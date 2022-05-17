package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"math"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/sstable"
)

type pebbleIterator struct {
	iter    *pebble.Iterator
	options pebble.IterOptions

	keyBuf []byte

	lowerBoundBuf            [2][]byte
	upperBoundBuf            [2][]byte
	curBuf                   int
	testingSetBoundsListener testingSetBoundsListener

	prefix bool

	reusable bool
	inuse    bool

	mvccDirIsReverse bool

	mvccDone bool

	timeBoundNumSSTables int
}

var _ MVCCIterator = &pebbleIterator{}
var _ EngineIterator = &pebbleIterator{}

var pebbleIterPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(642845)
		return &pebbleIterator{}
	},
}

type cloneableIter interface {
	Clone() (*pebble.Iterator, error)
}

type testingSetBoundsListener interface {
	postSetBounds(lower, upper []byte)
}

func newPebbleIterator(
	handle pebble.Reader,
	iterToClone cloneableIter,
	opts IterOptions,
	durability DurabilityRequirement,
) *pebbleIterator {
	__antithesis_instrumentation__.Notify(642846)
	iter := pebbleIterPool.Get().(*pebbleIterator)
	iter.reusable = false
	iter.init(handle, iterToClone, opts, durability)
	return iter
}

func (p *pebbleIterator) init(
	handle pebble.Reader,
	iterToClone cloneableIter,
	opts IterOptions,
	durability DurabilityRequirement,
) {
	*p = pebbleIterator{
		keyBuf:        p.keyBuf,
		lowerBoundBuf: p.lowerBoundBuf,
		upperBoundBuf: p.upperBoundBuf,
		prefix:        opts.Prefix,
		reusable:      p.reusable,
	}

	if !opts.Prefix && len(opts.UpperBound) == 0 && len(opts.LowerBound) == 0 {
		panic("iterator must set prefix or upper bound or lower bound")
	}

	p.options.OnlyReadGuaranteedDurable = false
	if durability == GuaranteedDurability {
		p.options.OnlyReadGuaranteedDurable = true
	}
	if opts.LowerBound != nil {

		p.lowerBoundBuf[0] = append(p.lowerBoundBuf[0][:0], opts.LowerBound...)
		p.lowerBoundBuf[0] = append(p.lowerBoundBuf[0], 0x00)
		p.options.LowerBound = p.lowerBoundBuf[0]
	}
	if opts.UpperBound != nil {

		p.upperBoundBuf[0] = append(p.upperBoundBuf[0][:0], opts.UpperBound...)
		p.upperBoundBuf[0] = append(p.upperBoundBuf[0], 0x00)
		p.options.UpperBound = p.upperBoundBuf[0]
	}

	doClone := iterToClone != nil
	if !opts.MaxTimestampHint.IsEmpty() {
		doClone = false
		encodedMinTS := string(encodeMVCCTimestamp(opts.MinTimestampHint))
		encodedMaxTS := string(encodeMVCCTimestamp(opts.MaxTimestampHint))
		p.options.TableFilter = func(userProps map[string]string) bool {
			tableMinTS := userProps["crdb.ts.min"]
			if len(tableMinTS) == 0 {
				if opts.WithStats {
					p.timeBoundNumSSTables++
				}
				return true
			}
			tableMaxTS := userProps["crdb.ts.max"]
			if len(tableMaxTS) == 0 {
				if opts.WithStats {
					p.timeBoundNumSSTables++
				}
				return true
			}
			used := encodedMaxTS >= tableMinTS && encodedMinTS <= tableMaxTS
			if used && opts.WithStats {
				p.timeBoundNumSSTables++
			}
			return used
		}

		p.options.PointKeyFilters = []pebble.BlockPropertyFilter{
			sstable.NewBlockIntervalFilter(mvccWallTimeIntervalCollector,
				uint64(opts.MinTimestampHint.WallTime),
				uint64(opts.MaxTimestampHint.WallTime)+1),
		}
	} else if !opts.MinTimestampHint.IsEmpty() {
		panic("min timestamp hint set without max timestamp hint")
	}

	if doClone {
		var err error
		if p.iter, err = iterToClone.Clone(); err != nil {
			panic(err)
		}
		p.iter.SetBounds(p.options.LowerBound, p.options.UpperBound)
	} else {
		if handle == nil {
			panic("handle is nil for non-cloning path")
		}
		p.iter = handle.NewIter(&p.options)
	}
	if p.iter == nil {
		panic("unable to create iterator")
	}

	p.inuse = true
}

