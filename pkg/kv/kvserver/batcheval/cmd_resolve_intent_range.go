package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.ResolveIntentRange, declareKeysResolveIntentRange, ResolveIntentRange)
}

func declareKeysResolveIntentRange(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97399)
	declareKeysResolveIntentCombined(rs, req, latchSpans)
}

func ResolveIntentRange(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97400)
	args := cArgs.Args.(*roachpb.ResolveIntentRangeRequest)
	h := cArgs.Header
	ms := cArgs.Stats

	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97405)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97406)
	}
	__antithesis_instrumentation__.Notify(97401)

	update := args.AsLockUpdate()
	numKeys, resumeSpan, err := storage.MVCCResolveWriteIntentRange(
		ctx, readWriter, ms, update, h.MaxSpanRequestKeys)
	if err != nil {
		__antithesis_instrumentation__.Notify(97407)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97408)
	}
	__antithesis_instrumentation__.Notify(97402)
	reply := resp.(*roachpb.ResolveIntentRangeResponse)
	reply.NumKeys = numKeys
	if resumeSpan != nil {
		__antithesis_instrumentation__.Notify(97409)
		update.EndKey = resumeSpan.Key
		reply.ResumeSpan = resumeSpan

		reply.ResumeReason = roachpb.RESUME_KEY_LIMIT
	} else {
		__antithesis_instrumentation__.Notify(97410)
	}
	__antithesis_instrumentation__.Notify(97403)

	var res result.Result
	res.Local.ResolvedLocks = []roachpb.LockUpdate{update}
	res.Local.Metrics = resolveToMetricType(args.Status, args.Poison)

	if WriteAbortSpanOnResolve(args.Status, args.Poison, numKeys > 0) {
		__antithesis_instrumentation__.Notify(97411)
		if err := UpdateAbortSpan(ctx, cArgs.EvalCtx, readWriter, ms, args.IntentTxn, args.Poison); err != nil {
			__antithesis_instrumentation__.Notify(97412)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97413)
		}
	} else {
		__antithesis_instrumentation__.Notify(97414)
	}
	__antithesis_instrumentation__.Notify(97404)
	return res, nil
}
