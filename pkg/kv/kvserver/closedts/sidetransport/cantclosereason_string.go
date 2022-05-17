package sidetransport

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(98572)

	var x [1]struct{}
	_ = x[ReasonUnknown-0]
	_ = x[ReplicaDestroyed-1]
	_ = x[InvalidLease-2]
	_ = x[TargetOverLeaseExpiration-3]
	_ = x[MergeInProgress-4]
	_ = x[ProposalsInFlight-5]
	_ = x[RequestsEvaluatingBelowTarget-6]
	_ = x[MaxReason-7]
}

const _CantCloseReason_name = "ReasonUnknownReplicaDestroyedInvalidLeaseTargetOverLeaseExpirationMergeInProgressProposalsInFlightRequestsEvaluatingBelowTargetMaxReason"

var _CantCloseReason_index = [...]uint8{0, 13, 29, 41, 66, 81, 98, 127, 136}

func (i CantCloseReason) String() string {
	__antithesis_instrumentation__.Notify(98573)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(98575)
		return i >= CantCloseReason(len(_CantCloseReason_index)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(98576)
		return "CantCloseReason(" + strconv.FormatInt(int64(i), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(98577)
	}
	__antithesis_instrumentation__.Notify(98574)
	return _CantCloseReason_name[_CantCloseReason_index[i]:_CantCloseReason_index[i+1]]
}
