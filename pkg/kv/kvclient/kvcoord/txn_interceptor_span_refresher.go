package kvcoord

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/metric"
	"github.com/cockroachdb/errors"
)

const (
	maxTxnRefreshAttempts = 5
)

var MaxTxnRefreshSpansBytes = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.transaction.max_refresh_spans_bytes",
	"maximum number of bytes used to track refresh spans in serializable transactions",
	256*1000,
).WithPublic()

type txnSpanRefresher struct {
	st      *cluster.Settings
	knobs   *ClientTestingKnobs
	riGen   rangeIteratorFactory
	wrapped lockedSender

	refreshFootprint condensableSpanSet

	refreshInvalid bool

	refreshedTimestamp hlc.Timestamp

	canAutoRetry bool

	refreshSuccess                *metric.Counter
	refreshFail                   *metric.Counter
	refreshFailWithCondensedSpans *metric.Counter
	refreshMemoryLimitExceeded    *metric.Counter
	refreshAutoRetries            *metric.Counter
}

func (sr *txnSpanRefresher) SendLocked(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89066)
	batchReadTimestamp := ba.Txn.ReadTimestamp
	if sr.refreshedTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(89072)

		sr.refreshedTimestamp = batchReadTimestamp
	} else {
		__antithesis_instrumentation__.Notify(89073)
		if batchReadTimestamp.Less(sr.refreshedTimestamp) {
			__antithesis_instrumentation__.Notify(89074)

			ba.Txn.ReadTimestamp.Forward(sr.refreshedTimestamp)
			ba.Txn.WriteTimestamp.Forward(sr.refreshedTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(89075)
			if sr.refreshedTimestamp != batchReadTimestamp {
				__antithesis_instrumentation__.Notify(89076)
				return nil, roachpb.NewError(errors.AssertionFailedf(
					"unexpected batch read timestamp: %s. Expected refreshed timestamp: %s. ba: %s. txn: %s",
					batchReadTimestamp, sr.refreshedTimestamp, ba, ba.Txn))
			} else {
				__antithesis_instrumentation__.Notify(89077)
			}
		}
	}
	__antithesis_instrumentation__.Notify(89067)

	ba.CanForwardReadTimestamp = sr.canForwardReadTimestampWithoutRefresh(ba.Txn)

	ba, pErr := sr.maybeRefreshPreemptivelyLocked(ctx, ba, false)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89078)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89079)
	}
	__antithesis_instrumentation__.Notify(89068)

	br, pErr := sr.sendLockedWithRefreshAttempts(ctx, ba, sr.maxRefreshAttempts())
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89080)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89081)
	}
	__antithesis_instrumentation__.Notify(89069)

	if br.Txn.Status != roachpb.PENDING {
		__antithesis_instrumentation__.Notify(89082)
		return br, nil
	} else {
		__antithesis_instrumentation__.Notify(89083)
	}
	__antithesis_instrumentation__.Notify(89070)

	if !sr.refreshInvalid {
		__antithesis_instrumentation__.Notify(89084)
		if err := sr.appendRefreshSpans(ctx, ba, br); err != nil {
			__antithesis_instrumentation__.Notify(89086)
			return nil, roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(89087)
		}
		__antithesis_instrumentation__.Notify(89085)

		maxBytes := MaxTxnRefreshSpansBytes.Get(&sr.st.SV)
		if sr.refreshFootprint.bytes >= maxBytes {
			__antithesis_instrumentation__.Notify(89088)
			condensedBefore := sr.refreshFootprint.condensed
			condensedSufficient := sr.tryCondenseRefreshSpans(ctx, maxBytes)
			if condensedSufficient {
				__antithesis_instrumentation__.Notify(89090)
				log.VEventf(ctx, 2, "condensed refresh spans for txn %s to %d bytes",
					br.Txn, sr.refreshFootprint.bytes)
			} else {
				__antithesis_instrumentation__.Notify(89091)

				log.VEventf(ctx, 2, "condensed refresh spans didn't save enough memory. txn %s. "+
					"refresh spans after condense: %d bytes",
					br.Txn, sr.refreshFootprint.bytes)
				sr.refreshInvalid = true
				sr.refreshFootprint.clear()
			}
			__antithesis_instrumentation__.Notify(89089)

			if sr.refreshFootprint.condensed && func() bool {
				__antithesis_instrumentation__.Notify(89092)
				return !condensedBefore == true
			}() == true {
				__antithesis_instrumentation__.Notify(89093)
				sr.refreshMemoryLimitExceeded.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(89094)
			}
		} else {
			__antithesis_instrumentation__.Notify(89095)
		}
	} else {
		__antithesis_instrumentation__.Notify(89096)
	}
	__antithesis_instrumentation__.Notify(89071)
	return br, nil
}

