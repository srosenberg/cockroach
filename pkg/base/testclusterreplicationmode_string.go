package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(1763)

	var x [1]struct{}
	_ = x[ReplicationAuto-0]
	_ = x[ReplicationManual-1]
}

const _TestClusterReplicationMode_name = "ReplicationAutoReplicationManual"

var _TestClusterReplicationMode_index = [...]uint8{0, 15, 32}

func (i TestClusterReplicationMode) String() string {
	__antithesis_instrumentation__.Notify(1764)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(1766)
		return i >= TestClusterReplicationMode(len(_TestClusterReplicationMode_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(1767)
		return "TestClusterReplicationMode(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(1768)
	}
	__antithesis_instrumentation__.Notify(1765)
	return _TestClusterReplicationMode_name[_TestClusterReplicationMode_index[i]:_TestClusterReplicationMode_index[i+1]]
}
