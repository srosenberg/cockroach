package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

type Processor interface {
	OutputTypes() []*types.T

	MustBeStreaming() bool

	Run(context.Context)
}

type DoesNotUseTxn interface {
	DoesNotUseTxn() bool
}

type ProcOutputHelper struct {
	numInternalCols int
	RowAlloc        rowenc.EncDatumRowAlloc

	renderExprs []execinfrapb.ExprHelper

	outputCols []uint32

	outputRow rowenc.EncDatumRow

	OutputTypes []*types.T

	offset uint64

	maxRowIdx uint64

	rowIdx uint64
}

func (h *ProcOutputHelper) Reset() {
	__antithesis_instrumentation__.Notify(471138)
	*h = ProcOutputHelper{
		renderExprs: h.renderExprs[:0],
		OutputTypes: h.OutputTypes[:0],
	}
}

func (h *ProcOutputHelper) Init(
	post *execinfrapb.PostProcessSpec,
	coreOutputTypes []*types.T,
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
) error {
	__antithesis_instrumentation__.Notify(471139)
	if !post.Projection && func() bool {
		__antithesis_instrumentation__.Notify(471145)
		return len(post.OutputColumns) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(471146)
		return errors.Errorf("post-processing has projection unset but output columns set: %s", post)
	} else {
		__antithesis_instrumentation__.Notify(471147)
	}
	__antithesis_instrumentation__.Notify(471140)
	if post.Projection && func() bool {
		__antithesis_instrumentation__.Notify(471148)
		return len(post.RenderExprs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(471149)
		return errors.Errorf("post-processing has both projection and rendering: %s", post)
	} else {
		__antithesis_instrumentation__.Notify(471150)
	}
	__antithesis_instrumentation__.Notify(471141)
	h.numInternalCols = len(coreOutputTypes)
	if post.Projection {
		__antithesis_instrumentation__.Notify(471151)
		for _, col := range post.OutputColumns {
			__antithesis_instrumentation__.Notify(471155)
			if int(col) >= h.numInternalCols {
				__antithesis_instrumentation__.Notify(471156)
				return errors.Errorf("invalid output column %d (only %d available)", col, h.numInternalCols)
			} else {
				__antithesis_instrumentation__.Notify(471157)
			}
		}
		__antithesis_instrumentation__.Notify(471152)
		h.outputCols = post.OutputColumns
		if h.outputCols == nil {
			__antithesis_instrumentation__.Notify(471158)

			h.outputCols = make([]uint32, 0)
		} else {
			__antithesis_instrumentation__.Notify(471159)
		}
		__antithesis_instrumentation__.Notify(471153)
		nOutputCols := len(h.outputCols)
		if cap(h.OutputTypes) >= nOutputCols {
			__antithesis_instrumentation__.Notify(471160)
			h.OutputTypes = h.OutputTypes[:nOutputCols]
		} else {
			__antithesis_instrumentation__.Notify(471161)
			h.OutputTypes = make([]*types.T, nOutputCols)
		}
		__antithesis_instrumentation__.Notify(471154)
		for i, c := range h.outputCols {
			__antithesis_instrumentation__.Notify(471162)
			h.OutputTypes[i] = coreOutputTypes[c]
		}
	} else {
		__antithesis_instrumentation__.Notify(471163)
		if nRenders := len(post.RenderExprs); nRenders > 0 {
			__antithesis_instrumentation__.Notify(471164)
			if cap(h.renderExprs) >= nRenders {
				__antithesis_instrumentation__.Notify(471167)
				h.renderExprs = h.renderExprs[:nRenders]
			} else {
				__antithesis_instrumentation__.Notify(471168)
				h.renderExprs = make([]execinfrapb.ExprHelper, nRenders)
			}
			__antithesis_instrumentation__.Notify(471165)
			if cap(h.OutputTypes) >= nRenders {
				__antithesis_instrumentation__.Notify(471169)
				h.OutputTypes = h.OutputTypes[:nRenders]
			} else {
				__antithesis_instrumentation__.Notify(471170)
				h.OutputTypes = make([]*types.T, nRenders)
			}
			__antithesis_instrumentation__.Notify(471166)
			for i, expr := range post.RenderExprs {
				__antithesis_instrumentation__.Notify(471171)
				h.renderExprs[i] = execinfrapb.ExprHelper{}
				if err := h.renderExprs[i].Init(expr, coreOutputTypes, semaCtx, evalCtx); err != nil {
					__antithesis_instrumentation__.Notify(471173)
					return err
				} else {
					__antithesis_instrumentation__.Notify(471174)
				}
				__antithesis_instrumentation__.Notify(471172)
				h.OutputTypes[i] = h.renderExprs[i].Expr.ResolvedType()
			}
		} else {
			__antithesis_instrumentation__.Notify(471175)

			if cap(h.OutputTypes) >= len(coreOutputTypes) {
				__antithesis_instrumentation__.Notify(471177)
				h.OutputTypes = h.OutputTypes[:len(coreOutputTypes)]
			} else {
				__antithesis_instrumentation__.Notify(471178)
				h.OutputTypes = make([]*types.T, len(coreOutputTypes))
			}
			__antithesis_instrumentation__.Notify(471176)
			copy(h.OutputTypes, coreOutputTypes)
		}
	}
	__antithesis_instrumentation__.Notify(471142)
	if h.outputCols != nil || func() bool {
		__antithesis_instrumentation__.Notify(471179)
		return len(h.renderExprs) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(471180)

		h.outputRow = h.RowAlloc.AllocRow(len(h.OutputTypes))
	} else {
		__antithesis_instrumentation__.Notify(471181)
	}
	__antithesis_instrumentation__.Notify(471143)

	h.offset = post.Offset
	if post.Limit == 0 || func() bool {
		__antithesis_instrumentation__.Notify(471182)
		return post.Limit >= math.MaxUint64-h.offset == true
	}() == true {
		__antithesis_instrumentation__.Notify(471183)
		h.maxRowIdx = math.MaxUint64
	} else {
		__antithesis_instrumentation__.Notify(471184)
		h.maxRowIdx = h.offset + post.Limit
	}
	__antithesis_instrumentation__.Notify(471144)

	return nil
}

func (h *ProcOutputHelper) NeededColumns() (colIdxs util.FastIntSet) {
	__antithesis_instrumentation__.Notify(471185)
	if h.outputCols == nil && func() bool {
		__antithesis_instrumentation__.Notify(471189)
		return len(h.renderExprs) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(471190)

		colIdxs.AddRange(0, h.numInternalCols-1)
		return colIdxs
	} else {
		__antithesis_instrumentation__.Notify(471191)
	}
	__antithesis_instrumentation__.Notify(471186)

	for _, c := range h.outputCols {
		__antithesis_instrumentation__.Notify(471192)
		colIdxs.Add(int(c))
	}
	__antithesis_instrumentation__.Notify(471187)

	for i := 0; i < h.numInternalCols; i++ {
		__antithesis_instrumentation__.Notify(471193)

		for j := range h.renderExprs {
			__antithesis_instrumentation__.Notify(471194)
			if h.renderExprs[j].Vars.IndexedVarUsed(i) {
				__antithesis_instrumentation__.Notify(471195)
				colIdxs.Add(i)
				break
			} else {
				__antithesis_instrumentation__.Notify(471196)
			}
		}
	}
	__antithesis_instrumentation__.Notify(471188)

	return colIdxs
}

func (h *ProcOutputHelper) EmitRow(
	ctx context.Context, row rowenc.EncDatumRow, output RowReceiver,
) (ConsumerStatus, error) {
	__antithesis_instrumentation__.Notify(471197)
	if output == nil {
		__antithesis_instrumentation__.Notify(471205)
		panic("output RowReceiver is not set for emitting rows")
	} else {
		__antithesis_instrumentation__.Notify(471206)
	}
	__antithesis_instrumentation__.Notify(471198)

	outRow, ok, err := h.ProcessRow(ctx, row)
	if err != nil {
		__antithesis_instrumentation__.Notify(471207)

		return NeedMoreRows, err
	} else {
		__antithesis_instrumentation__.Notify(471208)
	}
	__antithesis_instrumentation__.Notify(471199)
	if outRow == nil {
		__antithesis_instrumentation__.Notify(471209)
		if ok {
			__antithesis_instrumentation__.Notify(471211)
			return NeedMoreRows, nil
		} else {
			__antithesis_instrumentation__.Notify(471212)
		}
		__antithesis_instrumentation__.Notify(471210)
		return DrainRequested, nil
	} else {
		__antithesis_instrumentation__.Notify(471213)
	}
	__antithesis_instrumentation__.Notify(471200)

	if log.V(3) {
		__antithesis_instrumentation__.Notify(471214)
		log.InfofDepth(ctx, 1, "pushing row %s", outRow.String(h.OutputTypes))
	} else {
		__antithesis_instrumentation__.Notify(471215)
	}
	__antithesis_instrumentation__.Notify(471201)
	if r := output.Push(outRow, nil); r != NeedMoreRows {
		__antithesis_instrumentation__.Notify(471216)
		log.VEventf(ctx, 1, "no more rows required. drain requested: %t",
			r == DrainRequested)
		return r, nil
	} else {
		__antithesis_instrumentation__.Notify(471217)
	}
	__antithesis_instrumentation__.Notify(471202)
	if h.rowIdx == h.maxRowIdx {
		__antithesis_instrumentation__.Notify(471218)
		log.VEventf(ctx, 1, "hit row limit; asking producer to drain")
		return DrainRequested, nil
	} else {
		__antithesis_instrumentation__.Notify(471219)
	}
	__antithesis_instrumentation__.Notify(471203)
	status := NeedMoreRows
	if !ok {
		__antithesis_instrumentation__.Notify(471220)
		status = DrainRequested
	} else {
		__antithesis_instrumentation__.Notify(471221)
	}
	__antithesis_instrumentation__.Notify(471204)
	return status, nil
}

func (h *ProcOutputHelper) ProcessRow(
	ctx context.Context, row rowenc.EncDatumRow,
) (_ rowenc.EncDatumRow, moreRowsOK bool, _ error) {
	__antithesis_instrumentation__.Notify(471222)
	if h.rowIdx >= h.maxRowIdx {
		__antithesis_instrumentation__.Notify(471226)
		return nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(471227)
	}
	__antithesis_instrumentation__.Notify(471223)

	h.rowIdx++
	if h.rowIdx <= h.offset {
		__antithesis_instrumentation__.Notify(471228)

		return nil, true, nil
	} else {
		__antithesis_instrumentation__.Notify(471229)
	}
	__antithesis_instrumentation__.Notify(471224)

	if len(h.renderExprs) > 0 {
		__antithesis_instrumentation__.Notify(471230)

		for i := range h.renderExprs {
			__antithesis_instrumentation__.Notify(471231)
			datum, err := h.renderExprs[i].Eval(row)
			if err != nil {
				__antithesis_instrumentation__.Notify(471233)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(471234)
			}
			__antithesis_instrumentation__.Notify(471232)
			h.outputRow[i] = rowenc.DatumToEncDatum(h.OutputTypes[i], datum)
		}
	} else {
		__antithesis_instrumentation__.Notify(471235)
		if h.outputCols != nil {
			__antithesis_instrumentation__.Notify(471236)

			for i, col := range h.outputCols {
				__antithesis_instrumentation__.Notify(471237)
				h.outputRow[i] = row[col]
			}
		} else {
			__antithesis_instrumentation__.Notify(471238)

			return row, h.rowIdx < h.maxRowIdx, nil
		}
	}
	__antithesis_instrumentation__.Notify(471225)

	return h.outputRow, h.rowIdx < h.maxRowIdx, nil
}

func (h *ProcOutputHelper) consumerClosed() {
	__antithesis_instrumentation__.Notify(471239)
	h.rowIdx = h.maxRowIdx
}

func (h *ProcOutputHelper) Stats() execinfrapb.OutputStats {
	__antithesis_instrumentation__.Notify(471240)
	return execinfrapb.OutputStats{
		NumTuples: optional.MakeUint(h.rowIdx),
	}
}

type ProcessorConstructor func(
	ctx context.Context,
	flowCtx *FlowCtx,
	processorID int32,
	core *execinfrapb.ProcessorCoreUnion,
	post *execinfrapb.PostProcessSpec,
	inputs []RowSource,
	outputs []RowReceiver,
	localProcessors []LocalProcessor,
) (Processor, error)

type ProcessorBase struct {
	ProcessorBaseNoHelper

	OutputHelper ProcOutputHelper

	MemMonitor *mon.BytesMonitor

	SemaCtx tree.SemaContext
}

type ProcessorBaseNoHelper struct {
	self RowSource

	ProcessorID int32

	Output RowReceiver

	FlowCtx *FlowCtx

	EvalCtx *tree.EvalContext

	Closed bool

	Ctx  context.Context
	span *tracing.Span

	origCtx context.Context

	State procState

	ExecStatsForTrace func() *execinfrapb.ComponentStats

	trailingMetaCallback func() []execinfrapb.ProducerMetadata

	trailingMeta []execinfrapb.ProducerMetadata

	inputsToDrain []RowSource

	curInputToDrain int
}

func (pb *ProcessorBaseNoHelper) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(471241)
	return false
}

