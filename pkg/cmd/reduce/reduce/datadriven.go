package reduce

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"log"
	"os"
	"testing"

	"github.com/cockroachdb/datadriven"
)

func Walk(
	t *testing.T,
	path string,
	filter func(string) (string, error),
	interesting func(contains string) InterestingFn,
	mode Mode,
	cr ChunkReducer,
	passes []Pass,
) {
	__antithesis_instrumentation__.Notify(41809)
	datadriven.Walk(t, path, func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(41810)
		RunTest(t, path, filter, interesting, mode, cr, passes)
	})
}

func RunTest(
	t *testing.T,
	path string,
	filter func(string) (string, error),
	interesting func(contains string) InterestingFn,
	mode Mode,
	cr ChunkReducer,
	passes []Pass,
) {
	__antithesis_instrumentation__.Notify(41811)
	var contains string
	var logger *log.Logger
	if testing.Verbose() {
		__antithesis_instrumentation__.Notify(41813)
		logger = log.New(os.Stderr, "", 0)
	} else {
		__antithesis_instrumentation__.Notify(41814)
	}
	__antithesis_instrumentation__.Notify(41812)
	datadriven.RunTest(t, path, func(t *testing.T, d *datadriven.TestData) string {
		__antithesis_instrumentation__.Notify(41815)
		switch d.Cmd {
		case "contains":
			__antithesis_instrumentation__.Notify(41816)
			contains = d.Input
			return ""
		case "reduce":
			__antithesis_instrumentation__.Notify(41817)
			input := d.Input
			if filter != nil {
				__antithesis_instrumentation__.Notify(41821)
				var err error
				input, err = filter(input)
				if err != nil {
					__antithesis_instrumentation__.Notify(41822)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(41823)
				}
			} else {
				__antithesis_instrumentation__.Notify(41824)
			}
			__antithesis_instrumentation__.Notify(41818)
			output, err := Reduce(logger, input, interesting(contains), 0, mode, cr, passes...)
			if err != nil {
				__antithesis_instrumentation__.Notify(41825)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(41826)
			}
			__antithesis_instrumentation__.Notify(41819)
			return output + "\n"
		default:
			__antithesis_instrumentation__.Notify(41820)
			t.Fatalf("unknown command: %s", d.Cmd)
			return ""
		}
	})
}
