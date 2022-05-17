package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/binfetcher"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/workload/tpch"
	"github.com/stretchr/testify/require"
)

const tpchVecPerfSlownessThreshold = 1.5

var tpchTables = []string{
	"nation", "region", "part", "supplier",
	"partsupp", "customer", "orders", "lineitem",
}

type tpchVecTestRunConfig struct {
	numRunsPerQuery int

	queriesToRun []int

	clusterSetups [][]string

	setupNames []string
}

func performClusterSetup(t test.Test, conn *gosql.DB, clusterSetup []string) {
	__antithesis_instrumentation__.Notify(51967)
	for _, query := range clusterSetup {
		__antithesis_instrumentation__.Notify(51968)
		if _, err := conn.Exec(query); err != nil {
			__antithesis_instrumentation__.Notify(51969)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51970)
		}
	}
}

type tpchVecTestCase interface {
	getRunConfig() tpchVecTestRunConfig

	preTestRunHook(ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, clusterSetup []string)

	postQueryRunHook(t test.Test, output []byte, setupIdx int)

	postTestRunHook(ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB)
}

type tpchVecTestCaseBase struct{}

func (b tpchVecTestCaseBase) getRunConfig() tpchVecTestRunConfig {
	__antithesis_instrumentation__.Notify(51971)
	runConfig := tpchVecTestRunConfig{
		numRunsPerQuery: 1,
		clusterSetups: [][]string{{
			"RESET CLUSTER SETTING sql.distsql.temp_storage.workmem",
			"SET CLUSTER SETTING sql.defaults.vectorize=on",
		}},
		setupNames: []string{"default"},
	}
	for queryNum := 1; queryNum <= tpch.NumQueries; queryNum++ {
		__antithesis_instrumentation__.Notify(51973)
		runConfig.queriesToRun = append(runConfig.queriesToRun, queryNum)
	}
	__antithesis_instrumentation__.Notify(51972)
	return runConfig
}

func (b tpchVecTestCaseBase) preTestRunHook(
	t test.Test, conn *gosql.DB, clusterSetup []string, createStats bool,
) {
	__antithesis_instrumentation__.Notify(51974)
	performClusterSetup(t, conn, clusterSetup)
	if createStats {
		__antithesis_instrumentation__.Notify(51975)
		createStatsFromTables(t, conn, tpchTables)
	} else {
		__antithesis_instrumentation__.Notify(51976)
	}
}

func (b tpchVecTestCaseBase) postQueryRunHook(test.Test, []byte, int) {
	__antithesis_instrumentation__.Notify(51977)
}

func (b tpchVecTestCaseBase) postTestRunHook(
	context.Context, test.Test, cluster.Cluster, *gosql.DB,
) {
	__antithesis_instrumentation__.Notify(51978)
}

type tpchVecPerfHelper struct {
	timeByQueryNum []map[int][]float64
}

func newTpchVecPerfHelper(numSetups int) *tpchVecPerfHelper {
	__antithesis_instrumentation__.Notify(51979)
	timeByQueryNum := make([]map[int][]float64, numSetups)
	for i := range timeByQueryNum {
		__antithesis_instrumentation__.Notify(51981)
		timeByQueryNum[i] = make(map[int][]float64)
	}
	__antithesis_instrumentation__.Notify(51980)
	return &tpchVecPerfHelper{
		timeByQueryNum: timeByQueryNum,
	}
}