func (pb *ProcessorBaseNoHelper) Reset() {
	__antithesis_instrumentation__.Notify(471242)

	for i := range pb.trailingMeta {
		__antithesis_instrumentation__.Notify(471245)
		pb.trailingMeta[i] = execinfrapb.ProducerMetadata{}
	}
	__antithesis_instrumentation__.Notify(471243)
	for i := range pb.inputsToDrain {
		__antithesis_instrumentation__.Notify(471246)
		pb.inputsToDrain[i] = nil
	}
	__antithesis_instrumentation__.Notify(471244)
	*pb = ProcessorBaseNoHelper{
		trailingMeta:  pb.trailingMeta[:0],
		inputsToDrain: pb.inputsToDrain[:0],
	}
}

func (pb *ProcessorBase) Reset() {
	__antithesis_instrumentation__.Notify(471247)
	pb.ProcessorBaseNoHelper.Reset()
	pb.OutputHelper.Reset()
	*pb = ProcessorBase{
		ProcessorBaseNoHelper: pb.ProcessorBaseNoHelper,
		OutputHelper:          pb.OutputHelper,
	}
}

type procState int

const (
	StateRunning procState = iota

	StateDraining

	StateTrailingMeta

	StateExhausted
)

func (pb *ProcessorBaseNoHelper) MoveToDraining(err error) {
	__antithesis_instrumentation__.Notify(471248)
	if pb.State != StateRunning {
		__antithesis_instrumentation__.Notify(471251)

		if err != nil {
			__antithesis_instrumentation__.Notify(471253)
			logcrash.ReportOrPanic(
				pb.Ctx,
				&pb.FlowCtx.Cfg.Settings.SV,
				"MoveToDraining called in state %s with err: %+v",
				pb.State, err)
		} else {
			__antithesis_instrumentation__.Notify(471254)
		}
		__antithesis_instrumentation__.Notify(471252)
		return
	} else {
		__antithesis_instrumentation__.Notify(471255)
	}
	__antithesis_instrumentation__.Notify(471249)

	if err != nil {
		__antithesis_instrumentation__.Notify(471256)
		pb.trailingMeta = append(pb.trailingMeta, execinfrapb.ProducerMetadata{Err: err})
	} else {
		__antithesis_instrumentation__.Notify(471257)
	}
	__antithesis_instrumentation__.Notify(471250)
	if pb.curInputToDrain < len(pb.inputsToDrain) {
		__antithesis_instrumentation__.Notify(471258)

		pb.State = StateDraining
		for _, input := range pb.inputsToDrain[pb.curInputToDrain:] {
			__antithesis_instrumentation__.Notify(471259)
			input.ConsumerDone()
		}
	} else {
		__antithesis_instrumentation__.Notify(471260)
		pb.moveToTrailingMeta()
	}
}

