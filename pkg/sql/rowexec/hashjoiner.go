package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/util/cancelchecker"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/errors"
)

type hashJoinerState int

const (
	hjStateUnknown hashJoinerState = iota

	hjBuilding

	hjReadingLeftSide

	hjProbingRow

	hjEmittingRightUnmatched
)

type hashJoiner struct {
	joinerBase

	eqCols [2][]uint32

	runningState hashJoinerState

	diskMonitor *mon.BytesMonitor

	leftSource, rightSource execinfra.RowSource

	nullEquality bool

	hashTable *rowcontainer.HashDiskBackedRowContainer

	probingRowState struct {
		row rowenc.EncDatumRow

		iter rowcontainer.RowMarkerIterator

		matched bool
	}

	emittingRightUnmatchedState struct {
		iter rowcontainer.RowIterator
	}

	cancelChecker cancelchecker.CancelChecker
}

var _ execinfra.Processor = &hashJoiner{}
var _ execinfra.RowSource = &hashJoiner{}
var _ execinfra.OpNode = &hashJoiner{}

const hashJoinerProcName = "hash joiner"

func newHashJoiner(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec *execinfrapb.HashJoinerSpec,
	leftSource execinfra.RowSource,
	rightSource execinfra.RowSource,
	post *execinfrapb.PostProcessSpec,
	output execinfra.RowReceiver,
) (*hashJoiner, error) {
	__antithesis_instrumentation__.Notify(572285)
	h := &hashJoiner{
		leftSource:   leftSource,
		rightSource:  rightSource,
		nullEquality: spec.Type.IsSetOpJoin(),
		eqCols: [2][]uint32{
			leftSide:  spec.LeftEqColumns,
			rightSide: spec.RightEqColumns,
		},
	}

	if err := h.joinerBase.init(
		h,
		flowCtx,
		processorID,
		leftSource.OutputTypes(),
		rightSource.OutputTypes(),
		spec.Type,
		spec.OnExpr,
		false,
		post,
		output,
		execinfra.ProcStateOpts{
			InputsToDrain: []execinfra.RowSource{h.leftSource, h.rightSource},
			TrailingMetaCallback: func() []execinfrapb.ProducerMetadata {
				__antithesis_instrumentation__.Notify(572288)
				h.close()
				return nil
			},
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(572289)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(572290)
	}
	__antithesis_instrumentation__.Notify(572286)

	ctx := h.FlowCtx.EvalCtx.Ctx()

	h.MemMonitor = execinfra.NewLimitedMonitor(ctx, flowCtx.EvalCtx.Mon, flowCtx, "hashjoiner-limited")
	h.diskMonitor = execinfra.NewMonitor(ctx, flowCtx.DiskMonitor, "hashjoiner-disk")
	h.hashTable = rowcontainer.NewHashDiskBackedRowContainer(
		h.EvalCtx, h.MemMonitor, h.diskMonitor, h.FlowCtx.Cfg.TempStorage,
	)

	if execinfra.ShouldCollectStats(ctx, flowCtx) {
		__antithesis_instrumentation__.Notify(572291)
		h.leftSource = newInputStatCollector(h.leftSource)
		h.rightSource = newInputStatCollector(h.rightSource)
		h.ExecStatsForTrace = h.execStatsForTrace
	} else {
		__antithesis_instrumentation__.Notify(572292)
	}
	__antithesis_instrumentation__.Notify(572287)

	return h, h.hashTable.Init(
		h.Ctx,
		shouldMarkRightSide(h.joinType),
		h.rightSource.OutputTypes(),
		h.eqCols[rightSide],
		h.nullEquality,
	)
}

func (h *hashJoiner) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(572293)
	ctx = h.StartInternal(ctx, hashJoinerProcName)
	h.leftSource.Start(ctx)
	h.rightSource.Start(ctx)
	h.cancelChecker.Reset(ctx)
	h.runningState = hjBuilding
}

