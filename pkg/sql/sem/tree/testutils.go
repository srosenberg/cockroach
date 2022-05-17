package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
)

var presetTypesForTesting map[string]*types.T

func MockNameTypes(types map[string]*types.T) func() {
	__antithesis_instrumentation__.Notify(614527)
	presetTypesForTesting = types
	return func() {
		__antithesis_instrumentation__.Notify(614528)
		presetTypesForTesting = nil
	}
}

func SampleDatum(t *types.T) Datum {
	__antithesis_instrumentation__.Notify(614529)
	switch t.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(614530)
		a, _ := NewDBitArrayFromInt(123, 40)
		return a
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(614531)
		return MakeDBool(true)
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(614532)
		return NewDInt(123)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(614533)
		f := DFloat(123.456)
		return &f
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(614534)
		d := &DDecimal{}

		d.Decimal.SetFinite(3, 6)
		return d
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(614535)
		return NewDString("Carl")
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(614536)
		return NewDBytes("Princess")
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(614537)
		return NewDDate(pgdate.MakeCompatibleDateFromDisk(123123))
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(614538)
		return MakeDTime(timeofday.FromInt(789))
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(614539)
		return NewDTimeTZFromOffset(timeofday.FromInt(345), 5*60*60)
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(614540)
		return MustMakeDTimestamp(timeutil.Unix(123, 123), time.Second)
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(614541)
		return MustMakeDTimestampTZ(timeutil.Unix(123, 123), time.Second)
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(614542)
		i, _ := ParseDInterval(duration.IntervalStyle_POSTGRES, "1h1m1s")
		return i
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(614543)
		u, _ := ParseDUuidFromString("3189ad07-52f2-4d60-83e8-4a8347fef718")
		return u
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(614544)
		i, _ := ParseDIPAddrFromINetString("127.0.0.1")
		return i
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(614545)
		j, _ := ParseDJSON(`{"a": "b"}`)
		return j
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(614546)
		return NewDOid(DInt(1009))
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(614547)
		b := geo.NewCartesianBoundingBox().AddPoint(1, 2).AddPoint(3, 4)
		return NewDBox2D(*b)
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(614548)
		return NewDGeography(geo.MustParseGeographyFromEWKB([]byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0\x3f\x00\x00\x00\x00\x00\x00\xf0\x3f")))
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(614549)
		return NewDGeometry(geo.MustParseGeometryFromEWKB([]byte("\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\xf0\x3f\x00\x00\x00\x00\x00\x00\xf0\x3f")))
	default:
		__antithesis_instrumentation__.Notify(614550)
		panic(errors.AssertionFailedf("SampleDatum not implemented for %s", t))
	}
}
