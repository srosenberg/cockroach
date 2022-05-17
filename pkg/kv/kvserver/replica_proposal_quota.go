package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/quotapool"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/tracker"
)

func (r *Replica) maybeAcquireProposalQuota(
	ctx context.Context, quota uint64,
) (*quotapool.IntAlloc, error) {
	__antithesis_instrumentation__.Notify(118270)
	r.mu.RLock()
	quotaPool := r.mu.proposalQuota
	desc := *r.mu.state.Desc
	r.mu.RUnlock()

	if quotaPool == nil {
		__antithesis_instrumentation__.Notify(118275)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(118276)
	}
	__antithesis_instrumentation__.Notify(118271)

	if !quotaPoolEnabledForRange(desc) {
		__antithesis_instrumentation__.Notify(118277)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(118278)
	}
	__antithesis_instrumentation__.Notify(118272)

	if log.HasSpanOrEvent(ctx) {
		__antithesis_instrumentation__.Notify(118279)
		if q := quotaPool.ApproximateQuota(); q < quotaPool.Capacity()/10 {
			__antithesis_instrumentation__.Notify(118280)
			log.Eventf(ctx, "quota running low, currently available ~%d", q)
		} else {
			__antithesis_instrumentation__.Notify(118281)
		}
	} else {
		__antithesis_instrumentation__.Notify(118282)
	}
	__antithesis_instrumentation__.Notify(118273)
	alloc, err := quotaPool.Acquire(ctx, quota)

	if errors.HasType(err, (*quotapool.ErrClosed)(nil)) {
		__antithesis_instrumentation__.Notify(118283)
		err = nil
	} else {
		__antithesis_instrumentation__.Notify(118284)
	}
	__antithesis_instrumentation__.Notify(118274)
	return alloc, err
}

func quotaPoolEnabledForRange(desc roachpb.RangeDescriptor) bool {
	__antithesis_instrumentation__.Notify(118285)

	return !bytes.HasPrefix(desc.StartKey, keys.NodeLivenessPrefix)
}

var logSlowRaftProposalQuotaAcquisition = quotapool.OnSlowAcquisition(
	base.SlowRequestThreshold, quotapool.LogSlowAcquisition,
)

