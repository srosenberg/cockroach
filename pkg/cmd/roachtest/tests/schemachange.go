package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/roachprod/logger"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

func registerSchemaChangeDuringKV(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50575)
	r.Add(registry.TestSpec{
		Name:    `schemachange/during/kv`,
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50576)
			const fixturePath = `gs://cockroach-fixtures/workload/tpch/scalefactor=10/backup?AUTH=implicit`

			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Put(ctx, t.DeprecatedWorkload(), "./workload")

			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())
			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			m := c.NewMonitor(ctx, c.All())
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50580)
				t.Status("loading fixture")
				if _, err := db.Exec(`RESTORE DATABASE tpch FROM $1`, fixturePath); err != nil {
					__antithesis_instrumentation__.Notify(50582)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(50583)
				}
				__antithesis_instrumentation__.Notify(50581)
				return nil
			})
			__antithesis_instrumentation__.Notify(50577)
			m.Wait()

			c.Run(ctx, c.Node(1), `./workload init kv --drop --db=test`)
			for node := 1; node <= c.Spec().NodeCount; node++ {
				__antithesis_instrumentation__.Notify(50584)
				node := node

				go func() {
					__antithesis_instrumentation__.Notify(50585)
					const cmd = `./workload run kv --tolerate-errors --min-block-bytes=8 --max-block-bytes=127 --db=test`
					l, err := t.L().ChildLogger(fmt.Sprintf(`kv-%d`, node))
					if err != nil {
						__antithesis_instrumentation__.Notify(50587)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(50588)
					}
					__antithesis_instrumentation__.Notify(50586)
					defer l.Close()
					_ = c.RunE(ctx, c.Node(node), cmd)
				}()
			}
			__antithesis_instrumentation__.Notify(50578)

			m = c.NewMonitor(ctx, c.All())
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50589)
				t.Status("running schema change tests")
				return waitForSchemaChanges(ctx, t.L(), db)
			})
			__antithesis_instrumentation__.Notify(50579)
			m.Wait()
		},
	})
}

func waitForSchemaChanges(ctx context.Context, l *logger.Logger, db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(50590)
	start := timeutil.Now()

	l.Printf("running schema changes over tpch.customer\n")
	schemaChanges := []string{
		"ALTER TABLE tpch.customer ADD COLUMN newcol INT DEFAULT 23456",
		"CREATE INDEX foo ON tpch.customer (c_name)",
	}
	if err := runSchemaChanges(ctx, l, db, schemaChanges); err != nil {
		__antithesis_instrumentation__.Notify(50594)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50595)
	}
	__antithesis_instrumentation__.Notify(50591)

	validationQueries := []string{
		"SELECT count(*) FROM tpch.customer AS OF SYSTEM TIME %s",
		"SELECT count(newcol) FROM tpch.customer AS OF SYSTEM TIME %s",
		"SELECT count(c_name) FROM tpch.customer@foo AS OF SYSTEM TIME %s",
	}
	if err := runValidationQueries(ctx, l, db, start, validationQueries, nil); err != nil {
		__antithesis_instrumentation__.Notify(50596)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50597)
	}
	__antithesis_instrumentation__.Notify(50592)

	l.Printf("running schema changes over test.kv\n")
	schemaChanges = []string{
		"ALTER TABLE test.kv ADD COLUMN created_at TIMESTAMP DEFAULT now()",
		"CREATE INDEX foo ON test.kv (v)",
	}
	if err := runSchemaChanges(ctx, l, db, schemaChanges); err != nil {
		__antithesis_instrumentation__.Notify(50598)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50599)
	}
	__antithesis_instrumentation__.Notify(50593)

	validationQueries = []string{
		"SELECT count(*) FROM test.kv AS OF SYSTEM TIME %s",
		"SELECT count(v) FROM test.kv AS OF SYSTEM TIME %s",
		"SELECT count(v) FROM test.kv@foo AS OF SYSTEM TIME %s",
	}

	indexValidationQueries := []string{
		"SELECT count(k) FROM test.kv@kv_pkey AS OF SYSTEM TIME %s WHERE created_at > $1 AND created_at <= $2",
		"SELECT count(v) FROM test.kv@foo AS OF SYSTEM TIME %s WHERE created_at > $1 AND created_at <= $2",
	}
	return runValidationQueries(ctx, l, db, start, validationQueries, indexValidationQueries)
}

