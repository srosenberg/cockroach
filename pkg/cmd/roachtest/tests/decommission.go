package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
)

func registerDecommission(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46882)
	{
		__antithesis_instrumentation__.Notify(46883)
		numNodes := 4
		duration := time.Hour

		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("decommission/nodes=%d/duration=%s", numNodes, duration),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46884)
				if c.IsLocal() {
					__antithesis_instrumentation__.Notify(46886)
					duration = 5 * time.Minute
					t.L().Printf("running with duration=%s in local mode\n", duration)
				} else {
					__antithesis_instrumentation__.Notify(46887)
				}
				__antithesis_instrumentation__.Notify(46885)
				runDecommission(ctx, t, c, numNodes, duration)
			},
		})
	}
	{
		__antithesis_instrumentation__.Notify(46888)
		numNodes := 9
		duration := 30 * time.Minute
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("drain-and-decommission/nodes=%d", numNodes),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46889)
				runDrainAndDecommission(ctx, t, c, numNodes, duration)
			},
		})
	}
	{
		__antithesis_instrumentation__.Notify(46890)
		numNodes := 4
		r.Add(registry.TestSpec{
			Name:    "decommission/drains",
			Owner:   registry.OwnerServer,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46891)
				runDecommissionDrains(ctx, t, c)
			},
		})
	}
	{
		__antithesis_instrumentation__.Notify(46892)
		numNodes := 6
		r.Add(registry.TestSpec{
			Name:    "decommission/randomized",
			Owner:   registry.OwnerKV,
			Timeout: 10 * time.Minute,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46893)
				runDecommissionRandomized(ctx, t, c)
			},
		})
	}
	{
		__antithesis_instrumentation__.Notify(46894)
		numNodes := 4
		r.Add(registry.TestSpec{
			Name:    "decommission/mixed-versions",
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46895)
				runDecommissionMixedVersions(ctx, t, c, *t.BuildVersion())
			},
		})
	}
}

func runDrainAndDecommission(
	ctx context.Context, t test.Test, c cluster.Cluster, nodes int, duration time.Duration,
) {
	__antithesis_instrumentation__.Notify(46896)
	const defaultReplicationFactor = 5
	if defaultReplicationFactor > nodes {
		__antithesis_instrumentation__.Notify(46903)
		t.Fatal("improper configuration: replication factor greater than number of nodes in the test")
	} else {
		__antithesis_instrumentation__.Notify(46904)
	}
	__antithesis_instrumentation__.Notify(46897)
	pinnedNode := 1
	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	for i := 1; i <= nodes; i++ {
		__antithesis_instrumentation__.Notify(46905)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(i))
	}
	__antithesis_instrumentation__.Notify(46898)
	c.Run(ctx, c.Node(pinnedNode), `./cockroach workload init kv --drop --splits 1000`)

	run := func(stmt string) {
		__antithesis_instrumentation__.Notify(46906)
		db := c.Conn(ctx, t.L(), pinnedNode)
		defer db.Close()

		t.Status(stmt)
		_, err := db.ExecContext(ctx, stmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(46908)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46909)
		}
		__antithesis_instrumentation__.Notify(46907)
		t.L().Printf("run: %s\n", stmt)
	}

	{
		__antithesis_instrumentation__.Notify(46910)
		db := c.Conn(ctx, t.L(), pinnedNode)
		defer db.Close()

		run(fmt.Sprintf(`ALTER RANGE default CONFIGURE ZONE USING num_replicas=%d`, defaultReplicationFactor))
		run(fmt.Sprintf(`ALTER DATABASE system CONFIGURE ZONE USING num_replicas=%d`, defaultReplicationFactor))

		run(`SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='2GiB'`)
		run(`SET CLUSTER SETTING kv.snapshot_recovery.max_rate='2GiB'`)

		err := WaitFor3XReplication(ctx, t, db)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(46899)

	var m *errgroup.Group
	m, ctx = errgroup.WithContext(ctx)
	m.Go(
		func() error {
			__antithesis_instrumentation__.Notify(46911)
			return c.RunE(ctx, c.Node(pinnedNode),
				fmt.Sprintf("./cockroach workload run kv --max-rate 500 --tolerate-errors --duration=%s {pgurl:1-%d}",
					duration.String(), nodes-4,
				),
			)
		},
	)
	__antithesis_instrumentation__.Notify(46900)

	time.Sleep(1 * time.Minute)

	for nodeID := nodes - 2; nodeID <= nodes; nodeID++ {
		__antithesis_instrumentation__.Notify(46912)
		id := nodeID
		m.Go(func() error {
			__antithesis_instrumentation__.Notify(46913)
			drain := func(id int) error {
				__antithesis_instrumentation__.Notify(46915)
				t.Status(fmt.Sprintf("draining node %d", id))
				return c.RunE(ctx, c.Node(id), "./cockroach node drain --insecure")
			}
			__antithesis_instrumentation__.Notify(46914)
			return drain(id)
		})
	}
	__antithesis_instrumentation__.Notify(46901)

	time.Sleep(30 * time.Second)

	m.Go(func() error {
		__antithesis_instrumentation__.Notify(46916)

		id := nodes - 3
		decom := func(id int) error {
			__antithesis_instrumentation__.Notify(46918)
			t.Status(fmt.Sprintf("decommissioning node %d", id))
			return c.RunE(ctx, c.Node(id), "./cockroach node decommission --self --insecure")
		}
		__antithesis_instrumentation__.Notify(46917)
		return decom(id)
	})
	__antithesis_instrumentation__.Notify(46902)
	if err := m.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(46919)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46920)
	}
}

