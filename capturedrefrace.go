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

// Package capturedrefrace defines an Analyzer that checks for captured
// references in goroutine closures.
//
// # Analyzer capturedrefrace
//
// capturedrefrace: report captured variable references in goroutine
// closures.
//
// Goroutines that run function closures can capture reference variables from
// outer scopes which could lead to data races. This analyzer checks closures
// run by goroutines and reports uses of all variables declared in outer
// scopes, as well as arguments to the closure with a pointer type.
//
// For example:
//
//	func (r *Record) CapturedReference() {
//		capturedReference := 0
//		spline := &Spline{Curvature: 5.0}
//
//		go func(s *Spline) {
//			capturedReference += 1 // closure captures the variable
//				// 'capturedReference' in a goroutine, which could
//				// lead to data races
//
//			if capturedReference > 0 {
//				r.reticulateSplines() // goroutine closure captures 'r'
//			}
//
//			s.Curvature = 3.0 // 's' is a pointer type which could
//				// lead to data races
//		}(spline)
//	}
package capturedrefrace

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/exp/slices"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "capturedrefrace",
	Doc:      "reports captured references in goroutine closures",
	Run:      run,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.GoStmt)(nil),
	}

	inspect.Preorder(
		nodeFilter,
		func(node ast.Node) {
			// Find `go` statements.
			goStmt, ok := node.(*ast.GoStmt)
			if !ok {
				return
			}

			var funcLit *ast.FuncLit

			switch goStmtFunc := goStmt.Call.Fun.(type) {
			case *ast.FuncLit:
				// Look for a function literal after the `go` statement.
				funcLit, ok = goStmt.Call.Fun.(*ast.FuncLit)

			case *ast.Ident:
				// Get a function literal stored in a local variable.
				funcLit, ok = funcLitFromIdent(goStmtFunc)

			default:
				// The function being run by the `go` statement is an
				// unsupported type.
				return
			}
			if !ok {
				return
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
		},
	)

	return nil, nil
}

// funcLitFromIdent takes a variable identifier and looks for a function
// literal assigned to it, returning the function literal.
func funcLitFromIdent(funcIdent *ast.Ident) (funcLit *ast.FuncLit, ok bool) {
	// Find the assignment where the function literal was defined.
	assignStmt, ok := funcIdent.Obj.Decl.(*ast.AssignStmt)
	if !ok {
		return nil, ok
	}

	funcVariableName := funcIdent.Name

	// Find the index of funcVariableName in assignStmt.Lhs. We need this to
	// find the closure in the corresponding assignStmt.Rhs list.
	assignmentIndex := -1
	for i, expr := range assignStmt.Lhs {
		lhsIdent, ok := expr.(*ast.Ident)
		if !ok {
			continue
		}

		if lhsIdent.Name == funcVariableName {
			assignmentIndex = i
			break
		}
	}

	if assignmentIndex == -1 ||
		assignmentIndex > len(assignStmt.Rhs)-1 {

		return nil, false
	}

	expr := assignStmt.Rhs[assignmentIndex]
	funcLit, ok = expr.(*ast.FuncLit)
	if !ok {
		return nil, ok
	}

	return funcLit, true
}

// checkClosure reports variables used in funcLit that are captured from an
// outer scope.
func checkClosure(pass *analysis.Pass, funcLit *ast.FuncLit) {
	// Get the closure's scope.
	funcScope := pass.TypesInfo.Scopes[funcLit.Type]

	// Build a list of assignments local to funcLit. These will be ignored as
	// shadowed variables.
	localVarDeclarations := findLocalVarDeclarations(funcLit)

	ast.Inspect(
		funcLit,
		func(node ast.Node) bool {
			// fmt.Printf("node: %#v\n", node)

			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Obj == nil {
				return true
			}

			// Ignore shadowed variables.
			for _, declarationIdent := range localVarDeclarations {
				if ident == declarationIdent {
					return true
				}
			}

			// Find out whether `ident` was defined in an outer scope.
			scope, scopeObj := funcScope.LookupParent(ident.Name, ident.NamePos)
			// fmt.Printf("ident: %#v\n\t%#v\n\t%#v\n", ident, scope, scopeObj)

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

			if ident.Name == "copied" {
				fmt.Printf("identifier: %#v, obj: %#v, decl: %#v\n", ident, ident.Obj, ident.Obj.Decl)
				fmt.Printf("  scope: %#v\n", scope)
				fmt.Printf("    Names: %#v\n", scope.Names())
				fmt.Printf("  funcScope: %#v\n", funcScope)
				fmt.Printf("    Names: %#v\n", funcScope.Names())
				fmt.Println()
			}

			za := slices.Index(funcScope.Names(), ident.Name)
			// if za == -1 {
			// 	return true
			// }
			zb := slices.Index(scope.Names(), ident.Name)
			// if zb == -1 {
			// 	return true
			// }

			// TODO: Broken by golang/go a27a525d1b4df74989ac9f6ad10394391fe3eb88
			// Test with golang.org/x/tools@v0.15.0
			// Identifier was defined in a different scope.
			// if funcScope != scope {
			if za != -1 && zb != -1 && funcScope.Names()[za] == scope.Names()[zb] {
				pass.Reportf(
					ident.Pos(),
					"captured reference %s in goroutine closure",
					ident,
				)

				return true
			}

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
	funcLit *ast.FuncLit,
) (declarations []*ast.Ident) {
	declarations = []*ast.Ident{}

	ast.Inspect(
		funcLit,
		func(node ast.Node) bool {
			switch node := node.(type) {
			case *ast.AssignStmt:
				assignments := assignmentDefinitions(node)
				declarations = append(declarations, assignments...)

			case *ast.GenDecl:
				decls := varDeclarations(node)
				if decls != nil {
					declarations = append(declarations, decls...)
				}
			}

			return true
		},
	)

	return declarations
}

// assignmentDefinitions returns the identifiers corresponding to variable
// assignments in assignStmt, or nil if assignStmt does not declare any
// variables.
func assignmentDefinitions(
	assignStmt *ast.AssignStmt,
) (assignments []*ast.Ident) {
	assignments = []*ast.Ident{}

	if assignStmt.Tok != token.DEFINE {
		return nil
	}

	for _, lhs := range assignStmt.Lhs {
		ident, ok := lhs.(*ast.Ident)
		if !ok {
			continue
		}

		if ident == nil {
			continue
		}

		assignments = append(assignments, ident)
	}

	return assignments
}

// varDeclarations returns the identifiers corresponding to variable
// declarations in decl, or nil if decl is not a variable declaration.
func varDeclarations(decl *ast.GenDecl) (declarations []*ast.Ident) {
	if decl.Tok != token.VAR {
		return nil
	}

	declarations = []*ast.Ident{}

	for _, spec := range decl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			return declarations
		}

		for _, ident := range valueSpec.Names {
			declarations = append(declarations, ident)
		}
	}

	return declarations
}
