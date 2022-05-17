package valueside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/errors"
)

func Decode(
	a *tree.DatumAlloc, valType *types.T, b []byte,
) (_ tree.Datum, remaining []byte, _ error) {
	__antithesis_instrumentation__.Notify(571254)
	_, dataOffset, _, typ, err := encoding.DecodeValueTag(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(571258)
		return nil, b, err
	} else {
		__antithesis_instrumentation__.Notify(571259)
	}
	__antithesis_instrumentation__.Notify(571255)

	if typ == encoding.Null {
		__antithesis_instrumentation__.Notify(571260)
		return tree.DNull, b[dataOffset:], nil
	} else {
		__antithesis_instrumentation__.Notify(571261)
	}
	__antithesis_instrumentation__.Notify(571256)

	if valType.Family() != types.BoolFamily {
		__antithesis_instrumentation__.Notify(571262)
		b = b[dataOffset:]
	} else {
		__antithesis_instrumentation__.Notify(571263)
	}
	__antithesis_instrumentation__.Notify(571257)
	return DecodeUntaggedDatum(a, valType, b)
}

func DecodeUntaggedDatum(
	a *tree.DatumAlloc, t *types.T, buf []byte,
) (_ tree.Datum, remaining []byte, _ error) {
	__antithesis_instrumentation__.Notify(571264)
	switch t.Family() {
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(571265)
		b, i, err := encoding.DecodeUntaggedIntValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571311)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571312)
		}
		__antithesis_instrumentation__.Notify(571266)
		return a.NewDInt(tree.DInt(i)), b, nil
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(571267)
		b, data, err := encoding.DecodeUntaggedBytesValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571313)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571314)
		}
		__antithesis_instrumentation__.Notify(571268)
		return a.NewDString(tree.DString(data)), b, nil
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(571269)
		b, data, err := encoding.DecodeUntaggedBytesValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571315)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571316)
		}
		__antithesis_instrumentation__.Notify(571270)
		d, err := a.NewDCollatedString(string(data), t.Locale())
		return d, b, err
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(571271)
		b, data, err := encoding.DecodeUntaggedBitArrayValue(buf)
		return a.NewDBitArray(tree.DBitArray{BitArray: data}), b, err
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(571272)

		b, data, err := encoding.DecodeBoolValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571317)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571318)
		}
		__antithesis_instrumentation__.Notify(571273)
		return tree.MakeDBool(tree.DBool(data)), b, nil
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(571274)
		b, data, err := encoding.DecodeUntaggedFloatValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571319)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571320)
		}
		__antithesis_instrumentation__.Notify(571275)
		return a.NewDFloat(tree.DFloat(data)), b, nil
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(571276)
		b, data, err := encoding.DecodeUntaggedDecimalValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571321)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571322)
		}
		__antithesis_instrumentation__.Notify(571277)
		return a.NewDDecimal(tree.DDecimal{Decimal: data}), b, nil
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(571278)
		b, data, err := encoding.DecodeUntaggedBytesValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571323)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571324)
		}
		__antithesis_instrumentation__.Notify(571279)
		return a.NewDBytes(tree.DBytes(data)), b, nil
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(571280)
		b, data, err := encoding.DecodeUntaggedIntValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571325)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571326)
		}
		__antithesis_instrumentation__.Notify(571281)
		return a.NewDDate(tree.MakeDDate(pgdate.MakeCompatibleDateFromDisk(data))), b, nil
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(571282)
		b, data, err := encoding.DecodeUntaggedBox2DValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571327)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571328)
		}
		__antithesis_instrumentation__.Notify(571283)
		return a.NewDBox2D(tree.DBox2D{
			CartesianBoundingBox: geo.CartesianBoundingBox{BoundingBox: data},
		}), b, nil
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(571284)
		g := a.NewDGeographyEmpty()
		so := g.Geography.SpatialObjectRef()
		b, err := encoding.DecodeUntaggedGeoValue(buf, so)
		a.DoneInitNewDGeo(so)
		if err != nil {
			__antithesis_instrumentation__.Notify(571329)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571330)
		}
		__antithesis_instrumentation__.Notify(571285)
		return g, b, nil
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(571286)
		g := a.NewDGeometryEmpty()
		so := g.Geometry.SpatialObjectRef()
		b, err := encoding.DecodeUntaggedGeoValue(buf, so)
		a.DoneInitNewDGeo(so)
		if err != nil {
			__antithesis_instrumentation__.Notify(571331)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571332)
		}
		__antithesis_instrumentation__.Notify(571287)
		return g, b, nil
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(571288)
		b, data, err := encoding.DecodeUntaggedIntValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571333)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571334)
		}
		__antithesis_instrumentation__.Notify(571289)
		return a.NewDTime(tree.DTime(data)), b, nil
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(571290)
		b, data, err := encoding.DecodeUntaggedTimeTZValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571335)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571336)
		}
		__antithesis_instrumentation__.Notify(571291)
		return a.NewDTimeTZ(tree.DTimeTZ{TimeTZ: data}), b, nil
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(571292)
		b, data, err := encoding.DecodeUntaggedTimeValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571337)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571338)
		}
		__antithesis_instrumentation__.Notify(571293)
		return a.NewDTimestamp(tree.DTimestamp{Time: data}), b, nil
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(571294)
		b, data, err := encoding.DecodeUntaggedTimeValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571339)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571340)
		}
		__antithesis_instrumentation__.Notify(571295)
		return a.NewDTimestampTZ(tree.DTimestampTZ{Time: data}), b, nil
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(571296)
		b, data, err := encoding.DecodeUntaggedDurationValue(buf)
		return a.NewDInterval(tree.DInterval{Duration: data}), b, err
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(571297)
		b, data, err := encoding.DecodeUntaggedUUIDValue(buf)
		return a.NewDUuid(tree.DUuid{UUID: data}), b, err
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(571298)
		b, data, err := encoding.DecodeUntaggedIPAddrValue(buf)
		return a.NewDIPAddr(tree.DIPAddr{IPAddr: data}), b, err
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(571299)
		b, data, err := encoding.DecodeUntaggedBytesValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571341)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571342)
		}
		__antithesis_instrumentation__.Notify(571300)

		cpy := make([]byte, len(data))
		copy(cpy, data)
		j, err := json.FromEncoding(cpy)
		if err != nil {
			__antithesis_instrumentation__.Notify(571343)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571344)
		}
		__antithesis_instrumentation__.Notify(571301)
		return a.NewDJSON(tree.DJSON{JSON: j}), b, nil
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(571302)
		b, data, err := encoding.DecodeUntaggedIntValue(buf)
		return a.NewDOid(tree.MakeDOid(tree.DInt(data))), b, err
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(571303)

		b, _, _, err := encoding.DecodeNonsortingUvarint(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571345)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(571346)
		}
		__antithesis_instrumentation__.Notify(571304)
		return decodeArray(a, t, b)
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(571305)
		return decodeTuple(a, t, buf)
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(571306)
		b, data, err := encoding.DecodeUntaggedBytesValue(buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(571347)
			return nil, b, err
		} else {
			__antithesis_instrumentation__.Notify(571348)
		}
		__antithesis_instrumentation__.Notify(571307)
		phys, log, err := tree.GetEnumComponentsFromPhysicalRep(t, data)
		if err != nil {
			__antithesis_instrumentation__.Notify(571349)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(571350)
		}
		__antithesis_instrumentation__.Notify(571308)
		return a.NewDEnum(tree.DEnum{EnumTyp: t, PhysicalRep: phys, LogicalRep: log}), b, nil
	case types.VoidFamily:
		__antithesis_instrumentation__.Notify(571309)
		return a.NewDVoid(), buf, nil
	default:
		__antithesis_instrumentation__.Notify(571310)
		return nil, buf, errors.Errorf("couldn't decode type %s", t)
	}
}

