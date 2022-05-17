package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/cat"
	"github.com/cockroachdb/cockroach/pkg/sql/opt/memo"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/sql/stmtdiagnostics"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/memzipper"
	"github.com/cockroachdb/cockroach/pkg/util/tracing"
	"github.com/cockroachdb/errors"
)

func setExplainBundleResult(
	ctx context.Context,
	res RestrictedCommandResult,
	bundle diagnosticsBundle,
	execCfg *ExecutorConfig,
) error {
	__antithesis_instrumentation__.Notify(491013)
	res.ResetStmtType(&tree.ExplainAnalyze{})
	res.SetColumns(ctx, colinfo.ExplainPlanColumns)

	var text []string
	if bundle.collectionErr != nil {
		__antithesis_instrumentation__.Notify(491017)

		text = []string{fmt.Sprintf("Error generating bundle: %v", bundle.collectionErr)}
	} else {
		__antithesis_instrumentation__.Notify(491018)
		if execCfg.Codec.ForSystemTenant() {
			__antithesis_instrumentation__.Notify(491019)
			text = []string{
				"Statement diagnostics bundle generated. Download from the Admin UI (Advanced",
				"Debug -> Statement Diagnostics History), via the direct link below, or using",
				"the SQL shell or command line.",
				fmt.Sprintf("Admin UI: %s", execCfg.AdminURL()),
				fmt.Sprintf("Direct link: %s/_admin/v1/stmtbundle/%d", execCfg.AdminURL(), bundle.diagID),
				fmt.Sprintf("SQL shell: \\statement-diag download %d", bundle.diagID),
				fmt.Sprintf("Command line: cockroach statement-diag download %d", bundle.diagID),
			}
		} else {
			__antithesis_instrumentation__.Notify(491020)

			text = []string{
				"Statement diagnostics bundle generated. Download using the SQL shell or command",
				"line.",
				fmt.Sprintf("SQL shell: \\statement-diag download %d", bundle.diagID),
				fmt.Sprintf("Command line: cockroach statement-diag download %d", bundle.diagID),
			}
		}
	}
	__antithesis_instrumentation__.Notify(491014)

	if err := res.Err(); err != nil {
		__antithesis_instrumentation__.Notify(491021)

		res.SetError(errors.WithDetail(err, strings.Join(text, "\n")))
		return nil
	} else {
		__antithesis_instrumentation__.Notify(491022)
	}
	__antithesis_instrumentation__.Notify(491015)

	for _, line := range text {
		__antithesis_instrumentation__.Notify(491023)
		if err := res.AddRow(ctx, tree.Datums{tree.NewDString(line)}); err != nil {
			__antithesis_instrumentation__.Notify(491024)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491025)
		}
	}
	__antithesis_instrumentation__.Notify(491016)
	return nil
}

type diagnosticsBundle struct {
	zip []byte

	collectionErr error

	diagID stmtdiagnostics.CollectedInstanceID
}

func buildStatementBundle(
	ctx context.Context,
	db *kv.DB,
	ie *InternalExecutor,
	plan *planTop,
	planString string,
	trace tracing.Recording,
	placeholders *tree.PlaceholderInfo,
) diagnosticsBundle {
	__antithesis_instrumentation__.Notify(491026)
	if plan == nil {
		__antithesis_instrumentation__.Notify(491029)
		return diagnosticsBundle{collectionErr: errors.AssertionFailedf("execution terminated early")}
	} else {
		__antithesis_instrumentation__.Notify(491030)
	}
	__antithesis_instrumentation__.Notify(491027)
	b := makeStmtBundleBuilder(db, ie, plan, trace, placeholders)

	b.addStatement()
	b.addOptPlans()
	b.addExecPlan(planString)
	b.addDistSQLDiagrams()
	b.addExplainVec()
	b.addTrace()
	b.addEnv(ctx)

	buf, err := b.finalize()
	if err != nil {
		__antithesis_instrumentation__.Notify(491031)
		return diagnosticsBundle{collectionErr: err}
	} else {
		__antithesis_instrumentation__.Notify(491032)
	}
	__antithesis_instrumentation__.Notify(491028)
	return diagnosticsBundle{zip: buf.Bytes()}
}

