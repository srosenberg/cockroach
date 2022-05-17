package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanlatch"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var MaxLockWaitQueueLength = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.lock_table.maximum_lock_wait_queue_length",
	"the maximum length of a lock wait-queue that read-write requests are willing "+
		"to enter and wait in. The setting can be used to ensure some level of quality-of-service "+
		"under severe per-key contention. If set to a non-zero value and an existing lock "+
		"wait-queue is already equal to or exceeding this length, requests will be rejected "+
		"eagerly instead of entering the queue and waiting. Set to 0 to disable.",
	0,
	func(v int64) error {
		__antithesis_instrumentation__.Notify(98946)
		if v < 0 {
			__antithesis_instrumentation__.Notify(98950)
			return errors.Errorf("cannot be set to a negative value: %d", v)
		} else {
			__antithesis_instrumentation__.Notify(98951)
		}
		__antithesis_instrumentation__.Notify(98947)
		if v == 0 {
			__antithesis_instrumentation__.Notify(98952)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(98953)
		}
		__antithesis_instrumentation__.Notify(98948)

		const minSafeMaxLength = 3
		if v < minSafeMaxLength {
			__antithesis_instrumentation__.Notify(98954)
			return errors.Errorf("cannot be set below %d: %d", minSafeMaxLength, v)
		} else {
			__antithesis_instrumentation__.Notify(98955)
		}
		__antithesis_instrumentation__.Notify(98949)
		return nil
	},
)

var DiscoveredLocksThresholdToConsultFinalizedTxnCache = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.lock_table.discovered_locks_threshold_for_consulting_finalized_txn_cache",
	"the maximum number of discovered locks by a waiter, above which the finalized txn cache"+
		"is consulted and resolvable locks are not added to the lock table -- this should be a small"+
		"fraction of the maximum number of locks in the lock table",
	200,
	settings.NonNegativeInt,
)

type managerImpl struct {
	st *cluster.Settings

	lm latchManager

	lt lockTable

	ltw lockTableWaiter

	twq txnWaitQueue
}

type Config struct {
	NodeDesc  *roachpb.NodeDescriptor
	RangeDesc *roachpb.RangeDescriptor

	Settings       *cluster.Settings
	DB             *kv.DB
	Clock          *hlc.Clock
	Stopper        *stop.Stopper
	IntentResolver IntentResolver

	TxnWaitMetrics *txnwait.Metrics
	SlowLatchGauge *metric.Gauge

	MaxLockTableSize  int64
	DisableTxnPushing bool
	TxnWaitKnobs      txnwait.TestingKnobs
}

func (c *Config) initDefaults() {
	__antithesis_instrumentation__.Notify(98956)
	if c.MaxLockTableSize == 0 {
		__antithesis_instrumentation__.Notify(98957)
		c.MaxLockTableSize = defaultLockTableSize
	} else {
		__antithesis_instrumentation__.Notify(98958)
	}
}

func NewManager(cfg Config) Manager {
	__antithesis_instrumentation__.Notify(98959)
	cfg.initDefaults()
	m := new(managerImpl)
	lt := newLockTable(cfg.MaxLockTableSize, cfg.RangeDesc.RangeID, cfg.Clock)
	*m = managerImpl{
		st: cfg.Settings,

		lm: &latchManagerImpl{
			m: spanlatch.Make(
				cfg.Stopper,
				cfg.SlowLatchGauge,
			),
		},
		lt: lt,
		ltw: &lockTableWaiterImpl{
			st:                cfg.Settings,
			clock:             cfg.Clock,
			stopper:           cfg.Stopper,
			ir:                cfg.IntentResolver,
			lt:                lt,
			disableTxnPushing: cfg.DisableTxnPushing,
		},

		twq: txnwait.NewQueue(txnwait.Config{
			RangeDesc: cfg.RangeDesc,
			DB:        cfg.DB,
			Clock:     cfg.Clock,
			Stopper:   cfg.Stopper,
			Metrics:   cfg.TxnWaitMetrics,
			Knobs:     cfg.TxnWaitKnobs,
		}),
	}
	return m
}