func (h *hashJoiner) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572294)
	for h.State == execinfra.StateRunning {
		__antithesis_instrumentation__.Notify(572296)
		var row rowenc.EncDatumRow
		var meta *execinfrapb.ProducerMetadata
		switch h.runningState {
		case hjBuilding:
			__antithesis_instrumentation__.Notify(572300)
			h.runningState, row, meta = h.build()
		case hjReadingLeftSide:
			__antithesis_instrumentation__.Notify(572301)
			h.runningState, row, meta = h.readLeftSide()
		case hjProbingRow:
			__antithesis_instrumentation__.Notify(572302)
			h.runningState, row, meta = h.probeRow()
		case hjEmittingRightUnmatched:
			__antithesis_instrumentation__.Notify(572303)
			h.runningState, row, meta = h.emitRightUnmatched()
		default:
			__antithesis_instrumentation__.Notify(572304)
			log.Fatalf(h.Ctx, "unsupported state: %d", h.runningState)
		}
		__antithesis_instrumentation__.Notify(572297)

		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(572305)
			return meta == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(572306)
			continue
		} else {
			__antithesis_instrumentation__.Notify(572307)
		}
		__antithesis_instrumentation__.Notify(572298)
		if meta != nil {
			__antithesis_instrumentation__.Notify(572308)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572309)
		}
		__antithesis_instrumentation__.Notify(572299)
		if outRow := h.ProcessRowHelper(row); outRow != nil {
			__antithesis_instrumentation__.Notify(572310)
			return outRow, nil
		} else {
			__antithesis_instrumentation__.Notify(572311)
		}
	}
	__antithesis_instrumentation__.Notify(572295)
	return nil, h.DrainHelper()
}

func (h *hashJoiner) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(572312)
	h.close()
}

func (h *hashJoiner) build() (hashJoinerState, rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(572313)
	for {
		__antithesis_instrumentation__.Notify(572314)
		row, meta, emitDirectly, err := h.receiveNext(rightSide)
		if err != nil {
			__antithesis_instrumentation__.Notify(572317)
			h.MoveToDraining(err)
			return hjStateUnknown, nil, h.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572318)
			if meta != nil {
				__antithesis_instrumentation__.Notify(572319)
				if meta.Err != nil {
					__antithesis_instrumentation__.Notify(572321)
					h.MoveToDraining(nil)
					return hjStateUnknown, nil, meta
				} else {
					__antithesis_instrumentation__.Notify(572322)
				}
				__antithesis_instrumentation__.Notify(572320)
				return hjBuilding, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(572323)
				if emitDirectly {
					__antithesis_instrumentation__.Notify(572324)
					return hjBuilding, row, nil
				} else {
					__antithesis_instrumentation__.Notify(572325)
				}
			}
		}
		__antithesis_instrumentation__.Notify(572315)

		if row == nil {
			__antithesis_instrumentation__.Notify(572326)

			if h.hashTable.IsEmpty() && func() bool {
				__antithesis_instrumentation__.Notify(572329)
				return h.joinType.IsEmptyOutputWhenRightIsEmpty() == true
			}() == true {
				__antithesis_instrumentation__.Notify(572330)
				h.MoveToDraining(nil)
				return hjStateUnknown, nil, h.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(572331)
			}
			__antithesis_instrumentation__.Notify(572327)

			if err = h.hashTable.ReserveMarkMemoryMaybe(h.Ctx); err != nil {
				__antithesis_instrumentation__.Notify(572332)
				h.MoveToDraining(err)
				return hjStateUnknown, nil, h.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(572333)
			}
			__antithesis_instrumentation__.Notify(572328)
			return hjReadingLeftSide, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(572334)
		}
		__antithesis_instrumentation__.Notify(572316)

		err = h.hashTable.AddRow(h.Ctx, row)

		if err != nil {
			__antithesis_instrumentation__.Notify(572335)
			h.MoveToDraining(err)
			return hjStateUnknown, nil, h.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572336)
		}
	}
}

func (h *hashJoiner) readLeftSide() (
	hashJoinerState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(572337)
	row, meta, emitDirectly, err := h.receiveNext(leftSide)
	if err != nil {
		__antithesis_instrumentation__.Notify(572341)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572342)
		if meta != nil {
			__antithesis_instrumentation__.Notify(572343)
			if meta.Err != nil {
				__antithesis_instrumentation__.Notify(572345)
				h.MoveToDraining(nil)
				return hjStateUnknown, nil, meta
			} else {
				__antithesis_instrumentation__.Notify(572346)
			}
			__antithesis_instrumentation__.Notify(572344)
			return hjReadingLeftSide, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(572347)
			if emitDirectly {
				__antithesis_instrumentation__.Notify(572348)
				return hjReadingLeftSide, row, nil
			} else {
				__antithesis_instrumentation__.Notify(572349)
			}
		}
	}
	__antithesis_instrumentation__.Notify(572338)

	if row == nil {
		__antithesis_instrumentation__.Notify(572350)

		if shouldEmitUnmatchedRow(rightSide, h.joinType) {
			__antithesis_instrumentation__.Notify(572352)
			i := h.hashTable.NewUnmarkedIterator(h.Ctx)
			i.Rewind()
			h.emittingRightUnmatchedState.iter = i
			return hjEmittingRightUnmatched, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(572353)
		}
		__antithesis_instrumentation__.Notify(572351)
		h.MoveToDraining(nil)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572354)
	}
	__antithesis_instrumentation__.Notify(572339)

	h.probingRowState.row = row
	h.probingRowState.matched = false
	if h.probingRowState.iter == nil {
		__antithesis_instrumentation__.Notify(572355)
		i, err := h.hashTable.NewBucketIterator(h.Ctx, row, h.eqCols[leftSide])
		if err != nil {
			__antithesis_instrumentation__.Notify(572357)
			h.MoveToDraining(err)
			return hjStateUnknown, nil, h.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572358)
		}
		__antithesis_instrumentation__.Notify(572356)
		h.probingRowState.iter = i
	} else {
		__antithesis_instrumentation__.Notify(572359)
		if err := h.probingRowState.iter.Reset(h.Ctx, row); err != nil {
			__antithesis_instrumentation__.Notify(572360)
			h.MoveToDraining(err)
			return hjStateUnknown, nil, h.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572361)
		}
	}
	__antithesis_instrumentation__.Notify(572340)
	h.probingRowState.iter.Rewind()
	return hjProbingRow, nil, nil
}

