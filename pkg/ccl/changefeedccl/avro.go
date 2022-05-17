package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/json"
	"math/big"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/errors"
	"github.com/linkedin/goavro/v2"
)

type avroSchemaType interface{}

const (
	avroSchemaArray   = `array`
	avroSchemaBoolean = `boolean`
	avroSchemaBytes   = `bytes`
	avroSchemaDouble  = `double`
	avroSchemaInt     = `int`
	avroSchemaLong    = `long`
	avroSchemaNull    = `null`
	avroSchemaString  = `string`
)

type avroLogicalType struct {
	SchemaType  avroSchemaType `json:"type"`
	LogicalType string         `json:"logicalType"`
	Precision   int            `json:"precision,omitempty"`
	Scale       int            `json:"scale,omitempty"`
}

type avroArrayType struct {
	SchemaType avroSchemaType `json:"type"`
	Items      avroSchemaType `json:"items"`
}

func avroUnionKey(t avroSchemaType) string {
	__antithesis_instrumentation__.Notify(14334)
	switch s := t.(type) {
	case string:
		__antithesis_instrumentation__.Notify(14335)
		return s
	case avroLogicalType:
		__antithesis_instrumentation__.Notify(14336)
		return avroUnionKey(s.SchemaType) + `.` + s.LogicalType
	case avroArrayType:
		__antithesis_instrumentation__.Notify(14337)
		return avroUnionKey(s.SchemaType)
	case *avroRecord:
		__antithesis_instrumentation__.Notify(14338)
		if s.Namespace == "" {
			__antithesis_instrumentation__.Notify(14341)
			return s.Name
		} else {
			__antithesis_instrumentation__.Notify(14342)
		}
		__antithesis_instrumentation__.Notify(14339)
		return s.Namespace + `.` + s.Name
	default:
		__antithesis_instrumentation__.Notify(14340)
		panic(errors.AssertionFailedf(`unsupported type %T %v`, t, t))
	}
}

type datumToNativeFn func(datum tree.Datum, memo interface{}) (interface{}, error)

type avroSchemaField struct {
	SchemaType avroSchemaType `json:"type"`
	Name       string         `json:"name"`
	Default    *string        `json:"default"`
	Metadata   string         `json:"__crdb__,omitempty"`
	Namespace  string         `json:"namespace,omitempty"`

	typ *types.T

	encodeFn func(datum tree.Datum) (interface{}, error)

	encodeDatum datumToNativeFn

	decodeFn func(interface{}) (tree.Datum, error)

	nativeEncoded              map[string]interface{}
	nativeEncodedSecondaryType map[string]interface{}
}

type avroRecord struct {
	SchemaType string             `json:"type"`
	Name       string             `json:"name"`
	Fields     []*avroSchemaField `json:"fields"`
	Namespace  string             `json:"namespace,omitempty"`
	codec      *goavro.Codec
}

type avroDataRecord struct {
	avroRecord

	colIdxByFieldIdx map[int]int
	fieldIdxByName   map[string]int
	fieldIdxByColIdx map[int]int

	native map[string]interface{}
	alloc  tree.DatumAlloc
}

type avroMetadata map[string]interface{}

type avroEnvelopeOpts struct {
	beforeField, afterField     bool
	updatedField, resolvedField bool
}

type avroEnvelopeRecord struct {
	avroRecord

	opts          avroEnvelopeOpts
	before, after *avroDataRecord
}

