// Package raftentry provides a cache for entries to avoid extra
// deserializations.
package raftentry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"sync/atomic"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type Cache struct {
	metrics  Metrics
	maxBytes int32

	bytes   int32
	entries int32

	mu    syncutil.Mutex
	lru   partitionList
	parts map[roachpb.RangeID]*partition
}

type partition struct {
	id roachpb.RangeID

	mu syncutil.RWMutex
	ringBuf

	size cacheSize

	next, prev *partition
}

const partitionSize = int32(unsafe.Sizeof(partition{}))

type rangeCache interface {
	add(ent []raftpb.Entry) (bytesAdded, entriesAdded int32)
	truncateFrom(lo uint64) (bytesRemoved, entriesRemoved int32)
	clearTo(hi uint64) (bytesRemoved, entriesRemoved int32)
	get(index uint64) (raftpb.Entry, bool)
	scan(ents []raftpb.Entry, lo, hi, maxBytes uint64) (
		_ []raftpb.Entry, bytes uint64, nextIdx uint64, exceededMaxBytes bool)
}

var _ rangeCache = (*ringBuf)(nil)

func NewCache(maxBytes uint64) *Cache {
	__antithesis_instrumentation__.Notify(113229)
	if maxBytes > math.MaxInt32 {
		__antithesis_instrumentation__.Notify(113231)
		maxBytes = math.MaxInt32
	} else {
		__antithesis_instrumentation__.Notify(113232)
	}
	__antithesis_instrumentation__.Notify(113230)
	return &Cache{
		maxBytes: int32(maxBytes),
		metrics:  makeMetrics(),
		parts:    map[roachpb.RangeID]*partition{},
	}
}

func (c *Cache) Metrics() Metrics {
	__antithesis_instrumentation__.Notify(113233)
	return c.metrics
}

func (c *Cache) Drop(id roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(113234)
	c.mu.Lock()
	defer c.mu.Unlock()
	p := c.getPartLocked(id, false, false)
	if p != nil {
		__antithesis_instrumentation__.Notify(113235)
		c.updateGauges(c.evictPartitionLocked(p))
	} else {
		__antithesis_instrumentation__.Notify(113236)
	}
}

