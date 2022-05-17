package execinfrapb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/optional"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/dustin/go-humanize"
	"github.com/gogo/protobuf/types"
)

func ProcessorComponentID(
	instanceID base.SQLInstanceID, flowID FlowID, processorID int32,
) ComponentID {
	__antithesis_instrumentation__.Notify(472665)
	return ComponentID{
		FlowID:        flowID,
		Type:          ComponentID_PROCESSOR,
		ID:            processorID,
		SQLInstanceID: instanceID,
	}
}

func StreamComponentID(
	originInstanceID base.SQLInstanceID, flowID FlowID, streamID StreamID,
) ComponentID {
	__antithesis_instrumentation__.Notify(472666)
	return ComponentID{
		FlowID:        flowID,
		Type:          ComponentID_STREAM,
		ID:            int32(streamID),
		SQLInstanceID: originInstanceID,
	}
}

func FlowComponentID(instanceID base.SQLInstanceID, flowID FlowID) ComponentID {
	__antithesis_instrumentation__.Notify(472667)
	return ComponentID{
		FlowID:        flowID,
		Type:          ComponentID_FLOW,
		SQLInstanceID: instanceID,
	}
}

const (
	FlowIDTagKey = tracing.TagPrefix + "flowid"

	StreamIDTagKey = tracing.TagPrefix + "streamid"

	ProcessorIDTagKey = tracing.TagPrefix + "processorid"
)

func (s *ComponentStats) StatsForQueryPlan() []string {
	__antithesis_instrumentation__.Notify(472668)
	result := make([]string, 0, 4)
	s.formatStats(func(key string, value interface{}) {
		__antithesis_instrumentation__.Notify(472670)
		result = append(result, fmt.Sprintf("%s: %v", key, value))
	})
	__antithesis_instrumentation__.Notify(472669)
	return result
}

