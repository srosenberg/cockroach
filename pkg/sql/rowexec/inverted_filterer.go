package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/inverted"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/invertedidx"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type invertedFiltererState int

const (
	ifrStateUnknown invertedFiltererState = iota

	ifrReadingInput

	ifrEmittingRows
)

type invertedFilterer struct {
	execinfra.ProcessorBase
	runningState   invertedFiltererState
	input          execinfra.RowSource
	invertedColIdx uint32

	diskMonitor *mon.BytesMonitor
	rc          *rowcontainer.DiskBackedNumberedRowContainer

	invertedEval batchedInvertedExprEvaluator

	evalResult []KeyIndex

	resultIdx int

	keyRow rowenc.EncDatumRow

	outputRow rowenc.EncDatumRow
}

var _ execinfra.Processor = &invertedFilterer{}
var _ execinfra.RowSource = &invertedFilterer{}
var _ execinfra.OpNode = &invertedFilterer{}

const invertedFiltererProcName = "inverted filterer"

func newInvertedFilterer(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.InvertedFiltererSpec,
	input execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (execinfra.RowSourcedProcessor, error) {
	__antithesis_instrumentation__.Notify(572752)
	ifr := &invertedFilterer{
		input:          input,
		invertedColIdx: spec.InvertedColIdx,
		invertedEval: batchedInvertedExprEvaluator{
			exprs: []*inverted.SpanExpressionProto{&spec.InvertedExpr},
		},
	}

	outputColTypes := input.OutputTypes()
	rcColTypes := make([]*types.T, len(outputColTypes)-1)
	copy(rcColTypes, outputColTypes[:ifr.invertedColIdx])
	copy(rcColTypes[ifr.invertedColIdx:], outputColTypes[ifr.invertedColIdx+1:])
	ifr.keyRow = make(rowenc.EncDatumRow, len(rcColTypes))
	ifr.outputRow = make(rowenc.EncDatumRow, len(outputColTypes))
	ifr.outputRow[ifr.invertedColIdx].Datum = tree.DNull

	if err := ifr.ProcessorBase.Init(
		ifr, post, outputColTypes, flowCtx, processorID, output, nil,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{ifr.input},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(572757)
				ifr.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(572758)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572759)
	}
	__antithesis_instrumentation__.Notify(572753)

	ctx := flowCtx.EvalCtx.Ctx()

	ifr.MemMonitor = execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, "inverted-filterer-limited")
	ifr.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, "inverted-filterer-disk")
	ifr.rc = rowcontainer.NewDiskBackedNumberedRowContainer(
		true,
		rcColTypes,
		ifr.EvalCtx,
		ifr.FlowCtx.Cfg.TempStorage,
		ifr.MemMonitor,
		ifr.diskMonitor,
	)

	if execinfra.ShouldCollectStats(flowCtx.EvalCtx.Ctx(), flowCtx) {
		__antithesis_instrumentation__.Notify(572760)
		ifr.input = newInputStatCollector(ifr.input)
		ifr.ExecStatsForTrace = ifr.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572761)
	}
	__antithesis_instrumentation__.Notify(572754)

	if spec.PreFiltererSpec != nil {
		__antithesis_instrumentation__.Notify(572762)
		semaCtx := flowCtx.NewSemaContext(flowCtx.EvalCtx.Txn)
		var exprHelper execinfrapb.ExprHelper
		colTypes := []*types.T{spec.PreFiltererSpec.Type}
		if err := exprHelper.Init(spec.PreFiltererSpec.Expression, colTypes, semaCtx, ifr.EvalCtx); err != nil {
			__antithesis_instrumentation__.Notify(572765)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(572766)
		}
		__antithesis_instrumentation__.Notify(572763)
		preFilterer, preFiltererState, err := invertedidx.NewBoundPreFilterer(
			spec.PreFiltererSpec.Type, exprHelper.Expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(572767)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(572768)
		}
		__antithesis_instrumentation__.Notify(572764)
		ifr.invertedEval.filterer = preFilterer
		ifr.invertedEval.preFilterState = append(ifr.invertedEval.preFilterState, preFiltererState)
	} else {
		__antithesis_instrumentation__.Notify(572769)
	}
	__antithesis_instrumentation__.Notify(572755)

	_, err := ifr.invertedEval.init()
	if err != nil {
		__antithesis_instrumentation__.Notify(572770)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572771)
	}
	__antithesis_instrumentation__.Notify(572756)

	return ifr, nil
}

