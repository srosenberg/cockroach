package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/cdcutils"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/schemafeed"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/metric/aggmetric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var enableSLIMetrics = envutil.EnvOrDefaultBool(
	"COCKROACH_EXPERIMENTAL_ENABLE_PER_CHANGEFEED_METRICS", false)

const maxSLIScopeNameLen = 128

const defaultSLIScope = "default"

type AggMetrics struct {
	EmittedMessages       *aggmetric.AggCounter
	MessageSize           *aggmetric.AggHistogram
	EmittedBytes          *aggmetric.AggCounter
	FlushedBytes          *aggmetric.AggCounter
	BatchHistNanos        *aggmetric.AggHistogram
	Flushes               *aggmetric.AggCounter
	FlushHistNanos        *aggmetric.AggHistogram
	CommitLatency         *aggmetric.AggHistogram
	BackfillCount         *aggmetric.AggGauge
	BackfillPendingRanges *aggmetric.AggGauge
	ErrorRetries          *aggmetric.AggCounter
	AdmitLatency          *aggmetric.AggHistogram
	RunningCount          *aggmetric.AggGauge

	mu struct {
		syncutil.Mutex
		sliMetrics map[string]*sliMetrics
	}
}

func (a *AggMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(17480) }

type sliMetrics struct {
	EmittedMessages       *aggmetric.Counter
	MessageSize           *aggmetric.Histogram
	EmittedBytes          *aggmetric.Counter
	FlushedBytes          *aggmetric.Counter
	BatchHistNanos        *aggmetric.Histogram
	Flushes               *aggmetric.Counter
	FlushHistNanos        *aggmetric.Histogram
	CommitLatency         *aggmetric.Histogram
	ErrorRetries          *aggmetric.Counter
	AdmitLatency          *aggmetric.Histogram
	BackfillCount         *aggmetric.Gauge
	BackfillPendingRanges *aggmetric.Gauge
	RunningCount          *aggmetric.Gauge
}

const sinkDoesNotCompress = -1

type recordOneMessageCallback func(mvcc hlc.Timestamp, bytes int, compressedBytes int)

func (m *sliMetrics) recordOneMessage() recordOneMessageCallback {
	__antithesis_instrumentation__.Notify(17481)
	if m == nil {
		__antithesis_instrumentation__.Notify(17483)
		return func(mvcc hlc.Timestamp, bytes int, compressedBytes int) { __antithesis_instrumentation__.Notify(17484) }
	} else {
		__antithesis_instrumentation__.Notify(17485)
	}
	__antithesis_instrumentation__.Notify(17482)

	start := timeutil.Now()
	return func(mvcc hlc.Timestamp, bytes int, compressedBytes int) {
		__antithesis_instrumentation__.Notify(17486)
		m.MessageSize.RecordValue(int64(bytes))
		m.recordEmittedBatch(start, 1, mvcc, bytes, compressedBytes)
	}
}

func (m *sliMetrics) recordMessageSize(sz int64) {
	__antithesis_instrumentation__.Notify(17487)
	if m != nil {
		__antithesis_instrumentation__.Notify(17488)
		m.MessageSize.RecordValue(sz)
	} else {
		__antithesis_instrumentation__.Notify(17489)
	}
}

func (m *sliMetrics) recordEmittedBatch(
	startTime time.Time, numMessages int, mvcc hlc.Timestamp, bytes int, compressedBytes int,
) {
	__antithesis_instrumentation__.Notify(17490)
	if m == nil {
		__antithesis_instrumentation__.Notify(17493)
		return
	} else {
		__antithesis_instrumentation__.Notify(17494)
	}
	__antithesis_instrumentation__.Notify(17491)
	emitNanos := timeutil.Since(startTime).Nanoseconds()
	m.EmittedMessages.Inc(int64(numMessages))
	m.EmittedBytes.Inc(int64(bytes))
	if compressedBytes == sinkDoesNotCompress {
		__antithesis_instrumentation__.Notify(17495)
		compressedBytes = bytes
	} else {
		__antithesis_instrumentation__.Notify(17496)
	}
	__antithesis_instrumentation__.Notify(17492)
	m.FlushedBytes.Inc(int64(compressedBytes))
	m.BatchHistNanos.RecordValue(emitNanos)
	if m.BackfillCount.Value() == 0 {
		__antithesis_instrumentation__.Notify(17497)
		m.CommitLatency.RecordValue(timeutil.Since(mvcc.GoTime()).Nanoseconds())
	} else {
		__antithesis_instrumentation__.Notify(17498)
	}
}

