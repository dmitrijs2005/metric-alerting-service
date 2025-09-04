package grpc

import (
	"context"
	"crypto/rsa"
	"net"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"google.golang.org/grpc"
)

// MetricsServer implements the gRPC MetricServiceServer.
//
// It provides handlers for updating metric values (plain and encrypted),
// enforces optional trusted subnet restrictions, and manages lifecycle
// of the gRPC server instance.
type MetricsServer struct {
	pb.UnimplementedMetricServiceServer
	address       string
	storage       storage.Storage
	logger        logger.Logger
	trustedSubnet *net.IPNet
	privateKey    *rsa.PrivateKey
}

// NewgRPCMetricsServer creates a new instance of MetricsServer.
//
// Parameters:
//   - a: TCP address to bind the server on (e.g. ":8080").
//   - s: storage backend implementing Storage interface.
//   - l: logger for structured logging.
//   - trustedSubnet: optional CIDR string to restrict access
//     by client IP address (empty string disables check).
//   - cryptoKey: optional PEM-encoded RSA private key used for
//     decrypting encrypted requests (empty string disables encryption).
//
// Returns a configured MetricsServer instance or an error if parsing
// trustedSubnet or loading cryptoKey fails.
func NewgRPCMetricsServer(a string, s storage.Storage, l logger.Logger, trustedSubnet string, cryptoKey string) (*MetricsServer, error) {

	var cidr *net.IPNet
	var privKey *rsa.PrivateKey
	var err error
	if trustedSubnet != "" {
		_, cidr, err = net.ParseCIDR(trustedSubnet)
		if err != nil {
			return nil, err
		}
	}

	if cryptoKey != "" {
		privKey, err = secure.LoadRSAPrivateKeyFromPEM(cryptoKey)
		if err != nil {
			return nil, err
		}
	}

	return &MetricsServer{address: a, storage: s, logger: l, trustedSubnet: cidr, privateKey: privKey}, nil
}

// UpdateMetricValue handles an UpdateMetricValueRequest by creating or updating
// a metric in the storage backend.
//
// If the metric does not exist, it is created. Otherwise, its value is updated.
// Returns the resulting metric value as a string in the response.
func (s *MetricsServer) Run(ctx context.Context) error {

	// announces address
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	var opts []grpc.ServerOption

	if s.trustedSubnet != nil {
		opts = append(opts, grpc.UnaryInterceptor(NewTrustedSubnetInterceptor(s.trustedSubnet)))
	}

	// creates gRPC-server
	srv := grpc.NewServer(opts...)

	// registers service
	pb.RegisterMetricServiceServer(srv, s)

	go func() {
		<-ctx.Done()
		s.logger.Info("Stopping gPRC server...")
		srv.GracefulStop()
	}()

	s.logger.Infow("Starting gRPC server", "address", s.address)

	// starts accepting incoming connections
	if err := srv.Serve(listen); err != nil {
		return err
	}

	return nil
}
