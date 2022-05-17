package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvbase"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/contextutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
)

const abortTxnAsyncTimeout = time.Minute

type txnHeartbeater struct {
	log.AmbientContext
	stopper      *stop.Stopper
	clock        *hlc.Clock
	metrics      *TxnMetrics
	loopInterval time.Duration

	wrapped lockedSender

	gatekeeper lockedSender

	mu struct {
		sync.Locker

		txn *roachpb.Transaction

		loopStarted bool

		loopCancel func()

		finalObservedStatus roachpb.TransactionStatus

		ifReqs uint8

		abortTxnAsyncPending bool

		abortTxnAsyncResultC chan abortTxnAsyncResult
	}
}

type abortTxnAsyncResult struct {
	br   *roachpb.BatchResponse
	pErr *roachpb.Error
}

func (h *txnHeartbeater) init(
	ac log.AmbientContext,
	stopper *stop.Stopper,
	clock *hlc.Clock,
	metrics *TxnMetrics,
	loopInterval time.Duration,
	gatekeeper lockedSender,
	mu sync.Locker,
	txn *roachpb.Transaction,
) {
	h.AmbientContext = ac
	h.stopper = stopper
	h.clock = clock
	h.metrics = metrics
	h.loopInterval = loopInterval
	h.gatekeeper = gatekeeper
	h.mu.Locker = mu
	h.mu.txn = txn
}

func (h *txnHeartbeater) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88607)
	etArg, hasET := ba.GetArg(roachpb.EndTxn)
	firstLockingIndex, pErr := firstLockingIndex(&ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88612)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88613)
	}
	__antithesis_instrumentation__.Notify(88608)
	if firstLockingIndex != -1 {
		__antithesis_instrumentation__.Notify(88614)

		if len(h.mu.txn.Key) == 0 {
			__antithesis_instrumentation__.Notify(88616)
			anchor := ba.Requests[firstLockingIndex].GetInner().Header().Key
			h.mu.txn.Key = anchor

			ba.Txn.Key = anchor
		} else {
			__antithesis_instrumentation__.Notify(88617)
		}
		__antithesis_instrumentation__.Notify(88615)

		if !h.mu.loopStarted {
			__antithesis_instrumentation__.Notify(88618)
			h.startHeartbeatLoopLocked(ctx)
		} else {
			__antithesis_instrumentation__.Notify(88619)
		}
	} else {
		__antithesis_instrumentation__.Notify(88620)
	}
	__antithesis_instrumentation__.Notify(88609)

	if hasET {
		__antithesis_instrumentation__.Notify(88621)
		et := etArg.(*roachpb.EndTxnRequest)

		if !et.Commit {
			__antithesis_instrumentation__.Notify(88622)
			h.cancelHeartbeatLoopLocked()

			if resultC := h.mu.abortTxnAsyncResultC; resultC != nil {
				__antithesis_instrumentation__.Notify(88623)

				h.mu.Unlock()
				defer h.mu.Lock()
				select {
				case res := <-resultC:
					__antithesis_instrumentation__.Notify(88624)
					return res.br, res.pErr
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(88625)
					return nil, roachpb.NewError(ctx.Err())
				}
			} else {
				__antithesis_instrumentation__.Notify(88626)
			}
		} else {
			__antithesis_instrumentation__.Notify(88627)
		}
	} else {
		__antithesis_instrumentation__.Notify(88628)
	}
	__antithesis_instrumentation__.Notify(88610)

	h.mu.ifReqs++
	br, pErr := h.wrapped.SendLocked(ctx, ba)
	h.mu.ifReqs--

	if h.mu.abortTxnAsyncPending && func() bool {
		__antithesis_instrumentation__.Notify(88629)
		return h.mu.ifReqs == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(88630)
		h.abortTxnAsyncLocked(ctx)
		h.mu.abortTxnAsyncPending = false
	} else {
		__antithesis_instrumentation__.Notify(88631)
	}
	__antithesis_instrumentation__.Notify(88611)

	return br, pErr
}

