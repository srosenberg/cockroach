package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type projectSetNode struct {
	source planNode
	projectSetPlanningInfo
}

type projectSetPlanningInfo struct {
	columns colinfo.ResultColumns

	numColsInSource int

	exprs tree.TypedExprs

	numColsPerGen []int
}

func (n *projectSetNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(563352)
	panic("projectSetNode can't be run in local mode")
}

func (n *projectSetNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(563353)
	panic("projectSetNode can't be run in local mode")
}

func (n *projectSetNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(563354)
	panic("projectSetNode can't be run in local mode")
}

func (n *projectSetNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(563355)
	n.source.Close(ctx)
}
