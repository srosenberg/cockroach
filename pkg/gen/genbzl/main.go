// Command genbzl is used to generate bazel files which then get imported by
// the gen package's BUILD.bazel to facilitate hoisting these generated files
// back into the source tree.
//
// The flow is that we invoke this binary inside the
// bazelutil/bazel-generate.sh script which writes out some bzl files
// defining lists of targets for generation and hoisting.
//
// The program assumes it will be run from the root of the cockroach workspace.
// If it's not, errors will occur.
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
)

var (
	outDir = flag.String(
		"out-dir", "",
		"directory in which to place the generated files",
	)
)

func main() {
	__antithesis_instrumentation__.Notify(58485)
	flag.Parse()
	if err := generate(*outDir); err != nil {
		__antithesis_instrumentation__.Notify(58486)
		fmt.Fprintf(os.Stderr, "failed to generate files: %v\n", err)
		exit.WithCode(exit.UnspecifiedError())
	} else {
		__antithesis_instrumentation__.Notify(58487)
	}
}

func generate(outDir string) error {
	__antithesis_instrumentation__.Notify(58488)
	qd, err := getQueryData()
	if err != nil {
		__antithesis_instrumentation__.Notify(58491)
		return err
	} else {
		__antithesis_instrumentation__.Notify(58492)
	}
	__antithesis_instrumentation__.Notify(58489)
	for _, q := range targets {
		__antithesis_instrumentation__.Notify(58493)
		if q.doNotGenerate {
			__antithesis_instrumentation__.Notify(58496)
			continue
		} else {
			__antithesis_instrumentation__.Notify(58497)
		}
		__antithesis_instrumentation__.Notify(58494)
		out, err := q.target.execQuery(qd)
		if err != nil {
			__antithesis_instrumentation__.Notify(58498)
			return err
		} else {
			__antithesis_instrumentation__.Notify(58499)
		}
		__antithesis_instrumentation__.Notify(58495)
		if err := q.target.write(outDir, out); err != nil {
			__antithesis_instrumentation__.Notify(58500)
			return err
		} else {
			__antithesis_instrumentation__.Notify(58501)
		}
	}
	__antithesis_instrumentation__.Notify(58490)
	return nil
}

func getQueryData() (*queryData, error) {
	__antithesis_instrumentation__.Notify(58502)
	dirs := []string{"build", "docs"}
	ents, err := os.ReadDir("pkg")
	if err != nil {
		__antithesis_instrumentation__.Notify(58506)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(58507)
	}
	__antithesis_instrumentation__.Notify(58503)
	toSkip := map[string]struct{}{"ui": {}, "gen": {}}
	for _, e := range ents {
		__antithesis_instrumentation__.Notify(58508)
		if _, shouldSkip := toSkip[e.Name()]; shouldSkip || func() bool {
			__antithesis_instrumentation__.Notify(58510)
			return !e.IsDir() == true
		}() == true {
			__antithesis_instrumentation__.Notify(58511)
			continue
		} else {
			__antithesis_instrumentation__.Notify(58512)
		}
		__antithesis_instrumentation__.Notify(58509)
		dirs = append(dirs, filepath.Join("pkg", e.Name()))
	}
	__antithesis_instrumentation__.Notify(58504)
	exprs := make([]string, 0, len(dirs))
	for _, dir := range dirs {
		__antithesis_instrumentation__.Notify(58513)
		exprs = append(exprs, "//"+dir+"/...:*")
	}
	__antithesis_instrumentation__.Notify(58505)
	return &queryData{
		All: "(" + strings.Join(exprs, " + ") + ")",
	}, nil
}
