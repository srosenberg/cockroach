package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

func distImport(
	ctx context.Context,
	execCtx sql.JobExecContext,
	job *jobs.Job,
	tables map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
	typeDescs []*descpb.TypeDescriptor,
	from []string,
	format roachpb.IOFileFormat,
	walltime int64,
	alwaysFlushProgress bool,
	procsPerNode int,
) (roachpb.BulkOpSummary, error) {
	__antithesis_instrumentation__.Notify(494689)
	dsp := execCtx.DistSQLPlanner()
	evalCtx := execCtx.ExtendedEvalContext()

	planCtx, sqlInstanceIDs, err := dsp.SetupAllNodesPlanning(ctx, evalCtx, execCtx.ExecCfg())
	if err != nil {
		__antithesis_instrumentation__.Notify(494700)
		return roachpb.BulkOpSummary{}, err
	} else {
		__antithesis_instrumentation__.Notify(494701)
	}
	__antithesis_instrumentation__.Notify(494690)

	accumulatedBulkSummary := struct {
		syncutil.Mutex
		roachpb.BulkOpSummary
	}{}
	accumulatedBulkSummary.Lock()
	accumulatedBulkSummary.BulkOpSummary = getLastImportSummary(job)
	accumulatedBulkSummary.Unlock()

	inputSpecs := makeImportReaderSpecs(job, tables, typeDescs, from, format, sqlInstanceIDs, walltime,
		execCtx.User(), procsPerNode)

	p := planCtx.NewPhysicalPlan()

	corePlacement := make([]physicalplan.ProcessorCorePlacement, len(inputSpecs))
	for i := range inputSpecs {
		__antithesis_instrumentation__.Notify(494702)
		corePlacement[i].SQLInstanceID = sqlInstanceIDs[i%len(sqlInstanceIDs)]
		corePlacement[i].Core.ReadImport = inputSpecs[i]
	}
	__antithesis_instrumentation__.Notify(494691)
	p.AddNoInputStage(
		corePlacement,
		execinfrapb.PostProcessSpec{},

		[]*types.T{types.Bytes, types.Bytes},
		execinfrapb.Ordering{},
	)

	p.PlanToStreamColMap = []int{0, 1}

	dsp.FinalizePlan(planCtx, p)

	importDetails := job.Progress().Details.(*jobspb.Progress_Import).Import
	if importDetails.ReadProgress == nil {
		__antithesis_instrumentation__.Notify(494703)

		if err := job.FractionProgressed(ctx, nil,
			func(ctx context.Context, details jobspb.ProgressDetails) float32 {
				__antithesis_instrumentation__.Notify(494704)
				prog := details.(*jobspb.Progress_Import).Import
				prog.ReadProgress = make([]float32, len(from))
				prog.ResumePos = make([]int64, len(from))
				if prog.SequenceDetails == nil {
					__antithesis_instrumentation__.Notify(494706)
					prog.SequenceDetails = make([]*jobspb.SequenceDetails, len(from))
					for i := range prog.SequenceDetails {
						__antithesis_instrumentation__.Notify(494707)
						prog.SequenceDetails[i] = &jobspb.SequenceDetails{}
					}
				} else {
					__antithesis_instrumentation__.Notify(494708)
				}
				__antithesis_instrumentation__.Notify(494705)

				return 0.0
			},
		); err != nil {
			__antithesis_instrumentation__.Notify(494709)
			return roachpb.BulkOpSummary{}, err
		} else {
			__antithesis_instrumentation__.Notify(494710)
		}
	} else {
		__antithesis_instrumentation__.Notify(494711)
	}
	__antithesis_instrumentation__.Notify(494692)

	rowProgress := make([]int64, len(from))
	fractionProgress := make([]uint32, len(from))

	updateJobProgress := func() error {
		__antithesis_instrumentation__.Notify(494712)
		return job.FractionProgressed(ctx, nil,
			func(ctx context.Context, details jobspb.ProgressDetails) float32 {
				__antithesis_instrumentation__.Notify(494713)
				var overall float32
				prog := details.(*jobspb.Progress_Import).Import
				for i := range rowProgress {
					__antithesis_instrumentation__.Notify(494716)
					prog.ResumePos[i] = atomic.LoadInt64(&rowProgress[i])
				}
				__antithesis_instrumentation__.Notify(494714)
				for i := range fractionProgress {
					__antithesis_instrumentation__.Notify(494717)
					fileProgress := math.Float32frombits(atomic.LoadUint32(&fractionProgress[i]))
					prog.ReadProgress[i] = fileProgress
					overall += fileProgress
				}
				__antithesis_instrumentation__.Notify(494715)

				accumulatedBulkSummary.Lock()
				prog.Summary.Add(accumulatedBulkSummary.BulkOpSummary)
				accumulatedBulkSummary.Reset()
				accumulatedBulkSummary.Unlock()
				return overall / float32(len(from))
			},
		)
	}
	__antithesis_instrumentation__.Notify(494693)

	metaFn := func(_ context.Context, meta *execinfrapb.ProducerMetadata) error {
		__antithesis_instrumentation__.Notify(494718)
		if meta.BulkProcessorProgress != nil {
			__antithesis_instrumentation__.Notify(494720)
			for i, v := range meta.BulkProcessorProgress.ResumePos {
				__antithesis_instrumentation__.Notify(494723)
				atomic.StoreInt64(&rowProgress[i], v)
			}
			__antithesis_instrumentation__.Notify(494721)
			for i, v := range meta.BulkProcessorProgress.CompletedFraction {
				__antithesis_instrumentation__.Notify(494724)
				atomic.StoreUint32(&fractionProgress[i], math.Float32bits(v))
			}
			__antithesis_instrumentation__.Notify(494722)

			accumulatedBulkSummary.Lock()
			accumulatedBulkSummary.Add(meta.BulkProcessorProgress.BulkSummary)
			accumulatedBulkSummary.Unlock()

			if alwaysFlushProgress {
				__antithesis_instrumentation__.Notify(494725)
				return updateJobProgress()
			} else {
				__antithesis_instrumentation__.Notify(494726)
			}
		} else {
			__antithesis_instrumentation__.Notify(494727)
		}
		__antithesis_instrumentation__.Notify(494719)
		return nil
	}
	__antithesis_instrumentation__.Notify(494694)

	var res roachpb.BulkOpSummary
	rowResultWriter := sql.NewCallbackResultWriter(func(ctx context.Context, row tree.Datums) error {
		__antithesis_instrumentation__.Notify(494728)
		var counts roachpb.BulkOpSummary
		if err := protoutil.Unmarshal([]byte(*row[0].(*tree.DBytes)), &counts); err != nil {
			__antithesis_instrumentation__.Notify(494730)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494731)
		}
		__antithesis_instrumentation__.Notify(494729)
		res.Add(counts)
		return nil
	})
	__antithesis_instrumentation__.Notify(494695)

	if evalCtx.Codec.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(494732)
		if err := presplitTableBoundaries(ctx, execCtx.ExecCfg(), tables); err != nil {
			__antithesis_instrumentation__.Notify(494733)
			return roachpb.BulkOpSummary{}, err
		} else {
			__antithesis_instrumentation__.Notify(494734)
		}
	} else {
		__antithesis_instrumentation__.Notify(494735)
	}
	__antithesis_instrumentation__.Notify(494696)

	recv := sql.MakeDistSQLReceiver(
		ctx,
		sql.NewMetadataCallbackWriter(rowResultWriter, metaFn),
		tree.Rows,
		nil,
		nil,
		nil,
		evalCtx.Tracing,
		evalCtx.ExecCfg.ContentionRegistry,
		nil,
	)
	defer recv.Release()

	stopProgress := make(chan struct{})
	g := ctxgroup.WithContext(ctx)
	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(494736)
		tick := time.NewTicker(time.Second * 10)
		defer tick.Stop()
		done := ctx.Done()
		for {
			__antithesis_instrumentation__.Notify(494737)
			select {
			case <-stopProgress:
				__antithesis_instrumentation__.Notify(494738)
				return nil
			case <-done:
				__antithesis_instrumentation__.Notify(494739)
				return ctx.Err()
			case <-tick.C:
				__antithesis_instrumentation__.Notify(494740)
				if err := updateJobProgress(); err != nil {
					__antithesis_instrumentation__.Notify(494741)
					return err
				} else {
					__antithesis_instrumentation__.Notify(494742)
				}
			}
		}
	})
	__antithesis_instrumentation__.Notify(494697)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(494743)
		defer close(stopProgress)

		evalCtxCopy := *evalCtx
		dsp.Run(ctx, planCtx, nil, p, recv, &evalCtxCopy, nil)()
		return rowResultWriter.Err()
	})
	__antithesis_instrumentation__.Notify(494698)

	if err := g.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(494744)
		return roachpb.BulkOpSummary{}, err
	} else {
		__antithesis_instrumentation__.Notify(494745)
	}
	__antithesis_instrumentation__.Notify(494699)

	return res, nil
}