func runSchemaChanges(
	ctx context.Context, l *logger.Logger, db *gosql.DB, schemaChanges []string,
) error {
	__antithesis_instrumentation__.Notify(50600)
	for _, cmd := range schemaChanges {
		__antithesis_instrumentation__.Notify(50602)
		start := timeutil.Now()
		l.Printf("starting schema change: %s\n", cmd)
		if _, err := db.Exec(cmd); err != nil {
			__antithesis_instrumentation__.Notify(50604)
			l.Errorf("hit schema change error: %s, for %s, in %s\n", err, cmd, timeutil.Since(start))
			return err
		} else {
			__antithesis_instrumentation__.Notify(50605)
		}
		__antithesis_instrumentation__.Notify(50603)
		l.Printf("completed schema change: %s, in %s\n", cmd, timeutil.Since(start))

	}
	__antithesis_instrumentation__.Notify(50601)

	return nil
}

func runValidationQueries(
	ctx context.Context,
	l *logger.Logger,
	db *gosql.DB,
	start time.Time,
	validationQueries []string,
	indexValidationQueries []string,
) error {
	__antithesis_instrumentation__.Notify(50606)

	time.Sleep(5 * time.Second)

	var nowString string
	if err := db.QueryRow("SELECT cluster_logical_timestamp()").Scan(&nowString); err != nil {
		__antithesis_instrumentation__.Notify(50610)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50611)
	}
	__antithesis_instrumentation__.Notify(50607)
	var nowInNanos int64
	if _, err := fmt.Sscanf(nowString, "%d", &nowInNanos); err != nil {
		__antithesis_instrumentation__.Notify(50612)
		return err
	} else {
		__antithesis_instrumentation__.Notify(50613)
	}
	__antithesis_instrumentation__.Notify(50608)
	now := timeutil.Unix(0, nowInNanos)

	var eCount int64
	for i := range validationQueries {
		__antithesis_instrumentation__.Notify(50614)
		var count int64
		q := fmt.Sprintf(validationQueries[i], nowString)
		if err := db.QueryRow(q).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(50617)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50618)
		}
		__antithesis_instrumentation__.Notify(50615)
		l.Printf("query: %s, found %d rows\n", q, count)
		if count == 0 {
			__antithesis_instrumentation__.Notify(50619)
			return errors.Errorf("%s: %d rows found", q, count)
		} else {
			__antithesis_instrumentation__.Notify(50620)
		}
		__antithesis_instrumentation__.Notify(50616)
		if eCount == 0 {
			__antithesis_instrumentation__.Notify(50621)
			eCount = count

			if indexValidationQueries != nil {
				__antithesis_instrumentation__.Notify(50622)
				sp := timeSpan{start: start, end: now}
				if err := findIndexProblem(
					ctx, l, db, sp, nowString, indexValidationQueries,
				); err != nil {
					__antithesis_instrumentation__.Notify(50623)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50624)
				}
			} else {
				__antithesis_instrumentation__.Notify(50625)
			}
		} else {
			__antithesis_instrumentation__.Notify(50626)
			if count != eCount {
				__antithesis_instrumentation__.Notify(50627)
				return errors.Errorf("%s: %d rows found, expected %d rows", q, count, eCount)
			} else {
				__antithesis_instrumentation__.Notify(50628)
			}
		}
	}
	__antithesis_instrumentation__.Notify(50609)
	return nil
}

type timeSpan struct {
	start, end time.Time
}

