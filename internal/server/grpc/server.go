package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/usecase"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MetricsServer struct {
	pb.UnimplementedMetricServiceServer
	address string
	storage storage.Storage
	logger  logger.Logger
}

func NewgRPCMetricsServer(a string, s storage.Storage, l logger.Logger) (*MetricsServer, error) {
	return &MetricsServer{address: a, storage: s, logger: l}, nil
}

func (s *MetricsServer) Run(ctx context.Context) error {

	// announces address
	listen, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}

	// creates gRPC-server
	srv := grpc.NewServer()

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

func (s *MetricsServer) UpdateMetricValue(ctx context.Context, req *pb.UpdateMetricValueRequest) (*pb.UpdateMetricValueResponse, error) {
	var response pb.UpdateMetricValueResponse

	m, err := usecase.RetrieveMetric(ctx, s.storage, req.MetricType, req.MetricName)

	if err != nil {
		if !errors.Is(err, common.ErrorMetricDoesNotExist) {
			return nil, status.Error(codes.Internal, err.Error())
		} else {
			m, err = usecase.AddNewMetric(ctx, s.storage, req.MetricType, req.MetricName, req.MetricValue)
			if err != nil {
				return nil, err
			}
		}
	} else {
		err = usecase.UpdateMetric(ctx, s.storage, m, req.MetricValue)
		if err != nil {
			return nil, err
		}
	}

	response.Value = fmt.Sprintf("%v", m.GetValue())
	return &response, nil
}
