package jobs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
)

const CreatedByScheduledJobs = "crdb_schedule"

type jobScheduler struct {
	*scheduledjobs.JobExecutionConfig
	env      scheduledjobs.JobSchedulerEnv
	registry *metric.Registry
	metrics  SchedulerMetrics
}

func newJobScheduler(
	cfg *scheduledjobs.JobExecutionConfig,
	env scheduledjobs.JobSchedulerEnv,
	registry *metric.Registry,
) *jobScheduler {
	__antithesis_instrumentation__.Notify(70263)
	if env == nil {
		__antithesis_instrumentation__.Notify(70265)
		env = scheduledjobs.ProdJobSchedulerEnv
	} else {
		__antithesis_instrumentation__.Notify(70266)
	}
	__antithesis_instrumentation__.Notify(70264)

	stats := MakeSchedulerMetrics()
	registry.AddMetricStruct(stats)

	return &jobScheduler{
		JobExecutionConfig: cfg,
		env:                env,
		registry:           registry,
		metrics:            stats,
	}
}

const allSchedules = 0

var errScheduleNotRunnable = errors.New("schedule not runnable")

func loadCandidateScheduleForExecution(
	ctx context.Context,
	scheduleID int64,
	env scheduledjobs.JobSchedulerEnv,
	ie sqlutil.InternalExecutor,
	txn *kv.Txn,
) (*ScheduledJob, error) {
	__antithesis_instrumentation__.Notify(70267)
	lookupStmt := fmt.Sprintf(
		"SELECT * FROM %s WHERE schedule_id=%d AND next_run < %s FOR UPDATE",
		env.ScheduledJobsTableName(), scheduleID, env.NowExpr())
	row, cols, err := ie.QueryRowExWithCols(
		ctx, "find-scheduled-jobs-exec",
		txn,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		lookupStmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(70271)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70272)
	}
	__antithesis_instrumentation__.Notify(70268)

	if row == nil {
		__antithesis_instrumentation__.Notify(70273)
		return nil, errScheduleNotRunnable
	} else {
		__antithesis_instrumentation__.Notify(70274)
	}
	__antithesis_instrumentation__.Notify(70269)

	j := NewScheduledJob(env)
	if err := j.InitFromDatums(row, cols); err != nil {
		__antithesis_instrumentation__.Notify(70275)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70276)
	}
	__antithesis_instrumentation__.Notify(70270)
	return j, nil
}

func lookupNumRunningJobs(
	ctx context.Context,
	scheduleID int64,
	env scheduledjobs.JobSchedulerEnv,
	ie sqlutil.InternalExecutor,
) (int64, error) {
	__antithesis_instrumentation__.Notify(70277)
	lookupStmt := fmt.Sprintf(
		"SELECT count(*) FROM %s WHERE created_by_type = '%s' AND created_by_id = %d AND status IN %s",
		env.SystemJobsTableName(), CreatedByScheduledJobs, scheduleID, NonTerminalStatusTupleString)
	row, err := ie.QueryRowEx(
		ctx, "lookup-num-running",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		lookupStmt)
	if err != nil {
		__antithesis_instrumentation__.Notify(70279)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(70280)
	}
	__antithesis_instrumentation__.Notify(70278)
	return int64(tree.MustBeDInt(row[0])), nil
}

const recheckRunningAfter = 1 * time.Minute

