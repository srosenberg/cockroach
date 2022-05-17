package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/closedts/tracker"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type propBuf struct {
	p        proposer
	clock    *hlc.Clock
	settings *cluster.Settings

	evalTracker tracker.Tracker
	full        sync.Cond

	arr propBufArray

	allocatedIdx int64

	assignedLAI uint64

	assignedClosedTimestamp hlc.Timestamp

	scratchFooter kvserverpb.RaftCommandFooter

	testing struct {
		leaseIndexFilter func(*ProposalData) (indexOverride uint64)

		insertFilter func(*ProposalData) error

		submitProposalFilter func(*ProposalData) (drop bool, err error)

		allowLeaseProposalWhenNotLeader bool

		dontCloseTimestamps bool
	}
}

type rangeLeaderInfo struct {
	leaderKnown bool

	leader roachpb.ReplicaID

	iAmTheLeader bool

	leaderEligibleForLease bool
}

type proposer interface {
	locker() sync.Locker
	rlocker() sync.Locker

	getReplicaID() roachpb.ReplicaID
	destroyed() destroyStatus
	leaseAppliedIndex() uint64
	enqueueUpdateCheck()
	closedTimestampTarget() hlc.Timestamp

	withGroupLocked(func(proposerRaft) error) error
	registerProposalLocked(*ProposalData)
	leaderStatusRLocked(raftGroup proposerRaft) rangeLeaderInfo
	ownsValidLeaseRLocked(ctx context.Context, now hlc.ClockTimestamp) bool

	rejectProposalWithRedirectLocked(
		ctx context.Context,
		prop *ProposalData,
		redirectTo roachpb.ReplicaID,
	)

	leaseDebugRLocked() string
}

type proposerRaft interface {
	Step(raftpb.Message) error
	BasicStatus() raft.BasicStatus
	ProposeConfChange(raftpb.ConfChangeI) error
}

func (b *propBuf) Init(
	p proposer, tracker tracker.Tracker, clock *hlc.Clock, settings *cluster.Settings,
) {
	__antithesis_instrumentation__.Notify(118044)
	b.p = p
	b.full.L = p.rlocker()
	b.clock = clock
	b.evalTracker = tracker
	b.settings = settings
	b.assignedLAI = p.leaseAppliedIndex()
}

func (b *propBuf) AllocatedIdx() int {
	__antithesis_instrumentation__.Notify(118045)
	return int(atomic.LoadInt64(&b.allocatedIdx))
}

func (b *propBuf) clearAllocatedIdx() int {
	__antithesis_instrumentation__.Notify(118046)
	return int(atomic.SwapInt64(&b.allocatedIdx, 0))
}

func (b *propBuf) incAllocatedIdx() int {
	__antithesis_instrumentation__.Notify(118047)
	return int(atomic.AddInt64(&b.allocatedIdx, 1)) - 1
}

func (b *propBuf) Insert(ctx context.Context, p *ProposalData, tok TrackedRequestToken) error {
	__antithesis_instrumentation__.Notify(118048)
	defer tok.DoneIfNotMoved(ctx)

	b.p.rlocker().Lock()
	defer b.p.rlocker().Unlock()

	if filter := b.testing.insertFilter; filter != nil {
		__antithesis_instrumentation__.Notify(118052)
		if err := filter(p); err != nil {
			__antithesis_instrumentation__.Notify(118053)
			return err
		} else {
			__antithesis_instrumentation__.Notify(118054)
		}
	} else {
		__antithesis_instrumentation__.Notify(118055)
	}
	__antithesis_instrumentation__.Notify(118049)

	idx, err := b.allocateIndex(ctx, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(118056)
		return err
	} else {
		__antithesis_instrumentation__.Notify(118057)
	}
	__antithesis_instrumentation__.Notify(118050)

	if log.V(4) {
		__antithesis_instrumentation__.Notify(118058)
		log.Infof(p.ctx, "submitting proposal %x", p.idKey)
	} else {
		__antithesis_instrumentation__.Notify(118059)
	}
	__antithesis_instrumentation__.Notify(118051)

	p.tok = tok.Move(ctx)
	b.insertIntoArray(p, idx)
	return nil
}

