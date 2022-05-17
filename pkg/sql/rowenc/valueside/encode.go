package valueside

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/errors"
)

func Encode(appendTo []byte, colID ColumnIDDelta, val tree.Datum, scratch []byte) ([]byte, error) {
	__antithesis_instrumentation__.Notify(571370)
	if val == tree.DNull {
		__antithesis_instrumentation__.Notify(571372)
		return encoding.EncodeNullValue(appendTo, uint32(colID)), nil
	} else {
		__antithesis_instrumentation__.Notify(571373)
	}
	__antithesis_instrumentation__.Notify(571371)
	switch t := tree.UnwrapDatum(nil, val).(type) {
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(571374)
		return encoding.EncodeBitArrayValue(appendTo, uint32(colID), t.BitArray), nil
	case *tree.DBool:
		__antithesis_instrumentation__.Notify(571375)
		return encoding.EncodeBoolValue(appendTo, uint32(colID), bool(*t)), nil
	case *tree.DInt:
		__antithesis_instrumentation__.Notify(571376)
		return encoding.EncodeIntValue(appendTo, uint32(colID), int64(*t)), nil
	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(571377)
		return encoding.EncodeFloatValue(appendTo, uint32(colID), float64(*t)), nil
	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(571378)
		return encoding.EncodeDecimalValue(appendTo, uint32(colID), &t.Decimal), nil
	case *tree.DString:
		__antithesis_instrumentation__.Notify(571379)
		return encoding.EncodeBytesValue(appendTo, uint32(colID), []byte(*t)), nil
	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(571380)
		return encoding.EncodeBytesValue(appendTo, uint32(colID), []byte(*t)), nil
	case *tree.DDate:
		__antithesis_instrumentation__.Notify(571381)
		return encoding.EncodeIntValue(appendTo, uint32(colID), t.UnixEpochDaysWithOrig()), nil
	case *tree.DBox2D:
		__antithesis_instrumentation__.Notify(571382)
		return encoding.EncodeBox2DValue(appendTo, uint32(colID), t.CartesianBoundingBox.BoundingBox)
	case *tree.DGeography:
		__antithesis_instrumentation__.Notify(571383)
		return encoding.EncodeGeoValue(appendTo, uint32(colID), t.SpatialObjectRef())
	case *tree.DGeometry:
		__antithesis_instrumentation__.Notify(571384)
		return encoding.EncodeGeoValue(appendTo, uint32(colID), t.SpatialObjectRef())
	case *tree.DTime:
		__antithesis_instrumentation__.Notify(571385)
		return encoding.EncodeIntValue(appendTo, uint32(colID), int64(*t)), nil
	case *tree.DTimeTZ:
		__antithesis_instrumentation__.Notify(571386)
		return encoding.EncodeTimeTZValue(appendTo, uint32(colID), t.TimeTZ), nil
	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(571387)
		return encoding.EncodeTimeValue(appendTo, uint32(colID), t.Time), nil
	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(571388)
		return encoding.EncodeTimeValue(appendTo, uint32(colID), t.Time), nil
	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(571389)
		return encoding.EncodeDurationValue(appendTo, uint32(colID), t.Duration), nil
	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(571390)
		return encoding.EncodeUUIDValue(appendTo, uint32(colID), t.UUID), nil
	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(571391)
		return encoding.EncodeIPAddrValue(appendTo, uint32(colID), t.IPAddr), nil
	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(571392)
		encoded, err := json.EncodeJSON(scratch, t.JSON)
		if err != nil {
			__antithesis_instrumentation__.Notify(571402)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571403)
		}
		__antithesis_instrumentation__.Notify(571393)
		return encoding.EncodeJSONValue(appendTo, uint32(colID), encoded), nil
	case *tree.DArray:
		__antithesis_instrumentation__.Notify(571394)
		a, err := encodeArray(t, scratch)
		if err != nil {
			__antithesis_instrumentation__.Notify(571404)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(571405)
		}
		__antithesis_instrumentation__.Notify(571395)
		return encoding.EncodeArrayValue(appendTo, uint32(colID), a), nil
	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(571396)
		return encodeTuple(t, appendTo, uint32(colID), scratch)
	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(571397)
		return encoding.EncodeBytesValue(appendTo, uint32(colID), []byte(t.Contents)), nil
	case *tree.DOid:
		__antithesis_instrumentation__.Notify(571398)
		return encoding.EncodeIntValue(appendTo, uint32(colID), int64(t.DInt)), nil
	case *tree.DEnum:
		__antithesis_instrumentation__.Notify(571399)
		return encoding.EncodeBytesValue(appendTo, uint32(colID), t.PhysicalRep), nil
	case *tree.DVoid:
		__antithesis_instrumentation__.Notify(571400)
		return encoding.EncodeVoidValue(appendTo, uint32(colID)), nil
	default:
		__antithesis_instrumentation__.Notify(571401)
		return nil, errors.Errorf("unable to encode table value: %T", t)
	}
}

type ColumnIDDelta uint32

const NoColumnID = ColumnIDDelta(encoding.NoColumnID)

func MakeColumnIDDelta(previous, current descpb.ColumnID) ColumnIDDelta {
	__antithesis_instrumentation__.Notify(571406)
	if previous > current {
		__antithesis_instrumentation__.Notify(571408)
		panic(errors.AssertionFailedf("cannot write column id %d after %d", current, previous))
	} else {
		__antithesis_instrumentation__.Notify(571409)
	}
	__antithesis_instrumentation__.Notify(571407)
	return ColumnIDDelta(current - previous)
}
