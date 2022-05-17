package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

func makeKey(keys ...[]byte) []byte {
	__antithesis_instrumentation__.Notify(85219)
	return bytes.Join(keys, nil)
}

func MakeStoreKey(suffix, detail roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85220)
	key := make(roachpb.Key, 0, len(LocalStorePrefix)+len(suffix)+len(detail))
	key = append(key, LocalStorePrefix...)
	key = append(key, suffix...)
	key = append(key, detail...)
	return key
}

func DecodeStoreKey(key roachpb.Key) (suffix, detail roachpb.RKey, err error) {
	__antithesis_instrumentation__.Notify(85221)
	if !bytes.HasPrefix(key, LocalStorePrefix) {
		__antithesis_instrumentation__.Notify(85224)
		return nil, nil, errors.Errorf("key %s does not have %s prefix", key, LocalStorePrefix)
	} else {
		__antithesis_instrumentation__.Notify(85225)
	}
	__antithesis_instrumentation__.Notify(85222)

	key = key[len(LocalStorePrefix):]
	if len(key) < localSuffixLength {
		__antithesis_instrumentation__.Notify(85226)
		return nil, nil, errors.Errorf("malformed key does not contain local store suffix")
	} else {
		__antithesis_instrumentation__.Notify(85227)
	}
	__antithesis_instrumentation__.Notify(85223)
	suffix = roachpb.RKey(key[:localSuffixLength])
	detail = roachpb.RKey(key[localSuffixLength:])
	return suffix, detail, nil
}

func StoreIdentKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85228)
	return MakeStoreKey(localStoreIdentSuffix, nil)
}

func StoreGossipKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85229)
	return MakeStoreKey(localStoreGossipSuffix, nil)
}

func StoreClusterVersionKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85230)
	return MakeStoreKey(localStoreClusterVersionSuffix, nil)
}

func StoreLastUpKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85231)
	return MakeStoreKey(localStoreLastUpSuffix, nil)
}

func StoreHLCUpperBoundKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85232)
	return MakeStoreKey(localStoreHLCUpperBoundSuffix, nil)
}

func StoreNodeTombstoneKey(nodeID roachpb.NodeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85233)
	return MakeStoreKey(localStoreNodeTombstoneSuffix, encoding.EncodeUint32Ascending(nil, uint32(nodeID)))
}

func DecodeNodeTombstoneKey(key roachpb.Key) (roachpb.NodeID, error) {
	__antithesis_instrumentation__.Notify(85234)
	suffix, detail, err := DecodeStoreKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85238)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85239)
	}
	__antithesis_instrumentation__.Notify(85235)
	if !suffix.Equal(localStoreNodeTombstoneSuffix) {
		__antithesis_instrumentation__.Notify(85240)
		return 0, errors.Errorf("key with suffix %q != %q", suffix, localStoreNodeTombstoneSuffix)
	} else {
		__antithesis_instrumentation__.Notify(85241)
	}
	__antithesis_instrumentation__.Notify(85236)
	detail, nodeID, err := encoding.DecodeUint32Ascending(detail)
	if len(detail) != 0 {
		__antithesis_instrumentation__.Notify(85242)
		return 0, errors.Errorf("invalid key has trailing garbage: %q", detail)
	} else {
		__antithesis_instrumentation__.Notify(85243)
	}
	__antithesis_instrumentation__.Notify(85237)
	return roachpb.NodeID(nodeID), err
}

func StoreCachedSettingsKey(settingKey roachpb.Key) roachpb.Key {
	__antithesis_instrumentation__.Notify(85244)
	return MakeStoreKey(localStoreCachedSettingsSuffix, encoding.EncodeBytesAscending(nil, settingKey))
}

func DecodeStoreCachedSettingsKey(key roachpb.Key) (settingKey roachpb.Key, err error) {
	__antithesis_instrumentation__.Notify(85245)
	var suffix, detail roachpb.RKey
	suffix, detail, err = DecodeStoreKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85249)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(85250)
	}
	__antithesis_instrumentation__.Notify(85246)
	if !suffix.Equal(localStoreCachedSettingsSuffix) {
		__antithesis_instrumentation__.Notify(85251)
		return nil, errors.Errorf(
			"key with suffix %q != %q",
			suffix,
			localStoreCachedSettingsSuffix,
		)
	} else {
		__antithesis_instrumentation__.Notify(85252)
	}
	__antithesis_instrumentation__.Notify(85247)
	detail, settingKey, err = encoding.DecodeBytesAscending(detail, nil)
	if len(detail) != 0 {
		__antithesis_instrumentation__.Notify(85253)
		return nil, errors.Errorf("invalid key has trailing garbage: %q", detail)
	} else {
		__antithesis_instrumentation__.Notify(85254)
	}
	__antithesis_instrumentation__.Notify(85248)
	return
}

func StoreUnsafeReplicaRecoveryKey(uuid uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85255)
	key := make(roachpb.Key, 0, len(LocalStoreUnsafeReplicaRecoveryKeyMin)+len(uuid))
	key = append(key, LocalStoreUnsafeReplicaRecoveryKeyMin...)
	key = append(key, uuid.GetBytes()...)
	return key
}

