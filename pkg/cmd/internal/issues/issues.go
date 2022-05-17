package issues

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/exec"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	CockroachPkgPrefix = "github.com/cockroachdb/cockroach/pkg/"

	githubIssueBodyMaximumLength = 60000
)

func enforceMaxLength(s string) string {
	__antithesis_instrumentation__.Notify(41134)
	if len(s) > githubIssueBodyMaximumLength {
		__antithesis_instrumentation__.Notify(41136)
		return s[:githubIssueBodyMaximumLength]
	} else {
		__antithesis_instrumentation__.Notify(41137)
	}
	__antithesis_instrumentation__.Notify(41135)
	return s
}

var (
	issueLabels = []string{"O-robot", "C-test-failure"}

	searchLabel = issueLabels[1]
)

type postCtx struct {
	context.Context
	strings.Builder
}

func (ctx *postCtx) Printf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(41138)
	if n := len(format); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(41140)
		return format[n-1] != '\n' == true
	}() == true {
		__antithesis_instrumentation__.Notify(41141)
		format += "\n"
	} else {
		__antithesis_instrumentation__.Notify(41142)
	}
	__antithesis_instrumentation__.Notify(41139)
	fmt.Fprintf(&ctx.Builder, format, args...)
}

func getLatestTag() (string, error) {
	__antithesis_instrumentation__.Notify(41143)
	cmd := exec.Command("git", "describe", "--abbrev=0", "--tags", "--match=v[0-9]*")
	out, err := cmd.CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(41145)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41146)
	}
	__antithesis_instrumentation__.Notify(41144)
	return strings.TrimSpace(string(out)), nil
}

