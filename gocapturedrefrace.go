package gocapturedrefrace

import (
	"go/ast"
	"go/types"

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
				// Find `go` statements.
				goStmt, ok := node.(*ast.GoStmt)
				if !ok {
					return true
				}

				// Look for a function literal after the `go` statement.
				funcLit, ok := goStmt.Call.Fun.(*ast.FuncLit)
				if !ok {
					return true
				}

				// Inspect closure argument list.
				for _, arg := range funcLit.Type.Params.List {
					// Report reference arguments.
					_, ok := arg.Type.(*ast.StarExpr)
					if !ok {
						continue
					}

					pass.Reportf(
						arg.Pos(),
						"reference %s in goroutine closure",
						arg.Names[0],
					)
				}

				// Get the closure's scope.
				funcScope := pass.TypesInfo.Scopes[funcLit.Type]

				checkClosure(pass, funcLit, funcScope)

				return true
			},
		)
	}

	return nil, nil
}

func checkClosure(
	pass *analysis.Pass,
	funcLit *ast.FuncLit,
	funcScope *types.Scope,
) {
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

			// Find out whether `ident` was defined in an outer scope.
			scope, scopeObj := funcScope.LookupParent(ident.Name, ident.NamePos)

			// Identifier is local to the closure.
			if scope == nil && scopeObj == nil {
				return true
			}

			// Ignore non-variable identifiers.
			_, ok = scopeObj.(*types.Var)
			if !ok {
				return true
			}

			// Identifier was defined in a different scope.
			if funcScope != scope {
				pass.Reportf(
					ident.Pos(),
					"captured reference %s in goroutine closure",
					ident,
				)
			}

			return true
		},
	)
}
