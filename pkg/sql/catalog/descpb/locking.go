package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/errors"
)

func (s ScanLockingStrength) PrettyString() string {
	__antithesis_instrumentation__.Notify(252492)
	switch s {
	case ScanLockingStrength_FOR_NONE:
		__antithesis_instrumentation__.Notify(252493)
		return "for none"
	case ScanLockingStrength_FOR_KEY_SHARE:
		__antithesis_instrumentation__.Notify(252494)
		return "for key share"
	case ScanLockingStrength_FOR_SHARE:
		__antithesis_instrumentation__.Notify(252495)
		return "for share"
	case ScanLockingStrength_FOR_NO_KEY_UPDATE:
		__antithesis_instrumentation__.Notify(252496)
		return "for no key update"
	case ScanLockingStrength_FOR_UPDATE:
		__antithesis_instrumentation__.Notify(252497)
		return "for update"
	default:
		__antithesis_instrumentation__.Notify(252498)
		panic(errors.AssertionFailedf("unexpected strength"))
	}
}

func ToScanLockingStrength(s tree.LockingStrength) ScanLockingStrength {
	__antithesis_instrumentation__.Notify(252499)
	switch s {
	case tree.ForNone:
		__antithesis_instrumentation__.Notify(252500)
		return ScanLockingStrength_FOR_NONE
	case tree.ForKeyShare:
		__antithesis_instrumentation__.Notify(252501)
		return ScanLockingStrength_FOR_KEY_SHARE
	case tree.ForShare:
		__antithesis_instrumentation__.Notify(252502)
		return ScanLockingStrength_FOR_SHARE
	case tree.ForNoKeyUpdate:
		__antithesis_instrumentation__.Notify(252503)
		return ScanLockingStrength_FOR_NO_KEY_UPDATE
	case tree.ForUpdate:
		__antithesis_instrumentation__.Notify(252504)
		return ScanLockingStrength_FOR_UPDATE
	default:
		__antithesis_instrumentation__.Notify(252505)
		panic(errors.AssertionFailedf("unknown locking strength %s", s))
	}
}

func (wp ScanLockingWaitPolicy) PrettyString() string {
	__antithesis_instrumentation__.Notify(252506)
	switch wp {
	case ScanLockingWaitPolicy_BLOCK:
		__antithesis_instrumentation__.Notify(252507)
		return "block"
	case ScanLockingWaitPolicy_SKIP:
		__antithesis_instrumentation__.Notify(252508)
		return "skip locked"
	case ScanLockingWaitPolicy_ERROR:
		__antithesis_instrumentation__.Notify(252509)
		return "nowait"
	default:
		__antithesis_instrumentation__.Notify(252510)
		panic(errors.AssertionFailedf("unexpected wait policy"))
	}
}

func ToScanLockingWaitPolicy(wp tree.LockingWaitPolicy) ScanLockingWaitPolicy {
	__antithesis_instrumentation__.Notify(252511)
	switch wp {
	case tree.LockWaitBlock:
		__antithesis_instrumentation__.Notify(252512)
		return ScanLockingWaitPolicy_BLOCK
	case tree.LockWaitSkip:
		__antithesis_instrumentation__.Notify(252513)
		return ScanLockingWaitPolicy_SKIP
	case tree.LockWaitError:
		__antithesis_instrumentation__.Notify(252514)
		return ScanLockingWaitPolicy_ERROR
	default:
		__antithesis_instrumentation__.Notify(252515)
		panic(errors.AssertionFailedf("unknown locking wait policy %s", wp))
	}
}
