// Package config handles initialization and validation of the agent configuration,
// including reading from environment variables and command-line flags.
package config

import (
	"time"
)

func NewConfig() *Config {
	return &Config{
		EndpointAddr:   ":8080",
		ReportInterval: 10 * time.Second,
		PollInterval:   2 * time.Second,
		SendRateLimit:  3,
		Key:            "",
		CryptoKey:      "",
		UseGRPC:        false,
	}
}

type Config struct {
	EndpointAddr   string
	Key            string
	ReportInterval time.Duration
	PollInterval   time.Duration
	SendRateLimit  int
	CryptoKey      string
	UseGRPC        bool
}

func LoadConfig() *Config {
	config := NewConfig()

	parseJson(config)
	parseFlags(config)
	parseEnv(config)

	return config
}
