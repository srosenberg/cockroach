package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

func registerSchemaChangeDatabaseVersionUpgrade(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50496)

	r.Add(registry.TestSpec{
		Name:    "schemachange/database-version-upgrade",
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50497)
			runSchemaChangeDatabaseVersionUpgrade(ctx, t, c, *t.BuildVersion())
		},
	})
}

func uploadAndStart(nodes option.NodeListOption, v string) versionStep {
	__antithesis_instrumentation__.Notify(50498)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50499)

		binary := u.uploadVersion(ctx, t, nodes, v)

		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.Sequential = false
		settings := install.MakeClusterSettings(install.BinaryOption(binary))
		u.c.Start(ctx, t.L(), startOpts, settings, nodes)
	}
}

func runSchemaChangeDatabaseVersionUpgrade(
	ctx context.Context, t test.Test, c cluster.Cluster, buildVersion version.Version,
) {
	__antithesis_instrumentation__.Notify(50500)

	const mainVersion = ""
	predecessorVersion, err := PredecessorVersion(buildVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(50511)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50512)
	}
	__antithesis_instrumentation__.Notify(50501)

	publicSchemaWithDescriptorsVersion := clusterversion.ByKey(clusterversion.PublicSchemasWithDescriptors)

	createDatabaseWithTableStep := func(dbName string) versionStep {
		__antithesis_instrumentation__.Notify(50513)
		t.L().Printf("creating database %s", dbName)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(50514)
			db := u.conn(ctx, t, 1)
			_, err := db.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s; CREATE TABLE %s.t(a INT)`, dbName, dbName))
			require.NoError(t, err)
		}
	}
	__antithesis_instrumentation__.Notify(50502)

	assertDatabaseResolvable := func(ctx context.Context, db *gosql.DB, dbName string) error {
		__antithesis_instrumentation__.Notify(50515)
		var tblName string
		row := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT table_name FROM [SHOW TABLES FROM %s]`, dbName))
		if err := row.Scan(&tblName); err != nil {
			__antithesis_instrumentation__.Notify(50518)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50519)
		}
		__antithesis_instrumentation__.Notify(50516)
		if tblName != "t" {
			__antithesis_instrumentation__.Notify(50520)
			return errors.AssertionFailedf("unexpected table name %s", tblName)
		} else {
			__antithesis_instrumentation__.Notify(50521)
		}
		__antithesis_instrumentation__.Notify(50517)
		return nil
	}
	__antithesis_instrumentation__.Notify(50503)

	assertDatabaseNotResolvable := func(ctx context.Context, db *gosql.DB, dbName string) error {
		__antithesis_instrumentation__.Notify(50522)
		_, err = db.ExecContext(ctx, fmt.Sprintf(`SELECT table_name FROM [SHOW TABLES FROM %s]`, dbName))
		if err == nil || func() bool {
			__antithesis_instrumentation__.Notify(50524)
			return err.Error() != "pq: target database or schema does not exist" == true
		}() == true {
			__antithesis_instrumentation__.Notify(50525)
			return errors.Newf("unexpected error: %s", pgerror.FullError(err))
		} else {
			__antithesis_instrumentation__.Notify(50526)
		}
		__antithesis_instrumentation__.Notify(50523)
		return nil
	}
	__antithesis_instrumentation__.Notify(50504)

	runSchemaChangesStep := func(dbName string) versionStep {
		__antithesis_instrumentation__.Notify(50527)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(50528)
			t.L().Printf("running schema changes on %s", dbName)
			newDbName := dbName + "_new_name"
			dbNode1 := u.conn(ctx, t, 1)
			dbNode2 := u.conn(ctx, t, 2)

			_, err := dbNode1.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE %s RENAME TO %s`, dbName, newDbName))
			require.NoError(t, err)

			if err := assertDatabaseResolvable(ctx, dbNode1, newDbName); err != nil {
				__antithesis_instrumentation__.Notify(50536)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50537)
			}
			__antithesis_instrumentation__.Notify(50529)
			if err := assertDatabaseNotResolvable(ctx, dbNode1, dbName); err != nil {
				__antithesis_instrumentation__.Notify(50538)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50539)
			}
			__antithesis_instrumentation__.Notify(50530)

			if err := testutils.SucceedsSoonError(func() error {
				__antithesis_instrumentation__.Notify(50540)
				return assertDatabaseResolvable(ctx, dbNode2, newDbName)
			}); err != nil {
				__antithesis_instrumentation__.Notify(50541)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50542)
			}
			__antithesis_instrumentation__.Notify(50531)
			if err := testutils.SucceedsSoonError(func() error {
				__antithesis_instrumentation__.Notify(50543)
				return assertDatabaseNotResolvable(ctx, dbNode2, dbName)
			}); err != nil {
				__antithesis_instrumentation__.Notify(50544)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50545)
			}
			__antithesis_instrumentation__.Notify(50532)

			_, err = dbNode1.ExecContext(ctx, fmt.Sprintf(`DROP DATABASE %s CASCADE`, newDbName))
			require.NoError(t, err)

			if err := assertDatabaseNotResolvable(ctx, dbNode1, newDbName); err != nil {
				__antithesis_instrumentation__.Notify(50546)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50547)
			}
			__antithesis_instrumentation__.Notify(50533)
			if err := testutils.SucceedsSoonError(func() error {
				__antithesis_instrumentation__.Notify(50548)
				return assertDatabaseNotResolvable(ctx, dbNode2, newDbName)
			}); err != nil {
				__antithesis_instrumentation__.Notify(50549)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50550)
			}
			__antithesis_instrumentation__.Notify(50534)

			_, err = dbNode1.ExecContext(ctx, fmt.Sprintf(`CREATE DATABASE %s; CREATE TABLE %s.t(a INT)`, dbName, dbName))
			require.NoError(t, err)

			if err := assertDatabaseResolvable(ctx, dbNode1, dbName); err != nil {
				__antithesis_instrumentation__.Notify(50551)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50552)
			}
			__antithesis_instrumentation__.Notify(50535)
			if err := testutils.SucceedsSoonError(func() error {
				__antithesis_instrumentation__.Notify(50553)
				return assertDatabaseResolvable(ctx, dbNode1, dbName)
			}); err != nil {
				__antithesis_instrumentation__.Notify(50554)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50555)
			}
		}
	}
	__antithesis_instrumentation__.Notify(50505)

	createParentDatabaseStep := func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50556)
		if publicSchemaWithDescriptorsVersion.LessEq(u.clusterVersion(ctx, t, 1)) {
			__antithesis_instrumentation__.Notify(50558)
			return
		} else {
			__antithesis_instrumentation__.Notify(50559)
		}
		__antithesis_instrumentation__.Notify(50557)
		t.L().Printf("creating parent database")
		db := u.conn(ctx, t, 1)
		_, err := db.ExecContext(ctx, `CREATE DATABASE new_parent_db`)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(50506)

	reparentDatabaseStep := func(dbName string) versionStep {
		__antithesis_instrumentation__.Notify(50560)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(50561)
			if publicSchemaWithDescriptorsVersion.LessEq(u.clusterVersion(ctx, t, 1)) {
				__antithesis_instrumentation__.Notify(50563)
				return
			} else {
				__antithesis_instrumentation__.Notify(50564)
			}
			__antithesis_instrumentation__.Notify(50562)
			db := u.conn(ctx, t, 1)
			t.L().Printf("reparenting database %s", dbName)
			_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER DATABASE %s CONVERT TO SCHEMA WITH PARENT new_parent_db;`, dbName))
			require.NoError(t, err)
		}
	}
	__antithesis_instrumentation__.Notify(50507)

	validationStep := func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50565)
		t.L().Printf("validating")
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(1),
			[]string{"./cockroach debug doctor cluster", "--url {pgurl:1}"}...)
		require.NoError(t, err)
		t.L().Printf("%s", result.Stdout)
	}
	__antithesis_instrumentation__.Notify(50508)

	interactWithReparentedSchemaStep := func(schemaName string) versionStep {
		__antithesis_instrumentation__.Notify(50566)
		return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(50567)
			if publicSchemaWithDescriptorsVersion.LessEq(u.clusterVersion(ctx, t, 1)) {
				__antithesis_instrumentation__.Notify(50569)
				return
			} else {
				__antithesis_instrumentation__.Notify(50570)
			}
			__antithesis_instrumentation__.Notify(50568)
			t.L().Printf("running schema changes on %s", schemaName)
			db := u.conn(ctx, t, 1)

			_, err = db.ExecContext(ctx, `USE new_parent_db`)
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, fmt.Sprintf(`INSERT INTO %s.t VALUES (1)`, schemaName))
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER TABLE %s.t ADD COLUMN b INT`, schemaName))
			require.NoError(t, err)

			_, err = db.ExecContext(ctx, fmt.Sprintf(`CREATE TABLE %s.t2()`, schemaName))
			require.NoError(t, err)

			newSchemaName := schemaName + "_new"
			_, err = db.ExecContext(ctx, fmt.Sprintf(`ALTER SCHEMA %s RENAME TO %s`, schemaName, newSchemaName))
			require.NoError(t, err)
		}
	}
	__antithesis_instrumentation__.Notify(50509)

	dropDatabaseCascadeStep := func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(50571)
		if publicSchemaWithDescriptorsVersion.LessEq(u.clusterVersion(ctx, t, 1)) {
			__antithesis_instrumentation__.Notify(50573)
			return
		} else {
			__antithesis_instrumentation__.Notify(50574)
		}
		__antithesis_instrumentation__.Notify(50572)
		t.L().Printf("dropping parent database")
		db := u.conn(ctx, t, 1)
		_, err = db.ExecContext(ctx, `
