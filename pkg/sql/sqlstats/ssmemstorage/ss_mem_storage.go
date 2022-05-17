package ssmemstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type stmtKey struct {
	sampledPlanKey
	planHash                 uint64
	transactionFingerprintID roachpb.TransactionFingerprintID
}

type sampledPlanKey struct {
	anonymizedStmt string
	failed         bool
	implicitTxn    bool
	database       string
}

func (p sampledPlanKey) size() int64 {
	__antithesis_instrumentation__.Notify(625515)
	return int64(unsafe.Sizeof(p)) + int64(len(p.anonymizedStmt)) + int64(len(p.database))
}

func (s stmtKey) String() string {
	__antithesis_instrumentation__.Notify(625516)
	if s.failed {
		__antithesis_instrumentation__.Notify(625518)
		return "!" + s.anonymizedStmt
	} else {
		__antithesis_instrumentation__.Notify(625519)
	}
	__antithesis_instrumentation__.Notify(625517)
	return s.anonymizedStmt
}

func (s stmtKey) size() int64 {
	__antithesis_instrumentation__.Notify(625520)
	return s.sampledPlanKey.size() + int64(unsafe.Sizeof(invalidStmtFingerprintID))
}

const invalidStmtFingerprintID = 0

type Container struct {
	st      *cluster.Settings
	appName string

	uniqueStmtFingerprintLimit *settings.IntSetting

	uniqueTxnFingerprintLimit *settings.IntSetting

	atomic struct {
		uniqueStmtFingerprintCount *int64

		uniqueTxnFingerprintCount *int64
	}

	mu struct {
		syncutil.Mutex

		acc mon.BoundAccount

		stmts map[stmtKey]*stmtStats
		txns  map[roachpb.TransactionFingerprintID]*txnStats

		sampledPlanMetadataCache map[sampledPlanKey]time.Time
	}

	txnCounts transactionCounts
	mon       *mon.BytesMonitor

	knobs *sqlstats.TestingKnobs
}

var _ sqlstats.ApplicationStats = &Container{}

func New(
	st *cluster.Settings,
	uniqueStmtFingerprintLimit *settings.IntSetting,
	uniqueTxnFingerprintLimit *settings.IntSetting,
	uniqueStmtFingerprintCount *int64,
	uniqueTxnFingerprintCount *int64,
	mon *mon.BytesMonitor,
	appName string,
	knobs *sqlstats.TestingKnobs,
) *Container {
	__antithesis_instrumentation__.Notify(625521)
	s := &Container{
		st:                         st,
		appName:                    appName,
		uniqueStmtFingerprintLimit: uniqueStmtFingerprintLimit,
		uniqueTxnFingerprintLimit:  uniqueTxnFingerprintLimit,
		mon:                        mon,
		knobs:                      knobs,
	}

	if mon != nil {
		__antithesis_instrumentation__.Notify(625523)
		s.mu.acc = mon.MakeBoundAccount()
	} else {
		__antithesis_instrumentation__.Notify(625524)
	}
	__antithesis_instrumentation__.Notify(625522)

	s.mu.stmts = make(map[stmtKey]*stmtStats)
	s.mu.txns = make(map[roachpb.TransactionFingerprintID]*txnStats)
	s.mu.sampledPlanMetadataCache = make(map[sampledPlanKey]time.Time)

	s.atomic.uniqueStmtFingerprintCount = uniqueStmtFingerprintCount
	s.atomic.uniqueTxnFingerprintCount = uniqueTxnFingerprintCount

	return s
}

func (s *Container) IterateAggregatedTransactionStats(
	_ context.Context, _ *sqlstats.IteratorOptions, visitor sqlstats.AggregatedTransactionVisitor,
) error {
	__antithesis_instrumentation__.Notify(625525)
	var txnStat roachpb.TxnStats
	s.txnCounts.mu.Lock()
	txnStat = s.txnCounts.mu.TxnStats
	s.txnCounts.mu.Unlock()

	err := visitor(s.appName, &txnStat)
	if err != nil {
		__antithesis_instrumentation__.Notify(625527)
		return errors.Wrap(err, "sql stats iteration abort")
	} else {
		__antithesis_instrumentation__.Notify(625528)
	}
	__antithesis_instrumentation__.Notify(625526)

	return nil
}

