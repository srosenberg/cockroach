package pgwire

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"encoding/binary"
	"math"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/lex"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgwirebase"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/sqltelemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/errorutil/unimplemented"
	"github.com/cockroachdb/cockroach/pkg/util/ipaddr"
	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/timetz"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

type pgType struct {
	oid oid.Oid

	size int
}

func pgTypeForParserType(t *types.T) pgType {
	__antithesis_instrumentation__.Notify(561777)
	size := -1
	if s, variable := tree.DatumTypeSize(t); !variable {
		__antithesis_instrumentation__.Notify(561779)
		size = int(s)
	} else {
		__antithesis_instrumentation__.Notify(561780)
	}
	__antithesis_instrumentation__.Notify(561778)
	return pgType{
		oid:  t.Oid(),
		size: size,
	}
}

func writeTextBool(b *writeBuffer, v bool) {
	__antithesis_instrumentation__.Notify(561781)
	b.putInt32(1)
	b.writeByte(tree.PgwireFormatBool(v))
}

func writeTextInt64(b *writeBuffer, v int64) {
	__antithesis_instrumentation__.Notify(561782)

	s := strconv.AppendInt(b.putbuf[4:4], v, 10)
	b.putInt32(int32(len(s)))
	b.write(s)
}

func writeTextFloat64(b *writeBuffer, fl float64, conv sessiondatapb.DataConversionConfig) {
	__antithesis_instrumentation__.Notify(561783)
	var s []byte

	if math.IsInf(fl, 1) {
		__antithesis_instrumentation__.Notify(561785)
		s = []byte("Infinity")
	} else {
		__antithesis_instrumentation__.Notify(561786)
		if math.IsInf(fl, -1) {
			__antithesis_instrumentation__.Notify(561787)
			s = []byte("-Infinity")
		} else {
			__antithesis_instrumentation__.Notify(561788)

			s = strconv.AppendFloat(b.putbuf[4:4], fl, 'g', conv.GetFloatPrec(), 64)
		}
	}
	__antithesis_instrumentation__.Notify(561784)
	b.putInt32(int32(len(s)))
	b.write(s)
}

func writeTextBytes(b *writeBuffer, v string, conv sessiondatapb.DataConversionConfig) {
	__antithesis_instrumentation__.Notify(561789)
	result := lex.EncodeByteArrayToRawBytes(v, conv.BytesEncodeFormat, false)
	b.putInt32(int32(len(result)))
	b.write([]byte(result))
}

func writeTextUUID(b *writeBuffer, v uuid.UUID) {
	__antithesis_instrumentation__.Notify(561790)

	s := b.putbuf[4 : 4+36]
	v.StringBytes(s)
	b.putInt32(int32(len(s)))
	b.write(s)
}

func writeTextString(b *writeBuffer, v string, t *types.T) {
	__antithesis_instrumentation__.Notify(561791)
	b.writeLengthPrefixedString(tree.ResolveBlankPaddedChar(v, t))
}

func writeTextTimestamp(b *writeBuffer, v time.Time) {
	__antithesis_instrumentation__.Notify(561792)

	s := formatTs(v, nil, b.putbuf[4:4])
	b.putInt32(int32(len(s)))
	b.write(s)
}

func writeTextTimestampTZ(b *writeBuffer, v time.Time, sessionLoc *time.Location) {
	__antithesis_instrumentation__.Notify(561793)

	s := formatTs(v, sessionLoc, b.putbuf[4:4])
	b.putInt32(int32(len(s)))
	b.write(s)
}

func (b *writeBuffer) writeTextDatum(
	ctx context.Context,
	d tree.Datum,
	conv sessiondatapb.DataConversionConfig,
	sessionLoc *time.Location,
	t *types.T,
) {
	__antithesis_instrumentation__.Notify(561794)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(561797)
		log.Infof(ctx, "pgwire writing TEXT datum of type: %T, %#v", d, d)
	} else {
		__antithesis_instrumentation__.Notify(561798)
	}
	__antithesis_instrumentation__.Notify(561795)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(561799)

		b.putInt32(-1)
		return
	} else {
		__antithesis_instrumentation__.Notify(561800)
	}
	__antithesis_instrumentation__.Notify(561796)
	writeTextDatumNotNull(b, d, conv, sessionLoc, t)
}

