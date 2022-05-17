package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"math/rand"
	"sync"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/kv/kvserver/concurrency/lock"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble"
)

type intentInterleavingIterConstraint int8

const (
	notConstrained intentInterleavingIterConstraint = iota
	constrainedToLocal
	constrainedToGlobal
)

type intentInterleavingIter struct {
	prefix     bool
	constraint intentInterleavingIterConstraint

	iter MVCCIterator

	iterValid bool

	iterKey MVCCKey

	intentIter      EngineIterator
	intentIterState pebble.IterValidityState

	intentKey                            roachpb.Key
	intentKeyAsNoTimestampMVCCKey        []byte
	intentKeyAsNoTimestampMVCCKeyBacking []byte

	intentCmp int

	dir   int
	valid bool
	err   error

	intentKeyBuf      []byte
	intentLimitKeyBuf []byte
}

var _ MVCCIterator = &intentInterleavingIter{}

var intentInterleavingIterPool = sync.Pool{
	New: func() interface{} {
		__antithesis_instrumentation__.Notify(639196)
		return &intentInterleavingIter{}
	},
}

func isLocal(k roachpb.Key) bool {
	__antithesis_instrumentation__.Notify(639197)
	return len(k) == 0 || func() bool {
		__antithesis_instrumentation__.Notify(639198)
		return keys.IsLocal(k) == true
	}() == true
}

func newIntentInterleavingIterator(reader Reader, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(639199)
	if !opts.MinTimestampHint.IsEmpty() || func() bool {
		__antithesis_instrumentation__.Notify(639207)
		return !opts.MaxTimestampHint.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(639208)
		panic("intentInterleavingIter must not be used with timestamp hints")
	} else {
		__antithesis_instrumentation__.Notify(639209)
	}
	__antithesis_instrumentation__.Notify(639200)
	var lowerIsLocal, upperIsLocal bool
	var constraint intentInterleavingIterConstraint
	if opts.LowerBound != nil {
		__antithesis_instrumentation__.Notify(639210)
		lowerIsLocal = isLocal(opts.LowerBound)
		if lowerIsLocal {
			__antithesis_instrumentation__.Notify(639211)
			constraint = constrainedToLocal
		} else {
			__antithesis_instrumentation__.Notify(639212)
			constraint = constrainedToGlobal
		}
	} else {
		__antithesis_instrumentation__.Notify(639213)
	}
	__antithesis_instrumentation__.Notify(639201)
	if opts.UpperBound != nil {
		__antithesis_instrumentation__.Notify(639214)
		upperIsLocal = isLocal(opts.UpperBound) || func() bool {
			__antithesis_instrumentation__.Notify(639216)
			return bytes.Equal(opts.UpperBound, keys.LocalMax) == true
		}() == true
		if opts.LowerBound != nil && func() bool {
			__antithesis_instrumentation__.Notify(639217)
			return lowerIsLocal != upperIsLocal == true
		}() == true {
			__antithesis_instrumentation__.Notify(639218)
			panic(fmt.Sprintf(
				"intentInterleavingIter cannot span from lowerIsLocal %t, %s to upperIsLocal %t, %s",
				lowerIsLocal, opts.LowerBound.String(), upperIsLocal, opts.UpperBound.String()))
		} else {
			__antithesis_instrumentation__.Notify(639219)
		}
		__antithesis_instrumentation__.Notify(639215)
		if upperIsLocal {
			__antithesis_instrumentation__.Notify(639220)
			constraint = constrainedToLocal
		} else {
			__antithesis_instrumentation__.Notify(639221)
			constraint = constrainedToGlobal
		}
	} else {
		__antithesis_instrumentation__.Notify(639222)
	}
	__antithesis_instrumentation__.Notify(639202)
	if !opts.Prefix {
		__antithesis_instrumentation__.Notify(639223)
		if opts.LowerBound == nil && func() bool {
			__antithesis_instrumentation__.Notify(639226)
			return opts.UpperBound == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(639227)

			panic("iterator must set prefix or upper bound or lower bound")
		} else {
			__antithesis_instrumentation__.Notify(639228)
		}
		__antithesis_instrumentation__.Notify(639224)

		if opts.LowerBound == nil && func() bool {
			__antithesis_instrumentation__.Notify(639229)
			return constraint == constrainedToGlobal == true
		}() == true {
			__antithesis_instrumentation__.Notify(639230)

			opts.LowerBound = keys.LocalMax
		} else {
			__antithesis_instrumentation__.Notify(639231)
		}
		__antithesis_instrumentation__.Notify(639225)
		if opts.UpperBound == nil && func() bool {
			__antithesis_instrumentation__.Notify(639232)
			return constraint == constrainedToLocal == true
		}() == true {
			__antithesis_instrumentation__.Notify(639233)

			opts.UpperBound = keys.LocalRangeLockTablePrefix
		} else {
			__antithesis_instrumentation__.Notify(639234)
		}
	} else {
		__antithesis_instrumentation__.Notify(639235)
	}
	__antithesis_instrumentation__.Notify(639203)

	iiIter := intentInterleavingIterPool.Get().(*intentInterleavingIter)
	intentOpts := opts
	intentKeyBuf := iiIter.intentKeyBuf
	intentLimitKeyBuf := iiIter.intentLimitKeyBuf
	if opts.LowerBound != nil {
		__antithesis_instrumentation__.Notify(639236)
		intentOpts.LowerBound, intentKeyBuf = keys.LockTableSingleKey(opts.LowerBound, intentKeyBuf)
	} else {
		__antithesis_instrumentation__.Notify(639237)
		if !opts.Prefix {
			__antithesis_instrumentation__.Notify(639238)

			intentOpts.LowerBound = keys.LockTableSingleKeyStart
		} else {
			__antithesis_instrumentation__.Notify(639239)
		}
	}
	__antithesis_instrumentation__.Notify(639204)
	if opts.UpperBound != nil {
		__antithesis_instrumentation__.Notify(639240)
		intentOpts.UpperBound, intentLimitKeyBuf =
			keys.LockTableSingleKey(opts.UpperBound, intentLimitKeyBuf)
	} else {
		__antithesis_instrumentation__.Notify(639241)
		if !opts.Prefix {
			__antithesis_instrumentation__.Notify(639242)

			intentOpts.UpperBound = keys.LockTableSingleKeyEnd
		} else {
			__antithesis_instrumentation__.Notify(639243)
		}
	}
	__antithesis_instrumentation__.Notify(639205)

	intentIter := reader.NewEngineIterator(intentOpts)

	var iter MVCCIterator
	if reader.ConsistentIterators() {
		__antithesis_instrumentation__.Notify(639244)
		iter = reader.NewMVCCIterator(MVCCKeyIterKind, opts)
	} else {
		__antithesis_instrumentation__.Notify(639245)
		iter = newMVCCIteratorByCloningEngineIter(intentIter, opts)
	}
	__antithesis_instrumentation__.Notify(639206)

	*iiIter = intentInterleavingIter{
		prefix:                               opts.Prefix,
		constraint:                           constraint,
		iter:                                 iter,
		intentIter:                           intentIter,
		intentKeyAsNoTimestampMVCCKeyBacking: iiIter.intentKeyAsNoTimestampMVCCKeyBacking,
		intentKeyBuf:                         intentKeyBuf,
		intentLimitKeyBuf:                    intentLimitKeyBuf,
	}
	return iiIter
}

