package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build"
	"github.com/cockroachdb/cockroach/pkg/gossip"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/cgroups"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/system"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/codahale/hdrhistogram"
	humanize "github.com/dustin/go-humanize"
	"github.com/elastic/gosigar"
)

const (
	storeTimeSeriesPrefix = "cr.store.%s"

	nodeTimeSeriesPrefix = "cr.node.%s"

	advertiseAddrLabelKey = "advertise-addr"
	httpAddrLabelKey      = "http-addr"
	sqlAddrLabelKey       = "sql-addr"
)

type quantile struct {
	suffix   string
	quantile float64
}

var recordHistogramQuantiles = []quantile{
	{"-max", 100},
	{"-p99.999", 99.999},
	{"-p99.99", 99.99},
	{"-p99.9", 99.9},
	{"-p99", 99},
	{"-p90", 90},
	{"-p75", 75},
	{"-p50", 50},
}

type storeMetrics interface {
	StoreID() roachpb.StoreID
	Descriptor(context.Context, bool) (*roachpb.StoreDescriptor, error)
	Registry() *metric.Registry
}

var childMetricsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable, "server.child_metrics.enabled",
	"enables the exporting of child metrics, additional prometheus time series with extra labels",
	false).WithPublic()

type MetricsRecorder struct {
	*HealthChecker
	gossip       *gossip.Gossip
	nodeLiveness *liveness.NodeLiveness
	rpcContext   *rpc.Context
	settings     *cluster.Settings
	clock        *hlc.Clock

	lastDataCount        int64
	lastSummaryCount     int64
	lastNodeMetricCount  int64
	lastStoreMetricCount int64

	mu struct {
		syncutil.RWMutex

		nodeRegistry *metric.Registry
		desc         roachpb.NodeDescriptor
		startedAt    int64

		storeRegistries map[roachpb.StoreID]*metric.Registry
		stores          map[roachpb.StoreID]storeMetrics
	}

	prometheusExporter metric.PrometheusExporter

	writeSummaryMu syncutil.Mutex
}

func NewMetricsRecorder(
	clock *hlc.Clock,
	nodeLiveness *liveness.NodeLiveness,
	rpcContext *rpc.Context,
	gossip *gossip.Gossip,
	settings *cluster.Settings,
) *MetricsRecorder {
	__antithesis_instrumentation__.Notify(235463)
	mr := &MetricsRecorder{
		HealthChecker: NewHealthChecker(trackedMetrics),
		nodeLiveness:  nodeLiveness,
		rpcContext:    rpcContext,
		gossip:        gossip,
		settings:      settings,
	}
	mr.mu.storeRegistries = make(map[roachpb.StoreID]*metric.Registry)
	mr.mu.stores = make(map[roachpb.StoreID]storeMetrics)
	mr.prometheusExporter = metric.MakePrometheusExporter()
	mr.clock = clock
	return mr
}

func (mr *MetricsRecorder) AddNode(
	reg *metric.Registry,
	desc roachpb.NodeDescriptor,
	startedAt int64,
	advertiseAddr, httpAddr, sqlAddr string,
) {
	__antithesis_instrumentation__.Notify(235464)
	mr.mu.Lock()
	defer mr.mu.Unlock()
	mr.mu.nodeRegistry = reg
	mr.mu.desc = desc
	mr.mu.startedAt = startedAt

	metadata := metric.Metadata{
		Name:        "node-id",
		Help:        "node ID with labels for advertised RPC and HTTP addresses",
		Measurement: "Node ID",
		Unit:        metric.Unit_CONST,
	}

	metadata.AddLabel(advertiseAddrLabelKey, advertiseAddr)
	metadata.AddLabel(httpAddrLabelKey, httpAddr)
	metadata.AddLabel(sqlAddrLabelKey, sqlAddr)
	nodeIDGauge := metric.NewGauge(metadata)
	nodeIDGauge.Update(int64(desc.NodeID))
	reg.AddMetric(nodeIDGauge)
}

func (mr *MetricsRecorder) AddStore(store storeMetrics) {
	__antithesis_instrumentation__.Notify(235465)
	mr.mu.Lock()
	defer mr.mu.Unlock()
	storeID := store.StoreID()
	store.Registry().AddLabel("store", strconv.Itoa(int(storeID)))
	mr.mu.storeRegistries[storeID] = store.Registry()
	mr.mu.stores[storeID] = store
}

