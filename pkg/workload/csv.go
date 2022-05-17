package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/encoding/csv"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const (
	rowStartParam = `row-start`
	rowEndParam   = `row-end`
)

func WriteCSVRows(
	ctx context.Context, w io.Writer, table Table, rowStart, rowEnd int, sizeBytesLimit int64,
) (rowBatchIdx int, err error) {
	__antithesis_instrumentation__.Notify(693943)
	cb := coldata.NewMemBatchWithCapacity(nil, 0, coldata.StandardColumnFactory)
	var a bufalloc.ByteAllocator

	bytesWrittenW := &bytesWrittenWriter{w: w}
	csvW := csv.NewWriter(bytesWrittenW)
	var rowStrings []string
	for rowBatchIdx = rowStart; rowBatchIdx < rowEnd; rowBatchIdx++ {
		__antithesis_instrumentation__.Notify(693945)
		if sizeBytesLimit > 0 && func() bool {
			__antithesis_instrumentation__.Notify(693949)
			return bytesWrittenW.written > sizeBytesLimit == true
		}() == true {
			__antithesis_instrumentation__.Notify(693950)
			break
		} else {
			__antithesis_instrumentation__.Notify(693951)
		}
		__antithesis_instrumentation__.Notify(693946)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(693952)
			return 0, ctx.Err()
		default:
			__antithesis_instrumentation__.Notify(693953)
		}
		__antithesis_instrumentation__.Notify(693947)
		a = a.Truncate()
		table.InitialRows.FillBatch(rowBatchIdx, cb, &a)
		if numCols := cb.Width(); cap(rowStrings) < numCols {
			__antithesis_instrumentation__.Notify(693954)
			rowStrings = make([]string, numCols)
		} else {
			__antithesis_instrumentation__.Notify(693955)
			rowStrings = rowStrings[:numCols]
		}
		__antithesis_instrumentation__.Notify(693948)
		for rowIdx, numRows := 0, cb.Length(); rowIdx < numRows; rowIdx++ {
			__antithesis_instrumentation__.Notify(693956)
			for colIdx, col := range cb.ColVecs() {
				__antithesis_instrumentation__.Notify(693958)
				rowStrings[colIdx] = colDatumToCSVString(col, rowIdx)
			}
			__antithesis_instrumentation__.Notify(693957)
			if err := csvW.Write(rowStrings); err != nil {
				__antithesis_instrumentation__.Notify(693959)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(693960)
			}
		}
	}
	__antithesis_instrumentation__.Notify(693944)
	csvW.Flush()
	return rowBatchIdx, csvW.Error()
}

type csvRowsReader struct {
	t                    Table
	batchStart, batchEnd int

	buf  bytes.Buffer
	csvW *csv.Writer

	batchIdx int
	cb       coldata.Batch
	a        bufalloc.ByteAllocator

	stringsBuf []string
}

func (r *csvRowsReader) Read(p []byte) (n int, err error) {
	__antithesis_instrumentation__.Notify(693961)
	if r.cb == nil {
		__antithesis_instrumentation__.Notify(693963)
		r.cb = coldata.NewMemBatchWithCapacity(nil, 0, coldata.StandardColumnFactory)
	} else {
		__antithesis_instrumentation__.Notify(693964)
	}
	__antithesis_instrumentation__.Notify(693962)

	for {
		__antithesis_instrumentation__.Notify(693965)
		if r.buf.Len() > 0 {
			__antithesis_instrumentation__.Notify(693970)
			return r.buf.Read(p)
		} else {
			__antithesis_instrumentation__.Notify(693971)
		}
		__antithesis_instrumentation__.Notify(693966)
		r.buf.Reset()
		if r.batchIdx == r.batchEnd {
			__antithesis_instrumentation__.Notify(693972)
			return 0, io.EOF
		} else {
			__antithesis_instrumentation__.Notify(693973)
		}
		__antithesis_instrumentation__.Notify(693967)
		r.a = r.a.Truncate()
		r.t.InitialRows.FillBatch(r.batchIdx, r.cb, &r.a)
		r.batchIdx++
		if numCols := r.cb.Width(); cap(r.stringsBuf) < numCols {
			__antithesis_instrumentation__.Notify(693974)
			r.stringsBuf = make([]string, numCols)
		} else {
			__antithesis_instrumentation__.Notify(693975)
			r.stringsBuf = r.stringsBuf[:numCols]
		}
		__antithesis_instrumentation__.Notify(693968)
		for rowIdx, numRows := 0, r.cb.Length(); rowIdx < numRows; rowIdx++ {
			__antithesis_instrumentation__.Notify(693976)
			for colIdx, col := range r.cb.ColVecs() {
				__antithesis_instrumentation__.Notify(693978)
				r.stringsBuf[colIdx] = colDatumToCSVString(col, rowIdx)
			}
			__antithesis_instrumentation__.Notify(693977)
			if err := r.csvW.Write(r.stringsBuf); err != nil {
				__antithesis_instrumentation__.Notify(693979)
				return 0, err
			} else {
				__antithesis_instrumentation__.Notify(693980)
			}
		}
		__antithesis_instrumentation__.Notify(693969)
		r.csvW.Flush()
	}
}

