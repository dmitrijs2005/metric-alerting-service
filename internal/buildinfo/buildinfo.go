package buildinfo

import (
	"fmt"
	"io"
)

var buildVersion, buildDate, buildCommit string

func printBuildParam(w io.Writer, name string, value string) {
	if value == "" {
		value = "N/A"
	}
	fmt.Fprintf(w, "%s: %s\n", name, value)
}

func PrintBuildData(w io.Writer) {
	printBuildParam(w, "Build version", buildVersion)
	printBuildParam(w, "Build date", buildDate)
	printBuildParam(w, "Build commit", buildCommit)
}
