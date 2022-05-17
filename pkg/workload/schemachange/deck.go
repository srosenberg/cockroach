package schemachange

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type deck struct {
	rng *rand.Rand
	mu  struct {
		syncutil.Mutex
		index int
		vals  []int
	}
}

func newDeck(rng *rand.Rand, weights ...int) *deck {
	__antithesis_instrumentation__.Notify(695724)
	var sum int
	for i := range weights {
		__antithesis_instrumentation__.Notify(695727)
		sum += weights[i]
	}
	__antithesis_instrumentation__.Notify(695725)
	vals := make([]int, 0, sum)
	for i := range weights {
		__antithesis_instrumentation__.Notify(695728)
		for j := 0; j < weights[i]; j++ {
			__antithesis_instrumentation__.Notify(695729)
			vals = append(vals, i)
		}
	}
	__antithesis_instrumentation__.Notify(695726)
	d := &deck{
		rng: rng,
	}
	d.mu.index = len(vals)
	d.mu.vals = vals
	return d
}

func (d *deck) Int() int {
	__antithesis_instrumentation__.Notify(695730)
	d.mu.Lock()
	if d.mu.index == len(d.mu.vals) {
		__antithesis_instrumentation__.Notify(695732)
		d.rng.Shuffle(len(d.mu.vals), func(i, j int) {
			__antithesis_instrumentation__.Notify(695734)
			d.mu.vals[i], d.mu.vals[j] = d.mu.vals[j], d.mu.vals[i]
		})
		__antithesis_instrumentation__.Notify(695733)
		d.mu.index = 0
	} else {
		__antithesis_instrumentation__.Notify(695735)
	}
	__antithesis_instrumentation__.Notify(695731)
	result := d.mu.vals[d.mu.index]
	d.mu.index++
	d.mu.Unlock()
	return result
}
