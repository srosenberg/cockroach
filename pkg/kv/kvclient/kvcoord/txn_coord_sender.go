package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"runtime/debug"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"go.opentelemetry.io/otel/attribute"
)

const (
	OpTxnCoordSender = "txn coordinator send"
)

var DisableCommitSanityCheck = envutil.EnvOrDefaultBool("COCKROACH_DISABLE_COMMIT_SANITY_CHECK", false)

type txnState int

const (
	txnPending txnState = iota

	txnRetryableError

	txnError

	txnFinalized
)

type TxnCoordSender struct {
	mu struct {
		syncutil.Mutex

		txnState txnState

		storedRetryableErr *roachpb.TransactionRetryWithProtoRefreshError

		storedErr *roachpb.Error

		active bool

		closed bool

		txn roachpb.Transaction

		userPriority roachpb.UserPriority

		commitWaitDeferred bool
	}

	*TxnCoordSenderFactory

	interceptorStack []txnInterceptor
	interceptorAlloc struct {
		arr [6]txnInterceptor
		txnHeartbeater
		txnSeqNumAllocator
		txnPipeliner
		txnSpanRefresher
		txnCommitter
		txnMetricRecorder
		txnLockGatekeeper
	}

	typ kv.TxnType
}

var _ kv.TxnSender = &TxnCoordSender{}

type txnInterceptor interface {
	lockedSender

	setWrapped(wrapped lockedSender)

	populateLeafInputState(*roachpb.LeafTxnInputState)

	populateLeafFinalState(*roachpb.LeafTxnFinalState)

	importLeafFinalState(context.Context, *roachpb.LeafTxnFinalState)

	epochBumpedLocked()

	createSavepointLocked(context.Context, *savepoint)

	rollbackToSavepointLocked(context.Context, savepoint)

	closeLocked()
}

func newRootTxnCoordSender(
	tcf *TxnCoordSenderFactory, txn *roachpb.Transaction, pri roachpb.UserPriority,
) kv.TxnSender {
	__antithesis_instrumentation__.Notify(88106)
	txn.AssertInitialized(context.TODO())

	if txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(88109)
		log.Fatalf(context.TODO(), "unexpected non-pending txn in RootTransactionalSender: %s", txn)
	} else {
		__antithesis_instrumentation__.Notify(88110)
	}
	__antithesis_instrumentation__.Notify(88107)
	if txn.Sequence != 0 {
		__antithesis_instrumentation__.Notify(88111)
		log.Fatalf(context.TODO(), "cannot initialize root txn with seq != 0: %s", txn)
	} else {
		__antithesis_instrumentation__.Notify(88112)
	}
	__antithesis_instrumentation__.Notify(88108)

	tcs := &TxnCoordSender{
		typ:                   kv.RootTxn,
		TxnCoordSenderFactory: tcf,
	}
	tcs.mu.txnState = txnPending
	tcs.mu.userPriority = pri

	tcs.interceptorAlloc.txnHeartbeater.init(
		tcf.AmbientContext,
		tcs.stopper,
		tcs.clock,
		&tcs.metrics,
		tcs.heartbeatInterval,
		&tcs.interceptorAlloc.txnLockGatekeeper,
		&tcs.mu.Mutex,
		&tcs.mu.txn,
	)
	tcs.interceptorAlloc.txnCommitter = txnCommitter{
		st:      tcf.st,
		stopper: tcs.stopper,
		mu:      &tcs.mu.Mutex,
	}
	tcs.interceptorAlloc.txnMetricRecorder = txnMetricRecorder{
		metrics: &tcs.metrics,
		clock:   tcs.clock,
		txn:     &tcs.mu.txn,
	}
	tcs.initCommonInterceptors(tcf, txn, kv.RootTxn)

	tcs.interceptorAlloc.arr = [...]txnInterceptor{
		&tcs.interceptorAlloc.txnHeartbeater,

		&tcs.interceptorAlloc.txnSeqNumAllocator,

		&tcs.interceptorAlloc.txnPipeliner,

		&tcs.interceptorAlloc.txnSpanRefresher,

		&tcs.interceptorAlloc.txnCommitter,

		&tcs.interceptorAlloc.txnMetricRecorder,
	}
	tcs.interceptorStack = tcs.interceptorAlloc.arr[:]

	tcs.connectInterceptors()

	tcs.mu.txn.Update(txn)
	return tcs
}

func (tc *TxnCoordSender) initCommonInterceptors(
	tcf *TxnCoordSenderFactory, txn *roachpb.Transaction, typ kv.TxnType,
) {
	__antithesis_instrumentation__.Notify(88113)
	var riGen rangeIteratorFactory
	if ds, ok := tcf.wrapped.(*DistSender); ok {
		__antithesis_instrumentation__.Notify(88115)
		riGen.ds = ds
	} else {
		__antithesis_instrumentation__.Notify(88116)
	}
	__antithesis_instrumentation__.Notify(88114)
	tc.interceptorAlloc.txnPipeliner = txnPipeliner{
		st:                     tcf.st,
		riGen:                  riGen,
		txnMetrics:             &tc.metrics,
		condensedIntentsEveryN: &tc.TxnCoordSenderFactory.condensedIntentsEveryN,
	}
	tc.interceptorAlloc.txnSpanRefresher = txnSpanRefresher{
		st:    tcf.st,
		knobs: &tcf.testingKnobs,
		riGen: riGen,

		canAutoRetry:                  typ == kv.RootTxn,
		refreshSuccess:                tc.metrics.RefreshSuccess,
		refreshFail:                   tc.metrics.RefreshFail,
		refreshFailWithCondensedSpans: tc.metrics.RefreshFailWithCondensedSpans,
		refreshMemoryLimitExceeded:    tc.metrics.RefreshMemoryLimitExceeded,
		refreshAutoRetries:            tc.metrics.RefreshAutoRetries,
	}
	tc.interceptorAlloc.txnLockGatekeeper = txnLockGatekeeper{
		wrapped:                 tc.wrapped,
		mu:                      &tc.mu.Mutex,
		allowConcurrentRequests: typ == kv.LeafTxn,
	}
	tc.interceptorAlloc.txnSeqNumAllocator.writeSeq = txn.Sequence
}

