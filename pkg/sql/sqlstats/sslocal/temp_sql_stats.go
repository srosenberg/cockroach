package sslocal

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sort"

	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlstats/ssmemstorage"
)

type stmtResponseList []serverpb.StatementsResponse_CollectedStatementStatistics

var _ sort.Interface = stmtResponseList{}

func (s stmtResponseList) Len() int {
	__antithesis_instrumentation__.Notify(625469)
	return len(s)
}

func (s stmtResponseList) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(625470)
	return s[i].Key.KeyData.App < s[j].Key.KeyData.App
}

func (s stmtResponseList) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(625471)
	s[i], s[j] = s[j], s[i]
}

type txnResponseList []serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics

var _ sort.Interface = txnResponseList{}

func (t txnResponseList) Len() int {
	__antithesis_instrumentation__.Notify(625472)
	return len(t)
}

func (t txnResponseList) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(625473)
	return t[i].StatsData.App < t[j].StatsData.App
}

func (t txnResponseList) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(625474)
	t[i], t[j] = t[j], t[i]
}

func NewTempSQLStatsFromExistingStmtStats(
	statistics []serverpb.StatementsResponse_CollectedStatementStatistics,
) (*SQLStats, error) {
	__antithesis_instrumentation__.Notify(625475)
	sort.Sort(stmtResponseList(statistics))

	var err error
	s := &SQLStats{}
	s.mu.apps = make(map[string]*ssmemstorage.Container)

	for len(statistics) > 0 {
		__antithesis_instrumentation__.Notify(625477)
		appName := statistics[0].Key.KeyData.App
		var container *ssmemstorage.Container

		container, statistics, err =
			ssmemstorage.NewTempContainerFromExistingStmtStats(statistics)
		if err != nil {
			__antithesis_instrumentation__.Notify(625479)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625480)
		}
		__antithesis_instrumentation__.Notify(625478)

		s.mu.apps[appName] = container
	}
	__antithesis_instrumentation__.Notify(625476)

	return s, nil
}

func NewTempSQLStatsFromExistingTxnStats(
	statistics []serverpb.StatementsResponse_ExtendedCollectedTransactionStatistics,
) (*SQLStats, error) {
	__antithesis_instrumentation__.Notify(625481)
	sort.Sort(txnResponseList(statistics))

	var err error
	s := &SQLStats{}
	s.mu.apps = make(map[string]*ssmemstorage.Container)

	for len(statistics) > 0 {
		__antithesis_instrumentation__.Notify(625483)
		appName := statistics[0].StatsData.App
		var container *ssmemstorage.Container

		container, statistics, err =
			ssmemstorage.NewTempContainerFromExistingTxnStats(statistics)
		if err != nil {
			__antithesis_instrumentation__.Notify(625485)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625486)
		}
		__antithesis_instrumentation__.Notify(625484)

		s.mu.apps[appName] = container
	}
	__antithesis_instrumentation__.Notify(625482)

	return s, nil
}
