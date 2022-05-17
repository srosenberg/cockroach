package entitynodetest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(578573)

	var x [1]struct{}
	_ = x[i8-0]
	_ = x[pi8-1]
	_ = x[i16-2]
	_ = x[value-3]
	_ = x[left-4]
	_ = x[right-5]
}

const _testAttr_name = "i8pi8i16valueleftright"

var _testAttr_index = [...]uint8{0, 2, 5, 8, 13, 17, 22}

func (i testAttr) String() string {
	__antithesis_instrumentation__.Notify(578574)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(578576)
		return i >= testAttr(len(_testAttr_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(578577)
		return "testAttr(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(578578)
	}
	__antithesis_instrumentation__.Notify(578575)
	return _testAttr_name[_testAttr_index[i]:_testAttr_index[i+1]]
}
