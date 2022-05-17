package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"regexp"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

var psycopgReleaseTagRegex = regexp.MustCompile(`^(?P<major>\d+)(?:_(?P<minor>\d+)(?:_(?P<point>\d+)(?:_(?P<subpoint>\d+))?)?)?$`)
var supportedPsycopgTag = "2_8_6"

func registerPsycopg(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49829)
	runPsycopg := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(49831)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49844)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49845)
		}
		__antithesis_instrumentation__.Notify(49832)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49846)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49847)
		}
		__antithesis_instrumentation__.Notify(49833)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49848)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49849)
		}
		__antithesis_instrumentation__.Notify(49834)

		t.Status("cloning psycopg and installing prerequisites")
		latestTag, err := repeatGetLatestTag(ctx, t, "psycopg", "psycopg2", psycopgReleaseTagRegex)
		if err != nil {
			__antithesis_instrumentation__.Notify(49850)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49851)
		}
		__antithesis_instrumentation__.Notify(49835)
		t.L().Printf("Latest Psycopg release is %s.", latestTag)
		t.L().Printf("Supported Psycopg release is %s.", supportedPsycopgTag)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49852)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49853)
		}
		__antithesis_instrumentation__.Notify(49836)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install make python3 libpq-dev python-dev gcc python3-setuptools python-setuptools`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49854)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49855)
		}
		__antithesis_instrumentation__.Notify(49837)

		if err := repeatRunE(
			ctx, t, c, node, "remove old Psycopg", `sudo rm -rf /mnt/data1/psycopg`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49856)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49857)
		}
		__antithesis_instrumentation__.Notify(49838)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/psycopg/psycopg2.git",
			"/mnt/data1/psycopg",
			supportedPsycopgTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(49858)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49859)
		}
		__antithesis_instrumentation__.Notify(49839)

		t.Status("building Psycopg")
		if err := repeatRunE(
			ctx, t, c, node, "building Psycopg", `cd /mnt/data1/psycopg/ && make`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49860)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49861)
		}
		__antithesis_instrumentation__.Notify(49840)

		blocklistName, expectedFailures, ignoredlistName, ignoredlist := psycopgBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(49862)
			t.Fatalf("No psycopg blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49863)
		}
		__antithesis_instrumentation__.Notify(49841)
		if ignoredlist == nil {
			__antithesis_instrumentation__.Notify(49864)
			t.Fatalf("No psycopg ignorelist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49865)
		}
		__antithesis_instrumentation__.Notify(49842)
		t.L().Printf("Running cockroach version %s, using blocklist %s, using ignoredlist %s",
			version, blocklistName, ignoredlistName)

		t.Status("running psycopg test suite")

		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			`cd /mnt/data1/psycopg/ &&
			export PSYCOPG2_TESTDB=defaultdb &&
			export PSYCOPG2_TESTDB_USER=root &&
			export PSYCOPG2_TESTDB_PORT=26257 &&
			export PSYCOPG2_TESTDB_HOST=localhost &&
			make check`,
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(49866)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(49867)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49868)
			}
		} else {
			__antithesis_instrumentation__.Notify(49869)
		}
		__antithesis_instrumentation__.Notify(49843)

		rawResults := []byte(result.Stdout + result.Stderr)

		t.Status("collating the test results")
		t.L().Printf("Test Results: %s", rawResults)

		results := newORMTestsResults()
		results.parsePythonUnitTestOutput(rawResults, expectedFailures, ignoredlist)
		results.summarizeAll(
			t, "psycopg", blocklistName, expectedFailures,
			version, supportedPsycopgTag,
		)
	}
	__antithesis_instrumentation__.Notify(49830)

	r.Add(registry.TestSpec{
		Name:    "psycopg",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `driver`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49870)
			runPsycopg(ctx, t, c)
		},
	})
}
