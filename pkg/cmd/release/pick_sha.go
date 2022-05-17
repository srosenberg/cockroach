package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"html/template"
	"os"

	"github.com/spf13/cobra"
)

const (
	qualifyBucket       = "qualify-bucket"
	qualifyObjectPrefix = "qualify-object-prefix"
	releaseBucket       = "release-bucket"
	releaseObjectPrefix = "release-object-prefix"
)

var pickSHAFlags = struct {
	qualifyBucket       string
	qualifyObjectPrefix string
	releaseBucket       string
	releaseObjectPrefix string
	releaseSeries       string
	templatesDir        string
	smtpUser            string
	smtpHost            string
	smtpPort            int
	emailAddresses      []string
	dryRun              bool
}{}

var pickSHACmd = &cobra.Command{
	Use:   "pick-sha",
	Short: "Pick release git SHA for a particular version and communicate the choice via Jira and email",

	Long: "Pick release git SHA for a particular version and communicate the choice via Jira and email",
	RunE: pickSHA,
}

func init() {

	pickSHACmd.Flags().StringVar(&pickSHAFlags.qualifyBucket, qualifyBucket, "", "release qualification metadata GCS bucket")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.qualifyObjectPrefix, qualifyObjectPrefix, "",
		"release qualification object prefix")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.releaseBucket, releaseBucket, "", "release candidates metadata GCS bucket")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.releaseObjectPrefix, releaseObjectPrefix, "", "release candidate object prefix")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.releaseSeries, releaseSeries, "", "major release series")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.templatesDir, templatesDir, "", "templates directory")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.smtpUser, smtpUser, os.Getenv(envSMTPUser), "SMTP user name")
	pickSHACmd.Flags().StringVar(&pickSHAFlags.smtpHost, smtpHost, "", "SMTP host")
	pickSHACmd.Flags().IntVar(&pickSHAFlags.smtpPort, smtpPort, 0, "SMTP port")
	pickSHACmd.Flags().StringArrayVar(&pickSHAFlags.emailAddresses, emailAddresses, []string{}, "email addresses")
	pickSHACmd.Flags().BoolVar(&pickSHAFlags.dryRun, dryRun, false, "use dry run Jira project for issues")

	requiredFlags := []string{
		qualifyBucket,
		qualifyObjectPrefix,
		releaseBucket,
		releaseObjectPrefix,
		releaseSeries,
		smtpUser,
		smtpHost,
		smtpPort,
		emailAddresses,
	}
	for _, flag := range requiredFlags {
		if err := pickSHACmd.MarkFlagRequired(flag); err != nil {
			panic(err)
		}
	}
}

func pickSHA(_ *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(42837)
	smtpPassword := os.Getenv("SMTP_PASSWORD")
	if smtpPassword == "" {
		__antithesis_instrumentation__.Notify(42847)
		return fmt.Errorf("SMTP_PASSWORD environment variable should be set")
	} else {
		__antithesis_instrumentation__.Notify(42848)
	}
	__antithesis_instrumentation__.Notify(42838)
	jiraUsername := os.Getenv("JIRA_USERNAME")
	if jiraUsername == "" {
		__antithesis_instrumentation__.Notify(42849)
		return fmt.Errorf("JIRA_USERNAME environment variable should be set")
	} else {
		__antithesis_instrumentation__.Notify(42850)
	}
	__antithesis_instrumentation__.Notify(42839)
	jiraToken := os.Getenv("JIRA_TOKEN")
	if jiraToken == "" {
		__antithesis_instrumentation__.Notify(42851)
		return fmt.Errorf("JIRA_TOKEN environment variable should be set")
	} else {
		__antithesis_instrumentation__.Notify(42852)
	}
	__antithesis_instrumentation__.Notify(42840)

	nextRelease, err := findNextRelease(pickSHAFlags.releaseSeries)
	if err != nil {
		__antithesis_instrumentation__.Notify(42853)
		return fmt.Errorf("cannot find next release: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42854)
	}
	__antithesis_instrumentation__.Notify(42841)

	fmt.Println("Previous version:", nextRelease.prevReleaseVersion)
	fmt.Println("Next version:", nextRelease.nextReleaseVersion)
	fmt.Println("Release SHA:", nextRelease.buildInfo.SHA)

	releaseInfoPath := fmt.Sprintf("%s/%s.json", pickSHAFlags.releaseObjectPrefix, nextRelease.nextReleaseVersion)
	fmt.Println("Publishing release candidate metadata")
	if err := publishReleaseCandidateInfo(context.Background(), nextRelease, pickSHAFlags.releaseBucket, releaseInfoPath); err != nil {
		__antithesis_instrumentation__.Notify(42855)
		return fmt.Errorf("cannot publish release metadata: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42856)
	}
	__antithesis_instrumentation__.Notify(42842)

	fmt.Println("Creating SRE issue")
	jiraClient, err := newJiraClient(jiraBaseURL, jiraUsername, jiraToken)
	if err != nil {
		__antithesis_instrumentation__.Notify(42857)
		return fmt.Errorf("cannot create Jira client: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42858)
	}
	__antithesis_instrumentation__.Notify(42843)
	sreIssue, err := createSREIssue(jiraClient, nextRelease, pickSHAFlags.dryRun)
	if err != nil {
		__antithesis_instrumentation__.Notify(42859)
		return err
	} else {
		__antithesis_instrumentation__.Notify(42860)
	}
	__antithesis_instrumentation__.Notify(42844)

	fmt.Println("Creating tracking issue")
	trackingIssue, err := createTrackingIssue(jiraClient, nextRelease, sreIssue, pickSHAFlags.dryRun)
	if err != nil {
		__antithesis_instrumentation__.Notify(42861)
		return fmt.Errorf("cannot create tracking issue: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42862)
	}
	__antithesis_instrumentation__.Notify(42845)
	diffURL := template.URL(
		fmt.Sprintf("https://github.com/cockroachdb/cockroach/compare/%s...%s",
			nextRelease.prevReleaseVersion,
			nextRelease.buildInfo.SHA))
	args := messageDataPickSHA{
		Version:          nextRelease.nextReleaseVersion,
		SHA:              nextRelease.buildInfo.SHA,
		TrackingIssue:    trackingIssue.Key,
		TrackingIssueURL: template.URL(trackingIssue.url()),
		DiffURL:          diffURL,
	}
	opts := sendOpts{
		templatesDir: pickSHAFlags.templatesDir,
		from:         fmt.Sprintf("Justin Beaver <%s>", pickSHAFlags.smtpUser),
		host:         pickSHAFlags.smtpHost,
		port:         pickSHAFlags.smtpPort,
		user:         pickSHAFlags.smtpUser,
		password:     smtpPassword,
		to:           pickSHAFlags.emailAddresses,
	}
	fmt.Println("Sending email")
	if err := sendMailPickSHA(args, opts); err != nil {
		__antithesis_instrumentation__.Notify(42863)
		return fmt.Errorf("cannot send email: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42864)
	}
	__antithesis_instrumentation__.Notify(42846)
	return nil
}