func (sr *txnSpanRefresher) tryCondenseRefreshSpans(ctx context.Context, maxBytes int64) bool {
	__antithesis_instrumentation__.Notify(89097)
	if sr.knobs.CondenseRefreshSpansFilter != nil && func() bool {
		__antithesis_instrumentation__.Notify(89099)
		return !sr.knobs.CondenseRefreshSpansFilter() == true
	}() == true {
		__antithesis_instrumentation__.Notify(89100)
		return false
	} else {
		__antithesis_instrumentation__.Notify(89101)
	}
	__antithesis_instrumentation__.Notify(89098)
	sr.refreshFootprint.maybeCondense(ctx, sr.riGen, maxBytes)
	return sr.refreshFootprint.bytes < maxBytes
}

func (sr *txnSpanRefresher) sendLockedWithRefreshAttempts(
	ctx context.Context, ba roachpb.BatchRequest, maxRefreshAttempts int,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89102)
	if ba.Txn.WriteTooOld {
		__antithesis_instrumentation__.Notify(89107)

		log.Fatalf(ctx, "unexpected WriteTooOld request. ba: %s (txn: %s)",
			ba.String(), ba.Txn.String())
	} else {
		__antithesis_instrumentation__.Notify(89108)
	}
	__antithesis_instrumentation__.Notify(89103)
	br, pErr := sr.wrapped.SendLocked(ctx, ba)

	if pErr != nil && func() bool {
		__antithesis_instrumentation__.Notify(89109)
		return pErr.GetTxn() != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(89110)
		pErr.GetTxn().WriteTooOld = false
	} else {
		__antithesis_instrumentation__.Notify(89111)
	}
	__antithesis_instrumentation__.Notify(89104)

	if pErr == nil && func() bool {
		__antithesis_instrumentation__.Notify(89112)
		return br.Txn.WriteTooOld == true
	}() == true {
		__antithesis_instrumentation__.Notify(89113)

		bumpedTxn := br.Txn.Clone()
		bumpedTxn.WriteTooOld = false
		pErr = roachpb.NewErrorWithTxn(
			roachpb.NewTransactionRetryError(roachpb.RETRY_WRITE_TOO_OLD,
				"WriteTooOld flag converted to WriteTooOldError"),
			bumpedTxn)
		br = nil
	} else {
		__antithesis_instrumentation__.Notify(89114)
	}
	__antithesis_instrumentation__.Notify(89105)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89115)
		if maxRefreshAttempts > 0 {
			__antithesis_instrumentation__.Notify(89116)
			br, pErr = sr.maybeRefreshAndRetrySend(ctx, ba, pErr, maxRefreshAttempts)
		} else {
			__antithesis_instrumentation__.Notify(89117)
			log.VEventf(ctx, 2, "not checking error for refresh; refresh attempts exhausted")
		}
	} else {
		__antithesis_instrumentation__.Notify(89118)
	}
	__antithesis_instrumentation__.Notify(89106)
	sr.forwardRefreshTimestampOnResponse(br, pErr)
	return br, pErr
}

