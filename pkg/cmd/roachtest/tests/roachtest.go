package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"math/rand"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
)

func registerRoachtest(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50419)
	r.Add(registry.TestSpec{
		Name:    "roachtest/noop",
		Tags:    []string{"roachtest"},
		Owner:   registry.OwnerTestEng,
		Run:     func(_ context.Context, _ test.Test, _ cluster.Cluster) { __antithesis_instrumentation__.Notify(50422) },
		Cluster: r.MakeClusterSpec(0),
	})
	__antithesis_instrumentation__.Notify(50420)
	r.Add(registry.TestSpec{
		Name:  "roachtest/noop-maybefail",
		Tags:  []string{"roachtest"},
		Owner: registry.OwnerTestEng,
		Run: func(_ context.Context, t test.Test, _ cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50423)
			if rand.Float64() <= 0.2 {
				__antithesis_instrumentation__.Notify(50424)
				t.Fatal("randomly failing")
			} else {
				__antithesis_instrumentation__.Notify(50425)
			}
		},
		Cluster: r.MakeClusterSpec(0),
	})
	__antithesis_instrumentation__.Notify(50421)

	r.Add(registry.TestSpec{
		Name:  "roachtest/hang",
		Tags:  []string{"roachtest"},
		Owner: registry.OwnerTestEng,
		Run: func(_ context.Context, t test.Test, c cluster.Cluster) {
			__antithesis_instrumentation__.Notify(50426)
			ctx := context.Background()
			c.Put(ctx, t.Cockroach(), "cockroach", c.All())
			c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.All())
			time.Sleep(time.Hour)
		},
		Timeout: 3 * time.Minute,
		Cluster: r.MakeClusterSpec(3),
	})
}
