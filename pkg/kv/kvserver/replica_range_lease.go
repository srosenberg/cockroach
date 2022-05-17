package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/constraint"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/logtags"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
)

var leaseStatusLogLimiter = func() *log.EveryN {
	__antithesis_instrumentation__.Notify(119554)
	e := log.Every(15 * time.Second)
	e.ShouldLog()
	return &e
}()

type leaseRequestHandle struct {
	p *pendingLeaseRequest
	c chan *roachpb.Error
}

func (h *leaseRequestHandle) C() <-chan *roachpb.Error {
	__antithesis_instrumentation__.Notify(119555)
	if h.c == nil {
		__antithesis_instrumentation__.Notify(119557)
		panic("handle already canceled")
	} else {
		__antithesis_instrumentation__.Notify(119558)
	}
	__antithesis_instrumentation__.Notify(119556)
	return h.c
}

func (h *leaseRequestHandle) Cancel() {
	__antithesis_instrumentation__.Notify(119559)
	h.p.repl.mu.Lock()
	defer h.p.repl.mu.Unlock()
	if len(h.c) == 0 {
		__antithesis_instrumentation__.Notify(119561)

		delete(h.p.llHandles, h)

		if len(h.p.llHandles) == 0 {
			__antithesis_instrumentation__.Notify(119562)
			h.p.cancelLocked()
		} else {
			__antithesis_instrumentation__.Notify(119563)
		}
	} else {
		__antithesis_instrumentation__.Notify(119564)
	}
	__antithesis_instrumentation__.Notify(119560)

	h.c = nil
}

func (h *leaseRequestHandle) resolve(pErr *roachpb.Error) {
	__antithesis_instrumentation__.Notify(119565)
	h.c <- pErr
}

type pendingLeaseRequest struct {
	repl *Replica

	llHandles map[*leaseRequestHandle]struct{}

	cancelLocked func()

	nextLease roachpb.Lease
}

func makePendingLeaseRequest(repl *Replica) pendingLeaseRequest {
	__antithesis_instrumentation__.Notify(119566)
	return pendingLeaseRequest{
		repl:      repl,
		llHandles: make(map[*leaseRequestHandle]struct{}),
	}
}

func (p *pendingLeaseRequest) RequestPending() (roachpb.Lease, bool) {
	__antithesis_instrumentation__.Notify(119567)
	pending := p.cancelLocked != nil
	if pending {
		__antithesis_instrumentation__.Notify(119569)
		return p.nextLease, true
	} else {
		__antithesis_instrumentation__.Notify(119570)
	}
	__antithesis_instrumentation__.Notify(119568)
	return roachpb.Lease{}, false
}

func (p *pendingLeaseRequest) InitOrJoinRequest(
	ctx context.Context,
	nextLeaseHolder roachpb.ReplicaDescriptor,
	status kvserverpb.LeaseStatus,
	startKey roachpb.Key,
	transfer bool,
) *leaseRequestHandle {
	__antithesis_instrumentation__.Notify(119571)
	if nextLease, ok := p.RequestPending(); ok {
		__antithesis_instrumentation__.Notify(119577)
		if nextLease.Replica.ReplicaID == nextLeaseHolder.ReplicaID {
			__antithesis_instrumentation__.Notify(119579)

			return p.JoinRequest()
		} else {
			__antithesis_instrumentation__.Notify(119580)
		}
		__antithesis_instrumentation__.Notify(119578)

		return p.newResolvedHandle(roachpb.NewErrorf(
			"request for different replica in progress (requesting: %+v, in progress: %+v)",
			nextLeaseHolder.ReplicaID, nextLease.Replica.ReplicaID))
	} else {
		__antithesis_instrumentation__.Notify(119581)
	}
	__antithesis_instrumentation__.Notify(119572)

	acquisition := !status.Lease.OwnedBy(p.repl.store.StoreID())
	extension := !transfer && func() bool {
		__antithesis_instrumentation__.Notify(119582)
		return !acquisition == true
	}() == true
	_ = extension

	if acquisition {
		__antithesis_instrumentation__.Notify(119583)

		if status.State != kvserverpb.LeaseState_EXPIRED {
			__antithesis_instrumentation__.Notify(119584)
			log.Fatalf(ctx, "cannot acquire lease from another node before it has expired: %v", status)
		} else {
			__antithesis_instrumentation__.Notify(119585)
		}
	} else {
		__antithesis_instrumentation__.Notify(119586)
	}
	__antithesis_instrumentation__.Notify(119573)

	llHandle := p.newHandle()
	reqHeader := roachpb.RequestHeader{
		Key: startKey,
	}
	reqLease := roachpb.Lease{
		Start:      status.Now,
		Replica:    nextLeaseHolder,
		ProposedTS: &status.Now,
	}

	if p.repl.requiresExpiringLeaseRLocked() {
		__antithesis_instrumentation__.Notify(119587)
		reqLease.Expiration = &hlc.Timestamp{}
		*reqLease.Expiration = status.Now.ToTimestamp().Add(int64(p.repl.store.cfg.RangeLeaseActiveDuration()), 0)
	} else {
		__antithesis_instrumentation__.Notify(119588)

		l, ok := p.repl.store.cfg.NodeLiveness.GetLiveness(nextLeaseHolder.NodeID)
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(119590)
			return l.Epoch == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(119591)
			llHandle.resolve(roachpb.NewError(&roachpb.LeaseRejectedError{
				Existing:  status.Lease,
				Requested: reqLease,
				Message:   fmt.Sprintf("couldn't request lease for %+v: %v", nextLeaseHolder, liveness.ErrRecordCacheMiss),
			}))
			return llHandle
		} else {
			__antithesis_instrumentation__.Notify(119592)
		}
		__antithesis_instrumentation__.Notify(119589)
		reqLease.Epoch = l.Epoch
	}
	__antithesis_instrumentation__.Notify(119574)

	var leaseReq roachpb.Request
	if transfer {
		__antithesis_instrumentation__.Notify(119593)
		leaseReq = &roachpb.TransferLeaseRequest{
			RequestHeader: reqHeader,
			Lease:         reqLease,
			PrevLease:     status.Lease,
		}
	} else {
		__antithesis_instrumentation__.Notify(119594)
		minProposedTS := p.repl.mu.minLeaseProposedTS
		leaseReq = &roachpb.RequestLeaseRequest{
			RequestHeader: reqHeader,
			Lease:         reqLease,

			PrevLease:     status.Lease,
			MinProposedTS: &minProposedTS,
		}
	}
	__antithesis_instrumentation__.Notify(119575)

	if err := p.requestLeaseAsync(ctx, nextLeaseHolder, reqLease, status, leaseReq); err != nil {
		__antithesis_instrumentation__.Notify(119595)

		llHandle.resolve(roachpb.NewError(
			newNotLeaseHolderError(roachpb.Lease{}, p.repl.store.StoreID(), p.repl.mu.state.Desc,
				"lease acquisition task couldn't be started; node is shutting down")))
		return llHandle
	} else {
		__antithesis_instrumentation__.Notify(119596)
	}
	__antithesis_instrumentation__.Notify(119576)

	p.llHandles[llHandle] = struct{}{}
	p.nextLease = reqLease
	return llHandle
}

