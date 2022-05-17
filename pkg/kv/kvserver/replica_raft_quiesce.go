package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func (r *Replica) quiesceLocked(ctx context.Context, lagging laggingReplicaSet) {
	__antithesis_instrumentation__.Notify(118973)
	if !r.mu.quiescent {
		__antithesis_instrumentation__.Notify(118974)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(118976)
			log.Infof(ctx, "quiescing %d", r.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(118977)
		}
		__antithesis_instrumentation__.Notify(118975)
		r.mu.quiescent = true
		r.mu.laggingFollowersOnQuiesce = lagging
		r.store.unquiescedReplicas.Lock()
		delete(r.store.unquiescedReplicas.m, r.RangeID)
		r.store.unquiescedReplicas.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(118978)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(118979)
			log.Infof(ctx, "already quiesced")
		} else {
			__antithesis_instrumentation__.Notify(118980)
		}
	}
}

func (r *Replica) maybeUnquiesce() bool {
	__antithesis_instrumentation__.Notify(118981)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.maybeUnquiesceLocked()
}

func (r *Replica) maybeUnquiesceLocked() bool {
	__antithesis_instrumentation__.Notify(118982)
	return r.maybeUnquiesceWithOptionsLocked(true)
}

func (r *Replica) maybeUnquiesceWithOptionsLocked(campaignOnWake bool) bool {
	__antithesis_instrumentation__.Notify(118983)
	if !r.canUnquiesceRLocked() {
		__antithesis_instrumentation__.Notify(118987)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118988)
	}
	__antithesis_instrumentation__.Notify(118984)
	ctx := r.AnnotateCtx(context.TODO())
	if log.V(3) {
		__antithesis_instrumentation__.Notify(118989)
		log.Infof(ctx, "unquiescing %d", r.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(118990)
	}
	__antithesis_instrumentation__.Notify(118985)
	r.mu.quiescent = false
	r.mu.laggingFollowersOnQuiesce = nil
	r.store.unquiescedReplicas.Lock()
	r.store.unquiescedReplicas.m[r.RangeID] = struct{}{}
	r.store.unquiescedReplicas.Unlock()
	if campaignOnWake {
		__antithesis_instrumentation__.Notify(118991)
		r.maybeCampaignOnWakeLocked(ctx)
	} else {
		__antithesis_instrumentation__.Notify(118992)
	}
	__antithesis_instrumentation__.Notify(118986)

	r.mu.lastUpdateTimes.updateOnUnquiesce(
		r.mu.state.Desc.Replicas().Descriptors(), r.raftStatusRLocked().Progress, timeutil.Now(),
	)
	return true
}

func (r *Replica) maybeUnquiesceAndWakeLeaderLocked() bool {
	__antithesis_instrumentation__.Notify(118993)
	if !r.canUnquiesceRLocked() {
		__antithesis_instrumentation__.Notify(118996)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118997)
	}
	__antithesis_instrumentation__.Notify(118994)
	ctx := r.AnnotateCtx(context.TODO())
	if log.V(3) {
		__antithesis_instrumentation__.Notify(118998)
		log.Infof(ctx, "unquiescing %d: waking leader", r.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(118999)
	}
	__antithesis_instrumentation__.Notify(118995)
	r.mu.quiescent = false
	r.mu.laggingFollowersOnQuiesce = nil
	r.store.unquiescedReplicas.Lock()
	r.store.unquiescedReplicas.m[r.RangeID] = struct{}{}
	r.store.unquiescedReplicas.Unlock()
	r.maybeCampaignOnWakeLocked(ctx)

	data := kvserverbase.EncodeRaftCommand(kvserverbase.RaftVersionStandard, makeIDKey(), nil)
	_ = r.mu.internalRaftGroup.Propose(data)
	return true
}

func (r *Replica) canUnquiesceRLocked() bool {
	__antithesis_instrumentation__.Notify(119000)
	return r.mu.quiescent && func() bool {
		__antithesis_instrumentation__.Notify(119001)
		return r.isInitializedRLocked() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(119002)
		return r.mu.internalRaftGroup != nil == true
	}() == true
}

