package rowexec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type streamMerger struct {
	left       streamGroupAccumulator
	right      streamGroupAccumulator
	leftGroup  []rowenc.EncDatumRow
	rightGroup []rowenc.EncDatumRow

	nullEquality bool
	datumAlloc   tree.DatumAlloc
}

func (sm *streamMerger) start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575054)
	sm.left.start(ctx)
	sm.right.start(ctx)
}

func (sm *streamMerger) NextBatch(
	ctx context.Context, evalCtx *tree.EvalContext,
) ([]rowenc.EncDatumRow, []rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575055)
	if sm.leftGroup == nil {
		__antithesis_instrumentation__.Notify(575064)
		var meta *execinfrapb.ProducerMetadata
		sm.leftGroup, meta = sm.left.nextGroup(ctx, evalCtx)
		if meta != nil {
			__antithesis_instrumentation__.Notify(575065)
			return nil, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(575066)
		}
	} else {
		__antithesis_instrumentation__.Notify(575067)
	}
	__antithesis_instrumentation__.Notify(575056)
	if sm.rightGroup == nil {
		__antithesis_instrumentation__.Notify(575068)
		var meta *execinfrapb.ProducerMetadata
		sm.rightGroup, meta = sm.right.nextGroup(ctx, evalCtx)
		if meta != nil {
			__antithesis_instrumentation__.Notify(575069)
			return nil, nil, meta
		} else {
			__antithesis_instrumentation__.Notify(575070)
		}
	} else {
		__antithesis_instrumentation__.Notify(575071)
	}
	__antithesis_instrumentation__.Notify(575057)
	if sm.leftGroup == nil && func() bool {
		__antithesis_instrumentation__.Notify(575072)
		return sm.rightGroup == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(575073)
		return nil, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(575074)
	}
	__antithesis_instrumentation__.Notify(575058)

	var lrow, rrow rowenc.EncDatumRow
	if len(sm.leftGroup) > 0 {
		__antithesis_instrumentation__.Notify(575075)
		lrow = sm.leftGroup[0]
	} else {
		__antithesis_instrumentation__.Notify(575076)
	}
	__antithesis_instrumentation__.Notify(575059)
	if len(sm.rightGroup) > 0 {
		__antithesis_instrumentation__.Notify(575077)
		rrow = sm.rightGroup[0]
	} else {
		__antithesis_instrumentation__.Notify(575078)
	}
	__antithesis_instrumentation__.Notify(575060)

	cmp, err := CompareEncDatumRowForMerge(
		sm.left.types, lrow, rrow, sm.left.ordering, sm.right.ordering,
		sm.nullEquality, &sm.datumAlloc, evalCtx,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(575079)
		return nil, nil, &execinfrapb.ProducerMetadata{Err: err}
	} else {
		__antithesis_instrumentation__.Notify(575080)
	}
	__antithesis_instrumentation__.Notify(575061)
	var leftGroup, rightGroup []rowenc.EncDatumRow
	if cmp <= 0 {
		__antithesis_instrumentation__.Notify(575081)
		leftGroup = sm.leftGroup
		sm.leftGroup = nil
	} else {
		__antithesis_instrumentation__.Notify(575082)
	}
	__antithesis_instrumentation__.Notify(575062)
	if cmp >= 0 {
		__antithesis_instrumentation__.Notify(575083)
		rightGroup = sm.rightGroup
		sm.rightGroup = nil
	} else {
		__antithesis_instrumentation__.Notify(575084)
	}
	__antithesis_instrumentation__.Notify(575063)
	return leftGroup, rightGroup, nil
}

