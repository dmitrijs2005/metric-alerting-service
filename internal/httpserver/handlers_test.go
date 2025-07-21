package httpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/db"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	metric1 = &metric.Counter{Name: "counter1", Value: 1}
	metric2 = &metric.Gauge{Name: "gauge1", Value: 1.234}
)

func prepareTestServer() *HTTPServer {
	a := "http://localhost:8080"
	s := memory.NewMemStorage()

	s.Data["counter|counter1"] = metric1
	s.Data["gauge|gauge1"] = metric2

	return &HTTPServer{
		Address: a,
		Storage: s,
	}

}

func TestHTTPServer_UpdateHandler(t *testing.T) {

	a := "http://localhost:8080"
	s := memory.NewMemStorage()

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		want   want
		name   string
		method string
		url    string
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
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		want   want
		name   string
		method string
		url    string
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

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		want   want
		name   string
		method string
		url    string
	}{

		{name: "Counter OK", method: http.MethodGet, url: "/value/counter/counter1", want: want{code: 200, response: fmt.Sprintf("%v", metric1.GetValue()), contentType: "text/plain; charset=UTF-8"}},
		{name: "Gauge OK", method: http.MethodGet, url: "/value/gauge/gauge1", want: want{code: 200, response: fmt.Sprintf("%v", metric2.GetValue()), contentType: "text/plain; charset=UTF-8"}},
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

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	s := &HTTPServer{
		Address: addr,
		Storage: stor,
	}

	e := echo.New()

	// Load templates
	r := &Template{
		templates: template.Must(template.ParseGlob("../../web/template/*.html")),
	}
	e.Renderer = r

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(request, rec)

	if assert.NoError(t, s.ListHandler(c)) {
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, strings.Contains(rec.Body.String(), metric1.Name))
		assert.True(t, strings.Contains(rec.Body.String(), metric1.Name))

		assert.Equal(t, "text/html; charset=UTF-8", rec.Header().Get("Content-Type"))
	}

}