func (p *pebbleIterator) setBounds(lowerBound, upperBound roachpb.Key) {
	__antithesis_instrumentation__.Notify(642847)

	boundsChanged := ((lowerBound == nil) != (p.options.LowerBound == nil)) || func() bool {
		__antithesis_instrumentation__.Notify(642852)
		return ((upperBound == nil) != (p.options.UpperBound == nil)) == true
	}() == true
	if !boundsChanged {
		__antithesis_instrumentation__.Notify(642853)

		if lowerBound != nil {
			__antithesis_instrumentation__.Notify(642855)

			if !bytes.Equal(p.options.LowerBound[:len(p.options.LowerBound)-1], lowerBound) {
				__antithesis_instrumentation__.Notify(642856)
				boundsChanged = true
			} else {
				__antithesis_instrumentation__.Notify(642857)
			}
		} else {
			__antithesis_instrumentation__.Notify(642858)
		}
		__antithesis_instrumentation__.Notify(642854)

		if !boundsChanged && func() bool {
			__antithesis_instrumentation__.Notify(642859)
			return upperBound != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(642860)

			if !bytes.Equal(p.options.UpperBound[:len(p.options.UpperBound)-1], upperBound) {
				__antithesis_instrumentation__.Notify(642861)
				boundsChanged = true
			} else {
				__antithesis_instrumentation__.Notify(642862)
			}
		} else {
			__antithesis_instrumentation__.Notify(642863)
		}
	} else {
		__antithesis_instrumentation__.Notify(642864)
	}
	__antithesis_instrumentation__.Notify(642848)
	if !boundsChanged {
		__antithesis_instrumentation__.Notify(642865)

		return
	} else {
		__antithesis_instrumentation__.Notify(642866)
	}
	__antithesis_instrumentation__.Notify(642849)

	p.options.LowerBound = nil
	p.options.UpperBound = nil
	p.curBuf = (p.curBuf + 1) % 2
	i := p.curBuf
	if lowerBound != nil {
		__antithesis_instrumentation__.Notify(642867)

		p.lowerBoundBuf[i] = append(p.lowerBoundBuf[i][:0], lowerBound...)
		p.lowerBoundBuf[i] = append(p.lowerBoundBuf[i], 0x00)
		p.options.LowerBound = p.lowerBoundBuf[i]
	} else {
		__antithesis_instrumentation__.Notify(642868)
	}
	__antithesis_instrumentation__.Notify(642850)
	if upperBound != nil {
		__antithesis_instrumentation__.Notify(642869)

		p.upperBoundBuf[i] = append(p.upperBoundBuf[i][:0], upperBound...)
		p.upperBoundBuf[i] = append(p.upperBoundBuf[i], 0x00)
		p.options.UpperBound = p.upperBoundBuf[i]
	} else {
		__antithesis_instrumentation__.Notify(642870)
	}
	__antithesis_instrumentation__.Notify(642851)
	p.iter.SetBounds(p.options.LowerBound, p.options.UpperBound)
	if p.testingSetBoundsListener != nil {
		__antithesis_instrumentation__.Notify(642871)
		p.testingSetBoundsListener.postSetBounds(p.options.LowerBound, p.options.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(642872)
	}
}

func (p *pebbleIterator) Close() {
	__antithesis_instrumentation__.Notify(642873)
	if !p.inuse {
		__antithesis_instrumentation__.Notify(642876)
		panic("closing idle iterator")
	} else {
		__antithesis_instrumentation__.Notify(642877)
	}
	__antithesis_instrumentation__.Notify(642874)
	p.inuse = false

	if p.reusable {
		__antithesis_instrumentation__.Notify(642878)
		p.iter.ResetStats()
		return
	} else {
		__antithesis_instrumentation__.Notify(642879)
	}
	__antithesis_instrumentation__.Notify(642875)

	p.destroy()

	pebbleIterPool.Put(p)
}

func (p *pebbleIterator) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(642880)
	p.mvccDirIsReverse = false
	p.mvccDone = false
	p.keyBuf = EncodeMVCCKeyToBuf(p.keyBuf[:0], key)
	if p.prefix {
		__antithesis_instrumentation__.Notify(642881)
		p.iter.SeekPrefixGE(p.keyBuf)
	} else {
		__antithesis_instrumentation__.Notify(642882)
		p.iter.SeekGE(p.keyBuf)
	}
}

