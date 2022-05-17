package contention

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/biogo/store/llrb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/contentionpb"
	"github.com/cockroachdb/cockroach/pkg/util/cache"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

type Registry struct {
	globalLock syncutil.Mutex

	indexMap *indexMap

	nonSQLKeysMap *nonSQLKeysMap

	eventStore *eventStore
}

var (
	indexMapMaxSize = 50

	orderedKeyMapMaxSize = 50

	maxNumTxns = 10

	_ eventReader = &Registry{}
)

var orderedKeyMapCfg = cache.Config{
	Policy: cache.CacheLRU,
	ShouldEvict: func(size int, _, _ interface{}) bool {
		__antithesis_instrumentation__.Notify(459119)
		return size > orderedKeyMapMaxSize
	},
}

var txnCacheCfg = cache.Config{
	Policy: cache.CacheLRU,
	ShouldEvict: func(size int, _, _ interface{}) bool {
		__antithesis_instrumentation__.Notify(459120)
		return size > maxNumTxns
	},
}

type comparableKey roachpb.Key

func (c comparableKey) Compare(other llrb.Comparable) int {
	__antithesis_instrumentation__.Notify(459121)
	return roachpb.Key(c).Compare(roachpb.Key(other.(comparableKey)))
}

type indexMapKey struct {
	tableID descpb.ID
	indexID descpb.IndexID
}

type indexMapValue struct {
	numContentionEvents uint64

	cumulativeContentionTime time.Duration

	orderedKeyMap *cache.OrderedCache
}

func newIndexMapValue(c roachpb.ContentionEvent) *indexMapValue {
	__antithesis_instrumentation__.Notify(459122)
	txnCache := cache.NewUnorderedCache(txnCacheCfg)
	txnCache.Add(c.TxnMeta.ID, uint64(1))
	keyMap := cache.NewOrderedCache(orderedKeyMapCfg)
	keyMap.Add(comparableKey(c.Key), txnCache)
	return &indexMapValue{
		numContentionEvents:      1,
		cumulativeContentionTime: c.Duration,
		orderedKeyMap:            keyMap,
	}
}

func (v *indexMapValue) addContentionEvent(c roachpb.ContentionEvent) {
	__antithesis_instrumentation__.Notify(459123)
	v.numContentionEvents++
	v.cumulativeContentionTime += c.Duration
	var numTimesThisTxnWasEncountered uint64
	txnCache, ok := v.orderedKeyMap.Get(comparableKey(c.Key))
	if ok {
		__antithesis_instrumentation__.Notify(459125)
		if txnVal, ok := txnCache.(*cache.UnorderedCache).Get(c.TxnMeta.ID); ok {
			__antithesis_instrumentation__.Notify(459126)
			numTimesThisTxnWasEncountered = txnVal.(uint64)
		} else {
			__antithesis_instrumentation__.Notify(459127)
		}
	} else {
		__antithesis_instrumentation__.Notify(459128)

		txnCache = cache.NewUnorderedCache(txnCacheCfg)
		v.orderedKeyMap.Add(comparableKey(c.Key), txnCache)
	}
	__antithesis_instrumentation__.Notify(459124)
	txnCache.(*cache.UnorderedCache).Add(c.TxnMeta.ID, numTimesThisTxnWasEncountered+1)
}

type indexMap struct {
	internalCache *cache.UnorderedCache
	scratchKey    indexMapKey
}

func newIndexMap() *indexMap {
	__antithesis_instrumentation__.Notify(459129)
	return &indexMap{
		internalCache: cache.NewUnorderedCache(cache.Config{
			Policy: cache.CacheLRU,
			ShouldEvict: func(size int, _, _ interface{}) bool {
				__antithesis_instrumentation__.Notify(459130)
				return size > indexMapMaxSize
			},
		}),
	}
}

func (m *indexMap) get(tableID descpb.ID, indexID descpb.IndexID) (*indexMapValue, bool) {
	__antithesis_instrumentation__.Notify(459131)
	m.scratchKey.tableID = tableID
	m.scratchKey.indexID = indexID
	v, ok := m.internalCache.Get(m.scratchKey)
	if !ok {
		__antithesis_instrumentation__.Notify(459133)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(459134)
	}
	__antithesis_instrumentation__.Notify(459132)
	return v.(*indexMapValue), true
}