func (tc *TxnCoordSender) connectInterceptors() {
	__antithesis_instrumentation__.Notify(88117)
	for i, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88118)
		if i < len(tc.interceptorStack)-1 {
			__antithesis_instrumentation__.Notify(88119)
			reqInt.setWrapped(tc.interceptorStack[i+1])
		} else {
			__antithesis_instrumentation__.Notify(88120)
			reqInt.setWrapped(&tc.interceptorAlloc.txnLockGatekeeper)
		}
	}
}

func newLeafTxnCoordSender(
	tcf *TxnCoordSenderFactory, tis *roachpb.LeafTxnInputState,
) kv.TxnSender {
	__antithesis_instrumentation__.Notify(88121)
	txn := &tis.Txn

	txn.WriteTooOld = false
	txn.AssertInitialized(context.TODO())

	if txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(88124)
		log.Fatalf(context.TODO(), "unexpected non-pending txn in LeafTransactionalSender: %s", tis)
	} else {
		__antithesis_instrumentation__.Notify(88125)
	}
	__antithesis_instrumentation__.Notify(88122)

	tcs := &TxnCoordSender{
		typ:                   kv.LeafTxn,
		TxnCoordSenderFactory: tcf,
	}
	tcs.mu.txnState = txnPending

	tcs.initCommonInterceptors(tcf, txn, kv.LeafTxn)

	tcs.interceptorAlloc.txnPipeliner.initializeLeaf(tis)
	tcs.interceptorAlloc.txnSeqNumAllocator.initializeLeaf(tis)

	tcs.interceptorAlloc.arr = [cap(tcs.interceptorAlloc.arr)]txnInterceptor{

		&tcs.interceptorAlloc.txnSeqNumAllocator,

		&tcs.interceptorAlloc.txnPipeliner,

		&tcs.interceptorAlloc.txnSpanRefresher,
	}

	if tis.RefreshInvalid {
		__antithesis_instrumentation__.Notify(88126)
		tcs.interceptorStack = tcs.interceptorAlloc.arr[:2]
	} else {
		__antithesis_instrumentation__.Notify(88127)
		tcs.interceptorStack = tcs.interceptorAlloc.arr[:3]
	}
	__antithesis_instrumentation__.Notify(88123)

	tcs.connectInterceptors()

	tcs.mu.txn.Update(txn)
	return tcs
}

func (tc *TxnCoordSender) DisablePipelining() error {
	__antithesis_instrumentation__.Notify(88128)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.mu.active {
		__antithesis_instrumentation__.Notify(88130)
		return errors.Errorf("cannot disable pipelining on a running transaction")
	} else {
		__antithesis_instrumentation__.Notify(88131)
	}
	__antithesis_instrumentation__.Notify(88129)
	tc.interceptorAlloc.txnPipeliner.disabled = true
	return nil
}

func generateTxnDeadlineExceededErr(
	txn *roachpb.Transaction, deadline hlc.Timestamp,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88132)
	exceededBy := txn.WriteTimestamp.GoTime().Sub(deadline.GoTime())
	extraMsg := fmt.Sprintf(
		"txn timestamp pushed too much; deadline exceeded by %s (%s > %s)",
		exceededBy, txn.WriteTimestamp, deadline)
	return roachpb.NewErrorWithTxn(
		roachpb.NewTransactionRetryError(roachpb.RETRY_COMMIT_DEADLINE_EXCEEDED, extraMsg), txn)
}

func (tc *TxnCoordSender) finalizeNonLockingTxnLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88133)
	et := ba.Requests[0].GetEndTxn()
	if et.Commit {
		__antithesis_instrumentation__.Notify(88136)
		deadline := et.Deadline
		if deadline != nil && func() bool {
			__antithesis_instrumentation__.Notify(88138)
			return !deadline.IsEmpty() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(88139)
			return deadline.LessEq(tc.mu.txn.WriteTimestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(88140)
			txn := tc.mu.txn.Clone()
			pErr := generateTxnDeadlineExceededErr(txn, *deadline)

			ba.Txn = txn
			return tc.updateStateLocked(ctx, ba, nil, pErr)
		} else {
			__antithesis_instrumentation__.Notify(88141)
		}
		__antithesis_instrumentation__.Notify(88137)

		tc.mu.txn.Status = roachpb.COMMITTED
	} else {
		__antithesis_instrumentation__.Notify(88142)
		tc.mu.txn.Status = roachpb.ABORTED
	}
	__antithesis_instrumentation__.Notify(88134)
	tc.finalizeAndCleanupTxnLocked(ctx)
	if et.Commit {
		__antithesis_instrumentation__.Notify(88143)
		if err := tc.maybeCommitWait(ctx, false); err != nil {
			__antithesis_instrumentation__.Notify(88144)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(88145)
		}
	} else {
		__antithesis_instrumentation__.Notify(88146)
	}
	__antithesis_instrumentation__.Notify(88135)
	return nil
}

