package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

const envYCSBFlags = "ROACHTEST_YCSB_FLAGS"

func registerYCSB(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52560)
	workloads := []string{"A", "B", "C", "D", "E", "F"}
	cpusConfigs := []int{8, 32}

	concurrencyConfigs := map[string]map[int]int{
		"A": {8: 96, 32: 144},
		"B": {8: 144, 32: 192},
		"C": {8: 144, 32: 192},
		"D": {8: 96, 32: 144},
		"E": {8: 96, 32: 144},
		"F": {8: 96, 32: 144},
	}

	runYCSB := func(ctx context.Context, t test.Test, c cluster.Cluster, wl string, cpus int) {
		__antithesis_instrumentation__.Notify(52562)

		if c.Spec().FileSystem == spec.Zfs && func() bool {
			__antithesis_instrumentation__.Notify(52566)
			return c.Spec().Cloud != spec.GCE == true
		}() == true {
			__antithesis_instrumentation__.Notify(52567)
			t.Skip("YCSB zfs benchmark can only be run on GCE", "")
		} else {
			__antithesis_instrumentation__.Notify(52568)
		}
		__antithesis_instrumentation__.Notify(52563)

		nodes := c.Spec().NodeCount - 1

		conc, ok := concurrencyConfigs[wl][cpus]
		if !ok {
			__antithesis_instrumentation__.Notify(52569)
			t.Fatalf("missing concurrency for (workload, cpus) = (%s, %d)", wl, cpus)
		} else {
			__antithesis_instrumentation__.Notify(52570)
		}
		__antithesis_instrumentation__.Notify(52564)

		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))
		err := WaitFor3XReplication(ctx, t, c.Conn(ctx, t.L(), 1))
		require.NoError(t, err)

		t.Status("running workload")
		m := c.NewMonitor(ctx, c.Range(1, nodes))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(52571)
			var args string
			args += fmt.Sprintf(" --select-for-update=%t", t.IsBuildVersion("v19.2.0"))
			args += " --ramp=" + ifLocal(c, "0s", "2m")
			args += " --duration=" + ifLocal(c, "10s", "30m")
			if envFlags := os.Getenv(envYCSBFlags); envFlags != "" {
				__antithesis_instrumentation__.Notify(52573)
				args += " " + envFlags
			} else {
				__antithesis_instrumentation__.Notify(52574)
			}
			__antithesis_instrumentation__.Notify(52572)
			cmd := fmt.Sprintf(
				"./workload run ycsb --init --insert-count=1000000 --workload=%s --concurrency=%d"+
					" --splits=%d --histograms="+t.PerfArtifactsDir()+"/stats.json"+args+
					" {pgurl:1-%d}",
				wl, conc, nodes, nodes)
			c.Run(ctx, c.Node(nodes+1), cmd)
			return nil
		})
		__antithesis_instrumentation__.Notify(52565)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(52561)

	for _, wl := range workloads {
		__antithesis_instrumentation__.Notify(52575)
		for _, cpus := range cpusConfigs {
			__antithesis_instrumentation__.Notify(52576)
			var name string
			if cpus == 8 {
				__antithesis_instrumentation__.Notify(52579)
				name = fmt.Sprintf("ycsb/%s/nodes=3", wl)
			} else {
				__antithesis_instrumentation__.Notify(52580)
				name = fmt.Sprintf("ycsb/%s/nodes=3/cpu=%d", wl, cpus)
			}
			__antithesis_instrumentation__.Notify(52577)
			wl, cpus := wl, cpus
			r.Add(registry.TestSpec{
				Name:    name,
				Owner:   registry.OwnerKV,
				Cluster: r.MakeClusterSpec(4, spec.CPU(cpus)),
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(52581)
					runYCSB(ctx, t, c, wl, cpus)
				},
			})
			__antithesis_instrumentation__.Notify(52578)

			if wl == "A" {
				__antithesis_instrumentation__.Notify(52582)
				r.Add(registry.TestSpec{
					Name:    fmt.Sprintf("zfs/ycsb/%s/nodes=3/cpu=%d", wl, cpus),
					Owner:   registry.OwnerStorage,
					Cluster: r.MakeClusterSpec(4, spec.CPU(cpus), spec.SetFileSystem(spec.Zfs)),
					Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
						__antithesis_instrumentation__.Notify(52583)
						runYCSB(ctx, t, c, wl, cpus)
					},
				})
			} else {
				__antithesis_instrumentation__.Notify(52584)
			}
		}
	}
}
