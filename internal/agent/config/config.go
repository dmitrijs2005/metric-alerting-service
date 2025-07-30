// Package config handles initialization and validation of the agent configuration,
// including reading from environment variables and command-line flags.
package config

import "time"

type Config struct {
	EndpointAddr   string
	Key            string
	ReportInterval time.Duration
	PollInterval   time.Duration
	SendRateLimit  int
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