func writeTextDatumNotNull(
	b *writeBuffer,
	d tree.Datum,
	conv sessiondatapb.DataConversionConfig,
	sessionLoc *time.Location,
	t *types.T,
) {
	__antithesis_instrumentation__.Notify(561801)
	oldDCC := b.textFormatter.SetDataConversionConfig(conv)
	defer b.textFormatter.SetDataConversionConfig(oldDCC)
	switch v := tree.UnwrapDatum(nil, d).(type) {
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(561802)
		b.textFormatter.FormatNode(v)
		b.writeFromFmtCtx(b.textFormatter)

	case *tree.DBool:
		__antithesis_instrumentation__.Notify(561803)
		writeTextBool(b, bool(*v))

	case *tree.DInt:
		__antithesis_instrumentation__.Notify(561804)
		writeTextInt64(b, int64(*v))

	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(561805)
		fl := float64(*v)
		writeTextFloat64(b, fl, conv)

	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(561806)
		b.writeLengthPrefixedDatum(v)

	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(561807)
		writeTextBytes(b, string(*v), conv)

	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(561808)
		writeTextUUID(b, v.UUID)

	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(561809)
		b.writeLengthPrefixedString(v.IPAddr.String())

	case *tree.DString:
		__antithesis_instrumentation__.Notify(561810)
		writeTextString(b, string(*v), t)

	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(561811)
		b.writeLengthPrefixedString(tree.ResolveBlankPaddedChar(v.Contents, t))

	case *tree.DDate:
		__antithesis_instrumentation__.Notify(561812)
		b.textFormatter.FormatNode(v)
		b.writeFromFmtCtx(b.textFormatter)

	case *tree.DTime:
		__antithesis_instrumentation__.Notify(561813)

		s := formatTime(timeofday.TimeOfDay(*v), b.putbuf[4:4])
		b.putInt32(int32(len(s)))
		b.write(s)

	case *tree.DTimeTZ:
		__antithesis_instrumentation__.Notify(561814)

		s := formatTimeTZ(v.TimeTZ, b.putbuf[4:4])
		b.putInt32(int32(len(s)))
		b.write(s)

	case *tree.DVoid:
		__antithesis_instrumentation__.Notify(561815)
		b.putInt32(0)

	case *tree.DBox2D:
		__antithesis_instrumentation__.Notify(561816)
		s := v.Repr()
		b.putInt32(int32(len(s)))
		b.write([]byte(s))

	case *tree.DGeography:
		__antithesis_instrumentation__.Notify(561817)
		s := v.Geography.EWKBHex()
		b.putInt32(int32(len(s)))
		b.write([]byte(s))

	case *tree.DGeometry:
		__antithesis_instrumentation__.Notify(561818)
		s := v.Geometry.EWKBHex()
		b.putInt32(int32(len(s)))
		b.write([]byte(s))

	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(561819)
		writeTextTimestamp(b, v.Time)

	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(561820)
		writeTextTimestampTZ(b, v.Time, sessionLoc)

	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(561821)
		b.textFormatter.FormatNode(v)
		b.writeFromFmtCtx(b.textFormatter)

	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(561822)
		b.writeLengthPrefixedString(v.JSON.String())

	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(561823)
		b.textFormatter.FormatNode(v)
		b.writeFromFmtCtx(b.textFormatter)

	case *tree.DArray:
		__antithesis_instrumentation__.Notify(561824)

		b.textFormatter.FormatNode(d)
		b.writeFromFmtCtx(b.textFormatter)

	case *tree.DOid:
		__antithesis_instrumentation__.Notify(561825)
		b.writeLengthPrefixedDatum(v)

	case *tree.DEnum:
		__antithesis_instrumentation__.Notify(561826)

		b.writeLengthPrefixedString(v.LogicalRep)

	default:
		__antithesis_instrumentation__.Notify(561827)
		b.setError(errors.Errorf("unsupported type %T", d))
	}
}

func getInt64(vecs *coldata.TypedVecs, vecIdx, rowIdx int, typ *types.T) int64 {
	__antithesis_instrumentation__.Notify(561828)
	colIdx := vecs.ColsMap[vecIdx]
	switch typ.Width() {
	case 16:
		__antithesis_instrumentation__.Notify(561829)
		return int64(vecs.Int16Cols[colIdx].Get(rowIdx))
	case 32:
		__antithesis_instrumentation__.Notify(561830)
		return int64(vecs.Int32Cols[colIdx].Get(rowIdx))
	default:
		__antithesis_instrumentation__.Notify(561831)
		return vecs.Int64Cols[colIdx].Get(rowIdx)
	}
}

