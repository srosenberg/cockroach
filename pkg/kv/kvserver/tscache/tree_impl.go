package tscache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const (
	defaultTreeImplSize = 64 << 20
)

func makeCacheEntry(key cache.IntervalKey, value cacheValue) *cache.Entry {
	__antithesis_instrumentation__.Notify(127128)
	alloc := struct {
		key   cache.IntervalKey
		value cacheValue
		entry cache.Entry
	}{
		key:   key,
		value: value,
	}
	alloc.entry.Key = &alloc.key
	alloc.entry.Value = &alloc.value
	return &alloc.entry
}

var cacheEntryOverhead = uint64(unsafe.Sizeof(cache.IntervalKey{}) +
	unsafe.Sizeof(cacheValue{}) + unsafe.Sizeof(cache.Entry{}))

func cacheEntrySize(start, end interval.Comparable) uint64 {
	__antithesis_instrumentation__.Notify(127129)
	n := uint64(cap(start))
	if end != nil && func() bool {
		__antithesis_instrumentation__.Notify(127131)
		return len(start) > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(127132)
		return len(end) > 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(127133)
		return &end[0] != &start[0] == true
	}() == true {
		__antithesis_instrumentation__.Notify(127134)

		n += uint64(cap(end))
	} else {
		__antithesis_instrumentation__.Notify(127135)
	}
	__antithesis_instrumentation__.Notify(127130)
	n += cacheEntryOverhead
	return n
}

type treeImpl struct {
	syncutil.RWMutex

	cache            *cache.IntervalCache
	lowWater, latest hlc.Timestamp

	bytes    uint64
	maxBytes uint64
	metrics  Metrics
}

var _ Cache = &treeImpl{}

func newTreeImpl(clock *hlc.Clock) *treeImpl {
	__antithesis_instrumentation__.Notify(127136)
	tc := &treeImpl{
		cache:    cache.NewIntervalCache(cache.Config{Policy: cache.CacheFIFO}),
		maxBytes: uint64(defaultTreeImplSize),
		metrics:  makeMetrics(),
	}
	tc.clear(clock.Now())
	tc.cache.Config.ShouldEvict = tc.shouldEvict
	tc.cache.Config.OnEvicted = tc.onEvicted
	return tc
}

func (tc *treeImpl) clear(lowWater hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(127137)
	tc.Lock()
	defer tc.Unlock()
	tc.cache.Clear()
	tc.lowWater = lowWater
	tc.latest = tc.lowWater
}

func (tc *treeImpl) len() int {
	__antithesis_instrumentation__.Notify(127138)
	tc.RLock()
	defer tc.RUnlock()
	return tc.cache.Len()
}

