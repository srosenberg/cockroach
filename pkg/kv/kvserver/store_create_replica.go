package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

var errRetry = errors.New("retry: orphaned replica")

func (s *Store) getOrCreateReplica(
	ctx context.Context,
	rangeID roachpb.RangeID,
	replicaID roachpb.ReplicaID,
	creatingReplica *roachpb.ReplicaDescriptor,
) (_ *Replica, created bool, _ error) {
	__antithesis_instrumentation__.Notify(124790)
	if replicaID == 0 {
		__antithesis_instrumentation__.Notify(124792)
		log.Fatalf(ctx, "cannot construct a Replica for range %d with 0 id", rangeID)
	} else {
		__antithesis_instrumentation__.Notify(124793)
	}
	__antithesis_instrumentation__.Notify(124791)

	r := retry.Start(retry.Options{
		InitialBackoff: time.Microsecond,

		MaxBackoff: 10 * time.Millisecond,
	})
	for {
		__antithesis_instrumentation__.Notify(124794)
		r.Next()
		r, created, err := s.tryGetOrCreateReplica(
			ctx,
			rangeID,
			replicaID,
			creatingReplica,
		)
		if errors.Is(err, errRetry) {
			__antithesis_instrumentation__.Notify(124797)
			continue
		} else {
			__antithesis_instrumentation__.Notify(124798)
		}
		__antithesis_instrumentation__.Notify(124795)
		if err != nil {
			__antithesis_instrumentation__.Notify(124799)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(124800)
		}
		__antithesis_instrumentation__.Notify(124796)
		return r, created, err
	}
}