func runDecommission(
	ctx context.Context, t test.Test, c cluster.Cluster, nodes int, duration time.Duration,
) {
	__antithesis_instrumentation__.Notify(46921)
	const defaultReplicationFactor = 3
	if defaultReplicationFactor > nodes {
		__antithesis_instrumentation__.Notify(46930)
		t.Fatal("improper configuration: replication factor greater than number of nodes in the test")
	} else {
		__antithesis_instrumentation__.Notify(46931)
	}
	__antithesis_instrumentation__.Notify(46922)

	numDecom := (defaultReplicationFactor - 1) / 2

	pinnedNode := 1
	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(pinnedNode))

	for i := 1; i <= nodes; i++ {
		__antithesis_instrumentation__.Notify(46932)
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, fmt.Sprintf("--attrs=node%d", i))
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(i))
	}
	__antithesis_instrumentation__.Notify(46923)
	c.Run(ctx, c.Node(pinnedNode), `./workload init kv --drop`)

	waitReplicatedAwayFrom := func(downNodeID int) error {
		__antithesis_instrumentation__.Notify(46933)
		db := c.Conn(ctx, t.L(), pinnedNode)
		defer func() {
			__antithesis_instrumentation__.Notify(46936)
			_ = db.Close()
		}()
		__antithesis_instrumentation__.Notify(46934)

		for {
			__antithesis_instrumentation__.Notify(46937)
			var count int
			if err := db.QueryRow(

				"SELECT count(*) FROM crdb_internal.ranges WHERE array_position(replicas, $1) IS NOT NULL",
				downNodeID,
			).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(46940)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46941)
			}
			__antithesis_instrumentation__.Notify(46938)
			if count == 0 {
				__antithesis_instrumentation__.Notify(46942)
				fullReplicated := false
				if err := db.QueryRow(

					"SELECT min(array_length(replicas, 1)) >= 3 FROM crdb_internal.ranges",
				).Scan(&fullReplicated); err != nil {
					__antithesis_instrumentation__.Notify(46944)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46945)
				}
				__antithesis_instrumentation__.Notify(46943)
				if fullReplicated {
					__antithesis_instrumentation__.Notify(46946)
					break
				} else {
					__antithesis_instrumentation__.Notify(46947)
				}
			} else {
				__antithesis_instrumentation__.Notify(46948)
			}
			__antithesis_instrumentation__.Notify(46939)
			time.Sleep(time.Second)
		}
		__antithesis_instrumentation__.Notify(46935)
		return nil
	}
	__antithesis_instrumentation__.Notify(46924)

	waitUpReplicated := func(targetNode, targetNodeID int) error {
		__antithesis_instrumentation__.Notify(46949)
		db := c.Conn(ctx, t.L(), pinnedNode)
		defer func() {
			__antithesis_instrumentation__.Notify(46952)
			_ = db.Close()
		}()
		__antithesis_instrumentation__.Notify(46950)

		var count int
		for {
			__antithesis_instrumentation__.Notify(46953)

			stmtReplicaCount := fmt.Sprintf(
				`SELECT count(*) FROM crdb_internal.ranges WHERE array_position(replicas, %d) IS NULL and database_name = 'kv';`, targetNodeID)
			if err := db.QueryRow(stmtReplicaCount).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(46956)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46957)
			}
			__antithesis_instrumentation__.Notify(46954)
			t.Status(fmt.Sprintf("node%d missing %d replica(s)", targetNode, count))
			if count == 0 {
				__antithesis_instrumentation__.Notify(46958)
				break
			} else {
				__antithesis_instrumentation__.Notify(46959)
			}
			__antithesis_instrumentation__.Notify(46955)
			time.Sleep(time.Second)
		}
		__antithesis_instrumentation__.Notify(46951)
		return nil
	}
	__antithesis_instrumentation__.Notify(46925)

	if err := waitReplicatedAwayFrom(0); err != nil {
		__antithesis_instrumentation__.Notify(46960)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(46961)
	}
	__antithesis_instrumentation__.Notify(46926)

	workloads := []string{

		fmt.Sprintf("./workload run kv --max-rate 500 --tolerate-errors --duration=%s {pgurl:1-%d}", duration.String(), nodes),
	}

	run := func(stmt string) {
		__antithesis_instrumentation__.Notify(46962)
		db := c.Conn(ctx, t.L(), pinnedNode)
		defer db.Close()

		t.Status(stmt)
		_, err := db.ExecContext(ctx, stmt)
		if err != nil {
			__antithesis_instrumentation__.Notify(46964)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46965)
		}
		__antithesis_instrumentation__.Notify(46963)
		t.L().Printf("run: %s\n", stmt)
	}
	__antithesis_instrumentation__.Notify(46927)

	var m *errgroup.Group
	m, ctx = errgroup.WithContext(ctx)
	for _, cmd := range workloads {
		__antithesis_instrumentation__.Notify(46966)
		cmd := cmd
		m.Go(func() error {
			__antithesis_instrumentation__.Notify(46967)
			return c.RunE(ctx, c.Node(pinnedNode), cmd)
		})
	}
	__antithesis_instrumentation__.Notify(46928)

	m.Go(func() error {
		__antithesis_instrumentation__.Notify(46968)
		getNodeID := func(node int) (int, error) {
			__antithesis_instrumentation__.Notify(46973)
			dbNode := c.Conn(ctx, t.L(), node)
			defer dbNode.Close()

			var nodeID int
			if err := dbNode.QueryRow(`SELECT node_id FROM crdb_internal.node_runtime_info LIMIT 1`).Scan(&nodeID); err != nil {
				__antithesis_instrumentation__.Notify(46975)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(46976)
			}
			__antithesis_instrumentation__.Notify(46974)
			return nodeID, nil
		}
		__antithesis_instrumentation__.Notify(46969)

		stop := func(node int) error {
			__antithesis_instrumentation__.Notify(46977)
			port := fmt.Sprintf("{pgport:%d}", node)
			defer time.Sleep(time.Second)
			return c.RunE(ctx, c.Node(node), "./cockroach quit --insecure --host=:"+port)
		}
		__antithesis_instrumentation__.Notify(46970)

		decom := func(id int) error {
			__antithesis_instrumentation__.Notify(46978)
			port := fmt.Sprintf("{pgport:%d}", pinnedNode)
			t.Status(fmt.Sprintf("decommissioning node %d", id))
			return c.RunE(ctx, c.Node(pinnedNode), fmt.Sprintf("./cockroach node decommission --insecure --wait=all --host=:%s %d", port, id))
		}
		__antithesis_instrumentation__.Notify(46971)

		tBegin, whileDown := timeutil.Now(), true
		node := nodes
		for timeutil.Since(tBegin) <= duration {
			__antithesis_instrumentation__.Notify(46979)

			whileDown = !whileDown

			node = nodes - (node % numDecom)
			if node == pinnedNode {
				__antithesis_instrumentation__.Notify(46989)
				t.Fatalf("programming error: not expecting to decommission/wipe node%d", pinnedNode)
			} else {
				__antithesis_instrumentation__.Notify(46990)
			}
			__antithesis_instrumentation__.Notify(46980)

			t.Status(fmt.Sprintf("decommissioning %d (down=%t)", node, whileDown))
			nodeID, err := getNodeID(node)
			if err != nil {
				__antithesis_instrumentation__.Notify(46991)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46992)
			}
			__antithesis_instrumentation__.Notify(46981)

			run(fmt.Sprintf(`ALTER RANGE default CONFIGURE ZONE = 'constraints: {"+node%d"}'`, node))
			if err := waitUpReplicated(node, nodeID); err != nil {
				__antithesis_instrumentation__.Notify(46993)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46994)
			}
			__antithesis_instrumentation__.Notify(46982)

			if whileDown {
				__antithesis_instrumentation__.Notify(46995)
				if err := stop(node); err != nil {
					__antithesis_instrumentation__.Notify(46996)
					return err
				} else {
					__antithesis_instrumentation__.Notify(46997)
				}
			} else {
				__antithesis_instrumentation__.Notify(46998)
			}
			__antithesis_instrumentation__.Notify(46983)

			run(fmt.Sprintf(`ALTER RANGE default CONFIGURE ZONE = 'constraints: {"-node%d"}'`, node))

			if err := decom(nodeID); err != nil {
				__antithesis_instrumentation__.Notify(46999)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47000)
			}
			__antithesis_instrumentation__.Notify(46984)

			if err := waitReplicatedAwayFrom(nodeID); err != nil {
				__antithesis_instrumentation__.Notify(47001)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47002)
			}
			__antithesis_instrumentation__.Notify(46985)

			if !whileDown {
				__antithesis_instrumentation__.Notify(47003)
				if err := stop(node); err != nil {
					__antithesis_instrumentation__.Notify(47004)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47005)
				}
			} else {
				__antithesis_instrumentation__.Notify(47006)
			}
			__antithesis_instrumentation__.Notify(46986)

			if err := c.RunE(ctx, c.Node(node), "rm -rf {store-dir}"); err != nil {
				__antithesis_instrumentation__.Notify(47007)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47008)
			}
			__antithesis_instrumentation__.Notify(46987)

			db := c.Conn(ctx, t.L(), pinnedNode)
			defer db.Close()

			internalAddrs, err := c.InternalAddr(ctx, t.L(), c.Node(pinnedNode))
			if err != nil {
				__antithesis_instrumentation__.Notify(47009)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47010)
			}
			__antithesis_instrumentation__.Notify(46988)
			startOpts := option.DefaultStartOpts()
			startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, fmt.Sprintf("--join %s --attrs=node%d", internalAddrs[0], node))
			if err := c.StartE(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(node)); err != nil {
				__antithesis_instrumentation__.Notify(47011)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47012)
			}
		}
		__antithesis_instrumentation__.Notify(46972)

		return nil
	})
	__antithesis_instrumentation__.Notify(46929)
	if err := m.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(47013)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47014)
	}
}

