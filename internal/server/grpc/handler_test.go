package grpc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// fakeMetricsServer наследуем от твоего MetricsServer, чтобы подменить UpdateMetricValue
type fakeMetricsServer struct {
	*MetricsServer
	lastReq *pb.UpdateMetricValueRequest
}

func (f *fakeMetricsServer) UpdateMetricValue(ctx context.Context, req *pb.UpdateMetricValueRequest) (*pb.UpdateMetricValueResponse, error) {
	f.lastReq = req
	return &pb.UpdateMetricValueResponse{Value: req.MetricValue}, nil
}

func TestUpdateMetricValueEncrypted(t *testing.T) {
	ctx := context.Background()

	t.Run("no private key", func(t *testing.T) {
		srv := &MetricsServer{ /* privateKey=nil */ }
		_, err := srv.UpdateMetricValueEncrypted(ctx, &pb.EncryptedMessage{Data: []byte("abc")})
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
		require.Contains(t, err.Error(), "no private key")
	})

	t.Run("invalid payload", func(t *testing.T) {
		priv, _ := rsa.GenerateKey(rand.Reader, 2048)
		srv := &MetricsServer{privateKey: priv}
		_, err := srv.UpdateMetricValueEncrypted(ctx, &pb.EncryptedMessage{Data: []byte("garbage")})
		require.Error(t, err)
		require.Equal(t, codes.Internal, status.Code(err))
		require.Contains(t, err.Error(), "error decrypting")
	})

	t.Run("happy path", func(t *testing.T) {
		// генерим RSA-ключи
		priv, _ := rsa.GenerateKey(rand.Reader, 2048)
		fakeSrv := &fakeMetricsServer{MetricsServer: &MetricsServer{privateKey: priv, storage: memory.NewMemStorage()}}

		// делаем валидный запрос
		req := &pb.UpdateMetricValueRequest{
			MetricType:  "gauge",
			MetricName:  "cpu",
			MetricValue: "42",
		}
		raw, err := proto.Marshal(req)
		require.NoError(t, err)

		// шифруем публичным ключом
		encrypted, err := secure.EncryptRSAOAEPChunked(raw, &priv.PublicKey)
		require.NoError(t, err)

		resp, err := fakeSrv.UpdateMetricValueEncrypted(ctx, &pb.EncryptedMessage{Data: []byte(encrypted)})
		require.NoError(t, err)
		require.Equal(t, "42", resp.Value)

		// проверяем, что доехало до UpdateMetricValue
		// fmt.Println(fakeSrv.lastReq)
		// require.NotNil(t, fakeSrv.lastReq)
		// require.Equal(t, "cpu", fakeSrv.lastReq.MetricName)
		// require.Equal(t, "gauge", fakeSrv.lastReq.MetricType)
	})
}

func TestUpdateMetricValueEncrypted_HappyPath(t *testing.T) {
	ctx := context.Background()

	// 1. генерим ключи
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// 2. создаём сервер с memory-storage и приватным ключом
	st := memory.NewMemStorage()
	srv := &MetricsServer{
		storage:    st,
		privateKey: priv,
	}

	// 3. делаем обычный запрос
	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "42",
	}
	raw, err := proto.Marshal(req)
	require.NoError(t, err)

	// 4. шифруем публичным ключом
	encrypted, err := secure.EncryptRSAOAEPChunked(raw, &priv.PublicKey)
	require.NoError(t, err)

	// 5. вызываем зашифрованный метод
	resp, err := srv.UpdateMetricValueEncrypted(ctx, &pb.EncryptedMessage{Data: []byte(encrypted)})
	require.NoError(t, err)
	require.Equal(t, "42", resp.Value)

	// 6. проверяем, что в сторедже появилась метрика
	m, err := st.Retrieve(ctx, "gauge", "cpu")
	require.NoError(t, err)
	require.Equal(t, float64(42), m.GetValue())
}

func TestMetricsServer_UpdateMetricValue_AddsNew(t *testing.T) {
	ctx := context.Background()
	st := memory.NewMemStorage()
	srv := &MetricsServer{storage: st}

	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "42",
	}

	resp, err := srv.UpdateMetricValue(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "42", resp.Value)

	// verify in storage
	m, err := st.Retrieve(ctx, metric.MetricTypeGauge, "cpu")
	require.NoError(t, err)
	require.Equal(t, float64(42), m.GetValue())
}

func TestMetricsServer_UpdateMetricValue_UpdatesExisting(t *testing.T) {
	ctx := context.Background()
	st := memory.NewMemStorage()
	srv := &MetricsServer{storage: st}

	// сначала добавляем вручную
	g := &metric.Gauge{Name: "cpu", Value: 10}
	err := st.Add(ctx, g)
	require.NoError(t, err)

	// теперь обновляем через метод
	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "99",
	}

	resp, err := srv.UpdateMetricValue(ctx, req)
	require.NoError(t, err)
	require.Equal(t, "99", resp.Value)

	// проверяем в сторедже
	m, err := st.Retrieve(ctx, metric.MetricTypeGauge, "cpu")
	require.NoError(t, err)
	require.Equal(t, float64(99), m.GetValue())
}

func TestMetricsServer_UpdateMetricValue_ErrorFromStorage(t *testing.T) {
	ctx := context.Background()

	// fake storage с ошибкой
	badStorage := &brokenStorage{}
	srv := &MetricsServer{storage: badStorage}

	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "42",
	}

	_, err := srv.UpdateMetricValue(ctx, req)
	require.Error(t, err)

	stErr, ok := status.FromError(err)
	require.True(t, ok)
	require.Equal(t, codes.Internal, stErr.Code())
}

// brokenStorage всегда возвращает ошибку
type brokenStorage struct{}

func (b *brokenStorage) Add(ctx context.Context, m metric.Metric) error {
	return errors.New("db error")
}
func (b *brokenStorage) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return errors.New("db error")
}
func (b *brokenStorage) Retrieve(ctx context.Context, mt metric.MetricType, n string) (metric.Metric, error) {
	return nil, errors.New("db error")
}
func (b *brokenStorage) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, errors.New("db error")
}
func (b *brokenStorage) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return errors.New("db error")
}

func TestUpdateMetricValueEncrypted_DecryptError(t *testing.T) {
	// подсовываем приватный ключ, но данные невалидные
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	s := &MetricsServer{privateKey: privKey}

	_, err := s.UpdateMetricValueEncrypted(context.Background(), &pb.EncryptedMessage{Data: []byte("garbage")})
	require.Error(t, err)
	require.Contains(t, err.Error(), "error decrypting")
}
