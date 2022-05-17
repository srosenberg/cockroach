package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(629003)

	var x [1]struct{}
	_ = x[noEvent-0]
	_ = x[txnStart-1]
	_ = x[txnCommit-2]
	_ = x[txnRollback-3]
	_ = x[txnRestart-4]
}

const _txnEventType_name = "noEventtxnStarttxnCommittxnRollbacktxnRestart"

var _txnEventType_index = [...]uint8{0, 7, 15, 24, 35, 45}

func (i txnEventType) String() string {
	__antithesis_instrumentation__.Notify(629004)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(629006)
		return i >= txnEventType(len(_txnEventType_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(629007)
		return "txnEventType(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(629008)
	}
	__antithesis_instrumentation__.Notify(629005)
	return _txnEventType_name[_txnEventType_index[i]:_txnEventType_index[i+1]]
}