func (m *sliMetrics) recordResolvedCallback() func() {
	__antithesis_instrumentation__.Notify(17499)
	if m == nil {
		__antithesis_instrumentation__.Notify(17501)
		return func() { __antithesis_instrumentation__.Notify(17502) }
	} else {
		__antithesis_instrumentation__.Notify(17503)
	}
	__antithesis_instrumentation__.Notify(17500)

	start := timeutil.Now()
	return func() {
		__antithesis_instrumentation__.Notify(17504)
		emitNanos := timeutil.Since(start).Nanoseconds()
		m.EmittedMessages.Inc(1)
		m.BatchHistNanos.RecordValue(emitNanos)
	}
}

func (m *sliMetrics) recordFlushRequestCallback() func() {
	__antithesis_instrumentation__.Notify(17505)
	if m == nil {
		__antithesis_instrumentation__.Notify(17507)
		return func() { __antithesis_instrumentation__.Notify(17508) }
	} else {
		__antithesis_instrumentation__.Notify(17509)
	}
	__antithesis_instrumentation__.Notify(17506)

	start := timeutil.Now()
	return func() {
		__antithesis_instrumentation__.Notify(17510)
		flushNanos := timeutil.Since(start).Nanoseconds()
		m.Flushes.Inc(1)
		m.FlushHistNanos.RecordValue(flushNanos)
	}
}

func (m *sliMetrics) getBackfillCallback() func() func() {
	__antithesis_instrumentation__.Notify(17511)
	return func() func() {
		__antithesis_instrumentation__.Notify(17512)
		m.BackfillCount.Inc(1)
		return func() {
			__antithesis_instrumentation__.Notify(17513)
			m.BackfillCount.Dec(1)
		}
	}
}

func (m *sliMetrics) getBackfillRangeCallback() func(int64) (func(), func()) {
	__antithesis_instrumentation__.Notify(17514)
	return func(initial int64) (dec func(), clear func()) {
		__antithesis_instrumentation__.Notify(17515)
		remaining := initial
		m.BackfillPendingRanges.Inc(initial)
		dec = func() {
			__antithesis_instrumentation__.Notify(17518)
			m.BackfillPendingRanges.Dec(1)
			atomic.AddInt64(&remaining, -1)
		}
		__antithesis_instrumentation__.Notify(17516)
		clear = func() {
			__antithesis_instrumentation__.Notify(17519)
			m.BackfillPendingRanges.Dec(remaining)
			atomic.AddInt64(&remaining, -remaining)
		}
		__antithesis_instrumentation__.Notify(17517)
		return
	}
}

const (
	changefeedCheckpointHistMaxLatency = 30 * time.Second
	changefeedBatchHistMaxLatency      = 30 * time.Second
	changefeedFlushHistMaxLatency      = 1 * time.Minute
	admitLatencyMaxValue               = 1 * time.Minute
	commitLatencyMaxValue              = 10 * time.Minute
)

