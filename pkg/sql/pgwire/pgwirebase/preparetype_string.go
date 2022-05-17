package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(561401)

	var x [1]struct{}
	_ = x[PrepareStatement-83]
	_ = x[PreparePortal-80]
}

const (
	_PrepareType_name_0 = "PreparePortal"
	_PrepareType_name_1 = "PrepareStatement"
)

func (i PrepareType) String() string {
	__antithesis_instrumentation__.Notify(561402)
	switch {
	case i == 80:
		__antithesis_instrumentation__.Notify(561403)
		return _PrepareType_name_0
	case i == 83:
		__antithesis_instrumentation__.Notify(561404)
		return _PrepareType_name_1
	default:
		__antithesis_instrumentation__.Notify(561405)
		return "PrepareType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
