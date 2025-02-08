package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) UpdateHandler(c echo.Context) error {

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	m, err := s.Storage.Retrieve(metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		m, err = metric.NewMetric(metric.MetricType(metricType), metricName)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		err = s.Storage.Add(m)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	err = m.Update(metricValue)
	if err != nil {

		if errors.Is(err, metric.ErrorInvalidMetricValue) {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	// if everything is correct and metric was saved
	return c.String(http.StatusOK, "OK")
}

func (s *HTTPServer) ValueHandler(c echo.Context) error {

	metricType := c.Param("type")
	metricName := c.Param("name")

	m, err := s.Storage.Retrieve(metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("%v", m.GetValue()))
}

func (s *HTTPServer) ListHandler(c echo.Context) error {

	metrics, err := s.Storage.RetrieveAll()

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].GetName() < metrics[j].GetName()
	})

	return c.Render(http.StatusOK, "list.html", metrics)
}
