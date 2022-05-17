package jobstest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/systemschema"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type EnvTablesType bool

const UseTestTables EnvTablesType = false

const UseSystemTables EnvTablesType = true

func NewJobSchedulerTestEnv(
	whichTables EnvTablesType, t time.Time, allowedExecutors ...tree.ScheduledJobExecutorType,
) *JobSchedulerTestEnv {
	__antithesis_instrumentation__.Notify(84176)
	var env *JobSchedulerTestEnv
	if whichTables == UseTestTables {
		__antithesis_instrumentation__.Notify(84179)
		env = &JobSchedulerTestEnv{
			scheduledJobsTableName: "defaultdb.scheduled_jobs",
			jobsTableName:          "defaultdb.system_jobs",
		}
	} else {
		__antithesis_instrumentation__.Notify(84180)
		env = &JobSchedulerTestEnv{
			scheduledJobsTableName: "system.scheduled_jobs",
			jobsTableName:          "system.jobs",
		}
	}
	__antithesis_instrumentation__.Notify(84177)
	env.mu.now = t
	if len(allowedExecutors) > 0 {
		__antithesis_instrumentation__.Notify(84181)
		env.allowedExecutors = make(map[string]struct{}, len(allowedExecutors))
		for _, e := range allowedExecutors {
			__antithesis_instrumentation__.Notify(84182)
			env.allowedExecutors[e.InternalName()] = struct{}{}
		}
	} else {
		__antithesis_instrumentation__.Notify(84183)
	}
	__antithesis_instrumentation__.Notify(84178)

	return env
}

type JobSchedulerTestEnv struct {
	scheduledJobsTableName string
	jobsTableName          string
	allowedExecutors       map[string]struct{}
	mu                     struct {
		syncutil.Mutex
		now time.Time
	}
}

var _ scheduledjobs.JobSchedulerEnv = &JobSchedulerTestEnv{}

func (e *JobSchedulerTestEnv) ScheduledJobsTableName() string {
	__antithesis_instrumentation__.Notify(84184)
	return e.scheduledJobsTableName
}

func (e *JobSchedulerTestEnv) SystemJobsTableName() string {
	__antithesis_instrumentation__.Notify(84185)
	return e.jobsTableName
}

func (e *JobSchedulerTestEnv) Now() time.Time {
	__antithesis_instrumentation__.Notify(84186)
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.mu.now
}

func (e *JobSchedulerTestEnv) AdvanceTime(d time.Duration) {
	__antithesis_instrumentation__.Notify(84187)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mu.now = e.mu.now.Add(d)
}

func (e *JobSchedulerTestEnv) SetTime(t time.Time) {
	__antithesis_instrumentation__.Notify(84188)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.mu.now = t
}

const timestampTZLayout = "2006-01-02 15:04:05.000000"

func (e *JobSchedulerTestEnv) NowExpr() string {
	__antithesis_instrumentation__.Notify(84189)
	e.mu.Lock()
	defer e.mu.Unlock()
	return fmt.Sprintf("TIMESTAMPTZ '%s'", e.mu.now.Format(timestampTZLayout))
}

func (e *JobSchedulerTestEnv) IsExecutorEnabled(name string) bool {
	__antithesis_instrumentation__.Notify(84190)
	enabled := e.allowedExecutors == nil
	if !enabled {
		__antithesis_instrumentation__.Notify(84192)
		_, enabled = e.allowedExecutors[name]
	} else {
		__antithesis_instrumentation__.Notify(84193)
	}
	__antithesis_instrumentation__.Notify(84191)
	return enabled
}

func GetScheduledJobsTableSchema(env scheduledjobs.JobSchedulerEnv) string {
	__antithesis_instrumentation__.Notify(84194)
	if env.ScheduledJobsTableName() == "system.jobs" {
		__antithesis_instrumentation__.Notify(84196)
		return systemschema.ScheduledJobsTableSchema
	} else {
		__antithesis_instrumentation__.Notify(84197)
	}
	__antithesis_instrumentation__.Notify(84195)
	return strings.Replace(systemschema.ScheduledJobsTableSchema,
		"system.scheduled_jobs", env.ScheduledJobsTableName(), 1)
}

func GetJobsTableSchema(env scheduledjobs.JobSchedulerEnv) string {
	__antithesis_instrumentation__.Notify(84198)
	if env.SystemJobsTableName() == "system.jobs" {
		__antithesis_instrumentation__.Notify(84200)
		return systemschema.JobsTableSchema
	} else {
		__antithesis_instrumentation__.Notify(84201)
	}
	__antithesis_instrumentation__.Notify(84199)
	return strings.Replace(systemschema.JobsTableSchema,
		"system.jobs", env.SystemJobsTableName(), 1)
}
