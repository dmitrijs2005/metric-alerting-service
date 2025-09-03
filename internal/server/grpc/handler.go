package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

func (s *MetricsServer) UpdateMetricValueEncrypted(ctx context.Context, req *pb.EncryptedMessage) (*pb.UpdateMetricValueResponse, error) {

	payload := req.Data

	if s.privateKey == nil {
		return nil, status.Error(codes.Internal, "no private key specified")
	}

	decrypted, err := secure.DecryptRSAOAEPChunked(string(payload), s.privateKey)
	if err != nil {
		return nil, status.Error(codes.Internal, "error decrypting")
	}

	m := &pb.UpdateMetricValueRequest{}
	err = proto.Unmarshal(decrypted, m)
	if err != nil {
		return nil, status.Error(codes.Internal, "error decrypting")
	}

	return s.UpdateMetricValue(ctx, m)

}
