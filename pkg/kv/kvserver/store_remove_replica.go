package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type RemoveOptions struct {
	DestroyData bool

	InsertPlaceholder bool
}

func (s *Store) RemoveReplica(
	ctx context.Context, rep *Replica, nextReplicaID roachpb.ReplicaID, opts RemoveOptions,
) error {
	__antithesis_instrumentation__.Notify(125644)
	rep.raftMu.Lock()
	defer rep.raftMu.Unlock()
	if opts.InsertPlaceholder {
		__antithesis_instrumentation__.Notify(125646)
		return errors.Errorf("InsertPlaceholder not supported in RemoveReplica")
	} else {
		__antithesis_instrumentation__.Notify(125647)
	}
	__antithesis_instrumentation__.Notify(125645)
	_, err := s.removeInitializedReplicaRaftMuLocked(ctx, rep, nextReplicaID, opts)
	return err
}

func (s *Store) removeReplicaRaftMuLocked(
	ctx context.Context, rep *Replica, nextReplicaID roachpb.ReplicaID, opts RemoveOptions,
) error {
	__antithesis_instrumentation__.Notify(125648)
	rep.raftMu.AssertHeld()
	if rep.IsInitialized() {
		__antithesis_instrumentation__.Notify(125650)
		if opts.InsertPlaceholder {
			__antithesis_instrumentation__.Notify(125652)
			return errors.Errorf("InsertPlaceholder unsupported in removeReplicaRaftMuLocked")
		} else {
			__antithesis_instrumentation__.Notify(125653)
		}
		__antithesis_instrumentation__.Notify(125651)
		_, err := s.removeInitializedReplicaRaftMuLocked(ctx, rep, nextReplicaID, opts)
		return errors.Wrap(err,
			"failed to remove replica")
	} else {
		__antithesis_instrumentation__.Notify(125654)
	}
	__antithesis_instrumentation__.Notify(125649)
	s.removeUninitializedReplicaRaftMuLocked(ctx, rep, nextReplicaID)
	return nil
}

