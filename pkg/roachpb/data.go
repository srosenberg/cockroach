package roachpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"hash/crc32"
	"math"
	"math/rand"
	"sort"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/keysbase"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/interval"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/timetz"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
	"go.etcd.io/etcd/raft/v3/raftpb"
)

const (
	localPrefixByte = '\x01'

	LocalMaxByte = '\x02'
)

var (
	RKeyMin = RKey("")

	KeyMin = Key(RKeyMin)

	RKeyMax = RKey(keysbase.KeyMax)

	KeyMax = Key(RKeyMax)

	LocalPrefix = Key{localPrefixByte}

	LocalMax = Key{LocalMaxByte}

	PrettyPrintKey func(valDirs []encoding.Direction, key Key) string

	PrettyPrintRange func(start, end Key, maxChars int) string
)

type RKey Key

func (rk RKey) AsRawKey() Key {
	__antithesis_instrumentation__.Notify(158798)
	return Key(rk)
}

func (rk RKey) Less(otherRK RKey) bool {
	__antithesis_instrumentation__.Notify(158799)
	return rk.Compare(otherRK) < 0
}

func (rk RKey) Compare(other RKey) int {
	__antithesis_instrumentation__.Notify(158800)
	return bytes.Compare(rk, other)
}

func (rk RKey) Equal(other []byte) bool {
	__antithesis_instrumentation__.Notify(158801)
	return bytes.Equal(rk, other)
}

func (rk RKey) Next() RKey {
	__antithesis_instrumentation__.Notify(158802)
	return RKey(BytesNext(rk))
}

func (rk RKey) PrefixEnd() RKey {
	__antithesis_instrumentation__.Notify(158803)
	return RKey(keysbase.PrefixEnd(rk))
}

func (rk RKey) String() string {
	__antithesis_instrumentation__.Notify(158804)
	return Key(rk).String()
}

func (rk RKey) StringWithDirs(valDirs []encoding.Direction, maxLen int) string {
	__antithesis_instrumentation__.Notify(158805)
	return Key(rk).StringWithDirs(valDirs, maxLen)
}

type Key []byte

func BytesNext(b []byte) []byte {
	__antithesis_instrumentation__.Notify(158806)
	if cap(b) > len(b) {
		__antithesis_instrumentation__.Notify(158808)
		bNext := b[:len(b)+1]
		if bNext[len(bNext)-1] == 0 {
			__antithesis_instrumentation__.Notify(158809)
			return bNext
		} else {
			__antithesis_instrumentation__.Notify(158810)
		}
	} else {
		__antithesis_instrumentation__.Notify(158811)
	}
	__antithesis_instrumentation__.Notify(158807)

	bn := make([]byte, len(b)+1)
	copy(bn, b)
	bn[len(bn)-1] = 0
	return bn
}

func (k Key) Clone() Key {
	__antithesis_instrumentation__.Notify(158812)
	if k == nil {
		__antithesis_instrumentation__.Notify(158814)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(158815)
	}
	__antithesis_instrumentation__.Notify(158813)
	c := make(Key, len(k))
	copy(c, k)
	return c
}

func (k Key) Next() Key {
	__antithesis_instrumentation__.Notify(158816)
	return Key(BytesNext(k))
}

func (k Key) IsPrev(m Key) bool {
	__antithesis_instrumentation__.Notify(158817)
	l := len(m) - 1
	return l == len(k) && func() bool {
		__antithesis_instrumentation__.Notify(158818)
		return m[l] == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(158819)
		return k.Equal(m[:l]) == true
	}() == true
}

func (k Key) PrefixEnd() Key {
	__antithesis_instrumentation__.Notify(158820)
	return Key(keysbase.PrefixEnd(k))
}

func (k Key) Equal(l Key) bool {
	__antithesis_instrumentation__.Notify(158821)
	return bytes.Equal(k, l)
}

func (k Key) Compare(b Key) int {
	__antithesis_instrumentation__.Notify(158822)
	return bytes.Compare(k, b)
}

func (k Key) String() string {
	__antithesis_instrumentation__.Notify(158823)
	return k.StringWithDirs(nil, 0)
}

func (k Key) StringWithDirs(valDirs []encoding.Direction, maxLen int) string {
	__antithesis_instrumentation__.Notify(158824)
	var s string
	if PrettyPrintKey != nil {
		__antithesis_instrumentation__.Notify(158827)
		s = PrettyPrintKey(valDirs, k)
	} else {
		__antithesis_instrumentation__.Notify(158828)
		s = fmt.Sprintf("%q", []byte(k))
	}
	__antithesis_instrumentation__.Notify(158825)
	if maxLen != 0 && func() bool {
		__antithesis_instrumentation__.Notify(158829)
		return len(s) > maxLen == true
	}() == true {
		__antithesis_instrumentation__.Notify(158830)
		return s[0:maxLen] + "..."
	} else {
		__antithesis_instrumentation__.Notify(158831)
	}
	__antithesis_instrumentation__.Notify(158826)
	return s
}

func (k Key) Format(f fmt.State, verb rune) {
	__antithesis_instrumentation__.Notify(158832)

	if verb == 'x' {
		__antithesis_instrumentation__.Notify(158833)
		fmt.Fprintf(f, "%x", []byte(k))
	} else {
		__antithesis_instrumentation__.Notify(158834)
		if PrettyPrintKey != nil {
			__antithesis_instrumentation__.Notify(158835)
			fmt.Fprint(f, PrettyPrintKey(nil, k))
		} else {
			__antithesis_instrumentation__.Notify(158836)
			fmt.Fprint(f, strconv.Quote(string(k)))
		}
	}
}

const (
	checksumUninitialized = 0
	checksumSize          = 4
	tagPos                = checksumSize
	headerSize            = tagPos + 1
)

func (v Value) checksum() uint32 {
	__antithesis_instrumentation__.Notify(158837)
	if len(v.RawBytes) < checksumSize {
		__antithesis_instrumentation__.Notify(158840)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(158841)
	}
	__antithesis_instrumentation__.Notify(158838)
	_, u, err := encoding.DecodeUint32Ascending(v.RawBytes[:checksumSize])
	if err != nil {
		__antithesis_instrumentation__.Notify(158842)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(158843)
	}
	__antithesis_instrumentation__.Notify(158839)
	return u
}

func (v *Value) setChecksum(cksum uint32) {
	__antithesis_instrumentation__.Notify(158844)
	if len(v.RawBytes) >= checksumSize {
		__antithesis_instrumentation__.Notify(158845)
		encoding.EncodeUint32Ascending(v.RawBytes[:0], cksum)
	} else {
		__antithesis_instrumentation__.Notify(158846)
	}
}

func (v *Value) InitChecksum(key []byte) {
	__antithesis_instrumentation__.Notify(158847)
	if v.RawBytes == nil {
		__antithesis_instrumentation__.Notify(158850)
		return
	} else {
		__antithesis_instrumentation__.Notify(158851)
	}
	__antithesis_instrumentation__.Notify(158848)

	if v.checksum() != checksumUninitialized {
		__antithesis_instrumentation__.Notify(158852)
		panic(fmt.Sprintf("initialized checksum = %x", v.checksum()))
	} else {
		__antithesis_instrumentation__.Notify(158853)
	}
	__antithesis_instrumentation__.Notify(158849)
	v.setChecksum(v.computeChecksum(key))
}

func (v *Value) ClearChecksum() {
	__antithesis_instrumentation__.Notify(158854)
	v.setChecksum(0)
}

func (v Value) Verify(key []byte) error {
	__antithesis_instrumentation__.Notify(158855)
	if err := v.VerifyHeader(); err != nil {
		__antithesis_instrumentation__.Notify(158858)
		return err
	} else {
		__antithesis_instrumentation__.Notify(158859)
	}
	__antithesis_instrumentation__.Notify(158856)
	if sum := v.checksum(); sum != 0 {
		__antithesis_instrumentation__.Notify(158860)
		if computedSum := v.computeChecksum(key); computedSum != sum {
			__antithesis_instrumentation__.Notify(158861)
			return fmt.Errorf("%s: invalid checksum (%x) value [% x]",
				Key(key), computedSum, v.RawBytes)
		} else {
			__antithesis_instrumentation__.Notify(158862)
		}
	} else {
		__antithesis_instrumentation__.Notify(158863)
	}
	__antithesis_instrumentation__.Notify(158857)
	return nil
}

func (v Value) VerifyHeader() error {
	__antithesis_instrumentation__.Notify(158864)
	if n := len(v.RawBytes); n > 0 && func() bool {
		__antithesis_instrumentation__.Notify(158866)
		return n < headerSize == true
	}() == true {
		__antithesis_instrumentation__.Notify(158867)
		return errors.Errorf("invalid header size: %d", n)
	} else {
		__antithesis_instrumentation__.Notify(158868)
	}
	__antithesis_instrumentation__.Notify(158865)
	return nil
}

func (v *Value) ShallowClone() *Value {
	__antithesis_instrumentation__.Notify(158869)
	if v == nil {
		__antithesis_instrumentation__.Notify(158871)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(158872)
	}
	__antithesis_instrumentation__.Notify(158870)
	t := *v
	return &t
}

func (v *Value) IsPresent() bool {
	__antithesis_instrumentation__.Notify(158873)
	return v != nil && func() bool {
		__antithesis_instrumentation__.Notify(158874)
		return len(v.RawBytes) != 0 == true
	}() == true
}

func MakeValueFromString(s string) Value {
	__antithesis_instrumentation__.Notify(158875)
	v := Value{}
	v.SetString(s)
	return v
}

func MakeValueFromBytes(bs []byte) Value {
	__antithesis_instrumentation__.Notify(158876)
	v := Value{}
	v.SetBytes(bs)
	return v
}

func MakeValueFromBytesAndTimestamp(bs []byte, t hlc.Timestamp) Value {
	__antithesis_instrumentation__.Notify(158877)
	v := Value{Timestamp: t}
	v.SetBytes(bs)
	return v
}

func (v Value) GetTag() ValueType {
	__antithesis_instrumentation__.Notify(158878)
	if len(v.RawBytes) <= tagPos {
		__antithesis_instrumentation__.Notify(158880)
		return ValueType_UNKNOWN
	} else {
		__antithesis_instrumentation__.Notify(158881)
	}
	__antithesis_instrumentation__.Notify(158879)
	return ValueType(v.RawBytes[tagPos])
}

func (v *Value) setTag(t ValueType) {
	__antithesis_instrumentation__.Notify(158882)
	v.RawBytes[tagPos] = byte(t)
}

func (v Value) dataBytes() []byte {
	__antithesis_instrumentation__.Notify(158883)
	return v.RawBytes[headerSize:]
}

func (v Value) TagAndDataBytes() []byte {
	__antithesis_instrumentation__.Notify(158884)
	return v.RawBytes[tagPos:]
}

func (v *Value) ensureRawBytes(size int) {
	__antithesis_instrumentation__.Notify(158885)
	if cap(v.RawBytes) < size {
		__antithesis_instrumentation__.Notify(158887)
		v.RawBytes = make([]byte, size)
		return
	} else {
		__antithesis_instrumentation__.Notify(158888)
	}
	__antithesis_instrumentation__.Notify(158886)
	v.RawBytes = v.RawBytes[:size]
	v.setChecksum(checksumUninitialized)
}

func (v Value) EqualTagAndData(o Value) bool {
	__antithesis_instrumentation__.Notify(158889)
	return bytes.Equal(v.TagAndDataBytes(), o.TagAndDataBytes())
}

func (v *Value) SetBytes(b []byte) {
	__antithesis_instrumentation__.Notify(158890)
	v.ensureRawBytes(headerSize + len(b))
	copy(v.dataBytes(), b)
	v.setTag(ValueType_BYTES)
}

func (v *Value) SetTagAndData(b []byte) {
	__antithesis_instrumentation__.Notify(158891)
	v.ensureRawBytes(checksumSize + len(b))
	copy(v.TagAndDataBytes(), b)
}

func (v *Value) SetString(s string) {
	__antithesis_instrumentation__.Notify(158892)
	v.ensureRawBytes(headerSize + len(s))
	copy(v.dataBytes(), s)
	v.setTag(ValueType_BYTES)
}

func (v *Value) SetFloat(f float64) {
	__antithesis_instrumentation__.Notify(158893)
	v.ensureRawBytes(headerSize + 8)
	encoding.EncodeUint64Ascending(v.RawBytes[headerSize:headerSize], math.Float64bits(f))
	v.setTag(ValueType_FLOAT)
}