func typeToAvroSchema(typ *types.T) (*avroSchemaField, error) {
	__antithesis_instrumentation__.Notify(14343)
	schema := &avroSchemaField{
		typ: typ,
	}

	setNullable := func(
		avroType avroSchemaType,
		encoder datumToNativeFn,
		decoder func(interface{}) (tree.Datum, error),
	) {
		__antithesis_instrumentation__.Notify(14347)

		schema.SchemaType = []avroSchemaType{avroSchemaNull, avroType}
		unionKey := avroUnionKey(avroType)
		schema.nativeEncoded = map[string]interface{}{unionKey: nil}
		schema.encodeDatum = encoder

		schema.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(14349)
			if d == tree.DNull {
				__antithesis_instrumentation__.Notify(14352)
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(14353)
			}
			__antithesis_instrumentation__.Notify(14350)
			encoded, err := encoder(d, schema.nativeEncoded[unionKey])
			if err != nil {
				__antithesis_instrumentation__.Notify(14354)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14355)
			}
			__antithesis_instrumentation__.Notify(14351)
			schema.nativeEncoded[unionKey] = encoded
			return schema.nativeEncoded, nil
		}
		__antithesis_instrumentation__.Notify(14348)
		schema.decodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(14356)
			if x == nil {
				__antithesis_instrumentation__.Notify(14358)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(14359)
			}
			__antithesis_instrumentation__.Notify(14357)
			return decoder(x.(map[string]interface{})[unionKey])
		}
	}
	__antithesis_instrumentation__.Notify(14344)

	setNullableWithStringFallback := func(
		avroType avroSchemaType,
		encoder datumToNativeFn,
		decoder func(interface{}) (tree.Datum, error),
	) {
		__antithesis_instrumentation__.Notify(14360)
		schema.SchemaType = []avroSchemaType{avroSchemaNull, avroType, avroSchemaString}
		mainUnionKey := avroUnionKey(avroType)
		stringUnionKey := avroUnionKey(avroSchemaString)
		schema.nativeEncoded = map[string]interface{}{mainUnionKey: nil}
		schema.nativeEncodedSecondaryType = map[string]interface{}{stringUnionKey: nil}
		schema.encodeDatum = encoder

		schema.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(14362)
			if d == tree.DNull {
				__antithesis_instrumentation__.Notify(14366)
				return nil, nil
			} else {
				__antithesis_instrumentation__.Notify(14367)
			}
			__antithesis_instrumentation__.Notify(14363)
			encoded, err := encoder(d, schema.nativeEncoded[mainUnionKey])
			if err != nil {
				__antithesis_instrumentation__.Notify(14368)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14369)
			}
			__antithesis_instrumentation__.Notify(14364)
			_, isString := encoded.(string)
			if isString {
				__antithesis_instrumentation__.Notify(14370)
				schema.nativeEncodedSecondaryType[stringUnionKey] = encoded
				return schema.nativeEncodedSecondaryType, nil
			} else {
				__antithesis_instrumentation__.Notify(14371)
			}
			__antithesis_instrumentation__.Notify(14365)
			schema.nativeEncoded[mainUnionKey] = encoded
			return schema.nativeEncoded, nil
		}
		__antithesis_instrumentation__.Notify(14361)
		schema.decodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(14372)
			if x == nil {
				__antithesis_instrumentation__.Notify(14374)
				return tree.DNull, nil
			} else {
				__antithesis_instrumentation__.Notify(14375)
			}
			__antithesis_instrumentation__.Notify(14373)
			return decoder(x.(map[string]interface{}))
		}
	}
	__antithesis_instrumentation__.Notify(14345)

	switch typ.Family() {
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(14376)
		setNullable(
			avroSchemaLong,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14401)
				return int64(*d.(*tree.DInt)), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14402)
				return tree.NewDInt(tree.DInt(x.(int64))), nil
			},
		)
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(14377)
		setNullable(
			avroSchemaBoolean,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14403)
				return bool(*d.(*tree.DBool)), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14404)
				return tree.MakeDBool(tree.DBool(x.(bool))), nil
			},
		)
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(14378)
		setNullable(
			avroArrayType{
				SchemaType: avroSchemaArray,
				Items:      avroSchemaLong,
			},
			func(d tree.Datum, memo interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14405)
				uints, lastBitsUsed := d.(*tree.DBitArray).EncodingParts()
				var signedLongs []interface{}

				if memo != nil {
					__antithesis_instrumentation__.Notify(14409)
					signedLongs = memo.([]interface{})
					if len(signedLongs) > len(uints)+1 {
						__antithesis_instrumentation__.Notify(14410)
						signedLongs = signedLongs[:len(uints)+1]
					} else {
						__antithesis_instrumentation__.Notify(14411)
					}
				} else {
					__antithesis_instrumentation__.Notify(14412)
				}
				__antithesis_instrumentation__.Notify(14406)
				if signedLongs == nil {
					__antithesis_instrumentation__.Notify(14413)
					signedLongs = make([]interface{}, len(uints)+1)
				} else {
					__antithesis_instrumentation__.Notify(14414)
				}
				__antithesis_instrumentation__.Notify(14407)
				signedLongs[0] = int64(lastBitsUsed)
				for idx, word := range uints {
					__antithesis_instrumentation__.Notify(14415)
					signedLongs[idx+1] = int64(word)
				}
				__antithesis_instrumentation__.Notify(14408)
				return signedLongs, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14416)
				arr := x.([]interface{})
				lastBitsUsed, ints := arr[0], arr[1:]
				uints := make([]uint64, len(ints))
				for idx, word := range ints {
					__antithesis_instrumentation__.Notify(14418)
					uints[idx] = uint64(word.(int64))
				}
				__antithesis_instrumentation__.Notify(14417)
				ba, err := bitarray.FromEncodingParts(uints, uint64(lastBitsUsed.(int64)))
				return &tree.DBitArray{BitArray: ba}, err
			},
		)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(14379)
		setNullable(
			avroSchemaDouble,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14419)
				return float64(*d.(*tree.DFloat)), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14420)
				return tree.NewDFloat(tree.DFloat(x.(float64))), nil
			},
		)
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(14380)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14421)
				return d.(*tree.DBox2D).CartesianBoundingBox.Repr(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14422)
				b, err := geo.ParseCartesianBoundingBox(x.(string))
				if err != nil {
					__antithesis_instrumentation__.Notify(14424)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(14425)
				}
				__antithesis_instrumentation__.Notify(14423)
				return tree.NewDBox2D(b), nil
			},
		)
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(14381)
		setNullable(
			avroSchemaBytes,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14426)
				return []byte(d.(*tree.DGeography).EWKB()), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14427)
				g, err := geo.ParseGeographyFromEWKBUnsafe(geopb.EWKB(x.([]byte)))
				if err != nil {
					__antithesis_instrumentation__.Notify(14429)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(14430)
				}
				__antithesis_instrumentation__.Notify(14428)
				return &tree.DGeography{Geography: g}, nil
			},
		)
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(14382)
		setNullable(
			avroSchemaBytes,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14431)
				return []byte(d.(*tree.DGeometry).EWKB()), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14432)
				g, err := geo.ParseGeometryFromEWKBUnsafe(geopb.EWKB(x.([]byte)))
				if err != nil {
					__antithesis_instrumentation__.Notify(14434)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(14435)
				}
				__antithesis_instrumentation__.Notify(14433)
				return &tree.DGeometry{Geometry: g}, nil
			},
		)
	case types.StringFamily:
		__antithesis_instrumentation__.Notify(14383)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14436)
				return string(*d.(*tree.DString)), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14437)
				return tree.NewDString(x.(string)), nil
			},
		)
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(14384)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14438)
				return d.(*tree.DCollatedString).Contents, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14439)
				return tree.NewDCollatedString(x.(string), typ.Locale(), &tree.CollationEnvironment{})
			},
		)
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(14385)
		setNullable(
			avroSchemaBytes,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14440)
				return []byte(*d.(*tree.DBytes)), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14441)
				return tree.NewDBytes(tree.DBytes(x.([]byte))), nil
			},
		)
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(14386)
		setNullable(
			avroLogicalType{
				SchemaType:  avroSchemaInt,
				LogicalType: `date`,
			},
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14442)
				date := *d.(*tree.DDate)
				if !date.IsFinite() {
					__antithesis_instrumentation__.Notify(14444)
					return nil, errors.Errorf(
						`infinite date not yet supported with avro`)
				} else {
					__antithesis_instrumentation__.Notify(14445)
				}
				__antithesis_instrumentation__.Notify(14443)

				return date.ToTime()
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14446)

				return tree.NewDDateFromTime(x.(time.Time))
			},
		)
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(14387)
		setNullable(
			avroLogicalType{
				SchemaType:  avroSchemaLong,
				LogicalType: `time-micros`,
			},
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14447)

				time := d.(*tree.DTime)
				return int64(*time), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14448)

				micros := x.(time.Duration) / time.Microsecond
				return tree.MakeDTime(timeofday.TimeOfDay(micros)), nil
			},
		)
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(14388)
		setNullable(
			avroSchemaString,

			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14449)
				return d.(*tree.DTimeTZ).TimeTZ.String(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14450)
				d, _, err := tree.ParseDTimeTZ(nil, x.(string), time.Microsecond)
				return d, err
			},
		)
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(14389)
		setNullable(
			avroLogicalType{
				SchemaType:  avroSchemaLong,
				LogicalType: `timestamp-micros`,
			},
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14451)
				return d.(*tree.DTimestamp).Time, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14452)
				return tree.MakeDTimestamp(x.(time.Time), time.Microsecond)
			},
		)
	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(14390)
		setNullable(
			avroLogicalType{
				SchemaType:  avroSchemaLong,
				LogicalType: `timestamp-micros`,
			},
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14453)
				return d.(*tree.DTimestampTZ).Time, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14454)
				return tree.MakeDTimestampTZ(x.(time.Time), time.Microsecond)
			},
		)
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(14391)
		setNullable(

			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14455)
				return d.(*tree.DInterval).ValueAsISO8601String(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14456)
				return tree.ParseDInterval(duration.IntervalStyle_ISO_8601, x.(string))
			},
		)
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(14392)
		if typ.Precision() == 0 {
			__antithesis_instrumentation__.Notify(14457)
			return nil, errors.Errorf(
				`decimal with no precision not yet supported with avro`)
		} else {
			__antithesis_instrumentation__.Notify(14458)
		}
		__antithesis_instrumentation__.Notify(14393)

		width := int(typ.Width())
		prec := int(typ.Precision())
		decimalType := avroLogicalType{
			SchemaType:  avroSchemaBytes,
			LogicalType: `decimal`,
			Precision:   prec,
			Scale:       width,
		}
		setNullableWithStringFallback(
			decimalType,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14459)
				dec := d.(*tree.DDecimal).Decimal

				if dec.Form != apd.Finite {
					__antithesis_instrumentation__.Notify(14463)
					return d.String(), nil
				} else {
					__antithesis_instrumentation__.Notify(14464)
				}
				__antithesis_instrumentation__.Notify(14460)

				if typ.Width() > -dec.Exponent {
					__antithesis_instrumentation__.Notify(14465)
					_, err := tree.DecimalCtx.WithPrecision(uint32(prec)).Quantize(&dec, &dec, -int32(width))
					if err != nil {
						__antithesis_instrumentation__.Notify(14466)

						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(14467)
					}
				} else {
					__antithesis_instrumentation__.Notify(14468)
				}
				__antithesis_instrumentation__.Notify(14461)

				rat, err := decimalToRat(dec, int32(width))
				if err != nil {
					__antithesis_instrumentation__.Notify(14469)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(14470)
				}
				__antithesis_instrumentation__.Notify(14462)
				return &rat, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14471)
				unionMap := x.(map[string]interface{})
				rat, ok := unionMap[avroUnionKey(decimalType)]
				if ok {
					__antithesis_instrumentation__.Notify(14473)
					return &tree.DDecimal{Decimal: ratToDecimal(*rat.(*big.Rat), int32(width))}, nil
				} else {
					__antithesis_instrumentation__.Notify(14474)
				}
				__antithesis_instrumentation__.Notify(14472)
				return tree.ParseDDecimal(unionMap[avroUnionKey(avroSchemaString)].(string))
			},
		)
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(14394)

		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14475)
				return d.(*tree.DUuid).UUID.String(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14476)
				return tree.ParseDUuidFromString(x.(string))
			},
		)
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(14395)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14477)
				return d.(*tree.DIPAddr).IPAddr.String(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14478)
				return tree.ParseDIPAddrFromINetString(x.(string))
			},
		)
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(14396)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14479)
				return d.(*tree.DJSON).JSON.String(), nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14480)
				return tree.ParseDJSON(x.(string))
			},
		)
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(14397)
		setNullable(
			avroSchemaString,
			func(d tree.Datum, _ interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14481)
				return d.(*tree.DEnum).LogicalRep, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14482)
				return tree.MakeDEnumFromLogicalRepresentation(typ, x.(string))
			},
		)
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(14398)
		itemSchema, err := typeToAvroSchema(typ.ArrayContents())
		if err != nil {
			__antithesis_instrumentation__.Notify(14483)
			return nil, errors.Wrapf(err, `could not create item schema for %s`,
				typ)
		} else {
			__antithesis_instrumentation__.Notify(14484)
		}
		__antithesis_instrumentation__.Notify(14399)
		itemUnionKey := avroUnionKey(itemSchema.SchemaType.([]avroSchemaType)[1])

		setNullable(
			avroArrayType{
				SchemaType: avroSchemaArray,
				Items:      itemSchema.SchemaType,
			},
			func(d tree.Datum, memo interface{}) (interface{}, error) {
				__antithesis_instrumentation__.Notify(14485)
				datumArr := d.(*tree.DArray)
				var avroArr []interface{}
				if memo != nil {
					__antithesis_instrumentation__.Notify(14488)
					avroArr = memo.([]interface{})
					if len(avroArr) > datumArr.Len() {
						__antithesis_instrumentation__.Notify(14489)
						avroArr = avroArr[:datumArr.Len()]
					} else {
						__antithesis_instrumentation__.Notify(14490)
					}
				} else {
					__antithesis_instrumentation__.Notify(14491)
					avroArr = make([]interface{}, 0, datumArr.Len())
				}
				__antithesis_instrumentation__.Notify(14486)

				for i, elt := range datumArr.Array {
					__antithesis_instrumentation__.Notify(14492)
					var encoded interface{}
					if elt == tree.DNull {
						__antithesis_instrumentation__.Notify(14494)
						encoded = nil
					} else {
						__antithesis_instrumentation__.Notify(14495)
						var encErr error
						if i < len(avroArr) {
							__antithesis_instrumentation__.Notify(14497)
							encoded, encErr = itemSchema.encodeDatum(elt, avroArr[i].(map[string]interface{})[itemUnionKey])
						} else {
							__antithesis_instrumentation__.Notify(14498)
							encoded, encErr = itemSchema.encodeDatum(elt, nil)
						}
						__antithesis_instrumentation__.Notify(14496)
						if encErr != nil {
							__antithesis_instrumentation__.Notify(14499)
							return nil, encErr
						} else {
							__antithesis_instrumentation__.Notify(14500)
						}
					}
					__antithesis_instrumentation__.Notify(14493)

					if i < len(avroArr) {
						__antithesis_instrumentation__.Notify(14501)

						if encoded == nil {
							__antithesis_instrumentation__.Notify(14502)
							avroArr[i] = encoded
						} else {
							__antithesis_instrumentation__.Notify(14503)
							if itemMap, ok := avroArr[i].(map[string]interface{}); ok {
								__antithesis_instrumentation__.Notify(14504)

								itemMap[itemUnionKey] = encoded
							} else {
								__antithesis_instrumentation__.Notify(14505)

								encMap := make(map[string]interface{})
								encMap[itemUnionKey] = encoded
								avroArr[i] = encMap
							}
						}
					} else {
						__antithesis_instrumentation__.Notify(14506)
						if encoded == nil {
							__antithesis_instrumentation__.Notify(14507)
							avroArr = append(avroArr, encoded)
						} else {
							__antithesis_instrumentation__.Notify(14508)
							encMap := make(map[string]interface{})
							encMap[itemUnionKey] = encoded
							avroArr = append(avroArr, encMap)
						}
					}
				}
				__antithesis_instrumentation__.Notify(14487)
				return avroArr, nil
			},
			func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(14509)
				datumArr := tree.NewDArray(itemSchema.typ)
				avroArr := x.([]interface{})
				for _, item := range avroArr {
					__antithesis_instrumentation__.Notify(14511)
					itemDatum, err := itemSchema.decodeFn(item)
					if err != nil {
						__antithesis_instrumentation__.Notify(14513)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(14514)
					}
					__antithesis_instrumentation__.Notify(14512)
					err = datumArr.Append(itemDatum)
					if err != nil {
						__antithesis_instrumentation__.Notify(14515)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(14516)
					}
				}
				__antithesis_instrumentation__.Notify(14510)
				return datumArr, nil
			},
		)

	default:
		__antithesis_instrumentation__.Notify(14400)
		return nil, errors.Errorf(`type %s not yet supported with avro`,
			typ.SQLString())
	}
	__antithesis_instrumentation__.Notify(14346)

	return schema, nil
}

