package kv

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"crypto/sha1"
	gosql "database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/spf13/pflag"
)

const (
	kvSchema = `(
		k BIGINT NOT NULL PRIMARY KEY,
		v BYTES NOT NULL
	)`
	kvSchemaWithIndex = `(
		k BIGINT NOT NULL PRIMARY KEY,
		v BYTES NOT NULL,
		INDEX (v)
	)`

	shardedKvSchema = `(
		k BIGINT NOT NULL,
		v BYTES NOT NULL,
		shard INT4 AS (mod(k, %d)) STORED CHECK (%s),
		PRIMARY KEY (shard, k)
	)`
	shardedKvSchemaWithIndex = `(
		k BIGINT NOT NULL,
		v BYTES NOT NULL,
		shard INT4 AS (mod(k, %d)) STORED CHECK (%s),
		PRIMARY KEY (shard, k),
		INDEX (v)
	)`
)

type kv struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	batchSize                            int
	minBlockSizeBytes, maxBlockSizeBytes int
	cycleLength                          int64
	readPercent                          int
	spanPercent                          int
	spanLimit                            int
	writesUseSelectForUpdate             bool
	seed                                 int64
	writeSeq                             string
	sequential                           bool
	zipfian                              bool
	splits                               int
	secondaryIndex                       bool
	shards                               int
	targetCompressionRatio               float64
	enum                                 bool
}

func init() {
	workload.Register(kvMeta)
}

