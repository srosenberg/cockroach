package pgwirecancel

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"math/rand"

	"github.com/cockroachdb/cockroach/pkg/base"
)

type BackendKeyData uint64

const (
	lower32BitsMask uint64 = 0x00000000FFFFFFFF
	lower52BitsMask uint64 = 0x000FFFFFFFFFFFFF
	leadingBitMask         = 1 << 63
)

func MakeBackendKeyData(rng *rand.Rand, sqlInstanceID base.SQLInstanceID) BackendKeyData {
	__antithesis_instrumentation__.Notify(561447)
	ret := rng.Uint64()
	if sqlInstanceID < 1<<11 {
		__antithesis_instrumentation__.Notify(561449)

		ret = ret & lower52BitsMask

		ret = ret | (uint64(sqlInstanceID) << 52)
	} else {
		__antithesis_instrumentation__.Notify(561450)

		ret = ret & lower32BitsMask

		ret = ret | leadingBitMask

		ret = ret | (uint64(sqlInstanceID) << 32)
	}
	__antithesis_instrumentation__.Notify(561448)
	return BackendKeyData(ret)
}

func (b BackendKeyData) GetSQLInstanceID() base.SQLInstanceID {
	__antithesis_instrumentation__.Notify(561451)
	bits := uint64(b)
	if bits&leadingBitMask == 0 {
		__antithesis_instrumentation__.Notify(561453)

		return base.SQLInstanceID(bits >> 52)
	} else {
		__antithesis_instrumentation__.Notify(561454)
	}
	__antithesis_instrumentation__.Notify(561452)

	bits = bits &^ leadingBitMask

	return base.SQLInstanceID(bits >> 32)

}