func (b *propBuf) ReinsertLocked(ctx context.Context, p *ProposalData) error {
	__antithesis_instrumentation__.Notify(118060)

	idx, err := b.allocateIndex(ctx, true)
	if err != nil {
		__antithesis_instrumentation__.Notify(118062)
		return err
	} else {
		__antithesis_instrumentation__.Notify(118063)
	}
	__antithesis_instrumentation__.Notify(118061)

	b.insertIntoArray(p, idx)
	return nil
}

func (b *propBuf) allocateIndex(ctx context.Context, wLocked bool) (int, error) {
	__antithesis_instrumentation__.Notify(118064)

	for {
		__antithesis_instrumentation__.Notify(118065)

		if status := b.p.destroyed(); !status.IsAlive() {
			__antithesis_instrumentation__.Notify(118067)
			return 0, status.err
		} else {
			__antithesis_instrumentation__.Notify(118068)
		}
		__antithesis_instrumentation__.Notify(118066)

		idx := b.incAllocatedIdx()
		if idx < b.arr.len() {
			__antithesis_instrumentation__.Notify(118069)

			return idx, nil
		} else {
			__antithesis_instrumentation__.Notify(118070)
			if wLocked {
				__antithesis_instrumentation__.Notify(118071)

				if err := b.flushLocked(ctx); err != nil {
					__antithesis_instrumentation__.Notify(118072)
					return 0, err
				} else {
					__antithesis_instrumentation__.Notify(118073)
				}
			} else {
				__antithesis_instrumentation__.Notify(118074)
				if idx == b.arr.len() {
					__antithesis_instrumentation__.Notify(118075)

					if err := b.flushRLocked(ctx); err != nil {
						__antithesis_instrumentation__.Notify(118076)
						return 0, err
					} else {
						__antithesis_instrumentation__.Notify(118077)
					}
				} else {
					__antithesis_instrumentation__.Notify(118078)

					b.full.Wait()
				}
			}
		}
	}
}

func (b *propBuf) insertIntoArray(p *ProposalData, idx int) {
	__antithesis_instrumentation__.Notify(118079)
	b.arr.asSlice()[idx] = p
	if idx == 0 {
		__antithesis_instrumentation__.Notify(118080)

		b.p.enqueueUpdateCheck()
	} else {
		__antithesis_instrumentation__.Notify(118081)
	}
}

func (b *propBuf) flushRLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(118082)

	b.p.rlocker().Unlock()
	defer b.p.rlocker().Lock()
	b.p.locker().Lock()
	defer b.p.locker().Unlock()
	if status := b.p.destroyed(); !status.IsAlive() {
		__antithesis_instrumentation__.Notify(118084)
		b.full.Broadcast()
		return status.err
	} else {
		__antithesis_instrumentation__.Notify(118085)
	}
	__antithesis_instrumentation__.Notify(118083)
	return b.flushLocked(ctx)
}

func (b *propBuf) flushLocked(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(118086)
	return b.p.withGroupLocked(func(raftGroup proposerRaft) error {
		__antithesis_instrumentation__.Notify(118087)
		_, err := b.FlushLockedWithRaftGroup(ctx, raftGroup)
		return err
	})
}