USE defaultdb;
DROP DATABASE new_parent_db CASCADE;
`)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(50510)

	u := newVersionUpgradeTest(c,
		uploadAndStart(c.All(), predecessorVersion),
		waitForUpgradeStep(c.All()),
		preventAutoUpgradeStep(1),

		createDatabaseWithTableStep("db_0"),
		createDatabaseWithTableStep("db_1"),
		createDatabaseWithTableStep("db_2"),
		createDatabaseWithTableStep("db_3"),
		createDatabaseWithTableStep("db_4"),
		createDatabaseWithTableStep("db_5"),

		binaryUpgradeStep(c.Node(1), mainVersion),

		runSchemaChangesStep("db_1"),

		binaryUpgradeStep(c.Nodes(2, 3), mainVersion),

		runSchemaChangesStep("db_2"),

		binaryUpgradeStep(c.Node(1), predecessorVersion),

		runSchemaChangesStep("db_3"),

		binaryUpgradeStep(c.Nodes(2, 3), predecessorVersion),

		runSchemaChangesStep("db_4"),

		binaryUpgradeStep(c.All(), mainVersion),

		runSchemaChangesStep("db_5"),

		allowAutoUpgradeStep(1),
		waitForUpgradeStep(c.All()),

		createParentDatabaseStep,
		reparentDatabaseStep("db_0"),
		reparentDatabaseStep("db_1"),
		reparentDatabaseStep("db_2"),
		reparentDatabaseStep("db_3"),
		reparentDatabaseStep("db_4"),
		reparentDatabaseStep("db_5"),
		validationStep,

		interactWithReparentedSchemaStep("db_0"),
		interactWithReparentedSchemaStep("db_1"),
		interactWithReparentedSchemaStep("db_2"),
		interactWithReparentedSchemaStep("db_3"),
		interactWithReparentedSchemaStep("db_4"),
		interactWithReparentedSchemaStep("db_5"),
		validationStep,

		dropDatabaseCascadeStep,
		validationStep,
	)
	u.run(ctx, t)
}
