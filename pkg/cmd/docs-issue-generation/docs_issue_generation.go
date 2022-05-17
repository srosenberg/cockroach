package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type commit struct {
	sha         string
	title       string
	releaseNote string
}

type pr struct {
	number      int
	mergeBranch string
	commits     []commit
}

type ghSearch struct {
	Items []ghSearchItem `json:"items"`
}

type ghSearchItem struct {
	Number      int            `json:"number"`
	PullRequest ghSearchItemPr `json:"pull_request"`
}

type ghSearchItemPr struct {
	MergedAt *time.Time `json:"merged_at"`
}

type ghPull struct {
	Base struct {
		Ref string `json:"ref"`
	} `json:"base"`
}

type ghPullCommit struct {
	Sha    string          `json:"sha"`
	Commit ghPullCommitMsg `json:"commit"`
}

type ghPullCommitMsg struct {
	Message string `json:"message"`
}

type parameters struct {
	Token string
	Sha   string
}

func docsIssueGeneration(params parameters) {
	__antithesis_instrumentation__.Notify(40002)
	var search ghSearch
	var prs []pr
	if err := httpGet("https://api.github.com/search/issues?q=sha:"+params.Sha+"+repo:cockroachdb/cockroach+is:merged", params.Token, &search); err != nil {
		__antithesis_instrumentation__.Notify(40004)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40005)
	}
	__antithesis_instrumentation__.Notify(40003)
	prs = prNums(search)
	for _, pr := range prs {
		__antithesis_instrumentation__.Notify(40006)
		var pull ghPull
		if err := httpGet("https://api.github.com/repos/cockroachdb/cockroach/pulls/"+strconv.Itoa(pr.number), params.Token, &pull); err != nil {
			__antithesis_instrumentation__.Notify(40009)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40010)
		}
		__antithesis_instrumentation__.Notify(40007)
		pr.mergeBranch = pull.Base.Ref
		var commits []ghPullCommit
		if err := httpGet("https://api.github.com/repos/cockroachdb/cockroach/pulls/"+strconv.Itoa(pr.number)+"/commits?per_page:250", params.Token, &commits); err != nil {
			__antithesis_instrumentation__.Notify(40011)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40012)
		}
		__antithesis_instrumentation__.Notify(40008)
		pr.commits = getCommits(commits, pr.number)
		pr.createDocsIssues(params.Token)
	}
}

func httpGet(url string, token string, out interface{}) error {
	__antithesis_instrumentation__.Notify(40013)
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(40018)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40019)
	}
	__antithesis_instrumentation__.Notify(40014)
	req.Header.Set("Authorization", "token "+token)
	res, err := client.Do(req)
	if err != nil {
		__antithesis_instrumentation__.Notify(40020)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40021)
	}
	__antithesis_instrumentation__.Notify(40015)
	bs, err := io.ReadAll(res.Body)
	if err != nil {
		__antithesis_instrumentation__.Notify(40022)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40023)
	}
	__antithesis_instrumentation__.Notify(40016)
	if err := json.Unmarshal(bs, out); err != nil {
		__antithesis_instrumentation__.Notify(40024)
		return err
	} else {
		__antithesis_instrumentation__.Notify(40025)
	}
	__antithesis_instrumentation__.Notify(40017)
	return nil
}

func prNums(search ghSearch) []pr {
	__antithesis_instrumentation__.Notify(40026)
	var result []pr
	for _, x := range search.Items {
		__antithesis_instrumentation__.Notify(40028)
		if x.PullRequest.MergedAt != nil {
			__antithesis_instrumentation__.Notify(40029)
			result = append(result, pr{number: x.Number})
		} else {
			__antithesis_instrumentation__.Notify(40030)
		}
	}
	__antithesis_instrumentation__.Notify(40027)
	return result
}

func getCommits(pullCommit []ghPullCommit, prNumber int) []commit {
	__antithesis_instrumentation__.Notify(40031)
	var result []commit
	for _, c := range pullCommit {
		__antithesis_instrumentation__.Notify(40033)
		message := c.Commit.Message
		rn := formatReleaseNote(message, prNumber, c.Sha)
		if rn != "" {
			__antithesis_instrumentation__.Notify(40034)
			x := commit{
				sha:         c.Sha,
				title:       formatTitle(message),
				releaseNote: rn,
			}
			result = append(result, x)
		} else {
			__antithesis_instrumentation__.Notify(40035)
		}
	}
	__antithesis_instrumentation__.Notify(40032)
	return result
}

func formatReleaseNote(message string, prNumber int, sha string) string {
	__antithesis_instrumentation__.Notify(40036)
	re := regexp.MustCompile(`(?s)[rR]elease [nN]ote \(.*`)
	reNeg := regexp.MustCompile(`([rR]elease [nN]ote \(bug fix.*)|([rR]elease [nN]ote: [nN]one)`)
	rn := re.FindString(message)
	if len(rn) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(40038)
		return !reNeg.MatchString(message) == true
	}() == true {
		__antithesis_instrumentation__.Notify(40039)
		return fmt.Sprintf(
			"Related PR: https://github.com/cockroachdb/cockroach/pull/%s\nCommit: https://github.com/cockroachdb/cockroach/commit/%s\n\n---\n\n%s",
			strconv.Itoa(prNumber),
			sha,
			rn,
		)
	} else {
		__antithesis_instrumentation__.Notify(40040)
	}
	__antithesis_instrumentation__.Notify(40037)
	return ""
}

func formatTitle(message string) string {
	__antithesis_instrumentation__.Notify(40041)
	if i := strings.IndexRune(message, '\n'); i > 0 {
		__antithesis_instrumentation__.Notify(40043)
		return message[:i]
	} else {
		__antithesis_instrumentation__.Notify(40044)
	}
	__antithesis_instrumentation__.Notify(40042)
	return message
}

func (p pr) createDocsIssues(token string) {
	__antithesis_instrumentation__.Notify(40045)
	postURL := "https://api.github.com/repos/cockroachdb/docs/issues"
	for _, commit := range p.commits {
		__antithesis_instrumentation__.Notify(40046)
		reqBody, err := json.Marshal(map[string]interface{}{
			"title":  commit.title,
			"labels": []string{"C-product-change", p.mergeBranch},
			"body":   commit.releaseNote,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(40048)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40049)
		}
		__antithesis_instrumentation__.Notify(40047)
		client := &http.Client{}
		req, _ := http.NewRequest("POST", postURL, bytes.NewBuffer(reqBody))
		req.Header.Set("Authorization", "token "+token)
		_, err = client.Do(req)
		if err != nil {
			__antithesis_instrumentation__.Notify(40050)
			log.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40051)
		}
	}
}
