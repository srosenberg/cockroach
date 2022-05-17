package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"fmt"
	"math"
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

func registerAllocator(r registry.Registry) {
	__antithesis_instrumentation__.Notify(45308)
	runAllocator := func(ctx context.Context, t test.Test, c cluster.Cluster, start int, maxStdDev float64) {
		__antithesis_instrumentation__.Notify(45312)
		c.Put(ctx, t.Cockroach(), "./cockroach")

		startOpts := option.DefaultStartOpts()
		startOpts.RoachprodOpts.ExtraArgs = []string{"--vmodule=store_rebalancer=5,allocator=5,allocator_scorer=5,replicate_queue=5"}
		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(1, start))
		db := c.Conn(ctx, t.L(), 1)
		defer db.Close()

		m := c.NewMonitor(ctx, c.Range(1, start))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45316)
			t.Status("loading fixture")
			if err := c.RunE(
				ctx, c.Node(1), "./cockroach", "workload", "fixtures", "import", "tpch", "--scale-factor", "10",
			); err != nil {
				__antithesis_instrumentation__.Notify(45318)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45319)
			}
			__antithesis_instrumentation__.Notify(45317)
			return nil
		})
		__antithesis_instrumentation__.Notify(45313)
		m.Wait()

		c.Start(ctx, t.L(), startOpts, install.MakeClusterSettings(), c.Range(start+1, c.Spec().NodeCount))

		c.Run(ctx, c.Node(1), `./cockroach workload init kv --drop`)
		for node := 1; node <= c.Spec().NodeCount; node++ {
			__antithesis_instrumentation__.Notify(45320)
			node := node

			go func() {
				__antithesis_instrumentation__.Notify(45321)
				const cmd = `./cockroach workload run kv --tolerate-errors --min-block-bytes=8 --max-block-bytes=127`
				l, err := t.L().ChildLogger(fmt.Sprintf(`kv-%d`, node))
				if err != nil {
					__antithesis_instrumentation__.Notify(45323)
					t.Fatal(err)
				} else {
					__antithesis_instrumentation__.Notify(45324)
				}
				__antithesis_instrumentation__.Notify(45322)
				defer l.Close()
				_ = c.RunE(ctx, c.Node(node), cmd)
			}()
		}
		__antithesis_instrumentation__.Notify(45314)

		m = c.NewMonitor(ctx, c.All())
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(45325)
			t.Status("waiting for reblance")
			return waitForRebalance(ctx, t.L(), db, maxStdDev)
		})
		__antithesis_instrumentation__.Notify(45315)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(45309)

	r.Add(registry.TestSpec{
		Name:    `replicate/up/1to3`,
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(3),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45326)
			runAllocator(ctx, t, c, 1, 10.0)
		},
	})
	__antithesis_instrumentation__.Notify(45310)
	r.Add(registry.TestSpec{
		Name:    `replicate/rebalance/3to5`,
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(5),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(45327)
			runAllocator(ctx, t, c, 3, 42.0)
		},
	})
	__antithesis_instrumentation__.Notify(45311)
	r.Add(registry.TestSpec{
		Name:    `replicate/wide`,
		Owner:   registry.OwnerKV,
		Timeout: 10 * time.Minute,
		Cluster: r.MakeClusterSpec(9, spec.CPU(1)),
		Run:     runWideReplication,
	})
}

