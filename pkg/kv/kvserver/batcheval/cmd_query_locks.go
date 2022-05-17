package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadOnlyCommand(roachpb.QueryLocks, declareKeysQueryLocks, QueryLocks)
}

func declareKeysQueryLocks(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97188)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

func QueryLocks(
	ctx context.Context, _ storage.Reader, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97189)
	args := cArgs.Args.(*roachpb.QueryLocksRequest)
	h := cArgs.Header
	reply := resp.(*roachpb.QueryLocksResponse)

	concurrencyManager := cArgs.EvalCtx.GetConcurrencyManager()
	keyScope := spanset.SpanGlobal
	if keys.IsLocal(args.Key) {
		__antithesis_instrumentation__.Notify(97191)
		keyScope = spanset.SpanLocal
	} else {
		__antithesis_instrumentation__.Notify(97192)
	}
	__antithesis_instrumentation__.Notify(97190)
	opts := concurrency.QueryLockTableOptions{
		KeyScope:           keyScope,
		MaxLocks:           h.MaxSpanRequestKeys,
		TargetBytes:        h.TargetBytes,
		IncludeUncontended: args.IncludeUncontended,
	}

	lockInfos, resumeState := concurrencyManager.QueryLockTableState(ctx, args.Span(), opts)

	reply.Locks = lockInfos
	reply.NumKeys = int64(len(lockInfos))
	reply.NumBytes = resumeState.TotalBytes
	reply.ResumeReason = resumeState.ResumeReason
	reply.ResumeSpan = resumeState.ResumeSpan
	reply.ResumeNextBytes = resumeState.ResumeNextBytes

	return result.Result{}, nil
}
