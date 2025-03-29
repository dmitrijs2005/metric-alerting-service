package httpserver

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// CreateBufferedLogger returns a Sugared Logger writing to a buffer
func CreateBufferedLogger(buf *bytes.Buffer) *zap.SugaredLogger {

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewProductionEncoderConfig()),
		zapcore.AddSync(buf), // Write logs to buffer
		zapcore.DebugLevel,   // Capture all logs (DEBUG and higher)
	)

	return zap.New(core).Sugar() // Return a SugaredLogger
}

func TestHTTPServer_RequestResponseInfoMiddleware(t *testing.T) {
	address := "http://localhost:8080"
	key := "secretkey"
	stor := memory.NewMemStorage()
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	buf := new(bytes.Buffer) // ✅ Initialize buffer
	log := CreateBufferedLogger(buf)

	tests := []struct {
		name   string
		method string
		url    string
		code   int
	}{
		{"Test 1 200", http.MethodGet, "/", 200},
		{"Test 2 405", http.MethodPost, "/", 405},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := NewHTTPServer(ctx, address, key, stor, log)
			e := s.ConfigureRoutes("../../web/template")

			request := httptest.NewRequest(tt.method, tt.url, nil)
			rec := httptest.NewRecorder()

			e.ServeHTTP(rec, request) // Ensures middleware runs
			log.Desugar().Sync()

			buffer := buf.String()

			assert.Equal(t, tt.code, rec.Code)

			assert.Contains(t, buffer, tt.method)
			assert.Contains(t, buffer, tt.url)

			assert.Contains(t, buffer, strconv.Itoa(tt.code))

		})
	}
}
