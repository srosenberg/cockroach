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

func registerPebbleWriteThroughput(r registry.Registry) {
	__antithesis_instrumentation__.Notify(49667)
	pebble := os.Getenv("PEBBLE_BIN")
	if pebble == "" {
		__antithesis_instrumentation__.Notify(49669)
		pebble = "./pebble.linux"
	} else {
		__antithesis_instrumentation__.Notify(49670)
	}
	__antithesis_instrumentation__.Notify(49668)

	size := 1024
	r.Add(registry.TestSpec{
		Name:    fmt.Sprintf("pebble/write/size=%d", size),
		Owner:   registry.OwnerStorage,
		Timeout: 10 * time.Hour,
		Cluster: r.MakeClusterSpec(5, spec.CPU(16), spec.SSD(16), spec.RAID0(true)),
		Tags:    []string{"pebble_nightly_write"},
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(49671)
			runPebbleWriteBenchmark(ctx, t, c, size, pebble)
		},
	})
}

func runPebbleWriteBenchmark(
	ctx context.Context, t test.Test, c cluster.Cluster, size int, bin string,
) {
	__antithesis_instrumentation__.Notify(49672)
	c.Put(ctx, bin, "./pebble")

	const (
		duration    = 8 * time.Hour
		dataDir     = "$(dirname {store-dir})"
		benchDir    = dataDir + "/bench"
		concurrency = 1024
		rateStart   = 30_000
	)

	runPebbleCmd(ctx, t, c, fmt.Sprintf(
		" ./pebble bench write %s"+
			" --wipe"+
			" --concurrency=%d"+
			" --duration=%s"+
			" --values=%d"+
			" --rate-start=%d"+
			" --debug > write.log 2>&1",
		benchDir, concurrency, duration, size, rateStart,
	))

	runPebbleCmd(ctx, t, c, "tar cvPf profiles.tar *.prof")

	dest := filepath.Join(t.ArtifactsDir(), "write.log")
	if err := c.Get(ctx, t.L(), "write.log", dest, c.All()); err != nil {
		__antithesis_instrumentation__.Notify(49675)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49676)
	}
	__antithesis_instrumentation__.Notify(49673)

	profilesName := "profiles.tar"
	dest = filepath.Join(t.ArtifactsDir(), profilesName)
	if err := c.Get(ctx, t.L(), profilesName, dest, c.All()); err != nil {
		__antithesis_instrumentation__.Notify(49677)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(49678)
	}
	__antithesis_instrumentation__.Notify(49674)

	runPebbleCmd(ctx, t, c, "rm -fr *.prof")
}
