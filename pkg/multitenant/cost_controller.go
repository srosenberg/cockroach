package multitenant

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlliveness"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
)

type TenantSideCostController interface {
	Start(
		ctx context.Context,
		stopper *stop.Stopper,
		instanceID base.SQLInstanceID,
		sessionID sqlliveness.SessionID,
		externalUsageFn ExternalUsageFn,
		nextLiveInstanceIDFn NextLiveInstanceIDFn,
	) error

	TenantSideKVInterceptor
}

type ExternalUsage struct {
	CPUSecs float64

	PGWireEgressBytes uint64
}

type ExternalUsageFn func(ctx context.Context) ExternalUsage

type NextLiveInstanceIDFn func(ctx context.Context) base.SQLInstanceID

type TenantSideKVInterceptor interface {
	OnRequestWait(ctx context.Context, info tenantcostmodel.RequestInfo) error

	OnResponse(
		ctx context.Context, req tenantcostmodel.RequestInfo, resp tenantcostmodel.ResponseInfo,
	)
}

func WithTenantCostControlExemption(ctx context.Context) context.Context {
	__antithesis_instrumentation__.Notify(128757)
	return context.WithValue(ctx, exemptCtxValue, exemptCtxValue)
}

func HasTenantCostControlExemption(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(128758)
	return ctx.Value(exemptCtxValue) != nil
}

type exemptCtxValueType struct{}

var exemptCtxValue interface{} = exemptCtxValueType{}
