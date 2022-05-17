// Package returncheck defines an Analyzer that detects unused or
// discarded roachpb.Error objects.
package returncheck

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "returncheck",
	Doc:      "`returncheck` : `roachpb.Error` :: `errcheck` : (stdlib)`error`",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runAnalyzer,
}

func runAnalyzer(pass *analysis.Pass) (interface{}, error) {
	__antithesis_instrumentation__.Notify(645111)
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	inspect.Preorder([]ast.Node{
		(*ast.AssignStmt)(nil),
		(*ast.DeferStmt)(nil),
		(*ast.ExprStmt)(nil),
		(*ast.GoStmt)(nil),
	}, func(n ast.Node) {
		__antithesis_instrumentation__.Notify(645113)
		switch stmt := n.(type) {
		case *ast.AssignStmt:
			__antithesis_instrumentation__.Notify(645114)

			for i := 0; i < len(stmt.Lhs); i++ {
				__antithesis_instrumentation__.Notify(645118)
				if id, ok := stmt.Lhs[i].(*ast.Ident); ok && func() bool {
					__antithesis_instrumentation__.Notify(645119)
					return id.Name == "_" == true
				}() == true {
					__antithesis_instrumentation__.Notify(645120)
					var rhs ast.Expr
					if len(stmt.Rhs) == 1 {
						__antithesis_instrumentation__.Notify(645122)

						rhs = stmt.Rhs[0]
					} else {
						__antithesis_instrumentation__.Notify(645123)

						rhs = stmt.Rhs[i]
					}
					__antithesis_instrumentation__.Notify(645121)
					if call, ok := rhs.(*ast.CallExpr); ok {
						__antithesis_instrumentation__.Notify(645124)
						recordUnchecked(pass, call, i)
					} else {
						__antithesis_instrumentation__.Notify(645125)
					}
				} else {
					__antithesis_instrumentation__.Notify(645126)
				}
			}
		case *ast.ExprStmt:
			__antithesis_instrumentation__.Notify(645115)
			if call, ok := stmt.X.(*ast.CallExpr); ok {
				__antithesis_instrumentation__.Notify(645127)
				recordUnchecked(pass, call, -1)
			} else {
				__antithesis_instrumentation__.Notify(645128)
			}
		case *ast.GoStmt:
			__antithesis_instrumentation__.Notify(645116)
			recordUnchecked(pass, stmt.Call, -1)
		case *ast.DeferStmt:
			__antithesis_instrumentation__.Notify(645117)
			recordUnchecked(pass, stmt.Call, -1)
		}
	})
	__antithesis_instrumentation__.Notify(645112)
	return nil, nil
}

func recordUnchecked(pass *analysis.Pass, call *ast.CallExpr, pos int) {
	__antithesis_instrumentation__.Notify(645129)
	isTarget := false
	switch t := pass.TypesInfo.Types[call].Type.(type) {
	case *types.Named:
		__antithesis_instrumentation__.Notify(645131)
		isTarget = isTargetType(t)
	case *types.Pointer:
		__antithesis_instrumentation__.Notify(645132)
		isTarget = isTargetType(t.Elem())
	case *types.Tuple:
		__antithesis_instrumentation__.Notify(645133)
		for i := 0; i < t.Len(); i++ {
			__antithesis_instrumentation__.Notify(645134)
			if pos >= 0 && func() bool {
				__antithesis_instrumentation__.Notify(645136)
				return i != pos == true
			}() == true {
				__antithesis_instrumentation__.Notify(645137)
				continue
			} else {
				__antithesis_instrumentation__.Notify(645138)
			}
			__antithesis_instrumentation__.Notify(645135)
			switch et := t.At(i).Type().(type) {
			case *types.Named:
				__antithesis_instrumentation__.Notify(645139)
				isTarget = isTargetType(et)
			case *types.Pointer:
				__antithesis_instrumentation__.Notify(645140)
				isTarget = isTargetType(et.Elem())
			}
		}
	}
	__antithesis_instrumentation__.Notify(645130)

	if isTarget {
		__antithesis_instrumentation__.Notify(645141)
		pass.Reportf(call.Pos(), "unchecked roachpb.Error value")
	} else {
		__antithesis_instrumentation__.Notify(645142)
	}
}

func isTargetType(t types.Type) bool {
	__antithesis_instrumentation__.Notify(645143)
	return t.String() == "github.com/cockroachdb/cockroach/pkg/roachpb.Error"
}
