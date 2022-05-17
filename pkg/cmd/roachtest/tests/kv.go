package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/ts/tspb"
	"github.com/cockroachdb/cockroach/pkg/util/httputil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/stretchr/testify/require"
)

const envKVFlags = "ROACHTEST_KV_FLAGS"

func registerKV(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48847)
	type kvOptions struct {
		nodes       int
		cpus        int
		readPercent int

		spanReads bool
		batchSize int
		blockSize int
		splits    int

		disableLoadSplits        bool
		encryption               bool
		sequential               bool
		admissionControlDisabled bool
		concMultiplier           int
		duration                 time.Duration
		tracing                  bool
		tags                     []string
		owner                    registry.Owner
	}
	computeNumSplits := func(opts kvOptions) int {
		__antithesis_instrumentation__.Notify(48850)

		const defaultNumSplits = 1000
		switch {
		case opts.splits == 0:
			__antithesis_instrumentation__.Notify(48851)
			return defaultNumSplits
		case opts.splits < 0:
			__antithesis_instrumentation__.Notify(48852)
			return 0
		default:
			__antithesis_instrumentation__.Notify(48853)
			return opts.splits
		}
	}
	__antithesis_instrumentation__.Notify(48848)
	runKV := func(ctx context.Context, t test.Test, c cluster.Cluster, opts kvOptions) {
		__antithesis_instrumentation__.Notify(48854)
		nodes := c.Spec().NodeCount - 1
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.EncryptedStores = opts.encryption
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()
		if opts.disableLoadSplits {
			__antithesis_instrumentation__.Notify(48858)
			if _, err := db.ExecContext(ctx, "SET CLUSTER SETTING kv.range_split.by_load_enabled = 'false'"); err != nil {
				__antithesis_instrumentation__.Notify(48859)
				t.Fatalf("failed to disable load based splitting: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(48860)
			}
		} else {
			__antithesis_instrumentation__.Notify(48861)
		}
		__antithesis_instrumentation__.Notify(48855)
		if opts.tracing {
			__antithesis_instrumentation__.Notify(48862)
			if _, err := db.ExecContext(ctx, "SET CLUSTER SETTING trace.debug.enable = true"); err != nil {
				__antithesis_instrumentation__.Notify(48863)
				t.Fatalf("failed to enable tracing: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(48864)
			}
		} else {
			__antithesis_instrumentation__.Notify(48865)
		}
		__antithesis_instrumentation__.Notify(48856)
		SetAdmissionControl(ctx, t, c, !opts.admissionControlDisabled)

		t.Status("running workload")
		m := c.NewMonitor(ctx, c.Range(1, nodes))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(48866)
			concurrencyMultiplier := 64
			if opts.concMultiplier != 0 {
				__antithesis_instrumentation__.Notify(48874)
				concurrencyMultiplier = opts.concMultiplier
			} else {
				__antithesis_instrumentation__.Notify(48875)
			}
			__antithesis_instrumentation__.Notify(48867)
			concurrency := ifLocal(c, "", " --concurrency="+fmt.Sprint(nodes*concurrencyMultiplier))

			splits := " --splits=" + strconv.Itoa(computeNumSplits(opts))
			if opts.duration == 0 {
				__antithesis_instrumentation__.Notify(48876)
				opts.duration = 30 * time.Minute
			} else {
				__antithesis_instrumentation__.Notify(48877)
			}
			__antithesis_instrumentation__.Notify(48868)
			duration := " --duration=" + ifLocal(c, "10s", opts.duration.String())
			var readPercent string
			if opts.spanReads {
				__antithesis_instrumentation__.Notify(48878)

				readPercent =
					fmt.Sprintf(" --span-percent=%d --span-limit=1 --sfu-writes=true --cycle-length=1000",
						opts.readPercent)
			} else {
				__antithesis_instrumentation__.Notify(48879)
				readPercent = fmt.Sprintf(" --read-percent=%d", opts.readPercent)
			}
			__antithesis_instrumentation__.Notify(48869)
			histograms := " --histograms=" + t.PerfArtifactsDir() + "/stats.json"
			var batchSize string
			if opts.batchSize > 0 {
				__antithesis_instrumentation__.Notify(48880)
				batchSize = fmt.Sprintf(" --batch=%d", opts.batchSize)
			} else {
				__antithesis_instrumentation__.Notify(48881)
			}
			__antithesis_instrumentation__.Notify(48870)

			var blockSize string
			if opts.blockSize > 0 {
				__antithesis_instrumentation__.Notify(48882)
				blockSize = fmt.Sprintf(" --min-block-bytes=%d --max-block-bytes=%d",
					opts.blockSize, opts.blockSize)
			} else {
				__antithesis_instrumentation__.Notify(48883)
			}
			__antithesis_instrumentation__.Notify(48871)

			var sequential string
			if opts.sequential {
				__antithesis_instrumentation__.Notify(48884)
				splits = ""
				sequential = " --sequential"
			} else {
				__antithesis_instrumentation__.Notify(48885)
			}
			__antithesis_instrumentation__.Notify(48872)

			var envFlags string
			if e := os.Getenv(envKVFlags); e != "" {
				__antithesis_instrumentation__.Notify(48886)
				envFlags = " " + e
			} else {
				__antithesis_instrumentation__.Notify(48887)
			}
			__antithesis_instrumentation__.Notify(48873)

			cmd := fmt.Sprintf("./workload run kv --init"+
				histograms+concurrency+splits+duration+readPercent+batchSize+blockSize+sequential+envFlags+
				" {pgurl:1-%d}", nodes)
			c.Run(ctx, c.Node(nodes+1), cmd)
			return nil
		})
		__antithesis_instrumentation__.Notify(48857)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(48849)

	for _, opts := range []kvOptions{

		{nodes: 1, cpus: 8, readPercent: 0},

		{nodes: 1, cpus: 8, readPercent: 50, concMultiplier: 8192},

		{nodes: 1, cpus: 8, readPercent: 0, concMultiplier: 4096, blockSize: 1 << 16},
		{nodes: 1, cpus: 8, readPercent: 95},
		{nodes: 1, cpus: 32, readPercent: 0},
		{nodes: 1, cpus: 32, readPercent: 95},
		{nodes: 3, cpus: 8, readPercent: 0},
		{nodes: 3, cpus: 8, readPercent: 95},
		{nodes: 3, cpus: 8, readPercent: 95, tracing: true, owner: registry.OwnerObsInf},
		{nodes: 3, cpus: 8, readPercent: 0, splits: -1},
		{nodes: 3, cpus: 8, readPercent: 95, splits: -1},
		{nodes: 3, cpus: 32, readPercent: 0},
		{nodes: 3, cpus: 32, readPercent: 95},
		{nodes: 3, cpus: 32, readPercent: 0, admissionControlDisabled: true},
		{nodes: 3, cpus: 32, readPercent: 95, admissionControlDisabled: true},
		{nodes: 3, cpus: 32, readPercent: 0, splits: -1},
		{nodes: 3, cpus: 32, readPercent: 95, splits: -1},

		{nodes: 3, cpus: 8, readPercent: 0, blockSize: 1 << 12},
		{nodes: 3, cpus: 8, readPercent: 95, blockSize: 1 << 12},
		{nodes: 3, cpus: 32, readPercent: 0, blockSize: 1 << 12},
		{nodes: 3, cpus: 32, readPercent: 95, blockSize: 1 << 12},
		{nodes: 3, cpus: 8, readPercent: 0, blockSize: 1 << 16},
		{nodes: 3, cpus: 8, readPercent: 95, blockSize: 1 << 16},
		{nodes: 3, cpus: 32, readPercent: 0, blockSize: 1 << 16},
		{nodes: 3, cpus: 32, readPercent: 95, blockSize: 1 << 16},
		{nodes: 3, cpus: 32, readPercent: 0, blockSize: 1 << 16,
			admissionControlDisabled: true},
		{nodes: 3, cpus: 32, readPercent: 95, blockSize: 1 << 16,
			admissionControlDisabled: true},

		{nodes: 3, cpus: 8, readPercent: 0, batchSize: 16},
		{nodes: 3, cpus: 8, readPercent: 95, batchSize: 16},

		{nodes: 3, cpus: 96, readPercent: 0},
		{nodes: 3, cpus: 96, readPercent: 95},
		{nodes: 4, cpus: 96, readPercent: 50, batchSize: 64},

		{nodes: 1, cpus: 8, readPercent: 0, encryption: true},
		{nodes: 1, cpus: 8, readPercent: 95, encryption: true},
		{nodes: 3, cpus: 8, readPercent: 0, encryption: true},
		{nodes: 3, cpus: 8, readPercent: 95, encryption: true},

		{nodes: 3, cpus: 32, readPercent: 0, sequential: true},
		{nodes: 3, cpus: 32, readPercent: 95, sequential: true},

		{nodes: 1, cpus: 8, readPercent: 95, spanReads: true, splits: -1, disableLoadSplits: true, sequential: true},
		{nodes: 1, cpus: 32, readPercent: 95, spanReads: true, splits: -1, disableLoadSplits: true, sequential: true},

		{nodes: 32, cpus: 8, readPercent: 0, tags: []string{"weekly"}, duration: time.Hour},
		{nodes: 32, cpus: 8, readPercent: 95, tags: []string{"weekly"}, duration: time.Hour},
	} {
		__antithesis_instrumentation__.Notify(48888)
		opts := opts

		var nameParts []string
		var limitedSpanStr string
		if opts.spanReads {
			__antithesis_instrumentation__.Notify(48901)
			limitedSpanStr = "limited-spans"
		} else {
			__antithesis_instrumentation__.Notify(48902)
		}
		__antithesis_instrumentation__.Notify(48889)
		nameParts = append(nameParts, fmt.Sprintf("kv%d%s", opts.readPercent, limitedSpanStr))
		if len(opts.tags) > 0 {
			__antithesis_instrumentation__.Notify(48903)
			nameParts = append(nameParts, strings.Join(opts.tags, "/"))
		} else {
			__antithesis_instrumentation__.Notify(48904)
		}
		__antithesis_instrumentation__.Notify(48890)
		nameParts = append(nameParts, fmt.Sprintf("enc=%t", opts.encryption))
		nameParts = append(nameParts, fmt.Sprintf("nodes=%d", opts.nodes))
		if opts.cpus != 8 {
			__antithesis_instrumentation__.Notify(48905)
			nameParts = append(nameParts, fmt.Sprintf("cpu=%d", opts.cpus))
		} else {
			__antithesis_instrumentation__.Notify(48906)
		}
		__antithesis_instrumentation__.Notify(48891)
		if opts.batchSize != 0 {
			__antithesis_instrumentation__.Notify(48907)
			nameParts = append(nameParts, fmt.Sprintf("batch=%d", opts.batchSize))
		} else {
			__antithesis_instrumentation__.Notify(48908)
		}
		__antithesis_instrumentation__.Notify(48892)
		if opts.blockSize != 0 {
			__antithesis_instrumentation__.Notify(48909)
			nameParts = append(nameParts, fmt.Sprintf("size=%dkb", opts.blockSize>>10))
		} else {
			__antithesis_instrumentation__.Notify(48910)
		}
		__antithesis_instrumentation__.Notify(48893)
		if opts.splits != 0 {
			__antithesis_instrumentation__.Notify(48911)
			nameParts = append(nameParts, fmt.Sprintf("splt=%d", computeNumSplits(opts)))
		} else {
			__antithesis_instrumentation__.Notify(48912)
		}
		__antithesis_instrumentation__.Notify(48894)
		if opts.sequential {
			__antithesis_instrumentation__.Notify(48913)
			nameParts = append(nameParts, "seq")
		} else {
			__antithesis_instrumentation__.Notify(48914)
		}
		__antithesis_instrumentation__.Notify(48895)
		if opts.admissionControlDisabled {
			__antithesis_instrumentation__.Notify(48915)
			nameParts = append(nameParts, "no-admission")
		} else {
			__antithesis_instrumentation__.Notify(48916)
		}
		__antithesis_instrumentation__.Notify(48896)
		if opts.concMultiplier != 0 {
			__antithesis_instrumentation__.Notify(48917)
			nameParts = append(nameParts, fmt.Sprintf("conc=%d", opts.concMultiplier))
		} else {
			__antithesis_instrumentation__.Notify(48918)
		}
		__antithesis_instrumentation__.Notify(48897)
		if opts.disableLoadSplits {
			__antithesis_instrumentation__.Notify(48919)
			nameParts = append(nameParts, "no-load-splitting")
		} else {
			__antithesis_instrumentation__.Notify(48920)
		}
		__antithesis_instrumentation__.Notify(48898)
		if opts.tracing {
			__antithesis_instrumentation__.Notify(48921)
			nameParts = append(nameParts, "tracing")
		} else {
			__antithesis_instrumentation__.Notify(48922)
		}
		__antithesis_instrumentation__.Notify(48899)
		owner := registry.OwnerKV
		if opts.owner != "" {
			__antithesis_instrumentation__.Notify(48923)
			owner = opts.owner
		} else {
			__antithesis_instrumentation__.Notify(48924)
		}
		__antithesis_instrumentation__.Notify(48900)
		r.Add(registry.TestSpec{
			Name:    strings.Join(nameParts, "/"),
			Owner:   owner,
			Cluster: r.MakeClusterSpec(opts.nodes+1, spec.CPU(opts.cpus)),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(48925)
				runKV(ctx, t, c, opts)
			},
			Tags: opts.tags,
		})
	}
}

