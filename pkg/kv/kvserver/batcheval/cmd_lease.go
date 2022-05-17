package batcheval

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/batcheval/result"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
)

func newFailedLeaseTrigger(isTransfer bool) result.Result {
	__antithesis_instrumentation__.Notify(96967)
	var trigger result.Result
	trigger.Local.Metrics = new(result.Metrics)
	if isTransfer {
		__antithesis_instrumentation__.Notify(96969)
		trigger.Local.Metrics.LeaseTransferError = 1
	} else {
		__antithesis_instrumentation__.Notify(96970)
		trigger.Local.Metrics.LeaseRequestError = 1
	}
	__antithesis_instrumentation__.Notify(96968)
	return trigger
}

func evalNewLease(
	ctx context.Context,
	rec EvalContext,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	lease roachpb.Lease,
	prevLease roachpb.Lease,
	priorReadSum *rspb.ReadSummary,
	isExtension bool,
	isTransfer bool,
) (result.Result, error) {
	__antithesis_instrumentation__.Notify(96971)

	if (lease.Type() == roachpb.LeaseExpiration && func() bool {
		__antithesis_instrumentation__.Notify(96980)
		return lease.GetExpiration().LessEq(lease.Start.ToTimestamp()) == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(96981)
		return (lease.Type() == roachpb.LeaseEpoch && func() bool {
			__antithesis_instrumentation__.Notify(96982)
			return lease.Expiration != nil == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(96983)

		return newFailedLeaseTrigger(isTransfer),
			&roachpb.LeaseRejectedError{
				Existing:  prevLease,
				Requested: lease,
				Message: fmt.Sprintf("illegal lease: epoch=%d, interval=[%s, %s)",
					lease.Epoch, lease.Start, lease.Expiration),
			}
	} else {
		__antithesis_instrumentation__.Notify(96984)
	}
	__antithesis_instrumentation__.Notify(96972)

	desc := rec.Desc()
	if _, ok := desc.GetReplicaDescriptor(lease.Replica.StoreID); !ok {
		__antithesis_instrumentation__.Notify(96985)
		return newFailedLeaseTrigger(isTransfer),
			&roachpb.LeaseRejectedError{
				Existing:  prevLease,
				Requested: lease,
				Message:   "replica not found",
			}
	} else {
		__antithesis_instrumentation__.Notify(96986)
	}
	__antithesis_instrumentation__.Notify(96973)

	if lease.Sequence != 0 {
		__antithesis_instrumentation__.Notify(96987)
		return newFailedLeaseTrigger(isTransfer),
			&roachpb.LeaseRejectedError{
				Existing:  prevLease,
				Requested: lease,
				Message:   "sequence number should not be set",
			}
	} else {
		__antithesis_instrumentation__.Notify(96988)
	}
	__antithesis_instrumentation__.Notify(96974)
	if prevLease.Equivalent(lease) {
		__antithesis_instrumentation__.Notify(96989)

		lease.Sequence = prevLease.Sequence
	} else {
		__antithesis_instrumentation__.Notify(96990)

		lease.Sequence = prevLease.Sequence + 1
	}
	__antithesis_instrumentation__.Notify(96975)

	if isTransfer {
		__antithesis_instrumentation__.Notify(96991)
		lease.AcquisitionType = roachpb.LeaseAcquisitionType_Transfer
	} else {
		__antithesis_instrumentation__.Notify(96992)
		lease.AcquisitionType = roachpb.LeaseAcquisitionType_Request
	}
	__antithesis_instrumentation__.Notify(96976)

	if err := MakeStateLoader(rec).SetLease(ctx, readWriter, ms, lease); err != nil {
		__antithesis_instrumentation__.Notify(96993)
		return newFailedLeaseTrigger(isTransfer), err
	} else {
		__antithesis_instrumentation__.Notify(96994)
	}
	__antithesis_instrumentation__.Notify(96977)

	var pd result.Result
	pd.Replicated.State = &kvserverpb.ReplicaState{
		Lease: &lease,
	}
	pd.Replicated.PrevLeaseProposal = prevLease.ProposedTS

	if priorReadSum != nil {
		__antithesis_instrumentation__.Notify(96995)
		if err := readsummary.Set(ctx, readWriter, rec.GetRangeID(), ms, priorReadSum); err != nil {
			__antithesis_instrumentation__.Notify(96997)
			return newFailedLeaseTrigger(isTransfer), err
		} else {
			__antithesis_instrumentation__.Notify(96998)
		}
		__antithesis_instrumentation__.Notify(96996)
		pd.Replicated.PriorReadSummary = priorReadSum
	} else {
		__antithesis_instrumentation__.Notify(96999)
	}
	__antithesis_instrumentation__.Notify(96978)

	pd.Local.Metrics = new(result.Metrics)
	if isTransfer {
		__antithesis_instrumentation__.Notify(97000)
		pd.Local.Metrics.LeaseTransferSuccess = 1
	} else {
		__antithesis_instrumentation__.Notify(97001)
		pd.Local.Metrics.LeaseRequestSuccess = 1
	}
	__antithesis_instrumentation__.Notify(96979)
	return pd, nil
}
