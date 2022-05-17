package rttanalysis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/tracing/tracingpb"
	"github.com/stretchr/testify/require"
)

type RoundTripBenchTestCase struct {
	Name  string
	Setup string
	Stmt  string
	Reset string
}

func runRoundTripBenchmark(b testingB, tests []RoundTripBenchTestCase, cc ClusterConstructor) {
	__antithesis_instrumentation__.Notify(1842)
	for _, tc := range tests {
		__antithesis_instrumentation__.Notify(1843)
		b.Run(tc.Name, func(b testingB) {
			__antithesis_instrumentation__.Notify(1844)
			executeRoundTripTest(b, tc, cc)
		})
	}
}

func runRoundTripBenchmarkTest(
	t *testing.T,
	scope *log.TestLogScope,
	results *resultSet,
	tests []RoundTripBenchTestCase,
	cc ClusterConstructor,
	numRuns int,
	limit *quotapool.IntPool,
) {
	__antithesis_instrumentation__.Notify(1845)
	skip.UnderMetamorphic(t, "changes the RTTs")
	var wg sync.WaitGroup
	for _, tc := range tests {
		__antithesis_instrumentation__.Notify(1847)
		wg.Add(1)
		go func(tc RoundTripBenchTestCase) {
			__antithesis_instrumentation__.Notify(1848)
			defer wg.Done()
			t.Run(tc.Name, func(t *testing.T) {
				__antithesis_instrumentation__.Notify(1849)
				runRoundTripBenchmarkTestCase(t, scope, results, tc, cc, numRuns, limit)
			})
		}(tc)
	}
	__antithesis_instrumentation__.Notify(1846)
	wg.Wait()
}

func runRoundTripBenchmarkTestCase(
	t *testing.T,
	scope *log.TestLogScope,
	results *resultSet,
	tc RoundTripBenchTestCase,
	cc ClusterConstructor,
	numRuns int,
	limit *quotapool.IntPool,
) {
	__antithesis_instrumentation__.Notify(1850)
	var wg sync.WaitGroup
	for i := 0; i < numRuns; i++ {
		__antithesis_instrumentation__.Notify(1852)
		alloc, err := limit.Acquire(context.Background(), 1)
		require.NoError(t, err)
		wg.Add(1)
		go func() {
			__antithesis_instrumentation__.Notify(1853)
			defer wg.Done()
			defer alloc.Release()
			executeRoundTripTest(tShim{
				T: t, results: results, scope: scope,
			}, tc, cc)
		}()
	}
	__antithesis_instrumentation__.Notify(1851)
	wg.Wait()
}

