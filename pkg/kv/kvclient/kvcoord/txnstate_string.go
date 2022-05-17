package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(89245)

	var x [1]struct{}
	_ = x[txnPending-0]
	_ = x[txnRetryableError-1]
	_ = x[txnError-2]
	_ = x[txnFinalized-3]
}

const _txnState_name = "txnPendingtxnRetryableErrortxnErrortxnFinalized"

var _txnState_index = [...]uint8{0, 10, 27, 35, 47}

func (i txnState) String() string {
	__antithesis_instrumentation__.Notify(89246)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(89248)
		return i >= txnState(len(_txnState_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(89249)
		return "txnState(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(89250)
	}
	__antithesis_instrumentation__.Notify(89247)
	return _txnState_name[_txnState_index[i]:_txnState_index[i+1]]
}
