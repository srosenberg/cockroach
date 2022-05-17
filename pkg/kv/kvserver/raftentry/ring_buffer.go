package raftentry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/bits"

	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type ringBuf struct {
	buf  []raftpb.Entry
	head int
	len  int
}

const (
	shrinkThreshold = 8
	minBufSize      = 16
)

func (b *ringBuf) add(ents []raftpb.Entry) (addedBytes, addedEntries int32) {
	__antithesis_instrumentation__.Notify(113344)
	if it := last(b); it.valid(b) && func() bool {
		__antithesis_instrumentation__.Notify(113349)
		return ents[0].Index > it.index(b)+1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(113350)

		removedBytes, removedEntries := b.clearTo(it.index(b) + 1)
		addedBytes, addedEntries = -1*removedBytes, -1*removedEntries
	} else {
		__antithesis_instrumentation__.Notify(113351)
	}
	__antithesis_instrumentation__.Notify(113345)
	before, after, ok := computeExtension(b, ents[0].Index, ents[len(ents)-1].Index)
	if !ok {
		__antithesis_instrumentation__.Notify(113352)
		return
	} else {
		__antithesis_instrumentation__.Notify(113353)
	}
	__antithesis_instrumentation__.Notify(113346)
	extend(b, before, after)
	it := first(b)
	if before == 0 && func() bool {
		__antithesis_instrumentation__.Notify(113354)
		return after != b.len == true
	}() == true {
		__antithesis_instrumentation__.Notify(113355)
		it, _ = iterateFrom(b, ents[0].Index)
	} else {
		__antithesis_instrumentation__.Notify(113356)
	}
	__antithesis_instrumentation__.Notify(113347)
	firstNewAfter := len(ents) - after
	for i, e := range ents {
		__antithesis_instrumentation__.Notify(113357)
		if i < before || func() bool {
			__antithesis_instrumentation__.Notify(113359)
			return i >= firstNewAfter == true
		}() == true {
			__antithesis_instrumentation__.Notify(113360)
			addedEntries++
			addedBytes += int32(e.Size())
		} else {
			__antithesis_instrumentation__.Notify(113361)
			addedBytes += int32(e.Size() - it.entry(b).Size())
		}
		__antithesis_instrumentation__.Notify(113358)
		it = it.push(b, e)
	}
	__antithesis_instrumentation__.Notify(113348)
	return
}

func (b *ringBuf) truncateFrom(lo uint64) (removedBytes, removedEntries int32) {
	__antithesis_instrumentation__.Notify(113362)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(113368)
		return
	} else {
		__antithesis_instrumentation__.Notify(113369)
	}
	__antithesis_instrumentation__.Notify(113363)
	if idx := first(b).index(b); idx > lo {
		__antithesis_instrumentation__.Notify(113370)

		lo = idx
	} else {
		__antithesis_instrumentation__.Notify(113371)
	}
	__antithesis_instrumentation__.Notify(113364)
	it, ok := iterateFrom(b, lo)
	for ok {
		__antithesis_instrumentation__.Notify(113372)
		removedBytes += int32(it.entry(b).Size())
		removedEntries++
		it.clear(b)
		it, ok = it.next(b)
	}
	__antithesis_instrumentation__.Notify(113365)
	b.len -= int(removedEntries)
	if b.len < (len(b.buf) / shrinkThreshold) {
		__antithesis_instrumentation__.Notify(113373)
		realloc(b, 0, b.len)
	} else {
		__antithesis_instrumentation__.Notify(113374)
	}
	__antithesis_instrumentation__.Notify(113366)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(113375)
		if b.len > 0 {
			__antithesis_instrumentation__.Notify(113376)
			if lastIdx := last(b).index(b); lastIdx >= lo {
				__antithesis_instrumentation__.Notify(113377)
				panic(errors.AssertionFailedf(
					"buffer truncated to [..., %d], but current last index is %d",
					lo, lastIdx,
				))
			} else {
				__antithesis_instrumentation__.Notify(113378)
			}
		} else {
			__antithesis_instrumentation__.Notify(113379)
		}
	} else {
		__antithesis_instrumentation__.Notify(113380)
	}
	__antithesis_instrumentation__.Notify(113367)
	return removedBytes, removedEntries
}

