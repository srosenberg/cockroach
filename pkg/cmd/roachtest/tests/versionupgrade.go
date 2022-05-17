package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/stretchr/testify/require"
)

var (
	v201 = roachpb.Version{Major: 20, Minor: 1}
	v202 = roachpb.Version{Major: 20, Minor: 2}
)

var versionUpgradeTestFeatures = versionFeatureStep{

	stmtFeatureTest("Object Access", v201, `
-- We should be able to successfully select from objects created in ancient
-- versions of CRDB using their FQNs. Prevents bugs such as #43141, where
-- databases created before a migration were inaccessible after the
-- migration.
--
-- NB: the data has been baked into the fixtures. Originally created via:
--   create database persistent_db
--   create table persistent_db.persistent_table(a int)"))
-- on CRDB v1.0
select * from persistent_db.persistent_table;
show tables from persistent_db;
`),
	stmtFeatureTest("JSONB", v201, `
CREATE DATABASE IF NOT EXISTS test;
CREATE TABLE test.t (j JSONB);
DROP TABLE test.t;
	`),
	stmtFeatureTest("Sequences", v201, `
CREATE DATABASE IF NOT EXISTS test;
CREATE SEQUENCE test.test_sequence;
DROP SEQUENCE test.test_sequence;
	`),
	stmtFeatureTest("Computed Columns", v201, `
CREATE DATABASE IF NOT EXISTS test;
CREATE TABLE test.t (x INT AS (3) STORED);
DROP TABLE test.t;
	`),
}

func runVersionUpgrade(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(52414)
	predecessorVersion, err := PredecessorVersion(*t.BuildVersion())
	if err != nil {
		__antithesis_instrumentation__.Notify(52418)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52419)
	}
	__antithesis_instrumentation__.Notify(52415)

	if false {
		__antithesis_instrumentation__.Notify(52420)

		makeFixtureVersion := "21.2.0"
		makeVersionFixtureAndFatal(ctx, t, c, makeFixtureVersion)
	} else {
		__antithesis_instrumentation__.Notify(52421)
	}
	__antithesis_instrumentation__.Notify(52416)

	testFeaturesStep := versionUpgradeTestFeatures.step(c.All())
	schemaChangeStep := runSchemaChangeWorkloadStep(c.All().RandNode()[0], 10, 2)

	_ = schemaChangeStep
	backupStep := func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52422)

		dest := fmt.Sprintf("nodelocal://0/%d", timeutil.Now().UnixNano())
		_, err := u.conn(ctx, t, 1).ExecContext(ctx, `BACKUP TO $1`, dest)
		require.NoError(t, err)
	}
	__antithesis_instrumentation__.Notify(52417)

	u := newVersionUpgradeTest(c,

		uploadAndStartFromCheckpointFixture(c.All(), predecessorVersion),
		uploadAndInitSchemaChangeWorkload(),
		waitForUpgradeStep(c.All()),
		testFeaturesStep,

		preventAutoUpgradeStep(1),

		binaryUpgradeStep(c.All(), ""),
		testFeaturesStep,

		backupStep,

		binaryUpgradeStep(c.All(), predecessorVersion),
		testFeaturesStep,

		backupStep,

		binaryUpgradeStep(c.All(), ""),
		allowAutoUpgradeStep(1),
		testFeaturesStep,

		backupStep,
		waitForUpgradeStep(c.All()),
		testFeaturesStep,

		backupStep,

		enableTracingGloballyStep,
		testFeaturesStep,

		backupStep,
	)

	u.run(ctx, t)
}

func (u *versionUpgradeTest) run(ctx context.Context, t test.Test) {
	__antithesis_instrumentation__.Notify(52423)
	defer func() {
		__antithesis_instrumentation__.Notify(52425)
		for _, db := range u.conns {
			__antithesis_instrumentation__.Notify(52426)
			_ = db.Close()
		}
	}()
	__antithesis_instrumentation__.Notify(52424)

	for _, step := range u.steps {
		__antithesis_instrumentation__.Notify(52427)
		if step != nil {
			__antithesis_instrumentation__.Notify(52428)
			step(ctx, t, u)
		} else {
			__antithesis_instrumentation__.Notify(52429)
		}
	}
}

type versionUpgradeTest struct {
	goOS  string
	c     cluster.Cluster
	steps []versionStep

	conns []*gosql.DB
}

func newVersionUpgradeTest(c cluster.Cluster, steps ...versionStep) *versionUpgradeTest {
	__antithesis_instrumentation__.Notify(52430)
	return &versionUpgradeTest{
		goOS:  ifLocal(c, runtime.GOOS, "linux"),
		c:     c,
		steps: steps,
	}
}

