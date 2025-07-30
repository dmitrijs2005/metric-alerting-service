package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "analyzer",
	Doc:  "check for unchecked errors",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {

		position := pass.Fset.Position(file.Package)
		if strings.Contains(position.Filename, "/.cache/go-build/") {
			continue
		}

		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {

			if fn, ok := node.(*ast.FuncDecl); ok {

				if fn.Name.Name != "main" {
					return true
				}

				ast.Inspect(fn.Body, func(node ast.Node) bool {

					call, ok := node.(*ast.CallExpr)

					if !ok {
						return true
					}

					selector, ok := call.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}

					pkg, ok := selector.X.(*ast.Ident)
					if !ok {
						return true
					}

					if selector.Sel.Name == "Exit" && pkg.Name == "os" {
						pass.Reportf(call.Pos(), "avoid using os.Exit inside main()")
					}

					return true
				})
			}

			return true
		})
	}
	return nil, nil
}
