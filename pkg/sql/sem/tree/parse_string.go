package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func ParseAndRequireString(
	t *types.T, s string, ctx ParseTimeContext,
) (d Datum, dependsOnContext bool, err error) {
	__antithesis_instrumentation__.Notify(611521)
	switch t.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(611524)
		d, dependsOnContext, err = ParseDArrayFromString(ctx, s, t.ArrayContents())
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(611525)
		r, err := ParseDBitArray(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(611552)
			return nil, false, err
		} else {
			__antithesis_instrumentation__.Notify(611553)
		}
		__antithesis_instrumentation__.Notify(611526)
		d = formatBitArrayToType(r, t)
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(611527)
		d, err = ParseDBool(strings.TrimSpace(s))
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(611528)
		d, err = ParseDByte(s)
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(611529)
		d, dependsOnContext, err = ParseDDate(ctx, s)
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(611530)
		d, err = ParseDDecimal(strings.TrimSpace(s))
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(611531)
		d, err = ParseDFloat(strings.TrimSpace(s))
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(611532)
		d, err = ParseDIPAddrFromINetString(s)
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(611533)
		d, err = ParseDInt(strings.TrimSpace(s))
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(611534)
		itm, typErr := t.IntervalTypeMetadata()
		if typErr != nil {
			__antithesis_instrumentation__.Notify(611554)
			return nil, false, typErr
		} else {
			__antithesis_instrumentation__.Notify(611555)
		}
		__antithesis_instrumentation__.Notify(611535)
		d, err = ParseDIntervalWithTypeMetadata(intervalStyle(ctx), s, itm)
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(611536)
		d, err = ParseDBox2D(s)
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(611537)
		d, err = ParseDGeography(s)
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(611538)
		d, err = ParseDGeometry(s)
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(611539)
		d, err = ParseDJSON(s)
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(611540)
		if t.Oid() != oid.T_oid && func() bool {
			__antithesis_instrumentation__.Notify(611556)
			return s == ZeroOidValue == true
		}() == true {
			__antithesis_instrumentation__.Notify(611557)
			d = wrapAsZeroOid(t)
		} else {
			__antithesis_instrumentation__.Notify(611558)
			i, err := ParseDInt(s)
			if err != nil {
				__antithesis_instrumentation__.Notify(611560)
				return nil, false, err
			} else {
				__antithesis_instrumentation__.Notify(611561)
			}
			__antithesis_instrumentation__.Notify(611559)
			d = NewDOid(*i)
		}
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(611541)

		if t.Width() > 0 {
			__antithesis_instrumentation__.Notify(611562)
			s = util.TruncateString(s, int(t.Width()))
		} else {
			__antithesis_instrumentation__.Notify(611563)
		}
		__antithesis_instrumentation__.Notify(611542)
		return NewDString(s), false, nil
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(611543)
		d, dependsOnContext, err = ParseDTime(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(611544)
		d, dependsOnContext, err = ParseDTimeTZ(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(611545)
		d, dependsOnContext, err = ParseDTimestamp(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(611546)
		d, dependsOnContext, err = ParseDTimestampTZ(ctx, s, TimeFamilyPrecisionToRoundDuration(t.Precision()))
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(611547)
		d, err = ParseDUuidFromString(s)
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(611548)
		d, err = MakeDEnumFromLogicalRepresentation(t, s)
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(611549)
		d, dependsOnContext, err = ParseDTupleFromString(ctx, s, t)
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(611550)
		d = DVoidDatum
	default:
		__antithesis_instrumentation__.Notify(611551)
		return nil, false, errors.AssertionFailedf("unknown type %s (%T)", t, t)
	}
	__antithesis_instrumentation__.Notify(611522)
	if err != nil {
		__antithesis_instrumentation__.Notify(611564)
		return d, dependsOnContext, err
	} else {
		__antithesis_instrumentation__.Notify(611565)
	}
	__antithesis_instrumentation__.Notify(611523)
	d, err = AdjustValueToType(t, d)
	return d, dependsOnContext, err
}
