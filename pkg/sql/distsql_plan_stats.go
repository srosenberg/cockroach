package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/span"
	"github.com/cockroachdb/cockroach/pkg/sql/stats"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type requestedStat struct {
	columns             []descpb.ColumnID
	histogram           bool
	histogramMaxBuckets uint32
	name                string
	inverted            bool
}

const histogramSamples = 10000

var maxTimestampAge = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.stats.max_timestamp_age",
	"maximum age of timestamp during table statistics collection",
	5*time.Minute,
)

func (dsp *DistSQLPlanner) createStatsPlan(
	ctx context.Context,
	planCtx *PlanningCtx,
	desc catalog.TableDescriptor,
	reqStats []requestedStat,
	jobID jobspb.JobID,
	details jobspb.CreateStatsDetails,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467737)
	if len(reqStats) == 0 {
		__antithesis_instrumentation__.Notify(467750)
		return nil, errors.New("no stats requested")
	} else {
		__antithesis_instrumentation__.Notify(467751)
	}
	__antithesis_instrumentation__.Notify(467738)

	var colCfg scanColumnsConfig
	var tableColSet catalog.TableColSet
	for _, s := range reqStats {
		__antithesis_instrumentation__.Notify(467752)
		for _, c := range s.columns {
			__antithesis_instrumentation__.Notify(467753)
			if !tableColSet.Contains(c) {
				__antithesis_instrumentation__.Notify(467754)
				tableColSet.Add(c)
				colCfg.wantedColumns = append(colCfg.wantedColumns, c)
			} else {
				__antithesis_instrumentation__.Notify(467755)
			}
		}
	}
	__antithesis_instrumentation__.Notify(467739)

	scan := scanNode{desc: desc}
	err := scan.initDescDefaults(colCfg)
	if err != nil {
		__antithesis_instrumentation__.Notify(467756)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467757)
	}
	__antithesis_instrumentation__.Notify(467740)
	var colIdxMap catalog.TableColMap
	for i, c := range scan.cols {
		__antithesis_instrumentation__.Notify(467758)
		colIdxMap.Set(c.GetID(), i)
	}
	__antithesis_instrumentation__.Notify(467741)
	var sb span.Builder
	sb.Init(planCtx.EvalContext(), planCtx.ExtendedEvalCtx.Codec, desc, scan.index)
	scan.spans, err = sb.UnconstrainedSpans()
	if err != nil {
		__antithesis_instrumentation__.Notify(467759)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467760)
	}
	__antithesis_instrumentation__.Notify(467742)
	scan.isFull = true

	p, err := dsp.createTableReaders(ctx, planCtx, &scan)
	if err != nil {
		__antithesis_instrumentation__.Notify(467761)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467762)
	}
	__antithesis_instrumentation__.Notify(467743)

	if details.AsOf != nil {
		__antithesis_instrumentation__.Notify(467763)

		val := maxTimestampAge.Get(&dsp.st.SV)
		for i := range p.Processors {
			__antithesis_instrumentation__.Notify(467764)
			spec := p.Processors[i].Spec.Core.TableReader
			spec.MaxTimestampAgeNanos = uint64(val)
		}
	} else {
		__antithesis_instrumentation__.Notify(467765)
	}
	__antithesis_instrumentation__.Notify(467744)

	var sketchSpecs, invSketchSpecs []execinfrapb.SketchSpec
	sampledColumnIDs := make([]descpb.ColumnID, len(scan.cols))
	for _, s := range reqStats {
		__antithesis_instrumentation__.Notify(467766)
		spec := execinfrapb.SketchSpec{
			SketchType:          execinfrapb.SketchType_HLL_PLUS_PLUS_V1,
			GenerateHistogram:   s.histogram,
			HistogramMaxBuckets: s.histogramMaxBuckets,
			Columns:             make([]uint32, len(s.columns)),
			StatName:            s.name,
		}
		for i, colID := range s.columns {
			__antithesis_instrumentation__.Notify(467768)
			colIdx, ok := colIdxMap.Get(colID)
			if !ok {
				__antithesis_instrumentation__.Notify(467770)
				panic("necessary column not scanned")
			} else {
				__antithesis_instrumentation__.Notify(467771)
			}
			__antithesis_instrumentation__.Notify(467769)
			streamColIdx := uint32(p.PlanToStreamColMap[colIdx])
			spec.Columns[i] = streamColIdx
			sampledColumnIDs[streamColIdx] = colID
		}
		__antithesis_instrumentation__.Notify(467767)
		if s.inverted {
			__antithesis_instrumentation__.Notify(467772)

			if len(s.columns) == 1 {
				__antithesis_instrumentation__.Notify(467774)
				col := s.columns[0]
				for _, index := range desc.PublicNonPrimaryIndexes() {
					__antithesis_instrumentation__.Notify(467775)
					if index.GetType() == descpb.IndexDescriptor_INVERTED && func() bool {
						__antithesis_instrumentation__.Notify(467776)
						return index.InvertedColumnID() == col == true
					}() == true {
						__antithesis_instrumentation__.Notify(467777)
						spec.Index = index.IndexDesc()
						break
					} else {
						__antithesis_instrumentation__.Notify(467778)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(467779)
			}
			__antithesis_instrumentation__.Notify(467773)

			invSketchSpecs = append(invSketchSpecs, spec)
		} else {
			__antithesis_instrumentation__.Notify(467780)
			sketchSpecs = append(sketchSpecs, spec)
		}
	}
	__antithesis_instrumentation__.Notify(467745)

	sampler := &execinfrapb.SamplerSpec{
		Sketches:         sketchSpecs,
		InvertedSketches: invSketchSpecs,
	}
	for _, s := range reqStats {
		__antithesis_instrumentation__.Notify(467781)
		sampler.MaxFractionIdle = details.MaxFractionIdle
		if s.histogram {
			__antithesis_instrumentation__.Notify(467782)
			sampler.SampleSize = histogramSamples

			sampler.MinSampleSize = s.histogramMaxBuckets
		} else {
			__antithesis_instrumentation__.Notify(467783)
		}
	}
	__antithesis_instrumentation__.Notify(467746)

	outTypes := make([]*types.T, 0, len(p.GetResultTypes())+5)
	outTypes = append(outTypes, p.GetResultTypes()...)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Bytes)

	outTypes = append(outTypes, types.Int)

	outTypes = append(outTypes, types.Bytes)

	p.AddNoGroupingStage(
		execinfrapb.ProcessorCoreUnion{Sampler: sampler},
		execinfrapb.PostProcessSpec{},
		outTypes,
		execinfrapb.Ordering{},
	)

	tableStats, err := planCtx.ExtendedEvalCtx.ExecCfg.TableStatsCache.GetTableStats(ctx, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(467784)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(467785)
	}
	__antithesis_instrumentation__.Notify(467747)

	var rowsExpected uint64
	if len(tableStats) > 0 {
		__antithesis_instrumentation__.Notify(467786)
		overhead := stats.AutomaticStatisticsFractionStaleRows.Get(&dsp.st.SV)

		rowsExpected = uint64(int64(

			float64(tableStats[0].RowCount) * (1 + overhead),
		))
	} else {
		__antithesis_instrumentation__.Notify(467787)
	}
	__antithesis_instrumentation__.Notify(467748)

	agg := &execinfrapb.SampleAggregatorSpec{
		Sketches:         sketchSpecs,
		InvertedSketches: invSketchSpecs,
		SampleSize:       sampler.SampleSize,
		MinSampleSize:    sampler.MinSampleSize,
		SampledColumnIDs: sampledColumnIDs,
		TableID:          desc.GetID(),
		JobID:            jobID,
		RowsExpected:     rowsExpected,
	}

	node := dsp.gatewaySQLInstanceID
	if len(p.ResultRouters) == 1 {
		__antithesis_instrumentation__.Notify(467788)
		node = p.Processors[p.ResultRouters[0]].SQLInstanceID
	} else {
		__antithesis_instrumentation__.Notify(467789)
	}
	__antithesis_instrumentation__.Notify(467749)
	p.AddSingleGroupStage(
		node,
		execinfrapb.ProcessorCoreUnion{SampleAggregator: agg},
		execinfrapb.PostProcessSpec{},
		[]*types.T{},
	)

	p.PlanToStreamColMap = []int{}
	return p, nil
}

