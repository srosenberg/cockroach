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
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(roachpb.QueryTxn, declareKeysQueryTransaction, QueryTxn)
}

func declareKeysQueryTransaction(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97227)
	qr := req.(*roachpb.QueryTxnRequest)
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.TransactionKey(qr.Txn.Key, qr.Txn.ID)})
}

func QueryTxn(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97228)
	args := cArgs.Args.(*roachpb.QueryTxnRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.QueryTxnResponse)

	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97234)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97235)
	}
	__antithesis_instrumentation__.Notify(97229)
	if h.WriteTimestamp().Less(args.Txn.MinTimestamp) {
		__antithesis_instrumentation__.Notify(97236)

		return result.Result{}, errors.AssertionFailedf("QueryTxn request timestamp %s less than txn MinTimestamp %s",
			h.Timestamp, args.Txn.MinTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97237)
	}
	__antithesis_instrumentation__.Notify(97230)
	if !args.Key.Equal(args.Txn.Key) {
		__antithesis_instrumentation__.Notify(97238)
		return result.Result{}, errors.AssertionFailedf("QueryTxn request key %s does not match txn key %s",
			args.Key, args.Txn.Key)
	} else {
		__antithesis_instrumentation__.Notify(97239)
	}
	__antithesis_instrumentation__.Notify(97231)
	key := keys.TransactionKey(args.Txn.Key, args.Txn.ID)

	ok, err := storage.MVCCGetProto(
		ctx, reader, key, hlc.Timestamp{}, &reply.QueriedTxn, storage.MVCCGetOptions{},
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(97240)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97241)
	}
	__antithesis_instrumentation__.Notify(97232)
	if ok {
		__antithesis_instrumentation__.Notify(97242)
		reply.TxnRecordExists = true
	} else {
		__antithesis_instrumentation__.Notify(97243)

		reply.QueriedTxn = SynthesizeTxnFromMeta(ctx, cArgs.EvalCtx, args.Txn)
	}
	__antithesis_instrumentation__.Notify(97233)

	reply.WaitingTxns = cArgs.EvalCtx.GetConcurrencyManager().GetDependents(args.Txn.ID)
	return result.Result{}, nil
}
