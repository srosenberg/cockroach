package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

var (
	MVCCKeyMax = MakeMVCCMetadataKey(roachpb.KeyMax)

	NilKey = MVCCKey{}
)

const (
	mvccEncodedTimeSentinelLen  = 1
	mvccEncodedTimeWallLen      = 8
	mvccEncodedTimeLogicalLen   = 4
	mvccEncodedTimeSyntheticLen = 1
	mvccEncodedTimeLengthLen    = 1
)

type MVCCKey struct {
	Key       roachpb.Key
	Timestamp hlc.Timestamp
}

func MakeMVCCMetadataKey(key roachpb.Key) MVCCKey {
	__antithesis_instrumentation__.Notify(641519)
	return MVCCKey{Key: key}
}

func (k MVCCKey) Next() MVCCKey {
	__antithesis_instrumentation__.Notify(641520)
	ts := k.Timestamp.Prev()
	if ts.IsEmpty() {
		__antithesis_instrumentation__.Notify(641522)
		return MVCCKey{
			Key: k.Key.Next(),
		}
	} else {
		__antithesis_instrumentation__.Notify(641523)
	}
	__antithesis_instrumentation__.Notify(641521)
	return MVCCKey{
		Key:       k.Key,
		Timestamp: ts,
	}
}

func (k MVCCKey) Compare(o MVCCKey) int {
	__antithesis_instrumentation__.Notify(641524)
	if c := k.Key.Compare(o.Key); c != 0 {
		__antithesis_instrumentation__.Notify(641526)
		return c
	} else {
		__antithesis_instrumentation__.Notify(641527)
	}
	__antithesis_instrumentation__.Notify(641525)
	if k.Timestamp.IsEmpty() && func() bool {
		__antithesis_instrumentation__.Notify(641528)
		return !o.Timestamp.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(641529)
		return -1
	} else {
		__antithesis_instrumentation__.Notify(641530)
		if !k.Timestamp.IsEmpty() && func() bool {
			__antithesis_instrumentation__.Notify(641531)
			return o.Timestamp.IsEmpty() == true
		}() == true {
			__antithesis_instrumentation__.Notify(641532)
			return 1
		} else {
			__antithesis_instrumentation__.Notify(641533)
			return -k.Timestamp.Compare(o.Timestamp)
		}
	}
}

func (k MVCCKey) Less(l MVCCKey) bool {
	__antithesis_instrumentation__.Notify(641534)
	if c := k.Key.Compare(l.Key); c != 0 {
		__antithesis_instrumentation__.Notify(641537)
		return c < 0
	} else {
		__antithesis_instrumentation__.Notify(641538)
	}
	__antithesis_instrumentation__.Notify(641535)
	if !k.IsValue() {
		__antithesis_instrumentation__.Notify(641539)
		return l.IsValue()
	} else {
		__antithesis_instrumentation__.Notify(641540)
		if !l.IsValue() {
			__antithesis_instrumentation__.Notify(641541)
			return false
		} else {
			__antithesis_instrumentation__.Notify(641542)
		}
	}
	__antithesis_instrumentation__.Notify(641536)
	return l.Timestamp.Less(k.Timestamp)
}

func (k MVCCKey) Equal(l MVCCKey) bool {
	__antithesis_instrumentation__.Notify(641543)
	return k.Key.Compare(l.Key) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(641544)
		return k.Timestamp.EqOrdering(l.Timestamp) == true
	}() == true
}

func (k MVCCKey) IsValue() bool {
	__antithesis_instrumentation__.Notify(641545)
	return !k.Timestamp.IsEmpty()
}

func (k MVCCKey) EncodedSize() int {
	__antithesis_instrumentation__.Notify(641546)
	n := len(k.Key) + 1
	if k.IsValue() {
		__antithesis_instrumentation__.Notify(641548)

		n += int(MVCCVersionTimestampSize)
	} else {
		__antithesis_instrumentation__.Notify(641549)
	}
	__antithesis_instrumentation__.Notify(641547)
	return n
}

