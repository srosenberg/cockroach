package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

const (
	numTopReplicasToTrack = 128
)

type replicaWithStats struct {
	repl *Replica
	qps  float64
}

type replicaRankings struct {
	mu struct {
		syncutil.Mutex
		qpsAccumulator *rrAccumulator
		byQPS          []replicaWithStats
	}
}

func newReplicaRankings() *replicaRankings {
	__antithesis_instrumentation__.Notify(120098)
	return &replicaRankings{}
}

func (rr *replicaRankings) newAccumulator() *rrAccumulator {
	__antithesis_instrumentation__.Notify(120099)
	res := &rrAccumulator{}
	res.qps.val = func(r replicaWithStats) float64 { __antithesis_instrumentation__.Notify(120101); return r.qps }
	__antithesis_instrumentation__.Notify(120100)
	return res
}

func (rr *replicaRankings) update(acc *rrAccumulator) {
	__antithesis_instrumentation__.Notify(120102)
	rr.mu.Lock()
	rr.mu.qpsAccumulator = acc
	rr.mu.Unlock()
}

func (rr *replicaRankings) topQPS() []replicaWithStats {
	__antithesis_instrumentation__.Notify(120103)
	rr.mu.Lock()
	defer rr.mu.Unlock()

	if rr.mu.qpsAccumulator != nil && func() bool {
		__antithesis_instrumentation__.Notify(120105)
		return rr.mu.qpsAccumulator.qps.Len() > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(120106)
		rr.mu.byQPS = consumeAccumulator(&rr.mu.qpsAccumulator.qps)
	} else {
		__antithesis_instrumentation__.Notify(120107)
	}
	__antithesis_instrumentation__.Notify(120104)
	return rr.mu.byQPS
}

type rrAccumulator struct {
	qps rrPriorityQueue
}

func (a *rrAccumulator) addReplica(repl replicaWithStats) {
	__antithesis_instrumentation__.Notify(120108)

	if a.qps.Len() < numTopReplicasToTrack {
		__antithesis_instrumentation__.Notify(120110)
		heap.Push(&a.qps, repl)
		return
	} else {
		__antithesis_instrumentation__.Notify(120111)
	}
	__antithesis_instrumentation__.Notify(120109)

	if repl.qps > a.qps.entries[0].qps {
		__antithesis_instrumentation__.Notify(120112)
		heap.Pop(&a.qps)
		heap.Push(&a.qps, repl)
	} else {
		__antithesis_instrumentation__.Notify(120113)
	}
}

func consumeAccumulator(pq *rrPriorityQueue) []replicaWithStats {
	__antithesis_instrumentation__.Notify(120114)
	length := pq.Len()
	sorted := make([]replicaWithStats, length)
	for i := 1; i <= length; i++ {
		__antithesis_instrumentation__.Notify(120116)
		sorted[length-i] = heap.Pop(pq).(replicaWithStats)
	}
	__antithesis_instrumentation__.Notify(120115)
	return sorted
}

type rrPriorityQueue struct {
	entries []replicaWithStats
	val     func(replicaWithStats) float64
}

func (pq rrPriorityQueue) Len() int {
	__antithesis_instrumentation__.Notify(120117)
	return len(pq.entries)
}

func (pq rrPriorityQueue) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(120118)
	return pq.val(pq.entries[i]) < pq.val(pq.entries[j])
}

func (pq rrPriorityQueue) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(120119)
	pq.entries[i], pq.entries[j] = pq.entries[j], pq.entries[i]
}

func (pq *rrPriorityQueue) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(120120)
	item := x.(replicaWithStats)
	pq.entries = append(pq.entries, item)
}

func (pq *rrPriorityQueue) Pop() interface{} {
	__antithesis_instrumentation__.Notify(120121)
	old := pq.entries
	n := len(old)
	item := old[n-1]
	pq.entries = old[0 : n-1]
	return item
}
