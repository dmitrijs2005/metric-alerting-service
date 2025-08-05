// Package analyzer defines custom static analysis tools for Go code.
// It includes checks like detection of direct calls to os.Exit in the main function.
package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// OsExitAnalyzer is a static analysis tool that checks for direct
// calls to os.Exit inside the main() function.
// It helps ensure graceful shutdown practices are followed.
var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  "check for direct os.Exit calls in main function",
	Run:  run,
}

// shouldSkipFile returns true if the given file should be skipped from analysis,
// e.g., if it's in the Go build cache.
func shouldSkipFile(pass *analysis.Pass, file *ast.File) bool {
	position := pass.Fset.Position(file.Package)
	return strings.Contains(position.Filename, "/.cache/go-build/")
}

// checkMainFunction finds the main() function in the given file and runs
// the os.Exit call check inside its body.
func checkMainFunction(pass *analysis.Pass, file *ast.File) {
	ast.Inspect(file, func(node ast.Node) bool {
		fn, ok := node.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			return true
		}

		checkForOsExit(pass, fn.Body)
		return false // stop after main is checked
	})
}

// checkForOsExit reports any direct calls to os.Exit inside the given block of code.
func checkForOsExit(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
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

// run is the entry point for the analyzer. It scans each file and applies the check.
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if shouldSkipFile(pass, file) {
			continue
		}

		if file.Name.Name != "main" {
			continue
		}

		checkMainFunction(pass, file)
	}

	return nil, nil
}