func (s *ComponentStats) formatStats(fn func(suffix string, value interface{})) {
	__antithesis_instrumentation__.Notify(472671)

	if s.NetRx.Latency.HasValue() {
		__antithesis_instrumentation__.Notify(472692)
		fn("network latency", humanizeutil.Duration(s.NetRx.Latency.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472693)
	}
	__antithesis_instrumentation__.Notify(472672)
	if s.NetRx.WaitTime.HasValue() {
		__antithesis_instrumentation__.Notify(472694)
		fn("network wait time", humanizeutil.Duration(s.NetRx.WaitTime.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472695)
	}
	__antithesis_instrumentation__.Notify(472673)
	if s.NetRx.DeserializationTime.HasValue() {
		__antithesis_instrumentation__.Notify(472696)
		fn("deserialization time", humanizeutil.Duration(s.NetRx.DeserializationTime.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472697)
	}
	__antithesis_instrumentation__.Notify(472674)
	if s.NetRx.TuplesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472698)
		fn("network rows received", humanizeutil.Count(s.NetRx.TuplesReceived.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472699)
	}
	__antithesis_instrumentation__.Notify(472675)
	if s.NetRx.BytesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472700)
		fn("network bytes received", humanize.IBytes(s.NetRx.BytesReceived.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472701)
	}
	__antithesis_instrumentation__.Notify(472676)
	if s.NetRx.MessagesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472702)
		fn("network messages received", humanizeutil.Count(s.NetRx.MessagesReceived.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472703)
	}
	__antithesis_instrumentation__.Notify(472677)

	if s.NetTx.TuplesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472704)
		fn("network rows sent", humanizeutil.Count(s.NetTx.TuplesSent.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472705)
	}
	__antithesis_instrumentation__.Notify(472678)
	if s.NetTx.BytesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472706)
		fn("network bytes sent", humanize.IBytes(s.NetTx.BytesSent.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472707)
	}
	__antithesis_instrumentation__.Notify(472679)
	if s.NetTx.MessagesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472708)
		fn("network messages sent", humanizeutil.Count(s.NetTx.MessagesSent.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472709)
	}
	__antithesis_instrumentation__.Notify(472680)

	switch len(s.Inputs) {
	case 1:
		__antithesis_instrumentation__.Notify(472710)
		if s.Inputs[0].NumTuples.HasValue() {
			__antithesis_instrumentation__.Notify(472717)
			fn("input rows", humanizeutil.Count(s.Inputs[0].NumTuples.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472718)
		}
		__antithesis_instrumentation__.Notify(472711)
		if s.Inputs[0].WaitTime.HasValue() {
			__antithesis_instrumentation__.Notify(472719)
			fn("input stall time", humanizeutil.Duration(s.Inputs[0].WaitTime.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472720)
		}

	case 2:
		__antithesis_instrumentation__.Notify(472712)
		if s.Inputs[0].NumTuples.HasValue() {
			__antithesis_instrumentation__.Notify(472721)
			fn("left rows", humanizeutil.Count(s.Inputs[0].NumTuples.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472722)
		}
		__antithesis_instrumentation__.Notify(472713)
		if s.Inputs[0].WaitTime.HasValue() {
			__antithesis_instrumentation__.Notify(472723)
			fn("left stall time", humanizeutil.Duration(s.Inputs[0].WaitTime.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472724)
		}
		__antithesis_instrumentation__.Notify(472714)
		if s.Inputs[1].NumTuples.HasValue() {
			__antithesis_instrumentation__.Notify(472725)
			fn("right rows", humanizeutil.Count(s.Inputs[1].NumTuples.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472726)
		}
		__antithesis_instrumentation__.Notify(472715)
		if s.Inputs[1].WaitTime.HasValue() {
			__antithesis_instrumentation__.Notify(472727)
			fn("right stall time", humanizeutil.Duration(s.Inputs[1].WaitTime.Value()))
		} else {
			__antithesis_instrumentation__.Notify(472728)
		}
	default:
		__antithesis_instrumentation__.Notify(472716)
	}
	__antithesis_instrumentation__.Notify(472681)

	if s.KV.KVTime.HasValue() {
		__antithesis_instrumentation__.Notify(472729)
		fn("KV time", humanizeutil.Duration(s.KV.KVTime.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472730)
	}
	__antithesis_instrumentation__.Notify(472682)
	if s.KV.ContentionTime.HasValue() {
		__antithesis_instrumentation__.Notify(472731)
		fn("KV contention time", humanizeutil.Duration(s.KV.ContentionTime.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472732)
	}
	__antithesis_instrumentation__.Notify(472683)
	if s.KV.TuplesRead.HasValue() {
		__antithesis_instrumentation__.Notify(472733)
		fn("KV rows read", humanizeutil.Count(s.KV.TuplesRead.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472734)
	}
	__antithesis_instrumentation__.Notify(472684)
	if s.KV.BytesRead.HasValue() {
		__antithesis_instrumentation__.Notify(472735)
		fn("KV bytes read", humanize.IBytes(s.KV.BytesRead.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472736)
	}
	__antithesis_instrumentation__.Notify(472685)
	if s.KV.NumInterfaceSteps.HasValue() {
		__antithesis_instrumentation__.Notify(472737)
		fn("MVCC step count (ext/int)",
			fmt.Sprintf("%s/%s",
				humanizeutil.Count(s.KV.NumInterfaceSteps.Value()),
				humanizeutil.Count(s.KV.NumInternalSteps.Value())),
		)
	} else {
		__antithesis_instrumentation__.Notify(472738)
	}
	__antithesis_instrumentation__.Notify(472686)
	if s.KV.NumInterfaceSeeks.HasValue() {
		__antithesis_instrumentation__.Notify(472739)
		fn("MVCC seek count (ext/int)",
			fmt.Sprintf("%s/%s",
				humanizeutil.Count(s.KV.NumInterfaceSeeks.Value()),
				humanizeutil.Count(s.KV.NumInternalSeeks.Value())),
		)
	} else {
		__antithesis_instrumentation__.Notify(472740)
	}
	__antithesis_instrumentation__.Notify(472687)

	if s.Exec.ExecTime.HasValue() {
		__antithesis_instrumentation__.Notify(472741)
		fn("execution time", humanizeutil.Duration(s.Exec.ExecTime.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472742)
	}
	__antithesis_instrumentation__.Notify(472688)
	if s.Exec.MaxAllocatedMem.HasValue() {
		__antithesis_instrumentation__.Notify(472743)
		fn("max memory allocated", humanize.IBytes(s.Exec.MaxAllocatedMem.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472744)
	}
	__antithesis_instrumentation__.Notify(472689)
	if s.Exec.MaxAllocatedDisk.HasValue() {
		__antithesis_instrumentation__.Notify(472745)
		fn("max sql temp disk usage", humanize.IBytes(s.Exec.MaxAllocatedDisk.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472746)
	}
	__antithesis_instrumentation__.Notify(472690)

	if s.Output.NumBatches.HasValue() {
		__antithesis_instrumentation__.Notify(472747)
		fn("batches output", humanizeutil.Count(s.Output.NumBatches.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472748)
	}
	__antithesis_instrumentation__.Notify(472691)
	if s.Output.NumTuples.HasValue() {
		__antithesis_instrumentation__.Notify(472749)
		fn("rows output", humanizeutil.Count(s.Output.NumTuples.Value()))
	} else {
		__antithesis_instrumentation__.Notify(472750)
	}
}

func (s *ComponentStats) Union(other *ComponentStats) *ComponentStats {
	__antithesis_instrumentation__.Notify(472751)
	result := *s

	if !result.NetRx.Latency.HasValue() {
		__antithesis_instrumentation__.Notify(472775)
		result.NetRx.Latency = other.NetRx.Latency
	} else {
		__antithesis_instrumentation__.Notify(472776)
	}
	__antithesis_instrumentation__.Notify(472752)
	if !result.NetRx.WaitTime.HasValue() {
		__antithesis_instrumentation__.Notify(472777)
		result.NetRx.WaitTime = other.NetRx.WaitTime
	} else {
		__antithesis_instrumentation__.Notify(472778)
	}
	__antithesis_instrumentation__.Notify(472753)
	if !result.NetRx.DeserializationTime.HasValue() {
		__antithesis_instrumentation__.Notify(472779)
		result.NetRx.DeserializationTime = other.NetRx.DeserializationTime
	} else {
		__antithesis_instrumentation__.Notify(472780)
	}
	__antithesis_instrumentation__.Notify(472754)
	if !result.NetRx.TuplesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472781)
		result.NetRx.TuplesReceived = other.NetRx.TuplesReceived
	} else {
		__antithesis_instrumentation__.Notify(472782)
	}
	__antithesis_instrumentation__.Notify(472755)
	if !result.NetRx.BytesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472783)
		result.NetRx.BytesReceived = other.NetRx.BytesReceived
	} else {
		__antithesis_instrumentation__.Notify(472784)
	}
	__antithesis_instrumentation__.Notify(472756)
	if !result.NetRx.MessagesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472785)
		result.NetRx.MessagesReceived = other.NetRx.MessagesReceived
	} else {
		__antithesis_instrumentation__.Notify(472786)
	}
	__antithesis_instrumentation__.Notify(472757)

	if !result.NetTx.TuplesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472787)
		result.NetTx.TuplesSent = other.NetTx.TuplesSent
	} else {
		__antithesis_instrumentation__.Notify(472788)
	}
	__antithesis_instrumentation__.Notify(472758)
	if !result.NetTx.BytesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472789)
		result.NetTx.BytesSent = other.NetTx.BytesSent
	} else {
		__antithesis_instrumentation__.Notify(472790)
	}
	__antithesis_instrumentation__.Notify(472759)

	result.Inputs = append([]InputStats(nil), s.Inputs...)
	result.Inputs = append(result.Inputs, other.Inputs...)

	if !result.KV.KVTime.HasValue() {
		__antithesis_instrumentation__.Notify(472791)
		result.KV.KVTime = other.KV.KVTime
	} else {
		__antithesis_instrumentation__.Notify(472792)
	}
	__antithesis_instrumentation__.Notify(472760)
	if !result.KV.ContentionTime.HasValue() {
		__antithesis_instrumentation__.Notify(472793)
		result.KV.ContentionTime = other.KV.ContentionTime
	} else {
		__antithesis_instrumentation__.Notify(472794)
	}
	__antithesis_instrumentation__.Notify(472761)
	if !result.KV.NumInterfaceSteps.HasValue() {
		__antithesis_instrumentation__.Notify(472795)
		result.KV.NumInterfaceSteps = other.KV.NumInterfaceSteps
	} else {
		__antithesis_instrumentation__.Notify(472796)
	}
	__antithesis_instrumentation__.Notify(472762)
	if !result.KV.NumInternalSteps.HasValue() {
		__antithesis_instrumentation__.Notify(472797)
		result.KV.NumInternalSteps = other.KV.NumInternalSteps
	} else {
		__antithesis_instrumentation__.Notify(472798)
	}
	__antithesis_instrumentation__.Notify(472763)
	if !result.KV.NumInterfaceSeeks.HasValue() {
		__antithesis_instrumentation__.Notify(472799)
		result.KV.NumInterfaceSeeks = other.KV.NumInterfaceSeeks
	} else {
		__antithesis_instrumentation__.Notify(472800)
	}
	__antithesis_instrumentation__.Notify(472764)
	if !result.KV.NumInternalSeeks.HasValue() {
		__antithesis_instrumentation__.Notify(472801)
		result.KV.NumInternalSeeks = other.KV.NumInternalSeeks
	} else {
		__antithesis_instrumentation__.Notify(472802)
	}
	__antithesis_instrumentation__.Notify(472765)
	if !result.KV.TuplesRead.HasValue() {
		__antithesis_instrumentation__.Notify(472803)
		result.KV.TuplesRead = other.KV.TuplesRead
	} else {
		__antithesis_instrumentation__.Notify(472804)
	}
	__antithesis_instrumentation__.Notify(472766)
	if !result.KV.BytesRead.HasValue() {
		__antithesis_instrumentation__.Notify(472805)
		result.KV.BytesRead = other.KV.BytesRead
	} else {
		__antithesis_instrumentation__.Notify(472806)
	}
	__antithesis_instrumentation__.Notify(472767)

	if !result.Exec.ExecTime.HasValue() {
		__antithesis_instrumentation__.Notify(472807)
		result.Exec.ExecTime = other.Exec.ExecTime
	} else {
		__antithesis_instrumentation__.Notify(472808)
	}
	__antithesis_instrumentation__.Notify(472768)
	if !result.Exec.MaxAllocatedMem.HasValue() {
		__antithesis_instrumentation__.Notify(472809)
		result.Exec.MaxAllocatedMem = other.Exec.MaxAllocatedMem
	} else {
		__antithesis_instrumentation__.Notify(472810)
	}
	__antithesis_instrumentation__.Notify(472769)
	if !result.Exec.MaxAllocatedDisk.HasValue() {
		__antithesis_instrumentation__.Notify(472811)
		result.Exec.MaxAllocatedDisk = other.Exec.MaxAllocatedDisk
	} else {
		__antithesis_instrumentation__.Notify(472812)
	}
	__antithesis_instrumentation__.Notify(472770)

	if !result.Output.NumBatches.HasValue() {
		__antithesis_instrumentation__.Notify(472813)
		result.Output.NumBatches = other.Output.NumBatches
	} else {
		__antithesis_instrumentation__.Notify(472814)
	}
	__antithesis_instrumentation__.Notify(472771)
	if !result.Output.NumTuples.HasValue() {
		__antithesis_instrumentation__.Notify(472815)
		result.Output.NumTuples = other.Output.NumTuples
	} else {
		__antithesis_instrumentation__.Notify(472816)
	}
	__antithesis_instrumentation__.Notify(472772)

	if !result.FlowStats.MaxMemUsage.HasValue() {
		__antithesis_instrumentation__.Notify(472817)
		result.FlowStats.MaxMemUsage = other.FlowStats.MaxMemUsage
	} else {
		__antithesis_instrumentation__.Notify(472818)
	}
	__antithesis_instrumentation__.Notify(472773)
	if !result.FlowStats.MaxDiskUsage.HasValue() {
		__antithesis_instrumentation__.Notify(472819)
		result.FlowStats.MaxDiskUsage = other.FlowStats.MaxDiskUsage
	} else {
		__antithesis_instrumentation__.Notify(472820)
	}
	__antithesis_instrumentation__.Notify(472774)

	return &result
}

func (s *ComponentStats) MakeDeterministic() {
	__antithesis_instrumentation__.Notify(472821)

	resetUint := func(v *optional.Uint) {
		__antithesis_instrumentation__.Notify(472829)
		if v.HasValue() {
			__antithesis_instrumentation__.Notify(472830)
			v.Set(0)
		} else {
			__antithesis_instrumentation__.Notify(472831)
		}
	}
	__antithesis_instrumentation__.Notify(472822)

	timeVal := func(v *optional.Duration) {
		__antithesis_instrumentation__.Notify(472832)
		if v.HasValue() {
			__antithesis_instrumentation__.Notify(472833)
			v.Set(0)
		} else {
			__antithesis_instrumentation__.Notify(472834)
		}
	}
	__antithesis_instrumentation__.Notify(472823)

	timeVal(&s.NetRx.Latency)
	timeVal(&s.NetRx.WaitTime)
	timeVal(&s.NetRx.DeserializationTime)
	if s.NetRx.BytesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472835)

		s.NetRx.BytesReceived.Set(8 * s.NetRx.TuplesReceived.Value())
	} else {
		__antithesis_instrumentation__.Notify(472836)
	}
	__antithesis_instrumentation__.Notify(472824)
	if s.NetRx.MessagesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(472837)

		s.NetRx.MessagesReceived.Set(s.NetRx.TuplesReceived.Value() / 2)
	} else {
		__antithesis_instrumentation__.Notify(472838)
	}
	__antithesis_instrumentation__.Notify(472825)

	if s.NetTx.BytesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472839)

		s.NetTx.BytesSent.Set(8 * s.NetTx.TuplesSent.Value())
	} else {
		__antithesis_instrumentation__.Notify(472840)
	}
	__antithesis_instrumentation__.Notify(472826)
	if s.NetTx.MessagesSent.HasValue() {
		__antithesis_instrumentation__.Notify(472841)

		s.NetTx.MessagesSent.Set(s.NetTx.TuplesSent.Value() / 2)
	} else {
		__antithesis_instrumentation__.Notify(472842)
	}
	__antithesis_instrumentation__.Notify(472827)

	timeVal(&s.KV.KVTime)
	timeVal(&s.KV.ContentionTime)
	resetUint(&s.KV.NumInterfaceSteps)
	resetUint(&s.KV.NumInternalSteps)
	resetUint(&s.KV.NumInterfaceSeeks)
	resetUint(&s.KV.NumInternalSeeks)
	if s.KV.BytesRead.HasValue() {
		__antithesis_instrumentation__.Notify(472843)

		s.KV.BytesRead.Set(8 * s.KV.TuplesRead.Value())
	} else {
		__antithesis_instrumentation__.Notify(472844)
	}
	__antithesis_instrumentation__.Notify(472828)

	timeVal(&s.Exec.ExecTime)
	resetUint(&s.Exec.MaxAllocatedMem)
	resetUint(&s.Exec.MaxAllocatedDisk)

	resetUint(&s.Output.NumBatches)

	for i := range s.Inputs {
		__antithesis_instrumentation__.Notify(472845)
		timeVal(&s.Inputs[i].WaitTime)
	}
}