func (h *txnHeartbeater) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(88632)
	h.wrapped = wrapped
}

func (*txnHeartbeater) populateLeafInputState(*roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(88633)
}

func (*txnHeartbeater) populateLeafFinalState(*roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88634)
}

func (*txnHeartbeater) importLeafFinalState(context.Context, *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88635)
}

func (h *txnHeartbeater) epochBumpedLocked() { __antithesis_instrumentation__.Notify(88636) }

func (*txnHeartbeater) createSavepointLocked(context.Context, *savepoint) {
	__antithesis_instrumentation__.Notify(88637)
}

func (*txnHeartbeater) rollbackToSavepointLocked(context.Context, savepoint) {
	__antithesis_instrumentation__.Notify(88638)
}

func (h *txnHeartbeater) closeLocked() {
	__antithesis_instrumentation__.Notify(88639)
	h.cancelHeartbeatLoopLocked()
}

func (h *txnHeartbeater) startHeartbeatLoopLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88640)
	if h.loopInterval < 0 {
		__antithesis_instrumentation__.Notify(88644)
		log.Infof(ctx, "coordinator heartbeat loop disabled")
		return
	} else {
		__antithesis_instrumentation__.Notify(88645)
	}
	__antithesis_instrumentation__.Notify(88641)
	if h.mu.loopStarted {
		__antithesis_instrumentation__.Notify(88646)
		log.Fatal(ctx, "attempting to start a second heartbeat loop")
	} else {
		__antithesis_instrumentation__.Notify(88647)
	}
	__antithesis_instrumentation__.Notify(88642)
	log.VEventf(ctx, 2, kvbase.SpawningHeartbeatLoopMsg)
	h.mu.loopStarted = true

	h.AmbientContext.AddLogTag("txn-hb", h.mu.txn.Short())

	hbCtx, hbCancel := context.WithCancel(h.AnnotateCtx(context.Background()))

	timer := time.AfterFunc(h.loopInterval, func() {
		__antithesis_instrumentation__.Notify(88648)
		const taskName = "kv.TxnCoordSender: heartbeat loop"
		var span *tracing.Span
		hbCtx, span = h.AmbientContext.Tracer.StartSpanCtx(hbCtx, taskName)
		defer span.Finish()

		_ = h.stopper.RunTask(hbCtx, taskName, h.heartbeatLoop)
	})
	__antithesis_instrumentation__.Notify(88643)

	h.mu.loopCancel = func() {
		__antithesis_instrumentation__.Notify(88649)
		timer.Stop()
		hbCancel()
	}
}

func (h *txnHeartbeater) cancelHeartbeatLoopLocked() {
	__antithesis_instrumentation__.Notify(88650)

	if h.heartbeatLoopRunningLocked() {
		__antithesis_instrumentation__.Notify(88651)
		h.mu.loopCancel()
		h.mu.loopCancel = nil
	} else {
		__antithesis_instrumentation__.Notify(88652)
	}
}

func (h *txnHeartbeater) heartbeatLoopRunningLocked() bool {
	__antithesis_instrumentation__.Notify(88653)
	return h.mu.loopCancel != nil
}

func (h *txnHeartbeater) heartbeatLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88654)
	defer func() {
		__antithesis_instrumentation__.Notify(88658)
		h.mu.Lock()
		h.cancelHeartbeatLoopLocked()
		h.mu.Unlock()
	}()
	__antithesis_instrumentation__.Notify(88655)

	var tickChan <-chan time.Time
	{
		__antithesis_instrumentation__.Notify(88659)
		ticker := time.NewTicker(h.loopInterval)
		tickChan = ticker.C
		defer ticker.Stop()
	}
	__antithesis_instrumentation__.Notify(88656)

	if !h.heartbeat(ctx) {
		__antithesis_instrumentation__.Notify(88660)
		return
	} else {
		__antithesis_instrumentation__.Notify(88661)
	}
	__antithesis_instrumentation__.Notify(88657)

	for {
		__antithesis_instrumentation__.Notify(88662)
		select {
		case <-tickChan:
			__antithesis_instrumentation__.Notify(88663)
			if !h.heartbeat(ctx) {
				__antithesis_instrumentation__.Notify(88666)

				return
			} else {
				__antithesis_instrumentation__.Notify(88667)
			}
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(88664)

			return
		case <-h.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(88665)
			return
		}
	}
}

