package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/intentresolver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel/attribute"
)

var LockTableLivenessPushDelay = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.lock_table.coordinator_liveness_push_delay",
	"the delay before pushing in order to detect coordinator failures of conflicting transactions",

	50*time.Millisecond,
)

var LockTableDeadlockDetectionPushDelay = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.lock_table.deadlock_detection_push_delay",
	"the delay before pushing in order to detect dependency cycles between transactions",

	100*time.Millisecond,
)

type lockTableWaiterImpl struct {
	st      *cluster.Settings
	clock   *hlc.Clock
	stopper *stop.Stopper
	ir      IntentResolver
	lt      lockTable

	disableTxnPushing bool

	onPushTimer func()
}

type IntentResolver interface {
	PushTransaction(
		context.Context, *enginepb.TxnMeta, roachpb.Header, roachpb.PushTxnType,
	) (*roachpb.Transaction, *Error)

	ResolveIntent(context.Context, roachpb.LockUpdate, intentresolver.ResolveOptions) *Error

	ResolveIntents(context.Context, []roachpb.LockUpdate, intentresolver.ResolveOptions) *Error
}

func (w *lockTableWaiterImpl) WaitOn(
	ctx context.Context, req Request, guard lockTableGuard,
) (err *Error) {
	__antithesis_instrumentation__.Notify(100200)
	newStateC := guard.NewStateChan()
	ctxDoneC := ctx.Done()
	shouldQuiesceC := w.stopper.ShouldQuiesce()

	var timer *timeutil.Timer
	var timerC <-chan time.Time
	var timerWaitingState waitingState

	var lockDeadline time.Time

	tracer := newContentionEventTracer(tracing.SpanFromContext(ctx), w.clock)

	defer tracer.notify(ctx, waitingState{kind: doneWaiting})

	for {
		__antithesis_instrumentation__.Notify(100201)
		select {

		case <-newStateC:
			__antithesis_instrumentation__.Notify(100202)
			timerC = nil
			state := guard.CurState()
			log.Eventf(ctx, "lock wait-queue event: %s", state)
			tracer.notify(ctx, state)
			switch state.kind {
			case waitFor, waitForDistinguished:
				__antithesis_instrumentation__.Notify(100211)
				if req.WaitPolicy == lock.WaitPolicy_Error {
					__antithesis_instrumentation__.Notify(100228)

					if state.held {
						__antithesis_instrumentation__.Notify(100231)
						err = w.pushLockTxn(ctx, req, state)
					} else {
						__antithesis_instrumentation__.Notify(100232)
						err = newWriteIntentErr(req, state, reasonWaitPolicy)
					}
					__antithesis_instrumentation__.Notify(100229)
					if err != nil {
						__antithesis_instrumentation__.Notify(100233)
						return err
					} else {
						__antithesis_instrumentation__.Notify(100234)
					}
					__antithesis_instrumentation__.Notify(100230)
					continue
				} else {
					__antithesis_instrumentation__.Notify(100235)
				}
				__antithesis_instrumentation__.Notify(100212)

				livenessPush := state.kind == waitForDistinguished
				deadlockPush := true

				if !state.held {
					__antithesis_instrumentation__.Notify(100236)
					livenessPush = false
				} else {
					__antithesis_instrumentation__.Notify(100237)
				}
				__antithesis_instrumentation__.Notify(100213)

				if req.Txn == nil {
					__antithesis_instrumentation__.Notify(100238)
					deadlockPush = false
				} else {
					__antithesis_instrumentation__.Notify(100239)
				}
				__antithesis_instrumentation__.Notify(100214)

				timeoutPush := req.LockTimeout != 0

				if !livenessPush && func() bool {
					__antithesis_instrumentation__.Notify(100240)
					return !deadlockPush == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(100241)
					return !timeoutPush == true
				}() == true {
					__antithesis_instrumentation__.Notify(100242)
					continue
				} else {
					__antithesis_instrumentation__.Notify(100243)
				}
				__antithesis_instrumentation__.Notify(100215)

				delay := time.Duration(math.MaxInt64)
				if livenessPush {
					__antithesis_instrumentation__.Notify(100244)
					delay = minDuration(delay, LockTableLivenessPushDelay.Get(&w.st.SV))
				} else {
					__antithesis_instrumentation__.Notify(100245)
				}
				__antithesis_instrumentation__.Notify(100216)
				if deadlockPush {
					__antithesis_instrumentation__.Notify(100246)
					delay = minDuration(delay, LockTableDeadlockDetectionPushDelay.Get(&w.st.SV))
				} else {
					__antithesis_instrumentation__.Notify(100247)
				}
				__antithesis_instrumentation__.Notify(100217)
				if timeoutPush {
					__antithesis_instrumentation__.Notify(100248)

					oldState := timerWaitingState
					newLock := !oldState.key.Equal(state.key) || func() bool {
						__antithesis_instrumentation__.Notify(100250)
						return oldState.txn.ID != state.txn.ID == true
					}() == true
					if newLock {
						__antithesis_instrumentation__.Notify(100251)
						lockDeadline = w.clock.PhysicalTime().Add(req.LockTimeout)
					} else {
						__antithesis_instrumentation__.Notify(100252)
					}
					__antithesis_instrumentation__.Notify(100249)
					delay = minDuration(delay, w.timeUntilDeadline(lockDeadline))
				} else {
					__antithesis_instrumentation__.Notify(100253)
				}
				__antithesis_instrumentation__.Notify(100218)

				if hasMinPriority(state.txn) || func() bool {
					__antithesis_instrumentation__.Notify(100254)
					return hasMaxPriority(req.Txn) == true
				}() == true {
					__antithesis_instrumentation__.Notify(100255)
					delay = 0
				} else {
					__antithesis_instrumentation__.Notify(100256)
				}
				__antithesis_instrumentation__.Notify(100219)

				if delay > 0 {
					__antithesis_instrumentation__.Notify(100257)
					if timer == nil {
						__antithesis_instrumentation__.Notify(100259)
						timer = timeutil.NewTimer()
						defer timer.Stop()
					} else {
						__antithesis_instrumentation__.Notify(100260)
					}
					__antithesis_instrumentation__.Notify(100258)
					timer.Reset(delay)
					timerC = timer.C
				} else {
					__antithesis_instrumentation__.Notify(100261)

					timerC = closedTimerC
				}
				__antithesis_instrumentation__.Notify(100220)
				timerWaitingState = state

			case waitElsewhere:
				__antithesis_instrumentation__.Notify(100221)

				if !state.held {
					__antithesis_instrumentation__.Notify(100262)

					return nil
				} else {
					__antithesis_instrumentation__.Notify(100263)
				}
				__antithesis_instrumentation__.Notify(100222)

				if req.LockTimeout != 0 {
					__antithesis_instrumentation__.Notify(100264)
					return doWithTimeoutAndFallback(
						ctx, req.LockTimeout,
						func(ctx context.Context) *Error {
							__antithesis_instrumentation__.Notify(100265)
							return w.pushLockTxn(ctx, req, state)
						},
						func(ctx context.Context) *Error {
							__antithesis_instrumentation__.Notify(100266)
							return w.pushLockTxnAfterTimeout(ctx, req, state)
						},
					)
				} else {
					__antithesis_instrumentation__.Notify(100267)
				}
				__antithesis_instrumentation__.Notify(100223)
				return w.pushLockTxn(ctx, req, state)

			case waitSelf:
				__antithesis_instrumentation__.Notify(100224)

			case waitQueueMaxLengthExceeded:
				__antithesis_instrumentation__.Notify(100225)

				return newWriteIntentErr(req, state, reasonWaitQueueMaxLengthExceeded)

			case doneWaiting:
				__antithesis_instrumentation__.Notify(100226)

				toResolve := guard.ResolveBeforeScanning()
				return w.ResolveDeferredIntents(ctx, toResolve)

			default:
				__antithesis_instrumentation__.Notify(100227)
				panic("unexpected waiting state")
			}

		case <-timerC:
			__antithesis_instrumentation__.Notify(100203)

			timerC = nil
			if timer != nil {
				__antithesis_instrumentation__.Notify(100268)
				timer.Read = true
			} else {
				__antithesis_instrumentation__.Notify(100269)
			}
			__antithesis_instrumentation__.Notify(100204)
			if w.onPushTimer != nil {
				__antithesis_instrumentation__.Notify(100270)
				w.onPushTimer()
			} else {
				__antithesis_instrumentation__.Notify(100271)
			}
			__antithesis_instrumentation__.Notify(100205)

			pushWait := func(ctx context.Context) *Error {
				__antithesis_instrumentation__.Notify(100272)

				if timerWaitingState.held {
					__antithesis_instrumentation__.Notify(100275)
					return w.pushLockTxn(ctx, req, timerWaitingState)
				} else {
					__antithesis_instrumentation__.Notify(100276)
				}
				__antithesis_instrumentation__.Notify(100273)

				pushCtx, pushCancel := context.WithCancel(ctx)
				defer pushCancel()
				go watchForNotifications(pushCtx, pushCancel, newStateC)
				err := w.pushRequestTxn(pushCtx, req, timerWaitingState)
				if errors.Is(pushCtx.Err(), context.Canceled) {
					__antithesis_instrumentation__.Notify(100277)

					err = nil
				} else {
					__antithesis_instrumentation__.Notify(100278)
				}
				__antithesis_instrumentation__.Notify(100274)
				return err
			}
			__antithesis_instrumentation__.Notify(100206)

			pushNoWait := func(ctx context.Context) *Error {
				__antithesis_instrumentation__.Notify(100279)

				if timerWaitingState.held {
					__antithesis_instrumentation__.Notify(100281)
					return w.pushLockTxnAfterTimeout(ctx, req, timerWaitingState)
				} else {
					__antithesis_instrumentation__.Notify(100282)
				}
				__antithesis_instrumentation__.Notify(100280)
				return newWriteIntentErr(req, timerWaitingState, reasonLockTimeout)
			}
			__antithesis_instrumentation__.Notify(100207)

			if !lockDeadline.IsZero() {
				__antithesis_instrumentation__.Notify(100283)
				untilDeadline := w.timeUntilDeadline(lockDeadline)
				if untilDeadline == 0 {
					__antithesis_instrumentation__.Notify(100284)

					err = pushNoWait(ctx)
				} else {
					__antithesis_instrumentation__.Notify(100285)

					err = doWithTimeoutAndFallback(ctx, untilDeadline, pushWait, pushNoWait)
				}
			} else {
				__antithesis_instrumentation__.Notify(100286)

				err = pushWait(ctx)
			}
			__antithesis_instrumentation__.Notify(100208)
			if err != nil {
				__antithesis_instrumentation__.Notify(100287)
				return err
			} else {
				__antithesis_instrumentation__.Notify(100288)
			}

		case <-ctxDoneC:
			__antithesis_instrumentation__.Notify(100209)
			return roachpb.NewError(ctx.Err())

		case <-shouldQuiesceC:
			__antithesis_instrumentation__.Notify(100210)
			return roachpb.NewError(&roachpb.NodeUnavailableError{})
		}
	}
}

