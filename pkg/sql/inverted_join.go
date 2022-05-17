package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type invertedJoinNode struct {
	input planNode
	table *scanNode

	joinType descpb.JoinType

	prefixEqCols []int

	invertedExpr tree.TypedExpr

	columns colinfo.ResultColumns

	onExpr tree.TypedExpr

	isFirstJoinInPairedJoiner bool

	reqOrdering ReqOrdering
}

func (ij *invertedJoinNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(498783)
	panic("invertedJoinNode cannot be run in local mode")
}

func (ij *invertedJoinNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(498784)
	panic("invertedJoinNode cannot be run in local mode")
}

func (ij *invertedJoinNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(498785)
	panic("invertedJoinNode cannot be run in local mode")
}

func (ij *invertedJoinNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(498786)
	ij.input.Close(ctx)
	ij.table.Close(ctx)
}
