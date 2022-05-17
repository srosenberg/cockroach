package sqltestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gosql "database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"testing"
	"text/tabwriter"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/server"
	"github.com/cockroachdb/cockroach/pkg/server/diagnostics"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/testutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/diagutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	"github.com/cockroachdb/cockroach/pkg/testutils/skip"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/cloudinfo"
	"github.com/cockroachdb/cockroach/pkg/util/treeprinter"
	"github.com/cockroachdb/datadriven"
	"github.com/cockroachdb/errors"
)

func TelemetryTest(t *testing.T, serverArgs []base.TestServerArgs, testTenant bool) {
	__antithesis_instrumentation__.Notify(625923)

	datadriven.Walk(t, "testdata/telemetry", func(t *testing.T, path string) {
		__antithesis_instrumentation__.Notify(625924)

		defer cloudinfo.Disable()()

		var test telemetryTest
		test.Start(t, serverArgs)
		defer test.Close()

		t.Run("server", func(t *testing.T) {
			__antithesis_instrumentation__.Notify(625926)
			datadriven.RunTest(t, path, func(t *testing.T, td *datadriven.TestData) string {
				__antithesis_instrumentation__.Notify(625927)
				sqlServer := test.server.SQLServer().(*sql.Server)
				reporter := test.server.DiagnosticsReporter().(*diagnostics.Reporter)
				return test.RunTest(td, test.serverDB, reporter.ReportDiagnostics, sqlServer)
			})
		})
		__antithesis_instrumentation__.Notify(625925)

		if testTenant {
			__antithesis_instrumentation__.Notify(625928)

			t.Run("tenant", func(t *testing.T) {
				__antithesis_instrumentation__.Notify(625929)

				switch path {
				case "testdata/telemetry/execution",
					"testdata/telemetry/planning",
					"testdata/telemetry/sql-stats":
					__antithesis_instrumentation__.Notify(625931)
					skip.WithIssue(t, 47893, "tenant clusters do not support SQL features used by this test")
				default:
					__antithesis_instrumentation__.Notify(625932)
				}
				__antithesis_instrumentation__.Notify(625930)

				datadriven.RunTest(t, path, func(t *testing.T, td *datadriven.TestData) string {
					__antithesis_instrumentation__.Notify(625933)
					sqlServer := test.server.SQLServer().(*sql.Server)
					reporter := test.tenant.DiagnosticsReporter().(*diagnostics.Reporter)
					return test.RunTest(td, test.tenantDB, reporter.ReportDiagnostics, sqlServer)
				})
			})
		} else {
			__antithesis_instrumentation__.Notify(625934)
		}
	})
}

type telemetryTest struct {
	t              *testing.T
	diagSrv        *diagutils.Server
	cluster        serverutils.TestClusterInterface
	server         serverutils.TestServerInterface
	serverDB       *gosql.DB
	tenant         serverutils.TestTenantInterface
	tenantDB       *gosql.DB
	tempDirCleanup func()
	allowlist      featureAllowlist
	rewrites       []rewrite
}

type rewrite struct {
	pattern     *regexp.Regexp
	replacement string
}

func (tt *telemetryTest) Start(t *testing.T, serverArgs []base.TestServerArgs) {
	__antithesis_instrumentation__.Notify(625935)
	tt.t = t
	tt.diagSrv = diagutils.NewServer()

	var tempExternalIODir string
	tempExternalIODir, tt.tempDirCleanup = testutils.TempDir(tt.t)

	diagSrvURL := tt.diagSrv.URL()
	mapServerArgs := make(map[int]base.TestServerArgs, len(serverArgs))
	for i, v := range serverArgs {
		__antithesis_instrumentation__.Notify(625937)
		v.Knobs.Server = &server.TestingKnobs{
			DiagnosticsTestingKnobs: diagnostics.TestingKnobs{
				OverrideReportingURL: &diagSrvURL,
			},
		}
		v.ExternalIODir = tempExternalIODir
		mapServerArgs[i] = v
	}
	__antithesis_instrumentation__.Notify(625936)
	tt.cluster = serverutils.StartNewTestCluster(
		tt.t,
		len(serverArgs),
		base.TestClusterArgs{ServerArgsPerNode: mapServerArgs},
	)
	tt.server = tt.cluster.Server(0)
	tt.serverDB = tt.cluster.ServerConn(0)
	tt.prepareCluster(tt.serverDB)

	tt.tenant, tt.tenantDB = serverutils.StartTenant(tt.t, tt.server, base.TestTenantArgs{
		TenantID:                    serverutils.TestTenantID(),
		AllowSettingClusterSettings: true,
		TestingKnobs:                mapServerArgs[0].Knobs,
	})
	tt.prepareCluster(tt.tenantDB)
}

