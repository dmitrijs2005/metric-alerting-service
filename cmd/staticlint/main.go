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