func registerKVContention(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48926)
	const nodes = 4
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("kv/contention/nodes=%d", nodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(nodes + 1),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48927)
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))

			settings := install.MakeClusterSettings()
			settings.Env = append(settings.Env, "COCKROACH_TXN_LIVENESS_HEARTBEAT_MULTIPLIER=600")
			c.Start(ctx, t.L(), option.DefaultStartOpts(), settings, c.Range(1, nodes))

			conn := c.Conn(ctx, t.L(), 1)

			if _, err := conn.Exec(`
				SET CLUSTER SETTING trace.debug.enable = true;
			`); err != nil {
				__antithesis_instrumentation__.Notify(48931)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48932)
			}
			__antithesis_instrumentation__.Notify(48928)

			if _, err := conn.Exec(`
				SET CLUSTER SETTING kv.lock_table.deadlock_detection_push_delay = '5ms'
			`); err != nil && func() bool {
				__antithesis_instrumentation__.Notify(48933)
				return !strings.Contains(err.Error(), "unknown cluster setting") == true
			}() == true {
				__antithesis_instrumentation__.Notify(48934)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(48935)
			}
			__antithesis_instrumentation__.Notify(48929)

			t.Status("running workload")
			m := c.NewMonitor(ctx, c.Range(1, nodes))
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(48936)

				const cycleLength = 512
				const concurrency = 128
				const avgConcPerKey = 1
				const batchSize = avgConcPerKey * (cycleLength / concurrency)

				splits := nodes

				const duration = 1 * time.Hour
				cmd := fmt.Sprintf("./workload run kv --init --secondary-index --duration=%s "+
					"--cycle-length=%d --concurrency=%d --batch=%d --splits=%d {pgurl:1-%d}",
					duration, cycleLength, concurrency, batchSize, splits, nodes)
				start := timeutil.Now()
				c.Run(ctx, c.Node(nodes+1), cmd)
				end := timeutil.Now()

				const minQPS = 50
				verifyTxnPerSecond(ctx, c, t, c.Node(1), start, end, minQPS, 0.1)
				return nil
			})
			__antithesis_instrumentation__.Notify(48930)
			m.Wait()
		},
	})
}

