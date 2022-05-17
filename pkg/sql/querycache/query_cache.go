package querycache

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
)

type C struct {
	totalMem int64

	mu struct {
		syncutil.Mutex

		availableMem int64

		used, free entry

		m map[string]*entry
	}
}

const avgCachedSize = 1024

const maxCachedSize = 128 * 1024

type CachedData struct {
	SQL  string
	Memo *memo.Memo

	PrepareMetadata *PrepareMetadata

	IsCorrelated bool
}

func (cd *CachedData) memoryEstimate() int64 {
	__antithesis_instrumentation__.Notify(563687)
	res := int64(len(cd.SQL)) + cd.Memo.MemoryEstimate()
	if cd.PrepareMetadata != nil {
		__antithesis_instrumentation__.Notify(563689)
		res += cd.PrepareMetadata.MemoryEstimate()
	} else {
		__antithesis_instrumentation__.Notify(563690)
	}
	__antithesis_instrumentation__.Notify(563688)
	return res
}

type entry struct {
	CachedData

	prev, next *entry
}

func (e *entry) clear() {
	__antithesis_instrumentation__.Notify(563691)
	e.CachedData = CachedData{}
}

func (e *entry) remove() {
	__antithesis_instrumentation__.Notify(563692)
	e.prev.next = e.next
	e.next.prev = e.prev
	e.prev = nil
	e.next = nil
}

func (e *entry) insertAfter(a *entry) {
	__antithesis_instrumentation__.Notify(563693)
	b := a.next

	e.prev = a
	e.next = b

	a.next = e
	b.prev = e
}

func New(memorySize int64) *C {
	__antithesis_instrumentation__.Notify(563694)
	if memorySize < avgCachedSize {
		__antithesis_instrumentation__.Notify(563697)
		memorySize = avgCachedSize
	} else {
		__antithesis_instrumentation__.Notify(563698)
	}
	__antithesis_instrumentation__.Notify(563695)
	numEntries := memorySize / avgCachedSize
	c := &C{totalMem: memorySize}
	c.mu.availableMem = memorySize
	c.mu.m = make(map[string]*entry, numEntries)
	entries := make([]entry, numEntries)

	c.mu.used.next = &c.mu.used
	c.mu.used.prev = &c.mu.used

	c.mu.free.next = &entries[0]
	c.mu.free.prev = &entries[numEntries-1]
	for i := range entries {
		__antithesis_instrumentation__.Notify(563699)
		if i > 0 {
			__antithesis_instrumentation__.Notify(563701)
			entries[i].prev = &entries[i-1]
		} else {
			__antithesis_instrumentation__.Notify(563702)
			entries[i].prev = &c.mu.free
		}
		__antithesis_instrumentation__.Notify(563700)
		if i+1 < len(entries) {
			__antithesis_instrumentation__.Notify(563703)
			entries[i].next = &entries[i+1]
		} else {
			__antithesis_instrumentation__.Notify(563704)
			entries[i].next = &c.mu.free
		}
	}
	__antithesis_instrumentation__.Notify(563696)
	return c
}

func (c *C) Find(session *Session, sql string) (_ CachedData, ok bool) {
	__antithesis_instrumentation__.Notify(563705)
	c.mu.Lock()
	defer c.mu.Unlock()

	e := c.mu.m[sql]
	if e == nil {
		__antithesis_instrumentation__.Notify(563707)
		session.registerMiss()
		return CachedData{}, false
	} else {
		__antithesis_instrumentation__.Notify(563708)
	}
	__antithesis_instrumentation__.Notify(563706)
	session.registerHit()

	e.remove()
	e.insertAfter(&c.mu.used)
	return e.CachedData, true
}

func (c *C) Add(session *Session, d *CachedData) {
	__antithesis_instrumentation__.Notify(563709)
	if session.highMissRatio() {
		__antithesis_instrumentation__.Notify(563713)

		if session.r == nil {
			__antithesis_instrumentation__.Notify(563715)
			session.r = rand.New(rand.NewSource(1))
		} else {
			__antithesis_instrumentation__.Notify(563716)
		}
		__antithesis_instrumentation__.Notify(563714)
		if session.r.Intn(100) != 0 {
			__antithesis_instrumentation__.Notify(563717)
			return
		} else {
			__antithesis_instrumentation__.Notify(563718)
		}
	} else {
		__antithesis_instrumentation__.Notify(563719)
	}
	__antithesis_instrumentation__.Notify(563710)
	mem := d.memoryEstimate()
	if d.SQL == "" || func() bool {
		__antithesis_instrumentation__.Notify(563720)
		return mem > maxCachedSize == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(563721)
		return mem > c.totalMem == true
	}() == true {
		__antithesis_instrumentation__.Notify(563722)
		return
	} else {
		__antithesis_instrumentation__.Notify(563723)
	}
	__antithesis_instrumentation__.Notify(563711)

	c.mu.Lock()
	defer c.mu.Unlock()

	e, ok := c.mu.m[d.SQL]
	if ok {
		__antithesis_instrumentation__.Notify(563724)

		e.remove()
		c.mu.availableMem += e.memoryEstimate()
	} else {
		__antithesis_instrumentation__.Notify(563725)

		e = c.getEntry()
		c.mu.m[d.SQL] = e
	}
	__antithesis_instrumentation__.Notify(563712)

	e.CachedData = *d

	c.makeSpace(mem)
	c.mu.availableMem -= mem

	e.insertAfter(&c.mu.used)
}

