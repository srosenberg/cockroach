package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/internal/team"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
)

var loadTeams = func() (team.Map, error) {
	__antithesis_instrumentation__.Notify(44708)
	return team.DefaultLoadTeams()
}

func ownerToAlias(o registry.Owner) team.Alias {
	__antithesis_instrumentation__.Notify(44709)
	return team.Alias(fmt.Sprintf("cockroachdb/%s", o))
}

type testRegistryImpl struct {
	m            map[string]*registry.TestSpec
	cloud        string
	instanceType string
	zones        string
	preferSSD    bool

	buildVersion version.Version
}

func makeTestRegistry(
	cloud string, instanceType string, zones string, preferSSD bool,
) (testRegistryImpl, error) {
	__antithesis_instrumentation__.Notify(44710)
	r := testRegistryImpl{
		cloud:        cloud,
		instanceType: instanceType,
		zones:        zones,
		preferSSD:    preferSSD,
		m:            make(map[string]*registry.TestSpec),
	}
	v := buildTag
	if v == "" {
		__antithesis_instrumentation__.Notify(44713)
		var err error
		v, err = loadBuildVersion()
		if err != nil {
			__antithesis_instrumentation__.Notify(44714)
			return testRegistryImpl{}, err
		} else {
			__antithesis_instrumentation__.Notify(44715)
		}
	} else {
		__antithesis_instrumentation__.Notify(44716)
	}
	__antithesis_instrumentation__.Notify(44711)
	if err := r.setBuildVersion(v); err != nil {
		__antithesis_instrumentation__.Notify(44717)
		return testRegistryImpl{}, err
	} else {
		__antithesis_instrumentation__.Notify(44718)
	}
	__antithesis_instrumentation__.Notify(44712)
	return r, nil
}

func (r *testRegistryImpl) Add(spec registry.TestSpec) {
	__antithesis_instrumentation__.Notify(44719)
	if _, ok := r.m[spec.Name]; ok {
		__antithesis_instrumentation__.Notify(44722)
		fmt.Fprintf(os.Stderr, "test %s already registered\n", spec.Name)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(44723)
	}
	__antithesis_instrumentation__.Notify(44720)
	if err := r.prepareSpec(&spec); err != nil {
		__antithesis_instrumentation__.Notify(44724)
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(44725)
	}
	__antithesis_instrumentation__.Notify(44721)
	r.m[spec.Name] = &spec
}

func (r *testRegistryImpl) MakeClusterSpec(nodeCount int, opts ...spec.Option) spec.ClusterSpec {
	__antithesis_instrumentation__.Notify(44726)

	var finalOpts []spec.Option
	if r.preferSSD {
		__antithesis_instrumentation__.Notify(44729)
		finalOpts = append(finalOpts, spec.PreferSSD())
	} else {
		__antithesis_instrumentation__.Notify(44730)
	}
	__antithesis_instrumentation__.Notify(44727)
	if r.zones != "" {
		__antithesis_instrumentation__.Notify(44731)
		finalOpts = append(finalOpts, spec.Zones(r.zones))
	} else {
		__antithesis_instrumentation__.Notify(44732)
	}
	__antithesis_instrumentation__.Notify(44728)
	finalOpts = append(finalOpts, opts...)
	return spec.MakeClusterSpec(r.cloud, r.instanceType, nodeCount, finalOpts...)
}

const testNameRE = "^[a-zA-Z0-9-_=/,]+$"