func (b *ringBuf) clearTo(hi uint64) (removedBytes, removedEntries int32) {
	__antithesis_instrumentation__.Notify(113381)
	if b.len == 0 || func() bool {
		__antithesis_instrumentation__.Notify(113386)
		return hi < first(b).index(b) == true
	}() == true {
		__antithesis_instrumentation__.Notify(113387)
		return
	} else {
		__antithesis_instrumentation__.Notify(113388)
	}
	__antithesis_instrumentation__.Notify(113382)
	it := first(b)
	ok := it.valid(b)
	firstIndex := it.index(b)
	for ok && func() bool {
		__antithesis_instrumentation__.Notify(113389)
		return it.index(b) < hi == true
	}() == true {
		__antithesis_instrumentation__.Notify(113390)
		removedBytes += int32(it.entry(b).Size())
		removedEntries++
		it.clear(b)
		it, ok = it.next(b)
	}
	__antithesis_instrumentation__.Notify(113383)
	offset := int(hi - firstIndex)
	if offset > b.len {
		__antithesis_instrumentation__.Notify(113391)
		offset = b.len
	} else {
		__antithesis_instrumentation__.Notify(113392)
	}
	__antithesis_instrumentation__.Notify(113384)
	b.len = b.len - offset
	b.head = (b.head + offset) % len(b.buf)
	if b.len < (len(b.buf) / shrinkThreshold) {
		__antithesis_instrumentation__.Notify(113393)
		realloc(b, 0, b.len)
	} else {
		__antithesis_instrumentation__.Notify(113394)
	}
	__antithesis_instrumentation__.Notify(113385)
	return
}

func (b *ringBuf) get(index uint64) (e raftpb.Entry, ok bool) {
	__antithesis_instrumentation__.Notify(113395)
	it, ok := iterateFrom(b, index)
	if !ok {
		__antithesis_instrumentation__.Notify(113397)
		return e, ok
	} else {
		__antithesis_instrumentation__.Notify(113398)
	}
	__antithesis_instrumentation__.Notify(113396)
	return *it.entry(b), ok
}

func (b *ringBuf) scan(
	ents []raftpb.Entry, lo, hi, maxBytes uint64,
) (_ []raftpb.Entry, bytes uint64, nextIdx uint64, exceededMaxBytes bool) {
	__antithesis_instrumentation__.Notify(113399)
	var it iterator
	nextIdx = lo
	it, ok := iterateFrom(b, lo)
	for ok && func() bool {
		__antithesis_instrumentation__.Notify(113401)
		return !exceededMaxBytes == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(113402)
		return it.index(b) < hi == true
	}() == true {
		__antithesis_instrumentation__.Notify(113403)
		e := it.entry(b)
		s := uint64(e.Size())
		exceededMaxBytes = bytes+s > maxBytes
		if exceededMaxBytes && func() bool {
			__antithesis_instrumentation__.Notify(113405)
			return len(ents) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(113406)
			break
		} else {
			__antithesis_instrumentation__.Notify(113407)
		}
		__antithesis_instrumentation__.Notify(113404)
		bytes += s
		ents = append(ents, *e)
		nextIdx++
		it, ok = it.next(b)
	}
	__antithesis_instrumentation__.Notify(113400)
	return ents, bytes, nextIdx, exceededMaxBytes
}

func realloc(b *ringBuf, before, newLen int) {
	__antithesis_instrumentation__.Notify(113408)
	newBuf := make([]raftpb.Entry, reallocLen(newLen))
	if b.head+b.len > len(b.buf) {
		__antithesis_instrumentation__.Notify(113410)
		n := copy(newBuf[before:], b.buf[b.head:])
		copy(newBuf[before+n:], b.buf[:(b.head+b.len)%len(b.buf)])
	} else {
		__antithesis_instrumentation__.Notify(113411)
		copy(newBuf[before:], b.buf[b.head:b.head+b.len])
	}
	__antithesis_instrumentation__.Notify(113409)
	b.buf = newBuf
	b.head = 0
	b.len = newLen
}

func reallocLen(need int) (newLen int) {
	__antithesis_instrumentation__.Notify(113412)
	if need <= minBufSize {
		__antithesis_instrumentation__.Notify(113414)
		return minBufSize
	} else {
		__antithesis_instrumentation__.Notify(113415)
	}
	__antithesis_instrumentation__.Notify(113413)
	return 1 << uint(bits.Len(uint(need)))
}

