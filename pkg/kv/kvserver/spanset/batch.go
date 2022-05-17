package spanset

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/storage/enginepb"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/pebble"
)

type MVCCIterator struct {
	i     storage.MVCCIterator
	spans *SpanSet

	spansOnly bool

	ts hlc.Timestamp

	err error

	invalid bool
}

var _ storage.MVCCIterator = &MVCCIterator{}

func NewIterator(iter storage.MVCCIterator, spans *SpanSet) *MVCCIterator {
	__antithesis_instrumentation__.Notify(122914)
	return &MVCCIterator{i: iter, spans: spans, spansOnly: true}
}

func NewIteratorAt(iter storage.MVCCIterator, spans *SpanSet, ts hlc.Timestamp) *MVCCIterator {
	__antithesis_instrumentation__.Notify(122915)
	return &MVCCIterator{i: iter, spans: spans, ts: ts}
}

func (i *MVCCIterator) Close() {
	__antithesis_instrumentation__.Notify(122916)
	i.i.Close()
}

func (i *MVCCIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(122917)
	if i.err != nil {
		__antithesis_instrumentation__.Notify(122920)
		return false, i.err
	} else {
		__antithesis_instrumentation__.Notify(122921)
	}
	__antithesis_instrumentation__.Notify(122918)
	ok, err := i.i.Valid()
	if err != nil {
		__antithesis_instrumentation__.Notify(122922)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(122923)
	}
	__antithesis_instrumentation__.Notify(122919)
	return ok && func() bool {
		__antithesis_instrumentation__.Notify(122924)
		return !i.invalid == true
	}() == true, nil
}

func (i *MVCCIterator) SeekGE(key storage.MVCCKey) {
	__antithesis_instrumentation__.Notify(122925)
	i.i.SeekGE(key)
	i.checkAllowed(roachpb.Span{Key: key.Key}, true)
}

func (i *MVCCIterator) SeekIntentGE(key roachpb.Key, txnUUID uuid.UUID) {
	__antithesis_instrumentation__.Notify(122926)
	i.i.SeekIntentGE(key, txnUUID)
	i.checkAllowed(roachpb.Span{Key: key}, true)
}

func (i *MVCCIterator) SeekLT(key storage.MVCCKey) {
	__antithesis_instrumentation__.Notify(122927)
	i.i.SeekLT(key)

	i.checkAllowed(roachpb.Span{EndKey: key.Key}, true)
}

func (i *MVCCIterator) Next() {
	__antithesis_instrumentation__.Notify(122928)
	i.i.Next()
	i.checkAllowed(roachpb.Span{Key: i.UnsafeKey().Key}, false)
}

func (i *MVCCIterator) Prev() {
	__antithesis_instrumentation__.Notify(122929)
	i.i.Prev()
	i.checkAllowed(roachpb.Span{Key: i.UnsafeKey().Key}, false)
}

func (i *MVCCIterator) NextKey() {
	__antithesis_instrumentation__.Notify(122930)
	i.i.NextKey()
	i.checkAllowed(roachpb.Span{Key: i.UnsafeKey().Key}, false)
}

func (i *MVCCIterator) checkAllowed(span roachpb.Span, errIfDisallowed bool) {
	__antithesis_instrumentation__.Notify(122931)
	i.invalid = false
	i.err = nil
	if ok, _ := i.i.Valid(); !ok {
		__antithesis_instrumentation__.Notify(122934)

		return
	} else {
		__antithesis_instrumentation__.Notify(122935)
	}
	__antithesis_instrumentation__.Notify(122932)
	var err error
	if i.spansOnly {
		__antithesis_instrumentation__.Notify(122936)
		err = i.spans.CheckAllowed(SpanReadOnly, span)
	} else {
		__antithesis_instrumentation__.Notify(122937)
		err = i.spans.CheckAllowedAt(SpanReadOnly, span, i.ts)
	}
	__antithesis_instrumentation__.Notify(122933)
	if errIfDisallowed {
		__antithesis_instrumentation__.Notify(122938)
		i.err = err
	} else {
		__antithesis_instrumentation__.Notify(122939)
		i.invalid = err != nil
	}
}

