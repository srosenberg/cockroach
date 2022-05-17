package faker

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "golang.org/x/exp/rand"

type Faker struct {
	addressFaker
	loremFaker
	nameFaker
}

func NewFaker() Faker {
	__antithesis_instrumentation__.Notify(694087)
	f := Faker{
		loremFaker: newLoremFaker(),
		nameFaker:  newNameFaker(),
	}
	f.addressFaker = newAddressFaker(f.nameFaker)
	return f
}

type weightedEntry struct {
	weight float64
	entry  interface{}
}

type weightedEntries struct {
	entries     []weightedEntry
	totalWeight float64
}

func makeWeightedEntries(entriesAndWeights ...interface{}) *weightedEntries {
	__antithesis_instrumentation__.Notify(694088)
	we := make([]weightedEntry, 0, len(entriesAndWeights)/2)
	var totalWeight float64
	for idx := 0; idx < len(entriesAndWeights); idx += 2 {
		__antithesis_instrumentation__.Notify(694090)
		e, w := entriesAndWeights[idx], entriesAndWeights[idx+1].(float64)
		we = append(we, weightedEntry{weight: w, entry: e})
		totalWeight += w
	}
	__antithesis_instrumentation__.Notify(694089)
	return &weightedEntries{entries: we, totalWeight: totalWeight}
}

func (e *weightedEntries) Rand(rng *rand.Rand) interface{} {
	__antithesis_instrumentation__.Notify(694091)
	rngWeight := rng.Float64() * e.totalWeight
	var w float64
	for i := range e.entries {
		__antithesis_instrumentation__.Notify(694093)
		w += e.entries[i].weight
		if w > rngWeight {
			__antithesis_instrumentation__.Notify(694094)
			return e.entries[i].entry
		} else {
			__antithesis_instrumentation__.Notify(694095)
		}
	}
	__antithesis_instrumentation__.Notify(694092)
	panic(`unreachable`)
}

func randInt(rng *rand.Rand, minInclusive, maxInclusive int) int {
	__antithesis_instrumentation__.Notify(694096)
	return rng.Intn(maxInclusive-minInclusive+1) + minInclusive
}
