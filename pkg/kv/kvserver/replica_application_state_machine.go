package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/ctpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/envutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"github.com/kr/pretty"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type applyCommittedEntriesStats struct {
	batchesProcessed     int
	entriesProcessed     int
	stateAssertions      int
	numEmptyEntries      int
	numConfChangeEntries int
}

type nonDeterministicFailure struct {
	wrapped  error
	safeExpl string
}

func makeNonDeterministicFailure(format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(115183)
	err := errors.AssertionFailedWithDepthf(1, format, args...)
	return &nonDeterministicFailure{
		wrapped:  err,
		safeExpl: err.Error(),
	}
}

func wrapWithNonDeterministicFailure(err error, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(115184)
	return &nonDeterministicFailure{
		wrapped:  errors.Wrapf(err, format, args...),
		safeExpl: fmt.Sprintf(format, args...),
	}
}

func (e *nonDeterministicFailure) Error() string {
	__antithesis_instrumentation__.Notify(115185)
	return fmt.Sprintf("non-deterministic failure: %s", e.wrapped.Error())
}

func (e *nonDeterministicFailure) Cause() error {
	__antithesis_instrumentation__.Notify(115186)
	return e.wrapped
}

func (e *nonDeterministicFailure) Unwrap() error {
	__antithesis_instrumentation__.Notify(115187)
	return e.wrapped
}

type replicaStateMachine struct {
	r *Replica

	batch replicaAppBatch

	ephemeralBatch ephemeralReplicaAppBatch

	stats applyCommittedEntriesStats
}

func (r *Replica) getStateMachine() *replicaStateMachine {
	__antithesis_instrumentation__.Notify(115188)
	sm := &r.raftMu.stateMachine
	sm.r = r
	return sm
}

func (r *Replica) shouldApplyCommand(
	ctx context.Context, cmd *replicatedCmd, replicaState *kvserverpb.ReplicaState,
) bool {
	__antithesis_instrumentation__.Notify(115189)
	cmd.leaseIndex, cmd.proposalRetry, cmd.forcedErr = checkForcedErr(
		ctx, cmd.idKey, &cmd.raftCmd, cmd.IsLocal(), replicaState,
	)

	if filter := r.store.cfg.TestingKnobs.TestingApplyFilter; cmd.forcedErr != nil || func() bool {
		__antithesis_instrumentation__.Notify(115191)
		return filter != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(115192)
		args := kvserverbase.ApplyFilterArgs{
			CmdID:                cmd.idKey,
			ReplicatedEvalResult: *cmd.replicatedResult(),
			StoreID:              r.store.StoreID(),
			RangeID:              r.RangeID,
			ForcedError:          cmd.forcedErr,
		}
		if cmd.forcedErr == nil {
			__antithesis_instrumentation__.Notify(115193)
			if cmd.IsLocal() {
				__antithesis_instrumentation__.Notify(115195)
				args.Req = cmd.proposal.Request
			} else {
				__antithesis_instrumentation__.Notify(115196)
			}
			__antithesis_instrumentation__.Notify(115194)
			newPropRetry, newForcedErr := filter(args)
			cmd.forcedErr = newForcedErr
			if cmd.proposalRetry == 0 {
				__antithesis_instrumentation__.Notify(115197)
				cmd.proposalRetry = proposalReevaluationReason(newPropRetry)
			} else {
				__antithesis_instrumentation__.Notify(115198)
			}
		} else {
			__antithesis_instrumentation__.Notify(115199)
			if feFilter := r.store.cfg.TestingKnobs.TestingApplyForcedErrFilter; feFilter != nil {
				__antithesis_instrumentation__.Notify(115200)
				newPropRetry, newForcedErr := filter(args)
				cmd.forcedErr = newForcedErr
				if cmd.proposalRetry == 0 {
					__antithesis_instrumentation__.Notify(115201)
					cmd.proposalRetry = proposalReevaluationReason(newPropRetry)
				} else {
					__antithesis_instrumentation__.Notify(115202)
				}
			} else {
				__antithesis_instrumentation__.Notify(115203)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(115204)
	}
	__antithesis_instrumentation__.Notify(115190)
	return cmd.forcedErr == nil
}

var noopOnEmptyRaftCommandErr = roachpb.NewErrorf("no-op on empty Raft entry")

var noopOnProbeCommandErr = roachpb.NewErrorf("no-op on ProbeRequest")

func checkForcedErr(
	ctx context.Context,
	idKey kvserverbase.CmdIDKey,
	raftCmd *kvserverpb.RaftCommand,
	isLocal bool,
	replicaState *kvserverpb.ReplicaState,
) (uint64, proposalReevaluationReason, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(115205)
	if raftCmd.ReplicatedEvalResult.IsProbe {
		__antithesis_instrumentation__.Notify(115213)

		return 0, proposalNoReevaluation, noopOnProbeCommandErr
	} else {
		__antithesis_instrumentation__.Notify(115214)
	}
	__antithesis_instrumentation__.Notify(115206)
	leaseIndex := replicaState.LeaseAppliedIndex
	isLeaseRequest := raftCmd.ReplicatedEvalResult.IsLeaseRequest
	var requestedLease roachpb.Lease
	if isLeaseRequest {
		__antithesis_instrumentation__.Notify(115215)
		requestedLease = *raftCmd.ReplicatedEvalResult.State.Lease
	} else {
		__antithesis_instrumentation__.Notify(115216)
	}
	__antithesis_instrumentation__.Notify(115207)
	if idKey == "" {
		__antithesis_instrumentation__.Notify(115217)

		return leaseIndex, proposalNoReevaluation, noopOnEmptyRaftCommandErr
	} else {
		__antithesis_instrumentation__.Notify(115218)
	}
	__antithesis_instrumentation__.Notify(115208)

	leaseMismatch := false
	if raftCmd.DeprecatedProposerLease != nil {
		__antithesis_instrumentation__.Notify(115219)

		leaseMismatch = !raftCmd.DeprecatedProposerLease.Equivalent(*replicaState.Lease)
	} else {
		__antithesis_instrumentation__.Notify(115220)
		leaseMismatch = raftCmd.ProposerLeaseSequence != replicaState.Lease.Sequence
		if !leaseMismatch && func() bool {
			__antithesis_instrumentation__.Notify(115221)
			return isLeaseRequest == true
		}() == true {
			__antithesis_instrumentation__.Notify(115222)

			if replicaState.Lease.Sequence == requestedLease.Sequence {
				__antithesis_instrumentation__.Notify(115224)

				leaseMismatch = !replicaState.Lease.Equivalent(requestedLease)
			} else {
				__antithesis_instrumentation__.Notify(115225)
			}
			__antithesis_instrumentation__.Notify(115223)

			if raftCmd.ReplicatedEvalResult.PrevLeaseProposal != nil && func() bool {
				__antithesis_instrumentation__.Notify(115226)
				return (!raftCmd.ReplicatedEvalResult.PrevLeaseProposal.Equal(replicaState.Lease.ProposedTS)) == true
			}() == true {
				__antithesis_instrumentation__.Notify(115227)
				leaseMismatch = true
			} else {
				__antithesis_instrumentation__.Notify(115228)
			}
		} else {
			__antithesis_instrumentation__.Notify(115229)
		}
	}
	__antithesis_instrumentation__.Notify(115209)
	if leaseMismatch {
		__antithesis_instrumentation__.Notify(115230)
		log.VEventf(
			ctx, 1,
			"command with lease #%d incompatible to %v",
			raftCmd.ProposerLeaseSequence, *replicaState.Lease,
		)
		if isLeaseRequest {
			__antithesis_instrumentation__.Notify(115232)

			return leaseIndex, proposalNoReevaluation, roachpb.NewError(&roachpb.LeaseRejectedError{
				Existing:  *replicaState.Lease,
				Requested: requestedLease,
				Message:   "proposed under invalid lease",
			})
		} else {
			__antithesis_instrumentation__.Notify(115233)
		}
		__antithesis_instrumentation__.Notify(115231)

		nlhe := newNotLeaseHolderError(
			*replicaState.Lease, 0, replicaState.Desc,
			fmt.Sprintf(
				"stale proposal: command was proposed under lease #%d but is being applied "+
					"under lease: %s", raftCmd.ProposerLeaseSequence, replicaState.Lease))
		return leaseIndex, proposalNoReevaluation, roachpb.NewError(nlhe)
	} else {
		__antithesis_instrumentation__.Notify(115234)
	}
	__antithesis_instrumentation__.Notify(115210)

	if isLeaseRequest {
		__antithesis_instrumentation__.Notify(115235)

		if _, ok := replicaState.Desc.GetReplicaDescriptor(requestedLease.Replica.StoreID); !ok {
			__antithesis_instrumentation__.Notify(115236)
			return leaseIndex, proposalNoReevaluation, roachpb.NewError(&roachpb.LeaseRejectedError{
				Existing:  *replicaState.Lease,
				Requested: requestedLease,
				Message:   "replica not part of range",
			})
		} else {
			__antithesis_instrumentation__.Notify(115237)
		}
	} else {
		__antithesis_instrumentation__.Notify(115238)
		if replicaState.LeaseAppliedIndex < raftCmd.MaxLeaseIndex {
			__antithesis_instrumentation__.Notify(115239)

			leaseIndex = raftCmd.MaxLeaseIndex
		} else {
			__antithesis_instrumentation__.Notify(115240)

			retry := proposalNoReevaluation
			if isLocal {
				__antithesis_instrumentation__.Notify(115242)
				log.VEventf(
					ctx, 1,
					"retry proposal %x: applied at lease index %d, required < %d",
					idKey, leaseIndex, raftCmd.MaxLeaseIndex,
				)
				retry = proposalIllegalLeaseIndex
			} else {
				__antithesis_instrumentation__.Notify(115243)
			}
			__antithesis_instrumentation__.Notify(115241)
			return leaseIndex, retry, roachpb.NewErrorf(
				"command observed at lease index %d, but required < %d", leaseIndex, raftCmd.MaxLeaseIndex,
			)
		}
	}
	__antithesis_instrumentation__.Notify(115211)

	wts := raftCmd.ReplicatedEvalResult.WriteTimestamp
	if !wts.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(115244)
		return wts.LessEq(*replicaState.GCThreshold) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115245)
		return leaseIndex, proposalNoReevaluation, roachpb.NewError(&roachpb.BatchTimestampBeforeGCError{
			Timestamp: wts,
			Threshold: *replicaState.GCThreshold,
		})
	} else {
		__antithesis_instrumentation__.Notify(115246)
	}
	__antithesis_instrumentation__.Notify(115212)
	return leaseIndex, proposalNoReevaluation, nil
}

