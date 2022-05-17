package stats

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

const DefaultHistogramBuckets = 200

var HistogramClusterMode = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"sql.stats.histogram_collection.enabled",
	"histogram collection mode",
	true,
).WithPublic()

type HistogramVersion uint32

const histVersion HistogramVersion = 1

func EquiDepthHistogram(
	evalCtx *tree.EvalContext,
	colType *types.T,
	samples tree.Datums,
	numRows, distinctCount int64,
	maxBuckets int,
) (HistogramData, []cat.HistogramBucket, error) {
	__antithesis_instrumentation__.Notify(626161)
	numSamples := len(samples)
	if numSamples == 0 {
		__antithesis_instrumentation__.Notify(626170)
		return HistogramData{ColumnType: colType}, nil, nil
	} else {
		__antithesis_instrumentation__.Notify(626171)
	}
	__antithesis_instrumentation__.Notify(626162)
	if maxBuckets < 2 {
		__antithesis_instrumentation__.Notify(626172)
		return HistogramData{}, nil, errors.Errorf("histogram requires at least two buckets")
	} else {
		__antithesis_instrumentation__.Notify(626173)
	}
	__antithesis_instrumentation__.Notify(626163)
	if numRows < int64(numSamples) {
		__antithesis_instrumentation__.Notify(626174)
		return HistogramData{}, nil, errors.Errorf("more samples than rows")
	} else {
		__antithesis_instrumentation__.Notify(626175)
	}
	__antithesis_instrumentation__.Notify(626164)
	if distinctCount == 0 {
		__antithesis_instrumentation__.Notify(626176)
		return HistogramData{}, nil, errors.Errorf("histogram requires distinctCount > 0")
	} else {
		__antithesis_instrumentation__.Notify(626177)
	}
	__antithesis_instrumentation__.Notify(626165)
	for _, d := range samples {
		__antithesis_instrumentation__.Notify(626178)
		if d == tree.DNull {
			__antithesis_instrumentation__.Notify(626179)
			return HistogramData{}, nil, errors.Errorf("NULL values not allowed in histogram")
		} else {
			__antithesis_instrumentation__.Notify(626180)
		}
	}
	__antithesis_instrumentation__.Notify(626166)

	sort.Slice(samples, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(626181)
		return samples[i].Compare(evalCtx, samples[j]) < 0
	})
	__antithesis_instrumentation__.Notify(626167)
	numBuckets := maxBuckets
	if maxBuckets > numSamples {
		__antithesis_instrumentation__.Notify(626182)
		numBuckets = numSamples
	} else {
		__antithesis_instrumentation__.Notify(626183)
	}
	__antithesis_instrumentation__.Notify(626168)
	h := histogram{buckets: make([]cat.HistogramBucket, 0, numBuckets)}
	lowerBound := samples[0]

	for i, b := 0, 0; b < numBuckets && func() bool {
		__antithesis_instrumentation__.Notify(626184)
		return i < numSamples == true
	}() == true; b++ {
		__antithesis_instrumentation__.Notify(626185)

		numSamplesInBucket := (numSamples - i) / (numBuckets - b)
		if i == 0 || func() bool {
			__antithesis_instrumentation__.Notify(626189)
			return numSamplesInBucket < 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(626190)
			numSamplesInBucket = 1
		} else {
			__antithesis_instrumentation__.Notify(626191)
		}
		__antithesis_instrumentation__.Notify(626186)
		upper := samples[i+numSamplesInBucket-1]

		numLess := 0
		for ; numLess < numSamplesInBucket-1; numLess++ {
			__antithesis_instrumentation__.Notify(626192)
			if c := samples[i+numLess].Compare(evalCtx, upper); c == 0 {
				__antithesis_instrumentation__.Notify(626193)
				break
			} else {
				__antithesis_instrumentation__.Notify(626194)
				if c > 0 {
					__antithesis_instrumentation__.Notify(626195)
					return HistogramData{}, nil, errors.AssertionFailedf("%+v", "samples not sorted")
				} else {
					__antithesis_instrumentation__.Notify(626196)
				}
			}
		}
		__antithesis_instrumentation__.Notify(626187)

		for ; i+numSamplesInBucket < numSamples; numSamplesInBucket++ {
			__antithesis_instrumentation__.Notify(626197)
			if samples[i+numSamplesInBucket].Compare(evalCtx, upper) != 0 {
				__antithesis_instrumentation__.Notify(626198)
				break
			} else {
				__antithesis_instrumentation__.Notify(626199)
			}
		}
		__antithesis_instrumentation__.Notify(626188)

		numEq := float64(numSamplesInBucket-numLess) * float64(numRows) / float64(numSamples)
		numRange := float64(numLess) * float64(numRows) / float64(numSamples)
		distinctRange := estimatedDistinctValuesInRange(evalCtx, numRange, lowerBound, upper)

		i += numSamplesInBucket
		h.buckets = append(h.buckets, cat.HistogramBucket{
			NumEq:         numEq,
			NumRange:      numRange,
			DistinctRange: distinctRange,
			UpperBound:    upper,
		})

		lowerBound = getNextLowerBound(evalCtx, upper)
	}
	__antithesis_instrumentation__.Notify(626169)

	h.adjustCounts(evalCtx, float64(numRows), float64(distinctCount))
	histogramData, err := h.toHistogramData(colType)
	return histogramData, h.buckets, err
}