func (h *hashJoiner) probeRow() (
	hashJoinerState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(572362)
	i := h.probingRowState.iter
	if ok, err := i.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(572371)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572372)
		if !ok {
			__antithesis_instrumentation__.Notify(572373)

			if h.probingRowState.matched {
				__antithesis_instrumentation__.Notify(572376)
				return hjReadingLeftSide, nil, nil
			} else {
				__antithesis_instrumentation__.Notify(572377)
			}
			__antithesis_instrumentation__.Notify(572374)

			if renderedRow, shouldEmit := h.shouldEmitUnmatched(
				h.probingRowState.row, leftSide,
			); shouldEmit {
				__antithesis_instrumentation__.Notify(572378)
				return hjReadingLeftSide, renderedRow, nil
			} else {
				__antithesis_instrumentation__.Notify(572379)
			}
			__antithesis_instrumentation__.Notify(572375)
			return hjReadingLeftSide, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(572380)
		}
	}
	__antithesis_instrumentation__.Notify(572363)

	if err := h.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(572381)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572382)
	}
	__antithesis_instrumentation__.Notify(572364)

	leftRow := h.probingRowState.row
	rightRow, err := i.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(572383)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572384)
	}
	__antithesis_instrumentation__.Notify(572365)
	defer i.Next()

	var renderedRow rowenc.EncDatumRow
	renderedRow, err = h.render(leftRow, rightRow)
	if err != nil {
		__antithesis_instrumentation__.Notify(572385)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572386)
	}
	__antithesis_instrumentation__.Notify(572366)

	if renderedRow == nil {
		__antithesis_instrumentation__.Notify(572387)
		return hjProbingRow, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(572388)
	}
	__antithesis_instrumentation__.Notify(572367)

	h.probingRowState.matched = true
	shouldEmit := true
	if shouldMarkRightSide(h.joinType) {
		__antithesis_instrumentation__.Notify(572389)
		if i.IsMarked(h.Ctx) {
			__antithesis_instrumentation__.Notify(572390)
			switch h.joinType {
			case descpb.RightSemiJoin:
				__antithesis_instrumentation__.Notify(572391)

				shouldEmit = false
			case descpb.IntersectAllJoin:
				__antithesis_instrumentation__.Notify(572392)

				shouldEmit = false
			case descpb.ExceptAllJoin:
				__antithesis_instrumentation__.Notify(572393)

				h.probingRowState.matched = false
			default:
				__antithesis_instrumentation__.Notify(572394)
			}
		} else {
			__antithesis_instrumentation__.Notify(572395)
			if err := i.Mark(h.Ctx); err != nil {
				__antithesis_instrumentation__.Notify(572396)
				h.MoveToDraining(err)
				return hjStateUnknown, nil, h.DrainHelper()
			} else {
				__antithesis_instrumentation__.Notify(572397)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(572398)
	}
	__antithesis_instrumentation__.Notify(572368)

	nextState := hjProbingRow
	switch h.joinType {
	case descpb.LeftSemiJoin:
		__antithesis_instrumentation__.Notify(572399)

		nextState = hjReadingLeftSide
	case descpb.LeftAntiJoin, descpb.RightAntiJoin:
		__antithesis_instrumentation__.Notify(572400)

		shouldEmit = false
	case descpb.ExceptAllJoin:
		__antithesis_instrumentation__.Notify(572401)

		shouldEmit = false
		if h.probingRowState.matched {
			__antithesis_instrumentation__.Notify(572404)

			nextState = hjReadingLeftSide
		} else {
			__antithesis_instrumentation__.Notify(572405)
		}
	case descpb.IntersectAllJoin:
		__antithesis_instrumentation__.Notify(572402)
		if shouldEmit {
			__antithesis_instrumentation__.Notify(572406)

			nextState = hjReadingLeftSide
		} else {
			__antithesis_instrumentation__.Notify(572407)
		}
	default:
		__antithesis_instrumentation__.Notify(572403)
	}
	__antithesis_instrumentation__.Notify(572369)

	if shouldEmit {
		__antithesis_instrumentation__.Notify(572408)
		return nextState, renderedRow, nil
	} else {
		__antithesis_instrumentation__.Notify(572409)
	}
	__antithesis_instrumentation__.Notify(572370)
	return nextState, nil, nil
}

func (h *hashJoiner) emitRightUnmatched() (
	hashJoinerState,
	rowenc.EncDatumRow,
	*execinfrapb.ProducerMetadata,
) {
	__antithesis_instrumentation__.Notify(572410)
	i := h.emittingRightUnmatchedState.iter
	if ok, err := i.Valid(); err != nil {
		__antithesis_instrumentation__.Notify(572414)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572415)
		if !ok {
			__antithesis_instrumentation__.Notify(572416)

			h.MoveToDraining(nil)
			return hjStateUnknown, nil, h.DrainHelper()
		} else {
			__antithesis_instrumentation__.Notify(572417)
		}
	}
	__antithesis_instrumentation__.Notify(572411)

	if err := h.cancelChecker.Check(); err != nil {
		__antithesis_instrumentation__.Notify(572418)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572419)
	}
	__antithesis_instrumentation__.Notify(572412)

	row, err := i.Row()
	if err != nil {
		__antithesis_instrumentation__.Notify(572420)
		h.MoveToDraining(err)
		return hjStateUnknown, nil, h.DrainHelper()
	} else {
		__antithesis_instrumentation__.Notify(572421)
	}
	__antithesis_instrumentation__.Notify(572413)
	defer i.Next()

	return hjEmittingRightUnmatched, h.renderUnmatchedRow(row, rightSide), nil
}

