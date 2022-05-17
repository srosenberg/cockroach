package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"go.etcd.io/etcd/raft/v3"
)

type ReplicaMetrics struct {
	Leader      bool
	LeaseValid  bool
	Leaseholder bool
	LeaseType   roachpb.LeaseType
	LeaseStatus kvserverpb.LeaseStatus

	Quiescent bool

	Ticking bool

	RangeCounter    bool
	Unavailable     bool
	Underreplicated bool
	Overreplicated  bool
	RaftLogTooLarge bool
	BehindCount     int64

	LatchMetrics     concurrency.LatchMetrics
	LockTableMetrics concurrency.LockTableMetrics
}

func (r *Replica) Metrics(
	ctx context.Context, now hlc.ClockTimestamp, livenessMap liveness.IsLiveMap, clusterNodes int,
) ReplicaMetrics {
	__antithesis_instrumentation__.Notify(117748)
	r.mu.RLock()
	raftStatus := r.raftStatusRLocked()
	leaseStatus := r.leaseStatusAtRLocked(ctx, now)
	quiescent := r.mu.quiescent
	desc := r.mu.state.Desc
	conf := r.mu.conf
	raftLogSize := r.mu.raftLogSize
	raftLogSizeTrusted := r.mu.raftLogSizeTrusted
	r.mu.RUnlock()

	r.store.unquiescedReplicas.Lock()
	_, ticking := r.store.unquiescedReplicas.m[r.RangeID]
	r.store.unquiescedReplicas.Unlock()

	latchMetrics := r.concMgr.LatchMetrics()
	lockTableMetrics := r.concMgr.LockTableMetrics()

	return calcReplicaMetrics(
		ctx,
		now.ToTimestamp(),
		&r.store.cfg.RaftConfig,
		conf,
		livenessMap,
		clusterNodes,
		desc,
		raftStatus,
		leaseStatus,
		r.store.StoreID(),
		quiescent,
		ticking,
		latchMetrics,
		lockTableMetrics,
		raftLogSize,
		raftLogSizeTrusted,
	)
}

func calcReplicaMetrics(
	_ context.Context,
	_ hlc.Timestamp,
	raftCfg *base.RaftConfig,
	conf roachpb.SpanConfig,
	livenessMap liveness.IsLiveMap,
	clusterNodes int,
	desc *roachpb.RangeDescriptor,
	raftStatus *raft.Status,
	leaseStatus kvserverpb.LeaseStatus,
	storeID roachpb.StoreID,
	quiescent bool,
	ticking bool,
	latchMetrics concurrency.LatchMetrics,
	lockTableMetrics concurrency.LockTableMetrics,
	raftLogSize int64,
	raftLogSizeTrusted bool,
) ReplicaMetrics {
	__antithesis_instrumentation__.Notify(117749)
	var m ReplicaMetrics

	var leaseOwner bool
	m.LeaseStatus = leaseStatus
	if leaseStatus.IsValid() {
		__antithesis_instrumentation__.Notify(117752)
		m.LeaseValid = true
		leaseOwner = leaseStatus.Lease.OwnedBy(storeID)
		m.LeaseType = leaseStatus.Lease.Type()
	} else {
		__antithesis_instrumentation__.Notify(117753)
	}
	__antithesis_instrumentation__.Notify(117750)
	m.Leaseholder = m.LeaseValid && func() bool {
		__antithesis_instrumentation__.Notify(117754)
		return leaseOwner == true
	}() == true
	m.Leader = isRaftLeader(raftStatus)
	m.Quiescent = quiescent
	m.Ticking = ticking

	m.RangeCounter, m.Unavailable, m.Underreplicated, m.Overreplicated = calcRangeCounter(
		storeID, desc, leaseStatus, livenessMap, conf.GetNumVoters(), conf.NumReplicas, clusterNodes)

	const raftLogTooLargeMultiple = 4
	m.RaftLogTooLarge = raftLogSize > (raftLogTooLargeMultiple*raftCfg.RaftLogTruncationThreshold) && func() bool {
		__antithesis_instrumentation__.Notify(117755)
		return raftLogSizeTrusted == true
	}() == true

	if m.Leader {
		__antithesis_instrumentation__.Notify(117756)
		m.BehindCount = calcBehindCount(raftStatus, desc, livenessMap)
	} else {
		__antithesis_instrumentation__.Notify(117757)
	}
	__antithesis_instrumentation__.Notify(117751)

	m.LatchMetrics = latchMetrics
	m.LockTableMetrics = lockTableMetrics

	return m
}

