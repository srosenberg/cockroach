package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/ioctx"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func runImport(
	ctx context.Context,
	flowCtx *execinfra.FlowCtx,
	spec *execinfrapb.ReadImportDataSpec,
	progCh chan execinfrapb.RemoteProducerMetadata_BulkProcessorProgress,
	seqChunkProvider *row.SeqChunkProvider,
) (*roachpb.BulkOpSummary, error) {
	__antithesis_instrumentation__.Notify(495025)

	kvCh := make(chan row.KVBatch, 10)

	importResolver := newImportTypeResolver(spec.Types)
	for _, table := range spec.Tables {
		__antithesis_instrumentation__.Notify(495031)
		if err := typedesc.HydrateTypesInTableDescriptor(ctx, table.Desc, importResolver); err != nil {
			__antithesis_instrumentation__.Notify(495032)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(495033)
		}
	}
	__antithesis_instrumentation__.Notify(495026)

	evalCtx := flowCtx.NewEvalCtx()

	evalCtx.DB = flowCtx.Cfg.DB
	evalCtx.Regions = makeImportRegionOperator(spec.DatabasePrimaryRegion)
	semaCtx := tree.MakeSemaContext()
	semaCtx.TypeResolver = importResolver
	conv, err := makeInputConverter(ctx, &semaCtx, spec, evalCtx, kvCh, seqChunkProvider)
	if err != nil {
		__antithesis_instrumentation__.Notify(495034)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495035)
	}
	__antithesis_instrumentation__.Notify(495027)

	group := ctxgroup.WithContext(ctx)
	conv.start(group)

	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(495036)
		defer close(kvCh)
		ctx, span := tracing.ChildSpan(ctx, "import-files-to-kvs")
		defer span.Finish()
		var inputs map[int32]string
		if spec.ResumePos != nil {
			__antithesis_instrumentation__.Notify(495038)

			inputs = make(map[int32]string)
			for id, name := range spec.Uri {
				__antithesis_instrumentation__.Notify(495039)
				if seek, ok := spec.ResumePos[id]; !ok || func() bool {
					__antithesis_instrumentation__.Notify(495040)
					return seek < math.MaxInt64 == true
				}() == true {
					__antithesis_instrumentation__.Notify(495041)
					inputs[id] = name
				} else {
					__antithesis_instrumentation__.Notify(495042)
				}
			}
		} else {
			__antithesis_instrumentation__.Notify(495043)
			inputs = spec.Uri
		}
		__antithesis_instrumentation__.Notify(495037)

		return conv.readFiles(ctx, inputs, spec.ResumePos, spec.Format, flowCtx.Cfg.ExternalStorage,
			spec.User())
	})
	__antithesis_instrumentation__.Notify(495028)

	var summary *roachpb.BulkOpSummary
	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(495044)
		summary, err = ingestKvs(ctx, flowCtx, spec, progCh, kvCh)
		if err != nil {
			__antithesis_instrumentation__.Notify(495047)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495048)
		}
		__antithesis_instrumentation__.Notify(495045)
		var prog execinfrapb.RemoteProducerMetadata_BulkProcessorProgress
		prog.ResumePos = make(map[int32]int64)
		prog.CompletedFraction = make(map[int32]float32)
		for i := range spec.Uri {
			__antithesis_instrumentation__.Notify(495049)
			prog.CompletedFraction[i] = 1.0
			prog.ResumePos[i] = math.MaxInt64
		}
		__antithesis_instrumentation__.Notify(495046)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(495050)
			return ctx.Err()
		case progCh <- prog:
			__antithesis_instrumentation__.Notify(495051)
			return nil
		}
	})
	__antithesis_instrumentation__.Notify(495029)

	if err = group.Wait(); err != nil {
		__antithesis_instrumentation__.Notify(495052)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(495053)
	}
	__antithesis_instrumentation__.Notify(495030)

	return summary, nil
}

type readFileFunc func(context.Context, *fileReader, int32, int64, chan string) error

func readInputFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	resumePos map[int32]int64,
	format roachpb.IOFileFormat,
	fileFunc readFileFunc,
	makeExternalStorage cloud.ExternalStorageFactory,
	user security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(495054)
	done := ctx.Done()

	fileSizes := make(map[int32]int64, len(dataFiles))

	for id, dataFile := range dataFiles {
		__antithesis_instrumentation__.Notify(495057)
		if err := func() error {
			__antithesis_instrumentation__.Notify(495058)

			conf, err := cloud.ExternalStorageConfFromURI(dataFile, user)
			if err != nil {
				__antithesis_instrumentation__.Notify(495063)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495064)
			}
			__antithesis_instrumentation__.Notify(495059)
			es, err := makeExternalStorage(ctx, conf)
			if err != nil {
				__antithesis_instrumentation__.Notify(495065)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495066)
			}
			__antithesis_instrumentation__.Notify(495060)
			defer es.Close()

			sz, err := es.Size(ctx, "")

			if sz <= 0 {
				__antithesis_instrumentation__.Notify(495067)

				log.Infof(ctx, "could not fetch file size; falling back to per-file progress: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(495068)
				fileSizes[id] = sz
			}
			__antithesis_instrumentation__.Notify(495061)

			if len(dataFiles) > 1 {
				__antithesis_instrumentation__.Notify(495069)

				raw, err := es.ReadFile(ctx, "")
				if err != nil {
					__antithesis_instrumentation__.Notify(495071)
					return err
				} else {
					__antithesis_instrumentation__.Notify(495072)
				}
				__antithesis_instrumentation__.Notify(495070)
				defer raw.Close(ctx)

				p := make([]byte, 1)
				if _, err := raw.Read(ctx, p); err != nil && func() bool {
					__antithesis_instrumentation__.Notify(495073)
					return err != io.EOF == true
				}() == true {
					__antithesis_instrumentation__.Notify(495074)

					return err
				} else {
					__antithesis_instrumentation__.Notify(495075)
				}

			} else {
				__antithesis_instrumentation__.Notify(495076)
			}
			__antithesis_instrumentation__.Notify(495062)

			return nil
		}(); err != nil {
			__antithesis_instrumentation__.Notify(495077)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495078)
		}
	}
	__antithesis_instrumentation__.Notify(495055)

	for dataFileIndex, dataFile := range dataFiles {
		__antithesis_instrumentation__.Notify(495079)
		select {
		case <-done:
			__antithesis_instrumentation__.Notify(495081)
			return ctx.Err()
		default:
			__antithesis_instrumentation__.Notify(495082)
		}
		__antithesis_instrumentation__.Notify(495080)
		if err := func() error {
			__antithesis_instrumentation__.Notify(495083)
			conf, err := cloud.ExternalStorageConfFromURI(dataFile, user)
			if err != nil {
				__antithesis_instrumentation__.Notify(495090)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495091)
			}
			__antithesis_instrumentation__.Notify(495084)
			es, err := makeExternalStorage(ctx, conf)
			if err != nil {
				__antithesis_instrumentation__.Notify(495092)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495093)
			}
			__antithesis_instrumentation__.Notify(495085)
			defer es.Close()
			raw, err := es.ReadFile(ctx, "")
			if err != nil {
				__antithesis_instrumentation__.Notify(495094)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495095)
			}
			__antithesis_instrumentation__.Notify(495086)
			defer raw.Close(ctx)

			src := &fileReader{total: fileSizes[dataFileIndex], counter: byteCounter{r: ioctx.ReaderCtxAdapter(ctx, raw)}}
			decompressed, err := decompressingReader(&src.counter, dataFile, format.Compression)
			if err != nil {
				__antithesis_instrumentation__.Notify(495096)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495097)
			}
			__antithesis_instrumentation__.Notify(495087)
			defer decompressed.Close()
			src.Reader = decompressed

			var rejected chan string
			if (format.Format == roachpb.IOFileFormat_CSV && func() bool {
				__antithesis_instrumentation__.Notify(495098)
				return format.SaveRejected == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(495099)
				return (format.Format == roachpb.IOFileFormat_MysqlOutfile && func() bool {
					__antithesis_instrumentation__.Notify(495100)
					return format.SaveRejected == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(495101)
				rejected = make(chan string)
			} else {
				__antithesis_instrumentation__.Notify(495102)
			}
			__antithesis_instrumentation__.Notify(495088)
			if rejected != nil {
				__antithesis_instrumentation__.Notify(495103)
				grp := ctxgroup.WithContext(ctx)
				grp.GoCtx(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(495106)
					var buf []byte
					var countRejected int64
					for s := range rejected {
						__antithesis_instrumentation__.Notify(495113)
						countRejected++
						if countRejected > 1000 {
							__antithesis_instrumentation__.Notify(495115)
							return pgerror.Newf(
								pgcode.DataCorrupted,
								"too many parsing errors (%d) encountered for file %s",
								countRejected,
								dataFile,
							)
						} else {
							__antithesis_instrumentation__.Notify(495116)
						}
						__antithesis_instrumentation__.Notify(495114)
						buf = append(buf, s...)
					}
					__antithesis_instrumentation__.Notify(495107)
					if countRejected == 0 {
						__antithesis_instrumentation__.Notify(495117)

						return nil
					} else {
						__antithesis_instrumentation__.Notify(495118)
					}
					__antithesis_instrumentation__.Notify(495108)
					rejFn, err := rejectedFilename(dataFile)
					if err != nil {
						__antithesis_instrumentation__.Notify(495119)
						return err
					} else {
						__antithesis_instrumentation__.Notify(495120)
					}
					__antithesis_instrumentation__.Notify(495109)
					conf, err := cloud.ExternalStorageConfFromURI(rejFn, user)
					if err != nil {
						__antithesis_instrumentation__.Notify(495121)
						return err
					} else {
						__antithesis_instrumentation__.Notify(495122)
					}
					__antithesis_instrumentation__.Notify(495110)
					rejectedStorage, err := makeExternalStorage(ctx, conf)
					if err != nil {
						__antithesis_instrumentation__.Notify(495123)
						return err
					} else {
						__antithesis_instrumentation__.Notify(495124)
					}
					__antithesis_instrumentation__.Notify(495111)
					defer rejectedStorage.Close()
					if err := cloud.WriteFile(ctx, rejectedStorage, "", bytes.NewReader(buf)); err != nil {
						__antithesis_instrumentation__.Notify(495125)
						return err
					} else {
						__antithesis_instrumentation__.Notify(495126)
					}
					__antithesis_instrumentation__.Notify(495112)
					return nil
				})
				__antithesis_instrumentation__.Notify(495104)

				grp.GoCtx(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(495127)
					defer close(rejected)
					if err := fileFunc(ctx, src, dataFileIndex, resumePos[dataFileIndex], rejected); err != nil {
						__antithesis_instrumentation__.Notify(495129)
						return err
					} else {
						__antithesis_instrumentation__.Notify(495130)
					}
					__antithesis_instrumentation__.Notify(495128)
					return nil
				})
				__antithesis_instrumentation__.Notify(495105)

				if err := grp.Wait(); err != nil {
					__antithesis_instrumentation__.Notify(495131)
					return errors.Wrapf(err, "%s", dataFile)
				} else {
					__antithesis_instrumentation__.Notify(495132)
				}
			} else {
				__antithesis_instrumentation__.Notify(495133)
				if err := fileFunc(ctx, src, dataFileIndex, resumePos[dataFileIndex], nil); err != nil {
					__antithesis_instrumentation__.Notify(495134)
					return errors.Wrapf(err, "%s", dataFile)
				} else {
					__antithesis_instrumentation__.Notify(495135)
				}
			}
			__antithesis_instrumentation__.Notify(495089)
			return nil
		}(); err != nil {
			__antithesis_instrumentation__.Notify(495136)
			return err
		} else {
			__antithesis_instrumentation__.Notify(495137)
		}
	}
	__antithesis_instrumentation__.Notify(495056)
	return nil
}