func (p *pebbleIterator) SeekIntentGE(key roachpb.Key, _ uuid.UUID) {
	__antithesis_instrumentation__.Notify(642883)
	p.SeekGE(MVCCKey{Key: key})
}

func (p *pebbleIterator) SeekEngineKeyGE(key EngineKey) (valid bool, err error) {
	__antithesis_instrumentation__.Notify(642884)
	p.keyBuf = key.EncodeToBuf(p.keyBuf[:0])
	var ok bool
	if p.prefix {
		__antithesis_instrumentation__.Notify(642887)
		ok = p.iter.SeekPrefixGE(p.keyBuf)
	} else {
		__antithesis_instrumentation__.Notify(642888)
		ok = p.iter.SeekGE(p.keyBuf)
	}
	__antithesis_instrumentation__.Notify(642885)

	if ok {
		__antithesis_instrumentation__.Notify(642889)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(642890)
	}
	__antithesis_instrumentation__.Notify(642886)
	return false, p.iter.Error()
}

func (p *pebbleIterator) SeekEngineKeyGEWithLimit(
	key EngineKey, limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(642891)
	p.keyBuf = key.EncodeToBuf(p.keyBuf[:0])
	if limit != nil {
		__antithesis_instrumentation__.Notify(642895)
		if p.prefix {
			__antithesis_instrumentation__.Notify(642897)
			panic("prefix iteration does not permit a limit")
		} else {
			__antithesis_instrumentation__.Notify(642898)
		}
		__antithesis_instrumentation__.Notify(642896)

		limit = append(limit, '\x00')
	} else {
		__antithesis_instrumentation__.Notify(642899)
	}
	__antithesis_instrumentation__.Notify(642892)
	if p.prefix {
		__antithesis_instrumentation__.Notify(642900)
		state = pebble.IterExhausted
		if p.iter.SeekPrefixGE(p.keyBuf) {
			__antithesis_instrumentation__.Notify(642901)
			state = pebble.IterValid
		} else {
			__antithesis_instrumentation__.Notify(642902)
		}
	} else {
		__antithesis_instrumentation__.Notify(642903)
		state = p.iter.SeekGEWithLimit(p.keyBuf, limit)
	}
	__antithesis_instrumentation__.Notify(642893)
	if state == pebble.IterExhausted {
		__antithesis_instrumentation__.Notify(642904)
		return state, p.iter.Error()
	} else {
		__antithesis_instrumentation__.Notify(642905)
	}
	__antithesis_instrumentation__.Notify(642894)
	return state, nil
}

