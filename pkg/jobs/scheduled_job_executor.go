package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type ScheduledJobExecutor interface {
	ExecuteJob(
		ctx context.Context,
		cfg *scheduledjobs.JobExecutionConfig,
		env scheduledjobs.JobSchedulerEnv,
		schedule *ScheduledJob,
		txn *kv.Txn,
	) error

	NotifyJobTermination(
		ctx context.Context,
		jobID jobspb.JobID,
		jobStatus Status,
		details jobspb.Details,
		env scheduledjobs.JobSchedulerEnv,
		schedule *ScheduledJob,
		ex sqlutil.InternalExecutor,
		txn *kv.Txn,
	) error

	Metrics() metric.Struct

	GetCreateScheduleStatement(
		ctx context.Context,
		env scheduledjobs.JobSchedulerEnv,
		txn *kv.Txn,
		descsCol *descs.Collection,
		sj *ScheduledJob,
		ex sqlutil.InternalExecutor,
	) (string, error)
}

type ScheduledJobController interface {
	OnDrop(
		ctx context.Context,
		scheduleControllerEnv scheduledjobs.ScheduleControllerEnv,
		env scheduledjobs.JobSchedulerEnv,
		schedule *ScheduledJob,
		txn *kv.Txn,
		descsCol *descs.Collection,
	) error
}

type ScheduledJobExecutorFactory = func() (ScheduledJobExecutor, error)

var executorRegistry struct {
	syncutil.Mutex
	factories map[string]ScheduledJobExecutorFactory
	executors map[string]ScheduledJobExecutor
}

func RegisterScheduledJobExecutorFactory(name string, factory ScheduledJobExecutorFactory) {
	__antithesis_instrumentation__.Notify(84876)
	executorRegistry.Lock()
	defer executorRegistry.Unlock()
	if executorRegistry.factories == nil {
		__antithesis_instrumentation__.Notify(84879)
		executorRegistry.factories = make(map[string]ScheduledJobExecutorFactory)
	} else {
		__antithesis_instrumentation__.Notify(84880)
	}
	__antithesis_instrumentation__.Notify(84877)

	if _, ok := executorRegistry.factories[name]; ok {
		__antithesis_instrumentation__.Notify(84881)
		panic("executor " + name + " already registered")
	} else {
		__antithesis_instrumentation__.Notify(84882)
	}
	__antithesis_instrumentation__.Notify(84878)
	executorRegistry.factories[name] = factory
}

func newScheduledJobExecutorLocked(name string) (ScheduledJobExecutor, error) {
	__antithesis_instrumentation__.Notify(84883)
	if factory, ok := executorRegistry.factories[name]; ok {
		__antithesis_instrumentation__.Notify(84885)
		return factory()
	} else {
		__antithesis_instrumentation__.Notify(84886)
	}
	__antithesis_instrumentation__.Notify(84884)
	return nil, errors.Newf("executor %q is not registered", name)
}

func GetScheduledJobExecutor(name string) (ScheduledJobExecutor, error) {
	__antithesis_instrumentation__.Notify(84887)
	executorRegistry.Lock()
	defer executorRegistry.Unlock()
	return getScheduledJobExecutorLocked(name)
}

func getScheduledJobExecutorLocked(name string) (ScheduledJobExecutor, error) {
	__antithesis_instrumentation__.Notify(84888)
	if executorRegistry.executors == nil {
		__antithesis_instrumentation__.Notify(84892)
		executorRegistry.executors = make(map[string]ScheduledJobExecutor)
	} else {
		__antithesis_instrumentation__.Notify(84893)
	}
	__antithesis_instrumentation__.Notify(84889)
	if ex, ok := executorRegistry.executors[name]; ok {
		__antithesis_instrumentation__.Notify(84894)
		return ex, nil
	} else {
		__antithesis_instrumentation__.Notify(84895)
	}
	__antithesis_instrumentation__.Notify(84890)
	ex, err := newScheduledJobExecutorLocked(name)
	if err != nil {
		__antithesis_instrumentation__.Notify(84896)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(84897)
	}
	__antithesis_instrumentation__.Notify(84891)
	executorRegistry.executors[name] = ex
	return ex, nil
}

func RegisterExecutorsMetrics(registry *metric.Registry) error {
	__antithesis_instrumentation__.Notify(84898)
	executorRegistry.Lock()
	defer executorRegistry.Unlock()

	for executorType := range executorRegistry.factories {
		__antithesis_instrumentation__.Notify(84900)
		ex, err := getScheduledJobExecutorLocked(executorType)
		if err != nil {
			__antithesis_instrumentation__.Notify(84902)
			return err
		} else {
			__antithesis_instrumentation__.Notify(84903)
		}
		__antithesis_instrumentation__.Notify(84901)
		if m := ex.Metrics(); m != nil {
			__antithesis_instrumentation__.Notify(84904)
			registry.AddMetricStruct(m)
		} else {
			__antithesis_instrumentation__.Notify(84905)
		}
	}
	__antithesis_instrumentation__.Notify(84899)

	return nil
}

func DefaultHandleFailedRun(schedule *ScheduledJob, fmtOrMsg string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(84906)
	switch schedule.ScheduleDetails().OnError {
	case jobspb.ScheduleDetails_RETRY_SOON:
		__antithesis_instrumentation__.Notify(84907)
		schedule.SetScheduleStatus("retrying: "+fmtOrMsg, args...)
		schedule.SetNextRun(schedule.env.Now().Add(retryFailedJobAfter))
	case jobspb.ScheduleDetails_PAUSE_SCHED:
		__antithesis_instrumentation__.Notify(84908)
		schedule.Pause()
		schedule.SetScheduleStatus("schedule paused: "+fmtOrMsg, args...)
	case jobspb.ScheduleDetails_RETRY_SCHED:
		__antithesis_instrumentation__.Notify(84909)
		schedule.SetScheduleStatus("reschedule: "+fmtOrMsg, args...)
	default:
		__antithesis_instrumentation__.Notify(84910)
	}
}

func NotifyJobTermination(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	jobID jobspb.JobID,
	jobStatus Status,
	jobDetails jobspb.Details,
	scheduleID int64,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(84911)
	if env == nil {
		__antithesis_instrumentation__.Notify(84916)
		env = scheduledjobs.ProdJobSchedulerEnv
	} else {
		__antithesis_instrumentation__.Notify(84917)
	}
	__antithesis_instrumentation__.Notify(84912)

	schedule, err := LoadScheduledJob(ctx, env, scheduleID, ex, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(84918)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84919)
	}
	__antithesis_instrumentation__.Notify(84913)
	executor, err := GetScheduledJobExecutor(schedule.ExecutorType())
	if err != nil {
		__antithesis_instrumentation__.Notify(84920)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84921)
	}
	__antithesis_instrumentation__.Notify(84914)

	err = executor.NotifyJobTermination(ctx, jobID, jobStatus, jobDetails, env, schedule, ex, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(84922)
		return err
	} else {
		__antithesis_instrumentation__.Notify(84923)
	}
	__antithesis_instrumentation__.Notify(84915)

	return schedule.Update(ctx, ex, txn)
}
