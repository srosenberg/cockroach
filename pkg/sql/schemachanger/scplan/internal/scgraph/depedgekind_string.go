package scgraph

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(594235)

	var x [1]struct{}
	_ = x[Precedence-1]
	_ = x[SameStagePrecedence-2]
}

const _DepEdgeKind_name = "PrecedenceSameStagePrecedence"

var _DepEdgeKind_index = [...]uint8{0, 10, 29}

func (i DepEdgeKind) String() string {
	__antithesis_instrumentation__.Notify(594236)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(594238)
		return i >= DepEdgeKind(len(_DepEdgeKind_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(594239)
		return "DepEdgeKind(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(594240)
	}
	__antithesis_instrumentation__.Notify(594237)
	return _DepEdgeKind_name[_DepEdgeKind_index[i]:_DepEdgeKind_index[i+1]]
}
