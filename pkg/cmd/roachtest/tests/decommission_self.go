package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
)

func runDecommissionSelf(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(47263)

	const mainVersion = ""

	allNodes := c.All()
	u := newVersionUpgradeTest(c,
		uploadVersionStep(allNodes, mainVersion),
		startVersion(allNodes, mainVersion),
		fullyDecommissionStep(2, 2, mainVersion),
		func(ctx context.Context, t test.Test, u *versionUpgradeTest) {
			__antithesis_instrumentation__.Notify(47265)

			u.c.Wipe(ctx, c.Node(2))
		},
		checkOneMembership(1, "decommissioned"),
	)
	__antithesis_instrumentation__.Notify(47264)

	u.run(ctx, t)
}
