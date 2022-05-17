package rttanalysis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/csv"
	"flag"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/system"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

type benchmarkResult struct {
	name   string
	result int
}

type benchmarkExpectation struct {
	name string

	min, max int
}

const expectationsFilename = "benchmark_expectations"

var expectationsHeader = []string{"exp", "benchmark"}

var (
	rewriteFlag = flag.Bool("rewrite", false,
		"if non-empty, a regexp of benchmarks to rewrite")
	rewriteIterations = flag.Int("rewrite-iterations", 50,
		"if re-writing, the number of times to execute each benchmark to "+
			"determine the range of possible values")
)

func runBenchmarkExpectationTests(t *testing.T, r *Registry) {
	__antithesis_instrumentation__.Notify(1909)
	if util.IsMetamorphicBuild() {
		__antithesis_instrumentation__.Notify(1913)
		execTestSubprocess(t)
		return
	} else {
		__antithesis_instrumentation__.Notify(1914)
	}
	__antithesis_instrumentation__.Notify(1910)

	scope := log.Scope(t)
	defer scope.Close(t)

	defer func() {
		__antithesis_instrumentation__.Notify(1915)
		if t.Failed() {
			__antithesis_instrumentation__.Notify(1916)
			t.Log("see the --rewrite flag to re-run the benchmarks and adjust the expectations")
		} else {
			__antithesis_instrumentation__.Notify(1917)
		}
	}()
	__antithesis_instrumentation__.Notify(1911)

	var results resultSet
	var wg sync.WaitGroup
	concurrency := ((system.NumCPU() - 1) / r.numNodes) + 1
	limiter := quotapool.NewIntPool("rttanalysis", uint64(concurrency))
	isRewrite := *rewriteFlag
	for b, cases := range r.r {
		__antithesis_instrumentation__.Notify(1918)
		wg.Add(1)
		go func(b string, cases []RoundTripBenchTestCase) {
			__antithesis_instrumentation__.Notify(1919)
			defer wg.Done()
			t.Run(b, func(t *testing.T) {
				__antithesis_instrumentation__.Notify(1920)
				runs := 1
				if isRewrite {
					__antithesis_instrumentation__.Notify(1922)
					runs = *rewriteIterations
				} else {
					__antithesis_instrumentation__.Notify(1923)
				}
				__antithesis_instrumentation__.Notify(1921)
				runRoundTripBenchmarkTest(t, scope, &results, cases, r.cc, runs, limiter)
			})
		}(b, cases)
	}
	__antithesis_instrumentation__.Notify(1912)
	wg.Wait()

	if isRewrite {
		__antithesis_instrumentation__.Notify(1924)
		writeExpectationsFile(t,
			mergeExpectations(
				readExpectationsFile(t),
				resultsToExpectations(t, results.toSlice()),
			))
	} else {
		__antithesis_instrumentation__.Notify(1925)
		checkResults(t, &results, readExpectationsFile(t))
	}
}

func checkResults(t *testing.T, results *resultSet, expectations benchmarkExpectations) {
	__antithesis_instrumentation__.Notify(1926)
	results.iterate(func(r benchmarkResult) {
		__antithesis_instrumentation__.Notify(1927)
		exp, ok := expectations.find(r.name)
		if !ok {
			__antithesis_instrumentation__.Notify(1929)
			t.Logf("no expectation for benchmark %s, got %d", r.name, r.result)
			return
		} else {
			__antithesis_instrumentation__.Notify(1930)
		}
		__antithesis_instrumentation__.Notify(1928)
		if !exp.matches(r.result) {
			__antithesis_instrumentation__.Notify(1931)
			t.Errorf("fail: expected %s to perform KV lookups in [%d, %d], got %d",
				r.name, exp.min, exp.max, r.result)
		} else {
			__antithesis_instrumentation__.Notify(1932)
			t.Logf("success: expected %s to perform KV lookups in [%d, %d], got %d",
				r.name, exp.min, exp.max, r.result)
		}
	})
}

func mergeExpectations(existing, new benchmarkExpectations) (merged benchmarkExpectations) {
	__antithesis_instrumentation__.Notify(1933)
	sort.Sort(existing)
	sort.Sort(new)
	pop := func(be *benchmarkExpectations) (ret benchmarkExpectation) {
		__antithesis_instrumentation__.Notify(1936)
		ret = (*be)[0]
		*be = (*be)[1:]
		return ret
	}
	__antithesis_instrumentation__.Notify(1934)
	for len(existing) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(1937)
		return len(new) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(1938)
		switch {
		case existing[0].name < new[0].name:
			__antithesis_instrumentation__.Notify(1939)
			merged = append(merged, pop(&existing))
		case existing[0].name > new[0].name:
			__antithesis_instrumentation__.Notify(1940)
			merged = append(merged, pop(&new))
		default:
			__antithesis_instrumentation__.Notify(1941)
			pop(&existing)
			merged = append(merged, pop(&new))
		}
	}
	__antithesis_instrumentation__.Notify(1935)

	merged = append(append(merged, new...), existing...)
	return merged
}

