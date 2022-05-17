package main

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"

	_ "net/http/pprof"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/petermattis/goid"
)

const perfArtifactsDir = "perf"

type testStatus struct {
	msg      string
	time     time.Time
	progress float64
}

type testImpl struct {
	spec *registry.TestSpec

	cockroach          string
	deprecatedWorkload string

	buildVersion version.Version

	l *logger.Logger

	runner string

	runnerID int64
	start    time.Time
	end      time.Time

	artifactsDir string

	artifactsSpec string

	mu struct {
		syncutil.RWMutex
		done    bool
		failed  bool
		timeout bool

		cancel  func()
		failLoc struct {
			file string
			line int
		}
		failureMsg string

		status map[int64]testStatus
		output []byte
	}

	versionsBinaryOverride map[string]string
}

func (t *testImpl) BuildVersion() *version.Version {
	__antithesis_instrumentation__.Notify(44586)
	return &t.buildVersion
}

func (t *testImpl) Cockroach() string {
	__antithesis_instrumentation__.Notify(44587)
	return t.cockroach
}

func (t *testImpl) DeprecatedWorkload() string {
	__antithesis_instrumentation__.Notify(44588)
	return t.deprecatedWorkload
}

func (t *testImpl) VersionsBinaryOverride() map[string]string {
	__antithesis_instrumentation__.Notify(44589)
	return t.versionsBinaryOverride
}

func (t *testImpl) Spec() interface{} {
	__antithesis_instrumentation__.Notify(44590)
	return t.spec
}

func (t *testImpl) Helper() { __antithesis_instrumentation__.Notify(44591) }

func (t *testImpl) Name() string {
	__antithesis_instrumentation__.Notify(44592)
	return t.spec.Name
}

func (t *testImpl) L() *logger.Logger {
	__antithesis_instrumentation__.Notify(44593)
	return t.l
}

func (t *testImpl) ReplaceL(l *logger.Logger) {
	__antithesis_instrumentation__.Notify(44594)

	t.l = l
}

func (t *testImpl) status(ctx context.Context, id int64, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44595)
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.mu.status == nil {
		__antithesis_instrumentation__.Notify(44598)
		t.mu.status = make(map[int64]testStatus)
	} else {
		__antithesis_instrumentation__.Notify(44599)
	}
	__antithesis_instrumentation__.Notify(44596)
	if len(args) == 0 {
		__antithesis_instrumentation__.Notify(44600)
		delete(t.mu.status, id)
		return
	} else {
		__antithesis_instrumentation__.Notify(44601)
	}
	__antithesis_instrumentation__.Notify(44597)
	msg := fmt.Sprint(args...)
	t.mu.status[id] = testStatus{
		msg:  msg,
		time: timeutil.Now(),
	}
	if !t.L().Closed() {
		__antithesis_instrumentation__.Notify(44602)
		if id == t.runnerID {
			__antithesis_instrumentation__.Notify(44603)
			t.L().PrintfCtxDepth(ctx, 3, "test status: %s", msg)
		} else {
			__antithesis_instrumentation__.Notify(44604)
			t.L().PrintfCtxDepth(ctx, 3, "test worker status: %s", msg)
		}
	} else {
		__antithesis_instrumentation__.Notify(44605)
	}
}

func (t *testImpl) Status(args ...interface{}) {
	__antithesis_instrumentation__.Notify(44606)
	t.status(context.TODO(), t.runnerID, args...)
}

func (t *testImpl) GetStatus() string {
	__antithesis_instrumentation__.Notify(44607)
	t.mu.Lock()
	defer t.mu.Unlock()
	status, ok := t.mu.status[t.runnerID]
	if ok {
		__antithesis_instrumentation__.Notify(44609)
		return fmt.Sprintf("%s (set %s ago)", status.msg, timeutil.Since(status.time).Round(time.Second))
	} else {
		__antithesis_instrumentation__.Notify(44610)
	}
	__antithesis_instrumentation__.Notify(44608)
	return "N/A"
}