func (i *intentInterleavingIter) makeUpperLimitKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(639246)
	key := i.iterKey.Key

	keyLen :=
		len(keys.LocalRangeLockTablePrefix) + len(keys.LockTableSingleKeyInfix) + len(key) + 3 + 2
	if cap(i.intentLimitKeyBuf) < keyLen {
		__antithesis_instrumentation__.Notify(639248)
		i.intentLimitKeyBuf = make([]byte, 0, keyLen)
	} else {
		__antithesis_instrumentation__.Notify(639249)
	}
	__antithesis_instrumentation__.Notify(639247)
	_, i.intentLimitKeyBuf = keys.LockTableSingleKey(key, i.intentLimitKeyBuf)

	i.intentLimitKeyBuf = append(i.intentLimitKeyBuf, '\x00')
	return i.intentLimitKeyBuf
}

func (i *intentInterleavingIter) makeLowerLimitKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(639250)
	key := i.iterKey.Key

	keyLen :=
		len(keys.LocalRangeLockTablePrefix) + len(keys.LockTableSingleKeyInfix) + len(key) + 3 + 1
	if cap(i.intentLimitKeyBuf) < keyLen {
		__antithesis_instrumentation__.Notify(639252)
		i.intentLimitKeyBuf = make([]byte, 0, keyLen)
	} else {
		__antithesis_instrumentation__.Notify(639253)
	}
	__antithesis_instrumentation__.Notify(639251)
	_, i.intentLimitKeyBuf = keys.LockTableSingleKey(key, i.intentLimitKeyBuf)
	return i.intentLimitKeyBuf
}

func (i *intentInterleavingIter) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(639254)
	i.dir = +1
	i.valid = true
	i.err = nil

	if i.constraint != notConstrained {
		__antithesis_instrumentation__.Notify(639259)
		i.checkConstraint(key.Key, false)
	} else {
		__antithesis_instrumentation__.Notify(639260)
	}
	__antithesis_instrumentation__.Notify(639255)
	i.iter.SeekGE(key)
	if err := i.tryDecodeKey(); err != nil {
		__antithesis_instrumentation__.Notify(639261)
		return
	} else {
		__antithesis_instrumentation__.Notify(639262)
	}
	__antithesis_instrumentation__.Notify(639256)
	var intentSeekKey roachpb.Key
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(639263)

		intentSeekKey, i.intentKeyBuf = keys.LockTableSingleKey(key.Key, i.intentKeyBuf)
	} else {
		__antithesis_instrumentation__.Notify(639264)
		if !i.prefix {
			__antithesis_instrumentation__.Notify(639265)

			intentSeekKey, i.intentKeyBuf = keys.LockTableSingleKey(key.Key.Next(), i.intentKeyBuf)
		} else {
			__antithesis_instrumentation__.Notify(639266)

			i.intentKey = nil
		}
	}
	__antithesis_instrumentation__.Notify(639257)
	if intentSeekKey != nil {
		__antithesis_instrumentation__.Notify(639267)
		var limitKey roachpb.Key
		if i.iterValid && func() bool {
			__antithesis_instrumentation__.Notify(639269)
			return !i.prefix == true
		}() == true {
			__antithesis_instrumentation__.Notify(639270)
			limitKey = i.makeUpperLimitKey()
		} else {
			__antithesis_instrumentation__.Notify(639271)
		}
		__antithesis_instrumentation__.Notify(639268)
		iterState, err := i.intentIter.SeekEngineKeyGEWithLimit(EngineKey{Key: intentSeekKey}, limitKey)
		if err = i.tryDecodeLockKey(iterState, err); err != nil {
			__antithesis_instrumentation__.Notify(639272)
			return
		} else {
			__antithesis_instrumentation__.Notify(639273)
		}
	} else {
		__antithesis_instrumentation__.Notify(639274)
	}
	__antithesis_instrumentation__.Notify(639258)
	i.computePos()
}