func rejectedFilename(datafile string) (string, error) {
	__antithesis_instrumentation__.Notify(495138)
	parsedURI, err := url.Parse(datafile)
	if err != nil {
		__antithesis_instrumentation__.Notify(495140)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(495141)
	}
	__antithesis_instrumentation__.Notify(495139)
	parsedURI.Path = parsedURI.Path + ".rejected"
	return parsedURI.String(), nil
}

func decompressingReader(
	in io.Reader, name string, hint roachpb.IOFileFormat_Compression,
) (io.ReadCloser, error) {
	__antithesis_instrumentation__.Notify(495142)
	switch guessCompressionFromName(name, hint) {
	case roachpb.IOFileFormat_Gzip:
		__antithesis_instrumentation__.Notify(495143)
		return gzip.NewReader(in)
	case roachpb.IOFileFormat_Bzip:
		__antithesis_instrumentation__.Notify(495144)
		return ioutil.NopCloser(bzip2.NewReader(in)), nil
	default:
		__antithesis_instrumentation__.Notify(495145)
		return ioutil.NopCloser(in), nil
	}
}

func guessCompressionFromName(
	name string, hint roachpb.IOFileFormat_Compression,
) roachpb.IOFileFormat_Compression {
	__antithesis_instrumentation__.Notify(495146)
	if hint != roachpb.IOFileFormat_Auto {
		__antithesis_instrumentation__.Notify(495148)
		return hint
	} else {
		__antithesis_instrumentation__.Notify(495149)
	}
	__antithesis_instrumentation__.Notify(495147)
	switch {
	case strings.HasSuffix(name, ".gz"):
		__antithesis_instrumentation__.Notify(495150)
		return roachpb.IOFileFormat_Gzip
	case strings.HasSuffix(name, ".bz2") || func() bool {
		__antithesis_instrumentation__.Notify(495154)
		return strings.HasSuffix(name, ".bz") == true
	}() == true:
		__antithesis_instrumentation__.Notify(495151)
		return roachpb.IOFileFormat_Bzip
	default:
		__antithesis_instrumentation__.Notify(495152)
		if parsed, err := url.Parse(name); err == nil && func() bool {
			__antithesis_instrumentation__.Notify(495155)
			return parsed.Path != name == true
		}() == true {
			__antithesis_instrumentation__.Notify(495156)
			return guessCompressionFromName(parsed.Path, hint)
		} else {
			__antithesis_instrumentation__.Notify(495157)
		}
		__antithesis_instrumentation__.Notify(495153)
		return roachpb.IOFileFormat_None
	}
}