func (s *Store) tryGetOrCreateReplica(
	ctx context.Context,
	rangeID roachpb.RangeID,
	replicaID roachpb.ReplicaID,
	creatingReplica *roachpb.ReplicaDescriptor,
) (_ *Replica, created bool, _ error) {
	__antithesis_instrumentation__.Notify(124801)

	if repl, ok := s.mu.replicasByRangeID.Load(rangeID); ok {
		__antithesis_instrumentation__.Notify(124806)
		repl.raftMu.Lock()
		repl.mu.RLock()

		if repl.mu.destroyStatus.Removed() {
			__antithesis_instrumentation__.Notify(124812)
			repl.mu.RUnlock()
			repl.raftMu.Unlock()
			return nil, false, errRetry
		} else {
			__antithesis_instrumentation__.Notify(124813)
		}
		__antithesis_instrumentation__.Notify(124807)

		if fromReplicaIsTooOldRLocked(repl, creatingReplica) {
			__antithesis_instrumentation__.Notify(124814)
			repl.mu.RUnlock()
			repl.raftMu.Unlock()
			return nil, false, roachpb.NewReplicaTooOldError(creatingReplica.ReplicaID)
		} else {
			__antithesis_instrumentation__.Notify(124815)
		}
		__antithesis_instrumentation__.Notify(124808)

		if toTooOld := repl.replicaID < replicaID; toTooOld {
			__antithesis_instrumentation__.Notify(124816)
			if shouldLog := log.V(1); shouldLog {
				__antithesis_instrumentation__.Notify(124819)
				log.Infof(ctx, "found message for replica ID %d which is newer than %v",
					replicaID, repl)
			} else {
				__antithesis_instrumentation__.Notify(124820)
			}
			__antithesis_instrumentation__.Notify(124817)

			repl.mu.RUnlock()
			if err := s.removeReplicaRaftMuLocked(ctx, repl, replicaID, RemoveOptions{
				DestroyData: true,
			}); err != nil {
				__antithesis_instrumentation__.Notify(124821)
				log.Fatalf(ctx, "failed to remove replica: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(124822)
			}
			__antithesis_instrumentation__.Notify(124818)
			repl.raftMu.Unlock()
			return nil, false, errRetry
		} else {
			__antithesis_instrumentation__.Notify(124823)
		}
		__antithesis_instrumentation__.Notify(124809)
		defer repl.mu.RUnlock()

		if repl.replicaID > replicaID {
			__antithesis_instrumentation__.Notify(124824)

			repl.raftMu.Unlock()
			return nil, false, &roachpb.RaftGroupDeletedError{}
		} else {
			__antithesis_instrumentation__.Notify(124825)
		}
		__antithesis_instrumentation__.Notify(124810)
		if repl.replicaID != replicaID {
			__antithesis_instrumentation__.Notify(124826)

			log.Fatalf(ctx, "intended replica id %d unexpectedly does not match the current replica %v",
				replicaID, repl)
		} else {
			__antithesis_instrumentation__.Notify(124827)
		}
		__antithesis_instrumentation__.Notify(124811)
		return repl, false, nil
	} else {
		__antithesis_instrumentation__.Notify(124828)
	}
	__antithesis_instrumentation__.Notify(124802)

	tombstoneKey := keys.RangeTombstoneKey(rangeID)
	var tombstone roachpb.RangeTombstone
	if ok, err := storage.MVCCGetProto(
		ctx, s.Engine(), tombstoneKey, hlc.Timestamp{}, &tombstone, storage.MVCCGetOptions{},
	); err != nil {
		__antithesis_instrumentation__.Notify(124829)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(124830)
		if ok && func() bool {
			__antithesis_instrumentation__.Notify(124831)
			return replicaID != 0 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(124832)
			return replicaID < tombstone.NextReplicaID == true
		}() == true {
			__antithesis_instrumentation__.Notify(124833)
			return nil, false, &roachpb.RaftGroupDeletedError{}
		} else {
			__antithesis_instrumentation__.Notify(124834)
		}
	}
	__antithesis_instrumentation__.Notify(124803)

	uninitializedDesc := &roachpb.RangeDescriptor{
		RangeID: rangeID,
	}
	repl := newUnloadedReplica(ctx, uninitializedDesc, s, replicaID)
	repl.creatingReplica = creatingReplica
	repl.raftMu.Lock()

	repl.readOnlyCmdMu.Lock()

	s.mu.Lock()

	repl.mu.Lock()
	repl.mu.tombstoneMinReplicaID = tombstone.NextReplicaID

	repl.mu.state.Desc = uninitializedDesc

	if err := s.addReplicaToRangeMapLocked(repl); err != nil {
		__antithesis_instrumentation__.Notify(124835)
		repl.mu.Unlock()
		s.mu.Unlock()
		repl.readOnlyCmdMu.Unlock()
		repl.raftMu.Unlock()
		return nil, false, errRetry
	} else {
		__antithesis_instrumentation__.Notify(124836)
	}
	__antithesis_instrumentation__.Notify(124804)
	s.mu.uninitReplicas[repl.RangeID] = repl
	s.mu.Unlock()

	if err := func() error {
		__antithesis_instrumentation__.Notify(124837)

		if ok, err := storage.MVCCGetProto(
			ctx, s.Engine(), tombstoneKey, hlc.Timestamp{}, &tombstone, storage.MVCCGetOptions{},
		); err != nil {
			__antithesis_instrumentation__.Notify(124841)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124842)
			if ok && func() bool {
				__antithesis_instrumentation__.Notify(124843)
				return replicaID < tombstone.NextReplicaID == true
			}() == true {
				__antithesis_instrumentation__.Notify(124844)
				return &roachpb.RaftGroupDeletedError{}
			} else {
				__antithesis_instrumentation__.Notify(124845)
			}
		}
		__antithesis_instrumentation__.Notify(124838)

		if hs, err := repl.mu.stateLoader.LoadHardState(ctx, s.Engine()); err != nil {
			__antithesis_instrumentation__.Notify(124846)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124847)
			if hs.Commit != 0 {
				__antithesis_instrumentation__.Notify(124848)
				log.Fatalf(ctx, "found non-zero HardState.Commit on uninitialized replica %s. HS=%+v", repl, hs)
			} else {
				__antithesis_instrumentation__.Notify(124849)
			}
		}
		__antithesis_instrumentation__.Notify(124839)

		if err := repl.mu.stateLoader.SetRaftReplicaID(ctx, s.Engine(), replicaID); err != nil {
			__antithesis_instrumentation__.Notify(124850)
			return err
		} else {
			__antithesis_instrumentation__.Notify(124851)
		}
		__antithesis_instrumentation__.Notify(124840)

		return repl.loadRaftMuLockedReplicaMuLocked(uninitializedDesc)
	}(); err != nil {
		__antithesis_instrumentation__.Notify(124852)

		repl.mu.destroyStatus.Set(errors.Wrapf(err, "%s: failed to initialize", repl), destroyReasonRemoved)
		repl.mu.Unlock()
		s.mu.Lock()
		s.unlinkReplicaByRangeIDLocked(ctx, rangeID)
		s.mu.Unlock()
		repl.readOnlyCmdMu.Unlock()
		repl.raftMu.Unlock()
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(124853)
	}
	__antithesis_instrumentation__.Notify(124805)
	repl.mu.Unlock()
	repl.readOnlyCmdMu.Unlock()
	return repl, true, nil
}

