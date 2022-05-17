// This utility detects new tests added in a given pull request, and runs them
// under stress in our CI infrastructure.
//
// Note that this program will directly invoke the build system, so there is no
// need to process its output. See build/teamcity-support.sh for usage examples.
//
// Note that our CI infrastructure has no notion of "pull requests", forcing
// the approach taken here be quite brute-force with respect to its use of the
// GitHub API.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	_ "github.com/cockroachdb/cockroach/pkg/testutils/buildutil"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	githubAPITokenEnv    = "GITHUB_API_TOKEN"
	teamcityVCSNumberEnv = "BUILD_VCS_NUMBER"
	targetEnv            = "TARGET"

	packageEnv    = "GHM_PACKAGES"
	forceBazelEnv = "GHM_FORCE_BAZEL"
)

const goTestStr = `func (Test[^a-z]\w*)\(.*\*testing\.TB?\) {$`

const bazelStressTarget = "@com_github_cockroachdb_stress//:stress"

var currentGoTestRE = regexp.MustCompile(`.*` + goTestStr)
var newGoTestRE = regexp.MustCompile(`^\+\s*` + goTestStr)

type pkg struct {
	tests []string
}

func pkgsFromGithubPRForSHA(
	ctx context.Context, org string, repo string, sha string,
) (map[string]pkg, error) {
	__antithesis_instrumentation__.Notify(40856)
	client := ghClient(ctx)
	currentPull := findPullRequest(ctx, client, org, repo, sha)
	if currentPull == nil {
		__antithesis_instrumentation__.Notify(40859)
		log.Printf("SHA %s not found in open pull requests, skipping stress", sha)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(40860)
	}
	__antithesis_instrumentation__.Notify(40857)

	diff, err := getDiff(ctx, client, org, repo, *currentPull.Number)
	if err != nil {
		__antithesis_instrumentation__.Notify(40861)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(40862)
	}
	__antithesis_instrumentation__.Notify(40858)

	return pkgsFromDiff(strings.NewReader(diff))
}

func pkgsFromDiff(r io.Reader) (map[string]pkg, error) {
	__antithesis_instrumentation__.Notify(40863)
	const newFilePrefix = "+++ b/"
	const replacement = "$1"

	pkgs := make(map[string]pkg)

	var curPkgName string
	var curTestName string
	var inPrefix bool
	for reader := bufio.NewReader(r); ; {
		__antithesis_instrumentation__.Notify(40864)
		line, isPrefix, err := reader.ReadLine()
		switch {
		case err == nil:
			__antithesis_instrumentation__.Notify(40867)
		case err == io.EOF:
			__antithesis_instrumentation__.Notify(40868)
			return pkgs, nil
		default:
			__antithesis_instrumentation__.Notify(40869)
			return nil, err
		}
		__antithesis_instrumentation__.Notify(40865)

		if isPrefix {
			__antithesis_instrumentation__.Notify(40870)
			inPrefix = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(40871)
			if inPrefix {
				__antithesis_instrumentation__.Notify(40872)
				inPrefix = false
				continue
			} else {
				__antithesis_instrumentation__.Notify(40873)
			}
		}
		__antithesis_instrumentation__.Notify(40866)

		switch {
		case bytes.HasPrefix(line, []byte(newFilePrefix)):
			__antithesis_instrumentation__.Notify(40874)
			curPkgName = filepath.Dir(string(bytes.TrimPrefix(line, []byte(newFilePrefix))))
		case newGoTestRE.Match(line):
			__antithesis_instrumentation__.Notify(40875)
			curPkg := pkgs[curPkgName]
			curPkg.tests = append(curPkg.tests, string(newGoTestRE.ReplaceAll(line, []byte(replacement))))
			pkgs[curPkgName] = curPkg
		case currentGoTestRE.Match(line):
			__antithesis_instrumentation__.Notify(40876)
			curTestName = ""
			if !bytes.HasPrefix(line, []byte{'-'}) {
				__antithesis_instrumentation__.Notify(40879)
				curTestName = string(currentGoTestRE.ReplaceAll(line, []byte(replacement)))
			} else {
				__antithesis_instrumentation__.Notify(40880)
			}
		case bytes.HasPrefix(line, []byte{'-'}) && func() bool {
			__antithesis_instrumentation__.Notify(40881)
			return bytes.Contains(line, []byte(".Skip")) == true
		}() == true:
			__antithesis_instrumentation__.Notify(40877)
			if curPkgName != "" && func() bool {
				__antithesis_instrumentation__.Notify(40882)
				return len(curTestName) > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(40883)
				curPkg := pkgs[curPkgName]
				curPkg.tests = append(curPkg.tests, curTestName)
				pkgs[curPkgName] = curPkg
			} else {
				__antithesis_instrumentation__.Notify(40884)
			}
		default:
			__antithesis_instrumentation__.Notify(40878)
		}
	}
}

