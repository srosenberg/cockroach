package testmodel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
)

type ModelDB struct {
	data                    map[string]DataSeries
	metricNameToDataSources map[string]map[string]struct{}
	seenDataSources         map[string]struct{}
}

func NewModelDB() *ModelDB {
	__antithesis_instrumentation__.Notify(648775)
	return &ModelDB{
		data:                    make(map[string]DataSeries),
		metricNameToDataSources: make(map[string]map[string]struct{}),
		seenDataSources:         make(map[string]struct{}),
	}
}

type seriesVisitor func(string, string, DataSeries) (DataSeries, bool)

func (mdb *ModelDB) UniqueSourceCount() int64 {
	__antithesis_instrumentation__.Notify(648776)
	return int64(len(mdb.seenDataSources))
}

func (mdb *ModelDB) VisitAllSeries(visitor seriesVisitor) {
	__antithesis_instrumentation__.Notify(648777)
	for k := range mdb.data {
		__antithesis_instrumentation__.Notify(648778)
		metricName, dataSource := splitSeriesName(k)
		replacement, replace := visitor(metricName, dataSource, mdb.data[k])
		if replace {
			__antithesis_instrumentation__.Notify(648779)
			mdb.data[k] = replacement
		} else {
			__antithesis_instrumentation__.Notify(648780)
		}
	}
}

func (mdb *ModelDB) VisitSeries(name string, visitor seriesVisitor) {
	__antithesis_instrumentation__.Notify(648781)
	sourceMap, ok := mdb.metricNameToDataSources[name]
	if !ok {
		__antithesis_instrumentation__.Notify(648783)
		return
	} else {
		__antithesis_instrumentation__.Notify(648784)
	}
	__antithesis_instrumentation__.Notify(648782)

	for source := range sourceMap {
		__antithesis_instrumentation__.Notify(648785)
		replacement, replace := visitor(name, source, mdb.getSeriesData(name, source))
		if replace {
			__antithesis_instrumentation__.Notify(648786)
			mdb.data[seriesName(name, source)] = replacement
		} else {
			__antithesis_instrumentation__.Notify(648787)
		}
	}
}

func (mdb *ModelDB) Record(metricName, dataSource string, data DataSeries) {
	__antithesis_instrumentation__.Notify(648788)
	dataSources, ok := mdb.metricNameToDataSources[metricName]
	if !ok {
		__antithesis_instrumentation__.Notify(648790)
		dataSources = make(map[string]struct{})
		mdb.metricNameToDataSources[metricName] = dataSources
	} else {
		__antithesis_instrumentation__.Notify(648791)
	}
	__antithesis_instrumentation__.Notify(648789)
	dataSources[dataSource] = struct{}{}
	mdb.seenDataSources[dataSource] = struct{}{}

	seriesName := seriesName(metricName, dataSource)
	mdb.data[seriesName] = append(mdb.data[seriesName], data...)
	sort.Stable(mdb.data[seriesName])
	mdb.data[seriesName] = mdb.data[seriesName].removeDuplicates()
}

