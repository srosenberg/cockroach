package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

func registerLedger(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49176)
	const nodes = 6

	const azs = "us-central1-f,us-central1-b,us-central1-c"
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("ledger/nodes=%d/multi-az", nodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(nodes+1, spec.CPU(16), spec.Geo(), spec.Zones(azs)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49177)
			roachNodes := c.Range(1, nodes)
			gatewayNodes := c.Range(1, nodes/3)
			loadNode := c.Node(nodes + 1)

			c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)

			t.Status("running workload")
			m := c.NewMonitor(ctx, roachNodes)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(49179)
				concurrency := ifLocal(c, "", " --concurrency="+fmt.Sprint(nodes*32))
				duration := " --duration=" + ifLocal(c, "10s", "10m")

				cmd := fmt.Sprintf("./workload run ledger --init --histograms="+t.PerfArtifactsDir()+"/stats.json"+
					concurrency+duration+" {pgurl%s}", gatewayNodes)
				c.Run(ctx, loadNode, cmd)
				return nil
			})
			__antithesis_instrumentation__.Notify(49178)
			m.Wait()
		},
	})
}