func (b *propBuf) FlushLockedWithRaftGroup(
	ctx context.Context, raftGroup proposerRaft,
) (int, error) {
	__antithesis_instrumentation__.Notify(118088)

	used := b.clearAllocatedIdx()

	defer b.arr.adjustSize(used)
	if used == 0 {
		__antithesis_instrumentation__.Notify(118093)

		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(118094)
		if used > b.arr.len() {
			__antithesis_instrumentation__.Notify(118095)

			used = b.arr.len()
			defer b.full.Broadcast()
		} else {
			__antithesis_instrumentation__.Notify(118096)
		}
	}
	__antithesis_instrumentation__.Notify(118089)

	buf := b.arr.asSlice()[:used]
	ents := make([]raftpb.Entry, 0, used)

	var leaderInfo rangeLeaderInfo
	if raftGroup != nil {
		__antithesis_instrumentation__.Notify(118097)
		leaderInfo = b.p.leaderStatusRLocked(raftGroup)

		if leaderInfo.leaderKnown && func() bool {
			__antithesis_instrumentation__.Notify(118098)
			return leaderInfo.leader == b.p.getReplicaID() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(118099)
			return !leaderInfo.iAmTheLeader == true
		}() == true {
			__antithesis_instrumentation__.Notify(118100)
			log.Fatalf(ctx,
				"inconsistent Raft state: state %s while the current replica is also the lead: %d",
				raftGroup.BasicStatus().RaftState, leaderInfo.leader)
		} else {
			__antithesis_instrumentation__.Notify(118101)
		}
	} else {
		__antithesis_instrumentation__.Notify(118102)
	}
	__antithesis_instrumentation__.Notify(118090)

	closedTSTarget := b.p.closedTimestampTarget()

	var firstErr error
	for i, p := range buf {
		__antithesis_instrumentation__.Notify(118103)
		if p == nil {
			__antithesis_instrumentation__.Notify(118110)
			log.Fatalf(ctx, "unexpected nil proposal in buffer")
			return 0, nil
		} else {
			__antithesis_instrumentation__.Notify(118111)
		}
		__antithesis_instrumentation__.Notify(118104)
		buf[i] = nil
		reproposal := !p.tok.stillTracked()

		if !leaderInfo.iAmTheLeader && func() bool {
			__antithesis_instrumentation__.Notify(118112)
			return p.Request.IsLeaseRequest() == true
		}() == true {
			__antithesis_instrumentation__.Notify(118113)
			leaderKnownAndEligible := leaderInfo.leaderKnown && func() bool {
				__antithesis_instrumentation__.Notify(118115)
				return leaderInfo.leaderEligibleForLease == true
			}() == true
			ownsCurrentLease := b.p.ownsValidLeaseRLocked(ctx, b.clock.NowAsClockTimestamp())
			if leaderKnownAndEligible && func() bool {
				__antithesis_instrumentation__.Notify(118116)
				return !ownsCurrentLease == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(118117)
				return !b.testing.allowLeaseProposalWhenNotLeader == true
			}() == true {
				__antithesis_instrumentation__.Notify(118118)
				log.VEventf(ctx, 2, "not proposing lease acquisition because we're not the leader; replica %d is",
					leaderInfo.leader)
				b.p.rejectProposalWithRedirectLocked(ctx, p, leaderInfo.leader)
				p.tok.doneIfNotMovedLocked(ctx)
				continue
			} else {
				__antithesis_instrumentation__.Notify(118119)
			}
			__antithesis_instrumentation__.Notify(118114)

			if ownsCurrentLease {
				__antithesis_instrumentation__.Notify(118120)
				log.VEventf(ctx, 2, "proposing lease extension even though we're not the leader; we hold the current lease")
			} else {
				__antithesis_instrumentation__.Notify(118121)
				if !leaderInfo.leaderKnown {
					__antithesis_instrumentation__.Notify(118122)
					log.VEventf(ctx, 2, "proposing lease acquisition even though we're not the leader; the leader is unknown")
				} else {
					__antithesis_instrumentation__.Notify(118123)
					log.VEventf(ctx, 2, "proposing lease acquisition even though we're not the leader; the leader is ineligible")
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(118124)
		}
		__antithesis_instrumentation__.Notify(118105)

		b.p.registerProposalLocked(p)

		if !reproposal && func() bool {
			__antithesis_instrumentation__.Notify(118125)
			return p.Request.AppliesTimestampCache() == true
		}() == true {
			__antithesis_instrumentation__.Notify(118126)

			wts := p.Request.WriteTimestamp()
			lb := b.evalTracker.LowerBound(ctx)
			if wts.Less(lb) {
				__antithesis_instrumentation__.Notify(118127)
				wts, lb := wts, lb
				log.Fatalf(ctx, "%v", errorutil.UnexpectedWithIssueErrorf(72428,
					"request writing below tracked lower bound: wts: %s < lb: %s; ba: %s; lease: %s.",
					wts, lb, p.Request, b.p.leaseDebugRLocked()))
			} else {
				__antithesis_instrumentation__.Notify(118128)
			}
		} else {
			__antithesis_instrumentation__.Notify(118129)
		}
		__antithesis_instrumentation__.Notify(118106)
		p.tok.doneIfNotMovedLocked(ctx)

		if raftGroup == nil || func() bool {
			__antithesis_instrumentation__.Notify(118130)
			return firstErr != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(118131)
			continue
		} else {
			__antithesis_instrumentation__.Notify(118132)
		}
		__antithesis_instrumentation__.Notify(118107)

		if !reproposal {
			__antithesis_instrumentation__.Notify(118133)
			lai, closedTimestamp, err := b.allocateLAIAndClosedTimestampLocked(ctx, p, closedTSTarget)
			if err != nil {
				__antithesis_instrumentation__.Notify(118135)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118136)
			}
			__antithesis_instrumentation__.Notify(118134)
			err = b.marshallLAIAndClosedTimestampToProposalLocked(ctx, p, lai, closedTimestamp)
			if err != nil {
				__antithesis_instrumentation__.Notify(118137)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118138)
			}
		} else {
			__antithesis_instrumentation__.Notify(118139)
		}
		__antithesis_instrumentation__.Notify(118108)

		if filter := b.testing.submitProposalFilter; filter != nil {
			__antithesis_instrumentation__.Notify(118140)
			if drop, err := filter(p); drop || func() bool {
				__antithesis_instrumentation__.Notify(118141)
				return err != nil == true
			}() == true {
				__antithesis_instrumentation__.Notify(118142)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118143)
			}
		} else {
			__antithesis_instrumentation__.Notify(118144)
		}
		__antithesis_instrumentation__.Notify(118109)

		if crt := p.command.ReplicatedEvalResult.ChangeReplicas; crt != nil {
			__antithesis_instrumentation__.Notify(118145)

			if err := proposeBatch(raftGroup, b.p.getReplicaID(), ents); err != nil {
				__antithesis_instrumentation__.Notify(118149)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118150)
			}
			__antithesis_instrumentation__.Notify(118146)
			ents = ents[len(ents):]

			confChangeCtx := kvserverpb.ConfChangeContext{
				CommandID: string(p.idKey),
				Payload:   p.encodedCommand,
			}
			encodedCtx, err := protoutil.Marshal(&confChangeCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(118151)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118152)
			}
			__antithesis_instrumentation__.Notify(118147)

			cc, err := crt.ConfChange(encodedCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(118153)
				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118154)
			}
			__antithesis_instrumentation__.Notify(118148)

			if err := raftGroup.ProposeConfChange(
				cc,
			); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(118155)
				return !errors.Is(err, raft.ErrProposalDropped) == true
			}() == true {
				__antithesis_instrumentation__.Notify(118156)

				firstErr = err
				continue
			} else {
				__antithesis_instrumentation__.Notify(118157)
			}
		} else {
			__antithesis_instrumentation__.Notify(118158)

			ents = append(ents, raftpb.Entry{
				Data: p.encodedCommand,
			})
		}
	}
	__antithesis_instrumentation__.Notify(118091)
	if firstErr != nil {
		__antithesis_instrumentation__.Notify(118159)
		return 0, firstErr
	} else {
		__antithesis_instrumentation__.Notify(118160)
	}
	__antithesis_instrumentation__.Notify(118092)
	return used, proposeBatch(raftGroup, b.p.getReplicaID(), ents)
}