func (s *jobScheduler) processSchedule(
	ctx context.Context, schedule *ScheduledJob, numRunning int64, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(70281)
	if numRunning > 0 {
		__antithesis_instrumentation__.Notify(70287)
		switch schedule.ScheduleDetails().Wait {
		case jobspb.ScheduleDetails_WAIT:
			__antithesis_instrumentation__.Notify(70288)

			schedule.SetNextRun(s.env.Now().Add(recheckRunningAfter))
			schedule.SetScheduleStatus("delayed due to %d already running", numRunning)
			s.metrics.RescheduleWait.Inc(1)
			return schedule.Update(ctx, s.InternalExecutor, txn)
		case jobspb.ScheduleDetails_SKIP:
			__antithesis_instrumentation__.Notify(70289)
			if err := schedule.ScheduleNextRun(); err != nil {
				__antithesis_instrumentation__.Notify(70292)
				return err
			} else {
				__antithesis_instrumentation__.Notify(70293)
			}
			__antithesis_instrumentation__.Notify(70290)
			schedule.SetScheduleStatus("rescheduled due to %d already running", numRunning)
			s.metrics.RescheduleSkip.Inc(1)
			return schedule.Update(ctx, s.InternalExecutor, txn)
		default:
			__antithesis_instrumentation__.Notify(70291)
		}
	} else {
		__antithesis_instrumentation__.Notify(70294)
	}
	__antithesis_instrumentation__.Notify(70282)

	schedule.ClearScheduleStatus()

	if schedule.HasRecurringSchedule() {
		__antithesis_instrumentation__.Notify(70295)
		if err := schedule.ScheduleNextRun(); err != nil {
			__antithesis_instrumentation__.Notify(70296)
			return err
		} else {
			__antithesis_instrumentation__.Notify(70297)
		}
	} else {
		__antithesis_instrumentation__.Notify(70298)

		schedule.SetNextRun(time.Time{})
	}
	__antithesis_instrumentation__.Notify(70283)

	if err := schedule.Update(ctx, s.InternalExecutor, txn); err != nil {
		__antithesis_instrumentation__.Notify(70299)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70300)
	}
	__antithesis_instrumentation__.Notify(70284)

	executor, err := GetScheduledJobExecutor(schedule.ExecutorType())
	if err != nil {
		__antithesis_instrumentation__.Notify(70301)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70302)
	}
	__antithesis_instrumentation__.Notify(70285)

	log.Infof(ctx,
		"Starting job for schedule %d (%q); scheduled to run at %s; next run scheduled for %s",
		schedule.ScheduleID(), schedule.ScheduleLabel(),
		schedule.ScheduledRunTime(), schedule.NextRun())

	execCtx := logtags.AddTag(ctx, "schedule", schedule.ScheduleID())
	if err := executor.ExecuteJob(execCtx, s.JobExecutionConfig, s.env, schedule, txn); err != nil {
		__antithesis_instrumentation__.Notify(70303)
		return errors.Wrapf(err, "executing schedule %d", schedule.ScheduleID())
	} else {
		__antithesis_instrumentation__.Notify(70304)
	}
	__antithesis_instrumentation__.Notify(70286)

	s.metrics.NumStarted.Inc(1)

	return schedule.Update(ctx, s.InternalExecutor, txn)
}

type savePointError struct {
	err error
}

func (e savePointError) Error() string {
	__antithesis_instrumentation__.Notify(70305)
	return e.err.Error()
}

func withSavePoint(ctx context.Context, txn *kv.Txn, fn func() error) error {
	__antithesis_instrumentation__.Notify(70306)
	sp, err := txn.CreateSavepoint(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(70311)
		return &savePointError{err}
	} else {
		__antithesis_instrumentation__.Notify(70312)
	}
	__antithesis_instrumentation__.Notify(70307)
	execErr := fn()

	if execErr == nil {
		__antithesis_instrumentation__.Notify(70313)
		if err := txn.ReleaseSavepoint(ctx, sp); err != nil {
			__antithesis_instrumentation__.Notify(70315)
			return &savePointError{err}
		} else {
			__antithesis_instrumentation__.Notify(70316)
		}
		__antithesis_instrumentation__.Notify(70314)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(70317)
	}
	__antithesis_instrumentation__.Notify(70308)

	if errors.HasType(execErr, (*roachpb.TransactionRetryWithProtoRefreshError)(nil)) {
		__antithesis_instrumentation__.Notify(70318)

		return &savePointError{execErr}
	} else {
		__antithesis_instrumentation__.Notify(70319)
	}
	__antithesis_instrumentation__.Notify(70309)

	if err := txn.RollbackToSavepoint(ctx, sp); err != nil {
		__antithesis_instrumentation__.Notify(70320)
		return &savePointError{errors.WithDetail(err, execErr.Error())}
	} else {
		__antithesis_instrumentation__.Notify(70321)
	}
	__antithesis_instrumentation__.Notify(70310)
	return execErr
}