func (p *pebbleIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(642906)
	if p.mvccDone {
		__antithesis_instrumentation__.Notify(642909)
		return false, nil
	} else {
		__antithesis_instrumentation__.Notify(642910)
	}
	__antithesis_instrumentation__.Notify(642907)

	if ok := p.iter.Valid(); ok {
		__antithesis_instrumentation__.Notify(642911)

		k := p.iter.Key()
		if len(k) == 0 {
			__antithesis_instrumentation__.Notify(642914)
			return false, errors.Errorf("iterator encountered 0 length key")
		} else {
			__antithesis_instrumentation__.Notify(642915)
		}
		__antithesis_instrumentation__.Notify(642912)

		versionLen := int(k[len(k)-1])
		if versionLen == engineKeyVersionLockTableLen+1 {
			__antithesis_instrumentation__.Notify(642916)
			p.mvccDone = true
			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(642917)
		}
		__antithesis_instrumentation__.Notify(642913)
		return ok, nil
	} else {
		__antithesis_instrumentation__.Notify(642918)
	}
	__antithesis_instrumentation__.Notify(642908)
	return false, p.iter.Error()
}

func (p *pebbleIterator) Next() {
	__antithesis_instrumentation__.Notify(642919)
	if p.mvccDirIsReverse {
		__antithesis_instrumentation__.Notify(642922)

		p.mvccDirIsReverse = false
		p.mvccDone = false
	} else {
		__antithesis_instrumentation__.Notify(642923)
	}
	__antithesis_instrumentation__.Notify(642920)
	if p.mvccDone {
		__antithesis_instrumentation__.Notify(642924)
		return
	} else {
		__antithesis_instrumentation__.Notify(642925)
	}
	__antithesis_instrumentation__.Notify(642921)
	p.iter.Next()
}

func (p *pebbleIterator) NextEngineKey() (valid bool, err error) {
	__antithesis_instrumentation__.Notify(642926)
	ok := p.iter.Next()

	if ok {
		__antithesis_instrumentation__.Notify(642928)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(642929)
	}
	__antithesis_instrumentation__.Notify(642927)
	return false, p.iter.Error()
}

func (p *pebbleIterator) NextEngineKeyWithLimit(
	limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(642930)
	if limit != nil {
		__antithesis_instrumentation__.Notify(642933)

		limit = append(limit, '\x00')
	} else {
		__antithesis_instrumentation__.Notify(642934)
	}
	__antithesis_instrumentation__.Notify(642931)
	state = p.iter.NextWithLimit(limit)
	if state == pebble.IterExhausted {
		__antithesis_instrumentation__.Notify(642935)
		return state, p.iter.Error()
	} else {
		__antithesis_instrumentation__.Notify(642936)
	}
	__antithesis_instrumentation__.Notify(642932)
	return state, nil
}

func (p *pebbleIterator) NextKey() {
	__antithesis_instrumentation__.Notify(642937)

	if p.mvccDirIsReverse {
		__antithesis_instrumentation__.Notify(642942)

		p.mvccDirIsReverse = false
		p.mvccDone = false
	} else {
		__antithesis_instrumentation__.Notify(642943)
	}
	__antithesis_instrumentation__.Notify(642938)
	if p.mvccDone {
		__antithesis_instrumentation__.Notify(642944)
		return
	} else {
		__antithesis_instrumentation__.Notify(642945)
	}
	__antithesis_instrumentation__.Notify(642939)
	if valid, err := p.Valid(); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(642946)
		return !valid == true
	}() == true {
		__antithesis_instrumentation__.Notify(642947)
		return
	} else {
		__antithesis_instrumentation__.Notify(642948)
	}
	__antithesis_instrumentation__.Notify(642940)
	p.keyBuf = append(p.keyBuf[:0], p.UnsafeKey().Key...)
	if !p.iter.Next() {
		__antithesis_instrumentation__.Notify(642949)
		return
	} else {
		__antithesis_instrumentation__.Notify(642950)
	}
	__antithesis_instrumentation__.Notify(642941)
	if bytes.Equal(p.keyBuf, p.UnsafeKey().Key) {
		__antithesis_instrumentation__.Notify(642951)

		p.iter.SeekGE(append(p.keyBuf, 0, 0))
	} else {
		__antithesis_instrumentation__.Notify(642952)
	}
}

