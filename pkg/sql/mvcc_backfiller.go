package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type IndexBackfillerMergePlanner struct {
	execCfg   *ExecutorConfig
	ieFactory sqlutil.SessionBoundInternalExecutorFactory
}

func NewIndexBackfillerMergePlanner(
	execCfg *ExecutorConfig, ieFactory sqlutil.SessionBoundInternalExecutorFactory,
) *IndexBackfillerMergePlanner {
	__antithesis_instrumentation__.Notify(501782)
	return &IndexBackfillerMergePlanner{execCfg: execCfg, ieFactory: ieFactory}
}

func (im *IndexBackfillerMergePlanner) plan(
	ctx context.Context,
	tableDesc catalog.TableDescriptor,
	todoSpanList [][]roachpb.Span,
	addedIndexes, temporaryIndexes []descpb.IndexID,
	metaFn func(_ context.Context, meta *execinfrapb.ProducerMetadata) error,
	mergeTimestamp hlc.Timestamp,
) (func(context.Context) error, error) {
	__antithesis_instrumentation__.Notify(501783)
	var p *PhysicalPlan
	var evalCtx extendedEvalContext
	var planCtx *PlanningCtx

	if err := DescsTxn(ctx, im.execCfg, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(501785)
		evalCtx = createSchemaChangeEvalCtx(ctx, im.execCfg, txn.ReadTimestamp(), descriptors)
		planCtx = im.execCfg.DistSQLPlanner.NewPlanningCtx(ctx, &evalCtx, nil, txn,
			DistributionTypeSystemTenantOnly)

		spec, err := initIndexBackfillMergerSpec(*tableDesc.TableDesc(), addedIndexes, temporaryIndexes, mergeTimestamp)
		if err != nil {
			__antithesis_instrumentation__.Notify(501787)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501788)
		}
		__antithesis_instrumentation__.Notify(501786)
		p, err = im.execCfg.DistSQLPlanner.createIndexBackfillerMergePhysicalPlan(ctx, planCtx, spec, todoSpanList)
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(501789)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(501790)
	}
	__antithesis_instrumentation__.Notify(501784)

	return func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(501791)
		cbw := MetadataCallbackWriter{rowResultWriter: &errOnlyResultWriter{}, fn: metaFn}
		recv := MakeDistSQLReceiver(
			ctx,
			&cbw,
			tree.Rows,
			im.execCfg.RangeDescriptorCache,
			nil,
			im.execCfg.Clock,
			evalCtx.Tracing,
			im.execCfg.ContentionRegistry,
			nil,
		)
		defer recv.Release()
		evalCtxCopy := evalCtx
		im.execCfg.DistSQLPlanner.Run(
			ctx,
			planCtx,
			nil,
			p, recv, &evalCtxCopy,
			nil,
		)()
		return cbw.Err()
	}, nil
}

type MergeProgress struct {
	TodoSpans [][]roachpb.Span

	MutationIdx []int

	AddedIndexes, TemporaryIndexes []descpb.IndexID
}

func (mp *MergeProgress) Copy() *MergeProgress {
	__antithesis_instrumentation__.Notify(501792)
	newp := &MergeProgress{
		TodoSpans:        make([][]roachpb.Span, len(mp.TodoSpans)),
		MutationIdx:      make([]int, len(mp.MutationIdx)),
		AddedIndexes:     make([]descpb.IndexID, len(mp.AddedIndexes)),
		TemporaryIndexes: make([]descpb.IndexID, len(mp.TemporaryIndexes)),
	}
	copy(newp.MutationIdx, mp.MutationIdx)
	copy(newp.AddedIndexes, mp.AddedIndexes)
	copy(newp.TemporaryIndexes, mp.TemporaryIndexes)
	for i, spanSlice := range mp.TodoSpans {
		__antithesis_instrumentation__.Notify(501794)
		newSpanSlice := make([]roachpb.Span, len(spanSlice))
		copy(newSpanSlice, spanSlice)
		newp.TodoSpans[i] = newSpanSlice
	}
	__antithesis_instrumentation__.Notify(501793)
	return newp
}

func (mp *MergeProgress) FlatSpans() []roachpb.Span {
	__antithesis_instrumentation__.Notify(501795)
	spans := []roachpb.Span{}
	for _, s := range mp.TodoSpans {
		__antithesis_instrumentation__.Notify(501797)
		spans = append(spans, s...)
	}
	__antithesis_instrumentation__.Notify(501796)
	return spans
}

type IndexMergeTracker struct {
	mu struct {
		syncutil.Mutex
		progress *MergeProgress

		hasOrigNRanges bool
		origNRanges    int
	}

	jobMu struct {
		syncutil.Mutex
		job *jobs.Job
	}

	rangeCounter   rangeCounter
	fractionScaler *multiStageFractionScaler
}

var _ scexec.BackfillProgressFlusher = (*IndexMergeTracker)(nil)

