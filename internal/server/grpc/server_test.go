package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestNewgRPCMetricsServer(t *testing.T) {
	s := memory.NewMemStorage()
	l := logger.GetLogger()
	// valid subnet
	srv, err := NewgRPCMetricsServer(":0", s, l, "192.168.0.0/24", "")
	require.NoError(t, err)
	require.NotNil(t, srv)

	// wrong format
	_, err = NewgRPCMetricsServer(":0", s, l, "invalid-subnet", "")
	require.Error(t, err)

	// empty subnet
	srv, err = NewgRPCMetricsServer(":0", s, l, "", "")
	require.NoError(t, err)
	require.Nil(t, srv.trustedSubnet)
}

func dialer(s *grpc.Server, lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestMetricsServer_Run(t *testing.T) {

	s := memory.NewMemStorage()
	l := logger.GetLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	lis := bufconn.Listen(1024 * 1024)
	srv := &MetricsServer{
		address: ":0", // bufconn
		logger:  l,
		storage: s,
	}

	go func() {
		_ = srv.Run(ctx)
	}()

	si := grpc.NewServer()
	pb.RegisterMetricServiceServer(si, srv)
	go si.Serve(lis)
	defer si.Stop()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(dialer(si, lis)),
		grpc.WithInsecure(),
	)

	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewMetricServiceClient(conn)
	_, err = client.UpdateMetricValue(ctx, &pb.UpdateMetricValueRequest{
		MetricType: "gauge", MetricName: "cpu", MetricValue: "42",
	})
	require.NoError(t, err)

	cancel()
}