func (p *pendingLeaseRequest) requestLeaseAsync(
	parentCtx context.Context,
	nextLeaseHolder roachpb.ReplicaDescriptor,
	reqLease roachpb.Lease,
	status kvserverpb.LeaseStatus,
	leaseReq roachpb.Request,
) error {
	__antithesis_instrumentation__.Notify(119597)

	ctx := p.repl.AnnotateCtx(context.Background())

	const opName = "request range lease"
	tr := p.repl.AmbientContext.Tracer
	tagsOpt := tracing.WithLogTags(logtags.FromContext(parentCtx))
	var sp *tracing.Span
	if parentSp := tracing.SpanFromContext(parentCtx); parentSp != nil {
		__antithesis_instrumentation__.Notify(119602)

		ctx, sp = tr.StartSpanCtx(
			ctx,
			opName,
			tracing.WithParent(parentSp),
			tracing.WithFollowsFrom(),
			tagsOpt,
		)
	} else {
		__antithesis_instrumentation__.Notify(119603)
		ctx, sp = tr.StartSpanCtx(ctx, opName, tagsOpt)
	}
	__antithesis_instrumentation__.Notify(119598)

	ctx, cancel := context.WithCancel(ctx)

	p.cancelLocked = func() {
		__antithesis_instrumentation__.Notify(119604)
		cancel()
		p.cancelLocked = nil
		p.nextLease = roachpb.Lease{}
	}
	__antithesis_instrumentation__.Notify(119599)

	err := p.repl.store.Stopper().RunAsyncTaskEx(
		ctx,
		stop.TaskOpts{
			TaskName: "pendingLeaseRequest: requesting lease",

			SpanOpt: stop.ChildSpan,
		},
		func(ctx context.Context) {
			__antithesis_instrumentation__.Notify(119605)
			defer sp.Finish()

			var pErr *roachpb.Error
			if reqLease.Type() == roachpb.LeaseEpoch && func() bool {
				__antithesis_instrumentation__.Notify(119610)
				return status.State == kvserverpb.LeaseState_EXPIRED == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(119611)
				return status.Lease.Type() == roachpb.LeaseEpoch == true
			}() == true {
				__antithesis_instrumentation__.Notify(119612)
				var err error

				if status.OwnedBy(nextLeaseHolder.StoreID) && func() bool {
					__antithesis_instrumentation__.Notify(119614)
					return p.repl.store.StoreID() == nextLeaseHolder.StoreID == true
				}() == true {
					__antithesis_instrumentation__.Notify(119615)
					if err = p.repl.store.cfg.NodeLiveness.Heartbeat(ctx, status.Liveness); err != nil {
						__antithesis_instrumentation__.Notify(119616)
						log.Errorf(ctx, "failed to heartbeat own liveness record: %s", err)
					} else {
						__antithesis_instrumentation__.Notify(119617)
					}
				} else {
					__antithesis_instrumentation__.Notify(119618)
					if status.Liveness.Epoch == status.Lease.Epoch {
						__antithesis_instrumentation__.Notify(119619)

						if live, liveErr := p.repl.store.cfg.NodeLiveness.IsLive(nextLeaseHolder.NodeID); !live || func() bool {
							__antithesis_instrumentation__.Notify(119620)
							return liveErr != nil == true
						}() == true {
							__antithesis_instrumentation__.Notify(119621)
							if liveErr != nil {
								__antithesis_instrumentation__.Notify(119623)
								err = errors.Wrapf(liveErr, "not incrementing epoch on n%d because next leaseholder (n%d) not live",
									status.Liveness.NodeID, nextLeaseHolder.NodeID)
							} else {
								__antithesis_instrumentation__.Notify(119624)
								err = errors.Errorf("not incrementing epoch on n%d because next leaseholder (n%d) not live (err = nil)",
									status.Liveness.NodeID, nextLeaseHolder.NodeID)
							}
							__antithesis_instrumentation__.Notify(119622)
							log.VEventf(ctx, 1, "%v", err)
						} else {
							__antithesis_instrumentation__.Notify(119625)
							if err = p.repl.store.cfg.NodeLiveness.IncrementEpoch(ctx, status.Liveness); err != nil {
								__antithesis_instrumentation__.Notify(119626)

								if !errors.Is(err, liveness.ErrEpochAlreadyIncremented) {
									__antithesis_instrumentation__.Notify(119627)
									log.Errorf(ctx, "failed to increment leaseholder's epoch: %s", err)
								} else {
									__antithesis_instrumentation__.Notify(119628)
								}
							} else {
								__antithesis_instrumentation__.Notify(119629)
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(119630)
					}
				}
				__antithesis_instrumentation__.Notify(119613)

				if err != nil {
					__antithesis_instrumentation__.Notify(119631)

					pErr = roachpb.NewError(newNotLeaseHolderError(
						status.Lease, p.repl.store.StoreID(), p.repl.Desc(),
						fmt.Sprintf("failed to manipulate liveness record: %s", err)))
				} else {
					__antithesis_instrumentation__.Notify(119632)
				}
			} else {
				__antithesis_instrumentation__.Notify(119633)
			}
			__antithesis_instrumentation__.Notify(119606)

			if pErr == nil {
				__antithesis_instrumentation__.Notify(119634)

				ba := roachpb.BatchRequest{}
				ba.Timestamp = p.repl.store.Clock().Now()
				ba.RangeID = p.repl.RangeID

				ba.Add(leaseReq)
				_, pErr = p.repl.Send(ctx, ba)
			} else {
				__antithesis_instrumentation__.Notify(119635)
			}
			__antithesis_instrumentation__.Notify(119607)

			p.repl.mu.Lock()
			defer p.repl.mu.Unlock()
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(119636)

				return
			} else {
				__antithesis_instrumentation__.Notify(119637)
			}
			__antithesis_instrumentation__.Notify(119608)

			for llHandle := range p.llHandles {
				__antithesis_instrumentation__.Notify(119638)

				if pErr != nil {
					__antithesis_instrumentation__.Notify(119640)
					pErrClone := *pErr

					pErrClone.SetTxn(pErr.GetTxn())
					llHandle.resolve(&pErrClone)
				} else {
					__antithesis_instrumentation__.Notify(119641)
					llHandle.resolve(nil)
				}
				__antithesis_instrumentation__.Notify(119639)
				delete(p.llHandles, llHandle)
			}
			__antithesis_instrumentation__.Notify(119609)
			p.cancelLocked()
		})
	__antithesis_instrumentation__.Notify(119600)
	if err != nil {
		__antithesis_instrumentation__.Notify(119642)
		p.cancelLocked()
		sp.Finish()
		return err
	} else {
		__antithesis_instrumentation__.Notify(119643)
	}
	__antithesis_instrumentation__.Notify(119601)
	return nil
}

func (p *pendingLeaseRequest) JoinRequest() *leaseRequestHandle {
	__antithesis_instrumentation__.Notify(119644)
	llHandle := p.newHandle()
	if _, ok := p.RequestPending(); !ok {
		__antithesis_instrumentation__.Notify(119646)
		llHandle.resolve(roachpb.NewErrorf("no request in progress"))
		return llHandle
	} else {
		__antithesis_instrumentation__.Notify(119647)
	}
	__antithesis_instrumentation__.Notify(119645)
	p.llHandles[llHandle] = struct{}{}
	return llHandle
}

func (p *pendingLeaseRequest) TransferInProgress(replicaID roachpb.ReplicaID) bool {
	__antithesis_instrumentation__.Notify(119648)
	if nextLease, ok := p.RequestPending(); ok {
		__antithesis_instrumentation__.Notify(119650)

		return replicaID != nextLease.Replica.ReplicaID
	} else {
		__antithesis_instrumentation__.Notify(119651)
	}
	__antithesis_instrumentation__.Notify(119649)
	return false
}

func (p *pendingLeaseRequest) newHandle() *leaseRequestHandle {
	__antithesis_instrumentation__.Notify(119652)
	return &leaseRequestHandle{
		p: p,
		c: make(chan *roachpb.Error, 1),
	}
}

func (p *pendingLeaseRequest) newResolvedHandle(pErr *roachpb.Error) *leaseRequestHandle {
	__antithesis_instrumentation__.Notify(119653)
	h := p.newHandle()
	h.resolve(pErr)
	return h
}

func (r *Replica) leaseStatus(
	ctx context.Context,
	lease roachpb.Lease,
	now hlc.ClockTimestamp,
	minProposedTS hlc.ClockTimestamp,
	reqTS hlc.Timestamp,
) kvserverpb.LeaseStatus {
	__antithesis_instrumentation__.Notify(119654)
	status := kvserverpb.LeaseStatus{
		Lease: lease,

		Now:         now,
		RequestTime: reqTS,
	}
	var expiration hlc.Timestamp
	if lease.Type() == roachpb.LeaseExpiration {
		__antithesis_instrumentation__.Notify(119657)
		expiration = lease.GetExpiration()
	} else {
		__antithesis_instrumentation__.Notify(119658)
		l, ok := r.store.cfg.NodeLiveness.GetLiveness(lease.Replica.NodeID)
		status.Liveness = l.Liveness
		if !ok || func() bool {
			__antithesis_instrumentation__.Notify(119661)
			return status.Liveness.Epoch < lease.Epoch == true
		}() == true {
			__antithesis_instrumentation__.Notify(119662)

			var msg redact.StringBuilder
			if !ok {
				__antithesis_instrumentation__.Notify(119665)
				msg.Printf("can't determine lease status of %s due to node liveness error: %v",
					lease.Replica, liveness.ErrRecordCacheMiss)
			} else {
				__antithesis_instrumentation__.Notify(119666)
				msg.Printf("can't determine lease status of %s because node liveness info for n%d is stale. lease: %s, liveness: %s",
					lease.Replica, lease.Replica.NodeID, lease, l.Liveness)
			}
			__antithesis_instrumentation__.Notify(119663)
			if leaseStatusLogLimiter.ShouldLog() {
				__antithesis_instrumentation__.Notify(119667)
				log.Infof(ctx, "%s", msg)
			} else {
				__antithesis_instrumentation__.Notify(119668)
			}
			__antithesis_instrumentation__.Notify(119664)
			status.State = kvserverpb.LeaseState_ERROR
			status.ErrInfo = msg.String()
			return status
		} else {
			__antithesis_instrumentation__.Notify(119669)
		}
		__antithesis_instrumentation__.Notify(119659)
		if status.Liveness.Epoch > lease.Epoch {
			__antithesis_instrumentation__.Notify(119670)
			status.State = kvserverpb.LeaseState_EXPIRED
			return status
		} else {
			__antithesis_instrumentation__.Notify(119671)
		}
		__antithesis_instrumentation__.Notify(119660)
		expiration = status.Liveness.Expiration.ToTimestamp()
	}
	__antithesis_instrumentation__.Notify(119655)
	maxOffset := r.store.Clock().MaxOffset()
	stasis := expiration.Add(-int64(maxOffset), 0)
	ownedLocally := lease.OwnedBy(r.store.StoreID())
	if expiration.LessEq(now.ToTimestamp()) {
		__antithesis_instrumentation__.Notify(119672)
		status.State = kvserverpb.LeaseState_EXPIRED
	} else {
		__antithesis_instrumentation__.Notify(119673)
		if stasis.LessEq(reqTS) {
			__antithesis_instrumentation__.Notify(119674)
			status.State = kvserverpb.LeaseState_UNUSABLE
		} else {
			__antithesis_instrumentation__.Notify(119675)
			if ownedLocally && func() bool {
				__antithesis_instrumentation__.Notify(119676)
				return lease.ProposedTS != nil == true
			}() == true && func() bool {
				__antithesis_instrumentation__.Notify(119677)
				return lease.ProposedTS.Less(minProposedTS) == true
			}() == true {
				__antithesis_instrumentation__.Notify(119678)

				status.State = kvserverpb.LeaseState_PROSCRIBED
			} else {
				__antithesis_instrumentation__.Notify(119679)
				status.State = kvserverpb.LeaseState_VALID
			}
		}
	}
	__antithesis_instrumentation__.Notify(119656)
	return status
}

func (r *Replica) CurrentLeaseStatus(ctx context.Context) kvserverpb.LeaseStatus {
	__antithesis_instrumentation__.Notify(119680)
	return r.LeaseStatusAt(ctx, r.Clock().NowAsClockTimestamp())
}

func (r *Replica) LeaseStatusAt(
	ctx context.Context, now hlc.ClockTimestamp,
) kvserverpb.LeaseStatus {
	__antithesis_instrumentation__.Notify(119681)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.leaseStatusAtRLocked(ctx, now)
}

func (r *Replica) leaseStatusAtRLocked(
	ctx context.Context, now hlc.ClockTimestamp,
) kvserverpb.LeaseStatus {
	__antithesis_instrumentation__.Notify(119682)
	return r.leaseStatusForRequestRLocked(ctx, now, hlc.Timestamp{})
}

func (r *Replica) leaseStatusForRequestRLocked(
	ctx context.Context, now hlc.ClockTimestamp, reqTS hlc.Timestamp,
) kvserverpb.LeaseStatus {
	__antithesis_instrumentation__.Notify(119683)
	if reqTS.IsEmpty() {
		__antithesis_instrumentation__.Notify(119685)

		reqTS = now.ToTimestamp()
	} else {
		__antithesis_instrumentation__.Notify(119686)
	}
	__antithesis_instrumentation__.Notify(119684)
	return r.leaseStatus(ctx, *r.mu.state.Lease, now, r.mu.minLeaseProposedTS, reqTS)
}

func (r *Replica) OwnsValidLease(ctx context.Context, now hlc.ClockTimestamp) bool {
	__antithesis_instrumentation__.Notify(119687)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.ownsValidLeaseRLocked(ctx, now)
}

func (r *Replica) ownsValidLeaseRLocked(ctx context.Context, now hlc.ClockTimestamp) bool {
	__antithesis_instrumentation__.Notify(119688)
	st := r.leaseStatusAtRLocked(ctx, now)
	return st.IsValid() && func() bool {
		__antithesis_instrumentation__.Notify(119689)
		return st.OwnedBy(r.store.StoreID()) == true
	}() == true
}

func (r *Replica) requiresExpiringLeaseRLocked() bool {
	__antithesis_instrumentation__.Notify(119690)
	return r.store.cfg.NodeLiveness == nil || func() bool {
		__antithesis_instrumentation__.Notify(119691)
		return r.mu.state.Desc.StartKey.Less(roachpb.RKey(keys.NodeLivenessKeyMax)) == true
	}() == true
}

func (r *Replica) requestLeaseLocked(
	ctx context.Context, status kvserverpb.LeaseStatus,
) *leaseRequestHandle {
	__antithesis_instrumentation__.Notify(119692)
	if r.store.TestingKnobs().LeaseRequestEvent != nil {
		__antithesis_instrumentation__.Notify(119697)
		if err := r.store.TestingKnobs().LeaseRequestEvent(status.Now.ToTimestamp(), r.StoreID(), r.GetRangeID()); err != nil {
			__antithesis_instrumentation__.Notify(119698)
			return r.mu.pendingLeaseRequest.newResolvedHandle(err)
		} else {
			__antithesis_instrumentation__.Notify(119699)
		}
	} else {
		__antithesis_instrumentation__.Notify(119700)
	}
	__antithesis_instrumentation__.Notify(119693)
	if pErr := r.store.TestingKnobs().PinnedLeases.rejectLeaseIfPinnedElsewhere(r); pErr != nil {
		__antithesis_instrumentation__.Notify(119701)
		return r.mu.pendingLeaseRequest.newResolvedHandle(pErr)
	} else {
		__antithesis_instrumentation__.Notify(119702)
	}
	__antithesis_instrumentation__.Notify(119694)

	if r.store.IsDraining() {
		__antithesis_instrumentation__.Notify(119703)

		if r.raftBasicStatusRLocked().RaftState != raft.StateLeader {
			__antithesis_instrumentation__.Notify(119705)
			log.VEventf(ctx, 2, "refusing to take the lease because we're draining")
			return r.mu.pendingLeaseRequest.newResolvedHandle(
				roachpb.NewError(
					newNotLeaseHolderError(
						roachpb.Lease{}, r.store.StoreID(), r.mu.state.Desc,
						"refusing to take the lease; node is draining",
					),
				),
			)
		} else {
			__antithesis_instrumentation__.Notify(119706)
		}
		__antithesis_instrumentation__.Notify(119704)
		log.Info(ctx, "trying to take the lease while we're draining since we're the raft leader")
	} else {
		__antithesis_instrumentation__.Notify(119707)
	}
	__antithesis_instrumentation__.Notify(119695)

	repDesc, err := r.getReplicaDescriptorRLocked()
	if err != nil {
		__antithesis_instrumentation__.Notify(119708)
		return r.mu.pendingLeaseRequest.newResolvedHandle(roachpb.NewError(err))
	} else {
		__antithesis_instrumentation__.Notify(119709)
	}
	__antithesis_instrumentation__.Notify(119696)
	return r.mu.pendingLeaseRequest.InitOrJoinRequest(
		ctx, repDesc, status, r.mu.state.Desc.StartKey.AsRawKey(), false)
}

func (r *Replica) AdminTransferLease(ctx context.Context, target roachpb.StoreID) error {
	__antithesis_instrumentation__.Notify(119710)

	var nextLeaseHolder roachpb.ReplicaDescriptor
	initTransferHelper := func() (extension, transfer *leaseRequestHandle, err error) {
		__antithesis_instrumentation__.Notify(119712)
		r.mu.Lock()
		defer r.mu.Unlock()

		now := r.store.Clock().NowAsClockTimestamp()
		status := r.leaseStatusAtRLocked(ctx, now)
		if status.Lease.OwnedBy(target) {
			__antithesis_instrumentation__.Notify(119717)

			return nil, nil, nil
		} else {
			__antithesis_instrumentation__.Notify(119718)
		}
		__antithesis_instrumentation__.Notify(119713)
		desc := r.mu.state.Desc
		if !status.Lease.OwnedBy(r.store.StoreID()) {
			__antithesis_instrumentation__.Notify(119719)
			return nil, nil, newNotLeaseHolderError(status.Lease, r.store.StoreID(), desc,
				"can't transfer the lease because this store doesn't own it")
		} else {
			__antithesis_instrumentation__.Notify(119720)
		}
		__antithesis_instrumentation__.Notify(119714)

		var ok bool
		if nextLeaseHolder, ok = desc.GetReplicaDescriptor(target); !ok {
			__antithesis_instrumentation__.Notify(119721)
			return nil, nil, errors.Errorf("unable to find store %d in range %+v", target, desc)
		} else {
			__antithesis_instrumentation__.Notify(119722)
		}
		__antithesis_instrumentation__.Notify(119715)

		if nextLease, ok := r.mu.pendingLeaseRequest.RequestPending(); ok && func() bool {
			__antithesis_instrumentation__.Notify(119723)
			return nextLease.Replica != nextLeaseHolder == true
		}() == true {
			__antithesis_instrumentation__.Notify(119724)
			repDesc, err := r.getReplicaDescriptorRLocked()
			if err != nil {
				__antithesis_instrumentation__.Notify(119727)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(119728)
			}
			__antithesis_instrumentation__.Notify(119725)
			if nextLease.Replica == repDesc {
				__antithesis_instrumentation__.Notify(119729)

				return r.mu.pendingLeaseRequest.JoinRequest(), nil, nil
			} else {
				__antithesis_instrumentation__.Notify(119730)
			}
			__antithesis_instrumentation__.Notify(119726)

			return nil, nil, newNotLeaseHolderError(nextLease, r.store.StoreID(), desc,
				"another transfer to a different store is in progress")
		} else {
			__antithesis_instrumentation__.Notify(119731)
		}
		__antithesis_instrumentation__.Notify(119716)

		transfer = r.mu.pendingLeaseRequest.InitOrJoinRequest(
			ctx, nextLeaseHolder, status, desc.StartKey.AsRawKey(), true,
		)
		return nil, transfer, nil
	}
	__antithesis_instrumentation__.Notify(119711)

	for {
		__antithesis_instrumentation__.Notify(119732)

		extension, transfer, err := initTransferHelper()
		if err != nil {
			__antithesis_instrumentation__.Notify(119736)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119737)
		}
		__antithesis_instrumentation__.Notify(119733)
		if extension == nil {
			__antithesis_instrumentation__.Notify(119738)
			if transfer == nil {
				__antithesis_instrumentation__.Notify(119740)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(119741)
			}
			__antithesis_instrumentation__.Notify(119739)
			select {
			case pErr := <-transfer.C():
				__antithesis_instrumentation__.Notify(119742)
				return pErr.GoError()
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(119743)
				transfer.Cancel()
				return ctx.Err()
			}
		} else {
			__antithesis_instrumentation__.Notify(119744)
		}
		__antithesis_instrumentation__.Notify(119734)

		if r.store.TestingKnobs().LeaseTransferBlockedOnExtensionEvent != nil {
			__antithesis_instrumentation__.Notify(119745)
			r.store.TestingKnobs().LeaseTransferBlockedOnExtensionEvent(nextLeaseHolder)
		} else {
			__antithesis_instrumentation__.Notify(119746)
		}
		__antithesis_instrumentation__.Notify(119735)
		select {
		case <-extension.C():
			__antithesis_instrumentation__.Notify(119747)
			continue
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(119748)
			extension.Cancel()
			return ctx.Err()
		}
	}
}

