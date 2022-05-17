package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/geo"
	"github.com/cockroachdb/cockroach/pkg/geo/geopb"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bitarray"
	"github.com/cockroachdb/cockroach/pkg/util/duration"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
	"github.com/lib/pq/oid"
)

const exportParquetFilePatternDefault = exportFilePatternPart + ".parquet"

type parquetExporter struct {
	buf            *bytes.Buffer
	parquetWriter  *goparquet.FileWriter
	schema         *parquetschema.SchemaDefinition
	parquetColumns []ParquetColumn
	compression    roachpb.IOFileFormat_Compression
}

func (c *parquetExporter) Write(record map[string]interface{}) error {
	__antithesis_instrumentation__.Notify(493186)
	return c.parquetWriter.AddData(record)
}

func (c *parquetExporter) Flush() error {
	__antithesis_instrumentation__.Notify(493187)
	return nil
}

func (c *parquetExporter) Close() error {
	__antithesis_instrumentation__.Notify(493188)
	return c.parquetWriter.Close()
}

func (c *parquetExporter) Bytes() []byte {
	__antithesis_instrumentation__.Notify(493189)
	return c.buf.Bytes()
}

func (c *parquetExporter) ResetBuffer() {
	__antithesis_instrumentation__.Notify(493190)
	c.buf.Reset()
	c.parquetWriter = c.buildFileWriter()
}

func (c *parquetExporter) Len() int {
	__antithesis_instrumentation__.Notify(493191)
	return c.buf.Len()
}

func (c *parquetExporter) FileName(spec execinfrapb.ExportSpec, part string) string {
	__antithesis_instrumentation__.Notify(493192)
	pattern := exportParquetFilePatternDefault
	if spec.NamePattern != "" {
		__antithesis_instrumentation__.Notify(493195)
		pattern = spec.NamePattern
	} else {
		__antithesis_instrumentation__.Notify(493196)
	}
	__antithesis_instrumentation__.Notify(493193)

	fileName := strings.Replace(pattern, exportFilePatternPart, part, -1)
	suffix := ""
	switch spec.Format.Compression {
	case roachpb.IOFileFormat_Gzip:
		__antithesis_instrumentation__.Notify(493197)
		suffix = ".gz"
	case roachpb.IOFileFormat_Snappy:
		__antithesis_instrumentation__.Notify(493198)
		suffix = ".snappy"
	default:
		__antithesis_instrumentation__.Notify(493199)
	}
	__antithesis_instrumentation__.Notify(493194)
	fileName += suffix
	return fileName
}

func (c *parquetExporter) buildFileWriter() *goparquet.FileWriter {
	__antithesis_instrumentation__.Notify(493200)
	var parquetCompression parquet.CompressionCodec
	switch c.compression {
	case roachpb.IOFileFormat_Gzip:
		__antithesis_instrumentation__.Notify(493202)
		parquetCompression = parquet.CompressionCodec_GZIP
	case roachpb.IOFileFormat_Snappy:
		__antithesis_instrumentation__.Notify(493203)
		parquetCompression = parquet.CompressionCodec_SNAPPY
	default:
		__antithesis_instrumentation__.Notify(493204)
		parquetCompression = parquet.CompressionCodec_UNCOMPRESSED
	}
	__antithesis_instrumentation__.Notify(493201)
	pw := goparquet.NewFileWriter(c.buf,
		goparquet.WithCompressionCodec(parquetCompression),
		goparquet.WithSchemaDefinition(c.schema),
	)
	return pw
}

func newParquetExporter(sp execinfrapb.ExportSpec, typs []*types.T) (*parquetExporter, error) {
	__antithesis_instrumentation__.Notify(493205)
	var exporter *parquetExporter

	buf := bytes.NewBuffer([]byte{})
	parquetColumns, err := newParquetColumns(typs, sp)
	if err != nil {
		__antithesis_instrumentation__.Notify(493207)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493208)
	}
	__antithesis_instrumentation__.Notify(493206)
	schema := newParquetSchema(parquetColumns)

	exporter = &parquetExporter{
		buf:            buf,
		schema:         schema,
		parquetColumns: parquetColumns,
		compression:    sp.Format.Compression,
	}
	return exporter, nil
}