func (i *intentInterleavingIter) SeekIntentGE(key roachpb.Key, txnUUID uuid.UUID) {
	__antithesis_instrumentation__.Notify(639275)
	i.dir = +1
	i.valid = true

	if i.constraint != notConstrained {
		__antithesis_instrumentation__.Notify(639280)
		i.checkConstraint(key, false)
	} else {
		__antithesis_instrumentation__.Notify(639281)
	}
	__antithesis_instrumentation__.Notify(639276)
	i.iter.SeekGE(MVCCKey{Key: key})
	if err := i.tryDecodeKey(); err != nil {
		__antithesis_instrumentation__.Notify(639282)
		return
	} else {
		__antithesis_instrumentation__.Notify(639283)
	}
	__antithesis_instrumentation__.Notify(639277)
	var engineKey EngineKey
	engineKey, i.intentKeyBuf = LockTableKey{
		Key:      key,
		Strength: lock.Exclusive,
		TxnUUID:  txnUUID[:],
	}.ToEngineKey(i.intentKeyBuf)
	var limitKey roachpb.Key
	if i.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(639284)
		return !i.prefix == true
	}() == true {
		__antithesis_instrumentation__.Notify(639285)
		limitKey = i.makeUpperLimitKey()
	} else {
		__antithesis_instrumentation__.Notify(639286)
	}
	__antithesis_instrumentation__.Notify(639278)
	iterState, err := i.intentIter.SeekEngineKeyGEWithLimit(engineKey, limitKey)
	if err = i.tryDecodeLockKey(iterState, err); err != nil {
		__antithesis_instrumentation__.Notify(639287)
		return
	} else {
		__antithesis_instrumentation__.Notify(639288)
	}
	__antithesis_instrumentation__.Notify(639279)
	i.computePos()
}

func (i *intentInterleavingIter) checkConstraint(k roachpb.Key, isExclusiveUpper bool) {
	__antithesis_instrumentation__.Notify(639289)
	kConstraint := constrainedToGlobal
	if isLocal(k) {
		__antithesis_instrumentation__.Notify(639291)
		if bytes.Compare(k, keys.LocalRangeLockTablePrefix) > 0 {
			__antithesis_instrumentation__.Notify(639293)
			panic(fmt.Sprintf("intentInterleavingIter cannot be used with invalid local keys %s",
				k.String()))
		} else {
			__antithesis_instrumentation__.Notify(639294)
		}
		__antithesis_instrumentation__.Notify(639292)
		kConstraint = constrainedToLocal
	} else {
		__antithesis_instrumentation__.Notify(639295)
		if isExclusiveUpper && func() bool {
			__antithesis_instrumentation__.Notify(639296)
			return bytes.Equal(k, keys.LocalMax) == true
		}() == true {
			__antithesis_instrumentation__.Notify(639297)
			kConstraint = constrainedToLocal
		} else {
			__antithesis_instrumentation__.Notify(639298)
		}
	}
	__antithesis_instrumentation__.Notify(639290)
	if kConstraint != i.constraint {
		__antithesis_instrumentation__.Notify(639299)
		panic(fmt.Sprintf(
			"iterator with constraint=%d is being used with key %s that has constraint=%d",
			i.constraint, k.String(), kConstraint))
	} else {
		__antithesis_instrumentation__.Notify(639300)
	}
}

func (i *intentInterleavingIter) tryDecodeKey() error {
	__antithesis_instrumentation__.Notify(639301)
	i.iterValid, i.err = i.iter.Valid()
	if i.iterValid {
		__antithesis_instrumentation__.Notify(639304)
		i.iterKey = i.iter.UnsafeKey()
	} else {
		__antithesis_instrumentation__.Notify(639305)
	}
	__antithesis_instrumentation__.Notify(639302)
	if i.err != nil {
		__antithesis_instrumentation__.Notify(639306)
		i.valid = false
	} else {
		__antithesis_instrumentation__.Notify(639307)
	}
	__antithesis_instrumentation__.Notify(639303)
	return i.err
}

func (i *intentInterleavingIter) computePos() {
	__antithesis_instrumentation__.Notify(639308)
	if !i.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(639311)
		return i.intentKey == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(639312)
		i.valid = false
		return
	} else {
		__antithesis_instrumentation__.Notify(639313)
	}
	__antithesis_instrumentation__.Notify(639309)

	if !i.iterValid {
		__antithesis_instrumentation__.Notify(639314)
		i.intentCmp = -i.dir
		return
	} else {
		__antithesis_instrumentation__.Notify(639315)
	}
	__antithesis_instrumentation__.Notify(639310)
	if i.intentKey == nil {
		__antithesis_instrumentation__.Notify(639316)
		i.intentCmp = i.dir
	} else {
		__antithesis_instrumentation__.Notify(639317)
		i.intentCmp = i.intentKey.Compare(i.iterKey.Key)
	}
}

func (i *intentInterleavingIter) tryDecodeLockKey(
	iterState pebble.IterValidityState, err error,
) error {
	__antithesis_instrumentation__.Notify(639318)
	if err != nil {
		__antithesis_instrumentation__.Notify(639324)
		i.err = err
		i.valid = false
		return err
	} else {
		__antithesis_instrumentation__.Notify(639325)
	}
	__antithesis_instrumentation__.Notify(639319)
	i.intentIterState = iterState
	if iterState != pebble.IterValid {
		__antithesis_instrumentation__.Notify(639326)

		i.intentKey = nil
		return nil
	} else {
		__antithesis_instrumentation__.Notify(639327)
	}
	__antithesis_instrumentation__.Notify(639320)
	engineKey, err := i.intentIter.UnsafeEngineKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(639328)
		i.err = err
		i.valid = false
		return err
	} else {
		__antithesis_instrumentation__.Notify(639329)
	}
	__antithesis_instrumentation__.Notify(639321)
	if i.intentKey, err = keys.DecodeLockTableSingleKey(engineKey.Key); err != nil {
		__antithesis_instrumentation__.Notify(639330)
		i.err = err
		i.valid = false
		return err
	} else {
		__antithesis_instrumentation__.Notify(639331)
	}
	__antithesis_instrumentation__.Notify(639322)

	i.intentKeyAsNoTimestampMVCCKey = nil
	if cap(i.intentKey) > len(i.intentKey) {
		__antithesis_instrumentation__.Notify(639332)
		prospectiveKey := i.intentKey[:len(i.intentKey)+1]
		if prospectiveKey[len(i.intentKey)] == 0 {
			__antithesis_instrumentation__.Notify(639333)
			i.intentKeyAsNoTimestampMVCCKey = prospectiveKey
		} else {
			__antithesis_instrumentation__.Notify(639334)
		}
	} else {
		__antithesis_instrumentation__.Notify(639335)
	}
	__antithesis_instrumentation__.Notify(639323)
	return nil
}

