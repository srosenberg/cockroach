package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/raftentry"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

var errSideloadedFileNotFound = errors.New("sideloaded file not found")

type SideloadStorage interface {
	Dir() string

	Put(_ context.Context, index, term uint64, contents []byte) error

	Get(_ context.Context, index, term uint64) ([]byte, error)

	Purge(_ context.Context, index, term uint64) (int64, error)

	Clear(context.Context) error

	TruncateTo(_ context.Context, index uint64) (freed, retained int64, _ error)

	BytesIfTruncatedFromTo(_ context.Context, from, to uint64) (freed, retained int64, _ error)

	Filename(_ context.Context, index, term uint64) (string, error)
}

func (r *Replica) maybeSideloadEntriesRaftMuLocked(
	ctx context.Context, entriesToAppend []raftpb.Entry,
) (_ []raftpb.Entry, sideloadedEntriesSize int64, _ error) {
	__antithesis_instrumentation__.Notify(120531)
	return maybeSideloadEntriesImpl(ctx, entriesToAppend, r.raftMu.sideloaded)
}

func maybeSideloadEntriesImpl(
	ctx context.Context, entriesToAppend []raftpb.Entry, sideloaded SideloadStorage,
) (_ []raftpb.Entry, sideloadedEntriesSize int64, _ error) {
	__antithesis_instrumentation__.Notify(120532)

	cow := false
	for i := range entriesToAppend {
		__antithesis_instrumentation__.Notify(120534)
		if sniffSideloadedRaftCommand(entriesToAppend[i].Data) {
			__antithesis_instrumentation__.Notify(120535)
			log.Event(ctx, "sideloading command in append")
			if !cow {
				__antithesis_instrumentation__.Notify(120541)

				log.Eventf(ctx, "copying entries slice of length %d", len(entriesToAppend))
				cow = true
				entriesToAppend = append([]raftpb.Entry(nil), entriesToAppend...)
			} else {
				__antithesis_instrumentation__.Notify(120542)
			}
			__antithesis_instrumentation__.Notify(120536)

			ent := &entriesToAppend[i]
			cmdID, data := kvserverbase.DecodeRaftCommand(ent.Data)

			var strippedCmd kvserverpb.RaftCommand
			if err := protoutil.Unmarshal(data, &strippedCmd); err != nil {
				__antithesis_instrumentation__.Notify(120543)
				return nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(120544)
			}
			__antithesis_instrumentation__.Notify(120537)

			if strippedCmd.ReplicatedEvalResult.AddSSTable == nil {
				__antithesis_instrumentation__.Notify(120545)

				log.Warning(ctx, "encountered sideloaded Raft command without inlined payload")
				continue
			} else {
				__antithesis_instrumentation__.Notify(120546)
			}
			__antithesis_instrumentation__.Notify(120538)

			dataToSideload := strippedCmd.ReplicatedEvalResult.AddSSTable.Data
			strippedCmd.ReplicatedEvalResult.AddSSTable.Data = nil

			{
				__antithesis_instrumentation__.Notify(120547)
				data := make([]byte, kvserverbase.RaftCommandPrefixLen+strippedCmd.Size())
				kvserverbase.EncodeRaftCommandPrefix(data[:kvserverbase.RaftCommandPrefixLen], kvserverbase.RaftVersionSideloaded, cmdID)
				_, err := protoutil.MarshalTo(&strippedCmd, data[kvserverbase.RaftCommandPrefixLen:])
				if err != nil {
					__antithesis_instrumentation__.Notify(120549)
					return nil, 0, errors.Wrap(err, "while marshaling stripped sideloaded command")
				} else {
					__antithesis_instrumentation__.Notify(120550)
				}
				__antithesis_instrumentation__.Notify(120548)
				ent.Data = data
			}
			__antithesis_instrumentation__.Notify(120539)

			log.Eventf(ctx, "writing payload at index=%d term=%d", ent.Index, ent.Term)
			if err := sideloaded.Put(ctx, ent.Index, ent.Term, dataToSideload); err != nil {
				__antithesis_instrumentation__.Notify(120551)
				return nil, 0, err
			} else {
				__antithesis_instrumentation__.Notify(120552)
			}
			__antithesis_instrumentation__.Notify(120540)
			sideloadedEntriesSize += int64(len(dataToSideload))
		} else {
			__antithesis_instrumentation__.Notify(120553)
		}
	}
	__antithesis_instrumentation__.Notify(120533)
	return entriesToAppend, sideloadedEntriesSize, nil
}

func sniffSideloadedRaftCommand(data []byte) (sideloaded bool) {
	__antithesis_instrumentation__.Notify(120554)
	return len(data) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(120555)
		return data[0] == byte(kvserverbase.RaftVersionSideloaded) == true
	}() == true
}