type rangeCounter func(ctx context.Context, spans []roachpb.Span) (int, error)

func NewIndexMergeTracker(
	progress *MergeProgress,
	job *jobs.Job,
	rangeCounter rangeCounter,
	scaler *multiStageFractionScaler,
) *IndexMergeTracker {
	__antithesis_instrumentation__.Notify(501798)
	imt := IndexMergeTracker{
		rangeCounter:   rangeCounter,
		fractionScaler: scaler,
	}
	imt.mu.hasOrigNRanges = false
	imt.mu.progress = progress.Copy()
	imt.jobMu.job = job
	return &imt
}

func (imt *IndexMergeTracker) FlushCheckpoint(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(501799)
	imt.jobMu.Lock()
	defer imt.jobMu.Unlock()

	imt.mu.Lock()
	if imt.mu.progress.TodoSpans == nil {
		__antithesis_instrumentation__.Notify(501803)
		imt.mu.Unlock()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(501804)
	}
	__antithesis_instrumentation__.Notify(501800)
	progress := imt.mu.progress.Copy()
	imt.mu.Unlock()

	details, ok := imt.jobMu.job.Details().(jobspb.SchemaChangeDetails)
	if !ok {
		__antithesis_instrumentation__.Notify(501805)
		return errors.Errorf("expected SchemaChangeDetails job type, got %T", imt.jobMu.job.Details())
	} else {
		__antithesis_instrumentation__.Notify(501806)
	}
	__antithesis_instrumentation__.Notify(501801)

	for idx := range progress.TodoSpans {
		__antithesis_instrumentation__.Notify(501807)
		details.ResumeSpanList[progress.MutationIdx[idx]].ResumeSpans = progress.TodoSpans[idx]
	}
	__antithesis_instrumentation__.Notify(501802)

	return imt.jobMu.job.SetDetails(ctx, nil, details)
}

func (imt *IndexMergeTracker) FlushFractionCompleted(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(501808)
	imt.mu.Lock()
	spans := imt.mu.progress.FlatSpans()
	imt.mu.Unlock()

	rangeCount, err := imt.rangeCounter(ctx, spans)
	if err != nil {
		__antithesis_instrumentation__.Notify(501811)
		return err
	} else {
		__antithesis_instrumentation__.Notify(501812)
	}
	__antithesis_instrumentation__.Notify(501809)

	orig := imt.maybeSetOrigNRanges(rangeCount)
	if orig >= rangeCount && func() bool {
		__antithesis_instrumentation__.Notify(501813)
		return orig != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(501814)
		fractionRangesFinished := float32(orig-rangeCount) / float32(orig)
		frac, err := imt.fractionScaler.fractionCompleteFromStageFraction(stageMerge, fractionRangesFinished)
		if err != nil {
			__antithesis_instrumentation__.Notify(501816)
			return err
		} else {
			__antithesis_instrumentation__.Notify(501817)
		}
		__antithesis_instrumentation__.Notify(501815)

		imt.jobMu.Lock()
		defer imt.jobMu.Unlock()
		if err := imt.jobMu.job.FractionProgressed(ctx, nil,
			jobs.FractionUpdater(frac)); err != nil {
			__antithesis_instrumentation__.Notify(501818)
			return jobs.SimplifyInvalidStatusError(err)
		} else {
			__antithesis_instrumentation__.Notify(501819)
		}
	} else {
		__antithesis_instrumentation__.Notify(501820)
	}
	__antithesis_instrumentation__.Notify(501810)
	return nil
}

func (imt *IndexMergeTracker) maybeSetOrigNRanges(count int) int {
	__antithesis_instrumentation__.Notify(501821)
	imt.mu.Lock()
	defer imt.mu.Unlock()
	if !imt.mu.hasOrigNRanges {
		__antithesis_instrumentation__.Notify(501823)
		imt.mu.hasOrigNRanges = true
		imt.mu.origNRanges = count
	} else {
		__antithesis_instrumentation__.Notify(501824)
	}
	__antithesis_instrumentation__.Notify(501822)
	return imt.mu.origNRanges
}

func (imt *IndexMergeTracker) UpdateMergeProgress(
	ctx context.Context, updateFn func(ctx context.Context, progress *MergeProgress),
) {
	__antithesis_instrumentation__.Notify(501825)
	imt.mu.Lock()
	defer imt.mu.Unlock()
	updateFn(ctx, imt.mu.progress)
}

func newPeriodicProgressFlusher(settings *cluster.Settings) scexec.PeriodicProgressFlusher {
	__antithesis_instrumentation__.Notify(501826)
	return scdeps.NewPeriodicProgressFlusher(
		func() time.Duration {
			__antithesis_instrumentation__.Notify(501827)
			return backfill.IndexBackfillCheckpointInterval.Get(&settings.SV)
		},
		func() time.Duration {
			__antithesis_instrumentation__.Notify(501828)

			const fractionInterval = 10 * time.Second
			return fractionInterval
		},
	)
}
