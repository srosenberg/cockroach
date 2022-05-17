package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/util/fileutil"
)

func TempDir(t testing.TB) (string, func()) {
	__antithesis_instrumentation__.Notify(644069)
	tmpDir := ""
	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(644073)

		tmpDir = bazel.TestTmpDir()
	} else {
		__antithesis_instrumentation__.Notify(644074)
	}
	__antithesis_instrumentation__.Notify(644070)

	dir, err := ioutil.TempDir(tmpDir, fileutil.EscapeFilename(t.Name()))
	if err != nil {
		__antithesis_instrumentation__.Notify(644075)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(644076)
	}
	__antithesis_instrumentation__.Notify(644071)
	cleanup := func() {
		__antithesis_instrumentation__.Notify(644077)
		if err := os.RemoveAll(dir); err != nil {
			__antithesis_instrumentation__.Notify(644078)
			t.Error(err)
		} else {
			__antithesis_instrumentation__.Notify(644079)
		}
	}
	__antithesis_instrumentation__.Notify(644072)
	return dir, cleanup
}
