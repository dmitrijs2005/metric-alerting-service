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
	Address        string          `json:"address"`
	ReportInterval common.Duration `json:"report_interval"`
	PollInterval   common.Duration `json:"poll_interval"`
	CryptoKey      string          `json:"crypto_key"`
	SendRateLimit  int             `json:"send_rate_limit"`
	Key            string          `json:"key"`
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
//   - ReportInterval
//   - PollInterval
//   - Key
//   - SendRateLimit
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
	config.ReportInterval = time.Duration(c.ReportInterval.Duration)
	config.PollInterval = time.Duration(c.PollInterval.Duration)
	config.Key = c.Key
	config.SendRateLimit = c.SendRateLimit
	config.CryptoKey = c.CryptoKey
}
