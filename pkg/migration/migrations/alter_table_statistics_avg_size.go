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

const addAvgSizeCol = `
ALTER TABLE system.table_statistics
ADD COLUMN IF NOT EXISTS "avgSize" INT8 NOT NULL DEFAULT (INT8 '0')
FAMILY "fam_0_tableID_statisticID_name_columnIDs_createdAt_rowCount_distinctCount_nullCount_histogram"
`

func alterSystemTableStatisticsAddAvgSize(
	ctx context.Context, cs clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128384)
	op := operation{
		name:           "add-table-statistics-avgSize-col",
		schemaList:     []string{"total_consumption"},
		query:          addAvgSizeCol,
		schemaExistsFn: hasColumn,
	}
	if err := migrateTable(ctx, cs, d, op, keys.TableStatisticsTableID, systemschema.TableStatisticsTable); err != nil {
		__antithesis_instrumentation__.Notify(128386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(128387)
	}
	__antithesis_instrumentation__.Notify(128385)
	return nil
}
