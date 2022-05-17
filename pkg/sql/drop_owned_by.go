package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
)

type dropOwnedByNode struct {
}

func (p *planner) DropOwnedBy(ctx context.Context) (planNode, error) {
	__antithesis_instrumentation__.Notify(469118)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"DROP OWNED BY",
	); err != nil {
		__antithesis_instrumentation__.Notify(469120)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(469121)
	}
	__antithesis_instrumentation__.Notify(469119)
	telemetry.Inc(sqltelemetry.CreateDropOwnedByCounter())

	return nil, unimplemented.NewWithIssue(55381, "drop owned by is not yet implemented")
}

func (n *dropOwnedByNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(469122)

	return nil
}
func (n *dropOwnedByNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(469123)
	return false, nil
}
func (n *dropOwnedByNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(469124)
	return tree.Datums{}
}
func (n *dropOwnedByNode) Close(context.Context) { __antithesis_instrumentation__.Notify(469125) }
