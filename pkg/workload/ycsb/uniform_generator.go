package ycsb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"golang.org/x/exp/rand"
)

type UniformGenerator struct {
	iMin uint64
	mu   struct {
		syncutil.Mutex
		r    *rand.Rand
		iMax uint64
	}
}

func NewUniformGenerator(rng *rand.Rand, iMin, iMax uint64) (*UniformGenerator, error) {
	__antithesis_instrumentation__.Notify(699304)

	z := UniformGenerator{}
	z.iMin = iMin
	z.mu.r = rng
	z.mu.iMax = iMax

	return &z, nil
}

func (z *UniformGenerator) IncrementIMax(count uint64) error {
	__antithesis_instrumentation__.Notify(699305)
	z.mu.Lock()
	defer z.mu.Unlock()
	z.mu.iMax += count
	return nil
}

func (z *UniformGenerator) Uint64() uint64 {
	__antithesis_instrumentation__.Notify(699306)
	z.mu.Lock()
	defer z.mu.Unlock()
	return (uint64)(z.mu.r.Int63n((int64)(z.mu.iMax-z.iMin+1))) + z.iMin
}
