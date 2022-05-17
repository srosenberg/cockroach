package clisqlshell

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strings"

type StatementsValue []string

func (s *StatementsValue) Type() string {
	__antithesis_instrumentation__.Notify(30229)
	return "<stmtlist>"
}

func (s *StatementsValue) String() string {
	__antithesis_instrumentation__.Notify(30230)
	return strings.Join(*s, ";")
}

func (s *StatementsValue) Set(value string) error {
	__antithesis_instrumentation__.Notify(30231)
	*s = append(*s, value)
	return nil
}
