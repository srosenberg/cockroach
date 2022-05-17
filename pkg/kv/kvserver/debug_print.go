package kvserver

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/kvserverpb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

func PrintEngineKeyValue(k storage.EngineKey, v []byte) {
	__antithesis_instrumentation__.Notify(101075)
	fmt.Println(SprintEngineKeyValue(k, v))
}

func PrintMVCCKeyValue(kv storage.MVCCKeyValue) {
	__antithesis_instrumentation__.Notify(101076)
	fmt.Println(SprintMVCCKeyValue(kv, true))
}

func SprintEngineKey(key storage.EngineKey) string {
	__antithesis_instrumentation__.Notify(101077)
	if key.IsMVCCKey() {
		__antithesis_instrumentation__.Notify(101079)
		if mvccKey, err := key.ToMVCCKey(); err == nil {
			__antithesis_instrumentation__.Notify(101080)
			return SprintMVCCKey(mvccKey)
		} else {
			__antithesis_instrumentation__.Notify(101081)
		}
	} else {
		__antithesis_instrumentation__.Notify(101082)
	}
	__antithesis_instrumentation__.Notify(101078)

	return fmt.Sprintf("%s %x (%#x): ", key.Key, key.Version, key.Encode())
}

func SprintMVCCKey(key storage.MVCCKey) string {
	__antithesis_instrumentation__.Notify(101083)
	return fmt.Sprintf("%s %s (%#x): ", key.Timestamp, key.Key, storage.EncodeMVCCKey(key))
}

func SprintEngineKeyValue(k storage.EngineKey, v []byte) string {
	__antithesis_instrumentation__.Notify(101084)
	if k.IsMVCCKey() {
		__antithesis_instrumentation__.Notify(101087)
		if key, err := k.ToMVCCKey(); err == nil {
			__antithesis_instrumentation__.Notify(101088)
			return SprintMVCCKeyValue(storage.MVCCKeyValue{Key: key, Value: v}, true)
		} else {
			__antithesis_instrumentation__.Notify(101089)
		}
	} else {
		__antithesis_instrumentation__.Notify(101090)
	}
	__antithesis_instrumentation__.Notify(101085)
	var sb strings.Builder
	fmt.Fprintf(&sb, "%s %x (%#x): ", k.Key, k.Version, k.Encode())
	if out, err := tryIntent(storage.MVCCKeyValue{Value: v}); err == nil {
		__antithesis_instrumentation__.Notify(101091)
		sb.WriteString(out)
	} else {
		__antithesis_instrumentation__.Notify(101092)
		fmt.Fprintf(&sb, "%x", v)
	}
	__antithesis_instrumentation__.Notify(101086)
	return sb.String()
}

var DebugSprintMVCCKeyValueDecoders []func(kv storage.MVCCKeyValue) (string, error)

func SprintMVCCKeyValue(kv storage.MVCCKeyValue, printKey bool) string {
	__antithesis_instrumentation__.Notify(101093)
	var sb strings.Builder
	if printKey {
		__antithesis_instrumentation__.Notify(101097)
		sb.WriteString(SprintMVCCKey(kv.Key))
	} else {
		__antithesis_instrumentation__.Notify(101098)
	}
	__antithesis_instrumentation__.Notify(101094)

	decoders := append(DebugSprintMVCCKeyValueDecoders,
		tryRaftLogEntry,
		tryRangeDescriptor,
		tryMeta,
		tryTxn,
		tryRangeIDKey,
		tryTimeSeries,
		tryIntent,
		func(kv storage.MVCCKeyValue) (string, error) {
			__antithesis_instrumentation__.Notify(101099)

			return fmt.Sprintf("%q", kv.Value), nil
		},
	)
	__antithesis_instrumentation__.Notify(101095)

	for _, decoder := range decoders {
		__antithesis_instrumentation__.Notify(101100)
		out, err := decoder(kv)
		if err != nil {
			__antithesis_instrumentation__.Notify(101102)
			continue
		} else {
			__antithesis_instrumentation__.Notify(101103)
		}
		__antithesis_instrumentation__.Notify(101101)
		sb.WriteString(out)
		return sb.String()
	}
	__antithesis_instrumentation__.Notify(101096)
	panic("unreachable")
}