var kvMeta = workload.Meta{
	Name:        `kv`,
	Description: `KV reads and writes to keys spread randomly across the cluster.`,
	Details: `
	By default, keys are picked uniformly at random across the cluster.
	--concurrency workers alternate between doing selects and upserts (according
	to a --read-percent ratio). Each select/upsert reads/writes a batch of --batch
	rows. The write keys are randomly generated in a deterministic fashion (or
	sequentially if --sequential is specified). Reads select a random batch of ids
	out of the ones previously written.
	--write-seq can be used to incorporate data produced by a previous run into
	the current run.
	`,
	Version:      `1.0.0`,
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(694349)
		g := &kv{}
		g.flags.FlagSet = pflag.NewFlagSet(`kv`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`batch`: {RuntimeOnly: true},
		}
		g.flags.IntVar(&g.batchSize, `batch`, 1,
			`Number of blocks to read/insert in a single SQL statement.`)
		g.flags.IntVar(&g.minBlockSizeBytes, `min-block-bytes`, 1,
			`Minimum amount of raw data written with each insertion.`)
		g.flags.IntVar(&g.maxBlockSizeBytes, `max-block-bytes`, 1,
			`Maximum amount of raw data written with each insertion`)
		g.flags.Int64Var(&g.cycleLength, `cycle-length`, math.MaxInt64,
			`Number of keys repeatedly accessed by each writer through upserts.`)
		g.flags.IntVar(&g.readPercent, `read-percent`, 0,
			`Percent (0-100) of operations that are reads of existing keys.`)
		g.flags.IntVar(&g.spanPercent, `span-percent`, 0,
			`Percent (0-100) of operations that are spanning queries of all ranges.`)
		g.flags.IntVar(&g.spanLimit, `span-limit`, 0,
			`LIMIT count for each spanning query, or 0 for no limit`)
		g.flags.BoolVar(&g.writesUseSelectForUpdate, `sfu-writes`, false,
			`Use SFU and transactional writes with a sleep after SFU.`)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.BoolVar(&g.zipfian, `zipfian`, false,
			`Pick keys in a zipfian distribution instead of randomly.`)
		g.flags.BoolVar(&g.sequential, `sequential`, false,
			`Pick keys sequentially instead of randomly.`)
		g.flags.StringVar(&g.writeSeq, `write-seq`, "",
			`Initial write sequence value. Can be used to use the data produced by a previous run. `+
				`It has to be of the form (R|S)<number>, where S implies that it was taken from a `+
				`previous --sequential run and R implies a previous random run.`)
		g.flags.IntVar(&g.splits, `splits`, 0,
			`Number of splits to perform before starting normal operations.`)
		g.flags.BoolVar(&g.secondaryIndex, `secondary-index`, false,
			`Add a secondary index to the schema`)
		g.flags.IntVar(&g.shards, `num-shards`, 0,
			`Number of shards to create on the primary key.`)
		g.flags.Float64Var(&g.targetCompressionRatio, `target-compression-ratio`, 1.0,
			`Target compression ratio for data blocks. Must be >= 1.0`)
		g.flags.BoolVar(&g.enum, `enum`, false,
			`Inject an enum column and use it`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*kv) Meta() workload.Meta { __antithesis_instrumentation__.Notify(694350); return kvMeta }

func (w *kv) Flags() workload.Flags { __antithesis_instrumentation__.Notify(694351); return w.flags }

func (w *kv) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(694352)
	return workload.Hooks{
		PostLoad: func(db *gosql.DB) error {
			__antithesis_instrumentation__.Notify(694353)
			if !w.enum {
				__antithesis_instrumentation__.Notify(694355)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(694356)
			}
			__antithesis_instrumentation__.Notify(694354)
			_, err := db.Exec(`
CREATE TYPE enum_type AS ENUM ('v');
ALTER TABLE kv ADD COLUMN e enum_type NOT NULL AS ('v') STORED;`)
			return err
		},
		Validate: func() error {
			__antithesis_instrumentation__.Notify(694357)
			if w.maxBlockSizeBytes < w.minBlockSizeBytes {
				__antithesis_instrumentation__.Notify(694363)
				return errors.Errorf("Value of 'max-block-bytes' (%d) must be greater than or equal to value of 'min-block-bytes' (%d)",
					w.maxBlockSizeBytes, w.minBlockSizeBytes)
			} else {
				__antithesis_instrumentation__.Notify(694364)
			}
			__antithesis_instrumentation__.Notify(694358)
			if w.sequential && func() bool {
				__antithesis_instrumentation__.Notify(694365)
				return w.splits > 0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(694366)
				return errors.New("'sequential' and 'splits' cannot both be enabled")
			} else {
				__antithesis_instrumentation__.Notify(694367)
			}
			__antithesis_instrumentation__.Notify(694359)
			if w.sequential && func() bool {
				__antithesis_instrumentation__.Notify(694368)
				return w.zipfian == true
			}() == true {
				__antithesis_instrumentation__.Notify(694369)
				return errors.New("'sequential' and 'zipfian' cannot both be enabled")
			} else {
				__antithesis_instrumentation__.Notify(694370)
			}
			__antithesis_instrumentation__.Notify(694360)
			if w.readPercent+w.spanPercent > 100 {
				__antithesis_instrumentation__.Notify(694371)
				return errors.New("'read-percent' and 'span-percent' higher than 100")
			} else {
				__antithesis_instrumentation__.Notify(694372)
			}
			__antithesis_instrumentation__.Notify(694361)
			if w.targetCompressionRatio < 1.0 || func() bool {
				__antithesis_instrumentation__.Notify(694373)
				return math.IsNaN(w.targetCompressionRatio) == true
			}() == true {
				__antithesis_instrumentation__.Notify(694374)
				return errors.New("'target-compression-ratio' must be a number >= 1.0")
			} else {
				__antithesis_instrumentation__.Notify(694375)
			}
			__antithesis_instrumentation__.Notify(694362)
			return nil
		},
	}
}