func (m *indexMap) add(tableID descpb.ID, indexID descpb.IndexID, v *indexMapValue) {
	__antithesis_instrumentation__.Notify(459135)
	m.scratchKey.tableID = tableID
	m.scratchKey.indexID = indexID
	m.internalCache.Add(m.scratchKey, v)
}

type nonSQLKeyMapValue struct {
	numContentionEvents uint64

	cumulativeContentionTime time.Duration

	txnCache *cache.UnorderedCache
}

func newNonSQLKeyMapValue(c roachpb.ContentionEvent) *nonSQLKeyMapValue {
	__antithesis_instrumentation__.Notify(459136)
	txnCache := cache.NewUnorderedCache(txnCacheCfg)
	txnCache.Add(c.TxnMeta.ID, uint64(1))
	return &nonSQLKeyMapValue{
		numContentionEvents:      1,
		cumulativeContentionTime: c.Duration,
		txnCache:                 txnCache,
	}
}

func (v *nonSQLKeyMapValue) addContentionEvent(c roachpb.ContentionEvent) {
	__antithesis_instrumentation__.Notify(459137)
	v.numContentionEvents++
	v.cumulativeContentionTime += c.Duration
	var numTimesThisTxnWasEncountered uint64
	if numTimes, ok := v.txnCache.Get(c.TxnMeta.ID); ok {
		__antithesis_instrumentation__.Notify(459139)
		numTimesThisTxnWasEncountered = numTimes.(uint64)
	} else {
		__antithesis_instrumentation__.Notify(459140)
	}
	__antithesis_instrumentation__.Notify(459138)
	v.txnCache.Add(c.TxnMeta.ID, numTimesThisTxnWasEncountered+1)
}

type nonSQLKeysMap struct {
	*cache.OrderedCache
}

func newNonSQLKeysMap() *nonSQLKeysMap {
	__antithesis_instrumentation__.Notify(459141)
	return &nonSQLKeysMap{
		OrderedCache: cache.NewOrderedCache(orderedKeyMapCfg),
	}
}

func NewRegistry(st *cluster.Settings, endpoint ResolverEndpoint, metrics *Metrics) *Registry {
	__antithesis_instrumentation__.Notify(459142)
	return &Registry{
		indexMap:      newIndexMap(),
		nonSQLKeysMap: newNonSQLKeysMap(),
		eventStore:    newEventStore(st, endpoint, timeutil.Now, metrics),
	}
}

func (r *Registry) Start(ctx context.Context, stopper *stop.Stopper) {
	__antithesis_instrumentation__.Notify(459143)
	r.eventStore.start(ctx, stopper)
}

func (r *Registry) AddContentionEvent(event contentionpb.ExtendedContentionEvent) {
	__antithesis_instrumentation__.Notify(459144)
	c := event.BlockingEvent
	r.globalLock.Lock()
	defer r.globalLock.Unlock()

	c.Key, _, _ = keys.DecodeTenantPrefix(c.Key)
	_, rawTableID, rawIndexID, err := keys.DecodeTableIDIndexID(c.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(459147)

		if v, ok := r.nonSQLKeysMap.Get(comparableKey(c.Key)); !ok {
			__antithesis_instrumentation__.Notify(459149)

			r.nonSQLKeysMap.Add(comparableKey(c.Key), newNonSQLKeyMapValue(c))
		} else {
			__antithesis_instrumentation__.Notify(459150)
			v.(*nonSQLKeyMapValue).addContentionEvent(c)
		}
		__antithesis_instrumentation__.Notify(459148)
		return
	} else {
		__antithesis_instrumentation__.Notify(459151)
	}
	__antithesis_instrumentation__.Notify(459145)
	tableID := descpb.ID(rawTableID)
	indexID := descpb.IndexID(rawIndexID)
	if v, ok := r.indexMap.get(tableID, indexID); !ok {
		__antithesis_instrumentation__.Notify(459152)

		r.indexMap.add(tableID, indexID, newIndexMapValue(c))
	} else {
		__antithesis_instrumentation__.Notify(459153)
		v.addContentionEvent(c)
	}
	__antithesis_instrumentation__.Notify(459146)

	r.eventStore.addEvent(event)
}

func (r *Registry) ForEachEvent(op func(event *contentionpb.ExtendedContentionEvent) error) error {
	__antithesis_instrumentation__.Notify(459154)
	return r.eventStore.ForEachEvent(op)
}

func (r *Registry) FlushEventsForTest(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(459155)
	return r.eventStore.flushAndResolve(ctx)
}

