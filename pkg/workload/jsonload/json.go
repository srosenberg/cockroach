package jsonload

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"crypto/sha1"
	gosql "database/sql"
	"encoding/binary"
	"fmt"
	"hash"
	"math"
	"math/rand"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

const (
	jsonSchema                  = `(k BIGINT NOT NULL PRIMARY KEY, v JSONB NOT NULL)`
	jsonSchemaWithInvertedIndex = `(k BIGINT NOT NULL PRIMARY KEY, v JSONB NOT NULL, INVERTED INDEX (v))`
	jsonSchemaWithComputed      = `(k BIGINT AS (v->>'key')::BIGINT STORED PRIMARY KEY, v JSONB NOT NULL)`
)

type jsonLoad struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	batchSize      int
	cycleLength    int64
	readPercent    int
	writeSeq, seed int64
	sequential     bool
	splits         int
	complexity     int
	inverted       bool
	computed       bool
}

func init() {
	workload.Register(jsonLoadMeta)
}

var jsonLoadMeta = workload.Meta{
	Name: `json`,
	Description: `JSON reads and writes to keys spread (by default, uniformly` +
		` at random) across the cluster`,
	Version: `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(694257)
		g := &jsonLoad{}
		g.flags.FlagSet = pflag.NewFlagSet(`json`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`batch`: {RuntimeOnly: true},
		}
		g.flags.IntVar(&g.batchSize, `batch`, 1, `Number of blocks to insert in a single SQL statement`)
		g.flags.Int64Var(&g.cycleLength, `cycle-length`, math.MaxInt64, `Number of keys repeatedly accessed by each writer`)
		g.flags.IntVar(&g.readPercent, `read-percent`, 0, `Percent (0-100) of operations that are reads of existing keys`)
		g.flags.Int64Var(&g.writeSeq, `write-seq`, 0, `Initial write sequence value.`)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.BoolVar(&g.sequential, `sequential`, false, `Pick keys sequentially instead of randomly`)
		g.flags.IntVar(&g.splits, `splits`, 0, `Number of splits to perform before starting normal operations`)
		g.flags.IntVar(&g.complexity, `complexity`, 20, `Complexity of generated JSON data`)
		g.flags.BoolVar(&g.inverted, `inverted`, false, `Whether to include an inverted index`)
		g.flags.BoolVar(&g.computed, `computed`, false, `Whether to use a computed primary key`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*jsonLoad) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(694258)
	return jsonLoadMeta
}

func (w *jsonLoad) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(694259)
	return w.flags
}

func (w *jsonLoad) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(694260)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(694261)
			if w.computed && func() bool {
				__antithesis_instrumentation__.Notify(694263)
				return w.inverted == true
			}() == true {
				__antithesis_instrumentation__.Notify(694264)
				return errors.Errorf("computed and inverted cannot be used together")
			} else {
				__antithesis_instrumentation__.Notify(694265)
			}
			__antithesis_instrumentation__.Notify(694262)
			return nil
		},
	}
}

func (w *jsonLoad) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(694266)
	schema := jsonSchema
	if w.inverted {
		__antithesis_instrumentation__.Notify(694269)
		schema = jsonSchemaWithInvertedIndex
	} else {
		__antithesis_instrumentation__.Notify(694270)
		if w.computed {
			__antithesis_instrumentation__.Notify(694271)
			schema = jsonSchemaWithComputed
		} else {
			__antithesis_instrumentation__.Notify(694272)
		}
	}
	__antithesis_instrumentation__.Notify(694267)
	table := workload.Table{
		Name:   `j`,
		Schema: schema,
		Splits: workload.Tuples(
			w.splits,
			func(splitIdx int) []interface{} {
				__antithesis_instrumentation__.Notify(694273)
				rng := rand.New(rand.NewSource(w.seed + int64(splitIdx)))
				g := newHashGenerator(&sequence{config: w, val: w.writeSeq})
				return []interface{}{
					int(g.hash(rng.Int63())),
				}
			},
		),
	}
	__antithesis_instrumentation__.Notify(694268)
	return []workload.Table{table}
}

func (w *jsonLoad) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(694274)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(694283)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694284)
	}
	__antithesis_instrumentation__.Notify(694275)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(694285)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694286)
	}
	__antithesis_instrumentation__.Notify(694276)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	var buf bytes.Buffer
	buf.WriteString(`SELECT k, v FROM j WHERE k IN (`)
	for i := 0; i < w.batchSize; i++ {
		__antithesis_instrumentation__.Notify(694287)
		if i > 0 {
			__antithesis_instrumentation__.Notify(694289)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(694290)
		}
		__antithesis_instrumentation__.Notify(694288)
		fmt.Fprintf(&buf, `$%d`, i+1)
	}
	__antithesis_instrumentation__.Notify(694277)
	buf.WriteString(`)`)
	readStmt, err := db.Prepare(buf.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(694291)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694292)
	}
	__antithesis_instrumentation__.Notify(694278)

	buf.Reset()
	if w.computed {
		__antithesis_instrumentation__.Notify(694293)
		buf.WriteString(`UPSERT INTO j (v) VALUES`)
	} else {
		__antithesis_instrumentation__.Notify(694294)
		buf.WriteString(`UPSERT INTO j (k, v) VALUES`)
	}
	__antithesis_instrumentation__.Notify(694279)

	for i := 0; i < w.batchSize; i++ {
		__antithesis_instrumentation__.Notify(694295)
		j := i * 2
		if i > 0 {
			__antithesis_instrumentation__.Notify(694297)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(694298)
		}
		__antithesis_instrumentation__.Notify(694296)
		if w.computed {
			__antithesis_instrumentation__.Notify(694299)
			fmt.Fprintf(&buf, ` ($%d)`, i+1)
		} else {
			__antithesis_instrumentation__.Notify(694300)
			fmt.Fprintf(&buf, ` ($%d, $%d)`, j+1, j+2)
		}
	}
	__antithesis_instrumentation__.Notify(694280)

	writeStmt, err := db.Prepare(buf.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(694301)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(694302)
	}
	__antithesis_instrumentation__.Notify(694281)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(694303)
		op := jsonOp{
			config:    w,
			hists:     reg.GetHandle(),
			db:        db,
			readStmt:  readStmt,
			writeStmt: writeStmt,
		}
		seq := &sequence{config: w, val: w.writeSeq}
		if w.sequential {
			__antithesis_instrumentation__.Notify(694305)
			op.g = newSequentialGenerator(seq)
		} else {
			__antithesis_instrumentation__.Notify(694306)
			op.g = newHashGenerator(seq)
		}
		__antithesis_instrumentation__.Notify(694304)
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(694282)
	return ql, nil
}

type jsonOp struct {
	config    *jsonLoad
	hists     *histogram.Histograms
	db        *gosql.DB
	readStmt  *gosql.Stmt
	writeStmt *gosql.Stmt
	g         keyGenerator
}

func (o *jsonOp) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(694307)
	if o.g.rand().Intn(100) < o.config.readPercent {
		__antithesis_instrumentation__.Notify(694311)
		args := make([]interface{}, o.config.batchSize)
		for i := 0; i < o.config.batchSize; i++ {
			__antithesis_instrumentation__.Notify(694315)
			args[i] = o.g.readKey()
		}
		__antithesis_instrumentation__.Notify(694312)
		start := timeutil.Now()
		rows, err := o.readStmt.Query(args...)
		if err != nil {
			__antithesis_instrumentation__.Notify(694316)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694317)
		}
		__antithesis_instrumentation__.Notify(694313)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(694318)
		}
		__antithesis_instrumentation__.Notify(694314)
		elapsed := timeutil.Since(start)
		o.hists.Get(`read`).Record(elapsed)
		return rows.Err()
	} else {
		__antithesis_instrumentation__.Notify(694319)
	}
	__antithesis_instrumentation__.Notify(694308)
	argCount := 2
	if o.config.computed {
		__antithesis_instrumentation__.Notify(694320)
		argCount = 1
	} else {
		__antithesis_instrumentation__.Notify(694321)
	}
	__antithesis_instrumentation__.Notify(694309)
	args := make([]interface{}, argCount*o.config.batchSize)
	for i := 0; i < o.config.batchSize*argCount; i += argCount {
		__antithesis_instrumentation__.Notify(694322)
		j := i
		if !o.config.computed {
			__antithesis_instrumentation__.Notify(694326)
			args[j] = o.g.writeKey()
			j++
		} else {
			__antithesis_instrumentation__.Notify(694327)
		}
		__antithesis_instrumentation__.Notify(694323)
		js, err := json.Random(o.config.complexity, o.g.rand())
		if err != nil {
			__antithesis_instrumentation__.Notify(694328)
			return err
		} else {
			__antithesis_instrumentation__.Notify(694329)
		}
		__antithesis_instrumentation__.Notify(694324)
		if o.config.computed {
			__antithesis_instrumentation__.Notify(694330)
			builder := json.NewObjectBuilder(2)
			builder.Add("key", json.FromInt64(o.g.writeKey()))
			builder.Add("data", js)
			js = builder.Build()
		} else {
			__antithesis_instrumentation__.Notify(694331)
		}
		__antithesis_instrumentation__.Notify(694325)
		args[j] = js.String()
	}
	__antithesis_instrumentation__.Notify(694310)
	start := timeutil.Now()
	_, err := o.writeStmt.Exec(args...)
	elapsed := timeutil.Since(start)
	o.hists.Get(`write`).Record(elapsed)
	return err
}

type sequence struct {
	config *jsonLoad
	val    int64
}

func (s *sequence) write() int64 {
	__antithesis_instrumentation__.Notify(694332)
	return (atomic.AddInt64(&s.val, 1) - 1) % s.config.cycleLength
}

func (s *sequence) read() int64 {
	__antithesis_instrumentation__.Notify(694333)
	return atomic.LoadInt64(&s.val) % s.config.cycleLength
}

type keyGenerator interface {
	writeKey() int64
	readKey() int64
	rand() *rand.Rand
}

type hashGenerator struct {
	seq    *sequence
	random *rand.Rand
	hasher hash.Hash
	buf    [sha1.Size]byte
}

func newHashGenerator(seq *sequence) *hashGenerator {
	__antithesis_instrumentation__.Notify(694334)
	return &hashGenerator{
		seq:    seq,
		random: rand.New(rand.NewSource(seq.config.seed)),
		hasher: sha1.New(),
	}
}

func (g *hashGenerator) hash(v int64) int64 {
	__antithesis_instrumentation__.Notify(694335)
	binary.BigEndian.PutUint64(g.buf[:8], uint64(v))
	binary.BigEndian.PutUint64(g.buf[8:16], uint64(g.seq.config.seed))
	g.hasher.Reset()
	_, _ = g.hasher.Write(g.buf[:16])
	g.hasher.Sum(g.buf[:0])
	return int64(binary.BigEndian.Uint64(g.buf[:8]))
}

func (g *hashGenerator) writeKey() int64 {
	__antithesis_instrumentation__.Notify(694336)
	return g.hash(g.seq.write())
}

func (g *hashGenerator) readKey() int64 {
	__antithesis_instrumentation__.Notify(694337)
	v := g.seq.read()
	if v == 0 {
		__antithesis_instrumentation__.Notify(694339)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(694340)
	}
	__antithesis_instrumentation__.Notify(694338)
	return g.hash(g.random.Int63n(v))
}

func (g *hashGenerator) rand() *rand.Rand {
	__antithesis_instrumentation__.Notify(694341)
	return g.random
}

type sequentialGenerator struct {
	seq    *sequence
	random *rand.Rand
}

func newSequentialGenerator(seq *sequence) *sequentialGenerator {
	__antithesis_instrumentation__.Notify(694342)
	return &sequentialGenerator{
		seq:    seq,
		random: rand.New(rand.NewSource(seq.config.seed)),
	}
}

func (g *sequentialGenerator) writeKey() int64 {
	__antithesis_instrumentation__.Notify(694343)
	return g.seq.write()
}

func (g *sequentialGenerator) readKey() int64 {
	__antithesis_instrumentation__.Notify(694344)
	v := g.seq.read()
	if v == 0 {
		__antithesis_instrumentation__.Notify(694346)
		return 0
	} else {
		__antithesis_instrumentation__.Notify(694347)
	}
	__antithesis_instrumentation__.Notify(694345)
	return g.random.Int63n(v)
}

func (g *sequentialGenerator) rand() *rand.Rand {
	__antithesis_instrumentation__.Notify(694348)
	return g.random
}
