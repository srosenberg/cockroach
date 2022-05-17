package tscache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"container/list"
	"context"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/andy-kimball/arenaskl"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

type rangeOptions int

const (
	excludeFrom = rangeOptions(1 << iota)

	excludeTo
)

type nodeOptions int

const (
	initialized = 1 << iota

	cantInit

	hasKey

	hasGap
)

const (
	encodedValSize = int(unsafe.Sizeof(cacheValue{}))

	initialSklPageSize = 128 << 10

	maximumSklPageSize = 32 << 20

	defaultMinSklPages = 2
)

var initialSklAllocSize = func() int {
	__antithesis_instrumentation__.Notify(126816)
	a := arenaskl.NewArena(1000)
	_ = arenaskl.NewSkiplist(a)
	return int(a.Size())
}()

type intervalSkl struct {
	rotMutex syncutil.RWMutex

	clock  *hlc.Clock
	minRet time.Duration

	pageSize      uint32
	pageSizeFixed bool

	pages    list.List
	minPages int

	floorTS hlc.Timestamp

	metrics sklMetrics
}

func newIntervalSkl(clock *hlc.Clock, minRet time.Duration, metrics sklMetrics) *intervalSkl {
	__antithesis_instrumentation__.Notify(126817)
	s := intervalSkl{
		clock:    clock,
		minRet:   minRet,
		pageSize: initialSklPageSize / 2,
		minPages: defaultMinSklPages,
		metrics:  metrics,
	}
	s.pushNewPage(0, nil)
	s.metrics.Pages.Update(1)
	return &s
}

func (s *intervalSkl) Add(key []byte, val cacheValue) {
	__antithesis_instrumentation__.Notify(126818)
	s.AddRange(nil, key, 0, val)
}

func (s *intervalSkl) AddRange(from, to []byte, opt rangeOptions, val cacheValue) {
	__antithesis_instrumentation__.Notify(126819)
	if from == nil && func() bool {
		__antithesis_instrumentation__.Notify(126823)
		return to == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(126824)
		panic("from and to keys cannot be nil")
	} else {
		__antithesis_instrumentation__.Notify(126825)
	}
	__antithesis_instrumentation__.Notify(126820)
	if encodedRangeSize(from, to, opt) > int(s.maximumPageSize())-initialSklAllocSize {
		__antithesis_instrumentation__.Notify(126826)

		panic("key range too large to fit in any page")
	} else {
		__antithesis_instrumentation__.Notify(126827)
	}
	__antithesis_instrumentation__.Notify(126821)

	if to != nil {
		__antithesis_instrumentation__.Notify(126828)
		cmp := 0
		if from != nil {
			__antithesis_instrumentation__.Notify(126830)
			cmp = bytes.Compare(from, to)
		} else {
			__antithesis_instrumentation__.Notify(126831)
		}
		__antithesis_instrumentation__.Notify(126829)

		switch {
		case cmp > 0:
			__antithesis_instrumentation__.Notify(126832)

			d := 0
			for d < len(from) && func() bool {
				__antithesis_instrumentation__.Notify(126837)
				return d < len(to) == true
			}() == true {
				__antithesis_instrumentation__.Notify(126838)
				if from[d] != to[d] {
					__antithesis_instrumentation__.Notify(126840)
					break
				} else {
					__antithesis_instrumentation__.Notify(126841)
				}
				__antithesis_instrumentation__.Notify(126839)
				d++
			}
			__antithesis_instrumentation__.Notify(126833)
			msg := fmt.Sprintf("inverted range (issue #32149): key lens = [%d,%d), diff @ index %d",
				len(from), len(to), d)
			log.Errorf(context.Background(), "%s, [%s,%s)", msg, from, to)
			panic(redact.Safe(msg))
		case cmp == 0:
			__antithesis_instrumentation__.Notify(126834)

			if opt == (excludeFrom | excludeTo) {
				__antithesis_instrumentation__.Notify(126842)

				return
			} else {
				__antithesis_instrumentation__.Notify(126843)
			}
			__antithesis_instrumentation__.Notify(126835)

			from = nil
			opt = 0
		default:
			__antithesis_instrumentation__.Notify(126836)
		}
	} else {
		__antithesis_instrumentation__.Notify(126844)
	}
	__antithesis_instrumentation__.Notify(126822)

	for {
		__antithesis_instrumentation__.Notify(126845)

		filledPage := s.addRange(from, to, opt, val)
		if filledPage == nil {
			__antithesis_instrumentation__.Notify(126847)
			break
		} else {
			__antithesis_instrumentation__.Notify(126848)
		}
		__antithesis_instrumentation__.Notify(126846)

		s.rotatePages(filledPage)
	}
}

