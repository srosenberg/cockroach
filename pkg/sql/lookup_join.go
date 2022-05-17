package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type lookupJoinNode struct {
	input planNode
	table *scanNode

	joinType descpb.JoinType

	eqCols []int

	eqColsAreKey bool

	lookupExpr tree.TypedExpr

	remoteLookupExpr tree.TypedExpr

	columns colinfo.ResultColumns

	onCond tree.TypedExpr

	isFirstJoinInPairedJoiner  bool
	isSecondJoinInPairedJoiner bool

	reqOrdering ReqOrdering

	limitHint int64
}

func (lj *lookupJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(501747)
	panic("lookupJoinNode cannot be run in local mode")
}

func (lj *lookupJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(501748)
	panic("lookupJoinNode cannot be run in local mode")
}

func (lj *lookupJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(501749)
	panic("lookupJoinNode cannot be run in local mode")
}

func (lj *lookupJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(501750)
	lj.input.Close(ctx)
	lj.table.Close(ctx)
}