func DecodeStoreUnsafeReplicaRecoveryKey(key roachpb.Key) (uuid.UUID, error) {
	__antithesis_instrumentation__.Notify(85256)
	if !bytes.HasPrefix(key, LocalStoreUnsafeReplicaRecoveryKeyMin) {
		__antithesis_instrumentation__.Notify(85259)
		return uuid.UUID{},
			errors.Errorf("key %q does not have %q prefix", string(key), LocalRangeIDPrefix)
	} else {
		__antithesis_instrumentation__.Notify(85260)
	}
	__antithesis_instrumentation__.Notify(85257)
	remainder := key[len(LocalStoreUnsafeReplicaRecoveryKeyMin):]
	entryID, err := uuid.FromBytes(remainder)
	if err != nil {
		__antithesis_instrumentation__.Notify(85261)
		return entryID, errors.Wrap(err, "failed to get uuid from unsafe replica recovery key")
	} else {
		__antithesis_instrumentation__.Notify(85262)
	}
	__antithesis_instrumentation__.Notify(85258)
	return entryID, nil
}

func NodeLivenessKey(nodeID roachpb.NodeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85263)
	key := make(roachpb.Key, 0, len(NodeLivenessPrefix)+9)
	key = append(key, NodeLivenessPrefix...)
	key = encoding.EncodeUvarintAscending(key, uint64(nodeID))
	return key
}

func NodeStatusKey(nodeID roachpb.NodeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85264)
	key := make(roachpb.Key, 0, len(StatusNodePrefix)+9)
	key = append(key, StatusNodePrefix...)
	key = encoding.EncodeUvarintAscending(key, uint64(nodeID))
	return key
}

func makePrefixWithRangeID(prefix []byte, rangeID roachpb.RangeID, infix roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85265)

	key := make(roachpb.Key, 0, 32)
	key = append(key, prefix...)
	key = encoding.EncodeUvarintAscending(key, uint64(rangeID))
	key = append(key, infix...)
	return key
}

func MakeRangeIDPrefix(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85266)
	return makePrefixWithRangeID(LocalRangeIDPrefix, rangeID, nil)
}

func MakeRangeIDReplicatedPrefix(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85267)
	return makePrefixWithRangeID(LocalRangeIDPrefix, rangeID, LocalRangeIDReplicatedInfix)
}

func makeRangeIDReplicatedKey(rangeID roachpb.RangeID, suffix, detail roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85268)
	if len(suffix) != localSuffixLength {
		__antithesis_instrumentation__.Notify(85270)
		panic(fmt.Sprintf("suffix len(%q) != %d", suffix, localSuffixLength))
	} else {
		__antithesis_instrumentation__.Notify(85271)
	}
	__antithesis_instrumentation__.Notify(85269)

	key := MakeRangeIDReplicatedPrefix(rangeID)
	key = append(key, suffix...)
	key = append(key, detail...)
	return key
}

func DecodeRangeIDKey(
	key roachpb.Key,
) (rangeID roachpb.RangeID, infix, suffix, detail roachpb.Key, err error) {
	__antithesis_instrumentation__.Notify(85272)
	if !bytes.HasPrefix(key, LocalRangeIDPrefix) {
		__antithesis_instrumentation__.Notify(85276)
		return 0, nil, nil, nil, errors.Errorf("key %s does not have %s prefix", key, LocalRangeIDPrefix)
	} else {
		__antithesis_instrumentation__.Notify(85277)
	}
	__antithesis_instrumentation__.Notify(85273)

	b := key[len(LocalRangeIDPrefix):]
	b, rangeInt, err := encoding.DecodeUvarintAscending(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(85278)
		return 0, nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(85279)
	}
	__antithesis_instrumentation__.Notify(85274)
	if len(b) < localSuffixLength+1 {
		__antithesis_instrumentation__.Notify(85280)
		return 0, nil, nil, nil, errors.Errorf("malformed key does not contain range ID infix and suffix")
	} else {
		__antithesis_instrumentation__.Notify(85281)
	}
	__antithesis_instrumentation__.Notify(85275)
	infix = b[:1]
	b = b[1:]
	suffix = b[:localSuffixLength]
	b = b[localSuffixLength:]

	return roachpb.RangeID(rangeInt), infix, suffix, b, nil
}

func AbortSpanKey(rangeID roachpb.RangeID, txnID uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85282)
	return MakeRangeIDPrefixBuf(rangeID).AbortSpanKey(txnID)
}

