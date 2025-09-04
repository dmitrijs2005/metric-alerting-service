package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/assets"
	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/usecase"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	metric1 = &metric.Counter{Name: "counter1", Value: 1}
	metric2 = &metric.Gauge{Name: "gauge1", Value: 1.234}
)

type faultyStorage struct{}

func (f faultyStorage) Add(ctx context.Context, m metric.Metric) error {
	return errors.New("forced error in Add")
}
func (f faultyStorage) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return errors.New("forced error in Update")
}
func (f faultyStorage) Retrieve(ctx context.Context, t metric.MetricType, n string) (metric.Metric, error) {
	return nil, errors.New("forced error in Retrieve")
}
func (f faultyStorage) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, errors.New("forced error in RetrieveAll")
}
func (f faultyStorage) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return errors.New("forced error in UpdateBatch")
}

func prepareTestSTorage() storage.Storage {
	s := memory.NewMemStorage()

	s.Data["counter|counter1"] = metric1
	s.Data["gauge|gauge1"] = metric2
	return s
}

func prepareTestServer() *HTTPServer {
	a := "http://localhost:8080"
	s := prepareTestSTorage()

	return &HTTPServer{
		Address: a,
		Storage: s,
	}

}

func TestHTTPServer_UpdateHandler(t *testing.T) {

	a := "http://localhost:8080"
	s := memory.NewMemStorage()

	type want struct {
		response    string
		contentType string
		code        int
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{

		{name: "Counter OK", method: http.MethodPost, url: "/update/counter/counter1/1", want: want{code: 200, response: "OK", contentType: "text/plain; charset=UTF-8"}},
		{name: "Invalid metric type - Bad Request", method: http.MethodPost, url: "/update/unknown/u1/1", want: want{code: http.StatusBadRequest, response: "invalid metric type", contentType: "text/plain; charset=UTF-8"}},
		{name: "Counter Bad request", method: http.MethodPost, url: "/update/counter/counter1/a", want: want{code: 400, response: "invalid metric value", contentType: "text/plain; charset=UTF-8"}},

		{name: "Gauge OK", method: http.MethodPost, url: "/update/gauge/gauge1/1.1", want: want{code: 200, response: "OK", contentType: "text/plain; charset=UTF-8"}},
		{name: "Gauge Bad request", method: http.MethodPost, url: "/update/gauge/gauge1/a", want: want{code: 400, response: "invalid metric value", contentType: "text/plain; charset=UTF-8"}},
		{name: "Gauge Bad request", method: http.MethodPost, url: "/update/gauge/gauge1/1,2", want: want{code: 400, response: "invalid metric value", contentType: "text/plain; charset=UTF-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: a,
				Storage: s,
			}

			e := echo.New()

			request := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(request, rec)

			parts := strings.Split(tt.url, "/")
			c.SetParamNames("type", "name", "value")
			c.SetParamValues(parts[2], parts[3], parts[4])

			if assert.NoError(t, s.UpdateHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.response, rec.Body.String())
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))
			}

		})
	}
}

func TestHTTPServer_UpdateHandler_404_405(t *testing.T) {

	a := "http://localhost:8080"
	s := memory.NewMemStorage()

	type want struct {
		response    string
		contentType string
		code        int
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{
		{name: "Get - Method Not Allowed", method: http.MethodGet, url: "/update/gauge/gauge1/1.234", want: want{code: 405, response: "Method Not Allowed\n", contentType: "text/plain; charset=utf-8"}},
		{name: "No metric name - Not Found", method: http.MethodPost, url: "/update/counter", want: want{code: 404, response: "Not Found\n", contentType: "text/plain; charset=utf-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: a,
				Storage: s,
			}

			e := echo.New()
			e.POST("/update/:type/:name/:value", s.UpdateHandler)

			request := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, request)
			assert.Equal(t, tt.want.code, rec.Code)

		})
	}
}

