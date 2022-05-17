package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary/rspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func isTrivial(r *kvserverpb.ReplicatedEvalResult) bool {
	__antithesis_instrumentation__.Notify(115092)

	if r.State != nil {
		__antithesis_instrumentation__.Notify(115094)
		stateAllowlist := *r.State

		if stateAllowlist.Stats != nil && func() bool {
			__antithesis_instrumentation__.Notify(115096)
			return (*stateAllowlist.Stats == enginepb.MVCCStats{}) == true
		}() == true {
			__antithesis_instrumentation__.Notify(115097)
			stateAllowlist.Stats = nil
		} else {
			__antithesis_instrumentation__.Notify(115098)
		}
		__antithesis_instrumentation__.Notify(115095)
		if stateAllowlist != (kvserverpb.ReplicaState{}) {
			__antithesis_instrumentation__.Notify(115099)
			return false
		} else {
			__antithesis_instrumentation__.Notify(115100)
		}
	} else {
		__antithesis_instrumentation__.Notify(115101)
	}
	__antithesis_instrumentation__.Notify(115093)

	allowlist := *r
	allowlist.Delta = enginepb.MVCCStatsDelta{}
	allowlist.WriteTimestamp = hlc.Timestamp{}
	allowlist.DeprecatedDelta = nil
	allowlist.PrevLeaseProposal = nil
	allowlist.State = nil
	return allowlist.IsZero()
}

func clearTrivialReplicatedEvalResultFields(r *kvserverpb.ReplicatedEvalResult) {
	__antithesis_instrumentation__.Notify(115102)

	r.IsLeaseRequest = false
	r.WriteTimestamp = hlc.Timestamp{}
	r.PrevLeaseProposal = nil

	if haveState := r.State != nil; haveState {
		__antithesis_instrumentation__.Notify(115104)
		r.State.Stats = nil

		r.State.RaftAppliedIndexTerm = 0
		if *r.State == (kvserverpb.ReplicaState{}) {
			__antithesis_instrumentation__.Notify(115105)
			r.State = nil
		} else {
			__antithesis_instrumentation__.Notify(115106)
		}
	} else {
		__antithesis_instrumentation__.Notify(115107)
	}
	__antithesis_instrumentation__.Notify(115103)
	r.Delta = enginepb.MVCCStatsDelta{}

	r.MVCCHistoryMutation = nil
}

func (r *Replica) prepareLocalResult(ctx context.Context, cmd *replicatedCmd) {
	__antithesis_instrumentation__.Notify(115108)
	if !cmd.IsLocal() {
		__antithesis_instrumentation__.Notify(115114)
		return
	} else {
		__antithesis_instrumentation__.Notify(115115)
	}
	__antithesis_instrumentation__.Notify(115109)

	var pErr *roachpb.Error
	if filter := r.store.cfg.TestingKnobs.TestingPostApplyFilter; filter != nil {
		__antithesis_instrumentation__.Notify(115116)
		var newPropRetry int
		newPropRetry, pErr = filter(kvserverbase.ApplyFilterArgs{
			CmdID:                cmd.idKey,
			ReplicatedEvalResult: *cmd.replicatedResult(),
			StoreID:              r.store.StoreID(),
			RangeID:              r.RangeID,
			Req:                  cmd.proposal.Request,
		})
		if cmd.proposalRetry == 0 {
			__antithesis_instrumentation__.Notify(115117)
			cmd.proposalRetry = proposalReevaluationReason(newPropRetry)
		} else {
			__antithesis_instrumentation__.Notify(115118)
		}
	} else {
		__antithesis_instrumentation__.Notify(115119)
	}
	__antithesis_instrumentation__.Notify(115110)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(115120)
		pErr = cmd.forcedErr
	} else {
		__antithesis_instrumentation__.Notify(115121)
	}
	__antithesis_instrumentation__.Notify(115111)

	if cmd.proposalRetry != proposalNoReevaluation && func() bool {
		__antithesis_instrumentation__.Notify(115122)
		return pErr == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(115123)
		log.Fatalf(ctx, "proposal with nontrivial retry behavior, but no error: %+v", cmd.proposal)
	} else {
		__antithesis_instrumentation__.Notify(115124)
	}
	__antithesis_instrumentation__.Notify(115112)
	if pErr != nil {
		__antithesis_instrumentation__.Notify(115125)

		switch cmd.proposalRetry {
		case proposalNoReevaluation:
			__antithesis_instrumentation__.Notify(115126)
			cmd.response.Err = pErr
		case proposalIllegalLeaseIndex:
			__antithesis_instrumentation__.Notify(115127)

			pErr = r.tryReproposeWithNewLeaseIndex(ctx, cmd)
			if pErr != nil {
				__antithesis_instrumentation__.Notify(115129)
				log.Warningf(ctx, "failed to repropose with new lease index: %s", pErr)
				cmd.response.Err = pErr
			} else {
				__antithesis_instrumentation__.Notify(115130)

				cmd.proposal = nil
				return
			}
		default:
			__antithesis_instrumentation__.Notify(115128)
			panic("unexpected")
		}
	} else {
		__antithesis_instrumentation__.Notify(115131)
		if cmd.proposal.Local.Reply != nil {
			__antithesis_instrumentation__.Notify(115132)
			cmd.response.Reply = cmd.proposal.Local.Reply
		} else {
			__antithesis_instrumentation__.Notify(115133)
			log.Fatalf(ctx, "proposal must return either a reply or an error: %+v", cmd.proposal)
		}
	}
	__antithesis_instrumentation__.Notify(115113)
	cmd.response.EncounteredIntents = cmd.proposal.Local.DetachEncounteredIntents()
	cmd.response.EndTxns = cmd.proposal.Local.DetachEndTxns(pErr != nil)
	if pErr == nil {
		__antithesis_instrumentation__.Notify(115134)
		cmd.localResult = cmd.proposal.Local
	} else {
		__antithesis_instrumentation__.Notify(115135)
		if cmd.localResult != nil {
			__antithesis_instrumentation__.Notify(115136)
			log.Fatalf(ctx, "shouldn't have a local result if command processing failed. pErr: %s", pErr)
		} else {
			__antithesis_instrumentation__.Notify(115137)
		}
	}
}

