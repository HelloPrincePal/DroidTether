package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/princePal/droidtether/config"
	"github.com/princePal/droidtether/internal/rndis"
	"github.com/princePal/droidtether/internal/tun"
	"github.com/princePal/droidtether/internal/usb"
)

// Daemon represents the main executing body of DroidTether.
type Daemon struct {
	cfg *config.Config
	wg  sync.WaitGroup
}

// New creates a new Daemon with the loaded configuration.
func New(cfg *config.Config) *Daemon {
	return &Daemon{
		cfg: cfg,
	}
}

// Run starts the daemon loop, USB watcher, and blocks until an interrupt signal is received.
func (d *Daemon) Run() error {
	d.setupLogging()
	log.Info().Msg("Starting DroidTether...")

	// Use config polling interval, fallback to 1000ms if not set
	pollInterval := time.Duration(d.cfg.USB.PollIntervalMS) * time.Millisecond
	if pollInterval <= 0 {
		pollInterval = 1000 * time.Millisecond
	}

	watcher := usb.NewWatcher(pollInterval)

	// Serialize device sessions. OnAttach fires from the watcher's poll loop and
	// previously did no bookkeeping, so a dying session's USB claim was still
	// held when the next attach arrived. The new session then failed for ~8s
	// with "libusb: bad access [code -3]" before the old one finished cleaning
	// up, and each retry armed another watchdog. One session at a time; a new
	// attach waits for the previous one to finish releasing the device.
	var sessionMu sync.Mutex
	// Consecutive RNDIS handshake failures. Sessions are serialized by
	// sessionMu, so a plain counter is enough.
	var handshakeFails int

	watcher.OnAttach(func(dev *usb.Device) {
		sessionMu.Lock()
		defer sessionMu.Unlock()

		d.wg.Add(1)
		defer d.wg.Done()
		log.Info().
			Str("component", "daemon").
			Msg("Android RNDIS device connected!")

		// Always release the USB interfaces this session claimed. Without
		// this, a relay that ends on a libusb error path leaves the claim
		// dangling and the next attach cannot claim the interface.
		defer func() {
			if err := dev.Close(); err != nil {
				log.Warn().Str("component", "daemon").Err(err).Msg("Failed to close USB device")
			} else {
				log.Debug().Str("component", "daemon").Msg("USB device released")
			}
		}()

		session := rndis.NewSession(dev)
		phoneMAC, err := session.Handshake()
		if err != nil {
			handshakeFails++
			log.Error().Str("component", "daemon").Err(err).
				Int("consecutive", handshakeFails).
				Msg("RNDIS Handshake failed")
			// After a Mac sleep Android typically disables USB
			// tethering itself: the phone still enumerates the RNDIS
			// config but its gadget never answers INIT. Only the user
			// can switch it back on, so tell them, and back off —
			// the 1s reconnect storm against a surprise-removed
			// device is what poisoned the libusb context (-99).
			if handshakeFails == 5 {
				go notifyPhoneTetheringOff()
			}
			backoff := time.Duration(handshakeFails) * 2 * time.Second
			if backoff > 30*time.Second {
				backoff = 30 * time.Second
			}
			select {
			case <-time.After(backoff):
			case <-watcher.Context().Done():
			}
			return
		}
		handshakeFails = 0

		iface, err := tun.OpenUTUN(0)
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Failed to create utun interface")
			return
		}

		// Session-scoped context: the config watchdog must stop when this
		// device session ends, not linger across reconnects.
		sessionCtx, cancelSession := context.WithCancel(watcher.Context())
		defer cancelSession()

		// Ensure everything closes when watcher is stopped
		go func() {
			<-watcher.Context().Done()
			iface.Close()
		}()
		defer iface.Close()

		// The watchdog is started once per session; DHCP renewals may re-fire
		// OnDHCP, but the watchdog must not stack.
		var watchdogStarted sync.Once

		relay, err := NewRelay(watcher.Context(), dev, iface, phoneMAC)
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Failed to initialize Relay")
			time.Sleep(2 * time.Second) // prevent busy loops on retry
			return
		}

		relay.OnDHCP = func(gateway, client string) {
			log.Info().Str("component", "daemon").Str("gateway", gateway).Str("client", client).Msg("🔥 DHCPOFFER Intercepted! Auto-configuring network...")

			mtuStr := fmt.Sprintf("%d", d.cfg.TUN.MTU)
			if d.cfg.TUN.MTU <= 0 {
				mtuStr = "1400" // fallback
			}

			if err := iface.Configure(client, gateway, mtuStr); err != nil {
				log.Warn().Str("component", "daemon").Err(err).Msg("Failed to auto-configure interface IP")
			} else {
				log.Info().Str("component", "daemon").Str("mtu", mtuStr).Msg("✨ Network auto-configured! Ping should now work natively!")

				// Inject default route if configured
				if d.cfg.Route.SetDefaultRoute {
					log.Info().Str("component", "daemon").Msg("Rerouting all system traffic through DroidTether...")
					if err := iface.SetDefaultRoute(gateway); err != nil {
						log.Warn().Str("component", "daemon").Err(err).Msg("Failed to set default route")
					}

					// Make droidtether a real configd service so the
					// primary election picks it and the derived DNS
					// resolver gets RequestA (mDNSResponder otherwise
					// suppresses every A/AAAA query).
					if err := iface.RegisterService(client, gateway); err != nil {
						log.Warn().Str("component", "daemon").Err(err).Msg("Failed to register service with configd")
					}

					// Set DNS to Google (Primary) and phone gateway (Secondary)
					log.Info().Str("component", "daemon").Msg("Setting system DNS to 8.8.8.8 (Google)...")
					if err := iface.SetDNS([]string{"8.8.8.8", gateway}); err != nil {
						log.Warn().Str("component", "daemon").Err(err).Msg("Failed to set DNS")
					}

					// macOS standby can silently remove everything we just
					// injected while the USB session survives (no DHCP
					// re-fire). Arm the watchdog that re-applies on drift.
					watchdogStarted.Do(func() {
						d.startConfigWatchdog(sessionCtx, iface, client, gateway, mtuStr)
						go d.startLinkWatchdog(sessionCtx, relay, gateway, client)
					})
				}
			}
		}

		// Start the relay loop in its own goroutine
		errChan := make(chan error, 1)
		go func() {
			errChan <- relay.Start()
		}()

		// Wait here until the relay ends (cable pull) OR daemon is shutting down (Ctrl+C)
		select {
		case err := <-errChan:
			if err != nil {
				log.Warn().Str("component", "daemon").Err(err).Msg("Relay session ended")
			} else {
				log.Info().Str("component", "daemon").Msg("Relay session closed")
			}
		case <-watcher.Context().Done():
			log.Info().Str("component", "daemon").Msg("Daemon shutting down — stopping active relay...")
			relay.Stop()
			<-errChan // Wait for cleanup to finish
		}
	})

	// Start the USB hotplug watcher
	watcher.Start()
	log.Debug().
		Str("component", "daemon").
		Dur("poll_interval", pollInterval).
		Msg("USB watcher started. Waiting for devices...")

	// Block until graceful shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Info().
		Str("component", "daemon").
		Interface("signal", sig).
		Msg("Received signal, shutting down gracefully...")

	// Clean up
	watcher.Stop()
	log.Info().Msg("Waiting for active sessions to close...")
	d.wg.Wait()
	log.Info().Msg("Shutdown complete.")

	return nil
}

// notifyPhoneTetheringOff shows a macOS user notification through the
// console user's GUI session (the daemon runs as root). Best-effort:
// every failure is ignored.
func notifyPhoneTetheringOff() {
	out, err := exec.Command("stat", "-f", "%u", "/dev/console").Output()
	if err != nil {
		return
	}
	uid := strings.TrimSpace(string(out))
	if uid == "" || uid == "0" {
		return
	}
	_ = exec.Command("launchctl", "asuser", uid, "osascript", "-e",
		`display notification "USB tethering is disabled on the phone. Enable it in Settings → Network & internet → Hotspot & tethering → USB tethering." with title "DroidTether" sound name "Glass"`).Run()
}

func (d *Daemon) setupLogging() {
	// Level
	level, err := zerolog.ParseLevel(d.cfg.Logging.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Format
	if d.cfg.Logging.Format == "text" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	}
}
