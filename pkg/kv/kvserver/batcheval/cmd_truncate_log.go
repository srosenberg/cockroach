package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func init() {
	RegisterReadWriteCommand(roachpb.TruncateLog, declareKeysTruncateLog, TruncateLog)
}

func declareKeysTruncateLog(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97533)
	prefix := keys.RaftLogPrefix(rs.GetRangeID())
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: prefix, EndKey: prefix.PrefixEnd()})
}

func TruncateLog(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97534)
	args := cArgs.Args.(*roachpb.TruncateLogRequest)

	rangeID := cArgs.EvalCtx.GetRangeID()
	if rangeID != args.RangeID {
		__antithesis_instrumentation__.Notify(97542)
		log.Infof(ctx, "attempting to truncate raft logs for another range: r%d. Normally this is due to a merge and can be ignored.",
			args.RangeID)
		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(97543)
	}
	__antithesis_instrumentation__.Notify(97535)

	firstIndex, err := cArgs.EvalCtx.GetFirstIndex()
	if err != nil {
		__antithesis_instrumentation__.Notify(97544)
		return result.Result{}, errors.Wrap(err, "getting first index")
	} else {
		__antithesis_instrumentation__.Notify(97545)
	}
	__antithesis_instrumentation__.Notify(97536)

	if firstIndex >= args.Index {
		__antithesis_instrumentation__.Notify(97546)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(97548)
			log.Infof(ctx, "attempting to truncate previously truncated raft log. FirstIndex:%d, TruncateFrom:%d",
				firstIndex, args.Index)
		} else {
			__antithesis_instrumentation__.Notify(97549)
		}
		__antithesis_instrumentation__.Notify(97547)
		return result.Result{}, nil
	} else {
		__antithesis_instrumentation__.Notify(97550)
	}
	__antithesis_instrumentation__.Notify(97537)

	term, err := cArgs.EvalCtx.GetTerm(args.Index - 1)
	if err != nil {
		__antithesis_instrumentation__.Notify(97551)
		return result.Result{}, errors.Wrap(err, "getting term")
	} else {
		__antithesis_instrumentation__.Notify(97552)
	}
	__antithesis_instrumentation__.Notify(97538)

	if args.ExpectedFirstIndex > firstIndex {
		__antithesis_instrumentation__.Notify(97553)
		firstIndex = args.ExpectedFirstIndex
	} else {
		__antithesis_instrumentation__.Notify(97554)
	}
	__antithesis_instrumentation__.Notify(97539)
	start := keys.RaftLogKey(rangeID, firstIndex)
	end := keys.RaftLogKey(rangeID, args.Index)

	iter := readWriter.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{UpperBound: end})
	defer iter.Close()

	ms, err := iter.ComputeStats(start, end, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(97555)
		return result.Result{}, errors.Wrap(err, "while computing stats of Raft log freed by truncation")
	} else {
		__antithesis_instrumentation__.Notify(97556)
	}
	__antithesis_instrumentation__.Notify(97540)
	ms.SysBytes = -ms.SysBytes

	tState := &roachpb.RaftTruncatedState{
		Index: args.Index - 1,
		Term:  term,
	}

	var pd result.Result
	pd.Replicated.State = &kvserverpb.ReplicaState{
		TruncatedState: tState,
	}
	pd.Replicated.RaftLogDelta = ms.SysBytes
	if cArgs.EvalCtx.ClusterSettings().Version.ActiveVersionOrEmpty(ctx).IsActive(
		clusterversion.LooselyCoupledRaftLogTruncation) {
		__antithesis_instrumentation__.Notify(97557)
		pd.Replicated.RaftExpectedFirstIndex = firstIndex
	} else {
		__antithesis_instrumentation__.Notify(97558)
	}
	__antithesis_instrumentation__.Notify(97541)
	return pd, nil
}
