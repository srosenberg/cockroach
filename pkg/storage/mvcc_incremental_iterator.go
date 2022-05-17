package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type MVCCIncrementalIterator struct {
	iter MVCCIterator

	timeBoundIter MVCCIterator

	startTime hlc.Timestamp
	endTime   hlc.Timestamp
	err       error
	valid     bool

	meta enginepb.MVCCMetadata

	intentPolicy MVCCIncrementalIterIntentPolicy
	inlinePolicy MVCCIncrementalIterInlinePolicy

	intents []roachpb.Intent
}

var _ SimpleMVCCIterator = &MVCCIncrementalIterator{}

type MVCCIncrementalIterIntentPolicy int

const (
	MVCCIncrementalIterIntentPolicyError MVCCIncrementalIterIntentPolicy = iota

	MVCCIncrementalIterIntentPolicyAggregate

	MVCCIncrementalIterIntentPolicyEmit
)

type MVCCIncrementalIterInlinePolicy int

const (
	MVCCIncrementalIterInlinePolicyError MVCCIncrementalIterInlinePolicy = iota

	MVCCIncrementalIterInlinePolicyEmit
)

type MVCCIncrementalIterOptions struct {
	EnableTimeBoundIteratorOptimization bool
	EndKey                              roachpb.Key

	StartTime hlc.Timestamp
	EndTime   hlc.Timestamp

	IntentPolicy MVCCIncrementalIterIntentPolicy
	InlinePolicy MVCCIncrementalIterInlinePolicy
}

func NewMVCCIncrementalIterator(
	reader Reader, opts MVCCIncrementalIterOptions,
) *MVCCIncrementalIterator {
	__antithesis_instrumentation__.Notify(641383)
	var iter MVCCIterator
	var timeBoundIter MVCCIterator
	if opts.EnableTimeBoundIteratorOptimization {
		__antithesis_instrumentation__.Notify(641385)

		iter = reader.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{
			UpperBound: opts.EndKey,
		})

		timeBoundIter = reader.NewMVCCIterator(MVCCKeyIterKind, IterOptions{
			UpperBound: opts.EndKey,

			MinTimestampHint: opts.StartTime.Next(),
			MaxTimestampHint: opts.EndTime,
		})
	} else {
		__antithesis_instrumentation__.Notify(641386)
		iter = reader.NewMVCCIterator(MVCCKeyAndIntentsIterKind, IterOptions{
			UpperBound: opts.EndKey,
		})
	}
	__antithesis_instrumentation__.Notify(641384)

	return &MVCCIncrementalIterator{
		iter:          iter,
		startTime:     opts.StartTime,
		endTime:       opts.EndTime,
		timeBoundIter: timeBoundIter,
		intentPolicy:  opts.IntentPolicy,
		inlinePolicy:  opts.InlinePolicy,
	}
}

func (i *MVCCIncrementalIterator) SeekGE(startKey MVCCKey) {
	__antithesis_instrumentation__.Notify(641387)
	if i.timeBoundIter != nil {
		__antithesis_instrumentation__.Notify(641390)

		i.timeBoundIter.SeekGE(startKey)
		if ok, err := i.timeBoundIter.Valid(); !ok {
			__antithesis_instrumentation__.Notify(641392)
			i.err = err
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(641393)
		}
		__antithesis_instrumentation__.Notify(641391)
		tbiKey := i.timeBoundIter.Key().Key
		if tbiKey.Compare(startKey.Key) > 0 {
			__antithesis_instrumentation__.Notify(641394)

			startKey = MakeMVCCMetadataKey(tbiKey)
		} else {
			__antithesis_instrumentation__.Notify(641395)
		}
	} else {
		__antithesis_instrumentation__.Notify(641396)
	}
	__antithesis_instrumentation__.Notify(641388)
	i.iter.SeekGE(startKey)
	if !i.checkValidAndSaveErr() {
		__antithesis_instrumentation__.Notify(641397)
		return
	} else {
		__antithesis_instrumentation__.Notify(641398)
	}
	__antithesis_instrumentation__.Notify(641389)
	i.err = nil
	i.valid = true
	i.advance()
}

func (i *MVCCIncrementalIterator) Close() {
	__antithesis_instrumentation__.Notify(641399)
	i.iter.Close()
	if i.timeBoundIter != nil {
		__antithesis_instrumentation__.Notify(641400)
		i.timeBoundIter.Close()
	} else {
		__antithesis_instrumentation__.Notify(641401)
	}
}