func (s *Container) StmtStatsIterator(options *sqlstats.IteratorOptions) *StmtStatsIterator {
	__antithesis_instrumentation__.Notify(625529)
	return NewStmtStatsIterator(s, options)
}

func (s *Container) TxnStatsIterator(options *sqlstats.IteratorOptions) *TxnStatsIterator {
	__antithesis_instrumentation__.Notify(625530)
	return NewTxnStatsIterator(s, options)
}

func (s *Container) IterateStatementStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.StatementVisitor,
) error {
	__antithesis_instrumentation__.Notify(625531)
	iter := s.StmtStatsIterator(options)

	for iter.Next() {
		__antithesis_instrumentation__.Notify(625533)
		if err := visitor(ctx, iter.Cur()); err != nil {
			__antithesis_instrumentation__.Notify(625534)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625535)
		}
	}
	__antithesis_instrumentation__.Notify(625532)

	return nil
}

func (s *Container) IterateTransactionStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.TransactionVisitor,
) error {
	__antithesis_instrumentation__.Notify(625536)
	iter := s.TxnStatsIterator(options)

	for iter.Next() {
		__antithesis_instrumentation__.Notify(625538)
		stats := iter.Cur()
		if err := visitor(ctx, stats); err != nil {
			__antithesis_instrumentation__.Notify(625539)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625540)
		}
	}
	__antithesis_instrumentation__.Notify(625537)

	return nil
}

func NewTempContainerFromExistingStmtStats(
	statistics []serverpb.StatementsResponse_CollectedStatementStatistics,
) (
	container *Container,
	remaining []serverpb.StatementsResponse_CollectedStatementStatistics,
	err error,
) {
	__antithesis_instrumentation__.Notify(625541)
	if len(statistics) == 0 {
		__antithesis_instrumentation__.Notify(625544)
		return nil, statistics, nil
	} else {
		__antithesis_instrumentation__.Notify(625545)
	}
	__antithesis_instrumentation__.Notify(625542)

	appName := statistics[0].Key.KeyData.App

	container = New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appName,
		nil,
	)

	for i := range statistics {
		__antithesis_instrumentation__.Notify(625546)
		if currentAppName := statistics[i].Key.KeyData.App; currentAppName != appName {
			__antithesis_instrumentation__.Notify(625551)
			return container, statistics[i:], nil
		} else {
			__antithesis_instrumentation__.Notify(625552)
		}
		__antithesis_instrumentation__.Notify(625547)
		key := stmtKey{
			sampledPlanKey: sampledPlanKey{
				anonymizedStmt: statistics[i].Key.KeyData.Query,
				failed:         statistics[i].Key.KeyData.Failed,
				implicitTxn:    statistics[i].Key.KeyData.ImplicitTxn,
				database:       statistics[i].Key.KeyData.Database,
			},
			planHash:                 statistics[i].Key.KeyData.PlanHash,
			transactionFingerprintID: statistics[i].Key.KeyData.TransactionFingerprintID,
		}
		stmtStats, _, throttled :=
			container.getStatsForStmtWithKeyLocked(key, statistics[i].ID, true)
		if throttled {
			__antithesis_instrumentation__.Notify(625553)
			return nil, nil, ErrFingerprintLimitReached
		} else {
			__antithesis_instrumentation__.Notify(625554)
		}
		__antithesis_instrumentation__.Notify(625548)

		stmtStats.mu.data.Add(&statistics[i].Stats)

		if stmtStats.mu.data.SensitiveInfo.LastErr == "" && func() bool {
			__antithesis_instrumentation__.Notify(625555)
			return key.failed == true
		}() == true {
			__antithesis_instrumentation__.Notify(625556)
			stmtStats.mu.data.SensitiveInfo.LastErr = statistics[i].Stats.SensitiveInfo.LastErr
		} else {
			__antithesis_instrumentation__.Notify(625557)
		}
		__antithesis_instrumentation__.Notify(625549)

		if stmtStats.mu.data.SensitiveInfo.MostRecentPlanTimestamp.Before(statistics[i].Stats.SensitiveInfo.MostRecentPlanTimestamp) {
			__antithesis_instrumentation__.Notify(625558)
			stmtStats.mu.data.SensitiveInfo.MostRecentPlanDescription = statistics[i].Stats.SensitiveInfo.MostRecentPlanDescription
			stmtStats.mu.data.SensitiveInfo.MostRecentPlanTimestamp = statistics[i].Stats.SensitiveInfo.MostRecentPlanTimestamp
		} else {
			__antithesis_instrumentation__.Notify(625559)
		}
		__antithesis_instrumentation__.Notify(625550)

		stmtStats.mu.vectorized = statistics[i].Key.KeyData.Vec
		stmtStats.mu.distSQLUsed = statistics[i].Key.KeyData.DistSQL
		stmtStats.mu.fullScan = statistics[i].Key.KeyData.FullScan
		stmtStats.mu.database = statistics[i].Key.KeyData.Database
		stmtStats.mu.querySummary = statistics[i].Key.KeyData.QuerySummary
	}
	__antithesis_instrumentation__.Notify(625543)

	return container, nil, nil
}

