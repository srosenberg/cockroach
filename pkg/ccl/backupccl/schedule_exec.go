package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
	pbtypes "github.com/gogo/protobuf/types"
)

type scheduledBackupExecutor struct {
	metrics backupMetrics
}

type backupMetrics struct {
	*jobs.ExecutorMetrics
	RpoMetric *metric.Gauge
}

var _ metric.Struct = &backupMetrics{}

func (m *backupMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(12564) }

var _ jobs.ScheduledJobExecutor = &scheduledBackupExecutor{}

func (e *scheduledBackupExecutor) ExecuteJob(
	ctx context.Context,
	cfg *scheduledjobs.JobExecutionConfig,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(12565)
	if err := e.executeBackup(ctx, cfg, sj, txn); err != nil {
		__antithesis_instrumentation__.Notify(12567)
		e.metrics.NumFailed.Inc(1)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12568)
	}
	__antithesis_instrumentation__.Notify(12566)
	e.metrics.NumStarted.Inc(1)
	return nil
}

func (e *scheduledBackupExecutor) executeBackup(
	ctx context.Context, cfg *scheduledjobs.JobExecutionConfig, sj *jobs.ScheduledJob, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(12569)
	backupStmt, err := extractBackupStatement(sj)
	if err != nil {
		__antithesis_instrumentation__.Notify(12576)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12577)
	}
	__antithesis_instrumentation__.Notify(12570)

	if !backupStmt.Options.Detached {
		__antithesis_instrumentation__.Notify(12578)
		backupStmt.Options.Detached = true
		log.Warningf(ctx, "force setting detached option for backup schedule %d",
			sj.ScheduleID())
	} else {
		__antithesis_instrumentation__.Notify(12579)
	}
	__antithesis_instrumentation__.Notify(12571)

	if sj.IsPaused() {
		__antithesis_instrumentation__.Notify(12580)
		return errors.New("scheduled unexpectedly paused")
	} else {
		__antithesis_instrumentation__.Notify(12581)
	}
	__antithesis_instrumentation__.Notify(12572)

	endTime, err := tree.MakeDTimestampTZ(sj.ScheduledRunTime(), time.Microsecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(12582)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12583)
	}
	__antithesis_instrumentation__.Notify(12573)
	backupStmt.AsOf = tree.AsOfClause{Expr: endTime}

	if knobs, ok := cfg.TestingKnobs.(*jobs.TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(12584)
		if knobs.OverrideAsOfClause != nil {
			__antithesis_instrumentation__.Notify(12585)
			knobs.OverrideAsOfClause(&backupStmt.AsOf)
		} else {
			__antithesis_instrumentation__.Notify(12586)
		}
	} else {
		__antithesis_instrumentation__.Notify(12587)
	}
	__antithesis_instrumentation__.Notify(12574)

	log.Infof(ctx, "Starting scheduled backup %d: %s",
		sj.ScheduleID(), tree.AsString(backupStmt))

	hook, cleanup := cfg.PlanHookMaker("exec-backup", txn, sj.Owner())
	defer cleanup()
	backupFn, err := planBackup(ctx, hook.(sql.PlanHookState), backupStmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(12588)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12589)
	}
	__antithesis_instrumentation__.Notify(12575)
	return invokeBackup(ctx, backupFn)
}

func invokeBackup(ctx context.Context, backupFn sql.PlanHookRowFn) error {
	__antithesis_instrumentation__.Notify(12590)
	resultCh := make(chan tree.Datums)
	g := ctxgroup.WithContext(ctx)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(12593)
		select {
		case <-resultCh:
			__antithesis_instrumentation__.Notify(12594)
			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(12595)
			return ctx.Err()
		}
	})
	__antithesis_instrumentation__.Notify(12591)

	g.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(12596)
		return backupFn(ctx, nil, resultCh)
	})
	__antithesis_instrumentation__.Notify(12592)

	return g.Wait()
}

