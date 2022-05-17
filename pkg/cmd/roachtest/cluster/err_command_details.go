package cluster

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type WithCommandDetails struct {
	Wrapped error
	Cmd     string
	Stderr  string
	Stdout  string
}

var _ error = (*WithCommandDetails)(nil)
var _ errors.Formatter = (*WithCommandDetails)(nil)

func (e *WithCommandDetails) Error() string {
	__antithesis_instrumentation__.Notify(43229)
	return e.Wrapped.Error()
}

func (e *WithCommandDetails) Cause() error {
	__antithesis_instrumentation__.Notify(43230)
	return e.Wrapped
}

func (e *WithCommandDetails) Format(s fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(43231)
	errors.FormatError(e, s, verb)
}

func (e *WithCommandDetails) FormatError(p errors.Printer) error {
	__antithesis_instrumentation__.Notify(43232)
	p.Printf("%s returned", e.Cmd)
	if p.Detail() {
		__antithesis_instrumentation__.Notify(43234)
		p.Printf("stderr:\n%s\nstdout:\n%s", e.Stderr, e.Stdout)
	} else {
		__antithesis_instrumentation__.Notify(43235)
	}
	__antithesis_instrumentation__.Notify(43233)
	return e.Wrapped
}

func GetStderr(err error) string {
	__antithesis_instrumentation__.Notify(43236)
	var c *WithCommandDetails
	if errors.As(err, &c) {
		__antithesis_instrumentation__.Notify(43238)
		return c.Stderr
	} else {
		__antithesis_instrumentation__.Notify(43239)
	}
	__antithesis_instrumentation__.Notify(43237)
	return ""
}