func (sm *replicaStateMachine) NewBatch(ephemeral bool) apply.Batch {
	__antithesis_instrumentation__.Notify(115247)
	r := sm.r
	if ephemeral {
		__antithesis_instrumentation__.Notify(115249)
		mb := &sm.ephemeralBatch
		mb.r = r
		r.mu.RLock()
		mb.state = r.mu.state
		r.mu.RUnlock()
		return mb
	} else {
		__antithesis_instrumentation__.Notify(115250)
	}
	__antithesis_instrumentation__.Notify(115248)
	b := &sm.batch
	b.r = r
	b.sm = sm
	b.batch = r.store.engine.NewBatch()
	r.mu.RLock()
	b.state = r.mu.state
	b.state.Stats = &b.stats
	*b.state.Stats = *r.mu.state.Stats
	b.closedTimestampSetter = r.mu.closedTimestampSetter
	r.mu.RUnlock()
	b.start = timeutil.Now()
	return b
}

type replicaAppBatch struct {
	r  *Replica
	sm *replicaStateMachine

	batch storage.Batch

	state kvserverpb.ReplicaState

	closedTimestampSetter closedTimestampSetterInfo

	stats enginepb.MVCCStats

	changeRemovesReplica bool

	entries      int
	emptyEntries int
	mutations    int
	start        time.Time
}

