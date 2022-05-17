package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type IndexBackfillPlanner struct {
	execCfg   *ExecutorConfig
	ieFactory sqlutil.SessionBoundInternalExecutorFactory
}

func NewIndexBackfiller(
	execCfg *ExecutorConfig, ieFactory sqlutil.SessionBoundInternalExecutorFactory,
) *IndexBackfillPlanner {
	__antithesis_instrumentation__.Notify(496658)
	return &IndexBackfillPlanner{execCfg: execCfg, ieFactory: ieFactory}
}

func (ib *IndexBackfillPlanner) MaybePrepareDestIndexesForBackfill(
	ctx context.Context, current scexec.BackfillProgress, td catalog.TableDescriptor,
) (scexec.BackfillProgress, error) {
	__antithesis_instrumentation__.Notify(496659)
	if !current.MinimumWriteTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(496663)
		return current, nil
	} else {
		__antithesis_instrumentation__.Notify(496664)
	}
	__antithesis_instrumentation__.Notify(496660)

	backfillReadTimestamp := ib.execCfg.Clock.Now()
	targetSpans := make([]roachpb.Span, len(current.DestIndexIDs))
	for i, idxID := range current.DestIndexIDs {
		__antithesis_instrumentation__.Notify(496665)
		targetSpans[i] = td.IndexSpan(ib.execCfg.Codec, idxID)
	}
	__antithesis_instrumentation__.Notify(496661)
	if err := scanTargetSpansToPushTimestampCache(
		ctx, ib.execCfg.DB, backfillReadTimestamp, targetSpans,
	); err != nil {
		__antithesis_instrumentation__.Notify(496666)
		return scexec.BackfillProgress{}, err
	} else {
		__antithesis_instrumentation__.Notify(496667)
	}
	__antithesis_instrumentation__.Notify(496662)
	return scexec.BackfillProgress{
		Backfill:              current.Backfill,
		MinimumWriteTimestamp: backfillReadTimestamp,
	}, nil
}

func (ib *IndexBackfillPlanner) BackfillIndex(
	ctx context.Context,
	progress scexec.BackfillProgress,
	tracker scexec.BackfillProgressWriter,
	descriptor catalog.TableDescriptor,
) error {
	__antithesis_instrumentation__.Notify(496668)
	var completed = struct {
		syncutil.Mutex
		g roachpb.SpanGroup
	}{}
	addCompleted := func(c ...roachpb.Span) []roachpb.Span {
		__antithesis_instrumentation__.Notify(496674)
		completed.Lock()
		defer completed.Unlock()
		completed.g.Add(c...)
		return completed.g.Slice()
	}
	__antithesis_instrumentation__.Notify(496669)
	updateFunc := func(
		ctx context.Context, meta *execinfrapb.ProducerMetadata,
	) error {
		__antithesis_instrumentation__.Notify(496675)
		if meta.BulkProcessorProgress == nil {
			__antithesis_instrumentation__.Notify(496677)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(496678)
		}
		__antithesis_instrumentation__.Notify(496676)
		progress.CompletedSpans = addCompleted(
			meta.BulkProcessorProgress.CompletedSpans...)
		return tracker.SetBackfillProgress(ctx, progress)
	}
	__antithesis_instrumentation__.Notify(496670)
	var spansToDo []roachpb.Span
	{
		__antithesis_instrumentation__.Notify(496679)
		sourceIndexSpan := descriptor.IndexSpan(ib.execCfg.Codec, progress.SourceIndexID)
		var g roachpb.SpanGroup
		g.Add(sourceIndexSpan)
		g.Sub(progress.CompletedSpans...)
		spansToDo = g.Slice()
	}
	__antithesis_instrumentation__.Notify(496671)
	if len(spansToDo) == 0 {
		__antithesis_instrumentation__.Notify(496680)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(496681)
	}
	__antithesis_instrumentation__.Notify(496672)
	now := ib.execCfg.DB.Clock().Now()
	run, err := ib.plan(
		ctx,
		descriptor,
		progress.MinimumWriteTimestamp,
		now,
		progress.MinimumWriteTimestamp,
		spansToDo,
		progress.DestIndexIDs,
		updateFunc,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(496682)
		return err
	} else {
		__antithesis_instrumentation__.Notify(496683)
	}
	__antithesis_instrumentation__.Notify(496673)
	return run(ctx)
}

func scanTargetSpansToPushTimestampCache(
	ctx context.Context, db *kv.DB, backfillTimestamp hlc.Timestamp, targetSpans []roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(496684)
	const pageSize = 10000
	return db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(496685)
		if err := txn.SetFixedTimestamp(ctx, backfillTimestamp); err != nil {
			__antithesis_instrumentation__.Notify(496688)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496689)
		}
		__antithesis_instrumentation__.Notify(496686)
		for _, span := range targetSpans {
			__antithesis_instrumentation__.Notify(496690)

			if err := txn.Iterate(ctx, span.Key, span.EndKey, pageSize, iterateNoop); err != nil {
				__antithesis_instrumentation__.Notify(496691)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496692)
			}
		}
		__antithesis_instrumentation__.Notify(496687)
		return nil
	})
}

func iterateNoop(_ []kv.KeyValue) error { __antithesis_instrumentation__.Notify(496693); return nil }

var _ scexec.Backfiller = (*IndexBackfillPlanner)(nil)

func (ib *IndexBackfillPlanner) plan(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	nowTimestamp, writeAsOf, readAsOf hlc.Timestamp,
	sourceSpans []roachpb.Span,
	indexesToBackfill []descpb.IndexID,
	callback func(_ context.Context, meta *execinfrapb.ProducerMetadata) error,
) (runFunc func(context.Context) error, _ error) {
	__antithesis_instrumentation__.Notify(496694)

	var p *PhysicalPlan
	var evalCtx extendedEvalContext
	var planCtx *PlanningCtx
	td := tabledesc.NewBuilder(tableDesc.TableDesc()).BuildExistingMutableTable()
	if err := DescsTxn(ctx, ib.execCfg, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(496696)
		evalCtx = createSchemaChangeEvalCtx(ctx, ib.execCfg, nowTimestamp, descriptors)
		planCtx = ib.execCfg.DistSQLPlanner.NewPlanningCtx(ctx, &evalCtx,
			nil, txn, DistributionTypeSystemTenantOnly)

		chunkSize := indexBackfillBatchSize.Get(&ib.execCfg.Settings.SV)
		spec, err := initIndexBackfillerSpec(*td.TableDesc(), writeAsOf, readAsOf, false, chunkSize, indexesToBackfill)
		if err != nil {
			__antithesis_instrumentation__.Notify(496698)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496699)
		}
		__antithesis_instrumentation__.Notify(496697)
		p, err = ib.execCfg.DistSQLPlanner.createBackfillerPhysicalPlan(ctx, planCtx, spec, sourceSpans)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(496700)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(496701)
	}
	__antithesis_instrumentation__.Notify(496695)

	return func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(496702)
		cbw := MetadataCallbackWriter{rowResultWriter: &errOnlyResultWriter{}, fn: callback}
		recv := MakeDistSQLReceiver(
			ctx,
			&cbw,
			tree.Rows,
			ib.execCfg.RangeDescriptorCache,
			nil,
			ib.execCfg.Clock,
			evalCtx.Tracing,
			ib.execCfg.ContentionRegistry,
			nil,
		)
		defer recv.Release()
		evalCtxCopy := evalCtx
		ib.execCfg.DistSQLPlanner.Run(ctx, planCtx, nil, p, recv, &evalCtxCopy, nil)()
		return cbw.Err()
	}, nil
}