func (s *intervalSkl) addRange(from, to []byte, opt rangeOptions, val cacheValue) *sklPage {
	__antithesis_instrumentation__.Notify(126849)

	s.rotMutex.RLock()
	defer s.rotMutex.RUnlock()

	if val.ts.Less(s.floorTS) {
		__antithesis_instrumentation__.Notify(126857)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126858)
	}
	__antithesis_instrumentation__.Notify(126850)

	fp := s.frontPage()

	var it arenaskl.Iterator
	it.Init(fp.list)

	var err error
	if to != nil {
		__antithesis_instrumentation__.Notify(126859)
		if (opt & excludeTo) == 0 {
			__antithesis_instrumentation__.Notify(126861)
			err = fp.addNode(&it, to, val, hasKey, true)
		} else {
			__antithesis_instrumentation__.Notify(126862)
			err = fp.addNode(&it, to, val, 0, true)
		}
		__antithesis_instrumentation__.Notify(126860)

		if errors.Is(err, arenaskl.ErrArenaFull) {
			__antithesis_instrumentation__.Notify(126863)
			return fp
		} else {
			__antithesis_instrumentation__.Notify(126864)
		}
	} else {
		__antithesis_instrumentation__.Notify(126865)
	}
	__antithesis_instrumentation__.Notify(126851)

	if from == nil {
		__antithesis_instrumentation__.Notify(126866)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126867)
	}
	__antithesis_instrumentation__.Notify(126852)

	if (opt & excludeFrom) == 0 {
		__antithesis_instrumentation__.Notify(126868)
		err = fp.addNode(&it, from, val, hasKey|hasGap, false)
	} else {
		__antithesis_instrumentation__.Notify(126869)
		err = fp.addNode(&it, from, val, hasGap, false)
	}
	__antithesis_instrumentation__.Notify(126853)

	if errors.Is(err, arenaskl.ErrArenaFull) {
		__antithesis_instrumentation__.Notify(126870)
		return fp
	} else {
		__antithesis_instrumentation__.Notify(126871)
	}
	__antithesis_instrumentation__.Notify(126854)

	if !it.Valid() || func() bool {
		__antithesis_instrumentation__.Notify(126872)
		return !bytes.Equal(it.Key(), from) == true
	}() == true {
		__antithesis_instrumentation__.Notify(126873)

		if it.Seek(from) {
			__antithesis_instrumentation__.Notify(126874)
			it.Next()
		} else {
			__antithesis_instrumentation__.Notify(126875)
		}
	} else {
		__antithesis_instrumentation__.Notify(126876)
		it.Next()
	}
	__antithesis_instrumentation__.Notify(126855)

	if !fp.ensureFloorValue(&it, to, val) {
		__antithesis_instrumentation__.Notify(126877)

		return fp
	} else {
		__antithesis_instrumentation__.Notify(126878)
	}
	__antithesis_instrumentation__.Notify(126856)

	return nil
}

func (s *intervalSkl) frontPage() *sklPage {
	__antithesis_instrumentation__.Notify(126879)
	return s.pages.Front().Value.(*sklPage)
}

func (s *intervalSkl) pushNewPage(maxTime ratchetingTime, arena *arenaskl.Arena) {
	__antithesis_instrumentation__.Notify(126880)
	size := s.nextPageSize()
	if arena != nil && func() bool {
		__antithesis_instrumentation__.Notify(126882)
		return arena.Cap() == size == true
	}() == true {
		__antithesis_instrumentation__.Notify(126883)

		arena.Reset()
	} else {
		__antithesis_instrumentation__.Notify(126884)

		arena = arenaskl.NewArena(size)
	}
	__antithesis_instrumentation__.Notify(126881)
	p := newSklPage(arena)
	p.maxTime = maxTime
	s.pages.PushFront(p)
}

