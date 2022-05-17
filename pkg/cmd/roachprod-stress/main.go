package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

var (
	flags       = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	flagP       = flags.Int("p", runtime.GOMAXPROCS(0), "run `N` processes in parallel")
	flagTimeout = flags.Duration("timeout", 0, "timeout each process after `duration`")
	_           = flags.Bool("kill", true, "kill timed out processes if true, otherwise just print pid (to attach with gdb)")
	flagFailure = flags.String("failure", "", "fail only if output matches `regexp`")
	flagIgnore  = flags.String("ignore", "", "ignore failure if output matches `regexp`")
	flagMaxTime = flags.Duration("maxtime", 0, "maximum time to run")
	flagMaxRuns = flags.Int("maxruns", 0, "maximum number of runs")
	_           = flags.Int("maxfails", 1, "maximum number of failures")
	flagStderr  = flags.Bool("stderr", true, "output failures to STDERR instead of to a temp file")
)

func roundToSeconds(d time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(43100)
	return time.Duration(d.Seconds()+0.5) * time.Second
}

func run() error {
	__antithesis_instrumentation__.Notify(43101)
	flags.Usage = func() {
		__antithesis_instrumentation__.Notify(43122)
		fmt.Fprintf(flags.Output(), "usage: %s <cluster> <pkg> [<flags>] -- [<args>]\n", flags.Name())
		flags.PrintDefaults()
	}
	__antithesis_instrumentation__.Notify(43102)

	if len(os.Args) < 2 {
		__antithesis_instrumentation__.Notify(43123)
		var b bytes.Buffer
		flags.SetOutput(&b)
		flags.Usage()
		return errors.Newf("%s", b.String())
	} else {
		__antithesis_instrumentation__.Notify(43124)
	}
	__antithesis_instrumentation__.Notify(43103)

	cluster := os.Args[1]
	if err := flags.Parse(os.Args[2:]); err != nil {
		__antithesis_instrumentation__.Notify(43125)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43126)
	}
	__antithesis_instrumentation__.Notify(43104)

	if !*flagStderr {
		__antithesis_instrumentation__.Notify(43127)
		return errors.New("-stderr=false is unsupported, please tee to a file (or implement the feature)")
	} else {
		__antithesis_instrumentation__.Notify(43128)
	}
	__antithesis_instrumentation__.Notify(43105)

	pkg := os.Args[2]
	localTestBin := filepath.Base(pkg) + ".test"
	{
		__antithesis_instrumentation__.Notify(43129)
		fi, err := os.Stat(pkg)
		if err != nil {
			__antithesis_instrumentation__.Notify(43133)
			return errors.Wrapf(err, "the pkg flag %q is not a directory relative to the current working directory", pkg)
		} else {
			__antithesis_instrumentation__.Notify(43134)
		}
		__antithesis_instrumentation__.Notify(43130)
		if !fi.Mode().IsDir() {
			__antithesis_instrumentation__.Notify(43135)
			return fmt.Errorf("the pkg flag %q is not a directory relative to the current working directory", pkg)
		} else {
			__antithesis_instrumentation__.Notify(43136)
		}
		__antithesis_instrumentation__.Notify(43131)

		fi, err = os.Stat(localTestBin)
		if err != nil {
			__antithesis_instrumentation__.Notify(43137)
			return errors.Wrapf(err, "test binary %q does not exist", localTestBin)
		} else {
			__antithesis_instrumentation__.Notify(43138)
		}
		__antithesis_instrumentation__.Notify(43132)
		if !fi.Mode().IsRegular() {
			__antithesis_instrumentation__.Notify(43139)
			return fmt.Errorf("test binary %q is not a file", localTestBin)
		} else {
			__antithesis_instrumentation__.Notify(43140)
		}
	}
	__antithesis_instrumentation__.Notify(43106)
	flagsAndArgs := os.Args[3:]
	stressArgs := flagsAndArgs
	var testArgs []string
	for i, arg := range flagsAndArgs {
		__antithesis_instrumentation__.Notify(43141)
		if arg == "--" {
			__antithesis_instrumentation__.Notify(43142)
			stressArgs = flagsAndArgs[:i]
			testArgs = flagsAndArgs[i+1:]
			break
		} else {
			__antithesis_instrumentation__.Notify(43143)
		}
	}
	__antithesis_instrumentation__.Notify(43107)

	if *flagP <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(43144)
		return *flagTimeout < 0 == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(43145)
		return len(flags.Args()) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(43146)
		var b bytes.Buffer
		flags.SetOutput(&b)
		flags.Usage()
		return errors.Newf("%s", b.String())
	} else {
		__antithesis_instrumentation__.Notify(43147)
	}
	__antithesis_instrumentation__.Notify(43108)
	if *flagFailure != "" {
		__antithesis_instrumentation__.Notify(43148)
		if _, err := regexp.Compile(*flagFailure); err != nil {
			__antithesis_instrumentation__.Notify(43149)
			return errors.Wrap(err, "bad failure regexp")
		} else {
			__antithesis_instrumentation__.Notify(43150)
		}
	} else {
		__antithesis_instrumentation__.Notify(43151)
	}
	__antithesis_instrumentation__.Notify(43109)
	if *flagIgnore != "" {
		__antithesis_instrumentation__.Notify(43152)
		if _, err := regexp.Compile(*flagIgnore); err != nil {
			__antithesis_instrumentation__.Notify(43153)
			return errors.Wrap(err, "bad ignore regexp")
		} else {
			__antithesis_instrumentation__.Notify(43154)
		}
	} else {
		__antithesis_instrumentation__.Notify(43155)
	}
	__antithesis_instrumentation__.Notify(43110)

	cmd := exec.Command("roachprod", "status", "-q", cluster)
	out, err := cmd.CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(43156)
		return errors.Wrapf(err, "%s", out)
	} else {
		__antithesis_instrumentation__.Notify(43157)
	}
	__antithesis_instrumentation__.Notify(43111)
	nodes := strings.Count(string(out), "\n") - 1

	const stressBin = "bin.docker_amd64/stress"

	cmd = exec.Command("roachprod", "put", cluster, stressBin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(43158)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43159)
	}
	__antithesis_instrumentation__.Notify(43112)

	const localLibDir = "lib.docker_amd64/"
	if fi, err := os.Stat(localLibDir); err == nil && func() bool {
		__antithesis_instrumentation__.Notify(43160)
		return fi.IsDir() == true
	}() == true {
		__antithesis_instrumentation__.Notify(43161)
		cmd = exec.Command("roachprod", "put", cluster, localLibDir, "lib")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			__antithesis_instrumentation__.Notify(43162)
			return err
		} else {
			__antithesis_instrumentation__.Notify(43163)
		}
	} else {
		__antithesis_instrumentation__.Notify(43164)
	}
	__antithesis_instrumentation__.Notify(43113)

	cmd = exec.Command("roachprod", "run", cluster, "mkdir -p "+pkg)
	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(43165)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43166)
	}
	__antithesis_instrumentation__.Notify(43114)
	testdataPath := filepath.Join(pkg, "testdata")
	if _, err := os.Stat(testdataPath); err == nil {
		__antithesis_instrumentation__.Notify(43167)

		tmpPath := "testdata" + strconv.Itoa(rand.Int())
		cmd = exec.Command("roachprod", "run", cluster, "--", "rm", "-rf", testdataPath)
		if output, err := cmd.CombinedOutput(); err != nil {
			__antithesis_instrumentation__.Notify(43170)
			return errors.Wrapf(err, "failed to remove old testdata:\n%s", output)
		} else {
			__antithesis_instrumentation__.Notify(43171)
		}
		__antithesis_instrumentation__.Notify(43168)
		cmd = exec.Command("roachprod", "put", cluster, testdataPath, tmpPath)
		if err := cmd.Run(); err != nil {
			__antithesis_instrumentation__.Notify(43172)
			return errors.Wrap(err, "failed to copy testdata")
		} else {
			__antithesis_instrumentation__.Notify(43173)
		}
		__antithesis_instrumentation__.Notify(43169)
		cmd = exec.Command("roachprod", "run", cluster, "mv", tmpPath, testdataPath)
		if err := cmd.Run(); err != nil {
			__antithesis_instrumentation__.Notify(43174)
			return errors.Wrap(err, "failed to move testdata")
		} else {
			__antithesis_instrumentation__.Notify(43175)
		}
	} else {
		__antithesis_instrumentation__.Notify(43176)
	}
	__antithesis_instrumentation__.Notify(43115)
	testBin := filepath.Join(pkg, localTestBin)
	cmd = exec.Command("roachprod", "put", cluster, localTestBin, testBin)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		__antithesis_instrumentation__.Notify(43177)
		return err
	} else {
		__antithesis_instrumentation__.Notify(43178)
	}
	__antithesis_instrumentation__.Notify(43116)

	c := make(chan os.Signal)
	defer close(c)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGTERM)
	defer signal.Stop(c)

	startTime := timeutil.Now()
	ctx, cancel := func(ctx context.Context) (context.Context, context.CancelFunc) {
		__antithesis_instrumentation__.Notify(43179)
		if *flagMaxTime > 0 {
			__antithesis_instrumentation__.Notify(43181)
			return context.WithTimeout(ctx, *flagMaxTime)
		} else {
			__antithesis_instrumentation__.Notify(43182)
		}
		__antithesis_instrumentation__.Notify(43180)
		return context.WithCancel(ctx)
	}(context.Background())
	__antithesis_instrumentation__.Notify(43117)
	defer cancel()

	go func() {
		__antithesis_instrumentation__.Notify(43183)
		<-ctx.Done()
		fmt.Printf("shutting down\n")
		_ = exec.Command("roachprod", "stop", cluster).Run()
	}()
	__antithesis_instrumentation__.Notify(43118)

	go func() {
		__antithesis_instrumentation__.Notify(43184)
		for range c {
			__antithesis_instrumentation__.Notify(43185)
			cancel()
		}
	}()
	__antithesis_instrumentation__.Notify(43119)

	var wg sync.WaitGroup
	defer wg.Wait()

	var runs, fails int32
	res := make(chan string)
	error := func(s string) {
		__antithesis_instrumentation__.Notify(43186)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(43187)
		case res <- s:
			__antithesis_instrumentation__.Notify(43188)
		}
	}
	__antithesis_instrumentation__.Notify(43120)

	statusRE := regexp.MustCompile(`(\d+) runs (so far|completed), (\d+) failures, over .*`)

	wg.Add(nodes)
	for i := 1; i <= nodes; i++ {
		__antithesis_instrumentation__.Notify(43189)
		go func(i int) {
			__antithesis_instrumentation__.Notify(43190)
			stdoutR, stdoutW := io.Pipe()
			defer func() {
				__antithesis_instrumentation__.Notify(43193)
				_ = stdoutW.Close()
				wg.Done()
			}()
			__antithesis_instrumentation__.Notify(43191)

			go func() {
				__antithesis_instrumentation__.Notify(43194)
				defer func() {
					__antithesis_instrumentation__.Notify(43196)
					_ = stdoutR.Close()
				}()
				__antithesis_instrumentation__.Notify(43195)

				var lastRuns, lastFails int
				scanner := bufio.NewScanner(stdoutR)
				for scanner.Scan() {
					__antithesis_instrumentation__.Notify(43197)
					m := statusRE.FindStringSubmatch(scanner.Text())
					if m == nil {
						__antithesis_instrumentation__.Notify(43202)
						continue
					} else {
						__antithesis_instrumentation__.Notify(43203)
					}
					__antithesis_instrumentation__.Notify(43198)
					curRuns, err := strconv.Atoi(m[1])
					if err != nil {
						__antithesis_instrumentation__.Notify(43204)
						error(fmt.Sprintf("%s", err))
						return
					} else {
						__antithesis_instrumentation__.Notify(43205)
					}
					__antithesis_instrumentation__.Notify(43199)
					curFails, err := strconv.Atoi(m[3])
					if err != nil {
						__antithesis_instrumentation__.Notify(43206)
						error(fmt.Sprintf("%s", err))
						return
					} else {
						__antithesis_instrumentation__.Notify(43207)
					}
					__antithesis_instrumentation__.Notify(43200)
					if m[2] == "completed" {
						__antithesis_instrumentation__.Notify(43208)
						break
					} else {
						__antithesis_instrumentation__.Notify(43209)
					}
					__antithesis_instrumentation__.Notify(43201)

					atomic.AddInt32(&runs, int32(curRuns-lastRuns))
					atomic.AddInt32(&fails, int32(curFails-lastFails))
					lastRuns, lastFails = curRuns, curFails

					if *flagMaxRuns > 0 && func() bool {
						__antithesis_instrumentation__.Notify(43210)
						return int(atomic.LoadInt32(&runs)) >= *flagMaxRuns == true
					}() == true {
						__antithesis_instrumentation__.Notify(43211)
						cancel()
					} else {
						__antithesis_instrumentation__.Notify(43212)
					}
				}
			}()
			__antithesis_instrumentation__.Notify(43192)
			var stderr bytes.Buffer
			cmd := exec.Command("roachprod",
				"ssh", fmt.Sprintf("%s:%d", cluster, i), "--",
				fmt.Sprintf("cd %s; GOTRACEBACK=all ~/stress %s ./%s %s",
					pkg,
					strings.Join(stressArgs, " "),
					filepath.Base(testBin),
					strings.Join(testArgs, " ")))
			cmd.Stdout = stdoutW
			cmd.Stderr = &stderr
			if err := cmd.Run(); err != nil {
				__antithesis_instrumentation__.Notify(43213)
				error(stderr.String())
			} else {
				__antithesis_instrumentation__.Notify(43214)
			}
		}(i)
	}
	__antithesis_instrumentation__.Notify(43121)

	ticker := time.NewTicker(5 * time.Second).C
	for {
		__antithesis_instrumentation__.Notify(43215)
		select {
		case out := <-res:
			__antithesis_instrumentation__.Notify(43216)
			cancel()
			fmt.Fprintf(os.Stderr, "\n%s\n", out)
		case <-ticker:
			__antithesis_instrumentation__.Notify(43217)
			fmt.Printf("%v runs so far, %v failures, over %s\n",
				atomic.LoadInt32(&runs), atomic.LoadInt32(&fails),
				roundToSeconds(timeutil.Since(startTime)))
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(43218)
			fmt.Printf("%v runs completed, %v failures, over %s\n",
				atomic.LoadInt32(&runs), atomic.LoadInt32(&fails),
				roundToSeconds(timeutil.Since(startTime)))

			err := ctx.Err()
			switch {

			case errors.Is(err, context.DeadlineExceeded):
				__antithesis_instrumentation__.Notify(43219)
				return nil
			case errors.Is(err, context.Canceled):
				__antithesis_instrumentation__.Notify(43220)
				if *flagMaxRuns > 0 && func() bool {
					__antithesis_instrumentation__.Notify(43223)
					return int(atomic.LoadInt32(&runs)) >= *flagMaxRuns == true
				}() == true {
					__antithesis_instrumentation__.Notify(43224)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(43225)
				}
				__antithesis_instrumentation__.Notify(43221)
				return err
			default:
				__antithesis_instrumentation__.Notify(43222)
				return errors.Wrap(err, "unexpected context error")
			}
		}
	}
}

func main() {
	__antithesis_instrumentation__.Notify(43226)
	if err := run(); err != nil {
		__antithesis_instrumentation__.Notify(43227)
		fmt.Fprintln(os.Stderr, err)
		fmt.Println("FAIL")
		os.Exit(1)
	} else {
		__antithesis_instrumentation__.Notify(43228)
		fmt.Println("SUCCESS")
	}
}
