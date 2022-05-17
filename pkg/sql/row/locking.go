package row

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

func getKeyLockingStrength(lockStrength descpb.ScanLockingStrength) lock.Strength {
	__antithesis_instrumentation__.Notify(568470)
	switch lockStrength {
	case descpb.ScanLockingStrength_FOR_NONE:
		__antithesis_instrumentation__.Notify(568471)
		return lock.None

	case descpb.ScanLockingStrength_FOR_KEY_SHARE:
		__antithesis_instrumentation__.Notify(568472)

		fallthrough
	case descpb.ScanLockingStrength_FOR_SHARE:
		__antithesis_instrumentation__.Notify(568473)

		return lock.None

	case descpb.ScanLockingStrength_FOR_NO_KEY_UPDATE:
		__antithesis_instrumentation__.Notify(568474)

		fallthrough
	case descpb.ScanLockingStrength_FOR_UPDATE:
		__antithesis_instrumentation__.Notify(568475)

		return lock.Exclusive

	default:
		__antithesis_instrumentation__.Notify(568476)
		panic(errors.AssertionFailedf("unknown locking strength %s", lockStrength))
	}
}

func GetWaitPolicy(lockWaitPolicy descpb.ScanLockingWaitPolicy) lock.WaitPolicy {
	__antithesis_instrumentation__.Notify(568477)
	switch lockWaitPolicy {
	case descpb.ScanLockingWaitPolicy_BLOCK:
		__antithesis_instrumentation__.Notify(568478)
		return lock.WaitPolicy_Block

	case descpb.ScanLockingWaitPolicy_SKIP:
		__antithesis_instrumentation__.Notify(568479)

		panic(errors.AssertionFailedf("unsupported wait policy %s", lockWaitPolicy))

	case descpb.ScanLockingWaitPolicy_ERROR:
		__antithesis_instrumentation__.Notify(568480)
		return lock.WaitPolicy_Error

	default:
		__antithesis_instrumentation__.Notify(568481)
		panic(errors.AssertionFailedf("unknown wait policy %s", lockWaitPolicy))
	}
}
