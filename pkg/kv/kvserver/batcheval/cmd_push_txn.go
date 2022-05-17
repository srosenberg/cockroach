package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/txnwait"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func init() {
	RegisterReadWriteCommand(roachpb.PushTxn, declareKeysPushTransaction, PushTxn)
}

func declareKeysPushTransaction(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97069)
	pr := req.(*roachpb.PushTxnRequest)
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.TransactionKey(pr.PusheeTxn.Key, pr.PusheeTxn.ID)})
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.AbortSpanKey(rs.GetRangeID(), pr.PusheeTxn.ID)})
}

func PushTxn(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97070)
	args := cArgs.Args.(*roachpb.PushTxnRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.PushTxnResponse)

	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97087)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97088)
	}
	__antithesis_instrumentation__.Notify(97071)
	if h.WriteTimestamp().Less(args.PushTo) {
		__antithesis_instrumentation__.Notify(97089)

		return result.Result{}, errors.AssertionFailedf("request timestamp %s less than PushTo timestamp %s",
			h.Timestamp, args.PushTo)
	} else {
		__antithesis_instrumentation__.Notify(97090)
	}
	__antithesis_instrumentation__.Notify(97072)
	if h.WriteTimestamp().Less(args.PusheeTxn.MinTimestamp) {
		__antithesis_instrumentation__.Notify(97091)

		return result.Result{}, errors.AssertionFailedf("request timestamp %s less than pushee txn MinTimestamp %s",
			h.Timestamp, args.PusheeTxn.MinTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97092)
	}
	__antithesis_instrumentation__.Notify(97073)
	if !bytes.Equal(args.Key, args.PusheeTxn.Key) {
		__antithesis_instrumentation__.Notify(97093)
		return result.Result{}, errors.AssertionFailedf("request key %s should match pushee txn key %s",
			args.Key, args.PusheeTxn.Key)
	} else {
		__antithesis_instrumentation__.Notify(97094)
	}
	__antithesis_instrumentation__.Notify(97074)
	key := keys.TransactionKey(args.PusheeTxn.Key, args.PusheeTxn.ID)

	var existTxn roachpb.Transaction
	ok, err := storage.MVCCGetProto(ctx, readWriter, key, hlc.Timestamp{}, &existTxn, storage.MVCCGetOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(97095)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97096)
		if !ok {
			__antithesis_instrumentation__.Notify(97097)
			log.VEventf(ctx, 2, "pushee txn record not found (pushee: %s)", args.PusheeTxn.Short())

			reply.PusheeTxn = SynthesizeTxnFromMeta(ctx, cArgs.EvalCtx, args.PusheeTxn)
			if reply.PusheeTxn.Status == roachpb.ABORTED {
				__antithesis_instrumentation__.Notify(97098)

				result := result.Result{}
				result.Local.UpdatedTxns = []*roachpb.Transaction{&reply.PusheeTxn}
				return result, nil
			} else {
				__antithesis_instrumentation__.Notify(97099)
			}
		} else {
			__antithesis_instrumentation__.Notify(97100)

			reply.PusheeTxn = existTxn
		}
	}
	__antithesis_instrumentation__.Notify(97075)

	if reply.PusheeTxn.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(97101)

		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(97102)
	}
	__antithesis_instrumentation__.Notify(97076)

	pushType := args.PushType
	if pushType == roachpb.PUSH_TIMESTAMP && func() bool {
		__antithesis_instrumentation__.Notify(97103)
		return args.PushTo.LessEq(reply.PusheeTxn.WriteTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97104)

		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(97105)
	}
	__antithesis_instrumentation__.Notify(97077)

	var knownHigherTimestamp, knownHigherEpoch bool
	if reply.PusheeTxn.WriteTimestamp.Less(args.PusheeTxn.WriteTimestamp) {
		__antithesis_instrumentation__.Notify(97106)
		reply.PusheeTxn.WriteTimestamp = args.PusheeTxn.WriteTimestamp
		knownHigherTimestamp = true
	} else {
		__antithesis_instrumentation__.Notify(97107)
	}
	__antithesis_instrumentation__.Notify(97078)
	if reply.PusheeTxn.Epoch < args.PusheeTxn.Epoch {
		__antithesis_instrumentation__.Notify(97108)
		reply.PusheeTxn.Epoch = args.PusheeTxn.Epoch
		knownHigherEpoch = true
	} else {
		__antithesis_instrumentation__.Notify(97109)
	}
	__antithesis_instrumentation__.Notify(97079)
	reply.PusheeTxn.UpgradePriority(args.PusheeTxn.Priority)

	if (knownHigherTimestamp || func() bool {
		__antithesis_instrumentation__.Notify(97110)
		return knownHigherEpoch == true
	}() == true) && func() bool {
		__antithesis_instrumentation__.Notify(97111)
		return reply.PusheeTxn.Status == roachpb.STAGING == true
	}() == true {
		__antithesis_instrumentation__.Notify(97112)
		reply.PusheeTxn.Status = roachpb.PENDING
		reply.PusheeTxn.InFlightWrites = nil

		if !knownHigherEpoch && func() bool {
			__antithesis_instrumentation__.Notify(97113)
			return pushType == roachpb.PUSH_TIMESTAMP == true
		}() == true {
			__antithesis_instrumentation__.Notify(97114)
			pushType = roachpb.PUSH_ABORT
		} else {
			__antithesis_instrumentation__.Notify(97115)
		}
	} else {
		__antithesis_instrumentation__.Notify(97116)
	}
	__antithesis_instrumentation__.Notify(97080)

	var pusherWins bool
	var reason string

	switch {
	case txnwait.IsExpired(cArgs.EvalCtx.Clock().Now(), &reply.PusheeTxn):
		__antithesis_instrumentation__.Notify(97117)
		reason = "pushee is expired"

		pushType = roachpb.PUSH_ABORT
		pusherWins = true
	case pushType == roachpb.PUSH_TOUCH:
		__antithesis_instrumentation__.Notify(97118)

		pusherWins = false
	case CanPushWithPriority(&args.PusherTxn, &reply.PusheeTxn):
		__antithesis_instrumentation__.Notify(97119)
		reason = "pusher has priority"
		pusherWins = true
	case args.Force:
		__antithesis_instrumentation__.Notify(97120)
		reason = "forced push"
		pusherWins = true
	default:
		__antithesis_instrumentation__.Notify(97121)
	}
	__antithesis_instrumentation__.Notify(97081)

	if log.V(1) && func() bool {
		__antithesis_instrumentation__.Notify(97122)
		return reason != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(97123)
		s := "pushed"
		if !pusherWins {
			__antithesis_instrumentation__.Notify(97125)
			s = "failed to push"
		} else {
			__antithesis_instrumentation__.Notify(97126)
		}
		__antithesis_instrumentation__.Notify(97124)
		log.Infof(ctx, "%s %s (push type=%s) %s: %s (pushee last active: %s)",
			args.PusherTxn.Short(), redact.Safe(s),
			redact.Safe(pushType),
			args.PusheeTxn.Short(),
			redact.Safe(reason),
			reply.PusheeTxn.LastActive())
	} else {
		__antithesis_instrumentation__.Notify(97127)
	}
	__antithesis_instrumentation__.Notify(97082)

	recoverOnFailedPush := cArgs.EvalCtx.EvalKnobs().RecoverIndeterminateCommitsOnFailedPushes
	if reply.PusheeTxn.Status == roachpb.STAGING && func() bool {
		__antithesis_instrumentation__.Notify(97128)
		return (pusherWins || func() bool {
			__antithesis_instrumentation__.Notify(97129)
			return recoverOnFailedPush == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97130)
		err := roachpb.NewIndeterminateCommitError(reply.PusheeTxn)
		log.VEventf(ctx, 1, "%v", err)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97131)
	}
	__antithesis_instrumentation__.Notify(97083)

	if !pusherWins {
		__antithesis_instrumentation__.Notify(97132)
		err := roachpb.NewTransactionPushError(reply.PusheeTxn)
		log.VEventf(ctx, 1, "%v", err)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97133)
	}
	__antithesis_instrumentation__.Notify(97084)

	reply.PusheeTxn.UpgradePriority(args.PusherTxn.Priority - 1)

	switch pushType {
	case roachpb.PUSH_ABORT:
		__antithesis_instrumentation__.Notify(97134)

		reply.PusheeTxn.Status = roachpb.ABORTED

		if ok {
			__antithesis_instrumentation__.Notify(97137)
			reply.PusheeTxn.WriteTimestamp.Forward(reply.PusheeTxn.LastActive())
		} else {
			__antithesis_instrumentation__.Notify(97138)
		}
	case roachpb.PUSH_TIMESTAMP:
		__antithesis_instrumentation__.Notify(97135)

		reply.PusheeTxn.WriteTimestamp.Forward(args.PushTo)
	default:
		__antithesis_instrumentation__.Notify(97136)
		return result.Result{}, errors.AssertionFailedf("unexpected push type: %v", pushType)
	}
	__antithesis_instrumentation__.Notify(97085)

	if ok {
		__antithesis_instrumentation__.Notify(97139)
		txnRecord := reply.PusheeTxn.AsRecord()
		if err := storage.MVCCPutProto(ctx, readWriter, cArgs.Stats, key, hlc.Timestamp{}, nil, &txnRecord); err != nil {
			__antithesis_instrumentation__.Notify(97140)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(97141)
		}
	} else {
		__antithesis_instrumentation__.Notify(97142)
	}
	__antithesis_instrumentation__.Notify(97086)

	result := result.Result{}
	result.Local.UpdatedTxns = []*roachpb.Transaction{&reply.PusheeTxn}
	return result, nil
}
