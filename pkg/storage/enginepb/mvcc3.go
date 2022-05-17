package enginepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

func (MVCCStatsDelta) SafeValue() { __antithesis_instrumentation__.Notify(635594) }

func (ms *MVCCStatsDelta) ToStats() MVCCStats {
	__antithesis_instrumentation__.Notify(635595)
	return MVCCStats(*ms)
}

func (ms *MVCCStats) ToStatsDelta() MVCCStatsDelta {
	__antithesis_instrumentation__.Notify(635596)
	return MVCCStatsDelta(*ms)
}

func (ms *MVCCPersistentStats) ToStats() MVCCStats {
	__antithesis_instrumentation__.Notify(635597)
	return MVCCStats(*ms)
}

func (ms *MVCCPersistentStats) ToStatsPtr() *MVCCStats {
	__antithesis_instrumentation__.Notify(635598)
	return (*MVCCStats)(ms)
}

func (ms *MVCCStats) SafeValue() { __antithesis_instrumentation__.Notify(635599) }

func (ms *MVCCStats) ToPersistentStats() MVCCPersistentStats {
	__antithesis_instrumentation__.Notify(635600)
	return MVCCPersistentStats(*ms)
}

func (op *MVCCLogicalOp) MustSetValue(value interface{}) {
	__antithesis_instrumentation__.Notify(635601)
	op.Reset()
	if !op.SetValue(value) {
		__antithesis_instrumentation__.Notify(635602)
		panic(errors.AssertionFailedf("%T excludes %T", op, value))
	} else {
		__antithesis_instrumentation__.Notify(635603)
	}
}
