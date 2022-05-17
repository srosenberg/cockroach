package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/sidetransport"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

func (r *Replica) BumpSideTransportClosed(
	ctx context.Context,
	now hlc.ClockTimestamp,
	targetByPolicy [roachpb.MAX_CLOSED_TIMESTAMP_POLICY]hlc.Timestamp,
) sidetransport.BumpSideTransportClosedResult {
	__antithesis_instrumentation__.Notify(115768)
	var res sidetransport.BumpSideTransportClosedResult
	r.mu.Lock()
	defer r.mu.Unlock()
	res.Desc = r.descRLocked()

	if _, err := r.isDestroyedRLocked(); err != nil {
		__antithesis_instrumentation__.Notify(115775)
		res.FailReason = sidetransport.ReplicaDestroyed
		return res
	} else {
		__antithesis_instrumentation__.Notify(115776)
	}
	__antithesis_instrumentation__.Notify(115769)

	lai := ctpb.LAI(r.mu.state.LeaseAppliedIndex)
	policy := r.closedTimestampPolicyRLocked()
	target := targetByPolicy[policy]
	st := r.leaseStatusForRequestRLocked(ctx, now, hlc.Timestamp{})

	valid := st.IsValid() || func() bool {
		__antithesis_instrumentation__.Notify(115777)
		return st.State == kvserverpb.LeaseState_UNUSABLE == true
	}() == true
	if !valid || func() bool {
		__antithesis_instrumentation__.Notify(115778)
		return !st.OwnedBy(r.StoreID()) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115779)
		res.FailReason = sidetransport.InvalidLease
		return res
	} else {
		__antithesis_instrumentation__.Notify(115780)
	}
	__antithesis_instrumentation__.Notify(115770)
	if st.ClosedTimestampUpperBound().Less(target) {
		__antithesis_instrumentation__.Notify(115781)
		res.FailReason = sidetransport.TargetOverLeaseExpiration
		return res
	} else {
		__antithesis_instrumentation__.Notify(115782)
	}
	__antithesis_instrumentation__.Notify(115771)

	if r.mergeInProgressRLocked() {
		__antithesis_instrumentation__.Notify(115783)
		res.FailReason = sidetransport.MergeInProgress
		return res
	} else {
		__antithesis_instrumentation__.Notify(115784)
	}
	__antithesis_instrumentation__.Notify(115772)

	if len(r.mu.proposals) > 0 || func() bool {
		__antithesis_instrumentation__.Notify(115785)
		return r.mu.applyingEntries == true
	}() == true {
		__antithesis_instrumentation__.Notify(115786)
		res.FailReason = sidetransport.ProposalsInFlight
		return res
	} else {
		__antithesis_instrumentation__.Notify(115787)
	}
	__antithesis_instrumentation__.Notify(115773)

	if !r.mu.proposalBuf.MaybeForwardClosedLocked(ctx, target) {
		__antithesis_instrumentation__.Notify(115788)
		res.FailReason = sidetransport.RequestsEvaluatingBelowTarget
		return res
	} else {
		__antithesis_instrumentation__.Notify(115789)
	}
	__antithesis_instrumentation__.Notify(115774)

	const knownApplied = true
	r.sideTransportClosedTimestamp.forward(ctx, target, lai, knownApplied)
	res.OK = true
	res.LAI = lai
	res.Policy = policy
	return res
}

func (r *Replica) closedTimestampTargetRLocked() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(115790)
	return closedts.TargetForPolicy(
		r.Clock().NowAsClockTimestamp(),
		r.Clock().MaxOffset(),
		closedts.TargetDuration.Get(&r.ClusterSettings().SV),
		closedts.LeadForGlobalReadsOverride.Get(&r.ClusterSettings().SV),
		closedts.SideTransportCloseInterval.Get(&r.ClusterSettings().SV),
		r.closedTimestampPolicyRLocked(),
	)
}

func (r *Replica) ForwardSideTransportClosedTimestamp(
	ctx context.Context, closed hlc.Timestamp, lai ctpb.LAI,
) {
	__antithesis_instrumentation__.Notify(115791)

	const knownApplied = false
	r.sideTransportClosedTimestamp.forward(ctx, closed, lai, knownApplied)
}

type sidetransportAccess struct {
	rangeID  roachpb.RangeID
	receiver sidetransportReceiver
	mu       struct {
		syncutil.RWMutex

		cur closedTimestamp

		next closedTimestamp
	}
}

type sidetransportReceiver interface {
	GetClosedTimestamp(
		ctx context.Context, rangeID roachpb.RangeID, leaseholderNode roachpb.NodeID,
	) (hlc.Timestamp, ctpb.LAI)
	HTML() string
}

type closedTimestamp struct {
	ts  hlc.Timestamp
	lai ctpb.LAI
}

func (a closedTimestamp) regression(b closedTimestamp) bool {
	__antithesis_instrumentation__.Notify(115792)
	if a.lai == b.lai {
		__antithesis_instrumentation__.Notify(115795)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115796)
	}
	__antithesis_instrumentation__.Notify(115793)
	if a.lai < b.lai {
		__antithesis_instrumentation__.Notify(115797)
		return b.ts.Less(a.ts)
	} else {
		__antithesis_instrumentation__.Notify(115798)
	}
	__antithesis_instrumentation__.Notify(115794)
	return a.ts.Less(b.ts)
}

