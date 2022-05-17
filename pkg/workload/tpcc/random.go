package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/workload/workloadimpl"
	"golang.org/x/exp/rand"
)

var cLastTokens = [...]string{
	"BAR", "OUGHT", "ABLE", "PRI", "PRES",
	"ESE", "ANTI", "CALLY", "ATION", "EING"}

func (w *tpcc) initNonUniformRandomConstants() {
	__antithesis_instrumentation__.Notify(698231)
	rng := rand.New(rand.NewSource(w.seed))
	w.cLoad = rng.Intn(256)
	w.cItemID = rng.Intn(1024)
	w.cCustomerID = rng.Intn(8192)
}

const precomputedLength = 10000
const aCharsAlphabet = `abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890`
const lettersAlphabet = `ABCDEFGHIJKLMNOPQRSTUVWXYZ`
const numbersAlphabet = `1234567890`

type tpccRand struct {
	*rand.Rand

	aChars, letters, numbers workloadimpl.PrecomputedRand
}

type aCharsOffset int
type lettersOffset int
type numbersOffset int

func randStringFromAlphabet(
	rng *rand.Rand,
	a *bufalloc.ByteAllocator,
	minLen, maxLen int,
	pr workloadimpl.PrecomputedRand,
	prOffset *int,
) []byte {
	__antithesis_instrumentation__.Notify(698232)
	size := maxLen
	if maxLen-minLen != 0 {
		__antithesis_instrumentation__.Notify(698235)
		size = int(randInt(rng, minLen, maxLen))
	} else {
		__antithesis_instrumentation__.Notify(698236)
	}
	__antithesis_instrumentation__.Notify(698233)
	if size == 0 {
		__antithesis_instrumentation__.Notify(698237)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(698238)
	}
	__antithesis_instrumentation__.Notify(698234)

	var b []byte
	*a, b = a.Alloc(size, 0)
	*prOffset = pr.FillBytes(*prOffset, b)
	return b
}

func randAStringInitialDataOnly(
	rng *tpccRand, ao *aCharsOffset, a *bufalloc.ByteAllocator, min, max int,
) []byte {
	__antithesis_instrumentation__.Notify(698239)
	return randStringFromAlphabet(rng.Rand, a, min, max, rng.aChars, (*int)(ao))
}

func randNStringInitialDataOnly(
	rng *tpccRand, no *numbersOffset, a *bufalloc.ByteAllocator, min, max int,
) []byte {
	__antithesis_instrumentation__.Notify(698240)
	return randStringFromAlphabet(rng.Rand, a, min, max, rng.numbers, (*int)(no))
}

func randStateInitialDataOnly(rng *tpccRand, lo *lettersOffset, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698241)
	return randStringFromAlphabet(rng.Rand, a, 2, 2, rng.letters, (*int)(lo))
}

func randOriginalStringInitialDataOnly(
	rng *tpccRand, ao *aCharsOffset, a *bufalloc.ByteAllocator,
) []byte {
	__antithesis_instrumentation__.Notify(698242)
	if rng.Rand.Intn(9) == 0 {
		__antithesis_instrumentation__.Notify(698244)
		l := int(randInt(rng.Rand, 26, 50))
		off := int(randInt(rng.Rand, 0, l-8))
		var buf []byte
		*a, buf = a.Alloc(l, 0)
		copy(buf[:off], randAStringInitialDataOnly(rng, ao, a, off, off))
		copy(buf[off:off+8], originalString)
		copy(buf[off+8:], randAStringInitialDataOnly(rng, ao, a, l-off-8, l-off-8))
		return buf
	} else {
		__antithesis_instrumentation__.Notify(698245)
	}
	__antithesis_instrumentation__.Notify(698243)
	return randAStringInitialDataOnly(rng, ao, a, 26, 50)
}

func randZipInitialDataOnly(rng *tpccRand, no *numbersOffset, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698246)
	var buf []byte
	*a, buf = a.Alloc(9, 0)
	copy(buf[:4], randNStringInitialDataOnly(rng, no, a, 4, 4))
	copy(buf[4:], `11111`)
	return buf
}

func randTax(rng *rand.Rand) float64 {
	__antithesis_instrumentation__.Notify(698247)
	return float64(randInt(rng, 0, 2000)) / float64(10000.0)
}

func randInt(rng *rand.Rand, min, max int) int64 {
	__antithesis_instrumentation__.Notify(698248)
	return int64(rng.Intn(max-min+1) + min)
}

func randCLastSyllables(n int, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698249)
	const scratchLen = 3 * 5
	var buf []byte
	*a, buf = a.Alloc(scratchLen, 0)
	buf = buf[:0]
	buf = append(buf, cLastTokens[n/100]...)
	n = n % 100
	buf = append(buf, cLastTokens[n/10]...)
	n = n % 10
	buf = append(buf, cLastTokens[n]...)
	return buf
}

func (w *tpcc) randCLast(rng *rand.Rand, a *bufalloc.ByteAllocator) []byte {
	__antithesis_instrumentation__.Notify(698250)
	return randCLastSyllables(((rng.Intn(256)|rng.Intn(1000))+w.cLoad)%1000, a)
}

func (w *tpcc) randCustomerID(rng *rand.Rand) int {
	__antithesis_instrumentation__.Notify(698251)
	return ((rng.Intn(1024) | (rng.Intn(3000) + 1) + w.cCustomerID) % 3000) + 1
}

func (w *tpcc) randItemID(rng *rand.Rand) int {
	__antithesis_instrumentation__.Notify(698252)
	return ((rng.Intn(8190) | (rng.Intn(100000) + 1) + w.cItemID) % 100000) + 1
}
