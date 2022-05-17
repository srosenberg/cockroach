package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

type EngineKey struct {
	Key     roachpb.Key
	Version []byte
}

const (
	engineKeyNoVersion                             = 0
	engineKeyVersionWallTimeLen                    = 8
	engineKeyVersionWallAndLogicalTimeLen          = 12
	engineKeyVersionWallLogicalAndSyntheticTimeLen = 13
	engineKeyVersionLockTableLen                   = 17
)

func (k EngineKey) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(633758)
	fmt.Fprintf(f, "%s/%x", k.Key, k.Version)
}

const (
	sentinel               = '\x00'
	sentinelLen            = 1
	suffixEncodedLengthLen = 1
)

func (k EngineKey) Copy() EngineKey {
	__antithesis_instrumentation__.Notify(633759)
	buf := make([]byte, len(k.Key)+len(k.Version))
	copy(buf, k.Key)
	k.Key = buf[:len(k.Key)]
	if len(k.Version) > 0 {
		__antithesis_instrumentation__.Notify(633761)
		versionCopy := buf[len(k.Key):]
		copy(versionCopy, k.Version)
		k.Version = versionCopy
	} else {
		__antithesis_instrumentation__.Notify(633762)
	}
	__antithesis_instrumentation__.Notify(633760)
	return k
}

func (k EngineKey) EncodedLen() int {
	__antithesis_instrumentation__.Notify(633763)
	n := len(k.Key) + suffixEncodedLengthLen
	versionLen := len(k.Version)
	if versionLen > 0 {
		__antithesis_instrumentation__.Notify(633765)
		n += sentinelLen + versionLen
	} else {
		__antithesis_instrumentation__.Notify(633766)
	}
	__antithesis_instrumentation__.Notify(633764)
	return n
}

func (k EngineKey) Encode() []byte {
	__antithesis_instrumentation__.Notify(633767)
	encodedLen := k.EncodedLen()
	buf := make([]byte, encodedLen)
	k.encodeToSizedBuf(buf)
	return buf
}

func (k EngineKey) EncodeToBuf(buf []byte) []byte {
	__antithesis_instrumentation__.Notify(633768)
	encodedLen := k.EncodedLen()
	if cap(buf) < encodedLen {
		__antithesis_instrumentation__.Notify(633770)
		buf = make([]byte, encodedLen)
	} else {
		__antithesis_instrumentation__.Notify(633771)
		buf = buf[:encodedLen]
	}
	__antithesis_instrumentation__.Notify(633769)
	k.encodeToSizedBuf(buf)
	return buf
}

func (k EngineKey) encodeToSizedBuf(buf []byte) {
	__antithesis_instrumentation__.Notify(633772)
	copy(buf, k.Key)
	pos := len(k.Key)

	suffixLen := len(buf) - pos - 1
	if suffixLen > 0 {
		__antithesis_instrumentation__.Notify(633774)
		buf[pos] = 0
		pos += sentinelLen
		copy(buf[pos:], k.Version)
	} else {
		__antithesis_instrumentation__.Notify(633775)
	}
	__antithesis_instrumentation__.Notify(633773)
	buf[len(buf)-1] = byte(suffixLen)
}

func (k EngineKey) IsMVCCKey() bool {
	__antithesis_instrumentation__.Notify(633776)
	l := len(k.Version)
	return l == engineKeyNoVersion || func() bool {
		__antithesis_instrumentation__.Notify(633777)
		return l == engineKeyVersionWallTimeLen == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(633778)
		return l == engineKeyVersionWallAndLogicalTimeLen == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(633779)
		return l == engineKeyVersionWallLogicalAndSyntheticTimeLen == true
	}() == true
}

func (k EngineKey) IsLockTableKey() bool {
	__antithesis_instrumentation__.Notify(633780)
	return len(k.Version) == engineKeyVersionLockTableLen
}

func (k EngineKey) ToMVCCKey() (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(633781)
	key := MVCCKey{Key: k.Key}
	switch len(k.Version) {
	case engineKeyNoVersion:
		__antithesis_instrumentation__.Notify(633783)

	case engineKeyVersionWallTimeLen:
		__antithesis_instrumentation__.Notify(633784)
		key.Timestamp.WallTime = int64(binary.BigEndian.Uint64(k.Version[0:8]))
	case engineKeyVersionWallAndLogicalTimeLen:
		__antithesis_instrumentation__.Notify(633785)
		key.Timestamp.WallTime = int64(binary.BigEndian.Uint64(k.Version[0:8]))
		key.Timestamp.Logical = int32(binary.BigEndian.Uint32(k.Version[8:12]))
	case engineKeyVersionWallLogicalAndSyntheticTimeLen:
		__antithesis_instrumentation__.Notify(633786)
		key.Timestamp.WallTime = int64(binary.BigEndian.Uint64(k.Version[0:8]))
		key.Timestamp.Logical = int32(binary.BigEndian.Uint32(k.Version[8:12]))
		key.Timestamp.Synthetic = k.Version[12] != 0
	default:
		__antithesis_instrumentation__.Notify(633787)
		return MVCCKey{}, errors.Errorf("version is not an encoded timestamp %x", k.Version)
	}
	__antithesis_instrumentation__.Notify(633782)
	return key, nil
}

