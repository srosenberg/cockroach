// Copyright 2020 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
//
// Command skip-test will skip a test in the CockroachDB repo.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var leakTestRegexp = regexp.MustCompile(`defer leaktest.AfterTest\(t\)\(\)`)

var (
	flags          = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flagIssueNum   = flags.Int("issue_num", 0, "issue to link the skip to; if unset skip-test will\ntry to search for existing issues based on the test name")
	flagReason     = flags.String("reason", "flaky test", "reason to put under skip")
	flagUnderRace  = flags.Bool("under_race", false, "if true, only skip under race")
	flagUnderBazel = flags.Bool("under_bazel", false, "if true, only skip under bazel")
)

const description = `The skip-test utility creates a pull request to skip a test.

Example usage:

    ./bin/skip-test -issue_num 1234 pkg/to/test:TestToSkip

The following options are available:

`

func usage() {
	__antithesis_instrumentation__.Notify(52680)
	fmt.Fprint(flags.Output(), description)
	flags.PrintDefaults()
	fmt.Println("")
}

func main() {
	__antithesis_instrumentation__.Notify(52681)
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		__antithesis_instrumentation__.Notify(52701)
		usage()
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52702)
	}
	__antithesis_instrumentation__.Notify(52682)

	if len(flags.Args()) != 1 {
		__antithesis_instrumentation__.Notify(52703)
		usage()
		log.Fatalf("missing required argument: `TestName` or `pkg/to/test:TestToSkip`")
	} else {
		__antithesis_instrumentation__.Notify(52704)
	}
	__antithesis_instrumentation__.Notify(52683)

	ctx := context.Background()

	remote, _ := capture("git", "config", "--get", "cockroach.remote")
	if remote == "" {
		__antithesis_instrumentation__.Notify(52705)
		log.Fatalf("set cockroach.remote to the name of the Git remote to push")
	} else {
		__antithesis_instrumentation__.Notify(52706)
	}
	__antithesis_instrumentation__.Notify(52684)

	var ghAuthClient *http.Client
	ghToken, _ := capture("git", "config", "--get", "cockroach.githubToken")
	if ghToken != "" {
		__antithesis_instrumentation__.Notify(52707)
		ghAuthClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: ghToken}))
	} else {
		__antithesis_instrumentation__.Notify(52708)
	}
	__antithesis_instrumentation__.Notify(52685)
	ghClient := github.NewClient(ghAuthClient)

	arg := flags.Args()[0]
	var pkgName, testName string
	splitArg := strings.Split(arg, ":")
	switch len(splitArg) {
	case 1:
		__antithesis_instrumentation__.Notify(52709)
		testName = splitArg[0]
	case 2:
		__antithesis_instrumentation__.Notify(52710)
		pkgName = splitArg[0]
		testName = splitArg[1]
	default:
		__antithesis_instrumentation__.Notify(52711)
		log.Fatalf("expected test to be of format `TestName` or `pkg/to/test:TestToSkip`, found %s", arg)
	}
	__antithesis_instrumentation__.Notify(52686)

	if *flagUnderBazel && func() bool {
		__antithesis_instrumentation__.Notify(52712)
		return *flagUnderRace == true
	}() == true {
		__antithesis_instrumentation__.Notify(52713)
		log.Fatal("cannot use both -under_race and -under_bazel")
	} else {
		__antithesis_instrumentation__.Notify(52714)
	}
	__antithesis_instrumentation__.Notify(52687)

	if err := spawn("git", "diff", "--exit-code"); err != nil {
		__antithesis_instrumentation__.Notify(52715)
		log.Fatal(errors.Wrap(err, "git state may not be clean, please use `git stash` or commit changes before proceeding."))
	} else {
		__antithesis_instrumentation__.Notify(52716)
	}
	__antithesis_instrumentation__.Notify(52688)

	if err := spawn("git", "fetch", "https://github.com/cockroachdb/cockroach.git", "master"); err != nil {
		__antithesis_instrumentation__.Notify(52717)
		log.Fatal(errors.Wrap(err, "failed to get CockroachDB master"))
	} else {
		__antithesis_instrumentation__.Notify(52718)
	}
	__antithesis_instrumentation__.Notify(52689)

	skipBranch := fmt.Sprintf("skip-test-%s", testName)

	if err := spawn("git", "checkout", "-b", skipBranch, "FETCH_HEAD"); err != nil {
		__antithesis_instrumentation__.Notify(52719)
		log.Fatal(errors.Wrapf(err, "failed to checkout branch %s", skipBranch))
	} else {
		__antithesis_instrumentation__.Notify(52720)
	}
	__antithesis_instrumentation__.Notify(52690)

	searchPath := pkgName
	if searchPath == "" {
		__antithesis_instrumentation__.Notify(52721)
		searchPath = "./pkg"
	} else {
		__antithesis_instrumentation__.Notify(52722)
	}
	__antithesis_instrumentation__.Notify(52691)
	fnGrep := fmt.Sprintf(`func %s(t \*testing\.T)`, regexp.QuoteMeta(testName))
	grepOutput, err := capture(
		"git",
		"grep",
		"-i",
		fnGrep,
		searchPath,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(52723)
		log.Fatal(errors.Wrapf(err, "failed to grep for the failing test"))
	} else {
		__antithesis_instrumentation__.Notify(52724)
	}
	__antithesis_instrumentation__.Notify(52692)

	grepOutputSplit := strings.Split(grepOutput, "\n")
	if len(grepOutputSplit) != 1 {
		__antithesis_instrumentation__.Notify(52725)
		log.Fatalf("expected 1 result for test %s, found %d:\n%s", arg, len(grepOutputSplit), grepOutput)
	} else {
		__antithesis_instrumentation__.Notify(52726)
	}
	__antithesis_instrumentation__.Notify(52693)
	fileName := strings.Split(grepOutput, ":")[0]

	pkgPrefix := strings.TrimPrefix(filepath.Dir(fileName), "pkg/")
	issueNum := findIssue(ctx, ghClient, pkgPrefix, testName)

	replaceFile(fileName, testName, issueNum)

	devPath, err := exec.LookPath("./dev")
	if err != nil {
		__antithesis_instrumentation__.Notify(52727)
		fmt.Printf("./dev not found, trying dev\n")
		devPath, err = exec.LookPath("dev")
		if err != nil {
			__antithesis_instrumentation__.Notify(52728)
			log.Fatal(errors.Wrapf(err, "no path found for dev"))
		} else {
			__antithesis_instrumentation__.Notify(52729)
		}
	} else {
		__antithesis_instrumentation__.Notify(52730)
	}
	__antithesis_instrumentation__.Notify(52694)
	if err := spawn(devPath, "generate", "bazel"); err != nil {
		__antithesis_instrumentation__.Notify(52731)
		log.Fatal(errors.Wrap(err, "failed to run bazel"))
	} else {
		__antithesis_instrumentation__.Notify(52732)
	}
	__antithesis_instrumentation__.Notify(52695)

	if err := spawn("git", "add", searchPath); err != nil {
		__antithesis_instrumentation__.Notify(52733)
		log.Fatal(errors.Wrapf(err, "failed to add %s to commit", searchPath))
	} else {
		__antithesis_instrumentation__.Notify(52734)
	}
	__antithesis_instrumentation__.Notify(52696)
	var modifierStr string
	if *flagUnderRace {
		__antithesis_instrumentation__.Notify(52735)
		modifierStr = " under race"
	} else {
		__antithesis_instrumentation__.Notify(52736)
		if *flagUnderBazel {
			__antithesis_instrumentation__.Notify(52737)
			modifierStr = " under bazel"
		} else {
			__antithesis_instrumentation__.Notify(52738)
		}
	}
	__antithesis_instrumentation__.Notify(52697)
	commitMsg := fmt.Sprintf(`%s: skip %s%s

Refs: #%d

Reason: %s

Generated by bin/skip-test.

Release justification: non-production code changes

Release note: None
`, pkgPrefix, testName, modifierStr, issueNum, *flagReason)
	if err := spawn("git", "commit", "-m", commitMsg); err != nil {
		__antithesis_instrumentation__.Notify(52739)
		log.Fatal(errors.Wrapf(err, "failed to commit %s", fileName))
	} else {
		__antithesis_instrumentation__.Notify(52740)
	}
	__antithesis_instrumentation__.Notify(52698)

	if err := spawn("git", "push", "--force", remote, fmt.Sprintf("%[1]s:%[1]s", skipBranch, skipBranch)); err != nil {
		__antithesis_instrumentation__.Notify(52741)
		log.Fatal(errors.Wrapf(err, "failed to push to remote"))
	} else {
		__antithesis_instrumentation__.Notify(52742)
	}
	__antithesis_instrumentation__.Notify(52699)

	if err := spawn(
		"python",
		"-c",
		"import sys, webbrowser; sys.exit(not webbrowser.open(sys.argv[1]))",
		fmt.Sprintf("https://github.com/cockroachdb/cockroach/compare/master...%s:%s?expand=1", remote, skipBranch),
	); err != nil {
		__antithesis_instrumentation__.Notify(52743)
		log.Fatal(errors.Wrapf(err, "failed to open web browser"))
	} else {
		__antithesis_instrumentation__.Notify(52744)
	}
	__antithesis_instrumentation__.Notify(52700)

	if err := checkoutPrevious(); err != nil {
		__antithesis_instrumentation__.Notify(52745)
		log.Fatal(errors.Wrapf(err, "failed to checkout previous branch"))
	} else {
		__antithesis_instrumentation__.Notify(52746)
	}
}

