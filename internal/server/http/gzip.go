package http

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

type HeaderProvider interface {
	Header() http.Header
}

type GzipWriter struct {
	http.ResponseWriter
	Writer         *gzip.Writer
	HeaderProvider HeaderProvider
	GzipWriterPool *sync.Pool
	used           bool
}

func (w *GzipWriter) Write(b []byte) (int, error) {

	// if content type not compressable, do nothing
	ct := w.HeaderProvider.Header().Get("Content-Type")

	if ContentTypeIsCompressable(ct) {
		w.used = true
		return w.Writer.Write(b)
	}

	// write uncompressed response
	return w.ResponseWriter.Write(b)

}

func (w *GzipWriter) Close() error {
	defer w.GzipWriterPool.Put(w.Writer)
	if w.used {
		return w.Writer.Close()
	}

	w.Writer.Reset(io.Discard)
	return nil
}

func NewGzipWriter(w http.ResponseWriter, hp HeaderProvider, pool *sync.Pool) (*GzipWriter, error) {

	gz := pool.Get().(*gzip.Writer)
	gz.Reset(w)

	return &GzipWriter{
		Writer:         gz,
		ResponseWriter: w,
		HeaderProvider: hp,
		GzipWriterPool: pool,
	}, nil

}

type GzipReader struct {
	rc io.ReadCloser
	zr *gzip.Reader
}

func NewGzipReader(rc io.ReadCloser) (*GzipReader, error) {
	zr, err := gzip.NewReader(rc)
	if err != nil {
		return nil, err
	}
	return &GzipReader{rc: rc, zr: zr}, nil
}

func (r *GzipReader) Close() error {
	if err := r.rc.Close(); err != nil {
		return err
	}
	return r.zr.Close()
}

func (r *GzipReader) Read(p []byte) (n int, err error) {
	return r.zr.Read(p)
}
