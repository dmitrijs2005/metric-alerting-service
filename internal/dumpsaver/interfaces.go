package dumpsaver

type DumpSaver interface {
	SaveDump() error
	RestoreDump() error
}
