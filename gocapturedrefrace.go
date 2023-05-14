package gocapturedrefrace

import (
	"bytes"
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

				pass.Reportf(
					goStmt.Pos(),
					"go statement found %q",
					printedNode,
				)

				return true
			},
		)
	}

	return nil, nil
}
