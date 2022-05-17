//go:build !acceptance
// +build !acceptance

package acceptance

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"testing"
)

func MainTest(m *testing.M) {
	__antithesis_instrumentation__.Notify(1156)
	fmt.Fprintln(os.Stderr, "not running with `acceptance` build tag; skipping")
}
