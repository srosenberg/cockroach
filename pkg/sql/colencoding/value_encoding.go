package colencoding

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
)

func DecodeTableValueToCol(
	da *tree.DatumAlloc,
	vecs *coldata.TypedVecs,
	vecIdx int,
	rowIdx int,
	typ encoding.Type,
	dataOffset int,
	valTyp *types.T,
	buf []byte,
) ([]byte, error) {
	__antithesis_instrumentation__.Notify(274245)

	if typ == encoding.Null {
		__antithesis_instrumentation__.Notify(274248)
		vecs.Nulls[vecIdx].SetNull(rowIdx)
		return buf[dataOffset:], nil
	} else {
		__antithesis_instrumentation__.Notify(274249)
	}
	__antithesis_instrumentation__.Notify(274246)

	origBuf := buf
	buf = buf[dataOffset:]

	colIdx := vecs.ColsMap[vecIdx]

	var err error
	switch valTyp.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(274250)
		var b bool

		buf, b, err = encoding.DecodeBoolValue(origBuf)
		vecs.BoolCols[colIdx][rowIdx] = b
	case types.BytesFamily, types.StringFamily:
		__antithesis_instrumentation__.Notify(274251)
		var data []byte
		buf, data, err = encoding.DecodeUntaggedBytesValue(buf)
		vecs.BytesCols[colIdx].Set(rowIdx, data)
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(274252)
		var i int64
		buf, i, err = encoding.DecodeUntaggedIntValue(buf)
		vecs.Int64Cols[colIdx][rowIdx] = i
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(274253)
		buf, err = encoding.DecodeIntoUntaggedDecimalValue(&vecs.DecimalCols[colIdx][rowIdx], buf)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(274254)
		var f float64
		buf, f, err = encoding.DecodeUntaggedFloatValue(buf)
		vecs.Float64Cols[colIdx][rowIdx] = f
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(274255)
		var i int64
		buf, i, err = encoding.DecodeUntaggedIntValue(buf)
		switch valTyp.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(274263)
			vecs.Int16Cols[colIdx][rowIdx] = int16(i)
		case 32:
			__antithesis_instrumentation__.Notify(274264)
			vecs.Int32Cols[colIdx][rowIdx] = int32(i)
		default:
			__antithesis_instrumentation__.Notify(274265)

			vecs.Int64Cols[colIdx][rowIdx] = i
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(274256)
		var data []byte
		buf, data, err = encoding.DecodeUntaggedBytesValue(buf)
		vecs.JSONCols[colIdx].Bytes.Set(rowIdx, data)
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(274257)
		var data uuid.UUID
		buf, data, err = encoding.DecodeUntaggedUUIDValue(buf)

		if err != nil {
			__antithesis_instrumentation__.Notify(274266)
			return buf, err
		} else {
			__antithesis_instrumentation__.Notify(274267)
		}
		__antithesis_instrumentation__.Notify(274258)
		vecs.BytesCols[colIdx].Set(rowIdx, data.GetBytes())
	case types.TimestampFamily, types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(274259)
		var t time.Time
		buf, t, err = encoding.DecodeUntaggedTimeValue(buf)
		vecs.TimestampCols[colIdx][rowIdx] = t
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(274260)
		var d duration.Duration
		buf, d, err = encoding.DecodeUntaggedDurationValue(buf)
		vecs.IntervalCols[colIdx][rowIdx] = d

	default:
		__antithesis_instrumentation__.Notify(274261)
		var d tree.Datum
		d, buf, err = valueside.DecodeUntaggedDatum(da, valTyp, buf)
		if err != nil {
			__antithesis_instrumentation__.Notify(274268)
			return buf, err
		} else {
			__antithesis_instrumentation__.Notify(274269)
		}
		__antithesis_instrumentation__.Notify(274262)
		vecs.DatumCols[colIdx].Set(rowIdx, d)
	}
	__antithesis_instrumentation__.Notify(274247)
	return buf, err
}
