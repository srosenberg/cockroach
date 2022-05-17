package colexecargs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var exprHelperPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(286135)
		return &ExprHelper{}
	},
}

func NewExprHelper() *ExprHelper {
	__antithesis_instrumentation__.Notify(286136)
	return exprHelperPool.Get().(*ExprHelper)
}

type ExprHelper struct {
	helper  execinfrapb.ExprHelper
	SemaCtx *tree.SemaContext
}

func (h *ExprHelper) ProcessExpr(
	expr execinfrapb.Expression, evalCtx *tree.EvalContext, typs []*types.T,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(286137)
	if expr.LocalExpr != nil {
		__antithesis_instrumentation__.Notify(286139)
		return expr.LocalExpr, nil
	} else {
		__antithesis_instrumentation__.Notify(286140)
	}
	__antithesis_instrumentation__.Notify(286138)
	h.helper.Types = typs
	tempVars := tree.MakeIndexedVarHelper(&h.helper, len(typs))
	return execinfrapb.DeserializeExpr(expr.Expr, h.SemaCtx, evalCtx, &tempVars)
}
