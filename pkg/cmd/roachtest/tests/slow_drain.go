package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

func registerSlowDrain(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50840)
	numNodes := 6
	duration := time.Minute

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("slow-drain/duration=%s", duration),
		Owner:   registry.OwnerServer,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50841)
			runSlowDrain(ctx, t, c, duration)
		},
	})
}

func runSlowDrain(ctx context.Context, t test.Test, c cluster.Cluster, duration time.Duration) {
	__antithesis_instrumentation__.Notify(50842)
	const (
		numNodes          = 6
		pinnedNodeID      = 1
		replicationFactor = 5
	)

	var verboseStoreLogRe = regexp.MustCompile("failed to transfer lease")

	err := c.PutE(ctx, t.L(), t.Cockroach(), "./cockroach", c.All())
	require.NoError(t, err)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

	run := func(stmt string) {
		__antithesis_instrumentation__.Notify(50846)
		db := c.Conn(ctx, t.L(), pinnedNodeID)
		defer db.Close()

		_, err = db.ExecContext(ctx, stmt)
		require.NoError(t, err)

		t.L().Printf("run: %s\n", stmt)
	}

	{
		__antithesis_instrumentation__.Notify(50847)
		db := c.Conn(ctx, t.L(), pinnedNodeID)
		defer db.Close()

		run(fmt.Sprintf(`ALTER RANGE default CONFIGURE ZONE USING num_replicas=%d`, replicationFactor))
		run(fmt.Sprintf(`ALTER DATABASE system CONFIGURE ZONE USING num_replicas=%d`, replicationFactor))

		err := WaitForReplication(ctx, t, db, replicationFactor)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(50843)

	m := c.NewMonitor(ctx)
	for nodeID := 2; nodeID <= numNodes; nodeID++ {
		__antithesis_instrumentation__.Notify(50848)
		id := nodeID
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(50849)
			drain := func(id int) error {
				__antithesis_instrumentation__.Notify(50851)
				t.Status(fmt.Sprintf("draining node %d", id))
				return c.RunE(ctx,
					c.Node(id),
					fmt.Sprintf("./cockroach node drain %d --insecure --drain-wait=%s", id, duration.String()),
				)
			}
			__antithesis_instrumentation__.Notify(50850)
			return drain(id)
		})
	}
	__antithesis_instrumentation__.Notify(50844)

	time.Sleep(30 * time.Second)

	t.Status("checking for stalling drain logging...")
	found := false
	for nodeID := 2; nodeID <= numNodes; nodeID++ {
		__antithesis_instrumentation__.Notify(50852)
		if err := c.RunE(ctx, c.Node(nodeID),
			fmt.Sprintf("grep -q '%s' logs/cockroach.log", verboseStoreLogRe),
		); err == nil {
			__antithesis_instrumentation__.Notify(50853)
			found = true
		} else {
			__antithesis_instrumentation__.Notify(50854)
		}
	}
	__antithesis_instrumentation__.Notify(50845)
	require.True(t, found)
	t.Status("log messages found")

	t.Status("waiting for the drain timeout to elapse...")
	err = m.WaitE()
	require.Error(t, err)
}
