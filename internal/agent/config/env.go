package config

import (
	"os"
	"strconv"
	"time"
)

func parseEnv(config *Config) {
	if envVar, ok := os.LookupEnv("ADDRESS"); ok && envVar != "" {
		config.EndpointAddr = envVar
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
	if envVar, ok := os.LookupEnv("KEY"); ok && envVar != "" {
		config.Key = envVar
	}

	if envVar, ok := os.LookupEnv("RATE_LIMIT"); ok && envVar != "" {
		val, err := strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
		config.SendRateLimit = val
	}

}