func NewTempContainerFromExistingTxnStats(
	statistics []serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics,
) (
	container *Container,
	remaining []serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics,
	err error,
) {
	__antithesis_instrumentation__.Notify(625560)
	if len(statistics) == 0 {
		__antithesis_instrumentation__.Notify(625563)
		return nil, statistics, nil
	} else {
		__antithesis_instrumentation__.Notify(625564)
	}
	__antithesis_instrumentation__.Notify(625561)

	appName := statistics[0].StatsData.App

	container = New(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		appName,
		nil,
	)

	for i := range statistics {
		__antithesis_instrumentation__.Notify(625565)
		if currentAppName := statistics[i].StatsData.App; currentAppName != appName {
			__antithesis_instrumentation__.Notify(625568)
			return container, statistics[i:], nil
		} else {
			__antithesis_instrumentation__.Notify(625569)
		}
		__antithesis_instrumentation__.Notify(625566)
		txnStats, _, throttled :=
			container.getStatsForTxnWithKeyLocked(
				statistics[i].StatsData.TransactionFingerprintID,
				statistics[i].StatsData.StatementFingerprintIDs,
				true)
		if throttled {
			__antithesis_instrumentation__.Notify(625570)
			return nil, nil, ErrFingerprintLimitReached
		} else {
			__antithesis_instrumentation__.Notify(625571)
		}
		__antithesis_instrumentation__.Notify(625567)
		txnStats.mu.data.Add(&statistics[i].StatsData.Stats)
	}
	__antithesis_instrumentation__.Notify(625562)

	return container, nil, nil
}

func (s *Container) NewApplicationStatsWithInheritedOptions() sqlstats.ApplicationStats {
	__antithesis_instrumentation__.Notify(625572)
	var (
		uniqueStmtFingerprintCount int64
		uniqueTxnFingerprintCount  int64
	)
	s.mu.Lock()
	defer s.mu.Unlock()
	return New(
		s.st,
		sqlstats.MaxSQLStatsStmtFingerprintsPerExplicitTxn,

		nil,
		&uniqueStmtFingerprintCount,
		&uniqueTxnFingerprintCount,
		s.mon,
		s.appName,
		s.knobs,
	)
}

type txnStats struct {
	statementFingerprintIDs []roachpb.StmtFingerprintID

	mu struct {
		syncutil.Mutex

		data roachpb.TransactionStatistics
	}
}

func (t *txnStats) sizeUnsafe() int64 {
	__antithesis_instrumentation__.Notify(625573)
	const txnStatsShallowSize = int64(unsafe.Sizeof(txnStats{}))
	stmtFingerprintIDsSize := int64(cap(t.statementFingerprintIDs)) *
		int64(unsafe.Sizeof(roachpb.StmtFingerprintID(0)))

	dataSize := -int64(unsafe.Sizeof(roachpb.TransactionStatistics{})) +
		int64(t.mu.data.Size())

	return txnStatsShallowSize + stmtFingerprintIDsSize + dataSize
}

func (t *txnStats) mergeStats(stats *roachpb.TransactionStatistics) {
	__antithesis_instrumentation__.Notify(625574)
	t.mu.Lock()
	defer t.mu.Unlock()
	t.mu.data.Add(stats)
}

type stmtStats struct {
	ID roachpb.StmtFingerprintID

	mu struct {
		syncutil.Mutex

		distSQLUsed bool

		vectorized bool

		fullScan bool

		database string

		querySummary string

		data roachpb.StatementStatistics
	}
}

