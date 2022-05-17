package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/search"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/ttycolor"
	"github.com/codahale/hdrhistogram"
)

type kvBenchKeyDistribution int

const (
	sequential kvBenchKeyDistribution = iota
	random
	zipfian
)

type kvBenchSpec struct {
	Nodes int
	CPUs  int

	NumShards       int
	KeyDistribution kvBenchKeyDistribution
	SecondaryIndex  bool

	EstimatedMaxThroughput int
	LatencyThresholdMs     float64
}

func registerKVBenchSpec(r registry.Registry, b kvBenchSpec) {
	__antithesis_instrumentation__.Notify(49090)
	nameParts := []string{
		"kv0bench",
		fmt.Sprintf("nodes=%d", b.Nodes),
		fmt.Sprintf("cpu=%d", b.CPUs),
	}
	if b.NumShards > 0 {
		__antithesis_instrumentation__.Notify(49094)
		nameParts = append(nameParts, fmt.Sprintf("shards=%d", b.NumShards))
	} else {
		__antithesis_instrumentation__.Notify(49095)
	}
	__antithesis_instrumentation__.Notify(49091)
	opts := []spec.Option{spec.CPU(b.CPUs)}
	switch b.KeyDistribution {
	case sequential:
		__antithesis_instrumentation__.Notify(49096)
		nameParts = append(nameParts, "sequential")
	case random:
		__antithesis_instrumentation__.Notify(49097)
		nameParts = append(nameParts, "random")
	case zipfian:
		__antithesis_instrumentation__.Notify(49098)
		nameParts = append(nameParts, "zipfian")
	default:
		__antithesis_instrumentation__.Notify(49099)
		panic("unexpected")
	}
	__antithesis_instrumentation__.Notify(49092)

	if b.SecondaryIndex {
		__antithesis_instrumentation__.Notify(49100)
		nameParts = append(nameParts, "2nd_idx")
	} else {
		__antithesis_instrumentation__.Notify(49101)
	}
	__antithesis_instrumentation__.Notify(49093)

	name := strings.Join(nameParts, "/")
	nodes := r.MakeClusterSpec(b.Nodes+1, opts...)
	r.Add(registry.TestSpec{
		Name: name,

		Tags:    []string{"manual"},
		Owner:   registry.OwnerKV,
		Cluster: nodes,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49102)
			runKVBench(ctx, t, c, b)
		},
	})
}

func registerKVBench(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49103)
	specs := []kvBenchSpec{
		{
			Nodes:                  5,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 30000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              10,
		},
		{
			Nodes:                  5,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 20000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              0,
		},
		{
			Nodes:                  10,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 60000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              20,
		},
		{
			Nodes:                  10,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 20000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              0,
		},
		{
			Nodes:                  20,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 100000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              80,
		},
		{
			Nodes:                  20,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 20000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         false,
			NumShards:              0,
		},
		{
			Nodes:                  5,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 10000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         true,
			NumShards:              10,
		},
		{
			Nodes:                  5,
			CPUs:                   8,
			KeyDistribution:        sequential,
			EstimatedMaxThroughput: 5000,
			LatencyThresholdMs:     10.0,
			SecondaryIndex:         true,
			NumShards:              0,
		},
	}

	for _, b := range specs {
		__antithesis_instrumentation__.Notify(49104)
		registerKVBenchSpec(r, b)
	}
}

func makeKVLoadGroup(c cluster.Cluster, numRoachNodes, numLoadNodes int) loadGroup {
	__antithesis_instrumentation__.Notify(49105)
	return loadGroup{
		roachNodes: c.Range(1, numRoachNodes),
		loadNodes:  c.Range(numRoachNodes+1, numRoachNodes+numLoadNodes),
	}
}