func (dsp *DistSQLPlanner) createPlanForCreateStats(
	ctx context.Context, planCtx *PlanningCtx, jobID jobspb.JobID, details jobspb.CreateStatsDetails,
) (*PhysicalPlan, error) {
	__antithesis_instrumentation__.Notify(467790)
	reqStats := make([]requestedStat, len(details.ColumnStats))
	histogramCollectionEnabled := stats.HistogramClusterMode.Get(&dsp.st.SV)
	for i := 0; i < len(reqStats); i++ {
		__antithesis_instrumentation__.Notify(467792)
		histogram := details.ColumnStats[i].HasHistogram && func() bool {
			__antithesis_instrumentation__.Notify(467794)
			return histogramCollectionEnabled == true
		}() == true
		var histogramMaxBuckets uint32 = stats.DefaultHistogramBuckets
		if details.ColumnStats[i].HistogramMaxBuckets > 0 {
			__antithesis_instrumentation__.Notify(467795)
			histogramMaxBuckets = details.ColumnStats[i].HistogramMaxBuckets
		} else {
			__antithesis_instrumentation__.Notify(467796)
		}
		__antithesis_instrumentation__.Notify(467793)
		reqStats[i] = requestedStat{
			columns:             details.ColumnStats[i].ColumnIDs,
			histogram:           histogram,
			histogramMaxBuckets: histogramMaxBuckets,
			name:                details.Name,
			inverted:            details.ColumnStats[i].Inverted,
		}
	}
	__antithesis_instrumentation__.Notify(467791)

	tableDesc := tabledesc.NewBuilder(&details.Table).BuildImmutableTable()
	return dsp.createStatsPlan(ctx, planCtx, tableDesc, reqStats, jobID, details)
}

func (dsp *DistSQLPlanner) planAndRunCreateStats(
	ctx context.Context,
	evalCtx *extendedEvalContext,
	planCtx *PlanningCtx,
	txn *kv.Txn,
	job *jobs.Job,
	resultWriter *RowResultWriter,
) error {
	__antithesis_instrumentation__.Notify(467797)
	ctx = logtags.AddTag(ctx, "create-stats-distsql", nil)

	details := job.Details().(jobspb.CreateStatsDetails)
	physPlan, err := dsp.createPlanForCreateStats(ctx, planCtx, job.ID(), details)
	if err != nil {
		__antithesis_instrumentation__.Notify(467799)
		return err
	} else {
		__antithesis_instrumentation__.Notify(467800)
	}
	__antithesis_instrumentation__.Notify(467798)

	dsp.FinalizePlan(planCtx, physPlan)

	recv := MakeDistSQLReceiver(
		ctx,
		resultWriter,
		tree.DDL,
		evalCtx.ExecCfg.RangeDescriptorCache,
		txn,
		evalCtx.ExecCfg.Clock,
		evalCtx.Tracing,
		evalCtx.ExecCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	dsp.Run(ctx, planCtx, txn, physPlan, recv, evalCtx, nil)()
	return resultWriter.Err()
}