func (r *Replica) GetLease() (roachpb.Lease, roachpb.Lease) {
	__antithesis_instrumentation__.Notify(119749)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.getLeaseRLocked()
}

func (r *Replica) getLeaseRLocked() (roachpb.Lease, roachpb.Lease) {
	__antithesis_instrumentation__.Notify(119750)
	if nextLease, ok := r.mu.pendingLeaseRequest.RequestPending(); ok {
		__antithesis_instrumentation__.Notify(119752)
		return *r.mu.state.Lease, nextLease
	} else {
		__antithesis_instrumentation__.Notify(119753)
	}
	__antithesis_instrumentation__.Notify(119751)
	return *r.mu.state.Lease, roachpb.Lease{}
}

func (r *Replica) RevokeLease(ctx context.Context, seq roachpb.LeaseSequence) {
	__antithesis_instrumentation__.Notify(119754)
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.mu.state.Lease.Sequence == seq {
		__antithesis_instrumentation__.Notify(119755)
		r.mu.minLeaseProposedTS = r.Clock().NowAsClockTimestamp()
	} else {
		__antithesis_instrumentation__.Notify(119756)
	}
}

func newNotLeaseHolderError(
	l roachpb.Lease, proposerStoreID roachpb.StoreID, rangeDesc *roachpb.RangeDescriptor, msg string,
) *roachpb.NotLeaseHolderError {
	__antithesis_instrumentation__.Notify(119757)
	err := &roachpb.NotLeaseHolderError{
		RangeID:   rangeDesc.RangeID,
		RangeDesc: *rangeDesc,
		CustomMsg: msg,
	}
	if proposerStoreID != 0 {
		__antithesis_instrumentation__.Notify(119760)
		err.Replica, _ = rangeDesc.GetReplicaDescriptor(proposerStoreID)
	} else {
		__antithesis_instrumentation__.Notify(119761)
	}
	__antithesis_instrumentation__.Notify(119758)
	if !l.Empty() {
		__antithesis_instrumentation__.Notify(119762)

		_, stillMember := rangeDesc.GetReplicaDescriptor(l.Replica.StoreID)
		if stillMember {
			__antithesis_instrumentation__.Notify(119763)
			err.Lease = new(roachpb.Lease)
			*err.Lease = l
			err.LeaseHolder = &err.Lease.Replica
		} else {
			__antithesis_instrumentation__.Notify(119764)
		}
	} else {
		__antithesis_instrumentation__.Notify(119765)
	}
	__antithesis_instrumentation__.Notify(119759)
	return err
}