func registerKVQuiescenceDead(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48937)
	r.Add(registry.TestSpec{
		Name:    "kv/quiescence/nodes=3",
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48938)
			nodes := c.Spec().NodeCount - 1
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))

			run := func(cmd string, lastDown bool) {
				__antithesis_instrumentation__.Notify(48944)
				n := nodes
				if lastDown {
					__antithesis_instrumentation__.Notify(48947)
					n--
				} else {
					__antithesis_instrumentation__.Notify(48948)
				}
				__antithesis_instrumentation__.Notify(48945)
				m := c.NewMonitor(ctx, c.Range(1, n))
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(48949)
					t.WorkerStatus(cmd)
					defer t.WorkerStatus()
					return c.RunE(ctx, c.Node(nodes+1), cmd)
				})
				__antithesis_instrumentation__.Notify(48946)
				m.Wait()
			}
			__antithesis_instrumentation__.Notify(48939)

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			err := WaitFor3XReplication(ctx, t, db)
			require.NoError(t, err)

			qps := func(f func()) float64 {
				__antithesis_instrumentation__.Notify(48950)

				numInserts := func() float64 {
					__antithesis_instrumentation__.Notify(48952)
					var v float64
					if err = db.QueryRowContext(
						ctx, `SELECT value FROM crdb_internal.node_metrics WHERE name = 'sql.insert.count'`,
					).Scan(&v); err != nil {
						__antithesis_instrumentation__.Notify(48954)
						t.Fatal(err)
					} else {
						__antithesis_instrumentation__.Notify(48955)
					}
					__antithesis_instrumentation__.Notify(48953)
					return v
				}
				__antithesis_instrumentation__.Notify(48951)

				tBegin := timeutil.Now()
				before := numInserts()
				f()
				after := numInserts()
				return (after - before) / timeutil.Since(tBegin).Seconds()
			}
			__antithesis_instrumentation__.Notify(48940)

			const kv = "./workload run kv --duration=10m --read-percent=0"

			run("./workload run kv --init --max-ops=1 --splits 10000 --concurrency 100 {pgurl:1}", false)
			run(kv+" --seed 0 {pgurl:1}", true)

			qpsAllUp := qps(func() {
				__antithesis_instrumentation__.Notify(48956)
				run(kv+" --seed 1 {pgurl:1}", true)
			})
			__antithesis_instrumentation__.Notify(48941)

			c.Run(ctx, c.Node(nodes), "./cockroach quit --insecure --host=:{pgport:3}")
			c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(nodes))

			qpsOneDown := qps(func() {
				__antithesis_instrumentation__.Notify(48957)

				run(kv+" --seed 2 {pgurl:1}", true)
			})
			__antithesis_instrumentation__.Notify(48942)
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(nodes))

			if minFrac, actFrac := 0.8, qpsOneDown/qpsAllUp; actFrac < minFrac {
				__antithesis_instrumentation__.Notify(48958)
				t.Fatalf(
					"QPS dropped from %.2f to %.2f (factor of %.2f, min allowed %.2f)",
					qpsAllUp, qpsOneDown, actFrac, minFrac,
				)
			} else {
				__antithesis_instrumentation__.Notify(48959)
			}
			__antithesis_instrumentation__.Notify(48943)
			t.L().Printf("QPS went from %.2f to %2.f with one node down\n", qpsAllUp, qpsOneDown)
		},
	})
}

