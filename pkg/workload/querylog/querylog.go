package querylog

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"archive/zip"
	"bufio"
	"context"
	gosql "database/sql"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/randgen"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/cockroach/pkg/workload"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	workloadrand "github.com/cockroachdb/cockroach/pkg/workload/rand"
	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"
	"github.com/lib/pq/oid"
	"github.com/spf13/pflag"
)

type querylog struct {
	flags     workload.Flags
	connFlags *workload.ConnFlags

	state querylogState

	dirPath            string
	zipPath            string
	filesToParse       int
	minSamplingProb    float64
	nullPct            int
	numSamples         int
	omitDeleteQueries  bool
	omitWriteQueries   bool
	probOfUsingSamples float64

	querybenchPath string
	count          uint
	seed           int64

	stmtTimeoutSeconds int64
	verbose            bool
}

type querylogState struct {
	tableNames             []string
	totalQueryCount        int
	queryCountPerTable     []int
	seenQueriesByTableName []map[string]int
	tableUsed              map[string]bool
	columnsByTableName     map[string][]columnInfo
}

func init() {
	workload.Register(querylogMeta)
}

var querylogMeta = workload.Meta{
	Name:        `querylog`,
	Description: `Querylog is a tool that produces a workload based on the provided query log.`,
	Version:     `1.0.0`,
	New: func() workload.Generator {
		__antithesis_instrumentation__.Notify(695104)
		g := &querylog{}
		g.flags.FlagSet = pflag.NewFlagSet(`querylog`, pflag.ContinueOnError)
		g.flags.Meta = map[string]workload.FlagMeta{
			`dir`:                   {RuntimeOnly: true},
			`files-to-parse`:        {RuntimeOnly: true},
			`min-sampling-prob`:     {RuntimeOnly: true},
			`null-percent`:          {RuntimeOnly: true},
			`num-samples`:           {RuntimeOnly: true},
			`omit-delete-queries`:   {RuntimeOnly: true},
			`omit-write-queries`:    {RuntimeOnly: true},
			`prob-of-using-samples`: {RuntimeOnly: true},
			`seed`:                  {RuntimeOnly: true},
			`statement-timeout`:     {RuntimeOnly: true},
			`verbose`:               {RuntimeOnly: true},
			`zip`:                   {RuntimeOnly: true},
		}
		g.flags.UintVar(&g.count, `count`, 100, `Number of queries to be written for querybench (used only if --querybench-path is specified).`)
		g.flags.StringVar(&g.dirPath, `dir`, ``, `Directory of the querylog files.`)
		g.flags.IntVar(&g.filesToParse, `files-to-parse`, 5, `Maximum number of files in the query log to process.`)
		g.flags.Float64Var(&g.minSamplingProb, `min-sampling-prob`, 0.01, `Minimum sampling probability defines the minimum chance `+
			`that a value will be chosen as a sample. The smaller the number is the more diverse samples will be (given that the table has enough values). `+
			`However, at the same time, the sampling process will be slower.`)
		g.flags.IntVar(&g.nullPct, `null-percent`, 5, `Percent random nulls.`)
		g.flags.IntVar(&g.numSamples, `num-samples`, 1000, `Number of samples to be taken from the tables. The bigger this number, `+
			`the more diverse values will be chosen for the generated queries.`)
		g.flags.BoolVar(&g.omitDeleteQueries, `omit-delete-queries`, true, `Indicates whether delete queries should be omitted.`)
		g.flags.BoolVar(&g.omitWriteQueries, `omit-write-queries`, true, `Indicates whether write queries (INSERTs and UPSERTs) should be omitted.`)
		g.flags.Float64Var(&g.probOfUsingSamples, `prob-of-using-samples`, 0.9, `Probability of using samples to generate values for `+
			`the placeholders. Say it is 0.9, then with 0.1 probability the values will be generated randomly.`)

		g.flags.StringVar(&g.querybenchPath, `querybench-path`, ``, `Path to write the generated queries to for querybench tool. `+
			`NOTE: at the moment --max-ops=1 is the best way to terminate the generator to produce the desired count of queries.`)
		g.flags.Int64Var(&g.seed, `seed`, 1, `Random number generator seed.`)
		g.flags.Int64Var(&g.stmtTimeoutSeconds, `statement-timeout`, 0, `Sets session's statement_timeout setting (in seconds).'`)
		g.flags.BoolVar(&g.verbose, `verbose`, false, `Indicates whether error messages should be printed out.`)
		g.flags.StringVar(&g.zipPath, `zip`, ``, `Path to the zip with the query log. Note: this zip will be extracted into a temporary `+
			`directory at the same path as zip just without '.zip' extension which will be removed after parsing is complete.`)

		g.connFlags = workload.NewConnFlags(&g.flags)
		return g
	},
}

func (*querylog) Meta() workload.Meta {
	__antithesis_instrumentation__.Notify(695105)
	return querylogMeta
}

func (w *querylog) Flags() workload.Flags {
	__antithesis_instrumentation__.Notify(695106)
	return w.flags
}

func (*querylog) Tables() []workload.Table {
	__antithesis_instrumentation__.Notify(695107)

	return []workload.Table{}
}

