package httpserver

import (
	"net/http"

	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
)

type HTTPServer struct {
	Address string
	Storage storage.Storage
}

func NewHTTPServer(address string, storage storage.Storage) *HTTPServer {
	return &HTTPServer{Address: address, Storage: storage}
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

func (s *HTTPServer) NewMetricServerMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc(`/update/`, s.UpdateHandler)
	return mux
}

func (s *HTTPServer) Run() error {

	h := s.NewMetricServerMux()

	err := http.ListenAndServe(s.Address, h)

	if err != nil {
		panic(err)
	}

	return nil
}
