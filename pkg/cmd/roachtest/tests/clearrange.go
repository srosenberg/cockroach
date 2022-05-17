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
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/util/version"
)

func registerClearRange(r registry.Registry) {
	__antithesis_instrumentation__.Notify(46560)
	for _, checks := range []bool{true, false} {
		__antithesis_instrumentation__.Notify(46561)
		checks := checks
		r.Add(registry.TestSpec{
			Name:  fmt.Sprintf(`clearrange/checks=%t`, checks),
			Owner: registry.OwnerStorage,

			Timeout: 5*time.Hour + 90*time.Minute,
			Cluster: r.MakeClusterSpec(10, spec.CPU(16)),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46563)
				runClearRange(ctx, t, c, checks)
			},
		})
		__antithesis_instrumentation__.Notify(46562)

		r.Add(registry.TestSpec{
			Name:  fmt.Sprintf(`clearrange/zfs/checks=%t`, checks),
			Skip:  "Consistently failing. See #70306 and #68420 for context.",
			Owner: registry.OwnerStorage,

			Timeout:         5*time.Hour + 120*time.Minute,
			Cluster:         r.MakeClusterSpec(10, spec.CPU(16), spec.SetFileSystem(spec.Zfs)),
			EncryptAtRandom: true,
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(46564)
				runClearRange(ctx, t, c, checks)
			},
		})
	}
}

