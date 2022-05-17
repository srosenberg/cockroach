package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanset"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

func init() {
	RegisterReadWriteCommand(roachpb.RequestLease, declareKeysRequestLease, RequestLease)
}

func declareKeysRequestLease(
	rs ImmutableRangeState,
	_ *roachpb.Header,
	_ roachpb.Request,
	latchSpans, _ *spanset.SpanSet,
	_ time.Duration,
) {
	__antithesis_instrumentation__.Notify(97007)

	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.RangeLeaseKey(rs.GetRangeID())})
	latchSpans.AddNonMVCC(spanset.SpanReadWrite, roachpb.Span{Key: keys.RangePriorReadSummaryKey(rs.GetRangeID())})
	latchSpans.AddNonMVCC(spanset.SpanReadOnly, roachpb.Span{Key: keys.RangeDescriptorKey(rs.GetStartKey())})
}

func RequestLease(
	ctx context.Context, readWriter storage.ReadWriter, cArgs CommandArgs, resp roachpb.Response,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(97008)

	args := cArgs.Args.(*roachpb.RequestLeaseRequest)

	prevLease, _ := cArgs.EvalCtx.GetLease()
	rErr := &roachpb.LeaseRejectedError{
		Existing:  prevLease,
		Requested: args.Lease,
	}

	if err := roachpb.CheckCanReceiveLease(args.Lease.Replica, cArgs.EvalCtx.Desc()); err != nil {
		__antithesis_instrumentation__.Notify(97014)
		rErr.Message = err.Error()
		return newFailedLeaseTrigger(false), rErr
	} else {
		__antithesis_instrumentation__.Notify(97015)
	}
	__antithesis_instrumentation__.Notify(97009)

	newLease := args.Lease
	if newLease.DeprecatedStartStasis == nil {
		__antithesis_instrumentation__.Notify(97016)
		newLease.DeprecatedStartStasis = newLease.Expiration
	} else {
		__antithesis_instrumentation__.Notify(97017)
	}
	__antithesis_instrumentation__.Notify(97010)

	isExtension := prevLease.Replica.StoreID == newLease.Replica.StoreID
	effectiveStart := newLease.Start

	if prevLease.Replica.StoreID == 0 || func() bool {
		__antithesis_instrumentation__.Notify(97018)
		return isExtension == true
	}() == true {
		__antithesis_instrumentation__.Notify(97019)
		effectiveStart.Backward(prevLease.Start)

		if ts := args.MinProposedTS; isExtension && func() bool {
			__antithesis_instrumentation__.Notify(97020)
			return ts != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(97021)
			effectiveStart.Forward(*ts)
		} else {
			__antithesis_instrumentation__.Notify(97022)
		}

	} else {
		__antithesis_instrumentation__.Notify(97023)
		if prevLease.Type() == roachpb.LeaseExpiration {
			__antithesis_instrumentation__.Notify(97024)
			effectiveStart.BackwardWithTimestamp(prevLease.Expiration.Next())
		} else {
			__antithesis_instrumentation__.Notify(97025)
		}
	}
	__antithesis_instrumentation__.Notify(97011)

	if isExtension {
		__antithesis_instrumentation__.Notify(97026)
		if effectiveStart.Less(prevLease.Start) {
			__antithesis_instrumentation__.Notify(97028)
			rErr.Message = "extension moved start timestamp backwards"
			return newFailedLeaseTrigger(false), rErr
		} else {
			__antithesis_instrumentation__.Notify(97029)
		}
		__antithesis_instrumentation__.Notify(97027)
		if newLease.Type() == roachpb.LeaseExpiration {
			__antithesis_instrumentation__.Notify(97030)

			t := *newLease.Expiration
			newLease.Expiration = &t
			newLease.Expiration.Forward(prevLease.GetExpiration())
		} else {
			__antithesis_instrumentation__.Notify(97031)
		}
	} else {
		__antithesis_instrumentation__.Notify(97032)
		if prevLease.Type() == roachpb.LeaseExpiration && func() bool {
			__antithesis_instrumentation__.Notify(97033)
			return effectiveStart.ToTimestamp().Less(prevLease.GetExpiration()) == true
		}() == true {
			__antithesis_instrumentation__.Notify(97034)
			rErr.Message = "requested lease overlaps previous lease"
			return newFailedLeaseTrigger(false), rErr
		} else {
			__antithesis_instrumentation__.Notify(97035)
		}
	}
	__antithesis_instrumentation__.Notify(97012)
	newLease.Start = effectiveStart

	var priorReadSum *rspb.ReadSummary
	if !prevLease.Equivalent(newLease) {
		__antithesis_instrumentation__.Notify(97036)

		worstCaseSum := rspb.FromTimestamp(newLease.Start.ToTimestamp())
		priorReadSum = &worstCaseSum
	} else {
		__antithesis_instrumentation__.Notify(97037)
	}
	__antithesis_instrumentation__.Notify(97013)

	return evalNewLease(ctx, cArgs.EvalCtx, readWriter, cArgs.Stats,
		newLease, prevLease, priorReadSum, isExtension, false)
}