func (w *lockTableWaiterImpl) pushLockTxn(
	ctx context.Context, req Request, ws waitingState,
) *Error {
	__antithesis_instrumentation__.Notify(100289)
	if w.disableTxnPushing {
		__antithesis_instrumentation__.Notify(100294)
		return newWriteIntentErr(req, ws, reasonWaitPolicy)
	} else {
		__antithesis_instrumentation__.Notify(100295)
	}
	__antithesis_instrumentation__.Notify(100290)

	h := w.pushHeader(req)
	var pushType roachpb.PushTxnType
	switch req.WaitPolicy {
	case lock.WaitPolicy_Block:
		__antithesis_instrumentation__.Notify(100296)

		switch ws.guardAccess {
		case spanset.SpanReadOnly:
			__antithesis_instrumentation__.Notify(100299)
			pushType = roachpb.PUSH_TIMESTAMP
			log.VEventf(ctx, 2, "pushing timestamp of txn %s above %s", ws.txn.ID.Short(), h.Timestamp)

		case spanset.SpanReadWrite:
			__antithesis_instrumentation__.Notify(100300)
			pushType = roachpb.PUSH_ABORT
			log.VEventf(ctx, 2, "pushing txn %s to abort", ws.txn.ID.Short())
		default:
			__antithesis_instrumentation__.Notify(100301)
		}

	case lock.WaitPolicy_Error:
		__antithesis_instrumentation__.Notify(100297)

		pushType = roachpb.PUSH_TOUCH
		log.VEventf(ctx, 2, "pushing txn %s to check if abandoned", ws.txn.ID.Short())

	default:
		__antithesis_instrumentation__.Notify(100298)
		log.Fatalf(ctx, "unexpected WaitPolicy: %v", req.WaitPolicy)
	}
	__antithesis_instrumentation__.Notify(100291)

	pusheeTxn, err := w.ir.PushTransaction(ctx, ws.txn, h, pushType)
	if err != nil {
		__antithesis_instrumentation__.Notify(100302)

		if _, ok := err.GetDetail().(*roachpb.TransactionPushError); ok && func() bool {
			__antithesis_instrumentation__.Notify(100304)
			return req.WaitPolicy == lock.WaitPolicy_Error == true
		}() == true {
			__antithesis_instrumentation__.Notify(100305)
			err = newWriteIntentErr(req, ws, reasonWaitPolicy)
		} else {
			__antithesis_instrumentation__.Notify(100306)
		}
		__antithesis_instrumentation__.Notify(100303)
		return err
	} else {
		__antithesis_instrumentation__.Notify(100307)
	}
	__antithesis_instrumentation__.Notify(100292)

	if pusheeTxn.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(100308)
		w.lt.TransactionIsFinalized(pusheeTxn)
	} else {
		__antithesis_instrumentation__.Notify(100309)
	}
	__antithesis_instrumentation__.Notify(100293)

	resolve := roachpb.MakeLockUpdate(pusheeTxn, roachpb.Span{Key: ws.key})
	opts := intentresolver.ResolveOptions{Poison: true}
	return w.ir.ResolveIntent(ctx, resolve, opts)
}

