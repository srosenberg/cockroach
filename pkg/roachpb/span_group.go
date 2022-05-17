package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/interval"

type SpanGroup struct {
	rg interval.RangeGroup
}

func (g *SpanGroup) checkInit() {
	__antithesis_instrumentation__.Notify(179844)
	if g.rg == nil {
		__antithesis_instrumentation__.Notify(179845)
		g.rg = interval.NewRangeTree()
	} else {
		__antithesis_instrumentation__.Notify(179846)
	}
}

func (g *SpanGroup) Add(spans ...Span) bool {
	__antithesis_instrumentation__.Notify(179847)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(179850)
		return false
	} else {
		__antithesis_instrumentation__.Notify(179851)
	}
	__antithesis_instrumentation__.Notify(179848)
	ret := false
	g.checkInit()
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(179852)
		ret = g.rg.Add(s2r(span)) || func() bool {
			__antithesis_instrumentation__.Notify(179853)
			return ret == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(179849)
	return ret
}

func (g *SpanGroup) Sub(spans ...Span) bool {
	__antithesis_instrumentation__.Notify(179854)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(179857)
		return false
	} else {
		__antithesis_instrumentation__.Notify(179858)
	}
	__antithesis_instrumentation__.Notify(179855)
	ret := false
	g.checkInit()
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(179859)
		ret = g.rg.Sub(s2r(span)) || func() bool {
			__antithesis_instrumentation__.Notify(179860)
			return ret == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(179856)
	return ret
}

func (g *SpanGroup) Contains(k Key) bool {
	__antithesis_instrumentation__.Notify(179861)
	if g.rg == nil {
		__antithesis_instrumentation__.Notify(179863)
		return false
	} else {
		__antithesis_instrumentation__.Notify(179864)
	}
	__antithesis_instrumentation__.Notify(179862)
	return g.rg.Encloses(interval.Range{
		Start: interval.Comparable(k),

		End: interval.Comparable(k.Next()),
	})
}

func (g *SpanGroup) Encloses(spans ...Span) bool {
	__antithesis_instrumentation__.Notify(179865)
	if g.rg == nil {
		__antithesis_instrumentation__.Notify(179868)
		return false
	} else {
		__antithesis_instrumentation__.Notify(179869)
	}
	__antithesis_instrumentation__.Notify(179866)
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(179870)
		if !g.rg.Encloses(s2r(span)) {
			__antithesis_instrumentation__.Notify(179871)
			return false
		} else {
			__antithesis_instrumentation__.Notify(179872)
		}
	}
	__antithesis_instrumentation__.Notify(179867)
	return true
}

func (g *SpanGroup) Len() int {
	__antithesis_instrumentation__.Notify(179873)
	if g.rg == nil {
		__antithesis_instrumentation__.Notify(179875)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(179876)
	}
	__antithesis_instrumentation__.Notify(179874)
	return g.rg.Len()
}

var _ = (*SpanGroup).Len

func (g *SpanGroup) Slice() []Span {
	__antithesis_instrumentation__.Notify(179877)
	rg := g.rg
	if rg == nil {
		__antithesis_instrumentation__.Notify(179880)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(179881)
	}
	__antithesis_instrumentation__.Notify(179878)
	ret := make([]Span, 0, rg.Len())
	it := rg.Iterator()
	for {
		__antithesis_instrumentation__.Notify(179882)
		rng, next := it.Next()
		if !next {
			__antithesis_instrumentation__.Notify(179884)
			break
		} else {
			__antithesis_instrumentation__.Notify(179885)
		}
		__antithesis_instrumentation__.Notify(179883)
		ret = append(ret, r2s(rng))
	}
	__antithesis_instrumentation__.Notify(179879)
	return ret
}

func s2r(s Span) interval.Range {
	__antithesis_instrumentation__.Notify(179886)

	var end = s.EndKey
	if len(end) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(179888)
		return s.Key.Equal(s.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(179889)
		end = s.Key.Next()
	} else {
		__antithesis_instrumentation__.Notify(179890)
	}
	__antithesis_instrumentation__.Notify(179887)
	return interval.Range{
		Start: interval.Comparable(s.Key),
		End:   interval.Comparable(end),
	}
}

func r2s(r interval.Range) Span {
	__antithesis_instrumentation__.Notify(179891)
	return Span{
		Key:    Key(r.Start),
		EndKey: Key(r.End),
	}
}
