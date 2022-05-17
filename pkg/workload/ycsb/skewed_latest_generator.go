package ycsb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"golang.org/x/exp/rand"
)

type SkewedLatestGenerator struct {
	mu struct {
		syncutil.Mutex
		iMax    uint64
		zipfGen *ZipfGenerator
	}
}

func NewSkewedLatestGenerator(
	rng *rand.Rand, iMin, iMax uint64, theta float64, verbose bool,
) (*SkewedLatestGenerator, error) {
	__antithesis_instrumentation__.Notify(699298)

	z := SkewedLatestGenerator{}
	z.mu.iMax = iMax
	zipfGen, err := NewZipfGenerator(rng, 0, iMax-iMin, theta, verbose)
	if err != nil {
		__antithesis_instrumentation__.Notify(699300)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(699301)
	}
	__antithesis_instrumentation__.Notify(699299)
	z.mu.zipfGen = zipfGen

	return &z, nil
}

func (z *SkewedLatestGenerator) IncrementIMax(count uint64) error {
	__antithesis_instrumentation__.Notify(699302)
	z.mu.Lock()
	defer z.mu.Unlock()
	z.mu.iMax += count
	return z.mu.zipfGen.IncrementIMax(count)
}

func (z *SkewedLatestGenerator) Uint64() uint64 {
	__antithesis_instrumentation__.Notify(699303)
	z.mu.Lock()
	defer z.mu.Unlock()
	return z.mu.iMax - z.mu.zipfGen.Uint64()
}
