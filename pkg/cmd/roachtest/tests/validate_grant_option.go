package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

func registerValidateGrantOption(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52269)
	r.Add(registry.TestSpec{
		Name:    "validate-grant-option",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(52270)
			runRegisterValidateGrantOption(ctx, t, c, *t.BuildVersion())
		},
	})
}

func execSQL(sqlStatement string, expectedErrText string, node int) versionStep {
	__antithesis_instrumentation__.Notify(52271)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52272)
		conn, err := u.c.ConnE(ctx, t.L(), node)
		require.NoError(t, err)
		t.L().PrintfCtx(ctx, "user root on node %d executing: %s", node, sqlStatement)
		_, err = conn.Exec(sqlStatement)
		if len(expectedErrText) == 0 {
			__antithesis_instrumentation__.Notify(52273)
			require.NoError(t, err)
		} else {
			__antithesis_instrumentation__.Notify(52274)
			require.EqualError(t, err, expectedErrText)
		}
	}
}

func execSQLAsUser(sqlStatement string, user string, expectedErrText string, node int) versionStep {
	__antithesis_instrumentation__.Notify(52275)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52276)
		conn, err := u.c.ConnEAsUser(ctx, t.L(), node, user)
		require.NoError(t, err)
		t.L().PrintfCtx(ctx, "user %s on node %d executing: %s", user, node, sqlStatement)
		_, err = conn.Exec(sqlStatement)
		if len(expectedErrText) == 0 {
			__antithesis_instrumentation__.Notify(52277)
			require.NoError(t, err)
		} else {
			__antithesis_instrumentation__.Notify(52278)
			require.EqualError(t, err, expectedErrText)
		}
	}
}

