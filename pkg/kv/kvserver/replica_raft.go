package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"sort"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/apply"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/poison"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/uncertainty"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
	"go.etcd.io/etcd/raft/v3/tracker"
)

var (
	raftLogTruncationClearRangeThreshold = uint64(util.ConstantWithMetamorphicTestRange(
		"raft-log-truncation-clearrange-threshold", 100000, 1, 1e6))
)

func makeIDKey() kvserverbase.CmdIDKey {
	__antithesis_instrumentation__.Notify(118372)
	idKeyBuf := make([]byte, 0, kvserverbase.RaftCommandIDLen)
	idKeyBuf = encoding.EncodeUint64Ascending(idKeyBuf, uint64(rand.Int63()))
	return kvserverbase.CmdIDKey(idKeyBuf)
}

func (r *Replica) evalAndPropose(
	ctx context.Context,
	ba *roachpb.BatchRequest,
	g *concurrency.Guard,
	st kvserverpb.LeaseStatus,
	ui uncertainty.Interval,
	tok TrackedRequestToken,
) (chan proposalResult, func(), kvserverbase.CmdIDKey, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(118373)
	defer tok.DoneIfNotMoved(ctx)
	idKey := makeIDKey()
	proposal, pErr := r.requestToProposal(ctx, idKey, ba, st, ui, g)
	log.Event(proposal.ctx, "evaluated request")

	if isConcurrencyRetryError(pErr) {
		__antithesis_instrumentation__.Notify(118384)
		pErr = maybeAttachLease(pErr, &st.Lease)
		return nil, nil, "", pErr
	} else {
		__antithesis_instrumentation__.Notify(118385)
		if _, ok := pErr.GetDetail().(*roachpb.ReplicaCorruptionError); ok {
			__antithesis_instrumentation__.Notify(118386)
			return nil, nil, "", pErr
		} else {
			__antithesis_instrumentation__.Notify(118387)
		}
	}
	__antithesis_instrumentation__.Notify(118374)

	proposal.ec = endCmds{repl: r, g: g, st: st}

	proposalCh := proposal.doneCh

	if proposal.command == nil {
		__antithesis_instrumentation__.Notify(118388)
		if proposal.Local.RequiresRaft() {
			__antithesis_instrumentation__.Notify(118390)
			return nil, nil, "", roachpb.NewError(errors.AssertionFailedf(
				"proposal resulting from batch %s erroneously bypassed Raft", ba))
		} else {
			__antithesis_instrumentation__.Notify(118391)
		}
		__antithesis_instrumentation__.Notify(118389)
		intents := proposal.Local.DetachEncounteredIntents()
		endTxns := proposal.Local.DetachEndTxns(pErr != nil)
		r.handleReadWriteLocalEvalResult(ctx, *proposal.Local)

		pr := proposalResult{
			Reply:              proposal.Local.Reply,
			Err:                pErr,
			EncounteredIntents: intents,
			EndTxns:            endTxns,
		}
		proposal.finishApplication(ctx, pr)
		return proposalCh, func() { __antithesis_instrumentation__.Notify(118392) }, "", nil
	} else {
		__antithesis_instrumentation__.Notify(118393)
	}
	__antithesis_instrumentation__.Notify(118375)

	log.VEventf(proposal.ctx, 2,
		"proposing command to write %d new keys, %d new values, %d new intents, "+
			"write batch size=%d bytes",
		proposal.command.ReplicatedEvalResult.Delta.KeyCount,
		proposal.command.ReplicatedEvalResult.Delta.ValCount,
		proposal.command.ReplicatedEvalResult.Delta.IntentCount,
		proposal.command.WriteBatch.Size(),
	)

	if ba.AsyncConsensus {
		__antithesis_instrumentation__.Notify(118394)
		if ets := proposal.Local.DetachEndTxns(false); len(ets) != 0 {
			__antithesis_instrumentation__.Notify(118396)

			return nil, nil, "", roachpb.NewErrorf("cannot perform consensus asynchronously for "+
				"proposal with EndTxnIntents=%v; %v", ets, ba)
		} else {
			__antithesis_instrumentation__.Notify(118397)
		}
		__antithesis_instrumentation__.Notify(118395)

		proposal.ctx, proposal.sp = tracing.ForkSpan(ctx, "async consensus")

		reply := *proposal.Local.Reply
		reply.Responses = append([]roachpb.ResponseUnion(nil), reply.Responses...)
		pr := proposalResult{
			Reply:              &reply,
			EncounteredIntents: proposal.Local.DetachEncounteredIntents(),
		}
		proposal.signalProposalResult(pr)

	} else {
		__antithesis_instrumentation__.Notify(118398)
	}
	__antithesis_instrumentation__.Notify(118376)

	if ba.IsSingleSkipsLeaseCheckRequest() {
		__antithesis_instrumentation__.Notify(118399)

		var seq roachpb.LeaseSequence
		switch t := ba.Requests[0].GetInner().(type) {
		case *roachpb.RequestLeaseRequest:
			__antithesis_instrumentation__.Notify(118401)
			seq = t.PrevLease.Sequence
		case *roachpb.TransferLeaseRequest:
			__antithesis_instrumentation__.Notify(118402)
			seq = t.PrevLease.Sequence
		default:
			__antithesis_instrumentation__.Notify(118403)
		}
		__antithesis_instrumentation__.Notify(118400)
		proposal.command.ProposerLeaseSequence = seq
	} else {
		__antithesis_instrumentation__.Notify(118404)
		if !st.Lease.OwnedBy(r.store.StoreID()) {
			__antithesis_instrumentation__.Notify(118405)

			log.Fatalf(ctx, "cannot propose %s on follower with remotely owned lease %s", ba, st.Lease)
		} else {
			__antithesis_instrumentation__.Notify(118406)
			proposal.command.ProposerLeaseSequence = st.Lease.Sequence
		}
	}
	__antithesis_instrumentation__.Notify(118377)

	quotaSize := uint64(proposal.command.Size())
	if maxSize := uint64(MaxCommandSize.Get(&r.store.cfg.Settings.SV)); quotaSize > maxSize {
		__antithesis_instrumentation__.Notify(118407)
		return nil, nil, "", roachpb.NewError(errors.Errorf(
			"command is too large: %d bytes (max: %d)", quotaSize, maxSize,
		))
	} else {
		__antithesis_instrumentation__.Notify(118408)
	}
	__antithesis_instrumentation__.Notify(118378)
	var err error
	proposal.quotaAlloc, err = r.maybeAcquireProposalQuota(ctx, quotaSize)
	if err != nil {
		__antithesis_instrumentation__.Notify(118409)
		return nil, nil, "", roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(118410)
	}
	__antithesis_instrumentation__.Notify(118379)

	defer func() {
		__antithesis_instrumentation__.Notify(118411)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(118412)
			proposal.releaseQuota()
		} else {
			__antithesis_instrumentation__.Notify(118413)
		}
	}()
	__antithesis_instrumentation__.Notify(118380)

	if filter := r.store.TestingKnobs().TestingProposalFilter; filter != nil {
		__antithesis_instrumentation__.Notify(118414)
		filterArgs := kvserverbase.ProposalFilterArgs{
			Ctx:        ctx,
			Cmd:        *proposal.command,
			QuotaAlloc: proposal.quotaAlloc,
			CmdID:      idKey,
			Req:        *ba,
		}
		if pErr = filter(filterArgs); pErr != nil {
			__antithesis_instrumentation__.Notify(118415)
			return nil, nil, "", pErr
		} else {
			__antithesis_instrumentation__.Notify(118416)
		}
	} else {
		__antithesis_instrumentation__.Notify(118417)
	}
	__antithesis_instrumentation__.Notify(118381)

	pErr = r.propose(ctx, proposal, tok.Move(ctx))
	if pErr != nil {
		__antithesis_instrumentation__.Notify(118418)
		return nil, nil, "", pErr
	} else {
		__antithesis_instrumentation__.Notify(118419)
	}
	__antithesis_instrumentation__.Notify(118382)

	abandon := func() {
		__antithesis_instrumentation__.Notify(118420)

		r.raftMu.Lock()
		defer r.raftMu.Unlock()
		r.mu.Lock()
		defer r.mu.Unlock()

		proposal.ctx = r.AnnotateCtx(context.TODO())
	}
	__antithesis_instrumentation__.Notify(118383)
	return proposalCh, abandon, idKey, nil
}