func (b *writeBuffer) writeTextColumnarElement(
	ctx context.Context,
	vecs *coldata.TypedVecs,
	vecIdx int,
	rowIdx int,
	conv sessiondatapb.DataConversionConfig,
	sessionLoc *time.Location,
) {
	__antithesis_instrumentation__.Notify(561832)
	oldDCC := b.textFormatter.SetDataConversionConfig(conv)
	defer b.textFormatter.SetDataConversionConfig(oldDCC)
	typ := vecs.Vecs[vecIdx].Type()
	if log.V(2) {
		__antithesis_instrumentation__.Notify(561835)
		log.Infof(ctx, "pgwire writing TEXT columnar element of type: %s", typ)
	} else {
		__antithesis_instrumentation__.Notify(561836)
	}
	__antithesis_instrumentation__.Notify(561833)
	if vecs.Nulls[vecIdx].MaybeHasNulls() && func() bool {
		__antithesis_instrumentation__.Notify(561837)
		return vecs.Nulls[vecIdx].NullAt(rowIdx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(561838)

		b.putInt32(-1)
		return
	} else {
		__antithesis_instrumentation__.Notify(561839)
	}
	__antithesis_instrumentation__.Notify(561834)
	colIdx := vecs.ColsMap[vecIdx]
	switch typ.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(561840)
		writeTextBool(b, vecs.BoolCols[colIdx].Get(rowIdx))

	case types.IntFamily:
		__antithesis_instrumentation__.Notify(561841)
		writeTextInt64(b, getInt64(vecs, vecIdx, rowIdx, typ))

	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(561842)
		writeTextFloat64(b, vecs.Float64Cols[colIdx].Get(rowIdx), conv)

	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(561843)
		d := vecs.DecimalCols[colIdx].Get(rowIdx)

		b.writeLengthPrefixedString(d.String())

	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(561844)
		writeTextBytes(b, string(vecs.BytesCols[colIdx].Get(rowIdx)), conv)

	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(561845)
		id, err := uuid.FromBytes(vecs.BytesCols[colIdx].Get(rowIdx))
		if err != nil {
			__antithesis_instrumentation__.Notify(561854)
			panic(errors.Wrap(err, "unexpectedly couldn't retrieve UUID object"))
		} else {
			__antithesis_instrumentation__.Notify(561855)
		}
		__antithesis_instrumentation__.Notify(561846)
		writeTextUUID(b, id)

	case types.StringFamily:
		__antithesis_instrumentation__.Notify(561847)
		writeTextString(b, string(vecs.BytesCols[colIdx].Get(rowIdx)), typ)

	case types.DateFamily:
		__antithesis_instrumentation__.Notify(561848)
		tree.FormatDate(pgdate.MakeCompatibleDateFromDisk(vecs.Int64Cols[colIdx].Get(rowIdx)), b.textFormatter)
		b.writeFromFmtCtx(b.textFormatter)

	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(561849)
		writeTextTimestamp(b, vecs.TimestampCols[colIdx].Get(rowIdx))

	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(561850)
		writeTextTimestampTZ(b, vecs.TimestampCols[colIdx].Get(rowIdx), sessionLoc)

	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(561851)
		tree.FormatDuration(vecs.IntervalCols[colIdx].Get(rowIdx), b.textFormatter)
		b.writeFromFmtCtx(b.textFormatter)

	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(561852)
		b.writeLengthPrefixedString(vecs.JSONCols[colIdx].Get(rowIdx).String())

	default:
		__antithesis_instrumentation__.Notify(561853)

		writeTextDatumNotNull(b, vecs.DatumCols[colIdx].Get(rowIdx).(tree.Datum), conv, sessionLoc, typ)
	}
}

func writeBinaryBool(b *writeBuffer, v bool) {
	__antithesis_instrumentation__.Notify(561856)
	b.putInt32(1)
	if v {
		__antithesis_instrumentation__.Notify(561857)
		b.writeByte(1)
	} else {
		__antithesis_instrumentation__.Notify(561858)
		b.writeByte(0)
	}
}

func writeBinaryInt(b *writeBuffer, v int64, t *types.T) {
	__antithesis_instrumentation__.Notify(561859)
	switch t.Oid() {
	case oid.T_int2:
		__antithesis_instrumentation__.Notify(561860)
		b.putInt32(2)
		b.putInt16(int16(v))
	case oid.T_int4:
		__antithesis_instrumentation__.Notify(561861)
		b.putInt32(4)
		b.putInt32(int32(v))
	case oid.T_int8:
		__antithesis_instrumentation__.Notify(561862)
		b.putInt32(8)
		b.putInt64(v)
	default:
		__antithesis_instrumentation__.Notify(561863)
		b.setError(errors.Errorf("unsupported int oid: %v", t.Oid()))
	}
}

func writeBinaryFloat(b *writeBuffer, v float64, t *types.T) {
	__antithesis_instrumentation__.Notify(561864)
	switch t.Oid() {
	case oid.T_float4:
		__antithesis_instrumentation__.Notify(561865)
		b.putInt32(4)
		b.putInt32(int32(math.Float32bits(float32(v))))
	case oid.T_float8:
		__antithesis_instrumentation__.Notify(561866)
		b.putInt32(8)
		b.putInt64(int64(math.Float64bits(v)))
	default:
		__antithesis_instrumentation__.Notify(561867)
		b.setError(errors.Errorf("unsupported float oid: %v", t.Oid()))
	}
}