func getLastImportSummary(job *jobs.Job) roachpb.BulkOpSummary {
	__antithesis_instrumentation__.Notify(494746)
	progress := job.Progress()
	importProgress := progress.GetImport()
	return importProgress.Summary
}

func makeImportReaderSpecs(
	job *jobs.Job,
	tables map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
	typeDescs []*descpb.TypeDescriptor,
	from []string,
	format roachpb.IOFileFormat,
	sqlInstanceIDs []base.SQLInstanceID,
	walltime int64,
	user security.SQLUsername,
	procsPerNode int,
) []*execinfrapb.ReadImportDataSpec {
	__antithesis_instrumentation__.Notify(494747)
	details := job.Details().(jobspb.ImportDetails)

	inputSpecs := make([]*execinfrapb.ReadImportDataSpec, 0, len(sqlInstanceIDs)*procsPerNode)
	progress := job.Progress()
	importProgress := progress.GetImport()
	for i, input := range from {
		__antithesis_instrumentation__.Notify(494750)

		if i < cap(inputSpecs) {
			__antithesis_instrumentation__.Notify(494752)
			spec := &execinfrapb.ReadImportDataSpec{
				JobID:  int64(job.ID()),
				Tables: tables,
				Types:  typeDescs,
				Format: format,
				Progress: execinfrapb.JobProgress{
					JobID: job.ID(),
					Slot:  int32(i),
				},
				WalltimeNanos:         walltime,
				Uri:                   make(map[int32]string),
				ResumePos:             make(map[int32]int64),
				UserProto:             user.EncodeProto(),
				DatabasePrimaryRegion: details.DatabasePrimaryRegion,
				InitialSplits:         int32(len(sqlInstanceIDs)),
			}
			inputSpecs = append(inputSpecs, spec)
		} else {
			__antithesis_instrumentation__.Notify(494753)
		}
		__antithesis_instrumentation__.Notify(494751)
		n := i % len(inputSpecs)
		inputSpecs[n].Uri[int32(i)] = input
		if importProgress.ResumePos != nil {
			__antithesis_instrumentation__.Notify(494754)
			inputSpecs[n].ResumePos[int32(i)] = importProgress.ResumePos[int32(i)]
		} else {
			__antithesis_instrumentation__.Notify(494755)
		}
	}
	__antithesis_instrumentation__.Notify(494748)

	for i := range inputSpecs {
		__antithesis_instrumentation__.Notify(494756)

		inputSpecs[i].Progress.Contribution = float32(len(inputSpecs[i].Uri)) / float32(len(from))
	}
	__antithesis_instrumentation__.Notify(494749)
	return inputSpecs
}

func presplitTableBoundaries(
	ctx context.Context,
	cfg *sql.ExecutorConfig,
	tables map[string]*execinfrapb.ReadImportDataSpec_ImportTable,
) error {
	__antithesis_instrumentation__.Notify(494757)
	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, "import-pre-splitting-table-boundaries")
	defer span.Finish()
	expirationTime := cfg.DB.Clock().Now().Add(time.Hour.Nanoseconds(), 0)
	for _, tbl := range tables {
		__antithesis_instrumentation__.Notify(494759)

		tblDesc := tabledesc.NewBuilder(tbl.Desc).BuildImmutableTable()
		for _, span := range tblDesc.AllIndexSpans(cfg.Codec) {
			__antithesis_instrumentation__.Notify(494760)
			if err := cfg.DB.AdminSplit(ctx, span.Key, expirationTime); err != nil {
				__antithesis_instrumentation__.Notify(494761)
				return err
			} else {
				__antithesis_instrumentation__.Notify(494762)
			}
		}
	}
	__antithesis_instrumentation__.Notify(494758)
	return nil
}