func (s *Store) removeInitializedReplicaRaftMuLocked(
	ctx context.Context, rep *Replica, nextReplicaID roachpb.ReplicaID, opts RemoveOptions,
) (*ReplicaPlaceholder, error) {
	__antithesis_instrumentation__.Notify(125655)
	rep.raftMu.AssertHeld()
	if !rep.IsInitialized() {
		__antithesis_instrumentation__.Notify(125663)
		return nil, errors.AssertionFailedf("cannot remove uninitialized replica %s", rep)
	} else {
		__antithesis_instrumentation__.Notify(125664)
	}
	__antithesis_instrumentation__.Notify(125656)

	if opts.InsertPlaceholder {
		__antithesis_instrumentation__.Notify(125665)
		if opts.DestroyData {
			__antithesis_instrumentation__.Notify(125666)
			return nil, errors.AssertionFailedf("cannot specify both InsertPlaceholder and DestroyData")
		} else {
			__antithesis_instrumentation__.Notify(125667)
		}

	} else {
		__antithesis_instrumentation__.Notify(125668)
	}
	__antithesis_instrumentation__.Notify(125657)

	var desc *roachpb.RangeDescriptor
	{
		__antithesis_instrumentation__.Notify(125669)
		rep.readOnlyCmdMu.Lock()
		rep.mu.Lock()

		if opts.DestroyData {
			__antithesis_instrumentation__.Notify(125673)

			if rep.mu.destroyStatus.Removed() {
				__antithesis_instrumentation__.Notify(125674)
				rep.mu.Unlock()
				rep.readOnlyCmdMu.Unlock()
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(125675)
			}
		} else {
			__antithesis_instrumentation__.Notify(125676)

			if !rep.mu.destroyStatus.Removed() {
				__antithesis_instrumentation__.Notify(125677)
				rep.mu.Unlock()
				rep.readOnlyCmdMu.Unlock()
				log.Fatalf(ctx, "replica not marked as destroyed but data already destroyed: %v", rep)
			} else {
				__antithesis_instrumentation__.Notify(125678)
			}
		}
		__antithesis_instrumentation__.Notify(125670)

		desc = rep.mu.state.Desc
		if repDesc, ok := desc.GetReplicaDescriptor(s.StoreID()); ok && func() bool {
			__antithesis_instrumentation__.Notify(125679)
			return repDesc.ReplicaID >= nextReplicaID == true
		}() == true {
			__antithesis_instrumentation__.Notify(125680)
			rep.mu.Unlock()
			rep.readOnlyCmdMu.Unlock()

			log.Fatalf(ctx, "replica descriptor's ID has changed (%s >= %s)",
				repDesc.ReplicaID, nextReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(125681)
		}
		__antithesis_instrumentation__.Notify(125671)

		if !rep.isInitializedRLocked() {
			__antithesis_instrumentation__.Notify(125682)
			rep.mu.Unlock()
			rep.readOnlyCmdMu.Unlock()
			log.Fatalf(ctx, "uninitialized replica cannot be removed with removeInitializedReplica: %v", rep)
		} else {
			__antithesis_instrumentation__.Notify(125683)
		}
		__antithesis_instrumentation__.Notify(125672)

		rep.mu.destroyStatus.Set(roachpb.NewRangeNotFoundError(rep.RangeID, rep.StoreID()),
			destroyReasonRemoved)
		rep.mu.Unlock()
		rep.readOnlyCmdMu.Unlock()
	}
	__antithesis_instrumentation__.Notify(125658)

	existing, err := s.GetReplica(rep.RangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(125684)
		log.Fatalf(ctx, "cannot remove replica which does not exist: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(125685)
		if existing != rep {
			__antithesis_instrumentation__.Notify(125686)
			log.Fatalf(ctx, "replica %v replaced by %v before being removed",
				rep, existing)
		} else {
			__antithesis_instrumentation__.Notify(125687)
		}
	}
	__antithesis_instrumentation__.Notify(125659)

	log.Infof(ctx, "removing replica r%d/%d", rep.RangeID, rep.replicaID)

	s.mu.Lock()
	if it := s.getOverlappingKeyRangeLocked(desc); it.repl != rep {
		__antithesis_instrumentation__.Notify(125688)

		s.mu.Unlock()
		log.Fatalf(ctx, "replica %+v unexpectedly overlapped by %+v", rep, it.item)
	} else {
		__antithesis_instrumentation__.Notify(125689)
	}
	__antithesis_instrumentation__.Notify(125660)

	s.metrics.subtractMVCCStats(ctx, rep.tenantMetricsRef, rep.GetMVCCStats())
	s.metrics.ReplicaCount.Dec(1)
	s.mu.Unlock()

	rep.disconnectRangefeedWithReason(
		roachpb.RangeFeedRetryError_REASON_REPLICA_REMOVED,
	)

	rep.disconnectReplicationRaftMuLocked(ctx)
	if opts.DestroyData {
		__antithesis_instrumentation__.Notify(125690)
		if err := rep.destroyRaftMuLocked(ctx, nextReplicaID); err != nil {
			__antithesis_instrumentation__.Notify(125691)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(125692)
		}
	} else {
		__antithesis_instrumentation__.Notify(125693)
	}
	__antithesis_instrumentation__.Notify(125661)

	ph := func() *ReplicaPlaceholder {
		__antithesis_instrumentation__.Notify(125694)
		s.mu.Lock()
		defer s.mu.Unlock()

		s.unlinkReplicaByRangeIDLocked(ctx, rep.RangeID)

		if ph, ok := s.mu.replicaPlaceholders[rep.RangeID]; ok {
			__antithesis_instrumentation__.Notify(125700)
			log.Fatalf(ctx, "initialized replica %s unexpectedly had a placeholder: %+v", rep, ph)
		} else {
			__antithesis_instrumentation__.Notify(125701)
		}
		__antithesis_instrumentation__.Notify(125695)
		desc := rep.Desc()
		ph := &ReplicaPlaceholder{
			rangeDesc: *roachpb.NewRangeDescriptor(desc.RangeID, desc.StartKey, desc.EndKey, desc.Replicas()),
		}

		if it := s.mu.replicasByKey.ReplaceOrInsertPlaceholder(ctx, ph); it.repl != rep {
			__antithesis_instrumentation__.Notify(125702)

			log.Fatalf(ctx, "replica %+v unexpectedly overlapped by %+v", rep, it.item)
		} else {
			__antithesis_instrumentation__.Notify(125703)
		}
		__antithesis_instrumentation__.Notify(125696)
		if exPH, ok := s.mu.replicaPlaceholders[desc.RangeID]; ok {
			__antithesis_instrumentation__.Notify(125704)
			log.Fatalf(ctx, "cannot insert placeholder %s, already have %s", ph, exPH)
		} else {
			__antithesis_instrumentation__.Notify(125705)
		}
		__antithesis_instrumentation__.Notify(125697)
		s.mu.replicaPlaceholders[desc.RangeID] = ph

		if opts.InsertPlaceholder {
			__antithesis_instrumentation__.Notify(125706)
			return ph
		} else {
			__antithesis_instrumentation__.Notify(125707)
		}
		__antithesis_instrumentation__.Notify(125698)

		s.mu.replicasByKey.DeletePlaceholder(ctx, ph)
		delete(s.mu.replicaPlaceholders, desc.RangeID)
		if it := s.getOverlappingKeyRangeLocked(desc); it.item != nil && func() bool {
			__antithesis_instrumentation__.Notify(125708)
			return it.item != ph == true
		}() == true {
			__antithesis_instrumentation__.Notify(125709)
			log.Fatalf(ctx, "corrupted replicasByKey map: %s and %s overlapped", rep, it.item)
		} else {
			__antithesis_instrumentation__.Notify(125710)
		}
		__antithesis_instrumentation__.Notify(125699)
		return nil
	}()
	__antithesis_instrumentation__.Notify(125662)

	s.maybeGossipOnCapacityChange(ctx, rangeRemoveEvent)
	s.scanner.RemoveReplica(rep)
	return ph, nil
}

func (s *Store) removeUninitializedReplicaRaftMuLocked(
	ctx context.Context, rep *Replica, nextReplicaID roachpb.ReplicaID,
) {
	__antithesis_instrumentation__.Notify(125711)
	rep.raftMu.AssertHeld()

	{
		__antithesis_instrumentation__.Notify(125716)
		rep.readOnlyCmdMu.Lock()
		rep.mu.Lock()

		if rep.mu.destroyStatus.Removed() {
			__antithesis_instrumentation__.Notify(125719)
			rep.mu.Unlock()
			rep.readOnlyCmdMu.Unlock()
			log.Fatalf(ctx, "uninitialized replica unexpectedly already removed")
		} else {
			__antithesis_instrumentation__.Notify(125720)
		}
		__antithesis_instrumentation__.Notify(125717)

		if rep.isInitializedRLocked() {
			__antithesis_instrumentation__.Notify(125721)
			rep.mu.Unlock()
			rep.readOnlyCmdMu.Unlock()
			log.Fatalf(ctx, "cannot remove initialized replica in removeUninitializedReplica: %v", rep)
		} else {
			__antithesis_instrumentation__.Notify(125722)
		}
		__antithesis_instrumentation__.Notify(125718)

		rep.mu.destroyStatus.Set(roachpb.NewRangeNotFoundError(rep.RangeID, rep.StoreID()),
			destroyReasonRemoved)

		rep.mu.Unlock()
		rep.readOnlyCmdMu.Unlock()
	}
	__antithesis_instrumentation__.Notify(125712)

	rep.disconnectReplicationRaftMuLocked(ctx)
	if err := rep.destroyRaftMuLocked(ctx, nextReplicaID); err != nil {
		__antithesis_instrumentation__.Notify(125723)
		log.Fatalf(ctx, "failed to remove uninitialized replica %v: %v", rep, err)
	} else {
		__antithesis_instrumentation__.Notify(125724)
	}
	__antithesis_instrumentation__.Notify(125713)

	s.mu.Lock()
	defer s.mu.Unlock()

	existing, stillExists := s.mu.replicasByRangeID.Load(rep.RangeID)
	if !stillExists {
		__antithesis_instrumentation__.Notify(125725)
		log.Fatalf(ctx, "uninitialized replica was removed in the meantime")
	} else {
		__antithesis_instrumentation__.Notify(125726)
	}
	__antithesis_instrumentation__.Notify(125714)
	if existing == rep {
		__antithesis_instrumentation__.Notify(125727)
		log.Infof(ctx, "removing uninitialized replica %v", rep)
	} else {
		__antithesis_instrumentation__.Notify(125728)
		log.Fatalf(ctx, "uninitialized replica %v was unexpectedly replaced", existing)
	}
	__antithesis_instrumentation__.Notify(125715)

	s.unlinkReplicaByRangeIDLocked(ctx, rep.RangeID)
}

func (s *Store) unlinkReplicaByRangeIDLocked(ctx context.Context, rangeID roachpb.RangeID) {
	__antithesis_instrumentation__.Notify(125729)
	s.mu.AssertHeld()
	s.unquiescedReplicas.Lock()
	delete(s.unquiescedReplicas.m, rangeID)
	s.unquiescedReplicas.Unlock()
	delete(s.mu.uninitReplicas, rangeID)
	s.replicaQueues.Delete(int64(rangeID))
	s.mu.replicasByRangeID.Delete(rangeID)
	s.unregisterLeaseholderByID(ctx, rangeID)
}

func (s *Store) removePlaceholder(
	ctx context.Context, ph *ReplicaPlaceholder, typ removePlaceholderType,
) (removed bool, _ error) {
	__antithesis_instrumentation__.Notify(125730)
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.removePlaceholderLocked(ctx, ph, typ)
}

type removePlaceholderType byte

const (
	removePlaceholderFilled removePlaceholderType = iota

	removePlaceholderDropped

	removePlaceholderFailed
)

func (s *Store) removePlaceholderLocked(
	ctx context.Context, inPH *ReplicaPlaceholder, typ removePlaceholderType,
) (removed bool, _ error) {
	__antithesis_instrumentation__.Notify(125731)
	rngID := inPH.Desc().RangeID
	placeholder, ok := s.mu.replicaPlaceholders[rngID]

	if wasTainted := !atomic.CompareAndSwapInt32(&inPH.tainted, 0, 1); wasTainted {
		__antithesis_instrumentation__.Notify(125738)
		if typ == removePlaceholderFilled {
			__antithesis_instrumentation__.Notify(125741)

			return false, errors.AssertionFailedf(
				"attempting to fill tainted placeholder %+v (stored placeholder: %+v)", inPH, placeholder,
			)
		} else {
			__antithesis_instrumentation__.Notify(125742)
		}
		__antithesis_instrumentation__.Notify(125739)

		if ok && func() bool {
			__antithesis_instrumentation__.Notify(125743)
			return inPH == placeholder == true
		}() == true {
			__antithesis_instrumentation__.Notify(125744)
			return false, errors.AssertionFailedf(
				"tainted placeholder %+v unexpectedly present in replicaPlaceholders: %+v", inPH, placeholder,
			)

		} else {
			__antithesis_instrumentation__.Notify(125745)
		}
		__antithesis_instrumentation__.Notify(125740)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(125746)
	}
	__antithesis_instrumentation__.Notify(125732)

	if !ok {
		__antithesis_instrumentation__.Notify(125747)
		return false, errors.AssertionFailedf("expected placeholder %+v to exist", inPH)
	} else {
		__antithesis_instrumentation__.Notify(125748)
	}
	__antithesis_instrumentation__.Notify(125733)

	if placeholder != inPH {
		__antithesis_instrumentation__.Notify(125749)

		return false, errors.AssertionFailedf(
			"placeholder %+v is being dropped or filled, but store has conflicting placeholder %+v",
			inPH, placeholder,
		)
	} else {
		__antithesis_instrumentation__.Notify(125750)
	}
	__antithesis_instrumentation__.Notify(125734)

	if it := s.mu.replicasByKey.DeletePlaceholder(ctx, placeholder); it.ph != placeholder {
		__antithesis_instrumentation__.Notify(125751)
		return false, errors.AssertionFailedf("placeholder %+v not found, got %+v", placeholder, it)
	} else {
		__antithesis_instrumentation__.Notify(125752)
	}
	__antithesis_instrumentation__.Notify(125735)
	delete(s.mu.replicaPlaceholders, rngID)
	if it := s.getOverlappingKeyRangeLocked(&placeholder.rangeDesc); it.item != nil {
		__antithesis_instrumentation__.Notify(125753)
		return false, errors.AssertionFailedf("corrupted replicasByKey map: %+v and %+v overlapped", it.ph, it.item)
	} else {
		__antithesis_instrumentation__.Notify(125754)
	}
	__antithesis_instrumentation__.Notify(125736)
	switch typ {
	case removePlaceholderDropped:
		__antithesis_instrumentation__.Notify(125755)
		atomic.AddInt32(&s.counts.droppedPlaceholders, 1)
	case removePlaceholderFailed:
		__antithesis_instrumentation__.Notify(125756)
		atomic.AddInt32(&s.counts.failedPlaceholders, 1)
	case removePlaceholderFilled:
		__antithesis_instrumentation__.Notify(125757)
		atomic.AddInt32(&s.counts.filledPlaceholders, 1)
	default:
		__antithesis_instrumentation__.Notify(125758)
	}
	__antithesis_instrumentation__.Notify(125737)
	return true, nil
}