type byteCounter struct {
	r io.Reader
	n int64
}

func (b *byteCounter) Read(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(495158)
	n, err := b.r.Read(p)
	b.n += int64(n)
	return n, err
}

type fileReader struct {
	io.Reader
	total   int64
	counter byteCounter
}

func (f fileReader) ReadFraction() float32 {
	__antithesis_instrumentation__.Notify(495159)
	if f.total == 0 {
		__antithesis_instrumentation__.Notify(495161)
		return 0.0
	} else {
		__antithesis_instrumentation__.Notify(495162)
	}
	__antithesis_instrumentation__.Notify(495160)
	return float32(f.counter.n) / float32(f.total)
}

type inputConverter interface {
	start(group ctxgroup.Group)
	readFiles(ctx context.Context, dataFiles map[int32]string, resumePos map[int32]int64,
		format roachpb.IOFileFormat, makeExternalStorage cloud.ExternalStorageFactory, user security.SQLUsername) error
}

func formatHasNamedColumns(format roachpb.IOFileFormat_FileFormat) bool {
	__antithesis_instrumentation__.Notify(495163)
	switch format {
	case roachpb.IOFileFormat_Avro,
		roachpb.IOFileFormat_Mysqldump,
		roachpb.IOFileFormat_PgDump:
		__antithesis_instrumentation__.Notify(495165)
		return true
	default:
		__antithesis_instrumentation__.Notify(495166)
	}
	__antithesis_instrumentation__.Notify(495164)
	return false
}

func isMultiTableFormat(format roachpb.IOFileFormat_FileFormat) bool {
	__antithesis_instrumentation__.Notify(495167)
	switch format {
	case roachpb.IOFileFormat_Mysqldump,
		roachpb.IOFileFormat_PgDump:
		__antithesis_instrumentation__.Notify(495169)
		return true
	default:
		__antithesis_instrumentation__.Notify(495170)
	}
	__antithesis_instrumentation__.Notify(495168)
	return false
}

func makeRowErr(row int64, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(495171)
	err := pgerror.NewWithDepthf(1, code, format, args...)
	err = errors.WrapWithDepthf(1, err, "row %d", row)
	return err
}

func wrapRowErr(err error, row int64, code pgcode.Code, format string, args ...interface{}) error {
	__antithesis_instrumentation__.Notify(495172)
	if format != "" || func() bool {
		__antithesis_instrumentation__.Notify(495175)
		return len(args) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(495176)
		err = errors.WrapWithDepthf(1, err, format, args...)
	} else {
		__antithesis_instrumentation__.Notify(495177)
	}
	__antithesis_instrumentation__.Notify(495173)
	err = errors.WrapWithDepthf(1, err, "row %d", row)
	if code != pgcode.Uncategorized {
		__antithesis_instrumentation__.Notify(495178)
		err = pgerror.WithCandidateCode(err, code)
	} else {
		__antithesis_instrumentation__.Notify(495179)
	}
	__antithesis_instrumentation__.Notify(495174)
	return err
}