func (bundle *diagnosticsBundle) insert(
	ctx context.Context,
	fingerprint string,
	ast tree.Statement,
	stmtDiagRecorder *stmtdiagnostics.Registry,
	diagRequestID stmtdiagnostics.RequestID,
) {
	__antithesis_instrumentation__.Notify(491033)
	var err error
	bundle.diagID, err = stmtDiagRecorder.InsertStatementDiagnostics(
		ctx,
		diagRequestID,
		fingerprint,
		tree.AsString(ast),
		bundle.zip,
		bundle.collectionErr,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(491034)
		log.Warningf(ctx, "failed to report statement diagnostics: %s", err)
		if bundle.collectionErr != nil {
			__antithesis_instrumentation__.Notify(491035)
			bundle.collectionErr = err
		} else {
			__antithesis_instrumentation__.Notify(491036)
		}
	} else {
		__antithesis_instrumentation__.Notify(491037)
	}
}

type stmtBundleBuilder struct {
	db *kv.DB
	ie *InternalExecutor

	plan         *planTop
	trace        tracing.Recording
	placeholders *tree.PlaceholderInfo

	z memzipper.Zipper
}

func makeStmtBundleBuilder(
	db *kv.DB,
	ie *InternalExecutor,
	plan *planTop,
	trace tracing.Recording,
	placeholders *tree.PlaceholderInfo,
) stmtBundleBuilder {
	__antithesis_instrumentation__.Notify(491038)
	b := stmtBundleBuilder{db: db, ie: ie, plan: plan, trace: trace, placeholders: placeholders}
	b.z.Init()
	return b
}

func (b *stmtBundleBuilder) addStatement() {
	__antithesis_instrumentation__.Notify(491039)
	cfg := tree.DefaultPrettyCfg()
	cfg.UseTabs = false
	cfg.LineWidth = 100
	cfg.TabWidth = 2
	cfg.Simplify = true
	cfg.Align = tree.PrettyNoAlign
	cfg.JSONFmt = true
	var output string

	switch {
	case b.plan.stmt == nil:
		__antithesis_instrumentation__.Notify(491042)
		output = "-- No Statement."
	case b.plan.stmt.AST == nil:
		__antithesis_instrumentation__.Notify(491043)
		output = "-- No AST."
	default:
		__antithesis_instrumentation__.Notify(491044)
		output = cfg.Pretty(b.plan.stmt.AST)
	}
	__antithesis_instrumentation__.Notify(491040)

	if b.placeholders != nil && func() bool {
		__antithesis_instrumentation__.Notify(491045)
		return len(b.placeholders.Values) != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(491046)
		var buf bytes.Buffer
		buf.WriteString(output)
		buf.WriteString("\n\n-- Arguments:\n")
		for i, v := range b.placeholders.Values {
			__antithesis_instrumentation__.Notify(491048)
			fmt.Fprintf(&buf, "--  %s: %v\n", tree.PlaceholderIdx(i), v)
		}
		__antithesis_instrumentation__.Notify(491047)
		output = buf.String()
	} else {
		__antithesis_instrumentation__.Notify(491049)
	}
	__antithesis_instrumentation__.Notify(491041)

	b.z.AddFile("statement.sql", output)
}

func (b *stmtBundleBuilder) addOptPlans() {
	__antithesis_instrumentation__.Notify(491050)
	if b.plan.mem == nil || func() bool {
		__antithesis_instrumentation__.Notify(491053)
		return b.plan.mem.RootExpr() == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(491054)

		return
	} else {
		__antithesis_instrumentation__.Notify(491055)
	}
	__antithesis_instrumentation__.Notify(491051)

	formatOptPlan := func(flags memo.ExprFmtFlags) string {
		__antithesis_instrumentation__.Notify(491056)
		f := memo.MakeExprFmtCtx(flags, b.plan.mem, b.plan.catalog)
		f.FormatExpr(b.plan.mem.RootExpr())
		return f.Buffer.String()
	}
	__antithesis_instrumentation__.Notify(491052)

	b.z.AddFile("opt.txt", formatOptPlan(memo.ExprFmtHideAll))
	b.z.AddFile("opt-v.txt", formatOptPlan(
		memo.ExprFmtHideQualifications|memo.ExprFmtHideScalars|memo.ExprFmtHideTypes,
	))
	b.z.AddFile("opt-vv.txt", formatOptPlan(memo.ExprFmtHideQualifications))
}