func extend(b *ringBuf, before, after int) {
	__antithesis_instrumentation__.Notify(113416)
	size := before + b.len + after
	if size > len(b.buf) {
		__antithesis_instrumentation__.Notify(113418)
		realloc(b, before, size)
	} else {
		__antithesis_instrumentation__.Notify(113419)
		b.head = (b.head - before) % len(b.buf)
		if b.head < 0 {
			__antithesis_instrumentation__.Notify(113420)
			b.head += len(b.buf)
		} else {
			__antithesis_instrumentation__.Notify(113421)
		}
	}
	__antithesis_instrumentation__.Notify(113417)
	b.len = size
}

func computeExtension(b *ringBuf, lo, hi uint64) (before, after int, ok bool) {
	__antithesis_instrumentation__.Notify(113422)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(113427)
		return 0, int(hi) - int(lo) + 1, true
	} else {
		__antithesis_instrumentation__.Notify(113428)
	}
	__antithesis_instrumentation__.Notify(113423)
	first, last := first(b).index(b), last(b).index(b)
	if lo > (last+1) || func() bool {
		__antithesis_instrumentation__.Notify(113429)
		return hi < (first - 1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(113430)
		return 0, 0, false
	} else {
		__antithesis_instrumentation__.Notify(113431)
	}
	__antithesis_instrumentation__.Notify(113424)
	if lo < first {
		__antithesis_instrumentation__.Notify(113432)
		before = int(first) - int(lo)
	} else {
		__antithesis_instrumentation__.Notify(113433)
	}
	__antithesis_instrumentation__.Notify(113425)
	if hi > last {
		__antithesis_instrumentation__.Notify(113434)
		after = int(hi) - int(last)
	} else {
		__antithesis_instrumentation__.Notify(113435)
	}
	__antithesis_instrumentation__.Notify(113426)
	return before, after, true
}

type iterator int

func iterateFrom(b *ringBuf, index uint64) (_ iterator, ok bool) {
	__antithesis_instrumentation__.Notify(113436)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(113439)
		return -1, false
	} else {
		__antithesis_instrumentation__.Notify(113440)
	}
	__antithesis_instrumentation__.Notify(113437)
	offset := int(index) - int(first(b).index(b))
	if offset < 0 || func() bool {
		__antithesis_instrumentation__.Notify(113441)
		return offset >= b.len == true
	}() == true {
		__antithesis_instrumentation__.Notify(113442)
		return -1, false
	} else {
		__antithesis_instrumentation__.Notify(113443)
	}
	__antithesis_instrumentation__.Notify(113438)
	return iterator((b.head + offset) % len(b.buf)), true
}

func first(b *ringBuf) iterator {
	__antithesis_instrumentation__.Notify(113444)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(113446)
		return iterator(-1)
	} else {
		__antithesis_instrumentation__.Notify(113447)
	}
	__antithesis_instrumentation__.Notify(113445)
	return iterator(b.head)
}

func last(b *ringBuf) iterator {
	__antithesis_instrumentation__.Notify(113448)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(113450)
		return iterator(-1)
	} else {
		__antithesis_instrumentation__.Notify(113451)
	}
	__antithesis_instrumentation__.Notify(113449)
	return iterator((b.head + b.len - 1) % len(b.buf))
}

func (it iterator) valid(b *ringBuf) bool {
	__antithesis_instrumentation__.Notify(113452)
	return it >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(113453)
		return int(it) < len(b.buf) == true
	}() == true
}

func (it iterator) index(b *ringBuf) uint64 {
	__antithesis_instrumentation__.Notify(113454)
	return b.buf[it].Index
}

func (it iterator) entry(b *ringBuf) *raftpb.Entry {
	__antithesis_instrumentation__.Notify(113455)
	return &b.buf[it]
}

func (it iterator) clear(b *ringBuf) {
	__antithesis_instrumentation__.Notify(113456)
	b.buf[it] = raftpb.Entry{}
}

func (it iterator) next(b *ringBuf) (_ iterator, ok bool) {
	__antithesis_instrumentation__.Notify(113457)
	if !it.valid(b) || func() bool {
		__antithesis_instrumentation__.Notify(113459)
		return it == last(b) == true
	}() == true {
		__antithesis_instrumentation__.Notify(113460)
		return -1, false
	} else {
		__antithesis_instrumentation__.Notify(113461)
	}
	__antithesis_instrumentation__.Notify(113458)
	return iterator(int(it+1) % len(b.buf)), true
}

func (it iterator) push(b *ringBuf, e raftpb.Entry) iterator {
	__antithesis_instrumentation__.Notify(113462)
	b.buf[it] = e
	it, _ = it.next(b)
	return it
}