func (s *intervalSkl) nextPageSize() uint32 {
	__antithesis_instrumentation__.Notify(126885)
	if s.pageSizeFixed || func() bool {
		__antithesis_instrumentation__.Notify(126888)
		return s.pageSize == maximumSklPageSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(126889)
		return s.pageSize
	} else {
		__antithesis_instrumentation__.Notify(126890)
	}
	__antithesis_instrumentation__.Notify(126886)
	s.pageSize *= 2
	if s.pageSize > maximumSklPageSize {
		__antithesis_instrumentation__.Notify(126891)
		s.pageSize = maximumSklPageSize
	} else {
		__antithesis_instrumentation__.Notify(126892)
	}
	__antithesis_instrumentation__.Notify(126887)
	return s.pageSize
}

func (s *intervalSkl) maximumPageSize() uint32 {
	__antithesis_instrumentation__.Notify(126893)
	if s.pageSizeFixed {
		__antithesis_instrumentation__.Notify(126895)
		return s.pageSize
	} else {
		__antithesis_instrumentation__.Notify(126896)
	}
	__antithesis_instrumentation__.Notify(126894)
	return maximumSklPageSize
}

func (s *intervalSkl) rotatePages(filledPage *sklPage) {
	__antithesis_instrumentation__.Notify(126897)

	s.rotMutex.Lock()
	defer s.rotMutex.Unlock()

	fp := s.frontPage()
	if filledPage != fp {
		__antithesis_instrumentation__.Notify(126901)

		return
	} else {
		__antithesis_instrumentation__.Notify(126902)
	}
	__antithesis_instrumentation__.Notify(126898)

	minTSToRetain := hlc.MaxTimestamp
	if s.clock != nil {
		__antithesis_instrumentation__.Notify(126903)
		minTSToRetain = s.clock.Now().Add(-s.minRet.Nanoseconds(), 0)
	} else {
		__antithesis_instrumentation__.Notify(126904)
	}
	__antithesis_instrumentation__.Notify(126899)

	back := s.pages.Back()
	var oldArena *arenaskl.Arena
	for s.pages.Len() >= s.minPages {
		__antithesis_instrumentation__.Notify(126905)
		bp := back.Value.(*sklPage)
		bpMaxTS := bp.getMaxTimestamp()
		if minTSToRetain.LessEq(bpMaxTS) {
			__antithesis_instrumentation__.Notify(126907)

			break
		} else {
			__antithesis_instrumentation__.Notify(126908)
		}
		__antithesis_instrumentation__.Notify(126906)

		s.floorTS.Forward(bpMaxTS)

		oldArena = bp.list.Arena()
		evict := back
		back = back.Prev()
		s.pages.Remove(evict)
	}
	__antithesis_instrumentation__.Notify(126900)

	s.pushNewPage(fp.maxTime, oldArena)

	s.metrics.Pages.Update(int64(s.pages.Len()))
	s.metrics.PageRotations.Inc(1)
}

func (s *intervalSkl) LookupTimestamp(key []byte) cacheValue {
	__antithesis_instrumentation__.Notify(126909)
	return s.LookupTimestampRange(nil, key, 0)
}

func (s *intervalSkl) LookupTimestampRange(from, to []byte, opt rangeOptions) cacheValue {
	__antithesis_instrumentation__.Notify(126910)
	if from == nil && func() bool {
		__antithesis_instrumentation__.Notify(126913)
		return to == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(126914)
		panic("from and to keys cannot be nil")
	} else {
		__antithesis_instrumentation__.Notify(126915)
	}
	__antithesis_instrumentation__.Notify(126911)

	s.rotMutex.RLock()
	defer s.rotMutex.RUnlock()

	var val cacheValue
	for e := s.pages.Front(); e != nil; e = e.Next() {
		__antithesis_instrumentation__.Notify(126916)
		p := e.Value.(*sklPage)

		maxTS := p.getMaxTimestamp()
		if maxTS.Less(val.ts) {
			__antithesis_instrumentation__.Notify(126918)
			break
		} else {
			__antithesis_instrumentation__.Notify(126919)
		}
		__antithesis_instrumentation__.Notify(126917)

		val2 := p.lookupTimestampRange(from, to, opt)
		val, _ = ratchetValue(val, val2)
	}
	__antithesis_instrumentation__.Notify(126912)

	floorVal := cacheValue{ts: s.floorTS, txnID: noTxnID}
	val, _ = ratchetValue(val, floorVal)

	return val
}

