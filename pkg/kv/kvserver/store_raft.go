package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/liveness/livenesspb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type raftRequestInfo struct {
	req        *kvserverpb.RaftMessageRequest
	respStream RaftMessageResponseStream
}

type raftRequestQueue struct {
	syncutil.Mutex
	infos []raftRequestInfo
}

func (q *raftRequestQueue) drain() ([]raftRequestInfo, bool) {
	__antithesis_instrumentation__.Notify(125260)
	q.Lock()
	defer q.Unlock()
	if len(q.infos) == 0 {
		__antithesis_instrumentation__.Notify(125262)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(125263)
	}
	__antithesis_instrumentation__.Notify(125261)
	infos := q.infos
	q.infos = nil
	return infos, true
}

func (q *raftRequestQueue) recycle(processed []raftRequestInfo) {
	__antithesis_instrumentation__.Notify(125264)
	if cap(processed) > 4 {
		__antithesis_instrumentation__.Notify(125266)
		return
	} else {
		__antithesis_instrumentation__.Notify(125267)
	}
	__antithesis_instrumentation__.Notify(125265)
	q.Lock()
	defer q.Unlock()
	if q.infos == nil {
		__antithesis_instrumentation__.Notify(125268)
		for i := range processed {
			__antithesis_instrumentation__.Notify(125270)
			processed[i] = raftRequestInfo{}
		}
		__antithesis_instrumentation__.Notify(125269)
		q.infos = processed[:0]
	} else {
		__antithesis_instrumentation__.Notify(125271)
	}
}

func (s *Store) HandleSnapshot(
	ctx context.Context, header *kvserverpb.SnapshotRequest_Header, stream SnapshotResponseStream,
) error {
	__antithesis_instrumentation__.Notify(125272)
	ctx = s.AnnotateCtx(ctx)
	const name = "storage.Store: handle snapshot"
	return s.stopper.RunTaskWithErr(ctx, name, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(125273)
		s.metrics.RaftRcvdMessages[raftpb.MsgSnap].Inc(1)

		return s.receiveSnapshot(ctx, header, stream)
	})
}

func (s *Store) uncoalesceBeats(
	ctx context.Context,
	beats []kvserverpb.RaftHeartbeat,
	fromReplica, toReplica roachpb.ReplicaDescriptor,
	msgT raftpb.MessageType,
	respStream RaftMessageResponseStream,
) {
	__antithesis_instrumentation__.Notify(125274)
	if len(beats) == 0 {
		__antithesis_instrumentation__.Notify(125278)
		return
	} else {
		__antithesis_instrumentation__.Notify(125279)
	}
	__antithesis_instrumentation__.Notify(125275)
	if log.V(4) {
		__antithesis_instrumentation__.Notify(125280)
		log.Infof(ctx, "uncoalescing %d beats of type %v: %+v", len(beats), msgT, beats)
	} else {
		__antithesis_instrumentation__.Notify(125281)
	}
	__antithesis_instrumentation__.Notify(125276)
	beatReqs := make([]kvserverpb.RaftMessageRequest, len(beats))
	var toEnqueue []roachpb.RangeID
	for i, beat := range beats {
		__antithesis_instrumentation__.Notify(125282)
		msg := raftpb.Message{
			Type:   msgT,
			From:   uint64(beat.FromReplicaID),
			To:     uint64(beat.ToReplicaID),
			Term:   beat.Term,
			Commit: beat.Commit,
		}
		beatReqs[i] = kvserverpb.RaftMessageRequest{
			RangeID: beat.RangeID,
			FromReplica: roachpb.ReplicaDescriptor{
				NodeID:    fromReplica.NodeID,
				StoreID:   fromReplica.StoreID,
				ReplicaID: beat.FromReplicaID,
			},
			ToReplica: roachpb.ReplicaDescriptor{
				NodeID:    toReplica.NodeID,
				StoreID:   toReplica.StoreID,
				ReplicaID: beat.ToReplicaID,
			},
			Message:                   msg,
			Quiesce:                   beat.Quiesce,
			LaggingFollowersOnQuiesce: beat.LaggingFollowersOnQuiesce,
		}
		if log.V(4) {
			__antithesis_instrumentation__.Notify(125284)
			log.Infof(ctx, "uncoalesced beat: %+v", beatReqs[i])
		} else {
			__antithesis_instrumentation__.Notify(125285)
		}
		__antithesis_instrumentation__.Notify(125283)

		enqueue := s.HandleRaftUncoalescedRequest(ctx, &beatReqs[i], respStream)
		if enqueue {
			__antithesis_instrumentation__.Notify(125286)
			toEnqueue = append(toEnqueue, beat.RangeID)
		} else {
			__antithesis_instrumentation__.Notify(125287)
		}
	}
	__antithesis_instrumentation__.Notify(125277)
	s.scheduler.EnqueueRaftRequests(toEnqueue...)
}