func (b *propBuf) allocateLAIAndClosedTimestampLocked(
	ctx context.Context, p *ProposalData, closedTSTarget hlc.Timestamp,
) (uint64, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(118161)

	var lai uint64
	if !p.Request.IsLeaseRequest() {
		__antithesis_instrumentation__.Notify(118168)
		b.assignedLAI++
		lai = b.assignedLAI
	} else {
		__antithesis_instrumentation__.Notify(118169)
	}
	__antithesis_instrumentation__.Notify(118162)

	if filter := b.testing.leaseIndexFilter; filter != nil {
		__antithesis_instrumentation__.Notify(118170)
		if override := filter(p); override != 0 {
			__antithesis_instrumentation__.Notify(118171)
			lai = override
		} else {
			__antithesis_instrumentation__.Notify(118172)
		}
	} else {
		__antithesis_instrumentation__.Notify(118173)
	}
	__antithesis_instrumentation__.Notify(118163)

	if b.testing.dontCloseTimestamps {
		__antithesis_instrumentation__.Notify(118174)
		return lai, hlc.Timestamp{}, nil
	} else {
		__antithesis_instrumentation__.Notify(118175)
	}
	__antithesis_instrumentation__.Notify(118164)

	if p.Request.IsLeaseRequest() {
		__antithesis_instrumentation__.Notify(118176)
		return lai, hlc.Timestamp{}, nil
	} else {
		__antithesis_instrumentation__.Notify(118177)
	}
	__antithesis_instrumentation__.Notify(118165)

	lb := b.evalTracker.LowerBound(ctx)
	if !lb.IsEmpty() {
		__antithesis_instrumentation__.Notify(118178)

		closedTSTarget.Backward(lb.FloorPrev())
	} else {
		__antithesis_instrumentation__.Notify(118179)
	}
	__antithesis_instrumentation__.Notify(118166)

	closedTSTarget.Backward(p.leaseStatus.ClosedTimestampUpperBound())

	if !b.forwardClosedTimestampLocked(closedTSTarget) {
		__antithesis_instrumentation__.Notify(118180)
		closedTSTarget = b.assignedClosedTimestamp
	} else {
		__antithesis_instrumentation__.Notify(118181)
	}
	__antithesis_instrumentation__.Notify(118167)

	return lai, closedTSTarget, nil
}

