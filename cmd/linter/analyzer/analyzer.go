package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer — статический анализатор, который проверяет:
// - использование panic
// - вызов log.Fatal / os.Exit вне main функции пакета main
var Analyzer = &analysis.Analyzer{
	Name: "safecalls",
	Doc:  "reports usage of panic, log.Fatal and os.Exit outside of main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			if n == nil {
				return true
			}

			// A CallExpr node represents an expression followed by an argument list
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			if ident, ok := call.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" {
					pass.Reportf(call.Pos(), "usage of panic is not allowed")
				}
			}

			if selectorExpr, ok := call.Fun.(*ast.SelectorExpr); ok {
				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				name := ident.Name + "." + selectorExpr.Sel.Name

				switch name {
				case "log.Fatal", "log.Panic":
					pass.Reportf(call.Pos(), "usage of %s is not allowed", name)
				case "os.Exit":
					pass.Reportf(call.Pos(), "usage of os.Exit is not allowed")
				}
			}
			return true
		})
	}

	return nil, nil
}
