package split

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

const (
	RecordDurationThreshold    = 10 * time.Second
	splitKeySampleSize         = 20
	splitKeyMinCounter         = 100
	splitKeyThreshold          = 0.25
	splitKeyContainedThreshold = 0.50
)

type sample struct {
	key                    roachpb.Key
	left, right, contained int
}

type Finder struct {
	startTime time.Time
	samples   [splitKeySampleSize]sample
	count     int
}

func NewFinder(startTime time.Time) *Finder {
	__antithesis_instrumentation__.Notify(123416)
	return &Finder{
		startTime: startTime,
	}
}

func (f *Finder) Ready(nowTime time.Time) bool {
	__antithesis_instrumentation__.Notify(123417)
	return nowTime.Sub(f.startTime) > RecordDurationThreshold
}

func (f *Finder) Record(span roachpb.Span, intNFn func(int) int) {
	__antithesis_instrumentation__.Notify(123418)
	if f == nil {
		__antithesis_instrumentation__.Notify(123421)
		return
	} else {
		__antithesis_instrumentation__.Notify(123422)
	}
	__antithesis_instrumentation__.Notify(123419)

	var idx int
	count := f.count
	f.count++
	if count < splitKeySampleSize {
		__antithesis_instrumentation__.Notify(123423)
		idx = count
	} else {
		__antithesis_instrumentation__.Notify(123424)
		if idx = intNFn(count); idx >= splitKeySampleSize {
			__antithesis_instrumentation__.Notify(123425)

			for i := range f.samples {
				__antithesis_instrumentation__.Notify(123427)
				if span.ProperlyContainsKey(f.samples[i].key) {
					__antithesis_instrumentation__.Notify(123428)
					f.samples[i].contained++
				} else {
					__antithesis_instrumentation__.Notify(123429)

					if comp := bytes.Compare(f.samples[i].key, span.Key); comp <= 0 {
						__antithesis_instrumentation__.Notify(123430)
						f.samples[i].right++
					} else {
						__antithesis_instrumentation__.Notify(123431)
						if comp > 0 {
							__antithesis_instrumentation__.Notify(123432)
							f.samples[i].left++
						} else {
							__antithesis_instrumentation__.Notify(123433)
						}
					}
				}
			}
			__antithesis_instrumentation__.Notify(123426)
			return
		} else {
			__antithesis_instrumentation__.Notify(123434)
		}
	}
	__antithesis_instrumentation__.Notify(123420)

	f.samples[idx] = sample{key: span.Key}
}

func (f *Finder) Key() roachpb.Key {
	__antithesis_instrumentation__.Notify(123435)
	if f == nil {
		__antithesis_instrumentation__.Notify(123439)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(123440)
	}
	__antithesis_instrumentation__.Notify(123436)

	var bestIdx = -1
	var bestScore float64 = 2
	for i, s := range f.samples {
		__antithesis_instrumentation__.Notify(123441)
		if s.left+s.right+s.contained < splitKeyMinCounter {
			__antithesis_instrumentation__.Notify(123444)
			continue
		} else {
			__antithesis_instrumentation__.Notify(123445)
		}
		__antithesis_instrumentation__.Notify(123442)
		balanceScore := math.Abs(float64(s.left-s.right)) / float64(s.left+s.right)
		containedScore := (float64(s.contained) / float64(s.left+s.right+s.contained))
		finalScore := balanceScore + containedScore
		if balanceScore >= splitKeyThreshold || func() bool {
			__antithesis_instrumentation__.Notify(123446)
			return containedScore >= splitKeyContainedThreshold == true
		}() == true {
			__antithesis_instrumentation__.Notify(123447)
			continue
		} else {
			__antithesis_instrumentation__.Notify(123448)
		}
		__antithesis_instrumentation__.Notify(123443)
		if finalScore < bestScore {
			__antithesis_instrumentation__.Notify(123449)
			bestIdx = i
			bestScore = finalScore
		} else {
			__antithesis_instrumentation__.Notify(123450)
		}
	}
	__antithesis_instrumentation__.Notify(123437)

	if bestIdx == -1 {
		__antithesis_instrumentation__.Notify(123451)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(123452)
	}
	__antithesis_instrumentation__.Notify(123438)
	return f.samples[bestIdx].key
}
