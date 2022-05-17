// Package workload provides an abstraction for generators of sql query loads
// (and requisite initial data) as well as tools for working with these
// generators.
package workload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"math/bits"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/col/typeconv"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

type Generator interface {
	Meta() Meta

	Tables() []Table
}

type FlagMeta struct {
	RuntimeOnly bool

	CheckConsistencyOnly bool
}

type Flags struct {
	*pflag.FlagSet

	Meta map[string]FlagMeta
}

type Flagser interface {
	Generator
	Flags() Flags
}

type Opser interface {
	Generator

	Ops(ctx context.Context, urls []string, reg *histogram.Registry) (QueryLoad, error)
}

type Hookser interface {
	Generator
	Hooks() Hooks
}

type Hooks struct {
	Validate func() error

	PreCreate func(*gosql.DB) error

	PreLoad func(*gosql.DB) error

	PostLoad func(*gosql.DB) error

	PostRun func(time.Duration) error

	CheckConsistency func(context.Context, *gosql.DB) error

	Partition func(*gosql.DB) error
}

type Meta struct {
	Name string

	Description string

	Details string

	Version string

	PublicFacing bool

	New func() Generator
}

type Table struct {
	Name string

	Schema string

	InitialRows BatchedTuples

	Splits BatchedTuples

	Stats []JSONStatistic
}

type BatchedTuples struct {
	NumBatches int

	FillBatch func(int, coldata.Batch, *bufalloc.ByteAllocator)
}

func Tuples(count int, fn func(int) []interface{}) BatchedTuples {
	__antithesis_instrumentation__.Notify(698965)
	return TypedTuples(count, nil, fn)
}

const (
	timestampOutputFormat = "2006-01-02 15:04:05.999999-07:00"
)

