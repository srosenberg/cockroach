package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/errors"
)

func TimeFamilyPrecisionToRoundDuration(precision int32) time.Duration {
	__antithesis_instrumentation__.Notify(614551)
	switch precision {
	case 0:
		__antithesis_instrumentation__.Notify(614553)
		return time.Second
	case 1:
		__antithesis_instrumentation__.Notify(614554)
		return time.Millisecond * 100
	case 2:
		__antithesis_instrumentation__.Notify(614555)
		return time.Millisecond * 10
	case 3:
		__antithesis_instrumentation__.Notify(614556)
		return time.Millisecond
	case 4:
		__antithesis_instrumentation__.Notify(614557)
		return time.Microsecond * 100
	case 5:
		__antithesis_instrumentation__.Notify(614558)
		return time.Microsecond * 10
	case 6:
		__antithesis_instrumentation__.Notify(614559)
		return time.Microsecond
	default:
		__antithesis_instrumentation__.Notify(614560)
	}
	__antithesis_instrumentation__.Notify(614552)
	panic(errors.Newf("unsupported precision: %d", precision))
}
