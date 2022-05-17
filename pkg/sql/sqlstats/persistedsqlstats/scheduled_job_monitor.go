package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/scheduledjobs"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var defaultScanInterval = time.Hour * 6

var (
	errScheduleNotFound = errors.New("sql stats compaction schedule not found")

	ErrScheduleIntervalTooLong = errors.New("sql stats compaction schedule interval too long")

	ErrSchedulePaused = errors.New("sql stats compaction schedule paused")

	ErrScheduleUndroppable = errors.New("sql stats compaction schedule cannot be dropped")
)

var longIntervalWarningThreshold = time.Hour * 24

type jobMonitor struct {
	st           *cluster.Settings
	ie           sqlutil.InternalExecutor
	db           *kv.DB
	scanInterval time.Duration
	jitterFn     func(time.Duration) time.Duration
}

func (j *jobMonitor) start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(624865)
	j.ensureSchedule(ctx)
	j.registerClusterSettingHook()

	_ = stopper.RunAsyncTask(ctx, "sql-stats-scheduled-compaction-job-monitor", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(624866)
		for timer := timeutil.NewTimer(); ; timer.Reset(j.jitterFn(j.scanInterval)) {
			__antithesis_instrumentation__.Notify(624867)
			select {
			case <-timer.C:
				__antithesis_instrumentation__.Notify(624869)
				timer.Read = true
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(624870)
				return
			}
			__antithesis_instrumentation__.Notify(624868)
			j.ensureSchedule(ctx)
		}
	})
}

func (j *jobMonitor) registerClusterSettingHook() {
	__antithesis_instrumentation__.Notify(624871)
	SQLStatsCleanupRecurrence.SetOnChange(&j.st.SV, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(624872)
		j.ensureSchedule(ctx)
		if err := j.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(624873)
			sj, err := j.getSchedule(ctx, txn)
			if err != nil {
				__antithesis_instrumentation__.Notify(624877)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624878)
			}
			__antithesis_instrumentation__.Notify(624874)
			cronExpr := SQLStatsCleanupRecurrence.Get(&j.st.SV)
			if err = sj.SetSchedule(cronExpr); err != nil {
				__antithesis_instrumentation__.Notify(624879)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624880)
			}
			__antithesis_instrumentation__.Notify(624875)
			if err = CheckScheduleAnomaly(sj); err != nil {
				__antithesis_instrumentation__.Notify(624881)
				log.Warningf(ctx, "schedule anomaly detected, disabled sql stats compaction may cause performance impact: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(624882)
			}
			__antithesis_instrumentation__.Notify(624876)
			return sj.Update(ctx, j.ie, txn)
		}); err != nil {
			__antithesis_instrumentation__.Notify(624883)
			log.Errorf(ctx, "unable to find sqlstats clean up schedule: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(624884)
		}
	})
}

func (j *jobMonitor) getSchedule(
	ctx context.Context, txn *kv.Txn,
) (sj *jobs.ScheduledJob, _ error) {
	__antithesis_instrumentation__.Notify(624885)
	row, err := j.ie.QueryRowEx(
		ctx,
		"load-sql-stats-scheduled-job",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		"SELECT schedule_id FROM system.scheduled_jobs WHERE schedule_name = $1",
		compactionScheduleName,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(624889)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624890)
	}
	__antithesis_instrumentation__.Notify(624886)

	if row == nil {
		__antithesis_instrumentation__.Notify(624891)
		return nil, errScheduleNotFound
	} else {
		__antithesis_instrumentation__.Notify(624892)
	}
	__antithesis_instrumentation__.Notify(624887)

	scheduledJobID := int64(tree.MustBeDInt(row[0]))

	sj, err = jobs.LoadScheduledJob(ctx, scheduledjobs.ProdJobSchedulerEnv, scheduledJobID, j.ie, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(624893)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(624894)
	}
	__antithesis_instrumentation__.Notify(624888)

	return sj, nil
}

func (j *jobMonitor) ensureSchedule(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624895)
	var sj *jobs.ScheduledJob
	var err error
	if err = j.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624897)

		sj, err = j.getSchedule(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624899)
			if !jobs.HasScheduledJobNotFoundError(err) && func() bool {
				__antithesis_instrumentation__.Notify(624901)
				return !errors.Is(err, errScheduleNotFound) == true
			}() == true {
				__antithesis_instrumentation__.Notify(624902)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624903)
			}
			__antithesis_instrumentation__.Notify(624900)
			sj, err = CreateSQLStatsCompactionScheduleIfNotYetExist(ctx, j.ie, txn, j.st)
			if err != nil {
				__antithesis_instrumentation__.Notify(624904)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624905)
			}
		} else {
			__antithesis_instrumentation__.Notify(624906)
		}
		__antithesis_instrumentation__.Notify(624898)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(624907)
		log.Errorf(ctx, "fail to ensure sql stats scheduled compaction job is created: %s", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(624908)
	}
	__antithesis_instrumentation__.Notify(624896)

	if err = CheckScheduleAnomaly(sj); err != nil {
		__antithesis_instrumentation__.Notify(624909)
		log.Warningf(ctx, "schedule anomaly detected: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(624910)
	}
}

func CheckScheduleAnomaly(sj *jobs.ScheduledJob) error {
	__antithesis_instrumentation__.Notify(624911)
	if (sj.NextRun() == time.Time{}) {
		__antithesis_instrumentation__.Notify(624914)
		return ErrSchedulePaused
	} else {
		__antithesis_instrumentation__.Notify(624915)
	}
	__antithesis_instrumentation__.Notify(624912)

	if nextRunInterval := sj.NextRun().Sub(timeutil.Now()); nextRunInterval > longIntervalWarningThreshold {
		__antithesis_instrumentation__.Notify(624916)
		return errors.Wrapf(ErrScheduleIntervalTooLong, "sql stats compaction schedule next run interval "+
			"(%s) exceeds warning threshold (%s)", nextRunInterval,
			longIntervalWarningThreshold)
	} else {
		__antithesis_instrumentation__.Notify(624917)
	}
	__antithesis_instrumentation__.Notify(624913)
	return nil
}