func (s *intervalSkl) FloorTS() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(126920)
	s.rotMutex.RLock()
	defer s.rotMutex.RUnlock()
	return s.floorTS
}

type sklPage struct {
	list    *arenaskl.Skiplist
	maxTime ratchetingTime
	isFull  int32
}

func newSklPage(arena *arenaskl.Arena) *sklPage {
	__antithesis_instrumentation__.Notify(126921)
	return &sklPage{list: arenaskl.NewSkiplist(arena)}
}

func (p *sklPage) lookupTimestampRange(from, to []byte, opt rangeOptions) cacheValue {
	__antithesis_instrumentation__.Notify(126922)
	if to != nil {
		__antithesis_instrumentation__.Notify(126924)
		cmp := 0
		if from != nil {
			__antithesis_instrumentation__.Notify(126927)
			cmp = bytes.Compare(from, to)
		} else {
			__antithesis_instrumentation__.Notify(126928)
		}
		__antithesis_instrumentation__.Notify(126925)

		if cmp > 0 {
			__antithesis_instrumentation__.Notify(126929)

			return cacheValue{}
		} else {
			__antithesis_instrumentation__.Notify(126930)
		}
		__antithesis_instrumentation__.Notify(126926)
		if cmp == 0 {
			__antithesis_instrumentation__.Notify(126931)

			if opt == (excludeFrom | excludeTo) {
				__antithesis_instrumentation__.Notify(126933)

				return cacheValue{}
			} else {
				__antithesis_instrumentation__.Notify(126934)
			}
			__antithesis_instrumentation__.Notify(126932)

			from = to
			opt = 0
		} else {
			__antithesis_instrumentation__.Notify(126935)
		}
	} else {
		__antithesis_instrumentation__.Notify(126936)
	}
	__antithesis_instrumentation__.Notify(126923)

	var it arenaskl.Iterator
	it.Init(p.list)
	it.SeekForPrev(from)

	return p.maxInRange(&it, from, to, opt)
}

func (p *sklPage) addNode(
	it *arenaskl.Iterator, key []byte, val cacheValue, opt nodeOptions, mustInit bool,
) error {
	__antithesis_instrumentation__.Notify(126937)

	var arr [encodedValSize * 2]byte
	var keyVal, gapVal cacheValue

	if (opt & hasKey) != 0 {
		__antithesis_instrumentation__.Notify(126943)
		keyVal = val
	} else {
		__antithesis_instrumentation__.Notify(126944)
	}
	__antithesis_instrumentation__.Notify(126938)

	if (opt & hasGap) != 0 {
		__antithesis_instrumentation__.Notify(126945)
		gapVal = val
	} else {
		__antithesis_instrumentation__.Notify(126946)
	}
	__antithesis_instrumentation__.Notify(126939)

	if !it.SeekForPrev(key) {
		__antithesis_instrumentation__.Notify(126947)

		prevGapVal := p.incomingGapVal(it, key)

		var err error
		if it.Valid() && func() bool {
			__antithesis_instrumentation__.Notify(126949)
			return bytes.Equal(it.Key(), key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(126950)

			err = arenaskl.ErrRecordExists
		} else {
			__antithesis_instrumentation__.Notify(126951)

			if _, update := ratchetValue(prevGapVal, val); !update {
				__antithesis_instrumentation__.Notify(126953)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(126954)
			}
			__antithesis_instrumentation__.Notify(126952)

			p.ratchetMaxTimestamp(val.ts)

			b, meta := encodeValueSet(arr[:0], keyVal, gapVal)
			err = it.Add(key, b, meta)
		}
		__antithesis_instrumentation__.Notify(126948)

		switch {
		case errors.Is(err, arenaskl.ErrArenaFull):
			__antithesis_instrumentation__.Notify(126955)
			atomic.StoreInt32(&p.isFull, 1)
			return err
		case errors.Is(err, arenaskl.ErrRecordExists):
			__antithesis_instrumentation__.Notify(126956)

		case err == nil:
			__antithesis_instrumentation__.Notify(126957)

			return p.ensureInitialized(it, key)
		default:
			__antithesis_instrumentation__.Notify(126958)
			panic(fmt.Sprintf("unexpected error: %v", err))
		}
	} else {
		__antithesis_instrumentation__.Notify(126959)
	}
	__antithesis_instrumentation__.Notify(126940)

	if (it.Meta()&initialized) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(126960)
		return mustInit == true
	}() == true {
		__antithesis_instrumentation__.Notify(126961)
		if err := p.ensureInitialized(it, key); err != nil {
			__antithesis_instrumentation__.Notify(126962)
			return err
		} else {
			__antithesis_instrumentation__.Notify(126963)
		}
	} else {
		__antithesis_instrumentation__.Notify(126964)
	}
	__antithesis_instrumentation__.Notify(126941)

	if opt == 0 {
		__antithesis_instrumentation__.Notify(126965)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(126966)
	}
	__antithesis_instrumentation__.Notify(126942)
	return p.ratchetValueSet(it, always, keyVal, gapVal, false)
}