type histogram struct {
	buckets []cat.HistogramBucket
}

func (h *histogram) adjustCounts(
	evalCtx *tree.EvalContext, rowCountTotal, distinctCountTotal float64,
) {
	__antithesis_instrumentation__.Notify(626200)

	var rowCountRange, rowCountEq float64

	var distinctCountRange float64

	var distinctCountEq float64
	for i := range h.buckets {
		__antithesis_instrumentation__.Notify(626208)
		rowCountRange += h.buckets[i].NumRange
		rowCountEq += h.buckets[i].NumEq
		distinctCountRange += h.buckets[i].DistinctRange
		if h.buckets[i].NumEq > 0 {
			__antithesis_instrumentation__.Notify(626209)
			distinctCountEq++
		} else {
			__antithesis_instrumentation__.Notify(626210)
		}
	}
	__antithesis_instrumentation__.Notify(626201)

	if rowCountEq <= 0 {
		__antithesis_instrumentation__.Notify(626211)
		panic(errors.AssertionFailedf("expected a positive value for rowCountEq"))
	} else {
		__antithesis_instrumentation__.Notify(626212)
	}
	__antithesis_instrumentation__.Notify(626202)

	if distinctCountEq >= distinctCountTotal {
		__antithesis_instrumentation__.Notify(626213)
		adjustmentFactorNumEq := rowCountTotal / rowCountEq
		for i := range h.buckets {
			__antithesis_instrumentation__.Notify(626215)
			h.buckets[i].NumRange = 0
			h.buckets[i].DistinctRange = 0
			h.buckets[i].NumEq *= adjustmentFactorNumEq
		}
		__antithesis_instrumentation__.Notify(626214)
		return
	} else {
		__antithesis_instrumentation__.Notify(626216)
	}
	__antithesis_instrumentation__.Notify(626203)

	remDistinctCount := distinctCountTotal - distinctCountEq
	if rowCountEq+remDistinctCount >= rowCountTotal {
		__antithesis_instrumentation__.Notify(626217)
		targetRowCountEq := rowCountTotal - remDistinctCount
		adjustmentFactorNumEq := targetRowCountEq / rowCountEq
		for i := range h.buckets {
			__antithesis_instrumentation__.Notify(626219)
			h.buckets[i].NumEq *= adjustmentFactorNumEq
		}
		__antithesis_instrumentation__.Notify(626218)
		rowCountEq = targetRowCountEq
	} else {
		__antithesis_instrumentation__.Notify(626220)
	}
	__antithesis_instrumentation__.Notify(626204)

	if remDistinctCount > distinctCountRange {
		__antithesis_instrumentation__.Notify(626221)
		remDistinctCount -= distinctCountRange

		maxDistinctCountRange := float64(math.MaxInt64)
		lowerBound := h.buckets[0].UpperBound
		upperBound := h.buckets[len(h.buckets)-1].UpperBound
		if maxDistinct, ok := tree.MaxDistinctCount(evalCtx, lowerBound, upperBound); ok {
			__antithesis_instrumentation__.Notify(626223)

			maxDistinctCountRange = float64(maxDistinct) - distinctCountEq - distinctCountRange
		} else {
			__antithesis_instrumentation__.Notify(626224)
		}
		__antithesis_instrumentation__.Notify(626222)

		if maxDistinctCountRange > 0 {
			__antithesis_instrumentation__.Notify(626225)
			if remDistinctCount > maxDistinctCountRange {
				__antithesis_instrumentation__.Notify(626227)

				remDistinctCount = maxDistinctCountRange
			} else {
				__antithesis_instrumentation__.Notify(626228)
			}
			__antithesis_instrumentation__.Notify(626226)
			avgRemPerBucket := remDistinctCount / float64(len(h.buckets)-1)
			for i := 1; i < len(h.buckets); i++ {
				__antithesis_instrumentation__.Notify(626229)
				lowerBound := h.buckets[i-1].UpperBound
				upperBound := h.buckets[i].UpperBound
				maxDistRange, countable := maxDistinctRange(evalCtx, lowerBound, upperBound)

				inc := avgRemPerBucket
				if countable {
					__antithesis_instrumentation__.Notify(626231)
					maxDistRange -= h.buckets[i].DistinctRange

					inc = remDistinctCount * (maxDistRange / maxDistinctCountRange)
				} else {
					__antithesis_instrumentation__.Notify(626232)
				}
				__antithesis_instrumentation__.Notify(626230)

				h.buckets[i].NumRange += inc
				h.buckets[i].DistinctRange += inc
				rowCountRange += inc
				distinctCountRange += inc
			}
		} else {
			__antithesis_instrumentation__.Notify(626233)
		}
	} else {
		__antithesis_instrumentation__.Notify(626234)
	}
	__antithesis_instrumentation__.Notify(626205)

	remDistinctCount = distinctCountTotal - distinctCountRange - distinctCountEq
	if remDistinctCount > 0 {
		__antithesis_instrumentation__.Notify(626235)
		h.addOuterBuckets(
			evalCtx, remDistinctCount, &rowCountEq, &distinctCountEq, &rowCountRange, &distinctCountRange,
		)
	} else {
		__antithesis_instrumentation__.Notify(626236)
	}
	__antithesis_instrumentation__.Notify(626206)

	adjustmentFactorDistinctRange := float64(1)
	if distinctCountRange > 0 {
		__antithesis_instrumentation__.Notify(626237)
		adjustmentFactorDistinctRange = (distinctCountTotal - distinctCountEq) / distinctCountRange
	} else {
		__antithesis_instrumentation__.Notify(626238)
	}
	__antithesis_instrumentation__.Notify(626207)
	adjustmentFactorRowCount := rowCountTotal / (rowCountRange + rowCountEq)
	for i := range h.buckets {
		__antithesis_instrumentation__.Notify(626239)
		h.buckets[i].DistinctRange *= adjustmentFactorDistinctRange
		h.buckets[i].NumRange *= adjustmentFactorRowCount
		h.buckets[i].NumEq *= adjustmentFactorRowCount
	}
}

