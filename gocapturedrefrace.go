package gocapturedrefrace

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
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

				// TODO: How to get types.Func or {ast,types}.Scope of function literal?
				funcIdent, ok := goStmt.Call.Fun.(*ast.Ident)
				funcObj := pass.TypesInfo.ObjectOf(funcIdent)
				if funcObj != nil {
					theFunc, ok := funcObj.(*types.Func)
					if ok {
						fmt.Printf("func scope: %#v\n", theFunc.Scope())
					} else {
						fmt.Println("func scope: error")
					}
				} else {
					fmt.Println("funcObj: error")
				}

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
	formalParams := []*ast.Object{}
	for _, field := range funcLit.Type.Params.List {
		formalParams = append(formalParams, field.Names[0].Obj)
	}
	fmt.Printf("formalParams: %#v\n", formalParams)
	// TODO: Ensure argument types are not references
	// TODO: goStmt.Call.Args should also give us something like this.

	// TODO: Build a list of variables created in the closure
	assignments := assignmentsInFunc(pass, funcLit)
	fmt.Printf("variable declarations: %#v\n", assignments)
	// TODO: Use ast.GenDecl instead
	// ast.Scope?

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

			for _, param := range formalParams {
				if ident.Obj == param {
					return true
				}
			}

			pass.Reportf(
				ident.Pos(),
				"variable found %q",
				ident,
			)

			return true
		},
	)
}

func assignmentsInFunc(
	pass *analysis.Pass,
	funcLit *ast.FuncLit,
) []string {
	assignments := []string{}
	// ) []*ast.Object {
	// 	assignments := []*ast.Object{}

	ast.Inspect(
		funcLit,
		func(node ast.Node) bool {
			// decl, ok := node.(*ast.GenDecl)
			// if !ok {
			// 	return true
			// }
			//
			// fmt.Printf("decl: %#v\n", decl)
			//
			// if decl.Tok != token.VAR {
			// 	return true
			// }
			//
			// for _, spec := range decl.Specs {
			// 	valueSpec, ok := spec.(*ast.ValueSpec)
			// 	if !ok {
			// 		return true
			// 	}
			//
			// 	fmt.Printf("valueSpec: %#v\n", valueSpec)
			//
			// 	assignments = append(assignments, valueSpec.Names[0].Obj)
			// }

			// decl, ok := node.(*ast.DeclStmt)
			// if !ok {
			// 	return true
			// }
			//
			// fmt.Printf("decl: %#v\n", decl)

			ident, ok := node.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Obj == nil || ident.Obj.Decl == nil {
				return true
			}

			assignment, ok := ident.Obj.Decl.(*ast.AssignStmt)
			if !ok {
				return true
			}

			// fmt.Printf("assignment: %#v\n", assignment.Tok)
			if assignment.Tok == token.DEFINE {
				fmt.Printf("assignment: %v is DEFINE\n", ident.Name)
			} else if assignment.Tok == token.ASSIGN {
				fmt.Printf("assignment: %v is ASSIGN\n", ident.Name)
			} else {
				fmt.Printf("assignment: %v\n", assignment.Tok)
			}

			if pass.TypesInfo.Defs[ident] != nil {
				fmt.Println("DEFINE:", ident)
			} else {
				fmt.Println("ASSIGN:", ident)
			}

			obj := pass.TypesInfo.ObjectOf(ident)
			if obj != nil {
				fmt.Printf("obj: %#v\n", obj)

				theVar, ok := obj.(*types.Var)
				if !ok {
					return true
				}

				fmt.Printf("obj origin: %#v\n", theVar.Origin())
				fmt.Printf("obj parent: %#v\n", theVar.Parent())
			}

			assignments = append(assignments, ident.Name)

			return true
		},
	)

	return assignments
}