func (r *Replica) tryReproposeWithNewLeaseIndex(
	ctx context.Context, cmd *replicatedCmd,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(115138)

	p := cmd.proposal
	if p.applied || func() bool {
		__antithesis_instrumentation__.Notify(115142)
		return cmd.raftCmd.MaxLeaseIndex != p.command.MaxLeaseIndex == true
	}() == true {
		__antithesis_instrumentation__.Notify(115143)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(115144)
	}
	__antithesis_instrumentation__.Notify(115139)

	minTS, tok := r.mu.proposalBuf.TrackEvaluatingRequest(ctx, p.Request.WriteTimestamp())
	defer tok.DoneIfNotMoved(ctx)

	if p.Request.AppliesTimestampCache() && func() bool {
		__antithesis_instrumentation__.Notify(115145)
		return p.Request.WriteTimestamp().LessEq(minTS) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115146)

		err := newNotLeaseHolderError(
			*r.mu.state.Lease,
			r.store.StoreID(),
			r.mu.state.Desc,
			"reproposal failed due to closed timestamp",
		)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(115147)
	}
	__antithesis_instrumentation__.Notify(115140)

	log.VEventf(ctx, 2, "retry: proposalIllegalLeaseIndex")

	pErr := r.propose(ctx, p, tok.Move(ctx))
	if pErr != nil {
		__antithesis_instrumentation__.Notify(115148)
		return pErr
	} else {
		__antithesis_instrumentation__.Notify(115149)
	}
	__antithesis_instrumentation__.Notify(115141)
	log.VEventf(ctx, 2, "reproposed command %x", cmd.idKey)
	return nil
}

func (r *Replica) handleSplitResult(ctx context.Context, split *kvserverpb.Split) {
	__antithesis_instrumentation__.Notify(115150)
	splitPostApply(ctx, split.RHSDelta, &split.SplitTrigger, r)
}

