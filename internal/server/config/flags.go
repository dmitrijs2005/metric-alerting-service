package config

import (
	"flag"
)

func parseFlags(config *Config) {
	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&config.StoreInterval, "i", 30, "saved metric store interval")
	flag.StringVar(&config.FileStoragePath, "f", "/tmp/tmp.sav", "saved metric file storage path")
	flag.BoolVar(&config.Restore, "r", true, "restore saved metrics")
	flag.Parse()
}
