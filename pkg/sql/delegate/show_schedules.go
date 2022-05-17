package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

const commandColumn = `crdb_internal.pb_to_json('cockroach.jobs.jobspb.ExecutionArguments', execution_args, false, true)->'args'`

func (d *delegator) delegateShowSchedules(n *tree.ShowSchedules) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465761)
	sqltelemetry.IncrementShowCounter(sqltelemetry.Schedules)

	columnExprs := []string{
		"schedule_id as id",
		"schedule_name as label",
		"(CASE WHEN next_run IS NULL THEN 'PAUSED' ELSE 'ACTIVE' END) AS schedule_status",
		"next_run",
		"crdb_internal.pb_to_json('cockroach.jobs.jobspb.ScheduleState', schedule_state)->>'status' as state",
		"(CASE WHEN schedule_expr IS NULL THEN 'NEVER' ELSE schedule_expr END) as recurrence",
		fmt.Sprintf(`(
SELECT count(*) FROM system.jobs
WHERE status='%s' AND created_by_type='%s' AND created_by_id=schedule_id
) AS jobsRunning`, jobs.StatusRunning, jobs.CreatedByScheduledJobs),
		"owner",
		"created",
	}

	var whereExprs []string

	switch n.WhichSchedules {
	case tree.PausedSchedules:
		__antithesis_instrumentation__.Notify(465766)
		whereExprs = append(whereExprs, "next_run IS NULL")
	case tree.ActiveSchedules:
		__antithesis_instrumentation__.Notify(465767)
		whereExprs = append(whereExprs, "next_run IS NOT NULL")
	default:
		__antithesis_instrumentation__.Notify(465768)
	}
	__antithesis_instrumentation__.Notify(465762)

	switch n.ExecutorType {
	case tree.ScheduledBackupExecutor:
		__antithesis_instrumentation__.Notify(465769)
		whereExprs = append(whereExprs, fmt.Sprintf(
			"executor_type = '%s'", tree.ScheduledBackupExecutor.InternalName()))
		columnExprs = append(columnExprs, fmt.Sprintf(
			"%s->>'backup_statement' AS command", commandColumn))
	case tree.ScheduledSQLStatsCompactionExecutor:
		__antithesis_instrumentation__.Notify(465770)
		whereExprs = append(whereExprs, fmt.Sprintf(
			"executor_type = '%s'", tree.ScheduledSQLStatsCompactionExecutor.InternalName()))
	default:
		__antithesis_instrumentation__.Notify(465771)

		columnExprs = append(columnExprs, fmt.Sprintf("%s #-'{@type}' AS command", commandColumn))
	}
	__antithesis_instrumentation__.Notify(465763)

	if n.ScheduleID != nil {
		__antithesis_instrumentation__.Notify(465772)
		whereExprs = append(whereExprs,
			fmt.Sprintf("schedule_id=(%s)", tree.AsString(n.ScheduleID)))
	} else {
		__antithesis_instrumentation__.Notify(465773)
	}
	__antithesis_instrumentation__.Notify(465764)

	var whereClause string
	if len(whereExprs) > 0 {
		__antithesis_instrumentation__.Notify(465774)
		whereClause = fmt.Sprintf("WHERE (%s)", strings.Join(whereExprs, " AND "))
	} else {
		__antithesis_instrumentation__.Notify(465775)
	}
	__antithesis_instrumentation__.Notify(465765)
	return parse(fmt.Sprintf(
		"SELECT %s FROM system.scheduled_jobs %s",
		strings.Join(columnExprs, ","),
		whereClause,
	))
}
