package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTempJSON(t *testing.T, dir, name string, data map[string]any) string {
	t.Helper()
	if dir == "" {
		dir = t.TempDir()
	}
	if name == "" {
		name = "cfg.json"
	}
	path := filepath.Join(dir, name)
	b, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, b, 0o600))
	return path
}

func Test_parseJson_SourcesAndPrecedence(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	dir := t.TempDir()
	pathEnv := writeTempJSON(t, dir, "env.json", map[string]any{
		"address":         "env.example:9000",
		"report_interval": "10s",
		"poll_interval":   "2s",
		"crypto_key":      "/env/key.pem",
		"send_rate_limit": 9,
		"key":             "ENVKEY",
	})
	pathFlag := writeTempJSON(t, dir, "flag.json", map[string]any{
		"address":         "flag.example:8000",
		"report_interval": "5s",
		"poll_interval":   "1s",
		"crypto_key":      "/flag/key.pem",
		"send_rate_limit": 3,
		"key":             "FLAGKEY",
	})

	t.Run("ENV wins over flags", func(t *testing.T) {
		t.Setenv("CONFIG", pathEnv)
		os.Args = []string{"testbin", "-c", pathFlag}

		cfg := &Config{}
		parseJson(cfg)

		assert.Equal(t, "env.example:9000", cfg.EndpointAddr)
		assert.Equal(t, 10*time.Second, cfg.ReportInterval)
		assert.Equal(t, 2*time.Second, cfg.PollInterval)
		assert.Equal(t, "/env/key.pem", cfg.CryptoKey)
		assert.Equal(t, 9, cfg.SendRateLimit)
		assert.Equal(t, "ENVKEY", cfg.Key)
	})

	t.Run("loads from flags when ENV is empty", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin", "-config", pathFlag}

		cfg := &Config{}
		parseJson(cfg)

		assert.Equal(t, "flag.example:8000", cfg.EndpointAddr)
		assert.Equal(t, 5*time.Second, cfg.ReportInterval)
		assert.Equal(t, 1*time.Second, cfg.PollInterval)
		assert.Equal(t, "/flag/key.pem", cfg.CryptoKey)
		assert.Equal(t, 3, cfg.SendRateLimit)
		assert.Equal(t, "FLAGKEY", cfg.Key)
	})

	t.Run("no CONFIG and no flags → no changes", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin"}

		cfg := &Config{
			EndpointAddr:   "defaults:1234",
			ReportInterval: 42 * time.Second,
			PollInterval:   7 * time.Second,
			SendRateLimit:  1,
			CryptoKey:      "ck",
			Key:            "k",
		}
		parseJson(cfg)

		assert.Equal(t, "defaults:1234", cfg.EndpointAddr)
		assert.Equal(t, 42*time.Second, cfg.ReportInterval)
		assert.Equal(t, 7*time.Second, cfg.PollInterval)
		assert.Equal(t, 1, cfg.SendRateLimit)
		assert.Equal(t, "ck", cfg.CryptoKey)
		assert.Equal(t, "k", cfg.Key)
	})

	t.Run("bad path → panics", func(t *testing.T) {
		t.Setenv("CONFIG", filepath.Join(dir, "missing.json"))
		os.Args = []string{"testbin"}
		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})

	t.Run("invalid JSON → panics", func(t *testing.T) {
		bad := filepath.Join(dir, "bad.json")
		require.NoError(t, os.WriteFile(bad, []byte(`{ this is not valid json`), 0o600))

		t.Setenv("CONFIG", bad)
		os.Args = []string{"testbin"}

		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})
}