func registerKVGracefulDraining(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48960)
	r.Add(registry.TestSpec{
		Name:    "kv/gracefuldraining/nodes=3",
		Owner:   registry.OwnerKV,
		Skip:    "https://github.com/cockroachdb/cockroach/issues/59094",
		Cluster: r.MakeClusterSpec(4),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48961)
			nodes := c.Spec().NodeCount - 1
			c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
			c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))

			t.Status("starting cluster")

			startOpts := option.DefaultStartOpts()
			startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--vmodule=store=2,store_rebalancer=2")
			c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			err := WaitFor3XReplication(ctx, t, db)
			require.NoError(t, err)

			t.Status("initializing workload")

			splitCmd := "./workload run kv --init --max-ops=1 --splits 100 {pgurl:1}"
			c.Run(ctx, c.Node(nodes+1), splitCmd)

			m := c.NewMonitor(ctx, c.Nodes(1, nodes))

			const specifiedQPS = 1000

			expectedQPS := specifiedQPS * 0.9

			t.Status("starting workload")
			workloadStartTime := timeutil.Now()
			desiredRunDuration := 5 * time.Minute
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(48964)
				cmd := fmt.Sprintf(
					"./workload run kv --duration=%s --read-percent=0 --tolerate-errors --max-rate=%d {pgurl:1-%d}",
					desiredRunDuration,
					specifiedQPS, nodes-1)
				t.WorkerStatus(cmd)
				defer func() {
					__antithesis_instrumentation__.Notify(48966)
					t.WorkerStatus("workload command completed")
					t.WorkerStatus()
				}()
				__antithesis_instrumentation__.Notify(48965)
				return c.RunE(ctx, c.Node(nodes+1), cmd)
			})
			__antithesis_instrumentation__.Notify(48962)

			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(48967)
				defer t.WorkerStatus()

				t.WorkerStatus("waiting for perf to stabilize")

				adminURLs, err := c.ExternalAdminUIAddr(ctx, t.L(), c.Node(1))
				if err != nil {
					__antithesis_instrumentation__.Notify(48975)
					return err
				} else {
					__antithesis_instrumentation__.Notify(48976)
				}
				__antithesis_instrumentation__.Notify(48968)
				url := "http://" + adminURLs[0] + "/ts/query"
				getQPSTimeSeries := func(start, end time.Time) ([]tspb.TimeSeriesDatapoint, error) {
					__antithesis_instrumentation__.Notify(48977)
					request := tspb.TimeSeriesQueryRequest{
						StartNanos: start.UnixNano(),
						EndNanos:   end.UnixNano(),

						SampleNanos: base.DefaultMetricsSampleInterval.Nanoseconds(),
						Queries: []tspb.Query{
							{
								Name:             "cr.node.sql.query.count",
								Downsampler:      tspb.TimeSeriesQueryAggregator_AVG.Enum(),
								SourceAggregator: tspb.TimeSeriesQueryAggregator_SUM.Enum(),
								Derivative:       tspb.TimeSeriesQueryDerivative_NON_NEGATIVE_DERIVATIVE.Enum(),
							},
						},
					}
					var response tspb.TimeSeriesQueryResponse
					if err := httputil.PostJSON(http.Client{}, url, &request, &response); err != nil {
						__antithesis_instrumentation__.Notify(48980)
						return nil, err
					} else {
						__antithesis_instrumentation__.Notify(48981)
					}
					__antithesis_instrumentation__.Notify(48978)
					if len(response.Results[0].Datapoints) <= 1 {
						__antithesis_instrumentation__.Notify(48982)
						return nil, errors.Newf("not enough datapoints in timeseries query response: %+v", response)
					} else {
						__antithesis_instrumentation__.Notify(48983)
					}
					__antithesis_instrumentation__.Notify(48979)
					return response.Results[0].Datapoints, nil
				}
				__antithesis_instrumentation__.Notify(48969)

				waitBegin := timeutil.Now()

				if err := retry.ForDuration(1*time.Minute, func() (err error) {
					__antithesis_instrumentation__.Notify(48984)
					defer func() {
						__antithesis_instrumentation__.Notify(48988)
						if timeutil.Since(waitBegin) > 3*time.Second && func() bool {
							__antithesis_instrumentation__.Notify(48989)
							return err != nil == true
						}() == true {
							__antithesis_instrumentation__.Notify(48990)
							t.Status(fmt.Sprintf("perf not stable yet: %v", err))
						} else {
							__antithesis_instrumentation__.Notify(48991)
						}
					}()
					__antithesis_instrumentation__.Notify(48985)
					now := timeutil.Now()
					datapoints, err := getQPSTimeSeries(workloadStartTime, now)
					if err != nil {
						__antithesis_instrumentation__.Notify(48992)
						return err
					} else {
						__antithesis_instrumentation__.Notify(48993)
					}
					__antithesis_instrumentation__.Notify(48986)

					dp := datapoints[len(datapoints)-1]
					if qps := dp.Value; qps < expectedQPS {
						__antithesis_instrumentation__.Notify(48994)
						return errors.Newf(
							"QPS of %.2f at time %v is below minimum allowable QPS of %.2f; entire timeseries: %+v",
							qps, timeutil.Unix(0, dp.TimestampNanos), expectedQPS, datapoints)
					} else {
						__antithesis_instrumentation__.Notify(48995)
					}
					__antithesis_instrumentation__.Notify(48987)

					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(48996)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(48997)
				}
				__antithesis_instrumentation__.Notify(48970)
				t.Status("detected stable perf before restarts: OK")

				stablePerfStartTime := timeutil.Now()

				t.WorkerStatus("gracefully draining and restarting nodes")

				for i := 0; i < 2; i++ {
					__antithesis_instrumentation__.Notify(48998)
					if i > 0 {
						__antithesis_instrumentation__.Notify(49001)

						t.Status("letting workload run with all nodes")
						select {
						case <-ctx.Done():
							__antithesis_instrumentation__.Notify(49002)
							return nil
						case <-time.After(1 * time.Minute):
							__antithesis_instrumentation__.Notify(49003)
						}
					} else {
						__antithesis_instrumentation__.Notify(49004)
					}
					__antithesis_instrumentation__.Notify(48999)
					m.ExpectDeath()
					c.Run(ctx, c.Node(nodes), "./cockroach quit --insecure --host=:{pgport:3}")
					c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(nodes))
					t.Status("letting workload run with one node down")
					select {
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(49005)
						return nil
					case <-time.After(1 * time.Minute):
						__antithesis_instrumentation__.Notify(49006)
					}
					__antithesis_instrumentation__.Notify(49000)
					c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(nodes))
					m.ResetDeaths()
				}
				__antithesis_instrumentation__.Notify(48971)

				t.WorkerStatus("checking workload is still running")
				runDuration := timeutil.Since(workloadStartTime)
				if runDuration > desiredRunDuration-10*time.Second {
					__antithesis_instrumentation__.Notify(49007)
					t.Fatalf("not enough workload time left to reliably determine performance (%s left)",
						desiredRunDuration-runDuration)
				} else {
					__antithesis_instrumentation__.Notify(49008)
				}
				__antithesis_instrumentation__.Notify(48972)

				t.WorkerStatus("checking for perf throughout the test")

				endTestTime := timeutil.Now()
				datapoints, err := getQPSTimeSeries(stablePerfStartTime, endTestTime)
				if err != nil {
					__antithesis_instrumentation__.Notify(49009)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(49010)
				}
				__antithesis_instrumentation__.Notify(48973)

				for _, dp := range datapoints {
					__antithesis_instrumentation__.Notify(49011)
					if qps := dp.Value; qps < expectedQPS {
						__antithesis_instrumentation__.Notify(49012)
						t.Fatalf(
							"QPS of %.2f at time %v is below minimum allowable QPS of %.2f; entire timeseries: %+v",
							qps, timeutil.Unix(0, dp.TimestampNanos), expectedQPS, datapoints)
					} else {
						__antithesis_instrumentation__.Notify(49013)
					}
				}
				__antithesis_instrumentation__.Notify(48974)
				t.Status("perf is OK!")
				t.WorkerStatus("waiting for workload to complete")
				return nil
			})
			__antithesis_instrumentation__.Notify(48963)

			m.Wait()
		},
	})
}

