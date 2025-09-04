// Package grpc implements the gRPC server layer for the metric alerting service.
// It provides a MetricsServer that exposes gRPC endpoints for updating metrics,
// optionally with encryption, and supports access control by trusted subnet.
package grpc
