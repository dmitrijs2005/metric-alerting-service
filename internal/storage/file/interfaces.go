package file

import "context"

// DumpSaver defines methods for persisting metrics to and from storage.
type DumpSaver interface {
	// SaveDump saves all metrics to persistent storage (e.g., file).
	SaveDump(ctx context.Context) error

	// RestoreDump restores metrics from persistent storage.
	RestoreDump(ctx context.Context) error
}
