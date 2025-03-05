package config

import (
	"flag"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestParseFlags(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		args     []string
		expected *Config
	}{
		{"Test1 iP:port", []string{"cmd", "-a=127.0.0.1:9090", "-i", "30", "-f", "/tmp/tmp.sav", "-d", "db", "-r", "true"},
			&Config{"127.0.0.1:9090", 30 * time.Second, "/tmp/tmp.sav", true, "db"}},
		{"Test2 :port", []string{"cmd"}, &Config{":8080", 30 * time.Second, "/tmp/tmp.sav", true, ""}},             // Default value
		{"Test3 empty string", []string{"cmd", "-a", ""}, &Config{"", 30 * time.Second, "/tmp/tmp.sav", true, ""}}, // Edge case: empty value
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			os.Args = tt.args

			config := &Config{}
			parseFlags(config)

			if diff := cmp.Diff(config, tt.expected); diff != "" {
				t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
			}
		})
	}
}