func SprintIntent(value []byte) string {
	__antithesis_instrumentation__.Notify(101104)
	if out, err := tryIntent(storage.MVCCKeyValue{Value: value}); err == nil {
		__antithesis_instrumentation__.Notify(101106)
		return out
	} else {
		__antithesis_instrumentation__.Notify(101107)
	}
	__antithesis_instrumentation__.Notify(101105)
	return fmt.Sprintf("%x", value)
}

func tryRangeDescriptor(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101108)
	if err := IsRangeDescriptorKey(kv.Key); err != nil {
		__antithesis_instrumentation__.Notify(101111)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101112)
	}
	__antithesis_instrumentation__.Notify(101109)
	var desc roachpb.RangeDescriptor
	if err := getProtoValue(kv.Value, &desc); err != nil {
		__antithesis_instrumentation__.Notify(101113)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101114)
	}
	__antithesis_instrumentation__.Notify(101110)
	return descStr(desc), nil
}

func tryIntent(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101115)
	if len(kv.Value) == 0 {
		__antithesis_instrumentation__.Notify(101119)
		return "", errors.New("empty")
	} else {
		__antithesis_instrumentation__.Notify(101120)
	}
	__antithesis_instrumentation__.Notify(101116)
	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(kv.Value, &meta); err != nil {
		__antithesis_instrumentation__.Notify(101121)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101122)
	}
	__antithesis_instrumentation__.Notify(101117)
	s := fmt.Sprintf("%+v", &meta)
	if meta.Txn != nil {
		__antithesis_instrumentation__.Notify(101123)
		s = meta.Txn.WriteTimestamp.String() + " " + s
	} else {
		__antithesis_instrumentation__.Notify(101124)
	}
	__antithesis_instrumentation__.Notify(101118)
	return s, nil
}

func decodeWriteBatch(writeBatch *kvserverpb.WriteBatch) (string, error) {
	__antithesis_instrumentation__.Notify(101125)
	if writeBatch == nil {
		__antithesis_instrumentation__.Notify(101129)
		return "<nil>\n", nil
	} else {
		__antithesis_instrumentation__.Notify(101130)
	}
	__antithesis_instrumentation__.Notify(101126)

	r, err := storage.NewRocksDBBatchReader(writeBatch.Data)
	if err != nil {
		__antithesis_instrumentation__.Notify(101131)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101132)
	}
	__antithesis_instrumentation__.Notify(101127)

	var sb strings.Builder
	for r.Next() {
		__antithesis_instrumentation__.Notify(101133)
		switch r.BatchType() {
		case storage.BatchTypeDeletion:
			__antithesis_instrumentation__.Notify(101134)
			engineKey, err := r.EngineKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101146)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101147)
			}
			__antithesis_instrumentation__.Notify(101135)
			sb.WriteString(fmt.Sprintf("Delete: %s\n", SprintEngineKey(engineKey)))
		case storage.BatchTypeValue:
			__antithesis_instrumentation__.Notify(101136)
			engineKey, err := r.EngineKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101148)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101149)
			}
			__antithesis_instrumentation__.Notify(101137)
			sb.WriteString(fmt.Sprintf("Put: %s\n", SprintEngineKeyValue(engineKey, r.Value())))
		case storage.BatchTypeMerge:
			__antithesis_instrumentation__.Notify(101138)
			engineKey, err := r.EngineKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101150)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101151)
			}
			__antithesis_instrumentation__.Notify(101139)
			sb.WriteString(fmt.Sprintf("Merge: %s\n", SprintEngineKeyValue(engineKey, r.Value())))
		case storage.BatchTypeSingleDeletion:
			__antithesis_instrumentation__.Notify(101140)
			engineKey, err := r.EngineKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101152)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101153)
			}
			__antithesis_instrumentation__.Notify(101141)
			sb.WriteString(fmt.Sprintf("Single Delete: %s\n", SprintEngineKey(engineKey)))
		case storage.BatchTypeRangeDeletion:
			__antithesis_instrumentation__.Notify(101142)
			engineStartKey, err := r.EngineKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101154)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101155)
			}
			__antithesis_instrumentation__.Notify(101143)
			engineEndKey, err := r.EngineEndKey()
			if err != nil {
				__antithesis_instrumentation__.Notify(101156)
				return sb.String(), err
			} else {
				__antithesis_instrumentation__.Notify(101157)
			}
			__antithesis_instrumentation__.Notify(101144)
			sb.WriteString(fmt.Sprintf(
				"Delete Range: [%s, %s)\n", SprintEngineKey(engineStartKey), SprintEngineKey(engineEndKey),
			))
		default:
			__antithesis_instrumentation__.Notify(101145)
			sb.WriteString(fmt.Sprintf("unsupported batch type: %d\n", r.BatchType()))
		}
	}
	__antithesis_instrumentation__.Notify(101128)
	return sb.String(), r.Error()
}

