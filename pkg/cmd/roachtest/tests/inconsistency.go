package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

func registerInconsistency(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48526)
	r.Add(registry.TestSpec{
		Name:    "inconsistency",
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(3),
		Run:     runInconsistency,
	})
}

func runInconsistency(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(48527)
	startOps := option.DefaultStartOpts()

	nodes := c.Range(1, 3)
	c.Put(ctx, t.Cockroach(), "./cockroach", nodes)
	c.Start(ctx, t.L(), startOps, install.MakeClusterSettings(), nodes)

	{
		__antithesis_instrumentation__.Notify(48537)
		db := c.Conn(ctx, t.L(), 1)

		_, err := db.ExecContext(ctx, `SET CLUSTER SETTING server.consistency_check.interval = '0'`)
		require.NoError(t, err)
		err = WaitFor3XReplication(ctx, t, db)
		require.NoError(t, err)
		_, db = db.Close(), nil
	}
	__antithesis_instrumentation__.Notify(48528)

	stopOpts := option.DefaultStopOpts()
	stopOpts.RoachprodOpts.Wait = false
	stopOpts.RoachprodOpts.Sig = 2
	c.Stop(ctx, t.L(), stopOpts, nodes)
	stopOpts.RoachprodOpts.Wait = true
	c.Stop(ctx, t.L(), stopOpts, nodes)

	c.Run(ctx, c.Node(1), "./cockroach debug pebble db set {store-dir} "+
		"hex:016b1202000174786e2d0000000000000000000000000000000000 "+
		"hex:120408001000180020002800322a0a10000000000000000000000000000000001a1266616b65207472616e73616374696f6e20302a004a00")

	m := c.NewMonitor(ctx)

	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--vmodule=consistency_queue=5,replica_consistency=5,queue=5")
	c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), nodes)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(48538)
		select {
		case <-time.After(5 * time.Minute):
			__antithesis_instrumentation__.Notify(48540)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(48541)
		}
		__antithesis_instrumentation__.Notify(48539)
		return nil
	})
	__antithesis_instrumentation__.Notify(48529)

	time.Sleep(10 * time.Second)

	{
		__antithesis_instrumentation__.Notify(48542)
		db := c.Conn(ctx, t.L(), 2)
		_, err := db.ExecContext(ctx, `SET CLUSTER SETTING server.consistency_check.interval = '10ms'`)
		if err != nil {
			__antithesis_instrumentation__.Notify(48544)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48545)
		}
		__antithesis_instrumentation__.Notify(48543)
		_ = db.Close()
	}
	__antithesis_instrumentation__.Notify(48530)

	if err := m.WaitE(); err == nil {
		__antithesis_instrumentation__.Notify(48546)
		t.Fatal("expected a node to crash")
	} else {
		__antithesis_instrumentation__.Notify(48547)
	}
	__antithesis_instrumentation__.Notify(48531)

	time.Sleep(20 * time.Second)

	db := c.Conn(ctx, t.L(), 2)
	rows, err := db.Query(`SELECT node_id FROM crdb_internal.gossip_nodes WHERE is_live = false;`)
	if err != nil {
		__antithesis_instrumentation__.Notify(48548)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48549)
	}
	__antithesis_instrumentation__.Notify(48532)
	var ids []int
	for rows.Next() {
		__antithesis_instrumentation__.Notify(48550)
		var id int
		if err := rows.Scan(&id); err != nil {
			__antithesis_instrumentation__.Notify(48552)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48553)
		}
		__antithesis_instrumentation__.Notify(48551)
		ids = append(ids, id)
	}
	__antithesis_instrumentation__.Notify(48533)
	if err := rows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(48554)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48555)
	}
	__antithesis_instrumentation__.Notify(48534)
	if len(ids) != 1 {
		__antithesis_instrumentation__.Notify(48556)
		t.Fatalf("expected one dead NodeID, got %v", ids)
	} else {
		__antithesis_instrumentation__.Notify(48557)
	}
	__antithesis_instrumentation__.Notify(48535)
	const expr = "this.node.is.terminating.because.a.replica.inconsistency.was.detected"
	c.Run(ctx, c.Node(1), "grep "+
		expr+" "+"{log-dir}/cockroach.log")

	if err := c.StartE(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(1)); err == nil {
		__antithesis_instrumentation__.Notify(48558)

		t.Fatalf("node restart should have failed")
	} else {
		__antithesis_instrumentation__.Notify(48559)
	}
	__antithesis_instrumentation__.Notify(48536)

	c.Wipe(ctx, c.Node(1))
}
