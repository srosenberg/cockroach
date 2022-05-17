package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/errors"
)

type timeSeriesSpan []roachpb.InternalTimeSeriesData

type timeSeriesSpanIterator struct {
	span      timeSeriesSpan
	total     int
	outer     int
	inner     int
	timestamp int64
	length    int
}

func makeTimeSeriesSpanIterator(span timeSeriesSpan) timeSeriesSpanIterator {
	__antithesis_instrumentation__.Notify(648125)
	iterator := timeSeriesSpanIterator{
		span: span,
	}
	iterator.computeLength()
	iterator.computeTimestamp()
	return iterator
}

func (tsi *timeSeriesSpanIterator) computeLength() {
	__antithesis_instrumentation__.Notify(648126)
	tsi.length = 0
	for _, data := range tsi.span {
		__antithesis_instrumentation__.Notify(648127)
		tsi.length += data.SampleCount()
	}
}

func (tsi *timeSeriesSpanIterator) computeTimestamp() {
	__antithesis_instrumentation__.Notify(648128)
	if !tsi.isValid() {
		__antithesis_instrumentation__.Notify(648130)
		tsi.timestamp = 0
		return
	} else {
		__antithesis_instrumentation__.Notify(648131)
	}
	__antithesis_instrumentation__.Notify(648129)
	data := tsi.span[tsi.outer]
	tsi.timestamp = data.StartTimestampNanos + data.SampleDurationNanos*int64(tsi.offset())
}

func (tsi *timeSeriesSpanIterator) forward() {
	__antithesis_instrumentation__.Notify(648132)
	if !tsi.isValid() {
		__antithesis_instrumentation__.Notify(648135)
		return
	} else {
		__antithesis_instrumentation__.Notify(648136)
	}
	__antithesis_instrumentation__.Notify(648133)
	tsi.total++
	tsi.inner++
	if tsi.inner >= tsi.span[tsi.outer].SampleCount() {
		__antithesis_instrumentation__.Notify(648137)
		tsi.inner = 0
		tsi.outer++
	} else {
		__antithesis_instrumentation__.Notify(648138)
	}
	__antithesis_instrumentation__.Notify(648134)
	tsi.computeTimestamp()
}

func (tsi *timeSeriesSpanIterator) backward() {
	__antithesis_instrumentation__.Notify(648139)
	if tsi.outer == 0 && func() bool {
		__antithesis_instrumentation__.Notify(648142)
		return tsi.inner == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(648143)
		return
	} else {
		__antithesis_instrumentation__.Notify(648144)
	}
	__antithesis_instrumentation__.Notify(648140)
	tsi.total--
	if tsi.inner == 0 {
		__antithesis_instrumentation__.Notify(648145)
		tsi.outer--
		tsi.inner = tsi.span[tsi.outer].SampleCount() - 1
	} else {
		__antithesis_instrumentation__.Notify(648146)
		tsi.inner--
	}
	__antithesis_instrumentation__.Notify(648141)
	tsi.computeTimestamp()
}

func (tsi *timeSeriesSpanIterator) seekIndex(index int) {
	__antithesis_instrumentation__.Notify(648147)
	if index >= tsi.length {
		__antithesis_instrumentation__.Notify(648151)
		tsi.total = tsi.length
		tsi.inner = 0
		tsi.outer = len(tsi.span)
		tsi.timestamp = 0
		return
	} else {
		__antithesis_instrumentation__.Notify(648152)
	}
	__antithesis_instrumentation__.Notify(648148)

	if index < 0 {
		__antithesis_instrumentation__.Notify(648153)
		index = 0
	} else {
		__antithesis_instrumentation__.Notify(648154)
	}
	__antithesis_instrumentation__.Notify(648149)

	remaining := index
	newOuter := 0
	for len(tsi.span) > newOuter && func() bool {
		__antithesis_instrumentation__.Notify(648155)
		return remaining >= tsi.span[newOuter].SampleCount() == true
	}() == true {
		__antithesis_instrumentation__.Notify(648156)
		remaining -= tsi.span[newOuter].SampleCount()
		newOuter++
	}
	__antithesis_instrumentation__.Notify(648150)
	tsi.inner = remaining
	tsi.outer = newOuter
	tsi.total = index
	tsi.computeTimestamp()
}

func (tsi *timeSeriesSpanIterator) seekTimestamp(timestamp int64) {
	__antithesis_instrumentation__.Notify(648157)
	seeker := *tsi
	index := sort.Search(tsi.length, func(i int) bool {
		__antithesis_instrumentation__.Notify(648159)
		seeker.seekIndex(i)
		return seeker.timestamp >= timestamp
	})
	__antithesis_instrumentation__.Notify(648158)
	tsi.seekIndex(index)
}

func (tsi *timeSeriesSpanIterator) isColumnar() bool {
	__antithesis_instrumentation__.Notify(648160)
	return tsi.span[tsi.outer].IsColumnar()
}

func (tsi *timeSeriesSpanIterator) isRollup() bool {
	__antithesis_instrumentation__.Notify(648161)
	return tsi.span[tsi.outer].IsRollup()
}

