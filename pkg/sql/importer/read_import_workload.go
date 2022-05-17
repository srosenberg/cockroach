package importer

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"unsafe"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/row"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil/pgdate"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/errors"
)

type workloadReader struct {
	semaCtx     *tree.SemaContext
	evalCtx     *tree.EvalContext
	table       catalog.TableDescriptor
	kvCh        chan row.KVBatch
	parallelism int
}

var _ inputConverter = &workloadReader{}

func newWorkloadReader(
	semaCtx *tree.SemaContext,
	evalCtx *tree.EvalContext,
	table catalog.TableDescriptor,
	kvCh chan row.KVBatch,
	parallelism int,
) *workloadReader {
	__antithesis_instrumentation__.Notify(496549)
	return &workloadReader{semaCtx: semaCtx, evalCtx: evalCtx, table: table, kvCh: kvCh, parallelism: parallelism}
}

func (w *workloadReader) start(ctx ctxgroup.Group) {
	__antithesis_instrumentation__.Notify(496550)
}

func makeDatumFromColOffset(
	alloc *tree.DatumAlloc, hint *types.T, evalCtx *tree.EvalContext, col coldata.Vec, rowIdx int,
) (tree.Datum, error) {
	__antithesis_instrumentation__.Notify(496551)
	if col.Nulls().NullAt(rowIdx) {
		__antithesis_instrumentation__.Notify(496554)
		return tree.DNull, nil
	} else {
		__antithesis_instrumentation__.Notify(496555)
	}
	__antithesis_instrumentation__.Notify(496552)
	switch t := col.Type(); col.CanonicalTypeFamily() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(496556)
		return tree.MakeDBool(tree.DBool(col.Bool()[rowIdx])), nil
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(496557)
		switch t.Width() {
		case 0, 64:
			__antithesis_instrumentation__.Notify(496561)
			switch hint.Family() {
			case types.IntFamily:
				__antithesis_instrumentation__.Notify(496564)
				return alloc.NewDInt(tree.DInt(col.Int64()[rowIdx])), nil
			case types.DecimalFamily:
				__antithesis_instrumentation__.Notify(496565)
				d := *apd.New(col.Int64()[rowIdx], 0)
				return alloc.NewDDecimal(tree.DDecimal{Decimal: d}), nil
			case types.DateFamily:
				__antithesis_instrumentation__.Notify(496566)
				date, err := pgdate.MakeDateFromUnixEpoch(col.Int64()[rowIdx])
				if err != nil {
					__antithesis_instrumentation__.Notify(496569)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(496570)
				}
				__antithesis_instrumentation__.Notify(496567)
				return alloc.NewDDate(tree.DDate{Date: date}), nil
			default:
				__antithesis_instrumentation__.Notify(496568)
			}
		case 16:
			__antithesis_instrumentation__.Notify(496562)
			switch hint.Family() {
			case types.IntFamily:
				__antithesis_instrumentation__.Notify(496571)
				return alloc.NewDInt(tree.DInt(col.Int16()[rowIdx])), nil
			default:
				__antithesis_instrumentation__.Notify(496572)
			}
		default:
			__antithesis_instrumentation__.Notify(496563)
		}
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(496558)
		switch hint.Family() {
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(496573)
			return alloc.NewDFloat(tree.DFloat(col.Float64()[rowIdx])), nil
		case types.DecimalFamily:
			__antithesis_instrumentation__.Notify(496574)
			var d apd.Decimal
			if _, err := d.SetFloat64(col.Float64()[rowIdx]); err != nil {
				__antithesis_instrumentation__.Notify(496577)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(496578)
			}
			__antithesis_instrumentation__.Notify(496575)
			return alloc.NewDDecimal(tree.DDecimal{Decimal: d}), nil
		default:
			__antithesis_instrumentation__.Notify(496576)
		}
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(496559)
		switch hint.Family() {
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(496579)
			return alloc.NewDBytes(tree.DBytes(col.Bytes().Get(rowIdx))), nil
		case types.StringFamily:
			__antithesis_instrumentation__.Notify(496580)
			data := col.Bytes().Get(rowIdx)
			str := *(*string)(unsafe.Pointer(&data))
			return alloc.NewDString(tree.DString(str)), nil
		default:
			__antithesis_instrumentation__.Notify(496581)
			data := col.Bytes().Get(rowIdx)
			str := *(*string)(unsafe.Pointer(&data))
			return rowenc.ParseDatumStringAs(hint, str, evalCtx)
		}
	default:
		__antithesis_instrumentation__.Notify(496560)
	}
	__antithesis_instrumentation__.Notify(496553)
	return nil, errors.Errorf(
		`don't know how to interpret %s column as %s`, col.Type(), hint)
}

