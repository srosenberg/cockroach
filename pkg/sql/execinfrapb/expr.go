package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/transform"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

type ivarBinder struct {
	h   *tree.IndexedVarHelper
	err error
}

func (v *ivarBinder) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	__antithesis_instrumentation__.Notify(477969)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(477972)
		return false, expr
	} else {
		__antithesis_instrumentation__.Notify(477973)
	}
	__antithesis_instrumentation__.Notify(477970)
	if ivar, ok := expr.(*tree.IndexedVar); ok {
		__antithesis_instrumentation__.Notify(477974)
		newVar, err := v.h.BindIfUnbound(ivar)
		if err != nil {
			__antithesis_instrumentation__.Notify(477976)
			v.err = err
			return false, expr
		} else {
			__antithesis_instrumentation__.Notify(477977)
		}
		__antithesis_instrumentation__.Notify(477975)
		return false, newVar
	} else {
		__antithesis_instrumentation__.Notify(477978)
	}
	__antithesis_instrumentation__.Notify(477971)
	return true, expr
}

func (*ivarBinder) VisitPost(expr tree.Expr) tree.Expr {
	__antithesis_instrumentation__.Notify(477979)
	return expr
}

func processExpression(
	exprSpec Expression,
	evalCtx *tree.EvalContext,
	semaCtx *tree.SemaContext,
	h *tree.IndexedVarHelper,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(477980)
	if exprSpec.Expr == "" {
		__antithesis_instrumentation__.Notify(477986)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(477987)
	}
	__antithesis_instrumentation__.Notify(477981)
	expr, err := parser.ParseExprWithInt(
		exprSpec.Expr,
		parser.NakedIntTypeFromDefaultIntSize(evalCtx.SessionData().DefaultIntSize),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(477988)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(477989)
	}
	__antithesis_instrumentation__.Notify(477982)

	v := ivarBinder{h: h, err: nil}
	expr, _ = tree.WalkExpr(&v, expr)
	if v.err != nil {
		__antithesis_instrumentation__.Notify(477990)
		return nil, v.err
	} else {
		__antithesis_instrumentation__.Notify(477991)
	}
	__antithesis_instrumentation__.Notify(477983)

	semaCtx.IVarContainer = h.Container()

	typedExpr, err := tree.TypeCheck(evalCtx.Context, expr, semaCtx, types.Any)
	if err != nil {
		__antithesis_instrumentation__.Notify(477992)

		return nil, errors.NewAssertionErrorWithWrappedErrf(err, "%s", expr)
	} else {
		__antithesis_instrumentation__.Notify(477993)
	}
	__antithesis_instrumentation__.Notify(477984)

	c := tree.MakeConstantEvalVisitor(evalCtx)
	expr, _ = tree.WalkExpr(&c, typedExpr)
	if err := c.Err(); err != nil {
		__antithesis_instrumentation__.Notify(477994)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(477995)
	}
	__antithesis_instrumentation__.Notify(477985)

	return expr.(tree.TypedExpr), nil
}

type ExprHelper struct {
	_ util.NoCopy

	Expr tree.TypedExpr

	Vars tree.IndexedVarHelper

	evalCtx *tree.EvalContext

	Types      []*types.T
	Row        rowenc.EncDatumRow
	datumAlloc tree.DatumAlloc
}

func (eh *ExprHelper) String() string {
	__antithesis_instrumentation__.Notify(477996)
	if eh.Expr == nil {
		__antithesis_instrumentation__.Notify(477998)
		return "none"
	} else {
		__antithesis_instrumentation__.Notify(477999)
	}
	__antithesis_instrumentation__.Notify(477997)
	return eh.Expr.String()
}

var _ tree.IndexedVarContainer = &ExprHelper{}

func (eh *ExprHelper) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(478000)
	return eh.Types[idx]
}

func (eh *ExprHelper) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(478001)
	err := eh.Row[idx].EnsureDecoded(eh.Types[idx], &eh.datumAlloc)
	if err != nil {
		__antithesis_instrumentation__.Notify(478003)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(478004)
	}
	__antithesis_instrumentation__.Notify(478002)
	return eh.Row[idx].Datum.Eval(ctx)
}

