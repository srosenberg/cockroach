package persistedsqlstats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/sslocal"
)

type memStmtStatsIterator struct {
	*sslocal.StmtStatsIterator
	aggregatedTs time.Time
	aggInterval  time.Duration
}

func newMemStmtStatsIterator(
	stats *sslocal.SQLStats,
	options *sqlstats.IteratorOptions,
	aggregatedTS time.Time,
	aggInterval time.Duration,
) *memStmtStatsIterator {
	__antithesis_instrumentation__.Notify(624841)
	return &memStmtStatsIterator{
		StmtStatsIterator: stats.StmtStatsIterator(options),
		aggregatedTs:      aggregatedTS,
		aggInterval:       aggInterval,
	}
}

func (m *memStmtStatsIterator) Cur() *roachpb.CollectedStatementStatistics {
	__antithesis_instrumentation__.Notify(624842)
	c := m.StmtStatsIterator.Cur()
	c.AggregatedTs = m.aggregatedTs
	c.AggregationInterval = m.aggInterval
	return c
}

type memTxnStatsIterator struct {
	*sslocal.TxnStatsIterator
	aggregatedTs time.Time
	aggInterval  time.Duration
}

func newMemTxnStatsIterator(
	stats *sslocal.SQLStats,
	options *sqlstats.IteratorOptions,
	aggregatedTS time.Time,
	aggInterval time.Duration,
) *memTxnStatsIterator {
	__antithesis_instrumentation__.Notify(624843)
	return &memTxnStatsIterator{
		TxnStatsIterator: stats.TxnStatsIterator(options),
		aggregatedTs:     aggregatedTS,
		aggInterval:      aggInterval,
	}
}

func (m *memTxnStatsIterator) Cur() *roachpb.CollectedTransactionStatistics {
	__antithesis_instrumentation__.Notify(624844)
	stats := m.TxnStatsIterator.Cur()
	stats.AggregatedTs = m.aggregatedTs
	stats.AggregationInterval = m.aggInterval
	return stats
}
