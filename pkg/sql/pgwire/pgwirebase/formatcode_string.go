package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(561381)

	var x [1]struct{}
	_ = x[FormatText-0]
	_ = x[FormatBinary-1]
}

const _FormatCode_name = "FormatTextFormatBinary"

var _FormatCode_index = [...]uint8{0, 10, 22}

func (i FormatCode) String() string {
	__antithesis_instrumentation__.Notify(561382)
	if i >= FormatCode(len(_FormatCode_index)-1) {
		__antithesis_instrumentation__.Notify(561384)
		return "FormatCode(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(561385)
	}
	__antithesis_instrumentation__.Notify(561383)
	return _FormatCode_name[_FormatCode_index[i]:_FormatCode_index[i+1]]
}
