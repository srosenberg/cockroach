package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlclient"
	"github.com/cockroachdb/cockroach/pkg/cli/democluster"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var debugStatementBundleCmd = &cobra.Command{
	Use:     "statement-bundle [command]",
	Aliases: []string{"sb"},
	Short:   "run a cockroach debug statement-bundle tool command",
	Long: `
debug statement-bundle is a suite of tools for debugging and manipulating statement
bundles created with EXPLAIN ANALYZE (DEBUG).
`,
}

var statementBundleRecreateCmd = &cobra.Command{
	Use:   "recreate <stmt bundle zipdir>",
	Short: "recreate the statement bundle in a demo cluster",
	Long: `
Run the recreate tool to populate a demo cluster with the environment, schema,
and stats in an unzipped statement bundle directory.
`,
	Args: cobra.ExactArgs(1),
}

var placeholderPairs []string
var explainPrefix string

func init() {
	statementBundleRecreateCmd.RunE = clierrorplus.MaybeDecorateError(func(cmd *cobra.Command, args []string) error {
		return runBundleRecreate(cmd, args)
	})

	statementBundleRecreateCmd.Flags().StringArrayVar(&placeholderPairs, "placeholder", nil,
		"pass in a map of placeholder id to fully-qualified table column to get the program to produce all optimal"+
			" of explain plans with each of the histogram values for each column replaced in its placeholder.")
	statementBundleRecreateCmd.Flags().StringVar(&explainPrefix, "explain-cmd", "EXPLAIN",
		"set the EXPLAIN command used to produce the final output when displaying all optimal explain plans with"+
			" --placeholder. Example: EXPLAIN(OPT)")
}

type statementBundle struct {
	env       []byte
	schema    []byte
	statement []byte
	stats     [][]byte
}

func loadStatementBundle(zipdir string) (*statementBundle, error) {
	__antithesis_instrumentation__.Notify(34365)
	ret := &statementBundle{}
	var err error
	ret.env, err = ioutil.ReadFile(filepath.Join(zipdir, "env.sql"))
	if err != nil {
		__antithesis_instrumentation__.Notify(34369)
		return ret, err
	} else {
		__antithesis_instrumentation__.Notify(34370)
	}
	__antithesis_instrumentation__.Notify(34366)
	ret.schema, err = ioutil.ReadFile(filepath.Join(zipdir, "schema.sql"))
	if err != nil {
		__antithesis_instrumentation__.Notify(34371)
		return ret, err
	} else {
		__antithesis_instrumentation__.Notify(34372)
	}
	__antithesis_instrumentation__.Notify(34367)
	ret.statement, err = ioutil.ReadFile(filepath.Join(zipdir, "statement.txt"))
	if err != nil {
		__antithesis_instrumentation__.Notify(34373)
		return ret, err
	} else {
		__antithesis_instrumentation__.Notify(34374)
	}
	__antithesis_instrumentation__.Notify(34368)

	return ret, filepath.WalkDir(zipdir, func(path string, d fs.DirEntry, _ error) error {
		__antithesis_instrumentation__.Notify(34375)
		if d.IsDir() {
			__antithesis_instrumentation__.Notify(34379)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(34380)
		}
		__antithesis_instrumentation__.Notify(34376)
		if !strings.HasPrefix(d.Name(), "stats-") {
			__antithesis_instrumentation__.Notify(34381)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(34382)
		}
		__antithesis_instrumentation__.Notify(34377)
		f, err := ioutil.ReadFile(path)
		if err != nil {
			__antithesis_instrumentation__.Notify(34383)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34384)
		}
		__antithesis_instrumentation__.Notify(34378)
		ret.stats = append(ret.stats, f)
		return nil
	})
}