func (b *replicaAppBatch) Stage(
	ctx context.Context, cmdI apply.Command,
) (apply.CheckedCommand, error) {
	__antithesis_instrumentation__.Notify(115251)
	cmd := cmdI.(*replicatedCmd)
	if cmd.ent.Index == 0 {
		__antithesis_instrumentation__.Notify(115261)
		return nil, makeNonDeterministicFailure("processRaftCommand requires a non-zero index")
	} else {
		__antithesis_instrumentation__.Notify(115262)
	}
	__antithesis_instrumentation__.Notify(115252)
	if idx, applied := cmd.ent.Index, b.state.RaftAppliedIndex; idx != applied+1 {
		__antithesis_instrumentation__.Notify(115263)

		return nil, makeNonDeterministicFailure("applied index jumped from %d to %d", applied, idx)
	} else {
		__antithesis_instrumentation__.Notify(115264)
	}
	__antithesis_instrumentation__.Notify(115253)
	if log.V(4) {
		__antithesis_instrumentation__.Notify(115265)
		log.Infof(ctx, "processing command %x: raftIndex=%d maxLeaseIndex=%d closedts=%s",
			cmd.idKey, cmd.ent.Index, cmd.raftCmd.MaxLeaseIndex, cmd.raftCmd.ClosedTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(115266)
	}
	__antithesis_instrumentation__.Notify(115254)

	if !b.r.shouldApplyCommand(ctx, cmd, &b.state) {
		__antithesis_instrumentation__.Notify(115267)
		log.VEventf(ctx, 1, "applying command with forced error: %s", cmd.forcedErr)

		cmd.raftCmd.ReplicatedEvalResult = kvserverpb.ReplicatedEvalResult{}
		cmd.raftCmd.WriteBatch = nil
		cmd.raftCmd.LogicalOpLog = nil
		cmd.raftCmd.ClosedTimestamp = nil
	} else {
		__antithesis_instrumentation__.Notify(115268)
		if err := b.assertNoCmdClosedTimestampRegression(ctx, cmd); err != nil {
			__antithesis_instrumentation__.Notify(115271)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(115272)
		}
		__antithesis_instrumentation__.Notify(115269)
		if err := b.assertNoWriteBelowClosedTimestamp(cmd); err != nil {
			__antithesis_instrumentation__.Notify(115273)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(115274)
		}
		__antithesis_instrumentation__.Notify(115270)
		log.Event(ctx, "applying command")
	}
	__antithesis_instrumentation__.Notify(115255)

	if splitMergeUnlock, err := b.r.maybeAcquireSplitMergeLock(ctx, cmd.raftCmd); err != nil {
		__antithesis_instrumentation__.Notify(115275)
		if cmd.raftCmd.ReplicatedEvalResult.Split != nil {
			__antithesis_instrumentation__.Notify(115277)
			err = wrapWithNonDeterministicFailure(err, "unable to acquire split lock")
		} else {
			__antithesis_instrumentation__.Notify(115278)
			err = wrapWithNonDeterministicFailure(err, "unable to acquire merge lock")
		}
		__antithesis_instrumentation__.Notify(115276)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(115279)
		if splitMergeUnlock != nil {
			__antithesis_instrumentation__.Notify(115280)

			cmd.splitMergeUnlock = splitMergeUnlock
		} else {
			__antithesis_instrumentation__.Notify(115281)
		}
	}
	__antithesis_instrumentation__.Notify(115256)

	b.migrateReplicatedResult(ctx, cmd)

	if err := b.runPreApplyTriggersBeforeStagingWriteBatch(ctx, cmd); err != nil {
		__antithesis_instrumentation__.Notify(115282)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(115283)
	}
	__antithesis_instrumentation__.Notify(115257)

	if err := b.stageWriteBatch(ctx, cmd); err != nil {
		__antithesis_instrumentation__.Notify(115284)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(115285)
	}
	__antithesis_instrumentation__.Notify(115258)

	if err := b.runPreApplyTriggersAfterStagingWriteBatch(ctx, cmd); err != nil {
		__antithesis_instrumentation__.Notify(115286)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(115287)
	}
	__antithesis_instrumentation__.Notify(115259)

	b.stageTrivialReplicatedEvalResult(ctx, cmd)
	b.entries++
	if len(cmd.ent.Data) == 0 {
		__antithesis_instrumentation__.Notify(115288)
		b.emptyEntries++
	} else {
		__antithesis_instrumentation__.Notify(115289)
	}
	__antithesis_instrumentation__.Notify(115260)

	return cmd, nil
}

func (b *replicaAppBatch) migrateReplicatedResult(ctx context.Context, cmd *replicatedCmd) {
	__antithesis_instrumentation__.Notify(115290)

	res := cmd.replicatedResult()
	if deprecatedDelta := res.DeprecatedDelta; deprecatedDelta != nil {
		__antithesis_instrumentation__.Notify(115291)
		if res.Delta != (enginepb.MVCCStatsDelta{}) {
			__antithesis_instrumentation__.Notify(115293)
			log.Fatalf(ctx, "stats delta not empty but deprecated delta provided: %+v", cmd)
		} else {
			__antithesis_instrumentation__.Notify(115294)
		}
		__antithesis_instrumentation__.Notify(115292)
		res.Delta = deprecatedDelta.ToStatsDelta()
		res.DeprecatedDelta = nil
	} else {
		__antithesis_instrumentation__.Notify(115295)
	}
}

func (b *replicaAppBatch) stageWriteBatch(ctx context.Context, cmd *replicatedCmd) error {
	__antithesis_instrumentation__.Notify(115296)
	wb := cmd.raftCmd.WriteBatch
	if wb == nil {
		__antithesis_instrumentation__.Notify(115300)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115301)
	}
	__antithesis_instrumentation__.Notify(115297)
	if mutations, err := storage.RocksDBBatchCount(wb.Data); err != nil {
		__antithesis_instrumentation__.Notify(115302)
		log.Errorf(ctx, "unable to read header of committed WriteBatch: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(115303)
		b.mutations += mutations
	}
	__antithesis_instrumentation__.Notify(115298)
	if err := b.batch.ApplyBatchRepr(wb.Data, false); err != nil {
		__antithesis_instrumentation__.Notify(115304)
		return wrapWithNonDeterministicFailure(err, "unable to apply WriteBatch")
	} else {
		__antithesis_instrumentation__.Notify(115305)
	}
	__antithesis_instrumentation__.Notify(115299)
	return nil
}

func changeRemovesStore(
	desc *roachpb.RangeDescriptor, change *kvserverpb.ChangeReplicas, storeID roachpb.StoreID,
) (removesStore bool) {
	__antithesis_instrumentation__.Notify(115306)

	_, existsInChange := change.Desc.GetReplicaDescriptor(storeID)
	return !existsInChange
}

func (b *replicaAppBatch) runPreApplyTriggersBeforeStagingWriteBatch(
	ctx context.Context, cmd *replicatedCmd,
) error {
	__antithesis_instrumentation__.Notify(115307)
	if ops := cmd.raftCmd.LogicalOpLog; ops != nil {
		__antithesis_instrumentation__.Notify(115309)
		b.r.populatePrevValsInLogicalOpLogRaftMuLocked(ctx, ops, b.batch)
	} else {
		__antithesis_instrumentation__.Notify(115310)
	}
	__antithesis_instrumentation__.Notify(115308)
	return nil
}

func (b *replicaAppBatch) runPreApplyTriggersAfterStagingWriteBatch(
	ctx context.Context, cmd *replicatedCmd,
) error {
	__antithesis_instrumentation__.Notify(115311)
	res := cmd.replicatedResult()

	if res.MVCCHistoryMutation != nil {
		__antithesis_instrumentation__.Notify(115319)
		for _, span := range res.MVCCHistoryMutation.Spans {
			__antithesis_instrumentation__.Notify(115320)
			b.r.disconnectRangefeedSpanWithErr(span, roachpb.NewError(&roachpb.MVCCHistoryMutationError{
				Span: span,
			}))
		}
	} else {
		__antithesis_instrumentation__.Notify(115321)
	}
	__antithesis_instrumentation__.Notify(115312)

	if res.AddSSTable != nil {
		__antithesis_instrumentation__.Notify(115322)
		copied := addSSTablePreApply(
			ctx,
			b.r.store.cfg.Settings,
			b.r.store.engine,
			b.r.raftMu.sideloaded,
			cmd.ent.Term,
			cmd.ent.Index,
			*res.AddSSTable,
			b.r.store.limiters.BulkIOWriteRate,
		)
		b.r.store.metrics.AddSSTableApplications.Inc(1)
		if copied {
			__antithesis_instrumentation__.Notify(115326)
			b.r.store.metrics.AddSSTableApplicationCopies.Inc(1)
		} else {
			__antithesis_instrumentation__.Notify(115327)
		}
		__antithesis_instrumentation__.Notify(115323)
		if added := res.Delta.KeyCount; added > 0 {
			__antithesis_instrumentation__.Notify(115328)
			b.r.writeStats.recordCount(float64(added), 0)
		} else {
			__antithesis_instrumentation__.Notify(115329)
		}
		__antithesis_instrumentation__.Notify(115324)
		if res.AddSSTable.AtWriteTimestamp {
			__antithesis_instrumentation__.Notify(115330)
			b.r.handleSSTableRaftMuLocked(
				ctx, res.AddSSTable.Data, res.AddSSTable.Span, res.WriteTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(115331)
		}
		__antithesis_instrumentation__.Notify(115325)
		res.AddSSTable = nil
	} else {
		__antithesis_instrumentation__.Notify(115332)
	}
	__antithesis_instrumentation__.Notify(115313)

	if res.Split != nil {
		__antithesis_instrumentation__.Notify(115333)

		splitPreApply(ctx, b.r, b.batch, res.Split.SplitTrigger, cmd.raftCmd.ClosedTimestamp)

		b.r.disconnectRangefeedWithReason(
			roachpb.RangeFeedRetryError_REASON_RANGE_SPLIT,
		)
	} else {
		__antithesis_instrumentation__.Notify(115334)
	}
	__antithesis_instrumentation__.Notify(115314)

	if merge := res.Merge; merge != nil {
		__antithesis_instrumentation__.Notify(115335)

		rhsRepl, err := b.r.store.GetReplica(merge.RightDesc.RangeID)
		if err != nil {
			__antithesis_instrumentation__.Notify(115338)
			return wrapWithNonDeterministicFailure(err, "unable to get replica for merge")
		} else {
			__antithesis_instrumentation__.Notify(115339)
		}
		__antithesis_instrumentation__.Notify(115336)

		rhsRepl.raftMu.AssertHeld()

		rhsRepl.readOnlyCmdMu.Lock()
		rhsRepl.mu.Lock()
		rhsRepl.mu.destroyStatus.Set(
			roachpb.NewRangeNotFoundError(rhsRepl.RangeID, rhsRepl.store.StoreID()),
			destroyReasonRemoved)
		rhsRepl.mu.Unlock()
		rhsRepl.readOnlyCmdMu.Unlock()

		const clearRangeIDLocalOnly = true
		const mustClearRange = false
		if err := rhsRepl.preDestroyRaftMuLocked(
			ctx, b.batch, b.batch, mergedTombstoneReplicaID, clearRangeIDLocalOnly, mustClearRange,
		); err != nil {
			__antithesis_instrumentation__.Notify(115340)
			return wrapWithNonDeterministicFailure(err, "unable to destroy replica before merge")
		} else {
			__antithesis_instrumentation__.Notify(115341)
		}
		__antithesis_instrumentation__.Notify(115337)

		b.r.disconnectRangefeedWithReason(
			roachpb.RangeFeedRetryError_REASON_RANGE_MERGED,
		)
		rhsRepl.disconnectRangefeedWithReason(
			roachpb.RangeFeedRetryError_REASON_RANGE_MERGED,
		)
	} else {
		__antithesis_instrumentation__.Notify(115342)
	}
	__antithesis_instrumentation__.Notify(115315)

	if res.State != nil && func() bool {
		__antithesis_instrumentation__.Notify(115343)
		return res.State.TruncatedState != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(115344)
		var err error

		looselyCoupledTruncation := isLooselyCoupledRaftLogTruncationEnabled(ctx, b.r.ClusterSettings())

		apply := !looselyCoupledTruncation || func() bool {
			__antithesis_instrumentation__.Notify(115346)
			return res.RaftExpectedFirstIndex == 0 == true
		}() == true
		if apply {
			__antithesis_instrumentation__.Notify(115347)
			if apply, err = handleTruncatedStateBelowRaftPreApply(
				ctx, b.state.TruncatedState, res.State.TruncatedState, b.r.raftMu.stateLoader, b.batch,
			); err != nil {
				__antithesis_instrumentation__.Notify(115348)
				return wrapWithNonDeterministicFailure(err, "unable to handle truncated state")
			} else {
				__antithesis_instrumentation__.Notify(115349)
			}
		} else {
			__antithesis_instrumentation__.Notify(115350)
			b.r.store.raftTruncator.addPendingTruncation(
				ctx, (*raftTruncatorReplica)(b.r), *res.State.TruncatedState, res.RaftExpectedFirstIndex,
				res.RaftLogDelta)
		}
		__antithesis_instrumentation__.Notify(115345)
		if !apply {
			__antithesis_instrumentation__.Notify(115351)

			res.State.TruncatedState = nil
			res.RaftLogDelta = 0
			res.RaftExpectedFirstIndex = 0
			if !looselyCoupledTruncation {
				__antithesis_instrumentation__.Notify(115352)

				b.r.mu.Lock()
				b.r.mu.raftLogSizeTrusted = false
				b.r.mu.Unlock()
			} else {
				__antithesis_instrumentation__.Notify(115353)
			}
		} else {
			__antithesis_instrumentation__.Notify(115354)
		}
	} else {
		__antithesis_instrumentation__.Notify(115355)
	}
	__antithesis_instrumentation__.Notify(115316)

	if change := res.ChangeReplicas; change != nil && func() bool {
		__antithesis_instrumentation__.Notify(115356)
		return changeRemovesStore(b.state.Desc, change, b.r.store.StoreID()) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(115357)
		return !b.r.store.TestingKnobs().DisableEagerReplicaRemoval == true
	}() == true {
		__antithesis_instrumentation__.Notify(115358)

		b.r.readOnlyCmdMu.Lock()
		b.r.mu.Lock()
		b.r.mu.destroyStatus.Set(
			roachpb.NewRangeNotFoundError(b.r.RangeID, b.r.store.StoreID()),
			destroyReasonRemoved)
		b.r.mu.Unlock()
		b.r.readOnlyCmdMu.Unlock()
		b.changeRemovesReplica = true

		if err := b.r.preDestroyRaftMuLocked(
			ctx,
			b.batch,
			b.batch,
			change.NextReplicaID(),
			false,
			false,
		); err != nil {
			__antithesis_instrumentation__.Notify(115359)
			return wrapWithNonDeterministicFailure(err, "unable to destroy replica before removal")
		} else {
			__antithesis_instrumentation__.Notify(115360)
		}
	} else {
		__antithesis_instrumentation__.Notify(115361)
	}
	__antithesis_instrumentation__.Notify(115317)

	if ops := cmd.raftCmd.LogicalOpLog; cmd.raftCmd.WriteBatch != nil {
		__antithesis_instrumentation__.Notify(115362)
		b.r.handleLogicalOpLogRaftMuLocked(ctx, ops, b.batch)
	} else {
		__antithesis_instrumentation__.Notify(115363)
		if ops != nil {
			__antithesis_instrumentation__.Notify(115364)
			log.Fatalf(ctx, "non-nil logical op log with nil write batch: %v", cmd.raftCmd)
		} else {
			__antithesis_instrumentation__.Notify(115365)
		}
	}
	__antithesis_instrumentation__.Notify(115318)

	return nil
}

func (b *replicaAppBatch) stageTrivialReplicatedEvalResult(
	ctx context.Context, cmd *replicatedCmd,
) {
	__antithesis_instrumentation__.Notify(115366)
	raftAppliedIndex := cmd.ent.Index
	if raftAppliedIndex == 0 {
		__antithesis_instrumentation__.Notify(115371)
		log.Fatalf(ctx, "raft entry with index 0")
	} else {
		__antithesis_instrumentation__.Notify(115372)
	}
	__antithesis_instrumentation__.Notify(115367)
	b.state.RaftAppliedIndex = raftAppliedIndex
	rs := cmd.decodedRaftEntry.replicatedResult().State

	if b.state.RaftAppliedIndexTerm > 0 || func() bool {
		__antithesis_instrumentation__.Notify(115373)
		return (rs != nil && func() bool {
			__antithesis_instrumentation__.Notify(115374)
			return rs.RaftAppliedIndexTerm == stateloader.RaftLogTermSignalForAddRaftAppliedIndexTermMigration == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115375)

		b.state.RaftAppliedIndexTerm = cmd.ent.Term
	} else {
		__antithesis_instrumentation__.Notify(115376)
	}
	__antithesis_instrumentation__.Notify(115368)

	if leaseAppliedIndex := cmd.leaseIndex; leaseAppliedIndex != 0 {
		__antithesis_instrumentation__.Notify(115377)
		b.state.LeaseAppliedIndex = leaseAppliedIndex
	} else {
		__antithesis_instrumentation__.Notify(115378)
	}
	__antithesis_instrumentation__.Notify(115369)
	if cts := cmd.raftCmd.ClosedTimestamp; cts != nil && func() bool {
		__antithesis_instrumentation__.Notify(115379)
		return !cts.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(115380)
		b.state.RaftClosedTimestamp = *cts
		b.closedTimestampSetter.record(cmd, b.state.Lease)
	} else {
		__antithesis_instrumentation__.Notify(115381)
	}
	__antithesis_instrumentation__.Notify(115370)

	res := cmd.replicatedResult()

	deltaStats := res.Delta.ToStats()
	b.state.Stats.Add(deltaStats)
}

func (b *replicaAppBatch) ApplyToStateMachine(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(115382)
	if log.V(4) {
		__antithesis_instrumentation__.Notify(115392)
		log.Infof(ctx, "flushing batch %v of %d entries", b.state, b.entries)
	} else {
		__antithesis_instrumentation__.Notify(115393)
	}
	__antithesis_instrumentation__.Notify(115383)

	if !b.changeRemovesReplica {
		__antithesis_instrumentation__.Notify(115394)
		if err := b.addAppliedStateKeyToBatch(ctx); err != nil {
			__antithesis_instrumentation__.Notify(115395)
			return err
		} else {
			__antithesis_instrumentation__.Notify(115396)
		}
	} else {
		__antithesis_instrumentation__.Notify(115397)
	}
	__antithesis_instrumentation__.Notify(115384)

	sync := b.changeRemovesReplica
	if err := b.batch.Commit(sync); err != nil {
		__antithesis_instrumentation__.Notify(115398)
		return wrapWithNonDeterministicFailure(err, "unable to commit Raft entry batch")
	} else {
		__antithesis_instrumentation__.Notify(115399)
	}
	__antithesis_instrumentation__.Notify(115385)
	b.batch.Close()
	b.batch = nil

	r := b.r
	r.mu.Lock()
	r.mu.state.RaftAppliedIndex = b.state.RaftAppliedIndex

	r.mu.state.RaftAppliedIndexTerm = b.state.RaftAppliedIndexTerm
	r.mu.state.LeaseAppliedIndex = b.state.LeaseAppliedIndex

	existingClosed := r.mu.state.RaftClosedTimestamp
	newClosed := b.state.RaftClosedTimestamp
	if !newClosed.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(115400)
		return newClosed.Less(existingClosed) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(115401)
		return raftClosedTimestampAssertionsEnabled == true
	}() == true {
		__antithesis_instrumentation__.Notify(115402)
		return errors.AssertionFailedf(
			"raft closed timestamp regression; replica has: %s, new batch has: %s.",
			existingClosed.String(), newClosed.String())
	} else {
		__antithesis_instrumentation__.Notify(115403)
	}
	__antithesis_instrumentation__.Notify(115386)
	r.mu.closedTimestampSetter = b.closedTimestampSetter

	closedTimestampUpdated := r.mu.state.RaftClosedTimestamp.Forward(b.state.RaftClosedTimestamp)
	prevStats := *r.mu.state.Stats
	*r.mu.state.Stats = *b.state.Stats

	if r.mu.largestPreviousMaxRangeSizeBytes > 0 && func() bool {
		__antithesis_instrumentation__.Notify(115404)
		return b.state.Stats.Total() < r.mu.conf.RangeMaxBytes == true
	}() == true {
		__antithesis_instrumentation__.Notify(115405)
		r.mu.largestPreviousMaxRangeSizeBytes = 0
	} else {
		__antithesis_instrumentation__.Notify(115406)
	}
	__antithesis_instrumentation__.Notify(115387)

	needsSplitBySize := r.needsSplitBySizeRLocked()
	needsMergeBySize := r.needsMergeBySizeRLocked()
	needsTruncationByLogSize := r.needsRaftLogTruncationLocked()
	r.mu.Unlock()
	if closedTimestampUpdated {
		__antithesis_instrumentation__.Notify(115407)
		r.handleClosedTimestampUpdateRaftMuLocked(ctx, b.state.RaftClosedTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(115408)
	}
	__antithesis_instrumentation__.Notify(115388)

	deltaStats := *b.state.Stats
	deltaStats.Subtract(prevStats)
	r.store.metrics.addMVCCStats(ctx, r.tenantMetricsRef, deltaStats)

	b.r.writeStats.recordCount(float64(b.mutations), 0)

	now := timeutil.Now()
	if needsSplitBySize && func() bool {
		__antithesis_instrumentation__.Notify(115409)
		return r.splitQueueThrottle.ShouldProcess(now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115410)
		r.store.splitQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(115411)
	}
	__antithesis_instrumentation__.Notify(115389)
	if needsMergeBySize && func() bool {
		__antithesis_instrumentation__.Notify(115412)
		return r.mergeQueueThrottle.ShouldProcess(now) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115413)
		r.store.mergeQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(115414)
	}
	__antithesis_instrumentation__.Notify(115390)
	if needsTruncationByLogSize {
		__antithesis_instrumentation__.Notify(115415)
		r.store.raftLogQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(115416)
	}
	__antithesis_instrumentation__.Notify(115391)

	b.recordStatsOnCommit()
	return nil
}

func (b *replicaAppBatch) addAppliedStateKeyToBatch(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(115417)

	loader := &b.r.raftMu.stateLoader
	return loader.SetRangeAppliedState(
		ctx, b.batch, b.state.RaftAppliedIndex, b.state.LeaseAppliedIndex, b.state.RaftAppliedIndexTerm,
		b.state.Stats, &b.state.RaftClosedTimestamp,
	)
}

func (b *replicaAppBatch) recordStatsOnCommit() {
	__antithesis_instrumentation__.Notify(115418)
	b.sm.stats.entriesProcessed += b.entries
	b.sm.stats.numEmptyEntries += b.emptyEntries
	b.sm.stats.batchesProcessed++

	elapsed := timeutil.Since(b.start)
	b.r.store.metrics.RaftCommandCommitLatency.RecordValue(elapsed.Nanoseconds())
}

func (b *replicaAppBatch) Close() {
	__antithesis_instrumentation__.Notify(115419)
	if b.batch != nil {
		__antithesis_instrumentation__.Notify(115421)
		b.batch.Close()
	} else {
		__antithesis_instrumentation__.Notify(115422)
	}
	__antithesis_instrumentation__.Notify(115420)
	*b = replicaAppBatch{}
}

var raftClosedTimestampAssertionsEnabled = envutil.EnvOrDefaultBool("COCKROACH_RAFT_CLOSEDTS_ASSERTIONS_ENABLED", true)

func (b *replicaAppBatch) assertNoWriteBelowClosedTimestamp(cmd *replicatedCmd) error {
	__antithesis_instrumentation__.Notify(115423)
	if !cmd.IsLocal() || func() bool {
		__antithesis_instrumentation__.Notify(115427)
		return !cmd.proposal.Request.AppliesTimestampCache() == true
	}() == true {
		__antithesis_instrumentation__.Notify(115428)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115429)
	}
	__antithesis_instrumentation__.Notify(115424)
	if !raftClosedTimestampAssertionsEnabled {
		__antithesis_instrumentation__.Notify(115430)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115431)
	}
	__antithesis_instrumentation__.Notify(115425)
	wts := cmd.raftCmd.ReplicatedEvalResult.WriteTimestamp
	if !wts.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(115432)
		return wts.LessEq(b.state.RaftClosedTimestamp) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115433)
		wts := wts
		var req redact.StringBuilder
		if cmd.proposal != nil {
			__antithesis_instrumentation__.Notify(115435)
			req.Print(cmd.proposal.Request)
		} else {
			__antithesis_instrumentation__.Notify(115436)
			req.SafeString("request unknown; not leaseholder")
		}
		__antithesis_instrumentation__.Notify(115434)
		return wrapWithNonDeterministicFailure(errors.AssertionFailedf(
			"command writing below closed timestamp; cmd: %x, write ts: %s, "+
				"batch state closed: %s, command closed: %s, request: %s, lease: %s.\n"+
				"This assertion will fire again on restart; to ignore run with env var\n"+
				"COCKROACH_RAFT_CLOSEDTS_ASSERTIONS_ENABLED=false",
			cmd.idKey, wts,
			b.state.RaftClosedTimestamp, cmd.raftCmd.ClosedTimestamp,
			req, b.state.Lease),
			"command writing below closed timestamp")
	} else {
		__antithesis_instrumentation__.Notify(115437)
	}
	__antithesis_instrumentation__.Notify(115426)
	return nil
}

