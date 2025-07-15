// Package httpserver provides HTTP handlers for interacting with the metrics storage.
//
// This package implements endpoints for updating, retrieving, and listing application metrics.
// It supports both individual and batch operations via JSON or path parameters,
// and includes support for health checks and HTML rendering of metrics.
//
// The main handlers include:
//
//   - UpdateHandler: updates a metric via path parameters
//   - UpdateJSONHandler: updates a metric via JSON payload
//   - UpdatesJSONHandler: batch update of multiple metrics via JSON
//   - ValueHandler: retrieves a metric value via path parameters
//   - ValueJSONHandler: retrieves a metric value via JSON payload
//   - ListHandler: renders all metrics as HTML
//   - PingHandler: health check endpoint to verify DB connectivity
//
// All handlers are implemented as methods on the HTTPServer struct,
// and rely on a shared metric storage layer and logging interface.
package httpserver
