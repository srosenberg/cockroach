package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"
	"unicode/utf8"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/lexbase"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/timeofday"
	"github.com/cockroachdb/errors"
	"github.com/linkedin/goavro/v2"
)

func nativeTimeToDatum(t time.Time, targetT *types.T) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(494866)
	duration := tree.TimeFamilyPrecisionToRoundDuration(targetT.Precision())
	switch targetT.Family() {
	case types.DateFamily:
		__antithesis_instrumentation__.Notify(494867)
		return tree.NewDDateFromTime(t)
	case types.TimestampFamily:
		__antithesis_instrumentation__.Notify(494868)
		return tree.MakeDTimestamp(t, duration)
	default:
		__antithesis_instrumentation__.Notify(494869)
		return nil, errors.New("type not supported")
	}
}

func nativeToDatum(
	x interface{}, targetT *types.T, avroT []string, evalCtx *tree.EvalContext,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(494870)
	var d tree.Datum

	switch v := x.(type) {
	case nil:
		__antithesis_instrumentation__.Notify(494874)

		return tree.DNull, nil
	case bool:
		__antithesis_instrumentation__.Notify(494875)
		if v {
			__antithesis_instrumentation__.Notify(494890)
			d = tree.DBoolTrue
		} else {
			__antithesis_instrumentation__.Notify(494891)
			d = tree.DBoolFalse
		}
	case int:
		__antithesis_instrumentation__.Notify(494876)
		d = tree.NewDInt(tree.DInt(v))
	case int32:
		__antithesis_instrumentation__.Notify(494877)
		d = tree.NewDInt(tree.DInt(v))
	case int64:
		__antithesis_instrumentation__.Notify(494878)
		d = tree.NewDInt(tree.DInt(v))
	case float32:
		__antithesis_instrumentation__.Notify(494879)
		d = tree.NewDFloat(tree.DFloat(v))
	case float64:
		__antithesis_instrumentation__.Notify(494880)
		d = tree.NewDFloat(tree.DFloat(v))
	case time.Time:
		__antithesis_instrumentation__.Notify(494881)
		return nativeTimeToDatum(v, targetT)
	case time.Duration:
		__antithesis_instrumentation__.Notify(494882)

		dU := v / time.Microsecond
		d = tree.MakeDTime(timeofday.TimeOfDay(dU))
	case []byte:
		__antithesis_instrumentation__.Notify(494883)
		if targetT.Identical(types.Bytes) {
			__antithesis_instrumentation__.Notify(494892)
			d = tree.NewDBytes(tree.DBytes(v))
		} else {
			__antithesis_instrumentation__.Notify(494893)

			return rowenc.ParseDatumStringAs(targetT, string(v), evalCtx)
		}
	case string:
		__antithesis_instrumentation__.Notify(494884)

		return rowenc.ParseDatumStringAs(targetT, v, evalCtx)
	case map[string]interface{}:
		__antithesis_instrumentation__.Notify(494885)
		for _, aT := range avroT {
			__antithesis_instrumentation__.Notify(494894)

			if val, ok := v[aT]; ok {
				__antithesis_instrumentation__.Notify(494895)
				return nativeToDatum(val, targetT, avroT, evalCtx)
			} else {
				__antithesis_instrumentation__.Notify(494896)
			}
		}
	case []interface{}:
		__antithesis_instrumentation__.Notify(494886)

		if targetT.ArrayContents() == nil {
			__antithesis_instrumentation__.Notify(494897)
			return nil, fmt.Errorf("cannot convert array to non-array type %s", targetT)
		} else {
			__antithesis_instrumentation__.Notify(494898)
		}
		__antithesis_instrumentation__.Notify(494887)
		eltAvroT, ok := familyToAvroT[targetT.ArrayContents().Family()]
		if !ok {
			__antithesis_instrumentation__.Notify(494899)
			return nil, fmt.Errorf("cannot convert avro array element to %s", targetT.ArrayContents())
		} else {
			__antithesis_instrumentation__.Notify(494900)
		}
		__antithesis_instrumentation__.Notify(494888)

		arr := tree.NewDArray(targetT.ArrayContents())
		for _, elt := range v {
			__antithesis_instrumentation__.Notify(494901)
			eltDatum, err := nativeToDatum(elt, targetT.ArrayContents(), eltAvroT, evalCtx)
			if err == nil {
				__antithesis_instrumentation__.Notify(494903)
				err = arr.Append(eltDatum)
			} else {
				__antithesis_instrumentation__.Notify(494904)
			}
			__antithesis_instrumentation__.Notify(494902)
			if err != nil {
				__antithesis_instrumentation__.Notify(494905)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(494906)
			}
		}
		__antithesis_instrumentation__.Notify(494889)
		d = arr
	}
	__antithesis_instrumentation__.Notify(494871)

	if d == nil {
		__antithesis_instrumentation__.Notify(494907)
		return nil, fmt.Errorf("cannot handle type %T when converting to %s", x, targetT)
	} else {
		__antithesis_instrumentation__.Notify(494908)
	}
	__antithesis_instrumentation__.Notify(494872)

	if !targetT.Equivalent(d.ResolvedType()) {
		__antithesis_instrumentation__.Notify(494909)
		return nil, fmt.Errorf("cannot convert type %s to %s", d.ResolvedType(), targetT)
	} else {
		__antithesis_instrumentation__.Notify(494910)
	}
	__antithesis_instrumentation__.Notify(494873)

	return d, nil
}

