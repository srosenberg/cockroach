package txnwait

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"container/list"
	"context"
	"runtime/pprof"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const maxWaitForQueryTxn = 50 * time.Millisecond

var TxnLivenessHeartbeatMultiplier = envutil.EnvOrDefaultInt(
	"COCKROACH_TXN_LIVENESS_HEARTBEAT_MULTIPLIER", 5)

var TxnLivenessThreshold = time.Duration(TxnLivenessHeartbeatMultiplier) * base.DefaultTxnHeartbeatInterval

func TestingOverrideTxnLivenessThreshold(t time.Duration) func() {
	__antithesis_instrumentation__.Notify(127288)
	old := TxnLivenessThreshold
	TxnLivenessThreshold = t
	return func() {
		__antithesis_instrumentation__.Notify(127289)
		TxnLivenessThreshold = old
	}
}

func ShouldPushImmediately(req *roachpb.PushTxnRequest) bool {
	__antithesis_instrumentation__.Notify(127290)
	if req.Force {
		__antithesis_instrumentation__.Notify(127294)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127295)
	}
	__antithesis_instrumentation__.Notify(127291)
	if !(req.PushType == roachpb.PUSH_ABORT || func() bool {
		__antithesis_instrumentation__.Notify(127296)
		return req.PushType == roachpb.PUSH_TIMESTAMP == true
	}() == true) {
		__antithesis_instrumentation__.Notify(127297)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127298)
	}
	__antithesis_instrumentation__.Notify(127292)
	p1, p2 := req.PusherTxn.Priority, req.PusheeTxn.Priority
	if p1 > p2 && func() bool {
		__antithesis_instrumentation__.Notify(127299)
		return (p1 == enginepb.MaxTxnPriority || func() bool {
			__antithesis_instrumentation__.Notify(127300)
			return p2 == enginepb.MinTxnPriority == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(127301)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127302)
	}
	__antithesis_instrumentation__.Notify(127293)
	return false
}

func isPushed(req *roachpb.PushTxnRequest, txn *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(127303)
	return (txn.Status.IsFinalized() || func() bool {
		__antithesis_instrumentation__.Notify(127304)
		return (req.PushType == roachpb.PUSH_TIMESTAMP && func() bool {
			__antithesis_instrumentation__.Notify(127305)
			return req.PushTo.LessEq(txn.WriteTimestamp) == true
		}() == true) == true
	}() == true)
}

func TxnExpiration(txn *roachpb.Transaction) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127306)
	return txn.LastActive().Add(TxnLivenessThreshold.Nanoseconds(), 0)
}

func IsExpired(now hlc.Timestamp, txn *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(127307)
	return TxnExpiration(txn).Less(now)
}

func createPushTxnResponse(txn *roachpb.Transaction) *roachpb.PushTxnResponse {
	__antithesis_instrumentation__.Notify(127308)
	return &roachpb.PushTxnResponse{PusheeTxn: *txn}
}

type waitingPush struct {
	req *roachpb.PushTxnRequest

	pending chan *roachpb.Transaction
	mu      struct {
		syncutil.Mutex
		dependents map[uuid.UUID]struct{}
	}
}

type waitingQueries struct {
	pending chan struct{}
	count   int
}

type pendingTxn struct {
	txn atomic.Value

	waitingPushes      *list.List
	waitingPushesAlloc list.List
}

func (pt *pendingTxn) getTxn() *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(127309)
	return pt.txn.Load().(*roachpb.Transaction)
}

func (pt *pendingTxn) takeWaitingPushes() *list.List {
	__antithesis_instrumentation__.Notify(127310)
	wp := pt.waitingPushes
	pt.waitingPushes = nil
	return wp
}

func (pt *pendingTxn) getDependentsSet() map[uuid.UUID]struct{} {
	__antithesis_instrumentation__.Notify(127311)
	set := map[uuid.UUID]struct{}{}
	for e := pt.waitingPushes.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(127313)
		push := e.Value.(*waitingPush)
		if id := push.req.PusherTxn.ID; id != (uuid.UUID{}) {
			__antithesis_instrumentation__.Notify(127314)
			set[id] = struct{}{}
			push.mu.Lock()
			for txnID := range push.mu.dependents {
				__antithesis_instrumentation__.Notify(127316)
				set[txnID] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(127315)
			push.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(127317)
		}
	}
	__antithesis_instrumentation__.Notify(127312)
	return set
}

