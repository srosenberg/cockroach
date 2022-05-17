package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
	"github.com/robfig/cron/v3"
)

var SQLStatsFlushInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.stats.flush.interval",
	"the interval at which SQL execution statistics are flushed to disk, "+
		"this value must be less than or equal to sql.stats.aggregation.interval",
	time.Minute*10,
	settings.NonNegativeDurationWithMaximum(time.Hour*24),
).WithPublic()

var MinimumInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.stats.flush.minimum_interval",
	"the minimum interval that SQL stats can be flushes to disk. If a "+
		"flush operation starts within less than the minimum interval, the flush "+
		"operation will be aborted",
	0,
	settings.NonNegativeDuration,
)

var DiscardInMemoryStatsWhenFlushDisabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.stats.flush.force_cleanup.enabled",
	"if set, older SQL stats are discarded periodically when flushing to "+
		"persisted tables is disabled",
	false,
)

var SQLStatsFlushEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.stats.flush.enabled",
	"if set, SQL execution statistics are periodically flushed to disk",
	true,
).WithPublic()

var SQLStatsFlushJitter = settings.RegisterFloatSetting(
	settings.TenantWritable,
	"sql.stats.flush.jitter",
	"jitter fraction on the duration between sql stats flushes",
	0.15,
	func(f float64) error {
		__antithesis_instrumentation__.Notify(624432)
		if f < 0 || func() bool {
			__antithesis_instrumentation__.Notify(624434)
			return f > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(624435)
			return errors.Newf("%f is not in [0, 1]", f)
		} else {
			__antithesis_instrumentation__.Notify(624436)
		}
		__antithesis_instrumentation__.Notify(624433)
		return nil
	},
)

var SQLStatsMaxPersistedRows = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.stats.persisted_rows.max",
	"maximum number of rows of statement and transaction"+
		" statistics that will be persisted in the system tables",
	1000000,
).WithPublic()

var SQLStatsCleanupRecurrence = settings.RegisterValidatedStringSetting(
	settings.TenantWritable,
	"sql.stats.cleanup.recurrence",
	"cron-tab recurrence for SQL Stats cleanup job",
	"@hourly",
	func(_ *settings.Values, s string) error {
		__antithesis_instrumentation__.Notify(624437)
		if _, err := cron.ParseStandard(s); err != nil {
			__antithesis_instrumentation__.Notify(624439)
			return errors.Wrap(err, "invalid cron expression")
		} else {
			__antithesis_instrumentation__.Notify(624440)
		}
		__antithesis_instrumentation__.Notify(624438)
		return nil
	},
).WithPublic()

var SQLStatsAggregationInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"sql.stats.aggregation.interval",
	"the interval at which we aggregate SQL execution statistics upon flush, "+
		"this value must be greater than or equal to sql.stats.flush.interval",
	time.Hour,
	settings.NonNegativeDurationWithMaximum(time.Hour*24),
)

var CompactionJobRowsToDeletePerTxn = settings.RegisterIntSetting(
	settings.TenantWritable,
	"sql.stats.cleanup.rows_to_delete_per_txn",
	"number of rows the compaction job deletes from system table per iteration",
	1024,
	settings.NonNegativeInt,
)