func planBackup(
	ctx context.Context, p sql.PlanHookState, backupStmt tree.Statement,
) (sql.PlanHookRowFn, error) {
	__antithesis_instrumentation__.Notify(12597)
	fn, cols, _, _, err := backupPlanHook(ctx, backupStmt, p)
	if err != nil {
		__antithesis_instrumentation__.Notify(12601)
		return nil, errors.Wrapf(err, "backup eval: %q", tree.AsString(backupStmt))
	} else {
		__antithesis_instrumentation__.Notify(12602)
	}
	__antithesis_instrumentation__.Notify(12598)
	if fn == nil {
		__antithesis_instrumentation__.Notify(12603)
		return nil, errors.Newf("backup eval: %q", tree.AsString(backupStmt))
	} else {
		__antithesis_instrumentation__.Notify(12604)
	}
	__antithesis_instrumentation__.Notify(12599)
	if len(cols) != len(jobs.DetachedJobExecutionResultHeader) {
		__antithesis_instrumentation__.Notify(12605)
		return nil, errors.Newf("unexpected result columns")
	} else {
		__antithesis_instrumentation__.Notify(12606)
	}
	__antithesis_instrumentation__.Notify(12600)
	return fn, nil
}

func (e *scheduledBackupExecutor) NotifyJobTermination(
	ctx context.Context,
	jobID jobspb.JobID,
	jobStatus jobs.Status,
	details jobspb.Details,
	env scheduledjobs.JobSchedulerEnv,
	schedule *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(12607)
	if jobStatus == jobs.StatusSucceeded {
		__antithesis_instrumentation__.Notify(12609)
		e.metrics.NumSucceeded.Inc(1)
		log.Infof(ctx, "backup job %d scheduled by %d succeeded", jobID, schedule.ScheduleID())
		return e.backupSucceeded(ctx, schedule, details, env, ex, txn)
	} else {
		__antithesis_instrumentation__.Notify(12610)
	}
	__antithesis_instrumentation__.Notify(12608)

	e.metrics.NumFailed.Inc(1)
	err := errors.Errorf(
		"backup job %d scheduled by %d failed with status %s",
		jobID, schedule.ScheduleID(), jobStatus)
	log.Errorf(ctx, "backup error: %v	", err)
	jobs.DefaultHandleFailedRun(schedule, "backup job %d failed with err=%v", jobID, err)
	return nil
}

