package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func registerReplicaGC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50177)
	for _, restart := range []bool{true, false} {
		__antithesis_instrumentation__.Notify(50178)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("replicagc-changed-peers/restart=%t", restart),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(6),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(50179)
				runReplicaGCChangedPeers(ctx, t, c, restart)
			},
		})
	}
}

var deadNodeAttr = "deadnode"

func runReplicaGCChangedPeers(
	ctx context.Context, t test.Test, c cluster.Cluster, withRestart bool,
) {
	__antithesis_instrumentation__.Notify(50180)
	if c.Spec().NodeCount != 6 {
		__antithesis_instrumentation__.Notify(50186)
		t.Fatal("test needs to be run with 6 nodes")
	} else {
		__antithesis_instrumentation__.Notify(50187)
	}
	__antithesis_instrumentation__.Notify(50181)

	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(1))
	settings := install.MakeClusterSettings(install.EnvOption([]string{"COCKROACH_SCAN_MAX_IDLE_TIME=5ms"}))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Range(1, 3))

	h := &replicagcTestHelper{c: c, t: t}

	t.Status("waiting for full replication")
	h.waitForFullReplication(ctx)

	c.Run(ctx, c.Node(1), "./workload init kv {pgurl:1} --splits 100")

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(3))

	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Range(4, 6))

	if err := h.decommission(ctx, c.Range(1, 3), 2, "--wait=none"); err != nil {
		__antithesis_instrumentation__.Notify(50188)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50189)
	}
	__antithesis_instrumentation__.Notify(50182)

	t.Status("waiting for zero replicas on n1")
	h.waitForZeroReplicas(ctx, 1)

	t.Status("waiting for zero replicas on n2")
	h.waitForZeroReplicas(ctx, 2)

	t.Status("waiting for zero replicas on n3")
	waitForZeroReplicasOnN3(ctx, t, c.Conn(ctx, t.L(), 1))

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Range(1, 2))

	h.isolateDeadNodes(ctx, 4)

	if err := h.recommission(ctx, c.Range(1, 3), 4); err != nil {
		__antithesis_instrumentation__.Notify(50190)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50191)
	}
	__antithesis_instrumentation__.Notify(50183)

	if withRestart {
		__antithesis_instrumentation__.Notify(50192)

		c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Range(4, 6))
		c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Range(4, 6))
	} else {
		__antithesis_instrumentation__.Notify(50193)
	}
	__antithesis_instrumentation__.Notify(50184)

	internalAddrs, err := c.InternalAddr(ctx, t.L(), c.Node(4))
	if err != nil {
		__antithesis_instrumentation__.Notify(50194)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50195)
	}
	__antithesis_instrumentation__.Notify(50185)
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--join="+internalAddrs[0], "--attrs="+deadNodeAttr, "--vmodule=raft=5,replicate_queue=5,allocator=5")
	c.Start(ctx, t.L(), startOpts, settings, c.Node(3))

	h.waitForZeroReplicas(ctx, 3)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, 2))
}

type replicagcTestHelper struct {
	t test.Test
	c cluster.Cluster
}

func (h *replicagcTestHelper) waitForFullReplication(ctx context.Context) {
	__antithesis_instrumentation__.Notify(50196)
	db := h.c.Conn(ctx, h.t.L(), 1)
	defer func() {
		__antithesis_instrumentation__.Notify(50198)
		_ = db.Close()
	}()
	__antithesis_instrumentation__.Notify(50197)

	for {
		__antithesis_instrumentation__.Notify(50199)
		var fullReplicated bool
		if err := db.QueryRow(

			"SELECT min(array_length(replicas, 1)) >= 3 FROM crdb_internal.ranges",
		).Scan(&fullReplicated); err != nil {
			__antithesis_instrumentation__.Notify(50202)
			h.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50203)
		}
		__antithesis_instrumentation__.Notify(50200)
		if fullReplicated {
			__antithesis_instrumentation__.Notify(50204)
			break
		} else {
			__antithesis_instrumentation__.Notify(50205)
		}
		__antithesis_instrumentation__.Notify(50201)
		time.Sleep(time.Second)
	}
}

