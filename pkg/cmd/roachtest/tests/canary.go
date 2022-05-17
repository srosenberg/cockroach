package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

type blocklist map[string]string

type blocklistForVersion struct {
	versionPrefix  string
	blocklistname  string
	blocklist      blocklist
	ignorelistname string
	ignorelist     blocklist
}

type blocklistsForVersion []blocklistForVersion

func (b blocklistsForVersion) getLists(version string) (string, blocklist, string, blocklist) {
	__antithesis_instrumentation__.Notify(45915)
	for _, info := range b {
		__antithesis_instrumentation__.Notify(45917)
		if strings.HasPrefix(version, info.versionPrefix) {
			__antithesis_instrumentation__.Notify(45918)
			return info.blocklistname, info.blocklist, info.ignorelistname, info.ignorelist
		} else {
			__antithesis_instrumentation__.Notify(45919)
		}
	}
	__antithesis_instrumentation__.Notify(45916)
	return "", nil, "", nil
}

func fetchCockroachVersion(
	ctx context.Context, l *logger.Logger, c cluster.Cluster, nodeIndex int,
) (string, error) {
	__antithesis_instrumentation__.Notify(45920)
	db, err := c.ConnE(ctx, l, nodeIndex)
	if err != nil {
		__antithesis_instrumentation__.Notify(45923)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(45924)
	}
	__antithesis_instrumentation__.Notify(45921)
	defer db.Close()
	var version string
	if err := db.QueryRowContext(ctx,
		`SELECT value FROM crdb_internal.node_build_info where field = 'Version'`,
	).Scan(&version); err != nil {
		__antithesis_instrumentation__.Notify(45925)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(45926)
	}
	__antithesis_instrumentation__.Notify(45922)
	return version, nil
}

func maybeAddGithubLink(issue string) string {
	__antithesis_instrumentation__.Notify(45927)
	if len(issue) == 0 {
		__antithesis_instrumentation__.Notify(45930)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(45931)
	}
	__antithesis_instrumentation__.Notify(45928)
	issueNum, err := strconv.Atoi(issue)
	if err != nil {
		__antithesis_instrumentation__.Notify(45932)
		return issue
	} else {
		__antithesis_instrumentation__.Notify(45933)
	}
	__antithesis_instrumentation__.Notify(45929)
	return fmt.Sprintf("https://github.com/cockroachdb/cockroach/issues/%d", issueNum)
}

var canaryRetryOptions = retry.Options{
	InitialBackoff: 10 * time.Second,
	Multiplier:     2,
	MaxBackoff:     5 * time.Minute,
	MaxRetries:     10,
}

func repeatRunE(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	node option.NodeListOption,
	operation string,
	args ...string,
) error {
	__antithesis_instrumentation__.Notify(45934)
	var lastError error
	for attempt, r := 0, retry.StartWithCtx(ctx, canaryRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(45936)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(45940)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(45941)
		}
		__antithesis_instrumentation__.Notify(45937)
		if t.Failed() {
			__antithesis_instrumentation__.Notify(45942)
			return fmt.Errorf("test has failed")
		} else {
			__antithesis_instrumentation__.Notify(45943)
		}
		__antithesis_instrumentation__.Notify(45938)
		attempt++
		t.L().Printf("attempt %d - %s", attempt, operation)
		lastError = c.RunE(ctx, node, args...)
		if lastError != nil {
			__antithesis_instrumentation__.Notify(45944)
			t.L().Printf("error - retrying: %s", lastError)
			continue
		} else {
			__antithesis_instrumentation__.Notify(45945)
		}
		__antithesis_instrumentation__.Notify(45939)
		return nil
	}
	__antithesis_instrumentation__.Notify(45935)
	return errors.Wrapf(lastError, "all attempts failed for %s", operation)
}

