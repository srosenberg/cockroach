package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/sstable"
	"github.com/cockroachdb/pebble/vfs"
)

type sstIterator struct {
	sst  *sstable.Reader
	iter sstable.Iterator

	mvccKey   MVCCKey
	value     []byte
	iterValid bool
	err       error

	keyBuf []byte

	verify bool

	prevSeekKey  MVCCKey
	seekGELastOp bool
}

func NewSSTIterator(file sstable.ReadableFile) (SimpleMVCCIterator, error) {
	__antithesis_instrumentation__.Notify(643801)
	sst, err := sstable.NewReader(file, sstable.ReaderOptions{
		Comparer: EngineComparer,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(643803)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(643804)
	}
	__antithesis_instrumentation__.Notify(643802)
	return &sstIterator{sst: sst}, nil
}

func NewMemSSTIterator(data []byte, verify bool) (SimpleMVCCIterator, error) {
	__antithesis_instrumentation__.Notify(643805)
	sst, err := sstable.NewReader(vfs.NewMemFile(data), sstable.ReaderOptions{
		Comparer: EngineComparer,
	})
	if err != nil {
		__antithesis_instrumentation__.Notify(643807)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(643808)
	}
	__antithesis_instrumentation__.Notify(643806)
	return &sstIterator{sst: sst, verify: verify}, nil
}

func (r *sstIterator) Close() {
	__antithesis_instrumentation__.Notify(643809)
	if r.iter != nil {
		__antithesis_instrumentation__.Notify(643811)
		r.err = errors.Wrap(r.iter.Close(), "closing sstable iterator")
	} else {
		__antithesis_instrumentation__.Notify(643812)
	}
	__antithesis_instrumentation__.Notify(643810)
	if err := r.sst.Close(); err != nil && func() bool {
		__antithesis_instrumentation__.Notify(643813)
		return r.err == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(643814)
		r.err = errors.Wrap(err, "closing sstable")
	} else {
		__antithesis_instrumentation__.Notify(643815)
	}
}

func (r *sstIterator) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(643816)
	if r.err != nil {
		__antithesis_instrumentation__.Notify(643822)
		return
	} else {
		__antithesis_instrumentation__.Notify(643823)
	}
	__antithesis_instrumentation__.Notify(643817)
	if r.iter == nil {
		__antithesis_instrumentation__.Notify(643824)

		r.iter, r.err = r.sst.NewIter(nil, nil)
		if r.err != nil {
			__antithesis_instrumentation__.Notify(643825)
			return
		} else {
			__antithesis_instrumentation__.Notify(643826)
		}
	} else {
		__antithesis_instrumentation__.Notify(643827)
	}
	__antithesis_instrumentation__.Notify(643818)
	r.keyBuf = EncodeMVCCKeyToBuf(r.keyBuf, key)
	var iKey *sstable.InternalKey
	trySeekUsingNext := false
	if r.seekGELastOp {
		__antithesis_instrumentation__.Notify(643828)

		trySeekUsingNext = !key.Less(r.prevSeekKey)
	} else {
		__antithesis_instrumentation__.Notify(643829)
	}
	__antithesis_instrumentation__.Notify(643819)

	iKey, r.value = r.iter.SeekGE(r.keyBuf, trySeekUsingNext)
	if iKey != nil {
		__antithesis_instrumentation__.Notify(643830)
		r.iterValid = true
		r.mvccKey, r.err = DecodeMVCCKey(iKey.UserKey)
	} else {
		__antithesis_instrumentation__.Notify(643831)
		r.iterValid = false
		r.err = r.iter.Error()
	}
	__antithesis_instrumentation__.Notify(643820)
	if r.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(643832)
		return r.err == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643833)
		return r.verify == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643834)
		return r.mvccKey.IsValue() == true
	}() == true {
		__antithesis_instrumentation__.Notify(643835)
		r.err = roachpb.Value{RawBytes: r.value}.Verify(r.mvccKey.Key)
	} else {
		__antithesis_instrumentation__.Notify(643836)
	}
	__antithesis_instrumentation__.Notify(643821)
	r.prevSeekKey.Key = append(r.prevSeekKey.Key[:0], r.mvccKey.Key...)
	r.prevSeekKey.Timestamp = r.mvccKey.Timestamp
	r.seekGELastOp = true
}

func (r *sstIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(643837)
	return r.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(643838)
		return r.err == nil == true
	}() == true, r.err
}

func (r *sstIterator) Next() {
	__antithesis_instrumentation__.Notify(643839)
	r.seekGELastOp = false
	if !r.iterValid || func() bool {
		__antithesis_instrumentation__.Notify(643842)
		return r.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(643843)
		return
	} else {
		__antithesis_instrumentation__.Notify(643844)
	}
	__antithesis_instrumentation__.Notify(643840)
	var iKey *sstable.InternalKey
	iKey, r.value = r.iter.Next()
	if iKey != nil {
		__antithesis_instrumentation__.Notify(643845)
		r.mvccKey, r.err = DecodeMVCCKey(iKey.UserKey)
	} else {
		__antithesis_instrumentation__.Notify(643846)
		r.iterValid = false
		r.err = r.iter.Error()
	}
	__antithesis_instrumentation__.Notify(643841)
	if r.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(643847)
		return r.err == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643848)
		return r.verify == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643849)
		return r.mvccKey.IsValue() == true
	}() == true {
		__antithesis_instrumentation__.Notify(643850)
		r.err = roachpb.Value{RawBytes: r.value}.Verify(r.mvccKey.Key)
	} else {
		__antithesis_instrumentation__.Notify(643851)
	}
}

func (r *sstIterator) NextKey() {
	__antithesis_instrumentation__.Notify(643852)
	r.seekGELastOp = false
	if !r.iterValid || func() bool {
		__antithesis_instrumentation__.Notify(643854)
		return r.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(643855)
		return
	} else {
		__antithesis_instrumentation__.Notify(643856)
	}
	__antithesis_instrumentation__.Notify(643853)
	r.keyBuf = append(r.keyBuf[:0], r.mvccKey.Key...)
	for r.Next(); r.iterValid && func() bool {
		__antithesis_instrumentation__.Notify(643857)
		return r.err == nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(643858)
		return bytes.Equal(r.keyBuf, r.mvccKey.Key) == true
	}() == true; r.Next() {
		__antithesis_instrumentation__.Notify(643859)
	}
}

func (r *sstIterator) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(643860)
	return r.mvccKey
}

func (r *sstIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(643861)
	return r.value
}
