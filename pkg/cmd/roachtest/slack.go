package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/slack-go/slack"
)

var slackToken string

func makeSlackClient() *slack.Client {
	__antithesis_instrumentation__.Notify(44395)
	if slackToken == "" {
		__antithesis_instrumentation__.Notify(44397)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(44398)
	}
	__antithesis_instrumentation__.Notify(44396)
	return slack.New(slackToken)
}

func findChannel(client *slack.Client, name string, nextCursor string) (string, error) {
	__antithesis_instrumentation__.Notify(44399)
	if client != nil {
		__antithesis_instrumentation__.Notify(44401)
		channels, cursor, err := client.GetConversationsForUser(
			&slack.GetConversationsForUserParameters{Cursor: nextCursor},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(44404)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(44405)
		}
		__antithesis_instrumentation__.Notify(44402)
		for _, channel := range channels {
			__antithesis_instrumentation__.Notify(44406)
			if channel.Name == name {
				__antithesis_instrumentation__.Notify(44407)
				return channel.ID, nil
			} else {
				__antithesis_instrumentation__.Notify(44408)
			}
		}
		__antithesis_instrumentation__.Notify(44403)
		if cursor != "" {
			__antithesis_instrumentation__.Notify(44409)
			return findChannel(client, name, cursor)
		} else {
			__antithesis_instrumentation__.Notify(44410)
		}
	} else {
		__antithesis_instrumentation__.Notify(44411)
	}
	__antithesis_instrumentation__.Notify(44400)
	return "", fmt.Errorf("not found")
}

func sortTests(tests []*testImpl) {
	__antithesis_instrumentation__.Notify(44412)
	sort.Slice(tests, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(44413)
		return tests[i].Name() < tests[j].Name()
	})
}

func postSlackReport(pass, fail, skip map[*testImpl]struct{}) {
	__antithesis_instrumentation__.Notify(44414)
	client := makeSlackClient()
	if client == nil {
		__antithesis_instrumentation__.Notify(44421)
		return
	} else {
		__antithesis_instrumentation__.Notify(44422)
	}
	__antithesis_instrumentation__.Notify(44415)

	channel, _ := findChannel(client, "production", "")
	if channel == "" {
		__antithesis_instrumentation__.Notify(44423)
		return
	} else {
		__antithesis_instrumentation__.Notify(44424)
	}
	__antithesis_instrumentation__.Notify(44416)

	branch := "<unknown branch>"
	if b := os.Getenv("TC_BUILD_BRANCH"); b != "" {
		__antithesis_instrumentation__.Notify(44425)
		branch = b
	} else {
		__antithesis_instrumentation__.Notify(44426)
	}
	__antithesis_instrumentation__.Notify(44417)

	var prefix string
	switch {
	case cloud != "":
		__antithesis_instrumentation__.Notify(44427)
		prefix = strings.ToUpper(cloud)
	case local:
		__antithesis_instrumentation__.Notify(44428)
		prefix = "LOCAL"
	default:
		__antithesis_instrumentation__.Notify(44429)
		prefix = "GCE"
	}
	__antithesis_instrumentation__.Notify(44418)
	message := fmt.Sprintf("[%s] %s: %d passed, %d failed, %d skipped",
		prefix, branch, len(pass), len(fail), len(skip))

	var attachments []slack.Attachment
	{
		__antithesis_instrumentation__.Notify(44430)
		status := "good"
		if len(fail) > 0 {
			__antithesis_instrumentation__.Notify(44433)
			status = "warning"
		} else {
			__antithesis_instrumentation__.Notify(44434)
		}
		__antithesis_instrumentation__.Notify(44431)
		var link string
		if buildID := os.Getenv("TC_BUILD_ID"); buildID != "" {
			__antithesis_instrumentation__.Notify(44435)
			link = fmt.Sprintf("https://teamcity.cockroachdb.com/viewLog.html?"+
				"buildId=%s&buildTypeId=Cockroach_Nightlies_WorkloadNightly",
				buildID)
		} else {
			__antithesis_instrumentation__.Notify(44436)
		}
		__antithesis_instrumentation__.Notify(44432)
		attachments = append(attachments,
			slack.Attachment{
				Color:     status,
				Title:     message,
				TitleLink: link,
				Fallback:  message,
			})
	}
	__antithesis_instrumentation__.Notify(44419)

	data := []struct {
		tests map[*testImpl]struct{}
		title string
		color string
	}{
		{pass, "Successes", "good"},
		{fail, "Failures", "danger"},
		{skip, "Skipped", "warning"},
	}
	for _, d := range data {
		__antithesis_instrumentation__.Notify(44437)
		tests := make([]*testImpl, 0, len(d.tests))
		for t := range d.tests {
			__antithesis_instrumentation__.Notify(44440)
			tests = append(tests, t)
		}
		__antithesis_instrumentation__.Notify(44438)
		sortTests(tests)

		var buf bytes.Buffer
		for _, t := range tests {
			__antithesis_instrumentation__.Notify(44441)
			fmt.Fprintf(&buf, "%s\n", t.Name())
		}
		__antithesis_instrumentation__.Notify(44439)
		attachments = append(attachments,
			slack.Attachment{
				Color:    d.color,
				Title:    fmt.Sprintf("%s: %d", d.title, len(tests)),
				Text:     buf.String(),
				Fallback: message,
			})
	}
	__antithesis_instrumentation__.Notify(44420)

	if _, _, err := client.PostMessage(
		channel,
		slack.MsgOptionUsername("roachtest"),
		slack.MsgOptionAttachments(attachments...),
	); err != nil {
		__antithesis_instrumentation__.Notify(44442)
		fmt.Println("unable to post slack report: ", err)
	} else {
		__antithesis_instrumentation__.Notify(44443)
	}
}