func runRegisterValidateGrantOption(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(52279)

	if buildVersion.Major() != 22 || func() bool {
		__antithesis_instrumentation__.Notify(52282)
		return buildVersion.Minor() != 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(52283)
		t.L().PrintfCtx(ctx, "skipping test because build version is %s", buildVersion)
		return
	} else {
		__antithesis_instrumentation__.Notify(52284)
	}
	__antithesis_instrumentation__.Notify(52280)
	const mainVersion = ""
	predecessorVersion, err := PredecessorVersion(buildVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(52285)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52286)
	}
	__antithesis_instrumentation__.Notify(52281)

	u := newVersionUpgradeTest(c,
		uploadAndStart(c.All(), predecessorVersion),
		waitForUpgradeStep(c.All()),
		preventAutoUpgradeStep(1),
		preventAutoUpgradeStep(2),
		preventAutoUpgradeStep(3),

		execSQL("CREATE USER foo;", "", 1),
		execSQL("CREATE USER foo2;", "", 2),
		execSQL("CREATE USER foo3;", "", 2),
		execSQL("CREATE USER foo4;", "", 2),
		execSQL("CREATE USER foo5;", "", 2),
		execSQL("CREATE USER target;", "", 1),

		execSQL("CREATE DATABASE d;", "", 2),
		execSQL("CREATE SCHEMA s;", "", 3),
		execSQL("CREATE TABLE t1();", "", 2),
		execSQL("CREATE TYPE ty AS ENUM();", "", 3),

		execSQL("GRANT ALL PRIVILEGES ON DATABASE d TO foo;", "", 1),
		execSQL("GRANT ALL PRIVILEGES ON SCHEMA s TO foo2;", "", 3),
		execSQL("GRANT ALL PRIVILEGES ON TABLE t1 TO foo3;", "", 1),
		execSQL("GRANT ALL PRIVILEGES ON TYPE ty TO foo4;", "", 1),

		execSQLAsUser("SELECT * FROM t1;", "foo3", "", 2),
		execSQLAsUser("GRANT CREATE ON DATABASE d TO target;", "foo", "", 1),
		execSQLAsUser("GRANT USAGE ON SCHEMA s TO target;", "foo2", "", 3),
		execSQLAsUser("GRANT SELECT ON TABLE t1 TO target;", "foo3", "", 1),
		execSQLAsUser("GRANT USAGE ON TYPE ty TO target;", "foo4", "", 1),

		binaryUpgradeStep(c.Node(1), mainVersion),
		allowAutoUpgradeStep(1),
		execSQLAsUser("GRANT CREATE ON DATABASE d TO target;", "foo", "", 1),
		execSQLAsUser("GRANT CREATE ON DATABASE defaultdb TO foo;", "root", "", 1),
		execSQLAsUser("CREATE TABLE t2();", "foo", "", 1),
		execSQLAsUser("CREATE TABLE t3();", "foo", "", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo2;", "foo", "", 1),
		execSQLAsUser("GRANT GRANT, CREATE ON TABLE t3 TO foo2;", "foo", "", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo3;", "foo", "", 1),
		execSQLAsUser("GRANT GRANT, CREATE ON TABLE t3 TO foo3;", "foo", "", 2),

		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo4 WITH GRANT OPTION;", "foo", "pq: version 21.2-22 must be finalized to use grant options", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO foo4 WITH GRANT OPTION;", "foo", "pq: at or near \"with\": syntax error", 2),

		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo3;", "foo2",
			"pq: user foo2 does not have GRANT privilege on relation t2", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo3;", "foo2",
			"pq: user foo2 does not have GRANT privilege on relation t2", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO foo3;", "foo2", "", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO foo3;", "foo2", "", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo4;", "foo3",
			"pq: user foo3 does not have GRANT privilege on relation t2", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo4;", "foo3",
			"pq: user foo3 does not have GRANT privilege on relation t2", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO foo4;", "foo3", "", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO foo4;", "foo3", "", 2),

		execSQLAsUser("GRANT USAGE ON SCHEMA s TO target;", "foo2", "", 3),
		execSQLAsUser("GRANT INSERT ON TABLE t1 TO target;", "foo3", "", 2),
		execSQLAsUser("GRANT GRANT ON TYPE ty TO target;", "foo4", "", 2),

		binaryUpgradeStep(c.Node(2), mainVersion),
		allowAutoUpgradeStep(2),
		execSQLAsUser("GRANT ALL PRIVILEGES ON DATABASE d TO target;", "foo", "", 3),
		execSQLAsUser("GRANT ALL PRIVILEGES ON SCHEMA s TO target;", "foo2", "", 2),
		execSQLAsUser("GRANT ALL PRIVILEGES ON TABLE t1 TO target;", "foo3", "", 1),
		execSQLAsUser("GRANT ALL PRIVILEGES ON TYPE ty TO target;", "foo4", "", 1),
		execSQLAsUser("GRANT GRANT ON DATABASE d TO foo2;", "foo", "", 1),
		execSQLAsUser("GRANT GRANT ON SCHEMA s TO foo3;", "foo2", "", 1),
		execSQLAsUser("GRANT GRANT, SELECT ON TABLE t1 TO foo4;", "foo3", "", 2),
		execSQLAsUser("GRANT DELETE ON TABLE t1 TO foo4;", "foo3", "", 2),

		binaryUpgradeStep(c.Node(3), mainVersion),
		allowAutoUpgradeStep(3),
		waitForUpgradeStep(c.All()),

		execSQLAsUser("GRANT DELETE ON TABLE t1 TO foo2;", "foo3", "", 3),
		execSQLAsUser("GRANT ALL PRIVILEGES ON SCHEMA s TO foo3;", "target", "", 2),
		execSQLAsUser("GRANT CREATE ON DATABASE d TO foo3;", "foo2", "pq: user foo2 does not have CREATE privilege on database d", 2),
		execSQLAsUser("GRANT USAGE ON SCHEMA s TO foo;", "foo3", "", 2),
		execSQLAsUser("GRANT INSERT ON TABLE t1 TO foo;", "foo4", "pq: user foo4 does not have INSERT privilege on relation t1", 1),
		execSQLAsUser("GRANT SELECT ON TABLE t1 TO foo;", "foo4", "", 1),
		execSQLAsUser("GRANT DELETE ON TABLE t1 TO target;", "foo2", "pq: user foo2 missing WITH GRANT OPTION privilege on DELETE", 1),
		execSQLAsUser("GRANT DELETE ON TABLE t1 TO target;", "foo4", "", 1),
		execSQLAsUser("GRANT SELECT ON TABLE t1 TO target;", "foo4", "", 1),

		execSQLAsUser("GRANT SELECT ON TABLE t2 TO foo4 WITH GRANT OPTION;", "foo", "", 1),
		execSQLAsUser("GRANT SELECT ON TABLE t3 TO foo4 WITH GRANT OPTION;", "foo", "", 2),

		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo3;", "foo2", "pq: user foo2 missing WITH GRANT OPTION privilege on CREATE", 1),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO target;", "foo2", "", 2),
		execSQLAsUser("GRANT CREATE ON TABLE t2 TO foo4;", "foo3", "pq: user foo3 missing WITH GRANT OPTION privilege on CREATE", 3),
		execSQLAsUser("GRANT CREATE ON TABLE t3 TO target;", "foo3", "", 3),
	)
	u.run(ctx, t)
}
