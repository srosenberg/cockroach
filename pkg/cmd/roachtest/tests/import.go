package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/errors"
)

func readCreateTableFromFixture(fixtureURI string, gatewayDB *gosql.DB) (string, error) {
	__antithesis_instrumentation__.Notify(48434)
	row := make([]byte, 0)
	err := gatewayDB.QueryRow(fmt.Sprintf(`SELECT crdb_internal.read_file('%s')`, fixtureURI)).Scan(&row)
	if err != nil {
		__antithesis_instrumentation__.Notify(48436)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(48437)
	}
	__antithesis_instrumentation__.Notify(48435)
	return string(row), err
}

func registerImportNodeShutdown(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48438)
	getImportRunner := func(ctx context.Context, t test.Test, gatewayNode int) jobStarter {
		__antithesis_instrumentation__.Notify(48441)
		startImport := func(c cluster.Cluster, t test.Test) (jobID string, err error) {
			__antithesis_instrumentation__.Notify(48443)

			tableName := "partsupp"
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(48447)

				tableName = "part"
			} else {
				__antithesis_instrumentation__.Notify(48448)
			}
			__antithesis_instrumentation__.Notify(48444)
			importStmt := fmt.Sprintf(`
				IMPORT INTO %[1]s
				CSV DATA (
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.1?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.2?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.3?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.4?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.5?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.6?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.7?AUTH=implicit',
				'gs://cockroach-fixtures/tpch-csv/sf-100/%[1]s.tbl.8?AUTH=implicit'
				) WITH  delimiter='|', detached
			`, tableName)
			gatewayDB := c.Conn(ctx, t.L(), gatewayNode)
			defer gatewayDB.Close()

			createStmt, err := readCreateTableFromFixture(
				fmt.Sprintf("gs://cockroach-fixtures/tpch-csv/schema/%s.sql?AUTH=implicit", tableName), gatewayDB)
			if err != nil {
				__antithesis_instrumentation__.Notify(48449)
				return "", err
			} else {
				__antithesis_instrumentation__.Notify(48450)
			}
			__antithesis_instrumentation__.Notify(48445)

			if _, err = gatewayDB.ExecContext(ctx, createStmt); err != nil {
				__antithesis_instrumentation__.Notify(48451)
				return jobID, err
			} else {
				__antithesis_instrumentation__.Notify(48452)
			}
			__antithesis_instrumentation__.Notify(48446)

			err = gatewayDB.QueryRowContext(ctx, importStmt).Scan(&jobID)
			return
		}
		__antithesis_instrumentation__.Notify(48442)

		return startImport
	}
	__antithesis_instrumentation__.Notify(48439)

	r.Add(registry.TestSpec{
		Name:    "import/nodeShutdown/worker",
		Owner:   registry.OwnerBulkIO,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48453)
			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
			gatewayNode := 2
			nodeToShutdown := 3
			startImport := getImportRunner(ctx, t, gatewayNode)

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, startImport)
		},
	})
	__antithesis_instrumentation__.Notify(48440)
	r.Add(registry.TestSpec{
		Name:    "import/nodeShutdown/coordinator",
		Owner:   registry.OwnerBulkIO,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48454)
			c.Put(ctx, t.Cockroach(), "./cockroach")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
			gatewayNode := 2
			nodeToShutdown := 2
			startImport := getImportRunner(ctx, t, gatewayNode)

			jobSurvivesNodeShutdown(ctx, t, c, nodeToShutdown, startImport)
		},
	})
}

