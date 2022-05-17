package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func New(
	settings *cluster.Settings,
	maxStmtFingerprints *settings.IntSetting,
	maxTxnFingerprints *settings.IntSetting,
	curMemoryBytesCount *metric.Gauge,
	maxMemoryBytesHist *metric.Histogram,
	pool *mon.BytesMonitor,
	reportingSink Sink,
	knobs *sqlstats.TestingKnobs,
) *SQLStats {
	__antithesis_instrumentation__.Notify(625403)
	return newSQLStats(settings, maxStmtFingerprints, maxTxnFingerprints,
		curMemoryBytesCount, maxMemoryBytesHist, pool,
		reportingSink, knobs)
}

var _ sqlstats.Provider = &SQLStats{}

func (s *SQLStats) GetController(
	server serverpb.SQLStatusServer, db *kv.DB, ie sqlutil.InternalExecutor,
) *Controller {
	__antithesis_instrumentation__.Notify(625404)
	return NewController(s, server)
}

func (s *SQLStats) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(625405)

	_ = stopper.RunAsyncTask(ctx, "sql-stats-clearer", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(625406)
		var timer timeutil.Timer
		for {
			__antithesis_instrumentation__.Notify(625407)
			s.mu.Lock()
			last := s.mu.lastReset
			s.mu.Unlock()

			next := last.Add(sqlstats.MaxSQLStatReset.Get(&s.st.SV))
			wait := next.Sub(timeutil.Now())
			if wait < 0 {
				__antithesis_instrumentation__.Notify(625408)
				err := s.Reset(ctx)
				if err != nil {
					__antithesis_instrumentation__.Notify(625409)
					if log.V(1) {
						__antithesis_instrumentation__.Notify(625410)
						log.Warningf(ctx, "unexpected error: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(625411)
					}
				} else {
					__antithesis_instrumentation__.Notify(625412)
				}
			} else {
				__antithesis_instrumentation__.Notify(625413)
				timer.Reset(wait)
				select {
				case <-stopper.ShouldQuiesce():
					__antithesis_instrumentation__.Notify(625414)
					return
				case <-timer.C:
					__antithesis_instrumentation__.Notify(625415)
					timer.Read = true
				}
			}
		}
	})
}

func (s *SQLStats) GetApplicationStats(appName string) sqlstats.ApplicationStats {
	__antithesis_instrumentation__.Notify(625416)
	s.mu.Lock()
	defer s.mu.Unlock()
	if a, ok := s.mu.apps[appName]; ok {
		__antithesis_instrumentation__.Notify(625418)
		return a
	} else {
		__antithesis_instrumentation__.Notify(625419)
	}
	__antithesis_instrumentation__.Notify(625417)
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

func (s *SQLStats) GetLastReset() time.Time {
	__antithesis_instrumentation__.Notify(625420)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.mu.lastReset
}

func (s *SQLStats) IterateStatementStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.StatementVisitor,
) error {
	__antithesis_instrumentation__.Notify(625421)
	iter := s.StmtStatsIterator(options)

	for iter.Next() {
		__antithesis_instrumentation__.Notify(625423)
		if err := visitor(ctx, iter.Cur()); err != nil {
			__antithesis_instrumentation__.Notify(625424)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625425)
		}
	}
	__antithesis_instrumentation__.Notify(625422)

	return nil
}

func (s *SQLStats) StmtStatsIterator(options *sqlstats.IteratorOptions) *StmtStatsIterator {
	__antithesis_instrumentation__.Notify(625426)
	return NewStmtStatsIterator(s, options)
}

func (s *SQLStats) IterateTransactionStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.TransactionVisitor,
) error {
	__antithesis_instrumentation__.Notify(625427)
	iter := s.TxnStatsIterator(options)

	for iter.Next() {
		__antithesis_instrumentation__.Notify(625429)
		stats := iter.Cur()
		if err := visitor(ctx, stats); err != nil {
			__antithesis_instrumentation__.Notify(625430)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625431)
		}
	}
	__antithesis_instrumentation__.Notify(625428)

	return nil
}

func (s *SQLStats) TxnStatsIterator(options *sqlstats.IteratorOptions) *TxnStatsIterator {
	__antithesis_instrumentation__.Notify(625432)
	return NewTxnStatsIterator(s, options)
}

func (s *SQLStats) IterateAggregatedTransactionStats(
	ctx context.Context,
	options *sqlstats.IteratorOptions,
	visitor sqlstats.AggregatedTransactionVisitor,
) error {
	__antithesis_instrumentation__.Notify(625433)
	appNames := s.getAppNames(options.SortedAppNames)

	for _, appName := range appNames {
		__antithesis_instrumentation__.Notify(625435)
		statsContainer := s.getStatsForApplication(appName)

		err := statsContainer.IterateAggregatedTransactionStats(ctx, options, visitor)
		if err != nil {
			__antithesis_instrumentation__.Notify(625436)
			return errors.Wrap(err, "sql stats iteration abort")
		} else {
			__antithesis_instrumentation__.Notify(625437)
		}
	}
	__antithesis_instrumentation__.Notify(625434)

	return nil
}

func (s *SQLStats) Reset(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(625438)
	return s.resetAndMaybeDumpStats(ctx, s.flushTarget)
}

func (s *SQLStats) getAppNames(sorted bool) []string {
	__antithesis_instrumentation__.Notify(625439)
	var appNames []string
	s.mu.Lock()
	for n := range s.mu.apps {
		__antithesis_instrumentation__.Notify(625442)
		appNames = append(appNames, n)
	}
	__antithesis_instrumentation__.Notify(625440)
	s.mu.Unlock()
	if sorted {
		__antithesis_instrumentation__.Notify(625443)
		sort.Strings(appNames)
	} else {
		__antithesis_instrumentation__.Notify(625444)
	}
	__antithesis_instrumentation__.Notify(625441)

	return appNames
}
