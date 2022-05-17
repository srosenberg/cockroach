/*
Package bulkingest defines a workload that is intended to stress some edge cases
in our bulk-ingestion infrastructure.

In both IMPORT and indexing, many readers scan though the source data (i.e. CSV
files or PK rows, respectively) and produce KVs to be ingested. However a given
range of that source data could produce any KVs -- i.e. in some schemas or
workloads, the produced KVs could have the same ordering or in some they could
be random and uniformly distributed in the keyspace. Additionally, both of the
processes often include concurrent producers, each scanning their own input
files or ranges of a table, and there the distribution could mean that
concurrent producers all produce different keys or all produce similar keys at
the same time, etc.

This workload is intended to produce testdata that emphasizes these cases. The
multi-column PK is intended to make it easy to independently control the prefix
of keys. Adding an index on the same columns with the columns reordered can then
control the flow of keys between prefixes, stressing any buffering, sorting or
other steps in the middle. This can be particularly interesting when concurrent
producers are a factor, as the distribution (or lack there of) of their output
prefixes at a given moment can cause hotspots.

The workload's schema is a table with columns a, b, and c plus a padding payload
string, with the primary key being (a,b,c).

Creating indexes on the different columns in this schema can then trigger
different distributions of produced index KVs -- i.e. an index on (b, c) would
see each range of PK data produce tightly grouped output that overlaps with the
output of A other ranges of the table.

The workload's main parameters are number of distinct values of a, b and c.
Initial data batches each correspond to one a/b pair containing c rows. By
default, batches are ordered by a then b (a=1/b=1, a=1/b=2, a=1,b=3, ...) though
this can optionally be inverted (a=1/b=1, a=2,b=1, a=3,b=1,...).

*/
package bulkingest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/col/coldata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const (
	bulkingestSchemaPrefix = `(
		a INT,
		b INT,
		c INT,
		payload STRING,
		PRIMARY KEY (a, b, c)`

	indexOnBCA = ",\n INDEX (b, c, a) STORING (payload)"

	defaultPayloadBytes = 100
)

type bulkingest struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed                                 int64
	aCount, bCount, cCount, payloadBytes int

	generateBsFirst bool
	indexBCA        bool
}

func init() {
	workload.Register(bulkingestMeta)
}

var bulkingestMeta = workload.Meta{
	Name:        `bulkingest`,
	Description: `bulkingest testdata is designed to produce a skewed distribution of KVs when ingested (in initial import or during later indexing)`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(693596)
		g := &bulkingest{}
		g.flags.FlagSet = pflag.NewFlagSet(`bulkingest`, pflag.ContinueOnError)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.IntVar(&g.aCount, `a`, 10, `number of values of A (i.e. pk prefix)`)
		g.flags.IntVar(&g.bCount, `b`, 10, `number of values of B (i.e. idx prefix)`)
		g.flags.IntVar(&g.cCount, `c`, 1000, `number of values of C (i.e. rows per A/B pair)`)
		g.flags.BoolVar(&g.generateBsFirst, `batches-by-b`, false, `generate all B batches for given A first`)
		g.flags.BoolVar(&g.indexBCA, `index-b-c-a`, true, `include an index on (B, C, A)`)
		g.flags.IntVar(&g.payloadBytes, `payload-bytes`, defaultPayloadBytes, `Size of the payload field in each row.`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*bulkingest) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(693597)
	return bulkingestMeta
}

func (w *bulkingest) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(693598)
	return w.flags
}

func (w *bulkingest) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(693599)
	return workload.Hooks{}
}