func (r *Replica) checkRequestTimeRLocked(now hlc.ClockTimestamp, reqTS hlc.Timestamp) error {
	__antithesis_instrumentation__.Notify(119766)
	var leaseRenewal time.Duration
	if r.requiresExpiringLeaseRLocked() {
		__antithesis_instrumentation__.Notify(119770)
		_, leaseRenewal = r.store.cfg.RangeLeaseDurations()
	} else {
		__antithesis_instrumentation__.Notify(119771)
		_, leaseRenewal = r.store.cfg.NodeLivenessDurations()
	}
	__antithesis_instrumentation__.Notify(119767)
	leaseRenewalMinusStasis := leaseRenewal - r.store.Clock().MaxOffset()
	if leaseRenewalMinusStasis < 0 {
		__antithesis_instrumentation__.Notify(119772)

		leaseRenewalMinusStasis = 0
	} else {
		__antithesis_instrumentation__.Notify(119773)
	}
	__antithesis_instrumentation__.Notify(119768)
	maxReqTS := now.ToTimestamp().Add(leaseRenewalMinusStasis.Nanoseconds(), 0)
	if maxReqTS.Less(reqTS) {
		__antithesis_instrumentation__.Notify(119774)
		return errors.Errorf("request timestamp %s too far in future (> %s)", reqTS, maxReqTS)
	} else {
		__antithesis_instrumentation__.Notify(119775)
	}
	__antithesis_instrumentation__.Notify(119769)
	return nil
}

