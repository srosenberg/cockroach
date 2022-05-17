package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats/sqlstatsutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func (s *PersistedSQLStats) Flush(ctx context.Context) {
	__antithesis_instrumentation__.Notify(624690)
	now := s.getTimeNow()

	allowDiscardWhenDisabled := DiscardInMemoryStatsWhenFlushDisabled.Get(&s.cfg.Settings.SV)
	minimumFlushInterval := MinimumInterval.Get(&s.cfg.Settings.SV)

	enabled := SQLStatsFlushEnabled.Get(&s.cfg.Settings.SV)
	flushingTooSoon := now.Before(s.lastFlushStarted.Add(minimumFlushInterval))

	shouldWipeInMemoryStats := enabled && func() bool {
		__antithesis_instrumentation__.Notify(624694)
		return !flushingTooSoon == true
	}() == true
	shouldWipeInMemoryStats = shouldWipeInMemoryStats || func() bool {
		__antithesis_instrumentation__.Notify(624695)
		return (!enabled && func() bool {
			__antithesis_instrumentation__.Notify(624696)
			return allowDiscardWhenDisabled == true
		}() == true) == true
	}() == true

	if shouldWipeInMemoryStats {
		__antithesis_instrumentation__.Notify(624697)
		defer func() {
			__antithesis_instrumentation__.Notify(624698)
			if err := s.SQLStats.Reset(ctx); err != nil {
				__antithesis_instrumentation__.Notify(624699)
				log.Warningf(ctx, "fail to reset in-memory SQL Stats: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(624700)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(624701)
	}
	__antithesis_instrumentation__.Notify(624691)

	if !enabled {
		__antithesis_instrumentation__.Notify(624702)
		return
	} else {
		__antithesis_instrumentation__.Notify(624703)
	}
	__antithesis_instrumentation__.Notify(624692)

	if flushingTooSoon {
		__antithesis_instrumentation__.Notify(624704)
		log.Infof(ctx, "flush aborted due to high flush frequency. "+
			"The minimum interval between flushes is %s", minimumFlushInterval.String())
		return
	} else {
		__antithesis_instrumentation__.Notify(624705)
	}
	__antithesis_instrumentation__.Notify(624693)

	s.lastFlushStarted = now
	log.Infof(ctx, "flushing %d stmt/txn fingerprints (%d bytes) after %s",
		s.SQLStats.GetTotalFingerprintCount(), s.SQLStats.GetTotalFingerprintBytes(), timeutil.Since(s.lastFlushStarted))

	aggregatedTs := s.ComputeAggregatedTs()

	s.flushStmtStats(ctx, aggregatedTs)
	s.flushTxnStats(ctx, aggregatedTs)
}

func (s *PersistedSQLStats) flushStmtStats(ctx context.Context, aggregatedTs time.Time) {
	__antithesis_instrumentation__.Notify(624706)

	_ = s.SQLStats.IterateStatementStats(ctx, &sqlstats.IteratorOptions{},
		func(ctx context.Context, statistics *roachpb.CollectedStatementStatistics) error {
			__antithesis_instrumentation__.Notify(624708)
			s.doFlush(ctx, func() error {
				__antithesis_instrumentation__.Notify(624710)
				return s.doFlushSingleStmtStats(ctx, statistics, aggregatedTs)
			}, "failed to flush statement statistics")
			__antithesis_instrumentation__.Notify(624709)

			return nil
		})
	__antithesis_instrumentation__.Notify(624707)

	if s.cfg.Knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(624711)
		return s.cfg.Knobs.OnStmtStatsFlushFinished != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(624712)
		s.cfg.Knobs.OnStmtStatsFlushFinished()
	} else {
		__antithesis_instrumentation__.Notify(624713)
	}
}

func (s *PersistedSQLStats) flushTxnStats(ctx context.Context, aggregatedTs time.Time) {
	__antithesis_instrumentation__.Notify(624714)
	_ = s.SQLStats.IterateTransactionStats(ctx, &sqlstats.IteratorOptions{},
		func(ctx context.Context, statistics *roachpb.CollectedTransactionStatistics) error {
			__antithesis_instrumentation__.Notify(624716)
			s.doFlush(ctx, func() error {
				__antithesis_instrumentation__.Notify(624718)
				return s.doFlushSingleTxnStats(ctx, statistics, aggregatedTs)
			}, "failed to flush transaction statistics")
			__antithesis_instrumentation__.Notify(624717)

			return nil
		})
	__antithesis_instrumentation__.Notify(624715)

	if s.cfg.Knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(624719)
		return s.cfg.Knobs.OnTxnStatsFlushFinished != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(624720)
		s.cfg.Knobs.OnTxnStatsFlushFinished()
	} else {
		__antithesis_instrumentation__.Notify(624721)
	}
}

func (s *PersistedSQLStats) doFlush(ctx context.Context, workFn func() error, errMsg string) {
	__antithesis_instrumentation__.Notify(624722)
	var err error
	flushBegin := s.getTimeNow()

	defer func() {
		__antithesis_instrumentation__.Notify(624724)
		if err != nil {
			__antithesis_instrumentation__.Notify(624726)
			s.cfg.FailureCounter.Inc(1)
			log.Warningf(ctx, "%s: %s", errMsg, err)
		} else {
			__antithesis_instrumentation__.Notify(624727)
		}
		__antithesis_instrumentation__.Notify(624725)
		flushDuration := s.getTimeNow().Sub(flushBegin)
		s.cfg.FlushDuration.RecordValue(flushDuration.Nanoseconds())
		s.cfg.FlushCounter.Inc(1)
	}()
	__antithesis_instrumentation__.Notify(624723)

	err = workFn()
}

func (s *PersistedSQLStats) doFlushSingleTxnStats(
	ctx context.Context, stats *roachpb.CollectedTransactionStatistics, aggregatedTs time.Time,
) error {
	__antithesis_instrumentation__.Notify(624728)
	return s.cfg.KvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624729)

		scopedStats := *stats

		serializedFingerprintID := sqlstatsutil.EncodeUint64ToBytes(uint64(stats.TransactionFingerprintID))

		insertFn := func(ctx context.Context, txn *kv.Txn) (alreadyExists bool, err error) {
			__antithesis_instrumentation__.Notify(624734)
			rowsAffected, err := s.insertTransactionStats(ctx, txn, aggregatedTs, serializedFingerprintID, &scopedStats)

			if err != nil {
				__antithesis_instrumentation__.Notify(624737)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(624738)
			}
			__antithesis_instrumentation__.Notify(624735)

			if rowsAffected == 0 {
				__antithesis_instrumentation__.Notify(624739)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(624740)
			}
			__antithesis_instrumentation__.Notify(624736)

			return false, nil
		}
		__antithesis_instrumentation__.Notify(624730)

		readFn := func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(624741)
			persistedData := roachpb.TransactionStatistics{}
			err := s.fetchPersistedTransactionStats(ctx, txn, aggregatedTs, serializedFingerprintID, scopedStats.App, &persistedData)
			if err != nil {
				__antithesis_instrumentation__.Notify(624743)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624744)
			}
			__antithesis_instrumentation__.Notify(624742)

			scopedStats.Stats.Add(&persistedData)
			return nil
		}
		__antithesis_instrumentation__.Notify(624731)

		updateFn := func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(624745)
			return s.updateTransactionStats(ctx, txn, aggregatedTs, serializedFingerprintID, &scopedStats)
		}
		__antithesis_instrumentation__.Notify(624732)

		err := s.doInsertElseDoUpdate(ctx, txn, insertFn, readFn, updateFn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624746)
			return errors.Wrapf(err, "flushing transaction %d's statistics", stats.TransactionFingerprintID)
		} else {
			__antithesis_instrumentation__.Notify(624747)
		}
		__antithesis_instrumentation__.Notify(624733)
		return nil
	})
}