func (w *lockTableWaiterImpl) pushLockTxnAfterTimeout(
	ctx context.Context, req Request, ws waitingState,
) *Error {
	__antithesis_instrumentation__.Notify(100310)
	req.WaitPolicy = lock.WaitPolicy_Error
	err := w.pushLockTxn(ctx, req, ws)
	if _, ok := err.GetDetail().(*roachpb.WriteIntentError); ok {
		__antithesis_instrumentation__.Notify(100312)
		err = newWriteIntentErr(req, ws, reasonLockTimeout)
	} else {
		__antithesis_instrumentation__.Notify(100313)
	}
	__antithesis_instrumentation__.Notify(100311)
	return err
}

func (w *lockTableWaiterImpl) pushRequestTxn(
	ctx context.Context, req Request, ws waitingState,
) *Error {
	__antithesis_instrumentation__.Notify(100314)

	h := w.pushHeader(req)
	pushType := roachpb.PUSH_ABORT
	log.VEventf(ctx, 3, "pushing txn %s to detect request deadlock", ws.txn.ID.Short())

	_, err := w.ir.PushTransaction(ctx, ws.txn, h, pushType)
	if err != nil {
		__antithesis_instrumentation__.Notify(100316)
		return err
	} else {
		__antithesis_instrumentation__.Notify(100317)
	}
	__antithesis_instrumentation__.Notify(100315)

	return nil
}

