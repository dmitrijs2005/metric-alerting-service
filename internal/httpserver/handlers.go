package httpserver

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
)

// type Metrics struct {
// 	ID    string   `json:"id"`              // имя метрики
// 	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
// 	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
// 	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
// }

func (s *HTTPServer) updateMetric(metricType string, metricName string, metricValue interface{}) (metric.Metric, error) {

	m, err := s.Storage.Retrieve(metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		m, err = metric.NewMetric(metric.MetricType(metricType), metricName)
		if err != nil {
			return nil, err
		}
		err = s.Storage.Add(m)
		if err != nil {
			return nil, err
		}
	}

	err = m.Update(metricValue)
	if err != nil {

		return nil, err
	}

	return m, nil

}

// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"g22","type":"gauge","value":123.12}'
// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"c33","type":"counter","delta":3}'

func (s *HTTPServer) UpdateJSONHandler(c echo.Context) error {

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	metricType := mDTO.MType
	metricName := mDTO.ID

	var metricValue interface{}
	switch metricType {
	case string(metric.MetricTypeCounter):
		if mDTO.Delta == nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
		metricValue = *mDTO.Delta
	case string(metric.MetricTypeGauge):
		if mDTO.Value == nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
		metricValue = *mDTO.Value
	default:
		return c.String(http.StatusBadRequest, metric.ErrorInvalidMetricType.Error())
	}

	//s.Logger.Info(fmt.Sprintf("update %s %s %v", metricType, metricName, metricValue))

	m, err := s.updateMetric(metricType, metricName, metricValue)

	if err != nil {

		isBadRequest := errors.Is(err, metric.ErrorInvalidMetricName) || errors.Is(err, metric.ErrorInvalidMetricType) || errors.Is(err, metric.ErrorInvalidMetricValue)

		if isBadRequest {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	val, ok := m.GetValue().(int64)
	if ok {
		float64Val := float64(val)
		mDTO.Value = &float64Val
	} else {
		val, ok := m.GetValue().(float64)
		if ok {
			mDTO.Value = &val
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	// // if everything is correct and metric was saved
	return c.JSON(http.StatusOK, mDTO)
}

func (s *HTTPServer) UpdateHandler(c echo.Context) error {

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	_, err := s.updateMetric(metricType, metricName, metricValue)

	if err != nil {

		isBadRequest := errors.Is(err, metric.ErrorInvalidMetricName) || errors.Is(err, metric.ErrorInvalidMetricType) || errors.Is(err, metric.ErrorInvalidMetricValue)

		if isBadRequest {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	// if everything is correct and metric was saved
	return c.String(http.StatusOK, "OK")
}

func (s *HTTPServer) ValueJSONHandler(c echo.Context) error {

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	metricType := mDTO.MType
	metricName := mDTO.ID

	m, err := s.Storage.Retrieve(metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		return c.String(http.StatusNotFound, err.Error())
	}

	val, ok := m.GetValue().(int64)
	if ok {
		//float64Val := float64(val)
		mDTO.Delta = &val
	} else {
		val, ok := m.GetValue().(float64)
		if ok {
			mDTO.Value = &val
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	//s.Logger.Info(fmt.Sprintf("value %s %s %v", metricType, metricName, mDTO))

	return c.JSON(http.StatusOK, mDTO)
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