func runBundleRecreate(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(34385)
	zipdir := args[0]
	bundle, err := loadStatementBundle(zipdir)
	if err != nil {
		__antithesis_instrumentation__.Notify(34395)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34396)
	}
	__antithesis_instrumentation__.Notify(34386)

	closeFn, err := sqlCtx.Open(os.Stdin)
	if err != nil {
		__antithesis_instrumentation__.Notify(34397)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34398)
	}
	__antithesis_instrumentation__.Notify(34387)
	defer closeFn()
	ctx := context.Background()
	c, err := democluster.NewDemoCluster(ctx, &demoCtx,
		log.Infof,
		log.Warningf,
		log.Ops.Shoutf,
		func(ctx context.Context) (*stop.Stopper, error) {
			__antithesis_instrumentation__.Notify(34399)

			serverCfg.Stores.Specs = nil
			return setupAndInitializeLoggingAndProfiling(ctx, cmd, false)
		},
		getAdminClient,
		func(ctx context.Context, ac serverpb.AdminClient) error {
			__antithesis_instrumentation__.Notify(34400)
			return drainAndShutdown(ctx, ac, "local")
		},
	)
	__antithesis_instrumentation__.Notify(34388)
	if err != nil {
		__antithesis_instrumentation__.Notify(34401)
		c.Close(ctx)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34402)
	}
	__antithesis_instrumentation__.Notify(34389)
	defer c.Close(ctx)

	initGEOS(ctx)

	if err := c.Start(ctx, runInitialSQL); err != nil {
		__antithesis_instrumentation__.Notify(34403)
		return clierrorplus.CheckAndMaybeShout(err)
	} else {
		__antithesis_instrumentation__.Notify(34404)
	}
	__antithesis_instrumentation__.Notify(34390)
	conn, err := sqlCtx.MakeConn(c.GetConnURL())
	if err != nil {
		__antithesis_instrumentation__.Notify(34405)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34406)
	}
	__antithesis_instrumentation__.Notify(34391)

	if err := conn.Exec(ctx,
		`SET CLUSTER SETTING sql.stats.automatic_collection.enabled = false`); err != nil {
		__antithesis_instrumentation__.Notify(34407)
		return err
	} else {
		__antithesis_instrumentation__.Notify(34408)
	}
	__antithesis_instrumentation__.Notify(34392)
	var initStmts = [][]byte{bundle.env, bundle.schema}
	initStmts = append(initStmts, bundle.stats...)
	for _, a := range initStmts {
		__antithesis_instrumentation__.Notify(34409)
		if err := conn.Exec(ctx, string(a)); err != nil {
			__antithesis_instrumentation__.Notify(34410)
			return errors.Wrapf(err, "failed to run %s", a)
		} else {
			__antithesis_instrumentation__.Notify(34411)
		}
	}
	__antithesis_instrumentation__.Notify(34393)

	cliCtx.PrintfUnlessEmbedded(`#
# Statement bundle %s loaded.
# Autostats disabled.
#
# Statement was:
#
# %s
`, zipdir, bundle.statement)

	if placeholderPairs != nil {
		__antithesis_instrumentation__.Notify(34412)
		placeholderToColMap := make(map[int]string)
		for _, placeholderPairStr := range placeholderPairs {
			__antithesis_instrumentation__.Notify(34415)
			pair := strings.Split(placeholderPairStr, "=")
			if len(pair) != 2 {
				__antithesis_instrumentation__.Notify(34418)
				return errors.New("use --placeholder='1=schema.table.col' --placeholder='2=schema.table.col...'")
			} else {
				__antithesis_instrumentation__.Notify(34419)
			}
			__antithesis_instrumentation__.Notify(34416)
			n, err := strconv.Atoi(pair[0])
			if err != nil {
				__antithesis_instrumentation__.Notify(34420)
				return err
			} else {
				__antithesis_instrumentation__.Notify(34421)
			}
			__antithesis_instrumentation__.Notify(34417)
			placeholderToColMap[n] = pair[1]
		}
		__antithesis_instrumentation__.Notify(34413)
		inputs, outputs, err := getExplainCombinations(conn, explainPrefix, placeholderToColMap, bundle)
		if err != nil {
			__antithesis_instrumentation__.Notify(34422)
			return err
		} else {
			__antithesis_instrumentation__.Notify(34423)
		}
		__antithesis_instrumentation__.Notify(34414)

		cliCtx.PrintfUnlessEmbedded("found %d unique explains:\n\n", len(inputs))
		for i, inputs := range inputs {
			__antithesis_instrumentation__.Notify(34424)
			cliCtx.PrintfUnlessEmbedded("Values %s: \n%s\n----\n\n", inputs, outputs[i])
		}
	} else {
		__antithesis_instrumentation__.Notify(34425)
	}
	__antithesis_instrumentation__.Notify(34394)

	sqlCtx.ShellCtx.DemoCluster = c
	return sqlCtx.Run(conn)
}

