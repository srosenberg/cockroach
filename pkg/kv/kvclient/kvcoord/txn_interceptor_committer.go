package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/multitenant"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
)

var parallelCommitsEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.transaction.parallel_commits_enabled",
	"if enabled, transactional commits will be parallelized with transactional writes",
	true,
)

type txnCommitter struct {
	st      *cluster.Settings
	stopper *stop.Stopper
	wrapped lockedSender
	mu      sync.Locker
}

func (tc *txnCommitter) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88476)

	rArgs, hasET := ba.GetArg(roachpb.EndTxn)
	if !hasET {
		__antithesis_instrumentation__.Notify(88486)
		return tc.wrapped.SendLocked(ctx, ba)
	} else {
		__antithesis_instrumentation__.Notify(88487)
	}
	__antithesis_instrumentation__.Notify(88477)
	et := rArgs.(*roachpb.EndTxnRequest)

	if err := tc.validateEndTxnBatch(ba); err != nil {
		__antithesis_instrumentation__.Notify(88488)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(88489)
	}
	__antithesis_instrumentation__.Notify(88478)

	if len(et.LockSpans) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(88490)
		return len(et.InFlightWrites) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(88491)
		return tc.sendLockedWithElidedEndTxn(ctx, ba, et)
	} else {
		__antithesis_instrumentation__.Notify(88492)
	}
	__antithesis_instrumentation__.Notify(88479)

	var etAttempt endTxnAttempt
	if et.Key == nil {
		__antithesis_instrumentation__.Notify(88493)
		et.Key = ba.Txn.Key
		etAttempt = endTxnFirstAttempt
	} else {
		__antithesis_instrumentation__.Notify(88494)

		etAttempt = endTxnRetry
		if len(et.InFlightWrites) > 0 {
			__antithesis_instrumentation__.Notify(88495)

			etCpy := *et
			ba.Requests[len(ba.Requests)-1].MustSetInner(&etCpy)
			et = &etCpy
		} else {
			__antithesis_instrumentation__.Notify(88496)
		}
	}
	__antithesis_instrumentation__.Notify(88480)

	if len(et.InFlightWrites) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(88497)
		return !tc.canCommitInParallel(ctx, ba, et, etAttempt) == true
	}() == true {
		__antithesis_instrumentation__.Notify(88498)

		et.LockSpans, ba.Header.DistinctSpans = mergeIntoSpans(et.LockSpans, et.InFlightWrites)

		et.InFlightWrites = nil
	} else {
		__antithesis_instrumentation__.Notify(88499)
	}
	__antithesis_instrumentation__.Notify(88481)

	if !et.Commit {
		__antithesis_instrumentation__.Notify(88500)
		return tc.wrapped.SendLocked(ctx, ba)
	} else {
		__antithesis_instrumentation__.Notify(88501)
	}
	__antithesis_instrumentation__.Notify(88482)

	br, pErr := tc.wrapped.SendLocked(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88502)

		if txn := pErr.GetTxn(); txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(88504)
			return txn.Status == roachpb.STAGING == true
		}() == true {
			__antithesis_instrumentation__.Notify(88505)
			pErr.SetTxn(cloneWithStatus(txn, roachpb.PENDING))
		} else {
			__antithesis_instrumentation__.Notify(88506)
		}
		__antithesis_instrumentation__.Notify(88503)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88507)
	}
	__antithesis_instrumentation__.Notify(88483)

	switch br.Txn.Status {
	case roachpb.STAGING:
		__antithesis_instrumentation__.Notify(88508)

	case roachpb.COMMITTED:
		__antithesis_instrumentation__.Notify(88509)

		return br, nil
	default:
		__antithesis_instrumentation__.Notify(88510)
		return nil, roachpb.NewErrorf("unexpected response status without error: %v", br.Txn)
	}
	__antithesis_instrumentation__.Notify(88484)

	if pErr := needTxnRetryAfterStaging(br); pErr != nil {
		__antithesis_instrumentation__.Notify(88511)
		log.VEventf(ctx, 2, "parallel commit failed since some writes were pushed. "+
			"Synthesized err: %s", pErr)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(88512)
	}
	__antithesis_instrumentation__.Notify(88485)

	mergedLockSpans, _ := mergeIntoSpans(et.LockSpans, et.InFlightWrites)
	tc.makeTxnCommitExplicitAsync(ctx, br.Txn, mergedLockSpans)

	br.Txn = cloneWithStatus(br.Txn, roachpb.COMMITTED)
	return br, nil
}

