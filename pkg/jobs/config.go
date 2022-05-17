package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const (
	intervalBaseSettingKey         = "jobs.registry.interval.base"
	adoptIntervalSettingKey        = "jobs.registry.interval.adopt"
	cancelIntervalSettingKey       = "jobs.registry.interval.cancel"
	gcIntervalSettingKey           = "jobs.registry.interval.gc"
	retentionTimeSettingKey        = "jobs.retention_time"
	cancelUpdateLimitKey           = "jobs.cancel_update_limit"
	retryInitialDelaySettingKey    = "jobs.registry.retry.initial_delay"
	retryMaxDelaySettingKey        = "jobs.registry.retry.max_delay"
	executionErrorsMaxEntriesKey   = "jobs.execution_errors.max_entries"
	executionErrorsMaxEntrySizeKey = "jobs.execution_errors.max_entry_size"
	debugPausePointsSettingKey     = "jobs.debug.pausepoints"
)

const (
	defaultAdoptInterval = 30 * time.Second

	defaultCancelInterval = 10 * time.Second

	defaultGcInterval = 1 * time.Hour

	defaultIntervalBase = 1.0

	defaultRetentionTime = 14 * 24 * time.Hour

	defaultCancellationsUpdateLimit int64 = 1000

	defaultRetryInitialDelay = 30 * time.Second

	defaultRetryMaxDelay = 24 * time.Hour

	defaultExecutionErrorsMaxEntries = 3

	defaultExecutionErrorsMaxEntrySize = 64 << 10
)

var (
	intervalBaseSetting = settings.RegisterFloatSetting(
		settings.TenantWritable,
		intervalBaseSettingKey,
		"the base multiplier for other intervals such as adopt, cancel, and gc",
		defaultIntervalBase,
		settings.PositiveFloat,
	)

	adoptIntervalSetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		adoptIntervalSettingKey,
		"the interval at which a node (a) claims some of the pending jobs and "+
			"(b) restart its already claimed jobs that are in running or reverting "+
			"states but are not running",
		defaultAdoptInterval,
		settings.PositiveDuration,
	)

	cancelIntervalSetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		cancelIntervalSettingKey,
		"the interval at which a node cancels the jobs belonging to the known "+
			"dead sessions",
		defaultCancelInterval,
		settings.PositiveDuration,
	)

	gcIntervalSetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		gcIntervalSettingKey,
		"the interval a node deletes expired job records that have exceeded their "+
			"retention duration",
		defaultGcInterval,
		settings.PositiveDuration,
	)

	retentionTimeSetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		retentionTimeSettingKey,
		"the amount of time to retain records for completed jobs before",
		defaultRetentionTime,
		settings.PositiveDuration,
	).WithPublic()

	cancellationsUpdateLimitSetting = settings.RegisterIntSetting(
		settings.TenantWritable,
		cancelUpdateLimitKey,
		"the number of jobs that can be updated when canceling jobs concurrently from dead sessions",
		defaultCancellationsUpdateLimit,
		settings.NonNegativeInt,
	)

	retryInitialDelaySetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		retryInitialDelaySettingKey,
		"the starting duration of exponential-backoff delay"+
			" to retry a job which encountered a retryable error or had its coordinator"+
			" fail. The delay doubles after each retry.",
		defaultRetryInitialDelay,
		settings.NonNegativeDuration,
	)

	retryMaxDelaySetting = settings.RegisterDurationSetting(
		settings.TenantWritable,
		retryMaxDelaySettingKey,
		"the maximum duration by which a job can be delayed to retry",
		defaultRetryMaxDelay,
		settings.PositiveDuration,
	)

	executionErrorsMaxEntriesSetting = settings.RegisterIntSetting(
		settings.TenantWritable,
		executionErrorsMaxEntriesKey,
		"the maximum number of retriable error entries which will be stored for introspection",
		defaultExecutionErrorsMaxEntries,
		settings.NonNegativeInt,
	)

	executionErrorsMaxEntrySize = settings.RegisterByteSizeSetting(
		settings.TenantWritable,
		executionErrorsMaxEntrySizeKey,
		"the maximum byte size of individual error entries which will be stored"+
			" for introspection",
		defaultExecutionErrorsMaxEntrySize,
		settings.NonNegativeInt,
	)

	debugPausepoints = settings.RegisterStringSetting(
		settings.TenantWritable,
		debugPausePointsSettingKey,
		"the list, comma separated, of named pausepoints currently enabled for debugging",
		"",
	)
)

func jitter(dur time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(70200)
	const jitter = 1.0 / 6.0
	jitterFraction := 1 + (2*rand.Float64()-1)*jitter
	return time.Duration(float64(dur) * jitterFraction)
}

type loopController struct {
	timer   *timeutil.Timer
	lastRun time.Time
	updated chan struct{}

	getInterval func() time.Duration
}

func makeLoopController(
	st *cluster.Settings, s *settings.DurationSetting, overrideKnob *time.Duration,
) (loopController, func()) {
	__antithesis_instrumentation__.Notify(70201)
	lc := loopController{
		timer:   timeutil.NewTimer(),
		lastRun: timeutil.Now(),
		updated: make(chan struct{}, 1),

		getInterval: func() time.Duration {
			__antithesis_instrumentation__.Notify(70204)
			if overrideKnob != nil {
				__antithesis_instrumentation__.Notify(70206)
				return *overrideKnob
			} else {
				__antithesis_instrumentation__.Notify(70207)
			}
			__antithesis_instrumentation__.Notify(70205)
			return time.Duration(intervalBaseSetting.Get(&st.SV) * float64(s.Get(&st.SV)))
		},
	}
	__antithesis_instrumentation__.Notify(70202)

	onChange := func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(70208)
		select {
		case lc.updated <- struct{}{}:
			__antithesis_instrumentation__.Notify(70209)
		default:
			__antithesis_instrumentation__.Notify(70210)
		}
	}
	__antithesis_instrumentation__.Notify(70203)

	s.SetOnChange(&st.SV, onChange)
	intervalBaseSetting.SetOnChange(&st.SV, onChange)

	lc.timer.Reset(jitter(lc.getInterval()))
	return lc, func() { __antithesis_instrumentation__.Notify(70211); lc.timer.Stop() }
}

func (lc *loopController) onUpdate() {
	__antithesis_instrumentation__.Notify(70212)
	lc.timer.Reset(timeutil.Until(lc.lastRun.Add(jitter(lc.getInterval()))))
}

func (lc *loopController) onExecute() {
	__antithesis_instrumentation__.Notify(70213)
	lc.lastRun = timeutil.Now()
	lc.timer.Reset(jitter(lc.getInterval()))
}