func (h *replicagcTestHelper) waitForZeroReplicas(ctx context.Context, targetNode int) {
	__antithesis_instrumentation__.Notify(50206)
	db := h.c.Conn(ctx, h.t.L(), targetNode)
	defer func() {
		__antithesis_instrumentation__.Notify(50209)
		_ = db.Close()
	}()
	__antithesis_instrumentation__.Notify(50207)

	var n = 0
	for tBegin := timeutil.Now(); timeutil.Since(tBegin) < 5*time.Minute; time.Sleep(5 * time.Second) {
		__antithesis_instrumentation__.Notify(50210)
		n = h.numReplicas(ctx, db, targetNode)
		if n == 0 {
			__antithesis_instrumentation__.Notify(50211)
			break
		} else {
			__antithesis_instrumentation__.Notify(50212)
		}
	}
	__antithesis_instrumentation__.Notify(50208)
	if n != 0 {
		__antithesis_instrumentation__.Notify(50213)
		h.t.Fatalf("replica count on n%d didn't drop to zero: %d", targetNode, n)
	} else {
		__antithesis_instrumentation__.Notify(50214)
	}
}

func (h *replicagcTestHelper) numReplicas(ctx context.Context, db *gosql.DB, targetNode int) int {
	__antithesis_instrumentation__.Notify(50215)
	var n int
	if err := db.QueryRowContext(
		ctx,
		`SELECT value FROM crdb_internal.node_metrics WHERE name = 'replicas'`,
	).Scan(&n); err != nil {
		__antithesis_instrumentation__.Notify(50217)
		h.t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50218)
	}
	__antithesis_instrumentation__.Notify(50216)
	h.t.L().Printf("found %d replicas found on n%d\n", n, targetNode)
	return n
}

func (h *replicagcTestHelper) decommission(
	ctx context.Context, targetNodes option.NodeListOption, runNode int, verbs ...string,
) error {
	__antithesis_instrumentation__.Notify(50219)
	args := []string{"node", "decommission"}
	args = append(args, verbs...)

	for _, target := range targetNodes {
		__antithesis_instrumentation__.Notify(50221)
		args = append(args, strconv.Itoa(target))
	}
	__antithesis_instrumentation__.Notify(50220)
	_, err := execCLI(ctx, h.t, h.c, runNode, args...)
	return err
}

func (h *replicagcTestHelper) recommission(
	ctx context.Context, targetNodes option.NodeListOption, runNode int, verbs ...string,
) error {
	__antithesis_instrumentation__.Notify(50222)
	args := []string{"node", "recommission"}
	args = append(args, verbs...)
	for _, target := range targetNodes {
		__antithesis_instrumentation__.Notify(50224)
		args = append(args, strconv.Itoa(target))
	}
	__antithesis_instrumentation__.Notify(50223)
	_, err := execCLI(ctx, h.t, h.c, runNode, args...)
	return err
}

func (h *replicagcTestHelper) isolateDeadNodes(ctx context.Context, runNode int) {
	__antithesis_instrumentation__.Notify(50225)
	db := h.c.Conn(ctx, h.t.L(), runNode)
	defer func() {
		__antithesis_instrumentation__.Notify(50227)
		_ = db.Close()
	}()
	__antithesis_instrumentation__.Notify(50226)

	for _, change := range []string{
		"RANGE default", "RANGE meta", "RANGE system", "RANGE liveness", "DATABASE system", "TABLE system.jobs",
	} {
		__antithesis_instrumentation__.Notify(50228)
		stmt := `ALTER ` + change + ` CONFIGURE ZONE = 'constraints: {"-` + deadNodeAttr + `"}'`
		h.t.L().Printf(stmt + "\n")
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			__antithesis_instrumentation__.Notify(50229)
			h.t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50230)
		}
	}
}

func waitForZeroReplicasOnN3(ctx context.Context, t test.Test, db *gosql.DB) {
	__antithesis_instrumentation__.Notify(50231)
	if err := retry.ForDuration(5*time.Minute, func() error {
		__antithesis_instrumentation__.Notify(50232)
		const q = `select range_id, replicas from crdb_internal.ranges_no_leases where replicas @> ARRAY[3];`
		rows, err := db.QueryContext(ctx, q)
		if err != nil {
			__antithesis_instrumentation__.Notify(50237)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50238)
		}
		__antithesis_instrumentation__.Notify(50233)
		m := make(map[int64]string)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50239)
			var rangeID int64
			var replicas string
			if err := rows.Scan(&rangeID, &replicas); err != nil {
				__antithesis_instrumentation__.Notify(50241)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50242)
			}
			__antithesis_instrumentation__.Notify(50240)
			m[rangeID] = replicas
		}
		__antithesis_instrumentation__.Notify(50234)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(50243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50244)
		}
		__antithesis_instrumentation__.Notify(50235)
		if len(m) == 0 {
			__antithesis_instrumentation__.Notify(50245)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(50246)
		}
		__antithesis_instrumentation__.Notify(50236)
		return errors.Errorf("ranges remained on n3 (according to meta2): %+v", m)
	}); err != nil {
		__antithesis_instrumentation__.Notify(50247)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50248)
	}
}
