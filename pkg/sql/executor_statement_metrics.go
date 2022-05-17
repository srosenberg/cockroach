package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

type EngineMetrics struct {
	DistSQLSelectCount *metric.Counter

	SQLOptFallbackCount   *metric.Counter
	SQLOptPlanCacheHits   *metric.Counter
	SQLOptPlanCacheMisses *metric.Counter

	DistSQLExecLatency    *metric.Histogram
	SQLExecLatency        *metric.Histogram
	DistSQLServiceLatency *metric.Histogram
	SQLServiceLatency     *metric.Histogram
	SQLTxnLatency         *metric.Histogram
	SQLTxnsOpen           *metric.Gauge
	SQLActiveStatements   *metric.Gauge

	TxnAbortCount *metric.Counter

	FailureCount *metric.Counter

	FullTableOrIndexScanCount *metric.Counter

	FullTableOrIndexScanRejectedCount *metric.Counter
}

var _ metric.Struct = EngineMetrics{}

func (EngineMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(490967) }

type StatsMetrics struct {
	SQLStatsMemoryMaxBytesHist  *metric.Histogram
	SQLStatsMemoryCurBytesCount *metric.Gauge

	ReportedSQLStatsMemoryMaxBytesHist  *metric.Histogram
	ReportedSQLStatsMemoryCurBytesCount *metric.Gauge

	DiscardedStatsCount *metric.Counter

	SQLStatsFlushStarted  *metric.Counter
	SQLStatsFlushFailure  *metric.Counter
	SQLStatsFlushDuration *metric.Histogram
	SQLStatsRemovedRows   *metric.Counter

	SQLTxnStatsCollectionOverhead *metric.Histogram
}

var _ metric.Struct = StatsMetrics{}

func (StatsMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(490968) }

type GuardrailMetrics struct {
	TxnRowsWrittenLogCount *metric.Counter
	TxnRowsWrittenErrCount *metric.Counter
	TxnRowsReadLogCount    *metric.Counter
	TxnRowsReadErrCount    *metric.Counter
}

var _ metric.Struct = GuardrailMetrics{}

func (GuardrailMetrics) MetricStruct() { __antithesis_instrumentation__.Notify(490969) }

