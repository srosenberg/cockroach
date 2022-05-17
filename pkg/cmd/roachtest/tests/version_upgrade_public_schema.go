package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

func registerVersionUpgradePublicSchema(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52373)
	r.Add(registry.TestSpec{
		Name:    "versionupgrade/publicschema",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52374)
			runVersionUpgradePublicSchema(ctx, t, c, *t.BuildVersion())
		},
	})
}

const loadNode = 1

func runVersionUpgradePublicSchema(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(52375)
	predecessorVersion, err := PredecessorVersion(buildVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(52377)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52378)
	}
	__antithesis_instrumentation__.Notify(52376)

	const currentVersion = ""

	steps := []versionStep{
		resetStep(),
		uploadAndStart(c.All(), predecessorVersion),
		waitForUpgradeStep(c.All()),

		preventAutoUpgradeStep(1),
	}

	steps = append(
		steps,
		createDatabaseStep("test1"),

		tryReparentingDatabase(false, ""),

		binaryUpgradeStep(c.Node(3), currentVersion),
		binaryUpgradeStep(c.Node(1), currentVersion),

		tryReparentingDatabase(false, ""),

		binaryUpgradeStep(c.Node(2), currentVersion),

		createDatabaseStep("test2"),

		allowAutoUpgradeStep(1),
		waitForUpgradeStep(c.All()),

		tryReparentingDatabase(true, "pq: cannot perform ALTER DATABASE CONVERT TO SCHEMA"),

		createDatabaseStep("test3"),

		createTableInDatabasePublicSchema("test1"),
		createTableInDatabasePublicSchema("test2"),
		createTableInDatabasePublicSchema("test3"),

		insertIntoTable("test1"),
		insertIntoTable("test2"),
		insertIntoTable("test3"),

		selectFromTable("test1"),
		selectFromTable("test2"),
		selectFromTable("test3"),

		dropTableInDatabase("test1"),
		dropTableInDatabase("test2"),
		dropTableInDatabase("test3"),
	)

	newVersionUpgradeTest(c, steps...).run(ctx, t)
}

func createDatabaseStep(dbName string) versionStep {
	__antithesis_instrumentation__.Notify(52379)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52380)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52382)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52381)
		require.NoError(t, err)
		_, err = conn.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		require.NoError(t, err)
	}
}

func createTableInDatabasePublicSchema(dbName string) versionStep {
	__antithesis_instrumentation__.Notify(52383)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52384)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52386)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52385)
		require.NoError(t, err)
		_, err = conn.Exec(fmt.Sprintf("CREATE TABLE %s.public.t(x INT)", dbName))
		require.NoError(t, err)
	}
}

func dropTableInDatabase(dbName string) versionStep {
	__antithesis_instrumentation__.Notify(52387)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52388)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52390)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52389)
		require.NoError(t, err)
		_, err = conn.Exec(fmt.Sprintf("DROP TABLE %s.public.t", dbName))
		require.NoError(t, err)
	}
}

func insertIntoTable(dbName string) versionStep {
	__antithesis_instrumentation__.Notify(52391)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52392)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52394)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52393)
		require.NoError(t, err)
		_, err = conn.Exec(fmt.Sprintf("INSERT INTO %s.public.t VALUES (0), (1), (2)", dbName))
		require.NoError(t, err)
	}
}

func selectFromTable(dbName string) versionStep {
	__antithesis_instrumentation__.Notify(52395)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52396)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52399)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52397)
		require.NoError(t, err)
		rows, err := conn.Query(fmt.Sprintf("SELECT x FROM %s.public.t ORDER BY x", dbName))
		defer func() {
			__antithesis_instrumentation__.Notify(52400)
			_ = rows.Close()
		}()
		__antithesis_instrumentation__.Notify(52398)
		require.NoError(t, err)
		numRows := 3
		var x int
		for i := 0; i < numRows; i++ {
			__antithesis_instrumentation__.Notify(52401)
			rows.Next()
			err := rows.Scan(&x)
			require.NoError(t, err)
			require.Equal(t, x, i)
		}
	}
}

func tryReparentingDatabase(shouldError bool, errRe string) versionStep {
	__antithesis_instrumentation__.Notify(52402)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52403)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		defer func() {
			__antithesis_instrumentation__.Notify(52406)
			_ = conn.Close()
		}()
		__antithesis_instrumentation__.Notify(52404)
		require.NoError(t, err)
		_, err = conn.Exec("CREATE DATABASE to_reparent;")
		require.NoError(t, err)
		_, err = conn.Exec("CREATE DATABASE new_parent")
		require.NoError(t, err)

		_, err = conn.Exec("ALTER DATABASE to_reparent CONVERT TO SCHEMA WITH PARENT new_parent;")

		if !shouldError {
			__antithesis_instrumentation__.Notify(52407)
			require.NoError(t, err)
		} else {
			__antithesis_instrumentation__.Notify(52408)
			if !testutils.IsError(err, errRe) {
				__antithesis_instrumentation__.Notify(52410)
				t.Fatalf("expected error '%s', got: %s", errRe, pgerror.FullError(err))
			} else {
				__antithesis_instrumentation__.Notify(52411)
			}
			__antithesis_instrumentation__.Notify(52409)

			_, err = conn.Exec("DROP DATABASE to_reparent")
			require.NoError(t, err)
		}
		__antithesis_instrumentation__.Notify(52405)

		_, err = conn.Exec("DROP DATABASE new_parent")
		require.NoError(t, err)
	}
}

func resetStep() versionStep {
	__antithesis_instrumentation__.Notify(52412)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52413)
		err := u.c.WipeE(ctx, t.L())
		require.NoError(t, err)
		err = u.c.RunE(ctx, u.c.All(), "rm -rf "+t.PerfArtifactsDir())
		require.NoError(t, err)
		err = u.c.RunE(ctx, u.c.All(), "rm -rf {store-dir}")
		require.NoError(t, err)
	}
}
