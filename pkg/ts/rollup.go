package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"sort"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type rollupDatapoint struct {
	timestampNanos int64
	first          float64
	last           float64
	min            float64
	max            float64
	sum            float64
	count          uint32
	variance       float64
}

type rollupData struct {
	name       string
	source     string
	datapoints []rollupDatapoint
}

func (rd *rollupData) toInternal(
	keyDuration, sampleDuration int64,
) ([]roachpb.InternalTimeSeriesData, error) {
	__antithesis_instrumentation__.Notify(648520)
	if err := tspb.VerifySlabAndSampleDuration(keyDuration, sampleDuration); err != nil {
		__antithesis_instrumentation__.Notify(648523)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(648524)
	}
	__antithesis_instrumentation__.Notify(648521)

	result := make([]roachpb.InternalTimeSeriesData, 0, len(rd.datapoints))

	resultByKeyTime := make(map[int64]*roachpb.InternalTimeSeriesData)

	for _, dp := range rd.datapoints {
		__antithesis_instrumentation__.Notify(648525)

		keyTime := normalizeToPeriod(dp.timestampNanos, keyDuration)
		itsd, ok := resultByKeyTime[keyTime]
		if !ok {
			__antithesis_instrumentation__.Notify(648527)
			result = append(result, roachpb.InternalTimeSeriesData{
				StartTimestampNanos: keyTime,
				SampleDurationNanos: sampleDuration,
			})
			itsd = &result[len(result)-1]
			resultByKeyTime[keyTime] = itsd
		} else {
			__antithesis_instrumentation__.Notify(648528)
		}
		__antithesis_instrumentation__.Notify(648526)

		itsd.Offset = append(itsd.Offset, itsd.OffsetForTimestamp(dp.timestampNanos))
		itsd.Last = append(itsd.Last, dp.last)
		itsd.First = append(itsd.First, dp.first)
		itsd.Min = append(itsd.Min, dp.min)
		itsd.Max = append(itsd.Max, dp.max)
		itsd.Count = append(itsd.Count, dp.count)
		itsd.Sum = append(itsd.Sum, dp.sum)
		itsd.Variance = append(itsd.Variance, dp.variance)
	}
	__antithesis_instrumentation__.Notify(648522)

	return result, nil
}

func computeRollupsFromData(data tspb.TimeSeriesData, rollupPeriodNanos int64) rollupData {
	__antithesis_instrumentation__.Notify(648529)
	rollup := rollupData{
		name:   data.Name,
		source: data.Source,
	}

	createRollupPoint := func(timestamp int64, dataSlice []tspb.TimeSeriesDatapoint) {
		__antithesis_instrumentation__.Notify(648532)
		result := rollupDatapoint{
			timestampNanos: timestamp,
			max:            -math.MaxFloat64,
			min:            math.MaxFloat64,
		}
		for i, dp := range dataSlice {
			__antithesis_instrumentation__.Notify(648534)
			if i == 0 {
				__antithesis_instrumentation__.Notify(648537)
				result.first = dp.Value
			} else {
				__antithesis_instrumentation__.Notify(648538)
			}
			__antithesis_instrumentation__.Notify(648535)
			result.last = dp.Value
			result.max = math.Max(result.max, dp.Value)
			result.min = math.Min(result.min, dp.Value)

			if result.count > 0 {
				__antithesis_instrumentation__.Notify(648539)
				result.variance = computeParallelVariance(
					parallelVarianceArgs{
						count:    result.count,
						average:  result.sum / float64(result.count),
						variance: result.variance,
					},
					parallelVarianceArgs{
						count:    1,
						average:  dp.Value,
						variance: 0,
					},
				)
			} else {
				__antithesis_instrumentation__.Notify(648540)
			}
			__antithesis_instrumentation__.Notify(648536)

			result.count++
			result.sum += dp.Value
		}
		__antithesis_instrumentation__.Notify(648533)

		rollup.datapoints = append(rollup.datapoints, result)
	}
	__antithesis_instrumentation__.Notify(648530)

	dps := data.Datapoints
	for len(dps) > 0 {
		__antithesis_instrumentation__.Notify(648541)
		rollupTimestamp := normalizeToPeriod(dps[0].TimestampNanos, rollupPeriodNanos)
		endIdx := sort.Search(len(dps), func(i int) bool {
			__antithesis_instrumentation__.Notify(648543)
			return normalizeToPeriod(dps[i].TimestampNanos, rollupPeriodNanos) > rollupTimestamp
		})
		__antithesis_instrumentation__.Notify(648542)
		createRollupPoint(rollupTimestamp, dps[:endIdx])
		dps = dps[endIdx:]
	}
	__antithesis_instrumentation__.Notify(648531)

	return rollup
}