func (w *workloadReader) readFiles(
	ctx context.Context,
	dataFiles map[int32]string,
	_ map[int32]int64,
	_ roachpb.IOFileFormat,
	_ cloud.ExternalStorageFactory,
	_ security.SQLUsername,
) error {
	__antithesis_instrumentation__.Notify(496582)

	wcs := make([]*WorkloadKVConverter, 0, len(dataFiles))
	for fileID, fileName := range dataFiles {
		__antithesis_instrumentation__.Notify(496585)
		conf, err := parseWorkloadConfig(fileName)
		if err != nil {
			__antithesis_instrumentation__.Notify(496592)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496593)
		}
		__antithesis_instrumentation__.Notify(496586)
		meta, err := workload.Get(conf.Generator)
		if err != nil {
			__antithesis_instrumentation__.Notify(496594)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496595)
		}
		__antithesis_instrumentation__.Notify(496587)

		if meta.Version != conf.Version {
			__antithesis_instrumentation__.Notify(496596)
			return errors.Errorf(
				`expected %s version "%s" but got "%s"`, meta.Name, conf.Version, meta.Version)
		} else {
			__antithesis_instrumentation__.Notify(496597)
		}
		__antithesis_instrumentation__.Notify(496588)
		gen := meta.New()
		if f, ok := gen.(workload.Flagser); ok {
			__antithesis_instrumentation__.Notify(496598)
			flags := f.Flags()
			if err := flags.Parse(conf.Flags); err != nil {
				__antithesis_instrumentation__.Notify(496599)
				return errors.Wrapf(err, `parsing parameters %s`, strings.Join(conf.Flags, ` `))
			} else {
				__antithesis_instrumentation__.Notify(496600)
			}
		} else {
			__antithesis_instrumentation__.Notify(496601)
		}
		__antithesis_instrumentation__.Notify(496589)
		var t workload.Table
		for _, tbl := range gen.Tables() {
			__antithesis_instrumentation__.Notify(496602)
			if tbl.Name == conf.Table {
				__antithesis_instrumentation__.Notify(496603)
				t = tbl
				break
			} else {
				__antithesis_instrumentation__.Notify(496604)
			}
		}
		__antithesis_instrumentation__.Notify(496590)
		if t.Name == `` {
			__antithesis_instrumentation__.Notify(496605)
			return errors.Errorf(`unknown table %s for generator %s`, conf.Table, meta.Name)
		} else {
			__antithesis_instrumentation__.Notify(496606)
		}
		__antithesis_instrumentation__.Notify(496591)

		wc := NewWorkloadKVConverter(
			fileID, w.table, t.InitialRows, int(conf.BatchBegin), int(conf.BatchEnd), w.kvCh)
		wcs = append(wcs, wc)
	}
	__antithesis_instrumentation__.Notify(496583)

	for _, wc := range wcs {
		__antithesis_instrumentation__.Notify(496607)
		if err := ctxgroup.GroupWorkers(ctx, w.parallelism, func(ctx context.Context, _ int) error {
			__antithesis_instrumentation__.Notify(496608)
			evalCtx := w.evalCtx.Copy()
			return wc.Worker(ctx, evalCtx, w.semaCtx)
		}); err != nil {
			__antithesis_instrumentation__.Notify(496609)
			return err
		} else {
			__antithesis_instrumentation__.Notify(496610)
		}
	}
	__antithesis_instrumentation__.Notify(496584)
	return nil
}

type WorkloadKVConverter struct {
	tableDesc      catalog.TableDescriptor
	rows           workload.BatchedTuples
	batchIdxAtomic int64
	batchEnd       int
	kvCh           chan row.KVBatch

	fileID                int32
	totalBatches          float32
	finishedBatchesAtomic int64
}

func NewWorkloadKVConverter(
	fileID int32,
	tableDesc catalog.TableDescriptor,
	rows workload.BatchedTuples,
	batchStart, batchEnd int,
	kvCh chan row.KVBatch,
) *WorkloadKVConverter {
	__antithesis_instrumentation__.Notify(496611)
	return &WorkloadKVConverter{
		tableDesc:      tableDesc,
		rows:           rows,
		batchIdxAtomic: int64(batchStart) - 1,
		batchEnd:       batchEnd,
		kvCh:           kvCh,
		totalBatches:   float32(batchEnd - batchStart),
		fileID:         fileID,
	}
}

