package config

import (
	"os"
	"testing"
	"time"

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
		databaseDSN     string
		expected        *Config
	}{
		{"Test1 ip:port", "127.0.0.1:9090", "30", "/tmp/save.sav", "true", "db1", &Config{"127.0.0.1:9090", 30 * time.Second, "/tmp/save.sav", true, "db1"}},
		{"Test1 :port", ":8080", "25", "/tmp/save2.sav", "false", "db2", &Config{":8080", 25 * time.Second, "/tmp/save2.sav", false, "db2"}}, // Default value
		{"Test1 empty string", "", "25", "/tmp/save2.sav", "false", "db3", &Config{"", 25 * time.Second, "/tmp/save2.sav", false, "db3"}},    // Edge case: empty value
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			oldAddr := os.Getenv("ADDRESS")
			oldStoreInterval := os.Getenv("STORE_INTERVAL")
			oldFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
			oldRestore := os.Getenv("RESTORE")
			oldDatabaseDSN := os.Getenv("DATABASE_DSN")

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

			if err := os.Setenv("DATABASE_DSN", tt.databaseDSN); err != nil {
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
			if err := os.Setenv("DATABASE_DSN", oldDatabaseDSN); err != nil {
				panic(err)
			}

			if diff := cmp.Diff(config, tt.expected); diff != "" {
				t.Errorf("Structs mismatch (-config +expected):\n%s", diff)
			}
		})
	}
}
