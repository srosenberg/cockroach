package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(114572)

	var x [1]struct{}
	_ = x[noReason-0]
	_ = x[reasonNewLeader-1]
	_ = x[reasonNewLeaderOrConfigChange-2]
	_ = x[reasonSnapshotApplied-3]
	_ = x[reasonTicks-4]
}

const _refreshRaftReason_name = "noReasonreasonNewLeaderreasonNewLeaderOrConfigChangereasonSnapshotAppliedreasonTicks"

var _refreshRaftReason_index = [...]uint8{0, 8, 23, 52, 73, 84}

func (i refreshRaftReason) String() string {
	__antithesis_instrumentation__.Notify(114573)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(114575)
		return i >= refreshRaftReason(len(_refreshRaftReason_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(114576)
		return "refreshRaftReason(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(114577)
	}
	__antithesis_instrumentation__.Notify(114574)
	return _refreshRaftReason_name[_refreshRaftReason_index[i]:_refreshRaftReason_index[i+1]]
}