func (s *jobScheduler) executeCandidateSchedule(
	ctx context.Context, candidate int64, txn *kv.Txn,
) error {
	__antithesis_instrumentation__.Notify(70322)
	schedule, err := loadCandidateScheduleForExecution(ctx, candidate, s.env, s.InternalExecutor, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(70327)
		if errors.Is(err, errScheduleNotRunnable) {
			__antithesis_instrumentation__.Notify(70329)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(70330)
		}
		__antithesis_instrumentation__.Notify(70328)
		s.metrics.NumMalformedSchedules.Inc(1)
		log.Errorf(ctx, "error parsing schedule %d: %s", candidate, err)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70331)
	}
	__antithesis_instrumentation__.Notify(70323)

	if !s.env.IsExecutorEnabled(schedule.ExecutorType()) {
		__antithesis_instrumentation__.Notify(70332)
		log.Infof(ctx, "Ignoring schedule %d: %s executor disabled",
			schedule.ScheduleID(), schedule.ExecutorType())
		return nil
	} else {
		__antithesis_instrumentation__.Notify(70333)
	}
	__antithesis_instrumentation__.Notify(70324)

	numRunning, err := lookupNumRunningJobs(ctx, schedule.ScheduleID(), s.env, s.InternalExecutor)
	if err != nil {
		__antithesis_instrumentation__.Notify(70334)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70335)
	}
	__antithesis_instrumentation__.Notify(70325)

	timeout := schedulerScheduleExecutionTimeout.Get(&s.Settings.SV)
	if processErr := withSavePoint(ctx, txn, func() error {
		__antithesis_instrumentation__.Notify(70336)
		if timeout > 0 {
			__antithesis_instrumentation__.Notify(70338)
			return contextutil.RunWithTimeout(
				ctx, fmt.Sprintf("process-schedule-%d", schedule.ScheduleID()), timeout,
				func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(70339)
					return s.processSchedule(ctx, schedule, numRunning, txn)
				})
		} else {
			__antithesis_instrumentation__.Notify(70340)
		}
		__antithesis_instrumentation__.Notify(70337)
		return s.processSchedule(ctx, schedule, numRunning, txn)
	}); processErr != nil {
		__antithesis_instrumentation__.Notify(70341)
		if errors.HasType(processErr, (*savePointError)(nil)) {
			__antithesis_instrumentation__.Notify(70343)
			return errors.Wrapf(processErr, "savepoint error for schedule %d", schedule.ScheduleID())
		} else {
			__antithesis_instrumentation__.Notify(70344)
		}
		__antithesis_instrumentation__.Notify(70342)

		s.metrics.NumErrSchedules.Inc(1)
		log.Errorf(ctx,
			"error processing schedule %d: %+v", schedule.ScheduleID(), processErr)

		if err := withSavePoint(ctx, txn, func() error {
			__antithesis_instrumentation__.Notify(70345)

			schedule.ClearDirty()
			DefaultHandleFailedRun(schedule,
				"failed to create job for schedule %d: err=%s",
				schedule.ScheduleID(), processErr)

			if schedule.HasRecurringSchedule() && func() bool {
				__antithesis_instrumentation__.Notify(70347)
				return schedule.ScheduleDetails().OnError == jobspb.ScheduleDetails_RETRY_SCHED == true
			}() == true {
				__antithesis_instrumentation__.Notify(70348)
				if err := schedule.ScheduleNextRun(); err != nil {
					__antithesis_instrumentation__.Notify(70349)
					return err
				} else {
					__antithesis_instrumentation__.Notify(70350)
				}
			} else {
				__antithesis_instrumentation__.Notify(70351)
			}
			__antithesis_instrumentation__.Notify(70346)
			return schedule.Update(ctx, s.InternalExecutor, txn)
		}); err != nil {
			__antithesis_instrumentation__.Notify(70352)
			if errors.HasType(err, (*savePointError)(nil)) {
				__antithesis_instrumentation__.Notify(70354)
				return errors.Wrapf(err,
					"savepoint error for schedule %d", schedule.ScheduleID())
			} else {
				__antithesis_instrumentation__.Notify(70355)
			}
			__antithesis_instrumentation__.Notify(70353)
			log.Errorf(ctx, "error recording processing error for schedule %d: %+v",
				schedule.ScheduleID(), err)
		} else {
			__antithesis_instrumentation__.Notify(70356)
		}
	} else {
		__antithesis_instrumentation__.Notify(70357)
	}
	__antithesis_instrumentation__.Notify(70326)
	return nil
}