func (tsi *timeSeriesSpanIterator) offset() int32 {
	__antithesis_instrumentation__.Notify(648162)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648164)
		return data.Offset[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648165)
	}
	__antithesis_instrumentation__.Notify(648163)
	return data.Samples[tsi.inner].Offset
}

func (tsi *timeSeriesSpanIterator) count() uint32 {
	__antithesis_instrumentation__.Notify(648166)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648168)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648170)
			return data.Count[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648171)
		}
		__antithesis_instrumentation__.Notify(648169)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(648172)
	}
	__antithesis_instrumentation__.Notify(648167)
	return data.Samples[tsi.inner].Count
}

func (tsi *timeSeriesSpanIterator) sum() float64 {
	__antithesis_instrumentation__.Notify(648173)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648175)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648177)
			return data.Sum[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648178)
		}
		__antithesis_instrumentation__.Notify(648176)
		return data.Last[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648179)
	}
	__antithesis_instrumentation__.Notify(648174)
	return data.Samples[tsi.inner].Sum
}

func (tsi *timeSeriesSpanIterator) max() float64 {
	__antithesis_instrumentation__.Notify(648180)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648183)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648185)
			return data.Max[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648186)
		}
		__antithesis_instrumentation__.Notify(648184)
		return data.Last[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648187)
	}
	__antithesis_instrumentation__.Notify(648181)
	if max := data.Samples[tsi.inner].Max; max != nil {
		__antithesis_instrumentation__.Notify(648188)
		return *max
	} else {
		__antithesis_instrumentation__.Notify(648189)
	}
	__antithesis_instrumentation__.Notify(648182)
	return data.Samples[tsi.inner].Sum
}

func (tsi *timeSeriesSpanIterator) min() float64 {
	__antithesis_instrumentation__.Notify(648190)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648193)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648195)
			return data.Min[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648196)
		}
		__antithesis_instrumentation__.Notify(648194)
		return data.Last[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648197)
	}
	__antithesis_instrumentation__.Notify(648191)
	if min := data.Samples[tsi.inner].Min; min != nil {
		__antithesis_instrumentation__.Notify(648198)
		return *min
	} else {
		__antithesis_instrumentation__.Notify(648199)
	}
	__antithesis_instrumentation__.Notify(648192)
	return data.Samples[tsi.inner].Sum
}

func (tsi *timeSeriesSpanIterator) first() float64 {
	__antithesis_instrumentation__.Notify(648200)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648202)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648204)
			return data.First[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648205)
		}
		__antithesis_instrumentation__.Notify(648203)
		return data.Last[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648206)
	}
	__antithesis_instrumentation__.Notify(648201)

	return data.Samples[tsi.inner].Sum
}

func (tsi *timeSeriesSpanIterator) last() float64 {
	__antithesis_instrumentation__.Notify(648207)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648209)
		return data.Last[tsi.inner]
	} else {
		__antithesis_instrumentation__.Notify(648210)
	}
	__antithesis_instrumentation__.Notify(648208)

	return data.Samples[tsi.inner].Sum
}

func (tsi *timeSeriesSpanIterator) variance() float64 {
	__antithesis_instrumentation__.Notify(648211)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648213)
		if tsi.isRollup() {
			__antithesis_instrumentation__.Notify(648215)
			return data.Variance[tsi.inner]
		} else {
			__antithesis_instrumentation__.Notify(648216)
		}
		__antithesis_instrumentation__.Notify(648214)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(648217)
	}
	__antithesis_instrumentation__.Notify(648212)

	return 0
}

func (tsi *timeSeriesSpanIterator) average() float64 {
	__antithesis_instrumentation__.Notify(648218)
	return tsi.sum() / float64(tsi.count())
}

func (tsi *timeSeriesSpanIterator) setOffset(value int32) {
	__antithesis_instrumentation__.Notify(648219)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648221)
		data.Offset[tsi.inner] = value
		return
	} else {
		__antithesis_instrumentation__.Notify(648222)
	}
	__antithesis_instrumentation__.Notify(648220)
	data.Samples[tsi.inner].Offset = value
}

func (tsi *timeSeriesSpanIterator) setSingleValue(value float64) {
	__antithesis_instrumentation__.Notify(648223)
	data := tsi.span[tsi.outer]
	if tsi.isColumnar() {
		__antithesis_instrumentation__.Notify(648225)
		data.Last[tsi.inner] = value
		return
	} else {
		__antithesis_instrumentation__.Notify(648226)
	}
	__antithesis_instrumentation__.Notify(648224)
	data.Samples[tsi.inner].Sum = value
	data.Samples[tsi.inner].Count = 1
	data.Samples[tsi.inner].Min = nil
	data.Samples[tsi.inner].Max = nil
}

