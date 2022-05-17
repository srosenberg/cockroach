package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func CountLeases(
	ctx context.Context, executor sqlutil.InternalExecutor, versions []IDVersion, at hlc.Timestamp,
) (int, error) {
	__antithesis_instrumentation__.Notify(266038)
	var whereClauses []string
	for _, t := range versions {
		__antithesis_instrumentation__.Notify(266042)
		whereClauses = append(whereClauses,
			fmt.Sprintf(`("descID" = %d AND version = %d AND expiration > $1)`,
				t.ID, t.Version),
		)
	}
	__antithesis_instrumentation__.Notify(266039)

	stmt := fmt.Sprintf(`SELECT count(1) FROM system.public.lease AS OF SYSTEM TIME '%s' WHERE `,
		at.AsOfSystemTime()) +
		strings.Join(whereClauses, " OR ")

	values, err := executor.QueryRowEx(
		ctx, "count-leases", nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		stmt, at.GoTime(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(266043)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(266044)
	}
	__antithesis_instrumentation__.Notify(266040)
	if values == nil {
		__antithesis_instrumentation__.Notify(266045)
		return 0, errors.New("failed to count leases")
	} else {
		__antithesis_instrumentation__.Notify(266046)
	}
	__antithesis_instrumentation__.Notify(266041)
	count := int(tree.MustBeDInt(values[0]))
	return count, nil
}