func printRebalanceStats(l *logger.Logger, db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(45328)

	{
		__antithesis_instrumentation__.Notify(45330)
		var rebalanceIntervalStr string
		if err := db.QueryRow(
			`SELECT (SELECT max(timestamp) FROM system.rangelog) - `+
				`(SELECT max(timestamp) FROM system.eventlog WHERE "eventType"=$1)`,
			`node_join`,
		).Scan(&rebalanceIntervalStr); err != nil {
			__antithesis_instrumentation__.Notify(45332)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45333)
		}
		__antithesis_instrumentation__.Notify(45331)
		l.Printf("cluster took %s to rebalance\n", rebalanceIntervalStr)
	}

	{
		__antithesis_instrumentation__.Notify(45334)
		var rangeEvents int64
		q := `SELECT count(*) from system.rangelog`
		if err := db.QueryRow(q).Scan(&rangeEvents); err != nil {
			__antithesis_instrumentation__.Notify(45336)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45337)
		}
		__antithesis_instrumentation__.Notify(45335)
		l.Printf("%d range events\n", rangeEvents)
	}

	{
		__antithesis_instrumentation__.Notify(45338)
		var stdDev float64
		if err := db.QueryRow(
			`SELECT stddev(range_count) FROM crdb_internal.kv_store_status`,
		).Scan(&stdDev); err != nil {
			__antithesis_instrumentation__.Notify(45340)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45341)
		}
		__antithesis_instrumentation__.Notify(45339)
		l.Printf("stdDev(replica count) = %.2f\n", stdDev)
	}

	{
		__antithesis_instrumentation__.Notify(45342)
		rows, err := db.Query(`SELECT store_id, range_count FROM crdb_internal.kv_store_status`)
		if err != nil {
			__antithesis_instrumentation__.Notify(45344)
			return err
		} else {
			__antithesis_instrumentation__.Notify(45345)
		}
		__antithesis_instrumentation__.Notify(45343)
		defer rows.Close()
		for rows.Next() {
			__antithesis_instrumentation__.Notify(45346)
			var storeID, rangeCount int64
			if err := rows.Scan(&storeID, &rangeCount); err != nil {
				__antithesis_instrumentation__.Notify(45348)
				return err
			} else {
				__antithesis_instrumentation__.Notify(45349)
			}
			__antithesis_instrumentation__.Notify(45347)
			l.Printf("s%d has %d ranges\n", storeID, rangeCount)
		}
	}
	__antithesis_instrumentation__.Notify(45329)

	return nil
}

type replicationStats struct {
	SecondsSinceLastEvent int64
	EventType             string
	RangeID               int64
	StoreID               int64
	ReplicaCountStdDev    float64
}

func (s replicationStats) String() string {
	__antithesis_instrumentation__.Notify(45350)
	return fmt.Sprintf("last range event: %s for range %d/store %d (%ds ago)",
		s.EventType, s.RangeID, s.StoreID, s.SecondsSinceLastEvent)
}

func allocatorStats(db *gosql.DB) (s replicationStats, err error) {
	__antithesis_instrumentation__.Notify(45351)
	defer func() {
		__antithesis_instrumentation__.Notify(45356)
		if err != nil {
			__antithesis_instrumentation__.Notify(45357)
			s.ReplicaCountStdDev = math.MaxFloat64
		} else {
			__antithesis_instrumentation__.Notify(45358)
		}
	}()
	__antithesis_instrumentation__.Notify(45352)

	eventTypes := []interface{}{

		`split`, `add_voter`, `remove_voter`,
	}

	q := `SELECT extract_duration(seconds FROM now()-timestamp), "rangeID", "storeID", "eventType"` +
		`FROM system.rangelog WHERE "eventType" IN ($1, $2, $3) ORDER BY timestamp DESC LIMIT 1`

	row := db.QueryRow(q, eventTypes...)
	if row == nil {
		__antithesis_instrumentation__.Notify(45359)

		return replicationStats{}, errors.New("couldn't find any range events")
	} else {
		__antithesis_instrumentation__.Notify(45360)
	}
	__antithesis_instrumentation__.Notify(45353)
	if err := row.Scan(&s.SecondsSinceLastEvent, &s.RangeID, &s.StoreID, &s.EventType); err != nil {
		__antithesis_instrumentation__.Notify(45361)
		return replicationStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(45362)
	}
	__antithesis_instrumentation__.Notify(45354)

	if err := db.QueryRow(
		`SELECT stddev(range_count) FROM crdb_internal.kv_store_status`,
	).Scan(&s.ReplicaCountStdDev); err != nil {
		__antithesis_instrumentation__.Notify(45363)
		return replicationStats{}, err
	} else {
		__antithesis_instrumentation__.Notify(45364)
	}
	__antithesis_instrumentation__.Notify(45355)

	return s, nil
}

