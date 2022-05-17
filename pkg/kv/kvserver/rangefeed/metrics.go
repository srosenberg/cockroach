package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
)

var (
	metaRangeFeedCatchUpScanNanos = metric.Metadata{
		Name:        "kv.rangefeed.catchup_scan_nanos",
		Help:        "Time spent in RangeFeed catchup scan",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaRangeFeedExhausted = metric.Metadata{
		Name:        "kv.rangefeed.budget_allocation_failed",
		Help:        "Number of times RangeFeed failed because memory budget was exceeded",
		Measurement: "Events",
		Unit:        metric.Unit_COUNT,
	}
	metaRangeFeedBudgetBlocked = metric.Metadata{
		Name:        "kv.rangefeed.budget_allocation_blocked",
		Help:        "Number of times RangeFeed waited for budget availability",
		Measurement: "Events",
		Unit:        metric.Unit_COUNT,
	}
)

type Metrics struct {
	RangeFeedCatchUpScanNanos *metric.Counter
	RangeFeedBudgetExhausted  *metric.Counter
	RangeFeedBudgetBlocked    *metric.Counter

	RangeFeedSlowClosedTimestampLogN  log.EveryN
	RangeFeedSlowClosedTimestampNudge singleflight.Group

	RangeFeedSlowClosedTimestampNudgeSem chan struct{}
}

func (*Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(113621) }

func NewMetrics() *Metrics {
	__antithesis_instrumentation__.Notify(113622)
	return &Metrics{
		RangeFeedCatchUpScanNanos:            metric.NewCounter(metaRangeFeedCatchUpScanNanos),
		RangeFeedBudgetExhausted:             metric.NewCounter(metaRangeFeedExhausted),
		RangeFeedBudgetBlocked:               metric.NewCounter(metaRangeFeedBudgetBlocked),
		RangeFeedSlowClosedTimestampLogN:     log.Every(5 * time.Second),
		RangeFeedSlowClosedTimestampNudgeSem: make(chan struct{}, 1024),
	}
}

type FeedBudgetPoolMetrics struct {
	SystemBytesCount *metric.Gauge
	SharedBytesCount *metric.Gauge
}

func (FeedBudgetPoolMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(113623) }

func NewFeedBudgetMetrics(histogramWindow time.Duration) *FeedBudgetPoolMetrics {
	__antithesis_instrumentation__.Notify(113624)
	makeMemMetricMetadata := func(name, help string) metric.Metadata {
		__antithesis_instrumentation__.Notify(113626)
		return metric.Metadata{
			Name:        "kv.rangefeed.mem_" + name,
			Help:        help,
			Measurement: "Memory",
			Unit:        metric.Unit_BYTES,
		}
	}
	__antithesis_instrumentation__.Notify(113625)

	return &FeedBudgetPoolMetrics{
		SystemBytesCount: metric.NewGauge(makeMemMetricMetadata("system",
			"Memory usage by rangefeeds on system ranges")),
		SharedBytesCount: metric.NewGauge(makeMemMetricMetadata("shared",
			"Memory usage by rangefeeds")),
	}
}