func (s *stmtStats) sizeUnsafe() int64 {
	__antithesis_instrumentation__.Notify(625575)
	const stmtStatsShallowSize = int64(unsafe.Sizeof(stmtStats{}))
	databaseNameSize := int64(len(s.mu.database))

	dataSize := -int64(unsafe.Sizeof(roachpb.StatementStatistics{})) +
		int64(s.mu.data.Size())

	return stmtStatsShallowSize + databaseNameSize + dataSize
}

func (s *stmtStats) recordExecStats(stats execstats.QueryLevelStats) {
	__antithesis_instrumentation__.Notify(625576)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.mu.data.ExecStats.Count++
	count := s.mu.data.ExecStats.Count
	s.mu.data.ExecStats.NetworkBytes.Record(count, float64(stats.NetworkBytesSent))
	s.mu.data.ExecStats.MaxMemUsage.Record(count, float64(stats.MaxMemUsage))
	s.mu.data.ExecStats.ContentionTime.Record(count, stats.ContentionTime.Seconds())
	s.mu.data.ExecStats.NetworkMessages.Record(count, float64(stats.NetworkMessages))
	s.mu.data.ExecStats.MaxDiskUsage.Record(count, float64(stats.MaxDiskUsage))
}

func (s *stmtStats) mergeStatsLocked(statistics *roachpb.CollectedStatementStatistics) {
	__antithesis_instrumentation__.Notify(625577)

	s.mu.data.Add(&statistics.Stats)

	if s.mu.data.SensitiveInfo.LastErr == "" && func() bool {
		__antithesis_instrumentation__.Notify(625580)
		return statistics.Key.Failed == true
	}() == true {
		__antithesis_instrumentation__.Notify(625581)
		s.mu.data.SensitiveInfo.LastErr = statistics.Stats.SensitiveInfo.LastErr
	} else {
		__antithesis_instrumentation__.Notify(625582)
	}
	__antithesis_instrumentation__.Notify(625578)

	if s.mu.data.SensitiveInfo.MostRecentPlanTimestamp.Before(statistics.Stats.SensitiveInfo.MostRecentPlanTimestamp) {
		__antithesis_instrumentation__.Notify(625583)
		s.mu.data.SensitiveInfo.MostRecentPlanDescription = statistics.Stats.SensitiveInfo.MostRecentPlanDescription
		s.mu.data.SensitiveInfo.MostRecentPlanTimestamp = statistics.Stats.SensitiveInfo.MostRecentPlanTimestamp
	} else {
		__antithesis_instrumentation__.Notify(625584)
	}
	__antithesis_instrumentation__.Notify(625579)

	s.mu.vectorized = statistics.Key.Vec
	s.mu.distSQLUsed = statistics.Key.DistSQL
	s.mu.fullScan = statistics.Key.FullScan
	s.mu.database = statistics.Key.Database
}

func (s *Container) getStatsForStmt(
	anonymizedStmt string,
	implicitTxn bool,
	database string,
	failed bool,
	planHash uint64,
	transactionFingerprintID roachpb.TransactionFingerprintID,
	createIfNonexistent bool,
) (
	stats *stmtStats,
	key stmtKey,
	stmtFingerprintID roachpb.StmtFingerprintID,
	created bool,
	throttled bool,
) {
	__antithesis_instrumentation__.Notify(625585)

	key = stmtKey{
		sampledPlanKey: sampledPlanKey{
			anonymizedStmt: anonymizedStmt,
			failed:         failed,
			implicitTxn:    implicitTxn,
			database:       database,
		},
		planHash:                 planHash,
		transactionFingerprintID: transactionFingerprintID,
	}

	stats, _, _ = s.getStatsForStmtWithKey(key, invalidStmtFingerprintID, false)
	if stats == nil {
		__antithesis_instrumentation__.Notify(625587)
		stmtFingerprintID = constructStatementFingerprintIDFromStmtKey(key)
		stats, created, throttled = s.getStatsForStmtWithKey(key, stmtFingerprintID, createIfNonexistent)
		return stats, key, stmtFingerprintID, created, throttled
	} else {
		__antithesis_instrumentation__.Notify(625588)
	}
	__antithesis_instrumentation__.Notify(625586)
	return stats, key, stats.ID, false, false
}