func (r *Replica) propose(
	ctx context.Context, p *ProposalData, tok TrackedRequestToken,
) (pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(118421)
	defer tok.DoneIfNotMoved(ctx)

	defer func(prev uint64) {
		__antithesis_instrumentation__.Notify(118430)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(118431)
			p.command.MaxLeaseIndex = prev
		} else {
			__antithesis_instrumentation__.Notify(118432)
		}
	}(p.command.MaxLeaseIndex)
	__antithesis_instrumentation__.Notify(118422)

	p.command.MaxLeaseIndex = 0

	prefix := true
	version := kvserverbase.RaftVersionStandard
	if crt := p.command.ReplicatedEvalResult.ChangeReplicas; crt != nil {
		__antithesis_instrumentation__.Notify(118433)

		log.Infof(p.ctx, "proposing %s", crt)
		prefix = false

		replID := r.ReplicaID()
		rDesc, ok := p.command.ReplicatedEvalResult.State.Desc.GetReplicaDescriptorByID(replID)
		hasVoterIncoming := p.command.ReplicatedEvalResult.State.Desc.ContainsVoterIncoming()
		lhRemovalAllowed := hasVoterIncoming && func() bool {
			__antithesis_instrumentation__.Notify(118434)
			return r.store.cfg.Settings.Version.IsActive(ctx,
				clusterversion.EnableLeaseHolderRemoval) == true
		}() == true

		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(118435)
			return (lhRemovalAllowed && func() bool {
				__antithesis_instrumentation__.Notify(118436)
				return !rDesc.IsAnyVoter() == true
			}() == true) == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(118437)
			return (!lhRemovalAllowed && func() bool {
				__antithesis_instrumentation__.Notify(118438)
				return !rDesc.IsVoterNewConfig() == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(118439)
			err := errors.Mark(errors.Newf("received invalid ChangeReplicasTrigger %s to remove "+
				"self (leaseholder); hasVoterIncoming: %v, lhRemovalAllowed: %v; proposed descriptor: %v",
				crt, hasVoterIncoming, lhRemovalAllowed, p.command.ReplicatedEvalResult.State.Desc),
				errMarkInvalidReplicationChange)
			log.Errorf(p.ctx, "%v", err)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(118440)
		}
	} else {
		__antithesis_instrumentation__.Notify(118441)
		if p.command.ReplicatedEvalResult.AddSSTable != nil {
			__antithesis_instrumentation__.Notify(118442)
			log.VEvent(p.ctx, 4, "sideloadable proposal detected")
			version = kvserverbase.RaftVersionSideloaded
			r.store.metrics.AddSSTableProposals.Inc(1)

			if p.command.ReplicatedEvalResult.AddSSTable.Data == nil {
				__antithesis_instrumentation__.Notify(118443)
				return roachpb.NewErrorf("cannot sideload empty SSTable")
			} else {
				__antithesis_instrumentation__.Notify(118444)
			}
		} else {
			__antithesis_instrumentation__.Notify(118445)
			if log.V(4) {
				__antithesis_instrumentation__.Notify(118446)
				log.Infof(p.ctx, "proposing command %x: %s", p.idKey, p.Request.Summary())
			} else {
				__antithesis_instrumentation__.Notify(118447)
			}
		}
	}
	__antithesis_instrumentation__.Notify(118423)

	preLen := 0
	if prefix {
		__antithesis_instrumentation__.Notify(118448)
		preLen = kvserverbase.RaftCommandPrefixLen
	} else {
		__antithesis_instrumentation__.Notify(118449)
	}
	__antithesis_instrumentation__.Notify(118424)
	cmdLen := p.command.Size()

	needed := preLen + cmdLen + kvserverpb.MaxRaftCommandFooterSize()
	data := make([]byte, preLen, needed)

	if prefix {
		__antithesis_instrumentation__.Notify(118450)
		kvserverbase.EncodeRaftCommandPrefix(data, version, p.idKey)
	} else {
		__antithesis_instrumentation__.Notify(118451)
	}
	__antithesis_instrumentation__.Notify(118425)

	data = data[:preLen+cmdLen]
	if _, err := protoutil.MarshalTo(p.command, data[preLen:]); err != nil {
		__antithesis_instrumentation__.Notify(118452)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(118453)
	}
	__antithesis_instrumentation__.Notify(118426)
	p.encodedCommand = data

	if false {
		__antithesis_instrumentation__.Notify(118454)
		log.Infof(p.ctx, `%s: proposal: %d
  RaftCommand.ReplicatedEvalResult:          %d
  RaftCommand.ReplicatedEvalResult.Delta:    %d
  RaftCommand.WriteBatch:                    %d
`, p.Request.Summary(), cmdLen,
			p.command.ReplicatedEvalResult.Size(),
			p.command.ReplicatedEvalResult.Delta.Size(),
			p.command.WriteBatch.Size(),
		)
	} else {
		__antithesis_instrumentation__.Notify(118455)
	}
	__antithesis_instrumentation__.Notify(118427)

	const largeProposalEventThresholdBytes = 2 << 19
	if cmdLen > largeProposalEventThresholdBytes {
		__antithesis_instrumentation__.Notify(118456)
		log.Eventf(p.ctx, "proposal is large: %s", humanizeutil.IBytes(int64(cmdLen)))
	} else {
		__antithesis_instrumentation__.Notify(118457)
	}
	__antithesis_instrumentation__.Notify(118428)

	err := r.mu.proposalBuf.Insert(ctx, p, tok.Move(ctx))
	if err != nil {
		__antithesis_instrumentation__.Notify(118458)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(118459)
	}
	__antithesis_instrumentation__.Notify(118429)
	return nil
}

func (r *Replica) numPendingProposalsRLocked() int {
	__antithesis_instrumentation__.Notify(118460)
	return len(r.mu.proposals) + r.mu.proposalBuf.AllocatedIdx()
}

func (r *Replica) hasPendingProposalsRLocked() bool {
	__antithesis_instrumentation__.Notify(118461)
	return r.numPendingProposalsRLocked() > 0
}

func (r *Replica) hasPendingProposalQuotaRLocked() bool {
	__antithesis_instrumentation__.Notify(118462)
	if r.mu.proposalQuota == nil {
		__antithesis_instrumentation__.Notify(118464)
		return true
	} else {
		__antithesis_instrumentation__.Notify(118465)
	}
	__antithesis_instrumentation__.Notify(118463)
	return !r.mu.proposalQuota.Full()
}

var errRemoved = errors.New("replica removed")

func (r *Replica) stepRaftGroup(req *kvserverpb.RaftMessageRequest) error {
	__antithesis_instrumentation__.Notify(118466)

	return r.withRaftGroup(false, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118467)

		r.maybeUnquiesceWithOptionsLocked(false)
		r.mu.lastUpdateTimes.update(req.FromReplica.ReplicaID, timeutil.Now())
		if req.Message.Type == raftpb.MsgSnap {
			__antithesis_instrumentation__.Notify(118470)

			if term := raftGroup.BasicStatus().Term; term > req.Message.Term {
				__antithesis_instrumentation__.Notify(118471)
				req.Message.Term = term
			} else {
				__antithesis_instrumentation__.Notify(118472)
			}
		} else {
			__antithesis_instrumentation__.Notify(118473)
		}
		__antithesis_instrumentation__.Notify(118468)
		err := raftGroup.Step(req.Message)
		if errors.Is(err, raft.ErrProposalDropped) {
			__antithesis_instrumentation__.Notify(118474)

			err = nil
		} else {
			__antithesis_instrumentation__.Notify(118475)
		}
		__antithesis_instrumentation__.Notify(118469)
		return false, err
	})
}

type handleSnapshotStats struct {
	offered bool
	applied bool
}

type handleRaftReadyStats struct {
	applyCommittedEntriesStats
	snap handleSnapshotStats
}

var noSnap IncomingSnapshot

func (r *Replica) handleRaftReady(
	ctx context.Context, inSnap IncomingSnapshot,
) (handleRaftReadyStats, string, error) {
	__antithesis_instrumentation__.Notify(118476)
	r.raftMu.Lock()
	defer r.raftMu.Unlock()
	return r.handleRaftReadyRaftMuLocked(ctx, inSnap)
}

