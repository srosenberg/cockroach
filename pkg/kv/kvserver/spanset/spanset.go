package spanset

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type SpanAccess int

const (
	SpanReadOnly SpanAccess = iota
	SpanReadWrite
	NumSpanAccess
)

func (a SpanAccess) String() string {
	__antithesis_instrumentation__.Notify(123233)
	switch a {
	case SpanReadOnly:
		__antithesis_instrumentation__.Notify(123234)
		return "read"
	case SpanReadWrite:
		__antithesis_instrumentation__.Notify(123235)
		return "write"
	default:
		__antithesis_instrumentation__.Notify(123236)
		panic("unreachable")
	}
}

type SpanScope int

const (
	SpanGlobal SpanScope = iota
	SpanLocal
	NumSpanScope
)

func (a SpanScope) String() string {
	__antithesis_instrumentation__.Notify(123237)
	switch a {
	case SpanGlobal:
		__antithesis_instrumentation__.Notify(123238)
		return "global"
	case SpanLocal:
		__antithesis_instrumentation__.Notify(123239)
		return "local"
	default:
		__antithesis_instrumentation__.Notify(123240)
		panic("unreachable")
	}
}

type Span struct {
	roachpb.Span
	Timestamp hlc.Timestamp
}

type SpanSet struct {
	spans [NumSpanAccess][NumSpanScope][]Span
}

var spanSetPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(123241); return new(SpanSet) },
}

func New() *SpanSet {
	__antithesis_instrumentation__.Notify(123242)
	return spanSetPool.Get().(*SpanSet)
}

func (s *SpanSet) Release() {
	__antithesis_instrumentation__.Notify(123243)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123245)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123246)

			const maxRecycleCap = 8
			var recycle []Span
			if sl := s.spans[sa][ss]; cap(sl) <= maxRecycleCap {
				__antithesis_instrumentation__.Notify(123248)
				for i := range sl {
					__antithesis_instrumentation__.Notify(123250)
					sl[i] = Span{}
				}
				__antithesis_instrumentation__.Notify(123249)
				recycle = sl[:0]
			} else {
				__antithesis_instrumentation__.Notify(123251)
			}
			__antithesis_instrumentation__.Notify(123247)
			s.spans[sa][ss] = recycle
		}
	}
	__antithesis_instrumentation__.Notify(123244)
	spanSetPool.Put(s)
}

func (s *SpanSet) String() string {
	__antithesis_instrumentation__.Notify(123252)
	var buf strings.Builder
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123254)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123255)
			for _, cur := range s.GetSpans(sa, ss) {
				__antithesis_instrumentation__.Notify(123256)
				fmt.Fprintf(&buf, "%s %s: %s at %s\n",
					sa, ss, cur.Span.String(), cur.Timestamp.String())
			}
		}
	}
	__antithesis_instrumentation__.Notify(123253)
	return buf.String()
}

func (s *SpanSet) Len() int {
	__antithesis_instrumentation__.Notify(123257)
	var count int
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123259)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123260)
			count += len(s.GetSpans(sa, ss))
		}
	}
	__antithesis_instrumentation__.Notify(123258)
	return count
}

func (s *SpanSet) Empty() bool {
	__antithesis_instrumentation__.Notify(123261)
	return s.Len() == 0
}

func (s *SpanSet) Copy() *SpanSet {
	__antithesis_instrumentation__.Notify(123262)
	n := New()
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123264)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123265)
			n.spans[sa][ss] = append(n.spans[sa][ss], s.spans[sa][ss]...)
		}
	}
	__antithesis_instrumentation__.Notify(123263)
	return n
}

func (s *SpanSet) Iterate(f func(SpanAccess, SpanScope, Span)) {
	__antithesis_instrumentation__.Notify(123266)
	if s == nil {
		__antithesis_instrumentation__.Notify(123268)
		return
	} else {
		__antithesis_instrumentation__.Notify(123269)
	}
	__antithesis_instrumentation__.Notify(123267)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123270)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123271)
			for _, span := range s.spans[sa][ss] {
				__antithesis_instrumentation__.Notify(123272)
				f(sa, ss, span)
			}
		}
	}
}

