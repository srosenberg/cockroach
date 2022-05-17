package clierror

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/log/logpb"
	"github.com/cockroachdb/errors"
)

func CheckAndMaybeLog(
	err error, logger func(context.Context, logpb.Severity, string, ...interface{}),
) error {
	__antithesis_instrumentation__.Notify(28274)
	if err == nil {
		__antithesis_instrumentation__.Notify(28277)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(28278)
	}
	__antithesis_instrumentation__.Notify(28275)
	severity := logpb.Severity_ERROR
	cause := err
	var ec *Error
	if errors.As(err, &ec) {
		__antithesis_instrumentation__.Notify(28279)
		severity = ec.GetSeverity()
		cause = ec.Unwrap()
	} else {
		__antithesis_instrumentation__.Notify(28280)
	}
	__antithesis_instrumentation__.Notify(28276)
	logger(context.Background(), severity, "%v", cause)
	return err
}