func (w *kv) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694376)
	table := workload.Table{
		Name: `kv`,

		Splits: workload.Tuples(
			w.splits,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(694379)
				stride := (float64(w.cycleLength) - float64(math.MinInt64)) / float64(w.splits+1)
				splitPoint := int(math.MinInt64 + float64(splitIdx+1)*stride)
				return []interface{}{splitPoint}
			},
		),
	}
	__antithesis_instrumentation__.Notify(694377)
	if w.shards > 0 {
		__antithesis_instrumentation__.Notify(694380)
		schema := shardedKvSchema
		if w.secondaryIndex {
			__antithesis_instrumentation__.Notify(694383)
			schema = shardedKvSchemaWithIndex
		} else {
			__antithesis_instrumentation__.Notify(694384)
		}
		__antithesis_instrumentation__.Notify(694381)
		checkConstraint := strings.Builder{}
		checkConstraint.WriteString(`shard IN (`)
		for i := 0; i < w.shards; i++ {
			__antithesis_instrumentation__.Notify(694385)
			if i != 0 {
				__antithesis_instrumentation__.Notify(694387)
				checkConstraint.WriteString(",")
			} else {
				__antithesis_instrumentation__.Notify(694388)
			}
			__antithesis_instrumentation__.Notify(694386)
			fmt.Fprintf(&checkConstraint, "%d", i)
		}
		__antithesis_instrumentation__.Notify(694382)
		checkConstraint.WriteString(")")
		table.Schema = fmt.Sprintf(schema, w.shards, checkConstraint.String())
	} else {
		__antithesis_instrumentation__.Notify(694389)
		if w.secondaryIndex {
			__antithesis_instrumentation__.Notify(694390)
			table.Schema = kvSchemaWithIndex
		} else {
			__antithesis_instrumentation__.Notify(694391)
			table.Schema = kvSchema
		}
	}
	__antithesis_instrumentation__.Notify(694378)
	return []workload.Table{table}
}

