package cdcutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type Throttler struct {
	name           string
	messageLimiter *quotapool.RateLimiter
	byteLimiter    *quotapool.RateLimiter
	flushLimiter   *quotapool.RateLimiter
	metrics        *Metrics
}

func (t *Throttler) AcquireMessageQuota(ctx context.Context, sz int) error {
	__antithesis_instrumentation__.Notify(15309)
	if t.messageLimiter.AdmitN(1) && func() bool {
		__antithesis_instrumentation__.Notify(15312)
		return t.byteLimiter.AdmitN(int64(sz)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(15313)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15314)
	}
	__antithesis_instrumentation__.Notify(15310)

	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, fmt.Sprintf("quota-wait-%s", t.name))
	defer span.Finish()

	if err := waitQuota(ctx, 1, t.messageLimiter, t.metrics.MessagesPushbackNanos); err != nil {
		__antithesis_instrumentation__.Notify(15315)
		return err
	} else {
		__antithesis_instrumentation__.Notify(15316)
	}
	__antithesis_instrumentation__.Notify(15311)
	return waitQuota(ctx, int64(sz), t.byteLimiter, t.metrics.BytesPushbackNanos)
}

func (t *Throttler) AcquireFlushQuota(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(15317)
	if t.flushLimiter.AdmitN(1) {
		__antithesis_instrumentation__.Notify(15319)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(15320)
	}
	__antithesis_instrumentation__.Notify(15318)

	var span *tracing.Span
	ctx, span = tracing.ChildSpan(ctx, fmt.Sprintf("quota-wait-flush-%s", t.name))
	defer span.Finish()
	return waitQuota(ctx, 1, t.flushLimiter, t.metrics.FlushPushbackNanos)
}

func (t *Throttler) updateConfig(config changefeedbase.SinkThrottleConfig) {
	__antithesis_instrumentation__.Notify(15321)
	setLimits := func(rl *quotapool.RateLimiter, rate, burst float64) {
		__antithesis_instrumentation__.Notify(15323)

		rateBudget := quotapool.Limit(math.MaxInt64)
		if rate > 0 {
			__antithesis_instrumentation__.Notify(15326)
			rateBudget = quotapool.Limit(rate)
		} else {
			__antithesis_instrumentation__.Notify(15327)
		}
		__antithesis_instrumentation__.Notify(15324)

		burstBudget := int64(burst)
		if burst < rate {
			__antithesis_instrumentation__.Notify(15328)
			burstBudget = int64(rate)
		} else {
			__antithesis_instrumentation__.Notify(15329)
		}
		__antithesis_instrumentation__.Notify(15325)
		rl.UpdateLimit(rateBudget, burstBudget)
	}
	__antithesis_instrumentation__.Notify(15322)

	setLimits(t.messageLimiter, config.MessageRate, config.MessageBurst)
	setLimits(t.byteLimiter, config.ByteRate, config.ByteBurst)
	setLimits(t.flushLimiter, config.FlushRate, config.FlushBurst)
}

func NewThrottler(name string, config changefeedbase.SinkThrottleConfig, m *Metrics) *Throttler {
	__antithesis_instrumentation__.Notify(15330)
	logSlowAcquisition := quotapool.OnSlowAcquisition(500*time.Millisecond, quotapool.LogSlowAcquisition)
	t := &Throttler{
		name: name,
		messageLimiter: quotapool.NewRateLimiter(
			fmt.Sprintf("%s-messages", name), 0, 0, logSlowAcquisition,
		),
		byteLimiter: quotapool.NewRateLimiter(
			fmt.Sprintf("%s-bytes", name), 0, 0, logSlowAcquisition,
		),
		flushLimiter: quotapool.NewRateLimiter(
			fmt.Sprintf("%s-flushes", name), 0, 0, logSlowAcquisition,
		),
		metrics: m,
	}
	t.updateConfig(config)
	return t
}

var nodeSinkThrottle = struct {
	sync.Once
	*Throttler
}{}

func NodeLevelThrottler(sv *settings.Values, metrics *Metrics) *Throttler {
	__antithesis_instrumentation__.Notify(15331)
	getConfig := func() (config changefeedbase.SinkThrottleConfig) {
		__antithesis_instrumentation__.Notify(15334)
		configStr := changefeedbase.NodeSinkThrottleConfig.Get(sv)
		if configStr != "" {
			__antithesis_instrumentation__.Notify(15336)
			if err := json.Unmarshal([]byte(configStr), &config); err != nil {
				__antithesis_instrumentation__.Notify(15337)
				log.Errorf(context.Background(),
					"failed to parse node throttle config %q: err=%v; throttling disabled", configStr, err)
			} else {
				__antithesis_instrumentation__.Notify(15338)
			}
		} else {
			__antithesis_instrumentation__.Notify(15339)
		}
		__antithesis_instrumentation__.Notify(15335)
		return
	}
	__antithesis_instrumentation__.Notify(15332)

	nodeSinkThrottle.Do(func() {
		__antithesis_instrumentation__.Notify(15340)
		if nodeSinkThrottle.Throttler != nil {
			__antithesis_instrumentation__.Notify(15342)
			panic("unexpected state")
		} else {
			__antithesis_instrumentation__.Notify(15343)
		}
		__antithesis_instrumentation__.Notify(15341)
		nodeSinkThrottle.Throttler = NewThrottler("cf.node.throttle", getConfig(), metrics)

		changefeedbase.NodeSinkThrottleConfig.SetOnChange(sv, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(15344)
			nodeSinkThrottle.Throttler.updateConfig(getConfig())
		})
	})
	__antithesis_instrumentation__.Notify(15333)

	return nodeSinkThrottle.Throttler
}

type Metrics struct {
	BytesPushbackNanos    *metric.Counter
	MessagesPushbackNanos *metric.Counter
	FlushPushbackNanos    *metric.Counter
}

func MakeMetrics(histogramWindow time.Duration) Metrics {
	__antithesis_instrumentation__.Notify(15345)
	makeMetric := func(n string) metric.Metadata {
		__antithesis_instrumentation__.Notify(15347)
		return metric.Metadata{
			Name:        fmt.Sprintf("changefeed.%s.messages_pushback_nanos", n),
			Help:        fmt.Sprintf("Total time spent throttled for %s quota", n),
			Measurement: "Nanoseconds",
			Unit:        metric.Unit_NANOSECONDS,
		}
	}
	__antithesis_instrumentation__.Notify(15346)

	return Metrics{
		BytesPushbackNanos:    metric.NewCounter(makeMetric("bytes")),
		MessagesPushbackNanos: metric.NewCounter(makeMetric("messages")),
		FlushPushbackNanos:    metric.NewCounter(makeMetric("flush")),
	}
}

var _ metric.Struct = (*Metrics)(nil)

func (m Metrics) MetricStruct() { __antithesis_instrumentation__.Notify(15348) }

func waitQuota(
	ctx context.Context, n int64, limit *quotapool.RateLimiter, c *metric.Counter,
) error {
	__antithesis_instrumentation__.Notify(15349)
	start := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(15351)
		c.Inc(int64(timeutil.Since(start)))
	}()
	__antithesis_instrumentation__.Notify(15350)
	return limit.WaitN(ctx, n)
}
