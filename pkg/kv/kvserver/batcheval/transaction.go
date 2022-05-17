package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

var ErrTransactionUnsupported = errors.AssertionFailedf("not supported within a transaction")

func VerifyTransaction(
	h roachpb.Header, args roachpb.Request, permittedStatuses ...roachpb.TransactionStatus,
) error {
	__antithesis_instrumentation__.Notify(97876)
	if h.Txn == nil {
		__antithesis_instrumentation__.Notify(97881)
		return errors.AssertionFailedf("no transaction specified to %s", args.Method())
	} else {
		__antithesis_instrumentation__.Notify(97882)
	}
	__antithesis_instrumentation__.Notify(97877)
	if !bytes.Equal(args.Header().Key, h.Txn.Key) {
		__antithesis_instrumentation__.Notify(97883)
		return errors.AssertionFailedf("request key %s should match txn key %s", args.Header().Key, h.Txn.Key)
	} else {
		__antithesis_instrumentation__.Notify(97884)
	}
	__antithesis_instrumentation__.Notify(97878)
	statusPermitted := false
	for _, s := range permittedStatuses {
		__antithesis_instrumentation__.Notify(97885)
		if h.Txn.Status == s {
			__antithesis_instrumentation__.Notify(97886)
			statusPermitted = true
			break
		} else {
			__antithesis_instrumentation__.Notify(97887)
		}
	}
	__antithesis_instrumentation__.Notify(97879)
	if !statusPermitted {
		__antithesis_instrumentation__.Notify(97888)
		reason := roachpb.TransactionStatusError_REASON_UNKNOWN
		if h.Txn.Status == roachpb.COMMITTED {
			__antithesis_instrumentation__.Notify(97890)
			reason = roachpb.TransactionStatusError_REASON_TXN_COMMITTED
		} else {
			__antithesis_instrumentation__.Notify(97891)
		}
		__antithesis_instrumentation__.Notify(97889)
		return roachpb.NewTransactionStatusError(reason,
			fmt.Sprintf("cannot perform %s with txn status %v", args.Method(), h.Txn.Status))
	} else {
		__antithesis_instrumentation__.Notify(97892)
	}
	__antithesis_instrumentation__.Notify(97880)
	return nil
}

func WriteAbortSpanOnResolve(status roachpb.TransactionStatus, poison, removedIntents bool) bool {
	__antithesis_instrumentation__.Notify(97893)
	if status != roachpb.ABORTED {
		__antithesis_instrumentation__.Notify(97896)

		return false
	} else {
		__antithesis_instrumentation__.Notify(97897)
	}
	__antithesis_instrumentation__.Notify(97894)
	if !poison {
		__antithesis_instrumentation__.Notify(97898)

		return true
	} else {
		__antithesis_instrumentation__.Notify(97899)
	}
	__antithesis_instrumentation__.Notify(97895)

	return removedIntents
}

func UpdateAbortSpan(
	ctx context.Context,
	rec EvalContext,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	txn enginepb.TxnMeta,
	poison bool,
) error {
	__antithesis_instrumentation__.Notify(97900)

	var curEntry roachpb.AbortSpanEntry
	exists, err := rec.AbortSpan().Get(ctx, readWriter, txn.ID, &curEntry)
	if err != nil {
		__antithesis_instrumentation__.Notify(97904)
		return err
	} else {
		__antithesis_instrumentation__.Notify(97905)
	}
	__antithesis_instrumentation__.Notify(97901)

	if !poison {
		__antithesis_instrumentation__.Notify(97906)
		if !exists {
			__antithesis_instrumentation__.Notify(97908)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(97909)
		}
		__antithesis_instrumentation__.Notify(97907)
		return rec.AbortSpan().Del(ctx, readWriter, ms, txn.ID)
	} else {
		__antithesis_instrumentation__.Notify(97910)
	}
	__antithesis_instrumentation__.Notify(97902)

	entry := roachpb.AbortSpanEntry{
		Key:       txn.Key,
		Timestamp: txn.WriteTimestamp,
		Priority:  txn.Priority,
	}
	if exists && func() bool {
		__antithesis_instrumentation__.Notify(97911)
		return curEntry.Equal(entry) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97912)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(97913)
	}
	__antithesis_instrumentation__.Notify(97903)

	curEntry = entry
	return rec.AbortSpan().Put(ctx, readWriter, ms, txn.ID, &curEntry)
}

func CanPushWithPriority(pusher, pushee *roachpb.Transaction) bool {
	__antithesis_instrumentation__.Notify(97914)
	return (pusher.Priority > enginepb.MinTxnPriority && func() bool {
		__antithesis_instrumentation__.Notify(97915)
		return pushee.Priority == enginepb.MinTxnPriority == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(97916)
		return (pusher.Priority == enginepb.MaxTxnPriority && func() bool {
			__antithesis_instrumentation__.Notify(97917)
			return pushee.Priority < pusher.Priority == true
		}() == true) == true
	}() == true
}

func CanCreateTxnRecord(ctx context.Context, rec EvalContext, txn *roachpb.Transaction) error {
	__antithesis_instrumentation__.Notify(97918)

	ok, minCommitTS, reason := rec.CanCreateTxnRecord(ctx, txn.ID, txn.Key, txn.MinTimestamp)
	if !ok {
		__antithesis_instrumentation__.Notify(97921)
		log.VEventf(ctx, 2, "txn tombstone present; transaction has been aborted")
		return roachpb.NewTransactionAbortedError(reason)
	} else {
		__antithesis_instrumentation__.Notify(97922)
	}
	__antithesis_instrumentation__.Notify(97919)
	if bumped := txn.WriteTimestamp.Forward(minCommitTS); bumped {
		__antithesis_instrumentation__.Notify(97923)
		log.VEventf(ctx, 2, "write timestamp bumped by txn tombstone to: %s", txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(97924)
	}
	__antithesis_instrumentation__.Notify(97920)
	return nil
}

func SynthesizeTxnFromMeta(
	ctx context.Context, rec EvalContext, txn enginepb.TxnMeta,
) roachpb.Transaction {
	__antithesis_instrumentation__.Notify(97925)
	synth := roachpb.TransactionRecord{
		TxnMeta: txn,
		Status:  roachpb.PENDING,

		LastHeartbeat: txn.MinTimestamp,
	}

	if txn.MinTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(97928)
		synth.Status = roachpb.ABORTED
		return synth.AsTransaction()
	} else {
		__antithesis_instrumentation__.Notify(97929)
	}
	__antithesis_instrumentation__.Notify(97926)

	ok, minCommitTS, _ := rec.CanCreateTxnRecord(ctx, txn.ID, txn.Key, txn.MinTimestamp)
	if ok {
		__antithesis_instrumentation__.Notify(97930)

		synth.WriteTimestamp.Forward(minCommitTS)
	} else {
		__antithesis_instrumentation__.Notify(97931)

		synth.Status = roachpb.ABORTED
	}
	__antithesis_instrumentation__.Notify(97927)
	return synth.AsTransaction()
}
