package ttlschedule

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

type rowLevelTTLExecutor struct {
	metrics rowLevelTTLMetrics
}

var _ jobs.ScheduledJobController = (*rowLevelTTLExecutor)(nil)

type rowLevelTTLMetrics struct {
	*jobs.ExecutorMetrics
}

var _ metric.Struct = &rowLevelTTLMetrics{}

func (m *rowLevelTTLMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(628864) }

func (s rowLevelTTLExecutor) OnDrop(
	ctx context.Context,
	scheduleControllerEnv scheduledjobs.ScheduleControllerEnv,
	env scheduledjobs.JobSchedulerEnv,
	schedule *jobs.ScheduledJob,
	txn *kv.Txn,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(628865)

	var args catpb.ScheduledRowLevelTTLArgs
	if err := pbtypes.UnmarshalAny(schedule.ExecutionArgs().Args, &args); err != nil {
		__antithesis_instrumentation__.Notify(628869)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628870)
	}
	__antithesis_instrumentation__.Notify(628866)

	canDrop, err := canDropTTLSchedule(ctx, txn, descsCol, schedule, args)
	if err != nil {
		__antithesis_instrumentation__.Notify(628871)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628872)
	}
	__antithesis_instrumentation__.Notify(628867)

	if !canDrop {
		__antithesis_instrumentation__.Notify(628873)
		tn, err := descs.GetTableNameByID(ctx, txn, descsCol, args.TableID)
		if err != nil {
			__antithesis_instrumentation__.Notify(628875)
			return err
		} else {
			__antithesis_instrumentation__.Notify(628876)
		}
		__antithesis_instrumentation__.Notify(628874)
		return errors.WithHintf(
			pgerror.Newf(
				pgcode.InvalidTableDefinition,
				"cannot drop a row level TTL schedule",
			),
			`use ALTER TABLE %s RESET (ttl) instead`,
			tn.FQString(),
		)
	} else {
		__antithesis_instrumentation__.Notify(628877)
	}
	__antithesis_instrumentation__.Notify(628868)
	return nil
}

func canDropTTLSchedule(
	ctx context.Context,
	txn *kv.Txn,
	descsCol *descs.Collection,
	schedule *jobs.ScheduledJob,
	args catpb.ScheduledRowLevelTTLArgs,
) (bool, error) {
	__antithesis_instrumentation__.Notify(628878)
	desc, err := descsCol.GetImmutableTableByID(ctx, txn, args.TableID, tree.ObjectLookupFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(628883)

		if sqlerrors.IsUndefinedRelationError(err) {
			__antithesis_instrumentation__.Notify(628885)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(628886)
		}
		__antithesis_instrumentation__.Notify(628884)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(628887)
	}
	__antithesis_instrumentation__.Notify(628879)
	if desc == nil {
		__antithesis_instrumentation__.Notify(628888)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(628889)
	}
	__antithesis_instrumentation__.Notify(628880)

	if !desc.HasRowLevelTTL() {
		__antithesis_instrumentation__.Notify(628890)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(628891)
	}
	__antithesis_instrumentation__.Notify(628881)

	if desc.GetRowLevelTTL().ScheduleID != schedule.ScheduleID() {
		__antithesis_instrumentation__.Notify(628892)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(628893)
	}
	__antithesis_instrumentation__.Notify(628882)
	return false, nil
}

func (s rowLevelTTLExecutor) ExecuteJob(
	ctx context.Context,
	cfg *scheduledjobs.JobExecutionConfig,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(628894)
	args := &catpb.ScheduledRowLevelTTLArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(628897)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628898)
	}
	__antithesis_instrumentation__.Notify(628895)

	p, cleanup := cfg.PlanHookMaker(
		fmt.Sprintf("invoke-row-level-ttl-%d", args.TableID),
		txn,
		security.NodeUserName(),
	)
	defer cleanup()

	if _, err := createRowLevelTTLJob(
		ctx,
		&jobs.CreatedByInfo{
			ID:   sj.ScheduleID(),
			Name: jobs.CreatedByScheduledJobs,
		},
		txn,
		p.(sql.PlanHookState).ExtendedEvalContext().Descs,
		p.(sql.PlanHookState).ExecCfg().JobRegistry,
		*args,
	); err != nil {
		__antithesis_instrumentation__.Notify(628899)
		s.metrics.NumFailed.Inc(1)
		return err
	} else {
		__antithesis_instrumentation__.Notify(628900)
	}
	__antithesis_instrumentation__.Notify(628896)
	s.metrics.NumStarted.Inc(1)
	return nil
}

