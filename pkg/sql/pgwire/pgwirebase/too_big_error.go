package pgwirebase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strconv"

	"github.com/cockroachdb/errors"
)

type withMessageTooBig struct {
	cause error
	size  int
}

var _ error = (*withMessageTooBig)(nil)
var _ errors.SafeDetailer = (*withMessageTooBig)(nil)

func (w *withMessageTooBig) Error() string {
	__antithesis_instrumentation__.Notify(561435)
	return w.cause.Error()
}
func (w *withMessageTooBig) Unwrap() error {
	__antithesis_instrumentation__.Notify(561436)
	return w.cause
}
func (w *withMessageTooBig) SafeDetails() []string {
	__antithesis_instrumentation__.Notify(561437)
	return []string{strconv.Itoa(w.size)}
}

func withMessageTooBigError(err error, size int) error {
	__antithesis_instrumentation__.Notify(561438)
	if err == nil {
		__antithesis_instrumentation__.Notify(561440)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(561441)
	}
	__antithesis_instrumentation__.Notify(561439)

	return &withMessageTooBig{cause: err, size: size}
}

func IsMessageTooBigError(err error) bool {
	__antithesis_instrumentation__.Notify(561442)
	var c withMessageTooBig
	return errors.HasType(err, &c)
}

func GetMessageTooBigSize(err error) int {
	__antithesis_instrumentation__.Notify(561443)
	if c := (*withMessageTooBig)(nil); errors.As(err, &c) {
		__antithesis_instrumentation__.Notify(561445)
		return c.size
	} else {
		__antithesis_instrumentation__.Notify(561446)
	}
	__antithesis_instrumentation__.Notify(561444)
	return -1
}
