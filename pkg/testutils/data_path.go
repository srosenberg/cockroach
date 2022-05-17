package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"path"
	"path/filepath"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/build/bazel"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/stretchr/testify/require"
)

func TestDataPath(t testing.TB, relative ...string) string {
	__antithesis_instrumentation__.Notify(644046)
	relative = append([]string{"testdata"}, relative...)

	if bazel.BuiltWithBazel() {
		__antithesis_instrumentation__.Notify(644048)

		cockroachWorkspace, set := envutil.EnvString("COCKROACH_WORKSPACE", 0)
		if set {
			__antithesis_instrumentation__.Notify(644050)
			return path.Join(cockroachWorkspace, bazel.RelativeTestTargetPath(), path.Join(relative...))
		} else {
			__antithesis_instrumentation__.Notify(644051)
		}
		__antithesis_instrumentation__.Notify(644049)
		runfiles, err := bazel.RunfilesPath()
		require.NoError(t, err)
		return path.Join(runfiles, bazel.RelativeTestTargetPath(), path.Join(relative...))
	} else {
		__antithesis_instrumentation__.Notify(644052)
	}
	__antithesis_instrumentation__.Notify(644047)

	ret := path.Join(relative...)
	ret, err := filepath.Abs(ret)
	require.NoError(t, err)
	return ret
}