func (b *propBuf) marshallLAIAndClosedTimestampToProposalLocked(
	ctx context.Context, p *ProposalData, lai uint64, closedTimestamp hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(118182)
	buf := &b.scratchFooter
	buf.MaxLeaseIndex = lai

	p.command.MaxLeaseIndex = lai

	buf.ClosedTimestamp = closedTimestamp

	if log.ExpensiveLogEnabled(ctx, 4) {
		__antithesis_instrumentation__.Notify(118184)
		log.VEventf(ctx, 4, "attaching closed timestamp %s to proposal %x",
			closedTimestamp, p.idKey)
	} else {
		__antithesis_instrumentation__.Notify(118185)
	}
	__antithesis_instrumentation__.Notify(118183)

	preLen := len(p.encodedCommand)
	p.encodedCommand = p.encodedCommand[:preLen+buf.Size()]
	_, err := protoutil.MarshalTo(buf, p.encodedCommand[preLen:])
	return err
}

func (b *propBuf) forwardAssignedLAILocked(v uint64) {
	__antithesis_instrumentation__.Notify(118186)
	if b.assignedLAI < v {
		__antithesis_instrumentation__.Notify(118187)
		b.assignedLAI = v
	} else {
		__antithesis_instrumentation__.Notify(118188)
	}
}

func (b *propBuf) forwardClosedTimestampLocked(closedTS hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(118189)
	return b.assignedClosedTimestamp.Forward(closedTS)
}