func (s *Store) HandleRaftRequest(
	ctx context.Context, req *kvserverpb.RaftMessageRequest, respStream RaftMessageResponseStream,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(125288)

	if len(req.Heartbeats)+len(req.HeartbeatResps) > 0 {
		__antithesis_instrumentation__.Notify(125291)
		if req.RangeID != 0 {
			__antithesis_instrumentation__.Notify(125293)
			log.Fatalf(ctx, "coalesced heartbeats must have rangeID == 0")
		} else {
			__antithesis_instrumentation__.Notify(125294)
		}
		__antithesis_instrumentation__.Notify(125292)
		s.uncoalesceBeats(ctx, req.Heartbeats, req.FromReplica, req.ToReplica, raftpb.MsgHeartbeat, respStream)
		s.uncoalesceBeats(ctx, req.HeartbeatResps, req.FromReplica, req.ToReplica, raftpb.MsgHeartbeatResp, respStream)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(125295)
	}
	__antithesis_instrumentation__.Notify(125289)
	enqueue := s.HandleRaftUncoalescedRequest(ctx, req, respStream)
	if enqueue {
		__antithesis_instrumentation__.Notify(125296)
		s.scheduler.EnqueueRaftRequest(req.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(125297)
	}
	__antithesis_instrumentation__.Notify(125290)
	return nil
}

func (s *Store) HandleRaftUncoalescedRequest(
	ctx context.Context, req *kvserverpb.RaftMessageRequest, respStream RaftMessageResponseStream,
) (enqueue bool) {
	__antithesis_instrumentation__.Notify(125298)
	if len(req.Heartbeats)+len(req.HeartbeatResps) > 0 {
		__antithesis_instrumentation__.Notify(125302)
		log.Fatalf(ctx, "HandleRaftUncoalescedRequest cannot be given coalesced heartbeats or heartbeat responses, received %s", req)
	} else {
		__antithesis_instrumentation__.Notify(125303)
	}
	__antithesis_instrumentation__.Notify(125299)

	s.metrics.RaftRcvdMessages[req.Message.Type].Inc(1)

	value, ok := s.replicaQueues.Load(int64(req.RangeID))
	if !ok {
		__antithesis_instrumentation__.Notify(125304)
		value, _ = s.replicaQueues.LoadOrStore(int64(req.RangeID), unsafe.Pointer(&raftRequestQueue{}))
	} else {
		__antithesis_instrumentation__.Notify(125305)
	}
	__antithesis_instrumentation__.Notify(125300)
	q := (*raftRequestQueue)(value)
	q.Lock()
	defer q.Unlock()
	if len(q.infos) >= replicaRequestQueueSize {
		__antithesis_instrumentation__.Notify(125306)

		s.metrics.RaftRcvdMsgDropped.Inc(1)
		return false
	} else {
		__antithesis_instrumentation__.Notify(125307)
	}
	__antithesis_instrumentation__.Notify(125301)
	q.infos = append(q.infos, raftRequestInfo{
		req:        req,
		respStream: respStream,
	})

	return len(q.infos) == 1
}

func (s *Store) withReplicaForRequest(
	ctx context.Context,
	req *kvserverpb.RaftMessageRequest,
	f func(context.Context, *Replica) *roachpb.Error,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(125308)

	r, _, err := s.getOrCreateReplica(
		ctx,
		req.RangeID,
		req.ToReplica.ReplicaID,
		&req.FromReplica,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(125310)
		return roachpb.NewError(err)
	} else {
		__antithesis_instrumentation__.Notify(125311)
	}
	__antithesis_instrumentation__.Notify(125309)
	defer r.raftMu.Unlock()
	r.setLastReplicaDescriptorsRaftMuLocked(req)
	return f(ctx, r)
}

func (s *Store) processRaftRequestWithReplica(
	ctx context.Context, r *Replica, req *kvserverpb.RaftMessageRequest,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(125312)
	if verboseRaftLoggingEnabled() {
		__antithesis_instrumentation__.Notify(125318)
		log.Infof(ctx, "incoming raft message:\n%s", raftDescribeMessage(req.Message, raftEntryFormatter))
	} else {
		__antithesis_instrumentation__.Notify(125319)
	}
	__antithesis_instrumentation__.Notify(125313)

	if req.Message.Type == raftpb.MsgSnap {
		__antithesis_instrumentation__.Notify(125320)
		log.Fatalf(ctx, "unexpected snapshot: %+v", req)
	} else {
		__antithesis_instrumentation__.Notify(125321)
	}
	__antithesis_instrumentation__.Notify(125314)

	if req.Quiesce {
		__antithesis_instrumentation__.Notify(125322)
		if req.Message.Type != raftpb.MsgHeartbeat {
			__antithesis_instrumentation__.Notify(125324)
			log.Fatalf(ctx, "unexpected quiesce: %+v", req)
		} else {
			__antithesis_instrumentation__.Notify(125325)
		}
		__antithesis_instrumentation__.Notify(125323)
		if r.maybeQuiesceOnNotify(
			ctx,
			req.Message,
			laggingReplicaSet(req.LaggingFollowersOnQuiesce),
		) {
			__antithesis_instrumentation__.Notify(125326)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(125327)
		}
	} else {
		__antithesis_instrumentation__.Notify(125328)
	}
	__antithesis_instrumentation__.Notify(125315)

	if req.ToReplica.ReplicaID == 0 {
		__antithesis_instrumentation__.Notify(125329)
		log.VEventf(ctx, 1, "refusing incoming Raft message %s from %+v to %+v",
			req.Message.Type, req.FromReplica, req.ToReplica)
		return roachpb.NewErrorf(
			"cannot recreate replica that is not a member of its range (StoreID %s not found in r%d)",
			r.store.StoreID(), req.RangeID,
		)
	} else {
		__antithesis_instrumentation__.Notify(125330)
	}
	__antithesis_instrumentation__.Notify(125316)

	drop := maybeDropMsgApp(ctx, (*replicaMsgAppDropper)(r), &req.Message, req.RangeStartKey)
	if !drop {
		__antithesis_instrumentation__.Notify(125331)
		if err := r.stepRaftGroup(req); err != nil {
			__antithesis_instrumentation__.Notify(125332)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(125333)
		}
	} else {
		__antithesis_instrumentation__.Notify(125334)
	}
	__antithesis_instrumentation__.Notify(125317)
	return nil
}

func (s *Store) processRaftSnapshotRequest(
	ctx context.Context, snapHeader *kvserverpb.SnapshotRequest_Header, inSnap IncomingSnapshot,
) *roachpb.Error {
	__antithesis_instrumentation__.Notify(125335)
	return s.withReplicaForRequest(ctx, &snapHeader.RaftMessageRequest, func(
		ctx context.Context, r *Replica,
	) (pErr *roachpb.Error) {
		__antithesis_instrumentation__.Notify(125336)
		ctx = r.AnnotateCtx(ctx)
		if snapHeader.RaftMessageRequest.Message.Type != raftpb.MsgSnap {
			__antithesis_instrumentation__.Notify(125342)
			log.Fatalf(ctx, "expected snapshot: %+v", snapHeader.RaftMessageRequest)
		} else {
			__antithesis_instrumentation__.Notify(125343)
		}
		__antithesis_instrumentation__.Notify(125337)

		var stats handleRaftReadyStats
		typ := removePlaceholderFailed
		defer func() {
			__antithesis_instrumentation__.Notify(125344)

			if inSnap.placeholder != nil {
				__antithesis_instrumentation__.Notify(125345)
				if _, err := s.removePlaceholder(ctx, inSnap.placeholder, typ); err != nil {
					__antithesis_instrumentation__.Notify(125346)
					log.Fatalf(ctx, "unable to remove placeholder: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(125347)
				}
			} else {
				__antithesis_instrumentation__.Notify(125348)
			}
		}()
		__antithesis_instrumentation__.Notify(125338)

		if snapHeader.RaftMessageRequest.Message.From == snapHeader.RaftMessageRequest.Message.To {
			__antithesis_instrumentation__.Notify(125349)

			snapHeader.RaftMessageRequest.Message.From = 0
		} else {
			__antithesis_instrumentation__.Notify(125350)
		}
		__antithesis_instrumentation__.Notify(125339)

		if err := r.stepRaftGroup(&snapHeader.RaftMessageRequest); err != nil {
			__antithesis_instrumentation__.Notify(125351)
			return roachpb.NewError(err)
		} else {
			__antithesis_instrumentation__.Notify(125352)
		}
		__antithesis_instrumentation__.Notify(125340)

		typ = removePlaceholderDropped
		var expl string
		var err error
		stats, expl, err = r.handleRaftReadyRaftMuLocked(ctx, inSnap)
		maybeFatalOnRaftReadyErr(ctx, expl, err)
		if !stats.snap.applied {
			__antithesis_instrumentation__.Notify(125353)

			log.Infof(ctx, "ignored stale snapshot at index %d", snapHeader.RaftMessageRequest.Message.Snapshot.Metadata.Index)
		} else {
			__antithesis_instrumentation__.Notify(125354)
		}
		__antithesis_instrumentation__.Notify(125341)
		return nil
	})
}

func (s *Store) HandleRaftResponse(
	ctx context.Context, resp *kvserverpb.RaftMessageResponse,
) error {
	__antithesis_instrumentation__.Notify(125355)
	ctx = s.AnnotateCtx(ctx)
	const name = "storage.Store: handle raft response"
	return s.stopper.RunTaskWithErr(ctx, name, func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(125356)
		repl, replErr := s.GetReplica(resp.RangeID)
		if replErr == nil {
			__antithesis_instrumentation__.Notify(125359)

			ctx = repl.AnnotateCtx(ctx)
		} else {
			__antithesis_instrumentation__.Notify(125360)
		}
		__antithesis_instrumentation__.Notify(125357)
		switch val := resp.Union.GetValue().(type) {
		case *roachpb.Error:
			__antithesis_instrumentation__.Notify(125361)
			switch tErr := val.GetDetail().(type) {
			case *roachpb.ReplicaTooOldError:
				__antithesis_instrumentation__.Notify(125363)
				if replErr != nil {
					__antithesis_instrumentation__.Notify(125371)

					if !errors.HasType(replErr, (*roachpb.RangeNotFoundError)(nil)) {
						__antithesis_instrumentation__.Notify(125373)
						log.Errorf(ctx, "%v", replErr)
					} else {
						__antithesis_instrumentation__.Notify(125374)
					}
					__antithesis_instrumentation__.Notify(125372)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(125375)
				}
				__antithesis_instrumentation__.Notify(125364)

				repl.raftMu.Lock()
				defer repl.raftMu.Unlock()
				repl.mu.Lock()

				if tErr.ReplicaID != repl.replicaID || func() bool {
					__antithesis_instrumentation__.Notify(125376)
					return !repl.mu.destroyStatus.IsAlive() == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(125377)
					return s.TestingKnobs().DisableEagerReplicaRemoval == true
				}() == true {
					__antithesis_instrumentation__.Notify(125378)
					repl.mu.Unlock()
					return nil
				} else {
					__antithesis_instrumentation__.Notify(125379)
				}
				__antithesis_instrumentation__.Notify(125365)

				if log.V(1) {
					__antithesis_instrumentation__.Notify(125380)
					log.Infof(ctx, "setting local replica to destroyed due to ReplicaTooOld error")
				} else {
					__antithesis_instrumentation__.Notify(125381)
				}
				__antithesis_instrumentation__.Notify(125366)

				repl.mu.Unlock()
				nextReplicaID := tErr.ReplicaID + 1
				return s.removeReplicaRaftMuLocked(ctx, repl, nextReplicaID, RemoveOptions{
					DestroyData: true,
				})
			case *roachpb.RaftGroupDeletedError:
				__antithesis_instrumentation__.Notify(125367)
				if replErr != nil {
					__antithesis_instrumentation__.Notify(125382)

					if !errors.HasType(replErr, (*roachpb.RangeNotFoundError)(nil)) {
						__antithesis_instrumentation__.Notify(125384)
						log.Errorf(ctx, "%v", replErr)
					} else {
						__antithesis_instrumentation__.Notify(125385)
					}
					__antithesis_instrumentation__.Notify(125383)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(125386)
				}
				__antithesis_instrumentation__.Notify(125368)

				s.replicaGCQueue.AddAsync(ctx, repl, replicaGCPriorityDefault)
			case *roachpb.StoreNotFoundError:
				__antithesis_instrumentation__.Notify(125369)
				log.Warningf(ctx, "raft error: node %d claims to not contain store %d for replica %s: %s",
					resp.FromReplica.NodeID, resp.FromReplica.StoreID, resp.FromReplica, val)
				return val.GetDetail()
			default:
				__antithesis_instrumentation__.Notify(125370)
				log.Warningf(ctx, "got error from r%d, replica %s: %s",
					resp.RangeID, resp.FromReplica, val)
			}
		default:
			__antithesis_instrumentation__.Notify(125362)
			log.Warningf(ctx, "got unknown raft response type %T from replica %s: %s", val, resp.FromReplica, val)
		}
		__antithesis_instrumentation__.Notify(125358)
		return nil
	})
}

func (s *Store) enqueueRaftUpdateCheck(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(125387)
	s.scheduler.EnqueueRaftReady(rangeID)
}

func (s *Store) processRequestQueue(ctx context.Context, rangeID roachpb.RangeID) bool {
	__antithesis_instrumentation__.Notify(125388)
	value, ok := s.replicaQueues.Load(int64(rangeID))
	if !ok {
		__antithesis_instrumentation__.Notify(125393)
		return false
	} else {
		__antithesis_instrumentation__.Notify(125394)
	}
	__antithesis_instrumentation__.Notify(125389)
	q := (*raftRequestQueue)(value)
	infos, ok := q.drain()
	if !ok {
		__antithesis_instrumentation__.Notify(125395)
		return false
	} else {
		__antithesis_instrumentation__.Notify(125396)
	}
	__antithesis_instrumentation__.Notify(125390)
	defer q.recycle(infos)

	var hadError bool
	for i := range infos {
		__antithesis_instrumentation__.Notify(125397)
		info := &infos[i]
		if pErr := s.withReplicaForRequest(
			ctx, info.req, func(_ context.Context, r *Replica) *roachpb.Error {
				__antithesis_instrumentation__.Notify(125398)
				return s.processRaftRequestWithReplica(r.raftCtx, r, info.req)
			},
		); pErr != nil {
			__antithesis_instrumentation__.Notify(125399)
			hadError = true
			if err := info.respStream.Send(newRaftMessageResponse(info.req, pErr)); err != nil {
				__antithesis_instrumentation__.Notify(125400)

				log.VEventf(ctx, 1, "error sending error: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(125401)
			}
		} else {
			__antithesis_instrumentation__.Notify(125402)
		}
	}
	__antithesis_instrumentation__.Notify(125391)

	if hadError {
		__antithesis_instrumentation__.Notify(125403)

		if _, exists := s.mu.replicasByRangeID.Load(rangeID); !exists {
			__antithesis_instrumentation__.Notify(125404)
			q.Lock()
			if len(q.infos) == 0 {
				__antithesis_instrumentation__.Notify(125406)
				s.replicaQueues.Delete(int64(rangeID))
			} else {
				__antithesis_instrumentation__.Notify(125407)
			}
			__antithesis_instrumentation__.Notify(125405)
			q.Unlock()
		} else {
			__antithesis_instrumentation__.Notify(125408)
		}
	} else {
		__antithesis_instrumentation__.Notify(125409)
	}
	__antithesis_instrumentation__.Notify(125392)

	return true
}

func (s *Store) processReady(rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(125410)
	r, ok := s.mu.replicasByRangeID.Load(rangeID)
	if !ok {
		__antithesis_instrumentation__.Notify(125412)
		return
	} else {
		__antithesis_instrumentation__.Notify(125413)
	}
	__antithesis_instrumentation__.Notify(125411)

	ctx := r.raftCtx
	start := timeutil.Now()
	stats, expl, err := r.handleRaftReady(ctx, noSnap)
	maybeFatalOnRaftReadyErr(ctx, expl, err)
	elapsed := timeutil.Since(start)
	s.metrics.RaftWorkingDurationNanos.Inc(elapsed.Nanoseconds())
	s.metrics.RaftHandleReadyLatency.RecordValue(elapsed.Nanoseconds())

	if elapsed >= defaultReplicaRaftMuWarnThreshold {
		__antithesis_instrumentation__.Notify(125414)
		log.Infof(ctx, "handle raft ready: %.1fs [applied=%d, batches=%d, state_assertions=%d]; node might be overloaded",
			elapsed.Seconds(), stats.entriesProcessed, stats.batchesProcessed, stats.stateAssertions)
	} else {
		__antithesis_instrumentation__.Notify(125415)
	}
}

func (s *Store) processTick(_ context.Context, rangeID roachpb.RangeID) bool {
	__antithesis_instrumentation__.Notify(125416)
	r, ok := s.mu.replicasByRangeID.Load(rangeID)
	if !ok {
		__antithesis_instrumentation__.Notify(125419)
		return false
	} else {
		__antithesis_instrumentation__.Notify(125420)
	}
	__antithesis_instrumentation__.Notify(125417)
	livenessMap, _ := s.livenessMap.Load().(liveness.IsLiveMap)

	start := timeutil.Now()
	ctx := r.raftCtx
	exists, err := r.tick(ctx, livenessMap)
	if err != nil {
		__antithesis_instrumentation__.Notify(125421)
		log.Errorf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(125422)
	}
	__antithesis_instrumentation__.Notify(125418)
	s.metrics.RaftTickingDurationNanos.Inc(timeutil.Since(start).Nanoseconds())
	return exists
}

func (s *Store) nodeIsLiveCallback(l livenesspb.Liveness) {
	__antithesis_instrumentation__.Notify(125423)
	s.updateLivenessMap()

	s.mu.replicasByRangeID.Range(func(r *Replica) {
		__antithesis_instrumentation__.Notify(125424)
		r.mu.RLock()
		quiescent := r.mu.quiescent
		lagging := r.mu.laggingFollowersOnQuiesce
		r.mu.RUnlock()
		if quiescent && func() bool {
			__antithesis_instrumentation__.Notify(125425)
			return lagging.MemberStale(l) == true
		}() == true {
			__antithesis_instrumentation__.Notify(125426)
			r.maybeUnquiesce()
		} else {
			__antithesis_instrumentation__.Notify(125427)
		}
	})
}

func (s *Store) processRaft(ctx context.Context) {
	__antithesis_instrumentation__.Notify(125428)
	if s.cfg.TestingKnobs.DisableProcessRaft {
		__antithesis_instrumentation__.Notify(125432)
		return
	} else {
		__antithesis_instrumentation__.Notify(125433)
	}
	__antithesis_instrumentation__.Notify(125429)

	s.scheduler.Start(s.stopper)

	if err := s.stopper.RunAsyncTask(ctx, "sched-wait", s.scheduler.Wait); err != nil {
		__antithesis_instrumentation__.Notify(125434)
		s.scheduler.Wait(ctx)
	} else {
		__antithesis_instrumentation__.Notify(125435)
	}
	__antithesis_instrumentation__.Notify(125430)

	_ = s.stopper.RunAsyncTask(ctx, "sched-tick-loop", s.raftTickLoop)
	_ = s.stopper.RunAsyncTask(ctx, "coalesced-hb-loop", s.coalescedHeartbeatsLoop)
	s.stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(125436)
		s.cfg.Transport.Stop(s.StoreID())
	}))
	__antithesis_instrumentation__.Notify(125431)

	s.stopper.AddCloser(stop.CloserFn(func() {
		__antithesis_instrumentation__.Notify(125437)
		s.VisitReplicas(func(r *Replica) (more bool) {
			__antithesis_instrumentation__.Notify(125438)
			r.mu.Lock()
			r.mu.proposalBuf.FlushLockedWithoutProposing(ctx)
			for k, prop := range r.mu.proposals {
				__antithesis_instrumentation__.Notify(125440)
				delete(r.mu.proposals, k)
				prop.finishApplication(
					context.Background(),
					proposalResult{
						Err: roachpb.NewError(roachpb.NewAmbiguousResultErrorf("store is stopping")),
					},
				)
			}
			__antithesis_instrumentation__.Notify(125439)
			r.mu.Unlock()
			return true
		})
	}))
}

