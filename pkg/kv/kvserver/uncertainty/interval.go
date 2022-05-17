package uncertainty

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/hlc"

type Interval struct {
	GlobalLimit hlc.Timestamp
	LocalLimit  hlc.ClockTimestamp
}

func (in Interval) IsUncertain(valueTs hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(127553)
	if !in.LocalLimit.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(127555)
		return !valueTs.Synthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(127556)
		return valueTs.LessEq(in.LocalLimit.ToTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(127557)
	}
	__antithesis_instrumentation__.Notify(127554)
	return valueTs.LessEq(in.GlobalLimit)
}
