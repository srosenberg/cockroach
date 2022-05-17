package valueside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

func encodeArray(d *tree.DArray, scratch []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571139)
	if err := d.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(571145)
		return scratch, err
	} else {
		__antithesis_instrumentation__.Notify(571146)
	}
	__antithesis_instrumentation__.Notify(571140)
	scratch = scratch[0:0]
	elementType, err := DatumTypeToArrayElementEncodingType(d.ParamTyp)

	if err != nil {
		__antithesis_instrumentation__.Notify(571147)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571148)
	}
	__antithesis_instrumentation__.Notify(571141)
	header := arrayHeader{
		hasNulls: d.HasNulls,

		numDimensions: 1,
		elementType:   elementType,
		length:        uint64(d.Len()),
	}
	scratch, err = encodeArrayHeader(header, scratch)
	if err != nil {
		__antithesis_instrumentation__.Notify(571149)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(571150)
	}
	__antithesis_instrumentation__.Notify(571142)
	nullBitmapStart := len(scratch)
	if d.HasNulls {
		__antithesis_instrumentation__.Notify(571151)
		for i := 0; i < numBytesInBitArray(d.Len()); i++ {
			__antithesis_instrumentation__.Notify(571152)
			scratch = append(scratch, 0)
		}
	} else {
		__antithesis_instrumentation__.Notify(571153)
	}
	__antithesis_instrumentation__.Notify(571143)
	for i, e := range d.Array {
		__antithesis_instrumentation__.Notify(571154)
		var err error
		if d.HasNulls && func() bool {
			__antithesis_instrumentation__.Notify(571155)
			return e == tree.DNull == true
		}() == true {
			__antithesis_instrumentation__.Notify(571156)
			setBit(scratch[nullBitmapStart:], i)
		} else {
			__antithesis_instrumentation__.Notify(571157)
			scratch, err = encodeArrayElement(scratch, e)
			if err != nil {
				__antithesis_instrumentation__.Notify(571158)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(571159)
			}
		}
	}
	__antithesis_instrumentation__.Notify(571144)
	return scratch, nil
}

func decodeArray(a *tree.DatumAlloc, arrayType *types.T, b []byte) (tree.Datum, []byte, error) {
	__antithesis_instrumentation__.Notify(571160)
	header, b, err := decodeArrayHeader(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(571164)
		return nil, b, err
	} else {
		__antithesis_instrumentation__.Notify(571165)
	}
	__antithesis_instrumentation__.Notify(571161)
	elementType := arrayType.ArrayContents()
	result := tree.DArray{
		Array:    make(tree.Datums, header.length),
		ParamTyp: elementType,
	}
	if err = result.MaybeSetCustomOid(arrayType); err != nil {
		__antithesis_instrumentation__.Notify(571166)
		return nil, b, err
	} else {
		__antithesis_instrumentation__.Notify(571167)
	}
	__antithesis_instrumentation__.Notify(571162)
	var val tree.Datum
	for i := uint64(0); i < header.length; i++ {
		__antithesis_instrumentation__.Notify(571168)
		if header.isNull(i) {
			__antithesis_instrumentation__.Notify(571169)
			result.Array[i] = tree.DNull
			result.HasNulls = true
		} else {
			__antithesis_instrumentation__.Notify(571170)
			result.HasNonNulls = true
			val, b, err = DecodeUntaggedDatum(a, elementType, b)
			if err != nil {
				__antithesis_instrumentation__.Notify(571172)
				return nil, b, err
			} else {
				__antithesis_instrumentation__.Notify(571173)
			}
			__antithesis_instrumentation__.Notify(571171)
			result.Array[i] = val
		}
	}
	__antithesis_instrumentation__.Notify(571163)
	return &result, b, nil
}

type arrayHeader struct {
	hasNulls bool

	numDimensions int

	elementType encoding.Type

	length uint64

	nullBitmap []byte
}

func (h arrayHeader) isNull(i uint64) bool {
	__antithesis_instrumentation__.Notify(571174)
	return h.hasNulls && func() bool {
		__antithesis_instrumentation__.Notify(571175)
		return ((h.nullBitmap[i/8] >> (i % 8)) & 1) == 1 == true
	}() == true
}

func setBit(bitmap []byte, idx int) {
	__antithesis_instrumentation__.Notify(571176)
	bitmap[idx/8] = bitmap[idx/8] | (1 << uint(idx%8))
}

func numBytesInBitArray(numBits int) int {
	__antithesis_instrumentation__.Notify(571177)
	return (numBits + 7) / 8
}

