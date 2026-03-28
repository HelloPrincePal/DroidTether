package usb

import (
	"context"
	"fmt"
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

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			// If we already have a device attached, check if it's still there.
			// Re-evaluating USB device states on macOS typically requires re-listing to clear ghosts.
			if w.activeDeviceHandle != nil {
				// We can simply continue polling, but an advanced implementation would verify device presence.
				// For the sake of simplicity, we assume robust handled detach scenarios for now and skip listing.
				continue
			}

			// Open devices listing.
			devs, err := w.usbCtx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
				// Pre-filter: We cannot know interfaces at this stage easily in some architectures without
				// checking configs, but we can verify class matching in MatchRNDIS.
				// OpenDevices pre-filters using a fast check before full open.
				for _, cfg := range desc.Configs {
					for _, intf := range cfg.Interfaces {
						for _, alt := range intf.AltSettings {
							if MatchRNDIS(uint16(desc.Vendor), uint16(desc.Product), uint8(alt.Class), uint8(alt.SubClass), uint8(alt.Protocol)) {
								return true
							}
						}
					}
				}
				// Test secondary strings or known vendors.
				if MatchRNDIS(uint16(desc.Vendor), uint16(desc.Product), 0xFF, 0x00, 0x00) {
					return true
				}
				return false
			})

			if err != nil {
				log.Error().
					Str("component", "usb").
					Err(err).
					Msg("Error scanning devices")
				for _, d := range devs {
					d.Close()
				}
				continue
			}

			if len(devs) > 0 {
				log.Debug().
					Str("component", "usb").
					Int("candidates", len(devs)).
					Msg("Found match, checking interface setup...")
				candidate := devs[0] // take the first matched valid device. MVP only supports one device.
				// Close the others quickly.
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
						// when the device eventually detaches or errors out...
						if w.onDetachFn != nil {
							log.Debug().
								Str("component", "usb").
								Msg("Device detached workflow triggered")
							w.onDetachFn()
						}
						// Clean up handles.
						w.activeDeviceHandle.Close()
						w.activeDeviceHandle = nil
					}()
				}
			}
		}
	}
}
