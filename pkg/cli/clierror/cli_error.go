package clierror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cli/exit"
	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/errors"
)

type Error struct {
	exitCode exit.Code
	severity logpb.Severity
	cause    error
}

func NewError(cause error, exitCode exit.Code) error {
	__antithesis_instrumentation__.Notify(28281)
	return NewErrorWithSeverity(cause, exitCode, severity.UNKNOWN)
}

func NewErrorWithSeverity(cause error, exitCode exit.Code, severity logpb.Severity) error {
	__antithesis_instrumentation__.Notify(28282)
	return &Error{
		exitCode: exitCode,
		severity: severity,
		cause:    cause,
	}
}

func (e *Error) GetExitCode() exit.Code {
	__antithesis_instrumentation__.Notify(28283)
	return e.exitCode
}

func (e *Error) GetSeverity() logpb.Severity {
	__antithesis_instrumentation__.Notify(28284)
	return e.severity
}

func (e *Error) Error() string { __antithesis_instrumentation__.Notify(28285); return e.cause.Error() }

func (e *Error) Cause() error { __antithesis_instrumentation__.Notify(28286); return e.cause }

func (e *Error) Unwrap() error { __antithesis_instrumentation__.Notify(28287); return e.cause }

func (e *Error) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(28288)
	errors.FormatError(e, s, verb)
}

func (e *Error) FormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(28289)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(28291)
		p.Printf("error with exit code: %d", e.exitCode)
	} else {
		__antithesis_instrumentation__.Notify(28292)
	}
	__antithesis_instrumentation__.Notify(28290)
	return e.cause
}
