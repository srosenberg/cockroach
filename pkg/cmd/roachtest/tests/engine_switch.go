package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
	"github.com/cockroachdb/errors"
	"golang.org/x/exp/rand"
)

func registerEngineSwitch(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47589)
	runEngineSwitch := func(ctx context.Context, t test.Test, c cluster.Cluster, additionalArgs ...string) {
		__antithesis_instrumentation__.Notify(47592)
		roachNodes := c.Range(1, c.Spec().NodeCount-1)
		loadNode := c.Node(c.Spec().NodeCount)
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", loadNode)
		c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)

		rockdbStartOpts := option.DefaultStartOpts()
		rockdbStartOpts.RoachprodOpts.ExtraArgs = append(rockdbStartOpts.RoachprodOpts.ExtraArgs, "--storage-engine=rocksdb")

		pebbleStartOpts := option.DefaultStartOpts()
		pebbleStartOpts.RoachprodOpts.ExtraArgs = append(pebbleStartOpts.RoachprodOpts.ExtraArgs, "--storage-engine=pebble")
		c.Start(ctx, t.L(), rockdbStartOpts, install.MakeClusterSettings(), roachNodes)
		stageDuration := 1 * time.Minute
		if c.IsLocal() {
			__antithesis_instrumentation__.Notify(47597)
			t.L().Printf("local mode: speeding up test\n")
			stageDuration = 10 * time.Second
		} else {
			__antithesis_instrumentation__.Notify(47598)
		}
		__antithesis_instrumentation__.Notify(47593)
		numIters := 5 * len(roachNodes)

		loadDuration := " --duration=" + (time.Duration(numIters) * stageDuration).String()

		var deprecatedWorkloadsStr string
		if !t.BuildVersion().AtLeast(version.MustParse("v20.2.0")) {
			__antithesis_instrumentation__.Notify(47599)
			deprecatedWorkloadsStr += " --deprecated-fk-indexes"
		} else {
			__antithesis_instrumentation__.Notify(47600)
		}
		__antithesis_instrumentation__.Notify(47594)

		workloads := []string{

			"./workload run tpcc --tolerate-errors --wait=false --drop --init" + deprecatedWorkloadsStr + " --warehouses=1 " + loadDuration + " {pgurl:1-%d}",
		}
		checkWorkloads := []string{
			"./workload check tpcc --warehouses=1 --expensive-checks=true {pgurl:1}",
		}
		m := c.NewMonitor(ctx, roachNodes)
		for _, cmd := range workloads {
			__antithesis_instrumentation__.Notify(47601)
			cmd := cmd
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(47602)
				cmd = fmt.Sprintf(cmd, len(roachNodes))
				return c.RunE(ctx, loadNode, cmd)
			})
		}
		__antithesis_instrumentation__.Notify(47595)

		usingPebble := make([]bool, len(roachNodes))
		rng := rand.New(rand.NewSource(uint64(timeutil.Now().UnixNano())))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(47603)
			l, err := t.L().ChildLogger("engine-switcher")
			if err != nil {
				__antithesis_instrumentation__.Notify(47607)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47608)
			}
			__antithesis_instrumentation__.Notify(47604)

			sleepAndCheck := func() error {
				__antithesis_instrumentation__.Notify(47609)
				t.WorkerStatus("sleeping")
				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(47612)
					return ctx.Err()
				case <-time.After(stageDuration):
					__antithesis_instrumentation__.Notify(47613)
				}
				__antithesis_instrumentation__.Notify(47610)

				for i := 1; i <= len(roachNodes); i++ {
					__antithesis_instrumentation__.Notify(47614)
					t.WorkerStatus("checking ", i)
					db := c.Conn(ctx, t.L(), i)
					defer db.Close()
					rows, err := db.Query(`SHOW DATABASES`)
					if err != nil {
						__antithesis_instrumentation__.Notify(47617)
						return err
					} else {
						__antithesis_instrumentation__.Notify(47618)
					}
					__antithesis_instrumentation__.Notify(47615)
					if err := rows.Close(); err != nil {
						__antithesis_instrumentation__.Notify(47619)
						return err
					} else {
						__antithesis_instrumentation__.Notify(47620)
					}
					__antithesis_instrumentation__.Notify(47616)
					if err := c.CheckReplicaDivergenceOnDB(ctx, t.L(), db); err != nil {
						__antithesis_instrumentation__.Notify(47621)
						return errors.Wrapf(err, "node %d", i)
					} else {
						__antithesis_instrumentation__.Notify(47622)
					}
				}
				__antithesis_instrumentation__.Notify(47611)
				return nil
			}
			__antithesis_instrumentation__.Notify(47605)

			for i := 0; i < numIters; i++ {
				__antithesis_instrumentation__.Notify(47623)

				if err := sleepAndCheck(); err != nil {
					__antithesis_instrumentation__.Notify(47628)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47629)
				}
				__antithesis_instrumentation__.Notify(47624)

				stop := func(node int) error {
					__antithesis_instrumentation__.Notify(47630)
					m.ExpectDeath()
					if rng.Intn(2) == 0 {
						__antithesis_instrumentation__.Notify(47632)
						l.Printf("stopping node gracefully %d\n", node)
						return c.StopCockroachGracefullyOnNode(ctx, t.L(), node)
					} else {
						__antithesis_instrumentation__.Notify(47633)
					}
					__antithesis_instrumentation__.Notify(47631)
					l.Printf("stopping node %d\n", node)
					c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(node))
					return nil
				}
				__antithesis_instrumentation__.Notify(47625)

				i := rng.Intn(len(roachNodes))
				var opts option.StartOpts
				usingPebble[i] = !usingPebble[i]
				if usingPebble[i] {
					__antithesis_instrumentation__.Notify(47634)
					opts = pebbleStartOpts
				} else {
					__antithesis_instrumentation__.Notify(47635)
					opts = rockdbStartOpts
				}
				__antithesis_instrumentation__.Notify(47626)
				t.WorkerStatus("switching ", i+1)
				l.Printf("switching %d\n", i+1)
				if err := stop(i + 1); err != nil {
					__antithesis_instrumentation__.Notify(47636)
					return err
				} else {
					__antithesis_instrumentation__.Notify(47637)
				}
				__antithesis_instrumentation__.Notify(47627)
				c.Start(ctx, t.L(), opts, install.MakeClusterSettings(), c.Node(i+1))
			}
			__antithesis_instrumentation__.Notify(47606)
			return sleepAndCheck()
		})
		__antithesis_instrumentation__.Notify(47596)
		m.Wait()

		for _, cmd := range checkWorkloads {
			__antithesis_instrumentation__.Notify(47638)
			c.Run(ctx, loadNode, cmd)
		}
	}
	__antithesis_instrumentation__.Notify(47590)

	n := 3
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("engine/switch/nodes=%d", n),
		Owner:   registry.OwnerStorage,
		Skip:    "rocksdb removed in 21.1",
		Cluster: r.MakeClusterSpec(n + 1),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47639)
			runEngineSwitch(ctx, t, c)
		},
	})
	__antithesis_instrumentation__.Notify(47591)
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("engine/switch/encrypted/nodes=%d", n),
		Owner:   registry.OwnerStorage,
		Skip:    "rocksdb removed in 21.1",
		Cluster: r.MakeClusterSpec(n + 1),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(47640)
			runEngineSwitch(ctx, t, c, "--encrypt=true")
		},
	})
}