func (p *sklPage) ensureInitialized(it *arenaskl.Iterator, key []byte) error {
	__antithesis_instrumentation__.Notify(126967)

	prevGapVal := p.incomingGapVal(it, key)

	if util.RaceEnabled && func() bool {
		__antithesis_instrumentation__.Notify(126969)
		return !bytes.Equal(it.Key(), key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(126970)
		panic("no node found")
	} else {
		__antithesis_instrumentation__.Notify(126971)
	}
	__antithesis_instrumentation__.Notify(126968)

	return p.ratchetValueSet(it, onlyIfUninitialized, prevGapVal, prevGapVal, true)
}

func (p *sklPage) ensureFloorValue(it *arenaskl.Iterator, to []byte, val cacheValue) bool {
	__antithesis_instrumentation__.Notify(126972)
	for it.Valid() {
		__antithesis_instrumentation__.Notify(126974)
		util.RacePreempt()

		if to != nil && func() bool {
			__antithesis_instrumentation__.Notify(126978)
			return bytes.Compare(it.Key(), to) >= 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(126979)
			break
		} else {
			__antithesis_instrumentation__.Notify(126980)
		}
		__antithesis_instrumentation__.Notify(126975)

		if atomic.LoadInt32(&p.isFull) == 1 {
			__antithesis_instrumentation__.Notify(126981)

			return false
		} else {
			__antithesis_instrumentation__.Notify(126982)
		}
		__antithesis_instrumentation__.Notify(126976)

		err := p.ratchetValueSet(it, always, val, val, false)
		switch {
		case err == nil:
			__antithesis_instrumentation__.Notify(126983)

		case errors.Is(err, arenaskl.ErrArenaFull):
			__antithesis_instrumentation__.Notify(126984)

			return false
		default:
			__antithesis_instrumentation__.Notify(126985)
			panic(fmt.Sprintf("unexpected error: %v", err))
		}
		__antithesis_instrumentation__.Notify(126977)

		it.Next()
	}
	__antithesis_instrumentation__.Notify(126973)

	return true
}

func (p *sklPage) ratchetMaxTimestamp(ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(126986)
	new := makeRatchetingTime(ts)
	for {
		__antithesis_instrumentation__.Notify(126987)
		old := ratchetingTime(atomic.LoadInt64((*int64)(&p.maxTime)))
		if new <= old {
			__antithesis_instrumentation__.Notify(126989)
			break
		} else {
			__antithesis_instrumentation__.Notify(126990)
		}
		__antithesis_instrumentation__.Notify(126988)

		if atomic.CompareAndSwapInt64((*int64)(&p.maxTime), int64(old), int64(new)) {
			__antithesis_instrumentation__.Notify(126991)
			break
		} else {
			__antithesis_instrumentation__.Notify(126992)
		}
	}
}

func (p *sklPage) getMaxTimestamp() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(126993)
	return ratchetingTime(atomic.LoadInt64((*int64)(&p.maxTime))).get()
}

type ratchetingTime int64

func makeRatchetingTime(ts hlc.Timestamp) ratchetingTime {
	__antithesis_instrumentation__.Notify(126994)

	rt := ratchetingTime(ts.WallTime)
	if ts.Logical > 0 {
		__antithesis_instrumentation__.Notify(126998)
		rt++
	} else {
		__antithesis_instrumentation__.Notify(126999)
	}
	__antithesis_instrumentation__.Notify(126995)

	if rt&1 == 1 {
		__antithesis_instrumentation__.Notify(127000)
		rt++
	} else {
		__antithesis_instrumentation__.Notify(127001)
	}
	__antithesis_instrumentation__.Notify(126996)
	if !ts.Synthetic {
		__antithesis_instrumentation__.Notify(127002)
		rt |= 1
	} else {
		__antithesis_instrumentation__.Notify(127003)
	}
	__antithesis_instrumentation__.Notify(126997)

	return rt
}

