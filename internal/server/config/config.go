package config

import "time"

//var config Config

type Config struct {
	EndpointAddr    string
	StoreInterval   time.Duration
	FileStoragePath string
	Restore         bool
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
