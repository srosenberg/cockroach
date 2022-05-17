package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

func init() {
	RegisterReadWriteCommand(roachpb.ResolveIntent, declareKeysResolveIntent, ResolveIntent)
}

func declareKeysResolveIntentCombined(
	rs ImmutableRangeState, req roachpb.Request, latchSpans *spanset.SpanSet,
) {
	__antithesis_instrumentation__.Notify(97374)
	var status roachpb.TransactionStatus
	var txnID uuid.UUID
	var minTxnTS hlc.Timestamp
	switch t := req.(type) {
	case *roachpb.ResolveIntentRequest:
		__antithesis_instrumentation__.Notify(97376)
		status = t.Status
		txnID = t.IntentTxn.ID
		minTxnTS = t.IntentTxn.MinTimestamp
	case *roachpb.ResolveIntentRangeRequest:
		__antithesis_instrumentation__.Notify(97377)
		status = t.Status
		txnID = t.IntentTxn.ID
		minTxnTS = t.IntentTxn.MinTimestamp
	}
	__antithesis_instrumentation__.Notify(97375)
	latchSpans.AddMVCC(spanset.SpanReadWrite, req.Header().Span(), minTxnTS)
	if status == roachpb.ABORTED {
		__antithesis_instrumentation__.Notify(97378)

		latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.AbortSpanKey(rs.GetRangeID(), txnID)})
	} else {
		__antithesis_instrumentation__.Notify(97379)
	}
}

func declareKeysResolveIntent(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97380)
	declareKeysResolveIntentCombined(rs, req, latchSpans)
}

func resolveToMetricType(status roachpb.TransactionStatus, poison bool) *result.Metrics {
	__antithesis_instrumentation__.Notify(97381)
	var typ result.Metrics
	if status == roachpb.ABORTED {
		__antithesis_instrumentation__.Notify(97383)
		if poison {
			__antithesis_instrumentation__.Notify(97384)
			typ.ResolvePoison = 1
		} else {
			__antithesis_instrumentation__.Notify(97385)
			typ.ResolveAbort = 1
		}
	} else {
		__antithesis_instrumentation__.Notify(97386)
		typ.ResolveCommit = 1
	}
	__antithesis_instrumentation__.Notify(97382)
	return &typ
}

func ResolveIntent(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97387)
	args := cArgs.Args.(*roachpb.ResolveIntentRequest)
	h := cArgs.Header
	ms := cArgs.Stats

	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97391)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97392)
	}
	__antithesis_instrumentation__.Notify(97388)

	update := args.AsLockUpdate()
	ok, err := storage.MVCCResolveWriteIntent(ctx, readWriter, ms, update)
	if err != nil {
		__antithesis_instrumentation__.Notify(97393)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97394)
	}
	__antithesis_instrumentation__.Notify(97389)

	var res result.Result
	res.Local.ResolvedLocks = []roachpb.LockUpdate{update}
	res.Local.Metrics = resolveToMetricType(args.Status, args.Poison)

	if WriteAbortSpanOnResolve(args.Status, args.Poison, ok) {
		__antithesis_instrumentation__.Notify(97395)
		if err := UpdateAbortSpan(ctx, cArgs.EvalCtx, readWriter, ms, args.IntentTxn, args.Poison); err != nil {
			__antithesis_instrumentation__.Notify(97396)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97397)
		}
	} else {
		__antithesis_instrumentation__.Notify(97398)
	}
	__antithesis_instrumentation__.Notify(97390)
	return res, nil
}