func runClearRange(ctx context.Context, t test.Test, c cluster.Cluster, aggressiveChecks bool) {
	__antithesis_instrumentation__.Notify(46565)
	c.Put(ctx, t.Cockroach(), "./cockroach")

	t.Status("restoring fixture")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	tBegin := timeutil.Now()
	c.Run(ctx, c.Node(1), "./cockroach", "workload", "fixtures", "import", "bank",
		"--payload-bytes=10240", "--ranges=10", "--rows=65104166", "--seed=4", "--db=bigbank")
	t.L().Printf("import took %.2fs", timeutil.Since(tBegin).Seconds())
	c.Stop(ctx, t.L(), option.DefaultStopOpts())
	t.Status()

	settings := install.MakeClusterSettings()
	if aggressiveChecks {
		__antithesis_instrumentation__.Notify(46571)

		settings.Env = append(settings.Env, []string{"COCKROACH_CONSISTENCY_AGGRESSIVE=true", "COCKROACH_ENFORCE_CONSISTENT_STATS=true"}...)
	} else {
		__antithesis_instrumentation__.Notify(46572)
	}
	__antithesis_instrumentation__.Notify(46566)

	c.Start(ctx, t.L(), option.DefaultStartOpts(), settings)

	t.Status(`restoring tiny table`)
	defer t.WorkerStatus()

	if t.BuildVersion().AtLeast(version.MustParse("v19.2.0")) {
		__antithesis_instrumentation__.Notify(46573)
		conn := c.Conn(ctx, t.L(), 1)
		if _, err := conn.ExecContext(ctx, `SET CLUSTER SETTING kv.bulk_io_write.concurrent_addsstable_requests = 8`); err != nil {
			__antithesis_instrumentation__.Notify(46575)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46576)
		}
		__antithesis_instrumentation__.Notify(46574)
		conn.Close()
	} else {
		__antithesis_instrumentation__.Notify(46577)
	}
	__antithesis_instrumentation__.Notify(46567)

	c.Run(ctx, c.Node(1), `COCKROACH_CONNECT_TIMEOUT=120 ./cockroach sql --insecure -e "DROP DATABASE IF EXISTS tinybank"`)
	c.Run(ctx, c.Node(1), "./cockroach", "workload", "fixtures", "import", "bank", "--db=tinybank",
		"--payload-bytes=100", "--ranges=10", "--rows=800", "--seed=1")

	t.Status()

	numBankRanges := func() func() int {
		__antithesis_instrumentation__.Notify(46578)
		conn := c.Conn(ctx, t.L(), 1)
		defer conn.Close()

		var startHex string
		if err := conn.QueryRow(
			`SELECT to_hex(start_key) FROM crdb_internal.ranges_no_leases WHERE database_name = 'bigbank' AND table_name = 'bank' ORDER BY start_key ASC LIMIT 1`,
		).Scan(&startHex); err != nil {
			__antithesis_instrumentation__.Notify(46580)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(46581)
		}
		__antithesis_instrumentation__.Notify(46579)
		return func() int {
			__antithesis_instrumentation__.Notify(46582)
			conn := c.Conn(ctx, t.L(), 1)
			defer conn.Close()
			var n int
			if err := conn.QueryRow(
				`SELECT count(*) FROM crdb_internal.ranges_no_leases WHERE substr(to_hex(start_key), 1, length($1::string)) = $1`, startHex,
			).Scan(&n); err != nil {
				__antithesis_instrumentation__.Notify(46584)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(46585)
			}
			__antithesis_instrumentation__.Notify(46583)
			return n
		}
	}()
	__antithesis_instrumentation__.Notify(46568)

	m := c.NewMonitor(ctx)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(46586)
		c.Run(ctx, c.Node(1), `./cockroach workload init kv`)
		c.Run(ctx, c.All(), `./cockroach workload run kv --concurrency=32 --duration=1h`)
		return nil
	})
	__antithesis_instrumentation__.Notify(46569)
	m.Go(func(ctx context.Context) error {
		__antithesis_instrumentation__.Notify(46587)
		conn := c.Conn(ctx, t.L(), 1)
		defer conn.Close()

		if _, err := conn.ExecContext(ctx, `SET CLUSTER SETTING kv.range_merge.queue_enabled = true`); err != nil {
			__antithesis_instrumentation__.Notify(46593)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46594)
		}
		__antithesis_instrumentation__.Notify(46588)

		if _, err := conn.ExecContext(ctx, `SET CLUSTER SETTING kv.range_merge.queue_interval = '0s'`); err != nil {
			__antithesis_instrumentation__.Notify(46595)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46596)
		}
		__antithesis_instrumentation__.Notify(46589)

		t.WorkerStatus("dropping table")
		defer t.WorkerStatus()

		if _, err := conn.ExecContext(ctx, `ALTER TABLE bigbank.bank CONFIGURE ZONE USING gc.ttlseconds = 1200`); err != nil {
			__antithesis_instrumentation__.Notify(46597)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46598)
		}
		__antithesis_instrumentation__.Notify(46590)

		t.WorkerStatus("computing number of ranges")
		initialBankRanges := numBankRanges()

		t.WorkerStatus("dropping bank table")
		if _, err := conn.ExecContext(ctx, `DROP TABLE bigbank.bank`); err != nil {
			__antithesis_instrumentation__.Notify(46599)
			return err
		} else {
			__antithesis_instrumentation__.Notify(46600)
		}
		__antithesis_instrumentation__.Notify(46591)

		const minDuration = 45 * time.Minute
		deadline := timeutil.Now().Add(minDuration)
		curBankRanges := numBankRanges()
		t.WorkerStatus("waiting for ~", curBankRanges, " merges to complete (and for at least ", minDuration, " to pass)")
		for timeutil.Now().Before(deadline) || func() bool {
			__antithesis_instrumentation__.Notify(46601)
			return curBankRanges > 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(46602)
			after := time.After(5 * time.Minute)
			curBankRanges = numBankRanges()
			t.WorkerProgress(1 - float64(curBankRanges)/float64(initialBankRanges))

			var count int

			if _, err := conn.ExecContext(ctx, `SET statement_timeout = '5s'`); err != nil {
				__antithesis_instrumentation__.Notify(46605)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46606)
			}
			__antithesis_instrumentation__.Notify(46603)

			if err := conn.QueryRowContext(ctx, `SELECT count(*) FROM tinybank.bank`).Scan(&count); err != nil {
				__antithesis_instrumentation__.Notify(46607)
				return err
			} else {
				__antithesis_instrumentation__.Notify(46608)
			}
			__antithesis_instrumentation__.Notify(46604)

			t.WorkerStatus("waiting for ~", curBankRanges, " merges to complete (and for at least ", timeutil.Since(deadline), " to pass)")
			select {
			case <-after:
				__antithesis_instrumentation__.Notify(46609)
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(46610)
				return ctx.Err()
			}
		}
		__antithesis_instrumentation__.Notify(46592)

		return nil
	})
	__antithesis_instrumentation__.Notify(46570)
	m.Wait()
}
