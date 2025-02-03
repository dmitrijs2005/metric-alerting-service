package config

import (
	"os"
)

func parseEnv() {
	if addr, ok := os.LookupEnv("ADDRESS"); ok {
		config.EndpointAddr = addr
	}
}
