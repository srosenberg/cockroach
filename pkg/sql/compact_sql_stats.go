package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
)

type sqlStatsCompactionResumer struct {
	job *jobs.Job
	st  *cluster.Settings
	sj  *jobs.ScheduledJob
}

var _ jobs.Resumer = &sqlStatsCompactionResumer{}

func (r *sqlStatsCompactionResumer) Resume(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(457084)
	log.Infof(ctx, "starting sql stats compaction job")
	p := execCtx.(JobExecContext)
	ie := p.ExecCfg().InternalExecutor
	db := p.ExecCfg().DB

	var (
		scheduledJobID int64
		err            error
	)

	if err = db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(457087)
		scheduledJobID, err = r.getScheduleID(ctx, ie, txn, scheduledjobs.ProdJobSchedulerEnv)
		if err != nil {
			__antithesis_instrumentation__.Notify(457090)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457091)
		}
		__antithesis_instrumentation__.Notify(457088)

		if scheduledJobID != jobs.InvalidScheduleID {
			__antithesis_instrumentation__.Notify(457092)
			r.sj, err = jobs.LoadScheduledJob(ctx, scheduledjobs.ProdJobSchedulerEnv, scheduledJobID, ie, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(457094)
				return err
			} else {
				__antithesis_instrumentation__.Notify(457095)
			}
			__antithesis_instrumentation__.Notify(457093)
			r.sj.SetScheduleStatus(string(jobs.StatusRunning))
			return r.sj.Update(ctx, ie, txn)
		} else {
			__antithesis_instrumentation__.Notify(457096)
		}
		__antithesis_instrumentation__.Notify(457089)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(457097)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457098)
	}
	__antithesis_instrumentation__.Notify(457085)

	statsCompactor := persistedsqlstats.NewStatsCompactor(
		r.st,
		ie,
		db,
		ie.s.ServerMetrics.StatsMetrics.SQLStatsRemovedRows,
		p.ExecCfg().SQLStatsTestingKnobs)
	if err = statsCompactor.DeleteOldestEntries(ctx); err != nil {
		__antithesis_instrumentation__.Notify(457099)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457100)
	}
	__antithesis_instrumentation__.Notify(457086)

	return r.maybeNotifyJobTerminated(
		ctx,
		ie,
		p.ExecCfg(),
		jobs.StatusSucceeded)
}

func (r *sqlStatsCompactionResumer) OnFailOrCancel(ctx context.Context, execCtx interface{}) error {
	__antithesis_instrumentation__.Notify(457101)
	p := execCtx.(JobExecContext)
	execCfg := p.ExecCfg()
	ie := execCfg.InternalExecutor
	return r.maybeNotifyJobTerminated(ctx, ie, execCfg, jobs.StatusFailed)
}

func (r *sqlStatsCompactionResumer) maybeNotifyJobTerminated(
	ctx context.Context, ie sqlutil.InternalExecutor, exec *ExecutorConfig, status jobs.Status,
) error {
	__antithesis_instrumentation__.Notify(457102)
	log.Infof(ctx, "sql stats compaction job terminated with status = %s", status)
	if r.sj != nil {
		__antithesis_instrumentation__.Notify(457104)
		env := scheduledjobs.ProdJobSchedulerEnv
		if knobs, ok := exec.DistSQLSrv.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs); ok {
			__antithesis_instrumentation__.Notify(457107)
			if knobs.JobSchedulerEnv != nil {
				__antithesis_instrumentation__.Notify(457108)
				env = knobs.JobSchedulerEnv
			} else {
				__antithesis_instrumentation__.Notify(457109)
			}
		} else {
			__antithesis_instrumentation__.Notify(457110)
		}
		__antithesis_instrumentation__.Notify(457105)
		if err := jobs.NotifyJobTermination(
			ctx, env, r.job.ID(), status, r.job.Details(), r.sj.ScheduleID(),
			ie, nil); err != nil {
			__antithesis_instrumentation__.Notify(457111)
			return err
		} else {
			__antithesis_instrumentation__.Notify(457112)
		}
		__antithesis_instrumentation__.Notify(457106)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(457113)
	}
	__antithesis_instrumentation__.Notify(457103)
	return nil
}

func (r *sqlStatsCompactionResumer) getScheduleID(
	ctx context.Context, ie sqlutil.InternalExecutor, txn *kv.Txn, env scheduledjobs.JobSchedulerEnv,
) (scheduleID int64, _ error) {
	__antithesis_instrumentation__.Notify(457114)
	row, err := ie.QueryRowEx(ctx, "lookup-sql-stats-schedule", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		fmt.Sprintf("SELECT created_by_id FROM %s WHERE id=$1 AND created_by_type=$2", env.SystemJobsTableName()),
		r.job.ID(), jobs.CreatedByScheduledJobs,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(457117)
		return jobs.InvalidScheduleID, errors.Wrap(err, "fail to look up scheduled information")
	} else {
		__antithesis_instrumentation__.Notify(457118)
	}
	__antithesis_instrumentation__.Notify(457115)

	if row == nil {
		__antithesis_instrumentation__.Notify(457119)

		return jobs.InvalidScheduleID, nil
	} else {
		__antithesis_instrumentation__.Notify(457120)
	}
	__antithesis_instrumentation__.Notify(457116)

	scheduleID = int64(tree.MustBeDInt(row[0]))
	return scheduleID, nil
}