func tryRaftLogEntry(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101158)
	var ent raftpb.Entry
	if err := maybeUnmarshalInline(kv.Value, &ent); err != nil {
		__antithesis_instrumentation__.Notify(101163)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101164)
	}
	__antithesis_instrumentation__.Notify(101159)

	var cmd kvserverpb.RaftCommand
	switch ent.Type {
	case raftpb.EntryNormal:
		__antithesis_instrumentation__.Notify(101165)
		if len(ent.Data) == 0 {
			__antithesis_instrumentation__.Notify(101171)
			return fmt.Sprintf("%s: EMPTY\n", &ent), nil
		} else {
			__antithesis_instrumentation__.Notify(101172)
		}
		__antithesis_instrumentation__.Notify(101166)
		_, cmdData := kvserverbase.DecodeRaftCommand(ent.Data)
		if err := protoutil.Unmarshal(cmdData, &cmd); err != nil {
			__antithesis_instrumentation__.Notify(101173)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(101174)
		}
	case raftpb.EntryConfChange, raftpb.EntryConfChangeV2:
		__antithesis_instrumentation__.Notify(101167)
		var c raftpb.ConfChangeI
		if ent.Type == raftpb.EntryConfChange {
			__antithesis_instrumentation__.Notify(101175)
			var cc raftpb.ConfChange
			if err := protoutil.Unmarshal(ent.Data, &cc); err != nil {
				__antithesis_instrumentation__.Notify(101177)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(101178)
			}
			__antithesis_instrumentation__.Notify(101176)
			c = cc
		} else {
			__antithesis_instrumentation__.Notify(101179)
			var cc raftpb.ConfChangeV2
			if err := protoutil.Unmarshal(ent.Data, &cc); err != nil {
				__antithesis_instrumentation__.Notify(101181)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(101182)
			}
			__antithesis_instrumentation__.Notify(101180)
			c = cc
		}
		__antithesis_instrumentation__.Notify(101168)

		var ctx kvserverpb.ConfChangeContext
		if err := protoutil.Unmarshal(c.AsV2().Context, &ctx); err != nil {
			__antithesis_instrumentation__.Notify(101183)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(101184)
		}
		__antithesis_instrumentation__.Notify(101169)
		if err := protoutil.Unmarshal(ctx.Payload, &cmd); err != nil {
			__antithesis_instrumentation__.Notify(101185)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(101186)
		}
	default:
		__antithesis_instrumentation__.Notify(101170)
		return "", fmt.Errorf("unknown log entry type: %s", &ent)
	}
	__antithesis_instrumentation__.Notify(101160)
	ent.Data = nil

	var leaseStr string
	if l := cmd.DeprecatedProposerLease; l != nil {
		__antithesis_instrumentation__.Notify(101187)
		leaseStr = l.String()
	} else {
		__antithesis_instrumentation__.Notify(101188)
		leaseStr = fmt.Sprintf("lease #%d", cmd.ProposerLeaseSequence)
	}
	__antithesis_instrumentation__.Notify(101161)

	wbStr, err := decodeWriteBatch(cmd.WriteBatch)
	if err != nil {
		__antithesis_instrumentation__.Notify(101189)
		wbStr = "failed to decode: " + err.Error() + "\nafter:\n" + wbStr
	} else {
		__antithesis_instrumentation__.Notify(101190)
	}
	__antithesis_instrumentation__.Notify(101162)
	cmd.WriteBatch = nil

	return fmt.Sprintf("%s by %s\n%s\nwrite batch:\n%s", &ent, leaseStr, &cmd, wbStr), nil
}