func makeBitVec(src []byte, length int) (b, bitVec []byte) {
	__antithesis_instrumentation__.Notify(571178)
	nullBitmapNumBytes := numBytesInBitArray(length)
	return src[nullBitmapNumBytes:], src[:nullBitmapNumBytes]
}

const hasNullFlag = 1 << 4

func encodeArrayHeader(h arrayHeader, buf []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571179)

	headerByte := h.numDimensions
	if h.hasNulls {
		__antithesis_instrumentation__.Notify(571181)
		headerByte = headerByte | hasNullFlag
	} else {
		__antithesis_instrumentation__.Notify(571182)
	}
	__antithesis_instrumentation__.Notify(571180)
	buf = append(buf, byte(headerByte))
	buf = encoding.EncodeValueTag(buf, encoding.NoColumnID, h.elementType)
	buf = encoding.EncodeNonsortingUvarint(buf, h.length)
	return buf, nil
}

func decodeArrayHeader(b []byte) (arrayHeader, []byte, error) {
	__antithesis_instrumentation__.Notify(571183)
	if len(b) < 2 {
		__antithesis_instrumentation__.Notify(571188)
		return arrayHeader{}, b, errors.Errorf("buffer too small")
	} else {
		__antithesis_instrumentation__.Notify(571189)
	}
	__antithesis_instrumentation__.Notify(571184)
	hasNulls := b[0]&hasNullFlag != 0
	b = b[1:]
	_, dataOffset, _, encType, err := encoding.DecodeValueTag(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(571190)
		return arrayHeader{}, b, err
	} else {
		__antithesis_instrumentation__.Notify(571191)
	}
	__antithesis_instrumentation__.Notify(571185)
	b = b[dataOffset:]
	b, _, length, err := encoding.DecodeNonsortingUvarint(b)
	if err != nil {
		__antithesis_instrumentation__.Notify(571192)
		return arrayHeader{}, b, err
	} else {
		__antithesis_instrumentation__.Notify(571193)
	}
	__antithesis_instrumentation__.Notify(571186)
	nullBitmap := []byte(nil)
	if hasNulls {
		__antithesis_instrumentation__.Notify(571194)
		b, nullBitmap = makeBitVec(b, int(length))
	} else {
		__antithesis_instrumentation__.Notify(571195)
	}
	__antithesis_instrumentation__.Notify(571187)
	return arrayHeader{
		hasNulls: hasNulls,

		numDimensions: 1,
		elementType:   encType,
		length:        length,
		nullBitmap:    nullBitmap,
	}, b, nil
}

func DatumTypeToArrayElementEncodingType(t *types.T) (encoding.Type, error) {
	__antithesis_instrumentation__.Notify(571196)
	switch t.Family() {
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(571197)
		return encoding.Int, nil
	case types.OidFamily:
		__antithesis_instrumentation__.Notify(571198)
		return encoding.Int, nil
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(571199)
		return encoding.Float, nil
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(571200)
		return encoding.Box2D, nil
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(571201)
		return encoding.Geo, nil
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(571202)
		return encoding.Geo, nil
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(571203)
		return encoding.Decimal, nil
	case types.BytesFamily, types.StringFamily, types.CollatedStringFamily, types.EnumFamily:
		__antithesis_instrumentation__.Notify(571204)
		return encoding.Bytes, nil
	case types.TimestampFamily, types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(571205)
		return encoding.Time, nil

	case types.DateFamily, types.TimeFamily:
		__antithesis_instrumentation__.Notify(571206)
		return encoding.Int, nil
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(571207)
		return encoding.TimeTZ, nil
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(571208)
		return encoding.Duration, nil
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(571209)
		return encoding.True, nil
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(571210)
		return encoding.BitArray, nil
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(571211)
		return encoding.UUID, nil
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(571212)
		return encoding.IPAddr, nil
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(571213)
		return encoding.JSON, nil
	case types.TupleFamily:
		__antithesis_instrumentation__.Notify(571214)
		return encoding.Tuple, nil
	default:
		__antithesis_instrumentation__.Notify(571215)
		return 0, errors.AssertionFailedf("no known encoding type for %s", t)
	}
}
func checkElementType(paramType *types.T, elemType *types.T) error {
	__antithesis_instrumentation__.Notify(571216)
	if paramType.Family() != elemType.Family() {
		__antithesis_instrumentation__.Notify(571219)
		return errors.Errorf("type of array contents %s doesn't match column type %s",
			paramType, elemType.Family())
	} else {
		__antithesis_instrumentation__.Notify(571220)
	}
	__antithesis_instrumentation__.Notify(571217)
	if paramType.Family() == types.CollatedStringFamily {
		__antithesis_instrumentation__.Notify(571221)
		if paramType.Locale() != elemType.Locale() {
			__antithesis_instrumentation__.Notify(571222)
			return errors.Errorf("locale of collated string array being inserted (%s) doesn't match locale of column type (%s)",
				paramType.Locale(), elemType.Locale())
		} else {
			__antithesis_instrumentation__.Notify(571223)
		}
	} else {
		__antithesis_instrumentation__.Notify(571224)
	}
	__antithesis_instrumentation__.Notify(571218)
	return nil
}

