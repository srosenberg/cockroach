package changefeeddist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/physicalplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

var ChangefeedResultTypes = []*types.T{
	types.Bytes,
	types.String,
	types.Bytes,
	types.Bytes,
}

func StartDistChangefeed(
	ctx context.Context,
	execCtx sql.JobExecContext,
	jobID jobspb.JobID,
	details jobspb.ChangefeedDetails,
	trackedSpans []roachpb.Span,
	initialHighWater hlc.Timestamp,
	checkpoint jobspb.ChangefeedProgress_Checkpoint,
	resultsCh chan<- tree.Datums,
	knobs TestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(16686)

	var noTxn *kv.Txn

	dsp := execCtx.DistSQLPlanner()
	evalCtx := execCtx.ExtendedEvalContext()
	planCtx := dsp.NewPlanningCtx(ctx, evalCtx, nil, noTxn,
		sql.DistributionTypeAlways)

	var spanPartitions []sql.SpanPartition
	if details.SinkURI == `` {
		__antithesis_instrumentation__.Notify(16692)

		spanPartitions = []sql.SpanPartition{{SQLInstanceID: dsp.GatewayID(), Spans: trackedSpans}}
	} else {
		__antithesis_instrumentation__.Notify(16693)

		var err error
		spanPartitions, err = dsp.PartitionSpans(ctx, planCtx, trackedSpans)
		if err != nil {
			__antithesis_instrumentation__.Notify(16694)
			return err
		} else {
			__antithesis_instrumentation__.Notify(16695)
		}
	}
	__antithesis_instrumentation__.Notify(16687)

	aggregatorCheckpoint := execinfrapb.ChangeAggregatorSpec_Checkpoint{
		Spans: checkpoint.Spans,
	}

	aggregatorSpecs := make([]*execinfrapb.ChangeAggregatorSpec, len(spanPartitions))
	for i, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(16696)
		watches := make([]execinfrapb.ChangeAggregatorSpec_Watch, len(sp.Spans))
		for watchIdx, nodeSpan := range sp.Spans {
			__antithesis_instrumentation__.Notify(16698)
			watches[watchIdx] = execinfrapb.ChangeAggregatorSpec_Watch{
				Span:            nodeSpan,
				InitialResolved: initialHighWater,
			}
		}
		__antithesis_instrumentation__.Notify(16697)

		aggregatorSpecs[i] = &execinfrapb.ChangeAggregatorSpec{
			Watches:    watches,
			Checkpoint: aggregatorCheckpoint,
			Feed:       details,
			UserProto:  execCtx.User().EncodeProto(),
			JobID:      jobID,
		}
	}
	__antithesis_instrumentation__.Notify(16688)

	changeFrontierSpec := execinfrapb.ChangeFrontierSpec{
		TrackedSpans: trackedSpans,
		Feed:         details,
		JobID:        jobID,
		UserProto:    execCtx.User().EncodeProto(),
	}

	if knobs.OnDistflowSpec != nil {
		__antithesis_instrumentation__.Notify(16699)
		knobs.OnDistflowSpec(aggregatorSpecs, &changeFrontierSpec)
	} else {
		__antithesis_instrumentation__.Notify(16700)
	}
	__antithesis_instrumentation__.Notify(16689)

	aggregatorCorePlacement := make([]physicalplan.ProcessorCorePlacement, len(spanPartitions))
	for i, sp := range spanPartitions {
		__antithesis_instrumentation__.Notify(16701)
		aggregatorCorePlacement[i].SQLInstanceID = sp.SQLInstanceID
		aggregatorCorePlacement[i].Core.ChangeAggregator = aggregatorSpecs[i]
	}
	__antithesis_instrumentation__.Notify(16690)

	p := planCtx.NewPhysicalPlan()
	p.AddNoInputStage(aggregatorCorePlacement, execinfrapb.PostProcessSpec{}, ChangefeedResultTypes, execinfrapb.Ordering{})
	p.AddSingleGroupStage(
		dsp.GatewayID(),
		execinfrapb.ProcessorCoreUnion{ChangeFrontier: &changeFrontierSpec},
		execinfrapb.PostProcessSpec{},
		ChangefeedResultTypes,
	)

	p.PlanToStreamColMap = []int{1, 2, 3}
	dsp.FinalizePlan(planCtx, p)

	resultRows := makeChangefeedResultWriter(resultsCh)
	recv := sql.MakeDistSQLReceiver(
		ctx,
		resultRows,
		tree.Rows,
		execCtx.ExecCfg().RangeDescriptorCache,
		noTxn,
		nil,
		evalCtx.Tracing,
		execCtx.ExecCfg().ContentionRegistry,
		nil,
	)
	defer recv.Release()

	var finishedSetupFn func()
	if details.SinkURI != `` {
		__antithesis_instrumentation__.Notify(16702)

		finishedSetupFn = func() { __antithesis_instrumentation__.Notify(16703); resultsCh <- tree.Datums(nil) }
	} else {
		__antithesis_instrumentation__.Notify(16704)
	}
	__antithesis_instrumentation__.Notify(16691)

	evalCtxCopy := *evalCtx
	dsp.Run(ctx, planCtx, noTxn, p, recv, &evalCtxCopy, finishedSetupFn)()
	return resultRows.Err()
}

type changefeedResultWriter struct {
	rowsCh       chan<- tree.Datums
	rowsAffected int
	err          error
}

func makeChangefeedResultWriter(rowsCh chan<- tree.Datums) *changefeedResultWriter {
	__antithesis_instrumentation__.Notify(16705)
	return &changefeedResultWriter{rowsCh: rowsCh}
}

func (w *changefeedResultWriter) AddRow(ctx context.Context, row tree.Datums) error {
	__antithesis_instrumentation__.Notify(16706)

	row = append(tree.Datums(nil), row...)

	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(16707)
		return ctx.Err()
	case w.rowsCh <- row:
		__antithesis_instrumentation__.Notify(16708)
		return nil
	}
}
func (w *changefeedResultWriter) IncrementRowsAffected(ctx context.Context, n int) {
	__antithesis_instrumentation__.Notify(16709)
	w.rowsAffected += n
}
func (w *changefeedResultWriter) SetError(err error) {
	__antithesis_instrumentation__.Notify(16710)
	w.err = err
}
func (w *changefeedResultWriter) Err() error {
	__antithesis_instrumentation__.Notify(16711)
	return w.err
}
