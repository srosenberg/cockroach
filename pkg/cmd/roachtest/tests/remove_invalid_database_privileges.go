package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

func registerRemoveInvalidDatabasePrivileges(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50158)
	r.Add(registry.TestSpec{
		Name:    "remove-invalid-database-privileges",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50159)
			runRemoveInvalidDatabasePrivileges(ctx, t, c, *t.BuildVersion())
		},
	})
}

func runRemoveInvalidDatabasePrivileges(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(50160)

	if buildVersion.Major() != 22 || func() bool {
		__antithesis_instrumentation__.Notify(50162)
		return buildVersion.Minor() != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(50163)
		t.L().PrintfCtx(ctx, "skipping test because build version is %s", buildVersion)
		return
	} else {
		__antithesis_instrumentation__.Notify(50164)
	}
	__antithesis_instrumentation__.Notify(50161)
	const mainVersion = ""

	const version21_1_12 = "21.1.12"
	const version21_2_4 = "21.2.4"
	u := newVersionUpgradeTest(c,
		uploadAndStart(c.All(), version21_1_12),
		waitForUpgradeStep(c.All()),
		preventAutoUpgradeStep(1),
		preventAutoUpgradeStep(2),
		preventAutoUpgradeStep(3),

		execSQL("CREATE USER testuser;", "", 1),
		execSQL("CREATE DATABASE test;", "", 1),
		execSQL("GRANT CREATE, SELECT, INSERT, UPDATE, DELETE ON DATABASE test TO testuser;", "", 1),

		showGrantsOnDatabase("system", [][]string{
			{"system", "admin", "GRANT"},
			{"system", "admin", "SELECT"},
			{"system", "root", "GRANT"},
			{"system", "root", "SELECT"},
		}),
		showGrantsOnDatabase("test", [][]string{
			{"test", "admin", "ALL"},
			{"test", "root", "ALL"},
			{"test", "testuser", "CREATE"},
			{"test", "testuser", "DELETE"},
			{"test", "testuser", "INSERT"},
			{"test", "testuser", "SELECT"},
			{"test", "testuser", "UPDATE"},
		}),

		binaryUpgradeStep(c.Node(1), version21_2_4),
		allowAutoUpgradeStep(1),
		binaryUpgradeStep(c.Node(2), version21_2_4),
		allowAutoUpgradeStep(2),
		binaryUpgradeStep(c.Node(3), version21_2_4),
		allowAutoUpgradeStep(3),

		waitForUpgradeStep(c.All()),

		binaryUpgradeStep(c.Node(1), mainVersion),
		allowAutoUpgradeStep(1),
		binaryUpgradeStep(c.Node(2), mainVersion),
		allowAutoUpgradeStep(2),
		binaryUpgradeStep(c.Node(3), mainVersion),
		allowAutoUpgradeStep(3),

		waitForUpgradeStep(c.All()),

		showGrantsOnDatabase("system", [][]string{
			{"system", "admin", "CONNECT"},
			{"system", "root", "CONNECT"},
		}),
		showGrantsOnDatabase("test", [][]string{
			{"test", "admin", "ALL"},
			{"test", "root", "ALL"},
			{"test", "testuser", "CREATE"},
		}),

		showDefaultPrivileges("test", [][]string{
			{"tables", "testuser", "SELECT"},
			{"tables", "testuser", "INSERT"},
			{"tables", "testuser", "DELETE"},
			{"tables", "testuser", "UPDATE"},
			{"types", "public", "USAGE"},
		}),

		showDefaultPrivileges("system", [][]string{
			{"types", "public", "USAGE"},
		}),
	)
	u.run(ctx, t)
}

func showGrantsOnDatabase(dbName string, expectedPrivileges [][]string) versionStep {
	__antithesis_instrumentation__.Notify(50165)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50166)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		require.NoError(t, err)
		rows, err := conn.Query(
			fmt.Sprintf("SELECT database_name, grantee, privilege_type FROM [SHOW GRANTS ON DATABASE %s]",
				dbName))
		require.NoError(t, err)
		var name, grantee, privilegeType string
		i := 0
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50168)
			privilegeRow := expectedPrivileges[i]
			err = rows.Scan(&name, &grantee, &privilegeType)
			require.NoError(t, err)

			require.Equal(t, privilegeRow[0], name)
			require.Equal(t, privilegeRow[1], grantee)
			require.Equal(t, privilegeRow[2], privilegeType)
			i++
		}
		__antithesis_instrumentation__.Notify(50167)

		if i != len(expectedPrivileges) {
			__antithesis_instrumentation__.Notify(50169)
			t.Errorf("expected %d rows, found %d rows", len(expectedPrivileges), i)
		} else {
			__antithesis_instrumentation__.Notify(50170)
		}
	}
}

func showDefaultPrivileges(dbName string, expectedDefaultPrivileges [][]string) versionStep {
	__antithesis_instrumentation__.Notify(50171)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50172)
		conn, err := u.c.ConnE(ctx, t.L(), loadNode)
		require.NoError(t, err)
		_, err = conn.Exec(fmt.Sprintf("USE %s", dbName))
		require.NoError(t, err)
		rows, err := conn.Query("SELECT object_type, grantee, privilege_type FROM [SHOW DEFAULT PRIVILEGES FOR ALL ROLES]")
		require.NoError(t, err)

		var objectType, grantee, privilegeType string
		i := 0
		for rows.Next() {
			__antithesis_instrumentation__.Notify(50174)
			defaultPrivilegeRow := expectedDefaultPrivileges[i]
			err = rows.Scan(&objectType, &grantee, &privilegeType)
			require.NoError(t, err)
			require.Equal(t, defaultPrivilegeRow[0], objectType)
			require.Equal(t, defaultPrivilegeRow[1], grantee)
			require.Equal(t, defaultPrivilegeRow[2], privilegeType)

			i++
		}
		__antithesis_instrumentation__.Notify(50173)

		if i != len(expectedDefaultPrivileges) {
			__antithesis_instrumentation__.Notify(50175)
			t.Errorf("expected %d rows, found %d rows", len(expectedDefaultPrivileges), i)
		} else {
			__antithesis_instrumentation__.Notify(50176)
		}
	}
}
