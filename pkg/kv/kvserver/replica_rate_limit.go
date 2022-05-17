package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func (r *Replica) maybeRateLimitBatch(ctx context.Context, ba *roachpb.BatchRequest) error {
	__antithesis_instrumentation__.Notify(120122)
	if r.tenantLimiter == nil {
		__antithesis_instrumentation__.Notify(120125)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120126)
	}
	__antithesis_instrumentation__.Notify(120123)
	tenantID, ok := roachpb.TenantFromContext(ctx)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(120127)
		return tenantID == roachpb.SystemTenantID == true
	}() == true {
		__antithesis_instrumentation__.Notify(120128)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(120129)
	}
	__antithesis_instrumentation__.Notify(120124)
	return r.tenantLimiter.Wait(ctx, tenantcostmodel.MakeRequestInfo(ba))
}

func (r *Replica) recordImpactOnRateLimiter(ctx context.Context, br *roachpb.BatchResponse) {
	__antithesis_instrumentation__.Notify(120130)
	if r.tenantLimiter == nil || func() bool {
		__antithesis_instrumentation__.Notify(120132)
		return br == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(120133)
		return
	} else {
		__antithesis_instrumentation__.Notify(120134)
	}
	__antithesis_instrumentation__.Notify(120131)

	r.tenantLimiter.RecordRead(ctx, tenantcostmodel.MakeResponseInfo(br))
}
