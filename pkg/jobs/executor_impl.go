package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
	"github.com/gogo/protobuf/types"
)

const InlineExecutorName = "inline"

type inlineScheduledJobExecutor struct{}

var _ ScheduledJobExecutor = &inlineScheduledJobExecutor{}

const retryFailedJobAfter = time.Minute

func (e *inlineScheduledJobExecutor) ExecuteJob(
	ctx context.Context,
	cfg *scheduledjobs.JobExecutionConfig,
	_ scheduledjobs.JobSchedulerEnv,
	schedule *ScheduledJob,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(70246)
	sqlArgs := &jobspb.SqlStatementExecutionArg{}

	if err := types.UnmarshalAny(schedule.ExecutionArgs().Args, sqlArgs); err != nil {
		__antithesis_instrumentation__.Notify(70249)
		return errors.Wrapf(err, "expected SqlStatementExecutionArg")
	} else {
		__antithesis_instrumentation__.Notify(70250)
	}
	__antithesis_instrumentation__.Notify(70247)

	_, err := cfg.InternalExecutor.ExecEx(ctx, "inline-exec", txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		sqlArgs.Statement,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(70251)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70252)
	}
	__antithesis_instrumentation__.Notify(70248)

	return nil
}

func (e *inlineScheduledJobExecutor) NotifyJobTermination(
	ctx context.Context,
	jobID jobspb.JobID,
	jobStatus Status,
	_ jobspb.Details,
	env scheduledjobs.JobSchedulerEnv,
	schedule *ScheduledJob,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(70253)

	if jobStatus == StatusFailed {
		__antithesis_instrumentation__.Notify(70255)
		DefaultHandleFailedRun(schedule, "job %d failed", jobID)
	} else {
		__antithesis_instrumentation__.Notify(70256)
	}
	__antithesis_instrumentation__.Notify(70254)
	return nil
}

func (e *inlineScheduledJobExecutor) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(70257)
	return nil
}

func (e *inlineScheduledJobExecutor) GetCreateScheduleStatement(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	descsCol *descs.Collection,
	sj *ScheduledJob,
	ex sqlutil.InternalExecutor,
) (string, error) {
	__antithesis_instrumentation__.Notify(70258)
	return "", errors.AssertionFailedf("unimplemented method: 'GetCreateScheduleStatement'")
}

func init() {
	RegisterScheduledJobExecutorFactory(
		InlineExecutorName,
		func() (ScheduledJobExecutor, error) {
			return &inlineScheduledJobExecutor{}, nil
		})
}