func encodeArrayElement(b []byte, d tree.Datum) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571225)
	switch t := tree.UnwrapDatum(nil, d).(type) {
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(571226)
		return encoding.EncodeUntaggedIntValue(b, int64(*t)), nil
	case *tree.DString:
		__antithesis_instrumentation__.Notify(571227)
		bytes := []byte(*t)
		b = encoding.EncodeUntaggedBytesValue(b, bytes)
		return b, nil
	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(571228)
		bytes := []byte(*t)
		b = encoding.EncodeUntaggedBytesValue(b, bytes)
		return b, nil
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(571229)
		return encoding.EncodeUntaggedBitArrayValue(b, t.BitArray), nil
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(571230)
		return encoding.EncodeUntaggedFloatValue(b, float64(*t)), nil
	case *tree.DBool:
		__antithesis_instrumentation__.Notify(571231)
		return encoding.EncodeBoolValue(b, encoding.NoColumnID, bool(*t)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(571232)
		return encoding.EncodeUntaggedDecimalValue(b, &t.Decimal), nil
	case *tree.DDate:
		__antithesis_instrumentation__.Notify(571233)
		return encoding.EncodeUntaggedIntValue(b, t.UnixEpochDaysWithOrig()), nil
	case *tree.DBox2D:
		__antithesis_instrumentation__.Notify(571234)
		return encoding.EncodeUntaggedBox2DValue(b, t.CartesianBoundingBox.BoundingBox)
	case *tree.DGeography:
		__antithesis_instrumentation__.Notify(571235)
		return encoding.EncodeUntaggedGeoValue(b, t.SpatialObjectRef())
	case *tree.DGeometry:
		__antithesis_instrumentation__.Notify(571236)
		return encoding.EncodeUntaggedGeoValue(b, t.SpatialObjectRef())
	case *tree.DTime:
		__antithesis_instrumentation__.Notify(571237)
		return encoding.EncodeUntaggedIntValue(b, int64(*t)), nil
	case *tree.DTimeTZ:
		__antithesis_instrumentation__.Notify(571238)
		return encoding.EncodeUntaggedTimeTZValue(b, t.TimeTZ), nil
	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(571239)
		return encoding.EncodeUntaggedTimeValue(b, t.Time), nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(571240)
		return encoding.EncodeUntaggedTimeValue(b, t.Time), nil
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(571241)
		return encoding.EncodeUntaggedDurationValue(b, t.Duration), nil
	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(571242)
		return encoding.EncodeUntaggedUUIDValue(b, t.UUID), nil
	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(571243)
		return encoding.EncodeUntaggedIPAddrValue(b, t.IPAddr), nil
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(571244)
		return encoding.EncodeUntaggedIntValue(b, int64(t.DInt)), nil
	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(571245)
		return encoding.EncodeUntaggedBytesValue(b, []byte(t.Contents)), nil
	case *tree.DOidWrapper:
		__antithesis_instrumentation__.Notify(571246)
		return encodeArrayElement(b, t.Wrapped)
	case *tree.DEnum:
		__antithesis_instrumentation__.Notify(571247)
		return encoding.EncodeUntaggedBytesValue(b, t.PhysicalRep), nil
	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(571248)
		encoded, err := json.EncodeJSON(nil, t.JSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(571252)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571253)
		}
		__antithesis_instrumentation__.Notify(571249)
		return encoding.EncodeUntaggedBytesValue(b, encoded), nil
	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(571250)
		return encodeUntaggedTuple(t, b, encoding.NoColumnID, nil)
	default:
		__antithesis_instrumentation__.Notify(571251)
		return nil, errors.Errorf("don't know how to encode %s (%T)", d, d)
	}
}
