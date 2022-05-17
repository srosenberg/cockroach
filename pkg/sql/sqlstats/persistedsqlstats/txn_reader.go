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

func (s *PersistedSQLStats) IterateTransactionStats(
	ctx context.Context, options *sqlstats.IteratorOptions, visitor sqlstats.TransactionVisitor,
) (err error) {
	__antithesis_instrumentation__.Notify(625316)

	options.SortedKey = true
	options.SortedAppNames = true

	curAggTs := s.ComputeAggregatedTs()
	aggInterval := s.GetAggregationInterval()
	memIter := newMemTxnStatsIterator(s.SQLStats, options, curAggTs, aggInterval)

	var persistedIter sqlutil.InternalRows
	var colCnt int
	persistedIter, colCnt, err = s.persistedTxnStatsIter(ctx, options)
	if err != nil {
		__antithesis_instrumentation__.Notify(625320)
		return err
	} else {
		__antithesis_instrumentation__.Notify(625321)
	}
	__antithesis_instrumentation__.Notify(625317)
	defer func() {
		__antithesis_instrumentation__.Notify(625322)
		closeError := persistedIter.Close()
		if closeError != nil {
			__antithesis_instrumentation__.Notify(625323)
			err = errors.CombineErrors(err, closeError)
		} else {
			__antithesis_instrumentation__.Notify(625324)
		}
	}()
	__antithesis_instrumentation__.Notify(625318)

	combinedIter := NewCombinedTxnStatsIterator(memIter, persistedIter, colCnt)

	for {
		__antithesis_instrumentation__.Notify(625325)
		var ok bool
		ok, err = combinedIter.Next(ctx)
		if err != nil {
			__antithesis_instrumentation__.Notify(625328)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625329)
		}
		__antithesis_instrumentation__.Notify(625326)

		if !ok {
			__antithesis_instrumentation__.Notify(625330)
			break
		} else {
			__antithesis_instrumentation__.Notify(625331)
		}
		__antithesis_instrumentation__.Notify(625327)

		stats := combinedIter.Cur()
		if err = visitor(ctx, stats); err != nil {
			__antithesis_instrumentation__.Notify(625332)
			return err
		} else {
			__antithesis_instrumentation__.Notify(625333)
		}
	}
	__antithesis_instrumentation__.Notify(625319)

	return nil
}

func (s *PersistedSQLStats) persistedTxnStatsIter(
	ctx context.Context, options *sqlstats.IteratorOptions,
) (iter sqlutil.InternalRows, expectedColCnt int, err error) {
	__antithesis_instrumentation__.Notify(625334)
	query, expectedColCnt := s.getFetchQueryForTxnStatsTable(options)

	persistedIter, err := s.cfg.InternalExecutor.QueryIteratorEx(
		ctx,
		"read-txn-stats",
		nil,
		sessiondata.InternalExecutorOverride{User: security.NodeUserName()},
		query,
	)

	if err != nil {
		__antithesis_instrumentation__.Notify(625336)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(625337)
	}
	__antithesis_instrumentation__.Notify(625335)

	return persistedIter, expectedColCnt, err
}

func (s *PersistedSQLStats) getFetchQueryForTxnStatsTable(
	options *sqlstats.IteratorOptions,
) (query string, colCnt int) {
	__antithesis_instrumentation__.Notify(625338)
	selectedColumns := []string{
		"aggregated_ts",
		"fingerprint_id",
		"app_name",
		"metadata",
		"statistics",
		"agg_interval",
	}

	query = `
SELECT 
  %[1]s
FROM
	system.transaction_statistics
%[2]s`

	followerReadClause := s.cfg.Knobs.GetAOSTClause()

	query = fmt.Sprintf(query, strings.Join(selectedColumns, ","), followerReadClause)

	orderByColumns := []string{"aggregated_ts"}
	if options.SortedAppNames {
		__antithesis_instrumentation__.Notify(625341)
		orderByColumns = append(orderByColumns, "app_name")
	} else {
		__antithesis_instrumentation__.Notify(625342)
	}
	__antithesis_instrumentation__.Notify(625339)

	if options.SortedKey {
		__antithesis_instrumentation__.Notify(625343)
		orderByColumns = append(orderByColumns, "fingerprint_id")
	} else {
		__antithesis_instrumentation__.Notify(625344)
	}
	__antithesis_instrumentation__.Notify(625340)

	query = fmt.Sprintf("%s ORDER BY %s", query, strings.Join(orderByColumns, ","))

	return query, len(selectedColumns)
}

func rowToTxnStats(row tree.Datums) (*roachpb.CollectedTransactionStatistics, error) {
	__antithesis_instrumentation__.Notify(625345)
	var stats roachpb.CollectedTransactionStatistics
	var err error

	stats.AggregatedTs = tree.MustBeDTimestampTZ(row[0]).Time

	value, err := sqlstatsutil.DatumToUint64(row[1])
	if err != nil {
		__antithesis_instrumentation__.Notify(625349)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625350)
	}
	__antithesis_instrumentation__.Notify(625346)
	stats.TransactionFingerprintID = roachpb.TransactionFingerprintID(value)

	stats.App = string(tree.MustBeDString(row[2]))

	metadata := tree.MustBeDJSON(row[3]).JSON
	if err = sqlstatsutil.DecodeTxnStatsMetadataJSON(metadata, &stats); err != nil {
		__antithesis_instrumentation__.Notify(625351)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625352)
	}
	__antithesis_instrumentation__.Notify(625347)

	statistics := tree.MustBeDJSON(row[4]).JSON
	if err = sqlstatsutil.DecodeTxnStatsStatisticsJSON(statistics, &stats.Stats); err != nil {
		__antithesis_instrumentation__.Notify(625353)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(625354)
	}
	__antithesis_instrumentation__.Notify(625348)

	aggInterval := tree.MustBeDInterval(row[5]).Duration
	stats.AggregationInterval = time.Duration(aggInterval.Nanos())

	return &stats, nil
}