func (w *querylog) Hooks() workload.Hooks {
	__antithesis_instrumentation__.Notify(695108)
	return workload.Hooks{
		Validate: func() error {
			__antithesis_instrumentation__.Notify(695109)
			if w.zipPath == "" && func() bool {
				__antithesis_instrumentation__.Notify(695116)
				return w.dirPath == "" == true
			}() == true {
				__antithesis_instrumentation__.Notify(695117)
				return errors.Errorf("Missing required argument: either `--zip` or `--dir` have to be specified.")
			} else {
				__antithesis_instrumentation__.Notify(695118)
			}
			__antithesis_instrumentation__.Notify(695110)
			if w.zipPath != "" {
				__antithesis_instrumentation__.Notify(695119)
				if w.zipPath[len(w.zipPath)-4:] != ".zip" {
					__antithesis_instrumentation__.Notify(695120)
					return errors.Errorf("Illegal argument: `--zip` is expected to end with '.zip'.")
				} else {
					__antithesis_instrumentation__.Notify(695121)
				}
			} else {
				__antithesis_instrumentation__.Notify(695122)
			}
			__antithesis_instrumentation__.Notify(695111)
			if w.minSamplingProb < 0.00000001 || func() bool {
				__antithesis_instrumentation__.Notify(695123)
				return w.minSamplingProb > 1.0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(695124)
				return errors.Errorf("Illegal argument: `--min-sampling-prob` must be in [0.00000001, 1.0] range.")
			} else {
				__antithesis_instrumentation__.Notify(695125)
			}
			__antithesis_instrumentation__.Notify(695112)
			if w.nullPct < 0 || func() bool {
				__antithesis_instrumentation__.Notify(695126)
				return w.nullPct > 100 == true
			}() == true {
				__antithesis_instrumentation__.Notify(695127)
				return errors.Errorf("Illegal argument: `--null-pct` must be in [0, 100] range.")
			} else {
				__antithesis_instrumentation__.Notify(695128)
			}
			__antithesis_instrumentation__.Notify(695113)
			if w.probOfUsingSamples < 0 || func() bool {
				__antithesis_instrumentation__.Notify(695129)
				return w.probOfUsingSamples > 1.0 == true
			}() == true {
				__antithesis_instrumentation__.Notify(695130)
				return errors.Errorf("Illegal argument: `--prob-of-using-samples` must be in [0.0, 1.0] range.")
			} else {
				__antithesis_instrumentation__.Notify(695131)
			}
			__antithesis_instrumentation__.Notify(695114)
			if w.stmtTimeoutSeconds < 0 {
				__antithesis_instrumentation__.Notify(695132)
				return errors.Errorf("Illegal argument: `--statement-timeout` must be a non-negative integer.")
			} else {
				__antithesis_instrumentation__.Notify(695133)
			}
			__antithesis_instrumentation__.Notify(695115)
			return nil
		},
	}
}

func (w *querylog) Ops(
	ctx context.Context, urls []string, reg *histogram.Registry,
) (workload.QueryLoad, error) {
	__antithesis_instrumentation__.Notify(695134)
	sqlDatabase, err := workload.SanitizeUrls(w, w.connFlags.DBOverride, urls)
	if err != nil {
		__antithesis_instrumentation__.Notify(695145)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695146)
	}
	__antithesis_instrumentation__.Notify(695135)
	db, err := gosql.Open(`cockroach`, strings.Join(urls, ` `))
	if err != nil {
		__antithesis_instrumentation__.Notify(695147)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695148)
	}
	__antithesis_instrumentation__.Notify(695136)

	db.SetMaxOpenConns(w.connFlags.Concurrency + 1)
	db.SetMaxIdleConns(w.connFlags.Concurrency + 1)

	if err = w.getTableNames(db); err != nil {
		__antithesis_instrumentation__.Notify(695149)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695150)
	}
	__antithesis_instrumentation__.Notify(695137)

	if err = w.processQueryLog(ctx); err != nil {
		__antithesis_instrumentation__.Notify(695151)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695152)
	}
	__antithesis_instrumentation__.Notify(695138)

	if err = w.getColumnsInfo(db); err != nil {
		__antithesis_instrumentation__.Notify(695153)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695154)
	}
	__antithesis_instrumentation__.Notify(695139)

	if err = w.populateSamples(ctx, db); err != nil {
		__antithesis_instrumentation__.Notify(695155)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695156)
	}
	__antithesis_instrumentation__.Notify(695140)

	if len(urls) != 1 {
		__antithesis_instrumentation__.Notify(695157)
		return workload.QueryLoad{}, errors.Errorf(
			"Exactly one connection string is supported at the moment.")
	} else {
		__antithesis_instrumentation__.Notify(695158)
	}
	__antithesis_instrumentation__.Notify(695141)
	connCfg, err := pgx.ParseConfig(urls[0])
	if err != nil {
		__antithesis_instrumentation__.Notify(695159)
		return workload.QueryLoad{}, err
	} else {
		__antithesis_instrumentation__.Notify(695160)
	}
	__antithesis_instrumentation__.Notify(695142)
	ql := workload.QueryLoad{SQLDatabase: sqlDatabase}
	if w.querybenchPath != `` {
		__antithesis_instrumentation__.Notify(695161)
		conn, err := pgx.ConnectConfig(ctx, connCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(695163)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695164)
		}
		__antithesis_instrumentation__.Notify(695162)
		worker := newQuerybenchWorker(w, reg, conn, 0)
		ql.WorkerFns = append(ql.WorkerFns, worker.querybenchRun)
		return ql, nil
	} else {
		__antithesis_instrumentation__.Notify(695165)
	}
	__antithesis_instrumentation__.Notify(695143)
	for i := 0; i < w.connFlags.Concurrency; i++ {
		__antithesis_instrumentation__.Notify(695166)
		conn, err := pgx.ConnectConfig(ctx, connCfg)
		if err != nil {
			__antithesis_instrumentation__.Notify(695168)
			return workload.QueryLoad{}, err
		} else {
			__antithesis_instrumentation__.Notify(695169)
		}
		__antithesis_instrumentation__.Notify(695167)
		worker := newWorker(w, reg, conn, i)
		ql.WorkerFns = append(ql.WorkerFns, worker.run)
	}
	__antithesis_instrumentation__.Notify(695144)
	return ql, nil
}

type worker struct {
	config *querylog
	hists  *histogram.Histograms

	conn *pgx.Conn
	id   int
	rng  *rand.Rand

	reWriteQuery *regexp.Regexp

	querybenchPath string
}

func newWorker(q *querylog, reg *histogram.Registry, conn *pgx.Conn, id int) *worker {
	__antithesis_instrumentation__.Notify(695170)
	return &worker{
		config:       q,
		hists:        reg.GetHandle(),
		conn:         conn,
		id:           id,
		rng:          rand.New(rand.NewSource(q.seed + int64(id))),
		reWriteQuery: regexp.MustCompile(regexWriteQueryPattern),
	}
}

func newQuerybenchWorker(q *querylog, reg *histogram.Registry, conn *pgx.Conn, id int) *worker {
	__antithesis_instrumentation__.Notify(695171)
	w := newWorker(q, reg, conn, id)
	w.querybenchPath = q.querybenchPath
	return w
}

