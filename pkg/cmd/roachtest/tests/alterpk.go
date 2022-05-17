package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/randutil"
)

func registerAlterPK(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45431)

	setupTest := func(ctx context.Context, t test.Test, c cluster.Cluster) (option.NodeListOption, option.NodeListOption) {
		__antithesis_instrumentation__.Notify(45436)
		roachNodes := c.Range(1, c.Spec().NodeCount-1)
		loadNode := c.Node(c.Spec().NodeCount)
		t.Status("copying binaries")
		c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)

		t.Status("starting cockroach nodes")
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)
		return roachNodes, loadNode
	}
	__antithesis_instrumentation__.Notify(45432)

	runAlterPKBank := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(45437)
		const numRows = 1000000
		const duration = 1 * time.Minute

		roachNodes, loadNode := setupTest(ctx, t, c)

		initDone := make(chan struct{}, 1)
		pkChangeDone := make(chan struct{}, 1)

		m := c.NewMonitor(ctx, roachNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45440)

			cmd := fmt.Sprintf("./workload init bank --drop --rows %d {pgurl%s}", numRows, roachNodes)
			if err := c.RunE(ctx, loadNode, cmd); err != nil {
				__antithesis_instrumentation__.Notify(45442)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45443)
			}
			__antithesis_instrumentation__.Notify(45441)
			initDone <- struct{}{}

			cmd = fmt.Sprintf("./workload run bank --duration=%s {pgurl%s}", duration, roachNodes)
			c.Run(ctx, loadNode, cmd)

			<-pkChangeDone
			t.Status("starting second run of the workload after primary key change")

			c.Run(ctx, loadNode, cmd)
			return nil
		})
		__antithesis_instrumentation__.Notify(45438)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45444)

			<-initDone
			time.Sleep(duration / 10)

			t.Status("beginning primary key change")
			db := c.Conn(ctx, t.L(), roachNodes[0])
			defer db.Close()
			cmd := `
			USE bank;
			ALTER TABLE bank ALTER COLUMN balance SET NOT NULL;
			ALTER TABLE bank ALTER PRIMARY KEY USING COLUMNS (id, balance)
			`
			if _, err := db.ExecContext(ctx, cmd); err != nil {
				__antithesis_instrumentation__.Notify(45446)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45447)
			}
			__antithesis_instrumentation__.Notify(45445)
			t.Status("primary key change finished")
			pkChangeDone <- struct{}{}
			return nil
		})
		__antithesis_instrumentation__.Notify(45439)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(45433)

	runAlterPKTPCC := func(ctx context.Context, t test.Test, c cluster.Cluster, warehouses int, expensiveChecks bool) {
		__antithesis_instrumentation__.Notify(45448)
		const duration = 10 * time.Minute

		roachNodes, loadNode := setupTest(ctx, t, c)

		cmd := fmt.Sprintf(
			"./cockroach workload fixtures import tpcc --warehouses=%d --db=tpcc",
			warehouses,
		)
		if err := c.RunE(ctx, c.Node(roachNodes[0]), cmd); err != nil {
			__antithesis_instrumentation__.Notify(45453)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45454)
		}
		__antithesis_instrumentation__.Notify(45449)

		m := c.NewMonitor(ctx, roachNodes)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45455)

			runCmd := fmt.Sprintf(
				"./workload run tpcc --warehouses=%d --split --scatter --duration=%s {pgurl%s}",
				warehouses,
				duration,
				roachNodes,
			)
			t.Status("beginning workload")
			c.Run(ctx, loadNode, runCmd)
			t.Status("finished running workload")
			return nil
		})
		__antithesis_instrumentation__.Notify(45450)
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45456)

			time.Sleep(duration / 10)

			alterStmts := []string{
				`ALTER TABLE warehouse ALTER PRIMARY KEY USING COLUMNS (w_id)`,
				`ALTER TABLE district ALTER PRIMARY KEY USING COLUMNS (d_w_id, d_id)`,
				`ALTER TABLE history ALTER PRIMARY KEY USING COLUMNS (h_w_id, rowid)`,
				`ALTER TABLE customer ALTER PRIMARY KEY USING COLUMNS (c_w_id, c_d_id, c_id)`,
				`ALTER TABLE "order" ALTER PRIMARY KEY USING COLUMNS (o_w_id, o_d_id, o_id DESC)`,
				`ALTER TABLE new_order ALTER PRIMARY KEY USING COLUMNS (no_w_id, no_d_id, no_o_id)`,
				`ALTER TABLE item ALTER PRIMARY KEY USING COLUMNS (i_id)`,
				`ALTER TABLE stock ALTER PRIMARY KEY USING COLUMNS (s_w_id, s_i_id)`,
				`ALTER TABLE order_line ALTER PRIMARY KEY USING COLUMNS (ol_w_id, ol_d_id, ol_o_id DESC, ol_number)`,
			}

			rand, _ := randutil.NewTestRand()
			randStmt := alterStmts[rand.Intn(len(alterStmts))]
			t.Status("Running command: ", randStmt)

			db := c.Conn(ctx, t.L(), roachNodes[0])
			defer db.Close()
			alterCmd := `USE tpcc; %s;`
			t.Status("beginning primary key change")
			if _, err := db.ExecContext(ctx, fmt.Sprintf(alterCmd, randStmt)); err != nil {
				__antithesis_instrumentation__.Notify(45458)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45459)
			}
			__antithesis_instrumentation__.Notify(45457)
			t.Status("primary key change finished")
			return nil
		})
		__antithesis_instrumentation__.Notify(45451)

		m.Wait()

		expensiveChecksArg := ""
		if expensiveChecks {
			__antithesis_instrumentation__.Notify(45460)
			expensiveChecksArg = "--expensive-checks"
		} else {
			__antithesis_instrumentation__.Notify(45461)
		}
		__antithesis_instrumentation__.Notify(45452)
		checkCmd := fmt.Sprintf(
			"./workload check tpcc --warehouses %d %s {pgurl%s}",
			warehouses,
			expensiveChecksArg,
			c.Node(roachNodes[0]),
		)
		t.Status("beginning database verification")
		c.Run(ctx, loadNode, checkCmd)
		t.Status("finished database verification")
	}
	__antithesis_instrumentation__.Notify(45434)
	r.Add(registry.TestSpec{
		Name:  "alterpk-bank",
		Owner: registry.OwnerSQLSchema,

		Cluster: r.MakeClusterSpec(4),
		Run:     runAlterPKBank,
	})
	r.Add(registry.TestSpec{
		Name:  "alterpk-tpcc-250",
		Owner: registry.OwnerSQLSchema,

		Cluster: r.MakeClusterSpec(4, spec.CPU(32)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45462)
			runAlterPKTPCC(ctx, t, c, 250, true)
		},
	})
	__antithesis_instrumentation__.Notify(45435)
	r.Add(registry.TestSpec{
		Name:  "alterpk-tpcc-500",
		Owner: registry.OwnerSQLSchema,

		Cluster: r.MakeClusterSpec(4, spec.CPU(16)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45463)
			runAlterPKTPCC(ctx, t, c, 500, false)
		},
	})
}
