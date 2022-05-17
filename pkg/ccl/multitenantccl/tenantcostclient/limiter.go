package tenantcostclient

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/multitenant/tenantcostmodel"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type limiter struct {
	timeSource timeutil.TimeSource
	testInstr  TestInstrumentation
	tb         tokenBucket
	qp         *quotapool.AbstractPool

	waitingRU int64
}

const initialRUs = 10000
const initialRate = 100

func (l *limiter) Init(
	timeSource timeutil.TimeSource, testInstr TestInstrumentation, notifyChan chan struct{},
) {
	__antithesis_instrumentation__.Notify(19970)
	*l = limiter{
		timeSource: timeSource,
		testInstr:  testInstr,
	}

	l.tb.Init(timeSource.Now(), notifyChan, initialRate, initialRUs)

	onWaitStartFn := func(ctx context.Context, poolName string, r quotapool.Request) {
		__antithesis_instrumentation__.Notify(19973)
		req := r.(*waitRequest)

		if !req.waitingRUAccounted {
			__antithesis_instrumentation__.Notify(19974)
			req.waitingRUAccounted = true
			atomic.AddInt64(&l.waitingRU, req.neededCeil())
			if l.testInstr != nil {
				__antithesis_instrumentation__.Notify(19975)
				l.testInstr.Event(l.timeSource.Now(), WaitingRUAccountedInCallback)
			} else {
				__antithesis_instrumentation__.Notify(19976)
			}
		} else {
			__antithesis_instrumentation__.Notify(19977)
		}
	}
	__antithesis_instrumentation__.Notify(19971)

	onWaitFinishFn := func(ctx context.Context, poolName string, r quotapool.Request, start time.Time) {
		__antithesis_instrumentation__.Notify(19978)

		if waitDuration := timeSource.Since(start); waitDuration > time.Second {
			__antithesis_instrumentation__.Notify(19979)
			log.VEventf(ctx, 1, "request waited for RUs for %s", waitDuration.String())
		} else {
			__antithesis_instrumentation__.Notify(19980)
		}
	}
	__antithesis_instrumentation__.Notify(19972)

	l.qp = quotapool.New(
		"tenant-side-limiter", l,
		quotapool.WithTimeSource(timeSource),
		quotapool.OnWaitStart(onWaitStartFn),
		quotapool.OnWaitFinish(onWaitFinishFn),
	)
}

func (l *limiter) Close() {
	__antithesis_instrumentation__.Notify(19981)
	l.qp.Close("shutting down")
}

func (l *limiter) Wait(ctx context.Context, needed tenantcostmodel.RU) error {
	__antithesis_instrumentation__.Notify(19982)
	r := newWaitRequest(needed)
	defer putWaitRequest(r)

	return l.qp.Acquire(ctx, r)
}

func (l *limiter) RemoveTokens(now time.Time, delta tenantcostmodel.RU) {
	__antithesis_instrumentation__.Notify(19983)
	l.qp.Update(func(res quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(19984)
		l.tb.RemoveTokens(now, delta)

		return false
	})
}

func (l *limiter) Reconfigure(now time.Time, args tokenBucketReconfigureArgs) {
	__antithesis_instrumentation__.Notify(19985)
	l.qp.Update(func(quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(19986)
		l.tb.Reconfigure(now, args)

		return true
	})
}

func (l *limiter) AvailableTokens(now time.Time) tenantcostmodel.RU {
	__antithesis_instrumentation__.Notify(19987)
	var result tenantcostmodel.RU
	l.qp.Update(func(quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(19989)
		result = l.tb.AvailableTokens(now)
		return false
	})
	__antithesis_instrumentation__.Notify(19988)

	result -= tenantcostmodel.RU(atomic.LoadInt64(&l.waitingRU))
	return result
}

func (l *limiter) SetupNotification(now time.Time, threshold tenantcostmodel.RU) {
	__antithesis_instrumentation__.Notify(19990)
	l.qp.Update(func(quotapool.Resource) (shouldNotify bool) {
		__antithesis_instrumentation__.Notify(19991)
		l.tb.SetupNotification(now, threshold)

		return true
	})
}

type waitRequest struct {
	needed tenantcostmodel.RU

	waitingRUAccounted bool
}

var _ quotapool.Request = (*waitRequest)(nil)

var waitRequestSyncPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(19992); return new(waitRequest) },
}

func newWaitRequest(needed tenantcostmodel.RU) *waitRequest {
	__antithesis_instrumentation__.Notify(19993)
	r := waitRequestSyncPool.Get().(*waitRequest)
	*r = waitRequest{needed: needed}
	return r
}

func putWaitRequest(r *waitRequest) {
	__antithesis_instrumentation__.Notify(19994)
	*r = waitRequest{}
	waitRequestSyncPool.Put(r)
}

func (req *waitRequest) neededCeil() int64 {
	__antithesis_instrumentation__.Notify(19995)
	return int64(math.Ceil(float64(req.needed)))
}

func (req *waitRequest) Acquire(
	ctx context.Context, res quotapool.Resource,
) (fulfilled bool, tryAgainAfter time.Duration) {
	__antithesis_instrumentation__.Notify(19996)
	l := res.(*limiter)
	now := l.timeSource.Now()
	fulfilled, tryAgainAfter = l.tb.TryToFulfill(now, req.needed)

	if !fulfilled {
		__antithesis_instrumentation__.Notify(19998)

		if !req.waitingRUAccounted {
			__antithesis_instrumentation__.Notify(19999)
			req.waitingRUAccounted = true
			atomic.AddInt64(&l.waitingRU, req.neededCeil())
		} else {
			__antithesis_instrumentation__.Notify(20000)
		}
	} else {
		__antithesis_instrumentation__.Notify(20001)
		if req.waitingRUAccounted {
			__antithesis_instrumentation__.Notify(20002)
			req.waitingRUAccounted = false
			atomic.AddInt64(&l.waitingRU, -req.neededCeil())
		} else {
			__antithesis_instrumentation__.Notify(20003)
		}
	}
	__antithesis_instrumentation__.Notify(19997)
	return fulfilled, tryAgainAfter
}

func (req *waitRequest) ShouldWait() bool {
	__antithesis_instrumentation__.Notify(20004)
	return true
}