func TestHTTPServer_ValueHandler(t *testing.T) {

	m1 := &metric.Counter{Name: "counter1", Value: 1}
	m2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	type want struct {
		response    string
		contentType string
		code        int
	}
	tests := []struct {
		name   string
		method string
		url    string
		want   want
	}{

		{name: "Counter OK", method: http.MethodGet, url: "/value/counter/counter1", want: want{code: 200, response: fmt.Sprintf("%v", m1.GetValue()), contentType: "text/plain; charset=UTF-8"}},
		{name: "Gauge OK", method: http.MethodGet, url: "/value/gauge/gauge1", want: want{code: 200, response: fmt.Sprintf("%v", m2.GetValue()), contentType: "text/plain; charset=UTF-8"}},
		{name: "Unnown metric", method: http.MethodGet, url: "/value/gauge/unknwn", want: want{code: 404, response: common.ErrorMetricDoesNotExist.Error(), contentType: "text/plain; charset=UTF-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: addr,
				Storage: stor,
			}

			e := echo.New()

			request := httptest.NewRequest(tt.method, "/", nil)
			rec := httptest.NewRecorder()

			c := e.NewContext(request, rec)

			parts := strings.Split(tt.url, "/")
			c.SetParamNames("type", "name")
			c.SetParamValues(parts[2], parts[3])

			if assert.NoError(t, s.ValueHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.response, rec.Body.String())
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))
			}

		})
	}
}

func TestHTTPServer_ListHandler(t *testing.T) {

	m1 := &metric.Counter{Name: "counter1", Value: 1}
	m2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	stor.Data["counter|counter1"] = m1
	stor.Data["gauge|gauge1"] = m2

	s := &HTTPServer{
		Address: addr,
		Storage: stor,
	}

	e := echo.New()

	// Load templates
	r := &Template{
		templates: template.Must(template.ParseFS(assets.WebTemplates, "template/*.html")),
	}
	e.Renderer = r

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(request, rec)

	if assert.NoError(t, s.ListHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, strings.Contains(rec.Body.String(), m1.Name))
		assert.True(t, strings.Contains(rec.Body.String(), m2.Name))

		assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
	}

}

func TestHTTPServer_ValueJSONHandler(t *testing.T) {

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	type want struct {
		response    *dto.Metrics
		contentType string
		code        int
	}
	tests := []struct {
		name    string
		method  string
		mtype   string
		mname   string
		mvalue  interface{}
		storage storage.Storage
		want    want
	}{

		{name: "Counter OK", storage: stor, method: http.MethodPost, mtype: "counter", mname: metric1.Name, mvalue: metric1.Value, want: want{code: 200, response: &dto.Metrics{ID: metric1.Name, Delta: &metric1.Value, MType: "counter"}, contentType: "application/json"}},
		{name: "Gauge OK", storage: stor, method: http.MethodGet, mtype: "gauge", mname: metric2.Name, mvalue: metric2.Value, want: want{code: 200, response: &dto.Metrics{ID: metric2.Name, Value: &metric2.Value, MType: "gauge"}, contentType: "application/json"}},
		{name: "Unnown metric", storage: stor, method: http.MethodPost, mtype: "unknown", mname: "", mvalue: 0, want: want{code: 404, response: nil, contentType: "text/plain; charset=UTF-8"}},
		{name: "Bad storage", storage: faultyStorage{}, method: http.MethodPost, mtype: "unknown", mname: "", mvalue: 0, want: want{code: 500, response: nil, contentType: "text/plain; charset=UTF-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: addr,
				Storage: tt.storage,
			}

			e := echo.New()

			payload := &dto.Metrics{ID: tt.mname, MType: tt.mtype}

			// Marshal the payload to JSON
			jsonData, err := json.Marshal(payload)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}

			request := httptest.NewRequest(tt.method, "/", bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			c := e.NewContext(request, rec)

			if assert.NoError(t, s.ValueJSONHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))

				if rec.Code == http.StatusOK {
					var response dto.Metrics
					err := json.Unmarshal(rec.Body.Bytes(), &response)

					assert.NoError(t, err)

					assert.Equal(t, payload.ID, tt.want.response.ID)
					assert.Equal(t, payload.MType, tt.want.response.MType)

					if tt.mtype == "counter" {
						assert.Equal(t, tt.mvalue, *tt.want.response.Delta)
					}
					if tt.mtype == "gauge" {
						assert.Equal(t, tt.mvalue, *tt.want.response.Value)
					}
				}
			}
		})
	}
}

type MockDBClient struct {
	mock.Mock
}

func (c *MockDBClient) Ping(ctx context.Context) error {
	return nil
}

func (c *MockDBClient) Add(ctx context.Context, m metric.Metric) error {
	return nil
}

func (c *MockDBClient) Update(ctx context.Context, m metric.Metric, v interface{}) error {
	return nil
}

func (c *MockDBClient) Retrieve(ctx context.Context, m metric.MetricType, n string) (metric.Metric, error) {
	return nil, nil
}