func (e *scheduledBackupExecutor) GetCreateScheduleStatement(
	ctx context.Context,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	descsCol *descs.Collection,
	sj *jobs.ScheduledJob,
	ex sqlutil.InternalExecutor,
) (string, error) {
	__antithesis_instrumentation__.Notify(12611)
	backupNode, err := extractBackupStatement(sj)
	if err != nil {
		__antithesis_instrumentation__.Notify(12623)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(12624)
	}
	__antithesis_instrumentation__.Notify(12612)

	args := &ScheduledBackupExecutionArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(12625)
		return "", errors.Wrap(err, "un-marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(12626)
	}
	__antithesis_instrumentation__.Notify(12613)

	recurrence := sj.ScheduleExpr()
	fullBackup := &tree.FullBackupClause{AlwaysFull: true}

	var dependentSchedule *jobs.ScheduledJob
	if args.DependentScheduleID != 0 {
		__antithesis_instrumentation__.Notify(12627)
		dependentSchedule, err = jobs.LoadScheduledJob(ctx, env, args.DependentScheduleID, ex, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(12629)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(12630)
		}
		__antithesis_instrumentation__.Notify(12628)

		fullBackup.AlwaysFull = false

		if backupNode.AppendToLatest {
			__antithesis_instrumentation__.Notify(12631)
			fullBackup.Recurrence = tree.NewDString(dependentSchedule.ScheduleExpr())
		} else {
			__antithesis_instrumentation__.Notify(12632)

			fullBackup.Recurrence = tree.NewDString(recurrence)
			recurrence = dependentSchedule.ScheduleExpr()
		}
	} else {
		__antithesis_instrumentation__.Notify(12633)

		if backupNode.AppendToLatest {
			__antithesis_instrumentation__.Notify(12634)
			fullBackup.AlwaysFull = false
			fullBackup.Recurrence = nil
		} else {
			__antithesis_instrumentation__.Notify(12635)
		}
	}
	__antithesis_instrumentation__.Notify(12614)

	firstRunTime := sj.ScheduledRunTime()
	if dependentSchedule != nil {
		__antithesis_instrumentation__.Notify(12636)
		dependentScheduleFirstRun := dependentSchedule.ScheduledRunTime()
		if firstRunTime.IsZero() {
			__antithesis_instrumentation__.Notify(12638)
			firstRunTime = dependentScheduleFirstRun
		} else {
			__antithesis_instrumentation__.Notify(12639)
		}
		__antithesis_instrumentation__.Notify(12637)
		if !dependentScheduleFirstRun.IsZero() && func() bool {
			__antithesis_instrumentation__.Notify(12640)
			return dependentScheduleFirstRun.Before(firstRunTime) == true
		}() == true {
			__antithesis_instrumentation__.Notify(12641)
			firstRunTime = dependentScheduleFirstRun
		} else {
			__antithesis_instrumentation__.Notify(12642)
		}
	} else {
		__antithesis_instrumentation__.Notify(12643)
	}
	__antithesis_instrumentation__.Notify(12615)
	if firstRunTime.IsZero() {
		__antithesis_instrumentation__.Notify(12644)
		firstRunTime = env.Now()
	} else {
		__antithesis_instrumentation__.Notify(12645)
	}
	__antithesis_instrumentation__.Notify(12616)

	firstRun, err := tree.MakeDTimestampTZ(firstRunTime, time.Microsecond)
	if err != nil {
		__antithesis_instrumentation__.Notify(12646)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(12647)
	}
	__antithesis_instrumentation__.Notify(12617)

	wait, err := parseOnPreviousRunningOption(sj.ScheduleDetails().Wait)
	if err != nil {
		__antithesis_instrumentation__.Notify(12648)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(12649)
	}
	__antithesis_instrumentation__.Notify(12618)
	onError, err := parseOnErrorOption(sj.ScheduleDetails().OnError)
	if err != nil {
		__antithesis_instrumentation__.Notify(12650)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(12651)
	}
	__antithesis_instrumentation__.Notify(12619)
	scheduleOptions := tree.KVOptions{
		tree.KVOption{
			Key:   optFirstRun,
			Value: firstRun,
		},
		tree.KVOption{
			Key:   optOnExecFailure,
			Value: tree.NewDString(onError),
		},
		tree.KVOption{
			Key:   optOnPreviousRunning,
			Value: tree.NewDString(wait),
		},
	}

	var destinations []string
	for i := range backupNode.To {
		__antithesis_instrumentation__.Notify(12652)
		dest, ok := backupNode.To[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(12654)
			return "", errors.Errorf("unexpected %T destination in backup statement", dest)
		} else {
			__antithesis_instrumentation__.Notify(12655)
		}
		__antithesis_instrumentation__.Notify(12653)
		destinations = append(destinations, dest.RawString())
	}
	__antithesis_instrumentation__.Notify(12620)

	var kmsURIs []string
	for i := range backupNode.Options.EncryptionKMSURI {
		__antithesis_instrumentation__.Notify(12656)
		kmsURI, ok := backupNode.Options.EncryptionKMSURI[i].(*tree.StrVal)
		if !ok {
			__antithesis_instrumentation__.Notify(12658)
			return "", errors.Errorf("unexpected %T kmsURI in backup statement", kmsURI)
		} else {
			__antithesis_instrumentation__.Notify(12659)
		}
		__antithesis_instrumentation__.Notify(12657)
		kmsURIs = append(kmsURIs, kmsURI.RawString())
	}
	__antithesis_instrumentation__.Notify(12621)

	redactedBackupNode, err := GetRedactedBackupNode(
		backupNode.Backup,
		destinations,
		nil,
		kmsURIs,
		"",
		nil,
		false)
	if err != nil {
		__antithesis_instrumentation__.Notify(12660)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(12661)
	}
	__antithesis_instrumentation__.Notify(12622)

	node := &tree.ScheduledBackup{
		ScheduleLabelSpec: tree.ScheduleLabelSpec{
			IfNotExists: false, Label: tree.NewDString(sj.ScheduleLabel()),
		},
		Recurrence:      tree.NewDString(recurrence),
		FullBackup:      fullBackup,
		Targets:         redactedBackupNode.Targets,
		To:              redactedBackupNode.To,
		BackupOptions:   redactedBackupNode.Options,
		ScheduleOptions: scheduleOptions,
	}

	return tree.AsString(node), nil
}

