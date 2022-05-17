package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/errors"
)

func initColumnBackfillerSpec(
	desc descpb.TableDescriptor, duration time.Duration, chunkSize int64, readAsOf hlc.Timestamp,
) (execinfrapb.BackfillerSpec, error) {
	__antithesis_instrumentation__.Notify(467601)
	return execinfrapb.BackfillerSpec{
		Table:     desc,
		Duration:  duration,
		ChunkSize: chunkSize,
		ReadAsOf:  readAsOf,
		Type:      execinfrapb.BackfillerSpec_Column,
	}, nil
}

func initIndexBackfillerSpec(
	desc descpb.TableDescriptor,
	writeAsOf, readAsOf hlc.Timestamp,
	writeAtBatchTimestamp bool,
	chunkSize int64,
	indexesToBackfill []descpb.IndexID,
) (execinfrapb.BackfillerSpec, error) {
	__antithesis_instrumentation__.Notify(467602)
	return execinfrapb.BackfillerSpec{
		Table:                 desc,
		WriteAsOf:             writeAsOf,
		WriteAtBatchTimestamp: writeAtBatchTimestamp,
		ReadAsOf:              readAsOf,
		Type:                  execinfrapb.BackfillerSpec_Index,
		ChunkSize:             chunkSize,
		IndexesToBackfill:     indexesToBackfill,
	}, nil
}

func initIndexBackfillMergerSpec(
	desc descpb.TableDescriptor,
	addedIndexes []descpb.IndexID,
	temporaryIndexes []descpb.IndexID,
	mergeTimestamp hlc.Timestamp,
) (execinfrapb.IndexBackfillMergerSpec, error) {
	__antithesis_instrumentation__.Notify(467603)
	return execinfrapb.IndexBackfillMergerSpec{
		Table:            desc,
		AddedIndexes:     addedIndexes,
		TemporaryIndexes: temporaryIndexes,
		MergeTimestamp:   mergeTimestamp,
	}, nil
}

func (dsp *DistSQLPlanner) createBackfillerPhysicalPlan(
	ctx context.Context, planCtx *PlanningCtx, spec execinfrapb.BackfillerSpec, spans []roachpb.Span,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467604)
	spanPartitions, err := dsp.PartitionSpans(ctx, planCtx, spans)
	if err != nil {
		__antithesis_instrumentation__.Notify(467607)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467608)
	}
	__antithesis_instrumentation__.Notify(467605)

	p := planCtx.NewPhysicalPlan()
	p.ResultRouters = make([]physicalplan.ProcessorIdx, len(spanPartitions))
	for i, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(467609)
		ib := &execinfrapb.BackfillerSpec{}
		*ib = spec
		ib.InitialSplits = int32(len(spanPartitions))
		ib.Spans = sp.Spans

		proc := physicalplan.Processor{
			SQLInstanceID: sp.SQLInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Core:        execinfrapb.ProcessorCoreUnion{Backfiller: ib},
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				ResultTypes: []*types.T{},
			},
		}

		pIdx := p.AddProcessor(proc)
		p.ResultRouters[i] = pIdx
	}
	__antithesis_instrumentation__.Notify(467606)
	dsp.FinalizePlan(planCtx, p)
	return p, nil
}

func (dsp *DistSQLPlanner) createIndexBackfillerMergePhysicalPlan(
	ctx context.Context,
	planCtx *PlanningCtx,
	spec execinfrapb.IndexBackfillMergerSpec,
	spans [][]roachpb.Span,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467610)

	var n int
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(467616)
		for range sp {
			__antithesis_instrumentation__.Notify(467617)
			n++
		}
	}
	__antithesis_instrumentation__.Notify(467611)
	indexSpans := make([]roachpb.Span, 0, n)
	spanIdxs := make([]spanAndIndex, 0, n)
	spanIdxTree := interval.NewTree(interval.ExclusiveOverlapper)
	for i := range spans {
		__antithesis_instrumentation__.Notify(467618)
		for j := range spans[i] {
			__antithesis_instrumentation__.Notify(467619)
			indexSpans = append(indexSpans, spans[i][j])
			spanIdxs = append(spanIdxs, spanAndIndex{Span: spans[i][j], idx: i})
			if err := spanIdxTree.Insert(&spanIdxs[len(spanIdxs)-1], true); err != nil {
				__antithesis_instrumentation__.Notify(467620)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(467621)
			}

		}
	}
	__antithesis_instrumentation__.Notify(467612)
	spanIdxTree.AdjustRanges()
	getIndex := func(sp roachpb.Span) (idx int) {
		__antithesis_instrumentation__.Notify(467622)
		if !spanIdxTree.DoMatching(func(i interval.Interface) (done bool) {
			__antithesis_instrumentation__.Notify(467624)
			idx = i.(*spanAndIndex).idx
			return true
		}, sp.AsRange()) {
			__antithesis_instrumentation__.Notify(467625)
			panic(errors.AssertionFailedf("no matching index found for span: %s", sp))
		} else {
			__antithesis_instrumentation__.Notify(467626)
		}
		__antithesis_instrumentation__.Notify(467623)
		return idx
	}
	__antithesis_instrumentation__.Notify(467613)

	spanPartitions, err := dsp.PartitionSpans(ctx, planCtx, indexSpans)
	if err != nil {
		__antithesis_instrumentation__.Notify(467627)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467628)
	}
	__antithesis_instrumentation__.Notify(467614)

	p := planCtx.NewPhysicalPlan()
	p.ResultRouters = make([]physicalplan.ProcessorIdx, len(spanPartitions))
	for i, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(467629)
		ibm := &execinfrapb.IndexBackfillMergerSpec{}
		*ibm = spec

		ibm.Spans = sp.Spans
		for _, sp := range ibm.Spans {
			__antithesis_instrumentation__.Notify(467631)
			ibm.SpanIdx = append(ibm.SpanIdx, int32(getIndex(sp)))
		}
		__antithesis_instrumentation__.Notify(467630)

		proc := physicalplan.Processor{
			SQLInstanceID: sp.SQLInstanceID,
			Spec: execinfrapb.ProcessorSpec{
				Core:        execinfrapb.ProcessorCoreUnion{IndexBackfillMerger: ibm},
				Output:      []execinfrapb.OutputRouterSpec{{Type: execinfrapb.OutputRouterSpec_PASS_THROUGH}},
				ResultTypes: []*types.T{},
			},
		}

		pIdx := p.AddProcessor(proc)
		p.ResultRouters[i] = pIdx
	}
	__antithesis_instrumentation__.Notify(467615)
	dsp.FinalizePlan(planCtx, p)
	return p, nil
}

type spanAndIndex struct {
	roachpb.Span
	idx int
}

var _ interval.Interface = (*spanAndIndex)(nil)

func (si *spanAndIndex) Range() interval.Range {
	__antithesis_instrumentation__.Notify(467632)
	return si.AsRange()
}
func (si *spanAndIndex) ID() uintptr {
	__antithesis_instrumentation__.Notify(467633)
	return uintptr(unsafe.Pointer(si))
}
