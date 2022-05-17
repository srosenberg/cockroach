package tscache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type sklImpl struct {
	cache   *intervalSkl
	clock   *hlc.Clock
	metrics Metrics
}

var _ Cache = &sklImpl{}

func newSklImpl(clock *hlc.Clock) *sklImpl {
	__antithesis_instrumentation__.Notify(127106)
	tc := sklImpl{clock: clock, metrics: makeMetrics()}
	tc.clear(clock.Now())
	return &tc
}

func (tc *sklImpl) clear(lowWater hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(127107)
	tc.cache = newIntervalSkl(tc.clock, MinRetentionWindow, tc.metrics.Skl)
	tc.cache.floorTS = lowWater
}

func (tc *sklImpl) Add(start, end roachpb.Key, ts hlc.Timestamp, txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(127108)
	start, end = tc.boundKeyLengths(start, end)

	val := cacheValue{ts: ts, txnID: txnID}
	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(127109)
		tc.cache.Add(nonNil(start), val)
	} else {
		__antithesis_instrumentation__.Notify(127110)
		tc.cache.AddRange(nonNil(start), end, excludeTo, val)
	}
}

func (tc *sklImpl) getLowWater() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127111)
	return tc.cache.FloorTS()
}

func (tc *sklImpl) GetMax(start, end roachpb.Key) (hlc.Timestamp, uuid.UUID) {
	__antithesis_instrumentation__.Notify(127112)
	var val cacheValue
	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(127114)
		val = tc.cache.LookupTimestamp(nonNil(start))
	} else {
		__antithesis_instrumentation__.Notify(127115)
		val = tc.cache.LookupTimestampRange(nonNil(start), end, excludeTo)
	}
	__antithesis_instrumentation__.Notify(127113)
	return val.ts, val.txnID
}

func (tc *sklImpl) boundKeyLengths(start, end roachpb.Key) (roachpb.Key, roachpb.Key) {
	__antithesis_instrumentation__.Notify(127116)

	maxKeySize := int(maximumSklPageSize / 32)

	if l := len(start); l > maxKeySize {
		__antithesis_instrumentation__.Notify(127119)
		start = start[:maxKeySize]
		log.Warningf(context.TODO(), "start key with length %d exceeds maximum key length of %d; "+
			"losing precision in timestamp cache", l, maxKeySize)
	} else {
		__antithesis_instrumentation__.Notify(127120)
	}
	__antithesis_instrumentation__.Notify(127117)
	if l := len(end); l > maxKeySize {
		__antithesis_instrumentation__.Notify(127121)
		end = end[:maxKeySize].PrefixEnd()
		log.Warningf(context.TODO(), "end key with length %d exceeds maximum key length of %d; "+
			"losing precision in timestamp cache", l, maxKeySize)
	} else {
		__antithesis_instrumentation__.Notify(127122)
	}
	__antithesis_instrumentation__.Notify(127118)
	return start, end
}

func (tc *sklImpl) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(127123)
	return tc.metrics
}

var emptyStartKey = []byte("")

func nonNil(b []byte) []byte {
	__antithesis_instrumentation__.Notify(127124)
	if b == nil {
		__antithesis_instrumentation__.Notify(127126)
		return emptyStartKey
	} else {
		__antithesis_instrumentation__.Notify(127127)
	}
	__antithesis_instrumentation__.Notify(127125)
	return b
}