type Config struct {
	RangeDesc *roachpb.RangeDescriptor
	DB        *kv.DB
	Clock     *hlc.Clock
	Stopper   *stop.Stopper
	Metrics   *Metrics
	Knobs     TestingKnobs
}

type TestingKnobs struct {
	OnPusherBlocked func(ctx context.Context, push *roachpb.PushTxnRequest)

	OnTxnUpdate func(ctx context.Context, txn *roachpb.Transaction)
}

type Queue struct {
	cfg Config
	mu  struct {
		syncutil.RWMutex
		txns    map[uuid.UUID]*pendingTxn
		queries map[uuid.UUID]*waitingQueries
	}
}

func NewQueue(cfg Config) *Queue {
	__antithesis_instrumentation__.Notify(127318)
	return &Queue{cfg: cfg}
}

func (q *Queue) Enable(_ roachpb.LeaseSequence) {
	__antithesis_instrumentation__.Notify(127319)
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.mu.txns == nil {
		__antithesis_instrumentation__.Notify(127321)
		q.mu.txns = map[uuid.UUID]*pendingTxn{}
	} else {
		__antithesis_instrumentation__.Notify(127322)
	}
	__antithesis_instrumentation__.Notify(127320)
	if q.mu.queries == nil {
		__antithesis_instrumentation__.Notify(127323)
		q.mu.queries = map[uuid.UUID]*waitingQueries{}
	} else {
		__antithesis_instrumentation__.Notify(127324)
	}
}

func (q *Queue) Clear(disable bool) {
	__antithesis_instrumentation__.Notify(127325)
	q.mu.Lock()
	waitingPushesLists := make([]*list.List, 0, len(q.mu.txns))
	waitingPushesCount := 0
	for _, pt := range q.mu.txns {
		__antithesis_instrumentation__.Notify(127331)
		waitingPushes := pt.takeWaitingPushes()
		waitingPushesLists = append(waitingPushesLists, waitingPushes)
		waitingPushesCount += waitingPushes.Len()
	}
	__antithesis_instrumentation__.Notify(127326)

	waitingQueriesMap := q.mu.queries
	waitingQueriesCount := 0
	for _, waitingQueries := range waitingQueriesMap {
		__antithesis_instrumentation__.Notify(127332)
		waitingQueriesCount += waitingQueries.count
	}
	__antithesis_instrumentation__.Notify(127327)

	metrics := q.cfg.Metrics
	metrics.PusheeWaiting.Dec(int64(len(q.mu.txns)))

	if log.V(1) {
		__antithesis_instrumentation__.Notify(127333)
		log.Infof(
			context.Background(),
			"clearing %d push waiters and %d query waiters",
			waitingPushesCount,
			waitingQueriesCount,
		)
	} else {
		__antithesis_instrumentation__.Notify(127334)
	}
	__antithesis_instrumentation__.Notify(127328)

	if disable {
		__antithesis_instrumentation__.Notify(127335)
		q.mu.txns = nil
		q.mu.queries = nil
	} else {
		__antithesis_instrumentation__.Notify(127336)
		q.mu.txns = map[uuid.UUID]*pendingTxn{}
		q.mu.queries = map[uuid.UUID]*waitingQueries{}
	}
	__antithesis_instrumentation__.Notify(127329)
	q.mu.Unlock()

	for _, w := range waitingPushesLists {
		__antithesis_instrumentation__.Notify(127337)
		for e := w.Front(); e != nil; e = e.Next() {
			__antithesis_instrumentation__.Notify(127338)
			push := e.Value.(*waitingPush)
			push.pending <- nil
		}
	}
	__antithesis_instrumentation__.Notify(127330)

	for _, w := range waitingQueriesMap {
		__antithesis_instrumentation__.Notify(127339)
		close(w.pending)
	}
}

func (q *Queue) IsEnabled() bool {
	__antithesis_instrumentation__.Notify(127340)
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.mu.txns != nil
}

func (q *Queue) OnRangeDescUpdated(desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(127341)
	q.mu.Lock()
	defer q.mu.Unlock()
	q.cfg.RangeDesc = desc
}

func (q *Queue) RangeContainsKeyLocked(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(127342)
	return kvserverbase.ContainsKey(q.cfg.RangeDesc, key)
}

