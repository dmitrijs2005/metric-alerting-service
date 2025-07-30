// Command staticlint runs a Go static analysis tool that combines built-in,
// custom, and third-party analyzers using the multichecker framework.
//
// Usage:
//
//	go run ./cmd/staticlint ./...
//
// Or build a binary:
//
//	go build -o staticlint ./cmd/staticlint
//	./staticlint ./...
//
// Included analyzers:
//   - All SA analyzers from the staticcheck package — detect potential bugs,
//     anti-patterns, and unused code. See: https://staticcheck.io/docs/checks/
//   - One analyzer from the simple suite (S1005, https://staticcheck.dev/docs/checks/#S1005) —
//     simplifies unnecessary use of the blank identifier `_`.
//   - One analyzer from the stylecheck suite (ST1008, https://staticcheck.dev/docs/checks/#ST1008) —
//     ensures that error return values are placed last in function signatures.
//   - One analyzer from the quickfix suite (QF1003, https://staticcheck.dev/docs/checks/#QF1003) — suggests replacing if/else chains with a switch statement.
//   - All built-in passes analyzers from golang.org/x/tools/go/analysis/passes —
//     provide low-level and syntactic checks.
//   - Public third-party analyzers:
//   - bodyclose (https://github.com/timakin/bodyclose): checks that `http.Response.Body` is properly closed.
//   - nakedret (https://github.com/alexkohler/nakedret): flags naked return statements in long functions.
//   - ineffassign (https://github.com/gordonklaus/ineffassign): detects ineffectual assignments (values assigned but never used).
//   - Custom OsExitAnalyzer: detects direct calls to os.Exit in the main function.
//
// This tool uses multichecker from golang.org/x/tools/go/analysis/multichecker,
// which allows combining multiple analyzers and running them as a unified analysis tool.
//
// More on multichecker: https://pkg.go.dev/golang.org/x/tools/go/analysis/multichecker
package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/analyzer"
	"github.com/dmitrijs2005/metric-alerting-service/internal/staticlint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	var allChecks []*analysis.Analyzer

	// adding staticcheck analyzers
	staticcheck := staticlint.GetStaticCheckAnalyzers()

	// adding Passes analyzers
	passes := staticlint.GetPassesAnalyzers()

	// adding some public analyzers
	public := staticlint.GetPublicAnalyzers()

	allChecks = append(allChecks, staticcheck...)
	allChecks = append(allChecks, passes...)
	allChecks = append(allChecks, public...)
	allChecks = append(allChecks, analyzer.OsExitAnalyzer)

	multichecker.Main(
		allChecks...,
	)

}
