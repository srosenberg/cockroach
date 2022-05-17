package codeowners

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/internal/reporoot"
	"github.com/cockroachdb/cockroach/pkg/internal/team"
	"github.com/cockroachdb/errors"
	"github.com/zabawaba99/go-gitignore"
)

type Rule struct {
	Pattern string
	Owners  []team.Alias
}

type CodeOwners struct {
	rules []Rule
	teams map[team.Alias]team.Team
}

func LoadCodeOwners(r io.Reader, teams map[team.Alias]team.Team) (*CodeOwners, error) {
	__antithesis_instrumentation__.Notify(68407)
	s := bufio.NewScanner(r)
	ret := &CodeOwners{
		teams: teams,
	}
	lineNum := 1
	for s.Scan() {
		__antithesis_instrumentation__.Notify(68409)
		lineNum++
		if s.Err() != nil {
			__antithesis_instrumentation__.Notify(68414)
			return nil, s.Err()
		} else {
			__antithesis_instrumentation__.Notify(68415)
		}
		__antithesis_instrumentation__.Notify(68410)
		t := s.Text()
		if strings.HasPrefix(t, "#") {
			__antithesis_instrumentation__.Notify(68416)
			continue
		} else {
			__antithesis_instrumentation__.Notify(68417)
		}
		__antithesis_instrumentation__.Notify(68411)
		t = strings.TrimSpace(t)
		if len(t) == 0 {
			__antithesis_instrumentation__.Notify(68418)
			continue
		} else {
			__antithesis_instrumentation__.Notify(68419)
		}
		__antithesis_instrumentation__.Notify(68412)

		fields := strings.Fields(t)
		rule := Rule{Pattern: fields[0]}
		for _, field := range fields[1:] {
			__antithesis_instrumentation__.Notify(68420)

			owner := team.Alias(strings.TrimSuffix(strings.TrimPrefix(field, "@"), "-noreview"))

			if _, ok := teams[owner]; !ok {
				__antithesis_instrumentation__.Notify(68422)
				return nil, errors.Newf("owner %s does not exist", owner)
			} else {
				__antithesis_instrumentation__.Notify(68423)
			}
			__antithesis_instrumentation__.Notify(68421)
			rule.Owners = append(rule.Owners, owner)
		}
		__antithesis_instrumentation__.Notify(68413)
		ret.rules = append(ret.rules, rule)
	}
	__antithesis_instrumentation__.Notify(68408)
	return ret, nil
}

func DefaultLoadCodeOwners() (*CodeOwners, error) {
	__antithesis_instrumentation__.Notify(68424)
	teams, err := team.DefaultLoadTeams()
	if err != nil {
		__antithesis_instrumentation__.Notify(68429)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(68430)
	}
	__antithesis_instrumentation__.Notify(68425)
	path := reporoot.GetFor(".", ".github/CODEOWNERS")
	if path == "" {
		__antithesis_instrumentation__.Notify(68431)
		return nil, errors.Errorf("CODEOWNERS not found")
	} else {
		__antithesis_instrumentation__.Notify(68432)
	}
	__antithesis_instrumentation__.Notify(68426)
	path = filepath.Join(path, ".github/CODEOWNERS")
	f, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(68433)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(68434)
	}
	__antithesis_instrumentation__.Notify(68427)
	defer func() { __antithesis_instrumentation__.Notify(68435); _ = f.Close() }()
	__antithesis_instrumentation__.Notify(68428)
	return LoadCodeOwners(f, teams)
}

func (co *CodeOwners) Match(filePath string) []team.Team {
	__antithesis_instrumentation__.Notify(68436)
	if filepath.IsAbs(filePath) {
		__antithesis_instrumentation__.Notify(68439)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(68440)
	}
	__antithesis_instrumentation__.Notify(68437)
	filePath = string(filepath.Separator) + filepath.Clean(filePath)

	lastFilePath := ""
	for filePath != lastFilePath {
		__antithesis_instrumentation__.Notify(68441)

		for i := len(co.rules) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(68443)
			rule := co.rules[i]

			if gitignore.Match(rule.Pattern, filePath) {
				__antithesis_instrumentation__.Notify(68444)
				teams := make([]team.Team, len(rule.Owners))
				for i, owner := range rule.Owners {
					__antithesis_instrumentation__.Notify(68446)
					teams[i] = co.teams[owner]
				}
				__antithesis_instrumentation__.Notify(68445)
				return teams
			} else {
				__antithesis_instrumentation__.Notify(68447)
			}
		}
		__antithesis_instrumentation__.Notify(68442)
		lastFilePath = filePath
		filePath = filepath.Dir(filePath)
	}
	__antithesis_instrumentation__.Notify(68438)
	return nil
}
