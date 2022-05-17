package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(613686)

	var x [1]struct{}
	_ = x[Ack-0]
	_ = x[DDL-1]
	_ = x[RowsAffected-2]
	_ = x[Rows-3]
	_ = x[CopyIn-4]
	_ = x[Unknown-5]
}

const _StatementReturnType_name = "AckDDLRowsAffectedRowsCopyInUnknown"

var _StatementReturnType_index = [...]uint8{0, 3, 6, 18, 22, 28, 35}

func (i StatementReturnType) String() string {
	__antithesis_instrumentation__.Notify(613687)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(613689)
		return i >= StatementReturnType(len(_StatementReturnType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(613690)
		return "StatementReturnType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(613691)
	}
	__antithesis_instrumentation__.Notify(613688)
	return _StatementReturnType_name[_StatementReturnType_index[i]:_StatementReturnType_index[i+1]]
}
