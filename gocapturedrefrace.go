package gocapturedrefrace

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "gocapturedrefrace",
	Doc:  "reports captured references in goroutine closures",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(
			file,
			func(node ast.Node) bool {
				goStmt, ok := node.(*ast.GoStmt)
				if !ok {
					return true
				}

				var printedNode bytes.Buffer
				err := printer.Fprint(&printedNode, pass.Fset, goStmt)
				if err != nil {
					panic(err)
				}

				fmt.Printf("%#v\n", goStmt)

				// TODO: Get func literal of go statement
				// TODO: Get variables in func literal
				funcLit, ok := goStmt.Call.Fun.(*ast.FuncLit)
				if !ok {
					return true
				}

				checkClosure(pass, funcLit)

				return true
			},
		)
	}

	return nil, nil
}

func checkClosure(pass *analysis.Pass, funcLit *ast.FuncLit) {
	fmt.Print("Params: ")
	for _, field := range funcLit.Type.Params.List {
		fmt.Printf("%#v, ", field.Names[0].Name)
	}
	fmt.Println()

	ast.Inspect(
		funcLit,
		func(node ast.Node) bool {
			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Obj == nil {
				return true
			}

			// TODO: Find out whether ident is a captured reference
			// Maybe check if variable was not assigned or passed as an argument?

			pass.Reportf(
				ident.Pos(),
				"variable found %q",
				ident,
			)

			return true
		},
	)
}