func (rt ratchetingTime) get() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(127004)
	var ts hlc.Timestamp
	ts.WallTime = int64(rt &^ 1)
	if rt&1 == 0 {
		__antithesis_instrumentation__.Notify(127006)
		ts.Synthetic = true
	} else {
		__antithesis_instrumentation__.Notify(127007)
	}
	__antithesis_instrumentation__.Notify(127005)
	return ts
}

type ratchetPolicy bool

const (
	always ratchetPolicy = false

	onlyIfUninitialized ratchetPolicy = true
)

func (p *sklPage) ratchetValueSet(
	it *arenaskl.Iterator, policy ratchetPolicy, keyVal, gapVal cacheValue, setInit bool,
) error {
	__antithesis_instrumentation__.Notify(127008)

	var arr [encodedValSize * 2]byte

	for {
		__antithesis_instrumentation__.Notify(127009)
		util.RacePreempt()

		meta := it.Meta()
		inited := (meta & initialized) != 0
		if inited && func() bool {
			__antithesis_instrumentation__.Notify(127013)
			return policy == onlyIfUninitialized == true
		}() == true {
			__antithesis_instrumentation__.Notify(127014)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(127015)
		}
		__antithesis_instrumentation__.Notify(127010)
		if (meta & cantInit) != 0 {
			__antithesis_instrumentation__.Notify(127016)

			return arenaskl.ErrArenaFull
		} else {
			__antithesis_instrumentation__.Notify(127017)
		}
		__antithesis_instrumentation__.Notify(127011)

		newMeta := meta
		updateInit := setInit && func() bool {
			__antithesis_instrumentation__.Notify(127018)
			return !inited == true
		}() == true
		if updateInit {
			__antithesis_instrumentation__.Notify(127019)
			newMeta |= initialized
		} else {
			__antithesis_instrumentation__.Notify(127020)
		}
		__antithesis_instrumentation__.Notify(127012)

		var keyValUpdate, gapValUpdate bool
		oldKeyVal, oldGapVal := decodeValueSet(it.Value(), meta)
		keyVal, keyValUpdate = ratchetValue(oldKeyVal, keyVal)
		gapVal, gapValUpdate = ratchetValue(oldGapVal, gapVal)
		updateVals := keyValUpdate || func() bool {
			__antithesis_instrumentation__.Notify(127021)
			return gapValUpdate == true
		}() == true

		if updateVals {
			__antithesis_instrumentation__.Notify(127022)

			maxTs := keyVal.ts
			maxTs.Forward(gapVal.ts)
			p.ratchetMaxTimestamp(maxTs)

			newMeta &^= (hasKey | hasGap)

			b, valMeta := encodeValueSet(arr[:0], keyVal, gapVal)
			newMeta |= valMeta

			err := it.Set(b, newMeta)
			switch {
			case err == nil:
				__antithesis_instrumentation__.Notify(127023)

				return nil
			case errors.Is(err, arenaskl.ErrRecordUpdated):
				__antithesis_instrumentation__.Notify(127024)

				continue
			case errors.Is(err, arenaskl.ErrArenaFull):
				__antithesis_instrumentation__.Notify(127025)

				atomic.StoreInt32(&p.isFull, 1)

				if !inited && func() bool {
					__antithesis_instrumentation__.Notify(127028)
					return (meta & cantInit) == 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(127029)
					err := it.SetMeta(meta | cantInit)
					switch {
					case errors.Is(err, arenaskl.ErrRecordUpdated):
						__antithesis_instrumentation__.Notify(127030)

						continue
					case errors.Is(err, arenaskl.ErrArenaFull):
						__antithesis_instrumentation__.Notify(127031)
						panic(fmt.Sprintf("SetMeta with larger meta should not return %v", err))
					default:
						__antithesis_instrumentation__.Notify(127032)
					}
				} else {
					__antithesis_instrumentation__.Notify(127033)
				}
				__antithesis_instrumentation__.Notify(127026)
				return arenaskl.ErrArenaFull
			default:
				__antithesis_instrumentation__.Notify(127027)
				panic(fmt.Sprintf("unexpected error: %v", err))
			}
		} else {
			__antithesis_instrumentation__.Notify(127034)
			if updateInit {
				__antithesis_instrumentation__.Notify(127035)

				err := it.SetMeta(newMeta)
				switch {
				case err == nil:
					__antithesis_instrumentation__.Notify(127036)

					return nil
				case errors.Is(err, arenaskl.ErrRecordUpdated):
					__antithesis_instrumentation__.Notify(127037)

					continue
				case errors.Is(err, arenaskl.ErrArenaFull):
					__antithesis_instrumentation__.Notify(127038)
					panic(fmt.Sprintf("SetMeta with larger meta should not return %v", err))
				default:
					__antithesis_instrumentation__.Notify(127039)
					panic(fmt.Sprintf("unexpected error: %v", err))
				}
			} else {
				__antithesis_instrumentation__.Notify(127040)
				return nil
			}
		}
	}
}