func writeBinaryDecimal(b *writeBuffer, v *apd.Decimal) {
	__antithesis_instrumentation__.Notify(561868)
	if v.Form != apd.Finite {
		__antithesis_instrumentation__.Notify(561876)
		b.putInt32(8)

		b.putInt32(0)
		if v.Form == apd.Infinite {
			__antithesis_instrumentation__.Notify(561878)

			if v.Negative {
				__antithesis_instrumentation__.Notify(561880)

				b.write([]byte{0xf0, 0, 0, 0x20})
			} else {
				__antithesis_instrumentation__.Notify(561881)

				b.write([]byte{0xd0, 0, 0, 0x20})
			}
			__antithesis_instrumentation__.Notify(561879)

			telemetry.Inc(sqltelemetry.BinaryDecimalInfinityCounter)
		} else {
			__antithesis_instrumentation__.Notify(561882)

			b.write([]byte{0xc0, 0, 0, 0})
		}
		__antithesis_instrumentation__.Notify(561877)
		return
	} else {
		__antithesis_instrumentation__.Notify(561883)
	}
	__antithesis_instrumentation__.Notify(561869)

	alloc := struct {
		pgNum pgwirebase.PGNumeric

		bigI apd.BigInt
	}{
		pgNum: pgwirebase.PGNumeric{

			Dscale: int16(-v.Exponent),
		},
	}

	if v.Sign() >= 0 {
		__antithesis_instrumentation__.Notify(561884)
		alloc.pgNum.Sign = pgwirebase.PGNumericPos
	} else {
		__antithesis_instrumentation__.Notify(561885)
		alloc.pgNum.Sign = pgwirebase.PGNumericNeg
	}
	__antithesis_instrumentation__.Notify(561870)

	isZero := func(r rune) bool {
		__antithesis_instrumentation__.Notify(561886)
		return r == '0'
	}
	__antithesis_instrumentation__.Notify(561871)

	digits := strings.TrimLeftFunc(alloc.bigI.Abs(&v.Coeff).String(), isZero)
	dweight := len(digits) - int(alloc.pgNum.Dscale) - 1
	digits = strings.TrimRightFunc(digits, isZero)

	if dweight >= 0 {
		__antithesis_instrumentation__.Notify(561887)
		alloc.pgNum.Weight = int16((dweight+1+pgwirebase.PGDecDigits-1)/pgwirebase.PGDecDigits - 1)
	} else {
		__antithesis_instrumentation__.Notify(561888)
		alloc.pgNum.Weight = int16(-((-dweight-1)/pgwirebase.PGDecDigits + 1))
	}
	__antithesis_instrumentation__.Notify(561872)
	offset := (int(alloc.pgNum.Weight)+1)*pgwirebase.PGDecDigits - (dweight + 1)
	alloc.pgNum.Ndigits = int16((len(digits) + offset + pgwirebase.PGDecDigits - 1) / pgwirebase.PGDecDigits)

	if len(digits) == 0 {
		__antithesis_instrumentation__.Notify(561889)
		offset = 0
		alloc.pgNum.Ndigits = 0
		alloc.pgNum.Weight = 0
	} else {
		__antithesis_instrumentation__.Notify(561890)
	}
	__antithesis_instrumentation__.Notify(561873)

	digitIdx := -offset

	nextDigit := func() int16 {
		__antithesis_instrumentation__.Notify(561891)
		var ndigit int16
		for nextDigitIdx := digitIdx + pgwirebase.PGDecDigits; digitIdx < nextDigitIdx; digitIdx++ {
			__antithesis_instrumentation__.Notify(561893)
			ndigit *= 10
			if digitIdx >= 0 && func() bool {
				__antithesis_instrumentation__.Notify(561894)
				return digitIdx < len(digits) == true
			}() == true {
				__antithesis_instrumentation__.Notify(561895)
				ndigit += int16(digits[digitIdx] - '0')
			} else {
				__antithesis_instrumentation__.Notify(561896)
			}
		}
		__antithesis_instrumentation__.Notify(561892)
		return ndigit
	}
	__antithesis_instrumentation__.Notify(561874)

	if alloc.pgNum.Dscale < 0 {
		__antithesis_instrumentation__.Notify(561897)
		alloc.pgNum.Dscale = 0
	} else {
		__antithesis_instrumentation__.Notify(561898)
	}
	__antithesis_instrumentation__.Notify(561875)

	b.putInt32(int32(2 * (4 + alloc.pgNum.Ndigits)))
	b.putInt16(alloc.pgNum.Ndigits)
	b.putInt16(alloc.pgNum.Weight)
	b.putInt16(int16(alloc.pgNum.Sign))
	b.putInt16(alloc.pgNum.Dscale)

	for digitIdx < len(digits) {
		__antithesis_instrumentation__.Notify(561899)
		b.putInt16(nextDigit())
	}
}

func writeBinaryBytes(b *writeBuffer, v []byte) {
	__antithesis_instrumentation__.Notify(561900)
	b.putInt32(int32(len(v)))
	b.write(v)
}