func (e *scheduledBackupExecutor) backupSucceeded(
	ctx context.Context,
	schedule *jobs.ScheduledJob,
	details jobspb.Details,
	env scheduledjobs.JobSchedulerEnv,
	ex sqlutil.InternalExecutor,
	txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(12662)
	args := &ScheduledBackupExecutionArgs{}
	if err := pbtypes.UnmarshalAny(schedule.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(12670)
		return errors.Wrap(err, "un-marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(12671)
	}
	__antithesis_instrumentation__.Notify(12663)

	if args.UpdatesLastBackupMetric {
		__antithesis_instrumentation__.Notify(12672)
		e.metrics.RpoMetric.Update(details.(jobspb.BackupDetails).EndTime.GoTime().Unix())
	} else {
		__antithesis_instrumentation__.Notify(12673)
	}
	__antithesis_instrumentation__.Notify(12664)

	if args.UnpauseOnSuccess == jobs.InvalidScheduleID {
		__antithesis_instrumentation__.Notify(12674)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12675)
	}
	__antithesis_instrumentation__.Notify(12665)

	s, err := jobs.LoadScheduledJob(ctx, env, args.UnpauseOnSuccess, ex, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(12676)
		if jobs.HasScheduledJobNotFoundError(err) {
			__antithesis_instrumentation__.Notify(12678)
			log.Warningf(ctx, "cannot find schedule %d to unpause; it may have been dropped",
				args.UnpauseOnSuccess)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(12679)
		}
		__antithesis_instrumentation__.Notify(12677)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12680)
	}
	__antithesis_instrumentation__.Notify(12666)
	s.ClearScheduleStatus()
	if s.HasRecurringSchedule() {
		__antithesis_instrumentation__.Notify(12681)
		if err := s.ScheduleNextRun(); err != nil {
			__antithesis_instrumentation__.Notify(12682)
			return err
		} else {
			__antithesis_instrumentation__.Notify(12683)
		}
	} else {
		__antithesis_instrumentation__.Notify(12684)
	}
	__antithesis_instrumentation__.Notify(12667)
	if err := s.Update(ctx, ex, txn); err != nil {
		__antithesis_instrumentation__.Notify(12685)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12686)
	}
	__antithesis_instrumentation__.Notify(12668)

	args.UnpauseOnSuccess = jobs.InvalidScheduleID
	any, err := pbtypes.MarshalAny(args)
	if err != nil {
		__antithesis_instrumentation__.Notify(12687)
		return errors.Wrap(err, "marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(12688)
	}
	__antithesis_instrumentation__.Notify(12669)
	schedule.SetExecutionDetails(
		schedule.ExecutorType(), jobspb.ExecutionArguments{Args: any},
	)

	return nil
}

func (e *scheduledBackupExecutor) Metrics() metric.Struct {
	__antithesis_instrumentation__.Notify(12689)
	return &e.metrics
}

