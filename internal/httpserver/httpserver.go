package httpserver

import (
	"compress/gzip"
	"context"
	"crypto/rsa"
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/dmitrijs2005/metric-alerting-service/internal/logger"
	"github.com/dmitrijs2005/metric-alerting-service/internal/secure"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage/file"
	"github.com/labstack/echo/v4"
)

type HTTPServer struct {
	GzipWriterPool *sync.Pool
	Storage        storage.Storage
	Saver          file.DumpSaver
	logger         logger.Logger
	Address        string
	Key            string
	StoreInterval  int
	Restore        bool
	PrivateKey     *rsa.PrivateKey
}

func NewHTTPServer(address string, key string, storage storage.Storage, logger logger.Logger, cryptoKey string) (*HTTPServer, error) {

	var privKey *rsa.PrivateKey
	var err error

	if cryptoKey != "" {
		privKey, err = secure.LoadRSAPrivateKeyFromPEM(cryptoKey)
		if err != nil {
			return nil, err
		}
	}

	pool := &sync.Pool{
		New: func() interface{} {
			w, err := gzip.NewWriterLevel(nil, gzip.BestSpeed)
			if err != nil {
				panic("failed to create gzip.Writer: " + err.Error())
			}
			return w
		},
	}

	return &HTTPServer{Address: address, Key: key, Storage: storage, logger: logger, GzipWriterPool: pool, PrivateKey: privKey}, nil
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
	e.POST("/update/", s.UpdateJSONHandler, s.middlewareIf(s.PrivateKey != nil, s.DecryptMiddleware)...)
	e.POST("/updates/", s.UpdatesJSONHandler, s.middlewareIf(s.PrivateKey != nil, s.DecryptMiddleware)...)
	e.POST("/update/:type/:name/:value", s.UpdateHandler)
	e.GET("/value/:type/:name", s.ValueHandler)
	e.GET("/ping", s.PingHandler)
	e.GET("/", s.ListHandler)

	e.Renderer = t
	return e
}

func (s *HTTPServer) middlewareIf(condtion bool, mw ...echo.MiddlewareFunc) []echo.MiddlewareFunc {
	if condtion {
		return mw
	}
	return nil
}

func (s *HTTPServer) Run(ctx context.Context) error {

	e := s.ConfigureRoutes("web/template")

	server := &http.Server{
		Addr:    s.Address,
		Handler: e,
	}

	go func() {
		<-ctx.Done()
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