func (p *poster) getProbableMilestone(ctx *postCtx) *int {
	__antithesis_instrumentation__.Notify(41147)
	tag, err := p.getLatestTag()
	if err != nil {
		__antithesis_instrumentation__.Notify(41152)
		ctx.Printf("unable to get latest tag to determine milestone: %s", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(41153)
	}
	__antithesis_instrumentation__.Notify(41148)

	v, err := version.Parse(tag)
	if err != nil {
		__antithesis_instrumentation__.Notify(41154)
		ctx.Printf("unable to parse version from tag to determine milestone: %s", err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(41155)
	}
	__antithesis_instrumentation__.Notify(41149)
	vstring := fmt.Sprintf("%d.%d", v.Major(), v.Minor())

	milestones, _, err := p.listMilestones(ctx, p.Org, p.Repo, &github.MilestoneListOptions{
		State: "open",
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41156)
		ctx.Printf("unable to list milestones for %s/%s: %v", p.Org, p.Repo, err)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(41157)
	}
	__antithesis_instrumentation__.Notify(41150)

	for _, m := range milestones {
		__antithesis_instrumentation__.Notify(41158)
		if m.GetTitle() == vstring {
			__antithesis_instrumentation__.Notify(41159)
			return m.Number
		} else {
			__antithesis_instrumentation__.Notify(41160)
		}
	}
	__antithesis_instrumentation__.Notify(41151)
	return nil
}

type poster struct {
	*Options

	createIssue func(ctx context.Context, owner string, repo string,
		issue *github.IssueRequest) (*github.Issue, *github.Response, error)
	searchIssues func(ctx context.Context, query string,
		opt *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error)
	createComment func(ctx context.Context, owner string, repo string, number int,
		comment *github.IssueComment) (*github.IssueComment, *github.Response, error)
	listCommits func(ctx context.Context, owner string, repo string,
		opts *github.CommitsListOptions) ([]*github.RepositoryCommit, *github.Response, error)
	listMilestones func(ctx context.Context, owner string, repo string,
		opt *github.MilestoneListOptions) ([]*github.Milestone, *github.Response, error)
	createProjectCard func(ctx context.Context, columnID int64,
		opt *github.ProjectCardOptions) (*github.ProjectCard, *github.Response, error)
}

func newPoster(client *github.Client, opts *Options) *poster {
	__antithesis_instrumentation__.Notify(41161)
	return &poster{
		Options:           opts,
		createIssue:       client.Issues.Create,
		searchIssues:      client.Search.Issues,
		createComment:     client.Issues.CreateComment,
		listCommits:       client.Repositories.ListCommits,
		listMilestones:    client.Issues.ListMilestones,
		createProjectCard: client.Projects.CreateProjectCard,
	}
}

type Options struct {
	Token        string
	Org          string
	Repo         string
	SHA          string
	BuildID      string
	ServerURL    string
	Branch       string
	Tags         string
	Goflags      string
	getLatestTag func() (string, error)
}

func DefaultOptionsFromEnv() *Options {
	__antithesis_instrumentation__.Notify(41162)

	const (
		githubOrgEnv           = "GITHUB_ORG"
		githubRepoEnv          = "GITHUB_REPO"
		githubAPITokenEnv      = "GITHUB_API_TOKEN"
		teamcityVCSNumberEnv   = "BUILD_VCS_NUMBER"
		teamcityBuildIDEnv     = "TC_BUILD_ID"
		teamcityServerURLEnv   = "TC_SERVER_URL"
		teamcityBuildBranchEnv = "TC_BUILD_BRANCH"
		tagsEnv                = "TAGS"
		goFlagsEnv             = "GOFLAGS"
	)

	return &Options{
		Token: maybeEnv(githubAPITokenEnv, ""),
		Org:   maybeEnv(githubOrgEnv, "cockroachdb"),
		Repo:  maybeEnv(githubRepoEnv, "cockroach"),

		SHA:          maybeEnv(teamcityVCSNumberEnv, "8548987813ff9e1b8a9878023d3abfc6911c16db"),
		BuildID:      maybeEnv(teamcityBuildIDEnv, "NOTFOUNDINENV"),
		ServerURL:    maybeEnv(teamcityServerURLEnv, "https://server-url-not-found-in-env.com"),
		Branch:       maybeEnv(teamcityBuildBranchEnv, "branch-not-found-in-env"),
		Tags:         maybeEnv(tagsEnv, ""),
		Goflags:      maybeEnv(goFlagsEnv, ""),
		getLatestTag: getLatestTag,
	}
}

func maybeEnv(envKey, defaultValue string) string {
	__antithesis_instrumentation__.Notify(41163)
	v := os.Getenv(envKey)
	if v == "" {
		__antithesis_instrumentation__.Notify(41165)
		return defaultValue
	} else {
		__antithesis_instrumentation__.Notify(41166)
	}
	__antithesis_instrumentation__.Notify(41164)
	return v
}

func (o *Options) CanPost() bool {
	__antithesis_instrumentation__.Notify(41167)
	return o.Token != ""
}

func (o *Options) IsReleaseBranch() bool {
	__antithesis_instrumentation__.Notify(41168)
	return o.Branch == "master" || func() bool {
		__antithesis_instrumentation__.Notify(41169)
		return strings.HasPrefix(o.Branch, "release-") == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(41170)
		return strings.HasPrefix(o.Branch, "provisional_") == true
	}() == true
}

type TemplateData struct {
	PostRequest

	PackageNameShort string

	Parameters []string

	CondensedMessage CondensedMessage

	Commit string

	CommitURL string

	Branch string

	ArtifactsURL string

	URL string

	RelatedIssues []github.Issue

	InternalLog string
}

func (p *poster) templateData(
	ctx context.Context, req PostRequest, relatedIssues []github.Issue,
) TemplateData {
	__antithesis_instrumentation__.Notify(41171)
	var artifactsURL string
	if req.Artifacts != "" {
		__antithesis_instrumentation__.Notify(41173)
		artifactsURL = p.teamcityArtifactsURL(req.Artifacts).String()
	} else {
		__antithesis_instrumentation__.Notify(41174)
	}
	__antithesis_instrumentation__.Notify(41172)
	return TemplateData{
		PostRequest:      req,
		Parameters:       p.parameters(),
		CondensedMessage: CondensedMessage(req.Message),
		Branch:           p.Branch,
		Commit:           p.SHA,
		ArtifactsURL:     artifactsURL,
		URL:              p.teamcityBuildLogURL().String(),
		RelatedIssues:    relatedIssues,
		PackageNameShort: strings.TrimPrefix(req.PackageName, CockroachPkgPrefix),
		CommitURL:        fmt.Sprintf("https://github.com/%s/%s/commits/%s", p.Org, p.Repo, p.SHA),
	}
}

func (p *poster) post(origCtx context.Context, formatter IssueFormatter, req PostRequest) error {
	__antithesis_instrumentation__.Notify(41175)
	ctx := &postCtx{Context: origCtx}
	data := p.templateData(
		ctx,
		req,
		nil,
	)

	title := formatter.Title(data)

	qBase := fmt.Sprintf(
		`repo:%q user:%q is:issue is:open in:title label:%q sort:created-desc %q`,
		p.Repo, p.Org, searchLabel, title)

	releaseLabel := fmt.Sprintf("branch-%s", p.Branch)
	qExisting := qBase + " label:" + releaseLabel
	qRelated := qBase + " -label:" + releaseLabel

	rExisting, _, err := p.searchIssues(ctx, qExisting, &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 1,
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41181)

		_ = err
		rExisting = &github.IssuesSearchResult{}
	} else {
		__antithesis_instrumentation__.Notify(41182)
	}
	__antithesis_instrumentation__.Notify(41176)

	rRelated, _, err := p.searchIssues(ctx, qRelated, &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 10,
		},
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(41183)

		_ = err
		rRelated = &github.IssuesSearchResult{}
	} else {
		__antithesis_instrumentation__.Notify(41184)
	}
	__antithesis_instrumentation__.Notify(41177)

	var foundIssue *int
	if len(rExisting.Issues) > 0 {
		__antithesis_instrumentation__.Notify(41185)

		foundIssue = rExisting.Issues[0].Number

		data.MentionOnCreate = nil
	} else {
		__antithesis_instrumentation__.Notify(41186)
	}
	__antithesis_instrumentation__.Notify(41178)

	data.RelatedIssues = rRelated.Issues
	data.InternalLog = ctx.Builder.String()
	r := &Renderer{}
	if err := formatter.Body(r, data); err != nil {
		__antithesis_instrumentation__.Notify(41187)

		_ = err
		fmt.Fprintln(&r.buf, "\nFailed to render body: "+err.Error())
	} else {
		__antithesis_instrumentation__.Notify(41188)
	}
	__antithesis_instrumentation__.Notify(41179)

	body := enforceMaxLength(r.buf.String())

	createLabels := append(issueLabels, releaseLabel)
	createLabels = append(createLabels, req.ExtraLabels...)
	if foundIssue == nil {
		__antithesis_instrumentation__.Notify(41189)
		issueRequest := github.IssueRequest{
			Title:     &title,
			Body:      github.String(body),
			Labels:    &createLabels,
			Milestone: p.getProbableMilestone(ctx),
		}
		issue, _, err := p.createIssue(ctx, p.Org, p.Repo, &issueRequest)
		if err != nil {
			__antithesis_instrumentation__.Notify(41191)
			return errors.Wrapf(err, "failed to create GitHub issue %s",
				github.Stringify(issueRequest))
		} else {
			__antithesis_instrumentation__.Notify(41192)
		}
		__antithesis_instrumentation__.Notify(41190)

		if req.ProjectColumnID != 0 {
			__antithesis_instrumentation__.Notify(41193)
			_, _, err := p.createProjectCard(ctx, int64(req.ProjectColumnID), &github.ProjectCardOptions{
				ContentID:   *issue.ID,
				ContentType: "Issue",
			})
			if err != nil {
				__antithesis_instrumentation__.Notify(41194)

				_ = err
			} else {
				__antithesis_instrumentation__.Notify(41195)
			}
		} else {
			__antithesis_instrumentation__.Notify(41196)
		}
	} else {
		__antithesis_instrumentation__.Notify(41197)
		comment := github.IssueComment{Body: github.String(body)}
		if _, _, err := p.createComment(
			ctx, p.Org, p.Repo, *foundIssue, &comment); err != nil {
			__antithesis_instrumentation__.Notify(41198)
			return errors.Wrapf(err, "failed to update issue #%d with %s",
				*foundIssue, github.Stringify(comment))
		} else {
			__antithesis_instrumentation__.Notify(41199)
		}
	}
	__antithesis_instrumentation__.Notify(41180)

	return nil
}

