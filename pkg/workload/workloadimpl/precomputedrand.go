package workloadimpl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"sync"

	"golang.org/x/exp/rand"
)

type PrecomputedRand []byte

func PrecomputedRandInit(rng rand.Source, length int, alphabet string) func() PrecomputedRand {
	__antithesis_instrumentation__.Notify(699063)
	var prOnce sync.Once
	var pr PrecomputedRand
	return func() PrecomputedRand {
		__antithesis_instrumentation__.Notify(699064)
		prOnce.Do(func() {
			__antithesis_instrumentation__.Notify(699066)
			pr = make(PrecomputedRand, length)
			RandStringFast(rng, pr, alphabet)
		})
		__antithesis_instrumentation__.Notify(699065)
		return pr
	}
}

func (pr PrecomputedRand) FillBytes(offset int, buf []byte) int {
	__antithesis_instrumentation__.Notify(699067)
	if len(pr) == 0 {
		__antithesis_instrumentation__.Notify(699070)
		panic(`cannot fill from empty precomputed rand`)
	} else {
		__antithesis_instrumentation__.Notify(699071)
	}
	__antithesis_instrumentation__.Notify(699068)
	prIdx := offset
	for bufIdx := 0; bufIdx < len(buf); {
		__antithesis_instrumentation__.Notify(699072)
		if prIdx == len(pr) {
			__antithesis_instrumentation__.Notify(699075)
			prIdx = 0
		} else {
			__antithesis_instrumentation__.Notify(699076)
		}
		__antithesis_instrumentation__.Notify(699073)
		need, remaining := len(buf)-bufIdx, len(pr)-prIdx
		copyLen := need
		if copyLen > remaining {
			__antithesis_instrumentation__.Notify(699077)
			copyLen = remaining
		} else {
			__antithesis_instrumentation__.Notify(699078)
		}
		__antithesis_instrumentation__.Notify(699074)
		newBufIdx, newPRIdx := bufIdx+copyLen, prIdx+copyLen
		copy(buf[bufIdx:newBufIdx], pr[prIdx:newPRIdx])
		bufIdx = newBufIdx
		prIdx = newPRIdx
	}
	__antithesis_instrumentation__.Notify(699069)
	return prIdx
}