func runDecommissionRandomized(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(47015)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	settings := install.MakeClusterSettings()
	settings.Env = append(settings.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings)

	h := newDecommTestHelper(t, c)

	firstNodeID := h.nodeIDs[0]
	retryOpts := retry.Options{
		InitialBackoff: time.Second,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2,
	}

	warningFilter := []string{
		"^warning: node [0-9]+ is already decommissioning or decommissioned$",
	}

	{
		__antithesis_instrumentation__.Notify(47018)
		targetNode, runNode := h.nodeIDs[0], h.getRandNode()
		t.L().Printf("partially decommissioning n%d from n%d\n", targetNode, runNode)
		o, err := h.decommission(ctx, c.Node(targetNode), runNode,
			"--wait=none", "--format=csv")
		if err != nil {
			__antithesis_instrumentation__.Notify(47020)
			t.Fatalf("decommission failed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(47021)
		}
		__antithesis_instrumentation__.Notify(47019)

		exp := [][]string{
			decommissionHeader,
			{strconv.Itoa(targetNode), "true", `\d+`, "true", "decommissioning", "false"},
		}
		if err := cli.MatchCSV(o, exp); err != nil {
			__antithesis_instrumentation__.Notify(47022)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47023)
		}

		{
			__antithesis_instrumentation__.Notify(47024)
			runNode = h.getRandNode()
			t.L().Printf("checking that `node status` (from n%d) shows n%d as decommissioning\n",
				runNode, targetNode)
			o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv", "--decommission")
			if err != nil {
				__antithesis_instrumentation__.Notify(47026)
				t.Fatalf("node-status failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47027)
			}
			__antithesis_instrumentation__.Notify(47025)

			numCols, err := cli.GetCsvNumCols(o)
			require.NoError(t, err)
			exp := h.expectCell(targetNode-1,
				statusHeaderMembershipColumnIdx, `decommissioning`, c.Spec().NodeCount, numCols)
			if err := cli.MatchCSV(o, exp); err != nil {
				__antithesis_instrumentation__.Notify(47028)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47029)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47030)
			runNode = h.getRandNode()
			t.L().Printf("recommissioning n%d (from n%d)\n", targetNode, runNode)
			if _, err := h.recommission(ctx, c.Node(targetNode), runNode); err != nil {
				__antithesis_instrumentation__.Notify(47031)
				t.Fatalf("recommission failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47032)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47033)
			runNode = h.getRandNode()
			t.L().Printf("checking that `node status` (from n%d) shows n%d as active\n",
				targetNode, runNode)
			o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv", "--decommission")
			if err != nil {
				__antithesis_instrumentation__.Notify(47035)
				t.Fatalf("node-status failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47036)
			}
			__antithesis_instrumentation__.Notify(47034)

			numCols, err := cli.GetCsvNumCols(o)
			require.NoError(t, err)
			exp := h.expectCell(targetNode-1,
				statusHeaderMembershipColumnIdx, `active`, c.Spec().NodeCount, numCols)
			if err := cli.MatchCSV(o, exp); err != nil {
				__antithesis_instrumentation__.Notify(47037)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47038)
			}
		}
	}

	{
		__antithesis_instrumentation__.Notify(47039)

		{
			__antithesis_instrumentation__.Notify(47040)

			if err := retry.WithMaxAttempts(ctx, retryOpts, 50, func() error {
				__antithesis_instrumentation__.Notify(47041)
				runNode := h.getRandNode()
				t.L().Printf("attempting to decommission all nodes from n%d\n", runNode)
				o, err := h.decommission(ctx, c.All(), runNode,
					"--wait=none", "--format=csv")
				if err != nil {
					__antithesis_instrumentation__.Notify(47045)
					t.Fatalf("decommission failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47046)
				}
				__antithesis_instrumentation__.Notify(47042)

				exp := [][]string{decommissionHeader}
				for i := 1; i <= c.Spec().NodeCount; i++ {
					__antithesis_instrumentation__.Notify(47047)
					rowRegex := []string{strconv.Itoa(i), "true", `\d+`, "true", "decommissioning", "false"}
					exp = append(exp, rowRegex)
				}
				__antithesis_instrumentation__.Notify(47043)
				if err := cli.MatchCSV(o, exp); err != nil {
					__antithesis_instrumentation__.Notify(47048)
					t.Fatalf("decommission failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47049)
				}
				__antithesis_instrumentation__.Notify(47044)
				return nil
			}); err != nil {
				__antithesis_instrumentation__.Notify(47050)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47051)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47052)
			runNode := h.getRandNode()
			t.L().Printf("checking that `node status` (from n%d) shows all nodes as decommissioning\n",
				runNode)
			o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv", "--decommission")
			if err != nil {
				__antithesis_instrumentation__.Notify(47055)
				t.Fatalf("node-status failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47056)
			}
			__antithesis_instrumentation__.Notify(47053)

			numCols, err := cli.GetCsvNumCols(o)
			require.NoError(t, err)
			var colRegex []string
			for i := 1; i <= c.Spec().NodeCount; i++ {
				__antithesis_instrumentation__.Notify(47057)
				colRegex = append(colRegex, `decommissioning`)
			}
			__antithesis_instrumentation__.Notify(47054)
			exp := h.expectColumn(statusHeaderMembershipColumnIdx, colRegex, c.Spec().NodeCount, numCols)
			if err := cli.MatchCSV(o, exp); err != nil {
				__antithesis_instrumentation__.Notify(47058)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47059)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47060)
			runNode := h.getRandNode()
			t.L().Printf("checking that we're able to create a database (from n%d)\n", runNode)
			db := c.Conn(ctx, t.L(), runNode)
			defer db.Close()

			if _, err := db.Exec(`create database still_working;`); err != nil {
				__antithesis_instrumentation__.Notify(47061)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47062)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47063)
			runNode := h.getRandNode()
			t.L().Printf("recommissioning all nodes (from n%d)\n", runNode)
			if _, err := h.recommission(ctx, c.All(), runNode); err != nil {
				__antithesis_instrumentation__.Notify(47064)
				t.Fatalf("recommission failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47065)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47066)
			runNode := h.getRandNode()
			t.L().Printf("checking that `node status` (from n%d) shows all nodes as active\n",
				runNode)
			o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv", "--decommission")
			if err != nil {
				__antithesis_instrumentation__.Notify(47069)
				t.Fatalf("node-status failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47070)
			}
			__antithesis_instrumentation__.Notify(47067)

			numCols, err := cli.GetCsvNumCols(o)
			require.NoError(t, err)
			var colRegex []string
			for i := 1; i <= c.Spec().NodeCount; i++ {
				__antithesis_instrumentation__.Notify(47071)
				colRegex = append(colRegex, `active`)
			}
			__antithesis_instrumentation__.Notify(47068)
			exp := h.expectColumn(statusHeaderMembershipColumnIdx, colRegex, c.Spec().NodeCount, numCols)
			if err := cli.MatchCSV(o, exp); err != nil {
				__antithesis_instrumentation__.Notify(47072)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47073)
			}
		}
	}
	__antithesis_instrumentation__.Notify(47016)

	decommissionedNodeA := h.getRandNode()
	h.blockFromRandNode(decommissionedNodeA)
	decommissionedNodeB := h.getRandNode()
	h.blockFromRandNode(decommissionedNodeB)
	{
		__antithesis_instrumentation__.Notify(47074)
		targetNodeA, targetNodeB := decommissionedNodeA, decommissionedNodeB
		if targetNodeB < targetNodeA {
			__antithesis_instrumentation__.Notify(47078)
			targetNodeB, targetNodeA = targetNodeA, targetNodeB
		} else {
			__antithesis_instrumentation__.Notify(47079)
		}
		__antithesis_instrumentation__.Notify(47075)

		runNode := h.getRandNode()
		waitStrategy := "all"
		if i := rand.Intn(2); i == 0 {
			__antithesis_instrumentation__.Notify(47080)
			waitStrategy = "none"
		} else {
			__antithesis_instrumentation__.Notify(47081)
		}
		__antithesis_instrumentation__.Notify(47076)

		t.L().Printf("fully decommissioning [n%d,n%d] from n%d, using --wait=%s\n",
			targetNodeA, targetNodeB, runNode, waitStrategy)

		maxAttempts := 50
		if waitStrategy == "all" {
			__antithesis_instrumentation__.Notify(47082)

			maxAttempts = 1
		} else {
			__antithesis_instrumentation__.Notify(47083)
		}
		__antithesis_instrumentation__.Notify(47077)

		if err := retry.WithMaxAttempts(ctx, retryOpts, maxAttempts, func() error {
			__antithesis_instrumentation__.Notify(47084)
			o, err := h.decommission(ctx, c.Nodes(targetNodeA, targetNodeB), runNode,
				fmt.Sprintf("--wait=%s", waitStrategy), "--format=csv")
			if err != nil {
				__antithesis_instrumentation__.Notify(47086)
				t.Fatalf("decommission failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47087)
			}
			__antithesis_instrumentation__.Notify(47085)

			exp := [][]string{
				decommissionHeader,
				{strconv.Itoa(targetNodeA), "true|false", "0", "true", "decommissioned", "false"},
				{strconv.Itoa(targetNodeB), "true|false", "0", "true", "decommissioned", "false"},
				decommissionFooter,
			}
			return cli.MatchCSV(cli.RemoveMatchingLines(o, warningFilter), exp)
		}); err != nil {
			__antithesis_instrumentation__.Notify(47088)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47089)
		}

		{
			__antithesis_instrumentation__.Notify(47090)
			if err := retry.WithMaxAttempts(ctx, retry.Options{}, 50, func() error {
				__antithesis_instrumentation__.Notify(47091)
				runNode = h.getRandNode()
				t.L().Printf("checking that `node ls` (from n%d) shows n-2 nodes\n", runNode)
				o, err := execCLI(ctx, t, c, runNode, "node", "ls", "--format=csv")
				if err != nil {
					__antithesis_instrumentation__.Notify(47094)
					t.Fatalf("node-ls failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47095)
				}
				__antithesis_instrumentation__.Notify(47092)
				exp := [][]string{{"id"}}
				for i := 1; i <= c.Spec().NodeCount; i++ {
					__antithesis_instrumentation__.Notify(47096)
					if _, ok := h.randNodeBlocklist[i]; ok {
						__antithesis_instrumentation__.Notify(47098)

						continue
					} else {
						__antithesis_instrumentation__.Notify(47099)
					}
					__antithesis_instrumentation__.Notify(47097)
					exp = append(exp, []string{strconv.Itoa(i)})
				}
				__antithesis_instrumentation__.Notify(47093)
				return cli.MatchCSV(o, exp)
			}); err != nil {
				__antithesis_instrumentation__.Notify(47100)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47101)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47102)
			if err := retry.WithMaxAttempts(ctx, retry.Options{}, 50, func() error {
				__antithesis_instrumentation__.Notify(47103)
				runNode = h.getRandNode()
				t.L().Printf("checking that `node status` (from n%d) shows only live nodes\n", runNode)
				o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv")
				if err != nil {
					__antithesis_instrumentation__.Notify(47106)
					t.Fatalf("node-status failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47107)
				}
				__antithesis_instrumentation__.Notify(47104)

				numCols, err := cli.GetCsvNumCols(o)
				require.NoError(t, err)

				colRegex := []string{}
				for i := 1; i <= c.Spec().NodeCount; i++ {
					__antithesis_instrumentation__.Notify(47108)
					if _, ok := h.randNodeBlocklist[i]; ok {
						__antithesis_instrumentation__.Notify(47110)
						continue
					} else {
						__antithesis_instrumentation__.Notify(47111)
					}
					__antithesis_instrumentation__.Notify(47109)
					colRegex = append(colRegex, strconv.Itoa(i))
				}
				__antithesis_instrumentation__.Notify(47105)
				exp := h.expectIDsInStatusOut(colRegex, numCols)
				return cli.MatchCSV(o, exp)
			}); err != nil {
				__antithesis_instrumentation__.Notify(47112)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47113)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47114)
			runNode = h.getRandNode()
			t.L().Printf("expected to fail: recommissioning [n%d,n%d] (from n%d)\n",
				targetNodeA, targetNodeB, runNode)
			if _, err := h.recommission(ctx, c.Nodes(targetNodeA, targetNodeB), runNode); err == nil {
				__antithesis_instrumentation__.Notify(47115)
				t.Fatal("expected recommission to fail")
			} else {
				__antithesis_instrumentation__.Notify(47116)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47117)
			runNode = h.getRandNode()
			t.L().Printf("checking that decommissioning [n%d,n%d] (from n%d) is a no-op\n",
				targetNodeA, targetNodeB, runNode)
			o, err := h.decommission(ctx, c.Nodes(targetNodeA, targetNodeB), runNode,
				"--wait=all", "--format=csv")
			if err != nil {
				__antithesis_instrumentation__.Notify(47119)
				t.Fatalf("decommission failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47120)
			}
			__antithesis_instrumentation__.Notify(47118)

			exp := [][]string{
				decommissionHeader,

				{strconv.Itoa(targetNodeA), "false", "0", "true", "decommissioned", "false"},
				{strconv.Itoa(targetNodeB), "false", "0", "true", "decommissioned", "false"},
				decommissionFooter,
			}
			if err := cli.MatchCSV(cli.RemoveMatchingLines(o, warningFilter), exp); err != nil {
				__antithesis_instrumentation__.Notify(47121)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47122)
			}
		}

		{
			__antithesis_instrumentation__.Notify(47123)
			runNode = h.getRandNode()
			t.L().Printf("expected to fail: restarting [n%d,n%d] and attempting to recommission through n%d\n",
				targetNodeA, targetNodeB, runNode)
			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Nodes(targetNodeA, targetNodeB))
			c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Nodes(targetNodeA, targetNodeB))

			if _, err := h.recommission(ctx, c.Nodes(targetNodeA, targetNodeB), runNode); err == nil {
				__antithesis_instrumentation__.Notify(47125)
				t.Fatalf("expected recommission to fail")
			} else {
				__antithesis_instrumentation__.Notify(47126)
			}
			__antithesis_instrumentation__.Notify(47124)

			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Nodes(targetNodeA, targetNodeB))
			c.Wipe(ctx, c.Nodes(targetNodeA, targetNodeB))
		}
	}

	{
		__antithesis_instrumentation__.Notify(47127)
		restartDownedNode := false
		if i := rand.Intn(2); i == 0 {
			__antithesis_instrumentation__.Notify(47135)
			restartDownedNode = true
		} else {
			__antithesis_instrumentation__.Notify(47136)
		}
		__antithesis_instrumentation__.Notify(47128)

		if !restartDownedNode {
			__antithesis_instrumentation__.Notify(47137)

			func() {
				__antithesis_instrumentation__.Notify(47138)
				db := c.Conn(ctx, t.L(), h.getRandNode())
				defer db.Close()
				const stmt = "SET CLUSTER SETTING server.time_until_store_dead = '1m15s'"
				if _, err := db.ExecContext(ctx, stmt); err != nil {
					__antithesis_instrumentation__.Notify(47139)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(47140)
				}
			}()
		} else {
			__antithesis_instrumentation__.Notify(47141)
		}
		__antithesis_instrumentation__.Notify(47129)

		targetNode := h.getRandNodeOtherThan(firstNodeID)
		h.blockFromRandNode(targetNode)
		t.L().Printf("intentionally killing n%d to later decommission it when down\n", targetNode)
		c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(targetNode))

		runNode := h.getRandNode()
		t.L().Printf("decommissioning n%d (from n%d) in absentia\n", targetNode, runNode)
		if _, err := h.decommission(ctx, c.Node(targetNode), runNode,
			"--wait=all", "--format=csv"); err != nil {
			__antithesis_instrumentation__.Notify(47142)
			t.Fatalf("decommission failed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(47143)
		}
		__antithesis_instrumentation__.Notify(47130)

		if restartDownedNode {
			__antithesis_instrumentation__.Notify(47144)
			t.L().Printf("restarting n%d for verification\n", targetNode)

			c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Node(targetNode))
		} else {
			__antithesis_instrumentation__.Notify(47145)
		}
		__antithesis_instrumentation__.Notify(47131)

		o, err := h.decommission(ctx, c.Node(targetNode), runNode,
			"--wait=all", "--format=csv")
		if err != nil {
			__antithesis_instrumentation__.Notify(47146)
			t.Fatalf("decommission failed: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(47147)
		}
		__antithesis_instrumentation__.Notify(47132)

		exp := [][]string{
			decommissionHeader,
			{strconv.Itoa(targetNode), "true|false", "0", "true", "decommissioned", "false"},
			decommissionFooter,
		}
		if err := cli.MatchCSV(cli.RemoveMatchingLines(o, warningFilter), exp); err != nil {
			__antithesis_instrumentation__.Notify(47148)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47149)
		}
		__antithesis_instrumentation__.Notify(47133)

		if !restartDownedNode {
			__antithesis_instrumentation__.Notify(47150)

			if err := retry.WithMaxAttempts(ctx, retryOpts, 50, func() error {
				__antithesis_instrumentation__.Notify(47152)
				runNode := h.getRandNode()
				o, err := execCLI(ctx, t, c, runNode, "node", "ls", "--format=csv")
				if err != nil {
					__antithesis_instrumentation__.Notify(47155)
					t.Fatalf("node-ls failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47156)
				}
				__antithesis_instrumentation__.Notify(47153)

				var exp [][]string

				for i := 1; i <= c.Spec().NodeCount-len(h.randNodeBlocklist); i++ {
					__antithesis_instrumentation__.Notify(47157)
					exp = append(exp, []string{fmt.Sprintf("[^%d]", targetNode)})
				}
				__antithesis_instrumentation__.Notify(47154)

				return cli.MatchCSV(o, exp)
			}); err != nil {
				__antithesis_instrumentation__.Notify(47158)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47159)
			}
			__antithesis_instrumentation__.Notify(47151)

			if err := retry.WithMaxAttempts(ctx, retryOpts, 50, func() error {
				__antithesis_instrumentation__.Notify(47160)
				runNode := h.getRandNode()
				o, err := execCLI(ctx, t, c, runNode, "node", "status", "--format=csv")
				if err != nil {
					__antithesis_instrumentation__.Notify(47163)
					t.Fatalf("node-status failed: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(47164)
				}
				__antithesis_instrumentation__.Notify(47161)

				numCols, err := cli.GetCsvNumCols(o)
				require.NoError(t, err)
				var expC []string
				for i := 1; i <= c.Spec().NodeCount-len(h.randNodeBlocklist); i++ {
					__antithesis_instrumentation__.Notify(47165)
					expC = append(expC, fmt.Sprintf("[^%d].*", targetNode))
				}
				__antithesis_instrumentation__.Notify(47162)
				exp := h.expectIDsInStatusOut(expC, numCols)
				return cli.MatchCSV(o, exp)
			}); err != nil {
				__antithesis_instrumentation__.Notify(47166)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47167)
			}
		} else {
			__antithesis_instrumentation__.Notify(47168)
		}

		{
			__antithesis_instrumentation__.Notify(47169)
			t.L().Printf("wiping n%d and adding it back to the cluster as a new node\n", targetNode)

			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(targetNode))
			c.Wipe(ctx, c.Node(targetNode))

			joinNode := h.getRandNode()
			internalAddrs, err := c.InternalAddr(ctx, t.L(), c.Node(joinNode))
			if err != nil {
				__antithesis_instrumentation__.Notify(47171)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47172)
			}
			__antithesis_instrumentation__.Notify(47170)
			joinAddr := internalAddrs[0]
			startOpts := option.DefaultStartOpts()
			startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, fmt.Sprintf("--join %s", joinAddr))
			c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(targetNode))
		}
		__antithesis_instrumentation__.Notify(47134)

		if err := retry.WithMaxAttempts(ctx, retryOpts, 50, func() error {
			__antithesis_instrumentation__.Notify(47173)
			o, err := execCLI(ctx, t, c, h.getRandNode(), "node", "status", "--format=csv")
			if err != nil {
				__antithesis_instrumentation__.Notify(47177)
				t.Fatalf("node-status failed: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47178)
			}
			__antithesis_instrumentation__.Notify(47174)
			numCols, err := cli.GetCsvNumCols(o)
			require.NoError(t, err)
			var expC []string

			re := `[^`
			for id := range h.randNodeBlocklist {
				__antithesis_instrumentation__.Notify(47179)
				re += fmt.Sprint(id)
			}
			__antithesis_instrumentation__.Notify(47175)
			re += `].*`

			for i := 1; i <= c.Spec().NodeCount-len(h.randNodeBlocklist)+1; i++ {
				__antithesis_instrumentation__.Notify(47180)
				expC = append(expC, re)
			}
			__antithesis_instrumentation__.Notify(47176)
			exp := h.expectIDsInStatusOut(expC, numCols)
			return cli.MatchCSV(o, exp)
		}); err != nil {
			__antithesis_instrumentation__.Notify(47181)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47182)
		}
	}
	__antithesis_instrumentation__.Notify(47017)

	if err := retry.ForDuration(time.Minute, func() error {
		__antithesis_instrumentation__.Notify(47183)

		db := c.Conn(ctx, t.L(), h.getRandNode())
		defer db.Close()

		rows, err := db.Query(`
			SELECT "eventType" FROM system.eventlog WHERE "eventType" IN ($1, $2, $3) ORDER BY timestamp
			`, "node_decommissioned", "node_decommissioning", "node_recommissioned",
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47187)
			t.L().Printf("retrying: %v\n", err)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47188)
		}
		__antithesis_instrumentation__.Notify(47184)
		defer rows.Close()

		matrix, err := sqlutils.RowsToStrMatrix(rows)
		if err != nil {
			__antithesis_instrumentation__.Notify(47189)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47190)
		}
		__antithesis_instrumentation__.Notify(47185)

		expMatrix := [][]string{

			{"node_decommissioning"},
			{"node_recommissioned"},

			{"node_decommissioning"},
			{"node_decommissioning"},
			{"node_decommissioning"},
			{"node_decommissioning"},
			{"node_decommissioning"},
			{"node_decommissioning"},

			{"node_recommissioned"},
			{"node_recommissioned"},
			{"node_recommissioned"},
			{"node_recommissioned"},
			{"node_recommissioned"},
			{"node_recommissioned"},

			{"node_decommissioning"},
			{"node_decommissioning"},
			{"node_decommissioned"},
			{"node_decommissioned"},

			{"node_decommissioning"},
			{"node_decommissioned"},
		}

		if !reflect.DeepEqual(matrix, expMatrix) {
			__antithesis_instrumentation__.Notify(47191)
			t.Fatalf("unexpected diff(matrix, expMatrix):\n%s\n%s\nvs.\n%s", pretty.Diff(matrix, expMatrix), matrix, expMatrix)
		} else {
			__antithesis_instrumentation__.Notify(47192)
		}
		__antithesis_instrumentation__.Notify(47186)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(47193)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47194)
	}
}

