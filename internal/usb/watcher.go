package usb

import (
	"context"
	"time"

	"github.com/google/gousb"
	"github.com/rs/zerolog/log"
)

// Watcher monitors USB hotplug events and triggers callbacks when RNDIS devices connect.
type Watcher struct {
	ctx          context.Context
	cancel       context.CancelFunc
	pollInterval time.Duration
	onAttachFn   func(dev *Device)
	onDetachFn   func()
	usbCtx       *gousb.Context

	activeDeviceHandle *gousb.Device // Track single device handle for MVP
}

// NewWatcher creates a new Watcher bound to a polling interval.
func NewWatcher(pollInterval time.Duration) *Watcher {
	ctx, cancel := context.WithCancel(context.Background())
	return &Watcher{
		ctx:          ctx,
		cancel:       cancel,
		pollInterval: pollInterval,
		usbCtx:       gousb.NewContext(),
	}
}

// OnAttach registers the callback given a successfully initialized *Device struct.
func (w *Watcher) OnAttach(fn func(dev *Device)) {
	w.onAttachFn = fn
}

// OnDetach registers a callback for when the active RNDIS device is removed.
func (w *Watcher) OnDetach(fn func()) {
	w.onDetachFn = fn
}

// Context returns the watcher's current context.
func (w *Watcher) Context() context.Context {
	return w.ctx
}

// Start begins a polling loop for finding new RNDIS devices.
func (w *Watcher) Start() {
	go w.pollLoop()
}

// Stop cleanly stops the USB watcher and releases contexts.
func (w *Watcher) Stop() {
	if w.activeDeviceHandle != nil {
		w.activeDeviceHandle.Close()
		w.activeDeviceHandle = nil
	}
	w.cancel()
	w.usbCtx.Close()
}

// pollLoop continuously looks for the first Android RNDIS device every few seconds.
func (w *Watcher) pollLoop() {
	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	// After a Mac sleep the phone's USB tethering function can be
	// disabled by Android itself (VBUS drop -> tethering toggle off).
	// The daemon then re-attaches every poll and races open/close on a
	// device that is surprise-removed, which poisons the libusb context
	// permanently: libusb_get_device_list then returns LIBUSB_ERROR_OTHER
	// (-99) forever and the daemon goes deaf — it never notices the user
	// re-enabling tethering. No in-process recovery of that context is
	// reliable, so the scan loop tracks consecutive failures and exits
	// the process once the context is clearly poisoned; launchd
	// (KeepAlive, ThrottleInterval 5) restarts it with a fresh libusb
	// context that sees the device again.
	scanErrStreak := 0
	const scanErrFatalAt = 10

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if w.activeDeviceHandle != nil {
				continue
			}

			devs, err := w.usbCtx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
				for _, cfg := range desc.Configs {
					for _, intf := range cfg.Interfaces {
						for _, alt := range intf.AltSettings {
							if MatchRNDIS(uint16(desc.Vendor), uint16(desc.Product), uint8(alt.Class), uint8(alt.SubClass), uint8(alt.Protocol)) {
								return true
							}
						}
					}
				}
				if MatchRNDIS(uint16(desc.Vendor), uint16(desc.Product), 0xFF, 0x00, 0x00) {
					return true
				}
				return false
			})

			if err != nil {
				scanErrStreak++
				log.Error().
					Str("component", "usb").
					Err(err).
					Int("consecutive", scanErrStreak).
					Msg("Error scanning devices")
				for _, d := range devs {
					d.Close()
				}
				if scanErrStreak >= scanErrFatalAt {
					log.Fatal().
						Str("component", "usb").
						Int("consecutive", scanErrStreak).
						Msg("USB device-list context is poisoned; exiting for launchd restart")
				}
				continue
			}
			scanErrStreak = 0

			if len(devs) > 0 {
				log.Debug().
					Str("component", "usb").
					Int("candidates", len(devs)).
					Msg("Found match, checking interface setup...")
				candidate := devs[0]
				for i := 1; i < len(devs); i++ {
					devs[i].Close()
				}

				deviceWrapper, err := NewDevice(candidate)
				if err != nil {
					log.Warn().
						Str("component", "usb").
						Err(err).
						Msg("Failed to initialize Android RNDIS device wrapper")
					candidate.Close()
					continue
				}

				w.activeDeviceHandle = candidate
				if w.onAttachFn != nil {
					go func() {
						w.onAttachFn(deviceWrapper)
						if w.onDetachFn != nil {
							log.Debug().
								Str("component", "usb").
								Msg("Device detached workflow triggered")
							w.onDetachFn()
						}
						if w.activeDeviceHandle != nil {
							w.activeDeviceHandle.Close()
							w.activeDeviceHandle = nil
						}
					}()
				}
			}
		}
	}
}