func (w *WorkloadKVConverter) Worker(
	ctx context.Context, evalCtx *tree.EvalContext, semaCtx *tree.SemaContext,
) error {
	__antithesis_instrumentation__.Notify(496612)
	conv, err := row.NewDatumRowConverter(ctx, semaCtx, w.tableDesc, nil,
		evalCtx, w.kvCh, nil, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(496616)
		return err
	} else {
		__antithesis_instrumentation__.Notify(496617)
	}
	__antithesis_instrumentation__.Notify(496613)
	conv.KvBatch.Source = w.fileID
	conv.FractionFn = func() float32 {
		__antithesis_instrumentation__.Notify(496618)
		return float32(atomic.LoadInt64(&w.finishedBatchesAtomic)) / w.totalBatches
	}
	__antithesis_instrumentation__.Notify(496614)
	var alloc tree.DatumAlloc
	var a bufalloc.ByteAllocator
	cb := coldata.NewMemBatchWithCapacity(nil, 0, coldata.StandardColumnFactory)

	for {
		__antithesis_instrumentation__.Notify(496619)
		batchIdx := int(atomic.AddInt64(&w.batchIdxAtomic, 1))
		if batchIdx >= w.batchEnd {
			__antithesis_instrumentation__.Notify(496622)
			break
		} else {
			__antithesis_instrumentation__.Notify(496623)
		}
		__antithesis_instrumentation__.Notify(496620)
		a = a.Truncate()
		w.rows.FillBatch(batchIdx, cb, &a)
		for rowIdx, numRows := 0, cb.Length(); rowIdx < numRows; rowIdx++ {
			__antithesis_instrumentation__.Notify(496624)
			for colIdx, col := range cb.ColVecs() {
				__antithesis_instrumentation__.Notify(496626)

				converted, err := makeDatumFromColOffset(
					&alloc, conv.VisibleColTypes[colIdx], evalCtx, col, rowIdx)
				if err != nil {
					__antithesis_instrumentation__.Notify(496628)
					return err
				} else {
					__antithesis_instrumentation__.Notify(496629)
				}
				__antithesis_instrumentation__.Notify(496627)
				conv.Datums[colIdx] = converted
			}
			__antithesis_instrumentation__.Notify(496625)

			fileIdx, timestamp := int32(batchIdx), int64(rowIdx)
			if err := conv.Row(ctx, fileIdx, timestamp); err != nil {
				__antithesis_instrumentation__.Notify(496630)
				return err
			} else {
				__antithesis_instrumentation__.Notify(496631)
			}
		}
		__antithesis_instrumentation__.Notify(496621)
		atomic.AddInt64(&w.finishedBatchesAtomic, 1)
	}
	__antithesis_instrumentation__.Notify(496615)
	return conv.SendBatch(ctx)
}

var errNotWorkloadURI = errors.New("not a workload URI")

func parseWorkloadConfig(fileName string) (workloadConfig, error) {
	__antithesis_instrumentation__.Notify(496632)
	c := workloadConfig{}

	uri, err := url.Parse(fileName)
	if err != nil {
		__antithesis_instrumentation__.Notify(496640)
		return c, err
	} else {
		__antithesis_instrumentation__.Notify(496641)
	}
	__antithesis_instrumentation__.Notify(496633)

	if uri.Scheme != "workload" {
		__antithesis_instrumentation__.Notify(496642)
		return c, errNotWorkloadURI
	} else {
		__antithesis_instrumentation__.Notify(496643)
	}
	__antithesis_instrumentation__.Notify(496634)
	pathParts := strings.Split(strings.Trim(uri.Path, `/`), `/`)
	if len(pathParts) != 3 {
		__antithesis_instrumentation__.Notify(496644)
		return c, errors.Errorf(
			`path must be of the form /<format>/<generator>/<table>: %s`, uri.Path)
	} else {
		__antithesis_instrumentation__.Notify(496645)
	}
	__antithesis_instrumentation__.Notify(496635)
	c.Format, c.Generator, c.Table = pathParts[0], pathParts[1], pathParts[2]
	q := uri.Query()
	if _, ok := q[`version`]; !ok {
		__antithesis_instrumentation__.Notify(496646)
		return c, errors.New(`parameter version is required`)
	} else {
		__antithesis_instrumentation__.Notify(496647)
	}
	__antithesis_instrumentation__.Notify(496636)
	c.Version = q.Get(`version`)
	q.Del(`version`)
	if s := q.Get(`row-start`); len(s) > 0 {
		__antithesis_instrumentation__.Notify(496648)
		q.Del(`row-start`)
		var err error
		if c.BatchBegin, err = strconv.ParseInt(s, 10, 64); err != nil {
			__antithesis_instrumentation__.Notify(496649)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(496650)
		}
	} else {
		__antithesis_instrumentation__.Notify(496651)
	}
	__antithesis_instrumentation__.Notify(496637)
	if e := q.Get(`row-end`); len(e) > 0 {
		__antithesis_instrumentation__.Notify(496652)
		q.Del(`row-end`)
		var err error
		if c.BatchEnd, err = strconv.ParseInt(e, 10, 64); err != nil {
			__antithesis_instrumentation__.Notify(496653)
			return c, err
		} else {
			__antithesis_instrumentation__.Notify(496654)
		}
	} else {
		__antithesis_instrumentation__.Notify(496655)
	}
	__antithesis_instrumentation__.Notify(496638)
	for k, vs := range q {
		__antithesis_instrumentation__.Notify(496656)
		for _, v := range vs {
			__antithesis_instrumentation__.Notify(496657)
			c.Flags = append(c.Flags, `--`+k+`=`+v)
		}
	}
	__antithesis_instrumentation__.Notify(496639)
	return c, nil
}

type workloadConfig struct {
	Generator  string
	Version    string
	Table      string
	Flags      []string
	Format     string
	BatchBegin int64
	BatchEnd   int64
}