func (w *worker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(695172)
	if w.config.stmtTimeoutSeconds != 0 {
		__antithesis_instrumentation__.Notify(695174)
		if _, err := w.conn.Exec(ctx, fmt.Sprintf("SET statement_timeout='%ds'", w.config.stmtTimeoutSeconds)); err != nil {
			__antithesis_instrumentation__.Notify(695175)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695176)
		}
	} else {
		__antithesis_instrumentation__.Notify(695177)
	}
	__antithesis_instrumentation__.Notify(695173)

	var start time.Time
	for {
		__antithesis_instrumentation__.Notify(695178)
		chosenQuery, tableName := w.chooseQuery()
		pholdersColumnNames, numRepeats, err := w.deduceColumnNamesForPlaceholders(ctx, chosenQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(695183)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695185)
				log.Infof(ctx, "Encountered an error %s while deducing column names corresponding to the placeholders", err.Error())
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695186)
			}
			__antithesis_instrumentation__.Notify(695184)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695187)
		}
		__antithesis_instrumentation__.Notify(695179)

		placeholders, err := w.generatePlaceholders(ctx, chosenQuery, pholdersColumnNames, numRepeats, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(695188)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695190)
				log.Infof(ctx, "Encountered an error %s while generating values for the placeholders", err.Error())
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695191)
			}
			__antithesis_instrumentation__.Notify(695189)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695192)
		}
		__antithesis_instrumentation__.Notify(695180)

		start = timeutil.Now()
		rows, err := w.conn.Query(ctx, chosenQuery, placeholders...)
		if err != nil {
			__antithesis_instrumentation__.Notify(695193)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695195)
				log.Infof(ctx, "Encountered an error %s while executing the query", err.Error())
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695196)
			}
			__antithesis_instrumentation__.Notify(695194)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695197)
		}
		__antithesis_instrumentation__.Notify(695181)

		for rows.Next() {
			__antithesis_instrumentation__.Notify(695198)
		}
		__antithesis_instrumentation__.Notify(695182)
		rows.Close()
		elapsed := timeutil.Since(start)

		w.hists.Get("").Record(elapsed)
	}
}

func (w *worker) chooseQuery() (chosenQuery string, tableName string) {
	__antithesis_instrumentation__.Notify(695199)
	prob := w.rng.Float64()
	count := 0
	for tableIdx, tableName := range w.config.state.tableNames {
		__antithesis_instrumentation__.Notify(695202)
		if probInInterval(count, w.config.state.queryCountPerTable[tableIdx], w.config.state.totalQueryCount, prob) {
			__antithesis_instrumentation__.Notify(695204)
			countForThisTable := 0
			totalForThisTable := w.config.state.queryCountPerTable[tableIdx]
			for query, frequency := range w.config.state.seenQueriesByTableName[tableIdx] {
				__antithesis_instrumentation__.Notify(695205)
				if probInInterval(countForThisTable, frequency, totalForThisTable, prob) {
					__antithesis_instrumentation__.Notify(695207)
					return query, tableName
				} else {
					__antithesis_instrumentation__.Notify(695208)
				}
				__antithesis_instrumentation__.Notify(695206)
				countForThisTable += frequency
			}
		} else {
			__antithesis_instrumentation__.Notify(695209)
		}
		__antithesis_instrumentation__.Notify(695203)
		count += w.config.state.queryCountPerTable[tableIdx]
	}
	__antithesis_instrumentation__.Notify(695200)

	for tableIdx, tableName := range w.config.state.tableNames {
		__antithesis_instrumentation__.Notify(695210)
		if w.config.state.queryCountPerTable[tableIdx] > 0 {
			__antithesis_instrumentation__.Notify(695211)
			for query := range w.config.state.seenQueriesByTableName[tableIdx] {
				__antithesis_instrumentation__.Notify(695212)
				return query, tableName
			}
		} else {
			__antithesis_instrumentation__.Notify(695213)
		}
	}
	__antithesis_instrumentation__.Notify(695201)
	panic("no queries were accumulated from the log")
}

func (w *worker) deduceColumnNamesForPlaceholders(
	ctx context.Context, query string,
) (pholdersColumnNames []string, numRepeats int, err error) {
	__antithesis_instrumentation__.Notify(695214)
	if isInsertOrUpsert(query) {
		__antithesis_instrumentation__.Notify(695217)

		if !w.reWriteQuery.Match([]byte(query)) {
			__antithesis_instrumentation__.Notify(695221)
			return nil, 0, errors.Errorf("Chosen write query didn't match the pattern.")
		} else {
			__antithesis_instrumentation__.Notify(695222)
		}
		__antithesis_instrumentation__.Notify(695218)
		submatch := w.reWriteQuery.FindSubmatch([]byte(query))
		columnsIdx := 3
		columnsNames := string(submatch[columnsIdx])
		pholdersColumnNames = strings.FieldsFunc(columnsNames, func(r rune) bool {
			__antithesis_instrumentation__.Notify(695223)
			return r == ' ' || func() bool {
				__antithesis_instrumentation__.Notify(695224)
				return r == ',' == true
			}() == true
		})
		__antithesis_instrumentation__.Notify(695219)
		lastPholderIdx := strings.LastIndex(query, "$")
		if lastPholderIdx == -1 {
			__antithesis_instrumentation__.Notify(695225)
			return nil, 0, errors.Errorf("Unexpected: no placeholders in the write query.")
		} else {
			__antithesis_instrumentation__.Notify(695226)
		}
		__antithesis_instrumentation__.Notify(695220)
		multipleValuesGroupIdx := 4

		numRepeats = 1 + strings.Count(string(submatch[multipleValuesGroupIdx]), "(")
		return pholdersColumnNames, numRepeats, nil
	} else {
		__antithesis_instrumentation__.Notify(695227)
	}
	__antithesis_instrumentation__.Notify(695215)

	pholdersColumnNames = make([]string, 0)
	for i := 1; ; i++ {
		__antithesis_instrumentation__.Notify(695228)
		pholder := fmt.Sprintf("$%d", i)
		pholderIdx := strings.Index(query, pholder)
		if pholderIdx == -1 {
			__antithesis_instrumentation__.Notify(695232)

			break
		} else {
			__antithesis_instrumentation__.Notify(695233)
		}
		__antithesis_instrumentation__.Notify(695229)
		tokens := strings.Fields(query[:pholderIdx])
		if len(tokens) < 2 {
			__antithesis_instrumentation__.Notify(695234)
			return nil, 0, errors.Errorf("assumption that there are at least two tokens before placeholder is wrong")
		} else {
			__antithesis_instrumentation__.Notify(695235)
		}
		__antithesis_instrumentation__.Notify(695230)
		column := tokens[len(tokens)-2]
		for column[0] == '(' {
			__antithesis_instrumentation__.Notify(695236)
			column = column[1:]
		}
		__antithesis_instrumentation__.Notify(695231)
		pholdersColumnNames = append(pholdersColumnNames, column)
	}
	__antithesis_instrumentation__.Notify(695216)
	return pholdersColumnNames, 1, nil
}

