package rditer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
)

type KeyRange struct {
	Start, End roachpb.Key
}

type ReplicaMVCCDataIterator struct {
	reader   storage.Reader
	curIndex int
	ranges   []KeyRange

	it      storage.MVCCIterator
	err     error
	reverse bool
}

type ReplicaEngineDataIterator struct {
	curIndex int
	ranges   []KeyRange
	it       storage.EngineIterator
	valid    bool
	err      error
}

func MakeAllKeyRanges(d *roachpb.RangeDescriptor) []KeyRange {
	__antithesis_instrumentation__.Notify(114151)
	return makeRangeKeyRanges(d, false)
}

func MakeReplicatedKeyRanges(d *roachpb.RangeDescriptor) []KeyRange {
	__antithesis_instrumentation__.Notify(114152)
	return makeRangeKeyRanges(d, true)
}

func makeRangeKeyRanges(d *roachpb.RangeDescriptor, replicatedOnly bool) []KeyRange {
	__antithesis_instrumentation__.Notify(114153)
	rangeIDLocal := MakeRangeIDLocalKeyRange(d.RangeID, replicatedOnly)
	rangeLocal := makeRangeLocalKeyRange(d)
	rangeLockTable := makeRangeLockTableKeyRanges(d)
	user := MakeUserKeyRange(d)
	ranges := make([]KeyRange, 5)
	ranges[0] = rangeIDLocal
	ranges[1] = rangeLocal
	if len(rangeLockTable) != 2 {
		__antithesis_instrumentation__.Notify(114155)
		panic("unexpected number of lock table ranges")
	} else {
		__antithesis_instrumentation__.Notify(114156)
	}
	__antithesis_instrumentation__.Notify(114154)
	ranges[2] = rangeLockTable[0]
	ranges[3] = rangeLockTable[1]
	ranges[4] = user
	return ranges
}

func MakeReplicatedKeyRangesExceptLockTable(d *roachpb.RangeDescriptor) []KeyRange {
	__antithesis_instrumentation__.Notify(114157)
	return []KeyRange{
		MakeRangeIDLocalKeyRange(d.RangeID, true),
		makeRangeLocalKeyRange(d),
		MakeUserKeyRange(d),
	}
}

func MakeReplicatedKeyRangesExceptRangeID(d *roachpb.RangeDescriptor) []KeyRange {
	__antithesis_instrumentation__.Notify(114158)
	rangeLocal := makeRangeLocalKeyRange(d)
	rangeLockTable := makeRangeLockTableKeyRanges(d)
	user := MakeUserKeyRange(d)
	ranges := make([]KeyRange, 4)
	ranges[0] = rangeLocal
	if len(rangeLockTable) != 2 {
		__antithesis_instrumentation__.Notify(114160)
		panic("unexpected number of lock table ranges")
	} else {
		__antithesis_instrumentation__.Notify(114161)
	}
	__antithesis_instrumentation__.Notify(114159)
	ranges[1] = rangeLockTable[0]
	ranges[2] = rangeLockTable[1]
	ranges[3] = user
	return ranges
}

func MakeRangeIDLocalKeyRange(rangeID roachpb.RangeID, replicatedOnly bool) KeyRange {
	__antithesis_instrumentation__.Notify(114162)
	var prefixFn func(roachpb.RangeID) roachpb.Key
	if replicatedOnly {
		__antithesis_instrumentation__.Notify(114164)
		prefixFn = keys.MakeRangeIDReplicatedPrefix
	} else {
		__antithesis_instrumentation__.Notify(114165)
		prefixFn = keys.MakeRangeIDPrefix
	}
	__antithesis_instrumentation__.Notify(114163)
	sysRangeIDKey := prefixFn(rangeID)
	return KeyRange{
		Start: sysRangeIDKey,
		End:   sysRangeIDKey.PrefixEnd(),
	}
}

func makeRangeLocalKeyRange(d *roachpb.RangeDescriptor) KeyRange {
	__antithesis_instrumentation__.Notify(114166)
	return KeyRange{
		Start: keys.MakeRangeKeyPrefix(d.StartKey),
		End:   keys.MakeRangeKeyPrefix(d.EndKey),
	}
}

func makeRangeLockTableKeyRanges(d *roachpb.RangeDescriptor) [2]KeyRange {
	__antithesis_instrumentation__.Notify(114167)

	startRangeLocal, _ := keys.LockTableSingleKey(keys.MakeRangeKeyPrefix(d.StartKey), nil)
	endRangeLocal, _ := keys.LockTableSingleKey(keys.MakeRangeKeyPrefix(d.EndKey), nil)

	globalStartKey := d.StartKey.AsRawKey()
	if d.StartKey.Equal(roachpb.RKeyMin) {
		__antithesis_instrumentation__.Notify(114169)
		globalStartKey = keys.LocalMax
	} else {
		__antithesis_instrumentation__.Notify(114170)
	}
	__antithesis_instrumentation__.Notify(114168)
	startGlobal, _ := keys.LockTableSingleKey(globalStartKey, nil)
	endGlobal, _ := keys.LockTableSingleKey(roachpb.Key(d.EndKey), nil)
	return [2]KeyRange{
		{
			Start: startRangeLocal,
			End:   endRangeLocal,
		},
		{
			Start: startGlobal,
			End:   endGlobal,
		},
	}
}