func (i *intentInterleavingIter) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(639336)
	return i.valid, i.err
}

func (i *intentInterleavingIter) Next() {
	__antithesis_instrumentation__.Notify(639337)
	if i.err != nil {
		__antithesis_instrumentation__.Notify(639341)
		return
	} else {
		__antithesis_instrumentation__.Notify(639342)
	}
	__antithesis_instrumentation__.Notify(639338)
	if i.dir < 0 {
		__antithesis_instrumentation__.Notify(639343)

		isCurAtIntent := i.isCurAtIntentIter()
		i.dir = +1
		if !i.valid {
			__antithesis_instrumentation__.Notify(639345)

			i.valid = true
			i.iter.Next()
			if err := i.tryDecodeKey(); err != nil {
				__antithesis_instrumentation__.Notify(639349)
				return
			} else {
				__antithesis_instrumentation__.Notify(639350)
			}
			__antithesis_instrumentation__.Notify(639346)
			var limitKey roachpb.Key
			if i.iterValid && func() bool {
				__antithesis_instrumentation__.Notify(639351)
				return !i.prefix == true
			}() == true {
				__antithesis_instrumentation__.Notify(639352)
				limitKey = i.makeUpperLimitKey()
			} else {
				__antithesis_instrumentation__.Notify(639353)
			}
			__antithesis_instrumentation__.Notify(639347)
			iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639354)
				return
			} else {
				__antithesis_instrumentation__.Notify(639355)
			}
			__antithesis_instrumentation__.Notify(639348)
			i.computePos()
			return
		} else {
			__antithesis_instrumentation__.Notify(639356)
		}
		__antithesis_instrumentation__.Notify(639344)

		if isCurAtIntent {
			__antithesis_instrumentation__.Notify(639357)

			i.iter.Next()
			if err := i.tryDecodeKey(); err != nil {
				__antithesis_instrumentation__.Notify(639360)
				return
			} else {
				__antithesis_instrumentation__.Notify(639361)
			}
			__antithesis_instrumentation__.Notify(639358)
			i.intentCmp = 0
			if !i.iterValid {
				__antithesis_instrumentation__.Notify(639362)
				i.err = errors.Errorf("intent has no provisional value")
				i.valid = false
				return
			} else {
				__antithesis_instrumentation__.Notify(639363)
			}
			__antithesis_instrumentation__.Notify(639359)
			if util.RaceEnabled {
				__antithesis_instrumentation__.Notify(639364)
				cmp := i.intentKey.Compare(i.iterKey.Key)
				if cmp != 0 {
					__antithesis_instrumentation__.Notify(639365)
					i.err = errors.Errorf("intent has no provisional value, cmp: %d", cmp)
					i.valid = false
					return
				} else {
					__antithesis_instrumentation__.Notify(639366)
				}
			} else {
				__antithesis_instrumentation__.Notify(639367)
			}
		} else {
			__antithesis_instrumentation__.Notify(639368)

			var limitKey roachpb.Key
			if !i.prefix {
				__antithesis_instrumentation__.Notify(639371)
				limitKey = i.makeUpperLimitKey()
			} else {
				__antithesis_instrumentation__.Notify(639372)
			}
			__antithesis_instrumentation__.Notify(639369)
			iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639373)
				return
			} else {
				__antithesis_instrumentation__.Notify(639374)
			}
			__antithesis_instrumentation__.Notify(639370)
			i.intentCmp = +1
			if util.RaceEnabled && func() bool {
				__antithesis_instrumentation__.Notify(639375)
				return iterState == pebble.IterValid == true
			}() == true {
				__antithesis_instrumentation__.Notify(639376)
				cmp := i.intentKey.Compare(i.iterKey.Key)
				if cmp <= 0 {
					__antithesis_instrumentation__.Notify(639377)
					i.err = errors.Errorf("intentIter incorrectly positioned, cmp: %d", cmp)
					i.valid = false
					return
				} else {
					__antithesis_instrumentation__.Notify(639378)
				}
			} else {
				__antithesis_instrumentation__.Notify(639379)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(639380)
	}
	__antithesis_instrumentation__.Notify(639339)
	if !i.valid {
		__antithesis_instrumentation__.Notify(639381)
		return
	} else {
		__antithesis_instrumentation__.Notify(639382)
	}
	__antithesis_instrumentation__.Notify(639340)
	if i.intentCmp <= 0 {
		__antithesis_instrumentation__.Notify(639383)

		if i.intentCmp != 0 {
			__antithesis_instrumentation__.Notify(639388)
			i.err = errors.Errorf("intentIter at intent, but iter not at provisional value")
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(639389)
		}
		__antithesis_instrumentation__.Notify(639384)
		if !i.iterValid {
			__antithesis_instrumentation__.Notify(639390)
			i.err = errors.Errorf("iter expected to be at provisional value, but is exhausted")
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(639391)
		}
		__antithesis_instrumentation__.Notify(639385)
		var limitKey roachpb.Key
		if !i.prefix {
			__antithesis_instrumentation__.Notify(639392)
			limitKey = i.makeUpperLimitKey()
		} else {
			__antithesis_instrumentation__.Notify(639393)
		}
		__antithesis_instrumentation__.Notify(639386)
		iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
		if err = i.tryDecodeLockKey(iterState, err); err != nil {
			__antithesis_instrumentation__.Notify(639394)
			return
		} else {
			__antithesis_instrumentation__.Notify(639395)
		}
		__antithesis_instrumentation__.Notify(639387)
		i.intentCmp = +1
		if util.RaceEnabled && func() bool {
			__antithesis_instrumentation__.Notify(639396)
			return i.intentKey != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(639397)
			cmp := i.intentKey.Compare(i.iterKey.Key)
			if cmp <= 0 {
				__antithesis_instrumentation__.Notify(639398)
				i.err = errors.Errorf("intentIter incorrectly positioned, cmp: %d", cmp)
				i.valid = false
				return
			} else {
				__antithesis_instrumentation__.Notify(639399)
			}
		} else {
			__antithesis_instrumentation__.Notify(639400)
		}
	} else {
		__antithesis_instrumentation__.Notify(639401)

		i.iter.Next()
		if err := i.tryDecodeKey(); err != nil {
			__antithesis_instrumentation__.Notify(639404)
			return
		} else {
			__antithesis_instrumentation__.Notify(639405)
		}
		__antithesis_instrumentation__.Notify(639402)
		if i.intentIterState == pebble.IterAtLimit && func() bool {
			__antithesis_instrumentation__.Notify(639406)
			return i.iterValid == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(639407)
			return !i.prefix == true
		}() == true {
			__antithesis_instrumentation__.Notify(639408)

			limitKey := i.makeUpperLimitKey()
			iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639409)
				return
			} else {
				__antithesis_instrumentation__.Notify(639410)
			}
		} else {
			__antithesis_instrumentation__.Notify(639411)
		}
		__antithesis_instrumentation__.Notify(639403)
		i.computePos()
	}
}