func columnToAvroSchema(col catalog.Column) (*avroSchemaField, error) {
	__antithesis_instrumentation__.Notify(14517)
	schema, err := typeToAvroSchema(col.GetType())
	if err != nil {
		__antithesis_instrumentation__.Notify(14519)
		return nil, errors.Wrapf(err, "column %s", col.GetName())
	} else {
		__antithesis_instrumentation__.Notify(14520)
	}
	__antithesis_instrumentation__.Notify(14518)
	schema.Name = SQLNameToAvroName(col.GetName())
	schema.Metadata = col.ColumnDesc().SQLStringNotHumanReadable()
	schema.Default = nil

	return schema, nil
}

func indexToAvroSchema(
	tableDesc catalog.TableDescriptor, index catalog.Index, sqlName string, namespace string,
) (*avroDataRecord, error) {
	__antithesis_instrumentation__.Notify(14521)
	schema := &avroDataRecord{
		avroRecord: avroRecord{
			Name:       SQLNameToAvroName(sqlName),
			SchemaType: `record`,
			Namespace:  namespace,
		},
		fieldIdxByName:   make(map[string]int),
		colIdxByFieldIdx: make(map[int]int),
		fieldIdxByColIdx: make(map[int]int),
	}
	colIdxByID := catalog.ColumnIDToOrdinalMap(tableDesc.PublicColumns())
	for i := 0; i < index.NumKeyColumns(); i++ {
		__antithesis_instrumentation__.Notify(14525)
		colID := index.GetKeyColumnID(i)
		colIdx, ok := colIdxByID.Get(colID)
		if !ok {
			__antithesis_instrumentation__.Notify(14528)
			return nil, errors.Errorf(`unknown column id: %d`, colID)
		} else {
			__antithesis_instrumentation__.Notify(14529)
		}
		__antithesis_instrumentation__.Notify(14526)
		col := tableDesc.PublicColumns()[colIdx]
		field, err := columnToAvroSchema(col)
		if err != nil {
			__antithesis_instrumentation__.Notify(14530)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14531)
		}
		__antithesis_instrumentation__.Notify(14527)
		schema.colIdxByFieldIdx[len(schema.Fields)] = colIdx
		schema.fieldIdxByColIdx[colIdx] = len(schema.Fields)
		schema.fieldIdxByName[field.Name] = len(schema.Fields)
		schema.Fields = append(schema.Fields, field)
	}
	__antithesis_instrumentation__.Notify(14522)
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(14532)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14533)
	}
	__antithesis_instrumentation__.Notify(14523)
	schema.codec, err = goavro.NewCodec(string(schemaJSON))
	if err != nil {
		__antithesis_instrumentation__.Notify(14534)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14535)
	}
	__antithesis_instrumentation__.Notify(14524)
	return schema, nil
}

