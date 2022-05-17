package rangefeed

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type CatchUpIterator struct {
	simpleCatchupIter
	close func()
}

type simpleCatchupIter interface {
	storage.SimpleMVCCIterator
	NextIgnoringTime()
}

type simpleCatchupIterAdapter struct {
	storage.SimpleMVCCIterator
}

func (i simpleCatchupIterAdapter) NextIgnoringTime() {
	__antithesis_instrumentation__.Notify(113541)
	i.SimpleMVCCIterator.Next()
}

var _ simpleCatchupIter = simpleCatchupIterAdapter{}

func NewCatchUpIterator(
	reader storage.Reader, args *roachpb.RangeFeedRequest, useTBI bool, closer func(),
) *CatchUpIterator {
	__antithesis_instrumentation__.Notify(113542)
	ret := &CatchUpIterator{
		close: closer,
	}
	if useTBI {
		__antithesis_instrumentation__.Notify(113544)
		ret.simpleCatchupIter = storage.NewMVCCIncrementalIterator(reader, storage.MVCCIncrementalIterOptions{
			EnableTimeBoundIteratorOptimization: true,
			EndKey:                              args.Span.EndKey,

			StartTime: args.Timestamp.Prev(),
			EndTime:   hlc.MaxTimestamp,

			IntentPolicy: storage.MVCCIncrementalIterIntentPolicyEmit,

			InlinePolicy: storage.MVCCIncrementalIterInlinePolicyEmit,
		})
	} else {
		__antithesis_instrumentation__.Notify(113545)
		iter := reader.NewMVCCIterator(storage.MVCCKeyAndIntentsIterKind, storage.IterOptions{
			UpperBound: args.Span.EndKey,
		})
		ret.simpleCatchupIter = simpleCatchupIterAdapter{SimpleMVCCIterator: iter}
	}
	__antithesis_instrumentation__.Notify(113543)

	return ret
}

func (i *CatchUpIterator) Close() {
	__antithesis_instrumentation__.Notify(113546)
	i.simpleCatchupIter.Close()
	if i.close != nil {
		__antithesis_instrumentation__.Notify(113547)
		i.close()
	} else {
		__antithesis_instrumentation__.Notify(113548)
	}
}

type outputEventFn func(e *roachpb.RangeFeedEvent) error