func registerKVSplits(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49014)
	for _, item := range []struct {
		quiesce bool
		splits  int
		timeout time.Duration
	}{

		{true, 300000, 2 * time.Hour},

		{false, 30000, 2 * time.Hour},
	} {
		__antithesis_instrumentation__.Notify(49015)
		item := item
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("kv/splits/nodes=3/quiesce=%t", item.quiesce),
			Owner:   registry.OwnerKV,
			Timeout: item.timeout,
			Cluster: r.MakeClusterSpec(4),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(49016)
				nodes := c.Spec().NodeCount - 1
				c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
				c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))

				settings := install.MakeClusterSettings()
				settings.Env = append(settings.Env, "COCKROACH_MEMPROF_INTERVAL=1m", "COCKROACH_DISABLE_QUIESCENCE="+strconv.FormatBool(!item.quiesce))
				startOpts := option.DefaultStartOpts()
				startOpts.RoachprodOpts.ExtraArgs = append(startOpts.RoachprodOpts.ExtraArgs, "--cache=256MiB")
				c.Start(ctx, t.L(), startOpts, settings, c.Range(1, nodes))

				t.Status("running workload")
				workloadCtx, workloadCancel := context.WithCancel(ctx)
				m := c.NewMonitor(workloadCtx, c.Range(1, nodes))
				m.Go(func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(49018)
					defer workloadCancel()
					concurrency := ifLocal(c, "", " --concurrency="+fmt.Sprint(nodes*64))
					splits := " --splits=" + ifLocal(c, "2000", fmt.Sprint(item.splits))
					cmd := fmt.Sprintf(
						"./workload run kv --init --max-ops=1"+
							concurrency+splits+
							" {pgurl:1-%d}",
						nodes)
					c.Run(ctx, c.Node(nodes+1), cmd)
					return nil
				})
				__antithesis_instrumentation__.Notify(49017)
				m.Wait()
			},
		})
	}
}

