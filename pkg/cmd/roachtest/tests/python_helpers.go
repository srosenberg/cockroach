package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
)

var pythonUnitTestOutputRegex = regexp.MustCompile(`(?P<name>.*) \((?P<class>.*)\) \.\.\. (?P<result>[^'"]*?)(?: u?['"](?P<reason>.*)['"])?$`)

func (r *ormTestsResults) parsePythonUnitTestOutput(
	input []byte, expectedFailures blocklist, ignoredList blocklist,
) {
	__antithesis_instrumentation__.Notify(49871)
	scanner := bufio.NewScanner(bytes.NewReader(input))
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(49872)
		match := pythonUnitTestOutputRegex.FindStringSubmatch(scanner.Text())
		if match != nil {
			__antithesis_instrumentation__.Notify(49873)
			groups := map[string]string{}
			for i, name := range match {
				__antithesis_instrumentation__.Notify(49877)
				groups[pythonUnitTestOutputRegex.SubexpNames()[i]] = name
			}
			__antithesis_instrumentation__.Notify(49874)
			test := fmt.Sprintf("%s.%s", groups["class"], groups["name"])
			skipped := groups["result"] == "skipped" || func() bool {
				__antithesis_instrumentation__.Notify(49878)
				return groups["result"] == "expected failure" == true
			}() == true
			skipReason := ""
			if skipped {
				__antithesis_instrumentation__.Notify(49879)
				skipReason = groups["reason"]
			} else {
				__antithesis_instrumentation__.Notify(49880)
			}
			__antithesis_instrumentation__.Notify(49875)
			pass := groups["result"] == "ok" || func() bool {
				__antithesis_instrumentation__.Notify(49881)
				return groups["result"] == "unexpected success" == true
			}() == true
			r.allTests = append(r.allTests, test)

			ignoredIssue, expectedIgnored := ignoredList[test]
			issue, expectedFailure := expectedFailures[test]
			switch {
			case expectedIgnored:
				__antithesis_instrumentation__.Notify(49882)
				r.results[test] = fmt.Sprintf("--- SKIP: %s due to %s (expected)", test, ignoredIssue)
				r.ignoredCount++
			case skipped && func() bool {
				__antithesis_instrumentation__.Notify(49890)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(49883)
				r.results[test] = fmt.Sprintf("--- SKIP: %s due to %s (unexpected)", test, skipReason)
				r.unexpectedSkipCount++
			case skipped:
				__antithesis_instrumentation__.Notify(49884)
				r.results[test] = fmt.Sprintf("--- SKIP: %s due to %s (expected)", test, skipReason)
				r.skipCount++
			case pass && func() bool {
				__antithesis_instrumentation__.Notify(49891)
				return !expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(49885)
				r.results[test] = fmt.Sprintf("--- PASS: %s (expected)", test)
				r.passExpectedCount++
			case pass && func() bool {
				__antithesis_instrumentation__.Notify(49892)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(49886)
				r.results[test] = fmt.Sprintf("--- PASS: %s - %s (unexpected)",
					test, maybeAddGithubLink(issue),
				)
				r.passUnexpectedCount++
			case !pass && func() bool {
				__antithesis_instrumentation__.Notify(49893)
				return expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(49887)
				r.results[test] = fmt.Sprintf("--- FAIL: %s - %s (expected)",
					test, maybeAddGithubLink(issue),
				)
				r.failExpectedCount++
				r.currentFailures = append(r.currentFailures, test)
			case !pass && func() bool {
				__antithesis_instrumentation__.Notify(49894)
				return !expectedFailure == true
			}() == true:
				__antithesis_instrumentation__.Notify(49888)
				r.results[test] = fmt.Sprintf("--- FAIL: %s (unexpected)", test)
				r.failUnexpectedCount++
				r.currentFailures = append(r.currentFailures, test)
			default:
				__antithesis_instrumentation__.Notify(49889)
			}
			__antithesis_instrumentation__.Notify(49876)
			r.runTests[test] = struct{}{}
		} else {
			__antithesis_instrumentation__.Notify(49895)
		}
	}
}
