package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/rowcontainer"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type joinReaderStrategy interface {
	getLookupRowsBatchSizeHint(*sessiondata.SessionData) int64

	getMaxLookupKeyCols() int

	generateRemoteSpans() (roachpb.Spans, error)

	generatedRemoteSpans() bool

	processLookupRows(rows []rowenc.EncDatumRow) (roachpb.Spans, error)

	processLookedUpRow(ctx context.Context, row rowenc.EncDatumRow, key roachpb.Key) (joinReaderState, error)

	prepareToEmit(ctx context.Context)

	nextRowToEmit(ctx context.Context) (rowenc.EncDatumRow, joinReaderState, error)

	spilled() bool

	close(ctx context.Context)
}

type joinReaderNoOrderingStrategy struct {
	*joinerBase
	joinReaderSpanGenerator
	isPartialJoin        bool
	inputRows            []rowenc.EncDatumRow
	remoteSpansGenerated bool

	scratchMatchingInputRowIndices []int

	emitState struct {
		processingLookupRow bool

		unmatchedInputRowIndicesCursor int

		unmatchedInputRowIndices            []int
		unmatchedInputRowIndicesInitialized bool

		matchingInputRowIndicesCursor int
		matchingInputRowIndices       []int
		lookedUpRow                   rowenc.EncDatumRow
	}

	groupingState *inputBatchGroupingState

	memAcc *mon.BoundAccount
}

func (s *joinReaderNoOrderingStrategy) getLookupRowsBatchSizeHint(*sessiondata.SessionData) int64 {
	__antithesis_instrumentation__.Notify(573719)
	return 2 << 20
}

func (s *joinReaderNoOrderingStrategy) getMaxLookupKeyCols() int {
	__antithesis_instrumentation__.Notify(573720)
	return s.maxLookupCols()
}

func (s *joinReaderNoOrderingStrategy) generateRemoteSpans() (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573721)
	gen, ok := s.joinReaderSpanGenerator.(*localityOptimizedSpanGenerator)
	if !ok {
		__antithesis_instrumentation__.Notify(573723)
		return nil, errors.AssertionFailedf("generateRemoteSpans can only be called for locality optimized lookup joins")
	} else {
		__antithesis_instrumentation__.Notify(573724)
	}
	__antithesis_instrumentation__.Notify(573722)
	s.remoteSpansGenerated = true
	return gen.generateRemoteSpans(s.Ctx, s.inputRows)
}

func (s *joinReaderNoOrderingStrategy) generatedRemoteSpans() bool {
	__antithesis_instrumentation__.Notify(573725)
	return s.remoteSpansGenerated
}

func (s *joinReaderNoOrderingStrategy) processLookupRows(
	rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573726)
	s.inputRows = rows
	s.remoteSpansGenerated = false
	s.emitState.unmatchedInputRowIndicesInitialized = false
	return s.generateSpans(s.Ctx, s.inputRows)
}

func (s *joinReaderNoOrderingStrategy) processLookedUpRow(
	_ context.Context, row rowenc.EncDatumRow, key roachpb.Key,
) (joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573727)
	matchingInputRowIndices := s.getMatchingRowIndices(key)
	if s.isPartialJoin {
		__antithesis_instrumentation__.Notify(573729)

		s.scratchMatchingInputRowIndices = s.scratchMatchingInputRowIndices[:0]
		for _, inputRowIdx := range matchingInputRowIndices {
			__antithesis_instrumentation__.Notify(573731)
			if !s.groupingState.getMatched(inputRowIdx) {
				__antithesis_instrumentation__.Notify(573732)
				s.scratchMatchingInputRowIndices = append(s.scratchMatchingInputRowIndices, inputRowIdx)
			} else {
				__antithesis_instrumentation__.Notify(573733)
			}
		}
		__antithesis_instrumentation__.Notify(573730)
		matchingInputRowIndices = s.scratchMatchingInputRowIndices

		if err := s.memAcc.ResizeTo(s.Ctx, s.memUsage()); err != nil {
			__antithesis_instrumentation__.Notify(573734)
			return jrStateUnknown, addWorkmemHint(err)
		} else {
			__antithesis_instrumentation__.Notify(573735)
		}
	} else {
		__antithesis_instrumentation__.Notify(573736)
	}
	__antithesis_instrumentation__.Notify(573728)
	s.emitState.processingLookupRow = true
	s.emitState.lookedUpRow = row
	s.emitState.matchingInputRowIndices = matchingInputRowIndices
	s.emitState.matchingInputRowIndicesCursor = 0
	return jrEmittingRows, nil
}

