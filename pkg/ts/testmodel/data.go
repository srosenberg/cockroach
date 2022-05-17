package testmodel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
)

func dp(timestamp int64, value float64) tspb.TimeSeriesDatapoint {
	__antithesis_instrumentation__.Notify(648676)
	return tspb.TimeSeriesDatapoint{
		TimestampNanos: timestamp,
		Value:          value,
	}
}

type DataSeries []tspb.TimeSeriesDatapoint

func (data DataSeries) Len() int { __antithesis_instrumentation__.Notify(648677); return len(data) }
func (data DataSeries) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(648678)
	data[i], data[j] = data[j], data[i]
}
func (data DataSeries) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(648679)
	return data[i].TimestampNanos < data[j].TimestampNanos
}

func normalizeTime(time, resolution int64) int64 {
	__antithesis_instrumentation__.Notify(648680)
	return time - time%resolution
}

func (data DataSeries) TimeSlice(start, end int64) DataSeries {
	__antithesis_instrumentation__.Notify(648681)
	startIdx := sort.Search(len(data), func(i int) bool {
		__antithesis_instrumentation__.Notify(648685)
		return data[i].TimestampNanos >= start
	})
	__antithesis_instrumentation__.Notify(648682)
	endIdx := sort.Search(len(data), func(i int) bool {
		__antithesis_instrumentation__.Notify(648686)
		return end <= data[i].TimestampNanos
	})
	__antithesis_instrumentation__.Notify(648683)

	result := data[startIdx:endIdx]
	if len(result) == 0 {
		__antithesis_instrumentation__.Notify(648687)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648688)
	}
	__antithesis_instrumentation__.Notify(648684)
	return result
}

func (data DataSeries) GroupByResolution(resolution int64, aggFunc aggFunc) DataSeries {
	__antithesis_instrumentation__.Notify(648689)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(648692)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648693)
	}
	__antithesis_instrumentation__.Notify(648690)

	result := make(DataSeries, 0)

	for len(data) > 0 {
		__antithesis_instrumentation__.Notify(648694)
		bucketTime := normalizeTime(data[0].TimestampNanos, resolution)

		bucketEndIdx := sort.Search(len(data), func(idx int) bool {
			__antithesis_instrumentation__.Notify(648696)
			return normalizeTime(data[idx].TimestampNanos, resolution) > bucketTime
		})
		__antithesis_instrumentation__.Notify(648695)

		result = append(result, dp(bucketTime, aggFunc(data[:bucketEndIdx])))
		data = data[bucketEndIdx:]
	}
	__antithesis_instrumentation__.Notify(648691)

	return result
}

func (data DataSeries) fillForResolution(resolution int64, fillFunc fillFunc) DataSeries {
	__antithesis_instrumentation__.Notify(648697)
	if len(data) < 2 {
		__antithesis_instrumentation__.Notify(648700)
		return data
	} else {
		__antithesis_instrumentation__.Notify(648701)
	}
	__antithesis_instrumentation__.Notify(648698)

	result := make(DataSeries, 0, len(data))
	result = append(result, data[0])
	for i := 1; i < len(data); i++ {
		__antithesis_instrumentation__.Notify(648702)
		if data[i].TimestampNanos-data[i-1].TimestampNanos > resolution {
			__antithesis_instrumentation__.Notify(648704)
			result = append(result, fillFunc(data[:i], data[i:], resolution)...)
		} else {
			__antithesis_instrumentation__.Notify(648705)
		}
		__antithesis_instrumentation__.Notify(648703)
		result = append(result, data[i])
	}
	__antithesis_instrumentation__.Notify(648699)

	return result
}

func (data DataSeries) rateOfChange(period int64) DataSeries {
	__antithesis_instrumentation__.Notify(648706)
	if len(data) < 2 {
		__antithesis_instrumentation__.Notify(648709)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648710)
	}
	__antithesis_instrumentation__.Notify(648707)

	result := make(DataSeries, len(data)-1)
	for i := 1; i < len(data); i++ {
		__antithesis_instrumentation__.Notify(648711)
		result[i-1] = dp(
			data[i].TimestampNanos,
			(data[i].Value-data[i-1].Value)/(float64(data[i].TimestampNanos-data[i-1].TimestampNanos)/float64(period)),
		)
	}
	__antithesis_instrumentation__.Notify(648708)
	return result
}

func (data DataSeries) nonNegative() DataSeries {
	__antithesis_instrumentation__.Notify(648712)
	result := make(DataSeries, len(data))
	for i := range data {
		__antithesis_instrumentation__.Notify(648714)
		if data[i].Value >= 0 {
			__antithesis_instrumentation__.Notify(648715)
			result[i] = data[i]
		} else {
			__antithesis_instrumentation__.Notify(648716)
			result[i] = dp(data[i].TimestampNanos, 0)
		}
	}
	__antithesis_instrumentation__.Notify(648713)
	return result
}

func (data DataSeries) adjustTimestamps(offset int64) DataSeries {
	__antithesis_instrumentation__.Notify(648717)
	result := make(DataSeries, len(data))
	for i := range data {
		__antithesis_instrumentation__.Notify(648719)
		result[i] = data[i]
		result[i].TimestampNanos += offset
	}
	__antithesis_instrumentation__.Notify(648718)
	return result
}

