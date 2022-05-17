package rttanalysis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/util/log"
)

type testingB interface {
	testing.TB
	N() int
	ResetTimer()
	StopTimer()
	StartTimer()
	ReportMetric(float64, string)
	Run(string, func(testingB))

	logScope() (getDirectory func() string, close func())
}

type bShim struct {
	*testing.B
}

var _ testingB = bShim{}

func (b bShim) logScope() (getDirectory func() string, close func()) {
	__antithesis_instrumentation__.Notify(1891)
	sc := log.Scope(b)
	return sc.GetDirectory, func() { __antithesis_instrumentation__.Notify(1892); sc.Close(b) }
}
func (b bShim) N() int { __antithesis_instrumentation__.Notify(1893); return b.B.N }
func (b bShim) Run(name string, f func(b testingB)) {
	__antithesis_instrumentation__.Notify(1894)
	b.B.Run(name, func(b *testing.B) {
		__antithesis_instrumentation__.Notify(1895)
		f(bShim{b})
	})
}

type tShim struct {
	*testing.T
	scope   *log.TestLogScope
	results *resultSet
}

var _ testingB = tShim{}

func (ts tShim) logScope() (getDirectory func() string, close func()) {
	__antithesis_instrumentation__.Notify(1896)
	return ts.scope.GetDirectory, func() { __antithesis_instrumentation__.Notify(1897) }
}
func (ts tShim) GetDirectory() string {
	__antithesis_instrumentation__.Notify(1898)
	return ts.scope.GetDirectory()
}
func (ts tShim) N() int      { __antithesis_instrumentation__.Notify(1899); return 2 }
func (ts tShim) ResetTimer() { __antithesis_instrumentation__.Notify(1900) }
func (ts tShim) StopTimer()  { __antithesis_instrumentation__.Notify(1901) }
func (ts tShim) StartTimer() { __antithesis_instrumentation__.Notify(1902) }
func (ts tShim) ReportMetric(f float64, s string) {
	__antithesis_instrumentation__.Notify(1903)
	if s == roundTripsMetric {
		__antithesis_instrumentation__.Notify(1904)
		ts.results.add(benchmarkResult{
			name:   ts.Name(),
			result: int(f),
		})
	} else {
		__antithesis_instrumentation__.Notify(1905)
	}
}
func (ts tShim) Name() string {
	__antithesis_instrumentation__.Notify(1906)

	tn := ts.T.Name()
	return tn[strings.Index(tn, "/")+1:]
}
func (ts tShim) Run(s string, f func(testingB)) {
	__antithesis_instrumentation__.Notify(1907)
	ts.T.Run(s, func(t *testing.T) {
		__antithesis_instrumentation__.Notify(1908)
		f(tShim{results: ts.results, T: t, scope: ts.scope})
	})
}
