package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/errors"
)

type mysqloutfileReader struct {
	importCtx *parallelImportContext
	opts      roachpb.MySQLOutfileOptions
}

var _ inputConverter = &mysqloutfileReader{}

func newMysqloutfileReader(
	semaCtx *tree.SemaContext,
	opts roachpb.MySQLOutfileOptions,
	kvCh chan row.KVBatch,
	walltime int64,
	parallelism int,
	tableDesc catalog.TableDescriptor,
	targetCols tree.NameList,
	evalCtx *tree.EvalContext,
) (*mysqloutfileReader, error) {
	__antithesis_instrumentation__.Notify(495638)
	return &mysqloutfileReader{
		importCtx: &parallelImportContext{
			semaCtx:    semaCtx,
			walltime:   walltime,
			numWorkers: parallelism,
			evalCtx:    evalCtx,
			tableDesc:  tableDesc,
			targetCols: targetCols,
			kvCh:       kvCh,
		},
		opts: opts,
	}, nil
}

func (d *mysqloutfileReader) start(ctx ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(495639)
}

func (d *mysqloutfileReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495640)
	return readInputFiles(ctx, dataFiles, resumePos, format, d.readFile, makeExternalStorage, user)
}

type delimitedProducer struct {
	importCtx *parallelImportContext
	opts      *roachpb.MySQLOutfileOptions
	input     *fileReader
	reader    *bufio.Reader
	row       []rune
	err       error
	eof       bool
}

var _ importRowProducer = &delimitedProducer{}

func (d *delimitedProducer) Scan() bool {
	__antithesis_instrumentation__.Notify(495641)
	d.row = nil
	var r rune
	var w int
	nextLiteral := false
	fieldEnclosed := false

	for {
		__antithesis_instrumentation__.Notify(495642)
		r, w, d.err = d.reader.ReadRune()
		if d.err == io.EOF {
			__antithesis_instrumentation__.Notify(495649)
			d.eof = true
			d.err = nil
		} else {
			__antithesis_instrumentation__.Notify(495650)
		}
		__antithesis_instrumentation__.Notify(495643)

		if d.eof {
			__antithesis_instrumentation__.Notify(495651)
			if d.row != nil {
				__antithesis_instrumentation__.Notify(495654)
				return true
			} else {
				__antithesis_instrumentation__.Notify(495655)
			}
			__antithesis_instrumentation__.Notify(495652)
			if nextLiteral {
				__antithesis_instrumentation__.Notify(495656)
				d.err = io.ErrUnexpectedEOF
			} else {
				__antithesis_instrumentation__.Notify(495657)
			}
			__antithesis_instrumentation__.Notify(495653)
			return false
		} else {
			__antithesis_instrumentation__.Notify(495658)
		}
		__antithesis_instrumentation__.Notify(495644)

		if d.err != nil {
			__antithesis_instrumentation__.Notify(495659)
			return false
		} else {
			__antithesis_instrumentation__.Notify(495660)
		}
		__antithesis_instrumentation__.Notify(495645)

		if r == unicode.ReplacementChar && func() bool {
			__antithesis_instrumentation__.Notify(495661)
			return w == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(495662)
			if d.err = d.reader.UnreadRune(); d.err != nil {
				__antithesis_instrumentation__.Notify(495665)
				return false
			} else {
				__antithesis_instrumentation__.Notify(495666)
			}
			__antithesis_instrumentation__.Notify(495663)
			var raw byte
			raw, d.err = d.reader.ReadByte()
			if d.err != nil {
				__antithesis_instrumentation__.Notify(495667)
				return false
			} else {
				__antithesis_instrumentation__.Notify(495668)
			}
			__antithesis_instrumentation__.Notify(495664)
			r = rune(raw)
		} else {
			__antithesis_instrumentation__.Notify(495669)
		}
		__antithesis_instrumentation__.Notify(495646)

		if r == d.opts.RowSeparator && func() bool {
			__antithesis_instrumentation__.Notify(495670)
			return !nextLiteral == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(495671)
			return !fieldEnclosed == true
		}() == true {
			__antithesis_instrumentation__.Notify(495672)
			return true
		} else {
			__antithesis_instrumentation__.Notify(495673)
		}
		__antithesis_instrumentation__.Notify(495647)

		d.row = append(d.row, r)

		if d.opts.HasEscape {
			__antithesis_instrumentation__.Notify(495674)
			nextLiteral = !nextLiteral && func() bool {
				__antithesis_instrumentation__.Notify(495675)
				return r == d.opts.Escape == true
			}() == true
		} else {
			__antithesis_instrumentation__.Notify(495676)
		}
		__antithesis_instrumentation__.Notify(495648)

		if d.opts.Enclose != roachpb.MySQLOutfileOptions_Never && func() bool {
			__antithesis_instrumentation__.Notify(495677)
			return r == d.opts.Encloser == true
		}() == true {
			__antithesis_instrumentation__.Notify(495678)

			fieldEnclosed = len(d.row) == 1
		} else {
			__antithesis_instrumentation__.Notify(495679)
		}
	}
}

