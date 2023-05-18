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

// TODO: package documentation.
package gocapturedrefrace

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var version = "0.0.1"

var Analyzer = &analysis.Analyzer{
	Name: "gocapturedrefrace",
	Doc:  "reports captured references in goroutine closures",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// TODO: Since we're calling ast.Inspect a bunch of times, maybe it's worthwhile using passes/inspect now.
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

	// Build a list of assignments local to funcLit. These will be ignored as
	// shadowed variables.
	localAssignments := findLocalVarDeclarations(pass, funcLit)

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

			// Ignore shadowed variables.
			for _, localAssignmentIdent := range localAssignments {
				if ident == localAssignmentIdent {
					return true
				}
			}

			// Find out whether `ident` was defined in an outer scope.
			scope, scopeObj := funcScope.LookupParent(ident.Name, ident.NamePos)

			// Identifier is local to the closure.
			if scope == nil && scopeObj == nil {
				return true
			}

			// Ignore non-variable identifiers.
			variable, ok := scopeObj.(*types.Var)
			if !ok {
				return true
			}

			// Ignore captured callable variables, like function arguments.
			_, isVariableTypeSignature := variable.Type().(*types.Signature)
			if isVariableTypeSignature {
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

// findLocalVarDeclarations returns a list of all variable declarations in
// funcLit.
func findLocalVarDeclarations(
	pass *analysis.Pass,
	funcLit *ast.FuncLit,
) (declarations []*ast.Ident) {
	declarations = []*ast.Ident{}

	ast.Inspect(
		funcLit,
		func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				// assignStmt, ok := node.(*ast.AssignStmt)
				// if !ok {
				// 	return true
				// }
				assignStmt := node

				if assignStmt.Tok != token.DEFINE {
					return true
				}

				for _, lhs := range assignStmt.Lhs {
					ident, ok := lhs.(*ast.Ident)
					if !ok {
						return true
					}

					if ident == nil {
						return true
					}

					declarations = append(declarations, ident)
				}

			case *ast.GenDecl:
				decl := varDeclaration(node)
				if decl != nil {
					declarations = append(declarations, decl)
				}
			}

			return true
		},
	)

	return declarations
}

// varDeclaration returns the identifier corresponding to variable declarations
// in decl, or nil if decl is not a variable declaration.
func varDeclaration(decl *ast.GenDecl) *ast.Ident {
	if decl.Tok != token.VAR {
		return nil
	}

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			return nil
		}

		for _, ident := range valueSpec.Names {
			return ident
		}
	}

	return nil
}
