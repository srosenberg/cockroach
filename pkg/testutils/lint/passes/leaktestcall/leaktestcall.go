// Package leaktestcall defines an Analyzer that detects correct use
// of leaktest.AfterTest(t).
package leaktestcall

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "leaktestcall",
	Doc:      "Check that the closure returned by leaktest.AfterFunc(t) is called",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(644856)
	astInspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	filter := []ast.Node{
		(*ast.DeferStmt)(nil),
	}

	astInspector.Preorder(filter, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(644858)
		def := n.(*ast.DeferStmt)
		switch funCall := def.Call.Fun.(type) {
		case *ast.SelectorExpr:
			__antithesis_instrumentation__.Notify(644859)
			packageIdent, ok := funCall.X.(*ast.Ident)
			if !ok {
				__antithesis_instrumentation__.Notify(644861)
				return
			} else {
				__antithesis_instrumentation__.Notify(644862)
			}
			__antithesis_instrumentation__.Notify(644860)
			if packageIdent.Name == "leaktest" && func() bool {
				__antithesis_instrumentation__.Notify(644863)
				return funCall.Sel != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(644864)
				return funCall.Sel.Name == "AfterTest" == true
			}() == true {
				__antithesis_instrumentation__.Notify(644865)
				pass.Reportf(def.Call.Pos(), "leaktest.AfterTest return not called")
			} else {
				__antithesis_instrumentation__.Notify(644866)
			}
		}
	})
	__antithesis_instrumentation__.Notify(644857)

	return nil, nil
}
