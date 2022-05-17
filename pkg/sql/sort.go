package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type sortNode struct {
	plan     planNode
	ordering colinfo.ColumnOrdering

	alreadyOrderedPrefix int
}

func (n *sortNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(623509)
	panic("sortNode cannot be run in local mode")
}

func (n *sortNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(623510)
	panic("sortNode cannot be run in local mode")
}

func (n *sortNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(623511)
	panic("sortNode cannot be run in local mode")
}

func (n *sortNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(623512)
	n.plan.Close(ctx)
}
