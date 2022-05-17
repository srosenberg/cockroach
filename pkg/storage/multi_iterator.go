package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"

	"github.com/cockroachdb/cockroach/pkg/keys"
)

const invalidIdxSentinel = -1

type multiIterator struct {
	iters []SimpleMVCCIterator

	currentIdx int

	itersWithCurrentKey []int

	itersWithCurrentKeyTimestamp []int

	err error
}

var _ SimpleMVCCIterator = &multiIterator{}

func MakeMultiIterator(iters []SimpleMVCCIterator) SimpleMVCCIterator {
	__antithesis_instrumentation__.Notify(640253)
	return &multiIterator{
		iters:                        iters,
		currentIdx:                   invalidIdxSentinel,
		itersWithCurrentKey:          make([]int, 0, len(iters)),
		itersWithCurrentKeyTimestamp: make([]int, 0, len(iters)),
	}
}

func (f *multiIterator) Close() {
	__antithesis_instrumentation__.Notify(640254)

}

func (f *multiIterator) SeekGE(key MVCCKey) {
	__antithesis_instrumentation__.Notify(640255)
	for _, iter := range f.iters {
		__antithesis_instrumentation__.Notify(640257)
		iter.SeekGE(key)
	}
	__antithesis_instrumentation__.Notify(640256)

	f.advance()
}

func (f *multiIterator) Valid() (bool, error) {
	__antithesis_instrumentation__.Notify(640258)
	valid := f.currentIdx != invalidIdxSentinel && func() bool {
		__antithesis_instrumentation__.Notify(640259)
		return f.err == nil == true
	}() == true
	return valid, f.err
}

func (f *multiIterator) UnsafeKey() MVCCKey {
	__antithesis_instrumentation__.Notify(640260)
	return f.iters[f.currentIdx].UnsafeKey()
}

func (f *multiIterator) UnsafeValue() []byte {
	__antithesis_instrumentation__.Notify(640261)
	return f.iters[f.currentIdx].UnsafeValue()
}

func (f *multiIterator) Next() {
	__antithesis_instrumentation__.Notify(640262)

	for _, iterIdx := range f.itersWithCurrentKeyTimestamp {
		__antithesis_instrumentation__.Notify(640264)
		f.iters[iterIdx].Next()
	}
	__antithesis_instrumentation__.Notify(640263)
	f.advance()
}

func (f *multiIterator) NextKey() {
	__antithesis_instrumentation__.Notify(640265)

	for _, iterIdx := range f.itersWithCurrentKey {
		__antithesis_instrumentation__.Notify(640267)
		f.iters[iterIdx].NextKey()
	}
	__antithesis_instrumentation__.Notify(640266)
	f.advance()
}

func (f *multiIterator) advance() {
	__antithesis_instrumentation__.Notify(640268)

	proposedNextIdx := invalidIdxSentinel
	for iterIdx, iter := range f.iters {
		__antithesis_instrumentation__.Notify(640270)

		if ok, err := iter.Valid(); err != nil {
			__antithesis_instrumentation__.Notify(640273)
			f.err = err
			return
		} else {
			__antithesis_instrumentation__.Notify(640274)
			if !ok {
				__antithesis_instrumentation__.Notify(640275)
				continue
			} else {
				__antithesis_instrumentation__.Notify(640276)
			}
		}
		__antithesis_instrumentation__.Notify(640271)

		proposedMVCCKey := MVCCKey{Key: keys.MaxKey}
		if proposedNextIdx != invalidIdxSentinel {
			__antithesis_instrumentation__.Notify(640277)
			proposedMVCCKey = f.iters[proposedNextIdx].UnsafeKey()
		} else {
			__antithesis_instrumentation__.Notify(640278)
		}
		__antithesis_instrumentation__.Notify(640272)

		iterMVCCKey := iter.UnsafeKey()
		if cmp := bytes.Compare(iterMVCCKey.Key, proposedMVCCKey.Key); cmp < 0 {
			__antithesis_instrumentation__.Notify(640279)

			f.itersWithCurrentKey = f.itersWithCurrentKey[:0]
			f.itersWithCurrentKey = append(f.itersWithCurrentKey, iterIdx)
			f.itersWithCurrentKeyTimestamp = f.itersWithCurrentKeyTimestamp[:0]
			f.itersWithCurrentKeyTimestamp = append(f.itersWithCurrentKeyTimestamp, iterIdx)
			proposedNextIdx = iterIdx
		} else {
			__antithesis_instrumentation__.Notify(640280)
			if cmp == 0 {
				__antithesis_instrumentation__.Notify(640281)

				f.itersWithCurrentKey = append(f.itersWithCurrentKey, iterIdx)
				if proposedMVCCKey.Timestamp.EqOrdering(iterMVCCKey.Timestamp) {
					__antithesis_instrumentation__.Notify(640282)

					f.itersWithCurrentKeyTimestamp = append(f.itersWithCurrentKeyTimestamp, iterIdx)
					proposedNextIdx = iterIdx
				} else {
					__antithesis_instrumentation__.Notify(640283)
					if iterMVCCKey.Less(proposedMVCCKey) {
						__antithesis_instrumentation__.Notify(640284)

						f.itersWithCurrentKeyTimestamp = f.itersWithCurrentKeyTimestamp[:0]
						f.itersWithCurrentKeyTimestamp = append(f.itersWithCurrentKeyTimestamp, iterIdx)
						proposedNextIdx = iterIdx
					} else {
						__antithesis_instrumentation__.Notify(640285)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(640286)
			}
		}
	}
	__antithesis_instrumentation__.Notify(640269)

	f.currentIdx = proposedNextIdx
}