type ParquetColumn struct {
	name     string
	crbdType *types.T

	definition *parquetschema.ColumnDefinition

	encodeFn func(datum tree.Datum) (interface{}, error)

	DecodeFn func(interface{}) (tree.Datum, error)
}

func newParquetColumns(typs []*types.T, sp execinfrapb.ExportSpec) ([]ParquetColumn, error) {
	__antithesis_instrumentation__.Notify(493209)
	parquetColumns := make([]ParquetColumn, len(typs))
	for i := 0; i < len(typs); i++ {
		__antithesis_instrumentation__.Notify(493211)
		parquetCol, err := NewParquetColumn(typs[i], sp.ColNames[i], sp.Format.Parquet.ColNullability[i])
		if err != nil {
			__antithesis_instrumentation__.Notify(493213)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(493214)
		}
		__antithesis_instrumentation__.Notify(493212)
		parquetColumns[i] = parquetCol
	}
	__antithesis_instrumentation__.Notify(493210)
	return parquetColumns, nil
}

func populateLogicalStringCol(schemaEl *parquet.SchemaElement) {
	__antithesis_instrumentation__.Notify(493215)
	schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
	schemaEl.LogicalType = parquet.NewLogicalType()
	schemaEl.LogicalType.STRING = parquet.NewStringType()
	schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_UTF8)
}

func RoundtripStringer(d tree.Datum) string {
	__antithesis_instrumentation__.Notify(493216)
	fmtCtx := tree.NewFmtCtx(tree.FmtBareStrings)
	d.Format(fmtCtx)
	return fmtCtx.CloseAndGetString()
}