func (w *kv) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(694392)
	writeSeq := 0
	if w.writeSeq != "" {
		__antithesis_instrumentation__.Notify(694401)
		first := w.writeSeq[0]
		if len(w.writeSeq) < 2 || func() bool {
			__antithesis_instrumentation__.Notify(694405)
			return (first != 'R' && func() bool {
				__antithesis_instrumentation__.Notify(694406)
				return first != 'S' == true
			}() == true) == true
		}() == true {
			__antithesis_instrumentation__.Notify(694407)
			return workload.QueryLoad{}, fmt.Errorf("--write-seq has to be of the form '(R|S)<num>'")
		} else {
			__antithesis_instrumentation__.Notify(694408)
		}
		__antithesis_instrumentation__.Notify(694402)
		rest := w.writeSeq[1:]
		var err error
		writeSeq, err = strconv.Atoi(rest)
		if err != nil {
			__antithesis_instrumentation__.Notify(694409)
			return workload.QueryLoad{}, fmt.Errorf("--write-seq has to be of the form '(R|S)<num>'")
		} else {
			__antithesis_instrumentation__.Notify(694410)
		}
		__antithesis_instrumentation__.Notify(694403)
		if first == 'R' && func() bool {
			__antithesis_instrumentation__.Notify(694411)
			return w.sequential == true
		}() == true {
			__antithesis_instrumentation__.Notify(694412)
			return workload.QueryLoad{}, fmt.Errorf("--sequential incompatible with a Random --write-seq")
		} else {
			__antithesis_instrumentation__.Notify(694413)
		}
		__antithesis_instrumentation__.Notify(694404)
		if first == 'S' && func() bool {
			__antithesis_instrumentation__.Notify(694414)
			return !w.sequential == true
		}() == true {
			__antithesis_instrumentation__.Notify(694415)
			return workload.QueryLoad{}, fmt.Errorf(
				"--sequential=false incompatible with a Sequential --write-seq")
		} else {
			__antithesis_instrumentation__.Notify(694416)
		}
	} else {
		__antithesis_instrumentation__.Notify(694417)
	}
	__antithesis_instrumentation__.Notify(694393)

	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694418)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694419)
	}
	__antithesis_instrumentation__.Notify(694394)
	cfg := workload.MultiConnPoolCfg{
		MaxTotalConnections: w.connFlags.Concurrency + 1,
	}
	mcp, err := workload.NewMultiConnPool(ctx, cfg, urls...)
	if err != nil {
		__antithesis_instrumentation__.Notify(694420)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694421)
	}
	__antithesis_instrumentation__.Notify(694395)

	var buf strings.Builder
	if w.shards == 0 {
		__antithesis_instrumentation__.Notify(694422)
		buf.WriteString(`SELECT k, v FROM kv WHERE k IN (`)
		for i := 0; i < w.batchSize; i++ {
			__antithesis_instrumentation__.Notify(694423)
			if i > 0 {
				__antithesis_instrumentation__.Notify(694425)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(694426)
			}
			__antithesis_instrumentation__.Notify(694424)
			fmt.Fprintf(&buf, `$%d`, i+1)
		}
	} else {
		__antithesis_instrumentation__.Notify(694427)
		if w.enum {
			__antithesis_instrumentation__.Notify(694428)
			buf.WriteString(`SELECT k, v, e FROM kv WHERE k IN (`)
			for i := 0; i < w.batchSize; i++ {
				__antithesis_instrumentation__.Notify(694429)
				if i > 0 {
					__antithesis_instrumentation__.Notify(694431)
					buf.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(694432)
				}
				__antithesis_instrumentation__.Notify(694430)
				fmt.Fprintf(&buf, `$%d`, i+1)
			}
		} else {
			__antithesis_instrumentation__.Notify(694433)

			buf.WriteString(`SELECT k, v FROM kv WHERE (shard, k) in (`)
			for i := 0; i < w.batchSize; i++ {
				__antithesis_instrumentation__.Notify(694434)
				if i > 0 {
					__antithesis_instrumentation__.Notify(694436)
					buf.WriteString(", ")
				} else {
					__antithesis_instrumentation__.Notify(694437)
				}
				__antithesis_instrumentation__.Notify(694435)
				fmt.Fprintf(&buf, `(mod($%d, %d), $%d)`, i+1, w.shards, i+1)
			}
		}
	}
	__antithesis_instrumentation__.Notify(694396)
	buf.WriteString(`)`)
	readStmtStr := buf.String()

	buf.Reset()
	buf.WriteString(`UPSERT INTO kv (k, v) VALUES`)
	for i := 0; i < w.batchSize; i++ {
		__antithesis_instrumentation__.Notify(694438)
		j := i * 2
		if i > 0 {
			__antithesis_instrumentation__.Notify(694440)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(694441)
		}
		__antithesis_instrumentation__.Notify(694439)
		fmt.Fprintf(&buf, ` ($%d, $%d)`, j+1, j+2)
	}
	__antithesis_instrumentation__.Notify(694397)
	writeStmtStr := buf.String()

	var sfuStmtStr string
	if w.writesUseSelectForUpdate {
		__antithesis_instrumentation__.Notify(694442)
		if w.shards != 0 {
			__antithesis_instrumentation__.Notify(694445)
			return workload.QueryLoad{}, fmt.Errorf("select for update in kv requires shard=0")
		} else {
			__antithesis_instrumentation__.Notify(694446)
		}
		__antithesis_instrumentation__.Notify(694443)
		buf.Reset()
		buf.WriteString(`SELECT k, v FROM kv WHERE k IN (`)
		for i := 0; i < w.batchSize; i++ {
			__antithesis_instrumentation__.Notify(694447)
			if i > 0 {
				__antithesis_instrumentation__.Notify(694449)
				buf.WriteString(", ")
			} else {
				__antithesis_instrumentation__.Notify(694450)
			}
			__antithesis_instrumentation__.Notify(694448)
			fmt.Fprintf(&buf, `$%d`, i+1)
		}
		__antithesis_instrumentation__.Notify(694444)
		buf.WriteString(`) FOR UPDATE`)
		sfuStmtStr = buf.String()
	} else {
		__antithesis_instrumentation__.Notify(694451)
	}
	__antithesis_instrumentation__.Notify(694398)

	buf.Reset()
	buf.WriteString(`SELECT count(v) FROM [SELECT v FROM kv`)
	if w.spanLimit > 0 {
		__antithesis_instrumentation__.Notify(694452)
		fmt.Fprintf(&buf, ` ORDER BY k LIMIT %d`, w.spanLimit)
	} else {
		__antithesis_instrumentation__.Notify(694453)
	}
	__antithesis_instrumentation__.Notify(694399)
	buf.WriteString(`]`)
	spanStmtStr := buf.String()

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	seq := &sequence{config: w, val: int64(writeSeq)}
	numEmptyResults := new(int64)
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(694454)
		op := &kvOp{
			config:          w,
			hists:           reg.GetHandle(),
			numEmptyResults: numEmptyResults,
		}
		op.readStmt = op.sr.Define(readStmtStr)
		op.writeStmt = op.sr.Define(writeStmtStr)
		if len(sfuStmtStr) > 0 {
			__antithesis_instrumentation__.Notify(694458)
			op.sfuStmt = op.sr.Define(sfuStmtStr)
		} else {
			__antithesis_instrumentation__.Notify(694459)
		}
		__antithesis_instrumentation__.Notify(694455)
		op.spanStmt = op.sr.Define(spanStmtStr)
		if err := op.sr.Init(ctx, "kv", mcp, w.connFlags); err != nil {
			__antithesis_instrumentation__.Notify(694460)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(694461)
		}
		__antithesis_instrumentation__.Notify(694456)
		op.mcp = mcp
		if w.sequential {
			__antithesis_instrumentation__.Notify(694462)
			op.g = newSequentialGenerator(seq)
		} else {
			__antithesis_instrumentation__.Notify(694463)
			if w.zipfian {
				__antithesis_instrumentation__.Notify(694464)
				op.g = newZipfianGenerator(seq)
			} else {
				__antithesis_instrumentation__.Notify(694465)
				op.g = newHashGenerator(seq)
			}
		}
		__antithesis_instrumentation__.Notify(694457)
		ql.WorkerFns = append(ql.WorkerFns, op.run)
		ql.Close = op.close
	}
	__antithesis_instrumentation__.Notify(694400)
	return ql, nil
}

