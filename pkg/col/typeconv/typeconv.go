package typeconv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/json"
)

var DatumVecCanonicalTypeFamily = types.Family(1000000)

func TypeFamilyToCanonicalTypeFamily(family types.Family) types.Family {
	__antithesis_instrumentation__.Notify(55762)
	switch family {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(55763)
		return types.BoolFamily
	case types.BytesFamily, types.StringFamily, types.UuidFamily:
		__antithesis_instrumentation__.Notify(55764)
		return types.BytesFamily
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(55765)
		return types.DecimalFamily
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(55766)
		return types.JsonFamily
	case types.IntFamily, types.DateFamily:
		__antithesis_instrumentation__.Notify(55767)
		return types.IntFamily
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(55768)
		return types.FloatFamily
	case types.TimestampTZFamily, types.TimestampFamily:
		__antithesis_instrumentation__.Notify(55769)
		return types.TimestampTZFamily
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(55770)
		return types.IntervalFamily
	default:
		__antithesis_instrumentation__.Notify(55771)

		return DatumVecCanonicalTypeFamily
	}
}

func ToCanonicalTypeFamilies(typs []*types.T) []types.Family {
	__antithesis_instrumentation__.Notify(55772)
	families := make([]types.Family, len(typs))
	for i := range typs {
		__antithesis_instrumentation__.Notify(55774)
		families[i] = TypeFamilyToCanonicalTypeFamily(typs[i].Family())
	}
	__antithesis_instrumentation__.Notify(55773)
	return families
}

func UnsafeFromGoType(v interface{}) *types.T {
	__antithesis_instrumentation__.Notify(55775)
	switch t := v.(type) {
	case int16:
		__antithesis_instrumentation__.Notify(55776)
		return types.Int2
	case int32:
		__antithesis_instrumentation__.Notify(55777)
		return types.Int4
	case int, int64:
		__antithesis_instrumentation__.Notify(55778)
		return types.Int
	case bool:
		__antithesis_instrumentation__.Notify(55779)
		return types.Bool
	case float64:
		__antithesis_instrumentation__.Notify(55780)
		return types.Float
	case []byte:
		__antithesis_instrumentation__.Notify(55781)
		return types.Bytes
	case string:
		__antithesis_instrumentation__.Notify(55782)
		return types.String
	case apd.Decimal:
		__antithesis_instrumentation__.Notify(55783)
		return types.Decimal
	case time.Time:
		__antithesis_instrumentation__.Notify(55784)
		return types.TimestampTZ
	case duration.Duration:
		__antithesis_instrumentation__.Notify(55785)
		return types.Interval
	case json.JSON:
		__antithesis_instrumentation__.Notify(55786)
		return types.Jsonb
	default:
		__antithesis_instrumentation__.Notify(55787)
		panic(fmt.Sprintf("type %s not supported yet", t))
	}
}

var TypesSupportedNatively []*types.T

func init() {
	for _, t := range types.Scalar {
		if TypeFamilyToCanonicalTypeFamily(t.Family()) == DatumVecCanonicalTypeFamily {
			continue
		}
		if t.Family() == types.IntFamily {
			TypesSupportedNatively = append(TypesSupportedNatively, types.Int2)
			TypesSupportedNatively = append(TypesSupportedNatively, types.Int4)
		}
		TypesSupportedNatively = append(TypesSupportedNatively, t)
	}
}