func (s *Store) raftTickLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(125441)
	ticker := time.NewTicker(s.cfg.RaftTickInterval)
	defer ticker.Stop()

	var rangeIDs []roachpb.RangeID

	for {
		__antithesis_instrumentation__.Notify(125442)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(125443)
			rangeIDs = rangeIDs[:0]

			if s.cfg.NodeLiveness != nil {
				__antithesis_instrumentation__.Notify(125447)
				s.updateLivenessMap()
			} else {
				__antithesis_instrumentation__.Notify(125448)
			}
			__antithesis_instrumentation__.Notify(125444)

			s.unquiescedReplicas.Lock()

			for rangeID := range s.unquiescedReplicas.m {
				__antithesis_instrumentation__.Notify(125449)
				rangeIDs = append(rangeIDs, rangeID)
			}
			__antithesis_instrumentation__.Notify(125445)
			s.unquiescedReplicas.Unlock()

			s.scheduler.EnqueueRaftTicks(rangeIDs...)
			s.metrics.RaftTicks.Inc(1)

		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(125446)
			return
		}
	}
}

func (s *Store) updateLivenessMap() {
	__antithesis_instrumentation__.Notify(125450)
	nextMap := s.cfg.NodeLiveness.GetIsLiveMap()
	for nodeID, entry := range nextMap {
		__antithesis_instrumentation__.Notify(125452)
		if entry.IsLive {
			__antithesis_instrumentation__.Notify(125454)
			continue
		} else {
			__antithesis_instrumentation__.Notify(125455)
		}
		__antithesis_instrumentation__.Notify(125453)

		entry.IsLive = (s.cfg.NodeDialer.ConnHealth(nodeID, rpc.SystemClass) == nil)
		nextMap[nodeID] = entry
	}
	__antithesis_instrumentation__.Notify(125451)
	s.livenessMap.Store(nextMap)
}