func (p *pebbleIterator) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(642953)
	if valid, err := p.Valid(); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(642956)
		return !valid == true
	}() == true {
		__antithesis_instrumentation__.Notify(642957)
		return MVCCKey{}
	} else {
		__antithesis_instrumentation__.Notify(642958)
	}
	__antithesis_instrumentation__.Notify(642954)

	mvccKey, err := DecodeMVCCKey(p.iter.Key())
	if err != nil {
		__antithesis_instrumentation__.Notify(642959)
		return MVCCKey{}
	} else {
		__antithesis_instrumentation__.Notify(642960)
	}
	__antithesis_instrumentation__.Notify(642955)

	return mvccKey
}

func (p *pebbleIterator) UnsafeEngineKey() (EngineKey, error) {
	__antithesis_instrumentation__.Notify(642961)
	engineKey, ok := DecodeEngineKey(p.iter.Key())
	if !ok {
		__antithesis_instrumentation__.Notify(642963)
		return engineKey, errors.Errorf("invalid encoded engine key: %x", p.iter.Key())
	} else {
		__antithesis_instrumentation__.Notify(642964)
	}
	__antithesis_instrumentation__.Notify(642962)
	return engineKey, nil
}

func (p *pebbleIterator) UnsafeRawKey() []byte {
	__antithesis_instrumentation__.Notify(642965)
	return p.iter.Key()
}

func (p *pebbleIterator) UnsafeRawMVCCKey() []byte {
	__antithesis_instrumentation__.Notify(642966)
	return p.iter.Key()
}

func (p *pebbleIterator) UnsafeRawEngineKey() []byte {
	__antithesis_instrumentation__.Notify(642967)
	return p.iter.Key()
}

func (p *pebbleIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(642968)
	if ok := p.iter.Valid(); !ok {
		__antithesis_instrumentation__.Notify(642970)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(642971)
	}
	__antithesis_instrumentation__.Notify(642969)
	return p.iter.Value()
}

func (p *pebbleIterator) SeekLT(key MVCCKey) {
	__antithesis_instrumentation__.Notify(642972)
	p.mvccDirIsReverse = true
	p.mvccDone = false
	p.keyBuf = EncodeMVCCKeyToBuf(p.keyBuf[:0], key)
	p.iter.SeekLT(p.keyBuf)
}

func (p *pebbleIterator) SeekEngineKeyLT(key EngineKey) (valid bool, err error) {
	__antithesis_instrumentation__.Notify(642973)
	p.keyBuf = key.EncodeToBuf(p.keyBuf[:0])
	ok := p.iter.SeekLT(p.keyBuf)

	if ok {
		__antithesis_instrumentation__.Notify(642975)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(642976)
	}
	__antithesis_instrumentation__.Notify(642974)
	return false, p.iter.Error()
}

func (p *pebbleIterator) SeekEngineKeyLTWithLimit(
	key EngineKey, limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(642977)
	p.keyBuf = key.EncodeToBuf(p.keyBuf[:0])
	if limit != nil {
		__antithesis_instrumentation__.Notify(642980)

		limit = append(limit, '\x00')
	} else {
		__antithesis_instrumentation__.Notify(642981)
	}
	__antithesis_instrumentation__.Notify(642978)
	state = p.iter.SeekLTWithLimit(p.keyBuf, limit)
	if state == pebble.IterExhausted {
		__antithesis_instrumentation__.Notify(642982)
		return state, p.iter.Error()
	} else {
		__antithesis_instrumentation__.Notify(642983)
	}
	__antithesis_instrumentation__.Notify(642979)
	return state, nil
}

func (p *pebbleIterator) Prev() {
	__antithesis_instrumentation__.Notify(642984)
	if !p.mvccDirIsReverse {
		__antithesis_instrumentation__.Notify(642987)

		p.mvccDirIsReverse = true
		p.mvccDone = false
	} else {
		__antithesis_instrumentation__.Notify(642988)
	}
	__antithesis_instrumentation__.Notify(642985)
	if p.mvccDone {
		__antithesis_instrumentation__.Notify(642989)
		return
	} else {
		__antithesis_instrumentation__.Notify(642990)
	}
	__antithesis_instrumentation__.Notify(642986)
	p.iter.Prev()
}

