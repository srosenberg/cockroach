package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type zeroNode struct {
	columns colinfo.ResultColumns
}

func newZeroNode(columns colinfo.ResultColumns) *zeroNode {
	__antithesis_instrumentation__.Notify(632860)
	return &zeroNode{columns: columns}
}

func (*zeroNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(632861)
	return nil
}
func (*zeroNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(632862)
	return false, nil
}
func (*zeroNode) Values() tree.Datums   { __antithesis_instrumentation__.Notify(632863); return nil }
func (*zeroNode) Close(context.Context) { __antithesis_instrumentation__.Notify(632864) }