type kvOp struct {
	config          *kv
	hists           *histogram.Histograms
	sr              workload.SQLRunner
	mcp             *workload.MultiConnPool
	readStmt        workload.StmtHandle
	writeStmt       workload.StmtHandle
	spanStmt        workload.StmtHandle
	sfuStmt         workload.StmtHandle
	g               keyGenerator
	numEmptyResults *int64
}

func (o *kvOp) run(ctx context.Context) (retErr error) {
	__antithesis_instrumentation__.Notify(694466)
	statementProbability := o.g.rand().Intn(100)
	if statementProbability < o.config.readPercent {
		__antithesis_instrumentation__.Notify(694472)
		args := make([]interface{}, o.config.batchSize)
		for i := 0; i < o.config.batchSize; i++ {
			__antithesis_instrumentation__.Notify(694477)
			args[i] = o.g.readKey()
		}
		__antithesis_instrumentation__.Notify(694473)
		start := timeutil.Now()
		rows, err := o.readStmt.Query(ctx, args...)
		if err != nil {
			__antithesis_instrumentation__.Notify(694478)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694479)
		}
		__antithesis_instrumentation__.Notify(694474)
		empty := true
		for rows.Next() {
			__antithesis_instrumentation__.Notify(694480)
			empty = false
		}
		__antithesis_instrumentation__.Notify(694475)
		if empty {
			__antithesis_instrumentation__.Notify(694481)
			atomic.AddInt64(o.numEmptyResults, 1)
		} else {
			__antithesis_instrumentation__.Notify(694482)
		}
		__antithesis_instrumentation__.Notify(694476)
		elapsed := timeutil.Since(start)
		o.hists.Get(`read`).Record(elapsed)
		return rows.Err()
	} else {
		__antithesis_instrumentation__.Notify(694483)
	}
	__antithesis_instrumentation__.Notify(694467)

	statementProbability -= o.config.readPercent
	if statementProbability < o.config.spanPercent {
		__antithesis_instrumentation__.Notify(694484)
		start := timeutil.Now()
		_, err := o.spanStmt.Exec(ctx)
		elapsed := timeutil.Since(start)
		o.hists.Get(`span`).Record(elapsed)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694485)
	}
	__antithesis_instrumentation__.Notify(694468)
	const argCount = 2
	writeArgs := make([]interface{}, argCount*o.config.batchSize)
	var sfuArgs []interface{}
	if o.config.writesUseSelectForUpdate {
		__antithesis_instrumentation__.Notify(694486)
		sfuArgs = make([]interface{}, o.config.batchSize)
	} else {
		__antithesis_instrumentation__.Notify(694487)
	}
	__antithesis_instrumentation__.Notify(694469)
	for i := 0; i < o.config.batchSize; i++ {
		__antithesis_instrumentation__.Notify(694488)
		j := i * argCount
		writeArgs[j+0] = o.g.writeKey()
		if sfuArgs != nil {
			__antithesis_instrumentation__.Notify(694490)
			sfuArgs[i] = writeArgs[j]
		} else {
			__antithesis_instrumentation__.Notify(694491)
		}
		__antithesis_instrumentation__.Notify(694489)
		writeArgs[j+1] = randomBlock(o.config, o.g.rand())
	}
	__antithesis_instrumentation__.Notify(694470)
	start := timeutil.Now()
	var err error
	if o.config.writesUseSelectForUpdate {
		__antithesis_instrumentation__.Notify(694492)

		var tx pgx.Tx
		if tx, err = o.mcp.Get().Begin(ctx); err != nil {
			__antithesis_instrumentation__.Notify(694498)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694499)
		}
		__antithesis_instrumentation__.Notify(694493)
		defer func() {
			__antithesis_instrumentation__.Notify(694500)
			rollbackErr := tx.Rollback(ctx)
			if !errors.Is(rollbackErr, pgx.ErrTxClosed) {
				__antithesis_instrumentation__.Notify(694501)
				retErr = errors.CombineErrors(retErr, rollbackErr)
			} else {
				__antithesis_instrumentation__.Notify(694502)
			}
		}()
		__antithesis_instrumentation__.Notify(694494)
		rows, err := o.sfuStmt.QueryTx(ctx, tx, sfuArgs...)
		if err != nil {
			__antithesis_instrumentation__.Notify(694503)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694504)
		}
		__antithesis_instrumentation__.Notify(694495)
		rows.Close()
		if err = rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(694505)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694506)
		}
		__antithesis_instrumentation__.Notify(694496)

		time.Sleep(10 * time.Millisecond)
		if _, err = o.writeStmt.ExecTx(ctx, tx, writeArgs...); err != nil {
			__antithesis_instrumentation__.Notify(694507)

			return o.tryHandleWriteErr("write-write-err", start, err)
		} else {
			__antithesis_instrumentation__.Notify(694508)
		}
		__antithesis_instrumentation__.Notify(694497)
		if err = tx.Commit(ctx); err != nil {
			__antithesis_instrumentation__.Notify(694509)
			return o.tryHandleWriteErr("write-commit-err", start, err)
		} else {
			__antithesis_instrumentation__.Notify(694510)
		}
	} else {
		__antithesis_instrumentation__.Notify(694511)
		_, err = o.writeStmt.Exec(ctx, writeArgs...)
	}
	__antithesis_instrumentation__.Notify(694471)
	elapsed := timeutil.Since(start)
	o.hists.Get(`write`).Record(elapsed)
	return err
}