func (h *hashJoiner) close() {
	__antithesis_instrumentation__.Notify(572422)
	if h.InternalClose() {
		__antithesis_instrumentation__.Notify(572423)
		h.hashTable.Close(h.Ctx)
		if h.probingRowState.iter != nil {
			__antithesis_instrumentation__.Notify(572426)
			h.probingRowState.iter.Close()
		} else {
			__antithesis_instrumentation__.Notify(572427)
		}
		__antithesis_instrumentation__.Notify(572424)
		if h.emittingRightUnmatchedState.iter != nil {
			__antithesis_instrumentation__.Notify(572428)
			h.emittingRightUnmatchedState.iter.Close()
		} else {
			__antithesis_instrumentation__.Notify(572429)
		}
		__antithesis_instrumentation__.Notify(572425)
		h.MemMonitor.Stop(h.Ctx)
		if h.diskMonitor != nil {
			__antithesis_instrumentation__.Notify(572430)
			h.diskMonitor.Stop(h.Ctx)
		} else {
			__antithesis_instrumentation__.Notify(572431)
		}
	} else {
		__antithesis_instrumentation__.Notify(572432)
	}
}

func (h *hashJoiner) receiveNext(
	side joinSide,
) (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata, bool, error) {
	__antithesis_instrumentation__.Notify(572433)
	source := h.leftSource
	if side == rightSide {
		__antithesis_instrumentation__.Notify(572435)
		source = h.rightSource
	} else {
		__antithesis_instrumentation__.Notify(572436)
	}
	__antithesis_instrumentation__.Notify(572434)
	for {
		__antithesis_instrumentation__.Notify(572437)
		if err := h.cancelChecker.Check(); err != nil {
			__antithesis_instrumentation__.Notify(572442)
			return nil, nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(572443)
		}
		__antithesis_instrumentation__.Notify(572438)
		row, meta := source.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(572444)
			return nil, meta, false, nil
		} else {
			__antithesis_instrumentation__.Notify(572445)
			if row == nil {
				__antithesis_instrumentation__.Notify(572446)
				return nil, nil, false, nil
			} else {
				__antithesis_instrumentation__.Notify(572447)
			}
		}
		__antithesis_instrumentation__.Notify(572439)

		hasNull := false
		for _, c := range h.eqCols[side] {
			__antithesis_instrumentation__.Notify(572448)
			if row[c].IsNull() {
				__antithesis_instrumentation__.Notify(572449)
				hasNull = true
				break
			} else {
				__antithesis_instrumentation__.Notify(572450)
			}
		}
		__antithesis_instrumentation__.Notify(572440)

		if !hasNull || func() bool {
			__antithesis_instrumentation__.Notify(572451)
			return h.nullEquality == true
		}() == true {
			__antithesis_instrumentation__.Notify(572452)
			return row, nil, false, nil
		} else {
			__antithesis_instrumentation__.Notify(572453)
		}
		__antithesis_instrumentation__.Notify(572441)

		if renderedRow, shouldEmit := h.shouldEmitUnmatched(row, side); shouldEmit {
			__antithesis_instrumentation__.Notify(572454)
			return renderedRow, nil, true, nil
		} else {
			__antithesis_instrumentation__.Notify(572455)
		}

	}
}