func (r *Replica) handleRaftReadyRaftMuLocked(
	ctx context.Context, inSnap IncomingSnapshot,
) (handleRaftReadyStats, string, error) {
	__antithesis_instrumentation__.Notify(118477)

	if ctx.Done() != nil {
		__antithesis_instrumentation__.Notify(118499)
		return handleRaftReadyStats{}, "", errors.AssertionFailedf(
			"handleRaftReadyRaftMuLocked cannot be called with a cancellable context")
	} else {
		__antithesis_instrumentation__.Notify(118500)
	}
	__antithesis_instrumentation__.Notify(118478)

	var stats handleRaftReadyStats
	if inSnap.Desc != nil {
		__antithesis_instrumentation__.Notify(118501)
		stats.snap.offered = true
	} else {
		__antithesis_instrumentation__.Notify(118502)
	}
	__antithesis_instrumentation__.Notify(118479)

	var hasReady bool
	var rd raft.Ready
	r.mu.Lock()
	lastIndex := r.mu.lastIndex
	lastTerm := r.mu.lastTerm
	raftLogSize := r.mu.raftLogSize
	leaderID := r.mu.leaderID
	lastLeaderID := leaderID
	err := r.withRaftGroupLocked(true, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118503)
		numFlushed, err := r.mu.proposalBuf.FlushLockedWithRaftGroup(ctx, raftGroup)
		if err != nil {
			__antithesis_instrumentation__.Notify(118506)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(118507)
		}
		__antithesis_instrumentation__.Notify(118504)
		if hasReady = raftGroup.HasReady(); hasReady {
			__antithesis_instrumentation__.Notify(118508)
			rd = raftGroup.Ready()
		} else {
			__antithesis_instrumentation__.Notify(118509)
		}
		__antithesis_instrumentation__.Notify(118505)

		unquiesceAndWakeLeader := hasReady || func() bool {
			__antithesis_instrumentation__.Notify(118510)
			return numFlushed > 0 == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(118511)
			return len(r.mu.proposals) > 0 == true
		}() == true
		return unquiesceAndWakeLeader, nil
	})
	__antithesis_instrumentation__.Notify(118480)
	r.mu.applyingEntries = len(rd.CommittedEntries) > 0
	r.mu.Unlock()
	if errors.Is(err, errRemoved) {
		__antithesis_instrumentation__.Notify(118512)

		return stats, "", nil
	} else {
		__antithesis_instrumentation__.Notify(118513)
		if err != nil {
			__antithesis_instrumentation__.Notify(118514)
			const expl = "while checking raft group for Ready"
			return stats, expl, errors.Wrap(err, expl)
		} else {
			__antithesis_instrumentation__.Notify(118515)
		}
	}
	__antithesis_instrumentation__.Notify(118481)
	if !hasReady {
		__antithesis_instrumentation__.Notify(118516)

		r.updateProposalQuotaRaftMuLocked(ctx, lastLeaderID)
		return stats, "", nil
	} else {
		__antithesis_instrumentation__.Notify(118517)
	}
	__antithesis_instrumentation__.Notify(118482)

	logRaftReady(ctx, rd)

	refreshReason := noReason
	if rd.SoftState != nil && func() bool {
		__antithesis_instrumentation__.Notify(118518)
		return leaderID != roachpb.ReplicaID(rd.SoftState.Lead) == true
	}() == true {
		__antithesis_instrumentation__.Notify(118519)

		if log.V(3) {
			__antithesis_instrumentation__.Notify(118522)
			log.Infof(ctx, "raft leader changed: %d -> %d", leaderID, rd.SoftState.Lead)
		} else {
			__antithesis_instrumentation__.Notify(118523)
		}
		__antithesis_instrumentation__.Notify(118520)
		if !r.store.TestingKnobs().DisableRefreshReasonNewLeader {
			__antithesis_instrumentation__.Notify(118524)
			refreshReason = reasonNewLeader
		} else {
			__antithesis_instrumentation__.Notify(118525)
		}
		__antithesis_instrumentation__.Notify(118521)
		leaderID = roachpb.ReplicaID(rd.SoftState.Lead)
	} else {
		__antithesis_instrumentation__.Notify(118526)
	}
	__antithesis_instrumentation__.Notify(118483)

	if inSnap.Desc != nil {
		__antithesis_instrumentation__.Notify(118527)
		if !raft.IsEmptySnap(rd.Snapshot) {
			__antithesis_instrumentation__.Notify(118528)
			snapUUID, err := uuid.FromBytes(rd.Snapshot.Data)
			if err != nil {
				__antithesis_instrumentation__.Notify(118533)
				const expl = "invalid snapshot id"
				return stats, expl, errors.Wrap(err, expl)
			} else {
				__antithesis_instrumentation__.Notify(118534)
			}
			__antithesis_instrumentation__.Notify(118529)
			if inSnap.SnapUUID == (uuid.UUID{}) {
				__antithesis_instrumentation__.Notify(118535)
				log.Fatalf(ctx, "programming error: a snapshot application was attempted outside of the streaming snapshot codepath")
			} else {
				__antithesis_instrumentation__.Notify(118536)
			}
			__antithesis_instrumentation__.Notify(118530)
			if snapUUID != inSnap.SnapUUID {
				__antithesis_instrumentation__.Notify(118537)
				log.Fatalf(ctx, "incoming snapshot id doesn't match raft snapshot id: %s != %s", snapUUID, inSnap.SnapUUID)
			} else {
				__antithesis_instrumentation__.Notify(118538)
			}
			__antithesis_instrumentation__.Notify(118531)

			subsumedRepls, releaseMergeLock := r.maybeAcquireSnapshotMergeLock(ctx, inSnap)
			defer releaseMergeLock()

			if err := r.applySnapshot(ctx, inSnap, rd.Snapshot, rd.HardState, subsumedRepls); err != nil {
				__antithesis_instrumentation__.Notify(118539)
				const expl = "while applying snapshot"
				return stats, expl, errors.Wrap(err, expl)
			} else {
				__antithesis_instrumentation__.Notify(118540)
			}
			__antithesis_instrumentation__.Notify(118532)
			stats.snap.applied = true

			r.mu.RLock()
			lastIndex = r.mu.lastIndex
			lastTerm = r.mu.lastTerm
			raftLogSize = r.mu.raftLogSize
			r.mu.RUnlock()

			if !r.store.TestingKnobs().DisableRefreshReasonSnapshotApplied && func() bool {
				__antithesis_instrumentation__.Notify(118541)
				return refreshReason == noReason == true
			}() == true {
				__antithesis_instrumentation__.Notify(118542)
				refreshReason = reasonSnapshotApplied
			} else {
				__antithesis_instrumentation__.Notify(118543)
			}
		} else {
			__antithesis_instrumentation__.Notify(118544)
		}
	} else {
		__antithesis_instrumentation__.Notify(118545)
		if !raft.IsEmptySnap(rd.Snapshot) {
			__antithesis_instrumentation__.Notify(118546)

			err := makeNonDeterministicFailure(
				"have inSnap=nil, but raft has a snapshot %s",
				raft.DescribeSnapshot(rd.Snapshot),
			)
			return stats, getNonDeterministicFailureExplanation(err), err
		} else {
			__antithesis_instrumentation__.Notify(118547)
		}
	}
	__antithesis_instrumentation__.Notify(118484)

	sm := r.getStateMachine()
	dec := r.getDecoder()
	appTask := apply.MakeTask(sm, dec)
	appTask.SetMaxBatchSize(r.store.TestingKnobs().MaxApplicationBatchSize)
	defer appTask.Close()
	if err := appTask.Decode(ctx, rd.CommittedEntries); err != nil {
		__antithesis_instrumentation__.Notify(118548)
		return stats, getNonDeterministicFailureExplanation(err), err
	} else {
		__antithesis_instrumentation__.Notify(118549)
	}
	__antithesis_instrumentation__.Notify(118485)
	if err := appTask.AckCommittedEntriesBeforeApplication(ctx, lastIndex); err != nil {
		__antithesis_instrumentation__.Notify(118550)
		return stats, getNonDeterministicFailureExplanation(err), err
	} else {
		__antithesis_instrumentation__.Notify(118551)
	}
	__antithesis_instrumentation__.Notify(118486)

	msgApps, otherMsgs := splitMsgApps(rd.Messages)
	r.traceMessageSends(msgApps, "sending msgApp")
	r.sendRaftMessagesRaftMuLocked(ctx, msgApps)

	batch := r.store.Engine().NewUnindexedBatch(false)
	defer batch.Close()

	prevLastIndex := lastIndex
	if len(rd.Entries) > 0 {
		__antithesis_instrumentation__.Notify(118552)

		thinEntries, sideLoadedEntriesSize, err := r.maybeSideloadEntriesRaftMuLocked(ctx, rd.Entries)
		if err != nil {
			__antithesis_instrumentation__.Notify(118554)
			const expl = "during sideloading"
			return stats, expl, errors.Wrap(err, expl)
		} else {
			__antithesis_instrumentation__.Notify(118555)
		}
		__antithesis_instrumentation__.Notify(118553)
		raftLogSize += sideLoadedEntriesSize
		if lastIndex, lastTerm, raftLogSize, err = r.append(
			ctx, batch, lastIndex, lastTerm, raftLogSize, thinEntries,
		); err != nil {
			__antithesis_instrumentation__.Notify(118556)
			const expl = "during append"
			return stats, expl, errors.Wrap(err, expl)
		} else {
			__antithesis_instrumentation__.Notify(118557)
		}
	} else {
		__antithesis_instrumentation__.Notify(118558)
	}
	__antithesis_instrumentation__.Notify(118487)
	if !raft.IsEmptyHardState(rd.HardState) {
		__antithesis_instrumentation__.Notify(118559)
		if !r.IsInitialized() && func() bool {
			__antithesis_instrumentation__.Notify(118561)
			return rd.HardState.Commit != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(118562)
			log.Fatalf(ctx, "setting non-zero HardState.Commit on uninitialized replica %s. HS=%+v", r, rd.HardState)
		} else {
			__antithesis_instrumentation__.Notify(118563)
		}
		__antithesis_instrumentation__.Notify(118560)

		if err := r.raftMu.stateLoader.SetHardState(ctx, batch, rd.HardState); err != nil {
			__antithesis_instrumentation__.Notify(118564)
			const expl = "during setHardState"
			return stats, expl, errors.Wrap(err, expl)
		} else {
			__antithesis_instrumentation__.Notify(118565)
		}
	} else {
		__antithesis_instrumentation__.Notify(118566)
	}
	__antithesis_instrumentation__.Notify(118488)

	commitStart := timeutil.Now()
	if err := batch.Commit(rd.MustSync && func() bool {
		__antithesis_instrumentation__.Notify(118567)
		return !disableSyncRaftLog.Get(&r.store.cfg.Settings.SV) == true
	}() == true); err != nil {
		__antithesis_instrumentation__.Notify(118568)
		const expl = "while committing batch"
		return stats, expl, errors.Wrap(err, expl)
	} else {
		__antithesis_instrumentation__.Notify(118569)
	}
	__antithesis_instrumentation__.Notify(118489)
	if rd.MustSync {
		__antithesis_instrumentation__.Notify(118570)
		elapsed := timeutil.Since(commitStart)
		r.store.metrics.RaftLogCommitLatency.RecordValue(elapsed.Nanoseconds())
	} else {
		__antithesis_instrumentation__.Notify(118571)
	}
	__antithesis_instrumentation__.Notify(118490)

	if len(rd.Entries) > 0 {
		__antithesis_instrumentation__.Notify(118572)

		firstPurge := rd.Entries[0].Index
		purgeTerm := rd.Entries[0].Term - 1
		lastPurge := prevLastIndex
		purgedSize, err := maybePurgeSideloaded(ctx, r.raftMu.sideloaded, firstPurge, lastPurge, purgeTerm)
		if err != nil {
			__antithesis_instrumentation__.Notify(118574)
			const expl = "while purging sideloaded storage"
			return stats, expl, err
		} else {
			__antithesis_instrumentation__.Notify(118575)
		}
		__antithesis_instrumentation__.Notify(118573)
		raftLogSize -= purgedSize
		if raftLogSize < 0 {
			__antithesis_instrumentation__.Notify(118576)

			raftLogSize = 0
		} else {
			__antithesis_instrumentation__.Notify(118577)
		}
	} else {
		__antithesis_instrumentation__.Notify(118578)
	}
	__antithesis_instrumentation__.Notify(118491)

	r.mu.Lock()
	r.mu.lastIndex = lastIndex
	r.mu.lastTerm = lastTerm
	r.mu.raftLogSize = raftLogSize
	var becameLeader bool
	if r.mu.leaderID != leaderID {
		__antithesis_instrumentation__.Notify(118579)
		r.mu.leaderID = leaderID

		becameLeader = r.mu.leaderID == r.replicaID
	} else {
		__antithesis_instrumentation__.Notify(118580)
	}
	__antithesis_instrumentation__.Notify(118492)
	r.mu.Unlock()

	if becameLeader && func() bool {
		__antithesis_instrumentation__.Notify(118581)
		return r.store.replicateQueue != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(118582)
		r.store.replicateQueue.MaybeAddAsync(ctx, r, r.store.Clock().NowAsClockTimestamp())
	} else {
		__antithesis_instrumentation__.Notify(118583)
	}
	__antithesis_instrumentation__.Notify(118493)

	r.store.raftEntryCache.Add(r.RangeID, rd.Entries, true)
	r.sendRaftMessagesRaftMuLocked(ctx, otherMsgs)
	r.traceEntries(rd.CommittedEntries, "committed, before applying any entries")

	applicationStart := timeutil.Now()
	if len(rd.CommittedEntries) > 0 {
		__antithesis_instrumentation__.Notify(118584)
		err := appTask.ApplyCommittedEntries(ctx)
		stats.applyCommittedEntriesStats = sm.moveStats()
		if errors.Is(err, apply.ErrRemoved) {
			__antithesis_instrumentation__.Notify(118586)

			return stats, "", err
		} else {
			__antithesis_instrumentation__.Notify(118587)
			if err != nil {
				__antithesis_instrumentation__.Notify(118588)
				return stats, getNonDeterministicFailureExplanation(err), err
			} else {
				__antithesis_instrumentation__.Notify(118589)
			}
		}
		__antithesis_instrumentation__.Notify(118585)

		if stats.numEmptyEntries > 0 {
			__antithesis_instrumentation__.Notify(118590)

			if !r.store.TestingKnobs().DisableRefreshReasonNewLeaderOrConfigChange {
				__antithesis_instrumentation__.Notify(118591)
				refreshReason = reasonNewLeaderOrConfigChange
			} else {
				__antithesis_instrumentation__.Notify(118592)
			}
		} else {
			__antithesis_instrumentation__.Notify(118593)
		}
	} else {
		__antithesis_instrumentation__.Notify(118594)
	}
	__antithesis_instrumentation__.Notify(118494)
	applicationElapsed := timeutil.Since(applicationStart).Nanoseconds()
	r.store.metrics.RaftApplyCommittedLatency.RecordValue(applicationElapsed)
	r.store.metrics.RaftCommandsApplied.Inc(int64(len(rd.CommittedEntries)))
	if r.store.TestingKnobs().EnableUnconditionalRefreshesInRaftReady {
		__antithesis_instrumentation__.Notify(118595)
		refreshReason = reasonNewLeaderOrConfigChange
	} else {
		__antithesis_instrumentation__.Notify(118596)
	}
	__antithesis_instrumentation__.Notify(118495)
	if refreshReason != noReason {
		__antithesis_instrumentation__.Notify(118597)
		r.mu.Lock()
		r.refreshProposalsLocked(ctx, 0, refreshReason)
		r.mu.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(118598)
	}
	__antithesis_instrumentation__.Notify(118496)

	const expl = "during advance"

	r.mu.Lock()
	err = r.withRaftGroupLocked(true, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118599)
		raftGroup.Advance(rd)
		if stats.numConfChangeEntries > 0 {
			__antithesis_instrumentation__.Notify(118602)

			if shouldCampaignAfterConfChange(ctx, r.store.StoreID(), r.descRLocked(), raftGroup) {
				__antithesis_instrumentation__.Notify(118603)
				r.campaignLocked(ctx)
			} else {
				__antithesis_instrumentation__.Notify(118604)
			}
		} else {
			__antithesis_instrumentation__.Notify(118605)
		}
		__antithesis_instrumentation__.Notify(118600)

		if raftGroup.HasReady() {
			__antithesis_instrumentation__.Notify(118606)
			r.store.enqueueRaftUpdateCheck(r.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(118607)
		}
		__antithesis_instrumentation__.Notify(118601)
		return true, nil
	})
	__antithesis_instrumentation__.Notify(118497)
	r.mu.applyingEntries = false
	r.mu.Unlock()
	if err != nil {
		__antithesis_instrumentation__.Notify(118608)
		return stats, expl, errors.Wrap(err, expl)
	} else {
		__antithesis_instrumentation__.Notify(118609)
	}
	__antithesis_instrumentation__.Notify(118498)

	r.updateProposalQuotaRaftMuLocked(ctx, lastLeaderID)
	return stats, "", nil
}

