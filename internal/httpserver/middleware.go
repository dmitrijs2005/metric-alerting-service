package httpserver

import (
	"fmt"
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