func (r *Replica) maybeQuiesceRaftMuLockedReplicaMuLocked(
	ctx context.Context, now hlc.ClockTimestamp, livenessMap liveness.IsLiveMap,
) bool {
	__antithesis_instrumentation__.Notify(119003)
	status, lagging, ok := shouldReplicaQuiesce(ctx, r, now, livenessMap)
	if !ok {
		__antithesis_instrumentation__.Notify(119005)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119006)
	}
	__antithesis_instrumentation__.Notify(119004)
	return r.quiesceAndNotifyRaftMuLockedReplicaMuLocked(ctx, status, lagging)
}

type quiescer interface {
	descRLocked() *roachpb.RangeDescriptor
	raftStatusRLocked() *raft.Status
	raftBasicStatusRLocked() raft.BasicStatus
	raftLastIndexLocked() (uint64, error)
	hasRaftReadyRLocked() bool
	hasPendingProposalsRLocked() bool
	hasPendingProposalQuotaRLocked() bool
	ownsValidLeaseRLocked(ctx context.Context, now hlc.ClockTimestamp) bool
	mergeInProgressRLocked() bool
	isDestroyedRLocked() (DestroyReason, error)
}

type laggingReplicaSet []livenesspb.Liveness

func (s laggingReplicaSet) MemberStale(l livenesspb.Liveness) bool {
	__antithesis_instrumentation__.Notify(119007)
	for _, laggingL := range s {
		__antithesis_instrumentation__.Notify(119009)
		if laggingL.NodeID == l.NodeID {
			__antithesis_instrumentation__.Notify(119010)
			return laggingL.Compare(l) < 0
		} else {
			__antithesis_instrumentation__.Notify(119011)
		}
	}
	__antithesis_instrumentation__.Notify(119008)
	return false
}

func (s laggingReplicaSet) AnyMemberStale(livenessMap liveness.IsLiveMap) bool {
	__antithesis_instrumentation__.Notify(119012)
	for _, laggingL := range s {
		__antithesis_instrumentation__.Notify(119014)
		if l, ok := livenessMap[laggingL.NodeID]; ok {
			__antithesis_instrumentation__.Notify(119015)
			if laggingL.Compare(l.Liveness) < 0 {
				__antithesis_instrumentation__.Notify(119016)
				return true
			} else {
				__antithesis_instrumentation__.Notify(119017)
			}
		} else {
			__antithesis_instrumentation__.Notify(119018)
		}
	}
	__antithesis_instrumentation__.Notify(119013)
	return false
}

func (s laggingReplicaSet) Len() int { __antithesis_instrumentation__.Notify(119019); return len(s) }
func (s laggingReplicaSet) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(119020)
	s[i], s[j] = s[j], s[i]
}
func (s laggingReplicaSet) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(119021)
	return s[i].NodeID < s[j].NodeID
}

