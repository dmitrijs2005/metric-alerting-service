package config

import (
	"os"
)

func parseEnv(config *Config) {
	if addr, ok := os.LookupEnv("ADDRESS"); ok && addr != "" {
		config.EndpointAddr = addr
	}
}
