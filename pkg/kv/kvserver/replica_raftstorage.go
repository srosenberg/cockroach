package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"
	"time"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/raftentry"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/rditer"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/readsummary"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/iterutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type replicaRaftStorage Replica

var _ raft.Storage = (*replicaRaftStorage)(nil)

func (r *replicaRaftStorage) InitialState() (raftpb.HardState, raftpb.ConfState, error) {
	__antithesis_instrumentation__.Notify(119184)
	ctx := r.AnnotateCtx(context.TODO())
	hs, err := r.mu.stateLoader.LoadHardState(ctx, r.store.Engine())

	if raft.IsEmptyHardState(hs) || func() bool {
		__antithesis_instrumentation__.Notify(119186)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(119187)
		return raftpb.HardState{}, raftpb.ConfState{}, err
	} else {
		__antithesis_instrumentation__.Notify(119188)
	}
	__antithesis_instrumentation__.Notify(119185)
	cs := r.mu.state.Desc.Replicas().ConfState()
	return hs, cs, nil
}

func (r *replicaRaftStorage) Entries(lo, hi, maxBytes uint64) ([]raftpb.Entry, error) {
	__antithesis_instrumentation__.Notify(119189)
	readonly := r.store.Engine().NewReadOnly(storage.StandardDurability)
	defer readonly.Close()
	ctx := r.AnnotateCtx(context.TODO())
	if r.raftMu.sideloaded == nil {
		__antithesis_instrumentation__.Notify(119191)
		return nil, errors.New("sideloaded storage is uninitialized")
	} else {
		__antithesis_instrumentation__.Notify(119192)
	}
	__antithesis_instrumentation__.Notify(119190)
	return entries(ctx, r.mu.stateLoader, readonly, r.RangeID, r.store.raftEntryCache,
		r.raftMu.sideloaded, lo, hi, maxBytes)
}

func (r *Replica) raftEntriesLocked(lo, hi, maxBytes uint64) ([]raftpb.Entry, error) {
	__antithesis_instrumentation__.Notify(119193)
	return (*replicaRaftStorage)(r).Entries(lo, hi, maxBytes)
}

func entries(
	ctx context.Context,
	rsl stateloader.StateLoader,
	reader storage.Reader,
	rangeID roachpb.RangeID,
	eCache *raftentry.Cache,
	sideloaded SideloadStorage,
	lo, hi, maxBytes uint64,
) ([]raftpb.Entry, error) {
	__antithesis_instrumentation__.Notify(119194)
	if lo > hi {
		__antithesis_instrumentation__.Notify(119206)
		return nil, errors.Errorf("lo:%d is greater than hi:%d", lo, hi)
	} else {
		__antithesis_instrumentation__.Notify(119207)
	}
	__antithesis_instrumentation__.Notify(119195)

	n := hi - lo
	if n > 100 {
		__antithesis_instrumentation__.Notify(119208)
		n = 100
	} else {
		__antithesis_instrumentation__.Notify(119209)
	}
	__antithesis_instrumentation__.Notify(119196)
	ents := make([]raftpb.Entry, 0, n)

	ents, size, hitIndex, exceededMaxBytes := eCache.Scan(ents, rangeID, lo, hi, maxBytes)

	if uint64(len(ents)) == hi-lo || func() bool {
		__antithesis_instrumentation__.Notify(119210)
		return exceededMaxBytes == true
	}() == true {
		__antithesis_instrumentation__.Notify(119211)
		return ents, nil
	} else {
		__antithesis_instrumentation__.Notify(119212)
	}
	__antithesis_instrumentation__.Notify(119197)

	expectedIndex := hitIndex

	canCache := true

	scanFunc := func(ent raftpb.Entry) error {
		__antithesis_instrumentation__.Notify(119213)

		if ent.Index != expectedIndex {
			__antithesis_instrumentation__.Notify(119218)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(119219)
		}
		__antithesis_instrumentation__.Notify(119214)
		expectedIndex++

		if sniffSideloadedRaftCommand(ent.Data) {
			__antithesis_instrumentation__.Notify(119220)
			canCache = canCache && func() bool {
				__antithesis_instrumentation__.Notify(119221)
				return sideloaded != nil == true
			}() == true
			if sideloaded != nil {
				__antithesis_instrumentation__.Notify(119222)
				newEnt, err := maybeInlineSideloadedRaftCommand(
					ctx, rangeID, ent, sideloaded, eCache,
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(119224)
					return err
				} else {
					__antithesis_instrumentation__.Notify(119225)
				}
				__antithesis_instrumentation__.Notify(119223)
				if newEnt != nil {
					__antithesis_instrumentation__.Notify(119226)
					ent = *newEnt
				} else {
					__antithesis_instrumentation__.Notify(119227)
				}
			} else {
				__antithesis_instrumentation__.Notify(119228)
			}
		} else {
			__antithesis_instrumentation__.Notify(119229)
		}
		__antithesis_instrumentation__.Notify(119215)

		size += uint64(ent.Size())
		if size > maxBytes {
			__antithesis_instrumentation__.Notify(119230)
			exceededMaxBytes = true
			if len(ents) > 0 {
				__antithesis_instrumentation__.Notify(119231)
				if exceededMaxBytes {
					__antithesis_instrumentation__.Notify(119233)
					return iterutil.StopIteration()
				} else {
					__antithesis_instrumentation__.Notify(119234)
				}
				__antithesis_instrumentation__.Notify(119232)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(119235)
			}
		} else {
			__antithesis_instrumentation__.Notify(119236)
		}
		__antithesis_instrumentation__.Notify(119216)
		ents = append(ents, ent)
		if exceededMaxBytes {
			__antithesis_instrumentation__.Notify(119237)
			return iterutil.StopIteration()
		} else {
			__antithesis_instrumentation__.Notify(119238)
		}
		__antithesis_instrumentation__.Notify(119217)
		return nil
	}
	__antithesis_instrumentation__.Notify(119198)

	if err := iterateEntries(ctx, reader, rangeID, expectedIndex, hi, scanFunc); err != nil {
		__antithesis_instrumentation__.Notify(119239)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(119240)
	}
	__antithesis_instrumentation__.Notify(119199)

	if canCache {
		__antithesis_instrumentation__.Notify(119241)
		eCache.Add(rangeID, ents, false)
	} else {
		__antithesis_instrumentation__.Notify(119242)
	}
	__antithesis_instrumentation__.Notify(119200)

	if uint64(len(ents)) == hi-lo {
		__antithesis_instrumentation__.Notify(119243)
		return ents, nil
	} else {
		__antithesis_instrumentation__.Notify(119244)
	}
	__antithesis_instrumentation__.Notify(119201)

	if exceededMaxBytes {
		__antithesis_instrumentation__.Notify(119245)
		return ents, nil
	} else {
		__antithesis_instrumentation__.Notify(119246)
	}
	__antithesis_instrumentation__.Notify(119202)

	if len(ents) > 0 {
		__antithesis_instrumentation__.Notify(119247)

		if ents[0].Index > lo {
			__antithesis_instrumentation__.Notify(119251)
			return nil, raft.ErrCompacted
		} else {
			__antithesis_instrumentation__.Notify(119252)
		}
		__antithesis_instrumentation__.Notify(119248)

		lastIndex, err := rsl.LoadLastIndex(ctx, reader)
		if err != nil {
			__antithesis_instrumentation__.Notify(119253)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(119254)
		}
		__antithesis_instrumentation__.Notify(119249)
		if lastIndex <= expectedIndex {
			__antithesis_instrumentation__.Notify(119255)
			return nil, raft.ErrUnavailable
		} else {
			__antithesis_instrumentation__.Notify(119256)
		}
		__antithesis_instrumentation__.Notify(119250)

		return nil, errors.Errorf("there is a gap in the index record between lo:%d and hi:%d at index:%d", lo, hi, expectedIndex)
	} else {
		__antithesis_instrumentation__.Notify(119257)
	}
	__antithesis_instrumentation__.Notify(119203)

	ts, err := rsl.LoadRaftTruncatedState(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(119258)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(119259)
	}
	__antithesis_instrumentation__.Notify(119204)
	if ts.Index >= lo {
		__antithesis_instrumentation__.Notify(119260)

		return nil, raft.ErrCompacted
	} else {
		__antithesis_instrumentation__.Notify(119261)
	}
	__antithesis_instrumentation__.Notify(119205)

	return nil, raft.ErrUnavailable
}

