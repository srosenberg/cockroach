package ttllogger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/spf13/pflag"
)

type ttlLogger struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	ttl                                time.Duration
	seed                               int64
	minRowsPerInsert, maxRowsPerInsert int
	tsAsPrimaryKey                     bool

	prometheus struct {
		insertedRows prometheus.Counter
	}
}

var ttlLoggerMeta = workload.Meta{
	Name:         "ttllogger",
	Description:  "Generates a simple log table with rows expiring after the given TTL.",
	Version:      "0.0.1",
	PublicFacing: true,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(698926)
		g := &ttlLogger{}
		g.flags.FlagSet = pflag.NewFlagSet(`ttllogger`, pflag.ContinueOnError)
		g.flags.DurationVar(&g.ttl, "ttl", time.Minute, `duration for the TTL to expire`)
		g.flags.Int64Var(&g.seed, `seed`, 1, `seed for randomization operations`)
		g.flags.IntVar(&g.minRowsPerInsert, `min-rows-per-insert`, 1, `minimum rows per insert per query`)
		g.flags.IntVar(&g.maxRowsPerInsert, `max-rows-per-insert`, 100, `maximum rows per insert per query`)
		g.flags.BoolVar(&g.tsAsPrimaryKey, `ts-as-primary-key`, true, `whether timestamp column for the table should be part of the primary key`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func init() {
	workload.Register(ttlLoggerMeta)
}

func (l ttlLogger) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(698927)
	return workload.Hooks{}
}

func (l *ttlLogger) setupMetrics(reg prometheus.Registerer) {
	__antithesis_instrumentation__.Notify(698928)
	p := promauto.With(reg)
	l.prometheus.insertedRows = p.NewCounter(
		prometheus.CounterOpts{
			Namespace: histogram.PrometheusNamespace,
			Subsystem: ttlLoggerMeta.Name,
			Name:      "rows_inserted",
			Help:      "Number of rows inserted.",
		},
	)
}

var logChars = []rune("abdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ !.")

func (l *ttlLogger) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(698929)
	sqlDatabase, err := workload.SanitizeUrls(l, l.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(698937)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698938)
	}
	__antithesis_instrumentation__.Notify(698930)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(698939)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698940)
	}
	__antithesis_instrumentation__.Notify(698931)

	db.SetMaxOpenConns(l.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(l.connFlags.Concurrency + 1)

	insertStmt, err := db.Prepare(`
		INSERT INTO logs (message) (SELECT ($2 || s) FROM generate_series(1, $1) s)`,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(698941)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698942)
	}
	__antithesis_instrumentation__.Notify(698932)
	selectElemSQL := `SELECT * FROM logs WHERE ts >= now() - $1::interval LIMIT 1`
	if !l.tsAsPrimaryKey {
		__antithesis_instrumentation__.Notify(698943)
		selectElemSQL = `SELECT * FROM logs WHERE id >= $1::string LIMIT 1`
	} else {
		__antithesis_instrumentation__.Notify(698944)
	}
	__antithesis_instrumentation__.Notify(698933)
	selectElemStmt, err := db.Prepare(selectElemSQL)
	if err != nil {
		__antithesis_instrumentation__.Notify(698945)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(698946)
	}
	__antithesis_instrumentation__.Notify(698934)

	if l.connFlags.Concurrency%2 != 0 {
		__antithesis_instrumentation__.Notify(698947)
		return workload.QueryLoad{}, errors.Newf("concurrency must be divisible by 2")
	} else {
		__antithesis_instrumentation__.Notify(698948)
	}
	__antithesis_instrumentation__.Notify(698935)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for len(ql.WorkerFns) < l.connFlags.Concurrency {
		__antithesis_instrumentation__.Notify(698949)
		rng := rand.New(rand.NewSource(l.seed + int64(len(ql.WorkerFns))))
		hists := reg.GetHandle()
		workerFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(698952)
			strLen := 1 + rng.Intn(100)
			str := make([]rune, strLen)
			for i := 0; i < strLen; i++ {
				__antithesis_instrumentation__.Notify(698954)
				str[i] = logChars[rand.Intn(len(logChars))]
			}
			__antithesis_instrumentation__.Notify(698953)
			rowsToInsert := l.minRowsPerInsert + rng.Intn(l.maxRowsPerInsert-l.minRowsPerInsert)

			start := timeutil.Now()
			_, err := insertStmt.Exec(rowsToInsert, string(str))
			elapsed := timeutil.Since(start)
			hists.Get(`log`).Record(elapsed)
			l.prometheus.insertedRows.Add(float64(rowsToInsert))
			return err
		}
		__antithesis_instrumentation__.Notify(698950)
		selectFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(698955)
			start := timeutil.Now()
			var placeholder interface{}
			placeholder = l.ttl / 2
			if !l.tsAsPrimaryKey {
				__antithesis_instrumentation__.Notify(698957)
				id := uuid.MakeV4()
				id.DeterministicV4(uint64(rng.Int63()), uint64(1<<63))
				placeholder = id.String()
			} else {
				__antithesis_instrumentation__.Notify(698958)
			}
			__antithesis_instrumentation__.Notify(698956)
			_, err := selectElemStmt.Exec(placeholder)
			elapsed := timeutil.Since(start)
			hists.Get(`select`).Record(elapsed)
			return err
		}
		__antithesis_instrumentation__.Notify(698951)
		ql.WorkerFns = append(ql.WorkerFns, selectFn, workerFn)
	}
	__antithesis_instrumentation__.Notify(698936)
	l.setupMetrics(reg.Registerer())
	return ql, nil
}

func (l ttlLogger) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(698959)
	return ttlLoggerMeta
}

func (l ttlLogger) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(698960)
	pk := `PRIMARY KEY (ts, id)`
	if !l.tsAsPrimaryKey {
		__antithesis_instrumentation__.Notify(698962)
		pk = `PRIMARY KEY (id)`
	} else {
		__antithesis_instrumentation__.Notify(698963)
	}
	__antithesis_instrumentation__.Notify(698961)
	return []workload.Table{
		{
			Name: "logs",
			Schema: fmt.Sprintf(`(
	ts TIMESTAMPTZ NOT NULL DEFAULT current_timestamp(),
	id TEXT NOT NULL DEFAULT gen_random_uuid()::string,
	message TEXT NOT NULL,
	%s
) WITH (ttl_expire_after = '%s', ttl_label_metrics = true, ttl_row_stats_poll_interval = '15s', ttl_job_cron = '* * * * *')`,
				pk,
				l.ttl.String(),
			),
		},
	}
}

func (l ttlLogger) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(698964)
	return l.flags
}
