package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

var hibernateReleaseTagRegex = regexp.MustCompile(`^(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var supportedHibernateTag = "5.4.30"

type hibernateOptions struct {
	testName string
	testDir  string
	buildCmd,
	testCmd string
	blocklists  blocklistsForVersion
	dbSetupFunc func(ctx context.Context, t test.Test, c cluster.Cluster)
}

var (
	hibernateOpts = hibernateOptions{
		testName: "hibernate",
		testDir:  "hibernate-core",
		buildCmd: `cd /mnt/data1/hibernate/hibernate-core/ && ./../gradlew test -Pdb=cockroachdb ` +
			`--tests org.hibernate.jdbc.util.BasicFormatterTest.*`,
		testCmd:     "cd /mnt/data1/hibernate/hibernate-core/ && ./../gradlew test -Pdb=cockroachdb",
		blocklists:  hibernateBlocklists,
		dbSetupFunc: nil,
	}
	hibernateSpatialOpts = hibernateOptions{
		testName: "hibernate-spatial",
		testDir:  "hibernate-spatial",
		buildCmd: `cd /mnt/data1/hibernate/hibernate-spatial/ && ./../gradlew test -Pdb=cockroachdb_spatial ` +
			`--tests org.hibernate.spatial.dialect.postgis.*`,
		testCmd: `cd /mnt/data1/hibernate/hibernate-spatial && ` +
			`HIBERNATE_CONNECTION_LEAK_DETECTION=true ./../gradlew test -Pdb=cockroachdb_spatial`,
		blocklists: hibernateSpatialBlocklists,
		dbSetupFunc: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48353)
			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()
			if _, err := db.ExecContext(
				ctx,
				"SET CLUSTER SETTING sql.spatial.experimental_box2d_comparison_operators.enabled = on",
			); err != nil {
				__antithesis_instrumentation__.Notify(48354)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48355)
			}
		},
	}
)

func registerHibernate(r registry.Registry, opt hibernateOptions) {
	__antithesis_instrumentation__.Notify(48356)
	runHibernate := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(48358)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48376)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(48377)
		}
		__antithesis_instrumentation__.Notify(48359)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		if err := c.PutLibraries(ctx, "./lib"); err != nil {
			__antithesis_instrumentation__.Notify(48378)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48379)
		}
		__antithesis_instrumentation__.Notify(48360)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		if opt.dbSetupFunc != nil {
			__antithesis_instrumentation__.Notify(48380)
			opt.dbSetupFunc(ctx, t, c)
		} else {
			__antithesis_instrumentation__.Notify(48381)
		}
		__antithesis_instrumentation__.Notify(48361)

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(48382)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48383)
		}
		__antithesis_instrumentation__.Notify(48362)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(48384)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48385)
		}
		__antithesis_instrumentation__.Notify(48363)

		t.Status("cloning hibernate and installing prerequisites")
		latestTag, err := repeatGetLatestTag(
			ctx, t, "hibernate", "hibernate-orm", hibernateReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48386)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48387)
		}
		__antithesis_instrumentation__.Notify(48364)
		t.L().Printf("Latest Hibernate release is %s.", latestTag)
		t.L().Printf("Supported Hibernate release is %s.", supportedHibernateTag)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48388)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48389)
		}
		__antithesis_instrumentation__.Notify(48365)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install default-jre openjdk-8-jdk-headless gradle`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48390)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48391)
		}
		__antithesis_instrumentation__.Notify(48366)

		if err := repeatRunE(
			ctx, t, c, node, "remove old Hibernate", `rm -rf /mnt/data1/hibernate`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48392)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48393)
		}
		__antithesis_instrumentation__.Notify(48367)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/hibernate/hibernate-orm.git",
			"/mnt/data1/hibernate",
			supportedHibernateTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(48394)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48395)
		}
		__antithesis_instrumentation__.Notify(48368)

		t.Status("building hibernate (without tests)")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"building hibernate (without tests)",
			opt.buildCmd,
		); err != nil {
			__antithesis_instrumentation__.Notify(48396)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48397)
		}
		__antithesis_instrumentation__.Notify(48369)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"delete test result from build output",
			fmt.Sprintf(`rm -rf /mnt/data1/hibernate/%s/target/test-results/test`, opt.testDir),
		); err != nil {
			__antithesis_instrumentation__.Notify(48398)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48399)
		}
		__antithesis_instrumentation__.Notify(48370)

		blocklistName, expectedFailures, _, _ := opt.blocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(48400)
			t.Fatalf("No hibernate blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(48401)
		}
		__antithesis_instrumentation__.Notify(48371)
		t.L().Printf("Running cockroach version %s, using blocklist %s", version, blocklistName)

		t.Status("running hibernate test suite, will take at least 3 hours")

		_ = c.RunE(ctx, node, opt.testCmd)

		t.Status("collecting the test results")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"copy html report",
			fmt.Sprintf(`cp /mnt/data1/hibernate/%s/target/reports/tests/test ~/logs/report -a`, opt.testDir),
		); err != nil {
			__antithesis_instrumentation__.Notify(48402)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48403)
		}
		__antithesis_instrumentation__.Notify(48372)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"copy test result files",
			fmt.Sprintf(`cp /mnt/data1/hibernate/%s/target/test-results/test ~/logs/report/results -a`, opt.testDir),
		); err != nil {
			__antithesis_instrumentation__.Notify(48404)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48405)
		}
		__antithesis_instrumentation__.Notify(48373)

		result, err := repeatRunWithDetailsSingleNode(
			ctx,
			c,
			t,
			node,
			"get list of test files",
			fmt.Sprintf(`ls /mnt/data1/hibernate/%s/target/test-results/test/*.xml`, opt.testDir),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48406)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48407)
		}
		__antithesis_instrumentation__.Notify(48374)
		output := []byte(result.Stdout + result.Stderr)
		if len(output) == 0 {
			__antithesis_instrumentation__.Notify(48408)
			t.Fatal("could not find any test result files")
		} else {
			__antithesis_instrumentation__.Notify(48409)
		}
		__antithesis_instrumentation__.Notify(48375)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "hibernate", output,
			blocklistName, expectedFailures, nil, version, supportedHibernateTag,
		)
	}
	__antithesis_instrumentation__.Notify(48357)

	r.Add(registry.TestSpec{
		Name:    opt.testName,
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48410)
			runHibernate(ctx, t, c)
		},
	})
}
