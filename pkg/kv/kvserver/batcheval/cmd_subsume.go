package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadWriteCommand(roachpb.Subsume, declareKeysSubsume, Subsume)
}

func declareKeysSubsume(
	_ ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97511)

	declareAllKeys(latchSpans)
}

func Subsume(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97512)
	args := cArgs.Args.(*roachpb.SubsumeRequest)
	reply := resp.(*roachpb.SubsumeResponse)

	desc := cArgs.EvalCtx.Desc()
	if !bytes.Equal(desc.StartKey, args.RightDesc.StartKey) || func() bool {
		__antithesis_instrumentation__.Notify(97518)
		return !bytes.Equal(desc.EndKey, args.RightDesc.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97519)
		return result.Result{}, errors.Errorf("RHS range bounds do not match: %s != %s",
			args.RightDesc, desc)
	} else {
		__antithesis_instrumentation__.Notify(97520)
	}
	__antithesis_instrumentation__.Notify(97513)

	if !bytes.Equal(args.LeftDesc.EndKey, desc.StartKey) {
		__antithesis_instrumentation__.Notify(97521)
		return result.Result{}, errors.Errorf("ranges are not adjacent: %s != %s",
			args.LeftDesc.EndKey, desc.StartKey)
	} else {
		__antithesis_instrumentation__.Notify(97522)
	}
	__antithesis_instrumentation__.Notify(97514)

	descKey := keys.RangeDescriptorKey(desc.StartKey)
	_, intent, err := storage.MVCCGet(ctx, readWriter, descKey, hlc.MaxTimestamp,
		storage.MVCCGetOptions{Inconsistent: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(97523)
		return result.Result{}, errors.Wrap(err, "fetching local range descriptor")
	} else {
		__antithesis_instrumentation__.Notify(97524)
		if intent == nil {
			__antithesis_instrumentation__.Notify(97525)
			return result.Result{}, errors.Errorf("range missing intent on its local descriptor")
		} else {
			__antithesis_instrumentation__.Notify(97526)
		}
	}
	__antithesis_instrumentation__.Notify(97515)
	val, _, err := storage.MVCCGetAsTxn(ctx, readWriter, descKey, intent.Txn.WriteTimestamp, intent.Txn)
	if err != nil {
		__antithesis_instrumentation__.Notify(97527)
		return result.Result{}, errors.Wrap(err, "fetching local range descriptor as txn")
	} else {
		__antithesis_instrumentation__.Notify(97528)
		if val != nil {
			__antithesis_instrumentation__.Notify(97529)
			return result.Result{}, errors.Errorf("non-deletion intent on local range descriptor")
		} else {
			__antithesis_instrumentation__.Notify(97530)
		}
	}
	__antithesis_instrumentation__.Notify(97516)

	if err := cArgs.EvalCtx.WatchForMerge(ctx); err != nil {
		__antithesis_instrumentation__.Notify(97531)
		return result.Result{}, errors.Wrap(err, "watching for merge during subsume")
	} else {
		__antithesis_instrumentation__.Notify(97532)
	}
	__antithesis_instrumentation__.Notify(97517)

	reply.MVCCStats = cArgs.EvalCtx.GetMVCCStats()
	reply.LeaseAppliedIndex = cArgs.EvalCtx.GetLeaseAppliedIndex()
	reply.FreezeStart = cArgs.EvalCtx.Clock().NowAsClockTimestamp()

	priorReadSum := cArgs.EvalCtx.GetCurrentReadSummary(ctx)

	priorReadSum.Merge(rspb.FromTimestamp(reply.FreezeStart.ToTimestamp()))
	reply.ReadSummary = &priorReadSum
	reply.ClosedTimestamp = cArgs.EvalCtx.GetClosedTimestamp(ctx)

	return result.Result{}, nil
}