type importRowError struct {
	err    error
	row    string
	rowNum int64
}

const (
	importRowErrMaxRuneCount    = 1024
	importRowErrTruncatedMarker = " -- TRUNCATED"
)

func (e *importRowError) Error() string {
	__antithesis_instrumentation__.Notify(495180)

	rowForLog := e.row
	if len(rowForLog) > importRowErrMaxRuneCount {
		__antithesis_instrumentation__.Notify(495182)
		rowForLog = util.TruncateString(rowForLog, importRowErrMaxRuneCount) + importRowErrTruncatedMarker
	} else {
		__antithesis_instrumentation__.Notify(495183)
	}
	__antithesis_instrumentation__.Notify(495181)
	return fmt.Sprintf("error parsing row %d: %v (row: %s)", e.rowNum, e.err, rowForLog)
}

func newImportRowError(err error, row string, num int64) error {
	__antithesis_instrumentation__.Notify(495184)
	return &importRowError{
		err:    err,
		row:    row,
		rowNum: num,
	}
}

type parallelImportContext struct {
	walltime         int64
	numWorkers       int
	batchSize        int
	semaCtx          *tree.SemaContext
	evalCtx          *tree.EvalContext
	tableDesc        catalog.TableDescriptor
	targetCols       tree.NameList
	kvCh             chan row.KVBatch
	seqChunkProvider *row.SeqChunkProvider
}

type importFileContext struct {
	source   int32
	skip     int64
	rejected chan string
	rowLimit int64
}

func handleCorruptRow(ctx context.Context, fileCtx *importFileContext, err error) error {
	__antithesis_instrumentation__.Notify(495185)
	log.Errorf(ctx, "%+v", err)

	if rowErr := (*importRowError)(nil); errors.As(err, &rowErr) && func() bool {
		__antithesis_instrumentation__.Notify(495187)
		return fileCtx.rejected != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(495188)
		fileCtx.rejected <- rowErr.row + "\n"
		return nil
	} else {
		__antithesis_instrumentation__.Notify(495189)
	}
	__antithesis_instrumentation__.Notify(495186)

	return err
}

func makeDatumConverter(
	ctx context.Context, importCtx *parallelImportContext, fileCtx *importFileContext,
) (*row.DatumRowConverter, error) {
	__antithesis_instrumentation__.Notify(495190)
	conv, err := row.NewDatumRowConverter(
		ctx, importCtx.semaCtx, importCtx.tableDesc, importCtx.targetCols, importCtx.evalCtx,
		importCtx.kvCh, importCtx.seqChunkProvider, nil)
	if err == nil {
		__antithesis_instrumentation__.Notify(495192)
		conv.KvBatch.Source = fileCtx.source
	} else {
		__antithesis_instrumentation__.Notify(495193)
	}
	__antithesis_instrumentation__.Notify(495191)
	return conv, err
}

type importRowProducer interface {
	Scan() bool

	Err() error

	Skip() error

	Row() (interface{}, error)

	Progress() float32
}

type importRowConsumer interface {
	FillDatums(row interface{}, rowNum int64, conv *row.DatumRowConverter) error
}

type batch struct {
	data     []interface{}
	startPos int64
	progress float32
}

type parallelImporter struct {
	b         batch
	batchSize int
	recordCh  chan batch
}

var parallelImporterReaderBatchSize = 500

func TestingSetParallelImporterReaderBatchSize(s int) func() {
	__antithesis_instrumentation__.Notify(495194)
	parallelImporterReaderBatchSize = s
	return func() {
		__antithesis_instrumentation__.Notify(495195)
		parallelImporterReaderBatchSize = 500
	}
}

