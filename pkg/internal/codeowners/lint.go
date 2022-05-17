package codeowners

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/stretchr/testify/require"
)

func LintEverythingIsOwned(
	t interface {
		Logf(string, ...interface{})
		Errorf(string, ...interface{})
		FailNow()
		Helper()
	},
	verbose bool,
	co *CodeOwners,
	repoRoot string,
	walkDir string,
) {
	__antithesis_instrumentation__.Notify(68448)
	debug := func(format string, args ...interface{}) {
		__antithesis_instrumentation__.Notify(68454)
		if !verbose {
			__antithesis_instrumentation__.Notify(68456)
			return
		} else {
			__antithesis_instrumentation__.Notify(68457)
		}
		__antithesis_instrumentation__.Notify(68455)
		t.Helper()
		t.Logf(format, args...)
	}
	__antithesis_instrumentation__.Notify(68449)

	skip := map[string]struct{}{
		filepath.Join("ccl", "ccl_init.go"): {},
		filepath.Join("node_modules"):       {},
		filepath.Join("yarn-vendor"):        {},
		"Makefile":                          {},
		"BUILD.bazel":                       {},
		".gitignore":                        {},
		"README.md":                         {},
	}

	unowned := map[string]string{}

	walkRoot := filepath.Join(repoRoot, walkDir)

	unownedWalkFn := func(path string, info os.FileInfo) error {
		__antithesis_instrumentation__.Notify(68458)
		teams := co.Match(path)
		if len(teams) > 0 {
			__antithesis_instrumentation__.Notify(68461)

			debug("%s <- has team(s) %v", path, teams)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(68462)
		}
		__antithesis_instrumentation__.Notify(68459)
		if !info.IsDir() {
			__antithesis_instrumentation__.Notify(68463)

			parts := strings.Split(path, string(filepath.Separator))
			var ok bool
			for i := range parts {
				__antithesis_instrumentation__.Notify(68465)
				prefix := filepath.Join(parts[:i+1]...)
				_, ok = unowned[prefix]
				if ok {
					__antithesis_instrumentation__.Notify(68466)
					debug("pruning %s; %s is already unowned", path, prefix)
					break
				} else {
					__antithesis_instrumentation__.Notify(68467)
				}
			}
			__antithesis_instrumentation__.Notify(68464)
			if !ok {
				__antithesis_instrumentation__.Notify(68468)
				debug("unowned: %s", path)
				unowned[filepath.Dir(path)] = path
			} else {
				__antithesis_instrumentation__.Notify(68469)
			}
		} else {
			__antithesis_instrumentation__.Notify(68470)
		}
		__antithesis_instrumentation__.Notify(68460)
		return nil
	}
	__antithesis_instrumentation__.Notify(68450)

	dirsToWalk := []string{walkRoot}
	for len(dirsToWalk) != 0 {
		__antithesis_instrumentation__.Notify(68471)

		require.NoError(t, filepath.Walk(dirsToWalk[0], func(path string, info os.FileInfo, err error) error {
			__antithesis_instrumentation__.Notify(68473)
			if err != nil {
				__antithesis_instrumentation__.Notify(68479)
				return err
			} else {
				__antithesis_instrumentation__.Notify(68480)
			}
			__antithesis_instrumentation__.Notify(68474)

			relPath, err := filepath.Rel(walkRoot, path)
			if err != nil {
				__antithesis_instrumentation__.Notify(68481)
				return err
			} else {
				__antithesis_instrumentation__.Notify(68482)
			}
			__antithesis_instrumentation__.Notify(68475)

			if _, ok := skip[relPath]; ok {
				__antithesis_instrumentation__.Notify(68483)
				debug("skipping %s", relPath)
				if info.IsDir() {
					__antithesis_instrumentation__.Notify(68485)
					return filepath.SkipDir
				} else {
					__antithesis_instrumentation__.Notify(68486)
				}
				__antithesis_instrumentation__.Notify(68484)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(68487)
			}
			__antithesis_instrumentation__.Notify(68476)
			fname := filepath.Base(relPath)
			if _, ok := skip[fname]; ok {
				__antithesis_instrumentation__.Notify(68488)
				debug("skipping %s", relPath)
				if info.IsDir() {
					__antithesis_instrumentation__.Notify(68490)
					return filepath.SkipDir
				} else {
					__antithesis_instrumentation__.Notify(68491)
				}
				__antithesis_instrumentation__.Notify(68489)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(68492)
			}
			__antithesis_instrumentation__.Notify(68477)

			if info.IsDir() {
				__antithesis_instrumentation__.Notify(68493)
				if path == dirsToWalk[0] {
					__antithesis_instrumentation__.Notify(68495)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(68496)
				}
				__antithesis_instrumentation__.Notify(68494)
				dirsToWalk = append(dirsToWalk, path)
				return filepath.SkipDir
			} else {
				__antithesis_instrumentation__.Notify(68497)
			}
			__antithesis_instrumentation__.Notify(68478)
			return unownedWalkFn(filepath.Join(walkDir, relPath), info)
		}))
		__antithesis_instrumentation__.Notify(68472)
		dirsToWalk = dirsToWalk[1:]
	}
	__antithesis_instrumentation__.Notify(68451)
	var sl []string
	for path := range unowned {
		__antithesis_instrumentation__.Notify(68498)
		sl = append(sl, path)
	}
	__antithesis_instrumentation__.Notify(68452)
	sort.Strings(sl)

	var buf strings.Builder
	for _, s := range sl {
		__antithesis_instrumentation__.Notify(68499)
		fmt.Fprintf(&buf, "%-28s @cockroachdb/<TODO>-noreview\n", string(filepath.Separator)+s+string(filepath.Separator))
	}
	__antithesis_instrumentation__.Notify(68453)
	if buf.Len() > 0 {
		__antithesis_instrumentation__.Notify(68500)
		t.Errorf(`unowned packages found, please fill out the below and augment .github/CODEOWNERS:
Remove the '-noreview' suffix if the team should be requested for Github reviews.

%s`, buf.String())
	} else {
		__antithesis_instrumentation__.Notify(68501)
	}
}
