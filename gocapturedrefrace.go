// Copyright (c) 2023  Teddy Wing
//
// This file is part of Gocapturedrefrace.
//
// Gocapturedrefrace is free software: you can redistribute it and/or
// modify it under the terms of the GNU General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// Gocapturedrefrace is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU
// General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Gocapturedrefrace. If not, see
// <https://www.gnu.org/licenses/>.


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
					// TODO: Ignore closures passed as arguments.

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

				checkClosure(pass, funcLit)

				return true
			},
		)
	}

	return nil, nil
}

// checkClosure reports variables used in funcLit that are captured from an
// outer scope.
func checkClosure(pass *analysis.Pass, funcLit *ast.FuncLit) {
	// Get the closure's scope.
	funcScope := pass.TypesInfo.Scopes[funcLit.Type]

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

			// TODO: Ignore shadowing variables.

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