func (w *lockTableWaiterImpl) pushHeader(req Request) roachpb.Header {
	__antithesis_instrumentation__.Notify(100318)
	h := roachpb.Header{
		Timestamp:    req.Timestamp,
		UserPriority: req.Priority,
	}
	if req.Txn != nil {
		__antithesis_instrumentation__.Notify(100320)

		h.Txn = req.Txn.Clone()

		uncertaintyLimit := req.Txn.GlobalUncertaintyLimit.WithSynthetic(true)
		if !h.Timestamp.Synthetic {
			__antithesis_instrumentation__.Notify(100322)

			uncertaintyLimit.Backward(w.clock.Now())
		} else {
			__antithesis_instrumentation__.Notify(100323)
		}
		__antithesis_instrumentation__.Notify(100321)
		h.Timestamp.Forward(uncertaintyLimit)
	} else {
		__antithesis_instrumentation__.Notify(100324)
	}
	__antithesis_instrumentation__.Notify(100319)
	return h
}

func (w *lockTableWaiterImpl) timeUntilDeadline(deadline time.Time) time.Duration {
	__antithesis_instrumentation__.Notify(100325)
	dur := deadline.Sub(w.clock.PhysicalTime())
	const soon = 250 * time.Microsecond
	if dur <= soon {
		__antithesis_instrumentation__.Notify(100327)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(100328)
	}
	__antithesis_instrumentation__.Notify(100326)
	return dur
}

