package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

const tpccOlapQuery = `SELECT
    i_id, s_w_id, s_quantity, i_price
FROM
    stock JOIN item ON s_i_id = i_id
WHERE
    s_quantity < 100 AND i_price > 90
ORDER BY
    i_price DESC, s_quantity ASC
LIMIT
    100;`

type tpccOLAPSpec struct {
	Nodes       int
	CPUs        int
	Warehouses  int
	Concurrency int
}

func (s tpccOLAPSpec) run(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(49648)
	crdbNodes, workloadNode := setupTPCC(
		ctx, t, c, tpccOptions{
			Warehouses: s.Warehouses, SetupType: usingImport,
		})
	const queryFileName = "queries.sql"

	queryLine := `"` + strings.Replace(tpccOlapQuery, "\n", " ", -1) + `"`
	c.Run(ctx, workloadNode, "echo", queryLine, "> "+queryFileName)
	t.Status("waiting")
	m := c.NewMonitor(ctx, crdbNodes)
	rampDuration := 2 * time.Minute
	duration := 3 * time.Minute
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49650)
		t.WorkerStatus("running querybench")
		cmd := fmt.Sprintf(
			"./workload run querybench --db tpcc"+
				" --tolerate-errors=t"+
				" --concurrency=%d"+
				" --query-file %s"+
				" --histograms="+t.PerfArtifactsDir()+"/stats.json "+
				" --ramp=%s --duration=%s {pgurl:1-%d}",
			s.Concurrency, queryFileName, rampDuration, duration, c.Spec().NodeCount-1)
		c.Run(ctx, workloadNode, cmd)
		return nil
	})
	__antithesis_instrumentation__.Notify(49649)
	m.Wait()
	verifyNodeLiveness(ctx, c, t, duration)
}

func verifyNodeLiveness(
	ctx context.Context, c cluster.Cluster, t test.Test, runDuration time.Duration,
) {
	__antithesis_instrumentation__.Notify(49651)
	const maxFailures = 10
	adminURLs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(1))
	if err != nil {
		__antithesis_instrumentation__.Notify(49655)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49656)
	}
	__antithesis_instrumentation__.Notify(49652)
	now := timeutil.Now()
	var response tspb.TimeSeriesQueryResponse

	if err := retry.WithMaxAttempts(ctx, retry.Options{
		MaxBackoff: 500 * time.Millisecond,
	}, 60, func() (err error) {
		__antithesis_instrumentation__.Notify(49657)
		response, err = getMetrics(adminURLs[0], now.Add(-runDuration), now, []tsQuery{
			{
				name:      "cr.node.liveness.heartbeatfailures",
				queryType: total,
			},
		})
		return err
	}); err != nil {
		__antithesis_instrumentation__.Notify(49658)
		t.Fatalf("failed to fetch liveness metrics: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(49659)
	}
	__antithesis_instrumentation__.Notify(49653)
	if len(response.Results[0].Datapoints) <= 1 {
		__antithesis_instrumentation__.Notify(49660)
		t.Fatalf("not enough datapoints in timeseries query response: %+v", response)
	} else {
		__antithesis_instrumentation__.Notify(49661)
	}
	__antithesis_instrumentation__.Notify(49654)
	datapoints := response.Results[0].Datapoints
	finalCount := int(datapoints[len(datapoints)-1].Value)
	initialCount := int(datapoints[0].Value)
	if failures := finalCount - initialCount; failures > maxFailures {
		__antithesis_instrumentation__.Notify(49662)
		t.Fatalf("Node liveness failed %d times, expected no more than %d",
			failures, maxFailures)
	} else {
		__antithesis_instrumentation__.Notify(49663)
		t.L().Printf("Node liveness failed %d times which is fewer than %d",
			failures, maxFailures)
	}
}

func registerTPCCOverloadSpec(r registry.Registry, s tpccOLAPSpec) {
	__antithesis_instrumentation__.Notify(49664)
	name := fmt.Sprintf("overload/tpcc_olap/nodes=%d/cpu=%d/w=%d/c=%d",
		s.Nodes, s.CPUs, s.Warehouses, s.Concurrency)
	r.Add(registry.TestSpec{
		Name:            name,
		Owner:           registry.OwnerKV,
		Cluster:         r.MakeClusterSpec(s.Nodes+1, spec.CPU(s.CPUs)),
		Run:             s.run,
		EncryptAtRandom: true,
		Timeout:         20 * time.Minute,
	})
}

func registerOverload(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49665)
	specs := []tpccOLAPSpec{
		{
			CPUs:        8,
			Concurrency: 96,
			Nodes:       3,
			Warehouses:  50,
		},
	}
	for _, s := range specs {
		__antithesis_instrumentation__.Notify(49666)
		registerTPCCOverloadSpec(r, s)
	}
}
