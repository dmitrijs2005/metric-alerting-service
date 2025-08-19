// Package config handles initialization and validation of the agent configuration,
// including reading from environment variables and command-line flags.
package config

import (
	"time"
)

func (c *Config) LoadDefaults() {
	c.EndpointAddr = ":8080"
	c.ReportInterval = time.Duration(10) * time.Second
	c.PollInterval = time.Duration(2) * time.Second
	c.SendRateLimit = 3
	c.Key = ""
	c.CryptoKey = ""
}

type Config struct {
	EndpointAddr   string
	Key            string
	ReportInterval time.Duration
	PollInterval   time.Duration
	SendRateLimit  int
	CryptoKey      string
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()

	parseJson(config)
	parseFlags(config)
	parseEnv(config)

	return config
}