func (h *txnHeartbeater) heartbeat(ctx context.Context) bool {
	__antithesis_instrumentation__.Notify(88668)

	h.mu.Lock()
	defer h.mu.Unlock()

	if ctx.Err() != nil {
		__antithesis_instrumentation__.Notify(88675)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88676)
	}
	__antithesis_instrumentation__.Notify(88669)

	if h.mu.txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(88677)
		if h.mu.txn.Status == roachpb.COMMITTED {
			__antithesis_instrumentation__.Notify(88679)
			log.Fatalf(ctx, "txn committed but heartbeat loop hasn't been signaled to stop: %s", h.mu.txn)
		} else {
			__antithesis_instrumentation__.Notify(88680)
		}
		__antithesis_instrumentation__.Notify(88678)

		return false
	} else {
		__antithesis_instrumentation__.Notify(88681)
	}
	__antithesis_instrumentation__.Notify(88670)

	txn := h.mu.txn.Clone()
	if txn.Key == nil {
		__antithesis_instrumentation__.Notify(88682)
		log.Fatalf(ctx, "attempting to heartbeat txn without anchor key: %v", txn)
	} else {
		__antithesis_instrumentation__.Notify(88683)
	}
	__antithesis_instrumentation__.Notify(88671)
	ba := roachpb.BatchRequest{}
	ba.Txn = txn
	ba.Add(&roachpb.HeartbeatTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txn.Key,
		},
		Now: h.clock.Now(),
	})

	log.VEvent(ctx, 2, "heartbeat")
	br, pErr := h.gatekeeper.SendLocked(ctx, ba)

	if h.mu.txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(88684)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88685)
	}
	__antithesis_instrumentation__.Notify(88672)

	var respTxn *roachpb.Transaction
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88686)
		log.VEventf(ctx, 2, "heartbeat failed for %s: %s", h.mu.txn, pErr)

		if _, ok := pErr.GetDetail().(*roachpb.TransactionAbortedError); ok {
			__antithesis_instrumentation__.Notify(88688)

			log.VEventf(ctx, 1, "Heartbeat detected aborted txn, cleaning up for %s", h.mu.txn)
			h.abortTxnAsyncLocked(ctx)
			h.mu.finalObservedStatus = roachpb.ABORTED
			return false
		} else {
			__antithesis_instrumentation__.Notify(88689)
		}
		__antithesis_instrumentation__.Notify(88687)

		respTxn = pErr.GetTxn()
	} else {
		__antithesis_instrumentation__.Notify(88690)
		respTxn = br.Txn
	}
	__antithesis_instrumentation__.Notify(88673)

	if respTxn != nil && func() bool {
		__antithesis_instrumentation__.Notify(88691)
		return respTxn.Status.IsFinalized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(88692)
		switch respTxn.Status {
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(88694)

		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(88695)

			log.VEventf(ctx, 1, "Heartbeat detected aborted txn, cleaning up for %s", h.mu.txn)
			h.abortTxnAsyncLocked(ctx)
		default:
			__antithesis_instrumentation__.Notify(88696)
		}
		__antithesis_instrumentation__.Notify(88693)
		h.mu.finalObservedStatus = respTxn.Status
		return false
	} else {
		__antithesis_instrumentation__.Notify(88697)
	}
	__antithesis_instrumentation__.Notify(88674)
	return true
}