func (i *MVCCIterator) Key() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(122940)
	return i.i.Key()
}

func (i *MVCCIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(122941)
	return i.i.Value()
}

func (i *MVCCIterator) ValueProto(msg protoutil.Message) error {
	__antithesis_instrumentation__.Notify(122942)
	return i.i.ValueProto(msg)
}

func (i *MVCCIterator) UnsafeKey() storage.MVCCKey {
	__antithesis_instrumentation__.Notify(122943)
	return i.i.UnsafeKey()
}

func (i *MVCCIterator) UnsafeRawKey() []byte {
	__antithesis_instrumentation__.Notify(122944)
	return i.i.UnsafeRawKey()
}

func (i *MVCCIterator) UnsafeRawMVCCKey() []byte {
	__antithesis_instrumentation__.Notify(122945)
	return i.i.UnsafeRawMVCCKey()
}

func (i *MVCCIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(122946)
	return i.i.UnsafeValue()
}

func (i *MVCCIterator) ComputeStats(
	start, end roachpb.Key, nowNanos int64,
) (enginepb.MVCCStats, error) {
	__antithesis_instrumentation__.Notify(122947)
	if i.spansOnly {
		__antithesis_instrumentation__.Notify(122949)
		if err := i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}); err != nil {
			__antithesis_instrumentation__.Notify(122950)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(122951)
		}
	} else {
		__antithesis_instrumentation__.Notify(122952)
		if err := i.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}, i.ts); err != nil {
			__antithesis_instrumentation__.Notify(122953)
			return enginepb.MVCCStats{}, err
		} else {
			__antithesis_instrumentation__.Notify(122954)
		}
	}
	__antithesis_instrumentation__.Notify(122948)
	return i.i.ComputeStats(start, end, nowNanos)
}

func (i *MVCCIterator) FindSplitKey(
	start, end, minSplitKey roachpb.Key, targetSize int64,
) (storage.MVCCKey, error) {
	__antithesis_instrumentation__.Notify(122955)
	if i.spansOnly {
		__antithesis_instrumentation__.Notify(122957)
		if err := i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}); err != nil {
			__antithesis_instrumentation__.Notify(122958)
			return storage.MVCCKey{}, err
		} else {
			__antithesis_instrumentation__.Notify(122959)
		}
	} else {
		__antithesis_instrumentation__.Notify(122960)
		if err := i.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}, i.ts); err != nil {
			__antithesis_instrumentation__.Notify(122961)
			return storage.MVCCKey{}, err
		} else {
			__antithesis_instrumentation__.Notify(122962)
		}
	}
	__antithesis_instrumentation__.Notify(122956)
	return i.i.FindSplitKey(start, end, minSplitKey, targetSize)
}

func (i *MVCCIterator) SetUpperBound(key roachpb.Key) {
	__antithesis_instrumentation__.Notify(122963)
	i.i.SetUpperBound(key)
}

func (i *MVCCIterator) Stats() storage.IteratorStats {
	__antithesis_instrumentation__.Notify(122964)
	return i.i.Stats()
}

func (i *MVCCIterator) SupportsPrev() bool {
	__antithesis_instrumentation__.Notify(122965)
	return i.i.SupportsPrev()
}

type EngineIterator struct {
	i         storage.EngineIterator
	spans     *SpanSet
	spansOnly bool
	ts        hlc.Timestamp
}

func (i *EngineIterator) Close() {
	__antithesis_instrumentation__.Notify(122966)
	i.i.Close()
}