func (r *Replica) updateProposalQuotaRaftMuLocked(
	ctx context.Context, lastLeaderID roachpb.ReplicaID,
) {
	__antithesis_instrumentation__.Notify(118286)
	r.mu.Lock()
	defer r.mu.Unlock()

	status := r.mu.internalRaftGroup.BasicStatus()
	if r.mu.leaderID != lastLeaderID {
		__antithesis_instrumentation__.Notify(118290)
		if r.replicaID == r.mu.leaderID {
			__antithesis_instrumentation__.Notify(118292)

			r.mu.proposalQuotaBaseIndex = status.Applied
			if r.mu.proposalQuota != nil {
				__antithesis_instrumentation__.Notify(118295)
				log.Fatal(ctx, "proposalQuota was not nil before becoming the leader")
			} else {
				__antithesis_instrumentation__.Notify(118296)
			}
			__antithesis_instrumentation__.Notify(118293)
			if releaseQueueLen := len(r.mu.quotaReleaseQueue); releaseQueueLen != 0 {
				__antithesis_instrumentation__.Notify(118297)
				log.Fatalf(ctx, "len(r.mu.quotaReleaseQueue) = %d, expected 0", releaseQueueLen)
			} else {
				__antithesis_instrumentation__.Notify(118298)
			}
			__antithesis_instrumentation__.Notify(118294)

			r.mu.proposalQuota = quotapool.NewIntPool(
				"raft proposal",
				uint64(r.store.cfg.RaftProposalQuota),
				logSlowRaftProposalQuotaAcquisition,
			)
			r.mu.lastUpdateTimes = make(map[roachpb.ReplicaID]time.Time)
			r.mu.lastUpdateTimes.updateOnBecomeLeader(r.mu.state.Desc.Replicas().Descriptors(), timeutil.Now())
		} else {
			__antithesis_instrumentation__.Notify(118299)
			if r.mu.proposalQuota != nil {
				__antithesis_instrumentation__.Notify(118300)

				r.mu.proposalQuota.Close("leader change")
				r.mu.proposalQuota.Release(r.mu.quotaReleaseQueue...)
				r.mu.quotaReleaseQueue = nil
				r.mu.proposalQuota = nil
				r.mu.lastUpdateTimes = nil
			} else {
				__antithesis_instrumentation__.Notify(118301)
			}
		}
		__antithesis_instrumentation__.Notify(118291)
		return
	} else {
		__antithesis_instrumentation__.Notify(118302)
		if r.mu.proposalQuota == nil {
			__antithesis_instrumentation__.Notify(118303)
			if r.replicaID == r.mu.leaderID {
				__antithesis_instrumentation__.Notify(118305)
				log.Fatal(ctx, "leader has uninitialized proposalQuota pool")
			} else {
				__antithesis_instrumentation__.Notify(118306)
			}
			__antithesis_instrumentation__.Notify(118304)

			return
		} else {
			__antithesis_instrumentation__.Notify(118307)
		}
	}
	__antithesis_instrumentation__.Notify(118287)

	now := timeutil.Now()

	commitIndex := status.Commit

	minIndex := status.Applied
	r.mu.internalRaftGroup.WithProgress(func(id uint64, _ raft.ProgressType, progress tracker.Progress) {
		__antithesis_instrumentation__.Notify(118308)
		rep, ok := r.mu.state.Desc.GetReplicaDescriptorByID(roachpb.ReplicaID(id))
		if !ok {
			__antithesis_instrumentation__.Notify(118314)
			return
		} else {
			__antithesis_instrumentation__.Notify(118315)
		}
		__antithesis_instrumentation__.Notify(118309)

		if !r.mu.lastUpdateTimes.isFollowerActiveSince(
			ctx, rep.ReplicaID, now, r.store.cfg.RangeLeaseActiveDuration(),
		) {
			__antithesis_instrumentation__.Notify(118316)
			return
		} else {
			__antithesis_instrumentation__.Notify(118317)
		}
		__antithesis_instrumentation__.Notify(118310)

		if err := r.store.cfg.NodeDialer.ConnHealth(rep.NodeID, r.connectionClass.get()); err != nil {
			__antithesis_instrumentation__.Notify(118318)
			return
		} else {
			__antithesis_instrumentation__.Notify(118319)
		}
		__antithesis_instrumentation__.Notify(118311)

		if progress.Match < r.mu.proposalQuotaBaseIndex {
			__antithesis_instrumentation__.Notify(118320)
			return
		} else {
			__antithesis_instrumentation__.Notify(118321)
		}
		__antithesis_instrumentation__.Notify(118312)
		if progress.Match > 0 && func() bool {
			__antithesis_instrumentation__.Notify(118322)
			return progress.Match < minIndex == true
		}() == true {
			__antithesis_instrumentation__.Notify(118323)
			minIndex = progress.Match
		} else {
			__antithesis_instrumentation__.Notify(118324)
		}
		__antithesis_instrumentation__.Notify(118313)

		if rep.ReplicaID == r.mu.lastReplicaAdded && func() bool {
			__antithesis_instrumentation__.Notify(118325)
			return progress.Match >= commitIndex == true
		}() == true {
			__antithesis_instrumentation__.Notify(118326)
			r.mu.lastReplicaAdded = 0
			r.mu.lastReplicaAddedTime = time.Time{}
		} else {
			__antithesis_instrumentation__.Notify(118327)
		}
	})
	__antithesis_instrumentation__.Notify(118288)

	if r.mu.proposalQuotaBaseIndex < minIndex {
		__antithesis_instrumentation__.Notify(118328)

		numReleases := minIndex - r.mu.proposalQuotaBaseIndex

		r.mu.proposalQuota.Release(r.mu.quotaReleaseQueue[:numReleases]...)
		r.mu.quotaReleaseQueue = r.mu.quotaReleaseQueue[numReleases:]
		r.mu.proposalQuotaBaseIndex += numReleases
	} else {
		__antithesis_instrumentation__.Notify(118329)
	}
	__antithesis_instrumentation__.Notify(118289)

	releasableIndex := r.mu.proposalQuotaBaseIndex + uint64(len(r.mu.quotaReleaseQueue))
	if releasableIndex != status.Applied {
		__antithesis_instrumentation__.Notify(118330)
		log.Fatalf(ctx, "proposalQuotaBaseIndex (%d) + quotaReleaseQueueLen (%d) = %d"+
			" must equal the applied index (%d)",
			r.mu.proposalQuotaBaseIndex, len(r.mu.quotaReleaseQueue), releasableIndex,
			status.Applied)
	} else {
		__antithesis_instrumentation__.Notify(118331)
	}
}