func (q *Queue) EnqueueTxn(txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(127343)
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.mu.txns == nil {
		__antithesis_instrumentation__.Notify(127345)

		return
	} else {
		__antithesis_instrumentation__.Notify(127346)
	}
	__antithesis_instrumentation__.Notify(127344)

	if pt, ok := q.mu.txns[txn.ID]; ok {
		__antithesis_instrumentation__.Notify(127347)
		pt.txn.Store(txn)
	} else {
		__antithesis_instrumentation__.Notify(127348)
		q.cfg.Metrics.PusheeWaiting.Inc(1)
		pt = &pendingTxn{}
		pt.txn.Store(txn)
		pt.waitingPushes = &pt.waitingPushesAlloc
		pt.waitingPushes.Init()
		q.mu.txns[txn.ID] = pt
	}
}

func (q *Queue) UpdateTxn(ctx context.Context, txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(127349)
	txn.AssertInitialized(ctx)
	q.mu.Lock()
	if f := q.cfg.Knobs.OnTxnUpdate; f != nil {
		__antithesis_instrumentation__.Notify(127354)
		f(ctx, txn)
	} else {
		__antithesis_instrumentation__.Notify(127355)
	}
	__antithesis_instrumentation__.Notify(127350)

	q.releaseWaitingQueriesLocked(ctx, txn.ID)

	if q.mu.txns == nil {
		__antithesis_instrumentation__.Notify(127356)

		q.mu.Unlock()
		return
	} else {
		__antithesis_instrumentation__.Notify(127357)
	}
	__antithesis_instrumentation__.Notify(127351)

	pending, ok := q.mu.txns[txn.ID]
	if !ok {
		__antithesis_instrumentation__.Notify(127358)
		q.mu.Unlock()
		return
	} else {
		__antithesis_instrumentation__.Notify(127359)
	}
	__antithesis_instrumentation__.Notify(127352)
	waitingPushes := pending.takeWaitingPushes()
	pending.txn.Store(txn)
	delete(q.mu.txns, txn.ID)
	q.mu.Unlock()

	metrics := q.cfg.Metrics
	metrics.PusheeWaiting.Dec(1)

	if log.V(1) && func() bool {
		__antithesis_instrumentation__.Notify(127360)
		return waitingPushes.Len() > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(127361)
		log.Infof(ctx, "updating %d push waiters for %s", waitingPushes.Len(), txn.ID.Short())
	} else {
		__antithesis_instrumentation__.Notify(127362)
	}
	__antithesis_instrumentation__.Notify(127353)

	for e := waitingPushes.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(127363)
		push := e.Value.(*waitingPush)
		push.pending <- txn
	}
}

func (q *Queue) GetDependents(txnID uuid.UUID) []uuid.UUID {
	__antithesis_instrumentation__.Notify(127364)
	q.mu.RLock()
	defer q.mu.RUnlock()
	if q.mu.txns == nil {
		__antithesis_instrumentation__.Notify(127367)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(127368)
	}
	__antithesis_instrumentation__.Notify(127365)
	if pending, ok := q.mu.txns[txnID]; ok {
		__antithesis_instrumentation__.Notify(127369)
		set := pending.getDependentsSet()
		dependents := make([]uuid.UUID, 0, len(set))
		for txnID := range set {
			__antithesis_instrumentation__.Notify(127371)
			dependents = append(dependents, txnID)
		}
		__antithesis_instrumentation__.Notify(127370)
		return dependents
	} else {
		__antithesis_instrumentation__.Notify(127372)
	}
	__antithesis_instrumentation__.Notify(127366)
	return nil
}

func (q *Queue) isTxnUpdated(pending *pendingTxn, req *roachpb.QueryTxnRequest) bool {
	__antithesis_instrumentation__.Notify(127373)

	txn := pending.getTxn()
	if txn.Status.IsFinalized() || func() bool {
		__antithesis_instrumentation__.Notify(127377)
		return txn.Priority > req.Txn.Priority == true
	}() == true {
		__antithesis_instrumentation__.Notify(127378)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127379)
	}
	__antithesis_instrumentation__.Notify(127374)

	set := pending.getDependentsSet()
	if len(req.KnownWaitingTxns) != len(set) {
		__antithesis_instrumentation__.Notify(127380)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127381)
	}
	__antithesis_instrumentation__.Notify(127375)
	for _, txnID := range req.KnownWaitingTxns {
		__antithesis_instrumentation__.Notify(127382)
		if _, ok := set[txnID]; !ok {
			__antithesis_instrumentation__.Notify(127383)
			return true
		} else {
			__antithesis_instrumentation__.Notify(127384)
		}
	}
	__antithesis_instrumentation__.Notify(127376)
	return false
}

