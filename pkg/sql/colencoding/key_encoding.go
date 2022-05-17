package colencoding

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/valueside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/redact"
)

func DecodeKeyValsToCols(
	da *tree.DatumAlloc,
	vecs *coldata.TypedVecs,
	rowIdx int,
	indexColIdx []int,
	checkAllColsForNull bool,
	keyCols []descpb.IndexFetchSpec_KeyColumn,
	unseen *util.FastIntSet,
	key []byte,
	scratch []byte,
) (remainingKey []byte, foundNull bool, retScratch []byte, _ error) {
	__antithesis_instrumentation__.Notify(274156)
	for j := range keyCols {
		__antithesis_instrumentation__.Notify(274158)
		var err error
		vecIdx := indexColIdx[j]
		if vecIdx == -1 {
			__antithesis_instrumentation__.Notify(274160)
			if checkAllColsForNull {
				__antithesis_instrumentation__.Notify(274162)
				isNull := encoding.PeekType(key) == encoding.Null
				foundNull = foundNull || func() bool {
					__antithesis_instrumentation__.Notify(274163)
					return isNull == true
				}() == true
			} else {
				__antithesis_instrumentation__.Notify(274164)
			}
			__antithesis_instrumentation__.Notify(274161)

			key, err = keyside.Skip(key)
		} else {
			__antithesis_instrumentation__.Notify(274165)
			if unseen != nil {
				__antithesis_instrumentation__.Notify(274167)
				unseen.Remove(vecIdx)
			} else {
				__antithesis_instrumentation__.Notify(274168)
			}
			__antithesis_instrumentation__.Notify(274166)
			var isNull bool
			key, isNull, scratch, err = decodeTableKeyToCol(
				da, vecs, vecIdx, rowIdx,
				keyCols[j].Type, key, keyCols[j].Direction, keyCols[j].IsInverted,
				scratch,
			)
			foundNull = isNull || func() bool {
				__antithesis_instrumentation__.Notify(274169)
				return foundNull == true
			}() == true
		}
		__antithesis_instrumentation__.Notify(274159)
		if err != nil {
			__antithesis_instrumentation__.Notify(274170)
			return nil, false, scratch, err
		} else {
			__antithesis_instrumentation__.Notify(274171)
		}
	}
	__antithesis_instrumentation__.Notify(274157)
	return key, foundNull, scratch, nil
}

func decodeTableKeyToCol(
	da *tree.DatumAlloc,
	vecs *coldata.TypedVecs,
	vecIdx int,
	rowIdx int,
	valType *types.T,
	key []byte,
	dir descpb.IndexDescriptor_Direction,
	isInverted bool,
	scratch []byte,
) (_ []byte, _ bool, retScratch []byte, _ error) {
	__antithesis_instrumentation__.Notify(274172)
	if (dir != descpb.IndexDescriptor_ASC) && func() bool {
		__antithesis_instrumentation__.Notify(274177)
		return (dir != descpb.IndexDescriptor_DESC) == true
	}() == true {
		__antithesis_instrumentation__.Notify(274178)
		return nil, false, scratch, errors.AssertionFailedf("invalid direction: %d", redact.Safe(dir))
	} else {
		__antithesis_instrumentation__.Notify(274179)
	}
	__antithesis_instrumentation__.Notify(274173)
	var isNull bool
	if key, isNull = encoding.DecodeIfNull(key); isNull {
		__antithesis_instrumentation__.Notify(274180)
		vecs.Nulls[vecIdx].SetNull(rowIdx)
		return key, true, scratch, nil
	} else {
		__antithesis_instrumentation__.Notify(274181)
	}
	__antithesis_instrumentation__.Notify(274174)

	colIdx := vecs.ColsMap[vecIdx]

	if isInverted {
		__antithesis_instrumentation__.Notify(274182)
		keyLen, err := encoding.PeekLength(key)
		if err != nil {
			__antithesis_instrumentation__.Notify(274184)
			return nil, false, scratch, err
		} else {
			__antithesis_instrumentation__.Notify(274185)
		}
		__antithesis_instrumentation__.Notify(274183)
		vecs.BytesCols[colIdx].Set(rowIdx, key[:keyLen])
		return key[keyLen:], false, scratch, nil
	} else {
		__antithesis_instrumentation__.Notify(274186)
	}
	__antithesis_instrumentation__.Notify(274175)

	var rkey []byte
	var err error
	switch valType.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(274187)
		var i int64
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274204)
			rkey, i, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(274205)
			rkey, i, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(274188)
		vecs.BoolCols[colIdx][rowIdx] = i != 0
	case types.IntFamily, types.DateFamily:
		__antithesis_instrumentation__.Notify(274189)
		var i int64
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274206)
			rkey, i, err = encoding.DecodeVarintAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(274207)
			rkey, i, err = encoding.DecodeVarintDescending(key)
		}
		__antithesis_instrumentation__.Notify(274190)
		switch valType.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(274208)
			vecs.Int16Cols[colIdx][rowIdx] = int16(i)
		case 32:
			__antithesis_instrumentation__.Notify(274209)
			vecs.Int32Cols[colIdx][rowIdx] = int32(i)
		case 0, 64:
			__antithesis_instrumentation__.Notify(274210)
			vecs.Int64Cols[colIdx][rowIdx] = i
		default:
			__antithesis_instrumentation__.Notify(274211)
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(274191)
		var f float64
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274212)
			rkey, f, err = encoding.DecodeFloatAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(274213)
			rkey, f, err = encoding.DecodeFloatDescending(key)
		}
		__antithesis_instrumentation__.Notify(274192)
		vecs.Float64Cols[colIdx][rowIdx] = f
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(274193)
		var d apd.Decimal
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274214)
			rkey, d, err = encoding.DecodeDecimalAscending(key, scratch[:0])
		} else {
			__antithesis_instrumentation__.Notify(274215)
			rkey, d, err = encoding.DecodeDecimalDescending(key, scratch[:0])
		}
		__antithesis_instrumentation__.Notify(274194)
		vecs.DecimalCols[colIdx][rowIdx] = d
	case types.BytesFamily, types.StringFamily, types.UuidFamily:
		__antithesis_instrumentation__.Notify(274195)
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274216)

			rkey, scratch, err = encoding.DecodeBytesAscendingDeepCopy(key, scratch[:0])
		} else {
			__antithesis_instrumentation__.Notify(274217)
			rkey, scratch, err = encoding.DecodeBytesDescending(key, scratch[:0])
		}
		__antithesis_instrumentation__.Notify(274196)

		vecs.BytesCols[colIdx].Set(rowIdx, scratch)
	case types.TimestampFamily, types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(274197)
		var t time.Time
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274218)
			rkey, t, err = encoding.DecodeTimeAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(274219)
			rkey, t, err = encoding.DecodeTimeDescending(key)
		}
		__antithesis_instrumentation__.Notify(274198)
		vecs.TimestampCols[colIdx][rowIdx] = t
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(274199)
		var d duration.Duration
		if dir == descpb.IndexDescriptor_ASC {
			__antithesis_instrumentation__.Notify(274220)
			rkey, d, err = encoding.DecodeDurationAscending(key)
		} else {
			__antithesis_instrumentation__.Notify(274221)
			rkey, d, err = encoding.DecodeDurationDescending(key)
		}
		__antithesis_instrumentation__.Notify(274200)
		vecs.IntervalCols[colIdx][rowIdx] = d
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(274201)

		var jsonLen int
		jsonLen, err = encoding.PeekLength(key)
		vecs.JSONCols[colIdx].Bytes.Set(rowIdx, key[:jsonLen])
		rkey = key[jsonLen:]
	default:
		__antithesis_instrumentation__.Notify(274202)
		var d tree.Datum
		encDir := encoding.Ascending
		if dir == descpb.IndexDescriptor_DESC {
			__antithesis_instrumentation__.Notify(274222)
			encDir = encoding.Descending
		} else {
			__antithesis_instrumentation__.Notify(274223)
		}
		__antithesis_instrumentation__.Notify(274203)
		d, rkey, err = keyside.Decode(da, valType, key, encDir)
		vecs.DatumCols[colIdx].Set(rowIdx, d)
	}
	__antithesis_instrumentation__.Notify(274176)
	return rkey, false, scratch, err
}

