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
	RegisterReadWriteCommand(roachpb.Put, declareKeysPut, Put)
}

func declareKeysPut(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(97143)
	args := req.(*roachpb.PutRequest)
	if args.Inline {
		__antithesis_instrumentation__.Notify(97144)
		DefaultDeclareKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	} else {
		__antithesis_instrumentation__.Notify(97145)
		DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)
	}
}

func Put(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97146)
	args := cArgs.Args.(*roachpb.PutRequest)
	h := cArgs.Header
	ms := cArgs.Stats

	var ts hlc.Timestamp
	if !args.Inline {
		__antithesis_instrumentation__.Notify(97149)
		ts = h.Timestamp
	} else {
		__antithesis_instrumentation__.Notify(97150)
	}
	__antithesis_instrumentation__.Notify(97147)
	var err error
	if args.Blind {
		__antithesis_instrumentation__.Notify(97151)
		err = storage.MVCCBlindPut(ctx, readWriter, ms, args.Key, ts, args.Value, h.Txn)
	} else {
		__antithesis_instrumentation__.Notify(97152)
		err = storage.MVCCPut(ctx, readWriter, ms, args.Key, ts, args.Value, h.Txn)
	}
	__antithesis_instrumentation__.Notify(97148)

	return result.FromAcquiredLocks(h.Txn, args.Key), err
}