func shouldReplicaQuiesce(
	ctx context.Context, q quiescer, now hlc.ClockTimestamp, livenessMap liveness.IsLiveMap,
) (*raft.Status, laggingReplicaSet, bool) {
	__antithesis_instrumentation__.Notify(119022)
	if testingDisableQuiescence {
		__antithesis_instrumentation__.Notify(119038)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119039)
	}
	__antithesis_instrumentation__.Notify(119023)
	if q.hasPendingProposalsRLocked() {
		__antithesis_instrumentation__.Notify(119040)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119042)
			log.Infof(ctx, "not quiescing: proposals pending")
		} else {
			__antithesis_instrumentation__.Notify(119043)
		}
		__antithesis_instrumentation__.Notify(119041)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119044)
	}
	__antithesis_instrumentation__.Notify(119024)

	if q.hasPendingProposalQuotaRLocked() {
		__antithesis_instrumentation__.Notify(119045)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119047)
			log.Infof(ctx, "not quiescing: replication quota outstanding")
		} else {
			__antithesis_instrumentation__.Notify(119048)
		}
		__antithesis_instrumentation__.Notify(119046)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119049)
	}
	__antithesis_instrumentation__.Notify(119025)
	if q.mergeInProgressRLocked() {
		__antithesis_instrumentation__.Notify(119050)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119052)
			log.Infof(ctx, "not quiescing: merge in progress")
		} else {
			__antithesis_instrumentation__.Notify(119053)
		}
		__antithesis_instrumentation__.Notify(119051)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119054)
	}
	__antithesis_instrumentation__.Notify(119026)
	if _, err := q.isDestroyedRLocked(); err != nil {
		__antithesis_instrumentation__.Notify(119055)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119057)
			log.Infof(ctx, "not quiescing: replica destroyed")
		} else {
			__antithesis_instrumentation__.Notify(119058)
		}
		__antithesis_instrumentation__.Notify(119056)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119059)
	}
	__antithesis_instrumentation__.Notify(119027)
	status := q.raftStatusRLocked()
	if status == nil {
		__antithesis_instrumentation__.Notify(119060)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119062)
			log.Infof(ctx, "not quiescing: dormant Raft group")
		} else {
			__antithesis_instrumentation__.Notify(119063)
		}
		__antithesis_instrumentation__.Notify(119061)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119064)
	}
	__antithesis_instrumentation__.Notify(119028)
	if status.SoftState.RaftState != raft.StateLeader {
		__antithesis_instrumentation__.Notify(119065)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119067)
			log.Infof(ctx, "not quiescing: not leader")
		} else {
			__antithesis_instrumentation__.Notify(119068)
		}
		__antithesis_instrumentation__.Notify(119066)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119069)
	}
	__antithesis_instrumentation__.Notify(119029)
	if status.LeadTransferee != 0 {
		__antithesis_instrumentation__.Notify(119070)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119072)
			log.Infof(ctx, "not quiescing: leader transfer to %d in progress", status.LeadTransferee)
		} else {
			__antithesis_instrumentation__.Notify(119073)
		}
		__antithesis_instrumentation__.Notify(119071)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119074)
	}
	__antithesis_instrumentation__.Notify(119030)

	if !q.ownsValidLeaseRLocked(ctx, now) {
		__antithesis_instrumentation__.Notify(119075)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119077)
			log.Infof(ctx, "not quiescing: not leaseholder")
		} else {
			__antithesis_instrumentation__.Notify(119078)
		}
		__antithesis_instrumentation__.Notify(119076)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119079)
	}
	__antithesis_instrumentation__.Notify(119031)

	if status.Applied != status.Commit {
		__antithesis_instrumentation__.Notify(119080)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119082)
			log.Infof(ctx, "not quiescing: applied (%d) != commit (%d)",
				status.Applied, status.Commit)
		} else {
			__antithesis_instrumentation__.Notify(119083)
		}
		__antithesis_instrumentation__.Notify(119081)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119084)
	}
	__antithesis_instrumentation__.Notify(119032)
	lastIndex, err := q.raftLastIndexLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(119085)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119087)
			log.Infof(ctx, "not quiescing: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(119088)
		}
		__antithesis_instrumentation__.Notify(119086)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119089)
	}
	__antithesis_instrumentation__.Notify(119033)
	if status.Commit != lastIndex {
		__antithesis_instrumentation__.Notify(119090)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119092)
			log.Infof(ctx, "not quiescing: commit (%d) != lastIndex (%d)",
				status.Commit, lastIndex)
		} else {
			__antithesis_instrumentation__.Notify(119093)
		}
		__antithesis_instrumentation__.Notify(119091)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119094)
	}
	__antithesis_instrumentation__.Notify(119034)

	var foundSelf bool
	var lagging laggingReplicaSet
	for _, rep := range q.descRLocked().Replicas().Descriptors() {
		__antithesis_instrumentation__.Notify(119095)
		if uint64(rep.ReplicaID) == status.ID {
			__antithesis_instrumentation__.Notify(119097)
			foundSelf = true
		} else {
			__antithesis_instrumentation__.Notify(119098)
		}
		__antithesis_instrumentation__.Notify(119096)
		if progress, ok := status.Progress[uint64(rep.ReplicaID)]; !ok {
			__antithesis_instrumentation__.Notify(119099)
			if log.V(4) {
				__antithesis_instrumentation__.Notify(119101)
				log.Infof(ctx, "not quiescing: could not locate replica %d in progress: %+v",
					rep.ReplicaID, progress)
			} else {
				__antithesis_instrumentation__.Notify(119102)
			}
			__antithesis_instrumentation__.Notify(119100)
			return nil, nil, false
		} else {
			__antithesis_instrumentation__.Notify(119103)
			if progress.Match != status.Applied {
				__antithesis_instrumentation__.Notify(119104)

				if l, ok := livenessMap[rep.NodeID]; ok && func() bool {
					__antithesis_instrumentation__.Notify(119107)
					return !l.IsLive == true
				}() == true {
					__antithesis_instrumentation__.Notify(119108)
					if log.V(4) {
						__antithesis_instrumentation__.Notify(119110)
						log.Infof(ctx, "skipping node %d because not live. Progress=%+v",
							rep.NodeID, progress)
					} else {
						__antithesis_instrumentation__.Notify(119111)
					}
					__antithesis_instrumentation__.Notify(119109)
					lagging = append(lagging, l.Liveness)
					continue
				} else {
					__antithesis_instrumentation__.Notify(119112)
				}
				__antithesis_instrumentation__.Notify(119105)
				if log.V(4) {
					__antithesis_instrumentation__.Notify(119113)
					log.Infof(ctx, "not quiescing: replica %d match (%d) != applied (%d)",
						rep.ReplicaID, progress.Match, status.Applied)
				} else {
					__antithesis_instrumentation__.Notify(119114)
				}
				__antithesis_instrumentation__.Notify(119106)
				return nil, nil, false
			} else {
				__antithesis_instrumentation__.Notify(119115)
			}
		}
	}
	__antithesis_instrumentation__.Notify(119035)
	sort.Sort(lagging)
	if !foundSelf {
		__antithesis_instrumentation__.Notify(119116)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119118)
			log.Infof(ctx, "not quiescing: %d not found in progress: %+v",
				status.ID, status.Progress)
		} else {
			__antithesis_instrumentation__.Notify(119119)
		}
		__antithesis_instrumentation__.Notify(119117)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119120)
	}
	__antithesis_instrumentation__.Notify(119036)
	if q.hasRaftReadyRLocked() {
		__antithesis_instrumentation__.Notify(119121)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119123)
			log.Infof(ctx, "not quiescing: raft ready")
		} else {
			__antithesis_instrumentation__.Notify(119124)
		}
		__antithesis_instrumentation__.Notify(119122)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(119125)
	}
	__antithesis_instrumentation__.Notify(119037)
	return status, lagging, true
}