func (s *jobScheduler) executeSchedules(ctx context.Context, maxSchedules int64) (retErr error) {
	__antithesis_instrumentation__.Notify(70358)

	limitClause := ""
	if maxSchedules != allSchedules {
		__antithesis_instrumentation__.Notify(70363)
		limitClause = fmt.Sprintf("LIMIT %d", maxSchedules)
	} else {
		__antithesis_instrumentation__.Notify(70364)
	}
	__antithesis_instrumentation__.Notify(70359)

	findSchedulesStmt := fmt.Sprintf(
		`SELECT schedule_id FROM %s WHERE next_run < %s ORDER BY random() %s`,
		s.env.ScheduledJobsTableName(), s.env.NowExpr(), limitClause)
	it, err := s.InternalExecutor.QueryIteratorEx(
		ctx, "find-scheduled-jobs",
		nil,
		sessiondata.InternalExecutorOverride{User: security.RootUserName()},
		findSchedulesStmt)

	if err != nil {
		__antithesis_instrumentation__.Notify(70365)
		return err
	} else {
		__antithesis_instrumentation__.Notify(70366)
	}
	__antithesis_instrumentation__.Notify(70360)

	defer func() {
		__antithesis_instrumentation__.Notify(70367)
		retErr = errors.CombineErrors(retErr, it.Close())
	}()
	__antithesis_instrumentation__.Notify(70361)

	var ok bool
	for ok, err = it.Next(ctx); ok; ok, err = it.Next(ctx) {
		__antithesis_instrumentation__.Notify(70368)
		row := it.Cur()
		candidateID := int64(tree.MustBeDInt(row[0]))
		if err := s.DB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(70369)
			return s.executeCandidateSchedule(ctx, candidateID, txn)
		}); err != nil {
			__antithesis_instrumentation__.Notify(70370)
			log.Errorf(ctx, "error executing candidate schedule %d: %s", candidateID, err)
		} else {
			__antithesis_instrumentation__.Notify(70371)
		}
	}
	__antithesis_instrumentation__.Notify(70362)

	return err
}

var schedulerRunsOnSingleNode = settings.RegisterBoolSetting(
	settings.TenantReadOnly,
	"jobs.scheduler.single_node_scheduler.enabled",
	"execute scheduler on a single node in a cluster",
	false,
)

func (s *jobScheduler) schedulerEnabledOnThisNode(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(70372)
	if s.ShouldRunScheduler == nil || func() bool {
		__antithesis_instrumentation__.Notify(70375)
		return !schedulerRunsOnSingleNode.Get(&s.Settings.SV) == true
	}() == true {
		__antithesis_instrumentation__.Notify(70376)
		return true
	} else {
		__antithesis_instrumentation__.Notify(70377)
	}
	__antithesis_instrumentation__.Notify(70373)

	enabled, err := s.ShouldRunScheduler(ctx, s.DB.Clock().NowAsClockTimestamp())
	if err != nil {
		__antithesis_instrumentation__.Notify(70378)
		log.Errorf(ctx, "error determining if the scheduler enabled: %v; will recheck after %s",
			err, recheckEnabledAfterPeriod)
		return false
	} else {
		__antithesis_instrumentation__.Notify(70379)
	}
	__antithesis_instrumentation__.Notify(70374)
	return enabled
}

type syncCancelFunc struct {
	syncutil.Mutex
	context.CancelFunc
}

func newCancelWhenDisabled(sv *settings.Values) *syncCancelFunc {
	__antithesis_instrumentation__.Notify(70380)
	sf := &syncCancelFunc{}
	schedulerEnabledSetting.SetOnChange(sv, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(70382)
		if !schedulerEnabledSetting.Get(sv) {
			__antithesis_instrumentation__.Notify(70383)
			sf.Lock()
			if sf.CancelFunc != nil {
				__antithesis_instrumentation__.Notify(70385)
				sf.CancelFunc()
			} else {
				__antithesis_instrumentation__.Notify(70386)
			}
			__antithesis_instrumentation__.Notify(70384)
			sf.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(70387)
		}
	})
	__antithesis_instrumentation__.Notify(70381)
	return sf
}

