package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const RevertTableDefaultBatchSize = 500000

var useTBIForRevertRange = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.bulk_io_write.revert_range_time_bound_iterator.enabled",
	"use the time-bound iterator optimization when processing a revert range request",
	true,
)

func RevertTables(
	ctx context.Context,
	db *kv.DB,
	execCfg *ExecutorConfig,
	tables []catalog.TableDescriptor,
	targetTime hlc.Timestamp,
	ignoreGCThreshold bool,
	batchSize int64,
) error {
	__antithesis_instrumentation__.Notify(567167)
	reverting := make(map[descpb.ID]bool, len(tables))
	for i := range tables {
		__antithesis_instrumentation__.Notify(567172)
		reverting[tables[i].GetID()] = true
	}
	__antithesis_instrumentation__.Notify(567168)

	spans := make([]roachpb.Span, 0, len(tables))

	for i := range tables {
		__antithesis_instrumentation__.Notify(567173)
		if tables[i].GetState() != descpb.DescriptorState_OFFLINE {
			__antithesis_instrumentation__.Notify(567176)
			return errors.New("only offline tables can be reverted")
		} else {
			__antithesis_instrumentation__.Notify(567177)
		}
		__antithesis_instrumentation__.Notify(567174)

		if !tables[i].IsPhysicalTable() {
			__antithesis_instrumentation__.Notify(567178)
			return errors.Errorf("cannot revert virtual table %s", tables[i].GetName())
		} else {
			__antithesis_instrumentation__.Notify(567179)
		}
		__antithesis_instrumentation__.Notify(567175)
		spans = append(spans, tables[i].TableSpan(execCfg.Codec))
	}
	__antithesis_instrumentation__.Notify(567169)

	for i := range tables {
		__antithesis_instrumentation__.Notify(567180)

		log.Infof(ctx, "reverting table %s (%d) to time %v", tables[i].GetName(), tables[i].GetID(), targetTime)
	}
	__antithesis_instrumentation__.Notify(567170)

	for len(spans) != 0 {
		__antithesis_instrumentation__.Notify(567181)
		var b kv.Batch
		for _, span := range spans {
			__antithesis_instrumentation__.Notify(567184)
			b.AddRawRequest(&roachpb.RevertRangeRequest{
				RequestHeader: roachpb.RequestHeader{
					Key:    span.Key,
					EndKey: span.EndKey,
				},
				TargetTime:                          targetTime,
				IgnoreGcThreshold:                   ignoreGCThreshold,
				EnableTimeBoundIteratorOptimization: useTBIForRevertRange.Get(&execCfg.Settings.SV),
			})
		}
		__antithesis_instrumentation__.Notify(567182)
		b.Header.MaxSpanRequestKeys = batchSize

		if err := db.Run(ctx, &b); err != nil {
			__antithesis_instrumentation__.Notify(567185)
			return err
		} else {
			__antithesis_instrumentation__.Notify(567186)
		}
		__antithesis_instrumentation__.Notify(567183)

		spans = spans[:0]
		for _, raw := range b.RawResponse().Responses {
			__antithesis_instrumentation__.Notify(567187)
			r := raw.GetRevertRange()
			if r.ResumeSpan != nil {
				__antithesis_instrumentation__.Notify(567188)
				if !r.ResumeSpan.Valid() {
					__antithesis_instrumentation__.Notify(567190)
					return errors.Errorf("invalid resume span: %s", r.ResumeSpan)
				} else {
					__antithesis_instrumentation__.Notify(567191)
				}
				__antithesis_instrumentation__.Notify(567189)
				spans = append(spans, *r.ResumeSpan)
			} else {
				__antithesis_instrumentation__.Notify(567192)
			}
		}
	}
	__antithesis_instrumentation__.Notify(567171)

	return nil
}
