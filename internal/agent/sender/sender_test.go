package sender

import (
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	pb "github.com/dmitrijs2005/metric-alerting-service/internal/proto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/proto"
)

func TestMetricToDto_ValidGauge(t *testing.T) {
	data := &sync.Map{}
	s, err := NewSender(data, time.Second, "http://localhost", "", 1, "", false)
	require.NoError(t, err)

	m := metric.NewGauge("cpu_load")
	m.Update(0.42)

	dto, err := s.MetricToDto(m)
	require.NoError(t, err)
	require.Equal(t, "cpu_load", dto.ID)
	require.Equal(t, "gauge", dto.MType)
	require.NotNil(t, dto.Value)
	require.Equal(t, 0.42, *dto.Value)
}

func TestSendMetric_Success(t *testing.T) {
	received := make(chan []byte, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		gr, err := gzip.NewReader(r.Body)
		require.NoError(t, err)
		defer gr.Close()

		body, _ := io.ReadAll(gr)
		received <- body

		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	s, _ := NewSender(data, time.Second, ts.URL, "", 1, "", false)

	m := metric.NewGauge("cpu_load")
	m.Update(1.23)

	err := s.SendMetric(m)
	require.NoError(t, err)

	var got map[string]interface{}
	require.NoError(t, json.Unmarshal(<-received, &got))
	require.Equal(t, "cpu_load", got["id"])
	require.Equal(t, "gauge", got["type"])
}

func TestSendAllMetricsInOneBatch_Success(t *testing.T) {
	received := make(chan []byte, 1)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gr, _ := gzip.NewReader(r.Body)
		defer gr.Close()
		body, _ := io.ReadAll(gr)
		received <- body
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	g := metric.NewGauge("temp")
	g.Update(99.9)
	data.Store("temp", g)

	s, _ := NewSender(data, time.Second, ts.URL, "", 1, "", false)

	err := s.SendAllMetricsInOneBatch()
	require.NoError(t, err)

	var arr []map[string]interface{}
	require.NoError(t, json.Unmarshal(<-received, &arr))
	require.Equal(t, "temp", arr[0]["id"])
	require.Equal(t, "gauge", arr[0]["type"])
}

func TestRun_SendsMetrics(t *testing.T) {
	var mu sync.Mutex
	count := 0

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		count++
		mu.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	data := &sync.Map{}
	g := metric.NewGauge("load")
	g.Update(0.99)
	data.Store("load", g)

	s, _ := NewSender(data, 50*time.Millisecond, ts.URL, "", 1, "", false)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go s.Run(ctx, &wg)

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	require.GreaterOrEqual(t, count, 1, "expected at least one batch sent")
}

func TestSender_GracefulShutdown_WaitsForInFlightMetrics(t *testing.T) {
	var mu sync.Mutex
	var received []string

	// fake server with artificial delay
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		// simulate slow server processing
		time.Sleep(300 * time.Millisecond)

		gr, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Errorf("failed to create gzip reader: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer gr.Close()

		var m dto.Metrics
		if err := json.NewDecoder(gr).Decode(&m); err != nil {
			t.Errorf("failed to decode json: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		received = append(received, m.ID)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// prepare sync.Map with a single metric
	data := &sync.Map{}
	data.Store("counter1", metric.NewCounter("counter1"))

	// create Sender with short report interval
	s, err := NewSender(data, 100*time.Millisecond, srv.URL, "", 1, "", false)
	if err != nil {
		t.Fatalf("failed to create sender: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go s.Run(ctx, &wg)

	// wait until first batch is definitely in-flight
	time.Sleep(150 * time.Millisecond)

	// initiate shutdown while request is still processing
	cancel()
	wg.Wait()

	// check that the in-flight request was completed
	mu.Lock()
	defer mu.Unlock()
	if len(received) != 1 {
		t.Fatalf("expected 1 metric, got %d: %v", len(received), received)
	}
}

const bufSize = 1024 * 1024

// bufconn dialer
func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

// plain server
type fakeMetricServer struct {
	pb.UnimplementedMetricServiceServer
	t    *testing.T
	priv *rsa.PrivateKey
}

func (f *fakeMetricServer) UpdateMetricValue(ctx context.Context, req *pb.UpdateMetricValueRequest) (*pb.UpdateMetricValueResponse, error) {
	require.Equal(f.t, "cpu", req.MetricName)
	require.Equal(f.t, "gauge", req.MetricType)
	require.Equal(f.t, "42", req.MetricValue)
	return &pb.UpdateMetricValueResponse{Value: req.MetricValue}, nil
}

// encrypted server
type fakeEncryptedServer struct {
	pb.UnimplementedMetricServiceServer
	t          *testing.T
	privateKey *rsa.PrivateKey
}

func (f *fakeEncryptedServer) UpdateMetricValueEncrypted(ctx context.Context, req *pb.EncryptedMessage) (*pb.UpdateMetricValueResponse, error) {
	decrypted, err := secure.DecryptRSAOAEPChunked(string(req.Data), f.privateKey)
	require.NoError(f.t, err)

	var realReq pb.UpdateMetricValueRequest
	err = proto.Unmarshal(decrypted, &realReq)
	require.NoError(f.t, err)

	require.Equal(f.t, "cpu", realReq.MetricName)
	require.Equal(f.t, "gauge", realReq.MetricType)
	require.Equal(f.t, "42", realReq.MetricValue)

	return &pb.UpdateMetricValueResponse{Value: realReq.MetricValue}, nil
}

func startBufconnServer(t *testing.T, srv pb.MetricServiceServer) (*grpc.ClientConn, func()) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	pb.RegisterMetricServiceServer(s, srv)

	go func() {
		if err := s.Serve(lis); err != nil {
			t.Logf("server exited: %v", err)
		}
	}()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	cleanup := func() {
		conn.Close()
		s.Stop()
	}
	return conn, cleanup
}

func TestSendMetricGRPC_Plain(t *testing.T) {
	conn, cleanup := startBufconnServer(t, &fakeMetricServer{t: t})
	defer cleanup()

	snd := &Sender{gRPCConn: conn} // PubKey=nil → UpdateMetricValue
	m := metric.MustNewGauge("cpu", 42)

	err := snd.SendMetricGRPC(m)
	require.NoError(t, err)
}

func TestSendMetricGRPC_Encrypted(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	pubKey := &privKey.PublicKey

	conn, cleanup := startBufconnServer(t, &fakeEncryptedServer{t: t, privateKey: privKey})
	defer cleanup()

	snd := &Sender{gRPCConn: conn, PubKey: pubKey} // PubKey!=nil → UpdateMetricValueEncrypted
	m := metric.MustNewGauge("cpu", 42)

	err = snd.SendMetricGRPC(m)
	require.NoError(t, err)
}

func TestSender_SendMetricGRPCEncrypted_HappyPath(t *testing.T) {

	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey

	lis := bufconn.Listen(1024 * 1024)
	s := grpc.NewServer()
	pb.RegisterMetricServiceServer(s, &fakeMetricServer{t: t, priv: priv})
	go s.Serve(lis)
	defer s.Stop()

	conn, err := grpc.NewClient(
		"passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, s string) (net.Conn, error) {
			return lis.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer conn.Close()

	client := pb.NewMetricServiceClient(conn)
	snd := &Sender{PubKey: pub}

	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "42",
	}

	err = snd.SendMetricGRPCEncrypted(metric.NewGauge("cpu"), client, req)
	require.NoError(t, err)
}

func (f *fakeMetricServer) UpdateMetricValueEncrypted(ctx context.Context, req *pb.EncryptedMessage) (*pb.UpdateMetricValueResponse, error) {
	decrypted, err := secure.DecryptRSAOAEPChunked(string(req.Data), f.priv)
	require.NoError(f.t, err)

	var m pb.UpdateMetricValueRequest
	err = proto.Unmarshal(decrypted, &m)
	require.NoError(f.t, err)

	require.Equal(f.t, "cpu", m.MetricName)
	return &pb.UpdateMetricValueResponse{Value: "42"}, nil
}

type fakeClient struct {
	resp *pb.UpdateMetricValueResponse
	err  error
}

func (f *fakeClient) UpdateMetricValueEncrypted(ctx context.Context, in *pb.EncryptedMessage, opts ...grpc.CallOption) (*pb.UpdateMetricValueResponse, error) {
	return f.resp, f.err
}

func (f *fakeClient) UpdateMetricValue(ctx context.Context, in *pb.UpdateMetricValueRequest, opts ...grpc.CallOption) (*pb.UpdateMetricValueResponse, error) {
	return nil, errors.New("not implemented")
}

func TestSendMetricGRPCEncrypted(t *testing.T) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)
	pub := &priv.PublicKey

	tests := []struct {
		name    string
		sender  *Sender
		client  pb.MetricServiceClient
		wantErr bool
	}{
		{
			name:    "happy path",
			sender:  &Sender{PubKey: pub},
			client:  &fakeClient{resp: &pb.UpdateMetricValueResponse{Value: "ok"}, err: nil},
			wantErr: false,
		},
		{
			name:    "encryption fails (nil key)",
			sender:  &Sender{PubKey: nil},
			client:  &fakeClient{resp: &pb.UpdateMetricValueResponse{Value: "ok"}, err: nil},
			wantErr: true,
		},
		{
			name:    "gRPC client returns error",
			sender:  &Sender{PubKey: pub},
			client:  &fakeClient{resp: nil, err: errors.New("network error")},
			wantErr: true,
		},
	}

	req := &pb.UpdateMetricValueRequest{
		MetricType:  "gauge",
		MetricName:  "cpu",
		MetricValue: "42",
	}
	m := metric.NewGauge("cpu")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sender.SendMetricGRPCEncrypted(m, tt.client, req)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSendMetricGRPCEncrypted_NoPubKey(t *testing.T) {
	s := &Sender{PubKey: nil}
	client := &fakeClient{}

	err := s.SendMetricGRPCEncrypted(metric.NewGauge("cpu"), client, &pb.UpdateMetricValueRequest{})
	require.Error(t, err)
}