func (s *PersistedSQLStats) doFlushSingleStmtStats(
	ctx context.Context, stats *roachpb.CollectedStatementStatistics, aggregatedTs time.Time,
) error {
	__antithesis_instrumentation__.Notify(624748)
	return s.cfg.KvDB.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(624749)

		scopedStats := *stats

		serializedFingerprintID := sqlstatsutil.EncodeUint64ToBytes(uint64(scopedStats.ID))
		serializedTransactionFingerprintID := sqlstatsutil.EncodeUint64ToBytes(uint64(scopedStats.Key.TransactionFingerprintID))
		serializedPlanHash := sqlstatsutil.EncodeUint64ToBytes(scopedStats.Key.PlanHash)

		insertFn := func(ctx context.Context, txn *kv.Txn) (alreadyExists bool, err error) {
			__antithesis_instrumentation__.Notify(624754)
			rowsAffected, err := s.insertStatementStats(
				ctx,
				txn,
				aggregatedTs,
				serializedFingerprintID,
				serializedTransactionFingerprintID,
				serializedPlanHash,
				&scopedStats,
			)

			if err != nil {
				__antithesis_instrumentation__.Notify(624757)
				return false, err
			} else {
				__antithesis_instrumentation__.Notify(624758)
			}
			__antithesis_instrumentation__.Notify(624755)

			if rowsAffected == 0 {
				__antithesis_instrumentation__.Notify(624759)
				return true, nil
			} else {
				__antithesis_instrumentation__.Notify(624760)
			}
			__antithesis_instrumentation__.Notify(624756)

			return false, nil
		}
		__antithesis_instrumentation__.Notify(624750)

		readFn := func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(624761)
			persistedData := roachpb.StatementStatistics{}
			err := s.fetchPersistedStatementStats(
				ctx,
				txn,
				aggregatedTs,
				serializedFingerprintID,
				serializedTransactionFingerprintID,
				serializedPlanHash,
				&scopedStats.Key,
				&persistedData,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(624763)
				return err
			} else {
				__antithesis_instrumentation__.Notify(624764)
			}
			__antithesis_instrumentation__.Notify(624762)

			scopedStats.Stats.Add(&persistedData)
			return nil
		}
		__antithesis_instrumentation__.Notify(624751)

		updateFn := func(ctx context.Context, txn *kv.Txn) error {
			__antithesis_instrumentation__.Notify(624765)
			return s.updateStatementStats(
				ctx,
				txn,
				aggregatedTs,
				serializedFingerprintID,
				serializedTransactionFingerprintID,
				serializedPlanHash,
				&scopedStats,
			)
		}
		__antithesis_instrumentation__.Notify(624752)

		err := s.doInsertElseDoUpdate(ctx, txn, insertFn, readFn, updateFn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624766)
			return errors.Wrapf(err, "flush statement %d's statistics", scopedStats.ID)
		} else {
			__antithesis_instrumentation__.Notify(624767)
		}
		__antithesis_instrumentation__.Notify(624753)
		return nil
	})
}