func execTestSubprocess(t *testing.T) {
	__antithesis_instrumentation__.Notify(1942)
	var args []string
	flag.CommandLine.Visit(func(f *flag.Flag) {
		__antithesis_instrumentation__.Notify(1944)
		vs := f.Value.String()
		switch f.Name {
		case "test.run":
			__antithesis_instrumentation__.Notify(1946)

			prefix := "^" + regexp.QuoteMeta(t.Name()) + "$"
			if idx := strings.Index(vs, "/"); idx >= 0 {
				__antithesis_instrumentation__.Notify(1949)
				vs = prefix + vs[idx:]
			} else {
				__antithesis_instrumentation__.Notify(1950)
				vs = prefix
			}
		case "test.bench":
			__antithesis_instrumentation__.Notify(1947)

			return
		default:
			__antithesis_instrumentation__.Notify(1948)
		}
		__antithesis_instrumentation__.Notify(1945)
		args = append(args, "--"+f.Name+"="+vs)
	})
	__antithesis_instrumentation__.Notify(1943)
	args = append(args, "--test.bench=^$")
	args = append(args, flag.CommandLine.Args()...)
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, util.DisableMetamorphicEnvVar+"=t")
	t.Log(cmd.Args)
	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(1951)
		t.FailNow()
	} else {
		__antithesis_instrumentation__.Notify(1952)
	}
}

type resultSet struct {
	mu struct {
		syncutil.Mutex
		results []benchmarkResult
	}
}

func (s *resultSet) add(result benchmarkResult) {
	__antithesis_instrumentation__.Notify(1953)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.results = append(s.mu.results, result)
}

func (s *resultSet) iterate(f func(res benchmarkResult)) {
	__antithesis_instrumentation__.Notify(1954)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, res := range s.mu.results {
		__antithesis_instrumentation__.Notify(1955)
		f(res)
	}
}

func (s *resultSet) toSlice() (res []benchmarkResult) {
	__antithesis_instrumentation__.Notify(1956)
	s.iterate(func(result benchmarkResult) {
		__antithesis_instrumentation__.Notify(1958)
		res = append(res, result)
	})
	__antithesis_instrumentation__.Notify(1957)
	return res
}

func resultsToExpectations(t *testing.T, results []benchmarkResult) benchmarkExpectations {
	__antithesis_instrumentation__.Notify(1959)
	sort.Slice(results, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(1964)
		return results[i].name < results[j].name
	})
	__antithesis_instrumentation__.Notify(1960)
	var res benchmarkExpectations
	var cur benchmarkExpectation
	for _, result := range results {
		__antithesis_instrumentation__.Notify(1965)
		if result.name != cur.name {
			__antithesis_instrumentation__.Notify(1968)
			if cur != (benchmarkExpectation{}) {
				__antithesis_instrumentation__.Notify(1970)
				res = append(res, cur)
				cur = benchmarkExpectation{}
			} else {
				__antithesis_instrumentation__.Notify(1971)
			}
			__antithesis_instrumentation__.Notify(1969)
			cur = benchmarkExpectation{
				name: result.name,
				min:  result.result,
				max:  result.result,
			}
		} else {
			__antithesis_instrumentation__.Notify(1972)
		}
		__antithesis_instrumentation__.Notify(1966)
		if result.result < cur.min {
			__antithesis_instrumentation__.Notify(1973)
			cur.min = result.result
		} else {
			__antithesis_instrumentation__.Notify(1974)
		}
		__antithesis_instrumentation__.Notify(1967)
		if result.result > cur.max {
			__antithesis_instrumentation__.Notify(1975)
			cur.max = result.result
		} else {
			__antithesis_instrumentation__.Notify(1976)
		}
	}
	__antithesis_instrumentation__.Notify(1961)
	if cur != (benchmarkExpectation{}) {
		__antithesis_instrumentation__.Notify(1977)
		res = append(res, cur)
	} else {
		__antithesis_instrumentation__.Notify(1978)
	}
	__antithesis_instrumentation__.Notify(1962)

	for i := 1; i < len(res); i++ {
		__antithesis_instrumentation__.Notify(1979)
		if res[i-1].name == res[i].name {
			__antithesis_instrumentation__.Notify(1980)
			t.Fatalf("duplicate expectations for Name %s", res[i].name)
		} else {
			__antithesis_instrumentation__.Notify(1981)
		}
	}
	__antithesis_instrumentation__.Notify(1963)
	return res
}