const (
	avroSchemaNoSuffix = ``
)

func tableToAvroSchema(
	tableDesc catalog.TableDescriptor,
	familyID descpb.FamilyID,
	nameSuffix string,
	namespace string,
	virtualColumnVisibility string,
) (*avroDataRecord, error) {
	__antithesis_instrumentation__.Notify(14536)
	family, err := tableDesc.FindFamilyByID(familyID)
	if err != nil {
		__antithesis_instrumentation__.Notify(14544)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14545)
	}
	__antithesis_instrumentation__.Notify(14537)
	var name string

	if tableDesc.NumFamilies() > 1 {
		__antithesis_instrumentation__.Notify(14546)
		name = SQLNameToAvroName(tableDesc.GetName() + "." + family.Name)
	} else {
		__antithesis_instrumentation__.Notify(14547)
		name = SQLNameToAvroName(tableDesc.GetName())
	}
	__antithesis_instrumentation__.Notify(14538)
	if nameSuffix != avroSchemaNoSuffix {
		__antithesis_instrumentation__.Notify(14548)
		name = name + `_` + nameSuffix
	} else {
		__antithesis_instrumentation__.Notify(14549)
	}
	__antithesis_instrumentation__.Notify(14539)
	schema := &avroDataRecord{
		avroRecord: avroRecord{
			Name:       name,
			SchemaType: `record`,
			Namespace:  namespace,
		},
		fieldIdxByName:   make(map[string]int),
		colIdxByFieldIdx: make(map[int]int),
		fieldIdxByColIdx: make(map[int]int),
	}

	include := make(map[descpb.ColumnID]struct{}, len(family.ColumnIDs))
	var yes struct{}
	for _, colID := range family.ColumnIDs {
		__antithesis_instrumentation__.Notify(14550)
		include[colID] = yes
	}
	__antithesis_instrumentation__.Notify(14540)

	for _, col := range tableDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(14551)
		_, inFamily := include[col.GetID()]
		virtual := col.IsVirtual() && func() bool {
			__antithesis_instrumentation__.Notify(14552)
			return virtualColumnVisibility == string(changefeedbase.OptVirtualColumnsNull) == true
		}() == true
		if inFamily || func() bool {
			__antithesis_instrumentation__.Notify(14553)
			return virtual == true
		}() == true {
			__antithesis_instrumentation__.Notify(14554)
			field, err := columnToAvroSchema(col)
			if err != nil {
				__antithesis_instrumentation__.Notify(14556)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14557)
			}
			__antithesis_instrumentation__.Notify(14555)
			schema.colIdxByFieldIdx[len(schema.Fields)] = col.Ordinal()
			schema.fieldIdxByName[field.Name] = len(schema.Fields)
			schema.fieldIdxByColIdx[col.Ordinal()] = len(schema.Fields)
			schema.Fields = append(schema.Fields, field)
		} else {
			__antithesis_instrumentation__.Notify(14558)
		}
	}
	__antithesis_instrumentation__.Notify(14541)
	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(14559)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14560)
	}
	__antithesis_instrumentation__.Notify(14542)
	schema.codec, err = goavro.NewCodec(string(schemaJSON))
	if err != nil {
		__antithesis_instrumentation__.Notify(14561)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14562)
	}
	__antithesis_instrumentation__.Notify(14543)
	return schema, nil
}