func (q *Queue) releaseWaitingQueriesLocked(ctx context.Context, txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(127385)
	if w, ok := q.mu.queries[txnID]; ok {
		__antithesis_instrumentation__.Notify(127386)
		log.VEventf(ctx, 2, "releasing %d waiting queries for %s", w.count, txnID.Short())
		close(w.pending)
		delete(q.mu.queries, txnID)
	} else {
		__antithesis_instrumentation__.Notify(127387)
	}
}

func (q *Queue) MaybeWaitForPush(
	ctx context.Context, req *roachpb.PushTxnRequest,
) (*roachpb.PushTxnResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127388)
	if ShouldPushImmediately(req) {
		__antithesis_instrumentation__.Notify(127397)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(127398)
	}
	__antithesis_instrumentation__.Notify(127389)

	q.mu.Lock()

	if q.mu.txns == nil || func() bool {
		__antithesis_instrumentation__.Notify(127399)
		return !q.RangeContainsKeyLocked(req.Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(127400)
		q.mu.Unlock()
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(127401)
	}
	__antithesis_instrumentation__.Notify(127390)

	pending, ok := q.mu.txns[req.PusheeTxn.ID]
	if !ok {
		__antithesis_instrumentation__.Notify(127402)
		q.mu.Unlock()
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(127403)
	}
	__antithesis_instrumentation__.Notify(127391)
	if txn := pending.getTxn(); isPushed(req, txn) {
		__antithesis_instrumentation__.Notify(127404)
		q.mu.Unlock()
		return createPushTxnResponse(txn), nil
	} else {
		__antithesis_instrumentation__.Notify(127405)
	}
	__antithesis_instrumentation__.Notify(127392)

	push := &waitingPush{
		req:     req,
		pending: make(chan *roachpb.Transaction, 1),
	}
	pushElem := pending.waitingPushes.PushBack(push)
	waitingPushesCount := pending.waitingPushes.Len()
	if f := q.cfg.Knobs.OnPusherBlocked; f != nil {
		__antithesis_instrumentation__.Notify(127406)
		f(ctx, req)
	} else {
		__antithesis_instrumentation__.Notify(127407)
	}
	__antithesis_instrumentation__.Notify(127393)

	q.releaseWaitingQueriesLocked(ctx, req.PusheeTxn.ID)
	q.mu.Unlock()

	defer func() {
		__antithesis_instrumentation__.Notify(127408)
		q.mu.Lock()
		pending.waitingPushes.Remove(pushElem)
		q.mu.Unlock()
	}()
	__antithesis_instrumentation__.Notify(127394)

	pusherStr := "non-txn"
	if req.PusherTxn.ID != (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(127409)
		pusherStr = req.PusherTxn.ID.Short()
	} else {
		__antithesis_instrumentation__.Notify(127410)
	}
	__antithesis_instrumentation__.Notify(127395)
	log.VEventf(
		ctx,
		2,
		"%s pushing %s (%d pending)",
		pusherStr,
		req.PusheeTxn.ID.Short(),
		waitingPushesCount,
	)
	var res *roachpb.PushTxnResponse
	var err *roachpb.Error
	labels := pprof.Labels("pushee", req.PusheeTxn.ID.String(), "pusher", pusherStr)
	pprof.Do(ctx, labels, func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(127411)
		res, err = q.waitForPush(ctx, req, push, pending)
	})
	__antithesis_instrumentation__.Notify(127396)
	return res, err
}

