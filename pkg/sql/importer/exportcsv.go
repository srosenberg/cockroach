package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/rowexec"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/builtins"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

const (
	exportFilePatternPart    = "%part%"
	exportFilePatternDefault = exportFilePatternPart + ".csv"
)

type csvExporter struct {
	compressor *gzip.Writer
	buf        *bytes.Buffer
	csvWriter  *csv.Writer
}

func (c *csvExporter) Write(record []string) error {
	__antithesis_instrumentation__.Notify(493086)
	return c.csvWriter.Write(record)
}

func (c *csvExporter) Close() error {
	__antithesis_instrumentation__.Notify(493087)
	if c.compressor != nil {
		__antithesis_instrumentation__.Notify(493089)
		return c.compressor.Close()
	} else {
		__antithesis_instrumentation__.Notify(493090)
	}
	__antithesis_instrumentation__.Notify(493088)
	return nil
}

func (c *csvExporter) Flush() error {
	__antithesis_instrumentation__.Notify(493091)
	c.csvWriter.Flush()
	if c.compressor != nil {
		__antithesis_instrumentation__.Notify(493093)
		return c.compressor.Flush()
	} else {
		__antithesis_instrumentation__.Notify(493094)
	}
	__antithesis_instrumentation__.Notify(493092)
	return nil
}

func (c *csvExporter) ResetBuffer() {
	__antithesis_instrumentation__.Notify(493095)
	c.buf.Reset()
	if c.compressor != nil {
		__antithesis_instrumentation__.Notify(493096)

		c.compressor.Reset(c.buf)
	} else {
		__antithesis_instrumentation__.Notify(493097)
	}
}

func (c *csvExporter) Bytes() []byte {
	__antithesis_instrumentation__.Notify(493098)
	return c.buf.Bytes()
}

func (c *csvExporter) Len() int {
	__antithesis_instrumentation__.Notify(493099)
	return c.buf.Len()
}

func (c *csvExporter) FileName(spec execinfrapb.ExportSpec, part string) string {
	__antithesis_instrumentation__.Notify(493100)
	pattern := exportFilePatternDefault
	if spec.NamePattern != "" {
		__antithesis_instrumentation__.Notify(493103)
		pattern = spec.NamePattern
	} else {
		__antithesis_instrumentation__.Notify(493104)
	}
	__antithesis_instrumentation__.Notify(493101)

	fileName := strings.Replace(pattern, exportFilePatternPart, part, -1)

	if c.compressor != nil {
		__antithesis_instrumentation__.Notify(493105)
		fileName += ".gz"
	} else {
		__antithesis_instrumentation__.Notify(493106)
	}
	__antithesis_instrumentation__.Notify(493102)
	return fileName
}

func newCSVExporter(sp execinfrapb.ExportSpec) *csvExporter {
	__antithesis_instrumentation__.Notify(493107)
	buf := bytes.NewBuffer([]byte{})
	var exporter *csvExporter
	switch sp.Format.Compression {
	case roachpb.IOFileFormat_Gzip:
		__antithesis_instrumentation__.Notify(493110)
		{
			__antithesis_instrumentation__.Notify(493112)
			writer := gzip.NewWriter(buf)
			exporter = &csvExporter{
				compressor: writer,
				buf:        buf,
				csvWriter:  csv.NewWriter(writer),
			}
		}
	default:
		__antithesis_instrumentation__.Notify(493111)
		{
			__antithesis_instrumentation__.Notify(493113)
			exporter = &csvExporter{
				buf:       buf,
				csvWriter: csv.NewWriter(buf),
			}
		}
	}
	__antithesis_instrumentation__.Notify(493108)
	if sp.Format.Csv.Comma != 0 {
		__antithesis_instrumentation__.Notify(493114)
		exporter.csvWriter.Comma = sp.Format.Csv.Comma
	} else {
		__antithesis_instrumentation__.Notify(493115)
	}
	__antithesis_instrumentation__.Notify(493109)
	return exporter
}

