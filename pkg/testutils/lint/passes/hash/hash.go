// Package hash defines an Analyzer that detects correct use of hash.Hash.
package hash

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
)

const Doc = `check for correct use of hash.Hash`

var Analyzer = &analysis.Analyzer{
	Name:     "hash",
	Doc:      Doc,
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644802)
	selectorIsHash := func(s *ast.SelectorExpr) bool {
		__antithesis_instrumentation__.Notify(644805)
		tv, ok := pass.TypesInfo.Types[s.X]
		if !ok {
			__antithesis_instrumentation__.Notify(644809)
			return false
		} else {
			__antithesis_instrumentation__.Notify(644810)
		}
		__antithesis_instrumentation__.Notify(644806)
		named, ok := tv.Type.(*types.Named)
		if !ok {
			__antithesis_instrumentation__.Notify(644811)
			return false
		} else {
			__antithesis_instrumentation__.Notify(644812)
		}
		__antithesis_instrumentation__.Notify(644807)
		if named.Obj().Type().String() != "hash.Hash" {
			__antithesis_instrumentation__.Notify(644813)
			return false
		} else {
			__antithesis_instrumentation__.Notify(644814)
		}
		__antithesis_instrumentation__.Notify(644808)
		return true
	}
	__antithesis_instrumentation__.Notify(644803)

	stack := make([]ast.Node, 0, 32)
	forAllFiles(pass.Files, func(n ast.Node) bool {
		__antithesis_instrumentation__.Notify(644815)
		if n == nil {
			__antithesis_instrumentation__.Notify(644825)
			stack = stack[:len(stack)-1]
			return true
		} else {
			__antithesis_instrumentation__.Notify(644826)
		}
		__antithesis_instrumentation__.Notify(644816)
		stack = append(stack, n)

		selExpr, ok := n.(*ast.SelectorExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(644827)
			return true
		} else {
			__antithesis_instrumentation__.Notify(644828)
		}
		__antithesis_instrumentation__.Notify(644817)
		if selExpr.Sel.Name != "Sum" {
			__antithesis_instrumentation__.Notify(644829)
			return true
		} else {
			__antithesis_instrumentation__.Notify(644830)
		}
		__antithesis_instrumentation__.Notify(644818)
		if !selectorIsHash(selExpr) {
			__antithesis_instrumentation__.Notify(644831)
			return true
		} else {
			__antithesis_instrumentation__.Notify(644832)
		}
		__antithesis_instrumentation__.Notify(644819)
		callExpr, ok := stack[len(stack)-2].(*ast.CallExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(644833)
			return true
		} else {
			__antithesis_instrumentation__.Notify(644834)
		}
		__antithesis_instrumentation__.Notify(644820)
		if len(callExpr.Args) != 1 {
			__antithesis_instrumentation__.Notify(644835)
			return true
		} else {
			__antithesis_instrumentation__.Notify(644836)
		}
		__antithesis_instrumentation__.Notify(644821)

		var nilArg bool
		if id, ok := callExpr.Args[0].(*ast.Ident); ok && func() bool {
			__antithesis_instrumentation__.Notify(644837)
			return id.Name == "nil" == true
		}() == true {
			__antithesis_instrumentation__.Notify(644838)
			nilArg = true
		} else {
			__antithesis_instrumentation__.Notify(644839)
		}
		__antithesis_instrumentation__.Notify(644822)

		var retUnused bool
	Switch:
		switch t := stack[len(stack)-3].(type) {
		case *ast.AssignStmt:
			__antithesis_instrumentation__.Notify(644840)
			for i := range t.Rhs {
				__antithesis_instrumentation__.Notify(644844)
				if t.Rhs[i] == stack[len(stack)-2] {
					__antithesis_instrumentation__.Notify(644845)
					if id, ok := t.Lhs[i].(*ast.Ident); ok && func() bool {
						__antithesis_instrumentation__.Notify(644847)
						return id.Name == "_" == true
					}() == true {
						__antithesis_instrumentation__.Notify(644848)

						retUnused = true
					} else {
						__antithesis_instrumentation__.Notify(644849)
					}
					__antithesis_instrumentation__.Notify(644846)
					break Switch
				} else {
					__antithesis_instrumentation__.Notify(644850)
				}
			}
			__antithesis_instrumentation__.Notify(644841)
			panic("unreachable")
		case *ast.ExprStmt:
			__antithesis_instrumentation__.Notify(644842)

			retUnused = true
		default:
			__antithesis_instrumentation__.Notify(644843)
		}
		__antithesis_instrumentation__.Notify(644823)

		if !nilArg && func() bool {
			__antithesis_instrumentation__.Notify(644851)
			return !retUnused == true
		}() == true {
			__antithesis_instrumentation__.Notify(644852)
			pass.Reportf(callExpr.Pos(), "probable misuse of hash.Hash.Sum: "+
				"provide parameter or use return value, but not both")
		} else {
			__antithesis_instrumentation__.Notify(644853)
		}
		__antithesis_instrumentation__.Notify(644824)
		return true
	})
	__antithesis_instrumentation__.Notify(644804)

	return nil, nil
}

func forAllFiles(files []*ast.File, fn func(node ast.Node) bool) {
	__antithesis_instrumentation__.Notify(644854)
	for _, f := range files {
		__antithesis_instrumentation__.Notify(644855)
		ast.Inspect(f, fn)
	}
}
