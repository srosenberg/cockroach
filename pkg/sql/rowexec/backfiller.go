package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/backfill"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

type chunkBackfiller interface {
	prepare(ctx context.Context) error

	close(ctx context.Context)

	runChunk(
		ctx context.Context,
		span roachpb.Span,
		chunkSize rowinfra.RowLimit,
		readAsOf hlc.Timestamp,
	) (roachpb.Key, error)

	CurrentBufferFill() float32

	flush(ctx context.Context) error
}

type backfiller struct {
	chunks chunkBackfiller

	name string

	filter backfill.MutationFilter

	spec        execinfrapb.BackfillerSpec
	output      execinfra.RowReceiver
	out         execinfra.ProcOutputHelper
	flowCtx     *execinfra.FlowCtx
	processorID int32
}

func (*backfiller) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(571905)

	return nil
}

func (*backfiller) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(571906)
	return false
}

func (b *backfiller) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(571907)
	opName := fmt.Sprintf("%sBackfiller", b.name)
	ctx = logtags.AddTag(ctx, opName, int(b.spec.Table.ID))
	ctx, span := execinfra.ProcessorSpan(ctx, opName)
	defer span.Finish()
	meta := b.doRun(ctx)
	execinfra.SendTraceData(ctx, b.output)
	if emitHelper(ctx, b.output, &b.out, nil, meta, func(ctx context.Context) { __antithesis_instrumentation__.Notify(571908) }) {
		__antithesis_instrumentation__.Notify(571909)
		b.output.ProducerDone()
	} else {
		__antithesis_instrumentation__.Notify(571910)
	}
}

func (b *backfiller) doRun(ctx context.Context) *execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(571911)
	semaCtx := tree.MakeSemaContext()
	if err := b.out.Init(&execinfrapb.PostProcessSpec{}, nil, &semaCtx, b.flowCtx.NewEvalCtx()); err != nil {
		__antithesis_instrumentation__.Notify(571914)
		return &execinfrapb.ProducerMetadata{Err: err}
	} else {
		__antithesis_instrumentation__.Notify(571915)
	}
	__antithesis_instrumentation__.Notify(571912)
	finishedSpans, err := b.mainLoop(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(571916)
		return &execinfrapb.ProducerMetadata{Err: err}
	} else {
		__antithesis_instrumentation__.Notify(571917)
	}
	__antithesis_instrumentation__.Notify(571913)
	var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
	prog.CompletedSpans = append(prog.CompletedSpans, finishedSpans...)
	return &execinfrapb.ProducerMetadata{BulkProcessorProgress: &prog}
}

func (b *backfiller) mainLoop(ctx context.Context) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(571918)
	if err := b.chunks.prepare(ctx); err != nil {
		__antithesis_instrumentation__.Notify(571922)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571923)
	}
	__antithesis_instrumentation__.Notify(571919)
	defer b.chunks.close(ctx)

	opportunisticCheckpointAfter := (b.spec.Duration * 4) / 5

	const opportunisticCheckpointThreshold = 0.8
	start := timeutil.Now()
	totalChunks := 0
	totalSpans := 0
	var finishedSpans roachpb.Spans

	for i := range b.spec.Spans {
		__antithesis_instrumentation__.Notify(571924)
		log.VEventf(ctx, 2, "%s backfiller starting span %d of %d: %s",
			b.name, i+1, len(b.spec.Spans), b.spec.Spans[i])
		chunks := 0
		todo := b.spec.Spans[i]
		for todo.Key != nil {
			__antithesis_instrumentation__.Notify(571927)
			log.VEventf(ctx, 3, "%s backfiller starting chunk %d: %s", b.name, chunks, todo)
			var err error
			todo.Key, err = b.chunks.runChunk(ctx, todo, rowinfra.RowLimit(b.spec.ChunkSize), b.spec.ReadAsOf)
			if err != nil {
				__antithesis_instrumentation__.Notify(571930)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(571931)
			}
			__antithesis_instrumentation__.Notify(571928)
			chunks++
			running := timeutil.Since(start)
			if running > opportunisticCheckpointAfter && func() bool {
				__antithesis_instrumentation__.Notify(571932)
				return b.chunks.CurrentBufferFill() > opportunisticCheckpointThreshold == true
			}() == true {
				__antithesis_instrumentation__.Notify(571933)
				break
			} else {
				__antithesis_instrumentation__.Notify(571934)
			}
			__antithesis_instrumentation__.Notify(571929)
			if running > b.spec.Duration {
				__antithesis_instrumentation__.Notify(571935)
				break
			} else {
				__antithesis_instrumentation__.Notify(571936)
			}
		}
		__antithesis_instrumentation__.Notify(571925)
		totalChunks += chunks

		if todo.Key != nil {
			__antithesis_instrumentation__.Notify(571937)
			log.VEventf(ctx, 2,
				"%s backfiller ran out of time on span %d of %d, will resume it at %s next time",
				b.name, i+1, len(b.spec.Spans), todo)
			finishedSpans = append(finishedSpans, roachpb.Span{Key: b.spec.Spans[i].Key, EndKey: todo.Key})
			break
		} else {
			__antithesis_instrumentation__.Notify(571938)
		}
		__antithesis_instrumentation__.Notify(571926)
		log.VEventf(ctx, 2, "%s backfiller finished span %d of %d: %s",
			b.name, i+1, len(b.spec.Spans), b.spec.Spans[i])
		totalSpans++
		finishedSpans = append(finishedSpans, b.spec.Spans[i])
	}
	__antithesis_instrumentation__.Notify(571920)

	log.VEventf(ctx, 3, "%s backfiller flushing...", b.name)
	if err := b.chunks.flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(571939)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571940)
	}
	__antithesis_instrumentation__.Notify(571921)
	log.VEventf(ctx, 2, "%s backfiller finished %d spans in %d chunks in %s",
		b.name, totalSpans, totalChunks, timeutil.Since(start))

	return finishedSpans, nil
}