func runDecommissionDrains(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(47195)
	var (
		numNodes     = 4
		pinnedNodeID = 1
		decommNodeID = numNodes
		decommNode   = c.Node(decommNodeID)
	)

	err := c.PutE(ctx, t.L(), t.Cockroach(), "./cockroach", c.All())
	require.NoError(t, err)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

	h := newDecommTestHelper(t, c)

	run := func(db *gosql.DB, query string) error {
		__antithesis_instrumentation__.Notify(47198)
		_, err = db.ExecContext(ctx, query)
		t.L().Printf("run: %s\n", query)
		return err
	}

	{
		__antithesis_instrumentation__.Notify(47199)
		db := c.Conn(ctx, t.L(), pinnedNodeID)
		defer db.Close()

		err = run(db, `SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='2GiB'`)
		require.NoError(t, err)
		err = run(db, `SET CLUSTER SETTING kv.snapshot_recovery.max_rate='2GiB'`)
		require.NoError(t, err)

		err := WaitFor3XReplication(ctx, t, db)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(47196)

	decommNodeDB := c.Conn(ctx, t.L(), decommNodeID)
	defer decommNodeDB.Close()

	var (
		maxAttempts = 50
		retryOpts   = retry.Options{
			InitialBackoff: time.Second,
			MaxBackoff:     5 * time.Second,
			Multiplier:     2,
		}

		expReplicasTransferred = [][]string{
			decommissionHeader,
			{strconv.Itoa(decommNodeID), "true|false", "0", "true", "decommissioning", "false"},
			decommissionFooter,
		}

		expDecommissioned = [][]string{
			decommissionHeader,
			{strconv.Itoa(decommNodeID), "true|false", "0", "true", "decommissioned", "false"},
			decommissionFooter,
		}
	)
	t.Status(fmt.Sprintf("decommissioning node %d", decommNodeID))
	e := retry.WithMaxAttempts(ctx, retryOpts, maxAttempts, func() error {
		__antithesis_instrumentation__.Notify(47200)
		o, err := h.decommission(ctx, decommNode, pinnedNodeID, "--wait=none", "--format=csv")
		require.NoError(t, errors.Wrapf(err, "decommission failed"))

		if err = cli.MatchCSV(o, expReplicasTransferred); err != nil {
			__antithesis_instrumentation__.Notify(47203)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47204)
		}
		__antithesis_instrumentation__.Notify(47201)

		if err = run(decommNodeDB, `SHOW DATABASES`); err != nil {
			__antithesis_instrumentation__.Notify(47205)
			if strings.Contains(err.Error(), "not accepting clients") {
				__antithesis_instrumentation__.Notify(47207)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(47208)
			}
			__antithesis_instrumentation__.Notify(47206)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47209)
		}
		__antithesis_instrumentation__.Notify(47202)

		return errors.New("not drained")
	})
	__antithesis_instrumentation__.Notify(47197)
	require.NoError(t, e)

	o, err := h.decommission(ctx, decommNode, pinnedNodeID, "--format=csv")
	require.NoError(t, err)
	require.NoError(t, cli.MatchCSV(o, expDecommissioned))
}

