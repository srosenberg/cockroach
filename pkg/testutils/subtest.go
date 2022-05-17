package testutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"testing"
)

func RunTrueAndFalse(t *testing.T, name string, fn func(t *testing.T, b bool)) {
	__antithesis_instrumentation__.Notify(646440)
	for _, b := range []bool{false, true} {
		__antithesis_instrumentation__.Notify(646441)
		t.Run(fmt.Sprintf("%s=%t", name, b), func(t *testing.T) {
			__antithesis_instrumentation__.Notify(646442)
			fn(t, b)
		})
	}
}

func RunValues(t *testing.T, name string, values []interface{}, fn func(*testing.T, interface{})) {
	__antithesis_instrumentation__.Notify(646443)
	t.Helper()
	for _, v := range values {
		__antithesis_instrumentation__.Notify(646444)
		t.Run(fmt.Sprintf("%s=%v", name, v), func(t *testing.T) {
			__antithesis_instrumentation__.Notify(646445)
			fn(t, v)
		})
	}
}
