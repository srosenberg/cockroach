package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
)

func RegisterDiskStalledDetection(r registry.Registry) {
	__antithesis_instrumentation__.Notify(47304)
	for _, affectsLogDir := range []bool{false, true} {
		__antithesis_instrumentation__.Notify(47305)
		for _, affectsDataDir := range []bool{false, true} {
			__antithesis_instrumentation__.Notify(47306)

			affectsLogDir := affectsLogDir
			affectsDataDir := affectsDataDir
			r.Add(registry.TestSpec{
				Name: fmt.Sprintf(
					"disk-stalled/log=%t,data=%t",
					affectsLogDir, affectsDataDir,
				),
				Owner:   registry.OwnerStorage,
				Cluster: r.MakeClusterSpec(1),
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(47307)
					runDiskStalledDetection(ctx, t, c, affectsLogDir, affectsDataDir)
				},
			})
		}
	}
}

func runDiskStalledDetection(
	ctx context.Context, t test.Test, c cluster.Cluster, affectsLogDir bool, affectsDataDir bool,
) {
	__antithesis_instrumentation__.Notify(47308)
	if c.IsLocal() && func() bool {
		__antithesis_instrumentation__.Notify(47317)
		return runtime.GOOS != "linux" == true
	}() == true {
		__antithesis_instrumentation__.Notify(47318)
		t.Fatalf("must run on linux os, found %s", runtime.GOOS)
	} else {
		__antithesis_instrumentation__.Notify(47319)
	}
	__antithesis_instrumentation__.Notify(47309)

	n := c.Node(1)

	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Run(ctx, n, "sudo umount -f {store-dir}/faulty || true")
	c.Run(ctx, n, "mkdir -p {store-dir}/{real,faulty} || true")

	c.Run(ctx, n, "rm -f logs && ln -s {store-dir}/real/logs logs || true")

	t.Status("setting up charybdefs")

	if err := c.Install(ctx, t.L(), n, "charybdefs"); err != nil {
		__antithesis_instrumentation__.Notify(47320)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47321)
	}
	__antithesis_instrumentation__.Notify(47310)
	c.Run(ctx, n, "sudo charybdefs {store-dir}/faulty -oallow_other,modules=subdir,subdir={store-dir}/real")
	c.Run(ctx, n, "sudo mkdir -p {store-dir}/real/logs")
	c.Run(ctx, n, "sudo chmod -R 777 {store-dir}/{real,faulty}")

	errCh := make(chan install.RunResultDetails)

	tooShortSync := 40 * time.Millisecond

	maxLogSync := time.Hour
	logDir := "real/logs"
	if affectsLogDir {
		__antithesis_instrumentation__.Notify(47322)
		logDir = "faulty/logs"
		maxLogSync = tooShortSync
	} else {
		__antithesis_instrumentation__.Notify(47323)
	}
	__antithesis_instrumentation__.Notify(47311)
	maxDataSync := time.Hour
	dataDir := "real"
	if affectsDataDir {
		__antithesis_instrumentation__.Notify(47324)
		maxDataSync = tooShortSync
		dataDir = "faulty"
	} else {
		__antithesis_instrumentation__.Notify(47325)
	}
	__antithesis_instrumentation__.Notify(47312)

	tStarted := timeutil.Now()
	dur := 10 * time.Minute
	if !affectsDataDir && func() bool {
		__antithesis_instrumentation__.Notify(47326)
		return !affectsLogDir == true
	}() == true {
		__antithesis_instrumentation__.Notify(47327)
		dur = 30 * time.Second
	} else {
		__antithesis_instrumentation__.Notify(47328)
	}
	__antithesis_instrumentation__.Notify(47313)

	go func() {
		__antithesis_instrumentation__.Notify(47329)
		t.WorkerStatus("running server")
		result, err := c.RunWithDetailsSingleNode(ctx, t.L(), n,
			fmt.Sprintf("timeout --signal 9 %ds env "+
				"COCKROACH_ENGINE_MAX_SYNC_DURATION_DEFAULT=%s "+
				"COCKROACH_LOG_MAX_SYNC_DURATION=%s "+
				"COCKROACH_AUTO_BALLAST=false "+
				"./cockroach start-single-node --insecure --store {store-dir}/%s --log '{sinks: {stderr: {filter: INFO}}, file-defaults: {dir: \"{store-dir}/%s\"}}'",
				int(dur.Seconds()), maxDataSync, maxLogSync, dataDir, logDir,
			),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47331)
			result.Err = err
		} else {
			__antithesis_instrumentation__.Notify(47332)
		}
		__antithesis_instrumentation__.Notify(47330)
		errCh <- result
	}()
	__antithesis_instrumentation__.Notify(47314)

	time.Sleep(time.Duration(rand.Intn(5)) * time.Second)

	t.Status("blocking storage")
	c.Run(ctx, n, "charybdefs-nemesis --delay")

	result := <-errCh
	if result.Err == nil {
		__antithesis_instrumentation__.Notify(47333)
		t.Fatalf("expected an error: %s", result.Stdout)
	} else {
		__antithesis_instrumentation__.Notify(47334)
	}
	__antithesis_instrumentation__.Notify(47315)

	expectMsg := affectsDataDir || func() bool {
		__antithesis_instrumentation__.Notify(47335)
		return affectsLogDir == true
	}() == true

	if expectMsg != strings.Contains(result.Stderr, "disk stall detected") {
		__antithesis_instrumentation__.Notify(47336)
		t.Fatalf("unexpected output: %v", result.Err)
	} else {
		__antithesis_instrumentation__.Notify(47337)
		if elapsed := timeutil.Since(tStarted); !expectMsg && func() bool {
			__antithesis_instrumentation__.Notify(47338)
			return elapsed < dur == true
		}() == true {
			__antithesis_instrumentation__.Notify(47339)
			t.Fatalf("no disk stall injected, but process terminated too early after %s (expected >= %s)", elapsed, dur)
		} else {
			__antithesis_instrumentation__.Notify(47340)
		}
	}
	__antithesis_instrumentation__.Notify(47316)

	c.Run(ctx, n, "charybdefs-nemesis --clear")
	c.Run(ctx, n, "sudo umount {store-dir}/faulty")
}