func (s *joinReaderNoOrderingStrategy) prepareToEmit(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573737)
}

func (s *joinReaderNoOrderingStrategy) nextRowToEmit(
	_ context.Context,
) (rowenc.EncDatumRow, joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573738)
	if !s.emitState.processingLookupRow {
		__antithesis_instrumentation__.Notify(573741)

		if !shouldEmitUnmatchedRow(leftSide, s.joinType) {
			__antithesis_instrumentation__.Notify(573746)

			return nil, jrReadingInput, nil
		} else {
			__antithesis_instrumentation__.Notify(573747)
		}
		__antithesis_instrumentation__.Notify(573742)

		if !s.emitState.unmatchedInputRowIndicesInitialized {
			__antithesis_instrumentation__.Notify(573748)
			s.emitState.unmatchedInputRowIndices = s.emitState.unmatchedInputRowIndices[:0]
			for inputRowIdx := range s.inputRows {
				__antithesis_instrumentation__.Notify(573750)
				if s.groupingState.isUnmatched(inputRowIdx) {
					__antithesis_instrumentation__.Notify(573751)
					s.emitState.unmatchedInputRowIndices = append(s.emitState.unmatchedInputRowIndices, inputRowIdx)
				} else {
					__antithesis_instrumentation__.Notify(573752)
				}
			}
			__antithesis_instrumentation__.Notify(573749)
			s.emitState.unmatchedInputRowIndicesInitialized = true
			s.emitState.unmatchedInputRowIndicesCursor = 0

			if err := s.memAcc.ResizeTo(s.Ctx, s.memUsage()); err != nil {
				__antithesis_instrumentation__.Notify(573753)
				return nil, jrStateUnknown, addWorkmemHint(err)
			} else {
				__antithesis_instrumentation__.Notify(573754)
			}
		} else {
			__antithesis_instrumentation__.Notify(573755)
		}
		__antithesis_instrumentation__.Notify(573743)

		if s.emitState.unmatchedInputRowIndicesCursor >= len(s.emitState.unmatchedInputRowIndices) {
			__antithesis_instrumentation__.Notify(573756)

			return nil, jrReadingInput, nil
		} else {
			__antithesis_instrumentation__.Notify(573757)
		}
		__antithesis_instrumentation__.Notify(573744)
		inputRow := s.inputRows[s.emitState.unmatchedInputRowIndices[s.emitState.unmatchedInputRowIndicesCursor]]
		s.emitState.unmatchedInputRowIndicesCursor++
		if !s.joinType.ShouldIncludeRightColsInOutput() {
			__antithesis_instrumentation__.Notify(573758)
			return inputRow, jrEmittingRows, nil
		} else {
			__antithesis_instrumentation__.Notify(573759)
		}
		__antithesis_instrumentation__.Notify(573745)
		return s.renderUnmatchedRow(inputRow, leftSide), jrEmittingRows, nil
	} else {
		__antithesis_instrumentation__.Notify(573760)
	}
	__antithesis_instrumentation__.Notify(573739)

	for s.emitState.matchingInputRowIndicesCursor < len(s.emitState.matchingInputRowIndices) {
		__antithesis_instrumentation__.Notify(573761)
		inputRowIdx := s.emitState.matchingInputRowIndices[s.emitState.matchingInputRowIndicesCursor]
		s.emitState.matchingInputRowIndicesCursor++
		inputRow := s.inputRows[inputRowIdx]
		if s.joinType == descpb.LeftSemiJoin && func() bool {
			__antithesis_instrumentation__.Notify(573766)
			return s.groupingState.getMatched(inputRowIdx) == true
		}() == true {
			__antithesis_instrumentation__.Notify(573767)

			continue
		} else {
			__antithesis_instrumentation__.Notify(573768)
		}
		__antithesis_instrumentation__.Notify(573762)

		outputRow, err := s.render(inputRow, s.emitState.lookedUpRow)
		if err != nil {
			__antithesis_instrumentation__.Notify(573769)
			return nil, jrStateUnknown, err
		} else {
			__antithesis_instrumentation__.Notify(573770)
		}
		__antithesis_instrumentation__.Notify(573763)
		if outputRow == nil {
			__antithesis_instrumentation__.Notify(573771)

			continue
		} else {
			__antithesis_instrumentation__.Notify(573772)
		}
		__antithesis_instrumentation__.Notify(573764)

		s.groupingState.setMatched(inputRowIdx)
		if !s.joinType.ShouldIncludeRightColsInOutput() {
			__antithesis_instrumentation__.Notify(573773)
			if s.joinType == descpb.LeftAntiJoin {
				__antithesis_instrumentation__.Notify(573775)

				continue
			} else {
				__antithesis_instrumentation__.Notify(573776)
			}
			__antithesis_instrumentation__.Notify(573774)
			return inputRow, jrEmittingRows, nil
		} else {
			__antithesis_instrumentation__.Notify(573777)
		}
		__antithesis_instrumentation__.Notify(573765)
		return outputRow, jrEmittingRows, nil
	}
	__antithesis_instrumentation__.Notify(573740)

	s.emitState.processingLookupRow = false
	return nil, jrPerformingLookup, nil
}