func (p *sklPage) maxInRange(it *arenaskl.Iterator, from, to []byte, opt rangeOptions) cacheValue {
	__antithesis_instrumentation__.Notify(127041)

	prevGapVal := p.incomingGapVal(it, from)

	if !it.Valid() {
		__antithesis_instrumentation__.Notify(127043)

		return prevGapVal
	} else {
		__antithesis_instrumentation__.Notify(127044)
		if bytes.Equal(it.Key(), from) {
			__antithesis_instrumentation__.Notify(127045)

			if (it.Meta() & initialized) != 0 {
				__antithesis_instrumentation__.Notify(127046)

				prevGapVal = cacheValue{}
			} else {
				__antithesis_instrumentation__.Notify(127047)
			}
		} else {
			__antithesis_instrumentation__.Notify(127048)

			opt &^= excludeFrom
		}
	}
	__antithesis_instrumentation__.Notify(127042)

	_, maxVal := p.scanTo(it, to, opt, prevGapVal)
	return maxVal
}

func (p *sklPage) incomingGapVal(it *arenaskl.Iterator, key []byte) cacheValue {
	__antithesis_instrumentation__.Notify(127049)

	prevInitNode(it)

	prevGapVal, _ := p.scanTo(it, key, 0, cacheValue{})
	return prevGapVal
}

