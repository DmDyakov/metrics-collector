package main

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// noOsExitAnalyzer запрещает прямой вызов os.Exit в функции main пакета main.
//
// Анализатор проверяет:
// - Что пакет называется main (точка входа программы)
// - Что функция называется main и не имеет ресивера
// - Что в теле функции нет вызова os.Exit
//
// Использование os.Exit в main-функции main-пакета считается плохой практикой,
// так как:
// - defer-функции не выполняются
// - Сложно тестировать
// - Нарушает принцип единственной точки выхода
//
// Рекомендуется возвращать ошибку из main и обрабатывать её на верхнем уровне.
var noOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "запрещает прямой вызов os.Exit в функции main пакета main",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	if pass.Pkg.Name() != "main" {
		return nil, nil
	}

	for _, file := range pass.Files {
		filename := pass.Fset.Position(file.Pos()).Filename
		if strings.Contains(filename, "go-build") {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			fn, ok := node.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" || fn.Recv != nil {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				pkg, ok := sel.X.(*ast.Ident)
				if ok && pkg.Name == "os" && sel.Sel.Name == "Exit" {
					pass.Reportf(call.Pos(), "прямой вызов os.Exit запрещён в main")
				}
				return true
			})
			return true
		})
	}
	return nil, nil
}