var decommissionHeader = []string{
	"id", "is_live", "replicas", "is_decommissioning", "membership", "is_draining",
}

var decommissionFooter = []string{
	"No more data reported on target nodes. " +
		"Please verify cluster health before removing the nodes.",
}

var statusHeader = []string{
	"id", "address", "sql_address", "build", "started_at", "updated_at", "locality", "is_available", "is_live",
}

var statusHeaderWithDecommission = []string{
	"id", "address", "sql_address", "build", "started_at", "updated_at", "locality", "is_available", "is_live",
	"gossiped_replicas", "is_decommissioning", "membership", "is_draining",
}

const statusHeaderMembershipColumnIdx = 11

type decommTestHelper struct {
	t       test.Test
	c       cluster.Cluster
	nodeIDs []int

	randNodeBlocklist map[int]struct{}
}

func newDecommTestHelper(t test.Test, c cluster.Cluster) *decommTestHelper {
	__antithesis_instrumentation__.Notify(47210)
	var nodeIDs []int
	for i := 1; i <= c.Spec().NodeCount; i++ {
		__antithesis_instrumentation__.Notify(47212)
		nodeIDs = append(nodeIDs, i)
	}
	__antithesis_instrumentation__.Notify(47211)
	return &decommTestHelper{
		t:                 t,
		c:                 c,
		nodeIDs:           nodeIDs,
		randNodeBlocklist: map[int]struct{}{},
	}
}

