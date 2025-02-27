package config

import (
	"os"
	"strconv"
	"time"
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
		config.ReportInterval = time.Duration(val) * time.Second
	}
	if envVar, ok := os.LookupEnv("POLL_INTERVAL"); ok {
		val, err := strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
		config.PollInterval = time.Duration(val) * time.Second
	}
}
