package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	replStatsRotateInterval = 5 * time.Minute
	decayFactor             = 0.8

	MinStatsDuration = 5 * time.Second
)

var AddSSTableRequestSizeFactor = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.replica_stats.addsst_request_size_factor",
	"the divisor that is applied to addsstable request sizes, then recorded in a leaseholders QPS; 0 means all requests are treated as cost 1",

	50000,
).WithPublic()

type localityOracle func(roachpb.NodeID) string

type perLocalityCounts map[string]float64

type replicaStats struct {
	clock           *hlc.Clock
	getNodeLocality localityOracle

	mu struct {
		syncutil.Mutex
		idx        int
		requests   [6]perLocalityCounts
		lastRotate time.Time
		lastReset  time.Time

		avgQPSForTesting float64
	}
}

func newReplicaStats(clock *hlc.Clock, getNodeLocality localityOracle) *replicaStats {
	__antithesis_instrumentation__.Notify(120779)
	rs := &replicaStats{
		clock:           clock,
		getNodeLocality: getNodeLocality,
	}
	rs.mu.requests[rs.mu.idx] = make(perLocalityCounts)
	rs.mu.lastRotate = timeutil.Unix(0, rs.clock.PhysicalNow())
	rs.mu.lastReset = rs.mu.lastRotate
	return rs
}

func (rs *replicaStats) splitRequestCounts(other *replicaStats) {
	__antithesis_instrumentation__.Notify(120780)
	other.mu.Lock()
	defer other.mu.Unlock()
	rs.mu.Lock()
	defer rs.mu.Unlock()

	other.mu.idx = rs.mu.idx
	other.mu.lastRotate = rs.mu.lastRotate
	other.mu.lastReset = rs.mu.lastReset

	for i := range rs.mu.requests {
		__antithesis_instrumentation__.Notify(120781)
		if rs.mu.requests[i] == nil {
			__antithesis_instrumentation__.Notify(120783)
			other.mu.requests[i] = nil
			continue
		} else {
			__antithesis_instrumentation__.Notify(120784)
		}
		__antithesis_instrumentation__.Notify(120782)
		other.mu.requests[i] = make(perLocalityCounts)
		for k := range rs.mu.requests[i] {
			__antithesis_instrumentation__.Notify(120785)
			newVal := rs.mu.requests[i][k] / 2.0
			rs.mu.requests[i][k] = newVal
			other.mu.requests[i][k] = newVal
		}
	}
}

func (rs *replicaStats) recordCount(count float64, nodeID roachpb.NodeID) {
	__antithesis_instrumentation__.Notify(120786)
	var locality string
	if rs.getNodeLocality != nil {
		__antithesis_instrumentation__.Notify(120788)
		locality = rs.getNodeLocality(nodeID)
	} else {
		__antithesis_instrumentation__.Notify(120789)
	}
	__antithesis_instrumentation__.Notify(120787)
	now := timeutil.Unix(0, rs.clock.PhysicalNow())

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.maybeRotateLocked(now)
	rs.mu.requests[rs.mu.idx][locality] += count
}

func (rs *replicaStats) maybeRotateLocked(now time.Time) {
	__antithesis_instrumentation__.Notify(120790)
	if now.Sub(rs.mu.lastRotate) >= replStatsRotateInterval {
		__antithesis_instrumentation__.Notify(120791)
		rs.rotateLocked()
		rs.mu.lastRotate = now
	} else {
		__antithesis_instrumentation__.Notify(120792)
	}
}

func (rs *replicaStats) rotateLocked() {
	__antithesis_instrumentation__.Notify(120793)
	rs.mu.idx = (rs.mu.idx + 1) % len(rs.mu.requests)
	rs.mu.requests[rs.mu.idx] = make(perLocalityCounts)
}