func replaceFile(fileName, testName string, issueNum int) {
	__antithesis_instrumentation__.Notify(52747)
	fileContents, err := ioutil.ReadFile(fileName)
	if err != nil {
		__antithesis_instrumentation__.Notify(52754)
		log.Fatal(errors.Wrapf(err, "failed to read file: %s", fileName))
	} else {
		__antithesis_instrumentation__.Notify(52755)
	}
	__antithesis_instrumentation__.Notify(52748)

	lines := strings.Split(string(fileContents), "\n")
	r := regexp.MustCompile(fmt.Sprintf(`func %s\(t \*testing\.T\)`, regexp.QuoteMeta(testName)))
	lineIdx := -1
	for i, line := range lines {
		__antithesis_instrumentation__.Notify(52756)
		if r.MatchString(line) {
			__antithesis_instrumentation__.Notify(52757)
			lineIdx = i
			break
		} else {
			__antithesis_instrumentation__.Notify(52758)
		}
	}
	__antithesis_instrumentation__.Notify(52749)
	if lineIdx == -1 {
		__antithesis_instrumentation__.Notify(52759)
		log.Fatalf("failed to find test output %s in %s", testName, fileName)
	} else {
		__antithesis_instrumentation__.Notify(52760)
	}
	__antithesis_instrumentation__.Notify(52750)

	insertLineIdx := lineIdx + 1
	for leakTestRegexp.MatchString(lines[insertLineIdx]) {
		__antithesis_instrumentation__.Notify(52761)
		insertLineIdx++
	}
	__antithesis_instrumentation__.Notify(52751)

	newLines := append(
		[]string{},
		lines[:insertLineIdx]...,
	)
	if *flagUnderRace {
		__antithesis_instrumentation__.Notify(52762)
		newLines = append(
			newLines,
			fmt.Sprintf(`skip.UnderRaceWithIssue(t, %d, "%s")`, issueNum, *flagReason),
		)
	} else {
		__antithesis_instrumentation__.Notify(52763)
		if *flagUnderBazel {
			__antithesis_instrumentation__.Notify(52764)
			newLines = append(
				newLines,
				fmt.Sprintf(`skip.UnderBazelWithIssue(t, %d, "%s")`, issueNum, *flagReason),
			)
		} else {
			__antithesis_instrumentation__.Notify(52765)
			newLines = append(
				newLines,
				fmt.Sprintf(`skip.WithIssue(t, %d, "%s")`, issueNum, *flagReason),
			)
		}
	}
	__antithesis_instrumentation__.Notify(52752)
	newLines = append(
		newLines,
		lines[insertLineIdx:]...,
	)

	if err := ioutil.WriteFile(fileName, []byte(strings.Join(newLines, "\n")), 0644); err != nil {
		__antithesis_instrumentation__.Notify(52766)
		log.Fatal(errors.Wrapf(err, "failed to write file %s", fileContents))
	} else {
		__antithesis_instrumentation__.Notify(52767)
	}
	__antithesis_instrumentation__.Notify(52753)

	if err := spawn("bin/crlfmt", "-w", "-tab", "2", fileName); err != nil {
		__antithesis_instrumentation__.Notify(52768)
		log.Fatal(errors.Wrapf(err, "failed to run crlfmt on %s", fileName))
	} else {
		__antithesis_instrumentation__.Notify(52769)
	}
}

