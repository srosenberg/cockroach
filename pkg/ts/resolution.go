package ts

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
)

type Resolution int64

func (r Resolution) String() string {
	__antithesis_instrumentation__.Notify(648490)
	switch r {
	case Resolution10s:
		__antithesis_instrumentation__.Notify(648492)
		return "10s"
	case Resolution30m:
		__antithesis_instrumentation__.Notify(648493)
		return "30m"
	case resolution1ns:
		__antithesis_instrumentation__.Notify(648494)
		return "1ns"
	case resolution50ns:
		__antithesis_instrumentation__.Notify(648495)
		return "50ns"
	case resolutionInvalid:
		__antithesis_instrumentation__.Notify(648496)
		return "BAD"
	default:
		__antithesis_instrumentation__.Notify(648497)
	}
	__antithesis_instrumentation__.Notify(648491)
	return fmt.Sprintf("%d", r)
}

const (
	Resolution10s Resolution = 1

	Resolution30m Resolution = 2

	resolution1ns Resolution = 998

	resolution50ns Resolution = 999

	resolutionInvalid Resolution = 1000
)

var sampleDurationByResolution = map[Resolution]int64{
	Resolution10s:     int64(time.Second * 10),
	Resolution30m:     int64(time.Minute * 30),
	resolution1ns:     1,
	resolution50ns:    50,
	resolutionInvalid: 10,
}

var slabDurationByResolution = map[Resolution]int64{
	Resolution10s:     int64(time.Hour),
	Resolution30m:     int64(time.Hour * 24),
	resolution1ns:     10,
	resolution50ns:    1000,
	resolutionInvalid: 11,
}

func (r Resolution) SampleDuration() int64 {
	__antithesis_instrumentation__.Notify(648498)
	duration, ok := sampleDurationByResolution[r]
	if !ok {
		__antithesis_instrumentation__.Notify(648500)
		panic(fmt.Sprintf("no sample duration found for resolution value %v", r))
	} else {
		__antithesis_instrumentation__.Notify(648501)
	}
	__antithesis_instrumentation__.Notify(648499)
	return duration
}

func (r Resolution) SlabDuration() int64 {
	__antithesis_instrumentation__.Notify(648502)
	duration, ok := slabDurationByResolution[r]
	if !ok {
		__antithesis_instrumentation__.Notify(648504)
		panic(fmt.Sprintf("no slab duration found for resolution value %v", r))
	} else {
		__antithesis_instrumentation__.Notify(648505)
	}
	__antithesis_instrumentation__.Notify(648503)
	return duration
}

func (r Resolution) IsRollup() bool {
	__antithesis_instrumentation__.Notify(648506)
	return r == Resolution30m || func() bool {
		__antithesis_instrumentation__.Notify(648507)
		return r == resolution50ns == true
	}() == true
}

func (r Resolution) TargetRollupResolution() (Resolution, bool) {
	__antithesis_instrumentation__.Notify(648508)
	switch r {
	case Resolution10s:
		__antithesis_instrumentation__.Notify(648510)
		return Resolution30m, true
	case resolution1ns:
		__antithesis_instrumentation__.Notify(648511)
		return resolution50ns, true
	default:
		__antithesis_instrumentation__.Notify(648512)
	}
	__antithesis_instrumentation__.Notify(648509)
	return r, false
}

func normalizeToPeriod(timestampNanos int64, period int64) int64 {
	__antithesis_instrumentation__.Notify(648513)
	return timestampNanos - timestampNanos%period
}

func (r Resolution) normalizeToSlab(timestampNanos int64) int64 {
	__antithesis_instrumentation__.Notify(648514)
	return normalizeToPeriod(timestampNanos, r.SlabDuration())
}

func ResolutionFromProto(r tspb.TimeSeriesResolution) Resolution {
	__antithesis_instrumentation__.Notify(648515)
	switch r {
	case tspb.TimeSeriesResolution_RESOLUTION_10S:
		__antithesis_instrumentation__.Notify(648517)
		return Resolution10s
	case tspb.TimeSeriesResolution_RESOLUTION_30M:
		__antithesis_instrumentation__.Notify(648518)
		return Resolution30m
	default:
		__antithesis_instrumentation__.Notify(648519)
	}
	__antithesis_instrumentation__.Notify(648516)
	return resolutionInvalid
}