func (pb *ProcessorBaseNoHelper) DrainHelper() *execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(471261)
	if pb.State == StateRunning {
		__antithesis_instrumentation__.Notify(471265)
		logcrash.ReportOrPanic(
			pb.Ctx,
			&pb.FlowCtx.Cfg.Settings.SV,
			"drain helper called in StateRunning",
		)
	} else {
		__antithesis_instrumentation__.Notify(471266)
	}
	__antithesis_instrumentation__.Notify(471262)

	if len(pb.trailingMeta) > 0 {
		__antithesis_instrumentation__.Notify(471267)
		return pb.popTrailingMeta()
	} else {
		__antithesis_instrumentation__.Notify(471268)
	}
	__antithesis_instrumentation__.Notify(471263)

	if pb.State != StateDraining {
		__antithesis_instrumentation__.Notify(471269)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(471270)
	}
	__antithesis_instrumentation__.Notify(471264)

	for {
		__antithesis_instrumentation__.Notify(471271)
		input := pb.inputsToDrain[pb.curInputToDrain]

		row, meta := input.Next()
		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(471273)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(471274)
			pb.curInputToDrain++
			if pb.curInputToDrain >= len(pb.inputsToDrain) {
				__antithesis_instrumentation__.Notify(471276)
				pb.moveToTrailingMeta()
				return pb.popTrailingMeta()
			} else {
				__antithesis_instrumentation__.Notify(471277)
			}
			__antithesis_instrumentation__.Notify(471275)
			continue
		} else {
			__antithesis_instrumentation__.Notify(471278)
		}
		__antithesis_instrumentation__.Notify(471272)
		if meta != nil {
			__antithesis_instrumentation__.Notify(471279)

			if ShouldSwallowReadWithinUncertaintyIntervalError(meta) {
				__antithesis_instrumentation__.Notify(471281)
				continue
			} else {
				__antithesis_instrumentation__.Notify(471282)
			}
			__antithesis_instrumentation__.Notify(471280)
			return meta
		} else {
			__antithesis_instrumentation__.Notify(471283)
		}
	}
}

