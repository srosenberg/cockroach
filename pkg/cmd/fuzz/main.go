// fuzz builds and executes fuzz tests.
//
// Fuzz tests can be added to CockroachDB by adding a function of the form:
//   func FuzzXXX(data []byte) int
// To help the fuzzer increase coverage, this function should return 1 on
// interesting input (for example, a parse succeeded) and 0 otherwise. Panics
// will be detected and reported.
//
// To exclude this file except during fuzzing, tag it with:
//   // +build gofuzz
package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"golang.org/x/tools/go/packages"
)

var (
	flags   = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	tests   = flags.String("tests", "", "regex filter for tests to run")
	timeout = flags.Duration("timeout", 1*time.Minute, "time to run each Fuzz func")
	verbose = flags.Bool("v", false, "verbose output")
)

func usage() {
	__antithesis_instrumentation__.Notify(40061)
	fmt.Fprintf(flags.Output(), "Usage of %s:\n", os.Args[0])
	flags.PrintDefaults()
	os.Exit(1)
}

func main() {
	__antithesis_instrumentation__.Notify(40062)

	for _, file := range []string{"go-fuzz", "go-fuzz-build"} {
		__antithesis_instrumentation__.Notify(40067)
		if _, err := exec.LookPath(file); err != nil {
			__antithesis_instrumentation__.Notify(40068)
			fmt.Println(file, "must be in your PATH")
			fmt.Println("Run `go get -u github.com/dvyukov/go-fuzz/...` to install.")
			os.Exit(1)
		} else {
			__antithesis_instrumentation__.Notify(40069)
		}
	}
	__antithesis_instrumentation__.Notify(40063)
	if err := flags.Parse(os.Args[1:]); err != nil {
		__antithesis_instrumentation__.Notify(40070)
		usage()
	} else {
		__antithesis_instrumentation__.Notify(40071)
	}
	__antithesis_instrumentation__.Notify(40064)
	patterns := flags.Args()
	if len(patterns) == 0 {
		__antithesis_instrumentation__.Notify(40072)
		fmt.Print("missing packages\n\n")
		usage()
	} else {
		__antithesis_instrumentation__.Notify(40073)
	}
	__antithesis_instrumentation__.Notify(40065)
	crashers, err := fuzz(patterns, *tests, *timeout)
	if err != nil {
		__antithesis_instrumentation__.Notify(40074)
		fmt.Println(err)
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(40075)
	}
	__antithesis_instrumentation__.Notify(40066)
	if crashers > 0 {
		__antithesis_instrumentation__.Notify(40076)
		fmt.Println(crashers, "crashers")
		os.Exit(2)
	} else {
		__antithesis_instrumentation__.Notify(40077)
	}
}

func fatal(arg interface{}) {
	__antithesis_instrumentation__.Notify(40078)
	panic(arg)
}

func log(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(40079)
	if !*verbose {
		__antithesis_instrumentation__.Notify(40081)
		return
	} else {
		__antithesis_instrumentation__.Notify(40082)
	}
	__antithesis_instrumentation__.Notify(40080)
	fmt.Fprintf(os.Stderr, format, args...)
}

func fuzz(patterns []string, tests string, timeout time.Duration) (int, error) {
	__antithesis_instrumentation__.Notify(40083)
	ctx := context.Background()
	pkgs, err := packages.Load(&packages.Config{
		Mode:       packages.NeedFiles,
		BuildFlags: []string{"-tags", "gofuzz"},
	}, patterns...)
	if err != nil {
		__antithesis_instrumentation__.Notify(40087)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(40088)
	}
	__antithesis_instrumentation__.Notify(40084)
	var testsRE *regexp.Regexp
	if tests != "" {
		__antithesis_instrumentation__.Notify(40089)
		testsRE, err = regexp.Compile(tests)
		if err != nil {
			__antithesis_instrumentation__.Notify(40090)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(40091)
		}
	} else {
		__antithesis_instrumentation__.Notify(40092)
	}
	__antithesis_instrumentation__.Notify(40085)
	crashers := 0
	for _, pkg := range pkgs {
		__antithesis_instrumentation__.Notify(40093)
		if len(pkg.Errors) > 0 {
			__antithesis_instrumentation__.Notify(40098)
			return 0, pkg.Errors[0]
		} else {
			__antithesis_instrumentation__.Notify(40099)
		}
		__antithesis_instrumentation__.Notify(40094)
		log("%s: searching for Fuzz funcs\n", pkg)
		fns, err := findFuncs(pkg)
		if err != nil {
			__antithesis_instrumentation__.Notify(40100)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(40101)
		}
		__antithesis_instrumentation__.Notify(40095)
		if len(fns) == 0 {
			__antithesis_instrumentation__.Notify(40102)
			continue
		} else {
			__antithesis_instrumentation__.Notify(40103)
		}
		__antithesis_instrumentation__.Notify(40096)
		dir := filepath.Dir(pkg.GoFiles[0])
		{
			__antithesis_instrumentation__.Notify(40104)
			log("%s: executing go-fuzz-build...", pkg)
			cmd := exec.Command("go-fuzz-build",

				"-preserve", "github.com/cockroachdb/cockroach/pkg/sql/stats,github.com/cockroachdb/cockroach/pkg/server/serverpb",
			)
			cmd.Dir = dir
			out, err := cmd.CombinedOutput()
			log(" done\n")
			if err != nil {
				__antithesis_instrumentation__.Notify(40105)
				log("%s\n", out)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(40106)
			}
		}
		__antithesis_instrumentation__.Notify(40097)
		for _, fn := range fns {
			__antithesis_instrumentation__.Notify(40107)
			if testsRE != nil && func() bool {
				__antithesis_instrumentation__.Notify(40109)
				return !testsRE.MatchString(fn) == true
			}() == true {
				__antithesis_instrumentation__.Notify(40110)
				continue
			} else {
				__antithesis_instrumentation__.Notify(40111)
			}
			__antithesis_instrumentation__.Notify(40108)
			crashers += execGoFuzz(ctx, pkg, dir, fn, timeout)
		}
	}
	__antithesis_instrumentation__.Notify(40086)
	return crashers, nil
}