var (
	metaChangefeedForwardedResolvedMessages = metric.Metadata{
		Name:        "changefeed.forwarded_resolved_messages",
		Help:        "Resolved timestamps forwarded from the change aggregator to the change frontier",
		Measurement: "Messages",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedErrorRetries = metric.Metadata{
		Name:        "changefeed.error_retries",
		Help:        "Total retryable errors encountered by all changefeeds",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedFailures = metric.Metadata{
		Name:        "changefeed.failures",
		Help:        "Total number of changefeed jobs which have failed",
		Measurement: "Errors",
		Unit:        metric.Unit_COUNT,
	}

	metaEventQueueTime = metric.Metadata{
		Name:        "changefeed.queue_time_nanos",
		Help:        "Time KV event spent waiting to be processed",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}

	metaChangefeedCheckpointHistNanos = metric.Metadata{
		Name:        "changefeed.checkpoint_hist_nanos",
		Help:        "Time spent checkpointing changefeed progress",
		Measurement: "Changefeeds",
		Unit:        metric.Unit_NANOSECONDS,
	}

	metaChangefeedMaxBehindNanos = metric.Metadata{
		Name:        "changefeed.max_behind_nanos",
		Help:        "Largest commit-to-emit duration of any running feed",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}

	metaChangefeedFrontierUpdates = metric.Metadata{
		Name:        "changefeed.frontier_updates",
		Help:        "Number of change frontier updates across all feeds",
		Measurement: "Updates",
		Unit:        metric.Unit_COUNT,
	}
)

func newAggregateMetrics(histogramWindow time.Duration) *AggMetrics {
	__antithesis_instrumentation__.Notify(17520)
	metaChangefeedEmittedMessages := metric.Metadata{
		Name:        "changefeed.emitted_messages",
		Help:        "Messages emitted by all feeds",
		Measurement: "Messages",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedEmittedBytes := metric.Metadata{
		Name:        "changefeed.emitted_bytes",
		Help:        "Bytes emitted by all feeds",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
	metaChangefeedFlushedBytes := metric.Metadata{
		Name:        "changefeed.flushed_bytes",
		Help:        "Bytes emitted by all feeds; maybe different from changefeed.emitted_bytes when compression is enabled",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}
	metaChangefeedFlushes := metric.Metadata{
		Name:        "changefeed.flushes",
		Help:        "Total flushes across all feeds",
		Measurement: "Flushes",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBatchHistNanos := metric.Metadata{
		Name:        "changefeed.sink_batch_hist_nanos",
		Help:        "Time spent batched in the sink buffer before being flushed and acknowledged",
		Measurement: "Changefeeds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaChangefeedFlushHistNanos := metric.Metadata{
		Name:        "changefeed.flush_hist_nanos",
		Help:        "Time spent flushing messages across all changefeeds",
		Measurement: "Changefeeds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaCommitLatency := metric.Metadata{
		Name: "changefeed.commit_latency",
		Help: "Event commit latency: a difference between event MVCC timestamp " +
			"and the time it was acknowledged by the downstream sink.  If the sink batches events, " +
			" then the difference between the oldest event in the batch and acknowledgement is recorded; " +
			"Excludes latency during backfill",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaAdmitLatency := metric.Metadata{
		Name: "changefeed.admit_latency",
		Help: "Event admission latency: a difference between event MVCC timestamp " +
			"and the time it was admitted into changefeed pipeline; " +
			"Note: this metric includes the time spent waiting until event can be processed due " +
			"to backpressure or time spent resolving schema descriptors. " +
			"Also note, this metric excludes latency during backfill",
		Measurement: "Nanoseconds",
		Unit:        metric.Unit_NANOSECONDS,
	}
	metaChangefeedBackfillCount := metric.Metadata{
		Name:        "changefeed.backfill_count",
		Help:        "Number of changefeeds currently executing backfill",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedBackfillPendingRanges := metric.Metadata{
		Name:        "changefeed.backfill_pending_ranges",
		Help:        "Number of ranges in an ongoing backfill that are yet to be fully emitted",
		Measurement: "Count",
		Unit:        metric.Unit_COUNT,
	}
	metaChangefeedRunning := metric.Metadata{
		Name:        "changefeed.running",
		Help:        "Number of currently running changefeeds, including sinkless",
		Measurement: "Changefeeds",
		Unit:        metric.Unit_COUNT,
	}
	metaMessageSize := metric.Metadata{
		Name:        "changefeed.message_size_hist",
		Help:        "Message size histogram",
		Measurement: "Bytes",
		Unit:        metric.Unit_BYTES,
	}

	b := aggmetric.MakeBuilder("scope")
	a := &AggMetrics{
		ErrorRetries:    b.Counter(metaChangefeedErrorRetries),
		EmittedMessages: b.Counter(metaChangefeedEmittedMessages),
		MessageSize: b.Histogram(metaMessageSize,
			histogramWindow, 10<<20, 1),
		EmittedBytes: b.Counter(metaChangefeedEmittedBytes),
		FlushedBytes: b.Counter(metaChangefeedFlushedBytes),
		Flushes:      b.Counter(metaChangefeedFlushes),

		BatchHistNanos: b.Histogram(metaChangefeedBatchHistNanos,
			histogramWindow, changefeedBatchHistMaxLatency.Nanoseconds(), 1),
		FlushHistNanos: b.Histogram(metaChangefeedFlushHistNanos,
			histogramWindow, changefeedFlushHistMaxLatency.Nanoseconds(), 2),
		CommitLatency: b.Histogram(metaCommitLatency,
			histogramWindow, commitLatencyMaxValue.Nanoseconds(), 1),
		AdmitLatency: b.Histogram(metaAdmitLatency, histogramWindow,
			admitLatencyMaxValue.Nanoseconds(), 1),
		BackfillCount:         b.Gauge(metaChangefeedBackfillCount),
		BackfillPendingRanges: b.Gauge(metaChangefeedBackfillPendingRanges),
		RunningCount:          b.Gauge(metaChangefeedRunning),
	}
	a.mu.sliMetrics = make(map[string]*sliMetrics)
	_, err := a.getOrCreateScope(defaultSLIScope)
	if err != nil {
		__antithesis_instrumentation__.Notify(17522)

		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(17523)
	}
	__antithesis_instrumentation__.Notify(17521)
	return a
}

func (a *AggMetrics) getOrCreateScope(scope string) (*sliMetrics, error) {
	__antithesis_instrumentation__.Notify(17524)
	a.mu.Lock()
	defer a.mu.Unlock()

	scope = strings.TrimSpace(strings.ToLower(scope))

	if scope == "" {
		__antithesis_instrumentation__.Notify(17529)
		scope = defaultSLIScope
	} else {
		__antithesis_instrumentation__.Notify(17530)
	}
	__antithesis_instrumentation__.Notify(17525)

	if len(scope) > maxSLIScopeNameLen {
		__antithesis_instrumentation__.Notify(17531)
		return nil, pgerror.Newf(pgcode.ConfigurationLimitExceeded,
			"scope name length must be less than %d bytes", maxSLIScopeNameLen)
	} else {
		__antithesis_instrumentation__.Notify(17532)
	}
	__antithesis_instrumentation__.Notify(17526)

	if s, ok := a.mu.sliMetrics[scope]; ok {
		__antithesis_instrumentation__.Notify(17533)
		return s, nil
	} else {
		__antithesis_instrumentation__.Notify(17534)
	}
	__antithesis_instrumentation__.Notify(17527)

	if scope != defaultSLIScope {
		__antithesis_instrumentation__.Notify(17535)
		if !enableSLIMetrics {
			__antithesis_instrumentation__.Notify(17537)
			return nil, errors.WithHint(
				pgerror.Newf(pgcode.ConfigurationLimitExceeded, "cannot create metrics scope %q", scope),
				"try restarting with COCKROACH_EXPERIMENTAL_ENABLE_PER_CHANGEFEED_METRICS=true",
			)
		} else {
			__antithesis_instrumentation__.Notify(17538)
		}
		__antithesis_instrumentation__.Notify(17536)
		const failSafeMax = 1024
		if len(a.mu.sliMetrics) == failSafeMax {
			__antithesis_instrumentation__.Notify(17539)
			return nil, pgerror.Newf(pgcode.ConfigurationLimitExceeded,
				"too many metrics labels; max %d", failSafeMax)
		} else {
			__antithesis_instrumentation__.Notify(17540)
		}
	} else {
		__antithesis_instrumentation__.Notify(17541)
	}
	__antithesis_instrumentation__.Notify(17528)

	sm := &sliMetrics{
		EmittedMessages:       a.EmittedMessages.AddChild(scope),
		MessageSize:           a.MessageSize.AddChild(scope),
		EmittedBytes:          a.EmittedBytes.AddChild(scope),
		FlushedBytes:          a.FlushedBytes.AddChild(scope),
		BatchHistNanos:        a.BatchHistNanos.AddChild(scope),
		Flushes:               a.Flushes.AddChild(scope),
		FlushHistNanos:        a.FlushHistNanos.AddChild(scope),
		CommitLatency:         a.CommitLatency.AddChild(scope),
		ErrorRetries:          a.ErrorRetries.AddChild(scope),
		AdmitLatency:          a.AdmitLatency.AddChild(scope),
		BackfillCount:         a.BackfillCount.AddChild(scope),
		BackfillPendingRanges: a.BackfillPendingRanges.AddChild(scope),
		RunningCount:          a.RunningCount.AddChild(scope),
	}

	a.mu.sliMetrics[scope] = sm
	return sm, nil
}

type Metrics struct {
	AggMetrics          *AggMetrics
	KVFeedMetrics       kvevent.Metrics
	SchemaFeedMetrics   schemafeed.Metrics
	Failures            *metric.Counter
	ResolvedMessages    *metric.Counter
	QueueTimeNanos      *metric.Counter
	CheckpointHistNanos *metric.Histogram
	FrontierUpdates     *metric.Counter
	ThrottleMetrics     cdcutils.Metrics

	mu struct {
		syncutil.Mutex
		id       int
		resolved map[int]hlc.Timestamp
	}
	MaxBehindNanos *metric.Gauge
}

func (*Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(17542) }

func (m *Metrics) getSLIMetrics(scope string) (*sliMetrics, error) {
	__antithesis_instrumentation__.Notify(17543)
	return m.AggMetrics.getOrCreateScope(scope)
}

func MakeMetrics(histogramWindow time.Duration) metric.Struct {
	__antithesis_instrumentation__.Notify(17544)
	m := &Metrics{
		AggMetrics:        newAggregateMetrics(histogramWindow),
		KVFeedMetrics:     kvevent.MakeMetrics(histogramWindow),
		SchemaFeedMetrics: schemafeed.MakeMetrics(histogramWindow),
		ResolvedMessages:  metric.NewCounter(metaChangefeedForwardedResolvedMessages),
		Failures:          metric.NewCounter(metaChangefeedFailures),
		QueueTimeNanos:    metric.NewCounter(metaEventQueueTime),
		CheckpointHistNanos: metric.NewHistogram(metaChangefeedCheckpointHistNanos, histogramWindow,
			changefeedCheckpointHistMaxLatency.Nanoseconds(), 2),
		FrontierUpdates: metric.NewCounter(metaChangefeedFrontierUpdates),
		ThrottleMetrics: cdcutils.MakeMetrics(histogramWindow),
	}

	m.mu.resolved = make(map[int]hlc.Timestamp)
	m.mu.id = 1
	m.MaxBehindNanos = metric.NewFunctionalGauge(metaChangefeedMaxBehindNanos, func() int64 {
		__antithesis_instrumentation__.Notify(17546)
		now := timeutil.Now()
		var maxBehind time.Duration
		m.mu.Lock()
		for _, resolved := range m.mu.resolved {
			__antithesis_instrumentation__.Notify(17548)
			if behind := now.Sub(resolved.GoTime()); behind > maxBehind {
				__antithesis_instrumentation__.Notify(17549)
				maxBehind = behind
			} else {
				__antithesis_instrumentation__.Notify(17550)
			}
		}
		__antithesis_instrumentation__.Notify(17547)
		m.mu.Unlock()
		return maxBehind.Nanoseconds()
	})
	__antithesis_instrumentation__.Notify(17545)
	return m
}

func init() {
	jobs.MakeChangefeedMetricsHook = MakeMetrics
}