func (t *testImpl) WorkerStatus(args ...interface{}) {
	__antithesis_instrumentation__.Notify(44611)
	t.status(context.TODO(), goid.Get(), args...)
}

func (t *testImpl) progress(id int64, frac float64) {
	__antithesis_instrumentation__.Notify(44612)
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.mu.status == nil {
		__antithesis_instrumentation__.Notify(44614)
		t.mu.status = make(map[int64]testStatus)
	} else {
		__antithesis_instrumentation__.Notify(44615)
	}
	__antithesis_instrumentation__.Notify(44613)
	status := t.mu.status[id]
	status.progress = frac
	t.mu.status[id] = status
}

func (t *testImpl) Progress(frac float64) {
	__antithesis_instrumentation__.Notify(44616)
	t.progress(t.runnerID, frac)
}

func (t *testImpl) WorkerProgress(frac float64) {
	__antithesis_instrumentation__.Notify(44617)
	t.progress(goid.Get(), frac)
}

var _ skip.SkippableTest = (*testImpl)(nil)

func (t *testImpl) Skip(args ...interface{}) {
	__antithesis_instrumentation__.Notify(44618)
	if len(args) > 0 {
		__antithesis_instrumentation__.Notify(44620)
		t.spec.Skip = fmt.Sprint(args[0])
		args = args[1:]
	} else {
		__antithesis_instrumentation__.Notify(44621)
	}
	__antithesis_instrumentation__.Notify(44619)
	t.spec.SkipDetails = fmt.Sprint(args...)
	panic(errTestFatal)
}

func (t *testImpl) Skipf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44622)
	t.spec.Skip = fmt.Sprintf(format, args...)
	panic(errTestFatal)
}

func (t *testImpl) Fatal(args ...interface{}) {
	__antithesis_instrumentation__.Notify(44623)
	t.markFailedInner("", args...)
	panic(errTestFatal)
}

func (t *testImpl) Fatalf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44624)
	t.markFailedInner(format, args...)
	panic(errTestFatal)
}

func (t *testImpl) FailNow() {
	__antithesis_instrumentation__.Notify(44625)
	t.Fatal()
}

func (t *testImpl) Errorf(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44626)
	t.markFailedInner(format, args...)
}

func (t *testImpl) markFailedInner(format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44627)

	if format != "" {
		__antithesis_instrumentation__.Notify(44628)
		t.printfAndFail(2, format, args...)
	} else {
		__antithesis_instrumentation__.Notify(44629)
		t.printAndFail(2, args...)
	}
}

func (t *testImpl) printAndFail(skip int, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44630)
	var msg string
	if len(args) == 1 {
		__antithesis_instrumentation__.Notify(44633)

		if err, ok := args[0].(error); ok {
			__antithesis_instrumentation__.Notify(44634)
			msg = fmt.Sprintf("%+v", err)
		} else {
			__antithesis_instrumentation__.Notify(44635)
		}
	} else {
		__antithesis_instrumentation__.Notify(44636)
	}
	__antithesis_instrumentation__.Notify(44631)
	if msg == "" {
		__antithesis_instrumentation__.Notify(44637)
		msg = fmt.Sprint(args...)
	} else {
		__antithesis_instrumentation__.Notify(44638)
	}
	__antithesis_instrumentation__.Notify(44632)
	t.failWithMsg(t.decorate(skip+1, msg))
}

func (t *testImpl) printfAndFail(skip int, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(44639)
	if format == "" {
		__antithesis_instrumentation__.Notify(44641)
		panic(fmt.Sprintf("invalid empty format. args: %s", args))
	} else {
		__antithesis_instrumentation__.Notify(44642)
	}
	__antithesis_instrumentation__.Notify(44640)
	t.failWithMsg(t.decorate(skip+1, fmt.Sprintf(format, args...)))
}

