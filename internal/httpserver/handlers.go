package httpserver

import (
	"fmt"
	"net/http"

	"github.com/dmitrijs2005/metric-alerting-service/internal/metrics"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) UpdateHandler(c echo.Context) error {

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	fmt.Println(metricType, metricName, metricValue)

	m, err := s.Storage.Retrieve(metrics.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		m, err = metrics.NewMetric(metrics.MetricType(metricType), metricName)
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

		if err.Error() == metrics.ErrorInvalidMetricValue {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	// if everything is correct and metric was saved
	return c.String(http.StatusOK, "OK")
}