func (w *worker) generatePlaceholders(
	ctx context.Context, query string, pholdersColumnNames []string, numRepeats int, tableName string,
) (placeholders []interface{}, err error) {
	__antithesis_instrumentation__.Notify(695237)
	isWriteQuery := isInsertOrUpsert(query)
	placeholders = make([]interface{}, 0, len(pholdersColumnNames)*numRepeats)
	for j := 0; j < numRepeats; j++ {
		__antithesis_instrumentation__.Notify(695239)
		for i, column := range pholdersColumnNames {
			__antithesis_instrumentation__.Notify(695240)
			columnMatched := false
			actualTableName := true
			if strings.Contains(column, ".") {
				__antithesis_instrumentation__.Notify(695244)
				actualTableName = false
				column = strings.Split(column, ".")[1]
			} else {
				__antithesis_instrumentation__.Notify(695245)
			}
			__antithesis_instrumentation__.Notify(695241)
			possibleTableNames := make([]string, 0, 1)
			possibleTableNames = append(possibleTableNames, tableName)
			if !actualTableName {
				__antithesis_instrumentation__.Notify(695246)

				for _, n := range w.config.state.tableNames {
					__antithesis_instrumentation__.Notify(695247)
					if n == tableName || func() bool {
						__antithesis_instrumentation__.Notify(695249)
						return !w.config.state.tableUsed[n] == true
					}() == true {
						__antithesis_instrumentation__.Notify(695250)
						continue
					} else {
						__antithesis_instrumentation__.Notify(695251)
					}
					__antithesis_instrumentation__.Notify(695248)
					possibleTableNames = append(possibleTableNames, n)
				}
			} else {
				__antithesis_instrumentation__.Notify(695252)
			}
			__antithesis_instrumentation__.Notify(695242)
			for _, tableName := range possibleTableNames {
				__antithesis_instrumentation__.Notify(695253)
				if columnMatched {
					__antithesis_instrumentation__.Notify(695255)
					break
				} else {
					__antithesis_instrumentation__.Notify(695256)
				}
				__antithesis_instrumentation__.Notify(695254)
				for _, c := range w.config.state.columnsByTableName[tableName] {
					__antithesis_instrumentation__.Notify(695257)
					if c.name == column {
						__antithesis_instrumentation__.Notify(695258)
						if w.rng.Float64() < w.config.probOfUsingSamples && func() bool {
							__antithesis_instrumentation__.Notify(695260)
							return c.samples != nil == true
						}() == true && func() bool {
							__antithesis_instrumentation__.Notify(695261)
							return !isWriteQuery == true
						}() == true {
							__antithesis_instrumentation__.Notify(695262)

							sampleIdx := w.rng.Intn(len(c.samples))
							placeholders = append(placeholders, c.samples[sampleIdx])
						} else {
							__antithesis_instrumentation__.Notify(695263)

							nullPct := 0
							if c.isNullable && func() bool {
								__antithesis_instrumentation__.Notify(695267)
								return w.config.nullPct > 0 == true
							}() == true {
								__antithesis_instrumentation__.Notify(695268)
								nullPct = 100 / w.config.nullPct
							} else {
								__antithesis_instrumentation__.Notify(695269)
							}
							__antithesis_instrumentation__.Notify(695264)
							d := randgen.RandDatumWithNullChance(w.rng, c.dataType, nullPct)
							if i, ok := d.(*tree.DInt); ok && func() bool {
								__antithesis_instrumentation__.Notify(695270)
								return c.intRange > 0 == true
							}() == true {
								__antithesis_instrumentation__.Notify(695271)
								j := int64(*i) % int64(c.intRange/2)
								d = tree.NewDInt(tree.DInt(j))
							} else {
								__antithesis_instrumentation__.Notify(695272)
							}
							__antithesis_instrumentation__.Notify(695265)
							p, err := workloadrand.DatumToGoSQL(d)
							if err != nil {
								__antithesis_instrumentation__.Notify(695273)
								return nil, err
							} else {
								__antithesis_instrumentation__.Notify(695274)
							}
							__antithesis_instrumentation__.Notify(695266)
							placeholders = append(placeholders, p)
						}
						__antithesis_instrumentation__.Notify(695259)
						columnMatched = true
						break
					} else {
						__antithesis_instrumentation__.Notify(695275)
					}
				}
			}
			__antithesis_instrumentation__.Notify(695243)
			if !columnMatched {
				__antithesis_instrumentation__.Notify(695276)
				d := w.rng.Int31n(10) + 1
				if w.config.verbose {
					__antithesis_instrumentation__.Notify(695279)
					log.Infof(ctx, "Couldn't deduce the corresponding to $%d, so generated %d (a small int)", i+1, d)
					printQueryShortened(ctx, query)
				} else {
					__antithesis_instrumentation__.Notify(695280)
				}
				__antithesis_instrumentation__.Notify(695277)
				p, err := workloadrand.DatumToGoSQL(tree.NewDInt(tree.DInt(d)))
				if err != nil {
					__antithesis_instrumentation__.Notify(695281)
					return nil, err
				} else {
					__antithesis_instrumentation__.Notify(695282)
				}
				__antithesis_instrumentation__.Notify(695278)
				placeholders = append(placeholders, p)
			} else {
				__antithesis_instrumentation__.Notify(695283)
			}
		}
	}
	__antithesis_instrumentation__.Notify(695238)
	return placeholders, nil
}

func (w *querylog) getTableNames(db *gosql.DB) (retErr error) {
	__antithesis_instrumentation__.Notify(695284)
	rows, err := db.Query(`SELECT table_name FROM [SHOW TABLES] ORDER BY table_name`)
	if err != nil {
		__antithesis_instrumentation__.Notify(695288)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695289)
	}
	__antithesis_instrumentation__.Notify(695285)
	defer func() {
		__antithesis_instrumentation__.Notify(695290)
		retErr = errors.CombineErrors(retErr, rows.Close())
	}()
	__antithesis_instrumentation__.Notify(695286)
	w.state.tableNames = make([]string, 0)
	for rows.Next() {
		__antithesis_instrumentation__.Notify(695291)
		var tableName string
		if err = rows.Scan(&tableName); err != nil {
			__antithesis_instrumentation__.Notify(695293)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695294)
		}
		__antithesis_instrumentation__.Notify(695292)
		w.state.tableNames = append(w.state.tableNames, tableName)
	}
	__antithesis_instrumentation__.Notify(695287)
	return rows.Err()
}