func newCSVWriterProcessor(
	flowCtx *execinfra.FlowCtx,
	processorID int32,
	spec execinfrapb.ExportSpec,
	input execinfra.RowSource,
	output execinfra.RowReceiver,
) (execinfra.Processor, error) {
	__antithesis_instrumentation__.Notify(493116)
	c := &csvWriter{
		flowCtx:     flowCtx,
		processorID: processorID,
		spec:        spec,
		input:       input,
		output:      output,
	}
	semaCtx := tree.MakeSemaContext()
	if err := c.out.Init(&execinfrapb.PostProcessSpec{}, c.OutputTypes(), &semaCtx, flowCtx.NewEvalCtx()); err != nil {
		__antithesis_instrumentation__.Notify(493118)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(493119)
	}
	__antithesis_instrumentation__.Notify(493117)
	return c, nil
}

type csvWriter struct {
	flowCtx     *execinfra.FlowCtx
	processorID int32
	spec        execinfrapb.ExportSpec
	input       execinfra.RowSource
	out         execinfra.ProcOutputHelper
	output      execinfra.RowReceiver
}

var _ execinfra.Processor = &csvWriter{}

func (sp *csvWriter) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(493120)
	res := make([]*types.T, len(colinfo.ExportColumns))
	for i := range res {
		__antithesis_instrumentation__.Notify(493122)
		res[i] = colinfo.ExportColumns[i].Typ
	}
	__antithesis_instrumentation__.Notify(493121)
	return res
}

func (sp *csvWriter) MustBeStreaming() bool {
	__antithesis_instrumentation__.Notify(493123)
	return false
}

