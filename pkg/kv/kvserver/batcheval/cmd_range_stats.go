package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadOnlyCommand(roachpb.RangeStats, declareKeysRangeStats, RangeStats)
}

func declareKeysRangeStats(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(97244)
	DefaultDeclareKeys(rs, header, req, latchSpans, lockSpans, maxOffset)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeLeaseKey(rs.GetRangeID())})
}

func RangeStats(
	ctx context.Context, _ storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97245)
	reply := resp.(*roachpb.RangeStatsResponse)
	reply.MVCCStats = cArgs.EvalCtx.GetMVCCStats()
	reply.DeprecatedLastQueriesPerSecond = cArgs.EvalCtx.GetLastSplitQPS()
	if qps, ok := cArgs.EvalCtx.GetMaxSplitQPS(); ok {
		__antithesis_instrumentation__.Notify(97247)
		reply.MaxQueriesPerSecond = qps
	} else {
		__antithesis_instrumentation__.Notify(97248)

		reply.MaxQueriesPerSecond = -1
	}
	__antithesis_instrumentation__.Notify(97246)
	reply.MaxQueriesPerSecondSet = true
	reply.RangeInfo = cArgs.EvalCtx.GetRangeInfo(ctx)
	return result.Result{}, nil
}