func (h *hashJoiner) shouldEmitUnmatched(
	row rowenc.EncDatumRow, side joinSide,
) (rowenc.EncDatumRow, bool) {
	__antithesis_instrumentation__.Notify(572456)
	if !shouldEmitUnmatchedRow(side, h.joinType) {
		__antithesis_instrumentation__.Notify(572458)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(572459)
	}
	__antithesis_instrumentation__.Notify(572457)
	return h.renderUnmatchedRow(row, side), true
}

func (h *hashJoiner) execStatsForTrace() *execinfrapb.ComponentStats {
	__antithesis_instrumentation__.Notify(572460)
	lis, ok := getInputStats(h.leftSource)
	if !ok {
		__antithesis_instrumentation__.Notify(572463)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572464)
	}
	__antithesis_instrumentation__.Notify(572461)
	ris, ok := getInputStats(h.rightSource)
	if !ok {
		__antithesis_instrumentation__.Notify(572465)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(572466)
	}
	__antithesis_instrumentation__.Notify(572462)
	return &execinfrapb.ComponentStats{
		Inputs: []execinfrapb.InputStats{lis, ris},
		Exec: execinfrapb.ExecStats{
			MaxAllocatedMem:  optional.MakeUint(uint64(h.MemMonitor.MaximumBytes())),
			MaxAllocatedDisk: optional.MakeUint(uint64(h.diskMonitor.MaximumBytes())),
		},
		Output: h.OutputHelper.Stats(),
	}
}

func shouldMarkRightSide(joinType descpb.JoinType) bool {
	__antithesis_instrumentation__.Notify(572467)
	switch joinType {
	case descpb.FullOuterJoin, descpb.RightOuterJoin, descpb.RightAntiJoin:
		__antithesis_instrumentation__.Notify(572468)

		return true
	case descpb.RightSemiJoin:
		__antithesis_instrumentation__.Notify(572469)

		return true
	case descpb.IntersectAllJoin, descpb.ExceptAllJoin:
		__antithesis_instrumentation__.Notify(572470)

		return true
	default:
		__antithesis_instrumentation__.Notify(572471)
		return false
	}
}

func (h *hashJoiner) ChildCount(verbose bool) int {
	__antithesis_instrumentation__.Notify(572472)
	if _, ok := h.leftSource.(execinfra.OpNode); ok {
		__antithesis_instrumentation__.Notify(572474)
		if _, ok := h.rightSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572475)
			return 2
		} else {
			__antithesis_instrumentation__.Notify(572476)
		}
	} else {
		__antithesis_instrumentation__.Notify(572477)
	}
	__antithesis_instrumentation__.Notify(572473)
	return 0
}

func (h *hashJoiner) Child(nth int, verbose bool) execinfra.OpNode {
	__antithesis_instrumentation__.Notify(572478)
	switch nth {
	case 0:
		__antithesis_instrumentation__.Notify(572479)
		if n, ok := h.leftSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572484)
			return n
		} else {
			__antithesis_instrumentation__.Notify(572485)
		}
		__antithesis_instrumentation__.Notify(572480)
		panic("left input to hashJoiner is not an execinfra.OpNode")
	case 1:
		__antithesis_instrumentation__.Notify(572481)
		if n, ok := h.rightSource.(execinfra.OpNode); ok {
			__antithesis_instrumentation__.Notify(572486)
			return n
		} else {
			__antithesis_instrumentation__.Notify(572487)
		}
		__antithesis_instrumentation__.Notify(572482)
		panic("right input to hashJoiner is not an execinfra.OpNode")
	default:
		__antithesis_instrumentation__.Notify(572483)
		panic(errors.AssertionFailedf("invalid index %d", nth))
	}
}