func (h *txnHeartbeater) abortTxnAsyncLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88698)

	if h.mu.ifReqs > 0 {
		__antithesis_instrumentation__.Notify(88700)
		h.mu.abortTxnAsyncPending = true
		log.VEventf(ctx, 2, "async abort waiting for in-flight request for txn %s", h.mu.txn)
		return
	} else {
		__antithesis_instrumentation__.Notify(88701)
	}
	__antithesis_instrumentation__.Notify(88699)

	txn := h.mu.txn.Clone()
	ba := roachpb.BatchRequest{}
	ba.Header = roachpb.Header{Txn: txn}
	ba.Add(&roachpb.EndTxnRequest{
		Commit: false,

		Poison: true,
	})

	const taskName = "txnHeartbeater: aborting txn"
	log.VEventf(ctx, 2, "async abort for txn: %s", txn)
	if err := h.stopper.RunAsyncTask(h.AnnotateCtx(context.Background()), taskName,
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(88702)
			if err := contextutil.RunWithTimeout(ctx, taskName, abortTxnAsyncTimeout,
				func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(88703)
					h.mu.Lock()
					defer h.mu.Unlock()

					if h.mu.abortTxnAsyncResultC != nil {
						__antithesis_instrumentation__.Notify(88707)
						log.VEventf(ctx, 2,
							"skipping async abort due to concurrent async abort for %s", txn)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(88708)
					}
					__antithesis_instrumentation__.Notify(88704)

					if h.mu.ifReqs > 0 {
						__antithesis_instrumentation__.Notify(88709)
						log.VEventf(ctx, 2,
							"skipping async abort due to client rollback for %s", txn)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(88710)
					}
					__antithesis_instrumentation__.Notify(88705)

					h.mu.abortTxnAsyncResultC = make(chan abortTxnAsyncResult, 1)

					br, pErr := h.wrapped.SendLocked(ctx, ba)
					if pErr != nil {
						__antithesis_instrumentation__.Notify(88711)
						log.VErrEventf(ctx, 1, "async abort failed for %s: %s ", txn, pErr)
						h.metrics.AsyncRollbacksFailed.Inc(1)
					} else {
						__antithesis_instrumentation__.Notify(88712)
					}
					__antithesis_instrumentation__.Notify(88706)

					h.mu.abortTxnAsyncResultC <- abortTxnAsyncResult{br: br, pErr: pErr}
					h.mu.abortTxnAsyncResultC = nil
					return nil
				},
			); err != nil {
				__antithesis_instrumentation__.Notify(88713)
				log.VEventf(ctx, 1, "async abort failed for %s: %s", txn, err)
			} else {
				__antithesis_instrumentation__.Notify(88714)
			}
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(88715)
		log.Warningf(ctx, "%v", err)
		h.metrics.AsyncRollbacksFailed.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(88716)
	}
}

func firstLockingIndex(ba *roachpb.BatchRequest) (int, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88717)
	for i, ru := range ba.Requests {
		__antithesis_instrumentation__.Notify(88719)
		args := ru.GetInner()
		if i < len(ba.Requests)-1 {
			__antithesis_instrumentation__.Notify(88721)
			if _, ok := args.(*roachpb.EndTxnRequest); ok {
				__antithesis_instrumentation__.Notify(88722)
				return -1, roachpb.NewErrorf("%s sent as non-terminal call", args.Method())
			} else {
				__antithesis_instrumentation__.Notify(88723)
			}
		} else {
			__antithesis_instrumentation__.Notify(88724)
		}
		__antithesis_instrumentation__.Notify(88720)
		if roachpb.IsLocking(args) {
			__antithesis_instrumentation__.Notify(88725)
			return i, nil
		} else {
			__antithesis_instrumentation__.Notify(88726)
		}
	}
	__antithesis_instrumentation__.Notify(88718)
	return -1, nil
}
