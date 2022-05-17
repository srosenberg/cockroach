package transform

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

type ExprTransformContext struct {
	normalizeVisitor   tree.NormalizeVisitor
	isAggregateVisitor IsAggregateVisitor
}

func (t *ExprTransformContext) NormalizeExpr(
	ctx *tree.EvalContext, typedExpr tree.TypedExpr,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(602910)
	if ctx.SkipNormalize {
		__antithesis_instrumentation__.Notify(602913)
		return typedExpr, nil
	} else {
		__antithesis_instrumentation__.Notify(602914)
	}
	__antithesis_instrumentation__.Notify(602911)
	t.normalizeVisitor = tree.MakeNormalizeVisitor(ctx)
	expr, _ := tree.WalkExpr(&t.normalizeVisitor, typedExpr)
	if err := t.normalizeVisitor.Err(); err != nil {
		__antithesis_instrumentation__.Notify(602915)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(602916)
	}
	__antithesis_instrumentation__.Notify(602912)
	return expr.(tree.TypedExpr), nil
}

func (t *ExprTransformContext) AggregateInExpr(
	expr tree.Expr, searchPath sessiondata.SearchPath,
) bool {
	__antithesis_instrumentation__.Notify(602917)
	if expr == nil {
		__antithesis_instrumentation__.Notify(602919)
		return false
	} else {
		__antithesis_instrumentation__.Notify(602920)
	}
	__antithesis_instrumentation__.Notify(602918)

	t.isAggregateVisitor = IsAggregateVisitor{
		searchPath: searchPath,
	}
	tree.WalkExprConst(&t.isAggregateVisitor, expr)
	return t.isAggregateVisitor.Aggregated
}