var familyToAvroT = map[types.Family][]string{

	types.BoolFamily:   {"bool", "boolean", "string"},
	types.IntFamily:    {"int", "long", "string"},
	types.FloatFamily:  {"float", "double", "string"},
	types.StringFamily: {"string", "bytes"},
	types.BytesFamily:  {"bytes", "string"},

	types.ArrayFamily: {"array", "string"},

	types.DateFamily:      {"string", "int.date"},
	types.TimeFamily:      {"string", "long.time-micros", "int.time-millis"},
	types.TimestampFamily: {"string", "long.timestamp-micros", "long.timestamp-millis"},

	types.TimeTZFamily:      {"string"},
	types.TimestampTZFamily: {"string"},

	types.IntervalFamily: {"string"},

	types.UuidFamily:           {"string"},
	types.CollatedStringFamily: {"string"},
	types.INetFamily:           {"string"},
	types.JsonFamily:           {"string"},
	types.BitFamily:            {"string"},
	types.DecimalFamily:        {"string"},
	types.EnumFamily:           {"string"},
}

type avroConsumer struct {
	importCtx      *parallelImportContext
	fieldNameToIdx map[string]int
	strict         bool
}

func (a *avroConsumer) convertNative(x interface{}, conv *row.DatumRowConverter) error {
	__antithesis_instrumentation__.Notify(494911)
	record, ok := x.(map[string]interface{})
	if !ok {
		__antithesis_instrumentation__.Notify(494914)
		return fmt.Errorf("unexpected native type; expected map[string]interface{} found %T instead", x)
	} else {
		__antithesis_instrumentation__.Notify(494915)
	}
	__antithesis_instrumentation__.Notify(494912)

	for f, v := range record {
		__antithesis_instrumentation__.Notify(494916)
		field := lexbase.NormalizeName(f)
		idx, ok := a.fieldNameToIdx[field]
		if !ok {
			__antithesis_instrumentation__.Notify(494920)
			if a.strict {
				__antithesis_instrumentation__.Notify(494922)
				return fmt.Errorf("could not find column for record field %s", field)
			} else {
				__antithesis_instrumentation__.Notify(494923)
			}
			__antithesis_instrumentation__.Notify(494921)
			continue
		} else {
			__antithesis_instrumentation__.Notify(494924)
		}
		__antithesis_instrumentation__.Notify(494917)

		typ := conv.VisibleColTypes[idx]
		avroT, ok := familyToAvroT[typ.Family()]
		if !ok {
			__antithesis_instrumentation__.Notify(494925)
			return fmt.Errorf("cannot convert avro value %v to col %s", v, conv.VisibleCols[idx].GetType().Name())
		} else {
			__antithesis_instrumentation__.Notify(494926)
		}
		__antithesis_instrumentation__.Notify(494918)
		datum, err := nativeToDatum(v, typ, avroT, conv.EvalCtx)
		if err != nil {
			__antithesis_instrumentation__.Notify(494927)
			return err
		} else {
			__antithesis_instrumentation__.Notify(494928)
		}
		__antithesis_instrumentation__.Notify(494919)
		conv.Datums[idx] = datum
	}
	__antithesis_instrumentation__.Notify(494913)
	return nil
}

