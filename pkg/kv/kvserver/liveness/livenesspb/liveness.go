package livenesspb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (l *Liveness) IsLive(now time.Time) bool {
	__antithesis_instrumentation__.Notify(108347)
	expiration := timeutil.Unix(0, l.Expiration.WallTime)
	return now.Before(expiration)
}

func (l *Liveness) IsDead(now time.Time, threshold time.Duration) bool {
	__antithesis_instrumentation__.Notify(108348)
	expiration := timeutil.Unix(0, l.Expiration.WallTime)
	deadAsOf := expiration.Add(threshold)
	return !now.Before(deadAsOf)
}

func (l *Liveness) Compare(o Liveness) int {
	__antithesis_instrumentation__.Notify(108349)

	if l.Epoch != o.Epoch {
		__antithesis_instrumentation__.Notify(108352)
		if l.Epoch < o.Epoch {
			__antithesis_instrumentation__.Notify(108354)
			return -1
		} else {
			__antithesis_instrumentation__.Notify(108355)
		}
		__antithesis_instrumentation__.Notify(108353)
		return +1
	} else {
		__antithesis_instrumentation__.Notify(108356)
	}
	__antithesis_instrumentation__.Notify(108350)
	if !l.Expiration.EqOrdering(o.Expiration) {
		__antithesis_instrumentation__.Notify(108357)
		if l.Expiration.Less(o.Expiration) {
			__antithesis_instrumentation__.Notify(108359)
			return -1
		} else {
			__antithesis_instrumentation__.Notify(108360)
		}
		__antithesis_instrumentation__.Notify(108358)
		return +1
	} else {
		__antithesis_instrumentation__.Notify(108361)
	}
	__antithesis_instrumentation__.Notify(108351)
	return 0
}

func (l Liveness) String() string {
	__antithesis_instrumentation__.Notify(108362)
	var extra string
	if l.Draining || func() bool {
		__antithesis_instrumentation__.Notify(108364)
		return l.Membership.Decommissioning() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(108365)
		return l.Membership.Decommissioned() == true
	}() == true {
		__antithesis_instrumentation__.Notify(108366)
		extra = fmt.Sprintf(" drain:%t membership:%s", l.Draining, l.Membership.String())
	} else {
		__antithesis_instrumentation__.Notify(108367)
	}
	__antithesis_instrumentation__.Notify(108363)
	return fmt.Sprintf("liveness(nid:%d epo:%d exp:%s%s)", l.NodeID, l.Epoch, l.Expiration, extra)
}

func (c MembershipStatus) Decommissioning() bool {
	__antithesis_instrumentation__.Notify(108368)
	return c == MembershipStatus_DECOMMISSIONING
}

func (c MembershipStatus) Decommissioned() bool {
	__antithesis_instrumentation__.Notify(108369)
	return c == MembershipStatus_DECOMMISSIONED
}

func (c MembershipStatus) Active() bool {
	__antithesis_instrumentation__.Notify(108370)
	return c == MembershipStatus_ACTIVE
}

func (c MembershipStatus) String() string {
	__antithesis_instrumentation__.Notify(108371)

	switch c {
	case MembershipStatus_ACTIVE:
		__antithesis_instrumentation__.Notify(108372)
		return "active"
	case MembershipStatus_DECOMMISSIONING:
		__antithesis_instrumentation__.Notify(108373)
		return "decommissioning"
	case MembershipStatus_DECOMMISSIONED:
		__antithesis_instrumentation__.Notify(108374)
		return "decommissioned"
	default:
		__antithesis_instrumentation__.Notify(108375)
		err := "unknown membership status, expected one of [active,decommissioning,decommissioned]"
		panic(err)
	}
}

func ValidateTransition(old, new Liveness) error {
	__antithesis_instrumentation__.Notify(108376)
	if old.Membership == new.Membership {
		__antithesis_instrumentation__.Notify(108381)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(108382)
	}
	__antithesis_instrumentation__.Notify(108377)

	if old.Membership.Decommissioned() && func() bool {
		__antithesis_instrumentation__.Notify(108383)
		return new.Membership.Decommissioning() == true
	}() == true {
		__antithesis_instrumentation__.Notify(108384)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(108385)
	}
	__antithesis_instrumentation__.Notify(108378)

	if new.Membership.Active() && func() bool {
		__antithesis_instrumentation__.Notify(108386)
		return !old.Membership.Decommissioning() == true
	}() == true {
		__antithesis_instrumentation__.Notify(108387)
		err := fmt.Sprintf("can only recommission a decommissioning node; n%d found to be %s",
			new.NodeID, old.Membership.String())
		return status.Error(codes.FailedPrecondition, err)
	} else {
		__antithesis_instrumentation__.Notify(108388)
	}
	__antithesis_instrumentation__.Notify(108379)

	if new.Membership.Decommissioned() && func() bool {
		__antithesis_instrumentation__.Notify(108389)
		return !old.Membership.Decommissioning() == true
	}() == true {
		__antithesis_instrumentation__.Notify(108390)
		err := fmt.Sprintf("can only fully decommission an already decommissioning node; n%d found to be %s",
			new.NodeID, old.Membership.String())
		return status.Error(codes.FailedPrecondition, err)
	} else {
		__antithesis_instrumentation__.Notify(108391)
	}
	__antithesis_instrumentation__.Notify(108380)

	return nil
}
