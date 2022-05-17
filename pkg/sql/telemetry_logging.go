package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const defaultMaxEventFrequency = 8

var telemetryMaxEventFrequency = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.telemetry.query_sampling.max_event_frequency",
	"the max event frequency at which we sample queries for telemetry, "+
		"note that this value shares a log-line limit of 10 logs per second on the "+
		"telemetry pipeline with all other telemetry events",
	defaultMaxEventFrequency,
	settings.NonNegativeInt,
)

type TelemetryLoggingMetrics struct {
	mu struct {
		syncutil.RWMutex

		lastEmittedTime time.Time
	}
	Knobs *TelemetryLoggingTestingKnobs

	skippedQueryCount uint64
}

type TelemetryLoggingTestingKnobs struct {
	getTimeNow func() time.Time
}

func (*TelemetryLoggingTestingKnobs) ModuleTestingKnobs() {
	__antithesis_instrumentation__.Notify(627831)
}

func (t *TelemetryLoggingMetrics) timeNow() time.Time {
	__antithesis_instrumentation__.Notify(627832)
	if t.Knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(627834)
		return t.Knobs.getTimeNow != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(627835)
		return t.Knobs.getTimeNow()
	} else {
		__antithesis_instrumentation__.Notify(627836)
	}
	__antithesis_instrumentation__.Notify(627833)
	return timeutil.Now()
}

func (t *TelemetryLoggingMetrics) maybeUpdateLastEmittedTime(
	newTime time.Time, requiredSecondsElapsed float64,
) bool {
	__antithesis_instrumentation__.Notify(627837)
	t.mu.Lock()
	defer t.mu.Unlock()

	lastEmittedTime := t.mu.lastEmittedTime

	if float64(newTime.Sub(lastEmittedTime))*1e-9 >= requiredSecondsElapsed {
		__antithesis_instrumentation__.Notify(627839)
		t.mu.lastEmittedTime = newTime
		return true
	} else {
		__antithesis_instrumentation__.Notify(627840)
	}
	__antithesis_instrumentation__.Notify(627838)

	return false
}

func (t *TelemetryLoggingMetrics) resetSkippedQueryCount() (res uint64) {
	__antithesis_instrumentation__.Notify(627841)
	return atomic.SwapUint64(&t.skippedQueryCount, 0)
}

func (t *TelemetryLoggingMetrics) incSkippedQueryCount() {
	__antithesis_instrumentation__.Notify(627842)
	atomic.AddUint64(&t.skippedQueryCount, 1)
}
