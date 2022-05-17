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
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

func registerHotSpotSplits(r registry.Registry) {
	__antithesis_instrumentation__.Notify(48411)

	runHotSpot := func(ctx context.Context, t test.Test, c cluster.Cluster, duration time.Duration, concurrency int) {
		__antithesis_instrumentation__.Notify(48413)
		roachNodes := c.Range(1, c.Spec().NodeCount-1)
		appNode := c.Node(c.Spec().NodeCount)

		c.Put(ctx, t.Cockroach(), "./cockroach", roachNodes)
		c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), roachNodes)

		c.Put(ctx, t.DeprecatedWorkload(), "./workload", appNode)
		c.Run(ctx, appNode, `./workload init kv --drop {pgurl:1}`)

		var m *errgroup.Group
		m, ctx = errgroup.WithContext(ctx)

		m.Go(func() error {
			__antithesis_instrumentation__.Notify(48416)
			t.L().Printf("starting load generator\n")

			const blockSize = 1 << 18
			return c.RunE(ctx, appNode, fmt.Sprintf(
				"./workload run kv --read-percent=0 --tolerate-errors --concurrency=%d "+
					"--min-block-bytes=%d --max-block-bytes=%d --duration=%s {pgurl:1-3}",
				concurrency, blockSize, blockSize, duration.String()))
		})
		__antithesis_instrumentation__.Notify(48414)

		m.Go(func() error {
			__antithesis_instrumentation__.Notify(48417)
			t.Status("starting checks for range sizes")
			const sizeLimit = 3 * (1 << 29)

			db := c.Conn(ctx, t.L(), 1)
			defer db.Close()

			var size = float64(0)
			for tBegin := timeutil.Now(); timeutil.Since(tBegin) <= duration; {
				__antithesis_instrumentation__.Notify(48419)
				if err := db.QueryRow(
					`select max(bytes_per_replica->'PMax') from crdb_internal.kv_store_status;`,
				).Scan(&size); err != nil {
					__antithesis_instrumentation__.Notify(48422)
					return err
				} else {
					__antithesis_instrumentation__.Notify(48423)
				}
				__antithesis_instrumentation__.Notify(48420)
				if size > sizeLimit {
					__antithesis_instrumentation__.Notify(48424)
					return errors.Errorf("range size %s exceeded %s", humanizeutil.IBytes(int64(size)),
						humanizeutil.IBytes(int64(sizeLimit)))
				} else {
					__antithesis_instrumentation__.Notify(48425)
				}
				__antithesis_instrumentation__.Notify(48421)

				t.Status(fmt.Sprintf("max range size %s", humanizeutil.IBytes(int64(size))))

				select {
				case <-ctx.Done():
					__antithesis_instrumentation__.Notify(48426)
					return ctx.Err()
				case <-time.After(5 * time.Second):
					__antithesis_instrumentation__.Notify(48427)
				}
			}
			__antithesis_instrumentation__.Notify(48418)

			return nil
		})
		__antithesis_instrumentation__.Notify(48415)
		if err := m.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(48428)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(48429)
		}
	}
	__antithesis_instrumentation__.Notify(48412)

	minutes := 10 * time.Minute
	numNodes := 4
	concurrency := 128

	r.Add(registry.TestSpec{
		Name:  fmt.Sprintf("hotspotsplits/nodes=%d", numNodes),
		Owner: registry.OwnerKV,

		Cluster: r.MakeClusterSpec(numNodes),
		Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(48430)
			if c.IsLocal() {
				__antithesis_instrumentation__.Notify(48432)
				concurrency = 32
				fmt.Printf("lowering concurrency to %d in local testing\n", concurrency)
			} else {
				__antithesis_instrumentation__.Notify(48433)
			}
			__antithesis_instrumentation__.Notify(48431)
			runHotSpot(ctx, t, c, minutes, concurrency)
		},
	})
}