func (tc *TxnCoordSender) Send(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88147)

	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.mu.active = true

	if pErr := tc.maybeRejectClientLocked(ctx, &ba); pErr != nil {
		__antithesis_instrumentation__.Notify(88157)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88158)
	}
	__antithesis_instrumentation__.Notify(88148)

	if ba.IsSingleEndTxnRequest() && func() bool {
		__antithesis_instrumentation__.Notify(88159)
		return !tc.interceptorAlloc.txnPipeliner.hasAcquiredLocks() == true
	}() == true {
		__antithesis_instrumentation__.Notify(88160)
		return nil, tc.finalizeNonLockingTxnLocked(ctx, ba)
	} else {
		__antithesis_instrumentation__.Notify(88161)
	}
	__antithesis_instrumentation__.Notify(88149)

	ctx, sp := tc.AnnotateCtxWithSpan(ctx, OpTxnCoordSender)
	defer sp.Finish()

	if tc.mu.txn.ID == (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(88162)
		log.Fatalf(ctx, "cannot send transactional request through unbound TxnCoordSender")
	} else {
		__antithesis_instrumentation__.Notify(88163)
	}
	__antithesis_instrumentation__.Notify(88150)
	if sp.IsVerbose() {
		__antithesis_instrumentation__.Notify(88164)
		sp.SetTag("txnID", attribute.StringValue(tc.mu.txn.ID.String()))
		ctx = logtags.AddTag(ctx, "txn", uuid.ShortStringer(tc.mu.txn.ID))
		if log.V(2) {
			__antithesis_instrumentation__.Notify(88165)
			ctx = logtags.AddTag(ctx, "ts", tc.mu.txn.WriteTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(88166)
		}
	} else {
		__antithesis_instrumentation__.Notify(88167)
	}
	__antithesis_instrumentation__.Notify(88151)

	if ba.ReadConsistency != roachpb.CONSISTENT {
		__antithesis_instrumentation__.Notify(88168)
		return nil, roachpb.NewErrorf("cannot use %s ReadConsistency in txn",
			ba.ReadConsistency)
	} else {
		__antithesis_instrumentation__.Notify(88169)
	}
	__antithesis_instrumentation__.Notify(88152)

	lastIndex := len(ba.Requests) - 1
	if lastIndex < 0 {
		__antithesis_instrumentation__.Notify(88170)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(88171)
	}
	__antithesis_instrumentation__.Notify(88153)

	ba.Txn = tc.mu.txn.Clone()

	br, pErr := tc.interceptorStack[0].SendLocked(ctx, ba)

	pErr = tc.updateStateLocked(ctx, ba, br, pErr)

	if req, ok := ba.GetArg(roachpb.EndTxn); ok {
		__antithesis_instrumentation__.Notify(88172)
		et := req.(*roachpb.EndTxnRequest)
		if (et.Commit && func() bool {
			__antithesis_instrumentation__.Notify(88173)
			return pErr == nil == true
		}() == true) || func() bool {
			__antithesis_instrumentation__.Notify(88174)
			return !et.Commit == true
		}() == true {
			__antithesis_instrumentation__.Notify(88175)
			tc.finalizeAndCleanupTxnLocked(ctx)
			if et.Commit {
				__antithesis_instrumentation__.Notify(88176)
				if err := tc.maybeCommitWait(ctx, false); err != nil {
					__antithesis_instrumentation__.Notify(88177)
					return nil, roachpb.NewError(err)
				} else {
					__antithesis_instrumentation__.Notify(88178)
				}
			} else {
				__antithesis_instrumentation__.Notify(88179)
			}
		} else {
			__antithesis_instrumentation__.Notify(88180)
		}
	} else {
		__antithesis_instrumentation__.Notify(88181)
	}
	__antithesis_instrumentation__.Notify(88154)

	if pErr != nil {
		__antithesis_instrumentation__.Notify(88182)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88183)
	}
	__antithesis_instrumentation__.Notify(88155)

	if br != nil && func() bool {
		__antithesis_instrumentation__.Notify(88184)
		return br.Error != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(88185)
		panic(roachpb.ErrorUnexpectedlySet(nil, br))
	} else {
		__antithesis_instrumentation__.Notify(88186)
	}
	__antithesis_instrumentation__.Notify(88156)

	return br, nil
}

