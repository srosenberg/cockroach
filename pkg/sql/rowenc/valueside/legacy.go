package valueside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func MarshalLegacy(colType *types.T, val tree.Datum) (roachpb.Value, error) {
	__antithesis_instrumentation__.Notify(571410)
	var r roachpb.Value

	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(571413)
		return r, nil
	} else {
		__antithesis_instrumentation__.Notify(571414)
	}
	__antithesis_instrumentation__.Notify(571411)

	switch colType.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(571415)
		if v, ok := val.(*tree.DBitArray); ok {
			__antithesis_instrumentation__.Notify(571439)
			r.SetBitArray(v.BitArray)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571440)
		}
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(571416)
		if v, ok := val.(*tree.DBool); ok {
			__antithesis_instrumentation__.Notify(571441)
			r.SetBool(bool(*v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571442)
		}
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(571417)
		if v, ok := tree.AsDInt(val); ok {
			__antithesis_instrumentation__.Notify(571443)
			r.SetInt(int64(v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571444)
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(571418)
		if v, ok := val.(*tree.DFloat); ok {
			__antithesis_instrumentation__.Notify(571445)
			r.SetFloat(float64(*v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571446)
		}
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(571419)
		if v, ok := val.(*tree.DDecimal); ok {
			__antithesis_instrumentation__.Notify(571447)
			err := r.SetDecimal(&v.Decimal)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(571448)
		}
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(571420)
		if v, ok := tree.AsDString(val); ok {
			__antithesis_instrumentation__.Notify(571449)
			r.SetString(string(v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571450)
		}
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(571421)
		if v, ok := val.(*tree.DBytes); ok {
			__antithesis_instrumentation__.Notify(571451)
			r.SetString(string(*v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571452)
		}
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(571422)
		if v, ok := val.(*tree.DDate); ok {
			__antithesis_instrumentation__.Notify(571453)
			r.SetInt(v.UnixEpochDaysWithOrig())
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571454)
		}
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(571423)
		if v, ok := val.(*tree.DBox2D); ok {
			__antithesis_instrumentation__.Notify(571455)
			r.SetBox2D(v.CartesianBoundingBox.BoundingBox)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571456)
		}
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(571424)
		if v, ok := val.(*tree.DGeography); ok {
			__antithesis_instrumentation__.Notify(571457)
			err := r.SetGeo(v.SpatialObject())
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(571458)
		}
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(571425)
		if v, ok := val.(*tree.DGeometry); ok {
			__antithesis_instrumentation__.Notify(571459)
			err := r.SetGeo(v.SpatialObject())
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(571460)
		}
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(571426)
		if v, ok := val.(*tree.DTime); ok {
			__antithesis_instrumentation__.Notify(571461)
			r.SetInt(int64(*v))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571462)
		}
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(571427)
		if v, ok := val.(*tree.DTimeTZ); ok {
			__antithesis_instrumentation__.Notify(571463)
			r.SetTimeTZ(v.TimeTZ)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571464)
		}
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(571428)
		if v, ok := val.(*tree.DTimestamp); ok {
			__antithesis_instrumentation__.Notify(571465)
			r.SetTime(v.Time)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571466)
		}
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(571429)
		if v, ok := val.(*tree.DTimestampTZ); ok {
			__antithesis_instrumentation__.Notify(571467)
			r.SetTime(v.Time)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571468)
		}
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(571430)
		if v, ok := val.(*tree.DInterval); ok {
			__antithesis_instrumentation__.Notify(571469)
			err := r.SetDuration(v.Duration)
			return r, err
		} else {
			__antithesis_instrumentation__.Notify(571470)
		}
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(571431)
		if v, ok := val.(*tree.DUuid); ok {
			__antithesis_instrumentation__.Notify(571471)
			r.SetBytes(v.GetBytes())
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571472)
		}
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(571432)
		if v, ok := val.(*tree.DIPAddr); ok {
			__antithesis_instrumentation__.Notify(571473)
			data := v.ToBuffer(nil)
			r.SetBytes(data)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571474)
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(571433)
		if v, ok := val.(*tree.DJSON); ok {
			__antithesis_instrumentation__.Notify(571475)
			data, err := json.EncodeJSON(nil, v.JSON)
			if err != nil {
				__antithesis_instrumentation__.Notify(571477)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(571478)
			}
			__antithesis_instrumentation__.Notify(571476)
			r.SetBytes(data)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571479)
		}
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(571434)
		if v, ok := val.(*tree.DArray); ok {
			__antithesis_instrumentation__.Notify(571480)
			if err := checkElementType(v.ParamTyp, colType.ArrayContents()); err != nil {
				__antithesis_instrumentation__.Notify(571483)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(571484)
			}
			__antithesis_instrumentation__.Notify(571481)
			b, err := encodeArray(v, nil)
			if err != nil {
				__antithesis_instrumentation__.Notify(571485)
				return r, err
			} else {
				__antithesis_instrumentation__.Notify(571486)
			}
			__antithesis_instrumentation__.Notify(571482)
			r.SetBytes(b)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571487)
		}
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(571435)
		if v, ok := val.(*tree.DCollatedString); ok {
			__antithesis_instrumentation__.Notify(571488)
			if lex.LocaleNamesAreEqual(v.Locale, colType.Locale()) {
				__antithesis_instrumentation__.Notify(571490)
				r.SetString(v.Contents)
				return r, nil
			} else {
				__antithesis_instrumentation__.Notify(571491)
			}
			__antithesis_instrumentation__.Notify(571489)

			return r, errors.AssertionFailedf(
				"locale mismatch %q vs %q",
				v.Locale, colType.Locale(),
			)
		} else {
			__antithesis_instrumentation__.Notify(571492)
		}
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(571436)
		if v, ok := val.(*tree.DOid); ok {
			__antithesis_instrumentation__.Notify(571493)
			r.SetInt(int64(v.DInt))
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571494)
		}
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(571437)
		if v, ok := val.(*tree.DEnum); ok {
			__antithesis_instrumentation__.Notify(571495)
			r.SetBytes(v.PhysicalRep)
			return r, nil
		} else {
			__antithesis_instrumentation__.Notify(571496)
		}
	default:
		__antithesis_instrumentation__.Notify(571438)
		return r, errors.AssertionFailedf("unsupported column type: %s", colType.Family())
	}
	__antithesis_instrumentation__.Notify(571412)
	return r, errors.AssertionFailedf("mismatched type %q vs %q", val.ResolvedType(), colType.Family())
}

