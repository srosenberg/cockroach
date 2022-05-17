package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
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

var activerecordResultRegex = regexp.MustCompile(`^(?P<test>[^\s]+#[^\s]+) = (?P<timing>\d+\.\d+ s) = (?P<result>.)$`)
var railsReleaseTagRegex = regexp.MustCompile(`^v(?P<major>\d+)\.(?P<minor>\d+)\.(?P<point>\d+)\.?(?P<subpoint>\d*)$`)
var supportedRailsVersion = "6.1.5"
var activerecordAdapterVersion = "v6.1.8"

func registerActiveRecord(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45230)
	runActiveRecord := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(45232)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(45252)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(45253)
		}
		__antithesis_instrumentation__.Notify(45233)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		if err := c.PutLibraries(ctx, "./lib"); err != nil {
			__antithesis_instrumentation__.Notify(45254)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45255)
		}
		__antithesis_instrumentation__.Notify(45234)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(45256)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45257)
		}
		__antithesis_instrumentation__.Notify(45235)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(45258)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45259)
		}
		__antithesis_instrumentation__.Notify(45236)

		t.Status("creating database used by tests")
		db, err := c.ConnE(ctx, t.L(), node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(45260)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45261)
		}
		__antithesis_instrumentation__.Notify(45237)
		defer db.Close()

		if _, err := db.ExecContext(
			ctx, `CREATE DATABASE activerecord_unittest;`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45262)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45263)
		}
		__antithesis_instrumentation__.Notify(45238)

		if _, err := db.ExecContext(
			ctx, `CREATE DATABASE activerecord_unittest2;`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45264)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45265)
		}
		__antithesis_instrumentation__.Notify(45239)

		t.Status("cloning rails and installing prerequisites")

		latestTag, err := repeatGetLatestTag(
			ctx, t, "rails", "rails", railsReleaseTagRegex,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(45266)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45267)
		}
		__antithesis_instrumentation__.Notify(45240)
		t.L().Printf("Latest rails release is %s.", latestTag)
		t.L().Printf("Supported rails release is %s.", supportedRailsVersion)
		t.L().Printf("Supported adapter version is %s.", activerecordAdapterVersion)

		if err := repeatRunE(
			ctx, t, c, node, "update apt-get", `sudo apt-get -qq update`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45268)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45269)
		}
		__antithesis_instrumentation__.Notify(45241)

		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"install dependencies",
			`sudo apt-get -qq install ruby-full ruby-dev rubygems build-essential zlib1g-dev libpq-dev libsqlite3-dev`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45270)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45271)
		}
		__antithesis_instrumentation__.Notify(45242)

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
			__antithesis_instrumentation__.Notify(45272)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45273)
		}
		__antithesis_instrumentation__.Notify(45243)

		if err := repeatRunE(
			ctx, t, c, node, "remove old activerecord adapter", `rm -rf /mnt/data1/activerecord-cockroachdb-adapter`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45274)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45275)
		}
		__antithesis_instrumentation__.Notify(45244)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/cockroachdb/activerecord-cockroachdb-adapter.git",
			"/mnt/data1/activerecord-cockroachdb-adapter",
			activerecordAdapterVersion,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(45276)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45277)
		}
		__antithesis_instrumentation__.Notify(45245)

		t.Status("installing bundler")
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"installing bundler",
			`cd /mnt/data1/activerecord-cockroachdb-adapter/ && sudo gem install bundler:2.1.4`,
		); err != nil {
			__antithesis_instrumentation__.Notify(45278)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45279)
		}
		__antithesis_instrumentation__.Notify(45246)

		t.Status("installing gems")
		if err := repeatRunE(
			ctx,
			t,
			c,
			node,
			"installing gems",
			fmt.Sprintf(
				`cd /mnt/data1/activerecord-cockroachdb-adapter/ && `+
					`RAILS_VERSION=%s sudo bundle install`, supportedRailsVersion),
		); err != nil {
			__antithesis_instrumentation__.Notify(45280)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45281)
		}
		__antithesis_instrumentation__.Notify(45247)

		blocklistName, expectedFailures, ignorelistName, ignorelist := activeRecordBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(45282)
			t.Fatalf("No activerecord blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(45283)
		}
		__antithesis_instrumentation__.Notify(45248)
		status := fmt.Sprintf("Running cockroach version %s, using blocklist %s", version, blocklistName)
		if ignorelist != nil {
			__antithesis_instrumentation__.Notify(45284)
			status = fmt.Sprintf("Running cockroach version %s, using blocklist %s, using ignorelist %s",
				version, blocklistName, ignorelistName)
		} else {
			__antithesis_instrumentation__.Notify(45285)
		}
		__antithesis_instrumentation__.Notify(45249)
		t.L().Printf("%s", status)

		t.Status("running activerecord test suite")

		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			`cd /mnt/data1/activerecord-cockroachdb-adapter/ && `+
				`sudo RUBYOPT="-W0" TESTOPTS="-v" bundle exec rake test`,
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(45286)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(45287)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45288)
			}
		} else {
			__antithesis_instrumentation__.Notify(45289)
		}
		__antithesis_instrumentation__.Notify(45250)

		rawResults := []byte(result.Stdout + result.Stderr)
		t.L().Printf("Test Results:\n%s", rawResults)

		results := newORMTestsResults()

		scanner := bufio.NewScanner(bytes.NewReader(rawResults))
		for scanner.Scan() {
			__antithesis_instrumentation__.Notify(45290)
			match := activerecordResultRegex.FindStringSubmatch(scanner.Text())
			if match == nil {
				__antithesis_instrumentation__.Notify(45293)
				continue
			} else {
				__antithesis_instrumentation__.Notify(45294)
			}
			__antithesis_instrumentation__.Notify(45291)
			test, result := match[1], match[3]
			pass := result == "."
			skipped := result == "S"
			results.allTests = append(results.allTests, test)

			ignoredIssue, expectedIgnored := ignorelist[test]
			issue, expectedFailure := expectedFailures[test]
			switch {
			case expectedIgnored:
				__antithesis_instrumentation__.Notify(45295)
				results.results[test] = fmt.Sprintf("--- SKIP: %s due to %s (expected)", test, ignoredIssue)
				results.ignoredCount++
			case skipped && func() bool {
				__antithesis_instrumentation__.Notify(45303)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(45296)
				results.results[test] = fmt.Sprintf("--- SKIP: %s (unexpected)", test)
				results.unexpectedSkipCount++
			case skipped:
				__antithesis_instrumentation__.Notify(45297)
				results.results[test] = fmt.Sprintf("--- SKIP: %s (expected)", test)
				results.skipCount++
			case pass && func() bool {
				__antithesis_instrumentation__.Notify(45304)
				return !expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(45298)
				results.results[test] = fmt.Sprintf("--- PASS: %s (expected)", test)
				results.passExpectedCount++
			case pass && func() bool {
				__antithesis_instrumentation__.Notify(45305)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(45299)
				results.results[test] = fmt.Sprintf("--- PASS: %s - %s (unexpected)",
					test, maybeAddGithubLink(issue),
				)
				results.passUnexpectedCount++
			case !pass && func() bool {
				__antithesis_instrumentation__.Notify(45306)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(45300)
				results.results[test] = fmt.Sprintf("--- FAIL: %s - %s (expected)",
					test, maybeAddGithubLink(issue),
				)
				results.failExpectedCount++
				results.currentFailures = append(results.currentFailures, test)
			case !pass && func() bool {
				__antithesis_instrumentation__.Notify(45307)
				return !expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(45301)
				results.results[test] = fmt.Sprintf("--- FAIL: %s (unexpected)", test)
				results.failUnexpectedCount++
				results.currentFailures = append(results.currentFailures, test)
			default:
				__antithesis_instrumentation__.Notify(45302)
			}
			__antithesis_instrumentation__.Notify(45292)
			results.runTests[test] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(45251)

		results.summarizeAll(
			t, "activerecord", blocklistName, expectedFailures, version, supportedRailsVersion,
		)
	}
	__antithesis_instrumentation__.Notify(45231)

	r.Add(registry.TestSpec{
		Name:    "activerecord",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run:     runActiveRecord,
	})
}