var placeholderRe = regexp.MustCompile(`\$(\d+): .*`)

var statsRe = regexp.MustCompile(`ALTER TABLE ([\w.]+) INJECT STATISTICS '`)

type bucketKey struct {
	NumEq         float64
	NumRange      float64
	DistinctRange float64
}

func getExplainCombinations(
	conn clisqlclient.Conn,
	explainPrefix string,
	placeholderToColMap map[int]string,
	bundle *statementBundle,
) (inputs [][]string, explainOutputs []string, err error) {
	__antithesis_instrumentation__.Notify(34426)

	stmtComponents := strings.Split(string(bundle.statement), "Arguments:")
	statement := strings.TrimSpace(stmtComponents[0])
	placeholders := stmtComponents[1]

	var stmtPlaceholders []int
	for _, line := range strings.Split(placeholders, "\n") {
		__antithesis_instrumentation__.Notify(34436)

		if matches := placeholderRe.FindStringSubmatch(line); len(matches) == 2 {
			__antithesis_instrumentation__.Notify(34437)

			n, err := strconv.Atoi(matches[1])
			if err != nil {
				__antithesis_instrumentation__.Notify(34439)
				return nil, nil, err
			} else {
				__antithesis_instrumentation__.Notify(34440)
			}
			__antithesis_instrumentation__.Notify(34438)
			stmtPlaceholders = append(stmtPlaceholders, n)
		} else {
			__antithesis_instrumentation__.Notify(34441)
		}
	}
	__antithesis_instrumentation__.Notify(34427)

	for _, n := range stmtPlaceholders {
		__antithesis_instrumentation__.Notify(34442)
		if placeholderToColMap[n] == "" {
			__antithesis_instrumentation__.Notify(34443)
			return nil, nil, errors.Errorf("specify --placeholder= for placeholder %d", n)
		} else {
			__antithesis_instrumentation__.Notify(34444)
		}
	}
	__antithesis_instrumentation__.Notify(34428)
	evalCtx := tree.MakeTestingEvalContext(cluster.MakeTestingClusterSettings())

	fmtCtx := tree.FmtBareStrings

	statsMap := make(map[string][]string)
	statsAge := make(map[string]time.Time)
	for _, statsBytes := range bundle.stats {
		__antithesis_instrumentation__.Notify(34445)
		statsStr := string(statsBytes)
		matches := statsRe.FindStringSubmatch(statsStr)
		if len(matches) != 2 {
			__antithesis_instrumentation__.Notify(34448)
			return nil, nil, errors.Errorf("invalid stats file %s", statsStr)
		} else {
			__antithesis_instrumentation__.Notify(34449)
		}
		__antithesis_instrumentation__.Notify(34446)
		tableName := matches[1]

		idx := bytes.IndexByte(statsBytes, '\'')
		var statsJSON []map[string]interface{}

		data := statsBytes[idx+1 : len(statsBytes)-3]
		if err := json.Unmarshal(data, &statsJSON); err != nil {
			__antithesis_instrumentation__.Notify(34450)
			return nil, nil, err
		} else {
			__antithesis_instrumentation__.Notify(34451)
		}
		__antithesis_instrumentation__.Notify(34447)

		for _, stat := range statsJSON {
			__antithesis_instrumentation__.Notify(34452)
			bucketMap := make(map[bucketKey][]string)
			columns := stat["columns"].([]interface{})
			if len(columns) > 1 {
				__antithesis_instrumentation__.Notify(34461)

				continue
			} else {
				__antithesis_instrumentation__.Notify(34462)
			}
			__antithesis_instrumentation__.Notify(34453)
			col := columns[0]
			fqColName := fmt.Sprintf("%s.%s", tableName, col)
			d, _, err := tree.ParseDTimestamp(nil, stat["created_at"].(string), time.Microsecond)
			if err != nil {
				__antithesis_instrumentation__.Notify(34463)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(34464)
			}
			__antithesis_instrumentation__.Notify(34454)
			if lastStat, ok := statsAge[fqColName]; ok && func() bool {
				__antithesis_instrumentation__.Notify(34465)
				return d.Before(lastStat) == true
			}() == true {
				__antithesis_instrumentation__.Notify(34466)

				continue
			} else {
				__antithesis_instrumentation__.Notify(34467)
			}
			__antithesis_instrumentation__.Notify(34455)
			statsAge[fqColName] = d.Time

			typ := stat["histo_col_type"].(string)
			if typ == "" {
				__antithesis_instrumentation__.Notify(34468)
				fmt.Println("Ignoring column with empty type ", col)
				continue
			} else {
				__antithesis_instrumentation__.Notify(34469)
			}
			__antithesis_instrumentation__.Notify(34456)
			colTypeRef, err := parser.GetTypeFromValidSQLSyntax(typ)
			if err != nil {
				__antithesis_instrumentation__.Notify(34470)
				return nil, nil, errors.Wrapf(err, "unable to parse type %s for col %s", typ, col)
			} else {
				__antithesis_instrumentation__.Notify(34471)
			}
			__antithesis_instrumentation__.Notify(34457)
			colType := tree.MustBeStaticallyKnownType(colTypeRef)
			buckets := stat["histo_buckets"].([]interface{})
			var maxUpperBound tree.Datum
			for _, b := range buckets {
				__antithesis_instrumentation__.Notify(34472)
				bucket := b.(map[string]interface{})
				numRange := bucket["num_range"].(float64)
				key := bucketKey{
					NumEq:         bucket["num_eq"].(float64),
					NumRange:      numRange,
					DistinctRange: bucket["distinct_range"].(float64),
				}
				upperBound := bucket["upper_bound"].(string)
				bucketMap[key] = []string{upperBound}
				datum, err := rowenc.ParseDatumStringAs(colType, upperBound, &evalCtx)
				if err != nil {
					__antithesis_instrumentation__.Notify(34475)
					panic("failed parsing datum string as " + datum.String() + " " + err.Error())
				} else {
					__antithesis_instrumentation__.Notify(34476)
				}
				__antithesis_instrumentation__.Notify(34473)
				if maxUpperBound == nil || func() bool {
					__antithesis_instrumentation__.Notify(34477)
					return maxUpperBound.Compare(&evalCtx, datum) < 0 == true
				}() == true {
					__antithesis_instrumentation__.Notify(34478)
					maxUpperBound = datum
				} else {
					__antithesis_instrumentation__.Notify(34479)
				}
				__antithesis_instrumentation__.Notify(34474)
				if numRange > 0 {
					__antithesis_instrumentation__.Notify(34480)
					if prev, ok := datum.Prev(&evalCtx); ok {
						__antithesis_instrumentation__.Notify(34481)
						bucketMap[key] = append(bucketMap[key], tree.AsStringWithFlags(prev, fmtCtx))
					} else {
						__antithesis_instrumentation__.Notify(34482)
					}
				} else {
					__antithesis_instrumentation__.Notify(34483)
				}
			}
			__antithesis_instrumentation__.Notify(34458)
			colSamples := make([]string, 0, len(bucketMap))
			for _, samples := range bucketMap {
				__antithesis_instrumentation__.Notify(34484)
				colSamples = append(colSamples, samples...)
			}
			__antithesis_instrumentation__.Notify(34459)

			if outside, ok := maxUpperBound.Next(&evalCtx); ok {
				__antithesis_instrumentation__.Notify(34485)
				colSamples = append(colSamples, tree.AsStringWithFlags(outside, fmtCtx))
			} else {
				__antithesis_instrumentation__.Notify(34486)
			}
			__antithesis_instrumentation__.Notify(34460)
			sort.Strings(colSamples)
			statsMap[fqColName] = colSamples
		}
	}
	__antithesis_instrumentation__.Notify(34429)

	for _, fqColName := range placeholderToColMap {
		__antithesis_instrumentation__.Notify(34487)
		if statsMap[fqColName] == nil {
			__antithesis_instrumentation__.Notify(34488)
			return nil, nil, errors.Errorf("no stats found for %s", fqColName)
		} else {
			__antithesis_instrumentation__.Notify(34489)
		}
	}
	__antithesis_instrumentation__.Notify(34430)

	combinations := getPlaceholderCombinations(stmtPlaceholders, placeholderToColMap, statsMap)

	outputs, err := getExplainOutputs(conn, "EXPLAIN(SHAPE)", statement, combinations)
	if err != nil {
		__antithesis_instrumentation__.Notify(34490)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(34491)
	}
	__antithesis_instrumentation__.Notify(34431)

	uniqueExplains := make(map[string][]string)
	for i := range combinations {
		__antithesis_instrumentation__.Notify(34492)
		uniqueExplains[outputs[i]] = combinations[i]
	}
	__antithesis_instrumentation__.Notify(34432)

	explains := make([]string, 0, len(uniqueExplains))
	for key := range uniqueExplains {
		__antithesis_instrumentation__.Notify(34493)
		explains = append(explains, key)
	}
	__antithesis_instrumentation__.Notify(34433)
	sort.Strings(explains)

	uniqueInputs := make([][]string, 0, len(uniqueExplains))
	for _, explain := range explains {
		__antithesis_instrumentation__.Notify(34494)
		input := uniqueExplains[explain]
		uniqueInputs = append(uniqueInputs, input)
	}
	__antithesis_instrumentation__.Notify(34434)
	outputs, err = getExplainOutputs(conn, explainPrefix, statement, uniqueInputs)
	if err != nil {
		__antithesis_instrumentation__.Notify(34495)
		return nil, nil, err
	} else {
		__antithesis_instrumentation__.Notify(34496)
	}
	__antithesis_instrumentation__.Notify(34435)

	return uniqueInputs, outputs, nil
}

