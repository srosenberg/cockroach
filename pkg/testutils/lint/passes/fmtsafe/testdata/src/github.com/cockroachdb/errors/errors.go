package errors

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

func New(msg string) error {
	__antithesis_instrumentation__.Notify(644693)
	return fmt.Errorf("abc")
}

func Newf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644694)
	return fmt.Errorf(format, args...)
}
