package colserde

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/apache/arrow/go/arrow/array"
	"github.com/apache/arrow/go/arrow/memory"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/memsize"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/errors"
)

type ArrowBatchConverter struct {
	typs []*types.T

	builders struct {
		boolBuilder *array.BooleanBuilder
	}

	scratch struct {
		arrowData []*array.Data

		buffers [][]*memory.Buffer
	}
}

func NewArrowBatchConverter(typs []*types.T) (*ArrowBatchConverter, error) {
	__antithesis_instrumentation__.Notify(54850)
	c := &ArrowBatchConverter{typs: typs}
	c.builders.boolBuilder = array.NewBooleanBuilder(memory.DefaultAllocator)
	c.scratch.arrowData = make([]*array.Data, len(typs))
	c.scratch.buffers = make([][]*memory.Buffer, len(typs))
	for i := range c.scratch.buffers {
		__antithesis_instrumentation__.Notify(54852)

		c.scratch.buffers[i] = make([]*memory.Buffer, 0, 3)
	}
	__antithesis_instrumentation__.Notify(54851)
	return c, nil
}

func (c *ArrowBatchConverter) BatchToArrow(batch coldata.Batch) ([]*array.Data, error) {
	__antithesis_instrumentation__.Notify(54853)
	if batch.Width() != len(c.typs) {
		__antithesis_instrumentation__.Notify(54856)
		return nil, errors.AssertionFailedf("mismatched batch width and schema length: %d != %d", batch.Width(), len(c.typs))
	} else {
		__antithesis_instrumentation__.Notify(54857)
	}
	__antithesis_instrumentation__.Notify(54854)
	n := batch.Length()
	for vecIdx, typ := range c.typs {
		__antithesis_instrumentation__.Notify(54858)
		vec := batch.ColVec(vecIdx)

		var nulls *coldata.Nulls
		if vec.MaybeHasNulls() {
			__antithesis_instrumentation__.Notify(54865)
			nulls = vec.Nulls()

			nulls.Truncate(batch.Length())
		} else {
			__antithesis_instrumentation__.Notify(54866)
		}
		__antithesis_instrumentation__.Notify(54859)

		if typ.Family() == types.BoolFamily {
			__antithesis_instrumentation__.Notify(54867)
			c.builders.boolBuilder.AppendValues(vec.Bool()[:n], nil)
			c.scratch.arrowData[vecIdx] = c.builders.boolBuilder.NewBooleanArray().Data()

			var arrowBitmap []byte
			if nulls != nil {
				__antithesis_instrumentation__.Notify(54869)
				arrowBitmap = nulls.NullBitmap()
			} else {
				__antithesis_instrumentation__.Notify(54870)
			}
			__antithesis_instrumentation__.Notify(54868)
			c.scratch.arrowData[vecIdx].Buffers()[0] = memory.NewBufferBytes(arrowBitmap)
			continue
		} else {
			__antithesis_instrumentation__.Notify(54871)
		}
		__antithesis_instrumentation__.Notify(54860)

		var (
			values       []byte
			offsetsBytes []byte

			dataHeader *reflect.SliceHeader

			datumSize int64
		)

		switch typeconv.TypeFamilyToCanonicalTypeFamily(typ.Family()) {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(54872)
			var offsets []int32
			values, offsets = vec.Bytes().ToArrowSerializationFormat(n)
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		case types.JsonFamily:
			__antithesis_instrumentation__.Notify(54873)
			var offsets []int32
			values, offsets = vec.JSON().Bytes.ToArrowSerializationFormat(n)
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		case types.IntFamily:
			__antithesis_instrumentation__.Notify(54874)
			switch typ.Width() {
			case 16:
				__antithesis_instrumentation__.Notify(54885)
				ints := vec.Int16()[:n]
				dataHeader = (*reflect.SliceHeader)(unsafe.Pointer(&ints))
				datumSize = memsize.Int16
			case 32:
				__antithesis_instrumentation__.Notify(54886)
				ints := vec.Int32()[:n]
				dataHeader = (*reflect.SliceHeader)(unsafe.Pointer(&ints))
				datumSize = memsize.Int32
			case 0, 64:
				__antithesis_instrumentation__.Notify(54887)
				ints := vec.Int64()[:n]
				dataHeader = (*reflect.SliceHeader)(unsafe.Pointer(&ints))
				datumSize = memsize.Int64
			default:
				__antithesis_instrumentation__.Notify(54888)
				panic(fmt.Sprintf("unexpected int width: %d", typ.Width()))
			}

		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(54875)
			floats := vec.Float64()[:n]
			dataHeader = (*reflect.SliceHeader)(unsafe.Pointer(&floats))
			datumSize = memsize.Float64

		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(54876)
			offsets := make([]int32, 0, n+1)
			decimals := vec.Decimal()[:n]

			values = make([]byte, 0, n*16)
			for i := range decimals {
				__antithesis_instrumentation__.Notify(54889)
				offsets = append(offsets, int32(len(values)))
				if nulls != nil && func() bool {
					__antithesis_instrumentation__.Notify(54891)
					return nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54892)
					continue
				} else {
					__antithesis_instrumentation__.Notify(54893)
				}
				__antithesis_instrumentation__.Notify(54890)

				values = decimals[i].Append(values, 'G')
			}
			__antithesis_instrumentation__.Notify(54877)
			offsets = append(offsets, int32(len(values)))
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		case types.IntervalFamily:
			__antithesis_instrumentation__.Notify(54878)
			offsets := make([]int32, 0, n+1)
			intervals := vec.Interval()[:n]
			intervalSize := int(memsize.Int64) * 3

			values = make([]byte, intervalSize*n)
			var curNonNullInterval int
			for i := range intervals {
				__antithesis_instrumentation__.Notify(54894)
				offsets = append(offsets, int32(curNonNullInterval*intervalSize))
				if nulls != nil && func() bool {
					__antithesis_instrumentation__.Notify(54897)
					return nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54898)
					continue
				} else {
					__antithesis_instrumentation__.Notify(54899)
				}
				__antithesis_instrumentation__.Notify(54895)
				nanos, months, days, err := intervals[i].Encode()
				if err != nil {
					__antithesis_instrumentation__.Notify(54900)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(54901)
				}
				__antithesis_instrumentation__.Notify(54896)
				curSlice := values[intervalSize*curNonNullInterval : intervalSize*(curNonNullInterval+1)]
				binary.LittleEndian.PutUint64(curSlice[0:memsize.Int64], uint64(nanos))
				binary.LittleEndian.PutUint64(curSlice[memsize.Int64:memsize.Int64*2], uint64(months))
				binary.LittleEndian.PutUint64(curSlice[memsize.Int64*2:memsize.Int64*3], uint64(days))
				curNonNullInterval++
			}
			__antithesis_instrumentation__.Notify(54879)
			values = values[:intervalSize*curNonNullInterval]
			offsets = append(offsets, int32(len(values)))
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		case types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(54880)
			offsets := make([]int32, 0, n+1)
			timestamps := vec.Timestamp()[:n]

			const timestampSize = 14
			values = make([]byte, 0, n*timestampSize)
			for i := range timestamps {
				__antithesis_instrumentation__.Notify(54902)
				offsets = append(offsets, int32(len(values)))
				if nulls != nil && func() bool {
					__antithesis_instrumentation__.Notify(54905)
					return nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54906)
					continue
				} else {
					__antithesis_instrumentation__.Notify(54907)
				}
				__antithesis_instrumentation__.Notify(54903)
				marshaled, err := timestamps[i].MarshalBinary()
				if err != nil {
					__antithesis_instrumentation__.Notify(54908)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(54909)
				}
				__antithesis_instrumentation__.Notify(54904)
				values = append(values, marshaled...)
			}
			__antithesis_instrumentation__.Notify(54881)
			offsets = append(offsets, int32(len(values)))
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(54882)
			offsets := make([]int32, 0, n+1)
			datums := vec.Datum().Window(0, n)

			size, _ := tree.DatumTypeSize(typ)
			values = make([]byte, 0, n*int(size))
			for i := 0; i < n; i++ {
				__antithesis_instrumentation__.Notify(54910)
				offsets = append(offsets, int32(len(values)))
				if nulls != nil && func() bool {
					__antithesis_instrumentation__.Notify(54912)
					return nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54913)
					continue
				} else {
					__antithesis_instrumentation__.Notify(54914)
				}
				__antithesis_instrumentation__.Notify(54911)
				var err error
				values, err = datums.MarshalAt(values, i)
				if err != nil {
					__antithesis_instrumentation__.Notify(54915)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(54916)
				}
			}
			__antithesis_instrumentation__.Notify(54883)
			offsets = append(offsets, int32(len(values)))
			unsafeCastOffsetsArray(offsets, &offsetsBytes)

		default:
			__antithesis_instrumentation__.Notify(54884)
			panic(fmt.Sprintf("unsupported type for conversion to arrow data %s", typ))
		}
		__antithesis_instrumentation__.Notify(54861)

		if values == nil {
			__antithesis_instrumentation__.Notify(54917)
			valuesHeader := (*reflect.SliceHeader)(unsafe.Pointer(&values))
			valuesHeader.Data = dataHeader.Data
			valuesHeader.Len = dataHeader.Len * int(datumSize)
			valuesHeader.Cap = dataHeader.Cap * int(datumSize)
		} else {
			__antithesis_instrumentation__.Notify(54918)
		}
		__antithesis_instrumentation__.Notify(54862)

		var arrowBitmap []byte
		if nulls != nil {
			__antithesis_instrumentation__.Notify(54919)
			arrowBitmap = nulls.NullBitmap()
		} else {
			__antithesis_instrumentation__.Notify(54920)
		}
		__antithesis_instrumentation__.Notify(54863)

		c.scratch.buffers[vecIdx] = c.scratch.buffers[vecIdx][:0]
		c.scratch.buffers[vecIdx] = append(c.scratch.buffers[vecIdx], memory.NewBufferBytes(arrowBitmap))
		if offsetsBytes != nil {
			__antithesis_instrumentation__.Notify(54921)
			c.scratch.buffers[vecIdx] = append(c.scratch.buffers[vecIdx], memory.NewBufferBytes(offsetsBytes))
		} else {
			__antithesis_instrumentation__.Notify(54922)
		}
		__antithesis_instrumentation__.Notify(54864)
		c.scratch.buffers[vecIdx] = append(c.scratch.buffers[vecIdx], memory.NewBufferBytes(values))

		c.scratch.arrowData[vecIdx] = array.NewData(
			nil, n, c.scratch.buffers[vecIdx], nil, 0, 0,
		)
	}
	__antithesis_instrumentation__.Notify(54855)
	return c.scratch.arrowData, nil
}