func (h *decommTestHelper) decommission(
	ctx context.Context, targetNodes option.NodeListOption, runNode int, verbs ...string,
) (string, error) {
	__antithesis_instrumentation__.Notify(47213)
	args := []string{"node", "decommission"}
	args = append(args, verbs...)

	if len(targetNodes) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(47215)
		return targetNodes[0] == runNode == true
	}() == true {
		__antithesis_instrumentation__.Notify(47216)
		args = append(args, "--self")
	} else {
		__antithesis_instrumentation__.Notify(47217)
		for _, target := range targetNodes {
			__antithesis_instrumentation__.Notify(47218)
			args = append(args, strconv.Itoa(target))
		}
	}
	__antithesis_instrumentation__.Notify(47214)
	return execCLI(ctx, h.t, h.c, runNode, args...)
}

func (h *decommTestHelper) recommission(
	ctx context.Context, targetNodes option.NodeListOption, runNode int, verbs ...string,
) (string, error) {
	__antithesis_instrumentation__.Notify(47219)
	args := []string{"node", "recommission"}
	args = append(args, verbs...)

	if len(targetNodes) == 1 && func() bool {
		__antithesis_instrumentation__.Notify(47221)
		return targetNodes[0] == runNode == true
	}() == true {
		__antithesis_instrumentation__.Notify(47222)
		args = append(args, "--self")
	} else {
		__antithesis_instrumentation__.Notify(47223)
		for _, target := range targetNodes {
			__antithesis_instrumentation__.Notify(47224)
			args = append(args, strconv.Itoa(target))
		}
	}
	__antithesis_instrumentation__.Notify(47220)
	return execCLI(ctx, h.t, h.c, runNode, args...)
}