func splitMsgApps(msgs []raftpb.Message) (msgApps, otherMsgs []raftpb.Message) {
	__antithesis_instrumentation__.Notify(118610)
	splitIdx := 0
	for i, msg := range msgs {
		__antithesis_instrumentation__.Notify(118612)
		if msg.Type == raftpb.MsgApp {
			__antithesis_instrumentation__.Notify(118613)
			msgs[i], msgs[splitIdx] = msgs[splitIdx], msgs[i]
			splitIdx++
		} else {
			__antithesis_instrumentation__.Notify(118614)
		}
	}
	__antithesis_instrumentation__.Notify(118611)
	return msgs[:splitIdx], msgs[splitIdx:]
}

func maybeFatalOnRaftReadyErr(ctx context.Context, expl string, err error) (removed bool) {
	__antithesis_instrumentation__.Notify(118615)
	switch {
	case err == nil:
		__antithesis_instrumentation__.Notify(118616)
		return false
	case errors.Is(err, apply.ErrRemoved):
		__antithesis_instrumentation__.Notify(118617)
		return true
	default:
		__antithesis_instrumentation__.Notify(118618)
		log.FatalfDepth(ctx, 1, "%s: %+v", redact.Safe(expl), err)
		panic("unreachable")
	}
}

func (r *Replica) tick(ctx context.Context, livenessMap liveness.IsLiveMap) (bool, error) {
	__antithesis_instrumentation__.Notify(118619)
	r.unreachablesMu.Lock()
	remotes := r.unreachablesMu.remotes
	r.unreachablesMu.remotes = nil
	r.unreachablesMu.Unlock()

	r.raftMu.Lock()
	defer r.raftMu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.mu.internalRaftGroup == nil {
		__antithesis_instrumentation__.Notify(118628)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(118629)
	}
	__antithesis_instrumentation__.Notify(118620)

	for remoteReplica := range remotes {
		__antithesis_instrumentation__.Notify(118630)
		r.mu.internalRaftGroup.ReportUnreachable(uint64(remoteReplica))
	}
	__antithesis_instrumentation__.Notify(118621)

	if r.mu.quiescent {
		__antithesis_instrumentation__.Notify(118631)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(118632)
	}
	__antithesis_instrumentation__.Notify(118622)

	now := r.store.Clock().NowAsClockTimestamp()
	if r.maybeQuiesceRaftMuLockedReplicaMuLocked(ctx, now, livenessMap) {
		__antithesis_instrumentation__.Notify(118633)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(118634)
	}
	__antithesis_instrumentation__.Notify(118623)

	r.maybeTransferRaftLeadershipToLeaseholderLocked(ctx, now)

	if r.replicaID == r.mu.leaderID {
		__antithesis_instrumentation__.Notify(118635)
		r.mu.lastUpdateTimes.update(r.replicaID, timeutil.Now())
	} else {
		__antithesis_instrumentation__.Notify(118636)
	}
	__antithesis_instrumentation__.Notify(118624)

	r.mu.ticks++
	preTickState := r.mu.internalRaftGroup.BasicStatus().RaftState
	r.mu.internalRaftGroup.Tick()
	postTickState := r.mu.internalRaftGroup.BasicStatus().RaftState
	if preTickState != postTickState {
		__antithesis_instrumentation__.Notify(118637)
		if postTickState == raft.StatePreCandidate {
			__antithesis_instrumentation__.Notify(118638)
			r.store.Metrics().RaftTimeoutCampaign.Inc(1)
			if k := r.store.TestingKnobs(); k != nil && func() bool {
				__antithesis_instrumentation__.Notify(118639)
				return k.OnRaftTimeoutCampaign != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(118640)
				k.OnRaftTimeoutCampaign(r.RangeID)
			} else {
				__antithesis_instrumentation__.Notify(118641)
			}
		} else {
			__antithesis_instrumentation__.Notify(118642)
		}
	} else {
		__antithesis_instrumentation__.Notify(118643)
	}
	__antithesis_instrumentation__.Notify(118625)

	refreshAtDelta := r.store.cfg.RaftElectionTimeoutTicks
	if knob := r.store.TestingKnobs().RefreshReasonTicksPeriod; knob > 0 {
		__antithesis_instrumentation__.Notify(118644)
		refreshAtDelta = knob
	} else {
		__antithesis_instrumentation__.Notify(118645)
	}
	__antithesis_instrumentation__.Notify(118626)
	if !r.store.TestingKnobs().DisableRefreshReasonTicks && func() bool {
		__antithesis_instrumentation__.Notify(118646)
		return r.mu.ticks%refreshAtDelta == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(118647)

		r.refreshProposalsLocked(ctx, refreshAtDelta, reasonTicks)
	} else {
		__antithesis_instrumentation__.Notify(118648)
	}
	__antithesis_instrumentation__.Notify(118627)
	return true, nil
}

func (r *Replica) hasRaftReadyRLocked() bool {
	__antithesis_instrumentation__.Notify(118649)
	return r.mu.internalRaftGroup.HasReady()
}

func (r *Replica) slowReplicationThreshold(ba *roachpb.BatchRequest) (time.Duration, bool) {
	__antithesis_instrumentation__.Notify(118650)
	if knobs := r.store.TestingKnobs(); knobs != nil && func() bool {
		__antithesis_instrumentation__.Notify(118652)
		return knobs.SlowReplicationThresholdOverride != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(118653)
		if dur := knobs.SlowReplicationThresholdOverride(ba); dur > 0 {
			__antithesis_instrumentation__.Notify(118654)
			return dur, true
		} else {
			__antithesis_instrumentation__.Notify(118655)
		}

	} else {
		__antithesis_instrumentation__.Notify(118656)
	}
	__antithesis_instrumentation__.Notify(118651)
	dur := replicaCircuitBreakerSlowReplicationThreshold.Get(&r.store.cfg.Settings.SV)
	return dur, dur > 0
}

type refreshRaftReason int

const (
	noReason refreshRaftReason = iota
	reasonNewLeader
	reasonNewLeaderOrConfigChange

	reasonSnapshotApplied
	reasonTicks
)

