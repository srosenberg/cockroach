package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math"
	"math/rand"
)

type zipf struct {
	imax         float64
	v            float64
	q            float64
	s            float64
	oneminusQ    float64
	oneminusQinv float64
	hxm          float64
	hx0minusHxm  float64
}

func (z *zipf) h(x float64) float64 {
	__antithesis_instrumentation__.Notify(694562)
	return math.Exp(z.oneminusQ*math.Log(z.v+x)) * z.oneminusQinv
}

func (z *zipf) hinv(x float64) float64 {
	__antithesis_instrumentation__.Notify(694563)
	return math.Exp(z.oneminusQinv*math.Log(z.oneminusQ*x)) - z.v
}

func newZipf(s float64, v float64, imax uint64) *zipf {
	__antithesis_instrumentation__.Notify(694564)
	z := new(zipf)
	if s <= 1.0 || func() bool {
		__antithesis_instrumentation__.Notify(694566)
		return v < 1 == true
	}() == true {
		__antithesis_instrumentation__.Notify(694567)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(694568)
	}
	__antithesis_instrumentation__.Notify(694565)
	z.imax = float64(imax)
	z.v = v
	z.q = s
	z.oneminusQ = 1.0 - z.q
	z.oneminusQinv = 1.0 / z.oneminusQ
	z.hxm = z.h(z.imax + 0.5)
	z.hx0minusHxm = z.h(0.5) - math.Exp(math.Log(z.v)*(-z.q)) - z.hxm
	z.s = 1 - z.hinv(z.h(1.5)-math.Exp(-z.q*math.Log(z.v+1.0)))
	return z
}

func (z *zipf) Uint64(random *rand.Rand) uint64 {
	__antithesis_instrumentation__.Notify(694569)
	if z == nil {
		__antithesis_instrumentation__.Notify(694572)
		panic("rand: nil Zipf")
	} else {
		__antithesis_instrumentation__.Notify(694573)
	}
	__antithesis_instrumentation__.Notify(694570)
	k := 0.0

	for {
		__antithesis_instrumentation__.Notify(694574)
		r := random.Float64()
		ur := z.hxm + r*z.hx0minusHxm
		x := z.hinv(ur)
		k = math.Floor(x + 0.5)
		if k-x <= z.s {
			__antithesis_instrumentation__.Notify(694576)
			break
		} else {
			__antithesis_instrumentation__.Notify(694577)
		}
		__antithesis_instrumentation__.Notify(694575)
		if ur >= z.h(k+0.5)-math.Exp(-math.Log(k+z.v)*z.q) {
			__antithesis_instrumentation__.Notify(694578)
			break
		} else {
			__antithesis_instrumentation__.Notify(694579)
		}
	}
	__antithesis_instrumentation__.Notify(694571)
	return uint64(int64(k))
}