func (s *joinReaderNoOrderingStrategy) spilled() bool {
	__antithesis_instrumentation__.Notify(573778)
	return false
}

func (s *joinReaderNoOrderingStrategy) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573779)
	s.memAcc.Close(ctx)
	s.joinReaderSpanGenerator.close(ctx)
	*s = joinReaderNoOrderingStrategy{}
}

func (s *joinReaderNoOrderingStrategy) memUsage() int64 {
	__antithesis_instrumentation__.Notify(573780)

	size := memsize.IntSliceOverhead + memsize.Int*int64(cap(s.scratchMatchingInputRowIndices))

	size += memsize.IntSliceOverhead + memsize.Int*int64(cap(s.emitState.unmatchedInputRowIndices))
	return size
}

type joinReaderIndexJoinStrategy struct {
	*joinerBase
	joinReaderSpanGenerator
	inputRows []rowenc.EncDatumRow

	emitState struct {
		processingLookupRow bool
		lookedUpRow         rowenc.EncDatumRow
	}

	memAcc *mon.BoundAccount
}

func (s *joinReaderIndexJoinStrategy) getLookupRowsBatchSizeHint(*sessiondata.SessionData) int64 {
	__antithesis_instrumentation__.Notify(573781)
	return 4 << 20
}

func (s *joinReaderIndexJoinStrategy) getMaxLookupKeyCols() int {
	__antithesis_instrumentation__.Notify(573782)
	return s.maxLookupCols()
}

func (s *joinReaderIndexJoinStrategy) generateRemoteSpans() (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573783)
	return nil, errors.AssertionFailedf("generateRemoteSpans called on an index join")
}

func (s *joinReaderIndexJoinStrategy) generatedRemoteSpans() bool {
	__antithesis_instrumentation__.Notify(573784)
	return false
}

func (s *joinReaderIndexJoinStrategy) processLookupRows(
	rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573785)
	s.inputRows = rows
	return s.generateSpans(s.Ctx, s.inputRows)
}

func (s *joinReaderIndexJoinStrategy) processLookedUpRow(
	_ context.Context, row rowenc.EncDatumRow, _ roachpb.Key,
) (joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573786)
	s.emitState.processingLookupRow = true
	s.emitState.lookedUpRow = row
	return jrEmittingRows, nil
}

func (s *joinReaderIndexJoinStrategy) prepareToEmit(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573787)
}

func (s *joinReaderIndexJoinStrategy) nextRowToEmit(
	ctx context.Context,
) (rowenc.EncDatumRow, joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573788)
	if !s.emitState.processingLookupRow {
		__antithesis_instrumentation__.Notify(573790)
		return nil, jrReadingInput, nil
	} else {
		__antithesis_instrumentation__.Notify(573791)
	}
	__antithesis_instrumentation__.Notify(573789)
	s.emitState.processingLookupRow = false
	return s.emitState.lookedUpRow, jrPerformingLookup, nil
}