func writeBinaryString(b *writeBuffer, v string, t *types.T) {
	__antithesis_instrumentation__.Notify(561901)
	s := tree.ResolveBlankPaddedChar(v, t)
	if t.Oid() == oid.T_char && func() bool {
		__antithesis_instrumentation__.Notify(561903)
		return s == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(561904)

		s = string([]byte{0})
	} else {
		__antithesis_instrumentation__.Notify(561905)
	}
	__antithesis_instrumentation__.Notify(561902)
	b.writeLengthPrefixedString(s)
}

func writeBinaryTimestamp(b *writeBuffer, v time.Time) {
	__antithesis_instrumentation__.Notify(561906)
	b.putInt32(8)
	b.putInt64(timeToPgBinary(v, nil))
}

func writeBinaryTimestampTZ(b *writeBuffer, v time.Time, sessionLoc *time.Location) {
	__antithesis_instrumentation__.Notify(561907)
	b.putInt32(8)
	b.putInt64(timeToPgBinary(v, sessionLoc))
}

func writeBinaryDate(b *writeBuffer, v pgdate.Date) {
	__antithesis_instrumentation__.Notify(561908)
	b.putInt32(4)
	b.putInt32(v.PGEpochDays())
}

func writeBinaryInterval(b *writeBuffer, v duration.Duration) {
	__antithesis_instrumentation__.Notify(561909)
	b.putInt32(16)
	b.putInt64(v.Nanos() / int64(time.Microsecond/time.Nanosecond))
	b.putInt32(int32(v.Days))
	b.putInt32(int32(v.Months))
}

func writeBinaryJSON(b *writeBuffer, v json.JSON) {
	__antithesis_instrumentation__.Notify(561910)
	s := v.String()
	b.putInt32(int32(len(s) + 1))

	b.writeByte(1)
	b.writeString(s)
}

func (b *writeBuffer) writeBinaryDatum(
	ctx context.Context, d tree.Datum, sessionLoc *time.Location, t *types.T,
) {
	__antithesis_instrumentation__.Notify(561911)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(561914)
		log.Infof(ctx, "pgwire writing BINARY datum of type: %T, %#v", d, d)
	} else {
		__antithesis_instrumentation__.Notify(561915)
	}
	__antithesis_instrumentation__.Notify(561912)
	if d == tree.DNull {
		__antithesis_instrumentation__.Notify(561916)

		b.putInt32(-1)
		return
	} else {
		__antithesis_instrumentation__.Notify(561917)
	}
	__antithesis_instrumentation__.Notify(561913)
	writeBinaryDatumNotNull(ctx, b, d, sessionLoc, t)
}

