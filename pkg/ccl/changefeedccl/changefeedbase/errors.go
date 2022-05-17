package changefeedbase

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs/joberror"
	"github.com/cockroachdb/errors"
)

type FailureType = string

const (
	ConnectionClosed FailureType = "connection_closed"

	UserInput FailureType = "user_input"

	OnStartup FailureType = "on_startup"

	UnknownError FailureType = "unknown_error"
)

type taggedError struct {
	wrapped error
	tag     string
}

func MarkTaggedError(e error, tag string) error {
	__antithesis_instrumentation__.Notify(16588)
	return &taggedError{wrapped: e, tag: tag}
}

func IsTaggedError(err error) (bool, string) {
	__antithesis_instrumentation__.Notify(16589)
	if err == nil {
		__antithesis_instrumentation__.Notify(16592)
		return false, ""
	} else {
		__antithesis_instrumentation__.Notify(16593)
	}
	__antithesis_instrumentation__.Notify(16590)
	if tagged := (*taggedError)(nil); errors.As(err, &tagged) {
		__antithesis_instrumentation__.Notify(16594)
		return true, tagged.tag
	} else {
		__antithesis_instrumentation__.Notify(16595)
	}
	__antithesis_instrumentation__.Notify(16591)
	return false, ""
}

func (e *taggedError) Error() string {
	__antithesis_instrumentation__.Notify(16596)
	return e.wrapped.Error()
}

func (e *taggedError) Cause() error { __antithesis_instrumentation__.Notify(16597); return e.wrapped }

func (e *taggedError) Unwrap() error { __antithesis_instrumentation__.Notify(16598); return e.wrapped }

const retryableErrorString = "retryable changefeed error"

type retryableError struct {
	wrapped error
}

func MarkRetryableError(e error) error {
	__antithesis_instrumentation__.Notify(16599)
	return &retryableError{wrapped: e}
}

func (e *retryableError) Error() string {
	__antithesis_instrumentation__.Notify(16600)
	return fmt.Sprintf("%s: %s", retryableErrorString, e.wrapped.Error())
}

func (e *retryableError) Cause() error {
	__antithesis_instrumentation__.Notify(16601)
	return e.wrapped
}

func (e *retryableError) Unwrap() error {
	__antithesis_instrumentation__.Notify(16602)
	return e.wrapped
}

func IsRetryableError(err error) bool {
	__antithesis_instrumentation__.Notify(16603)
	if err == nil {
		__antithesis_instrumentation__.Notify(16607)
		return false
	} else {
		__antithesis_instrumentation__.Notify(16608)
	}
	__antithesis_instrumentation__.Notify(16604)
	if errors.HasType(err, (*retryableError)(nil)) {
		__antithesis_instrumentation__.Notify(16609)
		return true
	} else {
		__antithesis_instrumentation__.Notify(16610)
	}
	__antithesis_instrumentation__.Notify(16605)

	errStr := err.Error()
	if strings.Contains(errStr, retryableErrorString) {
		__antithesis_instrumentation__.Notify(16611)

		return true
	} else {
		__antithesis_instrumentation__.Notify(16612)
	}
	__antithesis_instrumentation__.Notify(16606)

	return joberror.IsDistSQLRetryableError(err)
}

func MaybeStripRetryableErrorMarker(err error) error {
	__antithesis_instrumentation__.Notify(16613)

	if reflect.TypeOf(err) == retryableErrorType {
		__antithesis_instrumentation__.Notify(16615)
		err = errors.UnwrapOnce(err)
	} else {
		__antithesis_instrumentation__.Notify(16616)
	}
	__antithesis_instrumentation__.Notify(16614)
	return err
}

var retryableErrorType = reflect.TypeOf((*retryableError)(nil))