func runParallelImport(
	ctx context.Context,
	importCtx *parallelImportContext,
	fileCtx *importFileContext,
	producer importRowProducer,
	consumer importRowConsumer,
) error {
	__antithesis_instrumentation__.Notify(495196)
	batchSize := importCtx.batchSize
	if batchSize <= 0 {
		__antithesis_instrumentation__.Notify(495200)
		batchSize = parallelImporterReaderBatchSize
	} else {
		__antithesis_instrumentation__.Notify(495201)
	}
	__antithesis_instrumentation__.Notify(495197)
	importer := &parallelImporter{
		b: batch{
			data: make([]interface{}, 0, batchSize),
		},
		batchSize: batchSize,
		recordCh:  make(chan batch),
	}

	group := ctxgroup.WithContext(ctx)

	minEmited := make([]int64, importCtx.numWorkers)
	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(495202)
		var span *tracing.Span
		ctx, span = tracing.ChildSpan(ctx, "import-rows-to-datums")
		defer span.Finish()
		if importCtx.numWorkers <= 0 {
			__antithesis_instrumentation__.Notify(495204)
			return errors.AssertionFailedf("invalid parallelism: %d", importCtx.numWorkers)
		} else {
			__antithesis_instrumentation__.Notify(495205)
		}
		__antithesis_instrumentation__.Notify(495203)
		return ctxgroup.GroupWorkers(ctx, importCtx.numWorkers, func(ctx context.Context, id int) error {
			__antithesis_instrumentation__.Notify(495206)
			return importer.importWorker(ctx, id, consumer, importCtx, fileCtx, minEmited)
		})
	})
	__antithesis_instrumentation__.Notify(495198)

	group.GoCtx(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(495207)
		defer close(importer.recordCh)
		var span *tracing.Span
		ctx, span = tracing.ChildSpan(ctx, "import-file-to-rows")
		defer span.Finish()
		var numSkipped int64
		var count int64
		for producer.Scan() {
			__antithesis_instrumentation__.Notify(495210)

			count++
			if count <= fileCtx.skip {
				__antithesis_instrumentation__.Notify(495214)
				if err := producer.Skip(); err != nil {
					__antithesis_instrumentation__.Notify(495216)
					return err
				} else {
					__antithesis_instrumentation__.Notify(495217)
				}
				__antithesis_instrumentation__.Notify(495215)
				numSkipped++
				continue
			} else {
				__antithesis_instrumentation__.Notify(495218)
			}
			__antithesis_instrumentation__.Notify(495211)

			rowBeingProcessedIdx := count - numSkipped
			if fileCtx.rowLimit != 0 && func() bool {
				__antithesis_instrumentation__.Notify(495219)
				return rowBeingProcessedIdx > fileCtx.rowLimit == true
			}() == true {
				__antithesis_instrumentation__.Notify(495220)
				break
			} else {
				__antithesis_instrumentation__.Notify(495221)
			}
			__antithesis_instrumentation__.Notify(495212)

			data, err := producer.Row()
			if err != nil {
				__antithesis_instrumentation__.Notify(495222)
				if err = handleCorruptRow(ctx, fileCtx, err); err != nil {
					__antithesis_instrumentation__.Notify(495224)
					return err
				} else {
					__antithesis_instrumentation__.Notify(495225)
				}
				__antithesis_instrumentation__.Notify(495223)
				continue
			} else {
				__antithesis_instrumentation__.Notify(495226)
			}
			__antithesis_instrumentation__.Notify(495213)

			if err := importer.add(ctx, data, count, producer.Progress); err != nil {
				__antithesis_instrumentation__.Notify(495227)
				return err
			} else {
				__antithesis_instrumentation__.Notify(495228)
			}
		}
		__antithesis_instrumentation__.Notify(495208)

		if producer.Err() == nil {
			__antithesis_instrumentation__.Notify(495229)
			return importer.close(ctx)
		} else {
			__antithesis_instrumentation__.Notify(495230)
		}
		__antithesis_instrumentation__.Notify(495209)
		return producer.Err()
	})
	__antithesis_instrumentation__.Notify(495199)

	return group.Wait()
}

func (p *parallelImporter) add(
	ctx context.Context, data interface{}, pos int64, progress func() float32,
) error {
	__antithesis_instrumentation__.Notify(495231)
	if len(p.b.data) == 0 {
		__antithesis_instrumentation__.Notify(495234)
		p.b.startPos = pos
	} else {
		__antithesis_instrumentation__.Notify(495235)
	}
	__antithesis_instrumentation__.Notify(495232)
	p.b.data = append(p.b.data, data)

	if len(p.b.data) == p.batchSize {
		__antithesis_instrumentation__.Notify(495236)
		p.b.progress = progress()
		return p.flush(ctx)
	} else {
		__antithesis_instrumentation__.Notify(495237)
	}
	__antithesis_instrumentation__.Notify(495233)
	return nil
}