func checkIndexOverTimeSpan(
	ctx context.Context,
	l *logger.Logger,
	db *gosql.DB,
	s timeSpan,
	nowString string,
	indexValidationQueries []string,
) (bool, error) {
	__antithesis_instrumentation__.Notify(50629)
	var eCount int64
	q := fmt.Sprintf(indexValidationQueries[0], nowString)
	if err := db.QueryRow(q, s.start, s.end).Scan(&eCount); err != nil {
		__antithesis_instrumentation__.Notify(50632)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(50633)
	}
	__antithesis_instrumentation__.Notify(50630)
	var count int64
	q = fmt.Sprintf(indexValidationQueries[1], nowString)
	if err := db.QueryRow(q, s.start, s.end).Scan(&count); err != nil {
		__antithesis_instrumentation__.Notify(50634)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(50635)
	}
	__antithesis_instrumentation__.Notify(50631)
	l.Printf("counts seen %d, %d, over [%s, %s]\n", count, eCount, s.start, s.end)
	return count != eCount, nil
}

func findIndexProblem(
	ctx context.Context,
	l *logger.Logger,
	db *gosql.DB,
	s timeSpan,
	nowString string,
	indexValidationQueries []string,
) error {
	__antithesis_instrumentation__.Notify(50636)
	spans := []timeSpan{s}

	for len(spans) > 0 {
		__antithesis_instrumentation__.Notify(50638)
		s := spans[0]
		spans = spans[1:]

		leftSpan, rightSpan := s, s
		d := s.end.Sub(s.start) / 2
		if d < 50*time.Millisecond {
			__antithesis_instrumentation__.Notify(50644)
			l.Printf("problem seen over [%s, %s]\n", s.start, s.end)
			continue
		} else {
			__antithesis_instrumentation__.Notify(50645)
		}
		__antithesis_instrumentation__.Notify(50639)
		m := s.start.Add(d)
		leftSpan.end = m
		rightSpan.start = m

		leftState, err := checkIndexOverTimeSpan(
			ctx, l, db, leftSpan, nowString, indexValidationQueries)
		if err != nil {
			__antithesis_instrumentation__.Notify(50646)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50647)
		}
		__antithesis_instrumentation__.Notify(50640)
		rightState, err := checkIndexOverTimeSpan(
			ctx, l, db, rightSpan, nowString, indexValidationQueries)
		if err != nil {
			__antithesis_instrumentation__.Notify(50648)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50649)
		}
		__antithesis_instrumentation__.Notify(50641)
		if leftState {
			__antithesis_instrumentation__.Notify(50650)
			spans = append(spans, leftSpan)
		} else {
			__antithesis_instrumentation__.Notify(50651)
		}
		__antithesis_instrumentation__.Notify(50642)
		if rightState {
			__antithesis_instrumentation__.Notify(50652)
			spans = append(spans, rightSpan)
		} else {
			__antithesis_instrumentation__.Notify(50653)
		}
		__antithesis_instrumentation__.Notify(50643)
		if !(leftState || func() bool {
			__antithesis_instrumentation__.Notify(50654)
			return rightState == true
		}() == true) {
			__antithesis_instrumentation__.Notify(50655)
			l.Printf("no problem seen over [%s, %s]\n", s.start, s.end)
		} else {
			__antithesis_instrumentation__.Notify(50656)
		}
	}
	__antithesis_instrumentation__.Notify(50637)
	return nil
}

func registerSchemaChangeIndexTPCC1000(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50657)
	r.Add(makeIndexAddTpccTest(r.MakeClusterSpec(5, spec.CPU(16)), 1000, time.Hour*2))
}

func registerSchemaChangeIndexTPCC100(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50658)
	r.Add(makeIndexAddTpccTest(r.MakeClusterSpec(5), 100, time.Minute*15))
}

func makeIndexAddTpccTest(
	spec spec.ClusterSpec, warehouses int, length time.Duration,
) registry.TestSpec {
	__antithesis_instrumentation__.Notify(50659)
	return registry.TestSpec{
		Name:    fmt.Sprintf("schemachange/index/tpcc/w=%d", warehouses),
		Owner:   registry.OwnerSQLSchema,
		Cluster: spec,
		Timeout: length * 3,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50660)
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses: warehouses,

				ExtraRunArgs: fmt.Sprintf("--wait=false --tolerate-errors --workers=%d", warehouses),
				During: func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(50661)
					return runAndLogStmts(ctx, t, c, "addindex", []string{
						`CREATE UNIQUE INDEX ON tpcc.order (o_entry_d, o_w_id, o_d_id, o_carrier_id, o_id);`,
						`CREATE INDEX ON tpcc.order (o_carrier_id);`,
						`CREATE INDEX ON tpcc.customer (c_last, c_first);`,
					})
				},
				Duration:  length,
				SetupType: usingImport,
			})
		},
	}
}