func (h *tpchVecPerfHelper) parseQueryOutput(t test.Test, output []byte, setupIdx int) {
	__antithesis_instrumentation__.Notify(51982)
	runtimeRegex := regexp.MustCompile(`.*\[q([\d]+)\] returned \d+ rows after ([\d]+\.[\d]+) seconds.*`)
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(51983)
		line := scanner.Bytes()
		match := runtimeRegex.FindSubmatch(line)
		if match != nil {
			__antithesis_instrumentation__.Notify(51984)
			queryNum, err := strconv.Atoi(string(match[1]))
			if err != nil {
				__antithesis_instrumentation__.Notify(51987)
				t.Fatalf("failed parsing %q as int with %s", match[1], err)
			} else {
				__antithesis_instrumentation__.Notify(51988)
			}
			__antithesis_instrumentation__.Notify(51985)
			queryTime, err := strconv.ParseFloat(string(match[2]), 64)
			if err != nil {
				__antithesis_instrumentation__.Notify(51989)
				t.Fatalf("failed parsing %q as float with %s", match[2], err)
			} else {
				__antithesis_instrumentation__.Notify(51990)
			}
			__antithesis_instrumentation__.Notify(51986)
			h.timeByQueryNum[setupIdx][queryNum] = append(h.timeByQueryNum[setupIdx][queryNum], queryTime)
		} else {
			__antithesis_instrumentation__.Notify(51991)
		}
	}
}

const (
	tpchPerfTestVecOnConfigIdx  = 1
	tpchPerfTestVecOffConfigIdx = 0
)

type tpchVecPerfTest struct {
	tpchVecTestCaseBase
	*tpchVecPerfHelper

	disableStatsCreation bool
}

var _ tpchVecTestCase = &tpchVecPerfTest{}

func newTpchVecPerfTest(disableStatsCreation bool) *tpchVecPerfTest {
	__antithesis_instrumentation__.Notify(51992)
	return &tpchVecPerfTest{
		tpchVecPerfHelper:    newTpchVecPerfHelper(2),
		disableStatsCreation: disableStatsCreation,
	}
}

func (p tpchVecPerfTest) getRunConfig() tpchVecTestRunConfig {
	__antithesis_instrumentation__.Notify(51993)
	runConfig := p.tpchVecTestCaseBase.getRunConfig()
	if p.disableStatsCreation {
		__antithesis_instrumentation__.Notify(51995)

		runConfig.queriesToRun = append(runConfig.queriesToRun[:8], runConfig.queriesToRun[9:]...)
	} else {
		__antithesis_instrumentation__.Notify(51996)
	}
	__antithesis_instrumentation__.Notify(51994)
	runConfig.numRunsPerQuery = 3

	defaultSetup := runConfig.clusterSetups[0]
	runConfig.clusterSetups = append(runConfig.clusterSetups, make([]string, len(defaultSetup)))
	copy(runConfig.clusterSetups[1], defaultSetup)
	runConfig.clusterSetups[tpchPerfTestVecOffConfigIdx] = append(runConfig.clusterSetups[tpchPerfTestVecOffConfigIdx],
		"SET CLUSTER SETTING sql.defaults.vectorize=off")
	runConfig.clusterSetups[tpchPerfTestVecOnConfigIdx] = append(runConfig.clusterSetups[tpchPerfTestVecOnConfigIdx],
		"SET CLUSTER SETTING sql.defaults.vectorize=on")
	runConfig.setupNames = make([]string, 2)
	runConfig.setupNames[tpchPerfTestVecOffConfigIdx] = "off"
	runConfig.setupNames[tpchPerfTestVecOnConfigIdx] = "on"
	return runConfig
}

func (p tpchVecPerfTest) preTestRunHook(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, clusterSetup []string,
) {
	__antithesis_instrumentation__.Notify(51997)
	p.tpchVecTestCaseBase.preTestRunHook(t, conn, clusterSetup, !p.disableStatsCreation)
}

func (p *tpchVecPerfTest) postQueryRunHook(t test.Test, output []byte, setupIdx int) {
	__antithesis_instrumentation__.Notify(51998)
	p.parseQueryOutput(t, output, setupIdx)
}

