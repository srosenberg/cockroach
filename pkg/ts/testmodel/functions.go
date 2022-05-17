package testmodel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
)

type aggFunc func(DataSeries) float64
type fillFunc func(DataSeries, DataSeries, int64) DataSeries

func AggregateSum(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648832)
	total := 0.0
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648834)
		total += dp.Value
	}
	__antithesis_instrumentation__.Notify(648833)
	return total
}

func AggregateAverage(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648835)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(648837)
		return 0.0
	} else {
		__antithesis_instrumentation__.Notify(648838)
	}
	__antithesis_instrumentation__.Notify(648836)
	return AggregateSum(data) / float64(len(data))
}

func AggregateMax(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648839)
	max := -math.MaxFloat64
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648841)
		if dp.Value > max {
			__antithesis_instrumentation__.Notify(648842)
			max = dp.Value
		} else {
			__antithesis_instrumentation__.Notify(648843)
		}
	}
	__antithesis_instrumentation__.Notify(648840)
	return max
}

func AggregateMin(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648844)
	min := math.MaxFloat64
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648846)
		if dp.Value < min {
			__antithesis_instrumentation__.Notify(648847)
			min = dp.Value
		} else {
			__antithesis_instrumentation__.Notify(648848)
		}
	}
	__antithesis_instrumentation__.Notify(648845)
	return min
}

func AggregateFirst(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648849)
	return data[0].Value
}

func AggregateLast(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648850)
	return data[len(data)-1].Value
}

func AggregateVariance(data DataSeries) float64 {
	__antithesis_instrumentation__.Notify(648851)
	mean := 0.0
	meanSquaredDist := 0.0
	if len(data) < 2 {
		__antithesis_instrumentation__.Notify(648854)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(648855)
	}
	__antithesis_instrumentation__.Notify(648852)
	for i, dp := range data {
		__antithesis_instrumentation__.Notify(648856)

		delta := dp.Value - mean
		mean += delta / float64(i+1)
		delta2 := dp.Value - mean
		meanSquaredDist += delta * delta2
	}
	__antithesis_instrumentation__.Notify(648853)
	return meanSquaredDist / float64(len(data))
}

func getAggFunction(agg tspb.TimeSeriesQueryAggregator) aggFunc {
	__antithesis_instrumentation__.Notify(648857)
	switch agg {
	case tspb.TimeSeriesQueryAggregator_AVG:
		__antithesis_instrumentation__.Notify(648859)
		return AggregateAverage
	case tspb.TimeSeriesQueryAggregator_SUM:
		__antithesis_instrumentation__.Notify(648860)
		return AggregateSum
	case tspb.TimeSeriesQueryAggregator_MAX:
		__antithesis_instrumentation__.Notify(648861)
		return AggregateMax
	case tspb.TimeSeriesQueryAggregator_MIN:
		__antithesis_instrumentation__.Notify(648862)
		return AggregateMin
	case tspb.TimeSeriesQueryAggregator_FIRST:
		__antithesis_instrumentation__.Notify(648863)
		return AggregateFirst
	case tspb.TimeSeriesQueryAggregator_LAST:
		__antithesis_instrumentation__.Notify(648864)
		return AggregateLast
	case tspb.TimeSeriesQueryAggregator_VARIANCE:
		__antithesis_instrumentation__.Notify(648865)
		return AggregateVariance
	default:
		__antithesis_instrumentation__.Notify(648866)
	}
	__antithesis_instrumentation__.Notify(648858)

	panic(fmt.Sprintf("unknown aggregator option specified: %v", agg))
}

func fillFuncLinearInterpolate(before DataSeries, after DataSeries, resolution int64) DataSeries {
	__antithesis_instrumentation__.Notify(648867)
	start := before[len(before)-1]
	end := after[0]

	step := (end.Value - start.Value) / float64(end.TimestampNanos-start.TimestampNanos)

	result := make(DataSeries, (end.TimestampNanos-start.TimestampNanos)/resolution-1)
	for i := range result {
		__antithesis_instrumentation__.Notify(648869)
		result[i] = dp(
			start.TimestampNanos+(resolution*int64(i+1)),
			start.Value+(step*float64(i+1)*float64(resolution)),
		)
	}
	__antithesis_instrumentation__.Notify(648868)
	return result
}
