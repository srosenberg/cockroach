package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
)

type quitTest struct {
	t    test.Test
	c    cluster.Cluster
	args []string
	env  []string
}

func runQuitTransfersLeases(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	methodName string,
	method func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int),
) {
	__antithesis_instrumentation__.Notify(49936)
	q := quitTest{t: t, c: c}
	q.init(ctx)
	q.runTest(ctx, method)
}

func (q *quitTest) init(ctx context.Context) {
	q.args = []string{"--vmodule=store=1,replica=1,replica_proposal=1"}
	q.env = []string{"COCKROACH_SCAN_MAX_IDLE_TIME=5ms"}
	q.c.Put(ctx, q.t.Cockroach(), "./cockroach")
	settings := install.MakeClusterSettings(install.EnvOption(q.env))
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, q.args...)
	q.c.Start(ctx, q.t.L(), startOpts, settings)
}

func (q *quitTest) Fatal(args ...interface{}) {
	__antithesis_instrumentation__.Notify(49937)
	q.t.Fatal(args...)
}

func (q *quitTest) Fatalf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(49938)
	q.t.Fatalf(format, args...)
}

func (q *quitTest) runTest(
	ctx context.Context, method func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int),
) {
	__antithesis_instrumentation__.Notify(49939)
	q.waitForUpReplication(ctx)
	q.createRanges(ctx)
	q.setupIncrementalDrain(ctx)

	q.t.L().Printf("now running restart loop\n")
	for i := 0; i < 3; i++ {
		__antithesis_instrumentation__.Notify(49940)
		q.t.L().Printf("iteration %d\n", i)
		for nodeID := 1; nodeID <= q.c.Spec().NodeCount; nodeID++ {
			__antithesis_instrumentation__.Notify(49941)
			q.t.L().Printf("stopping node %d\n", nodeID)
			q.runWithTimeout(ctx, func(ctx context.Context) { __antithesis_instrumentation__.Notify(49944); method(ctx, q.t, q.c, nodeID) })
			__antithesis_instrumentation__.Notify(49942)
			q.runWithTimeout(ctx, func(ctx context.Context) { __antithesis_instrumentation__.Notify(49945); q.checkNoLeases(ctx, nodeID) })
			__antithesis_instrumentation__.Notify(49943)
			q.t.L().Printf("restarting node %d\n", nodeID)
			q.runWithTimeout(ctx, func(ctx context.Context) { __antithesis_instrumentation__.Notify(49946); q.restartNode(ctx, nodeID) })
		}
	}
}

func (q *quitTest) restartNode(ctx context.Context, nodeID int) {
	__antithesis_instrumentation__.Notify(49947)
	settings := install.MakeClusterSettings(install.EnvOption(q.env))
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, q.args...)
	q.c.Start(ctx, q.t.L(), startOpts, settings, q.c.Node(nodeID))

	q.t.L().Printf("waiting for readiness of node %d\n", nodeID)

	db := q.c.Conn(ctx, q.t.L(), nodeID)
	defer db.Close()
	if _, err := db.ExecContext(ctx, `TABLE crdb_internal.cluster_sessions`); err != nil {
		__antithesis_instrumentation__.Notify(49948)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49949)
	}
}

func (q *quitTest) waitForUpReplication(ctx context.Context) {
	__antithesis_instrumentation__.Notify(49950)
	db := q.c.Conn(ctx, q.t.L(), 1)
	defer db.Close()

	if _, err := db.ExecContext(ctx, `SET CLUSTER SETTING	kv.snapshot_rebalance.max_rate = '128MiB'`); err != nil {
		__antithesis_instrumentation__.Notify(49953)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49954)
	}
	__antithesis_instrumentation__.Notify(49951)

	err := retry.ForDuration(30*time.Second, func() error {
		__antithesis_instrumentation__.Notify(49955)
		q.t.L().Printf("waiting for up-replication\n")
		row := db.QueryRowContext(ctx, `SELECT min(array_length(replicas, 1)) FROM crdb_internal.ranges_no_leases`)
		minReplicas := 0
		if err := row.Scan(&minReplicas); err != nil {
			__antithesis_instrumentation__.Notify(49958)
			q.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49959)
		}
		__antithesis_instrumentation__.Notify(49956)
		if minReplicas < 3 {
			__antithesis_instrumentation__.Notify(49960)
			time.Sleep(time.Second)
			return errors.Newf("some ranges not up-replicated yet")
		} else {
			__antithesis_instrumentation__.Notify(49961)
		}
		__antithesis_instrumentation__.Notify(49957)
		return nil
	})
	__antithesis_instrumentation__.Notify(49952)
	if err != nil {
		__antithesis_instrumentation__.Notify(49962)
		q.Fatalf("cluster did not up-replicate: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(49963)
	}
}

