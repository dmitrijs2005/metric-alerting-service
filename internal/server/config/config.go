// Package config handles the configuration for the server component,
// including parsing environment variables and command-line flags.
package config

import "time"

type Config struct {
	EndpointAddr    string
	FileStoragePath string
	DatabaseDSN     string
	Key             string
	StoreInterval   time.Duration
	Restore         bool
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
