package clisqlclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type InitialSQLConnectionError struct {
	err error
}

func (i *InitialSQLConnectionError) Error() string {
	__antithesis_instrumentation__.Notify(28810)
	return i.err.Error()
}

func (i *InitialSQLConnectionError) Cause() error {
	__antithesis_instrumentation__.Notify(28811)
	return i.err
}

func (i *InitialSQLConnectionError) Unwrap() error {
	__antithesis_instrumentation__.Notify(28812)
	return i.err
}

func (i *InitialSQLConnectionError) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(28813)
	errors.FormatError(i, s, verb)
}

func (i *InitialSQLConnectionError) FormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(28814)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(28816)
		p.Print("error while establishing the SQL session")
	} else {
		__antithesis_instrumentation__.Notify(28817)
	}
	__antithesis_instrumentation__.Notify(28815)
	return i.err
}
