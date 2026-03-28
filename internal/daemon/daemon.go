package daemon

import (
	"os"
	"os/signal"
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
		log.Info().
			Str("component", "daemon").
			Msg("Android RNDIS device connected!")
		
		session := rndis.NewSession(dev)
		phoneMAC, err := session.Handshake()
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("RNDIS Handshake failed")
			return
		}

		// Milestone v0.4.0: Create virtual network interface
		iface, err := tun.OpenUTUN(0)
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Failed to create utun interface")
			return
		}
		defer iface.Close()

		log.Info().
			Str("component", "daemon").
			Str("interface", iface.Name()).
			Msg("Virtual network interface created and ACTIVE.")

		// Milestone v0.5.0: The Relay Engine
		relay, err := NewRelay(dev, iface, phoneMAC)
		if err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Failed to initialize Relay")
			time.Sleep(2 * time.Second) // prevent busy loops on retry
			return
		}

		stopChan := make(chan bool)
		watcher.OnDetach(func() {
			log.Info().Str("component", "daemon").Msg("Android RNDIS device detached.")
			relay.Stop()
			close(stopChan)
		})

		// Start the relay loop (blocks until error or stop)
		if err := relay.Start(); err != nil {
			log.Error().Str("component", "daemon").Err(err).Msg("Relay ended with error")
			time.Sleep(1 * time.Second)
		}

		// Wait here until phone is detached
		<-stopChan
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
