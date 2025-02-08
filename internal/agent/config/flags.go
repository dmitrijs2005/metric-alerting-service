package config

import (
	"flag"
)

func parseFlags(config *Config) {
	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&config.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&config.PollInterval, "p", 2, "poll interval")
	flag.Parse()
}