func (s *PersistedSQLStats) doInsertElseDoUpdate(
	ctx context.Context,
	txn *kv.Txn,
	insertFn func(context.Context, *kv.Txn) (alreadyExists bool, err error),
	readFn func(context.Context, *kv.Txn) error,
	updateFn func(context.Context, *kv.Txn) error,
) error {
	__antithesis_instrumentation__.Notify(624768)
	alreadyExists, err := insertFn(ctx, txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(624771)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624772)
	}
	__antithesis_instrumentation__.Notify(624769)

	if alreadyExists {
		__antithesis_instrumentation__.Notify(624773)
		err = readFn(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624775)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624776)
		}
		__antithesis_instrumentation__.Notify(624774)

		err = updateFn(ctx, txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(624777)
			return err
		} else {
			__antithesis_instrumentation__.Notify(624778)
		}
	} else {
		__antithesis_instrumentation__.Notify(624779)
	}
	__antithesis_instrumentation__.Notify(624770)

	return nil
}

func (s *PersistedSQLStats) ComputeAggregatedTs() time.Time {
	__antithesis_instrumentation__.Notify(624780)
	interval := SQLStatsAggregationInterval.Get(&s.cfg.Settings.SV)
	now := s.getTimeNow()

	aggTs := now.Truncate(interval)

	return aggTs
}

