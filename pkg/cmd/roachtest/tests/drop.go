package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func registerDrop(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47432)

	runDrop := func(ctx context.Context, t test.Test, c cluster.Cluster, warehouses, nodes int) {
		__antithesis_instrumentation__.Notify(47434)
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Range(1, nodes))
		settings := install.MakeClusterSettings()
		settings.Env = append(settings.Env, "COCKROACH_MEMPROF_INTERVAL=15s")
		c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Range(1, nodes))

		m := c.NewMonitor(ctx, c.Range(1, nodes))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(47436)
			t.WorkerStatus("importing TPCC fixture")
			c.Run(ctx, c.Node(1), tpccImportCmd(warehouses))

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			run := func(maybeExperimental bool, stmtStr string, args ...interface{}) {
				__antithesis_instrumentation__.Notify(47444)
				stmt := stmtStr

				if maybeExperimental {
					__antithesis_instrumentation__.Notify(47447)
					stmt = fmt.Sprintf(stmtStr, "", "=")
				} else {
					__antithesis_instrumentation__.Notify(47448)
				}
				__antithesis_instrumentation__.Notify(47445)
				t.WorkerStatus(stmt)
				_, err := db.ExecContext(ctx, stmt, args...)
				if err != nil && func() bool {
					__antithesis_instrumentation__.Notify(47449)
					return maybeExperimental == true
				}() == true && func() bool {
					__antithesis_instrumentation__.Notify(47450)
					return strings.Contains(err.Error(), "syntax error") == true
				}() == true {
					__antithesis_instrumentation__.Notify(47451)
					stmt = fmt.Sprintf(stmtStr, "EXPERIMENTAL", "")
					t.WorkerStatus(stmt)
					_, err = db.ExecContext(ctx, stmt, args...)
				} else {
					__antithesis_instrumentation__.Notify(47452)
				}
				__antithesis_instrumentation__.Notify(47446)
				if err != nil {
					__antithesis_instrumentation__.Notify(47453)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(47454)
				}
			}
			__antithesis_instrumentation__.Notify(47437)

			run(false, `SET CLUSTER SETTING trace.debug.enable = true`)

			const stmtDropConstraint = "ALTER TABLE tpcc.order_line DROP CONSTRAINT order_line_ol_supply_w_id_ol_i_id_fkey"
			run(false, stmtDropConstraint)

			var rows, minWarehouse, maxWarehouse int
			if err := db.QueryRow("select count(*), min(s_w_id), max(s_w_id) from tpcc.stock").Scan(&rows,
				&minWarehouse, &maxWarehouse); err != nil {
				__antithesis_instrumentation__.Notify(47455)
				t.Fatalf("failed to get range count: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(47456)
			}
			__antithesis_instrumentation__.Notify(47438)

			for j := 1; j <= nodes; j++ {
				__antithesis_instrumentation__.Notify(47457)
				size, err := getDiskUsageInBytes(ctx, c, t.L(), j)
				if err != nil {
					__antithesis_instrumentation__.Notify(47459)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47460)
				}
				__antithesis_instrumentation__.Notify(47458)

				t.L().Printf("Node %d space used: %s\n", j, humanizeutil.IBytes(int64(size)))
			}
			__antithesis_instrumentation__.Notify(47439)

			for i := minWarehouse; i <= maxWarehouse; i++ {
				__antithesis_instrumentation__.Notify(47461)
				t.Progress(float64(i) / float64(maxWarehouse))
				tBegin := timeutil.Now()
				run(false, "DELETE FROM tpcc.stock WHERE s_w_id = $1", i)
				elapsed := timeutil.Since(tBegin)

				t.L().Printf("deleted from tpcc.stock for warehouse %d (100k rows) in %s (%.2f rows/sec)\n", i, elapsed, 100000.0/elapsed.Seconds())
			}
			__antithesis_instrumentation__.Notify(47440)

			const stmtTruncate = "TRUNCATE TABLE tpcc.stock"
			run(false, stmtTruncate)

			const stmtDrop = "DROP DATABASE tpcc"
			run(false, stmtDrop)

			run(true, "ALTER RANGE default %[1]s CONFIGURE ZONE %[2]s '\ngc:\n  ttlseconds: 1\n'")

			var allNodesSpaceCleared bool
			var sizeReport string
			maxSizeBytes := 100 * 1024 * 1024
			if true {
				__antithesis_instrumentation__.Notify(47462)

				maxSizeBytes *= 100
			} else {
				__antithesis_instrumentation__.Notify(47463)
			}
			__antithesis_instrumentation__.Notify(47441)

			for i := 0; i < 10; i++ {
				__antithesis_instrumentation__.Notify(47464)
				sizeReport = ""
				allNodesSpaceCleared = true
				for j := 1; j <= nodes; j++ {
					__antithesis_instrumentation__.Notify(47467)
					size, err := getDiskUsageInBytes(ctx, c, t.L(), j)
					if err != nil {
						__antithesis_instrumentation__.Notify(47469)
						return err
					} else {
						__antithesis_instrumentation__.Notify(47470)
					}
					__antithesis_instrumentation__.Notify(47468)

					nodeSpaceUsed := fmt.Sprintf("Node %d space after deletion used: %s\n", j, humanizeutil.IBytes(int64(size)))
					t.L().Printf(nodeSpaceUsed)

					if size > maxSizeBytes {
						__antithesis_instrumentation__.Notify(47471)
						allNodesSpaceCleared = false
						sizeReport += nodeSpaceUsed
					} else {
						__antithesis_instrumentation__.Notify(47472)
					}
				}
				__antithesis_instrumentation__.Notify(47465)

				if allNodesSpaceCleared {
					__antithesis_instrumentation__.Notify(47473)
					break
				} else {
					__antithesis_instrumentation__.Notify(47474)
				}
				__antithesis_instrumentation__.Notify(47466)
				time.Sleep(time.Minute)
			}
			__antithesis_instrumentation__.Notify(47442)

			if !allNodesSpaceCleared {
				__antithesis_instrumentation__.Notify(47475)
				sizeReport += fmt.Sprintf("disk space usage has not dropped below %s on all nodes.",
					humanizeutil.IBytes(int64(maxSizeBytes)))
				t.Fatalf(sizeReport)
			} else {
				__antithesis_instrumentation__.Notify(47476)
			}
			__antithesis_instrumentation__.Notify(47443)

			return nil
		})
		__antithesis_instrumentation__.Notify(47435)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(47433)

	warehouses := 100
	numNodes := 9

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("drop/tpcc/w=%d,nodes=%d", warehouses, numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47477)

			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(47479)
				numNodes = 4
				warehouses = 1
				fmt.Printf("running with w=%d,nodes=%d in local mode\n", warehouses, numNodes)
			} else {
				__antithesis_instrumentation__.Notify(47480)
			}
			__antithesis_instrumentation__.Notify(47478)
			runDrop(ctx, t, c, warehouses, numNodes)
		},
	})
}