func (s *Container) getStatsForStmtWithKey(
	key stmtKey, stmtFingerprintID roachpb.StmtFingerprintID, createIfNonexistent bool,
) (stats *stmtStats, created, throttled bool) {
	__antithesis_instrumentation__.Notify(625589)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.getStatsForStmtWithKeyLocked(key, stmtFingerprintID, createIfNonexistent)
}

func (s *Container) getStatsForStmtWithKeyLocked(
	key stmtKey, stmtFingerprintID roachpb.StmtFingerprintID, createIfNonexistent bool,
) (stats *stmtStats, created, throttled bool) {
	__antithesis_instrumentation__.Notify(625590)

	stats, ok := s.mu.stmts[key]
	if !ok && func() bool {
		__antithesis_instrumentation__.Notify(625592)
		return createIfNonexistent == true
	}() == true {
		__antithesis_instrumentation__.Notify(625593)

		if s.atomic.uniqueStmtFingerprintCount != nil {
			__antithesis_instrumentation__.Notify(625595)

			limit := s.uniqueStmtFingerprintLimit.Get(&s.st.SV)
			incrementedFingerprintCount :=
				atomic.AddInt64(s.atomic.uniqueStmtFingerprintCount, int64(1))

			if incrementedFingerprintCount > limit {
				__antithesis_instrumentation__.Notify(625596)
				atomic.AddInt64(s.atomic.uniqueStmtFingerprintCount, -int64(1))
				return stats, false, true
			} else {
				__antithesis_instrumentation__.Notify(625597)
			}
		} else {
			__antithesis_instrumentation__.Notify(625598)
		}
		__antithesis_instrumentation__.Notify(625594)
		stats = &stmtStats{}
		stats.ID = stmtFingerprintID
		s.mu.stmts[key] = stats
		s.mu.sampledPlanMetadataCache[key.sampledPlanKey] = s.getTimeNow()

		return stats, true, false
	} else {
		__antithesis_instrumentation__.Notify(625599)
	}
	__antithesis_instrumentation__.Notify(625591)
	return stats, false, false
}

func (s *Container) getStatsForTxnWithKey(
	key roachpb.TransactionFingerprintID,
	stmtFingerprintIDs []roachpb.StmtFingerprintID,
	createIfNonexistent bool,
) (stats *txnStats, created, throttled bool) {
	__antithesis_instrumentation__.Notify(625600)
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.getStatsForTxnWithKeyLocked(key, stmtFingerprintIDs, createIfNonexistent)
}

func (s *Container) getStatsForTxnWithKeyLocked(
	key roachpb.TransactionFingerprintID,
	stmtFingerprintIDs []roachpb.StmtFingerprintID,
	createIfNonexistent bool,
) (stats *txnStats, created, throttled bool) {
	__antithesis_instrumentation__.Notify(625601)

	stats, ok := s.mu.txns[key]
	if !ok && func() bool {
		__antithesis_instrumentation__.Notify(625603)
		return createIfNonexistent == true
	}() == true {
		__antithesis_instrumentation__.Notify(625604)

		if s.atomic.uniqueTxnFingerprintCount != nil {
			__antithesis_instrumentation__.Notify(625606)
			limit := s.uniqueTxnFingerprintLimit.Get(&s.st.SV)
			incrementedFingerprintCount :=
				atomic.AddInt64(s.atomic.uniqueTxnFingerprintCount, int64(1))

			if incrementedFingerprintCount > limit {
				__antithesis_instrumentation__.Notify(625607)
				atomic.AddInt64(s.atomic.uniqueTxnFingerprintCount, -int64(1))
				return nil, false, true
			} else {
				__antithesis_instrumentation__.Notify(625608)
			}
		} else {
			__antithesis_instrumentation__.Notify(625609)
		}
		__antithesis_instrumentation__.Notify(625605)
		stats = &txnStats{}
		stats.statementFingerprintIDs = stmtFingerprintIDs
		s.mu.txns[key] = stats
		return stats, true, false
	} else {
		__antithesis_instrumentation__.Notify(625610)
	}
	__antithesis_instrumentation__.Notify(625602)
	return stats, false, false
}

