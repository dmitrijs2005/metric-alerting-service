package grpc

import (
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/require"
)

type dummyStorage struct{}

func TestNewgRPCMetricsServer(t *testing.T) {
	s := memory.NewMemStorage()
	l := logger.GetLogger()
	// valid subnet
	srv, err := NewgRPCMetricsServer(":0", s, l, "192.168.0.0/24")
	require.NoError(t, err)
	require.NotNil(t, srv)

	// wrong format
	_, err = NewgRPCMetricsServer(":0", s, l, "invalid-subnet")
	require.Error(t, err)

	// empty subnet
	srv, err = NewgRPCMetricsServer(":0", s, l, "")
	require.NoError(t, err)
	require.Nil(t, srv.trustedSubnet)
}
