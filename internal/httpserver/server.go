package httpserver

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	log "github.com/sirupsen/logrus"
)

type HTTPServer struct {
	Address string
	Storage storage.Storage
	Logger  *zap.SugaredLogger
}

func NewHTTPServer(address string, storage storage.Storage, logger *zap.SugaredLogger) *HTTPServer {
	return &HTTPServer{Address: address, Storage: storage, Logger: logger}
}

// Сервер должен быть доступен по адресу http://localhost:8080, а также:
// Принимать и хранить произвольные метрики двух типов:
// Тип gauge, float64 — новое значение должно замещать предыдущее.
// Тип counter, int64 — новое значение должно добавляться к предыдущему, если какое-то значение уже было известно серверу.

// Принимать метрики по протоколу HTTP методом POST.
// Принимать данные в формате http://<АДРЕС_СЕРВЕРА>/update/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ>/<ЗНАЧЕНИЕ_МЕТРИКИ>, Content-Type: text/plain.

// При успешном приёме возвращать http.StatusOK.
// При попытке передать запрос без имени метрики возвращать http.StatusNotFound.
// При попытке передать запрос с некорректным типом метрики или значением возвращать http.StatusBadRequest.

// Доработайте сервер так, чтобы в ответ на запрос GET http://<АДРЕС_СЕРВЕРА>/value/<ТИП_МЕТРИКИ>/<ИМЯ_МЕТРИКИ> он возвращал аккумулированное значение метрики в текстовом виде со статусом http.StatusOK.
// При попытке запроса неизвестной метрики сервер должен возвращать http.StatusNotFound.
// По запросу GET http://<АДРЕС_СЕРВЕРА>/ сервер должен отдавать HTML-страницу со списком имён и значений всех известных ему на текущий момент метрик.
// Хендлеры должны взаимодействовать с экземпляром MemStorage при помощи соответствующих интерфейсных методов.

func (s *HTTPServer) ConfigureRoutes(templatePath string) *echo.Echo {

	// Load templates
	t := &Template{
		templates: template.Must(template.ParseGlob(fmt.Sprintf("%s/*.html", templatePath))),
	}

	// Echo instance
	e := echo.New()

	e.Use(s.RequestResponseInfoMiddleware)

	e.POST("/update/:type/:name/:value", s.UpdateHandler)
	e.GET("/value/:type/:name", s.ValueHandler)
	e.GET("/", s.ListHandler)

	e.Renderer = t
	return e
}

func (s *HTTPServer) Run() error {

	e := s.ConfigureRoutes("web/template")

	server := http.Server{
		Addr:    s.Address,
		Handler: e,
	}
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}

	return nil
}