func unzip(src, dest string) error {
	__antithesis_instrumentation__.Notify(695295)
	r, err := zip.OpenReader(src)
	if err != nil {
		__antithesis_instrumentation__.Notify(695301)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695302)
	}
	__antithesis_instrumentation__.Notify(695296)
	defer func() {
		__antithesis_instrumentation__.Notify(695303)
		if err := r.Close(); err != nil {
			__antithesis_instrumentation__.Notify(695304)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(695305)
		}
	}()
	__antithesis_instrumentation__.Notify(695297)

	if err = os.MkdirAll(dest, 0755); err != nil {
		__antithesis_instrumentation__.Notify(695306)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695307)
	}
	__antithesis_instrumentation__.Notify(695298)

	extractAndWriteFile := func(f *zip.File) error {
		__antithesis_instrumentation__.Notify(695308)
		rc, err := f.Open()
		if err != nil {
			__antithesis_instrumentation__.Notify(695313)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695314)
		}
		__antithesis_instrumentation__.Notify(695309)
		defer func() {
			__antithesis_instrumentation__.Notify(695315)
			if err := rc.Close(); err != nil {
				__antithesis_instrumentation__.Notify(695316)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(695317)
			}
		}()
		__antithesis_instrumentation__.Notify(695310)

		path := filepath.Join(dest, f.Name)

		if !strings.HasPrefix(path, filepath.Clean(dest)+string(os.PathSeparator)) {
			__antithesis_instrumentation__.Notify(695318)
			return errors.Errorf("%s: illegal file path while extracting the zip. "+
				"Such a file path can be dangerous because of ZipSlip vulnerability. "+
				"Please reconsider whether the zip file is trustworthy.", path)
		} else {
			__antithesis_instrumentation__.Notify(695319)
		}
		__antithesis_instrumentation__.Notify(695311)

		if f.FileInfo().IsDir() {
			__antithesis_instrumentation__.Notify(695320)
			if err = os.MkdirAll(path, f.Mode()); err != nil {
				__antithesis_instrumentation__.Notify(695321)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695322)
			}
		} else {
			__antithesis_instrumentation__.Notify(695323)
			if err = os.MkdirAll(filepath.Dir(path), f.Mode()); err != nil {
				__antithesis_instrumentation__.Notify(695327)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695328)
			}
			__antithesis_instrumentation__.Notify(695324)
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				__antithesis_instrumentation__.Notify(695329)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695330)
			}
			__antithesis_instrumentation__.Notify(695325)
			defer func() {
				__antithesis_instrumentation__.Notify(695331)
				if err := f.Close(); err != nil {
					__antithesis_instrumentation__.Notify(695332)
					panic(err)
				} else {
					__antithesis_instrumentation__.Notify(695333)
				}
			}()
			__antithesis_instrumentation__.Notify(695326)

			_, err = io.Copy(f, rc)
			if err != nil {
				__antithesis_instrumentation__.Notify(695334)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695335)
			}
		}
		__antithesis_instrumentation__.Notify(695312)
		return nil
	}
	__antithesis_instrumentation__.Notify(695299)

	for _, f := range r.File {
		__antithesis_instrumentation__.Notify(695336)
		err := extractAndWriteFile(f)
		if err != nil {
			__antithesis_instrumentation__.Notify(695337)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695338)
		}
	}
	__antithesis_instrumentation__.Notify(695300)

	return nil
}

const (
	regexQueryLogFormat = `[^\s]+\s[^\s]+\s[^\s]+\s[^\s]+\s{2}[^\s]+\s[^\s]+\s[^\s]+\s".*"\s\{.*\}\s` +
		`"(.*)"` + `\s` + `(\{.*\})` + `\s[^\s]+\s[\d]+\s[^\s]+`

	regexQueryGroupIdx = 1

	queryLogHeaderLines = 5

	regexWriteQueryPattern = `(UPSERT|INSERT)\sINTO\s([^\s]+)\((.*)\)\sVALUES\s(\(.*\)\,\s)*(\(.*\))`
)

func (w *querylog) parseFile(ctx context.Context, fileInfo os.FileInfo, re *regexp.Regexp) error {
	__antithesis_instrumentation__.Notify(695339)
	start := timeutil.Now()
	file, err := os.Open(w.dirPath + "/" + fileInfo.Name())
	if err != nil {
		__antithesis_instrumentation__.Notify(695342)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695343)
	}
	__antithesis_instrumentation__.Notify(695340)

	reader := bufio.NewReader(file)
	buffer := make([]byte, 0)
	lineCount := 0
	for {
		__antithesis_instrumentation__.Notify(695344)
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			__antithesis_instrumentation__.Notify(695346)
			if err != io.EOF {
				__antithesis_instrumentation__.Notify(695348)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695349)
			}
			__antithesis_instrumentation__.Notify(695347)

			end := timeutil.Now()
			elapsed := end.Sub(start)
			log.Infof(ctx, "Processing of %s is done in %fs", fileInfo.Name(), elapsed.Seconds())
			break
		} else {
			__antithesis_instrumentation__.Notify(695350)
		}
		__antithesis_instrumentation__.Notify(695345)

		buffer = append(buffer, line...)
		if !isPrefix {
			__antithesis_instrumentation__.Notify(695351)

			if lineCount >= queryLogHeaderLines {
				__antithesis_instrumentation__.Notify(695353)

				if !re.Match(buffer) {
					__antithesis_instrumentation__.Notify(695355)
					return errors.Errorf("Line %d doesn't match the pattern", lineCount)
				} else {
					__antithesis_instrumentation__.Notify(695356)
				}
				__antithesis_instrumentation__.Notify(695354)

				groups := re.FindSubmatch(buffer)
				query := string(groups[regexQueryGroupIdx])
				isWriteQuery := isInsertOrUpsert(query)
				skipQuery := strings.HasPrefix(query, "ALTER") || func() bool {
					__antithesis_instrumentation__.Notify(695357)
					return (w.omitWriteQueries && func() bool {
						__antithesis_instrumentation__.Notify(695358)
						return isWriteQuery == true
					}() == true) == true
				}() == true || func() bool {
					__antithesis_instrumentation__.Notify(695359)
					return (w.omitDeleteQueries && func() bool {
						__antithesis_instrumentation__.Notify(695360)
						return strings.HasPrefix(query, "DELETE") == true
					}() == true) == true
				}() == true
				if !skipQuery {
					__antithesis_instrumentation__.Notify(695361)
					tableAssigned := false
					for i, tableName := range w.state.tableNames {
						__antithesis_instrumentation__.Notify(695363)

						if strings.Contains(query, tableName) {
							__antithesis_instrumentation__.Notify(695364)
							if !tableAssigned {
								__antithesis_instrumentation__.Notify(695366)

								w.state.seenQueriesByTableName[i][query]++
								w.state.queryCountPerTable[i]++
								w.state.totalQueryCount++
								tableAssigned = true
							} else {
								__antithesis_instrumentation__.Notify(695367)
							}
							__antithesis_instrumentation__.Notify(695365)
							w.state.tableUsed[tableName] = true
						} else {
							__antithesis_instrumentation__.Notify(695368)
						}
					}
					__antithesis_instrumentation__.Notify(695362)
					if !tableAssigned {
						__antithesis_instrumentation__.Notify(695369)
						return errors.Errorf("No table matched query %s while processing %s", query, fileInfo.Name())
					} else {
						__antithesis_instrumentation__.Notify(695370)
					}
				} else {
					__antithesis_instrumentation__.Notify(695371)
				}
			} else {
				__antithesis_instrumentation__.Notify(695372)
			}
			__antithesis_instrumentation__.Notify(695352)
			buffer = buffer[:0]
			lineCount++
		} else {
			__antithesis_instrumentation__.Notify(695373)
		}
	}
	__antithesis_instrumentation__.Notify(695341)
	return file.Close()
}