func (m *managerImpl) SequenceReq(
	ctx context.Context, prev *Guard, req Request, evalKind RequestEvalKind,
) (*Guard, Response, *Error) {
	__antithesis_instrumentation__.Notify(98960)
	var g *Guard
	if prev == nil {
		__antithesis_instrumentation__.Notify(98963)
		switch evalKind {
		case PessimisticEval:
			__antithesis_instrumentation__.Notify(98965)
			log.Event(ctx, "sequencing request")
		case OptimisticEval:
			__antithesis_instrumentation__.Notify(98966)
			log.Event(ctx, "optimistically sequencing request")
		case PessimisticAfterFailedOptimisticEval:
			__antithesis_instrumentation__.Notify(98967)
			panic("retry should have non-nil guard")
		default:
			__antithesis_instrumentation__.Notify(98968)
		}
		__antithesis_instrumentation__.Notify(98964)
		g = newGuard(req)
	} else {
		__antithesis_instrumentation__.Notify(98969)
		g = prev
		switch evalKind {
		case PessimisticEval:
			__antithesis_instrumentation__.Notify(98970)
			g.AssertNoLatches()
			log.Event(ctx, "re-sequencing request")
		case OptimisticEval:
			__antithesis_instrumentation__.Notify(98971)
			panic("optimistic eval cannot happen when re-sequencing")
		case PessimisticAfterFailedOptimisticEval:
			__antithesis_instrumentation__.Notify(98972)
			if !shouldIgnoreLatches(g.Req) {
				__antithesis_instrumentation__.Notify(98975)
				g.AssertLatches()
			} else {
				__antithesis_instrumentation__.Notify(98976)
			}
			__antithesis_instrumentation__.Notify(98973)
			log.Event(ctx, "re-sequencing request after optimistic sequencing failed")
		default:
			__antithesis_instrumentation__.Notify(98974)
		}
	}
	__antithesis_instrumentation__.Notify(98961)
	g.EvalKind = evalKind
	resp, err := m.sequenceReqWithGuard(ctx, g)
	if resp != nil || func() bool {
		__antithesis_instrumentation__.Notify(98977)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(98978)

		m.FinishReq(g)
		return nil, resp, err
	} else {
		__antithesis_instrumentation__.Notify(98979)
	}
	__antithesis_instrumentation__.Notify(98962)
	return g, nil, nil
}