func tryTxn(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101191)
	var txn roachpb.Transaction
	if err := maybeUnmarshalInline(kv.Value, &txn); err != nil {
		__antithesis_instrumentation__.Notify(101193)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101194)
	}
	__antithesis_instrumentation__.Notify(101192)
	return txn.String() + "\n", nil
}

func tryRangeIDKey(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101195)
	if !kv.Key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(101201)
		return "", fmt.Errorf("range ID keys shouldn't have timestamps: %s", kv.Key)
	} else {
		__antithesis_instrumentation__.Notify(101202)
	}
	__antithesis_instrumentation__.Notify(101196)
	_, _, suffix, _, err := keys.DecodeRangeIDKey(kv.Key.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(101203)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101204)
	}
	__antithesis_instrumentation__.Notify(101197)

	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(kv.Value, &meta); err != nil {
		__antithesis_instrumentation__.Notify(101205)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101206)
	}
	__antithesis_instrumentation__.Notify(101198)
	value := roachpb.Value{RawBytes: meta.RawBytes}

	var msg protoutil.Message
	switch {
	case bytes.Equal(suffix, keys.LocalAbortSpanSuffix):
		__antithesis_instrumentation__.Notify(101207)
		msg = &roachpb.AbortSpanEntry{}

	case bytes.Equal(suffix, keys.LocalRangeGCThresholdSuffix):
		__antithesis_instrumentation__.Notify(101208)
		msg = &hlc.Timestamp{}

	case bytes.Equal(suffix, keys.LocalRangeVersionSuffix):
		__antithesis_instrumentation__.Notify(101209)
		msg = &roachpb.Version{}

	case bytes.Equal(suffix, keys.LocalRangeTombstoneSuffix):
		__antithesis_instrumentation__.Notify(101210)
		msg = &roachpb.RangeTombstone{}

	case bytes.Equal(suffix, keys.LocalRaftTruncatedStateSuffix):
		__antithesis_instrumentation__.Notify(101211)
		msg = &roachpb.RaftTruncatedState{}

	case bytes.Equal(suffix, keys.LocalRangeLeaseSuffix):
		__antithesis_instrumentation__.Notify(101212)
		msg = &roachpb.Lease{}

	case bytes.Equal(suffix, keys.LocalRangeAppliedStateSuffix):
		__antithesis_instrumentation__.Notify(101213)
		msg = &enginepb.RangeAppliedState{}

	case bytes.Equal(suffix, keys.LocalRangeStatsLegacySuffix):
		__antithesis_instrumentation__.Notify(101214)
		msg = &enginepb.MVCCStats{}

	case bytes.Equal(suffix, keys.LocalRaftHardStateSuffix):
		__antithesis_instrumentation__.Notify(101215)
		msg = &raftpb.HardState{}

	case bytes.Equal(suffix, keys.LocalRangeLastReplicaGCTimestampSuffix):
		__antithesis_instrumentation__.Notify(101216)
		msg = &hlc.Timestamp{}

	default:
		__antithesis_instrumentation__.Notify(101217)
		return "", fmt.Errorf("unknown raft id key %s", suffix)
	}
	__antithesis_instrumentation__.Notify(101199)

	if err := value.GetProto(msg); err != nil {
		__antithesis_instrumentation__.Notify(101218)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101219)
	}
	__antithesis_instrumentation__.Notify(101200)
	return msg.String(), nil
}