func (a *avroConsumer) FillDatums(
	native interface{}, rowIndex int64, conv *row.DatumRowConverter,
) error {
	__antithesis_instrumentation__.Notify(494929)
	if err := a.convertNative(native, conv); err != nil {
		__antithesis_instrumentation__.Notify(494932)
		return err
	} else {
		__antithesis_instrumentation__.Notify(494933)
	}
	__antithesis_instrumentation__.Notify(494930)

	for i := range conv.Datums {
		__antithesis_instrumentation__.Notify(494934)
		if conv.TargetColOrds.Contains(i) && func() bool {
			__antithesis_instrumentation__.Notify(494935)
			return conv.Datums[i] == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(494936)
			if a.strict {
				__antithesis_instrumentation__.Notify(494938)
				return fmt.Errorf("field %s was not set in the avro import", conv.VisibleCols[i].GetName())
			} else {
				__antithesis_instrumentation__.Notify(494939)
			}
			__antithesis_instrumentation__.Notify(494937)
			conv.Datums[i] = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(494940)
		}
	}
	__antithesis_instrumentation__.Notify(494931)
	return nil
}

var _ importRowConsumer = &avroConsumer{}

type ocfStream struct {
	ocf      *goavro.OCFReader
	progress func() float32
	err      error
}

var _ importRowProducer = &ocfStream{}

func (o *ocfStream) Progress() float32 {
	__antithesis_instrumentation__.Notify(494941)
	if o.progress != nil {
		__antithesis_instrumentation__.Notify(494943)
		return o.progress()
	} else {
		__antithesis_instrumentation__.Notify(494944)
	}
	__antithesis_instrumentation__.Notify(494942)
	return 0
}

func (o *ocfStream) Scan() bool {
	__antithesis_instrumentation__.Notify(494945)
	return o.ocf.Scan()
}

func (o *ocfStream) Err() error {
	__antithesis_instrumentation__.Notify(494946)
	return o.err
}

func (o *ocfStream) Row() (interface{}, error) {
	__antithesis_instrumentation__.Notify(494947)
	return o.ocf.Read()
}

func (o *ocfStream) Skip() error {
	__antithesis_instrumentation__.Notify(494948)
	_, o.err = o.ocf.Read()
	return o.err
}

type avroRecordStream struct {
	importCtx  *parallelImportContext
	opts       *roachpb.AvroOptions
	input      *fileReader
	codec      *goavro.Codec
	row        interface{}
	buf        []byte
	eof        bool
	err        error
	trimLeft   bool
	maxBufSize int
	minBufSize int
	readSize   int
}

var _ importRowProducer = &avroRecordStream{}

func (r *avroRecordStream) Progress() float32 {
	__antithesis_instrumentation__.Notify(494949)
	return r.input.ReadFraction()
}

func (r *avroRecordStream) trimRecordSeparator() bool {
	__antithesis_instrumentation__.Notify(494950)
	if r.opts.RecordSeparator == 0 {
		__antithesis_instrumentation__.Notify(494953)
		return true
	} else {
		__antithesis_instrumentation__.Notify(494954)
	}
	__antithesis_instrumentation__.Notify(494951)

	if len(r.buf) > 0 {
		__antithesis_instrumentation__.Notify(494955)
		c, n := utf8.DecodeRune(r.buf)
		if n > 0 && func() bool {
			__antithesis_instrumentation__.Notify(494956)
			return c == r.opts.RecordSeparator == true
		}() == true {
			__antithesis_instrumentation__.Notify(494957)
			r.buf = r.buf[n:]
			return true
		} else {
			__antithesis_instrumentation__.Notify(494958)
		}
	} else {
		__antithesis_instrumentation__.Notify(494959)
	}
	__antithesis_instrumentation__.Notify(494952)
	return false
}

