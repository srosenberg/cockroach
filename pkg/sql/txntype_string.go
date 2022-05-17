package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(629009)

	var x [1]struct{}
	_ = x[implicitTxn-0]
	_ = x[explicitTxn-1]
}

const _txnType_name = "implicitTxnexplicitTxn"

var _txnType_index = [...]uint8{0, 11, 22}

func (i txnType) String() string {
	__antithesis_instrumentation__.Notify(629010)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(629012)
		return i >= txnType(len(_txnType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(629013)
		return "txnType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(629014)
	}
	__antithesis_instrumentation__.Notify(629011)
	return _txnType_name[_txnType_index[i]:_txnType_index[i+1]]
}
