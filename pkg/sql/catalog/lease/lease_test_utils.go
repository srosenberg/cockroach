package lease

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func TestingTableLeasesAreDisabled() bool {
	__antithesis_instrumentation__.Notify(266582)
	return testDisableTableLeases
}

var testDisableTableLeases bool

func TestingDisableTableLeases() func() {
	__antithesis_instrumentation__.Notify(266583)
	testDisableTableLeases = true
	return func() {
		__antithesis_instrumentation__.Notify(266584)
		testDisableTableLeases = false
	}
}
