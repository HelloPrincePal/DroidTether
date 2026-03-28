package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config represents the full configuration schema for DroidTether.
type Config struct {
	USB     USBConfig     `toml:"usb"`
	RNDIS   RNDISConfig   `toml:"rndis"`
	TUN     TUNConfig     `toml:"tun"`
	DHCP    DHCPConfig    `toml:"dhcp"`
	Logging LoggingConfig `toml:"logging"`
}

type USBConfig struct {
	PollIntervalMS int `toml:"poll_interval_ms"`
	ClaimTimeoutMS int `toml:"claim_timeout_ms"`
}

type RNDISConfig struct {
	MaxTransferSize int `toml:"max_transfer_size"`
	InitTimeoutMS   int `toml:"init_timeout_ms"`
}

type TUNConfig struct {
	InterfaceName string `toml:"interface_name"`
	MTU           int    `toml:"mtu"`
}

type DHCPConfig struct {
	TimeoutMS  int `toml:"timeout_ms"`
	RetryCount int `toml:"retry_count"`
}

type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
}

// Default paths for DroidTether config files on macOS.
var DefaultConfigPaths = []string{
	"/etc/droidtether/droidtether.toml",
	"/usr/local/etc/droidtether/droidtether.toml", // Intel Homebrew
	"/opt/homebrew/etc/droidtether/droidtether.toml", // ARM Homebrew
}

// Load reads and parses the configuration from the given path,
// or searches default paths if path is empty. [LLM]
func Load(path string) (*Config, error) {
	if path == "" {
		for _, p := range DefaultConfigPaths {
			if _, err := os.Stat(p); err == nil {
				path = p
				break
			}
		}
	}

	conf := &Config{}
	// If still no path, we'd normally load embedded defaults,
	// but for now we'll fail if no file is found and no path provided.
	if path == "" {
		return nil, fmt.Errorf("config: no config file found in default paths")
	}

	if _, err := toml.DecodeFile(path, conf); err != nil {
		return nil, fmt.Errorf("config: failed to decode %s: %w", path, err)
	}

	return conf, nil
}
