package kvevent

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "context"

type Alloc struct {
	bytes   int64
	entries int64
	ap      pool

	otherPoolAllocs map[pool]*Alloc
}

type pool interface {
	Release(ctx context.Context, bytes, entries int64)
}

func (a *Alloc) Release(ctx context.Context) {
	__antithesis_instrumentation__.Notify(16947)
	if a.isZero() {
		__antithesis_instrumentation__.Notify(16950)
		return
	} else {
		__antithesis_instrumentation__.Notify(16951)
	}
	__antithesis_instrumentation__.Notify(16948)
	for _, oa := range a.otherPoolAllocs {
		__antithesis_instrumentation__.Notify(16952)
		oa.Release(ctx)
	}
	__antithesis_instrumentation__.Notify(16949)
	a.ap.Release(ctx, a.bytes, a.entries)
	a.clear()
}

func (a *Alloc) Merge(other *Alloc) {
	__antithesis_instrumentation__.Notify(16953)
	defer other.clear()
	if a.isZero() {
		__antithesis_instrumentation__.Notify(16956)
		*a = *other
		return
	} else {
		__antithesis_instrumentation__.Notify(16957)
	}
	__antithesis_instrumentation__.Notify(16954)

	if other.otherPoolAllocs != nil {
		__antithesis_instrumentation__.Notify(16958)
		for _, oa := range other.otherPoolAllocs {
			__antithesis_instrumentation__.Notify(16960)
			a.Merge(oa)
		}
		__antithesis_instrumentation__.Notify(16959)
		other.otherPoolAllocs = nil
	} else {
		__antithesis_instrumentation__.Notify(16961)
	}
	__antithesis_instrumentation__.Notify(16955)

	if samePool := a.ap == other.ap; samePool {
		__antithesis_instrumentation__.Notify(16962)
		a.bytes += other.bytes
		a.entries += other.entries
	} else {
		__antithesis_instrumentation__.Notify(16963)

		if a.otherPoolAllocs == nil {
			__antithesis_instrumentation__.Notify(16965)
			a.otherPoolAllocs = make(map[pool]*Alloc, 1)
		} else {
			__antithesis_instrumentation__.Notify(16966)
		}
		__antithesis_instrumentation__.Notify(16964)
		if mergeAlloc, ok := a.otherPoolAllocs[other.ap]; ok {
			__antithesis_instrumentation__.Notify(16967)
			mergeAlloc.Merge(other)
		} else {
			__antithesis_instrumentation__.Notify(16968)
			otherCpy := *other
			a.otherPoolAllocs[other.ap] = &otherCpy
		}
	}
}

func (a *Alloc) clear()       { __antithesis_instrumentation__.Notify(16969); *a = Alloc{} }
func (a *Alloc) isZero() bool { __antithesis_instrumentation__.Notify(16970); return a.ap == nil }

func TestingMakeAlloc(bytes int64, p pool) Alloc {
	__antithesis_instrumentation__.Notify(16971)
	return Alloc{bytes: bytes, entries: 1, ap: p}
}
