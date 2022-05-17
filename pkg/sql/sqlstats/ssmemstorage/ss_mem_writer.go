package ssmemstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
)

var (
	ErrMemoryPressure = errors.New("insufficient sql stats memory")

	ErrFingerprintLimitReached = errors.New("sql stats fingerprint limit reached")

	ErrExecStatsFingerprintFlushed = errors.New("stmtStats flushed before execution stats can be recorded")
)

var timestampSize = int64(unsafe.Sizeof(time.Time{}))

var _ sqlstats.Writer = &Container{}

func (s *Container) RecordStatement(
	ctx context.Context, key roachpb.StatementStatisticsKey, value sqlstats.RecordedStmtStats,
) (roachpb.StmtFingerprintID, error) {
	__antithesis_instrumentation__.Notify(625697)
	createIfNonExistent := true

	t := sqlstats.StatsCollectionLatencyThreshold.Get(&s.st.SV)
	if !sqlstats.StmtStatsEnable.Get(&s.st.SV) || func() bool {
		__antithesis_instrumentation__.Notify(625705)
		return (t > 0 && func() bool {
			__antithesis_instrumentation__.Notify(625706)
			return t.Seconds() >= value.ServiceLatency == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(625707)
		createIfNonExistent = false
	} else {
		__antithesis_instrumentation__.Notify(625708)
	}
	__antithesis_instrumentation__.Notify(625698)

	stats, statementKey, stmtFingerprintID, created, throttled := s.getStatsForStmt(
		key.Query,
		key.ImplicitTxn,
		key.Database,
		key.Failed,
		key.PlanHash,
		key.TransactionFingerprintID,
		createIfNonExistent,
	)

	if throttled {
		__antithesis_instrumentation__.Notify(625709)
		return stmtFingerprintID, ErrFingerprintLimitReached
	} else {
		__antithesis_instrumentation__.Notify(625710)
	}
	__antithesis_instrumentation__.Notify(625699)

	if !createIfNonExistent {
		__antithesis_instrumentation__.Notify(625711)
		return stmtFingerprintID, nil
	} else {
		__antithesis_instrumentation__.Notify(625712)
	}
	__antithesis_instrumentation__.Notify(625700)

	stats.mu.Lock()
	defer stats.mu.Unlock()

	stats.mu.data.Count++
	if key.Failed {
		__antithesis_instrumentation__.Notify(625713)
		stats.mu.data.SensitiveInfo.LastErr = value.StatementError.Error()
	} else {
		__antithesis_instrumentation__.Notify(625714)
	}
	__antithesis_instrumentation__.Notify(625701)

	if value.Plan != nil {
		__antithesis_instrumentation__.Notify(625715)
		stats.mu.data.SensitiveInfo.MostRecentPlanDescription = *value.Plan
		stats.mu.data.SensitiveInfo.MostRecentPlanTimestamp = s.getTimeNow()
		s.setLogicalPlanLastSampled(statementKey.sampledPlanKey, stats.mu.data.SensitiveInfo.MostRecentPlanTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(625716)
	}
	__antithesis_instrumentation__.Notify(625702)
	if value.AutoRetryCount == 0 {
		__antithesis_instrumentation__.Notify(625717)
		stats.mu.data.FirstAttemptCount++
	} else {
		__antithesis_instrumentation__.Notify(625718)
		if int64(value.AutoRetryCount) > stats.mu.data.MaxRetries {
			__antithesis_instrumentation__.Notify(625719)
			stats.mu.data.MaxRetries = int64(value.AutoRetryCount)
		} else {
			__antithesis_instrumentation__.Notify(625720)
		}
	}
	__antithesis_instrumentation__.Notify(625703)
	stats.mu.data.SQLType = value.StatementType.String()
	stats.mu.data.NumRows.Record(stats.mu.data.Count, float64(value.RowsAffected))
	stats.mu.data.ParseLat.Record(stats.mu.data.Count, value.ParseLatency)
	stats.mu.data.PlanLat.Record(stats.mu.data.Count, value.PlanLatency)
	stats.mu.data.RunLat.Record(stats.mu.data.Count, value.RunLatency)
	stats.mu.data.ServiceLat.Record(stats.mu.data.Count, value.ServiceLatency)
	stats.mu.data.OverheadLat.Record(stats.mu.data.Count, value.OverheadLatency)
	stats.mu.data.BytesRead.Record(stats.mu.data.Count, float64(value.BytesRead))
	stats.mu.data.RowsRead.Record(stats.mu.data.Count, float64(value.RowsRead))
	stats.mu.data.RowsWritten.Record(stats.mu.data.Count, float64(value.RowsWritten))
	stats.mu.data.LastExecTimestamp = s.getTimeNow()
	stats.mu.data.Nodes = util.CombineUniqueInt64(stats.mu.data.Nodes, value.Nodes)
	stats.mu.data.PlanGists = util.CombineUniqueString(stats.mu.data.PlanGists, []string{value.PlanGist})

	stats.mu.vectorized = key.Vec
	stats.mu.distSQLUsed = key.DistSQL
	stats.mu.fullScan = key.FullScan
	stats.mu.database = key.Database
	stats.mu.querySummary = key.QuerySummary

	if created {
		__antithesis_instrumentation__.Notify(625721)

		estimatedMemoryAllocBytes := stats.sizeUnsafe() + statementKey.size() + 8

		estimatedMemoryAllocBytes += timestampSize + statementKey.sampledPlanKey.size() + 8
		s.mu.Lock()
		defer s.mu.Unlock()

		if s.mu.acc.Monitor() == nil {
			__antithesis_instrumentation__.Notify(625723)
			return stats.ID, nil
		} else {
			__antithesis_instrumentation__.Notify(625724)
		}
		__antithesis_instrumentation__.Notify(625722)

		if err := s.mu.acc.Grow(ctx, estimatedMemoryAllocBytes); err != nil {
			__antithesis_instrumentation__.Notify(625725)
			delete(s.mu.stmts, statementKey)
			return stats.ID, ErrMemoryPressure
		} else {
			__antithesis_instrumentation__.Notify(625726)
		}
	} else {
		__antithesis_instrumentation__.Notify(625727)
	}
	__antithesis_instrumentation__.Notify(625704)

	return stats.ID, nil
}

func (s *Container) RecordStatementExecStats(
	key roachpb.StatementStatisticsKey, stats execstats.QueryLevelStats,
) error {
	__antithesis_instrumentation__.Notify(625728)
	stmtStats, _, _, _, _ :=
		s.getStatsForStmt(
			key.Query,
			key.ImplicitTxn,
			key.Database,
			key.Failed,
			key.PlanHash,
			key.TransactionFingerprintID,
			false,
		)
	if stmtStats == nil {
		__antithesis_instrumentation__.Notify(625730)
		return ErrExecStatsFingerprintFlushed
	} else {
		__antithesis_instrumentation__.Notify(625731)
	}
	__antithesis_instrumentation__.Notify(625729)
	stmtStats.recordExecStats(stats)
	return nil
}

func (s *Container) ShouldSaveLogicalPlanDesc(
	fingerprint string, implicitTxn bool, database string,
) bool {
	__antithesis_instrumentation__.Notify(625732)
	lastSampled := s.getLogicalPlanLastSampled(sampledPlanKey{
		anonymizedStmt: fingerprint,
		implicitTxn:    implicitTxn,
		database:       database,
	})
	return s.shouldSaveLogicalPlanDescription(lastSampled)
}

func (s *Container) RecordTransaction(
	ctx context.Context, key roachpb.TransactionFingerprintID, value sqlstats.RecordedTxnStats,
) error {
	__antithesis_instrumentation__.Notify(625733)
	s.recordTransactionHighLevelStats(value.TransactionTimeSec, value.Committed, value.ImplicitTxn)

	if !sqlstats.TxnStatsEnable.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(625740)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(625741)
	}
	__antithesis_instrumentation__.Notify(625734)

	t := sqlstats.StatsCollectionLatencyThreshold.Get(&s.st.SV)
	if t > 0 {
		__antithesis_instrumentation__.Notify(625742)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(625743)
	}
	__antithesis_instrumentation__.Notify(625735)

	stats, created, throttled := s.getStatsForTxnWithKey(key, value.StatementFingerprintIDs, true)

	if throttled {
		__antithesis_instrumentation__.Notify(625744)
		return ErrFingerprintLimitReached
	} else {
		__antithesis_instrumentation__.Notify(625745)
	}
	__antithesis_instrumentation__.Notify(625736)

	stats.mu.Lock()
	defer stats.mu.Unlock()

	if created {
		__antithesis_instrumentation__.Notify(625746)
		estimatedMemAllocBytes :=
			stats.sizeUnsafe() + key.Size() + 8
		s.mu.Lock()

		if s.mu.acc.Monitor() != nil {
			__antithesis_instrumentation__.Notify(625748)
			if err := s.mu.acc.Grow(ctx, estimatedMemAllocBytes); err != nil {
				__antithesis_instrumentation__.Notify(625749)
				delete(s.mu.txns, key)
				s.mu.Unlock()
				return ErrMemoryPressure
			} else {
				__antithesis_instrumentation__.Notify(625750)
			}
		} else {
			__antithesis_instrumentation__.Notify(625751)
		}
		__antithesis_instrumentation__.Notify(625747)
		s.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(625752)
	}
	__antithesis_instrumentation__.Notify(625737)

	stats.mu.data.Count++

	stats.mu.data.NumRows.Record(stats.mu.data.Count, float64(value.RowsAffected))
	stats.mu.data.ServiceLat.Record(stats.mu.data.Count, value.ServiceLatency.Seconds())
	stats.mu.data.RetryLat.Record(stats.mu.data.Count, value.RetryLatency.Seconds())
	stats.mu.data.CommitLat.Record(stats.mu.data.Count, value.CommitLatency.Seconds())
	if value.RetryCount > stats.mu.data.MaxRetries {
		__antithesis_instrumentation__.Notify(625753)
		stats.mu.data.MaxRetries = value.RetryCount
	} else {
		__antithesis_instrumentation__.Notify(625754)
	}
	__antithesis_instrumentation__.Notify(625738)
	stats.mu.data.RowsRead.Record(stats.mu.data.Count, float64(value.RowsRead))
	stats.mu.data.RowsWritten.Record(stats.mu.data.Count, float64(value.RowsWritten))
	stats.mu.data.BytesRead.Record(stats.mu.data.Count, float64(value.BytesRead))

	if value.CollectedExecStats {
		__antithesis_instrumentation__.Notify(625755)
		stats.mu.data.ExecStats.Count++
		stats.mu.data.ExecStats.NetworkBytes.Record(stats.mu.data.ExecStats.Count, float64(value.ExecStats.NetworkBytesSent))
		stats.mu.data.ExecStats.MaxMemUsage.Record(stats.mu.data.ExecStats.Count, float64(value.ExecStats.MaxMemUsage))
		stats.mu.data.ExecStats.ContentionTime.Record(stats.mu.data.ExecStats.Count, value.ExecStats.ContentionTime.Seconds())
		stats.mu.data.ExecStats.NetworkMessages.Record(stats.mu.data.ExecStats.Count, float64(value.ExecStats.NetworkMessages))
		stats.mu.data.ExecStats.MaxDiskUsage.Record(stats.mu.data.ExecStats.Count, float64(value.ExecStats.MaxDiskUsage))
	} else {
		__antithesis_instrumentation__.Notify(625756)
	}
	__antithesis_instrumentation__.Notify(625739)

	return nil
}

func (s *Container) recordTransactionHighLevelStats(
	transactionTimeSec float64, committed bool, implicit bool,
) {
	__antithesis_instrumentation__.Notify(625757)
	if !sqlstats.TxnStatsEnable.Get(&s.st.SV) {
		__antithesis_instrumentation__.Notify(625759)
		return
	} else {
		__antithesis_instrumentation__.Notify(625760)
	}
	__antithesis_instrumentation__.Notify(625758)
	s.txnCounts.recordTransactionCounts(transactionTimeSec, committed, implicit)
}
