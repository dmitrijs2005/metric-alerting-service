package main

import (
	"github.com/dmitrijs2005/metric-alerting-service/internal/staticlint"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	// appfile, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// data, err := os.ReadFile(filepath.Join(filepath.Dir(appfile), Config))
	// if err != nil {
	// 	panic(err)
	// }
	// var cfg ConfigData
	// if err = json.Unmarshal(data, &cfg); err != nil {
	// 	panic(err)
	// }
	// mychecks := []*analysis.Analyzer{
	// 	errorcheck.ErrCheckAnalyzer,
	// 	printf.Analyzer,
	// 	shadow.Analyzer,
	// 	structtag.Analyzer,
	// }
	// checks := make(map[string]bool)
	// for _, v := range cfg.Staticcheck {
	// 	checks[v] = true
	// }
	// // добавляем анализаторы из staticcheck, которые указаны в файле конфигурации
	// for _, v := range staticcheck.Analyzers {
	// 	if checks[v.Analyzer.Name] {
	// 		mychecks = append(mychecks, v.Analyzer)
	// 	}
	// }
	// multichecker.Main(
	// 	mychecks...,
	// )

	//standard := GetStandardAnalyzers()
	//fmt.Println(standard)

	var allChecks []*analysis.Analyzer

	// adding staticcheck SA class analyzers
	staticcheck := staticlint.GetStaticCheckAnalyzers()
	passes := staticlint.GetPassesAnalyzers()

	allChecks = append(allChecks, staticcheck...)
	allChecks = append(allChecks, passes...)

	multichecker.Main(
		allChecks...,
	)

}