func (mr *MetricsRecorder) MarshalJSON() ([]byte, error) {
	__antithesis_instrumentation__.Notify(235466)
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	if mr.mu.nodeRegistry == nil {
		__antithesis_instrumentation__.Notify(235469)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(235471)
			log.Warning(context.TODO(), "MetricsRecorder.MarshalJSON() called before NodeID allocation")
		} else {
			__antithesis_instrumentation__.Notify(235472)
		}
		__antithesis_instrumentation__.Notify(235470)
		return []byte("{}"), nil
	} else {
		__antithesis_instrumentation__.Notify(235473)
	}
	__antithesis_instrumentation__.Notify(235467)
	topLevel := map[string]interface{}{
		fmt.Sprintf("node.%d", mr.mu.desc.NodeID): mr.mu.nodeRegistry,
	}

	storeLevel := make(map[string]interface{})
	for id, reg := range mr.mu.storeRegistries {
		__antithesis_instrumentation__.Notify(235474)
		storeLevel[strconv.Itoa(int(id))] = reg
	}
	__antithesis_instrumentation__.Notify(235468)
	topLevel["stores"] = storeLevel
	return json.Marshal(topLevel)
}

func (mr *MetricsRecorder) ScrapeIntoPrometheus(pm *metric.PrometheusExporter) {
	__antithesis_instrumentation__.Notify(235475)
	mr.mu.RLock()
	defer mr.mu.RUnlock()
	if mr.mu.nodeRegistry == nil {
		__antithesis_instrumentation__.Notify(235477)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(235478)
			log.Warning(context.TODO(), "MetricsRecorder asked to scrape metrics before NodeID allocation")
		} else {
			__antithesis_instrumentation__.Notify(235479)
		}
	} else {
		__antithesis_instrumentation__.Notify(235480)
	}
	__antithesis_instrumentation__.Notify(235476)
	includeChildMetrics := childMetricsEnabled.Get(&mr.settings.SV)
	pm.ScrapeRegistry(mr.mu.nodeRegistry, includeChildMetrics)
	for _, reg := range mr.mu.storeRegistries {
		__antithesis_instrumentation__.Notify(235481)
		pm.ScrapeRegistry(reg, includeChildMetrics)
	}
}

func (mr *MetricsRecorder) PrintAsText(w io.Writer) error {
	__antithesis_instrumentation__.Notify(235482)
	var buf bytes.Buffer
	if err := mr.prometheusExporter.ScrapeAndPrintAsText(&buf, mr.ScrapeIntoPrometheus); err != nil {
		__antithesis_instrumentation__.Notify(235484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(235485)
	}
	__antithesis_instrumentation__.Notify(235483)
	_, err := buf.WriteTo(w)
	return err
}

func (mr *MetricsRecorder) ExportToGraphite(
	ctx context.Context, endpoint string, pm *metric.PrometheusExporter,
) error {
	__antithesis_instrumentation__.Notify(235486)
	mr.ScrapeIntoPrometheus(pm)
	graphiteExporter := metric.MakeGraphiteExporter(pm)
	return graphiteExporter.Push(ctx, endpoint)
}

func (mr *MetricsRecorder) GetTimeSeriesData() []tspb.TimeSeriesData {
	__antithesis_instrumentation__.Notify(235487)
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.mu.nodeRegistry == nil {
		__antithesis_instrumentation__.Notify(235490)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(235492)
			log.Warning(context.TODO(), "MetricsRecorder.GetTimeSeriesData() called before NodeID allocation")
		} else {
			__antithesis_instrumentation__.Notify(235493)
		}
		__antithesis_instrumentation__.Notify(235491)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(235494)
	}
	__antithesis_instrumentation__.Notify(235488)

	lastDataCount := atomic.LoadInt64(&mr.lastDataCount)
	data := make([]tspb.TimeSeriesData, 0, lastDataCount)

	now := mr.clock.PhysicalNow()
	recorder := registryRecorder{
		registry:       mr.mu.nodeRegistry,
		format:         nodeTimeSeriesPrefix,
		source:         strconv.FormatInt(int64(mr.mu.desc.NodeID), 10),
		timestampNanos: now,
	}
	recorder.record(&data)

	for storeID, r := range mr.mu.storeRegistries {
		__antithesis_instrumentation__.Notify(235495)
		storeRecorder := registryRecorder{
			registry:       r,
			format:         storeTimeSeriesPrefix,
			source:         strconv.FormatInt(int64(storeID), 10),
			timestampNanos: now,
		}
		storeRecorder.record(&data)
	}
	__antithesis_instrumentation__.Notify(235489)
	atomic.CompareAndSwapInt64(&mr.lastDataCount, lastDataCount, int64(len(data)))
	return data
}