func (i *EngineIterator) SeekEngineKeyGE(key storage.EngineKey) (valid bool, err error) {
	__antithesis_instrumentation__.Notify(122967)
	valid, err = i.i.SeekEngineKeyGE(key)
	if !valid {
		__antithesis_instrumentation__.Notify(122970)
		return valid, err
	} else {
		__antithesis_instrumentation__.Notify(122971)
	}
	__antithesis_instrumentation__.Notify(122968)
	if key.IsMVCCKey() && func() bool {
		__antithesis_instrumentation__.Notify(122972)
		return !i.spansOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(122973)
		mvccKey, _ := key.ToMVCCKey()
		if err := i.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: mvccKey.Key}, i.ts); err != nil {
			__antithesis_instrumentation__.Notify(122974)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(122975)
		}
	} else {
		__antithesis_instrumentation__.Notify(122976)
		if err = i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(122977)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(122978)
		}
	}
	__antithesis_instrumentation__.Notify(122969)
	return valid, err
}

func (i *EngineIterator) SeekEngineKeyLT(key storage.EngineKey) (valid bool, err error) {
	__antithesis_instrumentation__.Notify(122979)
	valid, err = i.i.SeekEngineKeyLT(key)
	if !valid {
		__antithesis_instrumentation__.Notify(122982)
		return valid, err
	} else {
		__antithesis_instrumentation__.Notify(122983)
	}
	__antithesis_instrumentation__.Notify(122980)
	if key.IsMVCCKey() && func() bool {
		__antithesis_instrumentation__.Notify(122984)
		return !i.spansOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(122985)
		mvccKey, _ := key.ToMVCCKey()
		if err := i.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: mvccKey.Key}, i.ts); err != nil {
			__antithesis_instrumentation__.Notify(122986)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(122987)
		}
	} else {
		__antithesis_instrumentation__.Notify(122988)
		if err = i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{EndKey: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(122989)
			return false, err
		} else {
			__antithesis_instrumentation__.Notify(122990)
		}
	}
	__antithesis_instrumentation__.Notify(122981)
	return valid, err
}

func (i *EngineIterator) NextEngineKey() (valid bool, err error) {
	__antithesis_instrumentation__.Notify(122991)
	valid, err = i.i.NextEngineKey()
	if !valid {
		__antithesis_instrumentation__.Notify(122993)
		return valid, err
	} else {
		__antithesis_instrumentation__.Notify(122994)
	}
	__antithesis_instrumentation__.Notify(122992)
	return i.checkKeyAllowed()
}

func (i *EngineIterator) PrevEngineKey() (valid bool, err error) {
	__antithesis_instrumentation__.Notify(122995)
	valid, err = i.i.PrevEngineKey()
	if !valid {
		__antithesis_instrumentation__.Notify(122997)
		return valid, err
	} else {
		__antithesis_instrumentation__.Notify(122998)
	}
	__antithesis_instrumentation__.Notify(122996)
	return i.checkKeyAllowed()
}

func (i *EngineIterator) SeekEngineKeyGEWithLimit(
	key storage.EngineKey, limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(122999)
	state, err = i.i.SeekEngineKeyGEWithLimit(key, limit)
	if state != pebble.IterValid {
		__antithesis_instrumentation__.Notify(123002)
		return state, err
	} else {
		__antithesis_instrumentation__.Notify(123003)
	}
	__antithesis_instrumentation__.Notify(123000)
	if err = i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: key.Key}); err != nil {
		__antithesis_instrumentation__.Notify(123004)
		return pebble.IterExhausted, err
	} else {
		__antithesis_instrumentation__.Notify(123005)
	}
	__antithesis_instrumentation__.Notify(123001)
	return state, err
}

func (i *EngineIterator) SeekEngineKeyLTWithLimit(
	key storage.EngineKey, limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(123006)
	state, err = i.i.SeekEngineKeyLTWithLimit(key, limit)
	if state != pebble.IterValid {
		__antithesis_instrumentation__.Notify(123009)
		return state, err
	} else {
		__antithesis_instrumentation__.Notify(123010)
	}
	__antithesis_instrumentation__.Notify(123007)
	if err = i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{EndKey: key.Key}); err != nil {
		__antithesis_instrumentation__.Notify(123011)
		return pebble.IterExhausted, err
	} else {
		__antithesis_instrumentation__.Notify(123012)
	}
	__antithesis_instrumentation__.Notify(123008)
	return state, err
}