func unsafeCastOffsetsArray(offsetsInt32 []int32, offsetsBytes *[]byte) {
	__antithesis_instrumentation__.Notify(54923)

	int32Header := (*reflect.SliceHeader)(unsafe.Pointer(&offsetsInt32))
	bytesHeader := (*reflect.SliceHeader)(unsafe.Pointer(offsetsBytes))
	bytesHeader.Data = int32Header.Data
	bytesHeader.Len = int32Header.Len * int(memsize.Int32)
	bytesHeader.Cap = int32Header.Cap * int(memsize.Int32)
}

func (c *ArrowBatchConverter) ArrowToBatch(
	data []*array.Data, batchLength int, b coldata.Batch,
) error {
	__antithesis_instrumentation__.Notify(54924)
	if len(data) != len(c.typs) {
		__antithesis_instrumentation__.Notify(54927)
		return errors.Errorf("mismatched data and schema length: %d != %d", len(data), len(c.typs))
	} else {
		__antithesis_instrumentation__.Notify(54928)
	}
	__antithesis_instrumentation__.Notify(54925)

	for i, typ := range c.typs {
		__antithesis_instrumentation__.Notify(54929)
		vec := b.ColVec(i)
		d := data[i]

		data[i] = nil

		switch typeconv.TypeFamilyToCanonicalTypeFamily(typ.Family()) {
		case types.BoolFamily:
			__antithesis_instrumentation__.Notify(54930)
			boolArr := array.NewBooleanData(d)
			vec.Nulls().SetNullBitmap(boolArr.NullBitmapBytes(), batchLength)
			vecArr := vec.Bool()
			for i := 0; i < boolArr.Len(); i++ {
				__antithesis_instrumentation__.Notify(54947)
				vecArr[i] = boolArr.Value(i)
			}

		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(54931)
			deserializeArrowIntoBytes(d, vec.Nulls(), vec.Bytes(), batchLength)

		case types.JsonFamily:
			__antithesis_instrumentation__.Notify(54932)
			deserializeArrowIntoBytes(d, vec.Nulls(), &vec.JSON().Bytes, batchLength)

		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(54933)

			bytesArr := array.NewBinaryData(d)
			vec.Nulls().SetNullBitmap(bytesArr.NullBitmapBytes(), batchLength)

			var nulls *coldata.Nulls
			if vec.MaybeHasNulls() {
				__antithesis_instrumentation__.Notify(54948)
				nulls = vec.Nulls()
			} else {
				__antithesis_instrumentation__.Notify(54949)
			}
			__antithesis_instrumentation__.Notify(54934)
			bytes := bytesArr.ValueBytes()
			if bytes == nil {
				__antithesis_instrumentation__.Notify(54950)

				bytes = make([]byte, 0)
			} else {
				__antithesis_instrumentation__.Notify(54951)
			}
			__antithesis_instrumentation__.Notify(54935)
			offsets := bytesArr.ValueOffsets()
			vecArr := vec.Decimal()
			for i := 0; i < len(offsets)-1; i++ {
				__antithesis_instrumentation__.Notify(54952)
				if nulls == nil || func() bool {
					__antithesis_instrumentation__.Notify(54953)
					return !nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54954)
					if err := vecArr[i].UnmarshalText(bytes[offsets[i]:offsets[i+1]]); err != nil {
						__antithesis_instrumentation__.Notify(54955)
						return err
					} else {
						__antithesis_instrumentation__.Notify(54956)
					}
				} else {
					__antithesis_instrumentation__.Notify(54957)
				}
			}

		case types.TimestampTZFamily:
			__antithesis_instrumentation__.Notify(54936)

			bytesArr := array.NewBinaryData(d)
			vec.Nulls().SetNullBitmap(bytesArr.NullBitmapBytes(), batchLength)

			var nulls *coldata.Nulls
			if vec.MaybeHasNulls() {
				__antithesis_instrumentation__.Notify(54958)
				nulls = vec.Nulls()
			} else {
				__antithesis_instrumentation__.Notify(54959)
			}
			__antithesis_instrumentation__.Notify(54937)
			bytes := bytesArr.ValueBytes()
			if bytes == nil {
				__antithesis_instrumentation__.Notify(54960)

				bytes = make([]byte, 0)
			} else {
				__antithesis_instrumentation__.Notify(54961)
			}
			__antithesis_instrumentation__.Notify(54938)
			offsets := bytesArr.ValueOffsets()
			vecArr := vec.Timestamp()
			for i := 0; i < len(offsets)-1; i++ {
				__antithesis_instrumentation__.Notify(54962)
				if nulls == nil || func() bool {
					__antithesis_instrumentation__.Notify(54963)
					return !nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54964)
					if err := vecArr[i].UnmarshalBinary(bytes[offsets[i]:offsets[i+1]]); err != nil {
						__antithesis_instrumentation__.Notify(54965)
						return err
					} else {
						__antithesis_instrumentation__.Notify(54966)
					}
				} else {
					__antithesis_instrumentation__.Notify(54967)
				}
			}

		case types.IntervalFamily:
			__antithesis_instrumentation__.Notify(54939)

			bytesArr := array.NewBinaryData(d)
			vec.Nulls().SetNullBitmap(bytesArr.NullBitmapBytes(), batchLength)

			var nulls *coldata.Nulls
			if vec.MaybeHasNulls() {
				__antithesis_instrumentation__.Notify(54968)
				nulls = vec.Nulls()
			} else {
				__antithesis_instrumentation__.Notify(54969)
			}
			__antithesis_instrumentation__.Notify(54940)
			bytes := bytesArr.ValueBytes()
			if bytes == nil {
				__antithesis_instrumentation__.Notify(54970)

				bytes = make([]byte, 0)
			} else {
				__antithesis_instrumentation__.Notify(54971)
			}
			__antithesis_instrumentation__.Notify(54941)
			offsets := bytesArr.ValueOffsets()
			vecArr := vec.Interval()
			for i := 0; i < len(offsets)-1; i++ {
				__antithesis_instrumentation__.Notify(54972)
				if nulls == nil || func() bool {
					__antithesis_instrumentation__.Notify(54973)
					return !nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54974)
					intervalBytes := bytes[offsets[i]:offsets[i+1]]
					var err error
					vecArr[i], err = duration.Decode(
						int64(binary.LittleEndian.Uint64(intervalBytes[0:memsize.Int64])),
						int64(binary.LittleEndian.Uint64(intervalBytes[memsize.Int64:memsize.Int64*2])),
						int64(binary.LittleEndian.Uint64(intervalBytes[memsize.Int64*2:memsize.Int64*3])),
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(54975)
						return err
					} else {
						__antithesis_instrumentation__.Notify(54976)
					}
				} else {
					__antithesis_instrumentation__.Notify(54977)
				}
			}

		case typeconv.DatumVecCanonicalTypeFamily:
			__antithesis_instrumentation__.Notify(54942)
			bytesArr := array.NewBinaryData(d)
			vec.Nulls().SetNullBitmap(bytesArr.NullBitmapBytes(), batchLength)

			var nulls *coldata.Nulls
			if vec.MaybeHasNulls() {
				__antithesis_instrumentation__.Notify(54978)
				nulls = vec.Nulls()
			} else {
				__antithesis_instrumentation__.Notify(54979)
			}
			__antithesis_instrumentation__.Notify(54943)
			bytes := bytesArr.ValueBytes()
			if bytes == nil {
				__antithesis_instrumentation__.Notify(54980)

				bytes = make([]byte, 0)
			} else {
				__antithesis_instrumentation__.Notify(54981)
			}
			__antithesis_instrumentation__.Notify(54944)
			offsets := bytesArr.ValueOffsets()
			vecArr := vec.Datum()
			for i := 0; i < len(offsets)-1; i++ {
				__antithesis_instrumentation__.Notify(54982)
				if nulls == nil || func() bool {
					__antithesis_instrumentation__.Notify(54983)
					return !nulls.NullAt(i) == true
				}() == true {
					__antithesis_instrumentation__.Notify(54984)
					if err := vecArr.UnmarshalTo(i, bytes[offsets[i]:offsets[i+1]]); err != nil {
						__antithesis_instrumentation__.Notify(54985)
						return err
					} else {
						__antithesis_instrumentation__.Notify(54986)
					}
				} else {
					__antithesis_instrumentation__.Notify(54987)
				}
			}

		default:
			__antithesis_instrumentation__.Notify(54945)

			var col coldata.Column
			switch typeconv.TypeFamilyToCanonicalTypeFamily(typ.Family()) {
			case types.IntFamily:
				__antithesis_instrumentation__.Notify(54988)
				switch typ.Width() {
				case 16:
					__antithesis_instrumentation__.Notify(54991)
					intArr := array.NewInt16Data(d)
					vec.Nulls().SetNullBitmap(intArr.NullBitmapBytes(), batchLength)
					int16s := coldata.Int16s(intArr.Int16Values())
					col = int16s[:len(int16s):len(int16s)]
				case 32:
					__antithesis_instrumentation__.Notify(54992)
					intArr := array.NewInt32Data(d)
					vec.Nulls().SetNullBitmap(intArr.NullBitmapBytes(), batchLength)
					int32s := coldata.Int32s(intArr.Int32Values())
					col = int32s[:len(int32s):len(int32s)]
				case 0, 64:
					__antithesis_instrumentation__.Notify(54993)
					intArr := array.NewInt64Data(d)
					vec.Nulls().SetNullBitmap(intArr.NullBitmapBytes(), batchLength)
					int64s := coldata.Int64s(intArr.Int64Values())
					col = int64s[:len(int64s):len(int64s)]
				default:
					__antithesis_instrumentation__.Notify(54994)
					panic(fmt.Sprintf("unexpected int width: %d", typ.Width()))
				}
			case types.FloatFamily:
				__antithesis_instrumentation__.Notify(54989)
				floatArr := array.NewFloat64Data(d)
				vec.Nulls().SetNullBitmap(floatArr.NullBitmapBytes(), batchLength)
				float64s := coldata.Float64s(floatArr.Float64Values())
				col = float64s[:len(float64s):len(float64s)]
			default:
				__antithesis_instrumentation__.Notify(54990)
				panic(
					fmt.Sprintf("unsupported type for conversion to column batch %s", d.DataType().Name()),
				)
			}
			__antithesis_instrumentation__.Notify(54946)
			vec.SetCol(col)
		}
	}
	__antithesis_instrumentation__.Notify(54926)
	b.SetSelection(false)
	b.SetLength(batchLength)
	return nil
}

func deserializeArrowIntoBytes(
	d *array.Data, nulls *coldata.Nulls, bytes *coldata.Bytes, batchLength int,
) {
	__antithesis_instrumentation__.Notify(54995)
	bytesArr := array.NewBinaryData(d)
	nulls.SetNullBitmap(bytesArr.NullBitmapBytes(), batchLength)
	b := bytesArr.ValueBytes()
	if b == nil {
		__antithesis_instrumentation__.Notify(54997)

		b = make([]byte, 0)
	} else {
		__antithesis_instrumentation__.Notify(54998)
	}
	__antithesis_instrumentation__.Notify(54996)

	offsets := bytesArr.ValueOffsets()
	b = b[:len(b):len(b)]
	offsets = offsets[:len(offsets):len(offsets)]
	coldata.BytesFromArrowSerializationFormat(bytes, b, offsets)
}