func (q *quitTest) runWithTimeout(ctx context.Context, fn func(ctx context.Context)) {
	__antithesis_instrumentation__.Notify(49964)
	if err := contextutil.RunWithTimeout(ctx, "do", time.Minute, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49965)
		fn(ctx)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(49966)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49967)
	}
}

func (q *quitTest) setupIncrementalDrain(ctx context.Context) {
	__antithesis_instrumentation__.Notify(49968)
	db := q.c.Conn(ctx, q.t.L(), 1)
	defer db.Close()
	if _, err := db.ExecContext(ctx, `
SET CLUSTER SETTING server.shutdown.lease_transfer_wait = '10ms'`); err != nil {
		__antithesis_instrumentation__.Notify(49969)
		if strings.Contains(err.Error(), "unknown cluster setting") {
			__antithesis_instrumentation__.Notify(49970)

		} else {
			__antithesis_instrumentation__.Notify(49971)
			q.Fatal(err)
		}
	} else {
		__antithesis_instrumentation__.Notify(49972)
	}
}

func (q *quitTest) createRanges(ctx context.Context) {
	__antithesis_instrumentation__.Notify(49973)
	const numRanges = 500

	db := q.c.Conn(ctx, q.t.L(), 1)
	defer db.Close()
	if _, err := db.ExecContext(ctx, fmt.Sprintf(`
CREATE TABLE t(x, y, PRIMARY KEY(x)) AS SELECT @1, 1 FROM generate_series(1,%[1]d)`,
		numRanges)); err != nil {
		__antithesis_instrumentation__.Notify(49975)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49976)
	}
	__antithesis_instrumentation__.Notify(49974)

	for i := numRanges; i > 1; i -= 100 {
		__antithesis_instrumentation__.Notify(49977)
		q.t.L().Printf("creating %d ranges (%d-%d)...\n", numRanges, i, i-99)
		if _, err := db.ExecContext(ctx, fmt.Sprintf(`
ALTER TABLE t SPLIT AT TABLE generate_series(%[1]d,%[1]d-99,-1)`, i)); err != nil {
			__antithesis_instrumentation__.Notify(49978)
			q.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49979)
		}
	}
}