func (i *MVCCIncrementalIterator) Next() {
	__antithesis_instrumentation__.Notify(641402)
	i.iter.Next()
	if !i.checkValidAndSaveErr() {
		__antithesis_instrumentation__.Notify(641404)
		return
	} else {
		__antithesis_instrumentation__.Notify(641405)
	}
	__antithesis_instrumentation__.Notify(641403)
	i.advance()
}

func (i *MVCCIncrementalIterator) checkValidAndSaveErr() bool {
	__antithesis_instrumentation__.Notify(641406)
	if ok, err := i.iter.Valid(); !ok {
		__antithesis_instrumentation__.Notify(641408)
		i.err = err
		i.valid = false
		return false
	} else {
		__antithesis_instrumentation__.Notify(641409)
	}
	__antithesis_instrumentation__.Notify(641407)
	return true
}

func (i *MVCCIncrementalIterator) NextKey() {
	__antithesis_instrumentation__.Notify(641410)
	i.iter.NextKey()
	if !i.checkValidAndSaveErr() {
		__antithesis_instrumentation__.Notify(641412)
		return
	} else {
		__antithesis_instrumentation__.Notify(641413)
	}
	__antithesis_instrumentation__.Notify(641411)
	i.advance()
}

func (i *MVCCIncrementalIterator) maybeSkipKeys() {
	__antithesis_instrumentation__.Notify(641414)
	if i.timeBoundIter == nil {
		__antithesis_instrumentation__.Notify(641416)

		return
	} else {
		__antithesis_instrumentation__.Notify(641417)
	}
	__antithesis_instrumentation__.Notify(641415)
	tbiKey := i.timeBoundIter.UnsafeKey().Key
	iterKey := i.iter.UnsafeKey().Key
	if iterKey.Compare(tbiKey) > 0 {
		__antithesis_instrumentation__.Notify(641418)

		i.timeBoundIter.NextKey()
		if ok, err := i.timeBoundIter.Valid(); !ok {
			__antithesis_instrumentation__.Notify(641421)
			i.err = err
			i.valid = false
			return
		} else {
			__antithesis_instrumentation__.Notify(641422)
		}
		__antithesis_instrumentation__.Notify(641419)
		tbiKey = i.timeBoundIter.UnsafeKey().Key

		cmp := iterKey.Compare(tbiKey)

		if cmp > 0 {
			__antithesis_instrumentation__.Notify(641423)

			seekKey := MakeMVCCMetadataKey(iterKey)
			i.timeBoundIter.SeekGE(seekKey)
			if ok, err := i.timeBoundIter.Valid(); !ok {
				__antithesis_instrumentation__.Notify(641425)
				i.err = err
				i.valid = false
				return
			} else {
				__antithesis_instrumentation__.Notify(641426)
			}
			__antithesis_instrumentation__.Notify(641424)
			tbiKey = i.timeBoundIter.UnsafeKey().Key
			cmp = iterKey.Compare(tbiKey)
		} else {
			__antithesis_instrumentation__.Notify(641427)
		}
		__antithesis_instrumentation__.Notify(641420)

		if cmp < 0 {
			__antithesis_instrumentation__.Notify(641428)

			seekKey := MakeMVCCMetadataKey(tbiKey)
			i.iter.SeekGE(seekKey)
			if !i.checkValidAndSaveErr() {
				__antithesis_instrumentation__.Notify(641429)
				return
			} else {
				__antithesis_instrumentation__.Notify(641430)
			}
		} else {
			__antithesis_instrumentation__.Notify(641431)
		}
	} else {
		__antithesis_instrumentation__.Notify(641432)
	}
}