func (c *Cache) Add(id roachpb.RangeID, ents []raftpb.Entry, truncate bool) {
	__antithesis_instrumentation__.Notify(113237)
	if len(ents) == 0 {
		__antithesis_instrumentation__.Notify(113244)
		return
	} else {
		__antithesis_instrumentation__.Notify(113245)
	}
	__antithesis_instrumentation__.Notify(113238)
	bytesGuessed := analyzeEntries(ents)
	add := bytesGuessed <= c.maxBytes
	if !add {
		__antithesis_instrumentation__.Notify(113246)
		bytesGuessed = 0
	} else {
		__antithesis_instrumentation__.Notify(113247)
	}
	__antithesis_instrumentation__.Notify(113239)

	c.mu.Lock()

	p := c.getPartLocked(id, add, true)
	if bytesGuessed > 0 {
		__antithesis_instrumentation__.Notify(113248)
		c.evictLocked(bytesGuessed)
		if len(c.parts) == 0 {
			__antithesis_instrumentation__.Notify(113250)
			p = c.getPartLocked(id, true, false)
		} else {
			__antithesis_instrumentation__.Notify(113251)
		}
		__antithesis_instrumentation__.Notify(113249)

		for {
			__antithesis_instrumentation__.Notify(113252)
			prev := p.loadSize()
			if p.setSize(prev, prev.add(bytesGuessed, 0)) {
				__antithesis_instrumentation__.Notify(113253)
				break
			} else {
				__antithesis_instrumentation__.Notify(113254)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(113255)
	}
	__antithesis_instrumentation__.Notify(113240)
	c.mu.Unlock()
	if p == nil {
		__antithesis_instrumentation__.Notify(113256)

		return
	} else {
		__antithesis_instrumentation__.Notify(113257)
	}
	__antithesis_instrumentation__.Notify(113241)

	p.mu.Lock()
	defer p.mu.Unlock()
	var bytesAdded, entriesAdded, bytesRemoved, entriesRemoved int32

	if truncate {
		__antithesis_instrumentation__.Notify(113258)

		truncIdx := ents[0].Index
		bytesRemoved, entriesRemoved = p.truncateFrom(truncIdx)
	} else {
		__antithesis_instrumentation__.Notify(113259)
	}
	__antithesis_instrumentation__.Notify(113242)
	if add {
		__antithesis_instrumentation__.Notify(113260)
		bytesAdded, entriesAdded = p.add(ents)
	} else {
		__antithesis_instrumentation__.Notify(113261)
	}
	__antithesis_instrumentation__.Notify(113243)
	c.recordUpdate(p, bytesAdded-bytesRemoved, bytesGuessed, entriesAdded-entriesRemoved)
}

func (c *Cache) Clear(id roachpb.RangeID, hi uint64) {
	__antithesis_instrumentation__.Notify(113262)
	c.mu.Lock()
	p := c.getPartLocked(id, false, false)
	if p == nil {
		__antithesis_instrumentation__.Notify(113264)
		c.mu.Unlock()
		return
	} else {
		__antithesis_instrumentation__.Notify(113265)
	}
	__antithesis_instrumentation__.Notify(113263)
	c.mu.Unlock()
	p.mu.Lock()
	defer p.mu.Unlock()
	bytesRemoved, entriesRemoved := p.clearTo(hi)
	c.recordUpdate(p, -1*bytesRemoved, 0, -1*entriesRemoved)
}

func (c *Cache) Get(id roachpb.RangeID, idx uint64) (e raftpb.Entry, ok bool) {
	__antithesis_instrumentation__.Notify(113266)
	c.metrics.Accesses.Inc(1)
	c.mu.Lock()
	p := c.getPartLocked(id, false, true)
	c.mu.Unlock()
	if p == nil {
		__antithesis_instrumentation__.Notify(113269)
		return e, false
	} else {
		__antithesis_instrumentation__.Notify(113270)
	}
	__antithesis_instrumentation__.Notify(113267)
	p.mu.RLock()
	defer p.mu.RUnlock()
	e, ok = p.get(idx)
	if ok {
		__antithesis_instrumentation__.Notify(113271)
		c.metrics.Hits.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(113272)
	}
	__antithesis_instrumentation__.Notify(113268)
	return e, ok
}

func (c *Cache) Scan(
	ents []raftpb.Entry, id roachpb.RangeID, lo, hi, maxBytes uint64,
) (_ []raftpb.Entry, bytes uint64, nextIdx uint64, exceededMaxBytes bool) {
	__antithesis_instrumentation__.Notify(113273)
	c.metrics.Accesses.Inc(1)
	c.mu.Lock()
	p := c.getPartLocked(id, false, true)
	c.mu.Unlock()
	if p == nil {
		__antithesis_instrumentation__.Notify(113276)
		return ents, 0, lo, false
	} else {
		__antithesis_instrumentation__.Notify(113277)
	}
	__antithesis_instrumentation__.Notify(113274)
	p.mu.RLock()
	defer p.mu.RUnlock()

	ents, bytes, nextIdx, exceededMaxBytes = p.scan(ents, lo, hi, maxBytes)
	if nextIdx == hi || func() bool {
		__antithesis_instrumentation__.Notify(113278)
		return exceededMaxBytes == true
	}() == true {
		__antithesis_instrumentation__.Notify(113279)

		c.metrics.Hits.Inc(1)
	} else {
		__antithesis_instrumentation__.Notify(113280)
	}
	__antithesis_instrumentation__.Notify(113275)
	return ents, bytes, nextIdx, exceededMaxBytes
}

func (c *Cache) getPartLocked(id roachpb.RangeID, create, recordUse bool) *partition {
	__antithesis_instrumentation__.Notify(113281)
	part := c.parts[id]
	if create && func() bool {
		__antithesis_instrumentation__.Notify(113284)
		return part == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(113285)
		part = c.lru.pushFront(id)
		c.parts[id] = part
		c.addBytes(partitionSize)
	} else {
		__antithesis_instrumentation__.Notify(113286)
	}
	__antithesis_instrumentation__.Notify(113282)
	if recordUse && func() bool {
		__antithesis_instrumentation__.Notify(113287)
		return part != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(113288)
		c.lru.moveToFront(part)
	} else {
		__antithesis_instrumentation__.Notify(113289)
	}
	__antithesis_instrumentation__.Notify(113283)
	return part
}

func (c *Cache) evictLocked(toAdd int32) {
	__antithesis_instrumentation__.Notify(113290)
	bytes := c.addBytes(toAdd)
	for bytes > c.maxBytes && func() bool {
		__antithesis_instrumentation__.Notify(113291)
		return len(c.parts) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(113292)
		bytes, _ = c.evictPartitionLocked(c.lru.back())
	}
}

func (c *Cache) evictPartitionLocked(p *partition) (updatedBytes, updatedEntries int32) {
	__antithesis_instrumentation__.Notify(113293)
	delete(c.parts, p.id)
	c.lru.remove(p)
	pBytes, pEntries := p.evict()
	return c.addBytes(-1 * pBytes), c.addEntries(-1 * pEntries)
}

func (c *Cache) recordUpdate(p *partition, bytesAdded, bytesGuessed, entriesAdded int32) {
	__antithesis_instrumentation__.Notify(113294)

	delta := bytesAdded - bytesGuessed
	for {
		__antithesis_instrumentation__.Notify(113295)
		curSize := p.loadSize()
		if curSize == evicted {
			__antithesis_instrumentation__.Notify(113297)
			return
		} else {
			__antithesis_instrumentation__.Notify(113298)
		}
		__antithesis_instrumentation__.Notify(113296)
		newSize := curSize.add(delta, entriesAdded)
		if updated := p.setSize(curSize, newSize); updated {
			__antithesis_instrumentation__.Notify(113299)
			c.updateGauges(c.addBytes(delta), c.addEntries(entriesAdded))
			return
		} else {
			__antithesis_instrumentation__.Notify(113300)
		}
	}
}

func (c *Cache) addBytes(toAdd int32) int32 {
	__antithesis_instrumentation__.Notify(113301)
	return atomic.AddInt32(&c.bytes, toAdd)
}

func (c *Cache) addEntries(toAdd int32) int32 {
	__antithesis_instrumentation__.Notify(113302)
	return atomic.AddInt32(&c.entries, toAdd)
}

func (c *Cache) updateGauges(bytes, entries int32) {
	__antithesis_instrumentation__.Notify(113303)
	c.metrics.Bytes.Update(int64(bytes))
	c.metrics.Size.Update(int64(entries))
}

var initialSize = newCacheSize(partitionSize, 0)

func newPartition(id roachpb.RangeID) *partition {
	__antithesis_instrumentation__.Notify(113304)
	return &partition{
		id:   id,
		size: initialSize,
	}
}

const evicted cacheSize = 0

func (p *partition) evict() (bytes, entries int32) {
	__antithesis_instrumentation__.Notify(113305)

	cs := p.loadSize()
	for !p.setSize(cs, evicted) {
		__antithesis_instrumentation__.Notify(113307)
		cs = p.loadSize()
	}
	__antithesis_instrumentation__.Notify(113306)
	return cs.bytes(), cs.entries()
}

func (p *partition) loadSize() cacheSize {
	__antithesis_instrumentation__.Notify(113308)
	return cacheSize(atomic.LoadUint64((*uint64)(&p.size)))
}

func (p *partition) setSize(orig, new cacheSize) bool {
	__antithesis_instrumentation__.Notify(113309)
	return atomic.CompareAndSwapUint64((*uint64)(&p.size), uint64(orig), uint64(new))
}

func analyzeEntries(ents []raftpb.Entry) (size int32) {
	__antithesis_instrumentation__.Notify(113310)
	var prevIndex uint64
	var prevTerm uint64
	for i, e := range ents {
		__antithesis_instrumentation__.Notify(113312)
		if i != 0 && func() bool {
			__antithesis_instrumentation__.Notify(113315)
			return e.Index != prevIndex+1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(113316)
			panic(errors.AssertionFailedf("invalid non-contiguous set of entries %d and %d", prevIndex, e.Index))
		} else {
			__antithesis_instrumentation__.Notify(113317)
		}
		__antithesis_instrumentation__.Notify(113313)
		if i != 0 && func() bool {
			__antithesis_instrumentation__.Notify(113318)
			return e.Term < prevTerm == true
		}() == true {
			__antithesis_instrumentation__.Notify(113319)
			err := errors.AssertionFailedf("term regression idx %d: %d -> %d", prevIndex, prevTerm, e.Term)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(113320)
		}
		__antithesis_instrumentation__.Notify(113314)
		prevIndex = e.Index
		prevTerm = e.Term
		size += int32(e.Size())
	}
	__antithesis_instrumentation__.Notify(113311)
	return
}

type cacheSize uint64

func newCacheSize(bytes, entries int32) cacheSize {
	__antithesis_instrumentation__.Notify(113321)
	return cacheSize((uint64(entries) << 32) | uint64(bytes))
}

func (cs cacheSize) entries() int32 {
	__antithesis_instrumentation__.Notify(113322)
	return int32(cs >> 32)
}

func (cs cacheSize) bytes() int32 {
	__antithesis_instrumentation__.Notify(113323)
	return int32(cs & math.MaxUint32)
}

func (cs cacheSize) add(bytes, entries int32) cacheSize {
	__antithesis_instrumentation__.Notify(113324)
	return newCacheSize(cs.bytes()+bytes, cs.entries()+entries)
}

type partitionList struct {
	root partition
}

func (l *partitionList) lazyInit() {
	__antithesis_instrumentation__.Notify(113325)
	if l.root.next == nil {
		__antithesis_instrumentation__.Notify(113326)
		l.root.next = &l.root
		l.root.prev = &l.root
	} else {
		__antithesis_instrumentation__.Notify(113327)
	}
}

func (l *partitionList) pushFront(id roachpb.RangeID) *partition {
	__antithesis_instrumentation__.Notify(113328)
	l.lazyInit()
	return l.insert(newPartition(id), &l.root)
}

func (l *partitionList) moveToFront(p *partition) {
	__antithesis_instrumentation__.Notify(113329)
	l.insert(l.remove(p), &l.root)
}

func (l *partitionList) insert(e, at *partition) *partition {
	__antithesis_instrumentation__.Notify(113330)
	n := at.next
	at.next = e
	e.prev = at
	e.next = n
	n.prev = e
	return e
}

func (l *partitionList) back() *partition {
	__antithesis_instrumentation__.Notify(113331)
	if l.root.prev == nil || func() bool {
		__antithesis_instrumentation__.Notify(113333)
		return l.root.prev == &l.root == true
	}() == true {
		__antithesis_instrumentation__.Notify(113334)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(113335)
	}
	__antithesis_instrumentation__.Notify(113332)
	return l.root.prev
}

func (l *partitionList) remove(e *partition) *partition {
	__antithesis_instrumentation__.Notify(113336)
	if e == &l.root {
		__antithesis_instrumentation__.Notify(113339)
		panic("cannot remove root list node")
	} else {
		__antithesis_instrumentation__.Notify(113340)
	}
	__antithesis_instrumentation__.Notify(113337)
	if e.next != nil {
		__antithesis_instrumentation__.Notify(113341)
		e.prev.next = e.next
		e.next.prev = e.prev
		e.next = nil
		e.prev = nil
	} else {
		__antithesis_instrumentation__.Notify(113342)
	}
	__antithesis_instrumentation__.Notify(113338)
	return e
}