func NewParquetColumn(typ *types.T, name string, nullable bool) (ParquetColumn, error) {
	__antithesis_instrumentation__.Notify(493217)
	col := ParquetColumn{}
	col.definition = new(parquetschema.ColumnDefinition)
	col.definition.SchemaElement = parquet.NewSchemaElement()
	col.name = name
	col.crbdType = typ

	schemaEl := col.definition.SchemaElement

	schemaEl.RepetitionType = parquet.FieldRepetitionTypePtr(parquet.FieldRepetitionType_OPTIONAL)
	if !nullable {
		__antithesis_instrumentation__.Notify(493220)
		schemaEl.RepetitionType = parquet.FieldRepetitionTypePtr(parquet.FieldRepetitionType_REQUIRED)
	} else {
		__antithesis_instrumentation__.Notify(493221)
	}
	__antithesis_instrumentation__.Notify(493218)
	schemaEl.Name = col.name

	switch typ.Family() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(493222)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BOOLEAN)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493268)
			return bool(*d.(*tree.DBool)), nil
		}
		__antithesis_instrumentation__.Notify(493223)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493269)
			return tree.MakeDBool(tree.DBool(x.(bool))), nil
		}

	case types.StringFamily:
		__antithesis_instrumentation__.Notify(493224)
		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493270)
			return []byte(*d.(*tree.DString)), nil
		}
		__antithesis_instrumentation__.Notify(493225)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493271)
			return tree.NewDString(string(x.([]byte))), nil
		}
	case types.CollatedStringFamily:
		__antithesis_instrumentation__.Notify(493226)
		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493272)
			return []byte(d.(*tree.DCollatedString).Contents), nil
		}
		__antithesis_instrumentation__.Notify(493227)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493273)
			return tree.NewDCollatedString(string(x.([]byte)), typ.Locale(), &tree.CollationEnvironment{})
		}
	case types.INetFamily:
		__antithesis_instrumentation__.Notify(493228)
		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493274)
			return []byte(d.(*tree.DIPAddr).IPAddr.String()), nil
		}
		__antithesis_instrumentation__.Notify(493229)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493275)
			return tree.ParseDIPAddrFromINetString(string(x.([]byte)))
		}
	case types.JsonFamily:
		__antithesis_instrumentation__.Notify(493230)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.JSON = parquet.NewJsonType()
		schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_JSON)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493276)
			return []byte(d.(*tree.DJSON).JSON.String()), nil
		}
		__antithesis_instrumentation__.Notify(493231)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493277)
			jsonStr := string(x.([]byte))
			return tree.ParseDJSON(jsonStr)
		}

	case types.IntFamily:
		__antithesis_instrumentation__.Notify(493232)
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.INTEGER = parquet.NewIntType()
		schemaEl.LogicalType.INTEGER.IsSigned = true
		if typ.Oid() == oid.T_int8 {
			__antithesis_instrumentation__.Notify(493278)
			schemaEl.Type = parquet.TypePtr(parquet.Type_INT64)
			schemaEl.LogicalType.INTEGER.BitWidth = int8(64)
			schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_INT_64)
			col.encodeFn = func(d tree.Datum) (interface{}, error) {
				__antithesis_instrumentation__.Notify(493280)
				return int64(*d.(*tree.DInt)), nil
			}
			__antithesis_instrumentation__.Notify(493279)
			col.DecodeFn = func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(493281)
				return tree.NewDInt(tree.DInt(x.(int64))), nil
			}
		} else {
			__antithesis_instrumentation__.Notify(493282)
			schemaEl.Type = parquet.TypePtr(parquet.Type_INT32)
			schemaEl.LogicalType.INTEGER.BitWidth = int8(32)
			schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_INT_32)
			col.encodeFn = func(d tree.Datum) (interface{}, error) {
				__antithesis_instrumentation__.Notify(493284)
				return int32(*d.(*tree.DInt)), nil
			}
			__antithesis_instrumentation__.Notify(493283)
			col.DecodeFn = func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(493285)
				return tree.NewDInt(tree.DInt(x.(int32))), nil
			}
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(493233)
		if typ.Oid() == oid.T_float4 {
			__antithesis_instrumentation__.Notify(493286)
			schemaEl.Type = parquet.TypePtr(parquet.Type_FLOAT)
			col.encodeFn = func(d tree.Datum) (interface{}, error) {
				__antithesis_instrumentation__.Notify(493288)
				h := float32(*d.(*tree.DFloat))
				return h, nil
			}
			__antithesis_instrumentation__.Notify(493287)
			col.DecodeFn = func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(493289)

				hS := fmt.Sprintf("%f", x.(float32))
				return tree.ParseDFloat(hS)
			}
		} else {
			__antithesis_instrumentation__.Notify(493290)
			schemaEl.Type = parquet.TypePtr(parquet.Type_DOUBLE)
			col.encodeFn = func(d tree.Datum) (interface{}, error) {
				__antithesis_instrumentation__.Notify(493292)
				return float64(*d.(*tree.DFloat)), nil
			}
			__antithesis_instrumentation__.Notify(493291)
			col.DecodeFn = func(x interface{}) (tree.Datum, error) {
				__antithesis_instrumentation__.Notify(493293)
				return tree.NewDFloat(tree.DFloat(x.(float64))), nil
			}
		}
	case types.DecimalFamily:
		__antithesis_instrumentation__.Notify(493234)

		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)

		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.DECIMAL = parquet.NewDecimalType()

		schemaEl.LogicalType.DECIMAL.Scale = typ.Scale()
		schemaEl.LogicalType.DECIMAL.Precision = typ.Precision()
		schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_DECIMAL)

		if typ.Scale() == 0 {
			__antithesis_instrumentation__.Notify(493294)
			schemaEl.LogicalType.DECIMAL.Scale = math.MaxInt32
		} else {
			__antithesis_instrumentation__.Notify(493295)
		}
		__antithesis_instrumentation__.Notify(493235)
		if typ.Precision() == 0 {
			__antithesis_instrumentation__.Notify(493296)
			schemaEl.LogicalType.DECIMAL.Precision = math.MaxInt32
		} else {
			__antithesis_instrumentation__.Notify(493297)
		}
		__antithesis_instrumentation__.Notify(493236)

		schemaEl.Scale = &schemaEl.LogicalType.DECIMAL.Scale
		schemaEl.Precision = &schemaEl.LogicalType.DECIMAL.Precision

		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493298)
			dec := d.(*tree.DDecimal).Decimal
			return []byte(dec.String()), nil
		}
		__antithesis_instrumentation__.Notify(493237)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493299)

			return tree.ParseDDecimal(string(x.([]byte)))
		}
	case types.UuidFamily:
		__antithesis_instrumentation__.Notify(493238)

		schemaEl.Type = parquet.TypePtr(parquet.Type_FIXED_LEN_BYTE_ARRAY)
		byteArraySize := int32(uuid.Size)
		schemaEl.TypeLength = &byteArraySize
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.UUID = parquet.NewUUIDType()
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493300)
			return d.(*tree.DUuid).UUID.GetBytes(), nil
		}
		__antithesis_instrumentation__.Notify(493239)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493301)
			return tree.ParseDUuidFromBytes(x.([]byte))
		}
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(493240)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493302)
			return []byte(*d.(*tree.DBytes)), nil
		}
		__antithesis_instrumentation__.Notify(493241)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493303)
			return tree.NewDBytes(tree.DBytes(x.([]byte))), nil
		}
	case types.BitFamily:
		__antithesis_instrumentation__.Notify(493242)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493304)

			baS := RoundtripStringer(d.(*tree.DBitArray))
			return []byte(baS), nil
		}
		__antithesis_instrumentation__.Notify(493243)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493305)
			ba, err := bitarray.Parse(string(x.([]byte)))
			return &tree.DBitArray{BitArray: ba}, err
		}
	case types.EnumFamily:
		__antithesis_instrumentation__.Notify(493244)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.ENUM = parquet.NewEnumType()
		schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_ENUM)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493306)
			return []byte(d.(*tree.DEnum).LogicalRep), nil
		}
		__antithesis_instrumentation__.Notify(493245)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493307)
			return tree.MakeDEnumFromLogicalRepresentation(typ, string(x.([]byte)))
		}
	case types.Box2DFamily:
		__antithesis_instrumentation__.Notify(493246)
		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493308)
			return []byte(d.(*tree.DBox2D).CartesianBoundingBox.Repr()), nil
		}
		__antithesis_instrumentation__.Notify(493247)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493309)
			b, err := geo.ParseCartesianBoundingBox(string(x.([]byte)))
			if err != nil {
				__antithesis_instrumentation__.Notify(493311)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493312)
			}
			__antithesis_instrumentation__.Notify(493310)
			return tree.NewDBox2D(b), nil
		}
	case types.GeographyFamily:
		__antithesis_instrumentation__.Notify(493248)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493313)
			return []byte(d.(*tree.DGeography).EWKB()), nil
		}
		__antithesis_instrumentation__.Notify(493249)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493314)
			g, err := geo.ParseGeographyFromEWKB(geopb.EWKB(x.([]byte)))
			if err != nil {
				__antithesis_instrumentation__.Notify(493316)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493317)
			}
			__antithesis_instrumentation__.Notify(493315)
			return &tree.DGeography{Geography: g}, nil
		}
	case types.GeometryFamily:
		__antithesis_instrumentation__.Notify(493250)
		schemaEl.Type = parquet.TypePtr(parquet.Type_BYTE_ARRAY)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493318)
			return []byte(d.(*tree.DGeometry).EWKB()), nil
		}
		__antithesis_instrumentation__.Notify(493251)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493319)
			g, err := geo.ParseGeometryFromEWKBUnsafe(geopb.EWKB(x.([]byte)))
			if err != nil {
				__antithesis_instrumentation__.Notify(493321)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493322)
			}
			__antithesis_instrumentation__.Notify(493320)
			return &tree.DGeometry{Geometry: g}, nil
		}
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(493252)

		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493323)
			date := d.(*tree.DDate)
			ds := RoundtripStringer(date)
			return []byte(ds), nil
		}
		__antithesis_instrumentation__.Notify(493253)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493324)
			dStr := string(x.([]byte))
			d, dependCtx, err := tree.ParseDDate(nil, dStr)
			if dependCtx {
				__antithesis_instrumentation__.Notify(493326)
				return nil, errors.Newf("decoding date %s failed. depends on context", string(x.([]byte)))
			} else {
				__antithesis_instrumentation__.Notify(493327)
			}
			__antithesis_instrumentation__.Notify(493325)
			return d, err
		}
	case types.TimeFamily:
		__antithesis_instrumentation__.Notify(493254)
		schemaEl.Type = parquet.TypePtr(parquet.Type_INT64)
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.TIME = parquet.NewTimeType()
		t := parquet.NewTimeUnit()
		t.MICROS = parquet.NewMicroSeconds()
		schemaEl.LogicalType.TIME.Unit = t
		schemaEl.LogicalType.TIME.IsAdjustedToUTC = true
		schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_TIME_MICROS)

		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493328)

			time := d.(*tree.DTime)
			m := int64(*time)
			return m, nil
		}
		__antithesis_instrumentation__.Notify(493255)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493329)
			return tree.MakeDTime(timeofday.TimeOfDay(x.(int64))), nil
		}
	case types.TimeTZFamily:
		__antithesis_instrumentation__.Notify(493256)

		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493330)
			return []byte(d.(*tree.DTimeTZ).TimeTZ.String()), nil
		}
		__antithesis_instrumentation__.Notify(493257)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493331)
			d, dependsOnCtx, err := tree.ParseDTimeTZ(nil, string(x.([]byte)), time.Microsecond)
			if dependsOnCtx {
				__antithesis_instrumentation__.Notify(493333)
				return nil, errors.New("parsed time depends on context")
			} else {
				__antithesis_instrumentation__.Notify(493334)
			}
			__antithesis_instrumentation__.Notify(493332)
			return d, err
		}
	case types.IntervalFamily:
		__antithesis_instrumentation__.Notify(493258)

		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493335)
			return []byte(d.(*tree.DInterval).ValueAsISO8601String()), nil
		}
		__antithesis_instrumentation__.Notify(493259)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493336)
			return tree.ParseDInterval(duration.IntervalStyle_ISO_8601, string(x.([]byte)))
		}
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(493260)

		populateLogicalStringCol(schemaEl)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493337)
			ts := RoundtripStringer(d.(*tree.DTimestamp))
			return []byte(ts), nil
		}
		__antithesis_instrumentation__.Notify(493261)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493338)

			dtStr := string(x.([]byte))
			d, dependsOnCtx, err := tree.ParseDTimestamp(nil, dtStr, time.Microsecond)
			if dependsOnCtx {
				__antithesis_instrumentation__.Notify(493341)
				return nil, errors.New("TimestampTZ depends on context")
			} else {
				__antithesis_instrumentation__.Notify(493342)
			}
			__antithesis_instrumentation__.Notify(493339)
			if err != nil {
				__antithesis_instrumentation__.Notify(493343)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493344)
			}
			__antithesis_instrumentation__.Notify(493340)

			d.Time = d.Time.UTC()
			return d, nil
		}

	case types.TimestampTZFamily:
		__antithesis_instrumentation__.Notify(493262)

		populateLogicalStringCol(schemaEl)

		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493345)
			ts := RoundtripStringer(d.(*tree.DTimestampTZ))
			return []byte(ts), nil
		}
		__antithesis_instrumentation__.Notify(493263)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493346)
			dtStr := string(x.([]byte))
			d, dependsOnCtx, err := tree.ParseDTimestampTZ(nil, dtStr, time.Microsecond)
			if dependsOnCtx {
				__antithesis_instrumentation__.Notify(493349)
				return nil, errors.New("TimestampTZ depends on context")
			} else {
				__antithesis_instrumentation__.Notify(493350)
			}
			__antithesis_instrumentation__.Notify(493347)
			if err != nil {
				__antithesis_instrumentation__.Notify(493351)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(493352)
			}
			__antithesis_instrumentation__.Notify(493348)

			d.Time = d.Time.UTC()
			return d, nil
		}
	case types.ArrayFamily:
		__antithesis_instrumentation__.Notify(493264)

		grandChild, err := NewParquetColumn(typ.ArrayContents(), "element", true)
		if err != nil {
			__antithesis_instrumentation__.Notify(493353)
			return col, err
		} else {
			__antithesis_instrumentation__.Notify(493354)
		}
		__antithesis_instrumentation__.Notify(493265)

		child := &parquetschema.ColumnDefinition{}
		child.SchemaElement = parquet.NewSchemaElement()
		child.SchemaElement.RepetitionType = parquet.FieldRepetitionTypePtr(parquet.
			FieldRepetitionType_REPEATED)
		child.SchemaElement.Name = "list"
		child.Children = []*parquetschema.ColumnDefinition{grandChild.definition}
		ngc := int32(len(child.Children))
		child.SchemaElement.NumChildren = &ngc

		col.definition.Children = []*parquetschema.ColumnDefinition{child}
		nc := int32(len(col.definition.Children))
		child.SchemaElement.NumChildren = &nc
		schemaEl.LogicalType = parquet.NewLogicalType()
		schemaEl.LogicalType.LIST = parquet.NewListType()
		schemaEl.ConvertedType = parquet.ConvertedTypePtr(parquet.ConvertedType_LIST)
		col.encodeFn = func(d tree.Datum) (interface{}, error) {
			__antithesis_instrumentation__.Notify(493355)
			datumArr := d.(*tree.DArray)
			els := make([]map[string]interface{}, datumArr.Len())
			for i, elt := range datumArr.Array {
				__antithesis_instrumentation__.Notify(493357)
				var el interface{}
				if elt.ResolvedType().Family() == types.UnknownFamily {
					__antithesis_instrumentation__.Notify(493359)

				} else {
					__antithesis_instrumentation__.Notify(493360)
					el, err = grandChild.encodeFn(elt)
					if err != nil {
						__antithesis_instrumentation__.Notify(493361)
						return col, err
					} else {
						__antithesis_instrumentation__.Notify(493362)
					}
				}
				__antithesis_instrumentation__.Notify(493358)
				els[i] = map[string]interface{}{"element": el}
			}
			__antithesis_instrumentation__.Notify(493356)
			encEl := map[string]interface{}{"list": els}
			return encEl, nil
		}
		__antithesis_instrumentation__.Notify(493266)
		col.DecodeFn = func(x interface{}) (tree.Datum, error) {
			__antithesis_instrumentation__.Notify(493363)

			datumArr := tree.NewDArray(typ.ArrayContents())
			datumArr.Array = []tree.Datum{}

			intermediate := x.(map[string]interface{})
			vals := intermediate["list"].([]map[string]interface{})
			if _, nonEmpty := vals[0]["element"]; !nonEmpty {
				__antithesis_instrumentation__.Notify(493365)
				if len(vals) > 1 {
					__antithesis_instrumentation__.Notify(493366)
					return nil, errors.New("array is empty, it shouldn't have a length greater than 1")
				} else {
					__antithesis_instrumentation__.Notify(493367)
				}
			} else {
				__antithesis_instrumentation__.Notify(493368)
				for _, elMap := range vals {
					__antithesis_instrumentation__.Notify(493369)
					itemDatum, err := grandChild.DecodeFn(elMap["element"])
					if err != nil {
						__antithesis_instrumentation__.Notify(493371)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(493372)
					}
					__antithesis_instrumentation__.Notify(493370)
					err = datumArr.Append(itemDatum)
					if err != nil {
						__antithesis_instrumentation__.Notify(493373)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(493374)
					}
				}
			}
			__antithesis_instrumentation__.Notify(493364)
			return datumArr, nil
		}
	default:
		__antithesis_instrumentation__.Notify(493267)
		return col, errors.Errorf("parquet export does not support the %v type yet", typ.Family())
	}
	__antithesis_instrumentation__.Notify(493219)

	return col, nil
}