func (i *EngineIterator) NextEngineKeyWithLimit(
	limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(123013)
	return i.i.NextEngineKeyWithLimit(limit)
}

func (i *EngineIterator) PrevEngineKeyWithLimit(
	limit roachpb.Key,
) (state pebble.IterValidityState, err error) {
	__antithesis_instrumentation__.Notify(123014)
	return i.i.PrevEngineKeyWithLimit(limit)
}

func (i *EngineIterator) checkKeyAllowed() (valid bool, err error) {
	__antithesis_instrumentation__.Notify(123015)
	key, err := i.i.UnsafeEngineKey()
	if err != nil {
		__antithesis_instrumentation__.Notify(123018)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(123019)
	}
	__antithesis_instrumentation__.Notify(123016)
	if key.IsMVCCKey() && func() bool {
		__antithesis_instrumentation__.Notify(123020)
		return !i.spansOnly == true
	}() == true {
		__antithesis_instrumentation__.Notify(123021)
		mvccKey, _ := key.ToMVCCKey()
		if err := i.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: mvccKey.Key}, i.ts); err != nil {
			__antithesis_instrumentation__.Notify(123022)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(123023)
		}
	} else {
		__antithesis_instrumentation__.Notify(123024)
		if err = i.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(123025)

			return false, nil
		} else {
			__antithesis_instrumentation__.Notify(123026)
		}
	}
	__antithesis_instrumentation__.Notify(123017)
	return true, nil
}

func (i *EngineIterator) UnsafeEngineKey() (storage.EngineKey, error) {
	__antithesis_instrumentation__.Notify(123027)
	return i.i.UnsafeEngineKey()
}

func (i *EngineIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(123028)
	return i.i.UnsafeValue()
}

func (i *EngineIterator) EngineKey() (storage.EngineKey, error) {
	__antithesis_instrumentation__.Notify(123029)
	return i.i.EngineKey()
}

func (i *EngineIterator) Value() []byte {
	__antithesis_instrumentation__.Notify(123030)
	return i.i.Value()
}

func (i *EngineIterator) UnsafeRawEngineKey() []byte {
	__antithesis_instrumentation__.Notify(123031)
	return i.i.UnsafeRawEngineKey()
}

func (i *EngineIterator) SetUpperBound(key roachpb.Key) {
	__antithesis_instrumentation__.Notify(123032)
	i.i.SetUpperBound(key)
}

func (i *EngineIterator) GetRawIter() *pebble.Iterator {
	__antithesis_instrumentation__.Notify(123033)
	return i.i.GetRawIter()
}

func (i *EngineIterator) Stats() storage.IteratorStats {
	__antithesis_instrumentation__.Notify(123034)
	return i.i.Stats()
}

type spanSetReader struct {
	r     storage.Reader
	spans *SpanSet

	spansOnly bool
	ts        hlc.Timestamp
}

var _ storage.Reader = spanSetReader{}

func (s spanSetReader) Close() {
	__antithesis_instrumentation__.Notify(123035)
	s.r.Close()
}

func (s spanSetReader) Closed() bool {
	__antithesis_instrumentation__.Notify(123036)
	return s.r.Closed()
}

func (s spanSetReader) ExportMVCCToSst(
	ctx context.Context, exportOptions storage.ExportOptions, dest io.Writer,
) (roachpb.BulkOpSummary, roachpb.Key, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(123037)
	return s.r.ExportMVCCToSst(ctx, exportOptions, dest)
}

func (s spanSetReader) MVCCGet(key storage.MVCCKey) ([]byte, error) {
	__antithesis_instrumentation__.Notify(123038)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123040)
		if err := s.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(123041)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(123042)
		}
	} else {
		__antithesis_instrumentation__.Notify(123043)
		if err := s.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: key.Key}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123044)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(123045)
		}
	}
	__antithesis_instrumentation__.Notify(123039)

	return s.r.MVCCGet(key)
}