func (r *Replica) handleMergeResult(ctx context.Context, merge *kvserverpb.Merge) {
	__antithesis_instrumentation__.Notify(115151)
	if err := r.store.MergeRange(
		ctx,
		r,
		merge.LeftDesc,
		merge.RightDesc,
		merge.FreezeStart,
		merge.RightClosedTimestamp,
		merge.RightReadSummary,
	); err != nil {
		__antithesis_instrumentation__.Notify(115152)

		log.Fatalf(ctx, "failed to update store after merging range: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(115153)
	}
}

func (r *Replica) handleDescResult(ctx context.Context, desc *roachpb.RangeDescriptor) {
	__antithesis_instrumentation__.Notify(115154)
	r.setDescRaftMuLocked(ctx, desc)
}

func (r *Replica) handleLeaseResult(
	ctx context.Context, lease *roachpb.Lease, priorReadSum *rspb.ReadSummary,
) {
	__antithesis_instrumentation__.Notify(115155)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.leasePostApplyLocked(ctx,
		r.mu.state.Lease,
		lease,
		priorReadSum,
		assertNoLeaseJump)
}

func (r *Replica) handleTruncatedStateResult(
	ctx context.Context, t *roachpb.RaftTruncatedState, expectedFirstIndexPreTruncation uint64,
) (raftLogDelta int64, expectedFirstIndexWasAccurate bool) {
	__antithesis_instrumentation__.Notify(115156)
	r.mu.Lock()
	expectedFirstIndexWasAccurate =
		r.mu.state.TruncatedState.Index+1 == expectedFirstIndexPreTruncation
	r.mu.state.TruncatedState = t
	r.mu.Unlock()

	r.store.raftEntryCache.Clear(r.RangeID, t.Index+1)

	log.Eventf(ctx, "truncating sideloaded storage up to (and including) index %d", t.Index)
	size, _, err := r.raftMu.sideloaded.TruncateTo(ctx, t.Index+1)
	if err != nil {
		__antithesis_instrumentation__.Notify(115158)

		log.Errorf(ctx, "while removing sideloaded files during log truncation: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(115159)
	}
	__antithesis_instrumentation__.Notify(115157)
	return -size, expectedFirstIndexWasAccurate
}

func (r *Replica) handleGCThresholdResult(ctx context.Context, thresh *hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(115160)
	if thresh.IsEmpty() {
		__antithesis_instrumentation__.Notify(115162)
		return
	} else {
		__antithesis_instrumentation__.Notify(115163)
	}
	__antithesis_instrumentation__.Notify(115161)
	r.mu.Lock()
	r.mu.state.GCThreshold = thresh
	r.mu.Unlock()
}

func (r *Replica) handleVersionResult(ctx context.Context, version *roachpb.Version) {
	__antithesis_instrumentation__.Notify(115164)
	if (*version == roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(115166)
		log.Fatal(ctx, "not expecting empty replica version downstream of raft")
	} else {
		__antithesis_instrumentation__.Notify(115167)
	}
	__antithesis_instrumentation__.Notify(115165)
	r.mu.Lock()
	r.mu.state.Version = version
	r.mu.Unlock()
}

func (r *Replica) handleComputeChecksumResult(ctx context.Context, cc *kvserverpb.ComputeChecksum) {
	__antithesis_instrumentation__.Notify(115168)
	r.computeChecksumPostApply(ctx, *cc)
}

func (r *Replica) handleChangeReplicasResult(
	ctx context.Context, chng *kvserverpb.ChangeReplicas,
) (changeRemovedReplica bool) {
	__antithesis_instrumentation__.Notify(115169)

	if ds, _ := r.IsDestroyed(); ds != destroyReasonRemoved {
		__antithesis_instrumentation__.Notify(115174)
		return false
	} else {
		__antithesis_instrumentation__.Notify(115175)
	}
	__antithesis_instrumentation__.Notify(115170)

	if log.V(1) {
		__antithesis_instrumentation__.Notify(115176)
		log.Infof(ctx, "removing replica due to ChangeReplicasTrigger: %v", chng)
	} else {
		__antithesis_instrumentation__.Notify(115177)
	}
	__antithesis_instrumentation__.Notify(115171)

	if _, err := r.store.removeInitializedReplicaRaftMuLocked(ctx, r, chng.NextReplicaID(), RemoveOptions{

		DestroyData: false,
	}); err != nil {
		__antithesis_instrumentation__.Notify(115178)
		log.Fatalf(ctx, "failed to remove replica: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(115179)
	}
	__antithesis_instrumentation__.Notify(115172)

	if err := r.postDestroyRaftMuLocked(ctx, r.GetMVCCStats()); err != nil {
		__antithesis_instrumentation__.Notify(115180)
		log.Fatalf(ctx, "failed to run Replica postDestroy: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(115181)
	}
	__antithesis_instrumentation__.Notify(115173)

	return true
}

func (r *Replica) handleRaftLogDeltaResult(ctx context.Context, delta int64, isDeltaTrusted bool) {
	__antithesis_instrumentation__.Notify(115182)
	(*raftTruncatorReplica)(r).setTruncationDeltaAndTrusted(delta, isDeltaTrusted)
}
