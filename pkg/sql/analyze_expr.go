package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func (p *planner) analyzeExpr(
	ctx context.Context,
	raw tree.Expr,
	source *colinfo.DataSourceInfo,
	iVarHelper tree.IndexedVarHelper,
	expectedType *types.T,
	requireType bool,
	typingContext string,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(245227)

	resolved := raw
	if source != nil {
		__antithesis_instrumentation__.Notify(245231)
		var err error
		resolved, err = p.resolveNames(raw, source, iVarHelper)
		if err != nil {
			__antithesis_instrumentation__.Notify(245232)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(245233)
		}
	} else {
		__antithesis_instrumentation__.Notify(245234)
	}
	__antithesis_instrumentation__.Notify(245228)

	var typedExpr tree.TypedExpr
	var err error
	p.semaCtx.IVarContainer = iVarHelper.Container()
	if requireType {
		__antithesis_instrumentation__.Notify(245235)
		typedExpr, err = tree.TypeCheckAndRequire(ctx, resolved, &p.semaCtx,
			expectedType, typingContext)
	} else {
		__antithesis_instrumentation__.Notify(245236)
		typedExpr, err = tree.TypeCheck(ctx, resolved, &p.semaCtx, expectedType)
	}
	__antithesis_instrumentation__.Notify(245229)
	p.semaCtx.IVarContainer = nil
	if err != nil {
		__antithesis_instrumentation__.Notify(245237)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(245238)
	}
	__antithesis_instrumentation__.Notify(245230)

	return p.txCtx.NormalizeExpr(p.EvalContext(), typedExpr)
}