func (ex *connExecutor) recordStatementSummary(
	ctx context.Context,
	planner *planner,
	automaticRetryCount int,
	rowsAffected int,
	stmtErr error,
	stats topLevelQueryStats,
) {
	__antithesis_instrumentation__.Notify(490970)
	phaseTimes := ex.statsCollector.PhaseTimes()

	runLatRaw := phaseTimes.GetRunLatency()
	runLat := runLatRaw.Seconds()
	parseLat := phaseTimes.GetParsingLatency().Seconds()
	planLat := phaseTimes.GetPlanningLatency().Seconds()

	svcLatRaw := phaseTimes.GetServiceLatencyNoOverhead()
	svcLat := svcLatRaw.Seconds()

	processingLat := parseLat + planLat + runLat

	execOverhead := svcLat - processingLat

	stmt := &planner.stmt
	shouldIncludeInLatencyMetrics := shouldIncludeStmtInLatencyMetrics(stmt)
	flags := planner.curPlan.flags
	if automaticRetryCount == 0 {
		__antithesis_instrumentation__.Notify(490976)
		ex.updateOptCounters(flags)
		m := &ex.metrics.EngineMetrics
		if flags.IsDistributed() {
			__antithesis_instrumentation__.Notify(490978)
			if _, ok := stmt.AST.(*tree.Select); ok {
				__antithesis_instrumentation__.Notify(490980)
				m.DistSQLSelectCount.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(490981)
			}
			__antithesis_instrumentation__.Notify(490979)
			if shouldIncludeInLatencyMetrics {
				__antithesis_instrumentation__.Notify(490982)
				m.DistSQLExecLatency.RecordValue(runLatRaw.Nanoseconds())
				m.DistSQLServiceLatency.RecordValue(svcLatRaw.Nanoseconds())
			} else {
				__antithesis_instrumentation__.Notify(490983)
			}
		} else {
			__antithesis_instrumentation__.Notify(490984)
		}
		__antithesis_instrumentation__.Notify(490977)
		if shouldIncludeInLatencyMetrics {
			__antithesis_instrumentation__.Notify(490985)
			m.SQLExecLatency.RecordValue(runLatRaw.Nanoseconds())
			m.SQLServiceLatency.RecordValue(svcLatRaw.Nanoseconds())
		} else {
			__antithesis_instrumentation__.Notify(490986)
		}
	} else {
		__antithesis_instrumentation__.Notify(490987)
	}
	__antithesis_instrumentation__.Notify(490971)

	recordedStmtStatsKey := roachpb.StatementStatisticsKey{
		Query:        stmt.StmtNoConstants,
		QuerySummary: stmt.StmtSummary,
		DistSQL:      flags.IsDistributed(),
		Vec:          flags.IsSet(planFlagVectorized),
		ImplicitTxn:  flags.IsSet(planFlagImplicitTxn),
		FullScan: flags.IsSet(planFlagContainsFullIndexScan) || func() bool {
			__antithesis_instrumentation__.Notify(490988)
			return flags.IsSet(planFlagContainsFullTableScan) == true
		}() == true,
		Failed:   stmtErr != nil,
		Database: planner.SessionData().Database,
		PlanHash: planner.instrumentation.planGist.Hash(),
	}

	if ex.implicitTxn() {
		__antithesis_instrumentation__.Notify(490989)
		stmtFingerprintID := recordedStmtStatsKey.FingerprintID()
		txnFingerprintHash := util.MakeFNV64()
		txnFingerprintHash.Add(uint64(stmtFingerprintID))
		recordedStmtStatsKey.TransactionFingerprintID =
			roachpb.TransactionFingerprintID(txnFingerprintHash.Sum())
	} else {
		__antithesis_instrumentation__.Notify(490990)
	}
	__antithesis_instrumentation__.Notify(490972)

	recordedStmtStats := sqlstats.RecordedStmtStats{
		AutoRetryCount:  automaticRetryCount,
		RowsAffected:    rowsAffected,
		ParseLatency:    parseLat,
		PlanLatency:     planLat,
		RunLatency:      runLat,
		ServiceLatency:  svcLat,
		OverheadLatency: execOverhead,
		BytesRead:       stats.bytesRead,
		RowsRead:        stats.rowsRead,
		RowsWritten:     stats.rowsWritten,
		Nodes:           getNodesFromPlanner(planner),
		StatementType:   stmt.AST.StatementType(),
		Plan:            planner.instrumentation.PlanForStats(ctx),
		PlanGist:        planner.instrumentation.planGist.String(),
		StatementError:  stmtErr,
	}

	stmtFingerprintID, err :=
		ex.statsCollector.RecordStatement(ctx, recordedStmtStatsKey, recordedStmtStats)

	if err != nil {
		__antithesis_instrumentation__.Notify(490991)
		if log.V(1) {
			__antithesis_instrumentation__.Notify(490993)
			log.Warningf(ctx, "failed to record statement: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(490994)
		}
		__antithesis_instrumentation__.Notify(490992)
		ex.server.ServerMetrics.StatsMetrics.DiscardedStatsCount.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(490995)
	}
	__antithesis_instrumentation__.Notify(490973)

	maxStmtFingerprintIDsLen := sqlstats.TxnStatsNumStmtFingerprintIDsToRecord.Get(&ex.server.cfg.Settings.SV)
	if int64(len(ex.extraTxnState.transactionStatementFingerprintIDs)) < maxStmtFingerprintIDsLen {
		__antithesis_instrumentation__.Notify(490996)
		ex.extraTxnState.transactionStatementFingerprintIDs = append(
			ex.extraTxnState.transactionStatementFingerprintIDs, stmtFingerprintID)
	} else {
		__antithesis_instrumentation__.Notify(490997)
	}
	__antithesis_instrumentation__.Notify(490974)

	if ex.extraTxnState.transactionStatementsHash.IsInitialized() {
		__antithesis_instrumentation__.Notify(490998)
		ex.extraTxnState.transactionStatementsHash.Add(uint64(stmtFingerprintID))
	} else {
		__antithesis_instrumentation__.Notify(490999)
	}
	__antithesis_instrumentation__.Notify(490975)
	ex.extraTxnState.numRows += rowsAffected

	if log.V(2) {
		__antithesis_instrumentation__.Notify(491000)

		sessionAge := phaseTimes.GetSessionAge().Seconds()

		log.Infof(ctx,
			"query stats: %d rows, %d retries, "+
				"parse %.2fµs (%.1f%%), "+
				"plan %.2fµs (%.1f%%), "+
				"run %.2fµs (%.1f%%), "+
				"overhead %.2fµs (%.1f%%), "+
				"session age %.4fs",
			rowsAffected, automaticRetryCount,
			parseLat*1e6, 100*parseLat/svcLat,
			planLat*1e6, 100*planLat/svcLat,
			runLat*1e6, 100*runLat/svcLat,
			execOverhead*1e6, 100*execOverhead/svcLat,
			sessionAge,
		)
	} else {
		__antithesis_instrumentation__.Notify(491001)
	}
}

func (ex *connExecutor) updateOptCounters(planFlags planFlags) {
	__antithesis_instrumentation__.Notify(491002)
	m := &ex.metrics.EngineMetrics

	if planFlags.IsSet(planFlagOptCacheHit) {
		__antithesis_instrumentation__.Notify(491003)
		m.SQLOptPlanCacheHits.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(491004)
		if planFlags.IsSet(planFlagOptCacheMiss) {
			__antithesis_instrumentation__.Notify(491005)
			m.SQLOptPlanCacheMisses.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(491006)
		}
	}
}

func shouldIncludeStmtInLatencyMetrics(stmt *Statement) bool {
	__antithesis_instrumentation__.Notify(491007)
	return stmt.AST.StatementType() == tree.TypeDML
}

func getNodesFromPlanner(planner *planner) []int64 {
	__antithesis_instrumentation__.Notify(491008)

	var nodes []int64
	if planner.instrumentation.sp != nil {
		__antithesis_instrumentation__.Notify(491010)
		trace := planner.instrumentation.sp.GetRecording(tracing.RecordingStructured)

		execinfrapb.ExtractNodesFromSpans(planner.EvalContext().Context, trace).ForEach(func(i int) {
			__antithesis_instrumentation__.Notify(491011)
			nodes = append(nodes, int64(i))
		})
	} else {
		__antithesis_instrumentation__.Notify(491012)
	}
	__antithesis_instrumentation__.Notify(491009)

	return nodes
}