func (r *Replica) refreshProposalsLocked(
	ctx context.Context, refreshAtDelta int, reason refreshRaftReason,
) {
	__antithesis_instrumentation__.Notify(118657)
	if refreshAtDelta != 0 && func() bool {
		__antithesis_instrumentation__.Notify(118662)
		return reason != reasonTicks == true
	}() == true {
		__antithesis_instrumentation__.Notify(118663)
		log.Fatalf(ctx, "refreshAtDelta specified for reason %s != reasonTicks", reason)
	} else {
		__antithesis_instrumentation__.Notify(118664)
	}
	__antithesis_instrumentation__.Notify(118658)

	var maxSlowProposalDurationRequest *roachpb.BatchRequest

	var maxSlowProposalDuration time.Duration
	var slowProposalCount int64
	var reproposals pendingCmdSlice
	for _, p := range r.mu.proposals {
		__antithesis_instrumentation__.Notify(118665)
		slowReplicationThreshold, ok := r.slowReplicationThreshold(p.Request)

		inflightDuration := r.store.cfg.RaftTickInterval * time.Duration(r.mu.ticks-p.createdAtTicks)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(118667)
			return inflightDuration > slowReplicationThreshold == true
		}() == true {
			__antithesis_instrumentation__.Notify(118668)
			if maxSlowProposalDuration < inflightDuration {
				__antithesis_instrumentation__.Notify(118669)
				maxSlowProposalDuration = inflightDuration
				maxSlowProposalDurationRequest = p.Request
				slowProposalCount++
			} else {
				__antithesis_instrumentation__.Notify(118670)
			}
		} else {
			__antithesis_instrumentation__.Notify(118671)
		}
		__antithesis_instrumentation__.Notify(118666)
		switch reason {
		case reasonSnapshotApplied:
			__antithesis_instrumentation__.Notify(118672)

			if p.command.MaxLeaseIndex <= r.mu.state.LeaseAppliedIndex {
				__antithesis_instrumentation__.Notify(118676)
				r.cleanupFailedProposalLocked(p)
				log.Eventf(p.ctx, "retry proposal %x: %s", p.idKey, reason)
				p.finishApplication(ctx, proposalResult{
					Err: roachpb.NewError(
						roachpb.NewAmbiguousResultErrorf(
							"unable to determine whether command was applied via snapshot",
						),
					),
				})
			} else {
				__antithesis_instrumentation__.Notify(118677)
			}
			__antithesis_instrumentation__.Notify(118673)
			continue

		case reasonTicks:
			__antithesis_instrumentation__.Notify(118674)
			if p.proposedAtTicks <= r.mu.ticks-refreshAtDelta {
				__antithesis_instrumentation__.Notify(118678)

				reproposals = append(reproposals, p)
			} else {
				__antithesis_instrumentation__.Notify(118679)
			}

		default:
			__antithesis_instrumentation__.Notify(118675)

			reproposals = append(reproposals, p)
		}
	}
	__antithesis_instrumentation__.Notify(118659)

	r.store.metrics.SlowRaftRequests.Update(slowProposalCount)

	if maxSlowProposalDuration > 0 && func() bool {
		__antithesis_instrumentation__.Notify(118680)
		return r.breaker.Signal().Err() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(118681)
		err := errors.Errorf("have been waiting %.2fs for slow proposal %s",
			maxSlowProposalDuration.Seconds(), maxSlowProposalDurationRequest)
		log.Warningf(ctx, "%s", err)

		r.breaker.TripAsync(err)
	} else {
		__antithesis_instrumentation__.Notify(118682)
	}
	__antithesis_instrumentation__.Notify(118660)

	if log.V(1) && func() bool {
		__antithesis_instrumentation__.Notify(118683)
		return len(reproposals) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(118684)
		log.Infof(ctx,
			"pending commands: reproposing %d (at %d.%d) %s",
			len(reproposals), r.mu.state.RaftAppliedIndex,
			r.mu.state.LeaseAppliedIndex, reason)
	} else {
		__antithesis_instrumentation__.Notify(118685)
	}
	__antithesis_instrumentation__.Notify(118661)

	sort.Sort(reproposals)
	for _, p := range reproposals {
		__antithesis_instrumentation__.Notify(118686)
		log.Eventf(p.ctx, "re-submitting command %x to Raft: %s", p.idKey, reason)
		if err := r.mu.proposalBuf.ReinsertLocked(ctx, p); err != nil {
			__antithesis_instrumentation__.Notify(118687)
			r.cleanupFailedProposalLocked(p)
			p.finishApplication(ctx, proposalResult{
				Err: roachpb.NewError(roachpb.NewAmbiguousResultError(err)),
			})
		} else {
			__antithesis_instrumentation__.Notify(118688)
		}
	}
}

func (r *Replica) poisonInflightLatches(err error) {
	__antithesis_instrumentation__.Notify(118689)
	r.raftMu.Lock()
	defer r.raftMu.Unlock()
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.mu.proposals {
		__antithesis_instrumentation__.Notify(118690)
		p.ec.poison()
		if p.ec.g.Req.PoisonPolicy == poison.Policy_Error {
			__antithesis_instrumentation__.Notify(118691)
			aErr := roachpb.NewAmbiguousResultError(err)

			p.signalProposalResult(proposalResult{Err: roachpb.NewError(aErr)})
		} else {
			__antithesis_instrumentation__.Notify(118692)
		}
	}
}

func (r *Replica) maybeCoalesceHeartbeat(
	ctx context.Context,
	msg raftpb.Message,
	toReplica, fromReplica roachpb.ReplicaDescriptor,
	quiesce bool,
	lagging laggingReplicaSet,
) bool {
	__antithesis_instrumentation__.Notify(118693)
	var hbMap map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat
	switch msg.Type {
	case raftpb.MsgHeartbeat:
		__antithesis_instrumentation__.Notify(118696)
		r.store.coalescedMu.Lock()
		hbMap = r.store.coalescedMu.heartbeats
	case raftpb.MsgHeartbeatResp:
		__antithesis_instrumentation__.Notify(118697)
		r.store.coalescedMu.Lock()
		hbMap = r.store.coalescedMu.heartbeatResponses
	default:
		__antithesis_instrumentation__.Notify(118698)
		return false
	}
	__antithesis_instrumentation__.Notify(118694)
	beat := kvserverpb.RaftHeartbeat{
		RangeID:                           r.RangeID,
		ToReplicaID:                       toReplica.ReplicaID,
		FromReplicaID:                     fromReplica.ReplicaID,
		Term:                              msg.Term,
		Commit:                            msg.Commit,
		Quiesce:                           quiesce,
		LaggingFollowersOnQuiesce:         lagging,
		LaggingFollowersOnQuiesceAccurate: quiesce,
	}
	if log.V(4) {
		__antithesis_instrumentation__.Notify(118699)
		log.Infof(ctx, "coalescing beat: %+v", beat)
	} else {
		__antithesis_instrumentation__.Notify(118700)
	}
	__antithesis_instrumentation__.Notify(118695)
	toStore := roachpb.StoreIdent{
		StoreID: toReplica.StoreID,
		NodeID:  toReplica.NodeID,
	}
	hbMap[toStore] = append(hbMap[toStore], beat)
	r.store.coalescedMu.Unlock()
	return true
}

func (r *Replica) sendRaftMessagesRaftMuLocked(ctx context.Context, messages []raftpb.Message) {
	__antithesis_instrumentation__.Notify(118701)
	var lastAppResp raftpb.Message
	for _, message := range messages {
		__antithesis_instrumentation__.Notify(118703)
		drop := false
		switch message.Type {
		case raftpb.MsgApp:
			__antithesis_instrumentation__.Notify(118705)
			if util.RaceEnabled {
				__antithesis_instrumentation__.Notify(118708)

				prevTerm := message.LogTerm
				prevIndex := message.Index
				for j := range message.Entries {
					__antithesis_instrumentation__.Notify(118709)
					ent := &message.Entries[j]
					assertSideloadedRaftCommandInlined(ctx, ent)

					if prevIndex+1 != ent.Index {
						__antithesis_instrumentation__.Notify(118712)
						log.Fatalf(ctx,
							"index gap in outgoing MsgApp: idx %d followed by %d",
							prevIndex, ent.Index,
						)
					} else {
						__antithesis_instrumentation__.Notify(118713)
					}
					__antithesis_instrumentation__.Notify(118710)
					prevIndex = ent.Index
					if prevTerm > ent.Term {
						__antithesis_instrumentation__.Notify(118714)
						log.Fatalf(ctx,
							"term regression in outgoing MsgApp: idx %d at term=%d "+
								"appended with logterm=%d",
							ent.Index, ent.Term, message.LogTerm,
						)
					} else {
						__antithesis_instrumentation__.Notify(118715)
					}
					__antithesis_instrumentation__.Notify(118711)
					prevTerm = ent.Term
				}
			} else {
				__antithesis_instrumentation__.Notify(118716)
			}

		case raftpb.MsgAppResp:
			__antithesis_instrumentation__.Notify(118706)

			if !message.Reject && func() bool {
				__antithesis_instrumentation__.Notify(118717)
				return message.Index > lastAppResp.Index == true
			}() == true {
				__antithesis_instrumentation__.Notify(118718)
				lastAppResp = message
				drop = true
			} else {
				__antithesis_instrumentation__.Notify(118719)
			}
		default:
			__antithesis_instrumentation__.Notify(118707)
		}
		__antithesis_instrumentation__.Notify(118704)

		if !drop {
			__antithesis_instrumentation__.Notify(118720)
			r.sendRaftMessageRaftMuLocked(ctx, message)
		} else {
			__antithesis_instrumentation__.Notify(118721)
		}
	}
	__antithesis_instrumentation__.Notify(118702)
	if lastAppResp.Index > 0 {
		__antithesis_instrumentation__.Notify(118722)
		r.sendRaftMessageRaftMuLocked(ctx, lastAppResp)
	} else {
		__antithesis_instrumentation__.Notify(118723)
	}
}

