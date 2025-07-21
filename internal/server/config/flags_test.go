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
		name     string
		args     []string
		expected *Config
	}{
		{"Test1 iP:port", []string{"cmd", "-a=127.0.0.1:9090", "-i", "30", "-f", "/tmp/tmp.sav", "-d", "db", "-k", "secretkey1", "-r", "true"},
			&Config{"127.0.0.1:9090", 30 * time.Second, "/tmp/tmp.sav", true, "db", "secretkey1"}},
		{"Test2 :port", []string{"cmd"}, &Config{":8080", 30 * time.Second, "/tmp/tmp.sav", true, "", ""}},             // Default value
		{"Test3 empty string", []string{"cmd", "-a", ""}, &Config{"", 30 * time.Second, "/tmp/tmp.sav", true, "", ""}}, // Edge case: empty value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args

			config := &Config{}
			parseFlags(config)

			testutils.AssertEqualStructs(t, config, tt.expected)
		})
	}
}
