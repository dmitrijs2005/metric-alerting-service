package main

import (
	"flag"
)

type AgentOptions struct {
	EndpointAddr   string
	ReportInterval int
	PollInterval   int
}

var options AgentOptions

func parseFlags() {
	flag.StringVar(&options.EndpointAddr, "a", ":8080", "address and port to run server")
	flag.IntVar(&options.ReportInterval, "r", 10, "report interval")
	flag.IntVar(&options.PollInterval, "p", 2, "poll interval")
	flag.Parse()
}