func (p *pebbleIterator) PrevEngineKey() (valid bool, err error) {
	__antithesis_instrumentation__.Notify(642991)
	ok := p.iter.Prev()

	if ok {
		__antithesis_instrumentation__.Notify(642993)
		return true, nil
	} else {
		__antithesis_instrumentation__.Notify(642994)
	}
	__antithesis_instrumentation__.Notify(642992)
	return false, p.iter.Error()
}

func (p *pebbleIterator) PrevEngineKeyWithLimit(
	limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(642995)
	if limit != nil {
		__antithesis_instrumentation__.Notify(642998)

		limit = append(limit, '\x00')
	} else {
		__antithesis_instrumentation__.Notify(642999)
	}
	__antithesis_instrumentation__.Notify(642996)
	state = p.iter.PrevWithLimit(limit)
	if state == pebble.IterExhausted {
		__antithesis_instrumentation__.Notify(643000)
		return state, p.iter.Error()
	} else {
		__antithesis_instrumentation__.Notify(643001)
	}
	__antithesis_instrumentation__.Notify(642997)
	return state, nil
}

func (p *pebbleIterator) Key() MVCCKey {
	__antithesis_instrumentation__.Notify(643002)
	key := p.UnsafeKey()
	keyCopy := make([]byte, len(key.Key))
	copy(keyCopy, key.Key)
	key.Key = keyCopy
	return key
}

func (p *pebbleIterator) EngineKey() (EngineKey, error) {
	__antithesis_instrumentation__.Notify(643003)
	key, err := p.UnsafeEngineKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(643005)
		return key, err
	} else {
		__antithesis_instrumentation__.Notify(643006)
	}
	__antithesis_instrumentation__.Notify(643004)
	return key.Copy(), nil
}

func (p *pebbleIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(643007)
	value := p.UnsafeValue()
	valueCopy := make([]byte, len(value))
	copy(valueCopy, value)
	return valueCopy
}

func (p *pebbleIterator) ValueProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(643008)
	value := p.UnsafeValue()

	return protoutil.Unmarshal(value, msg)
}

func (p *pebbleIterator) ComputeStats(
	start, end roachpb.Key, nowNanos int64,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(643009)
	return ComputeStatsForRange(p, start, end, nowNanos)
}

func isValidSplitKey(key roachpb.Key, noSplitSpans []roachpb.Span) bool {
	__antithesis_instrumentation__.Notify(643010)
	if key.Equal(keys.Meta2KeyMax) {
		__antithesis_instrumentation__.Notify(643013)

		return false
	} else {
		__antithesis_instrumentation__.Notify(643014)
	}
	__antithesis_instrumentation__.Notify(643011)
	for i := range noSplitSpans {
		__antithesis_instrumentation__.Notify(643015)
		if noSplitSpans[i].ProperlyContainsKey(key) {
			__antithesis_instrumentation__.Notify(643016)
			return false
		} else {
			__antithesis_instrumentation__.Notify(643017)
		}
	}
	__antithesis_instrumentation__.Notify(643012)
	return true
}

func IsValidSplitKey(key roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(643018)
	return isValidSplitKey(key, keys.NoSplitSpans)
}

func (p *pebbleIterator) FindSplitKey(
	start, end, minSplitKey roachpb.Key, targetSize int64,
) (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(643019)
	return findSplitKeyUsingIterator(p, start, end, minSplitKey, targetSize)
}