func (w *bulkingest) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(693600)
	schema := bulkingestSchemaPrefix
	if w.indexBCA {
		__antithesis_instrumentation__.Notify(693603)
		schema += indexOnBCA
	} else {
		__antithesis_instrumentation__.Notify(693604)
	}
	__antithesis_instrumentation__.Notify(693601)
	schema += ")"

	var bulkingestTypes = []*types.T{
		types.Int,
		types.Int,
		types.Int,
		types.Bytes,
	}

	table := workload.Table{
		Name:   `bulkingest`,
		Schema: schema,
		InitialRows: workload.BatchedTuples{
			NumBatches: w.aCount * w.bCount,
			FillBatch: func(ab int, cb coldata.Batch, alloc *bufalloc.ByteAllocator) {
				__antithesis_instrumentation__.Notify(693605)
				a := ab / w.bCount
				b := ab % w.bCount
				if w.generateBsFirst {
					__antithesis_instrumentation__.Notify(693607)
					b = ab / w.aCount
					a = ab % w.aCount
				} else {
					__antithesis_instrumentation__.Notify(693608)
				}
				__antithesis_instrumentation__.Notify(693606)

				cb.Reset(bulkingestTypes, w.cCount, coldata.StandardColumnFactory)
				aCol := cb.ColVec(0).Int64()
				bCol := cb.ColVec(1).Int64()
				cCol := cb.ColVec(2).Int64()
				payloadCol := cb.ColVec(3).Bytes()

				rng := rand.New(rand.NewSource(w.seed + int64(ab)))
				var payload []byte
				*alloc, payload = alloc.Alloc(w.cCount*w.payloadBytes, 0)
				randutil.ReadTestdataBytes(rng, payload)
				payloadCol.Reset()
				for rowIdx := 0; rowIdx < w.cCount; rowIdx++ {
					__antithesis_instrumentation__.Notify(693609)
					c := rowIdx
					off := c * w.payloadBytes
					aCol[rowIdx] = int64(a)
					bCol[rowIdx] = int64(b)
					cCol[rowIdx] = int64(c)
					payloadCol.Set(rowIdx, payload[off:off+w.payloadBytes])
				}
			},
		},
	}
	__antithesis_instrumentation__.Notify(693602)
	return []workload.Table{table}
}

func (w *bulkingest) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(693610)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(693615)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693616)
	}
	__antithesis_instrumentation__.Notify(693611)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(693617)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693618)
	}
	__antithesis_instrumentation__.Notify(693612)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	updateStmt, err := db.Prepare(`
		UPDATE bulkingest
		SET payload = $4
		WHERE a = $1 AND b = $2 AND c = $3
	`)
	if err != nil {
		__antithesis_instrumentation__.Notify(693619)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(693620)
	}
	__antithesis_instrumentation__.Notify(693613)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(693621)
		rng := rand.New(rand.NewSource(w.seed))
		hists := reg.GetHandle()
		pad := make([]byte, w.payloadBytes)
		workerFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(693623)
			a := rng.Intn(w.aCount)
			b := rng.Intn(w.bCount)
			c := rng.Intn(w.cCount)
			randutil.ReadTestdataBytes(rng, pad)

			start := timeutil.Now()
			res, err := updateStmt.Exec(a, b, c, pad)
			elapsed := timeutil.Since(start)
			hists.Get(`update-payload`).Record(elapsed)
			if err != nil {
				__antithesis_instrumentation__.Notify(693626)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693627)
			}
			__antithesis_instrumentation__.Notify(693624)
			if affected, err := res.RowsAffected(); err != nil {
				__antithesis_instrumentation__.Notify(693628)
				return err
			} else {
				__antithesis_instrumentation__.Notify(693629)
				if affected != 1 {
					__antithesis_instrumentation__.Notify(693630)
					return errors.Errorf("expected 1 row affected, got %d", affected)
				} else {
					__antithesis_instrumentation__.Notify(693631)
				}
			}
			__antithesis_instrumentation__.Notify(693625)
			return nil
		}
		__antithesis_instrumentation__.Notify(693622)
		ql.WorkerFns = append(ql.WorkerFns, workerFn)
	}
	__antithesis_instrumentation__.Notify(693614)
	return ql, nil
}