func (m *managerImpl) sequenceReqWithGuard(ctx context.Context, g *Guard) (Response, *Error) {
	__antithesis_instrumentation__.Notify(98980)

	if shouldIgnoreLatches(g.Req) {
		__antithesis_instrumentation__.Notify(98984)
		log.Event(ctx, "not acquiring latches")
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(98985)
	}
	__antithesis_instrumentation__.Notify(98981)

	if shouldWaitOnLatchesWithoutAcquiring(g.Req) {
		__antithesis_instrumentation__.Notify(98986)
		log.Event(ctx, "waiting on latches without acquiring")
		return nil, m.lm.WaitFor(ctx, g.Req.LatchSpans, g.Req.PoisonPolicy)
	} else {
		__antithesis_instrumentation__.Notify(98987)
	}
	__antithesis_instrumentation__.Notify(98982)

	resp, err := m.maybeInterceptReq(ctx, g.Req)
	if resp != nil || func() bool {
		__antithesis_instrumentation__.Notify(98988)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(98989)
		return resp, err
	} else {
		__antithesis_instrumentation__.Notify(98990)
	}
	__antithesis_instrumentation__.Notify(98983)

	firstIteration := true
	for {
		__antithesis_instrumentation__.Notify(98991)
		if !g.HoldingLatches() {
			__antithesis_instrumentation__.Notify(98997)
			if g.EvalKind == OptimisticEval {
				__antithesis_instrumentation__.Notify(98998)
				if !firstIteration {
					__antithesis_instrumentation__.Notify(99000)

					panic("optimistic eval should not loop in sequenceReqWithGuard")
				} else {
					__antithesis_instrumentation__.Notify(99001)
				}
				__antithesis_instrumentation__.Notify(98999)
				log.Event(ctx, "optimistically acquiring latches")
				g.lg = m.lm.AcquireOptimistic(g.Req)
				g.lm = m.lm
			} else {
				__antithesis_instrumentation__.Notify(99002)

				log.Event(ctx, "acquiring latches")
				g.lg, err = m.lm.Acquire(ctx, g.Req)
				if err != nil {
					__antithesis_instrumentation__.Notify(99004)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(99005)
				}
				__antithesis_instrumentation__.Notify(99003)
				g.lm = m.lm
			}
		} else {
			__antithesis_instrumentation__.Notify(99006)
			if !firstIteration {
				__antithesis_instrumentation__.Notify(99009)
				panic(errors.AssertionFailedf("second or later iteration cannot be holding latches"))
			} else {
				__antithesis_instrumentation__.Notify(99010)
			}
			__antithesis_instrumentation__.Notify(99007)
			if g.EvalKind != PessimisticAfterFailedOptimisticEval {
				__antithesis_instrumentation__.Notify(99011)
				panic("must not be holding latches")
			} else {
				__antithesis_instrumentation__.Notify(99012)
			}
			__antithesis_instrumentation__.Notify(99008)
			log.Event(ctx, "optimistic failed, so waiting for latches")
			g.lg, err = m.lm.WaitUntilAcquired(ctx, g.lg)
			if err != nil {
				__antithesis_instrumentation__.Notify(99013)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(99014)
			}
		}
		__antithesis_instrumentation__.Notify(98992)
		firstIteration = false

		if g.Req.LockSpans.Empty() {
			__antithesis_instrumentation__.Notify(99015)
			return nil, nil
		} else {
			__antithesis_instrumentation__.Notify(99016)
		}
		__antithesis_instrumentation__.Notify(98993)

		if g.Req.MaxLockWaitQueueLength == 0 {
			__antithesis_instrumentation__.Notify(99017)
			g.Req.MaxLockWaitQueueLength = int(MaxLockWaitQueueLength.Get(&m.st.SV))
		} else {
			__antithesis_instrumentation__.Notify(99018)
		}
		__antithesis_instrumentation__.Notify(98994)

		if g.EvalKind == OptimisticEval {
			__antithesis_instrumentation__.Notify(99019)
			if g.ltg != nil {
				__antithesis_instrumentation__.Notify(99021)
				panic("Optimistic locking should not have a non-nil lockTableGuard")
			} else {
				__antithesis_instrumentation__.Notify(99022)
			}
			__antithesis_instrumentation__.Notify(99020)
			log.Event(ctx, "optimistically scanning lock table for conflicting locks")
			g.ltg = m.lt.ScanOptimistic(g.Req)
		} else {
			__antithesis_instrumentation__.Notify(99023)

			log.Event(ctx, "scanning lock table for conflicting locks")
			g.ltg = m.lt.ScanAndEnqueue(g.Req, g.ltg)
		}
		__antithesis_instrumentation__.Notify(98995)

		if g.ltg.ShouldWait() {
			__antithesis_instrumentation__.Notify(99024)
			m.lm.Release(g.moveLatchGuard())

			log.Event(ctx, "waiting in lock wait-queues")
			if err := m.ltw.WaitOn(ctx, g.Req, g.ltg); err != nil {
				__antithesis_instrumentation__.Notify(99026)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(99027)
			}
			__antithesis_instrumentation__.Notify(99025)
			continue
		} else {
			__antithesis_instrumentation__.Notify(99028)
		}
		__antithesis_instrumentation__.Notify(98996)
		return nil, nil
	}
}