func (tsi *timeSeriesSpanIterator) truncateSpan() {
	__antithesis_instrumentation__.Notify(648227)
	var outerExtent int
	if tsi.inner == 0 {
		__antithesis_instrumentation__.Notify(648231)
		outerExtent = tsi.outer
	} else {
		__antithesis_instrumentation__.Notify(648232)
		outerExtent = tsi.outer + 1
	}
	__antithesis_instrumentation__.Notify(648228)

	unused := tsi.span[outerExtent:]
	tsi.span = tsi.span[:outerExtent]
	for i := range unused {
		__antithesis_instrumentation__.Notify(648233)
		unused[i] = roachpb.InternalTimeSeriesData{}
	}
	__antithesis_instrumentation__.Notify(648229)

	if tsi.inner != 0 {
		__antithesis_instrumentation__.Notify(648234)
		data := tsi.span[tsi.outer]
		size := tsi.inner
		if data.IsColumnar() {
			__antithesis_instrumentation__.Notify(648236)
			data.Offset = data.Offset[:size]
			data.Last = data.Last[:size]
			if data.IsRollup() {
				__antithesis_instrumentation__.Notify(648237)
				data.First = data.First[:size]
				data.Min = data.Min[:size]
				data.Max = data.Max[:size]
				data.Count = data.Count[:size]
				data.Sum = data.Sum[:size]
				data.Variance = data.Variance[:size]
			} else {
				__antithesis_instrumentation__.Notify(648238)
			}
		} else {
			__antithesis_instrumentation__.Notify(648239)
			data.Samples = data.Samples[:size]
		}
		__antithesis_instrumentation__.Notify(648235)
		tsi.span[tsi.outer] = data
	} else {
		__antithesis_instrumentation__.Notify(648240)
	}
	__antithesis_instrumentation__.Notify(648230)

	tsi.computeLength()
	tsi.computeTimestamp()
}

func convertToSingleValue(span timeSeriesSpan) {
	__antithesis_instrumentation__.Notify(648241)
	for i := range span {
		__antithesis_instrumentation__.Notify(648242)
		if span[i].IsColumnar() {
			__antithesis_instrumentation__.Notify(648243)
			span[i].Count = nil
			span[i].Sum = nil
			span[i].Min = nil
			span[i].Max = nil
			span[i].First = nil
			span[i].Variance = nil
		} else {
			__antithesis_instrumentation__.Notify(648244)
		}
	}
}

func (tsi *timeSeriesSpanIterator) value(downsampler tspb.TimeSeriesQueryAggregator) float64 {
	__antithesis_instrumentation__.Notify(648245)
	if !tsi.isValid() {
		__antithesis_instrumentation__.Notify(648248)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(648249)
	}
	__antithesis_instrumentation__.Notify(648246)
	switch downsampler {
	case tspb.TimeSeriesQueryAggregator_AVG:
		__antithesis_instrumentation__.Notify(648250)
		return tsi.sum() / float64(tsi.count())
	case tspb.TimeSeriesQueryAggregator_MAX:
		__antithesis_instrumentation__.Notify(648251)
		return tsi.max()
	case tspb.TimeSeriesQueryAggregator_MIN:
		__antithesis_instrumentation__.Notify(648252)
		return tsi.min()
	case tspb.TimeSeriesQueryAggregator_SUM:
		__antithesis_instrumentation__.Notify(648253)
		return tsi.sum()
	default:
		__antithesis_instrumentation__.Notify(648254)
	}
	__antithesis_instrumentation__.Notify(648247)

	panic(fmt.Sprintf("unknown downsampler option encountered: %v", downsampler))
}

func (tsi *timeSeriesSpanIterator) valueAtTimestamp(
	timestamp int64, interpolationLimitNanos int64, downsampler tspb.TimeSeriesQueryAggregator,
) (float64, bool) {
	__antithesis_instrumentation__.Notify(648255)
	if !tsi.validAtTimestamp(timestamp, interpolationLimitNanos) {
		__antithesis_instrumentation__.Notify(648259)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(648260)
	}
	__antithesis_instrumentation__.Notify(648256)
	if tsi.timestamp == timestamp {
		__antithesis_instrumentation__.Notify(648261)
		return tsi.value(downsampler), true
	} else {
		__antithesis_instrumentation__.Notify(648262)
	}
	__antithesis_instrumentation__.Notify(648257)

	deriv, valid := tsi.derivative(downsampler)
	if !valid {
		__antithesis_instrumentation__.Notify(648263)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(648264)
	}
	__antithesis_instrumentation__.Notify(648258)
	return tsi.value(downsampler) - deriv*float64((tsi.timestamp-timestamp)/tsi.samplePeriod()), true
}

func (tsi *timeSeriesSpanIterator) validAtTimestamp(timestamp, interpolationLimitNanos int64) bool {
	__antithesis_instrumentation__.Notify(648265)
	if !tsi.isValid() {
		__antithesis_instrumentation__.Notify(648271)
		return false
	} else {
		__antithesis_instrumentation__.Notify(648272)
	}
	__antithesis_instrumentation__.Notify(648266)
	if tsi.timestamp == timestamp {
		__antithesis_instrumentation__.Notify(648273)
		return true
	} else {
		__antithesis_instrumentation__.Notify(648274)
	}
	__antithesis_instrumentation__.Notify(648267)

	if tsi.total == 0 {
		__antithesis_instrumentation__.Notify(648275)
		return false
	} else {
		__antithesis_instrumentation__.Notify(648276)
	}
	__antithesis_instrumentation__.Notify(648268)
	prev := *tsi
	prev.backward()

	if timestamp > tsi.timestamp || func() bool {
		__antithesis_instrumentation__.Notify(648277)
		return timestamp <= prev.timestamp == true
	}() == true {
		__antithesis_instrumentation__.Notify(648278)
		return false
	} else {
		__antithesis_instrumentation__.Notify(648279)
	}
	__antithesis_instrumentation__.Notify(648269)

	if interpolationLimitNanos > 0 && func() bool {
		__antithesis_instrumentation__.Notify(648280)
		return tsi.timestamp-prev.timestamp > interpolationLimitNanos == true
	}() == true {
		__antithesis_instrumentation__.Notify(648281)
		return false
	} else {
		__antithesis_instrumentation__.Notify(648282)
	}
	__antithesis_instrumentation__.Notify(648270)
	return true
}

