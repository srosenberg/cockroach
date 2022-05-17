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
	"github.com/stretchr/testify/require"
)

var gormReleaseTag = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var gormSupportedTag = "v1.23.1"

func registerGORM(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48097)
	runGORM := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(48099)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48110)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(48111)
		}
		__antithesis_instrumentation__.Notify(48100)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())
		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(48112)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48113)
		}
		__antithesis_instrumentation__.Notify(48101)
		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(48114)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48115)
		}
		__antithesis_instrumentation__.Notify(48102)

		t.Status("cloning gorm and installing prerequisites")
		latestTag, err := repeatGetLatestTag(
			ctx, t, "go-gorm", "gorm", gormReleaseTag)
		if err != nil {
			__antithesis_instrumentation__.Notify(48116)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48117)
		}
		__antithesis_instrumentation__.Notify(48103)
		t.L().Printf("Latest gorm release is %s.", latestTag)
		t.L().Printf("Supported gorm release is %s.", gormSupportedTag)

		installGolang(ctx, t, c, node)

		const (
			gormRepo     = "github.com/go-gorm/gorm"
			gormPath     = goPath + "/src/" + gormRepo
			gormTestPath = gormPath + "/tests/"
			resultsDir   = "~/logs/report/gorm"
			resultsPath  = resultsDir + "/report.xml"
		)

		if err := repeatRunE(
			ctx, t, c, node, "remove old gorm", fmt.Sprintf("rm -rf %s", gormPath),
		); err != nil {
			__antithesis_instrumentation__.Notify(48118)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48119)
		}
		__antithesis_instrumentation__.Notify(48104)

		if err := repeatRunE(
			ctx, t, c, node, "install go-junit-report", fmt.Sprintf("GOPATH=%s go get -u github.com/jstemmer/go-junit-report", goPath),
		); err != nil {
			__antithesis_instrumentation__.Notify(48120)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48121)
		}
		__antithesis_instrumentation__.Notify(48105)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			fmt.Sprintf("https://%s.git", gormRepo),
			gormPath,
			gormSupportedTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(48122)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48123)
		}
		__antithesis_instrumentation__.Notify(48106)

		if err := c.RunE(ctx, node, fmt.Sprintf("mkdir -p %s", resultsDir)); err != nil {
			__antithesis_instrumentation__.Notify(48124)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48125)
		}
		__antithesis_instrumentation__.Notify(48107)

		blocklistName, expectedFailures, ignorelistName, ignoredFailures := gormBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(48126)
			t.Fatalf("No gorm blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(48127)
		}
		__antithesis_instrumentation__.Notify(48108)
		t.L().Printf("Running cockroach version %s, using blocklist %s, using ignorelist %s", version, blocklistName, ignorelistName)

		err = c.RunE(ctx, node, `./cockroach sql -e "CREATE DATABASE gorm" --insecure`)
		require.NoError(t, err)

		t.Status("downloading go dependencies for tests")
		err = c.RunE(
			ctx,
			node,
			fmt.Sprintf(`cd %s && go get -u -t ./... && go mod download && go mod tidy `, gormTestPath),
		)
		require.NoError(t, err)

		t.Status("running gorm test suite and collecting results")

		err = c.RunE(
			ctx,
			node,
			fmt.Sprintf(`cd %s && GORM_DIALECT="postgres" GORM_DSN="user=root password= dbname=gorm host=localhost port=26257 sslmode=disable"
				go test -v ./... 2>&1 | %s/bin/go-junit-report > %s`,
				gormTestPath, goPath, resultsPath),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48128)
			t.L().Printf("error whilst running tests (may be expected): %#v", err)
		} else {
			__antithesis_instrumentation__.Notify(48129)
		}
		__antithesis_instrumentation__.Notify(48109)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "gorm", []byte(resultsPath),
			blocklistName, expectedFailures, ignoredFailures, version, latestTag,
		)
	}
	__antithesis_instrumentation__.Notify(48098)

	r.Add(registry.TestSpec{
		Name:    "gorm",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run:     runGORM,
	})
}