func DecodeAbortSpanKey(key roachpb.Key, dest []byte) (uuid.UUID, error) {
	__antithesis_instrumentation__.Notify(85283)
	_, _, suffix, detail, err := DecodeRangeIDKey(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85288)
		return uuid.UUID{}, err
	} else {
		__antithesis_instrumentation__.Notify(85289)
	}
	__antithesis_instrumentation__.Notify(85284)
	if !bytes.Equal(suffix, LocalAbortSpanSuffix) {
		__antithesis_instrumentation__.Notify(85290)
		return uuid.UUID{}, errors.Errorf("key %s does not contain the AbortSpan suffix %s",
			key, LocalAbortSpanSuffix)
	} else {
		__antithesis_instrumentation__.Notify(85291)
	}
	__antithesis_instrumentation__.Notify(85285)

	detail, idBytes, err := encoding.DecodeBytesAscending(detail, dest)
	if err != nil {
		__antithesis_instrumentation__.Notify(85292)
		return uuid.UUID{}, err
	} else {
		__antithesis_instrumentation__.Notify(85293)
	}
	__antithesis_instrumentation__.Notify(85286)
	if len(detail) > 0 {
		__antithesis_instrumentation__.Notify(85294)
		return uuid.UUID{}, errors.Errorf("key %q has leftover bytes after decode: %s; indicates corrupt key", key, detail)
	} else {
		__antithesis_instrumentation__.Notify(85295)
	}
	__antithesis_instrumentation__.Notify(85287)
	txnID, err := uuid.FromBytes(idBytes)
	return txnID, err
}

func RangeAppliedStateKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85296)
	return MakeRangeIDPrefixBuf(rangeID).RangeAppliedStateKey()
}

func RangeLeaseKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85297)
	return MakeRangeIDPrefixBuf(rangeID).RangeLeaseKey()
}

func RangePriorReadSummaryKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85298)
	return MakeRangeIDPrefixBuf(rangeID).RangePriorReadSummaryKey()
}

func RangeGCThresholdKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85299)
	return MakeRangeIDPrefixBuf(rangeID).RangeGCThresholdKey()
}

func RangeVersionKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85300)
	return MakeRangeIDPrefixBuf(rangeID).RangeVersionKey()
}

func MakeRangeIDUnreplicatedPrefix(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85301)
	return makePrefixWithRangeID(LocalRangeIDPrefix, rangeID, localRangeIDUnreplicatedInfix)
}

func makeRangeIDUnreplicatedKey(
	rangeID roachpb.RangeID, suffix roachpb.RKey, detail roachpb.RKey,
) roachpb.Key {
	__antithesis_instrumentation__.Notify(85302)
	if len(suffix) != localSuffixLength {
		__antithesis_instrumentation__.Notify(85304)
		panic(fmt.Sprintf("suffix len(%q) != %d", suffix, localSuffixLength))
	} else {
		__antithesis_instrumentation__.Notify(85305)
	}
	__antithesis_instrumentation__.Notify(85303)

	key := MakeRangeIDUnreplicatedPrefix(rangeID)
	key = append(key, suffix...)
	key = append(key, detail...)
	return key
}

func RangeTombstoneKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85306)
	return MakeRangeIDPrefixBuf(rangeID).RangeTombstoneKey()
}

func RaftTruncatedStateKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85307)
	return MakeRangeIDPrefixBuf(rangeID).RaftTruncatedStateKey()
}

func RaftHardStateKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85308)
	return MakeRangeIDPrefixBuf(rangeID).RaftHardStateKey()
}

func RaftLogPrefix(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85309)
	return MakeRangeIDPrefixBuf(rangeID).RaftLogPrefix()
}

func RaftLogKey(rangeID roachpb.RangeID, logIndex uint64) roachpb.Key {
	__antithesis_instrumentation__.Notify(85310)
	return MakeRangeIDPrefixBuf(rangeID).RaftLogKey(logIndex)
}

func RaftReplicaIDKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85311)
	return MakeRangeIDPrefixBuf(rangeID).RaftReplicaIDKey()
}

func RangeLastReplicaGCTimestampKey(rangeID roachpb.RangeID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85312)
	return MakeRangeIDPrefixBuf(rangeID).RangeLastReplicaGCTimestampKey()
}

func MakeRangeKey(key, suffix, detail roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85313)
	if len(suffix) != localSuffixLength {
		__antithesis_instrumentation__.Notify(85315)
		panic(fmt.Sprintf("suffix len(%q) != %d", suffix, localSuffixLength))
	} else {
		__antithesis_instrumentation__.Notify(85316)
	}
	__antithesis_instrumentation__.Notify(85314)
	buf := MakeRangeKeyPrefix(key)
	buf = append(buf, suffix...)
	buf = append(buf, detail...)
	return buf
}

func MakeRangeKeyPrefix(key roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85317)
	buf := make(roachpb.Key, 0, len(LocalRangePrefix)+len(key)+1)
	buf = append(buf, LocalRangePrefix...)
	buf = encoding.EncodeBytesAscending(buf, key)
	return buf
}