func (v *Value) SetGeo(so geopb.SpatialObject) error {
	__antithesis_instrumentation__.Notify(158894)
	bytes, err := protoutil.Marshal(&so)
	if err != nil {
		__antithesis_instrumentation__.Notify(158896)
		return err
	} else {
		__antithesis_instrumentation__.Notify(158897)
	}
	__antithesis_instrumentation__.Notify(158895)
	v.ensureRawBytes(headerSize + len(bytes))
	copy(v.dataBytes(), bytes)
	v.setTag(ValueType_GEO)
	return nil
}

func (v *Value) SetBox2D(b geopb.BoundingBox) {
	__antithesis_instrumentation__.Notify(158898)
	v.ensureRawBytes(headerSize + 32)
	encoding.EncodeUint64Ascending(v.RawBytes[headerSize:headerSize], math.Float64bits(b.LoX))
	encoding.EncodeUint64Ascending(v.RawBytes[headerSize+8:headerSize+8], math.Float64bits(b.HiX))
	encoding.EncodeUint64Ascending(v.RawBytes[headerSize+16:headerSize+16], math.Float64bits(b.LoY))
	encoding.EncodeUint64Ascending(v.RawBytes[headerSize+24:headerSize+24], math.Float64bits(b.HiY))
	v.setTag(ValueType_BOX2D)
}

func (v *Value) SetBool(b bool) {
	__antithesis_instrumentation__.Notify(158899)

	v.ensureRawBytes(headerSize + 1)
	i := int64(0)
	if b {
		__antithesis_instrumentation__.Notify(158901)
		i = 1
	} else {
		__antithesis_instrumentation__.Notify(158902)
	}
	__antithesis_instrumentation__.Notify(158900)
	_ = binary.PutVarint(v.RawBytes[headerSize:], i)
	v.setTag(ValueType_INT)
}

func (v *Value) SetInt(i int64) {
	__antithesis_instrumentation__.Notify(158903)
	v.ensureRawBytes(headerSize + binary.MaxVarintLen64)
	n := binary.PutVarint(v.RawBytes[headerSize:], i)
	v.RawBytes = v.RawBytes[:headerSize+n]
	v.setTag(ValueType_INT)
}

func (v *Value) SetProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(158904)

	v.ensureRawBytes(headerSize + msg.Size())
	if _, err := protoutil.MarshalTo(msg, v.RawBytes[headerSize:]); err != nil {
		__antithesis_instrumentation__.Notify(158907)
		return err
	} else {
		__antithesis_instrumentation__.Notify(158908)
	}
	__antithesis_instrumentation__.Notify(158905)

	if _, ok := msg.(*InternalTimeSeriesData); ok {
		__antithesis_instrumentation__.Notify(158909)
		v.setTag(ValueType_TIMESERIES)
	} else {
		__antithesis_instrumentation__.Notify(158910)
		v.setTag(ValueType_BYTES)
	}
	__antithesis_instrumentation__.Notify(158906)
	return nil
}

func (v *Value) SetTime(t time.Time) {
	__antithesis_instrumentation__.Notify(158911)
	const encodingSizeOverestimate = 11
	v.ensureRawBytes(headerSize + encodingSizeOverestimate)
	v.RawBytes = encoding.EncodeTimeAscending(v.RawBytes[:headerSize], t)
	v.setTag(ValueType_TIME)
}

func (v *Value) SetTimeTZ(t timetz.TimeTZ) {
	__antithesis_instrumentation__.Notify(158912)
	v.ensureRawBytes(headerSize + encoding.EncodedTimeTZMaxLen)
	v.RawBytes = encoding.EncodeTimeTZAscending(v.RawBytes[:headerSize], t)
	v.setTag(ValueType_TIMETZ)
}

func (v *Value) SetDuration(t duration.Duration) error {
	__antithesis_instrumentation__.Notify(158913)
	var err error
	v.ensureRawBytes(headerSize + encoding.EncodedDurationMaxLen)
	v.RawBytes, err = encoding.EncodeDurationAscending(v.RawBytes[:headerSize], t)
	if err != nil {
		__antithesis_instrumentation__.Notify(158915)
		return err
	} else {
		__antithesis_instrumentation__.Notify(158916)
	}
	__antithesis_instrumentation__.Notify(158914)
	v.setTag(ValueType_DURATION)
	return nil
}

func (v *Value) SetBitArray(t bitarray.BitArray) {
	__antithesis_instrumentation__.Notify(158917)
	words, _ := t.EncodingParts()
	v.ensureRawBytes(headerSize + encoding.MaxNonsortingUvarintLen + 8*len(words))
	v.RawBytes = encoding.EncodeUntaggedBitArrayValue(v.RawBytes[:headerSize], t)
	v.setTag(ValueType_BITARRAY)
}

func (v *Value) SetDecimal(dec *apd.Decimal) error {
	__antithesis_instrumentation__.Notify(158918)
	decSize := encoding.UpperBoundNonsortingDecimalSize(dec)
	v.ensureRawBytes(headerSize + decSize)
	v.RawBytes = encoding.EncodeNonsortingDecimal(v.RawBytes[:headerSize], dec)
	v.setTag(ValueType_DECIMAL)
	return nil
}

func (v *Value) SetTuple(data []byte) {
	__antithesis_instrumentation__.Notify(158919)
	v.ensureRawBytes(headerSize + len(data))
	copy(v.dataBytes(), data)
	v.setTag(ValueType_TUPLE)
}

func (v Value) GetBytes() ([]byte, error) {
	__antithesis_instrumentation__.Notify(158920)
	if tag := v.GetTag(); tag != ValueType_BYTES {
		__antithesis_instrumentation__.Notify(158922)
		return nil, fmt.Errorf("value type is not %s: %s", ValueType_BYTES, tag)
	} else {
		__antithesis_instrumentation__.Notify(158923)
	}
	__antithesis_instrumentation__.Notify(158921)
	return v.dataBytes(), nil
}

func (v Value) GetFloat() (float64, error) {
	__antithesis_instrumentation__.Notify(158924)
	if tag := v.GetTag(); tag != ValueType_FLOAT {
		__antithesis_instrumentation__.Notify(158928)
		return 0, fmt.Errorf("value type is not %s: %s", ValueType_FLOAT, tag)
	} else {
		__antithesis_instrumentation__.Notify(158929)
	}
	__antithesis_instrumentation__.Notify(158925)
	dataBytes := v.dataBytes()
	if len(dataBytes) != 8 {
		__antithesis_instrumentation__.Notify(158930)
		return 0, fmt.Errorf("float64 value should be exactly 8 bytes: %d", len(dataBytes))
	} else {
		__antithesis_instrumentation__.Notify(158931)
	}
	__antithesis_instrumentation__.Notify(158926)
	_, u, err := encoding.DecodeUint64Ascending(dataBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(158932)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(158933)
	}
	__antithesis_instrumentation__.Notify(158927)
	return math.Float64frombits(u), nil
}

func (v Value) GetGeo() (geopb.SpatialObject, error) {
	__antithesis_instrumentation__.Notify(158934)
	if tag := v.GetTag(); tag != ValueType_GEO {
		__antithesis_instrumentation__.Notify(158936)
		return geopb.SpatialObject{}, fmt.Errorf("value type is not %s: %s", ValueType_GEO, tag)
	} else {
		__antithesis_instrumentation__.Notify(158937)
	}
	__antithesis_instrumentation__.Notify(158935)
	var ret geopb.SpatialObject
	err := protoutil.Unmarshal(v.dataBytes(), &ret)
	return ret, err
}

func (v Value) GetBox2D() (geopb.BoundingBox, error) {
	__antithesis_instrumentation__.Notify(158938)
	box := geopb.BoundingBox{}
	if tag := v.GetTag(); tag != ValueType_BOX2D {
		__antithesis_instrumentation__.Notify(158945)
		return box, fmt.Errorf("value type is not %s: %s", ValueType_BOX2D, tag)
	} else {
		__antithesis_instrumentation__.Notify(158946)
	}
	__antithesis_instrumentation__.Notify(158939)
	dataBytes := v.dataBytes()
	if len(dataBytes) != 32 {
		__antithesis_instrumentation__.Notify(158947)
		return box, fmt.Errorf("float64 value should be exactly 32 bytes: %d", len(dataBytes))
	} else {
		__antithesis_instrumentation__.Notify(158948)
	}
	__antithesis_instrumentation__.Notify(158940)
	var err error
	var val uint64
	dataBytes, val, err = encoding.DecodeUint64Ascending(dataBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(158949)
		return box, err
	} else {
		__antithesis_instrumentation__.Notify(158950)
	}
	__antithesis_instrumentation__.Notify(158941)
	box.LoX = math.Float64frombits(val)
	dataBytes, val, err = encoding.DecodeUint64Ascending(dataBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(158951)
		return box, err
	} else {
		__antithesis_instrumentation__.Notify(158952)
	}
	__antithesis_instrumentation__.Notify(158942)
	box.HiX = math.Float64frombits(val)
	dataBytes, val, err = encoding.DecodeUint64Ascending(dataBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(158953)
		return box, err
	} else {
		__antithesis_instrumentation__.Notify(158954)
	}
	__antithesis_instrumentation__.Notify(158943)
	box.LoY = math.Float64frombits(val)
	_, val, err = encoding.DecodeUint64Ascending(dataBytes)
	if err != nil {
		__antithesis_instrumentation__.Notify(158955)
		return box, err
	} else {
		__antithesis_instrumentation__.Notify(158956)
	}
	__antithesis_instrumentation__.Notify(158944)
	box.HiY = math.Float64frombits(val)

	return box, nil
}

func (v Value) GetBool() (bool, error) {
	__antithesis_instrumentation__.Notify(158957)
	if tag := v.GetTag(); tag != ValueType_INT {
		__antithesis_instrumentation__.Notify(158961)
		return false, fmt.Errorf("value type is not %s: %s", ValueType_INT, tag)
	} else {
		__antithesis_instrumentation__.Notify(158962)
	}
	__antithesis_instrumentation__.Notify(158958)
	i, n := binary.Varint(v.dataBytes())
	if n <= 0 {
		__antithesis_instrumentation__.Notify(158963)
		return false, fmt.Errorf("int64 varint decoding failed: %d", n)
	} else {
		__antithesis_instrumentation__.Notify(158964)
	}
	__antithesis_instrumentation__.Notify(158959)
	if i > 1 || func() bool {
		__antithesis_instrumentation__.Notify(158965)
		return i < 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(158966)
		return false, fmt.Errorf("invalid bool: %d", i)
	} else {
		__antithesis_instrumentation__.Notify(158967)
	}
	__antithesis_instrumentation__.Notify(158960)
	return i != 0, nil
}

func (v Value) GetInt() (int64, error) {
	__antithesis_instrumentation__.Notify(158968)
	if tag := v.GetTag(); tag != ValueType_INT {
		__antithesis_instrumentation__.Notify(158971)
		return 0, fmt.Errorf("value type is not %s: %s", ValueType_INT, tag)
	} else {
		__antithesis_instrumentation__.Notify(158972)
	}
	__antithesis_instrumentation__.Notify(158969)
	i, n := binary.Varint(v.dataBytes())
	if n <= 0 {
		__antithesis_instrumentation__.Notify(158973)
		return 0, fmt.Errorf("int64 varint decoding failed: %d", n)
	} else {
		__antithesis_instrumentation__.Notify(158974)
	}
	__antithesis_instrumentation__.Notify(158970)
	return i, nil
}

func (v Value) GetProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(158975)
	expectedTag := ValueType_BYTES

	if _, ok := msg.(*InternalTimeSeriesData); ok {
		__antithesis_instrumentation__.Notify(158978)
		expectedTag = ValueType_TIMESERIES
	} else {
		__antithesis_instrumentation__.Notify(158979)
	}
	__antithesis_instrumentation__.Notify(158976)

	if tag := v.GetTag(); tag != expectedTag {
		__antithesis_instrumentation__.Notify(158980)
		return fmt.Errorf("value type is not %s: %s", expectedTag, tag)
	} else {
		__antithesis_instrumentation__.Notify(158981)
	}
	__antithesis_instrumentation__.Notify(158977)
	return protoutil.Unmarshal(v.dataBytes(), msg)
}

func (v Value) GetTime() (time.Time, error) {
	__antithesis_instrumentation__.Notify(158982)
	if tag := v.GetTag(); tag != ValueType_TIME {
		__antithesis_instrumentation__.Notify(158984)
		return time.Time{}, fmt.Errorf("value type is not %s: %s", ValueType_TIME, tag)
	} else {
		__antithesis_instrumentation__.Notify(158985)
	}
	__antithesis_instrumentation__.Notify(158983)
	_, t, err := encoding.DecodeTimeAscending(v.dataBytes())
	return t, err
}

