package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/rangecache"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const readerOverflowProtection = 1000000000000000

func LimitHint(specLimitHint int64, post *execinfrapb.PostProcessSpec) (limitHint int64) {
	__antithesis_instrumentation__.Notify(471385)

	if post.Limit != 0 && func() bool {
		__antithesis_instrumentation__.Notify(471387)
		return post.Limit <= readerOverflowProtection == true
	}() == true {
		__antithesis_instrumentation__.Notify(471388)
		limitHint = int64(post.Limit)
	} else {
		__antithesis_instrumentation__.Notify(471389)
		if specLimitHint != 0 && func() bool {
			__antithesis_instrumentation__.Notify(471390)
			return specLimitHint <= readerOverflowProtection == true
		}() == true {
			__antithesis_instrumentation__.Notify(471391)
			limitHint = specLimitHint
		} else {
			__antithesis_instrumentation__.Notify(471392)
		}
	}
	__antithesis_instrumentation__.Notify(471386)

	return limitHint
}

func MisplannedRanges(
	ctx context.Context, spans []roachpb.Span, nodeID roachpb.NodeID, rdc *rangecache.RangeCache,
) (misplannedRanges []roachpb.RangeInfo) {
	__antithesis_instrumentation__.Notify(471393)
	log.VEvent(ctx, 2, "checking range cache to see if range info updates should be communicated to the gateway")
	var misplanned map[roachpb.RangeID]struct{}
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(471396)
		rSpan, err := keys.SpanAddr(sp)
		if err != nil {
			__antithesis_instrumentation__.Notify(471398)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(471399)
		}
		__antithesis_instrumentation__.Notify(471397)
		overlapping := rdc.GetCachedOverlapping(ctx, rSpan)

		for _, ri := range overlapping {
			__antithesis_instrumentation__.Notify(471400)
			if _, ok := misplanned[ri.Desc().RangeID]; ok {
				__antithesis_instrumentation__.Notify(471402)

				continue
			} else {
				__antithesis_instrumentation__.Notify(471403)
			}
			__antithesis_instrumentation__.Notify(471401)

			l := ri.Lease()
			if l != nil && func() bool {
				__antithesis_instrumentation__.Notify(471404)
				return l.Replica.NodeID != nodeID == true
			}() == true {
				__antithesis_instrumentation__.Notify(471405)
				misplannedRanges = append(misplannedRanges, roachpb.RangeInfo{
					Desc:                  *ri.Desc(),
					Lease:                 *l,
					ClosedTimestampPolicy: ri.ClosedTimestampPolicy(),
				})

				if misplanned == nil {
					__antithesis_instrumentation__.Notify(471407)
					misplanned = make(map[roachpb.RangeID]struct{})
				} else {
					__antithesis_instrumentation__.Notify(471408)
				}
				__antithesis_instrumentation__.Notify(471406)
				misplanned[ri.Desc().RangeID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(471409)
			}
		}
	}
	__antithesis_instrumentation__.Notify(471394)

	if len(misplannedRanges) != 0 && func() bool {
		__antithesis_instrumentation__.Notify(471410)
		return log.ExpensiveLogEnabled(ctx, 2) == true
	}() == true {
		__antithesis_instrumentation__.Notify(471411)
		var b strings.Builder
		for i := range misplannedRanges {
			__antithesis_instrumentation__.Notify(471413)
			if i > 0 {
				__antithesis_instrumentation__.Notify(471416)
				b.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(471417)
			}
			__antithesis_instrumentation__.Notify(471414)
			if i > 3 {
				__antithesis_instrumentation__.Notify(471418)
				b.WriteString("...")
				break
			} else {
				__antithesis_instrumentation__.Notify(471419)
			}
			__antithesis_instrumentation__.Notify(471415)
			fmt.Fprintf(&b, "%+v", misplannedRanges[i])
		}
		__antithesis_instrumentation__.Notify(471412)
		log.VEventf(ctx, 2, "misplanned ranges: %s", b.String())
	} else {
		__antithesis_instrumentation__.Notify(471420)
	}
	__antithesis_instrumentation__.Notify(471395)

	return misplannedRanges
}

type SpansWithCopy struct {
	Spans     roachpb.Spans
	SpansCopy roachpb.Spans
}

func (s *SpansWithCopy) MakeSpansCopy() {
	__antithesis_instrumentation__.Notify(471421)
	if cap(s.SpansCopy) >= len(s.Spans) {
		__antithesis_instrumentation__.Notify(471423)
		s.SpansCopy = s.SpansCopy[:len(s.Spans)]
	} else {
		__antithesis_instrumentation__.Notify(471424)
		s.SpansCopy = make(roachpb.Spans, len(s.Spans))
	}
	__antithesis_instrumentation__.Notify(471422)
	copy(s.SpansCopy, s.Spans)
}

func (s *SpansWithCopy) Reset() {
	__antithesis_instrumentation__.Notify(471425)
	for i := range s.SpansCopy {
		__antithesis_instrumentation__.Notify(471427)
		s.SpansCopy[i] = roachpb.Span{}
	}
	__antithesis_instrumentation__.Notify(471426)
	s.Spans = nil
	s.SpansCopy = s.SpansCopy[:0]
}

type limitHintBatchCount int

const (
	limitHintFirstBatch limitHintBatchCount = iota
	limitHintSecondBatch
	limitHintDisabled
)

const limitHintSecondBatchFactor = 10

type LimitHintHelper struct {
	origLimitHint int64

	currentLimitHint int64
	limitHintIdx     limitHintBatchCount
}

func MakeLimitHintHelper(specLimitHint int64, post *execinfrapb.PostProcessSpec) LimitHintHelper {
	__antithesis_instrumentation__.Notify(471428)
	limitHint := LimitHint(specLimitHint, post)
	return LimitHintHelper{
		origLimitHint:    limitHint,
		currentLimitHint: limitHint,
		limitHintIdx:     limitHintFirstBatch,
	}
}

func (h *LimitHintHelper) LimitHint() int64 {
	__antithesis_instrumentation__.Notify(471429)
	return h.currentLimitHint
}

func (h *LimitHintHelper) ReadSomeRows(rowsRead int64) error {
	__antithesis_instrumentation__.Notify(471430)
	if h.currentLimitHint != 0 {
		__antithesis_instrumentation__.Notify(471432)
		h.currentLimitHint -= rowsRead
		if h.currentLimitHint == 0 {
			__antithesis_instrumentation__.Notify(471433)

			switch h.limitHintIdx {
			case limitHintFirstBatch:
				__antithesis_instrumentation__.Notify(471434)
				h.currentLimitHint = limitHintSecondBatchFactor * h.origLimitHint
				h.limitHintIdx = limitHintSecondBatch
			default:
				__antithesis_instrumentation__.Notify(471435)
				h.currentLimitHint = 0
				h.limitHintIdx = limitHintDisabled
			}
		} else {
			__antithesis_instrumentation__.Notify(471436)
			if h.currentLimitHint < 0 {
				__antithesis_instrumentation__.Notify(471437)
				return errors.AssertionFailedf(
					"unexpectedly the user of LimitHintHelper read " +
						"more rows that the current limit hint",
				)
			} else {
				__antithesis_instrumentation__.Notify(471438)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(471439)
	}
	__antithesis_instrumentation__.Notify(471431)
	return nil
}