func iterateEntries(
	ctx context.Context,
	reader storage.Reader,
	rangeID roachpb.RangeID,
	lo, hi uint64,
	f func(raftpb.Entry) error,
) error {
	__antithesis_instrumentation__.Notify(119262)
	key := keys.RaftLogKey(rangeID, lo)
	endKey := keys.RaftLogKey(rangeID, hi)
	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		UpperBound: endKey,
	})
	defer iter.Close()

	var meta enginepb.MVCCMetadata
	var ent raftpb.Entry

	iter.SeekGE(storage.MakeMVCCMetadataKey(key))
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(119263)
		if ok, err := iter.Valid(); err != nil || func() bool {
			__antithesis_instrumentation__.Notify(119267)
			return !ok == true
		}() == true {
			__antithesis_instrumentation__.Notify(119268)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119269)
		}
		__antithesis_instrumentation__.Notify(119264)

		if err := protoutil.Unmarshal(iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(119270)
			return errors.Wrap(err, "unable to decode MVCCMetadata")
		} else {
			__antithesis_instrumentation__.Notify(119271)
		}
		__antithesis_instrumentation__.Notify(119265)
		if err := storage.MakeValue(meta).GetProto(&ent); err != nil {
			__antithesis_instrumentation__.Notify(119272)
			return errors.Wrap(err, "unable to unmarshal raft Entry")
		} else {
			__antithesis_instrumentation__.Notify(119273)
		}
		__antithesis_instrumentation__.Notify(119266)
		if err := f(ent); err != nil {
			__antithesis_instrumentation__.Notify(119274)
			if iterutil.Done(err) {
				__antithesis_instrumentation__.Notify(119276)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(119277)
			}
			__antithesis_instrumentation__.Notify(119275)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119278)
		}
	}
}

const invalidLastTerm = 0

func (r *replicaRaftStorage) Term(i uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(119279)

	if r.mu.lastIndex == i && func() bool {
		__antithesis_instrumentation__.Notify(119282)
		return r.mu.lastTerm != invalidLastTerm == true
	}() == true {
		__antithesis_instrumentation__.Notify(119283)
		return r.mu.lastTerm, nil
	} else {
		__antithesis_instrumentation__.Notify(119284)
	}
	__antithesis_instrumentation__.Notify(119280)

	if e, ok := r.store.raftEntryCache.Get(r.RangeID, i); ok {
		__antithesis_instrumentation__.Notify(119285)
		return e.Term, nil
	} else {
		__antithesis_instrumentation__.Notify(119286)
	}
	__antithesis_instrumentation__.Notify(119281)
	readonly := r.store.Engine().NewReadOnly(storage.StandardDurability)
	defer readonly.Close()
	ctx := r.AnnotateCtx(context.TODO())
	return term(ctx, r.mu.stateLoader, readonly, r.RangeID, r.store.raftEntryCache, i)
}

func (r *Replica) raftTermRLocked(i uint64) (uint64, error) {
	__antithesis_instrumentation__.Notify(119287)
	return (*replicaRaftStorage)(r).Term(i)
}

func term(
	ctx context.Context,
	rsl stateloader.StateLoader,
	reader storage.Reader,
	rangeID roachpb.RangeID,
	eCache *raftentry.Cache,
	i uint64,
) (uint64, error) {
	__antithesis_instrumentation__.Notify(119288)

	ents, err := entries(ctx, rsl, reader, rangeID, eCache, nil, i, i+1, math.MaxUint64)
	if errors.Is(err, raft.ErrCompacted) {
		__antithesis_instrumentation__.Notify(119291)
		ts, err := rsl.LoadRaftTruncatedState(ctx, reader)
		if err != nil {
			__antithesis_instrumentation__.Notify(119294)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(119295)
		}
		__antithesis_instrumentation__.Notify(119292)
		if i == ts.Index {
			__antithesis_instrumentation__.Notify(119296)
			return ts.Term, nil
		} else {
			__antithesis_instrumentation__.Notify(119297)
		}
		__antithesis_instrumentation__.Notify(119293)
		return 0, raft.ErrCompacted
	} else {
		__antithesis_instrumentation__.Notify(119298)
		if err != nil {
			__antithesis_instrumentation__.Notify(119299)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(119300)
		}
	}
	__antithesis_instrumentation__.Notify(119289)
	if len(ents) == 0 {
		__antithesis_instrumentation__.Notify(119301)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(119302)
	}
	__antithesis_instrumentation__.Notify(119290)
	return ents[0].Term, nil
}

