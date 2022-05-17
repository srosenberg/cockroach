package stateloader

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type StateLoader struct {
	keys.RangeIDPrefixBuf
}

func Make(rangeID roachpb.RangeID) StateLoader {
	__antithesis_instrumentation__.Notify(123611)
	rsl := StateLoader{
		RangeIDPrefixBuf: keys.MakeRangeIDPrefixBuf(rangeID),
	}
	return rsl
}

func (rsl StateLoader) Load(
	ctx context.Context, reader storage.Reader, desc *roachpb.RangeDescriptor,
) (kvserverpb.ReplicaState, error) {
	__antithesis_instrumentation__.Notify(123612)
	var s kvserverpb.ReplicaState

	s.Desc = protoutil.Clone(desc).(*roachpb.RangeDescriptor)

	lease, err := rsl.LoadLease(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(123620)
		return kvserverpb.ReplicaState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123621)
	}
	__antithesis_instrumentation__.Notify(123613)
	s.Lease = &lease

	if s.GCThreshold, err = rsl.LoadGCThreshold(ctx, reader); err != nil {
		__antithesis_instrumentation__.Notify(123622)
		return kvserverpb.ReplicaState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123623)
	}
	__antithesis_instrumentation__.Notify(123614)

	as, err := rsl.LoadRangeAppliedState(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(123624)
		return kvserverpb.ReplicaState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123625)
	}
	__antithesis_instrumentation__.Notify(123615)
	s.RaftAppliedIndex = as.RaftAppliedIndex
	s.RaftAppliedIndexTerm = as.RaftAppliedIndexTerm
	s.LeaseAppliedIndex = as.LeaseAppliedIndex
	ms := as.RangeStats.ToStats()
	s.Stats = &ms
	if as.RaftClosedTimestamp != nil {
		__antithesis_instrumentation__.Notify(123626)
		s.RaftClosedTimestamp = *as.RaftClosedTimestamp
	} else {
		__antithesis_instrumentation__.Notify(123627)
	}
	__antithesis_instrumentation__.Notify(123616)

	truncState, err := rsl.LoadRaftTruncatedState(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(123628)
		return kvserverpb.ReplicaState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123629)
	}
	__antithesis_instrumentation__.Notify(123617)
	s.TruncatedState = &truncState

	version, err := rsl.LoadVersion(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(123630)
		return kvserverpb.ReplicaState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123631)
	}
	__antithesis_instrumentation__.Notify(123618)
	if (version != roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(123632)
		s.Version = &version
	} else {
		__antithesis_instrumentation__.Notify(123633)
	}
	__antithesis_instrumentation__.Notify(123619)

	return s, nil
}

func (rsl StateLoader) Save(
	ctx context.Context, readWriter storage.ReadWriter, state kvserverpb.ReplicaState,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(123634)
	ms := state.Stats
	if err := rsl.SetLease(ctx, readWriter, ms, *state.Lease); err != nil {
		__antithesis_instrumentation__.Notify(123640)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123641)
	}
	__antithesis_instrumentation__.Notify(123635)
	if err := rsl.SetGCThreshold(ctx, readWriter, ms, state.GCThreshold); err != nil {
		__antithesis_instrumentation__.Notify(123642)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123643)
	}
	__antithesis_instrumentation__.Notify(123636)
	if err := rsl.SetRaftTruncatedState(ctx, readWriter, state.TruncatedState); err != nil {
		__antithesis_instrumentation__.Notify(123644)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123645)
	}
	__antithesis_instrumentation__.Notify(123637)
	if state.Version != nil {
		__antithesis_instrumentation__.Notify(123646)
		if err := rsl.SetVersion(ctx, readWriter, ms, state.Version); err != nil {
			__antithesis_instrumentation__.Notify(123647)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(123648)
		}
	} else {
		__antithesis_instrumentation__.Notify(123649)
	}
	__antithesis_instrumentation__.Notify(123638)
	rai, lai, rait, ct := state.RaftAppliedIndex, state.LeaseAppliedIndex, state.RaftAppliedIndexTerm,
		&state.RaftClosedTimestamp
	if err := rsl.SetRangeAppliedState(ctx, readWriter, rai, lai, rait, ms, ct); err != nil {
		__antithesis_instrumentation__.Notify(123650)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123651)
	}
	__antithesis_instrumentation__.Notify(123639)
	return *ms, nil
}