func (h *decommTestHelper) expectColumn(
	column int, columnRegex []string, numRows, numCols int,
) [][]string {
	__antithesis_instrumentation__.Notify(47225)
	var res [][]string
	for r := 0; r < numRows; r++ {
		__antithesis_instrumentation__.Notify(47227)
		build := []string{}
		for c := 0; c < numCols; c++ {
			__antithesis_instrumentation__.Notify(47229)
			if c == column {
				__antithesis_instrumentation__.Notify(47230)
				build = append(build, columnRegex[r])
			} else {
				__antithesis_instrumentation__.Notify(47231)
				build = append(build, `.*`)
			}
		}
		__antithesis_instrumentation__.Notify(47228)
		res = append(res, build)
	}
	__antithesis_instrumentation__.Notify(47226)
	return res
}

func (h *decommTestHelper) expectCell(
	row, column int, regex string, numRows, numCols int,
) [][]string {
	__antithesis_instrumentation__.Notify(47232)
	var res [][]string
	for r := 0; r < numRows; r++ {
		__antithesis_instrumentation__.Notify(47234)
		build := []string{}
		for c := 0; c < numCols; c++ {
			__antithesis_instrumentation__.Notify(47236)
			if r == row && func() bool {
				__antithesis_instrumentation__.Notify(47237)
				return c == column == true
			}() == true {
				__antithesis_instrumentation__.Notify(47238)
				build = append(build, regex)
			} else {
				__antithesis_instrumentation__.Notify(47239)
				build = append(build, `.*`)
			}
		}
		__antithesis_instrumentation__.Notify(47235)
		res = append(res, build)
	}
	__antithesis_instrumentation__.Notify(47233)
	return res
}

