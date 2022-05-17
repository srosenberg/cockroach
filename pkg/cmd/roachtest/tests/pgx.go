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
	"github.com/cockroachdb/errors"
)

var pgxReleaseTagRegex = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var supportedPGXTag = "v4.15.0"

func registerPgx(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49762)
	runPgx := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(49764)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49776)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49777)
		}
		__antithesis_instrumentation__.Notify(49765)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49778)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49779)
		}
		__antithesis_instrumentation__.Notify(49766)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49780)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49781)
		}
		__antithesis_instrumentation__.Notify(49767)

		t.Status("setting up go")
		installGolang(ctx, t, c, node)

		t.Status("getting pgx")
		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/jackc/pgx.git",
			"/mnt/data1/pgx",
			supportedPGXTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(49782)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49783)
		}
		__antithesis_instrumentation__.Notify(49768)

		latestTag, err := repeatGetLatestTag(ctx, t, "jackc", "pgx", pgxReleaseTagRegex)
		if err != nil {
			__antithesis_instrumentation__.Notify(49784)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49785)
		}
		__antithesis_instrumentation__.Notify(49769)
		t.L().Printf("Latest jackc/pgx release is %s.", latestTag)
		t.L().Printf("Supported release is %s.", supportedPGXTag)

		t.Status("installing go-junit-report")
		if err := repeatRunE(
			ctx, t, c, node, "install go-junit-report", "go get -u github.com/jstemmer/go-junit-report",
		); err != nil {
			__antithesis_instrumentation__.Notify(49786)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49787)
		}
		__antithesis_instrumentation__.Notify(49770)

		t.Status("checking blocklist")
		blocklistName, expectedFailures, ignorelistName, ignorelist := pgxBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(49788)
			t.Fatalf("No pgx blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49789)
		}
		__antithesis_instrumentation__.Notify(49771)
		status := fmt.Sprintf("Running cockroach version %s, using blocklist %s", version, blocklistName)
		if ignorelist != nil {
			__antithesis_instrumentation__.Notify(49790)
			status = fmt.Sprintf("Running cockroach version %s, using blocklist %s, using ignorelist %s",
				version, blocklistName, ignorelistName)
		} else {
			__antithesis_instrumentation__.Notify(49791)
		}
		__antithesis_instrumentation__.Notify(49772)
		t.L().Printf("%s", status)

		t.Status("setting up test db")
		db, err := c.ConnE(ctx, t.L(), node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49792)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49793)
		}
		__antithesis_instrumentation__.Notify(49773)
		defer db.Close()
		if _, err = db.ExecContext(
			ctx, `drop database if exists pgx_test; create database pgx_test;`,
		); err != nil {
			__antithesis_instrumentation__.Notify(49794)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49795)
		}
		__antithesis_instrumentation__.Notify(49774)

		_, _ = db.ExecContext(
			ctx, `create domain uint64 as numeric(20,0);`,
		)

		t.Status("running pgx test suite")

		result, err := repeatRunWithDetailsSingleNode(
			ctx, c, t, node,
			"run pgx test suite",
			"cd /mnt/data1/pgx && "+
				"PGX_TEST_DATABASE='postgresql://root:@localhost:26257/pgx_test' go test -v 2>&1 | "+
				"`go env GOPATH`/bin/go-junit-report",
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(49796)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(49797)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(49798)
			}
		} else {
			__antithesis_instrumentation__.Notify(49799)
		}
		__antithesis_instrumentation__.Notify(49775)

		xmlResults := []byte(result.Stdout + result.Stderr)

		results := newORMTestsResults()
		results.parseJUnitXML(t, expectedFailures, ignorelist, xmlResults)
		results.summarizeAll(
			t, "pgx", blocklistName, expectedFailures, version, supportedPGXTag,
		)
	}
	__antithesis_instrumentation__.Notify(49763)

	r.Add(registry.TestSpec{
		Name:    "pgx",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `driver`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49800)
			runPgx(ctx, t, c)
		},
	})
}
