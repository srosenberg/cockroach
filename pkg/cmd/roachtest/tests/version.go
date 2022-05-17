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
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
)

func registerVersion(r registry.Registry) {
	__antithesis_instrumentation__.Notify(52287)
	runVersion := func(ctx context.Context, t test.Test, c cluster.Cluster, binaryVersion string) {
		__antithesis_instrumentation__.Notify(52289)
		nodes := c.Spec().NodeCount - 1

		if err := c.Stage(ctx, t.L(), "release", "v"+binaryVersion, "", c.Range(1, nodes)); err != nil {
			__antithesis_instrumentation__.Notify(52295)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(52296)
		}
		__antithesis_instrumentation__.Notify(52290)

		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))

		startOpts := option.DefaultStartOpts()
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

		stageDuration := 10 * time.Minute
		buffer := 10 * time.Minute
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(52297)
			t.L().Printf("local mode: speeding up test\n")
			stageDuration = 10 * time.Second
			buffer = time.Minute
		} else {
			__antithesis_instrumentation__.Notify(52298)
		}
		__antithesis_instrumentation__.Notify(52291)

		loadDuration := " --duration=" + (time.Duration(3*nodes+2)*stageDuration + buffer).String()

		var deprecatedWorkloadsStr string
		if !t.BuildVersion().AtLeast(version.MustParse("v20.2.0")) {
			__antithesis_instrumentation__.Notify(52299)
			deprecatedWorkloadsStr += " --deprecated-fk-indexes"
		} else {
			__antithesis_instrumentation__.Notify(52300)
		}
		__antithesis_instrumentation__.Notify(52292)

		workloads := []string{
			"./workload run tpcc --tolerate-errors --wait=false --drop --init --warehouses=1 " + deprecatedWorkloadsStr + loadDuration + " {pgurl:1-%d}",
			"./workload run kv --tolerate-errors --init" + loadDuration + " {pgurl:1-%d}",
		}

		m := c.NewMonitor(ctx, c.Range(1, nodes))
		for _, cmd := range workloads {
			__antithesis_instrumentation__.Notify(52301)
			cmd := cmd
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(52302)
				cmd = fmt.Sprintf(cmd, nodes)
				return c.RunE(ctx, c.Node(nodes+1), cmd)
			})
		}
		__antithesis_instrumentation__.Notify(52293)

		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(52303)
			l, err := t.L().ChildLogger("upgrader")
			if err != nil {
				__antithesis_instrumentation__.Notify(52316)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52317)
			}
			__antithesis_instrumentation__.Notify(52304)

			sleepAndCheck := func() error {
				__antithesis_instrumentation__.Notify(52318)
				t.WorkerStatus("sleeping")
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(52321)
					return ctx.Err()
				case <-time.After(stageDuration):
					__antithesis_instrumentation__.Notify(52322)
				}
				__antithesis_instrumentation__.Notify(52319)

				for i := 1; i <= nodes; i++ {
					__antithesis_instrumentation__.Notify(52323)
					t.WorkerStatus("checking ", i)
					db := c.Conn(ctx, t.L(), i)
					defer db.Close()
					rows, err := db.Query(`SHOW DATABASES`)
					if err != nil {
						__antithesis_instrumentation__.Notify(52326)
						return err
					} else {
						__antithesis_instrumentation__.Notify(52327)
					}
					__antithesis_instrumentation__.Notify(52324)
					if err := rows.Close(); err != nil {
						__antithesis_instrumentation__.Notify(52328)
						return err
					} else {
						__antithesis_instrumentation__.Notify(52329)
					}
					__antithesis_instrumentation__.Notify(52325)

					if !strings.HasPrefix(binaryVersion, "2.") {
						__antithesis_instrumentation__.Notify(52330)
						if err := c.CheckReplicaDivergenceOnDB(ctx, t.L(), db); err != nil {
							__antithesis_instrumentation__.Notify(52331)
							return errors.Wrapf(err, "node %d", i)
						} else {
							__antithesis_instrumentation__.Notify(52332)
						}
					} else {
						__antithesis_instrumentation__.Notify(52333)
					}
				}
				__antithesis_instrumentation__.Notify(52320)
				return nil
			}
			__antithesis_instrumentation__.Notify(52305)

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			db.SetMaxIdleConns(0)

			if err := sleepAndCheck(); err != nil {
				__antithesis_instrumentation__.Notify(52334)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52335)
			}
			__antithesis_instrumentation__.Notify(52306)

			stop := func(node int) error {
				__antithesis_instrumentation__.Notify(52336)
				m.ExpectDeath()
				l.Printf("stopping node %d\n", node)
				return c.StopCockroachGracefullyOnNode(ctx, t.L(), node)
			}
			__antithesis_instrumentation__.Notify(52307)

			var oldVersion string
			if err := db.QueryRowContext(ctx, `SHOW CLUSTER SETTING version`).Scan(&oldVersion); err != nil {
				__antithesis_instrumentation__.Notify(52337)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52338)
			}
			__antithesis_instrumentation__.Notify(52308)
			l.Printf("cluster version is %s\n", oldVersion)

			for i := 1; i < nodes; i++ {
				__antithesis_instrumentation__.Notify(52339)
				t.WorkerStatus("upgrading ", i)
				l.Printf("upgrading %d\n", i)
				if err := stop(i); err != nil {
					__antithesis_instrumentation__.Notify(52341)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52342)
				}
				__antithesis_instrumentation__.Notify(52340)
				c.Put(ctx, t.Cockroach(), "./cockroach", c.Node(i))

				c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(i))
				if err := sleepAndCheck(); err != nil {
					__antithesis_instrumentation__.Notify(52343)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52344)
				}
			}
			__antithesis_instrumentation__.Notify(52309)

			l.Printf("stopping last node\n")

			if err := stop(nodes); err != nil {
				__antithesis_instrumentation__.Notify(52345)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52346)
			}
			__antithesis_instrumentation__.Notify(52310)

			l.Printf("preventing automatic upgrade\n")
			if _, err := db.ExecContext(ctx,
				fmt.Sprintf("SET CLUSTER SETTING cluster.preserve_downgrade_option = '%s';", oldVersion),
			); err != nil {
				__antithesis_instrumentation__.Notify(52347)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52348)
			}
			__antithesis_instrumentation__.Notify(52311)

			l.Printf("upgrading last node\n")
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Node(nodes))
			c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(nodes))
			if err := sleepAndCheck(); err != nil {
				__antithesis_instrumentation__.Notify(52349)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52350)
			}
			__antithesis_instrumentation__.Notify(52312)

			for i := 1; i <= nodes; i++ {
				__antithesis_instrumentation__.Notify(52351)
				l.Printf("downgrading node %d\n", i)
				t.WorkerStatus("downgrading", i)
				if err := stop(i); err != nil {
					__antithesis_instrumentation__.Notify(52354)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52355)
				}
				__antithesis_instrumentation__.Notify(52352)
				if err := c.Stage(ctx, t.L(), "release", "v"+binaryVersion, "", c.Node(i)); err != nil {
					__antithesis_instrumentation__.Notify(52356)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(52357)
				}
				__antithesis_instrumentation__.Notify(52353)
				c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(i))
				if err := sleepAndCheck(); err != nil {
					__antithesis_instrumentation__.Notify(52358)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52359)
				}
			}
			__antithesis_instrumentation__.Notify(52313)

			for i := 1; i <= nodes; i++ {
				__antithesis_instrumentation__.Notify(52360)
				l.Printf("upgrading node %d (again)\n", i)
				t.WorkerStatus("upgrading", i, "(again)")
				if err := stop(i); err != nil {
					__antithesis_instrumentation__.Notify(52362)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52363)
				}
				__antithesis_instrumentation__.Notify(52361)
				c.Put(ctx, t.Cockroach(), "./cockroach", c.Node(i))
				c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Node(i))
				if err := sleepAndCheck(); err != nil {
					__antithesis_instrumentation__.Notify(52364)
					return err
				} else {
					__antithesis_instrumentation__.Notify(52365)
				}
			}
			__antithesis_instrumentation__.Notify(52314)

			l.Printf("reenabling auto-upgrade\n")
			if _, err := db.ExecContext(ctx,
				"RESET CLUSTER SETTING cluster.preserve_downgrade_option;",
			); err != nil {
				__antithesis_instrumentation__.Notify(52366)
				return err
			} else {
				__antithesis_instrumentation__.Notify(52367)
			}
			__antithesis_instrumentation__.Notify(52315)

			return sleepAndCheck()
		})
		__antithesis_instrumentation__.Notify(52294)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(52288)

	for _, n := range []int{3, 5} {
		__antithesis_instrumentation__.Notify(52368)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("version/mixed/nodes=%d", n),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(n + 1),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(52369)
				pred, err := PredecessorVersion(*t.BuildVersion())
				if err != nil {
					__antithesis_instrumentation__.Notify(52371)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(52372)
				}
				__antithesis_instrumentation__.Notify(52370)
				runVersion(ctx, t, c, pred)
			},
		})
	}
}
