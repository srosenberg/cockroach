package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

var rubyPGTestFailureRegex = regexp.MustCompile(`^rspec ./.*# .*`)
var testFailureFilenameRegexp = regexp.MustCompile("^rspec .*.rb.*([0-9]|]) # ")
var testSummaryRegexp = regexp.MustCompile("^([0-9]+) examples, [0-9]+ failures")
var rubyPGVersion = "v1.2.3"

func registerRubyPG(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50427)
	runRubyPGTest := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(50429)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(50446)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(50447)
		}
		__antithesis_instrumentation__.Notify(50430)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		if err := c.PutLibraries(ctx, "./lib"); err != nil {
			__antithesis_instrumentation__.Notify(50448)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50449)
		}
		__antithesis_instrumentation__.Notify(50431)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(50450)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50451)
		}
		__antithesis_instrumentation__.Notify(50432)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(50452)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50453)
		}
		__antithesis_instrumentation__.Notify(50433)

		t.Status("cloning rails and installing prerequisites")

		t.L().Printf("Supported ruby-pg version is %s.", rubyPGVersion)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50454)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50455)
		}
		__antithesis_instrumentation__.Notify(50434)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install ruby-full ruby-dev rubygems build-essential zlib1g-dev libpq-dev libsqlite3-dev`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50456)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50457)
		}
		__antithesis_instrumentation__.Notify(50435)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install ruby 2.7",
			`mkdir -p ruby-install && \
        curl -fsSL https://github.com/postmodern/ruby-install/archive/v0.6.1.tar.gz | tar --strip-components=1 -C ruby-install -xz && \
        sudo make -C ruby-install install && \
        sudo ruby-install --system ruby 2.7.1 && \
        sudo gem update --system`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50458)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50459)
		}
		__antithesis_instrumentation__.Notify(50436)

		if err := repeatRunE(
			ctx, t, c, node, "remove old ruby-pg", `sudo rm -rf /mnt/data1/ruby-pg`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50460)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50461)
		}
		__antithesis_instrumentation__.Notify(50437)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/ged/ruby-pg.git",
			"/mnt/data1/ruby-pg",
			rubyPGVersion,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(50462)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50463)
		}
		__antithesis_instrumentation__.Notify(50438)

		t.Status("installing bundler")
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"installing bundler",
			`cd /mnt/data1/ruby-pg/ && sudo gem install bundler:2.1.4`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50464)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50465)
		}
		__antithesis_instrumentation__.Notify(50439)

		t.Status("installing gems")
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"installing gems",
			`cd /mnt/data1/ruby-pg/ && sudo bundle install`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50466)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50467)
		}
		__antithesis_instrumentation__.Notify(50440)

		if err := repeatRunE(
			ctx, t, c, node, "remove old ruby-pg helpers.rb", `sudo rm /mnt/data1/ruby-pg/spec/helpers.rb`,
		); err != nil {
			__antithesis_instrumentation__.Notify(50468)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50469)
		}
		__antithesis_instrumentation__.Notify(50441)

		rubyPGHelpersFile := "./pkg/cmd/roachtest/tests/ruby_pg_helpers.rb"
		err = c.PutE(ctx, t.L(), rubyPGHelpersFile, "/mnt/data1/ruby-pg/spec/helpers.rb", c.All())
		require.NoError(t, err)

		t.Status("running ruby-pg test suite")

		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			`cd /mnt/data1/ruby-pg/ && sudo rake`,
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(50470)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(50471)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(50472)
			}
		} else {
			__antithesis_instrumentation__.Notify(50473)
		}
		__antithesis_instrumentation__.Notify(50442)

		rawResults := []byte(result.Stdout + result.Stderr)

		t.L().Printf("Test Results:\n%s", rawResults)

		results := newORMTestsResults()
		blocklistName, expectedFailures, _, _ := rubyPGBlocklist.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(50474)
			t.Fatalf("No ruby-pg blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(50475)
		}
		__antithesis_instrumentation__.Notify(50443)

		scanner := bufio.NewScanner(bytes.NewReader(rawResults))
		totalTests := int64(0)
		for scanner.Scan() {
			__antithesis_instrumentation__.Notify(50476)
			line := scanner.Text()
			testSummaryMatch := testSummaryRegexp.FindStringSubmatch(line)
			if testSummaryMatch != nil {
				__antithesis_instrumentation__.Notify(50482)
				totalTests, err = strconv.ParseInt(testSummaryMatch[1], 10, 64)
				require.NoError(t, err)
				continue
			} else {
				__antithesis_instrumentation__.Notify(50483)
			}
			__antithesis_instrumentation__.Notify(50477)

			match := rubyPGTestFailureRegex.FindStringSubmatch(line)
			if match == nil {
				__antithesis_instrumentation__.Notify(50484)
				continue
			} else {
				__antithesis_instrumentation__.Notify(50485)
			}
			__antithesis_instrumentation__.Notify(50478)
			if len(match) != 1 {
				__antithesis_instrumentation__.Notify(50486)
				log.Fatalf(ctx, "expected one match for test name, found: %d", len(match))
			} else {
				__antithesis_instrumentation__.Notify(50487)
			}
			__antithesis_instrumentation__.Notify(50479)

			test := match[0]

			strs := testFailureFilenameRegexp.Split(test, -1)
			if len(strs) != 2 {
				__antithesis_instrumentation__.Notify(50488)
				log.Fatalf(ctx, "expected test output line to be split into two strings")
			} else {
				__antithesis_instrumentation__.Notify(50489)
			}
			__antithesis_instrumentation__.Notify(50480)
			test = strs[1]

			issue, expectedFailure := expectedFailures[test]
			switch {
			case expectedFailure:
				__antithesis_instrumentation__.Notify(50490)
				results.results[test] = fmt.Sprintf("--- FAIL: %s - %s (expected)",
					test, maybeAddGithubLink(issue),
				)
				results.failExpectedCount++
				results.currentFailures = append(results.currentFailures, test)
			case !expectedFailure:
				__antithesis_instrumentation__.Notify(50491)
				results.results[test] = fmt.Sprintf("--- FAIL: %s - %s (unexpected)",
					test, maybeAddGithubLink(issue),
				)
				results.failUnexpectedCount++
				results.currentFailures = append(results.currentFailures, test)
			default:
				__antithesis_instrumentation__.Notify(50492)
			}
			__antithesis_instrumentation__.Notify(50481)
			results.runTests[test] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(50444)

		if totalTests == 0 {
			__antithesis_instrumentation__.Notify(50493)
			log.Fatalf(ctx, "failed to find total number of tests run")
		} else {
			__antithesis_instrumentation__.Notify(50494)
		}
		__antithesis_instrumentation__.Notify(50445)
		totalPasses := int(totalTests) - (results.failUnexpectedCount + results.failExpectedCount)
		results.passUnexpectedCount = len(expectedFailures) - results.failExpectedCount
		results.passExpectedCount = totalPasses - results.passUnexpectedCount

		results.summarizeAll(t, "ruby-pg", blocklistName, expectedFailures, version, rubyPGVersion)
	}
	__antithesis_instrumentation__.Notify(50428)

	r.Add(registry.TestSpec{
		Name:    "ruby-pg",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50495)
			runRubyPGTest(ctx, t, c)
		},
	})
}
