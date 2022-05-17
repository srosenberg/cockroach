package buildutil

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/build"
	"runtime"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/errors"
)

func short(in string) string {
	__antithesis_instrumentation__.Notify(643978)
	return strings.Replace(in, "github.com/cockroachdb/cockroach/pkg/", "./pkg/", -1)
}

func VerifyNoImports(
	t testing.TB,
	pkgPath string,
	cgo bool,
	forbiddenPkgs, forbiddenPrefixes []string,
	allowlist ...string,
) {
	__antithesis_instrumentation__.Notify(643979)

	if build.Default.GOPATH == "" {
		__antithesis_instrumentation__.Notify(643982)
		skip.IgnoreLint(t, "GOPATH isn't set")
	} else {
		__antithesis_instrumentation__.Notify(643983)
	}
	__antithesis_instrumentation__.Notify(643980)

	buildContext := build.Default
	buildContext.CgoEnabled = cgo

	checked := make(map[string]struct{})

	var check func(string) error
	check = func(path string) error {
		__antithesis_instrumentation__.Notify(643984)
		pkg, err := buildContext.Import(path, "", build.FindOnly)
		if err != nil {
			__antithesis_instrumentation__.Notify(643987)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(643988)
		}
		__antithesis_instrumentation__.Notify(643985)
		for _, imp := range pkg.Imports {
			__antithesis_instrumentation__.Notify(643989)
			for _, forbidden := range forbiddenPkgs {
				__antithesis_instrumentation__.Notify(643996)
				if forbidden == imp {
					__antithesis_instrumentation__.Notify(643998)
					allowlisted := false
					for _, w := range allowlist {
						__antithesis_instrumentation__.Notify(644000)
						if path == w {
							__antithesis_instrumentation__.Notify(644001)
							allowlisted = true
							break
						} else {
							__antithesis_instrumentation__.Notify(644002)
						}
					}
					__antithesis_instrumentation__.Notify(643999)
					if !allowlisted {
						__antithesis_instrumentation__.Notify(644003)
						return errors.Errorf("%s imports %s, which is forbidden", short(path), short(imp))
					} else {
						__antithesis_instrumentation__.Notify(644004)
					}
				} else {
					__antithesis_instrumentation__.Notify(644005)
				}
				__antithesis_instrumentation__.Notify(643997)
				if forbidden == "c-deps" && func() bool {
					__antithesis_instrumentation__.Notify(644006)
					return imp == "C" == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(644007)
					return strings.HasPrefix(path, "github.com/cockroachdb/cockroach/pkg") == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(644008)
					return path != "github.com/cockroachdb/cockroach/pkg/geo/geoproj" == true
				}() == true {
					__antithesis_instrumentation__.Notify(644009)
					for _, name := range pkg.CgoFiles {
						__antithesis_instrumentation__.Notify(644010)
						if strings.Contains(name, "zcgo_flags") {
							__antithesis_instrumentation__.Notify(644011)
							return errors.Errorf("%s imports %s (%s), which is forbidden", short(path), short(imp), name)
						} else {
							__antithesis_instrumentation__.Notify(644012)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(644013)
				}
			}
			__antithesis_instrumentation__.Notify(643990)
			for _, prefix := range forbiddenPrefixes {
				__antithesis_instrumentation__.Notify(644014)
				if strings.HasPrefix(imp, prefix) {
					__antithesis_instrumentation__.Notify(644015)
					return errors.Errorf("%s imports %s which has prefix %s, which is forbidden", short(path), short(imp), prefix)
				} else {
					__antithesis_instrumentation__.Notify(644016)
				}
			}
			__antithesis_instrumentation__.Notify(643991)

			if imp == "C" {
				__antithesis_instrumentation__.Notify(644017)
				continue
			} else {
				__antithesis_instrumentation__.Notify(644018)
			}
			__antithesis_instrumentation__.Notify(643992)

			importPkg, err := buildContext.Import(imp, pkg.Dir, build.FindOnly)
			if err != nil {
				__antithesis_instrumentation__.Notify(644019)

				if runtime.Compiler == "gccgo" {
					__antithesis_instrumentation__.Notify(644021)
					continue
				} else {
					__antithesis_instrumentation__.Notify(644022)
				}
				__antithesis_instrumentation__.Notify(644020)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(644023)
			}
			__antithesis_instrumentation__.Notify(643993)
			imp = importPkg.ImportPath
			if _, ok := checked[imp]; ok {
				__antithesis_instrumentation__.Notify(644024)
				continue
			} else {
				__antithesis_instrumentation__.Notify(644025)
			}
			__antithesis_instrumentation__.Notify(643994)
			if err := check(imp); err != nil {
				__antithesis_instrumentation__.Notify(644026)
				return errors.Wrapf(err, "%s depends on", short(path))
			} else {
				__antithesis_instrumentation__.Notify(644027)
			}
			__antithesis_instrumentation__.Notify(643995)
			checked[pkg.ImportPath] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(643986)
		return nil
	}
	__antithesis_instrumentation__.Notify(643981)
	if err := check(pkgPath); err != nil {
		__antithesis_instrumentation__.Notify(644028)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(644029)
	}
}
