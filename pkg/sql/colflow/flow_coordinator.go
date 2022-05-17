package colflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/colexec/colexecargs"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecerror"
	"github.com/cockroachdb/cockroach/pkg/sql/colexecop"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

type FlowCoordinator struct {
	execinfra.ProcessorBaseNoHelper
	colexecop.NonExplainable

	input execinfra.RowSource

	row  rowenc.EncDatumRow
	meta *execinfrapb.ProducerMetadata

	cancelFlow context.CancelFunc
}

var flowCoordinatorPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456077)
		return &FlowCoordinator{}
	},
}

func NewFlowCoordinator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	input execinfra.RowSource,
	output execinfra.RowReceiver,
	cancelFlow context.CancelFunc,
) *FlowCoordinator {
	__antithesis_instrumentation__.Notify(456078)
	f := flowCoordinatorPool.Get().(*FlowCoordinator)
	f.input = input
	f.cancelFlow = cancelFlow
	f.Init(
		f,
		flowCtx,

		flowCtx.EvalCtx,
		processorID,
		output,
		execinfra.ProcStateOpts{

			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(456080)

				f.close()
				return nil
			},
		},
	)
	__antithesis_instrumentation__.Notify(456079)
	f.AddInputToDrain(input)
	return f
}

var _ execinfra.OpNode = &FlowCoordinator{}
var _ execinfra.Processor = &FlowCoordinator{}
var _ execinfra.Releasable = &FlowCoordinator{}

func (f *FlowCoordinator) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(456081)
	return 1
}

func (f *FlowCoordinator) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(456082)
	if nth == 0 {
		__antithesis_instrumentation__.Notify(456084)

		return f.input.(execinfra.OpNode)
	} else {
		__antithesis_instrumentation__.Notify(456085)
	}
	__antithesis_instrumentation__.Notify(456083)
	colexecerror.InternalError(errors.AssertionFailedf("invalid index %d", nth))

	return nil
}

func (f *FlowCoordinator) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(456086)
	return f.input.OutputTypes()
}

func (f *FlowCoordinator) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456087)
	ctx = f.StartInternalNoSpan(ctx)
	if err := colexecerror.CatchVectorizedRuntimeError(func() {
		__antithesis_instrumentation__.Notify(456088)
		f.input.Start(ctx)
	}); err != nil {
		__antithesis_instrumentation__.Notify(456089)
		f.MoveToDraining(err)
	} else {
		__antithesis_instrumentation__.Notify(456090)
	}
}

func (f *FlowCoordinator) next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(456091)
	if f.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(456093)
		row, meta := f.input.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(456096)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(456098)
				f.MoveToDraining(nil)
			} else {
				__antithesis_instrumentation__.Notify(456099)
			}
			__antithesis_instrumentation__.Notify(456097)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(456100)
		}
		__antithesis_instrumentation__.Notify(456094)
		if row != nil {
			__antithesis_instrumentation__.Notify(456101)
			return row, nil
		} else {
			__antithesis_instrumentation__.Notify(456102)
		}
		__antithesis_instrumentation__.Notify(456095)

		f.MoveToDraining(nil)
	} else {
		__antithesis_instrumentation__.Notify(456103)
	}
	__antithesis_instrumentation__.Notify(456092)
	return nil, f.DrainHelper()
}

func (f *FlowCoordinator) nextAdapter() {
	__antithesis_instrumentation__.Notify(456104)
	f.row, f.meta = f.next()
}

func (f *FlowCoordinator) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(456105)
	if err := colexecerror.CatchVectorizedRuntimeError(f.nextAdapter); err != nil {
		__antithesis_instrumentation__.Notify(456107)
		f.MoveToDraining(err)
		return nil, f.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(456108)
	}
	__antithesis_instrumentation__.Notify(456106)
	return f.row, f.meta
}

func (f *FlowCoordinator) close() {
	__antithesis_instrumentation__.Notify(456109)
	if f.InternalClose() {
		__antithesis_instrumentation__.Notify(456110)
		f.cancelFlow()
	} else {
		__antithesis_instrumentation__.Notify(456111)
	}
}

func (f *FlowCoordinator) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(456112)
	f.close()
}

func (f *FlowCoordinator) Release() {
	__antithesis_instrumentation__.Notify(456113)
	f.ProcessorBaseNoHelper.Reset()
	*f = FlowCoordinator{

		ProcessorBaseNoHelper: f.ProcessorBaseNoHelper,
	}
	flowCoordinatorPool.Put(f)
}

type BatchFlowCoordinator struct {
	colexecop.OneInputNode
	colexecop.NonExplainable
	flowCtx     *execinfra.FlowCtx
	processorID int32

	input  colexecargs.OpWithMetaInfo
	output execinfra.BatchReceiver

	batch coldata.Batch

	cancelFlow context.CancelFunc
}

var batchFlowCoordinatorPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(456114)
		return &BatchFlowCoordinator{}
	},
}

func NewBatchFlowCoordinator(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	input colexecargs.OpWithMetaInfo,
	output execinfra.BatchReceiver,
	cancelFlow context.CancelFunc,
) *BatchFlowCoordinator {
	__antithesis_instrumentation__.Notify(456115)
	f := batchFlowCoordinatorPool.Get().(*BatchFlowCoordinator)
	*f = BatchFlowCoordinator{
		OneInputNode: colexecop.NewOneInputNode(input.Root),
		flowCtx:      flowCtx,
		processorID:  processorID,
		input:        input,
		output:       output,
		cancelFlow:   cancelFlow,
	}
	return f
}

var _ execinfra.OpNode = &BatchFlowCoordinator{}
var _ execinfra.Releasable = &BatchFlowCoordinator{}

