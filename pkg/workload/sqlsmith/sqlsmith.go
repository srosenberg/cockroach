package sqlsmith

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/internal/sqlsmith"
	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

type sqlSmith struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	seed          int64
	tables        int
	errorSettings int
}

type errorSettingTypes int

const (
	ignoreExecErrors errorSettingTypes = iota
	returnOnInternalError
	returnOnError
)

func init() {
	workload.Register(sqlSmithMeta)
}

var sqlSmithMeta = workload.Meta{
	Name:        `sqlsmith`,
	Description: `sqlsmith is a random SQL query generator`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(697395)
		g := &sqlSmith{}
		g.flags.FlagSet = pflag.NewFlagSet(`sqlsmith`, pflag.ContinueOnError)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Key hash seed.`)
		g.flags.IntVar(&g.tables, `tables`, 1, `Number of tables.`)
		g.flags.IntVar(&g.errorSettings, `error-sensitivity`, 0,
			`SQLSmith's sensitivity to errors. 0=ignore all errors. 1=quit on internal errors. 2=quit on any error.`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*sqlSmith) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(697396)
	return sqlSmithMeta
}

func (g *sqlSmith) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(697397)
	return g.flags
}

func (g *sqlSmith) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(697398)
	return workload.Hooks{}
}

func (g *sqlSmith) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(697399)
	rng := rand.New(rand.NewSource(g.seed))
	var tables []workload.Table
	for idx := 0; idx < g.tables; idx++ {
		__antithesis_instrumentation__.Notify(697401)
		schema := randgen.RandCreateTable(rng, "table", idx)
		table := workload.Table{
			Name:   schema.Table.String(),
			Schema: tree.Serialize(schema),
		}

		table.Schema = table.Schema[strings.Index(table.Schema, `(`):]
		tables = append(tables, table)
	}
	__antithesis_instrumentation__.Notify(697400)
	return tables
}

func (g *sqlSmith) handleError(err error) error {
	__antithesis_instrumentation__.Notify(697402)
	if err != nil {
		__antithesis_instrumentation__.Notify(697404)
		switch errorSettingTypes(g.errorSettings) {
		case ignoreExecErrors:
			__antithesis_instrumentation__.Notify(697405)
			return nil
		case returnOnInternalError:
			__antithesis_instrumentation__.Notify(697406)
			if strings.Contains(err.Error(), "internal error") {
				__antithesis_instrumentation__.Notify(697409)
				return err
			} else {
				__antithesis_instrumentation__.Notify(697410)
			}
		case returnOnError:
			__antithesis_instrumentation__.Notify(697407)
			return err
		default:
			__antithesis_instrumentation__.Notify(697408)
		}
	} else {
		__antithesis_instrumentation__.Notify(697411)
	}
	__antithesis_instrumentation__.Notify(697403)
	return nil
}

func (g *sqlSmith) validateErrorSetting() error {
	__antithesis_instrumentation__.Notify(697412)
	switch errorSettingTypes(g.errorSettings) {
	case ignoreExecErrors:
		__antithesis_instrumentation__.Notify(697414)
	case returnOnInternalError:
		__antithesis_instrumentation__.Notify(697415)
	case returnOnError:
		__antithesis_instrumentation__.Notify(697416)
	default:
		__antithesis_instrumentation__.Notify(697417)
		return errors.Newf("invalid value for error-sensitivity: %d", g.errorSettings)
	}
	__antithesis_instrumentation__.Notify(697413)
	return nil
}

func (g *sqlSmith) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(697418)
	if err := g.validateErrorSetting(); err != nil {
		__antithesis_instrumentation__.Notify(697423)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697424)
	}
	__antithesis_instrumentation__.Notify(697419)
	sqlDatabase, err := workload.SanitizeUrls(g, g.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(697425)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697426)
	}
	__antithesis_instrumentation__.Notify(697420)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(697427)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(697428)
	}
	__antithesis_instrumentation__.Notify(697421)

	db.SetMaxOpenConns(g.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(g.connFlags.Concurrency + 1)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < g.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(697429)
		rng := rand.New(rand.NewSource(g.seed + int64(i)))
		smither, err := sqlsmith.NewSmither(db, rng)
		if err != nil {
			__antithesis_instrumentation__.Notify(697432)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(697433)
		}
		__antithesis_instrumentation__.Notify(697430)

		hists := reg.GetHandle()
		workerFn := func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(697434)
			start := timeutil.Now()
			query := smither.Generate()
			elapsed := timeutil.Since(start)
			hists.Get(`generate`).Record(elapsed)

			start = timeutil.Now()
			_, err := db.ExecContext(ctx, query)
			if handledErr := g.handleError(err); handledErr != nil {
				__antithesis_instrumentation__.Notify(697436)
				return handledErr
			} else {
				__antithesis_instrumentation__.Notify(697437)
			}
			__antithesis_instrumentation__.Notify(697435)
			elapsed = timeutil.Since(start)

			hists.Get(`exec`).Record(elapsed)

			return nil
		}
		__antithesis_instrumentation__.Notify(697431)
		ql.WorkerFns = append(ql.WorkerFns, workerFn)
	}
	__antithesis_instrumentation__.Notify(697422)
	return ql, nil
}