func (s *joinReaderIndexJoinStrategy) spilled() bool {
	__antithesis_instrumentation__.Notify(573792)
	return false
}

func (s *joinReaderIndexJoinStrategy) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573793)
	s.memAcc.Close(ctx)
	s.joinReaderSpanGenerator.close(ctx)
	*s = joinReaderIndexJoinStrategy{}
}

var partialJoinSentinel = []int{-1}

type joinReaderOrderingStrategy struct {
	*joinerBase
	joinReaderSpanGenerator
	isPartialJoin bool

	inputRows            []rowenc.EncDatumRow
	remoteSpansGenerated bool

	inputRowIdxToLookedUpRowIndices [][]int

	lookedUpRows *rowcontainer.DiskBackedNumberedRowContainer

	emitCursor struct {
		inputRowIdx int

		outputRowIdx int
	}

	groupingState *inputBatchGroupingState

	outputGroupContinuationForLeftRow bool

	memAcc       *mon.BoundAccount
	accountedFor struct {
		sliceOverhead int64

		inputRowIdxToLookedUpRowIndices []int64
	}

	testingInfoSpilled bool
}

const joinReaderOrderingStrategyBatchSizeDefault = 10 << 10

var JoinReaderOrderingStrategyBatchSize = settings.RegisterByteSizeSetting(
	settings.TenantWritable,
	"sql.distsql.join_reader_ordering_strategy.batch_size",
	"size limit on the input rows to construct a single lookup KV batch",
	joinReaderOrderingStrategyBatchSizeDefault,
	settings.PositiveInt,
)

func (s *joinReaderOrderingStrategy) getLookupRowsBatchSizeHint(sd *sessiondata.SessionData) int64 {
	__antithesis_instrumentation__.Notify(573794)

	if sd.JoinReaderOrderingStrategyBatchSize == 0 {
		__antithesis_instrumentation__.Notify(573796)

		return joinReaderOrderingStrategyBatchSizeDefault
	} else {
		__antithesis_instrumentation__.Notify(573797)
	}
	__antithesis_instrumentation__.Notify(573795)
	return sd.JoinReaderOrderingStrategyBatchSize
}

func (s *joinReaderOrderingStrategy) getMaxLookupKeyCols() int {
	__antithesis_instrumentation__.Notify(573798)
	return s.maxLookupCols()
}

func (s *joinReaderOrderingStrategy) generateRemoteSpans() (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573799)
	gen, ok := s.joinReaderSpanGenerator.(*localityOptimizedSpanGenerator)
	if !ok {
		__antithesis_instrumentation__.Notify(573801)
		return nil, errors.AssertionFailedf("generateRemoteSpans can only be called for locality optimized lookup joins")
	} else {
		__antithesis_instrumentation__.Notify(573802)
	}
	__antithesis_instrumentation__.Notify(573800)
	s.remoteSpansGenerated = true
	return gen.generateRemoteSpans(s.Ctx, s.inputRows)
}

func (s *joinReaderOrderingStrategy) generatedRemoteSpans() bool {
	__antithesis_instrumentation__.Notify(573803)
	return s.remoteSpansGenerated
}

