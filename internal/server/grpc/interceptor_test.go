package grpc

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestTrustedSubnetInterceptor(t *testing.T) {
	_, subnet, _ := net.ParseCIDR("192.168.0.0/24")
	interceptor := NewTrustedSubnetInterceptor(subnet)

	// Заготовка handler, который возвращает "ok"
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "ok", nil
	}

	t.Run("allowed ip", func(t *testing.T) {
		md := metadata.New(map[string]string{"x-real-ip": "192.168.0.42"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		require.NoError(t, err)
		require.Equal(t, "ok", resp)
	})

	t.Run("denied ip", func(t *testing.T) {
		md := metadata.New(map[string]string{"x-real-ip": "10.0.0.5"})
		ctx := metadata.NewIncomingContext(context.Background(), md)

		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		st, _ := status.FromError(err)
		require.Equal(t, "PermissionDenied", st.Code().String())
	})

	t.Run("missing header", func(t *testing.T) {
		ctx := context.Background()
		_, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: "/test.Method"}, handler)
		st, _ := status.FromError(err)
		require.Equal(t, "PermissionDenied", st.Code().String())
	})
}