func (tsi *timeSeriesSpanIterator) derivative(
	downsampler tspb.TimeSeriesQueryAggregator,
) (float64, bool) {
	__antithesis_instrumentation__.Notify(648283)
	if !tsi.isValid() {
		__antithesis_instrumentation__.Notify(648286)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(648287)
	}
	__antithesis_instrumentation__.Notify(648284)

	if tsi.total == 0 {
		__antithesis_instrumentation__.Notify(648288)
		return 0, false
	} else {
		__antithesis_instrumentation__.Notify(648289)
	}
	__antithesis_instrumentation__.Notify(648285)

	prev := *tsi
	prev.backward()
	rateOfChange := (tsi.value(downsampler) - prev.value(downsampler)) / float64((tsi.timestamp-prev.timestamp)/tsi.samplePeriod())
	return rateOfChange, true
}

func (tsi *timeSeriesSpanIterator) samplePeriod() int64 {
	__antithesis_instrumentation__.Notify(648290)
	return tsi.span[0].SampleDurationNanos
}

func (tsi *timeSeriesSpanIterator) isValid() bool {
	__antithesis_instrumentation__.Notify(648291)
	return tsi.total < tsi.length
}

func (db *DB) Query(
	ctx context.Context,
	query tspb.Query,
	diskResolution Resolution,
	timespan QueryTimespan,
	mem QueryMemoryContext,
) ([]tspb.TimeSeriesDatapoint, []string, error) {
	__antithesis_instrumentation__.Notify(648292)
	timespan.normalize()

	if err := timespan.verifyBounds(); err != nil {
		__antithesis_instrumentation__.Notify(648301)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(648302)
	}
	__antithesis_instrumentation__.Notify(648293)
	if err := timespan.verifyDiskResolution(diskResolution); err != nil {
		__antithesis_instrumentation__.Notify(648303)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(648304)
	}
	__antithesis_instrumentation__.Notify(648294)
	if err := verifySourceAggregator(query.GetSourceAggregator()); err != nil {
		__antithesis_instrumentation__.Notify(648305)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(648306)
	}
	__antithesis_instrumentation__.Notify(648295)
	if err := verifyDownsampler(query.GetDownsampler()); err != nil {
		__antithesis_instrumentation__.Notify(648307)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(648308)
	}
	__antithesis_instrumentation__.Notify(648296)

	if err := timespan.adjustForCurrentTime(diskResolution); err != nil {
		__antithesis_instrumentation__.Notify(648309)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(648310)
	}
	__antithesis_instrumentation__.Notify(648297)

	var result []tspb.TimeSeriesDatapoint

	sourceSet := make(map[string]struct{})

	resolutions := []Resolution{diskResolution}
	if rollupResolution, ok := diskResolution.TargetRollupResolution(); ok {
		__antithesis_instrumentation__.Notify(648311)
		if timespan.verifyDiskResolution(rollupResolution) == nil {
			__antithesis_instrumentation__.Notify(648312)
			resolutions = []Resolution{rollupResolution, diskResolution}
		} else {
			__antithesis_instrumentation__.Notify(648313)
		}
	} else {
		__antithesis_instrumentation__.Notify(648314)
	}
	__antithesis_instrumentation__.Notify(648298)

	for _, resolution := range resolutions {
		__antithesis_instrumentation__.Notify(648315)

		maxTimespanWidth, err := mem.GetMaxTimespan(resolution)
		if err != nil {
			__antithesis_instrumentation__.Notify(648318)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(648319)
		}
		__antithesis_instrumentation__.Notify(648316)

		if maxTimespanWidth > timespan.width() {
			__antithesis_instrumentation__.Notify(648320)
			if err := db.queryChunk(
				ctx, query, resolution, timespan, mem, &result, sourceSet,
			); err != nil {
				__antithesis_instrumentation__.Notify(648321)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(648322)
			}
		} else {
			__antithesis_instrumentation__.Notify(648323)

			chunkTime := timespan
			chunkTime.EndNanos = chunkTime.StartNanos + maxTimespanWidth
			for ; chunkTime.StartNanos < timespan.EndNanos; chunkTime.moveForward(maxTimespanWidth + timespan.SampleDurationNanos) {
				__antithesis_instrumentation__.Notify(648324)
				if chunkTime.EndNanos > timespan.EndNanos {
					__antithesis_instrumentation__.Notify(648326)

					chunkTime.EndNanos = timespan.EndNanos
				} else {
					__antithesis_instrumentation__.Notify(648327)
				}
				__antithesis_instrumentation__.Notify(648325)
				if err := db.queryChunk(
					ctx, query, resolution, chunkTime, mem, &result, sourceSet,
				); err != nil {
					__antithesis_instrumentation__.Notify(648328)
					return nil, nil, err
				} else {
					__antithesis_instrumentation__.Notify(648329)
				}
			}
		}
		__antithesis_instrumentation__.Notify(648317)

		if len(resolutions) > 1 && func() bool {
			__antithesis_instrumentation__.Notify(648330)
			return len(result) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(648331)
			lastTime := result[len(result)-1].TimestampNanos
			if lastTime >= timespan.EndNanos {
				__antithesis_instrumentation__.Notify(648333)
				break
			} else {
				__antithesis_instrumentation__.Notify(648334)
			}
			__antithesis_instrumentation__.Notify(648332)
			timespan.StartNanos = lastTime
		} else {
			__antithesis_instrumentation__.Notify(648335)
		}
	}
	__antithesis_instrumentation__.Notify(648299)

	sources := make([]string, 0, len(sourceSet))
	for source := range sourceSet {
		__antithesis_instrumentation__.Notify(648336)
		sources = append(sources, source)
	}
	__antithesis_instrumentation__.Notify(648300)

	return result, sources, nil
}

