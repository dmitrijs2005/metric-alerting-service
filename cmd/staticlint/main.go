package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/staticlint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {

	var allChecks []*analysis.Analyzer

	// adding staticcheck SA class analyzers
	staticcheck := staticlint.GetStaticCheckAnalyzers()

	// adding staticcheck Passes analyzers
	passes := staticlint.GetPassesAnalyzers()

	allChecks = append(allChecks, staticcheck...)
	allChecks = append(allChecks, passes...)

	multichecker.Main(
		allChecks...,
	)

}