func (sf *syncCancelFunc) withCancelOnDisabled(
	ctx context.Context, sv *settings.Values, f func(ctx context.Context) error,
) error {
	__antithesis_instrumentation__.Notify(70388)
	ctx, cancel := func() (context.Context, context.CancelFunc) {
		__antithesis_instrumentation__.Notify(70390)
		sf.Lock()
		defer sf.Unlock()

		ctx, cancel := context.WithCancel(ctx)
		sf.CancelFunc = cancel

		if !schedulerEnabledSetting.Get(sv) {
			__antithesis_instrumentation__.Notify(70392)
			cancel()
		} else {
			__antithesis_instrumentation__.Notify(70393)
		}
		__antithesis_instrumentation__.Notify(70391)

		return ctx, func() {
			__antithesis_instrumentation__.Notify(70394)
			sf.Lock()
			defer sf.Unlock()
			cancel()
			sf.CancelFunc = nil
		}
	}()
	__antithesis_instrumentation__.Notify(70389)
	defer cancel()
	return f(ctx)
}

func (s *jobScheduler) runDaemon(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(70395)
	_ = stopper.RunAsyncTask(ctx, "job-scheduler", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(70396)
		ctx, cancel := stopper.WithCancelOnQuiesce(ctx)
		defer cancel()

		initialDelay := getInitialScanDelay(s.TestingKnobs)
		log.Infof(ctx, "waiting %v before scheduled jobs daemon start", initialDelay)

		if err := RegisterExecutorsMetrics(s.registry); err != nil {
			__antithesis_instrumentation__.Notify(70398)
			log.Errorf(ctx, "error registering executor metrics: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(70399)
		}
		__antithesis_instrumentation__.Notify(70397)

		whenDisabled := newCancelWhenDisabled(&s.Settings.SV)

		for timer := time.NewTimer(initialDelay); ; timer.Reset(
			getWaitPeriod(ctx, &s.Settings.SV, s.schedulerEnabledOnThisNode, jitter, s.TestingKnobs)) {
			__antithesis_instrumentation__.Notify(70400)
			select {
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(70401)
				return
			case <-timer.C:
				__antithesis_instrumentation__.Notify(70402)
				if !schedulerEnabledSetting.Get(&s.Settings.SV) || func() bool {
					__antithesis_instrumentation__.Notify(70404)
					return !s.schedulerEnabledOnThisNode(ctx) == true
				}() == true {
					__antithesis_instrumentation__.Notify(70405)
					continue
				} else {
					__antithesis_instrumentation__.Notify(70406)
				}
				__antithesis_instrumentation__.Notify(70403)

				maxSchedules := schedulerMaxJobsPerIterationSetting.Get(&s.Settings.SV)
				if err := whenDisabled.withCancelOnDisabled(ctx, &s.Settings.SV, func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(70407)
					return s.executeSchedules(ctx, maxSchedules)
				}); err != nil {
					__antithesis_instrumentation__.Notify(70408)
					log.Errorf(ctx, "error executing schedules: %+v", err)
				} else {
					__antithesis_instrumentation__.Notify(70409)
				}
			}
		}
	})
}

var schedulerEnabledSetting = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"jobs.scheduler.enabled",
	"enable/disable job scheduler",
	true,
)

var schedulerPaceSetting = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"jobs.scheduler.pace",
	"how often to scan system.scheduled_jobs table",
	time.Minute,
)

var schedulerMaxJobsPerIterationSetting = settings.RegisterIntSetting(
	settings.TenantWritable,
	"jobs.scheduler.max_jobs_per_iteration",
	"how many schedules to start per iteration; setting to 0 turns off this limit",
	5,
)

var schedulerScheduleExecutionTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"jobs.scheduler.schedule_execution.timeout",
	"sets a timeout on for schedule execution; 0 disables timeout",
	30*time.Second,
)

func getInitialScanDelay(knobs base.ModuleTestingKnobs) time.Duration {
	__antithesis_instrumentation__.Notify(70410)
	if k, ok := knobs.(*TestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(70412)
		return k.SchedulerDaemonInitialScanDelay != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70413)
		return k.SchedulerDaemonInitialScanDelay()
	} else {
		__antithesis_instrumentation__.Notify(70414)
	}
	__antithesis_instrumentation__.Notify(70411)

	return time.Minute * time.Duration(2+rand.Intn(3))
}

const minPacePeriod = 10 * time.Second

const recheckEnabledAfterPeriod = 5 * time.Minute

