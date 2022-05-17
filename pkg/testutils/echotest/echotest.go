package echotest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"testing"

	"github.com/cockroachdb/datadriven"
)

func Require(t *testing.T, act, path string) {
	__antithesis_instrumentation__.Notify(644193)
	var ran bool
	datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
		__antithesis_instrumentation__.Notify(644195)
		if d.Cmd != "echo" {
			__antithesis_instrumentation__.Notify(644197)
			return "only 'echo' is supported"
		} else {
			__antithesis_instrumentation__.Notify(644198)
		}
		__antithesis_instrumentation__.Notify(644196)
		ran = true
		return act
	})
	__antithesis_instrumentation__.Notify(644194)
	if !ran {
		__antithesis_instrumentation__.Notify(644199)

		t.Errorf("no tests run for %s, is the file empty?", path)
	} else {
		__antithesis_instrumentation__.Notify(644200)
	}
}