func writeExpectationsFile(t *testing.T, expectations benchmarkExpectations) {
	__antithesis_instrumentation__.Notify(1982)
	f, err := os.Create(testutils.TestDataPath(t, expectationsFilename))
	require.NoError(t, err)
	defer func() { __antithesis_instrumentation__.Notify(1985); require.NoError(t, f.Close()) }()
	__antithesis_instrumentation__.Notify(1983)
	w := csv.NewWriter(f)
	w.Comma = ','
	require.NoError(t, w.Write(expectationsHeader))
	for _, exp := range expectations {
		__antithesis_instrumentation__.Notify(1986)
		require.NoError(t, w.Write([]string{exp.String(), exp.name}))
	}
	__antithesis_instrumentation__.Notify(1984)
	w.Flush()
	require.NoError(t, w.Error())
}

func readExpectationsFile(t testing.TB) benchmarkExpectations {
	__antithesis_instrumentation__.Notify(1987)
	f, err := os.Open(testutils.TestDataPath(t, expectationsFilename))
	require.NoError(t, err)
	defer func() { __antithesis_instrumentation__.Notify(1991); _ = f.Close() }()
	__antithesis_instrumentation__.Notify(1988)

	r := csv.NewReader(f)
	r.Comma = ','
	records, err := r.ReadAll()
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(records), 1, "must have at least a header")
	require.Equal(t, expectationsHeader, records[0])
	records = records[1:]
	ret := make(benchmarkExpectations, len(records))

	parseExp := func(expStr string) (min, max int, err error) {
		__antithesis_instrumentation__.Notify(1992)
		split := strings.Split(expStr, "-")
		if len(split) > 2 {
			__antithesis_instrumentation__.Notify(1996)
			return 0, 0, errors.Errorf("expected <min>-<max>, got %q", expStr)
		} else {
			__antithesis_instrumentation__.Notify(1997)
		}
		__antithesis_instrumentation__.Notify(1993)
		min, err = strconv.Atoi(split[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(1998)
			return 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(1999)
		}
		__antithesis_instrumentation__.Notify(1994)
		if len(split) == 1 {
			__antithesis_instrumentation__.Notify(2000)
			max = min
			return min, max, err
		} else {
			__antithesis_instrumentation__.Notify(2001)
		}
		__antithesis_instrumentation__.Notify(1995)
		max, err = strconv.Atoi(split[1])
		return min, max, err
	}
	__antithesis_instrumentation__.Notify(1989)
	for i, r := range records {
		__antithesis_instrumentation__.Notify(2002)
		min, max, err := parseExp(r[0])
		require.NoErrorf(t, err, "line %d", i+1)
		ret[i] = benchmarkExpectation{min: min, max: max, name: r[1]}
	}
	__antithesis_instrumentation__.Notify(1990)
	sort.Sort(ret)
	return ret
}

func (b benchmarkExpectations) find(name string) (benchmarkExpectation, bool) {
	__antithesis_instrumentation__.Notify(2003)
	idx := sort.Search(len(b), func(i int) bool {
		__antithesis_instrumentation__.Notify(2006)
		return b[i].name >= name
	})
	__antithesis_instrumentation__.Notify(2004)
	if idx < len(b) && func() bool {
		__antithesis_instrumentation__.Notify(2007)
		return b[idx].name == name == true
	}() == true {
		__antithesis_instrumentation__.Notify(2008)
		return b[idx], true
	} else {
		__antithesis_instrumentation__.Notify(2009)
	}
	__antithesis_instrumentation__.Notify(2005)
	return benchmarkExpectation{}, false
}

func (e benchmarkExpectation) matches(roundTrips int) bool {
	__antithesis_instrumentation__.Notify(2010)

	return (e.min <= roundTrips && func() bool {
		__antithesis_instrumentation__.Notify(2011)
		return roundTrips <= e.max == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(2012)
		return (e.min == e.max && func() bool {
			__antithesis_instrumentation__.Notify(2013)
			return (roundTrips == e.min-1 || func() bool {
				__antithesis_instrumentation__.Notify(2014)
				return roundTrips == e.min+1 == true
			}() == true) == true
		}() == true) == true
	}() == true
}

func (e benchmarkExpectation) String() string {
	__antithesis_instrumentation__.Notify(2015)
	expStr := strconv.Itoa(e.min)
	if e.min != e.max {
		__antithesis_instrumentation__.Notify(2017)
		expStr += "-"
		expStr += strconv.Itoa(e.max)
	} else {
		__antithesis_instrumentation__.Notify(2018)
	}
	__antithesis_instrumentation__.Notify(2016)
	return expStr
}

type benchmarkExpectations []benchmarkExpectation

var _ sort.Interface = (benchmarkExpectations)(nil)

func (b benchmarkExpectations) Len() int { __antithesis_instrumentation__.Notify(2019); return len(b) }
func (b benchmarkExpectations) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(2020)
	return b[i].name < b[j].name
}
func (b benchmarkExpectations) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(2021)
	b[i], b[j] = b[j], b[i]
}
