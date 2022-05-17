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
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	"github.com/kr/pretty"
)

const ClearRangeBytesThreshold = 512 << 10

func init() {
	RegisterReadWriteCommand(roachpb.ClearRange, declareKeysClearRange, ClearRange)
}

func declareKeysClearRange(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(96434)
	DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

func ClearRange(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96435)
	if cArgs.Header.Txn != nil {
		__antithesis_instrumentation__.Notify(96442)
		return result.Result{}, errors.New("cannot execute ClearRange within a transaction")
	} else {
		__antithesis_instrumentation__.Notify(96443)
	}
	__antithesis_instrumentation__.Notify(96436)
	log.VEventf(ctx, 2, "ClearRange %+v", cArgs.Args)

	args := cArgs.Args.(*roachpb.ClearRangeRequest)
	from := args.Key
	to := args.EndKey

	if !args.Deadline.IsEmpty() {
		__antithesis_instrumentation__.Notify(96444)
		if now := cArgs.EvalCtx.Clock().Now(); args.Deadline.LessEq(now) {
			__antithesis_instrumentation__.Notify(96445)
			return result.Result{}, errors.Errorf("ClearRange has deadline %s <= %s", args.Deadline, now)
		} else {
			__antithesis_instrumentation__.Notify(96446)
		}
	} else {
		__antithesis_instrumentation__.Notify(96447)
	}
	__antithesis_instrumentation__.Notify(96437)

	pd := result.Result{
		Replicated: kvserverpb.ReplicatedEvalResult{
			MVCCHistoryMutation: &kvserverpb.ReplicatedEvalResult_MVCCHistoryMutation{
				Spans: []roachpb.Span{{Key: from, EndKey: to}},
			},
		},
	}

	maxIntents := storage.MaxIntentsPerWriteIntentError.Get(&cArgs.EvalCtx.ClusterSettings().SV)
	intents, err := storage.ScanIntents(ctx, readWriter, from, to, maxIntents, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(96448)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96449)
		if len(intents) > 0 {
			__antithesis_instrumentation__.Notify(96450)
			return result.Result{}, &roachpb.WriteIntentError{Intents: intents}
		} else {
			__antithesis_instrumentation__.Notify(96451)
		}
	}
	__antithesis_instrumentation__.Notify(96438)

	statsDelta, err := computeStatsDelta(ctx, readWriter, cArgs, from, to)
	if err != nil {
		__antithesis_instrumentation__.Notify(96452)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96453)
	}
	__antithesis_instrumentation__.Notify(96439)
	cArgs.Stats.Subtract(statsDelta)

	if statsDelta.ContainsEstimates == 0 && func() bool {
		__antithesis_instrumentation__.Notify(96454)
		return statsDelta.Total() < ClearRangeBytesThreshold == true
	}() == true {
		__antithesis_instrumentation__.Notify(96455)
		log.VEventf(ctx, 2, "delta=%d < threshold=%d; using non-range clear",
			statsDelta.Total(), ClearRangeBytesThreshold)
		iter := readWriter.NewMVCCIterator(storage.MVCCKeyAndIntentsIterKind, storage.IterOptions{
			LowerBound: from,
			UpperBound: to,
		})
		defer iter.Close()
		if err = readWriter.ClearIterRange(iter, from, to); err != nil {
			__antithesis_instrumentation__.Notify(96457)
			return result.Result{}, err
		} else {
			__antithesis_instrumentation__.Notify(96458)
		}
		__antithesis_instrumentation__.Notify(96456)
		return pd, nil
	} else {
		__antithesis_instrumentation__.Notify(96459)
	}
	__antithesis_instrumentation__.Notify(96440)

	if err := readWriter.ClearMVCCRangeAndIntents(from, to); err != nil {
		__antithesis_instrumentation__.Notify(96460)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(96461)
	}
	__antithesis_instrumentation__.Notify(96441)
	return pd, nil
}

func computeStatsDelta(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, from, to roachpb.Key,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(96462)
	desc := cArgs.EvalCtx.Desc()
	var delta enginepb.MVCCStats

	fast := desc.StartKey.Equal(from) && func() bool {
		__antithesis_instrumentation__.Notify(96465)
		return desc.EndKey.Equal(to) == true
	}() == true
	if fast {
		__antithesis_instrumentation__.Notify(96466)

		delta = cArgs.EvalCtx.GetMVCCStats()
		delta.SysCount, delta.SysBytes, delta.AbortSpanBytes = 0, 0, 0
	} else {
		__antithesis_instrumentation__.Notify(96467)
	}
	__antithesis_instrumentation__.Notify(96463)

	if !fast || func() bool {
		__antithesis_instrumentation__.Notify(96468)
		return util.RaceEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(96469)
		iter := readWriter.NewMVCCIterator(storage.MVCCKeyAndIntentsIterKind, storage.IterOptions{UpperBound: to})
		computed, err := iter.ComputeStats(from, to, delta.LastUpdateNanos)
		iter.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(96472)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(96473)
		}
		__antithesis_instrumentation__.Notify(96470)

		if fast {
			__antithesis_instrumentation__.Notify(96474)
			computed.ContainsEstimates = delta.ContainsEstimates
			if !delta.Equal(computed) {
				__antithesis_instrumentation__.Notify(96475)
				log.Fatalf(ctx, "fast-path MVCCStats computation gave wrong result: diff(fast, computed) = %s",
					pretty.Diff(delta, computed))
			} else {
				__antithesis_instrumentation__.Notify(96476)
			}
		} else {
			__antithesis_instrumentation__.Notify(96477)
		}
		__antithesis_instrumentation__.Notify(96471)
		delta = computed
	} else {
		__antithesis_instrumentation__.Notify(96478)
	}
	__antithesis_instrumentation__.Notify(96464)

	return delta, nil
}
