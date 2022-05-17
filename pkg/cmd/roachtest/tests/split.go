package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/testutils/sqlutils"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	humanize "github.com/dustin/go-humanize"
	"github.com/stretchr/testify/require"
)

type splitParams struct {
	maxSize       int
	concurrency   int
	readPercent   int
	spanPercent   int
	qpsThreshold  int
	minimumRanges int
	maximumRanges int
	sequential    bool
	waitDuration  time.Duration
}

func registerLoadSplits(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50858)
	const numNodes = 3

	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("splits/load/uniform/nodes=%d", numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50861)

			expSplits := 10
			runLoadSplits(ctx, t, c, splitParams{
				maxSize:       10 << 30,
				concurrency:   64,
				readPercent:   95,
				qpsThreshold:  100,
				minimumRanges: expSplits + 1,
				maximumRanges: math.MaxInt32,

				waitDuration: 10 * time.Minute,
			})
		},
	})
	__antithesis_instrumentation__.Notify(50859)
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("splits/load/sequential/nodes=%d", numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50862)
			runLoadSplits(ctx, t, c, splitParams{
				maxSize:       10 << 30,
				concurrency:   64,
				readPercent:   0,
				qpsThreshold:  100,
				minimumRanges: 1,

				maximumRanges: 3,
				sequential:    true,
				waitDuration:  60 * time.Second,
			})
		},
	})
	__antithesis_instrumentation__.Notify(50860)
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("splits/load/spanning/nodes=%d", numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50863)
			runLoadSplits(ctx, t, c, splitParams{
				maxSize:       10 << 30,
				concurrency:   64,
				readPercent:   0,
				spanPercent:   95,
				qpsThreshold:  100,
				minimumRanges: 1,
				maximumRanges: 1,
				waitDuration:  60 * time.Second,
			})
		},
	})
}

