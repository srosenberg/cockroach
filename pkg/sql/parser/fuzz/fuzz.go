//go:build gofuzz
// +build gofuzz

package fuzz

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/parser"

	_ "github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
)

func FuzzParse(data []byte) int {
	__antithesis_instrumentation__.Notify(552295)
	_, err := parser.Parse(string(data))
	if err != nil {
		__antithesis_instrumentation__.Notify(552297)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(552298)
	}
	__antithesis_instrumentation__.Notify(552296)
	return 1
}
