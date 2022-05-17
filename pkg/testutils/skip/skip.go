package skip

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"flag"
	"fmt"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/buildutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type SkippableTest interface {
	Helper()
	Skip(...interface{})
	Skipf(string, ...interface{})
}

func WithIssue(t SkippableTest, githubIssueID int, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645984)
	t.Helper()
	t.Skip(append([]interface{}{
		fmt.Sprintf("https://github.com/cockroachdb/cockroach/issues/%d", githubIssueID)},
		args...))
}

func IgnoreLint(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645985)
	t.Helper()
	t.Skip(args...)
}

func IgnoreLintf(t SkippableTest, format string, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645986)
	t.Helper()
	t.Skipf(format, args...)
}

func UnderDeadlock(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645987)
	t.Helper()
	if syncutil.DeadlockEnabled {
		__antithesis_instrumentation__.Notify(645988)
		t.Skip(append([]interface{}{"disabled under deadlock detector"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(645989)
	}
}

func UnderDeadlockWithIssue(t SkippableTest, githubIssueID int, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645990)
	t.Helper()
	if syncutil.DeadlockEnabled {
		__antithesis_instrumentation__.Notify(645991)
		t.Skip(append([]interface{}{fmt.Sprintf(
			"disabled under deadlock detector. issue: https://github.com/cockroachdb/cockroach/issue/%d",
			githubIssueID,
		)}, args...))
	} else {
		__antithesis_instrumentation__.Notify(645992)
	}
}

func UnderRace(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645993)
	t.Helper()
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(645994)
		t.Skip(append([]interface{}{"disabled under race"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(645995)
	}
}

func UnderRaceWithIssue(t SkippableTest, githubIssueID int, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645996)
	t.Helper()
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(645997)
		t.Skip(append([]interface{}{fmt.Sprintf(
			"disabled under race. issue: https://github.com/cockroachdb/cockroach/issue/%d", githubIssueID,
		)}, args...))
	} else {
		__antithesis_instrumentation__.Notify(645998)
	}
}

func UnderBazelWithIssue(t SkippableTest, githubIssueID int, args ...interface{}) {
	__antithesis_instrumentation__.Notify(645999)
	t.Helper()
	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(646000)
		t.Skip(append([]interface{}{fmt.Sprintf(
			"disabled under bazel. issue: https://github.com/cockroachdb/cockroach/issue/%d", githubIssueID,
		)}, args...))
	} else {
		__antithesis_instrumentation__.Notify(646001)
	}
}

var _ = UnderBazelWithIssue

func UnderShort(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646002)
	t.Helper()
	if testing.Short() {
		__antithesis_instrumentation__.Notify(646003)
		t.Skip(append([]interface{}{"disabled under -short"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(646004)
	}
}

func UnderStress(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646005)
	t.Helper()
	if NightlyStress() {
		__antithesis_instrumentation__.Notify(646006)
		t.Skip(append([]interface{}{"disabled under stress"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(646007)
	}
}

func UnderStressRace(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646008)
	t.Helper()
	if NightlyStress() && func() bool {
		__antithesis_instrumentation__.Notify(646009)
		return util.RaceEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(646010)
		t.Skip(append([]interface{}{"disabled under stressrace"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(646011)
	}
}

func UnderMetamorphic(t SkippableTest, args ...interface{}) {
	__antithesis_instrumentation__.Notify(646012)
	t.Helper()
	if util.IsMetamorphicBuild() {
		__antithesis_instrumentation__.Notify(646013)
		t.Skip(append([]interface{}{"disabled under metamorphic"}, args...))
	} else {
		__antithesis_instrumentation__.Notify(646014)
	}
}

func UnderNonTestBuild(t SkippableTest) {
	__antithesis_instrumentation__.Notify(646015)
	if !buildutil.CrdbTestBuild {
		__antithesis_instrumentation__.Notify(646016)
		t.Skip("crdb_test tag required for this test")
	} else {
		__antithesis_instrumentation__.Notify(646017)
	}
}

func UnderBench() bool {
	__antithesis_instrumentation__.Notify(646018)

	f := flag.Lookup("test.bench")
	return f != nil && func() bool {
		__antithesis_instrumentation__.Notify(646019)
		return f.Value.String() != "" == true
	}() == true
}
