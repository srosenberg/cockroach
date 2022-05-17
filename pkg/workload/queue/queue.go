package queue

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/spf13/pflag"
)

const (
	queueSchema = `(ts BIGINT NOT NULL, id BIGINT NOT NULL, PRIMARY KEY(ts, id))`
)

type queue struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags
	batchSize int
}

func init() {
	workload.Register(queueMeta)
}

var queueMeta = workload.Meta{
	Name: `queue`,
	Description: `A simple queue-like application load: inserts into a table in sequence ` +
		`(ordered by primary key), followed by the deletion of inserted rows starting from the ` +
		`beginning of the sequence.`,
	Version: `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(695544)
		g := &queue{}
		g.flags.FlagSet = pflag.NewFlagSet(`queue`, pflag.ContinueOnError)
		g.connFlags = workload.NewConnFlags(&g.flags)
		g.flags.IntVar(&g.batchSize, `batch`, 1, `Number of blocks to insert in a single SQL statement`)
		return g
	},
}

func (*queue) Meta() workload.Meta { __antithesis_instrumentation__.Notify(695545); return queueMeta }

func (w *queue) Flags() workload.Flags { __antithesis_instrumentation__.Notify(695546); return w.flags }

func (w *queue) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(695547)
	table := workload.Table{
		Name:   `queue`,
		Schema: queueSchema,
	}
	return []workload.Table{table}
}

func (w *queue) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(695548)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(695555)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695556)
	}
	__antithesis_instrumentation__.Notify(695549)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(695557)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695558)
	}
	__antithesis_instrumentation__.Notify(695550)
	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	var buf bytes.Buffer
	buf.WriteString(`INSERT INTO queue (ts, id) VALUES`)
	for i := 0; i < w.batchSize; i++ {
		__antithesis_instrumentation__.Notify(695559)
		j := i * 2
		if i > 0 {
			__antithesis_instrumentation__.Notify(695561)
			buf.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(695562)
		}
		__antithesis_instrumentation__.Notify(695560)
		fmt.Fprintf(&buf, ` ($%d, $%d)`, j+1, j+2)
	}
	__antithesis_instrumentation__.Notify(695551)
	insertStmt, err := db.Prepare(buf.String())
	if err != nil {
		__antithesis_instrumentation__.Notify(695563)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695564)
	}
	__antithesis_instrumentation__.Notify(695552)

	deleteStmt, err := db.Prepare(`DELETE FROM queue WHERE ts < $1`)
	if err != nil {
		__antithesis_instrumentation__.Notify(695565)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695566)
	}
	__antithesis_instrumentation__.Notify(695553)

	seqFunc := makeSequenceFunc()

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(695567)
		op := queueOp{
			workerID:   i + 1,
			config:     w,
			hists:      reg.GetHandle(),
			db:         db,
			insertStmt: insertStmt,
			deleteStmt: deleteStmt,
			getSeq:     seqFunc,
		}
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(695554)
	return ql, nil
}

type queueOp struct {
	workerID   int
	config     *queue
	hists      *histogram.Histograms
	db         *gosql.DB
	insertStmt *gosql.Stmt
	deleteStmt *gosql.Stmt
	getSeq     func() int
}

func (o *queueOp) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(695568)
	count := o.getSeq()
	start := count * o.config.batchSize
	end := start + o.config.batchSize

	params := make([]interface{}, 2*o.config.batchSize)
	for i := 0; i < o.config.batchSize; i++ {
		__antithesis_instrumentation__.Notify(695571)
		paramOffset := i * 2
		params[paramOffset+0] = start + i
		params[paramOffset+1] = o.workerID
	}
	__antithesis_instrumentation__.Notify(695569)
	startTime := timeutil.Now()
	_, err := o.insertStmt.Exec(params...)
	if err != nil {
		__antithesis_instrumentation__.Notify(695572)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695573)
	}
	__antithesis_instrumentation__.Notify(695570)
	elapsed := timeutil.Since(startTime)
	o.hists.Get("write").Record(elapsed)

	startTime = timeutil.Now()
	_, err = o.deleteStmt.Exec(end)
	elapsed = timeutil.Since(startTime)
	o.hists.Get(`delete`).Record(elapsed)
	return err
}

func makeSequenceFunc() func() int {
	__antithesis_instrumentation__.Notify(695574)
	i := int64(0)
	return func() int {
		__antithesis_instrumentation__.Notify(695575)
		return int(atomic.AddInt64(&i, 1))
	}
}