func (w *lockTableWaiterImpl) ResolveDeferredIntents(
	ctx context.Context, deferredResolution []roachpb.LockUpdate,
) *Error {
	__antithesis_instrumentation__.Notify(100329)
	if len(deferredResolution) == 0 {
		__antithesis_instrumentation__.Notify(100331)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(100332)
	}
	__antithesis_instrumentation__.Notify(100330)

	opts := intentresolver.ResolveOptions{Poison: true}
	return w.ir.ResolveIntents(ctx, deferredResolution, opts)
}

func doWithTimeoutAndFallback(
	ctx context.Context,
	timeout time.Duration,
	withTimeout, afterTimeout func(ctx context.Context) *Error,
) *Error {
	__antithesis_instrumentation__.Notify(100333)
	timeoutCtx, timeoutCancel := context.WithTimeout(ctx, timeout)
	defer timeoutCancel()
	err := withTimeout(timeoutCtx)
	if !errors.Is(timeoutCtx.Err(), context.DeadlineExceeded) {
		__antithesis_instrumentation__.Notify(100335)

		return err
	} else {
		__antithesis_instrumentation__.Notify(100336)
	}
	__antithesis_instrumentation__.Notify(100334)

	return afterTimeout(ctx)
}

func watchForNotifications(ctx context.Context, cancel func(), newStateC chan struct{}) {
	__antithesis_instrumentation__.Notify(100337)
	select {
	case <-newStateC:
		__antithesis_instrumentation__.Notify(100338)

		select {
		case newStateC <- struct{}{}:
			__antithesis_instrumentation__.Notify(100341)
		default:
			__antithesis_instrumentation__.Notify(100342)
		}
		__antithesis_instrumentation__.Notify(100339)

		cancel()
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(100340)
	}
}

type txnCache struct {
	mu   syncutil.Mutex
	txns [8]*roachpb.Transaction
}

func (c *txnCache) get(id uuid.UUID) (*roachpb.Transaction, bool) {
	__antithesis_instrumentation__.Notify(100343)
	c.mu.Lock()
	defer c.mu.Unlock()
	if idx := c.getIdxLocked(id); idx >= 0 {
		__antithesis_instrumentation__.Notify(100345)
		txn := c.txns[idx]
		c.moveFrontLocked(txn, idx)
		return txn, true
	} else {
		__antithesis_instrumentation__.Notify(100346)
	}
	__antithesis_instrumentation__.Notify(100344)
	return nil, false
}

