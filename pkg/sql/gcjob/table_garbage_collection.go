package gcjob

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/spanconfig"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func gcTables(
	ctx context.Context, execCfg *sql.ExecutorConfig, progress *jobspb.SchemaChangeGCProgress,
) error {
	__antithesis_instrumentation__.Notify(492674)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(492677)
		log.Infof(ctx, "GC is being considered for tables: %+v", progress.Tables)
	} else {
		__antithesis_instrumentation__.Notify(492678)
	}
	__antithesis_instrumentation__.Notify(492675)
	for _, droppedTable := range progress.Tables {
		__antithesis_instrumentation__.Notify(492679)
		if droppedTable.Status != jobspb.SchemaChangeGCProgress_DELETING {
			__antithesis_instrumentation__.Notify(492687)

			continue
		} else {
			__antithesis_instrumentation__.Notify(492688)
		}
		__antithesis_instrumentation__.Notify(492680)

		var table catalog.TableDescriptor
		if err := sql.DescsTxn(ctx, execCfg, func(ctx context.Context, txn *kv.Txn, col *descs.Collection) (err error) {
			__antithesis_instrumentation__.Notify(492689)
			table, err = col.Direct().MustGetTableDescByID(ctx, txn, droppedTable.ID)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(492690)
			if errors.Is(err, catalog.ErrDescriptorNotFound) {
				__antithesis_instrumentation__.Notify(492692)

				log.Warningf(ctx, "table descriptor %d not found while attempting to GC, skipping", droppedTable.ID)

				markTableGCed(ctx, droppedTable.ID, progress)
				continue
			} else {
				__antithesis_instrumentation__.Notify(492693)
			}
			__antithesis_instrumentation__.Notify(492691)
			return errors.Wrapf(err, "fetching table %d", droppedTable.ID)
		} else {
			__antithesis_instrumentation__.Notify(492694)
		}
		__antithesis_instrumentation__.Notify(492681)

		if !table.Dropped() {
			__antithesis_instrumentation__.Notify(492695)

			continue
		} else {
			__antithesis_instrumentation__.Notify(492696)
		}
		__antithesis_instrumentation__.Notify(492682)

		if err := ClearTableData(
			ctx, execCfg.DB, execCfg.DistSender, execCfg.Codec, &execCfg.Settings.SV, table,
		); err != nil {
			__antithesis_instrumentation__.Notify(492697)
			return errors.Wrapf(err, "clearing data for table %d", table.GetID())
		} else {
			__antithesis_instrumentation__.Notify(492698)
		}
		__antithesis_instrumentation__.Notify(492683)

		delta, err := spanconfig.Delta(ctx, execCfg.SpanConfigSplitter, table, nil)
		if err != nil {
			__antithesis_instrumentation__.Notify(492699)
			return err
		} else {
			__antithesis_instrumentation__.Notify(492700)
		}
		__antithesis_instrumentation__.Notify(492684)

		if err := execCfg.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(492701)
			_, err := execCfg.SpanConfigLimiter.ShouldLimit(ctx, txn, delta)
			return err
		}); err != nil {
			__antithesis_instrumentation__.Notify(492702)
			return errors.Wrapf(err, "deducting span count for table %d", table.GetID())
		} else {
			__antithesis_instrumentation__.Notify(492703)
		}
		__antithesis_instrumentation__.Notify(492685)

		if err := sql.DeleteTableDescAndZoneConfig(
			ctx, execCfg.DB, execCfg.Settings, execCfg.Codec, table,
		); err != nil {
			__antithesis_instrumentation__.Notify(492704)
			return errors.Wrapf(err, "dropping table descriptor for table %d", table.GetID())
		} else {
			__antithesis_instrumentation__.Notify(492705)
		}
		__antithesis_instrumentation__.Notify(492686)

		markTableGCed(ctx, table.GetID(), progress)
	}
	__antithesis_instrumentation__.Notify(492676)
	return nil
}

func ClearTableData(
	ctx context.Context,
	db *kv.DB,
	distSender *kvcoord.DistSender,
	codec keys.SQLCodec,
	sv *settings.Values,
	table catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(492706)
	log.Infof(ctx, "clearing data for table %d", table.GetID())
	tableKey := roachpb.RKey(codec.TablePrefix(uint32(table.GetID())))
	tableSpan := roachpb.RSpan{Key: tableKey, EndKey: tableKey.PrefixEnd()}
	return clearSpanData(ctx, db, distSender, tableSpan)
}

func clearSpanData(
	ctx context.Context, db *kv.DB, distSender *kvcoord.DistSender, span roachpb.RSpan,
) error {
	__antithesis_instrumentation__.Notify(492707)

	const batchSize = 100
	const waitTime = 500 * time.Millisecond

	var n int
	lastKey := span.Key
	ri := kvcoord.MakeRangeIterator(distSender)
	timer := timeutil.NewTimer()
	defer timer.Stop()

	for ri.Seek(ctx, span.Key, kvcoord.Ascending); ; ri.Next(ctx) {
		__antithesis_instrumentation__.Notify(492709)
		if !ri.Valid() {
			__antithesis_instrumentation__.Notify(492712)
			return ri.Error()
		} else {
			__antithesis_instrumentation__.Notify(492713)
		}
		__antithesis_instrumentation__.Notify(492710)

		if n++; n >= batchSize || func() bool {
			__antithesis_instrumentation__.Notify(492714)
			return !ri.NeedAnother(span) == true
		}() == true {
			__antithesis_instrumentation__.Notify(492715)
			endKey := ri.Desc().EndKey
			if span.EndKey.Less(endKey) {
				__antithesis_instrumentation__.Notify(492718)
				endKey = span.EndKey
			} else {
				__antithesis_instrumentation__.Notify(492719)
			}
			__antithesis_instrumentation__.Notify(492716)
			var b kv.Batch
			b.AddRawRequest(&roachpb.ClearRangeRequest{
				RequestHeader: roachpb.RequestHeader{
					Key:    lastKey.AsRawKey(),
					EndKey: endKey.AsRawKey(),
				},
			})
			log.VEventf(ctx, 2, "ClearRange %s - %s", lastKey, endKey)
			if err := db.Run(ctx, &b); err != nil {
				__antithesis_instrumentation__.Notify(492720)
				return errors.Wrapf(err, "clear range %s - %s", lastKey, endKey)
			} else {
				__antithesis_instrumentation__.Notify(492721)
			}
			__antithesis_instrumentation__.Notify(492717)
			n = 0
			lastKey = endKey
			timer.Reset(waitTime)
			select {
			case <-timer.C:
				__antithesis_instrumentation__.Notify(492722)
				timer.Read = true
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(492723)
				return ctx.Err()
			}
		} else {
			__antithesis_instrumentation__.Notify(492724)
		}
		__antithesis_instrumentation__.Notify(492711)

		if !ri.NeedAnother(span) {
			__antithesis_instrumentation__.Notify(492725)
			break
		} else {
			__antithesis_instrumentation__.Notify(492726)
		}
	}
	__antithesis_instrumentation__.Notify(492708)

	return nil
}