type sqlStatsCompactionMetrics struct {
	*jobs.ExecutorMetrics
}

var _ metric.Struct = &sqlStatsCompactionMetrics{}

func (m *sqlStatsCompactionMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(457121) }

type scheduledSQLStatsCompactionExecutor struct {
	metrics sqlStatsCompactionMetrics
}

var _ jobs.ScheduledJobExecutor = &scheduledSQLStatsCompactionExecutor{}
var _ jobs.ScheduledJobController = &scheduledSQLStatsCompactionExecutor{}

func (e *scheduledSQLStatsCompactionExecutor) OnDrop(
	ctx context.Context,
	scheduleControllerEnv scheduledjobs.ScheduleControllerEnv,
	env scheduledjobs.JobSchedulerEnv,
	schedule *jobs.ScheduledJob,
	txn *kv.Txn,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(457122)
	return persistedsqlstats.ErrScheduleUndroppable
}

func (e *scheduledSQLStatsCompactionExecutor) ExecuteJob(
	ctx context.Context,
	cfg *scheduledjobs.JobExecutionConfig,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(457123)
	if err := e.createSQLStatsCompactionJob(ctx, cfg, sj, txn); err != nil {
		__antithesis_instrumentation__.Notify(457125)
		e.metrics.NumFailed.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(457126)
	}
	__antithesis_instrumentation__.Notify(457124)

	e.metrics.NumStarted.Inc(1)
	return nil
}

func (e *scheduledSQLStatsCompactionExecutor) createSQLStatsCompactionJob(
	ctx context.Context, cfg *scheduledjobs.JobExecutionConfig, sj *jobs.ScheduledJob, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(457127)
	p, cleanup := cfg.PlanHookMaker("invoke-sql-stats-compact", txn, security.NodeUserName())
	defer cleanup()

	_, err :=
		persistedsqlstats.CreateCompactionJob(ctx, &jobs.CreatedByInfo{
			ID:   sj.ScheduleID(),
			Name: jobs.CreatedByScheduledJobs,
		}, txn, p.(*planner).ExecCfg().JobRegistry)

	if err != nil {
		__antithesis_instrumentation__.Notify(457129)
		return err
	} else {
		__antithesis_instrumentation__.Notify(457130)
	}
	__antithesis_instrumentation__.Notify(457128)

	return nil
}

func (e *scheduledSQLStatsCompactionExecutor) NotifyJobTermination(
	ctx context.Context,
	jobID jobspb.JobID,
	jobStatus jobs.Status,
	details jobspb.Details,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(457131)
	if jobStatus == jobs.StatusFailed {
		__antithesis_instrumentation__.Notify(457134)
		jobs.DefaultHandleFailedRun(sj, "sql stats compaction %d failed", jobID)
		e.metrics.NumFailed.Inc(1)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(457135)
	}
	__antithesis_instrumentation__.Notify(457132)

	if jobStatus == jobs.StatusSucceeded {
		__antithesis_instrumentation__.Notify(457136)
		e.metrics.NumSucceeded.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(457137)
	}
	__antithesis_instrumentation__.Notify(457133)

	sj.SetScheduleStatus(string(jobStatus))

	return nil
}

func (e *scheduledSQLStatsCompactionExecutor) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(457138)
	return &e.metrics
}

func (e *scheduledSQLStatsCompactionExecutor) GetCreateScheduleStatement(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	descsCol *descs.Collection,
	sj *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
) (string, error) {
	__antithesis_instrumentation__.Notify(457139)
	return "SELECT crdb_internal.schedule_sql_stats_compact()", nil
}

func init() {
	jobs.RegisterConstructor(jobspb.TypeAutoSQLStatsCompaction, func(job *jobs.Job, settings *cluster.Settings) jobs.Resumer {
		return &sqlStatsCompactionResumer{
			job: job,
			st:  settings,
		}
	})

	jobs.RegisterScheduledJobExecutorFactory(
		tree.ScheduledSQLStatsCompactionExecutor.InternalName(),
		func() (jobs.ScheduledJobExecutor, error) {
			m := jobs.MakeExecutorMetrics(tree.ScheduledSQLStatsCompactionExecutor.InternalName())
			return &scheduledSQLStatsCompactionExecutor{
				metrics: sqlStatsCompactionMetrics{
					ExecutorMetrics: &m,
				},
			}, nil
		})
}
