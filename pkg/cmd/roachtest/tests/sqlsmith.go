package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/internal/sqlsmith"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
	"github.com/cockroachdb/errors"
)

func registerSQLSmith(r registry.Registry) {
	__antithesis_instrumentation__.Notify(51046)
	const numNodes = 4
	setups := map[string]sqlsmith.Setup{
		"empty":                     sqlsmith.Setups["empty"],
		"seed":                      sqlsmith.Setups["seed"],
		sqlsmith.RandTableSetupName: sqlsmith.Setups[sqlsmith.RandTableSetupName],
		"tpch-sf1": func(r *rand.Rand) []string {
			__antithesis_instrumentation__.Notify(51051)
			return []string{`RESTORE TABLE tpch.* FROM 'gs://cockroach-fixtures/workload/tpch/scalefactor=1/backup?AUTH=implicit' WITH into_db = 'defaultdb';`}
		},
		"tpcc": func(r *rand.Rand) []string {
			__antithesis_instrumentation__.Notify(51052)
			const version = "version=2.1.0,fks=true,interleaved=false,seed=1,warehouses=1"
			var stmts []string
			for _, t := range []string{
				"customer",
				"district",
				"history",
				"item",
				"new_order",
				"order",
				"order_line",
				"stock",
				"warehouse",
			} {
				__antithesis_instrumentation__.Notify(51054)
				stmts = append(
					stmts,
					fmt.Sprintf("RESTORE TABLE tpcc.%s FROM 'gs://cockroach-fixtures/workload/tpcc/%[2]s/%[1]s?AUTH=implicit' WITH into_db = 'defaultdb';",
						t, version,
					),
				)
			}
			__antithesis_instrumentation__.Notify(51053)
			return stmts
		},
	}
	__antithesis_instrumentation__.Notify(51047)
	settings := map[string]sqlsmith.SettingFunc{
		"default":      sqlsmith.Settings["default"],
		"no-mutations": sqlsmith.Settings["no-mutations"],
		"no-ddl":       sqlsmith.Settings["no-ddl"],
	}

	runSQLSmith := func(ctx context.Context, t test.Test, c cluster.Cluster, setupName, settingName string) {
		__antithesis_instrumentation__.Notify(51055)

		smithLog, err := os.Create(filepath.Join(t.ArtifactsDir(), "sqlsmith.log"))
		if err != nil {
			__antithesis_instrumentation__.Notify(51067)
			t.Fatalf("could not create sqlsmith.log: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(51068)
		}
		__antithesis_instrumentation__.Notify(51056)
		defer smithLog.Close()
		logStmt := func(stmt string) {
			__antithesis_instrumentation__.Notify(51069)
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				__antithesis_instrumentation__.Notify(51072)
				return
			} else {
				__antithesis_instrumentation__.Notify(51073)
			}
			__antithesis_instrumentation__.Notify(51070)
			fmt.Fprint(smithLog, stmt)
			if !strings.HasSuffix(stmt, ";") {
				__antithesis_instrumentation__.Notify(51074)
				fmt.Fprint(smithLog, ";")
			} else {
				__antithesis_instrumentation__.Notify(51075)
			}
			__antithesis_instrumentation__.Notify(51071)
			fmt.Fprint(smithLog, "\n\n")
		}
		__antithesis_instrumentation__.Notify(51057)

		rng, seed := randutil.NewTestRand()
		t.L().Printf("seed: %d", seed)

		c.Put(ctx, t.Cockroach(), "./cockroach")
		if err := c.PutLibraries(ctx, "./lib"); err != nil {
			__antithesis_instrumentation__.Notify(51076)
			t.Fatalf("could not initialize libraries: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(51077)
		}
		__antithesis_instrumentation__.Notify(51058)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

		setupFunc, ok := setups[setupName]
		if !ok {
			__antithesis_instrumentation__.Notify(51078)
			t.Fatalf("unknown setup %s", setupName)
		} else {
			__antithesis_instrumentation__.Notify(51079)
		}
		__antithesis_instrumentation__.Notify(51059)
		settingFunc, ok := settings[settingName]
		if !ok {
			__antithesis_instrumentation__.Notify(51080)
			t.Fatalf("unknown setting %s", settingName)
		} else {
			__antithesis_instrumentation__.Notify(51081)
		}
		__antithesis_instrumentation__.Notify(51060)

		setup := setupFunc(rng)
		setting := settingFunc(rng)

		allConns := make([]*gosql.DB, 0, numNodes)
		for node := 1; node <= numNodes; node++ {
			__antithesis_instrumentation__.Notify(51082)
			allConns = append(allConns, c.Conn(ctx, t.L(), node))
		}
		__antithesis_instrumentation__.Notify(51061)
		conn := allConns[0]
		t.Status("executing setup")
		t.L().Printf("setup:\n%s", strings.Join(setup, "\n"))
		for _, stmt := range setup {
			__antithesis_instrumentation__.Notify(51083)
			if _, err := conn.Exec(stmt); err != nil {
				__antithesis_instrumentation__.Notify(51084)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51085)
				logStmt(stmt)
			}
		}
		__antithesis_instrumentation__.Notify(51062)

		if settingName == "multi-region" {
			__antithesis_instrumentation__.Notify(51086)
			regionsSet := make(map[string]struct{})
			var region, zone string
			rows, err := conn.Query("SHOW REGIONS FROM CLUSTER")
			if err != nil {
				__antithesis_instrumentation__.Notify(51091)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51092)
			}
			__antithesis_instrumentation__.Notify(51087)
			for rows.Next() {
				__antithesis_instrumentation__.Notify(51093)
				if err := rows.Scan(&region, &zone); err != nil {
					__antithesis_instrumentation__.Notify(51095)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(51096)
				}
				__antithesis_instrumentation__.Notify(51094)
				regionsSet[region] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(51088)

			var regionList []string
			for region := range regionsSet {
				__antithesis_instrumentation__.Notify(51097)
				regionList = append(regionList, region)
			}
			__antithesis_instrumentation__.Notify(51089)

			if len(regionList) == 0 {
				__antithesis_instrumentation__.Notify(51098)
				t.Fatal(errors.New("no regions, cannot run multi-region config"))
			} else {
				__antithesis_instrumentation__.Notify(51099)
			}
			__antithesis_instrumentation__.Notify(51090)

			if _, err := conn.Exec(
				fmt.Sprintf(`ALTER DATABASE defaultdb SET PRIMARY REGION "%s";
ALTER TABLE seed_mr_table SET LOCALITY REGIONAL BY ROW;
INSERT INTO seed_mr_table DEFAULT VALUES;`, regionList[0]),
			); err != nil {
				__antithesis_instrumentation__.Notify(51100)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(51101)
			}
		} else {
			__antithesis_instrumentation__.Notify(51102)
		}
		__antithesis_instrumentation__.Notify(51063)

		const timeout = time.Minute
		setStmtTimeout := fmt.Sprintf("SET statement_timeout='%s';", timeout.String())
		t.Status("setting statement_timeout")
		t.L().Printf("statement timeout:\n%s", setStmtTimeout)
		if _, err := conn.Exec(setStmtTimeout); err != nil {
			__antithesis_instrumentation__.Notify(51103)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51104)
		}
		__antithesis_instrumentation__.Notify(51064)
		logStmt(setStmtTimeout)

		smither, err := sqlsmith.NewSmither(conn, rng, setting.Options...)
		if err != nil {
			__antithesis_instrumentation__.Notify(51105)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51106)
		}
		__antithesis_instrumentation__.Notify(51065)

		injectPanicsStmt := "SET testing_vectorize_inject_panics=true;"
		if _, err := conn.Exec(injectPanicsStmt); err != nil {
			__antithesis_instrumentation__.Notify(51107)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(51108)
		}
		__antithesis_instrumentation__.Notify(51066)
		logStmt(injectPanicsStmt)

		t.Status("smithing")
		until := time.After(t.Spec().(*registry.TestSpec).Timeout / 2)
		done := ctx.Done()
		for i := 1; ; i++ {
			__antithesis_instrumentation__.Notify(51109)
			if i%10000 == 0 {
				__antithesis_instrumentation__.Notify(51114)
				t.Status("smithing: ", i, " statements completed")
			} else {
				__antithesis_instrumentation__.Notify(51115)
			}
			__antithesis_instrumentation__.Notify(51110)
			select {
			case <-done:
				__antithesis_instrumentation__.Notify(51116)
				return
			case <-until:
				__antithesis_instrumentation__.Notify(51117)
				return
			default:
				__antithesis_instrumentation__.Notify(51118)
			}
			__antithesis_instrumentation__.Notify(51111)

			stmt := ""
			err := func() error {
				__antithesis_instrumentation__.Notify(51119)
				done := make(chan error, 1)
				go func(context.Context) {
					__antithesis_instrumentation__.Notify(51121)

					defer func() {
						__antithesis_instrumentation__.Notify(51125)
						if r := recover(); r != nil {
							__antithesis_instrumentation__.Notify(51126)
							done <- errors.Newf("Caught error %s", r)
							return
						} else {
							__antithesis_instrumentation__.Notify(51127)
						}
					}()
					__antithesis_instrumentation__.Notify(51122)

					stmt = smither.Generate()
					if stmt == "" {
						__antithesis_instrumentation__.Notify(51128)

						done <- errors.Newf("Empty statement returned by generate")
						return
					} else {
						__antithesis_instrumentation__.Notify(51129)
					}
					__antithesis_instrumentation__.Notify(51123)

					_, err := conn.Exec(stmt)
					if err == nil {
						__antithesis_instrumentation__.Notify(51130)
						logStmt(stmt)
					} else {
						__antithesis_instrumentation__.Notify(51131)
					}
					__antithesis_instrumentation__.Notify(51124)
					done <- err
				}(ctx)
				__antithesis_instrumentation__.Notify(51120)
				select {
				case <-time.After(timeout * 2):
					__antithesis_instrumentation__.Notify(51132)

					t.L().Printf("query timed out, but did not cancel execution:\n%s;", stmt)
					return nil
				case err := <-done:
					__antithesis_instrumentation__.Notify(51133)
					return err
				}
			}()
			__antithesis_instrumentation__.Notify(51112)
			if err != nil {
				__antithesis_instrumentation__.Notify(51134)
				es := err.Error()
				if strings.Contains(es, "internal error") {
					__antithesis_instrumentation__.Notify(51135)

					var expectedError bool
					for _, exp := range []string{
						"could not parse \"0E-2019\" as type decimal",
					} {
						__antithesis_instrumentation__.Notify(51137)
						expectedError = expectedError || func() bool {
							__antithesis_instrumentation__.Notify(51138)
							return strings.Contains(es, exp) == true
						}() == true
					}
					__antithesis_instrumentation__.Notify(51136)
					if !expectedError {
						__antithesis_instrumentation__.Notify(51139)
						logStmt(stmt)
						t.Fatalf("error: %s\nstmt:\n%s;", err, stmt)
					} else {
						__antithesis_instrumentation__.Notify(51140)
					}
				} else {
					__antithesis_instrumentation__.Notify(51141)
					if strings.Contains(es, "Empty statement returned by generate") || func() bool {
						__antithesis_instrumentation__.Notify(51142)
						return stmt == "" == true
					}() == true {
						__antithesis_instrumentation__.Notify(51143)

						t.Fatalf("Failed generating a query %s", err)
					} else {
						__antithesis_instrumentation__.Notify(51144)
					}
				}

			} else {
				__antithesis_instrumentation__.Notify(51145)
			}
			__antithesis_instrumentation__.Notify(51113)

			for idx, c := range allConns {
				__antithesis_instrumentation__.Notify(51146)
				if err := c.PingContext(ctx); err != nil {
					__antithesis_instrumentation__.Notify(51147)
					logStmt(stmt)
					t.Fatalf("ping node %d: %v\nprevious sql:\n%s;", idx+1, err, stmt)
				} else {
					__antithesis_instrumentation__.Notify(51148)
				}
			}
		}
	}
	__antithesis_instrumentation__.Notify(51048)

	register := func(setup, setting string) {
		__antithesis_instrumentation__.Notify(51149)
		var clusterSpec spec.ClusterSpec
		if strings.Contains(setting, "multi-region") {
			__antithesis_instrumentation__.Notify(51151)
			clusterSpec = r.MakeClusterSpec(numNodes, spec.Geo())
		} else {
			__antithesis_instrumentation__.Notify(51152)
			clusterSpec = r.MakeClusterSpec(numNodes)
		}
		__antithesis_instrumentation__.Notify(51150)
		r.Add(registry.TestSpec{
			Name: fmt.Sprintf("sqlsmith/setup=%s/setting=%s", setup, setting),

			Owner:   registry.OwnerSQLQueries,
			Cluster: clusterSpec,
			Timeout: time.Minute * 20,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(51153)
				runSQLSmith(ctx, t, c, setup, setting)
			},
		})
	}
	__antithesis_instrumentation__.Notify(51049)

	for setup := range setups {
		__antithesis_instrumentation__.Notify(51154)
		for setting := range settings {
			__antithesis_instrumentation__.Notify(51155)
			register(setup, setting)
		}
	}
	__antithesis_instrumentation__.Notify(51050)
	setups["seed-multi-region"] = sqlsmith.Setups["seed-multi-region"]
	settings["ddl-nodrop"] = sqlsmith.Settings["ddl-nodrop"]
	settings["multi-region"] = sqlsmith.Settings["multi-region"]
	register("tpcc", "ddl-nodrop")
	register("seed-multi-region", "multi-region")
}