func (k MVCCKey) String() string {
	__antithesis_instrumentation__.Notify(641550)
	if !k.IsValue() {
		__antithesis_instrumentation__.Notify(641552)
		return k.Key.String()
	} else {
		__antithesis_instrumentation__.Notify(641553)
	}
	__antithesis_instrumentation__.Notify(641551)
	return fmt.Sprintf("%s/%s", k.Key, k.Timestamp)
}

func (k MVCCKey) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(641554)
	fmt.Fprintf(f, "%s/%s", k.Key, k.Timestamp)
}

func (k MVCCKey) Len() int {
	__antithesis_instrumentation__.Notify(641555)
	return encodedMVCCKeyLength(k)
}

func EncodeMVCCKey(key MVCCKey) []byte {
	__antithesis_instrumentation__.Notify(641556)
	keyLen := encodedMVCCKeyLength(key)
	buf := make([]byte, keyLen)
	encodeMVCCKeyToBuf(buf, key, keyLen)
	return buf
}

func EncodeMVCCKeyToBuf(buf []byte, key MVCCKey) []byte {
	__antithesis_instrumentation__.Notify(641557)
	keyLen := encodedMVCCKeyLength(key)
	if cap(buf) < keyLen {
		__antithesis_instrumentation__.Notify(641559)
		buf = make([]byte, keyLen)
	} else {
		__antithesis_instrumentation__.Notify(641560)
		buf = buf[:keyLen]
	}
	__antithesis_instrumentation__.Notify(641558)
	encodeMVCCKeyToBuf(buf, key, keyLen)
	return buf
}

func EncodeMVCCKeyPrefix(key roachpb.Key) []byte {
	__antithesis_instrumentation__.Notify(641561)
	return EncodeMVCCKey(MVCCKey{Key: key})
}

func encodeMVCCKeyToBuf(buf []byte, key MVCCKey, keyLen int) {
	__antithesis_instrumentation__.Notify(641562)
	copy(buf, key.Key)
	pos := len(key.Key)

	buf[pos] = 0
	pos += mvccEncodedTimeSentinelLen

	tsLen := keyLen - pos - mvccEncodedTimeLengthLen
	if tsLen > 0 {
		__antithesis_instrumentation__.Notify(641563)
		encodeMVCCTimestampToBuf(buf[pos:], key.Timestamp)
		pos += tsLen
		buf[pos] = byte(tsLen + mvccEncodedTimeLengthLen)
	} else {
		__antithesis_instrumentation__.Notify(641564)
	}
}

func encodeMVCCTimestamp(ts hlc.Timestamp) []byte {
	__antithesis_instrumentation__.Notify(641565)
	tsLen := encodedMVCCTimestampLength(ts)
	if tsLen == 0 {
		__antithesis_instrumentation__.Notify(641567)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641568)
	}
	__antithesis_instrumentation__.Notify(641566)
	buf := make([]byte, tsLen)
	encodeMVCCTimestampToBuf(buf, ts)
	return buf
}

func EncodeMVCCTimestampSuffix(ts hlc.Timestamp) []byte {
	__antithesis_instrumentation__.Notify(641569)
	tsLen := encodedMVCCTimestampLength(ts)
	if tsLen == 0 {
		__antithesis_instrumentation__.Notify(641571)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641572)
	}
	__antithesis_instrumentation__.Notify(641570)
	buf := make([]byte, tsLen+mvccEncodedTimeLengthLen)
	encodeMVCCTimestampToBuf(buf, ts)
	buf[tsLen] = byte(tsLen + mvccEncodedTimeLengthLen)
	return buf
}

func encodeMVCCTimestampToBuf(buf []byte, ts hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(641573)
	binary.BigEndian.PutUint64(buf, uint64(ts.WallTime))
	if ts.Logical != 0 || func() bool {
		__antithesis_instrumentation__.Notify(641574)
		return ts.Synthetic == true
	}() == true {
		__antithesis_instrumentation__.Notify(641575)
		binary.BigEndian.PutUint32(buf[mvccEncodedTimeWallLen:], uint32(ts.Logical))
		if ts.Synthetic {
			__antithesis_instrumentation__.Notify(641576)
			buf[mvccEncodedTimeWallLen+mvccEncodedTimeLogicalLen] = 1
		} else {
			__antithesis_instrumentation__.Notify(641577)
		}
	} else {
		__antithesis_instrumentation__.Notify(641578)
	}
}