func (h *decommTestHelper) expectIDsInStatusOut(ids []string, numCols int) [][]string {
	__antithesis_instrumentation__.Notify(47240)
	var res [][]string
	switch numCols {
	case len(statusHeader):
		__antithesis_instrumentation__.Notify(47243)
		res = append(res, statusHeader)
	case len(statusHeaderWithDecommission):
		__antithesis_instrumentation__.Notify(47244)
		res = append(res, statusHeaderWithDecommission)
	default:
		__antithesis_instrumentation__.Notify(47245)
		h.t.Fatalf(
			"Expected status output numCols to be one of %d or %d, found %d",
			len(statusHeader),
			len(statusHeaderWithDecommission),
			numCols,
		)
	}
	__antithesis_instrumentation__.Notify(47241)
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(47246)
		build := []string{id}
		for i := 0; i < numCols-1; i++ {
			__antithesis_instrumentation__.Notify(47248)
			build = append(build, `.*`)
		}
		__antithesis_instrumentation__.Notify(47247)
		res = append(res, build)
	}
	__antithesis_instrumentation__.Notify(47242)
	return res
}

func (h *decommTestHelper) blockFromRandNode(id int) {
	__antithesis_instrumentation__.Notify(47249)
	h.randNodeBlocklist[id] = struct{}{}
}

func (h *decommTestHelper) getRandNode() int {
	__antithesis_instrumentation__.Notify(47250)
	return h.getRandNodeOtherThan()
}

func (h *decommTestHelper) getRandNodeOtherThan(ids ...int) int {
	__antithesis_instrumentation__.Notify(47251)
	m := map[int]struct{}{}
	for _, id := range ids {
		__antithesis_instrumentation__.Notify(47255)
		m[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(47252)
	for id := range h.randNodeBlocklist {
		__antithesis_instrumentation__.Notify(47256)
		m[id] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(47253)
	if len(m) == len(h.nodeIDs) {
		__antithesis_instrumentation__.Notify(47257)
		h.t.Fatal("all nodes are blocked")
	} else {
		__antithesis_instrumentation__.Notify(47258)
	}
	__antithesis_instrumentation__.Notify(47254)
	for {
		__antithesis_instrumentation__.Notify(47259)
		id := h.nodeIDs[rand.Intn(len(h.nodeIDs))]
		if _, ok := m[id]; !ok {
			__antithesis_instrumentation__.Notify(47260)
			return id
		} else {
			__antithesis_instrumentation__.Notify(47261)
		}
	}
}

func execCLI(
	ctx context.Context, t test.Test, c cluster.Cluster, runNode int, extraArgs ...string,
) (string, error) {
	__antithesis_instrumentation__.Notify(47262)
	args := []string{"./cockroach"}
	args = append(args, extraArgs...)
	args = append(args, "--insecure")
	args = append(args, fmt.Sprintf("--port={pgport:%d}", runNode))
	result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(runNode), args...)
	t.L().Printf("%s\n", result.Stdout)
	return result.Stdout, err
}
