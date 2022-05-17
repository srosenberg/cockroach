package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

func alterZoneConfigAndClusterSettings(
	ctx context.Context, t test.Test, version string, c cluster.Cluster, nodeIdx int,
) error {
	__antithesis_instrumentation__.Notify(49601)
	db, err := c.ConnE(ctx, t.L(), nodeIdx)
	if err != nil {
		__antithesis_instrumentation__.Notify(49604)
		return err
	} else {
		__antithesis_instrumentation__.Notify(49605)
	}
	__antithesis_instrumentation__.Notify(49602)
	defer db.Close()

	for _, cmd := range []string{
		`ALTER RANGE default CONFIGURE ZONE USING num_replicas = 1, gc.ttlseconds = 30;`,
		`ALTER TABLE system.public.jobs CONFIGURE ZONE USING num_replicas = 1, gc.ttlseconds = 30;`,
		`ALTER RANGE meta CONFIGURE ZONE USING num_replicas = 1, gc.ttlseconds = 30;`,
		`ALTER RANGE system CONFIGURE ZONE USING num_replicas = 1, gc.ttlseconds = 30;`,
		`ALTER RANGE liveness CONFIGURE ZONE USING num_replicas = 1, gc.ttlseconds = 30;`,

		`SET CLUSTER SETTING kv.range_merge.queue_interval = '50ms'`,
		`SET CLUSTER SETTING kv.raft_log.disable_synchronization_unsafe = 'true'`,
		`SET CLUSTER SETTING jobs.registry.interval.cancel = '180s';`,
		`SET CLUSTER SETTING jobs.registry.interval.gc = '30s';`,
		`SET CLUSTER SETTING jobs.retention_time = '15s';`,
		`SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false;`,
		`SET CLUSTER SETTING kv.range_split.by_load_merge_delay = '5s';`,

		`SET CLUSTER SETTING server.user_login.password_encryption = 'scram-sha-256';`,

		`SET CLUSTER SETTING sql.defaults.experimental_temporary_tables.enabled = 'true';`,
		`SET CLUSTER SETTING sql.defaults.datestyle.enabled = true`,
		`SET CLUSTER SETTING sql.defaults.intervalstyle.enabled = true;`,
	} {
		__antithesis_instrumentation__.Notify(49606)
		if _, err := db.ExecContext(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(49607)
			return err
		} else {
			__antithesis_instrumentation__.Notify(49608)
		}
	}
	__antithesis_instrumentation__.Notify(49603)

	return nil
}

type ormTestsResults struct {
	currentFailures, allTests                    []string
	failUnexpectedCount, failExpectedCount       int
	ignoredCount, skipCount, unexpectedSkipCount int
	passUnexpectedCount, passExpectedCount       int

	results map[string]string

	allIssueHints map[string]string
	runTests      map[string]struct{}
}

func newORMTestsResults() *ormTestsResults {
	__antithesis_instrumentation__.Notify(49609)
	return &ormTestsResults{
		results:       make(map[string]string),
		allIssueHints: make(map[string]string),
		runTests:      make(map[string]struct{}),
	}
}

func (r *ormTestsResults) summarizeAll(
	t test.Test, ormName, blocklistName string, expectedFailures blocklist, version, tag string,
) {
	__antithesis_instrumentation__.Notify(49610)

	notRunCount := 0
	for test, issue := range expectedFailures {
		__antithesis_instrumentation__.Notify(49613)
		if _, ok := r.runTests[test]; ok {
			__antithesis_instrumentation__.Notify(49615)
			continue
		} else {
			__antithesis_instrumentation__.Notify(49616)
		}
		__antithesis_instrumentation__.Notify(49614)
		r.allTests = append(r.allTests, test)
		r.results[test] = fmt.Sprintf("--- FAIL: %s - %s (not run)", test, maybeAddGithubLink(issue))
		notRunCount++
	}
	__antithesis_instrumentation__.Notify(49611)

	sort.Strings(r.allTests)
	for _, test := range r.allTests {
		__antithesis_instrumentation__.Notify(49617)
		result, ok := r.results[test]
		if !ok {
			__antithesis_instrumentation__.Notify(49619)
			t.Fatalf("can't find %s in test result list", test)
		} else {
			__antithesis_instrumentation__.Notify(49620)
		}
		__antithesis_instrumentation__.Notify(49618)
		t.L().Printf("%s\n", result)
	}
	__antithesis_instrumentation__.Notify(49612)

	t.L().Printf("------------------------\n")

	r.summarizeFailed(
		t, ormName, blocklistName, expectedFailures, version, tag, notRunCount,
	)
}

