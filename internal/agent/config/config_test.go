package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Defaults(t *testing.T) {
	cfg := LoadConfig()

	require.Equal(t, ":8080", cfg.EndpointAddr)
	require.Equal(t, 10*time.Second, cfg.ReportInterval)
	require.Equal(t, 2*time.Second, cfg.PollInterval)
	require.Equal(t, 3, cfg.SendRateLimit)
	require.Equal(t, "", cfg.Key)
	require.Equal(t, "", cfg.CryptoKey)
	require.False(t, cfg.UseGRPC)
}