func (sp *csvWriter) Run(ctx context.Context) {
	__antithesis_instrumentation__.Notify(493124)
	ctx, span := tracing.ChildSpan(ctx, "csvWriter")
	defer span.Finish()

	instanceID := sp.flowCtx.EvalCtx.NodeID.SQLInstanceID()
	uniqueID := builtins.GenerateUniqueInt(instanceID)

	err := func() error {
		__antithesis_instrumentation__.Notify(493126)
		typs := sp.input.OutputTypes()
		sp.input.Start(ctx)
		input := execinfra.MakeNoMetadataRowSource(sp.input, sp.output)

		alloc := &tree.DatumAlloc{}

		writer := newCSVExporter(sp.spec)

		var nullsAs string
		if sp.spec.Format.Csv.NullEncoding != nil {
			__antithesis_instrumentation__.Notify(493129)
			nullsAs = *sp.spec.Format.Csv.NullEncoding
		} else {
			__antithesis_instrumentation__.Notify(493130)
		}
		__antithesis_instrumentation__.Notify(493127)
		f := tree.NewFmtCtx(tree.FmtExport)
		defer f.Close()

		csvRow := make([]string, len(typs))

		chunk := 0
		done := false
		for {
			__antithesis_instrumentation__.Notify(493131)
			var rows int64
			writer.ResetBuffer()
			for {
				__antithesis_instrumentation__.Notify(493141)

				if int64(writer.buf.Len()) >= sp.spec.ChunkSize {
					__antithesis_instrumentation__.Notify(493147)
					break
				} else {
					__antithesis_instrumentation__.Notify(493148)
				}
				__antithesis_instrumentation__.Notify(493142)
				if sp.spec.ChunkRows > 0 && func() bool {
					__antithesis_instrumentation__.Notify(493149)
					return rows >= sp.spec.ChunkRows == true
				}() == true {
					__antithesis_instrumentation__.Notify(493150)
					break
				} else {
					__antithesis_instrumentation__.Notify(493151)
				}
				__antithesis_instrumentation__.Notify(493143)
				row, err := input.NextRow()
				if err != nil {
					__antithesis_instrumentation__.Notify(493152)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493153)
				}
				__antithesis_instrumentation__.Notify(493144)
				if row == nil {
					__antithesis_instrumentation__.Notify(493154)
					done = true
					break
				} else {
					__antithesis_instrumentation__.Notify(493155)
				}
				__antithesis_instrumentation__.Notify(493145)
				rows++

				for i, ed := range row {
					__antithesis_instrumentation__.Notify(493156)
					if ed.IsNull() {
						__antithesis_instrumentation__.Notify(493159)
						if sp.spec.Format.Csv.NullEncoding != nil {
							__antithesis_instrumentation__.Notify(493160)
							csvRow[i] = nullsAs
							continue
						} else {
							__antithesis_instrumentation__.Notify(493161)
							return errors.New("NULL value encountered during EXPORT, " +
								"use `WITH nullas` to specify the string representation of NULL")
						}
					} else {
						__antithesis_instrumentation__.Notify(493162)
					}
					__antithesis_instrumentation__.Notify(493157)
					if err := ed.EnsureDecoded(typs[i], alloc); err != nil {
						__antithesis_instrumentation__.Notify(493163)
						return err
					} else {
						__antithesis_instrumentation__.Notify(493164)
					}
					__antithesis_instrumentation__.Notify(493158)
					ed.Datum.Format(f)
					csvRow[i] = f.String()
					f.Reset()
				}
				__antithesis_instrumentation__.Notify(493146)
				if err := writer.Write(csvRow); err != nil {
					__antithesis_instrumentation__.Notify(493165)
					return err
				} else {
					__antithesis_instrumentation__.Notify(493166)
				}
			}
			__antithesis_instrumentation__.Notify(493132)
			if rows < 1 {
				__antithesis_instrumentation__.Notify(493167)
				break
			} else {
				__antithesis_instrumentation__.Notify(493168)
			}
			__antithesis_instrumentation__.Notify(493133)
			if err := writer.Flush(); err != nil {
				__antithesis_instrumentation__.Notify(493169)
				return errors.Wrap(err, "failed to flush csv writer")
			} else {
				__antithesis_instrumentation__.Notify(493170)
			}
			__antithesis_instrumentation__.Notify(493134)

			conf, err := cloud.ExternalStorageConfFromURI(sp.spec.Destination, sp.spec.User())
			if err != nil {
				__antithesis_instrumentation__.Notify(493171)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493172)
			}
			__antithesis_instrumentation__.Notify(493135)
			es, err := sp.flowCtx.Cfg.ExternalStorage(ctx, conf)
			if err != nil {
				__antithesis_instrumentation__.Notify(493173)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493174)
			}
			__antithesis_instrumentation__.Notify(493136)
			defer es.Close()

			part := fmt.Sprintf("n%d.%d", uniqueID, chunk)
			chunk++
			filename := writer.FileName(sp.spec, part)

			err = writer.Close()
			if err != nil {
				__antithesis_instrumentation__.Notify(493175)
				return errors.Wrapf(err, "failed to close exporting writer")
			} else {
				__antithesis_instrumentation__.Notify(493176)
			}
			__antithesis_instrumentation__.Notify(493137)

			size := writer.Len()

			if err := cloud.WriteFile(ctx, es, filename, bytes.NewReader(writer.Bytes())); err != nil {
				__antithesis_instrumentation__.Notify(493177)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493178)
			}
			__antithesis_instrumentation__.Notify(493138)
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
				__antithesis_instrumentation__.Notify(493179)
				return err
			} else {
				__antithesis_instrumentation__.Notify(493180)
			}
			__antithesis_instrumentation__.Notify(493139)
			if cs != execinfra.NeedMoreRows {
				__antithesis_instrumentation__.Notify(493181)

				return errors.New("unexpected closure of consumer")
			} else {
				__antithesis_instrumentation__.Notify(493182)
			}
			__antithesis_instrumentation__.Notify(493140)
			if done {
				__antithesis_instrumentation__.Notify(493183)
				break
			} else {
				__antithesis_instrumentation__.Notify(493184)
			}
		}
		__antithesis_instrumentation__.Notify(493128)

		return nil
	}()
	__antithesis_instrumentation__.Notify(493125)

	execinfra.DrainAndClose(
		ctx, sp.output, err, func(context.Context) { __antithesis_instrumentation__.Notify(493185) }, sp.input)
}

func init() {
	rowexec.NewCSVWriterProcessor = newCSVWriterProcessor
}