func (r *testRegistryImpl) prepareSpec(spec *registry.TestSpec) error {
	__antithesis_instrumentation__.Notify(44733)
	if matched, err := regexp.MatchString(testNameRE, spec.Name); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(44742)
		return !matched == true
	}() == true {
		__antithesis_instrumentation__.Notify(44743)
		return fmt.Errorf("%s: Name must match this regexp: %s", spec.Name, testNameRE)
	} else {
		__antithesis_instrumentation__.Notify(44744)
	}
	__antithesis_instrumentation__.Notify(44734)

	if spec.Run == nil {
		__antithesis_instrumentation__.Notify(44745)
		return fmt.Errorf("%s: must specify Run", spec.Name)
	} else {
		__antithesis_instrumentation__.Notify(44746)
	}
	__antithesis_instrumentation__.Notify(44735)

	if spec.Cluster.ReusePolicy == nil {
		__antithesis_instrumentation__.Notify(44747)
		return fmt.Errorf("%s: must specify a ClusterReusePolicy", spec.Name)
	} else {
		__antithesis_instrumentation__.Notify(44748)
	}
	__antithesis_instrumentation__.Notify(44736)

	if spec.Owner == `` {
		__antithesis_instrumentation__.Notify(44749)
		return fmt.Errorf(`%s: unspecified owner`, spec.Name)
	} else {
		__antithesis_instrumentation__.Notify(44750)
	}
	__antithesis_instrumentation__.Notify(44737)
	teams, err := loadTeams()
	if err != nil {
		__antithesis_instrumentation__.Notify(44751)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44752)
	}
	__antithesis_instrumentation__.Notify(44738)
	if _, ok := teams[ownerToAlias(spec.Owner)]; !ok {
		__antithesis_instrumentation__.Notify(44753)
		return fmt.Errorf(`%s: unknown owner [%s]`, spec.Name, spec.Owner)
	} else {
		__antithesis_instrumentation__.Notify(44754)
	}
	__antithesis_instrumentation__.Notify(44739)
	if len(spec.Tags) == 0 {
		__antithesis_instrumentation__.Notify(44755)
		spec.Tags = []string{registry.DefaultTag}
	} else {
		__antithesis_instrumentation__.Notify(44756)
	}
	__antithesis_instrumentation__.Notify(44740)
	spec.Tags = append(spec.Tags, "owner-"+string(spec.Owner))

	const maxTimeout = 18 * time.Hour
	if spec.Timeout > maxTimeout {
		__antithesis_instrumentation__.Notify(44757)
		var weekly bool
		for _, tag := range spec.Tags {
			__antithesis_instrumentation__.Notify(44759)
			if tag == "weekly" {
				__antithesis_instrumentation__.Notify(44760)
				weekly = true
			} else {
				__antithesis_instrumentation__.Notify(44761)
			}
		}
		__antithesis_instrumentation__.Notify(44758)
		if !weekly {
			__antithesis_instrumentation__.Notify(44762)
			return fmt.Errorf(
				"%s: timeout %s exceeds the maximum allowed of %s", spec.Name, spec.Timeout, maxTimeout,
			)
		} else {
			__antithesis_instrumentation__.Notify(44763)
		}
	} else {
		__antithesis_instrumentation__.Notify(44764)
	}
	__antithesis_instrumentation__.Notify(44741)

	return nil
}

func (r testRegistryImpl) GetTests(
	ctx context.Context, filter *registry.TestFilter,
) []registry.TestSpec {
	__antithesis_instrumentation__.Notify(44765)
	var tests []registry.TestSpec
	for _, t := range r.m {
		__antithesis_instrumentation__.Notify(44768)
		if !t.MatchOrSkip(filter) {
			__antithesis_instrumentation__.Notify(44770)
			continue
		} else {
			__antithesis_instrumentation__.Notify(44771)
		}
		__antithesis_instrumentation__.Notify(44769)
		tests = append(tests, *t)
	}
	__antithesis_instrumentation__.Notify(44766)
	sort.Slice(tests, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(44772)
		return tests[i].Name < tests[j].Name
	})
	__antithesis_instrumentation__.Notify(44767)
	return tests
}

func (r testRegistryImpl) List(ctx context.Context, filters []string) []registry.TestSpec {
	__antithesis_instrumentation__.Notify(44773)
	filter := registry.NewTestFilter(filters)
	tests := r.GetTests(ctx, filter)
	sort.Slice(tests, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(44775)
		return tests[i].Name < tests[j].Name
	})
	__antithesis_instrumentation__.Notify(44774)
	return tests
}

func (r *testRegistryImpl) setBuildVersion(buildTag string) error {
	__antithesis_instrumentation__.Notify(44776)
	v, err := version.Parse(buildTag)
	if err != nil {
		__antithesis_instrumentation__.Notify(44778)
		return err
	} else {
		__antithesis_instrumentation__.Notify(44779)
	}
	__antithesis_instrumentation__.Notify(44777)
	r.buildVersion = *v
	return err
}

func loadBuildVersion() (string, error) {
	__antithesis_instrumentation__.Notify(44780)
	cmd := exec.Command("git", "describe", "--abbrev=0", "--tags", "--match=v[0-9]*")
	out, err := cmd.CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(44782)
		return "", errors.Wrapf(
			err, "failed to get version tag from git. Are you running in the "+
				"cockroach repo directory? err=%s, out=%s",
			err, out)
	} else {
		__antithesis_instrumentation__.Notify(44783)
	}
	__antithesis_instrumentation__.Notify(44781)
	return strings.TrimSpace(string(out)), nil
}