func (m *managerImpl) maybeInterceptReq(ctx context.Context, req Request) (Response, *Error) {
	__antithesis_instrumentation__.Notify(99029)
	switch {
	case req.isSingle(roachpb.PushTxn):
		__antithesis_instrumentation__.Notify(99031)

		t := req.Requests[0].GetPushTxn()
		resp, err := m.twq.MaybeWaitForPush(ctx, t)
		if err != nil {
			__antithesis_instrumentation__.Notify(99034)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(99035)
			if resp != nil {
				__antithesis_instrumentation__.Notify(99036)
				return makeSingleResponse(resp), nil
			} else {
				__antithesis_instrumentation__.Notify(99037)
			}
		}
	case req.isSingle(roachpb.QueryTxn):
		__antithesis_instrumentation__.Notify(99032)

		t := req.Requests[0].GetQueryTxn()
		return nil, m.twq.MaybeWaitForQuery(ctx, t)
	default:
		__antithesis_instrumentation__.Notify(99033)

	}
	__antithesis_instrumentation__.Notify(99030)
	return nil, nil
}

func shouldIgnoreLatches(req Request) bool {
	__antithesis_instrumentation__.Notify(99038)
	switch {
	case req.ReadConsistency != roachpb.CONSISTENT:
		__antithesis_instrumentation__.Notify(99040)

		return true
	case req.isSingle(roachpb.RequestLease):
		__antithesis_instrumentation__.Notify(99041)

		return true
	default:
		__antithesis_instrumentation__.Notify(99042)
	}
	__antithesis_instrumentation__.Notify(99039)
	return false
}

func shouldWaitOnLatchesWithoutAcquiring(req Request) bool {
	__antithesis_instrumentation__.Notify(99043)
	return req.isSingle(roachpb.Barrier)
}

func (m *managerImpl) PoisonReq(g *Guard) {
	__antithesis_instrumentation__.Notify(99044)

	if g.lg != nil {
		__antithesis_instrumentation__.Notify(99045)
		m.lm.Poison(g.lg)
	} else {
		__antithesis_instrumentation__.Notify(99046)
	}
}

func (m *managerImpl) FinishReq(g *Guard) {
	__antithesis_instrumentation__.Notify(99047)

	if lg := g.moveLatchGuard(); lg != nil {
		__antithesis_instrumentation__.Notify(99050)
		m.lm.Release(lg)
	} else {
		__antithesis_instrumentation__.Notify(99051)
	}
	__antithesis_instrumentation__.Notify(99048)
	if ltg := g.moveLockTableGuard(); ltg != nil {
		__antithesis_instrumentation__.Notify(99052)
		m.lt.Dequeue(ltg)
	} else {
		__antithesis_instrumentation__.Notify(99053)
	}
	__antithesis_instrumentation__.Notify(99049)
	releaseGuard(g)
}