func ShouldSwallowReadWithinUncertaintyIntervalError(meta *execinfrapb.ProducerMetadata) bool {
	__antithesis_instrumentation__.Notify(471284)
	if err := meta.Err; err != nil {
		__antithesis_instrumentation__.Notify(471286)

		if ure := (*roachpb.UnhandledRetryableError)(nil); errors.As(err, &ure) {
			__antithesis_instrumentation__.Notify(471287)
			if _, uncertain := ure.PErr.GetDetail().(*roachpb.ReadWithinUncertaintyIntervalError); uncertain {
				__antithesis_instrumentation__.Notify(471288)
				return true
			} else {
				__antithesis_instrumentation__.Notify(471289)
			}
		} else {
			__antithesis_instrumentation__.Notify(471290)
		}
	} else {
		__antithesis_instrumentation__.Notify(471291)
	}
	__antithesis_instrumentation__.Notify(471285)
	return false
}

func (pb *ProcessorBaseNoHelper) popTrailingMeta() *execinfrapb.ProducerMetadata {
	__antithesis_instrumentation__.Notify(471292)
	if len(pb.trailingMeta) > 0 {
		__antithesis_instrumentation__.Notify(471294)
		meta := &pb.trailingMeta[0]
		pb.trailingMeta = pb.trailingMeta[1:]
		return meta
	} else {
		__antithesis_instrumentation__.Notify(471295)
	}
	__antithesis_instrumentation__.Notify(471293)
	pb.State = StateExhausted
	return nil
}

