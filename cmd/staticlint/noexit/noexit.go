// Package noexit содержит собственный анализатор, который запрещает
// прямой вызов os.Exit в функции main пакета main.
package noexit

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// Analyzer запрещает прямой вызов os.Exit в функции main пакета main.
var MyAnalyzer = &analysis.Analyzer{
	Name: "noexit",
	Doc:  "запрещает вызывать os.Exit напрямую в функции main пакета main",
	Run:  run,
}

// run обходит дерево AST и ищет вызовы os.Exit в функции main пакета main.
func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	var currentFuncName string

	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {

			if fn, ok := node.(*ast.FuncDecl); ok {
				currentFuncName = fn.Name.Name
				return true
			}

			callExpr, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			if sel, ok := callExpr.Fun.(*ast.SelectorExpr); ok {
				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						if currentFuncName == "main" {
							pass.Reportf(
								callExpr.Pos(),
								"запрещён прямой вызов os.Exit в функции main пакета main",
							)
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