func MakeUserKeyRange(d *roachpb.RangeDescriptor) KeyRange {
	__antithesis_instrumentation__.Notify(114171)
	userKeys := d.KeySpan()
	return KeyRange{
		Start: userKeys.Key.AsRawKey(),
		End:   userKeys.EndKey.AsRawKey(),
	}
}

func NewReplicaMVCCDataIterator(
	d *roachpb.RangeDescriptor, reader storage.Reader, seekEnd bool,
) *ReplicaMVCCDataIterator {
	__antithesis_instrumentation__.Notify(114172)
	if !reader.ConsistentIterators() {
		__antithesis_instrumentation__.Notify(114175)
		panic("ReplicaMVCCDataIterator needs a Reader that provides ConsistentIterators")
	} else {
		__antithesis_instrumentation__.Notify(114176)
	}
	__antithesis_instrumentation__.Notify(114173)
	ri := &ReplicaMVCCDataIterator{
		reader:  reader,
		ranges:  MakeReplicatedKeyRangesExceptLockTable(d),
		reverse: seekEnd,
	}
	if ri.reverse {
		__antithesis_instrumentation__.Notify(114177)
		ri.curIndex = len(ri.ranges) - 1
	} else {
		__antithesis_instrumentation__.Notify(114178)
		ri.curIndex = 0
	}
	__antithesis_instrumentation__.Notify(114174)
	ri.tryCloseAndCreateIter()
	return ri
}

func (ri *ReplicaMVCCDataIterator) tryCloseAndCreateIter() {
	__antithesis_instrumentation__.Notify(114179)
	for {
		__antithesis_instrumentation__.Notify(114180)
		if ri.it != nil {
			__antithesis_instrumentation__.Notify(114185)
			ri.it.Close()
			ri.it = nil
		} else {
			__antithesis_instrumentation__.Notify(114186)
		}
		__antithesis_instrumentation__.Notify(114181)
		if ri.curIndex < 0 || func() bool {
			__antithesis_instrumentation__.Notify(114187)
			return ri.curIndex >= len(ri.ranges) == true
		}() == true {
			__antithesis_instrumentation__.Notify(114188)
			return
		} else {
			__antithesis_instrumentation__.Notify(114189)
		}
		__antithesis_instrumentation__.Notify(114182)
		ri.it = ri.reader.NewMVCCIterator(
			storage.MVCCKeyAndIntentsIterKind,
			storage.IterOptions{
				LowerBound: ri.ranges[ri.curIndex].Start,
				UpperBound: ri.ranges[ri.curIndex].End,
			})
		if ri.reverse {
			__antithesis_instrumentation__.Notify(114190)
			ri.it.SeekLT(storage.MakeMVCCMetadataKey(ri.ranges[ri.curIndex].End))
		} else {
			__antithesis_instrumentation__.Notify(114191)
			ri.it.SeekGE(storage.MakeMVCCMetadataKey(ri.ranges[ri.curIndex].Start))
		}
		__antithesis_instrumentation__.Notify(114183)
		if valid, err := ri.it.Valid(); valid || func() bool {
			__antithesis_instrumentation__.Notify(114192)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(114193)
			ri.err = err
			return
		} else {
			__antithesis_instrumentation__.Notify(114194)
		}
		__antithesis_instrumentation__.Notify(114184)
		if ri.reverse {
			__antithesis_instrumentation__.Notify(114195)
			ri.curIndex--
		} else {
			__antithesis_instrumentation__.Notify(114196)
			ri.curIndex++
		}
	}
}

func (ri *ReplicaMVCCDataIterator) Close() {
	__antithesis_instrumentation__.Notify(114197)
	if ri.it != nil {
		__antithesis_instrumentation__.Notify(114198)
		ri.it.Close()
		ri.it = nil
	} else {
		__antithesis_instrumentation__.Notify(114199)
	}
}

func (ri *ReplicaMVCCDataIterator) Next() {
	__antithesis_instrumentation__.Notify(114200)
	if ri.reverse {
		__antithesis_instrumentation__.Notify(114203)
		panic("Next called on reverse iterator")
	} else {
		__antithesis_instrumentation__.Notify(114204)
	}
	__antithesis_instrumentation__.Notify(114201)
	ri.it.Next()
	valid, err := ri.it.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(114205)
		ri.err = err
		return
	} else {
		__antithesis_instrumentation__.Notify(114206)
	}
	__antithesis_instrumentation__.Notify(114202)
	if !valid {
		__antithesis_instrumentation__.Notify(114207)
		ri.curIndex++
		ri.tryCloseAndCreateIter()
	} else {
		__antithesis_instrumentation__.Notify(114208)
	}
}