func (o *kvOp) tryHandleWriteErr(name string, start time.Time, err error) error {
	__antithesis_instrumentation__.Notify(694512)

	pgErr := new(pgconn.PgError)
	if !errors.As(err, &pgErr) {
		__antithesis_instrumentation__.Notify(694515)
		return err
	} else {
		__antithesis_instrumentation__.Notify(694516)
	}
	__antithesis_instrumentation__.Notify(694513)

	if pgcode.MakeCode(pgErr.Code) == pgcode.SerializationFailure {
		__antithesis_instrumentation__.Notify(694517)
		elapsed := timeutil.Since(start)
		o.hists.Get(name).Record(elapsed)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(694518)
	}
	__antithesis_instrumentation__.Notify(694514)
	return err
}

func (o *kvOp) close(context.Context) {
	__antithesis_instrumentation__.Notify(694519)
	if empty := atomic.LoadInt64(o.numEmptyResults); empty != 0 {
		__antithesis_instrumentation__.Notify(694522)
		fmt.Printf("Number of reads that didn't return any results: %d.\n", empty)
	} else {
		__antithesis_instrumentation__.Notify(694523)
	}
	__antithesis_instrumentation__.Notify(694520)
	seq := o.g.sequence()
	var ch string
	if o.config.sequential {
		__antithesis_instrumentation__.Notify(694524)
		ch = "S"
	} else {
		__antithesis_instrumentation__.Notify(694525)
		ch = "R"
	}
	__antithesis_instrumentation__.Notify(694521)
	fmt.Printf("Highest sequence written: %d. Can be passed as --write-seq=%s%d to the next run.\n",
		seq, ch, seq)
}