func (tc *TxnCoordSender) maybeCommitWait(ctx context.Context, deferred bool) error {
	__antithesis_instrumentation__.Notify(88187)
	if tc.mu.txn.Status != roachpb.COMMITTED {
		__antithesis_instrumentation__.Notify(88193)
		log.Fatalf(ctx, "maybeCommitWait called when not committed")
	} else {
		__antithesis_instrumentation__.Notify(88194)
	}
	__antithesis_instrumentation__.Notify(88188)
	if tc.mu.commitWaitDeferred && func() bool {
		__antithesis_instrumentation__.Notify(88195)
		return !deferred == true
	}() == true {
		__antithesis_instrumentation__.Notify(88196)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(88197)
	}
	__antithesis_instrumentation__.Notify(88189)

	commitTS := tc.mu.txn.WriteTimestamp
	readOnly := tc.mu.txn.Sequence == 0
	linearizable := tc.linearizable
	needWait := commitTS.Synthetic || func() bool {
		__antithesis_instrumentation__.Notify(88198)
		return (linearizable && func() bool {
			__antithesis_instrumentation__.Notify(88199)
			return !readOnly == true
		}() == true) == true
	}() == true
	if !needWait {
		__antithesis_instrumentation__.Notify(88200)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(88201)
	}
	__antithesis_instrumentation__.Notify(88190)

	waitUntil := commitTS
	if linearizable && func() bool {
		__antithesis_instrumentation__.Notify(88202)
		return !readOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(88203)
		waitUntil = waitUntil.Add(tc.clock.MaxOffset().Nanoseconds(), 0)
	} else {
		__antithesis_instrumentation__.Notify(88204)
	}
	__antithesis_instrumentation__.Notify(88191)

	before := tc.clock.PhysicalTime()
	est := waitUntil.GoTime().Sub(before)
	log.VEventf(ctx, 2, "performing commit-wait sleep for ~%s", est)

	tc.mu.Unlock()
	err := tc.clock.SleepUntil(ctx, waitUntil)
	tc.mu.Lock()
	if err != nil {
		__antithesis_instrumentation__.Notify(88205)
		return err
	} else {
		__antithesis_instrumentation__.Notify(88206)
	}
	__antithesis_instrumentation__.Notify(88192)

	after := tc.clock.PhysicalTime()
	log.VEventf(ctx, 2, "completed commit-wait sleep, took %s", after.Sub(before))
	tc.metrics.CommitWaits.Inc(1)
	return nil
}

func (tc *TxnCoordSender) maybeRejectClientLocked(
	ctx context.Context, ba *roachpb.BatchRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88207)
	if ba != nil && func() bool {
		__antithesis_instrumentation__.Notify(88211)
		return ba.IsSingleAbortTxnRequest() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(88212)
		return tc.mu.txn.Status != roachpb.COMMITTED == true
	}() == true {
		__antithesis_instrumentation__.Notify(88213)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(88214)
	}
	__antithesis_instrumentation__.Notify(88208)

	switch tc.mu.txnState {
	case txnPending:
		__antithesis_instrumentation__.Notify(88215)

	case txnRetryableError:
		__antithesis_instrumentation__.Notify(88216)
		return roachpb.NewError(tc.mu.storedRetryableErr)
	case txnError:
		__antithesis_instrumentation__.Notify(88217)
		return tc.mu.storedErr
	case txnFinalized:
		__antithesis_instrumentation__.Notify(88218)
		msg := fmt.Sprintf("client already committed or rolled back the transaction. "+
			"Trying to execute: %s", ba.Summary())
		stack := string(debug.Stack())
		log.Errorf(ctx, "%s. stack:\n%s", msg, stack)
		reason := roachpb.TransactionStatusError_REASON_UNKNOWN
		if tc.mu.txn.Status == roachpb.COMMITTED {
			__antithesis_instrumentation__.Notify(88221)
			reason = roachpb.TransactionStatusError_REASON_TXN_COMMITTED
		} else {
			__antithesis_instrumentation__.Notify(88222)
		}
		__antithesis_instrumentation__.Notify(88219)
		return roachpb.NewErrorWithTxn(roachpb.NewTransactionStatusError(reason, msg), &tc.mu.txn)
	default:
		__antithesis_instrumentation__.Notify(88220)
	}
	__antithesis_instrumentation__.Notify(88209)

	protoStatus := tc.mu.txn.Status
	hbObservedStatus := tc.interceptorAlloc.txnHeartbeater.mu.finalObservedStatus
	switch {
	case protoStatus == roachpb.ABORTED:
		__antithesis_instrumentation__.Notify(88223)

		fallthrough
	case protoStatus != roachpb.COMMITTED && func() bool {
		__antithesis_instrumentation__.Notify(88227)
		return hbObservedStatus == roachpb.ABORTED == true
	}() == true:
		__antithesis_instrumentation__.Notify(88224)

		abortedErr := roachpb.NewErrorWithTxn(
			roachpb.NewTransactionAbortedError(roachpb.ABORT_REASON_CLIENT_REJECT), &tc.mu.txn)
		return roachpb.NewError(tc.handleRetryableErrLocked(ctx, abortedErr))
	case protoStatus != roachpb.PENDING || func() bool {
		__antithesis_instrumentation__.Notify(88228)
		return hbObservedStatus != roachpb.PENDING == true
	}() == true:
		__antithesis_instrumentation__.Notify(88225)

		return roachpb.NewErrorf(
			"unexpected txn state: %s; heartbeat observed status: %s", tc.mu.txn, hbObservedStatus)
	default:
		__antithesis_instrumentation__.Notify(88226)

	}
	__antithesis_instrumentation__.Notify(88210)
	return nil
}

func (tc *TxnCoordSender) finalizeAndCleanupTxnLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88229)
	tc.mu.txnState = txnFinalized
	tc.cleanupTxnLocked(ctx)
}

func (tc *TxnCoordSender) cleanupTxnLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88230)
	if tc.mu.closed {
		__antithesis_instrumentation__.Notify(88232)
		return
	} else {
		__antithesis_instrumentation__.Notify(88233)
	}
	__antithesis_instrumentation__.Notify(88231)
	tc.mu.closed = true

	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88234)
		reqInt.closeLocked()
	}
}

func (tc *TxnCoordSender) UpdateStateOnRemoteRetryableErr(
	ctx context.Context, pErr *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88235)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return roachpb.NewError(tc.handleRetryableErrLocked(ctx, pErr))
}