type ExecStatsForTraceHijacker interface {
	HijackExecStatsForTrace() func() *execinfrapb.ComponentStats
}

var _ ExecStatsForTraceHijacker = &ProcessorBase{}

func (pb *ProcessorBase) HijackExecStatsForTrace() func() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(471296)
	execStatsForTrace := pb.ExecStatsForTrace
	pb.ExecStatsForTrace = nil
	return execStatsForTrace
}

func (pb *ProcessorBaseNoHelper) moveToTrailingMeta() {
	__antithesis_instrumentation__.Notify(471297)
	if pb.State == StateTrailingMeta || func() bool {
		__antithesis_instrumentation__.Notify(471301)
		return pb.State == StateExhausted == true
	}() == true {
		__antithesis_instrumentation__.Notify(471302)
		logcrash.ReportOrPanic(
			pb.Ctx,
			&pb.FlowCtx.Cfg.Settings.SV,
			"moveToTrailingMeta called in state: %s",
			pb.State,
		)
	} else {
		__antithesis_instrumentation__.Notify(471303)
	}
	__antithesis_instrumentation__.Notify(471298)

	pb.State = StateTrailingMeta
	if pb.span != nil {
		__antithesis_instrumentation__.Notify(471304)
		if pb.ExecStatsForTrace != nil {
			__antithesis_instrumentation__.Notify(471306)
			if stats := pb.ExecStatsForTrace(); stats != nil {
				__antithesis_instrumentation__.Notify(471307)
				stats.Component = pb.FlowCtx.ProcessorComponentID(pb.ProcessorID)
				pb.span.RecordStructured(stats)
			} else {
				__antithesis_instrumentation__.Notify(471308)
			}
		} else {
			__antithesis_instrumentation__.Notify(471309)
		}
		__antithesis_instrumentation__.Notify(471305)
		if trace := pb.span.GetConfiguredRecording(); trace != nil {
			__antithesis_instrumentation__.Notify(471310)
			pb.trailingMeta = append(pb.trailingMeta, execinfrapb.ProducerMetadata{TraceData: trace})
		} else {
			__antithesis_instrumentation__.Notify(471311)
		}
	} else {
		__antithesis_instrumentation__.Notify(471312)
	}
	__antithesis_instrumentation__.Notify(471299)

	if buildutil.CrdbTestBuild && func() bool {
		__antithesis_instrumentation__.Notify(471313)
		return pb.Ctx == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(471314)
		panic(
			errors.AssertionFailedf(
				"unexpected nil ProcessorBase.Ctx when draining. Was StartInternal called?",
			),
		)
	} else {
		__antithesis_instrumentation__.Notify(471315)
	}
	__antithesis_instrumentation__.Notify(471300)

	if pb.trailingMetaCallback != nil {
		__antithesis_instrumentation__.Notify(471316)
		pb.trailingMeta = append(pb.trailingMeta, pb.trailingMetaCallback()...)
	} else {
		__antithesis_instrumentation__.Notify(471317)
		pb.InternalClose()
	}
}