func (i *intentInterleavingIter) NextKey() {
	__antithesis_instrumentation__.Notify(639412)

	if i.dir < 0 {
		__antithesis_instrumentation__.Notify(639418)
		i.err = errors.Errorf("NextKey cannot be used to switch iteration direction")
		i.valid = false
		return
	} else {
		__antithesis_instrumentation__.Notify(639419)
	}
	__antithesis_instrumentation__.Notify(639413)
	if !i.valid {
		__antithesis_instrumentation__.Notify(639420)
		return
	} else {
		__antithesis_instrumentation__.Notify(639421)
	}
	__antithesis_instrumentation__.Notify(639414)
	if i.intentCmp <= 0 {
		__antithesis_instrumentation__.Notify(639422)

		if i.intentCmp != 0 {
			__antithesis_instrumentation__.Notify(639427)
			i.err = errors.Errorf("intentIter at intent, but iter not at provisional value")
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(639428)
		}
		__antithesis_instrumentation__.Notify(639423)

		i.iter.NextKey()
		if err := i.tryDecodeKey(); err != nil {
			__antithesis_instrumentation__.Notify(639429)
			return
		} else {
			__antithesis_instrumentation__.Notify(639430)
		}
		__antithesis_instrumentation__.Notify(639424)
		var limitKey roachpb.Key
		if i.iterValid && func() bool {
			__antithesis_instrumentation__.Notify(639431)
			return !i.prefix == true
		}() == true {
			__antithesis_instrumentation__.Notify(639432)
			limitKey = i.makeUpperLimitKey()
		} else {
			__antithesis_instrumentation__.Notify(639433)
		}
		__antithesis_instrumentation__.Notify(639425)
		iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
		if err := i.tryDecodeLockKey(iterState, err); err != nil {
			__antithesis_instrumentation__.Notify(639434)
			return
		} else {
			__antithesis_instrumentation__.Notify(639435)
		}
		__antithesis_instrumentation__.Notify(639426)
		i.computePos()
		return
	} else {
		__antithesis_instrumentation__.Notify(639436)
	}
	__antithesis_instrumentation__.Notify(639415)

	i.iter.NextKey()
	if err := i.tryDecodeKey(); err != nil {
		__antithesis_instrumentation__.Notify(639437)
		return
	} else {
		__antithesis_instrumentation__.Notify(639438)
	}
	__antithesis_instrumentation__.Notify(639416)
	if i.intentIterState == pebble.IterAtLimit && func() bool {
		__antithesis_instrumentation__.Notify(639439)
		return i.iterValid == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(639440)
		return !i.prefix == true
	}() == true {
		__antithesis_instrumentation__.Notify(639441)
		limitKey := i.makeUpperLimitKey()
		iterState, err := i.intentIter.NextEngineKeyWithLimit(limitKey)
		if err = i.tryDecodeLockKey(iterState, err); err != nil {
			__antithesis_instrumentation__.Notify(639442)
			return
		} else {
			__antithesis_instrumentation__.Notify(639443)
		}
	} else {
		__antithesis_instrumentation__.Notify(639444)
	}
	__antithesis_instrumentation__.Notify(639417)
	i.computePos()
}

func (i *intentInterleavingIter) isCurAtIntentIter() bool {
	__antithesis_instrumentation__.Notify(639445)

	return (i.dir > 0 && func() bool {
		__antithesis_instrumentation__.Notify(639446)
		return i.intentCmp <= 0 == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(639447)
		return (i.dir < 0 && func() bool {
			__antithesis_instrumentation__.Notify(639448)
			return i.intentCmp > 0 == true
		}() == true) == true
	}() == true
}

func (i *intentInterleavingIter) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(639449)

	if i.isCurAtIntentIter() {
		__antithesis_instrumentation__.Notify(639451)
		return MVCCKey{Key: i.intentKey}
	} else {
		__antithesis_instrumentation__.Notify(639452)
	}
	__antithesis_instrumentation__.Notify(639450)
	return i.iterKey
}

func (i *intentInterleavingIter) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(639453)
	if i.isCurAtIntentIter() {
		__antithesis_instrumentation__.Notify(639455)
		return i.intentIter.UnsafeValue()
	} else {
		__antithesis_instrumentation__.Notify(639456)
	}
	__antithesis_instrumentation__.Notify(639454)
	return i.iter.UnsafeValue()
}

func (i *intentInterleavingIter) Key() MVCCKey {
	__antithesis_instrumentation__.Notify(639457)
	key := i.UnsafeKey()
	keyCopy := make([]byte, len(key.Key))
	copy(keyCopy, key.Key)
	key.Key = keyCopy
	return key
}