func (r *replicaRaftStorage) LastIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(119303)
	return r.mu.lastIndex, nil
}

func (r *Replica) raftLastIndexLocked() (uint64, error) {
	__antithesis_instrumentation__.Notify(119304)
	return (*replicaRaftStorage)(r).LastIndex()
}

func (r *Replica) raftTruncatedStateLocked(
	ctx context.Context,
) (roachpb.RaftTruncatedState, error) {
	__antithesis_instrumentation__.Notify(119305)
	if r.mu.state.TruncatedState != nil {
		__antithesis_instrumentation__.Notify(119309)
		return *r.mu.state.TruncatedState, nil
	} else {
		__antithesis_instrumentation__.Notify(119310)
	}
	__antithesis_instrumentation__.Notify(119306)
	ts, err := r.mu.stateLoader.LoadRaftTruncatedState(ctx, r.store.Engine())
	if err != nil {
		__antithesis_instrumentation__.Notify(119311)
		return ts, err
	} else {
		__antithesis_instrumentation__.Notify(119312)
	}
	__antithesis_instrumentation__.Notify(119307)
	if ts.Index != 0 {
		__antithesis_instrumentation__.Notify(119313)
		r.mu.state.TruncatedState = &ts
	} else {
		__antithesis_instrumentation__.Notify(119314)
	}
	__antithesis_instrumentation__.Notify(119308)
	return ts, nil
}

func (r *replicaRaftStorage) FirstIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(119315)
	ctx := r.AnnotateCtx(context.TODO())
	ts, err := (*Replica)(r).raftTruncatedStateLocked(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(119317)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(119318)
	}
	__antithesis_instrumentation__.Notify(119316)
	return ts.Index + 1, nil
}

func (r *Replica) raftFirstIndexLocked() (uint64, error) {
	__antithesis_instrumentation__.Notify(119319)
	return (*replicaRaftStorage)(r).FirstIndex()
}

func (r *Replica) GetFirstIndex() (uint64, error) {
	__antithesis_instrumentation__.Notify(119320)
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.raftFirstIndexLocked()
}

func (r *Replica) GetLeaseAppliedIndex() uint64 {
	__antithesis_instrumentation__.Notify(119321)
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mu.state.LeaseAppliedIndex
}

func (r *replicaRaftStorage) Snapshot() (raftpb.Snapshot, error) {
	__antithesis_instrumentation__.Notify(119322)
	r.mu.AssertHeld()
	appliedIndex := r.mu.state.RaftAppliedIndex
	term, err := r.Term(appliedIndex)
	if err != nil {
		__antithesis_instrumentation__.Notify(119324)
		return raftpb.Snapshot{}, err
	} else {
		__antithesis_instrumentation__.Notify(119325)
	}
	__antithesis_instrumentation__.Notify(119323)
	return raftpb.Snapshot{
		Metadata: raftpb.SnapshotMetadata{
			Index: appliedIndex,
			Term:  term,
		},
	}, nil
}

func (r *Replica) raftSnapshotLocked() (raftpb.Snapshot, error) {
	__antithesis_instrumentation__.Notify(119326)
	return (*replicaRaftStorage)(r).Snapshot()
}

func (r *Replica) GetSnapshot(
	ctx context.Context, snapType kvserverpb.SnapshotRequest_Type, recipientStore roachpb.StoreID,
) (_ *OutgoingSnapshot, err error) {
	__antithesis_instrumentation__.Notify(119327)
	snapUUID := uuid.MakeV4()

	r.raftMu.Lock()
	snap := r.store.engine.NewSnapshot()
	{
		__antithesis_instrumentation__.Notify(119333)
		r.mu.Lock()

		appliedIndex := r.mu.state.RaftAppliedIndex

		r.addSnapshotLogTruncationConstraintLocked(ctx, snapUUID, appliedIndex, recipientStore)
		r.mu.Unlock()
	}
	__antithesis_instrumentation__.Notify(119328)
	r.raftMu.Unlock()

	release := func() {
		__antithesis_instrumentation__.Notify(119334)
		now := timeutil.Now()
		r.completeSnapshotLogTruncationConstraint(ctx, snapUUID, now)
	}
	__antithesis_instrumentation__.Notify(119329)

	defer func() {
		__antithesis_instrumentation__.Notify(119335)
		if err != nil {
			__antithesis_instrumentation__.Notify(119336)
			release()
			snap.Close()
		} else {
			__antithesis_instrumentation__.Notify(119337)
		}
	}()
	__antithesis_instrumentation__.Notify(119330)

	r.mu.RLock()
	defer r.mu.RUnlock()
	rangeID := r.RangeID

	startKey := r.mu.state.Desc.StartKey
	ctx, sp := r.AnnotateCtxWithSpan(ctx, "snapshot")
	defer sp.Finish()

	log.Eventf(ctx, "new engine snapshot for replica %s", r)

	withSideloaded := func(fn func(SideloadStorage) error) error {
		__antithesis_instrumentation__.Notify(119338)
		r.raftMu.Lock()
		defer r.raftMu.Unlock()
		return fn(r.raftMu.sideloaded)
	}
	__antithesis_instrumentation__.Notify(119331)

	snapData, err := snapshot(
		ctx, snapUUID, stateloader.Make(rangeID), snapType,
		snap, rangeID, r.store.raftEntryCache, withSideloaded, startKey,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(119339)
		log.Errorf(ctx, "error generating snapshot: %+v", err)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(119340)
	}
	__antithesis_instrumentation__.Notify(119332)
	snapData.onClose = release
	return &snapData, nil
}

type OutgoingSnapshot struct {
	SnapUUID uuid.UUID

	RaftSnap raftpb.Snapshot

	EngineSnap storage.Reader

	Iter *rditer.ReplicaEngineDataIterator

	State kvserverpb.ReplicaState

	WithSideloaded func(func(SideloadStorage) error) error
	RaftEntryCache *raftentry.Cache
	snapType       kvserverpb.SnapshotRequest_Type
	onClose        func()
}