func (tc *txnCommitter) validateEndTxnBatch(ba roachpb.BatchRequest) error {
	__antithesis_instrumentation__.Notify(88513)

	if ba.Header.MaxSpanRequestKeys == 0 {
		__antithesis_instrumentation__.Notify(88516)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(88517)
	}
	__antithesis_instrumentation__.Notify(88514)
	e, endTxn := ba.GetArg(roachpb.EndTxn)
	_, delRange := ba.GetArg(roachpb.DeleteRange)
	if delRange && func() bool {
		__antithesis_instrumentation__.Notify(88518)
		return endTxn == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(88519)
		return !e.(*roachpb.EndTxnRequest).Require1PC == true
	}() == true {
		__antithesis_instrumentation__.Notify(88520)
		return errors.Errorf("possible 1PC batch cannot contain EndTxn without setting Require1PC; see #37457")
	} else {
		__antithesis_instrumentation__.Notify(88521)
	}
	__antithesis_instrumentation__.Notify(88515)
	return nil
}

func (tc *txnCommitter) sendLockedWithElidedEndTxn(
	ctx context.Context, ba roachpb.BatchRequest, et *roachpb.EndTxnRequest,
) (br *roachpb.BatchResponse, pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(88522)

	ba.Requests = ba.Requests[:len(ba.Requests)-1]
	if len(ba.Requests) > 0 {
		__antithesis_instrumentation__.Notify(88526)
		br, pErr = tc.wrapped.SendLocked(ctx, ba)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(88527)
			return nil, pErr
		} else {
			__antithesis_instrumentation__.Notify(88528)
		}
	} else {
		__antithesis_instrumentation__.Notify(88529)
		br = &roachpb.BatchResponse{}

		br.Txn = ba.Txn
	}
	__antithesis_instrumentation__.Notify(88523)

	if et.Commit {
		__antithesis_instrumentation__.Notify(88530)
		deadline := et.Deadline
		if deadline != nil && func() bool {
			__antithesis_instrumentation__.Notify(88531)
			return !deadline.IsEmpty() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(88532)
			return deadline.LessEq(br.Txn.WriteTimestamp) == true
		}() == true {
			__antithesis_instrumentation__.Notify(88533)
			return nil, generateTxnDeadlineExceededErr(ba.Txn, *deadline)
		} else {
			__antithesis_instrumentation__.Notify(88534)
		}
	} else {
		__antithesis_instrumentation__.Notify(88535)
	}
	__antithesis_instrumentation__.Notify(88524)

	status := roachpb.ABORTED
	if et.Commit {
		__antithesis_instrumentation__.Notify(88536)
		status = roachpb.COMMITTED
	} else {
		__antithesis_instrumentation__.Notify(88537)
	}
	__antithesis_instrumentation__.Notify(88525)
	br.Txn = cloneWithStatus(br.Txn, status)

	br.Add(&roachpb.EndTxnResponse{})
	return br, nil
}

type endTxnAttempt int

const (
	endTxnFirstAttempt endTxnAttempt = iota
	endTxnRetry
)

