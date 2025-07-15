// Package metric defines core types and interfaces used to represent and manage application metrics.
//
// It includes a Metric interface that abstracts different types of metrics (e.g., gauge and counter),
// and a MetricType enumeration to distinguish between metric kinds.
//
// The supported metric types are:
//   - gauge   — a float64 representing a value that can go up or down (e.g., memory usage)
//   - counter — an int64 representing a monotonically increasing value (e.g., number of requests)
//
// Example usage:
//
//	var m metric.Metric
//	switch m.GetType() {
//	case metric.MetricTypeGauge:
//	    fmt.Println("Gauge:", m.GetValue())
//	case metric.MetricTypeCounter:
//	    fmt.Println("Counter:", m.GetValue())
//	}

package metric