func (p *tpchVecPerfTest) postTestRunHook(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB,
) {
	__antithesis_instrumentation__.Notify(51999)
	runConfig := p.getRunConfig()
	t.Status("comparing the runtimes (only median values for each query are compared)")
	for _, queryNum := range runConfig.queriesToRun {
		__antithesis_instrumentation__.Notify(52000)
		findMedian := func(times []float64) float64 {
			__antithesis_instrumentation__.Notify(52005)
			sort.Float64s(times)
			return times[len(times)/2]
		}
		__antithesis_instrumentation__.Notify(52001)
		vecOnTimes := p.timeByQueryNum[tpchPerfTestVecOnConfigIdx][queryNum]
		vecOffTimes := p.timeByQueryNum[tpchPerfTestVecOffConfigIdx][queryNum]
		if len(vecOnTimes) != runConfig.numRunsPerQuery {
			__antithesis_instrumentation__.Notify(52006)
			t.Fatal(fmt.Sprintf("[q%d] unexpectedly wrong number of run times "+
				"recorded with vec ON config: %v", queryNum, vecOnTimes))
		} else {
			__antithesis_instrumentation__.Notify(52007)
		}
		__antithesis_instrumentation__.Notify(52002)
		if len(vecOffTimes) != runConfig.numRunsPerQuery {
			__antithesis_instrumentation__.Notify(52008)
			t.Fatal(fmt.Sprintf("[q%d] unexpectedly wrong number of run times "+
				"recorded with vec OFF config: %v", queryNum, vecOffTimes))
		} else {
			__antithesis_instrumentation__.Notify(52009)
		}
		__antithesis_instrumentation__.Notify(52003)
		vecOnTime := findMedian(vecOnTimes)
		vecOffTime := findMedian(vecOffTimes)
		if vecOffTime < vecOnTime {
			__antithesis_instrumentation__.Notify(52010)
			t.L().Printf(
				fmt.Sprintf("[q%d] vec OFF was faster by %.2f%%: "+
					"%.2fs ON vs %.2fs OFF --- WARNING\n"+
					"vec ON times: %v\t vec OFF times: %v",
					queryNum, 100*(vecOnTime-vecOffTime)/vecOffTime,
					vecOnTime, vecOffTime, vecOnTimes, vecOffTimes))
		} else {
			__antithesis_instrumentation__.Notify(52011)
			t.L().Printf(
				fmt.Sprintf("[q%d] vec ON was faster by %.2f%%: "+
					"%.2fs ON vs %.2fs OFF\n"+
					"vec ON times: %v\t vec OFF times: %v",
					queryNum, 100*(vecOffTime-vecOnTime)/vecOnTime,
					vecOnTime, vecOffTime, vecOnTimes, vecOffTimes))
		}
		__antithesis_instrumentation__.Notify(52004)
		if vecOnTime >= tpchVecPerfSlownessThreshold*vecOffTime {
			__antithesis_instrumentation__.Notify(52012)

			for setupIdx, setup := range runConfig.clusterSetups {
				__antithesis_instrumentation__.Notify(52014)
				performClusterSetup(t, conn, setup)

				tempConn := c.Conn(ctx, t.L(), 1)
				defer tempConn.Close()
				if _, err := tempConn.Exec("USE tpch;"); err != nil {
					__antithesis_instrumentation__.Notify(52016)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(52017)
				}
				__antithesis_instrumentation__.Notify(52015)
				for i := 0; i < runConfig.numRunsPerQuery; i++ {
					__antithesis_instrumentation__.Notify(52018)
					t.Status(fmt.Sprintf("\nRunning EXPLAIN ANALYZE (DEBUG) for setup=%s\n", runConfig.setupNames[setupIdx]))
					rows, err := tempConn.Query(fmt.Sprintf(
						"EXPLAIN ANALYZE (DEBUG) %s;", tpch.QueriesByNumber[queryNum],
					))
					if err != nil {
						__antithesis_instrumentation__.Notify(52023)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(52024)
					}
					__antithesis_instrumentation__.Notify(52019)

					directLinkPrefix := "Direct link: "
					var line, url, debugOutput string
					for rows.Next() {
						__antithesis_instrumentation__.Notify(52025)
						if err = rows.Scan(&line); err != nil {
							__antithesis_instrumentation__.Notify(52027)
							t.Fatal(err)
						} else {
							__antithesis_instrumentation__.Notify(52028)
						}
						__antithesis_instrumentation__.Notify(52026)
						debugOutput += line + "\n"
						if strings.HasPrefix(line, directLinkPrefix) {
							__antithesis_instrumentation__.Notify(52029)
							url = line[len(directLinkPrefix):]
							break
						} else {
							__antithesis_instrumentation__.Notify(52030)
						}
					}
					__antithesis_instrumentation__.Notify(52020)
					if err = rows.Close(); err != nil {
						__antithesis_instrumentation__.Notify(52031)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(52032)
					}
					__antithesis_instrumentation__.Notify(52021)
					if url == "" {
						__antithesis_instrumentation__.Notify(52033)
						t.Fatal(fmt.Sprintf("unexpectedly didn't find a line "+
							"with %q prefix in EXPLAIN ANALYZE (DEBUG) output\n%s",
							directLinkPrefix, debugOutput))
					} else {
						__antithesis_instrumentation__.Notify(52034)
					}
					__antithesis_instrumentation__.Notify(52022)

					curlCmd := fmt.Sprintf(
						"curl %s > logs/bundle_%s_%d.zip", url, runConfig.setupNames[setupIdx], i,
					)
					if err = c.RunE(ctx, c.Node(1), curlCmd); err != nil {
						__antithesis_instrumentation__.Notify(52035)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(52036)
					}
				}
			}
			__antithesis_instrumentation__.Notify(52013)
			t.Fatal(fmt.Sprintf(
				"[q%d] vec ON is slower by %.2f%% than vec OFF\n"+
					"vec ON times: %v\nvec OFF times: %v",
				queryNum, 100*(vecOnTime-vecOffTime)/vecOffTime, vecOnTimes, vecOffTimes))
		} else {
			__antithesis_instrumentation__.Notify(52037)
		}
	}
}

