package workloadimpl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"

	"golang.org/x/exp/rand"
)

func RandStringFast(rng rand.Source, buf []byte, alphabet string) {
	__antithesis_instrumentation__.Notify(699079)

	alphabetLen := uint64(len(alphabet))

	lettersCharsPerRand := uint64(math.Log(float64(math.MaxUint64)) / math.Log(float64(alphabetLen)))

	var r, charsLeft uint64
	for i := 0; i < len(buf); i++ {
		__antithesis_instrumentation__.Notify(699080)
		if charsLeft == 0 {
			__antithesis_instrumentation__.Notify(699082)
			r = rng.Uint64()
			charsLeft = lettersCharsPerRand
		} else {
			__antithesis_instrumentation__.Notify(699083)
		}
		__antithesis_instrumentation__.Notify(699081)
		buf[i] = alphabet[r%alphabetLen]
		r = r / alphabetLen
		charsLeft--
	}
}
