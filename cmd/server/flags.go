package main

import (
	"flag"
)

var flagEndpointAddr string

func parseFlags() {
	flag.StringVar(&flagEndpointAddr, "a", ":8080", "address and port to run server")
	flag.Parse()
}
