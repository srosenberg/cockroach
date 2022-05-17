package metamorphic

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
)

type operandType int

const (
	operandTransaction operandType = iota
	operandReadWriter
	operandMVCCKey
	operandUnusedMVCCKey
	operandPastTS
	operandNextTS
	operandValue
	operandIterator
	operandFloat
	operandBool
	operandSavepoint
)

const (
	maxValueSize = 16
)

type operandGenerator interface {
	get(args []string) string

	getNew(args []string) string

	opener() string

	count(args []string) int

	closeAll()
}

func generateBytes(rng *rand.Rand, min int, max int) []byte {
	__antithesis_instrumentation__.Notify(639772)

	iterations := min + rng.Intn(max-min)
	result := make([]byte, 0, iterations)

	for i := 0; i < iterations; i++ {
		__antithesis_instrumentation__.Notify(639774)
		result = append(result, byte(rng.Float64()*float64('z'-'a')+'a'))
	}
	__antithesis_instrumentation__.Notify(639773)
	return result
}

type keyGenerator struct {
	liveKeys    []storage.MVCCKey
	rng         *rand.Rand
	tsGenerator *tsGenerator
}

var _ operandGenerator = &keyGenerator{}

func (k *keyGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639775)
	return ""
}

func (k *keyGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639776)

	return len(k.liveKeys) + 1
}

func (k *keyGenerator) open() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(639777)
	var key storage.MVCCKey
	key.Key = generateBytes(k.rng, 8, maxValueSize)
	key.Timestamp = k.tsGenerator.lastTS
	k.liveKeys = append(k.liveKeys, key)

	return key
}

func (k *keyGenerator) toString(key storage.MVCCKey) string {
	__antithesis_instrumentation__.Notify(639778)
	return fmt.Sprintf("%s/%d", key.Key, key.Timestamp.WallTime)
}

func (k *keyGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639779)

	if len(k.liveKeys) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(639781)
		return k.rng.Float64() < 0.30 == true
	}() == true {
		__antithesis_instrumentation__.Notify(639782)
		return k.toString(k.open())
	} else {
		__antithesis_instrumentation__.Notify(639783)
	}
	__antithesis_instrumentation__.Notify(639780)

	return k.toString(k.liveKeys[k.rng.Intn(len(k.liveKeys))])
}

func (k *keyGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639784)
	return k.get(args)
}

func (k *keyGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639785)

}

func (k *keyGenerator) parse(input string) storage.MVCCKey {
	__antithesis_instrumentation__.Notify(639786)
	var key storage.MVCCKey
	key.Key = make([]byte, 0, maxValueSize)
	_, err := fmt.Sscanf(input, "%q/%d", &key.Key, &key.Timestamp.WallTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(639788)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639789)
	}
	__antithesis_instrumentation__.Notify(639787)
	return key
}

type txnKeyGenerator struct {
	txns *txnGenerator
	keys *keyGenerator
}

var _ operandGenerator = &keyGenerator{}

func (k *txnKeyGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639790)
	return ""
}

func (k *txnKeyGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639791)

	return len(k.txns.inUseKeys) + 1
}

func (k *txnKeyGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639792)
	writer := readWriterID(args[0])
	txn := txnID(args[1])

	for {
		__antithesis_instrumentation__.Notify(639793)
		rawKey := k.keys.get(args)
		key := k.keys.parse(rawKey)

		conflictFound := false
		k.txns.forEachConflict(writer, txn, key.Key, nil, func(span roachpb.Span) bool {
			__antithesis_instrumentation__.Notify(639795)
			conflictFound = true
			return false
		})
		__antithesis_instrumentation__.Notify(639794)
		if !conflictFound {
			__antithesis_instrumentation__.Notify(639796)
			return rawKey
		} else {
			__antithesis_instrumentation__.Notify(639797)
		}
	}
}

func (k *txnKeyGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639798)
	return k.get(args)
}

func (k *txnKeyGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639799)

}