func UnmarshalLegacy(a *tree.DatumAlloc, typ *types.T, value roachpb.Value) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(571497)
	if value.RawBytes == nil {
		__antithesis_instrumentation__.Notify(571499)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(571500)
	}
	__antithesis_instrumentation__.Notify(571498)

	switch typ.Family() {
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(571501)
		d, err := value.GetBitArray()
		if err != nil {
			__antithesis_instrumentation__.Notify(571553)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571554)
		}
		__antithesis_instrumentation__.Notify(571502)
		return a.NewDBitArray(tree.DBitArray{BitArray: d}), nil
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(571503)
		v, err := value.GetBool()
		if err != nil {
			__antithesis_instrumentation__.Notify(571555)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571556)
		}
		__antithesis_instrumentation__.Notify(571504)
		return tree.MakeDBool(tree.DBool(v)), nil
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(571505)
		v, err := value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(571557)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571558)
		}
		__antithesis_instrumentation__.Notify(571506)
		return a.NewDInt(tree.DInt(v)), nil
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(571507)
		v, err := value.GetFloat()
		if err != nil {
			__antithesis_instrumentation__.Notify(571559)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571560)
		}
		__antithesis_instrumentation__.Notify(571508)
		return a.NewDFloat(tree.DFloat(v)), nil
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(571509)
		v, err := value.GetDecimal()
		if err != nil {
			__antithesis_instrumentation__.Notify(571561)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571562)
		}
		__antithesis_instrumentation__.Notify(571510)
		dd := a.NewDDecimal(tree.DDecimal{Decimal: v})
		return dd, nil
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(571511)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571563)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571564)
		}
		__antithesis_instrumentation__.Notify(571512)
		if typ.Oid() == oid.T_name {
			__antithesis_instrumentation__.Notify(571565)
			return a.NewDName(tree.DString(v)), nil
		} else {
			__antithesis_instrumentation__.Notify(571566)
		}
		__antithesis_instrumentation__.Notify(571513)
		return a.NewDString(tree.DString(v)), nil
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(571514)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571567)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571568)
		}
		__antithesis_instrumentation__.Notify(571515)
		return a.NewDBytes(tree.DBytes(v)), nil
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(571516)
		v, err := value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(571569)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571570)
		}
		__antithesis_instrumentation__.Notify(571517)
		return a.NewDDate(tree.MakeDDate(pgdate.MakeCompatibleDateFromDisk(v))), nil
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(571518)
		v, err := value.GetBox2D()
		if err != nil {
			__antithesis_instrumentation__.Notify(571571)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571572)
		}
		__antithesis_instrumentation__.Notify(571519)
		return a.NewDBox2D(tree.DBox2D{
			CartesianBoundingBox: geo.CartesianBoundingBox{BoundingBox: v},
		}), nil
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(571520)
		v, err := value.GetGeo()
		if err != nil {
			__antithesis_instrumentation__.Notify(571573)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571574)
		}
		__antithesis_instrumentation__.Notify(571521)
		return a.NewDGeography(tree.DGeography{Geography: geo.MakeGeographyUnsafe(v)}), nil
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(571522)
		v, err := value.GetGeo()
		if err != nil {
			__antithesis_instrumentation__.Notify(571575)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571576)
		}
		__antithesis_instrumentation__.Notify(571523)
		return a.NewDGeometry(tree.DGeometry{Geometry: geo.MakeGeometryUnsafe(v)}), nil
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(571524)
		v, err := value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(571577)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571578)
		}
		__antithesis_instrumentation__.Notify(571525)
		return a.NewDTime(tree.DTime(v)), nil
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(571526)
		v, err := value.GetTimeTZ()
		if err != nil {
			__antithesis_instrumentation__.Notify(571579)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571580)
		}
		__antithesis_instrumentation__.Notify(571527)
		return a.NewDTimeTZ(tree.DTimeTZ{TimeTZ: v}), nil
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(571528)
		v, err := value.GetTime()
		if err != nil {
			__antithesis_instrumentation__.Notify(571581)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571582)
		}
		__antithesis_instrumentation__.Notify(571529)
		return a.NewDTimestamp(tree.DTimestamp{Time: v}), nil
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(571530)
		v, err := value.GetTime()
		if err != nil {
			__antithesis_instrumentation__.Notify(571583)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571584)
		}
		__antithesis_instrumentation__.Notify(571531)
		return a.NewDTimestampTZ(tree.DTimestampTZ{Time: v}), nil
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(571532)
		d, err := value.GetDuration()
		if err != nil {
			__antithesis_instrumentation__.Notify(571585)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571586)
		}
		__antithesis_instrumentation__.Notify(571533)
		return a.NewDInterval(tree.DInterval{Duration: d}), nil
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(571534)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571587)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571588)
		}
		__antithesis_instrumentation__.Notify(571535)
		return a.NewDCollatedString(string(v), typ.Locale())
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(571536)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571589)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571590)
		}
		__antithesis_instrumentation__.Notify(571537)
		u, err := uuid.FromBytes(v)
		if err != nil {
			__antithesis_instrumentation__.Notify(571591)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571592)
		}
		__antithesis_instrumentation__.Notify(571538)
		return a.NewDUuid(tree.DUuid{UUID: u}), nil
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(571539)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571593)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571594)
		}
		__antithesis_instrumentation__.Notify(571540)
		var ipAddr ipaddr.IPAddr
		_, err = ipAddr.FromBuffer(v)
		if err != nil {
			__antithesis_instrumentation__.Notify(571595)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571596)
		}
		__antithesis_instrumentation__.Notify(571541)
		return a.NewDIPAddr(tree.DIPAddr{IPAddr: ipAddr}), nil
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(571542)
		v, err := value.GetInt()
		if err != nil {
			__antithesis_instrumentation__.Notify(571597)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571598)
		}
		__antithesis_instrumentation__.Notify(571543)
		return a.NewDOid(tree.MakeDOid(tree.DInt(v))), nil
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(571544)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571599)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571600)
		}
		__antithesis_instrumentation__.Notify(571545)
		datum, _, err := decodeArray(a, typ, v)

		return datum, err
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(571546)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571601)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571602)
		}
		__antithesis_instrumentation__.Notify(571547)
		_, jsonDatum, err := json.DecodeJSON(v)
		if err != nil {
			__antithesis_instrumentation__.Notify(571603)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571604)
		}
		__antithesis_instrumentation__.Notify(571548)
		return tree.NewDJSON(jsonDatum), nil
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(571549)
		v, err := value.GetBytes()
		if err != nil {
			__antithesis_instrumentation__.Notify(571605)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571606)
		}
		__antithesis_instrumentation__.Notify(571550)
		phys, log, err := tree.GetEnumComponentsFromPhysicalRep(typ, v)
		if err != nil {
			__antithesis_instrumentation__.Notify(571607)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571608)
		}
		__antithesis_instrumentation__.Notify(571551)
		return a.NewDEnum(tree.DEnum{EnumTyp: typ, PhysicalRep: phys, LogicalRep: log}), nil
	default:
		__antithesis_instrumentation__.Notify(571552)
		return nil, errors.Errorf("unsupported column type: %s", typ.Family())
	}
}