func proposeBatch(raftGroup proposerRaft, replID roachpb.ReplicaID, ents []raftpb.Entry) error {
	__antithesis_instrumentation__.Notify(118190)
	if len(ents) == 0 {
		__antithesis_instrumentation__.Notify(118193)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(118194)
	}
	__antithesis_instrumentation__.Notify(118191)
	if err := raftGroup.Step(raftpb.Message{
		Type:    raftpb.MsgProp,
		From:    uint64(replID),
		Entries: ents,
	}); errors.Is(err, raft.ErrProposalDropped) {
		__antithesis_instrumentation__.Notify(118195)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(118196)
		if err != nil {
			__antithesis_instrumentation__.Notify(118197)
			return err
		} else {
			__antithesis_instrumentation__.Notify(118198)
		}
	}
	__antithesis_instrumentation__.Notify(118192)
	return nil
}

func (b *propBuf) FlushLockedWithoutProposing(ctx context.Context) {
	__antithesis_instrumentation__.Notify(118199)
	if _, err := b.FlushLockedWithRaftGroup(ctx, nil); err != nil {
		__antithesis_instrumentation__.Notify(118200)
		log.Fatalf(ctx, "unexpected error: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(118201)
	}
}

func (b *propBuf) OnLeaseChangeLocked(
	leaseOwned bool, appliedClosedTS hlc.Timestamp, appliedLAI uint64,
) {
	__antithesis_instrumentation__.Notify(118202)
	if leaseOwned {
		__antithesis_instrumentation__.Notify(118203)
		b.forwardClosedTimestampLocked(appliedClosedTS)
		b.forwardAssignedLAILocked(appliedLAI)
	} else {
		__antithesis_instrumentation__.Notify(118204)

		b.assignedClosedTimestamp = hlc.Timestamp{}
		b.assignedLAI = 0
	}
}

func (b *propBuf) EvaluatingRequestsCount() int {
	__antithesis_instrumentation__.Notify(118205)
	b.p.rlocker().Lock()
	defer b.p.rlocker().Unlock()
	return b.evalTracker.Count()
}

type TrackedRequestToken struct {
	done bool
	tok  tracker.RemovalToken
	b    *propBuf
}

func (t *TrackedRequestToken) DoneIfNotMoved(ctx context.Context) {
	__antithesis_instrumentation__.Notify(118206)
	if t.done {
		__antithesis_instrumentation__.Notify(118209)
		return
	} else {
		__antithesis_instrumentation__.Notify(118210)
	}
	__antithesis_instrumentation__.Notify(118207)
	if t.b != nil {
		__antithesis_instrumentation__.Notify(118211)
		t.b.p.locker().Lock()
		defer t.b.p.locker().Unlock()
	} else {
		__antithesis_instrumentation__.Notify(118212)
	}
	__antithesis_instrumentation__.Notify(118208)
	t.doneIfNotMovedLocked(ctx)
}

func (t *TrackedRequestToken) doneIfNotMovedLocked(ctx context.Context) {
	__antithesis_instrumentation__.Notify(118213)
	if t.done {
		__antithesis_instrumentation__.Notify(118215)
		return
	} else {
		__antithesis_instrumentation__.Notify(118216)
	}
	__antithesis_instrumentation__.Notify(118214)
	t.done = true
	if t.b != nil {
		__antithesis_instrumentation__.Notify(118217)
		t.b.evalTracker.Untrack(ctx, t.tok)
	} else {
		__antithesis_instrumentation__.Notify(118218)
	}
}

func (t *TrackedRequestToken) stillTracked() bool {
	__antithesis_instrumentation__.Notify(118219)
	return !t.done
}

func (t *TrackedRequestToken) Move(ctx context.Context) TrackedRequestToken {
	__antithesis_instrumentation__.Notify(118220)
	if t.done {
		__antithesis_instrumentation__.Notify(118222)
		log.Fatalf(ctx, "attempting to Move() after Done() call")
	} else {
		__antithesis_instrumentation__.Notify(118223)
	}
	__antithesis_instrumentation__.Notify(118221)
	cpy := *t
	t.done = true
	return cpy
}

func (b *propBuf) TrackEvaluatingRequest(
	ctx context.Context, wts hlc.Timestamp,
) (minTS hlc.Timestamp, _ TrackedRequestToken) {
	__antithesis_instrumentation__.Notify(118224)
	b.p.rlocker().Lock()
	defer b.p.rlocker().Unlock()

	minTS = b.assignedClosedTimestamp.Next()
	wts.Forward(minTS)
	tok := b.evalTracker.Track(ctx, wts)
	return minTS, TrackedRequestToken{tok: tok, b: b}
}

func (b *propBuf) MaybeForwardClosedLocked(ctx context.Context, target hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(118225)
	if lb := b.evalTracker.LowerBound(ctx); !lb.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(118227)
		return lb.LessEq(target) == true
	}() == true {
		__antithesis_instrumentation__.Notify(118228)
		return false
	} else {
		__antithesis_instrumentation__.Notify(118229)
	}
	__antithesis_instrumentation__.Notify(118226)
	return b.forwardClosedTimestampLocked(target)
}

