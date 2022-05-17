package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

const (
	dateFormatCommandLine = "2006-01-02"

	dateFormatEmail = "Monday, January 2"

	prepDate = "prep-date"

	publishDate = "publish-date"

	nextVersion = "next-version"

	NoProjectName = "No Project"

	eventRemovedProject = "removed_from_project"
	eventAddedProject   = "added_to_project"
)

var blockersFlags = struct {
	releaseSeries  string
	templatesDir   string
	prepDate       string
	publishDate    string
	nextVersion    string
	smtpUser       string
	smtpHost       string
	smtpPort       int
	emailAddresses []string
}{}

var postReleaseSeriesBlockersCmd = &cobra.Command{
	Use:   "post-blockers",
	Short: "Post blockers against release series",
	Long:  "Fetch post blockers against release series",
	RunE:  fetchReleaseSeriesBlockers,
}

func init() {
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.releaseSeries, releaseSeries, "", "major release series")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.templatesDir, templatesDir, "", "templates directory")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.prepDate, prepDate, "", "date to select candidate")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.publishDate, publishDate, "", "date to publish candidate")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.nextVersion, nextVersion, "", "next release version")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.smtpUser, smtpUser, os.Getenv(envSMTPUser), "SMTP user name")
	postReleaseSeriesBlockersCmd.Flags().StringVar(&blockersFlags.smtpHost, smtpHost, "", "SMTP host")
	postReleaseSeriesBlockersCmd.Flags().IntVar(&blockersFlags.smtpPort, smtpPort, 0, "SMTP port")
	postReleaseSeriesBlockersCmd.Flags().StringArrayVar(&blockersFlags.emailAddresses, emailAddresses, []string{}, "email addresses")

	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(releaseSeries)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(templatesDir)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(prepDate)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(publishDate)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(smtpHost)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(smtpPort)
	_ = postReleaseSeriesBlockersCmd.MarkFlagRequired(emailAddresses)
}