func (h *histogram) addOuterBuckets(
	evalCtx *tree.EvalContext,
	remDistinctCount float64,
	rowCountEq, distinctCountEq, rowCountRange, distinctCountRange *float64,
) {
	__antithesis_instrumentation__.Notify(626240)
	var maxDistinctCountExtraBuckets float64
	var addedMin, addedMax bool
	var newBuckets int
	if !h.buckets[0].UpperBound.IsMin(evalCtx) {
		__antithesis_instrumentation__.Notify(626248)
		if minVal, ok := h.buckets[0].UpperBound.Min(evalCtx); ok {
			__antithesis_instrumentation__.Notify(626249)
			lowerBound := minVal
			upperBound := h.buckets[0].UpperBound
			maxDistRange, _ := maxDistinctRange(evalCtx, lowerBound, upperBound)
			maxDistinctCountExtraBuckets += maxDistRange
			h.buckets = append([]cat.HistogramBucket{{UpperBound: minVal}}, h.buckets...)
			addedMin = true
			newBuckets++
		} else {
			__antithesis_instrumentation__.Notify(626250)
		}
	} else {
		__antithesis_instrumentation__.Notify(626251)
	}
	__antithesis_instrumentation__.Notify(626241)
	if !h.buckets[len(h.buckets)-1].UpperBound.IsMax(evalCtx) {
		__antithesis_instrumentation__.Notify(626252)
		if maxVal, ok := h.buckets[len(h.buckets)-1].UpperBound.Max(evalCtx); ok {
			__antithesis_instrumentation__.Notify(626253)
			lowerBound := h.buckets[len(h.buckets)-1].UpperBound
			upperBound := maxVal
			maxDistRange, _ := maxDistinctRange(evalCtx, lowerBound, upperBound)
			maxDistinctCountExtraBuckets += maxDistRange
			h.buckets = append(h.buckets, cat.HistogramBucket{UpperBound: maxVal})
			addedMax = true
			newBuckets++
		} else {
			__antithesis_instrumentation__.Notify(626254)
		}
	} else {
		__antithesis_instrumentation__.Notify(626255)
	}
	__antithesis_instrumentation__.Notify(626242)

	if newBuckets == 0 {
		__antithesis_instrumentation__.Notify(626256)

		return
	} else {
		__antithesis_instrumentation__.Notify(626257)
	}
	__antithesis_instrumentation__.Notify(626243)

	if typFam := h.buckets[0].UpperBound.ResolvedType().Family(); typFam == types.EnumFamily || func() bool {
		__antithesis_instrumentation__.Notify(626258)
		return typFam == types.BoolFamily == true
	}() == true {
		__antithesis_instrumentation__.Notify(626259)
		if addedMin {
			__antithesis_instrumentation__.Notify(626262)
			h.buckets[0].NumEq++
		} else {
			__antithesis_instrumentation__.Notify(626263)
		}
		__antithesis_instrumentation__.Notify(626260)
		if addedMax {
			__antithesis_instrumentation__.Notify(626264)
			h.buckets[len(h.buckets)-1].NumEq++
		} else {
			__antithesis_instrumentation__.Notify(626265)
		}
		__antithesis_instrumentation__.Notify(626261)
		*rowCountEq += float64(newBuckets)
		*distinctCountEq += float64(newBuckets)
		remDistinctCount -= float64(newBuckets)
	} else {
		__antithesis_instrumentation__.Notify(626266)
	}
	__antithesis_instrumentation__.Notify(626244)

	if remDistinctCount <= 0 {
		__antithesis_instrumentation__.Notify(626267)

		return
	} else {
		__antithesis_instrumentation__.Notify(626268)
	}
	__antithesis_instrumentation__.Notify(626245)

	bucIndexes := make([]int, 0, newBuckets)
	if addedMin {
		__antithesis_instrumentation__.Notify(626269)

		bucIndexes = append(bucIndexes, 1)
	} else {
		__antithesis_instrumentation__.Notify(626270)
	}
	__antithesis_instrumentation__.Notify(626246)
	if addedMax {
		__antithesis_instrumentation__.Notify(626271)
		bucIndexes = append(bucIndexes, len(h.buckets)-1)
	} else {
		__antithesis_instrumentation__.Notify(626272)
	}
	__antithesis_instrumentation__.Notify(626247)
	avgRemPerBucket := remDistinctCount / float64(newBuckets)
	for _, i := range bucIndexes {
		__antithesis_instrumentation__.Notify(626273)
		lowerBound := h.buckets[i-1].UpperBound
		upperBound := h.buckets[i].UpperBound
		maxDistRange, countable := maxDistinctRange(evalCtx, lowerBound, upperBound)

		inc := avgRemPerBucket
		if countable && func() bool {
			__antithesis_instrumentation__.Notify(626275)
			return h.buckets[0].UpperBound.ResolvedType().Family() == types.EnumFamily == true
		}() == true {
			__antithesis_instrumentation__.Notify(626276)

			inc = remDistinctCount * (maxDistRange / maxDistinctCountExtraBuckets)
		} else {
			__antithesis_instrumentation__.Notify(626277)
		}
		__antithesis_instrumentation__.Notify(626274)

		h.buckets[i].NumRange += inc
		h.buckets[i].DistinctRange += inc
		*rowCountRange += inc
		*distinctCountRange += inc
	}
}

