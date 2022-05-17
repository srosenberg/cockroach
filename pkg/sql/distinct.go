package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
)

type distinctNode struct {
	plan planNode

	distinctOnColIdxs util.FastIntSet

	columnsInOrder util.FastIntSet

	reqOrdering ReqOrdering

	nullsAreDistinct bool

	errorOnDup string
}

func (n *distinctNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(466207)
	panic("distinctNode can't be called in local mode")
}

func (n *distinctNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(466208)
	panic("distinctNode can't be called in local mode")
}

func (n *distinctNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(466209)
	panic("distinctNode can't be called in local mode")
}

func (n *distinctNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(466210)
	n.plan.Close(ctx)
}
