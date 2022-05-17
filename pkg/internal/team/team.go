// Package team involves processing team information based on a yaml
// file containing team metadata.
package team

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/internal/reporoot"
	"github.com/cockroachdb/errors"
	"gopkg.in/yaml.v2"
)

type Alias string

type Team struct {
	TeamName Alias

	Aliases map[Alias]Purpose `yaml:"aliases"`

	TriageColumnID int `yaml:"triage_column_id"`

	Email string `yaml:"email"`

	Slack string `yaml:"slack"`
}

func (t Team) Name() Alias {
	__antithesis_instrumentation__.Notify(70021)
	return t.TeamName
}

type Map map[Alias]Team

func (m Map) GetAliasesForPurpose(alias Alias, purpose Purpose) ([]Alias, bool) {
	__antithesis_instrumentation__.Notify(70022)
	tm, ok := m[alias]
	if !ok {
		__antithesis_instrumentation__.Notify(70026)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(70027)
	}
	__antithesis_instrumentation__.Notify(70023)
	var sl []Alias
	for hdl, purp := range tm.Aliases {
		__antithesis_instrumentation__.Notify(70028)
		if purpose != purp {
			__antithesis_instrumentation__.Notify(70030)
			continue
		} else {
			__antithesis_instrumentation__.Notify(70031)
		}
		__antithesis_instrumentation__.Notify(70029)
		sl = append(sl, hdl)
	}
	__antithesis_instrumentation__.Notify(70024)
	if len(sl) == 0 {
		__antithesis_instrumentation__.Notify(70032)
		sl = append(sl, tm.Name())
	} else {
		__antithesis_instrumentation__.Notify(70033)
	}
	__antithesis_instrumentation__.Notify(70025)
	return sl, true
}

func DefaultLoadTeams() (Map, error) {
	__antithesis_instrumentation__.Notify(70034)
	path := reporoot.GetFor(".", "TEAMS.yaml")
	if path == "" {
		__antithesis_instrumentation__.Notify(70038)
		return nil, errors.New("TEAMS.yaml not found")
	} else {
		__antithesis_instrumentation__.Notify(70039)
	}
	__antithesis_instrumentation__.Notify(70035)
	path = filepath.Join(path, "TEAMS.yaml")
	f, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(70040)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70041)
	}
	__antithesis_instrumentation__.Notify(70036)
	defer func() { __antithesis_instrumentation__.Notify(70042); _ = f.Close() }()
	__antithesis_instrumentation__.Notify(70037)
	return LoadTeams(f)
}

type Purpose string

const (
	PurposeOther = Purpose("other")

	PurposeRoachtest = Purpose("roachtest")
)

var validPurposes = map[Purpose]struct{}{
	PurposeOther:     {},
	PurposeRoachtest: {},
}

func LoadTeams(f io.Reader) (Map, error) {
	__antithesis_instrumentation__.Notify(70043)
	b, err := ioutil.ReadAll(f)
	if err != nil {
		__antithesis_instrumentation__.Notify(70047)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70048)
	}
	__antithesis_instrumentation__.Notify(70044)
	var src map[Alias]Team
	if err := yaml.UnmarshalStrict(b, &src); err != nil {
		__antithesis_instrumentation__.Notify(70049)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(70050)
	}
	__antithesis_instrumentation__.Notify(70045)

	dst := map[Alias]Team{}

	for k, v := range src {
		__antithesis_instrumentation__.Notify(70051)
		v.TeamName = k
		aliases := []Alias{v.TeamName}
		for alias, purpose := range v.Aliases {
			__antithesis_instrumentation__.Notify(70054)
			if _, ok := validPurposes[purpose]; !ok {
				__antithesis_instrumentation__.Notify(70056)
				return nil, errors.Errorf("team %s has alias %s with invalid purpose %s", k, alias, purpose)
			} else {
				__antithesis_instrumentation__.Notify(70057)
			}
			__antithesis_instrumentation__.Notify(70055)
			aliases = append(aliases, alias)
		}
		__antithesis_instrumentation__.Notify(70052)
		for _, alias := range aliases {
			__antithesis_instrumentation__.Notify(70058)
			if conflicting, ok := dst[alias]; ok {
				__antithesis_instrumentation__.Notify(70060)
				return nil, errors.Errorf(
					"team %s has alias %s which conflicts with team %s",
					k, alias, conflicting.Name(),
				)
			} else {
				__antithesis_instrumentation__.Notify(70061)
			}
			__antithesis_instrumentation__.Notify(70059)
			dst[alias] = v
		}
		__antithesis_instrumentation__.Notify(70053)
		dst[k] = v
	}
	__antithesis_instrumentation__.Notify(70046)
	return dst, nil
}