func (b *stmtBundleBuilder) addExecPlan(plan string) {
	__antithesis_instrumentation__.Notify(491057)
	if plan != "" {
		__antithesis_instrumentation__.Notify(491058)
		b.z.AddFile("plan.txt", plan)
	} else {
		__antithesis_instrumentation__.Notify(491059)
	}
}

func (b *stmtBundleBuilder) addDistSQLDiagrams() {
	__antithesis_instrumentation__.Notify(491060)
	for i, d := range b.plan.distSQLFlowInfos {
		__antithesis_instrumentation__.Notify(491061)
		d.diagram.AddSpans(b.trace)
		_, url, err := d.diagram.ToURL()

		var contents string
		if err != nil {
			__antithesis_instrumentation__.Notify(491064)
			contents = err.Error()
		} else {
			__antithesis_instrumentation__.Notify(491065)
			contents = fmt.Sprintf(`<meta http-equiv="Refresh" content="0; url=%s">`, url.String())
		}
		__antithesis_instrumentation__.Notify(491062)

		var filename string
		if len(b.plan.distSQLFlowInfos) == 1 {
			__antithesis_instrumentation__.Notify(491066)
			filename = "distsql.html"
		} else {
			__antithesis_instrumentation__.Notify(491067)
			filename = fmt.Sprintf("distsql-%d-%s.html", i+1, d.typ)
		}
		__antithesis_instrumentation__.Notify(491063)
		b.z.AddFile(filename, contents)
	}
}

func (b *stmtBundleBuilder) addExplainVec() {
	__antithesis_instrumentation__.Notify(491068)
	for i, d := range b.plan.distSQLFlowInfos {
		__antithesis_instrumentation__.Notify(491069)
		if len(d.explainVec) > 0 || func() bool {
			__antithesis_instrumentation__.Notify(491070)
			return len(d.explainVecVerbose) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(491071)
			extra := ""
			if len(b.plan.distSQLFlowInfos) > 1 {
				__antithesis_instrumentation__.Notify(491074)
				extra = fmt.Sprintf("-%d-%s", i+1, d.typ)
			} else {
				__antithesis_instrumentation__.Notify(491075)
			}
			__antithesis_instrumentation__.Notify(491072)
			if len(d.explainVec) > 0 {
				__antithesis_instrumentation__.Notify(491076)
				b.z.AddFile(fmt.Sprintf("vec%s.txt", extra), strings.Join(d.explainVec, "\n"))
			} else {
				__antithesis_instrumentation__.Notify(491077)
			}
			__antithesis_instrumentation__.Notify(491073)
			if len(d.explainVecVerbose) > 0 {
				__antithesis_instrumentation__.Notify(491078)
				b.z.AddFile(fmt.Sprintf("vec%s-v.txt", extra), strings.Join(d.explainVecVerbose, "\n"))
			} else {
				__antithesis_instrumentation__.Notify(491079)
			}
		} else {
			__antithesis_instrumentation__.Notify(491080)
		}
	}
}