func (db *DB) rollupTimeSeries(
	ctx context.Context,
	timeSeriesList []timeSeriesResolutionInfo,
	now hlc.Timestamp,
	qmc QueryMemoryContext,
) error {
	__antithesis_instrumentation__.Notify(648544)
	thresholds := db.computeThresholds(now.WallTime)
	for _, timeSeries := range timeSeriesList {
		__antithesis_instrumentation__.Notify(648546)

		targetResolution, hasRollup := timeSeries.Resolution.TargetRollupResolution()
		if !hasRollup {
			__antithesis_instrumentation__.Notify(648550)
			continue
		} else {
			__antithesis_instrumentation__.Notify(648551)
		}
		__antithesis_instrumentation__.Notify(648547)

		threshold := thresholds[timeSeries.Resolution]

		targetSpan := roachpb.Span{
			Key: MakeDataKey(timeSeries.Name, "", timeSeries.Resolution, 0),
			EndKey: MakeDataKey(
				timeSeries.Name, "", timeSeries.Resolution, threshold,
			),
		}

		rollupDataMap := make(map[string]rollupData)

		account := qmc.workerMonitor.MakeBoundAccount()
		defer account.Close(ctx)

		childQmc := QueryMemoryContext{
			workerMonitor:      qmc.workerMonitor,
			resultAccount:      &account,
			QueryMemoryOptions: qmc.QueryMemoryOptions,
		}
		for querySpan := targetSpan; querySpan.Valid(); {
			__antithesis_instrumentation__.Notify(648552)
			var err error
			querySpan, err = db.queryAndComputeRollupsForSpan(
				ctx, timeSeries, querySpan, targetResolution, rollupDataMap, childQmc,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(648553)
				return err
			} else {
				__antithesis_instrumentation__.Notify(648554)
			}
		}
		__antithesis_instrumentation__.Notify(648548)

		var rollupDataSlice []rollupData
		for _, data := range rollupDataMap {
			__antithesis_instrumentation__.Notify(648555)
			rollupDataSlice = append(rollupDataSlice, data)
		}
		__antithesis_instrumentation__.Notify(648549)
		if err := db.storeRollup(ctx, targetResolution, rollupDataSlice); err != nil {
			__antithesis_instrumentation__.Notify(648556)
			return err
		} else {
			__antithesis_instrumentation__.Notify(648557)
		}
	}
	__antithesis_instrumentation__.Notify(648545)
	return nil
}

