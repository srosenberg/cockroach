package scop

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(582479)

	var x [1]struct{}
	_ = x[MutationType-1]
	_ = x[BackfillType-2]
	_ = x[ValidationType-3]
}

const _Type_name = "MutationTypeBackfillTypeValidationType"

var _Type_index = [...]uint8{0, 12, 24, 38}

func (i Type) String() string {
	__antithesis_instrumentation__.Notify(582480)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(582482)
		return i >= Type(len(_Type_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(582483)
		return "Type(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(582484)
	}
	__antithesis_instrumentation__.Notify(582481)
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