func (t *testImpl) failWithMsg(msg string) {
	__antithesis_instrumentation__.Notify(44643)
	t.mu.Lock()
	defer t.mu.Unlock()

	prefix := ""
	if t.mu.failed {
		__antithesis_instrumentation__.Notify(44645)
		prefix = "[not the first failure] "

		msg = "\n" + msg
	} else {
		__antithesis_instrumentation__.Notify(44646)
	}
	__antithesis_instrumentation__.Notify(44644)
	t.L().Printf("%stest failure: %s", prefix, msg)

	t.mu.failed = true
	t.mu.failureMsg += msg
	t.mu.output = append(t.mu.output, msg...)
	if t.mu.cancel != nil {
		__antithesis_instrumentation__.Notify(44647)
		t.mu.cancel()
	} else {
		__antithesis_instrumentation__.Notify(44648)
	}
}

func (t *testImpl) decorate(skip int, s string) string {
	__antithesis_instrumentation__.Notify(44649)

	var pc [50]uintptr
	n := runtime.Callers(2+skip, pc[:])
	if n == 0 {
		__antithesis_instrumentation__.Notify(44654)
		panic("zero callers found")
	} else {
		__antithesis_instrumentation__.Notify(44655)
	}
	__antithesis_instrumentation__.Notify(44650)

	buf := new(bytes.Buffer)
	frames := runtime.CallersFrames(pc[:n])
	sep := "\t"
	runnerFound := false
	for {
		__antithesis_instrumentation__.Notify(44656)
		if runnerFound {
			__antithesis_instrumentation__.Notify(44662)
			break
		} else {
			__antithesis_instrumentation__.Notify(44663)
		}
		__antithesis_instrumentation__.Notify(44657)

		frame, more := frames.Next()
		if !more {
			__antithesis_instrumentation__.Notify(44664)
			break
		} else {
			__antithesis_instrumentation__.Notify(44665)
		}
		__antithesis_instrumentation__.Notify(44658)
		if frame.Function == t.runner {
			__antithesis_instrumentation__.Notify(44666)
			runnerFound = true

			if t.mu.failLoc.file == "" {
				__antithesis_instrumentation__.Notify(44667)
				t.mu.failLoc.file = frame.File
				t.mu.failLoc.line = frame.Line
			} else {
				__antithesis_instrumentation__.Notify(44668)
			}
		} else {
			__antithesis_instrumentation__.Notify(44669)
		}
		__antithesis_instrumentation__.Notify(44659)
		if !t.mu.failed && func() bool {
			__antithesis_instrumentation__.Notify(44670)
			return !runnerFound == true
		}() == true {
			__antithesis_instrumentation__.Notify(44671)

			t.mu.failLoc.file = frame.File
			t.mu.failLoc.line = frame.Line
		} else {
			__antithesis_instrumentation__.Notify(44672)
		}
		__antithesis_instrumentation__.Notify(44660)
		file := frame.File
		if index := strings.LastIndexByte(file, '/'); index >= 0 {
			__antithesis_instrumentation__.Notify(44673)
			file = file[index+1:]
		} else {
			__antithesis_instrumentation__.Notify(44674)
		}
		__antithesis_instrumentation__.Notify(44661)
		fmt.Fprintf(buf, "%s%s:%d", sep, file, frame.Line)
		sep = ","
	}
	__antithesis_instrumentation__.Notify(44651)
	buf.WriteString(": ")

	lines := strings.Split(s, "\n")
	if l := len(lines); l > 1 && func() bool {
		__antithesis_instrumentation__.Notify(44675)
		return lines[l-1] == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(44676)
		lines = lines[:l-1]
	} else {
		__antithesis_instrumentation__.Notify(44677)
	}
	__antithesis_instrumentation__.Notify(44652)
	for i, line := range lines {
		__antithesis_instrumentation__.Notify(44678)
		if i > 0 {
			__antithesis_instrumentation__.Notify(44680)
			buf.WriteString("\n\t\t")
		} else {
			__antithesis_instrumentation__.Notify(44681)
		}
		__antithesis_instrumentation__.Notify(44679)
		buf.WriteString(line)
	}
	__antithesis_instrumentation__.Notify(44653)
	buf.WriteByte('\n')
	return buf.String()
}

func (t *testImpl) duration() time.Duration {
	__antithesis_instrumentation__.Notify(44682)
	return t.end.Sub(t.start)
}

