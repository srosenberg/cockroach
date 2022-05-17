package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
)

type baseIterator struct {
	sqlStats *SQLStats
	idx      int
	appNames []string
	options  *sqlstats.IteratorOptions
}

type StmtStatsIterator struct {
	baseIterator
	curIter *ssmemstorage.StmtStatsIterator
}

func NewStmtStatsIterator(s *SQLStats, options *sqlstats.IteratorOptions) *StmtStatsIterator {
	__antithesis_instrumentation__.Notify(625383)
	appNames := s.getAppNames(options.SortedAppNames)

	return &StmtStatsIterator{
		baseIterator: baseIterator{
			sqlStats: s,
			idx:      -1,
			appNames: appNames,
			options:  options,
		},
	}
}

func (s *StmtStatsIterator) Next() bool {
	__antithesis_instrumentation__.Notify(625384)

	if s.curIter == nil || func() bool {
		__antithesis_instrumentation__.Notify(625386)
		return !s.curIter.Next() == true
	}() == true {
		__antithesis_instrumentation__.Notify(625387)
		s.idx++
		if s.idx >= len(s.appNames) {
			__antithesis_instrumentation__.Notify(625389)
			return false
		} else {
			__antithesis_instrumentation__.Notify(625390)
		}
		__antithesis_instrumentation__.Notify(625388)
		statsContainer := s.sqlStats.getStatsForApplication(s.appNames[s.idx])
		s.curIter = statsContainer.StmtStatsIterator(s.options)
		return s.Next()
	} else {
		__antithesis_instrumentation__.Notify(625391)
	}
	__antithesis_instrumentation__.Notify(625385)

	return true
}

func (s *StmtStatsIterator) Cur() *roachpb.CollectedStatementStatistics {
	__antithesis_instrumentation__.Notify(625392)
	return s.curIter.Cur()
}

type TxnStatsIterator struct {
	baseIterator
	curIter *ssmemstorage.TxnStatsIterator
}

func NewTxnStatsIterator(s *SQLStats, options *sqlstats.IteratorOptions) *TxnStatsIterator {
	__antithesis_instrumentation__.Notify(625393)
	appNames := s.getAppNames(options.SortedAppNames)

	return &TxnStatsIterator{
		baseIterator: baseIterator{
			sqlStats: s,
			idx:      -1,
			appNames: appNames,
			options:  options,
		},
	}
}

func (t *TxnStatsIterator) Next() bool {
	__antithesis_instrumentation__.Notify(625394)

	if t.curIter == nil || func() bool {
		__antithesis_instrumentation__.Notify(625396)
		return !t.curIter.Next() == true
	}() == true {
		__antithesis_instrumentation__.Notify(625397)
		t.idx++
		if t.idx >= len(t.appNames) {
			__antithesis_instrumentation__.Notify(625399)
			return false
		} else {
			__antithesis_instrumentation__.Notify(625400)
		}
		__antithesis_instrumentation__.Notify(625398)
		statsContainer := t.sqlStats.getStatsForApplication(t.appNames[t.idx])
		t.curIter = statsContainer.TxnStatsIterator(t.options)
		return t.Next()
	} else {
		__antithesis_instrumentation__.Notify(625401)
	}
	__antithesis_instrumentation__.Notify(625395)

	return true
}

func (t *TxnStatsIterator) Cur() *roachpb.CollectedTransactionStatistics {
	__antithesis_instrumentation__.Notify(625402)
	return t.curIter.Cur()
}
