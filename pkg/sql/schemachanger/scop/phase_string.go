package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(582473)

	var x [1]struct{}
	_ = x[StatementPhase-1]
	_ = x[PreCommitPhase-2]
	_ = x[PostCommitPhase-3]
	_ = x[PostCommitNonRevertiblePhase-4]
}

const _Phase_name = "StatementPhasePreCommitPhasePostCommitPhasePostCommitNonRevertiblePhase"

var _Phase_index = [...]uint8{0, 14, 28, 43, 71}

func (i Phase) String() string {
	__antithesis_instrumentation__.Notify(582474)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(582476)
		return i >= Phase(len(_Phase_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(582477)
		return "Phase(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(582478)
	}
	__antithesis_instrumentation__.Notify(582475)
	return _Phase_name[_Phase_index[i]:_Phase_index[i+1]]
}