func findPullRequest(
	ctx context.Context, client *github.Client, org, repo, sha string,
) *github.PullRequest {
	__antithesis_instrumentation__.Notify(40885)
	opts := &github.PullRequestListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	for {
		__antithesis_instrumentation__.Notify(40886)
		pulls, resp, err := client.PullRequests.List(ctx, org, repo, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(40890)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40891)
		}
		__antithesis_instrumentation__.Notify(40887)

		for _, pull := range pulls {
			__antithesis_instrumentation__.Notify(40892)
			if *pull.Head.SHA == sha {
				__antithesis_instrumentation__.Notify(40893)
				return pull
			} else {
				__antithesis_instrumentation__.Notify(40894)
			}
		}
		__antithesis_instrumentation__.Notify(40888)

		if resp.NextPage == 0 {
			__antithesis_instrumentation__.Notify(40895)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(40896)
		}
		__antithesis_instrumentation__.Notify(40889)
		opts.Page = resp.NextPage
	}
}

func ghClient(ctx context.Context) *github.Client {
	__antithesis_instrumentation__.Notify(40897)
	var httpClient *http.Client
	if token, ok := os.LookupEnv(githubAPITokenEnv); ok {
		__antithesis_instrumentation__.Notify(40899)
		httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: token},
		))
	} else {
		__antithesis_instrumentation__.Notify(40900)
		log.Printf("GitHub API token environment variable %s is not set", githubAPITokenEnv)
	}
	__antithesis_instrumentation__.Notify(40898)
	return github.NewClient(httpClient)
}

func getDiff(
	ctx context.Context, client *github.Client, org, repo string, prNum int,
) (string, error) {
	__antithesis_instrumentation__.Notify(40901)
	diff, _, err := client.PullRequests.GetRaw(
		ctx,
		org,
		repo,
		prNum,
		github.RawOptions{Type: github.Diff},
	)
	return diff, err
}

func parsePackagesFromEnvironment(input string) (map[string]pkg, error) {
	__antithesis_instrumentation__.Notify(40902)
	const expectedFormat = "PACKAGE_NAME=TEST_NAME[,TEST_NAME...][;PACKAGE_NAME=...]"
	pkgTestStrs := strings.Split(input, ";")
	pkgs := make(map[string]pkg, len(pkgTestStrs))
	for _, pts := range pkgTestStrs {
		__antithesis_instrumentation__.Notify(40904)
		ptsParts := strings.Split(pts, "=")
		if len(ptsParts) < 2 {
			__antithesis_instrumentation__.Notify(40906)
			return nil, fmt.Errorf("invalid format for package environment variable: %q (expected format: %s)",
				input, expectedFormat)
		} else {
			__antithesis_instrumentation__.Notify(40907)
		}
		__antithesis_instrumentation__.Notify(40905)
		pkgName := ptsParts[0]
		tests := ptsParts[1]
		pkgs[pkgName] = pkg{
			tests: strings.Split(tests, ","),
		}
	}
	__antithesis_instrumentation__.Notify(40903)
	return pkgs, nil
}