func (i *CatchUpIterator) CatchUpScan(
	startKey, endKey storage.MVCCKey,
	catchUpTimestamp hlc.Timestamp,
	withDiff bool,
	outputFn outputEventFn,
) error {
	__antithesis_instrumentation__.Notify(113549)
	var a bufalloc.ByteAllocator

	var lastKey roachpb.Key
	reorderBuf := make([]roachpb.RangeFeedEvent, 0, 5)
	addPrevToLastEvent := func(val []byte) {
		__antithesis_instrumentation__.Notify(113553)
		if l := len(reorderBuf); l > 0 {
			__antithesis_instrumentation__.Notify(113554)
			if reorderBuf[l-1].Val.PrevValue.IsPresent() {
				__antithesis_instrumentation__.Notify(113556)
				panic("RangeFeedValue.PrevVal unexpectedly set")
			} else {
				__antithesis_instrumentation__.Notify(113557)
			}
			__antithesis_instrumentation__.Notify(113555)

			reorderBuf[l-1].Val.PrevValue.RawBytes = val
		} else {
			__antithesis_instrumentation__.Notify(113558)
		}
	}
	__antithesis_instrumentation__.Notify(113550)

	outputEvents := func() error {
		__antithesis_instrumentation__.Notify(113559)
		for i := len(reorderBuf) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(113561)
			e := reorderBuf[i]
			if err := outputFn(&e); err != nil {
				__antithesis_instrumentation__.Notify(113563)
				return err
			} else {
				__antithesis_instrumentation__.Notify(113564)
			}
			__antithesis_instrumentation__.Notify(113562)
			reorderBuf[i] = roachpb.RangeFeedEvent{}
		}
		__antithesis_instrumentation__.Notify(113560)
		reorderBuf = reorderBuf[:0]
		return nil
	}
	__antithesis_instrumentation__.Notify(113551)

	var meta enginepb.MVCCMetadata
	i.SeekGE(startKey)
	for {
		__antithesis_instrumentation__.Notify(113565)
		if ok, err := i.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(113571)
			return err
		} else {
			__antithesis_instrumentation__.Notify(113572)
			if !ok {
				__antithesis_instrumentation__.Notify(113573)
				break
			} else {
				__antithesis_instrumentation__.Notify(113574)
			}
		}
		__antithesis_instrumentation__.Notify(113566)

		unsafeKey := i.UnsafeKey()
		unsafeVal := i.UnsafeValue()
		if !unsafeKey.IsValue() {
			__antithesis_instrumentation__.Notify(113575)

			if err := protoutil.Unmarshal(unsafeVal, &meta); err != nil {
				__antithesis_instrumentation__.Notify(113578)
				return errors.Wrapf(err, "unmarshaling mvcc meta: %v", unsafeKey)
			} else {
				__antithesis_instrumentation__.Notify(113579)
			}
			__antithesis_instrumentation__.Notify(113576)
			if !meta.IsInline() {
				__antithesis_instrumentation__.Notify(113580)

				i.Next()
				if ok, err := i.Valid(); err != nil {
					__antithesis_instrumentation__.Notify(113583)
					return errors.Wrap(err, "iterating to provisional value for intent")
				} else {
					__antithesis_instrumentation__.Notify(113584)
					if !ok {
						__antithesis_instrumentation__.Notify(113585)
						return errors.Errorf("expected provisional value for intent")
					} else {
						__antithesis_instrumentation__.Notify(113586)
					}
				}
				__antithesis_instrumentation__.Notify(113581)
				if !meta.Timestamp.ToTimestamp().EqOrdering(i.UnsafeKey().Timestamp) {
					__antithesis_instrumentation__.Notify(113587)
					return errors.Errorf("expected provisional value for intent with ts %s, found %s",
						meta.Timestamp, i.UnsafeKey().Timestamp)
				} else {
					__antithesis_instrumentation__.Notify(113588)
				}
				__antithesis_instrumentation__.Notify(113582)
				i.Next()
				continue
			} else {
				__antithesis_instrumentation__.Notify(113589)
			}
			__antithesis_instrumentation__.Notify(113577)

			unsafeVal = meta.RawBytes
		} else {
			__antithesis_instrumentation__.Notify(113590)
		}
		__antithesis_instrumentation__.Notify(113567)

		ts := unsafeKey.Timestamp
		ignore := !(ts.IsEmpty() || func() bool {
			__antithesis_instrumentation__.Notify(113591)
			return catchUpTimestamp.Less(ts) == true
		}() == true)
		if ignore && func() bool {
			__antithesis_instrumentation__.Notify(113592)
			return !withDiff == true
		}() == true {
			__antithesis_instrumentation__.Notify(113593)

			i.NextKey()
			continue
		} else {
			__antithesis_instrumentation__.Notify(113594)
		}
		__antithesis_instrumentation__.Notify(113568)

		sameKey := bytes.Equal(unsafeKey.Key, lastKey)
		if !sameKey {
			__antithesis_instrumentation__.Notify(113595)

			if err := outputEvents(); err != nil {
				__antithesis_instrumentation__.Notify(113597)
				return err
			} else {
				__antithesis_instrumentation__.Notify(113598)
			}
			__antithesis_instrumentation__.Notify(113596)
			a, lastKey = a.Copy(unsafeKey.Key, 0)
		} else {
			__antithesis_instrumentation__.Notify(113599)
		}
		__antithesis_instrumentation__.Notify(113569)
		key := lastKey

		if !ignore || func() bool {
			__antithesis_instrumentation__.Notify(113600)
			return (withDiff && func() bool {
				__antithesis_instrumentation__.Notify(113601)
				return len(reorderBuf) > 0 == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(113602)
			var val []byte
			a, val = a.Copy(unsafeVal, 0)
			if withDiff {
				__antithesis_instrumentation__.Notify(113604)

				addPrevToLastEvent(val)
			} else {
				__antithesis_instrumentation__.Notify(113605)
			}
			__antithesis_instrumentation__.Notify(113603)

			if !ignore {
				__antithesis_instrumentation__.Notify(113606)

				var event roachpb.RangeFeedEvent
				event.MustSetValue(&roachpb.RangeFeedValue{
					Key: key,
					Value: roachpb.Value{
						RawBytes:  val,
						Timestamp: ts,
					},
				})
				reorderBuf = append(reorderBuf, event)
			} else {
				__antithesis_instrumentation__.Notify(113607)
			}
		} else {
			__antithesis_instrumentation__.Notify(113608)
		}
		__antithesis_instrumentation__.Notify(113570)

		if ignore {
			__antithesis_instrumentation__.Notify(113609)

			i.NextKey()
		} else {
			__antithesis_instrumentation__.Notify(113610)

			if withDiff {
				__antithesis_instrumentation__.Notify(113611)

				i.NextIgnoringTime()
			} else {
				__antithesis_instrumentation__.Notify(113612)
				i.Next()
			}
		}
	}
	__antithesis_instrumentation__.Notify(113552)

	return outputEvents()
}