func (r *Replica) sendRaftMessageRaftMuLocked(ctx context.Context, msg raftpb.Message) {
	__antithesis_instrumentation__.Notify(118724)
	r.mu.RLock()
	fromReplica, fromErr := r.getReplicaDescriptorByIDRLocked(roachpb.ReplicaID(msg.From), r.raftMu.lastToReplica)
	toReplica, toErr := r.getReplicaDescriptorByIDRLocked(roachpb.ReplicaID(msg.To), r.raftMu.lastFromReplica)
	var startKey roachpb.RKey
	if msg.Type == raftpb.MsgApp && func() bool {
		__antithesis_instrumentation__.Notify(118730)
		return r.mu.internalRaftGroup != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(118731)

		_ = maybeDropMsgApp

		r.mu.internalRaftGroup.WithProgress(func(id uint64, _ raft.ProgressType, pr tracker.Progress) {
			__antithesis_instrumentation__.Notify(118732)
			if id == msg.To && func() bool {
				__antithesis_instrumentation__.Notify(118733)
				return pr.State == tracker.StateProbe == true
			}() == true {
				__antithesis_instrumentation__.Notify(118734)

				startKey = r.descRLocked().StartKey
			} else {
				__antithesis_instrumentation__.Notify(118735)
			}
		})
	} else {
		__antithesis_instrumentation__.Notify(118736)
	}
	__antithesis_instrumentation__.Notify(118725)
	r.mu.RUnlock()

	if fromErr != nil {
		__antithesis_instrumentation__.Notify(118737)
		log.Warningf(ctx, "failed to look up sender replica %d in r%d while sending %s: %s",
			msg.From, r.RangeID, msg.Type, fromErr)
		return
	} else {
		__antithesis_instrumentation__.Notify(118738)
	}
	__antithesis_instrumentation__.Notify(118726)
	if toErr != nil {
		__antithesis_instrumentation__.Notify(118739)
		log.Warningf(ctx, "failed to look up recipient replica %d in r%d while sending %s: %s",
			msg.To, r.RangeID, msg.Type, toErr)
		return
	} else {
		__antithesis_instrumentation__.Notify(118740)
	}
	__antithesis_instrumentation__.Notify(118727)

	if msg.Type == raftpb.MsgSnap {
		__antithesis_instrumentation__.Notify(118741)
		r.store.raftSnapshotQueue.AddAsync(ctx, r, raftSnapshotPriority)
		return
	} else {
		__antithesis_instrumentation__.Notify(118742)
	}
	__antithesis_instrumentation__.Notify(118728)

	if r.maybeCoalesceHeartbeat(ctx, msg, toReplica, fromReplica, false, nil) {
		__antithesis_instrumentation__.Notify(118743)
		return
	} else {
		__antithesis_instrumentation__.Notify(118744)
	}
	__antithesis_instrumentation__.Notify(118729)

	req := newRaftMessageRequest()
	*req = kvserverpb.RaftMessageRequest{
		RangeID:       r.RangeID,
		ToReplica:     toReplica,
		FromReplica:   fromReplica,
		Message:       msg,
		RangeStartKey: startKey,
	}
	if !r.sendRaftMessageRequest(ctx, req) {
		__antithesis_instrumentation__.Notify(118745)
		if err := r.withRaftGroup(true, func(raftGroup *raft.RawNode) (bool, error) {
			__antithesis_instrumentation__.Notify(118746)
			r.mu.droppedMessages++
			raftGroup.ReportUnreachable(msg.To)
			return true, nil
		}); err != nil && func() bool {
			__antithesis_instrumentation__.Notify(118747)
			return !errors.Is(err, errRemoved) == true
		}() == true {
			__antithesis_instrumentation__.Notify(118748)
			log.Fatalf(ctx, "%v", err)
		} else {
			__antithesis_instrumentation__.Notify(118749)
		}
	} else {
		__antithesis_instrumentation__.Notify(118750)
	}
}

func (r *Replica) addUnreachableRemoteReplica(remoteReplica roachpb.ReplicaID) {
	__antithesis_instrumentation__.Notify(118751)
	r.unreachablesMu.Lock()
	if r.unreachablesMu.remotes == nil {
		__antithesis_instrumentation__.Notify(118753)
		r.unreachablesMu.remotes = make(map[roachpb.ReplicaID]struct{})
	} else {
		__antithesis_instrumentation__.Notify(118754)
	}
	__antithesis_instrumentation__.Notify(118752)
	r.unreachablesMu.remotes[remoteReplica] = struct{}{}
	r.unreachablesMu.Unlock()
}

func (r *Replica) sendRaftMessageRequest(
	ctx context.Context, req *kvserverpb.RaftMessageRequest,
) bool {
	__antithesis_instrumentation__.Notify(118755)
	if log.V(4) {
		__antithesis_instrumentation__.Notify(118757)
		log.Infof(ctx, "sending raft request %+v", req)
	} else {
		__antithesis_instrumentation__.Notify(118758)
	}
	__antithesis_instrumentation__.Notify(118756)
	return r.store.cfg.Transport.SendAsync(req, r.connectionClass.get())
}

func (r *Replica) reportSnapshotStatus(ctx context.Context, to roachpb.ReplicaID, snapErr error) {
	__antithesis_instrumentation__.Notify(118759)
	r.raftMu.Lock()
	defer r.raftMu.Unlock()

	snapStatus := raft.SnapshotFinish
	if snapErr != nil {
		__antithesis_instrumentation__.Notify(118761)
		snapStatus = raft.SnapshotFailure
	} else {
		__antithesis_instrumentation__.Notify(118762)
	}
	__antithesis_instrumentation__.Notify(118760)

	if err := r.withRaftGroup(true, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118763)
		raftGroup.ReportSnapshot(uint64(to), snapStatus)
		return true, nil
	}); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(118764)
		return !errors.Is(err, errRemoved) == true
	}() == true {
		__antithesis_instrumentation__.Notify(118765)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(118766)
	}
}

type snapTruncationInfo struct {
	index          uint64
	recipientStore roachpb.StoreID
	deadline       time.Time
}

func (r *Replica) addSnapshotLogTruncationConstraint(
	ctx context.Context, snapUUID uuid.UUID, index uint64, recipientStore roachpb.StoreID,
) {
	__antithesis_instrumentation__.Notify(118767)
	r.mu.Lock()
	defer r.mu.Unlock()
	r.addSnapshotLogTruncationConstraintLocked(ctx, snapUUID, index, recipientStore)
}

func (r *Replica) addSnapshotLogTruncationConstraintLocked(
	ctx context.Context, snapUUID uuid.UUID, index uint64, recipientStore roachpb.StoreID,
) {
	__antithesis_instrumentation__.Notify(118768)
	if r.mu.snapshotLogTruncationConstraints == nil {
		__antithesis_instrumentation__.Notify(118771)
		r.mu.snapshotLogTruncationConstraints = make(map[uuid.UUID]snapTruncationInfo)
	} else {
		__antithesis_instrumentation__.Notify(118772)
	}
	__antithesis_instrumentation__.Notify(118769)
	item, ok := r.mu.snapshotLogTruncationConstraints[snapUUID]
	if ok {
		__antithesis_instrumentation__.Notify(118773)

		log.Warningf(ctx, "UUID collision at %s for %+v (index %d)", snapUUID, item, index)
		return
	} else {
		__antithesis_instrumentation__.Notify(118774)
	}
	__antithesis_instrumentation__.Notify(118770)

	r.mu.snapshotLogTruncationConstraints[snapUUID] = snapTruncationInfo{
		index:          index,
		recipientStore: recipientStore,
	}
}

func (r *Replica) completeSnapshotLogTruncationConstraint(
	ctx context.Context, snapUUID uuid.UUID, now time.Time,
) {
	__antithesis_instrumentation__.Notify(118775)
	r.mu.Lock()
	defer r.mu.Unlock()

	item, ok := r.mu.snapshotLogTruncationConstraints[snapUUID]
	if !ok {
		__antithesis_instrumentation__.Notify(118777)

		return
	} else {
		__antithesis_instrumentation__.Notify(118778)
	}
	__antithesis_instrumentation__.Notify(118776)

	deadline := now.Add(RaftLogQueuePendingSnapshotGracePeriod)
	item.deadline = deadline
	r.mu.snapshotLogTruncationConstraints[snapUUID] = item
}

func (r *Replica) getAndGCSnapshotLogTruncationConstraints(
	now time.Time, recipientStore roachpb.StoreID,
) (minSnapIndex uint64) {
	__antithesis_instrumentation__.Notify(118779)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getAndGCSnapshotLogTruncationConstraintsLocked(now, recipientStore)
}

func (r *Replica) getAndGCSnapshotLogTruncationConstraintsLocked(
	now time.Time, recipientStore roachpb.StoreID,
) (minSnapIndex uint64) {
	__antithesis_instrumentation__.Notify(118780)
	for snapUUID, item := range r.mu.snapshotLogTruncationConstraints {
		__antithesis_instrumentation__.Notify(118783)
		if item.deadline != (time.Time{}) && func() bool {
			__antithesis_instrumentation__.Notify(118786)
			return item.deadline.Before(now) == true
		}() == true {
			__antithesis_instrumentation__.Notify(118787)

			delete(r.mu.snapshotLogTruncationConstraints, snapUUID)
			continue
		} else {
			__antithesis_instrumentation__.Notify(118788)
		}
		__antithesis_instrumentation__.Notify(118784)
		if recipientStore != 0 && func() bool {
			__antithesis_instrumentation__.Notify(118789)
			return item.recipientStore != recipientStore == true
		}() == true {
			__antithesis_instrumentation__.Notify(118790)
			continue
		} else {
			__antithesis_instrumentation__.Notify(118791)
		}
		__antithesis_instrumentation__.Notify(118785)
		if minSnapIndex == 0 || func() bool {
			__antithesis_instrumentation__.Notify(118792)
			return minSnapIndex > item.index == true
		}() == true {
			__antithesis_instrumentation__.Notify(118793)
			minSnapIndex = item.index
		} else {
			__antithesis_instrumentation__.Notify(118794)
		}
	}
	__antithesis_instrumentation__.Notify(118781)
	if len(r.mu.snapshotLogTruncationConstraints) == 0 {
		__antithesis_instrumentation__.Notify(118795)

		r.mu.snapshotLogTruncationConstraints = nil
	} else {
		__antithesis_instrumentation__.Notify(118796)
	}
	__antithesis_instrumentation__.Notify(118782)
	return minSnapIndex
}

func isRaftLeader(raftStatus *raft.Status) bool {
	__antithesis_instrumentation__.Notify(118797)
	return raftStatus != nil && func() bool {
		__antithesis_instrumentation__.Notify(118798)
		return raftStatus.SoftState.RaftState == raft.StateLeader == true
	}() == true
}

func HasRaftLeader(raftStatus *raft.Status) bool {
	__antithesis_instrumentation__.Notify(118799)
	return raftStatus != nil && func() bool {
		__antithesis_instrumentation__.Notify(118800)
		return raftStatus.SoftState.Lead != 0 == true
	}() == true
}

type pendingCmdSlice []*ProposalData

func (s pendingCmdSlice) Len() int { __antithesis_instrumentation__.Notify(118801); return len(s) }
func (s pendingCmdSlice) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(118802)
	s[i], s[j] = s[j], s[i]
}
func (s pendingCmdSlice) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(118803)
	return s[i].command.MaxLeaseIndex < s[j].command.MaxLeaseIndex
}

