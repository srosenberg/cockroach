package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/errors"
)

func spanCountTableMigration(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128697)
	if d.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(128699)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128700)
	}
	__antithesis_instrumentation__.Notify(128698)

	return createSystemTable(
		ctx, d.DB, d.Codec, systemschema.SpanCountTable,
	)
}

func seedSpanCountTableMigration(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128701)
	if d.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(128703)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128704)
	}
	__antithesis_instrumentation__.Notify(128702)

	return d.CollectionFactory.Txn(ctx, d.InternalExecutor, d.DB, func(ctx context.Context, txn *kv.Txn, descriptors *descs.Collection) error {
		__antithesis_instrumentation__.Notify(128705)
		dbs, err := descriptors.GetAllDatabaseDescriptors(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(128711)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128712)
		}
		__antithesis_instrumentation__.Notify(128706)

		var spanCount int
		for _, db := range dbs {
			__antithesis_instrumentation__.Notify(128713)
			if db.GetID() == systemschema.SystemDB.GetID() {
				__antithesis_instrumentation__.Notify(128716)
				continue
			} else {
				__antithesis_instrumentation__.Notify(128717)
			}
			__antithesis_instrumentation__.Notify(128714)

			tables, err := descriptors.GetAllTableDescriptorsInDatabase(ctx, txn, db.GetID())
			if err != nil {
				__antithesis_instrumentation__.Notify(128718)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128719)
			}
			__antithesis_instrumentation__.Notify(128715)

			for _, table := range tables {
				__antithesis_instrumentation__.Notify(128720)
				splits, err := d.SpanConfig.Splitter.Splits(ctx, table)
				if err != nil {
					__antithesis_instrumentation__.Notify(128722)
					return err
				} else {
					__antithesis_instrumentation__.Notify(128723)
				}
				__antithesis_instrumentation__.Notify(128721)
				spanCount += splits
			}
		}
		__antithesis_instrumentation__.Notify(128707)

		const seedSpanCountStmt = `
INSERT INTO system.span_count (span_count) VALUES ($1)
ON CONFLICT (singleton)
DO UPDATE SET span_count = $1
RETURNING span_count
`
		datums, err := d.InternalExecutor.QueryRowEx(ctx, "seed-span-count", txn,
			sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			seedSpanCountStmt, spanCount)
		if err != nil {
			__antithesis_instrumentation__.Notify(128724)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128725)
		}
		__antithesis_instrumentation__.Notify(128708)
		if len(datums) != 1 {
			__antithesis_instrumentation__.Notify(128726)
			return errors.AssertionFailedf("expected to return 1 row, return %d", len(datums))
		} else {
			__antithesis_instrumentation__.Notify(128727)
		}
		__antithesis_instrumentation__.Notify(128709)
		if insertedSpanCount := int64(tree.MustBeDInt(datums[0])); insertedSpanCount != int64(spanCount) {
			__antithesis_instrumentation__.Notify(128728)
			return errors.AssertionFailedf("expected to insert %d, got %d", spanCount, insertedSpanCount)
		} else {
			__antithesis_instrumentation__.Notify(128729)
		}
		__antithesis_instrumentation__.Notify(128710)
		return nil
	})
}