func (i *intentInterleavingIter) Value() []byte {
	__antithesis_instrumentation__.Notify(639458)
	if i.isCurAtIntentIter() {
		__antithesis_instrumentation__.Notify(639460)
		return i.intentIter.Value()
	} else {
		__antithesis_instrumentation__.Notify(639461)
	}
	__antithesis_instrumentation__.Notify(639459)
	return i.iter.Value()
}

func (i *intentInterleavingIter) Close() {
	__antithesis_instrumentation__.Notify(639462)
	i.iter.Close()
	i.intentIter.Close()
	*i = intentInterleavingIter{
		intentKeyAsNoTimestampMVCCKeyBacking: i.intentKeyAsNoTimestampMVCCKeyBacking,
		intentKeyBuf:                         i.intentKeyBuf,
		intentLimitKeyBuf:                    i.intentLimitKeyBuf,
	}
	intentInterleavingIterPool.Put(i)
}

func (i *intentInterleavingIter) SeekLT(key MVCCKey) {
	__antithesis_instrumentation__.Notify(639463)
	i.dir = -1
	i.valid = true
	i.err = nil

	if i.prefix {
		__antithesis_instrumentation__.Notify(639470)
		i.err = errors.Errorf("prefix iteration is not permitted with SeekLT")
		i.valid = false
		return
	} else {
		__antithesis_instrumentation__.Notify(639471)
	}
	__antithesis_instrumentation__.Notify(639464)
	if i.constraint != notConstrained {
		__antithesis_instrumentation__.Notify(639472)
		i.checkConstraint(key.Key, true)
		if i.constraint == constrainedToLocal && func() bool {
			__antithesis_instrumentation__.Notify(639473)
			return bytes.Equal(key.Key, keys.LocalMax) == true
		}() == true {
			__antithesis_instrumentation__.Notify(639474)

			key.Key = keys.LocalRangeLockTablePrefix
		} else {
			__antithesis_instrumentation__.Notify(639475)
		}
	} else {
		__antithesis_instrumentation__.Notify(639476)
	}
	__antithesis_instrumentation__.Notify(639465)

	i.iter.SeekLT(key)
	if err := i.tryDecodeKey(); err != nil {
		__antithesis_instrumentation__.Notify(639477)
		return
	} else {
		__antithesis_instrumentation__.Notify(639478)
	}
	__antithesis_instrumentation__.Notify(639466)
	var intentSeekKey roachpb.Key
	if key.Timestamp.IsEmpty() {
		__antithesis_instrumentation__.Notify(639479)

		intentSeekKey, i.intentKeyBuf = keys.LockTableSingleKey(key.Key, i.intentKeyBuf)
	} else {
		__antithesis_instrumentation__.Notify(639480)

		intentSeekKey, i.intentKeyBuf = keys.LockTableSingleKey(key.Key.Next(), i.intentKeyBuf)
	}
	__antithesis_instrumentation__.Notify(639467)
	var limitKey roachpb.Key
	if i.iterValid {
		__antithesis_instrumentation__.Notify(639481)
		limitKey = i.makeLowerLimitKey()
	} else {
		__antithesis_instrumentation__.Notify(639482)
	}
	__antithesis_instrumentation__.Notify(639468)
	iterState, err := i.intentIter.SeekEngineKeyLTWithLimit(EngineKey{Key: intentSeekKey}, limitKey)
	if err = i.tryDecodeLockKey(iterState, err); err != nil {
		__antithesis_instrumentation__.Notify(639483)
		return
	} else {
		__antithesis_instrumentation__.Notify(639484)
	}
	__antithesis_instrumentation__.Notify(639469)
	i.computePos()
}