func (w *querylog) processQueryLog(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(695374)
	if w.zipPath != "" {
		__antithesis_instrumentation__.Notify(695381)
		log.Infof(ctx, "About to start unzipping %s", w.zipPath)
		w.dirPath = w.zipPath[:len(w.zipPath)-4]
		if err := unzip(w.zipPath, w.dirPath); err != nil {
			__antithesis_instrumentation__.Notify(695383)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695384)
		}
		__antithesis_instrumentation__.Notify(695382)
		log.Infof(ctx, "Unzipping to %s is complete", w.dirPath)
	} else {
		__antithesis_instrumentation__.Notify(695385)
	}
	__antithesis_instrumentation__.Notify(695375)

	files, err := ioutil.ReadDir(w.dirPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(695386)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695387)
	}
	__antithesis_instrumentation__.Notify(695376)

	w.state.queryCountPerTable = make([]int, len(w.state.tableNames))
	w.state.seenQueriesByTableName = make([]map[string]int, len(w.state.tableNames))
	for i := range w.state.tableNames {
		__antithesis_instrumentation__.Notify(695388)
		w.state.seenQueriesByTableName[i] = make(map[string]int)
	}
	__antithesis_instrumentation__.Notify(695377)
	w.state.tableUsed = make(map[string]bool)

	re := regexp.MustCompile(regexQueryLogFormat)
	log.Infof(ctx, "Starting to parse the query log")
	numFiles := w.filesToParse
	if numFiles > len(files) {
		__antithesis_instrumentation__.Notify(695389)
		numFiles = len(files)
	} else {
		__antithesis_instrumentation__.Notify(695390)
	}
	__antithesis_instrumentation__.Notify(695378)
	for fileNum, fileInfo := range files {
		__antithesis_instrumentation__.Notify(695391)
		if fileNum == w.filesToParse {
			__antithesis_instrumentation__.Notify(695394)

			break
		} else {
			__antithesis_instrumentation__.Notify(695395)
		}
		__antithesis_instrumentation__.Notify(695392)
		if fileInfo.IsDir() {
			__antithesis_instrumentation__.Notify(695396)
			if w.verbose {
				__antithesis_instrumentation__.Notify(695398)
				log.Infof(ctx, "Unexpected: a directory %s is encountered with the query log, skipping it.", fileInfo.Name())
			} else {
				__antithesis_instrumentation__.Notify(695399)
			}
			__antithesis_instrumentation__.Notify(695397)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695400)
		}
		__antithesis_instrumentation__.Notify(695393)
		log.Infof(ctx, "Processing %d out of %d", fileNum, numFiles)
		if err = w.parseFile(ctx, fileInfo, re); err != nil {
			__antithesis_instrumentation__.Notify(695401)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695402)
		}
	}
	__antithesis_instrumentation__.Notify(695379)
	log.Infof(ctx, "Query log processed")
	if w.zipPath != "" {
		__antithesis_instrumentation__.Notify(695403)
		log.Infof(ctx, "Unzipped files are about to be removed")
		return os.RemoveAll(w.dirPath)
	} else {
		__antithesis_instrumentation__.Notify(695404)
	}
	__antithesis_instrumentation__.Notify(695380)
	return nil
}

