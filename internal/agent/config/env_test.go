package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		name           string
		addr           string
		reportInterval string
		pollInterval   string
		expectPanic    bool
		expected       Config
	}{
		{"Test1 OK", "127.0.0.1:9090", "10", "5", false, Config{"127.0.0.1:9090", 10, 5}},
		{"Test2 incorrect report interval", "127.0.0.1:9090", "a", "5", true, Config{}},
		{"Test2 incorrect report interval", "127.0.0.1:9090", "20", "a", true, Config{}},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			oldAddr := os.Getenv("ADDRESS")
			oldRI := os.Getenv("REPORT_INTERVAL")
			oldPI := os.Getenv("POLL_INTERVAL")

			if err := os.Setenv("ADDRESS", tt.addr); err != nil {
				panic(err)
			}
			if err := os.Setenv("REPORT_INTERVAL", tt.reportInterval); err != nil {
				panic(err)
			}
			if err := os.Setenv("POLL_INTERVAL", tt.pollInterval); err != nil {
				panic(err)
			}

			if !tt.expectPanic {
				require.NotPanics(t, parseEnv)

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
				if config != tt.expected {
					t.Errorf("parseEnv() with args %v; expected %v, got %v", tt.addr, tt, config)
				}
			} else {
				require.Panics(t, parseEnv)
			}
		})
	}
}
