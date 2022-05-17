package sqlproxyccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/errors"
)

type errorCode int

const (
	_ errorCode = iota

	codeAuthFailed

	codeBackendReadFailed

	codeBackendWriteFailed

	codeClientReadFailed

	codeClientWriteFailed

	codeUnexpectedInsecureStartupMessage

	codeSNIRoutingFailed

	codeUnexpectedStartupMessage

	codeParamsRoutingFailed

	codeBackendDown

	codeBackendRefusedTLS

	codeBackendDisconnected

	codeClientDisconnected

	codeProxyRefusedConnection

	codeExpiredClientConnection

	codeIdleDisconnect

	codeUnavailable
)

type codeError struct {
	code errorCode
	err  error
}

func (e *codeError) Error() string {
	__antithesis_instrumentation__.Notify(21502)
	return fmt.Sprintf("%s: %s", e.code, e.err)
}

func newErrorf(code errorCode, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(21503)
	return &codeError{
		code: code,
		err:  errors.Errorf(format, args...),
	}
}