func (r *avroDataRecord) textualFromRow(row rowenc.EncDatumRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(14563)
	native, err := r.nativeFromRow(row)
	if err != nil {
		__antithesis_instrumentation__.Notify(14565)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14566)
	}
	__antithesis_instrumentation__.Notify(14564)
	return r.codec.TextualFromNative(nil, native)
}

func (r *avroDataRecord) BinaryFromRow(buf []byte, row rowenc.EncDatumRow) ([]byte, error) {
	__antithesis_instrumentation__.Notify(14567)
	native, err := r.nativeFromRow(row)
	if err != nil {
		__antithesis_instrumentation__.Notify(14569)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14570)
	}
	__antithesis_instrumentation__.Notify(14568)
	return r.codec.BinaryFromNative(buf, native)
}

func (r *avroDataRecord) rowFromTextual(buf []byte) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(14571)
	native, newBuf, err := r.codec.NativeFromTextual(buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(14574)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14575)
	}
	__antithesis_instrumentation__.Notify(14572)
	if len(newBuf) > 0 {
		__antithesis_instrumentation__.Notify(14576)
		return nil, errors.New(`only one row was expected`)
	} else {
		__antithesis_instrumentation__.Notify(14577)
	}
	__antithesis_instrumentation__.Notify(14573)
	return r.rowFromNative(native)
}