func maybeInlineSideloadedRaftCommand(
	ctx context.Context,
	rangeID roachpb.RangeID,
	ent raftpb.Entry,
	sideloaded SideloadStorage,
	entryCache *raftentry.Cache,
) (*raftpb.Entry, error) {
	__antithesis_instrumentation__.Notify(120556)
	if !sniffSideloadedRaftCommand(ent.Data) {
		__antithesis_instrumentation__.Notify(120563)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(120564)
	}
	__antithesis_instrumentation__.Notify(120557)
	log.Event(ctx, "inlining sideloaded SSTable")

	cachedSingleton, _, _, _ := entryCache.Scan(
		nil, rangeID, ent.Index, ent.Index+1, 1<<20,
	)

	if len(cachedSingleton) > 0 {
		__antithesis_instrumentation__.Notify(120565)
		log.Event(ctx, "using cache hit")
		return &cachedSingleton[0], nil
	} else {
		__antithesis_instrumentation__.Notify(120566)
	}
	__antithesis_instrumentation__.Notify(120558)

	entCpy := ent
	ent = entCpy

	log.Event(ctx, "inlined entry not cached")

	cmdID, data := kvserverbase.DecodeRaftCommand(ent.Data)

	var command kvserverpb.RaftCommand
	if err := protoutil.Unmarshal(data, &command); err != nil {
		__antithesis_instrumentation__.Notify(120567)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(120568)
	}
	__antithesis_instrumentation__.Notify(120559)

	if len(command.ReplicatedEvalResult.AddSSTable.Data) > 0 {
		__antithesis_instrumentation__.Notify(120569)

		log.Event(ctx, "entry already inlined")
		return &ent, nil
	} else {
		__antithesis_instrumentation__.Notify(120570)
	}
	__antithesis_instrumentation__.Notify(120560)

	sideloadedData, err := sideloaded.Get(ctx, ent.Index, ent.Term)
	if err != nil {
		__antithesis_instrumentation__.Notify(120571)
		return nil, errors.Wrap(err, "loading sideloaded data")
	} else {
		__antithesis_instrumentation__.Notify(120572)
	}
	__antithesis_instrumentation__.Notify(120561)
	command.ReplicatedEvalResult.AddSSTable.Data = sideloadedData
	{
		__antithesis_instrumentation__.Notify(120573)
		data := make([]byte, kvserverbase.RaftCommandPrefixLen+command.Size())
		kvserverbase.EncodeRaftCommandPrefix(data[:kvserverbase.RaftCommandPrefixLen], kvserverbase.RaftVersionSideloaded, cmdID)
		_, err := protoutil.MarshalTo(&command, data[kvserverbase.RaftCommandPrefixLen:])
		if err != nil {
			__antithesis_instrumentation__.Notify(120575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(120576)
		}
		__antithesis_instrumentation__.Notify(120574)
		ent.Data = data
	}
	__antithesis_instrumentation__.Notify(120562)
	return &ent, nil
}

func assertSideloadedRaftCommandInlined(ctx context.Context, ent *raftpb.Entry) {
	__antithesis_instrumentation__.Notify(120577)
	if !sniffSideloadedRaftCommand(ent.Data) {
		__antithesis_instrumentation__.Notify(120580)
		return
	} else {
		__antithesis_instrumentation__.Notify(120581)
	}
	__antithesis_instrumentation__.Notify(120578)

	var command kvserverpb.RaftCommand
	_, data := kvserverbase.DecodeRaftCommand(ent.Data)
	if err := protoutil.Unmarshal(data, &command); err != nil {
		__antithesis_instrumentation__.Notify(120582)
		log.Fatalf(ctx, "%v", err)
	} else {
		__antithesis_instrumentation__.Notify(120583)
	}
	__antithesis_instrumentation__.Notify(120579)

	if len(command.ReplicatedEvalResult.AddSSTable.Data) == 0 {
		__antithesis_instrumentation__.Notify(120584)

		log.Fatalf(ctx, "found thin sideloaded raft command: %+v", command)
	} else {
		__antithesis_instrumentation__.Notify(120585)
	}
}

func maybePurgeSideloaded(
	ctx context.Context, ss SideloadStorage, firstIndex, lastIndex uint64, term uint64,
) (int64, error) {
	__antithesis_instrumentation__.Notify(120586)
	var totalSize int64
	for i := firstIndex; i <= lastIndex; i++ {
		__antithesis_instrumentation__.Notify(120588)
		size, err := ss.Purge(ctx, i, term)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(120590)
			return !errors.Is(err, errSideloadedFileNotFound) == true
		}() == true {
			__antithesis_instrumentation__.Notify(120591)
			return totalSize, err
		} else {
			__antithesis_instrumentation__.Notify(120592)
		}
		__antithesis_instrumentation__.Notify(120589)
		totalSize += size
	}
	__antithesis_instrumentation__.Notify(120587)
	return totalSize, nil
}