func (tc *TxnCoordSender) handleRetryableErrLocked(
	ctx context.Context, pErr *roachpb.Error,
) *roachpb.TransactionRetryWithProtoRefreshError {
	__antithesis_instrumentation__.Notify(88236)

	switch tErr := pErr.GetDetail().(type) {
	case *roachpb.TransactionRetryError:
		__antithesis_instrumentation__.Notify(88240)
		switch tErr.Reason {
		case roachpb.RETRY_WRITE_TOO_OLD:
			__antithesis_instrumentation__.Notify(88246)
			tc.metrics.RestartsWriteTooOld.Inc()
		case roachpb.RETRY_SERIALIZABLE:
			__antithesis_instrumentation__.Notify(88247)
			tc.metrics.RestartsSerializable.Inc()
		case roachpb.RETRY_ASYNC_WRITE_FAILURE:
			__antithesis_instrumentation__.Notify(88248)
			tc.metrics.RestartsAsyncWriteFailure.Inc()
		case roachpb.RETRY_COMMIT_DEADLINE_EXCEEDED:
			__antithesis_instrumentation__.Notify(88249)
			tc.metrics.RestartsCommitDeadlineExceeded.Inc()
		default:
			__antithesis_instrumentation__.Notify(88250)
			tc.metrics.RestartsUnknown.Inc()
		}

	case *roachpb.WriteTooOldError:
		__antithesis_instrumentation__.Notify(88241)
		tc.metrics.RestartsWriteTooOldMulti.Inc()

	case *roachpb.ReadWithinUncertaintyIntervalError:
		__antithesis_instrumentation__.Notify(88242)
		tc.metrics.RestartsReadWithinUncertainty.Inc()

	case *roachpb.TransactionAbortedError:
		__antithesis_instrumentation__.Notify(88243)
		tc.metrics.RestartsTxnAborted.Inc()

	case *roachpb.TransactionPushError:
		__antithesis_instrumentation__.Notify(88244)
		tc.metrics.RestartsTxnPush.Inc()

	default:
		__antithesis_instrumentation__.Notify(88245)
		tc.metrics.RestartsUnknown.Inc()
	}
	__antithesis_instrumentation__.Notify(88237)
	errTxnID := pErr.GetTxn().ID
	newTxn := roachpb.PrepareTransactionForRetry(ctx, pErr, tc.mu.userPriority, tc.clock)

	retErr := roachpb.NewTransactionRetryWithProtoRefreshError(
		pErr.String(),
		errTxnID,
		newTxn)

	tc.mu.txnState = txnRetryableError
	tc.mu.storedRetryableErr = retErr

	if errTxnID != newTxn.ID {
		__antithesis_instrumentation__.Notify(88251)

		tc.mu.txn.Status = roachpb.ABORTED

		tc.interceptorAlloc.txnHeartbeater.abortTxnAsyncLocked(ctx)
		tc.cleanupTxnLocked(ctx)
		return retErr
	} else {
		__antithesis_instrumentation__.Notify(88252)
	}
	__antithesis_instrumentation__.Notify(88238)

	tc.mu.txn.Update(&newTxn)

	log.VEventf(ctx, 2, "resetting epoch-based coordinator state on retry")
	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88253)
		reqInt.epochBumpedLocked()
	}
	__antithesis_instrumentation__.Notify(88239)
	return retErr
}

func (tc *TxnCoordSender) updateStateLocked(
	ctx context.Context, ba roachpb.BatchRequest, br *roachpb.BatchResponse, pErr *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88254)

	if pErr == nil {
		__antithesis_instrumentation__.Notify(88259)
		tc.mu.txn.Update(br.Txn)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88260)
	}
	__antithesis_instrumentation__.Notify(88255)

	if pErr.TransactionRestart() != roachpb.TransactionRestart_NONE {
		__antithesis_instrumentation__.Notify(88261)
		if tc.typ == kv.LeafTxn {
			__antithesis_instrumentation__.Notify(88264)

			tc.mu.txnState = txnError
			tc.mu.storedErr = pErr

			tc.mu.txn.Update(pErr.GetTxn())
			tc.cleanupTxnLocked(ctx)
			return pErr
		} else {
			__antithesis_instrumentation__.Notify(88265)
		}
		__antithesis_instrumentation__.Notify(88262)

		txnID := ba.Txn.ID
		errTxnID := pErr.GetTxn().ID
		if errTxnID != txnID {
			__antithesis_instrumentation__.Notify(88266)

			log.Fatalf(ctx, "retryable error for the wrong txn. ba.Txn: %s. pErr: %s",
				ba.Txn, pErr)
		} else {
			__antithesis_instrumentation__.Notify(88267)
		}
		__antithesis_instrumentation__.Notify(88263)
		return roachpb.NewError(tc.handleRetryableErrLocked(ctx, pErr))
	} else {
		__antithesis_instrumentation__.Notify(88268)
	}
	__antithesis_instrumentation__.Notify(88256)

	if roachpb.ErrPriority(pErr.GoError()) != roachpb.ErrorScoreUnambiguousError {
		__antithesis_instrumentation__.Notify(88269)
		tc.mu.txnState = txnError
		tc.mu.storedErr = roachpb.NewError(&roachpb.TxnAlreadyEncounteredErrorError{
			PrevError: pErr.String(),
		})
	} else {
		__antithesis_instrumentation__.Notify(88270)
	}
	__antithesis_instrumentation__.Notify(88257)

	if errTxn := pErr.GetTxn(); errTxn != nil {
		__antithesis_instrumentation__.Notify(88271)
		if err := sanityCheckErrWithTxn(ctx, pErr, ba, &tc.testingKnobs); err != nil {
			__antithesis_instrumentation__.Notify(88273)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(88274)
		}
		__antithesis_instrumentation__.Notify(88272)
		tc.mu.txn.Update(errTxn)
	} else {
		__antithesis_instrumentation__.Notify(88275)
	}
	__antithesis_instrumentation__.Notify(88258)
	return pErr
}

