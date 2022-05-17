package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

var showCreateTableColumns = colinfo.ResultColumns{
	{Name: "schedule_id", Typ: types.Int},
	{Name: "create_statement", Typ: types.String},
}

const (
	scheduleID = iota
	createStmt
)

func loadSchedules(params runParams, n *tree.ShowCreateSchedules) ([]*jobs.ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(623108)
	env := JobSchedulerEnv(params.ExecCfg())
	var schedules []*jobs.ScheduledJob
	var rows []tree.Datums
	var cols colinfo.ResultColumns

	if n.ScheduleID != nil {
		__antithesis_instrumentation__.Notify(623111)
		sjID, err := strconv.Atoi(n.ScheduleID.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(623114)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623115)
		}
		__antithesis_instrumentation__.Notify(623112)

		datums, columns, err := params.ExecCfg().InternalExecutor.QueryRowExWithCols(
			params.ctx,
			"load-schedules",
			params.EvalContext().Txn, sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			fmt.Sprintf("SELECT * FROM %s WHERE schedule_id = $1", env.ScheduledJobsTableName()),
			tree.NewDInt(tree.DInt(sjID)))
		if err != nil {
			__antithesis_instrumentation__.Notify(623116)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623117)
		}
		__antithesis_instrumentation__.Notify(623113)
		rows = append(rows, datums)
		cols = columns
	} else {
		__antithesis_instrumentation__.Notify(623118)
		datums, columns, err := params.ExecCfg().InternalExecutor.QueryBufferedExWithCols(
			params.ctx,
			"load-schedules",
			params.EvalContext().Txn, sessiondata.InternalExecutorOverride{User: security.RootUserName()},
			fmt.Sprintf("SELECT * FROM %s", env.ScheduledJobsTableName()))
		if err != nil {
			__antithesis_instrumentation__.Notify(623120)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623121)
		}
		__antithesis_instrumentation__.Notify(623119)
		rows = append(rows, datums...)
		cols = columns
	}
	__antithesis_instrumentation__.Notify(623109)

	for _, row := range rows {
		__antithesis_instrumentation__.Notify(623122)
		schedule := jobs.NewScheduledJob(env)
		if err := schedule.InitFromDatums(row, cols); err != nil {
			__antithesis_instrumentation__.Notify(623124)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(623125)
		}
		__antithesis_instrumentation__.Notify(623123)
		schedules = append(schedules, schedule)
	}
	__antithesis_instrumentation__.Notify(623110)
	return schedules, nil
}

func (p *planner) ShowCreateSchedule(
	ctx context.Context, n *tree.ShowCreateSchedules,
) (planNode, error) {
	__antithesis_instrumentation__.Notify(623126)

	if userIsAdmin, err := p.UserHasAdminRole(ctx, p.User()); err != nil {
		__antithesis_instrumentation__.Notify(623128)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(623129)
		if !userIsAdmin {
			__antithesis_instrumentation__.Notify(623130)
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege,
				"user %s does not have admin role", p.User())
		} else {
			__antithesis_instrumentation__.Notify(623131)
		}
	}
	__antithesis_instrumentation__.Notify(623127)

	sqltelemetry.IncrementShowCounter(sqltelemetry.CreateSchedule)

	return &delayedNode{
		name:    fmt.Sprintf("SHOW CREATE SCHEDULE %d", n.ScheduleID),
		columns: showCreateTableColumns,
		constructor: func(ctx context.Context, p *planner) (planNode, error) {
			__antithesis_instrumentation__.Notify(623132)
			scheduledJobs, err := loadSchedules(
				runParams{ctx: ctx, p: p, extendedEvalCtx: &p.extendedEvalCtx}, n)
			if err != nil {
				__antithesis_instrumentation__.Notify(623136)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(623137)
			}
			__antithesis_instrumentation__.Notify(623133)

			var rows []tree.Datums
			for _, sj := range scheduledJobs {
				__antithesis_instrumentation__.Notify(623138)
				ex, err := jobs.GetScheduledJobExecutor(sj.ExecutorType())
				if err != nil {
					__antithesis_instrumentation__.Notify(623141)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623142)
				}
				__antithesis_instrumentation__.Notify(623139)

				createStmtStr, err := ex.GetCreateScheduleStatement(
					ctx,
					scheduledjobs.ProdJobSchedulerEnv,
					p.Txn(),
					p.Descriptors(),
					sj,
					p.ExecCfg().InternalExecutor,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(623143)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623144)
				}
				__antithesis_instrumentation__.Notify(623140)

				row := tree.Datums{
					scheduleID: tree.NewDInt(tree.DInt(sj.ScheduleID())),
					createStmt: tree.NewDString(createStmtStr),
				}
				rows = append(rows, row)
			}
			__antithesis_instrumentation__.Notify(623134)

			v := p.newContainerValuesNode(showCreateTableColumns, len(rows))
			for _, row := range rows {
				__antithesis_instrumentation__.Notify(623145)
				if _, err := v.rows.AddRow(ctx, row); err != nil {
					__antithesis_instrumentation__.Notify(623146)
					v.Close(ctx)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(623147)
				}
			}
			__antithesis_instrumentation__.Notify(623135)
			return v, nil
		},
	}, nil
}
