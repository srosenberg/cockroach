package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func init() {
	RegisterReadWriteCommand(roachpb.TransferLease, declareKeysTransferLease, TransferLease)
}

func declareKeysTransferLease(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97038)

	declareAllKeys(latchSpans)
}

func TransferLease(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97039)

	args := cArgs.Args.(*roachpb.TransferLeaseRequest)

	prevLease, _ := cArgs.EvalCtx.GetLease()

	newLease := args.Lease
	args.Lease = roachpb.Lease{}

	if err := roachpb.CheckCanReceiveLease(newLease.Replica, cArgs.EvalCtx.Desc()); err != nil {
		__antithesis_instrumentation__.Notify(97041)
		return newFailedLeaseTrigger(true), err
	} else {
		__antithesis_instrumentation__.Notify(97042)
	}
	__antithesis_instrumentation__.Notify(97040)

	cArgs.EvalCtx.RevokeLease(ctx, args.PrevLease.Sequence)

	newLease.Start.Forward(cArgs.EvalCtx.Clock().NowAsClockTimestamp())

	priorReadSum := cArgs.EvalCtx.GetCurrentReadSummary(ctx)

	priorReadSum.Merge(rspb.FromTimestamp(newLease.Start.ToTimestamp()))

	log.VEventf(ctx, 2, "lease transfer: prev lease: %+v, new lease: %+v", prevLease, newLease)
	return evalNewLease(ctx, cArgs.EvalCtx, readWriter, cArgs.Stats,
		newLease, prevLease, &priorReadSum, false, true)
}
