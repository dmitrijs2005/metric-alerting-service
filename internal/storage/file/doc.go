// Package file provides a file-based implementation of metric dump persistence.
//
// It defines the DumpSaver interface, which includes methods for saving and restoring
// application metrics. The FileSaver struct implements this interface by writing metrics
// to a plain text file and restoring them from it.
//
// The dump format is line-based:
//
//	metric_name:metric_type:metric_value
//
// For example:
//
//	requests_total:counter:42
//	temperature:gauge:36.6
//
// Typical usage:
//
//	saver := file.NewFileSaver("metrics.dump", metricStorage)
//	err := saver.RestoreDump(ctx)
//	// ... application logic ...
//	err := saver.SaveDump(ctx)
package file
