package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/workload/querybench"
)

type tpchBenchSpec struct {
	Nodes           int
	CPUs            int
	ScaleFactor     int
	benchType       string
	url             string
	numRunsPerQuery int

	minVersion string

	maxLatency time.Duration
}

func runTPCHBench(ctx context.Context, t test.Test, c cluster.Cluster, b tpchBenchSpec) {
	__antithesis_instrumentation__.Notify(51929)
	roachNodes := c.Range(1, c.Spec().NodeCount-1)
	loadNode := c.Node(c.Spec().NodeCount)

	t.Status("copying binaries")
	c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)

	filename := b.benchType
	t.Status(fmt.Sprintf("downloading %s query file from %s", filename, b.url))
	if err := c.RunE(ctx, loadNode, fmt.Sprintf("curl %s > %s", b.url, filename)); err != nil {
		__antithesis_instrumentation__.Notify(51932)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51933)
	}
	__antithesis_instrumentation__.Notify(51930)

	t.Status("starting nodes")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)

	m := c.NewMonitor(ctx, roachNodes)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(51934)
		t.Status("setting up dataset")
		err := loadTPCHDataset(ctx, t, c, b.ScaleFactor, m, roachNodes)
		if err != nil {
			__antithesis_instrumentation__.Notify(51938)
			return err
		} else {
			__antithesis_instrumentation__.Notify(51939)
		}
		__antithesis_instrumentation__.Notify(51935)

		t.L().Printf("running %s benchmark on tpch scale-factor=%d", filename, b.ScaleFactor)

		numQueries, err := getNumQueriesInFile(filename, b.url)
		if err != nil {
			__antithesis_instrumentation__.Notify(51940)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51941)
		}
		__antithesis_instrumentation__.Notify(51936)

		maxOps := b.numRunsPerQuery * numQueries

		cmd := fmt.Sprintf(
			"./workload run querybench --db=tpch --concurrency=1 --query-file=%s "+
				"--num-runs=%d --max-ops=%d {pgurl%s} "+
				"--histograms="+t.PerfArtifactsDir()+"/stats.json --histograms-max-latency=%s",
			filename,
			b.numRunsPerQuery,
			maxOps,
			roachNodes,
			b.maxLatency.String(),
		)
		if err := c.RunE(ctx, loadNode, cmd); err != nil {
			__antithesis_instrumentation__.Notify(51942)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51943)
		}
		__antithesis_instrumentation__.Notify(51937)
		return nil
	})
	__antithesis_instrumentation__.Notify(51931)
	m.Wait()
}

func getNumQueriesInFile(filename, url string) (int, error) {
	__antithesis_instrumentation__.Notify(51944)
	tempFile, err := downloadFile(filename, url)
	if err != nil {
		__antithesis_instrumentation__.Notify(51948)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(51949)
	}
	__antithesis_instrumentation__.Notify(51945)

	defer func() {
		__antithesis_instrumentation__.Notify(51950)
		_ = os.Remove(tempFile.Name())
	}()
	__antithesis_instrumentation__.Notify(51946)

	queries, err := querybench.GetQueries(tempFile.Name())
	if err != nil {
		__antithesis_instrumentation__.Notify(51951)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(51952)
	}
	__antithesis_instrumentation__.Notify(51947)
	return len(queries), nil
}

func downloadFile(filename string, url string) (*os.File, error) {
	__antithesis_instrumentation__.Notify(51953)

	httpClient := httputil.NewClientWithTimeout(30 * time.Second)

	resp, err := httpClient.Get(context.TODO(), url)
	if err != nil {
		__antithesis_instrumentation__.Notify(51956)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(51957)
	}
	__antithesis_instrumentation__.Notify(51954)
	defer resp.Body.Close()

	out, err := ioutil.TempFile(``, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(51958)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(51959)
	}
	__antithesis_instrumentation__.Notify(51955)
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return out, err
}

func registerTPCHBenchSpec(r registry.Registry, b tpchBenchSpec) {
	__antithesis_instrumentation__.Notify(51960)
	nameParts := []string{
		"tpchbench",
		b.benchType,
		fmt.Sprintf("nodes=%d", b.Nodes),
		fmt.Sprintf("cpu=%d", b.CPUs),
		fmt.Sprintf("sf=%d", b.ScaleFactor),
	}

	numNodes := b.Nodes + 1
	minVersion := b.minVersion
	if minVersion == `` {
		__antithesis_instrumentation__.Notify(51962)
		minVersion = "v19.1.0"
	} else {
		__antithesis_instrumentation__.Notify(51963)
	}
	__antithesis_instrumentation__.Notify(51961)

	r.Add(registry.TestSpec{
		Name:    strings.Join(nameParts, "/"),
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51964)
			runTPCHBench(ctx, t, c, b)
		},
	})
}

func registerTPCHBench(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51965)
	specs := []tpchBenchSpec{
		{
			Nodes:           3,
			CPUs:            4,
			ScaleFactor:     1,
			benchType:       `sql20`,
			url:             `https://raw.githubusercontent.com/cockroachdb/cockroach/master/pkg/workload/querybench/2.1-sql-20`,
			numRunsPerQuery: 3,
			maxLatency:      100 * time.Second,
		},
		{
			Nodes:           3,
			CPUs:            4,
			ScaleFactor:     1,
			benchType:       `tpch`,
			url:             `https://raw.githubusercontent.com/cockroachdb/cockroach/master/pkg/workload/querybench/tpch-queries`,
			numRunsPerQuery: 3,
			minVersion:      `v19.2.0`,
			maxLatency:      500 * time.Second,
		},
	}

	for _, b := range specs {
		__antithesis_instrumentation__.Notify(51966)
		registerTPCHBenchSpec(r, b)
	}
}