func (h histogram) toHistogramData(colType *types.T) (HistogramData, error) {
	__antithesis_instrumentation__.Notify(626278)
	histogramData := HistogramData{
		Buckets:    make([]HistogramData_Bucket, len(h.buckets)),
		ColumnType: colType,
		Version:    histVersion,
	}

	for i := range h.buckets {
		__antithesis_instrumentation__.Notify(626280)
		encoded, err := keyside.Encode(nil, h.buckets[i].UpperBound, encoding.Ascending)
		if err != nil {
			__antithesis_instrumentation__.Notify(626282)
			return HistogramData{}, err
		} else {
			__antithesis_instrumentation__.Notify(626283)
		}
		__antithesis_instrumentation__.Notify(626281)

		histogramData.Buckets[i] = HistogramData_Bucket{
			NumEq:         int64(math.Round(h.buckets[i].NumEq)),
			NumRange:      int64(math.Round(h.buckets[i].NumRange)),
			DistinctRange: h.buckets[i].DistinctRange,
			UpperBound:    encoded,
		}
	}
	__antithesis_instrumentation__.Notify(626279)

	return histogramData, nil
}

func estimatedDistinctValuesInRange(
	evalCtx *tree.EvalContext, numRange float64, lowerBound, upperBound tree.Datum,
) float64 {
	__antithesis_instrumentation__.Notify(626284)
	if numRange == 0 {
		__antithesis_instrumentation__.Notify(626288)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(626289)
	}
	__antithesis_instrumentation__.Notify(626285)
	rangeUpperBound, ok := upperBound.Prev(evalCtx)
	if !ok {
		__antithesis_instrumentation__.Notify(626290)
		rangeUpperBound = upperBound
	} else {
		__antithesis_instrumentation__.Notify(626291)
	}
	__antithesis_instrumentation__.Notify(626286)
	if maxDistinct, ok := tree.MaxDistinctCount(evalCtx, lowerBound, rangeUpperBound); ok {
		__antithesis_instrumentation__.Notify(626292)
		return expectedDistinctCount(numRange, float64(maxDistinct))
	} else {
		__antithesis_instrumentation__.Notify(626293)
	}
	__antithesis_instrumentation__.Notify(626287)
	return numRange
}