func (pb *ProcessorBase) ProcessRowHelper(row rowenc.EncDatumRow) rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(471318)
	outRow, ok, err := pb.OutputHelper.ProcessRow(pb.Ctx, row)
	if err != nil {
		__antithesis_instrumentation__.Notify(471321)
		pb.MoveToDraining(err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(471322)
	}
	__antithesis_instrumentation__.Notify(471319)
	if !ok {
		__antithesis_instrumentation__.Notify(471323)
		pb.MoveToDraining(nil)
	} else {
		__antithesis_instrumentation__.Notify(471324)
	}
	__antithesis_instrumentation__.Notify(471320)

	return outRow
}

func (pb *ProcessorBase) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(471325)
	return pb.OutputHelper.OutputTypes
}

func (pb *ProcessorBaseNoHelper) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(471326)
	if pb.Output == nil {
		__antithesis_instrumentation__.Notify(471328)
		panic("processor output is not set for emitting rows")
	} else {
		__antithesis_instrumentation__.Notify(471329)
	}
	__antithesis_instrumentation__.Notify(471327)
	pb.self.Start(ctx)
	Run(pb.Ctx, pb.self, pb.Output)
}

type ProcStateOpts struct {
	TrailingMetaCallback func() []execinfrapb.ProducerMetadata

	InputsToDrain []RowSource
}

func (pb *ProcessorBase) Init(
	self RowSource,
	post *execinfrapb.PostProcessSpec,
	coreOutputTypes []*types.T,
	flowCtx *FlowCtx,
	processorID int32,
	output RowReceiver,
	memMonitor *mon.BytesMonitor,
	opts ProcStateOpts,
) error {
	__antithesis_instrumentation__.Notify(471330)
	return pb.InitWithEvalCtx(
		self, post, coreOutputTypes, flowCtx, flowCtx.NewEvalCtx(), processorID, output, memMonitor, opts,
	)
}