func checkpointName(binaryVersion string) string {
	__antithesis_instrumentation__.Notify(52431)
	return "checkpoint-v" + binaryVersion
}

func (u *versionUpgradeTest) conn(ctx context.Context, t test.Test, i int) *gosql.DB {
	__antithesis_instrumentation__.Notify(52432)
	if u.conns == nil {
		__antithesis_instrumentation__.Notify(52434)
		for _, i := range u.c.All() {
			__antithesis_instrumentation__.Notify(52435)
			u.conns = append(u.conns, u.c.Conn(ctx, t.L(), i))
		}
	} else {
		__antithesis_instrumentation__.Notify(52436)
	}
	__antithesis_instrumentation__.Notify(52433)
	db := u.conns[i-1]

	_ = db.PingContext(ctx)
	return db
}

func uploadVersion(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	nodes option.NodeListOption,
	newVersion string,
) (binaryName string) {
	__antithesis_instrumentation__.Notify(52437)
	binaryName = "./cockroach"
	if newVersion == "" {
		__antithesis_instrumentation__.Notify(52439)
		if err := c.PutE(ctx, t.L(), t.Cockroach(), binaryName, nodes); err != nil {
			__antithesis_instrumentation__.Notify(52440)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52441)
		}
	} else {
		__antithesis_instrumentation__.Notify(52442)
		if binary, ok := t.VersionsBinaryOverride()[newVersion]; ok {
			__antithesis_instrumentation__.Notify(52443)

			t.L().Printf("using binary override for version %s: %s", newVersion, binary)
			binaryName = "./cockroach-" + newVersion
			if err := c.PutE(ctx, t.L(), binary, binaryName, nodes); err != nil {
				__antithesis_instrumentation__.Notify(52444)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52445)
			}
		} else {
			__antithesis_instrumentation__.Notify(52446)
			v := "v" + newVersion
			dir := v
			binaryName = filepath.Join(dir, "cockroach")

			if err := c.RunE(ctx, nodes, "test", "-e", binaryName); err != nil {
				__antithesis_instrumentation__.Notify(52447)
				if err := c.RunE(ctx, nodes, "mkdir", "-p", dir); err != nil {
					__antithesis_instrumentation__.Notify(52449)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(52450)
				}
				__antithesis_instrumentation__.Notify(52448)
				if err := c.Stage(ctx, t.L(), "release", v, dir, nodes); err != nil {
					__antithesis_instrumentation__.Notify(52451)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(52452)
				}
			} else {
				__antithesis_instrumentation__.Notify(52453)
			}
		}
	}
	__antithesis_instrumentation__.Notify(52438)
	return binaryPathFromVersion(newVersion)
}

func binaryPathFromVersion(v string) string {
	__antithesis_instrumentation__.Notify(52454)
	if v == "" {
		__antithesis_instrumentation__.Notify(52456)
		return "./cockroach"
	} else {
		__antithesis_instrumentation__.Notify(52457)
	}
	__antithesis_instrumentation__.Notify(52455)
	return filepath.Join("v"+v, "cockroach")
}

func (u *versionUpgradeTest) uploadVersion(
	ctx context.Context, t test.Test, nodes option.NodeListOption, newVersion string,
) string {
	__antithesis_instrumentation__.Notify(52458)
	return uploadVersion(ctx, t, u.c, nodes, newVersion)
}

func (u *versionUpgradeTest) binaryVersion(
	ctx context.Context, t test.Test, i int,
) roachpb.Version {
	__antithesis_instrumentation__.Notify(52459)
	db := u.conn(ctx, t, i)

	var sv string
	if err := db.QueryRow(`SELECT crdb_internal.node_executable_version();`).Scan(&sv); err != nil {
		__antithesis_instrumentation__.Notify(52463)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52464)
	}
	__antithesis_instrumentation__.Notify(52460)

	if len(sv) == 0 {
		__antithesis_instrumentation__.Notify(52465)
		t.Fatal("empty version")
	} else {
		__antithesis_instrumentation__.Notify(52466)
	}
	__antithesis_instrumentation__.Notify(52461)

	cv, err := roachpb.ParseVersion(sv)
	if err != nil {
		__antithesis_instrumentation__.Notify(52467)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52468)
	}
	__antithesis_instrumentation__.Notify(52462)
	return cv
}

