package concurrency

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/spanlatch"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
)

type LatchMetrics = spanlatch.Metrics

type TopKLockMetrics = [3]LockMetrics

type LockTableMetrics struct {
	Locks int64

	LocksHeld int64

	TotalLockHoldDurationNanos int64

	LocksWithReservation int64

	LocksWithWaitQueues int64

	Waiters int64

	WaitingReaders int64

	WaitingWriters int64

	TotalWaitDurationNanos int64

	TopKLocksByWaiters TopKLockMetrics

	TopKLocksByHoldDuration TopKLockMetrics

	TopKLocksByWaitDuration TopKLockMetrics
}

type LockMetrics struct {
	Key roachpb.Key

	Held bool

	HoldDurationNanos int64

	Waiters int64

	WaitingReaders int64

	WaitingWriters int64

	WaitDurationNanos int64

	MaxWaitDurationNanos int64
}

func (m *LockTableMetrics) addLockMetrics(lm LockMetrics) {
	__antithesis_instrumentation__.Notify(100801)
	m.Locks++
	if lm.Held {
		__antithesis_instrumentation__.Notify(100803)
		m.LocksHeld++
		m.TotalLockHoldDurationNanos += lm.HoldDurationNanos
		m.addToTopKLocksByHoldDuration(lm)
	} else {
		__antithesis_instrumentation__.Notify(100804)
		m.LocksWithReservation++
	}
	__antithesis_instrumentation__.Notify(100802)
	if lm.Waiters > 0 {
		__antithesis_instrumentation__.Notify(100805)
		m.LocksWithWaitQueues++
		m.Waiters += lm.Waiters
		m.WaitingReaders += lm.WaitingReaders
		m.WaitingWriters += lm.WaitingWriters
		m.TotalWaitDurationNanos += lm.WaitDurationNanos
		m.addToTopKLocksByWaiters(lm)
		m.addToTopKLocksByWaitDuration(lm)
	} else {
		__antithesis_instrumentation__.Notify(100806)
	}
}

func (m *LockTableMetrics) addToTopKLocksByWaiters(lm LockMetrics) {
	__antithesis_instrumentation__.Notify(100807)
	addToTopK(m.TopKLocksByWaiters[:], lm, func(cmp LockMetrics) int64 { __antithesis_instrumentation__.Notify(100808); return cmp.Waiters })
}

func (m *LockTableMetrics) addToTopKLocksByHoldDuration(lm LockMetrics) {
	__antithesis_instrumentation__.Notify(100809)
	addToTopK(m.TopKLocksByHoldDuration[:], lm, func(cmp LockMetrics) int64 {
		__antithesis_instrumentation__.Notify(100810)
		return cmp.HoldDurationNanos
	})
}

func (m *LockTableMetrics) addToTopKLocksByWaitDuration(lm LockMetrics) {
	__antithesis_instrumentation__.Notify(100811)
	addToTopK(m.TopKLocksByWaitDuration[:], lm, func(cmp LockMetrics) int64 {
		__antithesis_instrumentation__.Notify(100812)
		return cmp.MaxWaitDurationNanos
	})
}

func addToTopK(topK []LockMetrics, lm LockMetrics, cmp func(LockMetrics) int64) {
	__antithesis_instrumentation__.Notify(100813)
	cpy := false
	for i, cur := range topK {
		__antithesis_instrumentation__.Notify(100814)
		if cur.Key == nil {
			__antithesis_instrumentation__.Notify(100816)
			topK[i] = lm
			break
		} else {
			__antithesis_instrumentation__.Notify(100817)
		}
		__antithesis_instrumentation__.Notify(100815)
		if cpy || func() bool {
			__antithesis_instrumentation__.Notify(100818)
			return cmp(lm) > cmp(cur) == true
		}() == true {
			__antithesis_instrumentation__.Notify(100819)
			topK[i] = lm
			lm = cur
			cpy = true
		} else {
			__antithesis_instrumentation__.Notify(100820)
		}
	}
}
