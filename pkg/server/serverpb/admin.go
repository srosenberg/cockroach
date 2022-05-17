package serverpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (ts *TableStatsResponse) Add(ots *TableStatsResponse) {
	__antithesis_instrumentation__.Notify(196179)
	ts.RangeCount += ots.RangeCount
	ts.ReplicaCount += ots.ReplicaCount
	ts.ApproximateDiskBytes += ots.ApproximateDiskBytes
	ts.Stats.Add(ots.Stats)

	missingNodeIds := make(map[string]struct{})
	for _, nodeData := range ts.MissingNodes {
		__antithesis_instrumentation__.Notify(196181)
		missingNodeIds[nodeData.NodeID] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(196180)
	for _, nodeData := range ots.MissingNodes {
		__antithesis_instrumentation__.Notify(196182)
		if _, found := missingNodeIds[nodeData.NodeID]; !found {
			__antithesis_instrumentation__.Notify(196183)
			ts.MissingNodes = append(ts.MissingNodes, nodeData)
			ts.NodeCount--
		} else {
			__antithesis_instrumentation__.Notify(196184)
		}
	}
}
