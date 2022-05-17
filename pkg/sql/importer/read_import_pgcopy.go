package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
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

const defaultScanBuffer = 1024 * 1024 * 4

type pgCopyReader struct {
	importCtx *parallelImportContext
	opts      roachpb.PgCopyOptions
}

var _ inputConverter = &pgCopyReader{}

func newPgCopyReader(
	semaCtx *tree.SemaContext,
	opts roachpb.PgCopyOptions,
	kvCh chan row.KVBatch,
	walltime int64,
	parallelism int,
	tableDesc catalog.TableDescriptor,
	targetCols tree.NameList,
	evalCtx *tree.EvalContext,
) (*pgCopyReader, error) {
	__antithesis_instrumentation__.Notify(495766)
	return &pgCopyReader{
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

func (d *pgCopyReader) start(ctx ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(495767)
}

func (d *pgCopyReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495768)
	return readInputFiles(ctx, dataFiles, resumePos, format, d.readFile, makeExternalStorage, user)
}

type postgreStreamCopy struct {
	s         *bufio.Scanner
	delimiter rune
	null      string
}

func newPostgreStreamCopy(s *bufio.Scanner, delimiter rune, null string) *postgreStreamCopy {
	__antithesis_instrumentation__.Notify(495769)
	return &postgreStreamCopy{
		s:         s,
		delimiter: delimiter,
		null:      null,
	}
}

var errCopyDone = errors.New("COPY done")

func (p *postgreStreamCopy) Next() (copyData, error) {
	__antithesis_instrumentation__.Notify(495770)
	var row copyData

	var field []byte

	addField := func() error {
		__antithesis_instrumentation__.Notify(495778)

		if string(field) == p.null {
			__antithesis_instrumentation__.Notify(495780)
			row = append(row, nil)
		} else {
			__antithesis_instrumentation__.Notify(495781)

			nf := field[:0]
			for i := 0; i < len(field); i++ {
				__antithesis_instrumentation__.Notify(495783)
				if field[i] == '\\' {
					__antithesis_instrumentation__.Notify(495784)
					i++
					if len(field) <= i {
						__antithesis_instrumentation__.Notify(495786)
						return errors.New("unmatched escape")
					} else {
						__antithesis_instrumentation__.Notify(495787)
					}
					__antithesis_instrumentation__.Notify(495785)

					c := field[i]
					switch c {
					case 'b':
						__antithesis_instrumentation__.Notify(495788)
						nf = append(nf, '\b')
					case 'f':
						__antithesis_instrumentation__.Notify(495789)
						nf = append(nf, '\f')
					case 'n':
						__antithesis_instrumentation__.Notify(495790)
						nf = append(nf, '\n')
					case 'r':
						__antithesis_instrumentation__.Notify(495791)
						nf = append(nf, '\r')
					case 't':
						__antithesis_instrumentation__.Notify(495792)
						nf = append(nf, '\t')
					case 'v':
						__antithesis_instrumentation__.Notify(495793)
						nf = append(nf, '\v')
					default:
						__antithesis_instrumentation__.Notify(495794)
						if c == 'x' || func() bool {
							__antithesis_instrumentation__.Notify(495795)
							return (c >= '0' && func() bool {
								__antithesis_instrumentation__.Notify(495796)
								return c <= '9' == true
							}() == true) == true
						}() == true {
							__antithesis_instrumentation__.Notify(495797)

							if len(field) <= i+2 {
								__antithesis_instrumentation__.Notify(495801)
								return errors.Errorf("unsupported escape sequence: \\%s", field[i:])
							} else {
								__antithesis_instrumentation__.Notify(495802)
							}
							__antithesis_instrumentation__.Notify(495798)
							base := 8
							idx := 0
							if c == 'x' {
								__antithesis_instrumentation__.Notify(495803)
								base = 16
								idx = 1
							} else {
								__antithesis_instrumentation__.Notify(495804)
							}
							__antithesis_instrumentation__.Notify(495799)
							v, err := strconv.ParseInt(string(field[i+idx:i+3]), base, 8)
							if err != nil {
								__antithesis_instrumentation__.Notify(495805)
								return err
							} else {
								__antithesis_instrumentation__.Notify(495806)
							}
							__antithesis_instrumentation__.Notify(495800)
							i += 2
							nf = append(nf, byte(v))
						} else {
							__antithesis_instrumentation__.Notify(495807)
							nf = append(nf, string(c)...)
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(495808)
					nf = append(nf, field[i])
				}
			}
			__antithesis_instrumentation__.Notify(495782)
			ns := string(nf)
			row = append(row, &ns)
		}
		__antithesis_instrumentation__.Notify(495779)
		field = field[:0]
		return nil
	}
	__antithesis_instrumentation__.Notify(495771)

	scanned := p.s.Scan()
	if err := p.s.Err(); err != nil {
		__antithesis_instrumentation__.Notify(495809)
		if errors.Is(err, bufio.ErrTooLong) {
			__antithesis_instrumentation__.Notify(495811)
			err = wrapWithLineTooLongHint(
				errors.New("line too long"),
			)
		} else {
			__antithesis_instrumentation__.Notify(495812)
		}
		__antithesis_instrumentation__.Notify(495810)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495813)
	}
	__antithesis_instrumentation__.Notify(495772)
	if !scanned {
		__antithesis_instrumentation__.Notify(495814)
		return nil, io.EOF
	} else {
		__antithesis_instrumentation__.Notify(495815)
	}
	__antithesis_instrumentation__.Notify(495773)

	if bytes.Equal(p.s.Bytes(), []byte(`\.`)) {
		__antithesis_instrumentation__.Notify(495816)
		return nil, errCopyDone
	} else {
		__antithesis_instrumentation__.Notify(495817)
	}
	__antithesis_instrumentation__.Notify(495774)
	reader := bytes.NewReader(p.s.Bytes())

	var sawBackslash bool

	for {
		__antithesis_instrumentation__.Notify(495818)
		c, w, err := reader.ReadRune()
		if err == io.EOF {
			__antithesis_instrumentation__.Notify(495823)
			break
		} else {
			__antithesis_instrumentation__.Notify(495824)
		}
		__antithesis_instrumentation__.Notify(495819)
		if err != nil {
			__antithesis_instrumentation__.Notify(495825)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495826)
		}
		__antithesis_instrumentation__.Notify(495820)
		if c == unicode.ReplacementChar && func() bool {
			__antithesis_instrumentation__.Notify(495827)
			return w == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(495828)
			return nil, errors.New("error decoding UTF-8 Rune")
		} else {
			__antithesis_instrumentation__.Notify(495829)
		}
		__antithesis_instrumentation__.Notify(495821)

		if sawBackslash {
			__antithesis_instrumentation__.Notify(495830)
			sawBackslash = false
			if c == p.delimiter {
				__antithesis_instrumentation__.Notify(495832)
				field = append(field, string(p.delimiter)...)
			} else {
				__antithesis_instrumentation__.Notify(495833)
				field = append(field, '\\')
				field = append(field, string(c)...)
			}
			__antithesis_instrumentation__.Notify(495831)
			continue
		} else {
			__antithesis_instrumentation__.Notify(495834)
			if c == '\\' {
				__antithesis_instrumentation__.Notify(495835)
				sawBackslash = true
				continue
			} else {
				__antithesis_instrumentation__.Notify(495836)
			}
		}
		__antithesis_instrumentation__.Notify(495822)

		const rowSeparator = '\n'

		if c == p.delimiter || func() bool {
			__antithesis_instrumentation__.Notify(495837)
			return c == rowSeparator == true
		}() == true {
			__antithesis_instrumentation__.Notify(495838)
			if err := addField(); err != nil {
				__antithesis_instrumentation__.Notify(495839)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(495840)
			}
		} else {
			__antithesis_instrumentation__.Notify(495841)
			field = append(field, string(c)...)
		}
	}
	__antithesis_instrumentation__.Notify(495775)
	if sawBackslash {
		__antithesis_instrumentation__.Notify(495842)
		return nil, errors.Errorf("unmatched escape")
	} else {
		__antithesis_instrumentation__.Notify(495843)
	}
	__antithesis_instrumentation__.Notify(495776)

	if err := addField(); err != nil {
		__antithesis_instrumentation__.Notify(495844)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495845)
	}
	__antithesis_instrumentation__.Notify(495777)
	return row, nil
}

