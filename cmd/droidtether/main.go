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
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *showVersion {
		fmt.Printf("droidtether v%s\n", version)
		return
	}

	fmt.Printf("droidtether v%s starting...\n", version)

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "droidtether: config not found or unreadable, trying local default.toml...\n")
		cfg, err = config.Load("config/default.toml")
		if err != nil {
			fmt.Fprintf(os.Stderr, "droidtether: failed to load any configuration: %v\n", err)
			os.Exit(1)
		}
	}

	d := daemon.New(cfg)
	if err := d.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "droidtether: daemon encountered an error: %v\n", err)
		os.Exit(1)
	}
}
