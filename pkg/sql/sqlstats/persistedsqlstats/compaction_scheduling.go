package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

const compactionScheduleName = "sql-stats-compaction"

var ErrDuplicatedSchedules = errors.New("creating multiple sql stats compaction is disallowed")

func CreateSQLStatsCompactionScheduleIfNotYetExist(
	ctx context.Context, ie sqlutil.InternalExecutor, txn *kv.Txn, st *cluster.Settings,
) (*jobs.ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(624645)
	scheduleExists, err := checkExistingCompactionSchedule(ctx, ie, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(624651)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624652)
	}
	__antithesis_instrumentation__.Notify(624646)

	if scheduleExists {
		__antithesis_instrumentation__.Notify(624653)
		return nil, ErrDuplicatedSchedules
	} else {
		__antithesis_instrumentation__.Notify(624654)
	}
	__antithesis_instrumentation__.Notify(624647)

	compactionSchedule := jobs.NewScheduledJob(scheduledjobs.ProdJobSchedulerEnv)

	schedule := SQLStatsCleanupRecurrence.Get(&st.SV)
	if err := compactionSchedule.SetSchedule(schedule); err != nil {
		__antithesis_instrumentation__.Notify(624655)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624656)
	}
	__antithesis_instrumentation__.Notify(624648)

	compactionSchedule.SetScheduleDetails(jobspb.ScheduleDetails{
		Wait:    jobspb.ScheduleDetails_SKIP,
		OnError: jobspb.ScheduleDetails_RETRY_SCHED,
	})

	compactionSchedule.SetScheduleLabel(compactionScheduleName)
	compactionSchedule.SetOwner(security.NodeUserName())

	args, err := pbtypes.MarshalAny(&ScheduledSQLStatsCompactorExecutionArgs{})
	if err != nil {
		__antithesis_instrumentation__.Notify(624657)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624658)
	}
	__antithesis_instrumentation__.Notify(624649)
	compactionSchedule.SetExecutionDetails(
		tree.ScheduledSQLStatsCompactionExecutor.InternalName(),
		jobspb.ExecutionArguments{Args: args},
	)

	compactionSchedule.SetScheduleStatus(string(jobs.StatusPending))
	if err = compactionSchedule.Create(ctx, ie, txn); err != nil {
		__antithesis_instrumentation__.Notify(624659)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624660)
	}
	__antithesis_instrumentation__.Notify(624650)

	return compactionSchedule, nil
}

func CreateCompactionJob(
	ctx context.Context, createdByInfo *jobs.CreatedByInfo, txn *kv.Txn, jobRegistry *jobs.Registry,
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(624661)
	record := jobs.Record{
		Description: "automatic SQL Stats compaction",
		Username:    security.NodeUserName(),
		Details:     jobspb.AutoSQLStatsCompactionDetails{},
		Progress:    jobspb.AutoSQLStatsCompactionProgress{},
		CreatedBy:   createdByInfo,
	}

	jobID := jobRegistry.MakeJobID()
	if _, err := jobRegistry.CreateAdoptableJobWithTxn(ctx, record, jobID, txn); err != nil {
		__antithesis_instrumentation__.Notify(624663)
		return jobspb.InvalidJobID, err
	} else {
		__antithesis_instrumentation__.Notify(624664)
	}
	__antithesis_instrumentation__.Notify(624662)
	return jobID, nil
}

func checkExistingCompactionSchedule(
	ctx context.Context, ie sqlutil.InternalExecutor, txn *kv.Txn,
) (exists bool, _ error) {
	__antithesis_instrumentation__.Notify(624665)
	query := "SELECT count(*) FROM system.scheduled_jobs WHERE schedule_name = $1"

	row, err := ie.QueryRowEx(ctx, "check-existing-sql-stats-schedule", txn,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		query, compactionScheduleName,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(624669)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(624670)
	}
	__antithesis_instrumentation__.Notify(624666)

	if row == nil {
		__antithesis_instrumentation__.Notify(624671)
		return false, errors.AssertionFailedf("unexpected empty result when querying system.scheduled_job")
	} else {
		__antithesis_instrumentation__.Notify(624672)
	}
	__antithesis_instrumentation__.Notify(624667)

	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(624673)
		return false, errors.AssertionFailedf("unexpectedly received %d columns", len(row))
	} else {
		__antithesis_instrumentation__.Notify(624674)
	}
	__antithesis_instrumentation__.Notify(624668)

	return tree.MustBeDInt(row[0]) > 0, nil
}