func DecodeRangeKey(key roachpb.Key) (startKey, suffix, detail roachpb.Key, err error) {
	__antithesis_instrumentation__.Notify(85318)
	if !bytes.HasPrefix(key, LocalRangePrefix) {
		__antithesis_instrumentation__.Notify(85322)
		return nil, nil, nil, errors.Errorf("key %q does not have %q prefix",
			key, LocalRangePrefix)
	} else {
		__antithesis_instrumentation__.Notify(85323)
	}
	__antithesis_instrumentation__.Notify(85319)

	b := key[len(LocalRangePrefix):]
	b, startKey, err = encoding.DecodeBytesAscending(b, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(85324)
		return nil, nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(85325)
	}
	__antithesis_instrumentation__.Notify(85320)
	if len(b) < localSuffixLength {
		__antithesis_instrumentation__.Notify(85326)
		return nil, nil, nil, errors.Errorf("key %q does not have suffix of length %d",
			key, localSuffixLength)
	} else {
		__antithesis_instrumentation__.Notify(85327)
	}
	__antithesis_instrumentation__.Notify(85321)

	suffix = b[:localSuffixLength]
	detail = b[localSuffixLength:]
	return
}

func RangeDescriptorKey(key roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85328)
	return MakeRangeKey(key, LocalRangeDescriptorSuffix, nil)
}

func TransactionKey(key roachpb.Key, txnID uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85329)
	rk, err := Addr(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85331)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(85332)
	}
	__antithesis_instrumentation__.Notify(85330)
	return MakeRangeKey(rk, LocalTransactionSuffix, roachpb.RKey(txnID.GetBytes()))
}

func QueueLastProcessedKey(key roachpb.RKey, queue string) roachpb.Key {
	__antithesis_instrumentation__.Notify(85333)
	return MakeRangeKey(key, LocalQueueLastProcessedSuffix, roachpb.RKey(queue))
}

func RangeProbeKey(key roachpb.RKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(85334)
	return MakeRangeKey(key, LocalRangeProbeSuffix, nil)
}

func LockTableSingleKey(key roachpb.Key, buf []byte) (roachpb.Key, []byte) {
	__antithesis_instrumentation__.Notify(85335)

	keyLen := len(LocalRangeLockTablePrefix) + len(LockTableSingleKeyInfix) + len(key) + 3
	if cap(buf) < keyLen {
		__antithesis_instrumentation__.Notify(85337)
		buf = make([]byte, 0, keyLen)
	} else {
		__antithesis_instrumentation__.Notify(85338)
		buf = buf[:0]
	}
	__antithesis_instrumentation__.Notify(85336)

	buf = append(buf, LocalRangeLockTablePrefix...)
	buf = append(buf, LockTableSingleKeyInfix...)
	buf = encoding.EncodeBytesAscending(buf, key)
	return buf, buf
}

func DecodeLockTableSingleKey(key roachpb.Key) (lockedKey roachpb.Key, err error) {
	__antithesis_instrumentation__.Notify(85339)
	if !bytes.HasPrefix(key, LocalRangeLockTablePrefix) {
		__antithesis_instrumentation__.Notify(85344)
		return nil, errors.Errorf("key %q does not have %q prefix",
			key, LocalRangeLockTablePrefix)
	} else {
		__antithesis_instrumentation__.Notify(85345)
	}
	__antithesis_instrumentation__.Notify(85340)

	b := key[len(LocalRangeLockTablePrefix):]
	if !bytes.HasPrefix(b, LockTableSingleKeyInfix) {
		__antithesis_instrumentation__.Notify(85346)
		return nil, errors.Errorf("key %q is not for a single-key lock", key)
	} else {
		__antithesis_instrumentation__.Notify(85347)
	}
	__antithesis_instrumentation__.Notify(85341)
	b = b[len(LockTableSingleKeyInfix):]

	b, lockedKey, err = encoding.DecodeBytesAscending(b, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(85348)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(85349)
	}
	__antithesis_instrumentation__.Notify(85342)
	if len(b) != 0 {
		__antithesis_instrumentation__.Notify(85350)
		return nil, errors.Errorf("key %q has left-over bytes %d after decoding",
			key, len(b))
	} else {
		__antithesis_instrumentation__.Notify(85351)
	}
	__antithesis_instrumentation__.Notify(85343)
	return lockedKey, err
}

func IsLocal(k roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(85352)
	return bytes.HasPrefix(k, LocalPrefix)
}

func Addr(k roachpb.Key) (roachpb.RKey, error) {
	__antithesis_instrumentation__.Notify(85353)
	if !IsLocal(k) {
		__antithesis_instrumentation__.Notify(85356)
		return roachpb.RKey(k), nil
	} else {
		__antithesis_instrumentation__.Notify(85357)
	}
	__antithesis_instrumentation__.Notify(85354)

	for {
		__antithesis_instrumentation__.Notify(85358)
		if bytes.HasPrefix(k, LocalStorePrefix) {
			__antithesis_instrumentation__.Notify(85363)
			return nil, errors.Errorf("store-local key %q is not addressable", k)
		} else {
			__antithesis_instrumentation__.Notify(85364)
		}
		__antithesis_instrumentation__.Notify(85359)
		if bytes.HasPrefix(k, LocalRangeIDPrefix) {
			__antithesis_instrumentation__.Notify(85365)
			return nil, errors.Errorf("local range ID key %q is not addressable", k)
		} else {
			__antithesis_instrumentation__.Notify(85366)
		}
		__antithesis_instrumentation__.Notify(85360)
		if !bytes.HasPrefix(k, LocalRangePrefix) {
			__antithesis_instrumentation__.Notify(85367)
			return nil, errors.Errorf("local key %q malformed; should contain prefix %q",
				k, LocalRangePrefix)
		} else {
			__antithesis_instrumentation__.Notify(85368)
		}
		__antithesis_instrumentation__.Notify(85361)
		k = k[len(LocalRangePrefix):]
		var err error

		if _, k, err = encoding.DecodeBytesAscending(k, nil); err != nil {
			__antithesis_instrumentation__.Notify(85369)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(85370)
		}
		__antithesis_instrumentation__.Notify(85362)
		if !IsLocal(k) {
			__antithesis_instrumentation__.Notify(85371)
			break
		} else {
			__antithesis_instrumentation__.Notify(85372)
		}
	}
	__antithesis_instrumentation__.Notify(85355)
	return roachpb.RKey(k), nil
}

