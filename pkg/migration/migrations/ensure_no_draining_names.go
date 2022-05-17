package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
)

const query = `
WITH
	cte1
		AS (
			SELECT
				crdb_internal.pb_to_json('cockroach.sql.sqlbase.Descriptor', descriptor, false) AS d
			FROM
				system.descriptor
			ORDER BY
				id ASC
		),
	cte2 AS (SELECT COALESCE(d->'table', d->'database', d->'schema', d->'type') AS d FROM cte1)
SELECT
	(d->'id')::INT8 AS id, d->>'name' AS name
FROM
	cte2
WHERE
	COALESCE(json_array_length(d->'drainingNames'), 0) > 0
LIMIT 1;
`

func ensureNoDrainingNames(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128445)
	retryOpts := retry.Options{
		MaxBackoff: 10 * time.Second,
	}
	for r := retry.StartWithCtx(ctx, retryOpts); r.Next(); {
		__antithesis_instrumentation__.Notify(128447)
		rows, err := d.InternalExecutor.QueryBufferedEx(
			ctx,
			"ensure-no-draining-names",
			nil,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			query,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(128450)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128451)
		}
		__antithesis_instrumentation__.Notify(128448)
		if len(rows) == 0 {
			__antithesis_instrumentation__.Notify(128452)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(128453)
		}
		__antithesis_instrumentation__.Notify(128449)
		datums := rows[0]
		id := descpb.ID(*datums[0].(*tree.DInt))
		name := string(*datums[1].(*tree.DString))
		log.Infof(ctx, "descriptor with ID %d and name %q still has draining names", id, name)
	}
	__antithesis_instrumentation__.Notify(128446)
	return ctx.Err()
}