func (sr *txnSpanRefresher) maybeRefreshAndRetrySend(
	ctx context.Context, ba roachpb.BatchRequest, pErr *roachpb.Error, maxRefreshAttempts int,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89119)
	txn := pErr.GetTxn()
	if txn == nil || func() bool {
		__antithesis_instrumentation__.Notify(89125)
		return !sr.canForwardReadTimestamp(txn) == true
	}() == true {
		__antithesis_instrumentation__.Notify(89126)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89127)
	}
	__antithesis_instrumentation__.Notify(89120)

	ok, refreshTS := roachpb.TransactionRefreshTimestamp(pErr)
	if !ok {
		__antithesis_instrumentation__.Notify(89128)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89129)
	}
	__antithesis_instrumentation__.Notify(89121)
	refreshTxn := txn.Clone()
	refreshTxn.Refresh(refreshTS)
	log.VEventf(ctx, 2, "trying to refresh to %s because of %s", refreshTxn.ReadTimestamp, pErr)

	if refreshErr := sr.tryRefreshTxnSpans(ctx, refreshTxn); refreshErr != nil {
		__antithesis_instrumentation__.Notify(89130)
		log.Eventf(ctx, "refresh failed; propagating original retry error")

		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89131)
	}
	__antithesis_instrumentation__.Notify(89122)

	log.Eventf(ctx, "refresh succeeded; retrying original request")
	ba.UpdateTxn(refreshTxn)
	sr.refreshAutoRetries.Inc(1)

	args, hasET := ba.GetArg(roachpb.EndTxn)
	if len(ba.Requests) > 1 && func() bool {
		__antithesis_instrumentation__.Notify(89132)
		return hasET == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(89133)
		return !args.(*roachpb.EndTxnRequest).Require1PC == true
	}() == true {
		__antithesis_instrumentation__.Notify(89134)
		log.Eventf(ctx, "sending EndTxn separately from rest of batch on retry")
		return sr.splitEndTxnAndRetrySend(ctx, ba)
	} else {
		__antithesis_instrumentation__.Notify(89135)
	}
	__antithesis_instrumentation__.Notify(89123)

	retryBr, retryErr := sr.sendLockedWithRefreshAttempts(ctx, ba, maxRefreshAttempts-1)
	if retryErr != nil {
		__antithesis_instrumentation__.Notify(89136)
		log.VEventf(ctx, 2, "retry failed with %s", retryErr)
		return nil, retryErr
	} else {
		__antithesis_instrumentation__.Notify(89137)
	}
	__antithesis_instrumentation__.Notify(89124)

	log.VEventf(ctx, 2, "retry successful @%s", retryBr.Txn.ReadTimestamp)
	return retryBr, nil
}

func (sr *txnSpanRefresher) splitEndTxnAndRetrySend(
	ctx context.Context, ba roachpb.BatchRequest,
) (*roachpb.BatchResponse, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89138)

	etIdx := len(ba.Requests) - 1
	baPrefix := ba
	baPrefix.Requests = ba.Requests[:etIdx]
	brPrefix, pErr := sr.SendLocked(ctx, baPrefix)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89142)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89143)
	}
	__antithesis_instrumentation__.Notify(89139)

	baSuffix := ba
	baSuffix.Requests = ba.Requests[etIdx:]
	baSuffix.UpdateTxn(brPrefix.Txn)
	brSuffix, pErr := sr.SendLocked(ctx, baSuffix)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89144)
		return nil, pErr
	} else {
		__antithesis_instrumentation__.Notify(89145)
	}
	__antithesis_instrumentation__.Notify(89140)

	br := brPrefix
	br.Responses = append(br.Responses, roachpb.ResponseUnion{})
	if err := br.Combine(brSuffix, []int{etIdx}); err != nil {
		__antithesis_instrumentation__.Notify(89146)
		return nil, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(89147)
	}
	__antithesis_instrumentation__.Notify(89141)
	return br, nil
}

