package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/ghemawat/stream"
	"github.com/slack-go/slack"
)

var slackToken = flag.String("slack-token", "", "Slack bot token")
var slackChannel = flag.String("slack-channel", "test-infra-ops", "Slack channel")

type skippedTest struct {
	file string
	test string
}

func dirCmd(dir string, name string, args ...string) stream.Filter {
	__antithesis_instrumentation__.Notify(52807)
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	switch {
	case err == nil:
		__antithesis_instrumentation__.Notify(52809)
	case errors.HasType(err, (*exec.ExitError)(nil)):
		__antithesis_instrumentation__.Notify(52810)

	default:
		__antithesis_instrumentation__.Notify(52811)
		log.Fatal(err)
	}
	__antithesis_instrumentation__.Notify(52808)
	return stream.ReadLines(bytes.NewReader(out))
}

func makeSlackClient() *slack.Client {
	__antithesis_instrumentation__.Notify(52812)
	if *slackToken == "" {
		__antithesis_instrumentation__.Notify(52814)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(52815)
	}
	__antithesis_instrumentation__.Notify(52813)
	return slack.New(*slackToken)
}

func findChannel(client *slack.Client, name string, nextCursor string) (string, error) {
	__antithesis_instrumentation__.Notify(52816)
	if client != nil {
		__antithesis_instrumentation__.Notify(52818)
		channels, cursor, err := client.GetConversationsForUser(
			&slack.GetConversationsForUserParameters{Cursor: nextCursor},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(52821)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(52822)
		}
		__antithesis_instrumentation__.Notify(52819)
		for _, channel := range channels {
			__antithesis_instrumentation__.Notify(52823)
			if channel.Name == name {
				__antithesis_instrumentation__.Notify(52824)
				return channel.ID, nil
			} else {
				__antithesis_instrumentation__.Notify(52825)
			}
		}
		__antithesis_instrumentation__.Notify(52820)
		if cursor != "" {
			__antithesis_instrumentation__.Notify(52826)
			return findChannel(client, name, cursor)
		} else {
			__antithesis_instrumentation__.Notify(52827)
		}
	} else {
		__antithesis_instrumentation__.Notify(52828)
	}
	__antithesis_instrumentation__.Notify(52817)
	return "", fmt.Errorf("not found")
}

func postReport(skipped []skippedTest) {
	__antithesis_instrumentation__.Notify(52829)
	client := makeSlackClient()
	if client == nil {
		__antithesis_instrumentation__.Notify(52837)
		fmt.Printf("no slack client\n")
		return
	} else {
		__antithesis_instrumentation__.Notify(52838)
	}
	__antithesis_instrumentation__.Notify(52830)

	channel, _ := findChannel(client, *slackChannel, "")
	if channel == "" {
		__antithesis_instrumentation__.Notify(52839)
		fmt.Printf("unable to find slack channel: %q\n", *slackChannel)
		return
	} else {
		__antithesis_instrumentation__.Notify(52840)
	}
	__antithesis_instrumentation__.Notify(52831)

	status := "good"
	switch n := len(skipped); {
	case n >= 100:
		__antithesis_instrumentation__.Notify(52841)
		status = "danger"
	case n > 10:
		__antithesis_instrumentation__.Notify(52842)
		status = "warning"
	default:
		__antithesis_instrumentation__.Notify(52843)
	}
	__antithesis_instrumentation__.Notify(52832)

	message := fmt.Sprintf("%d skipped tests", len(skipped))
	fmt.Println(message)

	var attachments []slack.Attachment
	attachments = append(attachments,
		slack.Attachment{
			Color:    status,
			Title:    message,
			Fallback: message,
		})

	fileMap := make(map[string]int)
	for i := range skipped {
		__antithesis_instrumentation__.Notify(52844)
		fileMap[skipped[i].file]++
	}
	__antithesis_instrumentation__.Notify(52833)
	files := make([]string, 0, len(fileMap))
	for file := range fileMap {
		__antithesis_instrumentation__.Notify(52845)
		files = append(files, file)
	}
	__antithesis_instrumentation__.Notify(52834)
	sort.Strings(files)

	var buf bytes.Buffer
	for _, file := range files {
		__antithesis_instrumentation__.Notify(52846)
		fmt.Fprintf(&buf, "%3d %s\n", fileMap[file], file)
	}
	__antithesis_instrumentation__.Notify(52835)
	fmt.Print(buf.String())

	attachments = append(attachments,
		slack.Attachment{
			Color: status,
			Text:  fmt.Sprintf("```\n%s```\n", buf.String()),
		})

	if _, _, err := client.PostMessage(
		channel,
		slack.MsgOptionUsername("Craig Cockroach"),
		slack.MsgOptionAttachments(attachments...),
	); err != nil {
		__antithesis_instrumentation__.Notify(52847)
		fmt.Printf("unable to post slack report: %v\n", err)
		return
	} else {
		__antithesis_instrumentation__.Notify(52848)
	}
	__antithesis_instrumentation__.Notify(52836)

	fmt.Printf("posted slack report\n")
}

func main() {
	__antithesis_instrumentation__.Notify(52849)
	flag.Parse()

	const root = "github.com/cockroachdb/cockroach"

	crdb, err := build.Import(root, "", build.FindOnly)
	if err != nil {
		__antithesis_instrumentation__.Notify(52852)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52853)
	}
	__antithesis_instrumentation__.Notify(52850)
	pkgDir := filepath.Join(crdb.Dir, "pkg")

	testnameRE := regexp.MustCompile(`([^:]+):func (Test[^(]*).*`)

	filter := stream.Sequence(
		dirCmd(pkgDir, "git", "grep", "-E", `^func (Test|Benchmark)|\.Skipf?\(`),

		stream.GrepNot(`short|PKG specified`),
	)

	var skipped []skippedTest
	var lastTest string
	if err := stream.ForEach(filter, func(s string) {
		__antithesis_instrumentation__.Notify(52854)
		switch {
		case strings.Contains(s, ":func "):
			__antithesis_instrumentation__.Notify(52855)
			lastTest = s
		case strings.Contains(s, ".Skip"):
			__antithesis_instrumentation__.Notify(52856)
			m := testnameRE.FindStringSubmatch(lastTest)
			if m != nil {
				__antithesis_instrumentation__.Notify(52858)
				file, test := m[1], m[2]
				skipped = append(skipped, skippedTest{
					file: file,
					test: test,
				})

				lastTest = ""
			} else {
				__antithesis_instrumentation__.Notify(52859)
			}
		default:
			__antithesis_instrumentation__.Notify(52857)
		}
	}); err != nil {
		__antithesis_instrumentation__.Notify(52860)
		log.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(52861)
	}
	__antithesis_instrumentation__.Notify(52851)

	postReport(skipped)
}