func (pb *ProcessorBase) InitWithEvalCtx(
	self RowSource,
	post *execinfrapb.PostProcessSpec,
	coreOutputTypes []*types.T,
	flowCtx *FlowCtx,
	evalCtx *tree.EvalContext,
	processorID int32,
	output RowReceiver,
	memMonitor *mon.BytesMonitor,
	opts ProcStateOpts,
) error {
	__antithesis_instrumentation__.Notify(471331)
	pb.ProcessorBaseNoHelper.Init(
		self, flowCtx, evalCtx, processorID, output, opts,
	)
	pb.MemMonitor = memMonitor

	resolver := flowCtx.NewTypeResolver(evalCtx.Txn)
	if err := resolver.HydrateTypeSlice(evalCtx.Context, coreOutputTypes); err != nil {
		__antithesis_instrumentation__.Notify(471333)
		return err
	} else {
		__antithesis_instrumentation__.Notify(471334)
	}
	__antithesis_instrumentation__.Notify(471332)
	pb.SemaCtx = tree.MakeSemaContext()
	pb.SemaCtx.TypeResolver = &resolver

	return pb.OutputHelper.Init(post, coreOutputTypes, &pb.SemaCtx, pb.EvalCtx)
}

func (pb *ProcessorBaseNoHelper) Init(
	self RowSource,
	flowCtx *FlowCtx,
	evalCtx *tree.EvalContext,
	processorID int32,
	output RowReceiver,
	opts ProcStateOpts,
) {
	__antithesis_instrumentation__.Notify(471335)
	pb.self = self
	pb.FlowCtx = flowCtx
	pb.EvalCtx = evalCtx
	pb.ProcessorID = processorID
	pb.Output = output
	pb.trailingMetaCallback = opts.TrailingMetaCallback
	if opts.InputsToDrain != nil {
		__antithesis_instrumentation__.Notify(471336)

		pb.inputsToDrain = opts.InputsToDrain
	} else {
		__antithesis_instrumentation__.Notify(471337)
	}
}

func (pb *ProcessorBaseNoHelper) AddInputToDrain(input RowSource) {
	__antithesis_instrumentation__.Notify(471338)
	pb.inputsToDrain = append(pb.inputsToDrain, input)
}

func (pb *ProcessorBase) AppendTrailingMeta(meta execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(471339)
	pb.trailingMeta = append(pb.trailingMeta, meta)
}

func ProcessorSpan(ctx context.Context, name string) (context.Context, *tracing.Span) {
	__antithesis_instrumentation__.Notify(471340)
	sp := tracing.SpanFromContext(ctx)
	if sp == nil {
		__antithesis_instrumentation__.Notify(471342)
		return ctx, nil
	} else {
		__antithesis_instrumentation__.Notify(471343)
	}
	__antithesis_instrumentation__.Notify(471341)
	return sp.Tracer().StartSpanCtx(ctx, name,
		tracing.WithParent(sp), tracing.WithDetachedRecording())
}

func (pb *ProcessorBaseNoHelper) StartInternal(ctx context.Context, name string) context.Context {
	__antithesis_instrumentation__.Notify(471344)
	return pb.startImpl(ctx, true, name)
}

func (pb *ProcessorBaseNoHelper) StartInternalNoSpan(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(471345)
	return pb.startImpl(ctx, false, "")
}

func (pb *ProcessorBaseNoHelper) startImpl(
	ctx context.Context, createSpan bool, spanName string,
) context.Context {
	__antithesis_instrumentation__.Notify(471346)
	pb.origCtx = ctx
	if createSpan {
		__antithesis_instrumentation__.Notify(471348)
		pb.Ctx, pb.span = ProcessorSpan(ctx, spanName)
		if pb.span != nil && func() bool {
			__antithesis_instrumentation__.Notify(471349)
			return pb.span.IsVerbose() == true
		}() == true {
			__antithesis_instrumentation__.Notify(471350)
			pb.span.SetTag(execinfrapb.FlowIDTagKey, attribute.StringValue(pb.FlowCtx.ID.String()))
			pb.span.SetTag(execinfrapb.ProcessorIDTagKey, attribute.IntValue(int(pb.ProcessorID)))
		} else {
			__antithesis_instrumentation__.Notify(471351)
		}
	} else {
		__antithesis_instrumentation__.Notify(471352)
		pb.Ctx = ctx
	}
	__antithesis_instrumentation__.Notify(471347)
	pb.EvalCtx.Context = pb.Ctx
	return pb.Ctx
}

