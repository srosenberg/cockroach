package tracker

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type heapTracker struct {
	mu struct {
		syncutil.Mutex
		rs tsHeap
	}
}

var _ Tracker = &heapTracker{}

func newHeapTracker() Tracker {
	__antithesis_instrumentation__.Notify(98884)
	return &heapTracker{}
}

type item struct {
	ts hlc.Timestamp

	index int
}

type tsHeap []*item

var _ heap.Interface = &tsHeap{}

func (h tsHeap) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(98885)
	return h[i].ts.Less(h[j].ts)
}

func (h tsHeap) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(98886)
	tmp := h[i]
	h[i] = h[j]
	h[j] = tmp
	h[i].index = i
	h[j].index = j
}

func (h *tsHeap) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(98887)
	n := len(*h)
	item := x.(*item)
	item.index = n
	*h = append(*h, item)
}

func (h *tsHeap) Pop() interface{} {
	__antithesis_instrumentation__.Notify(98888)
	it := (*h)[len(*h)-1]

	it.index = -1
	*h = (*h)[0 : len(*h)-1]
	return it
}

func (h *tsHeap) Len() int {
	__antithesis_instrumentation__.Notify(98889)
	return len(*h)
}

type heapToken struct {
	*item
}

func (heapToken) RemovalTokenMarker() { __antithesis_instrumentation__.Notify(98890) }

var _ RemovalToken = heapToken{}

func (h *heapTracker) Track(ctx context.Context, ts hlc.Timestamp) RemovalToken {
	__antithesis_instrumentation__.Notify(98891)
	h.mu.Lock()
	defer h.mu.Unlock()
	i := &item{ts: ts}
	heap.Push(&h.mu.rs, i)
	return heapToken{i}
}

func (h *heapTracker) Untrack(ctx context.Context, tok RemovalToken) {
	__antithesis_instrumentation__.Notify(98892)
	idx := tok.(heapToken).index
	if idx == -1 {
		__antithesis_instrumentation__.Notify(98894)
		log.Fatalf(ctx, "attempting to untrack already-untracked item")
	} else {
		__antithesis_instrumentation__.Notify(98895)
	}
	__antithesis_instrumentation__.Notify(98893)
	h.mu.Lock()
	defer h.mu.Unlock()
	heap.Remove(&h.mu.rs, idx)
}

func (h *heapTracker) LowerBound(ctx context.Context) hlc.Timestamp {
	__antithesis_instrumentation__.Notify(98896)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.mu.rs.Len() == 0 {
		__antithesis_instrumentation__.Notify(98898)
		return hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(98899)
	}
	__antithesis_instrumentation__.Notify(98897)
	return h.mu.rs[0].ts
}

func (h *heapTracker) Count() int {
	__antithesis_instrumentation__.Notify(98900)
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.mu.rs.Len()
}