func serializeTxnCache(txnCache *cache.UnorderedCache) []contentionpb.SingleTxnContention {
	__antithesis_instrumentation__.Notify(459156)
	txns := make([]contentionpb.SingleTxnContention, txnCache.Len())
	var txnCount int
	txnCache.Do(func(e *cache.Entry) {
		__antithesis_instrumentation__.Notify(459158)
		txns[txnCount].TxnID = e.Key.(uuid.UUID)
		txns[txnCount].Count = e.Value.(uint64)
		txnCount++
	})
	__antithesis_instrumentation__.Notify(459157)
	sortSingleTxnContention(txns)
	return txns
}

func (r *Registry) Serialize() contentionpb.SerializedRegistry {
	__antithesis_instrumentation__.Notify(459159)
	r.globalLock.Lock()
	defer r.globalLock.Unlock()
	var resp contentionpb.SerializedRegistry

	resp.IndexContentionEvents = make([]contentionpb.IndexContentionEvents, r.indexMap.internalCache.Len())
	var iceCount int
	r.indexMap.internalCache.Do(func(e *cache.Entry) {
		__antithesis_instrumentation__.Notify(459162)
		ice := &resp.IndexContentionEvents[iceCount]
		key := e.Key.(indexMapKey)
		ice.TableID = key.tableID
		ice.IndexID = key.indexID
		v := e.Value.(*indexMapValue)
		ice.NumContentionEvents = v.numContentionEvents
		ice.CumulativeContentionTime = v.cumulativeContentionTime
		ice.Events = make([]contentionpb.SingleKeyContention, v.orderedKeyMap.Len())
		var skcCount int
		v.orderedKeyMap.Do(func(k, txnCacheInterface interface{}) bool {
			__antithesis_instrumentation__.Notify(459164)
			txnCache := txnCacheInterface.(*cache.UnorderedCache)
			skc := &ice.Events[skcCount]
			skc.Key = roachpb.Key(k.(comparableKey))
			skc.Txns = serializeTxnCache(txnCache)
			skcCount++
			return false
		})
		__antithesis_instrumentation__.Notify(459163)
		iceCount++
	})
	__antithesis_instrumentation__.Notify(459160)
	sortIndexContentionEvents(resp.IndexContentionEvents)

	resp.NonSQLKeysContention = make([]contentionpb.SingleNonSQLKeyContention, r.nonSQLKeysMap.Len())
	var snkcCount int
	r.nonSQLKeysMap.Do(func(k, nonSQLKeyMapVal interface{}) bool {
		__antithesis_instrumentation__.Notify(459165)
		snkc := &resp.NonSQLKeysContention[snkcCount]
		snkc.Key = roachpb.Key(k.(comparableKey))
		v := nonSQLKeyMapVal.(*nonSQLKeyMapValue)
		snkc.NumContentionEvents = v.numContentionEvents
		snkc.CumulativeContentionTime = v.cumulativeContentionTime
		snkc.Txns = serializeTxnCache(v.txnCache)
		snkcCount++
		return false
	})
	__antithesis_instrumentation__.Notify(459161)

	return resp
}

func sortIndexContentionEvents(ice []contentionpb.IndexContentionEvents) {
	__antithesis_instrumentation__.Notify(459166)
	sort.Slice(ice, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(459167)
		if ice[i].NumContentionEvents != ice[j].NumContentionEvents {
			__antithesis_instrumentation__.Notify(459169)
			return ice[i].NumContentionEvents > ice[j].NumContentionEvents
		} else {
			__antithesis_instrumentation__.Notify(459170)
		}
		__antithesis_instrumentation__.Notify(459168)
		return ice[i].CumulativeContentionTime > ice[j].CumulativeContentionTime
	})
}

func sortSingleTxnContention(txns []contentionpb.SingleTxnContention) {
	__antithesis_instrumentation__.Notify(459171)
	sort.Slice(txns, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(459172)
		return txns[i].Count > txns[j].Count
	})
}

func (r *Registry) String() string {
	__antithesis_instrumentation__.Notify(459173)
	var b strings.Builder
	serialized := r.Serialize()
	for i := range serialized.IndexContentionEvents {
		__antithesis_instrumentation__.Notify(459176)
		b.WriteString(serialized.IndexContentionEvents[i].String())
	}
	__antithesis_instrumentation__.Notify(459174)
	for i := range serialized.NonSQLKeysContention {
		__antithesis_instrumentation__.Notify(459177)
		b.WriteString(serialized.NonSQLKeysContention[i].String())
	}
	__antithesis_instrumentation__.Notify(459175)
	return b.String()
}