func (db *DB) queryChunk(
	ctx context.Context,
	query tspb.Query,
	diskResolution Resolution,
	timespan QueryTimespan,
	mem QueryMemoryContext,
	dest *[]tspb.TimeSeriesDatapoint,
	sourceSet map[string]struct{},
) error {
	__antithesis_instrumentation__.Notify(648337)
	acc := mem.workerMonitor.MakeBoundAccount()
	defer acc.Close(ctx)

	diskTimespan := timespan
	diskTimespan.expand(mem.InterpolationLimitNanos)

	var data []kv.KeyValue
	var err error
	if len(query.Sources) == 0 {
		__antithesis_instrumentation__.Notify(648345)
		data, err = db.readAllSourcesFromDatabase(ctx, query.Name, diskResolution, diskTimespan)
	} else {
		__antithesis_instrumentation__.Notify(648346)
		data, err = db.readFromDatabase(ctx, query.Name, diskResolution, diskTimespan, query.Sources)
	}
	__antithesis_instrumentation__.Notify(648338)

	if err != nil {
		__antithesis_instrumentation__.Notify(648347)
		return err
	} else {
		__antithesis_instrumentation__.Notify(648348)
	}
	__antithesis_instrumentation__.Notify(648339)

	sourceSpans, err := convertKeysToSpans(ctx, data, &acc)
	if err != nil {
		__antithesis_instrumentation__.Notify(648349)
		return err
	} else {
		__antithesis_instrumentation__.Notify(648350)
	}
	__antithesis_instrumentation__.Notify(648340)
	if len(sourceSpans) == 0 {
		__antithesis_instrumentation__.Notify(648351)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(648352)
	}
	__antithesis_instrumentation__.Notify(648341)

	if timespan.SampleDurationNanos != diskResolution.SampleDuration() {
		__antithesis_instrumentation__.Notify(648353)
		downsampleSpans(sourceSpans, timespan.SampleDurationNanos, query.GetDownsampler())

		query.Downsampler = tspb.TimeSeriesQueryAggregator_SUM.Enum()
	} else {
		__antithesis_instrumentation__.Notify(648354)
	}
	__antithesis_instrumentation__.Notify(648342)

	oldCap := cap(*dest)
	aggregateSpansToDatapoints(sourceSpans, query, timespan, mem.InterpolationLimitNanos, dest)
	if oldCap > cap(*dest) {
		__antithesis_instrumentation__.Notify(648355)
		if err := mem.resultAccount.Grow(ctx, sizeOfDataPoint*int64(cap(*dest)-oldCap)); err != nil {
			__antithesis_instrumentation__.Notify(648356)
			return err
		} else {
			__antithesis_instrumentation__.Notify(648357)
		}
	} else {
		__antithesis_instrumentation__.Notify(648358)
	}
	__antithesis_instrumentation__.Notify(648343)

	for k := range sourceSpans {
		__antithesis_instrumentation__.Notify(648359)
		sourceSet[k] = struct{}{}
	}
	__antithesis_instrumentation__.Notify(648344)
	return nil
}

