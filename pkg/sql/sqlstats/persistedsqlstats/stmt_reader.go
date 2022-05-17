package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/persistedsqlstats/sqlstatsutil"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/errors"
)

func (s *PersistedSQLStats) IterateStatementStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.StatementVisitor,
) (err error) {
	__antithesis_instrumentation__.Notify(625268)

	options.SortedKey = true
	options.SortedAppNames = true

	curAggTs := s.ComputeAggregatedTs()
	aggInterval := s.GetAggregationInterval()
	memIter := newMemStmtStatsIterator(s.SQLStats, options, curAggTs, aggInterval)

	var persistedIter sqlutil.InternalRows
	var colCnt int
	persistedIter, colCnt, err = s.persistedStmtStatsIter(ctx, options)
	if err != nil {
		__antithesis_instrumentation__.Notify(625272)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625273)
	}
	__antithesis_instrumentation__.Notify(625269)
	defer func() {
		__antithesis_instrumentation__.Notify(625274)
		closeError := persistedIter.Close()
		if closeError != nil {
			__antithesis_instrumentation__.Notify(625275)
			err = errors.CombineErrors(err, closeError)
		} else {
			__antithesis_instrumentation__.Notify(625276)
		}
	}()
	__antithesis_instrumentation__.Notify(625270)

	combinedIter := NewCombinedStmtStatsIterator(memIter, persistedIter, colCnt)

	for {
		__antithesis_instrumentation__.Notify(625277)
		var ok bool
		ok, err = combinedIter.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(625280)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625281)
		}
		__antithesis_instrumentation__.Notify(625278)

		if !ok {
			__antithesis_instrumentation__.Notify(625282)
			break
		} else {
			__antithesis_instrumentation__.Notify(625283)
		}
		__antithesis_instrumentation__.Notify(625279)

		stats := combinedIter.Cur()
		if err = visitor(ctx, stats); err != nil {
			__antithesis_instrumentation__.Notify(625284)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625285)
		}
	}
	__antithesis_instrumentation__.Notify(625271)

	return nil
}

func (s *PersistedSQLStats) persistedStmtStatsIter(
	ctx context.Context, options *sqlstats.IteratorOptions,
) (iter sqlutil.InternalRows, expectedColCnt int, err error) {
	__antithesis_instrumentation__.Notify(625286)
	query, expectedColCnt := s.getFetchQueryForStmtStatsTable(options)

	persistedIter, err := s.cfg.InternalExecutor.QueryIteratorEx(
		ctx,
		"read-stmt-stats",
		nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		query,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(625288)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(625289)
	}
	__antithesis_instrumentation__.Notify(625287)

	return persistedIter, expectedColCnt, err
}

func (s *PersistedSQLStats) getFetchQueryForStmtStatsTable(
	options *sqlstats.IteratorOptions,
) (query string, colCnt int) {
	__antithesis_instrumentation__.Notify(625290)
	selectedColumns := []string{
		"aggregated_ts",
		"fingerprint_id",
		"transaction_fingerprint_id",
		"plan_hash",
		"app_name",
		"metadata",
		"statistics",
		"plan",
		"agg_interval",
	}

	query = `
SELECT 
  %[1]s
FROM
	system.statement_statistics
%[2]s`

	followerReadClause := s.cfg.Knobs.GetAOSTClause()

	query = fmt.Sprintf(query, strings.Join(selectedColumns, ","), followerReadClause)

	orderByColumns := []string{"aggregated_ts"}
	if options.SortedAppNames {
		__antithesis_instrumentation__.Notify(625293)
		orderByColumns = append(orderByColumns, "app_name")
	} else {
		__antithesis_instrumentation__.Notify(625294)
	}
	__antithesis_instrumentation__.Notify(625291)

	if options.SortedKey {
		__antithesis_instrumentation__.Notify(625295)
		orderByColumns = append(orderByColumns, "metadata ->> 'query'")
	} else {
		__antithesis_instrumentation__.Notify(625296)
	}
	__antithesis_instrumentation__.Notify(625292)

	orderByColumns = append(orderByColumns, "transaction_fingerprint_id")
	query = fmt.Sprintf("%s ORDER BY %s", query, strings.Join(orderByColumns, ","))

	return query, len(selectedColumns)
}

func rowToStmtStats(row tree.Datums) (*roachpb.CollectedStatementStatistics, error) {
	__antithesis_instrumentation__.Notify(625297)
	var stats roachpb.CollectedStatementStatistics
	stats.AggregatedTs = tree.MustBeDTimestampTZ(row[0]).Time

	stmtFingerprintID, err := sqlstatsutil.DatumToUint64(row[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(625304)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625305)
	}
	__antithesis_instrumentation__.Notify(625298)
	stats.ID = roachpb.StmtFingerprintID(stmtFingerprintID)

	transactionFingerprintID, err := sqlstatsutil.DatumToUint64(row[2])
	if err != nil {
		__antithesis_instrumentation__.Notify(625306)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625307)
	}
	__antithesis_instrumentation__.Notify(625299)
	stats.Key.TransactionFingerprintID =
		roachpb.TransactionFingerprintID(transactionFingerprintID)

	stats.Key.PlanHash, err = sqlstatsutil.DatumToUint64(row[3])
	if err != nil {
		__antithesis_instrumentation__.Notify(625308)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625309)
	}
	__antithesis_instrumentation__.Notify(625300)

	stats.Key.App = string(tree.MustBeDString(row[4]))

	metadata := tree.MustBeDJSON(row[5]).JSON
	if err = sqlstatsutil.DecodeStmtStatsMetadataJSON(metadata, &stats); err != nil {
		__antithesis_instrumentation__.Notify(625310)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625311)
	}
	__antithesis_instrumentation__.Notify(625301)

	statistics := tree.MustBeDJSON(row[6]).JSON
	if err = sqlstatsutil.DecodeStmtStatsStatisticsJSON(statistics, &stats.Stats); err != nil {
		__antithesis_instrumentation__.Notify(625312)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625313)
	}
	__antithesis_instrumentation__.Notify(625302)

	jsonPlan := tree.MustBeDJSON(row[7]).JSON
	plan, err := sqlstatsutil.JSONToExplainTreePlanNode(jsonPlan)
	if err != nil {
		__antithesis_instrumentation__.Notify(625314)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625315)
	}
	__antithesis_instrumentation__.Notify(625303)
	stats.Stats.SensitiveInfo.MostRecentPlanDescription = *plan

	aggInterval := tree.MustBeDInterval(row[8]).Duration
	stats.AggregationInterval = time.Duration(aggInterval.Nanos())

	return &stats, nil
}