func (r *avroRecordStream) fill(sz int) {
	__antithesis_instrumentation__.Notify(494960)
	if r.eof || func() bool {
		__antithesis_instrumentation__.Notify(494962)
		return r.err != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(494963)
		return
	} else {
		__antithesis_instrumentation__.Notify(494964)
	}
	__antithesis_instrumentation__.Notify(494961)

	sink := bytes.NewBuffer(r.buf)
	_, r.err = io.CopyN(sink, r.input, int64(sz))
	r.buf = sink.Bytes()

	if r.err == io.EOF {
		__antithesis_instrumentation__.Notify(494965)
		r.eof = true
		r.err = nil
	} else {
		__antithesis_instrumentation__.Notify(494966)
	}
}

func (r *avroRecordStream) Scan() bool {
	__antithesis_instrumentation__.Notify(494967)
	if r.row != nil {
		__antithesis_instrumentation__.Notify(494969)
		panic("must call Row() or Skip() before calling Scan()")
	} else {
		__antithesis_instrumentation__.Notify(494970)
	}
	__antithesis_instrumentation__.Notify(494968)

	r.readNative()
	return r.err == nil && func() bool {
		__antithesis_instrumentation__.Notify(494971)
		return (!r.eof || func() bool {
			__antithesis_instrumentation__.Notify(494972)
			return r.row != nil == true
		}() == true) == true
	}() == true
}

func (r *avroRecordStream) Err() error {
	__antithesis_instrumentation__.Notify(494973)
	return r.err
}

func (r *avroRecordStream) decode() (interface{}, []byte, error) {
	__antithesis_instrumentation__.Notify(494974)
	if r.opts.Format == roachpb.AvroOptions_BIN_RECORDS {
		__antithesis_instrumentation__.Notify(494976)
		return r.codec.NativeFromBinary(r.buf)
	} else {
		__antithesis_instrumentation__.Notify(494977)
	}
	__antithesis_instrumentation__.Notify(494975)
	return r.codec.NativeFromTextual(r.buf)
}

