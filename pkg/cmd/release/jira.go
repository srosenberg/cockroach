package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/andygrunwald/go-jira"
)

const jiraBaseURL = "https://cockroachlabs.atlassian.net/"
const dryRunProject = "RE"

const customFieldHasSLAKey = "customfield_10073"

const trackingIssueTemplate = `
* Version: *{{ .Version }}*
* SHA: [{{ .SHA }}|https://github.com/cockroachlabs/release-staging/commit/{{ .SHA }}]
* Tag: [{{ .Tag }}|https://github.com/cockroachlabs/release-staging/releases/tag/{{ .Tag }}]
* SRE issue: [{{ .SREIssue }}]
* Deployment status: _fillmein_
* Publish Cockroach Release: _fillmein_

h2. [Release process checklist|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process]

* Assign the SRE issue [{{ .SREIssue }}] (use "/genie whoisoncall" in Slack). They will be notified by Jira.
* [5-8. Verify node crash reports|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/328859690/Release+Qualification#Verify-node-crash-reports-appear-in-sentry.io]

h2. Do not proceed below until the release date.

Release date: _fillmein_

* [9. Publish Cockroach Release|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-9.PublishTheRelease]
* Ack security@ and release-engineering-team@ on the generated AWS S3 bucket write alert to confirm these writes were part of a planned release
* [10. Check binaries|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-10.CheckBinaries]
* [12. Announce the release is cut to releases@|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-13.Announcethereleaseiscuttoreleases@]
* [13. Update version numbers|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-14.Updateversionnumbers]
* For production or stable releases in the latest [major release|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-Knowifthereleaseisonthelatestmajorreleaseseries] series only (in August 2020, this is the v20.1 series):
* Update [Brew Recipe|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-Brewrecipe]
* Update [Orchestrator configurations:CRDB|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-Orchestratorconfigurations:CRDB]
* Update [Orchestrator configurations:Helm Charts|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-Orchestratorconfigurations:HelmCharts]

For all production or stable releases:
* Create a ticket in the [Dev Inf tracker|https://cockroachlabs.atlassian.net/wiki/spaces/devinf/pages/429097164/Submitting+Issues+Requests+to+the+Developer+Infrastructure+team] to update the Red Hat Container Image Repository
* *After docs are updated* [Announce version to registration cluster|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-AnnounceVersionToRegCluster]
* [Update version map in bin/roachtest (all stable releases) and regenerate test fixtures (only major release)|https://cockroachlabs.atlassian.net/wiki/spaces/ENG/pages/73105625/Release+process#Releaseprocess-Updateversionmapinroachtestandregeneratetestfixtures]
* Update docs (handled by Docs team)
* External communications for release (handled by Marketing team)
`
const sreIssueTemplate = `
Could you deploy the Docker image with the following tag to the release qualification CC cluster?

* Version: {{ .Version }}
* Build ID: {{ .Tag }}

Please follow [this playbook|https://github.com/cockroachlabs/production/wiki/Deploy-release-qualification-versions]

Thank you\!
`

type jiraClient struct {
	client *jira.Client
}

type trackingIssueTemplateArgs struct {
	Version  string
	SHA      string
	Tag      string
	SREIssue string
}

type sreIssueTemplateArgs struct {
	Version string
	Tag     string
}

type jiraIssue struct {
	ID           string
	Key          string
	TypeName     string
	ProjectKey   string
	Summary      string
	Description  string
	CustomFields jira.CustomFields
}