func (tc *txnCommitter) canCommitInParallel(
	ctx context.Context, ba roachpb.BatchRequest, et *roachpb.EndTxnRequest, etAttempt endTxnAttempt,
) bool {
	__antithesis_instrumentation__.Notify(88538)
	if !parallelCommitsEnabled.Get(&tc.st.SV) {
		__antithesis_instrumentation__.Notify(88544)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88545)
	}
	__antithesis_instrumentation__.Notify(88539)

	if etAttempt == endTxnRetry {
		__antithesis_instrumentation__.Notify(88546)
		log.VEventf(ctx, 2, "retrying batch not eligible for parallel commit")
		return false
	} else {
		__antithesis_instrumentation__.Notify(88547)
	}
	__antithesis_instrumentation__.Notify(88540)

	if !et.Commit {
		__antithesis_instrumentation__.Notify(88548)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88549)
	}
	__antithesis_instrumentation__.Notify(88541)

	if et.InternalCommitTrigger != nil {
		__antithesis_instrumentation__.Notify(88550)
		return false
	} else {
		__antithesis_instrumentation__.Notify(88551)
	}
	__antithesis_instrumentation__.Notify(88542)

	for _, ru := range ba.Requests[:len(ba.Requests)-1] {
		__antithesis_instrumentation__.Notify(88552)
		req := ru.GetInner()
		switch {
		case roachpb.IsIntentWrite(req):
			__antithesis_instrumentation__.Notify(88553)
			if roachpb.IsRange(req) {
				__antithesis_instrumentation__.Notify(88556)

				return false
			} else {
				__antithesis_instrumentation__.Notify(88557)
			}

		case req.Method() == roachpb.QueryIntent:
			__antithesis_instrumentation__.Notify(88554)

		default:
			__antithesis_instrumentation__.Notify(88555)

			return false
		}
	}
	__antithesis_instrumentation__.Notify(88543)
	return true
}

func mergeIntoSpans(s []roachpb.Span, ws []roachpb.SequencedWrite) ([]roachpb.Span, bool) {
	__antithesis_instrumentation__.Notify(88558)
	m := make([]roachpb.Span, len(s)+len(ws))
	copy(m, s)
	for i, w := range ws {
		__antithesis_instrumentation__.Notify(88560)
		m[len(s)+i] = roachpb.Span{Key: w.Key}
	}
	__antithesis_instrumentation__.Notify(88559)
	return roachpb.MergeSpans(&m)
}

func needTxnRetryAfterStaging(br *roachpb.BatchResponse) *roachpb.Error {
	__antithesis_instrumentation__.Notify(88561)
	if len(br.Responses) == 0 {
		__antithesis_instrumentation__.Notify(88566)
		return roachpb.NewErrorf("no responses in BatchResponse: %v", br)
	} else {
		__antithesis_instrumentation__.Notify(88567)
	}
	__antithesis_instrumentation__.Notify(88562)
	lastResp := br.Responses[len(br.Responses)-1].GetInner()
	etResp, ok := lastResp.(*roachpb.EndTxnResponse)
	if !ok {
		__antithesis_instrumentation__.Notify(88568)
		return roachpb.NewErrorf("unexpected response in BatchResponse: %v", lastResp)
	} else {
		__antithesis_instrumentation__.Notify(88569)
	}
	__antithesis_instrumentation__.Notify(88563)
	if etResp.StagingTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(88570)
		return roachpb.NewErrorf("empty StagingTimestamp in EndTxnResponse: %v", etResp)
	} else {
		__antithesis_instrumentation__.Notify(88571)
	}
	__antithesis_instrumentation__.Notify(88564)
	if etResp.StagingTimestamp.Less(br.Txn.WriteTimestamp) {
		__antithesis_instrumentation__.Notify(88572)

		reason := roachpb.RETRY_SERIALIZABLE
		if br.Txn.WriteTooOld {
			__antithesis_instrumentation__.Notify(88574)
			reason = roachpb.RETRY_WRITE_TOO_OLD
		} else {
			__antithesis_instrumentation__.Notify(88575)
		}
		__antithesis_instrumentation__.Notify(88573)
		err := roachpb.NewTransactionRetryError(
			reason, "serializability failure concurrent with STAGING")
		txn := cloneWithStatus(br.Txn, roachpb.PENDING)
		return roachpb.NewErrorWithTxn(err, txn)
	} else {
		__antithesis_instrumentation__.Notify(88576)
	}
	__antithesis_instrumentation__.Notify(88565)
	return nil
}