func registerKVScalability(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49019)
	runScalability := func(ctx context.Context, t test.Test, c cluster.Cluster, percent int) {
		__antithesis_instrumentation__.Notify(49021)
		nodes := c.Spec().NodeCount - 1

		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))

		const maxPerNodeConcurrency = 64
		for i := nodes; i <= nodes*maxPerNodeConcurrency; i += nodes {
			__antithesis_instrumentation__.Notify(49022)
			c.Wipe(ctx, c.Range(1, nodes))
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))

			t.Status("running workload")
			m := c.NewMonitor(ctx, c.Range(1, nodes))
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(49024)
				cmd := fmt.Sprintf("./workload run kv --init --read-percent=%d "+
					"--splits=1000 --duration=1m "+fmt.Sprintf("--concurrency=%d", i)+
					" {pgurl:1-%d}",
					percent, nodes)

				return c.RunE(ctx, c.Node(nodes+1), cmd)
			})
			__antithesis_instrumentation__.Notify(49023)
			m.Wait()
		}
	}
	__antithesis_instrumentation__.Notify(49020)

	if false {
		__antithesis_instrumentation__.Notify(49025)
		for _, p := range []int{0, 95} {
			__antithesis_instrumentation__.Notify(49026)
			p := p
			r.Add(registry.TestSpec{
				Name:    fmt.Sprintf("kv%d/scale/nodes=6", p),
				Owner:   registry.OwnerKV,
				Cluster: r.MakeClusterSpec(7, spec.CPU(8)),
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(49027)
					runScalability(ctx, t, c, p)
				},
			})
		}
	} else {
		__antithesis_instrumentation__.Notify(49028)
	}
}