func (s *PersistedSQLStats) GetAggregationInterval() time.Duration {
	__antithesis_instrumentation__.Notify(624781)
	return SQLStatsAggregationInterval.Get(&s.cfg.Settings.SV)
}

func (s *PersistedSQLStats) getTimeNow() time.Time {
	__antithesis_instrumentation__.Notify(624782)
	if s.cfg.Knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(624784)
		return s.cfg.Knobs.StubTimeNow != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(624785)
		return s.cfg.Knobs.StubTimeNow()
	} else {
		__antithesis_instrumentation__.Notify(624786)
	}
	__antithesis_instrumentation__.Notify(624783)

	return timeutil.Now()
}

func (s *PersistedSQLStats) insertTransactionStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	stats *roachpb.CollectedTransactionStatistics,
) (rowsAffected int, err error) {
	__antithesis_instrumentation__.Notify(624787)
	insertStmt := `
INSERT INTO system.transaction_statistics
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_shard_8, aggregated_ts, fingerprint_id, app_name, node_id)
DO NOTHING
`

	aggInterval := s.GetAggregationInterval()

	metadataJSON, err := sqlstatsutil.BuildTxnMetadataJSON(stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624790)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(624791)
	}
	__antithesis_instrumentation__.Notify(624788)
	metadata := tree.NewDJSON(metadataJSON)

	statisticsJSON, err := sqlstatsutil.BuildTxnStatisticsJSON(stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624792)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(624793)
	}
	__antithesis_instrumentation__.Notify(624789)
	statistics := tree.NewDJSON(statisticsJSON)

	rowsAffected, err = s.cfg.InternalExecutor.ExecEx(
		ctx,
		"insert-txn-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		insertStmt,
		aggregatedTs,
		serializedFingerprintID,
		stats.App,
		s.cfg.SQLIDContainer.SQLInstanceID(),
		aggInterval,
		metadata,
		statistics,
	)

	return rowsAffected, err
}
func (s *PersistedSQLStats) updateTransactionStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	stats *roachpb.CollectedTransactionStatistics,
) error {
	__antithesis_instrumentation__.Notify(624794)
	updateStmt := `
UPDATE system.transaction_statistics
SET statistics = $1
WHERE fingerprint_id = $2
	AND aggregated_ts = $3
  AND app_name = $4
  AND node_id = $5
`

	statisticsJSON, err := sqlstatsutil.BuildTxnStatisticsJSON(stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624798)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624799)
	}
	__antithesis_instrumentation__.Notify(624795)
	statistics := tree.NewDJSON(statisticsJSON)

	rowsAffected, err := s.cfg.InternalExecutor.ExecEx(
		ctx,
		"update-stmt-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		updateStmt,
		statistics,
		serializedFingerprintID,
		aggregatedTs,
		stats.App,
		s.cfg.SQLIDContainer.SQLInstanceID(),
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(624800)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624801)
	}
	__antithesis_instrumentation__.Notify(624796)

	if rowsAffected == 0 {
		__antithesis_instrumentation__.Notify(624802)
		return errors.AssertionFailedf("failed to update transaction statistics for  fingerprint_id: %s, app: %s, aggregated_ts: %s, node_id: %d",
			serializedFingerprintID, stats.App, aggregatedTs,
			s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624803)
	}
	__antithesis_instrumentation__.Notify(624797)

	return nil
}