func TypedTuples(count int, typs []*types.T, fn func(int) []interface{}) BatchedTuples {
	__antithesis_instrumentation__.Notify(698966)

	var typesOnce sync.Once

	t := BatchedTuples{
		NumBatches: count,
	}
	if fn != nil {
		__antithesis_instrumentation__.Notify(698968)
		t.FillBatch = func(batchIdx int, cb coldata.Batch, _ *bufalloc.ByteAllocator) {
			__antithesis_instrumentation__.Notify(698969)
			row := fn(batchIdx)

			typesOnce.Do(func() {
				__antithesis_instrumentation__.Notify(698971)
				if typs == nil {
					__antithesis_instrumentation__.Notify(698972)
					typs = make([]*types.T, len(row))
					for i, datum := range row {
						__antithesis_instrumentation__.Notify(698973)
						if datum == nil {
							__antithesis_instrumentation__.Notify(698974)
							panic(fmt.Sprintf(
								`can't determine type of nil column; call TypedTuples directly: %v`, row))
						} else {
							__antithesis_instrumentation__.Notify(698975)
							switch datum.(type) {
							case time.Time:
								__antithesis_instrumentation__.Notify(698976)
								typs[i] = types.Bytes
							default:
								__antithesis_instrumentation__.Notify(698977)
								typs[i] = typeconv.UnsafeFromGoType(datum)
							}
						}
					}
				} else {
					__antithesis_instrumentation__.Notify(698978)
				}
			})
			__antithesis_instrumentation__.Notify(698970)

			cb.Reset(typs, 1, coldata.StandardColumnFactory)
			for colIdx, col := range cb.ColVecs() {
				__antithesis_instrumentation__.Notify(698979)
				switch d := row[colIdx].(type) {
				case nil:
					__antithesis_instrumentation__.Notify(698980)
					col.Nulls().SetNull(0)
				case bool:
					__antithesis_instrumentation__.Notify(698981)
					col.Bool()[0] = d
				case int:
					__antithesis_instrumentation__.Notify(698982)
					col.Int64()[0] = int64(d)
				case float64:
					__antithesis_instrumentation__.Notify(698983)
					col.Float64()[0] = d
				case string:
					__antithesis_instrumentation__.Notify(698984)
					col.Bytes().Set(0, []byte(d))
				case []byte:
					__antithesis_instrumentation__.Notify(698985)
					col.Bytes().Set(0, d)
				case time.Time:
					__antithesis_instrumentation__.Notify(698986)
					col.Bytes().Set(0, []byte(d.Round(time.Microsecond).UTC().Format(timestampOutputFormat)))
				default:
					__antithesis_instrumentation__.Notify(698987)
					panic(errors.AssertionFailedf(`unhandled datum type %T`, d))
				}
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(698988)
	}
	__antithesis_instrumentation__.Notify(698967)
	return t
}

func (b BatchedTuples) BatchRows(batchIdx int) [][]interface{} {
	__antithesis_instrumentation__.Notify(698989)
	cb := coldata.NewMemBatchWithCapacity(nil, 0, coldata.StandardColumnFactory)
	var a bufalloc.ByteAllocator
	b.FillBatch(batchIdx, cb, &a)
	return ColBatchToRows(cb)
}

func ColBatchToRows(cb coldata.Batch) [][]interface{} {
	__antithesis_instrumentation__.Notify(698990)
	numRows, numCols := cb.Length(), cb.Width()

	datums := make([]interface{}, numRows*numCols)
	for colIdx, col := range cb.ColVecs() {
		__antithesis_instrumentation__.Notify(698993)
		nulls := col.Nulls()
		switch col.CanonicalTypeFamily() {
		case types.BoolFamily:
			__antithesis_instrumentation__.Notify(698994)
			for rowIdx, datum := range col.Bool()[:numRows] {
				__antithesis_instrumentation__.Notify(698999)
				if !nulls.NullAt(rowIdx) {
					__antithesis_instrumentation__.Notify(699000)
					datums[rowIdx*numCols+colIdx] = datum
				} else {
					__antithesis_instrumentation__.Notify(699001)
				}
			}
		case types.IntFamily:
			__antithesis_instrumentation__.Notify(698995)
			switch col.Type().Width() {
			case 0, 64:
				__antithesis_instrumentation__.Notify(699002)
				for rowIdx, datum := range col.Int64()[:numRows] {
					__antithesis_instrumentation__.Notify(699005)
					if !nulls.NullAt(rowIdx) {
						__antithesis_instrumentation__.Notify(699006)
						datums[rowIdx*numCols+colIdx] = datum
					} else {
						__antithesis_instrumentation__.Notify(699007)
					}
				}
			case 16:
				__antithesis_instrumentation__.Notify(699003)
				for rowIdx, datum := range col.Int16()[:numRows] {
					__antithesis_instrumentation__.Notify(699008)
					if !nulls.NullAt(rowIdx) {
						__antithesis_instrumentation__.Notify(699009)
						datums[rowIdx*numCols+colIdx] = datum
					} else {
						__antithesis_instrumentation__.Notify(699010)
					}
				}
			default:
				__antithesis_instrumentation__.Notify(699004)
				panic(fmt.Sprintf(`unhandled type %s`, col.Type()))
			}
		case types.FloatFamily:
			__antithesis_instrumentation__.Notify(698996)
			for rowIdx, datum := range col.Float64()[:numRows] {
				__antithesis_instrumentation__.Notify(699011)
				if !nulls.NullAt(rowIdx) {
					__antithesis_instrumentation__.Notify(699012)
					datums[rowIdx*numCols+colIdx] = datum
				} else {
					__antithesis_instrumentation__.Notify(699013)
				}
			}
		case types.BytesFamily:
			__antithesis_instrumentation__.Notify(698997)

			colBytes := col.Bytes()
			for rowIdx := 0; rowIdx < numRows; rowIdx++ {
				__antithesis_instrumentation__.Notify(699014)
				if !nulls.NullAt(rowIdx) {
					__antithesis_instrumentation__.Notify(699015)
					datums[rowIdx*numCols+colIdx] = colBytes.Get(rowIdx)
				} else {
					__antithesis_instrumentation__.Notify(699016)
				}
			}
		default:
			__antithesis_instrumentation__.Notify(698998)
			panic(fmt.Sprintf(`unhandled type %s`, col.Type()))
		}
	}
	__antithesis_instrumentation__.Notify(698991)
	rows := make([][]interface{}, numRows)
	for rowIdx := 0; rowIdx < numRows; rowIdx++ {
		__antithesis_instrumentation__.Notify(699017)
		rows[rowIdx] = datums[rowIdx*numCols : (rowIdx+1)*numCols]
	}
	__antithesis_instrumentation__.Notify(698992)
	return rows
}

type InitialDataLoader interface {
	InitialDataLoad(context.Context, *gosql.DB, Generator) (int64, error)
}

var ImportDataLoader InitialDataLoader = requiresCCLBinaryDataLoader(`IMPORT`)

type requiresCCLBinaryDataLoader string

func (l requiresCCLBinaryDataLoader) InitialDataLoad(
	context.Context, *gosql.DB, Generator,
) (int64, error) {
	__antithesis_instrumentation__.Notify(699018)
	return 0, errors.Errorf(`loading initial data with %s requires a CCL binary`, l)
}

type QueryLoad struct {
	SQLDatabase string

	WorkerFns []func(context.Context) error

	Close func(context.Context)

	ResultHist string
}

var registered = make(map[string]Meta)

func Register(m Meta) {
	__antithesis_instrumentation__.Notify(699019)
	if _, ok := registered[m.Name]; ok {
		__antithesis_instrumentation__.Notify(699021)
		panic(m.Name + " is already registered")
	} else {
		__antithesis_instrumentation__.Notify(699022)
	}
	__antithesis_instrumentation__.Notify(699020)
	registered[m.Name] = m
}

func Get(name string) (Meta, error) {
	__antithesis_instrumentation__.Notify(699023)
	m, ok := registered[name]
	if !ok {
		__antithesis_instrumentation__.Notify(699025)
		return Meta{}, errors.Errorf("unknown generator: %s", name)
	} else {
		__antithesis_instrumentation__.Notify(699026)
	}
	__antithesis_instrumentation__.Notify(699024)
	return m, nil
}

func Registered() []Meta {
	__antithesis_instrumentation__.Notify(699027)
	gens := make([]Meta, 0, len(registered))
	for _, gen := range registered {
		__antithesis_instrumentation__.Notify(699030)
		gens = append(gens, gen)
	}
	__antithesis_instrumentation__.Notify(699028)
	sort.Slice(gens, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(699031)
		return strings.Compare(gens[i].Name, gens[j].Name) < 0
	})
	__antithesis_instrumentation__.Notify(699029)
	return gens
}

func FromFlags(meta Meta, flags ...string) Generator {
	__antithesis_instrumentation__.Notify(699032)
	gen := meta.New()
	if len(flags) > 0 {
		__antithesis_instrumentation__.Notify(699035)
		f, ok := gen.(Flagser)
		if !ok {
			__antithesis_instrumentation__.Notify(699037)
			panic(fmt.Sprintf(`generator %s does not accept flags: %v`, meta.Name, flags))
		} else {
			__antithesis_instrumentation__.Notify(699038)
		}
		__antithesis_instrumentation__.Notify(699036)
		flagsStruct := f.Flags()
		if err := flagsStruct.Parse(flags); err != nil {
			__antithesis_instrumentation__.Notify(699039)
			panic(fmt.Sprintf(`generator %s parsing flags %v: %v`, meta.Name, flags, err))
		} else {
			__antithesis_instrumentation__.Notify(699040)
		}
	} else {
		__antithesis_instrumentation__.Notify(699041)
	}
	__antithesis_instrumentation__.Notify(699033)
	if h, ok := gen.(Hookser); ok {
		__antithesis_instrumentation__.Notify(699042)
		if err := h.Hooks().Validate(); err != nil {
			__antithesis_instrumentation__.Notify(699043)
			panic(fmt.Sprintf(`generator %s flags %s did not validate: %v`, meta.Name, flags, err))
		} else {
			__antithesis_instrumentation__.Notify(699044)
		}
	} else {
		__antithesis_instrumentation__.Notify(699045)
	}
	__antithesis_instrumentation__.Notify(699034)
	return gen
}

func ApproxDatumSize(x interface{}) int64 {
	__antithesis_instrumentation__.Notify(699046)
	if x == nil {
		__antithesis_instrumentation__.Notify(699048)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(699049)
	}
	__antithesis_instrumentation__.Notify(699047)
	switch t := x.(type) {
	case bool:
		__antithesis_instrumentation__.Notify(699050)
		return 1
	case int:
		__antithesis_instrumentation__.Notify(699051)
		if t < 0 {
			__antithesis_instrumentation__.Notify(699061)
			t = -t
		} else {
			__antithesis_instrumentation__.Notify(699062)
		}
		__antithesis_instrumentation__.Notify(699052)

		return int64(bits.Len(uint(t))+8) / 8
	case int64:
		__antithesis_instrumentation__.Notify(699053)
		return int64(bits.Len64(uint64(t))+8) / 8
	case int16:
		__antithesis_instrumentation__.Notify(699054)
		return int64(bits.Len64(uint64(t))+8) / 8
	case uint64:
		__antithesis_instrumentation__.Notify(699055)
		return int64(bits.Len64(t)+8) / 8
	case float64:
		__antithesis_instrumentation__.Notify(699056)
		return int64(bits.Len64(math.Float64bits(t))+8) / 8
	case string:
		__antithesis_instrumentation__.Notify(699057)
		return int64(len(t))
	case []byte:
		__antithesis_instrumentation__.Notify(699058)
		return int64(len(t))
	case time.Time:
		__antithesis_instrumentation__.Notify(699059)
		return 12
	default:
		__antithesis_instrumentation__.Notify(699060)
		panic(errors.AssertionFailedf("unsupported type %T: %v", x, x))
	}
}
