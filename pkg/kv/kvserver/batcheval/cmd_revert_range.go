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
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func init() {
	RegisterReadWriteCommand(roachpb.RevertRange, declareKeysRevertRange, RevertRange)
}

func declareKeysRevertRange(
	rs ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(97441)
	DefaultDeclareIsolatedKeys(rs, header, req, latchSpans, lockSpans, maxOffset)

	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeGCThresholdKey(rs.GetRangeID())})
}

func isEmptyKeyTimeRange(
	readWriter storage.ReadWriter, from, to roachpb.Key, since, until hlc.Timestamp,
) (bool, error) {
	__antithesis_instrumentation__.Notify(97442)

	iter := readWriter.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		LowerBound: from, UpperBound: to,
		MinTimestampHint: since.Next(), MaxTimestampHint: until,
	})
	defer iter.Close()
	iter.SeekGE(storage.MVCCKey{Key: from})
	ok, err := iter.Valid()
	return !ok, err
}

const maxRevertRangeBatchBytes = 32 << 20

func RevertRange(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97443)
	if cArgs.Header.Txn != nil {
		__antithesis_instrumentation__.Notify(97448)
		return result.Result{}, ErrTransactionUnsupported
	} else {
		__antithesis_instrumentation__.Notify(97449)
	}
	__antithesis_instrumentation__.Notify(97444)
	log.VEventf(ctx, 2, "RevertRange %+v", cArgs.Args)

	args := cArgs.Args.(*roachpb.RevertRangeRequest)
	reply := resp.(*roachpb.RevertRangeResponse)
	pd := result.Result{
		Replicated: kvserverpb.ReplicatedEvalResult{
			MVCCHistoryMutation: &kvserverpb.ReplicatedEvalResult_MVCCHistoryMutation{
				Spans: []roachpb.Span{{Key: args.Key, EndKey: args.EndKey}},
			},
		},
	}

	if empty, err := isEmptyKeyTimeRange(
		readWriter, args.Key, args.EndKey, args.TargetTime, cArgs.Header.Timestamp,
	); err != nil {
		__antithesis_instrumentation__.Notify(97450)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97451)
		if empty {
			__antithesis_instrumentation__.Notify(97452)
			log.VEventf(ctx, 2, "no keys to clear in specified time range")
			return result.Result{}, nil
		} else {
			__antithesis_instrumentation__.Notify(97453)
		}
	}
	__antithesis_instrumentation__.Notify(97445)

	log.VEventf(ctx, 2, "clearing keys with timestamp (%v, %v]", args.TargetTime, cArgs.Header.Timestamp)

	resume, err := storage.MVCCClearTimeRange(ctx, readWriter, cArgs.Stats, args.Key, args.EndKey,
		args.TargetTime, cArgs.Header.Timestamp, cArgs.Header.MaxSpanRequestKeys,
		maxRevertRangeBatchBytes,
		args.EnableTimeBoundIteratorOptimization)
	if err != nil {
		__antithesis_instrumentation__.Notify(97454)
		return result.Result{}, err
	} else {
		__antithesis_instrumentation__.Notify(97455)
	}
	__antithesis_instrumentation__.Notify(97446)

	if resume != nil {
		__antithesis_instrumentation__.Notify(97456)
		log.VEventf(ctx, 2, "hit limit while clearing keys, resume span [%v, %v)", resume.Key, resume.EndKey)
		reply.ResumeSpan = resume

		reply.NumKeys = cArgs.Header.MaxSpanRequestKeys
		reply.ResumeReason = roachpb.RESUME_KEY_LIMIT
	} else {
		__antithesis_instrumentation__.Notify(97457)
	}
	__antithesis_instrumentation__.Notify(97447)

	return pd, nil
}
