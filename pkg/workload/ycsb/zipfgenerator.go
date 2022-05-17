package ycsb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"math"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/exp/rand"
)

const (
	defaultIMax  = 10000000000
	defaultTheta = 0.99
	defaultZetaN = 26.46902820178302
)

type ZipfGenerator struct {
	zipfGenMu ZipfGeneratorMu

	theta float64
	iMin  uint64

	alpha, zeta2, halfPowTheta float64
	verbose                    bool
}

type ZipfGeneratorMu struct {
	mu    syncutil.Mutex
	r     *rand.Rand
	iMax  uint64
	eta   float64
	zetaN float64
}

func NewZipfGenerator(
	rng *rand.Rand, iMin, iMax uint64, theta float64, verbose bool,
) (*ZipfGenerator, error) {
	__antithesis_instrumentation__.Notify(699507)
	if iMin > iMax {
		__antithesis_instrumentation__.Notify(699512)
		return nil, errors.Errorf("iMin %d > iMax %d", iMin, iMax)
	} else {
		__antithesis_instrumentation__.Notify(699513)
	}
	__antithesis_instrumentation__.Notify(699508)
	if theta < 0.0 || func() bool {
		__antithesis_instrumentation__.Notify(699514)
		return theta == 1.0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(699515)
		return nil, errors.Errorf("0 < theta, and theta != 1")
	} else {
		__antithesis_instrumentation__.Notify(699516)
	}
	__antithesis_instrumentation__.Notify(699509)

	z := ZipfGenerator{
		iMin: iMin,
		zipfGenMu: ZipfGeneratorMu{
			r:    rng,
			iMax: iMax,
		},
		theta:   theta,
		verbose: verbose,
	}
	z.zipfGenMu.mu.Lock()
	defer z.zipfGenMu.mu.Unlock()

	zeta2, err := computeZetaFromScratch(2, theta)
	if err != nil {
		__antithesis_instrumentation__.Notify(699517)
		return nil, errors.Wrap(err, "Could not compute zeta(2,theta)")
	} else {
		__antithesis_instrumentation__.Notify(699518)
	}
	__antithesis_instrumentation__.Notify(699510)
	var zetaN float64
	zetaN, err = computeZetaFromScratch(iMax+1-iMin, theta)
	if err != nil {
		__antithesis_instrumentation__.Notify(699519)
		return nil, errors.Wrapf(err, "Could not compute zeta(%d,theta)", iMax)
	} else {
		__antithesis_instrumentation__.Notify(699520)
	}
	__antithesis_instrumentation__.Notify(699511)
	z.alpha = 1.0 / (1.0 - theta)
	z.zipfGenMu.eta = (1 - math.Pow(2.0/float64(z.zipfGenMu.iMax+1-z.iMin), 1.0-theta)) / (1.0 - zeta2/zetaN)
	z.zipfGenMu.zetaN = zetaN
	z.zeta2 = zeta2
	z.halfPowTheta = 1.0 + math.Pow(0.5, z.theta)
	return &z, nil
}

func computeZetaIncrementally(oldIMax, iMax uint64, theta float64, sum float64) (float64, error) {
	__antithesis_instrumentation__.Notify(699521)
	if iMax < oldIMax {
		__antithesis_instrumentation__.Notify(699524)
		return 0, errors.Errorf("Can't increment iMax backwards!")
	} else {
		__antithesis_instrumentation__.Notify(699525)
	}
	__antithesis_instrumentation__.Notify(699522)
	for i := oldIMax + 1; i <= iMax; i++ {
		__antithesis_instrumentation__.Notify(699526)
		sum += 1.0 / math.Pow(float64(i), theta)
	}
	__antithesis_instrumentation__.Notify(699523)
	return sum, nil
}

func computeZetaFromScratch(n uint64, theta float64) (float64, error) {
	__antithesis_instrumentation__.Notify(699527)
	if n == defaultIMax && func() bool {
		__antithesis_instrumentation__.Notify(699530)
		return theta == defaultTheta == true
	}() == true {
		__antithesis_instrumentation__.Notify(699531)

		return defaultZetaN, nil
	} else {
		__antithesis_instrumentation__.Notify(699532)
	}
	__antithesis_instrumentation__.Notify(699528)
	zeta, err := computeZetaIncrementally(0, n, theta, 0.0)
	if err != nil {
		__antithesis_instrumentation__.Notify(699533)
		return zeta, errors.Wrap(err, "could not compute zeta")
	} else {
		__antithesis_instrumentation__.Notify(699534)
	}
	__antithesis_instrumentation__.Notify(699529)
	return zeta, nil
}

func (z *ZipfGenerator) Uint64() uint64 {
	__antithesis_instrumentation__.Notify(699535)
	z.zipfGenMu.mu.Lock()
	u := z.zipfGenMu.r.Float64()
	uz := u * z.zipfGenMu.zetaN
	var result uint64
	if uz < 1.0 {
		__antithesis_instrumentation__.Notify(699538)
		result = z.iMin
	} else {
		__antithesis_instrumentation__.Notify(699539)
		if uz < z.halfPowTheta {
			__antithesis_instrumentation__.Notify(699540)
			result = z.iMin + 1
		} else {
			__antithesis_instrumentation__.Notify(699541)
			spread := float64(z.zipfGenMu.iMax + 1 - z.iMin)
			result = z.iMin + uint64(int64(spread*math.Pow(z.zipfGenMu.eta*u-z.zipfGenMu.eta+1.0, z.alpha)))
		}
	}
	__antithesis_instrumentation__.Notify(699536)
	if z.verbose {
		__antithesis_instrumentation__.Notify(699542)
		fmt.Printf("Uint64[%d, %d] -> %d\n", z.iMin, z.zipfGenMu.iMax, result)
	} else {
		__antithesis_instrumentation__.Notify(699543)
	}
	__antithesis_instrumentation__.Notify(699537)
	z.zipfGenMu.mu.Unlock()
	return result
}

func (z *ZipfGenerator) IncrementIMax(count uint64) error {
	__antithesis_instrumentation__.Notify(699544)
	z.zipfGenMu.mu.Lock()
	zetaN, err := computeZetaIncrementally(
		z.zipfGenMu.iMax+1-z.iMin, z.zipfGenMu.iMax+count+1-z.iMin, z.theta, z.zipfGenMu.zetaN)
	if err != nil {
		__antithesis_instrumentation__.Notify(699546)
		z.zipfGenMu.mu.Unlock()
		return errors.Wrap(err, "Could not incrementally compute zeta")
	} else {
		__antithesis_instrumentation__.Notify(699547)
	}
	__antithesis_instrumentation__.Notify(699545)
	z.zipfGenMu.iMax += count
	eta := (1 - math.Pow(2.0/float64(z.zipfGenMu.iMax+1-z.iMin), 1.0-z.theta)) / (1.0 - z.zeta2/zetaN)
	z.zipfGenMu.eta = eta
	z.zipfGenMu.zetaN = zetaN
	z.zipfGenMu.mu.Unlock()
	return nil
}
