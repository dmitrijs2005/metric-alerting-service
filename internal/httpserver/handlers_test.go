package httpserver

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
		{name: "Get - Method Not Allowed", method: http.MethodGet, url: "/update", want: want{code: 405, response: "Method Not Allowed\n", contentType: "text/plain; charset=utf-8"}},
		{name: "No metric name - Not Found", method: http.MethodPost, url: "/update/counter", want: want{code: 404, response: "Not Found\n", contentType: "text/plain; charset=utf-8"}},
		{name: "Invalid metric type - Bad Request", method: http.MethodPost, url: "/update/unknown/u1/1", want: want{code: 400, response: "invalid metric type\n", contentType: "text/plain; charset=utf-8"}},
		{name: "Counter OK", method: http.MethodPost, url: "/update/counter/counter1/1", want: want{code: 200, response: "OK", contentType: "text/plain; charset=utf-8"}},
		{name: "Counter Bad request", method: http.MethodPost, url: "/update/counter/counter1/a", want: want{code: 400, response: "invalid metric value\n", contentType: "text/plain; charset=utf-8"}},
		{name: "Gauge OK", method: http.MethodPost, url: "/update/gauge/gauge1/1.1", want: want{code: 200, response: "OK", contentType: "text/plain; charset=utf-8"}},
		{name: "Gauge Bad request", method: http.MethodPost, url: "/update/gauge/gauge1/a", want: want{code: 400, response: "invalid metric value\n", contentType: "text/plain; charset=utf-8"}},
		{name: "Gauge Bad request", method: http.MethodPost, url: "/update/gauge/gauge1/1,2", want: want{code: 400, response: "invalid metric value\n", contentType: "text/plain; charset=utf-8"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &HTTPServer{
				Address: a,
				Storage: s,
			}

			request := httptest.NewRequest(tt.method, tt.url, nil)
			w := httptest.NewRecorder()

			s.UpdateHandler(w, request)

			res := w.Result()

			// получаем и проверяем тело запроса
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)

			// проверяем код ответа
			assert.Equal(t, tt.want.code, res.StatusCode)
			assert.Equal(t, tt.want.response, string(resBody))
			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

		})
	}
}
