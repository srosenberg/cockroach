package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(613692)

	var x [1]struct{}
	_ = x[TypeDDL-0]
	_ = x[TypeDML-1]
	_ = x[TypeDCL-2]
	_ = x[TypeTCL-3]
}

const _StatementType_name = "TypeDDLTypeDMLTypeDCLTypeTCL"

var _StatementType_index = [...]uint8{0, 7, 14, 21, 28}

func (i StatementType) String() string {
	__antithesis_instrumentation__.Notify(613693)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(613695)
		return i >= StatementType(len(_StatementType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(613696)
		return "StatementType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(613697)
	}
	__antithesis_instrumentation__.Notify(613694)
	return _StatementType_name[_StatementType_index[i]:_StatementType_index[i+1]]
}
