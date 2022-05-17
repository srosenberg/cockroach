package cloud

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachprod/config"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/roachprod/vm"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/slack-go/slack"
)

var errNoSlackClient = fmt.Errorf("no Slack client")

type status struct {
	good    []*Cluster
	warn    []*Cluster
	destroy []*Cluster
}

func (s *status) add(c *Cluster, now time.Time) {
	__antithesis_instrumentation__.Notify(180037)
	exp := c.ExpiresAt()
	if exp.After(now) {
		__antithesis_instrumentation__.Notify(180038)
		if exp.Before(now.Add(2 * time.Hour)) {
			__antithesis_instrumentation__.Notify(180039)
			s.warn = append(s.warn, c)
		} else {
			__antithesis_instrumentation__.Notify(180040)
			s.good = append(s.good, c)
		}
	} else {
		__antithesis_instrumentation__.Notify(180041)
		s.destroy = append(s.destroy, c)
	}
}

func (s *status) notificationHash() string {
	__antithesis_instrumentation__.Notify(180042)

	hash := fnv.New32a()

	for i, list := range [][]*Cluster{s.good, s.warn, s.destroy} {
		__antithesis_instrumentation__.Notify(180044)
		_, _ = hash.Write([]byte{byte(i)})

		var data []string
		for _, c := range list {
			__antithesis_instrumentation__.Notify(180046)

			data = append(data, fmt.Sprintf("%s %s", c.Name, c.ExpiresAt()))
		}
		__antithesis_instrumentation__.Notify(180045)

		sort.Strings(data)

		for _, d := range data {
			__antithesis_instrumentation__.Notify(180047)
			_, _ = hash.Write([]byte(d))
		}
	}
	__antithesis_instrumentation__.Notify(180043)

	bytes := hash.Sum(nil)
	return base64.StdEncoding.EncodeToString(bytes)
}

func makeSlackClient() *slack.Client {
	__antithesis_instrumentation__.Notify(180048)
	if config.SlackToken == "" {
		__antithesis_instrumentation__.Notify(180050)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(180051)
	}
	__antithesis_instrumentation__.Notify(180049)
	client := slack.New(config.SlackToken)

	return client
}

func findChannel(client *slack.Client, name string, nextCursor string) (string, error) {
	__antithesis_instrumentation__.Notify(180052)
	if client != nil {
		__antithesis_instrumentation__.Notify(180054)
		channels, cursor, err := client.GetConversationsForUser(
			&slack.GetConversationsForUserParameters{Cursor: nextCursor},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(180057)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(180058)
		}
		__antithesis_instrumentation__.Notify(180055)
		for _, channel := range channels {
			__antithesis_instrumentation__.Notify(180059)
			if channel.Name == name {
				__antithesis_instrumentation__.Notify(180060)
				return channel.ID, nil
			} else {
				__antithesis_instrumentation__.Notify(180061)
			}
		}
		__antithesis_instrumentation__.Notify(180056)
		if cursor != "" {
			__antithesis_instrumentation__.Notify(180062)
			return findChannel(client, name, cursor)
		} else {
			__antithesis_instrumentation__.Notify(180063)
		}
	} else {
		__antithesis_instrumentation__.Notify(180064)
	}
	__antithesis_instrumentation__.Notify(180053)
	return "", fmt.Errorf("not found")
}

func findUserChannel(client *slack.Client, email string) (string, error) {
	__antithesis_instrumentation__.Notify(180065)
	if client == nil {
		__antithesis_instrumentation__.Notify(180068)
		return "", errNoSlackClient
	} else {
		__antithesis_instrumentation__.Notify(180069)
	}
	__antithesis_instrumentation__.Notify(180066)
	u, err := client.GetUserByEmail(email)
	if err != nil {
		__antithesis_instrumentation__.Notify(180070)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(180071)
	}
	__antithesis_instrumentation__.Notify(180067)
	return u.ID, nil
}

