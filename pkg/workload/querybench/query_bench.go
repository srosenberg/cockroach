package querybench

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bufio"
	"context"
	gosql "database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
	"github.com/spf13/pflag"
)

type queryBench struct {
	flags           workload.Flags
	connFlags       *workload.ConnFlags
	queryFile       string
	numRunsPerQuery int
	vectorize       string
	verbose         bool

	queries []string
}

func init() {
	workload.Register(queryBenchMeta)
}

var queryBenchMeta = workload.Meta{
	Name: `querybench`,
	Description: `QueryBench runs queries from the specified file. The queries are run ` +
		`sequentially in each concurrent worker.`,
	Version: `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(695017)
		g := &queryBench{}
		g.flags.FlagSet = pflag.NewFlagSet(`querybench`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`query-file`: {RuntimeOnly: true},
			`optimizer`:  {RuntimeOnly: true},
			`vectorize`:  {RuntimeOnly: true},
			`num-runs`:   {RuntimeOnly: true},
		}
		g.flags.StringVar(&g.queryFile, `query-file`, ``, `File of newline separated queries to run`)
		g.flags.IntVar(&g.numRunsPerQuery, `num-runs`, 0, `Specifies the number of times each query in the query file to be run `+
			`(note that --duration and --max-ops take precedence, so if duration or max-ops is reached, querybench will exit without honoring --num-runs)`)
		g.flags.StringVar(&g.vectorize, `vectorize`, "", `Set vectorize session variable`)
		g.flags.BoolVar(&g.verbose, `verbose`, true, `Prints out the queries being run as well as histograms`)
		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

var vectorizeSetting19_2Translation = map[string]string{
	"on": "experimental_on",
}

func (*queryBench) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(695018)
	return queryBenchMeta
}

func (g *queryBench) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(695019)
	return g.flags
}

func (g *queryBench) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(695020)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(695021)
			if g.queryFile == "" {
				__antithesis_instrumentation__.Notify(695026)
				return errors.Errorf("Missing required argument '--query-file'")
			} else {
				__antithesis_instrumentation__.Notify(695027)
			}
			__antithesis_instrumentation__.Notify(695022)
			queries, err := GetQueries(g.queryFile)
			if err != nil {
				__antithesis_instrumentation__.Notify(695028)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695029)
			}
			__antithesis_instrumentation__.Notify(695023)
			if len(queries) < 1 {
				__antithesis_instrumentation__.Notify(695030)
				return errors.New("no queries found in file")
			} else {
				__antithesis_instrumentation__.Notify(695031)
			}
			__antithesis_instrumentation__.Notify(695024)
			g.queries = queries
			if g.numRunsPerQuery < 0 {
				__antithesis_instrumentation__.Notify(695032)
				return errors.New("negative --num-runs specified")
			} else {
				__antithesis_instrumentation__.Notify(695033)
			}
			__antithesis_instrumentation__.Notify(695025)
			return nil
		},
	}
}

func (*queryBench) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(695034)

	return []workload.Table{}
}

func (g *queryBench) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(695035)
	sqlDatabase, err := workload.SanitizeUrls(g, g.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(695042)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695043)
	}
	__antithesis_instrumentation__.Notify(695036)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(695044)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695045)
	}
	__antithesis_instrumentation__.Notify(695037)

	db.SetMaxOpenConns(g.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(g.connFlags.Concurrency + 1)

	if g.vectorize != "" {
		__antithesis_instrumentation__.Notify(695046)
		_, err := db.Exec("SET vectorize=" + g.vectorize)
		if err != nil && func() bool {
			__antithesis_instrumentation__.Notify(695048)
			return strings.Contains(err.Error(), "invalid value") == true
		}() == true {
			__antithesis_instrumentation__.Notify(695049)
			if _, ok := vectorizeSetting19_2Translation[g.vectorize]; ok {
				__antithesis_instrumentation__.Notify(695050)

				_, err = db.Exec("SET vectorize=" + vectorizeSetting19_2Translation[g.vectorize])
			} else {
				__antithesis_instrumentation__.Notify(695051)
			}
		} else {
			__antithesis_instrumentation__.Notify(695052)
		}
		__antithesis_instrumentation__.Notify(695047)
		if err != nil {
			__antithesis_instrumentation__.Notify(695053)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695054)
		}
	} else {
		__antithesis_instrumentation__.Notify(695055)
	}
	__antithesis_instrumentation__.Notify(695038)

	stmts := make([]namedStmt, len(g.queries))
	for i, query := range g.queries {
		__antithesis_instrumentation__.Notify(695056)
		stmts[i] = namedStmt{

			name: fmt.Sprintf("%2d: %s", i+1, query),
		}
		stmt, err := db.Prepare(query)
		if err != nil {
			__antithesis_instrumentation__.Notify(695058)
			stmts[i].query = query
			continue
		} else {
			__antithesis_instrumentation__.Notify(695059)
		}
		__antithesis_instrumentation__.Notify(695057)
		stmts[i].preparedStmt = stmt
	}
	__antithesis_instrumentation__.Notify(695039)

	maxNumStmts := 0
	if g.numRunsPerQuery > 0 {
		__antithesis_instrumentation__.Notify(695060)
		maxNumStmts = g.numRunsPerQuery * len(g.queries)
	} else {
		__antithesis_instrumentation__.Notify(695061)
	}
	__antithesis_instrumentation__.Notify(695040)

	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	for i := 0; i < g.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(695062)
		op := queryBenchWorker{
			hists:       reg.GetHandle(),
			db:          db,
			stmts:       stmts,
			verbose:     g.verbose,
			maxNumStmts: maxNumStmts,
		}
		ql.WorkerFns = append(ql.WorkerFns, op.run)
	}
	__antithesis_instrumentation__.Notify(695041)
	return ql, nil
}

func GetQueries(path string) ([]string, error) {
	__antithesis_instrumentation__.Notify(695063)
	file, err := os.Open(path)
	if err != nil {
		__antithesis_instrumentation__.Notify(695067)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(695068)
	}
	__antithesis_instrumentation__.Notify(695064)
	defer file.Close()

	scanner := bufio.NewScanner(file)

	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	var lines []string
	for scanner.Scan() {
		__antithesis_instrumentation__.Notify(695069)
		line := scanner.Text()
		if len(line) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(695070)
			return line[0] != '#' == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(695071)
			return !strings.HasPrefix(line, "--") == true
		}() == true {
			__antithesis_instrumentation__.Notify(695072)
			lines = append(lines, line)
		} else {
			__antithesis_instrumentation__.Notify(695073)
		}
	}
	__antithesis_instrumentation__.Notify(695065)
	if err := scanner.Err(); err != nil {
		__antithesis_instrumentation__.Notify(695074)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(695075)
	}
	__antithesis_instrumentation__.Notify(695066)
	return lines, nil
}

type namedStmt struct {
	name string

	preparedStmt *gosql.Stmt
	query        string
}

type queryBenchWorker struct {
	hists *histogram.Histograms
	db    *gosql.DB
	stmts []namedStmt

	stmtIdx int
	verbose bool

	maxNumStmts int
}

func (o *queryBenchWorker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(695076)
	if o.maxNumStmts > 0 {
		__antithesis_instrumentation__.Notify(695081)
		if o.stmtIdx >= o.maxNumStmts {
			__antithesis_instrumentation__.Notify(695082)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(695083)
		}
	} else {
		__antithesis_instrumentation__.Notify(695084)
	}
	__antithesis_instrumentation__.Notify(695077)
	start := timeutil.Now()
	stmt := o.stmts[o.stmtIdx%len(o.stmts)]
	o.stmtIdx++

	exhaustRows := func(execFn func() (*gosql.Rows, error)) error {
		__antithesis_instrumentation__.Notify(695085)
		rows, err := execFn()
		if err != nil {
			__antithesis_instrumentation__.Notify(695089)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695090)
		}
		__antithesis_instrumentation__.Notify(695086)
		defer rows.Close()
		for rows.Next() {
			__antithesis_instrumentation__.Notify(695091)
		}
		__antithesis_instrumentation__.Notify(695087)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(695092)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695093)
		}
		__antithesis_instrumentation__.Notify(695088)
		return nil
	}
	__antithesis_instrumentation__.Notify(695078)
	if stmt.preparedStmt != nil {
		__antithesis_instrumentation__.Notify(695094)
		if err := exhaustRows(func() (*gosql.Rows, error) {
			__antithesis_instrumentation__.Notify(695095)
			return stmt.preparedStmt.Query()
		}); err != nil {
			__antithesis_instrumentation__.Notify(695096)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695097)
		}
	} else {
		__antithesis_instrumentation__.Notify(695098)
		if err := exhaustRows(func() (*gosql.Rows, error) {
			__antithesis_instrumentation__.Notify(695099)
			return o.db.Query(stmt.query)
		}); err != nil {
			__antithesis_instrumentation__.Notify(695100)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695101)
		}
	}
	__antithesis_instrumentation__.Notify(695079)
	elapsed := timeutil.Since(start)
	if o.verbose {
		__antithesis_instrumentation__.Notify(695102)
		o.hists.Get(stmt.name).Record(elapsed)
	} else {
		__antithesis_instrumentation__.Notify(695103)
		o.hists.Get("").Record(elapsed)
	}
	__antithesis_instrumentation__.Notify(695080)
	return nil
}