func (r *Replica) withRaftGroupLocked(
	mayCampaignOnWake bool, f func(r *raft.RawNode) (unquiesceAndWakeLeader bool, _ error),
) error {
	__antithesis_instrumentation__.Notify(118804)
	if r.mu.destroyStatus.Removed() {
		__antithesis_instrumentation__.Notify(118810)

		return errRemoved
	} else {
		__antithesis_instrumentation__.Notify(118811)
	}
	__antithesis_instrumentation__.Notify(118805)

	if r.mu.internalRaftGroup == nil {
		__antithesis_instrumentation__.Notify(118812)
		ctx := r.AnnotateCtx(context.TODO())
		raftGroup, err := raft.NewRawNode(newRaftConfig(
			raft.Storage((*replicaRaftStorage)(r)),
			uint64(r.replicaID),
			r.mu.state.RaftAppliedIndex,
			r.store.cfg,
			&raftLogger{ctx: ctx},
		))
		if err != nil {
			__antithesis_instrumentation__.Notify(118814)
			return err
		} else {
			__antithesis_instrumentation__.Notify(118815)
		}
		__antithesis_instrumentation__.Notify(118813)
		r.mu.internalRaftGroup = raftGroup

		if mayCampaignOnWake {
			__antithesis_instrumentation__.Notify(118816)
			r.maybeCampaignOnWakeLocked(ctx)
		} else {
			__antithesis_instrumentation__.Notify(118817)
		}
	} else {
		__antithesis_instrumentation__.Notify(118818)
	}
	__antithesis_instrumentation__.Notify(118806)

	unquiesce, err := func(rangeID roachpb.RangeID, raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118819)
		return f(raftGroup)
	}(r.RangeID, r.mu.internalRaftGroup)
	__antithesis_instrumentation__.Notify(118807)
	if r.mu.internalRaftGroup.BasicStatus().Lead == 0 {
		__antithesis_instrumentation__.Notify(118820)

		unquiesce = true
	} else {
		__antithesis_instrumentation__.Notify(118821)
	}
	__antithesis_instrumentation__.Notify(118808)
	if unquiesce {
		__antithesis_instrumentation__.Notify(118822)
		r.maybeUnquiesceAndWakeLeaderLocked()
	} else {
		__antithesis_instrumentation__.Notify(118823)
	}
	__antithesis_instrumentation__.Notify(118809)
	return err
}

func (r *Replica) withRaftGroup(
	mayCampaignOnWake bool, f func(r *raft.RawNode) (unquiesceAndWakeLeader bool, _ error),
) error {
	__antithesis_instrumentation__.Notify(118824)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.withRaftGroupLocked(mayCampaignOnWake, f)
}

func shouldCampaignOnWake(
	leaseStatus kvserverpb.LeaseStatus,
	storeID roachpb.StoreID,
	raftStatus raft.BasicStatus,
	livenessMap liveness.IsLiveMap,
	desc *roachpb.RangeDescriptor,
	requiresExpiringLease bool,
) bool {
	__antithesis_instrumentation__.Notify(118825)

	if leaseStatus.IsValid() && func() bool {
		__antithesis_instrumentation__.Notify(118832)
		return !leaseStatus.OwnedBy(storeID) == true
	}() == true {
		__antithesis_instrumentation__.Notify(118833)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118834)
	}
	__antithesis_instrumentation__.Notify(118826)

	if raftStatus.RaftState != raft.StateFollower {
		__antithesis_instrumentation__.Notify(118835)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118836)
	}
	__antithesis_instrumentation__.Notify(118827)

	if raftStatus.Lead == raft.None {
		__antithesis_instrumentation__.Notify(118837)
		return true
	} else {
		__antithesis_instrumentation__.Notify(118838)
	}
	__antithesis_instrumentation__.Notify(118828)

	if requiresExpiringLease {
		__antithesis_instrumentation__.Notify(118839)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118840)
	}
	__antithesis_instrumentation__.Notify(118829)

	replDesc, ok := desc.GetReplicaDescriptorByID(roachpb.ReplicaID(raftStatus.Lead))
	if !ok {
		__antithesis_instrumentation__.Notify(118841)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118842)
	}
	__antithesis_instrumentation__.Notify(118830)

	livenessEntry, ok := livenessMap[replDesc.NodeID]
	if !ok {
		__antithesis_instrumentation__.Notify(118843)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118844)
	}
	__antithesis_instrumentation__.Notify(118831)
	return !livenessEntry.IsLive
}

func (r *Replica) maybeCampaignOnWakeLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(118845)

	if _, currentMember := r.mu.state.Desc.GetReplicaDescriptorByID(r.replicaID); !currentMember {
		__antithesis_instrumentation__.Notify(118847)
		return
	} else {
		__antithesis_instrumentation__.Notify(118848)
	}
	__antithesis_instrumentation__.Notify(118846)

	leaseStatus := r.leaseStatusAtRLocked(ctx, r.store.Clock().NowAsClockTimestamp())
	raftStatus := r.mu.internalRaftGroup.BasicStatus()
	livenessMap, _ := r.store.livenessMap.Load().(liveness.IsLiveMap)
	if shouldCampaignOnWake(leaseStatus, r.store.StoreID(), raftStatus, livenessMap, r.descRLocked(), r.requiresExpiringLeaseRLocked()) {
		__antithesis_instrumentation__.Notify(118849)
		r.campaignLocked(ctx)
	} else {
		__antithesis_instrumentation__.Notify(118850)
	}
}

func (r *Replica) campaignLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(118851)
	log.VEventf(ctx, 3, "campaigning")
	if err := r.mu.internalRaftGroup.Campaign(); err != nil {
		__antithesis_instrumentation__.Notify(118853)
		log.VEventf(ctx, 1, "failed to campaign: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(118854)
	}
	__antithesis_instrumentation__.Notify(118852)
	r.store.enqueueRaftUpdateCheck(r.RangeID)
}

type lastUpdateTimesMap map[roachpb.ReplicaID]time.Time

func (m lastUpdateTimesMap) update(replicaID roachpb.ReplicaID, now time.Time) {
	__antithesis_instrumentation__.Notify(118855)
	if m == nil {
		__antithesis_instrumentation__.Notify(118857)
		return
	} else {
		__antithesis_instrumentation__.Notify(118858)
	}
	__antithesis_instrumentation__.Notify(118856)
	m[replicaID] = now
}

func (m lastUpdateTimesMap) updateOnUnquiesce(
	descs []roachpb.ReplicaDescriptor, prs map[uint64]tracker.Progress, now time.Time,
) {
	__antithesis_instrumentation__.Notify(118859)
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(118860)
		if prs[uint64(desc.ReplicaID)].State == tracker.StateReplicate {
			__antithesis_instrumentation__.Notify(118861)
			m.update(desc.ReplicaID, now)
		} else {
			__antithesis_instrumentation__.Notify(118862)
		}
	}
}

func (m lastUpdateTimesMap) updateOnBecomeLeader(descs []roachpb.ReplicaDescriptor, now time.Time) {
	__antithesis_instrumentation__.Notify(118863)
	for _, desc := range descs {
		__antithesis_instrumentation__.Notify(118864)
		m.update(desc.ReplicaID, now)
	}
}

func (m lastUpdateTimesMap) isFollowerActiveSince(
	ctx context.Context, replicaID roachpb.ReplicaID, now time.Time, threshold time.Duration,
) bool {
	__antithesis_instrumentation__.Notify(118865)
	lastUpdateTime, ok := m[replicaID]
	if !ok {
		__antithesis_instrumentation__.Notify(118867)

		return false
	} else {
		__antithesis_instrumentation__.Notify(118868)
	}
	__antithesis_instrumentation__.Notify(118866)
	return now.Sub(lastUpdateTime) <= threshold
}

