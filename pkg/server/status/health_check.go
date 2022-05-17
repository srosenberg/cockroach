package status

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/status/statuspb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type threshold struct {
	gauge bool
	min   int64
}

var (
	counterZero = threshold{}
	gaugeZero   = threshold{gauge: true}
)

var trackedMetrics = map[string]threshold{

	"ranges.unavailable":          gaugeZero,
	"ranges.underreplicated":      gaugeZero,
	"requests.backpressure.split": gaugeZero,
	"requests.slow.latch":         gaugeZero,
	"requests.slow.lease":         gaugeZero,
	"requests.slow.raft":          gaugeZero,

	"raft.process.logcommit.latency-90": {gauge: true, min: int64(100 * time.Millisecond)},
	"round-trip-latency-p90":            {gauge: true, min: int64(time.Second)},

	"liveness.heartbeatfailures": counterZero,
	"timeseries.write.errors":    counterZero,

	"queue.raftsnapshot.pending": {gauge: true, min: 100},
}

type metricsMap map[roachpb.StoreID]map[string]float64

func (d metricsMap) update(tracked map[string]threshold, m metricsMap) metricsMap {
	__antithesis_instrumentation__.Notify(235436)
	out := metricsMap{}
	for storeID := range m {
		__antithesis_instrumentation__.Notify(235438)
		for name, threshold := range tracked {
			__antithesis_instrumentation__.Notify(235439)
			val, ok := m[storeID][name]
			if !ok {
				__antithesis_instrumentation__.Notify(235442)
				continue
			} else {
				__antithesis_instrumentation__.Notify(235443)
			}
			__antithesis_instrumentation__.Notify(235440)

			if !threshold.gauge {
				__antithesis_instrumentation__.Notify(235444)
				prevVal, havePrev := d[storeID][name]
				if d[storeID] == nil {
					__antithesis_instrumentation__.Notify(235446)
					d[storeID] = map[string]float64{}
				} else {
					__antithesis_instrumentation__.Notify(235447)
				}
				__antithesis_instrumentation__.Notify(235445)
				d[storeID][name] = val
				if havePrev {
					__antithesis_instrumentation__.Notify(235448)
					val -= prevVal
				} else {
					__antithesis_instrumentation__.Notify(235449)

					val = 0
				}
			} else {
				__antithesis_instrumentation__.Notify(235450)
			}
			__antithesis_instrumentation__.Notify(235441)

			if val > float64(threshold.min) {
				__antithesis_instrumentation__.Notify(235451)
				if out[storeID] == nil {
					__antithesis_instrumentation__.Notify(235453)
					out[storeID] = map[string]float64{}
				} else {
					__antithesis_instrumentation__.Notify(235454)
				}
				__antithesis_instrumentation__.Notify(235452)
				out[storeID][name] = val
			} else {
				__antithesis_instrumentation__.Notify(235455)
			}
		}
	}
	__antithesis_instrumentation__.Notify(235437)
	return out
}

type HealthChecker struct {
	mu struct {
		syncutil.Mutex
		metricsMap
	}
	tracked map[string]threshold
}

func NewHealthChecker(trackedMetrics map[string]threshold) *HealthChecker {
	__antithesis_instrumentation__.Notify(235456)
	h := &HealthChecker{tracked: trackedMetrics}
	h.mu.metricsMap = metricsMap{}
	return h
}

func (h *HealthChecker) CheckHealth(
	ctx context.Context, nodeStatus statuspb.NodeStatus,
) statuspb.HealthCheckResult {
	__antithesis_instrumentation__.Notify(235457)
	h.mu.Lock()
	defer h.mu.Unlock()

	var alerts []statuspb.HealthAlert

	m := map[roachpb.StoreID]map[string]float64{
		0: nodeStatus.Metrics,
	}
	for _, storeStatus := range nodeStatus.StoreStatuses {
		__antithesis_instrumentation__.Notify(235460)
		m[storeStatus.Desc.StoreID] = storeStatus.Metrics
	}
	__antithesis_instrumentation__.Notify(235458)

	diffs := h.mu.update(h.tracked, m)

	for storeID, storeDiff := range diffs {
		__antithesis_instrumentation__.Notify(235461)
		for name, value := range storeDiff {
			__antithesis_instrumentation__.Notify(235462)
			alerts = append(alerts, statuspb.HealthAlert{
				StoreID:     storeID,
				Category:    statuspb.HealthAlert_METRICS,
				Description: name,
				Value:       value,
			})
		}
	}
	__antithesis_instrumentation__.Notify(235459)

	return statuspb.HealthCheckResult{Alerts: alerts}
}