func (r *avroDataRecord) RowFromBinary(buf []byte) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(14578)
	native, newBuf, err := r.codec.NativeFromBinary(buf)
	if err != nil {
		__antithesis_instrumentation__.Notify(14581)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14582)
	}
	__antithesis_instrumentation__.Notify(14579)
	if len(newBuf) > 0 {
		__antithesis_instrumentation__.Notify(14583)
		return nil, errors.New(`only one row was expected`)
	} else {
		__antithesis_instrumentation__.Notify(14584)
	}
	__antithesis_instrumentation__.Notify(14580)
	return r.rowFromNative(native)
}

func (r *avroDataRecord) nativeFromRow(row rowenc.EncDatumRow) (interface{}, error) {
	__antithesis_instrumentation__.Notify(14585)
	if r.native == nil {
		__antithesis_instrumentation__.Notify(14588)

		r.native = make(map[string]interface{}, len(r.Fields))
	} else {
		__antithesis_instrumentation__.Notify(14589)
	}
	__antithesis_instrumentation__.Notify(14586)

	for fieldIdx, field := range r.Fields {
		__antithesis_instrumentation__.Notify(14590)
		d := row[r.colIdxByFieldIdx[fieldIdx]]
		if err := d.EnsureDecoded(field.typ, &r.alloc); err != nil {
			__antithesis_instrumentation__.Notify(14592)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14593)
		}
		__antithesis_instrumentation__.Notify(14591)
		var err error
		if r.native[field.Name], err = field.encodeFn(d.Datum); err != nil {
			__antithesis_instrumentation__.Notify(14594)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14595)
		}
	}
	__antithesis_instrumentation__.Notify(14587)
	return r.native, nil
}

func (r *avroDataRecord) rowFromNative(native interface{}) (rowenc.EncDatumRow, error) {
	__antithesis_instrumentation__.Notify(14596)
	avroDatums, ok := native.(map[string]interface{})
	if !ok {
		__antithesis_instrumentation__.Notify(14600)
		return nil, errors.Errorf(`unknown avro native type: %T`, native)
	} else {
		__antithesis_instrumentation__.Notify(14601)
	}
	__antithesis_instrumentation__.Notify(14597)
	if len(r.Fields) != len(avroDatums) {
		__antithesis_instrumentation__.Notify(14602)
		return nil, errors.Errorf(
			`expected row with %d columns got %d`, len(r.Fields), len(avroDatums))
	} else {
		__antithesis_instrumentation__.Notify(14603)
	}
	__antithesis_instrumentation__.Notify(14598)
	row := make(rowenc.EncDatumRow, len(r.Fields))
	for fieldName, avroDatum := range avroDatums {
		__antithesis_instrumentation__.Notify(14604)
		fieldIdx := r.fieldIdxByName[fieldName]
		field := r.Fields[fieldIdx]
		decoded, err := field.decodeFn(avroDatum)
		if err != nil {
			__antithesis_instrumentation__.Notify(14606)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(14607)
		}
		__antithesis_instrumentation__.Notify(14605)
		row[r.colIdxByFieldIdx[fieldIdx]] = rowenc.DatumToEncDatum(field.typ, decoded)
	}
	__antithesis_instrumentation__.Notify(14599)
	return row, nil
}

