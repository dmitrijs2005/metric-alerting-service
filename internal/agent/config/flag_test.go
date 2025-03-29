package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
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
		{"Test1 OK", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "5", "-k", "secretkey", "-l", "3"}, false, &Config{"127.0.0.1:9090", 20 * time.Second, 5 * time.Second, "secretkey", 3}},
		{"Test2 incorrect report interval", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "a", "-p", "5"}, true, &Config{}},
		{"Test3 incorrect poll interval", []string{"cmd", "-a", "127.0.0.1:9090", "-r", "20", "-p", "a"}, true, &Config{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.PanicOnError)

			os.Args = tt.args

			config := &Config{}

			if !tt.expectPanic {

				require.NotPanics(t, func() { parseFlags(config) })

				if diff := cmp.Diff(config, tt.expected); diff != "" {
					t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
				}
			} else {
				require.Panics(t, func() { parseFlags(config) })
			}
		})
	}
}
