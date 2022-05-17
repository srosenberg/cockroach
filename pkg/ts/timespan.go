package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type QueryTimespan struct {
	StartNanos          int64
	EndNanos            int64
	NowNanos            int64
	SampleDurationNanos int64
}

func (qt *QueryTimespan) width() int64 {
	__antithesis_instrumentation__.Notify(648870)
	return qt.EndNanos - qt.StartNanos
}

func (qt *QueryTimespan) moveForward(forwardNanos int64) {
	__antithesis_instrumentation__.Notify(648871)
	qt.StartNanos += forwardNanos
	qt.EndNanos += forwardNanos
}

func (qt *QueryTimespan) expand(size int64) {
	__antithesis_instrumentation__.Notify(648872)
	qt.StartNanos -= size
	qt.EndNanos += size
}

func (qt *QueryTimespan) normalize() {
	__antithesis_instrumentation__.Notify(648873)
	qt.StartNanos -= qt.StartNanos % qt.SampleDurationNanos
	qt.EndNanos -= qt.EndNanos % qt.SampleDurationNanos
}

func (qt *QueryTimespan) verifyBounds() error {
	__antithesis_instrumentation__.Notify(648874)
	if qt.StartNanos > qt.EndNanos {
		__antithesis_instrumentation__.Notify(648876)
		return fmt.Errorf("startNanos %d was later than endNanos %d", qt.StartNanos, qt.EndNanos)
	} else {
		__antithesis_instrumentation__.Notify(648877)
	}
	__antithesis_instrumentation__.Notify(648875)
	return nil
}

func (qt *QueryTimespan) verifyDiskResolution(diskResolution Resolution) error {
	__antithesis_instrumentation__.Notify(648878)
	resolutionSampleDuration := diskResolution.SampleDuration()

	if qt.SampleDurationNanos < resolutionSampleDuration {
		__antithesis_instrumentation__.Notify(648881)
		return fmt.Errorf(
			"sampleDuration %d was not less that queryResolution.SampleDuration %d",
			qt.SampleDurationNanos,
			resolutionSampleDuration,
		)
	} else {
		__antithesis_instrumentation__.Notify(648882)
	}
	__antithesis_instrumentation__.Notify(648879)
	if qt.SampleDurationNanos%resolutionSampleDuration != 0 {
		__antithesis_instrumentation__.Notify(648883)
		return fmt.Errorf(
			"sampleDuration %d is not a multiple of queryResolution.SampleDuration %d",
			qt.SampleDurationNanos,
			resolutionSampleDuration,
		)
	} else {
		__antithesis_instrumentation__.Notify(648884)
	}
	__antithesis_instrumentation__.Notify(648880)
	return nil
}

func (qt *QueryTimespan) adjustForCurrentTime(diskResolution Resolution) error {
	__antithesis_instrumentation__.Notify(648885)

	cutoff := qt.NowNanos - qt.SampleDurationNanos

	if qt.StartNanos > cutoff {
		__antithesis_instrumentation__.Notify(648888)
		return fmt.Errorf(
			"cannot query time series in the future (start time %s was greater than "+
				"cutoff for current sample period %s); current time: %s; sample duration: %s",
			timeutil.Unix(0, qt.StartNanos),
			timeutil.Unix(0, cutoff),
			timeutil.Unix(0, qt.NowNanos),
			time.Duration(qt.SampleDurationNanos),
		)
	} else {
		__antithesis_instrumentation__.Notify(648889)
	}
	__antithesis_instrumentation__.Notify(648886)
	if qt.EndNanos > cutoff {
		__antithesis_instrumentation__.Notify(648890)
		qt.EndNanos = cutoff
	} else {
		__antithesis_instrumentation__.Notify(648891)
	}
	__antithesis_instrumentation__.Notify(648887)

	return nil
}