func envelopeToAvroSchema(
	topic string, opts avroEnvelopeOpts, before, after *avroDataRecord, namespace string,
) (*avroEnvelopeRecord, error) {
	__antithesis_instrumentation__.Notify(14608)
	schema := &avroEnvelopeRecord{
		avroRecord: avroRecord{
			Name:       SQLNameToAvroName(topic) + `_envelope`,
			SchemaType: `record`,
			Namespace:  namespace,
		},
		opts: opts,
	}

	if opts.beforeField {
		__antithesis_instrumentation__.Notify(14615)
		schema.before = before
		beforeField := &avroSchemaField{
			Name:       `before`,
			SchemaType: []avroSchemaType{avroSchemaNull, before},
			Default:    nil,
		}
		schema.Fields = append(schema.Fields, beforeField)
	} else {
		__antithesis_instrumentation__.Notify(14616)
	}
	__antithesis_instrumentation__.Notify(14609)
	if opts.afterField {
		__antithesis_instrumentation__.Notify(14617)
		schema.after = after
		afterField := &avroSchemaField{
			Name:       `after`,
			SchemaType: []avroSchemaType{avroSchemaNull, after},
			Default:    nil,
		}
		schema.Fields = append(schema.Fields, afterField)
	} else {
		__antithesis_instrumentation__.Notify(14618)
	}
	__antithesis_instrumentation__.Notify(14610)
	if opts.updatedField {
		__antithesis_instrumentation__.Notify(14619)
		updatedField := &avroSchemaField{
			SchemaType: []avroSchemaType{avroSchemaNull, avroSchemaString},
			Name:       `updated`,
			Default:    nil,
		}
		schema.Fields = append(schema.Fields, updatedField)
	} else {
		__antithesis_instrumentation__.Notify(14620)
	}
	__antithesis_instrumentation__.Notify(14611)
	if opts.resolvedField {
		__antithesis_instrumentation__.Notify(14621)
		resolvedField := &avroSchemaField{
			SchemaType: []avroSchemaType{avroSchemaNull, avroSchemaString},
			Name:       `resolved`,
			Default:    nil,
		}
		schema.Fields = append(schema.Fields, resolvedField)
	} else {
		__antithesis_instrumentation__.Notify(14622)
	}
	__antithesis_instrumentation__.Notify(14612)

	schemaJSON, err := json.Marshal(schema)
	if err != nil {
		__antithesis_instrumentation__.Notify(14623)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14624)
	}
	__antithesis_instrumentation__.Notify(14613)
	schema.codec, err = goavro.NewCodec(string(schemaJSON))
	if err != nil {
		__antithesis_instrumentation__.Notify(14625)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14626)
	}
	__antithesis_instrumentation__.Notify(14614)
	return schema, nil
}