func (r *Replica) quiesceAndNotifyRaftMuLockedReplicaMuLocked(
	ctx context.Context, status *raft.Status, lagging laggingReplicaSet,
) bool {
	__antithesis_instrumentation__.Notify(119126)
	fromReplica, fromErr := r.getReplicaDescriptorByIDRLocked(r.replicaID, r.raftMu.lastToReplica)
	if fromErr != nil {
		__antithesis_instrumentation__.Notify(119129)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119131)
			log.Infof(ctx, "not quiescing: cannot find from replica (%d)", r.replicaID)
		} else {
			__antithesis_instrumentation__.Notify(119132)
		}
		__antithesis_instrumentation__.Notify(119130)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119133)
	}
	__antithesis_instrumentation__.Notify(119127)

	r.quiesceLocked(ctx, lagging)

	for id, prog := range status.Progress {
		__antithesis_instrumentation__.Notify(119134)
		if roachpb.ReplicaID(id) == r.replicaID {
			__antithesis_instrumentation__.Notify(119138)
			continue
		} else {
			__antithesis_instrumentation__.Notify(119139)
		}
		__antithesis_instrumentation__.Notify(119135)
		toReplica, toErr := r.getReplicaDescriptorByIDRLocked(
			roachpb.ReplicaID(id), r.raftMu.lastFromReplica)
		if toErr != nil {
			__antithesis_instrumentation__.Notify(119140)
			if log.V(4) {
				__antithesis_instrumentation__.Notify(119142)
				log.Infof(ctx, "failed to quiesce: cannot find to replica (%d)", id)
			} else {
				__antithesis_instrumentation__.Notify(119143)
			}
			__antithesis_instrumentation__.Notify(119141)
			r.maybeUnquiesceLocked()
			return false
		} else {
			__antithesis_instrumentation__.Notify(119144)
		}
		__antithesis_instrumentation__.Notify(119136)

		commit := status.Commit
		quiesce := true
		curLagging := lagging
		if prog.Match < status.Commit {
			__antithesis_instrumentation__.Notify(119145)
			commit = prog.Match
			quiesce = false
			curLagging = nil
		} else {
			__antithesis_instrumentation__.Notify(119146)
		}
		__antithesis_instrumentation__.Notify(119137)
		msg := raftpb.Message{
			From:   uint64(r.replicaID),
			To:     id,
			Type:   raftpb.MsgHeartbeat,
			Term:   status.Term,
			Commit: commit,
		}

		if !r.maybeCoalesceHeartbeat(ctx, msg, toReplica, fromReplica, quiesce, curLagging) {
			__antithesis_instrumentation__.Notify(119147)
			log.Fatalf(ctx, "failed to coalesce known heartbeat: %v", msg)
		} else {
			__antithesis_instrumentation__.Notify(119148)
		}
	}
	__antithesis_instrumentation__.Notify(119128)
	return true
}