func (s *PersistedSQLStats) updateStatementStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	serializedTransactionFingerprintID []byte,
	serializedPlanHash []byte,
	stats *roachpb.CollectedStatementStatistics,
) error {
	__antithesis_instrumentation__.Notify(624804)
	updateStmt := `
UPDATE system.statement_statistics
SET statistics = $1
WHERE fingerprint_id = $2
  AND transaction_fingerprint_id = $3
	AND aggregated_ts = $4
  AND app_name = $5
  AND plan_hash = $6
  AND node_id = $7
`

	statisticsJSON, err := sqlstatsutil.BuildStmtStatisticsJSON(&stats.Stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624808)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624809)
	}
	__antithesis_instrumentation__.Notify(624805)
	statistics := tree.NewDJSON(statisticsJSON)

	rowsAffected, err := s.cfg.InternalExecutor.ExecEx(
		ctx,
		"update-stmt-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		updateStmt,
		statistics,
		serializedFingerprintID,
		serializedTransactionFingerprintID,
		aggregatedTs,
		stats.Key.App,
		serializedPlanHash,
		s.cfg.SQLIDContainer.SQLInstanceID(),
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(624810)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624811)
	}
	__antithesis_instrumentation__.Notify(624806)

	if rowsAffected == 0 {
		__antithesis_instrumentation__.Notify(624812)
		return errors.AssertionFailedf("failed to update statement statistics "+
			"for fingerprint_id: %s, "+
			"transaction_fingerprint_id: %s, "+
			"app: %s, "+
			"aggregated_ts: %s, "+
			"plan_hash: %d, "+
			"node_id: %d",
			serializedFingerprintID, serializedTransactionFingerprintID, stats.Key.App,
			aggregatedTs, serializedPlanHash, s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624813)
	}
	__antithesis_instrumentation__.Notify(624807)

	return nil
}

func (s *PersistedSQLStats) insertStatementStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	serializedTransactionFingerprintID []byte,
	serializedPlanHash []byte,
	stats *roachpb.CollectedStatementStatistics,
) (rowsAffected int, err error) {
	__antithesis_instrumentation__.Notify(624814)
	insertStmt := `
INSERT INTO system.statement_statistics
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
ON CONFLICT (crdb_internal_aggregated_ts_app_name_fingerprint_id_node_id_plan_hash_transaction_fingerprint_id_shard_8,
             aggregated_ts, fingerprint_id, transaction_fingerprint_id, app_name, plan_hash, node_id)
DO NOTHING
`
	aggInterval := s.GetAggregationInterval()

	metadataJSON, err := sqlstatsutil.BuildStmtMetadataJSON(stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624817)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(624818)
	}
	__antithesis_instrumentation__.Notify(624815)
	metadata := tree.NewDJSON(metadataJSON)

	statisticsJSON, err := sqlstatsutil.BuildStmtStatisticsJSON(&stats.Stats)
	if err != nil {
		__antithesis_instrumentation__.Notify(624819)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(624820)
	}
	__antithesis_instrumentation__.Notify(624816)
	statistics := tree.NewDJSON(statisticsJSON)

	plan := tree.NewDJSON(sqlstatsutil.ExplainTreePlanNodeToJSON(&stats.Stats.SensitiveInfo.MostRecentPlanDescription))

	rowsAffected, err = s.cfg.InternalExecutor.ExecEx(
		ctx,
		"insert-stmt-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		insertStmt,
		aggregatedTs,
		serializedFingerprintID,
		serializedTransactionFingerprintID,
		serializedPlanHash,
		stats.Key.App,
		s.cfg.SQLIDContainer.SQLInstanceID(),
		aggInterval,
		metadata,
		statistics,
		plan,
	)

	return rowsAffected, err
}

