package faker

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strconv"

	"golang.org/x/exp/rand"
)

type addressFaker struct {
	streetAddress *weightedEntries
	streetSuffix  *weightedEntries

	name nameFaker
}

func (f *addressFaker) StreetAddress(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694064)
	return f.streetAddress.Rand(rng).(func(rng *rand.Rand) string)(rng)
}

func (f *addressFaker) buildingNumber(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694065)
	return strconv.Itoa(randInt(rng, 1000, 99999))
}

func (f *addressFaker) streetName(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694066)
	return fmt.Sprintf(`%s %s`, f.firstOrLastName(rng), f.streetSuffix.Rand(rng))
}

func (f *addressFaker) firstOrLastName(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694067)
	switch rng.Intn(3) {
	case 0:
		__antithesis_instrumentation__.Notify(694069)
		return f.name.firstNameFemale.Rand(rng).(string)
	case 1:
		__antithesis_instrumentation__.Notify(694070)
		return f.name.firstNameMale.Rand(rng).(string)
	case 2:
		__antithesis_instrumentation__.Notify(694071)
		return f.name.lastName.Rand(rng).(string)
	default:
		__antithesis_instrumentation__.Notify(694072)
	}
	__antithesis_instrumentation__.Notify(694068)
	panic(`unreachable`)
}

func secondaryAddress(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694073)
	switch rng.Intn(2) {
	case 0:
		__antithesis_instrumentation__.Notify(694075)
		return fmt.Sprintf(`Apt. %d`, rng.Intn(100))
	case 1:
		__antithesis_instrumentation__.Notify(694076)
		return fmt.Sprintf(`Suite %d`, rng.Intn(100))
	default:
		__antithesis_instrumentation__.Notify(694077)
	}
	__antithesis_instrumentation__.Notify(694074)
	panic(`unreachable`)
}

func newAddressFaker(name nameFaker) addressFaker {
	__antithesis_instrumentation__.Notify(694078)
	f := addressFaker{name: name}
	f.streetSuffix = streetSuffix()
	f.streetAddress = makeWeightedEntries(
		func(rng *rand.Rand) string {
			__antithesis_instrumentation__.Notify(694080)
			return fmt.Sprintf(`%s %s`, f.buildingNumber(rng), f.streetName(rng))
		}, 0.5,
		func(rng *rand.Rand) string {
			__antithesis_instrumentation__.Notify(694081)
			return fmt.Sprintf(`%s %s %s`,
				f.buildingNumber(rng), f.streetName(rng), secondaryAddress(rng))
		}, 0.5,
	)
	__antithesis_instrumentation__.Notify(694079)
	return f
}
