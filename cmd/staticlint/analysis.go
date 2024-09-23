package main

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"honnef.co/go/tools/staticcheck"
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

func main() {

	var myChecks = []*analysis.Analyzer{
		ErrNoExitAnalizer,
	}

	for _, v := range staticcheck.Analyzers {
		myChecks = append(myChecks, v.Analyzer)
	}
	multichecker.Main(myChecks...)

}
