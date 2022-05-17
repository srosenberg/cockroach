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

var pgjdbcReleaseTagRegex = regexp.MustCompile(`^REL(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var supportedPGJDBCTag = "REL42.3.3"

func registerPgjdbc(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49706)
	runPgjdbc := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(49708)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49726)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49727)
		}
		__antithesis_instrumentation__.Notify(49709)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(install.SecureOption(true)), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49728)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49729)
		}
		__antithesis_instrumentation__.Notify(49710)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49730)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49731)
		}
		__antithesis_instrumentation__.Notify(49711)

		t.Status("create admin user for tests")
		db, err := c.ConnE(ctx, t.L(), node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49732)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49733)
		}
		__antithesis_instrumentation__.Notify(49712)
		defer db.Close()
		stmts := []string{
			"CREATE USER test_admin WITH PASSWORD 'testpw'",
			"GRANT admin TO test_admin",
			"ALTER ROLE ALL SET serial_normalization = 'sql_sequence_cached'",
			"ALTER ROLE ALL SET statement_timeout = '60s'",
		}
		for _, stmt := range stmts {
			__antithesis_instrumentation__.Notify(49734)
			_, err = db.ExecContext(ctx, stmt)
			if err != nil {
				__antithesis_instrumentation__.Notify(49735)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49736)
			}
		}
		__antithesis_instrumentation__.Notify(49713)

		t.Status("cloning pgjdbc and installing prerequisites")

		latestTag, err := repeatGetLatestTag(
			ctx, t, "pgjdbc", "pgjdbc", pgjdbcReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(49737)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49738)
		}
		__antithesis_instrumentation__.Notify(49714)
		t.L().Printf("Latest pgjdbc release is %s.", latestTag)
		t.L().Printf("Supported pgjdbc release is %s.", supportedPGJDBCTag)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49739)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49740)
		}
		__antithesis_instrumentation__.Notify(49715)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install default-jre openjdk-8-jdk-headless gradle`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49741)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49742)
		}
		__antithesis_instrumentation__.Notify(49716)

		if err := repeatRunE(
			ctx, t, c, node, "remove old pgjdbc", `rm -rf /mnt/data1/pgjdbc`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49743)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49744)
		}
		__antithesis_instrumentation__.Notify(49717)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/pgjdbc/pgjdbc.git",
			"/mnt/data1/pgjdbc",
			supportedPGJDBCTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(49745)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49746)
		}
		__antithesis_instrumentation__.Notify(49718)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"configuring tests for cockroach only",
			fmt.Sprintf(
				"echo \"%s\" > /mnt/data1/pgjdbc/build.local.properties", pgjdbcDatabaseParams,
			),
		); err != nil {
			__antithesis_instrumentation__.Notify(49747)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49748)
		}
		__antithesis_instrumentation__.Notify(49719)

		t.Status("building pgjdbc (without tests)")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"building pgjdbc (without tests)",
			`cd /mnt/data1/pgjdbc/pgjdbc/ && ../gradlew test --tests OidToStringTest`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49749)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49750)
		}
		__antithesis_instrumentation__.Notify(49720)

		blocklistName, expectedFailures, ignorelistName, ignorelist := pgjdbcBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(49751)
			t.Fatalf("No pgjdbc blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49752)
		}
		__antithesis_instrumentation__.Notify(49721)
		status := fmt.Sprintf("Running cockroach version %s, using blocklist %s", version, blocklistName)
		if ignorelist != nil {
			__antithesis_instrumentation__.Notify(49753)
			status = fmt.Sprintf("Running cockroach version %s, using blocklist %s, using ignorelist %s",
				version, blocklistName, ignorelistName)
		} else {
			__antithesis_instrumentation__.Notify(49754)
		}
		__antithesis_instrumentation__.Notify(49722)
		t.L().Printf("%s", status)

		t.Status("running pgjdbc test suite")

		_ = c.RunE(ctx, node,
			`cd /mnt/data1/pgjdbc/pgjdbc/ && ../gradlew test`,
		)

		_ = c.RunE(ctx, node,
			`mkdir -p ~/logs/report/pgjdbc-results`,
		)

		t.Status("collecting the test results")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"copy test result files",
			`cp /mnt/data1/pgjdbc/pgjdbc/build/test-results/test/ ~/logs/report/pgjdbc-results -a`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49755)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49756)
		}
		__antithesis_instrumentation__.Notify(49723)

		result, err := repeatRunWithDetailsSingleNode(
			ctx,
			c,
			t,
			node,
			"get list of test files",
			`ls /mnt/data1/pgjdbc/pgjdbc/build/test-results/test/*.xml`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(49757)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49758)
		}
		__antithesis_instrumentation__.Notify(49724)

		if len(result.Stdout) == 0 {
			__antithesis_instrumentation__.Notify(49759)
			t.Fatal("could not find any test result files")
		} else {
			__antithesis_instrumentation__.Notify(49760)
		}
		__antithesis_instrumentation__.Notify(49725)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "pgjdbc", []byte(result.Stdout),
			blocklistName, expectedFailures, ignorelist, version, supportedPGJDBCTag,
		)
	}
	__antithesis_instrumentation__.Notify(49707)

	r.Add(registry.TestSpec{
		Name:    "pgjdbc",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `driver`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49761)
			runPgjdbc(ctx, t, c)
		},
	})
}

const pgjdbcDatabaseParams = `
server=localhost
port=26257
secondaryServer=localhost
secondaryPort=5433
secondaryServer2=localhost
secondaryServerPort2=5434
database=defaultdb
username=test_admin
password=testpw
privilegedUser=test_admin
privilegedPassword=testpw
sspiusername=testsspi
preparethreshold=5
loggerLevel=DEBUG
loggerFile=target/pgjdbc-tests.log
protocolVersion=0
sslpassword=sslpwd
`
