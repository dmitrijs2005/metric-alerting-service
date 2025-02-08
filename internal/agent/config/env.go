package config

import (
	"os"
	"strconv"
)

func parseEnv() {
	if addr, ok := os.LookupEnv("ADDRESS"); ok {
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