func registerKVRangeLookups(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49029)
	type rangeLookupWorkloadType int
	const (
		splitWorkload rangeLookupWorkloadType = iota
		relocateWorkload
	)

	const (
		nodes = 8
		cpus  = 8
	)

	runRangeLookups := func(ctx context.Context, t test.Test, c cluster.Cluster, workers int, workloadType rangeLookupWorkloadType, maximumRangeLookupsPerSec float64) {
		__antithesis_instrumentation__.Notify(49031)
		nodes := c.Spec().NodeCount - 1
		doneInit := make(chan struct{})
		doneWorkload := make(chan struct{})
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, nodes))

		t.Status("running workload")

		conns := make([]*gosql.DB, nodes)
		for i := 0; i < nodes; i++ {
			__antithesis_instrumentation__.Notify(49036)
			conns[i] = c.Conn(ctx, t.L(), i+1)
		}
		__antithesis_instrumentation__.Notify(49032)
		defer func() {
			__antithesis_instrumentation__.Notify(49037)
			for i := 0; i < nodes; i++ {
				__antithesis_instrumentation__.Notify(49038)
				conns[i].Close()
			}
		}()
		__antithesis_instrumentation__.Notify(49033)
		err := WaitFor3XReplication(ctx, t, conns[0])
		require.NoError(t, err)

		m := c.NewMonitor(ctx, c.Range(1, nodes))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49039)
			defer close(doneWorkload)
			cmd := "./workload init kv --splits=1000 {pgurl:1}"
			if err = c.RunE(ctx, c.Node(nodes+1), cmd); err != nil {
				__antithesis_instrumentation__.Notify(49042)
				return err
			} else {
				__antithesis_instrumentation__.Notify(49043)
			}
			__antithesis_instrumentation__.Notify(49040)
			close(doneInit)
			concurrency := ifLocal(c, "", " --concurrency="+fmt.Sprint(nodes*64))
			duration := " --duration=10m"
			readPercent := " --read-percent=50"

			cmd = fmt.Sprintf("./workload run kv --tolerate-errors"+
				concurrency+duration+readPercent+
				" {pgurl:1-%d}", nodes)
			start := timeutil.Now()
			if err = c.RunE(ctx, c.Node(nodes+1), cmd); err != nil {
				__antithesis_instrumentation__.Notify(49044)
				return err
			} else {
				__antithesis_instrumentation__.Notify(49045)
			}
			__antithesis_instrumentation__.Notify(49041)
			end := timeutil.Now()
			verifyLookupsPerSec(ctx, c, t, c.Node(1), start, end, maximumRangeLookupsPerSec)
			return nil
		})
		__antithesis_instrumentation__.Notify(49034)

		<-doneInit
		for i := 0; i < workers; i++ {
			__antithesis_instrumentation__.Notify(49046)
			m.Go(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(49047)
				for {
					__antithesis_instrumentation__.Notify(49048)
					select {
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(49050)
						return ctx.Err()
					case <-doneWorkload:
						__antithesis_instrumentation__.Notify(49051)
						return nil
					default:
						__antithesis_instrumentation__.Notify(49052)
					}
					__antithesis_instrumentation__.Notify(49049)

					conn := conns[c.Range(1, nodes).RandNode()[0]-1]
					switch workloadType {
					case splitWorkload:
						__antithesis_instrumentation__.Notify(49053)
						_, err := conn.ExecContext(ctx, `
							ALTER TABLE
								kv.kv
							SPLIT AT
								VALUES (CAST(floor(random() * 9223372036854775808) AS INT))
						`)
						if err != nil && func() bool {
							__antithesis_instrumentation__.Notify(49056)
							return !pgerror.IsSQLRetryableError(err) == true
						}() == true {
							__antithesis_instrumentation__.Notify(49057)
							return err
						} else {
							__antithesis_instrumentation__.Notify(49058)
						}
					case relocateWorkload:
						__antithesis_instrumentation__.Notify(49054)
						newReplicas := rand.Perm(nodes)[:3]
						_, err := conn.ExecContext(ctx, `
							ALTER TABLE
								kv.kv
							EXPERIMENTAL_RELOCATE
								SELECT ARRAY[$1, $2, $3], CAST(floor(random() * 9223372036854775808) AS INT)
						`, newReplicas[0]+1, newReplicas[1]+1, newReplicas[2]+1)
						if err != nil && func() bool {
							__antithesis_instrumentation__.Notify(49059)
							return !pgerror.IsSQLRetryableError(err) == true
						}() == true && func() bool {
							__antithesis_instrumentation__.Notify(49060)
							return !kv.IsExpectedRelocateError(err) == true
						}() == true {
							__antithesis_instrumentation__.Notify(49061)
							return err
						} else {
							__antithesis_instrumentation__.Notify(49062)
						}
					default:
						__antithesis_instrumentation__.Notify(49055)
						panic("unexpected")
					}
				}
			})
		}
		__antithesis_instrumentation__.Notify(49035)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(49030)
	for _, item := range []struct {
		workers                   int
		workloadType              rangeLookupWorkloadType
		maximumRangeLookupsPerSec float64
	}{
		{2, splitWorkload, 15.0},

		{4, relocateWorkload, 50.0},
	} {
		__antithesis_instrumentation__.Notify(49063)

		item := item
		var workloadName string
		switch item.workloadType {
		case splitWorkload:
			__antithesis_instrumentation__.Notify(49065)
			workloadName = "split"
		case relocateWorkload:
			__antithesis_instrumentation__.Notify(49066)
			workloadName = "relocate"
		default:
			__antithesis_instrumentation__.Notify(49067)
			panic("unexpected")
		}
		__antithesis_instrumentation__.Notify(49064)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("kv50/rangelookups/%s/nodes=%d", workloadName, nodes),
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(nodes+1, spec.CPU(cpus)),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(49068)
				runRangeLookups(ctx, t, c, item.workers, item.workloadType, item.maximumRangeLookupsPerSec)
			},
		})
	}
}