const (
	copyDefaultDelimiter = '\t'
	copyDefaultNull      = `\N`
)

type copyData []*string

func (c copyData) String() string {
	__antithesis_instrumentation__.Notify(495846)
	var buf bytes.Buffer
	for i, s := range c {
		__antithesis_instrumentation__.Notify(495848)
		if i > 0 {
			__antithesis_instrumentation__.Notify(495850)
			buf.WriteByte(copyDefaultDelimiter)
		} else {
			__antithesis_instrumentation__.Notify(495851)
		}
		__antithesis_instrumentation__.Notify(495849)
		if s == nil {
			__antithesis_instrumentation__.Notify(495852)
			buf.WriteString(copyDefaultNull)
		} else {
			__antithesis_instrumentation__.Notify(495853)

			fmt.Fprintf(&buf, "%q", *s)
		}
	}
	__antithesis_instrumentation__.Notify(495847)
	return buf.String()
}

type pgCopyProducer struct {
	importCtx  *parallelImportContext
	opts       *roachpb.PgCopyOptions
	input      *fileReader
	copyStream *postgreStreamCopy
	row        copyData
	err        error
}

var _ importRowProducer = &pgCopyProducer{}

func (p *pgCopyProducer) Scan() bool {
	__antithesis_instrumentation__.Notify(495854)
	p.row, p.err = p.copyStream.Next()
	if p.err == io.EOF {
		__antithesis_instrumentation__.Notify(495856)
		p.err = nil
		return false
	} else {
		__antithesis_instrumentation__.Notify(495857)
	}
	__antithesis_instrumentation__.Notify(495855)

	return p.err == nil
}