func getNextLowerBound(evalCtx *tree.EvalContext, currentUpperBound tree.Datum) tree.Datum {
	__antithesis_instrumentation__.Notify(626294)
	nextLowerBound, ok := currentUpperBound.Next(evalCtx)
	if !ok {
		__antithesis_instrumentation__.Notify(626296)
		nextLowerBound = currentUpperBound
	} else {
		__antithesis_instrumentation__.Notify(626297)
	}
	__antithesis_instrumentation__.Notify(626295)
	return nextLowerBound
}

func maxDistinctRange(
	evalCtx *tree.EvalContext, lowerBound, upperBound tree.Datum,
) (_ float64, countable bool) {
	__antithesis_instrumentation__.Notify(626298)
	if maxDistinct, ok := tree.MaxDistinctCount(evalCtx, lowerBound, upperBound); ok {
		__antithesis_instrumentation__.Notify(626300)

		if maxDistinct < 2 {
			__antithesis_instrumentation__.Notify(626302)
			return 0, true
		} else {
			__antithesis_instrumentation__.Notify(626303)
		}
		__antithesis_instrumentation__.Notify(626301)
		return float64(maxDistinct - 2), true
	} else {
		__antithesis_instrumentation__.Notify(626304)
	}
	__antithesis_instrumentation__.Notify(626299)
	return float64(math.MaxInt64), false
}

func expectedDistinctCount(k, n float64) float64 {
	__antithesis_instrumentation__.Notify(626305)
	if n == 0 || func() bool {
		__antithesis_instrumentation__.Notify(626308)
		return k == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(626309)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(626310)
	}
	__antithesis_instrumentation__.Notify(626306)

	count := n * (1 - math.Pow((n-1)/n, k))

	if count == 0 {
		__antithesis_instrumentation__.Notify(626311)
		count = k
		if n < k {
			__antithesis_instrumentation__.Notify(626312)
			count = n
		} else {
			__antithesis_instrumentation__.Notify(626313)
		}
	} else {
		__antithesis_instrumentation__.Notify(626314)
	}
	__antithesis_instrumentation__.Notify(626307)
	return count
}
