package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
)

func ensureCommentsHaveNonDroppedIndexes(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128388)
	return d.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(128389)

		_, err := d.InternalExecutor.QueryBufferedEx(
			ctx,
			"select-comments-with-missing-indexes",
			txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			`DELETE FROM system.comments
      WHERE type = $1
            AND (object_id, sub_id)
				NOT IN (
						SELECT (descriptor_id, index_id)
						  FROM crdb_internal.table_indexes
					);`,
			keys.IndexCommentType,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(128391)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128392)
		}
		__antithesis_instrumentation__.Notify(128390)
		return txn.Commit(ctx)
	})
}