func (d *delimitedProducer) Err() error {
	__antithesis_instrumentation__.Notify(495680)
	return d.err
}

func (d *delimitedProducer) Skip() error {
	__antithesis_instrumentation__.Notify(495681)
	return nil
}

func (d *delimitedProducer) Row() (interface{}, error) {
	__antithesis_instrumentation__.Notify(495682)
	return d.row, d.err
}

func (d *delimitedProducer) Progress() float32 {
	__antithesis_instrumentation__.Notify(495683)
	return d.input.ReadFraction()
}

type delimitedConsumer struct {
	opts *roachpb.MySQLOutfileOptions
}

var _ importRowConsumer = &delimitedConsumer{}

func (d *delimitedConsumer) FillDatums(
	input interface{}, rowNum int64, conv *row.DatumRowConverter,
) error {
	__antithesis_instrumentation__.Notify(495684)
	data := input.([]rune)

	var fieldParts []rune

	var nextLiteral bool

	var readingField bool

	var gotEncloser bool

	var gotNull bool

	var datumIdx int

	addField := func() error {
		__antithesis_instrumentation__.Notify(495689)
		defer func() {
			__antithesis_instrumentation__.Notify(495695)
			fieldParts = fieldParts[:0]
			readingField = false
			gotEncloser = false
		}()
		__antithesis_instrumentation__.Notify(495690)
		if nextLiteral {
			__antithesis_instrumentation__.Notify(495696)
			return newImportRowError(errors.New("unmatched literal"), string(data), rowNum)
		} else {
			__antithesis_instrumentation__.Notify(495697)
		}
		__antithesis_instrumentation__.Notify(495691)

		var datum tree.Datum

		if gotEncloser {
			__antithesis_instrumentation__.Notify(495698)

			if readingField {
				__antithesis_instrumentation__.Notify(495699)
				fieldParts = fieldParts[:len(fieldParts)-1]
			} else {
				__antithesis_instrumentation__.Notify(495700)

				gotEncloser = false
				return newImportRowError(errors.New("unmatched field enclosure at end of field"),
					string(data), rowNum)
			}
		} else {
			__antithesis_instrumentation__.Notify(495701)
			if readingField {
				__antithesis_instrumentation__.Notify(495702)
				return newImportRowError(errors.New("unmatched field enclosure at start of field"),
					string(data), rowNum)
			} else {
				__antithesis_instrumentation__.Notify(495703)
			}
		}
		__antithesis_instrumentation__.Notify(495692)
		field := string(fieldParts)
		if datumIdx >= len(conv.VisibleCols) {
			__antithesis_instrumentation__.Notify(495704)
			return newImportRowError(
				fmt.Errorf("too many columns, got %d expected %d", datumIdx+1, len(conv.VisibleCols)),
				string(data), rowNum)
		} else {
			__antithesis_instrumentation__.Notify(495705)
		}
		__antithesis_instrumentation__.Notify(495693)

		if gotNull {
			__antithesis_instrumentation__.Notify(495706)
			gotNull = false
			if len(field) != 0 {
				__antithesis_instrumentation__.Notify(495708)
				return newImportRowError(fmt.Errorf("unexpected data after null encoding: %q", field),
					string(data), rowNum)
			} else {
				__antithesis_instrumentation__.Notify(495709)
			}
			__antithesis_instrumentation__.Notify(495707)
			datum = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(495710)
			if (!d.opts.HasEscape && func() bool {
				__antithesis_instrumentation__.Notify(495711)
				return field == "NULL" == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(495712)
				return (d.opts.NullEncoding != nil && func() bool {
					__antithesis_instrumentation__.Notify(495713)
					return field == *d.opts.NullEncoding == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495714)
				datum = tree.DNull
			} else {
				__antithesis_instrumentation__.Notify(495715)

				var err error
				datum, err = rowenc.ParseDatumStringAsWithRawBytes(conv.VisibleColTypes[datumIdx], field, conv.EvalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(495716)
					col := conv.VisibleCols[datumIdx]
					return newImportRowError(
						errors.Wrapf(err, "error while parse %q as %s", col.GetName(), col.GetType().SQLString()),
						string(data), rowNum)
				} else {
					__antithesis_instrumentation__.Notify(495717)
				}
			}
		}
		__antithesis_instrumentation__.Notify(495694)
		conv.Datums[datumIdx] = datum
		datumIdx++
		return nil
	}
	__antithesis_instrumentation__.Notify(495685)

	for _, c := range data {
		__antithesis_instrumentation__.Notify(495718)

		if d.opts.HasEscape {
			__antithesis_instrumentation__.Notify(495723)
			if nextLiteral {
				__antithesis_instrumentation__.Notify(495725)
				nextLiteral = false

				switch c {
				case '0':
					__antithesis_instrumentation__.Notify(495727)
					fieldParts = append(fieldParts, rune(0))
				case 'b':
					__antithesis_instrumentation__.Notify(495728)
					fieldParts = append(fieldParts, rune('\b'))
				case 'n':
					__antithesis_instrumentation__.Notify(495729)
					fieldParts = append(fieldParts, rune('\n'))
				case 'r':
					__antithesis_instrumentation__.Notify(495730)
					fieldParts = append(fieldParts, rune('\r'))
				case 't':
					__antithesis_instrumentation__.Notify(495731)
					fieldParts = append(fieldParts, rune('\t'))
				case 'Z':
					__antithesis_instrumentation__.Notify(495732)
					fieldParts = append(fieldParts, rune(byte(26)))
				case 'N':
					__antithesis_instrumentation__.Notify(495733)
					if gotNull {
						__antithesis_instrumentation__.Notify(495736)
						return newImportRowError(errors.New("unexpected null encoding"), string(data), rowNum)
					} else {
						__antithesis_instrumentation__.Notify(495737)
					}
					__antithesis_instrumentation__.Notify(495734)
					gotNull = true
				default:
					__antithesis_instrumentation__.Notify(495735)
					fieldParts = append(fieldParts, c)
				}
				__antithesis_instrumentation__.Notify(495726)
				gotEncloser = false
				continue
			} else {
				__antithesis_instrumentation__.Notify(495738)
			}
			__antithesis_instrumentation__.Notify(495724)

			if c == d.opts.Escape {
				__antithesis_instrumentation__.Notify(495739)
				nextLiteral = true
				gotEncloser = false
				continue
			} else {
				__antithesis_instrumentation__.Notify(495740)
			}
		} else {
			__antithesis_instrumentation__.Notify(495741)
		}
		__antithesis_instrumentation__.Notify(495719)

		if (!readingField || func() bool {
			__antithesis_instrumentation__.Notify(495742)
			return gotEncloser == true
		}() == true) && func() bool {
			__antithesis_instrumentation__.Notify(495743)
			return c == d.opts.FieldSeparator == true
		}() == true {
			__antithesis_instrumentation__.Notify(495744)
			if err := addField(); err != nil {
				__antithesis_instrumentation__.Notify(495746)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495747)
			}
			__antithesis_instrumentation__.Notify(495745)
			continue
		} else {
			__antithesis_instrumentation__.Notify(495748)
		}
		__antithesis_instrumentation__.Notify(495720)

		if gotEncloser {
			__antithesis_instrumentation__.Notify(495749)
			gotEncloser = false
		} else {
			__antithesis_instrumentation__.Notify(495750)
		}
		__antithesis_instrumentation__.Notify(495721)

		if d.opts.Enclose != roachpb.MySQLOutfileOptions_Never && func() bool {
			__antithesis_instrumentation__.Notify(495751)
			return c == d.opts.Encloser == true
		}() == true {
			__antithesis_instrumentation__.Notify(495752)
			if !readingField && func() bool {
				__antithesis_instrumentation__.Notify(495754)
				return len(fieldParts) == 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(495755)
				readingField = true
				continue
			} else {
				__antithesis_instrumentation__.Notify(495756)
			}
			__antithesis_instrumentation__.Notify(495753)
			gotEncloser = true
		} else {
			__antithesis_instrumentation__.Notify(495757)
		}
		__antithesis_instrumentation__.Notify(495722)
		fieldParts = append(fieldParts, c)
	}
	__antithesis_instrumentation__.Notify(495686)

	if err := addField(); err != nil {
		__antithesis_instrumentation__.Notify(495758)
		return err
	} else {
		__antithesis_instrumentation__.Notify(495759)
	}
	__antithesis_instrumentation__.Notify(495687)

	if datumIdx != len(conv.VisibleCols) {
		__antithesis_instrumentation__.Notify(495760)
		return newImportRowError(fmt.Errorf(
			"unexpected number of columns, expected %d got %d", len(conv.VisibleCols), datumIdx),
			string(data), rowNum)
	} else {
		__antithesis_instrumentation__.Notify(495761)
	}
	__antithesis_instrumentation__.Notify(495688)

	return nil
}

func (d *mysqloutfileReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(495762)
	producer := &delimitedProducer{
		importCtx: d.importCtx,
		opts:      &d.opts,
		input:     input,
		reader:    bufio.NewReaderSize(input, 64*1024),
	}
	consumer := &delimitedConsumer{opts: &d.opts}

	if resumePos < int64(d.opts.Skip) {
		__antithesis_instrumentation__.Notify(495764)
		resumePos = int64(d.opts.Skip)
	} else {
		__antithesis_instrumentation__.Notify(495765)
	}
	__antithesis_instrumentation__.Notify(495763)

	fileCtx := &importFileContext{
		source:   inputIdx,
		skip:     resumePos,
		rejected: rejected,
		rowLimit: d.opts.RowLimit,
	}

	return runParallelImport(ctx, d.importCtx, fileCtx, producer, consumer)
}
