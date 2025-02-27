package config

import "time"

type Config struct {
	EndpointAddr   string
	ReportInterval time.Duration
	PollInterval   time.Duration
}

func LoadConfig() *Config {
	config := &Config{}
	parseFlags(config)
	parseEnv(config)
	return config
}
