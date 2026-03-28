package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/princePal/droidtether/config"
	"github.com/princePal/droidtether/internal/daemon"
)

// Injected by make build: -ldflags="-X main.version=..."
var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("droidtether v%s\n", version)
		return
	}

	fmt.Printf("droidtether v%s starting...\n", version)

	// Load configuration
	// Load from system paths first, fallback to local source tree config
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "droidtether: system config not found, trying local config/default.toml...\n")
		cfg, err = config.Load("config/default.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "droidtether: failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	d := daemon.New(cfg)
	if err := d.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "droidtether: daemon encountered an error: %v\n", err)
		os.Exit(1)
	}
}
