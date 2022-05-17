package loqrecovery

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/loqrecovery/loqrecoverypb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/stateloader"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func CollectReplicaInfo(
	ctx context.Context, stores []storage.Engine,
) (loqrecoverypb.NodeReplicaInfo, error) {
	__antithesis_instrumentation__.Notify(108806)
	if len(stores) == 0 {
		__antithesis_instrumentation__.Notify(108809)
		return loqrecoverypb.NodeReplicaInfo{}, errors.New("no stores were provided for info collection")
	} else {
		__antithesis_instrumentation__.Notify(108810)
	}
	__antithesis_instrumentation__.Notify(108807)

	var replicas []loqrecoverypb.ReplicaInfo
	for _, reader := range stores {
		__antithesis_instrumentation__.Notify(108811)
		storeIdent, err := kvserver.ReadStoreIdent(ctx, reader)
		if err != nil {
			__antithesis_instrumentation__.Notify(108813)
			return loqrecoverypb.NodeReplicaInfo{}, err
		} else {
			__antithesis_instrumentation__.Notify(108814)
		}
		__antithesis_instrumentation__.Notify(108812)
		if err = kvserver.IterateRangeDescriptorsFromDisk(ctx, reader, func(desc roachpb.RangeDescriptor) error {
			__antithesis_instrumentation__.Notify(108815)
			rsl := stateloader.Make(desc.RangeID)
			rstate, err := rsl.Load(ctx, reader, &desc)
			if err != nil {
				__antithesis_instrumentation__.Notify(108819)
				return err
			} else {
				__antithesis_instrumentation__.Notify(108820)
			}
			__antithesis_instrumentation__.Notify(108816)
			hstate, err := rsl.LoadHardState(ctx, reader)
			if err != nil {
				__antithesis_instrumentation__.Notify(108821)
				return err
			} else {
				__antithesis_instrumentation__.Notify(108822)
			}
			__antithesis_instrumentation__.Notify(108817)

			rangeUpdates, err := GetDescriptorChangesFromRaftLog(desc.RangeID,
				rstate.RaftAppliedIndex+1, math.MaxInt64, reader)
			if err != nil {
				__antithesis_instrumentation__.Notify(108823)
				return err
			} else {
				__antithesis_instrumentation__.Notify(108824)
			}
			__antithesis_instrumentation__.Notify(108818)

			replicaData := loqrecoverypb.ReplicaInfo{
				StoreID:                  storeIdent.StoreID,
				NodeID:                   storeIdent.NodeID,
				Desc:                     desc,
				RaftAppliedIndex:         rstate.RaftAppliedIndex,
				RaftCommittedIndex:       hstate.Commit,
				RaftLogDescriptorChanges: rangeUpdates,
			}
			replicas = append(replicas, replicaData)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(108825)
			return loqrecoverypb.NodeReplicaInfo{}, err
		} else {
			__antithesis_instrumentation__.Notify(108826)
		}
	}
	__antithesis_instrumentation__.Notify(108808)
	return loqrecoverypb.NodeReplicaInfo{Replicas: replicas}, nil
}

func GetDescriptorChangesFromRaftLog(
	rangeID roachpb.RangeID, lo, hi uint64, reader storage.Reader,
) ([]loqrecoverypb.DescriptorChangeInfo, error) {
	__antithesis_instrumentation__.Notify(108827)
	var changes []loqrecoverypb.DescriptorChangeInfo

	key := keys.RaftLogKey(rangeID, lo)
	endKey := keys.RaftLogKey(rangeID, hi)
	iter := reader.NewMVCCIterator(storage.MVCCKeyIterKind, storage.IterOptions{
		UpperBound: endKey,
	})
	defer iter.Close()

	var meta enginepb.MVCCMetadata
	var ent raftpb.Entry

	decodeRaftChange := func(ccI raftpb.ConfChangeI) ([]byte, error) {
		__antithesis_instrumentation__.Notify(108829)
		var ccC kvserverpb.ConfChangeContext
		if err := protoutil.Unmarshal(ccI.AsV2().Context, &ccC); err != nil {
			__antithesis_instrumentation__.Notify(108831)
			return nil, errors.Wrap(err, "while unmarshaling CCContext")
		} else {
			__antithesis_instrumentation__.Notify(108832)
		}
		__antithesis_instrumentation__.Notify(108830)
		return ccC.Payload, nil
	}
	__antithesis_instrumentation__.Notify(108828)

	iter.SeekGE(storage.MakeMVCCMetadataKey(key))
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(108833)
		ok, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(108842)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(108843)
		}
		__antithesis_instrumentation__.Notify(108834)
		if !ok {
			__antithesis_instrumentation__.Notify(108844)
			return changes, nil
		} else {
			__antithesis_instrumentation__.Notify(108845)
		}
		__antithesis_instrumentation__.Notify(108835)
		if err := protoutil.Unmarshal(iter.UnsafeValue(), &meta); err != nil {
			__antithesis_instrumentation__.Notify(108846)
			return nil, errors.Wrap(err, "unable to decode raft log MVCCMetadata")
		} else {
			__antithesis_instrumentation__.Notify(108847)
		}
		__antithesis_instrumentation__.Notify(108836)
		if err := storage.MakeValue(meta).GetProto(&ent); err != nil {
			__antithesis_instrumentation__.Notify(108848)
			return nil, errors.Wrap(err, "unable to unmarshal raft Entry")
		} else {
			__antithesis_instrumentation__.Notify(108849)
		}
		__antithesis_instrumentation__.Notify(108837)
		if len(ent.Data) == 0 {
			__antithesis_instrumentation__.Notify(108850)
			continue
		} else {
			__antithesis_instrumentation__.Notify(108851)
		}
		__antithesis_instrumentation__.Notify(108838)

		var payload []byte
		switch ent.Type {
		case raftpb.EntryConfChange:
			__antithesis_instrumentation__.Notify(108852)
			var cc raftpb.ConfChange
			if err := protoutil.Unmarshal(ent.Data, &cc); err != nil {
				__antithesis_instrumentation__.Notify(108858)
				return nil, errors.Wrap(err, "while unmarshaling ConfChange")
			} else {
				__antithesis_instrumentation__.Notify(108859)
			}
			__antithesis_instrumentation__.Notify(108853)
			payload, err = decodeRaftChange(cc)
			if err != nil {
				__antithesis_instrumentation__.Notify(108860)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(108861)
			}
		case raftpb.EntryConfChangeV2:
			__antithesis_instrumentation__.Notify(108854)
			var cc raftpb.ConfChangeV2
			if err := protoutil.Unmarshal(ent.Data, &cc); err != nil {
				__antithesis_instrumentation__.Notify(108862)
				return nil, errors.Wrap(err, "while unmarshaling ConfChangeV2")
			} else {
				__antithesis_instrumentation__.Notify(108863)
			}
			__antithesis_instrumentation__.Notify(108855)
			payload, err = decodeRaftChange(cc)
			if err != nil {
				__antithesis_instrumentation__.Notify(108864)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(108865)
			}
		case raftpb.EntryNormal:
			__antithesis_instrumentation__.Notify(108856)
			_, payload = kvserverbase.DecodeRaftCommand(ent.Data)
		default:
			__antithesis_instrumentation__.Notify(108857)
			continue
		}
		__antithesis_instrumentation__.Notify(108839)
		if len(payload) == 0 {
			__antithesis_instrumentation__.Notify(108866)
			continue
		} else {
			__antithesis_instrumentation__.Notify(108867)
		}
		__antithesis_instrumentation__.Notify(108840)
		var raftCmd kvserverpb.RaftCommand
		if err := protoutil.Unmarshal(payload, &raftCmd); err != nil {
			__antithesis_instrumentation__.Notify(108868)
			return nil, errors.Wrap(err, "unable to unmarshal raft command")
		} else {
			__antithesis_instrumentation__.Notify(108869)
		}
		__antithesis_instrumentation__.Notify(108841)
		switch {
		case raftCmd.ReplicatedEvalResult.Split != nil:
			__antithesis_instrumentation__.Notify(108870)
			changes = append(changes,
				loqrecoverypb.DescriptorChangeInfo{
					ChangeType: loqrecoverypb.DescriptorChangeType_Split,
					Desc:       &raftCmd.ReplicatedEvalResult.Split.LeftDesc,
					OtherDesc:  &raftCmd.ReplicatedEvalResult.Split.RightDesc,
				})
		case raftCmd.ReplicatedEvalResult.Merge != nil:
			__antithesis_instrumentation__.Notify(108871)
			changes = append(changes,
				loqrecoverypb.DescriptorChangeInfo{
					ChangeType: loqrecoverypb.DescriptorChangeType_Merge,
					Desc:       &raftCmd.ReplicatedEvalResult.Merge.LeftDesc,
					OtherDesc:  &raftCmd.ReplicatedEvalResult.Merge.RightDesc,
				})
		case raftCmd.ReplicatedEvalResult.ChangeReplicas != nil:
			__antithesis_instrumentation__.Notify(108872)
			changes = append(changes, loqrecoverypb.DescriptorChangeInfo{
				ChangeType: loqrecoverypb.DescriptorChangeType_ReplicaChange,
				Desc:       raftCmd.ReplicatedEvalResult.ChangeReplicas.Desc,
			})
		default:
			__antithesis_instrumentation__.Notify(108873)
		}
	}
}
