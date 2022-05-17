package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadWriteCommand(roachpb.GC, declareKeysGC, GC)
}

func declareKeysGC(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(96873)

	gcr := req.(*roachpb.GCRequest)
	for _, key := range gcr.Keys {
		__antithesis_instrumentation__.Notify(96876)
		if keys.IsLocal(key.Key) {
			__antithesis_instrumentation__.Notify(96877)
			latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: key.Key})
		} else {
			__antithesis_instrumentation__.Notify(96878)
			latchSpans.AddMVCC(spanset.SpanReadWrite, roachpb.Span{Key: key.Key}, header.Timestamp)
		}
	}
	__antithesis_instrumentation__.Notify(96874)

	if !gcr.Threshold.IsEmpty() {
		__antithesis_instrumentation__.Notify(96879)
		latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.RangeGCThresholdKey(rs.GetRangeID())})
	} else {
		__antithesis_instrumentation__.Notify(96880)
	}
	__antithesis_instrumentation__.Notify(96875)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

func GC(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96881)
	args := cArgs.Args.(*roachpb.GCRequest)
	h := cArgs.Header

	if !args.Threshold.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(96886)
		return len(args.Keys) != 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(96887)
		return !cArgs.EvalCtx.EvalKnobs().AllowGCWithNewThresholdAndKeys == true
	}() == true {
		__antithesis_instrumentation__.Notify(96888)
		return result.Result{}, errors.AssertionFailedf(
			"GC request can set threshold or it can GC keys, but it is unsafe for it to do both")
	} else {
		__antithesis_instrumentation__.Notify(96889)
	}
	__antithesis_instrumentation__.Notify(96882)

	globalKeys := make([]roachpb.GCRequest_GCKey, 0, len(args.Keys))

	var localKeys []roachpb.GCRequest_GCKey
	for _, k := range args.Keys {
		__antithesis_instrumentation__.Notify(96890)
		if cArgs.EvalCtx.ContainsKey(k.Key) {
			__antithesis_instrumentation__.Notify(96891)
			if keys.IsLocal(k.Key) {
				__antithesis_instrumentation__.Notify(96892)
				localKeys = append(localKeys, k)
			} else {
				__antithesis_instrumentation__.Notify(96893)
				globalKeys = append(globalKeys, k)
			}
		} else {
			__antithesis_instrumentation__.Notify(96894)
		}
	}
	__antithesis_instrumentation__.Notify(96883)

	for _, gcKeys := range [][]roachpb.GCRequest_GCKey{localKeys, globalKeys} {
		__antithesis_instrumentation__.Notify(96895)
		if err := storage.MVCCGarbageCollect(
			ctx, readWriter, cArgs.Stats, gcKeys, h.Timestamp,
		); err != nil {
			__antithesis_instrumentation__.Notify(96896)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96897)
		}
	}
	__antithesis_instrumentation__.Notify(96884)

	var res result.Result
	if !args.Threshold.IsEmpty() {
		__antithesis_instrumentation__.Notify(96898)
		oldThreshold := cArgs.EvalCtx.GetGCThreshold()

		newThreshold := oldThreshold
		updated := newThreshold.Forward(args.Threshold)

		if updated {
			__antithesis_instrumentation__.Notify(96899)
			if err := MakeStateLoader(cArgs.EvalCtx).SetGCThreshold(
				ctx, readWriter, cArgs.Stats, &newThreshold,
			); err != nil {
				__antithesis_instrumentation__.Notify(96901)
				return result.Result{}, err
			} else {
				__antithesis_instrumentation__.Notify(96902)
			}
			__antithesis_instrumentation__.Notify(96900)

			res.Replicated.State = &kvserverpb.ReplicaState{
				GCThreshold: &newThreshold,
			}
		} else {
			__antithesis_instrumentation__.Notify(96903)
		}
	} else {
		__antithesis_instrumentation__.Notify(96904)
	}
	__antithesis_instrumentation__.Notify(96885)

	return res, nil
}
