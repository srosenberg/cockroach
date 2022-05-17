package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/errors"
)

var gopgReleaseTagRegex = regexp.MustCompile(`^v(?P<major>\d+)(?:\.(?P<minor>\d+)(?:\.(?P<point>\d+))?)?$`)
var gopgSupportedTag = "v10.9.0"

func registerGopg(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48015)
	const (
		destPath        = `/mnt/data1/go-pg/pg`
		resultsDirPath  = `~/logs/report/gopg-results`
		resultsFilePath = resultsDirPath + `/raw_results`
	)

	runGopg := func(
		ctx context.Context,
		t test.Test,
		c cluster.Cluster,
	) {
		__antithesis_instrumentation__.Notify(48017)
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48030)
			t.Fatal("cannot be run in local mode")
		} else {
			__antithesis_instrumentation__.Notify(48031)
		}
		__antithesis_instrumentation__.Notify(48018)
		node := c.Node(1)
		t.Status("setting up cockroach")
		c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())
		version, err := fetchCockroachVersion(ctx, t.L(), c, node[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(48032)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48033)
		}
		__antithesis_instrumentation__.Notify(48019)

		if err := alterZoneConfigAndClusterSettings(ctx, t, version, c, node[0]); err != nil {
			__antithesis_instrumentation__.Notify(48034)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48035)
		}
		__antithesis_instrumentation__.Notify(48020)

		t.Status("cloning gopg and installing prerequisites")
		gopgLatestTag, err := repeatGetLatestTag(ctx, t, "go-pg", "pg", gopgReleaseTagRegex)
		if err != nil {
			__antithesis_instrumentation__.Notify(48036)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48037)
		}
		__antithesis_instrumentation__.Notify(48021)
		t.L().Printf("Latest gopg release is %s.", gopgLatestTag)
		t.L().Printf("Supported gopg release is %s.", gopgSupportedTag)

		installGolang(ctx, t, c, node)

		if err := repeatRunE(
			ctx, t, c, node, "remove old gopg",
			fmt.Sprintf(`sudo rm -rf %s`, destPath),
		); err != nil {
			__antithesis_instrumentation__.Notify(48038)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48039)
		}
		__antithesis_instrumentation__.Notify(48022)

		if err := repeatGitCloneE(
			ctx,
			t,
			c,
			"https://github.com/go-pg/pg.git",
			destPath,
			gopgSupportedTag,
			node,
		); err != nil {
			__antithesis_instrumentation__.Notify(48040)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48041)
		}
		__antithesis_instrumentation__.Notify(48023)

		blocklistName, expectedFailures, ignorelistName, ignorelist := gopgBlocklists.getLists(version)
		if expectedFailures == nil {
			__antithesis_instrumentation__.Notify(48042)
			t.Fatalf("No gopg blocklist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(48043)
		}
		__antithesis_instrumentation__.Notify(48024)
		if ignorelist == nil {
			__antithesis_instrumentation__.Notify(48044)
			t.Fatalf("No gopg ignorelist defined for cockroach version %s", version)
		} else {
			__antithesis_instrumentation__.Notify(48045)
		}
		__antithesis_instrumentation__.Notify(48025)
		t.L().Printf("Running cockroach version %s, using blocklist %s, using ignorelist %s",
			version, blocklistName, ignorelistName)

		if err := c.RunE(ctx, node, fmt.Sprintf("mkdir -p %s", resultsDirPath)); err != nil {
			__antithesis_instrumentation__.Notify(48046)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48047)
		}
		__antithesis_instrumentation__.Notify(48026)
		t.Status("running gopg test suite")

		const removeColorCodes = `sed -r "s/\x1B\[([0-9]{1,2}(;[0-9]{1,2})?)?[mGK]//g"`

		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), node,
			fmt.Sprintf(
				`cd %s && PGPORT=26257 PGUSER=root PGSSLMODE=disable PGDATABASE=postgres go test -v ./... 2>&1 | %s | tee %s`,
				destPath, removeColorCodes, resultsFilePath),
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(48048)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(48049)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48050)
			}
		} else {
			__antithesis_instrumentation__.Notify(48051)
		}
		__antithesis_instrumentation__.Notify(48027)

		rawResults := []byte(result.Stdout + result.Stderr)
		t.Status("collating the test results")
		t.L().Printf("Test Results: %s", rawResults)
		results := newORMTestsResults()

		if err := gopgParseTestGinkgoOutput(
			results, rawResults, expectedFailures, ignorelist,
		); err != nil {
			__antithesis_instrumentation__.Notify(48052)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48053)
		}
		__antithesis_instrumentation__.Notify(48028)

		result, err = c.RunWithDetailsSingleNode(ctx, t.L(), node,

			fmt.Sprintf(`cd %s &&
							GOPATH=%s go get -u github.com/jstemmer/go-junit-report &&
							cat %s | %s/bin/go-junit-report`,
				destPath, goPath, resultsFilePath, goPath),
		)

		if err != nil {
			__antithesis_instrumentation__.Notify(48054)

			commandError := (*install.NonZeroExitCode)(nil)
			if !errors.As(err, &commandError) {
				__antithesis_instrumentation__.Notify(48055)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48056)
			}
		} else {
			__antithesis_instrumentation__.Notify(48057)
		}
		__antithesis_instrumentation__.Notify(48029)

		xmlResults := []byte(result.Stdout + result.Stderr)

		results.parseJUnitXML(t, expectedFailures, ignorelist, xmlResults)
		results.summarizeFailed(
			t, "gopg", blocklistName, expectedFailures, version, gopgSupportedTag,
			0,
		)
	}
	__antithesis_instrumentation__.Notify(48016)

	r.Add(registry.TestSpec{
		Name:    "gopg",
		Owner:   registry.OwnerSQLExperience,
		Cluster: r.MakeClusterSpec(1),
		Tags:    []string{`default`, `orm`},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48058)
			runGopg(ctx, t, c)
		},
	})
}

