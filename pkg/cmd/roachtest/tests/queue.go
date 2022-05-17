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
)

func registerQueue(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49896)

	const numNodes = 2
	r.Add(registry.TestSpec{
		Skip:    "https://github.com/cockroachdb/cockroach/issues/17229",
		Name:    fmt.Sprintf("queue/nodes=%d", numNodes-1),
		Owner:   registry.OwnerKV,
		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49897)
			runQueue(ctx, t, c)
		},
	})
}

func runQueue(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(49898)
	dbNodeCount := c.Spec().NodeCount - 1
	workloadNode := c.Spec().NodeCount

	c.Put(ctx, t.Cockroach(), "./cockroach", c.Range(1, dbNodeCount))
	c.Put(ctx, t.DeprecatedWorkload(), "./workload", c.Node(workloadNode))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Range(1, dbNodeCount))

	runQueueWorkload := func(duration time.Duration, initTables bool) {
		__antithesis_instrumentation__.Notify(49907)
		m := c.NewMonitor(ctx, c.Range(1, dbNodeCount))
		m.Go(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(49909)
			concurrency := ifLocal(c, "", " --concurrency="+fmt.Sprint(dbNodeCount*64))
			duration := fmt.Sprintf(" --duration=%s", duration.String())
			batch := " --batch 100"
			init := ""
			if initTables {
				__antithesis_instrumentation__.Notify(49911)
				init = " --init"
			} else {
				__antithesis_instrumentation__.Notify(49912)
			}
			__antithesis_instrumentation__.Notify(49910)
			cmd := fmt.Sprintf(
				"./workload run queue --histograms="+t.PerfArtifactsDir()+"/stats.json"+
					init+
					concurrency+
					duration+
					batch+
					" {pgurl:1-%d}",
				dbNodeCount,
			)
			c.Run(ctx, c.Node(workloadNode), cmd)
			return nil
		})
		__antithesis_instrumentation__.Notify(49908)
		m.Wait()
	}
	__antithesis_instrumentation__.Notify(49899)

	getQueueScanTime := func() time.Duration {
		__antithesis_instrumentation__.Notify(49913)
		db := c.Conn(ctx, t.L(), 1)
		sampleCount := 5
		samples := make([]time.Duration, sampleCount)
		for i := 0; i < sampleCount; i++ {
			__antithesis_instrumentation__.Notify(49916)
			startTime := timeutil.Now()
			var queueCount int
			row := db.QueryRow("SELECT count(*) FROM queue.queue WHERE ts < 1000")
			if err := row.Scan(&queueCount); err != nil {
				__antithesis_instrumentation__.Notify(49918)
				t.Fatalf("error running delete statement on queue: %s", err)
			} else {
				__antithesis_instrumentation__.Notify(49919)
			}
			__antithesis_instrumentation__.Notify(49917)
			endTime := timeutil.Now()
			samples[i] = endTime.Sub(startTime)
		}
		__antithesis_instrumentation__.Notify(49914)
		var sum time.Duration
		for _, sample := range samples {
			__antithesis_instrumentation__.Notify(49920)
			sum += sample
		}
		__antithesis_instrumentation__.Notify(49915)
		return sum / time.Duration(sampleCount)
	}
	__antithesis_instrumentation__.Notify(49900)

	t.Status("running initial workload")
	runQueueWorkload(10*time.Second, true)
	scanTimeBefore := getQueueScanTime()

	db := c.Conn(ctx, t.L(), 1)
	_, err := db.ExecContext(ctx, `ALTER TABLE queue.queue CONFIGURE ZONE USING gc.ttlseconds = 30`)
	if err != nil && func() bool {
		__antithesis_instrumentation__.Notify(49921)
		return strings.Contains(err.Error(), "syntax error") == true
	}() == true {
		__antithesis_instrumentation__.Notify(49922)

		_, err = db.ExecContext(ctx, `ALTER TABLE queue.queue EXPERIMENTAL CONFIGURE ZONE 'gc: {ttlseconds: 30}'`)
	} else {
		__antithesis_instrumentation__.Notify(49923)
	}
	__antithesis_instrumentation__.Notify(49901)
	if err != nil {
		__antithesis_instrumentation__.Notify(49924)
		t.Fatalf("error setting zone config TTL: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(49925)
	}
	__antithesis_instrumentation__.Notify(49902)

	if _, err := db.Exec("DELETE FROM queue.queue"); err != nil {
		__antithesis_instrumentation__.Notify(49926)
		t.Fatalf("error deleting rows after initial insertion: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(49927)
	}
	__antithesis_instrumentation__.Notify(49903)

	t.Status("running primary workload")
	runQueueWorkload(10*time.Minute, false)

	row := db.QueryRow("SELECT count(*) FROM queue.queue")
	var queueCount int
	if err := row.Scan(&queueCount); err != nil {
		__antithesis_instrumentation__.Notify(49928)
		t.Fatalf("error selecting queueCount from queue: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(49929)
	}
	__antithesis_instrumentation__.Notify(49904)
	maxRows := 100
	if c.IsLocal() {
		__antithesis_instrumentation__.Notify(49930)
		maxRows *= dbNodeCount * 64
	} else {
		__antithesis_instrumentation__.Notify(49931)
	}
	__antithesis_instrumentation__.Notify(49905)
	if queueCount > maxRows {
		__antithesis_instrumentation__.Notify(49932)
		t.Fatalf("resulting table had %d entries, expected %d or fewer", queueCount, maxRows)
	} else {
		__antithesis_instrumentation__.Notify(49933)
	}
	__antithesis_instrumentation__.Notify(49906)

	scanTimeAfter := getQueueScanTime()
	fmt.Printf("scan time before load: %s, scan time after: %s", scanTimeBefore, scanTimeAfter)
	fmt.Printf("scan time increase: %f (%f/%f)", float64(scanTimeAfter)/float64(scanTimeBefore), float64(scanTimeAfter), float64(scanTimeBefore))
	if scanTimeAfter > scanTimeBefore*30 {
		__antithesis_instrumentation__.Notify(49934)
		t.Fatalf(
			"scan time increased by factor of %f after queue workload",
			float64(scanTimeAfter)/float64(scanTimeBefore),
		)
	} else {
		__antithesis_instrumentation__.Notify(49935)
	}
}
