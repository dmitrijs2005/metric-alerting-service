package config

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		name        string
		args        []string
		expectPanic bool
		expected    *Config
	}{
		{"Test1 OK", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "5"}, false, &Config{EndpointAddr: "127.0.0.1:9090", ReportInterval: 20, PollInterval: 5}},
		{"Test2 incorrect report interval", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "a", "-p", "5"}, true, &Config{}},
		{"Test3 incorrect poll interval", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "a"}, true, &Config{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)

			os.Args = tt.args
			if !tt.expectPanic {

				require.NotPanics(t, parseFlags)

				if config.EndpointAddr != tt.expected.EndpointAddr {
					t.Errorf("parseFlags() with args %v; expected %q, got %q", tt.args, tt.expected.EndpointAddr, config.EndpointAddr)
				}

				if config.ReportInterval != tt.expected.ReportInterval {
					t.Errorf("parseFlags() with args %v; expected %d, got %d", tt.args, tt.expected.ReportInterval, config.ReportInterval)
				}

				if config.PollInterval != tt.expected.PollInterval {
					t.Errorf("parseFlags() with args %v; expected %d, got %d", tt.args, tt.expected.PollInterval, config.PollInterval)
				}
			} else {
				require.Panics(t, parseFlags)
			}
		})
	}
}