func fromReplicaIsTooOldRLocked(toReplica *Replica, fromReplica *roachpb.ReplicaDescriptor) bool {
	__antithesis_instrumentation__.Notify(124854)
	toReplica.mu.AssertRHeld()
	if fromReplica == nil {
		__antithesis_instrumentation__.Notify(124856)
		return false
	} else {
		__antithesis_instrumentation__.Notify(124857)
	}
	__antithesis_instrumentation__.Notify(124855)
	desc := toReplica.mu.state.Desc
	_, found := desc.GetReplicaDescriptorByID(fromReplica.ReplicaID)
	return !found && func() bool {
		__antithesis_instrumentation__.Notify(124858)
		return fromReplica.ReplicaID < desc.NextReplicaID == true
	}() == true
}

func (s *Store) addReplicaInternalLocked(repl *Replica) error {
	__antithesis_instrumentation__.Notify(124859)
	if !repl.IsInitialized() {
		__antithesis_instrumentation__.Notify(124864)
		return errors.Errorf("attempted to add uninitialized replica %s", repl)
	} else {
		__antithesis_instrumentation__.Notify(124865)
	}
	__antithesis_instrumentation__.Notify(124860)

	if err := s.addReplicaToRangeMapLocked(repl); err != nil {
		__antithesis_instrumentation__.Notify(124866)
		return err
	} else {
		__antithesis_instrumentation__.Notify(124867)
	}
	__antithesis_instrumentation__.Notify(124861)

	if it := s.getOverlappingKeyRangeLocked(repl.Desc()); it.item != nil {
		__antithesis_instrumentation__.Notify(124868)
		return errors.Errorf("%s: cannot addReplicaInternalLocked; range %s has overlapping range %s", s, repl, it.Desc())
	} else {
		__antithesis_instrumentation__.Notify(124869)
	}
	__antithesis_instrumentation__.Notify(124862)

	if it := s.mu.replicasByKey.ReplaceOrInsertReplica(context.Background(), repl); it.item != nil {
		__antithesis_instrumentation__.Notify(124870)
		return errors.Errorf("%s: cannot addReplicaInternalLocked; range for key %v already exists in replicasByKey btree", s,
			it.item.key())
	} else {
		__antithesis_instrumentation__.Notify(124871)
	}
	__antithesis_instrumentation__.Notify(124863)

	return nil
}

