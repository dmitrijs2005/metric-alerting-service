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

type MetricsServer struct {
	pb.UnimplementedMetricServiceServer
	address       string
	storage       storage.Storage
	logger        logger.Logger
	trustedSubnet *net.IPNet
	privateKey    *rsa.PrivateKey
}

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