func sanityCheckErrWithTxn(
	ctx context.Context,
	pErrWithTxn *roachpb.Error,
	ba roachpb.BatchRequest,
	knobs *ClientTestingKnobs,
) error {
	__antithesis_instrumentation__.Notify(88276)
	txn := pErrWithTxn.GetTxn()
	if txn.Status != roachpb.COMMITTED {
		__antithesis_instrumentation__.Notify(88280)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88281)
	}
	__antithesis_instrumentation__.Notify(88277)

	if ba.IsSingleAbortTxnRequest() {
		__antithesis_instrumentation__.Notify(88282)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88283)
	}
	__antithesis_instrumentation__.Notify(88278)

	err := errors.Wrapf(pErrWithTxn.GoError(),
		"transaction unexpectedly committed, ba: %s. txn: %s",
		ba, pErrWithTxn.GetTxn(),
	)
	err = errors.WithAssertionFailure(
		errors.WithIssueLink(err, errors.IssueLink{
			IssueURL: "https://github.com/cockroachdb/cockroach/issues/67765",
			Detail: "you have encountered a known bug in CockroachDB, please consider " +
				"reporting on the Github issue or reach out via Support. " +
				"This assertion can be disabled by setting the environment variable " +
				"COCKROACH_DISABLE_COMMIT_SANITY_CHECK=true",
		}))
	if !DisableCommitSanityCheck && func() bool {
		__antithesis_instrumentation__.Notify(88284)
		return !knobs.DisableCommitSanityCheck == true
	}() == true {
		__antithesis_instrumentation__.Notify(88285)
		log.Fatalf(ctx, "%s", err)
	} else {
		__antithesis_instrumentation__.Notify(88286)
	}
	__antithesis_instrumentation__.Notify(88279)
	return err
}

func (tc *TxnCoordSender) setTxnAnchorKeyLocked(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(88287)
	if len(tc.mu.txn.Key) != 0 {
		__antithesis_instrumentation__.Notify(88289)
		return errors.Errorf("transaction anchor key already set")
	} else {
		__antithesis_instrumentation__.Notify(88290)
	}
	__antithesis_instrumentation__.Notify(88288)
	tc.mu.txn.Key = key
	return nil
}

func (tc *TxnCoordSender) AnchorOnSystemConfigRange() error {
	__antithesis_instrumentation__.Notify(88291)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if bytes.Equal(tc.mu.txn.Key, keys.SystemConfigSpan.Key) {
		__antithesis_instrumentation__.Notify(88293)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88294)
	}
	__antithesis_instrumentation__.Notify(88292)

	return tc.setTxnAnchorKeyLocked(keys.SystemConfigSpan.Key)
}

func (tc *TxnCoordSender) TxnStatus() roachpb.TransactionStatus {
	__antithesis_instrumentation__.Notify(88295)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.Status
}

func (tc *TxnCoordSender) SetUserPriority(pri roachpb.UserPriority) error {
	__antithesis_instrumentation__.Notify(88296)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.mu.active && func() bool {
		__antithesis_instrumentation__.Notify(88298)
		return pri != tc.mu.userPriority == true
	}() == true {
		__antithesis_instrumentation__.Notify(88299)
		return errors.New("cannot change the user priority of a running transaction")
	} else {
		__antithesis_instrumentation__.Notify(88300)
	}
	__antithesis_instrumentation__.Notify(88297)
	tc.mu.userPriority = pri
	tc.mu.txn.Priority = roachpb.MakePriority(pri)
	return nil
}

func (tc *TxnCoordSender) SetDebugName(name string) {
	__antithesis_instrumentation__.Notify(88301)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.mu.txn.Name == name {
		__antithesis_instrumentation__.Notify(88304)
		return
	} else {
		__antithesis_instrumentation__.Notify(88305)
	}
	__antithesis_instrumentation__.Notify(88302)

	if tc.mu.active {
		__antithesis_instrumentation__.Notify(88306)
		panic("cannot change the debug name of a running transaction")
	} else {
		__antithesis_instrumentation__.Notify(88307)
	}
	__antithesis_instrumentation__.Notify(88303)
	tc.mu.txn.Name = name
}

func (tc *TxnCoordSender) String() string {
	__antithesis_instrumentation__.Notify(88308)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.String()
}

func (tc *TxnCoordSender) ReadTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(88309)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.ReadTimestamp
}

func (tc *TxnCoordSender) ProvisionalCommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(88310)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.WriteTimestamp
}

func (tc *TxnCoordSender) CommitTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(88311)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	txn := &tc.mu.txn
	tc.mu.txn.CommitTimestampFixed = true
	return txn.ReadTimestamp
}

func (tc *TxnCoordSender) CommitTimestampFixed() bool {
	__antithesis_instrumentation__.Notify(88312)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.CommitTimestampFixed
}