func runKVBench(ctx context.Context, t test.Test, c cluster.Cluster, b kvBenchSpec) {
	__antithesis_instrumentation__.Notify(49106)
	loadGrp := makeKVLoadGroup(c, b.Nodes, 1)
	roachNodes := loadGrp.roachNodes
	loadNodes := loadGrp.loadNodes

	if err := c.PutE(ctx, t.L(), t.Cockroach(), "./cockroach", roachNodes); err != nil {
		__antithesis_instrumentation__.Notify(49111)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49112)
	}
	__antithesis_instrumentation__.Notify(49107)
	if err := c.PutE(ctx, t.L(), t.DeprecatedWorkload(), "./workload", loadNodes); err != nil {
		__antithesis_instrumentation__.Notify(49113)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49114)
	}
	__antithesis_instrumentation__.Notify(49108)

	const restartWait = 15 * time.Second

	precision := int(math.Max(1.0, float64(b.EstimatedMaxThroughput/50)))
	initStepSize := 2 * precision

	resultsDir, err := ioutil.TempDir("", "roachtest-kvbench")
	if err != nil {
		__antithesis_instrumentation__.Notify(49115)
		t.Fatal(errors.Wrapf(err, `failed to create temp results dir`))
	} else {
		__antithesis_instrumentation__.Notify(49116)
	}
	__antithesis_instrumentation__.Notify(49109)
	s := search.NewLineSearcher(100, 10000000, b.EstimatedMaxThroughput, initStepSize, precision)
	searchPredicate := func(maxrate int) (bool, error) {
		__antithesis_instrumentation__.Notify(49117)
		m := c.NewMonitor(ctx, roachNodes)

		m.ExpectDeaths(int32(len(roachNodes)))

		c.Wipe(ctx, roachNodes)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
		time.Sleep(restartWait)

		resultChan := make(chan *kvBenchResult, 1)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49121)
			db := c.Conn(ctx, t.L(), 1)

			var initCmd strings.Builder

			fmt.Fprintf(&initCmd, `./workload init kv --num-shards=%d`,
				b.NumShards)
			if b.SecondaryIndex {
				__antithesis_instrumentation__.Notify(49130)
				initCmd.WriteString(` --secondary-index`)
			} else {
				__antithesis_instrumentation__.Notify(49131)
			}
			__antithesis_instrumentation__.Notify(49122)
			fmt.Fprintf(&initCmd, ` {pgurl%s}`, roachNodes)
			if err := c.RunE(ctx, loadNodes, initCmd.String()); err != nil {
				__antithesis_instrumentation__.Notify(49132)
				return err
			} else {
				__antithesis_instrumentation__.Notify(49133)
			}
			__antithesis_instrumentation__.Notify(49123)

			var splitCmd strings.Builder
			if b.NumShards > 0 {
				__antithesis_instrumentation__.Notify(49134)
				splitCmd.WriteString(`USE kv; ALTER TABLE kv SPLIT AT VALUES `)

				for i := 0; i < b.NumShards; i++ {
					__antithesis_instrumentation__.Notify(49136)
					if i != 0 {
						__antithesis_instrumentation__.Notify(49138)
						splitCmd.WriteString(`,`)
					} else {
						__antithesis_instrumentation__.Notify(49139)
					}
					__antithesis_instrumentation__.Notify(49137)
					fmt.Fprintf(&splitCmd, `(%d)`, i)
				}
				__antithesis_instrumentation__.Notify(49135)
				splitCmd.WriteString(`;`)
			} else {
				__antithesis_instrumentation__.Notify(49140)
			}
			__antithesis_instrumentation__.Notify(49124)
			if _, err := db.Exec(splitCmd.String()); err != nil {
				__antithesis_instrumentation__.Notify(49141)
				t.L().Printf(splitCmd.String())
				return err
			} else {
				__antithesis_instrumentation__.Notify(49142)
			}
			__antithesis_instrumentation__.Notify(49125)

			workloadCmd := strings.Builder{}
			clusterHistPath := fmt.Sprintf("%s/kvbench/maxrate=%d/stats.json",
				t.PerfArtifactsDir(), maxrate)

			const loadConcurrency = 192
			const ramp = time.Second * 300
			const duration = time.Second * 300

			fmt.Fprintf(&workloadCmd,
				`./workload run kv --ramp=%fs --duration=%fs {pgurl%s} --read-percent=0`+
					` --concurrency=%d --histograms=%s --max-rate=%d --num-shards=%d`,
				ramp.Seconds(), duration.Seconds(), roachNodes,
				b.CPUs*loadConcurrency, clusterHistPath, maxrate, b.NumShards)
			switch b.KeyDistribution {
			case sequential:
				__antithesis_instrumentation__.Notify(49143)
				workloadCmd.WriteString(` --sequential`)
			case zipfian:
				__antithesis_instrumentation__.Notify(49144)
				workloadCmd.WriteString(` --zipfian`)
			case random:
				__antithesis_instrumentation__.Notify(49145)
				workloadCmd.WriteString(` --random`)
			default:
				__antithesis_instrumentation__.Notify(49146)
				panic(`unexpected`)
			}
			__antithesis_instrumentation__.Notify(49126)

			err := c.RunE(ctx, loadNodes, workloadCmd.String())
			if err != nil {
				__antithesis_instrumentation__.Notify(49147)
				return errors.Wrapf(err, `error running workload`)
			} else {
				__antithesis_instrumentation__.Notify(49148)
			}
			__antithesis_instrumentation__.Notify(49127)

			localHistPath := filepath.Join(resultsDir, fmt.Sprintf(`kvbench-%d-stats.json`, maxrate))
			if err := c.Get(ctx, t.L(), clusterHistPath, localHistPath, loadNodes); err != nil {
				__antithesis_instrumentation__.Notify(49149)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49150)
			}
			__antithesis_instrumentation__.Notify(49128)

			snapshots, err := histogram.DecodeSnapshots(localHistPath)
			if err != nil {
				__antithesis_instrumentation__.Notify(49151)
				return errors.Wrapf(err, `failed to decode histogram snapshots`)
			} else {
				__antithesis_instrumentation__.Notify(49152)
			}
			__antithesis_instrumentation__.Notify(49129)

			res := newResultFromSnapshots(maxrate, snapshots)
			resultChan <- res
			return nil
		})
		__antithesis_instrumentation__.Notify(49118)

		if err := m.WaitE(); err != nil {
			__antithesis_instrumentation__.Notify(49153)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(49154)
		}
		__antithesis_instrumentation__.Notify(49119)
		close(resultChan)
		res := <-resultChan

		var color ttycolor.Code
		var msg string
		pass := res.latency() <= b.LatencyThresholdMs
		if pass {
			__antithesis_instrumentation__.Notify(49155)
			color = ttycolor.Green
			msg = "PASS"
		} else {
			__antithesis_instrumentation__.Notify(49156)
			color = ttycolor.Red
			msg = "FAIL"
		}
		__antithesis_instrumentation__.Notify(49120)
		ttycolor.Stdout(color)
		t.L().Printf(`--- SEARCH ITER %s: kv workload avg latency: %0.1fms (threshold: %0.1fms), avg throughput: %d`,
			msg, res.latency(), b.LatencyThresholdMs, res.throughput())
		ttycolor.Stdout(ttycolor.Reset)
		return pass, nil
	}
	__antithesis_instrumentation__.Notify(49110)
	if res, err := s.Search(searchPredicate); err != nil {
		__antithesis_instrumentation__.Notify(49157)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49158)
		ttycolor.Stdout(ttycolor.Green)
		t.L().Printf("-------\nMAX THROUGHPUT = %d\n--------\n\n", res)
		ttycolor.Stdout(ttycolor.Reset)
	}
}