func (v Value) GetTimeTZ() (timetz.TimeTZ, error) {
	__antithesis_instrumentation__.Notify(158986)
	if tag := v.GetTag(); tag != ValueType_TIMETZ {
		__antithesis_instrumentation__.Notify(158988)
		return timetz.TimeTZ{}, fmt.Errorf("value type is not %s: %s", ValueType_TIMETZ, tag)
	} else {
		__antithesis_instrumentation__.Notify(158989)
	}
	__antithesis_instrumentation__.Notify(158987)
	_, t, err := encoding.DecodeTimeTZAscending(v.dataBytes())
	return t, err
}

func (v Value) GetDuration() (duration.Duration, error) {
	__antithesis_instrumentation__.Notify(158990)
	if tag := v.GetTag(); tag != ValueType_DURATION {
		__antithesis_instrumentation__.Notify(158992)
		return duration.Duration{}, fmt.Errorf("value type is not %s: %s", ValueType_DURATION, tag)
	} else {
		__antithesis_instrumentation__.Notify(158993)
	}
	__antithesis_instrumentation__.Notify(158991)
	_, t, err := encoding.DecodeDurationAscending(v.dataBytes())
	return t, err
}

func (v Value) GetBitArray() (bitarray.BitArray, error) {
	__antithesis_instrumentation__.Notify(158994)
	if tag := v.GetTag(); tag != ValueType_BITARRAY {
		__antithesis_instrumentation__.Notify(158996)
		return bitarray.BitArray{}, fmt.Errorf("value type is not %s: %s", ValueType_BITARRAY, tag)
	} else {
		__antithesis_instrumentation__.Notify(158997)
	}
	__antithesis_instrumentation__.Notify(158995)
	_, t, err := encoding.DecodeUntaggedBitArrayValue(v.dataBytes())
	return t, err
}

func (v Value) GetDecimal() (apd.Decimal, error) {
	__antithesis_instrumentation__.Notify(158998)
	if tag := v.GetTag(); tag != ValueType_DECIMAL {
		__antithesis_instrumentation__.Notify(159000)
		return apd.Decimal{}, fmt.Errorf("value type is not %s: %s", ValueType_DECIMAL, tag)
	} else {
		__antithesis_instrumentation__.Notify(159001)
	}
	__antithesis_instrumentation__.Notify(158999)
	return encoding.DecodeNonsortingDecimal(v.dataBytes(), nil)
}

func (v Value) GetDecimalInto(d *apd.Decimal) error {
	__antithesis_instrumentation__.Notify(159002)
	if tag := v.GetTag(); tag != ValueType_DECIMAL {
		__antithesis_instrumentation__.Notify(159004)
		return fmt.Errorf("value type is not %s: %s", ValueType_DECIMAL, tag)
	} else {
		__antithesis_instrumentation__.Notify(159005)
	}
	__antithesis_instrumentation__.Notify(159003)
	return encoding.DecodeIntoNonsortingDecimal(d, v.dataBytes(), nil)
}

func (v Value) GetTimeseries() (InternalTimeSeriesData, error) {
	__antithesis_instrumentation__.Notify(159006)
	ts := InternalTimeSeriesData{}

	err := v.GetProto(&ts)
	return ts, err
}

func (v Value) GetTuple() ([]byte, error) {
	__antithesis_instrumentation__.Notify(159007)
	if tag := v.GetTag(); tag != ValueType_TUPLE {
		__antithesis_instrumentation__.Notify(159009)
		return nil, fmt.Errorf("value type is not %s: %s", ValueType_TUPLE, tag)
	} else {
		__antithesis_instrumentation__.Notify(159010)
	}
	__antithesis_instrumentation__.Notify(159008)
	return v.dataBytes(), nil
}

var crc32Pool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(159011)
		return crc32.NewIEEE()
	},
}

func computeChecksum(key, rawBytes []byte, crc hash.Hash32) uint32 {
	__antithesis_instrumentation__.Notify(159012)
	if len(rawBytes) < headerSize {
		__antithesis_instrumentation__.Notify(159017)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(159018)
	}
	__antithesis_instrumentation__.Notify(159013)
	if _, err := crc.Write(key); err != nil {
		__antithesis_instrumentation__.Notify(159019)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(159020)
	}
	__antithesis_instrumentation__.Notify(159014)
	if _, err := crc.Write(rawBytes[checksumSize:]); err != nil {
		__antithesis_instrumentation__.Notify(159021)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(159022)
	}
	__antithesis_instrumentation__.Notify(159015)
	sum := crc.Sum32()
	crc.Reset()

	if sum == checksumUninitialized {
		__antithesis_instrumentation__.Notify(159023)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(159024)
	}
	__antithesis_instrumentation__.Notify(159016)
	return sum
}

func (v Value) computeChecksum(key []byte) uint32 {
	__antithesis_instrumentation__.Notify(159025)
	crc := crc32Pool.Get().(hash.Hash32)
	sum := computeChecksum(key, v.RawBytes, crc)
	crc32Pool.Put(crc)
	return sum
}

func (v Value) PrettyPrint() string {
	__antithesis_instrumentation__.Notify(159026)
	if len(v.RawBytes) == 0 {
		__antithesis_instrumentation__.Notify(159030)
		return "/<empty>"
	} else {
		__antithesis_instrumentation__.Notify(159031)
	}
	__antithesis_instrumentation__.Notify(159027)
	var buf bytes.Buffer
	t := v.GetTag()
	buf.WriteRune('/')
	buf.WriteString(t.String())
	buf.WriteRune('/')

	var err error
	switch t {
	case ValueType_TUPLE:
		__antithesis_instrumentation__.Notify(159032)
		b := v.dataBytes()
		var colID uint32
		for i := 0; len(b) > 0; i++ {
			__antithesis_instrumentation__.Notify(159041)
			if i != 0 {
				__antithesis_instrumentation__.Notify(159045)
				buf.WriteRune('/')
			} else {
				__antithesis_instrumentation__.Notify(159046)
			}
			__antithesis_instrumentation__.Notify(159042)
			_, _, colIDDiff, typ, err := encoding.DecodeValueTag(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(159047)
				break
			} else {
				__antithesis_instrumentation__.Notify(159048)
			}
			__antithesis_instrumentation__.Notify(159043)
			colID += colIDDiff
			var s string
			b, s, err = encoding.PrettyPrintValueEncoded(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(159049)
				break
			} else {
				__antithesis_instrumentation__.Notify(159050)
			}
			__antithesis_instrumentation__.Notify(159044)
			fmt.Fprintf(&buf, "%d:%d:%s/%s", colIDDiff, colID, typ, s)
		}
	case ValueType_INT:
		__antithesis_instrumentation__.Notify(159033)
		var i int64
		i, err = v.GetInt()
		buf.WriteString(strconv.FormatInt(i, 10))
	case ValueType_FLOAT:
		__antithesis_instrumentation__.Notify(159034)
		var f float64
		f, err = v.GetFloat()
		buf.WriteString(strconv.FormatFloat(f, 'g', -1, 64))
	case ValueType_BYTES:
		__antithesis_instrumentation__.Notify(159035)
		var data []byte
		data, err = v.GetBytes()
		if encoding.PrintableBytes(data) {
			__antithesis_instrumentation__.Notify(159051)
			buf.WriteString(string(data))
		} else {
			__antithesis_instrumentation__.Notify(159052)
			buf.WriteString("0x")
			buf.WriteString(hex.EncodeToString(data))
		}
	case ValueType_BITARRAY:
		__antithesis_instrumentation__.Notify(159036)
		var data bitarray.BitArray
		data, err = v.GetBitArray()
		buf.WriteByte('B')
		data.Format(&buf)
	case ValueType_TIME:
		__antithesis_instrumentation__.Notify(159037)
		var t time.Time
		t, err = v.GetTime()
		buf.WriteString(t.UTC().Format(time.RFC3339Nano))
	case ValueType_DECIMAL:
		__antithesis_instrumentation__.Notify(159038)
		var d apd.Decimal
		d, err = v.GetDecimal()
		buf.WriteString(d.String())
	case ValueType_DURATION:
		__antithesis_instrumentation__.Notify(159039)
		var d duration.Duration
		d, err = v.GetDuration()
		buf.WriteString(d.StringNanos())
	default:
		__antithesis_instrumentation__.Notify(159040)
		err = errors.Errorf("unknown tag: %s", t)
	}
	__antithesis_instrumentation__.Notify(159028)
	if err != nil {
		__antithesis_instrumentation__.Notify(159053)

		return fmt.Sprintf("/<err: %s>", err)
	} else {
		__antithesis_instrumentation__.Notify(159054)
	}
	__antithesis_instrumentation__.Notify(159029)
	return buf.String()
}

func (ct InternalCommitTrigger) Kind() redact.SafeString {
	__antithesis_instrumentation__.Notify(159055)
	switch {
	case ct.SplitTrigger != nil:
		__antithesis_instrumentation__.Notify(159056)
		return "split"
	case ct.MergeTrigger != nil:
		__antithesis_instrumentation__.Notify(159057)
		return "merge"
	case ct.ChangeReplicasTrigger != nil:
		__antithesis_instrumentation__.Notify(159058)
		return "change-replicas"
	case ct.ModifiedSpanTrigger != nil:
		__antithesis_instrumentation__.Notify(159059)
		switch {
		case ct.ModifiedSpanTrigger.SystemConfigSpan:
			__antithesis_instrumentation__.Notify(159062)
			return "modified-span (system-config)"
		case ct.ModifiedSpanTrigger.NodeLivenessSpan != nil:
			__antithesis_instrumentation__.Notify(159063)
			return "modified-span (node-liveness)"
		default:
			__antithesis_instrumentation__.Notify(159064)
			panic("unknown modified-span commit trigger kind")
		}
	case ct.StickyBitTrigger != nil:
		__antithesis_instrumentation__.Notify(159060)
		return "sticky-bit"
	default:
		__antithesis_instrumentation__.Notify(159061)
		panic("unknown commit trigger kind")
	}
}

func (ts TransactionStatus) IsFinalized() bool {
	__antithesis_instrumentation__.Notify(159065)
	return ts == COMMITTED || func() bool {
		__antithesis_instrumentation__.Notify(159066)
		return ts == ABORTED == true
	}() == true
}

func (TransactionStatus) SafeValue() { __antithesis_instrumentation__.Notify(159067) }

func MakeTransaction(
	name string,
	baseKey Key,
	userPriority UserPriority,
	now hlc.Timestamp,
	maxOffsetNs int64,
	coordinatorNodeID int32,
) Transaction {
	__antithesis_instrumentation__.Notify(159068)
	u := uuid.FastMakeV4()

	gul := now.Add(maxOffsetNs, 0)

	return Transaction{
		TxnMeta: enginepb.TxnMeta{
			Key:               baseKey,
			ID:                u,
			WriteTimestamp:    now,
			MinTimestamp:      now,
			Priority:          MakePriority(userPriority),
			Sequence:          0,
			CoordinatorNodeID: coordinatorNodeID,
		},
		Name:                   name,
		LastHeartbeat:          now,
		ReadTimestamp:          now,
		GlobalUncertaintyLimit: gul,
	}
}

func (t Transaction) LastActive() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(159069)
	ts := t.LastHeartbeat
	if !t.ReadTimestamp.Synthetic {
		__antithesis_instrumentation__.Notify(159071)
		ts.Forward(t.ReadTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(159072)
	}
	__antithesis_instrumentation__.Notify(159070)
	return ts
}

func (t *Transaction) RequiredFrontier() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(159073)

	ts := t.ReadTimestamp

	ts.Forward(t.WriteTimestamp)

	ts.Forward(t.GlobalUncertaintyLimit)
	return ts
}

func (t Transaction) Clone() *Transaction {
	__antithesis_instrumentation__.Notify(159074)
	return &t
}

func (t *Transaction) AssertInitialized(ctx context.Context) {
	__antithesis_instrumentation__.Notify(159075)
	if t.ID == (uuid.UUID{}) || func() bool {
		__antithesis_instrumentation__.Notify(159076)
		return t.WriteTimestamp.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(159077)
		log.Fatalf(ctx, "uninitialized txn: %s", *t)
	} else {
		__antithesis_instrumentation__.Notify(159078)
	}
}