func (s *joinReaderOrderingStrategy) processLookupRows(
	rows []rowenc.EncDatumRow,
) (roachpb.Spans, error) {
	__antithesis_instrumentation__.Notify(573804)

	if cap(s.inputRowIdxToLookedUpRowIndices) >= len(rows) {
		__antithesis_instrumentation__.Notify(573808)

		s.inputRowIdxToLookedUpRowIndices = s.inputRowIdxToLookedUpRowIndices[:len(rows)]
		for i := range s.inputRowIdxToLookedUpRowIndices {
			__antithesis_instrumentation__.Notify(573809)
			s.inputRowIdxToLookedUpRowIndices[i] = s.inputRowIdxToLookedUpRowIndices[i][:0]
		}
	} else {
		__antithesis_instrumentation__.Notify(573810)

		oldSlices := s.inputRowIdxToLookedUpRowIndices

		oldSlices = oldSlices[:cap(oldSlices)]
		s.inputRowIdxToLookedUpRowIndices = make([][]int, len(rows))
		for i := range oldSlices {
			__antithesis_instrumentation__.Notify(573811)
			s.inputRowIdxToLookedUpRowIndices[i] = oldSlices[i][:0]
		}
	}
	__antithesis_instrumentation__.Notify(573805)

	if cap(s.accountedFor.inputRowIdxToLookedUpRowIndices) >= len(rows) {
		__antithesis_instrumentation__.Notify(573812)
		s.accountedFor.inputRowIdxToLookedUpRowIndices = s.accountedFor.inputRowIdxToLookedUpRowIndices[:len(rows)]
	} else {
		__antithesis_instrumentation__.Notify(573813)
		oldAccountedFor := s.accountedFor.inputRowIdxToLookedUpRowIndices
		s.accountedFor.inputRowIdxToLookedUpRowIndices = make([]int64, len(rows))

		copy(s.accountedFor.inputRowIdxToLookedUpRowIndices, oldAccountedFor[:cap(oldAccountedFor)])
	}
	__antithesis_instrumentation__.Notify(573806)

	sliceOverhead := memsize.IntSliceOverhead*int64(cap(s.inputRowIdxToLookedUpRowIndices)) +
		memsize.Int64*int64(cap(s.accountedFor.inputRowIdxToLookedUpRowIndices))
	if err := s.growMemoryAccount(sliceOverhead - s.accountedFor.sliceOverhead); err != nil {
		__antithesis_instrumentation__.Notify(573814)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(573815)
	}
	__antithesis_instrumentation__.Notify(573807)
	s.accountedFor.sliceOverhead = sliceOverhead

	s.inputRows = rows
	s.remoteSpansGenerated = false
	return s.generateSpans(s.Ctx, s.inputRows)
}

func (s *joinReaderOrderingStrategy) processLookedUpRow(
	ctx context.Context, row rowenc.EncDatumRow, key roachpb.Key,
) (joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573816)
	matchingInputRowIndices := s.getMatchingRowIndices(key)
	var containerIdx int
	if !s.isPartialJoin {
		__antithesis_instrumentation__.Notify(573821)

		for i := range row {
			__antithesis_instrumentation__.Notify(573823)
			if row[i].IsUnset() {
				__antithesis_instrumentation__.Notify(573824)
				row[i].Datum = tree.DNull
			} else {
				__antithesis_instrumentation__.Notify(573825)
			}
		}
		__antithesis_instrumentation__.Notify(573822)
		var err error
		containerIdx, err = s.lookedUpRows.AddRow(ctx, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(573826)
			return jrStateUnknown, err
		} else {
			__antithesis_instrumentation__.Notify(573827)
		}
	} else {
		__antithesis_instrumentation__.Notify(573828)
	}
	__antithesis_instrumentation__.Notify(573817)

	for _, inputRowIdx := range matchingInputRowIndices {
		__antithesis_instrumentation__.Notify(573829)
		if !s.isPartialJoin {
			__antithesis_instrumentation__.Notify(573831)
			s.inputRowIdxToLookedUpRowIndices[inputRowIdx] = append(
				s.inputRowIdxToLookedUpRowIndices[inputRowIdx], containerIdx)
			continue
		} else {
			__antithesis_instrumentation__.Notify(573832)
		}
		__antithesis_instrumentation__.Notify(573830)

		if !s.groupingState.getMatched(inputRowIdx) {
			__antithesis_instrumentation__.Notify(573833)
			renderedRow, err := s.render(s.inputRows[inputRowIdx], row)
			if err != nil {
				__antithesis_instrumentation__.Notify(573836)
				return jrStateUnknown, err
			} else {
				__antithesis_instrumentation__.Notify(573837)
			}
			__antithesis_instrumentation__.Notify(573834)
			if renderedRow == nil {
				__antithesis_instrumentation__.Notify(573838)

				continue
			} else {
				__antithesis_instrumentation__.Notify(573839)
			}
			__antithesis_instrumentation__.Notify(573835)
			s.groupingState.setMatched(inputRowIdx)
			s.inputRowIdxToLookedUpRowIndices[inputRowIdx] = partialJoinSentinel
		} else {
			__antithesis_instrumentation__.Notify(573840)
		}
	}
	__antithesis_instrumentation__.Notify(573818)

	var delta int64
	for _, idx := range matchingInputRowIndices {
		__antithesis_instrumentation__.Notify(573841)
		newSize := memsize.Int * int64(cap(s.inputRowIdxToLookedUpRowIndices[idx]))
		delta += newSize - s.accountedFor.inputRowIdxToLookedUpRowIndices[idx]
		s.accountedFor.inputRowIdxToLookedUpRowIndices[idx] = newSize
	}
	__antithesis_instrumentation__.Notify(573819)
	if err := s.growMemoryAccount(delta); err != nil {
		__antithesis_instrumentation__.Notify(573842)
		return jrStateUnknown, err
	} else {
		__antithesis_instrumentation__.Notify(573843)
	}
	__antithesis_instrumentation__.Notify(573820)

	return jrPerformingLookup, nil
}

