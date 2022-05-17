package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type topKNode struct {
	plan     planNode
	k        int64
	ordering colinfo.ColumnOrdering

	alreadyOrderedPrefix int
}

func (n *topKNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(628402)
	panic("topKNode cannot be run in local mode")
}

func (n *topKNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(628403)
	panic("topKNode cannot be run in local mode")
}

func (n *topKNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(628404)
	panic("topKNode cannot be run in local mode")
}

func (n *topKNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(628405)
	n.plan.Close(ctx)
}