func (r *avroEnvelopeRecord) BinaryFromRow(
	buf []byte, meta avroMetadata, beforeRow, afterRow rowenc.EncDatumRow,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(14627)
	native := map[string]interface{}{}
	if r.opts.beforeField {
		__antithesis_instrumentation__.Notify(14633)
		if beforeRow == nil {
			__antithesis_instrumentation__.Notify(14634)
			native[`before`] = nil
		} else {
			__antithesis_instrumentation__.Notify(14635)
			beforeNative, err := r.before.nativeFromRow(beforeRow)
			if err != nil {
				__antithesis_instrumentation__.Notify(14637)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14638)
			}
			__antithesis_instrumentation__.Notify(14636)
			native[`before`] = goavro.Union(avroUnionKey(&r.before.avroRecord), beforeNative)
		}
	} else {
		__antithesis_instrumentation__.Notify(14639)
	}
	__antithesis_instrumentation__.Notify(14628)
	if r.opts.afterField {
		__antithesis_instrumentation__.Notify(14640)
		if afterRow == nil {
			__antithesis_instrumentation__.Notify(14641)
			native[`after`] = nil
		} else {
			__antithesis_instrumentation__.Notify(14642)
			afterNative, err := r.after.nativeFromRow(afterRow)
			if err != nil {
				__antithesis_instrumentation__.Notify(14644)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(14645)
			}
			__antithesis_instrumentation__.Notify(14643)
			native[`after`] = goavro.Union(avroUnionKey(&r.after.avroRecord), afterNative)
		}
	} else {
		__antithesis_instrumentation__.Notify(14646)
	}
	__antithesis_instrumentation__.Notify(14629)
	if r.opts.updatedField {
		__antithesis_instrumentation__.Notify(14647)
		native[`updated`] = nil
		if u, ok := meta[`updated`]; ok {
			__antithesis_instrumentation__.Notify(14648)
			delete(meta, `updated`)
			ts, ok := u.(hlc.Timestamp)
			if !ok {
				__antithesis_instrumentation__.Notify(14650)
				return nil, errors.Errorf(`unknown metadata timestamp type: %T`, u)
			} else {
				__antithesis_instrumentation__.Notify(14651)
			}
			__antithesis_instrumentation__.Notify(14649)
			native[`updated`] = goavro.Union(avroUnionKey(avroSchemaString), ts.AsOfSystemTime())
		} else {
			__antithesis_instrumentation__.Notify(14652)
		}
	} else {
		__antithesis_instrumentation__.Notify(14653)
	}
	__antithesis_instrumentation__.Notify(14630)
	if r.opts.resolvedField {
		__antithesis_instrumentation__.Notify(14654)
		native[`resolved`] = nil
		if u, ok := meta[`resolved`]; ok {
			__antithesis_instrumentation__.Notify(14655)
			delete(meta, `resolved`)
			ts, ok := u.(hlc.Timestamp)
			if !ok {
				__antithesis_instrumentation__.Notify(14657)
				return nil, errors.Errorf(`unknown metadata timestamp type: %T`, u)
			} else {
				__antithesis_instrumentation__.Notify(14658)
			}
			__antithesis_instrumentation__.Notify(14656)
			native[`resolved`] = goavro.Union(avroUnionKey(avroSchemaString), ts.AsOfSystemTime())
		} else {
			__antithesis_instrumentation__.Notify(14659)
		}
	} else {
		__antithesis_instrumentation__.Notify(14660)
	}
	__antithesis_instrumentation__.Notify(14631)
	for k := range meta {
		__antithesis_instrumentation__.Notify(14661)
		return nil, errors.AssertionFailedf(`unhandled meta key: %s`, k)
	}
	__antithesis_instrumentation__.Notify(14632)
	return r.codec.BinaryFromNative(buf, native)
}

func (r *avroDataRecord) refreshTypeMetadata(tbl catalog.TableDescriptor) {
	__antithesis_instrumentation__.Notify(14662)
	for _, col := range tbl.UserDefinedTypeColumns() {
		__antithesis_instrumentation__.Notify(14663)
		if fieldIdx, ok := r.fieldIdxByColIdx[col.Ordinal()]; ok {
			__antithesis_instrumentation__.Notify(14664)
			r.Fields[fieldIdx].typ = col.GetType()
		} else {
			__antithesis_instrumentation__.Notify(14665)
		}
	}
}

func decimalToRat(dec apd.Decimal, scale int32) (big.Rat, error) {
	__antithesis_instrumentation__.Notify(14666)
	if dec.Form != apd.Finite {
		__antithesis_instrumentation__.Notify(14671)
		return big.Rat{}, errors.Errorf(`cannot convert %s form decimal`, dec.Form)
	} else {
		__antithesis_instrumentation__.Notify(14672)
	}
	__antithesis_instrumentation__.Notify(14667)
	if scale > 0 && func() bool {
		__antithesis_instrumentation__.Notify(14673)
		return scale != -dec.Exponent == true
	}() == true {
		__antithesis_instrumentation__.Notify(14674)
		return big.Rat{}, errors.Errorf(`%s will not roundtrip at scale %d`, &dec, scale)
	} else {
		__antithesis_instrumentation__.Notify(14675)
	}
	__antithesis_instrumentation__.Notify(14668)
	var r big.Rat
	if dec.Exponent >= 0 {
		__antithesis_instrumentation__.Notify(14676)
		exp := big.NewInt(10)
		exp = exp.Exp(exp, big.NewInt(int64(dec.Exponent)), nil)
		coeff := dec.Coeff.MathBigInt()
		r.SetFrac(coeff.Mul(coeff, exp), big.NewInt(1))
	} else {
		__antithesis_instrumentation__.Notify(14677)
		exp := big.NewInt(10)
		exp = exp.Exp(exp, big.NewInt(int64(-dec.Exponent)), nil)
		coeff := dec.Coeff.MathBigInt()
		r.SetFrac(coeff, exp)
	}
	__antithesis_instrumentation__.Notify(14669)
	if dec.Negative {
		__antithesis_instrumentation__.Notify(14678)
		r.Mul(&r, big.NewRat(-1, 1))
	} else {
		__antithesis_instrumentation__.Notify(14679)
	}
	__antithesis_instrumentation__.Notify(14670)
	return r, nil
}

func ratToDecimal(rat big.Rat, scale int32) apd.Decimal {
	__antithesis_instrumentation__.Notify(14680)
	num, denom := rat.Num(), rat.Denom()
	exp := big.NewInt(10)
	exp = exp.Exp(exp, big.NewInt(int64(scale)), nil)
	sf := denom.Div(exp, denom)
	var coeff apd.BigInt
	coeff.SetMathBigInt(num.Mul(num, sf))
	dec := apd.NewWithBigInt(&coeff, -scale)
	return *dec
}
