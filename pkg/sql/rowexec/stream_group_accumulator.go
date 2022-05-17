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
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type streamGroupAccumulator struct {
	src   execinfra.RowSource
	types []*types.T

	srcConsumed bool
	ordering    colinfo.ColumnOrdering

	curGroup   []rowenc.EncDatumRow
	datumAlloc tree.DatumAlloc

	leftoverRow rowenc.EncDatumRow

	rowAlloc rowenc.EncDatumRowAlloc

	memAcc mon.BoundAccount
}

func makeStreamGroupAccumulator(
	src execinfra.RowSource, ordering colinfo.ColumnOrdering, memMonitor *mon.BytesMonitor,
) streamGroupAccumulator {
	__antithesis_instrumentation__.Notify(575021)
	return streamGroupAccumulator{
		src:      src,
		types:    src.OutputTypes(),
		ordering: ordering,
		memAcc:   memMonitor.MakeBoundAccount(),
	}
}

func (s *streamGroupAccumulator) start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575022)
	s.src.Start(ctx)
}

func (s *streamGroupAccumulator) nextGroup(
	ctx context.Context, evalCtx *tree.EvalContext,
) ([]rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575023)
	if s.srcConsumed {
		__antithesis_instrumentation__.Notify(575026)

		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(575027)
	}
	__antithesis_instrumentation__.Notify(575024)

	if s.leftoverRow != nil {
		__antithesis_instrumentation__.Notify(575028)
		s.curGroup = append(s.curGroup, s.leftoverRow)
		s.leftoverRow = nil
	} else {
		__antithesis_instrumentation__.Notify(575029)
	}
	__antithesis_instrumentation__.Notify(575025)

	for {
		__antithesis_instrumentation__.Notify(575030)
		row, meta := s.src.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(575036)
			return nil, meta
		} else {
			__antithesis_instrumentation__.Notify(575037)
		}
		__antithesis_instrumentation__.Notify(575031)
		if row == nil {
			__antithesis_instrumentation__.Notify(575038)
			s.srcConsumed = true
			return s.curGroup, nil
		} else {
			__antithesis_instrumentation__.Notify(575039)
		}
		__antithesis_instrumentation__.Notify(575032)

		if err := s.memAcc.Grow(ctx, int64(row.Size())); err != nil {
			__antithesis_instrumentation__.Notify(575040)
			return nil, &execinfrapb.ProducerMetadata{Err: err}
		} else {
			__antithesis_instrumentation__.Notify(575041)
		}
		__antithesis_instrumentation__.Notify(575033)
		row = s.rowAlloc.CopyRow(row)

		if len(s.curGroup) == 0 {
			__antithesis_instrumentation__.Notify(575042)
			if s.curGroup == nil {
				__antithesis_instrumentation__.Notify(575044)
				s.curGroup = make([]rowenc.EncDatumRow, 0, 64)
			} else {
				__antithesis_instrumentation__.Notify(575045)
			}
			__antithesis_instrumentation__.Notify(575043)
			s.curGroup = append(s.curGroup, row)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575046)
		}
		__antithesis_instrumentation__.Notify(575034)

		cmp, err := s.curGroup[0].Compare(s.types, &s.datumAlloc, s.ordering, evalCtx, row)
		if err != nil {
			__antithesis_instrumentation__.Notify(575047)
			return nil, &execinfrapb.ProducerMetadata{Err: err}
		} else {
			__antithesis_instrumentation__.Notify(575048)
		}
		__antithesis_instrumentation__.Notify(575035)
		if cmp == 0 {
			__antithesis_instrumentation__.Notify(575049)
			s.curGroup = append(s.curGroup, row)
		} else {
			__antithesis_instrumentation__.Notify(575050)
			if cmp == 1 {
				__antithesis_instrumentation__.Notify(575051)
				return nil, &execinfrapb.ProducerMetadata{
					Err: errors.Errorf(
						"detected badly ordered input: %s > %s, but expected '<'",
						s.curGroup[0].String(s.types), row.String(s.types)),
				}
			} else {
				__antithesis_instrumentation__.Notify(575052)
				n := len(s.curGroup)
				ret := s.curGroup[:n:n]
				s.curGroup = s.curGroup[:0]
				s.memAcc.Empty(ctx)
				s.leftoverRow = row
				return ret, nil
			}
		}
	}
}

func (s *streamGroupAccumulator) close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575053)
	s.memAcc.Close(ctx)
}
