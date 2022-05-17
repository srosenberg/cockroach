package a

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"testing"

	"github.com/cockroachdb/cockroach/pkg/util/leaktest"
)

func foo(t *testing.T) {
	__antithesis_instrumentation__.Notify(644867)
	defer leaktest.AfterTest(t)
}
