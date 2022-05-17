package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/errors"
)

type filtererProcessor struct {
	execinfra.ProcessorBase
	input  execinfra.RowSource
	filter *execinfrapb.ExprHelper
}

var _ execinfra.Processor = &filtererProcessor{}
var _ execinfra.RowSource = &filtererProcessor{}
var _ execinfra.OpNode = &filtererProcessor{}

const filtererProcName = "filterer"

func newFiltererProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.FiltererSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*filtererProcessor, error) {
	__antithesis_instrumentation__.Notify(572240)
	f := &filtererProcessor{input: input}
	types := input.OutputTypes()
	if err := f.Init(
		f,
		post,
		types,
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{InputsToDrain: []execinfra.RowSource{f.input}},
	); err != nil {
		__antithesis_instrumentation__.Notify(572244)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572245)
	}
	__antithesis_instrumentation__.Notify(572241)

	f.filter = &execinfrapb.ExprHelper{}
	if err := f.filter.Init(spec.Filter, types, &f.SemaCtx, f.EvalCtx); err != nil {
		__antithesis_instrumentation__.Notify(572246)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572247)
	}
	__antithesis_instrumentation__.Notify(572242)

	ctx := flowCtx.EvalCtx.Ctx()
	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(572248)
		f.input = newInputStatCollector(f.input)
		f.ExecStatsForTrace = f.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572249)
	}
	__antithesis_instrumentation__.Notify(572243)
	return f, nil
}

func (f *filtererProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572250)
	ctx = f.StartInternal(ctx, filtererProcName)
	f.input.Start(ctx)
}

func (f *filtererProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572251)
	for f.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572253)
		row, meta := f.input.Next()

		if meta != nil {
			__antithesis_instrumentation__.Notify(572257)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572259)
				f.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(572260)
			}
			__antithesis_instrumentation__.Notify(572258)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572261)
		}
		__antithesis_instrumentation__.Notify(572254)
		if row == nil {
			__antithesis_instrumentation__.Notify(572262)
			f.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(572263)
		}
		__antithesis_instrumentation__.Notify(572255)

		passes, err := f.filter.EvalFilter(row)
		if err != nil {
			__antithesis_instrumentation__.Notify(572264)
			f.MoveToDraining(err)
			break
		} else {
			__antithesis_instrumentation__.Notify(572265)
		}
		__antithesis_instrumentation__.Notify(572256)
		if passes {
			__antithesis_instrumentation__.Notify(572266)
			if outRow := f.ProcessRowHelper(row); outRow != nil {
				__antithesis_instrumentation__.Notify(572267)
				return outRow, nil
			} else {
				__antithesis_instrumentation__.Notify(572268)
			}
		} else {
			__antithesis_instrumentation__.Notify(572269)
		}
	}
	__antithesis_instrumentation__.Notify(572252)
	return nil, f.DrainHelper()
}

func (f *filtererProcessor) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(572270)
	is, ok := getInputStats(f.input)
	if !ok {
		__antithesis_instrumentation__.Notify(572272)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572273)
	}
	__antithesis_instrumentation__.Notify(572271)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Output: f.OutputHelper.Stats(),
	}
}

func (f *filtererProcessor) ChildCount(bool) int {
	__antithesis_instrumentation__.Notify(572274)
	if _, ok := f.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(572276)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(572277)
	}
	__antithesis_instrumentation__.Notify(572275)
	return 0
}

func (f *filtererProcessor) Child(nth int, _ bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(572278)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(572280)
		if n, ok := f.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572282)
			return n
		} else {
			__antithesis_instrumentation__.Notify(572283)
		}
		__antithesis_instrumentation__.Notify(572281)
		panic("input to filterer is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(572284)
	}
	__antithesis_instrumentation__.Notify(572279)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
