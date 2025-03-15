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

	if envVar, ok := os.LookupEnv("STORE_INTERVAL"); ok {
		val, err := strconv.Atoi(envVar)
		if err != nil {
			panic(err)
		}
		config.StoreInterval = time.Duration(val) * time.Second
	}

	if envVar, ok := os.LookupEnv("FILE_STORAGE_PATH"); ok {
		config.FileStoragePath = envVar
	}

	if envVar, ok := os.LookupEnv("RESTORE"); ok {
		val, err := strconv.ParseBool(envVar)
		if err != nil {
			panic(err)
		}
		config.Restore = val
	}

	if envVar, ok := os.LookupEnv("DATABASE_DSN"); ok {
		config.DatabaseDSN = envVar
	}

}
