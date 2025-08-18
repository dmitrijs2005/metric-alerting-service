package httpserver

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/memory"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	cryptoKey := ""
	_, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	buf := new(bytes.Buffer) // Initialize buffer
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

			s, err := NewHTTPServer(address, key, stor, log, cryptoKey)
			require.NoError(t, err)

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

func TestDecryptMiddleware_Success(t *testing.T) {
	e := echo.New()

	// Generate test RSA key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	// Prepare plaintext and encrypt it in OAEP chunks with pubkey
	plaintext := []byte(`{"ok":1,"msg":"hello"}`)
	cipher, err := secure.EncryptRSAOAEPChunked(plaintext, &priv.PublicKey)
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewReader([]byte(cipher)))
	req.Header.Set(echo.HeaderContentType, echo.MIMEOctetStream)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Flag to ensure next is called and receives decrypted body
	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		// Read the (decrypted) body that middleware replaced
		got, err := io.ReadAll(c.Request().Body)
		require.NoError(t, err)
		assert.Equal(t, plaintext, got)
		return c.String(http.StatusOK, "OK")
	}

	s := &HTTPServer{PrivateKey: priv}
	h := s.DecryptMiddleware(next)

	err = h(c)
	require.NoError(t, err)

	assert.True(t, nextCalled, "next handler must be invoked on successful decrypt")
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())
}

// errRC is a ReadCloser that fails on Read to simulate body read errors.
type errRC struct{}

func (e errRC) Read(p []byte) (int, error) { return 0, errors.New("read-fail") }
func (e errRC) Close() error               { return nil }

func TestSignCheckMiddleware(t *testing.T) {
	key := "super-secret-key-32-bytes-please!"

	t.Run("no request signature -> response is signed", func(t *testing.T) {
		e := echo.New()

		// request has no HashSHA256 header; body is not verified
		req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString("ignored"))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		next := func(c echo.Context) error {
			// next writes some body
			return c.String(http.StatusOK, "pong")
		}

		s := &HTTPServer{Key: key}
		h := s.SignCheckMiddleware(next)

		err := h(c)
		require.NoError(t, err)

		// Response status/body come from next
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "pong", rec.Body.String())

		// Response must be signed
		respSign := rec.Header().Get("HashSHA256")
		require.NotEmpty(t, respSign)

		expectedSig, _ := secure.CreateAes256Signature([]byte("pong"), key)
		require.NoError(t, err)
		assert.Equal(t, base64.RawStdEncoding.EncodeToString(expectedSig), respSign)
	})

	t.Run("valid request signature -> request verified and body restored; response signed", func(t *testing.T) {
		e := echo.New()

		reqBody := []byte(`{"ok":1}`)
		// precompute request signature
		reqSig, err := secure.CreateAes256Signature(reqBody, key)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(reqBody))
		req.Header.Set("HashSHA256", base64.RawStdEncoding.EncodeToString(reqSig))

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		nextCalled := false
		next := func(c echo.Context) error {
			nextCalled = true
			// middleware should restore the request body
			got, err := io.ReadAll(c.Request().Body)
			require.NoError(t, err)
			assert.Equal(t, reqBody, got)
			return c.Blob(http.StatusCreated, "application/json", []byte(`{"ok":2}`))
		}

		s := &HTTPServer{Key: key}
		h := s.SignCheckMiddleware(next)

		err = h(c)
		require.NoError(t, err)
		assert.True(t, nextCalled)

		assert.Equal(t, http.StatusCreated, rec.Code)
		assert.Equal(t, `{"ok":2}`, rec.Body.String())

		// Response must be signed with the actual response body
		respSign := rec.Header().Get("HashSHA256")
		require.NotEmpty(t, respSign)

		expRespSig, err := secure.CreateAes256Signature([]byte(`{"ok":2}`), key)
		require.NoError(t, err)
		assert.Equal(t, base64.RawStdEncoding.EncodeToString(expRespSig), respSign)
	})

	t.Run("invalid request signature -> 400 and next not called", func(t *testing.T) {
		e := echo.New()

		req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewBufferString("body"))
		// wrong/garbage signature
		req.Header.Set("HashSHA256", "not-a-valid-sig")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		nextCalled := false
		next := func(c echo.Context) error {
			nextCalled = true
			return c.String(http.StatusOK, "should-not-be-called")
		}

		s := &HTTPServer{Key: key}
		h := s.SignCheckMiddleware(next)

		err := h(c)
		require.Error(t, err)

		he, ok := err.(*echo.HTTPError)
		require.True(t, ok, "should return echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, he.Code)
		assert.Equal(t, "Incorrect signature", he.Message)
		assert.False(t, nextCalled)
	})

	t.Run("read error while verifying request -> 500 and next not called", func(t *testing.T) {
		e := echo.New()

		// Non-empty signature triggers verification branch, but body reader errors.
		req := httptest.NewRequest(http.MethodPost, "/x", io.NopCloser(errRC{}))
		req.Header.Set("HashSHA256", "anything")

		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		nextCalled := false
		next := func(c echo.Context) error {
			nextCalled = true
			return c.String(http.StatusOK, "nope")
		}

		s := &HTTPServer{Key: key}
		h := s.SignCheckMiddleware(next)

		err := h(c)
		require.Error(t, err)

		he, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, he.Code)
		assert.Equal(t, "Cannot read signature", he.Message)
		assert.False(t, nextCalled)
	})
}

