package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(643662)

	var x [1]struct{}
	_ = x[ResourceLimitNotReached-0]
	_ = x[ResourceLimitReachedSoft-1]
	_ = x[ResourceLimitReachedHard-2]
}

const _ResourceLimitReached_name = "ResourceLimitNotReachedResourceLimitReachedSoftResourceLimitReachedHard"

var _ResourceLimitReached_index = [...]uint8{0, 23, 47, 71}

func (i ResourceLimitReached) String() string {
	__antithesis_instrumentation__.Notify(643663)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(643665)
		return i >= ResourceLimitReached(len(_ResourceLimitReached_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(643666)
		return "ResourceLimitReached(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(643667)
	}
	__antithesis_instrumentation__.Notify(643664)
	return _ResourceLimitReached_name[_ResourceLimitReached_index[i]:_ResourceLimitReached_index[i+1]]
}
