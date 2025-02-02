package main

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		args     []string
		expected AgentOptions
	}{
		{[]string{"cmd", "-a", "127.0.0.1:9090", "-r", "5", "-p", "20"}, AgentOptions{EndpointAddr: "127.0.0.1:9090", ReportInterval: 5, PollInterval: 20}},
	}

	for _, tt := range tests {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		os.Args = tt.args
		parseFlags()

		if options.EndpointAddr != tt.expected.EndpointAddr {
			t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt.expected.EndpointAddr, options.EndpointAddr)
		}

		if options.ReportInterval != tt.expected.ReportInterval {
			t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt.expected.ReportInterval, options.ReportInterval)
		}

		if options.PollInterval != tt.expected.PollInterval {
			t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt.expected.PollInterval, options.PollInterval)
		}

	}
}