func (s *SpanSet) Reserve(access SpanAccess, scope SpanScope, n int) {
	__antithesis_instrumentation__.Notify(123273)
	existing := s.spans[access][scope]
	if n <= cap(existing)-len(existing) {
		__antithesis_instrumentation__.Notify(123275)
		return
	} else {
		__antithesis_instrumentation__.Notify(123276)
	}
	__antithesis_instrumentation__.Notify(123274)
	s.spans[access][scope] = make([]Span, len(existing), n+len(existing))
	copy(s.spans[access][scope], existing)
}

func (s *SpanSet) AddNonMVCC(access SpanAccess, span roachpb.Span) {
	__antithesis_instrumentation__.Notify(123277)
	s.AddMVCC(access, span, hlc.Timestamp{})
}

func (s *SpanSet) AddMVCC(access SpanAccess, span roachpb.Span, timestamp hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(123278)
	scope := SpanGlobal
	if keys.IsLocal(span.Key) {
		__antithesis_instrumentation__.Notify(123280)
		scope = SpanLocal
		timestamp = hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(123281)
	}
	__antithesis_instrumentation__.Notify(123279)

	s.spans[access][scope] = append(s.spans[access][scope], Span{Span: span, Timestamp: timestamp})
}

func (s *SpanSet) Merge(s2 *SpanSet) {
	__antithesis_instrumentation__.Notify(123282)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123284)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123285)
			s.spans[sa][ss] = append(s.spans[sa][ss], s2.spans[sa][ss]...)
		}
	}
	__antithesis_instrumentation__.Notify(123283)
	s.SortAndDedup()
}

func (s *SpanSet) SortAndDedup() {
	__antithesis_instrumentation__.Notify(123286)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123287)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123288)
			s.spans[sa][ss], _ = mergeSpans(&s.spans[sa][ss])
		}
	}
}

func (s *SpanSet) GetSpans(access SpanAccess, scope SpanScope) []Span {
	__antithesis_instrumentation__.Notify(123289)
	return s.spans[access][scope]
}

func (s *SpanSet) BoundarySpan(scope SpanScope) roachpb.Span {
	__antithesis_instrumentation__.Notify(123290)
	var boundary roachpb.Span
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123292)
		for _, cur := range s.GetSpans(sa, scope) {
			__antithesis_instrumentation__.Notify(123293)
			if !boundary.Valid() {
				__antithesis_instrumentation__.Notify(123295)
				boundary = cur.Span
				continue
			} else {
				__antithesis_instrumentation__.Notify(123296)
			}
			__antithesis_instrumentation__.Notify(123294)
			boundary = boundary.Combine(cur.Span)
		}
	}
	__antithesis_instrumentation__.Notify(123291)
	return boundary
}

