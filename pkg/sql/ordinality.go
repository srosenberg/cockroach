package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

type ordinalityNode struct {
	source      planNode
	columns     colinfo.ResultColumns
	reqOrdering ReqOrdering
}

func (o *ordinalityNode) startExec(runParams) error {
	__antithesis_instrumentation__.Notify(551929)
	panic("ordinalityNode can't be run in local mode")
}

func (o *ordinalityNode) Next(params runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(551930)
	panic("ordinalityNode can't be run in local mode")
}

func (o *ordinalityNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(551931)
	panic("ordinalityNode can't be run in local mode")
}

func (o *ordinalityNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(551932)
	o.source.Close(ctx)
}