func writeBinaryDatumNotNull(
	ctx context.Context, b *writeBuffer, d tree.Datum, sessionLoc *time.Location, t *types.T,
) {
	__antithesis_instrumentation__.Notify(561918)
	switch v := tree.UnwrapDatum(nil, d).(type) {
	case *tree.DBitArray:
		__antithesis_instrumentation__.Notify(561919)
		words, lastBitsUsed := v.EncodingParts()
		if len(words) == 0 {
			__antithesis_instrumentation__.Notify(561952)
			b.putInt32(4)
		} else {
			__antithesis_instrumentation__.Notify(561953)

			b.putInt32(4 + int32(8*(len(words)-1)) + int32((lastBitsUsed+7)/8))
		}
		__antithesis_instrumentation__.Notify(561920)
		bitLen := v.BitLen()
		b.putInt32(int32(bitLen))
		var byteBuf [8]byte
		for i := 0; i < len(words)-1; i++ {
			__antithesis_instrumentation__.Notify(561954)
			w := words[i]
			binary.BigEndian.PutUint64(byteBuf[:], w)
			b.write(byteBuf[:])
		}
		__antithesis_instrumentation__.Notify(561921)
		if len(words) > 0 {
			__antithesis_instrumentation__.Notify(561955)
			w := words[len(words)-1]
			for i := uint(0); i < uint(lastBitsUsed); i += 8 {
				__antithesis_instrumentation__.Notify(561956)
				c := byte(w >> (56 - i))
				b.writeByte(c)
			}
		} else {
			__antithesis_instrumentation__.Notify(561957)
		}

	case *tree.DBool:
		__antithesis_instrumentation__.Notify(561922)
		writeBinaryBool(b, bool(*v))

	case *tree.DInt:
		__antithesis_instrumentation__.Notify(561923)
		writeBinaryInt(b, int64(*v), t)

	case *tree.DFloat:
		__antithesis_instrumentation__.Notify(561924)
		writeBinaryFloat(b, float64(*v), t)

	case *tree.DDecimal:
		__antithesis_instrumentation__.Notify(561925)
		writeBinaryDecimal(b, &v.Decimal)

	case *tree.DBytes:
		__antithesis_instrumentation__.Notify(561926)
		writeBinaryBytes(b, []byte(*v))

	case *tree.DUuid:
		__antithesis_instrumentation__.Notify(561927)
		writeBinaryBytes(b, v.GetBytes())

	case *tree.DIPAddr:
		__antithesis_instrumentation__.Notify(561928)

		const pgIPAddrBinaryHeaderSize = 4
		if v.Family == ipaddr.IPv4family {
			__antithesis_instrumentation__.Notify(561958)
			b.putInt32(net.IPv4len + pgIPAddrBinaryHeaderSize)
			b.writeByte(pgwirebase.PGBinaryIPv4family)
			b.writeByte(v.Mask)
			b.writeByte(0)
			b.writeByte(byte(net.IPv4len))
			err := v.Addr.WriteIPv4Bytes(b)
			if err != nil {
				__antithesis_instrumentation__.Notify(561959)
				b.setError(err)
			} else {
				__antithesis_instrumentation__.Notify(561960)
			}
		} else {
			__antithesis_instrumentation__.Notify(561961)
			if v.Family == ipaddr.IPv6family {
				__antithesis_instrumentation__.Notify(561962)
				b.putInt32(net.IPv6len + pgIPAddrBinaryHeaderSize)
				b.writeByte(pgwirebase.PGBinaryIPv6family)
				b.writeByte(v.Mask)
				b.writeByte(0)
				b.writeByte(byte(net.IPv6len))
				err := v.Addr.WriteIPv6Bytes(b)
				if err != nil {
					__antithesis_instrumentation__.Notify(561963)
					b.setError(err)
				} else {
					__antithesis_instrumentation__.Notify(561964)
				}
			} else {
				__antithesis_instrumentation__.Notify(561965)
				b.setError(errors.Errorf("error encoding inet to pgBinary: %v", v.IPAddr))
			}
		}

	case *tree.DEnum:
		__antithesis_instrumentation__.Notify(561929)
		b.writeLengthPrefixedString(v.LogicalRep)

	case *tree.DString:
		__antithesis_instrumentation__.Notify(561930)
		writeBinaryString(b, string(*v), t)

	case *tree.DCollatedString:
		__antithesis_instrumentation__.Notify(561931)
		b.writeLengthPrefixedString(tree.ResolveBlankPaddedChar(v.Contents, t))

	case *tree.DTimestamp:
		__antithesis_instrumentation__.Notify(561932)
		writeBinaryTimestamp(b, v.Time)

	case *tree.DTimestampTZ:
		__antithesis_instrumentation__.Notify(561933)
		writeBinaryTimestampTZ(b, v.Time, sessionLoc)

	case *tree.DDate:
		__antithesis_instrumentation__.Notify(561934)
		writeBinaryDate(b, v.Date)

	case *tree.DTime:
		__antithesis_instrumentation__.Notify(561935)
		b.putInt32(8)
		b.putInt64(int64(*v))

	case *tree.DTimeTZ:
		__antithesis_instrumentation__.Notify(561936)
		b.putInt32(12)
		b.putInt64(int64(v.TimeOfDay))
		b.putInt32(v.OffsetSecs)

	case *tree.DInterval:
		__antithesis_instrumentation__.Notify(561937)
		writeBinaryInterval(b, v.Duration)

	case *tree.DTuple:
		__antithesis_instrumentation__.Notify(561938)
		initialLen := b.Len()

		b.putInt32(int32(0))

		b.putInt32(int32(len(v.D)))
		tupleTypes := t.TupleContents()
		for i, elem := range v.D {
			__antithesis_instrumentation__.Notify(561966)
			oid := tupleTypes[i].Oid()
			b.putInt32(int32(oid))
			b.writeBinaryDatum(ctx, elem, sessionLoc, tupleTypes[i])
		}
		__antithesis_instrumentation__.Notify(561939)

		lengthToWrite := b.Len() - (initialLen + 4)
		b.putInt32AtIndex(initialLen, int32(lengthToWrite))

	case *tree.DVoid:
		__antithesis_instrumentation__.Notify(561940)
		b.putInt32(0)

	case *tree.DBox2D:
		__antithesis_instrumentation__.Notify(561941)
		b.putInt32(32)
		b.putInt64(int64(math.Float64bits(v.LoX)))
		b.putInt64(int64(math.Float64bits(v.HiX)))
		b.putInt64(int64(math.Float64bits(v.LoY)))
		b.putInt64(int64(math.Float64bits(v.HiY)))

	case *tree.DGeography:
		__antithesis_instrumentation__.Notify(561942)
		b.putInt32(int32(len(v.EWKB())))
		b.write(v.EWKB())

	case *tree.DGeometry:
		__antithesis_instrumentation__.Notify(561943)
		b.putInt32(int32(len(v.EWKB())))
		b.write(v.EWKB())

	case *tree.DArray:
		__antithesis_instrumentation__.Notify(561944)
		if v.ParamTyp.Family() == types.ArrayFamily {
			__antithesis_instrumentation__.Notify(561967)
			b.setError(unimplemented.NewWithIssueDetail(32552,
				"binenc", "unsupported binary serialization of multidimensional arrays"))
			return
		} else {
			__antithesis_instrumentation__.Notify(561968)
		}
		__antithesis_instrumentation__.Notify(561945)

		initialLen := b.Len()

		b.putInt32(int32(0))

		var ndims int32 = 1
		if v.Len() == 0 {
			__antithesis_instrumentation__.Notify(561969)
			ndims = 0
		} else {
			__antithesis_instrumentation__.Notify(561970)
		}
		__antithesis_instrumentation__.Notify(561946)
		b.putInt32(ndims)
		hasNulls := 0
		if v.HasNulls {
			__antithesis_instrumentation__.Notify(561971)
			hasNulls = 1
		} else {
			__antithesis_instrumentation__.Notify(561972)
		}
		__antithesis_instrumentation__.Notify(561947)
		oid := v.ParamTyp.Oid()
		b.putInt32(int32(hasNulls))
		b.putInt32(int32(oid))
		if v.Len() > 0 {
			__antithesis_instrumentation__.Notify(561973)
			b.putInt32(int32(v.Len()))

			b.putInt32(1)
			for _, elem := range v.Array {
				__antithesis_instrumentation__.Notify(561974)
				b.writeBinaryDatum(ctx, elem, sessionLoc, v.ParamTyp)
			}
		} else {
			__antithesis_instrumentation__.Notify(561975)
		}
		__antithesis_instrumentation__.Notify(561948)

		lengthToWrite := b.Len() - (initialLen + 4)
		b.putInt32AtIndex(initialLen, int32(lengthToWrite))

	case *tree.DJSON:
		__antithesis_instrumentation__.Notify(561949)
		writeBinaryJSON(b, v.JSON)

	case *tree.DOid:
		__antithesis_instrumentation__.Notify(561950)
		b.putInt32(4)
		b.putInt32(int32(v.DInt))
	default:
		__antithesis_instrumentation__.Notify(561951)
		b.setError(errors.AssertionFailedf("unsupported type %T", d))
	}
}

