package config

import (
	"encoding/json"
	"os"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
)

// JsonConfig defines a configuration structure tailored for JSON unmarshalling.
// It uses common.Duration for interval fields, which allows parsing both
// string values such as "1s" and integer nanoseconds.
//
// This struct is an intermediate DTO (Data Transfer Object) used only for
// reading JSON configuration files. After unmarshalling, its fields are
// copied into the runtime Config struct which uses time.Duration.
type JsonConfig struct {
	Address       string          `json:"address"`
	Restore       bool            `json:"restore"`
	StoreInterval common.Duration `json:"store_interval"`
	StoreFile     string          `json:"store_file"`
	DatabaseDsn   string          `json:"database_dsn"`
	Key           string          `json:"key"`
	CryptoKey     string          `json:"crypto_key"`
}

// parseJson loads configuration values from a JSON file into the provided
// Config instance.
//
// The lookup order for the JSON file path is:
//  1. The CONFIG environment variable (highest priority).
//  2. The -c or -config command-line flags.
//  3. If neither is set, no JSON file is loaded.
//
// If the file path is found, parseJson attempts to read and unmarshal it
// into a JsonConfig. The resulting values are copied into the target Config.
// If the file cannot be read or contains invalid JSON, the function panics.
//
// Fields populated:
//   - EndpointAddr
//   - FileStoragePath
//   - DatabaseDSN
//   - Key
//   - StoreInterval
//   - Restore
//   - CryptoKey
//
// The caller is expected to merge these values with defaults, environment
// variables, and command-line flags as part of the full configuration process.
func parseJson(config *Config) {
	// first try environment variable
	jsonConfigFile := common.JsonConfigEnv()

	// fallback to flags
	if jsonConfigFile == "" {
		jsonConfigFile = common.JsonConfigFlags()
	}

	c := &JsonConfig{}

	// nothing to load
	if jsonConfigFile == "" {
		return
	}

	file, err := os.ReadFile(jsonConfigFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(file, c)
	if err != nil {
		panic(err)
	}

	config.EndpointAddr = c.Address
	config.FileStoragePath = c.StoreFile
	config.DatabaseDSN = c.DatabaseDsn
	config.Key = c.Key
	config.StoreInterval = time.Duration(c.StoreInterval.Duration)
	config.Restore = c.Restore
	config.CryptoKey = c.CryptoKey
}