func (q *quitTest) checkNoLeases(ctx context.Context, nodeID int) {
	__antithesis_instrumentation__.Notify(49980)

	otherNodeID := 1 + nodeID%q.c.Spec().NodeCount

	if err := testutils.SucceedsSoonError(func() error {
		__antithesis_instrumentation__.Notify(49982)

		knownRanges := map[string]int{}

		invLeaseMap := map[int][]string{}
		for i := 1; i <= q.c.Spec().NodeCount; i++ {
			__antithesis_instrumentation__.Notify(49987)
			if i == nodeID {
				__antithesis_instrumentation__.Notify(49994)

				continue
			} else {
				__antithesis_instrumentation__.Notify(49995)
			}
			__antithesis_instrumentation__.Notify(49988)

			q.t.L().Printf("retrieving ranges for node %d\n", i)

			adminAddrs, err := q.c.InternalAdminUIAddr(ctx, q.t.L(), q.c.Node(otherNodeID))
			if err != nil {
				__antithesis_instrumentation__.Notify(49996)
				q.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49997)
			}
			__antithesis_instrumentation__.Notify(49989)
			result, err := q.c.RunWithDetailsSingleNode(ctx, q.t.L(), q.c.Node(otherNodeID),
				"curl", "-s", fmt.Sprintf("http://%s/_status/ranges/%d",
					adminAddrs[0], i))
			if err != nil {
				__antithesis_instrumentation__.Notify(49998)
				q.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49999)
			}
			__antithesis_instrumentation__.Notify(49990)

			type jsonOutput struct {
				Ranges []struct {
					State struct {
						State struct {
							Desc struct {
								RangeID string `json:"rangeId"`
							} `json:"desc"`
							Lease struct {
								Replica struct {
									NodeID int `json:"nodeId"`
								} `json:"replica"`
							} `json:"lease"`
						} `json:"state"`
					} `json:"state"`
				} `json:"ranges"`
			}
			var details jsonOutput
			if err := json.Unmarshal([]byte(result.Stdout), &details); err != nil {
				__antithesis_instrumentation__.Notify(50000)
				q.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50001)
			}
			__antithesis_instrumentation__.Notify(49991)

			if len(details.Ranges) == 0 {
				__antithesis_instrumentation__.Notify(50002)
				q.Fatal("expected some ranges from RPC, got none")
			} else {
				__antithesis_instrumentation__.Notify(50003)
			}
			__antithesis_instrumentation__.Notify(49992)

			var invalidLeases []string
			for _, r := range details.Ranges {
				__antithesis_instrumentation__.Notify(50004)

				if r.State.State.Lease.Replica.NodeID == 0 {
					__antithesis_instrumentation__.Notify(50007)
					q.Fatalf("expected a valid lease state, got %# v", pretty.Formatter(r))
				} else {
					__antithesis_instrumentation__.Notify(50008)
				}
				__antithesis_instrumentation__.Notify(50005)
				curLeaseHolder := knownRanges[r.State.State.Desc.RangeID]
				if r.State.State.Lease.Replica.NodeID == nodeID {
					__antithesis_instrumentation__.Notify(50009)

					invalidLeases = append(invalidLeases, r.State.State.Desc.RangeID)
				} else {
					__antithesis_instrumentation__.Notify(50010)

					curLeaseHolder = r.State.State.Lease.Replica.NodeID
				}
				__antithesis_instrumentation__.Notify(50006)
				knownRanges[r.State.State.Desc.RangeID] = curLeaseHolder
			}
			__antithesis_instrumentation__.Notify(49993)
			if len(invalidLeases) > 0 {
				__antithesis_instrumentation__.Notify(50011)
				invLeaseMap[i] = invalidLeases
			} else {
				__antithesis_instrumentation__.Notify(50012)
			}
		}
		__antithesis_instrumentation__.Notify(49983)

		var leftOver []string
		for r, n := range knownRanges {
			__antithesis_instrumentation__.Notify(50013)
			if n == 0 {
				__antithesis_instrumentation__.Notify(50014)
				leftOver = append(leftOver, r)
			} else {
				__antithesis_instrumentation__.Notify(50015)
			}
		}
		__antithesis_instrumentation__.Notify(49984)
		if len(leftOver) > 0 {
			__antithesis_instrumentation__.Notify(50016)
			q.Fatalf("(1) ranges with no lease outside of node %d: %# v", nodeID, pretty.Formatter(leftOver))
		} else {
			__antithesis_instrumentation__.Notify(50017)
		}
		__antithesis_instrumentation__.Notify(49985)

		if len(invLeaseMap) > 0 {
			__antithesis_instrumentation__.Notify(50018)
			err := errors.Newf(
				"(2) ranges with remaining leases on node %d, per node: %# v",
				nodeID, pretty.Formatter(invLeaseMap))
			q.t.L().Printf("condition failed: %v\n", err)
			q.t.L().Printf("retrying until SucceedsSoon has enough...\n")
			return err
		} else {
			__antithesis_instrumentation__.Notify(50019)
		}
		__antithesis_instrumentation__.Notify(49986)
		return nil
	}); err != nil {
		__antithesis_instrumentation__.Notify(50020)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50021)
	}
	__antithesis_instrumentation__.Notify(49981)

	db := q.c.Conn(ctx, q.t.L(), otherNodeID)
	defer db.Close()

	if _, err := db.ExecContext(ctx, `UPDATE t SET y = y + 1`); err != nil {
		__antithesis_instrumentation__.Notify(50022)
		q.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50023)
	}
}