func gopgParseTestGinkgoOutput(
	r *ormTestsResults, rawResults []byte, expectedFailures, ignorelist blocklist,
) (err error) {
	__antithesis_instrumentation__.Notify(48059)
	var (
		totalRunCount, totalTestCount int
		testSuiteStart                = regexp.MustCompile(`Running Suite: (?P<testsuite>[\w]+)`)
		summaryStartRegex             = regexp.MustCompile(`Summarizing [\d]+ Failures?:`)
		summaryEndRegex               = regexp.MustCompile(`Ran (?P<runCount>[\d]+) of (?P<testCount>[\d]+) Specs? in [\d]+\.[\d]+ seconds`)

		failureRegex                      = regexp.MustCompile(`\[(Fail|Panic!)\] (?P<class>.*) \[[\w]+\] (?P<name>.*)`)
		testGinkgoInternalTestNamePattern = `%s | %s | %s`
		testGinkgoInternalTestNameRE      = regexp.MustCompile(`.* | .* | .*`)
	)
	scanner := bufio.NewScanner(bytes.NewReader(rawResults))
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(48062)
		line := scanner.Bytes()
		if testSuiteStart.Match(line) {
			__antithesis_instrumentation__.Notify(48063)
			match := testSuiteStart.FindSubmatch(line)
			if match == nil {
				__antithesis_instrumentation__.Notify(48066)
				return errors.New("unexpectedly didn't find the name of the internal test suite")
			} else {
				__antithesis_instrumentation__.Notify(48067)
			}
			__antithesis_instrumentation__.Notify(48064)
			testSuiteName := string(match[1])

			for scanner.Scan() {
				__antithesis_instrumentation__.Notify(48068)
				line = scanner.Bytes()
				if testSuiteStart.Match(line) {
					__antithesis_instrumentation__.Notify(48069)

					match = testSuiteStart.FindSubmatch(line)
					if match == nil {
						__antithesis_instrumentation__.Notify(48071)
						return errors.New("unexpectedly didn't find the name of the internal test suite")
					} else {
						__antithesis_instrumentation__.Notify(48072)
					}
					__antithesis_instrumentation__.Notify(48070)
					testSuiteName = string(match[1])
				} else {
					__antithesis_instrumentation__.Notify(48073)
					if summaryStartRegex.Match(line) {
						__antithesis_instrumentation__.Notify(48074)
						break
					} else {
						__antithesis_instrumentation__.Notify(48075)
					}
				}
			}
			__antithesis_instrumentation__.Notify(48065)

			for scanner.Scan() {
				__antithesis_instrumentation__.Notify(48076)
				line = scanner.Bytes()

				match = summaryEndRegex.FindSubmatch(line)
				if match != nil {
					__antithesis_instrumentation__.Notify(48078)
					var runCount, totalCount int
					runCount, err = strconv.Atoi(string(match[1]))
					if err != nil {
						__antithesis_instrumentation__.Notify(48081)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48082)
					}
					__antithesis_instrumentation__.Notify(48079)
					totalCount, err = strconv.Atoi(string(match[2]))
					if err != nil {
						__antithesis_instrumentation__.Notify(48083)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48084)
					}
					__antithesis_instrumentation__.Notify(48080)
					totalRunCount += runCount
					totalTestCount += totalCount
					break
				} else {
					__antithesis_instrumentation__.Notify(48085)
				}
				__antithesis_instrumentation__.Notify(48077)

				match = failureRegex.FindSubmatch(line)
				if match != nil {
					__antithesis_instrumentation__.Notify(48086)
					class := string(match[2])
					name := strings.TrimSpace(string(match[3]))
					failedTest := fmt.Sprintf(testGinkgoInternalTestNamePattern, testSuiteName, class, name)
					if _, ignore := ignorelist[failedTest]; ignore {
						__antithesis_instrumentation__.Notify(48088)

						r.ignoredCount++
						continue
					} else {
						__antithesis_instrumentation__.Notify(48089)
					}
					__antithesis_instrumentation__.Notify(48087)
					r.currentFailures = append(r.currentFailures, failedTest)
					if _, ok := expectedFailures[failedTest]; ok {
						__antithesis_instrumentation__.Notify(48090)
						r.failExpectedCount++
					} else {
						__antithesis_instrumentation__.Notify(48091)
						r.failUnexpectedCount++
					}
				} else {
					__antithesis_instrumentation__.Notify(48092)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(48093)
		}
	}
	__antithesis_instrumentation__.Notify(48060)

	testGinkgoExpectedFailures := 0
	for failure := range expectedFailures {
		__antithesis_instrumentation__.Notify(48094)
		if testGinkgoInternalTestNameRE.MatchString(failure) {
			__antithesis_instrumentation__.Notify(48095)
			testGinkgoExpectedFailures++
		} else {
			__antithesis_instrumentation__.Notify(48096)
		}
	}
	__antithesis_instrumentation__.Notify(48061)

	passCount := totalRunCount - r.failExpectedCount - r.failUnexpectedCount
	r.passUnexpectedCount = testGinkgoExpectedFailures - r.failExpectedCount
	r.passExpectedCount = passCount - r.passUnexpectedCount
	return nil
}
