package delegate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

type ControlJobsDelegate struct {
	Command tree.JobCommand

	Schedules *tree.Select
	Type      string
}

var protobufNameForType = map[string]string{
	"changefeed": "changefeed",
	"backup":     "backup",
	"import":     "import",
	"restore":    "restore",
}

func (d *delegator) delegateJobControl(stmt ControlJobsDelegate) (tree.Statement, error) {
	__antithesis_instrumentation__.Notify(465434)

	validStartStatusForCommand := map[tree.JobCommand][]jobs.Status{
		tree.PauseJob:  {jobs.StatusPending, jobs.StatusRunning, jobs.StatusReverting},
		tree.ResumeJob: {jobs.StatusPaused},
		tree.CancelJob: {jobs.StatusPending, jobs.StatusRunning, jobs.StatusPaused},
	}

	var filterExprs []string
	var filterClause string
	if statuses, ok := validStartStatusForCommand[stmt.Command]; ok {
		__antithesis_instrumentation__.Notify(465438)
		for _, status := range statuses {
			__antithesis_instrumentation__.Notify(465440)
			filterExprs = append(filterExprs, fmt.Sprintf("'%s'", status))
		}
		__antithesis_instrumentation__.Notify(465439)
		filterClause = fmt.Sprint(strings.Join(filterExprs, ", "))
	} else {
		__antithesis_instrumentation__.Notify(465441)
		return nil, errors.New("unexpected Command encountered in schedule job control")
	}
	__antithesis_instrumentation__.Notify(465435)

	if stmt.Schedules != nil {
		__antithesis_instrumentation__.Notify(465442)
		return parse(fmt.Sprintf(`%s JOBS SELECT id FROM system.jobs WHERE jobs.created_by_type = '%s' 
AND jobs.status IN (%s) AND jobs.created_by_id IN (%s)`,
			tree.JobCommandToStatement[stmt.Command], jobs.CreatedByScheduledJobs, filterClause,
			stmt.Schedules))
	} else {
		__antithesis_instrumentation__.Notify(465443)
	}
	__antithesis_instrumentation__.Notify(465436)

	if stmt.Type != "" {
		__antithesis_instrumentation__.Notify(465444)
		if _, ok := protobufNameForType[stmt.Type]; !ok {
			__antithesis_instrumentation__.Notify(465446)
			return nil, errors.New("Unsupported job type")
		} else {
			__antithesis_instrumentation__.Notify(465447)
		}
		__antithesis_instrumentation__.Notify(465445)
		queryStrFormat := `%s JOBS (
  SELECT id
  FROM (
        SELECT id,
               status,
               (
                crdb_internal.pb_to_json(
                  'cockroach.sql.jobs.jobspb.Payload',
                  payload, false, true
                )->'%s'
               ) IS NOT NULL AS correct_type
          FROM system.jobs
         WHERE status IN (%s)
       )
  WHERE correct_type
);`
		return parse(fmt.Sprintf(queryStrFormat, tree.JobCommandToStatement[stmt.Command], protobufNameForType[stmt.Type], filterClause))
	} else {
		__antithesis_instrumentation__.Notify(465448)
	}
	__antithesis_instrumentation__.Notify(465437)

	return nil, errors.AssertionFailedf("Missing Schedules or Type clause in delegate parameters")
}
