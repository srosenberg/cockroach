package uncertainty

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func ComputeInterval(
	h *roachpb.Header, status kvserverpb.LeaseStatus, maxOffset time.Duration,
) Interval {
	__antithesis_instrumentation__.Notify(127533)
	if h.Txn != nil {
		__antithesis_instrumentation__.Notify(127535)
		return computeIntervalForTxn(h.Txn, status)
	} else {
		__antithesis_instrumentation__.Notify(127536)
	}
	__antithesis_instrumentation__.Notify(127534)
	return computeIntervalForNonTxn(h, status, maxOffset)
}

func computeIntervalForTxn(txn *roachpb.Transaction, status kvserverpb.LeaseStatus) Interval {
	__antithesis_instrumentation__.Notify(127537)
	in := Interval{

		GlobalLimit: txn.GlobalUncertaintyLimit,
	}

	if status.State != kvserverpb.LeaseState_VALID {
		__antithesis_instrumentation__.Notify(127540)

		return in
	} else {
		__antithesis_instrumentation__.Notify(127541)
	}
	__antithesis_instrumentation__.Notify(127538)

	obsTs, ok := txn.GetObservedTimestamp(status.Lease.Replica.NodeID)
	if !ok {
		__antithesis_instrumentation__.Notify(127542)
		return in
	} else {
		__antithesis_instrumentation__.Notify(127543)
	}
	__antithesis_instrumentation__.Notify(127539)
	in.LocalLimit = obsTs

	in.LocalLimit.Forward(minimumLocalLimitForLeaseholder(status.Lease))

	in.LocalLimit.BackwardWithTimestamp(in.GlobalLimit)
	return in
}

func computeIntervalForNonTxn(
	h *roachpb.Header, status kvserverpb.LeaseStatus, maxOffset time.Duration,
) Interval {
	__antithesis_instrumentation__.Notify(127544)
	if h.TimestampFromServerClock == nil || func() bool {
		__antithesis_instrumentation__.Notify(127547)
		return h.ReadConsistency != roachpb.CONSISTENT == true
	}() == true {
		__antithesis_instrumentation__.Notify(127548)

		return Interval{}
	} else {
		__antithesis_instrumentation__.Notify(127549)
	}
	__antithesis_instrumentation__.Notify(127545)

	in := Interval{

		GlobalLimit: h.TimestampFromServerClock.ToTimestamp().Add(maxOffset.Nanoseconds(), 0),
	}

	if status.State != kvserverpb.LeaseState_VALID {
		__antithesis_instrumentation__.Notify(127550)

		return in
	} else {
		__antithesis_instrumentation__.Notify(127551)
	}
	__antithesis_instrumentation__.Notify(127546)

	in.LocalLimit = *h.TimestampFromServerClock

	in.LocalLimit.Forward(minimumLocalLimitForLeaseholder(status.Lease))

	in.LocalLimit.BackwardWithTimestamp(in.GlobalLimit)
	return in
}

func minimumLocalLimitForLeaseholder(lease roachpb.Lease) hlc.ClockTimestamp {
	__antithesis_instrumentation__.Notify(127552)

	min := lease.Start

	return min
}
