package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func gcIndexes(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	parentID descpb.ID,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492464)
	droppedIndexes := progress.Indexes
	if log.V(2) {
		__antithesis_instrumentation__.Notify(492469)
		log.Infof(ctx, "GC is being considered on table %d for indexes indexes: %+v", parentID, droppedIndexes)
	} else {
		__antithesis_instrumentation__.Notify(492470)
	}
	__antithesis_instrumentation__.Notify(492465)

	parentDesc, err := sql.WaitToUpdateLeases(ctx, execCfg.LeaseManager, parentID)
	if err != nil {
		__antithesis_instrumentation__.Notify(492471)
		return err
	} else {
		__antithesis_instrumentation__.Notify(492472)
	}
	__antithesis_instrumentation__.Notify(492466)

	parentTable, isTable := parentDesc.(catalog.TableDescriptor)
	if !isTable {
		__antithesis_instrumentation__.Notify(492473)
		return errors.AssertionFailedf("expected descriptor %d to be a table, not %T", parentID, parentDesc)
	} else {
		__antithesis_instrumentation__.Notify(492474)
	}
	__antithesis_instrumentation__.Notify(492467)
	for _, index := range droppedIndexes {
		__antithesis_instrumentation__.Notify(492475)
		if index.Status != jobspb.SchemaChangeGCProgress_DELETING {
			__antithesis_instrumentation__.Notify(492480)
			continue
		} else {
			__antithesis_instrumentation__.Notify(492481)
		}
		__antithesis_instrumentation__.Notify(492476)

		if err := clearIndex(ctx, execCfg, parentTable, index.IndexID); err != nil {
			__antithesis_instrumentation__.Notify(492482)
			return errors.Wrapf(err, "clearing index %d from table %d", index.IndexID, parentTable.GetID())
		} else {
			__antithesis_instrumentation__.Notify(492483)
		}
		__antithesis_instrumentation__.Notify(492477)

		removeIndexZoneConfigs := func(
			ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
		) error {
			__antithesis_instrumentation__.Notify(492484)
			freshParentTableDesc, err := descriptors.GetMutableTableByID(
				ctx, txn, parentID, tree.ObjectLookupFlags{
					CommonLookupFlags: tree.CommonLookupFlags{
						AvoidLeased:    true,
						Required:       true,
						IncludeDropped: true,
						IncludeOffline: true,
					},
				})
			if err != nil {
				__antithesis_instrumentation__.Notify(492486)
				return err
			} else {
				__antithesis_instrumentation__.Notify(492487)
			}
			__antithesis_instrumentation__.Notify(492485)
			return sql.RemoveIndexZoneConfigs(
				ctx, txn, execCfg, freshParentTableDesc, []uint32{uint32(index.IndexID)},
			)
		}
		__antithesis_instrumentation__.Notify(492478)
		if err := sql.DescsTxn(ctx, execCfg, removeIndexZoneConfigs); err != nil {
			__antithesis_instrumentation__.Notify(492488)
			return errors.Wrapf(err, "removing index %d zone configs", index.IndexID)
		} else {
			__antithesis_instrumentation__.Notify(492489)
		}
		__antithesis_instrumentation__.Notify(492479)

		if err := completeDroppedIndex(
			ctx, execCfg, parentTable, index.IndexID, progress,
		); err != nil {
			__antithesis_instrumentation__.Notify(492490)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492491)
		}
	}
	__antithesis_instrumentation__.Notify(492468)
	return nil
}

func clearIndex(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	tableDesc catalog.TableDescriptor,
	indexID descpb.IndexID,
) error {
	__antithesis_instrumentation__.Notify(492492)
	log.Infof(ctx, "clearing index %d from table %d", indexID, tableDesc.GetID())

	sp := tableDesc.IndexSpan(execCfg.Codec, indexID)
	start, err := keys.Addr(sp.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(492495)
		return errors.Wrap(err, "failed to addr index start")
	} else {
		__antithesis_instrumentation__.Notify(492496)
	}
	__antithesis_instrumentation__.Notify(492493)
	end, err := keys.Addr(sp.EndKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(492497)
		return errors.Wrap(err, "failed to addr index end")
	} else {
		__antithesis_instrumentation__.Notify(492498)
	}
	__antithesis_instrumentation__.Notify(492494)
	rSpan := roachpb.RSpan{Key: start, EndKey: end}
	return clearSpanData(ctx, execCfg.DB, execCfg.DistSender, rSpan)
}

func completeDroppedIndex(
	ctx context.Context,
	execCfg *sql.ExecutorConfig,
	table catalog.TableDescriptor,
	indexID descpb.IndexID,
	progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492499)
	if err := updateDescriptorGCMutations(ctx, execCfg, table.GetID(), indexID); err != nil {
		__antithesis_instrumentation__.Notify(492501)
		return errors.Wrapf(err, "updating GC mutations")
	} else {
		__antithesis_instrumentation__.Notify(492502)
	}
	__antithesis_instrumentation__.Notify(492500)

	markIndexGCed(ctx, indexID, progress)

	return nil
}
