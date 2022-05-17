package tenantrate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
)

type systemLimiter struct {
	tenantMetrics
}

func (s systemLimiter) Wait(ctx context.Context, reqInfo tenantcostmodel.RequestInfo) error {
	__antithesis_instrumentation__.Notify(126713)
	if isWrite, writeBytes := reqInfo.IsWrite(); isWrite {
		__antithesis_instrumentation__.Notify(126715)
		s.writeRequestsAdmitted.Inc(1)
		s.writeBytesAdmitted.Inc(writeBytes)
	} else {
		__antithesis_instrumentation__.Notify(126716)
		s.readRequestsAdmitted.Inc(1)
	}
	__antithesis_instrumentation__.Notify(126714)
	return nil
}

func (s systemLimiter) RecordRead(ctx context.Context, respInfo tenantcostmodel.ResponseInfo) {
	__antithesis_instrumentation__.Notify(126717)
	s.readBytesAdmitted.Inc(respInfo.ReadBytes())
}

var _ Limiter = (*systemLimiter)(nil)