func (db *DB) queryAndComputeRollupsForSpan(
	ctx context.Context,
	series timeSeriesResolutionInfo,
	span roachpb.Span,
	targetResolution Resolution,
	rollupDataMap map[string]rollupData,
	qmc QueryMemoryContext,
) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(648558)
	b := &kv.Batch{}
	b.Header.MaxSpanRequestKeys = qmc.GetMaxRollupSlabs(series.Resolution)
	b.Scan(span.Key, span.EndKey)
	if err := db.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(648562)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(648563)
	}
	__antithesis_instrumentation__.Notify(648559)

	diskAccount := qmc.workerMonitor.MakeBoundAccount()
	defer diskAccount.Close(ctx)
	sourceSpans, err := convertKeysToSpans(ctx, b.Results[0].Rows, &diskAccount)
	if err != nil {
		__antithesis_instrumentation__.Notify(648564)
		return roachpb.Span{}, err
	} else {
		__antithesis_instrumentation__.Notify(648565)
	}
	__antithesis_instrumentation__.Notify(648560)

	for source, span := range sourceSpans {
		__antithesis_instrumentation__.Notify(648566)
		rollup, ok := rollupDataMap[source]
		if !ok {
			__antithesis_instrumentation__.Notify(648569)
			rollup = rollupData{
				name:   series.Name,
				source: source,
			}
			if err := qmc.resultAccount.Grow(ctx, int64(unsafe.Sizeof(rollup))); err != nil {
				__antithesis_instrumentation__.Notify(648570)
				return roachpb.Span{}, err
			} else {
				__antithesis_instrumentation__.Notify(648571)
			}
		} else {
			__antithesis_instrumentation__.Notify(648572)
		}
		__antithesis_instrumentation__.Notify(648567)

		var end timeSeriesSpanIterator
		for start := makeTimeSeriesSpanIterator(span); start.isValid(); start = end {
			__antithesis_instrumentation__.Notify(648573)
			rollupPeriod := targetResolution.SampleDuration()
			sampleTimestamp := normalizeToPeriod(start.timestamp, rollupPeriod)
			datapoint := rollupDatapoint{
				timestampNanos: sampleTimestamp,
				max:            -math.MaxFloat64,
				min:            math.MaxFloat64,
				first:          start.first(),
			}
			if err := qmc.resultAccount.Grow(ctx, int64(unsafe.Sizeof(datapoint))); err != nil {
				__antithesis_instrumentation__.Notify(648576)
				return roachpb.Span{}, err
			} else {
				__antithesis_instrumentation__.Notify(648577)
			}
			__antithesis_instrumentation__.Notify(648574)
			for end = start; end.isValid() && func() bool {
				__antithesis_instrumentation__.Notify(648578)
				return normalizeToPeriod(end.timestamp, rollupPeriod) == sampleTimestamp == true
			}() == true; end.forward() {
				__antithesis_instrumentation__.Notify(648579)
				datapoint.last = end.last()
				datapoint.max = math.Max(datapoint.max, end.max())
				datapoint.min = math.Min(datapoint.min, end.min())

				if datapoint.count > 0 {
					__antithesis_instrumentation__.Notify(648581)
					datapoint.variance = computeParallelVariance(
						parallelVarianceArgs{
							count:    end.count(),
							average:  end.average(),
							variance: end.variance(),
						},
						parallelVarianceArgs{
							count:    datapoint.count,
							average:  datapoint.sum / float64(datapoint.count),
							variance: datapoint.variance,
						},
					)
				} else {
					__antithesis_instrumentation__.Notify(648582)
				}
				__antithesis_instrumentation__.Notify(648580)

				datapoint.count += end.count()
				datapoint.sum += end.sum()
			}
			__antithesis_instrumentation__.Notify(648575)
			rollup.datapoints = append(rollup.datapoints, datapoint)
		}
		__antithesis_instrumentation__.Notify(648568)
		rollupDataMap[source] = rollup
	}
	__antithesis_instrumentation__.Notify(648561)
	return b.Results[0].ResumeSpanAsValue(), nil
}

type parallelVarianceArgs struct {
	count    uint32
	average  float64
	variance float64
}

func computeParallelVariance(left, right parallelVarianceArgs) float64 {
	__antithesis_instrumentation__.Notify(648583)
	leftCount := float64(left.count)
	rightCount := float64(right.count)
	totalCount := leftCount + rightCount
	averageDelta := left.average - right.average
	leftSumOfSquareDeviations := left.variance * leftCount
	rightSumOfSquareDeviations := right.variance * rightCount
	totalSumOfSquareDeviations := leftSumOfSquareDeviations + rightSumOfSquareDeviations + (averageDelta*averageDelta)*rightCount*leftCount/totalCount
	return totalSumOfSquareDeviations / totalCount
}