func (k *txnKeyGenerator) parse(input string) storage.MVCCKey {
	__antithesis_instrumentation__.Notify(639800)
	return k.keys.parse(input)
}

type valueGenerator struct {
	rng *rand.Rand
}

var _ operandGenerator = &valueGenerator{}

func (v *valueGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639801)
	return ""
}

func (v *valueGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639802)
	return 1
}

func (v *valueGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639803)
	return v.toString(generateBytes(v.rng, 4, maxValueSize))
}

func (v *valueGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639804)
	return v.get(args)
}

func (v *valueGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639805)

}

func (v *valueGenerator) toString(value []byte) string {
	__antithesis_instrumentation__.Notify(639806)
	return string(value)
}

func (v *valueGenerator) parse(input string) []byte {
	__antithesis_instrumentation__.Notify(639807)
	return []byte(input)
}

type txnID string

type writtenKeySpan struct {
	key    roachpb.Span
	writer readWriterID
	txn    txnID
}

type txnGenerator struct {
	rng         *rand.Rand
	testRunner  *metaTestRunner
	tsGenerator *tsGenerator
	liveTxns    []txnID
	txnIDMap    map[txnID]*roachpb.Transaction

	openBatches map[txnID]map[readWriterID]struct{}

	openSavepoints map[txnID]int

	inUseKeys []writtenKeySpan

	txnGenCounter uint64
}

var _ operandGenerator = &txnGenerator{}

func (t *txnGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639808)
	return "txn_open"
}

func (t *txnGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639809)
	return len(t.txnIDMap)
}

func (t *txnGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639810)
	if len(t.liveTxns) == 0 {
		__antithesis_instrumentation__.Notify(639812)
		panic("no open txns")
	} else {
		__antithesis_instrumentation__.Notify(639813)
	}
	__antithesis_instrumentation__.Notify(639811)
	return string(t.liveTxns[t.rng.Intn(len(t.liveTxns))])
}

func (t *txnGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639814)
	t.txnGenCounter++
	id := txnID(fmt.Sprintf("t%d", t.txnGenCounter))

	t.tsGenerator.generate()

	t.txnIDMap[id] = nil
	t.liveTxns = append(t.liveTxns, id)
	return string(id)
}

func (t *txnGenerator) generateClose(id txnID) {
	__antithesis_instrumentation__.Notify(639815)
	delete(t.openBatches, id)
	delete(t.txnIDMap, id)
	delete(t.openSavepoints, id)

	j := 0
	for i := range t.inUseKeys {
		__antithesis_instrumentation__.Notify(639817)
		if t.inUseKeys[i].txn != id {
			__antithesis_instrumentation__.Notify(639818)
			t.inUseKeys[j] = t.inUseKeys[i]
			j++
		} else {
			__antithesis_instrumentation__.Notify(639819)
		}
	}
	__antithesis_instrumentation__.Notify(639816)
	t.inUseKeys = t.inUseKeys[:j]

	for i := range t.liveTxns {
		__antithesis_instrumentation__.Notify(639820)
		if t.liveTxns[i] == id {
			__antithesis_instrumentation__.Notify(639821)
			t.liveTxns[i] = t.liveTxns[len(t.liveTxns)-1]
			t.liveTxns = t.liveTxns[:len(t.liveTxns)-1]
			break
		} else {
			__antithesis_instrumentation__.Notify(639822)
		}
	}
}

func (t *txnGenerator) clearBatch(batch readWriterID) {
	__antithesis_instrumentation__.Notify(639823)
	for _, batches := range t.openBatches {
		__antithesis_instrumentation__.Notify(639824)
		delete(batches, batch)
	}
}