func checkoutPrevious() error {
	__antithesis_instrumentation__.Notify(52770)
	branch, err := capture("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		__antithesis_instrumentation__.Notify(52774)
		return errors.Newf("looking up current branch name: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(52775)
	}
	__antithesis_instrumentation__.Notify(52771)
	if !regexp.MustCompile(`^skip-test-.*`).MatchString(branch) {
		__antithesis_instrumentation__.Notify(52776)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(52777)
	}
	__antithesis_instrumentation__.Notify(52772)
	if err := spawn("git", "checkout", "-"); err != nil {
		__antithesis_instrumentation__.Notify(52778)
		return errors.Newf("returning to previous branch: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(52779)
	}
	__antithesis_instrumentation__.Notify(52773)
	return nil
}

func findIssue(ctx context.Context, ghClient *github.Client, pkgPrefix, testName string) int {
	__antithesis_instrumentation__.Notify(52780)
	issueNum := *flagIssueNum
	if issueNum == 0 {
		__antithesis_instrumentation__.Notify(52782)
		searched, _, err := ghClient.Search.Issues(
			ctx,
			fmt.Sprintf(`"%s: %s failed" in:title is:open is:issue`, pkgPrefix, testName),
			nil,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(52785)
			log.Fatal(errors.Wrap(err, "failed searching for issue"))
		} else {
			__antithesis_instrumentation__.Notify(52786)
		}
		__antithesis_instrumentation__.Notify(52783)
		if len(searched.Issues) != 1 {
			__antithesis_instrumentation__.Notify(52787)
			var issues []string
			for _, issue := range searched.Issues {
				__antithesis_instrumentation__.Notify(52789)
				issues = append(issues, strconv.Itoa(issue.GetNumber()))
			}
			__antithesis_instrumentation__.Notify(52788)
			log.Fatal(errors.Newf("found 0 or multiple issues for %s: %s\nuse --issue_num=<num> to attach a created issue.", testName, strings.Join(issues, ",")))
		} else {
			__antithesis_instrumentation__.Notify(52790)
		}
		__antithesis_instrumentation__.Notify(52784)
		return searched.Issues[0].GetNumber()
	} else {
		__antithesis_instrumentation__.Notify(52791)
	}
	__antithesis_instrumentation__.Notify(52781)
	return issueNum
}
