package kvserverpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func (st LeaseStatus) IsValid() bool {
	__antithesis_instrumentation__.Notify(101858)
	return st.State == LeaseState_VALID
}

func (st LeaseStatus) OwnedBy(storeID roachpb.StoreID) bool {
	__antithesis_instrumentation__.Notify(101859)
	return st.Lease.OwnedBy(storeID)
}

func (st LeaseStatus) Expiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(101860)
	switch st.Lease.Type() {
	case roachpb.LeaseExpiration:
		__antithesis_instrumentation__.Notify(101861)
		return st.Lease.GetExpiration()
	case roachpb.LeaseEpoch:
		__antithesis_instrumentation__.Notify(101862)
		return st.Liveness.Expiration.ToTimestamp()
	default:
		__antithesis_instrumentation__.Notify(101863)
		panic("unexpected")
	}
}

func (st LeaseStatus) ClosedTimestampUpperBound() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(101864)

	return st.Expiration().WithSynthetic(true).WallPrev()
}