func waitForRebalance(
	ctx context.Context, l *logger.Logger, db *gosql.DB, maxStdDev float64,
) error {
	__antithesis_instrumentation__.Notify(45365)

	const statsInterval = 2 * time.Second
	const stableSeconds = 3 * 60

	var statsTimer timeutil.Timer
	defer statsTimer.Stop()
	statsTimer.Reset(statsInterval)
	for {
		__antithesis_instrumentation__.Notify(45366)
		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(45367)
			return ctx.Err()
		case <-statsTimer.C:
			__antithesis_instrumentation__.Notify(45368)
			statsTimer.Read = true
			stats, err := allocatorStats(db)
			if err != nil {
				__antithesis_instrumentation__.Notify(45371)
				return err
			} else {
				__antithesis_instrumentation__.Notify(45372)
			}
			__antithesis_instrumentation__.Notify(45369)

			l.Printf("%v\n", stats)
			if stableSeconds <= stats.SecondsSinceLastEvent {
				__antithesis_instrumentation__.Notify(45373)
				l.Printf("replica count stddev = %f, max allowed stddev = %f\n", stats.ReplicaCountStdDev, maxStdDev)
				if stats.ReplicaCountStdDev > maxStdDev {
					__antithesis_instrumentation__.Notify(45375)
					_ = printRebalanceStats(l, db)
					return errors.Errorf(
						"%ds elapsed without changes, but replica count standard "+
							"deviation is %.2f (>%.2f)", stats.SecondsSinceLastEvent,
						stats.ReplicaCountStdDev, maxStdDev)
				} else {
					__antithesis_instrumentation__.Notify(45376)
				}
				__antithesis_instrumentation__.Notify(45374)
				return printRebalanceStats(l, db)
			} else {
				__antithesis_instrumentation__.Notify(45377)
			}
			__antithesis_instrumentation__.Notify(45370)
			statsTimer.Reset(statsInterval)
		}
	}
}