func (c *MockDBClient) RetrieveAll(ctx context.Context) ([]metric.Metric, error) {
	return nil, nil
}

func NewMockDBClient() *MockDBClient {
	return &MockDBClient{}
}

func (c *MockDBClient) Close() error {
	return nil
}

func (c *MockDBClient) RunMigrations(ctx context.Context) error {
	return nil
}

func (c *MockDBClient) UpdateBatch(ctx context.Context, metrics *[]metric.Metric) error {
	return nil
}

func TestHTTPServer_PingHandler(t *testing.T) {

	addr := "http://localhost:8080"
	stor := NewMockDBClient()

	s := &HTTPServer{
		Address: addr,
		Storage: stor,
	}

	e := echo.New()

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(request, rec)

	if assert.NoError(t, s.PingHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "OK", rec.Body.String())
	}

}

func TestHTTPServer_UpdateJSONHandler(t *testing.T) {

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	ctr1 := &metric.Counter{Name: "ctr1"}
	g1 := &metric.Gauge{Name: "g1"}

	type want struct {
		name        string
		contentType string
		value       int64
		code        int
	}
	tests := []struct {
		name    string
		method  string
		payload any
		storage storage.Storage
		want    want
	}{

		{name: "Counter increment 1", method: http.MethodPost, payload: &dto.Metrics{ID: ctr1.GetName(), MType: string(ctr1.GetType()), Delta: int64Ptr(1)},
			storage: stor, want: want{code: 200, name: ctr1.Name, value: 1, contentType: "application/json"}},
		{name: "Counter increment 2", method: http.MethodPost, payload: &dto.Metrics{ID: ctr1.GetName(), MType: string(ctr1.GetType()), Delta: int64Ptr(2)},
			storage: stor, want: want{code: 200, name: ctr1.Name, value: 3, contentType: "application/json"}},
		{name: "Error1", method: http.MethodPost, payload: "123",
			storage: stor, want: want{code: 400, contentType: "text/plain; charset=UTF-8"}},
		{name: "Error 2", method: http.MethodPost, payload: &dto.Metrics{ID: ctr1.GetName(), MType: string(ctr1.GetType())},
			storage: stor, want: want{code: 400, contentType: "text/plain; charset=UTF-8"}},
		{name: "Error 3", method: http.MethodPost, payload: &dto.Metrics{ID: g1.GetName(), MType: string(g1.GetType())},
			storage: stor, want: want{code: 400, contentType: "text/plain; charset=UTF-8"}},
		{name: "Unknown metric type", method: http.MethodPost, payload: &dto.Metrics{ID: ctr1.GetName(), MType: "unknown", Delta: int64Ptr(2)},
			storage: stor, want: want{code: 400, contentType: "text/plain; charset=UTF-8"}},
		{name: "Bad storage", method: http.MethodPost, payload: &dto.Metrics{ID: ctr1.GetName(), MType: "unknown", Delta: int64Ptr(2)},
			storage: faultyStorage{}, want: want{code: 500, contentType: "text/plain; charset=UTF-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: addr,
				Storage: tt.storage,
			}

			e := echo.New()

			//payload := &dto.Metrics{ID: tt.mname, MType: string(tt.mtype), Delta: int64Ptr(tt.mvalue)}

			// Marshal the payload to JSON
			jsonData, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal JSON: %v", err)
			}

			request := httptest.NewRequest(tt.method, "/", bytes.NewBuffer(jsonData))
			request.Header.Set("Content-Type", "application/json")

			rec := httptest.NewRecorder()

			c := e.NewContext(request, rec)

			if assert.NoError(t, s.UpdateJSONHandler(c)) {
				assert.Equal(t, tt.want.code, rec.Code)
				assert.Equal(t, tt.want.contentType, rec.Header().Get("Content-Type"))

				if tt.want.code == http.StatusOK {
					var response dto.Metrics
					err := json.Unmarshal(rec.Body.Bytes(), &response)
					assert.NoError(t, err)
					assert.Equal(t, tt.want.value, *response.Delta)
				}
			}
		})
	}
}