func repeatRunWithDetailsSingleNode(
	ctx context.Context,
	c cluster.Cluster,
	t test.Test,
	node option.NodeListOption,
	operation string,
	args ...string,
) (install.RunResultDetails, error) {
	__antithesis_instrumentation__.Notify(45946)
	var (
		lastResult install.RunResultDetails
		lastError  error
	)
	for attempt, r := 0, retry.StartWithCtx(ctx, canaryRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(45948)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(45952)
			return lastResult, ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(45953)
		}
		__antithesis_instrumentation__.Notify(45949)
		if t.Failed() {
			__antithesis_instrumentation__.Notify(45954)
			return lastResult, fmt.Errorf("test has failed")
		} else {
			__antithesis_instrumentation__.Notify(45955)
		}
		__antithesis_instrumentation__.Notify(45950)
		attempt++
		t.L().Printf("attempt %d - %s", attempt, operation)
		lastResult, lastError = c.RunWithDetailsSingleNode(ctx, t.L(), node, args...)
		if lastError != nil {
			__antithesis_instrumentation__.Notify(45956)
			t.L().Printf("error - retrying: %s", lastError)
			continue
		} else {
			__antithesis_instrumentation__.Notify(45957)
		}
		__antithesis_instrumentation__.Notify(45951)
		return lastResult, nil
	}
	__antithesis_instrumentation__.Notify(45947)
	return lastResult, errors.Wrapf(lastError, "all attempts failed for %s", operation)
}

func repeatGitCloneE(
	ctx context.Context,
	t test.Test,
	c cluster.Cluster,
	src, dest, branch string,
	node option.NodeListOption,
) error {
	__antithesis_instrumentation__.Notify(45958)
	var lastError error
	for attempt, r := 0, retry.StartWithCtx(ctx, canaryRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(45960)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(45964)
			return ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(45965)
		}
		__antithesis_instrumentation__.Notify(45961)
		if t.Failed() {
			__antithesis_instrumentation__.Notify(45966)
			return fmt.Errorf("test has failed")
		} else {
			__antithesis_instrumentation__.Notify(45967)
		}
		__antithesis_instrumentation__.Notify(45962)
		attempt++
		t.L().Printf("attempt %d - clone %s", attempt, src)
		lastError = c.GitClone(ctx, t.L(), src, dest, branch, node)
		if lastError != nil {
			__antithesis_instrumentation__.Notify(45968)
			t.L().Printf("error - retrying: %s", lastError)
			continue
		} else {
			__antithesis_instrumentation__.Notify(45969)
		}
		__antithesis_instrumentation__.Notify(45963)
		return nil
	}
	__antithesis_instrumentation__.Notify(45959)
	return errors.Wrapf(lastError, "could not clone %s", src)
}