func (sr *txnSpanRefresher) maybeRefreshPreemptivelyLocked(
	ctx context.Context, ba roachpb.BatchRequest, force bool,
) (roachpb.BatchRequest, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89148)

	if ba.Txn.ReadTimestamp == ba.Txn.WriteTimestamp {
		__antithesis_instrumentation__.Notify(89153)
		return ba, nil
	} else {
		__antithesis_instrumentation__.Notify(89154)
	}
	__antithesis_instrumentation__.Notify(89149)

	refreshFree := ba.CanForwardReadTimestamp

	args, hasET := ba.GetArg(roachpb.EndTxn)
	refreshInevitable := hasET && func() bool {
		__antithesis_instrumentation__.Notify(89155)
		return args.(*roachpb.EndTxnRequest).Commit == true
	}() == true

	if !refreshFree && func() bool {
		__antithesis_instrumentation__.Notify(89156)
		return !refreshInevitable == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(89157)
		return !force == true
	}() == true {
		__antithesis_instrumentation__.Notify(89158)
		return ba, nil
	} else {
		__antithesis_instrumentation__.Notify(89159)
	}
	__antithesis_instrumentation__.Notify(89150)

	if !sr.canForwardReadTimestamp(ba.Txn) {
		__antithesis_instrumentation__.Notify(89160)
		return ba, newRetryErrorOnFailedPreemptiveRefresh(ba.Txn, nil)
	} else {
		__antithesis_instrumentation__.Notify(89161)
	}
	__antithesis_instrumentation__.Notify(89151)

	refreshTxn := ba.Txn.Clone()
	refreshTxn.Refresh(ba.Txn.WriteTimestamp)
	log.VEventf(ctx, 2, "preemptively refreshing to timestamp %s before issuing %s",
		refreshTxn.ReadTimestamp, ba)

	if refreshErr := sr.tryRefreshTxnSpans(ctx, refreshTxn); refreshErr != nil {
		__antithesis_instrumentation__.Notify(89162)
		log.Eventf(ctx, "preemptive refresh failed; propagating retry error")
		return roachpb.BatchRequest{}, newRetryErrorOnFailedPreemptiveRefresh(ba.Txn, refreshErr)
	} else {
		__antithesis_instrumentation__.Notify(89163)
	}
	__antithesis_instrumentation__.Notify(89152)

	log.Eventf(ctx, "preemptive refresh succeeded")
	ba.UpdateTxn(refreshTxn)
	return ba, nil
}

func newRetryErrorOnFailedPreemptiveRefresh(
	txn *roachpb.Transaction, refreshErr *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(89164)
	reason := roachpb.RETRY_SERIALIZABLE
	if txn.WriteTooOld {
		__antithesis_instrumentation__.Notify(89167)
		reason = roachpb.RETRY_WRITE_TOO_OLD
	} else {
		__antithesis_instrumentation__.Notify(89168)
	}
	__antithesis_instrumentation__.Notify(89165)
	msg := "failed preemptive refresh"
	if refreshErr != nil {
		__antithesis_instrumentation__.Notify(89169)
		if refreshErr, ok := refreshErr.GetDetail().(*roachpb.RefreshFailedError); ok {
			__antithesis_instrumentation__.Notify(89170)
			msg = fmt.Sprintf("%s due to a conflict: %s on key %s", msg, refreshErr.FailureReason(), refreshErr.Key)
		} else {
			__antithesis_instrumentation__.Notify(89171)
			msg = fmt.Sprintf("%s - unknown error: %s", msg, refreshErr)
		}
	} else {
		__antithesis_instrumentation__.Notify(89172)
	}
	__antithesis_instrumentation__.Notify(89166)
	retryErr := roachpb.NewTransactionRetryError(reason, msg)
	return roachpb.NewErrorWithTxn(retryErr, txn)
}

