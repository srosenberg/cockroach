package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(roachpb.QueryIntent, declareKeysQueryIntent, QueryIntent)
}

func declareKeysQueryIntent(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97153)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, req.Header().Span())
}

func QueryIntent(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97154)
	args := cArgs.Args.(*roachpb.QueryIntentRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.QueryIntentResponse)

	ownTxn := false
	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(97160)

		if h.Txn.ID == args.Txn.ID {
			__antithesis_instrumentation__.Notify(97161)
			ownTxn = true
		} else {
			__antithesis_instrumentation__.Notify(97162)
			return result.Result{}, ErrTransactionUnsupported
		}
	} else {
		__antithesis_instrumentation__.Notify(97163)
	}
	__antithesis_instrumentation__.Notify(97155)
	if h.WriteTimestamp().Less(args.Txn.WriteTimestamp) {
		__antithesis_instrumentation__.Notify(97164)

		return result.Result{}, errors.AssertionFailedf("QueryIntent request timestamp %s less than txn WriteTimestamp %s",
			h.Timestamp, args.Txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97165)
	}
	__antithesis_instrumentation__.Notify(97156)

	_, intent, err := storage.MVCCGet(ctx, reader, args.Key, hlc.MaxTimestamp, storage.MVCCGetOptions{

		Inconsistent: true,

		Txn: nil,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(97166)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97167)
	}
	__antithesis_instrumentation__.Notify(97157)

	var curIntentPushed bool
	if intent != nil {
		__antithesis_instrumentation__.Notify(97168)

		reply.FoundIntent = (args.Txn.ID == intent.Txn.ID) && func() bool {
			__antithesis_instrumentation__.Notify(97169)
			return (args.Txn.Epoch == intent.Txn.Epoch) == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(97170)
			return (args.Txn.Sequence <= intent.Txn.Sequence) == true
		}() == true

		if reply.FoundIntent {
			__antithesis_instrumentation__.Notify(97171)

			cmpTS := args.Txn.WriteTimestamp
			if ownTxn {
				__antithesis_instrumentation__.Notify(97173)
				cmpTS.Forward(h.Txn.WriteTimestamp)
			} else {
				__antithesis_instrumentation__.Notify(97174)
			}
			__antithesis_instrumentation__.Notify(97172)
			if cmpTS.Less(intent.Txn.WriteTimestamp) {
				__antithesis_instrumentation__.Notify(97175)

				curIntentPushed = true
				log.VEventf(ctx, 2, "found pushed intent")
				reply.FoundIntent = false

				if ownTxn {
					__antithesis_instrumentation__.Notify(97176)
					reply.Txn = h.Txn.Clone()
					reply.Txn.WriteTimestamp.Forward(intent.Txn.WriteTimestamp)
				} else {
					__antithesis_instrumentation__.Notify(97177)
				}
			} else {
				__antithesis_instrumentation__.Notify(97178)
			}
		} else {
			__antithesis_instrumentation__.Notify(97179)
		}
	} else {
		__antithesis_instrumentation__.Notify(97180)
	}
	__antithesis_instrumentation__.Notify(97158)

	if !reply.FoundIntent && func() bool {
		__antithesis_instrumentation__.Notify(97181)
		return args.ErrorIfMissing == true
	}() == true {
		__antithesis_instrumentation__.Notify(97182)
		if ownTxn && func() bool {
			__antithesis_instrumentation__.Notify(97184)
			return curIntentPushed == true
		}() == true {
			__antithesis_instrumentation__.Notify(97185)

			return result.Result{}, roachpb.NewTransactionRetryError(roachpb.RETRY_SERIALIZABLE, "intent pushed")
		} else {
			__antithesis_instrumentation__.Notify(97186)
		}
		__antithesis_instrumentation__.Notify(97183)
		return result.Result{}, roachpb.NewIntentMissingError(args.Key, intent)
	} else {
		__antithesis_instrumentation__.Notify(97187)
	}
	__antithesis_instrumentation__.Notify(97159)
	return result.Result{}, nil
}