func (q *Queue) waitForPush(
	ctx context.Context, req *roachpb.PushTxnRequest, push *waitingPush, pending *pendingTxn,
) (*roachpb.PushTxnResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127412)

	var queryPusherCh <-chan *roachpb.Transaction
	var queryPusherErrCh <-chan *roachpb.Error
	var readyCh chan struct{}

	if req.PusherTxn.ID != uuid.Nil && func() bool {
		__antithesis_instrumentation__.Notify(127415)
		return req.PusherTxn.IsLocking() == true
	}() == true {
		__antithesis_instrumentation__.Notify(127416)

		var cancel func()
		ctx, cancel = context.WithCancel(ctx)
		readyCh = make(chan struct{}, 1)
		queryPusherCh, queryPusherErrCh = q.startQueryPusherTxn(ctx, push, readyCh)

		defer func() {
			__antithesis_instrumentation__.Notify(127417)
			cancel()
			if queryPusherErrCh != nil {
				__antithesis_instrumentation__.Notify(127418)
				<-queryPusherErrCh
			} else {
				__antithesis_instrumentation__.Notify(127419)
			}
		}()
	} else {
		__antithesis_instrumentation__.Notify(127420)
	}
	__antithesis_instrumentation__.Notify(127413)
	pusherPriority := req.PusherTxn.Priority
	pusheePriority := req.PusheeTxn.Priority

	metrics := q.cfg.Metrics
	metrics.PusherWaiting.Inc(1)
	defer metrics.PusherWaiting.Dec(1)
	tBegin := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(127421)
		metrics.PusherWaitTime.RecordValue(timeutil.Since(tBegin).Nanoseconds())
	}()
	__antithesis_instrumentation__.Notify(127414)

	slowTimerThreshold := time.Minute
	slowTimer := timeutil.NewTimer()
	defer slowTimer.Stop()
	slowTimer.Reset(slowTimerThreshold)

	var pusheeTxnTimer timeutil.Timer
	defer pusheeTxnTimer.Stop()

	pusheeTxnTimer.Reset(0)
	for {
		__antithesis_instrumentation__.Notify(127422)
		select {
		case <-slowTimer.C:
			__antithesis_instrumentation__.Notify(127423)
			slowTimer.Read = true
			metrics.PusherSlow.Inc(1)
			log.Warningf(ctx, "pusher %s: have been waiting %.2fs for pushee %s",
				req.PusherTxn.ID.Short(),
				timeutil.Since(tBegin).Seconds(),
				req.PusheeTxn.ID.Short(),
			)
			defer func() {
				__antithesis_instrumentation__.Notify(127439)
				metrics.PusherSlow.Dec(1)
				log.Warningf(ctx, "pusher %s: finished waiting after %.2fs for pushee %s",
					req.PusherTxn.ID.Short(),
					timeutil.Since(tBegin).Seconds(),
					req.PusheeTxn.ID.Short(),
				)
			}()
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(127424)

			log.VEvent(ctx, 2, "pusher giving up due to context cancellation")
			return nil, roachpb.NewError(ctx.Err())
		case <-q.cfg.Stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(127425)

			return nil, nil
		case txn := <-push.pending:
			__antithesis_instrumentation__.Notify(127426)
			log.VEventf(ctx, 2, "result of pending push: %v", txn)

			if txn == nil {
				__antithesis_instrumentation__.Notify(127440)
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(127441)
			}
			__antithesis_instrumentation__.Notify(127427)

			if isPushed(req, txn) {
				__antithesis_instrumentation__.Notify(127442)
				log.VEvent(ctx, 2, "push request is satisfied")
				return createPushTxnResponse(txn), nil
			} else {
				__antithesis_instrumentation__.Notify(127443)
			}
			__antithesis_instrumentation__.Notify(127428)

			log.VEvent(ctx, 2, "not pushed; returning to caller")
			return nil, nil

		case <-pusheeTxnTimer.C:
			__antithesis_instrumentation__.Notify(127429)
			log.VEvent(ctx, 2, "querying pushee")
			pusheeTxnTimer.Read = true

			updatedPushee, _, pErr := q.queryTxnStatus(
				ctx, req.PusheeTxn, false, nil,
			)
			if pErr != nil {
				__antithesis_instrumentation__.Notify(127444)
				return nil, pErr
			} else {
				__antithesis_instrumentation__.Notify(127445)
				if updatedPushee == nil {
					__antithesis_instrumentation__.Notify(127446)

					log.VEvent(ctx, 2, "pushee not found, push should now succeed")
					return nil, nil
				} else {
					__antithesis_instrumentation__.Notify(127447)
				}
			}
			__antithesis_instrumentation__.Notify(127430)
			pusheePriority = updatedPushee.Priority
			pending.txn.Store(updatedPushee)
			if updatedPushee.Status.IsFinalized() {
				__antithesis_instrumentation__.Notify(127448)
				log.VEvent(ctx, 2, "push request is satisfied")
				if updatedPushee.Status == roachpb.ABORTED {
					__antithesis_instrumentation__.Notify(127450)

					q.UpdateTxn(ctx, updatedPushee)
				} else {
					__antithesis_instrumentation__.Notify(127451)
				}
				__antithesis_instrumentation__.Notify(127449)
				return createPushTxnResponse(updatedPushee), nil
			} else {
				__antithesis_instrumentation__.Notify(127452)
			}
			__antithesis_instrumentation__.Notify(127431)
			if IsExpired(q.cfg.Clock.Now(), updatedPushee) {
				__antithesis_instrumentation__.Notify(127453)
				log.VEventf(ctx, 1, "pushing expired txn %s", req.PusheeTxn.ID.Short())
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(127454)
			}
			__antithesis_instrumentation__.Notify(127432)

			expiration := TxnExpiration(updatedPushee).GoTime()
			now := q.cfg.Clock.Now().GoTime()
			pusheeTxnTimer.Reset(expiration.Sub(now))

		case updatedPusher := <-queryPusherCh:
			__antithesis_instrumentation__.Notify(127433)
			switch updatedPusher.Status {
			case roachpb.COMMITTED:
				__antithesis_instrumentation__.Notify(127455)
				log.VEventf(ctx, 1, "pusher committed: %v", updatedPusher)
				return nil, roachpb.NewErrorWithTxn(roachpb.NewTransactionStatusError(
					roachpb.TransactionStatusError_REASON_TXN_COMMITTED,
					"already committed"),
					updatedPusher)
			case roachpb.ABORTED:
				__antithesis_instrumentation__.Notify(127456)
				log.VEventf(ctx, 1, "pusher aborted: %v", updatedPusher)
				return nil, roachpb.NewErrorWithTxn(
					roachpb.NewTransactionAbortedError(roachpb.ABORT_REASON_PUSHER_ABORTED), updatedPusher)
			default:
				__antithesis_instrumentation__.Notify(127457)
			}
			__antithesis_instrumentation__.Notify(127434)
			log.VEventf(ctx, 2, "pusher was updated: %v", updatedPusher)
			if updatedPusher.Priority > pusherPriority {
				__antithesis_instrumentation__.Notify(127458)
				pusherPriority = updatedPusher.Priority
			} else {
				__antithesis_instrumentation__.Notify(127459)
			}
			__antithesis_instrumentation__.Notify(127435)

			push.mu.Lock()
			_, haveDependency := push.mu.dependents[req.PusheeTxn.ID]
			dependents := make([]string, 0, len(push.mu.dependents))
			for id := range push.mu.dependents {
				__antithesis_instrumentation__.Notify(127460)
				dependents = append(dependents, id.Short())
			}
			__antithesis_instrumentation__.Notify(127436)
			log.VEventf(
				ctx,
				2,
				"%s (%d), pushing %s (%d), has dependencies=%s",
				req.PusherTxn.ID.Short(),
				pusherPriority,
				req.PusheeTxn.ID.Short(),
				pusheePriority,
				dependents,
			)
			push.mu.Unlock()

			q.mu.Lock()
			q.releaseWaitingQueriesLocked(ctx, req.PusheeTxn.ID)
			q.mu.Unlock()

			if haveDependency {
				__antithesis_instrumentation__.Notify(127461)

				p1, p2 := pusheePriority, pusherPriority
				if p1 < p2 || func() bool {
					__antithesis_instrumentation__.Notify(127462)
					return (p1 == p2 && func() bool {
						__antithesis_instrumentation__.Notify(127463)
						return bytes.Compare(req.PusheeTxn.ID.GetBytes(), req.PusherTxn.ID.GetBytes()) < 0 == true
					}() == true) == true
				}() == true {
					__antithesis_instrumentation__.Notify(127464)
					log.VEventf(
						ctx,
						1,
						"%s breaking deadlock by force push of %s; dependencies=%s",
						req.PusherTxn.ID.Short(),
						req.PusheeTxn.ID.Short(),
						dependents,
					)
					metrics.DeadlocksTotal.Inc(1)
					return q.forcePushAbort(ctx, req)
				} else {
					__antithesis_instrumentation__.Notify(127465)
				}
			} else {
				__antithesis_instrumentation__.Notify(127466)
			}
			__antithesis_instrumentation__.Notify(127437)

			readyCh <- struct{}{}

		case pErr := <-queryPusherErrCh:
			__antithesis_instrumentation__.Notify(127438)
			queryPusherErrCh = nil
			return nil, pErr
		}
	}
}

