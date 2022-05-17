package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/errors"

type Volatility int8

const (
	VolatilityLeakProof Volatility = 1 + iota

	VolatilityImmutable

	VolatilityStable

	VolatilityVolatile
)

func (v Volatility) String() string {
	__antithesis_instrumentation__.Notify(615934)
	switch v {
	case VolatilityLeakProof:
		__antithesis_instrumentation__.Notify(615935)
		return "leak-proof"
	case VolatilityImmutable:
		__antithesis_instrumentation__.Notify(615936)
		return "immutable"
	case VolatilityStable:
		__antithesis_instrumentation__.Notify(615937)
		return "stable"
	case VolatilityVolatile:
		__antithesis_instrumentation__.Notify(615938)
		return "volatile"
	default:
		__antithesis_instrumentation__.Notify(615939)
		return "invalid"
	}
}

func (v Volatility) ToPostgres() (provolatile string, proleakproof bool) {
	__antithesis_instrumentation__.Notify(615940)
	switch v {
	case VolatilityLeakProof:
		__antithesis_instrumentation__.Notify(615941)
		return "i", true
	case VolatilityImmutable:
		__antithesis_instrumentation__.Notify(615942)
		return "i", false
	case VolatilityStable:
		__antithesis_instrumentation__.Notify(615943)
		return "s", false
	case VolatilityVolatile:
		__antithesis_instrumentation__.Notify(615944)
		return "v", false
	default:
		__antithesis_instrumentation__.Notify(615945)
		panic(errors.AssertionFailedf("invalid volatility %s", v))
	}
}

func VolatilityFromPostgres(provolatile string, proleakproof bool) (Volatility, error) {
	__antithesis_instrumentation__.Notify(615946)
	switch provolatile {
	case "i":
		__antithesis_instrumentation__.Notify(615947)
		if proleakproof {
			__antithesis_instrumentation__.Notify(615952)
			return VolatilityLeakProof, nil
		} else {
			__antithesis_instrumentation__.Notify(615953)
		}
		__antithesis_instrumentation__.Notify(615948)
		return VolatilityImmutable, nil
	case "s":
		__antithesis_instrumentation__.Notify(615949)
		return VolatilityStable, nil
	case "v":
		__antithesis_instrumentation__.Notify(615950)
		return VolatilityVolatile, nil
	default:
		__antithesis_instrumentation__.Notify(615951)
		return 0, errors.AssertionFailedf("invalid provolatile %s", provolatile)
	}
}
