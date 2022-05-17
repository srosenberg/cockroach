package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
)

func (d *delegator) delegateShowJobs(n *tree.ShowJobs) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465639)
	if n.Schedules != nil {
		__antithesis_instrumentation__.Notify(465643)

		return parse(fmt.Sprintf(`
SHOW JOBS SELECT id FROM system.jobs WHERE created_by_type='%s' and created_by_id IN (%s)
`, jobs.CreatedByScheduledJobs, n.Schedules.String()),
		)
	} else {
		__antithesis_instrumentation__.Notify(465644)
	}
	__antithesis_instrumentation__.Notify(465640)

	sqltelemetry.IncrementShowCounter(sqltelemetry.Jobs)

	const (
		selectClause = `
SELECT job_id, job_type, description, statement, user_name, status,
       running_status, created, started, finished, modified,
       fraction_completed, error, coordinator_id, trace_id, last_run,
       next_run, num_runs, execution_errors
  FROM crdb_internal.jobs`
	)
	var typePredicate, whereClause, orderbyClause string
	if n.Jobs == nil {
		__antithesis_instrumentation__.Notify(465645)

		{
			__antithesis_instrumentation__.Notify(465647)

			var predicate strings.Builder
			if n.Automatic {
				__antithesis_instrumentation__.Notify(465650)
				predicate.WriteString("job_type IN (")
			} else {
				__antithesis_instrumentation__.Notify(465651)
				predicate.WriteString("job_type IS NULL OR job_type NOT IN (")
			}
			__antithesis_instrumentation__.Notify(465648)
			for i, jobType := range jobspb.AutomaticJobTypes {
				__antithesis_instrumentation__.Notify(465652)
				if i != 0 {
					__antithesis_instrumentation__.Notify(465654)
					predicate.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(465655)
				}
				__antithesis_instrumentation__.Notify(465653)
				predicate.WriteByte('\'')
				predicate.WriteString(jobType.String())
				predicate.WriteByte('\'')
			}
			__antithesis_instrumentation__.Notify(465649)
			predicate.WriteByte(')')
			typePredicate = predicate.String()
		}
		__antithesis_instrumentation__.Notify(465646)

		whereClause = fmt.Sprintf(
			`WHERE %s AND (finished IS NULL OR finished > now() - '12h':::interval)`, typePredicate)

		orderbyClause = `ORDER BY COALESCE(finished, now()) DESC, started DESC`
	} else {
		__antithesis_instrumentation__.Notify(465656)

		whereClause = fmt.Sprintf(`WHERE job_id in (%s)`, n.Jobs.String())
	}
	__antithesis_instrumentation__.Notify(465641)

	sqlStmt := fmt.Sprintf("%s %s %s", selectClause, whereClause, orderbyClause)
	if n.Block {
		__antithesis_instrumentation__.Notify(465657)
		sqlStmt = fmt.Sprintf(
			`
    WITH jobs AS (SELECT * FROM [%s]),
       sleep_and_restart_if_unfinished AS (
              SELECT IF(pg_sleep(1), crdb_internal.force_retry('24h'), 1)
                     = 0 AS timed_out
                FROM (SELECT job_id FROM jobs WHERE finished IS NULL LIMIT 1)
             ),
       fail_if_slept_too_long AS (
                SELECT crdb_internal.force_error('55000', 'timed out waiting for jobs')
                  FROM sleep_and_restart_if_unfinished
                 WHERE timed_out
              )
SELECT *
  FROM jobs
 WHERE NOT EXISTS(SELECT * FROM fail_if_slept_too_long)`, sqlStmt)
	} else {
		__antithesis_instrumentation__.Notify(465658)
	}
	__antithesis_instrumentation__.Notify(465642)
	return parse(sqlStmt)
}
