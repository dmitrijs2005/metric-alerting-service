package httpserver

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) RequestResponseInfoMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		t := time.Now()

		if err := next(c); err != nil {
			c.Error(err)
		}

		timeTaken := time.Since(t)

		req := c.Request()
		resp := c.Response()

		s.Logger.Info(fmt.Sprintf("%s %s %s %d %d", req.URL, req.Method, timeTaken, resp.Status, resp.Size))

		return nil
	}
}

type gzipWriter struct {
	http.ResponseWriter
	Writer   io.Writer
	Response *echo.Response
}

func (w gzipWriter) Write(b []byte) (int, error) {
	// w.Writer будет отвечать за gzip-сжатие, поэтому пишем в него
	// if content type not compressable, do nothing

	ct := w.Response.Header().Get("Content-Type")

	if ct == "application/json" || strings.HasPrefix(ct, "text/html") {
		// write compressed response
		w.Header().Set("Content-Encoding", "gzip")
		return w.Writer.Write(b)
	}

	// write uncompressed response
	return w.ResponseWriter.Write(b)

}

func (s *HTTPServer) CompressingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		req := c.Request()
		resp := c.Response()

		// if gzip is not supported, do nothing
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {
			w := resp.Writer

			// создаём gzip.Writer поверх текущего w
			gz, err := gzip.NewWriterLevel(w, gzip.BestSpeed)
			if err != nil {
				io.WriteString(w, err.Error())
				return nil
			}
			defer gz.Close()

			resp.Writer = &gzipWriter{
				Writer:         gz,
				ResponseWriter: w,
				Response:       resp,
			}

		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
