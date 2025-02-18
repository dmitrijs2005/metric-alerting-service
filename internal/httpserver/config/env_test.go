package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		addr     string
		expected *Config
	}{
		{"Test1 ip:port", "127.0.0.1:9090", &Config{"127.0.0.1:9090"}},
		{"Test1 :port", ":8080", &Config{":8080"}}, // Default value
		{"Test1 empty string", "", &Config{""}},    // Edge case: empty value
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			oldAddr := os.Getenv("ADDRESS")

			err := os.Setenv("ADDRESS", tt.addr)
			require.NoError(t, err)

			config := &Config{}
			parseEnv(config)

			err = os.Setenv("ADDRESS", oldAddr)
			if err != nil {
				panic(err)
			}

			if diff := cmp.Diff(config, tt.expected); diff != "" {
				t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
			}
		})
	}
}
