package transform

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type IsAggregateVisitor struct {
	Aggregated bool

	searchPath sessiondata.SearchPath
}

var _ tree.Visitor = &IsAggregateVisitor{}

func (v *IsAggregateVisitor) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(602897)
	switch t := expr.(type) {
	case *tree.FuncExpr:
		__antithesis_instrumentation__.Notify(602899)
		if t.IsWindowFunctionApplication() {
			__antithesis_instrumentation__.Notify(602903)

			return true, expr
		} else {
			__antithesis_instrumentation__.Notify(602904)
		}
		__antithesis_instrumentation__.Notify(602900)
		fd, err := t.Func.Resolve(v.searchPath)
		if err != nil {
			__antithesis_instrumentation__.Notify(602905)
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(602906)
		}
		__antithesis_instrumentation__.Notify(602901)
		if fd.Class == tree.AggregateClass {
			__antithesis_instrumentation__.Notify(602907)
			v.Aggregated = true
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(602908)
		}
	case *tree.Subquery:
		__antithesis_instrumentation__.Notify(602902)
		return false, expr
	}
	__antithesis_instrumentation__.Notify(602898)

	return true, expr
}

func (*IsAggregateVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(602909)
	return expr
}