func (c *txnCache) add(txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(100347)
	c.mu.Lock()
	defer c.mu.Unlock()
	if idx := c.getIdxLocked(txn.ID); idx >= 0 {
		__antithesis_instrumentation__.Notify(100348)
		c.moveFrontLocked(txn, idx)
	} else {
		__antithesis_instrumentation__.Notify(100349)
		c.insertFrontLocked(txn)
	}
}

func (c *txnCache) clear() {
	__antithesis_instrumentation__.Notify(100350)
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := range c.txns {
		__antithesis_instrumentation__.Notify(100351)
		c.txns[i] = nil
	}
}

func (c *txnCache) getIdxLocked(id uuid.UUID) int {
	__antithesis_instrumentation__.Notify(100352)
	for i, txn := range c.txns {
		__antithesis_instrumentation__.Notify(100354)
		if txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(100355)
			return txn.ID == id == true
		}() == true {
			__antithesis_instrumentation__.Notify(100356)
			return i
		} else {
			__antithesis_instrumentation__.Notify(100357)
		}
	}
	__antithesis_instrumentation__.Notify(100353)
	return -1
}

func (c *txnCache) moveFrontLocked(txn *roachpb.Transaction, cur int) {
	__antithesis_instrumentation__.Notify(100358)
	copy(c.txns[1:cur+1], c.txns[:cur])
	c.txns[0] = txn
}

func (c *txnCache) insertFrontLocked(txn *roachpb.Transaction) {
	__antithesis_instrumentation__.Notify(100359)
	copy(c.txns[1:], c.txns[:])
	c.txns[0] = txn
}

const tagContentionTracer = "contention_tracer"

const tagWaitKey = "lock_wait_key"

const tagWaitStart = "lock_wait_start"

const tagLockHolderTxn = "lock_holder_txn"

const tagNumLocks = "lock_num"

const tagWaited = "lock_wait"

type contentionEventTracer struct {
	sp      *tracing.Span
	onEvent func(event *roachpb.ContentionEvent)
	tag     contentionTag
}

type contentionTag struct {
	clock *hlc.Clock
	mu    struct {
		syncutil.Mutex

		lockWait time.Duration

		waiting bool

		waitStart time.Time

		curState waitingState

		numLocks int
	}
}

func newContentionEventTracer(sp *tracing.Span, clock *hlc.Clock) *contentionEventTracer {
	__antithesis_instrumentation__.Notify(100360)
	t := &contentionEventTracer{}
	t.tag.clock = clock

	oldTag, ok := sp.GetLazyTag(tagContentionTracer)
	if ok {
		__antithesis_instrumentation__.Notify(100362)
		oldContentionTag := oldTag.(*contentionTag)
		oldContentionTag.mu.Lock()
		waiting := oldContentionTag.mu.waiting
		if waiting {
			__antithesis_instrumentation__.Notify(100364)
			oldContentionTag.mu.Unlock()
			panic("span already contains contention tag in the waiting state")
		} else {
			__antithesis_instrumentation__.Notify(100365)
		}
		__antithesis_instrumentation__.Notify(100363)
		t.tag.mu.numLocks = oldContentionTag.mu.numLocks
		t.tag.mu.lockWait = oldContentionTag.mu.lockWait
		oldContentionTag.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(100366)
	}
	__antithesis_instrumentation__.Notify(100361)

	sp.SetLazyTag(tagContentionTracer, &t.tag)
	t.sp = sp
	return t
}

func (h *contentionEventTracer) SetOnContentionEvent(f func(ev *roachpb.ContentionEvent)) {
	__antithesis_instrumentation__.Notify(100367)
	h.onEvent = f
}

var _ tracing.LazyTag = &contentionTag{}

