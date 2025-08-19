package config

import (
	"os"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/testutils"
)

func TestParseEnv(t *testing.T) {

	// Test cases
	tests := []struct {
		expected        *Config
		name            string
		addr            string
		storeInterval   string
		fileStoragePath string
		restore         string
		databaseDSN     string
		key             string
		cryptoKey       string
	}{
		{name: "Test1 ip:port", addr: "127.0.0.1:9090", storeInterval: "30", fileStoragePath: "/tmp/save.sav",
			restore: "true", databaseDSN: "db1", key: "secretkey1", cryptoKey: "some_file.pem",
			expected: &Config{EndpointAddr: "127.0.0.1:9090", StoreInterval: 30 * time.Second, FileStoragePath: "/tmp/save.sav",
				Restore: true, DatabaseDSN: "db1", Key: "secretkey1", CryptoKey: "some_file.pem"}},
		{name: "Test1 :port", addr: ":8080", storeInterval: "25", fileStoragePath: "/tmp/save2.sav",
			restore: "false", databaseDSN: "db2", key: "secretkey2",
			expected: &Config{EndpointAddr: ":8080", StoreInterval: 25 * time.Second, FileStoragePath: "/tmp/save2.sav",
				Restore: false, DatabaseDSN: "db2", Key: "secretkey2"}}, // Default value
		{name: "Test1 empty string", addr: "", storeInterval: "25", fileStoragePath: "/tmp/save2.sav",
			restore: "false", databaseDSN: "db3", key: "secretkey3",
			expected: &Config{EndpointAddr: "", StoreInterval: 25 * time.Second, FileStoragePath: "/tmp/save2.sav",
				Restore: false, DatabaseDSN: "db3", Key: "secretkey3"}}, // Edge case: empty value
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			oldAddr := os.Getenv("ADDRESS")
			oldStoreInterval := os.Getenv("STORE_INTERVAL")
			oldFileStoragePath := os.Getenv("FILE_STORAGE_PATH")
			oldRestore := os.Getenv("RESTORE")
			oldDatabaseDSN := os.Getenv("DATABASE_DSN")
			oldKey := os.Getenv("KEY")
			oldCryptoKey := os.Getenv("CRYPTO_KEY")

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

			if err := os.Setenv("KEY", tt.key); err != nil {
				panic(err)
			}

			if err := os.Setenv("CRYPTO_KEY", tt.cryptoKey); err != nil {
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
			if err := os.Setenv("KEY", oldKey); err != nil {
				panic(err)
			}
			if err := os.Setenv("CRYPTO_KEY", oldCryptoKey); err != nil {
				panic(err)
			}

			testutils.AssertEqualStructs(t, config, tt.expected)
		})
	}
}