type tpchVecBenchTest struct {
	tpchVecTestCaseBase
	*tpchVecPerfHelper

	numRunsPerQuery int
	queriesToRun    []int
	clusterSetups   [][]string
	setupNames      []string
}

var _ tpchVecTestCase = &tpchVecBenchTest{}

func newTpchVecBenchTest(
	numRunsPerQuery int, queriesToRun []int, clusterSetups [][]string, setupNames []string,
) *tpchVecBenchTest {
	__antithesis_instrumentation__.Notify(52038)
	return &tpchVecBenchTest{
		tpchVecPerfHelper: newTpchVecPerfHelper(len(setupNames)),
		numRunsPerQuery:   numRunsPerQuery,
		queriesToRun:      queriesToRun,
		clusterSetups:     clusterSetups,
		setupNames:        setupNames,
	}
}

func (b tpchVecBenchTest) getRunConfig() tpchVecTestRunConfig {
	__antithesis_instrumentation__.Notify(52039)
	runConfig := b.tpchVecTestCaseBase.getRunConfig()
	runConfig.numRunsPerQuery = b.numRunsPerQuery
	if b.queriesToRun != nil {
		__antithesis_instrumentation__.Notify(52042)
		runConfig.queriesToRun = b.queriesToRun
	} else {
		__antithesis_instrumentation__.Notify(52043)
	}
	__antithesis_instrumentation__.Notify(52040)
	defaultSetup := runConfig.clusterSetups[0]

	defaultSetup = defaultSetup[:len(defaultSetup):len(defaultSetup)]
	runConfig.clusterSetups = make([][]string, len(b.clusterSetups))
	runConfig.setupNames = b.setupNames
	for setupIdx, configSetup := range b.clusterSetups {
		__antithesis_instrumentation__.Notify(52044)
		runConfig.clusterSetups[setupIdx] = append(defaultSetup, configSetup...)
	}
	__antithesis_instrumentation__.Notify(52041)
	return runConfig
}