func runLoadSplits(ctx context.Context, t test.Test, c cluster.Cluster, params splitParams) {
	__antithesis_instrumentation__.Notify(50864)
	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(1))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())

	m := c.NewMonitor(ctx, c.All())
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(50866)
		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		t.Status("disable load based splitting")
		if err := disableLoadBasedSplitting(ctx, db); err != nil {
			__antithesis_instrumentation__.Notify(50875)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50876)
		}
		__antithesis_instrumentation__.Notify(50867)

		t.Status("increasing range_max_bytes")
		minBytes := 16 << 20
		setRangeMaxBytes := func(maxBytes int) {
			__antithesis_instrumentation__.Notify(50877)
			stmtZone := fmt.Sprintf(
				"ALTER RANGE default CONFIGURE ZONE USING range_max_bytes = %d, range_min_bytes = %d",
				maxBytes, minBytes)
			if _, err := db.Exec(stmtZone); err != nil {
				__antithesis_instrumentation__.Notify(50878)
				t.Fatalf("failed to set range_max_bytes: %v", err)
			} else {
				__antithesis_instrumentation__.Notify(50879)
			}
		}
		__antithesis_instrumentation__.Notify(50868)

		setRangeMaxBytes(params.maxSize)

		t.Status("running uniform kv workload")
		c.Run(ctx, c.Node(1), fmt.Sprintf("./workload init kv {pgurl:1-%d}", c.Spec().NodeCount))

		t.Status("checking initial range count")
		rangeCount := func() int {
			__antithesis_instrumentation__.Notify(50880)
			var ranges int
			const q = "SELECT count(*) FROM [SHOW RANGES FROM TABLE kv.kv]"
			if err := db.QueryRow(q).Scan(&ranges); err != nil {
				__antithesis_instrumentation__.Notify(50882)

				if strings.Contains(err.Error(), "syntax error at or near \"ranges\"") {
					__antithesis_instrumentation__.Notify(50884)
					err = db.QueryRow("SELECT count(*) FROM [SHOW EXPERIMENTAL_RANGES FROM TABLE kv.kv]").Scan(&ranges)
				} else {
					__antithesis_instrumentation__.Notify(50885)
				}
				__antithesis_instrumentation__.Notify(50883)
				if err != nil {
					__antithesis_instrumentation__.Notify(50886)
					t.Fatalf("failed to get range count: %v", err)
				} else {
					__antithesis_instrumentation__.Notify(50887)
				}
			} else {
				__antithesis_instrumentation__.Notify(50888)
			}
			__antithesis_instrumentation__.Notify(50881)
			return ranges
		}
		__antithesis_instrumentation__.Notify(50869)
		if rc := rangeCount(); rc != 1 {
			__antithesis_instrumentation__.Notify(50889)
			return errors.Errorf("kv.kv table split over multiple ranges.")
		} else {
			__antithesis_instrumentation__.Notify(50890)
		}
		__antithesis_instrumentation__.Notify(50870)

		if _, err := db.ExecContext(ctx, fmt.Sprintf("SET CLUSTER SETTING kv.range_split.load_qps_threshold = %d",
			params.qpsThreshold)); err != nil {
			__antithesis_instrumentation__.Notify(50891)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50892)
		}
		__antithesis_instrumentation__.Notify(50871)
		t.Status("enable load based splitting")
		if _, err := db.ExecContext(ctx, `SET CLUSTER SETTING kv.range_split.by_load_enabled = true`); err != nil {
			__antithesis_instrumentation__.Notify(50893)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50894)
		}
		__antithesis_instrumentation__.Notify(50872)
		var extraFlags string
		if params.sequential {
			__antithesis_instrumentation__.Notify(50895)
			extraFlags += "--sequential"
		} else {
			__antithesis_instrumentation__.Notify(50896)
		}
		__antithesis_instrumentation__.Notify(50873)
		c.Run(ctx, c.Node(1), fmt.Sprintf("./workload run kv "+
			"--init --concurrency=%d --read-percent=%d --span-percent=%d %s {pgurl:1-%d} --duration='%s'",
			params.concurrency, params.readPercent, params.spanPercent, extraFlags, c.Spec().NodeCount,
			params.waitDuration.String()))

		t.Status("waiting for splits")
		if rc := rangeCount(); rc < params.minimumRanges || func() bool {
			__antithesis_instrumentation__.Notify(50897)
			return rc > params.maximumRanges == true
		}() == true {
			__antithesis_instrumentation__.Notify(50898)
			return errors.Errorf("kv.kv has %d ranges, expected between %d and %d splits",
				rc, params.minimumRanges, params.maximumRanges)
		} else {
			__antithesis_instrumentation__.Notify(50899)
		}
		__antithesis_instrumentation__.Notify(50874)
		return nil
	})
	__antithesis_instrumentation__.Notify(50865)
	m.Wait()
}

func registerLargeRange(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50900)
	const size = 32 << 30
	const numNodes = 6
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("splits/largerange/size=%s,nodes=%d", bytesStr(size), numNodes),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Timeout: 5 * time.Hour,
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50901)
			runLargeRangeSplits(ctx, t, c, size)
		},
	})
}

func bytesStr(size uint64) string {
	__antithesis_instrumentation__.Notify(50902)
	return strings.Replace(humanize.IBytes(size), " ", "", -1)
}

func setRangeMaxBytes(t test.Test, db *gosql.DB, minBytes, maxBytes int) {
	__antithesis_instrumentation__.Notify(50903)
	stmtZone := fmt.Sprintf(
		"ALTER RANGE default CONFIGURE ZONE USING range_max_bytes = %d, range_min_bytes = %d",
		maxBytes, minBytes)
	_, err := db.Exec(stmtZone)
	require.NoError(t, err)
}