func (s spanSetReader) MVCCGetProto(
	key storage.MVCCKey, msg protoutil.Message,
) (bool, int64, int64, error) {
	__antithesis_instrumentation__.Notify(123046)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123048)
		if err := s.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(123049)
			return false, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(123050)
		}
	} else {
		__antithesis_instrumentation__.Notify(123051)
		if err := s.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: key.Key}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123052)
			return false, 0, 0, err
		} else {
			__antithesis_instrumentation__.Notify(123053)
		}
	}
	__antithesis_instrumentation__.Notify(123047)

	return s.r.MVCCGetProto(key, msg)
}

func (s spanSetReader) MVCCIterate(
	start, end roachpb.Key, iterKind storage.MVCCIterKind, f func(storage.MVCCKeyValue) error,
) error {
	__antithesis_instrumentation__.Notify(123054)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123056)
		if err := s.spans.CheckAllowed(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}); err != nil {
			__antithesis_instrumentation__.Notify(123057)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123058)
		}
	} else {
		__antithesis_instrumentation__.Notify(123059)
		if err := s.spans.CheckAllowedAt(SpanReadOnly, roachpb.Span{Key: start, EndKey: end}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123060)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123061)
		}
	}
	__antithesis_instrumentation__.Notify(123055)
	return s.r.MVCCIterate(start, end, iterKind, f)
}

func (s spanSetReader) NewMVCCIterator(
	iterKind storage.MVCCIterKind, opts storage.IterOptions,
) storage.MVCCIterator {
	__antithesis_instrumentation__.Notify(123062)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123064)
		return NewIterator(s.r.NewMVCCIterator(iterKind, opts), s.spans)
	} else {
		__antithesis_instrumentation__.Notify(123065)
	}
	__antithesis_instrumentation__.Notify(123063)
	return NewIteratorAt(s.r.NewMVCCIterator(iterKind, opts), s.spans, s.ts)
}

func (s spanSetReader) NewEngineIterator(opts storage.IterOptions) storage.EngineIterator {
	__antithesis_instrumentation__.Notify(123066)
	if !s.spansOnly {
		__antithesis_instrumentation__.Notify(123068)
		log.Warningf(context.Background(),
			"cannot do strict timestamp checking of EngineIterator, resorting to best effort")
	} else {
		__antithesis_instrumentation__.Notify(123069)
	}
	__antithesis_instrumentation__.Notify(123067)
	return &EngineIterator{
		i:         s.r.NewEngineIterator(opts),
		spans:     s.spans,
		spansOnly: s.spansOnly,
		ts:        s.ts,
	}
}

func (s spanSetReader) ConsistentIterators() bool {
	__antithesis_instrumentation__.Notify(123070)
	return s.r.ConsistentIterators()
}

func (s spanSetReader) PinEngineStateForIterators() error {
	__antithesis_instrumentation__.Notify(123071)
	return s.r.PinEngineStateForIterators()
}

type spanSetWriter struct {
	w     storage.Writer
	spans *SpanSet

	spansOnly bool
	ts        hlc.Timestamp
}

var _ storage.Writer = spanSetWriter{}

func (s spanSetWriter) ApplyBatchRepr(repr []byte, sync bool) error {
	__antithesis_instrumentation__.Notify(123072)

	return s.w.ApplyBatchRepr(repr, sync)
}

func (s spanSetWriter) checkAllowed(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123073)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123075)
		if err := s.spans.CheckAllowed(SpanReadWrite, roachpb.Span{Key: key}); err != nil {
			__antithesis_instrumentation__.Notify(123076)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123077)
		}
	} else {
		__antithesis_instrumentation__.Notify(123078)
		if err := s.spans.CheckAllowedAt(SpanReadWrite, roachpb.Span{Key: key}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123079)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123080)
		}
	}
	__antithesis_instrumentation__.Notify(123074)
	return nil
}

