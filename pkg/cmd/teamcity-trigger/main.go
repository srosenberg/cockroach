// teamcity-trigger launches a variety of nightly build jobs on TeamCity using
// its REST API. It is intended to be run from a meta-build on a schedule
// trigger.
//
// One might think that TeamCity would support scheduling the same build to run
// multiple times with different parameters, but alas. The feature request has
// been open for ten years: https://youtrack.jetbrains.com/issue/TW-6439
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/abourget/teamcity"
	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/cmd/cmdutil"
	"github.com/kisielk/gotool"
)

func main() {
	__antithesis_instrumentation__.Notify(53060)
	if len(os.Args) != 1 {
		__antithesis_instrumentation__.Notify(53062)
		fmt.Fprintf(os.Stderr, "usage: %s\n", os.Args[0])
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(53063)
	}
	__antithesis_instrumentation__.Notify(53061)

	branch := cmdutil.RequireEnv("TC_BUILD_BRANCH")
	serverURL := cmdutil.RequireEnv("TC_SERVER_URL")
	username := cmdutil.RequireEnv("TC_API_USER")
	password := cmdutil.RequireEnv("TC_API_PASSWORD")

	tcClient := teamcity.New(serverURL, username, password)
	runTC(func(buildID string, opts map[string]string) {
		__antithesis_instrumentation__.Notify(53064)
		build, err := tcClient.QueueBuild(buildID, branch, opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(53066)
			log.Fatalf("failed to create teamcity build (buildID=%s, branch=%s, opts=%+v): %s",
				build, branch, opts, err)
		} else {
			__antithesis_instrumentation__.Notify(53067)
		}
		__antithesis_instrumentation__.Notify(53065)
		log.Printf("created teamcity build (buildID=%s, branch=%s, opts=%+v): %s",
			buildID, branch, opts, build)
	})
}

func getBaseImportPath() string {
	__antithesis_instrumentation__.Notify(53068)
	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(53070)
		return "./"
	} else {
		__antithesis_instrumentation__.Notify(53071)
	}
	__antithesis_instrumentation__.Notify(53069)
	return "github.com/cockroachdb/cockroach/pkg/"
}

func runTC(queueBuild func(string, map[string]string)) {
	__antithesis_instrumentation__.Notify(53072)
	buildID := "Cockroach_Nightlies_Stress"
	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(53074)
		buildID = "Cockroach_Nightlies_StressBazel"
	} else {
		__antithesis_instrumentation__.Notify(53075)
	}
	__antithesis_instrumentation__.Notify(53073)
	baseImportPath := getBaseImportPath()
	importPaths := gotool.ImportPaths([]string{baseImportPath + "..."})

	for _, importPath := range importPaths {
		__antithesis_instrumentation__.Notify(53076)

		maxRuns := 100

		maxTime := 1 * time.Hour

		maxFails := 1

		testTimeout := 40 * time.Minute

		parallelism := 4

		opts := map[string]string{
			"env.PKG": importPath,
		}

		switch importPath {
		case baseImportPath + "kv/kvnemesis":
			__antithesis_instrumentation__.Notify(53081)

			maxRuns = 0
			if bazel.BuiltWithBazel() {
				__antithesis_instrumentation__.Notify(53084)
				opts["env.EXTRA_BAZEL_FLAGS"] = "--test_env COCKROACH_KVNEMESIS_STEPS=10000"
			} else {
				__antithesis_instrumentation__.Notify(53085)
				opts["env.COCKROACH_KVNEMESIS_STEPS"] = "10000"
			}
		case baseImportPath + "sql/logictest", baseImportPath + "kv/kvserver":
			__antithesis_instrumentation__.Notify(53082)

			parallelism /= 2

			testTimeout = 2 * time.Hour
			maxTime = 3 * time.Hour
		default:
			__antithesis_instrumentation__.Notify(53083)
		}
		__antithesis_instrumentation__.Notify(53077)

		if bazel.BuiltWithBazel() {
			__antithesis_instrumentation__.Notify(53086)

			opts["env.TESTTIMEOUTSECS"] = fmt.Sprintf("%.0f", (maxTime + time.Minute).Seconds())
		} else {
			__antithesis_instrumentation__.Notify(53087)
			opts["env.TESTTIMEOUT"] = testTimeout.String()
		}
		__antithesis_instrumentation__.Notify(53078)

		if bazel.BuiltWithBazel() {
			__antithesis_instrumentation__.Notify(53088)
			bazelFlags, ok := opts["env.EXTRA_BAZEL_FLAGS"]
			if ok {
				__antithesis_instrumentation__.Notify(53089)
				opts["env.EXTRA_BAZEL_FLAGS"] = fmt.Sprintf("%s --test_sharding_strategy=disabled --jobs %d", bazelFlags, parallelism)
			} else {
				__antithesis_instrumentation__.Notify(53090)
				opts["env.EXTRA_BAZEL_FLAGS"] = fmt.Sprintf("--test_sharding_strategy=disabled --jobs %d", parallelism)
			}
		} else {
			__antithesis_instrumentation__.Notify(53091)
			opts["env.GOFLAGS"] = fmt.Sprintf("-parallel=%d", parallelism)
		}
		__antithesis_instrumentation__.Notify(53079)
		opts["env.STRESSFLAGS"] = fmt.Sprintf("-maxruns %d -maxtime %s -maxfails %d -p %d",
			maxRuns, maxTime, maxFails, parallelism)
		queueBuild(buildID, opts)

		opts["env.TAGS"] = "deadlock"
		queueBuild(buildID, opts)
		delete(opts, "env.TAGS")

		noParallelism := 1
		if bazel.BuiltWithBazel() {
			__antithesis_instrumentation__.Notify(53092)
			extraBazelFlags := opts["env.EXTRA_BAZEL_FLAGS"]

			opts["env.EXTRA_BAZEL_FLAGS"] = fmt.Sprintf("%s --@io_bazel_rules_go//go/config:race --test_env=GORACE=halt_on_error=1", extraBazelFlags)
		} else {
			__antithesis_instrumentation__.Notify(53093)
			opts["env.GOFLAGS"] = fmt.Sprintf("-race -parallel=%d", parallelism)
		}
		__antithesis_instrumentation__.Notify(53080)
		opts["env.STRESSFLAGS"] = fmt.Sprintf("-maxruns %d -maxtime %s -maxfails %d -p %d",
			maxRuns, maxTime, maxFails, noParallelism)
		queueBuild(buildID, opts)
	}
}