func (tc *treeImpl) Add(start, end roachpb.Key, ts hlc.Timestamp, txnID uuid.UUID) {
	__antithesis_instrumentation__.Notify(127139)

	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(127141)
		end = start.Next()
		start = end[:len(start)]
	} else {
		__antithesis_instrumentation__.Notify(127142)
	}
	__antithesis_instrumentation__.Notify(127140)

	tc.Lock()
	defer tc.Unlock()
	tc.latest.Forward(ts)

	if tc.lowWater.Less(ts) {
		__antithesis_instrumentation__.Notify(127143)

		addRange := func(r interval.Range) {
			__antithesis_instrumentation__.Notify(127147)
			value := cacheValue{ts: ts, txnID: txnID}
			key := tc.cache.MakeKey(r.Start, r.End)
			entry := makeCacheEntry(key, value)
			tc.bytes += cacheEntrySize(r.Start, r.End)
			tc.cache.AddEntry(entry)
		}
		__antithesis_instrumentation__.Notify(127144)
		addEntryAfter := func(entry, after *cache.Entry) {
			__antithesis_instrumentation__.Notify(127148)
			ck := entry.Key.(*cache.IntervalKey)
			tc.bytes += cacheEntrySize(ck.Start, ck.End)
			tc.cache.AddEntryAfter(entry, after)
		}
		__antithesis_instrumentation__.Notify(127145)

		r := interval.Range{
			Start: interval.Comparable(start),
			End:   interval.Comparable(end),
		}

		for _, entry := range tc.cache.GetOverlaps(r.Start, r.End) {
			__antithesis_instrumentation__.Notify(127149)
			cv := entry.Value.(*cacheValue)
			key := entry.Key.(*cache.IntervalKey)
			sCmp := r.Start.Compare(key.Start)
			eCmp := r.End.Compare(key.End)

			oldSize := cacheEntrySize(key.Start, key.End)
			if cv.ts.Less(ts) {
				__antithesis_instrumentation__.Notify(127151)

				switch {
				case sCmp == 0 && func() bool {
					__antithesis_instrumentation__.Notify(127158)
					return eCmp == 0 == true
				}() == true:
					__antithesis_instrumentation__.Notify(127152)

					*cv = cacheValue{ts: ts, txnID: txnID}
					tc.cache.MoveToEnd(entry)
					return
				case sCmp <= 0 && func() bool {
					__antithesis_instrumentation__.Notify(127159)
					return eCmp >= 0 == true
				}() == true:
					__antithesis_instrumentation__.Notify(127153)

					tc.cache.DelEntry(entry)
					continue
				case sCmp > 0 && func() bool {
					__antithesis_instrumentation__.Notify(127160)
					return eCmp < 0 == true
				}() == true:
					__antithesis_instrumentation__.Notify(127154)

					oldEnd := key.End
					key.End = r.Start

					newKey := tc.cache.MakeKey(r.End, oldEnd)
					newEntry := makeCacheEntry(newKey, *cv)
					addEntryAfter(newEntry, entry)
				case eCmp >= 0:
					__antithesis_instrumentation__.Notify(127155)

					key.End = r.Start
				case sCmp <= 0:
					__antithesis_instrumentation__.Notify(127156)

					key.Start = r.End
				default:
					__antithesis_instrumentation__.Notify(127157)
					panic(fmt.Sprintf("no overlap between %v and %v", key.Range, r))
				}
			} else {
				__antithesis_instrumentation__.Notify(127161)
				if ts.Less(cv.ts) {
					__antithesis_instrumentation__.Notify(127162)

					switch {
					case sCmp >= 0 && func() bool {
						__antithesis_instrumentation__.Notify(127168)
						return eCmp <= 0 == true
					}() == true:
						__antithesis_instrumentation__.Notify(127163)

						return
					case sCmp < 0 && func() bool {
						__antithesis_instrumentation__.Notify(127169)
						return eCmp > 0 == true
					}() == true:
						__antithesis_instrumentation__.Notify(127164)

						lr := interval.Range{Start: r.Start, End: key.Start}
						addRange(lr)

						r.Start = key.End
					case eCmp > 0:
						__antithesis_instrumentation__.Notify(127165)

						r.Start = key.End
					case sCmp < 0:
						__antithesis_instrumentation__.Notify(127166)

						r.End = key.Start
					default:
						__antithesis_instrumentation__.Notify(127167)
						panic(fmt.Sprintf("no overlap between %v and %v", key.Range, r))
					}
				} else {
					__antithesis_instrumentation__.Notify(127170)
					if cv.txnID == txnID {
						__antithesis_instrumentation__.Notify(127171)

						switch {
						case sCmp >= 0 && func() bool {
							__antithesis_instrumentation__.Notify(127177)
							return eCmp <= 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127172)

							return
						case sCmp <= 0 && func() bool {
							__antithesis_instrumentation__.Notify(127178)
							return eCmp >= 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127173)

							tc.cache.DelEntry(entry)
							continue
						case eCmp >= 0:
							__antithesis_instrumentation__.Notify(127174)

							key.End = r.Start
						case sCmp <= 0:
							__antithesis_instrumentation__.Notify(127175)

							key.Start = r.End
						default:
							__antithesis_instrumentation__.Notify(127176)
							panic(fmt.Sprintf("no overlap between %v and %v", key.Range, r))
						}
					} else {
						__antithesis_instrumentation__.Notify(127179)

						switch {
						case sCmp == 0 && func() bool {
							__antithesis_instrumentation__.Notify(127190)
							return eCmp == 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127180)

							cv.txnID = noTxnID
							tc.bytes += cacheEntrySize(key.Start, key.End) - oldSize
							return
						case sCmp == 0 && func() bool {
							__antithesis_instrumentation__.Notify(127191)
							return eCmp > 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127181)

							cv.txnID = noTxnID
							r.Start = key.End
						case sCmp < 0 && func() bool {
							__antithesis_instrumentation__.Notify(127192)
							return eCmp == 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127182)

							cv.txnID = noTxnID
							r.End = key.Start
						case sCmp < 0 && func() bool {
							__antithesis_instrumentation__.Notify(127193)
							return eCmp > 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127183)

							cv.txnID = noTxnID

							newKey := tc.cache.MakeKey(r.Start, key.Start)
							newEntry := makeCacheEntry(newKey, cacheValue{ts: ts, txnID: txnID})
							addEntryAfter(newEntry, entry)
							r.Start = key.End
						case sCmp > 0 && func() bool {
							__antithesis_instrumentation__.Notify(127194)
							return eCmp < 0 == true
						}() == true:
							__antithesis_instrumentation__.Notify(127184)

							txnID = noTxnID
							oldEnd := key.End
							key.End = r.Start

							newKey := tc.cache.MakeKey(r.End, oldEnd)
							newEntry := makeCacheEntry(newKey, *cv)
							addEntryAfter(newEntry, entry)
						case eCmp == 0:
							__antithesis_instrumentation__.Notify(127185)

							txnID = noTxnID
							key.End = r.Start
						case sCmp == 0:
							__antithesis_instrumentation__.Notify(127186)

							txnID = noTxnID
							key.Start = r.End
						case eCmp > 0:
							__antithesis_instrumentation__.Notify(127187)

							key.End, r.Start = r.Start, key.End

							newKey := tc.cache.MakeKey(key.End, r.Start)
							newCV := cacheValue{ts: cv.ts}
							newEntry := makeCacheEntry(newKey, newCV)
							addEntryAfter(newEntry, entry)
						case sCmp < 0:
							__antithesis_instrumentation__.Notify(127188)

							key.Start, r.End = r.End, key.Start

							newKey := tc.cache.MakeKey(r.End, key.Start)
							newCV := cacheValue{ts: cv.ts}
							newEntry := makeCacheEntry(newKey, newCV)
							addEntryAfter(newEntry, entry)
						default:
							__antithesis_instrumentation__.Notify(127189)
							panic(fmt.Sprintf("no overlap between %v and %v", key.Range, r))
						}
					}
				}
			}
			__antithesis_instrumentation__.Notify(127150)
			tc.bytes += cacheEntrySize(key.Start, key.End) - oldSize
		}
		__antithesis_instrumentation__.Notify(127146)
		addRange(r)
	} else {
		__antithesis_instrumentation__.Notify(127195)
	}
}