func (t *testImpl) Failed() bool {
	__antithesis_instrumentation__.Notify(44683)
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mu.failed
}

func (t *testImpl) FailureMsg() string {
	__antithesis_instrumentation__.Notify(44684)
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.mu.failureMsg
}

func (t *testImpl) ArtifactsDir() string {
	__antithesis_instrumentation__.Notify(44685)
	return t.artifactsDir
}

func (t *testImpl) PerfArtifactsDir() string {
	__antithesis_instrumentation__.Notify(44686)
	return perfArtifactsDir
}

func (t *testImpl) IsBuildVersion(minVersion string) bool {
	__antithesis_instrumentation__.Notify(44687)
	vers, err := version.Parse(minVersion)
	if err != nil {
		__antithesis_instrumentation__.Notify(44690)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(44691)
	}
	__antithesis_instrumentation__.Notify(44688)
	if p := vers.PreRelease(); p != "" {
		__antithesis_instrumentation__.Notify(44692)
		panic("cannot specify a prerelease: " + p)
	} else {
		__antithesis_instrumentation__.Notify(44693)
	}
	__antithesis_instrumentation__.Notify(44689)

	vers = version.MustParse(minVersion + "-0")
	return t.BuildVersion().AtLeast(vers)
}

func teamCityEscape(s string) string {
	__antithesis_instrumentation__.Notify(44694)
	r := strings.NewReplacer(
		"\n", "|n",
		"'", "|'",
		"|", "||",
		"[", "|[",
		"]", "|]",
	)
	return r.Replace(s)
}

func teamCityNameEscape(name string) string {
	__antithesis_instrumentation__.Notify(44695)
	return strings.Replace(name, ",", "_", -1)
}

type testWithCount struct {
	spec registry.TestSpec

	count int
}

type clusterType int

const (
	localCluster clusterType = iota
	roachprodCluster
)

type loggingOpt struct {
	l *logger.Logger

	tee            logger.TeeOptType
	stdout, stderr io.Writer

	artifactsDir string

	literalArtifactsDir string

	runnerLogPath string
}

type workerStatus struct {
	name string
	mu   struct {
		syncutil.Mutex

		status string

		ttr testToRunRes
		t   *testImpl
		c   *clusterImpl
	}
}

func (w *workerStatus) Status() string {
	__antithesis_instrumentation__.Notify(44696)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.mu.status
}

func (w *workerStatus) SetStatus(status string) {
	__antithesis_instrumentation__.Notify(44697)
	w.mu.Lock()
	w.mu.status = status
	w.mu.Unlock()
}

func (w *workerStatus) Cluster() *clusterImpl {
	__antithesis_instrumentation__.Notify(44698)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.mu.c
}

func (w *workerStatus) SetCluster(c *clusterImpl) {
	__antithesis_instrumentation__.Notify(44699)
	w.mu.Lock()
	w.mu.c = c
	w.mu.Unlock()
}

func (w *workerStatus) TestToRun() testToRunRes {
	__antithesis_instrumentation__.Notify(44700)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.mu.ttr
}

func (w *workerStatus) Test() *testImpl {
	__antithesis_instrumentation__.Notify(44701)
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.mu.t
}

func (w *workerStatus) SetTest(t *testImpl, ttr testToRunRes) {
	__antithesis_instrumentation__.Notify(44702)
	w.mu.Lock()
	w.mu.t = t
	w.mu.ttr = ttr
	w.mu.Unlock()
}

func shout(
	ctx context.Context, l *logger.Logger, stdout io.Writer, format string, args ...interface{},
) {
	__antithesis_instrumentation__.Notify(44703)
	if len(format) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(44705)
		return format[len(format)-1] != '\n' == true
	}() == true {
		__antithesis_instrumentation__.Notify(44706)
		format += "\n"
	} else {
		__antithesis_instrumentation__.Notify(44707)
	}
	__antithesis_instrumentation__.Notify(44704)
	msg := fmt.Sprintf(format, args...)
	l.PrintfCtxDepth(ctx, 2, msg)
	fmt.Fprint(stdout, msg)
}