func downsampleSpans(
	spans map[string]timeSeriesSpan, duration int64, downsampler tspb.TimeSeriesQueryAggregator,
) {
	__antithesis_instrumentation__.Notify(648360)

	for k, span := range spans {
		__antithesis_instrumentation__.Notify(648361)
		nextInsert := makeTimeSeriesSpanIterator(span)
		for start, end := nextInsert, nextInsert; start.isValid(); start = end {
			__antithesis_instrumentation__.Notify(648363)
			sampleTimestamp := normalizeToPeriod(start.timestamp, duration)

			switch downsampler {
			case tspb.TimeSeriesQueryAggregator_MAX:
				__antithesis_instrumentation__.Notify(648365)
				max := -math.MaxFloat64
				for ; end.isValid() && func() bool {
					__antithesis_instrumentation__.Notify(648374)
					return normalizeToPeriod(end.timestamp, duration) == sampleTimestamp == true
				}() == true; end.forward() {
					__antithesis_instrumentation__.Notify(648375)
					max = math.Max(max, end.max())
				}
				__antithesis_instrumentation__.Notify(648366)
				nextInsert.setSingleValue(max)
			case tspb.TimeSeriesQueryAggregator_MIN:
				__antithesis_instrumentation__.Notify(648367)
				min := math.MaxFloat64
				for ; end.isValid() && func() bool {
					__antithesis_instrumentation__.Notify(648376)
					return normalizeToPeriod(end.timestamp, duration) == sampleTimestamp == true
				}() == true; end.forward() {
					__antithesis_instrumentation__.Notify(648377)
					min = math.Min(min, end.min())
				}
				__antithesis_instrumentation__.Notify(648368)
				nextInsert.setSingleValue(min)
			case tspb.TimeSeriesQueryAggregator_AVG:
				__antithesis_instrumentation__.Notify(648369)
				count, sum := uint32(0), 0.0
				for ; end.isValid() && func() bool {
					__antithesis_instrumentation__.Notify(648378)
					return normalizeToPeriod(end.timestamp, duration) == sampleTimestamp == true
				}() == true; end.forward() {
					__antithesis_instrumentation__.Notify(648379)
					count += end.count()
					sum += end.sum()
				}
				__antithesis_instrumentation__.Notify(648370)
				nextInsert.setSingleValue(sum / float64(count))
			case tspb.TimeSeriesQueryAggregator_SUM:
				__antithesis_instrumentation__.Notify(648371)
				sum := 0.0
				for ; end.isValid() && func() bool {
					__antithesis_instrumentation__.Notify(648380)
					return normalizeToPeriod(end.timestamp, duration) == sampleTimestamp == true
				}() == true; end.forward() {
					__antithesis_instrumentation__.Notify(648381)
					sum += end.sum()
				}
				__antithesis_instrumentation__.Notify(648372)
				nextInsert.setSingleValue(sum)
			default:
				__antithesis_instrumentation__.Notify(648373)
			}
			__antithesis_instrumentation__.Notify(648364)

			nextInsert.setOffset(span[nextInsert.outer].OffsetForTimestamp(sampleTimestamp))
			nextInsert.forward()
		}
		__antithesis_instrumentation__.Notify(648362)

		nextInsert.truncateSpan()
		span = nextInsert.span
		convertToSingleValue(span)
		spans[k] = span
	}
}

func aggregateSpansToDatapoints(
	spans map[string]timeSeriesSpan,
	query tspb.Query,
	timespan QueryTimespan,
	interpolationLimitNanos int64,
	dest *[]tspb.TimeSeriesDatapoint,
) {
	__antithesis_instrumentation__.Notify(648382)

	iterators := make([]timeSeriesSpanIterator, 0, len(spans))
	for _, span := range spans {
		__antithesis_instrumentation__.Notify(648385)
		iter := makeTimeSeriesSpanIterator(span)
		iter.seekTimestamp(timespan.StartNanos)
		iterators = append(iterators, iter)
	}
	__antithesis_instrumentation__.Notify(648383)

	var lowestTimestamp int64
	computeLowest := func() {
		__antithesis_instrumentation__.Notify(648386)
		lowestTimestamp = math.MaxInt64
		for _, iter := range iterators {
			__antithesis_instrumentation__.Notify(648387)
			if !iter.isValid() {
				__antithesis_instrumentation__.Notify(648389)
				continue
			} else {
				__antithesis_instrumentation__.Notify(648390)
			}
			__antithesis_instrumentation__.Notify(648388)
			if iter.timestamp < lowestTimestamp {
				__antithesis_instrumentation__.Notify(648391)
				lowestTimestamp = iter.timestamp
			} else {
				__antithesis_instrumentation__.Notify(648392)
			}
		}
	}
	__antithesis_instrumentation__.Notify(648384)

	aggregateValues := make([]float64, len(iterators))
	for computeLowest(); lowestTimestamp <= timespan.EndNanos; computeLowest() {
		__antithesis_instrumentation__.Notify(648393)
		aggregateValues = aggregateValues[:0]
		for i, iter := range iterators {
			__antithesis_instrumentation__.Notify(648397)
			var value float64
			var valid bool
			switch query.GetDerivative() {
			case tspb.TimeSeriesQueryDerivative_DERIVATIVE:
				__antithesis_instrumentation__.Notify(648400)
				valid = iter.validAtTimestamp(lowestTimestamp, interpolationLimitNanos)
				if valid {
					__antithesis_instrumentation__.Notify(648403)
					value, valid = iter.derivative(query.GetDownsampler())

					value *= float64(time.Second.Nanoseconds()) / float64(iter.samplePeriod())
				} else {
					__antithesis_instrumentation__.Notify(648404)
				}
			case tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE:
				__antithesis_instrumentation__.Notify(648401)
				valid = iter.validAtTimestamp(lowestTimestamp, interpolationLimitNanos)
				if valid {
					__antithesis_instrumentation__.Notify(648405)
					value, valid = iter.derivative(query.GetDownsampler())
					if value < 0 {
						__antithesis_instrumentation__.Notify(648406)
						value = 0
					} else {
						__antithesis_instrumentation__.Notify(648407)

						value *= float64(time.Second.Nanoseconds()) / float64(iter.samplePeriod())
					}
				} else {
					__antithesis_instrumentation__.Notify(648408)
				}
			default:
				__antithesis_instrumentation__.Notify(648402)
				value, valid = iter.valueAtTimestamp(
					lowestTimestamp, interpolationLimitNanos, query.GetDownsampler(),
				)
			}
			__antithesis_instrumentation__.Notify(648398)

			if valid {
				__antithesis_instrumentation__.Notify(648409)
				aggregateValues = append(aggregateValues, value)
			} else {
				__antithesis_instrumentation__.Notify(648410)
			}
			__antithesis_instrumentation__.Notify(648399)
			if iter.timestamp == lowestTimestamp {
				__antithesis_instrumentation__.Notify(648411)
				iterators[i].forward()
			} else {
				__antithesis_instrumentation__.Notify(648412)
			}
		}
		__antithesis_instrumentation__.Notify(648394)
		if len(aggregateValues) == 0 {
			__antithesis_instrumentation__.Notify(648413)
			continue
		} else {
			__antithesis_instrumentation__.Notify(648414)
		}
		__antithesis_instrumentation__.Notify(648395)

		if lowestTimestamp > timespan.NowNanos-timespan.SampleDurationNanos {
			__antithesis_instrumentation__.Notify(648415)
			if len(aggregateValues) < len(iterators) {
				__antithesis_instrumentation__.Notify(648416)
				continue
			} else {
				__antithesis_instrumentation__.Notify(648417)
			}
		} else {
			__antithesis_instrumentation__.Notify(648418)
		}
		__antithesis_instrumentation__.Notify(648396)

		*dest = append(*dest, tspb.TimeSeriesDatapoint{
			TimestampNanos: lowestTimestamp,
			Value:          aggregate(query.GetSourceAggregator(), aggregateValues),
		})
	}
}

