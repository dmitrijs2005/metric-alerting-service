package httpserver

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

func TestHTTPServer_UpdateHandler(t *testing.T) {

	a := "http://localhost:8080"
	s := storage.NewMemStorage()

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
	s := storage.NewMemStorage()

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
