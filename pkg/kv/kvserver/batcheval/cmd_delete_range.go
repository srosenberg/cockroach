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
	RegisterReadWriteCommand(roachpb.DeleteRange, declareKeysDeleteRange, DeleteRange)
}

func declareKeysDeleteRange(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(96497)
	args := req.(*roachpb.DeleteRangeRequest)
	if args.Inline {
		__antithesis_instrumentation__.Notify(96498)
		DefaultDeclareKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	} else {
		__antithesis_instrumentation__.Notify(96499)
		DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	}
}

func DeleteRange(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96500)
	args := cArgs.Args.(*roachpb.DeleteRangeRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.DeleteRangeResponse)

	var timestamp hlc.Timestamp
	if !args.Inline {
		__antithesis_instrumentation__.Notify(96504)
		timestamp = h.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(96505)
	}
	__antithesis_instrumentation__.Notify(96501)

	returnKeys := args.ReturnKeys || func() bool {
		__antithesis_instrumentation__.Notify(96506)
		return h.Txn != nil == true
	}() == true
	deleted, resumeSpan, num, err := storage.MVCCDeleteRange(
		ctx, readWriter, cArgs.Stats, args.Key, args.EndKey, h.MaxSpanRequestKeys, timestamp, h.Txn, returnKeys,
	)
	if err == nil && func() bool {
		__antithesis_instrumentation__.Notify(96507)
		return args.ReturnKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(96508)
		reply.Keys = deleted
	} else {
		__antithesis_instrumentation__.Notify(96509)
	}
	__antithesis_instrumentation__.Notify(96502)
	reply.NumKeys = num
	if resumeSpan != nil {
		__antithesis_instrumentation__.Notify(96510)
		reply.ResumeSpan = resumeSpan
		reply.ResumeReason = roachpb.RESUME_KEY_LIMIT
	} else {
		__antithesis_instrumentation__.Notify(96511)
	}
	__antithesis_instrumentation__.Notify(96503)

	return result.FromAcquiredLocks(h.Txn, deleted...), err
}
