package analyzer

import (
	"go/ast"
	"go/types"

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
		// search panic through all files
		ast.Inspect(file, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if ident, ok := call.Fun.(*ast.Ident); ok && ident.Name == "panic" {
				pass.Reportf(call.Pos(), "usage of panic is not allowed")
			}
			return true
		})

		// log.Fatal / os.Exit — search outside main() func of package main
		for _, decl := range file.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok {
				continue
			}

			if pass.Pkg.Name() == "main" && funcDecl.Name.Name == "main" {
				continue
			}

			ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				selectorExpr, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				ident, ok := selectorExpr.X.(*ast.Ident)
				if !ok {
					return true
				}

				pkgName, ok := pass.TypesInfo.Uses[ident].(*types.PkgName)
				if !ok {
					return true
				}
				pkgPath := pkgName.Imported().Path() // "os" или "log"
				funcName := selectorExpr.Sel.Name    // "Exit" или "Fatal"

				name := pkgPath + "." + funcName
				switch name {
				case "log.Fatal", "log.Panic":
					pass.Reportf(call.Pos(), "usage of %s is not allowed", name)
				case "os.Exit":
					pass.Reportf(call.Pos(), "usage of os.Exit is not allowed")
				}

				return true
			})
		}
	}

	return nil, nil
}
