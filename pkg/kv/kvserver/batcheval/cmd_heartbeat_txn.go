package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func init() {
	RegisterReadWriteCommand(roachpb.HeartbeatTxn, declareKeysHeartbeatTransaction, HeartbeatTxn)
}

func declareKeysHeartbeatTransaction(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(96942)
	declareKeysWriteTransaction(rs, header, req, latchSpans)
}

func HeartbeatTxn(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96943)
	args := cArgs.Args.(*roachpb.HeartbeatTxnRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.HeartbeatTxnResponse)

	if err := VerifyTransaction(h, args, roachpb.PENDING, roachpb.STAGING); err != nil {
		__antithesis_instrumentation__.Notify(96948)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96949)
	}
	__antithesis_instrumentation__.Notify(96944)

	if args.Now.IsEmpty() {
		__antithesis_instrumentation__.Notify(96950)
		return result.Result{}, fmt.Errorf("now not specified for heartbeat")
	} else {
		__antithesis_instrumentation__.Notify(96951)
	}
	__antithesis_instrumentation__.Notify(96945)

	key := keys.TransactionKey(h.Txn.Key, h.Txn.ID)

	var txn roachpb.Transaction
	if ok, err := storage.MVCCGetProto(
		ctx, readWriter, key, hlc.Timestamp{}, &txn, storage.MVCCGetOptions{},
	); err != nil {
		__antithesis_instrumentation__.Notify(96952)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96953)
		if !ok {
			__antithesis_instrumentation__.Notify(96954)

			txn = *h.Txn

			if err := CanCreateTxnRecord(ctx, cArgs.EvalCtx, &txn); err != nil {
				__antithesis_instrumentation__.Notify(96955)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96956)
			}
		} else {
			__antithesis_instrumentation__.Notify(96957)
		}
	}
	__antithesis_instrumentation__.Notify(96946)

	if !txn.Status.IsFinalized() {
		__antithesis_instrumentation__.Notify(96958)

		txn.LastHeartbeat.Forward(args.Now)
		txnRecord := txn.AsRecord()
		if err := storage.MVCCPutProto(ctx, readWriter, cArgs.Stats, key, hlc.Timestamp{}, nil, &txnRecord); err != nil {
			__antithesis_instrumentation__.Notify(96959)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96960)
		}
	} else {
		__antithesis_instrumentation__.Notify(96961)
	}
	__antithesis_instrumentation__.Notify(96947)

	reply.Txn = &txn
	return result.Result{}, nil
}