func (s *Store) coalescedHeartbeatsLoop(ctx context.Context) {
	__antithesis_instrumentation__.Notify(125456)
	ticker := time.NewTicker(s.cfg.CoalescedHeartbeatsInterval)
	defer ticker.Stop()

	for {
		__antithesis_instrumentation__.Notify(125457)
		select {
		case <-ticker.C:
			__antithesis_instrumentation__.Notify(125458)
			s.sendQueuedHeartbeats(ctx)
		case <-s.stopper.ShouldQuiesce():
			__antithesis_instrumentation__.Notify(125459)
			return
		}
	}
}

func (s *Store) sendQueuedHeartbeatsToNode(
	ctx context.Context, beats, resps []kvserverpb.RaftHeartbeat, to roachpb.StoreIdent,
) int {
	__antithesis_instrumentation__.Notify(125460)
	var msgType raftpb.MessageType

	if len(beats) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(125464)
		return len(resps) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(125465)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(125466)
		if len(resps) == 0 {
			__antithesis_instrumentation__.Notify(125467)
			msgType = raftpb.MsgHeartbeat
		} else {
			__antithesis_instrumentation__.Notify(125468)
			if len(beats) == 0 {
				__antithesis_instrumentation__.Notify(125469)
				msgType = raftpb.MsgHeartbeatResp
			} else {
				__antithesis_instrumentation__.Notify(125470)
				log.Fatal(ctx, "cannot coalesce both heartbeats and responses")
			}
		}
	}
	__antithesis_instrumentation__.Notify(125461)

	chReq := newRaftMessageRequest()
	*chReq = kvserverpb.RaftMessageRequest{
		RangeID: 0,
		ToReplica: roachpb.ReplicaDescriptor{
			NodeID:    to.NodeID,
			StoreID:   to.StoreID,
			ReplicaID: 0,
		},
		FromReplica: roachpb.ReplicaDescriptor{
			NodeID:  s.Ident.NodeID,
			StoreID: s.Ident.StoreID,
		},
		Message: raftpb.Message{
			Type: msgType,
		},
		Heartbeats:     beats,
		HeartbeatResps: resps,
	}

	if log.V(4) {
		__antithesis_instrumentation__.Notify(125471)
		log.Infof(ctx, "sending raft request (coalesced) %+v", chReq)
	} else {
		__antithesis_instrumentation__.Notify(125472)
	}
	__antithesis_instrumentation__.Notify(125462)

	if !s.cfg.Transport.SendAsync(chReq, rpc.SystemClass) {
		__antithesis_instrumentation__.Notify(125473)
		for _, beat := range beats {
			__antithesis_instrumentation__.Notify(125476)
			if repl, ok := s.mu.replicasByRangeID.Load(beat.RangeID); ok {
				__antithesis_instrumentation__.Notify(125477)
				repl.addUnreachableRemoteReplica(beat.ToReplicaID)
			} else {
				__antithesis_instrumentation__.Notify(125478)
			}
		}
		__antithesis_instrumentation__.Notify(125474)
		for _, resp := range resps {
			__antithesis_instrumentation__.Notify(125479)
			if repl, ok := s.mu.replicasByRangeID.Load(resp.RangeID); ok {
				__antithesis_instrumentation__.Notify(125480)
				repl.addUnreachableRemoteReplica(resp.ToReplicaID)
			} else {
				__antithesis_instrumentation__.Notify(125481)
			}
		}
		__antithesis_instrumentation__.Notify(125475)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(125482)
	}
	__antithesis_instrumentation__.Notify(125463)
	return len(beats) + len(resps)
}