func MakePriority(userPriority UserPriority) enginepb.TxnPriority {
	__antithesis_instrumentation__.Notify(159079)

	if userPriority < 0 {
		__antithesis_instrumentation__.Notify(159082)
		if -userPriority > UserPriority(math.MaxInt32) {
			__antithesis_instrumentation__.Notify(159084)
			panic(fmt.Sprintf("cannot set explicit priority to a value less than -%d", math.MaxInt32))
		} else {
			__antithesis_instrumentation__.Notify(159085)
		}
		__antithesis_instrumentation__.Notify(159083)
		return enginepb.TxnPriority(-userPriority)
	} else {
		__antithesis_instrumentation__.Notify(159086)
		if userPriority == 0 {
			__antithesis_instrumentation__.Notify(159087)
			userPriority = NormalUserPriority
		} else {
			__antithesis_instrumentation__.Notify(159088)
			if userPriority >= MaxUserPriority {
				__antithesis_instrumentation__.Notify(159089)
				return enginepb.MaxTxnPriority
			} else {
				__antithesis_instrumentation__.Notify(159090)
				if userPriority <= MinUserPriority {
					__antithesis_instrumentation__.Notify(159091)
					return enginepb.MinTxnPriority
				} else {
					__antithesis_instrumentation__.Notify(159092)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(159080)

	val := rand.ExpFloat64() * float64(userPriority)

	val = (val / (5 * float64(MaxUserPriority))) * math.MaxInt32
	if val < float64(enginepb.MinTxnPriority+1) {
		__antithesis_instrumentation__.Notify(159093)
		return enginepb.MinTxnPriority + 1
	} else {
		__antithesis_instrumentation__.Notify(159094)
		if val > float64(enginepb.MaxTxnPriority-1) {
			__antithesis_instrumentation__.Notify(159095)
			return enginepb.MaxTxnPriority - 1
		} else {
			__antithesis_instrumentation__.Notify(159096)
		}
	}
	__antithesis_instrumentation__.Notify(159081)
	return enginepb.TxnPriority(val)
}

func (t *Transaction) Restart(
	userPriority UserPriority, upgradePriority enginepb.TxnPriority, timestamp hlc.Timestamp,
) {
	__antithesis_instrumentation__.Notify(159097)
	t.BumpEpoch()
	if t.WriteTimestamp.Less(timestamp) {
		__antithesis_instrumentation__.Notify(159099)
		t.WriteTimestamp = timestamp
	} else {
		__antithesis_instrumentation__.Notify(159100)
	}
	__antithesis_instrumentation__.Notify(159098)
	t.ReadTimestamp = t.WriteTimestamp

	t.UpgradePriority(MakePriority(userPriority))
	t.UpgradePriority(upgradePriority)

	t.Sequence = 0
	t.WriteTooOld = false
	t.CommitTimestampFixed = false
	t.LockSpans = nil
	t.InFlightWrites = nil
	t.IgnoredSeqNums = nil
}

func (t *Transaction) BumpEpoch() {
	__antithesis_instrumentation__.Notify(159101)
	t.Epoch++
}

func (t *Transaction) Refresh(timestamp hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(159102)
	t.WriteTimestamp.Forward(timestamp)
	t.ReadTimestamp.Forward(t.WriteTimestamp)
	t.WriteTooOld = false
}

func (t *Transaction) Update(o *Transaction) {
	__antithesis_instrumentation__.Notify(159103)
	if o == nil {
		__antithesis_instrumentation__.Notify(159110)
		return
	} else {
		__antithesis_instrumentation__.Notify(159111)
	}
	__antithesis_instrumentation__.Notify(159104)
	o.AssertInitialized(context.TODO())
	if t.ID == (uuid.UUID{}) {
		__antithesis_instrumentation__.Notify(159112)
		*t = *o
		return
	} else {
		__antithesis_instrumentation__.Notify(159113)
		if t.ID != o.ID {
			__antithesis_instrumentation__.Notify(159114)
			log.Fatalf(context.Background(), "updating txn %s with different txn %s", t.String(), o.String())
			return
		} else {
			__antithesis_instrumentation__.Notify(159115)
		}
	}
	__antithesis_instrumentation__.Notify(159105)
	if len(t.Key) == 0 {
		__antithesis_instrumentation__.Notify(159116)
		t.Key = o.Key
	} else {
		__antithesis_instrumentation__.Notify(159117)
	}
	__antithesis_instrumentation__.Notify(159106)

	if t.Epoch < o.Epoch {
		__antithesis_instrumentation__.Notify(159118)

		t.Epoch = o.Epoch
		t.Status = o.Status
		t.WriteTooOld = o.WriteTooOld
		t.CommitTimestampFixed = o.CommitTimestampFixed
		t.Sequence = o.Sequence
		t.LockSpans = o.LockSpans
		t.InFlightWrites = o.InFlightWrites
		t.IgnoredSeqNums = o.IgnoredSeqNums
	} else {
		__antithesis_instrumentation__.Notify(159119)
		if t.Epoch == o.Epoch {
			__antithesis_instrumentation__.Notify(159120)

			switch t.Status {
			case PENDING:
				__antithesis_instrumentation__.Notify(159126)
				t.Status = o.Status
			case STAGING:
				__antithesis_instrumentation__.Notify(159127)
				if o.Status != PENDING {
					__antithesis_instrumentation__.Notify(159131)
					t.Status = o.Status
				} else {
					__antithesis_instrumentation__.Notify(159132)
				}
			case ABORTED:
				__antithesis_instrumentation__.Notify(159128)
				if o.Status == COMMITTED {
					__antithesis_instrumentation__.Notify(159133)
					log.Warningf(context.Background(), "updating ABORTED txn %s with COMMITTED txn %s", t.String(), o.String())
				} else {
					__antithesis_instrumentation__.Notify(159134)
				}
			case COMMITTED:
				__antithesis_instrumentation__.Notify(159129)
			default:
				__antithesis_instrumentation__.Notify(159130)

			}
			__antithesis_instrumentation__.Notify(159121)

			if t.ReadTimestamp == o.ReadTimestamp {
				__antithesis_instrumentation__.Notify(159135)

				t.WriteTooOld = t.WriteTooOld || func() bool {
					__antithesis_instrumentation__.Notify(159136)
					return o.WriteTooOld == true
				}() == true
				t.CommitTimestampFixed = t.CommitTimestampFixed || func() bool {
					__antithesis_instrumentation__.Notify(159137)
					return o.CommitTimestampFixed == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(159138)
				if t.ReadTimestamp.Less(o.ReadTimestamp) {
					__antithesis_instrumentation__.Notify(159139)

					t.WriteTooOld = o.WriteTooOld
					t.CommitTimestampFixed = o.CommitTimestampFixed
				} else {
					__antithesis_instrumentation__.Notify(159140)
				}
			}
			__antithesis_instrumentation__.Notify(159122)

			if t.Sequence < o.Sequence {
				__antithesis_instrumentation__.Notify(159141)
				t.Sequence = o.Sequence
			} else {
				__antithesis_instrumentation__.Notify(159142)
			}
			__antithesis_instrumentation__.Notify(159123)
			if len(o.LockSpans) > 0 {
				__antithesis_instrumentation__.Notify(159143)
				t.LockSpans = o.LockSpans
			} else {
				__antithesis_instrumentation__.Notify(159144)
			}
			__antithesis_instrumentation__.Notify(159124)
			if len(o.InFlightWrites) > 0 {
				__antithesis_instrumentation__.Notify(159145)
				t.InFlightWrites = o.InFlightWrites
			} else {
				__antithesis_instrumentation__.Notify(159146)
			}
			__antithesis_instrumentation__.Notify(159125)
			if len(o.IgnoredSeqNums) > 0 {
				__antithesis_instrumentation__.Notify(159147)
				t.IgnoredSeqNums = o.IgnoredSeqNums
			} else {
				__antithesis_instrumentation__.Notify(159148)
			}
		} else {
			__antithesis_instrumentation__.Notify(159149)

			switch o.Status {
			case ABORTED:
				__antithesis_instrumentation__.Notify(159150)

				t.Status = ABORTED
			case COMMITTED:
				__antithesis_instrumentation__.Notify(159151)
				log.Warningf(context.Background(), "updating txn %s with COMMITTED txn at earlier epoch %s", t.String(), o.String())
			default:
				__antithesis_instrumentation__.Notify(159152)
			}
		}
	}
	__antithesis_instrumentation__.Notify(159107)

	t.WriteTimestamp.Forward(o.WriteTimestamp)
	t.LastHeartbeat.Forward(o.LastHeartbeat)
	t.GlobalUncertaintyLimit.Forward(o.GlobalUncertaintyLimit)
	t.ReadTimestamp.Forward(o.ReadTimestamp)

	if t.MinTimestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(159153)
		t.MinTimestamp = o.MinTimestamp
	} else {
		__antithesis_instrumentation__.Notify(159154)
		if !o.MinTimestamp.IsEmpty() {
			__antithesis_instrumentation__.Notify(159155)
			t.MinTimestamp.Backward(o.MinTimestamp)
		} else {
			__antithesis_instrumentation__.Notify(159156)
		}
	}
	__antithesis_instrumentation__.Notify(159108)

	for _, v := range o.ObservedTimestamps {
		__antithesis_instrumentation__.Notify(159157)
		t.UpdateObservedTimestamp(v.NodeID, v.Timestamp)
	}
	__antithesis_instrumentation__.Notify(159109)

	t.UpgradePriority(o.Priority)
}

func (t *Transaction) UpgradePriority(minPriority enginepb.TxnPriority) {
	__antithesis_instrumentation__.Notify(159158)
	if minPriority > t.Priority && func() bool {
		__antithesis_instrumentation__.Notify(159159)
		return t.Priority != enginepb.MinTxnPriority == true
	}() == true {
		__antithesis_instrumentation__.Notify(159160)
		t.Priority = minPriority
	} else {
		__antithesis_instrumentation__.Notify(159161)
	}
}

func (t *Transaction) IsLocking() bool {
	__antithesis_instrumentation__.Notify(159162)
	return t.Key != nil
}

func (t *Transaction) LocksAsLockUpdates() []LockUpdate {
	__antithesis_instrumentation__.Notify(159163)
	ret := make([]LockUpdate, len(t.LockSpans))
	for i, sp := range t.LockSpans {
		__antithesis_instrumentation__.Notify(159165)
		ret[i] = MakeLockUpdate(t, sp)
	}
	__antithesis_instrumentation__.Notify(159164)
	return ret
}

func (t Transaction) String() string {
	__antithesis_instrumentation__.Notify(159166)
	return redact.StringWithoutMarkers(t)
}

func (t Transaction) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(159167)
	if len(t.Name) > 0 {
		__antithesis_instrumentation__.Notify(159171)
		w.Printf("%q ", redact.SafeString(t.Name))
	} else {
		__antithesis_instrumentation__.Notify(159172)
	}
	__antithesis_instrumentation__.Notify(159168)
	w.Printf("meta={%s} lock=%t stat=%s rts=%s wto=%t gul=%s",
		t.TxnMeta, t.IsLocking(), t.Status, t.ReadTimestamp, t.WriteTooOld, t.GlobalUncertaintyLimit)
	if ni := len(t.LockSpans); t.Status != PENDING && func() bool {
		__antithesis_instrumentation__.Notify(159173)
		return ni > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159174)
		w.Printf(" int=%d", ni)
	} else {
		__antithesis_instrumentation__.Notify(159175)
	}
	__antithesis_instrumentation__.Notify(159169)
	if nw := len(t.InFlightWrites); t.Status != PENDING && func() bool {
		__antithesis_instrumentation__.Notify(159176)
		return nw > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159177)
		w.Printf(" ifw=%d", nw)
	} else {
		__antithesis_instrumentation__.Notify(159178)
	}
	__antithesis_instrumentation__.Notify(159170)
	if ni := len(t.IgnoredSeqNums); ni > 0 {
		__antithesis_instrumentation__.Notify(159179)
		w.Printf(" isn=%d", ni)
	} else {
		__antithesis_instrumentation__.Notify(159180)
	}
}

func (t *Transaction) ResetObservedTimestamps() {
	__antithesis_instrumentation__.Notify(159181)
	t.ObservedTimestamps = nil
}

func (t *Transaction) UpdateObservedTimestamp(nodeID NodeID, timestamp hlc.ClockTimestamp) {
	__antithesis_instrumentation__.Notify(159182)

	if l := len(t.ObservedTimestamps); l == 0 {
		__antithesis_instrumentation__.Notify(159184)
		t.ObservedTimestamps = []ObservedTimestamp{{NodeID: nodeID, Timestamp: timestamp}}
		return
	} else {
		__antithesis_instrumentation__.Notify(159185)
		if l == 1 && func() bool {
			__antithesis_instrumentation__.Notify(159186)
			return t.ObservedTimestamps[0].NodeID == nodeID == true
		}() == true {
			__antithesis_instrumentation__.Notify(159187)
			if timestamp.Less(t.ObservedTimestamps[0].Timestamp) {
				__antithesis_instrumentation__.Notify(159189)
				t.ObservedTimestamps = []ObservedTimestamp{{NodeID: nodeID, Timestamp: timestamp}}
			} else {
				__antithesis_instrumentation__.Notify(159190)
			}
			__antithesis_instrumentation__.Notify(159188)
			return
		} else {
			__antithesis_instrumentation__.Notify(159191)
		}
	}
	__antithesis_instrumentation__.Notify(159183)
	s := observedTimestampSlice(t.ObservedTimestamps)
	t.ObservedTimestamps = s.update(nodeID, timestamp)
}

func (t *Transaction) GetObservedTimestamp(nodeID NodeID) (hlc.ClockTimestamp, bool) {
	__antithesis_instrumentation__.Notify(159192)
	s := observedTimestampSlice(t.ObservedTimestamps)
	return s.get(nodeID)
}

func (t *Transaction) AddIgnoredSeqNumRange(newRange enginepb.IgnoredSeqNumRange) {
	__antithesis_instrumentation__.Notify(159193)

	list := t.IgnoredSeqNums
	i := sort.Search(len(list), func(i int) bool {
		__antithesis_instrumentation__.Notify(159195)
		return list[i].End >= newRange.Start
	})
	__antithesis_instrumentation__.Notify(159194)

	cpy := make([]enginepb.IgnoredSeqNumRange, i+1)
	copy(cpy[:i], list[:i])
	cpy[i] = newRange
	t.IgnoredSeqNums = cpy
}

func (t *Transaction) AsRecord() TransactionRecord {
	__antithesis_instrumentation__.Notify(159196)
	var tr TransactionRecord
	tr.TxnMeta = t.TxnMeta
	tr.Status = t.Status
	tr.LastHeartbeat = t.LastHeartbeat
	tr.LockSpans = t.LockSpans
	tr.InFlightWrites = t.InFlightWrites
	tr.IgnoredSeqNums = t.IgnoredSeqNums
	return tr
}

func (tr *TransactionRecord) AsTransaction() Transaction {
	__antithesis_instrumentation__.Notify(159197)
	var t Transaction
	t.TxnMeta = tr.TxnMeta
	t.Status = tr.Status
	t.LastHeartbeat = tr.LastHeartbeat
	t.LockSpans = tr.LockSpans
	t.InFlightWrites = tr.InFlightWrites
	t.IgnoredSeqNums = tr.IgnoredSeqNums
	return t
}

func PrepareTransactionForRetry(
	ctx context.Context, pErr *Error, pri UserPriority, clock *hlc.Clock,
) Transaction {
	__antithesis_instrumentation__.Notify(159198)
	if pErr.TransactionRestart() == TransactionRestart_NONE {
		__antithesis_instrumentation__.Notify(159203)
		log.Fatalf(ctx, "invalid retryable err (%T): %s", pErr.GetDetail(), pErr)
	} else {
		__antithesis_instrumentation__.Notify(159204)
	}
	__antithesis_instrumentation__.Notify(159199)

	if pErr.GetTxn() == nil {
		__antithesis_instrumentation__.Notify(159205)
		log.Fatalf(ctx, "missing txn for retryable error: %s", pErr)
	} else {
		__antithesis_instrumentation__.Notify(159206)
	}
	__antithesis_instrumentation__.Notify(159200)

	txn := *pErr.GetTxn()
	aborted := false
	switch tErr := pErr.GetDetail().(type) {
	case *TransactionAbortedError:
		__antithesis_instrumentation__.Notify(159207)

		aborted = true

		errTxnPri := txn.Priority

		now := clock.NowAsClockTimestamp()
		txn = MakeTransaction(
			txn.Name,
			nil,

			NormalUserPriority,
			now.ToTimestamp(),
			clock.MaxOffset().Nanoseconds(),
			txn.CoordinatorNodeID,
		)

		txn.Priority = errTxnPri
	case *ReadWithinUncertaintyIntervalError:
		__antithesis_instrumentation__.Notify(159208)
		txn.WriteTimestamp.Forward(tErr.RetryTimestamp())
	case *TransactionPushError:
		__antithesis_instrumentation__.Notify(159209)

		txn.WriteTimestamp.Forward(tErr.PusheeTxn.WriteTimestamp)
		txn.UpgradePriority(tErr.PusheeTxn.Priority - 1)
	case *TransactionRetryError:
		__antithesis_instrumentation__.Notify(159210)

		if tErr.Reason == RETRY_SERIALIZABLE {
			__antithesis_instrumentation__.Notify(159213)

			now := clock.Now()
			txn.WriteTimestamp.Forward(now)
		} else {
			__antithesis_instrumentation__.Notify(159214)
		}
	case *WriteTooOldError:
		__antithesis_instrumentation__.Notify(159211)

		txn.WriteTimestamp.Forward(tErr.RetryTimestamp())
	default:
		__antithesis_instrumentation__.Notify(159212)
		log.Fatalf(ctx, "invalid retryable err (%T): %s", pErr.GetDetail(), pErr)
	}
	__antithesis_instrumentation__.Notify(159201)
	if !aborted {
		__antithesis_instrumentation__.Notify(159215)
		if txn.Status.IsFinalized() {
			__antithesis_instrumentation__.Notify(159217)
			log.Fatalf(ctx, "transaction unexpectedly finalized in (%T): %s", pErr.GetDetail(), pErr)
		} else {
			__antithesis_instrumentation__.Notify(159218)
		}
		__antithesis_instrumentation__.Notify(159216)
		txn.Restart(pri, txn.Priority, txn.WriteTimestamp)
	} else {
		__antithesis_instrumentation__.Notify(159219)
	}
	__antithesis_instrumentation__.Notify(159202)
	return txn
}

func TransactionRefreshTimestamp(pErr *Error) (bool, hlc.Timestamp) {
	__antithesis_instrumentation__.Notify(159220)
	txn := pErr.GetTxn()
	if txn == nil {
		__antithesis_instrumentation__.Notify(159223)
		return false, hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(159224)
	}
	__antithesis_instrumentation__.Notify(159221)
	timestamp := txn.WriteTimestamp
	switch err := pErr.GetDetail().(type) {
	case *TransactionRetryError:
		__antithesis_instrumentation__.Notify(159225)
		if err.Reason != RETRY_SERIALIZABLE && func() bool {
			__antithesis_instrumentation__.Notify(159229)
			return err.Reason != RETRY_WRITE_TOO_OLD == true
		}() == true {
			__antithesis_instrumentation__.Notify(159230)
			return false, hlc.Timestamp{}
		} else {
			__antithesis_instrumentation__.Notify(159231)
		}
	case *WriteTooOldError:
		__antithesis_instrumentation__.Notify(159226)

		timestamp.Forward(err.RetryTimestamp())
	case *ReadWithinUncertaintyIntervalError:
		__antithesis_instrumentation__.Notify(159227)
		timestamp.Forward(err.RetryTimestamp())
	default:
		__antithesis_instrumentation__.Notify(159228)
		return false, hlc.Timestamp{}
	}
	__antithesis_instrumentation__.Notify(159222)
	return true, timestamp
}

func (crt ChangeReplicasTrigger) Replicas() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(159232)
	return crt.Desc.Replicas().Descriptors()
}

