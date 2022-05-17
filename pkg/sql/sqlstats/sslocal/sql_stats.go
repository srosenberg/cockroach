package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type SQLStats struct {
	st *cluster.Settings

	uniqueStmtFingerprintLimit *settings.IntSetting

	uniqueTxnFingerprintLimit *settings.IntSetting

	mu struct {
		syncutil.Mutex

		mon *mon.BytesMonitor

		lastReset time.Time

		apps map[string]*ssmemstorage.Container
	}

	atomic struct {
		uniqueStmtFingerprintCount int64

		uniqueTxnFingerprintCount int64
	}

	flushTarget Sink

	knobs *sqlstats.TestingKnobs
}

func newSQLStats(
	st *cluster.Settings,
	uniqueStmtFingerprintLimit *settings.IntSetting,
	uniqueTxnFingerprintLimit *settings.IntSetting,
	curMemBytesCount *metric.Gauge,
	maxMemBytesHist *metric.Histogram,
	parentMon *mon.BytesMonitor,
	flushTarget Sink,
	knobs *sqlstats.TestingKnobs,
) *SQLStats {
	__antithesis_instrumentation__.Notify(625355)
	monitor := mon.NewMonitor(
		"SQLStats",
		mon.MemoryResource,
		curMemBytesCount,
		maxMemBytesHist,
		-1,
		math.MaxInt64,
		st,
	)
	s := &SQLStats{
		st:                         st,
		uniqueStmtFingerprintLimit: uniqueStmtFingerprintLimit,
		uniqueTxnFingerprintLimit:  uniqueTxnFingerprintLimit,
		flushTarget:                flushTarget,
		knobs:                      knobs,
	}
	s.mu.apps = make(map[string]*ssmemstorage.Container)
	s.mu.mon = monitor
	s.mu.mon.Start(context.Background(), parentMon, mon.BoundAccount{})
	return s
}

func (s *SQLStats) GetTotalFingerprintCount() int64 {
	__antithesis_instrumentation__.Notify(625356)
	return atomic.LoadInt64(&s.atomic.uniqueStmtFingerprintCount) + atomic.LoadInt64(&s.atomic.uniqueTxnFingerprintCount)
}

func (s *SQLStats) GetTotalFingerprintBytes() int64 {
	__antithesis_instrumentation__.Notify(625357)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.mon.AllocBytes()
}

func (s *SQLStats) getStatsForApplication(appName string) *ssmemstorage.Container {
	__antithesis_instrumentation__.Notify(625358)
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.mu.apps[appName]; ok {
		__antithesis_instrumentation__.Notify(625360)
		return a
	} else {
		__antithesis_instrumentation__.Notify(625361)
	}
	__antithesis_instrumentation__.Notify(625359)
	a := ssmemstorage.New(
		s.st,
		s.uniqueStmtFingerprintLimit,
		s.uniqueTxnFingerprintLimit,
		&s.atomic.uniqueStmtFingerprintCount,
		&s.atomic.uniqueTxnFingerprintCount,
		s.mu.mon,
		appName,
		s.knobs,
	)
	s.mu.apps[appName] = a
	return a
}

func (s *SQLStats) resetAndMaybeDumpStats(ctx context.Context, target Sink) (err error) {
	__antithesis_instrumentation__.Notify(625362)

	s.mu.Lock()

	for appName, statsContainer := range s.mu.apps {
		__antithesis_instrumentation__.Notify(625364)

		if sqlstats.DumpStmtStatsToLogBeforeReset.Get(&s.st.SV) {
			__antithesis_instrumentation__.Notify(625367)
			statsContainer.SaveToLog(ctx, appName)
		} else {
			__antithesis_instrumentation__.Notify(625368)
		}
		__antithesis_instrumentation__.Notify(625365)

		if target != nil {
			__antithesis_instrumentation__.Notify(625369)
			lastErr := target.AddAppStats(ctx, appName, statsContainer)

			if lastErr != nil {
				__antithesis_instrumentation__.Notify(625370)
				err = lastErr
			} else {
				__antithesis_instrumentation__.Notify(625371)
			}
		} else {
			__antithesis_instrumentation__.Notify(625372)
		}
		__antithesis_instrumentation__.Notify(625366)

		statsContainer.Clear(ctx)
	}
	__antithesis_instrumentation__.Notify(625363)
	s.mu.lastReset = timeutil.Now()
	s.mu.Unlock()

	return err
}
