package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

var sqlAlchemyResultRegex = regexp.MustCompile(`^(?P<test>test.*::.*::[^ \[\]]*(?:\[.*])?) (?P<result>\w+)\s+\[.+]$`)
var sqlAlchemyReleaseTagRegex = regexp.MustCompile(`^rel_(?P<major>\d+)_(?P<minor>\d+)_(?P<point>\d+)$`)

var supportedSQLAlchemyTag = "rel_1_4_26"

func registerSQLAlchemy(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50970)
	r.Add(registry.TestSpec{
		Name:    "sqlalchemy",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50971)
			runSQLAlchemy(ctx, t, c)
		},
	})
}

func runSQLAlchemy(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(50972)
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(50991)
		t.Fatal("cannot be run in local mode")
	} else {
		__antithesis_instrumentation__.Notify(50992)
	}
	__antithesis_instrumentation__.Notify(50973)

	node := c.Node(1)

	t.Status("cloning sqlalchemy and installing prerequisites")
	latestTag, err := repeatGetLatestTag(ctx, t, "sqlalchemy", "sqlalchemy", sqlAlchemyReleaseTagRegex)
	if err != nil {
		__antithesis_instrumentation__.Notify(50993)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50994)
	}
	__antithesis_instrumentation__.Notify(50974)
	t.L().Printf("Latest sqlalchemy release is %s.", latestTag)
	t.L().Printf("Supported sqlalchemy release is %s.", supportedSQLAlchemyTag)

	if err := repeatRunE(ctx, t, c, node, "update apt-get", `
		sudo add-apt-repository ppa:deadsnakes/ppa &&
		sudo apt-get -qq update
	`); err != nil {
		__antithesis_instrumentation__.Notify(50995)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50996)
	}
	__antithesis_instrumentation__.Notify(50975)

	if err := repeatRunE(ctx, t, c, node, "install dependencies", `
		sudo apt-get -qq install make python3.7 libpq-dev python3.7-dev gcc python3-setuptools python-setuptools build-essential python3.7-distutils
	`); err != nil {
		__antithesis_instrumentation__.Notify(50997)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(50998)
	}
	__antithesis_instrumentation__.Notify(50976)

	if err := repeatRunE(ctx, t, c, node, "set python3.7 as default", `
		sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.5 1
		sudo update-alternatives --install /usr/bin/python3 python3 /usr/bin/python3.7 2
		sudo update-alternatives --config python3
	`); err != nil {
		__antithesis_instrumentation__.Notify(50999)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51000)
	}
	__antithesis_instrumentation__.Notify(50977)

	if err := repeatRunE(ctx, t, c, node, "install pip", `
		curl https://bootstrap.pypa.io/get-pip.py | sudo -H python3.7
	`); err != nil {
		__antithesis_instrumentation__.Notify(51001)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51002)
	}
	__antithesis_instrumentation__.Notify(50978)

	if err := repeatRunE(ctx, t, c, node, "install pytest", `
		sudo pip3 install --upgrade --force-reinstall setuptools pytest==6.0.1 pytest-xdist psycopg2 alembic
	`); err != nil {
		__antithesis_instrumentation__.Notify(51003)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51004)
	}
	__antithesis_instrumentation__.Notify(50979)

	if err := repeatRunE(ctx, t, c, node, "remove old sqlalchemy-cockroachdb", `
		sudo rm -rf /mnt/data1/sqlalchemy-cockroachdb
	`); err != nil {
		__antithesis_instrumentation__.Notify(51005)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51006)
	}
	__antithesis_instrumentation__.Notify(50980)

	if err := repeatGitCloneE(ctx, t, c,
		"https://github.com/cockroachdb/sqlalchemy-cockroachdb.git", "/mnt/data1/sqlalchemy-cockroachdb",
		"master", node); err != nil {
		__antithesis_instrumentation__.Notify(51007)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51008)
	}
	__antithesis_instrumentation__.Notify(50981)

	t.Status("installing sqlalchemy-cockroachdb")
	if err := repeatRunE(ctx, t, c, node, "installing sqlalchemy=cockroachdb", `
		cd /mnt/data1/sqlalchemy-cockroachdb && sudo pip3 install .
	`); err != nil {
		__antithesis_instrumentation__.Notify(51009)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51010)
	}
	__antithesis_instrumentation__.Notify(50982)

	if err := repeatRunE(ctx, t, c, node, "remove old sqlalchemy", `
		sudo rm -rf /mnt/data1/sqlalchemy
	`); err != nil {
		__antithesis_instrumentation__.Notify(51011)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51012)
	}
	__antithesis_instrumentation__.Notify(50983)

	if err := repeatGitCloneE(ctx, t, c,
		"https://github.com/sqlalchemy/sqlalchemy.git", "/mnt/data1/sqlalchemy",
		supportedSQLAlchemyTag, node); err != nil {
		__antithesis_instrumentation__.Notify(51013)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51014)
	}
	__antithesis_instrumentation__.Notify(50984)

	t.Status("building sqlalchemy")
	if err := repeatRunE(ctx, t, c, node, "building sqlalchemy", `
		cd /mnt/data1/sqlalchemy && python3 setup.py build
	`); err != nil {
		__antithesis_instrumentation__.Notify(51015)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51016)
	}
	__antithesis_instrumentation__.Notify(50985)

	t.Status("setting up cockroach")
	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

	version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(51017)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51018)
	}
	__antithesis_instrumentation__.Notify(50986)

	if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
		__antithesis_instrumentation__.Notify(51019)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(51020)
	}
	__antithesis_instrumentation__.Notify(50987)

	blocklistName, expectedFailures, ignoredlistName, ignoredlist := sqlAlchemyBlocklists.getLists(version)
	if expectedFailures == nil {
		__antithesis_instrumentation__.Notify(51021)
		t.Fatalf("No sqlalchemy blocklist defined for cockroach version %s", version)
	} else {
		__antithesis_instrumentation__.Notify(51022)
	}
	__antithesis_instrumentation__.Notify(50988)
	t.L().Printf("Running cockroach version %s, using blocklist %s, using ignoredlist %s",
		version, blocklistName, ignoredlistName)

	t.Status("running sqlalchemy test suite")

	result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
		`cd /mnt/data1/sqlalchemy-cockroachdb/ && pytest --maxfail=0 \
		--dburi='cockroachdb://root@localhost:26257/defaultdb?sslmode=disable&disable_cockroachdb_telemetry=true' \
		test/test_suite_sqlalchemy.py
	`)

	if err != nil {
		__antithesis_instrumentation__.Notify(51023)

		commandError := (*install.NonZeroExitCode)(nil)
		if !errors.As(err, &commandError) {
			__antithesis_instrumentation__.Notify(51024)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51025)
		}
	} else {
		__antithesis_instrumentation__.Notify(51026)
	}
	__antithesis_instrumentation__.Notify(50989)

	rawResults := []byte(result.Stdout + result.Stderr)

	t.Status("collating the test results")
	t.L().Printf("Test Results: %s", rawResults)

	results := newORMTestsResults()

	scanner := bufio.NewScanner(bytes.NewReader(rawResults))
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(51027)
		match := sqlAlchemyResultRegex.FindStringSubmatch(scanner.Text())
		if match == nil {
			__antithesis_instrumentation__.Notify(51030)
			continue
		} else {
			__antithesis_instrumentation__.Notify(51031)
		}
		__antithesis_instrumentation__.Notify(51028)
		test, result := match[1], match[2]
		pass := result == "PASSED" || func() bool {
			__antithesis_instrumentation__.Notify(51032)
			return strings.Contains(result, "failed as expected") == true
		}() == true
		skipped := result == "SKIPPED"
		results.allTests = append(results.allTests, test)

		ignoredIssue, expectedIgnored := ignoredlist[test]
		issue, expectedFailure := expectedFailures[test]
		switch {
		case expectedIgnored:
			__antithesis_instrumentation__.Notify(51033)
			results.results[test] = fmt.Sprintf("--- SKIP: %s due to %s (expected)", test, ignoredIssue)
			results.ignoredCount++
		case skipped && func() bool {
			__antithesis_instrumentation__.Notify(51041)
			return expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(51034)
			results.results[test] = fmt.Sprintf("--- SKIP: %s (unexpected)", test)
			results.unexpectedSkipCount++
		case skipped:
			__antithesis_instrumentation__.Notify(51035)
			results.results[test] = fmt.Sprintf("--- SKIP: %s (expected)", test)
			results.skipCount++
		case pass && func() bool {
			__antithesis_instrumentation__.Notify(51042)
			return !expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(51036)
			results.results[test] = fmt.Sprintf("--- PASS: %s (expected)", test)
			results.passExpectedCount++
		case pass && func() bool {
			__antithesis_instrumentation__.Notify(51043)
			return expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(51037)
			results.results[test] = fmt.Sprintf("--- PASS: %s - %s (unexpected)",
				test, maybeAddGithubLink(issue),
			)
			results.passUnexpectedCount++
		case !pass && func() bool {
			__antithesis_instrumentation__.Notify(51044)
			return expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(51038)
			results.results[test] = fmt.Sprintf("--- FAIL: %s - %s (expected)",
				test, maybeAddGithubLink(issue),
			)
			results.failExpectedCount++
			results.currentFailures = append(results.currentFailures, test)
		case !pass && func() bool {
			__antithesis_instrumentation__.Notify(51045)
			return !expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(51039)
			results.results[test] = fmt.Sprintf("--- FAIL: %s (unexpected)", test)
			results.failUnexpectedCount++
			results.currentFailures = append(results.currentFailures, test)
		default:
			__antithesis_instrumentation__.Notify(51040)
		}
		__antithesis_instrumentation__.Notify(51029)
		results.runTests[test] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(50990)

	results.summarizeAll(
		t, "sqlalchemy", blocklistName, expectedFailures, version, supportedSQLAlchemyTag)
}