func ExtractStatsFromSpans(
	spans []tracingpb.RecordedSpan, makeDeterministic bool,
) map[ComponentID]*ComponentStats {
	__antithesis_instrumentation__.Notify(472846)
	statsMap := make(map[ComponentID]*ComponentStats)

	var componentStats ComponentStats
	for i := range spans {
		__antithesis_instrumentation__.Notify(472848)
		span := &spans[i]
		span.Structured(func(item *types.Any, _ time.Time) {
			__antithesis_instrumentation__.Notify(472849)
			if !types.Is(item, &componentStats) {
				__antithesis_instrumentation__.Notify(472854)
				return
			} else {
				__antithesis_instrumentation__.Notify(472855)
			}
			__antithesis_instrumentation__.Notify(472850)
			var stats ComponentStats
			if err := protoutil.Unmarshal(item.Value, &stats); err != nil {
				__antithesis_instrumentation__.Notify(472856)
				return
			} else {
				__antithesis_instrumentation__.Notify(472857)
			}
			__antithesis_instrumentation__.Notify(472851)
			if stats.Component == (ComponentID{}) {
				__antithesis_instrumentation__.Notify(472858)
				return
			} else {
				__antithesis_instrumentation__.Notify(472859)
			}
			__antithesis_instrumentation__.Notify(472852)
			if makeDeterministic {
				__antithesis_instrumentation__.Notify(472860)
				stats.MakeDeterministic()
			} else {
				__antithesis_instrumentation__.Notify(472861)
			}
			__antithesis_instrumentation__.Notify(472853)
			existing := statsMap[stats.Component]
			if existing == nil {
				__antithesis_instrumentation__.Notify(472862)
				statsMap[stats.Component] = &stats
			} else {
				__antithesis_instrumentation__.Notify(472863)

				statsMap[stats.Component] = existing.Union(&stats)
			}
		})
	}
	__antithesis_instrumentation__.Notify(472847)
	return statsMap
}

