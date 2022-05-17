package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

type countAggregator struct {
	execinfra.ProcessorBase

	input execinfra.RowSource
	count int
}

var _ execinfra.Processor = &countAggregator{}
var _ execinfra.RowSource = &countAggregator{}

const countRowsProcName = "count rows"

func newCountAggregator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*countAggregator, error) {
	__antithesis_instrumentation__.Notify(572087)
	ag := &countAggregator{}
	ag.input = input

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(572090)
		ag.input = newInputStatCollector(input)
		ag.ExecStatsForTrace = ag.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572091)
	}
	__antithesis_instrumentation__.Notify(572088)

	if err := ag.Init(
		ag,
		post,
		[]*types.T{types.Int},
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{ag.input},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(572092)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572093)
	}
	__antithesis_instrumentation__.Notify(572089)

	return ag, nil
}

func (ag *countAggregator) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572094)
	ctx = ag.StartInternal(ctx, countRowsProcName)
	ag.input.Start(ctx)
}

func (ag *countAggregator) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572095)
	for ag.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572097)
		row, meta := ag.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(572100)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572102)
				ag.MoveToDraining(meta.Err)
				break
			} else {
				__antithesis_instrumentation__.Notify(572103)
			}
			__antithesis_instrumentation__.Notify(572101)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572104)
		}
		__antithesis_instrumentation__.Notify(572098)
		if row == nil {
			__antithesis_instrumentation__.Notify(572105)
			ret := make(rowenc.EncDatumRow, 1)
			ret[0] = rowenc.EncDatum{Datum: tree.NewDInt(tree.DInt(ag.count))}
			rendered, _, err := ag.OutputHelper.ProcessRow(ag.Ctx, ret)

			ag.MoveToDraining(nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(572107)
				return nil, &execinfrapb.ProducerMetadata{Err: err}
			} else {
				__antithesis_instrumentation__.Notify(572108)
			}
			__antithesis_instrumentation__.Notify(572106)
			return rendered, nil
		} else {
			__antithesis_instrumentation__.Notify(572109)
		}
		__antithesis_instrumentation__.Notify(572099)
		ag.count++
	}
	__antithesis_instrumentation__.Notify(572096)
	return nil, ag.DrainHelper()
}

func (ag *countAggregator) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(572110)
	is, ok := getInputStats(ag.input)
	if !ok {
		__antithesis_instrumentation__.Notify(572112)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572113)
	}
	__antithesis_instrumentation__.Notify(572111)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Output: ag.OutputHelper.Stats(),
	}
}
