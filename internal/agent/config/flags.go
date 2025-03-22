package config

import (
	"flag"
	"time"
)

func parseFlags(config *Config) {

	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")
	reportInterval := flag.Int("r", 10, "report interval (in seconds)")
	pollInterval := flag.Int("p", 2, "poll interval (in seconds)")
	flag.StringVar(&config.Key, "k", "", "signing key")
	sendRateLimit := flag.Int("l", 2, "sending rate limit")

	flag.Parse()

	config.ReportInterval = time.Duration(*reportInterval) * time.Second
	config.PollInterval = time.Duration(*pollInterval) * time.Second

	config.SendRateLimit = *sendRateLimit

}
