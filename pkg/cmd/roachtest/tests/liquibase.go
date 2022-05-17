package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

var supportedLiquibaseHarnessCommit = "1790ddef2d0339c5c96839ac60ac424c130dadd8"

func registerLiquibase(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49205)
	runLiquibase := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(49207)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49221)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49222)
		}
		__antithesis_instrumentation__.Notify(49208)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49223)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49224)
		}
		__antithesis_instrumentation__.Notify(49209)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49225)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49226)
		}
		__antithesis_instrumentation__.Notify(49210)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49227)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49228)
		}
		__antithesis_instrumentation__.Notify(49211)

		t.Status("cloning liquibase test harness and installing prerequisites")
		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49229)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49230)
		}
		__antithesis_instrumentation__.Notify(49212)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install default-jre openjdk-8-jdk-headless maven`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49231)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49232)
		}
		__antithesis_instrumentation__.Notify(49213)

		if err := repeatRunE(
			ctx, t, c, node, "remove old liquibase test harness",
			`rm -rf /mnt/data1/liquibase-test-harness`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49233)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49234)
		}
		__antithesis_instrumentation__.Notify(49214)

		if err = c.RunE(ctx, node, "cd /mnt/data1/ && git clone https://github.com/liquibase/liquibase-test-harness.git"); err != nil {
			__antithesis_instrumentation__.Notify(49235)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49236)
		}
		__antithesis_instrumentation__.Notify(49215)
		if err = c.RunE(ctx, node, fmt.Sprintf("cd /mnt/data1/liquibase-test-harness/ && git checkout %s",
			supportedLiquibaseHarnessCommit)); err != nil {
			__antithesis_instrumentation__.Notify(49237)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49238)
		}
		__antithesis_instrumentation__.Notify(49216)

		t.Status("creating database/user used by tests")
		if err = c.RunE(ctx, node, `sudo mkdir /cockroach && sudo ln -sf /home/ubuntu/cockroach /cockroach/cockroach.sh`); err != nil {
			__antithesis_instrumentation__.Notify(49239)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49240)
		}
		__antithesis_instrumentation__.Notify(49217)
		if err = c.RunE(ctx, node, `/mnt/data1/liquibase-test-harness/src/test/resources/docker/setup_db.sh localhost`); err != nil {
			__antithesis_instrumentation__.Notify(49241)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49242)
		}
		__antithesis_instrumentation__.Notify(49218)

		t.Status("running liquibase test harness")
		blocklistName, expectedFailures, _, ignoreList := liquibaseBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(49243)
			t.Fatalf("No hibernate blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49244)
		}
		__antithesis_instrumentation__.Notify(49219)

		const (
			repoDir     = "/mnt/data1/liquibase-test-harness"
			resultsPath = repoDir + "/target/surefire-reports/TEST-liquibase.harness.LiquibaseHarnessSuiteTest.xml"
		)

		cmd := fmt.Sprintf("cd /mnt/data1/liquibase-test-harness/ && "+
			"mvn surefire-report:report-only test -Dtest=LiquibaseHarnessSuiteTest "+
			"-DdbName=cockroachdb -DdbVersion=20.2 -DoutputDirectory=%s", repoDir)

		err = c.RunE(ctx, node, cmd)
		if err != nil {
			__antithesis_instrumentation__.Notify(49245)
			t.L().Printf("error whilst running tests (may be expected): %#v", err)
		} else {
			__antithesis_instrumentation__.Notify(49246)
		}
		__antithesis_instrumentation__.Notify(49220)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "liquibase", []byte(resultsPath),
			blocklistName,
			expectedFailures,
			ignoreList,
			version,
			supportedLiquibaseHarnessCommit,
		)

	}
	__antithesis_instrumentation__.Notify(49206)

	r.Add(registry.TestSpec{
		Name:    "liquibase",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `tool`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49247)
			runLiquibase(ctx, t, c)
		},
	})
}