func (tc *TxnCoordSender) SetFixedTimestamp(ctx context.Context, ts hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(88313)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if !tc.interceptorAlloc.txnSpanRefresher.refreshFootprint.empty() {
		__antithesis_instrumentation__.Notify(88316)
		return errors.WithContextTags(errors.AssertionFailedf(
			"cannot set fixed timestamp, txn %s already performed reads", tc.mu.txn), ctx)
	} else {
		__antithesis_instrumentation__.Notify(88317)
	}
	__antithesis_instrumentation__.Notify(88314)
	if tc.mu.txn.Sequence != 0 {
		__antithesis_instrumentation__.Notify(88318)
		return errors.WithContextTags(errors.AssertionFailedf(
			"cannot set fixed timestamp, txn %s already performed writes", tc.mu.txn), ctx)
	} else {
		__antithesis_instrumentation__.Notify(88319)
	}
	__antithesis_instrumentation__.Notify(88315)

	tc.mu.txn.ReadTimestamp = ts
	tc.mu.txn.WriteTimestamp = ts
	tc.mu.txn.GlobalUncertaintyLimit = ts
	tc.mu.txn.CommitTimestampFixed = true

	tc.mu.txn.MinTimestamp.Backward(ts)
	return nil
}

func (tc *TxnCoordSender) RequiredFrontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(88320)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.RequiredFrontier()
}

func (tc *TxnCoordSender) ManualRestart(
	ctx context.Context, pri roachpb.UserPriority, ts hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(88321)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.mu.txnState == txnFinalized {
		__antithesis_instrumentation__.Notify(88324)
		log.Fatalf(ctx, "ManualRestart called on finalized txn: %s", tc.mu.txn)
	} else {
		__antithesis_instrumentation__.Notify(88325)
	}
	__antithesis_instrumentation__.Notify(88322)

	tc.mu.txn.Restart(pri, 0, ts)

	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88326)
		reqInt.epochBumpedLocked()
	}
	__antithesis_instrumentation__.Notify(88323)

	tc.mu.txnState = txnPending
}

func (tc *TxnCoordSender) IsSerializablePushAndRefreshNotPossible() bool {
	__antithesis_instrumentation__.Notify(88327)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	isTxnPushed := tc.mu.txn.WriteTimestamp != tc.mu.txn.ReadTimestamp
	refreshAttemptNotPossible := tc.interceptorAlloc.txnSpanRefresher.refreshInvalid || func() bool {
		__antithesis_instrumentation__.Notify(88328)
		return tc.mu.txn.CommitTimestampFixed == true
	}() == true

	return isTxnPushed && func() bool {
		__antithesis_instrumentation__.Notify(88329)
		return refreshAttemptNotPossible == true
	}() == true
}

func (tc *TxnCoordSender) Epoch() enginepb.TxnEpoch {
	__antithesis_instrumentation__.Notify(88330)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.Epoch
}

func (tc *TxnCoordSender) IsLocking() bool {
	__antithesis_instrumentation__.Notify(88331)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.IsLocking()
}

func (tc *TxnCoordSender) IsTracking() bool {
	__antithesis_instrumentation__.Notify(88332)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.interceptorAlloc.txnHeartbeater.heartbeatLoopRunningLocked()
}

func (tc *TxnCoordSender) Active() bool {
	__antithesis_instrumentation__.Notify(88333)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.active
}

func (tc *TxnCoordSender) GetLeafTxnInputState(
	ctx context.Context, opt kv.TxnStatusOpt,
) (*roachpb.LeafTxnInputState, error) {
	__antithesis_instrumentation__.Notify(88334)
	tis := new(roachpb.LeafTxnInputState)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err := tc.checkTxnStatusLocked(ctx, opt); err != nil {
		__antithesis_instrumentation__.Notify(88337)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(88338)
	}
	__antithesis_instrumentation__.Notify(88335)

	tis.Txn = tc.mu.txn
	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88339)
		reqInt.populateLeafInputState(tis)
	}
	__antithesis_instrumentation__.Notify(88336)

	tc.mu.active = true

	return tis, nil
}

func (tc *TxnCoordSender) GetLeafTxnFinalState(
	ctx context.Context, opt kv.TxnStatusOpt,
) (*roachpb.LeafTxnFinalState, error) {
	__antithesis_instrumentation__.Notify(88340)
	tfs := new(roachpb.LeafTxnFinalState)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if err := tc.checkTxnStatusLocked(ctx, opt); err != nil {
		__antithesis_instrumentation__.Notify(88344)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(88345)
	}
	__antithesis_instrumentation__.Notify(88341)

	if tc.mu.active {
		__antithesis_instrumentation__.Notify(88346)
		tfs.DeprecatedCommandCount = 1
	} else {
		__antithesis_instrumentation__.Notify(88347)
	}
	__antithesis_instrumentation__.Notify(88342)

	tfs.Txn = tc.mu.txn
	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88348)
		reqInt.populateLeafFinalState(tfs)
	}
	__antithesis_instrumentation__.Notify(88343)

	return tfs, nil
}

func (tc *TxnCoordSender) checkTxnStatusLocked(ctx context.Context, opt kv.TxnStatusOpt) error {
	__antithesis_instrumentation__.Notify(88349)
	switch opt {
	case kv.AnyTxnStatus:
		__antithesis_instrumentation__.Notify(88351)

	case kv.OnlyPending:
		__antithesis_instrumentation__.Notify(88352)

		rejectErr := tc.maybeRejectClientLocked(ctx, nil)
		if rejectErr != nil {
			__antithesis_instrumentation__.Notify(88354)
			return rejectErr.GoError()
		} else {
			__antithesis_instrumentation__.Notify(88355)
		}
	default:
		__antithesis_instrumentation__.Notify(88353)
		panic("unreachable")
	}
	__antithesis_instrumentation__.Notify(88350)
	return nil
}

