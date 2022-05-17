package physicalplan

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type ExprContext interface {
	EvalContext() *tree.EvalContext

	IsLocal() bool
}

type fakeExprContext struct{}

var _ ExprContext = fakeExprContext{}

func (fakeExprContext) EvalContext() *tree.EvalContext {
	__antithesis_instrumentation__.Notify(562082)
	return &tree.EvalContext{}
}

func (fakeExprContext) IsLocal() bool {
	__antithesis_instrumentation__.Notify(562083)
	return false
}

func MakeExpression(
	expr tree.TypedExpr, ctx ExprContext, indexVarMap []int,
) (execinfrapb.Expression, error) {
	__antithesis_instrumentation__.Notify(562084)
	if expr == nil {
		__antithesis_instrumentation__.Notify(562091)
		return execinfrapb.Expression{}, nil
	} else {
		__antithesis_instrumentation__.Notify(562092)
	}
	__antithesis_instrumentation__.Notify(562085)
	if ctx == nil {
		__antithesis_instrumentation__.Notify(562093)
		ctx = &fakeExprContext{}
	} else {
		__antithesis_instrumentation__.Notify(562094)
	}
	__antithesis_instrumentation__.Notify(562086)

	evalCtx := ctx.EvalContext()
	subqueryVisitor := &evalAndReplaceSubqueryVisitor{
		evalCtx: evalCtx,
	}
	outExpr, _ := tree.WalkExpr(subqueryVisitor, expr)
	if subqueryVisitor.err != nil {
		__antithesis_instrumentation__.Notify(562095)
		return execinfrapb.Expression{}, subqueryVisitor.err
	} else {
		__antithesis_instrumentation__.Notify(562096)
	}
	__antithesis_instrumentation__.Notify(562087)
	expr = outExpr.(tree.TypedExpr)

	if indexVarMap != nil {
		__antithesis_instrumentation__.Notify(562097)

		expr = RemapIVarsInTypedExpr(expr, indexVarMap)
	} else {
		__antithesis_instrumentation__.Notify(562098)
	}
	__antithesis_instrumentation__.Notify(562088)
	expression := execinfrapb.Expression{LocalExpr: expr}
	if ctx.IsLocal() {
		__antithesis_instrumentation__.Notify(562099)
		return expression, nil
	} else {
		__antithesis_instrumentation__.Notify(562100)
	}
	__antithesis_instrumentation__.Notify(562089)

	fmtCtx := execinfrapb.ExprFmtCtxBase(evalCtx)
	fmtCtx.FormatNode(expr)
	if log.V(1) {
		__antithesis_instrumentation__.Notify(562101)
		log.Infof(evalCtx.Ctx(), "Expr %s:\n%s", fmtCtx.String(), tree.ExprDebugString(expr))
	} else {
		__antithesis_instrumentation__.Notify(562102)
	}
	__antithesis_instrumentation__.Notify(562090)
	expression.Expr = fmtCtx.CloseAndGetString()
	return expression, nil
}

type evalAndReplaceSubqueryVisitor struct {
	evalCtx *tree.EvalContext
	err     error
}

var _ tree.Visitor = &evalAndReplaceSubqueryVisitor{}

func (e *evalAndReplaceSubqueryVisitor) VisitPre(expr tree.Expr) (bool, tree.Expr) {
	__antithesis_instrumentation__.Notify(562103)
	switch expr := expr.(type) {
	case *tree.Subquery:
		__antithesis_instrumentation__.Notify(562104)
		val, err := e.evalCtx.Planner.EvalSubquery(expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(562108)
			e.err = err
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(562109)
		}
		__antithesis_instrumentation__.Notify(562105)
		newExpr := tree.Expr(val)
		typ := expr.ResolvedType()
		if _, isTuple := val.(*tree.DTuple); !isTuple && func() bool {
			__antithesis_instrumentation__.Notify(562110)
			return typ.Family() != types.UnknownFamily == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(562111)
			return typ.Family() != types.TupleFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(562112)
			newExpr = tree.NewTypedCastExpr(val, typ)
		} else {
			__antithesis_instrumentation__.Notify(562113)
		}
		__antithesis_instrumentation__.Notify(562106)
		return false, newExpr
	default:
		__antithesis_instrumentation__.Notify(562107)
		return true, expr
	}
}

func (evalAndReplaceSubqueryVisitor) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(562114)
	return expr
}

func RemapIVarsInTypedExpr(expr tree.TypedExpr, indexVarMap []int) tree.TypedExpr {
	__antithesis_instrumentation__.Notify(562115)
	v := &ivarRemapper{indexVarMap: indexVarMap}
	newExpr, _ := tree.WalkExpr(v, expr)
	return newExpr.(tree.TypedExpr)
}

type ivarRemapper struct {
	indexVarMap []int
}

var _ tree.Visitor = &ivarRemapper{}

func (v *ivarRemapper) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(562116)
	if ivar, ok := expr.(*tree.IndexedVar); ok {
		__antithesis_instrumentation__.Notify(562118)
		newIvar := *ivar
		newIvar.Idx = v.indexVarMap[ivar.Idx]
		return false, &newIvar
	} else {
		__antithesis_instrumentation__.Notify(562119)
	}
	__antithesis_instrumentation__.Notify(562117)
	return true, expr
}

func (*ivarRemapper) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(562120)
	return expr
}