func registerImportTPCC(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48455)
	runImportTPCC := func(ctx context.Context, t test.Test, c cluster.Cluster, testName string,
		timeout time.Duration, warehouses int) {
		__antithesis_instrumentation__.Notify(48458)
		c.Put(ctx, t.Cockroach(), "./cockroach")
		c.Put(ctx, t.DeprecatedWorkload(), "./workload")
		t.Status("starting csv servers")
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
		c.Run(ctx, c.All(), `./workload csv-server --port=8081 &> logs/workload-csv-server.log < /dev/null &`)

		t.Status("running workload")
		m := c.NewMonitor(ctx)
		dul := NewDiskUsageLogger(t, c)
		m.Go(dul.Runner)
		hc := NewHealthChecker(t, c, c.All())
		m.Go(hc.Runner)

		tick, perfBuf := initBulkJobPerfArtifacts(testName, timeout)
		workloadStr := `./cockroach workload fixtures import tpcc --warehouses=%d --csv-server='http://localhost:8081'`
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(48460)
			defer dul.Done()
			defer hc.Done()
			cmd := fmt.Sprintf(workloadStr, warehouses)

			tick()
			c.Run(ctx, c.Node(1), cmd)
			tick()

			dest := filepath.Join(t.PerfArtifactsDir(), "stats.json")
			if err := c.RunE(ctx, c.Node(1), "mkdir -p "+filepath.Dir(dest)); err != nil {
				__antithesis_instrumentation__.Notify(48463)
				log.Errorf(ctx, "failed to create perf dir: %+v", err)
			} else {
				__antithesis_instrumentation__.Notify(48464)
			}
			__antithesis_instrumentation__.Notify(48461)
			if err := c.PutString(ctx, perfBuf.String(), dest, 0755, c.Node(1)); err != nil {
				__antithesis_instrumentation__.Notify(48465)
				log.Errorf(ctx, "failed to upload perf artifacts to node: %s", err.Error())
			} else {
				__antithesis_instrumentation__.Notify(48466)
			}
			__antithesis_instrumentation__.Notify(48462)
			return nil
		})
		__antithesis_instrumentation__.Notify(48459)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(48456)

	const warehouses = 1000
	for _, numNodes := range []int{4, 32} {
		__antithesis_instrumentation__.Notify(48467)
		testName := fmt.Sprintf("import/tpcc/warehouses=%d/nodes=%d", warehouses, numNodes)
		timeout := 5 * time.Hour
		r.Add(registry.TestSpec{
			Name:            testName,
			Owner:           registry.OwnerBulkIO,
			Cluster:         r.MakeClusterSpec(numNodes),
			Timeout:         timeout,
			EncryptAtRandom: true,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(48468)
				runImportTPCC(ctx, t, c, testName, timeout, warehouses)
			},
		})
	}
	__antithesis_instrumentation__.Notify(48457)
	const geoWarehouses = 4000
	const geoZones = "europe-west2-b,europe-west4-b,asia-northeast1-b,us-west1-b"
	r.Add(registry.TestSpec{
		Name:            fmt.Sprintf("import/tpcc/warehouses=%d/geo", geoWarehouses),
		Owner:           registry.OwnerBulkIO,
		Cluster:         r.MakeClusterSpec(8, spec.CPU(16), spec.Geo(), spec.Zones(geoZones)),
		Timeout:         5 * time.Hour,
		EncryptAtRandom: true,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48469)
			runImportTPCC(ctx, t, c, fmt.Sprintf("import/tpcc/warehouses=%d/geo", geoWarehouses),
				5*time.Hour, geoWarehouses)
		},
	})
}

func registerImportTPCH(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48470)
	for _, item := range []struct {
		nodes   int
		timeout time.Duration
	}{

		{8, 10 * time.Hour},
	} {
		__antithesis_instrumentation__.Notify(48471)
		item := item
		r.Add(registry.TestSpec{
			Name:            fmt.Sprintf(`import/tpch/nodes=%d`, item.nodes),
			Owner:           registry.OwnerBulkIO,
			Cluster:         r.MakeClusterSpec(item.nodes),
			Timeout:         item.timeout,
			EncryptAtRandom: true,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(48472)
				tick, perfBuf := initBulkJobPerfArtifacts(t.Name(), item.timeout)

				c.Put(ctx, t.Cockroach(), "./cockroach")
				c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
				conn := c.Conn(ctx, t.L(), 1)
				if _, err := conn.Exec(`CREATE DATABASE csv;`); err != nil {
					__antithesis_instrumentation__.Notify(48478)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48479)
				}
				__antithesis_instrumentation__.Notify(48473)
				if _, err := conn.Exec(`USE csv;`); err != nil {
					__antithesis_instrumentation__.Notify(48480)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48481)
				}
				__antithesis_instrumentation__.Notify(48474)
				if _, err := conn.Exec(
					`SET CLUSTER SETTING kv.bulk_ingest.max_index_buffer_size = '2gb'`,
				); err != nil && func() bool {
					__antithesis_instrumentation__.Notify(48482)
					return !strings.Contains(err.Error(), "unknown cluster setting") == true
				}() == true {
					__antithesis_instrumentation__.Notify(48483)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48484)
				}
				__antithesis_instrumentation__.Notify(48475)

				if err := retry.ForDuration(time.Second*30, func() error {
					__antithesis_instrumentation__.Notify(48485)
					var nodes int
					if err := conn.
						QueryRowContext(ctx, `select count(*) from crdb_internal.gossip_liveness where updated_at > now() - interval '8s'`).
						Scan(&nodes); err != nil {
						__antithesis_instrumentation__.Notify(48487)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(48488)
						if nodes != item.nodes {
							__antithesis_instrumentation__.Notify(48489)
							return errors.Errorf("expected %d nodes, got %d", item.nodes, nodes)
						} else {
							__antithesis_instrumentation__.Notify(48490)
						}
					}
					__antithesis_instrumentation__.Notify(48486)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(48491)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48492)
				}
				__antithesis_instrumentation__.Notify(48476)
				m := c.NewMonitor(ctx)
				dul := NewDiskUsageLogger(t, c)
				m.Go(dul.Runner)
				hc := NewHealthChecker(t, c, c.All())
				m.Go(hc.Runner)

				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(48493)
					defer dul.Done()
					defer hc.Done()
					t.WorkerStatus(`running import`)
					defer t.WorkerStatus()

					createStmt, err := readCreateTableFromFixture(
						"gs://cockroach-fixtures/tpch-csv/schema/lineitem.sql?AUTH=implicit", conn)
					if err != nil {
						__antithesis_instrumentation__.Notify(48499)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48500)
					}
					__antithesis_instrumentation__.Notify(48494)

					if _, err := conn.ExecContext(ctx, createStmt); err != nil {
						__antithesis_instrumentation__.Notify(48501)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48502)
					}
					__antithesis_instrumentation__.Notify(48495)

					tick()
					_, err = conn.Exec(`
						IMPORT INTO csv.lineitem
						CSV DATA (
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.1?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.2?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.3?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.4?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.5?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.6?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.7?AUTH=implicit',
						'gs://cockroach-fixtures/tpch-csv/sf-100/lineitem.tbl.8?AUTH=implicit'
						) WITH  delimiter='|'
					`)
					if err != nil {
						__antithesis_instrumentation__.Notify(48503)
						return errors.Wrap(err, "import failed")
					} else {
						__antithesis_instrumentation__.Notify(48504)
					}
					__antithesis_instrumentation__.Notify(48496)
					tick()

					dest := filepath.Join(t.PerfArtifactsDir(), "stats.json")
					if err := c.RunE(ctx, c.Node(1), "mkdir -p "+filepath.Dir(dest)); err != nil {
						__antithesis_instrumentation__.Notify(48505)
						log.Errorf(ctx, "failed to create perf dir: %+v", err)
					} else {
						__antithesis_instrumentation__.Notify(48506)
					}
					__antithesis_instrumentation__.Notify(48497)
					if err := c.PutString(ctx, perfBuf.String(), dest, 0755, c.Node(1)); err != nil {
						__antithesis_instrumentation__.Notify(48507)
						log.Errorf(ctx, "failed to upload perf artifacts to node: %s", err.Error())
					} else {
						__antithesis_instrumentation__.Notify(48508)
					}
					__antithesis_instrumentation__.Notify(48498)
					return nil
				})
				__antithesis_instrumentation__.Notify(48477)

				t.Status("waiting")
				m.Wait()
			},
		})
	}
}

