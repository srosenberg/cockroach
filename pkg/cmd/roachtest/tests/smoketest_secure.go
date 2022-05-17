package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/registry"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/stretchr/testify/require"
)

func registerSecure(r registry.Registry) {
	__antithesis_instrumentation__.Notify(50855)
	for _, numNodes := range []int{1, 3} {
		__antithesis_instrumentation__.Notify(50856)
		r.Add(registry.TestSpec{
			Name:    fmt.Sprintf("smoketest/secure/nodes=%d", numNodes),
			Tags:    []string{"smoketest", "weekly"},
			Owner:   registry.OwnerKV,
			Cluster: r.MakeClusterSpec(numNodes),
			Run: func(ctx context.Context, t test.Test, c cluster.Cluster) {
				__antithesis_instrumentation__.Notify(50857)
				c.Put(ctx, t.Cockroach(), "./cockroach")
				settings := install.MakeClusterSettings(install.SecureOption(true))
				c.Start(ctx, t.L(), option.DefaultStartOpts(), settings)
				db := c.Conn(ctx, t.L(), 1)
				defer db.Close()
				_, err := db.QueryContext(ctx, `SELECT 1`)
				require.NoError(t, err)
			},
		})
	}
}
