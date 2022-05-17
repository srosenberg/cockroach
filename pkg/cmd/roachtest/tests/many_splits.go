package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

func runManySplits(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(49248)
	c.Put(ctx, t.Cockroach(), "./cockroach")
	settings := install.MakeClusterSettings()
	settings.Env = append(settings.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings)

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	err := WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	m := c.NewMonitor(ctx, c.All())
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(49250)
		const numRanges = 2000
		t.L().Printf("creating %d ranges...", numRanges)
		if _, err := db.ExecContext(ctx, fmt.Sprintf(`
			CREATE TABLE t(x, PRIMARY KEY(x)) AS TABLE generate_series(1,%[1]d);
            ALTER TABLE t SPLIT AT TABLE generate_series(1,%[1]d);
		`, numRanges)); err != nil {
			__antithesis_instrumentation__.Notify(49252)
			return err
		} else {
			__antithesis_instrumentation__.Notify(49253)
		}
		__antithesis_instrumentation__.Notify(49251)
		return nil
	})
	__antithesis_instrumentation__.Notify(49249)
	m.Wait()
}
