package enginepb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"

	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func SplitMVCCKey(mvccKey []byte) (key []byte, ts []byte, ok bool) {
	__antithesis_instrumentation__.Notify(633835)
	if len(mvccKey) == 0 {
		__antithesis_instrumentation__.Notify(633839)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(633840)
	}
	__antithesis_instrumentation__.Notify(633836)
	tsLen := int(mvccKey[len(mvccKey)-1])
	keyPartEnd := len(mvccKey) - 1 - tsLen
	if keyPartEnd < 0 {
		__antithesis_instrumentation__.Notify(633841)
		return nil, nil, false
	} else {
		__antithesis_instrumentation__.Notify(633842)
	}
	__antithesis_instrumentation__.Notify(633837)

	key = mvccKey[:keyPartEnd]
	if tsLen > 0 {
		__antithesis_instrumentation__.Notify(633843)
		ts = mvccKey[keyPartEnd+1 : len(mvccKey)-1]
	} else {
		__antithesis_instrumentation__.Notify(633844)
	}
	__antithesis_instrumentation__.Notify(633838)
	return key, ts, true
}

func DecodeKey(encodedKey []byte) ([]byte, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(633845)
	key, encodedTS, ok := SplitMVCCKey(encodedKey)
	if !ok {
		__antithesis_instrumentation__.Notify(633848)
		return nil, hlc.Timestamp{}, errors.Errorf("invalid encoded mvcc key: %x", encodedKey)
	} else {
		__antithesis_instrumentation__.Notify(633849)
	}
	__antithesis_instrumentation__.Notify(633846)

	var timestamp hlc.Timestamp
	switch len(encodedTS) {
	case 0:
		__antithesis_instrumentation__.Notify(633850)

	case 8:
		__antithesis_instrumentation__.Notify(633851)
		timestamp.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
	case 12:
		__antithesis_instrumentation__.Notify(633852)
		timestamp.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
		timestamp.Logical = int32(binary.BigEndian.Uint32(encodedTS[8:12]))
	case 13:
		__antithesis_instrumentation__.Notify(633853)
		timestamp.WallTime = int64(binary.BigEndian.Uint64(encodedTS[0:8]))
		timestamp.Logical = int32(binary.BigEndian.Uint32(encodedTS[8:12]))
		timestamp.Synthetic = encodedTS[12] != 0
	default:
		__antithesis_instrumentation__.Notify(633854)
		return nil, hlc.Timestamp{}, errors.Errorf(
			"invalid encoded mvcc key: %x bad timestamp %x", encodedKey, encodedTS)
	}
	__antithesis_instrumentation__.Notify(633847)
	return key, timestamp, nil
}

const kvLenSize = 8

func ScanDecodeKeyValue(
	repr []byte,
) (key []byte, ts hlc.Timestamp, value []byte, orepr []byte, err error) {
	__antithesis_instrumentation__.Notify(633855)
	if len(repr) < kvLenSize {
		__antithesis_instrumentation__.Notify(633858)
		return key, ts, nil, repr, errors.Errorf("unexpected batch EOF")
	} else {
		__antithesis_instrumentation__.Notify(633859)
	}
	__antithesis_instrumentation__.Notify(633856)
	valSize := binary.LittleEndian.Uint32(repr)
	keyEnd := binary.LittleEndian.Uint32(repr[4:kvLenSize]) + kvLenSize
	if (keyEnd + valSize) > uint32(len(repr)) {
		__antithesis_instrumentation__.Notify(633860)
		return key, ts, nil, nil, errors.Errorf("expected %d bytes, but only %d remaining",
			keyEnd+valSize, len(repr))
	} else {
		__antithesis_instrumentation__.Notify(633861)
	}
	__antithesis_instrumentation__.Notify(633857)
	rawKey := repr[kvLenSize:keyEnd]
	value = repr[keyEnd : keyEnd+valSize]
	repr = repr[keyEnd+valSize:]
	key, ts, err = DecodeKey(rawKey)
	return key, ts, value, repr, err
}

func ScanDecodeKeyValueNoTS(repr []byte) (key []byte, value []byte, orepr []byte, err error) {
	__antithesis_instrumentation__.Notify(633862)
	if len(repr) < kvLenSize {
		__antithesis_instrumentation__.Notify(633866)
		return key, nil, repr, errors.Errorf("unexpected batch EOF")
	} else {
		__antithesis_instrumentation__.Notify(633867)
	}
	__antithesis_instrumentation__.Notify(633863)
	valSize := binary.LittleEndian.Uint32(repr)
	keyEnd := binary.LittleEndian.Uint32(repr[4:kvLenSize]) + kvLenSize
	if len(repr) < int(keyEnd+valSize) {
		__antithesis_instrumentation__.Notify(633868)
		return key, nil, nil, errors.Errorf("expected %d bytes, but only %d remaining",
			keyEnd+valSize, len(repr))
	} else {
		__antithesis_instrumentation__.Notify(633869)
	}
	__antithesis_instrumentation__.Notify(633864)

	ret := repr[keyEnd+valSize:]
	value = repr[keyEnd : keyEnd+valSize]
	var ok bool
	rawKey := repr[kvLenSize:keyEnd]
	key, _, ok = SplitMVCCKey(rawKey)
	if !ok {
		__antithesis_instrumentation__.Notify(633870)
		return nil, nil, nil, errors.Errorf("invalid encoded mvcc key: %x", rawKey)
	} else {
		__antithesis_instrumentation__.Notify(633871)
	}
	__antithesis_instrumentation__.Notify(633865)
	return key, value, ret, err
}