func MustAddr(k roachpb.Key) roachpb.RKey {
	__antithesis_instrumentation__.Notify(85373)
	rk, err := Addr(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(85375)
		panic(errors.Wrapf(err, "could not take address of '%s'", k))
	} else {
		__antithesis_instrumentation__.Notify(85376)
	}
	__antithesis_instrumentation__.Notify(85374)
	return rk
}

func AddrUpperBound(k roachpb.Key) (roachpb.RKey, error) {
	__antithesis_instrumentation__.Notify(85377)
	rk, err := Addr(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(85380)
		return rk, err
	} else {
		__antithesis_instrumentation__.Notify(85381)
	}
	__antithesis_instrumentation__.Notify(85378)

	if IsLocal(k) && func() bool {
		__antithesis_instrumentation__.Notify(85382)
		return !k.Equal(MakeRangeKeyPrefix(rk)) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85383)

		rk = rk.Next()
	} else {
		__antithesis_instrumentation__.Notify(85384)
	}
	__antithesis_instrumentation__.Notify(85379)
	return rk, nil
}

func SpanAddr(span roachpb.Span) (roachpb.RSpan, error) {
	__antithesis_instrumentation__.Notify(85385)
	rk, err := Addr(span.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85388)
		return roachpb.RSpan{}, err
	} else {
		__antithesis_instrumentation__.Notify(85389)
	}
	__antithesis_instrumentation__.Notify(85386)
	var rek roachpb.RKey
	if len(span.EndKey) > 0 {
		__antithesis_instrumentation__.Notify(85390)
		rek, err = Addr(span.EndKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(85391)
			return roachpb.RSpan{}, err
		} else {
			__antithesis_instrumentation__.Notify(85392)
		}
	} else {
		__antithesis_instrumentation__.Notify(85393)
	}
	__antithesis_instrumentation__.Notify(85387)
	return roachpb.RSpan{Key: rk, EndKey: rek}, nil
}

func RangeMetaKey(key roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(85394)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(85397)
		return roachpb.RKeyMin
	} else {
		__antithesis_instrumentation__.Notify(85398)
	}
	__antithesis_instrumentation__.Notify(85395)
	var prefix roachpb.Key
	switch key[0] {
	case meta1PrefixByte:
		__antithesis_instrumentation__.Notify(85399)
		return roachpb.RKeyMin
	case meta2PrefixByte:
		__antithesis_instrumentation__.Notify(85400)
		prefix = Meta1Prefix
		key = key[len(Meta2Prefix):]
	default:
		__antithesis_instrumentation__.Notify(85401)
		prefix = Meta2Prefix
	}
	__antithesis_instrumentation__.Notify(85396)

	buf := make(roachpb.RKey, 0, len(prefix)+len(key))
	buf = append(buf, prefix...)
	buf = append(buf, key...)
	return buf
}

func UserKey(key roachpb.RKey) roachpb.RKey {
	__antithesis_instrumentation__.Notify(85402)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(85405)
		return roachpb.RKey(Meta1Prefix)
	} else {
		__antithesis_instrumentation__.Notify(85406)
	}
	__antithesis_instrumentation__.Notify(85403)
	var prefix roachpb.Key
	switch key[0] {
	case meta1PrefixByte:
		__antithesis_instrumentation__.Notify(85407)
		prefix = Meta2Prefix
		key = key[len(Meta1Prefix):]
	case meta2PrefixByte:
		__antithesis_instrumentation__.Notify(85408)
		key = key[len(Meta2Prefix):]
	default:
		__antithesis_instrumentation__.Notify(85409)
	}
	__antithesis_instrumentation__.Notify(85404)

	buf := make(roachpb.RKey, 0, len(prefix)+len(key))
	buf = append(buf, prefix...)
	buf = append(buf, key...)
	return buf
}

func InMeta1(k roachpb.RKey) bool {
	__antithesis_instrumentation__.Notify(85410)
	return k.Equal(roachpb.RKeyMin) || func() bool {
		__antithesis_instrumentation__.Notify(85411)
		return bytes.HasPrefix(k, MustAddr(Meta1Prefix)) == true
	}() == true
}