const propBufArrayMinSize = 4
const propBufArrayMaxSize = 256
const propBufArrayShrinkDelay = 16

type propBufArray struct {
	small  [propBufArrayMinSize]*ProposalData
	large  []*ProposalData
	shrink int
}

func (a *propBufArray) asSlice() []*ProposalData {
	__antithesis_instrumentation__.Notify(118230)
	if a.large != nil {
		__antithesis_instrumentation__.Notify(118232)
		return a.large
	} else {
		__antithesis_instrumentation__.Notify(118233)
	}
	__antithesis_instrumentation__.Notify(118231)
	return a.small[:]
}

func (a *propBufArray) len() int {
	__antithesis_instrumentation__.Notify(118234)
	return len(a.asSlice())
}

func (a *propBufArray) adjustSize(used int) {
	__antithesis_instrumentation__.Notify(118235)
	cur := a.len()
	switch {
	case used <= cur/4:
		__antithesis_instrumentation__.Notify(118236)

		if cur <= propBufArrayMinSize {
			__antithesis_instrumentation__.Notify(118240)
			return
		} else {
			__antithesis_instrumentation__.Notify(118241)
		}
		__antithesis_instrumentation__.Notify(118237)
		a.shrink++

		if a.shrink == propBufArrayShrinkDelay {
			__antithesis_instrumentation__.Notify(118242)
			a.shrink = 0
			next := cur / 2
			if next <= propBufArrayMinSize {
				__antithesis_instrumentation__.Notify(118243)
				a.large = nil
			} else {
				__antithesis_instrumentation__.Notify(118244)
				a.large = make([]*ProposalData, next)
			}
		} else {
			__antithesis_instrumentation__.Notify(118245)
		}
	case used >= cur:
		__antithesis_instrumentation__.Notify(118238)

		a.shrink = 0
		next := 2 * cur
		if next <= propBufArrayMaxSize {
			__antithesis_instrumentation__.Notify(118246)
			a.large = make([]*ProposalData, next)
		} else {
			__antithesis_instrumentation__.Notify(118247)
		}
	default:
		__antithesis_instrumentation__.Notify(118239)

		a.shrink = 0
	}
}

type replicaProposer Replica

var _ proposer = &replicaProposer{}

func (rp *replicaProposer) locker() sync.Locker {
	__antithesis_instrumentation__.Notify(118248)
	return &rp.mu.RWMutex
}

func (rp *replicaProposer) rlocker() sync.Locker {
	__antithesis_instrumentation__.Notify(118249)
	return rp.mu.RWMutex.RLocker()
}

func (rp *replicaProposer) getReplicaID() roachpb.ReplicaID {
	__antithesis_instrumentation__.Notify(118250)
	return rp.replicaID
}

func (rp *replicaProposer) destroyed() destroyStatus {
	__antithesis_instrumentation__.Notify(118251)
	return rp.mu.destroyStatus
}

func (rp *replicaProposer) leaseAppliedIndex() uint64 {
	__antithesis_instrumentation__.Notify(118252)
	return rp.mu.state.LeaseAppliedIndex
}

func (rp *replicaProposer) enqueueUpdateCheck() {
	__antithesis_instrumentation__.Notify(118253)
	rp.store.enqueueRaftUpdateCheck(rp.RangeID)
}