func (rsl StateLoader) LoadLease(
	ctx context.Context, reader storage.Reader,
) (roachpb.Lease, error) {
	__antithesis_instrumentation__.Notify(123652)
	var lease roachpb.Lease
	_, err := storage.MVCCGetProto(ctx, reader, rsl.RangeLeaseKey(),
		hlc.Timestamp{}, &lease, storage.MVCCGetOptions{})
	return lease, err
}

func (rsl StateLoader) SetLease(
	ctx context.Context, readWriter storage.ReadWriter, ms *enginepb.MVCCStats, lease roachpb.Lease,
) error {
	__antithesis_instrumentation__.Notify(123653)
	return storage.MVCCPutProto(ctx, readWriter, ms, rsl.RangeLeaseKey(),
		hlc.Timestamp{}, nil, &lease)
}

func (rsl StateLoader) LoadRangeAppliedState(
	ctx context.Context, reader storage.Reader,
) (enginepb.RangeAppliedState, error) {
	__antithesis_instrumentation__.Notify(123654)
	var as enginepb.RangeAppliedState
	_, err := storage.MVCCGetProto(ctx, reader, rsl.RangeAppliedStateKey(), hlc.Timestamp{}, &as,
		storage.MVCCGetOptions{})
	return as, err
}

func (rsl StateLoader) LoadMVCCStats(
	ctx context.Context, reader storage.Reader,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(123655)

	as, err := rsl.LoadRangeAppliedState(ctx, reader)
	if err != nil {
		__antithesis_instrumentation__.Notify(123657)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123658)
	}
	__antithesis_instrumentation__.Notify(123656)
	return as.RangeStats.ToStats(), nil
}

func (rsl StateLoader) SetRangeAppliedState(
	ctx context.Context,
	readWriter storage.ReadWriter,
	appliedIndex, leaseAppliedIndex, appliedIndexTerm uint64,
	newMS *enginepb.MVCCStats,
	raftClosedTimestamp *hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(123659)
	as := enginepb.RangeAppliedState{
		RaftAppliedIndex:     appliedIndex,
		LeaseAppliedIndex:    leaseAppliedIndex,
		RangeStats:           newMS.ToPersistentStats(),
		RaftAppliedIndexTerm: appliedIndexTerm,
	}
	if raftClosedTimestamp != nil && func() bool {
		__antithesis_instrumentation__.Notify(123661)
		return !raftClosedTimestamp.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(123662)
		as.RaftClosedTimestamp = raftClosedTimestamp
	} else {
		__antithesis_instrumentation__.Notify(123663)
	}
	__antithesis_instrumentation__.Notify(123660)

	ms := (*enginepb.MVCCStats)(nil)
	return storage.MVCCPutProto(ctx, readWriter, ms, rsl.RangeAppliedStateKey(), hlc.Timestamp{}, nil, &as)
}

func (rsl StateLoader) SetMVCCStats(
	ctx context.Context, readWriter storage.ReadWriter, newMS *enginepb.MVCCStats,
) error {
	__antithesis_instrumentation__.Notify(123664)
	as, err := rsl.LoadRangeAppliedState(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(123666)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123667)
	}
	__antithesis_instrumentation__.Notify(123665)
	return rsl.SetRangeAppliedState(
		ctx, readWriter, as.RaftAppliedIndex, as.LeaseAppliedIndex, as.RaftAppliedIndexTerm, newMS,
		as.RaftClosedTimestamp)
}

func (rsl StateLoader) SetClosedTimestamp(
	ctx context.Context, readWriter storage.ReadWriter, closedTS *hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(123668)
	as, err := rsl.LoadRangeAppliedState(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(123670)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123671)
	}
	__antithesis_instrumentation__.Notify(123669)
	return rsl.SetRangeAppliedState(
		ctx, readWriter, as.RaftAppliedIndex, as.LeaseAppliedIndex, as.RaftAppliedIndexTerm,
		as.RangeStats.ToStatsPtr(), closedTS)
}

func (rsl StateLoader) LoadGCThreshold(
	ctx context.Context, reader storage.Reader,
) (*hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(123672)
	var t hlc.Timestamp
	_, err := storage.MVCCGetProto(ctx, reader, rsl.RangeGCThresholdKey(),
		hlc.Timestamp{}, &t, storage.MVCCGetOptions{})
	return &t, err
}