func (i *intentInterleavingIter) Prev() {
	__antithesis_instrumentation__.Notify(639485)
	if i.err != nil {
		__antithesis_instrumentation__.Notify(639489)
		return
	} else {
		__antithesis_instrumentation__.Notify(639490)
	}
	__antithesis_instrumentation__.Notify(639486)
	if i.dir > 0 {
		__antithesis_instrumentation__.Notify(639491)

		isCurAtIntent := i.isCurAtIntentIter()
		i.dir = -1
		if !i.valid {
			__antithesis_instrumentation__.Notify(639493)

			i.valid = true
			i.iter.Prev()
			if err := i.tryDecodeKey(); err != nil {
				__antithesis_instrumentation__.Notify(639497)
				return
			} else {
				__antithesis_instrumentation__.Notify(639498)
			}
			__antithesis_instrumentation__.Notify(639494)
			var limitKey roachpb.Key
			if i.iterValid {
				__antithesis_instrumentation__.Notify(639499)
				limitKey = i.makeLowerLimitKey()
			} else {
				__antithesis_instrumentation__.Notify(639500)
			}
			__antithesis_instrumentation__.Notify(639495)
			iterState, err := i.intentIter.PrevEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639501)
				return
			} else {
				__antithesis_instrumentation__.Notify(639502)
			}
			__antithesis_instrumentation__.Notify(639496)
			i.computePos()
			return
		} else {
			__antithesis_instrumentation__.Notify(639503)
		}
		__antithesis_instrumentation__.Notify(639492)

		if isCurAtIntent {
			__antithesis_instrumentation__.Notify(639504)

			if i.intentCmp != 0 {
				__antithesis_instrumentation__.Notify(639507)
				i.err = errors.Errorf("iter not at provisional value, cmp: %d", i.intentCmp)
				i.valid = false
				return
			} else {
				__antithesis_instrumentation__.Notify(639508)
			}
			__antithesis_instrumentation__.Notify(639505)
			i.iter.Prev()
			if err := i.tryDecodeKey(); err != nil {
				__antithesis_instrumentation__.Notify(639509)
				return
			} else {
				__antithesis_instrumentation__.Notify(639510)
			}
			__antithesis_instrumentation__.Notify(639506)
			i.intentCmp = +1
			if util.RaceEnabled && func() bool {
				__antithesis_instrumentation__.Notify(639511)
				return i.iterValid == true
			}() == true {
				__antithesis_instrumentation__.Notify(639512)
				cmp := i.intentKey.Compare(i.iterKey.Key)
				if cmp <= 0 {
					__antithesis_instrumentation__.Notify(639513)
					i.err = errors.Errorf("intentIter should be after iter, cmp: %d", cmp)
					i.valid = false
					return
				} else {
					__antithesis_instrumentation__.Notify(639514)
				}
			} else {
				__antithesis_instrumentation__.Notify(639515)
			}
		} else {
			__antithesis_instrumentation__.Notify(639516)

			limitKey := i.makeLowerLimitKey()
			iterState, err := i.intentIter.PrevEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639518)
				return
			} else {
				__antithesis_instrumentation__.Notify(639519)
			}
			__antithesis_instrumentation__.Notify(639517)
			if i.intentKey == nil {
				__antithesis_instrumentation__.Notify(639520)
				i.intentCmp = -1
			} else {
				__antithesis_instrumentation__.Notify(639521)
				i.intentCmp = i.intentKey.Compare(i.iterKey.Key)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(639522)
	}
	__antithesis_instrumentation__.Notify(639487)
	if !i.valid {
		__antithesis_instrumentation__.Notify(639523)
		return
	} else {
		__antithesis_instrumentation__.Notify(639524)
	}
	__antithesis_instrumentation__.Notify(639488)
	if i.intentCmp > 0 {
		__antithesis_instrumentation__.Notify(639525)

		var limitKey roachpb.Key
		if i.iterValid {
			__antithesis_instrumentation__.Notify(639529)
			limitKey = i.makeLowerLimitKey()
		} else {
			__antithesis_instrumentation__.Notify(639530)
		}
		__antithesis_instrumentation__.Notify(639526)
		intentIterState, err := i.intentIter.PrevEngineKeyWithLimit(limitKey)
		if err = i.tryDecodeLockKey(intentIterState, err); err != nil {
			__antithesis_instrumentation__.Notify(639531)
			return
		} else {
			__antithesis_instrumentation__.Notify(639532)
		}
		__antithesis_instrumentation__.Notify(639527)
		if !i.iterValid {
			__antithesis_instrumentation__.Notify(639533)

			if intentIterState != pebble.IterExhausted {
				__antithesis_instrumentation__.Notify(639535)
				i.err = errors.Errorf("reverse iteration discovered intent without provisional value")
			} else {
				__antithesis_instrumentation__.Notify(639536)
			}
			__antithesis_instrumentation__.Notify(639534)
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(639537)
		}
		__antithesis_instrumentation__.Notify(639528)

		i.intentCmp = -1
		if i.intentKey != nil {
			__antithesis_instrumentation__.Notify(639538)
			i.intentCmp = i.intentKey.Compare(i.iterKey.Key)
			if i.intentCmp > 0 {
				__antithesis_instrumentation__.Notify(639539)
				i.err = errors.Errorf("intentIter should not be after iter")
				i.valid = false
				return
			} else {
				__antithesis_instrumentation__.Notify(639540)
			}
		} else {
			__antithesis_instrumentation__.Notify(639541)
		}
	} else {
		__antithesis_instrumentation__.Notify(639542)

		i.iter.Prev()
		if err := i.tryDecodeKey(); err != nil {
			__antithesis_instrumentation__.Notify(639545)
			return
		} else {
			__antithesis_instrumentation__.Notify(639546)
		}
		__antithesis_instrumentation__.Notify(639543)
		if i.intentIterState == pebble.IterAtLimit && func() bool {
			__antithesis_instrumentation__.Notify(639547)
			return i.iterValid == true
		}() == true {
			__antithesis_instrumentation__.Notify(639548)

			limitKey := i.makeLowerLimitKey()
			iterState, err := i.intentIter.PrevEngineKeyWithLimit(limitKey)
			if err = i.tryDecodeLockKey(iterState, err); err != nil {
				__antithesis_instrumentation__.Notify(639549)
				return
			} else {
				__antithesis_instrumentation__.Notify(639550)
			}
		} else {
			__antithesis_instrumentation__.Notify(639551)
		}
		__antithesis_instrumentation__.Notify(639544)
		i.computePos()
	}
}

func (i *intentInterleavingIter) UnsafeRawKey() []byte {
	__antithesis_instrumentation__.Notify(639552)
	if i.isCurAtIntentIter() {
		__antithesis_instrumentation__.Notify(639554)
		return i.intentIter.UnsafeRawEngineKey()
	} else {
		__antithesis_instrumentation__.Notify(639555)
	}
	__antithesis_instrumentation__.Notify(639553)
	return i.iter.UnsafeRawKey()
}

func (i *intentInterleavingIter) UnsafeRawMVCCKey() []byte {
	__antithesis_instrumentation__.Notify(639556)
	if i.isCurAtIntentIter() {
		__antithesis_instrumentation__.Notify(639558)
		if i.intentKeyAsNoTimestampMVCCKey == nil {
			__antithesis_instrumentation__.Notify(639560)

			if cap(i.intentKeyAsNoTimestampMVCCKeyBacking) < len(i.intentKey)+1 {
				__antithesis_instrumentation__.Notify(639562)
				i.intentKeyAsNoTimestampMVCCKeyBacking = make([]byte, 0, len(i.intentKey)+1)
			} else {
				__antithesis_instrumentation__.Notify(639563)
			}
			__antithesis_instrumentation__.Notify(639561)
			i.intentKeyAsNoTimestampMVCCKeyBacking = append(
				i.intentKeyAsNoTimestampMVCCKeyBacking[:0], i.intentKey...)

			i.intentKeyAsNoTimestampMVCCKeyBacking = append(
				i.intentKeyAsNoTimestampMVCCKeyBacking, 0)
			i.intentKeyAsNoTimestampMVCCKey = i.intentKeyAsNoTimestampMVCCKeyBacking
		} else {
			__antithesis_instrumentation__.Notify(639564)
		}
		__antithesis_instrumentation__.Notify(639559)
		return i.intentKeyAsNoTimestampMVCCKey
	} else {
		__antithesis_instrumentation__.Notify(639565)
	}
	__antithesis_instrumentation__.Notify(639557)
	return i.iter.UnsafeRawKey()
}

