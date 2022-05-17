package errors

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os/exec"

	"github.com/cockroachdb/errors"
)

type Error interface {
	error

	ExitCode() int
}

const (
	cmdExitCode          = 20
	sshExitCode          = 10
	unclassifiedExitCode = 1
)

type Cmd struct {
	Err error
}

func (e Cmd) Error() string {
	__antithesis_instrumentation__.Notify(180406)
	return fmt.Sprintf("COMMAND_PROBLEM: %s", e.Err.Error())
}

func (e Cmd) ExitCode() int {
	__antithesis_instrumentation__.Notify(180407)
	return cmdExitCode
}

func (e Cmd) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(180408)
	errors.FormatError(e, s, verb)
}

func (e Cmd) Unwrap() error {
	__antithesis_instrumentation__.Notify(180409)
	return e.Err
}

type SSH struct {
	Err error
}

func (e SSH) Error() string {
	__antithesis_instrumentation__.Notify(180410)
	return fmt.Sprintf("SSH_PROBLEM: %s", e.Err.Error())
}

func (e SSH) ExitCode() int {
	__antithesis_instrumentation__.Notify(180411)
	return sshExitCode
}

func (e SSH) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(180412)
	errors.FormatError(e, s, verb)
}

func (e SSH) Unwrap() error {
	__antithesis_instrumentation__.Notify(180413)
	return e.Err
}

type Unclassified struct {
	Err error
}

func (e Unclassified) Error() string {
	__antithesis_instrumentation__.Notify(180414)
	return fmt.Sprintf("UNCLASSIFIED_PROBLEM: %s", e.Err.Error())
}

func (e Unclassified) ExitCode() int {
	__antithesis_instrumentation__.Notify(180415)
	return unclassifiedExitCode
}

func (e Unclassified) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(180416)
	errors.FormatError(e, s, verb)
}

func (e Unclassified) Unwrap() error {
	__antithesis_instrumentation__.Notify(180417)
	return e.Err
}

func ClassifyCmdError(err error) Error {
	__antithesis_instrumentation__.Notify(180418)
	if err == nil {
		__antithesis_instrumentation__.Notify(180421)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(180422)
	}
	__antithesis_instrumentation__.Notify(180419)

	if exitErr, ok := asExitError(err); ok {
		__antithesis_instrumentation__.Notify(180423)
		if exitErr.ExitCode() == 255 {
			__antithesis_instrumentation__.Notify(180425)
			return SSH{err}
		} else {
			__antithesis_instrumentation__.Notify(180426)
		}
		__antithesis_instrumentation__.Notify(180424)
		return Cmd{err}
	} else {
		__antithesis_instrumentation__.Notify(180427)
	}
	__antithesis_instrumentation__.Notify(180420)

	return Unclassified{err}
}

func asExitError(err error) (*exec.ExitError, bool) {
	__antithesis_instrumentation__.Notify(180428)
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		__antithesis_instrumentation__.Notify(180430)
		return exitErr, true
	} else {
		__antithesis_instrumentation__.Notify(180431)
	}
	__antithesis_instrumentation__.Notify(180429)
	return nil, false
}

func AsError(err error) (Error, bool) {
	__antithesis_instrumentation__.Notify(180432)
	var e Error
	if errors.As(err, &e) {
		__antithesis_instrumentation__.Notify(180434)
		return e, true
	} else {
		__antithesis_instrumentation__.Notify(180435)
	}
	__antithesis_instrumentation__.Notify(180433)
	return nil, false
}

func SelectPriorityError(errors []error) error {
	__antithesis_instrumentation__.Notify(180436)
	var result Error
	for _, err := range errors {
		__antithesis_instrumentation__.Notify(180440)
		if err == nil {
			__antithesis_instrumentation__.Notify(180443)
			continue
		} else {
			__antithesis_instrumentation__.Notify(180444)
		}
		__antithesis_instrumentation__.Notify(180441)

		rpErr, _ := AsError(err)
		if result == nil {
			__antithesis_instrumentation__.Notify(180445)
			result = rpErr
			continue
		} else {
			__antithesis_instrumentation__.Notify(180446)
		}
		__antithesis_instrumentation__.Notify(180442)

		if rpErr.ExitCode() > result.ExitCode() {
			__antithesis_instrumentation__.Notify(180447)
			result = rpErr
		} else {
			__antithesis_instrumentation__.Notify(180448)
		}
	}
	__antithesis_instrumentation__.Notify(180437)

	if result != nil {
		__antithesis_instrumentation__.Notify(180449)
		return result
	} else {
		__antithesis_instrumentation__.Notify(180450)
	}
	__antithesis_instrumentation__.Notify(180438)

	for _, err := range errors {
		__antithesis_instrumentation__.Notify(180451)
		if err != nil {
			__antithesis_instrumentation__.Notify(180452)
			return err
		} else {
			__antithesis_instrumentation__.Notify(180453)
		}
	}
	__antithesis_instrumentation__.Notify(180439)
	return nil
}

var _ = SelectPriorityError
