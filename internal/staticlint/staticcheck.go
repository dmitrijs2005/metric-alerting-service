package staticlint

import (
	"golang.org/x/tools/go/analysis"
	"honnef.co/go/tools/staticcheck"
)

func GetStaticCheckAnalyzers() []*analysis.Analyzer {
	var result []*analysis.Analyzer

	// 1. Все SA анализаторы
	for _, a := range staticcheck.Analyzers {
		result = append(result, a.Analyzer)
	}

	return result
}