func (b *replicaAppBatch) assertNoCmdClosedTimestampRegression(
	ctx context.Context, cmd *replicatedCmd,
) error {
	__antithesis_instrumentation__.Notify(115438)
	if !raftClosedTimestampAssertionsEnabled {
		__antithesis_instrumentation__.Notify(115441)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(115442)
	}
	__antithesis_instrumentation__.Notify(115439)
	existingClosed := &b.state.RaftClosedTimestamp
	newClosed := cmd.raftCmd.ClosedTimestamp
	if newClosed != nil && func() bool {
		__antithesis_instrumentation__.Notify(115443)
		return !newClosed.IsEmpty() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(115444)
		return newClosed.Less(*existingClosed) == true
	}() == true {
		__antithesis_instrumentation__.Notify(115445)
		var req redact.StringBuilder
		if cmd.IsLocal() {
			__antithesis_instrumentation__.Notify(115449)
			req.Print(cmd.proposal.Request)
		} else {
			__antithesis_instrumentation__.Notify(115450)
			req.SafeString("<unknown; not leaseholder>")
		}
		__antithesis_instrumentation__.Notify(115446)
		var prevReq redact.StringBuilder
		if req := b.closedTimestampSetter.leaseReq; req != nil {
			__antithesis_instrumentation__.Notify(115451)
			prevReq.Printf("lease acquisition: %s (prev: %s)", req.Lease, req.PrevLease)
		} else {
			__antithesis_instrumentation__.Notify(115452)
			prevReq.SafeString("<unknown; not leaseholder or not lease request>")
		}
		__antithesis_instrumentation__.Notify(115447)

		logTail, err := b.r.printRaftTail(ctx, 100, 2000)
		if err != nil {
			__antithesis_instrumentation__.Notify(115453)
			if logTail != "" {
				__antithesis_instrumentation__.Notify(115454)
				logTail = logTail + "\n; error printing log: " + err.Error()
			} else {
				__antithesis_instrumentation__.Notify(115455)
				logTail = "error printing log: " + err.Error()
			}
		} else {
			__antithesis_instrumentation__.Notify(115456)
		}
		__antithesis_instrumentation__.Notify(115448)

		return errors.AssertionFailedf(
			"raft closed timestamp regression in cmd: %x (term: %d, index: %d); batch state: %s, command: %s, lease: %s, req: %s, applying at LAI: %d.\n"+
				"Closed timestamp was set by req: %s under lease: %s; applied at LAI: %d. Batch idx: %d.\n"+
				"This assertion will fire again on restart; to ignore run with env var COCKROACH_RAFT_CLOSEDTS_ASSERTIONS_ENABLED=false\n"+
				"Raft log tail:\n%s",
			cmd.idKey, cmd.ent.Term, cmd.ent.Index, existingClosed, newClosed, b.state.Lease, req, cmd.leaseIndex,
			prevReq, b.closedTimestampSetter.lease, b.closedTimestampSetter.leaseIdx, b.entries,
			logTail)
	} else {
		__antithesis_instrumentation__.Notify(115457)
	}
	__antithesis_instrumentation__.Notify(115440)
	return nil
}

