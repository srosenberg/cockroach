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
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/stretchr/testify/require"
)

func registerDiskFull(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47266)
	r.Add(registry.TestSpec{
		Name:    "disk-full",
		Owner:   registry.OwnerStorage,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47267)
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(47271)
				t.Skip("you probably don't want to fill your local disk")
			} else {
				__antithesis_instrumentation__.Notify(47272)
			}
			__antithesis_instrumentation__.Notify(47268)

			nodes := c.Spec().NodeCount - 1
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, c.Spec().NodeCount))
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))

			db := c.Conn(ctx, t.L(), 1)
			err := WaitFor3XReplication(ctx, t, db)
			require.NoError(t, err)
			_ = db.Close()

			t.Status("running workload")
			m := c.NewMonitor(ctx, c.Range(1, nodes))
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(47273)
				cmd := fmt.Sprintf(
					"./cockroach workload run kv --tolerate-errors --init --read-percent=0"+
						" --concurrency=10 --duration=4m {pgurl:2-%d}",
					nodes)
				c.Run(ctx, c.Node(nodes+1), cmd)
				return nil
			})
			__antithesis_instrumentation__.Notify(47269)

			c.Run(ctx, c.Range(1, nodes), "stat {store-dir}/auxiliary/EMERGENCY_BALLAST")

			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(47274)
				const n = 1

				t.L().Printf("filling disk on %d\n", n)

				m.ExpectDeath()
				c.Run(ctx, c.Node(n), "./cockroach debug ballast {store-dir}/largefile --size=100% || true")

				for isLive := true; isLive; {
					__antithesis_instrumentation__.Notify(47278)
					db := c.Conn(ctx, t.L(), 2)
					err := db.QueryRow(`SELECT is_live FROM crdb_internal.gossip_nodes WHERE node_id = 1;`).Scan(&isLive)
					if err != nil {
						__antithesis_instrumentation__.Notify(47280)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(47281)
					}
					__antithesis_instrumentation__.Notify(47279)
					if isLive {
						__antithesis_instrumentation__.Notify(47282)
						t.L().Printf("waiting for n%d to die due to full disk\n", n)
						time.Sleep(time.Second)
					} else {
						__antithesis_instrumentation__.Notify(47283)
					}
				}
				__antithesis_instrumentation__.Notify(47275)

				t.L().Printf("node n%d died as expected\n", n)

				for start := timeutil.Now(); timeutil.Since(start) < 30*time.Second; {
					__antithesis_instrumentation__.Notify(47284)
					if t.Failed() {
						__antithesis_instrumentation__.Notify(47288)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(47289)
					}
					__antithesis_instrumentation__.Notify(47285)
					t.L().Printf("starting n%d when disk is full\n", n)

					m.ExpectDeath()

					err := c.StartE(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(n))
					t.L().Printf("starting n%d: error %v", n, err)
					if err == nil {
						__antithesis_instrumentation__.Notify(47290)
						t.Fatal("node successfully started unexpectedly")
					} else {
						__antithesis_instrumentation__.Notify(47291)
						if strings.Contains(cluster.GetStderr(err), "a panic has occurred") {
							__antithesis_instrumentation__.Notify(47292)
							t.Fatal(err)
						} else {
							__antithesis_instrumentation__.Notify(47293)
						}
					}
					__antithesis_instrumentation__.Notify(47286)

					result, err := c.RunWithDetailsSingleNode(ctx, t.L(), c.Node(n),
						`systemctl status cockroach.service | grep 'Main PID' | grep -oE '\((.+)\)'`,
					)
					if err != nil {
						__antithesis_instrumentation__.Notify(47294)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(47295)
					}
					__antithesis_instrumentation__.Notify(47287)
					exitLogs := strings.TrimSpace(result.Stdout)
					const want = `(code=exited, status=10)`
					if exitLogs != want {
						__antithesis_instrumentation__.Notify(47296)
						t.Fatalf("cockroach systemd status: got %q, want %q", exitLogs, want)
					} else {
						__antithesis_instrumentation__.Notify(47297)
					}
				}
				__antithesis_instrumentation__.Notify(47276)

				t.L().Printf("removing the emergency ballast on n%d\n", n)
				m.ExpectDeath()
				c.Run(ctx, c.Node(n), "rm -f {store-dir}/auxiliary/EMERGENCY_BALLAST")
				if err := c.StartE(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(n)); err != nil {
					__antithesis_instrumentation__.Notify(47298)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(47299)
				}
				__antithesis_instrumentation__.Notify(47277)
				m.ResetDeaths()

				time.Sleep(30 * time.Second)
				t.L().Printf("removing n%d's large file to free up available disk space.\n", n)
				c.Run(ctx, c.Node(n), "rm -f {store-dir}/largefile")

				t.L().Printf("waiting for node n%d's emergency ballast to be restored.\n", n)
				for {
					__antithesis_instrumentation__.Notify(47300)
					err := c.RunE(ctx, c.Node(1), "stat {store-dir}/auxiliary/EMERGENCY_BALLAST")
					if err == nil {
						__antithesis_instrumentation__.Notify(47302)
						return nil
					} else {
						__antithesis_instrumentation__.Notify(47303)
					}
					__antithesis_instrumentation__.Notify(47301)
					t.L().Printf("node n%d's emergency ballast doesn't exist yet\n", n)
					time.Sleep(time.Second)
				}
			})
			__antithesis_instrumentation__.Notify(47270)
			m.Wait()
		},
	})
}