func (r *Replica) leaseGoodToGoRLocked(
	ctx context.Context, now hlc.ClockTimestamp, reqTS hlc.Timestamp,
) (_ kvserverpb.LeaseStatus, shouldExtend bool, _ error) {
	__antithesis_instrumentation__.Notify(119776)
	st := r.leaseStatusForRequestRLocked(ctx, now, reqTS)
	shouldExtend, err := r.leaseGoodToGoForStatusRLocked(ctx, now, reqTS, st)
	if err != nil {
		__antithesis_instrumentation__.Notify(119778)
		return kvserverpb.LeaseStatus{}, false, err
	} else {
		__antithesis_instrumentation__.Notify(119779)
	}
	__antithesis_instrumentation__.Notify(119777)
	return st, shouldExtend, err
}

func (r *Replica) leaseGoodToGoForStatusRLocked(
	ctx context.Context, now hlc.ClockTimestamp, reqTS hlc.Timestamp, st kvserverpb.LeaseStatus,
) (shouldExtend bool, _ error) {
	__antithesis_instrumentation__.Notify(119780)
	if err := r.checkRequestTimeRLocked(now, reqTS); err != nil {
		__antithesis_instrumentation__.Notify(119784)

		return false, err
	} else {
		__antithesis_instrumentation__.Notify(119785)
	}
	__antithesis_instrumentation__.Notify(119781)
	if !st.IsValid() {
		__antithesis_instrumentation__.Notify(119786)

		return false, &roachpb.InvalidLeaseError{}
	} else {
		__antithesis_instrumentation__.Notify(119787)
	}
	__antithesis_instrumentation__.Notify(119782)
	if !st.Lease.OwnedBy(r.store.StoreID()) {
		__antithesis_instrumentation__.Notify(119788)

		_, stillMember := r.mu.state.Desc.GetReplicaDescriptor(st.Lease.Replica.StoreID)
		if !stillMember {
			__antithesis_instrumentation__.Notify(119790)

			log.Errorf(ctx, "lease %s owned by replica %+v that no longer exists",
				st.Lease, st.Lease.Replica)
		} else {
			__antithesis_instrumentation__.Notify(119791)
		}
		__antithesis_instrumentation__.Notify(119789)

		return false, newNotLeaseHolderError(
			st.Lease, r.store.StoreID(), r.descRLocked(), "lease held by different store",
		)
	} else {
		__antithesis_instrumentation__.Notify(119792)
	}
	__antithesis_instrumentation__.Notify(119783)

	return r.shouldExtendLeaseRLocked(st), nil
}