func (s spanSetWriter) ClearMVCC(key storage.MVCCKey) error {
	__antithesis_instrumentation__.Notify(123081)
	if err := s.checkAllowed(key.Key); err != nil {
		__antithesis_instrumentation__.Notify(123083)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123084)
	}
	__antithesis_instrumentation__.Notify(123082)
	return s.w.ClearMVCC(key)
}

func (s spanSetWriter) ClearUnversioned(key roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123085)
	if err := s.checkAllowed(key); err != nil {
		__antithesis_instrumentation__.Notify(123087)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123088)
	}
	__antithesis_instrumentation__.Notify(123086)
	return s.w.ClearUnversioned(key)
}

func (s spanSetWriter) ClearIntent(
	key roachpb.Key, txnDidNotUpdateMeta bool, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(123089)
	if err := s.checkAllowed(key); err != nil {
		__antithesis_instrumentation__.Notify(123091)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123092)
	}
	__antithesis_instrumentation__.Notify(123090)
	return s.w.ClearIntent(key, txnDidNotUpdateMeta, txnUUID)
}

func (s spanSetWriter) ClearEngineKey(key storage.EngineKey) error {
	__antithesis_instrumentation__.Notify(123093)
	if !s.spansOnly {
		__antithesis_instrumentation__.Notify(123096)
		panic("cannot do timestamp checking for clearing EngineKey")
	} else {
		__antithesis_instrumentation__.Notify(123097)
	}
	__antithesis_instrumentation__.Notify(123094)
	if err := s.spans.CheckAllowed(SpanReadWrite, roachpb.Span{Key: key.Key}); err != nil {
		__antithesis_instrumentation__.Notify(123098)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123099)
	}
	__antithesis_instrumentation__.Notify(123095)
	return s.w.ClearEngineKey(key)
}

func (s spanSetWriter) SingleClearEngineKey(key storage.EngineKey) error {
	__antithesis_instrumentation__.Notify(123100)

	return s.w.SingleClearEngineKey(key)
}

func (s spanSetWriter) checkAllowedRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123101)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123103)
		if err := s.spans.CheckAllowed(SpanReadWrite, roachpb.Span{Key: start, EndKey: end}); err != nil {
			__antithesis_instrumentation__.Notify(123104)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123105)
		}
	} else {
		__antithesis_instrumentation__.Notify(123106)
		if err := s.spans.CheckAllowedAt(SpanReadWrite, roachpb.Span{Key: start, EndKey: end}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123107)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123108)
		}
	}
	__antithesis_instrumentation__.Notify(123102)
	return nil
}

func (s spanSetWriter) ClearRawRange(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123109)
	if err := s.checkAllowedRange(start, end); err != nil {
		__antithesis_instrumentation__.Notify(123111)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123112)
	}
	__antithesis_instrumentation__.Notify(123110)
	return s.w.ClearRawRange(start, end)
}

func (s spanSetWriter) ClearMVCCRangeAndIntents(start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123113)
	if err := s.checkAllowedRange(start, end); err != nil {
		__antithesis_instrumentation__.Notify(123115)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123116)
	}
	__antithesis_instrumentation__.Notify(123114)
	return s.w.ClearMVCCRangeAndIntents(start, end)
}

func (s spanSetWriter) ClearMVCCRange(start, end storage.MVCCKey) error {
	__antithesis_instrumentation__.Notify(123117)
	if err := s.checkAllowedRange(start.Key, end.Key); err != nil {
		__antithesis_instrumentation__.Notify(123119)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123120)
	}
	__antithesis_instrumentation__.Notify(123118)
	return s.w.ClearMVCCRange(start, end)
}

func (s spanSetWriter) ClearIterRange(iter storage.MVCCIterator, start, end roachpb.Key) error {
	__antithesis_instrumentation__.Notify(123121)
	if err := s.checkAllowedRange(start, end); err != nil {
		__antithesis_instrumentation__.Notify(123123)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123124)
	}
	__antithesis_instrumentation__.Notify(123122)
	return s.w.ClearIterRange(iter, start, end)
}

