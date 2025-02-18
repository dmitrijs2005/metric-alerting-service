package config

import (
	"flag"
)

func parseFlags(config *Config) {
	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
}