type sequence struct {
	config *kv
	val    int64
}

func (s *sequence) write() int64 {
	__antithesis_instrumentation__.Notify(694526)
	return (atomic.AddInt64(&s.val, 1) - 1) % s.config.cycleLength
}

func (s *sequence) read() int64 {
	__antithesis_instrumentation__.Notify(694527)
	return atomic.LoadInt64(&s.val) % s.config.cycleLength
}

type keyGenerator interface {
	writeKey() int64
	readKey() int64
	rand() *rand.Rand
	sequence() int64
}

type hashGenerator struct {
	seq    *sequence
	random *rand.Rand
	hasher hash.Hash
	buf    [sha1.Size]byte
}

func newHashGenerator(seq *sequence) *hashGenerator {
	__antithesis_instrumentation__.Notify(694528)
	return &hashGenerator{
		seq:    seq,
		random: rand.New(rand.NewSource(timeutil.Now().UnixNano())),
		hasher: sha1.New(),
	}
}

func (g *hashGenerator) hash(v int64) int64 {
	__antithesis_instrumentation__.Notify(694529)
	binary.BigEndian.PutUint64(g.buf[:8], uint64(v))
	binary.BigEndian.PutUint64(g.buf[8:16], uint64(g.seq.config.seed))
	g.hasher.Reset()
	_, _ = g.hasher.Write(g.buf[:16])
	g.hasher.Sum(g.buf[:0])
	return int64(binary.BigEndian.Uint64(g.buf[:8]))
}

func (g *hashGenerator) writeKey() int64 {
	__antithesis_instrumentation__.Notify(694530)
	return g.hash(g.seq.write())
}

func (g *hashGenerator) readKey() int64 {
	__antithesis_instrumentation__.Notify(694531)
	v := g.seq.read()
	if v == 0 {
		__antithesis_instrumentation__.Notify(694533)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(694534)
	}
	__antithesis_instrumentation__.Notify(694532)
	return g.hash(g.random.Int63n(v))
}

func (g *hashGenerator) rand() *rand.Rand {
	__antithesis_instrumentation__.Notify(694535)
	return g.random
}

