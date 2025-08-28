package httpserver

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) SignCheckMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		req := c.Request()

		// checking if signature is received
		sign := req.Header.Get("HashSHA256")

		if sign != "" {

			body, err := io.ReadAll(req.Body)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Cannot read signature")
			}

			actualSign, err := secure.CreateAes256Signature(body, s.Key)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Cannot create signature")
			}

			if sign != base64.RawStdEncoding.EncodeToString(actualSign) {
				return echo.NewHTTPError(http.StatusBadRequest, "Incorrect signature")
			}

			// restoring requst body
			c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

		}

		// if key is specified we should read the output and sign it

		// buffer for output
		resBody := new(bytes.Buffer)

		// substitute response writer
		w := c.Response().Writer

		rec := NewResponseRecorder(w, resBody)
		c.Response().Writer = rec

		if err := next(c); err != nil {
			c.Error(err)
		}

		// signing the response
		body := resBody.Bytes()
		responseSign, err := secure.CreateAes256Signature(body, s.Key)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot create signature")
		}

		c.Response().Writer = w
		c.Response().WriteHeader(rec.status)

		hdr := c.Response().Header()
		hdr.Set("HashSHA256", base64.RawStdEncoding.EncodeToString(responseSign))

		return nil
	}
}

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
		if strings.Contains(req.Header.Get("Accept-Encoding"), "gzip") {

			// создаём gzip.Writer поверх текущего w
			gw, err := NewGzipWriter(w, resp, s.GzipWriterPool)

			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "Cannot initialize Gzip")
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
				return echo.NewHTTPError(http.StatusInternalServerError, "Cannot initialize Gzip")
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

func (s *HTTPServer) DecryptMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		req := c.Request()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot read request body")
		}

		qw, err := secure.DecryptRSAOAEPChunked(string(body), s.PrivateKey)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Cannot decrypt body")
		}

		c.Request().Body = io.NopCloser(bytes.NewBuffer(qw))

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil

	}
}

func (s *HTTPServer) CheckTrustedSubnetMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {

		req := c.Request()

		realIP := req.Header.Get("X-Real-IP")

		if realIP == "" {
			return echo.NewHTTPError(http.StatusForbidden, "Cannot find real IP header")
		}

		if !s.TrustedSubnet.Contains(net.ParseIP(realIP)) {
			return echo.NewHTTPError(http.StatusForbidden, "IP address is not in trusted subnet")
		}

		if err := next(c); err != nil {
			c.Error(err)
		}

		return nil

	}
}
