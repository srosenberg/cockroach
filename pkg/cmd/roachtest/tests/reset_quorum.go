package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/stretchr/testify/require"
)

func runResetQuorum(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(50249)
	skip.WithIssue(t, 58165)

	c.Put(ctx, t.Cockroach(), "./cockroach")

	settings := install.MakeClusterSettings(install.EnvOption([]string{"COCKROACH_SCAN_MAX_IDLE_TIME=5ms"}))
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--attrs=A")
	c.Start(ctx, t.L(), startOpts, settings, c.Range(1, 5))
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	rows, err := db.QueryContext(ctx, `SELECT target FROM crdb_internal.zones`)
	require.NoError(t, err)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(50253)
		var target string
		require.NoError(t, rows.Scan(&target))
		_, err = db.ExecContext(ctx, `ALTER `+target+` CONFIGURE ZONE USING constraints = '{"-B"}'`)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(50250)
	require.NoError(t, rows.Err())

	startOpts = option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--attrs=B")
	c.Start(ctx, t.L(), startOpts, settings, c.Range(6, 8))
	_, err = db.Exec(`CREATE TABLE lostrange (id INT PRIMARY KEY, v STRING)`)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO lostrange VALUES(1, 'foo')`)
	require.NoError(t, err)

	_, err = db.Exec(`ALTER TABLE lostrange CONFIGURE ZONE USING constraints = '{"+B"}'`)
	require.NoError(t, err)

	var lostRangeIDs map[int64]struct{}
	for i := 0; i < 100; i++ {
		__antithesis_instrumentation__.Notify(50254)
		lostRangeIDs = map[int64]struct{}{}
		rows, err := db.QueryContext(ctx, `
SELECT
	*
FROM
	[
		SELECT
			range_id, table_name, unnest(replicas) AS r
		FROM
			crdb_internal.ranges_no_leases
	]
WHERE
	(r IN (6, 7, 8)) -- intentionally do not exclude lostrange (to populate lostRangeIDs)
OR
	(r NOT IN (6, 7, 8) AND table_name = 'lostrange')
`)
		require.NoError(t, err)
		var buf strings.Builder
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50257)
			var rangeID int64
			var tableName string
			var storeID int
			require.NoError(t, rows.Scan(&rangeID, &tableName, &storeID))
			if tableName == "lostrange" && func() bool {
				__antithesis_instrumentation__.Notify(50258)
				return storeID >= 6 == true
			}() == true {
				__antithesis_instrumentation__.Notify(50259)
				lostRangeIDs[rangeID] = struct{}{}
			} else {
				__antithesis_instrumentation__.Notify(50260)
				fmt.Fprintf(&buf, "r%d still has a replica on s%d (table %q)\n", rangeID, storeID, tableName)
			}
		}
		__antithesis_instrumentation__.Notify(50255)
		require.NoError(t, rows.Err())
		if buf.Len() == 0 {
			__antithesis_instrumentation__.Notify(50261)
			break
		} else {
			__antithesis_instrumentation__.Notify(50262)
		}
		__antithesis_instrumentation__.Notify(50256)
		t.L().Printf("still waiting:\n" + buf.String())
		time.Sleep(5 * time.Second)
	}
	__antithesis_instrumentation__.Notify(50251)

	require.NotEmpty(t, lostRangeIDs)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Range(6, 8))
	c.Wipe(ctx, c.Range(6, 8))

	_, err = db.QueryContext(ctx, `SET statement_timeout = '15s'; SELECT * FROM lostrange;`)
	require.Error(t, err)
	t.L().Printf("table is now unavailable, as planned")

	const nodeID = 1
	for rangeID := range lostRangeIDs {
		__antithesis_instrumentation__.Notify(50263)
		c.Run(ctx, c.Node(nodeID), "./cockroach", "debug", "reset-quorum",
			fmt.Sprint(rangeID), "--insecure",
		)
	}
	__antithesis_instrumentation__.Notify(50252)

	var n int
	err = db.QueryRowContext(
		ctx, `SET statement_timeout = '120s'; SELECT count(*) FROM lostrange;`,
	).Scan(&n)
	require.NoError(t, err)
	require.Zero(t, n)

	for rangeID := range lostRangeIDs {
		__antithesis_instrumentation__.Notify(50264)
		var actNodeID int32

		err = db.QueryRowContext(ctx,
			`
SELECT
	r
FROM
	[
		SELECT
			range_id, unnest(replicas) AS r
		FROM
			crdb_internal.ranges_no_leases
	]
WHERE
	range_id = $1
`, rangeID).Scan(&actNodeID)
		require.NoError(t, err)
		require.EqualValues(t, nodeID, actNodeID)
	}
}