func (m *managerImpl) HandleWriterIntentError(
	ctx context.Context, g *Guard, seq roachpb.LeaseSequence, t *roachpb.WriteIntentError,
) (*Guard, *Error) {
	__antithesis_instrumentation__.Notify(99054)
	if g.ltg == nil {
		__antithesis_instrumentation__.Notify(99058)
		log.Fatalf(ctx, "cannot handle WriteIntentError %v for request without "+
			"lockTableGuard; were lock spans declared for this request?", t)
	} else {
		__antithesis_instrumentation__.Notify(99059)
	}
	__antithesis_instrumentation__.Notify(99055)

	consultFinalizedTxnCache :=
		int64(len(t.Intents)) > DiscoveredLocksThresholdToConsultFinalizedTxnCache.Get(&m.st.SV)
	for i := range t.Intents {
		__antithesis_instrumentation__.Notify(99060)
		intent := &t.Intents[i]
		added, err := m.lt.AddDiscoveredLock(intent, seq, consultFinalizedTxnCache, g.ltg)
		if err != nil {
			__antithesis_instrumentation__.Notify(99062)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(99063)
		}
		__antithesis_instrumentation__.Notify(99061)
		if !added {
			__antithesis_instrumentation__.Notify(99064)
			log.VEventf(ctx, 2,
				"intent on %s discovered but not added to disabled lock table",
				intent.Key.String())
		} else {
			__antithesis_instrumentation__.Notify(99065)
		}
	}
	__antithesis_instrumentation__.Notify(99056)

	m.lm.Release(g.moveLatchGuard())

	if toResolve := g.ltg.ResolveBeforeScanning(); len(toResolve) > 0 {
		__antithesis_instrumentation__.Notify(99066)
		if err := m.ltw.ResolveDeferredIntents(ctx, toResolve); err != nil {
			__antithesis_instrumentation__.Notify(99067)
			m.FinishReq(g)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(99068)
		}
	} else {
		__antithesis_instrumentation__.Notify(99069)
	}
	__antithesis_instrumentation__.Notify(99057)

	return g, nil
}

func (m *managerImpl) HandleTransactionPushError(
	ctx context.Context, g *Guard, t *roachpb.TransactionPushError,
) *Guard {
	__antithesis_instrumentation__.Notify(99070)
	m.twq.EnqueueTxn(&t.PusheeTxn)

	m.lm.Release(g.moveLatchGuard())
	return g
}

func (m *managerImpl) OnLockAcquired(ctx context.Context, acq *roachpb.LockAcquisition) {
	__antithesis_instrumentation__.Notify(99071)
	if err := m.lt.AcquireLock(&acq.Txn, acq.Key, lock.Exclusive, acq.Durability); err != nil {
		__antithesis_instrumentation__.Notify(99072)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(99073)
	}
}

func (m *managerImpl) OnLockUpdated(ctx context.Context, up *roachpb.LockUpdate) {
	__antithesis_instrumentation__.Notify(99074)
	if err := m.lt.UpdateLocks(up); err != nil {
		__antithesis_instrumentation__.Notify(99075)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(99076)
	}
}

func (m *managerImpl) QueryLockTableState(
	ctx context.Context, span roachpb.Span, opts QueryLockTableOptions,
) ([]roachpb.LockStateInfo, QueryLockTableResumeState) {
	__antithesis_instrumentation__.Notify(99077)
	return m.lt.QueryLockTableState(span, opts)
}

func (m *managerImpl) OnTransactionUpdated(ctx context.Context, txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(99078)
	m.twq.UpdateTxn(ctx, txn)
}

func (m *managerImpl) GetDependents(txnID uuid.UUID) []uuid.UUID {
	__antithesis_instrumentation__.Notify(99079)
	return m.twq.GetDependents(txnID)
}

func (m *managerImpl) OnRangeDescUpdated(desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(99080)
	m.twq.OnRangeDescUpdated(desc)
}

func (m *managerImpl) OnRangeLeaseUpdated(seq roachpb.LeaseSequence, isLeaseholder bool) {
	__antithesis_instrumentation__.Notify(99081)
	if isLeaseholder {
		__antithesis_instrumentation__.Notify(99082)
		m.lt.Enable(seq)
		m.twq.Enable(seq)
	} else {
		__antithesis_instrumentation__.Notify(99083)

		const disable = true
		m.lt.Clear(disable)
		m.twq.Clear(disable)
	}
}

func (m *managerImpl) OnRangeSplit() {
	__antithesis_instrumentation__.Notify(99084)

	const disable = false
	m.lt.Clear(disable)
	m.twq.Clear(disable)
}

func (m *managerImpl) OnRangeMerge() {
	__antithesis_instrumentation__.Notify(99085)

	const disable = true
	m.lt.Clear(disable)
	m.twq.Clear(disable)
}

