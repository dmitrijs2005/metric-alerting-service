package httpserver

import (
	"fmt"
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

		s.logger.Info(fmt.Sprintf("%s %s %s %d %d", req.URL, req.Method, timeTaken, resp.Status, resp.Size))

		return nil
	}
}

func (s *HTTPServer) CompressingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		req := c.Request()
		resp := c.Response()

		w := resp.Writer

		// if gzip is not supported, do nothing
		//&& (req.URL.Path == "/update/" || req.URL.Path == "/")
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {

			// создаём gzip.Writer поверх текущего w
			gw, err := NewGzipWriter(w, resp)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return nil
			}

			resp.Writer = gw
			defer gw.Close()

			resp.Before(func() {

				resp := c.Response()
				ct := resp.Header().Get("Content-Type")

				if resp.Status < 300 && ContentTypeIsCompressable(ct) {
					hdr := c.Response().Header()
					hdr.Set("Content-Encoding", "gzip")
				}
			})
		}

		// проверяем, что клиент отправил серверу сжатые данные в формате gzip
		ce := req.Header.Get("Content-Encoding")
		sendsGzip := strings.Contains(ce, "gzip")

		if sendsGzip {

			// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
			r, err := NewGzipReader(req.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return nil
			}

			req.Body = r
			req.Header.Del("Content-Encoding")

			defer r.Close()
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil
	}
}