func (rsl StateLoader) SetGCThreshold(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	threshold *hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(123673)
	if threshold == nil {
		__antithesis_instrumentation__.Notify(123675)
		return errors.New("cannot persist nil GCThreshold")
	} else {
		__antithesis_instrumentation__.Notify(123676)
	}
	__antithesis_instrumentation__.Notify(123674)
	return storage.MVCCPutProto(ctx, readWriter, ms,
		rsl.RangeGCThresholdKey(), hlc.Timestamp{}, nil, threshold)
}

func (rsl StateLoader) LoadVersion(
	ctx context.Context, reader storage.Reader,
) (roachpb.Version, error) {
	__antithesis_instrumentation__.Notify(123677)
	var version roachpb.Version
	_, err := storage.MVCCGetProto(ctx, reader, rsl.RangeVersionKey(),
		hlc.Timestamp{}, &version, storage.MVCCGetOptions{})
	return version, err
}

func (rsl StateLoader) SetVersion(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms *enginepb.MVCCStats,
	version *roachpb.Version,
) error {
	__antithesis_instrumentation__.Notify(123678)
	return storage.MVCCPutProto(ctx, readWriter, ms,
		rsl.RangeVersionKey(), hlc.Timestamp{}, nil, version)
}

func (rsl StateLoader) LoadLastIndex(ctx context.Context, reader storage.Reader) (uint64, error) {
	__antithesis_instrumentation__.Notify(123679)
	prefix := rsl.RaftLogPrefix()

	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{LowerBound: prefix})
	defer iter.Close()

	var lastIndex uint64
	iter.SeekLT(storage.MakeMVCCMetadataKey(rsl.RaftLogKey(math.MaxUint64)))
	if ok, _ := iter.Valid(); ok {
		__antithesis_instrumentation__.Notify(123682)
		key := iter.Key()
		var err error
		_, lastIndex, err = encoding.DecodeUint64Ascending(key.Key[len(prefix):])
		if err != nil {
			__antithesis_instrumentation__.Notify(123683)
			log.Fatalf(ctx, "unable to decode Raft log index key: %s", key)
		} else {
			__antithesis_instrumentation__.Notify(123684)
		}
	} else {
		__antithesis_instrumentation__.Notify(123685)
	}
	__antithesis_instrumentation__.Notify(123680)

	if lastIndex == 0 {
		__antithesis_instrumentation__.Notify(123686)

		lastEnt, err := rsl.LoadRaftTruncatedState(ctx, reader)
		if err != nil {
			__antithesis_instrumentation__.Notify(123688)
			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(123689)
		}
		__antithesis_instrumentation__.Notify(123687)
		lastIndex = lastEnt.Index
	} else {
		__antithesis_instrumentation__.Notify(123690)
	}
	__antithesis_instrumentation__.Notify(123681)
	return lastIndex, nil
}

func (rsl StateLoader) LoadRaftTruncatedState(
	ctx context.Context, reader storage.Reader,
) (roachpb.RaftTruncatedState, error) {
	__antithesis_instrumentation__.Notify(123691)
	var truncState roachpb.RaftTruncatedState
	if _, err := storage.MVCCGetProto(
		ctx, reader, rsl.RaftTruncatedStateKey(), hlc.Timestamp{}, &truncState, storage.MVCCGetOptions{},
	); err != nil {
		__antithesis_instrumentation__.Notify(123693)
		return roachpb.RaftTruncatedState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123694)
	}
	__antithesis_instrumentation__.Notify(123692)
	return truncState, nil
}

func (rsl StateLoader) SetRaftTruncatedState(
	ctx context.Context, writer storage.Writer, truncState *roachpb.RaftTruncatedState,
) error {
	__antithesis_instrumentation__.Notify(123695)
	if (*truncState == roachpb.RaftTruncatedState{}) {
		__antithesis_instrumentation__.Notify(123697)
		return errors.New("cannot persist empty RaftTruncatedState")
	} else {
		__antithesis_instrumentation__.Notify(123698)
	}
	__antithesis_instrumentation__.Notify(123696)

	return storage.MVCCBlindPutProto(
		ctx,
		writer,
		nil,
		rsl.RaftTruncatedStateKey(),
		hlc.Timestamp{},
		truncState,
		nil,
	)
}