func (w *querylog) getColumnsInfo(db *gosql.DB) (retErr error) {
	__antithesis_instrumentation__.Notify(695405)
	w.state.columnsByTableName = make(map[string][]columnInfo)
	for _, tableName := range w.state.tableNames {
		__antithesis_instrumentation__.Notify(695407)
		if !w.state.tableUsed[tableName] {
			__antithesis_instrumentation__.Notify(695419)

			continue
		} else {
			__antithesis_instrumentation__.Notify(695420)
		}
		__antithesis_instrumentation__.Notify(695408)

		columnTypeByColumnName := make(map[string]string)
		rows, err := db.Query(fmt.Sprintf("SELECT column_name, data_type FROM [SHOW COLUMNS FROM %s]", tableName))
		if err != nil {
			__antithesis_instrumentation__.Notify(695421)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695422)
		}
		__antithesis_instrumentation__.Notify(695409)
		defer func(rows *gosql.Rows) {
			__antithesis_instrumentation__.Notify(695423)
			retErr = errors.CombineErrors(retErr, rows.Close())
		}(rows)
		__antithesis_instrumentation__.Notify(695410)
		for rows.Next() {
			__antithesis_instrumentation__.Notify(695424)
			var columnName, dataType string
			if err = rows.Scan(&columnName, &dataType); err != nil {
				__antithesis_instrumentation__.Notify(695426)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695427)
			}
			__antithesis_instrumentation__.Notify(695425)
			columnTypeByColumnName[columnName] = dataType
		}
		__antithesis_instrumentation__.Notify(695411)
		if err = rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(695428)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695429)
		}
		__antithesis_instrumentation__.Notify(695412)

		var relid int
		if err := db.QueryRow(fmt.Sprintf("SELECT '%s'::REGCLASS::OID", tableName)).Scan(&relid); err != nil {
			__antithesis_instrumentation__.Notify(695430)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695431)
		}
		__antithesis_instrumentation__.Notify(695413)
		rows, err = db.Query(
			`
SELECT attname, atttypid, adsrc, NOT attnotnull
FROM pg_catalog.pg_attribute
LEFT JOIN pg_catalog.pg_attrdef
ON attrelid=adrelid AND attnum=adnum
WHERE attrelid=$1`, relid)
		if err != nil {
			__antithesis_instrumentation__.Notify(695432)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695433)
		}
		__antithesis_instrumentation__.Notify(695414)
		defer func(rows *gosql.Rows) {
			__antithesis_instrumentation__.Notify(695434)
			retErr = errors.CombineErrors(retErr, rows.Close())
		}(rows)
		__antithesis_instrumentation__.Notify(695415)

		var cols []columnInfo
		var numCols = 0

		for rows.Next() {
			__antithesis_instrumentation__.Notify(695435)
			var c columnInfo
			c.dataPrecision = 0
			c.dataScale = 0

			var typOid int
			if err := rows.Scan(&c.name, &typOid, &c.cdefault, &c.isNullable); err != nil {
				__antithesis_instrumentation__.Notify(695438)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695439)
			}
			__antithesis_instrumentation__.Notify(695436)
			c.dataType = types.OidToType[oid.Oid(typOid)]
			if c.dataType.Family() == types.IntFamily {
				__antithesis_instrumentation__.Notify(695440)
				actualType := columnTypeByColumnName[c.name]
				if actualType == `INT2` {
					__antithesis_instrumentation__.Notify(695441)
					c.intRange = 1 << 16
				} else {
					__antithesis_instrumentation__.Notify(695442)
					if actualType == `INT4` {
						__antithesis_instrumentation__.Notify(695443)
						c.intRange = 1 << 32
					} else {
						__antithesis_instrumentation__.Notify(695444)
					}
				}
			} else {
				__antithesis_instrumentation__.Notify(695445)
			}
			__antithesis_instrumentation__.Notify(695437)
			cols = append(cols, c)
			numCols++
		}
		__antithesis_instrumentation__.Notify(695416)
		if err = rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(695446)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695447)
		}
		__antithesis_instrumentation__.Notify(695417)
		if numCols == 0 {
			__antithesis_instrumentation__.Notify(695448)
			return errors.Errorf("no columns detected")
		} else {
			__antithesis_instrumentation__.Notify(695449)
		}
		__antithesis_instrumentation__.Notify(695418)
		w.state.columnsByTableName[tableName] = cols
	}
	__antithesis_instrumentation__.Notify(695406)
	return nil
}

func (w *querylog) populateSamples(ctx context.Context, db *gosql.DB) (retErr error) {
	__antithesis_instrumentation__.Notify(695450)
	log.Infof(ctx, "Populating samples started")
	for _, tableName := range w.state.tableNames {
		__antithesis_instrumentation__.Notify(695452)
		cols := w.state.columnsByTableName[tableName]
		if cols == nil {
			__antithesis_instrumentation__.Notify(695460)

			continue
		} else {
			__antithesis_instrumentation__.Notify(695461)
		}
		__antithesis_instrumentation__.Notify(695453)
		log.Infof(ctx, "Sampling %s", tableName)
		row := db.QueryRow(fmt.Sprintf(`SELECT count(*) FROM %s`, tableName))
		var numRows int
		if err := row.Scan(&numRows); err != nil {
			__antithesis_instrumentation__.Notify(695462)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695463)
		}
		__antithesis_instrumentation__.Notify(695454)

		columnNames := make([]string, len(cols))
		for i := range cols {
			__antithesis_instrumentation__.Notify(695464)
			columnNames[i] = cols[i].name
		}
		__antithesis_instrumentation__.Notify(695455)
		columnsOrdered := strings.Join(columnNames, ", ")

		var samplesQuery string
		if w.numSamples > numRows {
			__antithesis_instrumentation__.Notify(695465)
			samplesQuery = fmt.Sprintf(`SELECT %s FROM %s`, columnsOrdered, tableName)
		} else {
			__antithesis_instrumentation__.Notify(695466)
			samplingProb := float64(w.numSamples) / float64(numRows)
			if samplingProb < w.minSamplingProb {
				__antithesis_instrumentation__.Notify(695468)

				samplingProb = w.minSamplingProb
			} else {
				__antithesis_instrumentation__.Notify(695469)
			}
			__antithesis_instrumentation__.Notify(695467)
			samplesQuery = fmt.Sprintf(`SELECT %s FROM %s WHERE random() < %f LIMIT %d`,
				columnsOrdered, tableName, samplingProb, w.numSamples)
		}
		__antithesis_instrumentation__.Notify(695456)

		samples, err := db.Query(samplesQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(695470)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695471)
		}
		__antithesis_instrumentation__.Notify(695457)
		defer func() {
			__antithesis_instrumentation__.Notify(695472)
			retErr = errors.CombineErrors(retErr, samples.Close())
		}()
		__antithesis_instrumentation__.Notify(695458)
		for samples.Next() {
			__antithesis_instrumentation__.Notify(695473)
			rowOfSamples := make([]interface{}, len(cols))
			for i := range rowOfSamples {
				__antithesis_instrumentation__.Notify(695476)
				rowOfSamples[i] = new(interface{})
			}
			__antithesis_instrumentation__.Notify(695474)
			if err := samples.Scan(rowOfSamples...); err != nil {
				__antithesis_instrumentation__.Notify(695477)
				return err
			} else {
				__antithesis_instrumentation__.Notify(695478)
			}
			__antithesis_instrumentation__.Notify(695475)
			for i, sample := range rowOfSamples {
				__antithesis_instrumentation__.Notify(695479)
				cols[i].samples = append(cols[i].samples, sample)
			}
		}
		__antithesis_instrumentation__.Notify(695459)
		if err = samples.Err(); err != nil {
			__antithesis_instrumentation__.Notify(695480)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695481)
		}
	}
	__antithesis_instrumentation__.Notify(695451)
	log.Infof(ctx, "Populating samples is complete")
	return nil
}