func (ri *ReplicaMVCCDataIterator) Prev() {
	__antithesis_instrumentation__.Notify(114209)
	if !ri.reverse {
		__antithesis_instrumentation__.Notify(114212)
		panic("Prev called on forward iterator")
	} else {
		__antithesis_instrumentation__.Notify(114213)
	}
	__antithesis_instrumentation__.Notify(114210)
	ri.it.Prev()
	valid, err := ri.it.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(114214)
		ri.err = err
		return
	} else {
		__antithesis_instrumentation__.Notify(114215)
	}
	__antithesis_instrumentation__.Notify(114211)
	if !valid {
		__antithesis_instrumentation__.Notify(114216)
		ri.curIndex--
		ri.tryCloseAndCreateIter()
	} else {
		__antithesis_instrumentation__.Notify(114217)
	}
}

func (ri *ReplicaMVCCDataIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(114218)
	if ri.err != nil {
		__antithesis_instrumentation__.Notify(114221)
		return false, ri.err
	} else {
		__antithesis_instrumentation__.Notify(114222)
	}
	__antithesis_instrumentation__.Notify(114219)
	if ri.it == nil {
		__antithesis_instrumentation__.Notify(114223)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(114224)
	}
	__antithesis_instrumentation__.Notify(114220)
	return true, nil
}

func (ri *ReplicaMVCCDataIterator) Key() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(114225)
	return ri.it.Key()
}

func (ri *ReplicaMVCCDataIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(114226)
	return ri.it.Value()
}

func (ri *ReplicaMVCCDataIterator) UnsafeKey() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(114227)
	return ri.it.UnsafeKey()
}

func (ri *ReplicaMVCCDataIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(114228)
	return ri.it.UnsafeValue()
}

func NewReplicaEngineDataIterator(
	d *roachpb.RangeDescriptor, reader storage.Reader, replicatedOnly bool,
) *ReplicaEngineDataIterator {
	__antithesis_instrumentation__.Notify(114229)
	it := reader.NewEngineIterator(storage.IterOptions{UpperBound: d.EndKey.AsRawKey()})

	rangeFunc := MakeAllKeyRanges
	if replicatedOnly {
		__antithesis_instrumentation__.Notify(114231)
		rangeFunc = MakeReplicatedKeyRanges
	} else {
		__antithesis_instrumentation__.Notify(114232)
	}
	__antithesis_instrumentation__.Notify(114230)
	ri := &ReplicaEngineDataIterator{
		ranges: rangeFunc(d),
		it:     it,
	}
	ri.seekStart()
	return ri
}

func (ri *ReplicaEngineDataIterator) seekStart() {
	__antithesis_instrumentation__.Notify(114233)
	ri.curIndex = 0
	ri.valid, ri.err = ri.it.SeekEngineKeyGE(storage.EngineKey{Key: ri.ranges[ri.curIndex].Start})
	ri.advance()
}

func (ri *ReplicaEngineDataIterator) Close() {
	__antithesis_instrumentation__.Notify(114234)
	ri.valid = false
	ri.it.Close()
}

func (ri *ReplicaEngineDataIterator) Next() {
	__antithesis_instrumentation__.Notify(114235)
	ri.valid, ri.err = ri.it.NextEngineKey()
	ri.advance()
}

func (ri *ReplicaEngineDataIterator) advance() {
	__antithesis_instrumentation__.Notify(114236)
	for ri.valid {
		__antithesis_instrumentation__.Notify(114237)
		var k storage.EngineKey
		k, ri.err = ri.it.UnsafeEngineKey()
		if ri.err != nil {
			__antithesis_instrumentation__.Notify(114240)
			ri.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(114241)
		}
		__antithesis_instrumentation__.Notify(114238)
		if k.Key.Compare(ri.ranges[ri.curIndex].End) < 0 {
			__antithesis_instrumentation__.Notify(114242)
			return
		} else {
			__antithesis_instrumentation__.Notify(114243)
		}
		__antithesis_instrumentation__.Notify(114239)
		ri.curIndex++
		if ri.curIndex < len(ri.ranges) {
			__antithesis_instrumentation__.Notify(114244)
			ri.valid, ri.err = ri.it.SeekEngineKeyGE(
				storage.EngineKey{Key: ri.ranges[ri.curIndex].Start})
		} else {
			__antithesis_instrumentation__.Notify(114245)
			ri.valid = false
			return
		}
	}
}

func (ri *ReplicaEngineDataIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(114246)
	return ri.valid, ri.err
}

func (ri *ReplicaEngineDataIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(114247)
	value := ri.it.UnsafeValue()
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	return valueCopy
}

func (ri *ReplicaEngineDataIterator) UnsafeKey() storage.EngineKey {
	__antithesis_instrumentation__.Notify(114248)
	key, err := ri.it.UnsafeEngineKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(114250)

		panic("method called on an invalid iter")
	} else {
		__antithesis_instrumentation__.Notify(114251)
	}
	__antithesis_instrumentation__.Notify(114249)
	return key
}

func (ri *ReplicaEngineDataIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(114252)
	return ri.it.UnsafeValue()
}