func (ifr *invertedFilterer) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572772)

	for ifr.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572774)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch ifr.runningState {
		case ifrReadingInput:
			__antithesis_instrumentation__.Notify(572778)
			ifr.runningState, meta = ifr.readInput()
		case ifrEmittingRows:
			__antithesis_instrumentation__.Notify(572779)
			ifr.runningState, row, meta = ifr.emitRow()
		default:
			__antithesis_instrumentation__.Notify(572780)
			log.Fatalf(ifr.Ctx, "unsupported state: %d", ifr.runningState)
		}
		__antithesis_instrumentation__.Notify(572775)
		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(572781)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(572782)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572783)
		}
		__antithesis_instrumentation__.Notify(572776)
		if meta != nil {
			__antithesis_instrumentation__.Notify(572784)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572785)
		}
		__antithesis_instrumentation__.Notify(572777)
		if outRow := ifr.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(572786)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(572787)
		}
	}
	__antithesis_instrumentation__.Notify(572773)
	return nil, ifr.DrainHelper()
}

func (ifr *invertedFilterer) readInput() (invertedFiltererState, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572788)
	row, meta := ifr.input.Next()
	if meta != nil {
		__antithesis_instrumentation__.Notify(572795)
		if meta.Err != nil {
			__antithesis_instrumentation__.Notify(572797)
			ifr.MoveToDraining(nil)
			return ifrStateUnknown, meta
		} else {
			__antithesis_instrumentation__.Notify(572798)
		}
		__antithesis_instrumentation__.Notify(572796)
		return ifrReadingInput, meta
	} else {
		__antithesis_instrumentation__.Notify(572799)
	}
	__antithesis_instrumentation__.Notify(572789)
	if row == nil {
		__antithesis_instrumentation__.Notify(572800)
		log.VEventf(ifr.Ctx, 1, "no more input rows")
		evalResult := ifr.invertedEval.evaluate()
		ifr.rc.SetupForRead(ifr.Ctx, evalResult)

		ifr.evalResult = evalResult[0]
		return ifrEmittingRows, nil
	} else {
		__antithesis_instrumentation__.Notify(572801)
	}
	__antithesis_instrumentation__.Notify(572790)

	for i := range row {
		__antithesis_instrumentation__.Notify(572802)
		if row[i].IsUnset() {
			__antithesis_instrumentation__.Notify(572803)
			row[i].Datum = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(572804)
		}
	}
	__antithesis_instrumentation__.Notify(572791)

	enc := row[ifr.invertedColIdx].EncodedBytes()
	if len(enc) == 0 {
		__antithesis_instrumentation__.Notify(572805)

		if row[ifr.invertedColIdx].Datum == nil {
			__antithesis_instrumentation__.Notify(572808)
			ifr.MoveToDraining(errors.New("no datum found"))
			return ifrStateUnknown, ifr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572809)
		}
		__antithesis_instrumentation__.Notify(572806)
		if row[ifr.invertedColIdx].Datum.ResolvedType().Family() != types.BytesFamily {
			__antithesis_instrumentation__.Notify(572810)
			ifr.MoveToDraining(errors.New("inverted column should have type bytes"))
			return ifrStateUnknown, ifr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572811)
		}
		__antithesis_instrumentation__.Notify(572807)
		enc = []byte(*row[ifr.invertedColIdx].Datum.(*tree.DBytes))
	} else {
		__antithesis_instrumentation__.Notify(572812)
	}
	__antithesis_instrumentation__.Notify(572792)
	shouldAdd, err := ifr.invertedEval.prepareAddIndexRow(enc, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(572813)
		ifr.MoveToDraining(err)
		return ifrStateUnknown, ifr.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572814)
	}
	__antithesis_instrumentation__.Notify(572793)
	if shouldAdd {
		__antithesis_instrumentation__.Notify(572815)

		copy(ifr.keyRow, row[:ifr.invertedColIdx])
		copy(ifr.keyRow[ifr.invertedColIdx:], row[ifr.invertedColIdx+1:])
		keyIndex, err := ifr.rc.AddRow(ifr.Ctx, ifr.keyRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(572817)
			ifr.MoveToDraining(err)
			return ifrStateUnknown, ifr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572818)
		}
		__antithesis_instrumentation__.Notify(572816)
		if err = ifr.invertedEval.addIndexRow(keyIndex); err != nil {
			__antithesis_instrumentation__.Notify(572819)
			ifr.MoveToDraining(err)
			return ifrStateUnknown, ifr.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572820)
		}
	} else {
		__antithesis_instrumentation__.Notify(572821)
	}
	__antithesis_instrumentation__.Notify(572794)
	return ifrReadingInput, nil
}