func (p *poster) teamcityURL(tab, fragment string) *url.URL {
	__antithesis_instrumentation__.Notify(41200)
	options := url.Values{}
	options.Add("buildId", p.BuildID)
	options.Add("tab", tab)

	u, err := url.Parse(p.ServerURL)
	if err != nil {
		__antithesis_instrumentation__.Notify(41202)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(41203)
	}
	__antithesis_instrumentation__.Notify(41201)
	u.Scheme = "https"
	u.Path = "viewLog.html"
	u.RawQuery = options.Encode()
	u.Fragment = fragment
	return u
}

func (p *poster) teamcityBuildLogURL() *url.URL {
	__antithesis_instrumentation__.Notify(41204)
	return p.teamcityURL("buildLog", "")
}

func (p *poster) teamcityArtifactsURL(artifacts string) *url.URL {
	__antithesis_instrumentation__.Notify(41205)
	return p.teamcityURL("artifacts", artifacts)
}

func (p *poster) parameters() []string {
	__antithesis_instrumentation__.Notify(41206)
	var ps []string
	if p.Tags != "" {
		__antithesis_instrumentation__.Notify(41209)
		ps = append(ps, "TAGS="+p.Tags)
	} else {
		__antithesis_instrumentation__.Notify(41210)
	}
	__antithesis_instrumentation__.Notify(41207)
	if p.Goflags != "" {
		__antithesis_instrumentation__.Notify(41211)
		ps = append(ps, "GOFLAGS="+p.Goflags)
	} else {
		__antithesis_instrumentation__.Notify(41212)
	}
	__antithesis_instrumentation__.Notify(41208)
	return ps
}