func (p *pgCopyProducer) Err() error {
	__antithesis_instrumentation__.Notify(495858)
	return p.err
}

func (p *pgCopyProducer) Skip() error {
	__antithesis_instrumentation__.Notify(495859)
	return nil
}

func (p *pgCopyProducer) Row() (interface{}, error) {
	__antithesis_instrumentation__.Notify(495860)
	return p.row, p.err
}

func (p *pgCopyProducer) Progress() float32 {
	__antithesis_instrumentation__.Notify(495861)
	return p.input.ReadFraction()
}

type pgCopyConsumer struct {
	opts *roachpb.PgCopyOptions
}

var _ importRowConsumer = &pgCopyConsumer{}

func (p *pgCopyConsumer) FillDatums(
	row interface{}, rowNum int64, conv *row.DatumRowConverter,
) error {
	__antithesis_instrumentation__.Notify(495862)
	data := row.(copyData)
	var err error

	if len(data) != len(conv.VisibleColTypes) {
		__antithesis_instrumentation__.Notify(495865)
		return newImportRowError(fmt.Errorf(
			"unexpected number of columns, expected %d values, got %d",
			len(conv.VisibleColTypes), len(data)), data.String(), rowNum)
	} else {
		__antithesis_instrumentation__.Notify(495866)
	}
	__antithesis_instrumentation__.Notify(495863)

	for i, s := range data {
		__antithesis_instrumentation__.Notify(495867)
		if s == nil {
			__antithesis_instrumentation__.Notify(495868)
			conv.Datums[i] = tree.DNull
		} else {
			__antithesis_instrumentation__.Notify(495869)
			conv.Datums[i], err = rowenc.ParseDatumStringAs(conv.VisibleColTypes[i], *s, conv.EvalCtx)
			if err != nil {
				__antithesis_instrumentation__.Notify(495870)
				col := conv.VisibleCols[i]
				return newImportRowError(errors.Wrapf(
					err,
					"encountered error when attempting to parse %q as %s",
					col.GetName(), col.GetType().SQLString(),
				), data.String(), rowNum)
			} else {
				__antithesis_instrumentation__.Notify(495871)
			}
		}
	}
	__antithesis_instrumentation__.Notify(495864)
	return nil
}

func (d *pgCopyReader) readFile(
	ctx context.Context, input *fileReader, inputIdx int32, resumePos int64, rejected chan string,
) error {
	__antithesis_instrumentation__.Notify(495872)
	s := bufio.NewScanner(input)
	s.Split(bufio.ScanLines)
	s.Buffer(nil, int(d.opts.MaxRowSize))
	c := newPostgreStreamCopy(
		s,
		d.opts.Delimiter,
		d.opts.Null,
	)

	producer := &pgCopyProducer{
		importCtx:  d.importCtx,
		opts:       &d.opts,
		input:      input,
		copyStream: c,
	}

	consumer := &pgCopyConsumer{
		opts: &d.opts,
	}

	fileCtx := &importFileContext{
		source:   inputIdx,
		skip:     resumePos,
		rejected: rejected,
	}

	return runParallelImport(ctx, d.importCtx, fileCtx, producer, consumer)
}