func (i *MVCCIncrementalIterator) initMetaAndCheckForIntentOrInlineError() error {
	__antithesis_instrumentation__.Notify(641433)
	unsafeKey := i.iter.UnsafeKey()
	if unsafeKey.IsValue() {
		__antithesis_instrumentation__.Notify(641439)

		i.meta.Reset()
		i.meta.Timestamp = unsafeKey.Timestamp.ToLegacyTimestamp()
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641440)
	}
	__antithesis_instrumentation__.Notify(641434)

	if i.err = protoutil.Unmarshal(i.iter.UnsafeValue(), &i.meta); i.err != nil {
		__antithesis_instrumentation__.Notify(641441)
		i.valid = false
		return i.err
	} else {
		__antithesis_instrumentation__.Notify(641442)
	}
	__antithesis_instrumentation__.Notify(641435)

	if i.meta.IsInline() {
		__antithesis_instrumentation__.Notify(641443)
		switch i.inlinePolicy {
		case MVCCIncrementalIterInlinePolicyError:
			__antithesis_instrumentation__.Notify(641444)
			i.valid = false
			i.err = errors.Errorf("unexpected inline value found: %s", unsafeKey.Key)
			return i.err
		case MVCCIncrementalIterInlinePolicyEmit:
			__antithesis_instrumentation__.Notify(641445)
			return nil
		default:
			__antithesis_instrumentation__.Notify(641446)
			return errors.AssertionFailedf("unknown inline policy: %d", i.inlinePolicy)
		}
	} else {
		__antithesis_instrumentation__.Notify(641447)
	}
	__antithesis_instrumentation__.Notify(641436)

	if i.meta.Txn == nil {
		__antithesis_instrumentation__.Notify(641448)
		i.valid = false
		i.err = errors.Errorf("intent is missing a txn: %s", unsafeKey.Key)
	} else {
		__antithesis_instrumentation__.Notify(641449)
	}
	__antithesis_instrumentation__.Notify(641437)

	metaTimestamp := i.meta.Timestamp.ToTimestamp()
	if i.startTime.Less(metaTimestamp) && func() bool {
		__antithesis_instrumentation__.Notify(641450)
		return metaTimestamp.LessEq(i.endTime) == true
	}() == true {
		__antithesis_instrumentation__.Notify(641451)
		switch i.intentPolicy {
		case MVCCIncrementalIterIntentPolicyError:
			__antithesis_instrumentation__.Notify(641452)
			i.err = &roachpb.WriteIntentError{
				Intents: []roachpb.Intent{
					roachpb.MakeIntent(i.meta.Txn, i.iter.Key().Key),
				},
			}
			i.valid = false
			return i.err
		case MVCCIncrementalIterIntentPolicyAggregate:
			__antithesis_instrumentation__.Notify(641453)

			i.intents = append(i.intents, roachpb.MakeIntent(i.meta.Txn, i.iter.Key().Key))
			return nil
		case MVCCIncrementalIterIntentPolicyEmit:
			__antithesis_instrumentation__.Notify(641454)

			return nil
		default:
			__antithesis_instrumentation__.Notify(641455)
			return errors.AssertionFailedf("unknown intent policy: %d", i.intentPolicy)
		}
	} else {
		__antithesis_instrumentation__.Notify(641456)
	}
	__antithesis_instrumentation__.Notify(641438)
	return nil
}

func (i *MVCCIncrementalIterator) advance() {
	__antithesis_instrumentation__.Notify(641457)
	for {
		__antithesis_instrumentation__.Notify(641458)
		i.maybeSkipKeys()
		if !i.valid {
			__antithesis_instrumentation__.Notify(641464)
			return
		} else {
			__antithesis_instrumentation__.Notify(641465)
		}
		__antithesis_instrumentation__.Notify(641459)

		if err := i.initMetaAndCheckForIntentOrInlineError(); err != nil {
			__antithesis_instrumentation__.Notify(641466)
			return
		} else {
			__antithesis_instrumentation__.Notify(641467)
		}
		__antithesis_instrumentation__.Notify(641460)

		if i.meta.IsInline() && func() bool {
			__antithesis_instrumentation__.Notify(641468)
			return i.inlinePolicy == MVCCIncrementalIterInlinePolicyEmit == true
		}() == true {
			__antithesis_instrumentation__.Notify(641469)
			return
		} else {
			__antithesis_instrumentation__.Notify(641470)
		}
		__antithesis_instrumentation__.Notify(641461)

		if i.meta.Txn != nil {
			__antithesis_instrumentation__.Notify(641471)
			switch i.intentPolicy {
			case MVCCIncrementalIterIntentPolicyEmit:
				__antithesis_instrumentation__.Notify(641472)

			case MVCCIncrementalIterIntentPolicyError, MVCCIncrementalIterIntentPolicyAggregate:
				__antithesis_instrumentation__.Notify(641473)

				i.iter.Next()
				if !i.checkValidAndSaveErr() {
					__antithesis_instrumentation__.Notify(641476)
					return
				} else {
					__antithesis_instrumentation__.Notify(641477)
				}
				__antithesis_instrumentation__.Notify(641474)
				continue
			default:
				__antithesis_instrumentation__.Notify(641475)
			}
		} else {
			__antithesis_instrumentation__.Notify(641478)
		}
		__antithesis_instrumentation__.Notify(641462)

		metaTimestamp := i.meta.Timestamp.ToTimestamp()
		if i.endTime.Less(metaTimestamp) {
			__antithesis_instrumentation__.Notify(641479)
			i.iter.Next()
		} else {
			__antithesis_instrumentation__.Notify(641480)
			if metaTimestamp.LessEq(i.startTime) {
				__antithesis_instrumentation__.Notify(641481)
				i.iter.NextKey()
			} else {
				__antithesis_instrumentation__.Notify(641482)

				break
			}
		}
		__antithesis_instrumentation__.Notify(641463)
		if !i.checkValidAndSaveErr() {
			__antithesis_instrumentation__.Notify(641483)
			return
		} else {
			__antithesis_instrumentation__.Notify(641484)
		}
	}
}