type ephemeralReplicaAppBatch struct {
	r     *Replica
	state kvserverpb.ReplicaState
}

func (mb *ephemeralReplicaAppBatch) Stage(
	ctx context.Context, cmdI apply.Command,
) (apply.CheckedCommand, error) {
	__antithesis_instrumentation__.Notify(115458)
	cmd := cmdI.(*replicatedCmd)

	mb.r.shouldApplyCommand(ctx, cmd, &mb.state)
	mb.state.LeaseAppliedIndex = cmd.leaseIndex
	return cmd, nil
}

func (mb *ephemeralReplicaAppBatch) ApplyToStateMachine(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(115459)
	panic("cannot apply ephemeralReplicaAppBatch to state machine")
}

func (mb *ephemeralReplicaAppBatch) Close() {
	__antithesis_instrumentation__.Notify(115460)
	*mb = ephemeralReplicaAppBatch{}
}

func (sm *replicaStateMachine) ApplySideEffects(
	ctx context.Context, cmdI apply.CheckedCommand,
) (apply.AppliedCommand, error) {
	__antithesis_instrumentation__.Notify(115461)
	cmd := cmdI.(*replicatedCmd)

	if unlock := cmd.splitMergeUnlock; unlock != nil {
		__antithesis_instrumentation__.Notify(115467)
		defer unlock()
	} else {
		__antithesis_instrumentation__.Notify(115468)
	}
	__antithesis_instrumentation__.Notify(115462)

	sm.r.prepareLocalResult(ctx, cmd)
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(115469)
		log.VEventf(ctx, 2, "%v", cmd.localResult.String())
	} else {
		__antithesis_instrumentation__.Notify(115470)
	}
	__antithesis_instrumentation__.Notify(115463)

	clearTrivialReplicatedEvalResultFields(cmd.replicatedResult())
	if !cmd.IsTrivial() {
		__antithesis_instrumentation__.Notify(115471)
		shouldAssert, isRemoved := sm.handleNonTrivialReplicatedEvalResult(ctx, cmd.replicatedResult())
		if isRemoved {
			__antithesis_instrumentation__.Notify(115473)

			cmd.FinishNonLocal(ctx)
			return nil, apply.ErrRemoved
		} else {
			__antithesis_instrumentation__.Notify(115474)
		}
		__antithesis_instrumentation__.Notify(115472)

		if shouldAssert {
			__antithesis_instrumentation__.Notify(115475)

			sm.r.mu.RLock()
			sm.r.assertStateRaftMuLockedReplicaMuRLocked(ctx, sm.r.store.Engine())
			sm.r.mu.RUnlock()
			sm.stats.stateAssertions++
		} else {
			__antithesis_instrumentation__.Notify(115476)
		}
	} else {
		__antithesis_instrumentation__.Notify(115477)
		if res := cmd.replicatedResult(); !res.IsZero() {
			__antithesis_instrumentation__.Notify(115478)
			log.Fatalf(ctx, "failed to handle all side-effects of ReplicatedEvalResult: %v", res)
		} else {
			__antithesis_instrumentation__.Notify(115479)
		}
	}
	__antithesis_instrumentation__.Notify(115464)

	if err := sm.maybeApplyConfChange(ctx, cmd); err != nil {
		__antithesis_instrumentation__.Notify(115480)
		return nil, wrapWithNonDeterministicFailure(err, "unable to apply conf change")
	} else {
		__antithesis_instrumentation__.Notify(115481)
	}
	__antithesis_instrumentation__.Notify(115465)

	if cmd.IsLocal() {
		__antithesis_instrumentation__.Notify(115482)

		if cmd.localResult != nil {
			__antithesis_instrumentation__.Notify(115487)
			sm.r.handleReadWriteLocalEvalResult(ctx, *cmd.localResult)
		} else {
			__antithesis_instrumentation__.Notify(115488)
		}
		__antithesis_instrumentation__.Notify(115483)

		rejected := cmd.Rejected()
		higherReproposalsExist := cmd.raftCmd.MaxLeaseIndex != cmd.proposal.command.MaxLeaseIndex
		if !rejected && func() bool {
			__antithesis_instrumentation__.Notify(115489)
			return higherReproposalsExist == true
		}() == true {
			__antithesis_instrumentation__.Notify(115490)
			log.Fatalf(ctx, "finishing proposal with outstanding reproposal at a higher max lease index")
		} else {
			__antithesis_instrumentation__.Notify(115491)
		}
		__antithesis_instrumentation__.Notify(115484)
		if !rejected && func() bool {
			__antithesis_instrumentation__.Notify(115492)
			return cmd.proposal.applied == true
		}() == true {
			__antithesis_instrumentation__.Notify(115493)

			log.Fatalf(ctx, "command already applied: %+v; unexpected successful result", cmd)
		} else {
			__antithesis_instrumentation__.Notify(115494)
		}
		__antithesis_instrumentation__.Notify(115485)

		if higherReproposalsExist {
			__antithesis_instrumentation__.Notify(115495)
			sm.r.mu.Lock()
			delete(sm.r.mu.proposals, cmd.idKey)
			sm.r.mu.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(115496)
		}
		__antithesis_instrumentation__.Notify(115486)
		cmd.proposal.applied = true
	} else {
		__antithesis_instrumentation__.Notify(115497)
	}
	__antithesis_instrumentation__.Notify(115466)
	return cmd, nil
}

