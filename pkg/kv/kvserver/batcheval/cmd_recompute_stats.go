package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadOnlyCommand(roachpb.RecomputeStats, declareKeysRecomputeStats, RecomputeStats)
}

func declareKeysRecomputeStats(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97249)

	rdKey := keys.RangeDescriptorKey(rs.GetStartKey())
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: rdKey})
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.TransactionKey(rdKey, uuid.Nil)})
}

func RecomputeStats(
	ctx context.Context, reader storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97250)
	desc := cArgs.EvalCtx.Desc()
	args := cArgs.Args.(*roachpb.RecomputeStatsRequest)
	if !desc.StartKey.AsRawKey().Equal(args.Key) {
		__antithesis_instrumentation__.Notify(97255)
		return result.Result{}, errors.New("descriptor mismatch; range likely merged")
	} else {
		__antithesis_instrumentation__.Notify(97256)
	}
	__antithesis_instrumentation__.Notify(97251)
	dryRun := args.DryRun

	args = nil

	reader = spanset.DisableReaderAssertions(reader)

	actualMS, err := rditer.ComputeStatsForRange(desc, reader, cArgs.Header.Timestamp.WallTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(97257)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97258)
	}
	__antithesis_instrumentation__.Notify(97252)

	currentStats, err := MakeStateLoader(cArgs.EvalCtx).LoadMVCCStats(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(97259)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97260)
	}
	__antithesis_instrumentation__.Notify(97253)

	delta := actualMS
	delta.Subtract(currentStats)

	if !dryRun {
		__antithesis_instrumentation__.Notify(97261)

		cArgs.Stats.Add(delta)
	} else {
		__antithesis_instrumentation__.Notify(97262)
	}
	__antithesis_instrumentation__.Notify(97254)

	resp.(*roachpb.RecomputeStatsResponse).AddedDelta = enginepb.MVCCStatsDelta(delta)
	return result.Result{}, nil
}
