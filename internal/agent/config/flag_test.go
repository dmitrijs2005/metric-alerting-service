package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		expected    *Config
		name        string
		args        []string
		expectPanic bool
	}{
		{name: "Test1 OK", args: []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "5", "-k", "secretkey", "-l", "3", "-crypto-key", "some_file.pem"}, expectPanic: false,
			expected: &Config{EndpointAddr: "127.0.0.1:9090", ReportInterval: 20 * time.Second, PollInterval: 5 * time.Second, Key: "secretkey", SendRateLimit: 3, CryptoKey: "some_file.pem"}},
		{name: "Test2 incorrect report interval", args: []string{"cmd", "-a", "127.0.0.1:9090", "-r", "a", "-p", "5"}, expectPanic: true, expected: &Config{}},
		{name: "Test3 incorrect poll interval", args: []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "a"}, expectPanic: true, expected: &Config{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)

			os.Args = tt.args

			config := &Config{}

			if !tt.expectPanic {

				require.NotPanics(t, func() { parseFlags(config) })
				assert.Empty(t, cmp.Diff(config, tt.expected))
			} else {
				require.Panics(t, func() { parseFlags(config) })
			}
		})
	}
}