func (t *txnGenerator) forEachConflict(
	w readWriterID, txn txnID, key roachpb.Key, endKey roachpb.Key, fn func(roachpb.Span) bool,
) {
	__antithesis_instrumentation__.Notify(639825)
	if endKey == nil {
		__antithesis_instrumentation__.Notify(639829)
		endKey = key.Next()
	} else {
		__antithesis_instrumentation__.Notify(639830)
	}
	__antithesis_instrumentation__.Notify(639826)

	start := sort.Search(len(t.inUseKeys), func(i int) bool {
		__antithesis_instrumentation__.Notify(639831)
		return key.Compare(t.inUseKeys[i].key.EndKey) < 0
	})
	__antithesis_instrumentation__.Notify(639827)
	end := sort.Search(len(t.inUseKeys), func(i int) bool {
		__antithesis_instrumentation__.Notify(639832)
		return endKey.Compare(t.inUseKeys[i].key.Key) <= 0
	})
	__antithesis_instrumentation__.Notify(639828)

	for i := start; i < end; i++ {
		__antithesis_instrumentation__.Notify(639833)
		if t.inUseKeys[i].writer != w || func() bool {
			__antithesis_instrumentation__.Notify(639834)
			return t.inUseKeys[i].txn != txn == true
		}() == true {
			__antithesis_instrumentation__.Notify(639835)

			if cont := fn(t.inUseKeys[i].key); !cont {
				__antithesis_instrumentation__.Notify(639836)
				return
			} else {
				__antithesis_instrumentation__.Notify(639837)
			}
		} else {
			__antithesis_instrumentation__.Notify(639838)
		}
	}
}

func (t *txnGenerator) truncateSpanForConflicts(
	w readWriterID, txn txnID, key, endKey roachpb.Key,
) roachpb.Span {
	__antithesis_instrumentation__.Notify(639839)

	t.forEachConflict(w, txn, key, endKey, func(conflict roachpb.Span) bool {
		__antithesis_instrumentation__.Notify(639841)
		if conflict.ContainsKey(key) {
			__antithesis_instrumentation__.Notify(639843)
			key = append([]byte(nil), conflict.EndKey...)
			return true
		} else {
			__antithesis_instrumentation__.Notify(639844)
		}
		__antithesis_instrumentation__.Notify(639842)
		endKey = conflict.Key
		return false
	})
	__antithesis_instrumentation__.Notify(639840)
	result := roachpb.Span{
		Key:    key,
		EndKey: endKey,
	}
	return result
}

func (t *txnGenerator) addWrittenKeySpan(
	w readWriterID, txn txnID, key roachpb.Key, endKey roachpb.Key,
) {
	__antithesis_instrumentation__.Notify(639845)
	span := roachpb.Span{Key: key, EndKey: endKey}
	if endKey == nil {
		__antithesis_instrumentation__.Notify(639853)
		endKey = key.Next()
		span.EndKey = endKey
	} else {
		__antithesis_instrumentation__.Notify(639854)
	}
	__antithesis_instrumentation__.Notify(639846)

	start := sort.Search(len(t.inUseKeys), func(i int) bool {
		__antithesis_instrumentation__.Notify(639855)
		return key.Compare(t.inUseKeys[i].key.EndKey) < 0
	})
	__antithesis_instrumentation__.Notify(639847)
	end := sort.Search(len(t.inUseKeys), func(i int) bool {
		__antithesis_instrumentation__.Notify(639856)
		return endKey.Compare(t.inUseKeys[i].key.Key) <= 0
	})
	__antithesis_instrumentation__.Notify(639848)
	if start == len(t.inUseKeys) {
		__antithesis_instrumentation__.Notify(639857)

		t.inUseKeys = append(t.inUseKeys, writtenKeySpan{
			key:    span,
			writer: w,
			txn:    txn,
		})
		return
	} else {
		__antithesis_instrumentation__.Notify(639858)
	}
	__antithesis_instrumentation__.Notify(639849)
	if start == end {
		__antithesis_instrumentation__.Notify(639859)

		t.inUseKeys = append(t.inUseKeys, writtenKeySpan{})
		copy(t.inUseKeys[start+1:], t.inUseKeys[start:])
		t.inUseKeys[start] = writtenKeySpan{
			key:    span,
			writer: w,
			txn:    txn,
		}
		return
	} else {
		__antithesis_instrumentation__.Notify(639860)
		if start > end {
			__antithesis_instrumentation__.Notify(639861)
			panic(fmt.Sprintf("written keys not in sorted order: %d > %d", start, end))
		} else {
			__antithesis_instrumentation__.Notify(639862)
		}
	}
	__antithesis_instrumentation__.Notify(639850)

	if t.inUseKeys[start].key.Key.Compare(key) < 0 {
		__antithesis_instrumentation__.Notify(639863)

		span.Key = t.inUseKeys[start].key.Key
	} else {
		__antithesis_instrumentation__.Notify(639864)
	}
	__antithesis_instrumentation__.Notify(639851)

	if t.inUseKeys[end-1].key.EndKey.Compare(endKey) > 0 {
		__antithesis_instrumentation__.Notify(639865)

		span.EndKey = t.inUseKeys[end-1].key.EndKey
	} else {
		__antithesis_instrumentation__.Notify(639866)
	}
	__antithesis_instrumentation__.Notify(639852)

	t.inUseKeys[start] = writtenKeySpan{
		key:    span,
		writer: w,
		txn:    txn,
	}
	n := copy(t.inUseKeys[start+1:], t.inUseKeys[end:])
	t.inUseKeys = t.inUseKeys[:start+1+n]
}

