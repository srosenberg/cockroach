package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(242298)

	var x [1]struct{}
	_ = x[advanceUnknown-0]
	_ = x[stayInPlace-1]
	_ = x[advanceOne-2]
	_ = x[skipBatch-3]
	_ = x[rewind-4]
}

const _advanceCode_name = "advanceUnknownstayInPlaceadvanceOneskipBatchrewind"

var _advanceCode_index = [...]uint8{0, 14, 25, 35, 44, 50}

func (i advanceCode) String() string {
	__antithesis_instrumentation__.Notify(242299)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(242301)
		return i >= advanceCode(len(_advanceCode_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(242302)
		return "advanceCode(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(242303)
	}
	__antithesis_instrumentation__.Notify(242300)
	return _advanceCode_name[_advanceCode_index[i]:_advanceCode_index[i+1]]
}