func (s *Container) SaveToLog(ctx context.Context, appName string) {
	__antithesis_instrumentation__.Notify(625611)
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.mu.stmts) == 0 {
		__antithesis_instrumentation__.Notify(625614)
		return
	} else {
		__antithesis_instrumentation__.Notify(625615)
	}
	__antithesis_instrumentation__.Notify(625612)
	var buf bytes.Buffer
	for key, stats := range s.mu.stmts {
		__antithesis_instrumentation__.Notify(625616)
		stats.mu.Lock()
		json, err := json.Marshal(stats.mu.data)
		s.mu.Unlock()
		if err != nil {
			__antithesis_instrumentation__.Notify(625618)
			log.Errorf(ctx, "error while marshaling stats for %q // %q: %v", appName, key.String(), err)
			continue
		} else {
			__antithesis_instrumentation__.Notify(625619)
		}
		__antithesis_instrumentation__.Notify(625617)
		fmt.Fprintf(&buf, "%q: %s\n", key.String(), json)
	}
	__antithesis_instrumentation__.Notify(625613)
	log.Infof(ctx, "statistics for %q:\n%s", appName, buf.String())
}

func (s *Container) Clear(ctx context.Context) {
	__antithesis_instrumentation__.Notify(625620)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.freeLocked(ctx)

	s.mu.stmts = make(map[stmtKey]*stmtStats, len(s.mu.stmts)/2)
	s.mu.txns = make(map[roachpb.TransactionFingerprintID]*txnStats, len(s.mu.txns)/2)
	s.mu.sampledPlanMetadataCache = make(map[sampledPlanKey]time.Time, len(s.mu.sampledPlanMetadataCache)/2)
}

func (s *Container) Free(ctx context.Context) {
	__antithesis_instrumentation__.Notify(625621)
	s.mu.Lock()
	defer s.mu.Unlock()

	s.freeLocked(ctx)
}

func (s *Container) freeLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(625622)
	atomic.AddInt64(s.atomic.uniqueStmtFingerprintCount, int64(-len(s.mu.stmts)))
	atomic.AddInt64(s.atomic.uniqueTxnFingerprintCount, int64(-len(s.mu.txns)))

	s.mu.acc.Clear(ctx)
}