func (t *txnGenerator) trackTransactionalWrite(w readWriterID, txn txnID, key, endKey roachpb.Key) {
	__antithesis_instrumentation__.Notify(639867)
	if len(endKey) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(639871)
		return key.Compare(endKey) >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(639872)

		return
	} else {
		__antithesis_instrumentation__.Notify(639873)
	}
	__antithesis_instrumentation__.Notify(639868)
	t.addWrittenKeySpan(w, txn, key, endKey)
	if w == "engine" {
		__antithesis_instrumentation__.Notify(639874)
		return
	} else {
		__antithesis_instrumentation__.Notify(639875)
	}
	__antithesis_instrumentation__.Notify(639869)
	openBatches, ok := t.openBatches[txn]
	if !ok {
		__antithesis_instrumentation__.Notify(639876)
		t.openBatches[txn] = make(map[readWriterID]struct{})
		openBatches = t.openBatches[txn]
	} else {
		__antithesis_instrumentation__.Notify(639877)
	}
	__antithesis_instrumentation__.Notify(639870)
	openBatches[w] = struct{}{}
}

func (t *txnGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639878)
	t.liveTxns = nil
	t.txnIDMap = make(map[txnID]*roachpb.Transaction)
	t.openBatches = make(map[txnID]map[readWriterID]struct{})
	t.inUseKeys = t.inUseKeys[:0]
	t.openSavepoints = make(map[txnID]int)
}

type savepointGenerator struct {
	rng          *rand.Rand
	txnGenerator *txnGenerator
}

var _ operandGenerator = &savepointGenerator{}

func (s *savepointGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639879)

	id := txnID(args[len(args)-1])
	n := s.rng.Intn(s.txnGenerator.openSavepoints[id])
	return strconv.Itoa(n)
}

func (s *savepointGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639880)
	id := txnID(args[len(args)-1])
	s.txnGenerator.openSavepoints[id]++
	return strconv.Itoa(s.txnGenerator.openSavepoints[id] - 1)
}

func (s *savepointGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639881)
	return "txn_create_savepoint"
}

func (s *savepointGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639882)
	id := txnID(args[len(args)-1])
	return s.txnGenerator.openSavepoints[id]
}

func (s *savepointGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639883)

}

type pastTSGenerator struct {
	rng         *rand.Rand
	tsGenerator *tsGenerator
}

var _ operandGenerator = &pastTSGenerator{}

func (t *pastTSGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639884)
	return ""
}

func (t *pastTSGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639885)

	return int(t.tsGenerator.lastTS.WallTime) + 1
}

func (t *pastTSGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639886)

}

func (t *pastTSGenerator) toString(ts hlc.Timestamp) string {
	__antithesis_instrumentation__.Notify(639887)
	return fmt.Sprintf("%d", ts.WallTime)
}

