package closedts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

func TargetForPolicy(
	now hlc.ClockTimestamp,
	maxClockOffset time.Duration,
	lagTargetDuration time.Duration,
	leadTargetOverride time.Duration,
	sideTransportCloseInterval time.Duration,
	policy roachpb.RangeClosedTimestampPolicy,
) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(98561)
	var res hlc.Timestamp
	switch policy {
	case roachpb.LAG_BY_CLUSTER_SETTING:
		__antithesis_instrumentation__.Notify(98563)

		res = now.ToTimestamp().Add(-lagTargetDuration.Nanoseconds(), 0)
	case roachpb.LEAD_FOR_GLOBAL_READS:
		__antithesis_instrumentation__.Notify(98564)

		const maxNetworkRTT = 150 * time.Millisecond

		const raftTransportOverhead = 20 * time.Millisecond
		raftTransportPropTime := (maxNetworkRTT*3)/2 + raftTransportOverhead

		sideTransportPropTime := maxNetworkRTT/2 + sideTransportCloseInterval

		maxTransportPropTime := sideTransportPropTime
		if maxTransportPropTime < raftTransportPropTime {
			__antithesis_instrumentation__.Notify(98568)
			maxTransportPropTime = raftTransportPropTime
		} else {
			__antithesis_instrumentation__.Notify(98569)
		}
		__antithesis_instrumentation__.Notify(98565)

		const bufferTime = 25 * time.Millisecond
		leadTimeAtSender := maxTransportPropTime + maxClockOffset + bufferTime

		if leadTargetOverride != 0 {
			__antithesis_instrumentation__.Notify(98570)
			leadTimeAtSender = leadTargetOverride
		} else {
			__antithesis_instrumentation__.Notify(98571)
		}
		__antithesis_instrumentation__.Notify(98566)

		res = now.ToTimestamp().Add(leadTimeAtSender.Nanoseconds(), 0).WithSynthetic(true)
	default:
		__antithesis_instrumentation__.Notify(98567)
		panic("unexpected RangeClosedTimestampPolicy")
	}
	__antithesis_instrumentation__.Notify(98562)

	res.Logical = 0
	return res
}