func (r *Replica) leaseGoodToGo(
	ctx context.Context, now hlc.ClockTimestamp, reqTS hlc.Timestamp,
) (_ kvserverpb.LeaseStatus, shouldExtend bool, _ error) {
	__antithesis_instrumentation__.Notify(119793)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.leaseGoodToGoRLocked(ctx, now, reqTS)
}

func (r *Replica) redirectOnOrAcquireLease(
	ctx context.Context,
) (kvserverpb.LeaseStatus, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(119794)
	return r.redirectOnOrAcquireLeaseForRequest(ctx, hlc.Timestamp{}, r.breaker.Signal())
}

func (r *Replica) TestingAcquireLease(ctx context.Context) (kvserverpb.LeaseStatus, error) {
	__antithesis_instrumentation__.Notify(119795)
	ctx = r.AnnotateCtx(ctx)
	ctx = logtags.AddTag(ctx, "lease-acq", nil)
	l, pErr := r.redirectOnOrAcquireLease(ctx)
	return l, pErr.GoError()
}

func (r *Replica) redirectOnOrAcquireLeaseForRequest(
	ctx context.Context, reqTS hlc.Timestamp, brSig signaller,
) (kvserverpb.LeaseStatus, *roachpb.Error) {
	__antithesis_instrumentation__.Notify(119796)

	now := r.store.Clock().NowAsClockTimestamp()
	{
		__antithesis_instrumentation__.Notify(119799)
		status, shouldExtend, err := r.leaseGoodToGo(ctx, now, reqTS)
		if err == nil {
			__antithesis_instrumentation__.Notify(119800)
			if shouldExtend {
				__antithesis_instrumentation__.Notify(119802)
				r.maybeExtendLeaseAsync(ctx, status)
			} else {
				__antithesis_instrumentation__.Notify(119803)
			}
			__antithesis_instrumentation__.Notify(119801)
			return status, nil
		} else {
			__antithesis_instrumentation__.Notify(119804)
			if !errors.HasType(err, (*roachpb.InvalidLeaseError)(nil)) {
				__antithesis_instrumentation__.Notify(119805)
				return kvserverpb.LeaseStatus{}, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(119806)
			}
		}
	}
	__antithesis_instrumentation__.Notify(119797)

	if err := brSig.Err(); err != nil {
		__antithesis_instrumentation__.Notify(119807)
		return kvserverpb.LeaseStatus{}, roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(119808)
	}
	__antithesis_instrumentation__.Notify(119798)

	for attempt := 1; ; attempt++ {
		__antithesis_instrumentation__.Notify(119809)
		now = r.store.Clock().NowAsClockTimestamp()
		llHandle, status, transfer, pErr := func() (*leaseRequestHandle, kvserverpb.LeaseStatus, bool, *roachpb.Error) {
			__antithesis_instrumentation__.Notify(119814)
			r.mu.Lock()
			defer r.mu.Unlock()

			repDesc, err := r.getReplicaDescriptorRLocked()
			if err != nil {
				__antithesis_instrumentation__.Notify(119817)
				return nil, kvserverpb.LeaseStatus{}, false, roachpb.NewError(err)
			} else {
				__antithesis_instrumentation__.Notify(119818)
			}
			__antithesis_instrumentation__.Notify(119815)
			if ok := r.mu.pendingLeaseRequest.TransferInProgress(repDesc.ReplicaID); ok {
				__antithesis_instrumentation__.Notify(119819)
				return r.mu.pendingLeaseRequest.JoinRequest(), kvserverpb.LeaseStatus{}, true, nil
			} else {
				__antithesis_instrumentation__.Notify(119820)
			}
			__antithesis_instrumentation__.Notify(119816)

			status := r.leaseStatusForRequestRLocked(ctx, now, reqTS)
			switch status.State {
			case kvserverpb.LeaseState_ERROR:
				__antithesis_instrumentation__.Notify(119821)

				msg := status.ErrInfo
				if msg == "" {
					__antithesis_instrumentation__.Notify(119830)
					msg = "lease state could not be determined"
				} else {
					__antithesis_instrumentation__.Notify(119831)
				}
				__antithesis_instrumentation__.Notify(119822)
				log.VEventf(ctx, 2, "%s", msg)
				return nil, kvserverpb.LeaseStatus{}, false, roachpb.NewError(
					newNotLeaseHolderError(roachpb.Lease{}, r.store.StoreID(), r.mu.state.Desc, msg))

			case kvserverpb.LeaseState_VALID, kvserverpb.LeaseState_UNUSABLE:
				__antithesis_instrumentation__.Notify(119823)
				if !status.Lease.OwnedBy(r.store.StoreID()) {
					__antithesis_instrumentation__.Notify(119832)
					_, stillMember := r.mu.state.Desc.GetReplicaDescriptor(status.Lease.Replica.StoreID)
					if !stillMember {
						__antithesis_instrumentation__.Notify(119834)

						log.Errorf(ctx, "lease %s owned by replica %+v that no longer exists",
							status.Lease, status.Lease.Replica)
					} else {
						__antithesis_instrumentation__.Notify(119835)
					}
					__antithesis_instrumentation__.Notify(119833)

					return nil, kvserverpb.LeaseStatus{}, false, roachpb.NewError(
						newNotLeaseHolderError(status.Lease, r.store.StoreID(), r.mu.state.Desc,
							"lease held by different store"))
				} else {
					__antithesis_instrumentation__.Notify(119836)
				}
				__antithesis_instrumentation__.Notify(119824)

				if status.State == kvserverpb.LeaseState_UNUSABLE {
					__antithesis_instrumentation__.Notify(119837)
					return r.requestLeaseLocked(ctx, status), kvserverpb.LeaseStatus{}, false, nil
				} else {
					__antithesis_instrumentation__.Notify(119838)
				}
				__antithesis_instrumentation__.Notify(119825)

				return nil, status, false, nil

			case kvserverpb.LeaseState_EXPIRED:
				__antithesis_instrumentation__.Notify(119826)

				log.VEventf(ctx, 2, "request range lease (attempt #%d)", attempt)
				return r.requestLeaseLocked(ctx, status), kvserverpb.LeaseStatus{}, false, nil

			case kvserverpb.LeaseState_PROSCRIBED:
				__antithesis_instrumentation__.Notify(119827)

				if status.Lease.OwnedBy(r.store.StoreID()) {
					__antithesis_instrumentation__.Notify(119839)
					log.VEventf(ctx, 2, "request range lease (attempt #%d)", attempt)
					return r.requestLeaseLocked(ctx, status), kvserverpb.LeaseStatus{}, false, nil
				} else {
					__antithesis_instrumentation__.Notify(119840)
				}
				__antithesis_instrumentation__.Notify(119828)

				return nil, kvserverpb.LeaseStatus{}, false, roachpb.NewError(
					newNotLeaseHolderError(status.Lease, r.store.StoreID(), r.mu.state.Desc, "lease proscribed"))

			default:
				__antithesis_instrumentation__.Notify(119829)
				return nil, kvserverpb.LeaseStatus{}, false, roachpb.NewErrorf("unknown lease status state %v", status)
			}
		}()
		__antithesis_instrumentation__.Notify(119810)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(119841)
			return kvserverpb.LeaseStatus{}, pErr
		} else {
			__antithesis_instrumentation__.Notify(119842)
		}
		__antithesis_instrumentation__.Notify(119811)
		if llHandle == nil {
			__antithesis_instrumentation__.Notify(119843)

			log.Eventf(ctx, "valid lease %+v", status)
			return status, nil
		} else {
			__antithesis_instrumentation__.Notify(119844)
		}
		__antithesis_instrumentation__.Notify(119812)

		pErr = func() (pErr *roachpb.Error) {
			__antithesis_instrumentation__.Notify(119845)
			slowTimer := timeutil.NewTimer()
			defer slowTimer.Stop()
			slowTimer.Reset(base.SlowRequestThreshold)
			tBegin := timeutil.Now()
			for {
				__antithesis_instrumentation__.Notify(119846)
				select {
				case pErr = <-llHandle.C():
					__antithesis_instrumentation__.Notify(119847)
					if transfer {
						__antithesis_instrumentation__.Notify(119854)

						return nil
					} else {
						__antithesis_instrumentation__.Notify(119855)
					}
					__antithesis_instrumentation__.Notify(119848)

					if pErr != nil {
						__antithesis_instrumentation__.Notify(119856)
						goErr := pErr.GoError()
						switch {
						case errors.HasType(goErr, (*roachpb.AmbiguousResultError)(nil)):
							__antithesis_instrumentation__.Notify(119858)

							return nil
						case errors.HasType(goErr, (*roachpb.LeaseRejectedError)(nil)):
							__antithesis_instrumentation__.Notify(119859)
							var tErr *roachpb.LeaseRejectedError
							errors.As(goErr, &tErr)
							if tErr.Existing.OwnedBy(r.store.StoreID()) {
								__antithesis_instrumentation__.Notify(119863)

								return nil
							} else {
								__antithesis_instrumentation__.Notify(119864)
							}
							__antithesis_instrumentation__.Notify(119860)

							var err error
							if _, descErr := r.GetReplicaDescriptor(); descErr != nil {
								__antithesis_instrumentation__.Notify(119865)
								err = descErr
							} else {
								__antithesis_instrumentation__.Notify(119866)
								if st := r.CurrentLeaseStatus(ctx); !st.IsValid() {
									__antithesis_instrumentation__.Notify(119867)
									err = newNotLeaseHolderError(roachpb.Lease{}, r.store.StoreID(), r.Desc(),
										"lease acquisition attempt lost to another lease, which has expired in the meantime")
								} else {
									__antithesis_instrumentation__.Notify(119868)
									err = newNotLeaseHolderError(st.Lease, r.store.StoreID(), r.Desc(),
										"lease acquisition attempt lost to another lease")
								}
							}
							__antithesis_instrumentation__.Notify(119861)
							pErr = roachpb.NewError(err)
						default:
							__antithesis_instrumentation__.Notify(119862)
						}
						__antithesis_instrumentation__.Notify(119857)
						return pErr
					} else {
						__antithesis_instrumentation__.Notify(119869)
					}
					__antithesis_instrumentation__.Notify(119849)
					log.VEventf(ctx, 2, "lease acquisition succeeded: %+v", status.Lease)
					return nil
				case <-brSig.C():
					__antithesis_instrumentation__.Notify(119850)
					llHandle.Cancel()
					err := brSig.Err()
					log.VErrEventf(ctx, 2, "lease acquisition failed: %s", err)
					return roachpb.NewError(err)
				case <-slowTimer.C:
					__antithesis_instrumentation__.Notify(119851)
					slowTimer.Read = true
					log.Warningf(ctx, "have been waiting %s attempting to acquire lease (%d attempts)",
						base.SlowRequestThreshold, attempt)
					r.store.metrics.SlowLeaseRequests.Inc(1)
					defer func() {
						__antithesis_instrumentation__.Notify(119870)
						r.store.metrics.SlowLeaseRequests.Dec(1)
						log.Infof(ctx, "slow lease acquisition finished after %s with error %v after %d attempts", timeutil.Since(tBegin), pErr, attempt)
					}()
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(119852)
					llHandle.Cancel()
					log.VErrEventf(ctx, 2, "lease acquisition failed: %s", ctx.Err())
					return roachpb.NewError(newNotLeaseHolderError(roachpb.Lease{}, r.store.StoreID(), r.Desc(),
						"lease acquisition canceled because context canceled"))
				case <-r.store.Stopper().ShouldQuiesce():
					__antithesis_instrumentation__.Notify(119853)
					llHandle.Cancel()
					return roachpb.NewError(newNotLeaseHolderError(roachpb.Lease{}, r.store.StoreID(), r.Desc(),
						"lease acquisition canceled because node is stopping"))
				}
			}
		}()
		__antithesis_instrumentation__.Notify(119813)
		if pErr != nil {
			__antithesis_instrumentation__.Notify(119871)
			return kvserverpb.LeaseStatus{}, pErr
		} else {
			__antithesis_instrumentation__.Notify(119872)
		}

	}
}

