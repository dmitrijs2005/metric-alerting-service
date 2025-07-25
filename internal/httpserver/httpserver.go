package httpserver

import (
	"compress/gzip"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/file"
	"github.com/labstack/echo/v4"
)

type HTTPServer struct {
	ctx            context.Context
	Address        string
	StoreInterval  int
	Restore        bool
	Storage        storage.Storage
	Saver          file.DumpSaver
	Key            string
	GzipWriterPool *sync.Pool
	logger         logger.Logger
}

func NewHTTPServer(ctx context.Context, address string, key string, storage storage.Storage, logger logger.Logger) *HTTPServer {

	pool := &sync.Pool{
		New: func() interface{} {
			w, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
			if err != nil {
				panic("failed to create gzip.Writer: " + err.Error())
			}
			return w
		},
	}

	return &HTTPServer{ctx: ctx, Address: address, Key: key, Storage: storage, logger: logger, GzipWriterPool: pool}
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
	if s.Key != "" {
		e.Use(s.SignCheckMiddleware)
	}

	e.POST("/value/", s.ValueJSONHandler)
	e.POST("/update/", s.UpdateJSONHandler)
	e.POST("/updates/", s.UpdatesJSONHandler)
	e.POST("/update/:type/:name/:value", s.UpdateHandler)
	e.GET("/value/:type/:name", s.ValueHandler)
	e.GET("/ping", s.PingHandler)
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

	s.logger.Infow("Starting HTTP server", "address", server.Addr)

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}

	return nil
}
