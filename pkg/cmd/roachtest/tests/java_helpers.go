package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

var issueRegexp = regexp.MustCompile(`See: https://[^\s]+issues?/(\d+)`)

type status int

const (
	statusPass status = iota
	statusFail
	statusSkip
)

func extractFailureFromJUnitXML(contents []byte) ([]string, []status, map[string]string, error) {
	__antithesis_instrumentation__.Notify(48650)
	type Failure struct {
		Message string `xml:"message,attr"`
	}
	type Error struct {
		Message string `xml:"message,attr"`
	}
	type TestCase struct {
		Name      string    `xml:"name,attr"`
		ClassName string    `xml:"classname,attr"`
		Failure   Failure   `xml:"failure,omitempty"`
		Error     Error     `xml:"error,omitempty"`
		Skipped   *struct{} `xml:"skipped,omitempty"`
	}
	type TestSuite struct {
		XMLName   xml.Name   `xml:"testsuite"`
		TestCases []TestCase `xml:"testcase"`
	}
	type TestSuites struct {
		XMLName    xml.Name    `xml:"testsuites"`
		TestSuites []TestSuite `xml:"testsuite"`
	}

	var testSuite TestSuite
	_ = testSuite.XMLName
	var testSuites TestSuites
	_ = testSuites.XMLName

	var tests []string
	var testStatuses []status
	var failedTestToIssue = make(map[string]string)
	processTestSuite := func(testSuite TestSuite) {
		__antithesis_instrumentation__.Notify(48653)
		for _, testCase := range testSuite.TestCases {
			__antithesis_instrumentation__.Notify(48654)
			testName := fmt.Sprintf("%s.%s", testCase.ClassName, testCase.Name)
			testPassed := len(testCase.Failure.Message) == 0 && func() bool {
				__antithesis_instrumentation__.Notify(48655)
				return len(testCase.Error.Message) == 0 == true
			}() == true
			tests = append(tests, testName)
			if testCase.Skipped != nil {
				__antithesis_instrumentation__.Notify(48656)
				testStatuses = append(testStatuses, statusSkip)
			} else {
				__antithesis_instrumentation__.Notify(48657)
				if testPassed {
					__antithesis_instrumentation__.Notify(48658)
					testStatuses = append(testStatuses, statusPass)
				} else {
					__antithesis_instrumentation__.Notify(48659)
					testStatuses = append(testStatuses, statusFail)
					message := testCase.Failure.Message
					if len(message) == 0 {
						__antithesis_instrumentation__.Notify(48662)
						message = testCase.Error.Message
					} else {
						__antithesis_instrumentation__.Notify(48663)
					}
					__antithesis_instrumentation__.Notify(48660)

					issue := "unknown"
					match := issueRegexp.FindStringSubmatch(message)
					if match != nil {
						__antithesis_instrumentation__.Notify(48664)
						issue = match[1]
					} else {
						__antithesis_instrumentation__.Notify(48665)
					}
					__antithesis_instrumentation__.Notify(48661)
					failedTestToIssue[testName] = issue
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(48651)

	if err := xml.Unmarshal(contents, &testSuites); err == nil {
		__antithesis_instrumentation__.Notify(48666)

		for _, testSuite := range testSuites.TestSuites {
			__antithesis_instrumentation__.Notify(48667)
			processTestSuite(testSuite)
		}
	} else {
		__antithesis_instrumentation__.Notify(48668)

		if err := xml.Unmarshal(contents, &testSuite); err != nil {
			__antithesis_instrumentation__.Notify(48670)
			return nil, nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(48671)
		}
		__antithesis_instrumentation__.Notify(48669)
		processTestSuite(testSuite)
	}
	__antithesis_instrumentation__.Notify(48652)

	return tests, testStatuses, failedTestToIssue, nil
}

func (r *ormTestsResults) parseJUnitXML(
	t test.Test, expectedFailures, ignorelist blocklist, testOutputInJUnitXMLFormat []byte,
) {
	__antithesis_instrumentation__.Notify(48672)
	tests, statuses, issueHints, err := extractFailureFromJUnitXML(testOutputInJUnitXMLFormat)
	if err != nil {
		__antithesis_instrumentation__.Notify(48675)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(48676)
	}
	__antithesis_instrumentation__.Notify(48673)
	for testName, issue := range issueHints {
		__antithesis_instrumentation__.Notify(48677)
		r.allIssueHints[testName] = issue
	}
	__antithesis_instrumentation__.Notify(48674)
	for i, test := range tests {
		__antithesis_instrumentation__.Notify(48678)

		if _, alreadyTested := r.results[test]; alreadyTested {
			__antithesis_instrumentation__.Notify(48682)
			continue
		} else {
			__antithesis_instrumentation__.Notify(48683)
		}
		__antithesis_instrumentation__.Notify(48679)
		r.allTests = append(r.allTests, test)
		ignoredIssue, expectedIgnored := ignorelist[test]
		issue, expectedFailure := expectedFailures[test]
		if len(issue) == 0 || func() bool {
			__antithesis_instrumentation__.Notify(48684)
			return issue == "unknown" == true
		}() == true {
			__antithesis_instrumentation__.Notify(48685)
			issue = issueHints[test]
		} else {
			__antithesis_instrumentation__.Notify(48686)
		}
		__antithesis_instrumentation__.Notify(48680)
		status := statuses[i]
		switch {
		case expectedIgnored:
			__antithesis_instrumentation__.Notify(48687)
			r.results[test] = fmt.Sprintf("--- IGNORE: %s due to %s (expected)", test, ignoredIssue)
			r.ignoredCount++
		case status == statusSkip:
			__antithesis_instrumentation__.Notify(48688)
			r.results[test] = fmt.Sprintf("--- SKIP: %s", test)
			r.skipCount++
		case status == statusPass && func() bool {
			__antithesis_instrumentation__.Notify(48694)
			return !expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(48689)
			r.results[test] = fmt.Sprintf("--- PASS: %s (expected)", test)
			r.passExpectedCount++
		case status == statusPass && func() bool {
			__antithesis_instrumentation__.Notify(48695)
			return expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(48690)
			r.results[test] = fmt.Sprintf("--- PASS: %s - %s (unexpected)",
				test, maybeAddGithubLink(issue),
			)
			r.passUnexpectedCount++
		case status == statusFail && func() bool {
			__antithesis_instrumentation__.Notify(48696)
			return expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(48691)
			r.results[test] = fmt.Sprintf("--- FAIL: %s - %s (expected)",
				test, maybeAddGithubLink(issue),
			)
			r.failExpectedCount++
			r.currentFailures = append(r.currentFailures, test)
		case status == statusFail && func() bool {
			__antithesis_instrumentation__.Notify(48697)
			return !expectedFailure == true
		}() == true:
			__antithesis_instrumentation__.Notify(48692)
			r.results[test] = fmt.Sprintf("--- FAIL: %s - %s (unexpected)",
				test, maybeAddGithubLink(issue))
			r.failUnexpectedCount++
			r.currentFailures = append(r.currentFailures, test)
		default:
			__antithesis_instrumentation__.Notify(48693)
		}
		__antithesis_instrumentation__.Notify(48681)
		r.runTests[test] = struct{}{}
	}
}

func parseAndSummarizeJavaORMTestsResults(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	node option.NodeListOption,
	ormName string,
	testOutput []byte,
	blocklistName string,
	expectedFailures blocklist,
	ignorelist blocklist,
	version string,
	tag string,
) {
	__antithesis_instrumentation__.Notify(48698)
	results := newORMTestsResults()
	filesRaw := strings.Split(string(testOutput), "\n")

	var files []string
	for _, f := range filesRaw {
		__antithesis_instrumentation__.Notify(48701)
		file := strings.TrimSpace(f)
		if len(file) > 0 {
			__antithesis_instrumentation__.Notify(48702)
			files = append(files, file)
		} else {
			__antithesis_instrumentation__.Notify(48703)
		}
	}
	__antithesis_instrumentation__.Notify(48699)
	for i, file := range files {
		__antithesis_instrumentation__.Notify(48704)
		t.L().Printf("Parsing %d of %d: %s\n", i+1, len(files), file)
		result, err := repeatRunWithDetailsSingleNode(
			ctx,
			c,
			t,
			node,
			fmt.Sprintf("fetching results file %s", file),
			fmt.Sprintf("cat %s", file),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(48706)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48707)
		}
		__antithesis_instrumentation__.Notify(48705)

		results.parseJUnitXML(t, expectedFailures, ignorelist, []byte(result.Stdout))
	}
	__antithesis_instrumentation__.Notify(48700)

	results.summarizeAll(
		t, ormName, blocklistName, expectedFailures, version, tag,
	)
}