func (s spanSetWriter) Merge(key storage.MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(123125)
	if s.spansOnly {
		__antithesis_instrumentation__.Notify(123127)
		if err := s.spans.CheckAllowed(SpanReadWrite, roachpb.Span{Key: key.Key}); err != nil {
			__antithesis_instrumentation__.Notify(123128)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123129)
		}
	} else {
		__antithesis_instrumentation__.Notify(123130)
		if err := s.spans.CheckAllowedAt(SpanReadWrite, roachpb.Span{Key: key.Key}, s.ts); err != nil {
			__antithesis_instrumentation__.Notify(123131)
			return err
		} else {
			__antithesis_instrumentation__.Notify(123132)
		}
	}
	__antithesis_instrumentation__.Notify(123126)
	return s.w.Merge(key, value)
}

func (s spanSetWriter) PutMVCC(key storage.MVCCKey, value []byte) error {
	__antithesis_instrumentation__.Notify(123133)
	if err := s.checkAllowed(key.Key); err != nil {
		__antithesis_instrumentation__.Notify(123135)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123136)
	}
	__antithesis_instrumentation__.Notify(123134)
	return s.w.PutMVCC(key, value)
}

func (s spanSetWriter) PutUnversioned(key roachpb.Key, value []byte) error {
	__antithesis_instrumentation__.Notify(123137)
	if err := s.checkAllowed(key); err != nil {
		__antithesis_instrumentation__.Notify(123139)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123140)
	}
	__antithesis_instrumentation__.Notify(123138)
	return s.w.PutUnversioned(key, value)
}

func (s spanSetWriter) PutIntent(
	ctx context.Context, key roachpb.Key, value []byte, txnUUID uuid.UUID,
) error {
	__antithesis_instrumentation__.Notify(123141)
	if err := s.checkAllowed(key); err != nil {
		__antithesis_instrumentation__.Notify(123143)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123144)
	}
	__antithesis_instrumentation__.Notify(123142)
	return s.w.PutIntent(ctx, key, value, txnUUID)
}

func (s spanSetWriter) PutEngineKey(key storage.EngineKey, value []byte) error {
	__antithesis_instrumentation__.Notify(123145)
	if !s.spansOnly {
		__antithesis_instrumentation__.Notify(123148)
		panic("cannot do timestamp checking for putting EngineKey")
	} else {
		__antithesis_instrumentation__.Notify(123149)
	}
	__antithesis_instrumentation__.Notify(123146)
	if err := s.spans.CheckAllowed(SpanReadWrite, roachpb.Span{Key: key.Key}); err != nil {
		__antithesis_instrumentation__.Notify(123150)
		return err
	} else {
		__antithesis_instrumentation__.Notify(123151)
	}
	__antithesis_instrumentation__.Notify(123147)
	return s.w.PutEngineKey(key, value)
}

func (s spanSetWriter) LogData(data []byte) error {
	__antithesis_instrumentation__.Notify(123152)
	return s.w.LogData(data)
}

func (s spanSetWriter) LogLogicalOp(
	op storage.MVCCLogicalOpType, details storage.MVCCLogicalOpDetails,
) {
	__antithesis_instrumentation__.Notify(123153)
	s.w.LogLogicalOp(op, details)
}

type ReadWriter struct {
	spanSetReader
	spanSetWriter
}

var _ storage.ReadWriter = ReadWriter{}

func makeSpanSetReadWriter(rw storage.ReadWriter, spans *SpanSet) ReadWriter {
	__antithesis_instrumentation__.Notify(123154)
	spans = addLockTableSpans(spans)
	return ReadWriter{
		spanSetReader: spanSetReader{r: rw, spans: spans, spansOnly: true},
		spanSetWriter: spanSetWriter{w: rw, spans: spans, spansOnly: true},
	}
}

func makeSpanSetReadWriterAt(rw storage.ReadWriter, spans *SpanSet, ts hlc.Timestamp) ReadWriter {
	__antithesis_instrumentation__.Notify(123155)
	spans = addLockTableSpans(spans)
	return ReadWriter{
		spanSetReader: spanSetReader{r: rw, spans: spans, ts: ts},
		spanSetWriter: spanSetWriter{w: rw, spans: spans, ts: ts},
	}
}

