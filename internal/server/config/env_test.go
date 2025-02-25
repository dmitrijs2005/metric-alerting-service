package config

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		name            string
		addr            string
		storeInterval   string
		fileStoragePath string
		restore         string
		expected        *Config
	}{
		{"Test1 ip:port", "127.0.0.1:9090", "30", "/tmp/save.sav", "true", &Config{"127.0.0.1:9090", 30, "/tmp/save.sav", true}},
		{"Test1 :port", ":8080", "25", "/tmp/save2.sav", "false", &Config{":8080", 25, "/tmp/save2.sav", false}}, // Default value
		{"Test1 empty string", "", "25", "/tmp/save2.sav", "false", &Config{"", 25, "/tmp/save2.sav", false}},    // Edge case: empty value
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			oldAddr := os.Getenv("ADDRESS")
			oldStoreInterval := os.Getenv("STORE_INTERVAL")
			oldFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
			oldRestore := os.Getenv("RESTORE")

			if err := os.Setenv("ADDRESS", tt.addr); err != nil {
				panic(err)
			}

			if err := os.Setenv("STORE_INTERVAL", tt.storeInterval); err != nil {
				panic(err)
			}

			if err := os.Setenv("FILE_STORAGE_PATH", tt.fileStoragePath); err != nil {
				panic(err)
			}

			if err := os.Setenv("RESTORE", tt.restore); err != nil {
				panic(err)
			}

			config := &Config{}
			parseEnv(config)

			if err := os.Setenv("ADDRESS", oldAddr); err != nil {
				panic(err)
			}
			if err := os.Setenv("STORE_INTERVAL", oldStoreInterval); err != nil {
				panic(err)
			}
			if err := os.Setenv("FILE_STORAGE_PATH", oldFileStoragePath); err != nil {
				panic(err)
			}
			if err := os.Setenv("RESTORE", oldRestore); err != nil {
				panic(err)
			}

			if diff := cmp.Diff(config, tt.expected); diff != "" {
				t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
			}
		})
	}
}