func (r *Replica) maybeAcquireSnapshotMergeLock(
	ctx context.Context, inSnap IncomingSnapshot,
) (subsumedRepls []*Replica, releaseMergeLock func()) {
	__antithesis_instrumentation__.Notify(118869)

	endKey := r.Desc().EndKey
	if endKey == nil {
		__antithesis_instrumentation__.Notify(118872)

		return nil, func() { __antithesis_instrumentation__.Notify(118873) }
	} else {
		__antithesis_instrumentation__.Notify(118874)
	}
	__antithesis_instrumentation__.Notify(118870)
	for endKey.Less(inSnap.Desc.EndKey) {
		__antithesis_instrumentation__.Notify(118875)
		sRepl := r.store.LookupReplica(endKey)
		if sRepl == nil || func() bool {
			__antithesis_instrumentation__.Notify(118877)
			return !endKey.Equal(sRepl.Desc().StartKey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(118878)
			log.Fatalf(ctx, "snapshot widens existing replica, but no replica exists for subsumed key %s", endKey)
		} else {
			__antithesis_instrumentation__.Notify(118879)
		}
		__antithesis_instrumentation__.Notify(118876)
		sRepl.raftMu.Lock()
		subsumedRepls = append(subsumedRepls, sRepl)
		endKey = sRepl.Desc().EndKey
	}
	__antithesis_instrumentation__.Notify(118871)
	return subsumedRepls, func() {
		__antithesis_instrumentation__.Notify(118880)
		for _, sr := range subsumedRepls {
			__antithesis_instrumentation__.Notify(118881)
			sr.raftMu.Unlock()
		}
	}
}

func (r *Replica) maybeAcquireSplitMergeLock(
	ctx context.Context, raftCmd kvserverpb.RaftCommand,
) (func(), error) {
	__antithesis_instrumentation__.Notify(118882)
	if split := raftCmd.ReplicatedEvalResult.Split; split != nil {
		__antithesis_instrumentation__.Notify(118884)
		return r.acquireSplitLock(ctx, &split.SplitTrigger)
	} else {
		__antithesis_instrumentation__.Notify(118885)
		if merge := raftCmd.ReplicatedEvalResult.Merge; merge != nil {
			__antithesis_instrumentation__.Notify(118886)
			return r.acquireMergeLock(ctx, &merge.MergeTrigger)
		} else {
			__antithesis_instrumentation__.Notify(118887)
		}
	}
	__antithesis_instrumentation__.Notify(118883)
	return nil, nil
}

func (r *Replica) acquireSplitLock(
	ctx context.Context, split *roachpb.SplitTrigger,
) (func(), error) {
	__antithesis_instrumentation__.Notify(118888)
	rightReplDesc, _ := split.RightDesc.GetReplicaDescriptor(r.StoreID())
	rightRepl, _, err := r.store.getOrCreateReplica(
		ctx, split.RightDesc.RangeID, rightReplDesc.ReplicaID, nil,
	)

	if errors.HasType(err, (*roachpb.RaftGroupDeletedError)(nil)) {
		__antithesis_instrumentation__.Notify(118892)
		return func() { __antithesis_instrumentation__.Notify(118893) }, nil
	} else {
		__antithesis_instrumentation__.Notify(118894)
	}
	__antithesis_instrumentation__.Notify(118889)
	if err != nil {
		__antithesis_instrumentation__.Notify(118895)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(118896)
	}
	__antithesis_instrumentation__.Notify(118890)

	if rightRepl.IsInitialized() {
		__antithesis_instrumentation__.Notify(118897)
		return nil, errors.Errorf("RHS of split %s / %s already initialized before split application",
			&split.LeftDesc, &split.RightDesc)
	} else {
		__antithesis_instrumentation__.Notify(118898)
	}
	__antithesis_instrumentation__.Notify(118891)
	return rightRepl.raftMu.Unlock, nil
}

func (r *Replica) acquireMergeLock(
	ctx context.Context, merge *roachpb.MergeTrigger,
) (func(), error) {
	__antithesis_instrumentation__.Notify(118899)

	rightReplDesc, _ := merge.RightDesc.GetReplicaDescriptor(r.StoreID())
	rightRepl, _, err := r.store.getOrCreateReplica(
		ctx, merge.RightDesc.RangeID, rightReplDesc.ReplicaID, nil,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(118902)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(118903)
	}
	__antithesis_instrumentation__.Notify(118900)
	rightDesc := rightRepl.Desc()
	if !rightDesc.StartKey.Equal(merge.RightDesc.StartKey) || func() bool {
		__antithesis_instrumentation__.Notify(118904)
		return !rightDesc.EndKey.Equal(merge.RightDesc.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(118905)
		return nil, errors.Errorf("RHS of merge %s <- %s not present on store; found %s in place of the RHS",
			&merge.LeftDesc, &merge.RightDesc, rightDesc)
	} else {
		__antithesis_instrumentation__.Notify(118906)
	}
	__antithesis_instrumentation__.Notify(118901)
	return rightRepl.raftMu.Unlock, nil
}

func handleTruncatedStateBelowRaftPreApply(
	ctx context.Context,
	currentTruncatedState, suggestedTruncatedState *roachpb.RaftTruncatedState,
	loader stateloader.StateLoader,
	readWriter storage.ReadWriter,
) (_apply bool, _ error) {
	__antithesis_instrumentation__.Notify(118907)
	if suggestedTruncatedState.Index <= currentTruncatedState.Index {
		__antithesis_instrumentation__.Notify(118911)

		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(118912)
	}
	__antithesis_instrumentation__.Notify(118908)

	prefixBuf := &loader.RangeIDPrefixBuf
	numTruncatedEntries := suggestedTruncatedState.Index - currentTruncatedState.Index
	if numTruncatedEntries >= raftLogTruncationClearRangeThreshold {
		__antithesis_instrumentation__.Notify(118913)
		start := prefixBuf.RaftLogKey(currentTruncatedState.Index + 1).Clone()
		end := prefixBuf.RaftLogKey(suggestedTruncatedState.Index + 1).Clone()
		if err := readWriter.ClearRawRange(start, end); err != nil {
			__antithesis_instrumentation__.Notify(118914)
			return false, errors.Wrapf(err,
				"unable to clear truncated Raft entries for %+v between indexes %d-%d",
				suggestedTruncatedState, currentTruncatedState.Index+1, suggestedTruncatedState.Index+1)
		} else {
			__antithesis_instrumentation__.Notify(118915)
		}
	} else {
		__antithesis_instrumentation__.Notify(118916)

		for idx := currentTruncatedState.Index + 1; idx <= suggestedTruncatedState.Index; idx++ {
			__antithesis_instrumentation__.Notify(118917)
			if err := readWriter.ClearUnversioned(prefixBuf.RaftLogKey(idx)); err != nil {
				__antithesis_instrumentation__.Notify(118918)
				return false, errors.Wrapf(err, "unable to clear truncated Raft entries for %+v at index %d",
					suggestedTruncatedState, idx)
			} else {
				__antithesis_instrumentation__.Notify(118919)
			}
		}
	}
	__antithesis_instrumentation__.Notify(118909)

	if err := storage.MVCCPutProto(
		ctx, readWriter, nil, prefixBuf.RaftTruncatedStateKey(),
		hlc.Timestamp{}, nil, suggestedTruncatedState,
	); err != nil {
		__antithesis_instrumentation__.Notify(118920)
		return false, errors.Wrap(err, "unable to write RaftTruncatedState")
	} else {
		__antithesis_instrumentation__.Notify(118921)
	}
	__antithesis_instrumentation__.Notify(118910)

	return true, nil
}

func ComputeRaftLogSize(
	ctx context.Context, rangeID roachpb.RangeID, reader storage.Reader, sideloaded SideloadStorage,
) (int64, error) {
	__antithesis_instrumentation__.Notify(118922)
	prefix := keys.RaftLogPrefix(rangeID)
	prefixEnd := prefix.PrefixEnd()
	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		LowerBound: prefix,
		UpperBound: prefixEnd,
	})
	defer iter.Close()
	ms, err := iter.ComputeStats(prefix, prefixEnd, 0)
	if err != nil {
		__antithesis_instrumentation__.Notify(118925)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(118926)
	}
	__antithesis_instrumentation__.Notify(118923)
	var totalSideloaded int64
	if sideloaded != nil {
		__antithesis_instrumentation__.Notify(118927)
		var err error

		_, totalSideloaded, err = sideloaded.BytesIfTruncatedFromTo(ctx, 0, 0)
		if err != nil {
			__antithesis_instrumentation__.Notify(118928)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(118929)
		}
	} else {
		__antithesis_instrumentation__.Notify(118930)
	}
	__antithesis_instrumentation__.Notify(118924)
	return ms.SysBytes + totalSideloaded, nil
}

func shouldCampaignAfterConfChange(
	ctx context.Context,
	storeID roachpb.StoreID,
	desc *roachpb.RangeDescriptor,
	raftGroup *raft.RawNode,
) bool {
	__antithesis_instrumentation__.Notify(118931)

	st := raftGroup.BasicStatus()
	if st.Lead == 0 {
		__antithesis_instrumentation__.Notify(118935)

		return false
	} else {
		__antithesis_instrumentation__.Notify(118936)
	}
	__antithesis_instrumentation__.Notify(118932)
	if !desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(118937)

		return false
	} else {
		__antithesis_instrumentation__.Notify(118938)
	}
	__antithesis_instrumentation__.Notify(118933)

	_, leaderStillThere := desc.GetReplicaDescriptorByID(roachpb.ReplicaID(st.Lead))
	if !leaderStillThere && func() bool {
		__antithesis_instrumentation__.Notify(118939)
		return storeID == desc.Replicas().VoterDescriptors()[0].StoreID == true
	}() == true {
		__antithesis_instrumentation__.Notify(118940)
		log.VEventf(ctx, 3, "leader got removed by conf change")
		return true
	} else {
		__antithesis_instrumentation__.Notify(118941)
	}
	__antithesis_instrumentation__.Notify(118934)
	return false
}

func getNonDeterministicFailureExplanation(err error) string {
	__antithesis_instrumentation__.Notify(118942)
	if nd := (*nonDeterministicFailure)(nil); errors.As(err, &nd) {
		__antithesis_instrumentation__.Notify(118944)
		return nd.safeExpl
	} else {
		__antithesis_instrumentation__.Notify(118945)
	}
	__antithesis_instrumentation__.Notify(118943)
	return "???"
}

func (r *Replica) printRaftTail(
	ctx context.Context, maxEntries, maxCharsPerEntry int,
) (string, error) {
	__antithesis_instrumentation__.Notify(118946)
	start := keys.RaftLogPrefix(r.RangeID)
	end := keys.RaftLogPrefix(r.RangeID).PrefixEnd()

	it := r.Engine().NewEngineIterator(storage.IterOptions{LowerBound: start, UpperBound: end})
	valid, err := it.SeekEngineKeyLT(storage.EngineKey{Key: end})
	if err != nil {
		__antithesis_instrumentation__.Notify(118950)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(118951)
	}
	__antithesis_instrumentation__.Notify(118947)
	if !valid {
		__antithesis_instrumentation__.Notify(118952)
		return "", errors.AssertionFailedf("iterator invalid but no error")
	} else {
		__antithesis_instrumentation__.Notify(118953)
	}
	__antithesis_instrumentation__.Notify(118948)

	var sb strings.Builder
	for i := 0; i < maxEntries; i++ {
		__antithesis_instrumentation__.Notify(118954)
		key, err := it.EngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(118958)
			return sb.String(), err
		} else {
			__antithesis_instrumentation__.Notify(118959)
		}
		__antithesis_instrumentation__.Notify(118955)
		mvccKey, err := key.ToMVCCKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(118960)
			return sb.String(), err
		} else {
			__antithesis_instrumentation__.Notify(118961)
		}
		__antithesis_instrumentation__.Notify(118956)
		kv := storage.MVCCKeyValue{
			Key:   mvccKey,
			Value: it.Value(),
		}
		sb.WriteString(truncateEntryString(SprintMVCCKeyValue(kv, true), 2000))
		sb.WriteRune('\n')

		valid, err := it.PrevEngineKey()
		if err != nil {
			__antithesis_instrumentation__.Notify(118962)
			return sb.String(), err
		} else {
			__antithesis_instrumentation__.Notify(118963)
		}
		__antithesis_instrumentation__.Notify(118957)
		if !valid {
			__antithesis_instrumentation__.Notify(118964)

			break
		} else {
			__antithesis_instrumentation__.Notify(118965)
		}
	}
	__antithesis_instrumentation__.Notify(118949)
	return sb.String(), nil
}

func truncateEntryString(s string, maxChars int) string {
	__antithesis_instrumentation__.Notify(118966)
	res := s
	if len(s) > maxChars {
		__antithesis_instrumentation__.Notify(118968)
		if maxChars > 3 {
			__antithesis_instrumentation__.Notify(118970)
			maxChars -= 3
		} else {
			__antithesis_instrumentation__.Notify(118971)
		}
		__antithesis_instrumentation__.Notify(118969)
		res = s[0:maxChars] + "..."
	} else {
		__antithesis_instrumentation__.Notify(118972)
	}
	__antithesis_instrumentation__.Notify(118967)
	return res
}