func validateRangeMetaKey(key roachpb.RKey) error {
	__antithesis_instrumentation__.Notify(85412)

	if key.Equal(roachpb.RKeyMin) {
		__antithesis_instrumentation__.Notify(85417)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(85418)
	}
	__antithesis_instrumentation__.Notify(85413)

	if len(key) < len(Meta1Prefix) {
		__antithesis_instrumentation__.Notify(85419)
		return NewInvalidRangeMetaKeyError("too short", key)
	} else {
		__antithesis_instrumentation__.Notify(85420)
	}
	__antithesis_instrumentation__.Notify(85414)

	prefix, body := key[:len(Meta1Prefix)], key[len(Meta1Prefix):]
	if !prefix.Equal(Meta2Prefix) && func() bool {
		__antithesis_instrumentation__.Notify(85421)
		return !prefix.Equal(Meta1Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85422)
		return NewInvalidRangeMetaKeyError("not a meta key", key)
	} else {
		__antithesis_instrumentation__.Notify(85423)
	}
	__antithesis_instrumentation__.Notify(85415)

	if roachpb.RKeyMax.Less(body) {
		__antithesis_instrumentation__.Notify(85424)
		return NewInvalidRangeMetaKeyError("body of meta key range lookup is > KeyMax", key)
	} else {
		__antithesis_instrumentation__.Notify(85425)
	}
	__antithesis_instrumentation__.Notify(85416)
	return nil
}

func MetaScanBounds(key roachpb.RKey) (roachpb.RSpan, error) {
	__antithesis_instrumentation__.Notify(85426)
	if err := validateRangeMetaKey(key); err != nil {
		__antithesis_instrumentation__.Notify(85431)
		return roachpb.RSpan{}, err
	} else {
		__antithesis_instrumentation__.Notify(85432)
	}
	__antithesis_instrumentation__.Notify(85427)

	if key.Equal(Meta2KeyMax) {
		__antithesis_instrumentation__.Notify(85433)
		return roachpb.RSpan{},
			NewInvalidRangeMetaKeyError("Meta2KeyMax can't be used as the key of scan", key)
	} else {
		__antithesis_instrumentation__.Notify(85434)
	}
	__antithesis_instrumentation__.Notify(85428)

	if key.Equal(roachpb.RKeyMin) {
		__antithesis_instrumentation__.Notify(85435)

		return roachpb.RSpan{
			Key:    roachpb.RKey(Meta1Prefix),
			EndKey: roachpb.RKey(Meta1Prefix.PrefixEnd()),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(85436)
	}
	__antithesis_instrumentation__.Notify(85429)
	if key.Equal(Meta1KeyMax) {
		__antithesis_instrumentation__.Notify(85437)

		return roachpb.RSpan{
			Key:    roachpb.RKey(Meta1KeyMax),
			EndKey: roachpb.RKey(Meta1Prefix.PrefixEnd()),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(85438)
	}
	__antithesis_instrumentation__.Notify(85430)

	start := key.Next()
	end := key[:len(Meta1Prefix)].PrefixEnd()
	return roachpb.RSpan{Key: start, EndKey: end}, nil
}

func MetaReverseScanBounds(key roachpb.RKey) (roachpb.RSpan, error) {
	__antithesis_instrumentation__.Notify(85439)
	if err := validateRangeMetaKey(key); err != nil {
		__antithesis_instrumentation__.Notify(85443)
		return roachpb.RSpan{}, err
	} else {
		__antithesis_instrumentation__.Notify(85444)
	}
	__antithesis_instrumentation__.Notify(85440)

	if key.Equal(roachpb.RKeyMin) || func() bool {
		__antithesis_instrumentation__.Notify(85445)
		return key.Equal(Meta1Prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85446)
		return roachpb.RSpan{},
			NewInvalidRangeMetaKeyError("KeyMin and Meta1Prefix can't be used as the key of reverse scan", key)
	} else {
		__antithesis_instrumentation__.Notify(85447)
	}
	__antithesis_instrumentation__.Notify(85441)
	if key.Equal(Meta2Prefix) {
		__antithesis_instrumentation__.Notify(85448)

		return roachpb.RSpan{
			Key:    roachpb.RKey(Meta1Prefix),
			EndKey: roachpb.RKey(key.Next().AsRawKey()),
		}, nil
	} else {
		__antithesis_instrumentation__.Notify(85449)
	}
	__antithesis_instrumentation__.Notify(85442)

	start := key[:len(Meta1Prefix)]
	end := key.Next()
	return roachpb.RSpan{Key: start, EndKey: end}, nil
}

func MakeTableIDIndexID(key []byte, tableID uint32, indexID uint32) []byte {
	__antithesis_instrumentation__.Notify(85450)
	key = encoding.EncodeUvarintAscending(key, uint64(tableID))
	key = encoding.EncodeUvarintAscending(key, uint64(indexID))
	return key
}

func MakeFamilyKey(key []byte, famID uint32) []byte {
	__antithesis_instrumentation__.Notify(85451)
	if famID == 0 {
		__antithesis_instrumentation__.Notify(85453)

		return encoding.EncodeUvarintAscending(key, 0)
	} else {
		__antithesis_instrumentation__.Notify(85454)
	}
	__antithesis_instrumentation__.Notify(85452)
	size := len(key)
	key = encoding.EncodeUvarintAscending(key, uint64(famID))

	return encoding.EncodeUvarintAscending(key, uint64(len(key)-size))
}

func DecodeFamilyKey(key []byte) (uint32, error) {
	__antithesis_instrumentation__.Notify(85455)
	n, err := GetRowPrefixLength(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85460)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85461)
	}
	__antithesis_instrumentation__.Notify(85456)
	if n <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(85462)
		return n >= len(key) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85463)
		return 0, errors.Errorf("invalid row prefix, got prefix length %d for key %s", n, key)
	} else {
		__antithesis_instrumentation__.Notify(85464)
	}
	__antithesis_instrumentation__.Notify(85457)
	_, colFamilyID, err := encoding.DecodeUvarintAscending(key[n:])
	if err != nil {
		__antithesis_instrumentation__.Notify(85465)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85466)
	}
	__antithesis_instrumentation__.Notify(85458)
	if colFamilyID > math.MaxUint32 {
		__antithesis_instrumentation__.Notify(85467)
		return 0, errors.Errorf("column family ID overflow, got %d", colFamilyID)
	} else {
		__antithesis_instrumentation__.Notify(85468)
	}
	__antithesis_instrumentation__.Notify(85459)
	return uint32(colFamilyID), nil
}