func repeatGetLatestTag(
	ctx context.Context, t test.Test, user string, repo string, releaseRegex *regexp.Regexp,
) (string, error) {
	__antithesis_instrumentation__.Notify(45970)
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/tags", user, repo)
	httpClient := &http.Client{Timeout: 10 * time.Second}
	type Tag struct {
		Name string
	}
	type releaseTag struct {
		tag      string
		major    int
		minor    int
		point    int
		subpoint int
	}
	type Tags []Tag
	atoiOrZero := func(groups map[string]string, name string) int {
		__antithesis_instrumentation__.Notify(45973)
		value, ok := groups[name]
		if !ok {
			__antithesis_instrumentation__.Notify(45976)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(45977)
		}
		__antithesis_instrumentation__.Notify(45974)
		i, err := strconv.Atoi(value)
		if err != nil {
			__antithesis_instrumentation__.Notify(45978)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(45979)
		}
		__antithesis_instrumentation__.Notify(45975)
		return i
	}
	__antithesis_instrumentation__.Notify(45971)
	var lastError error
	for attempt, r := 0, retry.StartWithCtx(ctx, canaryRetryOptions); r.Next(); {
		__antithesis_instrumentation__.Notify(45980)
		if ctx.Err() != nil {
			__antithesis_instrumentation__.Notify(45989)
			return "", ctx.Err()
		} else {
			__antithesis_instrumentation__.Notify(45990)
		}
		__antithesis_instrumentation__.Notify(45981)
		if t.Failed() {
			__antithesis_instrumentation__.Notify(45991)
			return "", fmt.Errorf("test has failed")
		} else {
			__antithesis_instrumentation__.Notify(45992)
		}
		__antithesis_instrumentation__.Notify(45982)
		attempt++

		t.L().Printf("attempt %d - fetching %s", attempt, url)
		var resp *http.Response
		resp, lastError = httpClient.Get(url)
		if lastError != nil {
			__antithesis_instrumentation__.Notify(45993)
			t.L().Printf("error fetching - retrying: %s", lastError)
			continue
		} else {
			__antithesis_instrumentation__.Notify(45994)
		}
		__antithesis_instrumentation__.Notify(45983)
		defer resp.Body.Close()

		var tags Tags
		lastError = json.NewDecoder(resp.Body).Decode(&tags)
		if lastError != nil {
			__antithesis_instrumentation__.Notify(45995)
			t.L().Printf("error decoding - retrying: %s", lastError)
			continue
		} else {
			__antithesis_instrumentation__.Notify(45996)
		}
		__antithesis_instrumentation__.Notify(45984)
		if len(tags) == 0 {
			__antithesis_instrumentation__.Notify(45997)
			return "", fmt.Errorf("no tags found at %s", url)
		} else {
			__antithesis_instrumentation__.Notify(45998)
		}
		__antithesis_instrumentation__.Notify(45985)
		var releaseTags []releaseTag
		for _, t := range tags {
			__antithesis_instrumentation__.Notify(45999)
			match := releaseRegex.FindStringSubmatch(t.Name)
			if match == nil {
				__antithesis_instrumentation__.Notify(46003)
				continue
			} else {
				__antithesis_instrumentation__.Notify(46004)
			}
			__antithesis_instrumentation__.Notify(46000)
			groups := map[string]string{}
			for i, name := range match {
				__antithesis_instrumentation__.Notify(46005)
				groups[releaseRegex.SubexpNames()[i]] = name
			}
			__antithesis_instrumentation__.Notify(46001)
			if _, ok := groups["major"]; !ok {
				__antithesis_instrumentation__.Notify(46006)
				continue
			} else {
				__antithesis_instrumentation__.Notify(46007)
			}
			__antithesis_instrumentation__.Notify(46002)
			releaseTags = append(releaseTags, releaseTag{
				tag:      t.Name,
				major:    atoiOrZero(groups, "major"),
				minor:    atoiOrZero(groups, "minor"),
				point:    atoiOrZero(groups, "point"),
				subpoint: atoiOrZero(groups, "subpoint"),
			})
		}
		__antithesis_instrumentation__.Notify(45986)
		if len(releaseTags) == 0 {
			__antithesis_instrumentation__.Notify(46008)
			return "", fmt.Errorf("no tags match the given regex")
		} else {
			__antithesis_instrumentation__.Notify(46009)
		}
		__antithesis_instrumentation__.Notify(45987)
		sort.SliceStable(releaseTags, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(46010)
			return releaseTags[i].major < releaseTags[j].major || func() bool {
				__antithesis_instrumentation__.Notify(46011)
				return releaseTags[i].minor < releaseTags[j].minor == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(46012)
				return releaseTags[i].point < releaseTags[j].point == true
			}() == true || func() bool {
				__antithesis_instrumentation__.Notify(46013)
				return releaseTags[i].subpoint < releaseTags[j].subpoint == true
			}() == true
		})
		__antithesis_instrumentation__.Notify(45988)

		return releaseTags[len(releaseTags)-1].tag, nil
	}
	__antithesis_instrumentation__.Notify(45972)
	return "", errors.Wrapf(lastError, "could not get tags from %s", url)
}

func gitCloneWithRecurseSubmodules(
	ctx context.Context,
	c cluster.Cluster,
	l *logger.Logger,
	src, dest, branch string,
	node option.NodeListOption,
) error {
	__antithesis_instrumentation__.Notify(46014)
	cmd := []string{"bash", "-e", "-c", fmt.Sprintf(`'
		if ! test -d %[1]s; then
	  		git clone --recurse-submodules -b %[2]s --depth 1 %[3]s %[1]s
		else
	  		cd %[1]s
	  		git fetch origin
	  		git checkout origin/%[2]s
		fi
	'`, dest, branch, src),
	}
	return errors.Wrap(c.RunE(ctx, node, cmd...), "gitCloneWithRecurseSubmodules")
}
