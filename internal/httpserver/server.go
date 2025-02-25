package httpserver

import (
	"context"
	"fmt"
	"html/template"
	"net/http"

	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type HTTPServer struct {
	ctx           context.Context
	Address       string
	StoreInterval int
	Restore       bool
	Storage       storage.Storage
	Saver         storage.DumpSaver
	logger        *zap.SugaredLogger
}

func NewHTTPServer(ctx context.Context, address string, storage storage.Storage, logger *zap.SugaredLogger) *HTTPServer {

	return &HTTPServer{ctx: ctx, Address: address, Storage: storage, logger: logger}
}

func (s *HTTPServer) ConfigureRoutes(templatePath string) *echo.Echo {

	// Load templates
	t := &Template{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", templatePath))),
	}

	// Echo instance
	e := echo.New()

	e.Use(s.RequestResponseInfoMiddleware)
	e.Use(s.CompressingMiddleware)

	e.POST("/value/", s.ValueJSONHandler)
	e.POST("/update/", s.UpdateJSONHandler)
	e.POST("/update/:type/:name/:value", s.UpdateHandler)
	e.GET("/value/:type/:name", s.ValueHandler)
	e.GET("/", s.ListHandler)

	e.Renderer = t
	return e
}

func (s *HTTPServer) Run() error {

	e := s.ConfigureRoutes("web/template")

	server := &http.Server{
		Addr:    s.Address,
		Handler: e,
	}

	go func() {
		<-s.ctx.Done()
		s.logger.Info("Stopping HTTP server...")

		if err := server.Shutdown(context.Background()); err != nil {
			s.logger.Error("HTTP server shutdown error", "err", err)
		}
	}()

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		fmt.Println("SERVECLOSED")
		return err
	}

	return nil
}
