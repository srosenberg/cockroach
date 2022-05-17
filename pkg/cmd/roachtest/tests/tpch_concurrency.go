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
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/workload/tpch"
	"github.com/stretchr/testify/require"
)

func registerTPCHConcurrency(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51880)
	const numNodes = 4

	setupCluster := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(51885)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, numNodes-1))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(numNodes))
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, numNodes-1))

		conn := c.Conn(ctx, t.L(), 1)
		if _, err := conn.Exec("SET CLUSTER SETTING kv.allocator.min_lease_transfer_interval = '24h';"); err != nil {
			__antithesis_instrumentation__.Notify(51888)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51889)
		}
		__antithesis_instrumentation__.Notify(51886)
		if _, err := conn.Exec("SET CLUSTER SETTING kv.range_merge.queue_enabled = false;"); err != nil {
			__antithesis_instrumentation__.Notify(51890)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51891)
		}
		__antithesis_instrumentation__.Notify(51887)

		if err := loadTPCHDataset(ctx, t, c, 1, c.NewMonitor(ctx, c.Range(1, numNodes-1)), c.Range(1, numNodes-1)); err != nil {
			__antithesis_instrumentation__.Notify(51892)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51893)
		}
	}
	__antithesis_instrumentation__.Notify(51881)

	restartCluster := func(ctx context.Context, c cluster.Cluster, t test.Test) {
		__antithesis_instrumentation__.Notify(51894)
		c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Range(1, numNodes-1))
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, numNodes-1))
	}
	__antithesis_instrumentation__.Notify(51882)

	checkConcurrency := func(ctx context.Context, t test.Test, c cluster.Cluster, concurrency int) error {
		__antithesis_instrumentation__.Notify(51895)

		_ = c.RunE(ctx, c.Node(numNodes), "killall workload")

		restartCluster(ctx, c, t)

		conn := c.Conn(ctx, t.L(), 1)
		if _, err := conn.Exec("USE tpch;"); err != nil {
			__antithesis_instrumentation__.Notify(51899)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51900)
		}
		__antithesis_instrumentation__.Notify(51896)
		scatterTables(t, conn, tpchTables)
		err := WaitFor3XReplication(ctx, t, conn)
		require.NoError(t, err)

		for nodeIdx := 1; nodeIdx < numNodes; nodeIdx++ {
			__antithesis_instrumentation__.Notify(51901)
			node := c.Conn(ctx, t.L(), nodeIdx)
			if _, err := node.Exec("USE tpch;"); err != nil {
				__antithesis_instrumentation__.Notify(51903)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51904)
			}
			__antithesis_instrumentation__.Notify(51902)
			for _, table := range tpchTables {
				__antithesis_instrumentation__.Notify(51905)
				if _, err := node.Exec(fmt.Sprintf("SELECT count(*) FROM %s", table)); err != nil {
					__antithesis_instrumentation__.Notify(51906)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(51907)
				}
			}
		}
		__antithesis_instrumentation__.Notify(51897)

		m := c.NewMonitor(ctx, c.Range(1, numNodes-1))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(51908)
			t.Status(fmt.Sprintf("running with concurrency = %d", concurrency))

			for queryNum := 1; queryNum <= tpch.NumQueries; queryNum++ {
				__antithesis_instrumentation__.Notify(51910)
				t.Status("running Q", queryNum)

				rows, err := conn.Query("EXPLAIN (DISTSQL) " + tpch.QueriesByNumber[queryNum])
				if err != nil {
					__antithesis_instrumentation__.Notify(51913)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(51914)
				}
				__antithesis_instrumentation__.Notify(51911)
				defer rows.Close()
				for rows.Next() {
					__antithesis_instrumentation__.Notify(51915)
					var line string
					if err = rows.Scan(&line); err != nil {
						__antithesis_instrumentation__.Notify(51917)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(51918)
					}
					__antithesis_instrumentation__.Notify(51916)
					if strings.Contains(line, "Diagram:") {
						__antithesis_instrumentation__.Notify(51919)
						t.Status(line)
					} else {
						__antithesis_instrumentation__.Notify(51920)
					}
				}
				__antithesis_instrumentation__.Notify(51912)

				maxOps := concurrency / 10

				cmd := fmt.Sprintf(
					"./workload run tpch {pgurl:1-%d} --display-every=1ns --tolerate-errors "+
						"--count-errors --queries=%d --concurrency=%d --max-ops=%d",
					numNodes-1, queryNum, concurrency, maxOps,
				)
				if err := c.RunE(ctx, c.Node(numNodes), cmd); err != nil {
					__antithesis_instrumentation__.Notify(51921)
					return err
				} else {
					__antithesis_instrumentation__.Notify(51922)
				}
			}
			__antithesis_instrumentation__.Notify(51909)
			return nil
		})
		__antithesis_instrumentation__.Notify(51898)
		return m.WaitE()
	}
	__antithesis_instrumentation__.Notify(51883)

	runTPCHConcurrency := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(51923)
		setupCluster(ctx, t, c)

		minConcurrency, maxConcurrency := 48, 160

		for minConcurrency < maxConcurrency-1 {
			__antithesis_instrumentation__.Notify(51925)
			concurrency := (minConcurrency + maxConcurrency) / 2
			if err := checkConcurrency(ctx, t, c, concurrency); err != nil {
				__antithesis_instrumentation__.Notify(51926)
				maxConcurrency = concurrency
			} else {
				__antithesis_instrumentation__.Notify(51927)
				minConcurrency = concurrency
			}
		}
		__antithesis_instrumentation__.Notify(51924)

		restartCluster(ctx, c, t)
		t.Status(fmt.Sprintf("max supported concurrency is %d", minConcurrency))

		c.Run(ctx, c.Node(numNodes), "mkdir", t.PerfArtifactsDir())
		cmd := fmt.Sprintf(
			`echo '{ "max_concurrency": %d }' > %s/stats.json`,
			minConcurrency, t.PerfArtifactsDir(),
		)
		c.Run(ctx, c.Node(numNodes), cmd)
	}
	__antithesis_instrumentation__.Notify(51884)

	r.Add(registry.TestSpec{
		Name:    "tpch_concurrency",
		Owner:   registry.OwnerSQLQueries,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(51928)
			runTPCHConcurrency(ctx, t, c)
		},

		Timeout: 12 * time.Hour,
	})
}
