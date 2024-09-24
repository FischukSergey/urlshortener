package mycheck

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// переменные для анализатора с именем файла и функцией
const (
	funcName    = "main"
	packageName = "main"
)

// ErrNoExitAnalizer - анализатор, который проверяет наличие os.Exit в функции main
var ErrNoExitAnalizer = &analysis.Analyzer{
	Name: "ErrNoExitAnalizer",
	Doc:  "Check for os.Exit in main function",
	Run:  run,
}

/*
// Run - функция, которая выполняет анализ кода на наличие os.Exit в функции main
func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != packageName {
			continue
		}
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl: //если это функция
				if x.Name.Name == funcName { //и имя функции равно main
					ast.Inspect(x.Body, func(node ast.Node) bool { //итерируемся по всем операторам в теле функции
						switch stmt := node.(type) {
						case *ast.CallExpr: //если вызов функции
							switch fun := stmt.Fun.(type) {
							case *ast.SelectorExpr: //если вызов функции
								if fun.Sel.Name == "Exit" &&
									fun.X.(*ast.Ident).Name == "os" { //если имя функции равно Exit и имя пакета равно os
									pass.Reportf(stmt.Pos(), "os.Exit called in main function not allowed") //сообщение об ошибке
								}
							}
						}
						return true
					})
				}
			}
			return true
		})
	}
	return nil, nil
}
*/

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if file.Name.Name != packageName {
			continue
		}
		filename := pass.Fset.Position(file.Pos()).Filename
		if !strings.HasSuffix(filename, ".go") {
			continue
		}
		ast.Inspect(file, func(n ast.Node) bool {
			fun, ok := n.(*ast.FuncDecl)
			if ok {
				if fun.Name.Name == funcName {
					ast.Inspect(fun.Body, func(n ast.Node) bool {
						if call, ok := n.(*ast.CallExpr); ok {
							if isOsExitCall(call) {
								pass.Reportf(call.Pos(), "вызов Exit функции пакета os не рекомендуется")
							}
						}

						return true
					})
				}
			}

			return true
		})
	}

	return nil, nil
}

func isOsExitCall(call *ast.CallExpr) bool {
	if selectorIdent, ok := call.Fun.(*ast.SelectorExpr); ok {
		if parentIdent, ok := selectorIdent.X.(*ast.Ident); ok {
			if parentIdent.Name == "os" && selectorIdent.Sel.Name == "Exit" {
				return true
			}
		}
	}

	return false
}