func (tc *treeImpl) getLowWater() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127196)
	tc.RLock()
	defer tc.RUnlock()
	return tc.lowWater
}

func (tc *treeImpl) GetMax(start, end roachpb.Key) (hlc.Timestamp, uuid.UUID) {
	__antithesis_instrumentation__.Notify(127197)
	return tc.getMax(start, end)
}

func (tc *treeImpl) getMax(start, end roachpb.Key) (hlc.Timestamp, uuid.UUID) {
	__antithesis_instrumentation__.Notify(127198)
	tc.Lock()
	defer tc.Unlock()
	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(127201)
		end = start.Next()
	} else {
		__antithesis_instrumentation__.Notify(127202)
	}
	__antithesis_instrumentation__.Notify(127199)
	maxTS := tc.lowWater
	maxTxnID := noTxnID
	for _, o := range tc.cache.GetOverlaps(start, end) {
		__antithesis_instrumentation__.Notify(127203)
		ce := o.Value.(*cacheValue)
		if maxTS.Less(ce.ts) {
			__antithesis_instrumentation__.Notify(127204)
			maxTS = ce.ts
			maxTxnID = ce.txnID
		} else {
			__antithesis_instrumentation__.Notify(127205)
			if maxTS == ce.ts && func() bool {
				__antithesis_instrumentation__.Notify(127206)
				return maxTxnID != ce.txnID == true
			}() == true {
				__antithesis_instrumentation__.Notify(127207)
				maxTxnID = noTxnID
			} else {
				__antithesis_instrumentation__.Notify(127208)
			}
		}
	}
	__antithesis_instrumentation__.Notify(127200)
	return maxTS, maxTxnID
}

func (tc *treeImpl) shouldEvict(size int, key, value interface{}) bool {
	__antithesis_instrumentation__.Notify(127209)
	if tc.bytes <= tc.maxBytes {
		__antithesis_instrumentation__.Notify(127213)
		return false
	} else {
		__antithesis_instrumentation__.Notify(127214)
	}
	__antithesis_instrumentation__.Notify(127210)
	ce := value.(*cacheValue)

	if ce.ts.Less(tc.lowWater) {
		__antithesis_instrumentation__.Notify(127215)
		return true
	} else {
		__antithesis_instrumentation__.Notify(127216)
	}
	__antithesis_instrumentation__.Notify(127211)

	edge := tc.latest.Add(-MinRetentionWindow.Nanoseconds(), 0)

	if ce.ts.LessEq(edge) {
		__antithesis_instrumentation__.Notify(127217)
		tc.lowWater = ce.ts
		return true
	} else {
		__antithesis_instrumentation__.Notify(127218)
	}
	__antithesis_instrumentation__.Notify(127212)
	return false
}

func (tc *treeImpl) onEvicted(k, v interface{}) {
	__antithesis_instrumentation__.Notify(127219)
	ck := k.(*cache.IntervalKey)
	reqSize := cacheEntrySize(ck.Start, ck.End)
	if tc.bytes < reqSize {
		__antithesis_instrumentation__.Notify(127221)
		panic(fmt.Sprintf("bad reqSize: %d < %d", tc.bytes, reqSize))
	} else {
		__antithesis_instrumentation__.Notify(127222)
	}
	__antithesis_instrumentation__.Notify(127220)
	tc.bytes -= reqSize
}

func (tc *treeImpl) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(127223)
	return tc.metrics
}
