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
)

func init() {
	RegisterReadWriteCommand(roachpb.ConditionalPut, declareKeysConditionalPut, ConditionalPut)
}

func declareKeysConditionalPut(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(96481)
	args := req.(*roachpb.ConditionalPutRequest)
	if args.Inline {
		__antithesis_instrumentation__.Notify(96482)
		DefaultDeclareKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	} else {
		__antithesis_instrumentation__.Notify(96483)
		DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	}
}

func ConditionalPut(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96484)
	args := cArgs.Args.(*roachpb.ConditionalPutRequest)
	h := cArgs.Header

	var ts hlc.Timestamp
	if !args.Inline {
		__antithesis_instrumentation__.Notify(96488)
		ts = h.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(96489)
	}
	__antithesis_instrumentation__.Notify(96485)

	var expVal []byte
	if len(args.ExpBytes) != 0 {
		__antithesis_instrumentation__.Notify(96490)
		expVal = args.ExpBytes
	} else {
		__antithesis_instrumentation__.Notify(96491)

		if args.DeprecatedExpValue != nil {
			__antithesis_instrumentation__.Notify(96492)
			expVal = args.DeprecatedExpValue.TagAndDataBytes()
		} else {
			__antithesis_instrumentation__.Notify(96493)
		}
	}
	__antithesis_instrumentation__.Notify(96486)

	handleMissing := storage.CPutMissingBehavior(args.AllowIfDoesNotExist)
	var err error
	if args.Blind {
		__antithesis_instrumentation__.Notify(96494)
		err = storage.MVCCBlindConditionalPut(ctx, readWriter, cArgs.Stats, args.Key, ts, args.Value, expVal, handleMissing, h.Txn)
	} else {
		__antithesis_instrumentation__.Notify(96495)
		err = storage.MVCCConditionalPut(ctx, readWriter, cArgs.Stats, args.Key, ts, args.Value, expVal, handleMissing, h.Txn)
	}
	__antithesis_instrumentation__.Notify(96487)

	return result.FromAcquiredLocks(h.Txn, args.Key), err
}