func (p *parallelImporter) close(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(495238)
	if len(p.b.data) > 0 {
		__antithesis_instrumentation__.Notify(495240)
		return p.flush(ctx)
	} else {
		__antithesis_instrumentation__.Notify(495241)
	}
	__antithesis_instrumentation__.Notify(495239)
	return nil
}

func (p *parallelImporter) flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(495242)
	select {
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(495243)
		return ctx.Err()
	case p.recordCh <- p.b:
		__antithesis_instrumentation__.Notify(495244)
		p.b = batch{
			data: make([]interface{}, 0, cap(p.b.data)),
		}
		return nil
	}
}

func timestampAfterEpoch(walltime int64) uint64 {
	__antithesis_instrumentation__.Notify(495245)
	epoch := time.Date(2015, time.January, 1, 0, 0, 0, 0, time.UTC).UnixNano()
	const precision = uint64(10 * time.Microsecond)
	return uint64(walltime-epoch) / precision
}

func (p *parallelImporter) importWorker(
	ctx context.Context,
	workerID int,
	consumer importRowConsumer,
	importCtx *parallelImportContext,
	fileCtx *importFileContext,
	minEmitted []int64,
) error {
	__antithesis_instrumentation__.Notify(495246)
	conv, err := makeDatumConverter(ctx, importCtx, fileCtx)
	if err != nil {
		__antithesis_instrumentation__.Notify(495251)
		return err
	} else {
		__antithesis_instrumentation__.Notify(495252)
	}
	__antithesis_instrumentation__.Notify(495247)
	if conv.EvalCtx.SessionData() == nil {
		__antithesis_instrumentation__.Notify(495253)
		panic("uninitialized session data")
	} else {
		__antithesis_instrumentation__.Notify(495254)
	}
	__antithesis_instrumentation__.Notify(495248)

	var rowNum int64
	timestamp := timestampAfterEpoch(importCtx.walltime)

	conv.CompletedRowFn = func() int64 {
		__antithesis_instrumentation__.Notify(495255)
		m := emittedRowLowWatermark(workerID, rowNum, minEmitted)
		return m
	}
	__antithesis_instrumentation__.Notify(495249)

	for batch := range p.recordCh {
		__antithesis_instrumentation__.Notify(495256)
		conv.KvBatch.Progress = batch.progress
		for batchIdx, record := range batch.data {
			__antithesis_instrumentation__.Notify(495257)
			rowNum = batch.startPos + int64(batchIdx)
			if err := consumer.FillDatums(record, rowNum, conv); err != nil {
				__antithesis_instrumentation__.Notify(495259)
				if err = handleCorruptRow(ctx, fileCtx, err); err != nil {
					__antithesis_instrumentation__.Notify(495261)
					return err
				} else {
					__antithesis_instrumentation__.Notify(495262)
				}
				__antithesis_instrumentation__.Notify(495260)
				continue
			} else {
				__antithesis_instrumentation__.Notify(495263)
			}
			__antithesis_instrumentation__.Notify(495258)

			rowIndex := int64(timestamp) + rowNum
			if err := conv.Row(ctx, conv.KvBatch.Source, rowIndex); err != nil {
				__antithesis_instrumentation__.Notify(495264)
				return newImportRowError(err, fmt.Sprintf("%v", record), rowNum)
			} else {
				__antithesis_instrumentation__.Notify(495265)
			}
		}
	}
	__antithesis_instrumentation__.Notify(495250)
	return conv.SendBatch(ctx)
}

func emittedRowLowWatermark(workerID int, emittedRow int64, minEmitted []int64) int64 {
	__antithesis_instrumentation__.Notify(495266)
	atomic.StoreInt64(&minEmitted[workerID], emittedRow)

	for i := 0; i < len(minEmitted); i++ {
		__antithesis_instrumentation__.Notify(495268)
		if i != workerID {
			__antithesis_instrumentation__.Notify(495269)
			w := atomic.LoadInt64(&minEmitted[i])
			if w < emittedRow {
				__antithesis_instrumentation__.Notify(495270)
				emittedRow = w
			} else {
				__antithesis_instrumentation__.Notify(495271)
			}
		} else {
			__antithesis_instrumentation__.Notify(495272)
		}
	}
	__antithesis_instrumentation__.Notify(495267)

	return emittedRow
}