func (b tpchVecBenchTest) preTestRunHook(
	_ context.Context, t test.Test, _ cluster.Cluster, conn *gosql.DB, clusterSetup []string,
) {
	__antithesis_instrumentation__.Notify(52045)
	b.tpchVecTestCaseBase.preTestRunHook(t, conn, clusterSetup, true)
}

func (b *tpchVecBenchTest) postQueryRunHook(t test.Test, output []byte, setupIdx int) {
	__antithesis_instrumentation__.Notify(52046)
	b.tpchVecPerfHelper.parseQueryOutput(t, output, setupIdx)
}

func (b *tpchVecBenchTest) postTestRunHook(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB,
) {
	__antithesis_instrumentation__.Notify(52047)
	runConfig := b.getRunConfig()
	t.Status("comparing the runtimes (average of values (excluding best and worst) for each query are compared)")

	scores := make([]float64, len(runConfig.setupNames))
	for _, queryNum := range runConfig.queriesToRun {
		__antithesis_instrumentation__.Notify(52050)

		findAvgTime := func(times []float64) float64 {
			__antithesis_instrumentation__.Notify(52053)
			if len(times) < 3 {
				__antithesis_instrumentation__.Notify(52056)
				t.Fatal(fmt.Sprintf("unexpectedly query %d ran %d times on one of the setups", queryNum, len(times)))
			} else {
				__antithesis_instrumentation__.Notify(52057)
			}
			__antithesis_instrumentation__.Notify(52054)
			sort.Float64s(times)
			sum, count := 0.0, 0
			for _, time := range times[1 : len(times)-1] {
				__antithesis_instrumentation__.Notify(52058)
				sum += time
				count++
			}
			__antithesis_instrumentation__.Notify(52055)
			return sum / float64(count)
		}
		__antithesis_instrumentation__.Notify(52051)
		bestTime := math.MaxFloat64
		var bestSetupIdx int
		for setupIdx := range runConfig.setupNames {
			__antithesis_instrumentation__.Notify(52059)
			setupTime := findAvgTime(b.timeByQueryNum[setupIdx][queryNum])
			if setupTime < bestTime {
				__antithesis_instrumentation__.Notify(52060)
				bestTime = setupTime
				bestSetupIdx = setupIdx
			} else {
				__antithesis_instrumentation__.Notify(52061)
			}
		}
		__antithesis_instrumentation__.Notify(52052)
		t.L().Printf(fmt.Sprintf("[q%d] best setup is %s", queryNum, runConfig.setupNames[bestSetupIdx]))
		for setupIdx, setupName := range runConfig.setupNames {
			__antithesis_instrumentation__.Notify(52062)
			setupTime := findAvgTime(b.timeByQueryNum[setupIdx][queryNum])
			scores[setupIdx] += setupTime / bestTime
			t.L().Printf(fmt.Sprintf("[q%d] setup %s took %.2fs", queryNum, setupName, setupTime))
		}
	}
	__antithesis_instrumentation__.Notify(52048)
	t.Status("----- scores of the setups -----")
	bestScore := math.MaxFloat64
	var bestSetupIdx int
	for setupIdx, setupName := range runConfig.setupNames {
		__antithesis_instrumentation__.Notify(52063)
		score := scores[setupIdx]
		t.L().Printf(fmt.Sprintf("score of %s is %.2f", setupName, score))
		if bestScore > score {
			__antithesis_instrumentation__.Notify(52064)
			bestScore = score
			bestSetupIdx = setupIdx
		} else {
			__antithesis_instrumentation__.Notify(52065)
		}
	}
	__antithesis_instrumentation__.Notify(52049)
	t.Status(fmt.Sprintf("----- best setup is %s -----", runConfig.setupNames[bestSetupIdx]))
}

type tpchVecDiskTest struct {
	tpchVecTestCaseBase
}