func (crt ChangeReplicasTrigger) NextReplicaID() ReplicaID {
	__antithesis_instrumentation__.Notify(159233)
	return crt.Desc.NextReplicaID
}

func (crt ChangeReplicasTrigger) ConfChange(encodedCtx []byte) (raftpb.ConfChangeI, error) {
	__antithesis_instrumentation__.Notify(159234)
	return confChangeImpl(crt, encodedCtx)
}

func (crt ChangeReplicasTrigger) alwaysV2() bool {
	__antithesis_instrumentation__.Notify(159235)

	return false
}

func confChangeImpl(
	crt interface {
		Added() []ReplicaDescriptor
		Removed() []ReplicaDescriptor
		Replicas() []ReplicaDescriptor
		alwaysV2() bool
	},
	encodedCtx []byte,
) (raftpb.ConfChangeI, error) {
	__antithesis_instrumentation__.Notify(159236)
	added, removed, replicas := crt.Added(), crt.Removed(), crt.Replicas()

	var sl []raftpb.ConfChangeSingle

	checkExists := func(in ReplicaDescriptor) error {
		__antithesis_instrumentation__.Notify(159244)
		for _, rDesc := range replicas {
			__antithesis_instrumentation__.Notify(159246)
			if rDesc.ReplicaID == in.ReplicaID {
				__antithesis_instrumentation__.Notify(159247)
				if a, b := in.GetType(), rDesc.GetType(); a != b {
					__antithesis_instrumentation__.Notify(159249)
					return errors.Errorf("have %s, but descriptor has %s", in, rDesc)
				} else {
					__antithesis_instrumentation__.Notify(159250)
				}
				__antithesis_instrumentation__.Notify(159248)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(159251)
			}
		}
		__antithesis_instrumentation__.Notify(159245)
		return errors.Errorf("%s missing from descriptors %v", in, replicas)
	}
	__antithesis_instrumentation__.Notify(159237)
	checkNotExists := func(in ReplicaDescriptor) error {
		__antithesis_instrumentation__.Notify(159252)
		for _, rDesc := range replicas {
			__antithesis_instrumentation__.Notify(159254)
			if rDesc.ReplicaID == in.ReplicaID {
				__antithesis_instrumentation__.Notify(159255)
				return errors.Errorf("%s must no longer be present in descriptor", in)
			} else {
				__antithesis_instrumentation__.Notify(159256)
			}
		}
		__antithesis_instrumentation__.Notify(159253)
		return nil
	}
	__antithesis_instrumentation__.Notify(159238)

	for _, rDesc := range removed {
		__antithesis_instrumentation__.Notify(159257)
		sl = append(sl, raftpb.ConfChangeSingle{
			Type:   raftpb.ConfChangeRemoveNode,
			NodeID: uint64(rDesc.ReplicaID),
		})

		switch rDesc.GetType() {
		case VOTER_OUTGOING:
			__antithesis_instrumentation__.Notify(159258)

			if err := checkExists(rDesc); err != nil {
				__antithesis_instrumentation__.Notify(159265)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(159266)
			}
		case VOTER_DEMOTING_LEARNER, VOTER_DEMOTING_NON_VOTER:
			__antithesis_instrumentation__.Notify(159259)

			if err := checkExists(rDesc); err != nil {
				__antithesis_instrumentation__.Notify(159267)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(159268)
			}
			__antithesis_instrumentation__.Notify(159260)

			sl = append(sl, raftpb.ConfChangeSingle{
				Type:   raftpb.ConfChangeAddLearnerNode,
				NodeID: uint64(rDesc.ReplicaID),
			})
		case LEARNER:
			__antithesis_instrumentation__.Notify(159261)

			if err := checkNotExists(rDesc); err != nil {
				__antithesis_instrumentation__.Notify(159269)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(159270)
			}
		case NON_VOTER:
			__antithesis_instrumentation__.Notify(159262)

			if err := checkNotExists(rDesc); err != nil {
				__antithesis_instrumentation__.Notify(159271)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(159272)
			}
		case VOTER_FULL:
			__antithesis_instrumentation__.Notify(159263)

			if err := checkNotExists(rDesc); err != nil {
				__antithesis_instrumentation__.Notify(159273)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(159274)
			}
		default:
			__antithesis_instrumentation__.Notify(159264)
			return nil, errors.Errorf("can't remove replica in state %v", rDesc.GetType())
		}
	}
	__antithesis_instrumentation__.Notify(159239)

	for _, rDesc := range added {
		__antithesis_instrumentation__.Notify(159275)

		if err := checkExists(rDesc); err != nil {
			__antithesis_instrumentation__.Notify(159278)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(159279)
		}
		__antithesis_instrumentation__.Notify(159276)

		var changeType raftpb.ConfChangeType
		switch rDesc.GetType() {
		case VOTER_FULL:
			__antithesis_instrumentation__.Notify(159280)

			changeType = raftpb.ConfChangeAddNode
		case VOTER_INCOMING:
			__antithesis_instrumentation__.Notify(159281)

			changeType = raftpb.ConfChangeAddNode
		case LEARNER, NON_VOTER:
			__antithesis_instrumentation__.Notify(159282)

			changeType = raftpb.ConfChangeAddLearnerNode
		default:
			__antithesis_instrumentation__.Notify(159283)

			return nil, errors.Errorf("can't add replica in state %v", rDesc.GetType())
		}
		__antithesis_instrumentation__.Notify(159277)
		sl = append(sl, raftpb.ConfChangeSingle{
			Type:   changeType,
			NodeID: uint64(rDesc.ReplicaID),
		})
	}
	__antithesis_instrumentation__.Notify(159240)

	var enteringJoint bool
	for _, rDesc := range replicas {
		__antithesis_instrumentation__.Notify(159284)
		switch rDesc.GetType() {
		case VOTER_INCOMING, VOTER_OUTGOING, VOTER_DEMOTING_LEARNER, VOTER_DEMOTING_NON_VOTER:
			__antithesis_instrumentation__.Notify(159285)
			enteringJoint = true
		default:
			__antithesis_instrumentation__.Notify(159286)
		}
	}
	__antithesis_instrumentation__.Notify(159241)
	wantLeaveJoint := len(added)+len(removed) == 0
	if !enteringJoint {
		__antithesis_instrumentation__.Notify(159287)
		if len(added)+len(removed) > 1 {
			__antithesis_instrumentation__.Notify(159288)
			return nil, errors.Errorf("change requires joint consensus")
		} else {
			__antithesis_instrumentation__.Notify(159289)
		}
	} else {
		__antithesis_instrumentation__.Notify(159290)
		if wantLeaveJoint {
			__antithesis_instrumentation__.Notify(159291)
			return nil, errors.Errorf("descriptor enters joint state, but trigger is requesting to leave one")
		} else {
			__antithesis_instrumentation__.Notify(159292)
		}
	}
	__antithesis_instrumentation__.Notify(159242)

	var cc raftpb.ConfChangeI

	if enteringJoint || func() bool {
		__antithesis_instrumentation__.Notify(159293)
		return crt.alwaysV2() == true
	}() == true {
		__antithesis_instrumentation__.Notify(159294)

		transition := raftpb.ConfChangeTransitionJointExplicit
		if !enteringJoint {
			__antithesis_instrumentation__.Notify(159296)

			transition = raftpb.ConfChangeTransitionAuto
		} else {
			__antithesis_instrumentation__.Notify(159297)
		}
		__antithesis_instrumentation__.Notify(159295)
		cc = raftpb.ConfChangeV2{
			Transition: transition,
			Changes:    sl,
			Context:    encodedCtx,
		}
	} else {
		__antithesis_instrumentation__.Notify(159298)
		if wantLeaveJoint {
			__antithesis_instrumentation__.Notify(159299)

			cc = raftpb.ConfChangeV2{
				Context: encodedCtx,
			}
		} else {
			__antithesis_instrumentation__.Notify(159300)

			cc = raftpb.ConfChange{
				Type:    sl[0].Type,
				NodeID:  sl[0].NodeID,
				Context: encodedCtx,
			}
		}
	}
	__antithesis_instrumentation__.Notify(159243)
	return cc, nil
}

