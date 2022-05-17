package stateloader

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

const (
	raftInitialLogIndex = 10
	raftInitialLogTerm  = 5

	RaftLogTermSignalForAddRaftAppliedIndexTermMigration = 3
)

func WriteInitialReplicaState(
	ctx context.Context,
	readWriter storage.ReadWriter,
	ms enginepb.MVCCStats,
	desc roachpb.RangeDescriptor,
	lease roachpb.Lease,
	gcThreshold hlc.Timestamp,
	replicaVersion roachpb.Version,
	writeRaftAppliedIndexTerm bool,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(123576)
	rsl := Make(desc.RangeID)
	var s kvserverpb.ReplicaState
	s.TruncatedState = &roachpb.RaftTruncatedState{
		Term:  raftInitialLogTerm,
		Index: raftInitialLogIndex,
	}
	s.RaftAppliedIndex = s.TruncatedState.Index
	if writeRaftAppliedIndexTerm {
		__antithesis_instrumentation__.Notify(123583)
		s.RaftAppliedIndexTerm = s.TruncatedState.Term
	} else {
		__antithesis_instrumentation__.Notify(123584)
	}
	__antithesis_instrumentation__.Notify(123577)
	s.Desc = &roachpb.RangeDescriptor{
		RangeID: desc.RangeID,
	}
	s.Stats = &ms
	s.Lease = &lease
	s.GCThreshold = &gcThreshold
	if (replicaVersion != roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(123585)
		s.Version = &replicaVersion
	} else {
		__antithesis_instrumentation__.Notify(123586)
	}
	__antithesis_instrumentation__.Notify(123578)

	if existingLease, err := rsl.LoadLease(ctx, readWriter); err != nil {
		__antithesis_instrumentation__.Notify(123587)
		return enginepb.MVCCStats{}, errors.Wrap(err, "error reading lease")
	} else {
		__antithesis_instrumentation__.Notify(123588)
		if (existingLease != roachpb.Lease{}) {
			__antithesis_instrumentation__.Notify(123589)
			log.Fatalf(ctx, "expected trivial lease, but found %+v", existingLease)
		} else {
			__antithesis_instrumentation__.Notify(123590)
		}
	}
	__antithesis_instrumentation__.Notify(123579)

	if existingGCThreshold, err := rsl.LoadGCThreshold(ctx, readWriter); err != nil {
		__antithesis_instrumentation__.Notify(123591)
		return enginepb.MVCCStats{}, errors.Wrap(err, "error reading GCThreshold")
	} else {
		__antithesis_instrumentation__.Notify(123592)
		if !existingGCThreshold.IsEmpty() {
			__antithesis_instrumentation__.Notify(123593)
			log.Fatalf(ctx, "expected trivial GCthreshold, but found %+v", existingGCThreshold)
		} else {
			__antithesis_instrumentation__.Notify(123594)
		}
	}
	__antithesis_instrumentation__.Notify(123580)

	if existingVersion, err := rsl.LoadVersion(ctx, readWriter); err != nil {
		__antithesis_instrumentation__.Notify(123595)
		return enginepb.MVCCStats{}, errors.Wrap(err, "error reading Version")
	} else {
		__antithesis_instrumentation__.Notify(123596)
		if (existingVersion != roachpb.Version{}) {
			__antithesis_instrumentation__.Notify(123597)
			log.Fatalf(ctx, "expected trivial version, but found %+v", existingVersion)
		} else {
			__antithesis_instrumentation__.Notify(123598)
		}
	}
	__antithesis_instrumentation__.Notify(123581)

	newMS, err := rsl.Save(ctx, readWriter, s)
	if err != nil {
		__antithesis_instrumentation__.Notify(123599)
		return enginepb.MVCCStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(123600)
	}
	__antithesis_instrumentation__.Notify(123582)

	return newMS, nil
}

func WriteInitialRangeState(
	ctx context.Context,
	readWriter storage.ReadWriter,
	desc roachpb.RangeDescriptor,
	replicaID roachpb.ReplicaID,
	replicaVersion roachpb.Version,
) error {
	__antithesis_instrumentation__.Notify(123601)
	initialLease := roachpb.Lease{}
	initialGCThreshold := hlc.Timestamp{}
	initialMS := enginepb.MVCCStats{}

	writeRaftAppliedIndexTerm :=
		clusterversion.ClusterVersion{Version: replicaVersion}.IsActiveVersion(
			clusterversion.ByKey(clusterversion.AddRaftAppliedIndexTermMigration))
	if _, err := WriteInitialReplicaState(
		ctx, readWriter, initialMS, desc, initialLease, initialGCThreshold,
		replicaVersion, writeRaftAppliedIndexTerm,
	); err != nil {
		__antithesis_instrumentation__.Notify(123605)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123606)
	}
	__antithesis_instrumentation__.Notify(123602)
	sl := Make(desc.RangeID)
	if err := sl.SynthesizeRaftState(ctx, readWriter); err != nil {
		__antithesis_instrumentation__.Notify(123607)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123608)
	}
	__antithesis_instrumentation__.Notify(123603)

	if err := sl.SetRaftReplicaID(ctx, readWriter, replicaID); err != nil {
		__antithesis_instrumentation__.Notify(123609)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123610)
	}
	__antithesis_instrumentation__.Notify(123604)
	return nil
}
