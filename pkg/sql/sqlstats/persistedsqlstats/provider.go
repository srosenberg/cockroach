package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/sslocal"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type Config struct {
	Settings         *cluster.Settings
	InternalExecutor sqlutil.InternalExecutor
	KvDB             *kv.DB
	SQLIDContainer   *base.SQLIDContainer
	JobRegistry      *jobs.Registry

	FlushCounter   *metric.Counter
	FlushDuration  *metric.Histogram
	FailureCounter *metric.Counter

	Knobs *sqlstats.TestingKnobs
}

type PersistedSQLStats struct {
	*sslocal.SQLStats

	cfg *Config

	memoryPressureSignal chan struct{}

	lastFlushStarted time.Time
	jobMonitor       jobMonitor

	atomic struct {
		nextFlushAt atomic.Value
	}
}

var _ sqlstats.Provider = &PersistedSQLStats{}

func New(cfg *Config, memSQLStats *sslocal.SQLStats) *PersistedSQLStats {
	__antithesis_instrumentation__.Notify(624845)
	p := &PersistedSQLStats{
		SQLStats:             memSQLStats,
		cfg:                  cfg,
		memoryPressureSignal: make(chan struct{}),
	}

	p.jobMonitor = jobMonitor{
		st:           cfg.Settings,
		ie:           cfg.InternalExecutor,
		db:           cfg.KvDB,
		scanInterval: defaultScanInterval,
		jitterFn:     p.jitterInterval,
	}

	return p
}

func (s *PersistedSQLStats) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(624846)
	s.startSQLStatsFlushLoop(ctx, stopper)
	s.jobMonitor.start(ctx, stopper)
}

func (s *PersistedSQLStats) GetController(server serverpb.SQLStatusServer) *Controller {
	__antithesis_instrumentation__.Notify(624847)
	return NewController(s, server, s.cfg.KvDB, s.cfg.InternalExecutor)
}

func (s *PersistedSQLStats) startSQLStatsFlushLoop(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(624848)
	_ = stopper.RunAsyncTask(ctx, "sql-stats-worker", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(624849)
		var resetIntervalChanged = make(chan struct{}, 1)

		SQLStatsFlushInterval.SetOnChange(&s.cfg.Settings.SV, func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(624851)
			select {
			case resetIntervalChanged <- struct{}{}:
				__antithesis_instrumentation__.Notify(624852)
			default:
				__antithesis_instrumentation__.Notify(624853)
			}
		})
		__antithesis_instrumentation__.Notify(624850)

		initialDelay := s.nextFlushInterval()
		timer := timeutil.NewTimer()
		timer.Reset(initialDelay)

		log.Infof(ctx, "starting sql-stats-worker with initial delay: %s", initialDelay)
		for {
			__antithesis_instrumentation__.Notify(624854)
			waitInterval := s.nextFlushInterval()
			timer.Reset(waitInterval)

			select {
			case <-timer.C:
				__antithesis_instrumentation__.Notify(624856)
				timer.Read = true
			case <-s.memoryPressureSignal:
				__antithesis_instrumentation__.Notify(624857)

			case <-resetIntervalChanged:
				__antithesis_instrumentation__.Notify(624858)

				continue
			case <-stopper.ShouldQuiesce():
				__antithesis_instrumentation__.Notify(624859)
				return
			}
			__antithesis_instrumentation__.Notify(624855)

			s.Flush(ctx)
		}
	})
}

func (s *PersistedSQLStats) GetLocalMemProvider() sqlstats.Provider {
	__antithesis_instrumentation__.Notify(624860)
	return s.SQLStats
}

func (s *PersistedSQLStats) GetNextFlushAt() time.Time {
	__antithesis_instrumentation__.Notify(624861)
	return s.atomic.nextFlushAt.Load().(time.Time)
}

func (s *PersistedSQLStats) nextFlushInterval() time.Duration {
	__antithesis_instrumentation__.Notify(624862)
	baseInterval := SQLStatsFlushInterval.Get(&s.cfg.Settings.SV)
	waitInterval := s.jitterInterval(baseInterval)

	nextFlushAt := s.getTimeNow().Add(waitInterval)
	s.atomic.nextFlushAt.Store(nextFlushAt)

	return waitInterval
}

func (s *PersistedSQLStats) jitterInterval(interval time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(624863)
	jitter := SQLStatsFlushJitter.Get(&s.cfg.Settings.SV)
	frac := 1 + (2*rand.Float64()-1)*jitter

	jitteredInterval := time.Duration(frac * float64(interval.Nanoseconds()))
	return jitteredInterval
}

func (s *PersistedSQLStats) GetApplicationStats(appName string) sqlstats.ApplicationStats {
	__antithesis_instrumentation__.Notify(624864)
	appStats := s.SQLStats.GetApplicationStats(appName)
	return &ApplicationStats{
		ApplicationStats:     appStats,
		memoryPressureSignal: s.memoryPressureSignal,
	}
}
