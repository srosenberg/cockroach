package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/server/telemetry"

func ScheduledBackupControlCounter(desiredStatus string) telemetry.Counter {
	__antithesis_instrumentation__.Notify(625825)
	return telemetry.GetCounter("sql.backup.scheduled.job.control." + desiredStatus)
}
