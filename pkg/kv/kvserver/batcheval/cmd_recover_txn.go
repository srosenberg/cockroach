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
	RegisterReadWriteCommand(roachpb.RecoverTxn, declareKeysRecoverTransaction, RecoverTxn)
}

func declareKeysRecoverTransaction(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97263)
	rr := req.(*roachpb.RecoverTxnRequest)
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.TransactionKey(rr.Txn.Key, rr.Txn.ID)})
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.AbortSpanKey(rs.GetRangeID(), rr.Txn.ID)})
}

func RecoverTxn(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97264)
	args := cArgs.Args.(*roachpb.RecoverTxnRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.RecoverTxnResponse)

	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97273)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97274)
	}
	__antithesis_instrumentation__.Notify(97265)
	if h.WriteTimestamp().Less(args.Txn.MinTimestamp) {
		__antithesis_instrumentation__.Notify(97275)

		return result.Result{}, errors.AssertionFailedf("RecoverTxn request timestamp %s less than txn MinTimestamp %s",
			h.Timestamp, args.Txn.MinTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97276)
	}
	__antithesis_instrumentation__.Notify(97266)
	if !args.Key.Equal(args.Txn.Key) {
		__antithesis_instrumentation__.Notify(97277)
		return result.Result{}, errors.AssertionFailedf("RecoverTxn request key %s does not match txn key %s",
			args.Key, args.Txn.Key)
	} else {
		__antithesis_instrumentation__.Notify(97278)
	}
	__antithesis_instrumentation__.Notify(97267)
	key := keys.TransactionKey(args.Txn.Key, args.Txn.ID)

	if ok, err := storage.MVCCGetProto(
		ctx, readWriter, key, hlc.Timestamp{}, &reply.RecoveredTxn, storage.MVCCGetOptions{},
	); err != nil {
		__antithesis_instrumentation__.Notify(97279)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97280)
		if !ok {
			__antithesis_instrumentation__.Notify(97281)

			synthTxn := SynthesizeTxnFromMeta(ctx, cArgs.EvalCtx, args.Txn)
			if synthTxn.Status != roachpb.ABORTED {
				__antithesis_instrumentation__.Notify(97283)
				err := errors.Errorf("txn record synthesized with non-ABORTED status: %v", synthTxn)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(97284)
			}
			__antithesis_instrumentation__.Notify(97282)
			reply.RecoveredTxn = synthTxn
			return result.Result{}, nil
		} else {
			__antithesis_instrumentation__.Notify(97285)
		}
	}
	__antithesis_instrumentation__.Notify(97268)

	if args.ImplicitlyCommitted {
		__antithesis_instrumentation__.Notify(97286)

		switch reply.RecoveredTxn.Status {
		case roachpb.PENDING, roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(97287)

			return result.Result{}, errors.AssertionFailedf(
				"programming error: found %s record for implicitly committed transaction: %v",
				reply.RecoveredTxn.Status, reply.RecoveredTxn,
			)
		case roachpb.STAGING, roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(97288)
			if was, is := args.Txn.Epoch, reply.RecoveredTxn.Epoch; was != is {
				__antithesis_instrumentation__.Notify(97292)
				return result.Result{}, errors.AssertionFailedf(
					"programming error: epoch change by implicitly committed transaction: %v->%v", was, is,
				)
			} else {
				__antithesis_instrumentation__.Notify(97293)
			}
			__antithesis_instrumentation__.Notify(97289)
			if was, is := args.Txn.WriteTimestamp, reply.RecoveredTxn.WriteTimestamp; was != is {
				__antithesis_instrumentation__.Notify(97294)
				return result.Result{}, errors.AssertionFailedf(
					"programming error: timestamp change by implicitly committed transaction: %v->%v", was, is,
				)
			} else {
				__antithesis_instrumentation__.Notify(97295)
			}
			__antithesis_instrumentation__.Notify(97290)
			if reply.RecoveredTxn.Status == roachpb.COMMITTED {
				__antithesis_instrumentation__.Notify(97296)

				return result.Result{}, nil
			} else {
				__antithesis_instrumentation__.Notify(97297)
			}

		default:
			__antithesis_instrumentation__.Notify(97291)
			return result.Result{}, errors.AssertionFailedf("bad txn status: %s", reply.RecoveredTxn)
		}
	} else {
		__antithesis_instrumentation__.Notify(97298)

		legalChange := args.Txn.Epoch < reply.RecoveredTxn.Epoch || func() bool {
			__antithesis_instrumentation__.Notify(97299)
			return args.Txn.WriteTimestamp.Less(reply.RecoveredTxn.WriteTimestamp) == true
		}() == true

		switch reply.RecoveredTxn.Status {
		case roachpb.ABORTED:
			__antithesis_instrumentation__.Notify(97300)

			return result.Result{}, nil
		case roachpb.COMMITTED:
			__antithesis_instrumentation__.Notify(97301)

			return result.Result{}, nil
		case roachpb.PENDING:
			__antithesis_instrumentation__.Notify(97302)
			if args.Txn.Epoch < reply.RecoveredTxn.Epoch {
				__antithesis_instrumentation__.Notify(97306)

				return result.Result{}, nil
			} else {
				__antithesis_instrumentation__.Notify(97307)
			}
			__antithesis_instrumentation__.Notify(97303)

			return result.Result{}, errors.AssertionFailedf(
				"programming error: cannot recover PENDING transaction in same epoch: %s", reply.RecoveredTxn,
			)
		case roachpb.STAGING:
			__antithesis_instrumentation__.Notify(97304)
			if legalChange {
				__antithesis_instrumentation__.Notify(97308)

				return result.Result{}, nil
			} else {
				__antithesis_instrumentation__.Notify(97309)
			}

		default:
			__antithesis_instrumentation__.Notify(97305)
			return result.Result{}, errors.AssertionFailedf("bad txn status: %s", reply.RecoveredTxn)
		}
	}
	__antithesis_instrumentation__.Notify(97269)

	for _, w := range reply.RecoveredTxn.InFlightWrites {
		__antithesis_instrumentation__.Notify(97310)
		sp := roachpb.Span{Key: w.Key}
		reply.RecoveredTxn.LockSpans = append(reply.RecoveredTxn.LockSpans, sp)
	}
	__antithesis_instrumentation__.Notify(97270)
	reply.RecoveredTxn.LockSpans, _ = roachpb.MergeSpans(&reply.RecoveredTxn.LockSpans)
	reply.RecoveredTxn.InFlightWrites = nil

	if args.ImplicitlyCommitted {
		__antithesis_instrumentation__.Notify(97311)
		reply.RecoveredTxn.Status = roachpb.COMMITTED
	} else {
		__antithesis_instrumentation__.Notify(97312)
		reply.RecoveredTxn.Status = roachpb.ABORTED
	}
	__antithesis_instrumentation__.Notify(97271)
	txnRecord := reply.RecoveredTxn.AsRecord()
	if err := storage.MVCCPutProto(ctx, readWriter, cArgs.Stats, key, hlc.Timestamp{}, nil, &txnRecord); err != nil {
		__antithesis_instrumentation__.Notify(97313)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97314)
	}
	__antithesis_instrumentation__.Notify(97272)

	result := result.Result{}
	result.Local.UpdatedTxns = []*roachpb.Transaction{&reply.RecoveredTxn}
	return result, nil
}