func (sm *replicaStateMachine) handleNonTrivialReplicatedEvalResult(
	ctx context.Context, rResult *kvserverpb.ReplicatedEvalResult,
) (shouldAssert, isRemoved bool) {
	__antithesis_instrumentation__.Notify(115498)

	if rResult.IsZero() {
		__antithesis_instrumentation__.Notify(115509)
		log.Fatalf(ctx, "zero-value ReplicatedEvalResult passed to handleNonTrivialReplicatedEvalResult")
	} else {
		__antithesis_instrumentation__.Notify(115510)
	}
	__antithesis_instrumentation__.Notify(115499)

	isRaftLogTruncationDeltaTrusted := true
	if rResult.State != nil {
		__antithesis_instrumentation__.Notify(115511)
		if newLease := rResult.State.Lease; newLease != nil {
			__antithesis_instrumentation__.Notify(115516)
			sm.r.handleLeaseResult(ctx, newLease, rResult.PriorReadSummary)
			rResult.State.Lease = nil
			rResult.PriorReadSummary = nil
		} else {
			__antithesis_instrumentation__.Notify(115517)
		}
		__antithesis_instrumentation__.Notify(115512)

		if newTruncState := rResult.State.TruncatedState; newTruncState != nil {
			__antithesis_instrumentation__.Notify(115518)
			raftLogDelta, expectedFirstIndexWasAccurate := sm.r.handleTruncatedStateResult(
				ctx, newTruncState, rResult.RaftExpectedFirstIndex)
			if !expectedFirstIndexWasAccurate && func() bool {
				__antithesis_instrumentation__.Notify(115520)
				return rResult.RaftExpectedFirstIndex != 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(115521)
				isRaftLogTruncationDeltaTrusted = false
			} else {
				__antithesis_instrumentation__.Notify(115522)
			}
			__antithesis_instrumentation__.Notify(115519)
			rResult.RaftLogDelta += raftLogDelta
			rResult.State.TruncatedState = nil
			rResult.RaftExpectedFirstIndex = 0
		} else {
			__antithesis_instrumentation__.Notify(115523)
		}
		__antithesis_instrumentation__.Notify(115513)

		if newThresh := rResult.State.GCThreshold; newThresh != nil {
			__antithesis_instrumentation__.Notify(115524)
			sm.r.handleGCThresholdResult(ctx, newThresh)
			rResult.State.GCThreshold = nil
		} else {
			__antithesis_instrumentation__.Notify(115525)
		}
		__antithesis_instrumentation__.Notify(115514)

		if newVersion := rResult.State.Version; newVersion != nil {
			__antithesis_instrumentation__.Notify(115526)
			sm.r.handleVersionResult(ctx, newVersion)
			rResult.State.Version = nil
		} else {
			__antithesis_instrumentation__.Notify(115527)
		}
		__antithesis_instrumentation__.Notify(115515)
		if (*rResult.State == kvserverpb.ReplicaState{}) {
			__antithesis_instrumentation__.Notify(115528)
			rResult.State = nil
		} else {
			__antithesis_instrumentation__.Notify(115529)
		}
	} else {
		__antithesis_instrumentation__.Notify(115530)
	}
	__antithesis_instrumentation__.Notify(115500)

	if rResult.RaftLogDelta != 0 {
		__antithesis_instrumentation__.Notify(115531)

		sm.r.handleRaftLogDeltaResult(ctx, rResult.RaftLogDelta, isRaftLogTruncationDeltaTrusted)
		rResult.RaftLogDelta = 0
	} else {
		__antithesis_instrumentation__.Notify(115532)
	}
	__antithesis_instrumentation__.Notify(115501)

	shouldAssert = !rResult.IsZero()
	if !shouldAssert {
		__antithesis_instrumentation__.Notify(115533)
		return false, false
	} else {
		__antithesis_instrumentation__.Notify(115534)
	}
	__antithesis_instrumentation__.Notify(115502)

	if rResult.Split != nil {
		__antithesis_instrumentation__.Notify(115535)
		sm.r.handleSplitResult(ctx, rResult.Split)
		rResult.Split = nil
	} else {
		__antithesis_instrumentation__.Notify(115536)
	}
	__antithesis_instrumentation__.Notify(115503)

	if rResult.Merge != nil {
		__antithesis_instrumentation__.Notify(115537)
		sm.r.handleMergeResult(ctx, rResult.Merge)
		rResult.Merge = nil
	} else {
		__antithesis_instrumentation__.Notify(115538)
	}
	__antithesis_instrumentation__.Notify(115504)

	if rResult.State != nil {
		__antithesis_instrumentation__.Notify(115539)
		if newDesc := rResult.State.Desc; newDesc != nil {
			__antithesis_instrumentation__.Notify(115541)
			sm.r.handleDescResult(ctx, newDesc)
			rResult.State.Desc = nil
		} else {
			__antithesis_instrumentation__.Notify(115542)
		}
		__antithesis_instrumentation__.Notify(115540)

		if (*rResult.State == kvserverpb.ReplicaState{}) {
			__antithesis_instrumentation__.Notify(115543)
			rResult.State = nil
		} else {
			__antithesis_instrumentation__.Notify(115544)
		}
	} else {
		__antithesis_instrumentation__.Notify(115545)
	}
	__antithesis_instrumentation__.Notify(115505)

	if rResult.ChangeReplicas != nil {
		__antithesis_instrumentation__.Notify(115546)
		isRemoved = sm.r.handleChangeReplicasResult(ctx, rResult.ChangeReplicas)
		rResult.ChangeReplicas = nil
	} else {
		__antithesis_instrumentation__.Notify(115547)
	}
	__antithesis_instrumentation__.Notify(115506)

	if rResult.ComputeChecksum != nil {
		__antithesis_instrumentation__.Notify(115548)
		sm.r.handleComputeChecksumResult(ctx, rResult.ComputeChecksum)
		rResult.ComputeChecksum = nil
	} else {
		__antithesis_instrumentation__.Notify(115549)
	}
	__antithesis_instrumentation__.Notify(115507)

	if !rResult.IsZero() {
		__antithesis_instrumentation__.Notify(115550)
		log.Fatalf(ctx, "unhandled field in ReplicatedEvalResult: %s", pretty.Diff(rResult, &kvserverpb.ReplicatedEvalResult{}))
	} else {
		__antithesis_instrumentation__.Notify(115551)
	}
	__antithesis_instrumentation__.Notify(115508)
	return true, isRemoved
}

