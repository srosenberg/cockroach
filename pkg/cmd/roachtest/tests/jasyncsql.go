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

var supportedJasyncCommit = "6301aa1b9ef8a0d4c5cf6f3c095b30a388c62dc0"

func registerJasyncSQL(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48610)
	runJasyncSQL := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(48612)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48625)
			t.Fatal("can not be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(48626)
		}
		__antithesis_instrumentation__.Notify(48613)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(48627)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48628)
		}
		__antithesis_instrumentation__.Notify(48614)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(48629)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48630)
		}
		__antithesis_instrumentation__.Notify(48615)

		t.Status("cloning jasync-sql and installing prerequisites")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"remove old jasync-sql",
			`rm -rf /mnt/data1/jasyncsql`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48631)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48632)
		}
		__antithesis_instrumentation__.Notify(48616)

		if err := c.RunE(
			ctx,
			node,
			"cd /mnt/data1 && git clone https://github.com/jasync-sql/jasync-sql.git",
		); err != nil {
			__antithesis_instrumentation__.Notify(48633)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48634)
		}
		__antithesis_instrumentation__.Notify(48617)

		if err := c.RunE(ctx, node, fmt.Sprintf("cd /mnt/data1/jasync-sql && git checkout %s",
			supportedJasyncCommit)); err != nil {
			__antithesis_instrumentation__.Notify(48635)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48636)
		}
		__antithesis_instrumentation__.Notify(48618)
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install java and gradle",
			`sudo apt-get -qq install default-jre openjdk-11-jdk-headless gradle`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48637)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48638)
		}
		__antithesis_instrumentation__.Notify(48619)
		t.Status("building jasyncsql (without tests)")

		blocklistName, expectedFailures, ignorelistName, ignorelist := jasyncsqlBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(48639)
			t.Fatalf("No jasyncsql blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(48640)
		}
		__antithesis_instrumentation__.Notify(48620)
		status := fmt.Sprintf("running cockraoch version %s, using blocklist %s", version, blocklistName)
		if ignorelist != nil {
			__antithesis_instrumentation__.Notify(48641)
			status = fmt.Sprintf(
				"Running cockroach %s, using blocklist %s, using ignorelist %s",
				version,
				blocklistName,
				ignorelistName)
		} else {
			__antithesis_instrumentation__.Notify(48642)
		}
		__antithesis_instrumentation__.Notify(48621)
		t.L().Printf("%s", status)

		t.Status("running jasyncsql test suite")

		_ = c.RunE(
			ctx,
			node,
			`cd /mnt/data1/jasync-sql && PGUSER=root PGHOST=localhost PGPORT=26257 PGDATABASE=defaultdb ./gradlew :postgresql-async:test`,
		)

		_ = c.RunE(ctx, node, `mkdir -p ~/logs/report/jasyncsql-results`)

		t.Status("making test directory")

		_ = c.RunE(ctx, node,
			`mkdir -p ~/logs/report/jasyncsql-results`,
		)

		t.Status("collecting the test results")

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"copy test result files",
			`cp /mnt/data1/jasync-sql/postgresql-async/build/test-results/test/*.xml ~/logs/report/jasyncsql-results -a`,
		); err != nil {
			__antithesis_instrumentation__.Notify(48643)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48644)
		}
		__antithesis_instrumentation__.Notify(48622)

		result, err := repeatRunWithDetailsSingleNode(
			ctx,
			c,
			t,
			node,
			"get list of test files",
			`ls /mnt/data1/jasync-sql/postgresql-async/build/test-results/test/*xml`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48645)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48646)
		}
		__antithesis_instrumentation__.Notify(48623)
		if len(result.Stdout) == 0 {
			__antithesis_instrumentation__.Notify(48647)
			t.Fatal("could not find any test result files")
		} else {
			__antithesis_instrumentation__.Notify(48648)
		}
		__antithesis_instrumentation__.Notify(48624)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "jasyncsql", []byte(result.Stdout),
			blocklistName, expectedFailures, ignorelist, version, supportedJasyncCommit)
	}
	__antithesis_instrumentation__.Notify(48611)

	r.Add(registry.TestSpec{
		Name:    "jasync",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48649)
			runJasyncSQL(ctx, t, c)
		},
	})
}