func postStatus(
	l *logger.Logger, client *slack.Client, channel string, dryrun bool, s *status, badVMs vm.List,
) {
	__antithesis_instrumentation__.Notify(180072)
	if dryrun {
		__antithesis_instrumentation__.Notify(180081)
		tw := tabwriter.NewWriter(l.Stdout, 0, 8, 2, ' ', 0)
		for _, c := range s.good {
			__antithesis_instrumentation__.Notify(180085)
			fmt.Fprintf(tw, "good:\t%s\t%s\t(%s)\n", c.Name,
				c.GCAt().Format(time.Stamp),
				c.LifetimeRemaining().Round(time.Second))
		}
		__antithesis_instrumentation__.Notify(180082)
		for _, c := range s.warn {
			__antithesis_instrumentation__.Notify(180086)
			fmt.Fprintf(tw, "warn:\t%s\t%s\t(%s)\n", c.Name,
				c.GCAt().Format(time.Stamp),
				c.LifetimeRemaining().Round(time.Second))
		}
		__antithesis_instrumentation__.Notify(180083)
		for _, c := range s.destroy {
			__antithesis_instrumentation__.Notify(180087)
			fmt.Fprintf(tw, "destroy:\t%s\t%s\t(%s)\n", c.Name,
				c.GCAt().Format(time.Stamp),
				c.LifetimeRemaining().Round(time.Second))
		}
		__antithesis_instrumentation__.Notify(180084)
		_ = tw.Flush()
	} else {
		__antithesis_instrumentation__.Notify(180088)
	}
	__antithesis_instrumentation__.Notify(180073)

	if client == nil || func() bool {
		__antithesis_instrumentation__.Notify(180089)
		return channel == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(180090)
		return
	} else {
		__antithesis_instrumentation__.Notify(180091)
	}
	__antithesis_instrumentation__.Notify(180074)

	if len(badVMs) == 0 {
		__antithesis_instrumentation__.Notify(180092)
		send, err := shouldSend(channel, s)
		if err != nil {
			__antithesis_instrumentation__.Notify(180094)
			log.Infof(context.Background(), "unable to deduplicate notification: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(180095)
		}
		__antithesis_instrumentation__.Notify(180093)
		if !send {
			__antithesis_instrumentation__.Notify(180096)
			return
		} else {
			__antithesis_instrumentation__.Notify(180097)
		}
	} else {
		__antithesis_instrumentation__.Notify(180098)
	}
	__antithesis_instrumentation__.Notify(180075)

	makeStatusFields := func(clusters []*Cluster) []slack.AttachmentField {
		__antithesis_instrumentation__.Notify(180099)
		var names []string
		var expirations []string
		for _, c := range clusters {
			__antithesis_instrumentation__.Notify(180101)
			names = append(names, c.Name)
			expirations = append(expirations,
				fmt.Sprintf("<!date^%[1]d^{date_short_pretty} {time}|%[2]s>",
					c.GCAt().Unix(),
					c.LifetimeRemaining().Round(time.Second)))
		}
		__antithesis_instrumentation__.Notify(180100)
		return []slack.AttachmentField{
			{
				Title: "name",
				Value: strings.Join(names, "\n"),
				Short: true,
			},
			{
				Title: "expiration",
				Value: strings.Join(expirations, "\n"),
				Short: true,
			},
		}
	}
	__antithesis_instrumentation__.Notify(180076)

	var attachments []slack.Attachment
	fallback := fmt.Sprintf("clusters: %d live, %d expired, %d destroyed",
		len(s.good), len(s.warn), len(s.destroy))
	if len(s.good) > 0 {
		__antithesis_instrumentation__.Notify(180102)
		attachments = append(attachments,
			slack.Attachment{
				Color:    "good",
				Title:    "Live Clusters",
				Fallback: fallback,
				Fields:   makeStatusFields(s.good),
			})
	} else {
		__antithesis_instrumentation__.Notify(180103)
	}
	__antithesis_instrumentation__.Notify(180077)
	if len(s.warn) > 0 {
		__antithesis_instrumentation__.Notify(180104)
		attachments = append(attachments,
			slack.Attachment{
				Color:    "warning",
				Title:    "Expiring Clusters",
				Fallback: fallback,
				Fields:   makeStatusFields(s.warn),
			})
	} else {
		__antithesis_instrumentation__.Notify(180105)
	}
	__antithesis_instrumentation__.Notify(180078)
	if len(s.destroy) > 0 {
		__antithesis_instrumentation__.Notify(180106)
		attachments = append(attachments,
			slack.Attachment{
				Color:    "danger",
				Title:    "Destroyed Clusters",
				Fallback: fallback,
				Fields:   makeStatusFields(s.destroy),
			})
	} else {
		__antithesis_instrumentation__.Notify(180107)
	}
	__antithesis_instrumentation__.Notify(180079)
	if len(badVMs) > 0 {
		__antithesis_instrumentation__.Notify(180108)
		var names []string
		for _, vm := range badVMs {
			__antithesis_instrumentation__.Notify(180110)
			names = append(names, vm.Name)
		}
		__antithesis_instrumentation__.Notify(180109)
		sort.Strings(names)
		attachments = append(attachments,
			slack.Attachment{
				Color: "danger",
				Title: "Bad VMs",
				Text:  strings.Join(names, "\n"),
			})
	} else {
		__antithesis_instrumentation__.Notify(180111)
	}
	__antithesis_instrumentation__.Notify(180080)
	_, _, err := client.PostMessage(
		channel,
		slack.MsgOptionUsername("roachprod"),
		slack.MsgOptionAttachments(attachments...),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(180112)
		log.Infof(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(180113)
	}
}

func postError(client *slack.Client, channel string, err error) {
	__antithesis_instrumentation__.Notify(180114)
	log.Infof(context.Background(), "%v", err)
	if client == nil || func() bool {
		__antithesis_instrumentation__.Notify(180116)
		return channel == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(180117)
		return
	} else {
		__antithesis_instrumentation__.Notify(180118)
	}
	__antithesis_instrumentation__.Notify(180115)

	_, _, err = client.PostMessage(
		channel,
		slack.MsgOptionUsername("roachprod"),
		slack.MsgOptionText(fmt.Sprintf("`%s`", err), false),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(180119)
		log.Infof(context.Background(), "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(180120)
	}
}

func shouldSend(channel string, status *status) (bool, error) {
	__antithesis_instrumentation__.Notify(180121)
	hashDir := os.ExpandEnv(filepath.Join("${HOME}", ".roachprod", "slack"))
	if err := os.MkdirAll(hashDir, 0755); err != nil {
		__antithesis_instrumentation__.Notify(180125)
		return true, err
	} else {
		__antithesis_instrumentation__.Notify(180126)
	}
	__antithesis_instrumentation__.Notify(180122)
	hashPath := os.ExpandEnv(filepath.Join(hashDir, "notification-"+channel))
	fileBytes, err := ioutil.ReadFile(hashPath)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(180127)
		return !oserror.IsNotExist(err) == true
	}() == true {
		__antithesis_instrumentation__.Notify(180128)
		return true, err
	} else {
		__antithesis_instrumentation__.Notify(180129)
	}
	__antithesis_instrumentation__.Notify(180123)
	oldHash := string(fileBytes)
	newHash := status.notificationHash()

	if newHash == oldHash {
		__antithesis_instrumentation__.Notify(180130)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(180131)
	}
	__antithesis_instrumentation__.Notify(180124)

	return true, ioutil.WriteFile(hashPath, []byte(newHash), 0644)
}

func GCClusters(l *logger.Logger, cloud *Cloud, dryrun bool) error {
	__antithesis_instrumentation__.Notify(180132)
	now := timeutil.Now()

	var names []string
	for name := range cloud.Clusters {
		__antithesis_instrumentation__.Notify(180138)
		if !config.IsLocalClusterName(name) {
			__antithesis_instrumentation__.Notify(180139)
			names = append(names, name)
		} else {
			__antithesis_instrumentation__.Notify(180140)
		}
	}
	__antithesis_instrumentation__.Notify(180133)
	sort.Strings(names)

	var s status
	users := make(map[string]*status)
	for _, name := range names {
		__antithesis_instrumentation__.Notify(180141)
		c := cloud.Clusters[name]
		u := users[c.User]
		if u == nil {
			__antithesis_instrumentation__.Notify(180143)
			u = &status{}
			users[c.User] = u
		} else {
			__antithesis_instrumentation__.Notify(180144)
		}
		__antithesis_instrumentation__.Notify(180142)
		s.add(c, now)
		u.add(c, now)
	}
	__antithesis_instrumentation__.Notify(180134)

	var badVMs vm.List
	for _, vm := range cloud.BadInstances {
		__antithesis_instrumentation__.Notify(180145)

		if now.Sub(vm.CreatedAt) >= time.Hour {
			__antithesis_instrumentation__.Notify(180146)
			badVMs = append(badVMs, vm)
		} else {
			__antithesis_instrumentation__.Notify(180147)
		}
	}
	__antithesis_instrumentation__.Notify(180135)

	client := makeSlackClient()
	channel, _ := findChannel(client, "roachprod-status", "")
	postStatus(l, client, channel, dryrun, &s, badVMs)

	for user, status := range users {
		__antithesis_instrumentation__.Notify(180148)
		if len(status.warn) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(180149)
			return len(status.destroy) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(180150)
			userChannel, err := findUserChannel(client, user+config.EmailDomain)
			if err == nil {
				__antithesis_instrumentation__.Notify(180151)
				postStatus(l, client, userChannel, dryrun, status, nil)
			} else {
				__antithesis_instrumentation__.Notify(180152)
				if !errors.Is(err, errNoSlackClient) {
					__antithesis_instrumentation__.Notify(180153)
					log.Infof(context.Background(), "could not deliver Slack DM to %s: %v", user+config.EmailDomain, err)
				} else {
					__antithesis_instrumentation__.Notify(180154)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(180155)
		}
	}
	__antithesis_instrumentation__.Notify(180136)

	if !dryrun {
		__antithesis_instrumentation__.Notify(180156)
		if len(badVMs) > 0 {
			__antithesis_instrumentation__.Notify(180158)

			err := vm.FanOut(badVMs, func(p vm.Provider, vms vm.List) error {
				__antithesis_instrumentation__.Notify(180160)
				return p.Delete(vms)
			})
			__antithesis_instrumentation__.Notify(180159)
			if err != nil {
				__antithesis_instrumentation__.Notify(180161)
				postError(client, channel, err)
			} else {
				__antithesis_instrumentation__.Notify(180162)
			}
		} else {
			__antithesis_instrumentation__.Notify(180163)
		}
		__antithesis_instrumentation__.Notify(180157)

		for _, c := range s.destroy {
			__antithesis_instrumentation__.Notify(180164)
			if err := DestroyCluster(c); err != nil {
				__antithesis_instrumentation__.Notify(180165)
				postError(client, channel, err)
			} else {
				__antithesis_instrumentation__.Notify(180166)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(180167)
	}
	__antithesis_instrumentation__.Notify(180137)
	return nil
}
