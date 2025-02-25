package config

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func parseFlags(config *Config) {
	// Print initial command-line args.
	fmt.Println("===")
	fmt.Println(os.Args)
	fmt.Println("===")

	// Setup flags.
	flag.StringVar(&config.EndpointAddr, "a", ":8080", "address and port to run server")
	reportInterval := flag.Int("r", 10, "report interval (in seconds)")
	pollInterval := flag.Int("p", 2, "poll interval (in seconds)")

	// Parse the flags; this updates the values.
	flag.Parse()

	// Now convert the integer values to time.Duration.
	config.ReportInterval = time.Duration(*reportInterval) * time.Second
	config.PollInterval = time.Duration(*pollInterval) * time.Second

	// Print the config after setting the values.
	fmt.Println(config)
}
