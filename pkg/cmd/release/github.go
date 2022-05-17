package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v42/github"
	"golang.org/x/oauth2"
)

const (
	githubPageSize = 100

	openIssueMaxPages = 3

	timelineEventMaxPages = 3
)

type githubClient interface {
	openIssues(labels []string) ([]githubIssue, error)
	issueEvents(issueNum int) ([]githubEvent, error)
}

type githubClientImpl struct {
	client *github.Client
	ctx    context.Context
	owner  string
	repo   string
}

var _ githubClient = &githubClientImpl{}

type githubIssue struct {
	Number      int
	ProjectName string
}

type githubEvent struct {
	CreatedAt time.Time

	Event       string
	ProjectName string
}

func newGithubClient(ctx context.Context, accessToken string) *githubClientImpl {
	__antithesis_instrumentation__.Notify(42697)
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	tc := oauth2.NewClient(ctx, ts)
	return &githubClientImpl{
		client: github.NewClient(tc),
		ctx:    ctx,
		owner:  "cockroachdb",
		repo:   "cockroach",
	}
}

func (c *githubClientImpl) openIssues(labels []string) ([]githubIssue, error) {
	__antithesis_instrumentation__.Notify(42698)
	var details []githubIssue
	for pageNum := 1; pageNum <= openIssueMaxPages; pageNum++ {
		__antithesis_instrumentation__.Notify(42700)

		issues, _, err := c.client.Issues.List(
			c.ctx,
			true,
			&github.IssueListOptions{
				Filter: "all",
				State:  "open",
				Labels: labels,
				ListOptions: github.ListOptions{
					PerPage: githubPageSize,
					Page:    pageNum,
				},
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(42703)
			return nil, fmt.Errorf("error calling Issues.List: %w", err)
		} else {
			__antithesis_instrumentation__.Notify(42704)
		}
		__antithesis_instrumentation__.Notify(42701)
		for _, i := range issues {
			__antithesis_instrumentation__.Notify(42705)
			issueNum := *i.Number
			projectName, err := mostRecentProjectName(c, issueNum)
			if err != nil {
				__antithesis_instrumentation__.Notify(42707)
				return nil, fmt.Errorf("error calling mostRecentProjectName: %w", err)
			} else {
				__antithesis_instrumentation__.Notify(42708)
			}
			__antithesis_instrumentation__.Notify(42706)

			details = append(details, githubIssue{
				Number:      issueNum,
				ProjectName: projectName,
			})
		}
		__antithesis_instrumentation__.Notify(42702)

		if len(issues) < githubPageSize {
			__antithesis_instrumentation__.Notify(42709)
			break
		} else {
			__antithesis_instrumentation__.Notify(42710)
		}
	}
	__antithesis_instrumentation__.Notify(42699)

	return details, nil
}

func (c *githubClientImpl) branchExists(branchName string) (bool, error) {
	__antithesis_instrumentation__.Notify(42711)
	protected := true
	branches, _, err := c.client.Repositories.ListBranches(
		c.ctx, c.owner, c.repo,
		&github.BranchListOptions{Protected: &protected},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(42714)
		return false, fmt.Errorf("error calling Repositories.ListBranches: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42715)
	}
	__antithesis_instrumentation__.Notify(42712)
	for _, b := range branches {
		__antithesis_instrumentation__.Notify(42716)
		if *b.Name == branchName {
			__antithesis_instrumentation__.Notify(42717)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(42718)
		}
	}
	__antithesis_instrumentation__.Notify(42713)
	return false, nil
}

func (c *githubClientImpl) issueEvents(issueNum int) ([]githubEvent, error) {
	__antithesis_instrumentation__.Notify(42719)
	var details []githubEvent

	for pageNum := 1; pageNum <= timelineEventMaxPages; pageNum++ {
		__antithesis_instrumentation__.Notify(42721)
		events, _, err := c.client.Issues.ListIssueTimeline(
			c.ctx, c.owner, c.repo, issueNum, &github.ListOptions{
				PerPage: githubPageSize,
				Page:    pageNum,
			})
		if err != nil {
			__antithesis_instrumentation__.Notify(42724)
			return nil, fmt.Errorf("error calling Issues.ListIssueTimeline: %w", err)
		} else {
			__antithesis_instrumentation__.Notify(42725)
		}
		__antithesis_instrumentation__.Notify(42722)
		for _, event := range events {
			__antithesis_instrumentation__.Notify(42726)
			detail := githubEvent{
				CreatedAt: *event.CreatedAt,
				Event:     *event.Event,
			}
			if event.ProjectCard != nil && func() bool {
				__antithesis_instrumentation__.Notify(42728)
				return event.ProjectCard.ProjectID != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(42729)
				project, _, err := c.client.Projects.GetProject(c.ctx, *event.ProjectCard.ProjectID)
				if err != nil {
					__antithesis_instrumentation__.Notify(42731)
					return nil, fmt.Errorf("error calling Projects.GetProject: %w", err)
				} else {
					__antithesis_instrumentation__.Notify(42732)
				}
				__antithesis_instrumentation__.Notify(42730)
				if project.Name != nil {
					__antithesis_instrumentation__.Notify(42733)
					detail.ProjectName = *project.Name
				} else {
					__antithesis_instrumentation__.Notify(42734)
				}
			} else {
				__antithesis_instrumentation__.Notify(42735)
			}
			__antithesis_instrumentation__.Notify(42727)
			details = append(details, detail)
		}
		__antithesis_instrumentation__.Notify(42723)
		if len(events) < githubPageSize {
			__antithesis_instrumentation__.Notify(42736)
			break
		} else {
			__antithesis_instrumentation__.Notify(42737)
		}
	}
	__antithesis_instrumentation__.Notify(42720)
	return details, nil
}
