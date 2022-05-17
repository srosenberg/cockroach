package cyclegraphtest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(578560)

	var x [1]struct{}
	_ = x[s-0]
	_ = x[s1-1]
	_ = x[s2-2]
	_ = x[c-3]
	_ = x[name-4]
}

const _testAttr_name = "ss1s2cname"

var _testAttr_index = [...]uint8{0, 1, 3, 5, 6, 10}

func (i testAttr) String() string {
	__antithesis_instrumentation__.Notify(578561)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(578563)
		return i >= testAttr(len(_testAttr_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(578564)
		return "testAttr(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(578565)
	}
	__antithesis_instrumentation__.Notify(578562)
	return _testAttr_name[_testAttr_index[i]:_testAttr_index[i+1]]
}