func registerKVMultiStoreWithOverload(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49069)
	runKV := func(ctx context.Context, t test.Test, c cluster.Cluster) {
		__antithesis_instrumentation__.Notify(49071)
		nodes := c.Spec().NodeCount - 1
		c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, nodes))
		c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(nodes+1))
		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.StoreCount = 2
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, nodes))

		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		for _, name := range []string{"db1", "db2"} {
			__antithesis_instrumentation__.Notify(49076)
			constraint := "store1"
			if name == "db2" {
				__antithesis_instrumentation__.Notify(49079)
				constraint = "store2"
			} else {
				__antithesis_instrumentation__.Notify(49080)
			}
			__antithesis_instrumentation__.Notify(49077)
			if _, err := db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", name)); err != nil {
				__antithesis_instrumentation__.Notify(49081)
				t.Fatalf("failed to create %s: %v", name, err)
			} else {
				__antithesis_instrumentation__.Notify(49082)
			}
			__antithesis_instrumentation__.Notify(49078)
			if _, err := db.ExecContext(ctx, fmt.Sprintf(
				"ALTER DATABASE %s CONFIGURE ZONE USING constraints = '[+%s]', "+
					"voter_constraints = '[+%s]', num_voters = 1", name, constraint, constraint)); err != nil {
				__antithesis_instrumentation__.Notify(49083)
				t.Fatalf("failed to configure zone for %s: %v", name, err)
			} else {
				__antithesis_instrumentation__.Notify(49084)
			}
		}
		__antithesis_instrumentation__.Notify(49072)

		SetAdmissionControl(ctx, t, c, true)
		if _, err := db.ExecContext(ctx,
			"SET CLUSTER SETTING kv.range_split.by_load_enabled = 'false'"); err != nil {
			__antithesis_instrumentation__.Notify(49085)
			t.Fatalf("failed to disable load based splitting: %v", err)
		} else {
			__antithesis_instrumentation__.Notify(49086)
		}
		__antithesis_instrumentation__.Notify(49073)
		t.Status("running workload")
		dur := 20 * time.Minute
		duration := " --duration=" + ifLocal(c, "10s", dur.String())
		histograms := " --histograms=" + t.PerfArtifactsDir() + "/stats.json"
		m1 := c.NewMonitor(ctx, c.Range(1, nodes))
		m1.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49087)
			dbRegular := " --db=db1"
			concurrencyRegular := ifLocal(c, "", " --concurrency=8")
			readPercentRegular := " --read-percent=95"
			cmdRegular := fmt.Sprintf("./workload run kv --init"+
				dbRegular+histograms+concurrencyRegular+duration+readPercentRegular+
				" {pgurl:1-%d}", nodes)
			c.Run(ctx, c.Node(nodes+1), cmdRegular)
			return nil
		})
		__antithesis_instrumentation__.Notify(49074)
		m2 := c.NewMonitor(ctx, c.Range(1, nodes))
		m2.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49088)
			dbOverload := " --db=db2"
			concurrencyOverload := ifLocal(c, "", " --concurrency=64")
			readPercentOverload := " --read-percent=0"
			bs := 1 << 16
			blockSizeOverload := fmt.Sprintf(" --min-block-bytes=%d --max-block-bytes=%d",
				bs, bs)
			cmdOverload := fmt.Sprintf("./workload run kv --init"+
				dbOverload+histograms+concurrencyOverload+duration+readPercentOverload+blockSizeOverload+
				" {pgurl:1-%d}", nodes)
			c.Run(ctx, c.Node(nodes+1), cmdOverload)
			return nil
		})
		__antithesis_instrumentation__.Notify(49075)
		m1.Wait()
		m2.Wait()
	}
	__antithesis_instrumentation__.Notify(49070)

	r.Add(registry.TestSpec{
		Name:    "kv/multi-store-with-overload",
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(2, spec.CPU(8), spec.SSD(2)),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49089)
			runKV(ctx, t, c)
		},
	})
}