func GetResumeSpans(
	ctx context.Context,
	jobsRegistry *jobs.Registry,
	txn *kv.Txn,
	codec keys.SQLCodec,
	col *descs.Collection,
	tableID descpb.ID,
	mutationID descpb.MutationID,
	filter backfill.MutationFilter,
) ([]roachpb.Span, *jobs.Job, int, error) {
	__antithesis_instrumentation__.Notify(571941)
	tableDesc, err := col.Direct().MustGetTableDescByID(ctx, txn, tableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(571950)
		return nil, nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(571951)
	}
	__antithesis_instrumentation__.Notify(571942)

	const noIndex = -1
	mutationIdx := noIndex
	for i, m := range tableDesc.AllMutations() {
		__antithesis_instrumentation__.Notify(571952)
		if m.MutationID() != mutationID {
			__antithesis_instrumentation__.Notify(571954)
			break
		} else {
			__antithesis_instrumentation__.Notify(571955)
		}
		__antithesis_instrumentation__.Notify(571953)
		if mutationIdx == noIndex && func() bool {
			__antithesis_instrumentation__.Notify(571956)
			return filter(m) == true
		}() == true {
			__antithesis_instrumentation__.Notify(571957)
			mutationIdx = i
		} else {
			__antithesis_instrumentation__.Notify(571958)
		}
	}
	__antithesis_instrumentation__.Notify(571943)

	if mutationIdx == noIndex {
		__antithesis_instrumentation__.Notify(571959)
		return nil, nil, 0, errors.AssertionFailedf(
			"mutation %d has completed", errors.Safe(mutationID))
	} else {
		__antithesis_instrumentation__.Notify(571960)
	}
	__antithesis_instrumentation__.Notify(571944)

	var jobID jobspb.JobID
	if len(tableDesc.GetMutationJobs()) > 0 {
		__antithesis_instrumentation__.Notify(571961)

		for _, job := range tableDesc.GetMutationJobs() {
			__antithesis_instrumentation__.Notify(571962)
			if job.MutationID == mutationID {
				__antithesis_instrumentation__.Notify(571963)
				jobID = job.JobID
				break
			} else {
				__antithesis_instrumentation__.Notify(571964)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(571965)
	}
	__antithesis_instrumentation__.Notify(571945)

	if jobID == 0 {
		__antithesis_instrumentation__.Notify(571966)
		log.Errorf(ctx, "mutation with no job: %d, table desc: %+v", mutationID, tableDesc)
		return nil, nil, 0, errors.AssertionFailedf(
			"no job found for mutation %d", errors.Safe(mutationID))
	} else {
		__antithesis_instrumentation__.Notify(571967)
	}
	__antithesis_instrumentation__.Notify(571946)

	job, err := jobsRegistry.LoadJobWithTxn(ctx, jobID, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(571968)
		return nil, nil, 0, errors.Wrapf(err, "can't find job %d", errors.Safe(jobID))
	} else {
		__antithesis_instrumentation__.Notify(571969)
	}
	__antithesis_instrumentation__.Notify(571947)
	details, ok := job.Details().(jobspb.SchemaChangeDetails)
	if !ok {
		__antithesis_instrumentation__.Notify(571970)
		return nil, nil, 0, errors.AssertionFailedf(
			"expected SchemaChangeDetails job type, got %T", job.Details())
	} else {
		__antithesis_instrumentation__.Notify(571971)
	}
	__antithesis_instrumentation__.Notify(571948)

	spanList := details.ResumeSpanList[mutationIdx].ResumeSpans
	prefix := codec.TenantPrefix()
	for i := range spanList {
		__antithesis_instrumentation__.Notify(571972)
		spanList[i], err = keys.RewriteSpanToTenantPrefix(spanList[i], prefix)
		if err != nil {
			__antithesis_instrumentation__.Notify(571973)
			return nil, nil, 0, err
		} else {
			__antithesis_instrumentation__.Notify(571974)
		}
	}
	__antithesis_instrumentation__.Notify(571949)

	return spanList, job, mutationIdx, nil
}

func SetResumeSpansInJob(
	ctx context.Context, spans []roachpb.Span, mutationIdx int, txn *kv.Txn, job *jobs.Job,
) error {
	__antithesis_instrumentation__.Notify(571975)
	details, ok := job.Details().(jobspb.SchemaChangeDetails)
	if !ok {
		__antithesis_instrumentation__.Notify(571977)
		return errors.Errorf("expected SchemaChangeDetails job type, got %T", job.Details())
	} else {
		__antithesis_instrumentation__.Notify(571978)
	}
	__antithesis_instrumentation__.Notify(571976)
	details.ResumeSpanList[mutationIdx].ResumeSpans = spans
	return job.SetDetails(ctx, txn, details)
}