func (w *worker) querybenchRun(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(695482)
	file, err := os.Create(w.querybenchPath)
	if err != nil {
		__antithesis_instrumentation__.Notify(695484)
		return err
	} else {
		__antithesis_instrumentation__.Notify(695485)
	}
	__antithesis_instrumentation__.Notify(695483)
	defer file.Close()

	reToAvoid := regexp.MustCompile(`\$[0-9]+`)
	writer := bufio.NewWriter(file)
	queryCount := uint(0)
	for {
		__antithesis_instrumentation__.Notify(695486)
		chosenQuery, tableName := w.chooseQuery()
		pholdersColumnNames, numRepeats, err := w.deduceColumnNamesForPlaceholders(ctx, chosenQuery)
		if err != nil {
			__antithesis_instrumentation__.Notify(695493)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695495)
				log.Infof(ctx, "Encountered an error %s while deducing column names corresponding to the placeholders", err.Error())
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695496)
			}
			__antithesis_instrumentation__.Notify(695494)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695497)
		}
		__antithesis_instrumentation__.Notify(695487)

		placeholders, err := w.generatePlaceholders(ctx, chosenQuery, pholdersColumnNames, numRepeats, tableName)
		if err != nil {
			__antithesis_instrumentation__.Notify(695498)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695500)
				log.Infof(ctx, "Encountered an error %s while generating values for the placeholders", err.Error())
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695501)
			}
			__antithesis_instrumentation__.Notify(695499)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695502)
		}
		__antithesis_instrumentation__.Notify(695488)

		query := chosenQuery
		skipQuery := false

		for i := len(placeholders) - 1; i >= 0; i-- {
			__antithesis_instrumentation__.Notify(695503)
			pholderString := fmt.Sprintf("$%d", i+1)
			if !strings.Contains(chosenQuery, pholderString) {
				__antithesis_instrumentation__.Notify(695506)
				skipQuery = true
				break
			} else {
				__antithesis_instrumentation__.Notify(695507)
			}
			__antithesis_instrumentation__.Notify(695504)
			pholderValue := printPlaceholder(placeholders[i])
			if reToAvoid.MatchString(pholderValue) {
				__antithesis_instrumentation__.Notify(695508)
				skipQuery = true
				break
			} else {
				__antithesis_instrumentation__.Notify(695509)
			}
			__antithesis_instrumentation__.Notify(695505)
			query = strings.Replace(query, pholderString, pholderValue, -1)
		}
		__antithesis_instrumentation__.Notify(695489)
		if skipQuery {
			__antithesis_instrumentation__.Notify(695510)
			if w.config.verbose {
				__antithesis_instrumentation__.Notify(695512)
				log.Infof(ctx, "Could not replace placeholders with values on query")
				printQueryShortened(ctx, chosenQuery)
			} else {
				__antithesis_instrumentation__.Notify(695513)
			}
			__antithesis_instrumentation__.Notify(695511)
			continue
		} else {
			__antithesis_instrumentation__.Notify(695514)
		}
		__antithesis_instrumentation__.Notify(695490)
		if _, err = writer.WriteString(query + "\n\n"); err != nil {
			__antithesis_instrumentation__.Notify(695515)
			return err
		} else {
			__antithesis_instrumentation__.Notify(695516)
		}
		__antithesis_instrumentation__.Notify(695491)

		queryCount++
		if queryCount%250 == 0 {
			__antithesis_instrumentation__.Notify(695517)
			log.Infof(ctx, "%d queries have been written", queryCount)
		} else {
			__antithesis_instrumentation__.Notify(695518)
		}
		__antithesis_instrumentation__.Notify(695492)
		if queryCount == w.config.count {
			__antithesis_instrumentation__.Notify(695519)
			writer.Flush()
			return nil
		} else {
			__antithesis_instrumentation__.Notify(695520)
		}
	}
}

func printPlaceholder(i interface{}) string {
	__antithesis_instrumentation__.Notify(695521)
	if ptr, ok := i.(*interface{}); ok {
		__antithesis_instrumentation__.Notify(695523)
		return printPlaceholder(*ptr)
	} else {
		__antithesis_instrumentation__.Notify(695524)
	}
	__antithesis_instrumentation__.Notify(695522)
	switch p := i.(type) {
	case bool:
		__antithesis_instrumentation__.Notify(695525)
		return fmt.Sprintf("%v", p)
	case int64:
		__antithesis_instrumentation__.Notify(695526)
		return fmt.Sprintf("%d", p)
	case float64:
		__antithesis_instrumentation__.Notify(695527)
		return fmt.Sprintf("%f", p)
	case []uint8:
		__antithesis_instrumentation__.Notify(695528)
		u, err := uuid.FromString(string(p))
		if err != nil {
			__antithesis_instrumentation__.Notify(695535)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(695536)
		}
		__antithesis_instrumentation__.Notify(695529)
		return fmt.Sprintf("'%s'", u.String())
	case uuid.UUID:
		__antithesis_instrumentation__.Notify(695530)
		return fmt.Sprintf("'%s'", p.String())
	case string:
		__antithesis_instrumentation__.Notify(695531)
		s := strings.Replace(p, "'", "''", -1)

		s = strings.Replace(s, "\n", "", -1)
		s = strings.Replace(s, "\r", "", -1)
		return fmt.Sprintf("'%s'", s)
	case time.Time:
		__antithesis_instrumentation__.Notify(695532)
		timestamp := p.String()

		idx := strings.Index(timestamp, `+0000`)
		timestamp = timestamp[:idx+5]
		return fmt.Sprintf("'%s':::TIMESTAMP", timestamp)
	case nil:
		__antithesis_instrumentation__.Notify(695533)
		return "NULL"
	default:
		__antithesis_instrumentation__.Notify(695534)
		panic(errors.AssertionFailedf("unsupported type: %T", i))
	}
}

type columnInfo struct {
	name          string
	dataType      *types.T
	dataPrecision int
	dataScale     int
	cdefault      gosql.NullString
	isNullable    bool

	intRange uint64
	samples  []interface{}
}

func printQueryShortened(ctx context.Context, query string) {
	__antithesis_instrumentation__.Notify(695537)
	if len(query) > 1000 {
		__antithesis_instrumentation__.Notify(695538)
		log.Infof(ctx, "%s...%s", query[:500], query[len(query)-500:])
	} else {
		__antithesis_instrumentation__.Notify(695539)
		log.Infof(ctx, "%s", query)
	}
}

func isInsertOrUpsert(query string) bool {
	__antithesis_instrumentation__.Notify(695540)
	return strings.HasPrefix(query, "INSERT") || func() bool {
		__antithesis_instrumentation__.Notify(695541)
		return strings.HasPrefix(query, "UPSERT") == true
	}() == true
}

func probInInterval(start, inc, total int, p float64) bool {
	__antithesis_instrumentation__.Notify(695542)
	return float64(start)/float64(total) <= p && func() bool {
		__antithesis_instrumentation__.Notify(695543)
		return float64(start+inc)/float64(total) > p == true
	}() == true
}