func (p *sklPage) scanTo(
	it *arenaskl.Iterator, to []byte, opt rangeOptions, initGapVal cacheValue,
) (prevGapVal, maxVal cacheValue) {
	__antithesis_instrumentation__.Notify(127050)
	prevGapVal, maxVal = initGapVal, initGapVal
	first := true
	for {
		__antithesis_instrumentation__.Notify(127051)
		util.RacePreempt()

		if !it.Valid() {
			__antithesis_instrumentation__.Notify(127058)

			return
		} else {
			__antithesis_instrumentation__.Notify(127059)
		}
		__antithesis_instrumentation__.Notify(127052)

		toCmp := bytes.Compare(it.Key(), to)
		if to == nil {
			__antithesis_instrumentation__.Notify(127060)

			toCmp = -1
		} else {
			__antithesis_instrumentation__.Notify(127061)
		}
		__antithesis_instrumentation__.Notify(127053)
		if toCmp > 0 || func() bool {
			__antithesis_instrumentation__.Notify(127062)
			return (toCmp == 0 && func() bool {
				__antithesis_instrumentation__.Notify(127063)
				return (opt & excludeTo) != 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(127064)

			return
		} else {
			__antithesis_instrumentation__.Notify(127065)
		}
		__antithesis_instrumentation__.Notify(127054)

		ratchetErr := p.ratchetValueSet(it, onlyIfUninitialized,
			prevGapVal, prevGapVal, false)

		keyVal, gapVal := decodeValueSet(it.Value(), it.Meta())
		if errors.Is(ratchetErr, arenaskl.ErrArenaFull) {
			__antithesis_instrumentation__.Notify(127066)

			keyVal, _ = ratchetValue(keyVal, prevGapVal)
			gapVal, _ = ratchetValue(gapVal, prevGapVal)
		} else {
			__antithesis_instrumentation__.Notify(127067)
		}
		__antithesis_instrumentation__.Notify(127055)

		if !(first && func() bool {
			__antithesis_instrumentation__.Notify(127068)
			return (opt & excludeFrom) != 0 == true
		}() == true) {
			__antithesis_instrumentation__.Notify(127069)

			maxVal, _ = ratchetValue(maxVal, keyVal)
		} else {
			__antithesis_instrumentation__.Notify(127070)
		}
		__antithesis_instrumentation__.Notify(127056)

		if toCmp == 0 {
			__antithesis_instrumentation__.Notify(127071)

			return
		} else {
			__antithesis_instrumentation__.Notify(127072)
		}
		__antithesis_instrumentation__.Notify(127057)

		maxVal, _ = ratchetValue(maxVal, gapVal)

		prevGapVal = gapVal
		first = false
		it.Next()
	}
}

func prevInitNode(it *arenaskl.Iterator) {
	__antithesis_instrumentation__.Notify(127073)
	for {
		__antithesis_instrumentation__.Notify(127074)
		util.RacePreempt()

		if !it.Valid() {
			__antithesis_instrumentation__.Notify(127077)

			it.SeekToFirst()
			break
		} else {
			__antithesis_instrumentation__.Notify(127078)
		}
		__antithesis_instrumentation__.Notify(127075)

		if (it.Meta() & initialized) != 0 {
			__antithesis_instrumentation__.Notify(127079)

			break
		} else {
			__antithesis_instrumentation__.Notify(127080)
		}
		__antithesis_instrumentation__.Notify(127076)

		it.Prev()
	}
}

func decodeValueSet(b []byte, meta uint16) (keyVal, gapVal cacheValue) {
	__antithesis_instrumentation__.Notify(127081)
	if (meta & hasKey) != 0 {
		__antithesis_instrumentation__.Notify(127084)
		b, keyVal = decodeValue(b)
	} else {
		__antithesis_instrumentation__.Notify(127085)
	}
	__antithesis_instrumentation__.Notify(127082)

	if (meta & hasGap) != 0 {
		__antithesis_instrumentation__.Notify(127086)
		_, gapVal = decodeValue(b)
	} else {
		__antithesis_instrumentation__.Notify(127087)
	}
	__antithesis_instrumentation__.Notify(127083)

	return
}

func encodeValueSet(b []byte, keyVal, gapVal cacheValue) (ret []byte, meta uint16) {
	__antithesis_instrumentation__.Notify(127088)
	if !keyVal.ts.IsEmpty() {
		__antithesis_instrumentation__.Notify(127091)
		b = encodeValue(b, keyVal)
		meta |= hasKey
	} else {
		__antithesis_instrumentation__.Notify(127092)
	}
	__antithesis_instrumentation__.Notify(127089)

	if !gapVal.ts.IsEmpty() {
		__antithesis_instrumentation__.Notify(127093)
		b = encodeValue(b, gapVal)
		meta |= hasGap
	} else {
		__antithesis_instrumentation__.Notify(127094)
	}
	__antithesis_instrumentation__.Notify(127090)

	ret = b
	return
}

func decodeValue(b []byte) (ret []byte, val cacheValue) {
	__antithesis_instrumentation__.Notify(127095)

	valPtr := (*[encodedValSize]byte)(unsafe.Pointer(&val))
	copy(valPtr[:], b)
	ret = b[encodedValSize:]
	return ret, val
}

func encodeValue(b []byte, val cacheValue) []byte {
	__antithesis_instrumentation__.Notify(127096)

	prev := len(b)
	b = b[:prev+encodedValSize]
	valPtr := (*[encodedValSize]byte)(unsafe.Pointer(&val))
	copy(b[prev:], valPtr[:])
	return b
}

func encodedRangeSize(from, to []byte, opt rangeOptions) int {
	__antithesis_instrumentation__.Notify(127097)
	vals := 1
	if (opt & excludeTo) == 0 {
		__antithesis_instrumentation__.Notify(127100)
		vals++
	} else {
		__antithesis_instrumentation__.Notify(127101)
	}
	__antithesis_instrumentation__.Notify(127098)
	if (opt & excludeFrom) == 0 {
		__antithesis_instrumentation__.Notify(127102)
		vals++
	} else {
		__antithesis_instrumentation__.Notify(127103)
	}
	__antithesis_instrumentation__.Notify(127099)

	return len(from) + len(to) + (vals * encodedValSize) + (2 * arenaskl.MaxNodeSize)
}