func (t *pastTSGenerator) parse(input string) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(639888)
	var ts hlc.Timestamp
	wallTime, err := strconv.ParseInt(input, 10, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(639890)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639891)
	}
	__antithesis_instrumentation__.Notify(639889)
	ts.WallTime = wallTime
	return ts
}

func (t *pastTSGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639892)
	return t.toString(t.tsGenerator.randomPastTimestamp(t.rng))
}

func (t *pastTSGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639893)
	return t.get(args)
}

type nextTSGenerator struct {
	pastTSGenerator
}

func (t *nextTSGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639894)
	return t.toString(t.tsGenerator.generate())
}

func (t *nextTSGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639895)
	return t.get(args)
}

type readWriterID string

type readWriterGenerator struct {
	rng             *rand.Rand
	m               *metaTestRunner
	liveBatches     []readWriterID
	batchIDMap      map[readWriterID]storage.Batch
	batchGenCounter uint64
}

var _ operandGenerator = &readWriterGenerator{}

func (w *readWriterGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639896)

	if len(w.liveBatches) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(639898)
		return w.rng.Float64() < 0.25 == true
	}() == true {
		__antithesis_instrumentation__.Notify(639899)
		return "engine"
	} else {
		__antithesis_instrumentation__.Notify(639900)
	}
	__antithesis_instrumentation__.Notify(639897)

	return string(w.liveBatches[w.rng.Intn(len(w.liveBatches))])
}

func (w *readWriterGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639901)
	w.batchGenCounter++
	id := readWriterID(fmt.Sprintf("batch%d", w.batchGenCounter))
	w.batchIDMap[id] = nil
	w.liveBatches = append(w.liveBatches, id)

	return string(id)
}

func (w *readWriterGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639902)
	return "batch_open"
}

func (w *readWriterGenerator) generateClose(id readWriterID) {
	__antithesis_instrumentation__.Notify(639903)
	if id == "engine" {
		__antithesis_instrumentation__.Notify(639906)
		return
	} else {
		__antithesis_instrumentation__.Notify(639907)
	}
	__antithesis_instrumentation__.Notify(639904)
	delete(w.batchIDMap, id)
	for i, batch := range w.liveBatches {
		__antithesis_instrumentation__.Notify(639908)
		if batch == id {
			__antithesis_instrumentation__.Notify(639909)
			w.liveBatches[i] = w.liveBatches[len(w.liveBatches)-1]
			w.liveBatches = w.liveBatches[:len(w.liveBatches)-1]
			break
		} else {
			__antithesis_instrumentation__.Notify(639910)
		}
	}
	__antithesis_instrumentation__.Notify(639905)
	w.m.txnGenerator.clearBatch(id)
}

func (w *readWriterGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639911)
	return len(w.batchIDMap) + 1
}

func (w *readWriterGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639912)
	for _, batch := range w.batchIDMap {
		__antithesis_instrumentation__.Notify(639914)
		if batch != nil {
			__antithesis_instrumentation__.Notify(639915)
			batch.Close()
		} else {
			__antithesis_instrumentation__.Notify(639916)
		}
	}
	__antithesis_instrumentation__.Notify(639913)
	w.liveBatches = w.liveBatches[:0]
	w.batchIDMap = make(map[readWriterID]storage.Batch)
}

type iteratorID string
type iteratorInfo struct {
	id          iteratorID
	iter        storage.MVCCIterator
	lowerBound  roachpb.Key
	isBatchIter bool
}

type iteratorGenerator struct {
	rng            *rand.Rand
	readerToIter   map[readWriterID][]iteratorID
	iterInfo       map[iteratorID]iteratorInfo
	liveIters      []iteratorID
	iterGenCounter uint64
}

var _ operandGenerator = &iteratorGenerator{}

func (i *iteratorGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639917)
	if len(i.liveIters) == 0 {
		__antithesis_instrumentation__.Notify(639919)
		panic("no open iterators")
	} else {
		__antithesis_instrumentation__.Notify(639920)
	}
	__antithesis_instrumentation__.Notify(639918)

	return string(i.liveIters[i.rng.Intn(len(i.liveIters))])
}

