package rpc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/VividCortex/ewma"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/montanaflynn/stats"
)

type RemoteClockMetrics struct {
	ClockOffsetMeanNanos   *metric.Gauge
	ClockOffsetStdDevNanos *metric.Gauge
	LatencyHistogramNanos  *metric.Histogram
}

const avgLatencyMeasurementAge = 20.0

var (
	metaClockOffsetMeanNanos = metric.Metadata{
		Name:        "clock-offset.meannanos",
		Help:        "Mean clock offset with other nodes",
		Measurement: "Clock Offset",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaClockOffsetStdDevNanos = metric.Metadata{
		Name:        "clock-offset.stddevnanos",
		Help:        "Stddev clock offset with other nodes",
		Measurement: "Clock Offset",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaLatencyHistogramNanos = metric.Metadata{
		Name:        "round-trip-latency",
		Help:        "Distribution of round-trip latencies with other nodes",
		Measurement: "Roundtrip Latency",
		Unit:        metric.Unit_NANOSECONDS,
	}
)

type resettingMaxTrigger bool

func (t *resettingMaxTrigger) triggers(value, resetThreshold, triggerThreshold float64) bool {
	__antithesis_instrumentation__.Notify(184194)
	if *t {
		__antithesis_instrumentation__.Notify(184196)

		if value < resetThreshold {
			__antithesis_instrumentation__.Notify(184197)
			*t = false
		} else {
			__antithesis_instrumentation__.Notify(184198)
		}
	} else {
		__antithesis_instrumentation__.Notify(184199)

		if value > triggerThreshold {
			__antithesis_instrumentation__.Notify(184200)
			*t = true
			return true
		} else {
			__antithesis_instrumentation__.Notify(184201)
		}
	}
	__antithesis_instrumentation__.Notify(184195)
	return false
}

type latencyInfo struct {
	avgNanos ewma.MovingAverage
	trigger  resettingMaxTrigger
}

type RemoteClockMonitor struct {
	clock     *hlc.Clock
	offsetTTL time.Duration

	mu struct {
		syncutil.RWMutex
		offsets      map[string]RemoteOffset
		latencyInfos map[string]*latencyInfo
	}

	metrics RemoteClockMetrics
}

func newRemoteClockMonitor(
	clock *hlc.Clock, offsetTTL time.Duration, histogramWindowInterval time.Duration,
) *RemoteClockMonitor {
	__antithesis_instrumentation__.Notify(184202)
	r := RemoteClockMonitor{
		clock:     clock,
		offsetTTL: offsetTTL,
	}
	r.mu.offsets = make(map[string]RemoteOffset)
	r.mu.latencyInfos = make(map[string]*latencyInfo)
	if histogramWindowInterval == 0 {
		__antithesis_instrumentation__.Notify(184204)
		histogramWindowInterval = time.Duration(math.MaxInt64)
	} else {
		__antithesis_instrumentation__.Notify(184205)
	}
	__antithesis_instrumentation__.Notify(184203)
	r.metrics = RemoteClockMetrics{
		ClockOffsetMeanNanos:   metric.NewGauge(metaClockOffsetMeanNanos),
		ClockOffsetStdDevNanos: metric.NewGauge(metaClockOffsetStdDevNanos),
		LatencyHistogramNanos:  metric.NewLatency(metaLatencyHistogramNanos, histogramWindowInterval),
	}
	return &r
}

func (r *RemoteClockMonitor) Metrics() *RemoteClockMetrics {
	__antithesis_instrumentation__.Notify(184206)
	return &r.metrics
}

func (r *RemoteClockMonitor) Latency(addr string) (time.Duration, bool) {
	__antithesis_instrumentation__.Notify(184207)
	r.mu.RLock()
	defer r.mu.RUnlock()
	if info, ok := r.mu.latencyInfos[addr]; ok && func() bool {
		__antithesis_instrumentation__.Notify(184209)
		return info.avgNanos.Value() != 0.0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(184210)
		return time.Duration(int64(info.avgNanos.Value())), true
	} else {
		__antithesis_instrumentation__.Notify(184211)
	}
	__antithesis_instrumentation__.Notify(184208)
	return 0, false
}

func (r *RemoteClockMonitor) AllLatencies() map[string]time.Duration {
	__antithesis_instrumentation__.Notify(184212)
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]time.Duration)
	for addr, info := range r.mu.latencyInfos {
		__antithesis_instrumentation__.Notify(184214)
		if info.avgNanos.Value() != 0.0 {
			__antithesis_instrumentation__.Notify(184215)
			result[addr] = time.Duration(int64(info.avgNanos.Value()))
		} else {
			__antithesis_instrumentation__.Notify(184216)
		}
	}
	__antithesis_instrumentation__.Notify(184213)
	return result
}