func successfulImportStep(warehouses, nodeID int) versionStep {
	__antithesis_instrumentation__.Notify(48509)
	return func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
		__antithesis_instrumentation__.Notify(48510)
		u.c.Run(ctx, u.c.Node(nodeID), tpccImportCmd(warehouses))
	}
}

func runImportMixedVersion(
	ctx context.Context, t test.Test, c cluster.Cluster, warehouses int, predecessorVersion string,
) {
	__antithesis_instrumentation__.Notify(48511)

	const mainVersion = ""
	roachNodes := c.All()

	t.Status("starting csv servers")

	u := newVersionUpgradeTest(c,
		uploadAndStartFromCheckpointFixture(roachNodes, predecessorVersion),
		waitForUpgradeStep(roachNodes),
		preventAutoUpgradeStep(1),

		binaryUpgradeStep(c.Node(1), mainVersion),
		binaryUpgradeStep(c.Node(2), mainVersion),

		successfulImportStep(warehouses, 1),
	)
	u.run(ctx, t)
}

func registerImportMixedVersion(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48512)
	r.Add(registry.TestSpec{
		Name:  "import/mixed-versions",
		Owner: registry.OwnerBulkIO,

		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48513)
			predV, err := PredecessorVersion(*t.BuildVersion())
			if err != nil {
				__antithesis_instrumentation__.Notify(48516)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48517)
			}
			__antithesis_instrumentation__.Notify(48514)
			warehouses := 100
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(48518)
				warehouses = 10
			} else {
				__antithesis_instrumentation__.Notify(48519)
			}
			__antithesis_instrumentation__.Notify(48515)
			runImportMixedVersion(ctx, t, c, warehouses, predV)
		},
	})
}

func registerImportDecommissioned(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48520)
	runImportDecommissioned := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(48522)
		warehouses := 100
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(48524)
			warehouses = 10
		} else {
			__antithesis_instrumentation__.Notify(48525)
		}
		__antithesis_instrumentation__.Notify(48523)

		c.Put(ctx, t.Cockroach(), "./cockroach")
		c.Put(ctx, t.DeprecatedWorkload(), "./workload")
		t.Status("starting csv servers")
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())
		c.Run(ctx, c.All(), `./workload csv-server --port=8081 &> logs/workload-csv-server.log < /dev/null &`)

		nodeToDecommission := 2
		t.Status(fmt.Sprintf("decommissioning node %d", nodeToDecommission))
		c.Run(ctx, c.Node(nodeToDecommission), `./cockroach node decommission --insecure --self --wait=all`)

		time.Sleep(10 * time.Second)

		t.Status("running workload")
		c.Run(ctx, c.Node(1), tpccImportCmd(warehouses))
	}
	__antithesis_instrumentation__.Notify(48521)

	r.Add(registry.TestSpec{
		Name:    "import/decommissioned",
		Owner:   registry.OwnerBulkIO,
		Cluster: r.MakeClusterSpec(4),
		Run:     runImportDecommissioned,
	})
}
