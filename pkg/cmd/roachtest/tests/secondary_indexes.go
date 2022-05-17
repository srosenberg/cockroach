package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/stretchr/testify/require"
)

func runIndexUpgrade(
	ctx context.Context, t test.Test, c cluster.Cluster, predecessorVersion string,
) {
	__antithesis_instrumentation__.Notify(50765)
	firstExpected := [][]int{
		{2, 3, 4},
		{6, 7, 8},
		{10, 11, 12},
		{14, 15, 17},
	}
	secondExpected := [][]int{
		{2, 3, 4},
		{6, 7, 8},
		{10, 11, 12},
		{14, 15, 17},
		{21, 25, 25},
	}

	roachNodes := c.All()

	const mainVersion = ""
	u := newVersionUpgradeTest(c,
		uploadAndStart(roachNodes, predecessorVersion),
		waitForUpgradeStep(roachNodes),

		createDataStep(),

		binaryUpgradeStep(c.Node(1), mainVersion),

		modifyData(1,
			`INSERT INTO t VALUES (13, 14, 15, 16)`,
			`UPDATE t SET w = 17 WHERE y = 14`,
		),

		verifyTableData(1, firstExpected),
		verifyTableData(2, firstExpected),
		verifyTableData(3, firstExpected),

		binaryUpgradeStep(c.Node(2), mainVersion),
		binaryUpgradeStep(c.Node(3), mainVersion),

		allowAutoUpgradeStep(1),
		waitForUpgradeStep(roachNodes),

		modifyData(1,
			`INSERT INTO t VALUES (20, 21, 22, 23)`,
			`UPDATE t SET w = 25, z = 25 WHERE y = 21`,
		),

		verifyTableData(1, secondExpected),
		verifyTableData(2, secondExpected),
		verifyTableData(3, secondExpected),
	)

	u.run(ctx, t)
}

func createDataStep() versionStep {
	__antithesis_instrumentation__.Notify(50766)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50767)
		conn := u.conn(ctx, t, 1)
		if _, err := conn.Exec(`
CREATE TABLE t (
	x INT PRIMARY KEY, y INT, z INT, w INT,
	INDEX i (y) STORING (z, w),
	FAMILY (x), FAMILY (y), FAMILY (z), FAMILY (w)
);
INSERT INTO t VALUES (1, 2, 3, 4), (5, 6, 7, 8), (9, 10, 11, 12);
`); err != nil {
			__antithesis_instrumentation__.Notify(50768)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50769)
		}
	}
}

func modifyData(node int, sql ...string) versionStep {
	__antithesis_instrumentation__.Notify(50770)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50771)

		conn := u.conn(ctx, t, node)
		for _, s := range sql {
			__antithesis_instrumentation__.Notify(50772)
			if _, err := conn.Exec(s); err != nil {
				__antithesis_instrumentation__.Notify(50773)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50774)
			}
		}
	}
}

func verifyTableData(node int, expected [][]int) versionStep {
	__antithesis_instrumentation__.Notify(50775)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50776)
		conn := u.conn(ctx, t, node)
		rows, err := conn.Query(`SELECT y, z, w FROM t@i ORDER BY y`)
		if err != nil {
			__antithesis_instrumentation__.Notify(50778)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50779)
		}
		__antithesis_instrumentation__.Notify(50777)
		var y, z, w int
		count := 0
		for ; rows.Next(); count++ {
			__antithesis_instrumentation__.Notify(50780)
			if err := rows.Scan(&y, &z, &w); err != nil {
				__antithesis_instrumentation__.Notify(50782)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50783)
			}
			__antithesis_instrumentation__.Notify(50781)
			found := []int{y, z, w}
			require.Equal(t, found, expected[count])
		}
	}
}

func registerSecondaryIndexesMultiVersionCluster(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50784)
	r.Add(registry.TestSpec{
		Name:    "schemachange/secondary-index-multi-version",
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50785)
			predV, err := PredecessorVersion(*t.BuildVersion())
			if err != nil {
				__antithesis_instrumentation__.Notify(50787)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50788)
			}
			__antithesis_instrumentation__.Notify(50786)
			runIndexUpgrade(ctx, t, c, predV)
		},
	})
}