func (m *managerImpl) OnReplicaSnapshotApplied() {
	__antithesis_instrumentation__.Notify(99086)

	const disable = false
	m.lt.Clear(disable)
}

func (m *managerImpl) LatchMetrics() LatchMetrics {
	__antithesis_instrumentation__.Notify(99087)
	return m.lm.Metrics()
}

func (m *managerImpl) LockTableMetrics() LockTableMetrics {
	__antithesis_instrumentation__.Notify(99088)
	return m.lt.Metrics()
}

func (m *managerImpl) TestingLockTableString() string {
	__antithesis_instrumentation__.Notify(99089)
	return m.lt.String()
}

func (m *managerImpl) TestingTxnWaitQueue() *txnwait.Queue {
	__antithesis_instrumentation__.Notify(99090)
	return m.twq.(*txnwait.Queue)
}

func (m *managerImpl) TestingSetMaxLocks(maxLocks int64) {
	__antithesis_instrumentation__.Notify(99091)
	m.lt.(*lockTableImpl).setMaxLocks(maxLocks)
}

func (r *Request) txnMeta() *enginepb.TxnMeta {
	__antithesis_instrumentation__.Notify(99092)
	if r.Txn == nil {
		__antithesis_instrumentation__.Notify(99094)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(99095)
	}
	__antithesis_instrumentation__.Notify(99093)
	return &r.Txn.TxnMeta
}

func (r *Request) isSingle(m roachpb.Method) bool {
	__antithesis_instrumentation__.Notify(99096)
	if len(r.Requests) != 1 {
		__antithesis_instrumentation__.Notify(99098)
		return false
	} else {
		__antithesis_instrumentation__.Notify(99099)
	}
	__antithesis_instrumentation__.Notify(99097)
	return r.Requests[0].GetInner().Method() == m
}

var guardPool = sync.Pool{
	New: func() interface{} { __antithesis_instrumentation__.Notify(99100); return new(Guard) },
}

func newGuard(req Request) *Guard {
	__antithesis_instrumentation__.Notify(99101)
	g := guardPool.Get().(*Guard)
	g.Req = req
	return g
}

func releaseGuard(g *Guard) {
	__antithesis_instrumentation__.Notify(99102)
	if g.Req.LatchSpans != nil {
		__antithesis_instrumentation__.Notify(99105)
		g.Req.LatchSpans.Release()
	} else {
		__antithesis_instrumentation__.Notify(99106)
	}
	__antithesis_instrumentation__.Notify(99103)
	if g.Req.LockSpans != nil {
		__antithesis_instrumentation__.Notify(99107)
		g.Req.LockSpans.Release()
	} else {
		__antithesis_instrumentation__.Notify(99108)
	}
	__antithesis_instrumentation__.Notify(99104)
	*g = Guard{}
	guardPool.Put(g)
}

func (g *Guard) LatchSpans() *spanset.SpanSet {
	__antithesis_instrumentation__.Notify(99109)
	return g.Req.LatchSpans
}

func (g *Guard) TakeSpanSets() (*spanset.SpanSet, *spanset.SpanSet) {
	__antithesis_instrumentation__.Notify(99110)
	la, lo := g.Req.LatchSpans, g.Req.LockSpans
	g.Req.LatchSpans, g.Req.LockSpans = nil, nil
	return la, lo
}

func (g *Guard) HoldingLatches() bool {
	__antithesis_instrumentation__.Notify(99111)
	return g != nil && func() bool {
		__antithesis_instrumentation__.Notify(99112)
		return g.lg != nil == true
	}() == true
}

func (g *Guard) AssertLatches() {
	__antithesis_instrumentation__.Notify(99113)
	if !shouldIgnoreLatches(g.Req) && func() bool {
		__antithesis_instrumentation__.Notify(99114)
		return !shouldWaitOnLatchesWithoutAcquiring(g.Req) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(99115)
		return !g.HoldingLatches() == true
	}() == true {
		__antithesis_instrumentation__.Notify(99116)
		panic("expected latches held, found none")
	} else {
		__antithesis_instrumentation__.Notify(99117)
	}
}