func getExplainOutputs(
	conn clisqlclient.Conn, explainPrefix string, statement string, inputs [][]string,
) (explainStrings []string, err error) {
	__antithesis_instrumentation__.Notify(34497)
	for _, values := range inputs {
		__antithesis_instrumentation__.Notify(34499)

		query := fmt.Sprintf("%s %s", explainPrefix, statement)
		args := make([]interface{}, len(values))
		for i, s := range values {
			__antithesis_instrumentation__.Notify(34505)
			args[i] = s
		}
		__antithesis_instrumentation__.Notify(34500)
		rows, err := conn.Query(context.Background(), query, args...)
		if err != nil {
			__antithesis_instrumentation__.Notify(34506)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(34507)
		}
		__antithesis_instrumentation__.Notify(34501)
		row := []driver.Value{""}
		var explainStr = strings.Builder{}
		for err = rows.Next(row); err == nil; err = rows.Next(row) {
			__antithesis_instrumentation__.Notify(34508)
			fmt.Fprintln(&explainStr, row[0])
		}
		__antithesis_instrumentation__.Notify(34502)
		if err != io.EOF {
			__antithesis_instrumentation__.Notify(34509)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(34510)
		}
		__antithesis_instrumentation__.Notify(34503)
		if err := rows.Close(); err != nil {
			__antithesis_instrumentation__.Notify(34511)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(34512)
		}
		__antithesis_instrumentation__.Notify(34504)
		explainStrings = append(explainStrings, explainStr.String())
	}
	__antithesis_instrumentation__.Notify(34498)
	return explainStrings, nil
}

func getPlaceholderCombinations(
	remainingPlaceholders []int, placeholderMap map[int]string, statsMap map[string][]string,
) [][]string {
	__antithesis_instrumentation__.Notify(34513)
	placeholder := remainingPlaceholders[0]
	fqColName := placeholderMap[placeholder]
	var rest = [][]string{nil}
	if len(remainingPlaceholders) > 1 {
		__antithesis_instrumentation__.Notify(34516)

		rest = getPlaceholderCombinations(remainingPlaceholders[1:], placeholderMap, statsMap)
	} else {
		__antithesis_instrumentation__.Notify(34517)
	}
	__antithesis_instrumentation__.Notify(34514)
	var ret [][]string
	for _, val := range statsMap[fqColName] {
		__antithesis_instrumentation__.Notify(34518)
		for _, inner := range rest {
			__antithesis_instrumentation__.Notify(34519)
			ret = append(ret, append([]string{val}, inner...))
		}
	}
	__antithesis_instrumentation__.Notify(34515)
	return ret
}
