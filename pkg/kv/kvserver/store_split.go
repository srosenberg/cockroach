package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
	raft "go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func splitPreApply(
	ctx context.Context,
	r *Replica,
	readWriter storage.ReadWriter,
	split roachpb.SplitTrigger,
	initClosedTS *hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(126331)

	rightDesc, hasRightDesc := split.RightDesc.GetReplicaDescriptor(r.StoreID())
	_, hasLeftDesc := split.LeftDesc.GetReplicaDescriptor(r.StoreID())
	if !hasRightDesc || func() bool {
		__antithesis_instrumentation__.Notify(126337)
		return !hasLeftDesc == true
	}() == true {
		__antithesis_instrumentation__.Notify(126338)
		log.Fatalf(ctx, "cannot process split on s%s which does not exist in the split: %+v",
			r.StoreID(), split)
	} else {
		__antithesis_instrumentation__.Notify(126339)
	}
	__antithesis_instrumentation__.Notify(126332)

	rightRepl := r.store.GetReplicaIfExists(split.RightDesc.RangeID)

	if rightRepl == nil || func() bool {
		__antithesis_instrumentation__.Notify(126340)
		return rightRepl.isNewerThanSplit(&split) == true
	}() == true {
		__antithesis_instrumentation__.Notify(126341)

		var hs raftpb.HardState
		if rightRepl != nil {
			__antithesis_instrumentation__.Notify(126345)

			if rightRepl.IsInitialized() {
				__antithesis_instrumentation__.Notify(126347)
				log.Fatalf(ctx, "unexpectedly found initialized newer RHS of split: %v", rightRepl.Desc())
			} else {
				__antithesis_instrumentation__.Notify(126348)
			}
			__antithesis_instrumentation__.Notify(126346)
			var err error
			hs, err = rightRepl.raftMu.stateLoader.LoadHardState(ctx, readWriter)
			if err != nil {
				__antithesis_instrumentation__.Notify(126349)
				log.Fatalf(ctx, "failed to load hard state for removed rhs: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(126350)
			}
		} else {
			__antithesis_instrumentation__.Notify(126351)
		}
		__antithesis_instrumentation__.Notify(126342)
		const rangeIDLocalOnly = false
		const mustUseClearRange = false
		if err := clearRangeData(&split.RightDesc, readWriter, readWriter, rangeIDLocalOnly, mustUseClearRange); err != nil {
			__antithesis_instrumentation__.Notify(126352)
			log.Fatalf(ctx, "failed to clear range data for removed rhs: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(126353)
		}
		__antithesis_instrumentation__.Notify(126343)
		if rightRepl != nil {
			__antithesis_instrumentation__.Notify(126354)

			if err := rightRepl.raftMu.stateLoader.SetHardState(ctx, readWriter, hs); err != nil {
				__antithesis_instrumentation__.Notify(126356)
				log.Fatalf(ctx, "failed to set hard state with 0 commit index for removed rhs: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(126357)
			}
			__antithesis_instrumentation__.Notify(126355)
			if err := rightRepl.raftMu.stateLoader.SetRaftReplicaID(
				ctx, readWriter, rightRepl.ReplicaID()); err != nil {
				__antithesis_instrumentation__.Notify(126358)
				log.Fatalf(ctx, "failed to set RaftReplicaID for removed rhs: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(126359)
			}
		} else {
			__antithesis_instrumentation__.Notify(126360)
		}
		__antithesis_instrumentation__.Notify(126344)
		return
	} else {
		__antithesis_instrumentation__.Notify(126361)
	}
	__antithesis_instrumentation__.Notify(126333)

	rsl := stateloader.Make(split.RightDesc.RangeID)
	if err := rsl.SynthesizeRaftState(ctx, readWriter); err != nil {
		__antithesis_instrumentation__.Notify(126362)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(126363)
	}
	__antithesis_instrumentation__.Notify(126334)

	if err := rsl.SetRaftReplicaID(ctx, readWriter, rightDesc.ReplicaID); err != nil {
		__antithesis_instrumentation__.Notify(126364)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(126365)
	}
	__antithesis_instrumentation__.Notify(126335)

	if initClosedTS == nil {
		__antithesis_instrumentation__.Notify(126366)
		initClosedTS = &hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(126367)
	}
	__antithesis_instrumentation__.Notify(126336)
	initClosedTS.Forward(r.GetClosedTimestamp(ctx))
	if err := rsl.SetClosedTimestamp(ctx, readWriter, initClosedTS); err != nil {
		__antithesis_instrumentation__.Notify(126368)
		log.Fatalf(ctx, "%s", err)
	} else {
		__antithesis_instrumentation__.Notify(126369)
	}
}

func splitPostApply(
	ctx context.Context, deltaMS enginepb.MVCCStats, split *roachpb.SplitTrigger, r *Replica,
) {
	__antithesis_instrumentation__.Notify(126370)

	rightReplOrNil := prepareRightReplicaForSplit(ctx, split, r)

	if err := r.store.SplitRange(ctx, r, rightReplOrNil, split); err != nil {
		__antithesis_instrumentation__.Notify(126373)

		log.Fatalf(ctx, "%s: failed to update Store after split: %+v", r, err)
	} else {
		__antithesis_instrumentation__.Notify(126374)
	}
	__antithesis_instrumentation__.Notify(126371)

	if rightReplOrNil != nil {
		__antithesis_instrumentation__.Notify(126375)
		if _, ok := rightReplOrNil.TenantID(); ok {
			__antithesis_instrumentation__.Notify(126376)

			rightReplOrNil.store.metrics.addMVCCStats(ctx, rightReplOrNil.tenantMetricsRef, deltaMS)
		} else {
			__antithesis_instrumentation__.Notify(126377)
			log.Fatalf(ctx, "%s: found replica which is RHS of a split "+
				"without a valid tenant ID", rightReplOrNil)
		}
	} else {
		__antithesis_instrumentation__.Notify(126378)
	}
	__antithesis_instrumentation__.Notify(126372)

	now := r.store.Clock().NowAsClockTimestamp()

	r.store.splitQueue.MaybeAddAsync(ctx, r, now)

	r.store.replicateQueue.MaybeAddAsync(ctx, r, now)

	if rightReplOrNil != nil {
		__antithesis_instrumentation__.Notify(126379)
		r.store.splitQueue.MaybeAddAsync(ctx, rightReplOrNil, now)
		r.store.replicateQueue.MaybeAddAsync(ctx, rightReplOrNil, now)
		if len(split.RightDesc.Replicas().Descriptors()) == 1 {
			__antithesis_instrumentation__.Notify(126380)

			r.store.enqueueRaftUpdateCheck(rightReplOrNil.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(126381)
		}
	} else {
		__antithesis_instrumentation__.Notify(126382)
	}
}

func prepareRightReplicaForSplit(
	ctx context.Context, split *roachpb.SplitTrigger, r *Replica,
) (rightReplicaOrNil *Replica) {
	__antithesis_instrumentation__.Notify(126383)

	r.mu.RLock()
	minLeaseProposedTS := r.mu.minLeaseProposedTS
	r.mu.RUnlock()

	rightRepl := r.store.GetReplicaIfExists(split.RightDesc.RangeID)

	_, _ = r.acquireSplitLock, splitPostApply
	if rightRepl == nil {
		__antithesis_instrumentation__.Notify(126389)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126390)
	}
	__antithesis_instrumentation__.Notify(126384)

	rightRepl.mu.Lock()
	defer rightRepl.mu.Unlock()

	if rightRepl.isNewerThanSplitRLocked(split) {
		__antithesis_instrumentation__.Notify(126391)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(126392)
	}
	__antithesis_instrumentation__.Notify(126385)

	err := rightRepl.loadRaftMuLockedReplicaMuLocked(&split.RightDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(126393)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(126394)
	}
	__antithesis_instrumentation__.Notify(126386)

	rightRepl.mu.minLeaseProposedTS = minLeaseProposedTS

	rightRepl.leasePostApplyLocked(ctx,
		rightRepl.mu.state.Lease,
		rightRepl.mu.state.Lease,
		nil,
		assertNoLeaseJump)

	err = rightRepl.withRaftGroupLocked(true, func(r *raft.RawNode) (unquiesceAndWakeLeader bool, _ error) {
		__antithesis_instrumentation__.Notify(126395)
		return true, nil
	})
	__antithesis_instrumentation__.Notify(126387)
	if err != nil {
		__antithesis_instrumentation__.Notify(126396)
		log.Fatalf(ctx, "unable to create raft group for right-hand range in split: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(126397)
	}
	__antithesis_instrumentation__.Notify(126388)

	return rightRepl
}

func (s *Store) SplitRange(
	ctx context.Context, leftRepl, rightReplOrNil *Replica, split *roachpb.SplitTrigger,
) error {
	__antithesis_instrumentation__.Notify(126398)
	rightDesc := &split.RightDesc
	newLeftDesc := &split.LeftDesc
	oldLeftDesc := leftRepl.Desc()
	if !bytes.Equal(oldLeftDesc.EndKey, rightDesc.EndKey) || func() bool {
		__antithesis_instrumentation__.Notify(126402)
		return bytes.Compare(oldLeftDesc.StartKey, rightDesc.StartKey) >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(126403)
		return errors.Errorf("left range is not splittable by right range: %+v, %+v", oldLeftDesc, rightDesc)
	} else {
		__antithesis_instrumentation__.Notify(126404)
	}
	__antithesis_instrumentation__.Notify(126399)

	s.mu.Lock()
	defer s.mu.Unlock()
	if exRng, ok := s.mu.uninitReplicas[rightDesc.RangeID]; rightReplOrNil != nil && func() bool {
		__antithesis_instrumentation__.Notify(126405)
		return ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(126406)

		if exRng != rightReplOrNil {
			__antithesis_instrumentation__.Notify(126408)
			log.Fatalf(ctx, "found unexpected uninitialized replica: %s vs %s", exRng, rightReplOrNil)
		} else {
			__antithesis_instrumentation__.Notify(126409)
		}
		__antithesis_instrumentation__.Notify(126407)

		delete(s.mu.uninitReplicas, rightDesc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(126410)
	}
	__antithesis_instrumentation__.Notify(126400)

	leftRepl.setDescRaftMuLocked(ctx, newLeftDesc)

	leftRepl.concMgr.OnRangeSplit()

	leftRepl.leaseholderStats.resetRequestCounts()

	if rightReplOrNil == nil {
		__antithesis_instrumentation__.Notify(126411)
		throwawayRightWriteStats := new(replicaStats)
		leftRepl.writeStats.splitRequestCounts(throwawayRightWriteStats)
	} else {
		__antithesis_instrumentation__.Notify(126412)
		rightRepl := rightReplOrNil
		leftRepl.writeStats.splitRequestCounts(rightRepl.writeStats)
		if err := s.addReplicaInternalLocked(rightRepl); err != nil {
			__antithesis_instrumentation__.Notify(126415)
			return errors.Wrapf(err, "unable to add replica %v", rightRepl)
		} else {
			__antithesis_instrumentation__.Notify(126416)
		}
		__antithesis_instrumentation__.Notify(126413)

		if err := rightRepl.updateRangeInfo(ctx, rightRepl.Desc()); err != nil {
			__antithesis_instrumentation__.Notify(126417)
			return err
		} else {
			__antithesis_instrumentation__.Notify(126418)
		}
		__antithesis_instrumentation__.Notify(126414)

		s.metrics.ReplicaCount.Inc(1)
		s.maybeGossipOnCapacityChange(ctx, rangeAddEvent)
	}
	__antithesis_instrumentation__.Notify(126401)

	return nil
}