func (g *Guard) AssertNoLatches() {
	__antithesis_instrumentation__.Notify(99118)
	if g.HoldingLatches() {
		__antithesis_instrumentation__.Notify(99119)
		panic("unexpected latches held")
	} else {
		__antithesis_instrumentation__.Notify(99120)
	}
}

func (g *Guard) IsolatedAtLaterTimestamps() bool {
	__antithesis_instrumentation__.Notify(99121)

	return len(g.Req.LatchSpans.GetSpans(spanset.SpanReadOnly, spanset.SpanGlobal)) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(99122)
		return len(g.Req.LockSpans.GetSpans(spanset.SpanReadOnly, spanset.SpanGlobal)) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(99123)
		return len(g.Req.LockSpans.GetSpans(spanset.SpanReadOnly, spanset.SpanLocal)) == 0 == true
	}() == true
}

func (g *Guard) CheckOptimisticNoConflicts(
	latchSpansRead *spanset.SpanSet, lockSpansRead *spanset.SpanSet,
) (ok bool) {
	__antithesis_instrumentation__.Notify(99124)
	if g.EvalKind != OptimisticEval {
		__antithesis_instrumentation__.Notify(99129)
		panic(errors.AssertionFailedf("unexpected EvalKind: %d", g.EvalKind))
	} else {
		__antithesis_instrumentation__.Notify(99130)
	}
	__antithesis_instrumentation__.Notify(99125)
	if g.lg == nil && func() bool {
		__antithesis_instrumentation__.Notify(99131)
		return g.ltg == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(99132)
		return true
	} else {
		__antithesis_instrumentation__.Notify(99133)
	}
	__antithesis_instrumentation__.Notify(99126)
	if g.lg == nil {
		__antithesis_instrumentation__.Notify(99134)
		panic("expected non-nil latchGuard")
	} else {
		__antithesis_instrumentation__.Notify(99135)
	}
	__antithesis_instrumentation__.Notify(99127)

	if g.lm.CheckOptimisticNoConflicts(g.lg, latchSpansRead) {
		__antithesis_instrumentation__.Notify(99136)
		return g.ltg.CheckOptimisticNoConflicts(lockSpansRead)
	} else {
		__antithesis_instrumentation__.Notify(99137)
	}
	__antithesis_instrumentation__.Notify(99128)
	return false
}

func (g *Guard) CheckOptimisticNoLatchConflicts() (ok bool) {
	__antithesis_instrumentation__.Notify(99138)
	if g.EvalKind != OptimisticEval {
		__antithesis_instrumentation__.Notify(99141)
		panic(errors.AssertionFailedf("unexpected EvalKind: %d", g.EvalKind))
	} else {
		__antithesis_instrumentation__.Notify(99142)
	}
	__antithesis_instrumentation__.Notify(99139)
	if g.lg == nil {
		__antithesis_instrumentation__.Notify(99143)
		return true
	} else {
		__antithesis_instrumentation__.Notify(99144)
	}
	__antithesis_instrumentation__.Notify(99140)
	return g.lm.CheckOptimisticNoConflicts(g.lg, g.Req.LatchSpans)
}

func (g *Guard) moveLatchGuard() latchGuard {
	__antithesis_instrumentation__.Notify(99145)
	lg := g.lg
	g.lg = nil
	g.lm = nil
	return lg
}

func (g *Guard) moveLockTableGuard() lockTableGuard {
	__antithesis_instrumentation__.Notify(99146)
	ltg := g.ltg
	g.ltg = nil
	return ltg
}

func makeSingleResponse(r roachpb.Response) Response {
	__antithesis_instrumentation__.Notify(99147)
	ru := make(Response, 1)
	ru[0].MustSetInner(r)
	return ru
}
