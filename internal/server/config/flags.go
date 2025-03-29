package config

import (
	"flag"
	"time"
)

func parseFlags(config *Config) {

	flag.StringVar(&config.DatabaseDSN, "d", "", "database dsn")

	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")

	var storeInterval int
	flag.IntVar(&storeInterval, "i", 30, "saved metric store interval")

	flag.StringVar(&config.FileStoragePath, "f", "/tmp/tmp.sav", "saved metric file storage path")

	flag.StringVar(&config.Key, "k", "", "signing key")
	flag.BoolVar(&config.Restore, "r", true, "restore saved metrics")

	flag.Parse()

	config.StoreInterval = time.Duration(storeInterval) * time.Second

}