func (s rowLevelTTLExecutor) NotifyJobTermination(
	ctx context.Context,
	jobID jobspb.JobID,
	jobStatus jobs.Status,
	details jobspb.Details,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(628901)
	if jobStatus == jobs.StatusFailed {
		__antithesis_instrumentation__.Notify(628904)
		jobs.DefaultHandleFailedRun(
			sj,
			"row level ttl for table [%d] job failed",
			details.(catpb.ScheduledRowLevelTTLArgs).TableID,
		)
		s.metrics.NumFailed.Inc(1)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(628905)
	}
	__antithesis_instrumentation__.Notify(628902)

	if jobStatus == jobs.StatusSucceeded {
		__antithesis_instrumentation__.Notify(628906)
		s.metrics.NumSucceeded.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(628907)
	}
	__antithesis_instrumentation__.Notify(628903)

	sj.SetScheduleStatus(string(jobStatus))
	return nil
}

func (s rowLevelTTLExecutor) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(628908)
	return &s.metrics
}

func (s rowLevelTTLExecutor) GetCreateScheduleStatement(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	descsCol *descs.Collection,
	sj *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
) (string, error) {
	__antithesis_instrumentation__.Notify(628909)
	args := &catpb.ScheduledRowLevelTTLArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(628912)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(628913)
	}
	__antithesis_instrumentation__.Notify(628910)
	tn, err := descs.GetTableNameByID(ctx, txn, descsCol, args.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(628914)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(628915)
	}
	__antithesis_instrumentation__.Notify(628911)
	return fmt.Sprintf(`ALTER TABLE %s WITH (ttl = 'on', ...)`, tn.FQString()), nil
}

func createRowLevelTTLJob(
	ctx context.Context,
	createdByInfo *jobs.CreatedByInfo,
	txn *kv.Txn,
	descsCol *descs.Collection,
	jobRegistry *jobs.Registry,
	ttlDetails catpb.ScheduledRowLevelTTLArgs,
) (jobspb.JobID, error) {
	__antithesis_instrumentation__.Notify(628916)
	tn, err := descs.GetTableNameByID(ctx, txn, descsCol, ttlDetails.TableID)
	if err != nil {
		__antithesis_instrumentation__.Notify(628919)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(628920)
	}
	__antithesis_instrumentation__.Notify(628917)
	record := jobs.Record{
		Description: fmt.Sprintf("ttl for %s", tn.FQString()),
		Username:    security.NodeUserName(),
		Details: jobspb.RowLevelTTLDetails{
			TableID: ttlDetails.TableID,
			Cutoff:  timeutil.Now(),
		},
		Progress:  jobspb.RowLevelTTLProgress{},
		CreatedBy: createdByInfo,
	}

	jobID := jobRegistry.MakeJobID()
	if _, err := jobRegistry.CreateAdoptableJobWithTxn(ctx, record, jobID, txn); err != nil {
		__antithesis_instrumentation__.Notify(628921)
		return jobspb.InvalidJobID, err
	} else {
		__antithesis_instrumentation__.Notify(628922)
	}
	__antithesis_instrumentation__.Notify(628918)
	return jobID, nil
}

func init() {
	jobs.RegisterScheduledJobExecutorFactory(
		tree.ScheduledRowLevelTTLExecutor.InternalName(),
		func() (jobs.ScheduledJobExecutor, error) {
			m := jobs.MakeExecutorMetrics(tree.ScheduledRowLevelTTLExecutor.InternalName())
			return &rowLevelTTLExecutor{
				metrics: rowLevelTTLMetrics{
					ExecutorMetrics: &m,
				},
			}, nil
		},
	)
}