func (mr *MetricsRecorder) GetMetricsMetadata() map[string]metric.Metadata {
	__antithesis_instrumentation__.Notify(235496)
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if mr.mu.nodeRegistry == nil {
		__antithesis_instrumentation__.Notify(235499)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(235501)
			log.Warning(context.TODO(), "MetricsRecorder.GetMetricsMetadata() called before NodeID allocation")
		} else {
			__antithesis_instrumentation__.Notify(235502)
		}
		__antithesis_instrumentation__.Notify(235500)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(235503)
	}
	__antithesis_instrumentation__.Notify(235497)

	metrics := make(map[string]metric.Metadata)

	mr.mu.nodeRegistry.WriteMetricsMetadata(metrics)

	var sID roachpb.StoreID

	for storeID := range mr.mu.storeRegistries {
		__antithesis_instrumentation__.Notify(235504)
		sID = storeID
		break
	}
	__antithesis_instrumentation__.Notify(235498)

	mr.mu.storeRegistries[sID].WriteMetricsMetadata(metrics)

	return metrics
}

func (mr *MetricsRecorder) getNetworkActivity(
	ctx context.Context,
) map[roachpb.NodeID]statuspb.NodeStatus_NetworkActivity {
	__antithesis_instrumentation__.Notify(235505)
	activity := make(map[roachpb.NodeID]statuspb.NodeStatus_NetworkActivity)
	if mr.nodeLiveness != nil && func() bool {
		__antithesis_instrumentation__.Notify(235507)
		return mr.gossip != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(235508)
		isLiveMap := mr.nodeLiveness.GetIsLiveMap()

		throughputMap := mr.rpcContext.GetStatsMap()
		var currentAverages map[string]time.Duration
		if mr.rpcContext.RemoteClocks != nil {
			__antithesis_instrumentation__.Notify(235510)
			currentAverages = mr.rpcContext.RemoteClocks.AllLatencies()
		} else {
			__antithesis_instrumentation__.Notify(235511)
		}
		__antithesis_instrumentation__.Notify(235509)
		for nodeID, entry := range isLiveMap {
			__antithesis_instrumentation__.Notify(235512)
			address, err := mr.gossip.GetNodeIDAddress(nodeID)
			if err != nil {
				__antithesis_instrumentation__.Notify(235516)
				if entry.IsLive {
					__antithesis_instrumentation__.Notify(235518)
					log.Warningf(ctx, "%v", err)
				} else {
					__antithesis_instrumentation__.Notify(235519)
				}
				__antithesis_instrumentation__.Notify(235517)
				continue
			} else {
				__antithesis_instrumentation__.Notify(235520)
			}
			__antithesis_instrumentation__.Notify(235513)
			na := statuspb.NodeStatus_NetworkActivity{}
			key := address.String()
			if tp, ok := throughputMap.Load(key); ok {
				__antithesis_instrumentation__.Notify(235521)
				stats := tp.(*rpc.Stats)
				na.Incoming = stats.Incoming()
				na.Outgoing = stats.Outgoing()
			} else {
				__antithesis_instrumentation__.Notify(235522)
			}
			__antithesis_instrumentation__.Notify(235514)
			if entry.IsLive {
				__antithesis_instrumentation__.Notify(235523)
				if latency, ok := currentAverages[key]; ok {
					__antithesis_instrumentation__.Notify(235524)
					na.Latency = latency.Nanoseconds()
				} else {
					__antithesis_instrumentation__.Notify(235525)
				}
			} else {
				__antithesis_instrumentation__.Notify(235526)
			}
			__antithesis_instrumentation__.Notify(235515)
			activity[nodeID] = na
		}
	} else {
		__antithesis_instrumentation__.Notify(235527)
	}
	__antithesis_instrumentation__.Notify(235506)
	return activity
}