func runWideReplication(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(45378)
	nodes := c.Spec().NodeCount
	if nodes != 9 {
		__antithesis_instrumentation__.Notify(45389)
		t.Fatalf("9-node cluster required")
	} else {
		__antithesis_instrumentation__.Notify(45390)
	}
	__antithesis_instrumentation__.Notify(45379)

	c.Put(ctx, t.Cockroach(), "./cockroach")
	startOpts := option.DefaultStartOpts()
	startOpts.RoachprodOpts.ExtraArgs = []string{"--vmodule=replicate_queue=6"}
	settings := install.MakeClusterSettings()
	settings.Env = append(settings.Env, "COCKROACH_SCAN_MAX_IDLE_TIME=5ms")
	c.Start(ctx, t.L(), startOpts, settings, c.All())

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()

	zones := func() []string {
		__antithesis_instrumentation__.Notify(45391)
		rows, err := db.Query(`SELECT target FROM crdb_internal.zones`)
		if err != nil {
			__antithesis_instrumentation__.Notify(45394)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45395)
		}
		__antithesis_instrumentation__.Notify(45392)
		defer rows.Close()
		var results []string
		for rows.Next() {
			__antithesis_instrumentation__.Notify(45396)
			var name string
			if err := rows.Scan(&name); err != nil {
				__antithesis_instrumentation__.Notify(45398)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45399)
			}
			__antithesis_instrumentation__.Notify(45397)
			results = append(results, name)
		}
		__antithesis_instrumentation__.Notify(45393)
		return results
	}
	__antithesis_instrumentation__.Notify(45380)

	run := func(stmt string) {
		__antithesis_instrumentation__.Notify(45400)
		t.L().Printf("%s\n", stmt)
		if _, err := db.ExecContext(ctx, stmt); err != nil {
			__antithesis_instrumentation__.Notify(45401)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45402)
		}
	}
	__antithesis_instrumentation__.Notify(45381)

	setReplication := func(width int) {
		__antithesis_instrumentation__.Notify(45403)

		for _, zone := range zones() {
			__antithesis_instrumentation__.Notify(45404)
			run(fmt.Sprintf(`ALTER %s CONFIGURE ZONE USING num_replicas = %d`, zone, width))
		}
	}
	__antithesis_instrumentation__.Notify(45382)
	setReplication(nodes)

	countMisreplicated := func(width int) int {
		__antithesis_instrumentation__.Notify(45405)
		var count int
		if err := db.QueryRow(
			"SELECT count(*) FROM crdb_internal.ranges WHERE array_length(replicas,1) != $1",
			width,
		).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(45407)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45408)
		}
		__antithesis_instrumentation__.Notify(45406)
		return count
	}
	__antithesis_instrumentation__.Notify(45383)

	waitForReplication := func(width int) {
		__antithesis_instrumentation__.Notify(45409)
		for count := -1; count != 0; time.Sleep(time.Second) {
			__antithesis_instrumentation__.Notify(45410)
			count = countMisreplicated(width)
			t.L().Printf("%d mis-replicated ranges\n", count)
		}
	}
	__antithesis_instrumentation__.Notify(45384)

	waitForReplication(nodes)

	numRanges := func() int {
		__antithesis_instrumentation__.Notify(45411)
		var count int
		if err := db.QueryRow(`SELECT count(*) FROM crdb_internal.ranges`).Scan(&count); err != nil {
			__antithesis_instrumentation__.Notify(45413)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(45414)
		}
		__antithesis_instrumentation__.Notify(45412)
		return count
	}()
	__antithesis_instrumentation__.Notify(45385)

	c.Stop(ctx, t.L(), option.DefaultStopOpts())
	tBeginDown := timeutil.Now()
	c.Start(ctx, t.L(), startOpts, settings, c.Range(1, 6))

	waitForUnderReplicated := func(count int) {
		__antithesis_instrumentation__.Notify(45415)
		for start := timeutil.Now(); ; time.Sleep(time.Second) {
			__antithesis_instrumentation__.Notify(45416)
			query := `
SELECT sum((metrics->>'ranges.unavailable')::DECIMAL)::INT AS ranges_unavailable,
       sum((metrics->>'ranges.underreplicated')::DECIMAL)::INT AS ranges_underreplicated
FROM crdb_internal.kv_store_status
`
			var unavailable, underReplicated int
			if err := db.QueryRow(query).Scan(&unavailable, &underReplicated); err != nil {
				__antithesis_instrumentation__.Notify(45419)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(45420)
			}
			__antithesis_instrumentation__.Notify(45417)
			t.L().Printf("%d unavailable, %d under-replicated ranges\n", unavailable, underReplicated)
			if unavailable != 0 {
				__antithesis_instrumentation__.Notify(45421)

				if timeutil.Since(start) >= 30*time.Second {
					__antithesis_instrumentation__.Notify(45423)
					t.Fatalf("%d unavailable ranges", unavailable)
				} else {
					__antithesis_instrumentation__.Notify(45424)
				}
				__antithesis_instrumentation__.Notify(45422)
				continue
			} else {
				__antithesis_instrumentation__.Notify(45425)
			}
			__antithesis_instrumentation__.Notify(45418)
			if underReplicated >= count {
				__antithesis_instrumentation__.Notify(45426)
				break
			} else {
				__antithesis_instrumentation__.Notify(45427)
			}
		}
	}
	__antithesis_instrumentation__.Notify(45386)

	waitForUnderReplicated(numRanges)
	if n := countMisreplicated(9); n != 0 {
		__antithesis_instrumentation__.Notify(45428)
		t.Fatalf("expected 0 mis-replicated ranges, but found %d", n)
	} else {
		__antithesis_instrumentation__.Notify(45429)
	}
	__antithesis_instrumentation__.Notify(45387)

	decom := func(id int) {
		__antithesis_instrumentation__.Notify(45430)
		c.Run(ctx, c.Node(1),
			fmt.Sprintf("./cockroach node decommission --insecure --wait=none %d", id))
	}
	__antithesis_instrumentation__.Notify(45388)

	decom(9)
	waitForReplication(7)

	run(`SET CLUSTER SETTING server.time_until_store_dead = '90s'`)

	time.Sleep(90*time.Second - timeutil.Since(tBeginDown))

	setReplication(5)
	waitForReplication(5)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(7, 9))
}