func (rp *replicaProposer) closedTimestampTarget() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(118254)
	return (*Replica)(rp).closedTimestampTargetRLocked()
}

func (rp *replicaProposer) withGroupLocked(fn func(raftGroup proposerRaft) error) error {
	__antithesis_instrumentation__.Notify(118255)

	return (*Replica)(rp).withRaftGroupLocked(true, func(raftGroup *raft.RawNode) (bool, error) {
		__antithesis_instrumentation__.Notify(118256)

		(*Replica)(rp).maybeUnquiesceLocked()
		return false, fn(raftGroup)
	})
}

func (rp *replicaProposer) leaseDebugRLocked() string {
	__antithesis_instrumentation__.Notify(118257)
	return rp.mu.state.Lease.String()
}

func (rp *replicaProposer) registerProposalLocked(p *ProposalData) {
	__antithesis_instrumentation__.Notify(118258)

	p.proposedAtTicks = rp.mu.ticks
	if p.createdAtTicks == 0 {
		__antithesis_instrumentation__.Notify(118260)
		p.createdAtTicks = rp.mu.ticks
	} else {
		__antithesis_instrumentation__.Notify(118261)
	}
	__antithesis_instrumentation__.Notify(118259)
	rp.mu.proposals[p.idKey] = p
}

func (rp *replicaProposer) leaderStatusRLocked(raftGroup proposerRaft) rangeLeaderInfo {
	__antithesis_instrumentation__.Notify(118262)
	r := (*Replica)(rp)

	status := raftGroup.BasicStatus()
	iAmTheLeader := status.RaftState == raft.StateLeader
	leader := status.Lead
	leaderKnown := leader != raft.None
	var leaderEligibleForLease bool
	rangeDesc := r.descRLocked()
	if leaderKnown {
		__antithesis_instrumentation__.Notify(118264)

		leaderRep, ok := rangeDesc.GetReplicaDescriptorByID(roachpb.ReplicaID(leader))
		if !ok {
			__antithesis_instrumentation__.Notify(118265)

			leaderEligibleForLease = true
		} else {
			__antithesis_instrumentation__.Notify(118266)
			err := roachpb.CheckCanReceiveLease(leaderRep, rangeDesc)
			leaderEligibleForLease = err == nil
		}
	} else {
		__antithesis_instrumentation__.Notify(118267)
	}
	__antithesis_instrumentation__.Notify(118263)
	return rangeLeaderInfo{
		leaderKnown:            leaderKnown,
		leader:                 roachpb.ReplicaID(leader),
		iAmTheLeader:           iAmTheLeader,
		leaderEligibleForLease: leaderEligibleForLease,
	}
}

func (rp *replicaProposer) ownsValidLeaseRLocked(ctx context.Context, now hlc.ClockTimestamp) bool {
	__antithesis_instrumentation__.Notify(118268)
	return (*Replica)(rp).ownsValidLeaseRLocked(ctx, now)
}

func (rp *replicaProposer) rejectProposalWithRedirectLocked(
	ctx context.Context, prop *ProposalData, redirectTo roachpb.ReplicaID,
) {
	__antithesis_instrumentation__.Notify(118269)
	r := (*Replica)(rp)
	rangeDesc := r.descRLocked()
	storeID := r.store.StoreID()
	r.store.metrics.LeaseRequestErrorCount.Inc(1)
	redirectRep, _ := rangeDesc.GetReplicaDescriptorByID(redirectTo)
	speculativeLease := roachpb.Lease{
		Replica: redirectRep,
	}
	log.VEventf(ctx, 2, "redirecting proposal to node %s; request: %s", redirectRep.NodeID, prop.Request)
	r.cleanupFailedProposalLocked(prop)
	prop.finishApplication(ctx, proposalResult{
		Err: roachpb.NewError(newNotLeaseHolderError(
			speculativeLease, storeID, rangeDesc, "refusing to acquire lease on follower")),
	})
}
