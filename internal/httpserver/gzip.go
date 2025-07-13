package httpserver

import (
	"compress/gzip"
	"io"
	"net/http"
	"sync"
)

var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

type HeaderProvider interface {
	Header() http.Header
}

type GzipWriter struct {
	http.ResponseWriter
	Writer         *gzip.Writer
	HeaderProvider HeaderProvider
}

func (w *GzipWriter) Write(b []byte) (int, error) {

	// if content type not compressable, do nothing
	ct := w.HeaderProvider.Header().Get("Content-Type")

	if ContentTypeIsCompressable(ct) {
		return w.Writer.Write(b)
	}

	// write uncompressed response
	return w.ResponseWriter.Write(b)

}

func (w *GzipWriter) Close() error {
	defer gzipWriterPool.Put(w.Writer)
	return w.Writer.Close()
}

func NewGzipWriter(w http.ResponseWriter, hp HeaderProvider) (*GzipWriter, error) {

	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(w)

	return &GzipWriter{
		Writer:         gz,
		ResponseWriter: w,
		HeaderProvider: hp,
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