func (a *closedTimestamp) merge(b closedTimestamp) {
	__antithesis_instrumentation__.Notify(115799)
	if a.lai == b.lai {
		__antithesis_instrumentation__.Notify(115800)
		a.ts.Forward(b.ts)
	} else {
		__antithesis_instrumentation__.Notify(115801)
		if a.lai < b.lai {
			__antithesis_instrumentation__.Notify(115802)
			a.ts = b.ts
			a.lai = b.lai
		} else {
			__antithesis_instrumentation__.Notify(115803)
		}
	}
}

func (st *sidetransportAccess) init(receiver sidetransportReceiver, rangeID roachpb.RangeID) {
	if receiver != nil {

		st.receiver = receiver
	}
	st.rangeID = rangeID
}

func (st *sidetransportAccess) forward(
	ctx context.Context, closed hlc.Timestamp, lai ctpb.LAI, knownApplied bool,
) closedTimestamp {
	__antithesis_instrumentation__.Notify(115804)
	st.mu.Lock()
	defer st.mu.Unlock()

	up := closedTimestamp{ts: closed, lai: lai}
	if st.mu.cur.lai != 0 {
		__antithesis_instrumentation__.Notify(115809)
		st.assertNoRegression(ctx, st.mu.cur, up)
	} else {
		__antithesis_instrumentation__.Notify(115810)
	}
	__antithesis_instrumentation__.Notify(115805)
	if st.mu.next.lai != 0 {
		__antithesis_instrumentation__.Notify(115811)
		st.assertNoRegression(ctx, st.mu.next, up)
	} else {
		__antithesis_instrumentation__.Notify(115812)
	}
	__antithesis_instrumentation__.Notify(115806)

	if up.lai <= st.mu.cur.lai {
		__antithesis_instrumentation__.Notify(115813)

		knownApplied = true
	} else {
		__antithesis_instrumentation__.Notify(115814)
	}
	__antithesis_instrumentation__.Notify(115807)
	if knownApplied {
		__antithesis_instrumentation__.Notify(115815)

		if up.lai >= st.mu.next.lai {
			__antithesis_instrumentation__.Notify(115817)
			up.merge(st.mu.next)
			st.mu.next = closedTimestamp{}
		} else {
			__antithesis_instrumentation__.Notify(115818)
		}
		__antithesis_instrumentation__.Notify(115816)
		st.mu.cur.merge(up)
	} else {
		__antithesis_instrumentation__.Notify(115819)

		st.mu.next.merge(up)
	}
	__antithesis_instrumentation__.Notify(115808)
	return st.mu.cur
}

func (st *sidetransportAccess) assertNoRegression(ctx context.Context, cur, up closedTimestamp) {
	__antithesis_instrumentation__.Notify(115820)
	if cur.regression(up) {
		__antithesis_instrumentation__.Notify(115821)
		log.Fatalf(ctx, "side-transport update saw closed timestamp regression on r%d: "+
			"(lai=%d, ts=%s) -> (lai=%d, ts=%s)", st.rangeID, cur.lai, cur.ts, up.lai, up.ts)
	} else {
		__antithesis_instrumentation__.Notify(115822)
	}
}

func (st *sidetransportAccess) get(
	ctx context.Context, leaseholder roachpb.NodeID, appliedLAI ctpb.LAI, sufficient hlc.Timestamp,
) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(115823)
	st.mu.RLock()
	cur, next := st.mu.cur, st.mu.next
	st.mu.RUnlock()

	if !sufficient.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(115828)
		return sufficient.LessEq(cur.ts) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115829)
		return cur.ts
	} else {
		__antithesis_instrumentation__.Notify(115830)
	}
	__antithesis_instrumentation__.Notify(115824)

	if next.lai != 0 {
		__antithesis_instrumentation__.Notify(115831)
		if next.lai > appliedLAI {
			__antithesis_instrumentation__.Notify(115833)

			return cur.ts
		} else {
			__antithesis_instrumentation__.Notify(115834)
		}
		__antithesis_instrumentation__.Notify(115832)

		cur = st.forward(ctx, next.ts, next.lai, true)
		next = closedTimestamp{}

		if !sufficient.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(115835)
			return sufficient.LessEq(cur.ts) == true
		}() == true {
			__antithesis_instrumentation__.Notify(115836)
			return cur.ts
		} else {
			__antithesis_instrumentation__.Notify(115837)
		}
	} else {
		__antithesis_instrumentation__.Notify(115838)
	}
	__antithesis_instrumentation__.Notify(115825)

	if st.receiver == nil {
		__antithesis_instrumentation__.Notify(115839)
		return cur.ts
	} else {
		__antithesis_instrumentation__.Notify(115840)
	}
	__antithesis_instrumentation__.Notify(115826)

	recTS, recLAI := st.receiver.GetClosedTimestamp(ctx, st.rangeID, leaseholder)
	if recTS.LessEq(cur.ts) {
		__antithesis_instrumentation__.Notify(115841)

		return cur.ts
	} else {
		__antithesis_instrumentation__.Notify(115842)
	}
	__antithesis_instrumentation__.Notify(115827)

	knownApplied := recLAI <= appliedLAI
	cur = st.forward(ctx, recTS, recLAI, knownApplied)
	return cur.ts
}
