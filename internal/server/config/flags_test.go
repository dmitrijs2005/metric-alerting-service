package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/testutils"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		expected *Config
		name     string
		args     []string
	}{
		{name: "Test1 iP:port", args: []string{"cmd", "-a=127.0.0.1:9090", "-i", "30", "-f", "/tmp/tmp.sav", "-d", "db",
			"-k", "secretkey1", "-crypto-key", "some_file.pem", "-t", "192.168.1.0/24", "-g", ":3200", "-r", "true"},
			expected: &Config{EndpointAddr: "127.0.0.1:9090", StoreInterval: 30 * time.Second,
				FileStoragePath: "/tmp/tmp.sav", Restore: true, DatabaseDSN: "db", Key: "secretkey1", CryptoKey: "some_file.pem",
				TrustedSubnet: "192.168.1.0/24", GRPCEndpointAddr: ":3200"}}, // Edge case: empty value
		{name: "Test2 :port", args: []string{"cmd"},
			expected: &Config{EndpointAddr: ":8080", StoreInterval: 30 * time.Second,
				FileStoragePath: "/tmp/tmp.sav", Restore: true, DatabaseDSN: "", Key: "", GRPCEndpointAddr: ":50051"}}, // Default value
		{name: "Test3 empty string", args: []string{"cmd", "-a", ""},
			expected: &Config{EndpointAddr: "", StoreInterval: 30 * time.Second,
				FileStoragePath: "/tmp/tmp.sav", Restore: true, DatabaseDSN: "", Key: "", GRPCEndpointAddr: ":50051"}}, // Edge case: empty value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args

			config := &Config{}
			config.LoadDefaults()
			parseFlags(config)

			testutils.AssertEqualStructs(t, config, tt.expected)
		})
	}
}
