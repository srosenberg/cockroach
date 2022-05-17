//go:build gofuzz
// +build gofuzz

package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/util/timeutil"

var (
	timeCtx = NewParseTimeContext(timeutil.Now())
)

func FuzzParseDDecimal(data []byte) int {
	__antithesis_instrumentation__.Notify(609908)
	_, err := ParseDDecimal(string(data))
	if err != nil {
		__antithesis_instrumentation__.Notify(609910)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(609911)
	}
	__antithesis_instrumentation__.Notify(609909)
	return 1
}

func FuzzParseDDate(data []byte) int {
	__antithesis_instrumentation__.Notify(609912)
	_, _, err := ParseDDate(timeCtx, string(data))
	if err != nil {
		__antithesis_instrumentation__.Notify(609914)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(609915)
	}
	__antithesis_instrumentation__.Notify(609913)
	return 1
}
