package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"os"
	"runtime"
	"runtime/pprof"
	"testing"
)

func WriteProfile(t testing.TB, name string, path string) {
	__antithesis_instrumentation__.Notify(645903)
	f, err := os.Create(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(645905)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645906)
	}
	__antithesis_instrumentation__.Notify(645904)
	defer f.Close()
	if err := pprof.Lookup(name).WriteTo(f, 0); err != nil {
		__antithesis_instrumentation__.Notify(645907)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(645908)
	}
}

func AllocProfileDiff(t testing.TB, beforePath, afterPath string, fn func()) {
	__antithesis_instrumentation__.Notify(645909)

	runtime.GC()
	WriteProfile(t, "allocs", beforePath)
	fn()
	runtime.GC()
	WriteProfile(t, "allocs", afterPath)
	t.Logf("to use your alloc profiles: go tool pprof -base %s %s", beforePath, afterPath)
}

var _ = WriteProfile
var _ = AllocProfileDiff