func shouldFollowerQuiesceOnNotify(
	ctx context.Context,
	q quiescer,
	msg raftpb.Message,
	lagging laggingReplicaSet,
	livenessMap liveness.IsLiveMap,
) bool {
	__antithesis_instrumentation__.Notify(119149)

	status := q.raftBasicStatusRLocked()
	if status.Term != msg.Term {
		__antithesis_instrumentation__.Notify(119155)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119157)
			log.Infof(ctx, "not quiescing: local raft term is %d, incoming term is %d", status.Term, msg.Term)
		} else {
			__antithesis_instrumentation__.Notify(119158)
		}
		__antithesis_instrumentation__.Notify(119156)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119159)
	}
	__antithesis_instrumentation__.Notify(119150)
	if status.Commit != msg.Commit {
		__antithesis_instrumentation__.Notify(119160)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119162)
			log.Infof(ctx, "not quiescing: local raft commit index is %d, incoming commit index is %d", status.Commit, msg.Commit)
		} else {
			__antithesis_instrumentation__.Notify(119163)
		}
		__antithesis_instrumentation__.Notify(119161)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119164)
	}
	__antithesis_instrumentation__.Notify(119151)
	if status.Lead != msg.From {
		__antithesis_instrumentation__.Notify(119165)
		if log.V(4) {
			__antithesis_instrumentation__.Notify(119167)
			log.Infof(ctx, "not quiescing: local raft leader is %d, incoming message from %d", status.Lead, msg.From)
		} else {
			__antithesis_instrumentation__.Notify(119168)
		}
		__antithesis_instrumentation__.Notify(119166)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119169)
	}
	__antithesis_instrumentation__.Notify(119152)

	if q.hasPendingProposalsRLocked() {
		__antithesis_instrumentation__.Notify(119170)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(119172)
			log.Infof(ctx, "not quiescing: pending commands")
		} else {
			__antithesis_instrumentation__.Notify(119173)
		}
		__antithesis_instrumentation__.Notify(119171)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119174)
	}
	__antithesis_instrumentation__.Notify(119153)

	if lagging.AnyMemberStale(livenessMap) {
		__antithesis_instrumentation__.Notify(119175)
		if log.V(3) {
			__antithesis_instrumentation__.Notify(119177)
			log.Infof(ctx, "not quiescing: liveness info about lagging replica stale")
		} else {
			__antithesis_instrumentation__.Notify(119178)
		}
		__antithesis_instrumentation__.Notify(119176)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119179)
	}
	__antithesis_instrumentation__.Notify(119154)
	return true
}

func (r *Replica) maybeQuiesceOnNotify(
	ctx context.Context, msg raftpb.Message, lagging laggingReplicaSet,
) bool {
	__antithesis_instrumentation__.Notify(119180)
	r.mu.Lock()
	defer r.mu.Unlock()

	livenessMap, _ := r.store.livenessMap.Load().(liveness.IsLiveMap)
	if !shouldFollowerQuiesceOnNotify(ctx, r, msg, lagging, livenessMap) {
		__antithesis_instrumentation__.Notify(119182)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119183)
	}
	__antithesis_instrumentation__.Notify(119181)

	r.quiesceLocked(ctx, lagging)
	return true
}
