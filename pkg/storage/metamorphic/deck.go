package metamorphic

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
		deck  []int
	}
}

func newDeck(rng *rand.Rand, weights ...int) *deck {
	__antithesis_instrumentation__.Notify(639628)
	var sum int
	for i := range weights {
		__antithesis_instrumentation__.Notify(639631)
		sum += weights[i]
	}
	__antithesis_instrumentation__.Notify(639629)
	expandedDeck := make([]int, 0, sum)
	for i := range weights {
		__antithesis_instrumentation__.Notify(639632)
		for j := 0; j < weights[i]; j++ {
			__antithesis_instrumentation__.Notify(639633)
			expandedDeck = append(expandedDeck, i)
		}
	}
	__antithesis_instrumentation__.Notify(639630)
	d := &deck{
		rng: rng,
	}
	d.mu.index = len(expandedDeck)
	d.mu.deck = expandedDeck
	return d
}

func (d *deck) Int() int {
	__antithesis_instrumentation__.Notify(639634)
	d.mu.Lock()
	if d.mu.index == len(d.mu.deck) {
		__antithesis_instrumentation__.Notify(639636)
		d.rng.Shuffle(len(d.mu.deck), func(i, j int) {
			__antithesis_instrumentation__.Notify(639638)
			d.mu.deck[i], d.mu.deck[j] = d.mu.deck[j], d.mu.deck[i]
		})
		__antithesis_instrumentation__.Notify(639637)
		d.mu.index = 0
	} else {
		__antithesis_instrumentation__.Notify(639639)
	}
	__antithesis_instrumentation__.Notify(639635)
	result := d.mu.deck[d.mu.index]
	d.mu.index++
	d.mu.Unlock()
	return result
}