func (rsl StateLoader) LoadHardState(
	ctx context.Context, reader storage.Reader,
) (raftpb.HardState, error) {
	__antithesis_instrumentation__.Notify(123699)
	var hs raftpb.HardState
	found, err := storage.MVCCGetProto(ctx, reader, rsl.RaftHardStateKey(),
		hlc.Timestamp{}, &hs, storage.MVCCGetOptions{})

	if !found || func() bool {
		__antithesis_instrumentation__.Notify(123701)
		return err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(123702)
		return raftpb.HardState{}, err
	} else {
		__antithesis_instrumentation__.Notify(123703)
	}
	__antithesis_instrumentation__.Notify(123700)
	return hs, nil
}

func (rsl StateLoader) SetHardState(
	ctx context.Context, writer storage.Writer, hs raftpb.HardState,
) error {
	__antithesis_instrumentation__.Notify(123704)

	return storage.MVCCBlindPutProto(
		ctx,
		writer,
		nil,
		rsl.RaftHardStateKey(),
		hlc.Timestamp{},
		&hs,
		nil,
	)
}

func (rsl StateLoader) SynthesizeRaftState(
	ctx context.Context, readWriter storage.ReadWriter,
) error {
	__antithesis_instrumentation__.Notify(123705)
	hs, err := rsl.LoadHardState(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(123709)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123710)
	}
	__antithesis_instrumentation__.Notify(123706)
	truncState, err := rsl.LoadRaftTruncatedState(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(123711)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123712)
	}
	__antithesis_instrumentation__.Notify(123707)
	as, err := rsl.LoadRangeAppliedState(ctx, readWriter)
	if err != nil {
		__antithesis_instrumentation__.Notify(123713)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123714)
	}
	__antithesis_instrumentation__.Notify(123708)
	return rsl.SynthesizeHardState(ctx, readWriter, hs, truncState, as.RaftAppliedIndex)
}

func (rsl StateLoader) SynthesizeHardState(
	ctx context.Context,
	readWriter storage.ReadWriter,
	oldHS raftpb.HardState,
	truncState roachpb.RaftTruncatedState,
	raftAppliedIndex uint64,
) error {
	__antithesis_instrumentation__.Notify(123715)
	newHS := raftpb.HardState{
		Term: truncState.Term,

		Commit: raftAppliedIndex,
	}

	if oldHS.Commit > newHS.Commit {
		__antithesis_instrumentation__.Notify(123719)
		return errors.Newf("can't decrease HardState.Commit from %d to %d",
			redact.Safe(oldHS.Commit), redact.Safe(newHS.Commit))
	} else {
		__antithesis_instrumentation__.Notify(123720)
	}
	__antithesis_instrumentation__.Notify(123716)
	if oldHS.Term > newHS.Term {
		__antithesis_instrumentation__.Notify(123721)

		newHS.Term = oldHS.Term
	} else {
		__antithesis_instrumentation__.Notify(123722)
	}
	__antithesis_instrumentation__.Notify(123717)

	if oldHS.Term == newHS.Term {
		__antithesis_instrumentation__.Notify(123723)
		newHS.Vote = oldHS.Vote
	} else {
		__antithesis_instrumentation__.Notify(123724)
	}
	__antithesis_instrumentation__.Notify(123718)
	err := rsl.SetHardState(ctx, readWriter, newHS)
	return errors.Wrapf(err, "writing HardState %+v", &newHS)
}

func (rsl StateLoader) SetRaftReplicaID(
	ctx context.Context, writer storage.Writer, replicaID roachpb.ReplicaID,
) error {
	__antithesis_instrumentation__.Notify(123725)
	rid := roachpb.RaftReplicaID{ReplicaID: replicaID}

	return storage.MVCCBlindPutProto(
		ctx,
		writer,
		nil,
		rsl.RaftReplicaIDKey(),
		hlc.Timestamp{},
		&rid,
		nil,
	)
}

func (rsl StateLoader) LoadRaftReplicaID(
	ctx context.Context, reader storage.Reader,
) (replicaID roachpb.RaftReplicaID, found bool, err error) {
	__antithesis_instrumentation__.Notify(123726)
	found, err = storage.MVCCGetProto(ctx, reader, rsl.RaftReplicaIDKey(),
		hlc.Timestamp{}, &replicaID, storage.MVCCGetOptions{})
	return
}
