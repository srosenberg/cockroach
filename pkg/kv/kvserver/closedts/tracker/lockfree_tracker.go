package tracker

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type lockfreeTracker struct {
	buckets [2]bucket
	b1, b2  *bucket
}

func NewLockfreeTracker() Tracker {
	__antithesis_instrumentation__.Notify(98901)
	t := lockfreeTracker{}
	t.b1 = &t.buckets[0]
	t.b2 = &t.buckets[1]
	return &t
}

func (t *lockfreeTracker) String() string {
	__antithesis_instrumentation__.Notify(98902)
	return fmt.Sprintf("b1: %s; b2: %s", t.b1, t.b2)
}

func (t *lockfreeTracker) Track(ctx context.Context, ts hlc.Timestamp) RemovalToken {
	__antithesis_instrumentation__.Notify(98903)

	b1, b2 := t.b1, t.b2

	wts := ts.WallTime

	t1, initialized := b1.timestamp()

	if !initialized || func() bool {
		__antithesis_instrumentation__.Notify(98905)
		return wts <= t1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(98906)
		return b1.extendAndJoin(ctx, wts, ts.Synthetic)
	} else {
		__antithesis_instrumentation__.Notify(98907)
	}
	__antithesis_instrumentation__.Notify(98904)

	return b2.extendAndJoin(ctx, wts, ts.Synthetic)
}

func (t *lockfreeTracker) Untrack(ctx context.Context, tok RemovalToken) {
	__antithesis_instrumentation__.Notify(98908)
	b := tok.(lockfreeToken).b

	b.refcnt--
	if b.refcnt < 0 {
		__antithesis_instrumentation__.Notify(98910)
		log.Fatalf(ctx, "negative bucket refcount: %d", b.refcnt)
	} else {
		__antithesis_instrumentation__.Notify(98911)
	}
	__antithesis_instrumentation__.Notify(98909)
	if b.refcnt == 0 {
		__antithesis_instrumentation__.Notify(98912)

		b.ts = 0
		b.synthetic = 0

		if b == t.b1 {
			__antithesis_instrumentation__.Notify(98913)
			t.b1 = t.b2
			t.b2 = b
		} else {
			__antithesis_instrumentation__.Notify(98914)
		}
	} else {
		__antithesis_instrumentation__.Notify(98915)
	}
}

func (t *lockfreeTracker) LowerBound(ctx context.Context) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(98916)

	ts, initialized := t.b1.timestamp()
	if !initialized {
		__antithesis_instrumentation__.Notify(98918)
		return hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(98919)
	}
	__antithesis_instrumentation__.Notify(98917)
	return hlc.Timestamp{
		WallTime:  ts,
		Logical:   0,
		Synthetic: t.b1.isSynthetic(),
	}
}

func (t *lockfreeTracker) Count() int {
	__antithesis_instrumentation__.Notify(98920)
	return int(t.b1.refcnt) + int(t.b2.refcnt)
}

type bucket struct {
	ts        int64
	refcnt    int32
	synthetic int32
}

func (b *bucket) String() string {
	__antithesis_instrumentation__.Notify(98921)
	ts := atomic.LoadInt64(&b.ts)
	if ts == 0 {
		__antithesis_instrumentation__.Notify(98923)
		return "uninit"
	} else {
		__antithesis_instrumentation__.Notify(98924)
	}
	__antithesis_instrumentation__.Notify(98922)
	refcnt := atomic.LoadInt32(&b.refcnt)
	return fmt.Sprintf("%d requests, lower bound: %s", refcnt, timeutil.Unix(0, ts))
}

func (b *bucket) timestamp() (int64, bool) {
	__antithesis_instrumentation__.Notify(98925)
	ts := atomic.LoadInt64(&b.ts)
	return ts, ts != 0
}

func (b *bucket) isSynthetic() bool {
	__antithesis_instrumentation__.Notify(98926)
	return atomic.LoadInt32(&b.synthetic) != 0
}

func (b *bucket) extendAndJoin(ctx context.Context, ts int64, synthetic bool) lockfreeToken {
	__antithesis_instrumentation__.Notify(98927)

	var t int64
	for {
		__antithesis_instrumentation__.Notify(98930)
		t = atomic.LoadInt64(&b.ts)
		if t != 0 && func() bool {
			__antithesis_instrumentation__.Notify(98932)
			return t <= ts == true
		}() == true {
			__antithesis_instrumentation__.Notify(98933)
			break
		} else {
			__antithesis_instrumentation__.Notify(98934)
		}
		__antithesis_instrumentation__.Notify(98931)
		if atomic.CompareAndSwapInt64(&b.ts, t, ts) {
			__antithesis_instrumentation__.Notify(98935)
			break
		} else {
			__antithesis_instrumentation__.Notify(98936)
		}
	}
	__antithesis_instrumentation__.Notify(98928)

	if t == 0 && func() bool {
		__antithesis_instrumentation__.Notify(98937)
		return synthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(98938)
		atomic.StoreInt32(&b.synthetic, 1)
	} else {
		__antithesis_instrumentation__.Notify(98939)
	}
	__antithesis_instrumentation__.Notify(98929)
	atomic.AddInt32(&b.refcnt, 1)
	return lockfreeToken{b: b}
}

type lockfreeToken struct {
	b *bucket
}

var _ RemovalToken = lockfreeToken{}

func (l lockfreeToken) RemovalTokenMarker() { __antithesis_instrumentation__.Notify(98940) }
