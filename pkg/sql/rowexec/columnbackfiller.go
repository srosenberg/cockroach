package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/util/admission"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type columnBackfiller struct {
	backfiller

	backfill.ColumnBackfiller

	desc catalog.TableDescriptor

	commitWaitFns []func(context.Context) error
}

var _ execinfra.Processor = &columnBackfiller{}
var _ chunkBackfiller = &columnBackfiller{}

var backfillerMaxCommitWaitFns = settings.RegisterIntSetting(
	settings.TenantWritable,
	"schemachanger.backfiller.max_commit_wait_fns",
	"the maximum number of commit-wait functions that the columnBackfiller will accumulate before consuming them to reclaim memory",
	128,
	settings.PositiveInt,
)

func newColumnBackfiller(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.BackfillerSpec,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*columnBackfiller, error) {
	__antithesis_instrumentation__.Notify(572055)
	columnBackfillerMon := execinfra.NewMonitor(ctx, flowCtx.Cfg.BackfillerMonitor,
		"column-backfill-mon")
	cb := &columnBackfiller{
		desc: flowCtx.TableDescriptor(&spec.Table),
		backfiller: backfiller{
			name:        "Column",
			filter:      backfill.ColumnMutationFilter,
			flowCtx:     flowCtx,
			processorID: processorID,
			output:      output,
			spec:        spec,
		},
	}
	cb.backfiller.chunks = cb

	if err := cb.ColumnBackfiller.InitForDistributedUse(ctx, flowCtx, cb.desc,
		columnBackfillerMon); err != nil {
		__antithesis_instrumentation__.Notify(572057)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572058)
	}
	__antithesis_instrumentation__.Notify(572056)

	return cb, nil
}

func (cb *columnBackfiller) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572059)
	cb.ColumnBackfiller.Close(ctx)
}

func (cb *columnBackfiller) prepare(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(572060)
	return nil
}
func (cb *columnBackfiller) flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(572061)
	return cb.runCommitWait(ctx)
}
func (cb *columnBackfiller) CurrentBufferFill() float32 {
	__antithesis_instrumentation__.Notify(572062)
	return 0
}

func (cb *columnBackfiller) runChunk(
	ctx context.Context, sp roachpb.Span, chunkSize rowinfra.RowLimit, _ hlc.Timestamp,
) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(572063)
	var key roachpb.Key
	var commitWaitFn func(context.Context) error
	err := cb.flowCtx.Cfg.DB.TxnWithAdmissionControl(
		ctx, roachpb.AdmissionHeader_FROM_SQL, admission.BulkNormalPri,
		func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(572066)
			if cb.flowCtx.Cfg.TestingKnobs.RunBeforeBackfillChunk != nil {
				__antithesis_instrumentation__.Notify(572069)
				if err := cb.flowCtx.Cfg.TestingKnobs.RunBeforeBackfillChunk(sp); err != nil {
					__antithesis_instrumentation__.Notify(572070)
					return err
				} else {
					__antithesis_instrumentation__.Notify(572071)
				}
			} else {
				__antithesis_instrumentation__.Notify(572072)
			}
			__antithesis_instrumentation__.Notify(572067)
			if cb.flowCtx.Cfg.TestingKnobs.RunAfterBackfillChunk != nil {
				__antithesis_instrumentation__.Notify(572073)
				defer cb.flowCtx.Cfg.TestingKnobs.RunAfterBackfillChunk()
			} else {
				__antithesis_instrumentation__.Notify(572074)
			}
			__antithesis_instrumentation__.Notify(572068)

			commitWaitFn = txn.DeferCommitWait(ctx)

			var err error
			key, err = cb.RunColumnBackfillChunk(
				ctx,
				txn,
				cb.desc,
				sp,
				chunkSize,
				true,
				false,
			)
			return err
		})
	__antithesis_instrumentation__.Notify(572064)
	if err == nil {
		__antithesis_instrumentation__.Notify(572075)
		cb.commitWaitFns = append(cb.commitWaitFns, commitWaitFn)
		maxCommitWaitFns := int(backfillerMaxCommitWaitFns.Get(&cb.flowCtx.Cfg.Settings.SV))
		if len(cb.commitWaitFns) >= maxCommitWaitFns {
			__antithesis_instrumentation__.Notify(572076)
			if err := cb.runCommitWait(ctx); err != nil {
				__antithesis_instrumentation__.Notify(572077)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(572078)
			}
		} else {
			__antithesis_instrumentation__.Notify(572079)
		}
	} else {
		__antithesis_instrumentation__.Notify(572080)
	}
	__antithesis_instrumentation__.Notify(572065)
	return key, err
}

func (cb *columnBackfiller) runCommitWait(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(572081)
	for i, fn := range cb.commitWaitFns {
		__antithesis_instrumentation__.Notify(572083)
		if err := fn(ctx); err != nil {
			__antithesis_instrumentation__.Notify(572085)
			return err
		} else {
			__antithesis_instrumentation__.Notify(572086)
		}
		__antithesis_instrumentation__.Notify(572084)
		cb.commitWaitFns[i] = nil
	}
	__antithesis_instrumentation__.Notify(572082)
	cb.commitWaitFns = cb.commitWaitFns[:0]
	return nil
}
