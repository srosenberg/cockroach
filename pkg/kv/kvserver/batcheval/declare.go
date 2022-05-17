package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
)

func DefaultDeclareKeys(
	_ ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97567)
	access := spanset.SpanReadWrite
	if roachpb.IsReadOnly(req) && func() bool {
		__antithesis_instrumentation__.Notify(97569)
		return !roachpb.IsLocking(req) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97570)
		access = spanset.SpanReadOnly
	} else {
		__antithesis_instrumentation__.Notify(97571)
	}
	__antithesis_instrumentation__.Notify(97568)
	latchSpans.AddMVCC(access, req.Header().Span(), header.Timestamp)
}

func DefaultDeclareIsolatedKeys(
	_ ImmutableRangeState,
	header *roachpb.Header,
	req roachpb.Request,
	latchSpans, lockSpans *spanset.SpanSet,
	maxOffset time.Duration,
) {
	__antithesis_instrumentation__.Notify(97572)
	access := spanset.SpanReadWrite
	timestamp := header.Timestamp
	if roachpb.IsReadOnly(req) && func() bool {
		__antithesis_instrumentation__.Notify(97574)
		return !roachpb.IsLocking(req) == true
	}() == true {
		__antithesis_instrumentation__.Notify(97575)
		access = spanset.SpanReadOnly

		in := uncertainty.ComputeInterval(header, kvserverpb.LeaseStatus{}, maxOffset)
		timestamp.Forward(in.GlobalLimit)
	} else {
		__antithesis_instrumentation__.Notify(97576)
	}
	__antithesis_instrumentation__.Notify(97573)
	latchSpans.AddMVCC(access, req.Header().Span(), timestamp)
	lockSpans.AddNonMVCC(access, req.Header().Span())
}

func DeclareKeysForBatch(
	rs ImmutableRangeState, header *roachpb.Header, latchSpans *spanset.SpanSet,
) {
	__antithesis_instrumentation__.Notify(97577)
	if header.Txn != nil {
		__antithesis_instrumentation__.Notify(97578)
		header.Txn.AssertInitialized(context.TODO())
		latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{
			Key: keys.AbortSpanKey(rs.GetRangeID(), header.Txn.ID),
		})
	} else {
		__antithesis_instrumentation__.Notify(97579)
	}
}

func declareAllKeys(latchSpans *spanset.SpanSet) {
	__antithesis_instrumentation__.Notify(97580)

	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.LocalPrefix, EndKey: keys.LocalMax})
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.MinKey, EndKey: keys.MaxKey})
}

type CommandArgs struct {
	EvalCtx EvalContext
	Header  roachpb.Header
	Args    roachpb.Request

	Stats       *enginepb.MVCCStats
	Uncertainty uncertainty.Interval
}
