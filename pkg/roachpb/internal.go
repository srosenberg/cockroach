package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (data *InternalTimeSeriesData) IsColumnar() bool {
	__antithesis_instrumentation__.Notify(170913)
	return len(data.Offset) > 0
}

func (data *InternalTimeSeriesData) IsRollup() bool {
	__antithesis_instrumentation__.Notify(170914)
	return len(data.Count) > 0
}

func (data *InternalTimeSeriesData) SampleCount() int {
	__antithesis_instrumentation__.Notify(170915)
	if data.IsColumnar() {
		__antithesis_instrumentation__.Notify(170917)
		return len(data.Offset)
	} else {
		__antithesis_instrumentation__.Notify(170918)
	}
	__antithesis_instrumentation__.Notify(170916)
	return len(data.Samples)
}

func (data *InternalTimeSeriesData) OffsetForTimestamp(timestampNanos int64) int32 {
	__antithesis_instrumentation__.Notify(170919)
	return int32((timestampNanos - data.StartTimestampNanos) / data.SampleDurationNanos)
}

func (data *InternalTimeSeriesData) TimestampForOffset(offset int32) int64 {
	__antithesis_instrumentation__.Notify(170920)
	return data.StartTimestampNanos + int64(offset)*data.SampleDurationNanos
}