func registerQuitTransfersLeases(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50024)
	registerTest := func(name, minver string, method func(context.Context, test.Test, cluster.Cluster, int)) {
		__antithesis_instrumentation__.Notify(50029)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("transfer-leases/%s", name),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(3),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(50030)
				runQuitTransfersLeases(ctx, t, c, name, method)
			},
		})
	}
	__antithesis_instrumentation__.Notify(50025)

	registerTest("signal", "v19.2.0", func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int) {
		__antithesis_instrumentation__.Notify(50031)
		stopOpts := option.DefaultStopOpts()
		stopOpts.RoachprodOpts.Sig = 15
		stopOpts.RoachprodOpts.Wait = true
		c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))

	})
	__antithesis_instrumentation__.Notify(50026)

	registerTest("quit", "v19.2.0", func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int) {
		__antithesis_instrumentation__.Notify(50032)
		_ = runQuit(ctx, t, c, nodeID)
	})
	__antithesis_instrumentation__.Notify(50027)

	registerTest("drain", "v20.1.0", func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int) {
		__antithesis_instrumentation__.Notify(50033)
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(nodeID),
			"./cockroach", "node", "drain", "--insecure", "--logtostderr=INFO",
			fmt.Sprintf("--port={pgport:%d}", nodeID),
		)
		t.L().Printf("cockroach node drain:\n%s\n", result.Stdout+result.Stdout)
		if err != nil {
			__antithesis_instrumentation__.Notify(50035)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50036)
		}
		__antithesis_instrumentation__.Notify(50034)

		stopOpts := option.DefaultStopOpts()
		stopOpts.RoachprodOpts.Sig = 1
		c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))

		stopOpts = option.DefaultStopOpts()
		stopOpts.RoachprodOpts.Sig = 9
		stopOpts.RoachprodOpts.Wait = true
		c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))
	})
	__antithesis_instrumentation__.Notify(50028)

	registerTest("drain-other-node", "v22.1.0", func(ctx context.Context, t test.Test, c cluster.Cluster, nodeID int) {
		__antithesis_instrumentation__.Notify(50037)

		otherNodeID := (nodeID % c.Spec().NodeCount) + 1
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(otherNodeID),
			"./cockroach", "node", "drain", "--insecure", "--logtostderr=INFO",
			fmt.Sprintf("--port={pgport:%d}", otherNodeID),
			fmt.Sprintf("%d", nodeID),
		)
		t.L().Printf("cockroach node drain:\n%s\n", result.Stdout+result.Stderr)
		if err != nil {
			__antithesis_instrumentation__.Notify(50039)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50040)
		}
		__antithesis_instrumentation__.Notify(50038)

		stopOpts := option.DefaultStopOpts()
		stopOpts.RoachprodOpts.Sig = 1
		c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))

		stopOpts.RoachprodOpts.Sig = 9
		stopOpts.RoachprodOpts.Wait = true
		c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))
	})
}

func runQuit(
	ctx context.Context, t test.Test, c cluster.Cluster, nodeID int, extraArgs ...string,
) []byte {
	__antithesis_instrumentation__.Notify(50041)
	args := append([]string{
		"./cockroach", "quit", "--insecure", "--logtostderr=INFO",
		fmt.Sprintf("--port={pgport:%d}", nodeID)},
		extraArgs...)
	result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(nodeID), args...)
	output := result.Stdout + result.Stderr
	t.L().Printf("cockroach quit:\n%s\n", output)
	if err != nil {
		__antithesis_instrumentation__.Notify(50043)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50044)
	}
	__antithesis_instrumentation__.Notify(50042)
	stopOpts := option.DefaultStopOpts()
	stopOpts.RoachprodOpts.Sig = 0
	stopOpts.RoachprodOpts.Wait = true
	c.Stop(ctx, t.L(), stopOpts, c.Node(nodeID))

	return []byte(output)
}

func registerQuitAllNodes(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50045)

	r.Add(registry.TestSpec{
		Name:    "quit-all-nodes",
		Owner:   registry.OwnerServer,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50046)
			q := quitTest{t: t, c: c}

			q.init(ctx)

			q.waitForUpReplication(ctx)

			q.runWithTimeout(ctx, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(50052)
				_ = runQuit(ctx, q.t, q.c, 5, "--drain-wait=1h")
			})
			__antithesis_instrumentation__.Notify(50047)

			q.runWithTimeout(ctx, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(50053)
				_ = runQuit(ctx, q.t, q.c, 4, "--drain-wait=4s")
			})
			__antithesis_instrumentation__.Notify(50048)
			q.runWithTimeout(ctx, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(50054)
				_ = runQuit(ctx, q.t, q.c, 3, "--drain-wait=4s")
			})
			__antithesis_instrumentation__.Notify(50049)

			q.runWithTimeout(ctx, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(50055)
				expectHardShutdown(ctx, q.t, runQuit(ctx, q.t, q.c, 2, "--drain-wait=4s"))
			})
			__antithesis_instrumentation__.Notify(50050)
			q.runWithTimeout(ctx, func(ctx context.Context) {
				__antithesis_instrumentation__.Notify(50056)
				expectHardShutdown(ctx, q.t, runQuit(ctx, q.t, q.c, 1, "--drain-wait=4s"))
			})
			__antithesis_instrumentation__.Notify(50051)

			settings := install.MakeClusterSettings(install.EnvOption(q.env))
			startOpts := option.DefaultStartOpts()
			startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, q.args...)
			q.c.Start(ctx, t.L(), startOpts, settings)
		},
	})
}

func expectHardShutdown(ctx context.Context, t test.Test, cmdOut []byte) {
	__antithesis_instrumentation__.Notify(50057)
	if !strings.Contains(string(cmdOut), "drain did not complete successfully") {
		__antithesis_instrumentation__.Notify(50058)
		t.Fatalf("expected 'drain did not complete successfully' in quit output, got:\n%s", cmdOut)
	} else {
		__antithesis_instrumentation__.Notify(50059)
	}
}
