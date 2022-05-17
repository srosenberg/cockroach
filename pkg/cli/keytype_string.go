package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(33251)

	var x [1]struct{}
	_ = x[raw-0]
	_ = x[human-1]
	_ = x[rangeID-2]
	_ = x[hex-3]
}

const _keyType_name = "rawhumanrangeIDhex"

var _keyType_index = [...]uint8{0, 3, 8, 15, 18}

func (i keyType) String() string {
	__antithesis_instrumentation__.Notify(33252)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(33254)
		return i >= keyType(len(_keyType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(33255)
		return "keyType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(33256)
	}
	__antithesis_instrumentation__.Notify(33253)
	return _keyType_name[_keyType_index[i]:_keyType_index[i+1]]
}