func (u *versionUpgradeTest) clusterVersion(
	ctx context.Context, t test.Test, i int,
) roachpb.Version {
	__antithesis_instrumentation__.Notify(52469)
	db := u.conn(ctx, t, i)

	var sv string
	if err := db.QueryRowContext(ctx, `SHOW CLUSTER SETTING version`).Scan(&sv); err != nil {
		__antithesis_instrumentation__.Notify(52472)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52473)
	}
	__antithesis_instrumentation__.Notify(52470)

	cv, err := roachpb.ParseVersion(sv)
	if err != nil {
		__antithesis_instrumentation__.Notify(52474)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52475)
	}
	__antithesis_instrumentation__.Notify(52471)
	return cv
}

type versionStep func(ctx context.Context, t test.Test, u *versionUpgradeTest)

func uploadAndStartFromCheckpointFixture(nodes option.NodeListOption, v string) versionStep {
	__antithesis_instrumentation__.Notify(52476)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52477)
		u.c.Run(ctx, nodes, "mkdir", "-p", "{store-dir}")
		vv := version.MustParse("v" + v)

		name := checkpointName(
			roachpb.Version{Major: int32(vv.Major()), Minor: int32(vv.Minor())}.String(),
		)
		for _, i := range nodes {
			__antithesis_instrumentation__.Notify(52479)
			u.c.Put(ctx,
				"pkg/cmd/roachtest/fixtures/"+strconv.Itoa(i)+"/"+name+".tgz",
				"{store-dir}/fixture.tgz", u.c.Node(i),
			)
		}
		__antithesis_instrumentation__.Notify(52478)

		u.c.Run(ctx, nodes, "cd {store-dir} && [ ! -f {store-dir}/CURRENT ] && tar -xf fixture.tgz")

		binary := u.uploadVersion(ctx, t, nodes, v)
		settings := install.MakeClusterSettings(install.BinaryOption(binary))
		startOpts := option.DefaultStartOpts()

		startOpts.RoachprodOpts.Sequential = false
		u.c.Start(ctx, t.L(), startOpts, settings, nodes)
	}
}

func binaryUpgradeStep(nodes option.NodeListOption, newVersion string) versionStep {
	__antithesis_instrumentation__.Notify(52480)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52481)
		upgradeNodes(ctx, nodes, newVersion, t, u.c)

	}
}

func upgradeNodes(
	ctx context.Context,
	nodes option.NodeListOption,
	newVersion string,
	t test.Test,
	c cluster.Cluster,
) {
	__antithesis_instrumentation__.Notify(52482)

	rand.Shuffle(len(nodes), func(i, j int) {
		__antithesis_instrumentation__.Notify(52484)
		nodes[i], nodes[j] = nodes[j], nodes[i]
	})
	__antithesis_instrumentation__.Notify(52483)
	for _, node := range nodes {
		__antithesis_instrumentation__.Notify(52485)
		v := newVersion
		if v == "" {
			__antithesis_instrumentation__.Notify(52488)
			v = "<latest>"
		} else {
			__antithesis_instrumentation__.Notify(52489)
		}
		__antithesis_instrumentation__.Notify(52486)
		newVersionMsg := newVersion
		if newVersion == "" {
			__antithesis_instrumentation__.Notify(52490)
			newVersionMsg = "<current>"
		} else {
			__antithesis_instrumentation__.Notify(52491)
		}
		__antithesis_instrumentation__.Notify(52487)
		t.L().Printf("restarting node %d into version %s", node, newVersionMsg)
		c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(node))

		binary := uploadVersion(ctx, t, c, c.Node(node), newVersion)
		settings := install.MakeClusterSettings(install.BinaryOption(binary))
		startOpts := option.DefaultStartOpts()
		c.Start(ctx, t.L(), startOpts, settings, c.Node(node))
	}
}

func enableTracingGloballyStep(ctx context.Context, t test.Test, u *versionUpgradeTest) {
	__antithesis_instrumentation__.Notify(52492)
	db := u.conn(ctx, t, 1)

	_, err := db.ExecContext(ctx, `SET CLUSTER SETTING trace.debug.enable = $1`, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(52493)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52494)
	}
}

func preventAutoUpgradeStep(node int) versionStep {
	__antithesis_instrumentation__.Notify(52495)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52496)
		db := u.conn(ctx, t, node)
		_, err := db.ExecContext(ctx, `SET CLUSTER SETTING cluster.preserve_downgrade_option = $1`, u.binaryVersion(ctx, t, node).String())
		if err != nil {
			__antithesis_instrumentation__.Notify(52497)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52498)
		}
	}
}