func (b *writeBuffer) writeBinaryColumnarElement(
	ctx context.Context, vecs *coldata.TypedVecs, vecIdx int, rowIdx int, sessionLoc *time.Location,
) {
	__antithesis_instrumentation__.Notify(561976)
	typ := vecs.Vecs[vecIdx].Type()
	if log.V(2) {
		__antithesis_instrumentation__.Notify(561979)
		log.Infof(ctx, "pgwire writing BINARY columnar element of type: %s", typ)
	} else {
		__antithesis_instrumentation__.Notify(561980)
	}
	__antithesis_instrumentation__.Notify(561977)
	if vecs.Nulls[vecIdx].MaybeHasNulls() && func() bool {
		__antithesis_instrumentation__.Notify(561981)
		return vecs.Nulls[vecIdx].NullAt(rowIdx) == true
	}() == true {
		__antithesis_instrumentation__.Notify(561982)

		b.putInt32(-1)
		return
	} else {
		__antithesis_instrumentation__.Notify(561983)
	}
	__antithesis_instrumentation__.Notify(561978)
	colIdx := vecs.ColsMap[vecIdx]
	switch typ.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(561984)
		writeBinaryBool(b, vecs.BoolCols[colIdx].Get(rowIdx))

	case types.IntFamily:
		__antithesis_instrumentation__.Notify(561985)
		writeBinaryInt(b, getInt64(vecs, vecIdx, rowIdx, typ), typ)

	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(561986)
		writeBinaryFloat(b, vecs.Float64Cols[colIdx].Get(rowIdx), typ)

	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(561987)
		v := vecs.DecimalCols[colIdx].Get(rowIdx)
		writeBinaryDecimal(b, &v)

	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(561988)
		writeBinaryBytes(b, vecs.BytesCols[colIdx].Get(rowIdx))

	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(561989)
		writeBinaryBytes(b, vecs.BytesCols[colIdx].Get(rowIdx))

	case types.StringFamily:
		__antithesis_instrumentation__.Notify(561990)
		writeBinaryString(b, string(vecs.BytesCols[colIdx].Get(rowIdx)), typ)

	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(561991)
		writeBinaryTimestamp(b, vecs.TimestampCols[colIdx].Get(rowIdx))

	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(561992)
		writeBinaryTimestampTZ(b, vecs.TimestampCols[colIdx].Get(rowIdx), sessionLoc)

	case types.DateFamily:
		__antithesis_instrumentation__.Notify(561993)
		writeBinaryDate(b, pgdate.MakeCompatibleDateFromDisk(vecs.Int64Cols[colIdx].Get(rowIdx)))

	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(561994)
		writeBinaryInterval(b, vecs.IntervalCols[colIdx].Get(rowIdx))

	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(561995)
		writeBinaryJSON(b, vecs.JSONCols[colIdx].Get(rowIdx))

	default:
		__antithesis_instrumentation__.Notify(561996)

		writeBinaryDatumNotNull(ctx, b, vecs.DatumCols[colIdx].Get(rowIdx).(tree.Datum), sessionLoc, typ)
	}
}

