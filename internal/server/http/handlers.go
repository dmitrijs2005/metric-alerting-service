package http

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/dmitrijs2005/metric-alerting-service/internal/common"
	"github.com/dmitrijs2005/metric-alerting-service/internal/dto"
	"github.com/dmitrijs2005/metric-alerting-service/internal/metric"
	"github.com/dmitrijs2005/metric-alerting-service/internal/server/usecase"
	"github.com/dmitrijs2005/metric-alerting-service/internal/storage"
	"github.com/labstack/echo/v4"
)

func (s *HTTPServer) MetricFromDto(mDTO dto.Metrics) (metric.Metric, error) {

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

func (s *HTTPServer) DTOFromMetric(m metric.Metric) (*dto.Metrics, error) {

	o := &dto.Metrics{ID: m.GetName(), MType: string(m.GetType())}

	if gauge, ok := m.(*metric.Gauge); ok {
		o.Value = float64Ptr(gauge.Value)
	} else if counter, ok := m.(*metric.Counter); ok {
		o.Delta = int64Ptr(counter.Value)
	} else {
		return nil, metric.ErrorInvalidMetricType
	}

	return o, nil

}

// UpdateJSONHandler handles an HTTP POST request with a single metric in JSON format.
//
// It expects a JSON body representing a metric (either `gauge` or `counter`) as defined by dto.Metrics.
//
// If the metric is valid, it updates the internal metric storage and returns the updated metric as JSON.
//
// Returns:
//   - 400 Bad Request: if the input is invalid or contains an unsupported metric type
//   - 500 Internal Server Error: if updating or retrieving the metric fails
//   - 200 OK: with the updated metric in JSON format
//
// Example JSON body:
//
//	{
//	  "id": "Alloc",
//	  "type": "gauge",
//	  "value": 123.4
//	}
//
// Supported metric types:
//   - gauge (float64)
//   - counter (int64)
func (s *HTTPServer) UpdateJSONHandler(c echo.Context) error {

	ctx := c.Request().Context()

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		s.logger.Errorw("Error parsing metric", "err", err)
		return c.String(http.StatusBadRequest, "bad request")
	}

	var metricValue interface{}

	switch metric.MetricType(mDTO.MType) {
	case metric.MetricTypeCounter:
		if mDTO.Delta == nil {
			msg := "wrong delta"
			s.logger.Errorw("Error parsing metric", "err", msg)
			return c.String(http.StatusBadRequest, msg)
		}
		metricValue = *mDTO.Delta
	case metric.MetricTypeGauge:
		if mDTO.Value == nil {
			msg := "wrong value"
			s.logger.Errorw("Error parsing metric", "err", msg)
			return c.String(http.StatusBadRequest, msg)
		}
		metricValue = *mDTO.Value
	default:
		return c.String(http.StatusBadRequest, metric.ErrorInvalidMetricType.Error())
	}

	m, err := usecase.UpdateMetricByValue(ctx, s.Storage, mDTO.MType, mDTO.ID, metricValue)
	if err != nil {
		s.logger.Errorw("Error updating metric", "err", err)

		isBadRequest := errors.Is(err, metric.ErrorInvalidMetricName) || errors.Is(err, metric.ErrorInvalidMetricType) || errors.Is(err, metric.ErrorInvalidMetricValue)

		if isBadRequest {
			return c.String(http.StatusBadRequest, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	updated, err := s.Storage.Retrieve(ctx, m.GetType(), m.GetName())
	if err != nil {
		s.logger.Errorw("Error retrieving metric", "err", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	result, err := s.DTOFromMetric(updated)
	if err != nil {
		s.logger.Errorw("Error retrieving metric", "err", err)
		return c.String(http.StatusInternalServerError, err.Error())
	}

	// // if everything is correct and metric was saved
	return c.JSON(http.StatusOK, result)
}

// UpdateHandler handles an HTTP POST request that updates a metric using URL path parameters.
//
// Expected URL path parameters:
//   - :type  — metric type ("gauge" or "counter")
//   - :name  — metric name
//   - :value — metric value (float64 for gauge, int64 for counter)
//
// Example request:
//
//	POST /update/counter/requests/42
//	POST /update/gauge/temperature/36.6
//
// Responses:
//   - 200 OK: if the metric was successfully updated
//   - 400 Bad Request: if the type, name, or value is invalid
//   - 500 Internal Server Error: if an unexpected error occurred
func (s *HTTPServer) UpdateHandler(c echo.Context) error {

	ctx := c.Request().Context()

	metricType := c.Param("type")
	metricName := c.Param("name")
	metricValue := c.Param("value")

	_, err := usecase.UpdateMetricByValue(ctx, s.Storage, metricType, metricName, metricValue)

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

// ValueJSONHandler handles an HTTP POST request that retrieves the current value of a metric specified in JSON format.
//
// It expects a JSON body with the metric's `id` and `type` (either "gauge" or "counter").
//
// If the metric exists, the response includes the same metric object with the `value` or `delta` field populated.
//
// Example request:
//
//	POST /value/
//	{
//	  "id": "Alloc",
//	  "type": "gauge"
//	}
//
// Example response:
//
//	{
//	  "id": "Alloc",
//	  "type": "gauge",
//	  "value": 123.45
//	}
//
// Responses:
//   - 200 OK: if the metric exists and its value is returned
//   - 400 Bad Request: if the request body is invalid
//   - 404 Not Found: if the metric does not exist
//   - 500 Internal Server Error: if a server-side error occurs
func (s *HTTPServer) ValueJSONHandler(c echo.Context) error {

	ctx := c.Request().Context()

	mDTO := new(dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		return c.String(http.StatusBadRequest, "bad request")
	}

	metricType := mDTO.MType
	metricName := mDTO.ID

	m, err := s.Storage.Retrieve(ctx, metric.MetricType(metricType), metricName)

	if m == nil {
		if errors.Is(err, common.ErrorMetricDoesNotExist) {
			return c.String(http.StatusNotFound, err.Error())
		} else {
			return c.String(http.StatusInternalServerError, err.Error())
		}
	}

	err = usecase.FillValue(m, mDTO)

	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, mDTO)
}

// ValueHandler handles an HTTP GET request that returns the current value of a metric,
// specified via URL path parameters.
//
// Expected URL path parameters:
//   - :type — metric type ("gauge" or "counter")
//   - :name — metric name
//
// Example request:
//
//	GET /value/counter/requests
//	GET /value/gauge/temperature
//
// Responses:
//   - 200 OK: returns the current value of the requested metric as plain text
//   - 404 Not Found: if the metric does not exist
func (s *HTTPServer) ValueHandler(c echo.Context) error {

	ctx := c.Request().Context()

	metricType := c.Param("type")
	metricName := c.Param("name")

	m, err := s.Storage.Retrieve(ctx, metric.MetricType(metricType), metricName)

	if m == nil && errors.Is(err, common.ErrorMetricDoesNotExist) {
		return c.String(http.StatusNotFound, err.Error())
	}

	return c.String(http.StatusOK, fmt.Sprintf("%v", m.GetValue()))
}

// ListHandler handles an HTTP GET request that renders a list of all stored metrics.
//
// It retrieves all available metrics from the storage, sorts them alphabetically by name,
// and renders them using the "list.html" template.
//
// Responses:
//   - 200 OK: renders the list of metrics
//   - 500 Internal Server Error: if metrics could not be retrieved from storage
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

// PingHandler handles a health check request to verify database connectivity.
//
// If the storage implements the DBStorage interface, it performs a database ping.
// If the ping is successful, it returns HTTP 200 with "OK".
// If the ping fails or the storage does not support DB access, it returns an error.
//
// Responses:
//   - 200 OK: if the database is reachable
//   - 500 Internal Server Error: if the ping fails or the storage does not support Ping
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

	return c.String(http.StatusInternalServerError, common.ErrorTypeNotImplemented.Error())

}

// UpdatesJSONHandler handles an HTTP POST request that updates multiple metrics in batch via JSON.
//
// It expects a JSON array of metric objects (either `gauge` or `counter`) as input.
// Each object is converted to an internal metric and passed to a batch update operation.
//
// After updating, it retrieves and returns the updated metrics with their current values.
//
// Example request:
//
//	POST /updates/
//	[
//	  {"id": "requests", "type": "counter", "delta": 5},
//	  {"id": "temperature", "type": "gauge", "value": 36.6}
//	]
//
// Example response:
//
//	[
//	  {"id": "requests", "type": "counter", "delta": 105},
//	  {"id": "temperature", "type": "gauge", "value": 36.6}
//	]
//
// Responses:
//   - 200 OK: if all metrics were successfully updated
//   - 400 Bad Request: if input is malformed or update fails
//   - 500 Internal Server Error: if retrieval or transformation fails
func (s *HTTPServer) UpdatesJSONHandler(c echo.Context) error {

	ctx := c.Request().Context()

	mDTO := new([]dto.Metrics)
	if err := c.Bind(mDTO); err != nil {
		s.logger.Errorw("Error converting DTO to metric", "err", err)
		return c.String(http.StatusBadRequest, "bad request")
	}

	metrics := make([]metric.Metric, len(*mDTO))

	for i, o := range *mDTO {
		m, err := s.MetricFromDto(o)
		if err != nil {
			s.logger.Errorw("Error initializing metric", "err", err)
			return c.String(http.StatusBadRequest, "bad request")
		}
		metrics[i] = m
	}

	err := s.Storage.UpdateBatch(ctx, &metrics)
	if err != nil {
		s.logger.Errorw("Error updating metrics", "err", err)
		return c.String(http.StatusBadRequest, err.Error())
	}

	results := make([]dto.Metrics, len(*mDTO))

	for i, o := range *mDTO {
		updated, err := s.Storage.Retrieve(ctx, metric.MetricType(o.MType), o.ID)
		if err != nil {
			s.logger.Errorw("Error retrieving metric", "err", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		result, err := s.DTOFromMetric(updated)
		if err != nil {
			s.logger.Errorw("Error retrieving metric", "err", err)
			return c.String(http.StatusInternalServerError, err.Error())
		}

		results[i] = *result

	}

	return c.JSON(http.StatusOK, results)
}