func TestHTTPServer_MetricFromDto(t *testing.T) {
	type args struct {
		mDTO dto.Metrics
	}
	tests := []struct {
		args    args
		want    metric.Metric
		name    string
		wantErr bool
	}{
		{name: "ok", args: args{mDTO: dto.Metrics{ID: "m1", MType: "counter", Delta: int64Ptr(1)}}, want: &metric.Counter{Name: "m1", Value: 1}, wantErr: false},
		{name: "error_unknown_type", args: args{mDTO: dto.Metrics{ID: "m1", MType: "unknown", Delta: int64Ptr(1)}}, want: &metric.Counter{Name: "m1", Value: 1}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.MetricFromDto(tt.args.mDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.MetricFromDto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("HTTPServer.MetricFromDto() = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestHTTPServer_DTOFromMetric(t *testing.T) {
	tests := []struct {
		m       metric.Metric
		want    *dto.Metrics
		name    string
		wantErr bool
	}{
		{name: "ok", m: &metric.Counter{Name: "c1", Value: 1}, want: &dto.Metrics{ID: "c1", MType: "counter", Delta: int64Ptr(1)}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.DTOFromMetric(tt.m)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.DTOFromMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.DTOFromMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_UpdatesJSONHandler(t *testing.T) {
	stor := prepareTestSTorage()
	ctx := context.Background()
	tests := []struct {
		name     string
		payload  any
		storage  storage.Storage
		want     []metric.Metric
		wantErr  bool
		wantCode int
	}{
		{name: "ok",
			payload: []dto.Metrics{{ID: metric1.Name, MType: "counter", Delta: int64Ptr(2)}, {ID: metric2.Name, MType: "gauge", Value: float64Ptr(2.345)}},
			want:    []metric.Metric{&metric.Counter{Name: metric1.Name, Value: 3}, &metric.Gauge{Name: metric2.Name, Value: 2.345}},
			storage: stor, wantErr: false, wantCode: 200},
		{name: "error1", storage: stor, payload: []dto.Metrics{{ID: metric1.Name, MType: "unknown", Delta: int64Ptr(2)}}, want: []metric.Metric{}, wantErr: false, wantCode: 400},
		{name: "error2", storage: stor, payload: "wrong body", want: []metric.Metric{}, wantErr: false, wantCode: 400},
		{name: "error3", payload: []dto.Metrics{{ID: metric1.GetName(), MType: string(metric1.GetType()), Delta: int64Ptr(2)}}, want: []metric.Metric{}, wantErr: false, wantCode: 400, storage: faultyStorage{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			body, err := json.Marshal(tt.payload)
			require.NoError(t, err)

			// Create request and recorder
			req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()

			// Create context
			e := echo.New()
			c := e.NewContext(req, rec)

			s := prepareTestServer()
			s.Storage = tt.storage
			if err := s.UpdatesJSONHandler(c); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.UpdatesJSONHandler() error = %v, wantErr %v", err, tt.wantErr)
			}

			assert.Equal(t, tt.wantCode, rec.Code)

			if tt.wantCode == 200 {
				for _, mwant := range tt.want {
					m, err := usecase.RetrieveMetric(ctx, s.Storage, string(mwant.GetType()), mwant.GetName())
					if (err != nil) != tt.wantErr {
						t.Errorf("HTTPServer.TestHTTPServer_UpdatesJSONHandler() error = %v, wantErr %v", err, tt.wantErr)
						return
					}

					if m.GetValue() != mwant.GetValue() {
						t.Errorf("error value: %v %v ", m.GetValue(), mwant.GetValue())
						return
					}
				}
			}

		})
	}
}

func BenchmarkHTTPServer_UpdateHandler(b *testing.B) {
	s := prepareTestServer()
	e := echo.New()

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("type", "name", "value")
	c.SetParamValues("counter", "counter1", "1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.UpdateHandler(c)
	}
}

func BenchmarkHTTPServer_ValueHandler(b *testing.B) {
	s := prepareTestServer()
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	c.SetParamNames("type", "name")
	c.SetParamValues("counter", "counter1")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ValueHandler(c)
	}
}

func BenchmarkHTTPServer_UpdateJSONHandler(b *testing.B) {
	s := prepareTestServer()
	e := echo.New()

	metric := &dto.Metrics{
		ID:    "counter1",
		MType: "counter",
		Delta: int64Ptr(1),
	}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = s.UpdateJSONHandler(c)
	}

}

func BenchmarkHTTPServer_ValueJSONHandler(b *testing.B) {
	s := prepareTestServer()
	e := echo.New()

	metric := &dto.Metrics{
		ID:    "counter1",
		MType: "counter",
	}
	body, _ := json.Marshal(metric)

	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = s.ValueJSONHandler(c)
	}
}
