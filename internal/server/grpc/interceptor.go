package grpc

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// NewTrustedSubnetInterceptor returns a gRPC UnaryServerInterceptor that
// validates the client IP address against the given trusted subnet.
//
// The interceptor extracts the "x-real-ip" header from incoming metadata
// and checks if it belongs to the trusted subnet. Requests from outside
// the subnet are rejected with PermissionDenied.
//
// If no "x-real-ip" header is present, the request is also denied.
func NewTrustedSubnetInterceptor(trustedSubnet *net.IPNet) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		var realIP string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("x-real-ip")
			if len(values) > 0 {
				realIP = values[0]
			}
		}

		if realIP == "" {
			return nil, status.Error(codes.PermissionDenied, "cannot find real ip header")
		}

		if !trustedSubnet.Contains(net.ParseIP(realIP)) {
			return nil, status.Error(codes.PermissionDenied, "ip address is not in trusted subnet")
		}

		return handler(ctx, req)
	}
}