func (s *Container) MergeApplicationStatementStats(
	ctx context.Context,
	other sqlstats.ApplicationStats,
	transformer func(*roachpb.CollectedStatementStatistics),
) (discardedStats uint64) {
	__antithesis_instrumentation__.Notify(625623)
	if err := other.IterateStatementStats(
		ctx,
		&sqlstats.IteratorOptions{},
		func(ctx context.Context, statistics *roachpb.CollectedStatementStatistics) error {
			__antithesis_instrumentation__.Notify(625625)
			if transformer != nil {
				__antithesis_instrumentation__.Notify(625629)
				transformer(statistics)
			} else {
				__antithesis_instrumentation__.Notify(625630)
			}
			__antithesis_instrumentation__.Notify(625626)
			key := stmtKey{
				sampledPlanKey: sampledPlanKey{
					anonymizedStmt: statistics.Key.Query,
					failed:         statistics.Key.Failed,
					implicitTxn:    statistics.Key.ImplicitTxn,
					database:       statistics.Key.Database,
				},
				planHash:                 statistics.Key.PlanHash,
				transactionFingerprintID: statistics.Key.TransactionFingerprintID,
			}

			stmtStats, _, throttled :=
				s.getStatsForStmtWithKey(key, statistics.ID, true)
			if throttled {
				__antithesis_instrumentation__.Notify(625631)
				discardedStats++
				return nil
			} else {
				__antithesis_instrumentation__.Notify(625632)
			}
			__antithesis_instrumentation__.Notify(625627)

			stmtStats.mu.Lock()
			defer stmtStats.mu.Unlock()

			stmtStats.mergeStatsLocked(statistics)
			planLastSampled := s.getLogicalPlanLastSampled(key.sampledPlanKey)
			if planLastSampled.Before(stmtStats.mu.data.SensitiveInfo.MostRecentPlanTimestamp) {
				__antithesis_instrumentation__.Notify(625633)
				s.setLogicalPlanLastSampled(key.sampledPlanKey, stmtStats.mu.data.SensitiveInfo.MostRecentPlanTimestamp)
			} else {
				__antithesis_instrumentation__.Notify(625634)
			}
			__antithesis_instrumentation__.Notify(625628)

			return nil
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(625635)

		panic(
			errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error returned when iterating through application stats"),
		)
	} else {
		__antithesis_instrumentation__.Notify(625636)
	}
	__antithesis_instrumentation__.Notify(625624)

	return discardedStats
}

func (s *Container) MergeApplicationTransactionStats(
	ctx context.Context, other sqlstats.ApplicationStats,
) (discardedStats uint64) {
	__antithesis_instrumentation__.Notify(625637)
	if err := other.IterateTransactionStats(
		ctx,
		&sqlstats.IteratorOptions{},
		func(ctx context.Context, statistics *roachpb.CollectedTransactionStatistics) error {
			__antithesis_instrumentation__.Notify(625639)
			txnStats, _, throttled :=
				s.getStatsForTxnWithKey(
					statistics.TransactionFingerprintID,
					statistics.StatementFingerprintIDs,
					true,
				)

			if throttled {
				__antithesis_instrumentation__.Notify(625641)
				discardedStats++
				return nil
			} else {
				__antithesis_instrumentation__.Notify(625642)
			}
			__antithesis_instrumentation__.Notify(625640)

			txnStats.mergeStats(&statistics.Stats)
			return nil
		}); err != nil {
		__antithesis_instrumentation__.Notify(625643)

		panic(
			errors.NewAssertionErrorWithWrappedErrf(err, "unexpected error returned when iterating through application stats"),
		)
	} else {
		__antithesis_instrumentation__.Notify(625644)
	}
	__antithesis_instrumentation__.Notify(625638)

	return discardedStats
}

func (s *Container) Add(ctx context.Context, other *Container) (err error) {
	__antithesis_instrumentation__.Notify(625645)
	other.mu.Lock()
	statMap := make(map[stmtKey]*stmtStats)
	for k, v := range other.mu.stmts {
		__antithesis_instrumentation__.Notify(625652)
		statMap[k] = v
	}
	__antithesis_instrumentation__.Notify(625646)
	other.mu.Unlock()

	for k, v := range statMap {
		__antithesis_instrumentation__.Notify(625653)
		v.mu.Lock()
		statCopy := &stmtStats{}
		statCopy.mu.data = v.mu.data
		v.mu.Unlock()
		statCopy.ID = v.ID
		statMap[k] = statCopy
	}
	__antithesis_instrumentation__.Notify(625647)

	for k, v := range statMap {
		__antithesis_instrumentation__.Notify(625654)
		stats, created, throttled := s.getStatsForStmtWithKey(k, v.ID, true)

		if throttled {
			__antithesis_instrumentation__.Notify(625657)
			continue
		} else {
			__antithesis_instrumentation__.Notify(625658)
		}
		__antithesis_instrumentation__.Notify(625655)

		stats.mu.Lock()

		if created {
			__antithesis_instrumentation__.Notify(625659)
			estimatedAllocBytes := stats.sizeUnsafe() + k.size() + 8

			s.mu.Lock()
			if latestErr := s.mu.acc.Grow(ctx, estimatedAllocBytes); latestErr != nil {
				__antithesis_instrumentation__.Notify(625661)
				stats.mu.Unlock()

				err = latestErr
				delete(s.mu.stmts, k)
				s.mu.Unlock()
				continue
			} else {
				__antithesis_instrumentation__.Notify(625662)
			}
			__antithesis_instrumentation__.Notify(625660)
			s.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(625663)
		}
		__antithesis_instrumentation__.Notify(625656)

		stats.mu.data.Add(&v.mu.data)
		stats.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(625648)

	other.mu.Lock()
	txnMap := make(map[roachpb.TransactionFingerprintID]*txnStats)
	for k, v := range other.mu.txns {
		__antithesis_instrumentation__.Notify(625664)
		txnMap[k] = v
	}
	__antithesis_instrumentation__.Notify(625649)
	other.mu.Unlock()

	for k, v := range txnMap {
		__antithesis_instrumentation__.Notify(625665)
		v.mu.Lock()
		txnCopy := &txnStats{}
		txnCopy.mu.data = v.mu.data
		v.mu.Unlock()
		txnCopy.statementFingerprintIDs = v.statementFingerprintIDs
		txnMap[k] = txnCopy
	}
	__antithesis_instrumentation__.Notify(625650)

	for k, v := range txnMap {
		__antithesis_instrumentation__.Notify(625666)

		t, created, throttled := s.getStatsForTxnWithKey(k, v.statementFingerprintIDs, true)

		if throttled {
			__antithesis_instrumentation__.Notify(625669)
			continue
		} else {
			__antithesis_instrumentation__.Notify(625670)
		}
		__antithesis_instrumentation__.Notify(625667)

		t.mu.Lock()
		if created {
			__antithesis_instrumentation__.Notify(625671)
			estimatedAllocBytes := t.sizeUnsafe() + k.Size() + 8

			s.mu.Lock()
			if latestErr := s.mu.acc.Grow(ctx, estimatedAllocBytes); latestErr != nil {
				__antithesis_instrumentation__.Notify(625673)
				t.mu.Unlock()

				err = latestErr
				delete(s.mu.txns, k)
				s.mu.Unlock()
				continue
			} else {
				__antithesis_instrumentation__.Notify(625674)
			}
			__antithesis_instrumentation__.Notify(625672)
			s.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(625675)
		}
		__antithesis_instrumentation__.Notify(625668)

		t.mu.data.Add(&v.mu.data)
		t.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(625651)

	other.txnCounts.mu.Lock()
	txnStats := other.txnCounts.mu.TxnStats
	other.txnCounts.mu.Unlock()

	s.txnCounts.mu.Lock()
	s.txnCounts.mu.TxnStats.Add(txnStats)
	s.txnCounts.mu.Unlock()

	return err
}

func (s *Container) getTimeNow() time.Time {
	__antithesis_instrumentation__.Notify(625676)
	if s.knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(625678)
		return s.knobs.StubTimeNow != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(625679)
		return s.knobs.StubTimeNow()
	} else {
		__antithesis_instrumentation__.Notify(625680)
	}
	__antithesis_instrumentation__.Notify(625677)

	return timeutil.Now()
}

func (s *transactionCounts) recordTransactionCounts(
	txnTimeSec float64, commit bool, implicit bool,
) {
	__antithesis_instrumentation__.Notify(625681)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.TxnCount++
	s.mu.TxnTimeSec.Record(s.mu.TxnCount, txnTimeSec)
	if commit {
		__antithesis_instrumentation__.Notify(625683)
		s.mu.CommittedCount++
	} else {
		__antithesis_instrumentation__.Notify(625684)
	}
	__antithesis_instrumentation__.Notify(625682)
	if implicit {
		__antithesis_instrumentation__.Notify(625685)
		s.mu.ImplicitCount++
	} else {
		__antithesis_instrumentation__.Notify(625686)
	}
}

func (s *Container) getLogicalPlanLastSampled(key sampledPlanKey) time.Time {
	__antithesis_instrumentation__.Notify(625687)
	s.mu.Lock()
	defer s.mu.Unlock()

	lastSampled, found := s.mu.sampledPlanMetadataCache[key]
	if !found {
		__antithesis_instrumentation__.Notify(625689)
		return time.Time{}
	} else {
		__antithesis_instrumentation__.Notify(625690)
	}
	__antithesis_instrumentation__.Notify(625688)

	return lastSampled
}

func (s *Container) setLogicalPlanLastSampled(key sampledPlanKey, time time.Time) {
	__antithesis_instrumentation__.Notify(625691)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.sampledPlanMetadataCache[key] = time
}

func (s *Container) shouldSaveLogicalPlanDescription(lastSampled time.Time) bool {
	__antithesis_instrumentation__.Notify(625692)
	if !sqlstats.SampleLogicalPlans.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(625694)
		return false
	} else {
		__antithesis_instrumentation__.Notify(625695)
	}
	__antithesis_instrumentation__.Notify(625693)
	now := s.getTimeNow()
	period := sqlstats.LogicalPlanCollectionPeriod.Get(&s.st.SV)
	return now.Sub(lastSampled) >= period
}

type transactionCounts struct {
	mu struct {
		syncutil.Mutex

		roachpb.TxnStats
	}
}

func constructStatementFingerprintIDFromStmtKey(key stmtKey) roachpb.StmtFingerprintID {
	__antithesis_instrumentation__.Notify(625696)
	return roachpb.ConstructStatementFingerprintID(
		key.anonymizedStmt, key.failed, key.implicitTxn, key.database,
	)
}