func runLargeRangeSplits(ctx context.Context, t test.Test, c cluster.Cluster, size int) {
	__antithesis_instrumentation__.Notify(50904)

	const payload = 100

	const rowOverheadEstimate = 160
	const rowEstimate = rowOverheadEstimate + payload

	rows := size / rowEstimate
	const minBytes = 16 << 20

	c.Put(ctx, t.Cockroach(), "./cockroach", c.All())
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.All())
	numNodes := c.Spec().NodeCount
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(1))

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	rangeCount := func(t test.Test) (int, string) {
		__antithesis_instrumentation__.Notify(50909)
		const q = "SHOW RANGES FROM TABLE bank.bank"
		m, err := sqlutils.RowsToStrMatrix(sqlutils.MakeSQLRunner(db).Query(t, q))
		if err != nil {
			__antithesis_instrumentation__.Notify(50911)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(50912)
		}
		__antithesis_instrumentation__.Notify(50910)
		return len(m), sqlutils.MatrixToStr(m)
	}
	__antithesis_instrumentation__.Notify(50905)

	retryOpts := func() (retry.Options, chan struct{}) {
		__antithesis_instrumentation__.Notify(50913)

		ch := make(chan struct{})
		return retry.Options{
			InitialBackoff:      10 * time.Second,
			MaxBackoff:          time.Minute,
			Multiplier:          2.0,
			RandomizationFactor: 1.0,
			Closer:              ch,
		}, ch
	}
	__antithesis_instrumentation__.Notify(50906)

	t.Status(fmt.Sprintf("creating large bank table range (%d rows at ~%s each)", rows, humanizeutil.IBytes(rowEstimate)))
	{
		__antithesis_instrumentation__.Notify(50914)
		m := c.NewMonitor(ctx, c.Node(1))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(50916)

			if err := disableLoadBasedSplitting(ctx, db); err != nil {
				__antithesis_instrumentation__.Notify(50921)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50922)
			}
			__antithesis_instrumentation__.Notify(50917)
			if _, err := db.ExecContext(ctx, `SET CLUSTER SETTING kv.snapshot_rebalance.max_rate='512MiB'`); err != nil {
				__antithesis_instrumentation__.Notify(50923)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50924)
			}
			__antithesis_instrumentation__.Notify(50918)
			if _, err := db.ExecContext(ctx, `SET CLUSTER SETTING kv.snapshot_recovery.max_rate='512MiB'`); err != nil {
				__antithesis_instrumentation__.Notify(50925)
				return err
			} else {
				__antithesis_instrumentation__.Notify(50926)
			}
			__antithesis_instrumentation__.Notify(50919)

			setRangeMaxBytes(t, db, minBytes, 10*size)

			c.Run(ctx, c.Node(1), fmt.Sprintf("./workload init bank "+
				"--rows=%d --payload-bytes=%d --data-loader INSERT --ranges=1 {pgurl:1}", rows, payload))

			if rc, s := rangeCount(t); rc != 1 {
				__antithesis_instrumentation__.Notify(50927)
				return errors.Errorf("bank table split over multiple ranges:\n%s", s)
			} else {
				__antithesis_instrumentation__.Notify(50928)
			}
			__antithesis_instrumentation__.Notify(50920)
			return nil
		})
		__antithesis_instrumentation__.Notify(50915)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(50907)

	t.Status("waiting for full replication")
	{
		__antithesis_instrumentation__.Notify(50929)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(2, numNodes))
		m := c.NewMonitor(ctx, c.All())

		const query = `
select concat('r', range_id::string) as range, voting_replicas
from crdb_internal.ranges_no_leases
where database_name = 'bank' and cardinality(voting_replicas) >= $1;`
		tBegin := timeutil.Now()
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(50931)
			opts, ch := retryOpts()
			defer time.AfterFunc(time.Hour, func() { __antithesis_instrumentation__.Notify(50933); close(ch) }).Stop()
			__antithesis_instrumentation__.Notify(50932)

			return opts.Do(ctx, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50934)
				m, err := sqlutils.RowsToStrMatrix(sqlutils.MakeSQLRunner(db).Query(t, query, 3))
				if err != nil {
					__antithesis_instrumentation__.Notify(50937)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50938)
				}
				__antithesis_instrumentation__.Notify(50935)
				t.L().Printf("waiting for range with >= 3 replicas:\n%s", sqlutils.MatrixToStr(m))
				if len(m) == 0 {
					__antithesis_instrumentation__.Notify(50939)
					return errors.New("not replicated yet")
				} else {
					__antithesis_instrumentation__.Notify(50940)
				}
				__antithesis_instrumentation__.Notify(50936)
				return nil
			})
		})
		__antithesis_instrumentation__.Notify(50930)
		m.Wait()

		mt, err := sqlutils.RowsToStrMatrix(sqlutils.MakeSQLRunner(db).Query(t, query, 0))
		require.NoError(t, err)
		t.L().Printf("bank table replicated after %s:\n%s", timeutil.Since(tBegin), sqlutils.MatrixToStr(mt))
	}
	__antithesis_instrumentation__.Notify(50908)

	rangeSize := 64 << 20
	expRC := size/rangeSize - 3
	expSplits := expRC - 1
	t.Status(fmt.Sprintf("waiting for %d splits and rebalancing", expSplits))
	{
		__antithesis_instrumentation__.Notify(50941)
		m := c.NewMonitor(ctx, c.All())
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(50943)
			setRangeMaxBytes(t, db, minBytes, rangeSize)

			{
				__antithesis_instrumentation__.Notify(50946)

				waitDuration := time.Duration(expSplits)*time.Second + 100*time.Second

				opts, timeoutCh := retryOpts()
				defer time.AfterFunc(waitDuration, func() { __antithesis_instrumentation__.Notify(50949); close(timeoutCh) }).Stop()
				__antithesis_instrumentation__.Notify(50947)
				if err := opts.Do(ctx, func(ctx context.Context) error {
					__antithesis_instrumentation__.Notify(50950)
					if rc, _ := rangeCount(t); rc < expRC {
						__antithesis_instrumentation__.Notify(50952)

						err := errors.Errorf("bank table split over %d ranges, expected at least %d", rc, expRC)
						t.L().Printf("%v", err)
						return err
					} else {
						__antithesis_instrumentation__.Notify(50953)
					}
					__antithesis_instrumentation__.Notify(50951)
					return nil
				}); err != nil {
					__antithesis_instrumentation__.Notify(50954)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50955)
				}
				__antithesis_instrumentation__.Notify(50948)
				t.L().Printf("splits complete")
			}
			__antithesis_instrumentation__.Notify(50944)

			opts, timeoutCh := retryOpts()
			defer time.AfterFunc(time.Hour, func() { __antithesis_instrumentation__.Notify(50956); close(timeoutCh) }).Stop()
			__antithesis_instrumentation__.Notify(50945)

			return opts.Do(ctx, func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(50957)

				const q = `
			WITH ranges AS (
				SELECT replicas FROM crdb_internal.ranges_no_leases
			), store_ids AS (
				SELECT unnest(replicas) AS store_id FROM ranges
			), store_id_count AS (
				SELECT store_id, count(1) AS num_replicas FROM store_ids GROUP BY store_id
			)
			SELECT min(num_replicas), max(num_replicas) FROM store_id_count;
			`
				var minRangeCount, maxRangeCount int
				if err := db.QueryRow(q).Scan(&minRangeCount, &maxRangeCount); err != nil {
					__antithesis_instrumentation__.Notify(50960)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50961)
				}
				__antithesis_instrumentation__.Notify(50958)
				if float64(minRangeCount) < 0.8*float64(maxRangeCount) {
					__antithesis_instrumentation__.Notify(50962)
					err := errors.Errorf("rebalancing incomplete: min_range_count=%d, max_range_count=%d",
						minRangeCount, maxRangeCount)
					t.L().Printf("%v", err)
					return err
				} else {
					__antithesis_instrumentation__.Notify(50963)
				}
				__antithesis_instrumentation__.Notify(50959)
				t.L().Printf("rebalancing complete: min_range_count=%d, max_range_count=%d", minRangeCount, maxRangeCount)
				return nil
			})
		})
		__antithesis_instrumentation__.Notify(50942)
		m.Wait()
	}
}

func disableLoadBasedSplitting(ctx context.Context, db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(50964)
	_, err := db.ExecContext(ctx, `SET CLUSTER SETTING kv.range_split.by_load_enabled = false`)
	if err != nil {
		__antithesis_instrumentation__.Notify(50966)

		if !strings.Contains(err.Error(), "unknown cluster setting") {
			__antithesis_instrumentation__.Notify(50967)
			return err
		} else {
			__antithesis_instrumentation__.Notify(50968)
		}
	} else {
		__antithesis_instrumentation__.Notify(50969)
	}
	__antithesis_instrumentation__.Notify(50965)
	return nil
}
