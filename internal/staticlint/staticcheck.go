package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"
)

func GetStaticCheckAnalyzers() []*analysis.Analyzer {
	var result []*analysis.Analyzer

	// All SA class analyzers
	for _, a := range staticcheck.Analyzers {
		result = append(result, a.Analyzer)
	}

	// S1005 - Drop unnecessary use of the blank identifier
	for _, a := range simple.Analyzers {
		if a.Analyzer.Name == "S1005" {
			result = append(result, a.Analyzer)
		}
	}

	// ST1008 - A function’s error value should be its last return value
	for _, a := range stylecheck.Analyzers {
		if a.Analyzer.Name == "ST1008" {
			result = append(result, a.Analyzer)
		}
	}

	// QF1003 - Convert if/else-if chain to tagged switch
	for _, a := range quickfix.Analyzers {
		if a.Analyzer.Name == "QF1003" {
			result = append(result, a.Analyzer)
		}
	}

	return result
}