func (b *stmtBundleBuilder) addTrace() {
	__antithesis_instrumentation__.Notify(491081)
	traceJSONStr, err := tracing.TraceToJSON(b.trace)
	if err != nil {
		__antithesis_instrumentation__.Notify(491083)
		b.z.AddFile("trace.json", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(491084)
		b.z.AddFile("trace.json", traceJSONStr)
	}
	__antithesis_instrumentation__.Notify(491082)

	cfg := tree.DefaultPrettyCfg()
	cfg.UseTabs = false
	cfg.LineWidth = 100
	cfg.TabWidth = 2
	cfg.Simplify = true
	cfg.Align = tree.PrettyNoAlign
	cfg.JSONFmt = true
	stmt := cfg.Pretty(b.plan.stmt.AST)

	b.z.AddFile("trace.txt", fmt.Sprintf("%s\n\n\n\n%s", stmt, b.trace.String()))

	comment := fmt.Sprintf(`This is a trace for SQL statement: %s
This trace can be imported into Jaeger for visualization. From the Jaeger Search screen, select the JSON File.
Jaeger can be started using docker with: docker run -d --name jaeger -p 16686:16686 jaegertracing/all-in-one:1.17
The UI can then be accessed at http://localhost:16686/search`, stmt)
	jaegerJSON, err := b.trace.ToJaegerJSON(stmt, comment, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(491085)
		b.z.AddFile("trace-jaeger.txt", err.Error())
	} else {
		__antithesis_instrumentation__.Notify(491086)
		b.z.AddFile("trace-jaeger.json", jaegerJSON)
	}
}

func (b *stmtBundleBuilder) addEnv(ctx context.Context) {
	__antithesis_instrumentation__.Notify(491087)
	c := makeStmtEnvCollector(ctx, b.ie)

	var buf bytes.Buffer
	if err := c.PrintVersion(&buf); err != nil {
		__antithesis_instrumentation__.Notify(491099)
		fmt.Fprintf(&buf, "-- error getting version: %v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(491100)
	}
	__antithesis_instrumentation__.Notify(491088)
	fmt.Fprintf(&buf, "\n")

	if err := c.PrintSessionSettings(&buf); err != nil {
		__antithesis_instrumentation__.Notify(491101)
		fmt.Fprintf(&buf, "-- error getting session settings: %v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(491102)
	}
	__antithesis_instrumentation__.Notify(491089)

	fmt.Fprintf(&buf, "\n")

	if err := c.PrintClusterSettings(&buf); err != nil {
		__antithesis_instrumentation__.Notify(491103)
		fmt.Fprintf(&buf, "-- error getting cluster settings: %v\n", err)
	} else {
		__antithesis_instrumentation__.Notify(491104)
	}
	__antithesis_instrumentation__.Notify(491090)

	b.z.AddFile("env.sql", buf.String())

	mem := b.plan.mem
	if mem == nil {
		__antithesis_instrumentation__.Notify(491105)

		return
	} else {
		__antithesis_instrumentation__.Notify(491106)
	}
	__antithesis_instrumentation__.Notify(491091)
	buf.Reset()

	var tables, sequences, views []tree.TableName
	err := b.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(491107)
		var err error
		tables, sequences, views, err = mem.Metadata().AllDataSourceNames(
			func(ds cat.DataSource) (cat.DataSourceName, error) {
				__antithesis_instrumentation__.Notify(491109)
				return b.plan.catalog.fullyQualifiedNameWithTxn(ctx, ds, txn)
			},
		)
		__antithesis_instrumentation__.Notify(491108)
		return err
	})
	__antithesis_instrumentation__.Notify(491092)
	if err != nil {
		__antithesis_instrumentation__.Notify(491110)
		b.z.AddFile("schema.sql", fmt.Sprintf("-- error getting data source names: %v\n", err))
		return
	} else {
		__antithesis_instrumentation__.Notify(491111)
	}
	__antithesis_instrumentation__.Notify(491093)

	if len(tables) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(491112)
		return len(sequences) == 0 == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(491113)
		return len(views) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(491114)
		return
	} else {
		__antithesis_instrumentation__.Notify(491115)
	}
	__antithesis_instrumentation__.Notify(491094)

	first := true
	blankLine := func() {
		__antithesis_instrumentation__.Notify(491116)
		if !first {
			__antithesis_instrumentation__.Notify(491118)
			buf.WriteByte('\n')
		} else {
			__antithesis_instrumentation__.Notify(491119)
		}
		__antithesis_instrumentation__.Notify(491117)
		first = false
	}
	__antithesis_instrumentation__.Notify(491095)
	for i := range sequences {
		__antithesis_instrumentation__.Notify(491120)
		blankLine()
		if err := c.PrintCreateSequence(&buf, &sequences[i]); err != nil {
			__antithesis_instrumentation__.Notify(491121)
			fmt.Fprintf(&buf, "-- error getting schema for sequence %s: %v\n", sequences[i].String(), err)
		} else {
			__antithesis_instrumentation__.Notify(491122)
		}
	}
	__antithesis_instrumentation__.Notify(491096)
	for i := range tables {
		__antithesis_instrumentation__.Notify(491123)
		blankLine()
		if err := c.PrintCreateTable(&buf, &tables[i]); err != nil {
			__antithesis_instrumentation__.Notify(491124)
			fmt.Fprintf(&buf, "-- error getting schema for table %s: %v\n", tables[i].String(), err)
		} else {
			__antithesis_instrumentation__.Notify(491125)
		}
	}
	__antithesis_instrumentation__.Notify(491097)
	for i := range views {
		__antithesis_instrumentation__.Notify(491126)
		blankLine()
		if err := c.PrintCreateView(&buf, &views[i]); err != nil {
			__antithesis_instrumentation__.Notify(491127)
			fmt.Fprintf(&buf, "-- error getting schema for view %s: %v\n", views[i].String(), err)
		} else {
			__antithesis_instrumentation__.Notify(491128)
		}
	}
	__antithesis_instrumentation__.Notify(491098)
	b.z.AddFile("schema.sql", buf.String())
	for i := range tables {
		__antithesis_instrumentation__.Notify(491129)
		buf.Reset()
		if err := c.PrintTableStats(&buf, &tables[i], false); err != nil {
			__antithesis_instrumentation__.Notify(491131)
			fmt.Fprintf(&buf, "-- error getting statistics for table %s: %v\n", tables[i].String(), err)
		} else {
			__antithesis_instrumentation__.Notify(491132)
		}
		__antithesis_instrumentation__.Notify(491130)
		b.z.AddFile(fmt.Sprintf("stats-%s.sql", tables[i].String()), buf.String())
	}
}

