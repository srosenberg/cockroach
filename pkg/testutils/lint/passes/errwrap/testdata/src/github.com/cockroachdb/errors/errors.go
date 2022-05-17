package errors

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

func New(msg string) error {
	__antithesis_instrumentation__.Notify(644581)
	return fmt.Errorf(msg)
}

func Newf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644582)
	return fmt.Errorf(format, args...)
}

func Wrap(_ error, msg string) error {
	__antithesis_instrumentation__.Notify(644583)
	return fmt.Errorf(msg)
}

func Wrapf(_ error, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644584)
	return fmt.Errorf(format, args...)
}

func WrapWithDepth(depth int, err error, msg string) error {
	__antithesis_instrumentation__.Notify(644585)
	return fmt.Errorf(msg)
}

func WrapWithDepthf(depth int, err error, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644586)
	return fmt.Errorf(format, args...)
}

func NewWithDepth(_ int, msg string) error {
	__antithesis_instrumentation__.Notify(644587)
	return fmt.Errorf(msg)
}

func NewWithDepthf(_ int, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644588)
	return fmt.Errorf(format, args...)
}

func AssertionFailedf(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644589)
	return fmt.Errorf(format, args...)
}

func AssertionFailedWithDepthf(_ int, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644590)
	return fmt.Errorf(format, args...)
}

func NewAssertionErrorWithWrappedErrf(_ error, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(644591)
	return fmt.Errorf(format, args...)
}