func allowAutoUpgradeStep(node int) versionStep {
	__antithesis_instrumentation__.Notify(52499)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52500)
		db := u.conn(ctx, t, node)
		_, err := db.ExecContext(ctx, `RESET CLUSTER SETTING cluster.preserve_downgrade_option`)
		if err != nil {
			__antithesis_instrumentation__.Notify(52501)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52502)
		}
	}
}

func waitForUpgradeStep(nodes option.NodeListOption) versionStep {
	__antithesis_instrumentation__.Notify(52503)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52504)
		newVersion := u.binaryVersion(ctx, t, nodes[0]).String()
		t.L().Printf("%s: waiting for cluster to auto-upgrade\n", newVersion)

		for _, i := range nodes {
			__antithesis_instrumentation__.Notify(52506)
			err := retry.ForDuration(5*time.Minute, func() error {
				__antithesis_instrumentation__.Notify(52508)
				currentVersion := u.clusterVersion(ctx, t, i).String()
				if currentVersion != newVersion {
					__antithesis_instrumentation__.Notify(52510)
					return fmt.Errorf("%d: expected version %s, got %s", i, newVersion, currentVersion)
				} else {
					__antithesis_instrumentation__.Notify(52511)
				}
				__antithesis_instrumentation__.Notify(52509)
				t.L().Printf("%s: acked by n%d", currentVersion, i)
				return nil
			})
			__antithesis_instrumentation__.Notify(52507)
			if err != nil {
				__antithesis_instrumentation__.Notify(52512)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52513)
			}
		}
		__antithesis_instrumentation__.Notify(52505)

		t.L().Printf("%s: nodes %v are upgraded\n", newVersion, nodes)

	}
}

func setClusterSettingVersionStep(ctx context.Context, t test.Test, u *versionUpgradeTest) {
	__antithesis_instrumentation__.Notify(52514)
	db := u.conn(ctx, t, 1)
	t.L().Printf("bumping cluster version")

	if _, err := db.ExecContext(
		ctx, `SET CLUSTER SETTING version = crdb_internal.node_executable_version()`,
	); err != nil {
		__antithesis_instrumentation__.Notify(52516)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52517)
	}
	__antithesis_instrumentation__.Notify(52515)
	t.L().Printf("cluster version bumped")
}

type versionFeatureTest struct {
	name string
	fn   func(context.Context, test.Test, *versionUpgradeTest, option.NodeListOption) (skipped bool)
}

type versionFeatureStep []versionFeatureTest

func (vs versionFeatureStep) step(nodes option.NodeListOption) versionStep {
	__antithesis_instrumentation__.Notify(52518)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52519)
		for _, feature := range vs {
			__antithesis_instrumentation__.Notify(52520)
			t.L().Printf("checking %s", feature.name)
			tBegin := timeutil.Now()
			skipped := feature.fn(ctx, t, u, nodes)
			dur := fmt.Sprintf("%.2fs", timeutil.Since(tBegin).Seconds())
			if skipped {
				__antithesis_instrumentation__.Notify(52521)
				t.L().Printf("^-- skip (%s)", dur)
			} else {
				__antithesis_instrumentation__.Notify(52522)
				t.L().Printf("^-- ok (%s)", dur)
			}
		}
	}
}

func stmtFeatureTest(
	name string, minVersion roachpb.Version, stmt string, args ...interface{},
) versionFeatureTest {
	__antithesis_instrumentation__.Notify(52523)
	return versionFeatureTest{
		name: name,
		fn: func(ctx context.Context, t test.Test, u *versionUpgradeTest, nodes option.NodeListOption) (skipped bool) {
			__antithesis_instrumentation__.Notify(52524)
			i := nodes.RandNode()[0]
			if u.clusterVersion(ctx, t, i).Less(minVersion) {
				__antithesis_instrumentation__.Notify(52527)
				return true
			} else {
				__antithesis_instrumentation__.Notify(52528)
			}
			__antithesis_instrumentation__.Notify(52525)
			db := u.conn(ctx, t, i)
			if _, err := db.ExecContext(ctx, stmt, args...); err != nil {
				__antithesis_instrumentation__.Notify(52529)
				if testutils.IsError(err, "no inbound stream connection") && func() bool {
					__antithesis_instrumentation__.Notify(52531)
					return u.clusterVersion(ctx, t, i).Less(v202) == true
				}() == true {
					__antithesis_instrumentation__.Notify(52532)

					return true
				} else {
					__antithesis_instrumentation__.Notify(52533)
				}
				__antithesis_instrumentation__.Notify(52530)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(52534)
			}
			__antithesis_instrumentation__.Notify(52526)
			return false
		},
	}
}

