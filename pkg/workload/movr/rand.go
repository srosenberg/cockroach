package movr

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"

	"golang.org/x/exp/rand"
)

const numerals = `1234567890`

var vehicleTypes = [...]string{`skateboard`, `bike`, `scooter`}
var vehicleColors = [...]string{`red`, `yellow`, `blue`, `green`, `black`}
var bikeBrands = [...]string{
	`Merida`, `Fuji`, `Cervelo`, `Pinarello`, `Santa Cruz`, `Kona`, `Schwinn`}

func randString(rng *rand.Rand, length int, alphabet string) string {
	__antithesis_instrumentation__.Notify(694857)
	buf := make([]byte, length)
	for i := range buf {
		__antithesis_instrumentation__.Notify(694859)
		buf[i] = alphabet[rng.Intn(len(alphabet))]
	}
	__antithesis_instrumentation__.Notify(694858)
	return string(buf)
}

func randCreditCard(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694860)
	return randString(rng, 10, numerals)
}

func randVehicleType(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694861)
	return vehicleTypes[rng.Intn(len(vehicleTypes))]
}

func randVehicleStatus(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694862)
	r := rng.Intn(100)
	switch {
	case r < 40:
		__antithesis_instrumentation__.Notify(694863)
		return `available`
	case r < 95:
		__antithesis_instrumentation__.Notify(694864)
		return `in_use`
	default:
		__antithesis_instrumentation__.Notify(694865)
		return `lost`
	}
}

func randLatLong(rng *rand.Rand) (float64, float64) {
	__antithesis_instrumentation__.Notify(694866)
	lat, long := float64(-180+rng.Intn(360)), float64(-90+rng.Intn(180))
	return lat, long
}

func randCity(rng *rand.Rand) string {
	__antithesis_instrumentation__.Notify(694867)
	idx := rng.Int31n(int32(len(cities)))
	return cities[idx].city
}

func randVehicleMetadata(rng *rand.Rand, vehicleType string) string {
	__antithesis_instrumentation__.Notify(694868)
	m := map[string]string{
		`color`: vehicleColors[rng.Intn(len(vehicleColors))],
	}
	switch vehicleType {
	case `bike`:
		__antithesis_instrumentation__.Notify(694871)
		m[`brand`] = bikeBrands[rng.Intn(len(bikeBrands))]
	default:
		__antithesis_instrumentation__.Notify(694872)
	}
	__antithesis_instrumentation__.Notify(694869)
	j, err := json.Marshal(m)
	if err != nil {
		__antithesis_instrumentation__.Notify(694873)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(694874)
	}
	__antithesis_instrumentation__.Notify(694870)
	return string(j)
}