func (q *Queue) MaybeWaitForQuery(
	ctx context.Context, req *roachpb.QueryTxnRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(127467)
	if !req.WaitForUpdate {
		__antithesis_instrumentation__.Notify(127474)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(127475)
	}
	__antithesis_instrumentation__.Notify(127468)
	q.mu.Lock()

	if q.mu.txns == nil || func() bool {
		__antithesis_instrumentation__.Notify(127476)
		return !q.RangeContainsKeyLocked(req.Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(127477)
		q.mu.Unlock()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(127478)
	}
	__antithesis_instrumentation__.Notify(127469)

	var maxWaitCh <-chan time.Time

	if pending, ok := q.mu.txns[req.Txn.ID]; ok && func() bool {
		__antithesis_instrumentation__.Notify(127479)
		return q.isTxnUpdated(pending, req) == true
	}() == true {
		__antithesis_instrumentation__.Notify(127480)
		q.mu.Unlock()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(127481)
		if !ok {
			__antithesis_instrumentation__.Notify(127482)

			maxWaitCh = time.After(maxWaitForQueryTxn)
		} else {
			__antithesis_instrumentation__.Notify(127483)
		}
	}
	__antithesis_instrumentation__.Notify(127470)

	query, ok := q.mu.queries[req.Txn.ID]
	if ok {
		__antithesis_instrumentation__.Notify(127484)
		query.count++
	} else {
		__antithesis_instrumentation__.Notify(127485)
		query = &waitingQueries{
			pending: make(chan struct{}),
			count:   1,
		}
		q.mu.queries[req.Txn.ID] = query
	}
	__antithesis_instrumentation__.Notify(127471)
	q.mu.Unlock()

	defer func() {
		__antithesis_instrumentation__.Notify(127486)
		q.mu.Lock()
		query.count--
		if query.count == 0 && func() bool {
			__antithesis_instrumentation__.Notify(127488)
			return query == q.mu.queries[req.Txn.ID] == true
		}() == true {
			__antithesis_instrumentation__.Notify(127489)
			delete(q.mu.queries, req.Txn.ID)
		} else {
			__antithesis_instrumentation__.Notify(127490)
		}
		__antithesis_instrumentation__.Notify(127487)
		q.mu.Unlock()
	}()
	__antithesis_instrumentation__.Notify(127472)

	metrics := q.cfg.Metrics
	metrics.QueryWaiting.Inc(1)
	defer metrics.QueryWaiting.Dec(1)
	tBegin := timeutil.Now()
	defer func() {
		__antithesis_instrumentation__.Notify(127491)
		metrics.QueryWaitTime.RecordValue(timeutil.Since(tBegin).Nanoseconds())
	}()
	__antithesis_instrumentation__.Notify(127473)

	log.VEventf(ctx, 2, "waiting on query for %s", req.Txn.ID.Short())
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(127492)

		return roachpb.NewError(ctx.Err())
	case <-maxWaitCh:
		__antithesis_instrumentation__.Notify(127493)
		return nil
	case <-query.pending:
		__antithesis_instrumentation__.Notify(127494)
		return nil
	}
}