func (pb *ProcessorBase) InternalClose() bool {
	__antithesis_instrumentation__.Notify(471353)
	return pb.InternalCloseEx(nil)
}

func (pb *ProcessorBase) InternalCloseEx(onClose func()) bool {
	__antithesis_instrumentation__.Notify(471354)
	closing := pb.ProcessorBaseNoHelper.InternalCloseEx(onClose)
	if closing {
		__antithesis_instrumentation__.Notify(471356)

		pb.OutputHelper.consumerClosed()
	} else {
		__antithesis_instrumentation__.Notify(471357)
	}
	__antithesis_instrumentation__.Notify(471355)
	return closing
}

func (pb *ProcessorBaseNoHelper) InternalClose() bool {
	__antithesis_instrumentation__.Notify(471358)
	return pb.InternalCloseEx(nil)
}

func (pb *ProcessorBaseNoHelper) InternalCloseEx(onClose func()) bool {
	__antithesis_instrumentation__.Notify(471359)

	if pb.Closed {
		__antithesis_instrumentation__.Notify(471363)
		return false
	} else {
		__antithesis_instrumentation__.Notify(471364)
	}
	__antithesis_instrumentation__.Notify(471360)
	for _, input := range pb.inputsToDrain[pb.curInputToDrain:] {
		__antithesis_instrumentation__.Notify(471365)
		input.ConsumerClosed()
	}
	__antithesis_instrumentation__.Notify(471361)

	if onClose != nil {
		__antithesis_instrumentation__.Notify(471366)
		onClose()
	} else {
		__antithesis_instrumentation__.Notify(471367)
	}
	__antithesis_instrumentation__.Notify(471362)

	pb.Closed = true
	pb.span.Finish()
	pb.span = nil

	pb.Ctx = pb.origCtx
	pb.EvalCtx.Context = pb.origCtx
	return true
}

func (pb *ProcessorBaseNoHelper) ConsumerDone() {
	__antithesis_instrumentation__.Notify(471368)
	pb.MoveToDraining(nil)
}

func (pb *ProcessorBaseNoHelper) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(471369)

	pb.InternalClose()
}

func NewMonitor(ctx context.Context, parent *mon.BytesMonitor, name string) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(471370)
	monitor := mon.NewMonitorInheritWithLimit(name, 0, parent)
	monitor.Start(ctx, parent, mon.BoundAccount{})
	return monitor
}

func NewLimitedMonitor(
	ctx context.Context, parent *mon.BytesMonitor, flowCtx *FlowCtx, name string,
) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(471371)
	limitedMon := mon.NewMonitorInheritWithLimit(name, GetWorkMemLimit(flowCtx), parent)
	limitedMon.Start(ctx, parent, mon.BoundAccount{})
	return limitedMon
}

func NewLimitedMonitorNoFlowCtx(
	ctx context.Context,
	parent *mon.BytesMonitor,
	config *ServerConfig,
	sd *sessiondata.SessionData,
	name string,
) *mon.BytesMonitor {
	__antithesis_instrumentation__.Notify(471372)

	flowCtx := &FlowCtx{
		Cfg: config,
		EvalCtx: &tree.EvalContext{
			SessionDataStack: sessiondata.NewStack(sd),
		},
	}
	return NewLimitedMonitor(ctx, parent, flowCtx, name)
}

type LocalProcessor interface {
	RowSourcedProcessor

	InitWithOutput(flowCtx *FlowCtx, post *execinfrapb.PostProcessSpec, output RowReceiver) error

	SetInput(ctx context.Context, input RowSource) error
}

func HasParallelProcessors(flow *execinfrapb.FlowSpec) bool {
	__antithesis_instrumentation__.Notify(471373)
	var seen util.FastIntSet
	for _, p := range flow.Processors {
		__antithesis_instrumentation__.Notify(471375)
		if seen.Contains(int(p.StageID)) {
			__antithesis_instrumentation__.Notify(471377)
			return true
		} else {
			__antithesis_instrumentation__.Notify(471378)
		}
		__antithesis_instrumentation__.Notify(471376)
		seen.Add(int(p.StageID))
	}
	__antithesis_instrumentation__.Notify(471374)
	return false
}
