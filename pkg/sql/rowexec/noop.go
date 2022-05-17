package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/errors"
)

type noopProcessor struct {
	execinfra.ProcessorBase
	input execinfra.RowSource
}

var _ execinfra.Processor = &noopProcessor{}
var _ execinfra.RowSource = &noopProcessor{}
var _ execinfra.OpNode = &noopProcessor{}

const noopProcName = "noop"

func newNoopProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*noopProcessor, error) {
	__antithesis_instrumentation__.Notify(574012)
	n := &noopProcessor{input: input}
	if err := n.Init(
		n,
		post,
		input.OutputTypes(),
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{InputsToDrain: []execinfra.RowSource{n.input}},
	); err != nil {
		__antithesis_instrumentation__.Notify(574015)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574016)
	}
	__antithesis_instrumentation__.Notify(574013)
	ctx := flowCtx.EvalCtx.Ctx()
	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(574017)
		n.input = newInputStatCollector(n.input)
		n.ExecStatsForTrace = n.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(574018)
	}
	__antithesis_instrumentation__.Notify(574014)
	return n, nil
}

func (n *noopProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574019)
	ctx = n.StartInternal(ctx, noopProcName)
	n.input.Start(ctx)
}

func (n *noopProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574020)
	for n.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(574022)
		row, meta := n.input.Next()

		if meta != nil {
			__antithesis_instrumentation__.Notify(574025)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574027)
				n.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(574028)
			}
			__antithesis_instrumentation__.Notify(574026)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(574029)
		}
		__antithesis_instrumentation__.Notify(574023)
		if row == nil {
			__antithesis_instrumentation__.Notify(574030)
			n.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(574031)
		}
		__antithesis_instrumentation__.Notify(574024)

		if outRow := n.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(574032)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(574033)
		}
	}
	__antithesis_instrumentation__.Notify(574021)
	return nil, n.DrainHelper()
}

func (n *noopProcessor) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(574034)
	is, ok := getInputStats(n.input)
	if !ok {
		__antithesis_instrumentation__.Notify(574036)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(574037)
	}
	__antithesis_instrumentation__.Notify(574035)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Output: n.OutputHelper.Stats(),
	}
}

func (n *noopProcessor) ChildCount(bool) int {
	__antithesis_instrumentation__.Notify(574038)
	if _, ok := n.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(574040)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(574041)
	}
	__antithesis_instrumentation__.Notify(574039)
	return 0
}

func (n *noopProcessor) Child(nth int, _ bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(574042)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(574044)
		if n, ok := n.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(574046)
			return n
		} else {
			__antithesis_instrumentation__.Notify(574047)
		}
		__antithesis_instrumentation__.Notify(574045)
		panic("input to noop is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(574048)
	}
	__antithesis_instrumentation__.Notify(574043)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
