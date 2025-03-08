package file

import "context"

type DumpSaver interface {
	SaveDump(ctx context.Context) error
	RestoreDump(ctx context.Context) error
}
