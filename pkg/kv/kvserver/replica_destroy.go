package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

type DestroyReason int

const (
	destroyReasonAlive DestroyReason = iota

	destroyReasonRemoved

	destroyReasonMergePending
)

type destroyStatus struct {
	reason DestroyReason
	err    error
}

func (s destroyStatus) String() string {
	__antithesis_instrumentation__.Notify(117135)
	return fmt.Sprintf("{%v %d}", s.err, s.reason)
}

func (s *destroyStatus) Set(err error, reason DestroyReason) {
	__antithesis_instrumentation__.Notify(117136)
	s.err = err
	s.reason = reason
}

func (s destroyStatus) IsAlive() bool {
	__antithesis_instrumentation__.Notify(117137)
	return s.reason == destroyReasonAlive
}

func (s destroyStatus) Removed() bool {
	__antithesis_instrumentation__.Notify(117138)
	return s.reason == destroyReasonRemoved
}

const mergedTombstoneReplicaID roachpb.ReplicaID = math.MaxInt32

func (r *Replica) preDestroyRaftMuLocked(
	ctx context.Context,
	reader storage.Reader,
	writer storage.Writer,
	nextReplicaID roachpb.ReplicaID,
	clearRangeIDLocalOnly bool,
	mustUseClearRange bool,
) error {
	__antithesis_instrumentation__.Notify(117139)
	r.mu.RLock()
	desc := r.descRLocked()
	removed := r.mu.destroyStatus.Removed()
	r.mu.RUnlock()

	if !removed {
		__antithesis_instrumentation__.Notify(117142)
		log.Fatalf(ctx, "replica not marked as destroyed before call to preDestroyRaftMuLocked: %v", r)
	} else {
		__antithesis_instrumentation__.Notify(117143)
	}
	__antithesis_instrumentation__.Notify(117140)

	err := clearRangeData(desc, reader, writer, clearRangeIDLocalOnly, mustUseClearRange)
	if err != nil {
		__antithesis_instrumentation__.Notify(117144)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117145)
	}
	__antithesis_instrumentation__.Notify(117141)

	return r.setTombstoneKey(ctx, writer, nextReplicaID)
}

func (r *Replica) postDestroyRaftMuLocked(ctx context.Context, ms enginepb.MVCCStats) error {
	__antithesis_instrumentation__.Notify(117146)

	if r.raftMu.sideloaded != nil {
		__antithesis_instrumentation__.Notify(117150)
		if err := r.raftMu.sideloaded.Clear(ctx); err != nil {
			__antithesis_instrumentation__.Notify(117151)
			return err
		} else {
			__antithesis_instrumentation__.Notify(117152)
		}
	} else {
		__antithesis_instrumentation__.Notify(117153)
	}
	__antithesis_instrumentation__.Notify(117147)

	if r.tenantMetricsRef != nil {
		__antithesis_instrumentation__.Notify(117154)
		r.store.metrics.releaseTenant(ctx, r.tenantMetricsRef)
	} else {
		__antithesis_instrumentation__.Notify(117155)
	}
	__antithesis_instrumentation__.Notify(117148)

	if r.tenantLimiter != nil {
		__antithesis_instrumentation__.Notify(117156)
		r.store.tenantRateLimiters.Release(r.tenantLimiter)
		r.tenantLimiter = nil
	} else {
		__antithesis_instrumentation__.Notify(117157)
	}
	__antithesis_instrumentation__.Notify(117149)

	return nil
}