func (r *RemoteClockMonitor) UpdateOffset(
	ctx context.Context, addr string, offset RemoteOffset, roundTripLatency time.Duration,
) {
	__antithesis_instrumentation__.Notify(184217)
	emptyOffset := offset == RemoteOffset{}

	r.mu.Lock()
	defer r.mu.Unlock()

	if oldOffset, ok := r.mu.offsets[addr]; !ok {
		__antithesis_instrumentation__.Notify(184220)

		if !emptyOffset {
			__antithesis_instrumentation__.Notify(184221)
			r.mu.offsets[addr] = offset
		} else {
			__antithesis_instrumentation__.Notify(184222)
		}
	} else {
		__antithesis_instrumentation__.Notify(184223)
		if oldOffset.isStale(r.offsetTTL, r.clock.PhysicalTime()) {
			__antithesis_instrumentation__.Notify(184224)

			if !emptyOffset {
				__antithesis_instrumentation__.Notify(184225)
				r.mu.offsets[addr] = offset
			} else {
				__antithesis_instrumentation__.Notify(184226)
				delete(r.mu.offsets, addr)
			}
		} else {
			__antithesis_instrumentation__.Notify(184227)
			if offset.Uncertainty < oldOffset.Uncertainty {
				__antithesis_instrumentation__.Notify(184228)

				if !emptyOffset {
					__antithesis_instrumentation__.Notify(184229)
					r.mu.offsets[addr] = offset
				} else {
					__antithesis_instrumentation__.Notify(184230)
				}
			} else {
				__antithesis_instrumentation__.Notify(184231)
			}
		}
	}
	__antithesis_instrumentation__.Notify(184218)

	if roundTripLatency > 0 {
		__antithesis_instrumentation__.Notify(184232)
		info, ok := r.mu.latencyInfos[addr]
		if !ok {
			__antithesis_instrumentation__.Notify(184234)
			info = &latencyInfo{
				avgNanos: ewma.NewMovingAverage(avgLatencyMeasurementAge),
			}
			r.mu.latencyInfos[addr] = info
		} else {
			__antithesis_instrumentation__.Notify(184235)
		}
		__antithesis_instrumentation__.Notify(184233)

		newLatencyf := float64(roundTripLatency.Nanoseconds())
		prevAvg := info.avgNanos.Value()
		info.avgNanos.Add(newLatencyf)
		r.metrics.LatencyHistogramNanos.RecordValue(roundTripLatency.Nanoseconds())

		if newLatencyf > 1e6 && func() bool {
			__antithesis_instrumentation__.Notify(184236)
			return prevAvg > 0.0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(184237)
			return info.trigger.triggers(newLatencyf, prevAvg*1.4, prevAvg*1.5) == true
		}() == true {
			__antithesis_instrumentation__.Notify(184238)
			log.Health.Warningf(ctx, "latency jump (prev avg %.2fms, current %.2fms)",
				prevAvg/1e6, newLatencyf/1e6)
		} else {
			__antithesis_instrumentation__.Notify(184239)
		}
	} else {
		__antithesis_instrumentation__.Notify(184240)
	}
	__antithesis_instrumentation__.Notify(184219)

	if log.V(2) {
		__antithesis_instrumentation__.Notify(184241)
		log.Dev.Infof(ctx, "update offset: %s %v", addr, r.mu.offsets[addr])
	} else {
		__antithesis_instrumentation__.Notify(184242)
	}
}

