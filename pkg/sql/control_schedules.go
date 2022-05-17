package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/errors"
)

type controlSchedulesNode struct {
	rows    planNode
	command tree.ScheduleCommand
	numRows int
}

func collectTelemetry(command tree.ScheduleCommand) {
	__antithesis_instrumentation__.Notify(460554)
	switch command {
	case tree.PauseSchedule:
		__antithesis_instrumentation__.Notify(460555)
		telemetry.Inc(sqltelemetry.ScheduledBackupControlCounter("pause"))
	case tree.ResumeSchedule:
		__antithesis_instrumentation__.Notify(460556)
		telemetry.Inc(sqltelemetry.ScheduledBackupControlCounter("resume"))
	case tree.DropSchedule:
		__antithesis_instrumentation__.Notify(460557)
		telemetry.Inc(sqltelemetry.ScheduledBackupControlCounter("drop"))
	default:
		__antithesis_instrumentation__.Notify(460558)
	}
}

func (n *controlSchedulesNode) FastPathResults() (int, bool) {
	__antithesis_instrumentation__.Notify(460559)
	return n.numRows, true
}

func JobSchedulerEnv(execCfg *ExecutorConfig) scheduledjobs.JobSchedulerEnv {
	__antithesis_instrumentation__.Notify(460560)
	if knobs, ok := execCfg.DistSQLSrv.TestingKnobs.JobsTestingKnobs.(*jobs.TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(460562)
		if knobs.JobSchedulerEnv != nil {
			__antithesis_instrumentation__.Notify(460563)
			return knobs.JobSchedulerEnv
		} else {
			__antithesis_instrumentation__.Notify(460564)
		}
	} else {
		__antithesis_instrumentation__.Notify(460565)
	}
	__antithesis_instrumentation__.Notify(460561)
	return scheduledjobs.ProdJobSchedulerEnv
}

func loadSchedule(params runParams, scheduleID tree.Datum) (*jobs.ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(460566)
	env := JobSchedulerEnv(params.ExecCfg())
	schedule := jobs.NewScheduledJob(env)

	datums, cols, err := params.ExecCfg().InternalExecutor.QueryRowExWithCols(
		params.ctx,
		"load-schedule",
		params.EvalContext().Txn, sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(
			"SELECT schedule_id, next_run, schedule_expr, executor_type, execution_args FROM %s WHERE schedule_id = $1",
			env.ScheduledJobsTableName(),
		),
		scheduleID)
	if err != nil {
		__antithesis_instrumentation__.Notify(460570)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460571)
	}
	__antithesis_instrumentation__.Notify(460567)

	if datums == nil {
		__antithesis_instrumentation__.Notify(460572)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(460573)
	}
	__antithesis_instrumentation__.Notify(460568)

	if err := schedule.InitFromDatums(datums, cols); err != nil {
		__antithesis_instrumentation__.Notify(460574)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(460575)
	}
	__antithesis_instrumentation__.Notify(460569)
	return schedule, nil
}

func updateSchedule(params runParams, schedule *jobs.ScheduledJob) error {
	__antithesis_instrumentation__.Notify(460576)
	return schedule.Update(
		params.ctx,
		params.ExecCfg().InternalExecutor,
		params.EvalContext().Txn,
	)
}

func DeleteSchedule(
	ctx context.Context, execCfg *ExecutorConfig, txn *kv.Txn, scheduleID int64,
) error {
	__antithesis_instrumentation__.Notify(460577)
	env := JobSchedulerEnv(execCfg)
	_, err := execCfg.InternalExecutor.ExecEx(
		ctx,
		"delete-schedule",
		txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		fmt.Sprintf(
			"DELETE FROM %s WHERE schedule_id = $1",
			env.ScheduledJobsTableName(),
		),
		scheduleID,
	)
	return err
}

