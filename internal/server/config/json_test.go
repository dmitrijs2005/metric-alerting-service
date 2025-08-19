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
	p := filepath.Join(dir, name)
	b, err := json.Marshal(data)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(p, b, 0o600))
	return p
}

func Test_parseJson_PrecedenceAndParsing(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	tmp := t.TempDir()

	// JSON for ENV path (durations as string)
	envPath := writeTempJSON(t, tmp, "env.json", map[string]any{
		"address":        "env.example:9000",
		"store_file":     "/env/metrics.db",
		"database_dsn":   "postgres://env",
		"key":            "ENVKEY",
		"store_interval": "10s",
		"restore":        true,
		"crypto_key":     "/env/key.pem",
	})

	// JSON for flag path (durations as number, ns)
	flagPath := writeTempJSON(t, tmp, "flag.json", map[string]any{
		"address":        "flag.example:8000",
		"store_file":     "/flag/metrics.db",
		"database_dsn":   "postgres://flag",
		"key":            "FLAGKEY",
		"store_interval": 2_000_000_000, // 2s in ns
		"restore":        false,
		"crypto_key":     "/flag/key.pem",
	})

	t.Run("ENV path wins over flags path", func(t *testing.T) {
		t.Setenv("CONFIG", envPath)
		os.Args = []string{"testbin", "-c", flagPath}

		cfg := &Config{
			// preset to non-zero to ensure they change to JSON values
			EndpointAddr:    "preset:1",
			FileStoragePath: "preset",
			DatabaseDSN:     "preset",
			Key:             "preset",
			StoreInterval:   42 * time.Second,
			Restore:         false,
			CryptoKey:       "preset",
		}

		parseJson(cfg)

		assert.Equal(t, "env.example:9000", cfg.EndpointAddr)
		assert.Equal(t, "/env/metrics.db", cfg.FileStoragePath)
		assert.Equal(t, "postgres://env", cfg.DatabaseDSN)
		assert.Equal(t, "ENVKEY", cfg.Key)
		assert.Equal(t, 10*time.Second, cfg.StoreInterval)
		assert.Equal(t, true, cfg.Restore)
		assert.Equal(t, "/env/key.pem", cfg.CryptoKey)
	})

	t.Run("falls back to flags when CONFIG is empty", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin", "-config", flagPath}

		cfg := &Config{}
		parseJson(cfg)

		assert.Equal(t, "flag.example:8000", cfg.EndpointAddr)
		assert.Equal(t, "/flag/metrics.db", cfg.FileStoragePath)
		assert.Equal(t, "postgres://flag", cfg.DatabaseDSN)
		assert.Equal(t, "FLAGKEY", cfg.Key)
		assert.Equal(t, 2*time.Second, cfg.StoreInterval) // from ns
		assert.Equal(t, false, cfg.Restore)
		assert.Equal(t, "/flag/key.pem", cfg.CryptoKey)
	})

	t.Run("no CONFIG and no flags → no changes", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin"}

		orig := &Config{
			EndpointAddr:    "defaults:1234",
			FileStoragePath: "/tmp/db.json",
			DatabaseDSN:     "dsn",
			Key:             "k",
			StoreInterval:   3 * time.Second,
			Restore:         true,
			CryptoKey:       "ck",
		}
		cfg := *orig // copy
		parseJson(&cfg)
		assert.Equal(t, *orig, cfg)
	})

	t.Run("bad path in CONFIG → panics", func(t *testing.T) {
		t.Setenv("CONFIG", filepath.Join(tmp, "missing.json"))
		os.Args = []string{"testbin"}
		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})

	t.Run("invalid JSON → panics", func(t *testing.T) {
		bad := filepath.Join(tmp, "bad.json")
		require.NoError(t, os.WriteFile(bad, []byte(`{ not: "valid",`), 0o600))

		t.Setenv("CONFIG", bad)
		os.Args = []string{"testbin"}

		cfg := &Config{}
		require.Panics(t, func() { parseJson(cfg) })
	})
}

func Test_parseJson_DurationForms(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	tmp := t.TempDir()

	pathStr := writeTempJSON(t, tmp, "dur_str.json", map[string]any{
		"store_interval": "750ms",
	})
	pathNum := writeTempJSON(t, tmp, "dur_num.json", map[string]any{
		"store_interval": 750_000_000, // 750ms in ns
	})

	t.Run("duration as string", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin", "-c", pathStr}
		cfg := &Config{}
		parseJson(cfg)
		assert.Equal(t, 750*time.Millisecond, cfg.StoreInterval)
	})

	t.Run("duration as number (ns)", func(t *testing.T) {
		t.Setenv("CONFIG", "")
		os.Args = []string{"testbin", "-config", pathNum}
		cfg := &Config{}
		parseJson(cfg)
		assert.Equal(t, 750*time.Millisecond, cfg.StoreInterval)
	})
}