func (r *ormTestsResults) summarizeFailed(
	t test.Test,
	ormName, blocklistName string,
	expectedFailures blocklist,
	version, latestTag string,
	notRunCount int,
) {
	__antithesis_instrumentation__.Notify(49621)
	var bResults strings.Builder
	fmt.Fprintf(&bResults, "Tests run on Cockroach %s\n", version)
	fmt.Fprintf(&bResults, "Tests run against %s %s\n", ormName, latestTag)
	totalTestsRun := r.passExpectedCount + r.passUnexpectedCount + r.failExpectedCount + r.failUnexpectedCount
	fmt.Fprintf(&bResults, "%d Total Tests Run\n",
		totalTestsRun,
	)
	if totalTestsRun == 0 {
		__antithesis_instrumentation__.Notify(49625)
		t.Fatal("No tests ran! Fix the testing commands.")
	} else {
		__antithesis_instrumentation__.Notify(49626)
	}
	__antithesis_instrumentation__.Notify(49622)

	p := func(msg string, count int) {
		__antithesis_instrumentation__.Notify(49627)
		testString := "tests"
		if count == 1 {
			__antithesis_instrumentation__.Notify(49629)
			testString = "test"
		} else {
			__antithesis_instrumentation__.Notify(49630)
		}
		__antithesis_instrumentation__.Notify(49628)
		fmt.Fprintf(&bResults, "%d %s %s\n", count, testString, msg)
	}
	__antithesis_instrumentation__.Notify(49623)
	p("passed", r.passUnexpectedCount+r.passExpectedCount)
	p("failed", r.failUnexpectedCount+r.failExpectedCount)
	p("skipped", r.skipCount)
	p("ignored", r.ignoredCount)
	p("passed unexpectedly", r.passUnexpectedCount)
	p("failed unexpectedly", r.failUnexpectedCount)
	p("expected failed but skipped", r.unexpectedSkipCount)
	p("expected failed but not run", notRunCount)

	fmt.Fprintf(&bResults, "---\n")
	for _, result := range r.results {
		__antithesis_instrumentation__.Notify(49631)
		if strings.Contains(result, "unexpected") {
			__antithesis_instrumentation__.Notify(49632)
			fmt.Fprintf(&bResults, "%s\n", result)
		} else {
			__antithesis_instrumentation__.Notify(49633)
		}
	}
	__antithesis_instrumentation__.Notify(49624)

	fmt.Fprintf(&bResults, "For a full summary look at the %s artifacts \n", ormName)
	t.L().Printf("%s\n", bResults.String())
	t.L().Printf("------------------------\n")

	if r.failUnexpectedCount > 0 || func() bool {
		__antithesis_instrumentation__.Notify(49634)
		return r.passUnexpectedCount > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(49635)
		return notRunCount > 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(49636)
		return r.unexpectedSkipCount > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(49637)

		sort.Strings(r.currentFailures)
		var b strings.Builder
		fmt.Fprintf(&b, "Here is new %s blocklist that can be used to update the test:\n\n", ormName)
		fmt.Fprintf(&b, "var %s = blocklist{\n", blocklistName)
		for _, test := range r.currentFailures {
			__antithesis_instrumentation__.Notify(49639)
			issue := expectedFailures[test]
			if len(issue) == 0 || func() bool {
				__antithesis_instrumentation__.Notify(49642)
				return issue == "unknown" == true
			}() == true {
				__antithesis_instrumentation__.Notify(49643)
				issue = r.allIssueHints[test]
			} else {
				__antithesis_instrumentation__.Notify(49644)
			}
			__antithesis_instrumentation__.Notify(49640)
			if len(issue) == 0 {
				__antithesis_instrumentation__.Notify(49645)
				issue = "unknown"
			} else {
				__antithesis_instrumentation__.Notify(49646)
			}
			__antithesis_instrumentation__.Notify(49641)
			fmt.Fprintf(&b, "  \"%s\": \"%s\",\n", test, issue)
		}
		__antithesis_instrumentation__.Notify(49638)
		fmt.Fprintf(&b, "}\n\n")
		t.L().Printf("\n\n%s\n\n", b.String())
		t.L().Printf("------------------------\n")
		t.Fatalf("\n%s\nAn updated blocklist (%s) is available in the artifacts' %s log\n",
			bResults.String(),
			blocklistName,
			ormName,
		)
	} else {
		__antithesis_instrumentation__.Notify(49647)
	}
}