func aggSum(data []float64) float64 {
	__antithesis_instrumentation__.Notify(648419)
	total := 0.0
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648421)
		total += dp
	}
	__antithesis_instrumentation__.Notify(648420)
	return total
}

func aggAvg(data []float64) float64 {
	__antithesis_instrumentation__.Notify(648422)
	if len(data) == 0 {
		__antithesis_instrumentation__.Notify(648424)
		return 0.0
	} else {
		__antithesis_instrumentation__.Notify(648425)
	}
	__antithesis_instrumentation__.Notify(648423)
	return aggSum(data) / float64(len(data))
}

func aggMax(data []float64) float64 {
	__antithesis_instrumentation__.Notify(648426)
	max := -math.MaxFloat64
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648428)
		if dp > max {
			__antithesis_instrumentation__.Notify(648429)
			max = dp
		} else {
			__antithesis_instrumentation__.Notify(648430)
		}
	}
	__antithesis_instrumentation__.Notify(648427)
	return max
}

func aggMin(data []float64) float64 {
	__antithesis_instrumentation__.Notify(648431)
	min := math.MaxFloat64
	for _, dp := range data {
		__antithesis_instrumentation__.Notify(648433)
		if dp < min {
			__antithesis_instrumentation__.Notify(648434)
			min = dp
		} else {
			__antithesis_instrumentation__.Notify(648435)
		}
	}
	__antithesis_instrumentation__.Notify(648432)
	return min
}

func aggregate(agg tspb.TimeSeriesQueryAggregator, values []float64) float64 {
	__antithesis_instrumentation__.Notify(648436)
	switch agg {
	case tspb.TimeSeriesQueryAggregator_AVG:
		__antithesis_instrumentation__.Notify(648438)
		return aggAvg(values)
	case tspb.TimeSeriesQueryAggregator_SUM:
		__antithesis_instrumentation__.Notify(648439)
		return aggSum(values)
	case tspb.TimeSeriesQueryAggregator_MAX:
		__antithesis_instrumentation__.Notify(648440)
		return aggMax(values)
	case tspb.TimeSeriesQueryAggregator_MIN:
		__antithesis_instrumentation__.Notify(648441)
		return aggMin(values)
	default:
		__antithesis_instrumentation__.Notify(648442)
	}
	__antithesis_instrumentation__.Notify(648437)

	panic(fmt.Sprintf("unknown aggregator option encountered: %v", agg))
}

func (db *DB) readFromDatabase(
	ctx context.Context,
	seriesName string,
	diskResolution Resolution,
	timespan QueryTimespan,
	sources []string,
) ([]kv.KeyValue, error) {
	__antithesis_instrumentation__.Notify(648443)

	b := &kv.Batch{}
	startTimestamp := diskResolution.normalizeToSlab(timespan.StartNanos)
	kd := diskResolution.SlabDuration()
	for currentTimestamp := startTimestamp; currentTimestamp <= timespan.EndNanos; currentTimestamp += kd {
		__antithesis_instrumentation__.Notify(648447)
		for _, source := range sources {
			__antithesis_instrumentation__.Notify(648448)
			key := MakeDataKey(seriesName, source, diskResolution, currentTimestamp)
			b.Get(key)
		}
	}
	__antithesis_instrumentation__.Notify(648444)
	if err := db.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(648449)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(648450)
	}
	__antithesis_instrumentation__.Notify(648445)
	var rows []kv.KeyValue
	for _, result := range b.Results {
		__antithesis_instrumentation__.Notify(648451)
		row := result.Rows[0]
		if row.Value == nil {
			__antithesis_instrumentation__.Notify(648453)
			continue
		} else {
			__antithesis_instrumentation__.Notify(648454)
		}
		__antithesis_instrumentation__.Notify(648452)
		rows = append(rows, row)
	}
	__antithesis_instrumentation__.Notify(648446)
	return rows, nil
}

