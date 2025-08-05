package staticlint

import (
	"github.com/alexkohler/nakedret"
	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"github.com/timakin/bodyclose/passes/bodyclose"
	"golang.org/x/tools/go/analysis"
)

func GetPublicAnalyzers() []*analysis.Analyzer {
	var result []*analysis.Analyzer

	// Bodyclose is a static analysis tool which checks whether res.Body is correctly closed.
	result = append(result, bodyclose.Analyzer)

	// Nakedret is a Go static analysis tool to find naked returns in functions greater than a specified function length.
	result = append(result, nakedret.NakedReturnAnalyzer(3))

	// Detect ineffectual assignments in Go code. An assignment is ineffectual if the variable assigned is not thereafter used.
	result = append(result, ineffassign.Analyzer)

	return result
}
