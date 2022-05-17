package kvclientutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

type PushExpectation int

const (
	ExpectPusheeTxnRecovery PushExpectation = iota

	ExpectPusheeTxnRecordNotFound

	DontExpectAnything
)

type ExpectedTxnResolution int

const (
	ExpectAborted ExpectedTxnResolution = iota

	ExpectCommitted
)

func CheckPushResult(
	ctx context.Context,
	db *kv.DB,
	tr *tracing.Tracer,
	txn roachpb.Transaction,
	expResolution ExpectedTxnResolution,
	pushExpectation PushExpectation,
) error {
	__antithesis_instrumentation__.Notify(644430)
	pushReq := roachpb.PushTxnRequest{
		RequestHeader: roachpb.RequestHeader{
			Key: txn.Key,
		},
		PusheeTxn: txn.TxnMeta,
		PushTo:    hlc.Timestamp{},
		PushType:  roachpb.PUSH_ABORT,

		Force: true,
	}
	ba := roachpb.BatchRequest{}
	ba.Add(&pushReq)

	recCtx, collectRecAndFinish := tracing.ContextWithRecordingSpan(ctx, tr, "test trace")
	defer collectRecAndFinish()

	resp, pErr := db.NonTransactionalSender().Send(recCtx, ba)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(644434)
		return pErr.GoError()
	} else {
		__antithesis_instrumentation__.Notify(644435)
	}
	__antithesis_instrumentation__.Notify(644431)

	var statusErr error
	pusheeStatus := resp.Responses[0].GetPushTxn().PusheeTxn.Status
	switch pusheeStatus {
	case roachpb.ABORTED:
		__antithesis_instrumentation__.Notify(644436)
		if expResolution != ExpectAborted {
			__antithesis_instrumentation__.Notify(644439)
			statusErr = errors.Errorf("transaction unexpectedly aborted")
		} else {
			__antithesis_instrumentation__.Notify(644440)
		}
	case roachpb.COMMITTED:
		__antithesis_instrumentation__.Notify(644437)
		if expResolution != ExpectCommitted {
			__antithesis_instrumentation__.Notify(644441)
			statusErr = errors.Errorf("transaction unexpectedly committed")
		} else {
			__antithesis_instrumentation__.Notify(644442)
		}
	default:
		__antithesis_instrumentation__.Notify(644438)
		return errors.Errorf("unexpected txn status: %s", pusheeStatus)
	}
	__antithesis_instrumentation__.Notify(644432)

	recording := collectRecAndFinish()
	var resolutionErr error
	switch pushExpectation {
	case ExpectPusheeTxnRecovery:
		__antithesis_instrumentation__.Notify(644443)
		expMsg := fmt.Sprintf("recovered txn %s", txn.ID.Short())
		if _, ok := recording.FindLogMessage(expMsg); !ok {
			__antithesis_instrumentation__.Notify(644447)
			resolutionErr = errors.Errorf(
				"recovery didn't run as expected (missing \"%s\"). recording: %s",
				expMsg, recording)
		} else {
			__antithesis_instrumentation__.Notify(644448)
		}
	case ExpectPusheeTxnRecordNotFound:
		__antithesis_instrumentation__.Notify(644444)
		expMsg := "pushee txn record not found"
		if _, ok := recording.FindLogMessage(expMsg); !ok {
			__antithesis_instrumentation__.Notify(644449)
			resolutionErr = errors.Errorf(
				"push didn't run as expected (missing \"%s\"). recording: %s",
				expMsg, recording)
		} else {
			__antithesis_instrumentation__.Notify(644450)
		}
	case DontExpectAnything:
		__antithesis_instrumentation__.Notify(644445)
	default:
		__antithesis_instrumentation__.Notify(644446)
	}
	__antithesis_instrumentation__.Notify(644433)

	return errors.CombineErrors(statusErr, resolutionErr)
}