func TestHTTPServer_ValueJSONHandler(t *testing.T) {

	addr := "http://localhost:8080"
	stor := memory.NewMemStorage()

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	stor.Data["counter|counter1"] = metric1
	stor.Data["gauge|gauge1"] = metric2

	type want struct {
		code        int
		response    *dto.Metrics
		contentType string
	}
	tests := []struct {
		want   want
		name   string
		method string
		mtype  string
		mname  string
		mvalue interface{}
	}{

		{name: "Counter OK", method: http.MethodPost, mtype: "counter", mname: metric1.Name, mvalue: metric1.Value, want: want{code: 200, response: &dto.Metrics{ID: metric1.Name, Delta: &metric1.Value, MType: "counter"}, contentType: "application/json"}},
		{name: "Gauge OK", method: http.MethodGet, mtype: "gauge", mname: metric2.Name, mvalue: metric2.Value, want: want{code: 200, response: &dto.Metrics{ID: metric2.Name, Value: &metric2.Value, MType: "gauge"}, contentType: "application/json"}},
		{name: "Unnown metric", method: http.MethodPost, mtype: "unknown", mname: "", mvalue: 0, want: want{code: 404, response: nil, contentType: "text/plain; charset=UTF-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: addr,
				Storage: stor,
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

func TestHTTPServer_PingHandler(t *testing.T) {

	addr := "http://localhost:8080"
	stor := db.NewMockDBClient()

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

	type want struct {
		code        int
		name        string
		value       int64
		contentType string
	}
	tests := []struct {
		want   want
		name   string
		method string
		mtype  metric.MetricType
		mname  string
		mvalue int64
	}{

		{name: "Counter increment 1", method: http.MethodPost, mtype: ctr1.GetType(), mname: ctr1.GetName(), mvalue: 1, want: want{code: 200, name: ctr1.Name, value: 1, contentType: "application/json"}},
		{name: "Counter increment 2", method: http.MethodPost, mtype: ctr1.GetType(), mname: ctr1.GetName(), mvalue: 2, want: want{code: 200, name: ctr1.Name, value: 3, contentType: "application/json"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: addr,
				Storage: stor,
			}

			e := echo.New()

			payload := &dto.Metrics{ID: tt.mname, MType: string(tt.mtype), Delta: int64Ptr(tt.mvalue)}

			// Marshal the payload to JSON
			jsonData, err := json.Marshal(payload)
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

				var response dto.Metrics
				err := json.Unmarshal(rec.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, tt.want.value, *response.Delta)

			}
		})
	}
}

func TestHTTPServer_retrieveMetric(t *testing.T) {

	ctx := context.Background()

	type args struct {
		metricType string
		metricName string
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{name: "Counter OK", args: args{string(metric.MetricTypeCounter), "counter1"}, want: &metric.Counter{Name: "counter1", Value: 1}},
		{name: "Gauge OK", args: args{string(metric.MetricTypeGauge), "gauge1"}, want: &metric.Gauge{Name: "gauge1", Value: 1.234}},
		{name: "Unknown", args: args{"unknown", "u1"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()

			got, err := s.retrieveMetric(ctx, tt.args.metricType, tt.args.metricName)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.retrieveMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.retrieveMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_updateMetric(t *testing.T) {

	a := "http://localhost:8080"
	s := memory.NewMemStorage()

	metric1 := &metric.Counter{Name: "counter1", Value: 1}
	metric2 := &metric.Gauge{Name: "gauge1", Value: 1.234}

	s.Data["counter|counter1"] = metric1
	s.Data["gauge|gauge1"] = metric2

	ctx := context.Background()

	type args struct {
		m           metric.Metric
		metricValue any
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		wantValue any
	}{
		{name: "Counter OK", args: args{m: metric1, metricValue: int64(2)}, wantErr: false, wantValue: int64(3)},
		{name: "Gauge OK", args: args{m: metric2, metricValue: float64(2.345)}, wantErr: false, wantValue: float64(2.345)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: a,
				Storage: s,
			}
			if err := s.updateMetric(ctx, tt.args.m, tt.args.metricValue); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			m, err := s.retrieveMetric(ctx, string(tt.args.m.GetType()), tt.args.m.GetName())
			if err != nil {
				t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if m.GetValue() != tt.wantValue {
				t.Errorf("HTTPServer.updateMetric() error = wrong value, %v, wanted: %v ", m.GetValue(), tt.wantValue)
			}

		})
	}
}

func TestHTTPServer_addNewMetric(t *testing.T) {

	ctx := context.Background()

	type args struct {
		metricType  string
		metricName  string
		metricValue any
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{name: "Counter OK", args: args{"counter", "c2", int64(1)}, wantErr: false, want: &metric.Counter{Name: "c2", Value: int64(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.addNewMetric(ctx, tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.addNewMetric() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.addNewMetric() = %v, want %v", got, tt.want)
			}
			m, err := s.retrieveMetric(ctx, tt.args.metricType, tt.args.metricName)
			if err != nil {
				t.Errorf("HTTPServer.updateMetric() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(m, tt.want) {
				t.Errorf("HTTPServer.addNewMetric() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_newMetricWithValue(t *testing.T) {
	type args struct {
		metricType  string
		metricName  string
		metricValue any
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{name: "Counter", args: args{"counter", "c1", int64(1)}, wantErr: false, want: &metric.Counter{Name: "c1", Value: int64(1)}},
		{name: "Gauge", args: args{"gauge", "g1", float64(1.234)}, wantErr: false, want: &metric.Gauge{Name: "g1", Value: float64(1.234)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.newMetricWithValue(tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.newMetricWithValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.newMetricWithValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_updateMetricByValue(t *testing.T) {
	ctx := context.Background()
	type args struct {
		metricType  string
		metricName  string
		metricValue interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{name: "Counter", args: args{"counter", "c1", int64(1)}, wantErr: false, want: &metric.Counter{Name: "c1", Value: int64(1)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.updateMetricByValue(ctx, tt.args.metricType, tt.args.metricName, tt.args.metricValue)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.updateMetricByValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.updateMetricByValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_fillValue(t *testing.T) {
	type args struct {
		m metric.Metric
		r *dto.Metrics
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "OK", args: args{m: &metric.Counter{Name: "c1", Value: int64(1)}, r: &dto.Metrics{}}, wantErr: false},
		{name: "Error", args: args{m: &metric.Gauge{Name: "g1", Value: float64(1.234)}, r: &dto.Metrics{}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			if err := s.fillValue(tt.args.m, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.fillValue() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPServer_MetricFromDto(t *testing.T) {
	type args struct {
		mDTO dto.Metrics
	}
	tests := []struct {
		name    string
		args    args
		want    metric.Metric
		wantErr bool
	}{
		{"ok", args{mDTO: dto.Metrics{ID: "m1", MType: "counter", Delta: int64Ptr(1)}}, &metric.Counter{Name: "m1", Value: 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := prepareTestServer()
			got, err := s.MetricFromDto(tt.args.mDTO)
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.MetricFromDto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("HTTPServer.MetricFromDto() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPServer_DTOFromMetric(t *testing.T) {
	tests := []struct {
		name    string
		m       metric.Metric
		want    *dto.Metrics
		wantErr bool
	}{
		{"ok", &metric.Counter{Name: "c1", Value: 1}, &dto.Metrics{ID: "c1", MType: "counter", Delta: int64Ptr(1)}, false},
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
	ctx := context.Background()
	tests := []struct {
		name    string
		payload []dto.Metrics
		want    []metric.Metric
		wantErr bool
	}{
		{"ok",
			[]dto.Metrics{{ID: metric1.Name, MType: "counter", Delta: int64Ptr(2)}, {ID: metric2.Name, MType: "gauge", Value: float64Ptr(2.345)}},
			[]metric.Metric{&metric.Counter{Name: metric1.Name, Value: 3}, &metric.Gauge{Name: metric2.Name, Value: 2.345}},
			false},
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
			if err := s.UpdatesJSONHandler(c); (err != nil) != tt.wantErr {
				t.Errorf("HTTPServer.UpdatesJSONHandler() error = %v, wantErr %v", err, tt.wantErr)
			}

			for _, m := range tt.want {
				m, err := s.retrieveMetric(ctx, string(m.GetType()), m.GetName())
				if (err != nil) != tt.wantErr {
					t.Errorf("HTTPServer.TestHTTPServer_UpdatesJSONHandler() error = %v, wantErr %v", err, tt.wantErr)
					return
				}

				if m.GetValue() != m.GetValue() {
					t.Errorf("error value: %v", m.GetValue())
					return
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