func executeRoundTripTest(b testingB, tc RoundTripBenchTestCase, cc ClusterConstructor) {
	__antithesis_instrumentation__.Notify(1854)
	getDir, cleanup := b.logScope()
	defer cleanup()

	cluster := cc(b)
	defer cluster.close()

	sql := sqlutils.MakeSQLRunner(cluster.conn())

	expData := readExpectationsFile(b)

	exp, haveExp := expData.find(strings.TrimPrefix(b.Name(), "Benchmark"))

	roundTrips := 0
	b.ResetTimer()
	b.StopTimer()
	var r tracing.Recording

	for i := 0; i < b.N()+1; i++ {
		__antithesis_instrumentation__.Notify(1857)
		sql.Exec(b, "CREATE DATABASE bench;")
		sql.Exec(b, tc.Setup)
		cluster.clearStatementTrace(tc.Stmt)

		b.StartTimer()
		sql.Exec(b, tc.Stmt)
		b.StopTimer()
		var ok bool
		r, ok = cluster.getStatementTrace(tc.Stmt)
		if !ok {
			__antithesis_instrumentation__.Notify(1860)
			b.Fatalf(
				"could not find number of round trips for statement: %s",
				tc.Stmt,
			)
		} else {
			__antithesis_instrumentation__.Notify(1861)
		}
		__antithesis_instrumentation__.Notify(1858)

		rt, hasRetry := countKvBatchRequestsInRecording(r)
		if hasRetry {
			__antithesis_instrumentation__.Notify(1862)
			i--
		} else {
			__antithesis_instrumentation__.Notify(1863)
			if i > 0 {
				__antithesis_instrumentation__.Notify(1864)
				roundTrips += rt
			} else {
				__antithesis_instrumentation__.Notify(1865)
			}
		}
		__antithesis_instrumentation__.Notify(1859)

		sql.Exec(b, "DROP DATABASE bench;")
		sql.Exec(b, tc.Reset)
	}
	__antithesis_instrumentation__.Notify(1855)

	res := float64(roundTrips) / float64(b.N())

	if haveExp && func() bool {
		__antithesis_instrumentation__.Notify(1866)
		return !exp.matches(int(res)) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(1867)
		return !*rewriteFlag == true
	}() == true {
		__antithesis_instrumentation__.Notify(1868)
		b.Errorf(`%s: got %v, expected %v`, b.Name(), res, exp)
		dir := getDir()
		jaegerJSON, err := r.ToJaegerJSON(tc.Stmt, "", "n0")
		require.NoError(b, err)
		path := filepath.Join(dir, strings.Replace(b.Name(), "/", "_", -1)) + ".jaeger.json"
		require.NoError(b, ioutil.WriteFile(path, []byte(jaegerJSON), 0666))
		b.Errorf("wrote jaeger trace to %s", path)
	} else {
		__antithesis_instrumentation__.Notify(1869)
	}
	__antithesis_instrumentation__.Notify(1856)
	b.ReportMetric(res, roundTripsMetric)
}

const roundTripsMetric = "roundtrips"

func countKvBatchRequestsInRecording(r tracing.Recording) (sends int, hasRetry bool) {
	__antithesis_instrumentation__.Notify(1870)
	root := r[0]
	return countKvBatchRequestsInSpan(r, root)
}

func countKvBatchRequestsInSpan(r tracing.Recording, sp tracingpb.RecordedSpan) (int, bool) {
	__antithesis_instrumentation__.Notify(1871)
	count := 0

	if sp.Operation == kvcoord.OpTxnCoordSender {
		__antithesis_instrumentation__.Notify(1875)
		count++
	} else {
		__antithesis_instrumentation__.Notify(1876)
	}
	__antithesis_instrumentation__.Notify(1872)
	if logsContainRetry(sp.Logs) {
		__antithesis_instrumentation__.Notify(1877)
		return 0, true
	} else {
		__antithesis_instrumentation__.Notify(1878)
	}
	__antithesis_instrumentation__.Notify(1873)

	for _, osp := range r {
		__antithesis_instrumentation__.Notify(1879)
		if osp.ParentSpanID != sp.SpanID {
			__antithesis_instrumentation__.Notify(1882)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1883)
		}
		__antithesis_instrumentation__.Notify(1880)

		subCount, hasRetry := countKvBatchRequestsInSpan(r, osp)
		if hasRetry {
			__antithesis_instrumentation__.Notify(1884)
			return 0, true
		} else {
			__antithesis_instrumentation__.Notify(1885)
		}
		__antithesis_instrumentation__.Notify(1881)
		count += subCount
	}
	__antithesis_instrumentation__.Notify(1874)

	return count, false
}

func logsContainRetry(logs []tracingpb.LogRecord) bool {
	__antithesis_instrumentation__.Notify(1886)
	for _, l := range logs {
		__antithesis_instrumentation__.Notify(1888)
		if strings.Contains(l.String(), "TransactionRetryWithProtoRefreshError") {
			__antithesis_instrumentation__.Notify(1889)
			return true
		} else {
			__antithesis_instrumentation__.Notify(1890)
		}
	}
	__antithesis_instrumentation__.Notify(1887)
	return false
}