func (i *iteratorGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639921)
	i.iterGenCounter++
	id := fmt.Sprintf("iter%d", i.iterGenCounter)
	return id
}

func (i *iteratorGenerator) generateOpen(rwID readWriterID, id iteratorID) {
	__antithesis_instrumentation__.Notify(639922)
	i.iterInfo[id] = iteratorInfo{
		id:          id,
		lowerBound:  nil,
		isBatchIter: rwID != "engine",
	}
	i.readerToIter[rwID] = append(i.readerToIter[rwID], id)
	i.liveIters = append(i.liveIters, id)
}

func (i *iteratorGenerator) generateClose(id iteratorID) {
	__antithesis_instrumentation__.Notify(639923)
	delete(i.iterInfo, id)

	for reader, iters := range i.readerToIter {
		__antithesis_instrumentation__.Notify(639925)
		for j, id2 := range iters {
			__antithesis_instrumentation__.Notify(639926)
			if id == id2 {
				__antithesis_instrumentation__.Notify(639927)

				iters[j] = iters[len(iters)-1]
				i.readerToIter[reader] = iters[:len(iters)-1]

				break
			} else {
				__antithesis_instrumentation__.Notify(639928)
			}
		}
	}
	__antithesis_instrumentation__.Notify(639924)

	for j, iter := range i.liveIters {
		__antithesis_instrumentation__.Notify(639929)
		if id == iter {
			__antithesis_instrumentation__.Notify(639930)
			i.liveIters[j] = i.liveIters[len(i.liveIters)-1]
			i.liveIters = i.liveIters[:len(i.liveIters)-1]
			break
		} else {
			__antithesis_instrumentation__.Notify(639931)
		}
	}
}

func (i *iteratorGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639932)
	return "iterator_open"
}

func (i *iteratorGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639933)
	return len(i.iterInfo)
}

func (i *iteratorGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639934)
	i.liveIters = nil
	i.iterInfo = make(map[iteratorID]iteratorInfo)
	i.readerToIter = make(map[readWriterID][]iteratorID)
}

type floatGenerator struct {
	rng *rand.Rand
}

func (f *floatGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639935)
	return fmt.Sprintf("%.4f", f.rng.Float32())
}

func (f *floatGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639936)
	return f.get(args)
}

func (f *floatGenerator) parse(input string) float32 {
	__antithesis_instrumentation__.Notify(639937)
	var result float32
	if _, err := fmt.Sscanf(input, "%f", &result); err != nil {
		__antithesis_instrumentation__.Notify(639939)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639940)
	}
	__antithesis_instrumentation__.Notify(639938)
	return result
}

func (f *floatGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639941)

	return ""
}

func (f *floatGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639942)
	return 1
}

func (f *floatGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639943)

}

type boolGenerator struct {
	rng *rand.Rand
}

func (f *boolGenerator) get(args []string) string {
	__antithesis_instrumentation__.Notify(639944)
	return fmt.Sprintf("%t", f.rng.Float32() < 0.5)
}

func (f *boolGenerator) getNew(args []string) string {
	__antithesis_instrumentation__.Notify(639945)
	return f.get(args)
}

func (f *boolGenerator) parse(input string) bool {
	__antithesis_instrumentation__.Notify(639946)
	var result bool
	if _, err := fmt.Sscanf(input, "%t", &result); err != nil {
		__antithesis_instrumentation__.Notify(639948)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(639949)
	}
	__antithesis_instrumentation__.Notify(639947)
	return result
}

func (f *boolGenerator) opener() string {
	__antithesis_instrumentation__.Notify(639950)

	return ""
}

func (f *boolGenerator) count(args []string) int {
	__antithesis_instrumentation__.Notify(639951)
	return 1
}

func (f *boolGenerator) closeAll() {
	__antithesis_instrumentation__.Notify(639952)

}
