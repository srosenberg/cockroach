//go:build gofuzz
// +build gofuzz

package hba

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/kr/pretty"
)

func FuzzParseAndNormalize(data []byte) int {
	__antithesis_instrumentation__.Notify(559833)
	conf, err := ParseAndNormalize(string(data))
	if err != nil {
		__antithesis_instrumentation__.Notify(559837)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(559838)
	}
	__antithesis_instrumentation__.Notify(559834)
	s := conf.String()
	conf2, err := ParseAndNormalize(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(559839)
		panic(fmt.Errorf(`-- original:
%s
-- parsed:
%# v
-- new:
%s
-- error:
%v`,
			string(data),
			pretty.Formatter(conf),
			s,
			err))
	} else {
		__antithesis_instrumentation__.Notify(559840)
	}
	__antithesis_instrumentation__.Notify(559835)
	s2 := conf2.String()
	if s != s2 {
		__antithesis_instrumentation__.Notify(559841)
		panic(fmt.Errorf(`reparse mismatch:
-- original:
%s
-- new:
%# v
-- printed:
%s`,
			s,
			pretty.Formatter(conf2),
			s2))
	} else {
		__antithesis_instrumentation__.Notify(559842)
	}
	__antithesis_instrumentation__.Notify(559836)
	return 1
}