func (q *Queue) startQueryPusherTxn(
	ctx context.Context, push *waitingPush, readyCh <-chan struct{},
) (<-chan *roachpb.Transaction, <-chan *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127495)
	ch := make(chan *roachpb.Transaction, 1)
	errCh := make(chan *roachpb.Error, 1)
	push.mu.Lock()
	var waitingTxns []uuid.UUID
	if push.mu.dependents != nil {
		__antithesis_instrumentation__.Notify(127498)
		waitingTxns = make([]uuid.UUID, 0, len(push.mu.dependents))
		for txnID := range push.mu.dependents {
			__antithesis_instrumentation__.Notify(127499)
			waitingTxns = append(waitingTxns, txnID)
		}
	} else {
		__antithesis_instrumentation__.Notify(127500)
	}
	__antithesis_instrumentation__.Notify(127496)
	pusher := push.req.PusherTxn.Clone()
	push.mu.Unlock()

	if err := q.cfg.Stopper.RunAsyncTask(
		ctx, "monitoring pusher txn",
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(127501)

			for r := retry.StartWithCtx(ctx, base.DefaultRetryOptions()); r.Next(); {
				__antithesis_instrumentation__.Notify(127503)
				var pErr *roachpb.Error
				var updatedPusher *roachpb.Transaction
				updatedPusher, waitingTxns, pErr = q.queryTxnStatus(
					ctx, pusher.TxnMeta, true, waitingTxns,
				)
				if pErr != nil {
					__antithesis_instrumentation__.Notify(127508)
					errCh <- pErr
					return
				} else {
					__antithesis_instrumentation__.Notify(127509)
					if updatedPusher == nil {
						__antithesis_instrumentation__.Notify(127510)

						log.Event(ctx, "no pusher found; backing off")
						continue
					} else {
						__antithesis_instrumentation__.Notify(127511)
					}
				}
				__antithesis_instrumentation__.Notify(127504)

				push.mu.Lock()
				if push.mu.dependents == nil {
					__antithesis_instrumentation__.Notify(127512)
					push.mu.dependents = map[uuid.UUID]struct{}{}
				} else {
					__antithesis_instrumentation__.Notify(127513)
				}
				__antithesis_instrumentation__.Notify(127505)
				for _, txnID := range waitingTxns {
					__antithesis_instrumentation__.Notify(127514)
					push.mu.dependents[txnID] = struct{}{}
				}
				__antithesis_instrumentation__.Notify(127506)
				push.mu.Unlock()

				pusher.Update(updatedPusher)
				ch <- pusher

				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(127515)
					errCh <- roachpb.NewError(ctx.Err())
					return
				case <-readyCh:
					__antithesis_instrumentation__.Notify(127516)
				}
				__antithesis_instrumentation__.Notify(127507)

				r.Reset()
			}
			__antithesis_instrumentation__.Notify(127502)
			errCh <- roachpb.NewError(ctx.Err())
		}); err != nil {
		__antithesis_instrumentation__.Notify(127517)
		errCh <- roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(127518)
	}
	__antithesis_instrumentation__.Notify(127497)
	return ch, errCh
}

