package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

func registerTPCE(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51852)
	type tpceOptions struct {
		customers int
		nodes     int
		cpus      int
		ssds      int

		tags    []string
		timeout time.Duration
	}

	runTPCE := func(ctx context.Context, t test.Test, c cluster.Cluster, opts tpceOptions) {
		__antithesis_instrumentation__.Notify(51854)
		roachNodes := c.Range(1, opts.nodes)
		loadNode := c.Node(opts.nodes + 1)
		racks := opts.nodes

		t.Status("installing cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)

		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.StoreCount = opts.ssds
		settings := install.MakeClusterSettings(install.NumRacksOption(racks))
		c.Start(ctx, t.L(), startOpts, settings, roachNodes)

		t.Status("installing docker")
		if err := c.Install(ctx, t.L(), loadNode, "docker"); err != nil {
			__antithesis_instrumentation__.Notify(51858)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51859)
		}
		__antithesis_instrumentation__.Notify(51855)

		func() {
			__antithesis_instrumentation__.Notify(51860)
			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()
			if _, err := db.ExecContext(
				ctx, "SET CLUSTER SETTING kv.bulk_io_write.concurrent_addsstable_requests = $1", 4*opts.ssds,
			); err != nil {
				__antithesis_instrumentation__.Notify(51862)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51863)
			}
			__antithesis_instrumentation__.Notify(51861)
			if _, err := db.ExecContext(
				ctx, "SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false",
			); err != nil {
				__antithesis_instrumentation__.Notify(51864)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51865)
			}
		}()
		__antithesis_instrumentation__.Notify(51856)

		m := c.NewMonitor(ctx, roachNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51866)
			const dockerRun = `sudo docker run cockroachdb/tpc-e:latest`

			roachNodeIPs, err := c.InternalIP(ctx, t.L(), roachNodes)
			if err != nil {
				__antithesis_instrumentation__.Notify(51871)
				return err
			} else {
				__antithesis_instrumentation__.Notify(51872)
			}
			__antithesis_instrumentation__.Notify(51867)
			roachNodeIPFlags := make([]string, len(roachNodeIPs))
			for i, ip := range roachNodeIPs {
				__antithesis_instrumentation__.Notify(51873)
				roachNodeIPFlags[i] = fmt.Sprintf("--hosts=%s", ip)
			}
			__antithesis_instrumentation__.Notify(51868)

			t.Status("preparing workload")
			c.Run(ctx, loadNode, fmt.Sprintf("%s --customers=%d --racks=%d --init %s",
				dockerRun, opts.customers, racks, roachNodeIPFlags[0]))

			t.Status("running workload")
			duration := 2 * time.Hour
			threads := opts.nodes * opts.cpus
			result, err := c.RunWithDetailsSingleNode(ctx, t.L(), loadNode,
				fmt.Sprintf("%s --customers=%d --racks=%d --duration=%s --threads=%d %s",
					dockerRun, opts.customers, racks, duration, threads, strings.Join(roachNodeIPFlags, " ")))
			if err != nil {
				__antithesis_instrumentation__.Notify(51874)
				t.Fatal(err.Error())
			} else {
				__antithesis_instrumentation__.Notify(51875)
			}
			__antithesis_instrumentation__.Notify(51869)
			t.L().Printf("workload output:\n%s\n", result.Stdout)
			if strings.Contains(result.Stdout, "Reported tpsE :    --   (not between 80% and 100%)") {
				__antithesis_instrumentation__.Notify(51876)
				return errors.New("invalid tpsE fraction")
			} else {
				__antithesis_instrumentation__.Notify(51877)
			}
			__antithesis_instrumentation__.Notify(51870)
			return nil
		})
		__antithesis_instrumentation__.Notify(51857)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(51853)

	for _, opts := range []tpceOptions{

		{customers: 5_000, nodes: 3, cpus: 4, ssds: 1},

		{customers: 100_000, nodes: 5, cpus: 32, ssds: 2, tags: []string{"weekly"}, timeout: 36 * time.Hour},
	} {
		__antithesis_instrumentation__.Notify(51878)
		opts := opts
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("tpce/c=%d/nodes=%d", opts.customers, opts.nodes),
			Owner:   registry.OwnerKV,
			Tags:    opts.tags,
			Timeout: opts.timeout,
			Cluster: r.MakeClusterSpec(opts.nodes+1, spec.CPU(opts.cpus), spec.SSD(opts.ssds)),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(51879)
				runTPCE(ctx, t, c, opts)
			},
		})
	}
}