type kvBenchResult struct {
	Cumulative map[string]*hdrhistogram.Histogram
	Elapsed    time.Duration
}

func newResultFromSnapshots(
	maxrate int, snapshots map[string][]histogram.SnapshotTick,
) *kvBenchResult {
	__antithesis_instrumentation__.Notify(49159)
	var start, end time.Time
	ret := make(map[string]*hdrhistogram.Histogram, len(snapshots))
	for n, snaps := range snapshots {
		__antithesis_instrumentation__.Notify(49161)
		var cur *hdrhistogram.Histogram
		for _, s := range snaps {
			__antithesis_instrumentation__.Notify(49163)
			h := hdrhistogram.Import(s.Hist)
			if cur == nil {
				__antithesis_instrumentation__.Notify(49166)
				cur = h
			} else {
				__antithesis_instrumentation__.Notify(49167)
				cur.Merge(h)
			}
			__antithesis_instrumentation__.Notify(49164)
			if start.IsZero() || func() bool {
				__antithesis_instrumentation__.Notify(49168)
				return s.Now.Before(start) == true
			}() == true {
				__antithesis_instrumentation__.Notify(49169)
				start = s.Now
			} else {
				__antithesis_instrumentation__.Notify(49170)
			}
			__antithesis_instrumentation__.Notify(49165)
			if sEnd := s.Now.Add(s.Elapsed); end.IsZero() || func() bool {
				__antithesis_instrumentation__.Notify(49171)
				return sEnd.After(end) == true
			}() == true {
				__antithesis_instrumentation__.Notify(49172)
				end = sEnd
			} else {
				__antithesis_instrumentation__.Notify(49173)
			}
		}
		__antithesis_instrumentation__.Notify(49162)
		ret[n] = cur
	}
	__antithesis_instrumentation__.Notify(49160)
	return &kvBenchResult{
		Cumulative: ret,
		Elapsed:    end.Sub(start),
	}
}

func (r kvBenchResult) latency() float64 {
	__antithesis_instrumentation__.Notify(49174)
	return time.Duration(r.Cumulative[`write`].Mean()).Seconds() * 1000
}

func (r kvBenchResult) throughput() int {
	__antithesis_instrumentation__.Notify(49175)

	return int(float64(r.Cumulative[`write`].TotalCount()) / r.Elapsed.Seconds())
}
