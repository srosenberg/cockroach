package execstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/cockroachdb/errors"
)

type streamStats struct {
	originSQLInstanceID      base.SQLInstanceID
	destinationSQLInstanceID base.SQLInstanceID
	stats                    *execinfrapb.ComponentStats
}

type flowStats struct {
	stats []*execinfrapb.ComponentStats
}

type FlowsMetadata struct {
	flowID execinfrapb.FlowID

	processorStats map[execinfrapb.ProcessorID]*execinfrapb.ComponentStats

	streamStats map[execinfrapb.StreamID]*streamStats

	flowStats map[base.SQLInstanceID]*flowStats
}

func NewFlowsMetadata(flows map[base.SQLInstanceID]*execinfrapb.FlowSpec) *FlowsMetadata {
	__antithesis_instrumentation__.Notify(490822)
	a := &FlowsMetadata{
		processorStats: make(map[execinfrapb.ProcessorID]*execinfrapb.ComponentStats),
		streamStats:    make(map[execinfrapb.StreamID]*streamStats),
		flowStats:      make(map[base.SQLInstanceID]*flowStats),
	}

	for sqlInstanceID, flow := range flows {
		__antithesis_instrumentation__.Notify(490824)
		if a.flowID.IsUnset() {
			__antithesis_instrumentation__.Notify(490826)
			a.flowID = flow.FlowID
		} else {
			__antithesis_instrumentation__.Notify(490827)
			if buildutil.CrdbTestBuild && func() bool {
				__antithesis_instrumentation__.Notify(490828)
				return !a.flowID.Equal(flow.FlowID) == true
			}() == true {
				__antithesis_instrumentation__.Notify(490829)
				panic(
					errors.AssertionFailedf(
						"expected the same FlowID to be used for all flows. UUID of first flow: %v, UUID of flow on node %s: %v",
						a.flowID, sqlInstanceID, flow.FlowID),
				)
			} else {
				__antithesis_instrumentation__.Notify(490830)
			}
		}
		__antithesis_instrumentation__.Notify(490825)
		a.flowStats[sqlInstanceID] = &flowStats{}
		for _, proc := range flow.Processors {
			__antithesis_instrumentation__.Notify(490831)
			procID := execinfrapb.ProcessorID(proc.ProcessorID)
			a.processorStats[procID] = &execinfrapb.ComponentStats{}
			a.processorStats[procID].Component.SQLInstanceID = sqlInstanceID

			for _, output := range proc.Output {
				__antithesis_instrumentation__.Notify(490832)
				for _, stream := range output.Streams {
					__antithesis_instrumentation__.Notify(490833)
					a.streamStats[stream.StreamID] = &streamStats{
						originSQLInstanceID:      sqlInstanceID,
						destinationSQLInstanceID: stream.TargetNodeID,
					}
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(490823)

	return a
}

type NodeLevelStats struct {
	NetworkBytesSentGroupedByNode map[base.SQLInstanceID]int64
	MaxMemoryUsageGroupedByNode   map[base.SQLInstanceID]int64
	MaxDiskUsageGroupedByNode     map[base.SQLInstanceID]int64
	KVBytesReadGroupedByNode      map[base.SQLInstanceID]int64
	KVRowsReadGroupedByNode       map[base.SQLInstanceID]int64
	KVTimeGroupedByNode           map[base.SQLInstanceID]time.Duration
	NetworkMessagesGroupedByNode  map[base.SQLInstanceID]int64
	ContentionTimeGroupedByNode   map[base.SQLInstanceID]time.Duration
}

type QueryLevelStats struct {
	NetworkBytesSent int64
	MaxMemUsage      int64
	MaxDiskUsage     int64
	KVBytesRead      int64
	KVRowsRead       int64
	KVTime           time.Duration
	NetworkMessages  int64
	ContentionTime   time.Duration
	Regions          []string
}

func (s *QueryLevelStats) Accumulate(other QueryLevelStats) {
	__antithesis_instrumentation__.Notify(490834)
	s.NetworkBytesSent += other.NetworkBytesSent
	if other.MaxMemUsage > s.MaxMemUsage {
		__antithesis_instrumentation__.Notify(490837)
		s.MaxMemUsage = other.MaxMemUsage
	} else {
		__antithesis_instrumentation__.Notify(490838)
	}
	__antithesis_instrumentation__.Notify(490835)
	if other.MaxDiskUsage > s.MaxDiskUsage {
		__antithesis_instrumentation__.Notify(490839)
		s.MaxDiskUsage = other.MaxDiskUsage
	} else {
		__antithesis_instrumentation__.Notify(490840)
	}
	__antithesis_instrumentation__.Notify(490836)
	s.KVBytesRead += other.KVBytesRead
	s.KVRowsRead += other.KVRowsRead
	s.KVTime += other.KVTime
	s.NetworkMessages += other.NetworkMessages
	s.ContentionTime += other.ContentionTime
	s.Regions = util.CombineUniqueString(s.Regions, other.Regions)
}

type TraceAnalyzer struct {
	*FlowsMetadata
	nodeLevelStats  NodeLevelStats
	queryLevelStats QueryLevelStats
}

func NewTraceAnalyzer(flowsMetadata *FlowsMetadata) *TraceAnalyzer {
	__antithesis_instrumentation__.Notify(490841)
	return &TraceAnalyzer{FlowsMetadata: flowsMetadata}
}

func (a *TraceAnalyzer) AddTrace(trace []tracingpb.RecordedSpan, makeDeterministic bool) error {
	__antithesis_instrumentation__.Notify(490842)
	m := execinfrapb.ExtractStatsFromSpans(trace, makeDeterministic)

	for component, componentStats := range m {
		__antithesis_instrumentation__.Notify(490844)
		if !component.FlowID.Equal(a.flowID) {
			__antithesis_instrumentation__.Notify(490846)

			continue
		} else {
			__antithesis_instrumentation__.Notify(490847)
		}
		__antithesis_instrumentation__.Notify(490845)
		switch component.Type {
		case execinfrapb.ComponentID_PROCESSOR:
			__antithesis_instrumentation__.Notify(490848)
			id := component.ID
			a.processorStats[execinfrapb.ProcessorID(id)] = componentStats

		case execinfrapb.ComponentID_STREAM:
			__antithesis_instrumentation__.Notify(490849)
			id := component.ID
			streamStats := a.streamStats[execinfrapb.StreamID(id)]
			if streamStats == nil {
				__antithesis_instrumentation__.Notify(490854)
				return errors.Errorf("trace has span for stream %d but the stream does not exist in the physical plan", id)
			} else {
				__antithesis_instrumentation__.Notify(490855)
			}
			__antithesis_instrumentation__.Notify(490850)
			streamStats.stats = componentStats

		case execinfrapb.ComponentID_FLOW:
			__antithesis_instrumentation__.Notify(490851)
			flowStats := a.flowStats[component.SQLInstanceID]
			if flowStats == nil {
				__antithesis_instrumentation__.Notify(490856)
				return errors.Errorf(
					"trace has span for flow %s on node %s but the flow does not exist in the physical plan",
					component.FlowID,
					component.SQLInstanceID,
				)
			} else {
				__antithesis_instrumentation__.Notify(490857)
			}
			__antithesis_instrumentation__.Notify(490852)
			flowStats.stats = append(flowStats.stats, componentStats)
		default:
			__antithesis_instrumentation__.Notify(490853)
		}
	}
	__antithesis_instrumentation__.Notify(490843)

	return nil
}

func (a *TraceAnalyzer) ProcessStats() error {
	__antithesis_instrumentation__.Notify(490858)

	a.nodeLevelStats = NodeLevelStats{
		NetworkBytesSentGroupedByNode: make(map[base.SQLInstanceID]int64),
		MaxMemoryUsageGroupedByNode:   make(map[base.SQLInstanceID]int64),
		MaxDiskUsageGroupedByNode:     make(map[base.SQLInstanceID]int64),
		KVBytesReadGroupedByNode:      make(map[base.SQLInstanceID]int64),
		KVRowsReadGroupedByNode:       make(map[base.SQLInstanceID]int64),
		KVTimeGroupedByNode:           make(map[base.SQLInstanceID]time.Duration),
		NetworkMessagesGroupedByNode:  make(map[base.SQLInstanceID]int64),
		ContentionTimeGroupedByNode:   make(map[base.SQLInstanceID]time.Duration),
	}
	var errs error

	for _, stats := range a.processorStats {
		__antithesis_instrumentation__.Notify(490870)
		if stats == nil {
			__antithesis_instrumentation__.Notify(490872)
			continue
		} else {
			__antithesis_instrumentation__.Notify(490873)
		}
		__antithesis_instrumentation__.Notify(490871)
		instanceID := stats.Component.SQLInstanceID
		a.nodeLevelStats.KVBytesReadGroupedByNode[instanceID] += int64(stats.KV.BytesRead.Value())
		a.nodeLevelStats.KVRowsReadGroupedByNode[instanceID] += int64(stats.KV.TuplesRead.Value())
		a.nodeLevelStats.KVTimeGroupedByNode[instanceID] += stats.KV.KVTime.Value()
		a.nodeLevelStats.ContentionTimeGroupedByNode[instanceID] += stats.KV.ContentionTime.Value()
	}
	__antithesis_instrumentation__.Notify(490859)

	for _, stats := range a.streamStats {
		__antithesis_instrumentation__.Notify(490874)
		if stats.stats == nil {
			__antithesis_instrumentation__.Notify(490879)
			continue
		} else {
			__antithesis_instrumentation__.Notify(490880)
		}
		__antithesis_instrumentation__.Notify(490875)
		originInstanceID := stats.originSQLInstanceID

		bytes, err := getNetworkBytesFromComponentStats(stats.stats)
		if err != nil {
			__antithesis_instrumentation__.Notify(490881)
			errs = errors.CombineErrors(errs, errors.Wrap(err, "error calculating network bytes sent"))
		} else {
			__antithesis_instrumentation__.Notify(490882)
			a.nodeLevelStats.NetworkBytesSentGroupedByNode[originInstanceID] += bytes
		}
		__antithesis_instrumentation__.Notify(490876)

		if stats.stats.FlowStats.MaxMemUsage.HasValue() {
			__antithesis_instrumentation__.Notify(490883)
			memUsage := int64(stats.stats.FlowStats.MaxMemUsage.Value())
			if memUsage > a.nodeLevelStats.MaxMemoryUsageGroupedByNode[originInstanceID] {
				__antithesis_instrumentation__.Notify(490884)
				a.nodeLevelStats.MaxMemoryUsageGroupedByNode[originInstanceID] = memUsage
			} else {
				__antithesis_instrumentation__.Notify(490885)
			}
		} else {
			__antithesis_instrumentation__.Notify(490886)
		}
		__antithesis_instrumentation__.Notify(490877)
		if stats.stats.FlowStats.MaxDiskUsage.HasValue() {
			__antithesis_instrumentation__.Notify(490887)
			if diskUsage := int64(stats.stats.FlowStats.MaxDiskUsage.Value()); diskUsage > a.nodeLevelStats.MaxDiskUsageGroupedByNode[originInstanceID] {
				__antithesis_instrumentation__.Notify(490888)
				a.nodeLevelStats.MaxDiskUsageGroupedByNode[originInstanceID] = diskUsage
			} else {
				__antithesis_instrumentation__.Notify(490889)
			}
		} else {
			__antithesis_instrumentation__.Notify(490890)
		}
		__antithesis_instrumentation__.Notify(490878)

		numMessages, err := getNumNetworkMessagesFromComponentsStats(stats.stats)
		if err != nil {
			__antithesis_instrumentation__.Notify(490891)
			errs = errors.CombineErrors(errs, errors.Wrap(err, "error calculating number of network messages"))
		} else {
			__antithesis_instrumentation__.Notify(490892)
			a.nodeLevelStats.NetworkMessagesGroupedByNode[originInstanceID] += numMessages
		}
	}
	__antithesis_instrumentation__.Notify(490860)

	for instanceID, stats := range a.flowStats {
		__antithesis_instrumentation__.Notify(490893)
		if stats.stats == nil {
			__antithesis_instrumentation__.Notify(490895)
			continue
		} else {
			__antithesis_instrumentation__.Notify(490896)
		}
		__antithesis_instrumentation__.Notify(490894)

		for _, v := range stats.stats {
			__antithesis_instrumentation__.Notify(490897)
			if v.FlowStats.MaxMemUsage.HasValue() {
				__antithesis_instrumentation__.Notify(490899)
				if memUsage := int64(v.FlowStats.MaxMemUsage.Value()); memUsage > a.nodeLevelStats.MaxMemoryUsageGroupedByNode[instanceID] {
					__antithesis_instrumentation__.Notify(490900)
					a.nodeLevelStats.MaxMemoryUsageGroupedByNode[instanceID] = memUsage
				} else {
					__antithesis_instrumentation__.Notify(490901)
				}
			} else {
				__antithesis_instrumentation__.Notify(490902)
			}
			__antithesis_instrumentation__.Notify(490898)
			if v.FlowStats.MaxDiskUsage.HasValue() {
				__antithesis_instrumentation__.Notify(490903)
				if diskUsage := int64(v.FlowStats.MaxDiskUsage.Value()); diskUsage > a.nodeLevelStats.MaxDiskUsageGroupedByNode[instanceID] {
					__antithesis_instrumentation__.Notify(490904)
					a.nodeLevelStats.MaxDiskUsageGroupedByNode[instanceID] = diskUsage
				} else {
					__antithesis_instrumentation__.Notify(490905)
				}

			} else {
				__antithesis_instrumentation__.Notify(490906)
			}
		}
	}
	__antithesis_instrumentation__.Notify(490861)

	a.queryLevelStats = QueryLevelStats{}

	for _, bytesSentByNode := range a.nodeLevelStats.NetworkBytesSentGroupedByNode {
		__antithesis_instrumentation__.Notify(490907)
		a.queryLevelStats.NetworkBytesSent += bytesSentByNode
	}
	__antithesis_instrumentation__.Notify(490862)

	for _, maxMemUsage := range a.nodeLevelStats.MaxMemoryUsageGroupedByNode {
		__antithesis_instrumentation__.Notify(490908)
		if maxMemUsage > a.queryLevelStats.MaxMemUsage {
			__antithesis_instrumentation__.Notify(490909)
			a.queryLevelStats.MaxMemUsage = maxMemUsage
		} else {
			__antithesis_instrumentation__.Notify(490910)
		}
	}
	__antithesis_instrumentation__.Notify(490863)

	for _, maxDiskUsage := range a.nodeLevelStats.MaxDiskUsageGroupedByNode {
		__antithesis_instrumentation__.Notify(490911)
		if maxDiskUsage > a.queryLevelStats.MaxDiskUsage {
			__antithesis_instrumentation__.Notify(490912)
			a.queryLevelStats.MaxDiskUsage = maxDiskUsage
		} else {
			__antithesis_instrumentation__.Notify(490913)
		}
	}
	__antithesis_instrumentation__.Notify(490864)

	for _, kvBytesRead := range a.nodeLevelStats.KVBytesReadGroupedByNode {
		__antithesis_instrumentation__.Notify(490914)
		a.queryLevelStats.KVBytesRead += kvBytesRead
	}
	__antithesis_instrumentation__.Notify(490865)

	for _, kvRowsRead := range a.nodeLevelStats.KVRowsReadGroupedByNode {
		__antithesis_instrumentation__.Notify(490915)
		a.queryLevelStats.KVRowsRead += kvRowsRead
	}
	__antithesis_instrumentation__.Notify(490866)

	for _, kvTime := range a.nodeLevelStats.KVTimeGroupedByNode {
		__antithesis_instrumentation__.Notify(490916)
		a.queryLevelStats.KVTime += kvTime
	}
	__antithesis_instrumentation__.Notify(490867)

	for _, networkMessages := range a.nodeLevelStats.NetworkMessagesGroupedByNode {
		__antithesis_instrumentation__.Notify(490917)
		a.queryLevelStats.NetworkMessages += networkMessages
	}
	__antithesis_instrumentation__.Notify(490868)

	for _, contentionTime := range a.nodeLevelStats.ContentionTimeGroupedByNode {
		__antithesis_instrumentation__.Notify(490918)
		a.queryLevelStats.ContentionTime += contentionTime
	}
	__antithesis_instrumentation__.Notify(490869)
	return errs
}

func getNetworkBytesFromComponentStats(v *execinfrapb.ComponentStats) (int64, error) {
	__antithesis_instrumentation__.Notify(490919)

	if v.NetRx.BytesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(490922)
		if v.NetTx.BytesSent.HasValue() {
			__antithesis_instrumentation__.Notify(490924)
			return 0, errors.Errorf("could not get network bytes; both BytesReceived and BytesSent are set")
		} else {
			__antithesis_instrumentation__.Notify(490925)
		}
		__antithesis_instrumentation__.Notify(490923)
		return int64(v.NetRx.BytesReceived.Value()), nil
	} else {
		__antithesis_instrumentation__.Notify(490926)
	}
	__antithesis_instrumentation__.Notify(490920)
	if v.NetTx.BytesSent.HasValue() {
		__antithesis_instrumentation__.Notify(490927)
		return int64(v.NetTx.BytesSent.Value()), nil
	} else {
		__antithesis_instrumentation__.Notify(490928)
	}
	__antithesis_instrumentation__.Notify(490921)

	return 0, nil
}

func getNumNetworkMessagesFromComponentsStats(v *execinfrapb.ComponentStats) (int64, error) {
	__antithesis_instrumentation__.Notify(490929)

	if v.NetRx.MessagesReceived.HasValue() {
		__antithesis_instrumentation__.Notify(490932)
		if v.NetTx.MessagesSent.HasValue() {
			__antithesis_instrumentation__.Notify(490934)
			return 0, errors.Errorf("could not get network messages; both MessagesReceived and MessagesSent are set")
		} else {
			__antithesis_instrumentation__.Notify(490935)
		}
		__antithesis_instrumentation__.Notify(490933)
		return int64(v.NetRx.MessagesReceived.Value()), nil
	} else {
		__antithesis_instrumentation__.Notify(490936)
	}
	__antithesis_instrumentation__.Notify(490930)
	if v.NetTx.MessagesSent.HasValue() {
		__antithesis_instrumentation__.Notify(490937)
		return int64(v.NetTx.MessagesSent.Value()), nil
	} else {
		__antithesis_instrumentation__.Notify(490938)
	}
	__antithesis_instrumentation__.Notify(490931)

	return 0, nil
}

func (a *TraceAnalyzer) GetNodeLevelStats() NodeLevelStats {
	__antithesis_instrumentation__.Notify(490939)
	return a.nodeLevelStats
}

func (a *TraceAnalyzer) GetQueryLevelStats() QueryLevelStats {
	__antithesis_instrumentation__.Notify(490940)
	return a.queryLevelStats
}

func GetQueryLevelStats(
	trace []tracingpb.RecordedSpan, deterministicExplainAnalyze bool, flowsMetadata []*FlowsMetadata,
) (QueryLevelStats, error) {
	__antithesis_instrumentation__.Notify(490941)
	var queryLevelStats QueryLevelStats
	var errs error
	for _, metadata := range flowsMetadata {
		__antithesis_instrumentation__.Notify(490943)
		analyzer := NewTraceAnalyzer(metadata)
		if err := analyzer.AddTrace(trace, deterministicExplainAnalyze); err != nil {
			__antithesis_instrumentation__.Notify(490946)
			errs = errors.CombineErrors(errs, errors.Wrap(err, "error analyzing trace statistics"))
			continue
		} else {
			__antithesis_instrumentation__.Notify(490947)
		}
		__antithesis_instrumentation__.Notify(490944)

		if err := analyzer.ProcessStats(); err != nil {
			__antithesis_instrumentation__.Notify(490948)
			errs = errors.CombineErrors(errs, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(490949)
		}
		__antithesis_instrumentation__.Notify(490945)
		queryLevelStats.Accumulate(analyzer.GetQueryLevelStats())
	}
	__antithesis_instrumentation__.Notify(490942)
	return queryLevelStats, errs
}