const (
	pgTimeFormat              = "15:04:05.999999"
	pgTimeTZFormat            = pgTimeFormat + "-07"
	pgDateFormat              = "2006-01-02"
	pgTimeStampFormatNoOffset = pgDateFormat + " " + pgTimeFormat
	pgTimeStampFormat         = pgTimeStampFormatNoOffset + "-07"
	pgTime2400Format          = "24:00:00"
)

func formatTime(t timeofday.TimeOfDay, tmp []byte) []byte {
	__antithesis_instrumentation__.Notify(561997)

	if t == timeofday.Time2400 {
		__antithesis_instrumentation__.Notify(561999)
		return []byte(pgTime2400Format)
	} else {
		__antithesis_instrumentation__.Notify(562000)
	}
	__antithesis_instrumentation__.Notify(561998)
	return t.ToTime().AppendFormat(tmp, pgTimeFormat)
}

func formatTimeTZ(t timetz.TimeTZ, tmp []byte) []byte {
	__antithesis_instrumentation__.Notify(562001)
	format := pgTimeTZFormat
	if t.OffsetSecs%60 != 0 {
		__antithesis_instrumentation__.Notify(562004)
		format += ":00:00"
	} else {
		__antithesis_instrumentation__.Notify(562005)
		if t.OffsetSecs%3600 != 0 {
			__antithesis_instrumentation__.Notify(562006)
			format += ":00"
		} else {
			__antithesis_instrumentation__.Notify(562007)
		}
	}
	__antithesis_instrumentation__.Notify(562002)
	ret := t.ToTime().AppendFormat(tmp, format)

	if t.TimeOfDay == timeofday.Time2400 {
		__antithesis_instrumentation__.Notify(562008)

		var newRet []byte
		newRet = append(newRet, pgTime2400Format...)
		newRet = append(newRet, ret[len(pgTime2400Format):]...)
		ret = newRet
	} else {
		__antithesis_instrumentation__.Notify(562009)
	}
	__antithesis_instrumentation__.Notify(562003)
	return ret
}

func formatTs(t time.Time, offset *time.Location, tmp []byte) (b []byte) {
	__antithesis_instrumentation__.Notify(562010)
	var format string
	if offset != nil {
		__antithesis_instrumentation__.Notify(562012)
		format = pgTimeStampFormat
		if _, offsetSeconds := t.In(offset).Zone(); offsetSeconds%60 != 0 {
			__antithesis_instrumentation__.Notify(562013)
			format += ":00:00"
		} else {
			__antithesis_instrumentation__.Notify(562014)
			if offsetSeconds%3600 != 0 {
				__antithesis_instrumentation__.Notify(562015)
				format += ":00"
			} else {
				__antithesis_instrumentation__.Notify(562016)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(562017)
		format = pgTimeStampFormatNoOffset
	}
	__antithesis_instrumentation__.Notify(562011)
	return formatTsWithFormat(format, t, offset, tmp)
}

func formatTsWithFormat(format string, t time.Time, offset *time.Location, tmp []byte) (b []byte) {
	__antithesis_instrumentation__.Notify(562018)

	if offset != nil {
		__antithesis_instrumentation__.Notify(562022)
		t = t.In(offset)
	} else {
		__antithesis_instrumentation__.Notify(562023)
	}
	__antithesis_instrumentation__.Notify(562019)

	bc := false
	if t.Year() <= 0 {
		__antithesis_instrumentation__.Notify(562024)

		t = t.AddDate((-t.Year())*2+1, 0, 0)
		bc = true
	} else {
		__antithesis_instrumentation__.Notify(562025)
	}
	__antithesis_instrumentation__.Notify(562020)

	b = t.AppendFormat(tmp, format)
	if bc {
		__antithesis_instrumentation__.Notify(562026)
		b = append(b, " BC"...)
	} else {
		__antithesis_instrumentation__.Notify(562027)
	}
	__antithesis_instrumentation__.Notify(562021)
	return b
}

func timeToPgBinary(t time.Time, offset *time.Location) int64 {
	__antithesis_instrumentation__.Notify(562028)
	if offset != nil {
		__antithesis_instrumentation__.Notify(562030)
		t = t.In(offset)
	} else {
		__antithesis_instrumentation__.Notify(562031)
		t = t.UTC()
	}
	__antithesis_instrumentation__.Notify(562029)
	return duration.DiffMicros(t, pgwirebase.PGEpochJDate)
}