func (h *contentionEventTracer) notify(ctx context.Context, s waitingState) {
	__antithesis_instrumentation__.Notify(100368)
	if h.sp == nil {
		__antithesis_instrumentation__.Notify(100370)

		return
	} else {
		__antithesis_instrumentation__.Notify(100371)
	}
	__antithesis_instrumentation__.Notify(100369)

	event := h.tag.notify(ctx, s)
	if event != nil {
		__antithesis_instrumentation__.Notify(100372)
		h.emit(event)
	} else {
		__antithesis_instrumentation__.Notify(100373)
	}
}

func (h *contentionEventTracer) emit(event *roachpb.ContentionEvent) {
	__antithesis_instrumentation__.Notify(100374)
	if event == nil {
		__antithesis_instrumentation__.Notify(100377)
		return
	} else {
		__antithesis_instrumentation__.Notify(100378)
	}
	__antithesis_instrumentation__.Notify(100375)
	if h.onEvent != nil {
		__antithesis_instrumentation__.Notify(100379)

		h.onEvent(event)
	} else {
		__antithesis_instrumentation__.Notify(100380)
	}
	__antithesis_instrumentation__.Notify(100376)
	h.sp.RecordStructured(event)
}

func (tag *contentionTag) generateEventLocked() *roachpb.ContentionEvent {
	__antithesis_instrumentation__.Notify(100381)
	if !tag.mu.waiting {
		__antithesis_instrumentation__.Notify(100383)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(100384)
	}
	__antithesis_instrumentation__.Notify(100382)

	return &roachpb.ContentionEvent{
		Key:      tag.mu.curState.key,
		TxnMeta:  *tag.mu.curState.txn,
		Duration: tag.clock.PhysicalTime().Sub(tag.mu.curState.lockWaitStart),
	}
}

