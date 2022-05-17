package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/spec"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

func registerPebbleYCSB(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49679)
	pebble := os.Getenv("PEBBLE_BIN")
	if pebble == "" {
		__antithesis_instrumentation__.Notify(49681)
		pebble = "./pebble.linux"
	} else {
		__antithesis_instrumentation__.Notify(49682)
	}
	__antithesis_instrumentation__.Notify(49680)

	for _, dur := range []int64{10, 90} {
		__antithesis_instrumentation__.Notify(49683)
		for _, size := range []int{64, 1024} {
			__antithesis_instrumentation__.Notify(49684)
			size := size

			name := fmt.Sprintf("pebble/ycsb/size=%d", size)
			tag := "pebble_nightly_ycsb"

			if dur != 90 {
				__antithesis_instrumentation__.Notify(49686)
				tag = "pebble"
				name += fmt.Sprintf("/duration=%d", dur)
			} else {
				__antithesis_instrumentation__.Notify(49687)
			}
			__antithesis_instrumentation__.Notify(49685)

			d := dur
			r.Add(registry.TestSpec{
				Name:    name,
				Owner:   registry.OwnerStorage,
				Timeout: 12 * time.Hour,
				Cluster: r.MakeClusterSpec(5, spec.CPU(16)),
				Tags:    []string{tag},
				Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
					__antithesis_instrumentation__.Notify(49688)
					runPebbleYCSB(ctx, t, c, size, pebble, d)
				},
			})
		}
	}
}

func runPebbleYCSB(
	ctx context.Context, t test.Test, c cluster.Cluster, size int, bin string, dur int64,
) {
	__antithesis_instrumentation__.Notify(49689)
	c.Put(ctx, bin, "./pebble")

	const initialKeys = 10_000_000
	const cache = 4 << 30
	const dataDir = "$(dirname {store-dir})"
	const dataTar = dataDir + "/data.tar"
	const benchDir = dataDir + "/bench"

	var duration = time.Duration(dur) * time.Minute

	runPebbleCmd(ctx, t, c, fmt.Sprintf(
		"(./pebble bench ycsb %s"+
			" --wipe "+
			" --workload=read=100"+
			" --concurrency=1"+
			" --values=%d"+
			" --initial-keys=%d"+
			" --cache=%d"+
			" --num-ops=1 && "+
			"rm -f %s && tar cvPf %s %s) > init.log 2>&1",
		benchDir, size, initialKeys, cache, dataTar, dataTar, benchDir))

	for _, workload := range []string{"A", "B", "C", "D", "E", "F"} {
		__antithesis_instrumentation__.Notify(49690)
		keys := "zipf"
		switch workload {
		case "D":
			__antithesis_instrumentation__.Notify(49694)
			keys = "uniform"
		default:
			__antithesis_instrumentation__.Notify(49695)
		}
		__antithesis_instrumentation__.Notify(49691)

		runPebbleCmd(ctx, t, c, fmt.Sprintf(
			"rm -fr %s && tar xPf %s &&"+
				" ./pebble bench ycsb %s"+
				" --workload=%s"+
				" --concurrency=256"+
				" --values=%d"+
				" --keys=%s"+
				" --initial-keys=0"+
				" --prepopulated-keys=%d"+
				" --cache=%d"+
				" --duration=%s > ycsb.log 2>&1",
			benchDir, dataTar, benchDir, workload, size, keys, initialKeys, cache, duration))

		runPebbleCmd(ctx, t, c, fmt.Sprintf("tar cvPf profiles_%s.tar *.prof", workload))

		dest := filepath.Join(t.ArtifactsDir(), fmt.Sprintf("ycsb_%s.log", workload))
		if err := c.Get(ctx, t.L(), "ycsb.log", dest, c.All()); err != nil {
			__antithesis_instrumentation__.Notify(49696)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49697)
		}
		__antithesis_instrumentation__.Notify(49692)

		profilesName := fmt.Sprintf("profiles_%s.tar", workload)
		dest = filepath.Join(t.ArtifactsDir(), profilesName)
		if err := c.Get(ctx, t.L(), profilesName, dest, c.All()); err != nil {
			__antithesis_instrumentation__.Notify(49698)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(49699)
		}
		__antithesis_instrumentation__.Notify(49693)

		runPebbleCmd(ctx, t, c, "rm -fr *.prof")
	}
}

func runPebbleCmd(ctx context.Context, t test.Test, c cluster.Cluster, cmd string) {
	__antithesis_instrumentation__.Notify(49700)
	t.L().PrintfCtx(ctx, "> %s", cmd)
	err := c.RunE(ctx, c.All(), cmd)
	t.L().Printf("> result: %+v", err)
	if err := ctx.Err(); err != nil {
		__antithesis_instrumentation__.Notify(49702)
		t.L().Printf("(note: incoming context was canceled: %s", err)
	} else {
		__antithesis_instrumentation__.Notify(49703)
	}
	__antithesis_instrumentation__.Notify(49701)
	if err != nil {
		__antithesis_instrumentation__.Notify(49704)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49705)
	}
}