func (s OutgoingSnapshot) String() string {
	__antithesis_instrumentation__.Notify(119341)
	return redact.StringWithoutMarkers(s)
}

func (s OutgoingSnapshot) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(119342)
	w.Printf("%s snapshot %s at applied index %d",
		s.snapType, redact.Safe(s.SnapUUID.Short()), s.State.RaftAppliedIndex)
}

func (s *OutgoingSnapshot) Close() {
	__antithesis_instrumentation__.Notify(119343)
	s.Iter.Close()
	s.EngineSnap.Close()
	if s.onClose != nil {
		__antithesis_instrumentation__.Notify(119344)
		s.onClose()
	} else {
		__antithesis_instrumentation__.Notify(119345)
	}
}

type IncomingSnapshot struct {
	SnapUUID uuid.UUID

	SSTStorageScratch *SSTSnapshotStorageScratch
	FromReplica       roachpb.ReplicaDescriptor

	Desc             *roachpb.RangeDescriptor
	DataSize         int64
	snapType         kvserverpb.SnapshotRequest_Type
	placeholder      *ReplicaPlaceholder
	raftAppliedIndex uint64
}

func (s IncomingSnapshot) String() string {
	__antithesis_instrumentation__.Notify(119346)
	return redact.StringWithoutMarkers(s)
}

func (s IncomingSnapshot) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(119347)
	w.Printf("%s snapshot %s from %s at applied index %d",
		s.snapType, redact.Safe(s.SnapUUID.Short()), s.FromReplica, s.raftAppliedIndex)
}

