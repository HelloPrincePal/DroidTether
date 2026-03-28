package daemon

import (
	"fmt"
	"os"
	"os/signal"
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

	watcher.OnAttach(func(dev *usb.Device) {
		d.wg.Add(1)
		defer d.wg.Done()
		log.Info().
			Str("component", "daemon").
			Msg("Android RNDIS device connected!")
		
		session := rndis.NewSession(dev)
		phoneMAC, err := session.Handshake()
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("RNDIS Handshake failed")
			return
		}

		iface, err := tun.OpenUTUN(0)
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Failed to create utun interface")
			return
		}

		// Ensure everything closes when watcher is stopped
		go func() {
			<-watcher.Context().Done()
			iface.Close()
		}()
		defer iface.Close()

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

					// Set DNS to Google (Primary) and phone gateway (Secondary)
					log.Info().Str("component", "daemon").Msg("Setting system DNS to 8.8.8.8 (Google)...")
					if err := iface.SetDNS([]string{"8.8.8.8", gateway}); err != nil {
						log.Warn().Str("component", "daemon").Err(err).Msg("Failed to set DNS")
					}
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
