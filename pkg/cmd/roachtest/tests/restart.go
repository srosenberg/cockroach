package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func runRestart(ctx context.Context, t test.Test, c cluster.Cluster, downDuration time.Duration) {
	__antithesis_instrumentation__.Notify(50265)
	crdbNodes := c.Range(1, c.Spec().NodeCount)
	workloadNode := c.Node(1)
	const restartNode = 3

	t.Status("installing cockroach")
	c.Put(ctx, t.Cockroach(), "./cockroach", crdbNodes)
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--vmodule=raft_log_queue=3")
	c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), crdbNodes)

	t.Status("importing tpcc fixture")
	c.Run(ctx, workloadNode,
		"./cockroach workload fixtures import tpcc --warehouses=100 --fks=false --checks=false")

	t.Status("waiting for addsstable truncations")
	time.Sleep(11 * time.Minute)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(restartNode))

	c.Run(ctx, workloadNode, "./cockroach workload run tpcc --warehouses=100 "+
		fmt.Sprintf("--tolerate-errors --wait=false --duration=%s", downDuration))

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(restartNode))

	time.Sleep(15 * time.Second)

	start := timeutil.Now()
	restartNodeDB := c.Conn(ctx, t.L(), restartNode)
	if _, err := restartNodeDB.Exec(`SELECT count(*) FROM tpcc.order_line`); err != nil {
		__antithesis_instrumentation__.Notify(50267)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50268)
	}
	__antithesis_instrumentation__.Notify(50266)
	if took := timeutil.Since(start); took > downDuration {
		__antithesis_instrumentation__.Notify(50269)
		t.Fatalf(`expected to recover within %s took %s`, downDuration, took)
	} else {
		__antithesis_instrumentation__.Notify(50270)
		t.L().Printf(`connecting and query finished in %s`, took)
	}
}

func registerRestart(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50271)
	r.Add(registry.TestSpec{
		Name:    "restart/down-for-2m",
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(3),

		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50272)
			runRestart(ctx, t, c, 2*time.Minute)
		},
	})
}