func (b *stmtBundleBuilder) finalize() (*bytes.Buffer, error) {
	__antithesis_instrumentation__.Notify(491133)
	return b.z.Finalize()
}

type stmtEnvCollector struct {
	ctx context.Context
	ie  *InternalExecutor
}

func makeStmtEnvCollector(ctx context.Context, ie *InternalExecutor) stmtEnvCollector {
	__antithesis_instrumentation__.Notify(491134)
	return stmtEnvCollector{ctx: ctx, ie: ie}
}

func (c *stmtEnvCollector) query(query string) (string, error) {
	__antithesis_instrumentation__.Notify(491135)
	row, err := c.ie.QueryRowEx(
		c.ctx,
		"stmtEnvCollector",
		nil,
		sessiondata.NoSessionDataOverride,
		query,
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(491139)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(491140)
	}
	__antithesis_instrumentation__.Notify(491136)

	if len(row) != 1 {
		__antithesis_instrumentation__.Notify(491141)
		return "", errors.AssertionFailedf(
			"expected env query %q to return a single column, returned %d",
			query, len(row),
		)
	} else {
		__antithesis_instrumentation__.Notify(491142)
	}
	__antithesis_instrumentation__.Notify(491137)

	s, ok := row[0].(*tree.DString)
	if !ok {
		__antithesis_instrumentation__.Notify(491143)
		return "", errors.AssertionFailedf(
			"expected env query %q to return a DString, returned %T",
			query, row[0],
		)
	} else {
		__antithesis_instrumentation__.Notify(491144)
	}
	__antithesis_instrumentation__.Notify(491138)

	return string(*s), nil
}

var testingOverrideExplainEnvVersion string

func TestingOverrideExplainEnvVersion(ver string) func() {
	__antithesis_instrumentation__.Notify(491145)
	prev := testingOverrideExplainEnvVersion
	testingOverrideExplainEnvVersion = ver
	return func() { __antithesis_instrumentation__.Notify(491146); testingOverrideExplainEnvVersion = prev }
}

func (c *stmtEnvCollector) PrintVersion(w io.Writer) error {
	__antithesis_instrumentation__.Notify(491147)
	version, err := c.query("SELECT version()")
	if err != nil {
		__antithesis_instrumentation__.Notify(491150)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491151)
	}
	__antithesis_instrumentation__.Notify(491148)
	if testingOverrideExplainEnvVersion != "" {
		__antithesis_instrumentation__.Notify(491152)
		version = testingOverrideExplainEnvVersion
	} else {
		__antithesis_instrumentation__.Notify(491153)
	}
	__antithesis_instrumentation__.Notify(491149)
	fmt.Fprintf(w, "-- Version: %s\n", version)
	return err
}

