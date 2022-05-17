package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

var leaseHistoryMaxEntries = envutil.EnvOrDefaultInt("COCKROACH_LEASE_HISTORY", 5)

type leaseHistory struct {
	syncutil.Mutex
	index   int
	history []roachpb.Lease
}

func newLeaseHistory() *leaseHistory {
	__antithesis_instrumentation__.Notify(107943)
	lh := &leaseHistory{
		history: make([]roachpb.Lease, 0, leaseHistoryMaxEntries),
	}
	return lh
}

func (lh *leaseHistory) add(lease roachpb.Lease) {
	__antithesis_instrumentation__.Notify(107944)
	lh.Lock()
	defer lh.Unlock()

	if lh.index == len(lh.history) {
		__antithesis_instrumentation__.Notify(107946)
		lh.history = append(lh.history, lease)
	} else {
		__antithesis_instrumentation__.Notify(107947)
		lh.history[lh.index] = lease
	}
	__antithesis_instrumentation__.Notify(107945)
	lh.index++
	if lh.index >= leaseHistoryMaxEntries {
		__antithesis_instrumentation__.Notify(107948)
		lh.index = 0
	} else {
		__antithesis_instrumentation__.Notify(107949)
	}
}

func (lh *leaseHistory) get() []roachpb.Lease {
	__antithesis_instrumentation__.Notify(107950)
	lh.Lock()
	defer lh.Unlock()
	if len(lh.history) == 0 {
		__antithesis_instrumentation__.Notify(107953)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(107954)
	}
	__antithesis_instrumentation__.Notify(107951)
	if len(lh.history) < leaseHistoryMaxEntries || func() bool {
		__antithesis_instrumentation__.Notify(107955)
		return lh.index == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(107956)
		result := make([]roachpb.Lease, len(lh.history))
		copy(result, lh.history)
		return result
	} else {
		__antithesis_instrumentation__.Notify(107957)
	}
	__antithesis_instrumentation__.Notify(107952)
	first := lh.history[lh.index:]
	second := lh.history[:lh.index]
	result := make([]roachpb.Lease, len(first)+len(second))
	copy(result, first)
	copy(result[len(first):], second)
	return result
}
