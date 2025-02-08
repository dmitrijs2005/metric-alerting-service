package config

import (
	"os"
	"strconv"
)

func parseEnv(config *Config) {
	if addr, ok := os.LookupEnv("ADDRESS"); ok && addr != "" {
		config.EndpointAddr = addr
	}
	if envVar, ok := os.LookupEnv("REPORT_INTERVAL"); ok {
		val, err := strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
		config.ReportInterval = val
	}
	if envVar, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		val, err := strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
		config.PollInterval = val
	}
}