func (eh *ExprHelper) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(478005)
	n := tree.Name(fmt.Sprintf("$%d", idx))
	return &n
}

func DeserializeExpr(
	expr string, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext, vars *tree.IndexedVarHelper,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(478006)
	if expr == "" {
		__antithesis_instrumentation__.Notify(478010)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(478011)
	}
	__antithesis_instrumentation__.Notify(478007)

	deserializedExpr, err := processExpression(Expression{Expr: expr}, evalCtx, semaCtx, vars)
	if err != nil {
		__antithesis_instrumentation__.Notify(478012)
		return deserializedExpr, err
	} else {
		__antithesis_instrumentation__.Notify(478013)
	}
	__antithesis_instrumentation__.Notify(478008)
	var t transform.ExprTransformContext
	if t.AggregateInExpr(deserializedExpr, evalCtx.SessionData().SearchPath) {
		__antithesis_instrumentation__.Notify(478014)
		return nil, errors.Errorf("expression '%s' has aggregate", deserializedExpr)
	} else {
		__antithesis_instrumentation__.Notify(478015)
	}
	__antithesis_instrumentation__.Notify(478009)
	return deserializedExpr, nil
}

func (eh *ExprHelper) Init(
	expr Expression, types []*types.T, semaCtx *tree.SemaContext, evalCtx *tree.EvalContext,
) error {
	__antithesis_instrumentation__.Notify(478016)
	if expr.Empty() {
		__antithesis_instrumentation__.Notify(478020)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(478021)
	}
	__antithesis_instrumentation__.Notify(478017)
	eh.evalCtx = evalCtx
	eh.Types = types
	eh.Vars = tree.MakeIndexedVarHelper(eh, len(types))

	if expr.LocalExpr != nil {
		__antithesis_instrumentation__.Notify(478022)
		eh.Expr = expr.LocalExpr

		eh.Vars.Rebind(eh.Expr)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(478023)
	}
	__antithesis_instrumentation__.Notify(478018)
	if semaCtx.TypeResolver != nil {
		__antithesis_instrumentation__.Notify(478024)
		for _, t := range types {
			__antithesis_instrumentation__.Notify(478025)
			if err := typedesc.EnsureTypeIsHydrated(evalCtx.Context, t, semaCtx.TypeResolver.(catalog.TypeDescriptorResolver)); err != nil {
				__antithesis_instrumentation__.Notify(478026)
				return err
			} else {
				__antithesis_instrumentation__.Notify(478027)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(478028)
	}
	__antithesis_instrumentation__.Notify(478019)
	var err error
	eh.Expr, err = DeserializeExpr(expr.Expr, semaCtx, evalCtx, &eh.Vars)
	return err
}

func (eh *ExprHelper) EvalFilter(row rowenc.EncDatumRow) (bool, error) {
	__antithesis_instrumentation__.Notify(478029)
	eh.Row = row
	eh.evalCtx.PushIVarContainer(eh)
	pass, err := RunFilter(eh.Expr, eh.evalCtx)
	eh.evalCtx.PopIVarContainer()
	return pass, err
}

func RunFilter(filter tree.TypedExpr, evalCtx *tree.EvalContext) (bool, error) {
	__antithesis_instrumentation__.Notify(478030)
	if filter == nil {
		__antithesis_instrumentation__.Notify(478033)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(478034)
	}
	__antithesis_instrumentation__.Notify(478031)

	d, err := filter.Eval(evalCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(478035)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(478036)
	}
	__antithesis_instrumentation__.Notify(478032)

	return d == tree.DBoolTrue, nil
}

func (eh *ExprHelper) Eval(row rowenc.EncDatumRow) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(478037)
	eh.Row = row

	eh.evalCtx.PushIVarContainer(eh)
	d, err := eh.Expr.Eval(eh.evalCtx)
	eh.evalCtx.PopIVarContainer()
	return d, err
}