func NewCSVRowsReader(t Table, batchStart, batchEnd int) io.Reader {
	__antithesis_instrumentation__.Notify(693981)
	if batchEnd == 0 {
		__antithesis_instrumentation__.Notify(693983)
		batchEnd = t.InitialRows.NumBatches
	} else {
		__antithesis_instrumentation__.Notify(693984)
	}
	__antithesis_instrumentation__.Notify(693982)
	r := &csvRowsReader{t: t, batchStart: batchStart, batchEnd: batchEnd, batchIdx: batchStart}
	r.csvW = csv.NewWriter(&r.buf)
	return r
}

func colDatumToCSVString(col coldata.Vec, rowIdx int) string {
	__antithesis_instrumentation__.Notify(693985)
	if col.Nulls().NullAt(rowIdx) {
		__antithesis_instrumentation__.Notify(693988)
		return `NULL`
	} else {
		__antithesis_instrumentation__.Notify(693989)
	}
	__antithesis_instrumentation__.Notify(693986)
	switch col.CanonicalTypeFamily() {
	case types.BoolFamily:
		__antithesis_instrumentation__.Notify(693990)
		return strconv.FormatBool(col.Bool()[rowIdx])
	case types.IntFamily:
		__antithesis_instrumentation__.Notify(693991)
		return strconv.FormatInt(col.Int64()[rowIdx], 10)
	case types.FloatFamily:
		__antithesis_instrumentation__.Notify(693992)
		return strconv.FormatFloat(col.Float64()[rowIdx], 'f', -1, 64)
	case types.BytesFamily:
		__antithesis_instrumentation__.Notify(693993)

		bytes := col.Bytes().Get(rowIdx)
		return *(*string)(unsafe.Pointer(&bytes))
	default:
		__antithesis_instrumentation__.Notify(693994)
	}
	__antithesis_instrumentation__.Notify(693987)
	panic(fmt.Sprintf(`unhandled type %s`, col.Type()))
}