func makeVersionFixtureAndFatal(
	ctx context.Context, t test.Test, c cluster.Cluster, makeFixtureVersion string,
) {
	__antithesis_instrumentation__.Notify(52535)
	var useLocalBinary bool
	if makeFixtureVersion == "" {
		__antithesis_instrumentation__.Notify(52539)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(1))
		require.NoError(t, c.Conn(ctx, t.L(), 1).QueryRowContext(
			ctx,
			`select regexp_extract(value, '^v([0-9]+\.[0-9]+\.[0-9]+)') from crdb_internal.node_build_info where field = 'Version';`,
		).Scan(&makeFixtureVersion))
		c.Wipe(ctx, c.Node(1))
		useLocalBinary = true
	} else {
		__antithesis_instrumentation__.Notify(52540)
	}
	__antithesis_instrumentation__.Notify(52536)

	predecessorVersion, err := PredecessorVersion(*version.MustParse("v" + makeFixtureVersion))
	if err != nil {
		__antithesis_instrumentation__.Notify(52541)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52542)
	}
	__antithesis_instrumentation__.Notify(52537)

	t.L().Printf("making fixture for %s (starting at %s)", makeFixtureVersion, predecessorVersion)

	if useLocalBinary {
		__antithesis_instrumentation__.Notify(52543)

		makeFixtureVersion = ""
	} else {
		__antithesis_instrumentation__.Notify(52544)
	}
	__antithesis_instrumentation__.Notify(52538)

	newVersionUpgradeTest(c,

		uploadAndStartFromCheckpointFixture(c.All(), predecessorVersion),
		waitForUpgradeStep(c.All()),

		binaryUpgradeStep(c.All(), makeFixtureVersion),
		waitForUpgradeStep(c.All()),

		func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(52545)

			name := checkpointName(u.binaryVersion(ctx, t, 1).String())
			u.c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.All())

			c.Run(ctx, c.All(), binaryPathFromVersion(makeFixtureVersion), "debug", "pebble", "db", "checkpoint",
				"{store-dir}", "{store-dir}/"+name)

			c.Run(ctx, c.Node(1), "cp", "{store-dir}/cluster-bootstrapped", "{store-dir}/"+name)
			c.Run(ctx, c.All(), "tar", "-C", "{store-dir}/"+name, "-czf", "{log-dir}/"+name+".tgz", ".")
			t.Fatalf(`successfully created checkpoints; failing test on purpose.

Invoke the following to move the archives to the right place and commit the
result:

for i in 1 2 3 4; do
  mkdir -p pkg/cmd/roachtest/fixtures/${i} && \
  mv artifacts/acceptance/version-upgrade/run_1/logs/${i}.unredacted/checkpoint-*.tgz \
     pkg/cmd/roachtest/fixtures/${i}/
done
`)
		}).run(ctx, t)
}

func importTPCCStep(
	oldV string, headroomWarehouses int, crdbNodes option.NodeListOption,
) versionStep {
	__antithesis_instrumentation__.Notify(52546)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52547)

		var cmd string
		if oldV == "" {
			__antithesis_instrumentation__.Notify(52550)
			cmd = tpccImportCmd(headroomWarehouses)
		} else {
			__antithesis_instrumentation__.Notify(52551)
			cmd = tpccImportCmdWithCockroachBinary(filepath.Join("v"+oldV, "cockroach"), headroomWarehouses, "--checks=false")
		}
		__antithesis_instrumentation__.Notify(52548)

		m := u.c.NewMonitor(ctx, crdbNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(52552)
			return u.c.RunE(ctx, u.c.Node(crdbNodes[0]), cmd)
		})
		__antithesis_instrumentation__.Notify(52549)
		m.Wait()
	}
}

func importLargeBankStep(oldV string, rows int, crdbNodes option.NodeListOption) versionStep {
	__antithesis_instrumentation__.Notify(52553)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(52554)

		binary := "./cockroach"
		if oldV != "" {
			__antithesis_instrumentation__.Notify(52557)
			binary = filepath.Join("v"+oldV, "cockroach")
		} else {
			__antithesis_instrumentation__.Notify(52558)
		}
		__antithesis_instrumentation__.Notify(52555)

		m := u.c.NewMonitor(ctx, crdbNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(52559)
			return u.c.RunE(ctx, u.c.Node(crdbNodes[0]), binary, "workload", "fixtures", "import", "bank",
				"--payload-bytes=10240", "--rows="+fmt.Sprint(rows), "--seed=4", "--db=bigbank")
		})
		__antithesis_instrumentation__.Notify(52556)
		m.Wait()
	}
}