func (tc *txnCommitter) makeTxnCommitExplicitAsync(
	ctx context.Context, txn *roachpb.Transaction, lockSpans []roachpb.Span,
) {
	__antithesis_instrumentation__.Notify(88577)

	log.VEventf(ctx, 2, "making txn commit explicit: %s", txn)
	asyncCtx := context.Background()

	if multitenant.HasTenantCostControlExemption(ctx) {
		__antithesis_instrumentation__.Notify(88579)
		asyncCtx = multitenant.WithTenantCostControlExemption(asyncCtx)
	} else {
		__antithesis_instrumentation__.Notify(88580)
	}
	__antithesis_instrumentation__.Notify(88578)
	if err := tc.stopper.RunAsyncTask(
		asyncCtx, "txnCommitter: making txn commit explicit", func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(88581)
			tc.mu.Lock()
			defer tc.mu.Unlock()
			if err := makeTxnCommitExplicitLocked(ctx, tc.wrapped, txn, lockSpans); err != nil {
				__antithesis_instrumentation__.Notify(88582)
				log.Errorf(ctx, "making txn commit explicit failed for %s: %v", txn, err)
			} else {
				__antithesis_instrumentation__.Notify(88583)
			}
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(88584)
		log.VErrEventf(ctx, 1, "failed to make txn commit explicit: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(88585)
	}
}

func makeTxnCommitExplicitLocked(
	ctx context.Context, s lockedSender, txn *roachpb.Transaction, lockSpans []roachpb.Span,
) error {
	__antithesis_instrumentation__.Notify(88586)

	txn = txn.Clone()

	ba := roachpb.BatchRequest{}
	ba.Header = roachpb.Header{Txn: txn}
	et := roachpb.EndTxnRequest{Commit: true}
	et.Key = txn.Key
	et.LockSpans = lockSpans
	ba.Add(&et)

	_, pErr := s.SendLocked(ctx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(88588)
		switch t := pErr.GetDetail().(type) {
		case *roachpb.TransactionStatusError:
			__antithesis_instrumentation__.Notify(88590)

			if t.Reason == roachpb.TransactionStatusError_REASON_TXN_COMMITTED {
				__antithesis_instrumentation__.Notify(88593)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(88594)
			}
		case *roachpb.TransactionRetryError:
			__antithesis_instrumentation__.Notify(88591)
			logFunc := log.Errorf
			if util.RaceEnabled {
				__antithesis_instrumentation__.Notify(88595)
				logFunc = log.Fatalf
			} else {
				__antithesis_instrumentation__.Notify(88596)
			}
			__antithesis_instrumentation__.Notify(88592)
			logFunc(ctx, "unexpected retry error when making commit explicit for %s: %v", txn, t)
		}
		__antithesis_instrumentation__.Notify(88589)
		return pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(88597)
	}
	__antithesis_instrumentation__.Notify(88587)
	return nil
}

func (tc *txnCommitter) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(88598)
	tc.wrapped = wrapped
}

func (*txnCommitter) populateLeafInputState(*roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(88599)
}

func (*txnCommitter) populateLeafFinalState(*roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88600)
}

func (*txnCommitter) importLeafFinalState(context.Context, *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(88601)
}

func (tc *txnCommitter) epochBumpedLocked() { __antithesis_instrumentation__.Notify(88602) }

func (*txnCommitter) createSavepointLocked(context.Context, *savepoint) {
	__antithesis_instrumentation__.Notify(88603)
}

func (*txnCommitter) rollbackToSavepointLocked(context.Context, savepoint) {
	__antithesis_instrumentation__.Notify(88604)
}

func (tc *txnCommitter) closeLocked() { __antithesis_instrumentation__.Notify(88605) }

func cloneWithStatus(txn *roachpb.Transaction, s roachpb.TransactionStatus) *roachpb.Transaction {
	__antithesis_instrumentation__.Notify(88606)
	clone := txn.Clone()
	clone.Status = s
	return clone
}