func DecodeTableIDIndexID(key []byte) ([]byte, uint32, uint32, error) {
	__antithesis_instrumentation__.Notify(85469)
	var tableID uint64
	var indexID uint64
	var err error

	key, tableID, err = encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85472)
		return nil, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(85473)
	}
	__antithesis_instrumentation__.Notify(85470)
	key, indexID, err = encoding.DecodeUvarintAscending(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85474)
		return nil, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(85475)
	}
	__antithesis_instrumentation__.Notify(85471)
	return key, uint32(tableID), uint32(indexID), nil
}

func GetRowPrefixLength(key roachpb.Key) (int, error) {
	__antithesis_instrumentation__.Notify(85476)
	n := len(key)

	sqlKey, _, err := DecodeTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85482)
		return 0, errors.Errorf("%s: not a valid table key", key)
	} else {
		__antithesis_instrumentation__.Notify(85483)
	}
	__antithesis_instrumentation__.Notify(85477)
	sqlN := len(sqlKey)

	if encoding.PeekType(sqlKey) != encoding.Int {
		__antithesis_instrumentation__.Notify(85484)

		return n, nil
	} else {
		__antithesis_instrumentation__.Notify(85485)
	}
	__antithesis_instrumentation__.Notify(85478)

	colFamIDLenByte := sqlKey[sqlN-1:]
	if encoding.PeekType(colFamIDLenByte) != encoding.Int {
		__antithesis_instrumentation__.Notify(85486)

		return 0, errors.Errorf("%s: not a valid table key", key)
	} else {
		__antithesis_instrumentation__.Notify(85487)
	}
	__antithesis_instrumentation__.Notify(85479)

	_, colFamIDLen, err := encoding.DecodeUvarintAscending(colFamIDLenByte)
	if err != nil {
		__antithesis_instrumentation__.Notify(85488)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85489)
	}
	__antithesis_instrumentation__.Notify(85480)

	if colFamIDLen > uint64(sqlN-1) {
		__antithesis_instrumentation__.Notify(85490)

		return 0, errors.Errorf("%s: malformed table key", key)
	} else {
		__antithesis_instrumentation__.Notify(85491)
	}
	__antithesis_instrumentation__.Notify(85481)
	return n - int(colFamIDLen) - 1, nil
}

func EnsureSafeSplitKey(key roachpb.Key) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(85492)

	idx, err := GetRowPrefixLength(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85494)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(85495)
	}
	__antithesis_instrumentation__.Notify(85493)
	return key[:idx], nil
}