func (sm *replicaStateMachine) maybeApplyConfChange(ctx context.Context, cmd *replicatedCmd) error {
	__antithesis_instrumentation__.Notify(115552)
	switch cmd.ent.Type {
	case raftpb.EntryNormal:
		__antithesis_instrumentation__.Notify(115553)
		return nil
	case raftpb.EntryConfChange, raftpb.EntryConfChangeV2:
		__antithesis_instrumentation__.Notify(115554)
		sm.stats.numConfChangeEntries++
		if cmd.Rejected() {
			__antithesis_instrumentation__.Notify(115557)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(115558)
		}
		__antithesis_instrumentation__.Notify(115555)
		return sm.r.withRaftGroup(true, func(rn *raft.RawNode) (bool, error) {
			__antithesis_instrumentation__.Notify(115559)

			rn.ApplyConfChange(cmd.confChange.ConfChangeI)
			return true, nil
		})
	default:
		__antithesis_instrumentation__.Notify(115556)
		panic("unexpected")
	}
}

func (sm *replicaStateMachine) moveStats() applyCommittedEntriesStats {
	__antithesis_instrumentation__.Notify(115560)
	stats := sm.stats
	sm.stats = applyCommittedEntriesStats{}
	return stats
}

type closedTimestampSetterInfo struct {
	lease *roachpb.Lease

	leaseIdx ctpb.LAI

	leaseReq *roachpb.RequestLeaseRequest

	split, merge bool
}