func (tag *contentionTag) notify(ctx context.Context, s waitingState) *roachpb.ContentionEvent {
	__antithesis_instrumentation__.Notify(100385)
	tag.mu.Lock()
	defer tag.mu.Unlock()

	switch s.kind {
	case waitFor, waitForDistinguished, waitSelf, waitElsewhere:
		__antithesis_instrumentation__.Notify(100387)

		var differentLock bool
		if !tag.mu.waiting {
			__antithesis_instrumentation__.Notify(100393)
			differentLock = true
		} else {
			__antithesis_instrumentation__.Notify(100394)
			curLockHolder, curKey := tag.mu.curState.txn.ID, tag.mu.curState.key
			differentLock = !curLockHolder.Equal(s.txn.ID) || func() bool {
				__antithesis_instrumentation__.Notify(100395)
				return !curKey.Equal(s.key) == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(100388)
		var res *roachpb.ContentionEvent
		if differentLock {
			__antithesis_instrumentation__.Notify(100396)
			res = tag.generateEventLocked()
		} else {
			__antithesis_instrumentation__.Notify(100397)
		}
		__antithesis_instrumentation__.Notify(100389)
		tag.mu.curState = s
		tag.mu.waiting = true
		if differentLock {
			__antithesis_instrumentation__.Notify(100398)
			tag.mu.waitStart = tag.clock.PhysicalTime()
			tag.mu.numLocks++
			return res
		} else {
			__antithesis_instrumentation__.Notify(100399)
		}
		__antithesis_instrumentation__.Notify(100390)
		return nil
	case doneWaiting, waitQueueMaxLengthExceeded:
		__antithesis_instrumentation__.Notify(100391)

		res := tag.generateEventLocked()
		tag.mu.waiting = false
		tag.mu.curState = waitingState{}
		tag.mu.lockWait += tag.clock.PhysicalTime().Sub(tag.mu.waitStart)

		tag.mu.waitStart = time.Time{}
		return res
	default:
		__antithesis_instrumentation__.Notify(100392)
		kind := s.kind
		log.Fatalf(ctx, "unhandled waitingState.kind: %v", kind)
	}
	__antithesis_instrumentation__.Notify(100386)
	panic("unreachable")
}

func (tag *contentionTag) Render() []attribute.KeyValue {
	__antithesis_instrumentation__.Notify(100400)
	tag.mu.Lock()
	defer tag.mu.Unlock()
	tags := make([]attribute.KeyValue, 0, 4)
	if tag.mu.numLocks > 0 {
		__antithesis_instrumentation__.Notify(100405)
		tags = append(tags, attribute.KeyValue{
			Key:   tagNumLocks,
			Value: attribute.IntValue(tag.mu.numLocks),
		})
	} else {
		__antithesis_instrumentation__.Notify(100406)
	}
	__antithesis_instrumentation__.Notify(100401)

	lockWait := tag.mu.lockWait
	if !tag.mu.waitStart.IsZero() {
		__antithesis_instrumentation__.Notify(100407)
		lockWait += tag.clock.PhysicalTime().Sub(tag.mu.waitStart)
	} else {
		__antithesis_instrumentation__.Notify(100408)
	}
	__antithesis_instrumentation__.Notify(100402)
	if lockWait != 0 {
		__antithesis_instrumentation__.Notify(100409)
		tags = append(tags, attribute.KeyValue{
			Key:   tagWaited,
			Value: attribute.StringValue(string(humanizeutil.Duration(lockWait))),
		})
	} else {
		__antithesis_instrumentation__.Notify(100410)
	}
	__antithesis_instrumentation__.Notify(100403)

	if tag.mu.waiting {
		__antithesis_instrumentation__.Notify(100411)
		tags = append(tags, attribute.KeyValue{
			Key:   tagWaitKey,
			Value: attribute.StringValue(tag.mu.curState.key.String()),
		})
		tags = append(tags, attribute.KeyValue{
			Key:   tagLockHolderTxn,
			Value: attribute.StringValue(tag.mu.curState.txn.ID.String()),
		})
		tags = append(tags, attribute.KeyValue{
			Key:   tagWaitStart,
			Value: attribute.StringValue(tag.mu.curState.lockWaitStart.Format("15:04:05.123")),
		})
	} else {
		__antithesis_instrumentation__.Notify(100412)
	}
	__antithesis_instrumentation__.Notify(100404)
	return tags
}

const (
	reasonWaitPolicy                 = roachpb.WriteIntentError_REASON_WAIT_POLICY
	reasonLockTimeout                = roachpb.WriteIntentError_REASON_LOCK_TIMEOUT
	reasonWaitQueueMaxLengthExceeded = roachpb.WriteIntentError_REASON_LOCK_WAIT_QUEUE_MAX_LENGTH_EXCEEDED
)

func newWriteIntentErr(
	req Request, ws waitingState, reason roachpb.WriteIntentError_Reason,
) *Error {
	__antithesis_instrumentation__.Notify(100413)
	err := roachpb.NewError(&roachpb.WriteIntentError{
		Intents: []roachpb.Intent{roachpb.MakeIntent(ws.txn, ws.key)},
		Reason:  reason,
	})

	if len(req.Requests) == 1 {
		__antithesis_instrumentation__.Notify(100415)
		err.SetErrorIndex(0)
	} else {
		__antithesis_instrumentation__.Notify(100416)
	}
	__antithesis_instrumentation__.Notify(100414)
	return err
}

func hasMinPriority(txn *enginepb.TxnMeta) bool {
	__antithesis_instrumentation__.Notify(100417)
	return txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(100418)
		return txn.Priority == enginepb.MinTxnPriority == true
	}() == true
}

func hasMaxPriority(txn *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(100419)
	return txn != nil && func() bool {
		__antithesis_instrumentation__.Notify(100420)
		return txn.Priority == enginepb.MaxTxnPriority == true
	}() == true
}

func minDuration(a, b time.Duration) time.Duration {
	__antithesis_instrumentation__.Notify(100421)
	if a < b {
		__antithesis_instrumentation__.Notify(100423)
		return a
	} else {
		__antithesis_instrumentation__.Notify(100424)
	}
	__antithesis_instrumentation__.Notify(100422)
	return b
}

var closedTimerC chan time.Time

func init() {
	closedTimerC = make(chan time.Time)
	close(closedTimerC)
}