func (s *SpanSet) Intersects(other *SpanSet) bool {
	__antithesis_instrumentation__.Notify(123297)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123299)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123300)
			otherSpans := other.GetSpans(sa, ss)
			for _, span := range otherSpans {
				__antithesis_instrumentation__.Notify(123301)

				if err := s.CheckAllowed(sa, span.Span); err == nil {
					__antithesis_instrumentation__.Notify(123302)
					return true
				} else {
					__antithesis_instrumentation__.Notify(123303)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(123298)
	return false
}

func (s *SpanSet) AssertAllowed(access SpanAccess, span roachpb.Span) {
	__antithesis_instrumentation__.Notify(123304)
	if err := s.CheckAllowed(access, span); err != nil {
		__antithesis_instrumentation__.Notify(123305)
		log.Fatalf(context.TODO(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(123306)
	}
}

func (s *SpanSet) CheckAllowed(access SpanAccess, span roachpb.Span) error {
	__antithesis_instrumentation__.Notify(123307)
	return s.checkAllowed(access, span, func(_ SpanAccess, _ Span) bool {
		__antithesis_instrumentation__.Notify(123308)
		return true
	})
}

func (s *SpanSet) CheckAllowedAt(
	access SpanAccess, span roachpb.Span, timestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(123309)
	mvcc := !timestamp.IsEmpty()
	return s.checkAllowed(access, span, func(declAccess SpanAccess, declSpan Span) bool {
		__antithesis_instrumentation__.Notify(123310)
		declTimestamp := declSpan.Timestamp
		if declTimestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(123312)

			return true
		} else {
			__antithesis_instrumentation__.Notify(123313)
		}
		__antithesis_instrumentation__.Notify(123311)

		switch declAccess {
		case SpanReadOnly:
			__antithesis_instrumentation__.Notify(123314)
			switch access {
			case SpanReadOnly:
				__antithesis_instrumentation__.Notify(123317)

				return mvcc && func() bool {
					__antithesis_instrumentation__.Notify(123320)
					return timestamp.LessEq(declTimestamp) == true
				}() == true
			case SpanReadWrite:
				__antithesis_instrumentation__.Notify(123318)

				panic("unexpected SpanReadWrite access")
			default:
				__antithesis_instrumentation__.Notify(123319)
				panic("unexpected span access")
			}
		case SpanReadWrite:
			__antithesis_instrumentation__.Notify(123315)
			switch access {
			case SpanReadOnly:
				__antithesis_instrumentation__.Notify(123321)

				return mvcc
			case SpanReadWrite:
				__antithesis_instrumentation__.Notify(123322)

				return mvcc && func() bool {
					__antithesis_instrumentation__.Notify(123324)
					return declTimestamp.LessEq(timestamp) == true
				}() == true
			default:
				__antithesis_instrumentation__.Notify(123323)
				panic("unexpected span access")
			}
		default:
			__antithesis_instrumentation__.Notify(123316)
			panic("unexpected span access")
		}
	})
}

func (s *SpanSet) checkAllowed(
	access SpanAccess, span roachpb.Span, check func(SpanAccess, Span) bool,
) error {
	__antithesis_instrumentation__.Notify(123325)
	scope := SpanGlobal
	if (span.Key != nil && func() bool {
		__antithesis_instrumentation__.Notify(123328)
		return keys.IsLocal(span.Key) == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(123329)
		return (span.EndKey != nil && func() bool {
			__antithesis_instrumentation__.Notify(123330)
			return keys.IsLocal(span.EndKey) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(123331)
		scope = SpanLocal
	} else {
		__antithesis_instrumentation__.Notify(123332)
	}
	__antithesis_instrumentation__.Notify(123326)

	for ac := access; ac < NumSpanAccess; ac++ {
		__antithesis_instrumentation__.Notify(123333)
		for _, cur := range s.spans[ac][scope] {
			__antithesis_instrumentation__.Notify(123334)
			if contains(cur.Span, span) && func() bool {
				__antithesis_instrumentation__.Notify(123335)
				return check(ac, cur) == true
			}() == true {
				__antithesis_instrumentation__.Notify(123336)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(123337)
			}
		}
	}
	__antithesis_instrumentation__.Notify(123327)

	return errors.Errorf("cannot %s undeclared span %s\ndeclared:\n%s\nstack:\n%s", access, span, s, debug.Stack())
}

func contains(s1, s2 roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(123338)
	if s2.Key != nil {
		__antithesis_instrumentation__.Notify(123341)

		return s1.Contains(s2)
	} else {
		__antithesis_instrumentation__.Notify(123342)
	}
	__antithesis_instrumentation__.Notify(123339)

	if s1.EndKey == nil {
		__antithesis_instrumentation__.Notify(123343)
		return s1.Key.IsPrev(s2.EndKey)
	} else {
		__antithesis_instrumentation__.Notify(123344)
	}
	__antithesis_instrumentation__.Notify(123340)

	return s1.Key.Compare(s2.EndKey) < 0 && func() bool {
		__antithesis_instrumentation__.Notify(123345)
		return s1.EndKey.Compare(s2.EndKey) >= 0 == true
	}() == true
}

func (s *SpanSet) Validate() error {
	__antithesis_instrumentation__.Notify(123346)
	for sa := SpanAccess(0); sa < NumSpanAccess; sa++ {
		__antithesis_instrumentation__.Notify(123348)
		for ss := SpanScope(0); ss < NumSpanScope; ss++ {
			__antithesis_instrumentation__.Notify(123349)
			for _, cur := range s.GetSpans(sa, ss) {
				__antithesis_instrumentation__.Notify(123350)
				if len(cur.EndKey) > 0 && func() bool {
					__antithesis_instrumentation__.Notify(123351)
					return cur.Key.Compare(cur.EndKey) >= 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(123352)
					return errors.Errorf("inverted span %s %s", cur.Key, cur.EndKey)
				} else {
					__antithesis_instrumentation__.Notify(123353)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(123347)

	return nil
}