func (r *Replica) destroyRaftMuLocked(ctx context.Context, nextReplicaID roachpb.ReplicaID) error {
	__antithesis_instrumentation__.Notify(117158)
	startTime := timeutil.Now()

	ms := r.GetMVCCStats()
	batch := r.Engine().NewUnindexedBatch(true)
	defer batch.Close()
	clearRangeIDLocalOnly := !r.IsInitialized()
	if err := r.preDestroyRaftMuLocked(
		ctx,
		r.Engine(),
		batch,
		nextReplicaID,
		clearRangeIDLocalOnly,
		false,
	); err != nil {
		__antithesis_instrumentation__.Notify(117163)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117164)
	}
	__antithesis_instrumentation__.Notify(117159)
	preTime := timeutil.Now()

	if err := batch.Commit(true); err != nil {
		__antithesis_instrumentation__.Notify(117165)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117166)
	}
	__antithesis_instrumentation__.Notify(117160)
	commitTime := timeutil.Now()

	if err := r.postDestroyRaftMuLocked(ctx, ms); err != nil {
		__antithesis_instrumentation__.Notify(117167)
		return err
	} else {
		__antithesis_instrumentation__.Notify(117168)
	}
	__antithesis_instrumentation__.Notify(117161)
	if r.IsInitialized() {
		__antithesis_instrumentation__.Notify(117169)
		log.Infof(ctx, "removed %d (%d+%d) keys in %0.0fms [clear=%0.0fms commit=%0.0fms]",
			ms.KeyCount+ms.SysCount, ms.KeyCount, ms.SysCount,
			commitTime.Sub(startTime).Seconds()*1000,
			preTime.Sub(startTime).Seconds()*1000,
			commitTime.Sub(preTime).Seconds()*1000)
	} else {
		__antithesis_instrumentation__.Notify(117170)
		log.Infof(ctx, "removed uninitialized range in %0.0fms [clear=%0.0fms commit=%0.0fms]",
			commitTime.Sub(startTime).Seconds()*1000,
			preTime.Sub(startTime).Seconds()*1000,
			commitTime.Sub(preTime).Seconds()*1000)
	}
	__antithesis_instrumentation__.Notify(117162)
	return nil
}

func (r *Replica) disconnectReplicationRaftMuLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(117171)
	r.raftMu.AssertHeld()
	r.mu.Lock()
	defer r.mu.Unlock()

	if pq := r.mu.proposalQuota; pq != nil {
		__antithesis_instrumentation__.Notify(117174)
		pq.Close("destroyed")
	} else {
		__antithesis_instrumentation__.Notify(117175)
	}
	__antithesis_instrumentation__.Notify(117172)
	r.mu.proposalBuf.FlushLockedWithoutProposing(ctx)
	for _, p := range r.mu.proposals {
		__antithesis_instrumentation__.Notify(117176)
		r.cleanupFailedProposalLocked(p)

		p.finishApplication(ctx, proposalResult{
			Err: roachpb.NewError(
				roachpb.NewAmbiguousResultError(apply.ErrRemoved)),
		})
	}
	__antithesis_instrumentation__.Notify(117173)
	r.mu.internalRaftGroup = nil
}

func (r *Replica) setTombstoneKey(
	ctx context.Context, writer storage.Writer, externalNextReplicaID roachpb.ReplicaID,
) error {
	__antithesis_instrumentation__.Notify(117177)
	r.mu.Lock()
	nextReplicaID := r.mu.state.Desc.NextReplicaID
	if nextReplicaID < externalNextReplicaID {
		__antithesis_instrumentation__.Notify(117180)
		nextReplicaID = externalNextReplicaID
	} else {
		__antithesis_instrumentation__.Notify(117181)
	}
	__antithesis_instrumentation__.Notify(117178)
	if nextReplicaID > r.mu.tombstoneMinReplicaID {
		__antithesis_instrumentation__.Notify(117182)
		r.mu.tombstoneMinReplicaID = nextReplicaID
	} else {
		__antithesis_instrumentation__.Notify(117183)
	}
	__antithesis_instrumentation__.Notify(117179)
	r.mu.Unlock()
	return writeTombstoneKey(ctx, writer, r.RangeID, nextReplicaID)
}

func writeTombstoneKey(
	ctx context.Context,
	writer storage.Writer,
	rangeID roachpb.RangeID,
	nextReplicaID roachpb.ReplicaID,
) error {
	__antithesis_instrumentation__.Notify(117184)
	tombstoneKey := keys.RangeTombstoneKey(rangeID)
	tombstone := &roachpb.RangeTombstone{
		NextReplicaID: nextReplicaID,
	}

	return storage.MVCCBlindPutProto(ctx, writer, nil, tombstoneKey,
		hlc.Timestamp{}, tombstone, nil)
}
