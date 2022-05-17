package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
)

func (p *planner) Deallocate(ctx context.Context, s *tree.Deallocate) (planNode, error) {
	__antithesis_instrumentation__.Notify(465370)
	if s.Name == "" {
		__antithesis_instrumentation__.Notify(465372)
		p.preparedStatements.DeleteAll(ctx)
	} else {
		__antithesis_instrumentation__.Notify(465373)
		if found := p.preparedStatements.Delete(ctx, string(s.Name)); !found {
			__antithesis_instrumentation__.Notify(465374)
			return nil, pgerror.Newf(pgcode.InvalidSQLStatementName,
				"prepared statement %q does not exist", s.Name)
		} else {
			__antithesis_instrumentation__.Notify(465375)
		}
	}
	__antithesis_instrumentation__.Notify(465371)
	return newZeroNode(nil), nil
}