func (mdb *ModelDB) Query(
	name string,
	sources []string,
	downsample, agg tspb.TimeSeriesQueryAggregator,
	derivative tspb.TimeSeriesQueryDerivative,
	slabDuration, sampleDuration, start, end, interpolationLimit, now int64,
) DataSeries {
	__antithesis_instrumentation__.Notify(648792)
	start = normalizeTime(start, sampleDuration)

	cutoff := now - sampleDuration
	if start > cutoff {
		__antithesis_instrumentation__.Notify(648798)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648799)
	}
	__antithesis_instrumentation__.Notify(648793)
	if end > cutoff {
		__antithesis_instrumentation__.Notify(648800)
		end = cutoff
	} else {
		__antithesis_instrumentation__.Notify(648801)
	}
	__antithesis_instrumentation__.Notify(648794)

	end++

	if len(sources) == 0 {
		__antithesis_instrumentation__.Notify(648802)
		sourceMap, ok := mdb.metricNameToDataSources[name]
		if !ok {
			__antithesis_instrumentation__.Notify(648804)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(648805)
		}
		__antithesis_instrumentation__.Notify(648803)
		sources = make([]string, 0, len(sourceMap))
		for k := range sourceMap {
			__antithesis_instrumentation__.Notify(648806)
			sources = append(sources, k)
		}
	} else {
		__antithesis_instrumentation__.Notify(648807)
	}
	__antithesis_instrumentation__.Notify(648795)

	queryData := make([]DataSeries, 0, len(sources))
	for _, source := range sources {
		__antithesis_instrumentation__.Notify(648808)
		queryData = append(queryData, mdb.getSeriesData(name, source))
	}
	__antithesis_instrumentation__.Notify(648796)

	beforeFill := make([]DataSeries, len(queryData))

	adjustedStart := normalizeTime(start-interpolationLimit, slabDuration)
	adjustedEnd := normalizeTime(end+interpolationLimit-1, slabDuration) + slabDuration
	for i := range queryData {
		__antithesis_instrumentation__.Notify(648809)
		data := queryData[i]

		data = data.TimeSlice(adjustedStart, adjustedEnd)

		data = data.GroupByResolution(sampleDuration, getAggFunction(downsample))

		beforeFill[i] = data

		if len(queryData) > 1 {
			__antithesis_instrumentation__.Notify(648812)
			data = data.fillForResolution(
				sampleDuration,
				func(before DataSeries, after DataSeries, res int64) DataSeries {
					__antithesis_instrumentation__.Notify(648813)

					start := before[len(before)-1]
					end := after[0]
					if interpolationLimit > 0 && func() bool {
						__antithesis_instrumentation__.Notify(648815)
						return end.TimestampNanos-start.TimestampNanos > interpolationLimit == true
					}() == true {
						__antithesis_instrumentation__.Notify(648816)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(648817)
					}
					__antithesis_instrumentation__.Notify(648814)

					return fillFuncLinearInterpolate(before, after, res)
				},
			)
		} else {
			__antithesis_instrumentation__.Notify(648818)
		}
		__antithesis_instrumentation__.Notify(648810)

		if derivative != tspb.TimeSeriesQueryDerivative_NONE {
			__antithesis_instrumentation__.Notify(648819)
			data = data.rateOfChange(time.Second.Nanoseconds())
			if derivative == tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE {
				__antithesis_instrumentation__.Notify(648820)
				data = data.nonNegative()
			} else {
				__antithesis_instrumentation__.Notify(648821)
			}
		} else {
			__antithesis_instrumentation__.Notify(648822)
		}
		__antithesis_instrumentation__.Notify(648811)

		queryData[i] = data
	}
	__antithesis_instrumentation__.Notify(648797)

	result := groupSeriesByTimestamp(queryData, getAggFunction(agg))
	result = result.TimeSlice(start, end)
	result = result.intersectTimestamps(beforeFill...)
	return result
}

func (mdb *ModelDB) getSeriesData(metricName, dataSource string) DataSeries {
	__antithesis_instrumentation__.Notify(648823)
	seriesName := seriesName(metricName, dataSource)
	data, ok := mdb.data[seriesName]
	if !ok {
		__antithesis_instrumentation__.Notify(648825)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648826)
	}
	__antithesis_instrumentation__.Notify(648824)
	return data
}

func seriesName(metricName, dataSource string) string {
	__antithesis_instrumentation__.Notify(648827)
	return fmt.Sprintf("%s$$%s", metricName, dataSource)
}

func splitSeriesName(seriesName string) (string, string) {
	__antithesis_instrumentation__.Notify(648828)
	split := strings.Split(seriesName, "$$")
	if len(split) != 2 {
		__antithesis_instrumentation__.Notify(648830)
		panic(fmt.Sprintf("attempt to split invalid series name %s", seriesName))
	} else {
		__antithesis_instrumentation__.Notify(648831)
	}
	__antithesis_instrumentation__.Notify(648829)
	return split[0], split[1]
}
