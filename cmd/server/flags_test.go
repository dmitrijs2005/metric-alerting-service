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
		expected string
	}{
		{[]string{"cmd", "-a", "127.0.0.1:9090"}, "127.0.0.1:9090"},
		{[]string{"cmd"}, ":8080"},      // Default value
		{[]string{"cmd", "-a", ""}, ""}, // Edge case: empty value
	}

	for _, tt := range tests {
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

		os.Args = tt.args
		parseFlags()

		if flagEndpointAddr != tt.expected {
			t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt.expected, flagEndpointAddr)
		}
	}
}
