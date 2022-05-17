package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
)

type Filter struct {
	needPrevVals interval.RangeGroup
	needVals     interval.RangeGroup
}

func newFilterFromRegistry(reg *registry) *Filter {
	__antithesis_instrumentation__.Notify(113613)
	f := &Filter{
		needPrevVals: interval.NewRangeList(),
		needVals:     interval.NewRangeList(),
	}
	reg.tree.Do(func(i interval.Interface) (done bool) {
		__antithesis_instrumentation__.Notify(113615)
		r := i.(*registration)
		if r.withDiff {
			__antithesis_instrumentation__.Notify(113617)
			f.needPrevVals.Add(r.Range())
		} else {
			__antithesis_instrumentation__.Notify(113618)
		}
		__antithesis_instrumentation__.Notify(113616)
		f.needVals.Add(r.Range())
		return false
	})
	__antithesis_instrumentation__.Notify(113614)
	return f
}

func (r *Filter) NeedPrevVal(s roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(113619)
	return r.needPrevVals.Overlaps(s.AsRange())
}

func (r *Filter) NeedVal(s roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(113620)
	return r.needVals.Overlaps(s.AsRange())
}