func HandleCSV(w http.ResponseWriter, req *http.Request, prefix string, meta Meta) error {
	__antithesis_instrumentation__.Notify(693995)
	ctx := context.Background()
	if err := req.ParseForm(); err != nil {
		__antithesis_instrumentation__.Notify(694003)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694004)
	}
	__antithesis_instrumentation__.Notify(693996)

	gen := meta.New()
	if f, ok := gen.(Flagser); ok {
		__antithesis_instrumentation__.Notify(694005)
		var flags []string
		f.Flags().VisitAll(func(f *pflag.Flag) {
			__antithesis_instrumentation__.Notify(694007)
			if vals, ok := req.Form[f.Name]; ok {
				__antithesis_instrumentation__.Notify(694008)
				for _, val := range vals {
					__antithesis_instrumentation__.Notify(694009)
					flags = append(flags, fmt.Sprintf(`--%s=%s`, f.Name, val))
				}
			} else {
				__antithesis_instrumentation__.Notify(694010)
			}
		})
		__antithesis_instrumentation__.Notify(694006)
		if err := f.Flags().Parse(flags); err != nil {
			__antithesis_instrumentation__.Notify(694011)
			return errors.Wrapf(err, `parsing parameters %s`, strings.Join(flags, ` `))
		} else {
			__antithesis_instrumentation__.Notify(694012)
		}
	} else {
		__antithesis_instrumentation__.Notify(694013)
	}
	__antithesis_instrumentation__.Notify(693997)

	tableName := strings.TrimPrefix(req.URL.Path, prefix)
	var table *Table
	for _, t := range gen.Tables() {
		__antithesis_instrumentation__.Notify(694014)
		if t.Name == tableName {
			__antithesis_instrumentation__.Notify(694015)
			table = &t
			break
		} else {
			__antithesis_instrumentation__.Notify(694016)
		}
	}
	__antithesis_instrumentation__.Notify(693998)
	if table == nil {
		__antithesis_instrumentation__.Notify(694017)
		return errors.Errorf(`could not find table %s in generator %s`, tableName, meta.Name)
	} else {
		__antithesis_instrumentation__.Notify(694018)
	}
	__antithesis_instrumentation__.Notify(693999)
	if table.InitialRows.FillBatch == nil {
		__antithesis_instrumentation__.Notify(694019)
		return errors.Errorf(`csv-server is not supported for workload %s`, meta.Name)
	} else {
		__antithesis_instrumentation__.Notify(694020)
	}
	__antithesis_instrumentation__.Notify(694000)

	rowStart, rowEnd := 0, table.InitialRows.NumBatches
	if vals, ok := req.Form[rowStartParam]; ok && func() bool {
		__antithesis_instrumentation__.Notify(694021)
		return len(vals) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(694022)
		var err error
		rowStart, err = strconv.Atoi(vals[len(vals)-1])
		if err != nil {
			__antithesis_instrumentation__.Notify(694023)
			return errors.Wrapf(err, `parsing %s`, rowStartParam)
		} else {
			__antithesis_instrumentation__.Notify(694024)
		}
	} else {
		__antithesis_instrumentation__.Notify(694025)
	}
	__antithesis_instrumentation__.Notify(694001)
	if vals, ok := req.Form[rowEndParam]; ok && func() bool {
		__antithesis_instrumentation__.Notify(694026)
		return len(vals) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(694027)
		var err error
		rowEnd, err = strconv.Atoi(vals[len(vals)-1])
		if err != nil {
			__antithesis_instrumentation__.Notify(694028)
			return errors.Wrapf(err, `parsing %s`, rowEndParam)
		} else {
			__antithesis_instrumentation__.Notify(694029)
		}
	} else {
		__antithesis_instrumentation__.Notify(694030)
	}
	__antithesis_instrumentation__.Notify(694002)

	w.Header().Set(`Content-Type`, `text/csv`)
	_, err := WriteCSVRows(ctx, w, *table, rowStart, rowEnd, -1)
	return err
}

type bytesWrittenWriter struct {
	w       io.Writer
	written int64
}

func (w *bytesWrittenWriter) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(694031)
	n, err := w.w.Write(p)
	w.written += int64(n)
	return n, err
}

func CSVMux(metas []Meta) *http.ServeMux {
	__antithesis_instrumentation__.Notify(694032)
	mux := http.NewServeMux()
	for _, meta := range metas {
		__antithesis_instrumentation__.Notify(694034)
		meta := meta
		prefix := fmt.Sprintf(`/csv/%s/`, meta.Name)
		mux.HandleFunc(prefix, func(w http.ResponseWriter, req *http.Request) {
			__antithesis_instrumentation__.Notify(694035)
			if err := HandleCSV(w, req, prefix, meta); err != nil {
				__antithesis_instrumentation__.Notify(694036)
				http.Error(w, err.Error(), http.StatusInternalServerError)
			} else {
				__antithesis_instrumentation__.Notify(694037)
			}
		})
	}
	__antithesis_instrumentation__.Notify(694033)
	return mux
}