func (c *C) makeSpace(needed int64) {
	__antithesis_instrumentation__.Notify(563726)
	for c.mu.availableMem < needed {
		__antithesis_instrumentation__.Notify(563727)

		c.evict().insertAfter(&c.mu.free)
	}
}

func (c *C) evict() *entry {
	__antithesis_instrumentation__.Notify(563728)
	e := c.mu.used.prev
	if e == &c.mu.used {
		__antithesis_instrumentation__.Notify(563730)
		panic("no more used entries")
	} else {
		__antithesis_instrumentation__.Notify(563731)
	}
	__antithesis_instrumentation__.Notify(563729)
	e.remove()
	c.mu.availableMem += e.memoryEstimate()
	delete(c.mu.m, e.SQL)
	e.clear()

	return e
}

func (c *C) getEntry() *entry {
	__antithesis_instrumentation__.Notify(563732)
	if e := c.mu.free.next; e != &c.mu.free {
		__antithesis_instrumentation__.Notify(563734)
		e.remove()
		return e
	} else {
		__antithesis_instrumentation__.Notify(563735)
	}
	__antithesis_instrumentation__.Notify(563733)

	return c.evict()
}

func (c *C) Clear() {
	__antithesis_instrumentation__.Notify(563736)
	c.mu.Lock()
	defer c.mu.Unlock()

	for sql, e := range c.mu.m {
		__antithesis_instrumentation__.Notify(563737)

		c.mu.availableMem += e.memoryEstimate()
		delete(c.mu.m, sql)
		e.remove()
		e.clear()
		e.insertAfter(&c.mu.free)
	}
}

func (c *C) Purge(sql string) {
	__antithesis_instrumentation__.Notify(563738)
	c.mu.Lock()
	defer c.mu.Unlock()

	if e := c.mu.m[sql]; e != nil {
		__antithesis_instrumentation__.Notify(563739)
		c.mu.availableMem += e.memoryEstimate()
		delete(c.mu.m, sql)
		e.clear()
		e.remove()
		e.insertAfter(&c.mu.free)
	} else {
		__antithesis_instrumentation__.Notify(563740)
	}
}

func (c *C) check() {
	__antithesis_instrumentation__.Notify(563741)
	c.mu.Lock()
	defer c.mu.Unlock()

	numUsed := 0
	memUsed := int64(0)
	for e := c.mu.used.next; e != &c.mu.used; e = e.next {
		__antithesis_instrumentation__.Notify(563744)
		numUsed++
		memUsed += e.memoryEstimate()
		if e.SQL == "" {
			__antithesis_instrumentation__.Notify(563746)
			panic(errors.AssertionFailedf("used entry with empty SQL"))
		} else {
			__antithesis_instrumentation__.Notify(563747)
		}
		__antithesis_instrumentation__.Notify(563745)
		if me, ok := c.mu.m[e.SQL]; !ok {
			__antithesis_instrumentation__.Notify(563748)
			panic(errors.AssertionFailedf("used entry %s not in map", e.SQL))
		} else {
			__antithesis_instrumentation__.Notify(563749)
			if e != me {
				__antithesis_instrumentation__.Notify(563750)
				panic(errors.AssertionFailedf("map entry for %s doesn't match used entry", e.SQL))
			} else {
				__antithesis_instrumentation__.Notify(563751)
			}
		}
	}
	__antithesis_instrumentation__.Notify(563742)

	if numUsed != len(c.mu.m) {
		__antithesis_instrumentation__.Notify(563752)
		panic(errors.AssertionFailedf("map length %d doesn't match used list size %d", len(c.mu.m), numUsed))
	} else {
		__antithesis_instrumentation__.Notify(563753)
	}
	__antithesis_instrumentation__.Notify(563743)

	if memUsed+c.mu.availableMem != c.totalMem {
		__antithesis_instrumentation__.Notify(563754)
		panic(errors.AssertionFailedf(
			"memory usage doesn't add up: used=%d available=%d total=%d",
			memUsed, c.mu.availableMem, c.totalMem,
		))
	} else {
		__antithesis_instrumentation__.Notify(563755)
	}
}

type Session struct {
	missRatioMMA int64

	r *rand.Rand
}

const mmaN = 1024
const mmaScale = 1000000000

func (s *Session) Init() {
	__antithesis_instrumentation__.Notify(563756)
	s.missRatioMMA = 0
	s.r = nil
}

func (s *Session) registerHit() {
	__antithesis_instrumentation__.Notify(563757)
	s.missRatioMMA = s.missRatioMMA * (mmaN - 1) / mmaN
}

func (s *Session) registerMiss() {
	__antithesis_instrumentation__.Notify(563758)
	s.missRatioMMA = (s.missRatioMMA*(mmaN-1) + mmaScale) / mmaN
}

func (s *Session) highMissRatio() bool {
	__antithesis_instrumentation__.Notify(563759)
	const threshold = mmaScale * 80 / 100
	return s.missRatioMMA > threshold
}