func UnmarshalColumnValueToCol(
	da *tree.DatumAlloc,
	vecs *coldata.TypedVecs,
	vecIdx, rowIdx int,
	typ *types.T,
	value roachpb.Value,
) error {
	__antithesis_instrumentation__.Notify(274224)
	if value.RawBytes == nil {
		__antithesis_instrumentation__.Notify(274227)
		vecs.Nulls[vecIdx].SetNull(rowIdx)
	} else {
		__antithesis_instrumentation__.Notify(274228)
	}
	__antithesis_instrumentation__.Notify(274225)

	colIdx := vecs.ColsMap[vecIdx]

	var err error
	switch typ.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(274229)
		var v bool
		v, err = value.GetBool()
		vecs.BoolCols[colIdx][rowIdx] = v
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(274230)
		var v int64
		v, err = value.GetInt()
		switch typ.Width() {
		case 16:
			__antithesis_instrumentation__.Notify(274240)
			vecs.Int16Cols[colIdx][rowIdx] = int16(v)
		case 32:
			__antithesis_instrumentation__.Notify(274241)
			vecs.Int32Cols[colIdx][rowIdx] = int32(v)
		default:
			__antithesis_instrumentation__.Notify(274242)

			vecs.Int64Cols[colIdx][rowIdx] = v
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(274231)
		var v float64
		v, err = value.GetFloat()
		vecs.Float64Cols[colIdx][rowIdx] = v
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(274232)
		err = value.GetDecimalInto(&vecs.DecimalCols[colIdx][rowIdx])
	case types.BytesFamily, types.StringFamily, types.UuidFamily:
		__antithesis_instrumentation__.Notify(274233)
		var v []byte
		v, err = value.GetBytes()
		vecs.BytesCols[colIdx].Set(rowIdx, v)
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(274234)
		var v int64
		v, err = value.GetInt()
		vecs.Int64Cols[colIdx][rowIdx] = v
	case types.TimestampFamily, types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(274235)
		var v time.Time
		v, err = value.GetTime()
		vecs.TimestampCols[colIdx][rowIdx] = v
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(274236)
		var v duration.Duration
		v, err = value.GetDuration()
		vecs.IntervalCols[colIdx][rowIdx] = v
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(274237)
		var v []byte
		v, err = value.GetBytes()
		vecs.JSONCols[colIdx].Bytes.Set(rowIdx, v)

	default:
		__antithesis_instrumentation__.Notify(274238)
		var d tree.Datum
		d, err = valueside.UnmarshalLegacy(da, typ, value)
		if err != nil {
			__antithesis_instrumentation__.Notify(274243)
			return err
		} else {
			__antithesis_instrumentation__.Notify(274244)
		}
		__antithesis_instrumentation__.Notify(274239)
		vecs.DatumCols[colIdx].Set(rowIdx, d)
	}
	__antithesis_instrumentation__.Notify(274226)
	return err
}