func (f *BatchFlowCoordinator) init(ctx context.Context) error {
	return colexecerror.CatchVectorizedRuntimeError(func() {
		f.input.Root.Init(ctx)
	})
}

func (f *BatchFlowCoordinator) nextAdapter() {
	__antithesis_instrumentation__.Notify(456116)
	f.batch = f.input.Root.Next()
}

func (f *BatchFlowCoordinator) next() error {
	__antithesis_instrumentation__.Notify(456117)
	return colexecerror.CatchVectorizedRuntimeError(f.nextAdapter)
}

func (f *BatchFlowCoordinator) pushError(err error) execinfra.ConsumerStatus {
	__antithesis_instrumentation__.Notify(456118)
	meta := execinfrapb.GetProducerMeta()
	meta.Err = err
	return f.output.PushBatch(nil, meta)
}

func (f *BatchFlowCoordinator) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(456119)
	status := execinfra.NeedMoreRows

	ctx, span := execinfra.ProcessorSpan(ctx, "batch flow coordinator")
	if span != nil {
		__antithesis_instrumentation__.Notify(456125)
		if span.IsVerbose() {
			__antithesis_instrumentation__.Notify(456126)
			span.SetTag(execinfrapb.FlowIDTagKey, attribute.StringValue(f.flowCtx.ID.String()))
			span.SetTag(execinfrapb.ProcessorIDTagKey, attribute.IntValue(int(f.processorID)))
		} else {
			__antithesis_instrumentation__.Notify(456127)
		}
	} else {
		__antithesis_instrumentation__.Notify(456128)
	}
	__antithesis_instrumentation__.Notify(456120)

	defer func() {
		__antithesis_instrumentation__.Notify(456129)
		if err := f.close(ctx); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(456131)
			return status != execinfra.ConsumerClosed == true
		}() == true {
			__antithesis_instrumentation__.Notify(456132)
			f.pushError(err)
		} else {
			__antithesis_instrumentation__.Notify(456133)
		}
		__antithesis_instrumentation__.Notify(456130)
		f.output.ProducerDone()

		span.Finish()
	}()
	__antithesis_instrumentation__.Notify(456121)

	if err := f.init(ctx); err != nil {
		__antithesis_instrumentation__.Notify(456134)
		f.pushError(err)

		return
	} else {
		__antithesis_instrumentation__.Notify(456135)
	}
	__antithesis_instrumentation__.Notify(456122)

	for status == execinfra.NeedMoreRows {
		__antithesis_instrumentation__.Notify(456136)
		err := f.next()
		if err != nil {
			__antithesis_instrumentation__.Notify(456139)
			switch status = f.pushError(err); status {
			case execinfra.ConsumerClosed:
				__antithesis_instrumentation__.Notify(456141)
				return
			default:
				__antithesis_instrumentation__.Notify(456142)
			}
			__antithesis_instrumentation__.Notify(456140)
			continue
		} else {
			__antithesis_instrumentation__.Notify(456143)
		}
		__antithesis_instrumentation__.Notify(456137)
		if f.batch.Length() == 0 {
			__antithesis_instrumentation__.Notify(456144)

			break
		} else {
			__antithesis_instrumentation__.Notify(456145)
		}
		__antithesis_instrumentation__.Notify(456138)
		switch status = f.output.PushBatch(f.batch, nil); status {
		case execinfra.ConsumerClosed:
			__antithesis_instrumentation__.Notify(456146)
			return
		default:
			__antithesis_instrumentation__.Notify(456147)
		}
	}
	__antithesis_instrumentation__.Notify(456123)

	if span != nil {
		__antithesis_instrumentation__.Notify(456148)
		for _, s := range f.input.StatsCollectors {
			__antithesis_instrumentation__.Notify(456150)
			span.RecordStructured(s.GetStats())
		}
		__antithesis_instrumentation__.Notify(456149)
		if meta := execinfra.GetTraceDataAsMetadata(span); meta != nil {
			__antithesis_instrumentation__.Notify(456151)
			status = f.output.PushBatch(nil, meta)
			if status == execinfra.ConsumerClosed {
				__antithesis_instrumentation__.Notify(456152)
				return
			} else {
				__antithesis_instrumentation__.Notify(456153)
			}
		} else {
			__antithesis_instrumentation__.Notify(456154)
		}
	} else {
		__antithesis_instrumentation__.Notify(456155)
	}
	__antithesis_instrumentation__.Notify(456124)

	drainedMeta := f.input.MetadataSources.DrainMeta()
	for i := range drainedMeta {
		__antithesis_instrumentation__.Notify(456156)
		if execinfra.ShouldSwallowReadWithinUncertaintyIntervalError(&drainedMeta[i]) {
			__antithesis_instrumentation__.Notify(456158)

			continue
		} else {
			__antithesis_instrumentation__.Notify(456159)
		}
		__antithesis_instrumentation__.Notify(456157)
		status = f.output.PushBatch(nil, &drainedMeta[i])
		if status == execinfra.ConsumerClosed {
			__antithesis_instrumentation__.Notify(456160)
			return
		} else {
			__antithesis_instrumentation__.Notify(456161)
		}
	}
}

func (f *BatchFlowCoordinator) close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(456162)
	f.cancelFlow()
	var lastErr error
	for _, toClose := range f.input.ToClose {
		__antithesis_instrumentation__.Notify(456164)
		if err := toClose.Close(ctx); err != nil {
			__antithesis_instrumentation__.Notify(456165)
			lastErr = err
		} else {
			__antithesis_instrumentation__.Notify(456166)
		}
	}
	__antithesis_instrumentation__.Notify(456163)
	return lastErr
}

func (f *BatchFlowCoordinator) Release() {
	__antithesis_instrumentation__.Notify(456167)
	*f = BatchFlowCoordinator{}
	batchFlowCoordinatorPool.Put(f)
}