func encodedMVCCKeyLength(key MVCCKey) int {
	__antithesis_instrumentation__.Notify(641579)

	keyLen := len(key.Key) + mvccEncodedTimeSentinelLen
	if !key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(641581)
		keyLen += mvccEncodedTimeWallLen + mvccEncodedTimeLengthLen
		if key.Timestamp.Logical != 0 || func() bool {
			__antithesis_instrumentation__.Notify(641582)
			return key.Timestamp.Synthetic == true
		}() == true {
			__antithesis_instrumentation__.Notify(641583)
			keyLen += mvccEncodedTimeLogicalLen
			if key.Timestamp.Synthetic {
				__antithesis_instrumentation__.Notify(641584)
				keyLen += mvccEncodedTimeSyntheticLen
			} else {
				__antithesis_instrumentation__.Notify(641585)
			}
		} else {
			__antithesis_instrumentation__.Notify(641586)
		}
	} else {
		__antithesis_instrumentation__.Notify(641587)
	}
	__antithesis_instrumentation__.Notify(641580)
	return keyLen
}

func encodedMVCCKeyPrefixLength(key roachpb.Key) int {
	__antithesis_instrumentation__.Notify(641588)
	return len(key) + mvccEncodedTimeSentinelLen
}

func encodedMVCCTimestampLength(ts hlc.Timestamp) int {
	__antithesis_instrumentation__.Notify(641589)

	tsLen := encodedMVCCKeyLength(MVCCKey{Timestamp: ts}) - mvccEncodedTimeSentinelLen
	if tsLen > 0 {
		__antithesis_instrumentation__.Notify(641591)
		tsLen -= mvccEncodedTimeLengthLen
	} else {
		__antithesis_instrumentation__.Notify(641592)
	}
	__antithesis_instrumentation__.Notify(641590)
	return tsLen
}

func encodedMVCCTimestampSuffixLength(ts hlc.Timestamp) int {
	__antithesis_instrumentation__.Notify(641593)

	return encodedMVCCKeyLength(MVCCKey{Timestamp: ts}) - mvccEncodedTimeSentinelLen
}

func DecodeMVCCKey(encodedKey []byte) (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(641594)
	k, ts, err := enginepb.DecodeKey(encodedKey)
	return MVCCKey{k, ts}, err
}

func decodeMVCCTimestamp(encodedTS []byte) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(641595)

	var ts hlc.Timestamp
	switch len(encodedTS) {
	case 0:
		__antithesis_instrumentation__.Notify(641597)

	case 8:
		__antithesis_instrumentation__.Notify(641598)
		ts.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
	case 12:
		__antithesis_instrumentation__.Notify(641599)
		ts.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
		ts.Logical = int32(binary.BigEndian.Uint32(encodedTS[8:12]))
	case 13:
		__antithesis_instrumentation__.Notify(641600)
		ts.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
		ts.Logical = int32(binary.BigEndian.Uint32(encodedTS[8:12]))
		ts.Synthetic = encodedTS[12] != 0
	default:
		__antithesis_instrumentation__.Notify(641601)
		return hlc.Timestamp{}, errors.Errorf("bad timestamp %x", encodedTS)
	}
	__antithesis_instrumentation__.Notify(641596)
	return ts, nil
}

func decodeMVCCTimestampSuffix(encodedTS []byte) (hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(641602)
	if len(encodedTS) == 0 {
		__antithesis_instrumentation__.Notify(641605)
		return hlc.Timestamp{}, nil
	} else {
		__antithesis_instrumentation__.Notify(641606)
	}
	__antithesis_instrumentation__.Notify(641603)
	encodedLen := len(encodedTS)
	if suffixLen := int(encodedTS[encodedLen-1]); suffixLen != encodedLen {
		__antithesis_instrumentation__.Notify(641607)
		return hlc.Timestamp{}, errors.Errorf(
			"bad timestamp: found length suffix %d, actual length %d", suffixLen, encodedLen)
	} else {
		__antithesis_instrumentation__.Notify(641608)
	}
	__antithesis_instrumentation__.Notify(641604)
	return decodeMVCCTimestamp(encodedTS[:encodedLen-1])
}