func (q *Queue) queryTxnStatus(
	ctx context.Context, txnMeta enginepb.TxnMeta, wait bool, dependents []uuid.UUID,
) (*roachpb.Transaction, []uuid.UUID, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127519)
	b := &kv.Batch{}
	b.Header.Timestamp = q.cfg.Clock.Now()
	b.AddRawRequest(&roachpb.QueryTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txnMeta.Key,
		},
		Txn:              txnMeta,
		WaitForUpdate:    wait,
		KnownWaitingTxns: dependents,
	})
	if err := q.cfg.DB.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(127522)

		return nil, nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(127523)
	}
	__antithesis_instrumentation__.Notify(127520)
	br := b.RawResponse()
	resp := br.Responses[0].GetInner().(*roachpb.QueryTxnResponse)

	if updatedTxn := &resp.QueriedTxn; updatedTxn.ID != (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(127524)
		return updatedTxn, resp.WaitingTxns, nil
	} else {
		__antithesis_instrumentation__.Notify(127525)
	}
	__antithesis_instrumentation__.Notify(127521)
	return nil, nil, nil
}

func (q *Queue) forcePushAbort(
	ctx context.Context, req *roachpb.PushTxnRequest,
) (*roachpb.PushTxnResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(127526)
	log.VEventf(ctx, 1, "force pushing %v to break deadlock", req.PusheeTxn.ID)
	forcePush := *req
	forcePush.Force = true
	forcePush.PushType = roachpb.PUSH_ABORT
	b := &kv.Batch{}
	b.Header.Timestamp = q.cfg.Clock.Now()
	b.Header.Timestamp.Forward(req.PushTo)
	b.AddRawRequest(&forcePush)
	if err := q.cfg.DB.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(127528)
		return nil, b.MustPErr()
	} else {
		__antithesis_instrumentation__.Notify(127529)
	}
	__antithesis_instrumentation__.Notify(127527)
	return b.RawResponse().Responses[0].GetPushTxn(), nil
}

func (q *Queue) TrackedTxns() map[uuid.UUID]struct{} {
	__antithesis_instrumentation__.Notify(127530)
	m := make(map[uuid.UUID]struct{})
	q.mu.RLock()
	for k := range q.mu.txns {
		__antithesis_instrumentation__.Notify(127532)
		m[k] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(127531)
	q.mu.RUnlock()
	return m
}