func findSplitKeyUsingIterator(
	iter MVCCIterator, start, end, minSplitKey roachpb.Key, targetSize int64,
) (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(643020)
	const timestampLen = 12

	sizeSoFar := int64(0)
	bestDiff := int64(math.MaxInt64)
	bestSplitKey := MVCCKey{}

	found := false
	prevKey := MVCCKey{}

	noSplitSpans := keys.NoSplitSpans
	for i := range noSplitSpans {
		__antithesis_instrumentation__.Notify(643024)
		if minSplitKey.Compare(noSplitSpans[i].EndKey) <= 0 {
			__antithesis_instrumentation__.Notify(643025)
			noSplitSpans = noSplitSpans[i:]
			break
		} else {
			__antithesis_instrumentation__.Notify(643026)
		}
	}
	__antithesis_instrumentation__.Notify(643021)

	mvccMinSplitKey := MakeMVCCMetadataKey(minSplitKey)
	iter.SeekGE(MakeMVCCMetadataKey(start))
	for ; ; iter.Next() {
		__antithesis_instrumentation__.Notify(643027)
		valid, err := iter.Valid()
		if err != nil {
			__antithesis_instrumentation__.Notify(643035)
			return MVCCKey{}, err
		} else {
			__antithesis_instrumentation__.Notify(643036)
		}
		__antithesis_instrumentation__.Notify(643028)
		if !valid {
			__antithesis_instrumentation__.Notify(643037)
			break
		} else {
			__antithesis_instrumentation__.Notify(643038)
		}
		__antithesis_instrumentation__.Notify(643029)
		mvccKey := iter.UnsafeKey()

		diff := targetSize - sizeSoFar
		if diff < 0 {
			__antithesis_instrumentation__.Notify(643039)
			diff = -diff
		} else {
			__antithesis_instrumentation__.Notify(643040)
		}
		__antithesis_instrumentation__.Notify(643030)
		if diff > bestDiff {
			__antithesis_instrumentation__.Notify(643041)

			break
		} else {
			__antithesis_instrumentation__.Notify(643042)
		}
		__antithesis_instrumentation__.Notify(643031)

		if mvccMinSplitKey.Key != nil && func() bool {
			__antithesis_instrumentation__.Notify(643043)
			return !mvccKey.Less(mvccMinSplitKey) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643044)

			mvccMinSplitKey.Key = nil
		} else {
			__antithesis_instrumentation__.Notify(643045)
		}
		__antithesis_instrumentation__.Notify(643032)

		if mvccMinSplitKey.Key == nil && func() bool {
			__antithesis_instrumentation__.Notify(643046)
			return diff < bestDiff == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(643047)
			return (len(noSplitSpans) == 0 || func() bool {
				__antithesis_instrumentation__.Notify(643048)
				return isValidSplitKey(mvccKey.Key, noSplitSpans) == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643049)

			bestDiff = diff
			found = true

			bestSplitKey.Key = bestSplitKey.Key[:0]
		} else {
			__antithesis_instrumentation__.Notify(643050)
			if found && func() bool {
				__antithesis_instrumentation__.Notify(643051)
				return len(bestSplitKey.Key) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(643052)

				bestSplitKey.Timestamp = prevKey.Timestamp
				bestSplitKey.Key = append(bestSplitKey.Key[:0], prevKey.Key...)
			} else {
				__antithesis_instrumentation__.Notify(643053)
			}
		}
		__antithesis_instrumentation__.Notify(643033)

		sizeSoFar += int64(len(iter.UnsafeValue()))
		if mvccKey.IsValue() && func() bool {
			__antithesis_instrumentation__.Notify(643054)
			return bytes.Equal(prevKey.Key, mvccKey.Key) == true
		}() == true {
			__antithesis_instrumentation__.Notify(643055)

			sizeSoFar += timestampLen
		} else {
			__antithesis_instrumentation__.Notify(643056)
			sizeSoFar += int64(len(mvccKey.Key) + 1)
			if mvccKey.IsValue() {
				__antithesis_instrumentation__.Notify(643057)
				sizeSoFar += timestampLen
			} else {
				__antithesis_instrumentation__.Notify(643058)
			}
		}
		__antithesis_instrumentation__.Notify(643034)

		prevKey.Key = append(prevKey.Key[:0], mvccKey.Key...)
		prevKey.Timestamp = mvccKey.Timestamp
	}
	__antithesis_instrumentation__.Notify(643022)

	if found && func() bool {
		__antithesis_instrumentation__.Notify(643059)
		return len(bestSplitKey.Key) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(643060)

		return prevKey, nil
	} else {
		__antithesis_instrumentation__.Notify(643061)
	}
	__antithesis_instrumentation__.Notify(643023)
	return bestSplitKey, nil
}

