package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type indexJoinNode struct {
	input planNode

	keyCols []int

	table *scanNode

	cols []catalog.Column

	resultColumns colinfo.ResultColumns

	reqOrdering ReqOrdering

	limitHint int64
}

func (n *indexJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(496703)
	panic("indexJoinNode cannot be run in local mode")
}

func (n *indexJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(496704)
	panic("indexJoinNode cannot be run in local mode")
}

func (n *indexJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(496705)
	panic("indexJoinNode cannot be run in local mode")
}

func (n *indexJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(496706)
	n.input.Close(ctx)
	n.table.Close(ctx)
}