func (s *closedTimestampSetterInfo) record(cmd *replicatedCmd, lease *roachpb.Lease) {
	__antithesis_instrumentation__.Notify(115561)
	*s = closedTimestampSetterInfo{}
	s.leaseIdx = ctpb.LAI(cmd.leaseIndex)
	s.lease = lease
	if !cmd.IsLocal() {
		__antithesis_instrumentation__.Notify(115563)
		return
	} else {
		__antithesis_instrumentation__.Notify(115564)
	}
	__antithesis_instrumentation__.Notify(115562)
	req := cmd.proposal.Request
	et, ok := req.GetArg(roachpb.EndTxn)
	if ok {
		__antithesis_instrumentation__.Notify(115565)
		endTxn := et.(*roachpb.EndTxnRequest)
		if trig := endTxn.InternalCommitTrigger; trig != nil {
			__antithesis_instrumentation__.Notify(115566)
			if trig.SplitTrigger != nil {
				__antithesis_instrumentation__.Notify(115567)
				s.split = true
			} else {
				__antithesis_instrumentation__.Notify(115568)
				if trig.MergeTrigger != nil {
					__antithesis_instrumentation__.Notify(115569)
					s.merge = true
				} else {
					__antithesis_instrumentation__.Notify(115570)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(115571)
		}
	} else {
		__antithesis_instrumentation__.Notify(115572)
		if req.IsLeaseRequest() {
			__antithesis_instrumentation__.Notify(115573)

			lr, _ := req.GetArg(roachpb.RequestLease)
			s.leaseReq = protoutil.Clone(lr).(*roachpb.RequestLeaseRequest)
		} else {
			__antithesis_instrumentation__.Notify(115574)
		}
	}
}