var warnIfPaceTooLow = log.Every(time.Minute)

type jitterFn func(duration time.Duration) time.Duration

func getWaitPeriod(
	ctx context.Context,
	sv *settings.Values,
	enabledOnThisNode func(ctx context.Context) bool,
	jitter jitterFn,
	knobs base.ModuleTestingKnobs,
) time.Duration {
	__antithesis_instrumentation__.Notify(70415)
	if k, ok := knobs.(*TestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(70420)
		return k.SchedulerDaemonScanDelay != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70421)
		return k.SchedulerDaemonScanDelay()
	} else {
		__antithesis_instrumentation__.Notify(70422)
	}
	__antithesis_instrumentation__.Notify(70416)

	if !schedulerEnabledSetting.Get(sv) {
		__antithesis_instrumentation__.Notify(70423)
		return recheckEnabledAfterPeriod
	} else {
		__antithesis_instrumentation__.Notify(70424)
	}
	__antithesis_instrumentation__.Notify(70417)

	if enabledOnThisNode != nil && func() bool {
		__antithesis_instrumentation__.Notify(70425)
		return !enabledOnThisNode(ctx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(70426)
		return recheckEnabledAfterPeriod
	} else {
		__antithesis_instrumentation__.Notify(70427)
	}
	__antithesis_instrumentation__.Notify(70418)

	pace := schedulerPaceSetting.Get(sv)
	if pace < minPacePeriod {
		__antithesis_instrumentation__.Notify(70428)
		if warnIfPaceTooLow.ShouldLog() {
			__antithesis_instrumentation__.Notify(70430)
			log.Warningf(ctx,
				"job.scheduler.pace setting too low (%s < %s)", pace, minPacePeriod)
		} else {
			__antithesis_instrumentation__.Notify(70431)
		}
		__antithesis_instrumentation__.Notify(70429)
		pace = minPacePeriod
	} else {
		__antithesis_instrumentation__.Notify(70432)
	}
	__antithesis_instrumentation__.Notify(70419)

	return jitter(pace)
}

func StartJobSchedulerDaemon(
	ctx context.Context,
	stopper *stop.Stopper,
	registry *metric.Registry,
	cfg *scheduledjobs.JobExecutionConfig,
	env scheduledjobs.JobSchedulerEnv,
) {
	__antithesis_instrumentation__.Notify(70433)
	schedulerEnv := env
	var daemonKnobs *TestingKnobs
	if jobsKnobs, ok := cfg.TestingKnobs.(*TestingKnobs); ok {
		__antithesis_instrumentation__.Notify(70439)
		daemonKnobs = jobsKnobs
	} else {
		__antithesis_instrumentation__.Notify(70440)
	}
	__antithesis_instrumentation__.Notify(70434)

	if daemonKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(70441)
		return daemonKnobs.CaptureJobExecutionConfig != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70442)
		daemonKnobs.CaptureJobExecutionConfig(cfg)
	} else {
		__antithesis_instrumentation__.Notify(70443)
	}
	__antithesis_instrumentation__.Notify(70435)
	if daemonKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(70444)
		return daemonKnobs.JobSchedulerEnv != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70445)
		schedulerEnv = daemonKnobs.JobSchedulerEnv
	} else {
		__antithesis_instrumentation__.Notify(70446)
	}
	__antithesis_instrumentation__.Notify(70436)

	scheduler := newJobScheduler(cfg, schedulerEnv, registry)

	if daemonKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(70447)
		return daemonKnobs.TakeOverJobsScheduling != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70448)
		daemonKnobs.TakeOverJobsScheduling(
			func(ctx context.Context, maxSchedules int64) error {
				__antithesis_instrumentation__.Notify(70450)
				return scheduler.executeSchedules(ctx, maxSchedules)
			})
		__antithesis_instrumentation__.Notify(70449)
		return
	} else {
		__antithesis_instrumentation__.Notify(70451)
	}
	__antithesis_instrumentation__.Notify(70437)

	if daemonKnobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(70452)
		return daemonKnobs.CaptureJobScheduler != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(70453)
		daemonKnobs.CaptureJobScheduler(scheduler)
	} else {
		__antithesis_instrumentation__.Notify(70454)
	}
	__antithesis_instrumentation__.Notify(70438)

	scheduler.runDaemon(ctx, stopper)
}