func newParquetSchema(parquetFields []ParquetColumn) *parquetschema.SchemaDefinition {
	__antithesis_instrumentation__.Notify(493375)
	schemaDefinition := new(parquetschema.SchemaDefinition)
	schemaDefinition.RootColumn = new(parquetschema.ColumnDefinition)
	schemaDefinition.RootColumn.SchemaElement = parquet.NewSchemaElement()

	for i := 0; i < len(parquetFields); i++ {
		__antithesis_instrumentation__.Notify(493377)
		schemaDefinition.RootColumn.Children = append(schemaDefinition.RootColumn.Children,
			parquetFields[i].definition)
		schemaDefinition.RootColumn.SchemaElement.Name = "root"
	}
	__antithesis_instrumentation__.Notify(493376)
	return schemaDefinition
}

func newParquetWriterProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.ExportSpec,
	input execinfra.RowSource,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(493378)
	c := &parquetWriterProcessor{
		flowCtx:     flowCtx,
		processorID: processorID,
		spec:        spec,
		input:       input,
		output:      output,
	}
	semaCtx := tree.MakeSemaContext()
	if err := c.out.Init(&execinfrapb.PostProcessSpec{}, c.OutputTypes(), &semaCtx, flowCtx.NewEvalCtx()); err != nil {
		__antithesis_instrumentation__.Notify(493380)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493381)
	}
	__antithesis_instrumentation__.Notify(493379)
	return c, nil
}

