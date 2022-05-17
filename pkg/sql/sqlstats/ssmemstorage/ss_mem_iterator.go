package ssmemstorage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats"
)

type baseIterator struct {
	container *Container
	idx       int
}

type StmtStatsIterator struct {
	baseIterator
	stmtKeys     stmtList
	currentValue *roachpb.CollectedStatementStatistics
}

func NewStmtStatsIterator(
	container *Container, options *sqlstats.IteratorOptions,
) *StmtStatsIterator {
	__antithesis_instrumentation__.Notify(625487)
	var stmtKeys stmtList
	container.mu.Lock()
	for k := range container.mu.stmts {
		__antithesis_instrumentation__.Notify(625490)
		stmtKeys = append(stmtKeys, k)
	}
	__antithesis_instrumentation__.Notify(625488)
	container.mu.Unlock()
	if options.SortedKey {
		__antithesis_instrumentation__.Notify(625491)
		sort.Sort(stmtKeys)
	} else {
		__antithesis_instrumentation__.Notify(625492)
	}
	__antithesis_instrumentation__.Notify(625489)

	return &StmtStatsIterator{
		baseIterator: baseIterator{
			container: container,
			idx:       -1,
		},
		stmtKeys: stmtKeys,
	}
}

func (s *StmtStatsIterator) Next() bool {
	__antithesis_instrumentation__.Notify(625493)
	s.idx++
	if s.idx >= len(s.stmtKeys) {
		__antithesis_instrumentation__.Notify(625496)
		return false
	} else {
		__antithesis_instrumentation__.Notify(625497)
	}
	__antithesis_instrumentation__.Notify(625494)

	stmtKey := s.stmtKeys[s.idx]

	stmtFingerprintID := constructStatementFingerprintIDFromStmtKey(stmtKey)
	statementStats, _, _ :=
		s.container.getStatsForStmtWithKey(stmtKey, invalidStmtFingerprintID, false)

	if statementStats == nil {
		__antithesis_instrumentation__.Notify(625498)
		return s.Next()
	} else {
		__antithesis_instrumentation__.Notify(625499)
	}
	__antithesis_instrumentation__.Notify(625495)

	statementStats.mu.Lock()
	data := statementStats.mu.data
	distSQLUsed := statementStats.mu.distSQLUsed
	vectorized := statementStats.mu.vectorized
	fullScan := statementStats.mu.fullScan
	database := statementStats.mu.database
	querySummary := statementStats.mu.querySummary
	statementStats.mu.Unlock()

	s.currentValue = &roachpb.CollectedStatementStatistics{
		Key: roachpb.StatementStatisticsKey{
			Query:                    stmtKey.anonymizedStmt,
			QuerySummary:             querySummary,
			DistSQL:                  distSQLUsed,
			Vec:                      vectorized,
			ImplicitTxn:              stmtKey.implicitTxn,
			FullScan:                 fullScan,
			Failed:                   stmtKey.failed,
			App:                      s.container.appName,
			Database:                 database,
			PlanHash:                 stmtKey.planHash,
			TransactionFingerprintID: stmtKey.transactionFingerprintID,
		},
		ID:    stmtFingerprintID,
		Stats: data,
	}

	return true
}

func (s *StmtStatsIterator) Cur() *roachpb.CollectedStatementStatistics {
	__antithesis_instrumentation__.Notify(625500)
	return s.currentValue
}

type TxnStatsIterator struct {
	baseIterator
	txnKeys  txnList
	curValue *roachpb.CollectedTransactionStatistics
}

func NewTxnStatsIterator(
	container *Container, options *sqlstats.IteratorOptions,
) *TxnStatsIterator {
	__antithesis_instrumentation__.Notify(625501)
	var txnKeys txnList
	container.mu.Lock()
	for k := range container.mu.txns {
		__antithesis_instrumentation__.Notify(625504)
		txnKeys = append(txnKeys, k)
	}
	__antithesis_instrumentation__.Notify(625502)
	container.mu.Unlock()
	if options.SortedKey {
		__antithesis_instrumentation__.Notify(625505)
		sort.Sort(txnKeys)
	} else {
		__antithesis_instrumentation__.Notify(625506)
	}
	__antithesis_instrumentation__.Notify(625503)

	return &TxnStatsIterator{
		baseIterator: baseIterator{
			container: container,
			idx:       -1,
		},
		txnKeys: txnKeys,
	}
}

func (t *TxnStatsIterator) Next() bool {
	__antithesis_instrumentation__.Notify(625507)
	t.idx++
	if t.idx >= len(t.txnKeys) {
		__antithesis_instrumentation__.Notify(625510)
		return false
	} else {
		__antithesis_instrumentation__.Notify(625511)
	}
	__antithesis_instrumentation__.Notify(625508)

	txnKey := t.txnKeys[t.idx]

	txnStats, _, _ := t.container.getStatsForTxnWithKey(txnKey, nil, false)

	if txnStats == nil {
		__antithesis_instrumentation__.Notify(625512)
		return t.Next()
	} else {
		__antithesis_instrumentation__.Notify(625513)
	}
	__antithesis_instrumentation__.Notify(625509)

	txnStats.mu.Lock()
	defer txnStats.mu.Unlock()

	t.curValue = &roachpb.CollectedTransactionStatistics{
		StatementFingerprintIDs:  txnStats.statementFingerprintIDs,
		App:                      t.container.appName,
		Stats:                    txnStats.mu.data,
		TransactionFingerprintID: txnKey,
	}

	return true
}

func (t *TxnStatsIterator) Cur() *roachpb.CollectedTransactionStatistics {
	__antithesis_instrumentation__.Notify(625514)
	return t.curValue
}