func (s *joinReaderOrderingStrategy) prepareToEmit(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573844)
	if !s.isPartialJoin {
		__antithesis_instrumentation__.Notify(573845)
		s.lookedUpRows.SetupForRead(ctx, s.inputRowIdxToLookedUpRowIndices)
	} else {
		__antithesis_instrumentation__.Notify(573846)
	}
}

func (s *joinReaderOrderingStrategy) nextRowToEmit(
	ctx context.Context,
) (rowenc.EncDatumRow, joinReaderState, error) {
	__antithesis_instrumentation__.Notify(573847)
	if s.emitCursor.inputRowIdx >= len(s.inputRowIdxToLookedUpRowIndices) {
		__antithesis_instrumentation__.Notify(573854)
		log.VEventf(ctx, 1, "done emitting rows")

		s.emitCursor.outputRowIdx = 0
		s.emitCursor.inputRowIdx = 0
		if err := s.lookedUpRows.UnsafeReset(ctx); err != nil {
			__antithesis_instrumentation__.Notify(573856)
			return nil, jrStateUnknown, err
		} else {
			__antithesis_instrumentation__.Notify(573857)
		}
		__antithesis_instrumentation__.Notify(573855)
		return nil, jrReadingInput, nil
	} else {
		__antithesis_instrumentation__.Notify(573858)
	}
	__antithesis_instrumentation__.Notify(573848)

	inputRow := s.inputRows[s.emitCursor.inputRowIdx]
	lookedUpRows := s.inputRowIdxToLookedUpRowIndices[s.emitCursor.inputRowIdx]
	if s.emitCursor.outputRowIdx >= len(lookedUpRows) {
		__antithesis_instrumentation__.Notify(573859)

		inputRowIdx := s.emitCursor.inputRowIdx
		s.emitCursor.inputRowIdx++
		s.emitCursor.outputRowIdx = 0
		if s.groupingState.isUnmatched(inputRowIdx) {
			__antithesis_instrumentation__.Notify(573861)
			switch s.joinType {
			case descpb.LeftOuterJoin:
				__antithesis_instrumentation__.Notify(573862)

				if renderedRow := s.renderUnmatchedRow(inputRow, leftSide); renderedRow != nil {
					__antithesis_instrumentation__.Notify(573865)
					if s.outputGroupContinuationForLeftRow {
						__antithesis_instrumentation__.Notify(573867)

						renderedRow = append(renderedRow, falseEncDatum)
					} else {
						__antithesis_instrumentation__.Notify(573868)
					}
					__antithesis_instrumentation__.Notify(573866)
					return renderedRow, jrEmittingRows, nil
				} else {
					__antithesis_instrumentation__.Notify(573869)
				}
			case descpb.LeftAntiJoin:
				__antithesis_instrumentation__.Notify(573863)

				return inputRow, jrEmittingRows, nil
			default:
				__antithesis_instrumentation__.Notify(573864)
			}
		} else {
			__antithesis_instrumentation__.Notify(573870)
		}
		__antithesis_instrumentation__.Notify(573860)
		return nil, jrEmittingRows, nil
	} else {
		__antithesis_instrumentation__.Notify(573871)
	}
	__antithesis_instrumentation__.Notify(573849)

	lookedUpRowIdx := lookedUpRows[s.emitCursor.outputRowIdx]
	s.emitCursor.outputRowIdx++
	switch s.joinType {
	case descpb.LeftSemiJoin:
		__antithesis_instrumentation__.Notify(573872)

		return inputRow, jrEmittingRows, nil
	case descpb.LeftAntiJoin:
		__antithesis_instrumentation__.Notify(573873)

		return nil, jrEmittingRows, nil
	default:
		__antithesis_instrumentation__.Notify(573874)
	}
	__antithesis_instrumentation__.Notify(573850)

	lookedUpRow, err := s.lookedUpRows.GetRow(s.Ctx, lookedUpRowIdx, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(573875)
		return nil, jrStateUnknown, err
	} else {
		__antithesis_instrumentation__.Notify(573876)
	}
	__antithesis_instrumentation__.Notify(573851)
	outputRow, err := s.render(inputRow, lookedUpRow)
	if err != nil {
		__antithesis_instrumentation__.Notify(573877)
		return nil, jrStateUnknown, err
	} else {
		__antithesis_instrumentation__.Notify(573878)
	}
	__antithesis_instrumentation__.Notify(573852)
	if outputRow != nil {
		__antithesis_instrumentation__.Notify(573879)
		wasAlreadyMatched := s.groupingState.setMatched(s.emitCursor.inputRowIdx)
		if s.outputGroupContinuationForLeftRow {
			__antithesis_instrumentation__.Notify(573880)
			if wasAlreadyMatched {
				__antithesis_instrumentation__.Notify(573881)

				outputRow = append(outputRow, trueEncDatum)
			} else {
				__antithesis_instrumentation__.Notify(573882)

				outputRow = append(outputRow, falseEncDatum)
			}
		} else {
			__antithesis_instrumentation__.Notify(573883)
		}
	} else {
		__antithesis_instrumentation__.Notify(573884)
	}
	__antithesis_instrumentation__.Notify(573853)
	return outputRow, jrEmittingRows, nil
}