func (data DataSeries) removeDuplicates() DataSeries {
	__antithesis_instrumentation__.Notify(648720)
	if len(data) < 2 {
		__antithesis_instrumentation__.Notify(648723)
		return data
	} else {
		__antithesis_instrumentation__.Notify(648724)
	}
	__antithesis_instrumentation__.Notify(648721)

	result := make(DataSeries, len(data))
	result[0] = data[0]
	for i, j := 1, 0; i < len(data); i++ {
		__antithesis_instrumentation__.Notify(648725)
		if result[j].TimestampNanos == data[i].TimestampNanos {
			__antithesis_instrumentation__.Notify(648726)

			result[j].Value = data[i].Value
			result = result[:len(result)-1]
		} else {
			__antithesis_instrumentation__.Notify(648727)
			j++
			result[j] = data[i]
		}
	}
	__antithesis_instrumentation__.Notify(648722)

	return result
}

func groupSeriesByTimestamp(datas []DataSeries, aggFunc aggFunc) DataSeries {
	__antithesis_instrumentation__.Notify(648728)
	if len(datas) == 0 {
		__antithesis_instrumentation__.Notify(648731)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648732)
	}
	__antithesis_instrumentation__.Notify(648729)

	results := make(DataSeries, 0)
	dataPointsToAggregate := make(DataSeries, 0, len(datas))
	for {
		__antithesis_instrumentation__.Notify(648733)

		origDatas := datas
		datas = datas[:0]
		for _, data := range origDatas {
			__antithesis_instrumentation__.Notify(648737)
			if len(data) > 0 {
				__antithesis_instrumentation__.Notify(648738)
				datas = append(datas, data)
			} else {
				__antithesis_instrumentation__.Notify(648739)
			}
		}
		__antithesis_instrumentation__.Notify(648734)
		if len(datas) == 0 {
			__antithesis_instrumentation__.Notify(648740)
			break
		} else {
			__antithesis_instrumentation__.Notify(648741)
		}
		__antithesis_instrumentation__.Notify(648735)

		earliestTime := int64(math.MaxInt64)
		for _, data := range datas {
			__antithesis_instrumentation__.Notify(648742)
			if data[0].TimestampNanos < earliestTime {
				__antithesis_instrumentation__.Notify(648744)

				dataPointsToAggregate = dataPointsToAggregate[:0]
				earliestTime = data[0].TimestampNanos
			} else {
				__antithesis_instrumentation__.Notify(648745)
			}
			__antithesis_instrumentation__.Notify(648743)
			if data[0].TimestampNanos == earliestTime {
				__antithesis_instrumentation__.Notify(648746)

				dataPointsToAggregate = append(dataPointsToAggregate, data[0])
			} else {
				__antithesis_instrumentation__.Notify(648747)
			}
		}
		__antithesis_instrumentation__.Notify(648736)
		results = append(results, dp(earliestTime, aggFunc(dataPointsToAggregate)))
		for i := range datas {
			__antithesis_instrumentation__.Notify(648748)
			if datas[i][0].TimestampNanos == earliestTime {
				__antithesis_instrumentation__.Notify(648749)
				datas[i] = datas[i][1:]
			} else {
				__antithesis_instrumentation__.Notify(648750)
			}
		}
	}
	__antithesis_instrumentation__.Notify(648730)

	return results
}

func (data DataSeries) intersectTimestamps(datas ...DataSeries) DataSeries {
	__antithesis_instrumentation__.Notify(648751)
	seenTimestamps := make(map[int64]struct{})
	for _, ds := range datas {
		__antithesis_instrumentation__.Notify(648754)
		for _, dp := range ds {
			__antithesis_instrumentation__.Notify(648755)
			seenTimestamps[dp.TimestampNanos] = struct{}{}
		}
	}
	__antithesis_instrumentation__.Notify(648752)

	result := make(DataSeries, 0, len(data))
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648756)
		if _, ok := seenTimestamps[dp.TimestampNanos]; ok {
			__antithesis_instrumentation__.Notify(648757)
			result = append(result, dp)
		} else {
			__antithesis_instrumentation__.Notify(648758)
		}
	}
	__antithesis_instrumentation__.Notify(648753)
	return result
}

const floatTolerance float64 = 0.0001

func floatEquals(a, b float64) bool {
	__antithesis_instrumentation__.Notify(648759)
	if (a-b) < floatTolerance && func() bool {
		__antithesis_instrumentation__.Notify(648761)
		return (b - a) < floatTolerance == true
	}() == true {
		__antithesis_instrumentation__.Notify(648762)
		return true
	} else {
		__antithesis_instrumentation__.Notify(648763)
	}
	__antithesis_instrumentation__.Notify(648760)
	return false
}

func DataSeriesEquivalent(a, b DataSeries) bool {
	__antithesis_instrumentation__.Notify(648764)
	if len(a) != len(b) {
		__antithesis_instrumentation__.Notify(648767)
		return false
	} else {
		__antithesis_instrumentation__.Notify(648768)
	}
	__antithesis_instrumentation__.Notify(648765)
	for i := range a {
		__antithesis_instrumentation__.Notify(648769)
		if a[i].TimestampNanos != b[i].TimestampNanos {
			__antithesis_instrumentation__.Notify(648771)
			return false
		} else {
			__antithesis_instrumentation__.Notify(648772)
		}
		__antithesis_instrumentation__.Notify(648770)
		if !floatEquals(a[i].Value, b[i].Value) {
			__antithesis_instrumentation__.Notify(648773)
			return false
		} else {
			__antithesis_instrumentation__.Notify(648774)
		}
	}
	__antithesis_instrumentation__.Notify(648766)
	return true
}