func (s *Store) sendQueuedHeartbeats(ctx context.Context) {
	__antithesis_instrumentation__.Notify(125483)
	s.coalescedMu.Lock()
	heartbeats := s.coalescedMu.heartbeats
	heartbeatResponses := s.coalescedMu.heartbeatResponses
	s.coalescedMu.heartbeats = map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat{}
	s.coalescedMu.heartbeatResponses = map[roachpb.StoreIdent][]kvserverpb.RaftHeartbeat{}
	s.coalescedMu.Unlock()

	var beatsSent int

	for to, beats := range heartbeats {
		__antithesis_instrumentation__.Notify(125486)
		beatsSent += s.sendQueuedHeartbeatsToNode(ctx, beats, nil, to)
	}
	__antithesis_instrumentation__.Notify(125484)
	for to, resps := range heartbeatResponses {
		__antithesis_instrumentation__.Notify(125487)
		beatsSent += s.sendQueuedHeartbeatsToNode(ctx, nil, resps, to)
	}
	__antithesis_instrumentation__.Notify(125485)
	s.metrics.RaftCoalescedHeartbeatsPending.Update(int64(beatsSent))
}

func (s *Store) updateCapacityGauges(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(125488)
	desc, err := s.Descriptor(ctx, false)
	if err != nil {
		__antithesis_instrumentation__.Notify(125490)
		return err
	} else {
		__antithesis_instrumentation__.Notify(125491)
	}
	__antithesis_instrumentation__.Notify(125489)
	s.metrics.Capacity.Update(desc.Capacity.Capacity)
	s.metrics.Available.Update(desc.Capacity.Available)
	s.metrics.Used.Update(desc.Capacity.Used)

	return nil
}