var _ fmt.Stringer = &ChangeReplicasTrigger{}

func (crt ChangeReplicasTrigger) String() string {
	__antithesis_instrumentation__.Notify(159301)
	return redact.StringWithoutMarkers(crt)
}

func (crt ChangeReplicasTrigger) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(159302)
	var nextReplicaID ReplicaID
	var afterReplicas []ReplicaDescriptor
	added, removed := crt.Added(), crt.Removed()
	nextReplicaID = crt.Desc.NextReplicaID

	afterReplicas = crt.Desc.InternalReplicas
	cc, err := crt.ConfChange(nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(159306)
		w.Printf("<malformed ChangeReplicasTrigger: %s>", err)
	} else {
		__antithesis_instrumentation__.Notify(159307)
		ccv2 := cc.AsV2()
		if ccv2.LeaveJoint() {
			__antithesis_instrumentation__.Notify(159308)

			w.SafeString("LEAVE_JOINT")
		} else {
			__antithesis_instrumentation__.Notify(159309)
			if _, ok := ccv2.EnterJoint(); ok {
				__antithesis_instrumentation__.Notify(159310)
				w.Printf("ENTER_JOINT(%s) ", confChangesToRedactableString(ccv2.Changes))
			} else {
				__antithesis_instrumentation__.Notify(159311)
				w.Printf("SIMPLE(%s) ", confChangesToRedactableString(ccv2.Changes))
			}
		}
	}
	__antithesis_instrumentation__.Notify(159303)
	if len(added) > 0 {
		__antithesis_instrumentation__.Notify(159312)
		w.Printf("%s", added)
	} else {
		__antithesis_instrumentation__.Notify(159313)
	}
	__antithesis_instrumentation__.Notify(159304)
	if len(removed) > 0 {
		__antithesis_instrumentation__.Notify(159314)
		if len(added) > 0 {
			__antithesis_instrumentation__.Notify(159316)
			w.SafeString(", ")
		} else {
			__antithesis_instrumentation__.Notify(159317)
		}
		__antithesis_instrumentation__.Notify(159315)
		w.Printf("%s", removed)
	} else {
		__antithesis_instrumentation__.Notify(159318)
	}
	__antithesis_instrumentation__.Notify(159305)
	w.Printf(": after=%s next=%d", afterReplicas, nextReplicaID)
}

func confChangesToRedactableString(ccs []raftpb.ConfChangeSingle) redact.RedactableString {
	__antithesis_instrumentation__.Notify(159319)
	return redact.Sprintfn(func(w redact.SafePrinter) {
		__antithesis_instrumentation__.Notify(159320)
		for i, cc := range ccs {
			__antithesis_instrumentation__.Notify(159321)
			if i > 0 {
				__antithesis_instrumentation__.Notify(159324)
				w.SafeRune(' ')
			} else {
				__antithesis_instrumentation__.Notify(159325)
			}
			__antithesis_instrumentation__.Notify(159322)
			switch cc.Type {
			case raftpb.ConfChangeAddNode:
				__antithesis_instrumentation__.Notify(159326)
				w.SafeRune('v')
			case raftpb.ConfChangeAddLearnerNode:
				__antithesis_instrumentation__.Notify(159327)
				w.SafeRune('l')
			case raftpb.ConfChangeRemoveNode:
				__antithesis_instrumentation__.Notify(159328)
				w.SafeRune('r')
			case raftpb.ConfChangeUpdateNode:
				__antithesis_instrumentation__.Notify(159329)
				w.SafeRune('u')
			default:
				__antithesis_instrumentation__.Notify(159330)
				w.SafeString("unknown")
			}
			__antithesis_instrumentation__.Notify(159323)
			w.Print(cc.NodeID)
		}
	})
}

func (crt ChangeReplicasTrigger) Added() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(159331)
	return crt.InternalAddedReplicas
}

func (crt ChangeReplicasTrigger) Removed() []ReplicaDescriptor {
	__antithesis_instrumentation__.Notify(159332)
	return crt.InternalRemovedReplicas
}

type LeaseSequence int64

func (s LeaseSequence) SafeValue() { __antithesis_instrumentation__.Notify(159333) }

var _ fmt.Stringer = &Lease{}

func (l Lease) String() string {
	__antithesis_instrumentation__.Notify(159334)
	return redact.StringWithoutMarkers(l)
}

func (l Lease) SafeFormat(w redact.SafePrinter, _ rune) {
	__antithesis_instrumentation__.Notify(159335)
	if l.Empty() {
		__antithesis_instrumentation__.Notify(159338)
		w.SafeString("<empty>")
		return
	} else {
		__antithesis_instrumentation__.Notify(159339)
	}
	__antithesis_instrumentation__.Notify(159336)
	if l.Type() == LeaseExpiration {
		__antithesis_instrumentation__.Notify(159340)
		w.Printf("repl=%s seq=%d start=%s exp=%s", l.Replica, l.Sequence, l.Start, l.Expiration)
	} else {
		__antithesis_instrumentation__.Notify(159341)
		w.Printf("repl=%s seq=%d start=%s epo=%d", l.Replica, l.Sequence, l.Start, l.Epoch)
	}
	__antithesis_instrumentation__.Notify(159337)
	if l.ProposedTS != nil {
		__antithesis_instrumentation__.Notify(159342)
		w.Printf(" pro=%s", l.ProposedTS)
	} else {
		__antithesis_instrumentation__.Notify(159343)
	}
}

func (l *Lease) Empty() bool {
	__antithesis_instrumentation__.Notify(159344)
	return l == nil || func() bool {
		__antithesis_instrumentation__.Notify(159345)
		return *l == (Lease{}) == true
	}() == true
}

func (l Lease) OwnedBy(storeID StoreID) bool {
	__antithesis_instrumentation__.Notify(159346)
	return l.Replica.StoreID == storeID
}

type LeaseType int

const (
	LeaseNone LeaseType = iota

	LeaseExpiration

	LeaseEpoch
)

func (l Lease) Type() LeaseType {
	__antithesis_instrumentation__.Notify(159347)
	if l.Epoch == 0 {
		__antithesis_instrumentation__.Notify(159349)
		return LeaseExpiration
	} else {
		__antithesis_instrumentation__.Notify(159350)
	}
	__antithesis_instrumentation__.Notify(159348)
	return LeaseEpoch
}

func (l Lease) Speculative() bool {
	__antithesis_instrumentation__.Notify(159351)
	return l.Sequence == 0
}

func (l Lease) Equivalent(newL Lease) bool {
	__antithesis_instrumentation__.Notify(159352)

	l.ProposedTS, newL.ProposedTS = nil, nil
	l.DeprecatedStartStasis, newL.DeprecatedStartStasis = nil, nil

	l.Sequence, newL.Sequence = 0, 0

	l.AcquisitionType, newL.AcquisitionType = 0, 0

	l.Replica.Type, newL.Replica.Type = nil, nil

	switch l.Type() {
	case LeaseEpoch:
		__antithesis_instrumentation__.Notify(159354)

		l.Expiration, newL.Expiration = nil, nil

		if l.Epoch == newL.Epoch {
			__antithesis_instrumentation__.Notify(159357)
			l.Epoch, newL.Epoch = 0, 0
		} else {
			__antithesis_instrumentation__.Notify(159358)
		}
	case LeaseExpiration:
		__antithesis_instrumentation__.Notify(159355)

		l.Epoch, newL.Epoch = 0, 0

		if l.GetExpiration().LessEq(newL.GetExpiration()) {
			__antithesis_instrumentation__.Notify(159359)
			l.Expiration, newL.Expiration = nil, nil
		} else {
			__antithesis_instrumentation__.Notify(159360)
		}
	default:
		__antithesis_instrumentation__.Notify(159356)
	}
	__antithesis_instrumentation__.Notify(159353)
	return l == newL
}

func (l Lease) GetExpiration() hlc.Timestamp {
	__antithesis_instrumentation__.Notify(159361)
	if l.Expiration == nil {
		__antithesis_instrumentation__.Notify(159363)
		return hlc.Timestamp{}
	} else {
		__antithesis_instrumentation__.Notify(159364)
	}
	__antithesis_instrumentation__.Notify(159362)
	return *l.Expiration
}

func equivalentTimestamps(a, b *hlc.Timestamp) bool {
	__antithesis_instrumentation__.Notify(159365)
	if a == nil {
		__antithesis_instrumentation__.Notify(159367)
		if b == nil {
			__antithesis_instrumentation__.Notify(159369)
			return true
		} else {
			__antithesis_instrumentation__.Notify(159370)
		}
		__antithesis_instrumentation__.Notify(159368)
		if b.IsEmpty() {
			__antithesis_instrumentation__.Notify(159371)
			return true
		} else {
			__antithesis_instrumentation__.Notify(159372)
		}
	} else {
		__antithesis_instrumentation__.Notify(159373)
		if b == nil {
			__antithesis_instrumentation__.Notify(159374)
			if a.IsEmpty() {
				__antithesis_instrumentation__.Notify(159375)
				return true
			} else {
				__antithesis_instrumentation__.Notify(159376)
			}
		} else {
			__antithesis_instrumentation__.Notify(159377)
		}
	}
	__antithesis_instrumentation__.Notify(159366)
	return a.Equal(b)
}