func fetchReleaseSeriesBlockers(_ *cobra.Command, _ []string) error {
	__antithesis_instrumentation__.Notify(42538)
	smtpPassword := os.Getenv(envSMTPPassword)
	if smtpPassword == "" {
		__antithesis_instrumentation__.Notify(42550)
		return fmt.Errorf("%s environment variable should be set", envSMTPPassword)
	} else {
		__antithesis_instrumentation__.Notify(42551)
	}
	__antithesis_instrumentation__.Notify(42539)
	githubToken := os.Getenv(envGithubToken)
	if githubToken == "" {
		__antithesis_instrumentation__.Notify(42552)
		return fmt.Errorf("%s environment variable should be set", envGithubToken)
	} else {
		__antithesis_instrumentation__.Notify(42553)
	}
	__antithesis_instrumentation__.Notify(42540)
	if blockersFlags.smtpUser == "" {
		__antithesis_instrumentation__.Notify(42554)
		return fmt.Errorf("either %s environment variable or %s flag should be set", envSMTPUser, smtpUser)
	} else {
		__antithesis_instrumentation__.Notify(42555)
	}
	__antithesis_instrumentation__.Notify(42541)
	releasePrepDate, err := time.Parse(dateFormatCommandLine, blockersFlags.prepDate)
	if err != nil {
		__antithesis_instrumentation__.Notify(42556)
		return fmt.Errorf("%s is not parseable into %s date layout", blockersFlags.prepDate, dateFormatCommandLine)
	} else {
		__antithesis_instrumentation__.Notify(42557)
	}
	__antithesis_instrumentation__.Notify(42542)
	releasePublishDate, err := time.Parse(dateFormatCommandLine, blockersFlags.publishDate)
	if err != nil {
		__antithesis_instrumentation__.Notify(42558)
		return fmt.Errorf("%s is not parseable into %s date layout", blockersFlags.publishDate, dateFormatCommandLine)
	} else {
		__antithesis_instrumentation__.Notify(42559)
	}
	__antithesis_instrumentation__.Notify(42543)
	if blockersFlags.nextVersion == "" {
		__antithesis_instrumentation__.Notify(42560)
		var err error
		blockersFlags.nextVersion, err = findNextVersion(blockersFlags.releaseSeries)
		if err != nil {
			__antithesis_instrumentation__.Notify(42561)
			return fmt.Errorf("cannot find next release version: %w", err)
		} else {
			__antithesis_instrumentation__.Notify(42562)
		}
	} else {
		__antithesis_instrumentation__.Notify(42563)
	}
	__antithesis_instrumentation__.Notify(42544)

	if !strings.HasPrefix(blockersFlags.nextVersion, fmt.Sprintf("v%s.", blockersFlags.releaseSeries)) {
		__antithesis_instrumentation__.Notify(42564)
		return fmt.Errorf("version %s does not match release series %s", blockersFlags.nextVersion, blockersFlags.releaseSeries)
	} else {
		__antithesis_instrumentation__.Notify(42565)
	}
	__antithesis_instrumentation__.Notify(42545)

	blockersURL := "go.crdb.dev/blockers/" + blockersFlags.releaseSeries
	releaseBranch := "release-" + blockersFlags.releaseSeries
	releaseBranchLabel := fmt.Sprintf("branch-release-%s", blockersFlags.releaseSeries)

	client := newGithubClient(context.Background(), githubToken)
	branchExists, err := client.branchExists(releaseBranch)
	if err != nil {
		__antithesis_instrumentation__.Notify(42566)
		return fmt.Errorf("cannot fetch branches: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42567)
	}
	__antithesis_instrumentation__.Notify(42546)
	if !branchExists {
		__antithesis_instrumentation__.Notify(42568)
		blockersURL = "go.crdb.dev/blockers"
		releaseBranch = "master"
		releaseBranchLabel = "branch-master"
	} else {
		__antithesis_instrumentation__.Notify(42569)
	}
	__antithesis_instrumentation__.Notify(42547)

	blockers, err := fetchOpenBlockers(client, releaseBranchLabel)
	if err != nil {
		__antithesis_instrumentation__.Notify(42570)
		return fmt.Errorf("cannot fetch blockers: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42571)
	}
	__antithesis_instrumentation__.Notify(42548)

	args := messageDataPostBlockers{
		Version:       blockersFlags.nextVersion,
		PrepDate:      releasePrepDate.Format(dateFormatEmail),
		ReleaseDate:   releasePublishDate.Format(dateFormatEmail),
		TotalBlockers: blockers.TotalBlockers,
		BlockersURL:   blockersURL,
		ReleaseBranch: releaseBranch,
		BlockerList:   blockers.BlockerList,
	}
	opts := sendOpts{
		templatesDir: blockersFlags.templatesDir,
		from:         fmt.Sprintf("Justin Beaver <%s>", blockersFlags.smtpUser),
		host:         blockersFlags.smtpHost,
		port:         blockersFlags.smtpPort,
		user:         blockersFlags.smtpUser,
		password:     smtpPassword,
		to:           blockersFlags.emailAddresses,
	}

	fmt.Println("Sending email")
	if err := sendMailPostBlockers(args, opts); err != nil {
		__antithesis_instrumentation__.Notify(42572)
		return fmt.Errorf("cannot send email: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42573)
	}
	__antithesis_instrumentation__.Notify(42549)
	return nil
}

type openBlockers struct {
	TotalBlockers int
	BlockerList   []ProjectBlocker
}

func fetchOpenBlockers(client githubClient, releaseBranchLabel string) (*openBlockers, error) {
	__antithesis_instrumentation__.Notify(42574)
	issues, err := client.openIssues([]string{
		"release-blocker",
		releaseBranchLabel,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(42579)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(42580)
	}
	__antithesis_instrumentation__.Notify(42575)
	if len(issues) == 0 {
		__antithesis_instrumentation__.Notify(42581)
		return &openBlockers{}, nil
	} else {
		__antithesis_instrumentation__.Notify(42582)
	}
	__antithesis_instrumentation__.Notify(42576)
	numBlockersByProject := make(map[string]int)
	for _, issue := range issues {
		__antithesis_instrumentation__.Notify(42583)
		if len(issue.ProjectName) == 0 {
			__antithesis_instrumentation__.Notify(42585)
			issue.ProjectName = NoProjectName
		} else {
			__antithesis_instrumentation__.Notify(42586)
		}
		__antithesis_instrumentation__.Notify(42584)
		total := numBlockersByProject[issue.ProjectName]
		numBlockersByProject[issue.ProjectName] = total + 1
	}
	__antithesis_instrumentation__.Notify(42577)
	var blockers projectBlockers
	for projectName, numBlockers := range numBlockersByProject {
		__antithesis_instrumentation__.Notify(42587)
		blockers = append(blockers, ProjectBlocker{
			ProjectName: projectName,
			NumBlockers: numBlockers,
		})
	}
	__antithesis_instrumentation__.Notify(42578)

	sort.Sort(blockers)
	return &openBlockers{
		TotalBlockers: len(issues),
		BlockerList:   blockers,
	}, nil
}

func mostRecentProjectName(client githubClient, issueNum int) (string, error) {
	__antithesis_instrumentation__.Notify(42588)
	removedProjects := make(map[string]bool)
	events, err := client.issueEvents(issueNum)
	if err != nil {
		__antithesis_instrumentation__.Notify(42591)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(42592)
	}
	__antithesis_instrumentation__.Notify(42589)

	sort.Sort(githubEvents(events))

	for i := len(events) - 1; i >= 0; i-- {
		__antithesis_instrumentation__.Notify(42593)
		projectName := events[i].ProjectName
		if len(projectName) == 0 {
			__antithesis_instrumentation__.Notify(42595)
			continue
		} else {
			__antithesis_instrumentation__.Notify(42596)
		}
		__antithesis_instrumentation__.Notify(42594)

		if events[i].Event == eventRemovedProject {
			__antithesis_instrumentation__.Notify(42597)

			removedProjects[projectName] = true
		} else {
			__antithesis_instrumentation__.Notify(42598)
			if events[i].Event == eventAddedProject {
				__antithesis_instrumentation__.Notify(42599)

				if _, exists := removedProjects[projectName]; exists {
					__antithesis_instrumentation__.Notify(42600)

					delete(removedProjects, projectName)
				} else {
					__antithesis_instrumentation__.Notify(42601)

					return projectName, err
				}
			} else {
				__antithesis_instrumentation__.Notify(42602)
			}
		}

	}
	__antithesis_instrumentation__.Notify(42590)

	return "", nil
}

type githubEvents []githubEvent

func (e githubEvents) Len() int {
	__antithesis_instrumentation__.Notify(42603)
	return len(e)
}

func (e githubEvents) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(42604)
	return e[i].CreatedAt.Before(e[j].CreatedAt)
}

func (e githubEvents) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(42605)
	e[i], e[j] = e[j], e[i]
}

type projectBlockers []ProjectBlocker

func (p projectBlockers) Len() int {
	__antithesis_instrumentation__.Notify(42606)
	return len(p)
}

func (p projectBlockers) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(42607)
	if p[i].ProjectName == NoProjectName {
		__antithesis_instrumentation__.Notify(42610)
		return false
	} else {
		__antithesis_instrumentation__.Notify(42611)
	}
	__antithesis_instrumentation__.Notify(42608)
	if p[j].ProjectName == NoProjectName {
		__antithesis_instrumentation__.Notify(42612)
		return true
	} else {
		__antithesis_instrumentation__.Notify(42613)
	}
	__antithesis_instrumentation__.Notify(42609)
	return p[i].ProjectName < p[j].ProjectName
}

func (p projectBlockers) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(42614)
	p[i], p[j] = p[j], p[i]
}
