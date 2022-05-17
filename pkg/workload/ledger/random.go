package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
const aChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randInt(rng *rand.Rand, min, max int) int {
	__antithesis_instrumentation__.Notify(694684)
	return rng.Intn(max-min+1) + min
}

func randStringFromAlphabet(rng *rand.Rand, minLen, maxLen int, alphabet string) string {
	__antithesis_instrumentation__.Notify(694685)
	size := maxLen
	if maxLen-minLen != 0 {
		__antithesis_instrumentation__.Notify(694689)
		size = randInt(rng, minLen, maxLen)
	} else {
		__antithesis_instrumentation__.Notify(694690)
	}
	__antithesis_instrumentation__.Notify(694686)
	if size == 0 {
		__antithesis_instrumentation__.Notify(694691)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(694692)
	}
	__antithesis_instrumentation__.Notify(694687)

	b := make([]byte, size)
	for i := range b {
		__antithesis_instrumentation__.Notify(694693)
		b[i] = alphabet[rng.Intn(len(alphabet))]
	}
	__antithesis_instrumentation__.Notify(694688)
	return string(b)
}

func randAString(rng *rand.Rand, min, max int) string {
	__antithesis_instrumentation__.Notify(694694)
	return randStringFromAlphabet(rng, min, max, aChars)
}

func randPaymentID(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694695)
	uuidStr := uuid.MakeV4().String()
	return paymentIDPrefix + uuidStr
}

func randContext(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694696)
	return randAString(rng, 56, 56)
}

func randUsername(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694697)
	return randAString(rng, 18, 20)
}

func randResponse(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694698)
	return randAString(rng, 400, 400)
}

func randCurrencyCode(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694699)
	return randStringFromAlphabet(rng, 3, 3, letters)
}

func randTimestamp(rng *rand.Rand) time.Time {
	__antithesis_instrumentation__.Notify(694700)
	return timeutil.Unix(rng.Int63n(1600000000), rng.Int63())
}

func randSessionID(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694701)
	return randAString(rng, 60, 62)
}

func randSessionData(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694702)
	return randAString(rng, 160, 160)
}

func randAmount(rng *rand.Rand) float64 {
	__antithesis_instrumentation__.Notify(694703)
	return float64(randInt(rng, 100, 100000)) / 100
}

func (w ledger) randCustomer(rng *rand.Rand) int {
	__antithesis_instrumentation__.Notify(694704)
	return randInt(rng, 0, w.customers-1)
}
