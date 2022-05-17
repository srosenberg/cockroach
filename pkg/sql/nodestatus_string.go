package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(501841)

	var x [1]struct{}
	_ = x[NodeOK-0]
	_ = x[NodeUnhealthy-1]
	_ = x[NodeDistSQLVersionIncompatible-2]
}

const _NodeStatus_name = "NodeOKNodeUnhealthyNodeDistSQLVersionIncompatible"

var _NodeStatus_index = [...]uint8{0, 6, 19, 49}

func (i NodeStatus) String() string {
	__antithesis_instrumentation__.Notify(501842)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(501844)
		return i >= NodeStatus(len(_NodeStatus_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(501845)
		return "NodeStatus(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(501846)
	}
	__antithesis_instrumentation__.Notify(501843)
	return _NodeStatus_name[_NodeStatus_index[i]:_NodeStatus_index[i+1]]
}
