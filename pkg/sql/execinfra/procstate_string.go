package execinfra

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(471379)

	var x [1]struct{}
	_ = x[StateRunning-0]
	_ = x[StateDraining-1]
	_ = x[StateTrailingMeta-2]
	_ = x[StateExhausted-3]
}

const _procState_name = "StateRunningStateDrainingStateTrailingMetaStateExhausted"

var _procState_index = [...]uint8{0, 12, 25, 42, 56}

func (i procState) String() string {
	__antithesis_instrumentation__.Notify(471380)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(471382)
		return i >= procState(len(_procState_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(471383)
		return "procState(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(471384)
	}
	__antithesis_instrumentation__.Notify(471381)
	return _procState_name[_procState_index[i]:_procState_index[i+1]]
}