func (k EngineKey) ToLockTableKey() (LockTableKey, error) {
	__antithesis_instrumentation__.Notify(633788)
	lockedKey, err := keys.DecodeLockTableSingleKey(k.Key)
	if err != nil {
		__antithesis_instrumentation__.Notify(633791)
		return LockTableKey{}, err
	} else {
		__antithesis_instrumentation__.Notify(633792)
	}
	__antithesis_instrumentation__.Notify(633789)
	key := LockTableKey{Key: lockedKey}
	switch len(k.Version) {
	case engineKeyVersionLockTableLen:
		__antithesis_instrumentation__.Notify(633793)
		key.Strength = lock.Strength(k.Version[0])
		if key.Strength < lock.None || func() bool {
			__antithesis_instrumentation__.Notify(633796)
			return key.Strength > lock.Exclusive == true
		}() == true {
			__antithesis_instrumentation__.Notify(633797)
			return LockTableKey{}, errors.Errorf("unknown strength %d", key.Strength)
		} else {
			__antithesis_instrumentation__.Notify(633798)
		}
		__antithesis_instrumentation__.Notify(633794)
		key.TxnUUID = k.Version[1:]
	default:
		__antithesis_instrumentation__.Notify(633795)
		return LockTableKey{}, errors.Errorf("version is not valid for a LockTableKey %x", k.Version)
	}
	__antithesis_instrumentation__.Notify(633790)
	return key, nil
}

func (k EngineKey) Validate() error {
	__antithesis_instrumentation__.Notify(633799)
	_, errMVCC := k.ToMVCCKey()
	_, errLock := k.ToLockTableKey()
	if errMVCC != nil && func() bool {
		__antithesis_instrumentation__.Notify(633801)
		return errLock != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(633802)
		return errors.Newf("key %s is neither an MVCCKey or LockTableKey", k)
	} else {
		__antithesis_instrumentation__.Notify(633803)
	}
	__antithesis_instrumentation__.Notify(633800)
	return nil
}

func DecodeEngineKey(b []byte) (key EngineKey, ok bool) {
	__antithesis_instrumentation__.Notify(633804)
	if len(b) == 0 {
		__antithesis_instrumentation__.Notify(633808)
		return EngineKey{}, false
	} else {
		__antithesis_instrumentation__.Notify(633809)
	}
	__antithesis_instrumentation__.Notify(633805)

	versionLen := int(b[len(b)-1])

	keyPartEnd := len(b) - 1 - versionLen
	if keyPartEnd < 0 {
		__antithesis_instrumentation__.Notify(633810)
		return EngineKey{}, false
	} else {
		__antithesis_instrumentation__.Notify(633811)
	}
	__antithesis_instrumentation__.Notify(633806)

	key.Key = b[:keyPartEnd]
	if versionLen > 0 {
		__antithesis_instrumentation__.Notify(633812)

		key.Version = b[keyPartEnd+1 : len(b)-1]
	} else {
		__antithesis_instrumentation__.Notify(633813)
	}
	__antithesis_instrumentation__.Notify(633807)
	return key, true
}

func GetKeyPartFromEngineKey(engineKey []byte) (key []byte, ok bool) {
	__antithesis_instrumentation__.Notify(633814)
	if len(engineKey) == 0 {
		__antithesis_instrumentation__.Notify(633817)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(633818)
	}
	__antithesis_instrumentation__.Notify(633815)

	versionLen := int(engineKey[len(engineKey)-1])

	keyPartEnd := len(engineKey) - 1 - versionLen
	if keyPartEnd < 0 {
		__antithesis_instrumentation__.Notify(633819)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(633820)
	}
	__antithesis_instrumentation__.Notify(633816)

	return engineKey[:keyPartEnd], true
}

type EngineKeyFormatter struct {
	key EngineKey
}

var _ fmt.Formatter = EngineKeyFormatter{}

func (m EngineKeyFormatter) Format(f fmt.State, c rune) {
	__antithesis_instrumentation__.Notify(633821)
	m.key.Format(f, c)
}

type LockTableKey struct {
	Key      roachpb.Key
	Strength lock.Strength

	TxnUUID []byte
}

func (lk LockTableKey) ToEngineKey(buf []byte) (EngineKey, []byte) {
	__antithesis_instrumentation__.Notify(633822)
	if len(lk.TxnUUID) != uuid.Size {
		__antithesis_instrumentation__.Notify(633827)
		panic("invalid TxnUUID")
	} else {
		__antithesis_instrumentation__.Notify(633828)
	}
	__antithesis_instrumentation__.Notify(633823)
	if lk.Strength != lock.Exclusive {
		__antithesis_instrumentation__.Notify(633829)
		panic("unsupported lock strength")
	} else {
		__antithesis_instrumentation__.Notify(633830)
	}
	__antithesis_instrumentation__.Notify(633824)

	estimatedLen :=
		(len(keys.LocalRangeLockTablePrefix) + len(keys.LockTableSingleKeyInfix) + len(lk.Key) + 3) +
			engineKeyVersionLockTableLen
	if cap(buf) < estimatedLen {
		__antithesis_instrumentation__.Notify(633831)
		buf = make([]byte, 0, estimatedLen)
	} else {
		__antithesis_instrumentation__.Notify(633832)
	}
	__antithesis_instrumentation__.Notify(633825)
	ltKey, buf := keys.LockTableSingleKey(lk.Key, buf)
	k := EngineKey{Key: ltKey}
	if cap(buf)-len(buf) >= engineKeyVersionLockTableLen {
		__antithesis_instrumentation__.Notify(633833)
		k.Version = buf[len(buf) : len(buf)+engineKeyVersionLockTableLen]
	} else {
		__antithesis_instrumentation__.Notify(633834)

		k.Version = make([]byte, engineKeyVersionLockTableLen)
	}
	__antithesis_instrumentation__.Notify(633826)
	k.Version[0] = byte(lk.Strength)
	copy(k.Version[1:], lk.TxnUUID)
	return k, buf
}
