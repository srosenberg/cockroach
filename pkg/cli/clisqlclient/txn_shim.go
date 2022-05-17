package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
)

type sqlTxnShim struct {
	conn *sqlConn
}

var _ crdb.Tx = sqlTxnShim{}

func (t sqlTxnShim) Commit(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(28975)
	return t.conn.Exec(ctx, `COMMIT`)
}

func (t sqlTxnShim) Rollback(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(28976)
	return t.conn.Exec(ctx, `ROLLBACK`)
}

func (t sqlTxnShim) Exec(ctx context.Context, query string, values ...interface{}) error {
	__antithesis_instrumentation__.Notify(28977)
	if len(values) != 0 {
		__antithesis_instrumentation__.Notify(28979)
		panic("sqlTxnShim.ExecContext must not be called with values")
	} else {
		__antithesis_instrumentation__.Notify(28980)
	}
	__antithesis_instrumentation__.Notify(28978)
	return t.conn.Exec(ctx, query)
}
