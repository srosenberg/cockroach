package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/errors"
)

type csvInputReader struct {
	importCtx *parallelImportContext

	numExpectedDataCols int
	opts                roachpb.CSVOptions
}

var _ inputConverter = &csvInputReader{}

func newCSVInputReader(
	semaCtx *tree.SemaContext,
	kvCh chan row.KVBatch,
	opts roachpb.CSVOptions,
	walltime int64,
	parallelism int,
	tableDesc catalog.TableDescriptor,
	targetCols tree.NameList,
	evalCtx *tree.EvalContext,
	seqChunkProvider *row.SeqChunkProvider,
) *csvInputReader {
	__antithesis_instrumentation__.Notify(495273)
	numExpectedDataCols := len(targetCols)
	if numExpectedDataCols == 0 {
		__antithesis_instrumentation__.Notify(495275)
		numExpectedDataCols = len(tableDesc.VisibleColumns())
	} else {
		__antithesis_instrumentation__.Notify(495276)
	}
	__antithesis_instrumentation__.Notify(495274)

	return &csvInputReader{
		importCtx: &parallelImportContext{
			semaCtx:          semaCtx,
			walltime:         walltime,
			numWorkers:       parallelism,
			evalCtx:          evalCtx,
			tableDesc:        tableDesc,
			targetCols:       targetCols,
			kvCh:             kvCh,
			seqChunkProvider: seqChunkProvider,
		},
		numExpectedDataCols: numExpectedDataCols,
		opts:                opts,
	}
}

func (c *csvInputReader) start(group ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(495277)
}

func (c *csvInputReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495278)
	return readInputFiles(ctx, dataFiles, resumePos, format, c.readFile, makeExternalStorage, user)
}

func (c *csvInputReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(495279)
	producer, consumer := newCSVPipeline(c, input)

	if resumePos < int64(c.opts.Skip) {
		__antithesis_instrumentation__.Notify(495281)
		resumePos = int64(c.opts.Skip)
	} else {
		__antithesis_instrumentation__.Notify(495282)
	}
	__antithesis_instrumentation__.Notify(495280)

	fileCtx := &importFileContext{
		source:   inputIdx,
		skip:     resumePos,
		rejected: rejected,
		rowLimit: c.opts.RowLimit,
	}

	return runParallelImport(ctx, c.importCtx, fileCtx, producer, consumer)
}

type csvRowProducer struct {
	importCtx          *parallelImportContext
	opts               *roachpb.CSVOptions
	csv                *csv.Reader
	rowNum             int64
	err                error
	record             []string
	progress           func() float32
	numExpectedColumns int
}

var _ importRowProducer = &csvRowProducer{}

func (p *csvRowProducer) Scan() bool {
	__antithesis_instrumentation__.Notify(495283)
	p.record, p.err = p.csv.Read()

	if p.err == io.EOF {
		__antithesis_instrumentation__.Notify(495285)
		p.err = nil
		return false
	} else {
		__antithesis_instrumentation__.Notify(495286)
	}
	__antithesis_instrumentation__.Notify(495284)

	return p.err == nil
}

func (p *csvRowProducer) Err() error {
	__antithesis_instrumentation__.Notify(495287)
	return p.err
}

func (p *csvRowProducer) Skip() error {
	__antithesis_instrumentation__.Notify(495288)

	return nil
}

func strRecord(record []string, sep rune) string {
	__antithesis_instrumentation__.Notify(495289)
	csvSep := ","
	if sep != 0 {
		__antithesis_instrumentation__.Notify(495291)
		csvSep = string(sep)
	} else {
		__antithesis_instrumentation__.Notify(495292)
	}
	__antithesis_instrumentation__.Notify(495290)
	return strings.Join(record, csvSep)
}

func (p *csvRowProducer) Row() (interface{}, error) {
	__antithesis_instrumentation__.Notify(495293)
	p.rowNum++
	expectedColsLen := p.numExpectedColumns

	if len(p.record) == expectedColsLen {
		__antithesis_instrumentation__.Notify(495295)

	} else {
		__antithesis_instrumentation__.Notify(495296)
		if len(p.record) == expectedColsLen+1 && func() bool {
			__antithesis_instrumentation__.Notify(495297)
			return p.record[expectedColsLen] == "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(495298)

			p.record = p.record[:expectedColsLen]
		} else {
			__antithesis_instrumentation__.Notify(495299)
			return nil, newImportRowError(
				errors.Errorf("expected %d fields, got %d", expectedColsLen, len(p.record)),
				strRecord(p.record, p.opts.Comma),
				p.rowNum)
		}
	}
	__antithesis_instrumentation__.Notify(495294)
	return p.record, nil
}

func (p *csvRowProducer) Progress() float32 {
	__antithesis_instrumentation__.Notify(495300)
	return p.progress()
}

type csvRowConsumer struct {
	importCtx *parallelImportContext
	opts      *roachpb.CSVOptions
}

var _ importRowConsumer = &csvRowConsumer{}

func (c *csvRowConsumer) FillDatums(
	row interface{}, rowNum int64, conv *row.DatumRowConverter,
) error {
	__antithesis_instrumentation__.Notify(495301)
	record := row.([]string)
	datumIdx := 0

	for i, field := range record {
		__antithesis_instrumentation__.Notify(495303)

		if !conv.TargetColOrds.Contains(i) {
			__antithesis_instrumentation__.Notify(495306)
			continue
		} else {
			__antithesis_instrumentation__.Notify(495307)
		}
		__antithesis_instrumentation__.Notify(495304)

		if c.opts.NullEncoding != nil && func() bool {
			__antithesis_instrumentation__.Notify(495308)
			return field == *c.opts.NullEncoding == true
		}() == true {
			__antithesis_instrumentation__.Notify(495309)
			conv.Datums[datumIdx] = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(495310)
			var err error
			conv.Datums[datumIdx], err = rowenc.ParseDatumStringAs(conv.VisibleColTypes[i], field, conv.EvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(495311)
				col := conv.VisibleCols[i]
				return newImportRowError(
					errors.Wrapf(err, "parse %q as %s", col.GetName(), col.GetType().SQLString()),
					strRecord(record, c.opts.Comma),
					rowNum)
			} else {
				__antithesis_instrumentation__.Notify(495312)
			}
		}
		__antithesis_instrumentation__.Notify(495305)
		datumIdx++
	}
	__antithesis_instrumentation__.Notify(495302)
	return nil
}

func newCSVPipeline(c *csvInputReader, input *fileReader) (*csvRowProducer, *csvRowConsumer) {
	__antithesis_instrumentation__.Notify(495313)
	cr := csv.NewReader(input)
	if c.opts.Comma != 0 {
		__antithesis_instrumentation__.Notify(495316)
		cr.Comma = c.opts.Comma
	} else {
		__antithesis_instrumentation__.Notify(495317)
	}
	__antithesis_instrumentation__.Notify(495314)
	cr.FieldsPerRecord = -1
	cr.LazyQuotes = !c.opts.StrictQuotes
	cr.Comment = c.opts.Comment

	producer := &csvRowProducer{
		importCtx:          c.importCtx,
		opts:               &c.opts,
		csv:                cr,
		progress:           func() float32 { __antithesis_instrumentation__.Notify(495318); return input.ReadFraction() },
		numExpectedColumns: c.numExpectedDataCols,
	}
	__antithesis_instrumentation__.Notify(495315)
	consumer := &csvRowConsumer{
		importCtx: c.importCtx,
		opts:      &c.opts,
	}

	return producer, consumer
}