func (r *avroRecordStream) readNative() {
	__antithesis_instrumentation__.Notify(494978)
	var remaining []byte
	var decodeErr error
	r.row = nil

	canReadMoreData := func() bool {
		__antithesis_instrumentation__.Notify(494982)
		return !r.eof && func() bool {
			__antithesis_instrumentation__.Notify(494983)
			return len(r.buf) < r.maxBufSize == true
		}() == true
	}
	__antithesis_instrumentation__.Notify(494979)

	for sz := r.readSize; r.row == nil && func() bool {
		__antithesis_instrumentation__.Notify(494984)
		return (len(r.buf) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(494985)
			return canReadMoreData() == true
		}() == true) == true
	}() == true; sz *= 2 {
		__antithesis_instrumentation__.Notify(494986)
		r.fill(sz)

		if r.trimLeft {
			__antithesis_instrumentation__.Notify(494989)
			r.trimLeft = !r.trimRecordSeparator()
		} else {
			__antithesis_instrumentation__.Notify(494990)
		}
		__antithesis_instrumentation__.Notify(494987)

		if len(r.buf) > 0 {
			__antithesis_instrumentation__.Notify(494991)
			r.row, remaining, decodeErr = r.decode()
		} else {
			__antithesis_instrumentation__.Notify(494992)
		}
		__antithesis_instrumentation__.Notify(494988)

		if decodeErr != nil && func() bool {
			__antithesis_instrumentation__.Notify(494993)
			return (r.eof || func() bool {
				__antithesis_instrumentation__.Notify(494994)
				return len(r.buf) > r.maxBufSize == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(494995)
			break
		} else {
			__antithesis_instrumentation__.Notify(494996)
		}
	}
	__antithesis_instrumentation__.Notify(494980)

	if decodeErr != nil {
		__antithesis_instrumentation__.Notify(494997)
		r.err = decodeErr
		return
	} else {
		__antithesis_instrumentation__.Notify(494998)
	}
	__antithesis_instrumentation__.Notify(494981)

	r.buf = remaining
	r.trimLeft = !r.trimRecordSeparator()
}

func (r *avroRecordStream) Skip() error {
	__antithesis_instrumentation__.Notify(494999)
	r.row = nil
	return nil
}

func (r *avroRecordStream) Row() (interface{}, error) {
	__antithesis_instrumentation__.Notify(495000)
	res := r.row
	r.row = nil
	return res, nil
}

func newImportAvroPipeline(
	avro *avroInputReader, input *fileReader,
) (importRowProducer, importRowConsumer, error) {
	__antithesis_instrumentation__.Notify(495001)
	fieldIdxByName := make(map[string]int)
	for idx, col := range avro.importContext.tableDesc.VisibleColumns() {
		__antithesis_instrumentation__.Notify(495006)
		fieldIdxByName[col.GetName()] = idx
	}
	__antithesis_instrumentation__.Notify(495002)

	consumer := &avroConsumer{
		importCtx:      avro.importContext,
		fieldNameToIdx: fieldIdxByName,
		strict:         avro.opts.StrictMode,
	}

	if avro.opts.Format == roachpb.AvroOptions_OCF {
		__antithesis_instrumentation__.Notify(495007)
		ocf, err := goavro.NewOCFReader(bufio.NewReaderSize(input, 64<<10))
		if err != nil {
			__antithesis_instrumentation__.Notify(495010)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(495011)
		}
		__antithesis_instrumentation__.Notify(495008)
		producer := &ocfStream{
			ocf:      ocf,
			progress: func() float32 { __antithesis_instrumentation__.Notify(495012); return input.ReadFraction() },
		}
		__antithesis_instrumentation__.Notify(495009)
		return producer, consumer, nil
	} else {
		__antithesis_instrumentation__.Notify(495013)
	}
	__antithesis_instrumentation__.Notify(495003)

	codec, err := goavro.NewCodec(avro.opts.SchemaJSON)
	if err != nil {
		__antithesis_instrumentation__.Notify(495014)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(495015)
	}
	__antithesis_instrumentation__.Notify(495004)

	producer := &avroRecordStream{
		importCtx: avro.importContext,
		opts:      &avro.opts,
		input:     input,
		codec:     codec,

		minBufSize: 512,
		maxBufSize: 4 << 20,
		readSize:   4 << 10,
	}

	if int(avro.opts.MaxRecordSize) > producer.maxBufSize {
		__antithesis_instrumentation__.Notify(495016)
		producer.maxBufSize = int(avro.opts.MaxRecordSize)
	} else {
		__antithesis_instrumentation__.Notify(495017)
	}
	__antithesis_instrumentation__.Notify(495005)

	return producer, consumer, nil
}

type avroInputReader struct {
	importContext *parallelImportContext
	opts          roachpb.AvroOptions
}

var _ inputConverter = &avroInputReader{}

func newAvroInputReader(
	semaCtx *tree.SemaContext,
	kvCh chan row.KVBatch,
	tableDesc catalog.TableDescriptor,
	avroOpts roachpb.AvroOptions,
	walltime int64,
	parallelism int,
	evalCtx *tree.EvalContext,
) (*avroInputReader, error) {
	__antithesis_instrumentation__.Notify(495018)

	return &avroInputReader{
		importContext: &parallelImportContext{
			semaCtx:    semaCtx,
			walltime:   walltime,
			numWorkers: parallelism,
			evalCtx:    evalCtx,
			tableDesc:  tableDesc,
			kvCh:       kvCh,
		},
		opts: avroOpts,
	}, nil
}

func (a *avroInputReader) start(group ctxgroup.Group) { __antithesis_instrumentation__.Notify(495019) }

func (a *avroInputReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495020)
	return readInputFiles(ctx, dataFiles, resumePos, format, a.readFile, makeExternalStorage, user)
}

func (a *avroInputReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(495021)
	producer, consumer, err := newImportAvroPipeline(a, input)
	if err != nil {
		__antithesis_instrumentation__.Notify(495023)
		return err
	} else {
		__antithesis_instrumentation__.Notify(495024)
	}
	__antithesis_instrumentation__.Notify(495022)

	fileCtx := &importFileContext{
		source:   inputIdx,
		skip:     resumePos,
		rejected: rejected,
		rowLimit: a.opts.RowLimit,
	}
	return runParallelImport(ctx, a.importContext, fileCtx, producer, consumer)
}
