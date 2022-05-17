package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type condensableSpanSet struct {
	s     []roachpb.Span
	bytes int64

	condensed bool

	sAlloc [2]roachpb.Span
}

func (s *condensableSpanSet) insert(spans ...roachpb.Span) {
	__antithesis_instrumentation__.Notify(86978)
	if cap(s.s) == 0 {
		__antithesis_instrumentation__.Notify(86980)
		s.s = s.sAlloc[:0]
	} else {
		__antithesis_instrumentation__.Notify(86981)
	}
	__antithesis_instrumentation__.Notify(86979)
	s.s = append(s.s, spans...)
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(86982)
		s.bytes += spanSize(sp)
	}
}

func (s *condensableSpanSet) mergeAndSort() {
	__antithesis_instrumentation__.Notify(86983)
	oldLen := len(s.s)
	s.s, _ = roachpb.MergeSpans(&s.s)

	if oldLen != len(s.s) {
		__antithesis_instrumentation__.Notify(86984)
		s.bytes = 0
		for _, sp := range s.s {
			__antithesis_instrumentation__.Notify(86985)
			s.bytes += spanSize(sp)
		}
	} else {
		__antithesis_instrumentation__.Notify(86986)
	}
}

func (s *condensableSpanSet) maybeCondense(
	ctx context.Context, riGen rangeIteratorFactory, maxBytes int64,
) bool {
	__antithesis_instrumentation__.Notify(86987)
	if s.bytes <= maxBytes {
		__antithesis_instrumentation__.Notify(86993)
		return false
	} else {
		__antithesis_instrumentation__.Notify(86994)
	}
	__antithesis_instrumentation__.Notify(86988)

	s.mergeAndSort()
	if s.bytes <= maxBytes {
		__antithesis_instrumentation__.Notify(86995)
		return false
	} else {
		__antithesis_instrumentation__.Notify(86996)
	}
	__antithesis_instrumentation__.Notify(86989)

	ri := riGen.newRangeIterator()

	type spanBucket struct {
		rangeID roachpb.RangeID
		bytes   int64
		spans   []roachpb.Span
	}
	var buckets []spanBucket
	var localSpans []roachpb.Span
	for _, sp := range s.s {
		__antithesis_instrumentation__.Notify(86997)
		if keys.IsLocal(sp.Key) {
			__antithesis_instrumentation__.Notify(87001)
			localSpans = append(localSpans, sp)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87002)
		}
		__antithesis_instrumentation__.Notify(86998)
		ri.Seek(ctx, roachpb.RKey(sp.Key), Ascending)
		if !ri.Valid() {
			__antithesis_instrumentation__.Notify(87003)

			log.Warningf(ctx, "failed to condense lock spans: %v", ri.Error())
			return false
		} else {
			__antithesis_instrumentation__.Notify(87004)
		}
		__antithesis_instrumentation__.Notify(86999)
		rangeID := ri.Desc().RangeID
		if l := len(buckets); l > 0 && func() bool {
			__antithesis_instrumentation__.Notify(87005)
			return buckets[l-1].rangeID == rangeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(87006)
			buckets[l-1].spans = append(buckets[l-1].spans, sp)
		} else {
			__antithesis_instrumentation__.Notify(87007)
			buckets = append(buckets, spanBucket{
				rangeID: rangeID, spans: []roachpb.Span{sp},
			})
		}
		__antithesis_instrumentation__.Notify(87000)
		buckets[len(buckets)-1].bytes += spanSize(sp)
	}
	__antithesis_instrumentation__.Notify(86990)

	sort.Slice(buckets, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(87008)
		return buckets[i].bytes > buckets[j].bytes
	})
	__antithesis_instrumentation__.Notify(86991)
	s.s = localSpans
	for _, bucket := range buckets {
		__antithesis_instrumentation__.Notify(87009)

		if s.bytes <= maxBytes/2 {
			__antithesis_instrumentation__.Notify(87012)

			s.s = append(s.s, bucket.spans...)
			continue
		} else {
			__antithesis_instrumentation__.Notify(87013)
		}
		__antithesis_instrumentation__.Notify(87010)
		s.bytes -= bucket.bytes

		cs := bucket.spans[0]
		for _, s := range bucket.spans[1:] {
			__antithesis_instrumentation__.Notify(87014)
			cs = cs.Combine(s)
			if !cs.Valid() {
				__antithesis_instrumentation__.Notify(87015)

				log.Fatalf(ctx, "failed to condense lock spans: "+
					"combining span %s yielded invalid result", s)
			} else {
				__antithesis_instrumentation__.Notify(87016)
			}
		}
		__antithesis_instrumentation__.Notify(87011)
		s.bytes += spanSize(cs)
		s.s = append(s.s, cs)
	}
	__antithesis_instrumentation__.Notify(86992)
	s.condensed = true
	return true
}

func (s *condensableSpanSet) asSlice() []roachpb.Span {
	__antithesis_instrumentation__.Notify(87017)
	l := len(s.s)
	return s.s[:l:l]
}

func (s *condensableSpanSet) empty() bool {
	__antithesis_instrumentation__.Notify(87018)
	return len(s.s) == 0
}

func (s *condensableSpanSet) clear() {
	__antithesis_instrumentation__.Notify(87019)
	*s = condensableSpanSet{}
}

func (s *condensableSpanSet) estimateSize(spans []roachpb.Span, mergeThresholdBytes int64) int64 {
	__antithesis_instrumentation__.Notify(87020)
	var bytes int64
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(87025)
		bytes += spanSize(sp)
	}
	{
		__antithesis_instrumentation__.Notify(87026)
		estimate := s.bytes + bytes
		if estimate <= mergeThresholdBytes {
			__antithesis_instrumentation__.Notify(87027)
			return estimate
		} else {
			__antithesis_instrumentation__.Notify(87028)
		}
	}
	__antithesis_instrumentation__.Notify(87021)

	s.mergeAndSort()

	estimate := s.bytes + bytes
	if estimate <= mergeThresholdBytes {
		__antithesis_instrumentation__.Notify(87029)
		return estimate
	} else {
		__antithesis_instrumentation__.Notify(87030)
	}
	__antithesis_instrumentation__.Notify(87022)

	spans = append(spans, s.s...)
	lenBeforeMerge := len(spans)
	spans, _ = roachpb.MergeSpans(&spans)
	if len(spans) == lenBeforeMerge {
		__antithesis_instrumentation__.Notify(87031)

		return estimate
	} else {
		__antithesis_instrumentation__.Notify(87032)
	}
	__antithesis_instrumentation__.Notify(87023)

	bytes = 0
	for _, sp := range spans {
		__antithesis_instrumentation__.Notify(87033)
		bytes += spanSize(sp)
	}
	__antithesis_instrumentation__.Notify(87024)
	return bytes
}

func (s *condensableSpanSet) bytesSize() int64 {
	__antithesis_instrumentation__.Notify(87034)
	return s.bytes
}

func spanSize(sp roachpb.Span) int64 {
	__antithesis_instrumentation__.Notify(87035)
	return int64(len(sp.Key) + len(sp.EndKey))
}

func keySize(k roachpb.Key) int64 {
	__antithesis_instrumentation__.Notify(87036)
	return int64(len(k))
}
