package leaktest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "testing"

func AfterTest(testing.TB) func() {
	__antithesis_instrumentation__.Notify(644868)
	return func() { __antithesis_instrumentation__.Notify(644869) }
}