func newJiraClient(baseURL string, username string, password string) (*jiraClient, error) {
	__antithesis_instrumentation__.Notify(42738)
	tp := jira.BasicAuthTransport{
		Username: username,
		Password: password,
	}
	client, err := jira.NewClient(tp.Client(), baseURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(42740)
		return nil, fmt.Errorf("cannot create Jira client: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42741)
	}
	__antithesis_instrumentation__.Notify(42739)
	return &jiraClient{
		client: client,
	}, nil
}

func (j *jiraClient) getIssueDetails(issueID string) (jiraIssue, error) {
	__antithesis_instrumentation__.Notify(42742)
	issue, _, err := j.client.Issue.Get(issueID, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(42745)
		return jiraIssue{}, err
	} else {
		__antithesis_instrumentation__.Notify(42746)
	}
	__antithesis_instrumentation__.Notify(42743)
	customFields, _, err := j.client.Issue.GetCustomFields(issueID)
	if err != nil {
		__antithesis_instrumentation__.Notify(42747)
		return jiraIssue{}, err
	} else {
		__antithesis_instrumentation__.Notify(42748)
	}
	__antithesis_instrumentation__.Notify(42744)
	return jiraIssue{
		ID:           issue.ID,
		Key:          issue.Key,
		TypeName:     issue.Fields.Type.Name,
		ProjectKey:   issue.Fields.Project.Name,
		Summary:      issue.Fields.Summary,
		Description:  issue.Fields.Description,
		CustomFields: customFields,
	}, nil
}

func newIssue(details *jiraIssue) *jira.Issue {
	__antithesis_instrumentation__.Notify(42749)
	var issue jira.Issue
	issue.Fields = &jira.IssueFields{}
	issue.Fields.Project = jira.Project{
		Key: details.ProjectKey,
	}
	issue.Fields.Type = jira.IssueType{
		Name: details.TypeName,
	}
	issue.Fields.Summary = details.Summary
	issue.Fields.Description = details.Description

	if details.CustomFields != nil {
		__antithesis_instrumentation__.Notify(42751)
		issue.Fields.Unknowns = make(map[string]interface{})
		for key, value := range details.CustomFields {
			__antithesis_instrumentation__.Notify(42752)
			issue.Fields.Unknowns[key] = map[string]string{"value": value}
		}
	} else {
		__antithesis_instrumentation__.Notify(42753)
	}
	__antithesis_instrumentation__.Notify(42750)
	return &issue
}

func (d jiraIssue) url() string {
	__antithesis_instrumentation__.Notify(42754)
	return fmt.Sprintf("%s/browse/%s", strings.TrimSuffix(jiraBaseURL, "/"), d.Key)
}

func createJiraIssue(client *jiraClient, issue *jira.Issue) (jiraIssue, error) {
	__antithesis_instrumentation__.Notify(42755)
	newIssue, _, err := client.client.Issue.Create(issue)
	if err != nil {
		__antithesis_instrumentation__.Notify(42758)
		return jiraIssue{}, err
	} else {
		__antithesis_instrumentation__.Notify(42759)
	}
	__antithesis_instrumentation__.Notify(42756)
	details, err := client.getIssueDetails(newIssue.ID)
	if err != nil {
		__antithesis_instrumentation__.Notify(42760)
		return jiraIssue{}, err
	} else {
		__antithesis_instrumentation__.Notify(42761)
	}
	__antithesis_instrumentation__.Notify(42757)
	return details, nil
}

func createTrackingIssue(
	client *jiraClient, release releaseInfo, sreIssue jiraIssue, dryRun bool,
) (jiraIssue, error) {
	__antithesis_instrumentation__.Notify(42762)
	templateArgs := trackingIssueTemplateArgs{
		Version:  release.nextReleaseVersion,
		Tag:      release.buildInfo.Tag,
		SHA:      release.buildInfo.SHA,
		SREIssue: sreIssue.Key,
	}
	description, err := templateToText(trackingIssueTemplate, templateArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(42765)
		return jiraIssue{}, fmt.Errorf("cannot parse tracking issue template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42766)
	}
	__antithesis_instrumentation__.Notify(42763)
	summary := fmt.Sprintf("Release: %s", release.nextReleaseVersion)
	projectKey := "RE"
	if dryRun {
		__antithesis_instrumentation__.Notify(42767)
		projectKey = dryRunProject
	} else {
		__antithesis_instrumentation__.Notify(42768)
	}
	__antithesis_instrumentation__.Notify(42764)
	issue := newIssue(&jiraIssue{

		ProjectKey: projectKey,

		TypeName:    "Task",
		Summary:     summary,
		Description: description,
	})
	return createJiraIssue(client, issue)
}

func createSREIssue(client *jiraClient, release releaseInfo, dryRun bool) (jiraIssue, error) {
	__antithesis_instrumentation__.Notify(42769)
	templateArgs := sreIssueTemplateArgs{
		Version: release.nextReleaseVersion,
		Tag:     release.buildInfo.Tag,
	}
	description, err := templateToHTML(sreIssueTemplate, templateArgs)
	if err != nil {
		__antithesis_instrumentation__.Notify(42772)
		return jiraIssue{}, fmt.Errorf("cannot parse SRE issue template: %w", err)
	} else {
		__antithesis_instrumentation__.Notify(42773)
	}
	__antithesis_instrumentation__.Notify(42770)
	projectKey := "SREOPS"
	summary := fmt.Sprintf("Deploy %s to release qualification cluster", release.nextReleaseVersion)
	customFields := make(jira.CustomFields)
	customFields[customFieldHasSLAKey] = "Yes"
	if dryRun {
		__antithesis_instrumentation__.Notify(42774)
		projectKey = dryRunProject
		customFields = nil
	} else {
		__antithesis_instrumentation__.Notify(42775)
	}
	__antithesis_instrumentation__.Notify(42771)
	issue := newIssue(&jiraIssue{
		ProjectKey:   projectKey,
		TypeName:     "Task",
		Summary:      summary,
		Description:  description,
		CustomFields: customFields,
	})
	return createJiraIssue(client, issue)
}