func (c *stmtEnvCollector) PrintSessionSettings(w io.Writer) error {
	__antithesis_instrumentation__.Notify(491154)

	boolToOnOff := func(boolStr string) string {
		__antithesis_instrumentation__.Notify(491159)
		switch boolStr {
		case "true":
			__antithesis_instrumentation__.Notify(491161)
			return "on"
		case "false":
			__antithesis_instrumentation__.Notify(491162)
			return "off"
		default:
			__antithesis_instrumentation__.Notify(491163)
		}
		__antithesis_instrumentation__.Notify(491160)
		return boolStr
	}
	__antithesis_instrumentation__.Notify(491155)

	distsqlConv := func(enumVal string) string {
		__antithesis_instrumentation__.Notify(491164)
		n, err := strconv.ParseInt(enumVal, 10, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(491166)
			return enumVal
		} else {
			__antithesis_instrumentation__.Notify(491167)
		}
		__antithesis_instrumentation__.Notify(491165)
		return sessiondatapb.DistSQLExecMode(n).String()
	}
	__antithesis_instrumentation__.Notify(491156)

	vectorizeConv := func(enumVal string) string {
		__antithesis_instrumentation__.Notify(491168)
		n, err := strconv.ParseInt(enumVal, 10, 32)
		if err != nil {
			__antithesis_instrumentation__.Notify(491170)
			return enumVal
		} else {
			__antithesis_instrumentation__.Notify(491171)
		}
		__antithesis_instrumentation__.Notify(491169)
		return sessiondatapb.VectorizeExecMode(n).String()
	}
	__antithesis_instrumentation__.Notify(491157)

	relevantSettings := []struct {
		sessionSetting string
		clusterSetting settings.NonMaskedSetting
		convFunc       func(string) string
	}{
		{sessionSetting: "reorder_joins_limit", clusterSetting: ReorderJoinsLimitClusterValue},
		{sessionSetting: "enable_zigzag_join", clusterSetting: zigzagJoinClusterMode, convFunc: boolToOnOff},
		{sessionSetting: "optimizer_use_histograms", clusterSetting: optUseHistogramsClusterMode, convFunc: boolToOnOff},
		{sessionSetting: "optimizer_use_multicol_stats", clusterSetting: optUseMultiColStatsClusterMode, convFunc: boolToOnOff},
		{sessionSetting: "locality_optimized_partitioned_index_scan", clusterSetting: localityOptimizedSearchMode, convFunc: boolToOnOff},
		{sessionSetting: "propagate_input_ordering", clusterSetting: propagateInputOrdering, convFunc: boolToOnOff},
		{sessionSetting: "prefer_lookup_joins_for_fks", clusterSetting: preferLookupJoinsForFKs, convFunc: boolToOnOff},
		{sessionSetting: "disallow_full_table_scans", clusterSetting: disallowFullTableScans, convFunc: boolToOnOff},
		{sessionSetting: "large_full_scan_rows", clusterSetting: largeFullScanRows},
		{sessionSetting: "cost_scans_with_default_col_size", clusterSetting: costScansWithDefaultColSize, convFunc: boolToOnOff},
		{sessionSetting: "default_transaction_quality_of_service"},
		{sessionSetting: "distsql", clusterSetting: DistSQLClusterExecMode, convFunc: distsqlConv},
		{sessionSetting: "vectorize", clusterSetting: VectorizeClusterMode, convFunc: vectorizeConv},
	}

	for _, s := range relevantSettings {
		__antithesis_instrumentation__.Notify(491172)
		value, err := c.query(fmt.Sprintf("SHOW %s", s.sessionSetting))
		if err != nil {
			__antithesis_instrumentation__.Notify(491176)
			return err
		} else {
			__antithesis_instrumentation__.Notify(491177)
		}
		__antithesis_instrumentation__.Notify(491173)

		var def string
		if s.clusterSetting == nil {
			__antithesis_instrumentation__.Notify(491178)

			def = sessiondatapb.Normal.String()
		} else {
			__antithesis_instrumentation__.Notify(491179)
			def = s.clusterSetting.EncodedDefault()
		}
		__antithesis_instrumentation__.Notify(491174)
		if s.convFunc != nil {
			__antithesis_instrumentation__.Notify(491180)

			def = s.convFunc(def)
		} else {
			__antithesis_instrumentation__.Notify(491181)
		}
		__antithesis_instrumentation__.Notify(491175)

		if value == def {
			__antithesis_instrumentation__.Notify(491182)
			fmt.Fprintf(w, "-- %s has the default value: %s\n", s.sessionSetting, value)
		} else {
			__antithesis_instrumentation__.Notify(491183)
			fmt.Fprintf(w, "SET %s = %s;  -- default value: %s\n", s.sessionSetting, value, def)
		}
	}
	__antithesis_instrumentation__.Notify(491158)
	return nil
}