func (d tpchVecDiskTest) preTestRunHook(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, clusterSetup []string,
) {
	__antithesis_instrumentation__.Notify(52066)
	d.tpchVecTestCaseBase.preTestRunHook(t, conn, clusterSetup, true)

	rng, _ := randutil.NewTestRand()
	workmemInKiB := 650 + rng.Intn(1350)
	workmem := fmt.Sprintf("%dKiB", workmemInKiB)
	t.Status(fmt.Sprintf("setting workmem='%s'", workmem))
	if _, err := conn.Exec(fmt.Sprintf("SET CLUSTER SETTING sql.distsql.temp_storage.workmem='%s'", workmem)); err != nil {
		__antithesis_instrumentation__.Notify(52067)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52068)
	}
}

func baseTestRun(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, tc tpchVecTestCase,
) {
	__antithesis_instrumentation__.Notify(52069)
	firstNode := c.Node(1)
	runConfig := tc.getRunConfig()
	for setupIdx, setup := range runConfig.clusterSetups {
		__antithesis_instrumentation__.Notify(52070)
		t.Status(fmt.Sprintf("running setup=%s", runConfig.setupNames[setupIdx]))
		tc.preTestRunHook(ctx, t, c, conn, setup)
		for _, queryNum := range runConfig.queriesToRun {
			__antithesis_instrumentation__.Notify(52071)

			cmd := fmt.Sprintf("./workload run tpch --concurrency=1 --db=tpch "+
				"--default-vectorize --max-ops=%d --queries=%d {pgurl:1} --enable-checks=true",
				runConfig.numRunsPerQuery, queryNum)
			result, err := c.RunWithDetailsSingleNode(ctx, t.L(), firstNode, cmd)
			workloadOutput := result.Stdout + result.Stderr
			t.L().Printf(workloadOutput)
			if err != nil {
				__antithesis_instrumentation__.Notify(52073)

				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52074)
			}
			__antithesis_instrumentation__.Notify(52072)
			tc.postQueryRunHook(t, []byte(workloadOutput), setupIdx)
		}
	}
}

type tpchVecSmithcmpTest struct {
	tpchVecTestCaseBase
}

const tpchVecSmithcmp = "smithcmp"

func (s tpchVecSmithcmpTest) preTestRunHook(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, clusterSetup []string,
) {
	__antithesis_instrumentation__.Notify(52075)
	s.tpchVecTestCaseBase.preTestRunHook(t, conn, clusterSetup, true)
	const smithcmpSHA = "a3f41f5ba9273249c5ecfa6348ea8ee3ac4b77e3"
	node := c.Node(1)
	if c.IsLocal() && func() bool {
		__antithesis_instrumentation__.Notify(52078)
		return runtime.GOOS != "linux" == true
	}() == true {
		__antithesis_instrumentation__.Notify(52079)
		t.Fatalf("must run on linux os, found %s", runtime.GOOS)
	} else {
		__antithesis_instrumentation__.Notify(52080)
	}
	__antithesis_instrumentation__.Notify(52076)

	smithcmp, err := binfetcher.Download(ctx, binfetcher.Options{
		Component: tpchVecSmithcmp,
		Binary:    tpchVecSmithcmp,
		Version:   smithcmpSHA,
		GOOS:      "linux",
		GOARCH:    "amd64",
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(52081)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52082)
	}
	__antithesis_instrumentation__.Notify(52077)
	c.Put(ctx, smithcmp, "./"+tpchVecSmithcmp, node)
}

func smithcmpTestRun(
	ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, tc tpchVecTestCase,
) {
	__antithesis_instrumentation__.Notify(52083)
	runConfig := tc.getRunConfig()
	tc.preTestRunHook(ctx, t, c, conn, runConfig.clusterSetups[0])
	const (
		configFile = `tpchvec_smithcmp.toml`
		configURL  = `https://raw.githubusercontent.com/cockroachdb/cockroach/master/pkg/cmd/roachtest/tests/` + configFile
	)
	firstNode := c.Node(1)
	if err := c.RunE(ctx, firstNode, fmt.Sprintf("curl %s > %s", configURL, configFile)); err != nil {
		__antithesis_instrumentation__.Notify(52085)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52086)
	}
	__antithesis_instrumentation__.Notify(52084)
	cmd := fmt.Sprintf("./%s %s", tpchVecSmithcmp, configFile)
	if err := c.RunE(ctx, firstNode, cmd); err != nil {
		__antithesis_instrumentation__.Notify(52087)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52088)
	}
}

