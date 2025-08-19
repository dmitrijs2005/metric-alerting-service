package httpserver

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeHP struct{ h http.Header }

func (f fakeHP) Header() http.Header { return f.h }

// helper: make a gzip pool like production
func newGzipPool() *sync.Pool {
	return &sync.Pool{
		New: func() any {
			return gzip.NewWriter(io.Discard)
		},
	}
}

func TestGzipWriter_Compressible_WritesGzip(t *testing.T) {
	rr := httptest.NewRecorder()
	h := http.Header{}
	h.Set("Content-Type", "application/json") // assumed compressible
	hp := fakeHP{h: h}

	pool := newGzipPool()

	gw, err := NewGzipWriter(rr, hp, pool)
	require.NoError(t, err)

	plain := []byte(`{"msg":"hello world"}`)
	n, err := gw.Write(plain)
	require.NoError(t, err)
	assert.Equal(t, len(plain), n)

	require.NoError(t, gw.Close())

	// Body must be gzip-compressed (should NOT equal plain)
	got := rr.Body.Bytes()
	assert.NotEqual(t, plain, got, "response must be compressed, not plain")

	// Decompress and compare
	zr, err := gzip.NewReader(bytes.NewReader(got))
	require.NoError(t, err)
	defer zr.Close()

	var decompressed bytes.Buffer
	_, err = io.Copy(&decompressed, zr)
	require.NoError(t, err)
	assert.Equal(t, plain, decompressed.Bytes())
}

func TestGzipWriter_Close_PutsWriterBackToPool(t *testing.T) {
	rr := httptest.NewRecorder()
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	hp := fakeHP{h: h}

	pool := newGzipPool()
	gw, err := NewGzipWriter(rr, hp, pool)
	require.NoError(t, err)

	// Capture the pointer identity of the gzip.Writer borrowed from the pool.
	gotPtr := gw.Writer

	require.NoError(t, gw.Close())

	// Next Get() should return the same instance (pool returns what we Put).
	next := pool.Get().(*gzip.Writer)
	// put it back to not starve the pool
	defer pool.Put(next)

	assert.Same(t, gotPtr, next, "gzip.Writer should be returned to the pool on Close")
}

func TestGzipWriter_NonCompressible_WritesPlain(t *testing.T) {
	rr := httptest.NewRecorder()
	h := http.Header{}
	h.Set("Content-Type", "image/png") // assumed NOT compressible
	hp := fakeHP{h: h}

	pool := newGzipPool()
	gw, err := NewGzipWriter(rr, hp, pool)
	require.NoError(t, err)

	plain := []byte("raw-binary-data")
	n, err := gw.Write(plain)
	require.NoError(t, err)
	assert.Equal(t, len(plain), n)

	// Close underlying gzip writer (even though we didn't use gzip path)
	require.NoError(t, gw.Close())

	got := rr.Body.Bytes()

	// Expect exactly the same bytes when not compressible.
	// If this assertion fails, Close might be appending gzip trailer bytes to the response.
	assert.Equal(t, plain, got, "non-compressible responses should be written as-is")

	// And attempting to treat it as gzip should fail.
	_, err = gzip.NewReader(bytes.NewReader(got))
	assert.Error(t, err, "plain response should not be recognized as gzip")
}

func TestGzipReader_ReadOK(t *testing.T) {
	// Prepare gzipped data
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	plain := []byte("hello gz")
	_, err := zw.Write(plain)
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	rc := io.NopCloser(bytes.NewReader(buf.Bytes()))

	gr, err := NewGzipReader(rc)
	require.NoError(t, err)

	var out bytes.Buffer
	_, err = io.Copy(&out, gr)
	require.NoError(t, err)
	assert.Equal(t, plain, out.Bytes())

	require.NoError(t, gr.Close())
}

func TestGzipReader_New_ErrorOnNonGzip(t *testing.T) {
	rc := io.NopCloser(bytes.NewReader([]byte("not-gzip")))
	gr, err := NewGzipReader(rc)
	assert.Nil(t, gr)
	require.Error(t, err)
}

type errOnCloseRC struct {
	io.Reader
}

func (e errOnCloseRC) Close() error { return errors.New("close-fail") }

func TestGzipReader_Close_PropagatesRCError(t *testing.T) {
	// Need a valid gzip stream to construct GzipReader
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	_, err := zw.Write([]byte("x"))
	require.NoError(t, err)
	require.NoError(t, zw.Close())

	rc := errOnCloseRC{Reader: bytes.NewReader(buf.Bytes())}
	gr, err := NewGzipReader(rc)
	require.NoError(t, err)

	// Close should return error from underlying rc.Close()
	err = gr.Close()
	require.Error(t, err)
	assert.Equal(t, "close-fail", err.Error())
}