func (sr *txnSpanRefresher) tryRefreshTxnSpans(
	ctx context.Context, refreshTxn *roachpb.Transaction,
) (err *roachpb.Error) {
	__antithesis_instrumentation__.Notify(89173)

	defer func() {
		__antithesis_instrumentation__.Notify(89178)
		if err == nil {
			__antithesis_instrumentation__.Notify(89179)
			sr.refreshSuccess.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(89180)
			sr.refreshFail.Inc(1)
			if sr.refreshFootprint.condensed {
				__antithesis_instrumentation__.Notify(89181)
				sr.refreshFailWithCondensedSpans.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(89182)
			}
		}
	}()
	__antithesis_instrumentation__.Notify(89174)

	if sr.refreshInvalid {
		__antithesis_instrumentation__.Notify(89183)
		log.VEvent(ctx, 2, "can't refresh txn spans; not valid")
		return roachpb.NewError(errors.AssertionFailedf("can't refresh txn spans; not valid"))
	} else {
		__antithesis_instrumentation__.Notify(89184)
		if sr.refreshFootprint.empty() {
			__antithesis_instrumentation__.Notify(89185)
			log.VEvent(ctx, 2, "there are no txn spans to refresh")
			sr.refreshedTimestamp.Forward(refreshTxn.ReadTimestamp)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(89186)
		}
	}
	__antithesis_instrumentation__.Notify(89175)

	refreshSpanBa := roachpb.BatchRequest{}
	refreshSpanBa.Txn = refreshTxn
	addRefreshes := func(refreshes *condensableSpanSet) {
		__antithesis_instrumentation__.Notify(89187)

		for _, u := range refreshes.asSlice() {
			__antithesis_instrumentation__.Notify(89188)
			var req roachpb.Request
			if len(u.EndKey) == 0 {
				__antithesis_instrumentation__.Notify(89190)
				req = &roachpb.RefreshRequest{
					RequestHeader: roachpb.RequestHeaderFromSpan(u),
					RefreshFrom:   sr.refreshedTimestamp,
				}
			} else {
				__antithesis_instrumentation__.Notify(89191)
				req = &roachpb.RefreshRangeRequest{
					RequestHeader: roachpb.RequestHeaderFromSpan(u),
					RefreshFrom:   sr.refreshedTimestamp,
				}
			}
			__antithesis_instrumentation__.Notify(89189)
			refreshSpanBa.Add(req)
			log.VEventf(ctx, 2, "updating span %s @%s - @%s to avoid serializable restart",
				req.Header().Span(), sr.refreshedTimestamp, refreshTxn.WriteTimestamp)
		}
	}
	__antithesis_instrumentation__.Notify(89176)
	addRefreshes(&sr.refreshFootprint)

	if _, batchErr := sr.wrapped.SendLocked(ctx, refreshSpanBa); batchErr != nil {
		__antithesis_instrumentation__.Notify(89192)
		log.VEventf(ctx, 2, "failed to refresh txn spans (%s)", batchErr)
		return batchErr
	} else {
		__antithesis_instrumentation__.Notify(89193)
	}
	__antithesis_instrumentation__.Notify(89177)

	sr.refreshedTimestamp.Forward(refreshTxn.ReadTimestamp)
	return nil
}

func (sr *txnSpanRefresher) appendRefreshSpans(
	ctx context.Context, ba roachpb.BatchRequest, br *roachpb.BatchResponse,
) error {
	__antithesis_instrumentation__.Notify(89194)
	readTimestamp := br.Txn.ReadTimestamp
	if readTimestamp.Less(sr.refreshedTimestamp) {
		__antithesis_instrumentation__.Notify(89197)

		return errors.AssertionFailedf("attempting to append refresh spans after the tracked"+
			" timestamp has moved forward. batchTimestamp: %s refreshedTimestamp: %s ba: %s",
			errors.Safe(readTimestamp), errors.Safe(sr.refreshedTimestamp), ba)
	} else {
		__antithesis_instrumentation__.Notify(89198)
	}
	__antithesis_instrumentation__.Notify(89195)

	ba.RefreshSpanIterate(br, func(span roachpb.Span) {
		__antithesis_instrumentation__.Notify(89199)
		if log.ExpensiveLogEnabled(ctx, 3) {
			__antithesis_instrumentation__.Notify(89201)
			log.VEventf(ctx, 3, "recording span to refresh: %s", span.String())
		} else {
			__antithesis_instrumentation__.Notify(89202)
		}
		__antithesis_instrumentation__.Notify(89200)
		sr.refreshFootprint.insert(span)
	})
	__antithesis_instrumentation__.Notify(89196)
	return nil
}

func (sr *txnSpanRefresher) canForwardReadTimestamp(txn *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(89203)
	return sr.canAutoRetry && func() bool {
		__antithesis_instrumentation__.Notify(89204)
		return !txn.CommitTimestampFixed == true
	}() == true
}

func (sr *txnSpanRefresher) canForwardReadTimestampWithoutRefresh(txn *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(89205)
	return sr.canForwardReadTimestamp(txn) && func() bool {
		__antithesis_instrumentation__.Notify(89206)
		return !sr.refreshInvalid == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(89207)
		return sr.refreshFootprint.empty() == true
	}() == true
}

