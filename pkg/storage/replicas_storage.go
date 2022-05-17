package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

type ReplicaState int

const (
	UninitializedStateMachine ReplicaState = 0

	InitializedStateMachine

	DeletedReplica
)

type FullReplicaID struct {
	RangeID roachpb.RangeID

	ReplicaID roachpb.ReplicaID
}

type ReplicaInfo struct {
	FullReplicaID

	State ReplicaState
}

type MutationBatch interface {
	Commit(sync bool) error

	Batch() *Batch
}

type RaftMutationBatch struct {
	MutationBatch

	Lo, Hi uint64

	HardState *raftpb.HardState

	MustSync bool
}

type RangeStorage interface {
	FullReplicaID() FullReplicaID

	State() ReplicaState

	CurrentRaftEntriesRange() (lo uint64, hi uint64, err error)

	GetHardState() (raftpb.HardState, error)

	CanTruncateRaftIfStateMachineIsDurable(index uint64)

	DoRaftMutation(rBatch RaftMutationBatch) error

	IngestRangeSnapshot(
		span roachpb.Span, raftAppliedIndex uint64, raftAppliedIndexTerm uint64,
		sstPaths []string, subsumedRangeIDs []roachpb.RangeID) error

	ApplyCommittedUsingIngest(sstPaths []string, highestRaftIndex uint64) error

	ApplyCommittedBatch(smBatch MutationBatch) error

	SyncStateMachine() error
}

type ReplicasStorage interface {
	Init()

	CurrentRanges() []ReplicaInfo

	GetRangeTombstone(rangeID roachpb.RangeID) (nextReplicaID roachpb.ReplicaID, err error)

	GetHandle(rr FullReplicaID) (RangeStorage, error)

	CreateUninitializedRange(rr FullReplicaID) (RangeStorage, error)

	SplitReplica(
		r RangeStorage, rhsRR FullReplicaID, rhsSpan roachpb.Span, smBatch MutationBatch,
	) (RangeStorage, error)

	MergeReplicas(lhsRS RangeStorage, rhsRS RangeStorage, smBatch MutationBatch) error

	DiscardReplica(r RangeStorage, nextReplicaID roachpb.ReplicaID) error
}

func MakeSingleEngineReplicasStorage(storeID roachpb.StoreID, eng Engine) ReplicasStorage {
	__antithesis_instrumentation__.Notify(643650)

	return nil
}