func (tc *TxnCoordSender) UpdateRootWithLeafFinalState(
	ctx context.Context, tfs *roachpb.LeafTxnFinalState,
) {
	__antithesis_instrumentation__.Notify(88356)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	if tc.mu.txn.ID == (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(88360)
		log.Fatalf(ctx, "cannot UpdateRootWithLeafFinalState on unbound TxnCoordSender. input id: %s", tfs.Txn.ID)
	} else {
		__antithesis_instrumentation__.Notify(88361)
	}
	__antithesis_instrumentation__.Notify(88357)

	if tc.mu.txn.ID != tfs.Txn.ID {
		__antithesis_instrumentation__.Notify(88362)
		return
	} else {
		__antithesis_instrumentation__.Notify(88363)
	}
	__antithesis_instrumentation__.Notify(88358)

	if tfs.Txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(88364)
		return
	} else {
		__antithesis_instrumentation__.Notify(88365)
	}
	__antithesis_instrumentation__.Notify(88359)

	tc.mu.txn.Update(&tfs.Txn)
	for _, reqInt := range tc.interceptorStack {
		__antithesis_instrumentation__.Notify(88366)
		reqInt.importLeafFinalState(ctx, tfs)
	}
}

func (tc *TxnCoordSender) TestingCloneTxn() *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(88367)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.mu.txn.Clone()
}

func (tc *TxnCoordSender) PrepareRetryableError(ctx context.Context, msg string) error {
	__antithesis_instrumentation__.Notify(88368)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return roachpb.NewTransactionRetryWithProtoRefreshError(
		msg, tc.mu.txn.ID, tc.mu.txn)
}

func (tc *TxnCoordSender) Step(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(88369)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.interceptorAlloc.txnSeqNumAllocator.stepLocked(ctx)
}

func (tc *TxnCoordSender) SetReadSeqNum(seq enginepb.TxnSeq) error {
	__antithesis_instrumentation__.Notify(88370)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if seq < 0 || func() bool {
		__antithesis_instrumentation__.Notify(88372)
		return seq > tc.interceptorAlloc.txnSeqNumAllocator.writeSeq == true
	}() == true {
		__antithesis_instrumentation__.Notify(88373)
		return errors.AssertionFailedf("invalid read seq num < 0 || > writeSeq (%d): %d",
			tc.interceptorAlloc.txnSeqNumAllocator.writeSeq, seq)
	} else {
		__antithesis_instrumentation__.Notify(88374)
	}
	__antithesis_instrumentation__.Notify(88371)
	tc.interceptorAlloc.txnSeqNumAllocator.readSeq = seq
	return nil
}

func (tc *TxnCoordSender) ConfigureStepping(
	ctx context.Context, mode kv.SteppingMode,
) (prevMode kv.SteppingMode) {
	__antithesis_instrumentation__.Notify(88375)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	return tc.interceptorAlloc.txnSeqNumAllocator.configureSteppingLocked(mode)
}

func (tc *TxnCoordSender) GetSteppingMode(ctx context.Context) (curMode kv.SteppingMode) {
	__antithesis_instrumentation__.Notify(88376)
	curMode = kv.SteppingDisabled
	if tc.interceptorAlloc.txnSeqNumAllocator.steppingModeEnabled {
		__antithesis_instrumentation__.Notify(88378)
		curMode = kv.SteppingEnabled
	} else {
		__antithesis_instrumentation__.Notify(88379)
	}
	__antithesis_instrumentation__.Notify(88377)
	return curMode
}

func (tc *TxnCoordSender) ManualRefresh(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(88380)
	tc.mu.Lock()
	defer tc.mu.Unlock()

	var ba roachpb.BatchRequest
	ba.Txn = tc.mu.txn.Clone()
	const force = true
	refreshedBa, pErr := tc.interceptorAlloc.txnSpanRefresher.maybeRefreshPreemptivelyLocked(ctx, ba, force)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88382)
		pErr = tc.updateStateLocked(ctx, ba, nil, pErr)
	} else {
		__antithesis_instrumentation__.Notify(88383)
		var br roachpb.BatchResponse
		br.Txn = refreshedBa.Txn
		pErr = tc.updateStateLocked(ctx, ba, &br, nil)
	}
	__antithesis_instrumentation__.Notify(88381)
	return pErr.GoError()
}

func (tc *TxnCoordSender) DeferCommitWait(ctx context.Context) func(context.Context) error {
	__antithesis_instrumentation__.Notify(88384)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	tc.mu.commitWaitDeferred = true
	return func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(88385)
		tc.mu.Lock()
		defer tc.mu.Unlock()
		if tc.mu.txn.Status != roachpb.COMMITTED {
			__antithesis_instrumentation__.Notify(88387)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(88388)
		}
		__antithesis_instrumentation__.Notify(88386)
		return tc.maybeCommitWait(ctx, true)
	}
}

func (tc *TxnCoordSender) GetTxnRetryableErr(
	ctx context.Context,
) *roachpb.TransactionRetryWithProtoRefreshError {
	__antithesis_instrumentation__.Notify(88389)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.mu.txnState == txnRetryableError {
		__antithesis_instrumentation__.Notify(88391)
		return tc.mu.storedRetryableErr
	} else {
		__antithesis_instrumentation__.Notify(88392)
	}
	__antithesis_instrumentation__.Notify(88390)
	return nil
}

func (tc *TxnCoordSender) ClearTxnRetryableErr(ctx context.Context) {
	__antithesis_instrumentation__.Notify(88393)
	tc.mu.Lock()
	defer tc.mu.Unlock()
	if tc.mu.txnState == txnRetryableError {
		__antithesis_instrumentation__.Notify(88394)
		tc.mu.storedRetryableErr = nil
		tc.mu.txnState = txnPending
	} else {
		__antithesis_instrumentation__.Notify(88395)
	}
}
