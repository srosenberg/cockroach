package rel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(579262)

	var x [1]struct{}
	_ = x[Type-63]
	_ = x[Self-62]
}

const _systemAttribute_name = "SelfType"

var _systemAttribute_index = [...]uint8{0, 4, 8}

func (i systemAttribute) String() string {
	__antithesis_instrumentation__.Notify(579263)
	i -= 62
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(579265)
		return i >= systemAttribute(len(_systemAttribute_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(579266)
		return "systemAttribute(" + strconv.FormatInt(int64(i+62), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(579267)
	}
	__antithesis_instrumentation__.Notify(579264)
	return _systemAttribute_name[_systemAttribute_index[i]:_systemAttribute_index[i+1]]
}
