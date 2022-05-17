package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type invertedFilterNode struct {
	input           planNode
	expression      *inverted.SpanExpression
	preFiltererExpr tree.TypedExpr
	preFiltererType *types.T
	invColumn       int
	resultColumns   colinfo.ResultColumns
}

func (n *invertedFilterNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(498779)
	panic("invertedFiltererNode can't be run in local mode")
}
func (n *invertedFilterNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(498780)
	n.input.Close(ctx)
}
func (n *invertedFilterNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(498781)
	panic("invertedFiltererNode can't be run in local mode")
}
func (n *invertedFilterNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(498782)
	panic("invertedFiltererNode can't be run in local mode")
}
