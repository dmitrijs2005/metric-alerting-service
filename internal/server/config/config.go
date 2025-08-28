// Package config handles the configuration for the server component,
// including parsing environment variables and command-line flags.
package config

import "time"

func (c *Config) LoadDefaults() {
	c.DatabaseDSN = ""
	c.EndpointAddr = ":8080"
	c.StoreInterval = time.Duration(30) * time.Second
	c.FileStoragePath = "/tmp/tmp.sav"
	c.Key = ""
	c.Restore = true
	c.CryptoKey = ""
	c.TrustedSubnet = ""
}

type Config struct {
	EndpointAddr    string
	FileStoragePath string
	DatabaseDSN     string
	Key             string
	StoreInterval   time.Duration
	Restore         bool
	CryptoKey       string
	TrustedSubnet   string
}

func LoadConfig() *Config {
	config := &Config{}
	config.LoadDefaults()

	parseJson(config)
	parseFlags(config)
	parseEnv(config)

	return config
}
