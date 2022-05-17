package randgen

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "math/rand"

func RandString(rng *rand.Rand, length int, alphabet string) string {
	__antithesis_instrumentation__.Notify(564699)
	buf := make([]byte, length)
	for i := range buf {
		__antithesis_instrumentation__.Notify(564701)
		buf[i] = alphabet[rng.Intn(len(alphabet))]
	}
	__antithesis_instrumentation__.Notify(564700)
	return string(buf)
}
