package tspb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

func (ts TimeSeriesData) ToInternal(
	keyDuration, sampleDuration int64, columnar bool,
) ([]roachpb.InternalTimeSeriesData, error) {
	__antithesis_instrumentation__.Notify(648892)
	if err := VerifySlabAndSampleDuration(keyDuration, sampleDuration); err != nil {
		__antithesis_instrumentation__.Notify(648895)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(648896)
	}
	__antithesis_instrumentation__.Notify(648893)

	result := make([]roachpb.InternalTimeSeriesData, 0, len(ts.Datapoints))

	resultByKeyTime := make(map[int64]*roachpb.InternalTimeSeriesData)

	for _, dp := range ts.Datapoints {
		__antithesis_instrumentation__.Notify(648897)

		keyTime := (dp.TimestampNanos / keyDuration) * keyDuration
		itsd, ok := resultByKeyTime[keyTime]
		if !ok {
			__antithesis_instrumentation__.Notify(648899)
			result = append(result, roachpb.InternalTimeSeriesData{
				StartTimestampNanos: keyTime,
				SampleDurationNanos: sampleDuration,
			})
			itsd = &result[len(result)-1]
			resultByKeyTime[keyTime] = itsd
		} else {
			__antithesis_instrumentation__.Notify(648900)
		}
		__antithesis_instrumentation__.Notify(648898)

		if columnar {
			__antithesis_instrumentation__.Notify(648901)
			itsd.Offset = append(itsd.Offset, itsd.OffsetForTimestamp(dp.TimestampNanos))
			itsd.Last = append(itsd.Last, dp.Value)
		} else {
			__antithesis_instrumentation__.Notify(648902)
			itsd.Samples = append(itsd.Samples, roachpb.InternalTimeSeriesSample{
				Offset: itsd.OffsetForTimestamp(dp.TimestampNanos),
				Count:  1,
				Sum:    dp.Value,
			})
		}
	}
	__antithesis_instrumentation__.Notify(648894)

	return result, nil
}

func VerifySlabAndSampleDuration(slabDuration, sampleDuration int64) error {
	__antithesis_instrumentation__.Notify(648903)
	if slabDuration%sampleDuration != 0 {
		__antithesis_instrumentation__.Notify(648906)
		return fmt.Errorf(
			"sample duration %d does not evenly divide key duration %d",
			sampleDuration, slabDuration)
	} else {
		__antithesis_instrumentation__.Notify(648907)
	}
	__antithesis_instrumentation__.Notify(648904)
	if slabDuration < sampleDuration {
		__antithesis_instrumentation__.Notify(648908)
		return fmt.Errorf(
			"sample duration %d is not less than or equal to key duration %d",
			sampleDuration, slabDuration)
	} else {
		__antithesis_instrumentation__.Notify(648909)
	}
	__antithesis_instrumentation__.Notify(648905)

	return nil
}