type Decoder struct {
	colIdxMap catalog.TableColMap
	types     []*types.T
}

func MakeDecoder(cols []catalog.Column) Decoder {
	__antithesis_instrumentation__.Notify(571351)
	var d Decoder
	d.types = make([]*types.T, len(cols))
	for i, col := range cols {
		__antithesis_instrumentation__.Notify(571353)
		d.colIdxMap.Set(col.GetID(), i)
		d.types[i] = col.GetType()
	}
	__antithesis_instrumentation__.Notify(571352)
	return d
}

func (d *Decoder) Decode(a *tree.DatumAlloc, bytes []byte) (tree.Datums, error) {
	__antithesis_instrumentation__.Notify(571354)
	datums := make(tree.Datums, len(d.types))
	for i := range datums {
		__antithesis_instrumentation__.Notify(571357)
		datums[i] = tree.DNull
	}
	__antithesis_instrumentation__.Notify(571355)

	var lastColID descpb.ColumnID
	for len(bytes) > 0 {
		__antithesis_instrumentation__.Notify(571358)
		_, dataOffset, colIDDiff, typ, err := encoding.DecodeValueTag(bytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(571361)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571362)
		}
		__antithesis_instrumentation__.Notify(571359)
		colID := lastColID + descpb.ColumnID(colIDDiff)
		lastColID = colID
		idx, ok := d.colIdxMap.Get(colID)
		if !ok {
			__antithesis_instrumentation__.Notify(571363)

			l, err := encoding.PeekValueLengthWithOffsetsAndType(bytes, dataOffset, typ)
			if err != nil {
				__antithesis_instrumentation__.Notify(571365)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(571366)
			}
			__antithesis_instrumentation__.Notify(571364)
			bytes = bytes[l:]
			continue
		} else {
			__antithesis_instrumentation__.Notify(571367)
		}
		__antithesis_instrumentation__.Notify(571360)
		datums[idx], bytes, err = Decode(a, d.types[idx], bytes)
		if err != nil {
			__antithesis_instrumentation__.Notify(571368)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571369)
		}
	}
	__antithesis_instrumentation__.Notify(571356)
	return datums, nil
}
