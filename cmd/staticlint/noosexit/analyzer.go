// Package noosexit отвечает за анализатор, запрещающий использовать
// прямой вызов os.Exit в функции main пакета main
package noosexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// doc - описание работы анализатора
const doc = `noosexit checks for direct calls to os.Exit in main function of main package

This analyzer reports direct calls to os.Exit in the main function
of the main package as they can prevent proper cleanup and testing.`

// Analyzer - анализатор для проверки os.Exit
var Analyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// run - запускает анализатор
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.File)(nil),
		(*ast.CallExpr)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		// Проверяем, что мы в пакете main
		file, ok := n.(*ast.File)
		if ok {
			if file.Name.Name != "main" {
				return
			}
			return
		}

		call, ok := n.(*ast.CallExpr)
		if !ok {
			return
		}

		// Проверяем, что это вызов функции
		fun, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		// Проверяем, что это os.Exit
		if ident, ok := fun.X.(*ast.Ident); ok {
			if ident.Name == "os" && fun.Sel.Name == "Exit" {
				pass.Reportf(call.Pos(), "direct call to os.Exit in main package is forbidden")
			}
		}
	})

	return nil, nil
}