func NewReadWriterAt(rw storage.ReadWriter, spans *SpanSet, ts hlc.Timestamp) storage.ReadWriter {
	__antithesis_instrumentation__.Notify(123156)
	return makeSpanSetReadWriterAt(rw, spans, ts)
}

type spanSetBatch struct {
	ReadWriter
	b     storage.Batch
	spans *SpanSet

	spansOnly bool
	ts        hlc.Timestamp
}

var _ storage.Batch = spanSetBatch{}

func (s spanSetBatch) Commit(sync bool) error {
	__antithesis_instrumentation__.Notify(123157)
	return s.b.Commit(sync)
}

func (s spanSetBatch) Empty() bool {
	__antithesis_instrumentation__.Notify(123158)
	return s.b.Empty()
}

func (s spanSetBatch) Count() uint32 {
	__antithesis_instrumentation__.Notify(123159)
	return s.b.Count()
}

func (s spanSetBatch) Len() int {
	__antithesis_instrumentation__.Notify(123160)
	return s.b.Len()
}

func (s spanSetBatch) Repr() []byte {
	__antithesis_instrumentation__.Notify(123161)
	return s.b.Repr()
}

func NewBatch(b storage.Batch, spans *SpanSet) storage.Batch {
	__antithesis_instrumentation__.Notify(123162)
	return &spanSetBatch{
		ReadWriter: makeSpanSetReadWriter(b, spans),
		b:          b,
		spans:      spans,
		spansOnly:  true,
	}
}

func NewBatchAt(b storage.Batch, spans *SpanSet, ts hlc.Timestamp) storage.Batch {
	__antithesis_instrumentation__.Notify(123163)
	return &spanSetBatch{
		ReadWriter: makeSpanSetReadWriterAt(b, spans, ts),
		b:          b,
		spans:      spans,
		ts:         ts,
	}
}

func DisableReaderAssertions(reader storage.Reader) storage.Reader {
	__antithesis_instrumentation__.Notify(123164)
	switch v := reader.(type) {
	case ReadWriter:
		__antithesis_instrumentation__.Notify(123165)
		return DisableReaderAssertions(v.r)
	case *spanSetBatch:
		__antithesis_instrumentation__.Notify(123166)
		return DisableReaderAssertions(v.r)
	default:
		__antithesis_instrumentation__.Notify(123167)
		return reader
	}
}

func addLockTableSpans(spans *SpanSet) *SpanSet {
	__antithesis_instrumentation__.Notify(123168)
	withLocks := spans.Copy()
	spans.Iterate(func(sa SpanAccess, _ SpanScope, span Span) {
		__antithesis_instrumentation__.Notify(123170)

		if bytes.HasPrefix(span.Key, keys.LocalRangeLockTablePrefix) || func() bool {
			__antithesis_instrumentation__.Notify(123173)
			return bytes.HasPrefix(span.EndKey, keys.LocalRangeLockTablePrefix) == true
		}() == true {
			__antithesis_instrumentation__.Notify(123174)
			panic(fmt.Sprintf(
				"declaring raw lock table spans is illegal, use main key spans instead (found %s)", span))
		} else {
			__antithesis_instrumentation__.Notify(123175)
		}
		__antithesis_instrumentation__.Notify(123171)
		ltKey, _ := keys.LockTableSingleKey(span.Key, nil)
		var ltEndKey roachpb.Key
		if span.EndKey != nil {
			__antithesis_instrumentation__.Notify(123176)
			ltEndKey, _ = keys.LockTableSingleKey(span.EndKey, nil)
		} else {
			__antithesis_instrumentation__.Notify(123177)
		}
		__antithesis_instrumentation__.Notify(123172)
		withLocks.AddNonMVCC(sa, roachpb.Span{Key: ltKey, EndKey: ltEndKey})
	})
	__antithesis_instrumentation__.Notify(123169)
	return withLocks
}