func registerSchemaChangeBulkIngest(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50662)
	r.Add(makeSchemaChangeBulkIngestTest(r, 5, 100000000, time.Minute*20))
}

func makeSchemaChangeBulkIngestTest(
	r registry.Registry, numNodes, numRows int, length time.Duration,
) registry.TestSpec {
	__antithesis_instrumentation__.Notify(50663)
	return registry.TestSpec{
		Name:    "schemachange/bulkingest",
		Owner:   registry.OwnerSQLSchema,
		Cluster: r.MakeClusterSpec(numNodes),
		Timeout: length * 2,

		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50664)

			aNum := numRows
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(50669)
				aNum = 100000
			} else {
				__antithesis_instrumentation__.Notify(50670)
			}
			__antithesis_instrumentation__.Notify(50665)
			bNum := 1
			cNum := 1
			payloadBytes := 4

			crdbNodes := c.Range(1, c.Spec().NodeCount-1)
			workloadNode := c.Node(c.Spec().NodeCount)

			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", workloadNode)

			settings := install.MakeClusterSettings(install.EnvOption([]string{"COCKROACH_IMPORT_WORKLOAD_FASTER=true"}))
			c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, crdbNodes)

			cmdWrite := fmt.Sprintf(

				"./cockroach workload fixtures import bulkingest {pgurl:1} --a %d --b %d --c %d --payload-bytes %d --index-b-c-a=false",
				aNum, bNum, cNum, payloadBytes,
			)

			c.Run(ctx, workloadNode, cmdWrite)

			m := c.NewMonitor(ctx, crdbNodes)

			indexDuration := length
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(50671)
				indexDuration = time.Second * 30
			} else {
				__antithesis_instrumentation__.Notify(50672)
			}
			__antithesis_instrumentation__.Notify(50666)
			cmdWriteAndRead := fmt.Sprintf(
				"./workload run bulkingest --duration %s {pgurl:1-%d} --a %d --b %d --c %d --payload-bytes %d",
				indexDuration.String(), c.Spec().NodeCount-1, aNum, bNum, cNum, payloadBytes,
			)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50673)
				c.Run(ctx, workloadNode, cmdWriteAndRead)
				return nil
			})
			__antithesis_instrumentation__.Notify(50667)

			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50674)
				db := c.Conn(ctx, t.L(), 1)
				defer db.Close()

				if !c.IsLocal() {
					__antithesis_instrumentation__.Notify(50677)

					sleepInterval := time.Minute * 5
					maxSleep := length / 2
					if sleepInterval > maxSleep {
						__antithesis_instrumentation__.Notify(50679)
						sleepInterval = maxSleep
					} else {
						__antithesis_instrumentation__.Notify(50680)
					}
					__antithesis_instrumentation__.Notify(50678)
					time.Sleep(sleepInterval)
				} else {
					__antithesis_instrumentation__.Notify(50681)
				}
				__antithesis_instrumentation__.Notify(50675)

				t.L().Printf("Creating index")
				before := timeutil.Now()
				if _, err := db.Exec(`CREATE INDEX payload_a ON bulkingest.bulkingest (payload, a)`); err != nil {
					__antithesis_instrumentation__.Notify(50682)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(50683)
				}
				__antithesis_instrumentation__.Notify(50676)
				t.L().Printf("CREATE INDEX took %v\n", timeutil.Since(before))
				return nil
			})
			__antithesis_instrumentation__.Notify(50668)

			m.Wait()
		},
	}
}

func registerSchemaChangeDuringTPCC1000(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50684)
	r.Add(makeSchemaChangeDuringTPCC(r.MakeClusterSpec(5, spec.CPU(16)), 1000, time.Hour*3))
}