func (l *Lease) Equal(that interface{}) bool {
	__antithesis_instrumentation__.Notify(159378)
	if that == nil {
		__antithesis_instrumentation__.Notify(159389)
		return l == nil
	} else {
		__antithesis_instrumentation__.Notify(159390)
	}
	__antithesis_instrumentation__.Notify(159379)

	that1, ok := that.(*Lease)
	if !ok {
		__antithesis_instrumentation__.Notify(159391)
		that2, ok := that.(Lease)
		if ok {
			__antithesis_instrumentation__.Notify(159392)
			that1 = &that2
		} else {
			__antithesis_instrumentation__.Notify(159393)
			panic(errors.AssertionFailedf("attempting to compare lease to %T", that))
		}
	} else {
		__antithesis_instrumentation__.Notify(159394)
	}
	__antithesis_instrumentation__.Notify(159380)
	if that1 == nil {
		__antithesis_instrumentation__.Notify(159395)
		return l == nil
	} else {
		__antithesis_instrumentation__.Notify(159396)
		if l == nil {
			__antithesis_instrumentation__.Notify(159397)
			return false
		} else {
			__antithesis_instrumentation__.Notify(159398)
		}
	}
	__antithesis_instrumentation__.Notify(159381)

	if !l.Start.Equal(&that1.Start) {
		__antithesis_instrumentation__.Notify(159399)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159400)
	}
	__antithesis_instrumentation__.Notify(159382)
	if !equivalentTimestamps(l.Expiration, that1.Expiration) {
		__antithesis_instrumentation__.Notify(159401)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159402)
	}
	__antithesis_instrumentation__.Notify(159383)
	if !l.Replica.Equal(&that1.Replica) {
		__antithesis_instrumentation__.Notify(159403)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159404)
	}
	__antithesis_instrumentation__.Notify(159384)
	if !equivalentTimestamps(l.DeprecatedStartStasis, that1.DeprecatedStartStasis) {
		__antithesis_instrumentation__.Notify(159405)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159406)
	}
	__antithesis_instrumentation__.Notify(159385)
	if !l.ProposedTS.Equal(that1.ProposedTS) {
		__antithesis_instrumentation__.Notify(159407)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159408)
	}
	__antithesis_instrumentation__.Notify(159386)
	if l.Epoch != that1.Epoch {
		__antithesis_instrumentation__.Notify(159409)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159410)
	}
	__antithesis_instrumentation__.Notify(159387)
	if l.Sequence != that1.Sequence {
		__antithesis_instrumentation__.Notify(159411)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159412)
	}
	__antithesis_instrumentation__.Notify(159388)
	return true
}

func MakeIntent(txn *enginepb.TxnMeta, key Key) Intent {
	__antithesis_instrumentation__.Notify(159413)
	var i Intent
	i.Key = key
	i.Txn = *txn
	return i
}

func AsIntents(txn *enginepb.TxnMeta, keys []Key) []Intent {
	__antithesis_instrumentation__.Notify(159414)
	ret := make([]Intent, len(keys))
	for i := range keys {
		__antithesis_instrumentation__.Notify(159416)
		ret[i] = MakeIntent(txn, keys[i])
	}
	__antithesis_instrumentation__.Notify(159415)
	return ret
}

func MakeLockAcquisition(txn *Transaction, key Key, dur lock.Durability) LockAcquisition {
	__antithesis_instrumentation__.Notify(159417)
	return LockAcquisition{Span: Span{Key: key}, Txn: txn.TxnMeta, Durability: dur}
}

func MakeLockUpdate(txn *Transaction, span Span) LockUpdate {
	__antithesis_instrumentation__.Notify(159418)
	u := LockUpdate{Span: span}
	u.SetTxn(txn)
	return u
}

func (u *LockUpdate) SetTxn(txn *Transaction) {
	__antithesis_instrumentation__.Notify(159419)
	u.Txn = txn.TxnMeta
	u.Status = txn.Status
	u.IgnoredSeqNums = txn.IgnoredSeqNums
}

func (ls LockStateInfo) SafeFormat(w redact.SafePrinter, r rune) {
	__antithesis_instrumentation__.Notify(159420)
	expand := w.Flag('+')
	w.Printf("range_id=%d key=%s ", ls.RangeID, ls.Key)
	redactableLockHolder := redact.Sprint(nil)
	if ls.LockHolder != nil {
		__antithesis_instrumentation__.Notify(159422)
		if expand {
			__antithesis_instrumentation__.Notify(159423)
			redactableLockHolder = redact.Sprint(ls.LockHolder.ID)
		} else {
			__antithesis_instrumentation__.Notify(159424)
			redactableLockHolder = redact.Sprint(ls.LockHolder.Short())
		}
	} else {
		__antithesis_instrumentation__.Notify(159425)
	}
	__antithesis_instrumentation__.Notify(159421)
	w.Printf("holder=%s ", redactableLockHolder)
	w.Printf("durability=%s ", ls.Durability)
	w.Printf("duration=%s", ls.HoldDuration)
	if len(ls.Waiters) > 0 {
		__antithesis_instrumentation__.Notify(159426)
		w.Printf("\n waiters:")

		for _, lw := range ls.Waiters {
			__antithesis_instrumentation__.Notify(159427)
			if expand {
				__antithesis_instrumentation__.Notify(159428)
				w.Printf("\n  %+v", lw)
			} else {
				__antithesis_instrumentation__.Notify(159429)
				w.Printf("\n  %s", lw)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(159430)
	}
}

func (ls LockStateInfo) String() string {
	__antithesis_instrumentation__.Notify(159431)
	return redact.StringWithoutMarkers(ls)
}

func (s Span) EqualValue(o Span) bool {
	__antithesis_instrumentation__.Notify(159432)
	return s.Key.Equal(o.Key) && func() bool {
		__antithesis_instrumentation__.Notify(159433)
		return s.EndKey.Equal(o.EndKey) == true
	}() == true
}

func (s Span) Equal(o Span) bool {
	__antithesis_instrumentation__.Notify(159434)
	return s.Key.Equal(o.Key) && func() bool {
		__antithesis_instrumentation__.Notify(159435)
		return s.EndKey.Equal(o.EndKey) == true
	}() == true
}

func (s Span) Overlaps(o Span) bool {
	__antithesis_instrumentation__.Notify(159436)
	if !s.Valid() || func() bool {
		__antithesis_instrumentation__.Notify(159439)
		return !o.Valid() == true
	}() == true {
		__antithesis_instrumentation__.Notify(159440)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159441)
	}
	__antithesis_instrumentation__.Notify(159437)

	if len(s.EndKey) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(159442)
		return len(o.EndKey) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159443)
		return s.Key.Equal(o.Key)
	} else {
		__antithesis_instrumentation__.Notify(159444)
		if len(s.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(159445)
			return bytes.Compare(s.Key, o.Key) >= 0 && func() bool {
				__antithesis_instrumentation__.Notify(159446)
				return bytes.Compare(s.Key, o.EndKey) < 0 == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(159447)
			if len(o.EndKey) == 0 {
				__antithesis_instrumentation__.Notify(159448)
				return bytes.Compare(o.Key, s.Key) >= 0 && func() bool {
					__antithesis_instrumentation__.Notify(159449)
					return bytes.Compare(o.Key, s.EndKey) < 0 == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(159450)
			}
		}
	}
	__antithesis_instrumentation__.Notify(159438)
	return bytes.Compare(s.EndKey, o.Key) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(159451)
		return bytes.Compare(s.Key, o.EndKey) < 0 == true
	}() == true
}

func (s Span) Intersect(o Span) Span {
	__antithesis_instrumentation__.Notify(159452)

	if !s.Overlaps(o) {
		__antithesis_instrumentation__.Notify(159458)
		return Span{}
	} else {
		__antithesis_instrumentation__.Notify(159459)
	}
	__antithesis_instrumentation__.Notify(159453)

	if len(s.EndKey) == 0 {
		__antithesis_instrumentation__.Notify(159460)
		return s
	} else {
		__antithesis_instrumentation__.Notify(159461)
	}
	__antithesis_instrumentation__.Notify(159454)
	if len(o.EndKey) == 0 {
		__antithesis_instrumentation__.Notify(159462)
		return o
	} else {
		__antithesis_instrumentation__.Notify(159463)
	}
	__antithesis_instrumentation__.Notify(159455)

	key := s.Key
	if key.Compare(o.Key) < 0 {
		__antithesis_instrumentation__.Notify(159464)
		key = o.Key
	} else {
		__antithesis_instrumentation__.Notify(159465)
	}
	__antithesis_instrumentation__.Notify(159456)
	endKey := s.EndKey
	if endKey.Compare(o.EndKey) > 0 {
		__antithesis_instrumentation__.Notify(159466)
		endKey = o.EndKey
	} else {
		__antithesis_instrumentation__.Notify(159467)
	}
	__antithesis_instrumentation__.Notify(159457)
	return Span{key, endKey}
}

func (s Span) Combine(o Span) Span {
	__antithesis_instrumentation__.Notify(159468)
	if !s.Valid() || func() bool {
		__antithesis_instrumentation__.Notify(159474)
		return !o.Valid() == true
	}() == true {
		__antithesis_instrumentation__.Notify(159475)
		return Span{}
	} else {
		__antithesis_instrumentation__.Notify(159476)
	}
	__antithesis_instrumentation__.Notify(159469)

	min := s.Key
	max := s.Key
	if len(s.EndKey) > 0 {
		__antithesis_instrumentation__.Notify(159477)
		max = s.EndKey
	} else {
		__antithesis_instrumentation__.Notify(159478)
	}
	__antithesis_instrumentation__.Notify(159470)
	if o.Key.Compare(min) < 0 {
		__antithesis_instrumentation__.Notify(159479)
		min = o.Key
	} else {
		__antithesis_instrumentation__.Notify(159480)
		if o.Key.Compare(max) > 0 {
			__antithesis_instrumentation__.Notify(159481)
			max = o.Key
		} else {
			__antithesis_instrumentation__.Notify(159482)
		}
	}
	__antithesis_instrumentation__.Notify(159471)
	if len(o.EndKey) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(159483)
		return o.EndKey.Compare(max) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159484)
		max = o.EndKey
	} else {
		__antithesis_instrumentation__.Notify(159485)
	}
	__antithesis_instrumentation__.Notify(159472)
	if min.Equal(max) {
		__antithesis_instrumentation__.Notify(159486)
		return Span{Key: min}
	} else {
		__antithesis_instrumentation__.Notify(159487)
		if s.Key.Equal(max) || func() bool {
			__antithesis_instrumentation__.Notify(159488)
			return o.Key.Equal(max) == true
		}() == true {
			__antithesis_instrumentation__.Notify(159489)
			return Span{Key: min, EndKey: max.Next()}
		} else {
			__antithesis_instrumentation__.Notify(159490)
		}
	}
	__antithesis_instrumentation__.Notify(159473)
	return Span{Key: min, EndKey: max}
}

func (s Span) Contains(o Span) bool {
	__antithesis_instrumentation__.Notify(159491)
	if !s.Valid() || func() bool {
		__antithesis_instrumentation__.Notify(159494)
		return !o.Valid() == true
	}() == true {
		__antithesis_instrumentation__.Notify(159495)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159496)
	}
	__antithesis_instrumentation__.Notify(159492)

	if len(s.EndKey) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(159497)
		return len(o.EndKey) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159498)
		return s.Key.Equal(o.Key)
	} else {
		__antithesis_instrumentation__.Notify(159499)
		if len(s.EndKey) == 0 {
			__antithesis_instrumentation__.Notify(159500)
			return false
		} else {
			__antithesis_instrumentation__.Notify(159501)
			if len(o.EndKey) == 0 {
				__antithesis_instrumentation__.Notify(159502)
				return bytes.Compare(o.Key, s.Key) >= 0 && func() bool {
					__antithesis_instrumentation__.Notify(159503)
					return bytes.Compare(o.Key, s.EndKey) < 0 == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(159504)
			}
		}
	}
	__antithesis_instrumentation__.Notify(159493)
	return bytes.Compare(s.Key, o.Key) <= 0 && func() bool {
		__antithesis_instrumentation__.Notify(159505)
		return bytes.Compare(s.EndKey, o.EndKey) >= 0 == true
	}() == true
}

func (s Span) ContainsKey(key Key) bool {
	__antithesis_instrumentation__.Notify(159506)
	return bytes.Compare(key, s.Key) >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(159507)
		return bytes.Compare(key, s.EndKey) < 0 == true
	}() == true
}

func (s Span) CompareKey(key Key) int {
	__antithesis_instrumentation__.Notify(159508)
	if bytes.Compare(key, s.Key) >= 0 {
		__antithesis_instrumentation__.Notify(159510)
		if bytes.Compare(key, s.EndKey) < 0 {
			__antithesis_instrumentation__.Notify(159512)
			return 0
		} else {
			__antithesis_instrumentation__.Notify(159513)
		}
		__antithesis_instrumentation__.Notify(159511)
		return 1
	} else {
		__antithesis_instrumentation__.Notify(159514)
	}
	__antithesis_instrumentation__.Notify(159509)
	return -1
}

func (s Span) ProperlyContainsKey(key Key) bool {
	__antithesis_instrumentation__.Notify(159515)
	return bytes.Compare(key, s.Key) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(159516)
		return bytes.Compare(key, s.EndKey) < 0 == true
	}() == true
}

