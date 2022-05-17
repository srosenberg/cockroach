package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type windowNode struct {
	plan planNode

	columns colinfo.ResultColumns

	windowRender []tree.TypedExpr

	funcs []*windowFuncHolder

	colAndAggContainer windowNodeColAndAggContainer
}

func (n *windowNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(632832)
	panic("windowNode can't be run in local mode")
}

func (n *windowNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(632833)
	panic("windowNode can't be run in local mode")
}

func (n *windowNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(632834)
	panic("windowNode can't be run in local mode")
}

func (n *windowNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(632835)
	n.plan.Close(ctx)
}

var _ tree.TypedExpr = &windowFuncHolder{}
var _ tree.VariableExpr = &windowFuncHolder{}

type windowFuncHolder struct {
	window *windowNode

	expr *tree.FuncExpr
	args []tree.Expr

	argsIdxs     []uint32
	filterColIdx int
	outputColIdx int

	partitionIdxs  []int
	columnOrdering colinfo.ColumnOrdering
	frame          *tree.WindowFrame
}

func (w *windowFuncHolder) samePartition(other *windowFuncHolder) bool {
	__antithesis_instrumentation__.Notify(632836)
	if len(w.partitionIdxs) != len(other.partitionIdxs) {
		__antithesis_instrumentation__.Notify(632839)
		return false
	} else {
		__antithesis_instrumentation__.Notify(632840)
	}
	__antithesis_instrumentation__.Notify(632837)
	for i, p := range w.partitionIdxs {
		__antithesis_instrumentation__.Notify(632841)
		if p != other.partitionIdxs[i] {
			__antithesis_instrumentation__.Notify(632842)
			return false
		} else {
			__antithesis_instrumentation__.Notify(632843)
		}
	}
	__antithesis_instrumentation__.Notify(632838)
	return true
}

func (*windowFuncHolder) Variable() { __antithesis_instrumentation__.Notify(632844) }

func (w *windowFuncHolder) Format(ctx *tree.FmtCtx) {
	__antithesis_instrumentation__.Notify(632845)

	w.expr.Format(ctx)
}

func (w *windowFuncHolder) String() string {
	__antithesis_instrumentation__.Notify(632846)
	return tree.AsString(w)
}

func (w *windowFuncHolder) Walk(v tree.Visitor) tree.Expr {
	__antithesis_instrumentation__.Notify(632847)
	return w
}

func (w *windowFuncHolder) TypeCheck(
	_ context.Context, _ *tree.SemaContext, desired *types.T,
) (tree.TypedExpr, error) {
	__antithesis_instrumentation__.Notify(632848)
	return w, nil
}

func (w *windowFuncHolder) Eval(ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(632849)
	panic("windowFuncHolder should not be evaluated directly")
}

func (w *windowFuncHolder) ResolvedType() *types.T {
	__antithesis_instrumentation__.Notify(632850)
	return w.expr.ResolvedType()
}

type windowNodeColAndAggContainer struct {
	idxMap map[int]int

	sourceInfo *colinfo.DataSourceInfo

	aggFuncs map[int]*tree.FuncExpr

	startAggIdx int
}

func (c *windowNodeColAndAggContainer) IndexedVarEval(
	idx int, ctx *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(632851)
	panic("IndexedVarEval should not be called on windowNodeColAndAggContainer")
}

func (c *windowNodeColAndAggContainer) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(632852)
	if idx >= c.startAggIdx {
		__antithesis_instrumentation__.Notify(632854)
		return c.aggFuncs[idx].ResolvedType()
	} else {
		__antithesis_instrumentation__.Notify(632855)
	}
	__antithesis_instrumentation__.Notify(632853)
	return c.sourceInfo.SourceColumns[idx].Typ
}

func (c *windowNodeColAndAggContainer) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(632856)
	if idx >= c.startAggIdx {
		__antithesis_instrumentation__.Notify(632858)

		return c.aggFuncs[idx]
	} else {
		__antithesis_instrumentation__.Notify(632859)
	}
	__antithesis_instrumentation__.Notify(632857)

	return c.sourceInfo.NodeFormatter(idx)
}
