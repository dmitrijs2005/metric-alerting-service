package config

import (
	"os"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		expected       *Config
		name           string
		addr           string
		reportInterval string
		pollInterval   string
		key            string
		rateLimit      string
		expectPanic    bool
		cryptoKey      string
	}{
		{name: "Test1 OK", addr: "127.0.0.1:9090", reportInterval: "10", pollInterval: "5",
			key: "secretkey", rateLimit: "3", expectPanic: false, cryptoKey: "some_file.pem",
			expected: &Config{EndpointAddr: "127.0.0.1:9090", ReportInterval: 10 * time.Second,
				PollInterval: 5 * time.Second, Key: "secretkey", SendRateLimit: 3, CryptoKey: "some_file.pem"}},
		{name: "Test2 incorrect report interval", addr: "127.0.0.1:9090", reportInterval: "a", pollInterval: "5", key: "secretkey", rateLimit: "3", expectPanic: true, expected: &Config{}},
		{name: "Test2 incorrect report interval", addr: "127.0.0.1:9090", reportInterval: "20", pollInterval: "a", key: "secretkey", rateLimit: "3", expectPanic: true, expected: &Config{}},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			oldAddr := os.Getenv("ADDRESS")
			oldRI := os.Getenv("REPORT_INTERVAL")
			oldPI := os.Getenv("POLL_INTERVAL")
			oldKey := os.Getenv("KEY")
			oldRateLimit := os.Getenv("RATE_LIMIT")
			oldCryptoKey := os.Getenv("CRYPTO_KEY")

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
			if err := os.Setenv("CRYPTO_KEY", tt.cryptoKey); err != nil {
				panic(err)
			}

			config := &Config{}

			if !tt.expectPanic {
				require.NotPanics(t, func() { parseEnv(config) })

				err := os.Setenv("ADDRESS", oldAddr)
				if err != nil {
					panic(err)
				}

				if err = os.Setenv("ADDRESS", oldAddr); err != nil {
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
				if err = os.Setenv("CRYPTO_KEY", oldCryptoKey); err != nil {
					panic(err)
				}

				testutils.AssertEqualStructs(t, config, tt.expected)

			} else {
				require.Panics(t, func() { parseEnv(config) })
			}
		})
	}
}
