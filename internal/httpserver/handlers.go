package httpserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) updateMetric(ctx context.Context, metricType string, metricName string, metricValue interface{}) (metric.Metric, error) {

	m, err := s.Storage.Retrieve(ctx, metric.MetricType(metricType), metricName)

	if err != nil {
		if err.Error() != storage.MetricDoesNotExist {
			return nil, err
		} else {
			m, err = metric.NewMetric(metric.MetricType(metricType), metricName)
			if err != nil {
				return nil, err
			}

			if gauge, ok := m.(*metric.Gauge); ok {
				if err := gauge.Update(metricValue); err != nil {
					return nil, err
				}
			} else if counter, ok := m.(*metric.Counter); ok {
				if err := counter.Update(metricValue); err != nil {
					return nil, err
				}
			} else {
				return nil, metric.ErrorInvalidMetricType
			}

			err = s.Storage.Add(ctx, m)
			if err != nil {
				return nil, err
			}
		}
	} else {
		err = s.Storage.Update(ctx, m, metricValue)
		if err != nil {

			return nil, err
		}
	}

	return m, nil

}

// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"g22","type":"gauge","value":123.12}'
// curl -v -X POST 'http://localhost:8080/update/' -H "Content-Type: application/json" -d '{"id":"c33","type":"counter","delta":3}'

func (s *HTTPServer) fillValue(m metric.Metric, r *dto.Metrics) error {
	switch m.GetType() {
	case metric.MetricTypeCounter:
		int64Val, ok := m.GetValue().(int64)
		if ok {
			r.Delta = &int64Val
		} else {
			return ErrorTypeConversion
		}
	case metric.MetricTypeGauge:
		float64Val, ok := m.GetValue().(float64)
		if ok {
			r.Value = &float64Val
		} else {
			return ErrorTypeConversion
		}
	default:
		return metric.ErrorInvalidMetricType
	}
	return nil
}

func (s *HTTPServer) FromDto(mDTO dto.Metrics) (metric.Metric, error) {

	m, err := metric.NewMetric(metric.MetricType(mDTO.MType), mDTO.ID)
	if err != nil {
		return nil, err
	}

	if gauge, ok := m.(*metric.Gauge); ok {
		err := gauge.Update(*mDTO.Value)
		if err != nil {
			return nil, err
		}
	} else if counter, ok := m.(*metric.Counter); ok {
		err := counter.Update(*mDTO.Delta)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, metric.ErrorInvalidMetricType
	}

	return m, nil

}

func (s *HTTPServer) UpdateJSONHandler(c echo.Context) error {

	ctx := c.Request().Context()

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	var metricValue interface{}

	switch metric.MetricType(mDTO.MType) {
	case metric.MetricTypeCounter:
		if mDTO.Delta == nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
		metricValue = *mDTO.Delta
	case metric.MetricTypeGauge:
		if mDTO.Value == nil {
			return c.String(http.StatusBadRequest, "bad request")
		}
		metricValue = *mDTO.Value
	default:
		return c.String(http.StatusBadRequest, metric.ErrorInvalidMetricType.Error())
	}

	m, err := s.updateMetric(ctx, mDTO.MType, mDTO.ID, metricValue)

	if err != nil {

		s.logger.Errorw("Error updating metric", "err", err)

		isBadRequest := errors.Is(err, metric.ErrorInvalidMetricName) || errors.Is(err, metric.ErrorInvalidMetricType) || errors.Is(err, metric.ErrorInvalidMetricValue)

		if isBadRequest {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	err = s.fillValue(m, mDTO)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// // if everything is correct and metric was saved
	return c.JSON(http.StatusOK, mDTO)
}

func (s *HTTPServer) UpdateHandler(c echo.Context) error {

	ctx := c.Request().Context()

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	_, err := s.updateMetric(ctx, metricType, metricName, metricValue)

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

	ctx := c.Request().Context()

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	metricType := mDTO.MType
	metricName := mDTO.ID

	m, err := s.Storage.Retrieve(ctx, metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		return c.String(http.StatusNotFound, err.Error())
	}

	err = s.fillValue(m, mDTO)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mDTO)
}

func (s *HTTPServer) ValueHandler(c echo.Context) error {

	ctx := c.Request().Context()

	metricType := c.Param("type")
	metricName := c.Param("name")

	m, err := s.Storage.Retrieve(ctx, metric.MetricType(metricType), metricName)

	if m == nil && err.Error() == storage.MetricDoesNotExist {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("%v", m.GetValue()))
}

func (s *HTTPServer) ListHandler(c echo.Context) error {

	ctx := c.Request().Context()
	metrics, err := s.Storage.RetrieveAll(ctx)

	if err != nil {
		s.logger.Errorw("Error retrieving metrics", "err", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	sort.Slice(metrics, func(i, j int) bool {
		return metrics[i].GetName() < metrics[j].GetName()
	})

	return c.Render(http.StatusOK, "list.html", metrics)
}

func (s *HTTPServer) PingHandler(c echo.Context) error {

	ctx := c.Request().Context()

	db, ok := s.Storage.(storage.DBStorage)
	if ok {
		err := db.Ping(ctx)

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}

		return c.String(http.StatusOK, "OK")
	}

	return c.String(http.StatusInternalServerError, ErrorTypeNotImplemented.Error())

}

func (s *HTTPServer) UpdatesJSONHandler(c echo.Context) error {

	ctx := c.Request().Context()

	mDTO := new([]dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		s.logger.Errorw("Error converting DTO to metric", "err", err)
		return c.String(http.StatusBadRequest, "bad request")
	}

	metrics := make([]metric.Metric, len(*mDTO))

	for i, o := range *mDTO {
		m, err := s.FromDto(o)
		if err != nil {
			s.logger.Errorw("Error initializing metric", "err", err)
			return c.String(http.StatusBadRequest, "bad request")
		}
		metrics[i] = m
	}
	s.Storage.UpdateBatch(ctx, &metrics)

	return c.JSON(http.StatusOK, mDTO)
}