type PostRequest struct {
	PackageName string

	TestName string

	Message string

	Artifacts string

	MentionOnCreate []string

	HelpCommand func(*Renderer)

	ExtraLabels []string

	ProjectColumnID int
}

func Post(ctx context.Context, formatter IssueFormatter, req PostRequest) error {
	__antithesis_instrumentation__.Notify(41213)
	opts := DefaultOptionsFromEnv()
	if !opts.CanPost() {
		__antithesis_instrumentation__.Notify(41215)
		return errors.Newf("GITHUB_API_TOKEN env variable is not set; cannot post issue")
	} else {
		__antithesis_instrumentation__.Notify(41216)
	}
	__antithesis_instrumentation__.Notify(41214)

	client := github.NewClient(oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: opts.Token},
	)))
	return newPoster(client, opts).post(ctx, formatter, req)
}

func ReproductionCommandFromString(repro string) func(*Renderer) {
	__antithesis_instrumentation__.Notify(41217)
	if repro == "" {
		__antithesis_instrumentation__.Notify(41219)
		return func(*Renderer) { __antithesis_instrumentation__.Notify(41220) }
	} else {
		__antithesis_instrumentation__.Notify(41221)
	}
	__antithesis_instrumentation__.Notify(41218)
	return func(r *Renderer) {
		__antithesis_instrumentation__.Notify(41222)
		r.Escaped("To reproduce, try:\n")
		r.CodeBlock("bash", repro)
	}
}

func HelpCommandAsLink(title, href string) func(r *Renderer) {
	__antithesis_instrumentation__.Notify(41223)
	return func(r *Renderer) {
		__antithesis_instrumentation__.Notify(41224)

		r.Escaped("\n\nSee: ")
		r.A(title, href)
		r.Escaped("\n\n")
	}
}