func (g *hashGenerator) sequence() int64 {
	__antithesis_instrumentation__.Notify(694536)
	return atomic.LoadInt64(&g.seq.val)
}

type sequentialGenerator struct {
	seq    *sequence
	random *rand.Rand
}

func newSequentialGenerator(seq *sequence) *sequentialGenerator {
	__antithesis_instrumentation__.Notify(694537)
	return &sequentialGenerator{
		seq:    seq,
		random: rand.New(rand.NewSource(timeutil.Now().UnixNano())),
	}
}

func (g *sequentialGenerator) writeKey() int64 {
	__antithesis_instrumentation__.Notify(694538)
	return g.seq.write()
}

func (g *sequentialGenerator) readKey() int64 {
	__antithesis_instrumentation__.Notify(694539)
	v := g.seq.read()
	if v == 0 {
		__antithesis_instrumentation__.Notify(694541)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(694542)
	}
	__antithesis_instrumentation__.Notify(694540)
	return g.random.Int63n(v)
}

func (g *sequentialGenerator) rand() *rand.Rand {
	__antithesis_instrumentation__.Notify(694543)
	return g.random
}

func (g *sequentialGenerator) sequence() int64 {
	__antithesis_instrumentation__.Notify(694544)
	return atomic.LoadInt64(&g.seq.val)
}

type zipfGenerator struct {
	seq    *sequence
	random *rand.Rand
	zipf   *zipf
}

func newZipfianGenerator(seq *sequence) *zipfGenerator {
	__antithesis_instrumentation__.Notify(694545)
	random := rand.New(rand.NewSource(timeutil.Now().UnixNano()))
	return &zipfGenerator{
		seq:    seq,
		random: random,
		zipf:   newZipf(1.1, 1, uint64(math.MaxInt64)),
	}
}

func (g *zipfGenerator) zipfian(seed int64) int64 {
	__antithesis_instrumentation__.Notify(694546)
	randomWithSeed := rand.New(rand.NewSource(seed))
	return int64(g.zipf.Uint64(randomWithSeed))
}

func (g *zipfGenerator) writeKey() int64 {
	__antithesis_instrumentation__.Notify(694547)
	return g.zipfian(g.seq.write())
}

func (g *zipfGenerator) readKey() int64 {
	__antithesis_instrumentation__.Notify(694548)
	v := g.seq.read()
	if v == 0 {
		__antithesis_instrumentation__.Notify(694550)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(694551)
	}
	__antithesis_instrumentation__.Notify(694549)
	return g.zipfian(g.random.Int63n(v))
}

func (g *zipfGenerator) rand() *rand.Rand {
	__antithesis_instrumentation__.Notify(694552)
	return g.random
}

func (g *zipfGenerator) sequence() int64 {
	__antithesis_instrumentation__.Notify(694553)
	return atomic.LoadInt64(&g.seq.val)
}

func randomBlock(config *kv, r *rand.Rand) []byte {
	__antithesis_instrumentation__.Notify(694554)
	blockSize := r.Intn(config.maxBlockSizeBytes-config.minBlockSizeBytes+1) + config.minBlockSizeBytes
	blockData := make([]byte, blockSize)
	uniqueSize := int(float64(blockSize) / config.targetCompressionRatio)
	if uniqueSize < 1 {
		__antithesis_instrumentation__.Notify(694557)
		uniqueSize = 1
	} else {
		__antithesis_instrumentation__.Notify(694558)
	}
	__antithesis_instrumentation__.Notify(694555)
	for i := range blockData {
		__antithesis_instrumentation__.Notify(694559)
		if i >= uniqueSize {
			__antithesis_instrumentation__.Notify(694560)
			blockData[i] = blockData[i-uniqueSize]
		} else {
			__antithesis_instrumentation__.Notify(694561)
			blockData[i] = byte(r.Int() & 0xff)
		}
	}
	__antithesis_instrumentation__.Notify(694556)
	return blockData
}
