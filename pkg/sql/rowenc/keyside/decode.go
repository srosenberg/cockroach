package keyside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/timetz"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func Decode(
	a *tree.DatumAlloc, valType *types.T, key []byte, dir encoding.Direction,
) (_ tree.Datum, remainingKey []byte, _ error) {
	__antithesis_instrumentation__.Notify(570673)
	if (dir != encoding.Ascending) && func() bool {
		__antithesis_instrumentation__.Notify(570676)
		return (dir != encoding.Descending) == true
	}() == true {
		__antithesis_instrumentation__.Notify(570677)
		return nil, nil, errors.Errorf("invalid direction: %d", dir)
	} else {
		__antithesis_instrumentation__.Notify(570678)
	}
	__antithesis_instrumentation__.Notify(570674)
	var isNull bool
	if key, isNull = encoding.DecodeIfNull(key); isNull {
		__antithesis_instrumentation__.Notify(570679)
		return tree.DNull, key, nil
	} else {
		__antithesis_instrumentation__.Notify(570680)
	}
	__antithesis_instrumentation__.Notify(570675)
	var rkey []byte
	var err error

	switch valType.Family() {
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(570681)
		return decodeArrayKey(a, valType, key, dir)
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(570682)
		var r bitarray.BitArray
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570734)
			rkey, r, err = encoding.DecodeBitArrayAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570735)
			rkey, r, err = encoding.DecodeBitArrayDescending(key)
		}
		__antithesis_instrumentation__.Notify(570683)
		return a.NewDBitArray(tree.DBitArray{BitArray: r}), rkey, err
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(570684)
		var i int64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570736)
			rkey, i, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570737)
			rkey, i, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(570685)

		return tree.MakeDBool(tree.DBool(i != 0)), rkey, err
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(570686)
		var i int64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570738)
			rkey, i, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570739)
			rkey, i, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(570687)
		return a.NewDInt(tree.DInt(i)), rkey, err
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(570688)
		var f float64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570740)
			rkey, f, err = encoding.DecodeFloatAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570741)
			rkey, f, err = encoding.DecodeFloatDescending(key)
		}
		__antithesis_instrumentation__.Notify(570689)
		return a.NewDFloat(tree.DFloat(f)), rkey, err
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(570690)
		var d apd.Decimal
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570742)
			rkey, d, err = encoding.DecodeDecimalAscending(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570743)
			rkey, d, err = encoding.DecodeDecimalDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570691)
		dd := a.NewDDecimal(tree.DDecimal{Decimal: d})
		return dd, rkey, err
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(570692)
		var r string
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570744)

			rkey, r, err = encoding.DecodeUnsafeStringAscendingDeepCopy(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570745)
			rkey, r, err = encoding.DecodeUnsafeStringDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570693)
		if valType.Oid() == oid.T_name {
			__antithesis_instrumentation__.Notify(570746)
			return a.NewDName(tree.DString(r)), rkey, err
		} else {
			__antithesis_instrumentation__.Notify(570747)
		}
		__antithesis_instrumentation__.Notify(570694)
		return a.NewDString(tree.DString(r)), rkey, err
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(570695)
		var r string
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570748)

			rkey, r, err = encoding.DecodeUnsafeStringAscendingDeepCopy(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570749)
			rkey, r, err = encoding.DecodeUnsafeStringDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570696)
		if err != nil {
			__antithesis_instrumentation__.Notify(570750)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570751)
		}
		__antithesis_instrumentation__.Notify(570697)
		d, err := a.NewDCollatedString(r, valType.Locale())
		return d, rkey, err
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(570698)

		jsonLen, err := encoding.PeekLength(key)
		if err != nil {
			__antithesis_instrumentation__.Notify(570752)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570753)
		}
		__antithesis_instrumentation__.Notify(570699)
		return tree.DNull, key[jsonLen:], nil
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(570700)
		var r []byte
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570754)

			rkey, r, err = encoding.DecodeBytesAscending(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570755)
			rkey, r, err = encoding.DecodeBytesDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570701)
		return a.NewDBytes(tree.DBytes(r)), rkey, err
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(570702)
		rkey, err = encoding.DecodeVoidAscendingOrDescending(key)
		return a.NewDVoid(), rkey, err
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(570703)
		var r geopb.BoundingBox
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570756)
			rkey, r, err = encoding.DecodeBox2DAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570757)
			rkey, r, err = encoding.DecodeBox2DDescending(key)
		}
		__antithesis_instrumentation__.Notify(570704)
		return a.NewDBox2D(tree.DBox2D{
			CartesianBoundingBox: geo.CartesianBoundingBox{BoundingBox: r},
		}), rkey, err
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(570705)
		g := a.NewDGeographyEmpty()
		so := g.Geography.SpatialObjectRef()
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570758)
			rkey, err = encoding.DecodeGeoAscending(key, so)
		} else {
			__antithesis_instrumentation__.Notify(570759)
			rkey, err = encoding.DecodeGeoDescending(key, so)
		}
		__antithesis_instrumentation__.Notify(570706)
		a.DoneInitNewDGeo(so)
		return g, rkey, err
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(570707)
		g := a.NewDGeometryEmpty()
		so := g.Geometry.SpatialObjectRef()
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570760)
			rkey, err = encoding.DecodeGeoAscending(key, so)
		} else {
			__antithesis_instrumentation__.Notify(570761)
			rkey, err = encoding.DecodeGeoDescending(key, so)
		}
		__antithesis_instrumentation__.Notify(570708)
		a.DoneInitNewDGeo(so)
		return g, rkey, err
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(570709)
		var t int64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570762)
			rkey, t, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570763)
			rkey, t, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(570710)
		return a.NewDDate(tree.MakeDDate(pgdate.MakeCompatibleDateFromDisk(t))), rkey, err
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(570711)
		var t int64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570764)
			rkey, t, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570765)
			rkey, t, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(570712)
		return a.NewDTime(tree.DTime(t)), rkey, err
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(570713)
		var t timetz.TimeTZ
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570766)
			rkey, t, err = encoding.DecodeTimeTZAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570767)
			rkey, t, err = encoding.DecodeTimeTZDescending(key)
		}
		__antithesis_instrumentation__.Notify(570714)
		return a.NewDTimeTZ(tree.DTimeTZ{TimeTZ: t}), rkey, err
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(570715)
		var t time.Time
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570768)
			rkey, t, err = encoding.DecodeTimeAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570769)
			rkey, t, err = encoding.DecodeTimeDescending(key)
		}
		__antithesis_instrumentation__.Notify(570716)
		return a.NewDTimestamp(tree.DTimestamp{Time: t}), rkey, err
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(570717)
		var t time.Time
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570770)
			rkey, t, err = encoding.DecodeTimeAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570771)
			rkey, t, err = encoding.DecodeTimeDescending(key)
		}
		__antithesis_instrumentation__.Notify(570718)
		return a.NewDTimestampTZ(tree.DTimestampTZ{Time: t}), rkey, err
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(570719)
		var d duration.Duration
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570772)
			rkey, d, err = encoding.DecodeDurationAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570773)
			rkey, d, err = encoding.DecodeDurationDescending(key)
		}
		__antithesis_instrumentation__.Notify(570720)
		return a.NewDInterval(tree.DInterval{Duration: d}), rkey, err
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(570721)
		var r []byte
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570774)

			rkey, r, err = encoding.DecodeBytesAscending(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570775)
			rkey, r, err = encoding.DecodeBytesDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570722)
		if err != nil {
			__antithesis_instrumentation__.Notify(570776)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570777)
		}
		__antithesis_instrumentation__.Notify(570723)
		u, err := uuid.FromBytes(r)
		return a.NewDUuid(tree.DUuid{UUID: u}), rkey, err
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(570724)
		var r []byte
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570778)

			rkey, r, err = encoding.DecodeBytesAscending(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570779)
			rkey, r, err = encoding.DecodeBytesDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570725)
		if err != nil {
			__antithesis_instrumentation__.Notify(570780)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570781)
		}
		__antithesis_instrumentation__.Notify(570726)
		var ipAddr ipaddr.IPAddr
		_, err := ipAddr.FromBuffer(r)
		return a.NewDIPAddr(tree.DIPAddr{IPAddr: ipAddr}), rkey, err
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(570727)
		var i int64
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570782)
			rkey, i, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(570783)
			rkey, i, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(570728)
		return a.NewDOid(tree.MakeDOid(tree.DInt(i))), rkey, err
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(570729)
		var r []byte
		if dir == encoding.Ascending {
			__antithesis_instrumentation__.Notify(570784)

			rkey, r, err = encoding.DecodeBytesAscending(key, nil)
		} else {
			__antithesis_instrumentation__.Notify(570785)
			rkey, r, err = encoding.DecodeBytesDescending(key, nil)
		}
		__antithesis_instrumentation__.Notify(570730)
		if err != nil {
			__antithesis_instrumentation__.Notify(570786)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570787)
		}
		__antithesis_instrumentation__.Notify(570731)
		phys, log, err := tree.GetEnumComponentsFromPhysicalRep(valType, r)
		if err != nil {
			__antithesis_instrumentation__.Notify(570788)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(570789)
		}
		__antithesis_instrumentation__.Notify(570732)
		return a.NewDEnum(tree.DEnum{EnumTyp: valType, PhysicalRep: phys, LogicalRep: log}), rkey, nil
	default:
		__antithesis_instrumentation__.Notify(570733)
		return nil, nil, errors.Errorf("unable to decode table key: %s", valType)
	}
}

func Skip(key []byte) (remainingKey []byte, _ error) {
	__antithesis_instrumentation__.Notify(570790)
	skipLen, err := encoding.PeekLength(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(570792)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(570793)
	}
	__antithesis_instrumentation__.Notify(570791)
	return key[skipLen:], nil
}