func extractBackupStatement(sj *jobs.ScheduledJob) (*annotatedBackupStatement, error) {
	__antithesis_instrumentation__.Notify(12690)
	args := &ScheduledBackupExecutionArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(12694)
		return nil, errors.Wrap(err, "un-marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(12695)
	}
	__antithesis_instrumentation__.Notify(12691)

	node, err := parser.ParseOne(args.BackupStatement)
	if err != nil {
		__antithesis_instrumentation__.Notify(12696)
		return nil, errors.Wrap(err, "parsing backup statement")
	} else {
		__antithesis_instrumentation__.Notify(12697)
	}
	__antithesis_instrumentation__.Notify(12692)

	if backupStmt, ok := node.AST.(*tree.Backup); ok {
		__antithesis_instrumentation__.Notify(12698)
		return &annotatedBackupStatement{
			Backup: backupStmt,
			CreatedByInfo: &jobs.CreatedByInfo{
				Name: jobs.CreatedByScheduledJobs,
				ID:   sj.ScheduleID(),
			},
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(12699)
	}
	__antithesis_instrumentation__.Notify(12693)

	return nil, errors.Newf("unexpect node type %T", node)
}

var _ jobs.ScheduledJobController = &scheduledBackupExecutor{}

func unlinkDependentSchedule(
	ctx context.Context,
	scheduleControllerEnv scheduledjobs.ScheduleControllerEnv,
	env scheduledjobs.JobSchedulerEnv,
	txn *kv.Txn,
	args *ScheduledBackupExecutionArgs,
) error {
	__antithesis_instrumentation__.Notify(12700)
	if args.DependentScheduleID == 0 {
		__antithesis_instrumentation__.Notify(12704)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(12705)
	}
	__antithesis_instrumentation__.Notify(12701)

	dependentSj, dependentArgs, err := getScheduledBackupExecutionArgsFromSchedule(ctx, env, txn,
		scheduleControllerEnv.InternalExecutor().(*sql.InternalExecutor), args.DependentScheduleID)
	if err != nil {
		__antithesis_instrumentation__.Notify(12706)
		if jobs.HasScheduledJobNotFoundError(err) {
			__antithesis_instrumentation__.Notify(12708)
			log.Warningf(ctx, "failed to resolve dependent schedule %d", args.DependentScheduleID)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(12709)
		}
		__antithesis_instrumentation__.Notify(12707)
		return errors.Wrapf(err, "failed to resolve dependent schedule %d", args.DependentScheduleID)
	} else {
		__antithesis_instrumentation__.Notify(12710)
	}
	__antithesis_instrumentation__.Notify(12702)

	dependentArgs.DependentScheduleID = 0
	any, err := pbtypes.MarshalAny(dependentArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(12711)
		return err
	} else {
		__antithesis_instrumentation__.Notify(12712)
	}
	__antithesis_instrumentation__.Notify(12703)
	dependentSj.SetExecutionDetails(dependentSj.ExecutorType(), jobspb.ExecutionArguments{Args: any})
	return dependentSj.Update(ctx, scheduleControllerEnv.InternalExecutor(), txn)
}

func (e *scheduledBackupExecutor) OnDrop(
	ctx context.Context,
	scheduleControllerEnv scheduledjobs.ScheduleControllerEnv,
	env scheduledjobs.JobSchedulerEnv,
	sj *jobs.ScheduledJob,
	txn *kv.Txn,
	descsCol *descs.Collection,
) error {
	__antithesis_instrumentation__.Notify(12713)
	args := &ScheduledBackupExecutionArgs{}
	if err := pbtypes.UnmarshalAny(sj.ExecutionArgs().Args, args); err != nil {
		__antithesis_instrumentation__.Notify(12716)
		return errors.Wrap(err, "un-marshaling args")
	} else {
		__antithesis_instrumentation__.Notify(12717)
	}
	__antithesis_instrumentation__.Notify(12714)

	if err := unlinkDependentSchedule(ctx, scheduleControllerEnv, env, txn, args); err != nil {
		__antithesis_instrumentation__.Notify(12718)
		return errors.Wrap(err, "failed to unlink dependent schedule")
	} else {
		__antithesis_instrumentation__.Notify(12719)
	}
	__antithesis_instrumentation__.Notify(12715)
	return releaseProtectedTimestamp(ctx, txn, scheduleControllerEnv.PTSProvider(),
		args.ProtectedTimestampRecord)
}

func init() {
	jobs.RegisterScheduledJobExecutorFactory(
		tree.ScheduledBackupExecutor.InternalName(),
		func() (jobs.ScheduledJobExecutor, error) {
			m := jobs.MakeExecutorMetrics(tree.ScheduledBackupExecutor.UserName())
			return &scheduledBackupExecutor{
				metrics: backupMetrics{
					ExecutorMetrics: &m,
					RpoMetric: metric.NewGauge(metric.Metadata{
						Name:        "schedules.BACKUP.last-completed-time",
						Help:        "The unix timestamp of the most recently completed backup by a schedule specified as maintaining this metric",
						Measurement: "Jobs",
						Unit:        metric.Unit_TIMESTAMP_SEC,
					}),
				},
			}, nil
		})
}
