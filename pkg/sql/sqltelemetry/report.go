package sqltelemetry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/logcrash"
	"github.com/cockroachdb/errors"
)

func RecordError(ctx context.Context, err error, sv *settings.Values) {
	__antithesis_instrumentation__.Notify(625813)

	telemetry.RecordError(err)

	code := pgerror.GetPGCode(err)
	switch {
	case code == pgcode.Uncategorized:
		__antithesis_instrumentation__.Notify(625814)

		telemetry.Inc(UncategorizedErrorCounter)

	case code == pgcode.Internal || func() bool {
		__antithesis_instrumentation__.Notify(625817)
		return errors.HasAssertionFailure(err) == true
	}() == true:
		__antithesis_instrumentation__.Notify(625815)

		log.Errorf(ctx, "encountered internal error:\n%+v", err)

		if logcrash.ShouldSendReport(sv) {
			__antithesis_instrumentation__.Notify(625818)
			event, extraDetails := errors.BuildSentryReport(err)
			logcrash.SendReport(ctx, logcrash.ReportTypeError, event, extraDetails)
		} else {
			__antithesis_instrumentation__.Notify(625819)
		}
	default:
		__antithesis_instrumentation__.Notify(625816)
	}
}