func (s Span) AsRange() interval.Range {
	__antithesis_instrumentation__.Notify(159517)
	startKey := s.Key
	endKey := s.EndKey
	if len(endKey) == 0 {
		__antithesis_instrumentation__.Notify(159519)
		endKey = s.Key.Next()
		startKey = endKey[:len(startKey)]
	} else {
		__antithesis_instrumentation__.Notify(159520)
	}
	__antithesis_instrumentation__.Notify(159518)
	return interval.Range{
		Start: interval.Comparable(startKey),
		End:   interval.Comparable(endKey),
	}
}

func (s Span) String() string {
	__antithesis_instrumentation__.Notify(159521)
	const maxChars = math.MaxInt32
	return PrettyPrintRange(s.Key, s.EndKey, maxChars)
}

func (s Span) SplitOnKey(key Key) (left Span, right Span) {
	__antithesis_instrumentation__.Notify(159522)

	if bytes.Compare(key, s.Key) <= 0 || func() bool {
		__antithesis_instrumentation__.Notify(159524)
		return bytes.Compare(key, s.EndKey) >= 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159525)
		return s, Span{}
	} else {
		__antithesis_instrumentation__.Notify(159526)
	}
	__antithesis_instrumentation__.Notify(159523)

	return Span{Key: s.Key, EndKey: key}, Span{Key: key, EndKey: s.EndKey}
}

func (s Span) Valid() bool {
	__antithesis_instrumentation__.Notify(159527)

	if len(s.Key) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(159531)
		return len(s.EndKey) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(159532)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159533)
	}
	__antithesis_instrumentation__.Notify(159528)

	if len(s.EndKey) == 0 {
		__antithesis_instrumentation__.Notify(159534)
		return true
	} else {
		__antithesis_instrumentation__.Notify(159535)
	}
	__antithesis_instrumentation__.Notify(159529)

	if bytes.Compare(s.Key, s.EndKey) >= 0 {
		__antithesis_instrumentation__.Notify(159536)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159537)
	}
	__antithesis_instrumentation__.Notify(159530)

	return true
}

const SpanOverhead = int64(unsafe.Sizeof(Span{}))

func (s Span) MemUsage() int64 {
	__antithesis_instrumentation__.Notify(159538)
	return SpanOverhead + int64(cap(s.Key)) + int64(cap(s.EndKey))
}

type Spans []Span

func (a Spans) Len() int      { __antithesis_instrumentation__.Notify(159539); return len(a) }
func (a Spans) Swap(i, j int) { __antithesis_instrumentation__.Notify(159540); a[i], a[j] = a[j], a[i] }
func (a Spans) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(159541)
	return a[i].Key.Compare(a[j].Key) < 0
}

func (a Spans) ContainsKey(key Key) bool {
	__antithesis_instrumentation__.Notify(159542)
	for _, span := range a {
		__antithesis_instrumentation__.Notify(159544)
		if span.ContainsKey(key) {
			__antithesis_instrumentation__.Notify(159545)
			return true
		} else {
			__antithesis_instrumentation__.Notify(159546)
		}
	}
	__antithesis_instrumentation__.Notify(159543)

	return false
}

const SpansOverhead = int64(unsafe.Sizeof(Spans{}))

func (a Spans) MemUsage() int64 {
	__antithesis_instrumentation__.Notify(159547)

	aCap := a[:cap(a)]
	size := SpansOverhead
	for i := range aCap {
		__antithesis_instrumentation__.Notify(159549)
		size += aCap[i].MemUsage()
	}
	__antithesis_instrumentation__.Notify(159548)
	return size
}

func (a Spans) String() string {
	__antithesis_instrumentation__.Notify(159550)
	var buf bytes.Buffer
	for i, span := range a {
		__antithesis_instrumentation__.Notify(159552)
		if i != 0 {
			__antithesis_instrumentation__.Notify(159554)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(159555)
		}
		__antithesis_instrumentation__.Notify(159553)
		buf.WriteString(span.String())
	}
	__antithesis_instrumentation__.Notify(159551)
	return buf.String()
}

type RSpan struct {
	Key, EndKey RKey
}

func (rs RSpan) Equal(o RSpan) bool {
	__antithesis_instrumentation__.Notify(159556)
	return rs.Key.Equal(o.Key) && func() bool {
		__antithesis_instrumentation__.Notify(159557)
		return rs.EndKey.Equal(o.EndKey) == true
	}() == true
}

func (rs RSpan) ContainsKey(key RKey) bool {
	__antithesis_instrumentation__.Notify(159558)
	return bytes.Compare(key, rs.Key) >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(159559)
		return bytes.Compare(key, rs.EndKey) < 0 == true
	}() == true
}

func (rs RSpan) ContainsKeyInverted(key RKey) bool {
	__antithesis_instrumentation__.Notify(159560)
	return bytes.Compare(key, rs.Key) > 0 && func() bool {
		__antithesis_instrumentation__.Notify(159561)
		return bytes.Compare(key, rs.EndKey) <= 0 == true
	}() == true
}

func (rs RSpan) ContainsKeyRange(start, end RKey) bool {
	__antithesis_instrumentation__.Notify(159562)
	if len(end) == 0 {
		__antithesis_instrumentation__.Notify(159565)
		return rs.ContainsKey(start)
	} else {
		__antithesis_instrumentation__.Notify(159566)
	}
	__antithesis_instrumentation__.Notify(159563)
	if comp := bytes.Compare(end, start); comp < 0 {
		__antithesis_instrumentation__.Notify(159567)
		return false
	} else {
		__antithesis_instrumentation__.Notify(159568)
		if comp == 0 {
			__antithesis_instrumentation__.Notify(159569)
			return rs.ContainsKey(start)
		} else {
			__antithesis_instrumentation__.Notify(159570)
		}
	}
	__antithesis_instrumentation__.Notify(159564)
	return bytes.Compare(start, rs.Key) >= 0 && func() bool {
		__antithesis_instrumentation__.Notify(159571)
		return bytes.Compare(rs.EndKey, end) >= 0 == true
	}() == true
}

func (rs RSpan) String() string {
	__antithesis_instrumentation__.Notify(159572)
	const maxChars = math.MaxInt32
	return PrettyPrintRange(Key(rs.Key), Key(rs.EndKey), maxChars)
}

func (rs RSpan) Intersect(desc *RangeDescriptor) (RSpan, error) {
	__antithesis_instrumentation__.Notify(159573)
	if !rs.Key.Less(desc.EndKey) || func() bool {
		__antithesis_instrumentation__.Notify(159577)
		return !desc.StartKey.Less(rs.EndKey) == true
	}() == true {
		__antithesis_instrumentation__.Notify(159578)
		return rs, errors.Errorf("span and descriptor's range do not overlap: %s vs %s", rs, desc)
	} else {
		__antithesis_instrumentation__.Notify(159579)
	}
	__antithesis_instrumentation__.Notify(159574)

	key := rs.Key
	if key.Less(desc.StartKey) {
		__antithesis_instrumentation__.Notify(159580)
		key = desc.StartKey
	} else {
		__antithesis_instrumentation__.Notify(159581)
	}
	__antithesis_instrumentation__.Notify(159575)
	endKey := rs.EndKey
	if !desc.ContainsKeyRange(desc.StartKey, endKey) {
		__antithesis_instrumentation__.Notify(159582)
		endKey = desc.EndKey
	} else {
		__antithesis_instrumentation__.Notify(159583)
	}
	__antithesis_instrumentation__.Notify(159576)
	return RSpan{key, endKey}, nil
}

func (rs RSpan) AsRawSpanWithNoLocals() Span {
	__antithesis_instrumentation__.Notify(159584)
	return Span{
		Key:    Key(rs.Key),
		EndKey: Key(rs.EndKey),
	}
}

type KeyValueByKey []KeyValue

func (kv KeyValueByKey) Len() int {
	__antithesis_instrumentation__.Notify(159585)
	return len(kv)
}

func (kv KeyValueByKey) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(159586)
	return bytes.Compare(kv[i].Key, kv[j].Key) < 0
}

func (kv KeyValueByKey) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(159587)
	kv[i], kv[j] = kv[j], kv[i]
}

var _ sort.Interface = KeyValueByKey{}

type observedTimestampSlice []ObservedTimestamp

func (s observedTimestampSlice) index(nodeID NodeID) int {
	__antithesis_instrumentation__.Notify(159588)
	return sort.Search(len(s),
		func(i int) bool {
			__antithesis_instrumentation__.Notify(159589)
			return s[i].NodeID >= nodeID
		},
	)
}

func (s observedTimestampSlice) get(nodeID NodeID) (hlc.ClockTimestamp, bool) {
	__antithesis_instrumentation__.Notify(159590)
	i := s.index(nodeID)
	if i < len(s) && func() bool {
		__antithesis_instrumentation__.Notify(159592)
		return s[i].NodeID == nodeID == true
	}() == true {
		__antithesis_instrumentation__.Notify(159593)
		return s[i].Timestamp, true
	} else {
		__antithesis_instrumentation__.Notify(159594)
	}
	__antithesis_instrumentation__.Notify(159591)
	return hlc.ClockTimestamp{}, false
}

func (s observedTimestampSlice) update(
	nodeID NodeID, timestamp hlc.ClockTimestamp,
) observedTimestampSlice {
	__antithesis_instrumentation__.Notify(159595)
	i := s.index(nodeID)
	if i < len(s) && func() bool {
		__antithesis_instrumentation__.Notify(159597)
		return s[i].NodeID == nodeID == true
	}() == true {
		__antithesis_instrumentation__.Notify(159598)
		if timestamp.Less(s[i].Timestamp) {
			__antithesis_instrumentation__.Notify(159600)

			cpy := make(observedTimestampSlice, len(s))
			copy(cpy, s)
			cpy[i].Timestamp = timestamp
			return cpy
		} else {
			__antithesis_instrumentation__.Notify(159601)
		}
		__antithesis_instrumentation__.Notify(159599)
		return s
	} else {
		__antithesis_instrumentation__.Notify(159602)
	}
	__antithesis_instrumentation__.Notify(159596)

	cpy := make(observedTimestampSlice, len(s)+1)
	copy(cpy[:i], s[:i])
	cpy[i] = ObservedTimestamp{NodeID: nodeID, Timestamp: timestamp}
	copy(cpy[i+1:], s[i:])
	return cpy
}

type SequencedWriteBySeq []SequencedWrite

func (s SequencedWriteBySeq) Len() int { __antithesis_instrumentation__.Notify(159603); return len(s) }

func (s SequencedWriteBySeq) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(159604)
	return s[i].Sequence < s[j].Sequence
}

func (s SequencedWriteBySeq) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(159605)
	s[i], s[j] = s[j], s[i]
}

var _ sort.Interface = SequencedWriteBySeq{}

func (s SequencedWriteBySeq) Find(seq enginepb.TxnSeq) int {
	__antithesis_instrumentation__.Notify(159606)
	if util.RaceEnabled {
		__antithesis_instrumentation__.Notify(159609)
		if !sort.IsSorted(s) {
			__antithesis_instrumentation__.Notify(159610)
			panic("SequencedWriteBySeq must be sorted")
		} else {
			__antithesis_instrumentation__.Notify(159611)
		}
	} else {
		__antithesis_instrumentation__.Notify(159612)
	}
	__antithesis_instrumentation__.Notify(159607)
	if i := sort.Search(len(s), func(i int) bool {
		__antithesis_instrumentation__.Notify(159613)
		return s[i].Sequence >= seq
	}); i < len(s) && func() bool {
		__antithesis_instrumentation__.Notify(159614)
		return s[i].Sequence == seq == true
	}() == true {
		__antithesis_instrumentation__.Notify(159615)
		return i
	} else {
		__antithesis_instrumentation__.Notify(159616)
	}
	__antithesis_instrumentation__.Notify(159608)
	return -1
}

var _ = (SequencedWriteBySeq{}).Find

func init() {

	enginepb.FormatBytesAsKey = func(k []byte) string { return Key(k).String() }
	enginepb.FormatBytesAsValue = func(v []byte) string { return Value{RawBytes: v}.PrettyPrint() }
}

func (ReplicaChangeType) SafeValue() { __antithesis_instrumentation__.Notify(159617) }

func (ri RangeInfo) String() string {
	__antithesis_instrumentation__.Notify(159618)
	return fmt.Sprintf("desc: %s, lease: %s, closed_timestamp_policy: %s",
		ri.Desc, ri.Lease, ri.ClosedTimestampPolicy)
}

func (r *RowCount) Add(other RowCount) {
	__antithesis_instrumentation__.Notify(159619)
	r.DataSize += other.DataSize
	r.Rows += other.Rows
	r.IndexEntries += other.IndexEntries
}
