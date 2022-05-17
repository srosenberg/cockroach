package joberror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv/kvclient/kvcoord"
	"github.com/cockroachdb/cockroach/pkg/sql/flowinfra"
	"github.com/cockroachdb/cockroach/pkg/util/circuit"
	"github.com/cockroachdb/cockroach/pkg/util/grpcutil"
	"github.com/cockroachdb/cockroach/pkg/util/sysutil"
	"github.com/cockroachdb/errors"
)

func IsDistSQLRetryableError(err error) bool {
	__antithesis_instrumentation__.Notify(70455)
	if err == nil {
		__antithesis_instrumentation__.Notify(70457)
		return false
	} else {
		__antithesis_instrumentation__.Notify(70458)
	}
	__antithesis_instrumentation__.Notify(70456)

	errStr := err.Error()

	return strings.Contains(errStr, `rpc error`)
}

func isBreakerOpenError(err error) bool {
	__antithesis_instrumentation__.Notify(70459)
	return errors.Is(err, circuit.ErrBreakerOpen)
}

func IsPermanentBulkJobError(err error) bool {
	__antithesis_instrumentation__.Notify(70460)
	if err == nil {
		__antithesis_instrumentation__.Notify(70462)
		return false
	} else {
		__antithesis_instrumentation__.Notify(70463)
	}
	__antithesis_instrumentation__.Notify(70461)

	return !IsDistSQLRetryableError(err) && func() bool {
		__antithesis_instrumentation__.Notify(70464)
		return !grpcutil.IsClosedConnection(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70465)
		return !flowinfra.IsNoInboundStreamConnectionError(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70466)
		return !kvcoord.IsSendError(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70467)
		return !isBreakerOpenError(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70468)
		return !sysutil.IsErrConnectionReset(err) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(70469)
		return !sysutil.IsErrConnectionRefused(err) == true
	}() == true
}