func calcRangeCounter(
	storeID roachpb.StoreID,
	desc *roachpb.RangeDescriptor,
	leaseStatus kvserverpb.LeaseStatus,
	livenessMap liveness.IsLiveMap,
	numVoters, numReplicas int32,
	clusterNodes int,
) (rangeCounter, unavailable, underreplicated, overreplicated bool) {
	__antithesis_instrumentation__.Notify(117758)

	if livenessMap[leaseStatus.Lease.Replica.NodeID].IsLive {
		__antithesis_instrumentation__.Notify(117761)
		rangeCounter = leaseStatus.OwnedBy(storeID)
	} else {
		__antithesis_instrumentation__.Notify(117762)

		for _, rd := range desc.Replicas().Descriptors() {
			__antithesis_instrumentation__.Notify(117763)
			if livenessMap[rd.NodeID].IsLive {
				__antithesis_instrumentation__.Notify(117764)
				rangeCounter = rd.StoreID == storeID
				break
			} else {
				__antithesis_instrumentation__.Notify(117765)
			}
		}
	}
	__antithesis_instrumentation__.Notify(117759)

	if rangeCounter {
		__antithesis_instrumentation__.Notify(117766)
		neededVoters := GetNeededVoters(numVoters, clusterNodes)
		neededNonVoters := GetNeededNonVoters(int(numVoters), int(numReplicas-numVoters), clusterNodes)
		status := desc.Replicas().ReplicationStatus(func(rDesc roachpb.ReplicaDescriptor) bool {
			__antithesis_instrumentation__.Notify(117768)
			return livenessMap[rDesc.NodeID].IsLive
		},

			0)
		__antithesis_instrumentation__.Notify(117767)
		unavailable = !status.Available
		liveVoters := calcLiveVoterReplicas(desc, livenessMap)
		liveNonVoters := calcLiveNonVoterReplicas(desc, livenessMap)
		if neededVoters > liveVoters || func() bool {
			__antithesis_instrumentation__.Notify(117769)
			return neededNonVoters > liveNonVoters == true
		}() == true {
			__antithesis_instrumentation__.Notify(117770)
			underreplicated = true
		} else {
			__antithesis_instrumentation__.Notify(117771)
			if neededVoters < liveVoters || func() bool {
				__antithesis_instrumentation__.Notify(117772)
				return neededNonVoters < liveNonVoters == true
			}() == true {
				__antithesis_instrumentation__.Notify(117773)
				overreplicated = true
			} else {
				__antithesis_instrumentation__.Notify(117774)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(117775)
	}
	__antithesis_instrumentation__.Notify(117760)
	return
}

func calcLiveVoterReplicas(desc *roachpb.RangeDescriptor, livenessMap liveness.IsLiveMap) int {
	__antithesis_instrumentation__.Notify(117776)
	return calcLiveReplicas(desc.Replicas().VoterDescriptors(), livenessMap)
}

func calcLiveNonVoterReplicas(desc *roachpb.RangeDescriptor, livenessMap liveness.IsLiveMap) int {
	__antithesis_instrumentation__.Notify(117777)
	return calcLiveReplicas(desc.Replicas().NonVoterDescriptors(), livenessMap)
}

func calcLiveReplicas(repls []roachpb.ReplicaDescriptor, livenessMap liveness.IsLiveMap) int {
	__antithesis_instrumentation__.Notify(117778)
	var live int
	for _, rd := range repls {
		__antithesis_instrumentation__.Notify(117780)
		if livenessMap[rd.NodeID].IsLive {
			__antithesis_instrumentation__.Notify(117781)
			live++
		} else {
			__antithesis_instrumentation__.Notify(117782)
		}
	}
	__antithesis_instrumentation__.Notify(117779)
	return live
}

func calcBehindCount(
	raftStatus *raft.Status, desc *roachpb.RangeDescriptor, livenessMap liveness.IsLiveMap,
) int64 {
	__antithesis_instrumentation__.Notify(117783)
	var behindCount int64
	for _, rd := range desc.Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(117785)
		if progress, ok := raftStatus.Progress[uint64(rd.ReplicaID)]; ok {
			__antithesis_instrumentation__.Notify(117786)
			if progress.Match > 0 && func() bool {
				__antithesis_instrumentation__.Notify(117787)
				return progress.Match < raftStatus.Commit == true
			}() == true {
				__antithesis_instrumentation__.Notify(117788)
				behindCount += int64(raftStatus.Commit) - int64(progress.Match)
			} else {
				__antithesis_instrumentation__.Notify(117789)
			}
		} else {
			__antithesis_instrumentation__.Notify(117790)
		}
	}
	__antithesis_instrumentation__.Notify(117784)

	return behindCount
}

func (r *Replica) QueriesPerSecond() (float64, time.Duration) {
	__antithesis_instrumentation__.Notify(117791)
	return r.leaseholderStats.avgQPS()
}

func (r *Replica) WritesPerSecond() float64 {
	__antithesis_instrumentation__.Notify(117792)
	wps, _ := r.writeStats.avgQPS()
	return wps
}

func (r *Replica) needsSplitBySizeRLocked() bool {
	__antithesis_instrumentation__.Notify(117793)
	exceeded, _ := r.exceedsMultipleOfSplitSizeRLocked(1)
	return exceeded
}

func (r *Replica) needsMergeBySizeRLocked() bool {
	__antithesis_instrumentation__.Notify(117794)
	return r.mu.state.Stats.Total() < r.mu.conf.RangeMinBytes
}

func (r *Replica) needsRaftLogTruncationLocked() bool {
	__antithesis_instrumentation__.Notify(117795)

	checkRaftLog := r.mu.raftLogSize-r.mu.raftLogLastCheckSize >= RaftLogQueueStaleSize
	if checkRaftLog {
		__antithesis_instrumentation__.Notify(117797)
		r.mu.raftLogLastCheckSize = r.mu.raftLogSize
	} else {
		__antithesis_instrumentation__.Notify(117798)
	}
	__antithesis_instrumentation__.Notify(117796)
	return checkRaftLog
}

func (r *Replica) exceedsMultipleOfSplitSizeRLocked(mult float64) (exceeded bool, bytesOver int64) {
	__antithesis_instrumentation__.Notify(117799)
	maxBytes := r.mu.conf.RangeMaxBytes
	if r.mu.largestPreviousMaxRangeSizeBytes > maxBytes {
		__antithesis_instrumentation__.Notify(117802)
		maxBytes = r.mu.largestPreviousMaxRangeSizeBytes
	} else {
		__antithesis_instrumentation__.Notify(117803)
	}
	__antithesis_instrumentation__.Notify(117800)
	size := r.mu.state.Stats.Total()
	maxSize := int64(float64(maxBytes)*mult) + 1
	if maxBytes <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(117804)
		return size <= maxSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(117805)
		return false, 0
	} else {
		__antithesis_instrumentation__.Notify(117806)
	}
	__antithesis_instrumentation__.Notify(117801)
	return true, size - maxSize
}