func (rs *replicaStats) perLocalityDecayingQPS() (perLocalityCounts, time.Duration) {
	__antithesis_instrumentation__.Notify(120794)
	now := timeutil.Unix(0, rs.clock.PhysicalNow())

	rs.mu.Lock()
	defer rs.mu.Unlock()

	rs.maybeRotateLocked(now)

	timeSinceRotate := now.Sub(rs.mu.lastRotate)
	fractionOfRotation := float64(timeSinceRotate) / float64(replStatsRotateInterval)

	counts := make(perLocalityCounts)
	var duration time.Duration
	for i := range rs.mu.requests {
		__antithesis_instrumentation__.Notify(120797)

		requestsIdx := (rs.mu.idx + len(rs.mu.requests) - i) % len(rs.mu.requests)
		if cur := rs.mu.requests[requestsIdx]; cur != nil {
			__antithesis_instrumentation__.Notify(120798)
			decay := math.Pow(decayFactor, float64(i)+fractionOfRotation)
			if i == 0 {
				__antithesis_instrumentation__.Notify(120800)
				duration += time.Duration(float64(timeSinceRotate) * decay)
			} else {
				__antithesis_instrumentation__.Notify(120801)
				duration += time.Duration(float64(replStatsRotateInterval) * decay)
			}
			__antithesis_instrumentation__.Notify(120799)
			for k, v := range cur {
				__antithesis_instrumentation__.Notify(120802)
				counts[k] += v * decay
			}
		} else {
			__antithesis_instrumentation__.Notify(120803)
		}
	}
	__antithesis_instrumentation__.Notify(120795)

	if duration.Seconds() > 0 {
		__antithesis_instrumentation__.Notify(120804)
		for k := range counts {
			__antithesis_instrumentation__.Notify(120805)
			counts[k] = counts[k] / duration.Seconds()
		}
	} else {
		__antithesis_instrumentation__.Notify(120806)
	}
	__antithesis_instrumentation__.Notify(120796)
	return counts, now.Sub(rs.mu.lastReset)
}

func (rs *replicaStats) sumQueriesLocked() (float64, int) {
	__antithesis_instrumentation__.Notify(120807)
	var sum float64
	var windowsUsed int
	for i := range rs.mu.requests {
		__antithesis_instrumentation__.Notify(120809)

		requestsIdx := (rs.mu.idx + len(rs.mu.requests) - i) % len(rs.mu.requests)
		if cur := rs.mu.requests[requestsIdx]; cur != nil {
			__antithesis_instrumentation__.Notify(120810)
			windowsUsed++
			for _, v := range cur {
				__antithesis_instrumentation__.Notify(120811)
				sum += v
			}
		} else {
			__antithesis_instrumentation__.Notify(120812)
		}
	}
	__antithesis_instrumentation__.Notify(120808)
	return sum, windowsUsed
}

func (rs *replicaStats) avgQPS() (float64, time.Duration) {
	__antithesis_instrumentation__.Notify(120813)
	now := timeutil.Unix(0, rs.clock.PhysicalNow())

	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.mu.avgQPSForTesting != 0 {
		__antithesis_instrumentation__.Notify(120817)
		return rs.mu.avgQPSForTesting, 0
	} else {
		__antithesis_instrumentation__.Notify(120818)
	}
	__antithesis_instrumentation__.Notify(120814)

	rs.maybeRotateLocked(now)

	sum, windowsUsed := rs.sumQueriesLocked()
	if windowsUsed <= 0 {
		__antithesis_instrumentation__.Notify(120819)
		return 0, 0
	} else {
		__antithesis_instrumentation__.Notify(120820)
	}
	__antithesis_instrumentation__.Notify(120815)
	duration := now.Sub(rs.mu.lastRotate) + time.Duration(windowsUsed-1)*replStatsRotateInterval
	if duration == 0 {
		__antithesis_instrumentation__.Notify(120821)
		return 0, 0
	} else {
		__antithesis_instrumentation__.Notify(120822)
	}
	__antithesis_instrumentation__.Notify(120816)
	return sum / duration.Seconds(), duration
}

func (rs *replicaStats) resetRequestCounts() {
	__antithesis_instrumentation__.Notify(120823)
	rs.mu.Lock()
	defer rs.mu.Unlock()

	for i := range rs.mu.requests {
		__antithesis_instrumentation__.Notify(120825)
		rs.mu.requests[i] = nil
	}
	__antithesis_instrumentation__.Notify(120824)
	rs.mu.requests[rs.mu.idx] = make(perLocalityCounts)
	rs.mu.lastRotate = timeutil.Unix(0, rs.clock.PhysicalNow())
	rs.mu.lastReset = rs.mu.lastRotate
}

func (rs *replicaStats) setAvgQPSForTesting(qps float64) {
	__antithesis_instrumentation__.Notify(120826)
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.mu.avgQPSForTesting = qps
}
