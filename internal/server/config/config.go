// Package config handles the configuration for the server component,
// including parsing environment variables and command-line flags.
package config

import "time"

type Config struct {
	EndpointAddr    string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
	DatabaseDSN     string
	Key             string
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