type parquetWriterProcessor struct {
	flowCtx     *execinfra.FlowCtx
	processorID int32
	spec        execinfrapb.ExportSpec
	input       execinfra.RowSource
	out         execinfra.ProcOutputHelper
	output      execinfra.RowReceiver
}

var _ execinfra.Processor = &parquetWriterProcessor{}

func (sp *parquetWriterProcessor) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(493382)
	res := make([]*types.T, len(colinfo.ExportColumns))
	for i := range res {
		__antithesis_instrumentation__.Notify(493384)
		res[i] = colinfo.ExportColumns[i].Typ
	}
	__antithesis_instrumentation__.Notify(493383)
	return res
}

func (sp *parquetWriterProcessor) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(493385)
	return false
}

func (sp *parquetWriterProcessor) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(493386)
	ctx, span := tracing.ChildSpan(ctx, "parquetWriter")
	defer span.Finish()

	instanceID := sp.flowCtx.EvalCtx.NodeID.SQLInstanceID()
	uniqueID := builtins.GenerateUniqueInt(instanceID)

	err := func() error {
		__antithesis_instrumentation__.Notify(493388)
		typs := sp.input.OutputTypes()
		sp.input.Start(ctx)
		input := execinfra.MakeNoMetadataRowSource(sp.input, sp.output)
		alloc := &tree.DatumAlloc{}

		exporter, err := newParquetExporter(sp.spec, typs)
		if err != nil {
			__antithesis_instrumentation__.Notify(493391)
			return err
		} else {
			__antithesis_instrumentation__.Notify(493392)
		}
		__antithesis_instrumentation__.Notify(493389)

		parquetRow := make(map[string]interface{}, len(typs))
		chunk := 0
		done := false
		for {
			__antithesis_instrumentation__.Notify(493393)
			var rows int64
			exporter.ResetBuffer()
			for {
				__antithesis_instrumentation__.Notify(493403)

				if int64(exporter.buf.Len()) >= sp.spec.ChunkSize {
					__antithesis_instrumentation__.Notify(493409)
					break
				} else {
					__antithesis_instrumentation__.Notify(493410)
				}
				__antithesis_instrumentation__.Notify(493404)
				if sp.spec.ChunkRows > 0 && func() bool {
					__antithesis_instrumentation__.Notify(493411)
					return rows >= sp.spec.ChunkRows == true
				}() == true {
					__antithesis_instrumentation__.Notify(493412)
					break
				} else {
					__antithesis_instrumentation__.Notify(493413)
				}
				__antithesis_instrumentation__.Notify(493405)
				row, err := input.NextRow()
				if err != nil {
					__antithesis_instrumentation__.Notify(493414)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493415)
				}
				__antithesis_instrumentation__.Notify(493406)
				if row == nil {
					__antithesis_instrumentation__.Notify(493416)
					done = true
					break
				} else {
					__antithesis_instrumentation__.Notify(493417)
				}
				__antithesis_instrumentation__.Notify(493407)
				rows++

				for i, ed := range row {
					__antithesis_instrumentation__.Notify(493418)
					if ed.IsNull() {
						__antithesis_instrumentation__.Notify(493419)
						parquetRow[exporter.parquetColumns[i].name] = nil
					} else {
						__antithesis_instrumentation__.Notify(493420)
						if err := ed.EnsureDecoded(typs[i], alloc); err != nil {
							__antithesis_instrumentation__.Notify(493423)
							return err
						} else {
							__antithesis_instrumentation__.Notify(493424)
						}
						__antithesis_instrumentation__.Notify(493421)

						edNative, err := exporter.parquetColumns[i].encodeFn(tree.UnwrapDatum(nil, ed.Datum))
						if err != nil {
							__antithesis_instrumentation__.Notify(493425)
							return err
						} else {
							__antithesis_instrumentation__.Notify(493426)
						}
						__antithesis_instrumentation__.Notify(493422)
						parquetRow[exporter.parquetColumns[i].name] = edNative
					}
				}
				__antithesis_instrumentation__.Notify(493408)
				if err := exporter.Write(parquetRow); err != nil {
					__antithesis_instrumentation__.Notify(493427)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493428)
				}
			}
			__antithesis_instrumentation__.Notify(493394)
			if rows < 1 {
				__antithesis_instrumentation__.Notify(493429)
				break
			} else {
				__antithesis_instrumentation__.Notify(493430)
			}
			__antithesis_instrumentation__.Notify(493395)
			if err := exporter.Flush(); err != nil {
				__antithesis_instrumentation__.Notify(493431)
				return errors.Wrap(err, "failed to flush parquet exporter")
			} else {
				__antithesis_instrumentation__.Notify(493432)
			}
			__antithesis_instrumentation__.Notify(493396)

			err = exporter.Close()
			if err != nil {
				__antithesis_instrumentation__.Notify(493433)
				return errors.Wrapf(err, "failed to close exporting exporter")
			} else {
				__antithesis_instrumentation__.Notify(493434)
			}
			__antithesis_instrumentation__.Notify(493397)

			conf, err := cloud.ExternalStorageConfFromURI(sp.spec.Destination, sp.spec.User())
			if err != nil {
				__antithesis_instrumentation__.Notify(493435)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493436)
			}
			__antithesis_instrumentation__.Notify(493398)
			es, err := sp.flowCtx.Cfg.ExternalStorage(ctx, conf)
			if err != nil {
				__antithesis_instrumentation__.Notify(493437)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493438)
			}
			__antithesis_instrumentation__.Notify(493399)
			defer es.Close()

			part := fmt.Sprintf("n%d.%d", uniqueID, chunk)
			chunk++
			filename := exporter.FileName(sp.spec, part)

			size := exporter.Len()

			if err := cloud.WriteFile(ctx, es, filename, bytes.NewReader(exporter.Bytes())); err != nil {
				__antithesis_instrumentation__.Notify(493439)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493440)
			}
			__antithesis_instrumentation__.Notify(493400)
			res := rowenc.EncDatumRow{
				rowenc.DatumToEncDatum(
					types.String,
					tree.NewDString(filename),
				),
				rowenc.DatumToEncDatum(
					types.Int,
					tree.NewDInt(tree.DInt(rows)),
				),
				rowenc.DatumToEncDatum(
					types.Int,
					tree.NewDInt(tree.DInt(size)),
				),
			}

			cs, err := sp.out.EmitRow(ctx, res, sp.output)
			if err != nil {
				__antithesis_instrumentation__.Notify(493441)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493442)
			}
			__antithesis_instrumentation__.Notify(493401)
			if cs != execinfra.NeedMoreRows {
				__antithesis_instrumentation__.Notify(493443)

				return errors.New("unexpected closure of consumer")
			} else {
				__antithesis_instrumentation__.Notify(493444)
			}
			__antithesis_instrumentation__.Notify(493402)
			if done {
				__antithesis_instrumentation__.Notify(493445)
				break
			} else {
				__antithesis_instrumentation__.Notify(493446)
			}
		}
		__antithesis_instrumentation__.Notify(493390)

		return nil
	}()
	__antithesis_instrumentation__.Notify(493387)

	execinfra.DrainAndClose(
		ctx, sp.output, err, func(context.Context) { __antithesis_instrumentation__.Notify(493447) }, sp.input)
}

func init() {
	rowexec.NewParquetWriterProcessor = newParquetWriterProcessor
}