func (db *DB) readAllSourcesFromDatabase(
	ctx context.Context, seriesName string, diskResolution Resolution, timespan QueryTimespan,
) ([]kv.KeyValue, error) {
	__antithesis_instrumentation__.Notify(648455)

	startKey := MakeDataKey(
		seriesName, "", diskResolution, timespan.StartNanos,
	)
	endKey := MakeDataKey(
		seriesName, "", diskResolution, timespan.EndNanos,
	).PrefixEnd()
	b := &kv.Batch{}
	b.Scan(startKey, endKey)

	if err := db.db.Run(ctx, b); err != nil {
		__antithesis_instrumentation__.Notify(648457)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(648458)
	}
	__antithesis_instrumentation__.Notify(648456)
	return b.Results[0].Rows, nil
}

func convertKeysToSpans(
	ctx context.Context, data []kv.KeyValue, acc *mon.BoundAccount,
) (map[string]timeSeriesSpan, error) {
	__antithesis_instrumentation__.Notify(648459)
	sourceSpans := make(map[string]timeSeriesSpan)
	for _, row := range data {
		__antithesis_instrumentation__.Notify(648461)
		var data roachpb.InternalTimeSeriesData
		if err := row.ValueProto(&data); err != nil {
			__antithesis_instrumentation__.Notify(648466)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(648467)
		}
		__antithesis_instrumentation__.Notify(648462)
		_, source, _, _, err := DecodeDataKey(row.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(648468)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(648469)
		}
		__antithesis_instrumentation__.Notify(648463)
		sampleSize := sizeOfSample
		if data.IsColumnar() {
			__antithesis_instrumentation__.Notify(648470)
			sampleSize = sizeOfInt32 + sizeOfFloat64
		} else {
			__antithesis_instrumentation__.Notify(648471)
		}
		__antithesis_instrumentation__.Notify(648464)
		if err := acc.Grow(
			ctx, sampleSize*int64(data.SampleCount())+sizeOfTimeSeriesData,
		); err != nil {
			__antithesis_instrumentation__.Notify(648472)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(648473)
		}
		__antithesis_instrumentation__.Notify(648465)
		sourceSpans[source] = append(sourceSpans[source], data)
	}
	__antithesis_instrumentation__.Notify(648460)
	return sourceSpans, nil
}

func verifySourceAggregator(agg tspb.TimeSeriesQueryAggregator) error {
	__antithesis_instrumentation__.Notify(648474)
	switch agg {
	case tspb.TimeSeriesQueryAggregator_AVG:
		__antithesis_instrumentation__.Notify(648476)
		return nil
	case tspb.TimeSeriesQueryAggregator_SUM:
		__antithesis_instrumentation__.Notify(648477)
		return nil
	case tspb.TimeSeriesQueryAggregator_MIN:
		__antithesis_instrumentation__.Notify(648478)
		return nil
	case tspb.TimeSeriesQueryAggregator_MAX:
		__antithesis_instrumentation__.Notify(648479)
		return nil
	case tspb.TimeSeriesQueryAggregator_FIRST,
		tspb.TimeSeriesQueryAggregator_LAST,
		tspb.TimeSeriesQueryAggregator_VARIANCE:
		__antithesis_instrumentation__.Notify(648480)
		return errors.Errorf("aggregator %s is not yet supported", agg.String())
	default:
		__antithesis_instrumentation__.Notify(648481)
	}
	__antithesis_instrumentation__.Notify(648475)
	return errors.Errorf("query specified unknown time series aggregator %s", agg.String())
}

func verifyDownsampler(downsampler tspb.TimeSeriesQueryAggregator) error {
	__antithesis_instrumentation__.Notify(648482)
	switch downsampler {
	case tspb.TimeSeriesQueryAggregator_AVG:
		__antithesis_instrumentation__.Notify(648484)
		return nil
	case tspb.TimeSeriesQueryAggregator_SUM:
		__antithesis_instrumentation__.Notify(648485)
		return nil
	case tspb.TimeSeriesQueryAggregator_MIN:
		__antithesis_instrumentation__.Notify(648486)
		return nil
	case tspb.TimeSeriesQueryAggregator_MAX:
		__antithesis_instrumentation__.Notify(648487)
		return nil
	case tspb.TimeSeriesQueryAggregator_FIRST,
		tspb.TimeSeriesQueryAggregator_LAST,
		tspb.TimeSeriesQueryAggregator_VARIANCE:
		__antithesis_instrumentation__.Notify(648488)
		return errors.Errorf("downsampler %s is not yet supported", downsampler.String())
	default:
		__antithesis_instrumentation__.Notify(648489)
	}
	__antithesis_instrumentation__.Notify(648483)
	return errors.Errorf("query specified unknown time series downsampler %s", downsampler.String())
}
