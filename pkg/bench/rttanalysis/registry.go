package rttanalysis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type Registry struct {
	numNodes int
	cc       ClusterConstructor
	r        map[string][]RoundTripBenchTestCase
}

func NewRegistry(numNodes int, cc ClusterConstructor) *Registry {
	__antithesis_instrumentation__.Notify(1834)
	return &Registry{
		numNodes: numNodes,
		cc:       cc,
		r:        make(map[string][]RoundTripBenchTestCase),
	}
}

func (r *Registry) Run(b *testing.B) {
	__antithesis_instrumentation__.Notify(1835)
	tests, ok := r.r[bName(b)]
	require.True(b, ok)
	runRoundTripBenchmark(bShim{b}, tests, r.cc)
}

func (r *Registry) RunExpectations(t *testing.T) {
	__antithesis_instrumentation__.Notify(1836)
	skip.UnderStress(t)
	skip.UnderRace(t)
	skip.UnderShort(t)

	runBenchmarkExpectationTests(t, r)
}

func (r *Registry) Register(name string, tests []RoundTripBenchTestCase) {
	__antithesis_instrumentation__.Notify(1837)
	if _, exists := r.r[name]; exists {
		__antithesis_instrumentation__.Notify(1839)
		panic(errors.Errorf("Benchmark%s already registered", name))
	} else {
		__antithesis_instrumentation__.Notify(1840)
	}
	__antithesis_instrumentation__.Notify(1838)
	r.r[name] = tests
}

func bName(b *testing.B) string {
	__antithesis_instrumentation__.Notify(1841)
	return strings.TrimPrefix(b.Name(), "Benchmark")
}
