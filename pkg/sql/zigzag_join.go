package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type zigzagJoinNode struct {
	sides []zigzagJoinSide

	columns colinfo.ResultColumns

	onCond tree.TypedExpr

	reqOrdering ReqOrdering
}

type zigzagJoinSide struct {
	scan *scanNode

	eqCols []int

	fixedVals *valuesNode
}

func (zj *zigzagJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(632865)
	panic("zigzag joins cannot be executed outside of distsql")
}

func (zj *zigzagJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(632866)
	panic("zigzag joins cannot be executed outside of distsql")
}

func (zj *zigzagJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(632867)
	panic("zigzag joins cannot be executed outside of distsql")
}

func (zj *zigzagJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(632868)
}