func (r *Replica) shouldExtendLeaseRLocked(st kvserverpb.LeaseStatus) bool {
	__antithesis_instrumentation__.Notify(119873)
	if !r.requiresExpiringLeaseRLocked() {
		__antithesis_instrumentation__.Notify(119876)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119877)
	}
	__antithesis_instrumentation__.Notify(119874)
	if _, ok := r.mu.pendingLeaseRequest.RequestPending(); ok {
		__antithesis_instrumentation__.Notify(119878)
		return false
	} else {
		__antithesis_instrumentation__.Notify(119879)
	}
	__antithesis_instrumentation__.Notify(119875)
	renewal := st.Lease.Expiration.Add(-r.store.cfg.RangeLeaseRenewalDuration().Nanoseconds(), 0)
	return renewal.LessEq(st.Now.ToTimestamp())
}

func (r *Replica) maybeExtendLeaseAsync(ctx context.Context, st kvserverpb.LeaseStatus) {
	__antithesis_instrumentation__.Notify(119880)
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.shouldExtendLeaseRLocked(st) {
		__antithesis_instrumentation__.Notify(119883)
		return
	} else {
		__antithesis_instrumentation__.Notify(119884)
	}
	__antithesis_instrumentation__.Notify(119881)
	if log.ExpensiveLogEnabled(ctx, 2) {
		__antithesis_instrumentation__.Notify(119885)
		log.Infof(ctx, "extending lease %s at %s", st.Lease, st.Now)
	} else {
		__antithesis_instrumentation__.Notify(119886)
	}
	__antithesis_instrumentation__.Notify(119882)

	_ = r.requestLeaseLocked(ctx, st)
}

