package gc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
)

type gcIterator struct {
	it   *rditer.ReplicaMVCCDataIterator
	done bool
	err  error
	buf  gcIteratorRingBuf
}

func makeGCIterator(desc *roachpb.RangeDescriptor, snap storage.Reader) gcIterator {
	__antithesis_instrumentation__.Notify(101452)
	return gcIterator{
		it: rditer.NewReplicaMVCCDataIterator(desc, snap, true),
	}
}

type gcIteratorState struct {
	cur, next, afterNext *storage.MVCCKeyValue
}

func (s *gcIteratorState) curIsNewest() bool {
	__antithesis_instrumentation__.Notify(101453)
	return s.cur.Key.IsValue() && func() bool {
		__antithesis_instrumentation__.Notify(101454)
		return (s.next == nil || func() bool {
			__antithesis_instrumentation__.Notify(101455)
			return (s.afterNext != nil && func() bool {
				__antithesis_instrumentation__.Notify(101456)
				return !s.afterNext.Key.IsValue() == true
			}() == true) == true
		}() == true) == true
	}() == true
}

func (s *gcIteratorState) curIsNotValue() bool {
	__antithesis_instrumentation__.Notify(101457)
	return !s.cur.Key.IsValue()
}

func (s *gcIteratorState) curIsIntent() bool {
	__antithesis_instrumentation__.Notify(101458)
	return s.next != nil && func() bool {
		__antithesis_instrumentation__.Notify(101459)
		return !s.next.Key.IsValue() == true
	}() == true
}

func (it *gcIterator) state() (s gcIteratorState, ok bool) {
	__antithesis_instrumentation__.Notify(101460)

	s.cur, ok = it.peekAt(0)
	if !ok {
		__antithesis_instrumentation__.Notify(101466)
		return gcIteratorState{}, false
	} else {
		__antithesis_instrumentation__.Notify(101467)
	}
	__antithesis_instrumentation__.Notify(101461)
	next, ok := it.peekAt(1)
	if !ok && func() bool {
		__antithesis_instrumentation__.Notify(101468)
		return it.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(101469)
		return gcIteratorState{}, false
	} else {
		__antithesis_instrumentation__.Notify(101470)
	}
	__antithesis_instrumentation__.Notify(101462)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(101471)
		return !next.Key.Key.Equal(s.cur.Key.Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(101472)
		return s, true
	} else {
		__antithesis_instrumentation__.Notify(101473)
	}
	__antithesis_instrumentation__.Notify(101463)
	s.next = next
	afterNext, ok := it.peekAt(2)
	if !ok && func() bool {
		__antithesis_instrumentation__.Notify(101474)
		return it.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(101475)
		return gcIteratorState{}, false
	} else {
		__antithesis_instrumentation__.Notify(101476)
	}
	__antithesis_instrumentation__.Notify(101464)
	if !ok || func() bool {
		__antithesis_instrumentation__.Notify(101477)
		return !afterNext.Key.Key.Equal(s.cur.Key.Key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(101478)
		return s, true
	} else {
		__antithesis_instrumentation__.Notify(101479)
	}
	__antithesis_instrumentation__.Notify(101465)
	s.afterNext = afterNext
	return s, true
}

func (it *gcIterator) step() {
	__antithesis_instrumentation__.Notify(101480)
	it.buf.removeFront()
}

func (it *gcIterator) peekAt(i int) (*storage.MVCCKeyValue, bool) {
	__antithesis_instrumentation__.Notify(101481)
	if it.buf.len <= i {
		__antithesis_instrumentation__.Notify(101483)
		if !it.fillTo(i + 1) {
			__antithesis_instrumentation__.Notify(101484)
			return nil, false
		} else {
			__antithesis_instrumentation__.Notify(101485)
		}
	} else {
		__antithesis_instrumentation__.Notify(101486)
	}
	__antithesis_instrumentation__.Notify(101482)
	return it.buf.at(i), true
}

func (it *gcIterator) fillTo(targetLen int) (ok bool) {
	__antithesis_instrumentation__.Notify(101487)
	for it.buf.len < targetLen {
		__antithesis_instrumentation__.Notify(101489)
		if ok, err := it.it.Valid(); !ok {
			__antithesis_instrumentation__.Notify(101491)
			it.err, it.done = err, err == nil
			return false
		} else {
			__antithesis_instrumentation__.Notify(101492)
		}
		__antithesis_instrumentation__.Notify(101490)
		it.buf.pushBack(it.it)
		it.it.Prev()
	}
	__antithesis_instrumentation__.Notify(101488)
	return true
}

func (it *gcIterator) close() {
	__antithesis_instrumentation__.Notify(101493)
	it.it.Close()
	it.it = nil
}

const gcIteratorRingBufSize = 3

type gcIteratorRingBuf struct {
	allocs [gcIteratorRingBufSize]bufalloc.ByteAllocator
	buf    [gcIteratorRingBufSize]storage.MVCCKeyValue
	len    int
	head   int
}

func (b *gcIteratorRingBuf) at(i int) *storage.MVCCKeyValue {
	__antithesis_instrumentation__.Notify(101494)
	if i >= b.len {
		__antithesis_instrumentation__.Notify(101496)
		panic("index out of range")
	} else {
		__antithesis_instrumentation__.Notify(101497)
	}
	__antithesis_instrumentation__.Notify(101495)
	return &b.buf[(b.head+i)%gcIteratorRingBufSize]
}

func (b *gcIteratorRingBuf) removeFront() {
	__antithesis_instrumentation__.Notify(101498)
	if b.len == 0 {
		__antithesis_instrumentation__.Notify(101500)
		panic("cannot remove from empty gcIteratorRingBuf")
	} else {
		__antithesis_instrumentation__.Notify(101501)
	}
	__antithesis_instrumentation__.Notify(101499)
	b.buf[b.head] = storage.MVCCKeyValue{}
	b.head = (b.head + 1) % gcIteratorRingBufSize
	b.len--
}

type iterator interface {
	UnsafeKey() storage.MVCCKey
	UnsafeValue() []byte
}

func (b *gcIteratorRingBuf) pushBack(it iterator) {
	__antithesis_instrumentation__.Notify(101502)
	if b.len == gcIteratorRingBufSize {
		__antithesis_instrumentation__.Notify(101504)
		panic("cannot add to full gcIteratorRingBuf")
	} else {
		__antithesis_instrumentation__.Notify(101505)
	}
	__antithesis_instrumentation__.Notify(101503)
	i := (b.head + b.len) % gcIteratorRingBufSize
	b.allocs[i] = b.allocs[i].Truncate()
	k := it.UnsafeKey()
	v := it.UnsafeValue()
	b.allocs[i], k.Key = b.allocs[i].Copy(k.Key, len(v))
	b.allocs[i], v = b.allocs[i].Copy(v, 0)
	b.buf[i] = storage.MVCCKeyValue{
		Key:   k,
		Value: v,
	}
	b.len++
}