func (s *Store) addPlaceholderLocked(placeholder *ReplicaPlaceholder) error {
	__antithesis_instrumentation__.Notify(124872)
	rangeID := placeholder.Desc().RangeID
	if it := s.mu.replicasByKey.ReplaceOrInsertPlaceholder(context.Background(), placeholder); it.item != nil {
		__antithesis_instrumentation__.Notify(124875)
		return errors.Errorf("%s overlaps with existing replicaOrPlaceholder %+v in replicasByKey btree", placeholder, it.item)
	} else {
		__antithesis_instrumentation__.Notify(124876)
	}
	__antithesis_instrumentation__.Notify(124873)
	if exRng, ok := s.mu.replicaPlaceholders[rangeID]; ok {
		__antithesis_instrumentation__.Notify(124877)
		return errors.Errorf("%s has ID collision with placeholder %+v", placeholder, exRng)
	} else {
		__antithesis_instrumentation__.Notify(124878)
	}
	__antithesis_instrumentation__.Notify(124874)
	s.mu.replicaPlaceholders[rangeID] = placeholder
	return nil
}

func (s *Store) addReplicaToRangeMapLocked(repl *Replica) error {
	__antithesis_instrumentation__.Notify(124879)

	if existing, loaded := s.mu.replicasByRangeID.LoadOrStore(
		repl.RangeID, repl); loaded && func() bool {
		__antithesis_instrumentation__.Notify(124882)
		return existing != repl == true
	}() == true {
		__antithesis_instrumentation__.Notify(124883)
		return errors.Errorf("%s: replica already exists", repl)
	} else {
		__antithesis_instrumentation__.Notify(124884)
	}
	__antithesis_instrumentation__.Notify(124880)

	if !repl.mu.quiescent {
		__antithesis_instrumentation__.Notify(124885)
		s.unquiescedReplicas.Lock()
		s.unquiescedReplicas.m[repl.RangeID] = struct{}{}
		s.unquiescedReplicas.Unlock()
	} else {
		__antithesis_instrumentation__.Notify(124886)
	}
	__antithesis_instrumentation__.Notify(124881)
	return nil
}

func (s *Store) maybeMarkReplicaInitializedLockedReplLocked(
	ctx context.Context, lockedRepl *Replica,
) error {
	__antithesis_instrumentation__.Notify(124887)
	desc := lockedRepl.descRLocked()
	if !desc.IsInitialized() {
		__antithesis_instrumentation__.Notify(124893)
		return errors.Errorf("attempted to process uninitialized range %s", desc)
	} else {
		__antithesis_instrumentation__.Notify(124894)
	}
	__antithesis_instrumentation__.Notify(124888)

	rangeID := lockedRepl.RangeID
	if _, ok := s.mu.uninitReplicas[rangeID]; !ok {
		__antithesis_instrumentation__.Notify(124895)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(124896)
	}
	__antithesis_instrumentation__.Notify(124889)
	delete(s.mu.uninitReplicas, rangeID)

	if it := s.getOverlappingKeyRangeLocked(desc); it.item != nil {
		__antithesis_instrumentation__.Notify(124897)
		return errors.AssertionFailedf("%s: cannot initialize replica; %s has overlapping range %s",
			s, desc, it.Desc())
	} else {
		__antithesis_instrumentation__.Notify(124898)
	}
	__antithesis_instrumentation__.Notify(124890)

	lockedRepl.setStartKeyLocked(desc.StartKey)
	if it := s.mu.replicasByKey.ReplaceOrInsertReplica(ctx, lockedRepl); it.item != nil {
		__antithesis_instrumentation__.Notify(124899)
		return errors.AssertionFailedf("range for key %v already exists in replicasByKey btree: %+v",
			it.item.key(), it)
	} else {
		__antithesis_instrumentation__.Notify(124900)
	}
	__antithesis_instrumentation__.Notify(124891)

	if !lockedRepl.maybeUnquiesceWithOptionsLocked(false) {
		__antithesis_instrumentation__.Notify(124901)
		return errors.AssertionFailedf("expected replica %s to unquiesce after initialization", desc)
	} else {
		__antithesis_instrumentation__.Notify(124902)
	}
	__antithesis_instrumentation__.Notify(124892)

	s.metrics.ReplicaCount.Inc(1)
	s.maybeGossipOnCapacityChange(ctx, rangeAddEvent)

	return nil
}
