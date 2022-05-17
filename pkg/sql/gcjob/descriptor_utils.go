package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/config"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func updateDescriptorGCMutations(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	tableID descpb.ID,
	garbageCollectedIndexID descpb.IndexID,
) error {
	__antithesis_instrumentation__.Notify(492173)
	return sql.DescsTxn(ctx, execCfg, func(
		ctx context.Context, txn *kv.Txn, descsCol *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(492174)
		tbl, err := descsCol.GetMutableTableVersionByID(ctx, tableID, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(492178)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492179)
		}
		__antithesis_instrumentation__.Notify(492175)
		found := false
		for i := 0; i < len(tbl.GCMutations); i++ {
			__antithesis_instrumentation__.Notify(492180)
			other := tbl.GCMutations[i]
			if other.IndexID == garbageCollectedIndexID {
				__antithesis_instrumentation__.Notify(492181)
				tbl.GCMutations = append(tbl.GCMutations[:i], tbl.GCMutations[i+1:]...)
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(492182)
			}
		}
		__antithesis_instrumentation__.Notify(492176)
		if found {
			__antithesis_instrumentation__.Notify(492183)
			log.Infof(ctx, "updating GCMutations for table %d after removing index %d",
				tableID, garbageCollectedIndexID)

			b := txn.NewBatch()
			if err := descsCol.WriteDescToBatch(ctx, false, tbl, b); err != nil {
				__antithesis_instrumentation__.Notify(492185)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492186)
			}
			__antithesis_instrumentation__.Notify(492184)
			return txn.Run(ctx, b)
		} else {
			__antithesis_instrumentation__.Notify(492187)
		}
		__antithesis_instrumentation__.Notify(492177)
		return nil
	})
}

func deleteDatabaseZoneConfig(
	ctx context.Context,
	db *kv.DB,
	codec keys.SQLCodec,
	settings *cluster.Settings,
	databaseID descpb.ID,
) error {
	__antithesis_instrumentation__.Notify(492188)
	if databaseID == descpb.InvalidID {
		__antithesis_instrumentation__.Notify(492190)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(492191)
	}
	__antithesis_instrumentation__.Notify(492189)
	return db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(492192)
		if !settings.Version.IsActive(
			ctx, clusterversion.DisableSystemConfigGossipTrigger,
		) {
			__antithesis_instrumentation__.Notify(492194)
			if err := txn.DeprecatedSetSystemConfigTrigger(codec.ForSystemTenant()); err != nil {
				__antithesis_instrumentation__.Notify(492195)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492196)
			}
		} else {
			__antithesis_instrumentation__.Notify(492197)
		}
		__antithesis_instrumentation__.Notify(492193)
		b := &kv.Batch{}

		dbZoneKeyPrefix := config.MakeZoneKeyPrefix(codec, databaseID)
		b.DelRange(dbZoneKeyPrefix, dbZoneKeyPrefix.PrefixEnd(), false)
		return txn.Run(ctx, b)
	})
}