func (ifr *invertedFilterer) emitRow() (
	invertedFiltererState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(572822)
	drainFunc := func(err error) (
		invertedFiltererState,
		rowenc.EncDatumRow,
		*execinfrapb.ProducerMetadata,
	) {
		__antithesis_instrumentation__.Notify(572826)
		ifr.MoveToDraining(err)
		return ifrStateUnknown, nil, ifr.DrainHelper()
	}
	__antithesis_instrumentation__.Notify(572823)
	if ifr.resultIdx >= len(ifr.evalResult) {
		__antithesis_instrumentation__.Notify(572827)

		return drainFunc(ifr.rc.UnsafeReset(ifr.Ctx))
	} else {
		__antithesis_instrumentation__.Notify(572828)
	}
	__antithesis_instrumentation__.Notify(572824)
	curRowIdx := ifr.resultIdx
	ifr.resultIdx++
	keyRow, err := ifr.rc.GetRow(ifr.Ctx, ifr.evalResult[curRowIdx], false)
	if err != nil {
		__antithesis_instrumentation__.Notify(572829)
		return drainFunc(err)
	} else {
		__antithesis_instrumentation__.Notify(572830)
	}
	__antithesis_instrumentation__.Notify(572825)
	copy(ifr.outputRow[:ifr.invertedColIdx], keyRow[:ifr.invertedColIdx])
	copy(ifr.outputRow[ifr.invertedColIdx+1:], keyRow[ifr.invertedColIdx:])
	return ifrEmittingRows, ifr.outputRow, nil
}

func (ifr *invertedFilterer) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572831)
	ctx = ifr.StartInternal(ctx, invertedFiltererProcName)
	ifr.input.Start(ctx)
	ifr.runningState = ifrReadingInput
}

func (ifr *invertedFilterer) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(572832)

	ifr.close()
}

func (ifr *invertedFilterer) close() {
	__antithesis_instrumentation__.Notify(572833)
	if ifr.InternalClose() {
		__antithesis_instrumentation__.Notify(572834)
		ifr.rc.Close(ifr.Ctx)
		if ifr.MemMonitor != nil {
			__antithesis_instrumentation__.Notify(572836)
			ifr.MemMonitor.Stop(ifr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(572837)
		}
		__antithesis_instrumentation__.Notify(572835)
		if ifr.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(572838)
			ifr.diskMonitor.Stop(ifr.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(572839)
		}
	} else {
		__antithesis_instrumentation__.Notify(572840)
	}
}

func (ifr *invertedFilterer) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(572841)
	is, ok := getInputStats(ifr.input)
	if !ok {
		__antithesis_instrumentation__.Notify(572843)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572844)
	}
	__antithesis_instrumentation__.Notify(572842)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{is},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem:  optional.MakeUint(uint64(ifr.MemMonitor.MaximumBytes())),
			MaxAllocatedDisk: optional.MakeUint(uint64(ifr.diskMonitor.MaximumBytes())),
		},
		Output: ifr.OutputHelper.Stats(),
	}
}

func (ifr *invertedFilterer) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(572845)
	if _, ok := ifr.input.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(572847)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(572848)
	}
	__antithesis_instrumentation__.Notify(572846)
	return 0
}

func (ifr *invertedFilterer) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(572849)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(572851)
		if n, ok := ifr.input.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572853)
			return n
		} else {
			__antithesis_instrumentation__.Notify(572854)
		}
		__antithesis_instrumentation__.Notify(572852)
		panic("input to invertedFilterer is not an execinfra.OpNode")
	} else {
		__antithesis_instrumentation__.Notify(572855)
	}
	__antithesis_instrumentation__.Notify(572850)
	panic(errors.AssertionFailedf("invalid index %d", nth))
}
