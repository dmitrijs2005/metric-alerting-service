package config

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		args     []string
		expected Config
	}{
		{"Test1 iP:port", []string{"cmd", "-a", "127.0.0.1:9090"}, Config{"127.0.0.1:9090"}},
		{"Test2 :port", []string{"cmd"}, Config{":8080"}},             // Default value
		{"Test3 empty string", []string{"cmd", "-a", ""}, Config{""}}, // Edge case: empty value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args
			parseFlags()

			if config != tt.expected {
				t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt, config)
			}
		})
	}
}
