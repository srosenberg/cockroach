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

type ordinalityProcessor struct {
	execinfra.ProcessorBase

	input  execinfra.RowSource
	curCnt int64
}

var _ execinfra.Processor = &ordinalityProcessor{}
var _ execinfra.RowSource = &ordinalityProcessor{}

const ordinalityProcName = "ordinality"

func newOrdinalityProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.OrdinalitySpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.RowSourcedProcessor, error) {
	__antithesis_instrumentation__.Notify(574049)
	ctx := flowCtx.EvalCtx.Ctx()
	o := &ordinalityProcessor{input: input, curCnt: 1}

	colTypes := make([]*types.T, len(input.OutputTypes())+1)
	copy(colTypes, input.OutputTypes())
	colTypes[len(colTypes)-1] = types.Int
	if err := o.Init(
		o,
		post,
		colTypes,
		flowCtx,
		processorID,
		output,
		nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{o.input},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(574052)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(574053)
	}
	__antithesis_instrumentation__.Notify(574050)

	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(574054)
		o.input = newInputStatCollector(o.input)
		o.ExecStatsForTrace = o.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(574055)
	}
	__antithesis_instrumentation__.Notify(574051)

	return o, nil
}

func (o *ordinalityProcessor) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(574056)
	ctx = o.StartInternal(ctx, ordinalityProcName)
	o.input.Start(ctx)
}

func (o *ordinalityProcessor) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(574057)
	for o.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(574059)
		row, meta := o.input.Next()

		if meta != nil {
			__antithesis_instrumentation__.Notify(574062)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(574064)
				o.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(574065)
			}
			__antithesis_instrumentation__.Notify(574063)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(574066)
		}
		__antithesis_instrumentation__.Notify(574060)
		if row == nil {
			__antithesis_instrumentation__.Notify(574067)
			o.MoveToDraining(nil)
			break
		} else {
			__antithesis_instrumentation__.Notify(574068)
		}
		__antithesis_instrumentation__.Notify(574061)

		row = append(row, rowenc.DatumToEncDatum(types.Int, tree.NewDInt(tree.DInt(o.curCnt))))
		o.curCnt++
		if outRow := o.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(574069)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(574070)
		}
	}
	__antithesis_instrumentation__.Notify(574058)
	return nil, o.DrainHelper()

}

func (o *ordinalityProcessor) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(574071)
	is, ok := getInputStats(o.input)
	if !ok {
		__antithesis_instrumentation__.Notify(574073)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(574074)
	}
	__antithesis_instrumentation__.Notify(574072)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Output: o.OutputHelper.Stats(),
	}
}