func (sr *txnSpanRefresher) forwardRefreshTimestampOnResponse(
	br *roachpb.BatchResponse, pErr *roachpb.Error,
) {
	__antithesis_instrumentation__.Notify(89208)
	var txn *roachpb.Transaction
	if pErr != nil {
		__antithesis_instrumentation__.Notify(89210)
		txn = pErr.GetTxn()
	} else {
		__antithesis_instrumentation__.Notify(89211)
		txn = br.Txn
	}
	__antithesis_instrumentation__.Notify(89209)
	if txn != nil {
		__antithesis_instrumentation__.Notify(89212)
		sr.refreshedTimestamp.Forward(txn.ReadTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(89213)
	}
}

func (sr *txnSpanRefresher) maxRefreshAttempts() int {
	__antithesis_instrumentation__.Notify(89214)
	if knob := sr.knobs.MaxTxnRefreshAttempts; knob != 0 {
		__antithesis_instrumentation__.Notify(89216)
		if knob == -1 {
			__antithesis_instrumentation__.Notify(89218)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(89219)
		}
		__antithesis_instrumentation__.Notify(89217)
		return knob
	} else {
		__antithesis_instrumentation__.Notify(89220)
	}
	__antithesis_instrumentation__.Notify(89215)
	return maxTxnRefreshAttempts
}

func (sr *txnSpanRefresher) setWrapped(wrapped lockedSender) {
	__antithesis_instrumentation__.Notify(89221)
	sr.wrapped = wrapped
}

func (sr *txnSpanRefresher) populateLeafInputState(tis *roachpb.LeafTxnInputState) {
	__antithesis_instrumentation__.Notify(89222)
	tis.RefreshInvalid = sr.refreshInvalid
}

func (sr *txnSpanRefresher) populateLeafFinalState(tfs *roachpb.LeafTxnFinalState) {
	__antithesis_instrumentation__.Notify(89223)
	tfs.RefreshInvalid = sr.refreshInvalid
	if !sr.refreshInvalid {
		__antithesis_instrumentation__.Notify(89224)

		tfs.RefreshSpans = append([]roachpb.Span(nil), sr.refreshFootprint.asSlice()...)
	} else {
		__antithesis_instrumentation__.Notify(89225)
	}
}

func (sr *txnSpanRefresher) importLeafFinalState(
	ctx context.Context, tfs *roachpb.LeafTxnFinalState,
) {
	__antithesis_instrumentation__.Notify(89226)
	if tfs.RefreshInvalid {
		__antithesis_instrumentation__.Notify(89227)
		sr.refreshInvalid = true
		sr.refreshFootprint.clear()
	} else {
		__antithesis_instrumentation__.Notify(89228)
		if !sr.refreshInvalid {
			__antithesis_instrumentation__.Notify(89229)
			sr.refreshFootprint.insert(tfs.RefreshSpans...)
			sr.refreshFootprint.maybeCondense(ctx, sr.riGen, MaxTxnRefreshSpansBytes.Get(&sr.st.SV))
		} else {
			__antithesis_instrumentation__.Notify(89230)
		}
	}
}

func (sr *txnSpanRefresher) epochBumpedLocked() {
	__antithesis_instrumentation__.Notify(89231)
	sr.refreshFootprint.clear()
	sr.refreshInvalid = false
	sr.refreshedTimestamp.Reset()
}

func (sr *txnSpanRefresher) createSavepointLocked(ctx context.Context, s *savepoint) {
	__antithesis_instrumentation__.Notify(89232)
	s.refreshSpans = make([]roachpb.Span, len(sr.refreshFootprint.asSlice()))
	copy(s.refreshSpans, sr.refreshFootprint.asSlice())
	s.refreshInvalid = sr.refreshInvalid
}

func (sr *txnSpanRefresher) rollbackToSavepointLocked(ctx context.Context, s savepoint) {
	__antithesis_instrumentation__.Notify(89233)
	sr.refreshFootprint.clear()
	sr.refreshFootprint.insert(s.refreshSpans...)
	sr.refreshInvalid = s.refreshInvalid
}

func (*txnSpanRefresher) closeLocked() { __antithesis_instrumentation__.Notify(89234) }