func runTPCHVec(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	testCase tpchVecTestCase,
	testRun func(ctx context.Context, t test.Test, c cluster.Cluster, conn *gosql.DB, tc tpchVecTestCase),
) {
	__antithesis_instrumentation__.Notify(52089)
	firstNode := c.Node(1)
	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", firstNode)
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	conn := c.Conn(ctx, t.L(), 1)
	disableAutoStats(t, conn)
	t.Status("restoring TPCH dataset for Scale Factor 1")
	if err := loadTPCHDataset(ctx, t, c, 1, c.NewMonitor(ctx), c.All()); err != nil {
		__antithesis_instrumentation__.Notify(52092)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52093)
	}
	__antithesis_instrumentation__.Notify(52090)

	if _, err := conn.Exec("USE tpch;"); err != nil {
		__antithesis_instrumentation__.Notify(52094)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52095)
	}
	__antithesis_instrumentation__.Notify(52091)
	scatterTables(t, conn, tpchTables)
	t.Status("waiting for full replication")
	err := WaitFor3XReplication(ctx, t, conn)
	require.NoError(t, err)

	testRun(ctx, t, c, conn, testCase)
	testCase.postTestRunHook(ctx, t, c, conn)
}

const tpchVecNodeCount = 3

func registerTPCHVec(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52096)
	r.Add(registry.TestSpec{
		Name:    "tpchvec/perf",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(tpchVecNodeCount),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52101)
			runTPCHVec(ctx, t, c, newTpchVecPerfTest(false), baseTestRun)
		},
	})
	__antithesis_instrumentation__.Notify(52097)

	r.Add(registry.TestSpec{
		Name:    "tpchvec/disk",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(tpchVecNodeCount),

		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52102)
			runTPCHVec(ctx, t, c, tpchVecDiskTest{}, baseTestRun)
		},
	})
	__antithesis_instrumentation__.Notify(52098)

	r.Add(registry.TestSpec{
		Name:    "tpchvec/smithcmp",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(tpchVecNodeCount),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52103)
			runTPCHVec(ctx, t, c, tpchVecSmithcmpTest{}, smithcmpTestRun)
		},
	})
	__antithesis_instrumentation__.Notify(52099)

	r.Add(registry.TestSpec{
		Name:    "tpchvec/perf_no_stats",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(tpchVecNodeCount),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52104)
			runTPCHVec(ctx, t, c, newTpchVecPerfTest(true), baseTestRun)
		},
	})
	__antithesis_instrumentation__.Notify(52100)

	r.Add(registry.TestSpec{
		Name:    "tpchvec/bench",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(tpchVecNodeCount),
		Skip: "This config can be used to perform some benchmarking and is not " +
			"meant to be run on a nightly basis",
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52105)

			var clusterSetups [][]string
			var setupNames []string
			for _, batchSize := range []int{512, 1024, 1536} {
				__antithesis_instrumentation__.Notify(52107)
				clusterSetups = append(clusterSetups, []string{
					fmt.Sprintf("SET CLUSTER SETTING sql.testing.vectorize.batch_size=%d", batchSize),
				})
				setupNames = append(setupNames, fmt.Sprintf("%d", batchSize))
			}
			__antithesis_instrumentation__.Notify(52106)
			benchTest := newTpchVecBenchTest(
				5,
				nil,
				clusterSetups,
				setupNames,
			)
			runTPCHVec(ctx, t, c, benchTest, baseTestRun)
		},
	})
}