func MergeSerializedRegistries(
	first, second contentionpb.SerializedRegistry,
) contentionpb.SerializedRegistry {
	__antithesis_instrumentation__.Notify(459178)

	firstICE, secondICE := first.IndexContentionEvents, second.IndexContentionEvents
	for s := range secondICE {
		__antithesis_instrumentation__.Notify(459183)
		found := false
		for f := range firstICE {
			__antithesis_instrumentation__.Notify(459185)
			if firstICE[f].TableID == secondICE[s].TableID && func() bool {
				__antithesis_instrumentation__.Notify(459186)
				return firstICE[f].IndexID == secondICE[s].IndexID == true
			}() == true {
				__antithesis_instrumentation__.Notify(459187)
				firstICE[f] = mergeIndexContentionEvents(firstICE[f], secondICE[s])
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(459188)
			}
		}
		__antithesis_instrumentation__.Notify(459184)
		if !found {
			__antithesis_instrumentation__.Notify(459189)
			firstICE = append(firstICE, secondICE[s])
		} else {
			__antithesis_instrumentation__.Notify(459190)
		}
	}
	__antithesis_instrumentation__.Notify(459179)

	sortIndexContentionEvents(firstICE)
	if len(firstICE) > indexMapMaxSize {
		__antithesis_instrumentation__.Notify(459191)
		firstICE = firstICE[:indexMapMaxSize]
	} else {
		__antithesis_instrumentation__.Notify(459192)
	}
	__antithesis_instrumentation__.Notify(459180)
	first.IndexContentionEvents = firstICE

	firstNKC, secondNKC := first.NonSQLKeysContention, second.NonSQLKeysContention
	maxNumNonSQLKeys := len(firstNKC) + len(secondNKC)
	if maxNumNonSQLKeys > orderedKeyMapMaxSize {
		__antithesis_instrumentation__.Notify(459193)
		maxNumNonSQLKeys = orderedKeyMapMaxSize
	} else {
		__antithesis_instrumentation__.Notify(459194)
	}
	__antithesis_instrumentation__.Notify(459181)
	result := make([]contentionpb.SingleNonSQLKeyContention, maxNumNonSQLKeys)
	resultIdx, firstIdx, secondIdx := 0, 0, 0

	for resultIdx < maxNumNonSQLKeys && func() bool {
		__antithesis_instrumentation__.Notify(459195)
		return (firstIdx < len(firstNKC) || func() bool {
			__antithesis_instrumentation__.Notify(459196)
			return secondIdx < len(secondNKC) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(459197)
		var cmp int
		if firstIdx == len(firstNKC) {
			__antithesis_instrumentation__.Notify(459200)

			cmp = 1
		} else {
			__antithesis_instrumentation__.Notify(459201)
			if secondIdx == len(secondNKC) {
				__antithesis_instrumentation__.Notify(459202)

				cmp = -1
			} else {
				__antithesis_instrumentation__.Notify(459203)
				cmp = firstNKC[firstIdx].Key.Compare(secondNKC[secondIdx].Key)
			}
		}
		__antithesis_instrumentation__.Notify(459198)
		if cmp == 0 {
			__antithesis_instrumentation__.Notify(459204)

			f := firstNKC[firstIdx]
			s := secondNKC[secondIdx]
			result[resultIdx].Key = f.Key
			result[resultIdx].NumContentionEvents = f.NumContentionEvents + s.NumContentionEvents
			result[resultIdx].CumulativeContentionTime = f.CumulativeContentionTime + s.CumulativeContentionTime
			result[resultIdx].Txns = mergeSingleKeyContention(f.Txns, s.Txns)
			firstIdx++
			secondIdx++
		} else {
			__antithesis_instrumentation__.Notify(459205)
			if cmp < 0 {
				__antithesis_instrumentation__.Notify(459206)

				result[resultIdx] = firstNKC[firstIdx]
				firstIdx++
			} else {
				__antithesis_instrumentation__.Notify(459207)

				result[resultIdx] = secondNKC[secondIdx]
				secondIdx++
			}
		}
		__antithesis_instrumentation__.Notify(459199)
		resultIdx++
	}
	__antithesis_instrumentation__.Notify(459182)

	result = result[:resultIdx]
	first.NonSQLKeysContention = result

	return first
}

func mergeIndexContentionEvents(
	first contentionpb.IndexContentionEvents, second contentionpb.IndexContentionEvents,
) contentionpb.IndexContentionEvents {
	__antithesis_instrumentation__.Notify(459208)
	if first.TableID != second.TableID || func() bool {
		__antithesis_instrumentation__.Notify(459212)
		return first.IndexID != second.IndexID == true
	}() == true {
		__antithesis_instrumentation__.Notify(459213)
		panic(fmt.Sprintf("attempting to merge contention events from different indexes\n%s%s", first.String(), second.String()))
	} else {
		__antithesis_instrumentation__.Notify(459214)
	}
	__antithesis_instrumentation__.Notify(459209)
	var result contentionpb.IndexContentionEvents
	result.TableID = first.TableID
	result.IndexID = first.IndexID
	result.NumContentionEvents = first.NumContentionEvents + second.NumContentionEvents
	result.CumulativeContentionTime = first.CumulativeContentionTime + second.CumulativeContentionTime

	maxNumEvents := len(first.Events) + len(second.Events)
	if maxNumEvents > orderedKeyMapMaxSize {
		__antithesis_instrumentation__.Notify(459215)
		maxNumEvents = orderedKeyMapMaxSize
	} else {
		__antithesis_instrumentation__.Notify(459216)
	}
	__antithesis_instrumentation__.Notify(459210)
	result.Events = make([]contentionpb.SingleKeyContention, maxNumEvents)
	resultIdx, firstIdx, secondIdx := 0, 0, 0

	for resultIdx < maxNumEvents && func() bool {
		__antithesis_instrumentation__.Notify(459217)
		return (firstIdx < len(first.Events) || func() bool {
			__antithesis_instrumentation__.Notify(459218)
			return secondIdx < len(second.Events) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(459219)
		var cmp int
		if firstIdx == len(first.Events) {
			__antithesis_instrumentation__.Notify(459222)

			cmp = 1
		} else {
			__antithesis_instrumentation__.Notify(459223)
			if secondIdx == len(second.Events) {
				__antithesis_instrumentation__.Notify(459224)

				cmp = -1
			} else {
				__antithesis_instrumentation__.Notify(459225)
				cmp = first.Events[firstIdx].Key.Compare(second.Events[secondIdx].Key)
			}
		}
		__antithesis_instrumentation__.Notify(459220)
		if cmp == 0 {
			__antithesis_instrumentation__.Notify(459226)

			f := first.Events[firstIdx]
			s := second.Events[secondIdx]
			result.Events[resultIdx].Key = f.Key
			result.Events[resultIdx].Txns = mergeSingleKeyContention(f.Txns, s.Txns)
			firstIdx++
			secondIdx++
		} else {
			__antithesis_instrumentation__.Notify(459227)
			if cmp < 0 {
				__antithesis_instrumentation__.Notify(459228)

				result.Events[resultIdx] = first.Events[firstIdx]
				firstIdx++
			} else {
				__antithesis_instrumentation__.Notify(459229)

				result.Events[resultIdx] = second.Events[secondIdx]
				secondIdx++
			}
		}
		__antithesis_instrumentation__.Notify(459221)
		resultIdx++
	}
	__antithesis_instrumentation__.Notify(459211)

	result.Events = result.Events[:resultIdx]
	return result
}

func mergeSingleKeyContention(
	first, second []contentionpb.SingleTxnContention,
) []contentionpb.SingleTxnContention {
	__antithesis_instrumentation__.Notify(459230)
	for s := range second {
		__antithesis_instrumentation__.Notify(459233)
		found := false
		for f := range first {
			__antithesis_instrumentation__.Notify(459235)
			if bytes.Equal(first[f].TxnID.GetBytes(), second[s].TxnID.GetBytes()) {
				__antithesis_instrumentation__.Notify(459236)
				first[f].Count += second[s].Count
				found = true
				break
			} else {
				__antithesis_instrumentation__.Notify(459237)
			}
		}
		__antithesis_instrumentation__.Notify(459234)
		if !found {
			__antithesis_instrumentation__.Notify(459238)
			first = append(first, second[s])
		} else {
			__antithesis_instrumentation__.Notify(459239)
		}
	}
	__antithesis_instrumentation__.Notify(459231)

	sortSingleTxnContention(first)
	if len(first) > maxNumTxns {
		__antithesis_instrumentation__.Notify(459240)
		first = first[:maxNumTxns]
	} else {
		__antithesis_instrumentation__.Notify(459241)
	}
	__antithesis_instrumentation__.Notify(459232)
	return first
}