func (s *PersistedSQLStats) fetchPersistedTransactionStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	appName string,
	result *roachpb.TransactionStatistics,
) error {
	__antithesis_instrumentation__.Notify(624821)

	readStmt := `
SELECT
    statistics
FROM
    system.transaction_statistics
WHERE fingerprint_id = $1
    AND app_name = $2
	  AND aggregated_ts = $3
    AND node_id = $4
FOR UPDATE
`

	row, err := s.cfg.InternalExecutor.QueryRowEx(
		ctx,
		"fetch-txn-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		readStmt,
		serializedFingerprintID,
		appName,
		aggregatedTs,
		s.cfg.SQLIDContainer.SQLInstanceID(),
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(624825)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624826)
	}
	__antithesis_instrumentation__.Notify(624822)

	if row == nil {
		__antithesis_instrumentation__.Notify(624827)
		return errors.AssertionFailedf("transaction statistics not found for fingerprint_id: %s, app: %s, aggregated_ts: %s, node_id: %d",
			serializedFingerprintID, appName, aggregatedTs,
			s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624828)
	}
	__antithesis_instrumentation__.Notify(624823)

	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(624829)
		return errors.AssertionFailedf("unexpectedly found %d returning columns for fingerprint_id: %s, app: %s, aggregated_ts: %s, node_id: %d",
			len(row), serializedFingerprintID, appName, aggregatedTs,
			s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624830)
	}
	__antithesis_instrumentation__.Notify(624824)

	statistics := tree.MustBeDJSON(row[0])
	return sqlstatsutil.DecodeTxnStatsStatisticsJSON(statistics.JSON, result)
}

func (s *PersistedSQLStats) fetchPersistedStatementStats(
	ctx context.Context,
	txn *kv.Txn,
	aggregatedTs time.Time,
	serializedFingerprintID []byte,
	serializedTransactionFingerprintID []byte,
	serializedPlanHash []byte,
	key *roachpb.StatementStatisticsKey,
	result *roachpb.StatementStatistics,
) error {
	__antithesis_instrumentation__.Notify(624831)
	readStmt := `
SELECT
    statistics
FROM
    system.statement_statistics
WHERE fingerprint_id = $1
    AND transaction_fingerprint_id = $2
    AND app_name = $3
	  AND aggregated_ts = $4
    AND plan_hash = $5
    AND node_id = $6
FOR UPDATE
`
	row, err := s.cfg.InternalExecutor.QueryRowEx(
		ctx,
		"fetch-stmt-stats",
		txn,
		sessiondata.InternalExecutorOverride{
			User: security.NodeUserName(),
		},
		readStmt,
		serializedFingerprintID,
		serializedTransactionFingerprintID,
		key.App,
		aggregatedTs,
		serializedPlanHash,
		s.cfg.SQLIDContainer.SQLInstanceID(),
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(624835)
		return err
	} else {
		__antithesis_instrumentation__.Notify(624836)
	}
	__antithesis_instrumentation__.Notify(624832)

	if row == nil {
		__antithesis_instrumentation__.Notify(624837)
		return errors.AssertionFailedf(
			"statement statistics not found fingerprint_id: %s, app: %s, aggregated_ts: %s, plan_hash: %d, node_id: %d",
			serializedFingerprintID, key.App, aggregatedTs, serializedPlanHash, s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624838)
	}
	__antithesis_instrumentation__.Notify(624833)

	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(624839)
		return errors.AssertionFailedf("unexpectedly found %d returning columns for fingerprint_id: %s, app: %s, aggregated_ts: %s, plan_hash %d, node_id: %d",
			len(row), serializedFingerprintID, key.App, aggregatedTs, serializedPlanHash,
			s.cfg.SQLIDContainer.SQLInstanceID())
	} else {
		__antithesis_instrumentation__.Notify(624840)
	}
	__antithesis_instrumentation__.Notify(624834)

	statistics := tree.MustBeDJSON(row[0])

	return sqlstatsutil.DecodeStmtStatsStatisticsJSON(statistics.JSON, result)
}