func (r *Replica) checkLeaseRespectsPreferences(ctx context.Context) (bool, error) {
	__antithesis_instrumentation__.Notify(119887)
	if !r.OwnsValidLease(ctx, r.store.cfg.Clock.NowAsClockTimestamp()) {
		__antithesis_instrumentation__.Notify(119892)
		return false, errors.Errorf("replica %s is not the leaseholder, cannot check lease preferences", r)
	} else {
		__antithesis_instrumentation__.Notify(119893)
	}
	__antithesis_instrumentation__.Notify(119888)
	conf := r.SpanConfig()
	if len(conf.LeasePreferences) == 0 {
		__antithesis_instrumentation__.Notify(119894)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(119895)
	}
	__antithesis_instrumentation__.Notify(119889)
	storeDesc, err := r.store.Descriptor(ctx, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(119896)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(119897)
	}
	__antithesis_instrumentation__.Notify(119890)
	for _, preference := range conf.LeasePreferences {
		__antithesis_instrumentation__.Notify(119898)
		if constraint.ConjunctionsCheck(*storeDesc, preference.Constraints) {
			__antithesis_instrumentation__.Notify(119899)
			return true, nil
		} else {
			__antithesis_instrumentation__.Notify(119900)
		}
	}
	__antithesis_instrumentation__.Notify(119891)
	return false, nil
}
