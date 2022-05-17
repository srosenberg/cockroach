package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(605373)

	var x [1]struct{}
	_ = x[Enum-1]
	_ = x[Composite-2]
	_ = x[Range-3]
	_ = x[Base-4]
	_ = x[Shell-5]
	_ = x[Domain-6]
}

const _CreateTypeVariety_name = "EnumCompositeRangeBaseShellDomain"

var _CreateTypeVariety_index = [...]uint8{0, 4, 13, 18, 22, 27, 33}

func (i CreateTypeVariety) String() string {
	__antithesis_instrumentation__.Notify(605374)
	i -= 1
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(605376)
		return i >= CreateTypeVariety(len(_CreateTypeVariety_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(605377)
		return "CreateTypeVariety(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(605378)
	}
	__antithesis_instrumentation__.Notify(605375)
	return _CreateTypeVariety_name[_CreateTypeVariety_index[i]:_CreateTypeVariety_index[i+1]]
}