func (p *pebbleIterator) SetUpperBound(upperBound roachpb.Key) {
	__antithesis_instrumentation__.Notify(643062)
	if upperBound == nil {
		__antithesis_instrumentation__.Notify(643066)
		panic("SetUpperBound must not use a nil key")
	} else {
		__antithesis_instrumentation__.Notify(643067)
	}
	__antithesis_instrumentation__.Notify(643063)
	if p.options.UpperBound != nil {
		__antithesis_instrumentation__.Notify(643068)

		if bytes.Equal(p.options.UpperBound[:len(p.options.UpperBound)-1], upperBound) {
			__antithesis_instrumentation__.Notify(643069)

			return
		} else {
			__antithesis_instrumentation__.Notify(643070)
		}
	} else {
		__antithesis_instrumentation__.Notify(643071)
	}
	__antithesis_instrumentation__.Notify(643064)
	p.curBuf = (p.curBuf + 1) % 2
	i := p.curBuf
	if p.options.LowerBound != nil {
		__antithesis_instrumentation__.Notify(643072)
		p.lowerBoundBuf[i] = append(p.lowerBoundBuf[i][:0], p.options.LowerBound...)
		p.options.LowerBound = p.lowerBoundBuf[i]
	} else {
		__antithesis_instrumentation__.Notify(643073)
	}
	__antithesis_instrumentation__.Notify(643065)
	p.upperBoundBuf[i] = append(p.upperBoundBuf[i][:0], upperBound...)
	p.upperBoundBuf[i] = append(p.upperBoundBuf[i], 0x00)
	p.options.UpperBound = p.upperBoundBuf[i]
	p.iter.SetBounds(p.options.LowerBound, p.options.UpperBound)
	if p.testingSetBoundsListener != nil {
		__antithesis_instrumentation__.Notify(643074)
		p.testingSetBoundsListener.postSetBounds(p.options.LowerBound, p.options.UpperBound)
	} else {
		__antithesis_instrumentation__.Notify(643075)
	}
}

func (p *pebbleIterator) Stats() IteratorStats {
	__antithesis_instrumentation__.Notify(643076)
	return IteratorStats{
		TimeBoundNumSSTs: p.timeBoundNumSSTables,
		Stats:            p.iter.Stats(),
	}
}

func (p *pebbleIterator) SupportsPrev() bool {
	__antithesis_instrumentation__.Notify(643077)
	return true
}

func (p *pebbleIterator) GetRawIter() *pebble.Iterator {
	__antithesis_instrumentation__.Notify(643078)
	return p.iter
}

func (p *pebbleIterator) destroy() {
	__antithesis_instrumentation__.Notify(643079)
	if p.inuse {
		__antithesis_instrumentation__.Notify(643082)
		panic("iterator still in use")
	} else {
		__antithesis_instrumentation__.Notify(643083)
	}
	__antithesis_instrumentation__.Notify(643080)
	if p.iter != nil {
		__antithesis_instrumentation__.Notify(643084)
		err := p.iter.Close()
		if err != nil {
			__antithesis_instrumentation__.Notify(643086)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(643087)
		}
		__antithesis_instrumentation__.Notify(643085)
		p.iter = nil
	} else {
		__antithesis_instrumentation__.Notify(643088)
	}
	__antithesis_instrumentation__.Notify(643081)

	*p = pebbleIterator{
		keyBuf:        p.keyBuf,
		lowerBoundBuf: p.lowerBoundBuf,
		upperBoundBuf: p.upperBoundBuf,
		reusable:      p.reusable,
	}
}