func snapshot(
	ctx context.Context,
	snapUUID uuid.UUID,
	rsl stateloader.StateLoader,
	snapType kvserverpb.SnapshotRequest_Type,
	snap storage.Reader,
	rangeID roachpb.RangeID,
	eCache *raftentry.Cache,
	withSideloaded func(func(SideloadStorage) error) error,
	startKey roachpb.RKey,
) (OutgoingSnapshot, error) {
	__antithesis_instrumentation__.Notify(119348)
	var desc roachpb.RangeDescriptor

	ok, err := storage.MVCCGetProto(ctx, snap, keys.RangeDescriptorKey(startKey),
		hlc.MaxTimestamp, &desc, storage.MVCCGetOptions{Inconsistent: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(119354)
		return OutgoingSnapshot{}, errors.Wrap(err, "failed to get desc")
	} else {
		__antithesis_instrumentation__.Notify(119355)
	}
	__antithesis_instrumentation__.Notify(119349)
	if !ok {
		__antithesis_instrumentation__.Notify(119356)
		return OutgoingSnapshot{}, errors.Mark(errors.Errorf("couldn't find range descriptor"), errMarkSnapshotError)
	} else {
		__antithesis_instrumentation__.Notify(119357)
	}
	__antithesis_instrumentation__.Notify(119350)

	state, err := rsl.Load(ctx, snap, &desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(119358)
		return OutgoingSnapshot{}, err
	} else {
		__antithesis_instrumentation__.Notify(119359)
	}
	__antithesis_instrumentation__.Notify(119351)

	term, err := term(ctx, rsl, snap, rangeID, eCache, state.RaftAppliedIndex)

	if state.RaftAppliedIndexTerm != 0 && func() bool {
		__antithesis_instrumentation__.Notify(119360)
		return term != state.RaftAppliedIndexTerm == true
	}() == true {
		__antithesis_instrumentation__.Notify(119361)
		return OutgoingSnapshot{},
			errors.AssertionFailedf("unequal terms %d != %d", term, state.RaftAppliedIndexTerm)
	} else {
		__antithesis_instrumentation__.Notify(119362)
	}
	__antithesis_instrumentation__.Notify(119352)
	if err != nil {
		__antithesis_instrumentation__.Notify(119363)
		return OutgoingSnapshot{}, errors.Wrapf(err, "failed to fetch term of %d", state.RaftAppliedIndex)
	} else {
		__antithesis_instrumentation__.Notify(119364)
	}
	__antithesis_instrumentation__.Notify(119353)

	iter := rditer.NewReplicaEngineDataIterator(&desc, snap, true)

	return OutgoingSnapshot{
		RaftEntryCache: eCache,
		WithSideloaded: withSideloaded,
		EngineSnap:     snap,
		Iter:           iter,
		State:          state,
		SnapUUID:       snapUUID,
		RaftSnap: raftpb.Snapshot{
			Data: snapUUID.GetBytes(),
			Metadata: raftpb.SnapshotMetadata{
				Index: state.RaftAppliedIndex,
				Term:  term,

				ConfState: desc.Replicas().ConfState(),
			},
		},
		snapType: snapType,
	}, nil
}

func (r *Replica) append(
	ctx context.Context,
	writer storage.Writer,
	prevLastIndex uint64,
	prevLastTerm uint64,
	prevRaftLogSize int64,
	entries []raftpb.Entry,
) (uint64, uint64, int64, error) {
	__antithesis_instrumentation__.Notify(119365)
	if len(entries) == 0 {
		__antithesis_instrumentation__.Notify(119369)
		return prevLastIndex, prevLastTerm, prevRaftLogSize, nil
	} else {
		__antithesis_instrumentation__.Notify(119370)
	}
	__antithesis_instrumentation__.Notify(119366)
	var diff enginepb.MVCCStats
	var value roachpb.Value
	for i := range entries {
		__antithesis_instrumentation__.Notify(119371)
		ent := &entries[i]
		key := r.raftMu.stateLoader.RaftLogKey(ent.Index)

		if err := value.SetProto(ent); err != nil {
			__antithesis_instrumentation__.Notify(119374)
			return 0, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(119375)
		}
		__antithesis_instrumentation__.Notify(119372)
		value.InitChecksum(key)
		var err error
		if ent.Index > prevLastIndex {
			__antithesis_instrumentation__.Notify(119376)
			err = storage.MVCCBlindPut(ctx, writer, &diff, key, hlc.Timestamp{}, value, nil)
		} else {
			__antithesis_instrumentation__.Notify(119377)

			eng, ok := writer.(storage.ReadWriter)
			if !ok {
				__antithesis_instrumentation__.Notify(119379)
				panic("expected writer to be a engine.ReadWriter when overwriting log entries")
			} else {
				__antithesis_instrumentation__.Notify(119380)
			}
			__antithesis_instrumentation__.Notify(119378)
			err = storage.MVCCPut(ctx, eng, &diff, key, hlc.Timestamp{}, value, nil)
		}
		__antithesis_instrumentation__.Notify(119373)
		if err != nil {
			__antithesis_instrumentation__.Notify(119381)
			return 0, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(119382)
		}
	}
	__antithesis_instrumentation__.Notify(119367)

	lastIndex := entries[len(entries)-1].Index
	lastTerm := entries[len(entries)-1].Term

	if prevLastIndex > 0 {
		__antithesis_instrumentation__.Notify(119383)

		eng, ok := writer.(storage.ReadWriter)
		if !ok {
			__antithesis_instrumentation__.Notify(119385)
			panic("expected writer to be a engine.ReadWriter when deleting log entries")
		} else {
			__antithesis_instrumentation__.Notify(119386)
		}
		__antithesis_instrumentation__.Notify(119384)
		for i := lastIndex + 1; i <= prevLastIndex; i++ {
			__antithesis_instrumentation__.Notify(119387)

			err := storage.MVCCDelete(ctx, eng, &diff, r.raftMu.stateLoader.RaftLogKey(i),
				hlc.Timestamp{}, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(119388)
				return 0, 0, 0, err
			} else {
				__antithesis_instrumentation__.Notify(119389)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(119390)
	}
	__antithesis_instrumentation__.Notify(119368)

	raftLogSize := prevRaftLogSize + diff.SysBytes
	return lastIndex, lastTerm, raftLogSize, nil
}

func (r *Replica) updateRangeInfo(ctx context.Context, desc *roachpb.RangeDescriptor) error {
	__antithesis_instrumentation__.Notify(119391)

	confReader, err := r.store.GetConfReader(ctx)
	if errors.Is(err, errSysCfgUnavailable) {
		__antithesis_instrumentation__.Notify(119395)

		log.Warningf(ctx, "unable to retrieve conf reader, cannot determine range MaxBytes")
		return nil
	} else {
		__antithesis_instrumentation__.Notify(119396)
	}
	__antithesis_instrumentation__.Notify(119392)
	if err != nil {
		__antithesis_instrumentation__.Notify(119397)
		return err
	} else {
		__antithesis_instrumentation__.Notify(119398)
	}
	__antithesis_instrumentation__.Notify(119393)

	conf, err := confReader.GetSpanConfigForKey(ctx, desc.StartKey)
	if err != nil {
		__antithesis_instrumentation__.Notify(119399)
		return errors.Wrapf(err, "%s: failed to lookup span config", r)
	} else {
		__antithesis_instrumentation__.Notify(119400)
	}
	__antithesis_instrumentation__.Notify(119394)

	r.SetSpanConfig(conf)
	return nil
}

func clearRangeData(
	desc *roachpb.RangeDescriptor,
	reader storage.Reader,
	writer storage.Writer,
	rangeIDLocalOnly bool,
	mustClearRange bool,
) error {
	__antithesis_instrumentation__.Notify(119401)
	var keyRanges []rditer.KeyRange
	if rangeIDLocalOnly {
		__antithesis_instrumentation__.Notify(119405)
		keyRanges = []rditer.KeyRange{rditer.MakeRangeIDLocalKeyRange(desc.RangeID, false)}
	} else {
		__antithesis_instrumentation__.Notify(119406)
		keyRanges = rditer.MakeAllKeyRanges(desc)
	}
	__antithesis_instrumentation__.Notify(119402)
	var clearRangeFn func(storage.Reader, storage.Writer, roachpb.Key, roachpb.Key) error
	if mustClearRange {
		__antithesis_instrumentation__.Notify(119407)
		clearRangeFn = func(reader storage.Reader, writer storage.Writer, start, end roachpb.Key) error {
			__antithesis_instrumentation__.Notify(119408)
			return writer.ClearRawRange(start, end)
		}
	} else {
		__antithesis_instrumentation__.Notify(119409)
		clearRangeFn = storage.ClearRangeWithHeuristic
	}
	__antithesis_instrumentation__.Notify(119403)

	for _, keyRange := range keyRanges {
		__antithesis_instrumentation__.Notify(119410)
		if err := clearRangeFn(reader, writer, keyRange.Start, keyRange.End); err != nil {
			__antithesis_instrumentation__.Notify(119411)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119412)
		}
	}
	__antithesis_instrumentation__.Notify(119404)
	return nil
}

func (r *Replica) applySnapshot(
	ctx context.Context,
	inSnap IncomingSnapshot,
	nonemptySnap raftpb.Snapshot,
	hs raftpb.HardState,
	subsumedRepls []*Replica,
) (err error) {
	__antithesis_instrumentation__.Notify(119413)
	desc := inSnap.Desc
	if desc.RangeID != r.RangeID {
		__antithesis_instrumentation__.Notify(119438)
		log.Fatalf(ctx, "unexpected range ID %d", desc.RangeID)
	} else {
		__antithesis_instrumentation__.Notify(119439)
	}
	__antithesis_instrumentation__.Notify(119414)

	isInitialSnap := !r.IsInitialized()
	{
		__antithesis_instrumentation__.Notify(119440)
		var from, to roachpb.RKey
		if isInitialSnap {
			__antithesis_instrumentation__.Notify(119441)

			d := inSnap.placeholder.Desc()
			from, to = d.StartKey, d.EndKey
			defer r.store.maybeAssertNoHole(ctx, from, to)()
		} else {
			__antithesis_instrumentation__.Notify(119442)

			d := r.Desc()
			from, to = d.EndKey, inSnap.Desc.EndKey
			defer r.store.maybeAssertNoHole(ctx, from, to)()
		}
	}
	__antithesis_instrumentation__.Notify(119415)
	defer func() {
		__antithesis_instrumentation__.Notify(119443)
		if e := recover(); e != nil {
			__antithesis_instrumentation__.Notify(119445)

			panic(e)
		} else {
			__antithesis_instrumentation__.Notify(119446)
		}
		__antithesis_instrumentation__.Notify(119444)
		if err == nil {
			__antithesis_instrumentation__.Notify(119447)
			desc, err := r.GetReplicaDescriptor()
			if err != nil {
				__antithesis_instrumentation__.Notify(119449)
				log.Fatalf(ctx, "could not fetch replica descriptor for range after applying snapshot: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(119450)
			}
			__antithesis_instrumentation__.Notify(119448)
			if isInitialSnap {
				__antithesis_instrumentation__.Notify(119451)
				r.store.metrics.RangeSnapshotsAppliedForInitialUpreplication.Inc(1)
			} else {
				__antithesis_instrumentation__.Notify(119452)
				switch typ := desc.GetType(); typ {

				case roachpb.VOTER_FULL, roachpb.VOTER_INCOMING, roachpb.VOTER_DEMOTING_LEARNER,
					roachpb.VOTER_OUTGOING, roachpb.LEARNER, roachpb.VOTER_DEMOTING_NON_VOTER:
					__antithesis_instrumentation__.Notify(119453)
					r.store.metrics.RangeSnapshotsAppliedByVoters.Inc(1)
				case roachpb.NON_VOTER:
					__antithesis_instrumentation__.Notify(119454)
					r.store.metrics.RangeSnapshotsAppliedByNonVoters.Inc(1)
				default:
					__antithesis_instrumentation__.Notify(119455)
					log.Fatalf(ctx, "unexpected replica type %s while applying snapshot", typ)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(119456)
		}
	}()
	__antithesis_instrumentation__.Notify(119416)

	if raft.IsEmptyHardState(hs) {
		__antithesis_instrumentation__.Notify(119457)

		log.Fatalf(ctx, "found empty HardState for non-empty Snapshot %+v", nonemptySnap)
	} else {
		__antithesis_instrumentation__.Notify(119458)
	}
	__antithesis_instrumentation__.Notify(119417)

	var stats struct {
		subsumedReplicas time.Time

		ingestion time.Time
	}
	log.Infof(ctx, "applying %s", inSnap)
	defer func(start time.Time) {
		__antithesis_instrumentation__.Notify(119459)
		var logDetails redact.StringBuilder
		logDetails.Printf("total=%0.0fms", timeutil.Since(start).Seconds()*1000)
		logDetails.Printf(" data=%s", humanizeutil.IBytes(inSnap.DataSize))
		if len(subsumedRepls) > 0 {
			__antithesis_instrumentation__.Notify(119461)
			logDetails.Printf(" subsumedReplicas=%d@%0.0fms",
				len(subsumedRepls), stats.subsumedReplicas.Sub(start).Seconds()*1000)
		} else {
			__antithesis_instrumentation__.Notify(119462)
		}
		__antithesis_instrumentation__.Notify(119460)
		logDetails.Printf(" ingestion=%d@%0.0fms", len(inSnap.SSTStorageScratch.SSTs()),
			stats.ingestion.Sub(stats.subsumedReplicas).Seconds()*1000)
		log.Infof(ctx, "applied %s (%s)", inSnap, logDetails)
	}(timeutil.Now())
	__antithesis_instrumentation__.Notify(119418)

	unreplicatedSSTFile := &storage.MemFile{}
	unreplicatedSST := storage.MakeIngestionSSTWriter(
		ctx, r.ClusterSettings(), unreplicatedSSTFile,
	)
	defer unreplicatedSST.Close()

	unreplicatedPrefixKey := keys.MakeRangeIDUnreplicatedPrefix(r.RangeID)
	unreplicatedStart := unreplicatedPrefixKey
	unreplicatedEnd := unreplicatedPrefixKey.PrefixEnd()
	if err = unreplicatedSST.ClearRawRange(unreplicatedStart, unreplicatedEnd); err != nil {
		__antithesis_instrumentation__.Notify(119463)
		return errors.Wrapf(err, "error clearing range of unreplicated SST writer")
	} else {
		__antithesis_instrumentation__.Notify(119464)
	}
	__antithesis_instrumentation__.Notify(119419)

	if err := r.raftMu.stateLoader.SetHardState(ctx, &unreplicatedSST, hs); err != nil {
		__antithesis_instrumentation__.Notify(119465)
		return errors.Wrapf(err, "unable to write HardState to unreplicated SST writer")
	} else {
		__antithesis_instrumentation__.Notify(119466)
	}
	__antithesis_instrumentation__.Notify(119420)

	if err := r.raftMu.stateLoader.SetRaftReplicaID(
		ctx, &unreplicatedSST, r.replicaID); err != nil {
		__antithesis_instrumentation__.Notify(119467)
		return errors.Wrapf(err, "unable to write RaftReplicaID to unreplicated SST writer")
	} else {
		__antithesis_instrumentation__.Notify(119468)
	}
	__antithesis_instrumentation__.Notify(119421)

	r.store.raftEntryCache.Drop(r.RangeID)

	if err := r.raftMu.stateLoader.SetRaftTruncatedState(
		ctx, &unreplicatedSST,
		&roachpb.RaftTruncatedState{
			Index: nonemptySnap.Metadata.Index,
			Term:  nonemptySnap.Metadata.Term,
		},
	); err != nil {
		__antithesis_instrumentation__.Notify(119469)
		return errors.Wrapf(err, "unable to write TruncatedState to unreplicated SST writer")
	} else {
		__antithesis_instrumentation__.Notify(119470)
	}
	__antithesis_instrumentation__.Notify(119422)

	if err := unreplicatedSST.Finish(); err != nil {
		__antithesis_instrumentation__.Notify(119471)
		return err
	} else {
		__antithesis_instrumentation__.Notify(119472)
	}
	__antithesis_instrumentation__.Notify(119423)
	if unreplicatedSST.DataSize > 0 {
		__antithesis_instrumentation__.Notify(119473)

		if err := inSnap.SSTStorageScratch.WriteSST(ctx, unreplicatedSSTFile.Data()); err != nil {
			__antithesis_instrumentation__.Notify(119474)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119475)
		}
	} else {
		__antithesis_instrumentation__.Notify(119476)
	}
	__antithesis_instrumentation__.Notify(119424)

	if err := r.clearSubsumedReplicaDiskData(ctx, inSnap.SSTStorageScratch, desc, subsumedRepls, mergedTombstoneReplicaID); err != nil {
		__antithesis_instrumentation__.Notify(119477)
		return err
	} else {
		__antithesis_instrumentation__.Notify(119478)
	}
	__antithesis_instrumentation__.Notify(119425)
	stats.subsumedReplicas = timeutil.Now()

	if fn := r.store.cfg.TestingKnobs.BeforeSnapshotSSTIngestion; fn != nil {
		__antithesis_instrumentation__.Notify(119479)
		if err := fn(inSnap, inSnap.snapType, inSnap.SSTStorageScratch.SSTs()); err != nil {
			__antithesis_instrumentation__.Notify(119480)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119481)
		}
	} else {
		__antithesis_instrumentation__.Notify(119482)
	}
	__antithesis_instrumentation__.Notify(119426)
	if err := r.store.engine.IngestExternalFiles(ctx, inSnap.SSTStorageScratch.SSTs()); err != nil {
		__antithesis_instrumentation__.Notify(119483)
		return errors.Wrapf(err, "while ingesting %s", inSnap.SSTStorageScratch.SSTs())
	} else {
		__antithesis_instrumentation__.Notify(119484)
	}
	__antithesis_instrumentation__.Notify(119427)
	stats.ingestion = timeutil.Now()

	state, err := stateloader.Make(desc.RangeID).Load(ctx, r.store.engine, desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(119485)
		log.Fatalf(ctx, "unable to load replica state: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(119486)
	}
	__antithesis_instrumentation__.Notify(119428)

	if state.RaftAppliedIndex != nonemptySnap.Metadata.Index {
		__antithesis_instrumentation__.Notify(119487)
		log.Fatalf(ctx, "snapshot RaftAppliedIndex %d doesn't match its metadata index %d",
			state.RaftAppliedIndex, nonemptySnap.Metadata.Index)
	} else {
		__antithesis_instrumentation__.Notify(119488)
	}
	__antithesis_instrumentation__.Notify(119429)

	if state.RaftAppliedIndexTerm != 0 && func() bool {
		__antithesis_instrumentation__.Notify(119489)
		return state.RaftAppliedIndexTerm != nonemptySnap.Metadata.Term == true
	}() == true {
		__antithesis_instrumentation__.Notify(119490)
		log.Fatalf(ctx, "snapshot RaftAppliedIndexTerm %d doesn't match its metadata term %d",
			state.RaftAppliedIndexTerm, nonemptySnap.Metadata.Term)
	} else {
		__antithesis_instrumentation__.Notify(119491)
	}
	__antithesis_instrumentation__.Notify(119430)

	subPHs, err := r.clearSubsumedReplicaInMemoryData(ctx, subsumedRepls, mergedTombstoneReplicaID)
	if err != nil {
		__antithesis_instrumentation__.Notify(119492)
		log.Fatalf(ctx, "failed to clear in-memory data of subsumed replicas while applying snapshot: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(119493)
	}
	__antithesis_instrumentation__.Notify(119431)

	prioReadSum, err := readsummary.Load(ctx, r.store.engine, r.RangeID)
	if err != nil {
		__antithesis_instrumentation__.Notify(119494)
		log.Fatalf(ctx, "failed to read prior read summary after applying snapshot: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(119495)
	}
	__antithesis_instrumentation__.Notify(119432)

	r.store.mu.Lock()
	if inSnap.placeholder != nil {
		__antithesis_instrumentation__.Notify(119496)
		subPHs = append(subPHs, inSnap.placeholder)
	} else {
		__antithesis_instrumentation__.Notify(119497)
	}
	__antithesis_instrumentation__.Notify(119433)
	for _, ph := range subPHs {
		__antithesis_instrumentation__.Notify(119498)
		_, err := r.store.removePlaceholderLocked(ctx, ph, removePlaceholderFilled)
		if err != nil {
			__antithesis_instrumentation__.Notify(119499)
			log.Fatalf(ctx, "unable to remove placeholder %s: %s", ph, err)
		} else {
			__antithesis_instrumentation__.Notify(119500)
		}
	}
	__antithesis_instrumentation__.Notify(119434)

	r.mu.Lock()
	r.setDescLockedRaftMuLocked(ctx, desc)
	if err := r.store.maybeMarkReplicaInitializedLockedReplLocked(ctx, r); err != nil {
		__antithesis_instrumentation__.Notify(119501)
		log.Fatalf(ctx, "unable to mark replica initialized while applying snapshot: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(119502)
	}
	__antithesis_instrumentation__.Notify(119435)

	r.store.mu.Unlock()

	r.mu.lastIndex = state.RaftAppliedIndex

	r.mu.lastTerm = invalidLastTerm
	r.mu.raftLogSize = 0

	r.store.metrics.subtractMVCCStats(ctx, r.tenantMetricsRef, *r.mu.state.Stats)
	r.store.metrics.addMVCCStats(ctx, r.tenantMetricsRef, *state.Stats)
	lastKnownLease := r.mu.state.Lease

	r.mu.state = state

	r.mu.raftLogSizeTrusted = false

	r.leasePostApplyLocked(ctx, lastKnownLease, state.Lease, prioReadSum, allowLeaseJump)

	if len(subsumedRepls) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(119503)
		return state.Lease.Replica.ReplicaID == r.replicaID == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(119504)
		return prioReadSum != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(119505)
		applyReadSummaryToTimestampCache(r.store.tsCache, r.descRLocked(), *prioReadSum)
	} else {
		__antithesis_instrumentation__.Notify(119506)
	}
	__antithesis_instrumentation__.Notify(119436)

	r.concMgr.OnReplicaSnapshotApplied()

	r.mu.Unlock()

	r.mu.RLock()
	r.assertStateRaftMuLockedReplicaMuRLocked(ctx, r.store.Engine())
	r.mu.RUnlock()

	r.disconnectRangefeedWithReason(
		roachpb.RangeFeedRetryError_REASON_RAFT_SNAPSHOT,
	)

	if err := r.updateRangeInfo(ctx, desc); err != nil {
		__antithesis_instrumentation__.Notify(119507)
		log.Fatalf(ctx, "unable to update range info while applying snapshot: %+v", err)
	} else {
		__antithesis_instrumentation__.Notify(119508)
	}
	__antithesis_instrumentation__.Notify(119437)

	return nil
}

func (r *Replica) clearSubsumedReplicaDiskData(
	ctx context.Context,
	scratch *SSTSnapshotStorageScratch,
	desc *roachpb.RangeDescriptor,
	subsumedRepls []*Replica,
	subsumedNextReplicaID roachpb.ReplicaID,
) error {
	__antithesis_instrumentation__.Notify(119509)

	getKeyRanges := rditer.MakeReplicatedKeyRangesExceptRangeID
	keyRanges := getKeyRanges(desc)
	totalKeyRanges := append([]rditer.KeyRange(nil), keyRanges...)
	for _, sr := range subsumedRepls {
		__antithesis_instrumentation__.Notify(119512)

		sr.readOnlyCmdMu.Lock()
		sr.mu.Lock()
		sr.mu.destroyStatus.Set(
			roachpb.NewRangeNotFoundError(sr.RangeID, sr.store.StoreID()),
			destroyReasonRemoved)
		sr.mu.Unlock()
		sr.readOnlyCmdMu.Unlock()

		subsumedReplSSTFile := &storage.MemFile{}
		subsumedReplSST := storage.MakeIngestionSSTWriter(
			ctx, r.ClusterSettings(), subsumedReplSSTFile,
		)
		defer subsumedReplSST.Close()

		if err := sr.preDestroyRaftMuLocked(
			ctx,
			r.store.Engine(),
			&subsumedReplSST,
			subsumedNextReplicaID,
			true,
			true,
		); err != nil {
			__antithesis_instrumentation__.Notify(119516)
			subsumedReplSST.Close()
			return err
		} else {
			__antithesis_instrumentation__.Notify(119517)
		}
		__antithesis_instrumentation__.Notify(119513)
		if err := subsumedReplSST.Finish(); err != nil {
			__antithesis_instrumentation__.Notify(119518)
			return err
		} else {
			__antithesis_instrumentation__.Notify(119519)
		}
		__antithesis_instrumentation__.Notify(119514)
		if subsumedReplSST.DataSize > 0 {
			__antithesis_instrumentation__.Notify(119520)

			if err := scratch.WriteSST(ctx, subsumedReplSSTFile.Data()); err != nil {
				__antithesis_instrumentation__.Notify(119521)
				return err
			} else {
				__antithesis_instrumentation__.Notify(119522)
			}
		} else {
			__antithesis_instrumentation__.Notify(119523)
		}
		__antithesis_instrumentation__.Notify(119515)

		srKeyRanges := getKeyRanges(sr.Desc())

		for i := range srKeyRanges {
			__antithesis_instrumentation__.Notify(119524)
			if srKeyRanges[i].Start.Compare(totalKeyRanges[i].Start) < 0 {
				__antithesis_instrumentation__.Notify(119526)
				totalKeyRanges[i].Start = srKeyRanges[i].Start
			} else {
				__antithesis_instrumentation__.Notify(119527)
			}
			__antithesis_instrumentation__.Notify(119525)
			if srKeyRanges[i].End.Compare(totalKeyRanges[i].End) > 0 {
				__antithesis_instrumentation__.Notify(119528)
				totalKeyRanges[i].End = srKeyRanges[i].End
			} else {
				__antithesis_instrumentation__.Notify(119529)
			}
		}
	}
	__antithesis_instrumentation__.Notify(119510)

	for i := range keyRanges {
		__antithesis_instrumentation__.Notify(119530)
		if totalKeyRanges[i].End.Compare(keyRanges[i].End) > 0 {
			__antithesis_instrumentation__.Notify(119532)
			subsumedReplSSTFile := &storage.MemFile{}
			subsumedReplSST := storage.MakeIngestionSSTWriter(
				ctx, r.ClusterSettings(), subsumedReplSSTFile,
			)
			defer subsumedReplSST.Close()
			if err := storage.ClearRangeWithHeuristic(
				r.store.Engine(),
				&subsumedReplSST,
				keyRanges[i].End,
				totalKeyRanges[i].End,
			); err != nil {
				__antithesis_instrumentation__.Notify(119535)
				subsumedReplSST.Close()
				return err
			} else {
				__antithesis_instrumentation__.Notify(119536)
			}
			__antithesis_instrumentation__.Notify(119533)
			if err := subsumedReplSST.Finish(); err != nil {
				__antithesis_instrumentation__.Notify(119537)
				return err
			} else {
				__antithesis_instrumentation__.Notify(119538)
			}
			__antithesis_instrumentation__.Notify(119534)
			if subsumedReplSST.DataSize > 0 {
				__antithesis_instrumentation__.Notify(119539)

				if err := scratch.WriteSST(ctx, subsumedReplSSTFile.Data()); err != nil {
					__antithesis_instrumentation__.Notify(119540)
					return err
				} else {
					__antithesis_instrumentation__.Notify(119541)
				}
			} else {
				__antithesis_instrumentation__.Notify(119542)
			}
		} else {
			__antithesis_instrumentation__.Notify(119543)
		}
		__antithesis_instrumentation__.Notify(119531)

		if totalKeyRanges[i].Start.Compare(keyRanges[i].Start) < 0 {
			__antithesis_instrumentation__.Notify(119544)
			log.Fatalf(ctx, "subsuming replica to our left; key range: %v; total key range %v",
				keyRanges[i], totalKeyRanges[i])
		} else {
			__antithesis_instrumentation__.Notify(119545)
		}
	}
	__antithesis_instrumentation__.Notify(119511)
	return nil
}

func (r *Replica) clearSubsumedReplicaInMemoryData(
	ctx context.Context, subsumedRepls []*Replica, subsumedNextReplicaID roachpb.ReplicaID,
) ([]*ReplicaPlaceholder, error) {
	__antithesis_instrumentation__.Notify(119546)

	var phs []*ReplicaPlaceholder
	for _, sr := range subsumedRepls {
		__antithesis_instrumentation__.Notify(119548)

		ph, err := r.store.removeInitializedReplicaRaftMuLocked(ctx, sr, subsumedNextReplicaID, RemoveOptions{

			DestroyData:       false,
			InsertPlaceholder: true,
		})
		if err != nil {
			__antithesis_instrumentation__.Notify(119550)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(119551)
		}
		__antithesis_instrumentation__.Notify(119549)
		phs = append(phs, ph)

		if err := sr.postDestroyRaftMuLocked(ctx, sr.GetMVCCStats()); err != nil {
			__antithesis_instrumentation__.Notify(119552)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(119553)
		}
	}
	__antithesis_instrumentation__.Notify(119547)
	return phs, nil
}