func (mr *MetricsRecorder) GenerateNodeStatus(ctx context.Context) *statuspb.NodeStatus {
	__antithesis_instrumentation__.Notify(235528)
	activity := mr.getNetworkActivity(ctx)

	mr.mu.RLock()
	defer mr.mu.RUnlock()

	if mr.mu.nodeRegistry == nil {
		__antithesis_instrumentation__.Notify(235534)

		if log.V(1) {
			__antithesis_instrumentation__.Notify(235536)
			log.Warning(ctx, "attempt to generate status summary before NodeID allocation.")
		} else {
			__antithesis_instrumentation__.Notify(235537)
		}
		__antithesis_instrumentation__.Notify(235535)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(235538)
	}
	__antithesis_instrumentation__.Notify(235529)

	now := mr.clock.PhysicalNow()

	lastSummaryCount := atomic.LoadInt64(&mr.lastSummaryCount)
	lastNodeMetricCount := atomic.LoadInt64(&mr.lastNodeMetricCount)
	lastStoreMetricCount := atomic.LoadInt64(&mr.lastStoreMetricCount)

	systemMemory, _, err := GetTotalMemoryWithoutLogging()
	if err != nil {
		__antithesis_instrumentation__.Notify(235539)
		log.Errorf(ctx, "could not get total system memory: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(235540)
	}
	__antithesis_instrumentation__.Notify(235530)

	nodeStat := &statuspb.NodeStatus{
		Desc:              mr.mu.desc,
		BuildInfo:         build.GetInfo(),
		UpdatedAt:         now,
		StartedAt:         mr.mu.startedAt,
		StoreStatuses:     make([]statuspb.StoreStatus, 0, lastSummaryCount),
		Metrics:           make(map[string]float64, lastNodeMetricCount),
		Args:              os.Args,
		Env:               flattenStrings(envutil.GetEnvVarsUsed()),
		Activity:          activity,
		NumCpus:           int32(system.NumCPU()),
		TotalSystemMemory: systemMemory,
	}

	eachRecordableValue(mr.mu.nodeRegistry, func(name string, val float64) {
		__antithesis_instrumentation__.Notify(235541)
		nodeStat.Metrics[name] = val
	})
	__antithesis_instrumentation__.Notify(235531)

	for storeID, r := range mr.mu.storeRegistries {
		__antithesis_instrumentation__.Notify(235542)
		storeMetrics := make(map[string]float64, lastStoreMetricCount)
		eachRecordableValue(r, func(name string, val float64) {
			__antithesis_instrumentation__.Notify(235545)
			storeMetrics[name] = val
		})
		__antithesis_instrumentation__.Notify(235543)

		descriptor, err := mr.mu.stores[storeID].Descriptor(ctx, false)
		if err != nil {
			__antithesis_instrumentation__.Notify(235546)
			log.Errorf(ctx, "could not record status summaries: Store %d could not return descriptor, error: %s", storeID, err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(235547)
		}
		__antithesis_instrumentation__.Notify(235544)

		nodeStat.StoreStatuses = append(nodeStat.StoreStatuses, statuspb.StoreStatus{
			Desc:    *descriptor,
			Metrics: storeMetrics,
		})
	}
	__antithesis_instrumentation__.Notify(235532)

	atomic.CompareAndSwapInt64(
		&mr.lastSummaryCount, lastSummaryCount, int64(len(nodeStat.StoreStatuses)))
	atomic.CompareAndSwapInt64(
		&mr.lastNodeMetricCount, lastNodeMetricCount, int64(len(nodeStat.Metrics)))
	if len(nodeStat.StoreStatuses) > 0 {
		__antithesis_instrumentation__.Notify(235548)
		atomic.CompareAndSwapInt64(
			&mr.lastStoreMetricCount, lastStoreMetricCount, int64(len(nodeStat.StoreStatuses[0].Metrics)))
	} else {
		__antithesis_instrumentation__.Notify(235549)
	}
	__antithesis_instrumentation__.Notify(235533)

	return nodeStat
}

func flattenStrings(s []redact.RedactableString) []string {
	__antithesis_instrumentation__.Notify(235550)
	res := make([]string, len(s))
	for i, v := range s {
		__antithesis_instrumentation__.Notify(235552)
		res[i] = v.StripMarkers()
	}
	__antithesis_instrumentation__.Notify(235551)
	return res
}

func (mr *MetricsRecorder) WriteNodeStatus(
	ctx context.Context, db *kv.DB, nodeStatus statuspb.NodeStatus, mustExist bool,
) error {
	__antithesis_instrumentation__.Notify(235553)
	mr.writeSummaryMu.Lock()
	defer mr.writeSummaryMu.Unlock()
	key := keys.NodeStatusKey(nodeStatus.Desc.NodeID)

	if mustExist {
		__antithesis_instrumentation__.Notify(235556)
		entry, err := db.Get(ctx, key)
		if err != nil {
			__antithesis_instrumentation__.Notify(235559)
			return err
		} else {
			__antithesis_instrumentation__.Notify(235560)
		}
		__antithesis_instrumentation__.Notify(235557)
		if entry.Value == nil {
			__antithesis_instrumentation__.Notify(235561)
			return errors.New("status entry not found, node may have been decommissioned")
		} else {
			__antithesis_instrumentation__.Notify(235562)
		}
		__antithesis_instrumentation__.Notify(235558)
		err = db.CPutInline(ctx, key, &nodeStatus, entry.Value.TagAndDataBytes())
		if detail := (*roachpb.ConditionFailedError)(nil); errors.As(err, &detail) {
			__antithesis_instrumentation__.Notify(235563)
			if detail.ActualValue == nil {
				__antithesis_instrumentation__.Notify(235565)
				return errors.New("status entry not found, node may have been decommissioned")
			} else {
				__antithesis_instrumentation__.Notify(235566)
			}
			__antithesis_instrumentation__.Notify(235564)
			return errors.New("status entry unexpectedly changed during update")
		} else {
			__antithesis_instrumentation__.Notify(235567)
			if err != nil {
				__antithesis_instrumentation__.Notify(235568)
				return err
			} else {
				__antithesis_instrumentation__.Notify(235569)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(235570)
		if err := db.PutInline(ctx, key, &nodeStatus); err != nil {
			__antithesis_instrumentation__.Notify(235571)
			return err
		} else {
			__antithesis_instrumentation__.Notify(235572)
		}
	}
	__antithesis_instrumentation__.Notify(235554)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(235573)
		statusJSON, err := json.Marshal(&nodeStatus)
		if err != nil {
			__antithesis_instrumentation__.Notify(235575)
			log.Errorf(ctx, "error marshaling nodeStatus to json: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(235576)
		}
		__antithesis_instrumentation__.Notify(235574)
		log.Infof(ctx, "node %d status: %s", nodeStatus.Desc.NodeID, statusJSON)
	} else {
		__antithesis_instrumentation__.Notify(235577)
	}
	__antithesis_instrumentation__.Notify(235555)
	return nil
}

type registryRecorder struct {
	registry       *metric.Registry
	format         string
	source         string
	timestampNanos int64
}

func extractValue(name string, mtr interface{}, fn func(string, float64)) error {
	__antithesis_instrumentation__.Notify(235578)

	type (
		float64Valuer   interface{ Value() float64 }
		int64Valuer     interface{ Value() int64 }
		int64Counter    interface{ Count() int64 }
		histogramValuer interface {
			Windowed() (*hdrhistogram.Histogram, time.Duration)
		}
	)
	switch mtr := mtr.(type) {
	case float64:
		__antithesis_instrumentation__.Notify(235580)
		fn(name, mtr)
	case float64Valuer:
		__antithesis_instrumentation__.Notify(235581)
		fn(name, mtr.Value())
	case int64Valuer:
		__antithesis_instrumentation__.Notify(235582)
		fn(name, float64(mtr.Value()))
	case int64Counter:
		__antithesis_instrumentation__.Notify(235583)
		fn(name, float64(mtr.Count()))
	case histogramValuer:
		__antithesis_instrumentation__.Notify(235584)

		curr, _ := mtr.Windowed()
		for _, pt := range recordHistogramQuantiles {
			__antithesis_instrumentation__.Notify(235587)
			fn(name+pt.suffix, float64(curr.ValueAtQuantile(pt.quantile)))
		}
		__antithesis_instrumentation__.Notify(235585)
		fn(name+"-count", float64(curr.TotalCount()))
	default:
		__antithesis_instrumentation__.Notify(235586)
		return errors.Errorf("cannot extract value for type %T", mtr)
	}
	__antithesis_instrumentation__.Notify(235579)
	return nil
}

func eachRecordableValue(reg *metric.Registry, fn func(string, float64)) {
	__antithesis_instrumentation__.Notify(235588)
	reg.Each(func(name string, mtr interface{}) {
		__antithesis_instrumentation__.Notify(235589)
		if err := extractValue(name, mtr, fn); err != nil {
			__antithesis_instrumentation__.Notify(235590)
			log.Warningf(context.TODO(), "%v", err)
			return
		} else {
			__antithesis_instrumentation__.Notify(235591)
		}
	})
}

func (rr registryRecorder) record(dest *[]tspb.TimeSeriesData) {
	__antithesis_instrumentation__.Notify(235592)
	eachRecordableValue(rr.registry, func(name string, val float64) {
		__antithesis_instrumentation__.Notify(235593)
		*dest = append(*dest, tspb.TimeSeriesData{
			Name:   fmt.Sprintf(rr.format, name),
			Source: rr.source,
			Datapoints: []tspb.TimeSeriesDatapoint{
				{
					TimestampNanos: rr.timestampNanos,
					Value:          val,
				},
			},
		})
	})
}

func GetTotalMemory(ctx context.Context) (int64, error) {
	__antithesis_instrumentation__.Notify(235594)
	memory, warning, err := GetTotalMemoryWithoutLogging()
	if err != nil {
		__antithesis_instrumentation__.Notify(235597)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(235598)
	}
	__antithesis_instrumentation__.Notify(235595)
	if warning != "" {
		__antithesis_instrumentation__.Notify(235599)
		log.Infof(ctx, "%s", warning)
	} else {
		__antithesis_instrumentation__.Notify(235600)
	}
	__antithesis_instrumentation__.Notify(235596)
	return memory, nil
}

func GetTotalMemoryWithoutLogging() (int64, string, error) {
	__antithesis_instrumentation__.Notify(235601)
	totalMem, err := func() (int64, error) {
		__antithesis_instrumentation__.Notify(235608)
		mem := gosigar.Mem{}
		if err := mem.Get(); err != nil {
			__antithesis_instrumentation__.Notify(235611)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(235612)
		}
		__antithesis_instrumentation__.Notify(235609)
		if mem.Total > math.MaxInt64 {
			__antithesis_instrumentation__.Notify(235613)
			return 0, fmt.Errorf("inferred memory size %s exceeds maximum supported memory size %s",
				humanize.IBytes(mem.Total), humanize.Bytes(math.MaxInt64))
		} else {
			__antithesis_instrumentation__.Notify(235614)
		}
		__antithesis_instrumentation__.Notify(235610)
		return int64(mem.Total), nil
	}()
	__antithesis_instrumentation__.Notify(235602)
	if err != nil {
		__antithesis_instrumentation__.Notify(235615)
		return 0, "", err
	} else {
		__antithesis_instrumentation__.Notify(235616)
	}
	__antithesis_instrumentation__.Notify(235603)
	checkTotal := func(x int64, warning string) (int64, string, error) {
		__antithesis_instrumentation__.Notify(235617)
		if x <= 0 {
			__antithesis_instrumentation__.Notify(235619)

			return 0, warning, fmt.Errorf("inferred memory size %d is suspicious, considering invalid", x)
		} else {
			__antithesis_instrumentation__.Notify(235620)
		}
		__antithesis_instrumentation__.Notify(235618)
		return x, warning, nil
	}
	__antithesis_instrumentation__.Notify(235604)
	if runtime.GOOS != "linux" {
		__antithesis_instrumentation__.Notify(235621)
		return checkTotal(totalMem, "")
	} else {
		__antithesis_instrumentation__.Notify(235622)
	}
	__antithesis_instrumentation__.Notify(235605)
	cgAvlMem, warning, err := cgroups.GetMemoryLimit()
	if err != nil {
		__antithesis_instrumentation__.Notify(235623)
		return checkTotal(totalMem,
			fmt.Sprintf("available memory from cgroups is unsupported, using system memory %s instead: %v",
				humanizeutil.IBytes(totalMem), err))
	} else {
		__antithesis_instrumentation__.Notify(235624)
	}
	__antithesis_instrumentation__.Notify(235606)
	if cgAvlMem == 0 || func() bool {
		__antithesis_instrumentation__.Notify(235625)
		return (totalMem > 0 && func() bool {
			__antithesis_instrumentation__.Notify(235626)
			return cgAvlMem > totalMem == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(235627)
		return checkTotal(totalMem,
			fmt.Sprintf("available memory from cgroups (%s) is unsupported, using system memory %s instead: %s",
				humanize.IBytes(uint64(cgAvlMem)), humanizeutil.IBytes(totalMem), warning))
	} else {
		__antithesis_instrumentation__.Notify(235628)
	}
	__antithesis_instrumentation__.Notify(235607)
	return checkTotal(cgAvlMem, "")
}