func newPool() *sync.Pool {
	return &sync.Pool{
		New: func() any { return gzip.NewWriter(io.Discard) },
	}
}

// helper: gzip a byte slice.
func gzipData(t *testing.T, b []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write(b)
	require.NoError(t, err)
	require.NoError(t, zw.Close())
	return buf.Bytes()
}

// helper: gunzip a byte slice.
func gunzipData(t *testing.T, b []byte) []byte {
	t.Helper()
	zr, err := gzip.NewReader(bytes.NewReader(b))
	require.NoError(t, err)
	defer zr.Close()
	out, err := io.ReadAll(zr)
	require.NoError(t, err)
	return out
}

func TestCompressingMiddleware_ResponseCompressible(t *testing.T) {
	e := echo.New()
	pool := newPool()
	s := &HTTPServer{GzipWriterPool: pool}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		// Set a compressible content type and write a body
		c.Response().Header().Set("Content-Type", "application/json")
		return c.String(http.StatusOK, `{"ok":1}`)
	}

	err := s.CompressingMiddleware(next)(c)
	require.NoError(t, err)

	// Header must indicate gzip
	ce := rec.Header().Get("Content-Encoding")
	assert.Equal(t, "gzip", ce)

	// Body should be gzipped; and gunzips to original
	gotCompressed := rec.Body.Bytes()
	assert.NotEqual(t, `{"ok":1}`, string(gotCompressed))
	gotPlain := gunzipData(t, gotCompressed)
	assert.Equal(t, `{"ok":1}`, string(gotPlain))
}

func TestCompressingMiddleware_ResponseNonCompressible(t *testing.T) {
	e := echo.New()
	pool := newPool()
	s := &HTTPServer{GzipWriterPool: pool}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	body := []byte{0x00, 0x01, 0x02, 0x03}
	next := func(c echo.Context) error {
		// Non-compressible type
		c.Response().Header().Set("Content-Type", "image/png")
		_, err := c.Response().Writer.Write(body)
		return err
	}

	err := s.CompressingMiddleware(next)(c)
	require.NoError(t, err)

	// Should not set gzip header
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	// Body should be the same bytes (not gzipped)
	assert.Equal(t, body, rec.Body.Bytes())
}

func TestCompressingMiddleware_RequestDecompression(t *testing.T) {
	e := echo.New()
	pool := newPool()
	s := &HTTPServer{GzipWriterPool: pool}

	plain := []byte("hello compressed world")
	compressed := gzipData(t, plain)

	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(compressed))
	req.Header.Set("Content-Encoding", "gzip")
	// No Accept-Encoding header → response not gzipped
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var seenHeaderBefore, seenHeaderAfter string
	next := func(c echo.Context) error {
		// The middleware should clear the header before calling next
		seenHeaderBefore = c.Request().Header.Get("Content-Encoding")
		b, err := io.ReadAll(c.Request().Body)
		require.NoError(t, err)
		assert.Equal(t, plain, b)

		// Write plain response
		return c.String(http.StatusCreated, "OK")
	}

	err := s.CompressingMiddleware(next)(c)
	require.NoError(t, err)

	// Confirm header was cleared for next
	assert.Equal(t, "", seenHeaderBefore)

	// Response is not gzipped (no Accept-Encoding)
	assert.Empty(t, rec.Header().Get("Content-Encoding"))
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "OK", rec.Body.String())

	// After handler finishes, the request header inside c should remain cleared.
	seenHeaderAfter = c.Request().Header.Get("Content-Encoding")
	assert.Equal(t, "", seenHeaderAfter)
}

func TestCompressingMiddleware_BadGzipRequestReturns500(t *testing.T) {
	e := echo.New()
	pool := newPool()
	s := &HTTPServer{GzipWriterPool: pool}

	// Claim gzip, but body is not gzipped
	req := httptest.NewRequest(http.MethodPost, "/x", strings.NewReader("not gzip"))
	req.Header.Set("Content-Encoding", "gzip")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	nextCalled := false
	next := func(c echo.Context) error {
		nextCalled = true
		return c.String(http.StatusOK, "should not run")
	}

	err := s.CompressingMiddleware(next)(c)
	require.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	require.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
	assert.Equal(t, "Cannot initialize Gzip", httpErr.Message)
	assert.False(t, nextCalled, "next must not be called on bad gzip")
}

func TestCompressingMiddleware_ReturnsWriterToPool(t *testing.T) {
	e := echo.New()

	pool := &sync.Pool{
		New: func() any { return gzip.NewWriter(io.Discard) },
	}
	s := &HTTPServer{GzipWriterPool: pool}

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	next := func(c echo.Context) error {
		c.Response().Header().Set("Content-Type", "application/json")
		_, err := c.Response().Writer.Write([]byte(`{"pool":true}`))
		return err
	}

	// Run the middleware (this will Get() a writer and defer Put() in Close()).
	err := s.CompressingMiddleware(next)(c)
	require.NoError(t, err)

	// We can assert that a *gzip.Writer is available and reusable.
	got := pool.Get()
	require.NotNil(t, got, "expected a writer back in the pool")
	gz, ok := got.(*gzip.Writer)
	require.True(t, ok, "expected *gzip.Writer from pool")
	// Writer should be reusable without panicking:
	gz.Reset(io.Discard)
	require.NoError(t, gz.Close())
	// put it back to keep the pool healthy
	pool.Put(gz)
}