func main() {
	__antithesis_instrumentation__.Notify(40908)
	sha, ok := os.LookupEnv(teamcityVCSNumberEnv)
	if !ok {
		__antithesis_instrumentation__.Notify(40916)
		log.Fatalf("VCS number environment variable %s is not set", teamcityVCSNumberEnv)
	} else {
		__antithesis_instrumentation__.Notify(40917)
	}
	__antithesis_instrumentation__.Notify(40909)

	target, ok := os.LookupEnv(targetEnv)
	if !ok {
		__antithesis_instrumentation__.Notify(40918)
		log.Fatalf("target variable %s is not set", targetEnv)
	} else {
		__antithesis_instrumentation__.Notify(40919)
	}
	__antithesis_instrumentation__.Notify(40910)
	if target != "stress" && func() bool {
		__antithesis_instrumentation__.Notify(40920)
		return target != "stressrace" == true
	}() == true {
		__antithesis_instrumentation__.Notify(40921)
		log.Fatalf("environment variable %s is %s; expected 'stress' or 'stressrace'", targetEnv, target)
	} else {
		__antithesis_instrumentation__.Notify(40922)
	}
	__antithesis_instrumentation__.Notify(40911)

	forceBazel := false
	if forceBazelStr, ok := os.LookupEnv(forceBazelEnv); ok {
		__antithesis_instrumentation__.Notify(40923)
		forceBazel, _ = strconv.ParseBool(forceBazelStr)
	} else {
		__antithesis_instrumentation__.Notify(40924)
	}
	__antithesis_instrumentation__.Notify(40912)
	var bazciPath string
	if bazel.BuiltWithBazel() || func() bool {
		__antithesis_instrumentation__.Notify(40925)
		return forceBazel == true
	}() == true {
		__antithesis_instrumentation__.Notify(40926)

		var err error
		bazciPath, err = exec.LookPath("bazci")
		if err != nil {
			__antithesis_instrumentation__.Notify(40927)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40928)
		}
	} else {
		__antithesis_instrumentation__.Notify(40929)
	}
	__antithesis_instrumentation__.Notify(40913)

	crdb, err := os.Getwd()
	if err != nil {
		__antithesis_instrumentation__.Notify(40930)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40931)
	}
	__antithesis_instrumentation__.Notify(40914)

	var pkgs map[string]pkg
	if pkgStr, ok := os.LookupEnv(packageEnv); ok {
		__antithesis_instrumentation__.Notify(40932)
		log.Printf("Using packages from environment variable %s", packageEnv)
		pkgs, err = parsePackagesFromEnvironment(pkgStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(40933)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40934)
		}

	} else {
		__antithesis_instrumentation__.Notify(40935)
		ctx := context.Background()
		const org = "cockroachdb"
		const repo = "cockroach"
		pkgs, err = pkgsFromGithubPRForSHA(ctx, org, repo, sha)
		if err != nil {
			__antithesis_instrumentation__.Notify(40936)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40937)
		}
	}
	__antithesis_instrumentation__.Notify(40915)

	if len(pkgs) > 0 {
		__antithesis_instrumentation__.Notify(40938)
		for name, pkg := range pkgs {
			__antithesis_instrumentation__.Notify(40939)

			duration := (20 * time.Minute) / time.Duration(len(pkgs))
			minDuration := (2 * time.Minute) * time.Duration(len(pkg.tests))
			if duration < minDuration {
				__antithesis_instrumentation__.Notify(40942)
				duration = minDuration
			} else {
				__antithesis_instrumentation__.Notify(40943)
			}
			__antithesis_instrumentation__.Notify(40940)

			timeout := (3 * duration) / 4

			parallelism := 16
			if target == "stressrace" {
				__antithesis_instrumentation__.Notify(40944)
				parallelism = 8
			} else {
				__antithesis_instrumentation__.Notify(40945)
			}
			__antithesis_instrumentation__.Notify(40941)

			var args []string
			if bazel.BuiltWithBazel() || func() bool {
				__antithesis_instrumentation__.Notify(40946)
				return forceBazel == true
			}() == true {
				__antithesis_instrumentation__.Notify(40947)
				args = append(args, "test")

				out, err := exec.Command("bazel", "query", fmt.Sprintf("kind(go_test, //%s:all)", name), "--output=label").Output()
				if err != nil {
					__antithesis_instrumentation__.Notify(40953)
					log.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(40954)
				}
				__antithesis_instrumentation__.Notify(40948)
				numTargets := 0
				for _, target := range strings.Split(string(out), "\n") {
					__antithesis_instrumentation__.Notify(40955)
					target = strings.TrimSpace(target)
					if target != "" {
						__antithesis_instrumentation__.Notify(40956)
						args = append(args, target)
						numTargets++
					} else {
						__antithesis_instrumentation__.Notify(40957)
					}
				}
				__antithesis_instrumentation__.Notify(40949)
				if numTargets == 0 {
					__antithesis_instrumentation__.Notify(40958)

					log.Printf("found no targets to test under package %s\n", name)
					continue
				} else {
					__antithesis_instrumentation__.Notify(40959)
				}
				__antithesis_instrumentation__.Notify(40950)
				args = append(args, "--")
				if target == "stressrace" {
					__antithesis_instrumentation__.Notify(40960)
					args = append(args, "--config=race")
				} else {
					__antithesis_instrumentation__.Notify(40961)
					args = append(args, "--test_sharding_strategy=disabled")
				}
				__antithesis_instrumentation__.Notify(40951)
				var filters []string
				for _, test := range pkg.tests {
					__antithesis_instrumentation__.Notify(40962)
					filters = append(filters, "^"+test+"$")
				}
				__antithesis_instrumentation__.Notify(40952)
				args = append(args, fmt.Sprintf("--test_filter=%s", strings.Join(filters, "|")))
				args = append(args, "--test_env=COCKROACH_NIGHTLY_STRESS=true")
				args = append(args, "--test_arg=-test.timeout", fmt.Sprintf("--test_arg=%s", timeout))

				args = append(args, fmt.Sprintf("--test_timeout=%d", int((duration+1*time.Minute).Seconds())))
				args = append(args, "--test_output", "streamed")

				args = append(args, "--run_under", fmt.Sprintf("%s -bazel -shardable-artifacts 'XML_OUTPUT_FILE=%s merge-test-xmls' -stderr -maxfails 1 -maxtime %s -p %d", bazelStressTarget, bazciPath, duration, parallelism))
				cmd := exec.Command("bazci", args...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				log.Println(cmd.Args)
				if err := cmd.Run(); err != nil {
					__antithesis_instrumentation__.Notify(40963)
					log.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(40964)
				}
			} else {
				__antithesis_instrumentation__.Notify(40965)
				tests := "-"
				if len(pkg.tests) > 0 {
					__antithesis_instrumentation__.Notify(40967)
					tests = "(" + strings.Join(pkg.tests, "$$|") + "$$)"
				} else {
					__antithesis_instrumentation__.Notify(40968)
				}
				__antithesis_instrumentation__.Notify(40966)

				args = append(
					args,
					target,
					fmt.Sprintf("PKG=./%s", name),
					fmt.Sprintf("TESTS=%s", tests),
					fmt.Sprintf("TESTTIMEOUT=%s", timeout),
					"GOTESTFLAGS=-json",
					fmt.Sprintf("STRESSFLAGS=-stderr -maxfails 1 -maxtime %s -p %d", duration, parallelism),
				)
				cmd := exec.Command("make", args...)
				cmd.Env = append(os.Environ(), "COCKROACH_NIGHTLY_STRESS=true")
				cmd.Dir = crdb
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				log.Println(cmd.Args)
				if err := cmd.Run(); err != nil {
					__antithesis_instrumentation__.Notify(40969)
					log.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(40970)
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(40971)
	}
}