func (tt *telemetryTest) Close() {
	__antithesis_instrumentation__.Notify(625938)
	tt.cluster.Stopper().Stop(context.Background())
	tt.diagSrv.Close()
	tt.tempDirCleanup()
}

func (tt *telemetryTest) RunTest(
	td *datadriven.TestData,
	db *gosql.DB,
	reportDiags func(ctx context.Context),
	sqlServer *sql.Server,
) (out string) {
	__antithesis_instrumentation__.Notify(625939)
	defer func() {
		__antithesis_instrumentation__.Notify(625941)
		if out == "" {
			__antithesis_instrumentation__.Notify(625943)
			return
		} else {
			__antithesis_instrumentation__.Notify(625944)
		}
		__antithesis_instrumentation__.Notify(625942)
		for _, r := range tt.rewrites {
			__antithesis_instrumentation__.Notify(625945)
			in := out
			out = r.pattern.ReplaceAllString(out, r.replacement)
			tt.t.Log(r.pattern, r.replacement, in == out, r.pattern.MatchString(out), out)
		}
	}()
	__antithesis_instrumentation__.Notify(625940)
	ctx := context.Background()
	switch td.Cmd {
	case "exec":
		__antithesis_instrumentation__.Notify(625946)
		_, err := db.Exec(td.Input)
		if err != nil {
			__antithesis_instrumentation__.Notify(625962)
			if errors.HasAssertionFailure(err) {
				__antithesis_instrumentation__.Notify(625964)
				td.Fatalf(tt.t, "%+v", err)
			} else {
				__antithesis_instrumentation__.Notify(625965)
			}
			__antithesis_instrumentation__.Notify(625963)
			return fmt.Sprintf("error: %v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(625966)
		}
		__antithesis_instrumentation__.Notify(625947)
		return ""

	case "schema":
		__antithesis_instrumentation__.Notify(625948)
		reportDiags(ctx)
		last := tt.diagSrv.LastRequestData()
		var buf bytes.Buffer
		for i := range last.Schema {
			__antithesis_instrumentation__.Notify(625967)
			buf.WriteString(formatTableDescriptor(&last.Schema[i]))
		}
		__antithesis_instrumentation__.Notify(625949)
		return buf.String()

	case "feature-allowlist":
		__antithesis_instrumentation__.Notify(625950)
		var err error
		tt.allowlist, err = makeAllowlist(strings.Split(td.Input, "\n"))
		if err != nil {
			__antithesis_instrumentation__.Notify(625968)
			td.Fatalf(tt.t, "error parsing feature regex: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(625969)
		}
		__antithesis_instrumentation__.Notify(625951)
		return ""

	case "feature-usage", "feature-counters":
		__antithesis_instrumentation__.Notify(625952)

		reportDiags(ctx)
		_, err := db.Exec(td.Input)
		var buf bytes.Buffer
		if err != nil {
			__antithesis_instrumentation__.Notify(625970)
			fmt.Fprintf(&buf, "error: %v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(625971)
		}
		__antithesis_instrumentation__.Notify(625953)
		reportDiags(ctx)
		last := tt.diagSrv.LastRequestData()
		usage := last.FeatureUsage
		keys := make([]string, 0, len(usage))
		for k, v := range usage {
			__antithesis_instrumentation__.Notify(625972)
			if v == 0 {
				__antithesis_instrumentation__.Notify(625975)

				continue
			} else {
				__antithesis_instrumentation__.Notify(625976)
			}
			__antithesis_instrumentation__.Notify(625973)
			if !tt.allowlist.Match(k) {
				__antithesis_instrumentation__.Notify(625977)

				continue
			} else {
				__antithesis_instrumentation__.Notify(625978)
			}
			__antithesis_instrumentation__.Notify(625974)
			keys = append(keys, k)
		}
		__antithesis_instrumentation__.Notify(625954)
		sort.Strings(keys)
		tw := tabwriter.NewWriter(&buf, 2, 1, 2, ' ', 0)
		for _, k := range keys {
			__antithesis_instrumentation__.Notify(625979)

			if td.Cmd == "feature-counters" {
				__antithesis_instrumentation__.Notify(625980)
				fmt.Fprintf(tw, "%s\t%d\n", k, usage[k])
			} else {
				__antithesis_instrumentation__.Notify(625981)
				fmt.Fprintf(tw, "%s\n", k)
			}
		}
		__antithesis_instrumentation__.Notify(625955)
		_ = tw.Flush()
		return buf.String()

	case "sql-stats":
		__antithesis_instrumentation__.Notify(625956)

		sqlServer.GetSQLStatsController().ResetLocalSQLStats(ctx)
		reportDiags(ctx)

		_, err := db.Exec(td.Input)
		var buf bytes.Buffer
		if err != nil {
			__antithesis_instrumentation__.Notify(625982)
			fmt.Fprintf(&buf, "error: %v\n", err)
		} else {
			__antithesis_instrumentation__.Notify(625983)
		}
		__antithesis_instrumentation__.Notify(625957)
		sqlServer.GetSQLStatsController().ResetLocalSQLStats(ctx)
		reportDiags(ctx)
		last := tt.diagSrv.LastRequestData()
		buf.WriteString(formatSQLStats(last.SqlStats))
		return buf.String()

	case "rewrite":
		__antithesis_instrumentation__.Notify(625958)
		lines := strings.Split(td.Input, "\n")
		if len(lines) != 2 {
			__antithesis_instrumentation__.Notify(625984)
			td.Fatalf(tt.t, "rewrite: expected two lines")
		} else {
			__antithesis_instrumentation__.Notify(625985)
		}
		__antithesis_instrumentation__.Notify(625959)
		pattern, err := regexp.Compile(lines[0])
		if err != nil {
			__antithesis_instrumentation__.Notify(625986)
			td.Fatalf(tt.t, "rewrite: invalid pattern: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(625987)
		}
		__antithesis_instrumentation__.Notify(625960)
		tt.rewrites = append(tt.rewrites, rewrite{
			pattern:     pattern,
			replacement: lines[1],
		})
		return ""

	default:
		__antithesis_instrumentation__.Notify(625961)
		td.Fatalf(tt.t, "unknown command %s", td.Cmd)
		return ""
	}
}

func (tt *telemetryTest) prepareCluster(db *gosql.DB) {
	__antithesis_instrumentation__.Notify(625988)
	runner := sqlutils.MakeSQLRunner(db)

	runner.Exec(tt.t, "SET CLUSTER SETTING diagnostics.reporting.enabled = false")
	runner.Exec(tt.t, "SET CLUSTER SETTING diagnostics.reporting.send_crash_reports = false")

	runner.Exec(tt.t, "SET CLUSTER SETTING sql.query_cache.enabled = false")
}

type featureAllowlist []*regexp.Regexp

func makeAllowlist(strings []string) (featureAllowlist, error) {
	__antithesis_instrumentation__.Notify(625989)
	w := make(featureAllowlist, len(strings))
	for i := range strings {
		__antithesis_instrumentation__.Notify(625991)
		var err error
		w[i], err = regexp.Compile("^" + strings[i] + "$")
		if err != nil {
			__antithesis_instrumentation__.Notify(625992)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(625993)
		}
	}
	__antithesis_instrumentation__.Notify(625990)
	return w, nil
}

func (w featureAllowlist) Match(feature string) bool {
	__antithesis_instrumentation__.Notify(625994)
	if w == nil {
		__antithesis_instrumentation__.Notify(625997)

		return true
	} else {
		__antithesis_instrumentation__.Notify(625998)
	}
	__antithesis_instrumentation__.Notify(625995)
	for _, r := range w {
		__antithesis_instrumentation__.Notify(625999)
		if r.MatchString(feature) {
			__antithesis_instrumentation__.Notify(626000)
			return true
		} else {
			__antithesis_instrumentation__.Notify(626001)
		}
	}
	__antithesis_instrumentation__.Notify(625996)
	return false
}

func formatTableDescriptor(desc *descpb.TableDescriptor) string {
	__antithesis_instrumentation__.Notify(626002)
	tp := treeprinter.New()
	n := tp.Childf("table:%s", desc.Name)
	cols := n.Child("columns")
	for _, col := range desc.Columns {
		__antithesis_instrumentation__.Notify(626005)
		var colBuf bytes.Buffer
		fmt.Fprintf(&colBuf, "%s:%s", col.Name, col.Type.String())
		if col.DefaultExpr != nil {
			__antithesis_instrumentation__.Notify(626008)
			fmt.Fprintf(&colBuf, " default: %s", *col.DefaultExpr)
		} else {
			__antithesis_instrumentation__.Notify(626009)
		}
		__antithesis_instrumentation__.Notify(626006)
		if col.ComputeExpr != nil {
			__antithesis_instrumentation__.Notify(626010)
			fmt.Fprintf(&colBuf, " computed: %s", *col.ComputeExpr)
		} else {
			__antithesis_instrumentation__.Notify(626011)
		}
		__antithesis_instrumentation__.Notify(626007)
		cols.Child(colBuf.String())
	}
	__antithesis_instrumentation__.Notify(626003)
	if len(desc.Checks) > 0 {
		__antithesis_instrumentation__.Notify(626012)
		checks := n.Child("checks")
		for _, chk := range desc.Checks {
			__antithesis_instrumentation__.Notify(626013)
			checks.Childf("%s: %s", chk.Name, chk.Expr)
		}
	} else {
		__antithesis_instrumentation__.Notify(626014)
	}
	__antithesis_instrumentation__.Notify(626004)
	return tp.String()
}

func formatSQLStats(stats []roachpb.CollectedStatementStatistics) string {
	__antithesis_instrumentation__.Notify(626015)
	bucketByApp := make(map[string][]roachpb.CollectedStatementStatistics)
	for i := range stats {
		__antithesis_instrumentation__.Notify(626019)
		s := &stats[i]

		if strings.HasPrefix(s.Key.App, catconstants.InternalAppNamePrefix) {
			__antithesis_instrumentation__.Notify(626021)

			continue
		} else {
			__antithesis_instrumentation__.Notify(626022)
		}
		__antithesis_instrumentation__.Notify(626020)
		bucketByApp[s.Key.App] = append(bucketByApp[s.Key.App], *s)
	}
	__antithesis_instrumentation__.Notify(626016)
	var apps []string
	for app, s := range bucketByApp {
		__antithesis_instrumentation__.Notify(626023)
		apps = append(apps, app)
		sort.Slice(s, func(i, j int) bool {
			__antithesis_instrumentation__.Notify(626025)
			return s[i].Key.Query < s[j].Key.Query
		})
		__antithesis_instrumentation__.Notify(626024)
		bucketByApp[app] = s
	}
	__antithesis_instrumentation__.Notify(626017)
	sort.Strings(apps)
	tp := treeprinter.New()
	n := tp.Child("sql-stats")

	for _, app := range apps {
		__antithesis_instrumentation__.Notify(626026)
		nodeApp := n.Child(app)
		for _, s := range bucketByApp[app] {
			__antithesis_instrumentation__.Notify(626027)
			var flags []string
			if s.Key.Failed {
				__antithesis_instrumentation__.Notify(626031)
				flags = append(flags, "failed")
			} else {
				__antithesis_instrumentation__.Notify(626032)
			}
			__antithesis_instrumentation__.Notify(626028)
			if !s.Key.DistSQL {
				__antithesis_instrumentation__.Notify(626033)
				flags = append(flags, "nodist")
			} else {
				__antithesis_instrumentation__.Notify(626034)
			}
			__antithesis_instrumentation__.Notify(626029)
			var buf bytes.Buffer
			if len(flags) > 0 {
				__antithesis_instrumentation__.Notify(626035)
				buf.WriteString("[")
				for i := range flags {
					__antithesis_instrumentation__.Notify(626037)
					if i > 0 {
						__antithesis_instrumentation__.Notify(626039)
						buf.WriteByte(',')
					} else {
						__antithesis_instrumentation__.Notify(626040)
					}
					__antithesis_instrumentation__.Notify(626038)
					buf.WriteString(flags[i])
				}
				__antithesis_instrumentation__.Notify(626036)
				buf.WriteString("] ")
			} else {
				__antithesis_instrumentation__.Notify(626041)
			}
			__antithesis_instrumentation__.Notify(626030)
			buf.WriteString(s.Key.Query)
			nodeApp.Child(buf.String())
		}
	}
	__antithesis_instrumentation__.Notify(626018)
	return tp.String()
}
