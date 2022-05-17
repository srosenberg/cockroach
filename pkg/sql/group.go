package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type groupNode struct {
	columns colinfo.ResultColumns

	plan planNode

	groupCols []int

	groupColOrdering colinfo.ColumnOrdering

	isScalar bool

	funcs []*aggregateFuncHolder

	reqOrdering ReqOrdering
}

func (n *groupNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(492988)
	panic("groupNode cannot be run in local mode")
}

func (n *groupNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(492989)
	panic("groupNode cannot be run in local mode")
}

func (n *groupNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(492990)
	panic("groupNode cannot be run in local mode")
}

func (n *groupNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(492991)
	n.plan.Close(ctx)
}

type aggregateFuncHolder struct {
	funcName string

	argRenderIdxs []int

	filterRenderIdx int

	arguments tree.Datums

	isDistinct bool
}

func newAggregateFuncHolder(
	funcName string, argRenderIdxs []int, arguments tree.Datums, isDistinct bool,
) *aggregateFuncHolder {
	__antithesis_instrumentation__.Notify(492992)
	res := &aggregateFuncHolder{
		funcName:        funcName,
		argRenderIdxs:   argRenderIdxs,
		filterRenderIdx: tree.NoColumnIdx,
		arguments:       arguments,
		isDistinct:      isDistinct,
	}
	return res
}

func (a *aggregateFuncHolder) hasFilter() bool {
	__antithesis_instrumentation__.Notify(492993)
	return a.filterRenderIdx != tree.NoColumnIdx
}
