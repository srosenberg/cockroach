// Check that GitHub PR descriptions and commit messages contain the
// expected epic and issue references.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "docs-issue-generation",
	Short: "Generate a new set of release issues in the docs repo for a given commit.",
	Run: func(_ *cobra.Command, args []string) {
		__antithesis_instrumentation__.Notify(40052)
		params := defaultEnvParameters()
		docsIssueGeneration(params)
	},
}

func main() {
	__antithesis_instrumentation__.Notify(40053)
	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(40054)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(40055)
	}
}

func defaultEnvParameters() parameters {
	__antithesis_instrumentation__.Notify(40056)
	const (
		githubAPITokenEnv = "GITHUB_API_TOKEN"
		buildVcsNumberEnv = "BUILD_VCS_NUMBER"
	)

	return parameters{
		Token: maybeEnv(githubAPITokenEnv, ""),
		Sha:   maybeEnv(buildVcsNumberEnv, "4dd8da9609adb3acce6795cea93b67ccacfc0270"),
	}
}

func maybeEnv(envKey, defaultValue string) string {
	__antithesis_instrumentation__.Notify(40057)
	v := os.Getenv(envKey)
	if v == "" {
		__antithesis_instrumentation__.Notify(40059)
		return defaultValue
	} else {
		__antithesis_instrumentation__.Notify(40060)
	}
	__antithesis_instrumentation__.Notify(40058)
	return v
}