func (r *RemoteClockMonitor) VerifyClockOffset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(184243)

	if maxOffset := r.clock.MaxOffset(); maxOffset != 0 {
		__antithesis_instrumentation__.Notify(184245)
		now := r.clock.PhysicalTime()

		healthyOffsetCount := 0

		r.mu.Lock()

		offsets := make(stats.Float64Data, 0, 2*len(r.mu.offsets))
		for addr, offset := range r.mu.offsets {
			__antithesis_instrumentation__.Notify(184250)
			if offset.isStale(r.offsetTTL, now) {
				__antithesis_instrumentation__.Notify(184252)
				delete(r.mu.offsets, addr)
				continue
			} else {
				__antithesis_instrumentation__.Notify(184253)
			}
			__antithesis_instrumentation__.Notify(184251)
			offsets = append(offsets, float64(offset.Offset+offset.Uncertainty))
			offsets = append(offsets, float64(offset.Offset-offset.Uncertainty))
			if offset.isHealthy(ctx, maxOffset) {
				__antithesis_instrumentation__.Notify(184254)
				healthyOffsetCount++
			} else {
				__antithesis_instrumentation__.Notify(184255)
			}
		}
		__antithesis_instrumentation__.Notify(184246)
		numClocks := len(r.mu.offsets)
		r.mu.Unlock()

		mean, err := offsets.Mean()
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(184256)
			return !errors.Is(err, stats.EmptyInput) == true
		}() == true {
			__antithesis_instrumentation__.Notify(184257)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184258)
		}
		__antithesis_instrumentation__.Notify(184247)
		stdDev, err := offsets.StandardDeviation()
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(184259)
			return !errors.Is(err, stats.EmptyInput) == true
		}() == true {
			__antithesis_instrumentation__.Notify(184260)
			return err
		} else {
			__antithesis_instrumentation__.Notify(184261)
		}
		__antithesis_instrumentation__.Notify(184248)
		r.metrics.ClockOffsetMeanNanos.Update(int64(mean))
		r.metrics.ClockOffsetStdDevNanos.Update(int64(stdDev))

		if numClocks > 0 && func() bool {
			__antithesis_instrumentation__.Notify(184262)
			return healthyOffsetCount <= numClocks/2 == true
		}() == true {
			__antithesis_instrumentation__.Notify(184263)
			return errors.Errorf(
				"clock synchronization error: this node is more than %s away from at least half of the known nodes (%d of %d are within the offset)",
				maxOffset, healthyOffsetCount, numClocks)
		} else {
			__antithesis_instrumentation__.Notify(184264)
		}
		__antithesis_instrumentation__.Notify(184249)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(184265)
			log.Dev.Infof(ctx, "%d of %d nodes are within the maximum clock offset of %s", healthyOffsetCount, numClocks, maxOffset)
		} else {
			__antithesis_instrumentation__.Notify(184266)
		}
	} else {
		__antithesis_instrumentation__.Notify(184267)
	}
	__antithesis_instrumentation__.Notify(184244)

	return nil
}

func (r RemoteOffset) isHealthy(ctx context.Context, maxOffset time.Duration) bool {
	__antithesis_instrumentation__.Notify(184268)

	toleratedOffset := maxOffset * 4 / 5

	absOffset := r.Offset
	if absOffset < 0 {
		__antithesis_instrumentation__.Notify(184270)
		absOffset = -absOffset
	} else {
		__antithesis_instrumentation__.Notify(184271)
	}
	__antithesis_instrumentation__.Notify(184269)
	switch {
	case time.Duration(absOffset-r.Uncertainty)*time.Nanosecond > toleratedOffset:
		__antithesis_instrumentation__.Notify(184272)

		return false

	case time.Duration(absOffset+r.Uncertainty)*time.Nanosecond < toleratedOffset:
		__antithesis_instrumentation__.Notify(184273)

		return true

	default:
		__antithesis_instrumentation__.Notify(184274)

		log.Health.Warningf(ctx, "uncertain remote offset %s for maximum tolerated offset %s, treating as healthy", r, toleratedOffset)
		return true
	}
}

func (r RemoteOffset) isStale(ttl time.Duration, now time.Time) bool {
	__antithesis_instrumentation__.Notify(184275)
	return r.measuredAt().Add(ttl).Before(now)
}