func CompareEncDatumRowForMerge(
	lhsTypes []*types.T,
	lhs, rhs rowenc.EncDatumRow,
	leftOrdering, rightOrdering colinfo.ColumnOrdering,
	nullEquality bool,
	da *tree.DatumAlloc,
	evalCtx *tree.EvalContext,
) (int, error) {
	__antithesis_instrumentation__.Notify(575085)
	if lhs == nil && func() bool {
		__antithesis_instrumentation__.Notify(575091)
		return rhs == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(575092)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(575093)
	}
	__antithesis_instrumentation__.Notify(575086)
	if lhs == nil {
		__antithesis_instrumentation__.Notify(575094)
		return 1, nil
	} else {
		__antithesis_instrumentation__.Notify(575095)
	}
	__antithesis_instrumentation__.Notify(575087)
	if rhs == nil {
		__antithesis_instrumentation__.Notify(575096)
		return -1, nil
	} else {
		__antithesis_instrumentation__.Notify(575097)
	}
	__antithesis_instrumentation__.Notify(575088)
	if len(leftOrdering) != len(rightOrdering) {
		__antithesis_instrumentation__.Notify(575098)
		return 0, errors.Errorf(
			"cannot compare two EncDatumRow types that have different length ColumnOrderings",
		)
	} else {
		__antithesis_instrumentation__.Notify(575099)
	}
	__antithesis_instrumentation__.Notify(575089)

	for i, ord := range leftOrdering {
		__antithesis_instrumentation__.Notify(575100)
		lIdx := ord.ColIdx
		rIdx := rightOrdering[i].ColIdx

		if lhs[lIdx].IsNull() && func() bool {
			__antithesis_instrumentation__.Notify(575103)
			return rhs[rIdx].IsNull() == true
		}() == true {
			__antithesis_instrumentation__.Notify(575104)
			if !nullEquality {
				__antithesis_instrumentation__.Notify(575106)

				return -1, nil
			} else {
				__antithesis_instrumentation__.Notify(575107)
			}
			__antithesis_instrumentation__.Notify(575105)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575108)
		}
		__antithesis_instrumentation__.Notify(575101)
		cmp, err := lhs[lIdx].Compare(lhsTypes[lIdx], da, evalCtx, &rhs[rIdx])
		if err != nil {
			__antithesis_instrumentation__.Notify(575109)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(575110)
		}
		__antithesis_instrumentation__.Notify(575102)
		if cmp != 0 {
			__antithesis_instrumentation__.Notify(575111)
			if leftOrdering[i].Direction == encoding.Descending {
				__antithesis_instrumentation__.Notify(575113)
				cmp = -cmp
			} else {
				__antithesis_instrumentation__.Notify(575114)
			}
			__antithesis_instrumentation__.Notify(575112)
			return cmp, nil
		} else {
			__antithesis_instrumentation__.Notify(575115)
		}
	}
	__antithesis_instrumentation__.Notify(575090)
	return 0, nil
}

func (sm *streamMerger) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575116)
	sm.left.close(ctx)
	sm.right.close(ctx)
}

func makeStreamMerger(
	leftSource execinfra.RowSource,
	leftOrdering colinfo.ColumnOrdering,
	rightSource execinfra.RowSource,
	rightOrdering colinfo.ColumnOrdering,
	nullEquality bool,
	memMonitor *mon.BytesMonitor,
) (streamMerger, error) {
	__antithesis_instrumentation__.Notify(575117)
	if len(leftOrdering) != len(rightOrdering) {
		__antithesis_instrumentation__.Notify(575120)
		return streamMerger{}, errors.Errorf(
			"ordering lengths don't match: %d and %d", len(leftOrdering), len(rightOrdering))
	} else {
		__antithesis_instrumentation__.Notify(575121)
	}
	__antithesis_instrumentation__.Notify(575118)
	for i, ord := range leftOrdering {
		__antithesis_instrumentation__.Notify(575122)
		if ord.Direction != rightOrdering[i].Direction {
			__antithesis_instrumentation__.Notify(575123)
			return streamMerger{}, errors.New("ordering mismatch")
		} else {
			__antithesis_instrumentation__.Notify(575124)
		}
	}
	__antithesis_instrumentation__.Notify(575119)

	return streamMerger{
		left:         makeStreamGroupAccumulator(leftSource, leftOrdering, memMonitor),
		right:        makeStreamGroupAccumulator(rightSource, rightOrdering, memMonitor),
		nullEquality: nullEquality,
	}, nil
}
