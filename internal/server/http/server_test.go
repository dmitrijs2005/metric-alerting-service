package http

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestEchoGracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	log := logger.GetLogger()
	stor := memory.NewMemStorage()

	s := &HTTPServer{Address: ":8080", Storage: stor, logger: log}

	e := echo.New()
	e.GET("/slow", func(c echo.Context) error {

		time.Sleep(500 * time.Millisecond) // slow request
		return c.String(http.StatusOK, "OK")
	})

	// starting the server
	go func() {
		if err := s.Run(ctx, e); err != nil && err != http.ErrServerClosed {
			t.Errorf("server error: %v", err)
		}
	}()

	// wait a little for server to start
	time.Sleep(50 * time.Millisecond)

	client := &http.Client{}
	doneCh := make(chan string, 1)

	// starting slow request
	go func() {
		resp, err := client.Get("http://localhost:8080/slow")
		if err != nil {
			t.Errorf("request failed: %v", err)
			return
		}
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		doneCh <- string(body)
	}()

	// waiting for request processing to start
	time.Sleep(100 * time.Millisecond)
	cancel() // shutdown

	// checking if request is processed
	select {
	case res := <-doneCh:
		if res != "OK" {
			t.Errorf("expected 'OK', got %q", res)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("request did not complete gracefully")
	}
}

func TestHTTPServer_getUpdateMiddlewares(t *testing.T) {
	s := &HTTPServer{}

	t.Run("no middlewares", func(t *testing.T) {
		mws := s.getUpdateMiddlewares()
		if len(mws) != 0 {
			t.Errorf("expected 0, got %d", len(mws))
		}
	})

	t.Run("with private key", func(t *testing.T) {
		s.PrivateKey = &rsa.PrivateKey{} // заглушка, можно сгенерить реальный ключ
		defer func() { s.PrivateKey = nil }()

		mws := s.getUpdateMiddlewares()
		if len(mws) != 1 {
			t.Errorf("expected 1, got %d", len(mws))
		}
	})

	t.Run("with trusted subnet", func(t *testing.T) {
		_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
		s.PrivateKey = nil
		s.TrustedSubnet = cidr
		defer func() { s.TrustedSubnet = nil }()

		mws := s.getUpdateMiddlewares()
		if len(mws) != 1 {
			t.Errorf("expected 1, got %d", len(mws))
		}
	})

	t.Run("with both", func(t *testing.T) {
		_, cidr, _ := net.ParseCIDR("192.168.0.0/24")
		s.PrivateKey = &rsa.PrivateKey{}
		s.TrustedSubnet = cidr

		mws := s.getUpdateMiddlewares()
		if len(mws) != 2 {
			t.Errorf("expected 2, got %d", len(mws))
		}
	})
}

func TestNewHTTPServer(t *testing.T) {
	st := memory.NewMemStorage()
	l := logger.GetLogger()

	t.Run("no cryptoKey, no subnet", func(t *testing.T) {
		srv, err := NewHTTPServer(":8080", "key", st, l, "", "")
		require.NoError(t, err)
		require.NotNil(t, srv)
		require.Nil(t, srv.PrivateKey)
		require.Nil(t, srv.TrustedSubnet)
	})

	t.Run("with valid cryptoKey file", func(t *testing.T) {
		privKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err)

		der := x509.MarshalPKCS1PrivateKey(privKey)
		block := &pem.Block{Type: "RSA PRIVATE KEY", Bytes: der}
		pemBytes := pem.EncodeToMemory(block)

		tmpFile, err := os.CreateTemp("", "testkey*.pem")
		require.NoError(t, err)
		defer os.Remove(tmpFile.Name())

		_, err = tmpFile.Write(pemBytes)
		require.NoError(t, err)
		tmpFile.Close()

		srv, err := NewHTTPServer(":8080", "key", st, l, tmpFile.Name(), "")
		require.NoError(t, err)
		require.NotNil(t, srv.PrivateKey)
	})

	t.Run("with invalid cryptoKey path", func(t *testing.T) {
		_, err := NewHTTPServer(":8080", "key", st, l, "/non/existing/path.pem", "")
		require.Error(t, err)
	})

	t.Run("with valid trustedSubnet", func(t *testing.T) {
		srv, err := NewHTTPServer(":8080", "key", st, l, "", "192.168.0.0/24")
		require.NoError(t, err)
		require.NotNil(t, srv.TrustedSubnet)
		require.Equal(t, "192.168.0.0/24", srv.TrustedSubnet.String())
	})

	t.Run("with invalid trustedSubnet", func(t *testing.T) {
		_, err := NewHTTPServer(":8080", "key", st, l, "", "not_a_cidr")
		require.Error(t, err)
	})
}
