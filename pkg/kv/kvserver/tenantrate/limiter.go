package tenantrate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type Limiter interface {
	Wait(ctx context.Context, reqInfo tenantcostmodel.RequestInfo) error

	RecordRead(ctx context.Context, respInfo tenantcostmodel.ResponseInfo)
}

type limiter struct {
	parent   *LimiterFactory
	tenantID roachpb.TenantID
	qp       *quotapool.AbstractPool
	metrics  tenantMetrics
}

func (rl *limiter) init(
	parent *LimiterFactory,
	tenantID roachpb.TenantID,
	config Config,
	metrics tenantMetrics,
	options ...quotapool.Option,
) {
	*rl = limiter{
		parent:   parent,
		tenantID: tenantID,
		metrics:  metrics,
	}

	bucket := &tokenBucket{}

	options = append(options,
		quotapool.OnWaitStart(
			func(ctx context.Context, poolName string, r quotapool.Request) {
				rl.metrics.currentBlocked.Inc(1)
			}),
		quotapool.OnWaitFinish(
			func(ctx context.Context, poolName string, r quotapool.Request, _ time.Time) {
				rl.metrics.currentBlocked.Dec(1)
			}),
	)

	rl.qp = quotapool.New(tenantID.String(), bucket, options...)
	bucket.init(config, rl.qp.TimeSource())
}

func (rl *limiter) Wait(ctx context.Context, reqInfo tenantcostmodel.RequestInfo) error {
	__antithesis_instrumentation__.Notify(126680)
	r := newWaitRequest(reqInfo)
	defer putWaitRequest(r)

	if err := rl.qp.Acquire(ctx, r); err != nil {
		__antithesis_instrumentation__.Notify(126683)
		return err
	} else {
		__antithesis_instrumentation__.Notify(126684)
	}
	__antithesis_instrumentation__.Notify(126681)

	if isWrite, writeBytes := reqInfo.IsWrite(); isWrite {
		__antithesis_instrumentation__.Notify(126685)
		rl.metrics.writeRequestsAdmitted.Inc(1)
		rl.metrics.writeBytesAdmitted.Inc(writeBytes)
	} else {
		__antithesis_instrumentation__.Notify(126686)

		rl.metrics.readRequestsAdmitted.Inc(1)
	}
	__antithesis_instrumentation__.Notify(126682)

	return nil
}

func (rl *limiter) RecordRead(ctx context.Context, respInfo tenantcostmodel.ResponseInfo) {
	__antithesis_instrumentation__.Notify(126687)
	rl.metrics.readBytesAdmitted.Inc(respInfo.ReadBytes())
	rl.qp.Update(func(res quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(126688)
		tb := res.(*tokenBucket)
		amount := float64(respInfo.ReadBytes()) * tb.config.ReadUnitsPerByte
		tb.Adjust(quotapool.Tokens(-amount))

		return false
	})
}

func (rl *limiter) updateConfig(config Config) {
	__antithesis_instrumentation__.Notify(126689)
	rl.qp.Update(func(res quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(126690)
		tb := res.(*tokenBucket)
		tb.config = config
		tb.UpdateConfig(quotapool.TokensPerSecond(config.Rate), quotapool.Tokens(config.Burst))
		return true
	})
}

type tokenBucket struct {
	quotapool.TokenBucket

	config Config
}

var _ quotapool.Resource = (*tokenBucket)(nil)

func (tb *tokenBucket) init(config Config, timeSource timeutil.TimeSource) {
	tb.TokenBucket.Init(
		quotapool.TokensPerSecond(config.Rate), quotapool.Tokens(config.Burst), timeSource,
	)
	tb.config = config
}

type waitRequest struct {
	info tenantcostmodel.RequestInfo
}

var _ quotapool.Request = (*waitRequest)(nil)

var waitRequestSyncPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(126691); return new(waitRequest) },
}

func newWaitRequest(info tenantcostmodel.RequestInfo) *waitRequest {
	__antithesis_instrumentation__.Notify(126692)
	r := waitRequestSyncPool.Get().(*waitRequest)
	*r = waitRequest{info: info}
	return r
}

func putWaitRequest(r *waitRequest) {
	__antithesis_instrumentation__.Notify(126693)
	*r = waitRequest{}
	waitRequestSyncPool.Put(r)
}

func (req *waitRequest) Acquire(
	ctx context.Context, res quotapool.Resource,
) (fulfilled bool, tryAgainAfter time.Duration) {
	__antithesis_instrumentation__.Notify(126694)
	tb := res.(*tokenBucket)
	var needed float64
	if isWrite, writeBytes := req.info.IsWrite(); isWrite {
		__antithesis_instrumentation__.Notify(126696)
		needed = tb.config.WriteRequestUnits + float64(writeBytes)*tb.config.WriteUnitsPerByte
	} else {
		__antithesis_instrumentation__.Notify(126697)
		needed = tb.config.ReadRequestUnits
	}
	__antithesis_instrumentation__.Notify(126695)
	return tb.TryToFulfill(quotapool.Tokens(needed))
}

func (req *waitRequest) ShouldWait() bool {
	__antithesis_instrumentation__.Notify(126698)
	return true
}
