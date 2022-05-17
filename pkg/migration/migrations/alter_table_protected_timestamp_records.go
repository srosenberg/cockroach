package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
)

const addTargetCol = `
ALTER TABLE system.protected_ts_records
ADD COLUMN IF NOT EXISTS target BYTES FAMILY "primary"
`

func alterTableProtectedTimestampRecords(
	ctx context.Context, cs clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128380)
	op := operation{
		name:           "add-table-pts-records-target-col",
		schemaList:     []string{"target"},
		query:          addTargetCol,
		schemaExistsFn: hasColumn,
	}
	if err := migrateTable(ctx, cs, d, op,
		keys.ProtectedTimestampsRecordsTableID,
		systemschema.ProtectedTimestampsRecordsTable); err != nil {
		__antithesis_instrumentation__.Notify(128382)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128383)
	}
	__antithesis_instrumentation__.Notify(128381)
	return nil
}