func (s *joinReaderOrderingStrategy) spilled() bool {
	__antithesis_instrumentation__.Notify(573885)
	if s.lookedUpRows != nil {
		__antithesis_instrumentation__.Notify(573887)
		return s.lookedUpRows.Spilled()
	} else {
		__antithesis_instrumentation__.Notify(573888)
	}
	__antithesis_instrumentation__.Notify(573886)

	return s.testingInfoSpilled
}

func (s *joinReaderOrderingStrategy) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(573889)
	s.memAcc.Close(ctx)
	s.joinReaderSpanGenerator.close(ctx)
	if s.lookedUpRows != nil {
		__antithesis_instrumentation__.Notify(573891)
		s.lookedUpRows.Close(ctx)
	} else {
		__antithesis_instrumentation__.Notify(573892)
	}
	__antithesis_instrumentation__.Notify(573890)
	*s = joinReaderOrderingStrategy{
		testingInfoSpilled: s.lookedUpRows.Spilled(),
	}
}

func (s *joinReaderOrderingStrategy) growMemoryAccount(delta int64) error {
	__antithesis_instrumentation__.Notify(573893)
	if err := s.memAcc.Grow(s.Ctx, delta); err != nil {
		__antithesis_instrumentation__.Notify(573895)

		spilled, spillErr := s.lookedUpRows.SpillToDisk(s.Ctx)
		if !spilled || func() bool {
			__antithesis_instrumentation__.Notify(573897)
			return spillErr != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(573898)
			return addWorkmemHint(errors.CombineErrors(err, spillErr))
		} else {
			__antithesis_instrumentation__.Notify(573899)
		}
		__antithesis_instrumentation__.Notify(573896)

		return addWorkmemHint(s.memAcc.Grow(s.Ctx, delta))
	} else {
		__antithesis_instrumentation__.Notify(573900)
	}
	__antithesis_instrumentation__.Notify(573894)
	return nil
}
