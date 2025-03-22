package config

import (
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		name           string
		addr           string
		reportInterval string
		pollInterval   string
		key            string
		rateLimit      string
		expectPanic    bool
		expected       *Config
	}{
		{"Test1 OK", "127.0.0.1:9090", "10", "5", "secretkey", "3", false, &Config{"127.0.0.1:9090", 10 * time.Second, 5 * time.Second, "secretkey", 3}},
		{"Test2 incorrect report interval", "127.0.0.1:9090", "a", "5", "secretkey", "3", true, &Config{}},
		{"Test2 incorrect report interval", "127.0.0.1:9090", "20", "a", "secretkey", "3", true, &Config{}},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			oldAddr := os.Getenv("ADDRESS")
			oldRI := os.Getenv("REPORT_INTERVAL")
			oldPI := os.Getenv("POLL_INTERVAL")
			oldKey := os.Getenv("KEY")
			oldRateLimit := os.Getenv("RATE_LIMIT")

			if err := os.Setenv("ADDRESS", tt.addr); err != nil {
				panic(err)
			}
			if err := os.Setenv("REPORT_INTERVAL", tt.reportInterval); err != nil {
				panic(err)
			}
			if err := os.Setenv("POLL_INTERVAL", tt.pollInterval); err != nil {
				panic(err)
			}
			if err := os.Setenv("KEY", tt.key); err != nil {
				panic(err)
			}
			if err := os.Setenv("RATE_LIMIT", tt.rateLimit); err != nil {
				panic(err)
			}

			config := &Config{}

			if !tt.expectPanic {
				require.NotPanics(t, func() { parseEnv(config) })

				err := os.Setenv("ADDRESS", oldAddr)
				if err != nil {
					panic(err)
				}

				if err := os.Setenv("ADDRESS", oldAddr); err != nil {
					panic(err)
				}
				if err = os.Setenv("REPORT_INTERVAL", oldRI); err != nil {
					panic(err)
				}
				if err = os.Setenv("POLL_INTERVAL", oldPI); err != nil {
					panic(err)
				}
				if err = os.Setenv("KEY", oldKey); err != nil {
					panic(err)
				}
				if err = os.Setenv("RATE_LIMIT", oldRateLimit); err != nil {
					panic(err)
				}

				if diff := cmp.Diff(config, tt.expected); diff != "" {
					t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
				}
			} else {
				require.Panics(t, func() { parseEnv(config) })
			}
		})
	}
}
