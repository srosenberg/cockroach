package migrations

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/migration"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func seedTenantSpanConfigsMigration(
	ctx context.Context, _ clusterversion.ClusterVersion, d migration.TenantDeps, _ *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(128671)
	if !d.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(128673)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(128674)
	}
	__antithesis_instrumentation__.Notify(128672)

	return d.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(128675)
		const getTenantIDsQuery = `SELECT id from system.tenants`
		it, err := d.InternalExecutor.QueryIteratorEx(ctx, "get-tenant-ids", txn,
			sessiondata.NodeUserSessionDataOverride, getTenantIDsQuery,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(128680)
			return errors.Wrap(err, "unable to fetch existing tenant IDs")
		} else {
			__antithesis_instrumentation__.Notify(128681)
		}
		__antithesis_instrumentation__.Notify(128676)

		var tenantIDs []roachpb.TenantID
		var ok bool
		for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
			__antithesis_instrumentation__.Notify(128682)
			row := it.Cur()
			tenantID := roachpb.MakeTenantID(uint64(tree.MustBeDInt(row[0])))
			tenantIDs = append(tenantIDs, tenantID)
		}
		__antithesis_instrumentation__.Notify(128677)
		if err != nil {
			__antithesis_instrumentation__.Notify(128683)
			return err
		} else {
			__antithesis_instrumentation__.Notify(128684)
		}
		__antithesis_instrumentation__.Notify(128678)

		scKVAccessor := d.SpanConfig.KVAccessor.WithTxn(ctx, txn)
		for _, tenantID := range tenantIDs {
			__antithesis_instrumentation__.Notify(128685)

			tenantSpanConfig := d.SpanConfig.Default
			tenantPrefix := keys.MakeTenantPrefix(tenantID)
			tenantTarget := spanconfig.MakeTargetFromSpan(roachpb.Span{
				Key:    tenantPrefix,
				EndKey: tenantPrefix.PrefixEnd(),
			})
			tenantSeedSpan := roachpb.Span{
				Key:    tenantPrefix,
				EndKey: tenantPrefix.Next(),
			}
			record, err := spanconfig.MakeRecord(spanconfig.MakeTargetFromSpan(tenantSeedSpan),
				tenantSpanConfig)
			if err != nil {
				__antithesis_instrumentation__.Notify(128689)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128690)
			}
			__antithesis_instrumentation__.Notify(128686)
			toUpsert := []spanconfig.Record{record}
			scRecords, err := scKVAccessor.GetSpanConfigRecords(ctx, []spanconfig.Target{tenantTarget})
			if err != nil {
				__antithesis_instrumentation__.Notify(128691)
				return err
			} else {
				__antithesis_instrumentation__.Notify(128692)
			}
			__antithesis_instrumentation__.Notify(128687)
			if len(scRecords) != 0 {
				__antithesis_instrumentation__.Notify(128693)

				continue
			} else {
				__antithesis_instrumentation__.Notify(128694)
			}
			__antithesis_instrumentation__.Notify(128688)
			if err := scKVAccessor.UpdateSpanConfigRecords(
				ctx, nil, toUpsert, hlc.MinTimestamp, hlc.MaxTimestamp,
			); err != nil {
				__antithesis_instrumentation__.Notify(128695)
				return errors.Wrapf(err, "failed to seed span config for tenant %d", tenantID)
			} else {
				__antithesis_instrumentation__.Notify(128696)
			}
		}
		__antithesis_instrumentation__.Notify(128679)

		return nil
	})
}
