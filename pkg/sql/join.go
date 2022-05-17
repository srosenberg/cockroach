package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type joinNode struct {
	left  planDataSource
	right planDataSource

	pred *joinPredicate

	mergeJoinOrdering colinfo.ColumnOrdering

	reqOrdering ReqOrdering

	columns colinfo.ResultColumns
}

func (p *planner) makeJoinNode(
	left planDataSource, right planDataSource, pred *joinPredicate,
) *joinNode {
	__antithesis_instrumentation__.Notify(498808)
	n := &joinNode{
		left:    left,
		right:   right,
		pred:    pred,
		columns: pred.cols,
	}
	return n
}

func (n *joinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(498809)
	panic("joinNode cannot be run in local mode")
}

func (n *joinNode) Next(params runParams) (res bool, err error) {
	__antithesis_instrumentation__.Notify(498810)
	panic("joinNode cannot be run in local mode")
}

func (n *joinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(498811)
	panic("joinNode cannot be run in local mode")
}

func (n *joinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(498812)
	n.right.plan.Close(ctx)
	n.left.plan.Close(ctx)
}
