package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

var libPQReleaseTagRegex = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)$`)
var libPQSupportedTag = "v1.10.5"

func registerLibPQ(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49180)
	runLibPQ := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(49182)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(49189)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(49190)
		}
		__antithesis_instrumentation__.Notify(49183)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(49191)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49192)
		}
		__antithesis_instrumentation__.Notify(49184)
		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(49193)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49194)
		}
		__antithesis_instrumentation__.Notify(49185)

		t.Status("cloning lib/pq and installing prerequisites")
		latestTag, err := repeatGetLatestTag(
			ctx, t, "lib", "pq", libPQReleaseTagRegex)
		require.NoError(t, err)
		t.L().Printf("Latest lib/pq release is %s.", latestTag)
		t.L().Printf("Supported lib/pq release is %s.", libPQSupportedTag)

		installGolang(ctx, t, c, node)

		const (
			libPQRepo   = "github.com/lib/pq"
			libPQPath   = goPath + "/src/" + libPQRepo
			resultsDir  = "~/logs/report/libpq-results"
			resultsPath = resultsDir + "/report.xml"
		)

		err = repeatRunE(
			ctx, t, c, node, "remove old lib/pq", fmt.Sprintf("rm -rf %s", libPQPath),
		)
		require.NoError(t, err)

		err = repeatRunE(ctx, t, c, node, "install go-junit-report",
			fmt.Sprintf("GOPATH=%s go get -u github.com/jstemmer/go-junit-report", goPath),
		)
		require.NoError(t, err)

		err = repeatGitCloneE(
			ctx,
			t,
			c,
			fmt.Sprintf("https://%s.git", libPQRepo),
			libPQPath,
			libPQSupportedTag,
			node,
		)
		require.NoError(t, err)
		if err := c.RunE(ctx, node, fmt.Sprintf("mkdir -p %s", resultsDir)); err != nil {
			__antithesis_instrumentation__.Notify(49195)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49196)
		}
		__antithesis_instrumentation__.Notify(49186)

		blocklistName, expectedFailures, ignorelistName, ignoreList := libPQBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(49197)
			t.Fatalf("No lib/pq blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(49198)
		}
		__antithesis_instrumentation__.Notify(49187)
		t.L().Printf("Running cockroach version %s, using blocklist %s, using ignorelist %s", version, blocklistName, ignorelistName)

		t.Status("running lib/pq test suite and collecting results")

		testListRegex := "^(Test|Example)"
		result, err := c.RunWithDetailsSingleNode(
			ctx, t.L(),
			node,
			fmt.Sprintf(`cd %s && PGPORT=26257 PGUSER=root PGSSLMODE=disable PGDATABASE=postgres go test -list "%s"`, libPQPath, testListRegex),
		)
		require.NoError(t, err)

		tests := strings.Fields(result.Stdout)
		var allowedTests []string
		testListR, err := regexp.Compile(testListRegex)
		require.NoError(t, err)

		for _, testName := range tests {
			__antithesis_instrumentation__.Notify(49199)

			if !testListR.MatchString(testName) {
				__antithesis_instrumentation__.Notify(49201)
				continue
			} else {
				__antithesis_instrumentation__.Notify(49202)
			}
			__antithesis_instrumentation__.Notify(49200)

			if _, ok := ignoreList[testName]; !ok {
				__antithesis_instrumentation__.Notify(49203)
				allowedTests = append(allowedTests, testName)
			} else {
				__antithesis_instrumentation__.Notify(49204)
			}
		}
		__antithesis_instrumentation__.Notify(49188)

		allowedTestsRegExp := fmt.Sprintf(`"^(%s)$"`, strings.Join(allowedTests, "|"))

		_ = c.RunE(
			ctx,
			node,
			fmt.Sprintf("cd %s && PGPORT=26257 PGUSER=root PGSSLMODE=disable PGDATABASE=postgres go test -run %s -v 2>&1 | %s/bin/go-junit-report > %s",
				libPQPath, allowedTestsRegExp, goPath, resultsPath),
		)

		parseAndSummarizeJavaORMTestsResults(
			ctx, t, c, node, "lib/pq", []byte(resultsPath),
			blocklistName, expectedFailures, ignoreList, version, libPQSupportedTag,
		)
	}
	__antithesis_instrumentation__.Notify(49181)

	r.Add(registry.TestSpec{
		Name:    "lib/pq",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `driver`},
		Run:     runLibPQ,
	})
}