func Range(reqs []roachpb.RequestUnion) (roachpb.RSpan, error) {
	__antithesis_instrumentation__.Notify(85496)
	from := roachpb.RKeyMax
	to := roachpb.RKeyMin
	for _, arg := range reqs {
		__antithesis_instrumentation__.Notify(85498)
		req := arg.GetInner()
		h := req.Header()
		if !roachpb.IsRange(req) && func() bool {
			__antithesis_instrumentation__.Notify(85506)
			return len(h.EndKey) != 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(85507)
			return roachpb.RSpan{}, errors.Errorf("end key specified for non-range operation: %s", req)
		} else {
			__antithesis_instrumentation__.Notify(85508)
		}
		__antithesis_instrumentation__.Notify(85499)

		key, err := Addr(h.Key)
		if err != nil {
			__antithesis_instrumentation__.Notify(85509)
			return roachpb.RSpan{}, err
		} else {
			__antithesis_instrumentation__.Notify(85510)
		}
		__antithesis_instrumentation__.Notify(85500)
		if key.Less(from) {
			__antithesis_instrumentation__.Notify(85511)

			from = key
		} else {
			__antithesis_instrumentation__.Notify(85512)
		}
		__antithesis_instrumentation__.Notify(85501)
		if !key.Less(to) {
			__antithesis_instrumentation__.Notify(85513)

			if bytes.Compare(key, roachpb.RKeyMax) > 0 {
				__antithesis_instrumentation__.Notify(85515)
				return roachpb.RSpan{}, errors.Errorf("%s must be less than KeyMax", key)
			} else {
				__antithesis_instrumentation__.Notify(85516)
			}
			__antithesis_instrumentation__.Notify(85514)
			to = key.Next()
		} else {
			__antithesis_instrumentation__.Notify(85517)
		}
		__antithesis_instrumentation__.Notify(85502)

		if len(h.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(85518)
			continue
		} else {
			__antithesis_instrumentation__.Notify(85519)
		}
		__antithesis_instrumentation__.Notify(85503)
		endKey, err := AddrUpperBound(h.EndKey)
		if err != nil {
			__antithesis_instrumentation__.Notify(85520)
			return roachpb.RSpan{}, err
		} else {
			__antithesis_instrumentation__.Notify(85521)
		}
		__antithesis_instrumentation__.Notify(85504)
		if bytes.Compare(roachpb.RKeyMax, endKey) < 0 {
			__antithesis_instrumentation__.Notify(85522)
			return roachpb.RSpan{}, errors.Errorf("%s must be less than or equal to KeyMax", endKey)
		} else {
			__antithesis_instrumentation__.Notify(85523)
		}
		__antithesis_instrumentation__.Notify(85505)
		if to.Less(endKey) {
			__antithesis_instrumentation__.Notify(85524)

			to = endKey
		} else {
			__antithesis_instrumentation__.Notify(85525)
		}
	}
	__antithesis_instrumentation__.Notify(85497)
	return roachpb.RSpan{Key: from, EndKey: to}, nil
}

type RangeIDPrefixBuf roachpb.Key

func MakeRangeIDPrefixBuf(rangeID roachpb.RangeID) RangeIDPrefixBuf {
	__antithesis_instrumentation__.Notify(85526)
	return RangeIDPrefixBuf(MakeRangeIDPrefix(rangeID))
}

func (b RangeIDPrefixBuf) replicatedPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85527)
	return append(roachpb.Key(b), LocalRangeIDReplicatedInfix...)
}

func (b RangeIDPrefixBuf) unreplicatedPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85528)
	return append(roachpb.Key(b), localRangeIDUnreplicatedInfix...)
}

func (b RangeIDPrefixBuf) AbortSpanKey(txnID uuid.UUID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85529)
	key := append(b.replicatedPrefix(), LocalAbortSpanSuffix...)
	return encoding.EncodeBytesAscending(key, txnID.GetBytes())
}

func (b RangeIDPrefixBuf) RangeAppliedStateKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85530)
	return append(b.replicatedPrefix(), LocalRangeAppliedStateSuffix...)
}

func (b RangeIDPrefixBuf) RangeLeaseKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85531)
	return append(b.replicatedPrefix(), LocalRangeLeaseSuffix...)
}

func (b RangeIDPrefixBuf) RangePriorReadSummaryKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85532)
	return append(b.replicatedPrefix(), LocalRangePriorReadSummarySuffix...)
}

func (b RangeIDPrefixBuf) RangeGCThresholdKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85533)
	return append(b.replicatedPrefix(), LocalRangeGCThresholdSuffix...)
}

func (b RangeIDPrefixBuf) RangeVersionKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85534)
	return append(b.replicatedPrefix(), LocalRangeVersionSuffix...)
}

func (b RangeIDPrefixBuf) RangeTombstoneKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85535)
	return append(b.unreplicatedPrefix(), LocalRangeTombstoneSuffix...)
}

func (b RangeIDPrefixBuf) RaftTruncatedStateKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85536)
	return append(b.unreplicatedPrefix(), LocalRaftTruncatedStateSuffix...)
}

func (b RangeIDPrefixBuf) RaftHardStateKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85537)
	return append(b.unreplicatedPrefix(), LocalRaftHardStateSuffix...)
}

func (b RangeIDPrefixBuf) RaftLogPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85538)
	return append(b.unreplicatedPrefix(), LocalRaftLogSuffix...)
}

func (b RangeIDPrefixBuf) RaftLogKey(logIndex uint64) roachpb.Key {
	__antithesis_instrumentation__.Notify(85539)
	return encoding.EncodeUint64Ascending(b.RaftLogPrefix(), logIndex)
}

func (b RangeIDPrefixBuf) RaftReplicaIDKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85540)
	return append(b.unreplicatedPrefix(), LocalRaftReplicaIDSuffix...)
}

func (b RangeIDPrefixBuf) RangeLastReplicaGCTimestampKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85541)
	return append(b.unreplicatedPrefix(), LocalRangeLastReplicaGCTimestampSuffix...)
}