var goFuzzRE = regexp.MustCompile(`crashers: (\d+)`)

func execGoFuzz(
	ctx context.Context, pkg *packages.Package, dir, fn string, timeout time.Duration,
) int {
	__antithesis_instrumentation__.Notify(40112)
	log("\n%s: fuzzing %s for %v\n", pkg, fn, timeout)
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	workdir := fmt.Sprintf("work-%s", fn)
	cmd := exec.CommandContext(ctx, "go-fuzz", "-func", fn, "-workdir", workdir)
	cmd.Dir = dir
	stderr, err := cmd.StderrPipe()
	if err != nil {
		__antithesis_instrumentation__.Notify(40118)
		fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40119)
	}
	__antithesis_instrumentation__.Notify(40113)
	if err := cmd.Start(); err != nil {
		__antithesis_instrumentation__.Notify(40120)
		fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40121)
	}
	__antithesis_instrumentation__.Notify(40114)
	crashers := 0
	scanner := bufio.NewScanner(stderr)
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(40122)
		line := scanner.Text()
		log("%s\n", line)
		matches := goFuzzRE.FindStringSubmatch(line)
		if len(matches) == 0 {
			__antithesis_instrumentation__.Notify(40125)
			continue
		} else {
			__antithesis_instrumentation__.Notify(40126)
		}
		__antithesis_instrumentation__.Notify(40123)
		i, err := strconv.Atoi(matches[1])
		if err != nil {
			__antithesis_instrumentation__.Notify(40127)
			fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(40128)
		}
		__antithesis_instrumentation__.Notify(40124)
		if i > crashers {
			__antithesis_instrumentation__.Notify(40129)
			if crashers == 0 {
				__antithesis_instrumentation__.Notify(40131)
				fmt.Printf("workdir: %s\n", filepath.Join(dir, workdir))
			} else {
				__antithesis_instrumentation__.Notify(40132)
			}
			__antithesis_instrumentation__.Notify(40130)
			crashers = i
			fmt.Printf("crashers: %d\n", crashers)
		} else {
			__antithesis_instrumentation__.Notify(40133)
		}
	}
	__antithesis_instrumentation__.Notify(40115)
	if err := scanner.Err(); err != nil {
		__antithesis_instrumentation__.Notify(40134)
		fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40135)
	}
	__antithesis_instrumentation__.Notify(40116)
	if err := cmd.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(40136)
		fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(40137)
	}
	__antithesis_instrumentation__.Notify(40117)
	return crashers
}

var fuzzFuncRE = regexp.MustCompile(`(?m)^func (Fuzz\w*)\(\w+ \[\]byte\) int {$`)

func findFuncs(pkg *packages.Package) ([]string, error) {
	__antithesis_instrumentation__.Notify(40138)
	var ret []string
	for _, file := range pkg.GoFiles {
		__antithesis_instrumentation__.Notify(40140)
		content, err := ioutil.ReadFile(file)
		if err != nil {
			__antithesis_instrumentation__.Notify(40142)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(40143)
		}
		__antithesis_instrumentation__.Notify(40141)
		matches := fuzzFuncRE.FindAllSubmatch(content, -1)
		for _, match := range matches {
			__antithesis_instrumentation__.Notify(40144)
			ret = append(ret, string(match[1]))
		}
	}
	__antithesis_instrumentation__.Notify(40139)
	return ret, nil
}
