package indexes

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uint128"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const (
	schemaBase = `(
		key     UUID  NOT NULL PRIMARY KEY,
		col0    INT   NOT NULL,
		col1    INT   NOT NULL,
		col2    INT   NOT NULL,
		col3    INT   NOT NULL,
		col4    INT   NOT NULL,
		col5    INT   NOT NULL,
		col6    INT   NOT NULL,
		col7    INT   NOT NULL,
		col8    INT   NOT NULL,
		col9    INT   NOT NULL,
		payload BYTES NOT NULL`
)

type indexes struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed        int64
	idxs        int
	unique      bool
	payload     int
	cycleLength uint64
}

func init() {
	workload.Register(indexesMeta)
}

var indexesMeta = workload.Meta{
	Name:        `indexes`,
	Description: `Indexes writes to a table with a variable number of secondary indexes`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(694211)
		g := &indexes{}
		g.flags.FlagSet = pflag.NewFlagSet(`indexes`, pflag.ContinueOnError)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.IntVar(&g.idxs, `secondary-indexes`, 1, `Number of indexes to add to the table.`)
		g.flags.BoolVar(&g.unique, `unique-indexes`, false, `Use UNIQUE secondary indexes.`)
		g.flags.IntVar(&g.payload, `payload`, 64, `Size of the unindexed payload column.`)
		g.flags.Uint64Var(&g.cycleLength, `cycle-length`, math.MaxUint64,
			`Number of keys repeatedly accessed by each writer through upserts.`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*indexes) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(694212)
	return indexesMeta
}

func (w *indexes) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(694213)
	return w.flags
}

func (w *indexes) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(694214)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(694215)
			if w.idxs < 0 || func() bool {
				__antithesis_instrumentation__.Notify(694218)
				return w.idxs > 99 == true
			}() == true {
				__antithesis_instrumentation__.Notify(694219)
				return errors.Errorf(`--secondary-indexes must be in range [0, 99]`)
			} else {
				__antithesis_instrumentation__.Notify(694220)
			}
			__antithesis_instrumentation__.Notify(694216)
			if w.payload < 1 {
				__antithesis_instrumentation__.Notify(694221)
				return errors.Errorf(`--payload size must be equal to or greater than 1`)
			} else {
				__antithesis_instrumentation__.Notify(694222)
			}
			__antithesis_instrumentation__.Notify(694217)
			return nil
		},
		PostLoad: func(sqlDB *gosql.DB) error {
			__antithesis_instrumentation__.Notify(694223)

			if err := maybeDisableMergeQueue(sqlDB); err != nil {
				__antithesis_instrumentation__.Notify(694226)
				return err
			} else {
				__antithesis_instrumentation__.Notify(694227)
			}
			__antithesis_instrumentation__.Notify(694224)

			for i := 0; i < w.idxs; i++ {
				__antithesis_instrumentation__.Notify(694228)
				split := fmt.Sprintf(`ALTER INDEX idx%d SPLIT AT VALUES (%d)`, i, math.MinInt64)
				if _, err := sqlDB.Exec(split); err != nil {
					__antithesis_instrumentation__.Notify(694229)
					return err
				} else {
					__antithesis_instrumentation__.Notify(694230)
				}
			}
			__antithesis_instrumentation__.Notify(694225)
			return nil
		},
	}
}

func maybeDisableMergeQueue(sqlDB *gosql.DB) error {
	__antithesis_instrumentation__.Notify(694231)
	var ok bool
	if err := sqlDB.QueryRow(
		`SELECT count(*) > 0 FROM [ SHOW ALL CLUSTER SETTINGS ] AS _ (v) WHERE v = 'kv.range_merge.queue_enabled'`,
	).Scan(&ok); err != nil || func() bool {
		__antithesis_instrumentation__.Notify(694233)
		return !ok == true
	}() == true {
		__antithesis_instrumentation__.Notify(694234)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694235)
	}
	__antithesis_instrumentation__.Notify(694232)
	_, err := sqlDB.Exec("SET CLUSTER SETTING kv.range_merge.queue_enabled = false")
	return err
}

func (w *indexes) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694236)

	var unique string
	if w.unique {
		__antithesis_instrumentation__.Notify(694239)
		unique = "UNIQUE "
	} else {
		__antithesis_instrumentation__.Notify(694240)
	}
	__antithesis_instrumentation__.Notify(694237)
	var b strings.Builder
	b.WriteString(schemaBase)
	for i := 0; i < w.idxs; i++ {
		__antithesis_instrumentation__.Notify(694241)
		col1, col2 := i/10, i%10
		if col1 == col2 {
			__antithesis_instrumentation__.Notify(694242)
			fmt.Fprintf(&b, ",\n\t\t%sINDEX idx%d (col%d)", unique, i, col1)
		} else {
			__antithesis_instrumentation__.Notify(694243)
			fmt.Fprintf(&b, ",\n\t\t%sINDEX idx%d (col%d, col%d)", unique, i, col1, col2)
		}
	}
	__antithesis_instrumentation__.Notify(694238)
	b.WriteString("\n)")

	return []workload.Table{{
		Name:   `indexes`,
		Schema: b.String(),
	}}
}

func (w *indexes) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(694244)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694248)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694249)
	}
	__antithesis_instrumentation__.Notify(694245)
	cfg := workload.MultiConnPoolCfg{
		MaxTotalConnections: w.connFlags.Concurrency + 1,
	}
	mcp, err := workload.NewMultiConnPool(ctx, cfg, urls...)
	if err != nil {
		__antithesis_instrumentation__.Notify(694250)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694251)
	}
	__antithesis_instrumentation__.Notify(694246)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	const stmt = `UPSERT INTO indexes VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(694252)
		op := &indexesOp{
			config: w,
			hists:  reg.GetHandle(),
			rand:   rand.New(rand.NewSource(int64((i + 1)) * w.seed)),
			buf:    make([]byte, w.payload),
		}
		op.stmt = op.sr.Define(stmt)
		if err := op.sr.Init(ctx, "indexes", mcp, w.connFlags); err != nil {
			__antithesis_instrumentation__.Notify(694254)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(694255)
		}
		__antithesis_instrumentation__.Notify(694253)
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(694247)
	return ql, nil
}

type indexesOp struct {
	config *indexes
	hists  *histogram.Histograms
	rand   *rand.Rand
	sr     workload.SQLRunner
	stmt   workload.StmtHandle
	buf    []byte
}

func (o *indexesOp) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(694256)
	keyLo := o.rand.Uint64() % o.config.cycleLength
	_, _ = o.rand.Read(o.buf[:])
	args := []interface{}{
		uuid.FromUint128(uint128.FromInts(0, keyLo)).String(),
		int64(keyLo + 0),
		int64(keyLo + 1),
		int64(keyLo + 2),
		int64(keyLo + 3),
		int64(keyLo + 4),
		int64(keyLo + 5),
		int64(keyLo + 6),
		int64(keyLo + 7),
		int64(keyLo + 8),
		int64(keyLo + 9),
		o.buf[:],
	}

	start := timeutil.Now()
	_, err := o.stmt.Exec(ctx, args...)
	elapsed := timeutil.Since(start)
	o.hists.Get(`write`).Record(elapsed)
	return err
}
