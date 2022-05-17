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

func registerRoachmart(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50407)
	runRoachmart := func(ctx context.Context, t test.Test, c cluster.Cluster, partition bool) {
		__antithesis_instrumentation__.Notify(50409)
		c.Put(ctx, t.Cockroach(), "./cockroach")
		c.Put(ctx, t.DeprecatedWorkload(), "./workload")
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

		nodes := []struct {
			i    int
			zone string
		}{
			{1, "us-central1-b"},
			{4, "us-west1-b"},
			{7, "europe-west2-b"},
		}

		roachmartRun := func(ctx context.Context, i int, args ...string) {
			__antithesis_instrumentation__.Notify(50412)
			args = append(args,
				"--local-zone="+nodes[i].zone,
				"--local-percent=90",
				"--users=10",
				"--orders=100",
				fmt.Sprintf("--partition=%v", partition))

			if err := c.RunE(ctx, c.Node(nodes[i].i), args...); err != nil {
				__antithesis_instrumentation__.Notify(50413)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50414)
			}
		}
		__antithesis_instrumentation__.Notify(50410)
		t.Status("initializing workload")
		roachmartRun(ctx, 0, "./workload", "init", "roachmart")

		duration := " --duration=" + ifLocal(c, "10s", "10m")

		t.Status("running workload")
		m := c.NewMonitor(ctx)
		for i := range nodes {
			__antithesis_instrumentation__.Notify(50415)
			i := i
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50416)
				roachmartRun(ctx, i, "./workload", "run", "roachmart", duration)
				return nil
			})
		}
		__antithesis_instrumentation__.Notify(50411)

		m.Wait()
	}
	__antithesis_instrumentation__.Notify(50408)

	for _, v := range []bool{true, false} {
		__antithesis_instrumentation__.Notify(50417)
		v := v
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("roachmart/partition=%v", v),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(9, spec.Geo(), spec.Zones("us-central1-b,us-west1-b,europe-west2-b")),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(50418)
				runRoachmart(ctx, t, c, v)
			},
		})
	}
}
