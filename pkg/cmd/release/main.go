package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{Use: "release"}

const (
	envSMTPUser     = "SMTP_USER"
	envSMTPPassword = "SMTP_PASSWORD"
	envGithubToken  = "GITHUB_TOKEN"
	releaseSeries   = "release-series"
	templatesDir    = "template-dir"
	smtpUser        = "smtp-user"
	smtpHost        = "smtp-host"
	smtpPort        = "smtp-port"
	emailAddresses  = "to"
	dryRun          = "dry-run"
)

func main() {
	__antithesis_instrumentation__.Notify(42776)
	if err := rootCmd.Execute(); err != nil {
		__antithesis_instrumentation__.Notify(42777)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(42778)
	}
}

func init() {
	rootCmd.AddCommand(pickSHACmd)
	rootCmd.AddCommand(postReleaseSeriesBlockersCmd)
	rootCmd.AddCommand(setOrchestrationVersionCmd)
}