func tryMeta(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101220)
	if !bytes.HasPrefix(kv.Key.Key, keys.Meta1Prefix) && func() bool {
		__antithesis_instrumentation__.Notify(101223)
		return !bytes.HasPrefix(kv.Key.Key, keys.Meta2Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(101224)
		return "", errors.New("not a meta key")
	} else {
		__antithesis_instrumentation__.Notify(101225)
	}
	__antithesis_instrumentation__.Notify(101221)
	value := roachpb.Value{
		Timestamp: kv.Key.Timestamp,
		RawBytes:  kv.Value,
	}
	var desc roachpb.RangeDescriptor
	if err := value.GetProto(&desc); err != nil {
		__antithesis_instrumentation__.Notify(101226)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101227)
	}
	__antithesis_instrumentation__.Notify(101222)
	return descStr(desc), nil
}

func tryTimeSeries(kv storage.MVCCKeyValue) (string, error) {
	__antithesis_instrumentation__.Notify(101228)
	if len(kv.Value) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(101232)
		return !bytes.HasPrefix(kv.Key.Key, keys.TimeseriesPrefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(101233)
		return "", errors.New("empty or not TS")
	} else {
		__antithesis_instrumentation__.Notify(101234)
	}
	__antithesis_instrumentation__.Notify(101229)
	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(kv.Value, &meta); err != nil {
		__antithesis_instrumentation__.Notify(101235)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101236)
	}
	__antithesis_instrumentation__.Notify(101230)
	v := roachpb.Value{RawBytes: meta.RawBytes}
	var ts roachpb.InternalTimeSeriesData
	if err := v.GetProto(&ts); err != nil {
		__antithesis_instrumentation__.Notify(101237)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(101238)
	}
	__antithesis_instrumentation__.Notify(101231)
	return fmt.Sprintf("%+v [mergeTS=%s]", &ts, meta.MergeTimestamp), nil
}

func IsRangeDescriptorKey(key storage.MVCCKey) error {
	__antithesis_instrumentation__.Notify(101239)
	_, suffix, _, err := keys.DecodeRangeKey(key.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(101242)
		return err
	} else {
		__antithesis_instrumentation__.Notify(101243)
	}
	__antithesis_instrumentation__.Notify(101240)
	if !bytes.Equal(suffix, keys.LocalRangeDescriptorSuffix) {
		__antithesis_instrumentation__.Notify(101244)
		return fmt.Errorf("wrong suffix: %s", suffix)
	} else {
		__antithesis_instrumentation__.Notify(101245)
	}
	__antithesis_instrumentation__.Notify(101241)
	return nil
}

func getProtoValue(data []byte, msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(101246)
	value := roachpb.Value{
		RawBytes: data,
	}
	return value.GetProto(msg)
}

func descStr(desc roachpb.RangeDescriptor) string {
	__antithesis_instrumentation__.Notify(101247)
	return fmt.Sprintf("[%s, %s)\n\tRaw:%s\n",
		desc.StartKey, desc.EndKey, &desc)
}

func maybeUnmarshalInline(v []byte, dest protoutil.Message) error {
	__antithesis_instrumentation__.Notify(101248)
	var meta enginepb.MVCCMetadata
	if err := protoutil.Unmarshal(v, &meta); err != nil {
		__antithesis_instrumentation__.Notify(101250)
		return err
	} else {
		__antithesis_instrumentation__.Notify(101251)
	}
	__antithesis_instrumentation__.Notify(101249)
	value := roachpb.Value{
		RawBytes: meta.RawBytes,
	}
	return value.GetProto(dest)
}

type stringifyWriteBatch kvserverpb.WriteBatch

func (s *stringifyWriteBatch) String() string {
	__antithesis_instrumentation__.Notify(101252)
	wbStr, err := decodeWriteBatch((*kvserverpb.WriteBatch)(s))
	if err == nil {
		__antithesis_instrumentation__.Notify(101254)
		return wbStr
	} else {
		__antithesis_instrumentation__.Notify(101255)
	}
	__antithesis_instrumentation__.Notify(101253)
	return fmt.Sprintf("failed to stringify write batch (%x): %s", s.Data, err)
}