func (n *controlSchedulesNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(460578)
	for {
		__antithesis_instrumentation__.Notify(460580)
		ok, err := n.rows.Next(params)
		if err != nil {
			__antithesis_instrumentation__.Notify(460587)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460588)
		}
		__antithesis_instrumentation__.Notify(460581)
		if !ok {
			__antithesis_instrumentation__.Notify(460589)
			break
		} else {
			__antithesis_instrumentation__.Notify(460590)
		}
		__antithesis_instrumentation__.Notify(460582)

		schedule, err := loadSchedule(params, n.rows.Values()[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(460591)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460592)
		}
		__antithesis_instrumentation__.Notify(460583)

		if schedule == nil {
			__antithesis_instrumentation__.Notify(460593)
			continue
		} else {
			__antithesis_instrumentation__.Notify(460594)
		}
		__antithesis_instrumentation__.Notify(460584)

		switch n.command {
		case tree.PauseSchedule:
			__antithesis_instrumentation__.Notify(460595)
			schedule.Pause()
			err = updateSchedule(params, schedule)
		case tree.ResumeSchedule:
			__antithesis_instrumentation__.Notify(460596)

			if schedule.IsPaused() {
				__antithesis_instrumentation__.Notify(460601)
				err = schedule.ScheduleNextRun()
				if err == nil {
					__antithesis_instrumentation__.Notify(460602)
					err = updateSchedule(params, schedule)
				} else {
					__antithesis_instrumentation__.Notify(460603)
				}
			} else {
				__antithesis_instrumentation__.Notify(460604)
			}
		case tree.DropSchedule:
			__antithesis_instrumentation__.Notify(460597)
			var ex jobs.ScheduledJobExecutor
			ex, err = jobs.GetScheduledJobExecutor(schedule.ExecutorType())
			if err != nil {
				__antithesis_instrumentation__.Notify(460605)
				return errors.Wrap(err, "failed to get scheduled job executor during drop")
			} else {
				__antithesis_instrumentation__.Notify(460606)
			}
			__antithesis_instrumentation__.Notify(460598)
			if controller, ok := ex.(jobs.ScheduledJobController); ok {
				__antithesis_instrumentation__.Notify(460607)
				scheduleControllerEnv := scheduledjobs.MakeProdScheduleControllerEnv(
					params.ExecCfg().ProtectedTimestampProvider, params.ExecCfg().InternalExecutor)
				if err := controller.OnDrop(
					params.ctx,
					scheduleControllerEnv,
					scheduledjobs.ProdJobSchedulerEnv,
					schedule,
					params.extendedEvalCtx.Txn,
					params.p.Descriptors(),
				); err != nil {
					__antithesis_instrumentation__.Notify(460608)
					return errors.Wrap(err, "failed to run OnDrop")
				} else {
					__antithesis_instrumentation__.Notify(460609)
				}
			} else {
				__antithesis_instrumentation__.Notify(460610)
			}
			__antithesis_instrumentation__.Notify(460599)
			err = DeleteSchedule(params.ctx, params.ExecCfg(), params.p.txn, schedule.ScheduleID())
		default:
			__antithesis_instrumentation__.Notify(460600)
			err = errors.AssertionFailedf("unhandled command %s", n.command)
		}
		__antithesis_instrumentation__.Notify(460585)
		collectTelemetry(n.command)

		if err != nil {
			__antithesis_instrumentation__.Notify(460611)
			return err
		} else {
			__antithesis_instrumentation__.Notify(460612)
		}
		__antithesis_instrumentation__.Notify(460586)
		n.numRows++
	}
	__antithesis_instrumentation__.Notify(460579)

	return nil
}

func (*controlSchedulesNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(460613)
	return false, nil
}

func (*controlSchedulesNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(460614)
	return nil
}

func (n *controlSchedulesNode) Close(ctx context.Context) {
	__antithesis_instrumentation__.Notify(460615)
	n.rows.Close(ctx)
}
