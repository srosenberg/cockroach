package scheduledjobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/protectedts"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type JobSchedulerEnv interface {
	ScheduledJobsTableName() string

	SystemJobsTableName() string

	Now() time.Time

	NowExpr() string

	IsExecutorEnabled(name string) bool
}

type JobExecutionConfig struct {
	Settings         *cluster.Settings
	InternalExecutor sqlutil.InternalExecutor
	DB               *kv.DB

	TestingKnobs base.ModuleTestingKnobs

	PlanHookMaker func(opName string, tnx *kv.Txn, user security.SQLUsername) (interface{}, func())

	ShouldRunScheduler func(ctx context.Context, ts hlc.ClockTimestamp) (bool, error)
}

type prodJobSchedulerEnvImpl struct{}

var ProdJobSchedulerEnv JobSchedulerEnv = &prodJobSchedulerEnvImpl{}

func (e *prodJobSchedulerEnvImpl) ScheduledJobsTableName() string {
	__antithesis_instrumentation__.Notify(185682)
	return "system.scheduled_jobs"
}

func (e *prodJobSchedulerEnvImpl) SystemJobsTableName() string {
	__antithesis_instrumentation__.Notify(185683)
	return "system.jobs"
}

func (e *prodJobSchedulerEnvImpl) Now() time.Time {
	__antithesis_instrumentation__.Notify(185684)
	return timeutil.Now()
}

func (e *prodJobSchedulerEnvImpl) NowExpr() string {
	__antithesis_instrumentation__.Notify(185685)
	return "current_timestamp()"
}

func (e *prodJobSchedulerEnvImpl) IsExecutorEnabled(name string) bool {
	__antithesis_instrumentation__.Notify(185686)
	return true
}

type ScheduleControllerEnv interface {
	InternalExecutor() sqlutil.InternalExecutor
	PTSProvider() protectedts.Provider
}

type ProdScheduleControllerEnvImpl struct {
	pts protectedts.Provider
	ie  sqlutil.InternalExecutor
}

func MakeProdScheduleControllerEnv(
	pts protectedts.Provider, ie sqlutil.InternalExecutor,
) *ProdScheduleControllerEnvImpl {
	__antithesis_instrumentation__.Notify(185687)
	return &ProdScheduleControllerEnvImpl{pts: pts, ie: ie}
}

func (c *ProdScheduleControllerEnvImpl) InternalExecutor() sqlutil.InternalExecutor {
	__antithesis_instrumentation__.Notify(185688)
	return c.ie
}

func (c *ProdScheduleControllerEnvImpl) PTSProvider() protectedts.Provider {
	__antithesis_instrumentation__.Notify(185689)
	return c.pts
}