func makeSchemaChangeDuringTPCC(
	spec spec.ClusterSpec, warehouses int, length time.Duration,
) registry.TestSpec {
	__antithesis_instrumentation__.Notify(50685)
	return registry.TestSpec{
		Name:    "schemachange/during/tpcc",
		Owner:   registry.OwnerSQLSchema,
		Cluster: spec,
		Timeout: length * 3,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50686)
			runTPCC(ctx, t, c, tpccOptions{
				Warehouses: warehouses,

				ExtraRunArgs: fmt.Sprintf("--wait=false --tolerate-errors --workers=%d", warehouses),
				During: func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(50687)
					if t.IsBuildVersion(`v19.2.0`) {
						__antithesis_instrumentation__.Notify(50689)
						if err := runAndLogStmts(ctx, t, c, "during-schema-changes-19.2", []string{

							`CREATE TABLE tpcc.orderpks (o_w_id, o_d_id, o_id, PRIMARY KEY(o_w_id, o_d_id, o_id)) AS select o_w_id, o_d_id, o_id FROM tpcc.order;`,
						}); err != nil {
							__antithesis_instrumentation__.Notify(50690)
							return err
						} else {
							__antithesis_instrumentation__.Notify(50691)
						}
					} else {
						__antithesis_instrumentation__.Notify(50692)
						if err := runAndLogStmts(ctx, t, c, "during-schema-changes-19.1", []string{
							`CREATE TABLE tpcc.orderpks (o_w_id INT, o_d_id INT, o_id INT, PRIMARY KEY(o_w_id, o_d_id, o_id));`,

							`INSERT INTO tpcc.orderpks SELECT o_w_id, o_d_id, o_id FROM tpcc.order LIMIT 10000;`,
						}); err != nil {
							__antithesis_instrumentation__.Notify(50693)
							return err
						} else {
							__antithesis_instrumentation__.Notify(50694)
						}
					}
					__antithesis_instrumentation__.Notify(50688)
					return runAndLogStmts(ctx, t, c, "during-schema-changes", []string{
						`CREATE INDEX ON tpcc.order (o_carrier_id);`,

						`CREATE TABLE tpcc.customerpks (c_w_id INT, c_d_id INT, c_id INT, FOREIGN KEY (c_w_id, c_d_id, c_id) REFERENCES tpcc.customer (c_w_id, c_d_id, c_id));`,

						`ALTER TABLE tpcc.order ADD COLUMN orderdiscount INT DEFAULT 0;`,
						`ALTER TABLE tpcc.order ADD CONSTRAINT nodiscount CHECK (orderdiscount = 0);`,

						`ALTER TABLE tpcc.orderpks ADD CONSTRAINT warehouse_id FOREIGN KEY (o_w_id) REFERENCES tpcc.warehouse (w_id);`,

						`ALTER TABLE tpcc.district VALIDATE CONSTRAINT district_d_w_id_fkey;`,

						`ALTER TABLE tpcc.orderpks RENAME TO tpcc.readytodrop;`,
						`TRUNCATE TABLE tpcc.readytodrop CASCADE;`,
						`DROP TABLE tpcc.readytodrop CASCADE;`,
					})
				},
				Duration:  length,
				SetupType: usingImport,
			})
		},
	}
}

func runAndLogStmts(
	ctx context.Context, t test.Test, c cluster.Cluster, prefix string, stmts []string,
) error {
	__antithesis_instrumentation__.Notify(50695)
	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()
	t.L().Printf("%s: running %d statements\n", prefix, len(stmts))
	start := timeutil.Now()
	for i, stmt := range stmts {
		__antithesis_instrumentation__.Notify(50697)

		time.Sleep(time.Minute)
		t.L().Printf("%s: running statement %d...\n", prefix, i+1)
		before := timeutil.Now()
		if _, err := db.Exec(stmt); err != nil {
			__antithesis_instrumentation__.Notify(50699)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50700)
		}
		__antithesis_instrumentation__.Notify(50698)
		t.L().Printf("%s: statement %d: %q took %v\n", prefix, i+1, stmt, timeutil.Since(before))
	}
	__antithesis_instrumentation__.Notify(50696)
	t.L().Printf("%s: ran %d statements in %v\n", prefix, len(stmts), timeutil.Since(start))
	return nil
}
