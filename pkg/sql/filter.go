package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type filterNode struct {
	source      planDataSource
	filter      tree.TypedExpr
	ivarHelper  tree.IndexedVarHelper
	reqOrdering ReqOrdering
}

var _ tree.IndexedVarContainer = &filterNode{}

func (f *filterNode) IndexedVarEval(idx int, ctx *tree.EvalContext) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(491500)
	return f.source.plan.Values()[idx].Eval(ctx)
}

func (f *filterNode) IndexedVarResolvedType(idx int) *types.T {
	__antithesis_instrumentation__.Notify(491501)
	return f.source.columns[idx].Typ
}

func (f *filterNode) IndexedVarNodeFormatter(idx int) tree.NodeFormatter {
	__antithesis_instrumentation__.Notify(491502)
	return f.source.columns.NodeFormatter(idx)
}

func (f *filterNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(491503)
	return nil
}

func (f *filterNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(491504)
	panic("filterNode cannot be run in local mode")
}

func (f *filterNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(491505)
	panic("filterNode cannot be run in local mode")
}

func (f *filterNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491506)
	f.source.plan.Close(ctx)
}