func (i *intentInterleavingIter) ValueProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(639566)
	value := i.UnsafeValue()
	return protoutil.Unmarshal(value, msg)
}

func (i *intentInterleavingIter) ComputeStats(
	start, end roachpb.Key, nowNanos int64,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(639567)
	return ComputeStatsForRange(i, start, end, nowNanos)
}

func (i *intentInterleavingIter) FindSplitKey(
	start, end, minSplitKey roachpb.Key, targetSize int64,
) (MVCCKey, error) {
	__antithesis_instrumentation__.Notify(639568)
	return findSplitKeyUsingIterator(i, start, end, minSplitKey, targetSize)
}

func (i *intentInterleavingIter) SetUpperBound(key roachpb.Key) {
	__antithesis_instrumentation__.Notify(639569)
	i.iter.SetUpperBound(key)

	if i.constraint != notConstrained {
		__antithesis_instrumentation__.Notify(639571)
		i.checkConstraint(key, true)
	} else {
		__antithesis_instrumentation__.Notify(639572)
	}
	__antithesis_instrumentation__.Notify(639570)
	var intentUpperBound roachpb.Key
	intentUpperBound, i.intentKeyBuf = keys.LockTableSingleKey(key, i.intentKeyBuf)
	i.intentIter.SetUpperBound(intentUpperBound)
}

func (i *intentInterleavingIter) Stats() IteratorStats {
	__antithesis_instrumentation__.Notify(639573)
	stats := i.iter.Stats()
	intentStats := i.intentIter.Stats()
	stats.InternalDeleteSkippedCount += intentStats.InternalDeleteSkippedCount
	stats.TimeBoundNumSSTs += intentStats.TimeBoundNumSSTs
	for i := pebble.IteratorStatsKind(0); i < pebble.NumStatsKind; i++ {
		__antithesis_instrumentation__.Notify(639575)
		stats.Stats.ForwardSeekCount[i] += intentStats.Stats.ForwardSeekCount[i]
		stats.Stats.ReverseSeekCount[i] += intentStats.Stats.ReverseSeekCount[i]
		stats.Stats.ForwardStepCount[i] += intentStats.Stats.ForwardStepCount[i]
		stats.Stats.ReverseStepCount[i] += intentStats.Stats.ReverseStepCount[i]
	}
	__antithesis_instrumentation__.Notify(639574)
	stats.Stats.InternalStats.Merge(intentStats.Stats.InternalStats)
	return stats
}

func (i *intentInterleavingIter) SupportsPrev() bool {
	__antithesis_instrumentation__.Notify(639576)
	return true
}

func newMVCCIteratorByCloningEngineIter(iter EngineIterator, opts IterOptions) MVCCIterator {
	__antithesis_instrumentation__.Notify(639577)
	pIter := iter.GetRawIter()
	it := newPebbleIterator(nil, pIter, opts, StandardDurability)
	if iter == nil {
		__antithesis_instrumentation__.Notify(639579)
		panic("couldn't create a new iterator")
	} else {
		__antithesis_instrumentation__.Notify(639580)
	}
	__antithesis_instrumentation__.Notify(639578)
	return it
}

type unsafeMVCCIterator struct {
	MVCCIterator
	keyBuf        []byte
	rawKeyBuf     []byte
	rawMVCCKeyBuf []byte
}

func wrapInUnsafeIter(iter MVCCIterator) MVCCIterator {
	__antithesis_instrumentation__.Notify(639581)
	return &unsafeMVCCIterator{MVCCIterator: iter}
}

var _ MVCCIterator = &unsafeMVCCIterator{}

func (i *unsafeMVCCIterator) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(639582)
	i.mangleBufs()
	i.MVCCIterator.SeekGE(key)
}

func (i *unsafeMVCCIterator) Next() {
	__antithesis_instrumentation__.Notify(639583)
	i.mangleBufs()
	i.MVCCIterator.Next()
}

func (i *unsafeMVCCIterator) NextKey() {
	__antithesis_instrumentation__.Notify(639584)
	i.mangleBufs()
	i.MVCCIterator.NextKey()
}

func (i *unsafeMVCCIterator) SeekLT(key MVCCKey) {
	__antithesis_instrumentation__.Notify(639585)
	i.mangleBufs()
	i.MVCCIterator.SeekLT(key)
}

func (i *unsafeMVCCIterator) Prev() {
	__antithesis_instrumentation__.Notify(639586)
	i.mangleBufs()
	i.MVCCIterator.Prev()
}

func (i *unsafeMVCCIterator) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(639587)
	rv := i.MVCCIterator.UnsafeKey()
	i.keyBuf = append(i.keyBuf[:0], rv.Key...)
	rv.Key = i.keyBuf
	return rv
}

func (i *unsafeMVCCIterator) UnsafeRawKey() []byte {
	__antithesis_instrumentation__.Notify(639588)
	rv := i.MVCCIterator.UnsafeRawKey()
	i.rawKeyBuf = append(i.rawKeyBuf[:0], rv...)
	return i.rawKeyBuf
}

func (i *unsafeMVCCIterator) UnsafeRawMVCCKey() []byte {
	__antithesis_instrumentation__.Notify(639589)
	rv := i.MVCCIterator.UnsafeRawMVCCKey()
	i.rawMVCCKeyBuf = append(i.rawMVCCKeyBuf[:0], rv...)
	return i.rawMVCCKeyBuf
}

func (i *unsafeMVCCIterator) mangleBufs() {
	__antithesis_instrumentation__.Notify(639590)
	if rand.Intn(2) == 0 {
		__antithesis_instrumentation__.Notify(639591)
		for _, b := range [3][]byte{i.keyBuf, i.rawKeyBuf, i.rawMVCCKeyBuf} {
			__antithesis_instrumentation__.Notify(639592)
			for i := range b {
				__antithesis_instrumentation__.Notify(639593)
				b[i] = 0
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(639594)
	}
}