func ExtractNodesFromSpans(ctx context.Context, spans []tracingpb.RecordedSpan) util.FastIntSet {
	__antithesis_instrumentation__.Notify(472864)
	var nodes util.FastIntSet

	var componentStats ComponentStats
	for i := range spans {
		__antithesis_instrumentation__.Notify(472866)
		span := &spans[i]
		span.Structured(func(item *types.Any, _ time.Time) {
			__antithesis_instrumentation__.Notify(472867)
			if !types.Is(item, &componentStats) {
				__antithesis_instrumentation__.Notify(472871)
				return
			} else {
				__antithesis_instrumentation__.Notify(472872)
			}
			__antithesis_instrumentation__.Notify(472868)
			var stats ComponentStats
			if err := protoutil.Unmarshal(item.Value, &stats); err != nil {
				__antithesis_instrumentation__.Notify(472873)
				return
			} else {
				__antithesis_instrumentation__.Notify(472874)
			}
			__antithesis_instrumentation__.Notify(472869)
			if stats.Component == (ComponentID{}) {
				__antithesis_instrumentation__.Notify(472875)
				return
			} else {
				__antithesis_instrumentation__.Notify(472876)
			}
			__antithesis_instrumentation__.Notify(472870)
			nodes.Add(int(stats.Component.SQLInstanceID))
		})
	}
	__antithesis_instrumentation__.Notify(472865)
	return nodes
}