func (i *MVCCIncrementalIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(641485)
	return i.valid, i.err
}

func (i *MVCCIncrementalIterator) Key() MVCCKey {
	__antithesis_instrumentation__.Notify(641486)
	return i.iter.Key()
}

func (i *MVCCIncrementalIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(641487)
	return i.iter.Value()
}

func (i *MVCCIncrementalIterator) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(641488)
	return i.iter.UnsafeKey()
}

func (i *MVCCIncrementalIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(641489)
	return i.iter.UnsafeValue()
}

func (i *MVCCIncrementalIterator) NextIgnoringTime() {
	__antithesis_instrumentation__.Notify(641490)
	for {
		__antithesis_instrumentation__.Notify(641491)
		i.iter.Next()
		if !i.checkValidAndSaveErr() {
			__antithesis_instrumentation__.Notify(641495)
			return
		} else {
			__antithesis_instrumentation__.Notify(641496)
		}
		__antithesis_instrumentation__.Notify(641492)

		if err := i.initMetaAndCheckForIntentOrInlineError(); err != nil {
			__antithesis_instrumentation__.Notify(641497)
			return
		} else {
			__antithesis_instrumentation__.Notify(641498)
		}
		__antithesis_instrumentation__.Notify(641493)

		if i.meta.Txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(641499)
			return i.intentPolicy != MVCCIncrementalIterIntentPolicyEmit == true
		}() == true {
			__antithesis_instrumentation__.Notify(641500)
			continue
		} else {
			__antithesis_instrumentation__.Notify(641501)
		}
		__antithesis_instrumentation__.Notify(641494)

		return
	}
}

func (i *MVCCIncrementalIterator) NextKeyIgnoringTime() {
	__antithesis_instrumentation__.Notify(641502)
	i.iter.NextKey()
	for {
		__antithesis_instrumentation__.Notify(641503)
		if !i.checkValidAndSaveErr() {
			__antithesis_instrumentation__.Notify(641507)
			return
		} else {
			__antithesis_instrumentation__.Notify(641508)
		}
		__antithesis_instrumentation__.Notify(641504)

		if err := i.initMetaAndCheckForIntentOrInlineError(); err != nil {
			__antithesis_instrumentation__.Notify(641509)
			return
		} else {
			__antithesis_instrumentation__.Notify(641510)
		}
		__antithesis_instrumentation__.Notify(641505)

		if i.meta.Txn != nil && func() bool {
			__antithesis_instrumentation__.Notify(641511)
			return i.intentPolicy != MVCCIncrementalIterIntentPolicyEmit == true
		}() == true {
			__antithesis_instrumentation__.Notify(641512)
			i.Next()
			continue
		} else {
			__antithesis_instrumentation__.Notify(641513)
		}
		__antithesis_instrumentation__.Notify(641506)

		return
	}
}

func (i *MVCCIncrementalIterator) NumCollectedIntents() int {
	__antithesis_instrumentation__.Notify(641514)
	return len(i.intents)
}

func (i *MVCCIncrementalIterator) TryGetIntentError() error {
	__antithesis_instrumentation__.Notify(641515)
	if len(i.intents) == 0 {
		__antithesis_instrumentation__.Notify(641517)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(641518)
	}
	__antithesis_instrumentation__.Notify(641516)
	return &roachpb.WriteIntentError{
		Intents: i.intents,
	}
}