func (c *stmtEnvCollector) PrintClusterSettings(w io.Writer) error {
	__antithesis_instrumentation__.Notify(491184)
	rows, err := c.ie.QueryBufferedEx(
		c.ctx,
		"stmtEnvCollector",
		nil,
		sessiondata.NoSessionDataOverride,
		"SELECT variable, value, description FROM [ SHOW ALL CLUSTER SETTINGS ]",
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(491187)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491188)
	}
	__antithesis_instrumentation__.Notify(491185)
	fmt.Fprintf(w, "-- Cluster settings:\n")
	for _, r := range rows {
		__antithesis_instrumentation__.Notify(491189)

		variable, ok1 := r[0].(*tree.DString)
		value, ok2 := r[1].(*tree.DString)
		description, ok3 := r[2].(*tree.DString)
		if ok1 && func() bool {
			__antithesis_instrumentation__.Notify(491190)
			return ok2 == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(491191)
			return ok3 == true
		}() == true {
			__antithesis_instrumentation__.Notify(491192)
			fmt.Fprintf(w, "--   %s = %s  (%s)\n", *variable, *value, *description)
		} else {
			__antithesis_instrumentation__.Notify(491193)
		}
	}
	__antithesis_instrumentation__.Notify(491186)
	return nil
}

func (c *stmtEnvCollector) PrintCreateTable(w io.Writer, tn *tree.TableName) error {
	__antithesis_instrumentation__.Notify(491194)
	createStatement, err := c.query(
		fmt.Sprintf("SELECT create_statement FROM [SHOW CREATE TABLE %s]", tn.String()),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(491196)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491197)
	}
	__antithesis_instrumentation__.Notify(491195)
	fmt.Fprintf(w, "%s;\n", createStatement)
	return nil
}

func (c *stmtEnvCollector) PrintCreateSequence(w io.Writer, tn *tree.TableName) error {
	__antithesis_instrumentation__.Notify(491198)
	createStatement, err := c.query(fmt.Sprintf(
		"SELECT create_statement FROM [SHOW CREATE SEQUENCE %s]", tn.String(),
	))
	if err != nil {
		__antithesis_instrumentation__.Notify(491200)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491201)
	}
	__antithesis_instrumentation__.Notify(491199)
	fmt.Fprintf(w, "%s;\n", createStatement)
	return nil
}

func (c *stmtEnvCollector) PrintCreateView(w io.Writer, tn *tree.TableName) error {
	__antithesis_instrumentation__.Notify(491202)
	createStatement, err := c.query(fmt.Sprintf(
		"SELECT create_statement FROM [SHOW CREATE VIEW %s]", tn.String(),
	))
	if err != nil {
		__antithesis_instrumentation__.Notify(491204)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491205)
	}
	__antithesis_instrumentation__.Notify(491203)
	fmt.Fprintf(w, "%s;\n", createStatement)
	return nil
}

func (c *stmtEnvCollector) PrintTableStats(
	w io.Writer, tn *tree.TableName, hideHistograms bool,
) error {
	__antithesis_instrumentation__.Notify(491206)
	var maybeRemoveHistoBuckets string
	if hideHistograms {
		__antithesis_instrumentation__.Notify(491209)
		maybeRemoveHistoBuckets = " - 'histo_buckets'"
	} else {
		__antithesis_instrumentation__.Notify(491210)
	}
	__antithesis_instrumentation__.Notify(491207)

	stats, err := c.query(fmt.Sprintf(
		`SELECT jsonb_pretty(COALESCE(json_agg(stat), '[]'))
		 FROM (
			 SELECT json_array_elements(statistics)%s AS stat
			 FROM [SHOW STATISTICS USING JSON FOR TABLE %s]
		 )`,
		maybeRemoveHistoBuckets, tn.String(),
	))
	if err != nil {
		__antithesis_instrumentation__.Notify(491211)
		return err
	} else {
		__antithesis_instrumentation__.Notify(491212)
	}
	__antithesis_instrumentation__.Notify(491208)

	stats = strings.Replace(stats, "'", "''", -1)

	explicitCatalog := tn.ExplicitCatalog
	tn.ExplicitCatalog = false
	fmt.Fprintf(w, "ALTER TABLE %s INJECT STATISTICS '%s';\n", tn.String(), stats)
	tn.ExplicitCatalog = explicitCatalog
	return nil
}
